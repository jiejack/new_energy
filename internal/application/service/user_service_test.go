package service

import (
	"context"
	"errors"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepositoryForUserService 用户服务用户仓储Mock
type MockUserRepositoryForUserService struct {
	mock.Mock
	users map[string]*entity.User
}

func NewMockUserRepositoryForUserService() *MockUserRepositoryForUserService {
	return &MockUserRepositoryForUserService{
		users: make(map[string]*entity.User),
	}
}

func (m *MockUserRepositoryForUserService) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	if args.Error(0) == nil {
		m.users[user.ID] = user
	}
	return args.Error(0)
}

func (m *MockUserRepositoryForUserService) Update(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	if args.Error(0) == nil {
		m.users[user.ID] = user
	}
	return args.Error(0)
}

func (m *MockUserRepositoryForUserService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if args.Error(0) == nil {
		delete(m.users, id)
	}
	return args.Error(0)
}

func (m *MockUserRepositoryForUserService) GetByID(ctx context.Context, id string) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepositoryForUserService) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepositoryForUserService) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepositoryForUserService) List(ctx context.Context, status *entity.UserStatus, page, pageSize int) ([]*entity.User, int64, error) {
	args := m.Called(ctx, status, page, pageSize)
	return args.Get(0).([]*entity.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepositoryForUserService) GetWithRoles(ctx context.Context, id string) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepositoryForUserService) GetWithPermissions(ctx context.Context, id string) (*entity.User, []*entity.Permission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(*entity.User), args.Get(1).([]*entity.Permission), args.Error(2)
}

func (m *MockUserRepositoryForUserService) AssignRole(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockUserRepositoryForUserService) RemoveRole(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockUserRepositoryForUserService) UpdateLastLogin(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepositoryForUserService) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockUserRepositoryForUserService) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(bool), args.Error(1)
}

// MockRoleRepositoryForUserService 角色仓储Mock
type MockRoleRepositoryForUserService struct {
	mock.Mock
}

func (m *MockRoleRepositoryForUserService) Create(ctx context.Context, role *entity.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepositoryForUserService) Update(ctx context.Context, role *entity.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepositoryForUserService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRoleRepositoryForUserService) GetByID(ctx context.Context, id string) (*entity.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Role), args.Error(1)
}

func (m *MockRoleRepositoryForUserService) GetByCode(ctx context.Context, code string) (*entity.Role, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Role), args.Error(1)
}

func (m *MockRoleRepositoryForUserService) List(ctx context.Context) ([]*entity.Role, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entity.Role), args.Error(1)
}

func (m *MockRoleRepositoryForUserService) GetWithPermissions(ctx context.Context, id string) (*entity.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Role), args.Error(1)
}

func (m *MockRoleRepositoryForUserService) AssignPermission(ctx context.Context, roleID, permissionID string) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}

func (m *MockRoleRepositoryForUserService) RemovePermission(ctx context.Context, roleID, permissionID string) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}

func (m *MockRoleRepositoryForUserService) ExistsByCode(ctx context.Context, code string) (bool, error) {
	args := m.Called(ctx, code)
	return args.Get(0).(bool), args.Error(1)
}

// MockOperationLogRepositoryForUserService 操作日志仓储Mock
type MockOperationLogRepositoryForUserService struct {
	mock.Mock
}

func (m *MockOperationLogRepositoryForUserService) Create(ctx context.Context, log *entity.OperationLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockOperationLogRepositoryForUserService) GetByID(ctx context.Context, id string) (*entity.OperationLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.OperationLog), args.Error(1)
}

func (m *MockOperationLogRepositoryForUserService) List(ctx context.Context, query *repository.OperationLogQuery) ([]*entity.OperationLog, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.OperationLog), args.Get(1).(int64), args.Error(2)
}

func (m *MockOperationLogRepositoryForUserService) DeleteBefore(ctx context.Context, before int64) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

func TestUserService_CreateUser_Success(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepositoryForUserService()
	mockRoleRepo := new(MockRoleRepositoryForUserService)
	mockLogRepo := new(MockOperationLogRepositoryForUserService)
	passwordManager := newTestPasswordManager()

	service := NewUserService(mockUserRepo, mockRoleRepo, mockLogRepo, passwordManager)

	// 设置Mock期望
	mockUserRepo.On("ExistsByUsername", ctx, "newuser").Return(false, nil)
	mockUserRepo.On("ExistsByEmail", ctx, "newuser@test.com").Return(false, nil)
	mockUserRepo.On("Create", ctx, mock.AnythingOfType("*entity.User")).Return(nil)
	mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	// 执行测试
	req := &CreateUserRequest{
		Username: "newuser",
		Password: "password123",
		Email:    "newuser@test.com",
		Phone:    "13800138000",
		RealName: "新用户",
	}
	user, err := service.CreateUser(ctx, req, "admin-001")

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "newuser", user.Username)
	assert.Equal(t, "newuser@test.com", user.Email)
	assert.Equal(t, "13800138000", user.Phone)
	assert.Equal(t, "新用户", user.RealName)
	assert.Equal(t, entity.UserStatusActive, user.Status)

	mockUserRepo.AssertExpectations(t)
	mockLogRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_UsernameExists(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepositoryForUserService()
	mockRoleRepo := new(MockRoleRepositoryForUserService)
	mockLogRepo := new(MockOperationLogRepositoryForUserService)
	passwordManager := newTestPasswordManager()

	service := NewUserService(mockUserRepo, mockRoleRepo, mockLogRepo, passwordManager)

	// 设置Mock期望 - 用户名已存在
	mockUserRepo.On("ExistsByUsername", ctx, "existinguser").Return(true, nil)

	// 执行测试
	req := &CreateUserRequest{
		Username: "existinguser",
		Password: "password123",
	}
	user, err := service.CreateUser(ctx, req, "admin-001")

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, ErrUsernameExists, err)
	assert.Nil(t, user)

	mockUserRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_EmailExists(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepositoryForUserService()
	mockRoleRepo := new(MockRoleRepositoryForUserService)
	mockLogRepo := new(MockOperationLogRepositoryForUserService)
	passwordManager := newTestPasswordManager()

	service := NewUserService(mockUserRepo, mockRoleRepo, mockLogRepo, passwordManager)

	// 设置Mock期望
	mockUserRepo.On("ExistsByUsername", ctx, "newuser").Return(false, nil)
	mockUserRepo.On("ExistsByEmail", ctx, "existing@test.com").Return(true, nil)

	// 执行测试
	req := &CreateUserRequest{
		Username: "newuser",
		Password: "password123",
		Email:    "existing@test.com",
	}
	user, err := service.CreateUser(ctx, req, "admin-001")

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, ErrEmailExists, err)
	assert.Nil(t, user)

	mockUserRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_Success(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepositoryForUserService()
	mockRoleRepo := new(MockRoleRepositoryForUserService)
	mockLogRepo := new(MockOperationLogRepositoryForUserService)
	passwordManager := newTestPasswordManager()

	service := NewUserService(mockUserRepo, mockRoleRepo, mockLogRepo, passwordManager)

	// 准备测试数据
	passwordHash, _ := passwordManager.HashPassword("password123")
	existingUser := entity.NewUser("testuser", passwordHash)
	existingUser.ID = "user-001"
	existingUser.Email = "old@test.com"

	// 设置Mock期望
	mockUserRepo.On("GetByID", ctx, "user-001").Return(existingUser, nil)
	mockUserRepo.On("ExistsByEmail", ctx, "new@test.com").Return(false, nil)
	mockUserRepo.On("Update", ctx, mock.AnythingOfType("*entity.User")).Return(nil)
	mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	// 执行测试
	activeStatus := 1
	req := &UpdateUserRequest{
		Email:    "new@test.com",
		Phone:    "13900139000",
		RealName: "更新后的名字",
		Status:   &activeStatus,
	}
	user, err := service.UpdateUser(ctx, "user-001", req, "admin-001")

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "new@test.com", user.Email)
	assert.Equal(t, "13900139000", user.Phone)
	assert.Equal(t, "更新后的名字", user.RealName)

	mockUserRepo.AssertExpectations(t)
	mockLogRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_NotFound(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepositoryForUserService()
	mockRoleRepo := new(MockRoleRepositoryForUserService)
	mockLogRepo := new(MockOperationLogRepositoryForUserService)
	passwordManager := newTestPasswordManager()

	service := NewUserService(mockUserRepo, mockRoleRepo, mockLogRepo, passwordManager)

	// 设置Mock期望 - 用户不存在
	mockUserRepo.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))

	// 执行测试
	req := &UpdateUserRequest{
		Email: "new@test.com",
	}
	user, err := service.UpdateUser(ctx, "nonexistent", req, "admin-001")

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
	assert.Nil(t, user)

	mockUserRepo.AssertExpectations(t)
}

func TestUserService_DeleteUser_Success(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepositoryForUserService()
	mockRoleRepo := new(MockRoleRepositoryForUserService)
	mockLogRepo := new(MockOperationLogRepositoryForUserService)
	passwordManager := newTestPasswordManager()

	service := NewUserService(mockUserRepo, mockRoleRepo, mockLogRepo, passwordManager)

	// 准备测试数据
	passwordHash, _ := passwordManager.HashPassword("password123")
	existingUser := entity.NewUser("testuser", passwordHash)
	existingUser.ID = "user-001"

	// 设置Mock期望
	mockUserRepo.On("GetByID", ctx, "user-001").Return(existingUser, nil)
	mockUserRepo.On("Delete", ctx, "user-001").Return(nil)
	mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	// 执行测试
	err := service.DeleteUser(ctx, "user-001", "admin-001")

	// 验证结果
	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
	mockLogRepo.AssertExpectations(t)
}

func TestUserService_DeleteUser_NotFound(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepositoryForUserService()
	mockRoleRepo := new(MockRoleRepositoryForUserService)
	mockLogRepo := new(MockOperationLogRepositoryForUserService)
	passwordManager := newTestPasswordManager()

	service := NewUserService(mockUserRepo, mockRoleRepo, mockLogRepo, passwordManager)

	// 设置Mock期望 - 用户不存在
	mockUserRepo.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))

	// 执行测试
	err := service.DeleteUser(ctx, "nonexistent", "admin-001")

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)

	mockUserRepo.AssertExpectations(t)
}

func TestUserService_GetUser_Success(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepositoryForUserService()
	mockRoleRepo := new(MockRoleRepositoryForUserService)
	mockLogRepo := new(MockOperationLogRepositoryForUserService)
	passwordManager := newTestPasswordManager()

	service := NewUserService(mockUserRepo, mockRoleRepo, mockLogRepo, passwordManager)

	// 准备测试数据
	passwordHash, _ := passwordManager.HashPassword("password123")
	existingUser := entity.NewUser("testuser", passwordHash)
	existingUser.ID = "user-001"

	// 设置Mock期望
	mockUserRepo.On("GetByID", ctx, "user-001").Return(existingUser, nil)

	// 执行测试
	user, err := service.GetUser(ctx, "user-001")

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "user-001", user.ID)
	assert.Equal(t, "testuser", user.Username)

	mockUserRepo.AssertExpectations(t)
}

func TestUserService_ListUsers(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepositoryForUserService()
	mockRoleRepo := new(MockRoleRepositoryForUserService)
	mockLogRepo := new(MockOperationLogRepositoryForUserService)
	passwordManager := newTestPasswordManager()

	service := NewUserService(mockUserRepo, mockRoleRepo, mockLogRepo, passwordManager)

	// 准备测试数据
	passwordHash, _ := passwordManager.HashPassword("password123")
	users := []*entity.User{
		entity.NewUser("user1", passwordHash),
		entity.NewUser("user2", passwordHash),
		entity.NewUser("user3", passwordHash),
	}

	// 设置Mock期望
	mockUserRepo.On("List", ctx, mock.Anything, 1, 10).Return(users, int64(3), nil)

	// 执行测试
	status := entity.UserStatusActive
	resp, err := service.ListUsers(ctx, &status, 1, 10)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Users, 3)
	assert.Equal(t, int64(3), resp.Total)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 10, resp.PageSize)

	mockUserRepo.AssertExpectations(t)
}

func TestUserService_ChangePassword_Success(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepositoryForUserService()
	mockRoleRepo := new(MockRoleRepositoryForUserService)
	mockLogRepo := new(MockOperationLogRepositoryForUserService)
	passwordManager := newTestPasswordManager()

	service := NewUserService(mockUserRepo, mockRoleRepo, mockLogRepo, passwordManager)

	// 准备测试数据
	passwordHash, _ := passwordManager.HashPassword("oldpassword")
	existingUser := entity.NewUser("testuser", passwordHash)
	existingUser.ID = "user-001"

	// 设置Mock期望
	mockUserRepo.On("GetByID", ctx, "user-001").Return(existingUser, nil)
	mockUserRepo.On("Update", ctx, mock.AnythingOfType("*entity.User")).Return(nil)
	mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	// 执行测试
	req := &ChangePasswordRequest{
		OldPassword: "oldpassword",
		NewPassword: "newpassword123",
	}
	err := service.ChangePassword(ctx, "user-001", req, "admin-001")

	// 验证结果
	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
	mockLogRepo.AssertExpectations(t)
}

func TestUserService_ChangePassword_WrongOldPassword(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepositoryForUserService()
	mockRoleRepo := new(MockRoleRepositoryForUserService)
	mockLogRepo := new(MockOperationLogRepositoryForUserService)
	passwordManager := newTestPasswordManager()

	service := NewUserService(mockUserRepo, mockRoleRepo, mockLogRepo, passwordManager)

	// 准备测试数据
	passwordHash, _ := passwordManager.HashPassword("correctpassword")
	existingUser := entity.NewUser("testuser", passwordHash)
	existingUser.ID = "user-001"

	// 设置Mock期望
	mockUserRepo.On("GetByID", ctx, "user-001").Return(existingUser, nil)

	// 执行测试 - 使用错误的旧密码
	req := &ChangePasswordRequest{
		OldPassword: "wrongpassword",
		NewPassword: "newpassword123",
	}
	err := service.ChangePassword(ctx, "user-001", req, "admin-001")

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidOldPassword, err)

	mockUserRepo.AssertExpectations(t)
}

func TestUserService_AssignRoles_Success(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepositoryForUserService()
	mockRoleRepo := new(MockRoleRepositoryForUserService)
	mockLogRepo := new(MockOperationLogRepositoryForUserService)
	passwordManager := newTestPasswordManager()

	service := NewUserService(mockUserRepo, mockRoleRepo, mockLogRepo, passwordManager)

	// 准备测试数据
	passwordHash, _ := passwordManager.HashPassword("password123")
	existingUser := entity.NewUser("testuser", passwordHash)
	existingUser.ID = "user-001"

	adminRole := entity.NewRole("admin", "管理员")
	adminRole.ID = "role-001"

	// 设置Mock期望
	mockUserRepo.On("GetByID", ctx, "user-001").Return(existingUser, nil)
	mockRoleRepo.On("GetByID", ctx, "role-001").Return(adminRole, nil)
	mockUserRepo.On("AssignRole", ctx, "user-001", "role-001").Return(nil)
	mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	// 执行测试
	req := &AssignRolesRequest{
		RoleIDs: []string{"role-001"},
	}
	err := service.AssignRoles(ctx, "user-001", req.RoleIDs, "admin-001")

	// 验证结果
	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
	mockLogRepo.AssertExpectations(t)
}

func TestUserService_AssignRoles_UserNotFound(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepositoryForUserService()
	mockRoleRepo := new(MockRoleRepositoryForUserService)
	mockLogRepo := new(MockOperationLogRepositoryForUserService)
	passwordManager := newTestPasswordManager()

	service := NewUserService(mockUserRepo, mockRoleRepo, mockLogRepo, passwordManager)

	// 设置Mock期望 - 用户不存在
	mockUserRepo.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))

	// 执行测试
	req := &AssignRolesRequest{
		RoleIDs: []string{"role-001"},
	}
	err := service.AssignRoles(ctx, "nonexistent", req.RoleIDs, "admin-001")

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)

	mockUserRepo.AssertExpectations(t)
}

func TestUserService_RemoveRole_Success(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepositoryForUserService()
	mockRoleRepo := new(MockRoleRepositoryForUserService)
	mockLogRepo := new(MockOperationLogRepositoryForUserService)
	passwordManager := newTestPasswordManager()

	service := NewUserService(mockUserRepo, mockRoleRepo, mockLogRepo, passwordManager)

	// 准备测试数据
	passwordHash, _ := passwordManager.HashPassword("password123")
	existingUser := entity.NewUser("testuser", passwordHash)
	existingUser.ID = "user-001"

	// 设置Mock期望
	mockUserRepo.On("GetByID", ctx, "user-001").Return(existingUser, nil)
	mockUserRepo.On("RemoveRole", ctx, "user-001", "role-001").Return(nil)
	mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	// 执行测试
	err := service.RemoveRole(ctx, "user-001", "role-001", "admin-001")

	// 验证结果
	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
	mockLogRepo.AssertExpectations(t)
}

func TestUserService_GetUserWithRoles(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepositoryForUserService()
	mockRoleRepo := new(MockRoleRepositoryForUserService)
	mockLogRepo := new(MockOperationLogRepositoryForUserService)
	passwordManager := newTestPasswordManager()

	service := NewUserService(mockUserRepo, mockRoleRepo, mockLogRepo, passwordManager)

	// 准备测试数据
	passwordHash, _ := passwordManager.HashPassword("password123")
	existingUser := entity.NewUser("testuser", passwordHash)
	existingUser.ID = "user-001"
	existingUser.Roles = []*entity.Role{
		entity.NewRole("admin", "管理员"),
	}

	// 设置Mock期望
	mockUserRepo.On("GetWithRoles", ctx, "user-001").Return(existingUser, nil)

	// 执行测试
	user, err := service.GetUserWithRoles(ctx, "user-001")

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Len(t, user.Roles, 1)

	mockUserRepo.AssertExpectations(t)
}
