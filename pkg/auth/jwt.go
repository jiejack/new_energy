package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTConfig struct {
	Secret        string
	AccessExpire  int64
	RefreshExpire int64
}

type JWTManager struct {
	config *JWTConfig
}

type Claims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrExpiredToken      = errors.New("token has expired")
	ErrInvalidClaims     = errors.New("invalid claims")
)

func NewJWTManager(config *JWTConfig) *JWTManager {
	return &JWTManager{config: config}
}

func (m *JWTManager) GenerateToken(userID, username string, roles, permissions []string) (*TokenPair, error) {
	now := time.Now()
	accessExpireAt := now.Add(time.Duration(m.config.AccessExpire) * time.Second)
	refreshExpireAt := now.Add(time.Duration(m.config.RefreshExpire) * time.Second)

	accessClaims := &Claims{
		UserID:      userID,
		Username:    username,
		Roles:       roles,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpireAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "new-energy-monitoring",
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(m.config.Secret))
	if err != nil {
		return nil, err
	}

	refreshClaims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpireAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "new-energy-monitoring",
		},
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(m.config.Secret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    m.config.AccessExpire,
		TokenType:    "Bearer",
	}, nil
}

func (m *JWTManager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(m.config.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

func (m *JWTManager) RefreshAccessToken(refreshToken string) (*TokenPair, error) {
	claims, err := m.ParseToken(refreshToken)
	if err != nil {
		return nil, err
	}

	return m.GenerateToken(claims.UserID, claims.Username, claims.Roles, claims.Permissions)
}

func (m *JWTManager) ValidateToken(tokenString string) (bool, error) {
	_, err := m.ParseToken(tokenString)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (m *JWTManager) GetUserIDFromToken(tokenString string) (string, error) {
	claims, err := m.ParseToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

func (m *JWTManager) GetUsernameFromToken(tokenString string) (string, error) {
	claims, err := m.ParseToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.Username, nil
}
