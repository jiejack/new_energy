package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/new-energy-monitoring/pkg/config"
)

// ConfigItem 配置项实体
type ConfigItem struct {
	ID          string            `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Key         string            `json:"key" gorm:"type:varchar(200);uniqueIndex;not null"`
	Value       string            `json:"value" gorm:"type:text"`
	ValueType   config.ValueType  `json:"value_type" gorm:"type:varchar(20);default:'string'"`
	Env         string            `json:"env" gorm:"type:varchar(20);not null;index"`
	Namespace   string            `json:"namespace" gorm:"type:varchar(100);default:'default';index"`
	Group       string            `json:"group" gorm:"type:varchar(100);default:'default';index"`
	Description string            `json:"description" gorm:"type:text"`
	Encrypted   bool              `json:"encrypted" gorm:"default:false"`
	Enabled     bool              `json:"enabled" gorm:"default:true"`
	CreatedAt   time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time         `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联
	Versions   []*ConfigVersion  `json:"versions,omitempty" gorm:"foreignKey:ConfigID"`
	Releases   []*ConfigRelease  `json:"releases,omitempty" gorm:"foreignKey:ConfigID"`
	Audits     []*ConfigAudit    `json:"audits,omitempty" gorm:"foreignKey:ConfigID"`
}

// TableName 返回表名
func (c *ConfigItem) TableName() string {
	return "config_items"
}

// NewConfigItem 创建配置项
func NewConfigItem(key, value string, env, namespace, group string) *ConfigItem {
	return &ConfigItem{
		ID:        uuid.New().String(),
		Key:       key,
		Value:     value,
		ValueType: config.ValueTypeString,
		Env:       env,
		Namespace: namespace,
		Group:     group,
		Enabled:   true,
	}
}

// SetValue 设置配置值
func (c *ConfigItem) SetValue(value string, valueType config.ValueType) {
	c.Value = value
	c.ValueType = valueType
	c.UpdatedAt = time.Now()
}

// SetEncrypted 设置加密状态
func (c *ConfigItem) SetEncrypted(encrypted bool) {
	c.Encrypted = encrypted
	c.UpdatedAt = time.Now()
}

// Enable 启用配置
func (c *ConfigItem) Enable() {
	c.Enabled = true
	c.UpdatedAt = time.Now()
}

// Disable 禁用配置
func (c *ConfigItem) Disable() {
	c.Enabled = false
	c.UpdatedAt = time.Now()
}

// IsEnabled 检查是否启用
func (c *ConfigItem) IsEnabled() bool {
	return c.Enabled
}

// CreateVersion 创建配置版本
func (c *ConfigItem) CreateVersion(version int, changeReason, changedBy string) *ConfigVersion {
	return &ConfigVersion{
		ID:           uuid.New().String(),
		ConfigID:     c.ID,
		Version:      version,
		Value:        c.Value,
		ChangeReason: changeReason,
		ChangedBy:    changedBy,
		CreatedAt:    time.Now(),
	}
}

// CreateRelease 创建配置发布
func (c *ConfigItem) CreateRelease(version int, env string, releaseType config.ReleaseType, targetInstances []string, releasedBy string) *ConfigRelease {
	return &ConfigRelease{
		ID:              uuid.New().String(),
		ConfigID:        c.ID,
		Version:         version,
		Env:             env,
		ReleaseType:     releaseType,
		TargetInstances: targetInstances,
		Status:          config.ReleaseStatusPending,
		ReleasedBy:      releasedBy,
		CreatedAt:       time.Now(),
	}
}

// CreateAudit 创建审计记录
func (c *ConfigItem) CreateAudit(action config.AuditAction, oldValue, newValue, operator, ipAddress string) *ConfigAudit {
	return &ConfigAudit{
		ID:        uuid.New().String(),
		ConfigID:  c.ID,
		Action:    action,
		OldValue:  oldValue,
		NewValue:  newValue,
		Operator:  operator,
		IPAddress: ipAddress,
		CreatedAt: time.Now(),
	}
}

// ConfigVersion 配置版本实体
type ConfigVersion struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	ConfigID     string    `json:"config_id" gorm:"type:varchar(36);index;not null"`
	Version      int       `json:"version" gorm:"not null"`
	Value        string    `json:"value" gorm:"type:text"`
	ChangeReason string    `json:"change_reason" gorm:"type:text"`
	ChangedBy    string    `json:"changed_by" gorm:"type:varchar(100)"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`

	// 关联
	Config *ConfigItem `json:"config,omitempty" gorm:"foreignKey:ConfigID"`
}

// TableName 返回表名
func (c *ConfigVersion) TableName() string {
	return "config_versions"
}

// NewConfigVersion 创建配置版本
func NewConfigVersion(configID string, version int, value, changeReason, changedBy string) *ConfigVersion {
	return &ConfigVersion{
		ID:           uuid.New().String(),
		ConfigID:     configID,
		Version:      version,
		Value:        value,
		ChangeReason: changeReason,
		ChangedBy:    changedBy,
		CreatedAt:    time.Now(),
	}
}

// ConfigRelease 配置发布实体
type ConfigRelease struct {
	ID              string                `json:"id" gorm:"primaryKey;type:varchar(36)"`
	ConfigID        string                `json:"config_id" gorm:"type:varchar(36);index;not null"`
	Version         int                   `json:"version" gorm:"not null"`
	Env             string                `json:"env" gorm:"type:varchar(20);not null;index"`
	ReleaseType     config.ReleaseType    `json:"release_type" gorm:"type:varchar(20);not null"`
	TargetInstances []string              `json:"target_instances" gorm:"type:text[]"`
	Status          config.ReleaseStatus  `json:"status" gorm:"type:varchar(20);default:'pending'"`
	ReleasedBy      string                `json:"released_by" gorm:"type:varchar(100)"`
	ReleasedAt      *time.Time            `json:"released_at"`
	CreatedAt       time.Time             `json:"created_at" gorm:"autoCreateTime"`

	// 关联
	Config *ConfigItem `json:"config,omitempty" gorm:"foreignKey:ConfigID"`
}

// TableName 返回表名
func (c *ConfigRelease) TableName() string {
	return "config_releases"
}

// NewConfigRelease 创建配置发布
func NewConfigRelease(configID string, version int, env string, releaseType config.ReleaseType, targetInstances []string, releasedBy string) *ConfigRelease {
	return &ConfigRelease{
		ID:              uuid.New().String(),
		ConfigID:        configID,
		Version:         version,
		Env:             env,
		ReleaseType:     releaseType,
		TargetInstances: targetInstances,
		Status:          config.ReleaseStatusPending,
		ReleasedBy:      releasedBy,
		CreatedAt:       time.Now(),
	}
}

// Start 开始发布
func (c *ConfigRelease) Start() {
	c.Status = config.ReleaseStatusReleasing
}

// Complete 完成发布
func (c *ConfigRelease) Complete() {
	now := time.Now()
	c.Status = config.ReleaseStatusSuccess
	c.ReleasedAt = &now
}

// Fail 发布失败
func (c *ConfigRelease) Fail() {
	c.Status = config.ReleaseStatusFailed
}

// IsPending 是否待发布
func (c *ConfigRelease) IsPending() bool {
	return c.Status == config.ReleaseStatusPending
}

// IsReleasing 是否发布中
func (c *ConfigRelease) IsReleasing() bool {
	return c.Status == config.ReleaseStatusReleasing
}

// IsSuccess 是否发布成功
func (c *ConfigRelease) IsSuccess() bool {
	return c.Status == config.ReleaseStatusSuccess
}

// IsFailed 是否发布失败
func (c *ConfigRelease) IsFailed() bool {
	return c.Status == config.ReleaseStatusFailed
}

// IsGrayRelease 是否灰度发布
func (c *ConfigRelease) IsGrayRelease() bool {
	return c.ReleaseType == config.ReleaseTypeGray
}

// IsFullRelease 是否全量发布
func (c *ConfigRelease) IsFullRelease() bool {
	return c.ReleaseType == config.ReleaseTypeFull
}

// IsRollback 是否回滚发布
func (c *ConfigRelease) IsRollback() bool {
	return c.ReleaseType == config.ReleaseTypeRollback
}

// ConfigAudit 配置审计实体
type ConfigAudit struct {
	ID        string              `json:"id" gorm:"primaryKey;type:varchar(36)"`
	ConfigID  string              `json:"config_id" gorm:"type:varchar(36);index;not null"`
	Action    config.AuditAction  `json:"action" gorm:"type:varchar(50);not null"`
	OldValue  string              `json:"old_value" gorm:"type:text"`
	NewValue  string              `json:"new_value" gorm:"type:text"`
	Operator  string              `json:"operator" gorm:"type:varchar(100)"`
	IPAddress string              `json:"ip_address" gorm:"type:varchar(50)"`
	CreatedAt time.Time           `json:"created_at" gorm:"autoCreateTime"`

	// 关联
	Config *ConfigItem `json:"config,omitempty" gorm:"foreignKey:ConfigID"`
}

// TableName 返回表名
func (c *ConfigAudit) TableName() string {
	return "config_audits"
}

// NewConfigAudit 创建审计记录
func NewConfigAudit(configID string, action config.AuditAction, oldValue, newValue, operator, ipAddress string) *ConfigAudit {
	return &ConfigAudit{
		ID:        uuid.New().String(),
		ConfigID:  configID,
		Action:    action,
		OldValue:  oldValue,
		NewValue:  newValue,
		Operator:  operator,
		IPAddress: ipAddress,
		CreatedAt: time.Now(),
	}
}

// IsCreate 是否创建操作
func (c *ConfigAudit) IsCreate() bool {
	return c.Action == config.AuditActionCreate
}

// IsUpdate 是否更新操作
func (c *ConfigAudit) IsUpdate() bool {
	return c.Action == config.AuditActionUpdate
}

// IsDelete 是否删除操作
func (c *ConfigAudit) IsDelete() bool {
	return c.Action == config.AuditActionDelete
}

// IsRollback 是否回滚操作
func (c *ConfigAudit) IsRollback() bool {
	return c.Action == config.AuditActionRollback
}

// IsRelease 是否发布操作
func (c *ConfigAudit) IsRelease() bool {
	return c.Action == config.AuditActionRelease
}

// ConfigItemFilter 配置项查询过滤器
type ConfigItemFilter struct {
	Env       *string
	Namespace *string
	Group     *string
	Key       *string
	Enabled   *bool
	Page      int
	PageSize  int
}

// ConfigVersionFilter 配置版本查询过滤器
type ConfigVersionFilter struct {
	ConfigID *string
	MinVersion *int
	MaxVersion *int
	Page     int
	PageSize int
}

// ConfigReleaseFilter 配置发布查询过滤器
type ConfigReleaseFilter struct {
	ConfigID   *string
	Env        *string
	Status     *config.ReleaseStatus
	ReleaseType *config.ReleaseType
	Page       int
	PageSize   int
}

// ConfigAuditFilter 配置审计查询过滤器
type ConfigAuditFilter struct {
	ConfigID *string
	Action   *config.AuditAction
	Operator *string
	StartTime *time.Time
	EndTime   *time.Time
	Page     int
	PageSize int
}
