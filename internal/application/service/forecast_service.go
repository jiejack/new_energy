package service

import (
	"context"
	"fmt"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/new-energy-monitoring/pkg/ai/forecast"
)

type ForecastService interface {
	PowerForecast(ctx context.Context, stationID string, forecastType entity.ForecastType) ([]*entity.ForecastResult, error)
	GetResults(ctx context.Context, stationID string, forecastType *entity.ForecastType, startTime, endTime *time.Time, page, pageSize int) ([]*entity.ForecastResult, int64, error)
	GetAccuracy(ctx context.Context, stationID string, forecastType *entity.ForecastType, start, end time.Time) (*repository.ForecastAccuracyStats, error)
	EvaluateModel(ctx context.Context, stationID string, forecastType *entity.ForecastType, start, end time.Time, installedCapacity float64) (*forecast.EvaluationReport, error)
	AttributionAnalysis(ctx context.Context, stationID string, targetTime string, predictedPower, actualPower float64) (*forecast.AttributionResult, error)
}

type forecastService struct {
	evaluator  forecast.ForecastEvaluator
	attributor forecast.AttributionAnalyzer
	resultRepo repository.ForecastResultRepository
}

func NewForecastService(
	evaluator forecast.ForecastEvaluator,
	attributor forecast.AttributionAnalyzer,
	resultRepo repository.ForecastResultRepository,
) ForecastService {
	return &forecastService{
		evaluator:  evaluator,
		attributor: attributor,
		resultRepo: resultRepo,
	}
}

func (s *forecastService) PowerForecast(ctx context.Context, stationID string, forecastType entity.ForecastType) ([]*entity.ForecastResult, error) {
	now := time.Now()
	var startTime time.Time
	switch forecastType {
	case entity.ForecastTypeUltraShortTerm:
		startTime = now.Add(-15 * time.Minute)
	case entity.ForecastTypeShortTerm:
		startTime = now.Add(-2 * time.Hour)
	case entity.ForecastTypeMediumTerm:
		startTime = now.AddDate(0, 0, -1)
	default:
		startTime = now.Add(-2 * time.Hour)
	}
	existing, _, err := s.resultRepo.ListByStation(ctx, stationID, &forecastType, &startTime, &now, 0, 100)
	if err == nil && len(existing) > 0 {
		return existing, nil
	}
	var intervals []time.Duration
	switch forecastType {
	case entity.ForecastTypeUltraShortTerm:
		intervals = []time.Duration{15 * time.Minute, 30 * time.Minute, 45 * time.Minute, 60 * time.Minute}
	case entity.ForecastTypeShortTerm:
		intervals = []time.Duration{1 * time.Hour, 2 * time.Hour, 3 * time.Hour, 6 * time.Hour, 12 * time.Hour, 24 * time.Hour}
	case entity.ForecastTypeMediumTerm:
		intervals = []time.Duration{24 * time.Hour, 48 * time.Hour, 72 * time.Hour, 96 * time.Hour, 120 * time.Hour, 144 * time.Hour, 168 * time.Hour}
	default:
		intervals = []time.Duration{1 * time.Hour, 2 * time.Hour, 3 * time.Hour}
	}
	results := make([]*entity.ForecastResult, 0, len(intervals))
	for i, offset := range intervals {
		targetTime := now.Add(offset)
		predictedPower := 0.0
		switch forecastType {
		case entity.ForecastTypeUltraShortTerm:
			predictedPower = 500.0 + float64(i)*50.0
		case entity.ForecastTypeShortTerm:
			predictedPower = 800.0 + float64(i)*30.0
		case entity.ForecastTypeMediumTerm:
			predictedPower = 1000.0 + float64(i)*20.0
		}
		confidence := 0.95 - float64(i)*0.05
		if confidence < 0.6 {
			confidence = 0.6
		}
		result := &entity.ForecastResult{
			StationID:      stationID,
			ForecastType:   forecastType,
			TargetTime:     targetTime,
			PredictedPower: predictedPower,
			Confidence:     &confidence,
		}
		result.SetConfidenceInterval(predictedPower*0.9, predictedPower*1.1)
		if err := s.resultRepo.Create(ctx, result); err != nil {
			return nil, fmt.Errorf("save forecast result failed: %w", err)
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *forecastService) GetResults(ctx context.Context, stationID string, forecastType *entity.ForecastType, startTime, endTime *time.Time, page, pageSize int) ([]*entity.ForecastResult, int64, error) {
	offset := (page - 1) * pageSize
	return s.resultRepo.ListByStation(ctx, stationID, forecastType, startTime, endTime, offset, pageSize)
}

func (s *forecastService) GetAccuracy(ctx context.Context, stationID string, forecastType *entity.ForecastType, start, end time.Time) (*repository.ForecastAccuracyStats, error) {
	return s.resultRepo.GetAccuracyStats(ctx, stationID, forecastType, start, end)
}

func (s *forecastService) EvaluateModel(ctx context.Context, stationID string, forecastType *entity.ForecastType, start, end time.Time, installedCapacity float64) (*forecast.EvaluationReport, error) {
	results, _, err := s.resultRepo.ListByStation(ctx, stationID, forecastType, &start, &end, 0, 10000)
	if err != nil {
		return nil, err
	}
	var predictions []forecast.PredictionWithActual
	for _, r := range results {
		if r.ActualPower != nil {
			predictions = append(predictions, forecast.PredictionWithActual{
				Predicted: r.PredictedPower,
				Actual:    *r.ActualPower,
				Time:      r.TargetTime,
			})
		}
	}
	return s.evaluator.Evaluate(predictions, installedCapacity)
}

func (s *forecastService) AttributionAnalysis(ctx context.Context, stationID string, targetTime string, predictedPower, actualPower float64) (*forecast.AttributionResult, error) {
	return s.attributor.Analyze(stationID, predictedPower, actualPower)
}
