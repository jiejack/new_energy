package persistence

import (
	"context"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type faultDetectionResultRepository struct {
	db *Database
}

func NewFaultDetectionResultRepository(db *Database) repository.FaultDetectionResultRepository {
	return &faultDetectionResultRepository{db: db}
}

func (r *faultDetectionResultRepository) Create(ctx context.Context, result *entity.FaultDetectionResult) error {
	return r.db.WithContext(ctx).Create(result).Error
}

func (r *faultDetectionResultRepository) GetByID(ctx context.Context, id string) (*entity.FaultDetectionResult, error) {
	var result entity.FaultDetectionResult
	if err := r.db.WithContext(ctx).First(&result, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *faultDetectionResultRepository) ListByDevice(ctx context.Context, deviceID string, severity *entity.FaultSeverity, status *entity.FaultDetectionStatus, offset, limit int) ([]*entity.FaultDetectionResult, int64, error) {
	var results []*entity.FaultDetectionResult
	var total int64
	query := r.db.WithContext(ctx).Model(&entity.FaultDetectionResult{}).Where("device_id = ?", deviceID)
	if severity != nil {
		query = query.Where("severity = ?", *severity)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&results).Error
	return results, total, err
}

func (r *faultDetectionResultRepository) ListByStation(ctx context.Context, stationID string, severity *entity.FaultSeverity, offset, limit int) ([]*entity.FaultDetectionResult, int64, error) {
	var results []*entity.FaultDetectionResult
	var total int64
	query := r.db.WithContext(ctx).Model(&entity.FaultDetectionResult{}).
		Joins("JOIN devices ON devices.id = fault_detection_results.device_id").
		Where("devices.station_id = ?", stationID)
	if severity != nil {
		query = query.Where("fault_detection_results.severity = ?", *severity)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("fault_detection_results.created_at DESC").Offset(offset).Limit(limit).Find(&results).Error
	return results, total, err
}

func (r *faultDetectionResultRepository) UpdateStatus(ctx context.Context, id string, status entity.FaultDetectionStatus) error {
	return r.db.WithContext(ctx).Model(&entity.FaultDetectionResult{}).Where("id = ?", id).Update("status", status).Error
}

func (r *faultDetectionResultRepository) UpdateRootCause(ctx context.Context, id, rootCause string) error {
	return r.db.WithContext(ctx).Model(&entity.FaultDetectionResult{}).Where("id = ?", id).Update("root_cause", rootCause).Error
}

func (r *faultDetectionResultRepository) LinkWorkOrder(ctx context.Context, id, workOrderID string) error {
	return r.db.WithContext(ctx).Model(&entity.FaultDetectionResult{}).Where("id = ?", id).Update("work_order_id", workOrderID).Error
}

func (r *faultDetectionResultRepository) CountBySeverity(ctx context.Context, deviceID *string) (map[entity.FaultSeverity]int64, error) {
	type countResult struct {
		Severity entity.FaultSeverity
		Count    int64
	}
	var counts []countResult
	query := r.db.WithContext(ctx).Model(&entity.FaultDetectionResult{}).
		Select("severity, count(*) as count").
		Where("status IN ?", []entity.FaultDetectionStatus{entity.FaultDetectionStatusPending, entity.FaultDetectionStatusConfirmed}).
		Group("severity")
	if deviceID != nil {
		query = query.Where("device_id = ?", *deviceID)
	}
	if err := query.Scan(&counts).Error; err != nil {
		return nil, err
	}
	result := make(map[entity.FaultSeverity]int64)
	for _, c := range counts {
		result[c.Severity] = c.Count
	}
	return result, nil
}
