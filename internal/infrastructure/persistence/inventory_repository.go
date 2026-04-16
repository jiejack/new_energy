package persistence

import (
	"context"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

// InventoryRepository 库存仓储实现
type InventoryRepository struct {
	db *Database
}

// NewInventoryRepository 创建库存仓储
func NewInventoryRepository(db *Database) repository.InventoryRepository {
	return &InventoryRepository{db: db}
}

// Create 创建库存
func (r *InventoryRepository) Create(ctx context.Context, inventory *entity.Inventory) error {
	return r.db.WithContext(ctx).Create(inventory).Error
}

// Update 更新库存
func (r *InventoryRepository) Update(ctx context.Context, inventory *entity.Inventory) error {
	return r.db.WithContext(ctx).Save(inventory).Error
}

// Delete 删除库存
func (r *InventoryRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Inventory{}, "id = ?", id).Error
}

// GetByID 根据ID获取库存
func (r *InventoryRepository) GetByID(ctx context.Context, id string) (*entity.Inventory, error) {
	var inventory entity.Inventory
	err := r.db.WithContext(ctx).First(&inventory, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &inventory, nil
}

// GetByCode 根据编码获取库存
func (r *InventoryRepository) GetByCode(ctx context.Context, code string) (*entity.Inventory, error) {
	var inventory entity.Inventory
	err := r.db.WithContext(ctx).First(&inventory, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &inventory, nil
}

// List 获取库存列表
func (r *InventoryRepository) List(ctx context.Context, filter interface{}) ([]*entity.Inventory, error) {
	var inventories []*entity.Inventory
	query := r.db.WithContext(ctx)

	// 这里可以根据filter参数添加过滤条件
	// 例如：if f, ok := filter.(*InventoryFilter); ok {
	//     if f.Type != "" {
	//         query = query.Where("type = ?", f.Type)
	//     }
	//     if f.Status != "" {
	//         query = query.Where("status = ?", f.Status)
	//     }
	// }

	err := query.Find(&inventories).Error
	return inventories, err
}

// Count 统计库存数量
func (r *InventoryRepository) Count(ctx context.Context, filter interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&entity.Inventory{})

	// 这里可以根据filter参数添加过滤条件

	err := query.Count(&count).Error
	return count, err
}

// UpdateQuantity 更新库存数量
func (r *InventoryRepository) UpdateQuantity(ctx context.Context, id string, quantity float64) error {
	return r.db.WithContext(ctx).Model(&entity.Inventory{}).Where("id = ?", id).Update("quantity", quantity).Error
}

// GetLowStockItems 获取低库存物品
func (r *InventoryRepository) GetLowStockItems(ctx context.Context) ([]*entity.Inventory, error) {
	var inventories []*entity.Inventory
	err := r.db.WithContext(ctx).Where("quantity < min_quantity").Find(&inventories).Error
	return inventories, err
}
