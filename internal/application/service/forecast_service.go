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
	return nil, fmt.Errorf("power forecast requires AI model integration, not yet connected")
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
