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

// InitUserRouter khởi tạo router cho người dùng với Rate Limiting
func (ur *UserRouter) InitUserRouter(router *gin.RouterGroup) {
	userCtrl := PkgUser.NewUserController()

	// Nhóm route public (không cần xác thực) - với Rate Limiting
	public := router.Group("/user/v1")
	{
		// Authentication routes với Rate Limiting nghiêm ngặt
		public.POST("/register",
			middleware.RegisterLimiter.Middleware(),
			response.Wrap(userCtrl.Register))

		public.POST("/login",
			middleware.LoginLimiter.Middleware(),
			response.Wrap(userCtrl.Login))

		public.POST("/refresh-token",
			middleware.LoginLimiter.Middleware(), // Sử dụng chung với login
			response.Wrap(userCtrl.RefreshToken))

		public.POST("/reset-password",
			middleware.ResetPasswordLimiter.Middleware(),
			response.Wrap(userCtrl.ResetPassword))

		public.POST("/confirm-reset",
			middleware.ResetPasswordLimiter.Middleware(), // Sử dụng chung với reset
			response.Wrap(userCtrl.ConfirmResetPassword))
	}

	// Nhóm route private (cần xác thực token JWT) - với Rate Limiting
	private := router.Group("/user/v2")
	private.Use(middleware.AuthMiddleware(PkgUser.User()))
	{
		// User profile routes với Rate Limiting
		private.GET("/profile", response.Wrap(userCtrl.GetInfor)) // Không cần rate limit cho GET

		private.PUT("/profile",
			middleware.UpdateProfileLimiter.Middleware(),
			response.Wrap(userCtrl.UpdateInforUser))

		private.PUT("/email",
			middleware.UpdateProfileLimiter.Middleware(), // Email update cần rate limit
			response.Wrap(userCtrl.UpdateEmail))

		// Authentication management routes với Rate Limiting
		private.POST("/logout", response.Wrap(userCtrl.Logout)) // Không cần rate limit cho logout

		private.POST("/logout-all",
			middleware.LoginLimiter.Middleware(), // Bảo vệ chống spam logout all
			response.Wrap(userCtrl.LogoutAllSessions))

		private.POST("/change-password",
			middleware.ChangePasswordLimiter.Middleware(),
			response.Wrap(userCtrl.ChangePassword))

		private.DELETE("/account",
			middleware.LoginLimiter.Middleware(), // Bảo vệ việc xóa tài khoản
			response.Wrap(userCtrl.DeleteAccount))

		// Token management routes
		private.GET("/tokens", response.Wrap(userCtrl.GetActiveTokens)) // Không cần rate limit cho GET

		private.DELETE("/tokens/:tokenID",
			middleware.UpdateProfileLimiter.Middleware(), // Rate limit cho token revoke
			response.Wrap(userCtrl.RevokeToken))
	}

	// Nhóm route admin (cần xác thực và quyền admin) - với Rate Limiting
	admin := router.Group("/user/v3")
	admin.Use(middleware.AuthMiddleware(PkgUser.User()))
	// admin.Use(middleware.AdminMiddleware()) // Middleware kiểm tra quyền admin
	{
		// User management routes với Rate Limiting
		admin.GET("/users", response.Wrap(userCtrl.GetUsersByRole)) // Không cần rate limit cho GET

		admin.POST("/users/search",
			middleware.SearchUserLimiter.Middleware(),
			response.Wrap(userCtrl.SearchUsers))

		admin.PUT("/users/:userID/role",
			middleware.UpdateProfileLimiter.Middleware(), // Rate limit cho update role
			response.Wrap(userCtrl.UpdateUserRole))

		admin.PUT("/users/:userID/deactivate",
			middleware.UpdateProfileLimiter.Middleware(), // Rate limit cho deactivate
			response.Wrap(userCtrl.DeactivateUser))

		admin.PUT("/users/:userID/activate",
			middleware.UpdateProfileLimiter.Middleware(), // Rate limit cho activate
			response.Wrap(userCtrl.ActivateUser))
	}
}
