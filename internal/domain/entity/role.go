package entity

import (
	"time"
)

type Role struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Code        string    `json:"code" gorm:"type:varchar(50);uniqueIndex;not null"`
	Name        string    `json:"name" gorm:"type:varchar(100);not null"`
	Description string    `json:"description" gorm:"type:text"`
	IsSystem    bool      `json:"is_system" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Permissions []*Permission `json:"permissions" gorm:"many2many:role_permissions;"`
	Users       []*User      `json:"users" gorm:"many2many:user_roles;"`
}

func (r *Role) TableName() string {
	return "roles"
}

func NewRole(code, name string) *Role {
	return &Role{
		Code: code,
		Name: name,
	}
}

func (r *Role) SetDescription(description string) {
	r.Description = description
}

func (r *Role) SetAsSystemRole() {
	r.IsSystem = true
}

func (r *Role) IsSystemRole() bool {
	return r.IsSystem
}

type RolePermission struct {
	RoleID       string    `json:"role_id" gorm:"type:varchar(36);primaryKey"`
	PermissionID string    `json:"permission_id" gorm:"type:varchar(36);primaryKey"`
	CreatedAt    time.Time `json:"created_at"`
}

func (rp *RolePermission) TableName() string {
	return "role_permissions"
}

const (
	RoleCodeSuperAdmin = "super_admin"
	RoleCodeAdmin      = "admin"
	RoleCodeOperator   = "operator"
	RoleCodeViewer     = "viewer"
)

func GetDefaultRoles() []*Role {
	return []*Role{
		{
			Code:        RoleCodeSuperAdmin,
			Name:        "超级管理员",
			Description: "系统最高权限",
			IsSystem:    true,
		},
		{
			Code:        RoleCodeAdmin,
			Name:        "系统管理员",
			Description: "系统管理权限",
			IsSystem:    true,
		},
		{
			Code:        RoleCodeOperator,
			Name:        "运维人员",
			Description: "设备操作权限",
			IsSystem:    true,
		},
		{
			Code:        RoleCodeViewer,
			Name:        "查看人员",
			Description: "只读权限",
			IsSystem:    true,
		},
	}
}
