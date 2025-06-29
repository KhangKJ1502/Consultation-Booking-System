// internal/initialize/run.go
package initialize

import (
	"cbs_backend/global"
	pkg "cbs_backend/pkg/configs"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

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

	// Step 6: Init Router
	router := InitRouter()
	global.Log.Info("✅ Router initialized")

	return router
}
