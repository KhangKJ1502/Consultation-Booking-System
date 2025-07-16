package initialize

import (
	"cbs_backend/internal/modules/bookings"
	"cbs_backend/internal/modules/dashboard"
	"cbs_backend/internal/modules/experts"
	"cbs_backend/internal/modules/users"
	"cbs_backend/internal/service/email"
	"cbs_backend/utils/cache"

	"github.com/bsm/redislock"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func InitServices(
	db *gorm.DB,
	redis *cache.RedisCache, // low‑level Redis cache
	log *zap.Logger,
) {

	redisLocker := redislock.New(redis.Client) // redis.Client là *redis.Client
	// 1. High‑level caches
	userCache := cache.NewRedisUserCache(redis, log)
	expertCache := cache.NewRedisExpertCache(redis)
	bookingCache := cache.NewRedisBookingCache(redis, log)

	// 2. Email
	emailSvc := email.NewEmailManager(db, log)
	_ = emailSvc // gán vào global.registry nếu bạn muốn dùng sau
	// 3. Users
	users.InitUserService(db, userCache, log)
	// 4. Experts
	experts.InitExpertService(db, expertCache, log)
	//5.Booking
	bookings.InitBookingService(db, bookingCache, log, redisLocker)
	dashboard.InitDashboardService(db, log)
}
