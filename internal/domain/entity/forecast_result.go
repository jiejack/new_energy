package entity

import "time"

type ForecastType string

const (
	ForecastTypeUltraShortTerm ForecastType = "ultra_short_term"
	ForecastTypeShortTerm      ForecastType = "short_term"
	ForecastTypeMediumTerm     ForecastType = "medium_term"
)

type ForecastResult struct {
	ID                string       `json:"id" gorm:"primaryKey;type:varchar(36)"`
	StationID         string       `json:"station_id" gorm:"type:varchar(36);not null;index"`
	ForecastType      ForecastType `json:"forecast_type" gorm:"type:varchar(20);not null"`
	TargetTime        time.Time    `json:"target_time" gorm:"not null;index"`
	PredictedPower    float64      `json:"predicted_power" gorm:"type:decimal(12,4);not null"`
	ActualPower       *float64     `json:"actual_power,omitempty" gorm:"type:decimal(12,4)"`
	Accuracy          *float64     `json:"accuracy,omitempty" gorm:"type:decimal(6,4)"`
	ConfidenceLower   *float64     `json:"confidence_lower,omitempty" gorm:"type:decimal(12,4)"`
	ConfidenceUpper   *float64     `json:"confidence_upper,omitempty" gorm:"type:decimal(12,4)"`
	ModelVersion      string       `json:"model_version" gorm:"type:varchar(50);not null"`
	AttributionType   *string      `json:"attribution_type,omitempty" gorm:"type:varchar(30)"`
	AttributionDetail *string      `json:"attribution_detail,omitempty" gorm:"type:text"`
	CreatedAt         time.Time    `json:"created_at"`
}

func (ForecastResult) TableName() string {
	return "forecast_results"
}

func NewForecastResult(stationID string, forecastType ForecastType, targetTime time.Time, predictedPower float64, modelVersion string) *ForecastResult {
	return &ForecastResult{
		StationID:      stationID,
		ForecastType:   forecastType,
		TargetTime:     targetTime,
		PredictedPower: predictedPower,
		ModelVersion:   modelVersion,
	}
}

func (f *ForecastResult) SetActualPower(actualPower float64) {
	f.ActualPower = &actualPower
}

func (f *ForecastResult) SetAccuracy(accuracy float64) {
	f.Accuracy = &accuracy
}

func (f *ForecastResult) SetConfidenceInterval(lower, upper float64) {
	f.ConfidenceLower = &lower
	f.ConfidenceUpper = &upper
}

func (f *ForecastResult) SetAttribution(attributionType, detail string) {
	f.AttributionType = &attributionType
	f.AttributionDetail = &detail
}
