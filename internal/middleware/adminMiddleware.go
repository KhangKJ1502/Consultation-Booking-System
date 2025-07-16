package middleware

import (
	"cbs_backend/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminMiddleware kiểm tra quyền admin của user
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lấy user từ context (đã được set bởi AuthMiddleware)
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, response.NewAPIError(
				http.StatusUnauthorized,
				"User not authenticated",
				nil,
			))
			c.Abort()
			return
		}

		// Type assertion để lấy user object
		user, ok := userInterface.(map[string]interface{})
		if !ok {
			c.JSON(http.StatusInternalServerError, response.NewAPIError(
				http.StatusInternalServerError,
				"Invalid user context",
				nil,
			))
			c.Abort()
			return
		}

		// Kiểm tra role của user
		role, roleExists := user["role"]
		if !roleExists {
			c.JSON(http.StatusForbidden, response.NewAPIError(
				http.StatusForbidden,
				"User role not found",
				nil,
			))
			c.Abort()
			return
		}

		// Kiểm tra có phải admin không
		if role != "admin" && role != "super_admin" {
			c.JSON(http.StatusForbidden, response.NewAPIError(
				http.StatusForbidden,
				"Access denied. Admin role required",
				nil,
			))
			c.Abort()
			return
		}
		c.Next()
	}
}
