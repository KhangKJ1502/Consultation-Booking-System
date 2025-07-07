package bookings

import (
	"cbs_backend/internal/kafka"
	"cbs_backend/internal/modules/bookings/dtobookings"
	entityBooking "cbs_backend/internal/modules/bookings/entity"
	"cbs_backend/internal/modules/experts/entity"
	"cbs_backend/internal/modules/realtime"
	entityUser "cbs_backend/internal/modules/users/entity"
	"cbs_backend/utils/cache"
	utils "cbs_backend/utils/cache"
	"cbs_backend/utils/helper"
	utilshelper "cbs_backend/utils/helper"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/bsm/redislock"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type bookingservice struct {
	db          *gorm.DB
	cache       utils.BookingCache
	logger      *zap.Logger
	helper      *utilshelper.HelperBooking
	redisLocker *redislock.Client
}

func NewBookingService(db *gorm.DB, cache cache.BookingCache, logger *zap.Logger, redisLocker *redislock.Client) *bookingservice {
	return &bookingservice{
		db:          db,
		cache:       cache,
		logger:      logger,
		helper:      helper.NewHelperBooking(db),
		redisLocker: redisLocker, // truyền vào đây!
	}
}

/*
	Thứ tự chuẩn xử lí

Input validation - Kiểm tra tất cả input trước
Calculate variables - Tính toán các biến cần thiết
Business logic validation - Kiểm tra logic nghiệp vụ
Create entity - Tạo đối tượng entity
Database transaction - Thực hiện transaction DB
Commit transaction - Commit trước khi làm side effects
Cache operations - Cập nhật cache sau khi DB success
Publish events - Publish event bất đồng bộ
Send notifications - Gửi thông báo bất đồng bộ
Build response - Tạo response object
Log & return - Log thành công và trả về
*/

func (bs *bookingservice) CreateBooking(ctx context.Context, req dtobookings.CreateBookingRequest) (*dtobookings.CreateBookingResponse, error) {
	now := time.Now()

	// 1. Validate input
	if req.BookingDatetime.Before(now) {
		return nil, fmt.Errorf("cannot book appointment in the past")
	}
	if req.DurationMinutes < 15 || req.DurationMinutes > 240 {
		return nil, fmt.Errorf("invalid duration: must be between 15-240 minutes")
	}
	if req.BookingDatetime.Sub(now) < 15*time.Minute {
		return nil, fmt.Errorf("cannot book less than 15 minutes before the appointment")
	}
	if req.BookingDatetime.After(now.Add(90 * 24 * time.Hour)) {
		return nil, fmt.Errorf("cannot book appointment more than 90 days in advance")
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}
	expertID, err := uuid.Parse(req.ExpertProfileID)
	if err != nil {
		return nil, fmt.Errorf("invalid expert profile ID format: %w", err)
	}

	// 2. Kiểm tra user/expert tồn tại, expert active
	var user entityUser.User
	if err := bs.db.WithContext(ctx).First(&user, "user_id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}
	var expert entity.ExpertProfile
	if err := bs.db.WithContext(ctx).First(&expert, "expert_profile_id = ?", expertID).Error; err != nil {
		return nil, fmt.Errorf("expert not found")
	}

	// if !expert.User.IsActive || expert.User == nil {
	// 	return nil, fmt.Errorf("expert is not active")
	// }

	// 3. Kiểm tra ngày đặt có nằm trong ngày làm việc của expert không
	goWeekday := int(req.BookingDatetime.Weekday()) // Go: 0=Chủ nhật, 1=Thứ hai, ..., 6=Thứ bảy
	dbWeekday := goWeekday + 1
	if dbWeekday > 7 {
		dbWeekday = 1 // Chủ nhật
	}
	var whCount int64
	if err := bs.db.WithContext(ctx).Model(&entity.ExpertWorkingHour{}).
		Where("expert_profile_id = ? AND day_of_week = ? AND is_active = true", expertID, dbWeekday).
		Count(&whCount).Error; err != nil {
		return nil, fmt.Errorf("failed to check expert working hours: %w", err)
	}
	if whCount == 0 {
		return nil, fmt.Errorf("expert does not work on this day")
	}

	// 4. Chặn spam đặt lịch liên tục
	var recentCount int64
	if err := bs.db.WithContext(ctx).Model(&entityBooking.ConsultationBooking{}).
		Where("user_id = ? AND booking_created_at > ?", userID, now.Add(-1*time.Minute)).
		Count(&recentCount).Error; err == nil && recentCount >= 3 {
		return nil, fmt.Errorf("too many booking requests, please wait a moment")
	}

	// 5. Redis distributed lock theo slot (expert + time)
	lockKey := fmt.Sprintf("booking:lock:%s:%s-%s", req.ExpertProfileID, req.BookingDatetime.Format(time.RFC3339), req.BookingDatetime.Add(time.Duration(req.DurationMinutes)*time.Minute).Format(time.RFC3339))
	if bs.redisLocker == nil {
		return nil, fmt.Errorf("redisLocker is not initialized")
	}
	lock, err := bs.redisLocker.Obtain(ctx, lockKey, 10*time.Second, &redislock.Options{
		RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(100*time.Millisecond), 30),
	})
	if err == redislock.ErrNotObtained {
		return nil, fmt.Errorf("another booking is being processed for this slot, please try again")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to acquire booking lock: %w", err)
	}
	defer func() {
		_ = lock.Release(ctx)
	}()

	// 6. Kiểm tra double booking và conflict trong DB (sau khi đã lock)
	startTime := req.BookingDatetime
	endTime := startTime.Add(time.Duration(req.DurationMinutes) * time.Minute)

	duplicate, err := bs.helper.CheckDuplicateBooking(ctx, req.UserID, req.ExpertProfileID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to check duplicate booking: %w", err)
	}
	if duplicate {
		return nil, fmt.Errorf("you have already booked this expert at this time")
	}

	hasConflict, err := bs.helper.CheckUserConflictDB(ctx, req.UserID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to check user booking conflicts: %w", err)
	}
	if hasConflict {
		return nil, fmt.Errorf("user has a conflicting booking at the requested time")
	}

	// 7. Transaction kiểm tra chuyên gia bị trùng lịch (FOR UPDATE)
	tx := bs.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	count, err := bs.helper.CheckExpertAvailabilityDB(ctx, req.ExpertProfileID, startTime, endTime)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to check expert availability: %w", err)
	}
	if count > 0 {
		tx.Rollback()
		return nil, fmt.Errorf("expert is not available for the requested time slot")
	}

	// 8. Tạo booking mới
	newBooking := &entityBooking.ConsultationBooking{
		UserID:           userID,
		ExpertProfileID:  expertID,
		BookingDatetime:  startTime,
		DurationMinutes:  req.DurationMinutes,
		ConsultationType: req.ConsultationType,
		BookingStatus:    "pending",
		UserNotes:        req.UserNotes,
		ConsultationFee:  req.ConsultationFee,
		PaymentStatus:    "pending",
	}
	if err := tx.Create(newBooking).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit booking: %w", err)
	}

	// 9. Cập nhật cache (nếu có)
	if bs.cache != nil {
		_ = bs.cache.CacheBooking(ctx, &cache.BookingCacheData{
			BookingID:        newBooking.BookingID.String(),
			UserID:           newBooking.UserID.String(),
			ExpertProfileID:  newBooking.ExpertProfileID.String(),
			BookingDatetime:  newBooking.BookingDatetime,
			DurationMinutes:  newBooking.DurationMinutes,
			BookingStatus:    newBooking.BookingStatus,
			ConsultationType: newBooking.ConsultationType,
		})
	}

	// 10. Gửi event, notification, trả response (giữ nguyên như cũ)
	go func() {
		// Lấy thông tin user và expert
		var user entityUser.User
		var expert entity.ExpertProfile

		if err := bs.db.First(&user, "user_id = ?", newBooking.UserID).Error; err != nil {
			log.Printf("Failed to get user info: %v", err)
			return
		}

		// Preload User khi lấy expert
		if err := bs.db.Preload("User").First(&expert, "expert_profile_id = ?", newBooking.ExpertProfileID).Error; err != nil {
			log.Printf("Failed to get expert info: %v", err)
			return
		}

		// Lấy thông tin doctorName, doctorSpecialty, email, fullName
		doctorName := ""
		doctorSpecialty := []string{}
		if expert.User != nil {
			doctorName = expert.User.FullName
		}
		if expert.SpecializationList != nil {
			doctorSpecialty = expert.SpecializationList
		}
		email := user.UserEmail
		fullName := user.FullName

		location := ""
		if newBooking.MeetingAddress != nil {
			location = *newBooking.MeetingAddress
		}
		meetingLink := ""
		if newBooking.MeetingLink != nil {
			meetingLink = *newBooking.MeetingLink
		}
		bookingNotes := ""
		if newBooking.UserNotes != nil {
			bookingNotes = *newBooking.UserNotes
		}
		amount := 0.0
		if newBooking.ConsultationFee != nil {
			amount = *newBooking.ConsultationFee
		}

		event := kafka.BookingCreatedEvent{
			EventType:          "booking_confirmation",
			UserID:             newBooking.UserID.String(),
			BookingID:          newBooking.BookingID.String(),
			ExpertID:           newBooking.ExpertProfileID.String(),
			DoctorName:         doctorName,
			DoctorSpecialty:    doctorSpecialty,
			ConsultationDate:   newBooking.BookingDatetime.Format("2006-01-02"),
			ConsultationTime:   newBooking.BookingDatetime.Format("15:04"),
			Duration:           newBooking.DurationMinutes,
			ConsultationType:   newBooking.ConsultationType,
			Location:           location,
			MeetingLink:        meetingLink,
			Amount:             amount,
			PaymentStatus:      newBooking.PaymentStatus,
			BookingNotes:       bookingNotes,
			CancellationPolicy: "Có thể hủy trước 24 giờ",
			Email:              email,
			FullName:           fullName,
			ConfirmedAt:        time.Time{}, // Chưa xác nhận, để rỗng hoặc nil
		}
		if err := kafka.PublishBookingCreatedEvent(event); err != nil {
			bs.logger.Warn("Failed to publish booking created event", zap.Error(err))
		}
	}()

	// 5. (Tùy chọn) Gửi notification realtime cho expert
	go func() {
		message := fmt.Sprintf("Bạn có booking mới! Mã: %s, Thời gian: %s",
			newBooking.BookingID.String(),
			newBooking.BookingDatetime.Format("02/01/2006 15:04"),
		)
		_ = realtime.Send(newBooking.ExpertProfileID.String(), message)
	}()

	response := &dtobookings.CreateBookingResponse{
		BookingID:        newBooking.BookingID.String(),
		UserID:           newBooking.UserID.String(),
		ExpertProfileID:  newBooking.ExpertProfileID.String(),
		BookingDatetime:  newBooking.BookingDatetime,
		DurationMinutes:  newBooking.DurationMinutes,
		ConsultationType: newBooking.ConsultationType,
		BookingStatus:    newBooking.BookingStatus,
		PaymentStatus:    newBooking.PaymentStatus,
		UserNotes:        newBooking.UserNotes,
		ConsultationFee:  newBooking.ConsultationFee,
		BookingCreatedAt: newBooking.BookingCreatedAt,
	}
	return response, nil
}
func (bs *bookingservice) ConfirmBooking(ctx context.Context, req dtobookings.ConfirmBooking) (*dtobookings.ConfirmBookingResponse, error) {
	var booking entityBooking.ConsultationBooking
	if err := bs.db.WithContext(ctx).First(&booking, "booking_id = ?", req.BookingID).Error; err != nil {
		return nil, err
	}

	if booking.ExpertProfileID.String() != req.ExpertID {
		return nil, fmt.Errorf("unauthorized")
	}

	if booking.BookingStatus != "pending" {
		return nil, fmt.Errorf("booking not in pending state")
	}

	booking.BookingStatus = "confirmed"
	if err := bs.db.WithContext(ctx).Save(&booking).Error; err != nil {
		return nil, err
	}

	// Gửi notification với error handling
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic in notification goroutine: %v", r)
			}
		}()

		log.Printf(">>> Start notification goroutine for booking %s", booking.BookingID.String())

		// Lấy thông tin user và expert
		var user entityUser.User
		var expert entity.ExpertProfile

		if err := bs.db.First(&user, "user_id = ?", booking.UserID).Error; err != nil {
			log.Printf("Failed to get user info: %v", err)
			return
		}

		// Preload User khi lấy expert
		if err := bs.db.Preload("User").First(&expert, "expert_profile_id = ?", booking.ExpertProfileID).Error; err != nil {
			log.Printf("Failed to get expert info: %v", err)
			return
		}

		// Kiểm tra expert.User nil
		doctorName := ""
		if expert.User != nil {
			doctorName = expert.User.FullName
		} else {
			log.Printf("WARNING: expert.User is nil for expert_profile_id=%s", booking.ExpertProfileID.String())
		}

		// Kiểm tra ConsultationFee nil
		var amount float64
		if booking.ConsultationFee != nil {
			amount = *booking.ConsultationFee
		} else {
			log.Printf("WARNING: ConsultationFee is nil for booking_id=%s", booking.BookingID.String())
		}

		// Tạo event với dữ liệu đầy đủ
		event := kafka.CreateBookingConfirmEvent(
			booking.UserID.String(),                      // userID
			booking.BookingID.String(),                   // bookingID
			booking.ExpertProfileID.String(),             // expertID
			user.UserEmail,                               // email
			user.FullName,                                // fullName
			doctorName,                                   // doctorName
			expert.SpecializationList,                    // doctorSpecialty (nếu có)
			booking.BookingDatetime.Format("2006-01-02"), // consultationDate
			booking.BookingDatetime.Format("15:04"),      // consultationTime
			booking.DurationMinutes,                      // duration
			booking.ConsultationType,                     // consultationType
			getLocationString(booking.MeetingAddress),    // location
			getMeetingLinkString(booking.MeetingLink),    // meetingLink
			amount,                                   // amount
			booking.PaymentStatus,                    // paymentStatus
			getBookingNotesString(booking.UserNotes), // bookingNotes
			"Có thể hủy trước 24 giờ",                // cancellationPolicy
		)

		// Publish event
		if err := kafka.PublishBookingConfirmEvent(event); err != nil {
			log.Printf("Failed to publish booking confirm event: %v", err)
		} else {
			log.Printf("✅ Booking confirm event published successfully for booking %s", booking.BookingID.String())
		}
	}()

	var meetingLink, meetingAddress string

	if booking.MeetingLink != nil {
		meetingLink = *booking.MeetingLink
	}
	if booking.MeetingAddress != nil {
		meetingAddress = *booking.MeetingAddress
	}

	res := &dtobookings.ConfirmBookingResponse{
		BookingID:       booking.BookingID.String(),
		ExpertID:        booking.ExpertProfileID.String(),
		UserID:          booking.UserID.String(),
		Status:          booking.BookingStatus,
		DurationMinutes: booking.DurationMinutes,
		MeetingLink:     meetingLink,
		MeetingAddress:  meetingAddress,
		ConfirmAt:       time.Now(),
	}

	return res, nil
}

func (bs *bookingservice) GetUpcomingBookingsForExpert(ctx context.Context, req dtobookings.GetUpcomingBookingForExpertRequest) ([]*dtobookings.BookingResponse, error) {
	// Chỉ lấy các booking với trạng thái "pending", "confirmed" và nằm trong khoảng thời gian yêu cầu
	var bookings []*entityBooking.ConsultationBooking
	err := bs.db.WithContext(ctx).
		Where("expert_profile_id = ? AND booking_datetime >= ? AND booking_datetime <= ? AND booking_status IN (?)",
			req.ExpertID, req.From, req.To, []string{"pending", "confirmed"}).
		Find(&bookings).Error

	if err != nil {
		return nil, err
	}

	// Chuyển đổi dữ liệu từ entity sang DTO
	var result []*dtobookings.BookingResponse
	for _, booking := range bookings {
		result = append(result, &dtobookings.BookingResponse{
			BookingID:        booking.BookingID.String(),
			ExpertProfileID:  booking.ExpertProfileID.String(),
			BookingDatetime:  booking.BookingDatetime,
			DurationMinutes:  booking.DurationMinutes,
			ConsultationType: booking.ConsultationType,
			BookingStatus:    booking.BookingStatus,
			UserNotes:        booking.UserNotes,
			ExpertNotes:      booking.ExpertNotes,
			MeetingLink:      booking.MeetingLink,
			MeetingAddress:   booking.MeetingAddress,
			ConsultationFee:  booking.ConsultationFee,
			PaymentStatus:    booking.PaymentStatus,
			BookingCreatedAt: booking.BookingCreatedAt,
		})
	}

	return result, nil
}

func (bs *bookingservice) CancelBooking(ctx context.Context, bookingID string, userID string) (*dtobookings.CancelResponse, error) {
	var booking entityBooking.ConsultationBooking

	// 1. Lấy thông tin booking
	if err := bs.db.WithContext(ctx).First(&booking, "booking_id = ?", bookingID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("booking not found")
		}
		return nil, err
	}

	// 2. Kiểm tra userID có quyền huỷ không
	if booking.UserID.String() != userID {
		return nil, fmt.Errorf("unauthorized: user does not own this booking")
	}

	// 3. Kiểm tra trạng thái
	if booking.BookingStatus == "completed" {
		return nil, fmt.Errorf("cannot cancel a completed booking")
	}
	if booking.BookingStatus == "cancelled" {
		return nil, fmt.Errorf("booking is already cancelled")
	}

	// 4. Kiểm tra thời gian hủy
	if time.Until(booking.BookingDatetime) < time.Hour {
		return nil, fmt.Errorf("cannot cancel booking less than 1 hour before the appointment")
	}

	// 5. Cập nhật trạng thái
	booking.BookingStatus = "cancelled"
	now := time.Now()
	booking.CancelledAt = &now
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format")
	}
	booking.CancelledByUserID = &userUUID
	if err := bs.db.WithContext(ctx).Save(&booking).Error; err != nil {
		return nil, err
	}

	// 6. Xóa cache
	if bs.cache != nil {
		_ = bs.cache.DeleteBooking(ctx, booking.BookingID.String())
	}

	// 7. Gửi notification realtime
	go realtime.Send(booking.ExpertProfileID.String(), fmt.Sprintf("Lịch hẹn %s đã bị hủy!", booking.BookingID.String()))
	go realtime.Send(booking.UserID.String(), fmt.Sprintf("Lịch hẹn %s của bạn đã bị hủy!", booking.BookingID.String()))

	// 8. Gửi Kafka event
	go func() {
		var user entityUser.User
		var expert entity.ExpertProfile

		if err := bs.db.First(&user, "user_id = ?", booking.UserID).Error; err != nil {
			log.Printf("❌ Failed to get user info: %v", err)
			return
		}
		if err := bs.db.Preload("User").First(&expert, "expert_profile_id = ?", booking.ExpertProfileID).Error; err != nil {
			log.Printf("❌ Failed to get expert info: %v", err)
			return
		}

		// Mapping data
		doctorName := ""
		if expert.User != nil {
			doctorName = expert.User.FullName
		}
		doctorSpecialty := expert.SpecializationList
		location := ""
		if booking.MeetingAddress != nil {
			location = *booking.MeetingAddress
		}
		meetingLink := ""
		if booking.MeetingLink != nil {
			meetingLink = *booking.MeetingLink
		}
		amount := 0.0
		if booking.ConsultationFee != nil {
			amount = *booking.ConsultationFee
		}
		cancellationBy := "system"
		if booking.CancelledByUserID != nil {
			if *booking.CancelledByUserID == booking.UserID {
				cancellationBy = "patient"
			} else {
				cancellationBy = "expert"
			}
		}
		cancellationNote := ""
		if booking.CancellationReason != nil {
			cancellationNote = *booking.CancellationReason
		}
		refundAmount := 0.0 // bạn có thể thêm trường refund trong DB nếu cần
		refundDays := 7     // giả định default xử lý trong 7 ngày
		cancelledAt := now

		event := kafka.BookingCancelledEvent{
			EventType:         "booking_cancelled",
			UserID:            booking.UserID.String(),
			BookingID:         booking.BookingID.String(),
			ExpertID:          booking.ExpertProfileID.String(),
			DoctorName:        doctorName,
			DoctorSpecialty:   doctorSpecialty,
			ConsultationDate:  booking.BookingDatetime.Format("02-01-2006"),
			ConsultationTime:  booking.BookingDatetime.Format("15:04"),
			Duration:          booking.DurationMinutes,
			ConsultationType:  booking.ConsultationType,
			Location:          location,
			MeetingLink:       meetingLink,
			Amount:            amount,
			PaymentStatus:     booking.PaymentStatus,
			Email:             user.UserEmail,
			FullName:          user.FullName,
			CancellationBy:    cancellationBy,
			CancellationNote:  cancellationNote,
			RefundAmount:      refundAmount,
			RefundProcessDays: refundDays,
			CancelledAt:       cancelledAt,
		}

		if err := kafka.PublishBookingCancelledEvent(event); err != nil {
			bs.logger.Warn("❌ Failed to publish booking cancelled event", zap.Error(err))
		}
	}()

	// 9. Trả response
	return &dtobookings.CancelResponse{
		BookingID:      bookingID,
		CancelByUserID: userID,
		Status:         booking.BookingStatus,
		CancelledAt:    now,
	}, nil
}

// Helper functions để xử lý pointer values
func getLocationString(location *string) string {
	if location != nil {
		return *location
	}
	return ""
}

func getMeetingLinkString(link *string) string {
	if link != nil {
		return *link
	}
	return ""
}

func getBookingNotesString(notes *string) string {
	if notes != nil {
		return *notes
	}
	return ""
}
func (bs *bookingservice) GetAvailableSlots(
	ctx context.Context,
	req dtobookings.GetAvailableSlotsRequest,
) (*dtobookings.GetAvailableSlotsResponse, error) {
	// 1. Parse and validate input
	expertID, err := uuid.Parse(req.ExpertProfileID)
	if err != nil {
		return nil, fmt.Errorf("invalid expert_profile_id: %w", err)
	}

	if req.FromDate.After(req.ToDate) {
		return nil, fmt.Errorf("from_date cannot be after to_date")
	}

	if req.FromDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return nil, fmt.Errorf("cannot get slots for past dates")
	}

	// 2. Get expert working hours
	var workingHours []dtobookings.WorkingHourRow
	if err := bs.db.WithContext(ctx).
		Model(&entity.ExpertWorkingHour{}).
		Select("day_of_week, start_time, end_time").
		Where("expert_profile_id = ? AND is_active = true", expertID).
		Scan(&workingHours).Error; err != nil {
		return nil, fmt.Errorf("failed to get expert working hours: %w", err)
	}

	if len(workingHours) == 0 {
		return &dtobookings.GetAvailableSlotsResponse{
			ExpertProfileID: req.ExpertProfileID,
			FromDate:        req.FromDate,
			ToDate:          req.ToDate,
			AvailableSlots:  []dtobookings.TimeSlot{},
			TotalSlots:      0,
			Message:         "Expert has no working hours configured",
		}, nil
	}

	// 3. Get existing bookings that might conflict
	var existingBookings []entityBooking.ConsultationBooking
	if err := bs.db.WithContext(ctx).
		Where("expert_profile_id = ? AND booking_datetime >= ? AND booking_datetime <= ? AND booking_status IN (?)",
			expertID, req.FromDate, req.ToDate.Add(24*time.Hour), []string{"confirmed", "pending"}).
		Find(&existingBookings).Error; err != nil {
		return nil, fmt.Errorf("failed to get existing bookings: %w", err)
	}

	// 4. Get unavailable times
	var unavailableTimes []dtobookings.UnavailableTime
	if err := bs.db.WithContext(ctx).
		Model(&entity.ExpertUnavailableTime{}).
		Select("unavailable_start_datetime, unavailable_end_datetime").
		Where("expert_profile_id = ? AND unavailable_start_datetime <= ? AND unavailable_end_datetime>= ?",
			expertID, req.ToDate.Add(24*time.Hour), req.FromDate).
		Scan(&unavailableTimes).Error; err != nil {
		return nil, fmt.Errorf("failed to get unavailable times: %w", err)
	}

	// 5. Generate available slots
	availableSlots := bs.helper.GenerateAvailableSlots(
		workingHours,
		existingBookings,
		unavailableTimes,
		req.FromDate,
		req.ToDate,
		req.SlotDurationMinutes,
	)

	return &dtobookings.GetAvailableSlotsResponse{
		ExpertProfileID: req.ExpertProfileID,
		FromDate:        req.FromDate,
		ToDate:          req.ToDate,
		AvailableSlots:  availableSlots,
		TotalSlots:      len(availableSlots),
	}, nil
}

func (bs *bookingservice) UpdateBookingNotes(ctx context.Context, req dtobookings.UpdateBookingNotesRequest) (*dtobookings.UpdateBookingNotesResponse, error) {
	// Validate booking ID
	bookingID, err := uuid.Parse(req.BookingID)
	if err != nil {
		bs.logger.Error("Invalid booking ID format", zap.String("bookingID", req.BookingID), zap.Error(err))
		return nil, fmt.Errorf("invalid booking ID format: %w", err)
	}

	// Validate user ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		bs.logger.Error("Invalid user ID format", zap.String("userID", req.UserID), zap.Error(err))
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	var booking entityBooking.ConsultationBooking

	// Get booking
	if err := bs.db.WithContext(ctx).First(&booking, "booking_id = ?", bookingID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("booking not found")
		}
		bs.logger.Error("Failed to get booking", zap.Error(err))
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	// Check authorization - chỉ user hoặc expert mới được update
	canUpdate := false
	updateField := ""
	updatedBy := ""

	if booking.UserID == userID {
		canUpdate = true
		updateField = "user_notes"
		updatedBy = "user"
	} else if booking.ExpertProfileID == userID {
		canUpdate = true
		updateField = "expert_notes"
		updatedBy = "expert"
	}

	if !canUpdate {
		return nil, fmt.Errorf("unauthorized: you don't have permission to update this booking")
	}

	// Không cho update nếu booking đã hoàn thành hoặc bị hủy
	if booking.BookingStatus == "completed" || booking.BookingStatus == "cancelled" {
		return nil, fmt.Errorf("cannot update notes for completed or cancelled booking")
	}

	// Không cho update nếu cuộc hẹn đã qua
	if booking.BookingDatetime.Before(time.Now()) {
		return nil, fmt.Errorf("cannot update notes for past appointments")
	}

	// Update notes
	now := time.Now()
	updates := map[string]interface{}{
		"booking_updated_at": now,
	}

	if updateField == "user_notes" {
		updates["user_notes"] = req.Notes
	} else {
		updates["expert_notes"] = req.Notes
	}

	if err := bs.db.WithContext(ctx).Model(&booking).Updates(updates).Error; err != nil {
		bs.logger.Error("Failed to update booking notes", zap.Error(err))
		return nil, fmt.Errorf("failed to update booking notes: %w", err)
	}

	// Log activity
	bs.logger.Info("Booking notes updated",
		zap.String("bookingID", req.BookingID),
		zap.String("userID", req.UserID),
		zap.String("field", updateField))

	// Return success response
	response := &dtobookings.UpdateBookingNotesResponse{
		BookingID: req.BookingID,
		UpdatedAt: now,
		UpdatedBy: updatedBy,
		Message:   "Booking notes updated successfully",
	}

	return response, nil
}

func (bs *bookingservice) GetBookingStatusHistory(ctx context.Context, req dtobookings.GetBookingStatusHistoryRequest) (*dtobookings.GetBookingStatusHistoryResponse, error) {
	// Validate booking ID
	bookingID, err := uuid.Parse(req.BookingID)
	if err != nil {
		bs.logger.Error("Invalid booking ID format", zap.String("bookingID", req.BookingID), zap.Error(err))
		return nil, fmt.Errorf("invalid booking ID format: %w", err)
	}

	// Check if booking exists and user has permission
	var booking entityBooking.ConsultationBooking
	if err := bs.db.WithContext(ctx).First(&booking, "booking_id = ?", bookingID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("booking not found")
		}
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	// Authorization check
	userUUID, _ := uuid.Parse(req.UserID)
	if booking.UserID != userUUID && booking.ExpertProfileID != userUUID {
		return nil, fmt.Errorf("unauthorized: you don't have permission to view this booking history")
	}

	// Get status history
	var historyRecords []struct {
		StatusHistoryID uuid.UUID `gorm:"column:status_history_id"`
		OldStatus       *string   `gorm:"column:old_status"`
		NewStatus       string    `gorm:"column:new_status"`
		ChangedByUserID uuid.UUID `gorm:"column:changed_by_user_id"`
		ChangeReason    *string   `gorm:"column:change_reason"`
		StatusChangedAt time.Time `gorm:"column:status_changed_at"`
		ChangedByName   string    `gorm:"column:changed_by_name"`
	}

	err = bs.db.WithContext(ctx).
		Table("tbl_booking_status_history h").
		Select("h.status_history_id, h.old_status, h.new_status, h.changed_by_user_id, h.change_reason, h.status_changed_at, u.full_name as changed_by_name").
		Joins("LEFT JOIN tbl_users u ON h.changed_by_user_id = u.user_id").
		Where("h.booking_id = ?", bookingID).
		Order("h.status_changed_at ASC").
		Scan(&historyRecords).Error

	if err != nil {
		bs.logger.Error("Failed to get booking status history", zap.Error(err))
		return nil, fmt.Errorf("failed to get booking status history: %w", err)
	}

	// Convert to response format
	var historyItems []dtobookings.StatusHistoryItem
	for _, record := range historyRecords {
		oldStatus := ""
		if record.OldStatus != nil {
			oldStatus = *record.OldStatus
		}

		reason := ""
		if record.ChangeReason != nil {
			reason = *record.ChangeReason
		}

		historyItems = append(historyItems, dtobookings.StatusHistoryItem{
			StatusHistoryID: record.StatusHistoryID.String(),
			OldStatus:       oldStatus,
			NewStatus:       record.NewStatus,
			ChangedByUserID: record.ChangedByUserID.String(),
			ChangedByName:   record.ChangedByName,
			ChangeReason:    reason,
			StatusChangedAt: record.StatusChangedAt,
		})
	}

	response := &dtobookings.GetBookingStatusHistoryResponse{
		BookingID:     req.BookingID,
		StatusHistory: historyItems,
		TotalRecords:  len(historyItems),
	}

	return response, nil
}
