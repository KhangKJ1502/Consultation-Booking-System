package entity

import (
	"time"

	"cbs_backend/internal/common"

	"github.com/google/uuid"
)

//Bảng entity để load dữ liệu khi cần

// User represents tbl_users table
type User struct {
	UserID               uuid.UUID    `json:"user_id" db:"user_id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserEmail            string       `json:"user_email" db:"user_email" gorm:"type:varchar(255);unique;not null"`
	PasswordHash         string       `json:"-" db:"password_hash" gorm:"type:varchar(255);not null"`
	FullName             string       `json:"full_name" db:"full_name" gorm:"type:varchar(255);not null"`
	PhoneNumber          *string      `json:"phone_number,omitempty" db:"phone_number" gorm:"type:varchar(20)"`
	AvatarURL            *string      `json:"avatar_url,omitempty" db:"avatar_url" gorm:"type:text"`
	Gender               *string      `json:"gender,omitempty" db:"gender" gorm:"type:varchar(10);check:gender IN ('male', 'female', 'other')"`
	UserRole             string       `json:"user_role" db:"user_role" gorm:"type:varchar(20);not null;default:'user';check:user_role IN ('user', 'expert', 'admin')"`
	BioDescription       *string      `json:"bio_description,omitempty" db:"bio_description" gorm:"type:text"`
	IsActive             bool         `json:"is_active" db:"is_active" gorm:"default:true"`
	EmailVerified        bool         `json:"email_verified" db:"email_verified" gorm:"default:false"`
	NotificationSettings common.JSONB `json:"notification_settings" db:"notification_settings" gorm:"type:jsonb;default:'{\"email\": true, \"push\": true, \"telegram\": false}'"`
	UserCreatedAt        time.Time    `json:"user_created_at" db:"user_created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UserUpdatedAt        time.Time    `json:"user_updated_at" db:"user_updated_at" gorm:"default:CURRENT_TIMESTAMP"`
	// Relationship
	// ActivityLog []entityactivitylog.ActivityLog `json:"activity_logs,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	// ExpertProfile *entityExperts.ExpertProfile `json:"expert_profile,omitempty" gorm:"foreignKey:UserID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	// // BookingsAsUser []entityBooking.ConsultationBooking `json:"bookings_as_user,omitempty" gorm:"foreignKey:UserID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	// Uncomment nếu cần dùng:
	// Reviews []entityReview.ConsultationReview `json:"reviews,omitempty" gorm:"foreignKey:ReviewerUserID"`
	// Notifications []entityNotification.SystemNotification `json:"notifications,omitempty" gorm:"foreignKey:RecipientUserID"`
}

func (User) TableName() string {
	return "tbl_users"
}
