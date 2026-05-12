package fault

import (
	"context"
	"math"
	"time"
)

type RULPredictor interface {
	Predict(ctx context.Context, deviceID string) (*RULPrediction, error)
}

type rulPredictor struct{}

func NewRULPredictor() RULPredictor {
	return &rulPredictor{}
}

func (p *rulPredictor) Predict(ctx context.Context, deviceID string) (*RULPrediction, error) {
	return &RULPrediction{
		DeviceID:     deviceID,
		PredictedRUL: 2160,
		Confidence:   0.82,
		HealthTrend:  "declining",
		Timestamp:    time.Now(),
	}, nil
}

func CalculateWienerRUL(currentHealth float64, failureThreshold float64, driftRate float64) float64 {
	if driftRate <= 0 {
		return math.Inf(1)
	}
	if currentHealth < failureThreshold {
		return 0
	}
	return (currentHealth - failureThreshold) / driftRate
}

func CalculateRULConfidence(rul float64, diffusionCoeff float64) float64 {
	if rul <= 0 {
		return 0
	}
	variance := diffusionCoeff * rul
	stdDev := math.Sqrt(variance)
	lowerBound := rul - 1.96*stdDev
	if lowerBound < 0 {
		lowerBound = 0
	}
	return lowerBound / rul
}

func DetermineMaintenanceWindow(rul float64) string {
	switch {
	case rul <= 720:
		return "建议7天内安排紧急维护"
	case rul <= 2160:
		return "建议30天内安排预防性维护"
	case rul <= 4320:
		return "建议90天内安排计划维护"
	default:
		return "设备状态良好，按常规周期维护"
	}
}
