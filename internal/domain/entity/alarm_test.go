package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAlarm(t *testing.T) {
	tests := []struct {
		name       string
		pointID    string
		deviceID   string
		stationID  string
		alarmType  AlarmType
		level      AlarmLevel
		title      string
		message    string
		want       *Alarm
	}{
		{
			name:      "创建高限告警",
			pointID:   "point001",
			deviceID:  "device001",
			stationID: "station001",
			alarmType: AlarmTypeLimit,
			level:     AlarmLevelWarning,
			title:     "电压高限告警",
			message:   "电压超过上限阈值",
			want: &Alarm{
				PointID:     "point001",
				DeviceID:    "device001",
				StationID:   "station001",
				Type:        AlarmTypeLimit,
				Level:       AlarmLevelWarning,
				Title:       "电压高限告警",
				Message:     "电压超过上限阈值",
				Status:      AlarmStatusActive,
			},
		},
		{
			name:      "创建设备故障告警",
			pointID:   "point002",
			deviceID:  "device002",
			stationID: "station001",
			alarmType: AlarmTypeDevice,
			level:     AlarmLevelCritical,
			title:     "设备故障",
			message:   "逆变器通信中断",
			want: &Alarm{
				PointID:     "point002",
				DeviceID:    "device002",
				StationID:   "station001",
				Type:        AlarmTypeDevice,
				Level:       AlarmLevelCritical,
				Title:       "设备故障",
				Message:     "逆变器通信中断",
				Status:      AlarmStatusActive,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAlarm(tt.pointID, tt.deviceID, tt.stationID, tt.alarmType, tt.level, tt.title, tt.message)
			assert.Equal(t, tt.want.PointID, got.PointID)
			assert.Equal(t, tt.want.DeviceID, got.DeviceID)
			assert.Equal(t, tt.want.StationID, got.StationID)
			assert.Equal(t, tt.want.Type, got.Type)
			assert.Equal(t, tt.want.Level, got.Level)
			assert.Equal(t, tt.want.Title, got.Title)
			assert.Equal(t, tt.want.Message, got.Message)
			assert.Equal(t, tt.want.Status, got.Status)
			assert.NotZero(t, got.TriggeredAt)
			assert.NotZero(t, got.CreatedAt)
		})
	}
}

func TestAlarm_Acknowledge(t *testing.T) {
	alarm := NewAlarm("point001", "device001", "station001", AlarmTypeLimit, AlarmLevelWarning, "告警", "测试告警")
	assert.Equal(t, AlarmStatusActive, alarm.Status)
	assert.Nil(t, alarm.AcknowledgedAt)
	assert.Empty(t, alarm.AcknowledgedBy)

	alarm.Acknowledge("user001")
	assert.Equal(t, AlarmStatusAcknowledged, alarm.Status)
	assert.NotNil(t, alarm.AcknowledgedAt)
	assert.Equal(t, "user001", alarm.AcknowledgedBy)
}

func TestAlarm_Clear(t *testing.T) {
	alarm := NewAlarm("point001", "device001", "station001", AlarmTypeLimit, AlarmLevelWarning, "告警", "测试告警")
	assert.Equal(t, AlarmStatusActive, alarm.Status)
	assert.Nil(t, alarm.ClearedAt)

	alarm.Clear()
	assert.Equal(t, AlarmStatusCleared, alarm.Status)
	assert.NotNil(t, alarm.ClearedAt)
}

func TestAlarm_Suppress(t *testing.T) {
	alarm := NewAlarm("point001", "device001", "station001", AlarmTypeLimit, AlarmLevelWarning, "告警", "测试告警")

	alarm.Suppress()
	assert.Equal(t, AlarmStatusSuppressed, alarm.Status)
}

func TestAlarm_IsActive(t *testing.T) {
	alarm := NewAlarm("point001", "device001", "station001", AlarmTypeLimit, AlarmLevelWarning, "告警", "测试告警")

	assert.True(t, alarm.IsActive())

	alarm.Acknowledge("user001")
	assert.False(t, alarm.IsActive())

	alarm.Clear()
	assert.False(t, alarm.IsActive())
}

func TestAlarm_IsAcknowledged(t *testing.T) {
	alarm := NewAlarm("point001", "device001", "station001", AlarmTypeLimit, AlarmLevelWarning, "告警", "测试告警")

	assert.False(t, alarm.IsAcknowledged())

	alarm.Acknowledge("user001")
	assert.True(t, alarm.IsAcknowledged())

	alarm.Clear()
	assert.False(t, alarm.IsAcknowledged())
}

func TestAlarm_IsCleared(t *testing.T) {
	alarm := NewAlarm("point001", "device001", "station001", AlarmTypeLimit, AlarmLevelWarning, "告警", "测试告警")

	assert.False(t, alarm.IsCleared())

	alarm.Clear()
	assert.True(t, alarm.IsCleared())
}

func TestAlarm_Duration(t *testing.T) {
	alarm := NewAlarm("point001", "device001", "station001", AlarmTypeLimit, AlarmLevelWarning, "告警", "测试告警")

	// 未清除时，持续时间应该接近0
	duration := alarm.Duration()
	assert.Less(t, duration.Seconds(), float64(1))

	// 清除后
	time.Sleep(10 * time.Millisecond)
	alarm.Clear()
	duration = alarm.Duration()
	assert.GreaterOrEqual(t, duration.Milliseconds(), int64(10))
}

func TestAlarm_TableName(t *testing.T) {
	alarm := Alarm{}
	assert.Equal(t, "alarms", alarm.TableName())
}

func TestAlarmLevel_Constants(t *testing.T) {
	assert.Equal(t, AlarmLevel(1), AlarmLevelInfo)
	assert.Equal(t, AlarmLevel(2), AlarmLevelWarning)
	assert.Equal(t, AlarmLevel(3), AlarmLevelMajor)
	assert.Equal(t, AlarmLevel(4), AlarmLevelCritical)
}

func TestAlarmType_Constants(t *testing.T) {
	assert.Equal(t, AlarmType("limit"), AlarmTypeLimit)
	assert.Equal(t, AlarmType("status"), AlarmTypeStatus)
	assert.Equal(t, AlarmType("comm"), AlarmTypeComm)
	assert.Equal(t, AlarmType("system"), AlarmTypeSystem)
	assert.Equal(t, AlarmType("device"), AlarmTypeDevice)
}

func TestAlarmStatus_Constants(t *testing.T) {
	assert.Equal(t, AlarmStatus(1), AlarmStatusActive)
	assert.Equal(t, AlarmStatus(2), AlarmStatusAcknowledged)
	assert.Equal(t, AlarmStatus(3), AlarmStatusCleared)
	assert.Equal(t, AlarmStatus(4), AlarmStatusSuppressed)
}
