package worker

import (
	"cbs_backend/internal/modules/bookings/entity"
	"cbs_backend/internal/modules/realtime"
	"log"
	"time"

	"github.com/robfig/cron"
	"gorm.io/gorm"
)

func StartReminderCron(db *gorm.DB, c *cron.Cron) {
	c.AddFunc("*/10 * * * *", func() {
		log.Println("[ReminderJob] Start checking bookings...")
		now := time.Now()
		remindFrom := now.Add(60 * time.Minute)

		var bookings []entity.ConsultationBooking
		err := db.Where("booking_datetime BETWEEN ? AND ? AND reminder_sent = false AND booking_status = ?", now, remindFrom, "confirmed").
			Find(&bookings).Error
		if err != nil {
			log.Printf("[ReminderJob] DB error: %v", err)
			return
		}

		for _, booking := range bookings {
			go realtime.Send(booking.UserID.String(), "Bạn sắp có lịch tư vấn lúc "+booking.BookingDatetime.Format("02/01/2006 15:04"))
			go realtime.Send(booking.ExpertProfileID.String(), "Bạn sắp có lịch tư vấn với khách lúc "+booking.BookingDatetime.Format("02/01/2006 15:04"))

			if err := db.Model(&booking).Update("reminder_sent", true).Error; err != nil {
				log.Printf("[ReminderJob] Failed to update reminder_sent: %v", err)
			}
		}
		log.Printf("[ReminderJob] Done. %d reminders sent.", len(bookings))
	})
}
