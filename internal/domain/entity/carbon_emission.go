package entity

import (
	"time"
)

type CarbonEmissionScope string
type CarbonEmissionStatus string

const (
	CarbonEmissionScope1 CarbonEmissionScope = "scope1"
	CarbonEmissionScope2 CarbonEmissionScope = "scope2"
	CarbonEmissionScope3 CarbonEmissionScope = "scope3"
)

const (
	CarbonEmissionStatusDraft     CarbonEmissionStatus = "draft"
	CarbonEmissionStatusPending   CarbonEmissionStatus = "pending"
	CarbonEmissionStatusApproved  CarbonEmissionStatus = "approved"
	CarbonEmissionStatusRejected  CarbonEmissionStatus = "rejected"
)

type CarbonEmissionFactor struct {
	ID          string               `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Name        string               `json:"name" gorm:"type:varchar(200);not null"`
	Code        string               `json:"code" gorm:"type:varchar(100);not null;uniqueIndex"`
	Scope       CarbonEmissionScope  `json:"scope" gorm:"type:varchar(20);not null;index:idx_cef_scope"`
	Source      string               `json:"source" gorm:"type:varchar(200);not null"`
	Value       float64              `json:"value" gorm:"type:decimal(20,6);not null"`
	Unit        string               `json:"unit" gorm:"type:varchar(50);not null"`
	Description string               `json:"description" gorm:"type:text"`
	Version     string               `json:"version" gorm:"type:varchar(50);not null"`
	EffectiveAt time.Time            `json:"effective_at" gorm:"type:timestamp;not null"`
	ExpiresAt   *time.Time           `json:"expires_at" gorm:"type:timestamp"`
	IsActive    bool                 `json:"is_active" gorm:"default:true;index:idx_cef_active"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:text;serializer:json"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

func (f *CarbonEmissionFactor) TableName() string {
	return "carbon_emission_factors"
}

type CarbonEmissionRecord struct {
	ID             string                 `json:"id" gorm:"primaryKey;type:varchar(36)"`
	RecordTime     time.Time              `json:"record_time" gorm:"type:timestamp;not null;index:idx_cer_record_time"`
	Scope          CarbonEmissionScope    `json:"scope" gorm:"type:varchar(20);not null;index:idx_cer_scope"`
	TargetID       string                 `json:"target_id" gorm:"type:varchar(36);not null;index:idx_cer_target"`
	TargetName     string                 `json:"target_name" gorm:"type:varchar(200);not null"`
	FactorID       string                 `json:"factor_id" gorm:"type:varchar(36);index:idx_cer_factor"`
	FactorCode     string                 `json:"factor_code" gorm:"type:varchar(100);not null"`
	FactorValue    float64                `json:"factor_value" gorm:"type:decimal(20,6);not null"`
	ActivityData   float64                `json:"activity_data" gorm:"type:decimal(20,4);not null"`
	ActivityUnit   string                 `json:"activity_unit" gorm:"type:varchar(50);not null"`
	EmissionValue  float64                `json:"emission_value" gorm:"type:decimal(20,6);not null;index:idx_cer_emission"`
	EmissionUnit   string                 `json:"emission_unit" gorm:"type:varchar(50);default:'tCO2e'"`
	Period         string                 `json:"period" gorm:"type:varchar(20);not null;index:idx_cer_period"`
	Status         CarbonEmissionStatus   `json:"status" gorm:"type:varchar(20);default:'draft';index:idx_cer_status"`
	Remark         string                 `json:"remark" gorm:"type:text"`
	Metadata       map[string]interface{} `json:"metadata" gorm:"type:text;serializer:json"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	CreatedBy      string                 `json:"created_by" gorm:"type:varchar(100)"`
	UpdatedBy      string                 `json:"updated_by" gorm:"type:varchar(100)"`
}

func (r *CarbonEmissionRecord) TableName() string {
	return "carbon_emission_records"
}

type CarbonEmissionSummary struct {
	ID               string                 `json:"id" gorm:"primaryKey;type:varchar(36)"`
	SummaryTime      time.Time              `json:"summary_time" gorm:"type:timestamp;not null;index:idx_ces_summary_time"`
	TargetID         string                 `json:"target_id" gorm:"type:varchar(36);not null;index:idx_ces_target"`
	TargetName       string                 `json:"target_name" gorm:"type:varchar(200);not null"`
	Period           string                 `json:"period" gorm:"type:varchar(20);not null;index:idx_ces_period"`
	PeriodStart      time.Time              `json:"period_start" gorm:"type:timestamp;not null"`
	PeriodEnd        time.Time              `json:"period_end" gorm:"type:timestamp;not null"`
	
	Scope1Emission   float64               `json:"scope1_emission" gorm:"type:decimal(20,6);default:0"`
	Scope2Emission   float64               `json:"scope2_emission" gorm:"type:decimal(20,6);default:0"`
	Scope3Emission   float64               `json:"scope3_emission" gorm:"type:decimal(20,6);default:0"`
	TotalEmission    float64               `json:"total_emission" gorm:"type:decimal(20,6);not null;index:idx_ces_total"`
	
	EmissionIntensity float64              `json:"emission_intensity" gorm:"type:decimal(20,6)"`
	IntensityUnit     string               `json:"intensity_unit" gorm:"type:varchar(50)"`
	
	YoYChange         float64               `json:"yoy_change" gorm:"type:decimal(10,4)"`
	MoMChange         float64               `json:"mom_change" gorm:"type:decimal(10,4)"`
	
	Trend            string                `json:"trend" gorm:"type:varchar(20)"`
	Status           CarbonEmissionStatus  `json:"status" gorm:"type:varchar(20);default:'draft'"`
	Metadata         map[string]interface{} `json:"metadata" gorm:"type:text;serializer:json"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
}

func (s *CarbonEmissionSummary) TableName() string {
	return "carbon_emission_summaries"
}

type CarbonReductionTarget struct {
	ID               string                 `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Name             string                 `json:"name" gorm:"type:varchar(200);not null"`
	Description      string                 `json:"description" gorm:"type:text"`
	TargetID         string                 `json:"target_id" gorm:"type:varchar(36);not null;index:idx_crt_target"`
	TargetName       string                 `json:"target_name" gorm:"type:varchar(200);not null"`
	
	BaseYear         int                    `json:"base_year" gorm:"not null"`
	BaseEmission     float64                `json:"base_emission" gorm:"type:decimal(20,6);not null"`
	
	TargetYear       int                    `json:"target_year" gorm:"not null"`
	TargetReduction  float64                `json:"target_reduction" gorm:"type:decimal(10,4);not null"`
	TargetEmission   float64                `json:"target_emission" gorm:"type:decimal(20,6);not null"`
	
	CurrentProgress  float64                `json:"current_progress" gorm:"type:decimal(10,4);default:0"`
	CurrentEmission  float64                `json:"current_emission" gorm:"type:decimal(20,6);default:0"`
	
	StartDate        time.Time              `json:"start_date" gorm:"type:timestamp;not null"`
	EndDate          time.Time              `json:"end_date" gorm:"type:timestamp;not null"`
	
	Status           string                 `json:"status" gorm:"type:varchar(20);default:'active';index:idx_crt_status"`
	Metadata         map[string]interface{} `json:"metadata" gorm:"type:text;serializer:json"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
	CreatedBy        string                 `json:"created_by" gorm:"type:varchar(100)"`
	UpdatedBy        string                 `json:"updated_by" gorm:"type:varchar(100)"`
}

func (t *CarbonReductionTarget) TableName() string {
	return "carbon_reduction_targets"
}

func NewCarbonEmissionFactor(
	name, code string,
	scope CarbonEmissionScope,
	source string,
	value float64,
	unit, version string,
	effectiveAt time.Time,
) *CarbonEmissionFactor {
	return &CarbonEmissionFactor{
		Name:        name,
		Code:        code,
		Scope:       scope,
		Source:      source,
		Value:       value,
		Unit:        unit,
		Version:     version,
		EffectiveAt: effectiveAt,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func NewCarbonEmissionRecord(
	recordTime time.Time,
	scope CarbonEmissionScope,
	targetID, targetName string,
	factorID, factorCode string,
	factorValue float64,
	activityData float64,
	activityUnit, period string,
) *CarbonEmissionRecord {
	emissionValue := factorValue * activityData
	return &CarbonEmissionRecord{
		RecordTime:    recordTime,
		Scope:         scope,
		TargetID:      targetID,
		TargetName:    targetName,
		FactorID:      factorID,
		FactorCode:    factorCode,
		FactorValue:   factorValue,
		ActivityData:  activityData,
		ActivityUnit:  activityUnit,
		EmissionValue: emissionValue,
		EmissionUnit:  "tCO2e",
		Period:        period,
		Status:        CarbonEmissionStatusDraft,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}
