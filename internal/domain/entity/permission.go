package entity

import (
	"time"
)

type Permission struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Code         string    `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name         string    `json:"name" gorm:"type:varchar(100);not null"`
	ResourceType string    `json:"resource_type" gorm:"type:varchar(50)"`
	ResourceID   string    `json:"resource_id" gorm:"type:varchar(36)"`
	Action       string    `json:"action" gorm:"type:varchar(50)"`
	Description  string    `json:"description" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at"`

	Roles []*Role `json:"roles" gorm:"many2many:role_permissions;"`
}

func (p *Permission) TableName() string {
	return "permissions"
}

func NewPermission(code, name string) *Permission {
	return &Permission{
		Code: code,
		Name: name,
	}
}

func (p *Permission) SetResource(resourceType, resourceID string) {
	p.ResourceType = resourceType
	p.ResourceID = resourceID
}

func (p *Permission) SetAction(action string) {
	p.Action = action
}

func (p *Permission) SetDescription(description string) {
	p.Description = description
}

const (
	ActionCreate = "create"
	ActionRead   = "read"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionControl = "control"
	ActionAck    = "ack"
)

const (
	ResourceStation = "station"
	ResourceDevice  = "device"
	ResourcePoint   = "point"
	ResourceAlarm   = "alarm"
	ResourceUser    = "user"
	ResourceRole    = "role"
	ResourceRegion  = "region"
)

func GetDefaultPermissions() []*Permission {
	return []*Permission{
		{
			Code:         "station:create",
			Name:         "创建厂站",
			ResourceType: ResourceStation,
			Action:       ActionCreate,
			Description:  "创建新厂站",
		},
		{
			Code:         "station:read",
			Name:         "查看厂站",
			ResourceType: ResourceStation,
			Action:       ActionRead,
			Description:  "查看厂站信息",
		},
		{
			Code:         "station:update",
			Name:         "更新厂站",
			ResourceType: ResourceStation,
			Action:       ActionUpdate,
			Description:  "更新厂站信息",
		},
		{
			Code:         "station:delete",
			Name:         "删除厂站",
			ResourceType: ResourceStation,
			Action:       ActionDelete,
			Description:  "删除厂站",
		},
		{
			Code:         "device:create",
			Name:         "创建设备",
			ResourceType: ResourceDevice,
			Action:       ActionCreate,
			Description:  "创建新设备",
		},
		{
			Code:         "device:read",
			Name:         "查看设备",
			ResourceType: ResourceDevice,
			Action:       ActionRead,
			Description:  "查看设备信息",
		},
		{
			Code:         "device:update",
			Name:         "更新设备",
			ResourceType: ResourceDevice,
			Action:       ActionUpdate,
			Description:  "更新设备信息",
		},
		{
			Code:         "device:delete",
			Name:         "删除设备",
			ResourceType: ResourceDevice,
			Action:       ActionDelete,
			Description:  "删除设备",
		},
		{
			Code:         "device:control",
			Name:         "设备控制",
			ResourceType: ResourceDevice,
			Action:       ActionControl,
			Description:  "设备控制操作",
		},
		{
			Code:         "alarm:read",
			Name:         "查看告警",
			ResourceType: ResourceAlarm,
			Action:       ActionRead,
			Description:  "查看告警信息",
		},
		{
			Code:         "alarm:ack",
			Name:         "告警确认",
			ResourceType: ResourceAlarm,
			Action:       ActionAck,
			Description:  "确认告警",
		},
		{
			Code:         "user:create",
			Name:         "创建用户",
			ResourceType: ResourceUser,
			Action:       ActionCreate,
			Description:  "创建新用户",
		},
		{
			Code:         "user:read",
			Name:         "查看用户",
			ResourceType: ResourceUser,
			Action:       ActionRead,
			Description:  "查看用户信息",
		},
		{
			Code:         "user:update",
			Name:         "更新用户",
			ResourceType: ResourceUser,
			Action:       ActionUpdate,
			Description:  "更新用户信息",
		},
		{
			Code:         "user:delete",
			Name:         "删除用户",
			ResourceType: ResourceUser,
			Action:       ActionDelete,
			Description:  "删除用户",
		},
		{
			Code:         "role:create",
			Name:         "创建角色",
			ResourceType: ResourceRole,
			Action:       ActionCreate,
			Description:  "创建新角色",
		},
		{
			Code:         "role:read",
			Name:         "查看角色",
			ResourceType: ResourceRole,
			Action:       ActionRead,
			Description:  "查看角色信息",
		},
		{
			Code:         "role:update",
			Name:         "更新角色",
			ResourceType: ResourceRole,
			Action:       ActionUpdate,
			Description:  "更新角色信息",
		},
		{
			Code:         "role:delete",
			Name:         "删除角色",
			ResourceType: ResourceRole,
			Action:       ActionDelete,
			Description:  "删除角色",
		},
	}
}
