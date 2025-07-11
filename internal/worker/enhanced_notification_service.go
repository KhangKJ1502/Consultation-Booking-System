package worker

import (
	entityNotify "cbs_backend/internal/modules/system_notification/entity"
	"cbs_backend/internal/service/interfaces"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ==================== CONSTANTS ====================

// Các loại job notification
const (
	JobTypeSendEmail               = "send_email"
	JobTypeSendTelegram            = "send_telegram"
	JobTypeSendSMS                 = "send_sms"
	JobTypeCleanupOldNotifications = "cleanup_old_notifications"
)

// Các phương thức gửi thông báo
const (
	DeliveryMethodEmail    = "email"
	DeliveryMethodTelegram = "telegram"
	DeliveryMethodSMS      = "sms"
)

// Trạng thái gửi thông báo
const (
	DeliveryStatusPending = "pending"
	DeliveryStatusSent    = "sent"
	DeliveryStatusFailed  = "failed"
)

// Cấu hình retry và cleanup
const (
	MaxRetries      = 3                   // Số lần thử lại tối đa
	RetryBaseDelay  = time.Second         // Thời gian delay base cho retry
	NotificationTTL = 3 * 24 * time.Hour  // Thời gian sống của notification (3 ngày)
	HistoryTTL      = 30 * 24 * time.Hour // Thời gian sống của history (30 ngày)
)

// ==================== STRUCTS ====================

// Service chính xử lý notification
type EnhancedNotificationService struct {
	db          *gorm.DB
	redisClient *redis.Client
	realtimeSvc *RealtimeService
	emailSvc    interfaces.EmailService
	ctx         context.Context
}

// Payload cho email - cải thiện để tương thích với ReminderService
type EmailPayload struct {
	From      string                 `json:"from"` // "user" hoặc "expert"
	UserID    string                 `json:"user_id" validate:"required,uuid"`
	Recipient string                 `json:"recipient" validate:"required,email"`
	Subject   string                 `json:"subject" validate:"required,min=1,max=200"`
	Body      string                 `json:"body" validate:"required,min=1"`
	Template  string                 `json:"template,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// Payload cho Telegram
type TelegramPayload struct {
	UserID  string                 `json:"user_id" validate:"required,uuid"`
	ChatID  string                 `json:"chat_id" validate:"required"`
	Message string                 `json:"message" validate:"required,min=1,max=4096"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// Payload cho SMS
type SMSPayload struct {
	UserID      string `json:"user_id" validate:"required,uuid"`
	PhoneNumber string `json:"phone_number" validate:"required,min=10,max=15"`
	Message     string `json:"message" validate:"required,min=1,max=160"`
}

// ==================== CONSTRUCTOR ====================

// Tạo instance mới của NotificationService
func NewEnhancedNotificationService(db *gorm.DB, redisClient *redis.Client, emailService interfaces.EmailService) *EnhancedNotificationService {
	return &EnhancedNotificationService{
		db:          db,
		redisClient: redisClient,
		realtimeSvc: NewRealtimeService(db, redisClient, nil),
		emailSvc:    emailService,
		ctx:         context.Background(),
	}
}

// Set realtime service sau khi khởi tạo
func (ens *EnhancedNotificationService) SetRealtimeService(realtimeSvc *RealtimeService) {
	ens.realtimeSvc = realtimeSvc
}

// ==================== NOTIFICATION DISPATCHER IMPLEMENTATION ====================

// DispatchNotification implements NotificationDispatcher interface
func (ens *EnhancedNotificationService) DispatchNotification(jobType string, payload interface{}) error {
	job := Job{
		Type:    jobType,
		Payload: payload,
	}

	return ens.ProcessNotificationJob(job)
}

// ==================== MAIN PROCESSOR ====================

// Xử lý job notification theo loại
func (ens *EnhancedNotificationService) ProcessNotificationJob(job Job) error {
	switch job.Type {
	case JobTypeSendEmail:
		return ens.processSendEmail(job)
	case JobTypeSendTelegram:
		return ens.processSendTelegram(job)
	case JobTypeSendSMS:
		return ens.processSendSMS(job)
	case JobTypeCleanupOldNotifications:
		return ens.cleanupOldNotifications()
	default:
		return fmt.Errorf("loại job không hỗ trợ: %s", job.Type)
	}
}

// ==================== EMAIL PROCESSING ====================

// Xử lý gửi email
func (ens *EnhancedNotificationService) processSendEmail(job Job) error {
	// Parse payload từ job
	payload, err := ens.parseEmailPayload(job.Payload)
	if err != nil {
		return fmt.Errorf("lỗi parse email payload: %w", err)
	}

	// Tạo record history để theo dõi
	history, err := ens.createNotificationHistory(payload.UserID, DeliveryMethodEmail, map[string]interface{}{
		"recipient": payload.Recipient,
		"subject":   payload.Subject,
		"from":      payload.From,
	})
	if err != nil {
		return fmt.Errorf("lỗi tạo notification history: %w", err)
	}

	// Gửi email với retry logic
	err = ens.sendEmailWithRetry(payload, &history)
	if err != nil {
		ens.updateHistoryFailed(&history, err)
		return fmt.Errorf("lỗi gửi email: %w", err)
	}

	ens.updateHistorySuccess(&history)
	return nil
}

// Gửi email với retry logic - cải thiện để xử lý cả user và expert
func (ens *EnhancedNotificationService) sendEmailWithRetry(payload *EmailPayload, history *entityNotify.NotificationHistory) error {
	var lastErr error
	log.Printf("🔄 Đang xử lý email payload: %+v", payload)

	// Chuyển đổi payload.Data thành ConsultationReminderData
	reminderData := interfaces.ConsultationReminderData{
		BookingID:        getString(payload.Data, "booking_id"),
		UserID:           payload.UserID,
		UserName:         getString(payload.Data, "user_name"),
		UserEmail:        getString(payload.Data, "user_email"),
		ExpertID:         getString(payload.Data, "expert_id"),
		ExpertName:       getString(payload.Data, "expert_name"),
		ExpertEmail:      getString(payload.Data, "expert_email"),
		ConsultationDate: getString(payload.Data, "consultation_date"),
		ConsultationTime: getString(payload.Data, "consultation_time"),
		MeetingLink:      getString(payload.Data, "meeting_link"),
		Location:         getString(payload.Data, "location"),
		ConsultationType: getString(payload.Data, "consultation_type"),
		TimeUntil:        getString(payload.Data, "time_until"),
	}

	// Thử gửi email với retry
	for attempt := 1; attempt <= MaxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(ens.ctx, 30*time.Second)

		var err error
		switch payload.From {
		case "user":
			err = ens.emailSvc.SendConsultationBookingRemindersToUser(ctx, payload.UserID, reminderData)
		case "expert":
			err = ens.emailSvc.SendConsultationBookingRemindersToExpert(ctx, payload.UserID, reminderData)
		default:
			// Fallback: gửi theo recipient
			if payload.Recipient == reminderData.UserEmail {
				err = ens.emailSvc.SendConsultationBookingRemindersToUser(ctx, payload.UserID, reminderData)
			} else {
				err = ens.emailSvc.SendConsultationBookingRemindersToExpert(ctx, payload.UserID, reminderData)
			}
		}

		cancel()

		if err == nil {
			log.Printf("✅ Email gửi thành công đến %s ở lần thử %d", payload.Recipient, attempt)
			return nil
		}

		lastErr = err
		log.Printf("❌ Gửi email lần %d thất bại cho user %s: %v", attempt, payload.UserID, err)

		// Delay trước khi retry
		if attempt < MaxRetries {
			backoffDelay := RetryBaseDelay * time.Duration(attempt)
			time.Sleep(backoffDelay)
		}
	}

	return fmt.Errorf("gửi email thất bại sau %d lần thử: %w", MaxRetries, lastErr)
}

// ==================== TELEGRAM PROCESSING ====================

// Xử lý gửi Telegram
func (ens *EnhancedNotificationService) processSendTelegram(job Job) error {
	// Parse payload từ job
	payload, err := ens.parseTelegramPayload(job.Payload)
	if err != nil {
		return fmt.Errorf("lỗi parse telegram payload: %w", err)
	}

	// Tạo record history để theo dõi
	history, err := ens.createNotificationHistory(payload.UserID, DeliveryMethodTelegram, map[string]interface{}{
		"chat_id": payload.ChatID,
		"message": payload.Message,
	})
	if err != nil {
		return fmt.Errorf("lỗi tạo notification history: %w", err)
	}

	// TODO: Implement telegram sending logic
	log.Printf("📱 Sending Telegram message to %s: %s", payload.ChatID, payload.Message)

	// Simulate success for now
	ens.updateHistorySuccess(&history)
	return nil
}

// ==================== SMS PROCESSING ====================

// Xử lý gửi SMS
func (ens *EnhancedNotificationService) processSendSMS(job Job) error {
	// Parse payload từ job
	payload, err := ens.parseSMSPayload(job.Payload)
	if err != nil {
		return fmt.Errorf("lỗi parse SMS payload: %w", err)
	}

	// Tạo record history để theo dõi
	history, err := ens.createNotificationHistory(payload.UserID, DeliveryMethodSMS, map[string]interface{}{
		"phone_number": payload.PhoneNumber,
		"message":      payload.Message,
	})
	if err != nil {
		return fmt.Errorf("lỗi tạo notification history: %w", err)
	}

	// TODO: Implement SMS sending logic
	log.Printf("📞 Sending SMS to %s: %s", payload.PhoneNumber, payload.Message)

	// Simulate success for now
	ens.updateHistorySuccess(&history)
	return nil
}

// ==================== HISTORY MANAGEMENT ====================

// Tạo record history để theo dõi notification
func (ens *EnhancedNotificationService) createNotificationHistory(userID, method string, data map[string]interface{}) (entityNotify.NotificationHistory, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return entityNotify.NotificationHistory{}, fmt.Errorf("format user ID không hợp lệ: %w", err)
	}

	history := entityNotify.NotificationHistory{
		UserID:         userUUID,
		DeliveryMethod: method,
		DeliveryStatus: DeliveryStatusPending,
		CreatedAt:      time.Now(),
		DeliveredAt:    nil,
	}

	if err := ens.db.Create(&history).Error; err != nil {
		return entityNotify.NotificationHistory{}, fmt.Errorf("lỗi tạo history record: %w", err)
	}

	return history, nil
}

// Cập nhật history khi gửi thành công
func (ens *EnhancedNotificationService) updateHistorySuccess(history *entityNotify.NotificationHistory) {
	now := time.Now()
	updates := map[string]interface{}{
		"delivery_status": DeliveryStatusSent,
		"delivered_at":    &now,
		"updated_at":      now,
	}

	if err := ens.db.Model(history).Updates(updates).Error; err != nil {
		log.Printf("❌ Lỗi cập nhật history thành công: %v", err)
	}
}

// Cập nhật history khi gửi thất bại
func (ens *EnhancedNotificationService) updateHistoryFailed(history *entityNotify.NotificationHistory, err error) {
	updates := map[string]interface{}{
		"delivery_status": DeliveryStatusFailed,
		"error_message":   err.Error(),
		"updated_at":      time.Now(),
	}

	if updateErr := ens.db.Model(history).Updates(updates).Error; updateErr != nil {
		log.Printf("❌ Lỗi cập nhật history thất bại: %v", updateErr)
	}
}

// ==================== PAYLOAD PARSING ====================

// Parse email payload từ job - cải thiện để tương thích với ReminderService
func (ens *EnhancedNotificationService) parseEmailPayload(payload interface{}) (*EmailPayload, error) {
	data, err := ens.parseToMap(payload)
	if err != nil {
		return nil, err
	}

	result := &EmailPayload{}

	// Validate và set các field bắt buộc
	if err := ens.setRequiredField(data, "user_id", &result.UserID, ens.validateUUID); err != nil {
		return nil, err
	}
	if err := ens.setRequiredField(data, "recipient", &result.Recipient, nil); err != nil {
		return nil, err
	}
	if err := ens.setRequiredField(data, "subject", &result.Subject, nil); err != nil {
		return nil, err
	}
	if err := ens.setRequiredField(data, "body", &result.Body, nil); err != nil {
		return nil, err
	}

	// Set các field optional
	if from, ok := data["from"].(string); ok {
		result.From = from
	}
	if template, ok := data["template"].(string); ok {
		result.Template = template
	}
	if emailData, ok := data["data"].(map[string]interface{}); ok {
		result.Data = emailData
	}

	return result, nil
}

// Parse telegram payload từ job
func (ens *EnhancedNotificationService) parseTelegramPayload(payload interface{}) (*TelegramPayload, error) {
	data, err := ens.parseToMap(payload)
	if err != nil {
		return nil, err
	}

	result := &TelegramPayload{}

	// Validate và set các field bắt buộc
	if err := ens.setRequiredField(data, "user_id", &result.UserID, ens.validateUUID); err != nil {
		return nil, err
	}
	if err := ens.setRequiredField(data, "chat_id", &result.ChatID, nil); err != nil {
		return nil, err
	}
	if err := ens.setRequiredField(data, "message", &result.Message, nil); err != nil {
		return nil, err
	}

	// Set field optional
	if telegramData, ok := data["data"].(map[string]interface{}); ok {
		result.Data = telegramData
	}

	return result, nil
}

// Parse SMS payload từ job
func (ens *EnhancedNotificationService) parseSMSPayload(payload interface{}) (*SMSPayload, error) {
	data, err := ens.parseToMap(payload)
	if err != nil {
		return nil, err
	}

	result := &SMSPayload{}

	// Validate và set các field bắt buộc
	if err := ens.setRequiredField(data, "user_id", &result.UserID, ens.validateUUID); err != nil {
		return nil, err
	}
	if err := ens.setRequiredField(data, "phone_number", &result.PhoneNumber, nil); err != nil {
		return nil, err
	}
	if err := ens.setRequiredField(data, "message", &result.Message, nil); err != nil {
		return nil, err
	}

	return result, nil
}

// ==================== HELPER METHODS ====================

// Parse payload thành map
func (ens *EnhancedNotificationService) parseToMap(payload interface{}) (map[string]interface{}, error) {
	switch v := payload.(type) {
	case map[string]interface{}:
		return v, nil
	case string:
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(v), &data); err != nil {
			return nil, fmt.Errorf("lỗi unmarshal JSON payload: %w", err)
		}
		return data, nil
	default:
		return nil, fmt.Errorf("format payload không hợp lệ: cần map hoặc JSON string")
	}
}

// Set field bắt buộc với validation
func (ens *EnhancedNotificationService) setRequiredField(data map[string]interface{}, key string, target *string, validator func(string) error) error {
	if value, ok := data[key].(string); ok && value != "" {
		if validator != nil {
			if err := validator(value); err != nil {
				return fmt.Errorf("field %s không hợp lệ: %w", key, err)
			}
		}
		*target = value
		return nil
	}
	return fmt.Errorf("field %s bị thiếu hoặc không hợp lệ", key)
}

// Validate UUID format
func (ens *EnhancedNotificationService) validateUUID(value string) error {
	if _, err := uuid.Parse(value); err != nil {
		return fmt.Errorf("format UUID không hợp lệ: %w", err)
	}
	return nil
}

// ==================== CLEANUP ====================

// Dọn dẹp các notification và history cũ
func (ens *EnhancedNotificationService) cleanupOldNotifications() error {
	log.Println("🧹 Bắt đầu dọn dẹp notification cũ...")

	return ens.db.Transaction(func(tx *gorm.DB) error {
		// Dọn dẹp notification đã đọc > 3 ngày
		notificationCutoff := time.Now().Add(-NotificationTTL)
		notificationResult := tx.Where("notification_created_at < ? AND is_read = ?", notificationCutoff, true).
			Delete(&entityNotify.SystemNotification{})

		if notificationResult.Error != nil {
			return fmt.Errorf("lỗi dọn dẹp notification cũ: %w", notificationResult.Error)
		}

		log.Printf("✅ Đã dọn dẹp %d notification cũ", notificationResult.RowsAffected)

		// Dọn dẹp history > 30 ngày
		historyCutoff := time.Now().Add(-HistoryTTL)
		historyResult := tx.Where("created_at < ?", historyCutoff).
			Delete(&entityNotify.NotificationHistory{})

		if historyResult.Error != nil {
			return fmt.Errorf("lỗi dọn dẹp notification history: %w", historyResult.Error)
		}

		log.Printf("✅ Đã dọn dẹp %d history cũ", historyResult.RowsAffected)

		return nil
	})
}

// ==================== VALIDATION METHODS ====================

// ValidateEmailPayload validates email payload structure
func (ens *EnhancedNotificationService) ValidateEmailPayload(payload map[string]interface{}) error {
	_, err := ens.parseEmailPayload(payload)
	return err
}

// ValidateTelegramPayload validates telegram payload structure
func (ens *EnhancedNotificationService) ValidateTelegramPayload(payload map[string]interface{}) error {
	_, err := ens.parseTelegramPayload(payload)
	return err
}

// ValidateSMSPayload validates SMS payload structure
func (ens *EnhancedNotificationService) ValidateSMSPayload(payload map[string]interface{}) error {
	_, err := ens.parseSMSPayload(payload)
	return err
}
