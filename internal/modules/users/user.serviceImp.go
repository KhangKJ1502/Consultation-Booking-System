package users

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"cbs_backend/global"
	"cbs_backend/internal/kafka"
	dtousergo "cbs_backend/internal/modules/users/dto.user.go"
	"cbs_backend/internal/modules/users/entity"
	entityuser "cbs_backend/internal/modules/users/entity"
	"cbs_backend/utils"
	utilsCache "cbs_backend/utils/cache"
	"cbs_backend/utils/helper"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"go.uber.org/zap"

	"golang.org/x/crypto/bcrypt"

	"gorm.io/gorm"
)

type userService struct {
	db         *gorm.DB
	cache      utilsCache.UserCache // renamed from userCache
	logger     *zap.Logger
	helperUser *helper.HelperUser
}

func NewUserService(
	db *gorm.DB,
	cache utilsCache.UserCache,
	logger *zap.Logger,
) *userService {
	return &userService{
		db:         db,
		cache:      cache,
		logger:     logger,
		helperUser: helper.NewHelperUser(db),
	}
}

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrCacheUnavailable = errors.New("cache service unavailable")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrUserNotFound     = errors.New("user not found")
	ErrEmailExists      = errors.New("email already exists")
)

func (us *userService) Register(ctx context.Context, req dtousergo.RegisterRequest) (*dtousergo.RegisterRespone, error) {
	if req.UserEmail == "" || req.Password == "" {
		return nil, fmt.Errorf("email or password must not be empty")
	}

	if !us.helperUser.IsValidEmailStrict(req.UserEmail) {
		return nil, fmt.Errorf("invalid email format")
	}
	// Kiểm tra xem email đã tồn tại chưa
	var existingUser entityuser.User
	err := us.db.WithContext(ctx).Where("user_email = ?", req.UserEmail).First(&existingUser).Error
	if err == nil {
		return nil, fmt.Errorf("email already registered")
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("database error: %v", err)
	}

	// Hash mật khẩu
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("password hash error: %v", err)
	}

	// Tạo user mới
	newUser := entityuser.User{
		UserEmail:    req.UserEmail,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
		UserRole:     "user",
	}

	if err := us.db.WithContext(ctx).Create(&newUser).Error; err != nil {
		global.Log.Error("Không thể thêm user", zap.Error(err))
		return nil, fmt.Errorf("failed to create user")
	}
	// 4. 🎯 PUBLISH USER REGISTERED EVENT
	event := kafka.UserRegisteredEvent{
		UserID:   newUser.UserID.String(),
		Email:    newUser.UserEmail,
		FullName: newUser.FullName,
	}

	if err := kafka.PublishUserRegisteredEvent(event); err != nil {
		log.Printf("⚠️ Failed to publish user registered event: %v", err)
		// Không return error để không fail registration
	}

	global.Log.Info("Thêm user thành công", zap.String("user_id", newUser.UserID.String()))

	resp := &dtousergo.RegisterRespone{
		UserID:    newUser.UserID,
		UserEmail: newUser.UserEmail,
		FullName:  newUser.FullName,
	}

	return resp, nil
}

func (us *userService) Login(ctx context.Context, req dtousergo.LoginRequest) (*dtousergo.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("email or password must not be empty")
	}

	// Tìm user theo email
	var user entityuser.User
	if err := us.db.WithContext(ctx).Where("user_email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid email or password")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	// So sánh password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Tạo access token (thời gian ngắn)
	token, err := utils.GenerateJWT(user.UserID, time.Now().Add(15*time.Minute).Unix())
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token")
	}

	// Use UserCache to store token if available
	if us.cache != nil {
		err = us.cache.SetToken(ctx, token, user.UserID, 15*time.Minute)
		if err != nil {
			us.logger.Error("Failed to save token using UserCache", zap.Error(err))
		} else {
			us.logger.Info("Token saved using UserCache successfully",
				zap.String("user_id", user.UserID.String()))
		}
	} else {
		us.logger.Warn("UserCache not available, token not cached")
	}

	// Tạo refresh token
	refreshToken, err := utils.GenerateJWT(user.UserID, time.Now().Add(time.Hour*72).Unix())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token")
	}
	hashed := utils.Hash(refreshToken)

	// Giới hạn số lượng refresh token (logic cũ giữ nguyên)
	const maxTokens = 5
	var tokens []entityuser.UserToken
	if err := us.db.WithContext(ctx).
		Where("user_id = ?", user.UserID).
		Order("created_at asc").
		Find(&tokens).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch existing refresh tokens: %v", err)
	}

	if len(tokens) >= maxTokens {
		tokensToDelete := tokens[:len(tokens)-maxTokens+1]
		ids := make([]uuid.UUID, len(tokensToDelete))
		for i, t := range tokensToDelete {
			ids[i] = t.TokenID
		}
		if err := us.db.WithContext(ctx).Where("token_id IN ?", ids).Delete(&entityuser.UserToken{}).Error; err != nil {
			return nil, fmt.Errorf("failed to delete old refresh tokens: %v", err)
		}
	}

	// Lưu refresh token mới
	refreshEntity := entityuser.UserToken{
		UserID:    user.UserID,
		TokenHash: hashed,
		ExpiresAt: time.Now().Add(time.Hour * 72),
		TokenType: "refresh",
		IsRevoked: false,
		CreatedAt: time.Now(),
	}

	if err := us.db.WithContext(ctx).Create(&refreshEntity).Error; err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %v", err)
	}

	// Trả về response
	resp := &dtousergo.LoginResponse{
		UserID:      user.UserID,
		FullName:    user.FullName,
		Token:       token,
		RefeshToken: refreshToken,
	}

	return resp, nil
}

func (us *userService) GetUserByID(ctx context.Context, userid uuid.UUID) (*dtousergo.UserProfileResponse, error) {
	if userid == uuid.Nil {
		return nil, fmt.Errorf("userid must not be empty")
	}

	// ✅ Query từ entity User (table users), KHÔNG phải từ DTO
	var user entityuser.User
	if err := us.db.WithContext(ctx).Where("user_id = ?", userid).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	// ✅ Chuyển đổi từ Entity sang DTO Response
	resp := &dtousergo.UserProfileResponse{
		UserID:        user.UserID,
		FullName:      user.FullName,
		UserEmail:     user.UserEmail,
		UserCreatedAt: user.UserCreatedAt,
		UserUpdatedAt: user.UserUpdatedAt,
	}

	// ✅ Handle pointer fields an toàn
	if user.PhoneNumber != nil {
		resp.PhoneNumber = *user.PhoneNumber
	}
	if user.AvatarURL != nil {
		resp.AvatarURL = *user.AvatarURL
	}
	if user.Gender != nil {
		resp.Gender = *user.Gender
	}
	if user.BioDescription != nil {
		resp.BioDescription = *user.BioDescription
	}

	return resp, nil
}

func (us *userService) UpdateInforUser(ctx context.Context, req dtousergo.InforUserUpdate, userId uuid.UUID) (*dtousergo.UserProfileResponse, error) {
	if userId == uuid.Nil {
		return nil, fmt.Errorf("userID must not be empty")
	}

	var user entityuser.User
	if err := us.db.WithContext(ctx).First(&user, "user_id = ?", userId).Error; err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	// Cập nhật các field nếu có
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.PhoneNumber != nil {
		user.PhoneNumber = req.PhoneNumber
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}
	if req.Gender != nil {
		user.Gender = req.Gender
	}
	if req.BioDescription != nil {
		user.BioDescription = req.BioDescription
	}

	user.UserUpdatedAt = time.Now()

	if err := us.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %v", err)
	}

	// FIX: Map sang DTO trả về an toàn với pointer fields
	res := &dtousergo.UserProfileResponse{
		UserID:        user.UserID,
		FullName:      user.FullName,
		UserEmail:     user.UserEmail,
		UserCreatedAt: user.UserCreatedAt,
		UserUpdatedAt: user.UserUpdatedAt,
	}

	// Handle pointer fields safely
	if user.PhoneNumber != nil {
		res.PhoneNumber = *user.PhoneNumber
	}
	if user.AvatarURL != nil {
		res.AvatarURL = *user.AvatarURL
	}
	if user.Gender != nil {
		res.Gender = *user.Gender
	}
	if user.BioDescription != nil {
		res.BioDescription = *user.BioDescription
	}

	return res, nil
}

func (us *userService) RefeshToken(ctx context.Context, refeshtoken string) (string, error) {
	hashed := utils.Hash(refeshtoken)
	var token entity.UserToken
	err := us.db.Where("token_hash = ? AND is_revoked = false AND expires_at > ? AND token_type = refresh", hashed, time.Now()).First(&token).Error
	if err != nil {
		return "", fmt.Errorf("invalid or expired refresh token")
	}

	// Tạo access token mới
	newToken, err := utils.GenerateJWT(token.UserID, time.Now().Add(time.Minute*15).Unix())
	if err != nil {
		return "", fmt.Errorf("failed to generate new access token")
	}

	// ✅ LƯU ACCESS TOKEN MỚI VÀO REDIS
	if us.cache != nil {
		err = us.cache.SetToken(ctx, newToken, token.UserID, time.Minute*15)
		if err != nil {
			us.logger.Error("Failed to save new token using UserCache", zap.Error(err))
			return "", fmt.Errorf("failed to save token to cache")
		}
		us.logger.Info("New token saved using UserCache successfully",
			zap.String("user_id", token.UserID.String()))
	} else {
		us.logger.Warn("UserCache not available, token not cached")
	}

	return newToken, nil
}

func (us *userService) Logout(ctx context.Context, token string, userID uuid.UUID) error {
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("token must not be empty")
	}

	if userID == uuid.Nil {
		return fmt.Errorf("userID must not be empty")
	}

	// Use UserCache to invalidate token
	if us.cache != nil {
		err := us.cache.InvalidateToken(ctx, token)
		if err != nil {
			us.logger.Error("Failed to invalidate token using UserCache", zap.Error(err))
			return fmt.Errorf("failed to logout: %w", err)
		}
		us.logger.Info("User logged out successfully using UserCache",
			zap.String("userID", userID.String()))
		return nil
	}

	// If UserCache is not available, log warning and return error
	us.logger.Error("UserCache not initialized, cannot logout")
	return fmt.Errorf("cache service unavailable")
}

//========================= REFACTORED VALIDATE TOKEN METHOD =========================

func (us *userService) ValidateToken(ctx context.Context, token string) (uuid.UUID, error) {
	// 1. Input validation
	if strings.TrimSpace(token) == "" {
		return uuid.Nil, ErrInvalidToken
	}

	// 2. Check if UserCache is available
	if us.cache == nil {
		us.logger.Error("User cache not initialized")
		return uuid.Nil, ErrCacheUnavailable
	}

	// 3. Create Redis key
	key := "auth:" + token

	// 4. Check if token exists using UserCache
	exists, err := us.cache.CheckTokenExists(ctx, key)
	if err != nil {
		us.logger.Error("Error checking token existence", zap.Error(err))
		return uuid.Nil, ErrCacheUnavailable
	}

	if !exists {
		us.logger.Debug("Token not found in cache", zap.String("key", key))
		return uuid.Nil, ErrInvalidToken
	}

	// 5. Get userID from cache using UserCache
	userID, err := us.cache.GetTokenFromCache(ctx, key)
	if err != nil {
		// Handle redis.Nil specifically
		if errors.Is(err, redis.Nil) {
			us.logger.Debug("Token not found in cache", zap.String("key", key))
			return uuid.Nil, ErrInvalidToken
		}

		us.logger.Error("Error retrieving user from cache", zap.Error(err))
		return uuid.Nil, ErrCacheUnavailable
	}

	// 6. Validate userID
	if userID == uuid.Nil {
		us.logger.Error("Invalid userID retrieved from cache")
		return uuid.Nil, ErrInvalidToken
	}

	us.logger.Debug("Token validated successfully", zap.String("userID", userID.String()))
	return userID, nil
}

// Optional: Add method to check Redis health using UserCache
func (us *userService) IsRedisHealthy(ctx context.Context) bool {
	if us.cache == nil {
		return false
	}
	return us.cache.IsRedisHealthy(ctx)
}

//========================= ADDITIONAL USER FEATURES =========================

// ChangePassword - Đổi mật khẩu
func (us *userService) ChangePassword(ctx context.Context, req dtousergo.ChangePasswordRequest, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return fmt.Errorf("userID must not be empty")
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		return fmt.Errorf("old password and new password must not be empty")
	}

	if req.OldPassword == req.NewPassword {
		return fmt.Errorf("new password must be different from old password")
	}

	// Validate password strength
	if err := us.helperUser.ValidatePasswordStrength(req.NewPassword); err != nil {
		return err
	}

	// Tìm user
	var user entityuser.User
	if err := us.db.WithContext(ctx).Where("user_id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrUserNotFound
		}
		return fmt.Errorf("database error: %v", err)
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return ErrInvalidPassword
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %v", err)
	}

	// Update password
	user.PasswordHash = string(hashedPassword)
	user.UserUpdatedAt = time.Now()

	if err := us.db.WithContext(ctx).Save(&user).Error; err != nil {
		return fmt.Errorf("failed to update password: %v", err)
	}

	// Revoke all refresh tokens for security
	if err := us.db.WithContext(ctx).Model(&entityuser.UserToken{}).
		Where("user_id = ?", userID).
		Update("is_revoked", true).Error; err != nil {
		us.logger.Error("Failed to revoke refresh tokens after password change", zap.Error(err))
	}

	// Invalidate all active sessions
	if us.cache != nil {
		if err := us.cache.InvalidateAllUserTokens(ctx, userID); err != nil {
			us.logger.Error("Failed to invalidate user tokens", zap.Error(err))
		}
	}

	us.logger.Info("Password changed successfully", zap.String("userID", userID.String()))
	return nil
}

// ResetPassword - Reset mật khẩu qua email
func (us *userService) ResetPassword(ctx context.Context, req dtousergo.ResetPasswordRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email must not be empty")
	}

	// Tìm user theo email
	var user entityuser.User
	if err := us.db.WithContext(ctx).Where("user_email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Don't reveal if email exists for security
			return nil
		}
		return fmt.Errorf("database error: %v", err)
	}

	// Generate reset token
	resetToken := us.helperUser.GenerateSecureToken(64)
	hashedToken := utils.Hash(resetToken)

	// Save reset token (expires in 1 hour)
	resetEntity := entityuser.UserToken{
		UserID:    user.UserID,
		TokenHash: hashedToken,
		TokenType: "password_reset",
		ExpiresAt: time.Now().Add(time.Hour),
		IsUsed:    false,
	}

	if err := us.db.WithContext(ctx).Create(&resetEntity).Error; err != nil {
		return fmt.Errorf("failed to save reset token: %v", err)
	}

	// Send reset email (integrate with email service)
	if err := us.sendResetEmail(user.UserEmail, resetToken); err != nil {
		us.logger.Error("Failed to send reset email", zap.Error(err))
		return fmt.Errorf("failed to send reset email")
	}

	us.logger.Info("Password reset email sent", zap.String("email", user.UserEmail))
	return nil
}

// ConfirmResetPassword - Xác nhận reset mật khẩu
func (us *userService) ConfirmResetPassword(ctx context.Context, req dtousergo.ConfirmResetPasswordRequest) error {
	if req.Token == "" || req.NewPassword == "" {
		return fmt.Errorf("token and new password must not be empty")
	}

	// Validate password strength
	if err := us.helperUser.ValidatePasswordStrength(req.NewPassword); err != nil {
		return err
	}

	hashedToken := utils.Hash(req.Token)

	// Find valid reset token
	var resetToken entityuser.UserToken
	if err := us.db.WithContext(ctx).Where("token_hash = ? AND is_used = false AND expires_at > ? AND token_type = password_reset",
		hashedToken, time.Now()).First(&resetToken).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("invalid or expired reset token")
		}
		return fmt.Errorf("database error: %v", err)
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// Update password
	if err := us.db.WithContext(ctx).Model(&entityuser.User{}).
		Where("user_id = ?", resetToken.UserID).
		Updates(map[string]interface{}{
			"password_hash":   string(hashedPassword),
			"user_updated_at": time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("failed to update password: %v", err)
	}

	// Mark token as used
	resetToken.IsUsed = true
	if err := us.db.WithContext(ctx).Save(&resetToken).Error; err != nil {
		us.logger.Error("Failed to mark reset token as used", zap.Error(err))
	}

	// Revoke all refresh tokens
	if err := us.db.WithContext(ctx).Model(&entityuser.UserToken{}).
		Where("user_id = ?", resetToken.UserID).
		Update("is_revoked", true).Error; err != nil {
		us.logger.Error("Failed to revoke refresh tokens", zap.Error(err))
	}

	// Invalidate all active sessions
	if us.cache != nil {
		if err := us.cache.InvalidateAllUserTokens(ctx, resetToken.UserID); err != nil {
			us.logger.Error("Failed to invalidate user tokens", zap.Error(err))
		}
	}

	us.logger.Info("Password reset successfully", zap.String("userID", resetToken.UserID.String()))
	return nil
}

// DeleteAccount - Xóa tài khoản
func (us *userService) DeleteAccount(ctx context.Context, req dtousergo.DeleteAccountRequest, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return fmt.Errorf("userID must not be empty")
	}

	if req.Password == "" {
		return fmt.Errorf("password must not be empty")
	}

	// Verify user and password
	var user entityuser.User
	if err := us.db.WithContext(ctx).Where("user_id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrUserNotFound
		}
		return fmt.Errorf("database error: %v", err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return ErrInvalidPassword
	}

	// Start transaction
	tx := us.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete refresh tokens
	if err := tx.Where("user_id = ?", userID).Delete(&entityuser.UserToken{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete refresh tokens: %v", err)
	}

	// Delete password reset tokens
	if err := tx.Where("user_id = ?", userID).Delete(&entityuser.UserToken{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete reset tokens: %v", err)
	}

	// Delete user
	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Invalidate all active sessions
	if us.cache != nil {
		if err := us.cache.InvalidateAllUserTokens(ctx, userID); err != nil {
			us.logger.Error("Failed to invalidate user tokens", zap.Error(err))
		}
	}

	// Publish user deleted event
	// event := kafka.UserDeletedEvent{
	// 	UserID: userID.String(),
	// 	Email:  user.UserEmail,
	// }
	// if err := kafka.PublishUserDeletedEvent(event); err != nil {
	// 	us.logger.Error("Failed to publish user deleted event", zap.Error(err))
	// }

	us.logger.Info("User account deleted successfully", zap.String("userID", userID.String()))
	return nil
}

// GetUsersByRole - Lấy danh sách user theo role (admin only)
func (us *userService) GetUsersByRole(ctx context.Context, role string, page, limit int) (*dtousergo.UserListResponse, error) {
	if role == "" {
		return nil, fmt.Errorf("role must not be empty")
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var users []entityuser.User
	var total int64

	// Count total
	if err := us.db.WithContext(ctx).Model(&entityuser.User{}).
		Where("user_role = ?", role).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count users: %v", err)
	}

	// Get users
	if err := us.db.WithContext(ctx).Where("user_role = ?", role).
		Order("user_created_at desc").
		Offset(offset).Limit(limit).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get users: %v", err)
	}

	// Convert to response
	userList := make([]dtousergo.UserProfileResponse, len(users))
	for i, user := range users {
		userList[i] = dtousergo.UserProfileResponse{
			UserID:        user.UserID,
			FullName:      user.FullName,
			UserEmail:     user.UserEmail,
			UserCreatedAt: user.UserCreatedAt,
			UserUpdatedAt: user.UserUpdatedAt,
		}

		// Handle pointer fields
		if user.PhoneNumber != nil {
			userList[i].PhoneNumber = *user.PhoneNumber
		}
		if user.AvatarURL != nil {
			userList[i].AvatarURL = *user.AvatarURL
		}
		if user.Gender != nil {
			userList[i].Gender = *user.Gender
		}
		if user.BioDescription != nil {
			userList[i].BioDescription = *user.BioDescription
		}
	}

	return &dtousergo.UserListResponse{
		Users:      userList,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: (total + int64(limit) - 1) / int64(limit),
	}, nil
}

// DeactivateUser - Vô hiệu hóa tài khoản (admin only)
func (us *userService) DeactivateUser(ctx context.Context, targetUserID uuid.UUID) error {
	if targetUserID == uuid.Nil {
		return fmt.Errorf("target user ID must not be empty")
	}

	var user entityuser.User
	if err := us.db.WithContext(ctx).Where("user_id = ?", targetUserID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrUserNotFound
		}
		return fmt.Errorf("database error: %v", err)
	}

	// Update user status
	user.IsActive = false
	user.UserUpdatedAt = time.Now()

	if err := us.db.WithContext(ctx).Save(&user).Error; err != nil {
		return fmt.Errorf("failed to deactivate user: %v", err)
	}

	// Revoke all refresh tokens
	if err := us.db.WithContext(ctx).Model(&entityuser.UserToken{}).
		Where("user_id = ?", targetUserID).
		Update("is_revoked", true).Error; err != nil {
		us.logger.Error("Failed to revoke refresh tokens", zap.Error(err))
	}

	// Invalidate all active sessions
	if us.cache != nil {
		if err := us.cache.InvalidateAllUserTokens(ctx, targetUserID); err != nil {
			us.logger.Error("Failed to invalidate user tokens", zap.Error(err))
		}
	}

	us.logger.Info("User deactivated successfully", zap.String("userID", targetUserID.String()))
	return nil
}

// ActivateUser - Kích hoạt lại tài khoản (admin only)
func (us *userService) ActivateUser(ctx context.Context, targetUserID uuid.UUID) error {
	if targetUserID == uuid.Nil {
		return fmt.Errorf("target user ID must not be empty")
	}

	var user entityuser.User
	if err := us.db.WithContext(ctx).Where("user_id = ?", targetUserID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrUserNotFound
		}
		return fmt.Errorf("database error: %v", err)
	}

	// Update user status
	user.IsActive = true
	user.UserUpdatedAt = time.Now()

	if err := us.db.WithContext(ctx).Save(&user).Error; err != nil {
		return fmt.Errorf("failed to activate user: %v", err)
	}

	us.logger.Info("User activated successfully", zap.String("userID", targetUserID.String()))
	return nil
}

// UpdateUserRole - Cập nhật role của user (admin only)
func (us *userService) UpdateUserRole(ctx context.Context, targetUserID uuid.UUID, newRole string) error {
	if targetUserID == uuid.Nil {
		return fmt.Errorf("target user ID must not be empty")
	}

	if newRole == "" {
		return fmt.Errorf("new role must not be empty")
	}

	// Validate role
	validRoles := map[string]bool{
		"user":      true,
		"admin":     true,
		"moderator": true,
	}
	if !validRoles[newRole] {
		return fmt.Errorf("invalid role: %s", newRole)
	}

	var user entityuser.User
	if err := us.db.WithContext(ctx).Where("user_id = ?", targetUserID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrUserNotFound
		}
		return fmt.Errorf("database error: %v", err)
	}

	oldRole := user.UserRole
	user.UserRole = newRole
	user.UserUpdatedAt = time.Now()

	if err := us.db.WithContext(ctx).Save(&user).Error; err != nil {
		return fmt.Errorf("failed to update user role: %v", err)
	}

	// // Publish role change event
	// event := kafka.UserRoleChangedEvent{
	// 	UserID:  targetUserID.String(),
	// 	OldRole: oldRole,
	// 	NewRole: newRole,
	// }
	// if err := kafka.PublishUserRoleChangedEvent(event); err != nil {
	// 	us.logger.Error("Failed to publish role changed event", zap.Error(err))
	// }

	us.logger.Info("User role updated successfully",
		zap.String("userID", targetUserID.String()),
		zap.String("oldRole", oldRole),
		zap.String("newRole", newRole))
	return nil
}

// SearchUsers - Tìm kiếm user
func (us *userService) SearchUsers(ctx context.Context, req dtousergo.SearchUsersRequest) (*dtousergo.UserListResponse, error) {
	if req.Query == "" {
		return nil, fmt.Errorf("search query must not be empty")
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}

	offset := (req.Page - 1) * req.Limit

	var users []entityuser.User
	var total int64

	// Build search query
	searchQuery := "%" + strings.ToLower(req.Query) + "%"

	baseQuery := us.db.WithContext(ctx).Model(&entityuser.User{}).
		Where("LOWER(full_name) LIKE ? OR LOWER(user_email) LIKE ?", searchQuery, searchQuery)

	// Filter by role if specified
	if req.Role != "" {
		baseQuery = baseQuery.Where("user_role = ?", req.Role)
	}

	// Filter by active status if specified
	if req.IsActive != nil {
		baseQuery = baseQuery.Where("is_active = ?", *req.IsActive)
	}

	// Count total
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count users: %v", err)
	}

	// Get users
	if err := baseQuery.Order("user_created_at desc").
		Offset(offset).Limit(req.Limit).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to search users: %v", err)
	}

	// Convert to response
	userList := make([]dtousergo.UserProfileResponse, len(users))
	for i, user := range users {
		userList[i] = dtousergo.UserProfileResponse{
			UserID:        user.UserID,
			FullName:      user.FullName,
			UserEmail:     user.UserEmail,
			UserCreatedAt: user.UserCreatedAt,
			UserUpdatedAt: user.UserUpdatedAt,
		}

		// Handle pointer fields
		if user.PhoneNumber != nil {
			userList[i].PhoneNumber = *user.PhoneNumber
		}
		if user.AvatarURL != nil {
			userList[i].AvatarURL = *user.AvatarURL
		}
		if user.Gender != nil {
			userList[i].Gender = *user.Gender
		}
		if user.BioDescription != nil {
			userList[i].BioDescription = *user.BioDescription
		}
	}

	return &dtousergo.UserListResponse{
		Users:      userList,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: (total + int64(req.Limit) - 1) / int64(req.Limit),
	}, nil
}

// LogoutAllSessions - Đăng xuất tất cả sessions
func (us *userService) LogoutAllSessions(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return fmt.Errorf("userID must not be empty")
	}

	// Revoke all refresh tokens
	if err := us.db.WithContext(ctx).Model(&entityuser.UserToken{}).
		Where("user_id = ?", userID).
		Update("is_revoked", true).Error; err != nil {
		return fmt.Errorf("failed to revoke refresh tokens: %v", err)
	}

	// Invalidate all active sessions
	if us.cache != nil {
		if err := us.cache.InvalidateAllUserTokens(ctx, userID); err != nil {
			us.logger.Error("Failed to invalidate user tokens", zap.Error(err))
			return fmt.Errorf("failed to invalidate active sessions: %v", err)
		}
	}

	us.logger.Info("All sessions logged out successfully", zap.String("userID", userID.String()))
	return nil
}

// UpdateEmail - Cập nhật email
func (us *userService) UpdateEmail(ctx context.Context, req dtousergo.UpdateEmailRequest, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return fmt.Errorf("userID must not be empty")
	}

	if req.NewEmail == "" || req.Password == "" {
		return fmt.Errorf("new email and password must not be empty")
	}

	// Validate email format
	if !us.helperUser.IsValidEmailStrict(req.NewEmail) {
		return fmt.Errorf("invalid email format")
	}

	// Check if new email already exists
	var existingUser entityuser.User
	if err := us.db.WithContext(ctx).Where("user_email = ? AND user_id != ?", req.NewEmail, userID).First(&existingUser).Error; err == nil {
		return ErrEmailExists
	}

	// Get current user
	var user entityuser.User
	if err := us.db.WithContext(ctx).Where("user_id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrUserNotFound
		}
		return fmt.Errorf("database error: %v", err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return ErrInvalidPassword
	}

	// Update email
	user.UserEmail = req.NewEmail
	user.UserUpdatedAt = time.Now()

	if err := us.db.WithContext(ctx).Save(&user).Error; err != nil {
		return fmt.Errorf("failed to update email: %v", err)
	}

	// Publish email changed event
	// event := kafka.UserEmailChangedEvent{
	// 	UserID:   userID.String(),
	// 	OldEmail: user.UserEmail,
	// 	NewEmail: req.NewEmail,
	// }
	// if err := kafka.PublishUserEmailChangedEvent(event); err != nil {
	// 	us.logger.Error("Failed to publish email changed event", zap.Error(err))
	// }

	us.logger.Info("Email updated successfully",
		zap.String("userID", userID.String()),
		zap.String("newEmail", req.NewEmail))
	return nil
}

// GetActiveTokens - Lấy danh sách token đang hoạt động
func (us *userService) GetActiveTokens(ctx context.Context, userID uuid.UUID) (*dtousergo.ActiveTokensResponse, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("userID must not be empty")
	}

	var tokens []entityuser.UserToken
	if err := us.db.WithContext(ctx).
		Where("user_id = ? AND is_revoked = false AND expires_at > ?", userID, time.Now()).
		Order("token_created_at desc").
		Find(&tokens).Error; err != nil {
		return nil, fmt.Errorf("failed to get active tokens: %v", err)
	}

	tokenList := make([]dtousergo.TokenInfo, len(tokens))
	for i, token := range tokens {
		tokenList[i] = dtousergo.TokenInfo{
			TokenID:   token.TokenID,
			CreatedAt: token.CreatedAt,
			ExpiresAt: token.ExpiresAt,
			IsActive:  !token.IsRevoked,
		}
	}

	return &dtousergo.ActiveTokensResponse{
		Tokens: tokenList,
		Total:  len(tokens),
	}, nil
}

// RevokeToken - Thu hồi một token cụ thể
func (us *userService) RevokeToken(ctx context.Context, tokenID uuid.UUID, userID uuid.UUID) error {
	if tokenID == uuid.Nil || userID == uuid.Nil {
		return fmt.Errorf("tokenID and userID must not be empty")
	}

	var token entityuser.UserToken
	if err := us.db.WithContext(ctx).
		Where("refresh_token_id = ? AND user_id = ?", tokenID, userID).
		First(&token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("token not found")
		}
		return fmt.Errorf("database error: %v", err)
	}

	// Mark as revoked
	token.IsRevoked = true
	if err := us.db.WithContext(ctx).Save(&token).Error; err != nil {
		return fmt.Errorf("failed to revoke token: %v", err)
	}

	us.logger.Info("Token revoked successfully",
		zap.String("tokenID", tokenID.String()),
		zap.String("userID", userID.String()))
	return nil
}

// ========================= HELPER METHODS =========================
// sendResetEmail - Gửi email reset password
func (us *userService) sendResetEmail(email, resetToken string) error {
	// Implement email sending logic here
	// This could integrate with services like SendGrid, AWS SES, etc.

	us.logger.Info("Reset email would be sent",
		zap.String("email", email),
		zap.String("token", resetToken))

	// For now, just log the reset token
	// In production, you would send an actual email
	return nil
}
