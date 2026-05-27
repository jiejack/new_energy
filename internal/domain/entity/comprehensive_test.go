package entity

import (
	"testing"
	"time"

	"github.com/new-energy-monitoring/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestConfigItem_CreationAndMethods(t *testing.T) {
	item := NewConfigItem("key1", "value1", "prod", "default", "basic")
	assert.NotEmpty(t, item.ID)
	assert.Equal(t, "key1", item.Key)
	assert.Equal(t, "value1", item.Value)
	assert.Equal(t, config.ValueTypeString, item.ValueType)
	assert.Equal(t, "prod", item.Env)
	assert.True(t, item.Enabled)

	item.SetValue("new_value", config.ValueTypeInt)
	assert.Equal(t, "new_value", item.Value)
	assert.Equal(t, config.ValueTypeInt, item.ValueType)

	item.SetEncrypted(true)
	assert.True(t, item.Encrypted)

	assert.True(t, item.IsEnabled())
	item.Disable()
	assert.False(t, item.IsEnabled())
	item.Enable()
	assert.True(t, item.IsEnabled())
}

func TestConfigItem_CreateVersionAndRelease(t *testing.T) {
	item := NewConfigItem("key1", "value1", "prod", "default", "basic")

	version := item.CreateVersion(1, "initial", "admin")
	assert.NotEmpty(t, version.ID)
	assert.Equal(t, item.ID, version.ConfigID)
	assert.Equal(t, 1, version.Version)
	assert.Equal(t, "value1", version.Value)

	release := item.CreateRelease(1, "prod", config.ReleaseTypeFull, nil, "admin")
	assert.NotEmpty(t, release.ID)
	assert.Equal(t, item.ID, release.ConfigID)
	assert.Equal(t, config.ReleaseStatusPending, release.Status)

	audit := item.CreateAudit(config.AuditActionCreate, "", "value1", "admin", "127.0.0.1")
	assert.NotEmpty(t, audit.ID)
	assert.Equal(t, item.ID, audit.ConfigID)
	assert.Equal(t, config.AuditActionCreate, audit.Action)
}

func TestConfigRelease_AllStatusAndTypeMethods(t *testing.T) {
	r := NewConfigRelease("cfg-001", 1, "prod", config.ReleaseTypeFull, nil, "admin")
	assert.True(t, r.IsPending())
	assert.False(t, r.IsReleasing())
	assert.False(t, r.IsSuccess())
	assert.False(t, r.IsFailed())
	assert.True(t, r.IsFullRelease())
	assert.False(t, r.IsGrayRelease())
	assert.False(t, r.IsRollback())

	r.Start()
	assert.True(t, r.IsReleasing())

	r.Complete()
	assert.True(t, r.IsSuccess())
	assert.NotNil(t, r.ReleasedAt)

	r2 := NewConfigRelease("cfg-001", 1, "prod", config.ReleaseTypeFull, nil, "admin")
	r2.Fail()
	assert.True(t, r2.IsFailed())

	gray := NewConfigRelease("cfg-001", 1, "prod", config.ReleaseTypeGray, nil, "admin")
	assert.True(t, gray.IsGrayRelease())

	rollback := NewConfigRelease("cfg-001", 1, "prod", config.ReleaseTypeRollback, nil, "admin")
	assert.True(t, rollback.IsRollback())
}

func TestConfigAudit_AllActionMethods(t *testing.T) {
	create := NewConfigAudit("cfg-001", config.AuditActionCreate, "", "val", "admin", "127.0.0.1")
	assert.True(t, create.IsCreate())
	assert.False(t, create.IsUpdate())

	update := NewConfigAudit("cfg-001", config.AuditActionUpdate, "old", "new", "admin", "127.0.0.1")
	assert.True(t, update.IsUpdate())

	delete := NewConfigAudit("cfg-001", config.AuditActionDelete, "val", "", "admin", "127.0.0.1")
	assert.True(t, delete.IsDelete())

	rollback := NewConfigAudit("cfg-001", config.AuditActionRollback, "new", "old", "admin", "127.0.0.1")
	assert.True(t, rollback.IsRollback())

	release := NewConfigAudit("cfg-001", config.AuditActionRelease, "val", "val", "admin", "127.0.0.1")
	assert.True(t, release.IsRelease())
}

func TestNewConfigVersion_Standalone(t *testing.T) {
	v := NewConfigVersion("cfg-001", 2, "new_value", "update", "admin")
	assert.NotEmpty(t, v.ID)
	assert.Equal(t, "cfg-001", v.ConfigID)
	assert.Equal(t, 2, v.Version)
}

func TestDevice_AllStatusAndCommMethods(t *testing.T) {
	d := NewDevice("INV_001", "Inverter 1", DeviceTypeInverter, "station-001")
	assert.Equal(t, "INV_001", d.Code)
	assert.Equal(t, DeviceTypeInverter, d.Type)
	assert.Equal(t, DeviceStatusOffline, d.Status)
	assert.False(t, d.IsOnline())

	d.SetOnline()
	assert.True(t, d.IsOnline())
	assert.Equal(t, DeviceStatusOnline, d.Status)
	assert.NotNil(t, d.LastOnline)

	d.SetOffline()
	assert.Equal(t, DeviceStatusOffline, d.Status)

	d.SetFault()
	assert.Equal(t, DeviceStatusFault, d.Status)

	d.SetMaintain()
	assert.Equal(t, DeviceStatusMaintain, d.Status)

	d.SetCommunication("modbus", "192.168.1.1", 502, 1)
	assert.Equal(t, "modbus", d.Protocol)
	assert.Equal(t, "192.168.1.1", d.IPAddress)
	assert.Equal(t, 502, d.Port)
	assert.Equal(t, 1, d.SlaveID)
}

func TestQASession_AllMethods(t *testing.T) {
	s := NewQASession("user-001", "Test Session")
	assert.NotEmpty(t, s.ID)
	assert.Equal(t, "user-001", s.UserID)
	assert.Equal(t, QASessionStatusActive, s.Status)
	assert.True(t, s.IsActive())

	msg := s.AddMessage(QAMessageRoleUser, "Hello")
	assert.NotEmpty(t, msg.ID)
	assert.Equal(t, s.ID, msg.SessionID)
	assert.Equal(t, QAMessageRoleUser, msg.Role)

	s.Archive()
	assert.True(t, s.IsArchived())
	assert.False(t, s.IsActive())

	s2 := NewQASession("user-001", "Test Session 2")
	s2.Delete()
	assert.True(t, s2.IsDeleted())
}

func TestQAMessage_AllRoleMethods(t *testing.T) {
	userMsg := NewQAMessage("s-001", QAMessageRoleUser, "Hello")
	assert.True(t, userMsg.IsUserMessage())
	assert.False(t, userMsg.IsAssistantMessage())

	assistantMsg := NewQAMessage("s-001", QAMessageRoleAssistant, "Hi there")
	assert.True(t, assistantMsg.IsAssistantMessage())

	systemMsg := NewQAMessage("s-001", QAMessageRoleSystem, "System message")
	assert.True(t, systemMsg.IsSystemMessage())
}

func TestUser_AllSetterMethods(t *testing.T) {
	u := NewUser("admin", "hash123")
	assert.Equal(t, "admin", u.Username)
	assert.Equal(t, "hash123", u.PasswordHash)
	assert.Equal(t, UserStatusActive, u.Status)
	assert.True(t, u.IsActive())

	u.SetEmail("admin@test.com")
	assert.Equal(t, "admin@test.com", u.Email)

	u.SetPhone("1234567890")
	assert.Equal(t, "1234567890", u.Phone)

	u.SetRealName("Admin User")
	assert.Equal(t, "Admin User", u.RealName)

	u.SetAvatar("avatar.png")
	assert.Equal(t, "avatar.png", u.Avatar)

	u.Deactivate()
	assert.False(t, u.IsActive())
	u.Activate()
	assert.True(t, u.IsActive())

	u.UpdateLastLogin()
	assert.NotNil(t, u.LastLogin)
	assert.Equal(t, 1, u.LoginCount)
	u.UpdateLastLogin()
	assert.Equal(t, 2, u.LoginCount)

	u.UpdatePassword("new_hash")
	assert.Equal(t, "new_hash", u.PasswordHash)
}

func TestFaultDetectionResult_CreationAndMethods(t *testing.T) {
	r := NewFaultDetectionResult("device-001", "temperature", FaultSeverityWarning, 0.9, "test", "1.0")
	assert.Equal(t, "device-001", r.DeviceID)
	assert.Equal(t, FaultSeverityWarning, r.Severity)
	assert.Equal(t, FaultDetectionStatusPending, r.Status)
}

func TestAlarm_CreationAndMethods(t *testing.T) {
	a := NewAlarm("p1", "d1", "s1", AlarmTypeLimit, AlarmLevelWarning, "Test Alarm", "msg")
	assert.Equal(t, AlarmLevelWarning, a.Level)
	assert.Equal(t, AlarmStatusActive, a.Status)

	a.Acknowledge("admin")
	assert.Equal(t, AlarmStatusAcknowledged, a.Status)
	assert.Equal(t, "admin", a.AcknowledgedBy)
	assert.NotNil(t, a.AcknowledgedAt)

	a.Clear()
	assert.Equal(t, AlarmStatusCleared, a.Status)
	assert.NotNil(t, a.ClearedAt)
}

func TestSystemConfig_CreationAndMethods(t *testing.T) {
	c := NewSystemConfig("basic", "system_name", "MyApp", SystemConfigValueTypeString, "App name")
	assert.NotEmpty(t, c.ID)
	assert.Equal(t, "basic", c.Category)
	assert.Equal(t, "system_name", c.Key)
	assert.Equal(t, "MyApp", c.Value)

	c.SetValue("NewApp", SystemConfigValueTypeString)
	assert.Equal(t, "NewApp", c.Value)
}

func TestOtherEntityConstructors(t *testing.T) {
	r := NewRole("admin", "Administrator")
	assert.Equal(t, "admin", r.Code)

	p := NewPermission("alarm:read", "Read Alarms")
	assert.Equal(t, "alarm:read", p.Code)

	log := NewOperationLog("user-001", "admin", "create")
	assert.Equal(t, "user-001", log.UserID)

	s := NewStation("SP001", "Solar Plant 1", StationTypePV, "region-001")
	assert.Equal(t, "Solar Plant 1", s.Name)

	pt := NewPoint("P_001", "Power", PointTypeYaoCe)
	assert.Equal(t, "P_001", pt.Code)

	region := NewRegion("EC", "East China", nil, 1)
	assert.Equal(t, "East China", region.Name)

	fr := NewForecastResult("station-001", ForecastTypeShortTerm, time.Now(), 500.0, "1.0")
	assert.Equal(t, "station-001", fr.StationID)

	mv := &ModelVersion{ModelName: "solar_v2", Version: "1.0.0", Status: ModelStatusStaging}
	assert.Equal(t, "solar_v2", mv.ModelName)

	en := &EdgeNode{Name: "Edge1", StationID: "station-001", IPAddress: "192.168.1.1", Status: EdgeNodeOffline}
	assert.Equal(t, "Edge1", en.Name)

	nc := &NotificationConfig{Name: "Test Config", Type: "email", Enabled: true}
	assert.Equal(t, "Test Config", nc.Name)

	ar := NewAlarmRule("High Temp", AlarmRuleTypeLimit, AlarmLevelCritical, "value > 80")
	assert.Equal(t, "High Temp", ar.Name)
}
