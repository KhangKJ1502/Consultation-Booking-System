package response

import (
	"github.com/gin-gonic/gin"
)

const (
	ErrCodeParamInvalid = 1001
	InvalidToken        = 4001
)

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// SuccessResponse returns a standardized success response
func SuccessResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(200, APIResponse{
		Code:    statusCode,
		Message: "success",
		Data:    data,
	})
}
func SuccessResponseNodata(c *gin.Context, statusCode int, Message string) {
	c.JSON(200, APIResponse{
		Code:    statusCode,
		Message: Message,
	})
}

// ErrorResponse returns a standardized error response
func ErrorResponse(c *gin.Context, statusCode int, message string, err interface{}) {
	c.JSON(401, APIResponse{
		Code:    statusCode,
		Message: message,
		Error:   err,
	})
}

func UnauthorizedResponse(c *gin.Context, statusCode int, mesage string, err interface{}) {
	c.JSON(402, APIResponse{
		Code:    statusCode,
		Message: mesage,
		Error:   err,
	})
}
