package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/new-energy-monitoring/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUserRepoForAuth struct {
	mock.Mock
}

func (m *mockUserRepoForAuth) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
func (m *mockUserRepoForAuth) Update(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
func (m *mockUserRepoForAuth) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockUserRepoForAuth) GetByID(ctx context.Context, id string) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}
func (m *mockUserRepoForAuth) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}
func (m *mockUserRepoForAuth) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}
func (m *mockUserRepoForAuth) List(ctx context.Context, status *entity.UserStatus, page, pageSize int) ([]*entity.User, int64, error) {
	args := m.Called(ctx, status, page, pageSize)
	return args.Get(0).([]*entity.User), args.Get(1).(int64), args.Error(2)
}
func (m *mockUserRepoForAuth) GetWithRoles(ctx context.Context, id string) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}
func (m *mockUserRepoForAuth) GetWithPermissions(ctx context.Context, id string) (*entity.User, []*entity.Permission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(*entity.User), args.Get(1).([]*entity.Permission), args.Error(2)
}
func (m *mockUserRepoForAuth) AssignRole(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}
func (m *mockUserRepoForAuth) RemoveRole(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}
func (m *mockUserRepoForAuth) UpdateLastLogin(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockUserRepoForAuth) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}
func (m *mockUserRepoForAuth) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

type mockRoleRepoForAuth struct {
	mock.Mock
}

func (m *mockRoleRepoForAuth) Create(ctx context.Context, role *entity.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}
func (m *mockRoleRepoForAuth) Update(ctx context.Context, role *entity.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}
func (m *mockRoleRepoForAuth) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockRoleRepoForAuth) GetByID(ctx context.Context, id string) (*entity.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Role), args.Error(1)
}
func (m *mockRoleRepoForAuth) GetByCode(ctx context.Context, code string) (*entity.Role, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Role), args.Error(1)
}
func (m *mockRoleRepoForAuth) List(ctx context.Context) ([]*entity.Role, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Role), args.Error(1)
}
func (m *mockRoleRepoForAuth) GetWithPermissions(ctx context.Context, id string) (*entity.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Role), args.Error(1)
}
func (m *mockRoleRepoForAuth) AssignPermission(ctx context.Context, roleID, permissionID string) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}
func (m *mockRoleRepoForAuth) RemovePermission(ctx context.Context, roleID, permissionID string) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}
func (m *mockRoleRepoForAuth) ExistsByCode(ctx context.Context, code string) (bool, error) {
	args := m.Called(ctx, code)
	return args.Bool(0), args.Error(1)
}

type mockPermRepoForAuth struct {
	mock.Mock
}

func (m *mockPermRepoForAuth) Create(ctx context.Context, permission *entity.Permission) error {
	args := m.Called(ctx, permission)
	return args.Error(0)
}
func (m *mockPermRepoForAuth) BatchCreate(ctx context.Context, permissions []*entity.Permission) error {
	args := m.Called(ctx, permissions)
	return args.Error(0)
}
func (m *mockPermRepoForAuth) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockPermRepoForAuth) GetByID(ctx context.Context, id string) (*entity.Permission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Permission), args.Error(1)
}
func (m *mockPermRepoForAuth) GetByCode(ctx context.Context, code string) (*entity.Permission, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Permission), args.Error(1)
}
func (m *mockPermRepoForAuth) List(ctx context.Context, resourceType *string) ([]*entity.Permission, error) {
	args := m.Called(ctx, resourceType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Permission), args.Error(1)
}
func (m *mockPermRepoForAuth) GetByRoleID(ctx context.Context, roleID string) ([]*entity.Permission, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Permission), args.Error(1)
}
func (m *mockPermRepoForAuth) GetByUserID(ctx context.Context, userID string) ([]*entity.Permission, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Permission), args.Error(1)
}
func (m *mockPermRepoForAuth) ExistsByCode(ctx context.Context, code string) (bool, error) {
	args := m.Called(ctx, code)
	return args.Bool(0), args.Error(1)
}

type mockLogRepoForAuth struct {
	mock.Mock
}

func (m *mockLogRepoForAuth) Create(ctx context.Context, log *entity.OperationLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}
func (m *mockLogRepoForAuth) GetByID(ctx context.Context, id string) (*entity.OperationLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.OperationLog), args.Error(1)
}
func (m *mockLogRepoForAuth) List(ctx context.Context, query *repository.OperationLogQuery) ([]*entity.OperationLog, int64, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]*entity.OperationLog), args.Get(1).(int64), args.Error(2)
}
func (m *mockLogRepoForAuth) DeleteBefore(ctx context.Context, before int64) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

func newTestAuthService(userRepo *mockUserRepoForAuth, roleRepo *mockRoleRepoForAuth, permRepo *mockPermRepoForAuth, logRepo *mockLogRepoForAuth) *service.AuthService {
	jwtManager := auth.NewJWTManager(&auth.JWTConfig{
		Secret:        "test-secret-key-for-testing",
		AccessExpire:  3600,
		RefreshExpire: 86400,
	})
	passwordManager := auth.NewPasswordManager(&auth.PasswordConfig{
		MinLength:        6,
		RequireUppercase: false,
		RequireLowercase: false,
		RequireDigit:     false,
	})
	return service.NewAuthService(userRepo, roleRepo, permRepo, logRepo, jwtManager, passwordManager)
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userRepo := new(mockUserRepoForAuth)
	roleRepo := new(mockRoleRepoForAuth)
	permRepo := new(mockPermRepoForAuth)
	logRepo := new(mockLogRepoForAuth)
	svc := newTestAuthService(userRepo, roleRepo, permRepo, logRepo)
	handler := NewAuthHandler(svc)

	r := gin.New()
	r.POST("/auth/login", handler.Login)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userRepo := new(mockUserRepoForAuth)
	roleRepo := new(mockRoleRepoForAuth)
	permRepo := new(mockPermRepoForAuth)
	logRepo := new(mockLogRepoForAuth)
	svc := newTestAuthService(userRepo, roleRepo, permRepo, logRepo)
	handler := NewAuthHandler(svc)

	r := gin.New()
	r.POST("/auth/login", handler.Login)

	userRepo.On("GetByUsername", mock.Anything, "admin").Return(nil, assert.AnError)
	logRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	body := service.LoginRequest{Username: "admin", Password: "wrongpass"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userRepo := new(mockUserRepoForAuth)
	roleRepo := new(mockRoleRepoForAuth)
	permRepo := new(mockPermRepoForAuth)
	logRepo := new(mockLogRepoForAuth)
	svc := newTestAuthService(userRepo, roleRepo, permRepo, logRepo)
	handler := NewAuthHandler(svc)

	r := gin.New()
	r.POST("/auth/login", handler.Login)

	hashedPwd, _ := auth.NewPasswordManager(&auth.PasswordConfig{MinLength: 1}).HashPassword("Password1")
	user := entity.NewUser("admin", hashedPwd)
	user.ID = "user-001"

	userRepo.On("GetByUsername", mock.Anything, "admin").Return(user, nil)
	userRepo.On("GetWithPermissions", mock.Anything, "user-001").Return(user, []*entity.Permission{}, nil)
	userRepo.On("UpdateLastLogin", mock.Anything, "user-001").Return(nil)
	logRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	body := service.LoginRequest{Username: "admin", Password: "Password1"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userRepo := new(mockUserRepoForAuth)
	roleRepo := new(mockRoleRepoForAuth)
	permRepo := new(mockPermRepoForAuth)
	logRepo := new(mockLogRepoForAuth)
	svc := newTestAuthService(userRepo, roleRepo, permRepo, logRepo)
	handler := NewAuthHandler(svc)

	r := gin.New()
	r.POST("/auth/logout", handler.Logout)

	logRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/logout", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_Logout_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userRepo := new(mockUserRepoForAuth)
	roleRepo := new(mockRoleRepoForAuth)
	permRepo := new(mockPermRepoForAuth)
	logRepo := new(mockLogRepoForAuth)
	svc := newTestAuthService(userRepo, roleRepo, permRepo, logRepo)
	handler := NewAuthHandler(svc)

	r := gin.New()
	r.POST("/auth/logout", handler.Logout)

	logRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).Return(assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/logout", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAuthHandler_RefreshToken_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userRepo := new(mockUserRepoForAuth)
	roleRepo := new(mockRoleRepoForAuth)
	permRepo := new(mockPermRepoForAuth)
	logRepo := new(mockLogRepoForAuth)
	svc := newTestAuthService(userRepo, roleRepo, permRepo, logRepo)
	handler := NewAuthHandler(svc)

	r := gin.New()
	r.POST("/auth/refresh", handler.RefreshToken)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/refresh", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_RefreshToken_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userRepo := new(mockUserRepoForAuth)
	roleRepo := new(mockRoleRepoForAuth)
	permRepo := new(mockPermRepoForAuth)
	logRepo := new(mockLogRepoForAuth)
	svc := newTestAuthService(userRepo, roleRepo, permRepo, logRepo)
	handler := NewAuthHandler(svc)

	r := gin.New()
	r.POST("/auth/refresh", handler.RefreshToken)

	body := map[string]string{"refresh_token": "invalid-token"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_NewAuthHandler(t *testing.T) {
	userRepo := new(mockUserRepoForAuth)
	roleRepo := new(mockRoleRepoForAuth)
	permRepo := new(mockPermRepoForAuth)
	logRepo := new(mockLogRepoForAuth)
	svc := newTestAuthService(userRepo, roleRepo, permRepo, logRepo)
	handler := NewAuthHandler(svc)
	assert.NotNil(t, handler)
}
