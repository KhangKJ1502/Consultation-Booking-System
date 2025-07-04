package initialize

import (
	"cbs_backend/global"
	"cbs_backend/internal/middleware"
	"cbs_backend/internal/modules/realtime"
	routerAll "cbs_backend/internal/router"
	"cbs_backend/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	// Initialize the router
	// This function will set up the routes and middleware for the application
	// It will return a gin.Engine instance that can be used to run the server

	var r *gin.Engine
	// Set the mode based on the environment
	if global.ConfigConection.ServerCF.GinMode == "dev" {
		gin.SetMode(gin.DebugMode)
		gin.ForceConsoleColor()
		r = gin.Default()
	} else {
		gin.SetMode(gin.ReleaseMode)
		r = gin.New()
	}
	r.Use(middleware.CORS) // cross
	r.Use(middleware.ValidatorMiddleware())
	// r.Use() // logging

	//Thêm websocket vào
	r.GET("/ws", realtime.WSHandler)
	// r.Use() // limiter global
	// r.Use(middlewares.Validator())      // middleware

	// r.Use(middlewares.NewRateLimiter().GlobalRateLimiter()) // 100 req/s

	// Route kiểm tra hoạt động
	r.GET("/ping/100", func(ctx *gin.Context) {
		response.SuccessResponse(ctx, http.StatusOK, "OK")
	})
	UserMainGroup := routerAll.RouterGroupApp.User
	ExpertMainGroup := routerAll.RouterGroupApp.Expert
	BookingMainGroup := routerAll.RouterGroupApp.Booking
	// Nhóm route chính (có thể đặt prefix như /api)
	apiGroup := r.Group("")
	{
		UserMainGroup.InitUserRouter(apiGroup) // Khởi tạo route user
		ExpertMainGroup.InitExpertRouter(apiGroup)
		BookingMainGroup.InitBookingRouter(apiGroup)
	}

	return r
}
