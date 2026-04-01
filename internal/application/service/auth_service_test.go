package service

import (
	"context"
	"errors"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepositoryForAuth 认证服务用户仓储Mock
type MockUserRepositoryForAuth struct {
	mock.Mock
}

func (m *MockUserRepositoryForAuth) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepositoryForAuth) Update(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepositoryForAuth) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepositoryForAuth) GetByID(ctx context.Context, id string) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepositoryForAuth) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepositoryForAuth) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepositoryForAuth) List(ctx context.Context, status *entity.UserStatus, page, pageSize int) ([]*entity.User, int64, error) {
	args := m.Called(ctx, status, page, pageSize)
	return args.Get(0).([]*entity.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepositoryForAuth) GetWithRoles(ctx context.Context, id string) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepositoryForAuth) GetWithPermissions(ctx context.Context, id string) (*entity.User, []*entity.Permission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(*entity.User), args.Get(1).([]*entity.Permission), args.Error(2)
}

func (m *MockUserRepositoryForAuth) AssignRole(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockUserRepositoryForAuth) RemoveRole(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockUserRepositoryForAuth) UpdateLastLogin(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepositoryForAuth) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockUserRepositoryForAuth) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(bool), args.Error(1)
}

// MockRoleRepositoryForAuth 角色仓储Mock
type MockRoleRepositoryForAuth struct {
	mock.Mock
}

func (m *MockRoleRepositoryForAuth) Create(ctx context.Context, role *entity.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepositoryForAuth) Update(ctx context.Context, role *entity.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepositoryForAuth) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRoleRepositoryForAuth) GetByID(ctx context.Context, id string) (*entity.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Role), args.Error(1)
}

func (m *MockRoleRepositoryForAuth) GetByCode(ctx context.Context, code string) (*entity.Role, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Role), args.Error(1)
}

func (m *MockRoleRepositoryForAuth) List(ctx context.Context) ([]*entity.Role, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entity.Role), args.Error(1)
}

func (m *MockRoleRepositoryForAuth) GetWithPermissions(ctx context.Context, id string) (*entity.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Role), args.Error(1)
}

func (m *MockRoleRepositoryForAuth) AssignPermission(ctx context.Context, roleID, permissionID string) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}

func (m *MockRoleRepositoryForAuth) RemovePermission(ctx context.Context, roleID, permissionID string) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}

func (m *MockRoleRepositoryForAuth) ExistsByCode(ctx context.Context, code string) (bool, error) {
	args := m.Called(ctx, code)
	return args.Get(0).(bool), args.Error(1)
}

// MockPermissionRepositoryForAuth 权限仓储Mock
type MockPermissionRepositoryForAuth struct {
	mock.Mock
}

func (m *MockPermissionRepositoryForAuth) Create(ctx context.Context, permission *entity.Permission) error {
	args := m.Called(ctx, permission)
	return args.Error(0)
}

func (m *MockPermissionRepositoryForAuth) BatchCreate(ctx context.Context, permissions []*entity.Permission) error {
	args := m.Called(ctx, permissions)
	return args.Error(0)
}

func (m *MockPermissionRepositoryForAuth) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPermissionRepositoryForAuth) GetByID(ctx context.Context, id string) (*entity.Permission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Permission), args.Error(1)
}

func (m *MockPermissionRepositoryForAuth) GetByCode(ctx context.Context, code string) (*entity.Permission, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Permission), args.Error(1)
}

func (m *MockPermissionRepositoryForAuth) List(ctx context.Context, resourceType *string) ([]*entity.Permission, error) {
	args := m.Called(ctx, resourceType)
	return args.Get(0).([]*entity.Permission), args.Error(1)
}

func (m *MockPermissionRepositoryForAuth) GetByRoleID(ctx context.Context, roleID string) ([]*entity.Permission, error) {
	args := m.Called(ctx, roleID)
	return args.Get(0).([]*entity.Permission), args.Error(1)
}

func (m *MockPermissionRepositoryForAuth) GetByUserID(ctx context.Context, userID string) ([]*entity.Permission, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*entity.Permission), args.Error(1)
}

func (m *MockPermissionRepositoryForAuth) ExistsByCode(ctx context.Context, code string) (bool, error) {
	args := m.Called(ctx, code)
	return args.Get(0).(bool), args.Error(1)
}

// MockOperationLogRepository 操作日志仓储Mock
type MockOperationLogRepository struct {
	mock.Mock
}

func (m *MockOperationLogRepository) Create(ctx context.Context, log *entity.OperationLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockOperationLogRepository) GetByID(ctx context.Context, id string) (*entity.OperationLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.OperationLog), args.Error(1)
}

func (m *MockOperationLogRepository) List(ctx context.Context, userID *string, action *string, startTime, endTime int64, page, pageSize int) ([]*entity.OperationLog, int64, error) {
	args := m.Called(ctx, userID, action, startTime, endTime, page, pageSize)
	return args.Get(0).([]*entity.OperationLog), args.Get(1).(int64), args.Error(2)
}

func (m *MockOperationLogRepository) GetByUserID(ctx context.Context, userID string, page, pageSize int) ([]*entity.OperationLog, int64, error) {
	args := m.Called(ctx, userID, page, pageSize)
	return args.Get(0).([]*entity.OperationLog), args.Get(1).(int64), args.Error(2)
}

// 创建测试用的JWT管理器
func newTestJWTManager() *auth.JWTManager {
	return auth.NewJWTManager(&auth.JWTConfig{
		Secret:        "test-secret-key-for-unit-testing",
		AccessExpire:  3600,
		RefreshExpire: 86400,
	})
}

// 创建测试用的密码管理器
func newTestPasswordManager() *auth.PasswordManager {
	return auth.NewPasswordManager(&auth.PasswordConfig{
		MinLength:        6,
		RequireUppercase: false,
		RequireLowercase: false,
		RequireDigit:     false,
	})
}

func TestAuthService_Login_Success(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := new(MockUserRepositoryForAuth)
	mockRoleRepo := new(MockRoleRepositoryForAuth)
	mockPermRepo := new(MockPermissionRepositoryForAuth)
	mockLogRepo := new(MockOperationLogRepository)
	jwtManager := newTestJWTManager()
	passwordManager := newTestPasswordManager()

	service := NewAuthService(mockUserRepo, mockRoleRepo, mockPermRepo, mockLogRepo, jwtManager, passwordManager)

	// 准备测试数据
	passwordHash, _ := passwordManager.HashPassword("password123")
	user := entity.NewUser("testuser", passwordHash)
	user.ID = "user-001"
	user.Email = "test@example.com"
	user.RealName = "测试用户"

	// 设置Mock期望
	mockUserRepo.On("GetByUsername", ctx, "testuser").Return(user, nil)
	mockUserRepo.On("GetWithPermissions", ctx, "user-001").Return(user, []*entity.Permission{}, nil)
	mockUserRepo.On("UpdateLastLogin", ctx, "user-001").Return(nil)
	mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	// 执行测试
	req := &LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	resp, err := service.Login(ctx, req, "127.0.0.1", "test-agent")

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.NotNil(t, resp.User)
	assert.Equal(t, "user-001", resp.User.ID)
	assert.Equal(t, "testuser", resp.User.Username)

	mockUserRepo.AssertExpectations(t)
	mockLogRepo.AssertExpectations(t)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := new(MockUserRepositoryForAuth)
	mockRoleRepo := new(MockRoleRepositoryForAuth)
	mockPermRepo := new(MockPermissionRepositoryForAuth)
	mockLogRepo := new(MockOperationLogRepository)
	jwtManager := newTestJWTManager()
	passwordManager := newTestPasswordManager()

	service := NewAuthService(mockUserRepo, mockRoleRepo, mockPermRepo, mockLogRepo, jwtManager, passwordManager)

	// 设置Mock期望 - 用户不存在
	mockUserRepo.On("GetByUsername", ctx, "nonexistent").Return(nil, errors.New("user not found"))
	mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	// 执行测试
	req := &LoginRequest{
		Username: "nonexistent",
		Password: "password123",
	}
	resp, err := service.Login(ctx, req, "127.0.0.1", "test-agent")

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, resp)

	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Login_UserDisabled(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := new(MockUserRepositoryForAuth)
	mockRoleRepo := new(MockRoleRepositoryForAuth)
	mockPermRepo := new(MockPermissionRepositoryForAuth)
	mockLogRepo := new(MockOperationLogRepository)
	jwtManager := newTestJWTManager()
	passwordManager := newTestPasswordManager()

	service := NewAuthService(mockUserRepo, mockRoleRepo, mockPermRepo, mockLogRepo, jwtManager, passwordManager)

	// 准备测试数据 - 已禁用用户
	passwordHash, _ := passwordManager.HashPassword("password123")
	user := entity.NewUser("disableduser", passwordHash)
	user.ID = "user-002"
	user.Deactivate()

	// 设置Mock期望
	mockUserRepo.On("GetByUsername", ctx, "disableduser").Return(user, nil)

	// 执行测试
	req := &LoginRequest{
		Username: "disableduser",
		Password: "password123",
	}
	resp, err := service.Login(ctx, req, "127.0.0.1", "test-agent")

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, ErrUserDisabled, err)
	assert.Nil(t, resp)

	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := new(MockUserRepositoryForAuth)
	mockRoleRepo := new(MockRoleRepositoryForAuth)
	mockPermRepo := new(MockPermissionRepositoryForAuth)
	mockLogRepo := new(MockOperationLogRepository)
	jwtManager := newTestJWTManager()
	passwordManager := newTestPasswordManager()

	service := NewAuthService(mockUserRepo, mockRoleRepo, mockPermRepo, mockLogRepo, jwtManager, passwordManager)

	// 准备测试数据
	passwordHash, _ := passwordManager.HashPassword("correctpassword")
	user := entity.NewUser("testuser", passwordHash)
	user.ID = "user-001"

	// 设置Mock期望
	mockUserRepo.On("GetByUsername", ctx, "testuser").Return(user, nil)
	mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	// 执行测试 - 使用错误密码
	req := &LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}
	resp, err := service.Login(ctx, req, "127.0.0.1", "test-agent")

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, resp)

	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_ValidateToken(t *testing.T) {
	_ = context.Background()

	jwtManager := newTestJWTManager()
	passwordManager := newTestPasswordManager()

	mockUserRepo := new(MockUserRepositoryForAuth)
	mockRoleRepo := new(MockRoleRepositoryForAuth)
	mockPermRepo := new(MockPermissionRepositoryForAuth)
	mockLogRepo := new(MockOperationLogRepository)

	service := NewAuthService(mockUserRepo, mockRoleRepo, mockPermRepo, mockLogRepo, jwtManager, passwordManager)

	// 生成测试Token
	tokenPair, _ := jwtManager.GenerateToken("user-001", "testuser", []string{"admin"}, []string{"user:read"})

	// 验证Token
	claims, err := service.ValidateToken(tokenPair.AccessToken)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "user-001", claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Contains(t, claims.Roles, "admin")
	assert.Contains(t, claims.Permissions, "user:read")
}

func TestAuthService_ValidateToken_Invalid(t *testing.T) {
	_ = context.Background()

	jwtManager := newTestJWTManager()
	passwordManager := newTestPasswordManager()

	mockUserRepo := new(MockUserRepositoryForAuth)
	mockRoleRepo := new(MockRoleRepositoryForAuth)
	mockPermRepo := new(MockPermissionRepositoryForAuth)
	mockLogRepo := new(MockOperationLogRepository)

	service := NewAuthService(mockUserRepo, mockRoleRepo, mockPermRepo, mockLogRepo, jwtManager, passwordManager)

	// 使用无效Token
	claims, err := service.ValidateToken("invalid-token")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestAuthService_RefreshToken(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := new(MockUserRepositoryForAuth)
	mockRoleRepo := new(MockRoleRepositoryForAuth)
	mockPermRepo := new(MockPermissionRepositoryForAuth)
	mockLogRepo := new(MockOperationLogRepository)
	jwtManager := newTestJWTManager()
	passwordManager := newTestPasswordManager()

	service := NewAuthService(mockUserRepo, mockRoleRepo, mockPermRepo, mockLogRepo, jwtManager, passwordManager)

	// 准备测试数据
	passwordHash, _ := passwordManager.HashPassword("password123")
	user := entity.NewUser("testuser", passwordHash)
	user.ID = "user-001"
	user.Email = "test@example.com"

	// 生成RefreshToken
	tokenPair, _ := jwtManager.GenerateToken("user-001", "testuser", []string{"admin"}, []string{"user:read"})

	// 设置Mock期望
	mockUserRepo.On("GetByID", ctx, "user-001").Return(user, nil)
	mockUserRepo.On("GetWithPermissions", ctx, "user-001").Return(user, []*entity.Permission{}, nil)

	// 执行测试
	resp, err := service.RefreshToken(ctx, tokenPair.RefreshToken)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)

	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Logout(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := new(MockUserRepositoryForAuth)
	mockRoleRepo := new(MockRoleRepositoryForAuth)
	mockPermRepo := new(MockPermissionRepositoryForAuth)
	mockLogRepo := new(MockOperationLogRepository)
	jwtManager := newTestJWTManager()
	passwordManager := newTestPasswordManager()

	service := NewAuthService(mockUserRepo, mockRoleRepo, mockPermRepo, mockLogRepo, jwtManager, passwordManager)

	// 设置Mock期望
	mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	// 执行测试
	err := service.Logout(ctx, "user-001", "127.0.0.1", "test-agent")

	// 验证结果
	assert.NoError(t, err)

	mockLogRepo.AssertExpectations(t)
}

func TestAuthService_Login_WithRolesAndPermissions(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := new(MockUserRepositoryForAuth)
	mockRoleRepo := new(MockRoleRepositoryForAuth)
	mockPermRepo := new(MockPermissionRepositoryForAuth)
	mockLogRepo := new(MockOperationLogRepository)
	jwtManager := newTestJWTManager()
	passwordManager := newTestPasswordManager()

	service := NewAuthService(mockUserRepo, mockRoleRepo, mockPermRepo, mockLogRepo, jwtManager, passwordManager)

	// 准备测试数据
	passwordHash, _ := passwordManager.HashPassword("password123")
	user := entity.NewUser("adminuser", passwordHash)
	user.ID = "user-001"
	user.Email = "admin@example.com"

	// 准备角色和权限
	roles := []*entity.Role{
		{ID: "role-001", Code: "admin", Name: "管理员"},
	}
	permissions := []*entity.Permission{
		{ID: "perm-001", Code: "user:read", Name: "查看用户"},
		{ID: "perm-002", Code: "user:create", Name: "创建用户"},
	}

	userWithRoles := &entity.User{
		ID:       "user-001",
		Username: "adminuser",
		Roles:    roles,
	}

	// 设置Mock期望
	mockUserRepo.On("GetByUsername", ctx, "adminuser").Return(user, nil)
	mockUserRepo.On("GetWithPermissions", ctx, "user-001").Return(userWithRoles, permissions, nil)
	mockUserRepo.On("UpdateLastLogin", ctx, "user-001").Return(nil)
	mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	// 执行测试
	req := &LoginRequest{
		Username: "adminuser",
		Password: "password123",
	}
	resp, err := service.Login(ctx, req, "127.0.0.1", "test-agent")

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, resp.User.Roles, "admin")
	assert.Contains(t, resp.User.Permissions, "user:read")
	assert.Contains(t, resp.User.Permissions, "user:create")

	mockUserRepo.AssertExpectations(t)
	mockLogRepo.AssertExpectations(t)
}
