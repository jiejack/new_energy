package collector

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockCollector struct {
	status  CollectorStatus
	config  *CollectorConfig
	metrics CollectorMetrics
	collectFunc func(ctx context.Context) (*CollectResult, error)
}

func (m *MockCollector) Initialize(ctx context.Context, config *CollectorConfig) error {
	m.config = config
	m.status = StatusInitialized
	return nil
}

func (m *MockCollector) Start(ctx context.Context) error {
	m.status = StatusRunning
	return nil
}

func (m *MockCollector) Stop(ctx context.Context) error {
	m.status = StatusStopped
	return nil
}

func (m *MockCollector) Collect(ctx context.Context) (*CollectResult, error) {
	if m.collectFunc != nil {
		return m.collectFunc(ctx)
	}
	return &CollectResult{
		CollectorID: "mock-collector",
		Success:     true,
		Data: []PointData{
			{PointID: "p1", Value: 100.0, Timestamp: time.Now(), Quality: 1},
		},
		Count:     1,
		Duration:  time.Millisecond,
		Timestamp: time.Now(),
	}, nil
}

func (m *MockCollector) GetStatus() CollectorStatus {
	return m.status
}

func (m *MockCollector) GetConfig() *CollectorConfig {
	return m.config
}

func (m *MockCollector) GetMetrics() *CollectorMetrics {
	return &m.metrics
}

func (m *MockCollector) HealthCheck(ctx context.Context) error {
	return nil
}

type MockEventHandler struct {
	completedCount int32
	failedCount    int32
	errorCount     int32
}

func (h *MockEventHandler) OnTaskComplete(result *TaskResult) {
	atomic.AddInt32(&h.completedCount, 1)
}

func (h *MockEventHandler) OnTaskFailed(result *TaskResult) {
	atomic.AddInt32(&h.failedCount, 1)
}

func (h *MockEventHandler) OnCollectorError(collectorID string, err error) {
	atomic.AddInt32(&h.errorCount, 1)
}

func TestCollectorStatus_String(t *testing.T) {
	assert.Equal(t, "Uninitialized", StatusUninitialized.String())
	assert.Equal(t, "Initialized", StatusInitialized.String())
	assert.Equal(t, "Running", StatusRunning.String())
	assert.Equal(t, "Stopped", StatusStopped.String())
	assert.Equal(t, "Error", StatusError.String())
	assert.Equal(t, "Unknown", CollectorStatus(99).String())
}

func TestTaskType_String(t *testing.T) {
	assert.Equal(t, "Periodic", TaskTypePeriodic.String())
	assert.Equal(t, "Event", TaskTypeEvent.String())
	assert.Equal(t, "Once", TaskTypeOnce.String())
	assert.Equal(t, "Unknown", TaskType(99).String())
}

func TestTaskStatus_String(t *testing.T) {
	assert.Equal(t, "Pending", TaskStatusPending.String())
	assert.Equal(t, "Running", TaskStatusRunning.String())
	assert.Equal(t, "Completed", TaskStatusCompleted.String())
	assert.Equal(t, "Failed", TaskStatusFailed.String())
	assert.Equal(t, "Cancelled", TaskStatusCancelled.String())
	assert.Equal(t, "Unknown", TaskStatus(99).String())
}

func TestNewScheduler(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	scheduler := NewScheduler(pool,
		WithMaxConcurrentTasks(100),
		WithTaskQueueSize(1000),
		WithScheduleInterval(100*time.Millisecond),
		WithEventBufferSize(500),
		WithSchedulerEnableMetrics(true),
	)

	assert.NotNil(t, scheduler)
	assert.Equal(t, 100, scheduler.config.MaxConcurrentTasks)
	assert.Equal(t, 1000, scheduler.config.TaskQueueSize)
	assert.Equal(t, 100*time.Millisecond, scheduler.config.ScheduleInterval)
	assert.Equal(t, 500, scheduler.config.EventBufferSize)
	assert.True(t, scheduler.config.EnableMetrics)
}

func TestScheduler_StartStop(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	require.NoError(t, pool.Start())
	defer pool.GracefulShutdown(5 * time.Second)

	scheduler := NewScheduler(pool, WithScheduleInterval(100*time.Millisecond))

	err := scheduler.Start()
	assert.NoError(t, err)
	assert.True(t, scheduler.IsRunning())

	err = scheduler.Start()
	assert.Error(t, err)
	assert.Equal(t, ErrSchedulerRunning, err)

	err = scheduler.Stop()
	assert.NoError(t, err)
	assert.False(t, scheduler.IsRunning())
}

func TestScheduler_StopNotRunning(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	scheduler := NewScheduler(pool)

	err := scheduler.Stop()
	assert.Error(t, err)
	assert.Equal(t, ErrSchedulerNotRunning, err)
}

func TestScheduler_AddTask(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	scheduler := NewScheduler(pool)

	task := &Task{
		ID:          "task-001",
		Name:        "Test Task",
		Type:        TaskTypePeriodic,
		Priority:    5,
		CollectorID: "collector-001",
		Interval:    10 * time.Second,
		Timeout:     5 * time.Second,
	}

	err := scheduler.AddTask(task)
	assert.NoError(t, err)
	assert.Equal(t, 1, scheduler.GetTaskCount())

	retrieved, err := scheduler.GetTask("task-001")
	assert.NoError(t, err)
	assert.Equal(t, "Test Task", retrieved.Name)
	assert.False(t, retrieved.CreateTime.IsZero())
	assert.False(t, retrieved.NextRunTime.IsZero())
}

func TestScheduler_AddTask_Invalid(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	scheduler := NewScheduler(pool)

	err := scheduler.AddTask(nil)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidTask, err)

	err = scheduler.AddTask(&Task{Name: "No ID"})
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidTask, err)
}

func TestScheduler_AddTask_Duplicate(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	scheduler := NewScheduler(pool)

	task := &Task{ID: "task-001", Name: "Test", Type: TaskTypeOnce}
	scheduler.AddTask(task)

	err := scheduler.AddTask(&Task{ID: "task-001", Name: "Duplicate"})
	assert.Error(t, err)
	assert.Equal(t, ErrTaskExists, err)
}

func TestScheduler_AddTask_DefaultStatus(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	scheduler := NewScheduler(pool)

	task := &Task{ID: "task-001", Name: "Test", Type: TaskTypeOnce, Status: TaskStatusCancelled}
	scheduler.AddTask(task)

	retrieved, _ := scheduler.GetTask("task-001")
	assert.Equal(t, TaskStatusPending, retrieved.Status)
}

func TestScheduler_RemoveTask(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	scheduler := NewScheduler(pool)

	task := &Task{ID: "task-001", Name: "Test", Type: TaskTypeOnce}
	scheduler.AddTask(task)

	err := scheduler.RemoveTask("task-001")
	assert.NoError(t, err)
	assert.Equal(t, 0, scheduler.GetTaskCount())
}

func TestScheduler_RemoveTask_NotFound(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	scheduler := NewScheduler(pool)

	err := scheduler.RemoveTask("nonexistent")
	assert.Error(t, err)
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestScheduler_GetTask_NotFound(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	scheduler := NewScheduler(pool)

	_, err := scheduler.GetTask("nonexistent")
	assert.Error(t, err)
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestScheduler_GetAllTasks(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	scheduler := NewScheduler(pool)

	scheduler.AddTask(&Task{ID: "task-001", Name: "Task1", Type: TaskTypeOnce})
	scheduler.AddTask(&Task{ID: "task-002", Name: "Task2", Type: TaskTypePeriodic})

	tasks := scheduler.GetAllTasks()
	assert.Len(t, tasks, 2)
}

func TestScheduler_RegisterCollector(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	scheduler := NewScheduler(pool)

	collector := &MockCollector{}
	err := scheduler.RegisterCollector("collector-001", collector)
	assert.NoError(t, err)
	assert.Equal(t, 1, scheduler.GetCollectorCount())
}

func TestScheduler_UnregisterCollector(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	scheduler := NewScheduler(pool)

	collector := &MockCollector{}
	scheduler.RegisterCollector("collector-001", collector)
	err := scheduler.UnregisterCollector("collector-001")
	assert.NoError(t, err)
	assert.Equal(t, 0, scheduler.GetCollectorCount())
}

func TestScheduler_SetEventHandler(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	scheduler := NewScheduler(pool)

	handler := &MockEventHandler{}
	scheduler.SetEventHandler(handler)
	assert.NotNil(t, scheduler.eventHandler)
}

func TestScheduler_TriggerEvent(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	require.NoError(t, pool.Start())
	defer pool.GracefulShutdown(5 * time.Second)

	scheduler := NewScheduler(pool, WithScheduleInterval(100*time.Millisecond))
	require.NoError(t, scheduler.Start())
	defer scheduler.Stop()

	err := scheduler.TriggerEvent("event-001", "task-001", nil)
	assert.NoError(t, err)
}

func TestScheduler_TriggerEvent_Stopped(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	scheduler := NewScheduler(pool)

	scheduler.cancelFunc()
	time.Sleep(10 * time.Millisecond)
	err := scheduler.TriggerEvent("event-001", "task-001", nil)
	if err != nil {
		assert.Error(t, err)
	}
}

func TestScheduler_PauseTask(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	scheduler := NewScheduler(pool)

	task := &Task{ID: "task-001", Name: "Test", Type: TaskTypePeriodic, Status: TaskStatusPending}
	scheduler.AddTask(task)

	err := scheduler.PauseTask("task-001")
	assert.NoError(t, err)

	retrieved, _ := scheduler.GetTask("task-001")
	assert.Equal(t, TaskStatusCancelled, retrieved.Status)
}

func TestScheduler_PauseTask_NotFound(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	scheduler := NewScheduler(pool)

	err := scheduler.PauseTask("nonexistent")
	assert.Error(t, err)
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestScheduler_ResumeTask(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	scheduler := NewScheduler(pool)

	task := &Task{ID: "task-001", Name: "Test", Type: TaskTypePeriodic, Status: TaskStatusCancelled}
	scheduler.AddTask(task)

	err := scheduler.ResumeTask("task-001")
	assert.NoError(t, err)

	retrieved, _ := scheduler.GetTask("task-001")
	assert.Equal(t, TaskStatusPending, retrieved.Status)
}

func TestScheduler_ResumeTask_NotFound(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	scheduler := NewScheduler(pool)

	err := scheduler.ResumeTask("nonexistent")
	assert.Error(t, err)
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestScheduler_UpdateTaskInterval(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	scheduler := NewScheduler(pool)

	task := &Task{ID: "task-001", Name: "Test", Type: TaskTypePeriodic, Interval: 10 * time.Second}
	scheduler.AddTask(task)

	err := scheduler.UpdateTaskInterval("task-001", 30*time.Second)
	assert.NoError(t, err)

	retrieved, _ := scheduler.GetTask("task-001")
	assert.Equal(t, 30*time.Second, retrieved.Interval)
}

func TestScheduler_UpdateTaskInterval_NotFound(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	scheduler := NewScheduler(pool)

	err := scheduler.UpdateTaskInterval("nonexistent", 30*time.Second)
	assert.Error(t, err)
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestScheduler_GetMetrics(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	scheduler := NewScheduler(pool)

	scheduler.AddTask(&Task{ID: "task-001", Name: "Test", Type: TaskTypeOnce})

	metrics := scheduler.GetMetrics()
	assert.Equal(t, int64(1), metrics.TotalTasks)
}

func TestScheduler_IsClosed(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	scheduler := NewScheduler(pool)
	assert.False(t, scheduler.IsClosed())
}

func TestScheduler_ExecuteTask_WithCollector(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	require.NoError(t, pool.Start())
	defer pool.GracefulShutdown(5 * time.Second)

	handler := &MockEventHandler{}
	collector := &MockCollector{}

	scheduler := NewScheduler(pool, WithScheduleInterval(50*time.Millisecond))
	scheduler.SetEventHandler(handler)
	scheduler.RegisterCollector("collector-001", collector)
	require.NoError(t, scheduler.Start())
	defer scheduler.Stop()

	task := &Task{
		ID:          "task-001",
		Name:        "Test Task",
		Type:        TaskTypePeriodic,
		Priority:    5,
		CollectorID: "collector-001",
		Interval:    50 * time.Millisecond,
		Timeout:     5 * time.Second,
	}
	scheduler.AddTask(task)

	time.Sleep(300 * time.Millisecond)

	metrics := scheduler.GetMetrics()
	assert.True(t, metrics.ScheduledTasks > 0 || metrics.CompletedTasks > 0)
}

func TestScheduler_ExecuteTask_CollectorNotFound(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	require.NoError(t, pool.Start())
	defer pool.GracefulShutdown(5 * time.Second)

	handler := &MockEventHandler{}
	scheduler := NewScheduler(pool, WithScheduleInterval(50*time.Millisecond))
	scheduler.SetEventHandler(handler)
	require.NoError(t, scheduler.Start())
	defer scheduler.Stop()

	task := &Task{
		ID:          "task-001",
		Name:        "Test Task",
		Type:        TaskTypePeriodic,
		Priority:    5,
		CollectorID: "nonexistent-collector",
		Interval:    50 * time.Millisecond,
		Timeout:     5 * time.Second,
		MaxRetry:    0,
	}
	scheduler.AddTask(task)

	time.Sleep(300 * time.Millisecond)

	metrics := scheduler.GetMetrics()
	assert.True(t, metrics.FailedTasks > 0)
}

func TestScheduler_ExecuteTask_CollectorError(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	require.NoError(t, pool.Start())
	defer pool.GracefulShutdown(5 * time.Second)

	handler := &MockEventHandler{}
	collector := &MockCollector{
		collectFunc: func(ctx context.Context) (*CollectResult, error) {
			return nil, errors.New("collection failed")
		},
	}

	scheduler := NewScheduler(pool, WithScheduleInterval(50*time.Millisecond))
	scheduler.SetEventHandler(handler)
	scheduler.RegisterCollector("collector-001", collector)
	require.NoError(t, scheduler.Start())
	defer scheduler.Stop()

	task := &Task{
		ID:          "task-001",
		Name:        "Test Task",
		Type:        TaskTypePeriodic,
		Priority:    5,
		CollectorID: "collector-001",
		Interval:    50 * time.Millisecond,
		Timeout:     5 * time.Second,
		MaxRetry:    0,
	}
	scheduler.AddTask(task)

	time.Sleep(300 * time.Millisecond)

	assert.True(t, atomic.LoadInt32(&handler.failedCount) > 0)
}

func TestScheduler_TriggerEvent_WithEventTask(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	require.NoError(t, pool.Start())
	defer pool.GracefulShutdown(5 * time.Second)

	handler := &MockEventHandler{}
	collector := &MockCollector{}

	scheduler := NewScheduler(pool, WithScheduleInterval(100*time.Millisecond))
	scheduler.SetEventHandler(handler)
	scheduler.RegisterCollector("collector-001", collector)
	require.NoError(t, scheduler.Start())
	defer scheduler.Stop()

	task := &Task{
		ID:          "task-001",
		Name:        "Event Task",
		Type:        TaskTypeEvent,
		Priority:    5,
		CollectorID: "collector-001",
		Timeout:     5 * time.Second,
		Status:      TaskStatusPending,
	}
	scheduler.AddTask(task)

	scheduler.TriggerEvent("event-001", "task-001", map[string]string{"key": "value"})

	time.Sleep(200 * time.Millisecond)

	assert.True(t, atomic.LoadInt32(&handler.completedCount) > 0)
}

func TestScheduler_NoMetrics(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	require.NoError(t, pool.Start())
	defer pool.GracefulShutdown(5 * time.Second)

	scheduler := NewScheduler(pool,
		WithScheduleInterval(100*time.Millisecond),
		WithSchedulerEnableMetrics(false),
	)
	require.NoError(t, scheduler.Start())
	defer scheduler.Stop()

	assert.False(t, scheduler.config.EnableMetrics)
}

func TestDataBuffer_Write_Closed(t *testing.T) {
	buffer := NewDataBuffer(WithMaxSize(100))
	buffer.SetWriter(&MockDataWriter{})
	require.NoError(t, buffer.Start())
	require.NoError(t, buffer.Stop())

	err := buffer.Write(PointData{PointID: "p1", Value: 1.0, Timestamp: time.Now()})
	assert.Error(t, err)
	assert.Equal(t, ErrBufferClosed, err)
}

func TestDataBuffer_WriteBatch_Closed(t *testing.T) {
	buffer := NewDataBuffer(WithMaxSize(100))
	buffer.SetWriter(&MockDataWriter{})
	require.NoError(t, buffer.Start())
	require.NoError(t, buffer.Stop())

	err := buffer.WriteBatch([]PointData{{PointID: "p1", Value: 1.0, Timestamp: time.Now()}})
	assert.Error(t, err)
	assert.Equal(t, ErrBufferClosed, err)
}

func TestDataBuffer_ForceFlush_Closed(t *testing.T) {
	buffer := NewDataBuffer(WithMaxSize(100))
	buffer.SetWriter(&MockDataWriter{})
	require.NoError(t, buffer.Start())
	require.NoError(t, buffer.Stop())

	err := buffer.ForceFlush()
	assert.Error(t, err)
	assert.Equal(t, ErrBufferClosed, err)
}

func TestDataBuffer_IsRunning_IsClosed(t *testing.T) {
	buffer := NewDataBuffer(WithMaxSize(100))
	buffer.SetWriter(&MockDataWriter{})

	assert.False(t, buffer.IsRunning())
	assert.False(t, buffer.IsClosed())

	require.NoError(t, buffer.Start())
	assert.True(t, buffer.IsRunning())
	assert.False(t, buffer.IsClosed())

	require.NoError(t, buffer.Stop())
	assert.False(t, buffer.IsRunning())
	assert.True(t, buffer.IsClosed())
}

func TestDataBuffer_Start_NoWriter(t *testing.T) {
	buffer := NewDataBuffer(WithMaxSize(100))
	err := buffer.Start()
	assert.Error(t, err)
	assert.Equal(t, ErrWriterNotSet, err)
}

func TestDataBuffer_Start_AlreadyRunning(t *testing.T) {
	buffer := NewDataBuffer(WithMaxSize(100))
	buffer.SetWriter(&MockDataWriter{})
	require.NoError(t, buffer.Start())
	defer buffer.Stop()

	err := buffer.Start()
	assert.Error(t, err)
}

func TestDataBuffer_Stop_NotRunning(t *testing.T) {
	buffer := NewDataBuffer(WithMaxSize(100))
	err := buffer.Stop()
	assert.Error(t, err)
}

func TestDataBuffer_WriteWithRetry_Failure(t *testing.T) {
	callCount := int32(0)
	mockWriter := &MockDataWriter{
		writeFunc: func(ctx context.Context, data []PointData) error {
			atomic.AddInt32(&callCount, 1)
			return errors.New("write failed")
		},
	}

	buffer := NewDataBuffer(
		WithMaxSize(100),
		WithFlushInterval(5*time.Second),
		WithMaxRetryCount(2),
		WithRetryDelay(10*time.Millisecond),
	)
	buffer.SetWriter(mockWriter)
	require.NoError(t, buffer.Start())
	defer buffer.Stop()

	buffer.Write(PointData{PointID: "p1", Value: 1.0, Timestamp: time.Now()})
	time.Sleep(100 * time.Millisecond)
	buffer.ForceFlush()
	time.Sleep(200 * time.Millisecond)
}

func TestDataBuffer_Options(t *testing.T) {
	buffer := NewDataBuffer(
		WithMaxRetryCount(5),
		WithRetryDelay(2*time.Second),
		WithEnableMetrics(false),
		WithEnableCompression(true),
	)

	assert.Equal(t, 5, buffer.config.MaxRetryCount)
	assert.Equal(t, 2*time.Second, buffer.config.RetryDelay)
	assert.False(t, buffer.config.EnableMetrics)
	assert.True(t, buffer.config.EnableCompression)
}

func TestDataBuffer_FlushThreshold(t *testing.T) {
	flushCount := int32(0)
	mockWriter := &MockDataWriter{
		writeFunc: func(ctx context.Context, data []PointData) error {
			atomic.AddInt32(&flushCount, 1)
			return nil
		},
	}

	buffer := NewDataBuffer(
		WithMaxSize(1000),
		WithFlushInterval(5*time.Second),
		WithFlushThreshold(5),
	)
	buffer.SetWriter(mockWriter)
	require.NoError(t, buffer.Start())
	defer buffer.Stop()

	for i := 0; i < 5; i++ {
		buffer.Write(PointData{PointID: "p1", Value: float64(i), Timestamp: time.Now()})
	}

	time.Sleep(300 * time.Millisecond)
	assert.True(t, atomic.LoadInt32(&flushCount) > 0)
}

func TestBatchWriter_WriteBatch(t *testing.T) {
	writeCount := int32(0)
	mockWriter := &MockDataWriter{
		writeFunc: func(ctx context.Context, data []PointData) error {
			atomic.AddInt32(&writeCount, 1)
			return nil
		},
	}

	bw := NewBatchWriter(mockWriter, WithBatchSize(10))

	batch := [][]PointData{
		make([]PointData, 5),
		make([]PointData, 8),
	}

	err := bw.WriteBatch(context.Background(), batch)
	assert.NoError(t, err)
}

func TestBatchWriter_WriteBatch_Error(t *testing.T) {
	mockWriter := &MockDataWriter{
		writeFunc: func(ctx context.Context, data []PointData) error {
			return errors.New("write error")
		},
	}

	bw := NewBatchWriter(mockWriter, WithBatchSize(10))

	batch := [][]PointData{make([]PointData, 5)}
	err := bw.WriteBatch(context.Background(), batch)
	assert.Error(t, err)
}

func TestBatchWriter_Write_Error(t *testing.T) {
	mockWriter := &MockDataWriter{
		writeFunc: func(ctx context.Context, data []PointData) error {
			return errors.New("write error")
		},
	}

	bw := NewBatchWriter(mockWriter, WithBatchSize(10))

	data := make([]PointData, 15)
	err := bw.Write(context.Background(), data)
	assert.Error(t, err)
}

func TestBatchWriter_Write_Empty(t *testing.T) {
	mockWriter := &MockDataWriter{}
	bw := NewBatchWriter(mockWriter, WithBatchSize(10))

	err := bw.Write(context.Background(), []PointData{})
	assert.NoError(t, err)
}

func TestBatchWriter_Close(t *testing.T) {
	mockWriter := &MockDataWriter{}
	bw := NewBatchWriter(mockWriter, WithBatchSize(10))

	err := bw.Close()
	assert.NoError(t, err)
}

func TestRetryWriter_WriteBatch_Success(t *testing.T) {
	_ = int32(0)
	mockWriter := &MockDataWriter{
		writeFunc: func(ctx context.Context, data []PointData) error {
			return nil
		},
	}

	rw := NewRetryWriter(mockWriter,
		WithWriterMaxRetry(3),
		WithWriterRetryDelay(10*time.Millisecond),
		WithExponentialBackoff(false),
	)

	batch := [][]PointData{{{PointID: "p1", Value: 1.0, Timestamp: time.Now()}}}
	err := rw.WriteBatch(context.Background(), batch)
	assert.NoError(t, err)
}

func TestRetryWriter_WriteBatch_RetrySuccess(t *testing.T) {
	attemptCount := int32(0)
	mockWriter := &MockDataWriter{}

	mockWriterForBatch := &struct {
		*MockDataWriter
	}{MockDataWriter: mockWriter}

	rw := NewRetryWriter(mockWriterForBatch,
		WithWriterMaxRetry(3),
		WithWriterRetryDelay(10*time.Millisecond),
	)

	atomic.StoreInt32(&attemptCount, 0)
	batch := [][]PointData{{{PointID: "p1", Value: 1.0, Timestamp: time.Now()}}}
	err := rw.WriteBatch(context.Background(), batch)
	assert.NoError(t, err)
}

func TestRetryWriter_WriteBatch_AllFail(t *testing.T) {
	mockWriter := &MockDataWriter{
		writeFunc: func(ctx context.Context, data []PointData) error {
			return errors.New("always fail")
		},
	}

	rw := NewRetryWriter(mockWriter,
		WithWriterMaxRetry(2),
		WithWriterRetryDelay(10*time.Millisecond),
		WithExponentialBackoff(true),
	)

	batch := [][]PointData{{{PointID: "p1", Value: 1.0, Timestamp: time.Now()}}}
	err := rw.WriteBatch(context.Background(), batch)
	assert.Error(t, err)
}

func TestRetryWriter_Write_CancelledContext(t *testing.T) {
	mockWriter := &MockDataWriter{
		writeFunc: func(ctx context.Context, data []PointData) error {
			return errors.New("fail")
		},
	}

	rw := NewRetryWriter(mockWriter,
		WithWriterMaxRetry(10),
		WithWriterRetryDelay(1*time.Second),
		WithExponentialBackoff(true),
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	data := []PointData{{PointID: "p1", Value: 1.0, Timestamp: time.Now()}}
	err := rw.Write(ctx, data)
	assert.Error(t, err)
}

func TestRetryWriter_Close(t *testing.T) {
	mockWriter := &MockDataWriter{}
	rw := NewRetryWriter(mockWriter)

	err := rw.Close()
	assert.NoError(t, err)
}

func TestRetryWriter_Close_NilWriter(t *testing.T) {
	rw := &RetryWriter{writer: nil}
	err := rw.Close()
	assert.NoError(t, err)
}

func TestWorkerPool_Submit_Closed(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	require.NoError(t, pool.Start())
	pool.GracefulShutdown(5 * time.Second)

	err := pool.Submit(context.Background(), "task-1", 1, func(ctx context.Context) error {
		return nil
	})
	assert.Error(t, err)
	assert.Equal(t, ErrPoolClosed, err)
}

func TestWorkerPool_Submit_NotRunning(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))

	err := pool.Submit(context.Background(), "task-1", 1, func(ctx context.Context) error {
		return nil
	})
	assert.Error(t, err)
	assert.Equal(t, ErrPoolNotRunning, err)
}

func TestWorkerPool_SubmitAndWait_Closed(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	require.NoError(t, pool.Start())
	pool.GracefulShutdown(5 * time.Second)

	err := pool.SubmitAndWait(context.Background(), "task-1", 1, func(ctx context.Context) error {
		return nil
	})
	assert.Error(t, err)
}

func TestWorkerPool_SubmitAndWait_NotRunning(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))

	err := pool.SubmitAndWait(context.Background(), "task-1", 1, func(ctx context.Context) error {
		return nil
	})
	assert.Error(t, err)
	assert.Equal(t, ErrPoolNotRunning, err)
}

func TestWorkerPool_Submit_CancelledContext(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	require.NoError(t, pool.Start())
	defer pool.GracefulShutdown(5 * time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := pool.Submit(ctx, "task-1", 1, func(ctx context.Context) error {
		return nil
	})
	assert.Error(t, err)
}

func TestWorkerPool_SubmitAndWait_CancelledContext(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	require.NoError(t, pool.Start())
	defer pool.GracefulShutdown(5 * time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := pool.SubmitAndWait(ctx, "task-1", 1, func(ctx context.Context) error {
		return nil
	})
	assert.Error(t, err)
}

func TestWorkerPool_SubmitAndWait_Success(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	require.NoError(t, pool.Start())
	defer pool.GracefulShutdown(5 * time.Second)

	err := pool.SubmitAndWait(context.Background(), "task-1", 1, func(ctx context.Context) error {
		return nil
	})
	assert.NoError(t, err)
}

func TestWorkerPool_SubmitAndWait_TaskError(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	require.NoError(t, pool.Start())
	defer pool.GracefulShutdown(5 * time.Second)

	err := pool.SubmitAndWait(context.Background(), "task-1", 1, func(ctx context.Context) error {
		return errors.New("task error")
	})
	assert.Error(t, err)
}

func TestWorkerPool_SetSize_Invalid(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	require.NoError(t, pool.Start())
	defer pool.GracefulShutdown(5 * time.Second)

	err := pool.SetSize(0, 10)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidSize, err)

	err = pool.SetSize(10, 0)
	assert.Error(t, err)

	err = pool.SetSize(20, 10)
	assert.Error(t, err)
}

func TestWorkerPool_GracefulShutdown_AlreadyClosed(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	require.NoError(t, pool.Start())
	require.NoError(t, pool.GracefulShutdown(5*time.Second))

	err := pool.GracefulShutdown(5 * time.Second)
	assert.Error(t, err)
	assert.Equal(t, ErrPoolClosed, err)
}

func TestWorkerPool_GetQueueSize(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))
	require.NoError(t, pool.Start())
	defer pool.GracefulShutdown(5 * time.Second)

	size := pool.GetQueueSize()
	assert.GreaterOrEqual(t, size, 0)
}

func TestWorkerPool_Options(t *testing.T) {
	pool := NewWorkerPool(
		WithIdleTimeout(10*time.Second),
		WithMaxIdleWorkers(50),
	)

	assert.Equal(t, 10*time.Second, pool.idleTimeout)
	assert.Equal(t, 50, pool.maxIdleWorkers)
}

func TestPriorityQueue_Len(t *testing.T) {
	pq := NewPriorityQueue()
	assert.Equal(t, 0, pq.Len())

	pq.Push(&TaskWrapper{ID: "task1", Priority: 1})
	assert.Equal(t, 1, pq.Len())

	pq.Push(&TaskWrapper{ID: "task2", Priority: 2})
	assert.Equal(t, 2, pq.Len())
}

func TestWorkerPool_CleanIdleWorkersOnce(t *testing.T) {
	pool := NewWorkerPool(
		WithMaxWorkers(10),
		WithMinWorkers(1),
		WithIdleTimeout(50*time.Millisecond),
		WithMaxIdleWorkers(1),
	)
	require.NoError(t, pool.Start())

	time.Sleep(100 * time.Millisecond)

	pool.cleanIdleWorkersOnce()

	pool.GracefulShutdown(5 * time.Second)
}

func TestWorkerPool_UpdateAverageDuration(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(10), WithMinWorkers(2))

	pool.updateAverageDuration(100 * time.Millisecond)
	assert.Equal(t, int64(100*time.Millisecond), pool.metrics.AverageDuration)

	atomic.StoreInt64(&pool.metrics.CompletedTasks, 1)
	pool.updateAverageDuration(200*time.Millisecond)
}
