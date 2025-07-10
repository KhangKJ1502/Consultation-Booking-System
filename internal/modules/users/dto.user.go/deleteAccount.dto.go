package dtousergo

type DeleteAccountRequest struct {
	Password string `json:"password" binding:"required"`
}
