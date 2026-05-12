package service

import (
	"context"
	"testing"
)

func newTestAIStatisticsService() AIStatisticsService {
	return NewAIStatisticsService()
}

func TestAIStatisticsService_DetectAnomaly_AnomalyCase(t *testing.T) {
	svc := newTestAIStatisticsService()
	values := []float64{100, 102, 98, 101, 99, 100, 103, 97, 101, 200}
	result, err := svc.DetectAnomaly(context.Background(), "station-001", "power", values)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.StationID != "station-001" {
		t.Errorf("expected stationID station-001, got %s", result.StationID)
	}
	if result.MetricName != "power" {
		t.Errorf("expected metricName power, got %s", result.MetricName)
	}
	if !result.IsAnomaly {
		t.Error("expected IsAnomaly to be true for outlier value 200")
	}
	if result.Score <= 2.0 {
		t.Errorf("expected anomaly score > 2.0, got %f", result.Score)
	}
	if result.Value != 200.0 {
		t.Errorf("expected last value 200.0, got %f", result.Value)
	}
}

func TestAIStatisticsService_DetectAnomaly_NormalCase(t *testing.T) {
	svc := newTestAIStatisticsService()
	values := []float64{100, 102, 98, 101, 99, 100, 103, 97, 101, 100}
	result, err := svc.DetectAnomaly(context.Background(), "station-001", "power", values)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsAnomaly {
		t.Error("expected IsAnomaly to be false for normal data")
	}
	if result.Score > 2.0 {
		t.Errorf("expected score <= 2.0 for normal data, got %f", result.Score)
	}
}

func TestAIStatisticsService_DetectAnomaly_InsufficientData(t *testing.T) {
	svc := newTestAIStatisticsService()
	values := []float64{100, 50}
	_, err := svc.DetectAnomaly(context.Background(), "station-001", "power", values)
	if err == nil {
		t.Fatal("expected error for insufficient data points, got nil")
	}
}

func TestAIStatisticsService_DetectAnomaly_EmptyData(t *testing.T) {
	svc := newTestAIStatisticsService()
	values := []float64{}
	_, err := svc.DetectAnomaly(context.Background(), "station-001", "power", values)
	if err == nil {
		t.Fatal("expected error for empty data points, got nil")
	}
}

func TestAIStatisticsService_DetectAnomaly_StatsCalculation(t *testing.T) {
	svc := newTestAIStatisticsService()
	values := []float64{10, 10, 10, 10}
	result, err := svc.DetectAnomaly(context.Background(), "station-001", "power", values)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Mean != 10.0 {
		t.Errorf("expected mean 10.0, got %f", result.Mean)
	}
	if result.StdDev != 0 {
		t.Errorf("expected stdDev 0 for constant values, got %f", result.StdDev)
	}
	if result.Score != 0 {
		t.Errorf("expected score 0 when stdDev is 0, got %f", result.Score)
	}
}

func TestAIStatisticsService_ForecastTrend_ValidHorizon(t *testing.T) {
	svc := newTestAIStatisticsService()
	result, err := svc.ForecastTrend(context.Background(), "station-001", "power", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.StationID != "station-001" {
		t.Errorf("expected stationID station-001, got %s", result.StationID)
	}
	if result.MetricName != "power" {
		t.Errorf("expected metricName power, got %s", result.MetricName)
	}
	if len(result.Points) != 20 {
		t.Errorf("expected 20 points, got %d", len(result.Points))
	}
	if result.Direction != "increasing" {
		t.Errorf("expected direction increasing, got %s", result.Direction)
	}
	if result.Slope <= 0 {
		t.Errorf("expected positive slope, got %f", result.Slope)
	}
}

func TestAIStatisticsService_ForecastTrend_ZeroHorizon(t *testing.T) {
	svc := newTestAIStatisticsService()
	result, err := svc.ForecastTrend(context.Background(), "station-001", "power", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Points) != 10 {
		t.Errorf("expected 10 points (default for zero horizon), got %d", len(result.Points))
	}
}

func TestAIStatisticsService_ForecastTrend_NegativeHorizon(t *testing.T) {
	svc := newTestAIStatisticsService()
	result, err := svc.ForecastTrend(context.Background(), "station-001", "power", -5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Points) != 10 {
		t.Errorf("expected 10 points (default for negative horizon), got %d", len(result.Points))
	}
}

func TestAIStatisticsService_ForecastTrend_LargeHorizon(t *testing.T) {
	svc := newTestAIStatisticsService()
	result, err := svc.ForecastTrend(context.Background(), "station-001", "power", 200)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Points) != 100 {
		t.Errorf("expected 100 points (capped at 100), got %d", len(result.Points))
	}
}

func TestAIStatisticsService_SmartAggregate_Hourly(t *testing.T) {
	svc := newTestAIStatisticsService()
	points, err := svc.SmartAggregate(context.Background(), "station-001", "power", "1h")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(points) != 24 {
		t.Errorf("expected 24 points, got %d", len(points))
	}
	for i, p := range points {
		if p.Count != 60 {
			t.Errorf("point %d: expected count 60, got %d", i, p.Count)
		}
		if p.Min > p.Value {
			t.Errorf("point %d: min %f should be <= value %f", i, p.Min, p.Value)
		}
		if p.Max < p.Value {
			t.Errorf("point %d: max %f should be >= value %f", i, p.Max, p.Value)
		}
	}
}

func TestAIStatisticsService_SmartAggregate_MinuteGranularity(t *testing.T) {
	svc := newTestAIStatisticsService()
	points, err := svc.SmartAggregate(context.Background(), "station-001", "power", "1m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(points) != 24 {
		t.Errorf("expected 24 points, got %d", len(points))
	}
}

func TestAIStatisticsService_SmartAggregate_DailyGranularity(t *testing.T) {
	svc := newTestAIStatisticsService()
	points, err := svc.SmartAggregate(context.Background(), "station-001", "power", "1d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(points) != 24 {
		t.Errorf("expected 24 points, got %d", len(points))
	}
}

func TestAIStatisticsService_SmartAggregate_DefaultGranularity(t *testing.T) {
	svc := newTestAIStatisticsService()
	points, err := svc.SmartAggregate(context.Background(), "station-001", "power", "unknown")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(points) != 24 {
		t.Errorf("expected 24 points for default granularity, got %d", len(points))
	}
}

func TestAIStatisticsService_SmartAggregate_TimestampsOrdered(t *testing.T) {
	svc := newTestAIStatisticsService()
	points, err := svc.SmartAggregate(context.Background(), "station-001", "power", "5m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 1; i < len(points); i++ {
		if points[i].Timestamp <= points[i-1].Timestamp {
			t.Errorf("point %d timestamp %d should be > point %d timestamp %d", i, points[i].Timestamp, i-1, points[i-1].Timestamp)
		}
	}
}
