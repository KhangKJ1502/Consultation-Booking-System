package email

import (
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
	var user struct {
		Email string `json:"email"`
	}
	err := ur.db.Table("users").Select("email").Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		ur.logger.Error("Failed to get user email", zap.Error(err), zap.String("userID", userID))
		return ""
	}
	return user.Email
}

func (ur *UserResolver) GetDoctorEmail(doctorID string) string {
	// TODO: Implement database query for doctor email
	ur.logger.Warn("getDoctorEmail not implemented", zap.String("doctorID", doctorID))
	return "doctor@example.com"
}
