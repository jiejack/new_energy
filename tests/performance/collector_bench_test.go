package performance

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/new-energy-monitoring/pkg/collector"
)

// BenchmarkCollectorMillionPoints 100万点位模拟采集性能测试
func BenchmarkCollectorMillionPoints(b *testing.B) {
	// 创建模拟采集器
	mockCollector := NewMockCollector("test-collector", 1000000) // 100万点位

	ctx := context.Background()
	config := &collector.CollectorConfig{
		ID:         "bench-collector",
		Name:       "Benchmark Collector",
		Protocol:   "mock",
		Interval:   time.Second,
		BufferSize: 100000,
		BatchSize:  10000,
		MaxWorkers: 100,
		MaxPoints:  1000000,
	}

	if err := mockCollector.Initialize(ctx, config); err != nil {
		b.Fatalf("Failed to initialize collector: %v", err)
	}

	if err := mockCollector.Start(ctx); err != nil {
		b.Fatalf("Failed to start collector: %v", err)
	}
	defer mockCollector.Stop(ctx)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			result, err := mockCollector.Collect(ctx)
			if err != nil {
				b.Errorf("Collect failed: %v", err)
			}
			if result.Count != 1000000 {
				b.Errorf("Expected 1000000 points, got %d", result.Count)
			}
		}
	})
}

// BenchmarkCollectorConcurrentConnections 并发连接测试
func BenchmarkCollectorConcurrentConnections(b *testing.B) {
	connectionCounts := []int{10, 50, 100, 500, 1000}

	for _, count := range connectionCounts {
		b.Run(fmt.Sprintf("Connections_%d", count), func(b *testing.B) {
			ctx := context.Background()

			collectors := make([]collector.Collector, count)
			for i := 0; i < count; i++ {
				collectors[i] = NewMockCollector(fmt.Sprintf("collector-%d", i), 1000)
				config := &collector.CollectorConfig{
					ID:         fmt.Sprintf("bench-collector-%d", i),
					Name:       fmt.Sprintf("Benchmark Collector %d", i),
					Protocol:   "mock",
					Interval:   time.Second,
					BufferSize: 1000,
					BatchSize:  100,
					MaxWorkers: 10,
					MaxPoints:  1000,
				}

				if err := collectors[i].Initialize(ctx, config); err != nil {
					b.Fatalf("Failed to initialize collector %d: %v", i, err)
				}

				if err := collectors[i].Start(ctx); err != nil {
					b.Fatalf("Failed to start collector %d: %v", i, err)
				}
			}

			defer func() {
				for _, c := range collectors {
					c.Stop(ctx)
				}
			}()

			b.ResetTimer()

			var wg sync.WaitGroup
			wg.Add(count)

			for i := 0; i < count; i++ {
				go func(idx int) {
					defer wg.Done()
					for j := 0; j < b.N; j++ {
						_, err := collectors[idx].Collect(ctx)
						if err != nil {
							b.Errorf("Collector %d collect failed: %v", idx, err)
						}
					}
				}(i)
			}

			wg.Wait()
		})
	}
}

// BenchmarkCollectorThroughput 数据吞吐量测试
func BenchmarkCollectorThroughput(b *testing.B) {
	pointCounts := []int{1000, 10000, 100000, 500000, 1000000}

	for _, points := range pointCounts {
		b.Run(fmt.Sprintf("Points_%d", points), func(b *testing.B) {
			mockCollector := NewMockCollector("throughput-test", points)

			ctx := context.Background()
			config := &collector.CollectorConfig{
				ID:         "throughput-collector",
				Name:       "Throughput Test Collector",
				Protocol:   "mock",
				Interval:   time.Millisecond * 100,
				BufferSize: points,
				BatchSize:  points / 10,
				MaxWorkers: 50,
				MaxPoints:  points,
			}

			if err := mockCollector.Initialize(ctx, config); err != nil {
				b.Fatalf("Failed to initialize collector: %v", err)
			}

			if err := mockCollector.Start(ctx); err != nil {
				b.Fatalf("Failed to start collector: %v", err)
			}
			defer mockCollector.Stop(ctx)

			b.ResetTimer()

			var totalPoints int64
			var totalDuration time.Duration

			for i := 0; i < b.N; i++ {
				start := time.Now()
				result, err := mockCollector.Collect(ctx)
				duration := time.Since(start)

				if err != nil {
					b.Errorf("Collect failed: %v", err)
					continue
				}

				atomic.AddInt64(&totalPoints, int64(result.Count))
				totalDuration += duration
			}

			b.ReportMetric(float64(totalPoints)/totalDuration.Seconds(), "points/sec")
		})
	}
}

// BenchmarkCollectorMemoryUsage 内存占用测试
func BenchmarkCollectorMemoryUsage(b *testing.B) {
	pointCounts := []int{10000, 50000, 100000, 500000, 1000000}

	for _, points := range pointCounts {
		b.Run(fmt.Sprintf("Memory_%d", points), func(b *testing.B) {
			mockCollector := NewMockCollector("memory-test", points)

			ctx := context.Background()
			config := &collector.CollectorConfig{
				ID:         "memory-collector",
				Name:       "Memory Test Collector",
				Protocol:   "mock",
				Interval:   time.Second,
				BufferSize: points,
				BatchSize:  points / 10,
				MaxWorkers: 20,
				MaxPoints:  points,
			}

			if err := mockCollector.Initialize(ctx, config); err != nil {
				b.Fatalf("Failed to initialize collector: %v", err)
			}

			if err := mockCollector.Start(ctx); err != nil {
				b.Fatalf("Failed to start collector: %v", err)
			}
			defer mockCollector.Stop(ctx)

			runtime.GC()
			var m1 runtime.MemStats
			runtime.ReadMemStats(&m1)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := mockCollector.Collect(ctx)
				if err != nil {
					b.Errorf("Collect failed: %v", err)
				}
			}

			runtime.GC()
			var m2 runtime.MemStats
			runtime.ReadMemStats(&m2)

			b.ReportMetric(float64(m2.Alloc-m1.Alloc)/1024/1024, "MB_allocated")
			b.ReportMetric(float64(m2.HeapAlloc)/1024/1024, "MB_heap")
			b.ReportMetric(float64(m2.Sys)/1024/1024, "MB_sys")
		})
	}
}

// BenchmarkCollectorCPUUsage CPU使用率测试
func BenchmarkCollectorCPUUsage(b *testing.B) {
	// 启动CPU profiling
	cpuProfile, err := os.Create("cpu_collector.prof")
	if err != nil {
		b.Fatalf("Failed to create CPU profile: %v", err)
	}
	defer cpuProfile.Close()

	if err := pprof.StartCPUProfile(cpuProfile); err != nil {
		b.Fatalf("Failed to start CPU profile: %v", err)
	}
	defer pprof.StopCPUProfile()

	mockCollector := NewMockCollector("cpu-test", 100000)

	ctx := context.Background()
	config := &collector.CollectorConfig{
		ID:         "cpu-collector",
		Name:       "CPU Test Collector",
		Protocol:   "mock",
		Interval:   time.Millisecond * 10,
		BufferSize: 100000,
		BatchSize:  10000,
		MaxWorkers: 100,
		MaxPoints:  100000,
	}

	if err := mockCollector.Initialize(ctx, config); err != nil {
		b.Fatalf("Failed to initialize collector: %v", err)
	}

	if err := mockCollector.Start(ctx); err != nil {
		b.Fatalf("Failed to start collector: %v", err)
	}
	defer mockCollector.Stop(ctx)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := mockCollector.Collect(ctx)
		if err != nil {
			b.Errorf("Collect failed: %v", err)
		}
	}
}

// BenchmarkCollectorWorkerPool 协程池性能测试
func BenchmarkCollectorWorkerPool(b *testing.B) {
	workerCounts := []int{10, 50, 100, 200, 500}

	for _, workers := range workerCounts {
		b.Run(fmt.Sprintf("Workers_%d", workers), func(b *testing.B) {
			pool := collector.NewWorkerPool(
				collector.WithMaxWorkers(workers),
				collector.WithMinWorkers(workers/10),
				collector.WithTaskQueueSize(100000),
			)

			if err := pool.Start(); err != nil {
				b.Fatalf("Failed to start worker pool: %v", err)
			}
			defer pool.GracefulShutdown(10 * time.Second)

			b.ResetTimer()

			var completedTasks int64

			for i := 0; i < b.N; i++ {
				taskID := fmt.Sprintf("task-%d", i)
				err := pool.Submit(context.Background(), taskID, 5, func(ctx context.Context) error {
					// 模拟采集任务
					time.Sleep(time.Microsecond * 100)
					atomic.AddInt64(&completedTasks, 1)
					return nil
				})

				if err != nil {
					b.Errorf("Failed to submit task: %v", err)
				}
			}

			// 等待所有任务完成
			for {
				metrics := pool.GetMetrics()
				if metrics.CompletedTasks >= int64(b.N) {
					break
				}
				time.Sleep(time.Millisecond * 10)
			}

			metrics := pool.GetMetrics()
			b.ReportMetric(float64(metrics.CompletedTasks)/float64(b.N)*100, "completion_rate")
			b.ReportMetric(float64(metrics.AverageDuration)/1e6, "avg_duration_ms")
		})
	}
}

// BenchmarkCollectorBuffer 缓冲区性能测试
func BenchmarkCollectorBuffer(b *testing.B) {
	bufferSizes := []int{1000, 5000, 10000, 50000, 100000}

	for _, size := range bufferSizes {
		b.Run(fmt.Sprintf("BufferSize_%d", size), func(b *testing.B) {
			buffer := collector.NewDataBuffer(size, 5*time.Minute)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				data := make([]collector.PointData, 100)
				for j := 0; j < 100; j++ {
					data[j] = collector.PointData{
						PointID:   fmt.Sprintf("point-%d-%d", i, j),
						Value:     float64(j),
						Quality:   192,
						Timestamp: time.Now(),
					}
				}

				if err := buffer.Write(data); err != nil {
					b.Errorf("Failed to write to buffer: %v", err)
				}
			}

			b.ReportMetric(float64(buffer.Size()), "buffer_size")
			b.ReportMetric(float64(buffer.Capacity()), "buffer_capacity")
		})
	}
}

// MockCollector 模拟采集器
type MockCollector struct {
	id       string
	status   collector.CollectorStatus
	config   *collector.CollectorConfig
	metrics  collector.CollectorMetrics
	pointCount int
	mu       sync.RWMutex
}

// NewMockCollector 创建模拟采集器
func NewMockCollector(id string, pointCount int) *MockCollector {
	return &MockCollector{
		id:         id,
		status:     collector.StatusUninitialized,
		pointCount: pointCount,
	}
}

// Initialize 初始化
func (m *MockCollector) Initialize(ctx context.Context, config *collector.CollectorConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = config
	m.status = collector.StatusInitialized
	return nil
}

// Start 启动
func (m *MockCollector) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.status = collector.StatusRunning
	return nil
}

// Stop 停止
func (m *MockCollector) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.status = collector.StatusStopped
	return nil
}

// Collect 采集数据
func (m *MockCollector) Collect(ctx context.Context) (*collector.CollectResult, error) {
	start := time.Now()

	m.mu.RLock()
	pointCount := m.pointCount
	m.mu.RUnlock()

	// 模拟采集延迟
	time.Sleep(time.Microsecond * time.Duration(pointCount/1000))

	// 生成模拟数据
	data := make([]collector.PointData, pointCount)
	for i := 0; i < pointCount; i++ {
		data[i] = collector.PointData{
			PointID:   fmt.Sprintf("point-%s-%d", m.id, i),
			Value:     float64(i) * 1.5,
			Quality:   192,
			Timestamp: time.Now(),
			Attributes: map[string]interface{}{
				"station_id": "station-001",
				"device_id":  "device-001",
			},
		}
	}

	m.mu.Lock()
	m.metrics.TotalCollects++
	m.metrics.SuccessCollects++
	m.metrics.TotalPoints += int64(pointCount)
	m.metrics.TotalDuration += time.Since(start)
	m.metrics.AverageDuration = time.Duration(int64(m.metrics.TotalDuration) / m.metrics.TotalCollects)
	m.metrics.LastCollectTime = time.Now()
	m.mu.Unlock()

	return &collector.CollectResult{
		CollectorID: m.id,
		Success:     true,
		Data:        data,
		Count:       pointCount,
		Duration:    time.Since(start),
		Timestamp:   time.Now(),
	}, nil
}

// GetStatus 获取状态
func (m *MockCollector) GetStatus() collector.CollectorStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

// GetConfig 获取配置
func (m *MockCollector) GetConfig() *collector.CollectorConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// GetMetrics 获取指标
func (m *MockCollector) GetMetrics() *collector.CollectorMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return &m.metrics
}

// HealthCheck 健康检查
func (m *MockCollector) HealthCheck(ctx context.Context) error {
	return nil
}

// DataBuffer 数据缓冲区（简化实现）
type DataBuffer struct {
	data      []collector.PointData
	size      int
	capacity  int
	ttl       time.Duration
	mu        sync.RWMutex
}

// NewDataBuffer 创建数据缓冲区
func NewDataBuffer(capacity int, ttl time.Duration) *DataBuffer {
	return &DataBuffer{
		data:     make([]collector.PointData, 0, capacity),
		capacity: capacity,
		ttl:      ttl,
	}
}

// Write 写入数据
func (b *DataBuffer) Write(data []collector.PointData) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.data)+len(data) > b.capacity {
		// 淘汰旧数据
		overflow := len(b.data) + len(data) - b.capacity
		if overflow > 0 && overflow < len(b.data) {
			b.data = b.data[overflow:]
		}
	}

	b.data = append(b.data, data...)
	b.size = len(b.data)
	return nil
}

// Size 获取大小
func (b *DataBuffer) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.size
}

// Capacity 获取容量
func (b *DataBuffer) Capacity() int {
	return b.capacity
}
