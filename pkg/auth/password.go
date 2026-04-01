package auth

import (
	"errors"
	"regexp"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordTooShort       = errors.New("password is too short")
	ErrPasswordNoUppercase    = errors.New("password must contain uppercase letter")
	ErrPasswordNoLowercase    = errors.New("password must contain lowercase letter")
	ErrPasswordNoDigit        = errors.New("password must contain digit")
)

type PasswordConfig struct {
	MinLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireDigit     bool
}

type PasswordManager struct {
	config *PasswordConfig
}

func NewPasswordManager(config *PasswordConfig) *PasswordManager {
	return &PasswordManager{config: config}
}

func (m *PasswordManager) HashPassword(password string) (string, error) {
	if err := m.ValidatePassword(password); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (m *PasswordManager) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (m *PasswordManager) ValidatePassword(password string) error {
	if len(password) < m.config.MinLength {
		return ErrPasswordTooShort
	}

	if m.config.RequireUppercase {
		if !hasUppercase(password) {
			return ErrPasswordNoUppercase
		}
	}

	if m.config.RequireLowercase {
		if !hasLowercase(password) {
			return ErrPasswordNoLowercase
		}
	}

	if m.config.RequireDigit {
		if !hasDigit(password) {
			return ErrPasswordNoDigit
		}
	}

	return nil
}

func hasUppercase(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

func hasLowercase(s string) bool {
	for _, r := range s {
		if unicode.IsLower(r) {
			return true
		}
	}
	return false
}

func hasDigit(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func IsValidPhone(phone string) bool {
	if phone == "" {
		return false
	}
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return phoneRegex.MatchString(phone)
}
