package persistence

import (
	"context"
	"errors"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"gorm.io/gorm"
)

// SystemConfigRepository 系统配置仓储实现
type SystemConfigRepository struct {
	db *Database
}

// NewSystemConfigRepository 创建系统配置仓储
func NewSystemConfigRepository(db *Database) *SystemConfigRepository {
	return &SystemConfigRepository{db: db}
}

// Create 创建配置
func (r *SystemConfigRepository) Create(ctx context.Context, config *entity.SystemConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// Update 更新配置
func (r *SystemConfigRepository) Update(ctx context.Context, config *entity.SystemConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

// Delete 删除配置
func (r *SystemConfigRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.SystemConfig{}, "id = ?", id).Error
}

// GetByID 根据ID获取配置
func (r *SystemConfigRepository) GetByID(ctx context.Context, id string) (*entity.SystemConfig, error) {
	var config entity.SystemConfig
	err := r.db.WithContext(ctx).First(&config, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &config, nil
}

// GetByKey 根据分类和键获取配置
func (r *SystemConfigRepository) GetByKey(ctx context.Context, category, key string) (*entity.SystemConfig, error) {
	var config entity.SystemConfig
	err := r.db.WithContext(ctx).
		Where("category = ? AND key = ?", category, key).
		First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &config, nil
}

// GetByCategory 获取指定分类的所有配置
func (r *SystemConfigRepository) GetByCategory(ctx context.Context, category string) ([]*entity.SystemConfig, error) {
	var configs []*entity.SystemConfig
	err := r.db.WithContext(ctx).
		Where("category = ?", category).
		Order("key ASC").
		Find(&configs).Error
	return configs, err
}

// GetAll 获取所有配置
func (r *SystemConfigRepository) GetAll(ctx context.Context) ([]*entity.SystemConfig, error) {
	var configs []*entity.SystemConfig
	err := r.db.WithContext(ctx).
		Order("category ASC, key ASC").
		Find(&configs).Error
	return configs, err
}

// List 分页查询配置列表
func (r *SystemConfigRepository) List(ctx context.Context, filter *entity.SystemConfigFilter) ([]*entity.SystemConfig, int64, error) {
	var configs []*entity.SystemConfig
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.SystemConfig{})

	// 应用过滤条件
	if filter.Category != nil {
		query = query.Where("category = ?", *filter.Category)
	}
	if filter.Key != nil {
		query = query.Where("key LIKE ?", "%"+*filter.Key+"%")
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (filter.Page - 1) * filter.PageSize
	err := query.Offset(offset).Limit(filter.PageSize).
		Order("category ASC, key ASC").
		Find(&configs).Error

	return configs, total, err
}

// BatchUpdate 批量更新配置
func (r *SystemConfigRepository) BatchUpdate(ctx context.Context, configs []*entity.SystemConfig) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, config := range configs {
			if err := tx.Save(config).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ExistsByKey 检查配置键是否存在
func (r *SystemConfigRepository) ExistsByKey(ctx context.Context, category, key string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.SystemConfig{}).
		Where("category = ? AND key = ?", category, key).
		Count(&count).Error
	return count > 0, err
}
