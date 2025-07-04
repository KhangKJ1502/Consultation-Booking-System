package main

import (
	"cbs_backend/global"
	"cbs_backend/internal/initialize"
	"cbs_backend/internal/worker"

	"github.com/robfig/cron"
)

func main() {
	r := initialize.Run()
	// Khởi tạo worker với config
	// Khởi tạo cron
	c := cron.New()

	// Đăng ký các cron-job
	worker.StartReminderCron(global.DB, c)

	// Khởi động cron
	c.Start()
	// 🔥 Quan trọng: Chạy server tại đây
	if err := r.Run(global.ConfigConection.ServerCF.Host + ":" + global.ConfigConection.ServerCF.Port); err != nil {
		panic(err)
	}
}
