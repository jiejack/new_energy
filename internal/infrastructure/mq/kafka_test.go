package mq

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopicConstants(t *testing.T) {
	assert.Equal(t, "data.collect", TopicDataCollect)
	assert.Equal(t, "alarm.event", TopicAlarmEvent)
	assert.Equal(t, "alarm.notify", TopicAlarmNotify)
	assert.Equal(t, "device.status", TopicDeviceStatus)
	assert.Equal(t, "compute.result", TopicComputeResult)
}

func TestCollectDataMessage(t *testing.T) {
	msg := CollectDataMessage{
		PointID:   "pt1",
		Value:     42.5,
		Quality:   1,
		Timestamp: 1700000000,
		Source:    "modbus",
	}
	assert.Equal(t, "pt1", msg.PointID)
	assert.Equal(t, 42.5, msg.Value)

	data, err := json.Marshal(msg)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "pt1")

	var parsed CollectDataMessage
	err = json.Unmarshal(data, &parsed)
	assert.NoError(t, err)
	assert.Equal(t, "pt1", parsed.PointID)
	assert.Equal(t, 42.5, parsed.Value)
}

func TestAlarmEventMessage(t *testing.T) {
	msg := AlarmEventMessage{
		AlarmID:   "alarm1",
		PointID:   "pt1",
		DeviceID:  "dev1",
		StationID: "st1",
		Level:     3,
		Type:      "limit",
		Title:     "High Temperature",
		Message:   "High temperature",
		Value:     95.5,
		Threshold: 90.0,
		Timestamp: 1700000000,
	}
	assert.Equal(t, "alarm1", msg.AlarmID)
	assert.Equal(t, 3, msg.Level)

	data, err := json.Marshal(msg)
	assert.NoError(t, err)

	var parsed AlarmEventMessage
	err = json.Unmarshal(data, &parsed)
	assert.NoError(t, err)
	assert.Equal(t, "alarm1", parsed.AlarmID)
}

func TestDeviceStatusMessage(t *testing.T) {
	msg := DeviceStatusMessage{
		DeviceID:  "dev1",
		StationID: "st1",
		Status:    1,
		Timestamp: 1700000000,
	}
	assert.Equal(t, "dev1", msg.DeviceID)
	assert.Equal(t, 1, msg.Status)

	data, err := json.Marshal(msg)
	assert.NoError(t, err)

	var parsed DeviceStatusMessage
	err = json.Unmarshal(data, &parsed)
	assert.NoError(t, err)
	assert.Equal(t, 1, parsed.Status)
}
