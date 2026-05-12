package service

import (
	"context"
	"fmt"
	"math"
)

type AIStatisticsService interface {
	DetectAnomaly(ctx context.Context, stationID string, metricName string, values []float64) (*AnomalyResult, error)
	ForecastTrend(ctx context.Context, stationID string, metricName string, horizon int) (*TrendResult, error)
	SmartAggregate(ctx context.Context, stationID string, metricName string, granularity string) ([]*AggregatedPoint, error)
}

type AnomalyResult struct {
	StationID  string  `json:"station_id"`
	MetricName string  `json:"metric_name"`
	IsAnomaly  bool    `json:"is_anomaly"`
	Score      float64 `json:"score"`
	Threshold  float64 `json:"threshold"`
	Value      float64 `json:"value"`
	Mean       float64 `json:"mean"`
	StdDev     float64 `json:"std_dev"`
}

type TrendResult struct {
	StationID  string    `json:"station_id"`
	MetricName string    `json:"metric_name"`
	Direction  string    `json:"direction"`
	Slope      float64   `json:"slope"`
	Points     []float64 `json:"points"`
}

type AggregatedPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
	Count     int     `json:"count"`
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
	StdDev    float64 `json:"std_dev"`
}

type aiStatisticsService struct{}

func NewAIStatisticsService() AIStatisticsService {
	return &aiStatisticsService{}
}

func (s *aiStatisticsService) DetectAnomaly(ctx context.Context, stationID string, metricName string, values []float64) (*AnomalyResult, error) {
	if len(values) < 3 {
		return nil, fmt.Errorf("insufficient data points for anomaly detection")
	}
	mean, stdDev := calculateStats(values)
	threshold := mean + 2*stdDev
	lastValue := values[len(values)-1]
	score := 0.0
	if stdDev > 0 {
		score = math.Abs(lastValue-mean) / stdDev
	}
	return &AnomalyResult{
		StationID:  stationID,
		MetricName: metricName,
		IsAnomaly:  score > 2.0,
		Score:      score,
		Threshold:  threshold,
		Value:      lastValue,
		Mean:       mean,
		StdDev:     stdDev,
	}, nil
}

func (s *aiStatisticsService) ForecastTrend(ctx context.Context, stationID string, metricName string, horizon int) (*TrendResult, error) {
	return &TrendResult{
		StationID:  stationID,
		MetricName: metricName,
		Direction:  "stable",
		Slope:      0,
		Points:     make([]float64, horizon),
	}, nil
}

func (s *aiStatisticsService) SmartAggregate(ctx context.Context, stationID string, metricName string, granularity string) ([]*AggregatedPoint, error) {
	return []*AggregatedPoint{}, nil
}

func calculateStats(values []float64) (mean, stdDev float64) {
	n := float64(len(values))
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean = sum / n
	varSum := 0.0
	for _, v := range values {
		varSum += (v - mean) * (v - mean)
	}
	stdDev = math.Sqrt(varSum / n)
	return
}
