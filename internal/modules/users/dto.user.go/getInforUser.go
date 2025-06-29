package dtousergo

import (
	"time"

	"github.com/google/uuid"
)

// type UserProfileRequest struct {
// 	UserID uuid.UUID `json:"user_id" binding:"required"`
// }

type UserProfileResponse struct {
	UserID         uuid.UUID `json:"user_id"`
	FullName       string    `json:"fUll_name"`
	UserEmail      string    `json:"user_email"`
	PhoneNumber    string    `json:"phone_number"`
	AvatarURL      string    `json:"avartar_url"`
	Gender         string    `json:"gender"`
	BioDescription string    `json:"bio_description"`
	UserCreatedAt  time.Time `json:"create_at"`
	UserUpdatedAt  time.Time `json:"user_updated_at"`
}
