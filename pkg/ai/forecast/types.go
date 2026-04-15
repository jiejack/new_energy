package forecast

import (
	"context"
	"time"
)

type Forecaster interface {
	Train(ctx context.Context, data []*TimeSeriesData) error
	Predict(ctx context.Context, horizon int) ([]*Prediction, error)
	GetModelInfo() *ModelInfo
	Save(ctx context.Context, path string) error
	Load(ctx context.Context, path string) error
}

type TimeSeriesData struct {
	Timestamp time.Time              `json:"timestamp"`
	Value     float64                `json:"value"`
	Features  map[string]float64     `json:"features,omitempty"`
}

type Prediction struct {
	Timestamp          time.Time `json:"timestamp"`
	Value              float64   `json:"value"`
	Confidence         float64   `json:"confidence,omitempty"`
	ConfidenceInterval [2]float64 `json:"confidence_interval,omitempty"`
}

type ModelInfo struct {
	ModelID      string                 `json:"model_id"`
	ModelType    string                 `json:"model_type"`
	Version      string                 `json:"version"`
	CreatedAt    time.Time              `json:"created_at"`
	TrainedAt    *time.Time             `json:"trained_at,omitempty"`
	Parameters   map[string]interface{} `json:"parameters"`
	Metrics      map[string]float64     `json:"metrics,omitempty"`
	Status       string                 `json:"status"`
}

type EvaluationMetrics struct {
	MAE  float64 `json:"mae"`
	MAPE float64 `json:"mape"`
	RMSE float64 `json:"rmse"`
	R2   float64 `json:"r2"`
}
