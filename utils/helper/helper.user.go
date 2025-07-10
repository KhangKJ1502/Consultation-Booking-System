package helper

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

type HelperUser struct {
	db *gorm.DB
}

func NewHelperUser(db *gorm.DB) *HelperUser {
	return &HelperUser{db: db}
}

var (
	ErrWeakPassword = errors.New("password is too weak")
)

// GenerateSecureToken tạo token bảo mật cao cho reset password
func (uh *HelperUser) GenerateSecureToken(length int) string {
	// Sử dụng charset an toàn, tránh các ký tự gây nhầm lẫn
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic("failed to generate random number: " + err.Error())
		}
		result[i] = charset[num.Int64()]
	}

	return string(result)
}

func (uh *HelperUser) ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return ErrWeakPassword
	}

	return nil
}

func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}
	email = strings.TrimSpace(email)
	if len(email) > 254 {
		return false
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return false
	}
	if strings.Contains(email, "..") || // consecutive dots
		strings.HasPrefix(email, ".") || // starts with dot
		strings.HasSuffix(email, ".") || // ends with dot
		strings.Contains(email, "@.") || // @ followed by dot
		strings.Contains(email, ".@") { // dot followed by @
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	local := parts[0]
	domain := parts[1]
	if len(local) == 0 || len(local) > 64 {
		return false
	}
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}
	if !strings.Contains(domain, ".") {
		return false
	}
	domainParts := strings.Split(domain, ".")
	lastPart := domainParts[len(domainParts)-1]
	if len(lastPart) < 2 {
		return false
	}

	return true
}
func (uh *HelperUser) IsValidEmailStrict(email string) bool {
	if !IsValidEmail(email) {
		return false
	}
	email = strings.TrimSpace(email)
	parts := strings.Split(email, "@")
	local := parts[0]
	domain := parts[1]
	if strings.HasPrefix(local, ".") || strings.HasSuffix(local, ".") ||
		strings.HasPrefix(local, "-") || strings.HasSuffix(local, "-") {
		return false
	}
	if strings.HasPrefix(domain, "-") || strings.HasSuffix(domain, "-") {
		return false
	}
	localRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+$`)
	if !localRegex.MatchString(local) {
		return false
	}
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
	if !domainRegex.MatchString(domain) {
		return false
	}

	return true
}
