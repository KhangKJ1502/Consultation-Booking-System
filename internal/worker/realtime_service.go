package worker

import (
	entityNotify "cbs_backend/internal/modules/system_notification/entity"
	entityUser "cbs_backend/internal/modules/users/entity"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Constants
const (
	// Notification types
	NotificationTypeBookingCreated   = "booking_created"
	NotificationTypeBookingCancelled = "booking_cancelled"
	NotificationTypeBookingReminder  = "booking_reminder"
	NotificationTypeBookingConfirmed = "booking_confirmed"

	// Redis configuration
	UserChannelPrefix   = "user:"
	RedisPublishTimeout = 5 * time.Second

	// Default pagination
	DefaultLimit = 20
)

// Structs
type RealtimeService struct {
	db          *gorm.DB
	redisClient *redis.Client
	scheduler   *WorkerScheduler
	ctx         context.Context
}

type RealtimeMessage struct {
	Type      string                 `json:"type"`
	UserID    string                 `json:"user_id"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

type NotificationTemplate struct {
	Title   string
	Message string
}

type Job struct {
	ID         string
	Type       string
	Payload    interface{}
	Priority   int
	RetryCount int
	MaxRetries int
	CreatedAt  time.Time
}

// Notification templates
var notificationTemplates = map[string]NotificationTemplate{
	NotificationTypeBookingCreated: {
		Title:   "Lịch hẹn mới được tạo",
		Message: "Bạn có một lịch hẹn mới được đặt",
	},
	NotificationTypeBookingCancelled: {
		Title:   "Lịch hẹn bị hủy",
		Message: "Lịch hẹn của bạn đã bị hủy",
	},
	NotificationTypeBookingReminder: {
		Title:   "Nhắc nhở lịch hẹn",
		Message: "Lịch hẹn của bạn sẽ diễn ra trong 60 phút",
	},
	NotificationTypeBookingConfirmed: {
		Title:   "Lịch hẹn được xác nhận",
		Message: "Lịch hẹn của bạn đã được xác nhận",
	},
}

// Constructor
func NewRealtimeService(db *gorm.DB, redisClient *redis.Client, scheduler *WorkerScheduler) *RealtimeService {
	return &RealtimeService{
		db:          db,
		redisClient: redisClient,
		scheduler:   scheduler,
		ctx:         context.Background(),
	}
}

// Setter methods
func (rs *RealtimeService) SetScheduler(scheduler *WorkerScheduler) {
	rs.scheduler = scheduler
}

// Core messaging methods
func (rs *RealtimeService) PublishToUser(userID string, message RealtimeMessage) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	channel := fmt.Sprintf("%s%s", UserChannelPrefix, userID)
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal realtime message: %w", err)
	}

	ctx, cancel := context.WithTimeout(rs.ctx, RedisPublishTimeout)
	defer cancel()

	result := rs.redisClient.Publish(ctx, channel, messageBytes)
	if result.Err() != nil {
		return fmt.Errorf("failed to publish message to Redis: %w", result.Err())
	}

	log.Printf("Published realtime message to channel %s, subscribers: %d", channel, result.Val())
	return nil
}

func (rs *RealtimeService) SendBookingNotification(bookingID, userID, expertID, notificationType string, data map[string]interface{}) error {
	if err := rs.validateBookingNotificationParams(bookingID, userID, notificationType); err != nil {
		return err
	}

	// Check if user has notifications enabled for this type
	if !rs.isNotificationEnabled(userID, notificationType) {
		log.Printf("Notification disabled for user %s, type %s", userID, notificationType)
		return nil
	}

	template := notificationTemplates[notificationType]
	userUUID, _ := uuid.Parse(userID) // Already validated

	// Prepare notification data
	if data == nil {
		data = make(map[string]interface{})
	}
	data["booking_id"] = bookingID
	data["expert_id"] = expertID
	data["user_id"] = userID

	// Create notification record
	notification := entityNotify.SystemNotification{
		RecipientUserID:       userUUID,
		NotificationType:      notificationType,
		NotificationTitle:     template.Title,
		NotificationMessage:   template.Message,
		NotificationData:      data,
		DeliveryMethods:       rs.getDeliveryMethods(userID),
		NotificationCreatedAt: time.Now(),
		IsRead:                false,
	}

	if err := rs.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to create notification record: %w", err)
	}

	// Send real-time notification
	realtimeMsg := RealtimeMessage{
		Type:      notificationType,
		UserID:    userID,
		Title:     template.Title,
		Message:   template.Message,
		Data:      data,
		Timestamp: time.Now(),
	}

	if err := rs.PublishToUser(userID, realtimeMsg); err != nil {
		log.Printf("Failed to send real-time notification to user %s: %v", userID, err)
	}

	// Process other delivery methods asynchronously
	go rs.processDeliveryMethods(notification)

	return nil
}

func (rs *RealtimeService) SendBulkNotifications(userIDs []string, notificationType string, data map[string]interface{}) error {
	if len(userIDs) == 0 {
		return fmt.Errorf("user IDs list cannot be empty")
	}

	if _, exists := notificationTemplates[notificationType]; !exists {
		return fmt.Errorf("invalid notification type: %s", notificationType)
	}

	var errors []string
	for _, userID := range userIDs {
		if err := rs.SendBookingNotification("", userID, "", notificationType, data); err != nil {
			errors = append(errors, fmt.Sprintf("user %s: %v", userID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send notifications to some users: %v", errors)
	}

	return nil
}

// Notification management methods
func (rs *RealtimeService) MarkNotificationAsRead(notificationID, userID string) error {
	if notificationID == "" || userID == "" {
		return fmt.Errorf("notification ID and user ID cannot be empty")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	notificationUUID, err := uuid.Parse(notificationID)
	if err != nil {
		return fmt.Errorf("invalid notification ID format: %w", err)
	}

	result := rs.db.Model(&entityNotify.SystemNotification{}).
		Where("notification_id = ? AND recipient_user_id = ?", notificationUUID, userUUID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to mark notification as read: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("notification not found or already read")
	}

	return nil
}

func (rs *RealtimeService) GetUserNotifications(userID string, limit, offset int) ([]entityNotify.SystemNotification, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	if limit <= 0 {
		limit = DefaultLimit
	}
	if offset < 0 {
		offset = 0
	}

	var notifications []entityNotify.SystemNotification
	err = rs.db.Where("recipient_user_id = ?", userUUID).
		Order("notification_created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user notifications: %w", err)
	}

	return notifications, nil
}

func (rs *RealtimeService) GetUnreadNotificationCount(userID string) (int64, error) {
	if userID == "" {
		return 0, fmt.Errorf("user ID cannot be empty")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID format: %w", err)
	}

	var count int64
	err = rs.db.Model(&entityNotify.SystemNotification{}).
		Where("recipient_user_id = ? AND is_read = ?", userUUID, false).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to get unread notification count: %w", err)
	}

	return count, nil
}

// Helper methods
func (rs *RealtimeService) validateBookingNotificationParams(bookingID, userID, notificationType string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	if bookingID == "" {
		return fmt.Errorf("booking ID cannot be empty")
	}
	if _, exists := notificationTemplates[notificationType]; !exists {
		return fmt.Errorf("invalid notification type: %s", notificationType)
	}
	if _, err := uuid.Parse(userID); err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}
	return nil
}

func (rs *RealtimeService) isNotificationEnabled(userID, notificationType string) bool {
	if userID == "" || notificationType == "" {
		return false
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Invalid user ID format: %v", err)
		return false
	}

	var setting entityUser.UserNotificationSetting
	err = rs.db.Where("user_id = ? AND notification_type = ?", userUUID, notificationType).
		First(&setting).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return true // Default to enabled
		}
		log.Printf("Error checking notification settings: %v", err)
		return true
	}

	return setting.IsEnabled
}

func (rs *RealtimeService) getDeliveryMethods(userID string) []string {
	if userID == "" {
		return []string{"app"}
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Invalid user ID format: %v", err)
		return []string{"app"}
	}

	var settings []entityUser.UserNotificationSetting
	err = rs.db.Where("user_id = ? AND is_enabled = ?", userUUID, true).
		Find(&settings).Error

	if err != nil {
		log.Printf("Error getting delivery methods: %v", err)
		return []string{"app"}
	}

	methods := []string{"app"}
	methodSet := map[string]bool{"app": true}

	for _, setting := range settings {
		if !methodSet[setting.DeliveryMethod] {
			methods = append(methods, setting.DeliveryMethod)
			methodSet[setting.DeliveryMethod] = true
		}
	}

	return methods
}

func (rs *RealtimeService) processDeliveryMethods(notification entityNotify.SystemNotification) {
	for _, method := range notification.DeliveryMethods {
		switch method {
		case "email":
			if err := rs.sendEmailNotification(notification); err != nil {
				log.Printf("Failed to send email notification: %v", err)
			}
		case "telegram":
			if err := rs.sendTelegramNotification(notification); err != nil {
				log.Printf("Failed to send telegram notification: %v", err)
			}
		case "sms":
			if err := rs.sendSMSNotification(notification); err != nil {
				log.Printf("Failed to send SMS notification: %v", err)
			}
		case "app":
			// App notification already handled by real-time publish
			continue
		default:
			log.Printf("Unknown delivery method: %s", method)
		}
	}
}

// Delivery method implementations
func (rs *RealtimeService) sendEmailNotification(notification entityNotify.SystemNotification) error {
	var user struct {
		UserEmail string `json:"user_email"`
		FullName  string `json:"full_name"`
	}

	err := rs.db.Table("tbl_users").
		Where("user_id = ?", notification.RecipientUserID).
		Select("user_email, full_name").
		Scan(&user).Error

	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	if user.UserEmail == "" {
		return fmt.Errorf("user email is empty")
	}

	emailData := make(map[string]interface{})
	if notification.NotificationData != nil {
		emailData = notification.NotificationData
	}
	emailData["user_name"] = user.FullName
	emailData["user_id"] = notification.RecipientUserID.String()

	emailJob := Job{
		ID:         generateJobID(),
		Type:       JobTypeSendEmail,
		Priority:   2,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
		Payload: map[string]interface{}{
			"user_id":   notification.RecipientUserID.String(),
			"recipient": user.UserEmail,
			"subject":   notification.NotificationTitle,
			"body":      notification.NotificationMessage,
			"template":  "notification",
			"data":      emailData,
		},
	}

	return rs.addJobToQueue(emailJob)
}

func (rs *RealtimeService) sendTelegramNotification(notification entityNotify.SystemNotification) error {
	var telegramChatID string
	err := rs.db.Table("tbl_user_settings").
		Where("user_id = ? AND setting_key = ?", notification.RecipientUserID, "telegram_chat_id").
		Select("setting_value").
		Scan(&telegramChatID).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("telegram chat ID not found for user")
		}
		return fmt.Errorf("failed to get telegram chat ID: %w", err)
	}

	if telegramChatID == "" {
		return fmt.Errorf("telegram chat ID is empty")
	}

	telegramJob := Job{
		ID:         generateJobID(),
		Type:       JobTypeSendTelegram,
		Priority:   2,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
		Payload: map[string]interface{}{
			"user_id": notification.RecipientUserID.String(),
			"chat_id": telegramChatID,
			"message": fmt.Sprintf("🔔 %s\n\n%s", notification.NotificationTitle, notification.NotificationMessage),
			"data":    notification.NotificationData,
		},
	}

	return rs.addJobToQueue(telegramJob)
}

func (rs *RealtimeService) sendSMSNotification(notification entityNotify.SystemNotification) error {
	var phoneNumber string
	err := rs.db.Table("tbl_users").
		Where("user_id = ?", notification.RecipientUserID).
		Select("phone_number").
		Scan(&phoneNumber).Error

	if err != nil {
		return fmt.Errorf("failed to get user phone number: %w", err)
	}

	if phoneNumber == "" {
		return fmt.Errorf("user phone number is empty")
	}

	smsJob := Job{
		ID:         generateJobID(),
		Type:       JobTypeSendSMS,
		Priority:   2,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
		Payload: map[string]interface{}{
			"user_id":      notification.RecipientUserID.String(),
			"phone_number": phoneNumber,
			"message":      fmt.Sprintf("%s: %s", notification.NotificationTitle, notification.NotificationMessage),
		},
	}

	return rs.addJobToQueue(smsJob)
}

func (rs *RealtimeService) addJobToQueue(job Job) error {
	if rs.scheduler == nil {
		return fmt.Errorf("scheduler is not initialized")
	}
	rs.scheduler.AddJob(job)
	return nil
}

// Note: The following functions are referenced but not defined in the original code
// You'll need to implement these based on your job system:
// - generateJobID() string
// - JobTypeSendEmail, JobTypeSendTelegram, JobTypeSendSMS constants
