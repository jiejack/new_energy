package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAlarmRule(t *testing.T) {
	rule := NewAlarmRule("Test Rule", AlarmRuleTypeLimit, AlarmLevelWarning, "value > threshold")
	assert.Equal(t, "Test Rule", rule.Name)
	assert.Equal(t, AlarmRuleTypeLimit, rule.Type)
	assert.Equal(t, AlarmLevelWarning, rule.Level)
	assert.Equal(t, "value > threshold", rule.Condition)
	assert.Equal(t, AlarmRuleStatusEnabled, rule.Status)
	assert.NotZero(t, rule.CreatedAt)
	assert.NotZero(t, rule.UpdatedAt)
}

func TestAlarmRule_TableName(t *testing.T) {
	rule := AlarmRule{}
	assert.Equal(t, "alarm_rules", rule.TableName())
}

func TestAlarmRuleStatus_Constants(t *testing.T) {
	assert.Equal(t, AlarmRuleStatus(0), AlarmRuleStatusDisabled)
	assert.Equal(t, AlarmRuleStatus(1), AlarmRuleStatusEnabled)
}

func TestAlarmRuleType_Constants(t *testing.T) {
	assert.Equal(t, AlarmRuleType("limit"), AlarmRuleTypeLimit)
	assert.Equal(t, AlarmRuleType("trend"), AlarmRuleTypeTrend)
	assert.Equal(t, AlarmRuleType("custom"), AlarmRuleTypeCustom)
}

func TestNewAlarmRule_WithPointID(t *testing.T) {
	pointID := "point-001"
	rule := NewAlarmRule("Point Rule", AlarmRuleTypeLimit, AlarmLevelWarning, "value > threshold")
	rule.PointID = &pointID
	assert.NotNil(t, rule.PointID)
	assert.Equal(t, "point-001", *rule.PointID)
}

func TestNewAlarmRule_WithDeviceID(t *testing.T) {
	deviceID := "device-001"
	rule := NewAlarmRule("Device Rule", AlarmRuleTypeLimit, AlarmLevelWarning, "value > threshold")
	rule.DeviceID = &deviceID
	assert.NotNil(t, rule.DeviceID)
	assert.Equal(t, "device-001", *rule.DeviceID)
}

func TestNewAlarmRule_WithStationID(t *testing.T) {
	stationID := "station-001"
	rule := NewAlarmRule("Station Rule", AlarmRuleTypeLimit, AlarmLevelWarning, "value > threshold")
	rule.StationID = &stationID
	assert.NotNil(t, rule.StationID)
	assert.Equal(t, "station-001", *rule.StationID)
}

func TestNewAlarmRule_WithThreshold(t *testing.T) {
	rule := NewAlarmRule("Threshold Rule", AlarmRuleTypeLimit, AlarmLevelWarning, "value > threshold")
	rule.Threshold = 85.0
	rule.Duration = 60
	assert.Equal(t, 85.0, rule.Threshold)
	assert.Equal(t, 60, rule.Duration)
}

func TestNewAlarmRule_WithNotifyChannels(t *testing.T) {
	rule := NewAlarmRule("Notify Rule", AlarmRuleTypeLimit, AlarmLevelWarning, "value > threshold")
	rule.NotifyChannels = []string{"email", "sms"}
	rule.NotifyUsers = []string{"user-001", "user-002"}
	assert.Equal(t, []string{"email", "sms"}, rule.NotifyChannels)
	assert.Equal(t, []string{"user-001", "user-002"}, rule.NotifyUsers)
}

func TestNewAlarmRule_Disabled(t *testing.T) {
	rule := NewAlarmRule("Disabled Rule", AlarmRuleTypeLimit, AlarmLevelWarning, "value > threshold")
	rule.Status = AlarmRuleStatusDisabled
	assert.Equal(t, AlarmRuleStatusDisabled, rule.Status)
}

func TestNewAlarmRule_AllLevels(t *testing.T) {
	levels := []AlarmLevel{AlarmLevelInfo, AlarmLevelWarning, AlarmLevelMajor, AlarmLevelCritical}
	types := []AlarmRuleType{AlarmRuleTypeLimit, AlarmRuleTypeTrend, AlarmRuleTypeCustom}

	for _, level := range levels {
		for _, ruleType := range types {
			rule := NewAlarmRule("Rule", ruleType, level, "condition")
			assert.Equal(t, level, rule.Level)
			assert.Equal(t, ruleType, rule.Type)
		}
	}
}
