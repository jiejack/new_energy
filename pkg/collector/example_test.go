package collector

import (
	"context"
	"fmt"
	"time"
)

// Example 展示如何使用数据采集框架
func Example() {
	// 1. 创建协程池
	pool := NewWorkerPool(
		WithMaxWorkers(1000),
		WithMinWorkers(10),
		WithPoolTaskQueueSize(100000),
		WithIdleTimeout(30*time.Second),
	)

	// 启动协程池
	if err := pool.Start(); err != nil {
		panic(err)
	}
	defer pool.GracefulShutdown(30 * time.Second)

	// 2. 创建调度器
	scheduler := NewScheduler(pool,
		WithMaxConcurrentTasks(1000),
		WithTaskQueueSize(100000),
		WithScheduleInterval(100*time.Millisecond),
	)

	// 启动调度器
	if err := scheduler.Start(); err != nil {
		panic(err)
	}
	defer scheduler.Stop()

	// 3. 创建数据缓冲区
	buffer := NewDataBuffer(
		WithMaxSize(1000000),
		WithFlushInterval(5*time.Second),
		WithFlushThreshold(10000),
		WithMaxRetryCount(3),
	)

	// 设置数据写入器
	writer := &ExampleDataWriter{}
	buffer.SetWriter(writer)

	// 启动缓冲区
	if err := buffer.Start(); err != nil {
		panic(err)
	}
	defer buffer.Stop()

	// 4. 创建并注册采集器
	collector := &ExampleCollector{
		id: "collector-1",
		name: "Example Collector",
	}
	
	if err := scheduler.RegisterCollector("collector-1", collector); err != nil {
		panic(err)
	}

	// 5. 添加采集任务
	task := &Task{
		ID:          "task-1",
		Name:        "Periodic Collection Task",
		Type:        TaskTypePeriodic,
		Priority:    5,
		CollectorID: "collector-1",
		Interval:    10 * time.Second,
		Timeout:     5 * time.Second,
		MaxRetry:    3,
	}

	if err := scheduler.AddTask(task); err != nil {
		panic(err)
	}

	// 6. 设置事件处理器
	eventHandler := &ExampleEventHandler{
		buffer: buffer,
	}
	scheduler.SetEventHandler(eventHandler)

	// 7. 运行一段时间
	time.Sleep(1 * time.Minute)

	// 8. 获取指标
	poolMetrics := pool.GetMetrics()
	fmt.Printf("Pool Metrics: %+v\n", poolMetrics)

	schedulerMetrics := scheduler.GetMetrics()
	fmt.Printf("Scheduler Metrics: %+v\n", schedulerMetrics)

	bufferMetrics := buffer.GetMetrics()
	fmt.Printf("Buffer Metrics: %+v\n", bufferMetrics)
}

// ExampleCollector 示例采集器
type ExampleCollector struct {
	id     string
	name   string
	status CollectorStatus
	config *CollectorConfig
	metrics CollectorMetrics
}

func (c *ExampleCollector) Initialize(ctx context.Context, config *CollectorConfig) error {
	c.config = config
	c.status = StatusInitialized
	return nil
}

func (c *ExampleCollector) Start(ctx context.Context) error {
	c.status = StatusRunning
	return nil
}

func (c *ExampleCollector) Stop(ctx context.Context) error {
	c.status = StatusStopped
	return nil
}

func (c *ExampleCollector) Collect(ctx context.Context) (*CollectResult, error) {
	startTime := time.Now()
	
	// 模拟数据采集
	data := make([]PointData, 10)
	for i := 0; i < 10; i++ {
		data[i] = PointData{
			PointID:   fmt.Sprintf("point-%d", i),
			Value:     float64(i * 100),
			Quality:   100,
			Timestamp: time.Now(),
		}
	}

	duration := time.Since(startTime)

	return &CollectResult{
		CollectorID: c.id,
		Success:     true,
		Data:        data,
		Count:       len(data),
		Duration:    duration,
		Timestamp:   time.Now(),
	}, nil
}

func (c *ExampleCollector) GetStatus() CollectorStatus {
	return c.status
}

func (c *ExampleCollector) GetConfig() *CollectorConfig {
	return c.config
}

func (c *ExampleCollector) GetMetrics() *CollectorMetrics {
	return &c.metrics
}

func (c *ExampleCollector) HealthCheck(ctx context.Context) error {
	return nil
}

// ExampleEventHandler 示例事件处理器
type ExampleEventHandler struct {
	buffer *DataBuffer
}

func (h *ExampleEventHandler) OnTaskComplete(result *TaskResult) {
	if result.Result != nil {
		// 将采集的数据写入缓冲区
		for _, data := range result.Result.Data {
			if err := h.buffer.Write(data); err != nil {
				fmt.Printf("Failed to write data to buffer: %v\n", err)
			}
		}
	}
	fmt.Printf("Task completed: %s, Duration: %v\n", result.TaskID, result.Duration)
}

func (h *ExampleEventHandler) OnTaskFailed(result *TaskResult) {
	fmt.Printf("Task failed: %s, Error: %v\n", result.TaskID, result.Error)
}

func (h *ExampleEventHandler) OnCollectorError(collectorID string, err error) {
	fmt.Printf("Collector error: %s, Error: %v\n", collectorID, err)
}

// ExampleDataWriter 示例数据写入器
type ExampleDataWriter struct{}

func (w *ExampleDataWriter) Write(ctx context.Context, data []PointData) error {
	// 模拟写入数据库或消息队列
	fmt.Printf("Writing %d data points\n", len(data))
	return nil
}

func (w *ExampleDataWriter) WriteBatch(ctx context.Context, batch [][]PointData) error {
	totalPoints := 0
	for _, data := range batch {
		totalPoints += len(data)
	}
	fmt.Printf("Writing batch with %d data points\n", totalPoints)
	return nil
}

func (w *ExampleDataWriter) Close() error {
	return nil
}
