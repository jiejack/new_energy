package repository

import (
	"context"

	"github.com/new-energy-monitoring/internal/domain/entity"
)

// SystemConfigRepository 系统配置仓储接口
type SystemConfigRepository interface {
	// Create 创建配置
	Create(ctx context.Context, config *entity.SystemConfig) error

	// Update 更新配置
	Update(ctx context.Context, config *entity.SystemConfig) error

	// Delete 删除配置
	Delete(ctx context.Context, id string) error

	// GetByID 根据ID获取配置
	GetByID(ctx context.Context, id string) (*entity.SystemConfig, error)

	// GetByKey 根据分类和键获取配置
	GetByKey(ctx context.Context, category, key string) (*entity.SystemConfig, error)

	// GetByCategory 获取指定分类的所有配置
	GetByCategory(ctx context.Context, category string) ([]*entity.SystemConfig, error)

	// GetAll 获取所有配置
	GetAll(ctx context.Context) ([]*entity.SystemConfig, error)

	// List 分页查询配置列表
	List(ctx context.Context, filter *entity.SystemConfigFilter) ([]*entity.SystemConfig, int64, error)

	// BatchUpdate 批量更新配置
	BatchUpdate(ctx context.Context, configs []*entity.SystemConfig) error

	// ExistsByKey 检查配置键是否存在
	ExistsByKey(ctx context.Context, category, key string) (bool, error)
}
