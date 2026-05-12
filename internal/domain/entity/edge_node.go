package entity

import "time"

type EdgeNodeStatus string

const (
	EdgeNodeOnline      EdgeNodeStatus = "online"
	EdgeNodeOffline     EdgeNodeStatus = "offline"
	EdgeNodeDegraded    EdgeNodeStatus = "degraded"
	EdgeNodeMaintenance EdgeNodeStatus = "maintenance"
)

type MapJSON map[string]string

type EdgeNode struct {
	ID            string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Name          string         `json:"name" gorm:"type:varchar(100);not null"`
	StationID     string         `json:"station_id" gorm:"type:varchar(36);not null;index"`
	IPAddress     string         `json:"ip_address" gorm:"type:varchar(45);not null"`
	Status        EdgeNodeStatus `json:"status" gorm:"type:varchar(20);default:'offline'"`
	CPUUsage      *float64       `json:"cpu_usage,omitempty" gorm:"type:decimal(5,2)"`
	MemoryUsage   *float64       `json:"memory_usage,omitempty" gorm:"type:decimal(5,2)"`
	ModelVersions MapJSON        `json:"model_versions" gorm:"type:jsonb"`
	LastHeartbeat *time.Time     `json:"last_heartbeat,omitempty"`
	ConfigVersion string         `json:"config_version" gorm:"type:varchar(50)"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func (EdgeNode) TableName() string {
	return "edge_nodes"
}
