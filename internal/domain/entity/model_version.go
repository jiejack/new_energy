package entity

import "time"

type ModelStatus string

const (
	ModelStatusTraining ModelStatus = "training"
	ModelStatusStaging  ModelStatus = "staging"
	ModelStatusProd     ModelStatus = "production"
	ModelStatusRetired  ModelStatus = "retired"
)

type ModelVersion struct {
	ID             string      `json:"id" gorm:"primaryKey;type:varchar(36)"`
	ModelName      string      `json:"model_name" gorm:"type:varchar(100);not null;index"`
	Version        string      `json:"version" gorm:"type:varchar(50);not null"`
	Status         ModelStatus `json:"status" gorm:"type:varchar(20);default:'training'"`
	Accuracy       *float64    `json:"accuracy,omitempty" gorm:"type:decimal(6,4)"`
	TrainingConfig string      `json:"training_config" gorm:"type:text"`
	ArtifactPath   string      `json:"artifact_path" gorm:"type:varchar(500)"`
	DeployedAt     *time.Time  `json:"deployed_at,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

func (ModelVersion) TableName() string {
	return "model_versions"
}
