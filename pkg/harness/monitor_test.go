package harness

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

// mockMonitor 是一个用于测试的 Monitor 实现
type mockMonitor struct {
	mu      sync.RWMutex
	metrics map[string][]Metric
	err     error
}

func newMockMonitor() *mockMonitor {
	return &mockMonitor{
		metrics: make(map[string][]Metric),
	}
}

func (m *mockMonitor) Record(ctx context.Context, metric string, value float64) error {
	if m.err != nil {
		return m.err
	}

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

func (m *mockMonitor) GetMetrics(ctx context.Context, pattern string) ([]Metric, error) {
	if m.err != nil {
		return nil, m.err
	}

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

func TestMonitor_Record(t *testing.T) {
	tests := []struct {
		name    string
		monitor Monitor
		metric  string
		value   float64
		wantErr bool
	}{
		{
			name:    "成功记录指标",
			monitor: newMockMonitor(),
			metric:  "cpu_usage",
			value:   75.5,
			wantErr: false,
		},
		{
			name:    "记录零值指标",
			monitor: newMockMonitor(),
			metric:  "error_count",
			value:   0,
			wantErr: false,
		},
		{
			name:    "记录负值指标",
			monitor: newMockMonitor(),
			metric:  "temperature_delta",
			value:   -5.5,
			wantErr: false,
		},
		{
			name: "记录失败",
			monitor: &mockMonitor{
				metrics: make(map[string][]Metric),
				err:     errors.New("storage error"),
			},
			metric:  "cpu_usage",
			value:   50.0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := tt.monitor.Record(ctx, tt.metric, tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("Record() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestMonitor_GetMetrics(t *testing.T) {
	tests := []struct {
		name       string
		monitor    Monitor
		setup      func(m Monitor)
		pattern    string
		wantCount  int
		wantErr    bool
		errMessage string
	}{
		{
			name:    "获取所有指标",
			monitor: newMockMonitor(),
			setup: func(m Monitor) {
				ctx := context.Background()
				m.Record(ctx, "cpu_usage", 75.5)
				m.Record(ctx, "memory_usage", 60.0)
				m.Record(ctx, "disk_usage", 80.0)
			},
			pattern:   "",
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:    "按模式过滤指标",
			monitor: newMockMonitor(),
			setup: func(m Monitor) {
				ctx := context.Background()
				m.Record(ctx, "cpu_usage", 75.5)
				m.Record(ctx, "cpu_temp", 65.0)
				m.Record(ctx, "memory_usage", 60.0)
			},
			pattern:   "cpu",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:    "获取不存在的指标",
			monitor: newMockMonitor(),
			setup: func(m Monitor) {
				ctx := context.Background()
				m.Record(ctx, "cpu_usage", 75.5)
			},
			pattern:   "network",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "获取指标失败",
			monitor: &mockMonitor{
				metrics: make(map[string][]Metric),
				err:     errors.New("query error"),
			},
			setup:      func(m Monitor) {},
			pattern:    "",
			wantErr:    true,
			errMessage: "query error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(tt.monitor)
			}

			ctx := context.Background()
			metrics, err := tt.monitor.GetMetrics(ctx, tt.pattern)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err.Error() != tt.errMessage {
					t.Errorf("GetMetrics() error message = %v, want %v", err.Error(), tt.errMessage)
				}
				return
			}

			if len(metrics) != tt.wantCount {
				t.Errorf("GetMetrics() returned %d metrics, want %d", len(metrics), tt.wantCount)
			}
		})
	}
}

func TestMetric_Structure(t *testing.T) {
	// 测试 Metric 结构体的字段
	metric := Metric{
		Name:      "cpu_usage",
		Value:     75.5,
		Timestamp: time.Now().Unix(),
		Labels: map[string]string{
			"host": "server-01",
			"env":  "production",
		},
	}

	if metric.Name != "cpu_usage" {
		t.Errorf("Metric.Name = %v, want cpu_usage", metric.Name)
	}

	if metric.Value != 75.5 {
		t.Errorf("Metric.Value = %v, want 75.5", metric.Value)
	}

	if metric.Timestamp == 0 {
		t.Error("Metric.Timestamp should not be zero")
	}

	if len(metric.Labels) != 2 {
		t.Errorf("Metric.Labels length = %v, want 2", len(metric.Labels))
	}

	if metric.Labels["host"] != "server-01" {
		t.Errorf("Metric.Labels[\"host\"] = %v, want server-01", metric.Labels["host"])
	}
}

func TestMonitor_Context_Cancellation(t *testing.T) {
	monitor := newMockMonitor()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 记录一个指标
	err := monitor.Record(ctx, "test_metric", 100.0)
	if err != nil {
		t.Errorf("Record() returned error: %v", err)
		return
	}

	// 取消上下文
	cancel()

	// 使用已取消的上下文记录指标（应该仍然成功，因为 mock 实现不检查上下文）
	ctx2 := context.Background()
	err = monitor.Record(ctx2, "test_metric2", 200.0)
	if err != nil {
		t.Errorf("Record() with new context returned error: %v", err)
	}

	// 获取指标
	metrics, err := monitor.GetMetrics(ctx2, "")
	if err != nil {
		t.Errorf("GetMetrics() returned error: %v", err)
		return
	}

	if len(metrics) != 2 {
		t.Errorf("GetMetrics() returned %d metrics, want 2", len(metrics))
	}
}

func TestMonitor_MultipleRecords(t *testing.T) {
	monitor := newMockMonitor()
	ctx := context.Background()

	// 记录多个相同名称的指标
	for i := 0; i < 5; i++ {
		err := monitor.Record(ctx, "cpu_usage", float64(i)*10)
		if err != nil {
			t.Errorf("Record() returned error: %v", err)
			return
		}
	}

	// 获取指标
	metrics, err := monitor.GetMetrics(ctx, "cpu_usage")
	if err != nil {
		t.Errorf("GetMetrics() returned error: %v", err)
		return
	}

	if len(metrics) != 5 {
		t.Errorf("GetMetrics() returned %d metrics, want 5", len(metrics))
	}

	// 验证值
	for i, m := range metrics {
		expected := float64(i) * 10
		if m.Value != expected {
			t.Errorf("metrics[%d].Value = %v, want %v", i, m.Value, expected)
		}
	}
}

func TestMonitor_ConcurrentAccess(t *testing.T) {
	monitor := newMockMonitor()
	ctx := context.Background()

	// 并发记录指标
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			for j := 0; j < 10; j++ {
				monitor.Record(ctx, "concurrent_metric", float64(id*10+j))
			}
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 获取指标
	metrics, err := monitor.GetMetrics(ctx, "concurrent_metric")
	if err != nil {
		t.Errorf("GetMetrics() returned error: %v", err)
		return
	}

	// 注意：由于并发访问 map 可能导致竞态条件，这里只检查总数
	// 在实际实现中应该使用适当的同步机制
	if len(metrics) != 100 {
		t.Errorf("GetMetrics() returned %d metrics, want 100", len(metrics))
	}
}
