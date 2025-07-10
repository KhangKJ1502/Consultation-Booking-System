package dtousergo

type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}
