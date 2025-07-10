package worker

import (
	entityActity "cbs_backend/internal/modules/activity_logs/entity"
	entityBackground "cbs_backend/internal/modules/background_job/entity"
	entityBooking "cbs_backend/internal/modules/bookings/entity"
	entityNotfy "cbs_backend/internal/modules/system_notification/entity"
	entityUser "cbs_backend/internal/modules/users/entity"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type CleanupService struct {
	db *gorm.DB
}

func NewCleanupService(db *gorm.DB) *CleanupService {
	return &CleanupService{db: db}
}

func (cs *CleanupService) CleanupOldData(days int) error {
	log.Printf("🧹 Cleaning up data older than %d days...", days)

	cutoffDate := time.Now().AddDate(0, 0, -days)

	// Cleanup old activity logs
	result := cs.db.Where("log_created_at < ?", cutoffDate).Delete(&entityActity.ActivityLog{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup activity logs: %w", result.Error)
	}
	log.Printf("✅ Cleaned up %d old activity logs", result.RowsAffected)

	// Cleanup old notifications (read notifications older than cutoff)
	result = cs.db.Where("notification_created_at < ? AND is_read = ?", cutoffDate, true).Delete(&entityNotfy.SystemNotification{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup notifications: %w", result.Error)
	}
	log.Printf("✅ Cleaned up %d old notifications", result.RowsAffected)

	// Cleanup expired user sessions
	result = cs.db.Where("expires_at < ?", time.Now()).Delete(&entityUser.UserSession{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", result.Error)
	}
	log.Printf("✅ Cleaned up %d expired sessions", result.RowsAffected)

	// Cleanup revoked refresh tokens
	result = cs.db.Where("expires_at < ? OR is_revoked = ? AND token_type = refresh", time.Now(), true).Delete(&entityUser.UserToken{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup refresh tokens: %w", result.Error)
	}
	log.Printf("✅ Cleaned up %d expired/revoked refresh tokens", result.RowsAffected)

	// Cleanup completed background jobs older than cutoff
	result = cs.db.Where("job_created_at < ? AND job_status IN (?)", cutoffDate, []string{"completed", "failed"}).Delete(&entityBackground.BackgroundJob{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup background jobs: %w", result.Error)
	}
	log.Printf("✅ Cleaned up %d old background jobs", result.RowsAffected)

	// Cleanup cancelled bookings older than cutoff
	result = cs.db.Where("booking_created_at < ? AND booking_status = ?", cutoffDate, "cancelled").Delete(&entityBooking.ConsultationBooking{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup cancelled bookings: %w", result.Error)
	}
	log.Printf("✅ Cleaned up %d old cancelled bookings", result.RowsAffected)

	log.Printf("✅ Cleanup completed successfully")
	return nil
}
