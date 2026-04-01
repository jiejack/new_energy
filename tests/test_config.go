package tests

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/pkg/auth"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestConfig 测试配置
type TestConfig struct {
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Server   ServerConfig
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret        string
	AccessExpire  int64
	RefreshExpire int64
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int
	Mode string
}

// DefaultTestConfig 默认测试配置
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		Database: DatabaseConfig{
			Host:     getEnv("TEST_DB_HOST", "localhost"),
			Port:     5432,
			User:     getEnv("TEST_DB_USER", "postgres"),
			Password: getEnv("TEST_DB_PASSWORD", "postgres"),
			DBName:   getEnv("TEST_DB_NAME", "nem_test"),
			SSLMode:  "disable",
		},
		Redis: RedisConfig{
			Host:     getEnv("TEST_REDIS_HOST", "localhost"),
			Port:     6379,
			Password: "",
			DB:       1,
		},
		JWT: JWTConfig{
			Secret:        "test-secret-key-for-unit-testing",
			AccessExpire:  3600,
			RefreshExpire: 86400,
		},
		Server: ServerConfig{
			Port: 8080,
			Mode: "test",
		},
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestDatabase 测试数据库
type TestDatabase struct {
	DB *gorm.DB
}

// NewTestDatabase 创建测试数据库连接
func NewTestDatabase() (*TestDatabase, error) {
	config := DefaultTestConfig()
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.DBName,
		config.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	return &TestDatabase{DB: db}, nil
}

// Migrate 运行数据库迁移
func (td *TestDatabase) Migrate() error {
	return td.DB.AutoMigrate(
		&entity.User{},
		&entity.Role{},
		&entity.Permission{},
		&entity.Region{},
		&entity.SubRegion{},
		&entity.Station{},
		&entity.Device{},
		&entity.Point{},
		&entity.Alarm{},
		&entity.OperationLog{},
		&entity.UserRole{},
		&entity.RolePermission{},
	)
}

// Cleanup 清理测试数据
func (td *TestDatabase) Cleanup() error {
	return td.DB.Exec("TRUNCATE TABLE users, roles, permissions, regions, sub_regions, stations, devices, points, alarms, operation_logs, user_roles, role_permissions CASCADE").Error
}

// Close 关闭数据库连接
func (td *TestDatabase) Close() error {
	sqlDB, err := td.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Transaction 在事务中执行测试
func (td *TestDatabase) Transaction(fn func(tx *gorm.DB) error) error {
	return td.DB.Transaction(fn)
}

// TestJWTManager 创建测试用JWT管理器
func TestJWTManager() *auth.JWTManager {
	config := DefaultTestConfig()
	return auth.NewJWTManager(&auth.JWTConfig{
		Secret:        config.JWT.Secret,
		AccessExpire:  config.JWT.AccessExpire,
		RefreshExpire: config.JWT.RefreshExpire,
	})
}

// TestPasswordManager 创建测试用密码管理器
func TestPasswordManager() *auth.PasswordManager {
	return auth.NewPasswordManager(&auth.PasswordConfig{
		MinLength:        6,
		RequireUppercase: false,
		RequireLowercase: false,
		RequireDigit:     false,
	})
}

// SetupTestUser 创建测试用户
func SetupTestUser(db *gorm.DB, username, password string) (*entity.User, error) {
	passwordManager := TestPasswordManager()
	passwordHash, err := passwordManager.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := entity.NewUser(username, passwordHash)
	user.Email = fmt.Sprintf("%s@test.com", username)
	user.RealName = username

	if err := db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// SetupTestRole 创建测试角色
func SetupTestRole(db *gorm.DB, code, name string) (*entity.Role, error) {
	role := entity.NewRole(code, name)
	if err := db.Create(role).Error; err != nil {
		return nil, err
	}
	return role, nil
}

// SetupTestPermission 创建测试权限
func SetupTestPermission(db *gorm.DB, code, name, resourceType, action string) (*entity.Permission, error) {
	perm := entity.NewPermission(code, name)
	perm.ResourceType = resourceType
	perm.Action = action
	if err := db.Create(perm).Error; err != nil {
		return nil, err
	}
	return perm, nil
}

// GenerateTestToken 生成测试Token
func GenerateTestToken(userID, username string, roles, permissions []string) (string, error) {
	jwtManager := TestJWTManager()
	tokenPair, err := jwtManager.GenerateToken(userID, username, roles, permissions)
	if err != nil {
		return "", err
	}
	return tokenPair.AccessToken, nil
}

// ContextWithTimeout 创建带超时的上下文
func ContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// DefaultTestContext 创建默认测试上下文
func DefaultTestContext() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	go func() {
		<-ctx.Done()
		cancel()
	}()
	return ctx
}

// MockRepository 模拟仓储接口
type MockRepository struct{}

// NewMockRepository 创建模拟仓储
func NewMockRepository() *MockRepository {
	return &MockRepository{}
}

// TestSuite 测试套件
type TestSuite struct {
	DB         *TestDatabase
	JWTManager *auth.JWTManager
	Password   *auth.PasswordManager
	Config     *TestConfig
}

// NewTestSuite 创建测试套件
func NewTestSuite() (*TestSuite, error) {
	config := DefaultTestConfig()
	
	db, err := NewTestDatabase()
	if err != nil {
		return nil, err
	}

	if err := db.Migrate(); err != nil {
		return nil, err
	}

	return &TestSuite{
		DB:         db,
		JWTManager: TestJWTManager(),
		Password:   TestPasswordManager(),
		Config:     config,
	}, nil
}

// Setup 设置测试环境
func (ts *TestSuite) Setup() error {
	return ts.DB.Migrate()
}

// Teardown 清理测试环境
func (ts *TestSuite) Teardown() error {
	if err := ts.DB.Cleanup(); err != nil {
		return err
	}
	return ts.DB.Close()
}

// CreateUser 创建用户
func (ts *TestSuite) CreateUser(username, password string) (*entity.User, error) {
	return SetupTestUser(ts.DB.DB, username, password)
}

// CreateRole 创建角色
func (ts *TestSuite) CreateRole(code, name string) (*entity.Role, error) {
	return SetupTestRole(ts.DB.DB, code, name)
}

// CreatePermission 创建权限
func (ts *TestSuite) CreatePermission(code, name, resourceType, action string) (*entity.Permission, error) {
	return SetupTestPermission(ts.DB.DB, code, name, resourceType, action)
}

// GenerateToken 生成Token
func (ts *TestSuite) GenerateToken(userID, username string, roles, permissions []string) (string, error) {
	return GenerateTestToken(userID, username, roles, permissions)
}
