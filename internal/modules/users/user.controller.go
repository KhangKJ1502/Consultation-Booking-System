package users

import (
	dtousergo "cbs_backend/internal/modules/users/dto.user.go"
	"cbs_backend/pkg/response"
	"fmt"
	"log"
	"net/http"
	"strconv"
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
	return res, nil
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

// RefreshToken - Refresh JWT token
func (uc *UserController) RefreshToken(ctx *gin.Context) (res interface{}, err error) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request payload", err.Error())
	}

	newToken, err := User().RefeshToken(ctx.Request.Context(), req.RefreshToken)
	if err != nil {
		return nil, response.NewAPIError(http.StatusUnauthorized, "Invalid refresh token", err.Error())
	}

	return map[string]interface{}{
		"access_token": newToken,
	}, nil
}

// ChangePassword - Change user password
func (uc *UserController) ChangePassword(ctx *gin.Context) (res interface{}, err error) {
	var req dtousergo.ChangePasswordRequest

	// Lấy userID từ context
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		return nil, response.NewAPIError(http.StatusUnauthorized, "Unauthorized", "UserID not found in context")
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return nil, response.NewAPIError(http.StatusInternalServerError, "Internal error", "Invalid userID type")
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request payload", err.Error())
	}

	err = User().ChangePassword(ctx.Request.Context(), req, userID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Failed to change password", err.Error())
	}

	return map[string]interface{}{
		"message": "Password changed successfully",
	}, nil
}

// ResetPassword - Request password reset
func (uc *UserController) ResetPassword(ctx *gin.Context) (res interface{}, err error) {
	var req dtousergo.ResetPasswordRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request payload", err.Error())
	}

	err = User().ResetPassword(ctx.Request.Context(), req)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Failed to reset password", err.Error())
	}

	return map[string]interface{}{
		"message": "Password reset email sent successfully",
	}, nil
}

// ConfirmResetPassword - Confirm password reset with token
func (uc *UserController) ConfirmResetPassword(ctx *gin.Context) (res interface{}, err error) {
	var req dtousergo.ConfirmResetPasswordRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request payload", err.Error())
	}

	err = User().ConfirmResetPassword(ctx.Request.Context(), req)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Failed to confirm password reset", err.Error())
	}

	return map[string]interface{}{
		"message": "Password reset successfully",
	}, nil
}

// DeleteAccount - Delete user account
func (uc *UserController) DeleteAccount(ctx *gin.Context) (res interface{}, err error) {
	var req dtousergo.DeleteAccountRequest

	// Lấy userID từ context
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		return nil, response.NewAPIError(http.StatusUnauthorized, "Unauthorized", "UserID not found in context")
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return nil, response.NewAPIError(http.StatusInternalServerError, "Internal error", "Invalid userID type")
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request payload", err.Error())
	}

	err = User().DeleteAccount(ctx.Request.Context(), req, userID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Failed to delete account", err.Error())
	}

	return map[string]interface{}{
		"message": "Account deleted successfully",
	}, nil
}

// GetUsersByRole - Get users by role (Admin function)
func (uc *UserController) GetUsersByRole(ctx *gin.Context) (res interface{}, err error) {
	role := ctx.Query("role")
	if role == "" {
		return nil, response.NewAPIError(http.StatusBadRequest, "Role parameter is required", "")
	}

	page := 1
	limit := 10

	if pageStr := ctx.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	users, err := User().GetUsersByRole(ctx.Request.Context(), role, page, limit)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, "Failed to get users", err.Error())
	}

	return users, nil
}

// DeactivateUser - Deactivate user account (Admin function)
func (uc *UserController) DeactivateUser(ctx *gin.Context) (res interface{}, err error) {
	targetUserIDStr := ctx.Param("userID")
	if targetUserIDStr == "" {
		return nil, response.NewAPIError(http.StatusBadRequest, "User ID is required", "")
	}

	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid user ID format", err.Error())
	}

	err = User().DeactivateUser(ctx.Request.Context(), targetUserID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Failed to deactivate user", err.Error())
	}

	return map[string]interface{}{
		"message": "User deactivated successfully",
	}, nil
}

// ActivateUser - Activate user account (Admin function)
func (uc *UserController) ActivateUser(ctx *gin.Context) (res interface{}, err error) {
	targetUserIDStr := ctx.Param("userID")
	if targetUserIDStr == "" {
		return nil, response.NewAPIError(http.StatusBadRequest, "User ID is required", "")
	}

	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid user ID format", err.Error())
	}

	err = User().ActivateUser(ctx.Request.Context(), targetUserID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Failed to activate user", err.Error())
	}

	return map[string]interface{}{
		"message": "User activated successfully",
	}, nil
}

// LogoutAllSessions - Logout all user sessions
func (uc *UserController) LogoutAllSessions(ctx *gin.Context) (res interface{}, err error) {
	// Lấy userID từ context
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		return nil, response.NewAPIError(http.StatusUnauthorized, "Unauthorized", "UserID not found in context")
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return nil, response.NewAPIError(http.StatusInternalServerError, "Internal error", "Invalid userID type")
	}

	err = User().LogoutAllSessions(ctx.Request.Context(), userID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, "Failed to logout all sessions", err.Error())
	}

	return map[string]interface{}{
		"message": "All sessions logged out successfully",
	}, nil
}

// UpdateUserRole - Update user role (Admin function)
func (uc *UserController) UpdateUserRole(ctx *gin.Context) (res interface{}, err error) {
	targetUserIDStr := ctx.Param("userID")
	if targetUserIDStr == "" {
		return nil, response.NewAPIError(http.StatusBadRequest, "User ID is required", "")
	}

	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid user ID format", err.Error())
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request payload", err.Error())
	}

	err = User().UpdateUserRole(ctx.Request.Context(), targetUserID, req.Role)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Failed to update user role", err.Error())
	}

	return map[string]interface{}{
		"message": "User role updated successfully",
	}, nil
}

// SearchUsers - Search users with filters
func (uc *UserController) SearchUsers(ctx *gin.Context) (res interface{}, err error) {
	var req dtousergo.SearchUsersRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request payload", err.Error())
	}

	users, err := User().SearchUsers(ctx.Request.Context(), req)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, "Failed to search users", err.Error())
	}

	return users, nil
}

// UpdateEmail - Update user email
func (uc *UserController) UpdateEmail(ctx *gin.Context) (res interface{}, err error) {
	var req dtousergo.UpdateEmailRequest

	// Lấy userID từ context
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		return nil, response.NewAPIError(http.StatusUnauthorized, "Unauthorized", "UserID not found in context")
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return nil, response.NewAPIError(http.StatusInternalServerError, "Internal error", "Invalid userID type")
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request payload", err.Error())
	}

	err = User().UpdateEmail(ctx.Request.Context(), req, userID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Failed to update email", err.Error())
	}

	return map[string]interface{}{
		"message": "Email updated successfully",
	}, nil
}

// GetActiveTokens - Get user's active tokens
func (uc *UserController) GetActiveTokens(ctx *gin.Context) (res interface{}, err error) {
	// Lấy userID từ context
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		return nil, response.NewAPIError(http.StatusUnauthorized, "Unauthorized", "UserID not found in context")
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return nil, response.NewAPIError(http.StatusInternalServerError, "Internal error", "Invalid userID type")
	}

	tokens, err := User().GetActiveTokens(ctx.Request.Context(), userID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, "Failed to get active tokens", err.Error())
	}

	return tokens, nil
}

// RevokeToken - Revoke specific token
func (uc *UserController) RevokeToken(ctx *gin.Context) (res interface{}, err error) {
	tokenIDStr := ctx.Param("tokenID")
	if tokenIDStr == "" {
		return nil, response.NewAPIError(http.StatusBadRequest, "Token ID is required", "")
	}

	tokenID, err := uuid.Parse(tokenIDStr)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid token ID format", err.Error())
	}

	// Lấy userID từ context
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		return nil, response.NewAPIError(http.StatusUnauthorized, "Unauthorized", "UserID not found in context")
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return nil, response.NewAPIError(http.StatusInternalServerError, "Internal error", "Invalid userID type")
	}

	err = User().RevokeToken(ctx.Request.Context(), tokenID, userID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Failed to revoke token", err.Error())
	}

	return map[string]interface{}{
		"message": "Token revoked successfully",
	}, nil
}
