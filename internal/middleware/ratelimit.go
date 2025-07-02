package middleware

import "net/http"

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check Redis or in-memory counter for user/IP
		// If exceed, return 429
		next.ServeHTTP(w, r)
	})
}
