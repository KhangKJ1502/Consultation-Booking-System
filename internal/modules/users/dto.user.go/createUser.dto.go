package dtousergo

import "github.com/google/uuid"

type RegisterRequest struct {
	UserEmail string `json:"user_email" binding:"required,email"`
	Password  string `json:"password" binding:"required"`
	FullName  string `json:"full_name"`
}

type RegisterRespone struct {
	UserID    uuid.UUID
	UserEmail string `json:"user_email"`
	FullName  string `json:"full_name"`
	Token     string `json:"user_token"`
}
