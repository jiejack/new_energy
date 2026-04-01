package entity

import (
	"time"
)

type UserStatus int

const (
	UserStatusDisabled UserStatus = 0
	UserStatusActive   UserStatus = 1
)

type User struct {
	ID           string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Username     string     `json:"username" gorm:"type:varchar(100);uniqueIndex;not null"`
	PasswordHash string     `json:"-" gorm:"type:varchar(256);not null"`
	Email        string     `json:"email" gorm:"type:varchar(200);uniqueIndex"`
	Phone        string     `json:"phone" gorm:"type:varchar(50)"`
	RealName     string     `json:"real_name" gorm:"type:varchar(100)"`
	Avatar       string     `json:"avatar" gorm:"type:varchar(500)"`
	Status       UserStatus `json:"status" gorm:"default:1"`
	LastLogin    *time.Time `json:"last_login"`
	LoginCount   int        `json:"login_count" gorm:"default:0"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	Roles        []*Role    `json:"roles" gorm:"many2many:user_roles;"`
}

func (u *User) TableName() string {
	return "users"
}

func NewUser(username, passwordHash string) *User {
	return &User{
		Username:     username,
		PasswordHash: passwordHash,
		Status:       UserStatusActive,
	}
}

func (u *User) SetEmail(email string) {
	u.Email = email
}

func (u *User) SetPhone(phone string) {
	u.Phone = phone
}

func (u *User) SetRealName(realName string) {
	u.RealName = realName
}

func (u *User) SetAvatar(avatar string) {
	u.Avatar = avatar
}

func (u *User) Activate() {
	u.Status = UserStatusActive
}

func (u *User) Deactivate() {
	u.Status = UserStatusDisabled
}

func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLogin = &now
	u.LoginCount++
}

func (u *User) UpdatePassword(passwordHash string) {
	u.PasswordHash = passwordHash
}

type UserRole struct {
	UserID    string    `json:"user_id" gorm:"type:varchar(36);primaryKey"`
	RoleID    string    `json:"role_id" gorm:"type:varchar(36);primaryKey"`
	CreatedAt time.Time `json:"created_at"`
}

func (ur *UserRole) TableName() string {
	return "user_roles"
}
