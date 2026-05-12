package persistence

import (
	"context"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type forecastResultRepository struct {
	db *Database
}

func NewForecastResultRepository(db *Database) repository.ForecastResultRepository {
	return &forecastResultRepository{db: db}
}

func (r *forecastResultRepository) Create(ctx context.Context, result *entity.ForecastResult) error {
	return r.db.WithContext(ctx).Create(result).Error
}

func (r *forecastResultRepository) GetByID(ctx context.Context, id string) (*entity.ForecastResult, error) {
	var result entity.ForecastResult
	if err := r.db.WithContext(ctx).First(&result, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *forecastResultRepository) ListByStation(ctx context.Context, stationID string, forecastType *entity.ForecastType, startTime, endTime *time.Time, offset, limit int) ([]*entity.ForecastResult, int64, error) {
	var results []*entity.ForecastResult
	var total int64
	query := r.db.WithContext(ctx).Model(&entity.ForecastResult{}).Where("station_id = ?", stationID)
	if forecastType != nil {
		query = query.Where("forecast_type = ?", *forecastType)
	}
	if startTime != nil {
		query = query.Where("target_time >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("target_time <= ?", *endTime)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("target_time DESC").Offset(offset).Limit(limit).Find(&results).Error
	return results, total, err
}

func (r *forecastResultRepository) UpdateActualPower(ctx context.Context, id string, actualPower float64) error {
	return r.db.WithContext(ctx).Model(&entity.ForecastResult{}).Where("id = ?", id).Update("actual_power", actualPower).Error
}

func (r *forecastResultRepository) GetAccuracyStats(ctx context.Context, stationID string, forecastType *entity.ForecastType, start, end time.Time) (*repository.ForecastAccuracyStats, error) {
	var results []*entity.ForecastResult
	query := r.db.WithContext(ctx).Where("station_id = ? AND actual_power IS NOT NULL", stationID)
	if forecastType != nil {
		query = query.Where("forecast_type = ?", *forecastType)
	}
	query = query.Where("target_time BETWEEN ? AND ?", start, end)
	if err := query.Find(&results).Error; err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return &repository.ForecastAccuracyStats{StationID: stationID}, nil
	}
	var sumSquaredErr, sumAbsErr, sumAccuracy float64
	for _, r := range results {
		if r.ActualPower != nil && *r.ActualPower != 0 {
			err := r.PredictedPower - *r.ActualPower
			sumSquaredErr += err * err
			sumAbsErr += abs(err)
			if r.Accuracy != nil {
				sumAccuracy += *r.Accuracy
			}
		}
	}
	n := float64(len(results))
	return &repository.ForecastAccuracyStats{
		StationID:    stationID,
		TotalPoints:  int64(len(results)),
		AvgAccuracy:  sumAccuracy / n,
		RMSE:         sqrt(sumSquaredErr / n),
		MAE:          sumAbsErr / n,
	}, nil
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}
