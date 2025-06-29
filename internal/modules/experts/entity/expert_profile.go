package entity

import (
	"time"

	entityUser "cbs_backend/internal/modules/users/entity"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ExpertProfile represents tbl_expert_profiles table
type ExpertProfile struct {
	ExpertProfileID    uuid.UUID      `json:"expert_profile_id" db:"expert_profile_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID             uuid.UUID      `json:"user_id" db:"user_id" gorm:"type:uuid;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	SpecializationList pq.StringArray `json:"specialization_list" db:"specialization_list" gorm:"type:text[]"`
	ExperienceYears    *int           `json:"experience_years,omitempty" db:"experience_years"`
	ExpertBio          *string        `json:"expert_bio,omitempty" db:"expert_bio" gorm:"type:text"`
	ConsultationFee    *float64       `json:"consultation_fee,omitempty" db:"consultation_fee" gorm:"type:decimal(10,2)"`
	AverageRating      float64        `json:"average_rating" db:"average_rating" gorm:"type:decimal(3,2);default:0.00"`
	TotalReviews       int            `json:"total_reviews" db:"total_reviews" gorm:"default:0"`
	IsVerified         bool           `json:"is_verified" db:"is_verified" gorm:"default:false"`
	LicenseNumber      *string        `json:"license_number,omitempty" db:"license_number" gorm:"type:varchar(100)"`
	AvailableOnline    bool           `json:"available_online" db:"available_online" gorm:"default:true"`
	AvailableOffline   bool           `json:"available_offline" db:"available_offline" gorm:"default:true"`
	ExpertCreatedAt    time.Time      `json:"expert_created_at" db:"expert_created_at" gorm:"default:CURRENT_TIMESTAMP"`
	ExpertUpdatedAt    time.Time      `json:"expert_updated_at" db:"expert_updated_at" gorm:"default:CURRENT_TIMESTAMP"`

	// 👇 Relationships: nên để *pointer và chỉ dùng nếu thực sự cần preload
	User *entityUser.User `json:"user" gorm:"foreignKey:UserID;references:UserID"`

	// Relationships
	// User entity.User `json:"user" gorm:"foreignKey:UserID"`
	WorkingHours     []ExpertWorkingHour     `json:"working_hours,omitempty" gorm:"foreignKey:ExpertProfileID"`
	UnavailableTimes []ExpertUnavailableTime `json:"unavailable_times,omitempty" gorm:"foreignKey:ExpertProfileID"`
	// Bookings            []ConsultationBooking   `json:"bookings,omitempty" gorm:"foreignKey:ExpertProfileID"`
	// Reviews             []ConsultationReview    `json:"reviews,omitempty" gorm:"foreignKey:ExpertProfileID"`
	Specializations []ExpertSpecialization `json:"specializations,omitempty" gorm:"foreignKey:ExpertProfileID"`
	// PaymentTransactions []PaymentTransaction    `json:"payment_transactions,omitempty" gorm:"foreignKey:ExpertProfileID"`
	// PricingConfigs      []PricingConfig         `json:"pricing_configs,omitempty" gorm:"foreignKey:ExpertProfileID"`
}

func (ExpertProfile) TableName() string {
	return "tbl_expert_profiles"
}
