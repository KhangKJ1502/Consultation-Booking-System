package worker

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

func generateJobID() string {
	return uuid.New().String()
}

// Helper function to get current time in Vietnam timezone
func getCurrentVietnamTime() time.Time {
	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	return time.Now().In(loc)
}

// Helper function to format time for Vietnamese locale
func formatVietnameseTime(t time.Time) string {
	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	return t.In(loc).Format("15:04 ngày 02/01/2006")
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		// Nếu là kiểu khác (ví dụ float64), bạn có thể convert sang string nếu muốn
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// Dành cho reminder gửi trước 1 tiếng
func FormatTimeUntil(bookingTime time.Time) string {
	now := time.Now()
	duration := bookingTime.Sub(now)

	if duration < 0 {
		return "Đã qua"
	}

	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("khoảng %d giờ %d phút nữa", hours, minutes)
	}
	return fmt.Sprintf("khoảng %d phút nữa", minutes)
}
