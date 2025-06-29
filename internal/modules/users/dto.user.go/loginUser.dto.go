package dtousergo

import "github.com/google/uuid"

type LoginRequest struct {
	Email    string `json:"user_email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	UserID      uuid.UUID `json:"user_id"`
	FullName    string    `json:"full_name"`
	Token       string    `json:"acess_token"`
	RefeshToken string    `json:"refesh_token"`
}
