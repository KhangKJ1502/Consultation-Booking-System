package utils

import (
	"cbs_backend/global"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

func GenerateJWT(userID uuid.UUID, exp int64) (string, error) {
	jwtSecret := global.ConfigConection.ServerCF
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     exp, // token hết hạn sau 72 giờ
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret.JWTSecret))
}
