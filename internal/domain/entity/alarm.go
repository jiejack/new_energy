package entity

import (
	"time"
)

type AlarmLevel int
type AlarmType string
type AlarmStatus int

const (
	AlarmLevelInfo     AlarmLevel = 1
	AlarmLevelWarning  AlarmLevel = 2
	AlarmLevelMajor    AlarmLevel = 3
	AlarmLevelCritical AlarmLevel = 4
)

const (
	AlarmTypeLimit     AlarmType = "limit"
	AlarmTypeStatus    AlarmType = "status"
	AlarmTypeComm      AlarmType = "comm"
	AlarmTypeSystem    AlarmType = "system"
	AlarmTypeDevice    AlarmType = "device"
)

const (
	AlarmStatusActive       AlarmStatus = 1
	AlarmStatusAcknowledged AlarmStatus = 2
	AlarmStatusCleared      AlarmStatus = 3
	AlarmStatusSuppressed   AlarmStatus = 4
)

type Alarm struct {
	ID          string       `json:"id" gorm:"primaryKey;type:varchar(36)"`
	PointID     string       `json:"point_id" gorm:"type:varchar(36);index"`
	DeviceID    string       `json:"device_id" gorm:"type:varchar(36);index"`
	StationID   string       `json:"station_id" gorm:"type:varchar(36);index"`
	
	Type        AlarmType    `json:"type" gorm:"type:varchar(20);not null"`
	Level       AlarmLevel   `json:"level" gorm:"not null"`
	
	Title       string       `json:"title" gorm:"type:varchar(200);not null"`
	Message     string       `json:"message" gorm:"type:text"`
	
	Value       float64      `json:"value"`
	Threshold   float64      `json:"threshold"`
	
	Status      AlarmStatus  `json:"status" gorm:"default:1"`
	
	TriggeredAt time.Time    `json:"triggered_at" gorm:"index"`
	AcknowledgedAt *time.Time `json:"acknowledged_at"`
	ClearedAt   *time.Time   `json:"cleared_at"`
	AcknowledgedBy string    `json:"acknowledged_by" gorm:"type:varchar(100)"`
	
	CreatedAt   time.Time    `json:"created_at"`
}

func (a *Alarm) TableName() string {
	return "alarms"
}

func NewAlarm(pointID, deviceID, stationID string, alarmType AlarmType, level AlarmLevel, title, message string) *Alarm {
	return &Alarm{
		PointID:     pointID,
		DeviceID:    deviceID,
		StationID:   stationID,
		Type:        alarmType,
		Level:       level,
		Title:       title,
		Message:     message,
		Status:      AlarmStatusActive,
		TriggeredAt: time.Now(),
		CreatedAt:   time.Now(),
	}
}

func (a *Alarm) Acknowledge(by string) {
	now := time.Now()
	a.Status = AlarmStatusAcknowledged
	a.AcknowledgedAt = &now
	a.AcknowledgedBy = by
}

func (a *Alarm) Clear() {
	now := time.Now()
	a.Status = AlarmStatusCleared
	a.ClearedAt = &now
}

func (a *Alarm) Suppress() {
	a.Status = AlarmStatusSuppressed
}

func (a *Alarm) IsActive() bool {
	return a.Status == AlarmStatusActive
}

func (a *Alarm) IsAcknowledged() bool {
	return a.Status == AlarmStatusAcknowledged
}

func (a *Alarm) IsCleared() bool {
	return a.Status == AlarmStatusCleared
}

func (a *Alarm) Duration() time.Duration {
	if a.ClearedAt != nil {
		return a.ClearedAt.Sub(a.TriggeredAt)
	}
	return time.Since(a.TriggeredAt)
}
