package entity

import (
	"time"
)

type EnergyEfficiencyType string
type EnergyEfficiencyLevel string

const (
	EnergyEfficiencyTypeDevice     EnergyEfficiencyType = "device"
	EnergyEfficiencyTypeStation    EnergyEfficiencyType = "station"
	EnergyEfficiencyTypeSystem     EnergyEfficiencyType = "system"
	EnergyEfficiencyTypeComprehensive EnergyEfficiencyType = "comprehensive"
)

const (
	EnergyEfficiencyLevelExcellent EnergyEfficiencyLevel = "excellent"
	EnergyEfficiencyLevelGood      EnergyEfficiencyLevel = "good"
	EnergyEfficiencyLevelNormal    EnergyEfficiencyLevel = "normal"
	EnergyEfficiencyLevelPoor      EnergyEfficiencyLevel = "poor"
)

type EnergyEfficiencyRecord struct {
	ID          string                  `json:"id" gorm:"primaryKey;type:varchar(36)"`
	RecordTime  time.Time               `json:"record_time" gorm:"type:timestamp;not null;index:idx_ee_record_time"`
	Type        EnergyEfficiencyType    `json:"type" gorm:"type:varchar(20);not null;index:idx_ee_type"`
	TargetID    string                  `json:"target_id" gorm:"type:varchar(36);not null;index:idx_ee_target"`
	TargetName  string                  `json:"target_name" gorm:"type:varchar(200);not null"`
	
	InputEnergy  float64                `json:"input_energy" gorm:"type:decimal(20,4);not null"`
	OutputEnergy float64                `json:"output_energy" gorm:"type:decimal(20,4);not null"`
	Efficiency   float64                `json:"efficiency" gorm:"type:decimal(10,4);not null;index:idx_ee_efficiency"`
	EfficiencyLevel EnergyEfficiencyLevel `json:"efficiency_level" gorm:"type:varchar(20);not null"`
	
	BenchmarkEfficiency float64         `json:"benchmark_efficiency" gorm:"type:decimal(10,4)"`
	ImprovementRate     float64         `json:"improvement_rate" gorm:"type:decimal(10,4)"`
	
	Unit         string                 `json:"unit" gorm:"type:varchar(20);default:'kWh'"`
	Period       string                 `json:"period" gorm:"type:varchar(20);not null;index:idx_ee_period"`
	
	Metadata     map[string]interface{} `json:"metadata" gorm:"type:text;serializer:json"`
	
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

func (r *EnergyEfficiencyRecord) TableName() string {
	return "energy_efficiency_records"
}

type EnergyEfficiencyAnalysis struct {
	ID               string                  `json:"id" gorm:"primaryKey;type:varchar(36)"`
	AnalysisTime     time.Time               `json:"analysis_time" gorm:"type:timestamp;not null"`
	Type             EnergyEfficiencyType    `json:"type" gorm:"type:varchar(20);not null"`
	TargetID         string                  `json:"target_id" gorm:"type:varchar(36);not null"`
	TargetName       string                  `json:"target_name" gorm:"type:varchar(200);not null"`
	
	TimeRangeStart   time.Time               `json:"time_range_start" gorm:"type:timestamp;not null"`
	TimeRangeEnd     time.Time               `json:"time_range_end" gorm:"type:timestamp;not null"`
	
	AvgEfficiency    float64                `json:"avg_efficiency" gorm:"type:decimal(10,4);not null"`
	MaxEfficiency    float64                `json:"max_efficiency" gorm:"type:decimal(10,4);not null"`
	MinEfficiency    float64                `json:"min_efficiency" gorm:"type:decimal(10,4);not null"`
	StdDevEfficiency float64                `json:"std_dev_efficiency" gorm:"type:decimal(10,4)"`
	
	YoYChange        float64                `json:"yoy_change" gorm:"type:decimal(10,4)"`
	MoMChange        float64                `json:"mom_change" gorm:"type:decimal(10,4)"`
	
	Trend            string                 `json:"trend" gorm:"type:varchar(20)"`
	OptimizationSuggestions []string         `json:"optimization_suggestions" gorm:"type:text;serializer:json"`
	SavingPotential  float64                `json:"saving_potential" gorm:"type:decimal(20,4)"`
	
	CreatedAt        time.Time              `json:"created_at"`
}

func (a *EnergyEfficiencyAnalysis) TableName() string {
	return "energy_efficiency_analyses"
}

func NewEnergyEfficiencyRecord(
	recordTime time.Time,
	eeType EnergyEfficiencyType,
	targetID, targetName string,
	inputEnergy, outputEnergy float64,
	period string,
) *EnergyEfficiencyRecord {
	efficiency := 0.0
	if inputEnergy > 0 {
		efficiency = outputEnergy / inputEnergy
	}
	
	level := EnergyEfficiencyLevelNormal
	switch {
	case efficiency >= 0.95:
		level = EnergyEfficiencyLevelExcellent
	case efficiency >= 0.85:
		level = EnergyEfficiencyLevelGood
	case efficiency < 0.70:
		level = EnergyEfficiencyLevelPoor
	}
	
	return &EnergyEfficiencyRecord{
		RecordTime:      recordTime,
		Type:            eeType,
		TargetID:        targetID,
		TargetName:      targetName,
		InputEnergy:     inputEnergy,
		OutputEnergy:    outputEnergy,
		Efficiency:      efficiency,
		EfficiencyLevel: level,
		Period:          period,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}
