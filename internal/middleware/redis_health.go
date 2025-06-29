// middleware/redis_health.go
package middleware

import (
	"cbs_backend/global"
	"cbs_backend/pkg/response"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RedisHealthMiddleware checks if Redis is available before processing auth
func RedisHealthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip health check for health endpoints
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/ping" {
			c.Next()
			return
		}

		// Check if Redis is required for this endpoint
		if !requiresAuth(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Quick Redis health check
		if !isRedisHealthy() {
			response.ErrorResponse(c, http.StatusServiceUnavailable, "Service Unavailable", "Authentication service temporarily unavailable")
			c.Abort()
			return
		}

		c.Next()
	}
}

// isRedisHealthy performs a quick Redis health check
func isRedisHealthy() bool {
	if global.Redis == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := global.Redis.Ping(ctx).Err()
	return err == nil
}

// requiresAuth checks if the endpoint requires authentication
func requiresAuth(path string) bool {
	publicPaths := []string{
		"/api/auth/login",
		"/api/auth/register",
		"/health",
		"/ping",
		"/api/public",
	}

	for _, publicPath := range publicPaths {
		if path == publicPath {
			return false
		}
	}

	return true
}

// Usage in main.go:
// r.Use(middleware.RedisHealthMiddleware())
// r.Use(middleware.AuthMiddleware(userService))
