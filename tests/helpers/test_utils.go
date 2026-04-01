package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/pkg/auth"
	"github.com/stretchr/testify/assert"
)

// TestConfig 测试配置
type TestConfig struct {
	JWTSecret        string
	AccessExpire     int64
	RefreshExpire    int64
	PasswordMinLen   int
}

// DefaultTestConfig 默认测试配置
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		JWTSecret:      "test-secret-key-for-testing",
		AccessExpire:   3600,      // 1小时
		RefreshExpire:  86400,     // 24小时
		PasswordMinLen: 6,
	}
}

// NewTestJWTManager 创建测试用JWT管理器
func NewTestJWTManager() *auth.JWTManager {
	config := DefaultTestConfig()
	return auth.NewJWTManager(&auth.JWTConfig{
		Secret:        config.JWTSecret,
		AccessExpire:  config.AccessExpire,
		RefreshExpire: config.RefreshExpire,
	})
}

// NewTestPasswordManager 创建测试用密码管理器
func NewTestPasswordManager() *auth.PasswordManager {
	config := DefaultTestConfig()
	return auth.NewPasswordManager(&auth.PasswordConfig{
		MinLength:        config.PasswordMinLen,
		RequireUppercase: false,
		RequireLowercase: false,
		RequireDigit:     false,
	})
}

// GenerateTestID 生成测试ID
func GenerateTestID() string {
	return uuid.New().String()
}

// CreateTestUser 创建测试用户
func CreateTestUser(username, password string) *entity.User {
	passwordManager := NewTestPasswordManager()
	passwordHash, _ := passwordManager.HashPassword(password)
	
	user := entity.NewUser(username, passwordHash)
	user.ID = GenerateTestID()
	user.Email = fmt.Sprintf("%s@test.com", username)
	user.RealName = username
	
	return user
}

// CreateTestRole 创建测试角色
func CreateTestRole(code, name string) *entity.Role {
	role := entity.NewRole(code, name)
	role.ID = GenerateTestID()
	return role
}

// CreateTestPermission 创建测试权限
func CreateTestPermission(code, name, resourceType, action string) *entity.Permission {
	perm := entity.NewPermission(code, name)
	perm.ID = GenerateTestID()
	perm.ResourceType = resourceType
	perm.Action = action
	return perm
}

// CreateTestRegion 创建测试区域
func CreateTestRegion(code, name string, parentID *string) *entity.Region {
	region := entity.NewRegion(code, name, parentID, 1)
	region.ID = GenerateTestID()
	return region
}

// CreateTestStation 创建测试电站
func CreateTestStation(code, name string, stationType entity.StationType, subRegionID string) *entity.Station {
	station := entity.NewStation(code, name, stationType, subRegionID)
	station.ID = GenerateTestID()
	return station
}

// CreateTestDevice 创建测试设备
func CreateTestDevice(code, name string, deviceType entity.DeviceType, stationID string) *entity.Device {
	device := entity.NewDevice(code, name, deviceType, stationID)
	device.ID = GenerateTestID()
	return device
}

// CreateTestAlarm 创建测试告警
func CreateTestAlarm(pointID, deviceID, stationID string, alarmType entity.AlarmType, level entity.AlarmLevel) *entity.Alarm {
	alarm := entity.NewAlarm(pointID, deviceID, stationID, alarmType, level, "测试告警", "这是一个测试告警")
	alarm.ID = GenerateTestID()
	return alarm
}

// CreateTestPoint 创建测试采集点
func CreateTestPoint(code, name string, pointType entity.PointType, deviceID, stationID string) *entity.Point {
	point := entity.NewPoint(code, name, pointType, deviceID, stationID)
	point.ID = GenerateTestID()
	return point
}

// MakeRequest 创建HTTP测试请求
func MakeRequest(method, path string, body interface{}, headers map[string]string) (*http.Request, error) {
	var req *http.Request
	var err error
	
	if body != nil {
		jsonBody, jsonErr := json.Marshal(body)
		if jsonErr != nil {
			return nil, jsonErr
		}
		req, err = http.NewRequest(method, path, strings.NewReader(string(jsonBody)))
	} else {
		req, err = http.NewRequest(method, path, nil)
	}
	
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	return req, nil
}

// ExecuteRequest 执行HTTP测试请求
func ExecuteRequest(router *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

// AssertResponseOK 断言响应状态码为200
func AssertResponseOK(t *testing.T, rr *httptest.ResponseRecorder) {
	assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200")
}

// AssertResponseCreated 断言响应状态码为201
func AssertResponseCreated(t *testing.T, rr *httptest.ResponseRecorder) {
	assert.Equal(t, http.StatusCreated, rr.Code, "Expected status code 201")
}

// AssertResponseBadRequest 断言响应状态码为400
func AssertResponseBadRequest(t *testing.T, rr *httptest.ResponseRecorder) {
	assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected status code 400")
}

// AssertResponseUnauthorized 断言响应状态码为401
func AssertResponseUnauthorized(t *testing.T, rr *httptest.ResponseRecorder) {
	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Expected status code 401")
}

// AssertResponseNotFound 断言响应状态码为404
func AssertResponseNotFound(t *testing.T, rr *httptest.ResponseRecorder) {
	assert.Equal(t, http.StatusNotFound, rr.Code, "Expected status code 404")
}

// ParseResponse 解析JSON响应
func ParseResponse(rr *httptest.ResponseRecorder, v interface{}) error {
	return json.Unmarshal(rr.Body.Bytes(), v)
}

// GenerateTestToken 生成测试用JWT Token
func GenerateTestToken(userID, username string, roles, permissions []string) (string, error) {
	jwtManager := NewTestJWTManager()
	tokenPair, err := jwtManager.GenerateToken(userID, username, roles, permissions)
	if err != nil {
		return "", err
	}
	return tokenPair.AccessToken, nil
}

// SetupTestRouter 创建测试路由
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())
	return router
}

// WaitForCondition 等待条件满足
func WaitForCondition(timeout time.Duration, condition func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// ContextWithTimeout 创建带超时的上下文
func ContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// MockUserRepository 模拟用户仓储
type MockUserRepository struct {
	users map[string]*entity.User
}

// NewMockUserRepository 创建模拟用户仓储
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*entity.User),
	}
}

// Create 创建用户
func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	m.users[user.ID] = user
	return nil
}

// Update 更新用户
func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	m.users[user.ID] = user
	return nil
}

// Delete 删除用户
func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	delete(m.users, id)
	return nil
}

// GetByID 根据ID获取用户
func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

// GetByUsername 根据用户名获取用户
func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

// GetByEmail 根据邮箱获取用户
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

// List 列出用户
func (m *MockUserRepository) List(ctx context.Context, status *entity.UserStatus, page, pageSize int) ([]*entity.User, int64, error) {
	users := make([]*entity.User, 0)
	for _, user := range m.users {
		if status == nil || user.Status == *status {
			users = append(users, user)
		}
	}
	return users, int64(len(users)), nil
}

// GetWithRoles 获取用户及角色
func (m *MockUserRepository) GetWithRoles(ctx context.Context, id string) (*entity.User, error) {
	return m.GetByID(ctx, id)
}

// GetWithPermissions 获取用户及权限
func (m *MockUserRepository) GetWithPermissions(ctx context.Context, id string) (*entity.User, []*entity.Permission, error) {
	user, err := m.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return user, []*entity.Permission{}, nil
}

// AssignRole 分配角色
func (m *MockUserRepository) AssignRole(ctx context.Context, userID, roleID string) error {
	return nil
}

// RemoveRole 移除角色
func (m *MockUserRepository) RemoveRole(ctx context.Context, userID, roleID string) error {
	return nil
}

// UpdateLastLogin 更新最后登录时间
func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	user, ok := m.users[id]
	if !ok {
		return fmt.Errorf("user not found")
	}
	user.UpdateLastLogin()
	return nil
}

// ExistsByUsername 检查用户名是否存在
func (m *MockUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	for _, user := range m.users {
		if user.Username == username {
			return true, nil
		}
	}
	return false, nil
}

// ExistsByEmail 检查邮箱是否存在
func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	for _, user := range m.users {
		if user.Email == email {
			return true, nil
		}
	}
	return false, nil
}

// AddUser 添加用户
func (m *MockUserRepository) AddUser(user *entity.User) {
	m.users[user.ID] = user
}
