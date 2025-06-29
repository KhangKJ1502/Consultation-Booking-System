package dtousergo

type InforUserUpdate struct {
	FullName string `json:"full_name"`
	// UserEmail string `json:"user_email" binding:"required, email"`
	PhoneNumber    *string `json:"phone_number"`
	AvatarURL      *string `json:"avatar_url"`
	Gender         *string `json:"gender"`
	BioDescription *string `json:"bio_description"`
}
