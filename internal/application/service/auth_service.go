package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/new-energy-monitoring/pkg/auth"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserDisabled       = errors.New("user is disabled")
	ErrUserLocked         = errors.New("user is locked due to too many failed attempts")
)

type AuthService struct {
	userRepo        repository.UserRepository
	roleRepo        repository.RoleRepository
	permissionRepo  repository.PermissionRepository
	logRepo         repository.OperationLogRepository
	jwtManager      *auth.JWTManager
	passwordManager *auth.PasswordManager
}

func NewAuthService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
	logRepo repository.OperationLogRepository,
	jwtManager *auth.JWTManager,
	passwordManager *auth.PasswordManager,
) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		roleRepo:        roleRepo,
		permissionRepo:  permissionRepo,
		logRepo:         logRepo,
		jwtManager:      jwtManager,
		passwordManager: passwordManager,
	}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
	User         *UserInfo `json:"user"`
}

type UserInfo struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	RealName    string   `json:"real_name"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

func (s *AuthService) Login(ctx context.Context, req *LoginRequest, ipAddress, userAgent string) (*LoginResponse, error) {
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		s.logLoginAttempt(ctx, "", req.Username, false, ipAddress, userAgent)
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive() {
		return nil, ErrUserDisabled
	}

	if !s.passwordManager.CheckPassword(req.Password, user.PasswordHash) {
		s.logLoginAttempt(ctx, user.ID, user.Username, false, ipAddress, userAgent)
		return nil, ErrInvalidCredentials
	}

	userWithPerms, permissions, err := s.userRepo.GetWithPermissions(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	roleCodes := make([]string, 0, len(userWithPerms.Roles))
	for _, role := range userWithPerms.Roles {
		roleCodes = append(roleCodes, role.Code)
	}

	permCodes := make([]string, 0, len(permissions))
	for _, perm := range permissions {
		permCodes = append(permCodes, perm.Code)
	}

	tokenPair, err := s.jwtManager.GenerateToken(user.ID, user.Username, roleCodes, permCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("failed to update last login: %w", err)
	}

	s.logLoginAttempt(ctx, user.ID, user.Username, true, ipAddress, userAgent)

	return &LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    tokenPair.TokenType,
		User: &UserInfo{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			RealName:    user.RealName,
			Roles:       roleCodes,
			Permissions: permCodes,
		},
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	claims, err := s.jwtManager.ParseToken(refreshToken)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.IsActive() {
		return nil, ErrUserDisabled
	}

	userWithPerms, permissions, err := s.userRepo.GetWithPermissions(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	roleCodes := make([]string, 0, len(userWithPerms.Roles))
	for _, role := range userWithPerms.Roles {
		roleCodes = append(roleCodes, role.Code)
	}

	permCodes := make([]string, 0, len(permissions))
	for _, perm := range permissions {
		permCodes = append(permCodes, perm.Code)
	}

	tokenPair, err := s.jwtManager.GenerateToken(user.ID, user.Username, roleCodes, permCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    tokenPair.TokenType,
		User: &UserInfo{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			RealName:    user.RealName,
			Roles:       roleCodes,
			Permissions: permCodes,
		},
	}, nil
}

func (s *AuthService) ValidateToken(token string) (*auth.Claims, error) {
	return s.jwtManager.ParseToken(token)
}

func (s *AuthService) Logout(ctx context.Context, userID string, ipAddress, userAgent string) error {
	log := entity.NewOperationLog(userID, "", entity.ActionLogout)
	log.SetRequestInfo(ipAddress, userAgent)
	return s.logRepo.Create(ctx, log)
}

func (s *AuthService) logLoginAttempt(ctx context.Context, userID, username string, success bool, ipAddress, userAgent string) {
	action := entity.ActionLogin
	details := entity.Details{
		"success": success,
	}

	log := entity.NewOperationLog(userID, username, action)
	log.SetDetails(details)
	log.SetRequestInfo(ipAddress, userAgent)

	_ = s.logRepo.Create(ctx, log)
}
