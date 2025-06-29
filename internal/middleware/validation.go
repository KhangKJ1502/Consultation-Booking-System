package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func ValidatorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		validate := validator.New()

		// set the middleware
		c.Set("validation", validate) // context

		c.Next()
	}
}
