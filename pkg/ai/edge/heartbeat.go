package edge

import (
	"context"
	"sync"
	"time"
)

type HeartbeatInfo struct {
	NodeID      string    `json:"node_id"`
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
}

type HeartbeatMonitor interface {
	Receive(ctx context.Context, info HeartbeatInfo) error
	CheckTimeout(ctx context.Context, timeout time.Duration) []string
	GetNodeStatus(ctx context.Context, nodeID string) (*HeartbeatInfo, error)
}

type heartbeatMonitor struct {
	mu    sync.RWMutex
	nodes map[string]*HeartbeatInfo
}

func NewHeartbeatMonitor() HeartbeatMonitor {
	return &heartbeatMonitor{
		nodes: make(map[string]*HeartbeatInfo),
	}
}

func (m *heartbeatMonitor) Receive(ctx context.Context, info HeartbeatInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	info.Timestamp = time.Now()
	m.nodes[info.NodeID] = &info
	return nil
}

func (m *heartbeatMonitor) CheckTimeout(ctx context.Context, timeout time.Duration) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var timedOut []string
	now := time.Now()
	for id, info := range m.nodes {
		if now.Sub(info.Timestamp) > timeout {
			timedOut = append(timedOut, id)
		}
	}
	return timedOut
}

func (m *heartbeatMonitor) GetNodeStatus(ctx context.Context, nodeID string) (*HeartbeatInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	info, ok := m.nodes[nodeID]
	if !ok {
		return nil, nil
	}
	return info, nil
}
