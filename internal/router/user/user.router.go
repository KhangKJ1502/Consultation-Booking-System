package user

import (
	"cbs_backend/internal/middleware"
	PkgUser "cbs_backend/internal/modules/users"
	"cbs_backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserRouter chứa các route liên quan đến người dùng
type UserRouter struct{}

// NewUserRouter khởi tạo UserRouter mới
func NewUserRouter() *UserRouter {
	return &UserRouter{}
}

// InitUserRouter khởi tạo router cho người dùng
func (ur *UserRouter) InitUserRouter(router *gin.RouterGroup) {
	userCtrl := PkgUser.NewUserController()

	// Nhóm route public (không cần xác thực)
	public := router.Group("/user/v1")
	{
		// Authentication routes
		public.POST("/register", response.Wrap(userCtrl.Register))                  // Đăng ký tài khoản
		public.POST("/login", response.Wrap(userCtrl.Login))                        // Đăng nhập
		public.POST("/refresh-token", response.Wrap(userCtrl.RefreshToken))         // Làm mới token
		public.POST("/reset-password", response.Wrap(userCtrl.ResetPassword))       // Yêu cầu reset mật khẩu
		public.POST("/confirm-reset", response.Wrap(userCtrl.ConfirmResetPassword)) // Xác nhận reset mật khẩu
	}

	// Nhóm route private (cần xác thực token JWT)
	private := router.Group("/user/v2")
	private.Use(middleware.AuthMiddleware(PkgUser.User())) // Middleware xác thực
	{
		// User profile routes
		private.GET("/profile", response.Wrap(userCtrl.GetInfor))        // Lấy thông tin người dùng
		private.PUT("/profile", response.Wrap(userCtrl.UpdateInforUser)) // Cập nhật thông tin người dùng
		private.PUT("/email", response.Wrap(userCtrl.UpdateEmail))       // Cập nhật email

		// Authentication management routes
		private.POST("/logout", response.Wrap(userCtrl.Logout))                  // Đăng xuất
		private.POST("/logout-all", response.Wrap(userCtrl.LogoutAllSessions))   // Đăng xuất tất cả phiên
		private.POST("/change-password", response.Wrap(userCtrl.ChangePassword)) // Đổi mật khẩu
		private.DELETE("/account", response.Wrap(userCtrl.DeleteAccount))        // Xóa tài khoản

		// Token management routes
		private.GET("/tokens", response.Wrap(userCtrl.GetActiveTokens))         // Lấy danh sách token đang hoạt động
		private.DELETE("/tokens/:tokenID", response.Wrap(userCtrl.RevokeToken)) // Thu hồi token cụ thể
	}

	// Nhóm route admin (cần xác thực và quyền admin)
	admin := router.Group("/user/v3")
	admin.Use(middleware.AuthMiddleware(PkgUser.User()))
	// admin.Use(middleware.AdminMiddleware()) // Middleware kiểm tra quyền admin
	{
		// User management routes
		admin.GET("/users", response.Wrap(userCtrl.GetUsersByRole))                    // Lấy người dùng theo role
		admin.POST("/users/search", response.Wrap(userCtrl.SearchUsers))               // Tìm kiếm người dùng
		admin.PUT("/users/:userID/role", response.Wrap(userCtrl.UpdateUserRole))       // Cập nhật role người dùng
		admin.PUT("/users/:userID/deactivate", response.Wrap(userCtrl.DeactivateUser)) // Vô hiệu hóa tài khoản
		admin.PUT("/users/:userID/activate", response.Wrap(userCtrl.ActivateUser))     // Kích hoạt tài khoản
	}
}

// InitUserRouterWithCustomMiddleware khởi tạo router với middleware tùy chỉnh
func (ur *UserRouter) InitUserRouterWithCustomMiddleware(
	router *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
	adminMiddleware gin.HandlerFunc,
) {
	userCtrl := PkgUser.NewUserController()

	// Public routes
	public := router.Group("/user/v1")
	{
		public.POST("/register", response.Wrap(userCtrl.Register))
		public.POST("/login", response.Wrap(userCtrl.Login))
		public.POST("/refresh-token", response.Wrap(userCtrl.RefreshToken))
		public.POST("/reset-password", response.Wrap(userCtrl.ResetPassword))
		public.POST("/confirm-reset", response.Wrap(userCtrl.ConfirmResetPassword))
	}

	// Private routes
	private := router.Group("/user/v2")
	private.Use(authMiddleware)
	{
		private.GET("/profile", response.Wrap(userCtrl.GetInfor))
		private.PUT("/profile", response.Wrap(userCtrl.UpdateInforUser))
		private.PUT("/email", response.Wrap(userCtrl.UpdateEmail))
		private.POST("/logout", response.Wrap(userCtrl.Logout))
		private.POST("/logout-all", response.Wrap(userCtrl.LogoutAllSessions))
		private.POST("/change-password", response.Wrap(userCtrl.ChangePassword))
		private.DELETE("/account", response.Wrap(userCtrl.DeleteAccount))
		private.GET("/tokens", response.Wrap(userCtrl.GetActiveTokens))
		private.DELETE("/tokens/:tokenID", response.Wrap(userCtrl.RevokeToken))
	}

	// Admin routes
	admin := router.Group("/user/v3")
	admin.Use(authMiddleware)
	admin.Use(adminMiddleware)
	{
		admin.GET("/users", response.Wrap(userCtrl.GetUsersByRole))
		admin.POST("/users/search", response.Wrap(userCtrl.SearchUsers))
		admin.PUT("/users/:userID/role", response.Wrap(userCtrl.UpdateUserRole))
		admin.PUT("/users/:userID/deactivate", response.Wrap(userCtrl.DeactivateUser))
		admin.PUT("/users/:userID/activate", response.Wrap(userCtrl.ActivateUser))
	}
}

// GetUserRoutes trả về danh sách các route được định nghĩa
func (ur *UserRouter) GetUserRoutes() map[string][]string {
	return map[string][]string{
		"public": {
			"POST /user/v1/register",
			"POST /user/v1/login",
			"POST /user/v1/refresh-token",
			"POST /user/v1/reset-password",
			"POST /user/v1/confirm-reset",
		},
		"private": {
			"GET /user/v2/profile",
			"PUT /user/v2/profile",
			"PUT /user/v2/email",
			"POST /user/v2/logout",
			"POST /user/v2/logout-all",
			"POST /user/v2/change-password",
			"DELETE /user/v2/account",
			"GET /user/v2/tokens",
			"DELETE /user/v2/tokens/:tokenID",
		},
		"admin": {
			"GET /user/v3/users",
			"POST /user/v3/users/search",
			"PUT /user/v3/users/:userID/role",
			"PUT /user/v3/users/:userID/deactivate",
			"PUT /user/v3/users/:userID/activate",
		},
	}
}
