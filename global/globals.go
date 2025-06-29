package global

import (
	"cbs_backend/internal/service/interfaces"
	"cbs_backend/pkg/configs"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ConfigConection *configs.ConectionConfigs
	DB              *gorm.DB
	Redis           *redis.Client
	Log             *zap.Logger
	// 🎯 GLOBAL EMAIL SERVICE
	EmailService interfaces.EmailService
)
