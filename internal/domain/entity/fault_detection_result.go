package entity

import "time"

type FaultSeverity string

const (
	FaultSeverityInfo     FaultSeverity = "info"
	FaultSeverityWarning  FaultSeverity = "warning"
	FaultSeverityCritical FaultSeverity = "critical"
	FaultSeverityFatal    FaultSeverity = "fatal"
)

type FaultDetectionStatus int

const (
	FaultDetectionStatusPending   FaultDetectionStatus = 1
	FaultDetectionStatusConfirmed FaultDetectionStatus = 2
	FaultDetectionStatusResolved  FaultDetectionStatus = 3
	FaultDetectionStatusIgnored   FaultDetectionStatus = 4
)

type FaultDetectionResult struct {
	ID                    string                `json:"id" gorm:"primaryKey;type:varchar(36)"`
	DeviceID              string                `json:"device_id" gorm:"type:varchar(36);not null;index"`
	FaultType             string                `json:"fault_type" gorm:"type:varchar(50);not null"`
	Severity              FaultSeverity         `json:"severity" gorm:"type:varchar(20);not null"`
	Confidence            float64               `json:"confidence" gorm:"type:decimal(6,4);not null"`
	Description           string                `json:"description" gorm:"type:text"`
	Status                FaultDetectionStatus  `json:"status" gorm:"type:int;default:1"`
	RootCause             *string               `json:"root_cause,omitempty" gorm:"type:text"`
	Recommendation        *string               `json:"recommendation,omitempty" gorm:"type:text"`
	RemainingUsefulLifeHrs *int                 `json:"remaining_useful_life_hrs,omitempty"`
	HealthScore           *int                  `json:"health_score,omitempty"`
	WorkOrderID           *string               `json:"work_order_id,omitempty" gorm:"type:varchar(36)"`
	ModelVersion          string                `json:"model_version" gorm:"type:varchar(50);not null"`
	CreatedAt             time.Time             `json:"created_at"`
	UpdatedAt             time.Time             `json:"updated_at"`
}

func (FaultDetectionResult) TableName() string {
	return "fault_detection_results"
}

func NewFaultDetectionResult(deviceID, faultType string, severity FaultSeverity, confidence float64, description, modelVersion string) *FaultDetectionResult {
	return &FaultDetectionResult{
		DeviceID:     deviceID,
		FaultType:    faultType,
		Severity:     severity,
		Confidence:   confidence,
		Description:  description,
		Status:       FaultDetectionStatusPending,
		ModelVersion: modelVersion,
	}
}

func (f *FaultDetectionResult) Confirm() {
	f.Status = FaultDetectionStatusConfirmed
}

func (f *FaultDetectionResult) Resolve() {
	f.Status = FaultDetectionStatusResolved
}

func (f *FaultDetectionResult) Ignore() {
	f.Status = FaultDetectionStatusIgnored
}

func (f *FaultDetectionResult) SetRootCause(cause string) {
	f.RootCause = &cause
}

func (f *FaultDetectionResult) SetRecommendation(rec string) {
	f.Recommendation = &rec
}

func (f *FaultDetectionResult) SetRUL(hours int) {
	f.RemainingUsefulLifeHrs = &hours
}

func (f *FaultDetectionResult) SetHealthScore(score int) {
	f.HealthScore = &score
}

func (f *FaultDetectionResult) LinkWorkOrder(workOrderID string) {
	f.WorkOrderID = &workOrderID
}
