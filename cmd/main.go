package main

import (
	"cbs_backend/global"
	"cbs_backend/internal/initialize"
	"cbs_backend/internal/worker"
)

func main() {
	r := initialize.Run()
	// Khởi tạo worker với config
	worker.StartBookingReminderWorker(global.DB)
	worker.StartNotificationCleanupWorker(global.DB)
	// 🔥 Quan trọng: Chạy server tại đây
	if err := r.Run(global.ConfigConection.ServerCF.Host + ":" + global.ConfigConection.ServerCF.Port); err != nil {
		panic(err)
	}
}
