package persistence

import (
	"context"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type carbonEmissionRepository struct {
	db *Database
}

func NewCarbonEmissionRepository(db *Database) repository.CarbonEmissionRepository {
	return &carbonEmissionRepository{db: db}
}

func (r *carbonEmissionRepository) CreateFactor(ctx context.Context, factor *entity.CarbonEmissionFactor) error {
	return r.db.WithContext(ctx).Create(factor).Error
}

func (r *carbonEmissionRepository) UpdateFactor(ctx context.Context, factor *entity.CarbonEmissionFactor) error {
	return r.db.WithContext(ctx).Save(factor).Error
}

func (r *carbonEmissionRepository) GetFactorByID(ctx context.Context, id string) (*entity.CarbonEmissionFactor, error) {
	var factor entity.CarbonEmissionFactor
	err := r.db.WithContext(ctx).First(&factor, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &factor, nil
}

func (r *carbonEmissionRepository) GetFactorByCode(ctx context.Context, code string) (*entity.CarbonEmissionFactor, error) {
	var factor entity.CarbonEmissionFactor
	err := r.db.WithContext(ctx).First(&factor, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &factor, nil
}

func (r *carbonEmissionRepository) ListFactors(ctx context.Context, query *repository.CarbonEmissionFactorQuery) ([]*entity.CarbonEmissionFactor, int64, error) {
	var factors []*entity.CarbonEmissionFactor
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.CarbonEmissionFactor{})

	if query.Scope != nil {
		db = db.Where("scope = ?", *query.Scope)
	}
	if query.Source != nil {
		db = db.Where("source LIKE ?", "%"+*query.Source+"%")
	}
	if query.IsActive != nil {
		db = db.Where("is_active = ?", *query.IsActive)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Order("created_at DESC").Find(&factors).Error; err != nil {
		return nil, 0, err
	}

	return factors, total, nil
}

func (r *carbonEmissionRepository) GetActiveFactors(ctx context.Context, scope *entity.CarbonEmissionScope) ([]*entity.CarbonEmissionFactor, error) {
	var factors []*entity.CarbonEmissionFactor
	db := r.db.WithContext(ctx).Where("is_active = ?", true)
	
	if scope != nil {
		db = db.Where("scope = ?", *scope)
	}
	
	err := db.Order("created_at DESC").Find(&factors).Error
	return factors, err
}

func (r *carbonEmissionRepository) CreateRecord(ctx context.Context, record *entity.CarbonEmissionRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *carbonEmissionRepository) BatchCreateRecords(ctx context.Context, records []*entity.CarbonEmissionRecord) error {
	return r.db.WithContext(ctx).CreateInBatches(records, 100).Error
}

func (r *carbonEmissionRepository) UpdateRecord(ctx context.Context, record *entity.CarbonEmissionRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

func (r *carbonEmissionRepository) GetRecordByID(ctx context.Context, id string) (*entity.CarbonEmissionRecord, error) {
	var record entity.CarbonEmissionRecord
	err := r.db.WithContext(ctx).First(&record, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *carbonEmissionRepository) ListRecords(ctx context.Context, query *repository.CarbonEmissionRecordQuery) ([]*entity.CarbonEmissionRecord, int64, error) {
	var records []*entity.CarbonEmissionRecord
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.CarbonEmissionRecord{})

	if query.Scope != nil {
		db = db.Where("scope = ?", *query.Scope)
	}
	if query.TargetID != nil {
		db = db.Where("target_id = ?", *query.TargetID)
	}
	if query.Period != nil {
		db = db.Where("period = ?", *query.Period)
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.StartTime != nil {
		db = db.Where("record_time >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("record_time <= ?", *query.EndTime)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Order("record_time DESC").Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

func (r *carbonEmissionRepository) GetRecordsByTimeRange(ctx context.Context, targetID string, scope *entity.CarbonEmissionScope, startTime, endTime time.Time) ([]*entity.CarbonEmissionRecord, error) {
	var records []*entity.CarbonEmissionRecord
	db := r.db.WithContext(ctx).
		Where("target_id = ? AND record_time >= ? AND record_time <= ?", targetID, startTime, endTime)
	
	if scope != nil {
		db = db.Where("scope = ?", *scope)
	}
	
	err := db.Order("record_time ASC").Find(&records).Error
	return records, err
}

func (r *carbonEmissionRepository) GetTotalEmissionByScope(ctx context.Context, targetID string, scope entity.CarbonEmissionScope, period string, startTime, endTime time.Time) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).
		Model(&entity.CarbonEmissionRecord{}).
		Select("COALESCE(SUM(emission_value), 0)").
		Where("target_id = ? AND scope = ? AND period = ? AND record_time >= ? AND record_time <= ?", targetID, scope, period, startTime, endTime).
		Scan(&total).Error
	return total, err
}

func (r *carbonEmissionRepository) CreateSummary(ctx context.Context, summary *entity.CarbonEmissionSummary) error {
	return r.db.WithContext(ctx).Create(summary).Error
}

func (r *carbonEmissionRepository) UpdateSummary(ctx context.Context, summary *entity.CarbonEmissionSummary) error {
	return r.db.WithContext(ctx).Save(summary).Error
}

func (r *carbonEmissionRepository) GetSummaryByID(ctx context.Context, id string) (*entity.CarbonEmissionSummary, error) {
	var summary entity.CarbonEmissionSummary
	err := r.db.WithContext(ctx).First(&summary, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

func (r *carbonEmissionRepository) GetLatestSummary(ctx context.Context, targetID string, period string) (*entity.CarbonEmissionSummary, error) {
	var summary entity.CarbonEmissionSummary
	err := r.db.WithContext(ctx).
		Where("target_id = ? AND period = ?", targetID, period).
		Order("summary_time DESC").
		First(&summary).Error
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

func (r *carbonEmissionRepository) ListSummaries(ctx context.Context, query *repository.CarbonEmissionSummaryQuery) ([]*entity.CarbonEmissionSummary, int64, error) {
	var summaries []*entity.CarbonEmissionSummary
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.CarbonEmissionSummary{})

	if query.TargetID != nil {
		db = db.Where("target_id = ?", *query.TargetID)
	}
	if query.Period != nil {
		db = db.Where("period = ?", *query.Period)
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.StartTime != nil {
		db = db.Where("summary_time >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("summary_time <= ?", *query.EndTime)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Order("summary_time DESC").Find(&summaries).Error; err != nil {
		return nil, 0, err
	}

	return summaries, total, nil
}

func (r *carbonEmissionRepository) CreateTarget(ctx context.Context, target *entity.CarbonReductionTarget) error {
	return r.db.WithContext(ctx).Create(target).Error
}

func (r *carbonEmissionRepository) UpdateTarget(ctx context.Context, target *entity.CarbonReductionTarget) error {
	return r.db.WithContext(ctx).Save(target).Error
}

func (r *carbonEmissionRepository) GetTargetByID(ctx context.Context, id string) (*entity.CarbonReductionTarget, error) {
	var target entity.CarbonReductionTarget
	err := r.db.WithContext(ctx).First(&target, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &target, nil
}

func (r *carbonEmissionRepository) ListTargets(ctx context.Context, query *repository.CarbonReductionTargetQuery) ([]*entity.CarbonReductionTarget, int64, error) {
	var targets []*entity.CarbonReductionTarget
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.CarbonReductionTarget{})

	if query.TargetID != nil {
		db = db.Where("target_id = ?", *query.TargetID)
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.StartYear != nil {
		db = db.Where("target_year >= ?", *query.StartYear)
	}
	if query.EndYear != nil {
		db = db.Where("target_year <= ?", *query.EndYear)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Order("created_at DESC").Find(&targets).Error; err != nil {
		return nil, 0, err
	}

	return targets, total, nil
}

func (r *carbonEmissionRepository) GetActiveTargets(ctx context.Context, targetID string) ([]*entity.CarbonReductionTarget, error) {
	var targets []*entity.CarbonReductionTarget
	err := r.db.WithContext(ctx).
		Where("target_id = ? AND status = ?", targetID, "active").
		Order("created_at DESC").
		Find(&targets).Error
	return targets, err
}
