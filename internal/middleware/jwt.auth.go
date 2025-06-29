// middleware/auth.go
package middleware

import (
	"cbs_backend/internal/modules/users"
	"cbs_backend/pkg/response"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AuthMiddleware(user users.IUser) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Kiểm tra Authorization header
		rawToken := c.GetHeader("Authorization")
		if rawToken == "" {
			response.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "Missing Authorization header")
			c.Abort()
			return
		}

		// 2. Validate Bearer format
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(rawToken, bearerPrefix) || len(rawToken) <= len(bearerPrefix) {
			response.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "Invalid token format")
			c.Abort()
			return
		}

		token := strings.TrimPrefix(rawToken, bearerPrefix)

		// 3. Validate token không rỗng
		if strings.TrimSpace(token) == "" {
			response.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "Empty token")
			c.Abort()
			return
		}

		// 4. Tạo context với timeout để tránh hang
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		// 5. Validate token với proper error handling
		userID, err := validateTokenSafely(ctx, user, token)
		if err != nil {
			// Log error for debugging
			fmt.Printf("❌ Token validation failed: %v\n", err)

			// Return generic error to client for security
			response.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "Invalid or expired token")
			c.Abort()
			return
		}

		// 6. Success - set userID và continue
		fmt.Printf("✅ Token validated successfully, userID: %s\n", userID.String())
		c.Set("userID", userID)
		c.Next()
	}
}

// validateTokenSafely wraps the token validation with proper error handling
func validateTokenSafely(ctx context.Context, user users.IUser, token string) (uuid.UUID, error) {
	// Recover from potential panics
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("❌ Panic in token validation: %v\n", r)
		}
	}()

	// Check if user service is available
	if user == nil {
		return uuid.Nil, fmt.Errorf("user service not available")
	}

	// Validate token
	userID, err := user.ValidateToken(ctx, token)
	if err != nil {
		return uuid.Nil, fmt.Errorf("token validation failed: %w", err)
	}

	return userID, nil
}
