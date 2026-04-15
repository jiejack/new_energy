package persistence

import (
	"context"
	"math"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"gorm.io/gorm"
)

type energyEfficiencyRepository struct {
	db *Database
}

func NewEnergyEfficiencyRepository(db *Database) repository.EnergyEfficiencyRepository {
	return &energyEfficiencyRepository{db: db}
}

func (r *energyEfficiencyRepository) CreateRecord(ctx context.Context, record *entity.EnergyEfficiencyRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *energyEfficiencyRepository) BatchCreateRecords(ctx context.Context, records []*entity.EnergyEfficiencyRecord) error {
	return r.db.WithContext(ctx).CreateInBatches(records, 100).Error
}

func (r *energyEfficiencyRepository) GetRecordByID(ctx context.Context, id string) (*entity.EnergyEfficiencyRecord, error) {
	var record entity.EnergyEfficiencyRecord
	err := r.db.WithContext(ctx).First(&record, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *energyEfficiencyRepository) ListRecords(ctx context.Context, query *repository.EnergyEfficiencyQuery) ([]*entity.EnergyEfficiencyRecord, int64, error) {
	var records []*entity.EnergyEfficiencyRecord
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.EnergyEfficiencyRecord{})

	if query.Type != nil {
		db = db.Where("type = ?", *query.Type)
	}
	if query.Level != nil {
		db = db.Where("efficiency_level = ?", *query.Level)
	}
	if query.TargetID != nil {
		db = db.Where("target_id = ?", *query.TargetID)
	}
	if query.Period != nil {
		db = db.Where("period = ?", *query.Period)
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

func (r *energyEfficiencyRepository) GetRecordsByTimeRange(ctx context.Context, targetID string, eeType entity.EnergyEfficiencyType, startTime, endTime time.Time) ([]*entity.EnergyEfficiencyRecord, error) {
	var records []*entity.EnergyEfficiencyRecord
	err := r.db.WithContext(ctx).
		Where("target_id = ? AND type = ? AND record_time >= ? AND record_time <= ?", targetID, eeType, startTime, endTime).
		Order("record_time ASC").
		Find(&records).Error
	return records, err
}

func (r *energyEfficiencyRepository) CreateAnalysis(ctx context.Context, analysis *entity.EnergyEfficiencyAnalysis) error {
	return r.db.WithContext(ctx).Create(analysis).Error
}

func (r *energyEfficiencyRepository) GetAnalysisByID(ctx context.Context, id string) (*entity.EnergyEfficiencyAnalysis, error) {
	var analysis entity.EnergyEfficiencyAnalysis
	err := r.db.WithContext(ctx).First(&analysis, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &analysis, nil
}

func (r *energyEfficiencyRepository) ListAnalyses(ctx context.Context, query *repository.EnergyEfficiencyAnalysisQuery) ([]*entity.EnergyEfficiencyAnalysis, int64, error) {
	var analyses []*entity.EnergyEfficiencyAnalysis
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.EnergyEfficiencyAnalysis{})

	if query.Type != nil {
		db = db.Where("type = ?", *query.Type)
	}
	if query.TargetID != nil {
		db = db.Where("target_id = ?", *query.TargetID)
	}
	if query.StartTime != nil {
		db = db.Where("analysis_time >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("analysis_time <= ?", *query.EndTime)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Order("analysis_time DESC").Find(&analyses).Error; err != nil {
		return nil, 0, err
	}

	return analyses, total, nil
}

func (r *energyEfficiencyRepository) GetLatestAnalysis(ctx context.Context, targetID string, eeType entity.EnergyEfficiencyType) (*entity.EnergyEfficiencyAnalysis, error) {
	var analysis entity.EnergyEfficiencyAnalysis
	err := r.db.WithContext(ctx).
		Where("target_id = ? AND type = ?", targetID, eeType).
		Order("analysis_time DESC").
		First(&analysis).Error
	if err != nil {
		return nil, err
	}
	return &analysis, nil
}

func (r *energyEfficiencyRepository) GetStatistics(ctx context.Context, targetID string, eeType entity.EnergyEfficiencyType, period string, startTime, endTime time.Time) (*repository.EnergyEfficiencyStatistics, error) {
	var stats repository.EnergyEfficiencyStatistics

	db := r.db.WithContext(ctx).Model(&entity.EnergyEfficiencyRecord{}).
		Where("target_id = ? AND type = ? AND period = ? AND record_time >= ? AND record_time <= ?", targetID, eeType, period, startTime, endTime)

	var records []*entity.EnergyEfficiencyRecord
	if err := db.Find(&records).Error; err != nil {
		return nil, err
	}

	stats.TotalRecords = int64(len(records))
	if stats.TotalRecords == 0 {
		return &stats, nil
	}

	var sumEfficiency, sumInput, sumOutput float64
	stats.MaxEfficiency = math.Inf(-1)
	stats.MinEfficiency = math.Inf(1)

	for _, record := range records {
		sumEfficiency += record.Efficiency
		sumInput += record.InputEnergy
		sumOutput += record.OutputEnergy

		if record.Efficiency > stats.MaxEfficiency {
			stats.MaxEfficiency = record.Efficiency
		}
		if record.Efficiency < stats.MinEfficiency {
			stats.MinEfficiency = record.Efficiency
		}

		switch record.EfficiencyLevel {
		case entity.EnergyEfficiencyLevelExcellent:
			stats.ExcellentCount++
		case entity.EnergyEfficiencyLevelGood:
			stats.GoodCount++
		case entity.EnergyEfficiencyLevelNormal:
			stats.NormalCount++
		case entity.EnergyEfficiencyLevelPoor:
			stats.PoorCount++
		}
	}

	stats.AvgEfficiency = sumEfficiency / float64(stats.TotalRecords)
	stats.TotalInputEnergy = sumInput
	stats.TotalOutputEnergy = sumOutput

	if stats.TotalRecords > 1 {
		var variance float64
		for _, record := range records {
			diff := record.Efficiency - stats.AvgEfficiency
			variance += diff * diff
		}
		stats.StdDevEfficiency = math.Sqrt(variance / float64(stats.TotalRecords-1))
	}

	return &stats, nil
}

func (r *energyEfficiencyRepository) GetBenchmark(ctx context.Context, eeType entity.EnergyEfficiencyType, targetID string) (float64, error) {
	var benchmark float64
	err := r.db.WithContext(ctx).
		Model(&entity.EnergyEfficiencyRecord{}).
		Select("COALESCE(AVG(efficiency), 0)").
		Where("type = ? AND target_id != ?", eeType, targetID).
		Scan(&benchmark).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}
	return benchmark, nil
}
