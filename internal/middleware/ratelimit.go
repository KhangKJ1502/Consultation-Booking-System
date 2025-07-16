// internal/middleware/ratelimit.go
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    burst,
	}
	go rl.cleanupVisitors()
	return rl
}

func (rl *RateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = limiter
	}
	return limiter
}

func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute * 5)
		rl.mu.Lock()
		for ip, limiter := range rl.visitors {
			if limiter.Allow() {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := rl.GetLimiter(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too many requests",
				"message": "Rate limit exceeded, please try again later",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Rate Limiting Configuration for CBS Project
var (
	// =============================================================================
	// HIGH PRIORITY (BẮT BUỘC) - Security Critical
	// =============================================================================

	// Authentication endpoints - Ngăn brute force
	LoginLimiter          = NewRateLimiter(rate.Every(3*time.Minute), 5)  // 5 per 15min
	RegisterLimiter       = NewRateLimiter(rate.Every(20*time.Minute), 3) // 3 per hour
	ResetPasswordLimiter  = NewRateLimiter(rate.Every(5*time.Minute), 3)  // 3 per 15min
	ChangePasswordLimiter = NewRateLimiter(rate.Every(2*time.Minute), 5)  // 5 per 10min

	// =============================================================================
	// MEDIUM PRIORITY (NÊN) - Business Logic Protection
	// =============================================================================

	// Booking endpoints - Ngăn spam booking
	CreateBookingLimiter     = NewRateLimiter(rate.Every(1*time.Minute), 10) // 10 per 10min
	CancelBookingLimiter     = NewRateLimiter(rate.Every(30*time.Second), 5) // 5 per 2.5min
	RescheduleBookingLimiter = NewRateLimiter(rate.Every(2*time.Minute), 3)  // 3 per 6min

	// Payment endpoints - Ngăn spam payment
	CreatePaymentLimiter  = NewRateLimiter(rate.Every(30*time.Second), 10) // 10 per 5min
	PaymentWebhookLimiter = NewRateLimiter(rate.Every(1*time.Second), 100) // 100 per 100s

	// Review endpoints - Ngăn spam reviews
	CreateReviewLimiter = NewRateLimiter(rate.Every(5*time.Minute), 5) // 5 per 25min

	// =============================================================================
	// LOW PRIORITY (TÙY CHỌN) - Performance Optimization
	// =============================================================================

	// Profile updates - Ngăn update liên tục
	UpdateProfileLimiter       = NewRateLimiter(rate.Every(30*time.Second), 10) // 10 per 5min
	UpdateExpertProfileLimiter = NewRateLimiter(rate.Every(1*time.Minute), 5)   // 5 per 5min

	// Expert configuration updates
	UpdateWorkingHoursLimiter    = NewRateLimiter(rate.Every(1*time.Minute), 5)   // 5 per 5min
	UpdatePricingLimiter         = NewRateLimiter(rate.Every(1*time.Minute), 5)   // 5 per 5min
	UpdateUnavailableTimeLimiter = NewRateLimiter(rate.Every(30*time.Second), 10) // 10 per 5min

	// Search endpoints - Ngăn spam search
	SearchBookingLimiter = NewRateLimiter(rate.Every(1*time.Second), 30) // 30 per 30s
	SearchUserLimiter    = NewRateLimiter(rate.Every(1*time.Second), 50) // 50 per 50s
	SearchExpertLimiter  = NewRateLimiter(rate.Every(1*time.Second), 50) // 50 per 50s

	// Dashboard endpoints - Ngăn spam reports
	DashboardStatsLimiter    = NewRateLimiter(rate.Every(10*time.Second), 20) // 20 per 200s
	RevenueReportLimiter     = NewRateLimiter(rate.Every(30*time.Second), 10) // 10 per 5min
	ExpertPerformanceLimiter = NewRateLimiter(rate.Every(30*time.Second), 10) // 10 per 5min

	// Email sending - Ngăn spam email
	SendEmailLimiter = NewRateLimiter(rate.Every(1*time.Minute), 10) // 10 per 10min

	// General API rate limiter
	APILimiter = NewRateLimiter(rate.Every(1*time.Second), 1000) // 1000 per 1000s
)
