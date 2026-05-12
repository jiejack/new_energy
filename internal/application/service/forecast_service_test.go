package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/new-energy-monitoring/pkg/ai/forecast"
)

var errTestRepoFailure = errors.New("repository failure")

type mockForecastResultRepository struct {
	createFn        func(ctx context.Context, result *entity.ForecastResult) error
	getByIDFn       func(ctx context.Context, id string) (*entity.ForecastResult, error)
	listByStationFn func(ctx context.Context, stationID string, forecastType *entity.ForecastType, startTime, endTime *time.Time, offset, limit int) ([]*entity.ForecastResult, int64, error)
	updateActualFn  func(ctx context.Context, id string, actualPower float64) error
	getAccuracyFn   func(ctx context.Context, stationID string, forecastType *entity.ForecastType, start, end time.Time) (*repository.ForecastAccuracyStats, error)
}

func (m *mockForecastResultRepository) Create(ctx context.Context, result *entity.ForecastResult) error {
	if m.createFn != nil {
		return m.createFn(ctx, result)
	}
	return nil
}

func (m *mockForecastResultRepository) GetByID(ctx context.Context, id string) (*entity.ForecastResult, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockForecastResultRepository) ListByStation(ctx context.Context, stationID string, forecastType *entity.ForecastType, startTime, endTime *time.Time, offset, limit int) ([]*entity.ForecastResult, int64, error) {
	if m.listByStationFn != nil {
		return m.listByStationFn(ctx, stationID, forecastType, startTime, endTime, offset, limit)
	}
	return nil, 0, nil
}

func (m *mockForecastResultRepository) UpdateActualPower(ctx context.Context, id string, actualPower float64) error {
	if m.updateActualFn != nil {
		return m.updateActualFn(ctx, id, actualPower)
	}
	return nil
}

func (m *mockForecastResultRepository) GetAccuracyStats(ctx context.Context, stationID string, forecastType *entity.ForecastType, start, end time.Time) (*repository.ForecastAccuracyStats, error) {
	if m.getAccuracyFn != nil {
		return m.getAccuracyFn(ctx, stationID, forecastType, start, end)
	}
	return nil, nil
}

type mockForecastEvaluator struct {
	evaluateFn         func(predictions []forecast.PredictionWithActual, installedCapacity float64) (*forecast.EvaluationReport, error)
	calculateAccuracyFn func(rmse, installedCapacity float64) float64
}

func (m *mockForecastEvaluator) Evaluate(predictions []forecast.PredictionWithActual, installedCapacity float64) (*forecast.EvaluationReport, error) {
	if m.evaluateFn != nil {
		return m.evaluateFn(predictions, installedCapacity)
	}
	return &forecast.EvaluationReport{}, nil
}

func (m *mockForecastEvaluator) CalculateAccuracy(rmse, installedCapacity float64) float64 {
	if m.calculateAccuracyFn != nil {
		return m.calculateAccuracyFn(rmse, installedCapacity)
	}
	return 0
}

type mockAttributionAnalyzer struct {
	analyzeFn func(stationID string, predictedPower, actualPower float64) (*forecast.AttributionResult, error)
}

func (m *mockAttributionAnalyzer) Analyze(stationID string, predictedPower, actualPower float64) (*forecast.AttributionResult, error) {
	if m.analyzeFn != nil {
		return m.analyzeFn(stationID, predictedPower, actualPower)
	}
	return &forecast.AttributionResult{}, nil
}

func TestForecastService_GetResults(t *testing.T) {
	now := time.Now()
	ft := entity.ForecastTypeShortTerm
	expectedResults := []*entity.ForecastResult{
		{
			StationID:      "station-001",
			ForecastType:   ft,
			TargetTime:     now.Add(1 * time.Hour),
			PredictedPower: 800.0,
		},
		{
			StationID:      "station-001",
			ForecastType:   ft,
			TargetTime:     now.Add(2 * time.Hour),
			PredictedPower: 830.0,
		},
	}

	repo := &mockForecastResultRepository{
		listByStationFn: func(ctx context.Context, stationID string, forecastType *entity.ForecastType, startTime, endTime *time.Time, offset, limit int) ([]*entity.ForecastResult, int64, error) {
			if stationID != "station-001" {
				t.Errorf("expected stationID station-001, got %s", stationID)
			}
			if offset != 0 {
				t.Errorf("expected offset 0, got %d", offset)
			}
			if limit != 10 {
				t.Errorf("expected limit 10, got %d", limit)
			}
			return expectedResults, 2, nil
		},
	}

	svc := NewForecastService(&mockForecastEvaluator{}, &mockAttributionAnalyzer{}, repo)
	results, total, err := svc.GetResults(context.Background(), "station-001", &ft, nil, nil, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestForecastService_GetResults_Pagination(t *testing.T) {
	ft := entity.ForecastTypeShortTerm
	repo := &mockForecastResultRepository{
		listByStationFn: func(ctx context.Context, stationID string, forecastType *entity.ForecastType, startTime, endTime *time.Time, offset, limit int) ([]*entity.ForecastResult, int64, error) {
			expectedOffset := (3 - 1) * 5
			if offset != expectedOffset {
				t.Errorf("expected offset %d, got %d", expectedOffset, offset)
			}
			if limit != 5 {
				t.Errorf("expected limit 5, got %d", limit)
			}
			return []*entity.ForecastResult{}, 25, nil
		},
	}

	svc := NewForecastService(&mockForecastEvaluator{}, &mockAttributionAnalyzer{}, repo)
	_, total, err := svc.GetResults(context.Background(), "station-001", &ft, nil, nil, 3, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 25 {
		t.Errorf("expected total 25, got %d", total)
	}
}

func TestForecastService_GetAccuracy(t *testing.T) {
	ft := entity.ForecastTypeShortTerm
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()
	expectedStats := &repository.ForecastAccuracyStats{
		StationID:    "station-001",
		ForecastType: "short_term",
		TotalPoints:  100,
		AvgAccuracy:  0.92,
		RMSE:         15.5,
		MAE:          12.3,
	}

	repo := &mockForecastResultRepository{
		getAccuracyFn: func(ctx context.Context, stationID string, forecastType *entity.ForecastType, s, e time.Time) (*repository.ForecastAccuracyStats, error) {
			if stationID != "station-001" {
				t.Errorf("expected stationID station-001, got %s", stationID)
			}
			return expectedStats, nil
		},
	}

	svc := NewForecastService(&mockForecastEvaluator{}, &mockAttributionAnalyzer{}, repo)
	stats, err := svc.GetAccuracy(context.Background(), "station-001", &ft, start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.AvgAccuracy != 0.92 {
		t.Errorf("expected avg accuracy 0.92, got %f", stats.AvgAccuracy)
	}
	if stats.TotalPoints != 100 {
		t.Errorf("expected total points 100, got %d", stats.TotalPoints)
	}
}

func TestForecastService_GetAccuracy_RepoError(t *testing.T) {
	ft := entity.ForecastTypeShortTerm
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()

	repo := &mockForecastResultRepository{
		getAccuracyFn: func(ctx context.Context, stationID string, forecastType *entity.ForecastType, s, e time.Time) (*repository.ForecastAccuracyStats, error) {
			return nil, errTestRepoFailure
		},
	}

	svc := NewForecastService(&mockForecastEvaluator{}, &mockAttributionAnalyzer{}, repo)
	_, err := svc.GetAccuracy(context.Background(), "station-001", &ft, start, end)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestForecastService_EvaluateModel(t *testing.T) {
	ft := entity.ForecastTypeShortTerm
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()
	actualPower := 780.0

	repo := &mockForecastResultRepository{
		listByStationFn: func(ctx context.Context, stationID string, forecastType *entity.ForecastType, startTime, endTime *time.Time, offset, limit int) ([]*entity.ForecastResult, int64, error) {
			return []*entity.ForecastResult{
				{
					PredictedPower: 800.0,
					ActualPower:    &actualPower,
					TargetTime:     start.Add(1 * time.Hour),
				},
				{
					PredictedPower: 830.0,
					ActualPower:    &actualPower,
					TargetTime:     start.Add(2 * time.Hour),
				},
			}, 2, nil
		},
	}

	evaluator := &mockForecastEvaluator{
		evaluateFn: func(predictions []forecast.PredictionWithActual, installedCapacity float64) (*forecast.EvaluationReport, error) {
			if len(predictions) != 2 {
				t.Errorf("expected 2 predictions, got %d", len(predictions))
			}
			if installedCapacity != 1000.0 {
				t.Errorf("expected installedCapacity 1000, got %f", installedCapacity)
			}
			return &forecast.EvaluationReport{
				TotalPoints:  2,
				RMSE:         25.0,
				MAE:          25.0,
				Accuracy:     0.975,
				InstalledCap: 1000.0,
			}, nil
		},
	}

	svc := NewForecastService(evaluator, &mockAttributionAnalyzer{}, repo)
	report, err := svc.EvaluateModel(context.Background(), "station-001", &ft, start, end, 1000.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.TotalPoints != 2 {
		t.Errorf("expected total points 2, got %d", report.TotalPoints)
	}
	if report.Accuracy != 0.975 {
		t.Errorf("expected accuracy 0.975, got %f", report.Accuracy)
	}
}

func TestForecastService_EvaluateModel_NoActualPower(t *testing.T) {
	ft := entity.ForecastTypeShortTerm
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()

	repo := &mockForecastResultRepository{
		listByStationFn: func(ctx context.Context, stationID string, forecastType *entity.ForecastType, startTime, endTime *time.Time, offset, limit int) ([]*entity.ForecastResult, int64, error) {
			return []*entity.ForecastResult{
				{
					PredictedPower: 800.0,
					ActualPower:    nil,
					TargetTime:     start.Add(1 * time.Hour),
				},
			}, 1, nil
		},
	}

	evaluator := &mockForecastEvaluator{
		evaluateFn: func(predictions []forecast.PredictionWithActual, installedCapacity float64) (*forecast.EvaluationReport, error) {
			if len(predictions) != 0 {
				t.Errorf("expected 0 predictions (no actual power), got %d", len(predictions))
			}
			return &forecast.EvaluationReport{}, nil
		},
	}

	svc := NewForecastService(evaluator, &mockAttributionAnalyzer{}, repo)
	report, err := svc.EvaluateModel(context.Background(), "station-001", &ft, start, end, 1000.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestForecastService_EvaluateModel_RepoError(t *testing.T) {
	ft := entity.ForecastTypeShortTerm
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()

	repo := &mockForecastResultRepository{
		listByStationFn: func(ctx context.Context, stationID string, forecastType *entity.ForecastType, startTime, endTime *time.Time, offset, limit int) ([]*entity.ForecastResult, int64, error) {
			return nil, 0, errTestRepoFailure
		},
	}

	svc := NewForecastService(&mockForecastEvaluator{}, &mockAttributionAnalyzer{}, repo)
	_, err := svc.EvaluateModel(context.Background(), "station-001", &ft, start, end, 1000.0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}


