package persistence

import (
	"context"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

// SupplierRepository 供应商仓储实现
type SupplierRepository struct {
	db *Database
}

// NewSupplierRepository 创建供应商仓储
func NewSupplierRepository(db *Database) repository.SupplierRepository {
	return &SupplierRepository{db: db}
}

// Create 创建供应商
func (r *SupplierRepository) Create(ctx context.Context, supplier *entity.Supplier) error {
	return r.db.WithContext(ctx).Create(supplier).Error
}

// Update 更新供应商
func (r *SupplierRepository) Update(ctx context.Context, supplier *entity.Supplier) error {
	return r.db.WithContext(ctx).Save(supplier).Error
}

// Delete 删除供应商
func (r *SupplierRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Supplier{}, "id = ?", id).Error
}

// GetByID 根据ID获取供应商
func (r *SupplierRepository) GetByID(ctx context.Context, id string) (*entity.Supplier, error) {
	var supplier entity.Supplier
	err := r.db.WithContext(ctx).First(&supplier, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &supplier, nil
}

// GetByCode 根据编码获取供应商
func (r *SupplierRepository) GetByCode(ctx context.Context, code string) (*entity.Supplier, error) {
	var supplier entity.Supplier
	err := r.db.WithContext(ctx).First(&supplier, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &supplier, nil
}

// List 获取供应商列表
func (r *SupplierRepository) List(ctx context.Context, filter interface{}) ([]*entity.Supplier, error) {
	var suppliers []*entity.Supplier
	query := r.db.WithContext(ctx)

	// 这里可以根据filter参数添加过滤条件
	// 例如：if f, ok := filter.(*SupplierFilter); ok {
	//     if f.Status != "" {
	//         query = query.Where("status = ?", f.Status)
	//     }
	//     if f.Name != "" {
	//         query = query.Where("name LIKE ?", "%"+f.Name+"%")
	//     }
	// }

	err := query.Find(&suppliers).Error
	return suppliers, err
}

// Count 统计供应商数量
func (r *SupplierRepository) Count(ctx context.Context, filter interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&entity.Supplier{})

	// 这里可以根据filter参数添加过滤条件

	err := query.Count(&count).Error
	return count, err
}
