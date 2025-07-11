package initialize

import (
	"cbs_backend/global"
	"cbs_backend/internal/service/email"
	"cbs_backend/internal/worker"
	pkg "cbs_backend/pkg/configs"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// WorkerScheduler global instance
var WorkerScheduler *worker.WorkerScheduler

// Run initializes and starts the application
func Run() *gin.Engine {
	fmt.Println("🚀 Starting ecommerce backend API...")

	// Step 0: Init logger
	fmt.Println("📝 Initializing logger...")
	InitLogger()
	global.Log.Info("✅ Logger initialized successfully")

	// Step 1: Load configuration
	cfg, err := pkg.LoadConfigs()
	if err != nil {
		global.Log.Fatal("❌ Failed to load config", zap.Error(err))
	}
	global.ConfigConection = cfg
	global.Log.Info("✅ Configuration loaded", zap.Any("config", cfg))

	// Step 2: Init DB
	db := InitPostgres()
	global.Log.Info("✅ Database initialized")

	// Step 3: Init Redis
	redis := InitRedis()
	global.Log.Info("✅ Redis initialized")

	// Step 4: Init Kafka
	InitKafka()

	// Step 5: Init Services
	InitServices(db, redis, global.Log)

	// Step 6: Init Worker Scheduler
	InitWorker()

	// Step 7: Init Router
	router := InitRouter()
	global.Log.Info("✅ Router initialized")

	return router
}

// InitWorker khởi tạo và start worker scheduler
func InitWorker() {
	fmt.Println("👷 Initializing Worker Scheduler...")

	maxWorkers := 5 // Có thể lấy từ config
	emailSvc := email.NewEmailManager(global.DB, global.Log)
	WorkerScheduler = worker.NewWorkerScheduler(global.DB, maxWorkers, emailSvc, global.Redis)

	if err := WorkerScheduler.Start(); err != nil {
		global.Log.Fatal("❌ Failed to start worker scheduler", zap.Error(err))
	}

	global.Log.Info("✅ Worker Scheduler initialized and started successfully")
}

// StopWorker dừng worker scheduler

func StopWorker() {
	if WorkerScheduler != nil {
		fmt.Println("🛑 Stopping Worker Scheduler...")
		WorkerScheduler.Stop()
		global.Log.Info("✅ Worker Scheduler stopped")
	}
}

// package initialize

// import (
// 	"cbs_backend/global"
// 	"cbs_backend/internal/worker/booking"
// 	"cbs_backend/internal/worker/notification/sender"
// 	"cbs_backend/internal/worker/scheduler"
// 	pkg "cbs_backend/pkg/configs"
// 	"context"
// 	"fmt"

// 	"github.com/gin-gonic/gin"
// 	"go.uber.org/zap"
// )

// var WorkerScheduler *scheduler.WorkerScheduler

// // Run initializes and starts the application
// func Run() *gin.Engine {
// 	fmt.Println("🚀 Starting ecommerce backend API...")

// 	// Step 0: Init logger
// 	fmt.Println("📝 Initializing logger...")
// 	InitLogger()
// 	global.Log.Info("✅ Logger initialized successfully")

// 	// Step 1: Load configuration
// 	cfg, err := pkg.LoadConfigs()
// 	if err != nil {
// 		global.Log.Fatal("❌ Failed to load config", zap.Error(err))
// 	}
// 	global.ConfigConection = cfg
// 	global.Log.Info("✅ Configuration loaded", zap.Any("config", cfg))

// 	// Step 2: Init DB
// 	db := InitPostgres()
// 	global.Log.Info("✅ Database initialized")

// 	// Step 3: Init Redis
// 	redis := InitRedis()
// 	global.Log.Info("✅ Redis initialized")

// 	// Step 4: Init Kafka
// 	InitKafka()

// 	// Step 5: Init Services
// 	InitServices(db, redis, global.Log)

// 	// Step 6: Init Worker Scheduler
// 	InitWorker()

// 	// Step 7: Init Router
// 	router := InitRouter()
// 	global.Log.Info("✅ Router initialized")

// 	return router
// }
// func InitWorker() {
// 	fmt.Println("👷 Initializing Worker Scheduler...")

// 	// Khởi tạo các sender
// 	emailSender := sender.NewEmailSender(global.EmailService)
// 	smsSender := sender.NewSMSSender(global.ConfigConection.SMSCF.SMSApiKey, global.ConfigConection.SMSCF.SMSApiURL)
// 	telegramSender := sender.NewTelegramSender(global.ConfigConection.TLGCF.TELEGRAM_BOT_TOKEN)

// 	// Khởi tạo notification service
// 	notificationService := sender.NewNotificationService(emailSender, smsSender, telegramSender)

// 	// Khởi tạo các service worker
// 	reminderSvc := booking.NewReminderService(notificationService)
// 	statsSvc := booking.NewStatsService()
// 	statusSvc := booking.NewStatusService()

// 	// Khởi tạo job processor
// 	jobProcessor := scheduler.NewJobProcessor(reminderSvc, statsSvc, statusSvc)

// 	// Khởi tạo worker scheduler
// 	maxWorkers := 5 // hoặc lấy từ config
// 	WorkerScheduler = scheduler.NewWorkerScheduler(jobProcessor, maxWorkers)

// 	// Start worker scheduler
// 	if err := WorkerScheduler.Start(context.Background()); err != nil {
// 		global.Log.Fatal("❌ Failed to start worker scheduler", zap.Error(err))
// 	}

// 	global.Log.Info("✅ Worker Scheduler initialized and started successfully")
// }

// // StopWorker dừng worker scheduler
// func StopWorker() {
// 	if WorkerScheduler != nil {
// 		fmt.Println("🛑 Stopping Worker Scheduler...")
// 		WorkerScheduler.Stop()
// 		global.Log.Info("✅ Worker Scheduler stopped")
// 	}
// }
