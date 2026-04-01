package entity

import (
	"time"

	"github.com/google/uuid"
)

// SystemConfigValueType 系统配置值类型
type SystemConfigValueType string

const (
	SystemConfigValueTypeString SystemConfigValueType = "string" // 字符串
	SystemConfigValueTypeInt    SystemConfigValueType = "int"    // 整数
	SystemConfigValueTypeBool   SystemConfigValueType = "bool"   // 布尔
	SystemConfigValueTypeJSON   SystemConfigValueType = "json"   // JSON对象
)

// SystemConfig 系统配置实体
type SystemConfig struct {
	ID          string                `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Category    string                `json:"category" gorm:"type:varchar(50);not null;uniqueIndex:idx_category_key"`
	Key         string                `json:"key" gorm:"type:varchar(100);not null;uniqueIndex:idx_category_key"`
	Value       string                `json:"value" gorm:"type:text"`
	ValueType   SystemConfigValueType `json:"value_type" gorm:"type:varchar(20);default:'string'"`
	Description string                `json:"description" gorm:"type:text"`
	
	CreatedAt   time.Time             `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 返回表名
func (c *SystemConfig) TableName() string {
	return "system_configs"
}

// NewSystemConfig 创建系统配置
func NewSystemConfig(category, key, value string, valueType SystemConfigValueType, description string) *SystemConfig {
	return &SystemConfig{
		ID:          uuid.New().String(),
		Category:    category,
		Key:         key,
		Value:       value,
		ValueType:   valueType,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// SetValue 设置配置值
func (c *SystemConfig) SetValue(value string, valueType SystemConfigValueType) {
	c.Value = value
	c.ValueType = valueType
	c.UpdatedAt = time.Now()
}

// SetDescription 设置描述
func (c *SystemConfig) SetDescription(description string) {
	c.Description = description
	c.UpdatedAt = time.Now()
}

// IsString 是否字符串类型
func (c *SystemConfig) IsString() bool {
	return c.ValueType == SystemConfigValueTypeString
}

// IsInt 是否整数类型
func (c *SystemConfig) IsInt() bool {
	return c.ValueType == SystemConfigValueTypeInt
}

// IsBool 是否布尔类型
func (c *SystemConfig) IsBool() bool {
	return c.ValueType == SystemConfigValueTypeBool
}

// IsJSON 是否JSON类型
func (c *SystemConfig) IsJSON() bool {
	return c.ValueType == SystemConfigValueTypeJSON
}

// SystemConfigFilter 系统配置查询过滤器
type SystemConfigFilter struct {
	Category *string
	Key      *string
	Page     int
	PageSize int
}
