package entity

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOperationLog(t *testing.T) {
	log := NewOperationLog("user-001", "admin", ActionLogin)
	assert.Empty(t, log.ID)
	assert.Equal(t, "user-001", log.UserID)
	assert.Equal(t, "admin", log.Username)
	assert.Equal(t, ActionLogin, log.Action)
}

func TestOperationLog_SetResource(t *testing.T) {
	log := NewOperationLog("user-001", "admin", ActionLogin)
	log.SetResource(ResourceSystemConfig, "config-001")
	assert.Equal(t, ResourceSystemConfig, log.ResourceType)
	assert.Equal(t, "config-001", log.ResourceID)
}

func TestOperationLog_SetDetails(t *testing.T) {
	log := NewOperationLog("user-001", "admin", ActionLogin)
	log.SetDetails(Details{"ip": "192.168.1.1", "action": "login"})
	assert.NotNil(t, log.Details)
	assert.Equal(t, "192.168.1.1", log.Details["ip"])
}

func TestOperationLog_SetRequestInfo(t *testing.T) {
	log := NewOperationLog("user-001", "admin", ActionLogin)
	log.SetRequestInfo("192.168.1.1", "Mozilla/5.0")
	assert.Equal(t, "192.168.1.1", log.IPAddress)
	assert.Equal(t, "Mozilla/5.0", log.UserAgent)
}

func TestOperationLog_TableName(t *testing.T) {
	log := OperationLog{}
	assert.Equal(t, "operation_logs", log.TableName())
}

func TestDetails_Value(t *testing.T) {
	d := Details{"key": "value"}
	v, err := d.Value()
	assert.NoError(t, err)
	assert.NotNil(t, v)

	var parsed map[string]interface{}
	err = json.Unmarshal(v.([]byte), &parsed)
	assert.NoError(t, err)
	assert.Equal(t, "value", parsed["key"])
}

func TestDetails_Value_Nil(t *testing.T) {
	var d Details
	v, err := d.Value()
	assert.NoError(t, err)
	assert.Nil(t, v)
}

func TestDetails_Scan(t *testing.T) {
	jsonData := `{"key":"value"}`
	d := Details{}
	err := d.Scan([]byte(jsonData))
	assert.NoError(t, err)
	assert.Equal(t, "value", d["key"])
}

func TestDetails_Scan_Nil(t *testing.T) {
	d := Details{"existing": "data"}
	err := d.Scan(nil)
	assert.NoError(t, err)
	assert.Nil(t, d)
}

func TestDetails_Scan_InvalidType(t *testing.T) {
	d := Details{}
	err := d.Scan(12345)
	assert.NoError(t, err)
}

func TestOperationAction_Constants(t *testing.T) {
	assert.Equal(t, "login", ActionLogin)
	assert.Equal(t, "logout", ActionLogout)
	assert.Equal(t, "create_user", ActionCreateUser)
	assert.Equal(t, "update_user", ActionUpdateUser)
	assert.Equal(t, "delete_user", ActionDeleteUser)
	assert.Equal(t, "change_password", ActionChangePassword)
	assert.Equal(t, "assign_role", ActionAssignRole)
	assert.Equal(t, "remove_role", ActionRemoveRole)
	assert.Equal(t, "create_role", ActionCreateRole)
	assert.Equal(t, "update_role", ActionUpdateRole)
	assert.Equal(t, "delete_role", ActionDeleteRole)
	assert.Equal(t, "assign_permission", ActionAssignPermission)
	assert.Equal(t, "remove_permission", ActionRemovePermission)
}

func TestResource_Constants_OperationLog(t *testing.T) {
	assert.Equal(t, "system_config", ResourceSystemConfig)
	assert.Equal(t, "alarm_rule", ResourceAlarmRule)
}
