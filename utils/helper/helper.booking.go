package helper

import (
	"cbs_backend/internal/modules/bookings/dtobookings"
	entityBooking "cbs_backend/internal/modules/bookings/entity"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// HelperBooking struct với exported name
type HelperBooking struct {
	db *gorm.DB
}

// NewHelperBooking constructor function
func NewHelperBooking(db *gorm.DB) *HelperBooking {
	return &HelperBooking{
		db: db,
	}
}

// CheckExpertAvailabilityDB kiểm tra availability của expert trong database
func (hb *HelperBooking) CheckExpertAvailabilityDB(ctx context.Context, expertID string, startTime, endTime time.Time) (bool, error) {
	var count int64
	err := hb.db.WithContext(ctx).Model(&entityBooking.ConsultationBooking{}).
		Where("expert_profile_id = ? AND booking_status NOT IN (?) AND ((booking_datetime <= ? AND booking_datetime + INTERVAL duration_minutes MINUTE > ?) OR (booking_datetime < ? AND booking_datetime + INTERVAL duration_minutes MINUTE >= ?))",
			expertID, []string{"cancelled", "completed"}, startTime, startTime, endTime, endTime).
		Count(&count).Error

	return count == 0, err
}

// CheckUserConflictDB kiểm tra conflict booking của user trong database
func (hb *HelperBooking) CheckUserConflictDB(ctx context.Context, userID string, startTime, endTime time.Time) (bool, error) {
	var count int64
	err := hb.db.WithContext(ctx).Model(&entityBooking.ConsultationBooking{}).
		Where("user_id = ? AND booking_status NOT IN (?) AND ((booking_datetime <= ? AND booking_datetime + INTERVAL duration_minutes MINUTE > ?) OR (booking_datetime < ? AND booking_datetime + INTERVAL duration_minutes MINUTE >= ?))",
			userID, []string{"cancelled", "completed"}, startTime, startTime, endTime, endTime).
		Count(&count).Error

	return count > 0, err
}

// GetBookingByID lấy booking theo ID
func (hb *HelperBooking) GetBookingByID(ctx context.Context, bookingID string) (*entityBooking.ConsultationBooking, error) {
	var booking entityBooking.ConsultationBooking
	err := hb.db.WithContext(ctx).Where("booking_id = ?", bookingID).First(&booking).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

// GetUserActiveBookings lấy danh sách booking active của user
func (hb *HelperBooking) GetUserActiveBookings(ctx context.Context, userID string) ([]entityBooking.ConsultationBooking, error) {
	var bookings []entityBooking.ConsultationBooking
	err := hb.db.WithContext(ctx).
		Where("user_id = ? AND booking_status NOT IN (?)", userID, []string{"cancelled", "completed"}).
		Order("booking_datetime ASC").
		Find(&bookings).Error

	return bookings, err
}

// GetExpertBookings lấy danh sách booking của expert
func (hb *HelperBooking) GetExpertBookings(ctx context.Context, expertID string, startDate, endDate time.Time) ([]entityBooking.ConsultationBooking, error) {
	var bookings []entityBooking.ConsultationBooking
	err := hb.db.WithContext(ctx).
		Where("expert_profile_id = ? AND booking_datetime BETWEEN ? AND ? AND booking_status NOT IN (?)",
			expertID, startDate, endDate, []string{"cancelled"}).
		Order("booking_datetime ASC").
		Find(&bookings).Error

	return bookings, err
}

// UpdateBookingStatus cập nhật trạng thái booking
func (hb *HelperBooking) UpdateBookingStatus(ctx context.Context, bookingID string, status string) error {
	return hb.db.WithContext(ctx).
		Model(&entityBooking.ConsultationBooking{}).
		Where("booking_id = ?", bookingID).
		Update("booking_status", status).Error
}

// IsTimeSlotAvailable kiểm tra time slot có available không
func (hb *HelperBooking) IsTimeSlotAvailable(ctx context.Context, expertID string, startTime time.Time, durationMinutes int) (bool, error) {
	endTime := startTime.Add(time.Duration(durationMinutes) * time.Minute)
	return hb.CheckExpertAvailabilityDB(ctx, expertID, startTime, endTime)
}

func (hb *HelperBooking) GenerateAvailableSlots(
	workingHours []dtobookings.WorkingHourRow,
	existingBookings []entityBooking.ConsultationBooking,
	unavailableTimes []dtobookings.UnavailableTime,
	fromDate, toDate time.Time,
	slotDuration int,
) []dtobookings.TimeSlot {
	var slots []dtobookings.TimeSlot
	slotDurationTime := time.Duration(slotDuration) * time.Minute

	// Loop through each day from fromDate to toDate (inclusive)
	for d := fromDate; !d.After(toDate); d = d.Add(24 * time.Hour) {
		dayOfWeek := int(d.Weekday())

		// Find working hours for this day
		var dayWorkingHours []dtobookings.WorkingHourRow
		for _, wh := range workingHours {
			if wh.DayOfWeek == dayOfWeek {
				dayWorkingHours = append(dayWorkingHours, wh)
			}
		}

		// Generate slots for each working hour period
		for _, wh := range dayWorkingHours {
			// Convert TimeOfDay to full datetime
			startDateTime := wh.StartTime.ToTime(d)
			endDateTime := wh.EndTime.ToTime(d)

			// Generate slots within working hours
			for slotStart := startDateTime; slotStart.Add(slotDurationTime).Before(endDateTime) || slotStart.Add(slotDurationTime).Equal(endDateTime); slotStart = slotStart.Add(slotDurationTime) {
				slotEnd := slotStart.Add(slotDurationTime)

				// Skip past slots with 15-minute buffer
				if slotStart.Before(time.Now().Add(15 * time.Minute)) {
					continue
				}

				// Check conflict with existing bookings
				isConflict := false
				for _, booking := range existingBookings {
					bookingEnd := booking.BookingDatetime.Add(time.Duration(booking.DurationMinutes) * time.Minute)
					// Check if slot overlaps with booking
					if slotStart.Before(bookingEnd) && slotEnd.After(booking.BookingDatetime) {
						isConflict = true
						break
					}
				}

				// Check conflict with unavailable times
				if !isConflict {
					for _, unavailable := range unavailableTimes {
						// Check if slot overlaps with unavailable time
						if slotStart.Before(unavailable.EndDatetime) && slotEnd.After(unavailable.StartDatetime) {
							isConflict = true
							break
						}
					}
				}

				// Add slot if no conflict
				if !isConflict {
					slots = append(slots, dtobookings.TimeSlot{
						StartTime:       slotStart,
						EndTime:         slotEnd,
						DurationMinutes: slotDuration,
						IsAvailable:     true,
					})
				}
			}
		}
	}

	return slots
}
func (hb *HelperBooking) updateExpertRating(tx *gorm.DB, expertProfileID uuid.UUID) error {
	var avgRating float64
	var totalReviews int64

	// Calculate new average rating
	err := tx.Table("tbl_consultation_reviews").
		Select("AVG(rating_score), COUNT(*)").
		Where("expert_profile_id = ? AND is_visible = true", expertProfileID).
		Row().Scan(&avgRating, &totalReviews)

	if err != nil {
		return fmt.Errorf("failed to calculate average rating: %w", err)
	}

	// Update expert profile
	return tx.Table("tbl_expert_profiles").
		Where("expert_profile_id = ?", expertProfileID).
		Updates(map[string]interface{}{
			"average_rating":    avgRating,
			"total_reviews":     totalReviews,
			"expert_updated_at": time.Now(),
		}).Error
}

// Helper function để log status change (có thể dùng cho các functions khác)
func (hb *HelperBooking) logStatusChange(ctx context.Context, bookingID uuid.UUID, oldStatus, newStatus string, changedByUserID uuid.UUID, reason string) error {
	statusHistory := struct {
		BookingID       uuid.UUID `gorm:"column:booking_id"`
		OldStatus       *string   `gorm:"column:old_status"`
		NewStatus       string    `gorm:"column:new_status"`
		ChangedByUserID uuid.UUID `gorm:"column:changed_by_user_id"`
		ChangeReason    *string   `gorm:"column:change_reason"`
		StatusChangedAt time.Time `gorm:"column:status_changed_at"`
	}{
		BookingID:       bookingID,
		NewStatus:       newStatus,
		ChangedByUserID: changedByUserID,
		StatusChangedAt: time.Now(),
	}

	if oldStatus != "" {
		statusHistory.OldStatus = &oldStatus
	}

	if reason != "" {
		statusHistory.ChangeReason = &reason
	}

	return hb.db.WithContext(ctx).Table("tbl_booking_status_history").Create(&statusHistory).Error
}

// func (hb *helperBooking) publishBookingEvent(ctx context.Context, event BookingEvent) error {
// 	eventData, err := json.Marshal(event)
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal booking event: %w", err)
// 	}

// 	topic := "booking-events"
// 	return kafka.PublishMessage(ctx, topic, event.BookingID, eventData)
// }
