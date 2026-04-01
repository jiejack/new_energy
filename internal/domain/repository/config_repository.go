package repository

import (
	"context"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/pkg/config"
)

// ConfigItemRepository 配置项仓储接口
type ConfigItemRepository interface {
	// Create 创建配置项
	Create(ctx context.Context, item *entity.ConfigItem) error
	
	// Update 更新配置项
	Update(ctx context.Context, item *entity.ConfigItem) error
	
	// Delete 删除配置项
	Delete(ctx context.Context, id string) error
	
	// GetByID 根据ID获取配置项
	GetByID(ctx context.Context, id string) (*entity.ConfigItem, error)
	
	// GetByKey 根据键获取配置项
	GetByKey(ctx context.Context, key, env, namespace, group string) (*entity.ConfigItem, error)
	
	// List 获取配置项列表
	List(ctx context.Context, filter *entity.ConfigItemFilter) ([]*entity.ConfigItem, int64, error)
	
	// GetByEnv 根据环境获取配置项列表
	GetByEnv(ctx context.Context, env string) ([]*entity.ConfigItem, error)
	
	// GetByNamespace 根据命名空间获取配置项列表
	GetByNamespace(ctx context.Context, namespace string) ([]*entity.ConfigItem, error)
	
	// GetByGroup 根据分组获取配置项列表
	GetByGroup(ctx context.Context, group string) ([]*entity.ConfigItem, error)
	
	// BatchCreate 批量创建配置项
	BatchCreate(ctx context.Context, items []*entity.ConfigItem) error
	
	// BatchUpdate 批量更新配置项
	BatchUpdate(ctx context.Context, items []*entity.ConfigItem) error
	
	// BatchDelete 批量删除配置项
	BatchDelete(ctx context.Context, ids []string) error
	
	// GetEnabledItems 获取启用的配置项
	GetEnabledItems(ctx context.Context) ([]*entity.ConfigItem, error)
	
	// GetByKeys 根据键列表获取配置项
	GetByKeys(ctx context.Context, keys []string, env string) ([]*entity.ConfigItem, error)
}

// ConfigVersionRepository 配置版本仓储接口
type ConfigVersionRepository interface {
	// Create 创建配置版本
	Create(ctx context.Context, version *entity.ConfigVersion) error
	
	// GetByID 根据ID获取配置版本
	GetByID(ctx context.Context, id string) (*entity.ConfigVersion, error)
	
	// GetByConfigID 根据配置ID获取版本列表
	GetByConfigID(ctx context.Context, configID string) ([]*entity.ConfigVersion, error)
	
	// GetByVersion 根据版本号获取配置版本
	GetByVersion(ctx context.Context, configID string, version int) (*entity.ConfigVersion, error)
	
	// GetLatestVersion 获取最新版本
	GetLatestVersion(ctx context.Context, configID string) (*entity.ConfigVersion, error)
	
	// List 获取版本列表
	List(ctx context.Context, filter *entity.ConfigVersionFilter) ([]*entity.ConfigVersion, int64, error)
	
	// Delete 删除配置版本
	Delete(ctx context.Context, id string) error
	
	// DeleteByConfigID 删除配置的所有版本
	DeleteByConfigID(ctx context.Context, configID string) error
	
	// GetVersionCount 获取版本数量
	GetVersionCount(ctx context.Context, configID string) (int64, error)
}

// ConfigReleaseRepository 配置发布仓储接口
type ConfigReleaseRepository interface {
	// Create 创建配置发布
	Create(ctx context.Context, release *entity.ConfigRelease) error
	
	// Update 更新配置发布
	Update(ctx context.Context, release *entity.ConfigRelease) error
	
	// GetByID 根据ID获取配置发布
	GetByID(ctx context.Context, id string) (*entity.ConfigRelease, error)
	
	// GetByConfigID 根据配置ID获取发布列表
	GetByConfigID(ctx context.Context, configID string) ([]*entity.ConfigRelease, error)
	
	// GetLatestRelease 获取最新发布
	GetLatestRelease(ctx context.Context, configID string, env string) (*entity.ConfigRelease, error)
	
	// List 获取发布列表
	List(ctx context.Context, filter *entity.ConfigReleaseFilter) ([]*entity.ConfigRelease, int64, error)
	
	// GetPendingReleases 获取待发布的配置
	GetPendingReleases(ctx context.Context) ([]*entity.ConfigRelease, error)
	
	// GetReleasingReleases 获取发布中的配置
	GetReleasingReleases(ctx context.Context) ([]*entity.ConfigRelease, error)
	
	// Delete 删除配置发布
	Delete(ctx context.Context, id string) error
	
	// DeleteByConfigID 删除配置的所有发布
	DeleteByConfigID(ctx context.Context, configID string) error
}

// ConfigAuditRepository 配置审计仓储接口
type ConfigAuditRepository interface {
	// Create 创建审计记录
	Create(ctx context.Context, audit *entity.ConfigAudit) error
	
	// GetByID 根据ID获取审计记录
	GetByID(ctx context.Context, id string) (*entity.ConfigAudit, error)
	
	// GetByConfigID 根据配置ID获取审计记录列表
	GetByConfigID(ctx context.Context, configID string) ([]*entity.ConfigAudit, error)
	
	// List 获取审计记录列表
	List(ctx context.Context, filter *entity.ConfigAuditFilter) ([]*entity.ConfigAudit, int64, error)
	
	// GetRecentAudits 获取最近的审计记录
	GetRecentAudits(ctx context.Context, limit int) ([]*entity.ConfigAudit, error)
	
	// GetByOperator 根据操作人获取审计记录
	GetByOperator(ctx context.Context, operator string) ([]*entity.ConfigAudit, error)
	
	// GetByAction 根据操作类型获取审计记录
	GetByAction(ctx context.Context, action config.AuditAction) ([]*entity.ConfigAudit, error)
	
	// Delete 删除审计记录
	Delete(ctx context.Context, id string) error
	
	// DeleteByConfigID 删除配置的所有审计记录
	DeleteByConfigID(ctx context.Context, configID string) error
	
	// DeleteOldAudits 删除旧的审计记录
	DeleteOldAudits(ctx context.Context, beforeDays int) error
}

// ConfigRepository 配置仓储聚合接口
type ConfigRepository interface {
	// Items 配置项仓储
	Items() ConfigItemRepository
	
	// Versions 配置版本仓储
	Versions() ConfigVersionRepository
	
	// Releases 配置发布仓储
	Releases() ConfigReleaseRepository
	
	// Audits 配置审计仓储
	Audits() ConfigAuditRepository
	
	// Transaction 事务
	Transaction(ctx context.Context, fn func(repo ConfigRepository) error) error
}

// ConfigItemRepositoryImpl 配置项仓储实现接口（用于基础设施层实现）
type ConfigItemRepositoryImpl interface {
	ConfigItemRepository
	
	// WithPreload 预加载关联
	WithPreload(preloads ...string) ConfigItemRepository
	
	// WithTransaction 使用事务
	WithTransaction(tx interface{}) ConfigItemRepository
}

// ConfigVersionRepositoryImpl 配置版本仓储实现接口（用于基础设施层实现）
type ConfigVersionRepositoryImpl interface {
	ConfigVersionRepository
	
	// WithPreload 预加载关联
	WithPreload(preloads ...string) ConfigVersionRepository
	
	// WithTransaction 使用事务
	WithTransaction(tx interface{}) ConfigVersionRepository
}

// ConfigReleaseRepositoryImpl 配置发布仓储实现接口（用于基础设施层实现）
type ConfigReleaseRepositoryImpl interface {
	ConfigReleaseRepository
	
	// WithPreload 预加载关联
	WithPreload(preloads ...string) ConfigReleaseRepository
	
	// WithTransaction 使用事务
	WithTransaction(tx interface{}) ConfigReleaseRepository
}

// ConfigAuditRepositoryImpl 配置审计仓储实现接口（用于基础设施层实现）
type ConfigAuditRepositoryImpl interface {
	ConfigAuditRepository
	
	// WithPreload 预加载关联
	WithPreload(preloads ...string) ConfigAuditRepository
	
	// WithTransaction 使用事务
	WithTransaction(tx interface{}) ConfigAuditRepository
}
