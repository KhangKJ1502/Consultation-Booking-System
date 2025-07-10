package main

import (
	"cbs_backend/global"
	"cbs_backend/internal/initialize"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Khởi tạo tất cả components (bao gồm worker)
	r := initialize.Run()

	// Tạo channel để handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Chạy server trong goroutine
	go func() {
		if err := r.Run(global.ConfigConection.ServerCF.Host + ":" + global.ConfigConection.ServerCF.Port); err != nil {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	log.Printf("🚀 Server started on %s:%s", global.ConfigConection.ServerCF.Host, global.ConfigConection.ServerCF.Port)

	// Chờ tín hiệu shutdown
	<-quit
	log.Println("🛑 Shutting down server...")
	// Dừng worker
	initialize.StopWorker()
	// Đợi worker hoàn thành
	time.Sleep(2 * time.Second)

	log.Println("✅ Server stopped gracefully")
}
