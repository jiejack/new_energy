package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name         string
		username     string
		passwordHash string
		want         *User
	}{
		{
			name:         "创建普通用户",
			username:     "testuser",
			passwordHash: "hashedpassword123",
			want: &User{
				Username:     "testuser",
				PasswordHash: "hashedpassword123",
				Status:       UserStatusActive,
			},
		},
		{
			name:         "创建管理员用户",
			username:     "admin",
			passwordHash: "adminhash456",
			want: &User{
				Username:     "admin",
				PasswordHash: "adminhash456",
				Status:       UserStatusActive,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewUser(tt.username, tt.passwordHash)
			assert.Equal(t, tt.want.Username, got.Username)
			assert.Equal(t, tt.want.PasswordHash, got.PasswordHash)
			assert.Equal(t, tt.want.Status, got.Status)
		})
	}
}

func TestUser_SetEmail(t *testing.T) {
	user := NewUser("testuser", "hash123")
	assert.Empty(t, user.Email)

	user.SetEmail("test@example.com")
	assert.Equal(t, "test@example.com", user.Email)
}

func TestUser_SetPhone(t *testing.T) {
	user := NewUser("testuser", "hash123")
	assert.Empty(t, user.Phone)

	user.SetPhone("13800138000")
	assert.Equal(t, "13800138000", user.Phone)
}

func TestUser_SetRealName(t *testing.T) {
	user := NewUser("testuser", "hash123")
	assert.Empty(t, user.RealName)

	user.SetRealName("张三")
	assert.Equal(t, "张三", user.RealName)
}

func TestUser_SetAvatar(t *testing.T) {
	user := NewUser("testuser", "hash123")
	assert.Empty(t, user.Avatar)

	user.SetAvatar("https://example.com/avatar.png")
	assert.Equal(t, "https://example.com/avatar.png", user.Avatar)
}

func TestUser_Activate(t *testing.T) {
	user := NewUser("testuser", "hash123")
	user.Status = UserStatusDisabled

	user.Activate()
	assert.Equal(t, UserStatusActive, user.Status)
}

func TestUser_Deactivate(t *testing.T) {
	user := NewUser("testuser", "hash123")
	assert.Equal(t, UserStatusActive, user.Status)

	user.Deactivate()
	assert.Equal(t, UserStatusDisabled, user.Status)
}

func TestUser_IsActive(t *testing.T) {
	user := NewUser("testuser", "hash123")
	assert.True(t, user.IsActive())

	user.Deactivate()
	assert.False(t, user.IsActive())

	user.Activate()
	assert.True(t, user.IsActive())
}

func TestUser_UpdateLastLogin(t *testing.T) {
	user := NewUser("testuser", "hash123")
	assert.Nil(t, user.LastLogin)
	assert.Equal(t, 0, user.LoginCount)

	user.UpdateLastLogin()
	assert.NotNil(t, user.LastLogin)
	assert.Equal(t, 1, user.LoginCount)

	user.UpdateLastLogin()
	assert.Equal(t, 2, user.LoginCount)
}

func TestUser_UpdatePassword(t *testing.T) {
	user := NewUser("testuser", "oldhash")
	assert.Equal(t, "oldhash", user.PasswordHash)

	user.UpdatePassword("newhash")
	assert.Equal(t, "newhash", user.PasswordHash)
}

func TestUser_TableName(t *testing.T) {
	user := User{}
	assert.Equal(t, "users", user.TableName())
}

func TestUserRole_TableName(t *testing.T) {
	userRole := UserRole{}
	assert.Equal(t, "user_roles", userRole.TableName())
}

func TestUserStatus_Constants(t *testing.T) {
	assert.Equal(t, UserStatus(0), UserStatusDisabled)
	assert.Equal(t, UserStatus(1), UserStatusActive)
}
