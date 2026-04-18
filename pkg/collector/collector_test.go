package collector

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockDataWriter 模拟数据写入器
type MockDataWriter struct {
	writeFunc    func(ctx context.Context, data []PointData) error
	writeBatchFunc func(ctx context.Context, batch [][]PointData) error
}

func (m *MockDataWriter) Write(ctx context.Context, data []PointData) error {
	if m.writeFunc != nil {
		return m.writeFunc(ctx, data)
	}
	return nil
}

func (m *MockDataWriter) WriteBatch(ctx context.Context, batch [][]PointData) error {
	if m.writeBatchFunc != nil {
		return m.writeBatchFunc(ctx, batch)
	}
	return nil
}

func (m *MockDataWriter) Close() error {
	return nil
}

func TestNewWorkerPool(t *testing.T) {
	pool := NewWorkerPool(
		WithMaxWorkers(100),
		WithMinWorkers(10),
		WithPoolTaskQueueSize(1000),
	)

	assert.NotNil(t, pool)
	assert.Equal(t, 100, pool.maxWorkers)
	assert.Equal(t, 10, pool.minWorkers)
	assert.Equal(t, 1000, pool.taskQueueSize)
}

func TestWorkerPool_StartStop(t *testing.T) {
	pool := NewWorkerPool(
		WithMaxWorkers(10),
		WithMinWorkers(2),
	)

	// 启动
	err := pool.Start()
	assert.NoError(t, err)
	assert.True(t, pool.IsRunning())

	// 重复启动
	err = pool.Start()
	assert.Error(t, err)

	// 停止
	err = pool.GracefulShutdown(5 * time.Second)
	assert.NoError(t, err)
	assert.True(t, pool.IsClosed())
}

func TestWorkerPool_Submit(t *testing.T) {
	pool := NewWorkerPool(
		WithMaxWorkers(10),
		WithMinWorkers(2),
	)

	err := pool.Start()
	assert.NoError(t, err)
	defer pool.GracefulShutdown(5 * time.Second)

	var mu sync.Mutex
	executed := false
	err = pool.Submit(context.Background(), "test-task", 1, func(ctx context.Context) error {
		mu.Lock()
		executed = true
		mu.Unlock()
		return nil
	})
	assert.NoError(t, err)

	// 等待任务执行
	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	assert.True(t, executed)
}

func TestWorkerPool_SubmitAndWait(t *testing.T) {
	pool := NewWorkerPool(
		WithMaxWorkers(10),
		WithMinWorkers(2),
	)

	err := pool.Start()
	assert.NoError(t, err)
	defer pool.GracefulShutdown(5 * time.Second)

	err = pool.SubmitAndWait(context.Background(), "test-task", 1, func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})
	assert.NoError(t, err)
}

func TestWorkerPool_GetMetrics(t *testing.T) {
	pool := NewWorkerPool(
		WithMaxWorkers(10),
		WithMinWorkers(2),
	)

	err := pool.Start()
	assert.NoError(t, err)
	defer pool.GracefulShutdown(5 * time.Second)

	// 提交一些任务
	for i := 0; i < 5; i++ {
		pool.Submit(context.Background(), "test-task", 1, func(ctx context.Context) error {
			time.Sleep(10 * time.Millisecond)
			return nil
		})
	}

	// 等待任务完成
	time.Sleep(200 * time.Millisecond)

	metrics := pool.GetMetrics()
	assert.Equal(t, int64(5), metrics.TotalTasks)
	assert.True(t, metrics.CompletedTasks > 0)
}

func TestWorkerPool_SetSize(t *testing.T) {
	pool := NewWorkerPool(
		WithMaxWorkers(10),
		WithMinWorkers(2),
	)

	err := pool.Start()
	assert.NoError(t, err)
	defer pool.GracefulShutdown(5 * time.Second)

	// 调整大小
	err = pool.SetSize(5, 20)
	assert.NoError(t, err)
	assert.Equal(t, 5, pool.minWorkers)
	assert.Equal(t, 20, pool.maxWorkers)

	// 无效大小
	err = pool.SetSize(0, 10)
	assert.Error(t, err)
}

func TestPriorityQueue(t *testing.T) {
	pq := NewPriorityQueue()

	// 添加任务（优先级低的先添加）
	pq.Push(&TaskWrapper{ID: "task1", Priority: 1})
	pq.Push(&TaskWrapper{ID: "task2", Priority: 3})
	pq.Push(&TaskWrapper{ID: "task3", Priority: 2})

	// 取出任务（优先级高的先出）
	task := pq.Pop()
	assert.NotNil(t, task)
	assert.Equal(t, "task2", task.ID)
	assert.Equal(t, 3, task.Priority)

	task = pq.Pop()
	assert.Equal(t, "task3", task.ID)

	task = pq.Pop()
	assert.Equal(t, "task1", task.ID)

	// 空队列
	task = pq.Pop()
	assert.Nil(t, task)
}

func TestNewDataBuffer(t *testing.T) {
	buffer := NewDataBuffer(
		WithMaxSize(10000),
		WithFlushInterval(5*time.Second),
		WithFlushThreshold(100),
	)

	assert.NotNil(t, buffer)
	assert.Equal(t, 10000, buffer.config.MaxSize)
	assert.Equal(t, 5*time.Second, buffer.config.FlushInterval)
	assert.Equal(t, 100, buffer.config.FlushThreshold)
}

func TestDataBuffer_Write(t *testing.T) {
	buffer := NewDataBuffer(
		WithMaxSize(100),
		WithFlushInterval(5*time.Second),
	)

	mockWriter := &MockDataWriter{}
	buffer.SetWriter(mockWriter)

	err := buffer.Start()
	assert.NoError(t, err)
	defer buffer.Stop()

	data := PointData{
		PointID:   "point001",
		Value:     100.5,
		Timestamp: time.Now(),
		Quality:   1,
	}

	err = buffer.Write(data)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int64(1), buffer.metrics.TotalData)
}

func TestDataBuffer_WriteBatch(t *testing.T) {
	buffer := NewDataBuffer(
		WithMaxSize(100),
		WithFlushInterval(5*time.Second),
	)

	mockWriter := &MockDataWriter{}
	buffer.SetWriter(mockWriter)

	err := buffer.Start()
	assert.NoError(t, err)
	defer buffer.Stop()

	dataPoints := []PointData{
		{PointID: "point001", Value: 100.0, Timestamp: time.Now(), Quality: 1},
		{PointID: "point002", Value: 200.0, Timestamp: time.Now(), Quality: 1},
		{PointID: "point003", Value: 300.0, Timestamp: time.Now(), Quality: 1},
	}

	err = buffer.WriteBatch(dataPoints)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int64(3), buffer.metrics.TotalData)
}

func TestDataBuffer_GetCurrentSize(t *testing.T) {
	buffer := NewDataBuffer(
		WithMaxSize(100),
		WithFlushInterval(5*time.Second),
	)

	mockWriter := &MockDataWriter{}
	buffer.SetWriter(mockWriter)

	err := buffer.Start()
	assert.NoError(t, err)
	defer buffer.Stop()

	// 添加数据
	for i := 0; i < 5; i++ {
		buffer.Write(PointData{
			PointID:   "point001",
			Value:     float64(i * 100),
			Timestamp: time.Now(),
			Quality:   1,
		})
	}

	time.Sleep(100 * time.Millisecond)
	size := buffer.GetCurrentSize()
	assert.Equal(t, 5, size)
}

func TestDataBuffer_Clear(t *testing.T) {
	buffer := NewDataBuffer(
		WithMaxSize(100),
		WithFlushInterval(5*time.Second),
	)

	mockWriter := &MockDataWriter{}
	buffer.SetWriter(mockWriter)

	err := buffer.Start()
	assert.NoError(t, err)
	defer buffer.Stop()

	// 添加数据
	for i := 0; i < 5; i++ {
		buffer.Write(PointData{
			PointID:   "point001",
			Value:     float64(i * 100),
			Timestamp: time.Now(),
			Quality:   1,
		})
	}

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 5, buffer.GetCurrentSize())

	// 清空
	buffer.Clear()
	assert.Equal(t, 0, buffer.GetCurrentSize())
}

func TestDataBuffer_ForceFlush(t *testing.T) {
	var mu sync.Mutex
	flushed := false
	mockWriter := &MockDataWriter{
		writeFunc: func(ctx context.Context, data []PointData) error {
			mu.Lock()
			flushed = true
			mu.Unlock()
			assert.Len(t, data, 3)
			return nil
		},
	}

	buffer := NewDataBuffer(
		WithMaxSize(100),
		WithFlushInterval(5*time.Second),
	)

	buffer.SetWriter(mockWriter)

	err := buffer.Start()
	assert.NoError(t, err)
	defer buffer.Stop()

	// 添加数据
	for i := 0; i < 3; i++ {
		buffer.Write(PointData{
			PointID:   "point001",
			Value:     float64(i * 100),
			Timestamp: time.Now(),
			Quality:   1,
		})
	}

	time.Sleep(100 * time.Millisecond)

	// 强制刷新
	err = buffer.ForceFlush()
	assert.NoError(t, err)

	// 等待刷新完成
	time.Sleep(200 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	assert.True(t, flushed)
}

func TestDataBuffer_GetMetrics(t *testing.T) {
	buffer := NewDataBuffer(
		WithMaxSize(100),
		WithFlushInterval(5*time.Second),
	)

	mockWriter := &MockDataWriter{}
	buffer.SetWriter(mockWriter)

	err := buffer.Start()
	assert.NoError(t, err)
	defer buffer.Stop()

	// 添加数据
	for i := 0; i < 5; i++ {
		buffer.Write(PointData{
			PointID:   "point001",
			Value:     float64(i * 100),
			Timestamp: time.Now(),
			Quality:   1,
		})
	}

	time.Sleep(100 * time.Millisecond)

	metrics := buffer.GetMetrics()
	assert.Equal(t, int64(5), metrics.TotalData)
}

func TestNewBatchWriter(t *testing.T) {
	mockWriter := &MockDataWriter{}

	bw := NewBatchWriter(mockWriter,
		WithBatchSize(100),
		WithParallelism(10),
		WithBatchTimeout(30*time.Second),
	)

	assert.NotNil(t, bw)
	assert.Equal(t, 100, bw.batchSize)
	assert.Equal(t, 10, bw.parallelism)
	assert.Equal(t, 30*time.Second, bw.timeout)
}

func TestBatchWriter_Write(t *testing.T) {
	var mu sync.Mutex
	writeCount := 0
	mockWriter := &MockDataWriter{
		writeFunc: func(ctx context.Context, data []PointData) error {
			mu.Lock()
			writeCount++
			mu.Unlock()
			return nil
		},
	}

	bw := NewBatchWriter(mockWriter, WithBatchSize(10))

	// 创建 25 个数据点，应该分成 3 批
	data := make([]PointData, 25)
	for i := 0; i < 25; i++ {
		data[i] = PointData{
			PointID:   "point001",
			Value:     float64(i),
			Timestamp: time.Now(),
			Quality:   1,
		}
	}

	err := bw.Write(context.Background(), data)
	assert.NoError(t, err)
	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 3, writeCount)
}

func TestNewRetryWriter(t *testing.T) {
	mockWriter := &MockDataWriter{}

	rw := NewRetryWriter(mockWriter,
		WithWriterMaxRetry(5),
		WithWriterRetryDelay(2*time.Second),
		WithExponentialBackoff(true),
	)

	assert.NotNil(t, rw)
	assert.Equal(t, 5, rw.maxRetry)
	assert.Equal(t, 2*time.Second, rw.retryDelay)
	assert.True(t, rw.exponentialBackoff)
}

func TestRetryWriter_Write_Success(t *testing.T) {
	attemptCount := 0
	mockWriter := &MockDataWriter{
		writeFunc: func(ctx context.Context, data []PointData) error {
			attemptCount++
			if attemptCount < 2 {
				return context.DeadlineExceeded
			}
			return nil
		},
	}

	rw := NewRetryWriter(mockWriter,
		WithWriterMaxRetry(3),
		WithWriterRetryDelay(10*time.Millisecond),
	)

	data := []PointData{
		{PointID: "point001", Value: 100.0, Timestamp: time.Now(), Quality: 1},
	}

	err := rw.Write(context.Background(), data)
	assert.NoError(t, err)
	assert.Equal(t, 2, attemptCount)
}

func TestRetryWriter_Write_AllFail(t *testing.T) {
	attemptCount := 0
	mockWriter := &MockDataWriter{
		writeFunc: func(ctx context.Context, data []PointData) error {
			attemptCount++
			return context.DeadlineExceeded
		},
	}

	rw := NewRetryWriter(mockWriter,
		WithWriterMaxRetry(3),
		WithWriterRetryDelay(10*time.Millisecond),
	)

	data := []PointData{
		{PointID: "point001", Value: 100.0, Timestamp: time.Now(), Quality: 1},
	}

	err := rw.Write(context.Background(), data)
	assert.Error(t, err)
	assert.Equal(t, 4, attemptCount) // 初始尝试 + 3次重试
}
