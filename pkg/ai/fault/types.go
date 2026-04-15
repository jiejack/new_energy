package fault

import (
	"context"
	"time"
)

type FaultDetector interface {
	Detect(ctx context.Context, data []*TimeSeriesData) ([]*Anomaly, error)
	Train(ctx context.Context, data []*TimeSeriesData) error
	GetDetectorInfo() *DetectorInfo
}

type FaultClassifier interface {
	Classify(ctx context.Context, anomaly *Anomaly) (*FaultClassification, error)
	Train(ctx context.Context, data []*FaultLabeledData) error
	GetClassifierInfo() *ClassifierInfo
}

type HealthAssessor interface {
	Assess(ctx context.Context, data []*TimeSeriesData) (*HealthAssessment, error)
	PredictRUL(ctx context.Context, data []*TimeSeriesData) (*RULPrediction, error)
	GetAssessorInfo() *AssessorInfo
}

type TimeSeriesData struct {
	Timestamp time.Time              `json:"timestamp"`
	Value     float64                `json:"value"`
	Features  map[string]float64     `json:"features,omitempty"`
	DeviceID  string                 `json:"device_id"`
	Metric    string                 `json:"metric"`
}

type Anomaly struct {
	ID             string                 `json:"id"`
	DeviceID       string                 `json:"device_id"`
	Timestamp      time.Time              `json:"timestamp"`
	Metric         string                 `json:"metric"`
	Value          float64                `json:"value"`
	ExpectedValue  float64                `json:"expected_value"`
	Deviation      float64                `json:"deviation"`
	Severity       AnomalySeverity        `json:"severity"`
	DetectorName   string                 `json:"detector_name"`
	Confidence     float64                `json:"confidence"`
	AdditionalInfo map[string]interface{} `json:"additional_info,omitempty"`
}

type AnomalySeverity string

const (
	SeverityLow      AnomalySeverity = "low"
	SeverityMedium   AnomalySeverity = "medium"
	SeverityHigh     AnomalySeverity = "high"
	SeverityCritical AnomalySeverity = "critical"
)

type FaultClassification struct {
	ID           string                 `json:"id"`
	AnomalyID    string                 `json:"anomaly_id"`
	FaultType    string                 `json:"fault_type"`
	FaultCode    string                 `json:"fault_code"`
	Description  string                 `json:"description"`
	Confidence   float64                `json:"confidence"`
	Recommendations []string               `json:"recommendations"`
	AdditionalInfo map[string]interface{} `json:"additional_info,omitempty"`
}

type HealthAssessment struct {
	DeviceID       string                 `json:"device_id"`
	Timestamp      time.Time              `json:"timestamp"`
	HealthScore    float64                `json:"health_score"`
	HealthStatus   HealthStatus           `json:"health_status"`
	ComponentHealth map[string]float64     `json:"component_health,omitempty"`
	Issues         []string               `json:"issues,omitempty"`
	Confidence     float64                `json:"confidence"`
	AdditionalInfo map[string]interface{} `json:"additional_info,omitempty"`
}

type HealthStatus string

const (
	HealthStatusExcellent HealthStatus = "excellent"
	HealthStatusGood      HealthStatus = "good"
	HealthStatusFair      HealthStatus = "fair"
	HealthStatusPoor      HealthStatus = "poor"
	HealthStatusCritical  HealthStatus = "critical"
)

type RULPrediction struct {
	DeviceID       string                 `json:"device_id"`
	Timestamp      time.Time              `json:"timestamp"`
	PredictedRUL   float64                `json:"predicted_rul"` // 剩余使用寿命（小时）
	Confidence     float64                `json:"confidence"`
	RULInterval    [2]float64            `json:"rul_interval"`  // 置信区间
	HealthTrend    string                 `json:"health_trend"`  // "improving", "stable", "declining"
	AdditionalInfo map[string]interface{} `json:"additional_info,omitempty"`
}

type FaultLabeledData struct {
	Anomaly    *Anomaly          `json:"anomaly"`
	FaultType  string             `json:"fault_type"`
	FaultCode  string             `json:"fault_code"`
	Label      bool               `json:"label"`
	Timestamp  time.Time          `json:"timestamp"`
	Features   map[string]float64 `json:"features"`
}

type DetectorInfo struct {
	DetectorID    string                 `json:"detector_id"`
	DetectorType  string                 `json:"detector_type"`
	Version       string                 `json:"version"`
	CreatedAt     time.Time              `json:"created_at"`
	TrainedAt     *time.Time             `json:"trained_at,omitempty"`
	Parameters    map[string]interface{} `json:"parameters"`
	Metrics       map[string]float64     `json:"metrics,omitempty"`
	Status        string                 `json:"status"`
}

type ClassifierInfo struct {
	ClassifierID  string                 `json:"classifier_id"`
	ClassifierType string                `json:"classifier_type"`
	Version       string                 `json:"version"`
	CreatedAt     time.Time              `json:"created_at"`
	TrainedAt     *time.Time             `json:"trained_at,omitempty"`
	Parameters    map[string]interface{} `json:"parameters"`
	Metrics       map[string]float64     `json:"metrics,omitempty"`
	Status        string                 `json:"status"`
}

type AssessorInfo struct {
	AssessorID    string                 `json:"assessor_id"`
	AssessorType  string                 `json:"assessor_type"`
	Version       string                 `json:"version"`
	CreatedAt     time.Time              `json:"created_at"`
	TrainedAt     *time.Time             `json:"trained_at,omitempty"`
	Parameters    map[string]interface{} `json:"parameters"`
	Metrics       map[string]float64     `json:"metrics,omitempty"`
	Status        string                 `json:"status"`
}

type FaultEvent struct {
	ID             string                 `json:"id"`
	DeviceID       string                 `json:"device_id"`
	Anomaly        *Anomaly               `json:"anomaly"`
	Classification *FaultClassification   `json:"classification,omitempty"`
	Assessment     *HealthAssessment      `json:"assessment,omitempty"`
	RUL            *RULPrediction        `json:"rul,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
	Status         string                 `json:"status"`
	Actions        []string               `json:"actions,omitempty"`
	AdditionalInfo map[string]interface{} `json:"additional_info,omitempty"`
}

type FaultService interface {
	DetectAnomalies(ctx context.Context, data []*TimeSeriesData) ([]*Anomaly, error)
	ClassifyFaults(ctx context.Context, anomalies []*Anomaly) ([]*FaultClassification, error)
	AssessHealth(ctx context.Context, deviceID string) (*HealthAssessment, error)
	PredictRUL(ctx context.Context, deviceID string) (*RULPrediction, error)
	GetFaultEvents(ctx context.Context, deviceID string, startTime, endTime time.Time) ([]*FaultEvent, error)
}
