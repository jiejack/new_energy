package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

var (
	ErrRoleNotFound       = errors.New("role not found")
	ErrRoleCodeExists     = errors.New("role code already exists")
	ErrPermissionNotFound = errors.New("permission not found")
	ErrPermissionCodeExists = errors.New("permission code already exists")
	ErrCannotDeleteSystemRole = errors.New("cannot delete system role")
)

type PermissionService struct {
	roleRepo       repository.RoleRepository
	permissionRepo repository.PermissionRepository
	logRepo        repository.OperationLogRepository
}

func NewPermissionService(
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
	logRepo repository.OperationLogRepository,
) *PermissionService {
	return &PermissionService{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		logRepo:        logRepo,
	}
}

type CreateRoleRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AssignPermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids" binding:"required"`
}

func (s *PermissionService) CreateRole(ctx context.Context, req *CreateRoleRequest, operatorID string) (*entity.Role, error) {
	exists, err := s.roleRepo.ExistsByCode(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to check role code: %w", err)
	}
	if exists {
		return nil, ErrRoleCodeExists
	}

	role := entity.NewRole(req.Code, req.Name)
	if req.Description != "" {
		role.SetDescription(req.Description)
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	s.logOperation(ctx, operatorID, entity.ActionCreateRole, entity.ResourceRole, role.ID, entity.Details{
		"code":        role.Code,
		"name":        role.Name,
		"description": role.Description,
	})

	return role, nil
}

func (s *PermissionService) UpdateRole(ctx context.Context, id string, req *UpdateRoleRequest, operatorID string) (*entity.Role, error) {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrRoleNotFound
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}

	if err := s.roleRepo.Update(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	s.logOperation(ctx, operatorID, entity.ActionUpdateRole, entity.ResourceRole, role.ID, entity.Details{
		"name":        role.Name,
		"description": role.Description,
	})

	return role, nil
}

func (s *PermissionService) DeleteRole(ctx context.Context, id string, operatorID string) error {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return ErrRoleNotFound
	}

	if role.IsSystemRole() {
		return ErrCannotDeleteSystemRole
	}

	if err := s.roleRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	s.logOperation(ctx, operatorID, entity.ActionDeleteRole, entity.ResourceRole, id, entity.Details{
		"code": role.Code,
		"name": role.Name,
	})

	return nil
}

func (s *PermissionService) GetRole(ctx context.Context, id string) (*entity.Role, error) {
	return s.roleRepo.GetByID(ctx, id)
}

func (s *PermissionService) GetRoleByCode(ctx context.Context, code string) (*entity.Role, error) {
	return s.roleRepo.GetByCode(ctx, code)
}

func (s *PermissionService) GetRoleWithPermissions(ctx context.Context, id string) (*entity.Role, error) {
	return s.roleRepo.GetWithPermissions(ctx, id)
}

func (s *PermissionService) ListRoles(ctx context.Context) ([]*entity.Role, error) {
	return s.roleRepo.List(ctx)
}

func (s *PermissionService) AssignPermissions(ctx context.Context, roleID string, permissionIDs []string, operatorID string) error {
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return ErrRoleNotFound
	}

	for _, permissionID := range permissionIDs {
		_, err := s.permissionRepo.GetByID(ctx, permissionID)
		if err != nil {
			return fmt.Errorf("permission not found: %s", permissionID)
		}

		if err := s.roleRepo.AssignPermission(ctx, roleID, permissionID); err != nil {
			return fmt.Errorf("failed to assign permission: %w", err)
		}
	}

	s.logOperation(ctx, operatorID, entity.ActionAssignPermission, entity.ResourceRole, roleID, entity.Details{
		"permission_ids": permissionIDs,
		"role_code":      role.Code,
		"role_name":      role.Name,
	})

	return nil
}

func (s *PermissionService) RemovePermission(ctx context.Context, roleID, permissionID string, operatorID string) error {
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return ErrRoleNotFound
	}

	if err := s.roleRepo.RemovePermission(ctx, roleID, permissionID); err != nil {
		return fmt.Errorf("failed to remove permission: %w", err)
	}

	s.logOperation(ctx, operatorID, entity.ActionRemovePermission, entity.ResourceRole, roleID, entity.Details{
		"permission_id": permissionID,
		"role_code":     role.Code,
	})

	return nil
}

type CreatePermissionRequest struct {
	Code         string `json:"code" binding:"required"`
	Name         string `json:"name" binding:"required"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	Action       string `json:"action"`
	Description  string `json:"description"`
}

func (s *PermissionService) CreatePermission(ctx context.Context, req *CreatePermissionRequest, operatorID string) (*entity.Permission, error) {
	exists, err := s.permissionRepo.ExistsByCode(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission code: %w", err)
	}
	if exists {
		return nil, ErrPermissionCodeExists
	}

	permission := entity.NewPermission(req.Code, req.Name)
	if req.ResourceType != "" {
		permission.SetResource(req.ResourceType, req.ResourceID)
	}
	if req.Action != "" {
		permission.SetAction(req.Action)
	}
	if req.Description != "" {
		permission.SetDescription(req.Description)
	}

	if err := s.permissionRepo.Create(ctx, permission); err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return permission, nil
}

func (s *PermissionService) BatchCreatePermissions(ctx context.Context, permissions []*entity.Permission) error {
	return s.permissionRepo.BatchCreate(ctx, permissions)
}

func (s *PermissionService) GetPermission(ctx context.Context, id string) (*entity.Permission, error) {
	return s.permissionRepo.GetByID(ctx, id)
}

func (s *PermissionService) GetPermissionByCode(ctx context.Context, code string) (*entity.Permission, error) {
	return s.permissionRepo.GetByCode(ctx, code)
}

func (s *PermissionService) ListPermissions(ctx context.Context, resourceType *string) ([]*entity.Permission, error) {
	return s.permissionRepo.List(ctx, resourceType)
}

func (s *PermissionService) GetUserPermissions(ctx context.Context, userID string) ([]*entity.Permission, error) {
	return s.permissionRepo.GetByUserID(ctx, userID)
}

func (s *PermissionService) GetRolePermissions(ctx context.Context, roleID string) ([]*entity.Permission, error) {
	return s.permissionRepo.GetByRoleID(ctx, roleID)
}

func (s *PermissionService) CheckPermission(ctx context.Context, userID, permissionCode string) (bool, error) {
	permissions, err := s.permissionRepo.GetByUserID(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, perm := range permissions {
		if perm.Code == permissionCode {
			return true, nil
		}
	}

	return false, nil
}

func (s *PermissionService) InitializeDefaultData(ctx context.Context) error {
	roles := entity.GetDefaultRoles()
	for _, role := range roles {
		exists, err := s.roleRepo.ExistsByCode(ctx, role.Code)
		if err != nil {
			return err
		}
		if !exists {
			if err := s.roleRepo.Create(ctx, role); err != nil {
				return fmt.Errorf("failed to create role %s: %w", role.Code, err)
			}
		}
	}

	permissions := entity.GetDefaultPermissions()
	for _, perm := range permissions {
		exists, err := s.permissionRepo.ExistsByCode(ctx, perm.Code)
		if err != nil {
			return err
		}
		if !exists {
			if err := s.permissionRepo.Create(ctx, perm); err != nil {
				return fmt.Errorf("failed to create permission %s: %w", perm.Code, err)
			}
		}
	}

	return nil
}

func (s *PermissionService) logOperation(ctx context.Context, operatorID, action, resourceType, resourceID string, details entity.Details) {
	log := entity.NewOperationLog(operatorID, "", action)
	log.SetResource(resourceType, resourceID)
	if details != nil {
		log.SetDetails(details)
	}
	_ = s.logRepo.Create(ctx, log)
}
