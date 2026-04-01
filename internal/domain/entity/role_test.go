package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRole(t *testing.T) {
	tests := []struct {
		name string
		code string
		roleName string
		want *Role
	}{
		{
			name: "创建管理员角色",
			code: "admin",
			roleName: "系统管理员",
			want: &Role{
				Code: "admin",
				Name: "系统管理员",
			},
		},
		{
			name: "创建操作员角色",
			code: "operator",
			roleName: "运维人员",
			want: &Role{
				Code: "operator",
				Name: "运维人员",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRole(tt.code, tt.roleName)
			assert.Equal(t, tt.want.Code, got.Code)
			assert.Equal(t, tt.want.Name, got.Name)
		})
	}
}

func TestRole_SetDescription(t *testing.T) {
	role := NewRole("admin", "管理员")
	assert.Empty(t, role.Description)

	role.SetDescription("系统最高权限管理员")
	assert.Equal(t, "系统最高权限管理员", role.Description)
}

func TestRole_SetAsSystemRole(t *testing.T) {
	role := NewRole("admin", "管理员")
	assert.False(t, role.IsSystem)

	role.SetAsSystemRole()
	assert.True(t, role.IsSystem)
}

func TestRole_IsSystemRole(t *testing.T) {
	role := NewRole("admin", "管理员")
	assert.False(t, role.IsSystemRole())

	role.SetAsSystemRole()
	assert.True(t, role.IsSystemRole())
}

func TestRole_TableName(t *testing.T) {
	role := Role{}
	assert.Equal(t, "roles", role.TableName())
}

func TestRolePermission_TableName(t *testing.T) {
	rp := RolePermission{}
	assert.Equal(t, "role_permissions", rp.TableName())
}

func TestGetDefaultRoles(t *testing.T) {
	roles := GetDefaultRoles()
	assert.Len(t, roles, 4)

	// 验证默认角色代码
	roleCodes := make(map[string]bool)
	for _, role := range roles {
		roleCodes[role.Code] = true
		assert.True(t, role.IsSystem)
	}

	assert.True(t, roleCodes[RoleCodeSuperAdmin])
	assert.True(t, roleCodes[RoleCodeAdmin])
	assert.True(t, roleCodes[RoleCodeOperator])
	assert.True(t, roleCodes[RoleCodeViewer])
}

func TestRoleCode_Constants(t *testing.T) {
	assert.Equal(t, "super_admin", RoleCodeSuperAdmin)
	assert.Equal(t, "admin", RoleCodeAdmin)
	assert.Equal(t, "operator", RoleCodeOperator)
	assert.Equal(t, "viewer", RoleCodeViewer)
}
