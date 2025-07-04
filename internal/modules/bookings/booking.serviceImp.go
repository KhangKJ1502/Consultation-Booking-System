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

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type bookingservice struct {
	db     *gorm.DB
	cache  utils.BookingCache
	logger *zap.Logger
	helper *utilshelper.HelperBooking
}

func NewBookingService(db *gorm.DB, cache cache.BookingCache, logger *zap.Logger) *bookingservice {
	return &bookingservice{
		db:     db,
		cache:  cache,
		logger: logger,
		helper: helper.NewHelperBooking(db),
	}
}

func (bs *bookingservice) CreateBooking(ctx context.Context, req dtobookings.CreateBookingRequest) (*dtobookings.CreateBookingResponse, error) {
	// Validate booking time (không được đặt quá khứ)
	if req.BookingDatetime.Before(time.Now()) {
		return nil, fmt.Errorf("cannot book appointment in the past")
	}

	// Validate duration (tối thiểu 15 phút, tối đa 4 giờ)
	if req.DurationMinutes < 15 || req.DurationMinutes > 240 {
		return nil, fmt.Errorf("invalid duration: must be between 15-240 minutes")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		bs.logger.Error("Invalid user ID format", zap.String("userID", req.UserID), zap.Error(err))
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	expertID, err := uuid.Parse(req.ExpertProfileID)
	if err != nil {
		bs.logger.Error("Invalid expert profile ID format", zap.String("expertProfileID", req.ExpertProfileID), zap.Error(err))
		return nil, fmt.Errorf("invalid expert profile ID format: %w", err)
	}

	startTime := req.BookingDatetime
	endTime := startTime.Add(time.Duration(req.DurationMinutes) * time.Minute)

	// Cache checks with Redis
	isAvailable, err := bs.cache.IsExpertAvailable(ctx, req.ExpertProfileID, startTime, endTime)
	if err != nil {
		bs.logger.Warn("Cache check failed, falling back to database", zap.Error(err))
		isAvailable, err = bs.helper.CheckExpertAvailabilityDB(ctx, req.ExpertProfileID, startTime, endTime)
		if err != nil {
			return nil, fmt.Errorf("failed to check expert availability: %w", err)
		}
	}

	if !isAvailable {
		return nil, fmt.Errorf("expert is not available for the requested time slot")
	}

	hasConflict, err := bs.cache.HasConflictingBooking(ctx, req.UserID, startTime, endTime)
	if err != nil {
		bs.logger.Warn("Cache conflict check failed, falling back to database", zap.Error(err))
		hasConflict, err = bs.helper.CheckUserConflictDB(ctx, req.UserID, startTime, endTime)
		if err != nil {
			return nil, fmt.Errorf("failed to check user booking conflicts: %w", err)
		}
	}

	if hasConflict {
		return nil, fmt.Errorf("user has a conflicting booking at the requested time")
	}
	realtime.Send(req.ExpertProfileID, "Bạn có booking mới!")

	// Create booking
	newBooking := &entityBooking.ConsultationBooking{
		UserID:           userID,
		ExpertProfileID:  expertID,
		BookingDatetime:  req.BookingDatetime,
		DurationMinutes:  req.DurationMinutes,
		ConsultationType: req.ConsultationType,
		BookingStatus:    "pending",
		UserNotes:        req.UserNotes,
		ConsultationFee:  req.ConsultationFee,
		PaymentStatus:    "pending",
	}

	// Database transaction
	tx := bs.db.WithContext(ctx).Begin()
	if err := tx.Create(newBooking).Error; err != nil {
		tx.Rollback()
		bs.logger.Error("Failed to create booking", zap.Error(err))
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	// Cache booking in Redis
	bookingCacheData := &cache.BookingCacheData{
		BookingID:        newBooking.BookingID.String(),
		UserID:           newBooking.UserID.String(),
		ExpertProfileID:  newBooking.ExpertProfileID.String(),
		BookingDatetime:  newBooking.BookingDatetime,
		DurationMinutes:  newBooking.DurationMinutes,
		BookingStatus:    newBooking.BookingStatus,
		ConsultationType: newBooking.ConsultationType,
	}

	if err := bs.cache.CacheBooking(ctx, bookingCacheData); err != nil {
		bs.logger.Warn("Failed to cache booking", zap.Error(err))
	}

	// Publish Kafka event
	event := kafka.BookingEvent{
		EventType: "booking-create",
		BookingID: newBooking.BookingID.String(),
		UserID:    newBooking.UserID.String(),
		ExpertID:  newBooking.ExpertProfileID.String(),
		Timestamp: time.Now(),
		EventData: map[string]interface{}{
			"consultation_type": newBooking.ConsultationType,
			"booking_datetime":  newBooking.BookingDatetime,
			"duration_minutes":  newBooking.DurationMinutes,
			"consultation_fee":  newBooking.ConsultationFee,
		},
	}

	if err := kafka.PublishBookingEvent(event); err != nil {
		bs.logger.Warn("Failed to publish booking created event", zap.Error(err))
	}

	if err := tx.Commit().Error; err != nil {
		bs.logger.Error("Failed to commit transaction", zap.Error(err))
		return nil, fmt.Errorf("failed to commit booking creation: %w", err)
	}
	// Gửi notification realtime cho chuyên gia (bất đồng bộ, sau khi commit)
	go realtime.Send(
		newBooking.ExpertProfileID.String(),
		fmt.Sprintf("Bạn có booking mới! Mã: %s, Thời gian: %s",
			newBooking.BookingID.String(),
			newBooking.BookingDatetime.Format("02/01/2006 15:04"),
		),
	)

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

	bs.logger.Info("Booking created successfully", zap.String("bookingID", response.BookingID))
	return response, nil
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

	// Lấy thông tin booking
	if err := bs.db.WithContext(ctx).First(&booking, "booking_id = ?", bookingID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("booking not found")
		}
		return nil, err
	}

	// Kiểm tra userID có quyền huỷ không (chỉ chủ sở hữu booking được huỷ)
	if booking.UserID.String() != userID {
		return nil, fmt.Errorf("unauthorized: user does not own this booking")
	}

	// Kiểm tra thời gian: phải huỷ trước ít nhất 1 giờ
	if time.Until(booking.BookingDatetime) < time.Hour {
		return nil, fmt.Errorf("cannot cancel booking less than 1 hour before the appointment")
	}

	// Kiểm tra trạng thái booking
	if booking.BookingStatus == "cancelled" {
		return nil, fmt.Errorf("booking is already cancelled")
	}

	// Cập nhật trạng thái
	booking.BookingStatus = "cancelled"
	if err := bs.db.WithContext(ctx).Save(&booking).Error; err != nil {
		return nil, err
	}
	// Gửi notification realtime cho cả user và expert
	go realtime.Send(
		booking.ExpertProfileID.String(),
		fmt.Sprintf("Lịch hẹn %s đã bị hủy!", booking.BookingID.String()),
	)
	go realtime.Send(
		booking.UserID.String(),
		fmt.Sprintf("Lịch hẹn %s của bạn đã bị hủy!", booking.BookingID.String()),
	)

	// Gửi notification cho chuyên gia (placeholder, bạn có thể tích hợp message queue / websocket)
	go func() {
		// bs.notificationService.NotifyExpert(booking.ExpertID, "Booking has been cancelled", ...)
		fmt.Printf("Notification sent to expert %s: booking %s cancelled\n", booking.ExpertProfileID.String(), booking.BookingID.String())
	}()

	return &dtobookings.CancelResponse{
		BookingID:      bookingID,
		CancelByUserID: userID,
		Status:         booking.BookingStatus,
		CancelledAt:    time.Now(),
	}, nil
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
		userID := booking.UserID.String()
		message := fmt.Sprintf("Lịch hẹn %s của bạn đã được chuyên gia xác nhận!", booking.BookingID.String())

		log.Printf("Attempting to send notification to user %s", userID)

		err := realtime.Send(userID, message)
		if err != nil {
			log.Printf("Failed to send realtime notification to user %s: %v", userID, err)

			// 🔁 Fallback: Gửi email thông qua Kafka
			// Cần lấy thông tin user và expert từ DB
			var user entityUser.User
			var expert entity.ExpertProfile

			// Lấy thông tin user
			if err := bs.db.First(&user, "user_id = ?", booking.UserID).Error; err != nil {
				log.Printf("Failed to get user info: %v", err)
				return
			}

			// Lấy thông tin expert
			if err := bs.db.First(&expert, "expert_profile_id = ?", booking.ExpertProfileID).Error; err != nil {
				log.Printf("Failed to get expert info: %v", err)
				return
			}

			// Tạo BookingConfirmEvent sử dụng helper function
			event := kafka.CreateBookingConfirmEvent(
				booking.UserID.String(),
				booking.BookingID.String(),
				booking.ExpertProfileID.String(),
				user.UserEmail,
				user.FullName,
				expert.User.FullName,
				booking.BookingDatetime.Format("2006-01-02"),
				booking.BookingDatetime.Format("15:04"),
				booking.DurationMinutes,
				booking.ConsultationType,
				getLocationString(booking.MeetingAddress),
				getMeetingLinkString(booking.MeetingLink),
				*booking.ConsultationFee,
				booking.PaymentStatus,
				getBookingNotesString(booking.UserNotes),
				"Có thể hủy trước 24 giờ",
			)

			// Publish event sử dụng dedicated publisher
			if err := kafka.PublishBookingConfirmEvent(event); err != nil {
				log.Printf("Failed to publish booking confirm event: %v", err)
			} else {
				log.Printf("Booking confirm event published successfully for booking %s", booking.BookingID.String())
			}
		} else {
			log.Printf("Notification sent successfully to user %s", userID)
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
