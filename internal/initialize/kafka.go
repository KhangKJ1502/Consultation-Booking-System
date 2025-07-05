package initialize

import (
	"fmt"
	"log"

	"cbs_backend/global"
	"cbs_backend/internal/kafka"
	"cbs_backend/internal/service/email"
	// emailservice "cbs_backend/internal/service"
)

func InitKafka() {
	cfg := global.ConfigConection.KafkaCF

	if len(cfg.Brokers) == 0 {
		log.Println("⚠️ Kafka brokers not configured, skipping Kafka initialization")
		return
	}

	// 1. Initialize Producer
	if err := kafka.InitProducer(cfg.Brokers); err != nil {
		log.Printf("❌ Failed to initialize Kafka producer: %v", err)
		return
	}

	// 2. Initialize EmailService
	// Cần truyền db và logger từ global hoặc parameter
	// emailService := emailservice.NewEmailService(global.DB, global.Log)
	// New way:
	emailService := email.NewEmailManager(global.DB, global.Log)

	// 3. Initialize EventHandler với EmailService
	handler := kafka.NewEventHandlerWithEmailService(emailService) // ← Fix: Truyền emailService

	// 4. Initialize Consumer
	consumer, err := kafka.NewConsumer(
		cfg.Brokers,
		cfg.GroupID,
		[]string{"user-events", "user-notifications", "booking-events"},
		handler,
	)
	if err != nil {
		log.Printf("❌ Failed to create Kafka consumer: %v", err)
		return
	}

	// 5. Start consumer in goroutine
	go consumer.Start()

	fmt.Printf("✅ Kafka initialized successfully:\n")
	fmt.Printf("  - Brokers: %v\n", cfg.Brokers)
	fmt.Printf("  - Group ID: %s\n", cfg.GroupID)
	fmt.Printf("  - Topics: user-events, user-notifications\n")
}
