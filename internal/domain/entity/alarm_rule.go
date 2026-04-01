package entity

import (
	"time"

	"github.com/google/uuid"
)

// AlarmRuleType 告警规则类型
type AlarmRuleType string

const (
	AlarmRuleTypeLimit  AlarmRuleType = "limit"  // 限值告警
	AlarmRuleTypeTrend  AlarmRuleType = "trend"  // 趋势告警
	AlarmRuleTypeCustom AlarmRuleType = "custom" // 自定义告警
)

// AlarmRuleStatus 告警规则状态
type AlarmRuleStatus int

const (
	AlarmRuleStatusDisabled AlarmRuleStatus = 0 // 禁用
	AlarmRuleStatusEnabled  AlarmRuleStatus = 1 // 启用
)

// AlarmRule 告警规则实体
type AlarmRule struct {
	ID          string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Name        string         `json:"name" gorm:"type:varchar(200);not null;comment:规则名称"`
	Description string         `json:"description" gorm:"type:text;comment:规则描述"`

	// 关联对象（可选）
	PointID   *string `json:"point_id" gorm:"type:varchar(36);index;comment:采集点ID"`
	DeviceID  *string `json:"device_id" gorm:"type:varchar(36);index;comment:设备ID"`
	StationID *string `json:"station_id" gorm:"type:varchar(36);index;comment:厂站ID"`

	// 规则配置
	Type      AlarmRuleType   `json:"type" gorm:"type:varchar(20);not null;comment:规则类型"`
	Level     AlarmLevel      `json:"level" gorm:"not null;comment:告警级别"`
	Condition string          `json:"condition" gorm:"type:varchar(500);comment:触发条件"`
	Threshold float64         `json:"threshold" gorm:"comment:阈值"`
	Duration  int             `json:"duration" gorm:"comment:持续时间(秒)"`

	// 通知配置
	NotifyChannels []string `json:"notify_channels" gorm:"type:text;serializer:json;comment:通知渠道"`
	NotifyUsers    []string `json:"notify_users" gorm:"type:text;serializer:json;comment:通知用户"`

	// 状态
	Status AlarmRuleStatus `json:"status" gorm:"default:1;comment:状态(0-禁用,1-启用)"`

	// 审计字段
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedBy string    `json:"created_by" gorm:"type:varchar(100);comment:创建人"`
	UpdatedBy string    `json:"updated_by" gorm:"type:varchar(100);comment:更新人"`
}

// TableName 指定表名
func (r *AlarmRule) TableName() string {
	return "alarm_rules"
}

// NewAlarmRule 创建告警规则
func NewAlarmRule(name string, ruleType AlarmRuleType, level AlarmLevel) *AlarmRule {
	return &AlarmRule{
		ID:             uuid.New().String(),
		Name:           name,
		Type:           ruleType,
		Level:          level,
		Status:         AlarmRuleStatusEnabled,
		NotifyChannels: []string{},
		NotifyUsers:    []string{},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// Enable 启用规则
func (r *AlarmRule) Enable() {
	r.Status = AlarmRuleStatusEnabled
	r.UpdatedAt = time.Now()
}

// Disable 禁用规则
func (r *AlarmRule) Disable() {
	r.Status = AlarmRuleStatusDisabled
	r.UpdatedAt = time.Now()
}

// IsEnabled 检查是否启用
func (r *AlarmRule) IsEnabled() bool {
	return r.Status == AlarmRuleStatusEnabled
}

// SetPoint 设置关联采集点
func (r *AlarmRule) SetPoint(pointID string) {
	r.PointID = &pointID
	r.UpdatedAt = time.Now()
}

// SetDevice 设置关联设备
func (r *AlarmRule) SetDevice(deviceID string) {
	r.DeviceID = &deviceID
	r.UpdatedAt = time.Now()
}

// SetStation 设置关联厂站
func (r *AlarmRule) SetStation(stationID string) {
	r.StationID = &stationID
	r.UpdatedAt = time.Now()
}

// SetNotifyChannels 设置通知渠道
func (r *AlarmRule) SetNotifyChannels(channels []string) {
	r.NotifyChannels = channels
	r.UpdatedAt = time.Now()
}

// SetNotifyUsers 设置通知用户
func (r *AlarmRule) SetNotifyUsers(users []string) {
	r.NotifyUsers = users
	r.UpdatedAt = time.Now()
}

// Update 更新规则
func (r *AlarmRule) Update(name, description string, level AlarmLevel, condition string, threshold float64, duration int) {
	r.Name = name
	r.Description = description
	r.Level = level
	r.Condition = condition
	r.Threshold = threshold
	r.Duration = duration
	r.UpdatedAt = time.Now()
}
