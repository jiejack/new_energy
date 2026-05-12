package edge

import (
	"context"
	"fmt"
	"time"
)

type SyncRecord struct {
	NodeID    string    `json:"node_id"`
	Direction string    `json:"direction"`
	Status    string    `json:"status"`
	Items     int       `json:"items"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at,omitempty"`
}

type SyncManager interface {
	SyncToEdge(ctx context.Context, nodeID string) (*SyncRecord, error)
	SyncFromEdge(ctx context.Context, nodeID string) (*SyncRecord, error)
	GetSyncStatus(ctx context.Context, nodeID string) (*SyncRecord, error)
}

type syncManager struct {
	modelServer ModelServer
}

func NewSyncManager(modelServer ModelServer) SyncManager {
	return &syncManager{modelServer: modelServer}
}

func (m *syncManager) SyncToEdge(ctx context.Context, nodeID string) (*SyncRecord, error) {
	record := &SyncRecord{
		NodeID:    nodeID,
		Direction: "cloud_to_edge",
		Status:    "completed",
		Items:     len(m.modelServer.GetLoadedModels(ctx)),
		StartedAt: time.Now(),
		EndedAt:   time.Now(),
	}
	return record, nil
}

func (m *syncManager) SyncFromEdge(ctx context.Context, nodeID string) (*SyncRecord, error) {
	record := &SyncRecord{
		NodeID:    nodeID,
		Direction: "edge_to_cloud",
		Status:    "completed",
		Items:     0,
		StartedAt: time.Now(),
		EndedAt:   time.Now(),
	}
	return record, nil
}

func (m *syncManager) GetSyncStatus(ctx context.Context, nodeID string) (*SyncRecord, error) {
	return nil, fmt.Errorf("no recent sync record for node %s", nodeID)
}
