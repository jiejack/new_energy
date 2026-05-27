package service

import (
	"context"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRoleRepo struct {
	mock.Mock
}

func (m *mockRoleRepo) Create(ctx context.Context, role *entity.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}
func (m *mockRoleRepo) Update(ctx context.Context, role *entity.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}
func (m *mockRoleRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockRoleRepo) GetByID(ctx context.Context, id string) (*entity.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Role), args.Error(1)
}
func (m *mockRoleRepo) GetByCode(ctx context.Context, code string) (*entity.Role, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Role), args.Error(1)
}
func (m *mockRoleRepo) List(ctx context.Context) ([]*entity.Role, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Role), args.Error(1)
}
func (m *mockRoleRepo) GetWithPermissions(ctx context.Context, id string) (*entity.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Role), args.Error(1)
}
func (m *mockRoleRepo) AssignPermission(ctx context.Context, roleID, permissionID string) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}
func (m *mockRoleRepo) RemovePermission(ctx context.Context, roleID, permissionID string) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}
func (m *mockRoleRepo) ExistsByCode(ctx context.Context, code string) (bool, error) {
	args := m.Called(ctx, code)
	return args.Bool(0), args.Error(1)
}

type mockPermRepo struct {
	mock.Mock
}

func (m *mockPermRepo) Create(ctx context.Context, permission *entity.Permission) error {
	args := m.Called(ctx, permission)
	return args.Error(0)
}
func (m *mockPermRepo) BatchCreate(ctx context.Context, permissions []*entity.Permission) error {
	args := m.Called(ctx, permissions)
	return args.Error(0)
}
func (m *mockPermRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockPermRepo) GetByID(ctx context.Context, id string) (*entity.Permission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Permission), args.Error(1)
}
func (m *mockPermRepo) GetByCode(ctx context.Context, code string) (*entity.Permission, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Permission), args.Error(1)
}
func (m *mockPermRepo) List(ctx context.Context, resourceType *string) ([]*entity.Permission, error) {
	args := m.Called(ctx, resourceType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Permission), args.Error(1)
}
func (m *mockPermRepo) GetByRoleID(ctx context.Context, roleID string) ([]*entity.Permission, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Permission), args.Error(1)
}
func (m *mockPermRepo) GetByUserID(ctx context.Context, userID string) ([]*entity.Permission, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Permission), args.Error(1)
}
func (m *mockPermRepo) ExistsByCode(ctx context.Context, code string) (bool, error) {
	args := m.Called(ctx, code)
	return args.Bool(0), args.Error(1)
}

type mockLogRepoForPerm struct {
	mock.Mock
}

func (m *mockLogRepoForPerm) Create(ctx context.Context, log *entity.OperationLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}
func (m *mockLogRepoForPerm) GetByID(ctx context.Context, id string) (*entity.OperationLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.OperationLog), args.Error(1)
}
func (m *mockLogRepoForPerm) List(ctx context.Context, query *repository.OperationLogQuery) ([]*entity.OperationLog, int64, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]*entity.OperationLog), args.Get(1).(int64), args.Error(2)
}
func (m *mockLogRepoForPerm) DeleteBefore(ctx context.Context, before int64) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

func setupPermissionService() (*PermissionService, *mockRoleRepo, *mockPermRepo, *mockLogRepoForPerm) {
	roleRepo := new(mockRoleRepo)
	permRepo := new(mockPermRepo)
	logRepo := new(mockLogRepoForPerm)
	svc := NewPermissionService(roleRepo, permRepo, logRepo)
	return svc, roleRepo, permRepo, logRepo
}

func TestPermissionService_CreateRole_Success(t *testing.T) {
	svc, roleRepo, _, logRepo := setupPermissionService()
	roleRepo.On("ExistsByCode", mock.Anything, "admin").Return(false, nil)
	roleRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Role")).Return(nil)
	logRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).Return(nil)
	role, err := svc.CreateRole(context.Background(), &CreateRoleRequest{Code: "admin", Name: "Admin"}, "op-001")
	assert.NoError(t, err)
	assert.NotNil(t, role)
	assert.Equal(t, "admin", role.Code)
}

func TestPermissionService_CreateRole_CodeExists(t *testing.T) {
	svc, roleRepo, _, _ := setupPermissionService()
	roleRepo.On("ExistsByCode", mock.Anything, "admin").Return(true, nil)
	role, err := svc.CreateRole(context.Background(), &CreateRoleRequest{Code: "admin", Name: "Admin"}, "op-001")
	assert.ErrorIs(t, err, ErrRoleCodeExists)
	assert.Nil(t, role)
}

func TestPermissionService_UpdateRole_Success(t *testing.T) {
	svc, roleRepo, _, logRepo := setupPermissionService()
	role := entity.NewRole("admin", "Admin")
	roleRepo.On("GetByID", mock.Anything, "role-001").Return(role, nil)
	roleRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.Role")).Return(nil)
	logRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).Return(nil)
	updated, err := svc.UpdateRole(context.Background(), "role-001", &UpdateRoleRequest{Name: "Super Admin"}, "op-001")
	assert.NoError(t, err)
	assert.Equal(t, "Super Admin", updated.Name)
}

func TestPermissionService_UpdateRole_NotFound(t *testing.T) {
	svc, roleRepo, _, _ := setupPermissionService()
	roleRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)
	_, err := svc.UpdateRole(context.Background(), "nonexistent", &UpdateRoleRequest{Name: "Test"}, "op-001")
	assert.ErrorIs(t, err, ErrRoleNotFound)
}

func TestPermissionService_DeleteRole_Success(t *testing.T) {
	svc, roleRepo, _, logRepo := setupPermissionService()
	role := entity.NewRole("custom", "Custom")
	roleRepo.On("GetByID", mock.Anything, "role-001").Return(role, nil)
	roleRepo.On("Delete", mock.Anything, "role-001").Return(nil)
	logRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).Return(nil)
	err := svc.DeleteRole(context.Background(), "role-001", "op-001")
	assert.NoError(t, err)
}

func TestPermissionService_DeleteRole_SystemRole(t *testing.T) {
	svc, roleRepo, _, _ := setupPermissionService()
	role := entity.NewRole("system", "System")
	role.SetAsSystemRole()
	roleRepo.On("GetByID", mock.Anything, "role-001").Return(role, nil)
	err := svc.DeleteRole(context.Background(), "role-001", "op-001")
	assert.ErrorIs(t, err, ErrCannotDeleteSystemRole)
}

func TestPermissionService_GetRole(t *testing.T) {
	svc, roleRepo, _, _ := setupPermissionService()
	role := entity.NewRole("admin", "Admin")
	roleRepo.On("GetByID", mock.Anything, "role-001").Return(role, nil)
	result, err := svc.GetRole(context.Background(), "role-001")
	assert.NoError(t, err)
	assert.Equal(t, "admin", result.Code)
}

func TestPermissionService_GetRoleByCode(t *testing.T) {
	svc, roleRepo, _, _ := setupPermissionService()
	role := entity.NewRole("admin", "Admin")
	roleRepo.On("GetByCode", mock.Anything, "admin").Return(role, nil)
	result, err := svc.GetRoleByCode(context.Background(), "admin")
	assert.NoError(t, err)
	assert.Equal(t, "admin", result.Code)
}

func TestPermissionService_GetRoleWithPermissions(t *testing.T) {
	svc, roleRepo, _, _ := setupPermissionService()
	role := entity.NewRole("admin", "Admin")
	roleRepo.On("GetWithPermissions", mock.Anything, "role-001").Return(role, nil)
	result, err := svc.GetRoleWithPermissions(context.Background(), "role-001")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestPermissionService_ListRoles(t *testing.T) {
	svc, roleRepo, _, _ := setupPermissionService()
	roles := []*entity.Role{entity.NewRole("admin", "Admin")}
	roleRepo.On("List", mock.Anything).Return(roles, nil)
	result, err := svc.ListRoles(context.Background())
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestPermissionService_AssignPermissions_Success(t *testing.T) {
	svc, roleRepo, permRepo, logRepo := setupPermissionService()
	role := entity.NewRole("admin", "Admin")
	roleRepo.On("GetByID", mock.Anything, "role-001").Return(role, nil)
	permRepo.On("GetByID", mock.Anything, "perm-001").Return(&entity.Permission{}, nil)
	roleRepo.On("AssignPermission", mock.Anything, "role-001", "perm-001").Return(nil)
	logRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).Return(nil)
	err := svc.AssignPermissions(context.Background(), "role-001", []string{"perm-001"}, "op-001")
	assert.NoError(t, err)
}

func TestPermissionService_AssignPermissions_RoleNotFound(t *testing.T) {
	svc, roleRepo, _, _ := setupPermissionService()
	roleRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)
	err := svc.AssignPermissions(context.Background(), "nonexistent", []string{"perm-001"}, "op-001")
	assert.ErrorIs(t, err, ErrRoleNotFound)
}

func TestPermissionService_RemovePermission_Success(t *testing.T) {
	svc, roleRepo, _, logRepo := setupPermissionService()
	role := entity.NewRole("admin", "Admin")
	roleRepo.On("GetByID", mock.Anything, "role-001").Return(role, nil)
	roleRepo.On("RemovePermission", mock.Anything, "role-001", "perm-001").Return(nil)
	logRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).Return(nil)
	err := svc.RemovePermission(context.Background(), "role-001", "perm-001", "op-001")
	assert.NoError(t, err)
}

func TestPermissionService_CreatePermission_Success(t *testing.T) {
	svc, _, permRepo, _ := setupPermissionService()
	permRepo.On("ExistsByCode", mock.Anything, "alarm:read").Return(false, nil)
	permRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Permission")).Return(nil)
	perm, err := svc.CreatePermission(context.Background(), &CreatePermissionRequest{Code: "alarm:read", Name: "Read Alarms"}, "op-001")
	assert.NoError(t, err)
	assert.NotNil(t, perm)
}

func TestPermissionService_CreatePermission_CodeExists(t *testing.T) {
	svc, _, permRepo, _ := setupPermissionService()
	permRepo.On("ExistsByCode", mock.Anything, "alarm:read").Return(true, nil)
	perm, err := svc.CreatePermission(context.Background(), &CreatePermissionRequest{Code: "alarm:read", Name: "Read Alarms"}, "op-001")
	assert.ErrorIs(t, err, ErrPermissionCodeExists)
	assert.Nil(t, perm)
}

func TestPermissionService_GetPermission(t *testing.T) {
	svc, _, permRepo, _ := setupPermissionService()
	perm := entity.NewPermission("alarm:read", "Read Alarms")
	permRepo.On("GetByID", mock.Anything, "perm-001").Return(perm, nil)
	result, err := svc.GetPermission(context.Background(), "perm-001")
	assert.NoError(t, err)
	assert.Equal(t, "alarm:read", result.Code)
}

func TestPermissionService_GetPermissionByCode(t *testing.T) {
	svc, _, permRepo, _ := setupPermissionService()
	perm := entity.NewPermission("alarm:read", "Read Alarms")
	permRepo.On("GetByCode", mock.Anything, "alarm:read").Return(perm, nil)
	result, err := svc.GetPermissionByCode(context.Background(), "alarm:read")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestPermissionService_ListPermissions(t *testing.T) {
	svc, _, permRepo, _ := setupPermissionService()
	perms := []*entity.Permission{entity.NewPermission("alarm:read", "Read Alarms")}
	permRepo.On("List", mock.Anything, (*string)(nil)).Return(perms, nil)
	result, err := svc.ListPermissions(context.Background(), nil)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestPermissionService_GetUserPermissions(t *testing.T) {
	svc, _, permRepo, _ := setupPermissionService()
	perms := []*entity.Permission{entity.NewPermission("alarm:read", "Read Alarms")}
	permRepo.On("GetByUserID", mock.Anything, "user-001").Return(perms, nil)
	result, err := svc.GetUserPermissions(context.Background(), "user-001")
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestPermissionService_GetRolePermissions(t *testing.T) {
	svc, _, permRepo, _ := setupPermissionService()
	perms := []*entity.Permission{entity.NewPermission("alarm:read", "Read Alarms")}
	permRepo.On("GetByRoleID", mock.Anything, "role-001").Return(perms, nil)
	result, err := svc.GetRolePermissions(context.Background(), "role-001")
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestPermissionService_CheckPermission_True(t *testing.T) {
	svc, _, permRepo, _ := setupPermissionService()
	perm := entity.NewPermission("alarm:read", "Read Alarms")
	permRepo.On("GetByUserID", mock.Anything, "user-001").Return([]*entity.Permission{perm}, nil)
	has, err := svc.CheckPermission(context.Background(), "user-001", "alarm:read")
	assert.NoError(t, err)
	assert.True(t, has)
}

func TestPermissionService_CheckPermission_False(t *testing.T) {
	svc, _, permRepo, _ := setupPermissionService()
	perm := entity.NewPermission("alarm:read", "Read Alarms")
	permRepo.On("GetByUserID", mock.Anything, "user-001").Return([]*entity.Permission{perm}, nil)
	has, err := svc.CheckPermission(context.Background(), "user-001", "alarm:write")
	assert.NoError(t, err)
	assert.False(t, has)
}

func TestPermissionService_BatchCreatePermissions(t *testing.T) {
	svc, _, permRepo, _ := setupPermissionService()
	perms := []*entity.Permission{entity.NewPermission("p1", "P1")}
	permRepo.On("BatchCreate", mock.Anything, perms).Return(nil)
	err := svc.BatchCreatePermissions(context.Background(), perms)
	assert.NoError(t, err)
}

func TestPermissionService_InitializeDefaultData(t *testing.T) {
	svc, roleRepo, permRepo, _ := setupPermissionService()
	roleRepo.On("ExistsByCode", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	roleRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Role")).Return(nil)
	permRepo.On("ExistsByCode", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	permRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Permission")).Return(nil)
	err := svc.InitializeDefaultData(context.Background())
	assert.NoError(t, err)
}
