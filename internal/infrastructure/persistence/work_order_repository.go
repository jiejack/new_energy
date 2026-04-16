package persistence

import (
	"context"

	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

// WorkOrderRepository 工单仓储
type WorkOrderRepository struct {
	db *Database
}

// NewWorkOrderRepository 创建工单仓储
func NewWorkOrderRepository(db *Database) repository.WorkOrderRepository {
	return &WorkOrderRepository{db: db}
}

// Create 创建工单
func (r *WorkOrderRepository) Create(ctx context.Context, workOrder *entity.WorkOrder) error {
	return r.db.WithContext(ctx).Create(workOrder).Error
}

// Update 更新工单
func (r *WorkOrderRepository) Update(ctx context.Context, workOrder *entity.WorkOrder) error {
	return r.db.WithContext(ctx).Save(workOrder).Error
}

// Delete 删除工单
func (r *WorkOrderRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.WorkOrder{}, "id = ?", id).Error
}

// GetByID 根据ID获取工单
func (r *WorkOrderRepository) GetByID(ctx context.Context, id string) (*entity.WorkOrder, error) {
	var workOrder entity.WorkOrder
	err := r.db.WithContext(ctx).First(&workOrder, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &workOrder, nil
}

// List 获取工单列表
func (r *WorkOrderRepository) List(ctx context.Context, filter interface{}) ([]*entity.WorkOrder, error) {
	var workOrders []*entity.WorkOrder
	query := r.db.WithContext(ctx)

	if f, ok := filter.(*service.WorkOrderFilter); ok {
		if f.DeviceID != nil {
			query = query.Where("device_id = ?", *f.DeviceID)
		}
		if f.Type != nil {
			query = query.Where("type = ?", *f.Type)
		}
		if f.Status != "" {
			query = query.Where("status = ?", f.Status)
		}
		if f.Priority != nil {
			query = query.Where("priority = ?", *f.Priority)
		}
		if f.Assignee != nil {
			query = query.Where("assignee = ?", *f.Assignee)
		}
		if f.CreatedBy != nil {
			query = query.Where("created_by = ?", *f.CreatedBy)
		}
		if f.StartDate != nil {
			query = query.Where("created_at >= ?", *f.StartDate)
		}
		if f.EndDate != nil {
			query = query.Where("created_at <= ?", *f.EndDate)
		}
	}

	err := query.Order("created_at DESC").Find(&workOrders).Error
	return workOrders, err
}

// Count 统计工单数量
func (r *WorkOrderRepository) Count(ctx context.Context, filter interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&entity.WorkOrder{})

	if f, ok := filter.(*service.WorkOrderFilter); ok {
		if f.DeviceID != nil {
			query = query.Where("device_id = ?", *f.DeviceID)
		}
		if f.Type != nil {
			query = query.Where("type = ?", *f.Type)
		}
		if f.Status != "" {
			query = query.Where("status = ?", f.Status)
		}
		if f.Priority != nil {
			query = query.Where("priority = ?", *f.Priority)
		}
		if f.Assignee != nil {
			query = query.Where("assignee = ?", *f.Assignee)
		}
		if f.CreatedBy != nil {
			query = query.Where("created_by = ?", *f.CreatedBy)
		}
		if f.StartDate != nil {
			query = query.Where("created_at >= ?", *f.StartDate)
		}
		if f.EndDate != nil {
			query = query.Where("created_at <= ?", *f.EndDate)
		}
	}

	err := query.Count(&count).Error
	return count, err
}
