package email

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserResolver struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewUserResolver(db *gorm.DB, logger *zap.Logger) *UserResolver {
	return &UserResolver{
		db:     db,
		logger: logger,
	}
}

func (ur *UserResolver) GetUserEmail(userID string) string {
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		ur.logger.Error("Invalid UUID format", zap.String("userID", userID), zap.Error(err))
		return ""
	}

	var email string
	err = ur.db.Table("tbl_users").
		Select("user_email").
		Where("user_id = ?", parsedID).
		Limit(1).
		Scan(&email).Error
	if err != nil {
		ur.logger.Error("Failed to get user email", zap.Error(err), zap.String("userID", userID))
		return ""
	}
	return email
}

func (ur *UserResolver) GetDoctorEmail(doctorID string) string {
	parsedID, err := uuid.Parse(doctorID)
	if err != nil {
		ur.logger.Error("Invalid UUID format", zap.String("doctorID", doctorID), zap.Error(err))
		return ""
	}

	var email string
	err = ur.db.Table("tbl_expert_profiles as ep").
		Select("u.user_email").
		Joins("join tbl_users u on ep.user_id = u.user_id").
		Where("ep.expert_profile_id = ?", parsedID).
		Limit(1).
		Scan(&email).Error

	if err != nil {
		ur.logger.Error("Failed to get doctor email via expert profile", zap.Error(err), zap.String("doctorID", doctorID))
		return ""
	}

	return email
}
