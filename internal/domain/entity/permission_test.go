package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPermission(t *testing.T) {
	tests := []struct {
		name string
		code string
		name string
		want *Permission
	}{
		{
			name: "创建用户查看权限",
			code: "user:read",
			name: "查看用户",
			want: &Permission{
				Code: "user:read",
				Name: "查看用户",
			},
		},
		{
			name: "创建设备控制权限",
			code: "device:control",
			name: "设备控制",
			want: &Permission{
				Code: "device:control",
				Name: "设备控制",
			},
		},
		{
			name: "创建告警确认权限",
			code: "alarm:ack",
			name: "告警确认",
			want: &Permission{
				Code: "alarm:ack",
				Name: "告警确认",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPermission(tt.code, tt.name)
			assert.Equal(t, tt.want.Code, got.Code)
			assert.Equal(t, tt.want.Name, got.Name)
		})
	}
}

func TestPermission_SetResource(t *testing.T) {
	perm := NewPermission("station:read", "查看厂站")
	assert.Empty(t, perm.ResourceType)
	assert.Empty(t, perm.ResourceID)

	perm.SetResource(ResourceStation, "station-001")
	assert.Equal(t, ResourceStation, perm.ResourceType)
	assert.Equal(t, "station-001", perm.ResourceID)
}

func TestPermission_SetAction(t *testing.T) {
	perm := NewPermission("device:control", "设备控制")
	assert.Empty(t, perm.Action)

	perm.SetAction(ActionControl)
	assert.Equal(t, ActionControl, perm.Action)
}

func TestPermission_SetDescription(t *testing.T) {
	perm := NewPermission("user:create", "创建用户")
	assert.Empty(t, perm.Description)

	perm.SetDescription("创建新用户的权限")
	assert.Equal(t, "创建新用户的权限", perm.Description)
}

func TestPermission_TableName(t *testing.T) {
	perm := Permission{}
	assert.Equal(t, "permissions", perm.TableName())
}

func TestAction_Constants(t *testing.T) {
	assert.Equal(t, "create", ActionCreate)
	assert.Equal(t, "read", ActionRead)
	assert.Equal(t, "update", ActionUpdate)
	assert.Equal(t, "delete", ActionDelete)
	assert.Equal(t, "control", ActionControl)
	assert.Equal(t, "ack", ActionAck)
}

func TestResource_Constants(t *testing.T) {
	assert.Equal(t, "station", ResourceStation)
	assert.Equal(t, "device", ResourceDevice)
	assert.Equal(t, "point", ResourcePoint)
	assert.Equal(t, "alarm", ResourceAlarm)
	assert.Equal(t, "user", ResourceUser)
	assert.Equal(t, "role", ResourceRole)
	assert.Equal(t, "region", ResourceRegion)
}

func TestGetDefaultPermissions(t *testing.T) {
	perms := GetDefaultPermissions()
	assert.NotEmpty(t, perms, "Default permissions should not be empty")

	// 验证默认权限包含必要的权限
	permCodes := make(map[string]bool)
	for _, perm := range perms {
		permCodes[perm.Code] = true
		assert.NotEmpty(t, perm.Code)
		assert.NotEmpty(t, perm.Name)
		assert.NotEmpty(t, perm.ResourceType)
		assert.NotEmpty(t, perm.Action)
	}

	// 验证关键权限存在
	assert.True(t, permCodes["station:create"], "station:create permission should exist")
	assert.True(t, permCodes["station:read"], "station:read permission should exist")
	assert.True(t, permCodes["device:create"], "device:create permission should exist")
	assert.True(t, permCodes["device:control"], "device:control permission should exist")
	assert.True(t, permCodes["alarm:read"], "alarm:read permission should exist")
	assert.True(t, permCodes["alarm:ack"], "alarm:ack permission should exist")
	assert.True(t, permCodes["user:create"], "user:create permission should exist")
	assert.True(t, permCodes["role:create"], "role:create permission should exist")
}

func TestPermission_ResourceTypes(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		action       string
	}{
		{"厂站创建权限", ResourceStation, ActionCreate},
		{"设备读取权限", ResourceDevice, ActionRead},
		{"告警确认权限", ResourceAlarm, ActionAck},
		{"用户更新权限", ResourceUser, ActionUpdate},
		{"角色删除权限", ResourceRole, ActionDelete},
		{"区域读取权限", ResourceRegion, ActionRead},
		{"采集点读取权限", ResourcePoint, ActionRead},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perm := NewPermission(tt.resourceType+":"+tt.action, tt.name)
			perm.SetResource(tt.resourceType, "")
			perm.SetAction(tt.action)

			assert.Equal(t, tt.resourceType, perm.ResourceType)
			assert.Equal(t, tt.action, perm.Action)
		})
	}
}

func TestPermission_FullPermission(t *testing.T) {
	perm := NewPermission("station:read", "查看厂站")
	perm.SetResource(ResourceStation, "station-001")
	perm.SetAction(ActionRead)
	perm.SetDescription("查看指定厂站的详细信息")

	assert.Equal(t, "station:read", perm.Code)
	assert.Equal(t, "查看厂站", perm.Name)
	assert.Equal(t, ResourceStation, perm.ResourceType)
	assert.Equal(t, "station-001", perm.ResourceID)
	assert.Equal(t, ActionRead, perm.Action)
	assert.Equal(t, "查看指定厂站的详细信息", perm.Description)
}
