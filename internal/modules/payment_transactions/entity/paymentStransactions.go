package entity

import (
	"time"

	"cbs_backend/internal/common"

	"github.com/google/uuid"
)

// PaymentTransaction represents tbl_payment_transactions table
type PaymentTransaction struct {
	TransactionID         uuid.UUID    `json:"transaction_id" db:"transaction_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	BookingID             uuid.UUID    `json:"booking_id" db:"booking_id" gorm:"type:uuid;not null;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	UserID                uuid.UUID    `json:"user_id" db:"user_id" gorm:"type:uuid;not null;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	ExpertProfileID       uuid.UUID    `json:"expert_profile_id" db:"expert_profile_id" gorm:"type:uuid;not null;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Amount                float64      `json:"amount" db:"amount" gorm:"type:decimal(10,2);not null"`
	Currency              string       `json:"currency" db:"currency" gorm:"type:varchar(3);default:'VND'"`
	PaymentMethod         *string      `json:"payment_method,omitempty" db:"payment_method" gorm:"type:varchar(50)"`
	TransactionStatus     string       `json:"transaction_status" db:"transaction_status" gorm:"type:varchar(20);default:'pending';check:transaction_status IN ('pending', 'processing', 'completed', 'failed', 'refunded', 'cancelled')"`
	ExternalTransactionID *string      `json:"external_transaction_id,omitempty" db:"external_transaction_id" gorm:"type:varchar(255)"`
	PaymentGateway        *string      `json:"payment_gateway,omitempty" db:"payment_gateway" gorm:"type:varchar(50)"`
	GatewayResponse       common.JSONB `json:"gateway_response,omitempty" db:"gateway_response" gorm:"type:jsonb"`
	ProcessedAt           *time.Time   `json:"processed_at,omitempty" db:"processed_at"`
	TransactionCreatedAt  time.Time    `json:"transaction_created_at" db:"transaction_created_at" gorm:"default:CURRENT_TIMESTAMP"`

	// // Relationships (optional)
	// // Booking       *entityBooking.ConsultationBooking `json:"booking,omitempty" gorm:"foreignKey:BookingID;references:BookingID"`
	// // User          *entityUser.User                   `json:"user,omitempty" gorm:"foreignKey:UserID;references:UserID"`
	// ExpertProfile *entityExpertProfile.ExpertProfile `json:"expert_profile,omitempty" gorm:"foreignKey:ExpertProfileID;references:ExpertProfileID"`
}
