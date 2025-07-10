package users

import (
	"context"

	dtousergo "cbs_backend/internal/modules/users/dto.user.go"
	"cbs_backend/utils/cache"

	"go.uber.org/zap"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

var (
	iUserService IUser
)

func InitUserService(db *gorm.DB, cache cache.UserCache, logger *zap.Logger) {
	iUserService = NewUserService(db, cache, logger)
}

func User() IUser {
	if iUserService == nil {
		panic("AuthService not initialized. Call InitAuthService(db) first.")
	}
	return iUserService
}

type IUser interface {
	UpdateInforUser(ctx context.Context, req dtousergo.InforUserUpdate, userid uuid.UUID) (*dtousergo.UserProfileResponse, error)
	GetUserByID(ctx context.Context, userid uuid.UUID) (*dtousergo.UserProfileResponse, error)
	Register(ctx context.Context, req dtousergo.RegisterRequest) (*dtousergo.RegisterRespone, error)
	Login(ctx context.Context, req dtousergo.LoginRequest) (*dtousergo.LoginResponse, error)
	ValidateToken(ctx context.Context, token string) (uuid.UUID, error)
	RefeshToken(ctx context.Context, refeshtoken string) (string, error)
	Logout(ctx context.Context, token string, userID uuid.UUID) error
	ChangePassword(ctx context.Context, req dtousergo.ChangePasswordRequest, userID uuid.UUID) error
	ResetPassword(ctx context.Context, req dtousergo.ResetPasswordRequest) error
	ConfirmResetPassword(ctx context.Context, req dtousergo.ConfirmResetPasswordRequest) error
	DeleteAccount(ctx context.Context, req dtousergo.DeleteAccountRequest, userID uuid.UUID) error
	GetUsersByRole(ctx context.Context, role string, page, limit int) (*dtousergo.UserListResponse, error)
	DeactivateUser(ctx context.Context, targetUserID uuid.UUID) error
	ActivateUser(ctx context.Context, targetUserID uuid.UUID) error
	LogoutAllSessions(ctx context.Context, userID uuid.UUID) error
	UpdateUserRole(ctx context.Context, targetUserID uuid.UUID, newRole string) error
	SearchUsers(ctx context.Context, req dtousergo.SearchUsersRequest) (*dtousergo.UserListResponse, error)
	UpdateEmail(ctx context.Context, req dtousergo.UpdateEmailRequest, userID uuid.UUID) error
	GetActiveTokens(ctx context.Context, userID uuid.UUID) (*dtousergo.ActiveTokensResponse, error)
	RevokeToken(ctx context.Context, tokenID uuid.UUID, userID uuid.UUID) error
}
