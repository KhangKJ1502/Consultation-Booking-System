package users

import (
	dtousergo "cbs_backend/internal/modules/users/dto.user.go"
	"cbs_backend/pkg/response"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserController struct{}

func NewUserController() *UserController {
	return &UserController{}
}

func (uc *UserController) Register(ctx *gin.Context) (res interface{}, err error) {
	var req dtousergo.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request payload", err.Error())
	}
	// Gọi service
	res, err = User().Register(ctx, req)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, "Failed to register", err.Error())

	}
	return req, nil
}

func (uc *UserController) Login(ctx *gin.Context) (res interface{}, err error) {
	var req dtousergo.LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request payload", err.Error())
	}

	resUser, err := User().Login(ctx, req)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, "Failed to login", err.Error())
	}
	ctx.Header("Authorization", "Bearer "+resUser.Token)
	return resUser, nil
}

func (uc *UserController) GetInfor(ctx *gin.Context) (res interface{}, err error) {
	// Lấy userID từ context (đã được middleware xác thực)
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		return nil, response.NewAPIError(401, "Unauthorized", "UserID not found in context")
	}
	// userIDValue, err := uuid.Parse("2dd31763-d147-4768-bd84-54413f1c9b08")
	if err != nil {
		// xử lý lỗi nếu UUID không hợp lệ
		log.Fatalf("UUID không hợp lệ: %v", err)
	}
	fmt.Println(userIDValue)

	// userID := userIDValue
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return nil, response.NewAPIError(500, "Internal error", "Invalid userID type")
	}

	// Lấy user từ DB
	user, err := User().GetUserByID(ctx, userID)
	if err != nil {
		return nil, response.NewAPIError(404, "User not found", err.Error())
	}

	// Trả về thông tin gọn
	res = dtousergo.UserProfileResponse{
		UserID:         user.UserID,
		FullName:       user.FullName,
		UserEmail:      user.UserEmail,
		PhoneNumber:    user.PhoneNumber,
		AvatarURL:      user.AvatarURL,
		Gender:         user.Gender,
		BioDescription: user.BioDescription,
		UserCreatedAt:  user.UserCreatedAt,
		UserUpdatedAt:  user.UserUpdatedAt,
	}
	return res, nil
}

func (uc *UserController) UpdateInforUser(ctx *gin.Context) (res interface{}, err error) {
	var req dtousergo.InforUserUpdate
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		return nil, response.NewAPIError(401, "Unauthorized", "UserID not found in context")
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return nil, response.NewAPIError(500, "Internal error", "Invalid userID type")
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request payload", err.Error())
	}

	// validation, exists := ctx.Get("validation")
	// if !exists {
	// 	response.ErrorResponse(ctx, http.StatusInternalServerError, "Internal error", "Validation not found in context")
	// 	return
	// }

	// if apiErr := utils.ValidateStruct(req, validation.(*validator.Validate)); apiErr != nil {
	// 	response.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", apiErr)
	// 	return
	// }

	userUpdate, err := User().UpdateInforUser(ctx, req, userID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusNotFound, "User not found", err.Error())
		return
	}
	res = dtousergo.UserProfileResponse{
		UserID:         userUpdate.UserID,
		FullName:       userUpdate.FullName,
		UserEmail:      userUpdate.UserEmail,
		PhoneNumber:    userUpdate.PhoneNumber,
		AvatarURL:      userUpdate.AvatarURL,
		Gender:         userUpdate.Gender,
		BioDescription: userUpdate.BioDescription,
		UserCreatedAt:  userUpdate.UserCreatedAt,
		UserUpdatedAt:  userUpdate.UserUpdatedAt,
	}
	return req, nil
}

// UserController Logout method
func (uc *UserController) Logout(ctx *gin.Context) (res interface{}, err error) {
	// 1. Lấy token từ Authorization header
	rawToken := ctx.GetHeader("Authorization")
	if rawToken == "" {
		return nil, fmt.Errorf("authorization header is required")
	}

	// 2. Validate Bearer format
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(rawToken, bearerPrefix) || len(rawToken) <= len(bearerPrefix) {
		return nil, fmt.Errorf("invalid token format")
	}

	// 3. Extract token
	token := strings.TrimPrefix(rawToken, bearerPrefix)
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("token is required")
	}

	// 4. Lấy userID từ context (đã được set bởi auth middleware)
	userIDInterface, exists := ctx.Get("userID")
	if !exists {
		return nil, fmt.Errorf("user not authenticated")
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("invalid user ID format")
	}

	// 5. Gọi service logout để blacklist token
	err = User().Logout(ctx.Request.Context(), token, userID)
	if err != nil {
		// Log error và trả về lỗi thực tế
		// log.Error("Logout failed", zap.Error(err), zap.String("userID", userID.String()))
		return nil, fmt.Errorf("logout failed: %w", err)
	}
	responseData := map[string]interface{}{
		"status":  "success",
		"code":    200,
		"message": "User logged out successfully",
		"data":    nil,
	}

	return responseData, nil
}
