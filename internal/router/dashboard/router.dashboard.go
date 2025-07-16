package dashboard

import (
	"cbs_backend/internal/middleware"
	PkgDashboard "cbs_backend/internal/modules/dashboard"
	PkgUser "cbs_backend/internal/modules/users"
	"cbs_backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type DashBoardRouter struct{}

func NewDashBoardRouter() *DashBoardRouter {
	return &DashBoardRouter{}
}

func (dr *DashBoardRouter) InitDashboardRouter(router *gin.RouterGroup) {
	dashboardCtrl := PkgDashboard.NewDashboardController()

	// Nhóm route private (cần xác thực token JWT)
	private := router.Group("/dashboard/v2")
	private.Use(middleware.AuthMiddleware(PkgUser.User())) // Middleware xác thực qua user service
	{
		// Dashboard statistics routes
		private.POST("/booking-stats", response.Wrap(dashboardCtrl.GetBookingStats))                    // Lấy thống kê booking
		private.GET("/system-overview", response.Wrap(dashboardCtrl.GetSystemOverview))                 // Lấy tổng quan hệ thống
		private.POST("/revenue-report", response.Wrap(dashboardCtrl.GetRevenueReport))                  // Lấy báo cáo doanh thu
		private.GET("/expert/:expertId/performance", response.Wrap(dashboardCtrl.GetExpertPerformance)) // Lấy hiệu suất chuyên gia
	}

	// Nhóm route admin (cần xác thực và quyền admin)
	admin := router.Group("/dashboard/v3")
	admin.Use(middleware.AuthMiddleware(PkgUser.User())) // Middleware xác thực qua user service
	admin.Use(middleware.AdminMiddleware())              // Middleware kiểm tra quyền admin
	{
		// Advanced dashboard routes for admin
		admin.POST("/booking-stats", response.Wrap(dashboardCtrl.GetBookingStats))                    // Lấy thống kê booking (admin)
		admin.GET("/system-overview", response.Wrap(dashboardCtrl.GetSystemOverview))                 // Lấy tổng quan hệ thống (admin)
		admin.POST("/revenue-report", response.Wrap(dashboardCtrl.GetRevenueReport))                  // Lấy báo cáo doanh thu (admin)
		admin.GET("/expert/:expertId/performance", response.Wrap(dashboardCtrl.GetExpertPerformance)) // Lấy hiệu suất chuyên gia (admin)
	}
}
