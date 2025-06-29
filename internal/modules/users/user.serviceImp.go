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

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"go.uber.org/zap"

	"golang.org/x/crypto/bcrypt"

	"gorm.io/gorm"
)

type userService struct {
	db     *gorm.DB
	cache  utilsCache.UserCache // renamed from userCache
	logger *zap.Logger
}

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrCacheUnavailable = errors.New("cache service unavailable")
)

func NewUserService(
	db *gorm.DB,
	cache utilsCache.UserCache,
	logger *zap.Logger,
) *userService {
	return &userService{db: db, cache: cache, logger: logger}
}

func (us *userService) Register(ctx context.Context, req dtousergo.RegisterRequest) (*dtousergo.RegisterRespone, error) {
	if req.UserEmail == "" || req.Password == "" {
		return nil, fmt.Errorf("email or password must not be empty")
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
	var tokens []entityuser.UserRefreshToken
	if err := us.db.WithContext(ctx).
		Where("user_id = ?", user.UserID).
		Order("token_created_at asc").
		Find(&tokens).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch existing refresh tokens: %v", err)
	}

	if len(tokens) >= maxTokens {
		tokensToDelete := tokens[:len(tokens)-maxTokens+1]
		ids := make([]uuid.UUID, len(tokensToDelete))
		for i, t := range tokensToDelete {
			ids[i] = t.RefreshTokenID
		}
		if err := us.db.WithContext(ctx).Where("refresh_token_id IN ?", ids).Delete(&entityuser.UserRefreshToken{}).Error; err != nil {
			return nil, fmt.Errorf("failed to delete old refresh tokens: %v", err)
		}
	}

	// Lưu refresh token mới
	refreshEntity := entityuser.UserRefreshToken{
		UserID:         user.UserID,
		TokenHash:      hashed,
		ExpiresAt:      time.Now().Add(time.Hour * 72),
		IsRevoked:      false,
		TokenCreatedAt: time.Now(),
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
	var token entity.UserRefreshToken
	err := us.db.Where("token_hash = ? AND is_revoked = false AND expires_at > ?", hashed, time.Now()).First(&token).Error
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
