// internal/worker/reminder_worker.go
package worker

import (
	"cbs_backend/internal/modules/bookings/entity"
	"cbs_backend/internal/modules/realtime"
	"time"

	"gorm.io/gorm"
)

func StartBookingReminderWorker(db *gorm.DB) {
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			now := time.Now()
			remindFrom := now.Add(60 * time.Minute)
			var bookings []entity.ConsultationBooking
			db.Where("booking_datetime BETWEEN ? AND ? AND reminder_sent = false AND booking_status = ?", now, remindFrom, "confirmed").
				Find(&bookings)
			for _, booking := range bookings {
				go realtime.Send(booking.UserID.String(), "Bạn sắp có lịch tư vấn lúc "+booking.BookingDatetime.Format("02/01/2006 15:04"))
				go realtime.Send(booking.ExpertProfileID.String(), "Bạn sắp có lịch tư vấn với khách lúc "+booking.BookingDatetime.Format("02/01/2006 15:04"))
				db.Model(&booking).Update("reminder_sent", true)
			}
		}
	}()
}
