package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandlerFunc defines the signature for wrapped handlers
type HandlerFunc func(ctx *gin.Context) (res interface{}, err error)

// Wrap standardizes how handlers return success or error responses
func Wrap(handler HandlerFunc) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		res, err := handler(ctx)
		if err != nil {
			if apiErr, ok := err.(*APIError); ok {
				ErrorResponse(ctx, apiErr.StatusCode, apiErr.Message, apiErr.Error())
			} else {
				ErrorResponse(ctx, http.StatusInternalServerError, "Internal Server Error", err.Error())
			}
			return
		}
		SuccessResponse(ctx, http.StatusOK, res)
	}
}
