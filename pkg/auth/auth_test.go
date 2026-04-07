package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewJWTManager(t *testing.T) {
	config := &JWTConfig{
		Secret:        "test-secret-key",
		AccessExpire:  3600,
		RefreshExpire: 86400,
	}

	manager := NewJWTManager(config)

	assert.NotNil(t, manager)
	assert.Equal(t, config, manager.config)
}

func TestJWTManager_GenerateToken(t *testing.T) {
	config := &JWTConfig{
		Secret:        "test-secret-key",
		AccessExpire:  3600,
		RefreshExpire: 86400,
	}

	manager := NewJWTManager(config)

	userID := "user-001"
	username := "testuser"
	roles := []string{"admin", "operator"}
	permissions := []string{"read", "write", "delete"}

	tokenPair, err := manager.GenerateToken(userID, username, roles, permissions)

	assert.NoError(t, err)
	assert.NotNil(t, tokenPair)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
}

func TestJWTManager_ValidateToken(t *testing.T) {
	config := &JWTConfig{
		Secret:        "test-secret-key",
		AccessExpire:  3600,
		RefreshExpire: 86400,
	}

	manager := NewJWTManager(config)

	// 生成 token
	tokenPair, err := manager.GenerateToken("user-001", "testuser", []string{"admin"}, []string{"read"})
	assert.NoError(t, err)

	// 验证 token
	valid, err := manager.ValidateToken(tokenPair.AccessToken)
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestJWTManager_ValidateToken_Invalid(t *testing.T) {
	config := &JWTConfig{
		Secret:        "test-secret-key",
		AccessExpire:  3600,
		RefreshExpire: 86400,
	}

	manager := NewJWTManager(config)

	// 验证无效 token
	valid, err := manager.ValidateToken("invalid-token")
	assert.Error(t, err)
	assert.False(t, valid)
}

func TestJWTManager_ParseToken(t *testing.T) {
	config := &JWTConfig{
		Secret:        "test-secret-key",
		AccessExpire:  3600,
		RefreshExpire: 86400,
	}

	manager := NewJWTManager(config)

	userID := "user-001"
	username := "testuser"
	roles := []string{"admin", "operator"}
	permissions := []string{"read", "write"}

	// 生成 token
	tokenPair, err := manager.GenerateToken(userID, username, roles, permissions)
	assert.NoError(t, err)

	// 解析 token
	claims, err := manager.ParseToken(tokenPair.AccessToken)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, roles, claims.Roles)
	assert.Equal(t, permissions, claims.Permissions)
}

func TestJWTManager_ParseToken_Invalid(t *testing.T) {
	config := &JWTConfig{
		Secret:        "test-secret-key",
		AccessExpire:  3600,
		RefreshExpire: 86400,
	}

	manager := NewJWTManager(config)

	// 解析无效 token
	claims, err := manager.ParseToken("invalid-token")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTManager_GetUserIDFromToken(t *testing.T) {
	config := &JWTConfig{
		Secret:        "test-secret-key",
		AccessExpire:  3600,
		RefreshExpire: 86400,
	}

	manager := NewJWTManager(config)

	userID := "user-001"

	// 生成 token
	tokenPair, err := manager.GenerateToken(userID, "testuser", []string{"admin"}, []string{"read"})
	assert.NoError(t, err)

	// 从 token 获取用户 ID
	extractedUserID, err := manager.GetUserIDFromToken(tokenPair.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, extractedUserID)
}

func TestJWTManager_RefreshAccessToken(t *testing.T) {
	config := &JWTConfig{
		Secret:        "test-secret-key",
		AccessExpire:  3600,
		RefreshExpire: 86400,
	}

	manager := NewJWTManager(config)

	// 生成 token
	tokenPair, err := manager.GenerateToken("user-001", "testuser", []string{"admin"}, []string{"read"})
	assert.NoError(t, err)

	// 等待一小段时间确保 access token 和 refresh token 不同
	time.Sleep(100 * time.Millisecond)

	// 刷新 access token
	newTokenPair, err := manager.RefreshAccessToken(tokenPair.RefreshToken)
	assert.NoError(t, err)
	assert.NotNil(t, newTokenPair)
	assert.NotEmpty(t, newTokenPair.AccessToken)
	assert.NotEmpty(t, newTokenPair.RefreshToken)
}

func TestJWTManager_RefreshAccessToken_Invalid(t *testing.T) {
	config := &JWTConfig{
		Secret:        "test-secret-key",
		AccessExpire:  3600,
		RefreshExpire: 86400,
	}

	manager := NewJWTManager(config)

	// 使用无效 refresh token
	newTokenPair, err := manager.RefreshAccessToken("invalid-refresh-token")
	assert.Error(t, err)
	assert.Nil(t, newTokenPair)
}

func TestNewPasswordManager(t *testing.T) {
	config := &PasswordConfig{
		MinLength:        8,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
	}

	manager := NewPasswordManager(config)

	assert.NotNil(t, manager)
	assert.Equal(t, config, manager.config)
}

func TestPasswordManager_HashPassword(t *testing.T) {
	config := &PasswordConfig{
		MinLength:        8,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
	}

	manager := NewPasswordManager(config)

	password := "TestPass123"

	hash, err := manager.HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestPasswordManager_HashPassword_TooShort(t *testing.T) {
	config := &PasswordConfig{
		MinLength:        8,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
	}

	manager := NewPasswordManager(config)

	password := "short"

	hash, err := manager.HashPassword(password)

	assert.Error(t, err)
	assert.Equal(t, ErrPasswordTooShort, err)
	assert.Empty(t, hash)
}

func TestPasswordManager_HashPassword_MissingUppercase(t *testing.T) {
	config := &PasswordConfig{
		MinLength:        8,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
	}

	manager := NewPasswordManager(config)

	password := "testpass123" // 没有大写字母

	hash, err := manager.HashPassword(password)

	assert.Error(t, err)
	assert.Empty(t, hash)
}

func TestPasswordManager_HashPassword_MissingLowercase(t *testing.T) {
	config := &PasswordConfig{
		MinLength:        8,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
	}

	manager := NewPasswordManager(config)

	password := "TESTPASS123" // 没有小写字母

	hash, err := manager.HashPassword(password)

	assert.Error(t, err)
	assert.Empty(t, hash)
}

func TestPasswordManager_HashPassword_MissingDigit(t *testing.T) {
	config := &PasswordConfig{
		MinLength:        8,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
	}

	manager := NewPasswordManager(config)

	password := "TestPassword" // 没有数字

	hash, err := manager.HashPassword(password)

	assert.Error(t, err)
	assert.Empty(t, hash)
}

func TestPasswordManager_CheckPassword(t *testing.T) {
	config := &PasswordConfig{
		MinLength:        8,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
	}

	manager := NewPasswordManager(config)

	password := "TestPass123"

	hash, err := manager.HashPassword(password)
	assert.NoError(t, err)

	// 验证正确的密码
	valid := manager.CheckPassword(password, hash)
	assert.True(t, valid)

	// 验证错误的密码
	valid = manager.CheckPassword("WrongPass123", hash)
	assert.False(t, valid)
}

func TestPasswordManager_ValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		config   *PasswordConfig
		password string
		hasError bool
	}{
		{
			name: "有效密码",
			config: &PasswordConfig{
				MinLength:        8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireDigit:     true,
			},
			password: "TestPass123",
			hasError: false,
		},
		{
			name: "密码太短",
			config: &PasswordConfig{
				MinLength:        8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireDigit:     true,
			},
			password: "short",
			hasError: true,
		},
		{
			name: "缺少大写字母",
			config: &PasswordConfig{
				MinLength:        8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireDigit:     true,
			},
			password: "testpass123",
			hasError: true,
		},
		{
			name: "缺少小写字母",
			config: &PasswordConfig{
				MinLength:        8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireDigit:     true,
			},
			password: "TESTPASS123",
			hasError: true,
		},
		{
			name: "缺少数字",
			config: &PasswordConfig{
				MinLength:        8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireDigit:     true,
			},
			password: "TestPassword",
			hasError: true,
		},
		{
			name: "不要求大写字母",
			config: &PasswordConfig{
				MinLength:        8,
				RequireUppercase: false,
				RequireLowercase: true,
				RequireDigit:     true,
			},
			password: "testpass123",
			hasError: false,
		},
		{
			name: "不要求小写字母",
			config: &PasswordConfig{
				MinLength:        8,
				RequireUppercase: true,
				RequireLowercase: false,
				RequireDigit:     true,
			},
			password: "TESTPASS123",
			hasError: false,
		},
		{
			name: "不求数字",
			config: &PasswordConfig{
				MinLength:        8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireDigit:     false,
			},
			password: "TestPassword",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewPasswordManager(tt.config)
			err := manager.ValidatePassword(tt.password)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHasUppercase(t *testing.T) {
	assert.True(t, hasUppercase("Test"))
	assert.True(t, hasUppercase("TEST"))
	assert.False(t, hasUppercase("test"))
	assert.False(t, hasUppercase("123"))
}

func TestHasLowercase(t *testing.T) {
	assert.True(t, hasLowercase("Test"))
	assert.True(t, hasLowercase("test"))
	assert.False(t, hasLowercase("TEST"))
	assert.False(t, hasLowercase("123"))
}

func TestHasDigit(t *testing.T) {
	assert.True(t, hasDigit("Test123"))
	assert.True(t, hasDigit("123"))
	assert.False(t, hasDigit("Test"))
	assert.False(t, hasDigit("test"))
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email   string
		isValid bool
	}{
		{"test@example.com", true},
		{"user.name@example.com", true},
		{"user+tag@example.com", true},
		{"test@subdomain.example.com", true},
		{"", false},
		{"invalid", false},
		{"invalid@", false},
		{"@example.com", false},
		{"test@", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := IsValidEmail(tt.email)
			assert.Equal(t, tt.isValid, result)
		})
	}
}

func TestIsValidPhone(t *testing.T) {
	tests := []struct {
		phone   string
		isValid bool
	}{
		{"13800138000", true},
		{"15912345678", true},
		{"18600000000", true},
		{"", false},
		{"12345678901", false}, // 不以1开头
		{"12800138000", false}, // 第二位不是3-9
		{"1380013800", false},  // 位数不够
		{"138001380001", false}, // 位数过多
	}

	for _, tt := range tests {
		t.Run(tt.phone, func(t *testing.T) {
			result := IsValidPhone(tt.phone)
			assert.Equal(t, tt.isValid, result)
		})
	}
}

func TestPasswordManager_ComplexPassword(t *testing.T) {
	config := &PasswordConfig{
		MinLength:        12,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
	}

	manager := NewPasswordManager(config)

	password := "ComplexPass123"

	hash, err := manager.HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	valid := manager.CheckPassword(password, hash)
	assert.True(t, valid)
}

func TestJWTManager_TokenExpiration(t *testing.T) {
	config := &JWTConfig{
		Secret:        "test-secret-key",
		AccessExpire:  1, // 1秒过期
		RefreshExpire: 2, // 2秒过期
	}

	manager := NewJWTManager(config)

	tokenPair, err := manager.GenerateToken("user-001", "testuser", []string{"admin"}, []string{"read"})
	assert.NoError(t, err)

	// 立即验证应该有效
	valid, err := manager.ValidateToken(tokenPair.AccessToken)
	assert.NoError(t, err)
	assert.True(t, valid)

	// 等待 access token 过期
	time.Sleep(2 * time.Second)

	// 过期后验证应该失败
	valid, err = manager.ValidateToken(tokenPair.AccessToken)
	assert.Error(t, err)
	assert.False(t, valid)
}

func TestJWTManager_DifferentSecrets(t *testing.T) {
	config1 := &JWTConfig{
		Secret:        "secret-key-1",
		AccessExpire:  3600,
		RefreshExpire: 86400,
	}

	config2 := &JWTConfig{
		Secret:        "secret-key-2",
		AccessExpire:  3600,
		RefreshExpire: 86400,
	}

	manager1 := NewJWTManager(config1)
	manager2 := NewJWTManager(config2)

	// 使用 manager1 生成 token
	tokenPair, err := manager1.GenerateToken("user-001", "testuser", []string{"admin"}, []string{"read"})
	assert.NoError(t, err)

	// 使用 manager2 验证应该失败（不同的 secret）
	valid, err := manager2.ValidateToken(tokenPair.AccessToken)
	assert.Error(t, err)
	assert.False(t, valid)

	// 使用 manager1 验证应该成功
	valid, err = manager1.ValidateToken(tokenPair.AccessToken)
	assert.NoError(t, err)
	assert.True(t, valid)
}
