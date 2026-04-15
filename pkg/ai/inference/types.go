package inference

import (
	"context"
	"time"
)

type ModelType string
type InferenceType string
type JobStatus string

const (
	ModelTypeSolarForecast  ModelType = "solar_forecast"
	ModelTypeWindForecast   ModelType = "wind_forecast"
	ModelTypeFaultDetector  ModelType = "fault_detector"
	ModelTypeHealthScore  ModelType = "health_score"

	InferenceTypeRealtime InferenceType = "realtime"
	InferenceTypeBatch    InferenceType = "batch"

	JobStatusQueued     JobStatus = "queued"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

type PredictRequest struct {
	ModelID   string                 `json:"model_id"`
	Version   string                 `json:"version"`
	Inputs    map[string]interface{} `json:"inputs"`
	Options   PredictOptions         `json:"options"`
}

type PredictOptions struct {
	IncludeConfidence bool `json:"include_confidence"`
	IncludeExplanation bool `json:"include_explanation"`
	CacheTTLSeconds   int  `json:"cache_ttl_seconds"`
}

type PredictResponse struct {
	RequestID   string          `json:"request_id"`
	ModelID     string          `json:"model_id"`
	Version     string          `json:"version"`
	Prediction  Prediction      `json:"prediction"`
	Explanation *Explanation    `json:"explanation,omitempty"`
	Metadata    InferenceMetadata `json:"metadata"`
}

type Prediction struct {
	Value               float64   `json:"value"`
	Unit                string    `json:"unit"`
	Confidence          float64   `json:"confidence,omitempty"`
	ConfidenceInterval  []float64 `json:"confidence_interval,omitempty"`
}

type Explanation struct {
	FeatureImportance map[string]float64 `json:"feature_importance,omitempty"`
	SHAPValues        map[string]float64 `json:"shap_values,omitempty"`
	LIMEExplanation   string             `json:"lime_explanation,omitempty"`
}

type InferenceMetadata struct {
	InferenceTimeMs int64     `json:"inference_time_ms"`
	Cached          bool      `json:"cached"`
	Timestamp       time.Time `json:"timestamp"`
}

type BatchPredictRequest struct {
	ModelID     string                   `json:"model_id"`
	Version     string                   `json:"version"`
	BatchID     string                   `json:"batch_id"`
	Inputs      []map[string]interface{} `json:"inputs"`
	CallbackURL string                   `json:"callback_url"`
	Options     PredictOptions           `json:"options"`
}

type BatchPredictResponse struct {
	JobID               string `json:"job_id"`
	Status              JobStatus `json:"status"`
	EstimatedTimeSeconds int    `json:"estimated_time_seconds"`
	QueuePosition        int    `json:"queue_position"`
}

type BatchJobStatus struct {
	JobID        string     `json:"job_id"`
	Status       JobStatus  `json:"status"`
	Progress     int        `json:"progress"`
	ResultURL    string     `json:"result_url,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	ItemCount    int        `json:"item_count"`
	ErrorMessage string     `json:"error_message,omitempty"`
}

type ModelInfo struct {
	ModelID      string    `json:"model_id"`
	Version      string    `json:"version"`
	Type         ModelType `json:"type"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Status       string    `json:"status"`
	Metrics      map[string]float64 `json:"metrics"`
}

type InferenceService interface {
	Predict(ctx context.Context, req *PredictRequest) (*PredictResponse, error)
	BatchPredict(ctx context.Context, req *BatchPredictRequest) (*BatchPredictResponse, error)
	GetBatchJobStatus(ctx context.Context, jobID string) (*BatchJobStatus, error)
	ListModels(ctx context.Context) ([]*ModelInfo, error)
	GetModel(ctx context.Context, modelID string) (*ModelInfo, error)
}

type CacheStrategy interface {
	Get(ctx context.Context, key string) (*PredictResponse, error)
	Set(ctx context.Context, key string, resp *PredictResponse, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeleteByPattern(ctx context.Context, pattern string) error
}

type ModelManager interface {
	LoadModel(ctx context.Context, modelID, version string) error
	UnloadModel(ctx context.Context, modelID string) error
	GetModelInfo(ctx context.Context, modelID string) (*ModelInfo, error)
	ListModels(ctx context.Context) ([]*ModelInfo, error)
	Predict(ctx context.Context, modelID string, inputs map[string]interface{}) (*Prediction, error)
}
