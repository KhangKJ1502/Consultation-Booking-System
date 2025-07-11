package worker

import (
	entityBooking "cbs_backend/internal/modules/bookings/entity"
	entityExpert "cbs_backend/internal/modules/experts/entity"
	entityNotify "cbs_backend/internal/modules/system_notification/entity"
	entitySystem "cbs_backend/internal/modules/system_setting/entity"
	"cbs_backend/internal/service/interfaces"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BookingData represents the booking information structure
type BookingData struct {
	BookingID        string    `json:"booking_id"`
	UserID           string    `json:"user_id"`
	ExpertProfileID  string    `json:"expert_profile_id"`
	BookingDatetime  time.Time `json:"booking_datetime"`
	UserEmail        string    `json:"user_email"`
	UserFullName     string    `json:"user_full_name"`
	ExpertEmail      string    `json:"expert_email"`
	ExpertFullName   string    `json:"expert_full_name"`
	ConsultationType string    `json:"consultation_type"`
	MeetingLink      string    `json:"meeting_link"`
	MeetingAddress   string    `json:"meeting_address"`
}

// WeeklyStats represents weekly statistics structure
type WeeklyStats struct {
	TotalBookings     int64   `json:"total_bookings"`
	CompletedBookings int64   `json:"completed_bookings"`
	CancelledBookings int64   `json:"cancelled_bookings"`
	MissedBookings    int64   `json:"missed_bookings"`
	RevenueTotal      float64 `json:"revenue_total"`
}

// DuplicateBooking represents duplicate booking information
type DuplicateBooking struct {
	ExpertProfileID string    `json:"expert_profile_id"`
	BookingDatetime time.Time `json:"booking_datetime"`
	Count           int       `json:"count"`
}

// NotificationDispatcher interface for dispatching notifications
type NotificationDispatcher interface {
	DispatchNotification(jobType string, payload interface{}) error
}

// ReminderService handles all booking reminder operations
type ReminderService struct {
	db           *gorm.DB
	dispatcher   NotificationDispatcher
	emailService interfaces.EmailService
}

// NewReminderService creates a new instance of ReminderService
func NewReminderService(db *gorm.DB, emailService interfaces.EmailService, dispatcher NotificationDispatcher) *ReminderService {
	return &ReminderService{
		db:           db,
		emailService: emailService,
		dispatcher:   dispatcher,
	}
}

// ===========================================
// MAIN REMINDER FUNCTIONALITY
// ===========================================

// SendBookingReminders sends reminders for upcoming bookings
func (rs *ReminderService) SendBookingReminders() error {
	log.Println("📧 Starting to send booking reminders...")

	bookings, err := rs.getUpcomingBookings()
	if err != nil {
		return fmt.Errorf("failed to get upcoming bookings: %w", err)
	}

	log.Printf("🔍 Found %d bookings to send reminders", len(bookings))

	successCount := 0
	for _, booking := range bookings {
		if rs.processBookingReminder(booking) {
			successCount++
		}
	}

	log.Printf("✅ Completed sending reminders for %d/%d bookings", successCount, len(bookings))
	return nil
}

// ===========================================
// BOOKING REMINDER HELPER METHODS
// ===========================================

// getUpcomingBookings retrieves bookings that need reminders
func (rs *ReminderService) getUpcomingBookings() ([]BookingData, error) {
	var bookings []BookingData
	query := `
		SELECT 
			cb.booking_id,
			cb.user_id,
			cb.expert_profile_id,
			cb.booking_datetime,
			u.user_email,
			u.full_name as user_full_name,
			eu.user_email as expert_email,
			eu.full_name as expert_full_name,
			cb.consultation_type,
			cb.meeting_link,
			cb.meeting_address
		FROM tbl_consultation_bookings cb
		JOIN tbl_users u ON cb.user_id = u.user_id
		JOIN tbl_expert_profiles ep ON cb.expert_profile_id = ep.expert_profile_id
		JOIN tbl_users eu ON ep.user_id = eu.user_id
		WHERE cb.booking_status = 'confirmed'
		AND cb.reminder_sent = false
		AND cb.booking_datetime BETWEEN NOW() AND NOW() + INTERVAL '1 hour'
	`

	return bookings, rs.db.Raw(query).Scan(&bookings).Error
}

// processBookingReminder processes a single booking reminder
func (rs *ReminderService) processBookingReminder(booking BookingData) bool {
	log.Printf("📧 Processing booking %s", booking.BookingID)

	// Send reminder to user
	if err := rs.sendReminderToUser(booking); err != nil {
		log.Printf("❌ Failed to send reminder to user %s: %v", booking.UserID, err)
		return false
	}

	// Send reminder to expert
	if err := rs.sendReminderToExpert(booking); err != nil {
		log.Printf("❌ Failed to send reminder to expert %s: %v", booking.ExpertProfileID, err)
		return false
	}

	// Mark reminder as sent
	if err := rs.markReminderAsSent(booking.BookingID); err != nil {
		log.Printf("❌ Failed to update reminder_sent for booking %s: %v", booking.BookingID, err)
		return false
	}

	log.Printf("✅ Sent reminder for booking %s", booking.BookingID)
	return true
}

// markReminderAsSent updates the reminder_sent flag
func (rs *ReminderService) markReminderAsSent(bookingID string) error {
	return rs.db.Model(&entityBooking.ConsultationBooking{}).
		Where("booking_id = ?", bookingID).
		Update("reminder_sent", true).Error
}

// ===========================================
// USER REMINDER FUNCTIONALITY
// ===========================================

// sendReminderToUser sends reminder notification and email to user
func (rs *ReminderService) sendReminderToUser(booking BookingData) error {
	log.Printf("📧 Sending reminder to user: %s (%s)", booking.UserFullName, booking.UserEmail)

	userUUID, err := uuid.Parse(booking.UserID)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	// Create and save notification
	notification := rs.createUserNotification(userUUID, booking)
	if err := rs.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Send email through dispatcher - SYNCHRONIZED WITH EnhancedNotificationService
	if rs.dispatcher != nil {
		payload := rs.createUserEmailPayload(booking)
		if err := rs.dispatcher.DispatchNotification(JobTypeSendEmail, payload); err != nil {
			log.Printf("❌ Failed to dispatch email notification: %v", err)
			return fmt.Errorf("failed to dispatch email notification: %w", err)
		}
		log.Printf("✅ Email notification dispatched to user: %s", booking.UserEmail)
	} else {
		log.Printf("⚠️ Dispatcher is not configured")
	}

	return nil
}

// createUserNotification creates notification for user
func (rs *ReminderService) createUserNotification(userUUID uuid.UUID, booking BookingData) entityNotify.SystemNotification {
	message := fmt.Sprintf(
		"Bạn có lịch tư vấn với %s vào lúc %s. Loại tư vấn: %s",
		booking.ExpertFullName,
		booking.BookingDatetime.Format("15:04 02/01/2006"),
		booking.ConsultationType,
	)

	// Add meeting information
	if booking.MeetingLink != "" {
		message += fmt.Sprintf("\nLink tham gia: %s", booking.MeetingLink)
	} else if booking.MeetingAddress != "" {
		message += fmt.Sprintf("\nĐịa chỉ: %s", booking.MeetingAddress)
	}

	return entityNotify.SystemNotification{
		RecipientUserID:     userUUID,
		NotificationType:    "booking_reminder",
		NotificationTitle:   "Nhắc nhở: Lịch tư vấn sắp diễn ra",
		NotificationMessage: message,
		NotificationData: map[string]interface{}{
			"booking_id":        booking.BookingID,
			"expert_name":       booking.ExpertFullName,
			"booking_datetime":  booking.BookingDatetime,
			"consultation_type": booking.ConsultationType,
			"meeting_link":      booking.MeetingLink,
			"meeting_address":   booking.MeetingAddress,
		},
		DeliveryMethods: []string{"app", "email"},
	}
}

// createUserEmailPayload creates email payload for user reminder - SYNCHRONIZED
func (rs *ReminderService) createUserEmailPayload(booking BookingData) map[string]interface{} {
	return map[string]interface{}{
		"from":      "user", // Thêm field "from" để sync với EnhancedNotificationService
		"user_id":   booking.UserID,
		"recipient": booking.UserEmail,
		"subject":   "Nhắc nhở: Lịch tư vấn sắp diễn ra",
		"body":      rs.createUserEmailBody(booking),
		"template":  "booking_reminder",
		"data": map[string]interface{}{
			"booking_id":        booking.BookingID,
			"user_name":         booking.UserFullName,
			"user_email":        booking.UserEmail,
			"expert_id":         booking.ExpertProfileID,
			"expert_name":       booking.ExpertFullName,
			"expert_email":      booking.ExpertEmail,
			"consultation_date": booking.BookingDatetime.Format("02/01/2006"),
			"consultation_time": booking.BookingDatetime.Format("15:04"),
			"meeting_link":      booking.MeetingLink,
			"location":          booking.MeetingAddress,
			"consultation_type": booking.ConsultationType,
			"time_until":        "1 giờ",
		},
	}
}

// createUserEmailBody creates email body for user
func (rs *ReminderService) createUserEmailBody(booking BookingData) string {
	body := fmt.Sprintf("Xin chào %s,\n\nBạn có lịch tư vấn với %s vào lúc %s.\n\nLoại tư vấn: %s\n",
		booking.UserFullName,
		booking.ExpertFullName,
		booking.BookingDatetime.Format("15:04 02/01/2006"),
		booking.ConsultationType,
	)

	if booking.MeetingLink != "" {
		body += fmt.Sprintf("Link tham gia: %s\n", booking.MeetingLink)
	} else if booking.MeetingAddress != "" {
		body += fmt.Sprintf("Địa chỉ: %s\n", booking.MeetingAddress)
	}

	body += "\nVui lòng chuẩn bị sẵn sàng cho buổi tư vấn.\n\nTrân trọng,\nĐội ngũ CBS"

	return body
}

// ===========================================
// EXPERT REMINDER FUNCTIONALITY
// ===========================================

// sendReminderToExpert sends reminder notification and email to expert
func (rs *ReminderService) sendReminderToExpert(booking BookingData) error {
	log.Printf("📧 Sending reminder to expert: %s (%s)", booking.ExpertFullName, booking.ExpertEmail)

	expertUserID, err := rs.getExpertUserID(booking.ExpertProfileID)
	if err != nil {
		return fmt.Errorf("failed to get expert user ID: %w", err)
	}

	expertUserUUID, err := uuid.Parse(expertUserID)
	if err != nil {
		return fmt.Errorf("invalid expert user ID format: %w", err)
	}

	// Create and save notification
	notification := rs.createExpertNotification(expertUserUUID, booking)
	if err := rs.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to create notification for expert: %w", err)
	}

	// Send email through dispatcher - SYNCHRONIZED WITH EnhancedNotificationService
	if rs.dispatcher != nil {
		payload := rs.createExpertEmailPayload(booking, expertUserID)
		if err := rs.dispatcher.DispatchNotification(JobTypeSendEmail, payload); err != nil {
			log.Printf("❌ Failed to dispatch email notification to expert: %v", err)
			return fmt.Errorf("failed to dispatch email notification to expert: %w", err)
		}
		log.Printf("✅ Email notification dispatched to expert: %s", booking.ExpertEmail)
	} else {
		log.Printf("⚠️ Dispatcher is not configured")
	}

	return nil
}

// getExpertUserID retrieves the user ID for an expert profile
func (rs *ReminderService) getExpertUserID(expertProfileID string) (string, error) {
	var expertUserID string
	err := rs.db.Model(&entityExpert.ExpertProfile{}).
		Where("expert_profile_id = ?", expertProfileID).
		Select("user_id").
		Scan(&expertUserID).Error
	return expertUserID, err
}

// createExpertNotification creates notification for expert
func (rs *ReminderService) createExpertNotification(expertUserUUID uuid.UUID, booking BookingData) entityNotify.SystemNotification {
	message := fmt.Sprintf(
		"Bạn có lịch tư vấn với %s vào lúc %s. Loại tư vấn: %s",
		booking.UserFullName,
		booking.BookingDatetime.Format("15:04 02/01/2006"),
		booking.ConsultationType,
	)

	// Add meeting information
	if booking.MeetingLink != "" {
		message += fmt.Sprintf("\nLink tham gia: %s", booking.MeetingLink)
	} else if booking.MeetingAddress != "" {
		message += fmt.Sprintf("\nĐịa chỉ: %s", booking.MeetingAddress)
	}

	return entityNotify.SystemNotification{
		RecipientUserID:     expertUserUUID,
		NotificationType:    "booking_reminder",
		NotificationTitle:   "Nhắc nhở: Lịch tư vấn sắp diễn ra",
		NotificationMessage: message,
		NotificationData: map[string]interface{}{
			"booking_id":        booking.BookingID,
			"user_name":         booking.UserFullName,
			"booking_datetime":  booking.BookingDatetime,
			"consultation_type": booking.ConsultationType,
			"meeting_link":      booking.MeetingLink,
			"meeting_address":   booking.MeetingAddress,
		},
		DeliveryMethods: []string{"app", "email"},
	}
}

// createExpertEmailPayload creates email payload for expert reminder - SYNCHRONIZED
func (rs *ReminderService) createExpertEmailPayload(booking BookingData, expertUserID string) map[string]interface{} {
	return map[string]interface{}{
		"from":      "expert", // Thêm field "from" để sync với EnhancedNotificationService
		"user_id":   expertUserID,
		"recipient": booking.ExpertEmail,
		"subject":   "Nhắc nhở: Lịch tư vấn sắp diễn ra",
		"body":      rs.createExpertEmailBody(booking),
		"template":  "booking_reminder",
		"data": map[string]interface{}{
			"booking_id":        booking.BookingID,
			"user_name":         booking.UserFullName,
			"user_email":        booking.UserEmail,
			"expert_id":         booking.ExpertProfileID,
			"expert_name":       booking.ExpertFullName,
			"expert_email":      booking.ExpertEmail,
			"consultation_date": booking.BookingDatetime.Format("02/01/2006"),
			"consultation_time": booking.BookingDatetime.Format("15:04"),
			"meeting_link":      booking.MeetingLink,
			"location":          booking.MeetingAddress,
			"consultation_type": booking.ConsultationType,
			"time_until":        "1 giờ",
		},
	}
}

// createExpertEmailBody creates email body for expert
func (rs *ReminderService) createExpertEmailBody(booking BookingData) string {
	body := fmt.Sprintf("Xin chào %s,\n\nBạn có lịch tư vấn với khách hàng %s vào lúc %s.\n\nLoại tư vấn: %s\n",
		booking.ExpertFullName,
		booking.UserFullName,
		booking.BookingDatetime.Format("15:04 02/01/2006"),
		booking.ConsultationType,
	)

	if booking.MeetingLink != "" {
		body += fmt.Sprintf("Link tham gia: %s\n", booking.MeetingLink)
	} else if booking.MeetingAddress != "" {
		body += fmt.Sprintf("Địa chỉ: %s\n", booking.MeetingAddress)
	}

	body += "\nVui lòng chuẩn bị sẵn sàng cho buổi tư vấn.\n\nTrân trọng,\nĐội ngũ CBS"

	return body
}

// ===========================================
// BOOKING STATUS MANAGEMENT
// ===========================================

// CheckMissedBookings updates overdue bookings to missed status
func (rs *ReminderService) CheckMissedBookings() error {
	log.Println("🔍 Checking missed bookings...")

	result := rs.db.Model(&entityBooking.ConsultationBooking{}).
		Where("booking_status = ? AND booking_datetime < NOW() - INTERVAL '15 minutes'", "confirmed").
		Update("booking_status", "missed")

	if result.Error != nil {
		return fmt.Errorf("failed to update missed bookings: %w", result.Error)
	}

	log.Printf("✅ Updated %d missed bookings", result.RowsAffected)
	return nil
}

// ===========================================
// DUPLICATE BOOKING HANDLING
// ===========================================

// HandleDuplicateBookings resolves conflicting bookings
func (rs *ReminderService) HandleDuplicateBookings() error {
	log.Println("🔄 Handling duplicate bookings...")

	duplicates, err := rs.findDuplicateBookings()
	if err != nil {
		return fmt.Errorf("failed to find duplicate bookings: %w", err)
	}

	successCount := 0
	for _, dup := range duplicates {
		if err := rs.resolveDuplicateBooking(dup.ExpertProfileID, dup.BookingDatetime); err != nil {
			log.Printf("❌ Failed to resolve duplicate booking: %v", err)
		} else {
			successCount++
		}
	}

	log.Printf("✅ Resolved %d/%d duplicate booking groups", successCount, len(duplicates))
	return nil
}

// findDuplicateBookings finds bookings with same expert and time
func (rs *ReminderService) findDuplicateBookings() ([]DuplicateBooking, error) {
	var duplicates []DuplicateBooking

	query := `
		SELECT 
			expert_profile_id,
			booking_datetime,
			COUNT(*) as count
		FROM tbl_consultation_bookings 
		WHERE booking_status IN ('pending', 'confirmed')
		GROUP BY expert_profile_id, booking_datetime
		HAVING COUNT(*) > 1
	`

	return duplicates, rs.db.Raw(query).Scan(&duplicates).Error
}

// resolveDuplicateBooking resolves conflicts by keeping first booking
func (rs *ReminderService) resolveDuplicateBooking(expertProfileID string, bookingDatetime time.Time) error {
	// Get all conflicting bookings, ordered by creation time
	var bookings []entityBooking.ConsultationBooking
	if err := rs.db.Where("expert_profile_id = ? AND booking_datetime = ? AND booking_status IN ('pending', 'confirmed')",
		expertProfileID, bookingDatetime).
		Order("booking_created_at ASC").
		Find(&bookings).Error; err != nil {
		return err
	}

	// Keep first booking, cancel others
	for i, booking := range bookings {
		if i == 0 {
			// Confirm first booking if pending
			if booking.BookingStatus == "pending" {
				rs.db.Model(&booking).Update("booking_status", "confirmed")
			}
		} else {
			// Cancel other bookings
			rs.db.Model(&booking).Updates(map[string]interface{}{
				"booking_status":      "cancelled",
				"cancellation_reason": "Tự động hủy do trùng lịch",
			})

			// Send cancellation notification
			rs.sendCancellationNotification(
				booking.UserID.String(),
				booking.BookingID.String(),
				"Lịch tư vấn đã bị hủy do trùng lịch. Vui lòng đặt lại lịch khác.",
			)
		}
	}

	return nil
}

// sendCancellationNotification sends notification for cancelled booking
func (rs *ReminderService) sendCancellationNotification(userID, bookingID, reason string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	notification := entityNotify.SystemNotification{
		RecipientUserID:     userUUID,
		NotificationType:    "booking_cancelled",
		NotificationTitle:   "Lịch tư vấn đã bị hủy",
		NotificationMessage: reason,
		NotificationData: map[string]interface{}{
			"booking_id": bookingID,
			"reason":     reason,
		},
		DeliveryMethods: []string{"app", "email"},
	}

	return rs.db.Create(&notification).Error
}

// ===========================================
// STATISTICS GENERATION
// ===========================================

// GenerateWeeklyStatistics creates weekly booking statistics
func (rs *ReminderService) GenerateWeeklyStatistics() error {
	log.Println("📊 Generating weekly statistics...")

	weekStart := time.Now().AddDate(0, 0, -7)
	stats, err := rs.calculateWeeklyStats(weekStart)
	if err != nil {
		return fmt.Errorf("failed to calculate weekly statistics: %w", err)
	}

	if err := rs.saveWeeklyStats(stats, weekStart); err != nil {
		return fmt.Errorf("failed to save weekly statistics: %w", err)
	}

	rs.logWeeklyStats(stats)
	return nil
}

// calculateWeeklyStats calculates statistics for the given week
func (rs *ReminderService) calculateWeeklyStats(weekStart time.Time) (*WeeklyStats, error) {
	stats := &WeeklyStats{}

	// Total bookings
	if err := rs.db.Model(&entityBooking.ConsultationBooking{}).
		Where("booking_created_at >= ?", weekStart).
		Count(&stats.TotalBookings).Error; err != nil {
		return nil, err
	}

	// Completed bookings
	if err := rs.db.Model(&entityBooking.ConsultationBooking{}).
		Where("booking_created_at >= ? AND booking_status = ?", weekStart, "completed").
		Count(&stats.CompletedBookings).Error; err != nil {
		return nil, err
	}

	// Cancelled bookings
	if err := rs.db.Model(&entityBooking.ConsultationBooking{}).
		Where("booking_created_at >= ? AND booking_status = ?", weekStart, "cancelled").
		Count(&stats.CancelledBookings).Error; err != nil {
		return nil, err
	}

	// Missed bookings
	if err := rs.db.Model(&entityBooking.ConsultationBooking{}).
		Where("booking_created_at >= ? AND booking_status = ?", weekStart, "missed").
		Count(&stats.MissedBookings).Error; err != nil {
		return nil, err
	}

	// Total revenue
	if err := rs.db.Model(&entityBooking.ConsultationBooking{}).
		Where("booking_created_at >= ? AND booking_status = ? AND payment_status = ?", weekStart, "completed", "paid").
		Select("COALESCE(SUM(consultation_fee), 0)").
		Scan(&stats.RevenueTotal).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// saveWeeklyStats saves statistics to system settings
func (rs *ReminderService) saveWeeklyStats(stats *WeeklyStats, weekStart time.Time) error {
	completionRate := 0.0
	if stats.TotalBookings > 0 {
		completionRate = float64(stats.CompletedBookings) / float64(stats.TotalBookings) * 100
	}

	statsData := map[string]interface{}{
		"period":             "weekly",
		"week_start":         weekStart,
		"week_end":           time.Now(),
		"total_bookings":     stats.TotalBookings,
		"completed_bookings": stats.CompletedBookings,
		"cancelled_bookings": stats.CancelledBookings,
		"missed_bookings":    stats.MissedBookings,
		"revenue_total":      stats.RevenueTotal,
		"completion_rate":    completionRate,
		"generated_at":       time.Now(),
	}

	setting := entitySystem.SystemSetting{
		SettingKey:         fmt.Sprintf("weekly_stats_%s", weekStart.Format("2006_01_02")),
		SettingValue:       statsData,
		SettingDescription: fmt.Sprintf("Weekly statistics for week starting %s", weekStart.Format("2006-01-02")),
	}

	return rs.db.Create(&setting).Error
}

// logWeeklyStats logs the generated statistics
func (rs *ReminderService) logWeeklyStats(stats *WeeklyStats) {
	completionRate := 0.0
	if stats.TotalBookings > 0 {
		completionRate = float64(stats.CompletedBookings) / float64(stats.TotalBookings) * 100
	}

	log.Printf("✅ Weekly statistics generated: %d total bookings, %d completed, %.2f%% completion rate, $%.2f revenue",
		stats.TotalBookings, stats.CompletedBookings, completionRate, stats.RevenueTotal)
}
