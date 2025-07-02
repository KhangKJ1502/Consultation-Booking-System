// internal/worker/notification_cleanup_worker.go
package worker

import (
	"gorm.io/gorm"
	"time"
)

func StartNotificationCleanupWorker(db *gorm.DB) {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for range ticker.C {
			threshold := time.Now().Add(-72 * time.Hour) // 3 ngày
			db.Exec("DELETE FROM tbl_system_notifications WHERE notification_created_at < ?", threshold)
		}
	}()
}
