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

type mockUserRepoForUser struct {
	mock.Mock
}

func (m *mockUserRepoForUser) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
func (m *mockUserRepoForUser) Update(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
func (m *mockUserRepoForUser) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockUserRepoForUser) GetByID(ctx context.Context, id string) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}
func (m *mockUserRepoForUser) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}
func (m *mockUserRepoForUser) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}
func (m *mockUserRepoForUser) List(ctx context.Context, status *entity.UserStatus, page, pageSize int) ([]*entity.User, int64, error) {
	args := m.Called(ctx, status, page, pageSize)
	return args.Get(0).([]*entity.User), args.Get(1).(int64), args.Error(2)
}
func (m *mockUserRepoForUser) GetWithRoles(ctx context.Context, id string) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}
func (m *mockUserRepoForUser) GetWithPermissions(ctx context.Context, id string) (*entity.User, []*entity.Permission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(*entity.User), args.Get(1).([]*entity.Permission), args.Error(2)
}
func (m *mockUserRepoForUser) AssignRole(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}
func (m *mockUserRepoForUser) RemoveRole(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}
func (m *mockUserRepoForUser) UpdateLastLogin(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockUserRepoForUser) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}
func (m *mockUserRepoForUser) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

type mockRoleRepoForUser struct {
	mock.Mock
}

func (m *mockRoleRepoForUser) Create(ctx context.Context, role *entity.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}
func (m *mockRoleRepoForUser) Update(ctx context.Context, role *entity.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}
func (m *mockRoleRepoForUser) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockRoleRepoForUser) GetByID(ctx context.Context, id string) (*entity.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Role), args.Error(1)
}
func (m *mockRoleRepoForUser) GetByCode(ctx context.Context, code string) (*entity.Role, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Role), args.Error(1)
}
func (m *mockRoleRepoForUser) List(ctx context.Context) ([]*entity.Role, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Role), args.Error(1)
}
func (m *mockRoleRepoForUser) GetWithPermissions(ctx context.Context, id string) (*entity.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Role), args.Error(1)
}
func (m *mockRoleRepoForUser) AssignPermission(ctx context.Context, roleID, permissionID string) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}
func (m *mockRoleRepoForUser) RemovePermission(ctx context.Context, roleID, permissionID string) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}
func (m *mockRoleRepoForUser) ExistsByCode(ctx context.Context, code string) (bool, error) {
	args := m.Called(ctx, code)
	return args.Bool(0), args.Error(1)
}

type mockLogRepoForUser struct {
	mock.Mock
}

func (m *mockLogRepoForUser) Create(ctx context.Context, log *entity.OperationLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}
func (m *mockLogRepoForUser) GetByID(ctx context.Context, id string) (*entity.OperationLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.OperationLog), args.Error(1)
}
func (m *mockLogRepoForUser) List(ctx context.Context, query *repository.OperationLogQuery) ([]*entity.OperationLog, int64, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]*entity.OperationLog), args.Get(1).(int64), args.Error(2)
}
func (m *mockLogRepoForUser) DeleteBefore(ctx context.Context, before int64) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

func setupUserHandler(userRepo *mockUserRepoForUser, roleRepo *mockRoleRepoForUser, logRepo *mockLogRepoForUser) (*UserHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	passwordManager := auth.NewPasswordManager(&auth.PasswordConfig{
		MinLength:        6,
		RequireUppercase: false,
		RequireLowercase: false,
		RequireDigit:     false,
	})
	svc := service.NewUserService(userRepo, roleRepo, logRepo, passwordManager)
	handler := NewUserHandler(svc)
	r := gin.New()
	return handler, r
}

func TestUserHandler_CreateUser_Success(t *testing.T) {
	userRepo := new(mockUserRepoForUser)
	roleRepo := new(mockRoleRepoForUser)
	logRepo := new(mockLogRepoForUser)
	handler, r := setupUserHandler(userRepo, roleRepo, logRepo)

	r.POST("/users", handler.CreateUser)

	userRepo.On("ExistsByUsername", mock.Anything, "testuser").Return(false, nil)
	userRepo.On("ExistsByEmail", mock.Anything, "test@example.com").Return(false, nil)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
	logRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	body := service.CreateUserRequest{Username: "testuser", Password: "Password1", Email: "test@example.com"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestUserHandler_CreateUser_InvalidJSON(t *testing.T) {
	userRepo := new(mockUserRepoForUser)
	roleRepo := new(mockRoleRepoForUser)
	logRepo := new(mockLogRepoForUser)
	handler, r := setupUserHandler(userRepo, roleRepo, logRepo)

	r.POST("/users", handler.CreateUser)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_CreateUser_Error(t *testing.T) {
	userRepo := new(mockUserRepoForUser)
	roleRepo := new(mockRoleRepoForUser)
	logRepo := new(mockLogRepoForUser)
	handler, r := setupUserHandler(userRepo, roleRepo, logRepo)

	r.POST("/users", handler.CreateUser)

	userRepo.On("ExistsByUsername", mock.Anything, "testuser").Return(true, nil)

	body := service.CreateUserRequest{Username: "testuser", Password: "Password1"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUserHandler_GetUser_Success(t *testing.T) {
	userRepo := new(mockUserRepoForUser)
	roleRepo := new(mockRoleRepoForUser)
	logRepo := new(mockLogRepoForUser)
	handler, r := setupUserHandler(userRepo, roleRepo, logRepo)

	r.GET("/users/:id", handler.GetUser)

	user := entity.NewUser("testuser", "hash")
	user.ID = "user-001"
	userRepo.On("GetByID", mock.Anything, "user-001").Return(user, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/user-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_GetUser_NotFound(t *testing.T) {
	userRepo := new(mockUserRepoForUser)
	roleRepo := new(mockRoleRepoForUser)
	logRepo := new(mockLogRepoForUser)
	handler, r := setupUserHandler(userRepo, roleRepo, logRepo)

	r.GET("/users/:id", handler.GetUser)

	userRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserHandler_ListUsers_Success(t *testing.T) {
	userRepo := new(mockUserRepoForUser)
	roleRepo := new(mockRoleRepoForUser)
	logRepo := new(mockLogRepoForUser)
	handler, r := setupUserHandler(userRepo, roleRepo, logRepo)

	r.GET("/users", handler.ListUsers)

	userRepo.On("List", mock.Anything, (*entity.UserStatus)(nil), 1, 20).Return([]*entity.User{}, int64(0), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_ListUsers_Error(t *testing.T) {
	userRepo := new(mockUserRepoForUser)
	roleRepo := new(mockRoleRepoForUser)
	logRepo := new(mockLogRepoForUser)
	handler, r := setupUserHandler(userRepo, roleRepo, logRepo)

	r.GET("/users", handler.ListUsers)

	userRepo.On("List", mock.Anything, (*entity.UserStatus)(nil), 1, 20).Return([]*entity.User{}, int64(0), assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUserHandler_UpdateUser_Success(t *testing.T) {
	userRepo := new(mockUserRepoForUser)
	roleRepo := new(mockRoleRepoForUser)
	logRepo := new(mockLogRepoForUser)
	handler, r := setupUserHandler(userRepo, roleRepo, logRepo)

	r.PUT("/users/:id", handler.UpdateUser)

	user := entity.NewUser("testuser", "hash")
	user.ID = "user-001"
	userRepo.On("GetByID", mock.Anything, "user-001").Return(user, nil)
	userRepo.On("ExistsByEmail", mock.Anything, "new@example.com").Return(false, nil)
	userRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
	logRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	body := service.UpdateUserRequest{Email: "new@example.com", RealName: "New Name"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/users/user-001", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_UpdateUser_InvalidJSON(t *testing.T) {
	userRepo := new(mockUserRepoForUser)
	roleRepo := new(mockRoleRepoForUser)
	logRepo := new(mockLogRepoForUser)
	handler, r := setupUserHandler(userRepo, roleRepo, logRepo)

	r.PUT("/users/:id", handler.UpdateUser)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/users/user-001", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_UpdateUser_NotFound(t *testing.T) {
	userRepo := new(mockUserRepoForUser)
	roleRepo := new(mockRoleRepoForUser)
	logRepo := new(mockLogRepoForUser)
	handler, r := setupUserHandler(userRepo, roleRepo, logRepo)

	r.PUT("/users/:id", handler.UpdateUser)

	userRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	body := service.UpdateUserRequest{Email: "new@example.com"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/users/nonexistent", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUserHandler_DeleteUser_Success(t *testing.T) {
	userRepo := new(mockUserRepoForUser)
	roleRepo := new(mockRoleRepoForUser)
	logRepo := new(mockLogRepoForUser)
	handler, r := setupUserHandler(userRepo, roleRepo, logRepo)

	r.DELETE("/users/:id", handler.DeleteUser)

	user := entity.NewUser("testuser", "hash")
	user.ID = "user-001"
	userRepo.On("GetByID", mock.Anything, "user-001").Return(user, nil)
	userRepo.On("Delete", mock.Anything, "user-001").Return(nil)
	logRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/users/user-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestUserHandler_DeleteUser_NotFound(t *testing.T) {
	userRepo := new(mockUserRepoForUser)
	roleRepo := new(mockRoleRepoForUser)
	logRepo := new(mockLogRepoForUser)
	handler, r := setupUserHandler(userRepo, roleRepo, logRepo)

	r.DELETE("/users/:id", handler.DeleteUser)

	userRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/users/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserHandler_ChangePassword_Success(t *testing.T) {
	userRepo := new(mockUserRepoForUser)
	roleRepo := new(mockRoleRepoForUser)
	logRepo := new(mockLogRepoForUser)
	handler, r := setupUserHandler(userRepo, roleRepo, logRepo)

	r.PUT("/users/:id/password", handler.ChangePassword)

	pm := auth.NewPasswordManager(&auth.PasswordConfig{MinLength: 1})
	hashedPwd, _ := pm.HashPassword("OldPass1")
	user := entity.NewUser("testuser", hashedPwd)
	user.ID = "user-001"
	userRepo.On("GetByID", mock.Anything, "user-001").Return(user, nil)
	userRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
	logRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	body := service.ChangePasswordRequest{OldPassword: "OldPass1", NewPassword: "NewPass1"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/users/user-001/password", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_ChangePassword_InvalidJSON(t *testing.T) {
	userRepo := new(mockUserRepoForUser)
	roleRepo := new(mockRoleRepoForUser)
	logRepo := new(mockLogRepoForUser)
	handler, r := setupUserHandler(userRepo, roleRepo, logRepo)

	r.PUT("/users/:id/password", handler.ChangePassword)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/users/user-001/password", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_ChangePassword_WrongOldPassword(t *testing.T) {
	userRepo := new(mockUserRepoForUser)
	roleRepo := new(mockRoleRepoForUser)
	logRepo := new(mockLogRepoForUser)
	handler, r := setupUserHandler(userRepo, roleRepo, logRepo)

	r.PUT("/users/:id/password", handler.ChangePassword)

	pm := auth.NewPasswordManager(&auth.PasswordConfig{MinLength: 1})
	hashedPwd, _ := pm.HashPassword("CorrectOld1")
	user := entity.NewUser("testuser", hashedPwd)
	user.ID = "user-001"
	userRepo.On("GetByID", mock.Anything, "user-001").Return(user, nil)

	body := service.ChangePasswordRequest{OldPassword: "WrongOld1", NewPassword: "NewPass1"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/users/user-001/password", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_NewUserHandler(t *testing.T) {
	userRepo := new(mockUserRepoForUser)
	roleRepo := new(mockRoleRepoForUser)
	logRepo := new(mockLogRepoForUser)
	pm := auth.NewPasswordManager(&auth.PasswordConfig{MinLength: 1})
	svc := service.NewUserService(userRepo, roleRepo, logRepo, pm)
	handler := NewUserHandler(svc)
	assert.NotNil(t, handler)
}
