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
	ErrUserNotFound       = errors.New("user not found")
	ErrUsernameExists     = errors.New("username already exists")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidOldPassword = errors.New("invalid old password")
)

type UserService struct {
	userRepo        repository.UserRepository
	roleRepo        repository.RoleRepository
	logRepo         repository.OperationLogRepository
	passwordManager *auth.PasswordManager
}

func NewUserService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	logRepo repository.OperationLogRepository,
	passwordManager *auth.PasswordManager,
) *UserService {
	return &UserService{
		userRepo:        userRepo,
		roleRepo:        roleRepo,
		logRepo:         logRepo,
		passwordManager: passwordManager,
	}
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	RealName string `json:"real_name"`
}

type UpdateUserRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	RealName string `json:"real_name"`
	Avatar   string `json:"avatar"`
	Status   *int   `json:"status"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type AssignRolesRequest struct {
	RoleIDs []string `json:"role_ids" binding:"required"`
}

type UserListResponse struct {
	Users []*entity.User `json:"users"`
	Total int64          `json:"total"`
	Page  int            `json:"page"`
	PageSize int         `json:"page_size"`
}

func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest, operatorID string) (*entity.User, error) {
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if exists {
		return nil, ErrUsernameExists
	}

	if req.Email != "" {
		exists, err = s.userRepo.ExistsByEmail(ctx, req.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to check email: %w", err)
		}
		if exists {
			return nil, ErrEmailExists
		}
	}

	passwordHash, err := s.passwordManager.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := entity.NewUser(req.Username, passwordHash)
	if req.Email != "" {
		user.SetEmail(req.Email)
	}
	if req.Phone != "" {
		user.SetPhone(req.Phone)
	}
	if req.RealName != "" {
		user.SetRealName(req.RealName)
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logOperation(ctx, operatorID, entity.ActionCreateUser, entity.ResourceUser, user.ID, entity.Details{
		"username": user.Username,
		"email":    user.Email,
	})

	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id string, req *UpdateUserRequest, operatorID string) (*entity.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if req.Email != "" && req.Email != user.Email {
		exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to check email: %w", err)
		}
		if exists {
			return nil, ErrEmailExists
		}
		user.SetEmail(req.Email)
	}

	if req.Phone != "" {
		user.SetPhone(req.Phone)
	}
	if req.RealName != "" {
		user.SetRealName(req.RealName)
	}
	if req.Avatar != "" {
		user.SetAvatar(req.Avatar)
	}
	if req.Status != nil {
		if *req.Status == 1 {
			user.Activate()
		} else {
			user.Deactivate()
		}
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	s.logOperation(ctx, operatorID, entity.ActionUpdateUser, entity.ResourceUser, user.ID, entity.Details{
		"email":    user.Email,
		"real_name": user.RealName,
	})

	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string, operatorID string) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	s.logOperation(ctx, operatorID, entity.ActionDeleteUser, entity.ResourceUser, id, entity.Details{
		"username": user.Username,
	})

	return nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (*entity.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *UserService) GetUserWithRoles(ctx context.Context, id string) (*entity.User, error) {
	return s.userRepo.GetWithRoles(ctx, id)
}

func (s *UserService) ListUsers(ctx context.Context, status *entity.UserStatus, page, pageSize int) (*UserListResponse, error) {
	users, total, err := s.userRepo.List(ctx, status, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return &UserListResponse{
		Users:    users,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *UserService) ChangePassword(ctx context.Context, id string, req *ChangePasswordRequest, operatorID string) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	if !s.passwordManager.CheckPassword(req.OldPassword, user.PasswordHash) {
		return ErrInvalidOldPassword
	}

	passwordHash, err := s.passwordManager.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.UpdatePassword(passwordHash)

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	s.logOperation(ctx, operatorID, entity.ActionChangePassword, entity.ResourceUser, id, nil)

	return nil
}

func (s *UserService) AssignRoles(ctx context.Context, userID string, roleIDs []string, operatorID string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	for _, roleID := range roleIDs {
		_, err := s.roleRepo.GetByID(ctx, roleID)
		if err != nil {
			return fmt.Errorf("role not found: %s", roleID)
		}

		if err := s.userRepo.AssignRole(ctx, userID, roleID); err != nil {
			return fmt.Errorf("failed to assign role: %w", err)
		}
	}

	s.logOperation(ctx, operatorID, entity.ActionAssignRole, entity.ResourceUser, userID, entity.Details{
		"role_ids": roleIDs,
		"username": user.Username,
	})

	return nil
}

func (s *UserService) RemoveRole(ctx context.Context, userID, roleID string, operatorID string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	if err := s.userRepo.RemoveRole(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	s.logOperation(ctx, operatorID, entity.ActionRemoveRole, entity.ResourceUser, userID, entity.Details{
		"role_id":  roleID,
		"username": user.Username,
	})

	return nil
}

func (s *UserService) logOperation(ctx context.Context, operatorID, action, resourceType, resourceID string, details entity.Details) {
	log := entity.NewOperationLog(operatorID, "", action)
	log.SetResource(resourceType, resourceID)
	if details != nil {
		log.SetDetails(details)
	}
	_ = s.logRepo.Create(ctx, log)
}
