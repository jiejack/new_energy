package service

import (
	"context"
	"fmt"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type DataLoopService interface {
	CollectFeedback(ctx context.Context, stationID string, forecastType entity.ForecastType, targetTime time.Time) (*FeedbackReport, error)
	EvaluateAndTrigger(ctx context.Context, stationID string) (*RetrainDecision, error)
	GetLoopStatus(ctx context.Context, stationID string) (*LoopStatus, error)
}

type FeedbackReport struct {
	StationID     string  `json:"station_id"`
	ForecastType  string  `json:"forecast_type"`
	TotalPoints   int     `json:"total_points"`
	MatchedPoints int     `json:"matched_points"`
	AvgError      float64 `json:"avg_error"`
	DriftDetected bool    `json:"drift_detected"`
}

type RetrainDecision struct {
	StationID     string  `json:"station_id"`
	ShouldRetrain bool    `json:"should_retrain"`
	Reason        string  `json:"reason"`
	Accuracy      float64 `json:"current_accuracy"`
	Threshold     float64 `json:"threshold"`
}

type LoopStatus struct {
	StationID       string    `json:"station_id"`
	LastFeedbackAt  time.Time `json:"last_feedback_at"`
	LastRetrainAt   time.Time `json:"last_retrain_at"`
	CurrentAccuracy float64   `json:"current_accuracy"`
	LoopCount       int       `json:"loop_count"`
}

type dataLoopService struct {
	forecastRepo repository.ForecastResultRepository
	modelSvc     ModelService
}

func NewDataLoopService(
	forecastRepo repository.ForecastResultRepository,
	modelSvc ModelService,
) DataLoopService {
	return &dataLoopService{
		forecastRepo: forecastRepo,
		modelSvc:     modelSvc,
	}
}

func (s *dataLoopService) CollectFeedback(ctx context.Context, stationID string, forecastType entity.ForecastType, targetTime time.Time) (*FeedbackReport, error) {
	stats, err := s.forecastRepo.GetAccuracyStats(ctx, stationID, &forecastType, targetTime.AddDate(0, -1, 0), targetTime)
	if err != nil {
		return nil, fmt.Errorf("get accuracy stats failed: %w", err)
	}
	report := &FeedbackReport{
		StationID:     stationID,
		ForecastType:  string(forecastType),
		TotalPoints:   int(stats.TotalPoints),
		MatchedPoints: int(stats.TotalPoints),
		AvgError:      stats.RMSE,
		DriftDetected: stats.AvgAccuracy < 0.80,
	}
	return report, nil
}

func (s *dataLoopService) EvaluateAndTrigger(ctx context.Context, stationID string) (*RetrainDecision, error) {
	ft := entity.ForecastTypeShortTerm
	stats, err := s.forecastRepo.GetAccuracyStats(ctx, stationID, &ft, time.Now().AddDate(0, -1, 0), time.Now())
	if err != nil {
		return &RetrainDecision{
			StationID:     stationID,
			ShouldRetrain: false,
			Reason:        "无法获取准确率统计",
		}, nil
	}
	threshold := 0.85
	decision := &RetrainDecision{
		StationID: stationID,
		Accuracy:  stats.AvgAccuracy,
		Threshold: threshold,
	}
	if stats.AvgAccuracy < threshold {
		decision.ShouldRetrain = true
		decision.Reason = fmt.Sprintf("准确率 %.2f%% 低于阈值 %.0f%%，建议触发重训练", stats.AvgAccuracy*100, threshold*100)
	} else {
		decision.ShouldRetrain = false
		decision.Reason = fmt.Sprintf("准确率 %.2f%% 满足阈值要求", stats.AvgAccuracy*100)
	}
	return decision, nil
}

func (s *dataLoopService) GetLoopStatus(ctx context.Context, stationID string) (*LoopStatus, error) {
	ft := entity.ForecastTypeShortTerm
	stats, err := s.forecastRepo.GetAccuracyStats(ctx, stationID, &ft, time.Now().AddDate(0, -1, 0), time.Now())
	accuracy := 0.0
	if err == nil && stats != nil {
		accuracy = stats.AvgAccuracy
	}
	return &LoopStatus{
		StationID:       stationID,
		LastFeedbackAt:  time.Now(),
		CurrentAccuracy: accuracy,
		LoopCount:       0,
	}, nil
}
