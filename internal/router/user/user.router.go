package user

import (
	"cbs_backend/internal/middleware"
	PkgUser "cbs_backend/internal/modules/users"
	"cbs_backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserRouter chứa các route liên quan đến người dùng
type UserRouter struct{}

// InitUserRouter khởi tạo router cho người dùng
func (ur *UserRouter) InitUserRouter(router *gin.RouterGroup) {
	userCtrl := PkgUser.NewUserController()

	// Nhóm route public (không cần xác thực)
	public := router.Group("/user/v1")
	{
		public.POST("/register", response.Wrap(userCtrl.Register)) // Đăng ký tài khoản
		public.POST("/login", response.Wrap(userCtrl.Login))       // Đăng nhập
	}

	// Nhóm route private (cần xác thực token JWT)
	private := router.Group("/user/v2")
	private.Use(middleware.AuthMiddleware(PkgUser.User())) // Middleware xác thực
	{
		private.GET("/GetInfor", response.Wrap(userCtrl.GetInfor)) // Lấy thông tin người dùng
		private.PUT("/update", response.Wrap(userCtrl.UpdateInforUser))
		private.GET("/logout", response.Wrap(userCtrl.Logout))
	}
}
