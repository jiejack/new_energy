package harness

import (
	"context"
	"strings"
	"sync"
	"time"
)

// Monitor 运行监控器接口
type Monitor interface {
	// Record 记录指标数据
	Record(ctx context.Context, metric string, value float64) error

	// GetMetrics 获取指标数据
	GetMetrics(ctx context.Context, pattern string) ([]Metric, error)
}

// Metric 指标数据
type Metric struct {
	Name      string
	Value     float64
	Timestamp int64
	Labels    map[string]string
}

// DefaultMonitor 默认监控器实现
type DefaultMonitor struct {
	mu      sync.RWMutex
	metrics map[string][]Metric
}

// NewDefaultMonitor 创建默认监控器
func NewDefaultMonitor() *DefaultMonitor {
	return &DefaultMonitor{
		metrics: make(map[string][]Metric),
	}
}

// Record 记录指标数据
func (m *DefaultMonitor) Record(ctx context.Context, metric string, value float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().Unix()
	m.metrics[metric] = append(m.metrics[metric], Metric{
		Name:      metric,
		Value:     value,
		Timestamp: now,
		Labels:    make(map[string]string),
	})
	return nil
}

// GetMetrics 获取指标数据
func (m *DefaultMonitor) GetMetrics(ctx context.Context, pattern string) ([]Metric, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []Metric
	for name, metrics := range m.metrics {
		if pattern == "" || strings.Contains(name, pattern) {
			result = append(result, metrics...)
		}
	}
	return result, nil
}
