package entity

import (
	"time"
)

type AlarmRuleStatus int
type AlarmRuleType string

const (
	AlarmRuleStatusDisabled AlarmRuleStatus = 0
	AlarmRuleStatusEnabled  AlarmRuleStatus = 1
)

const (
	AlarmRuleTypeLimit  AlarmRuleType = "limit"
	AlarmRuleTypeTrend  AlarmRuleType = "trend"
	AlarmRuleTypeCustom AlarmRuleType = "custom"
)

type AlarmRule struct {
	ID          string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Name        string          `json:"name" gorm:"type:varchar(100);not null;uniqueIndex"`
	Description string          `json:"description" gorm:"type:text"`

	PointID     *string         `json:"point_id" gorm:"type:varchar(36);index"`
	DeviceID    *string         `json:"device_id" gorm:"type:varchar(36);index"`
	StationID   *string         `json:"station_id" gorm:"type:varchar(36);index"`

	Type        AlarmRuleType   `json:"type" gorm:"type:varchar(20);not null"`
	Level       AlarmLevel      `json:"level" gorm:"not null"`

	Condition   string          `json:"condition" gorm:"type:text;not null"`
	Threshold   float64         `json:"threshold"`
	Duration    int             `json:"duration" gorm:"default:0"`

	NotifyChannels []string     `json:"notify_channels" gorm:"type:text;serializer:json"`
	NotifyUsers    []string     `json:"notify_users" gorm:"type:text;serializer:json"`

	Status      AlarmRuleStatus `json:"status" gorm:"default:1"`

	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	CreatedBy   string          `json:"created_by" gorm:"type:varchar(100)"`
	UpdatedBy   string          `json:"updated_by" gorm:"type:varchar(100)"`
}

func (r *AlarmRule) TableName() string {
	return "alarm_rules"
}

func NewAlarmRule(name string, ruleType AlarmRuleType, level AlarmLevel, condition string) *AlarmRule {
	return &AlarmRule{
		Name:      name,
		Type:      ruleType,
		Level:     level,
		Condition: condition,
		Status:    AlarmRuleStatusEnabled,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
