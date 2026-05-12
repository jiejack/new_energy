package persistence

import (
	"context"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type edgeNodeRepository struct {
	db *Database
}

func NewEdgeNodeRepository(db *Database) repository.EdgeNodeRepository {
	return &edgeNodeRepository{db: db}
}

func (r *edgeNodeRepository) Create(ctx context.Context, node *entity.EdgeNode) error {
	return r.db.WithContext(ctx).Create(node).Error
}

func (r *edgeNodeRepository) GetByID(ctx context.Context, id string) (*entity.EdgeNode, error) {
	var node entity.EdgeNode
	if err := r.db.WithContext(ctx).First(&node, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *edgeNodeRepository) List(ctx context.Context, stationID *string, status *entity.EdgeNodeStatus) ([]*entity.EdgeNode, error) {
	var nodes []*entity.EdgeNode
	query := r.db.WithContext(ctx)
	if stationID != nil {
		query = query.Where("station_id = ?", *stationID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	err := query.Order("created_at DESC").Find(&nodes).Error
	return nodes, err
}

func (r *edgeNodeRepository) Update(ctx context.Context, node *entity.EdgeNode) error {
	return r.db.WithContext(ctx).Save(node).Error
}

func (r *edgeNodeRepository) UpdateStatus(ctx context.Context, id string, status entity.EdgeNodeStatus) error {
	return r.db.WithContext(ctx).Model(&entity.EdgeNode{}).Where("id = ?", id).Update("status", status).Error
}

func (r *edgeNodeRepository) UpdateHeartbeat(ctx context.Context, id string, cpuUsage, memoryUsage float64) error {
	return r.db.WithContext(ctx).Model(&entity.EdgeNode{}).Where("id = ?", id).Updates(map[string]interface{}{
		"cpu_usage":    cpuUsage,
		"memory_usage": memoryUsage,
		"status":       entity.EdgeNodeOnline,
	}).Error
}

func (r *edgeNodeRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.EdgeNode{}, "id = ?", id).Error
}
