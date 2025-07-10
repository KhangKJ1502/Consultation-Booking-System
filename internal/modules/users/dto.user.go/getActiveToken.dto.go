package dtousergo

import (
	"time"

	"github.com/google/uuid"
)

type TokenInfo struct {
	TokenID   uuid.UUID `json:"token_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IsActive  bool      `json:"is_active"`
}

type ActiveTokensResponse struct {
	Tokens []TokenInfo `json:"tokens"`
	Total  int         `json:"total"`
}
