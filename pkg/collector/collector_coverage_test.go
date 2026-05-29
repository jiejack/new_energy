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

type mockDataWriter struct {
	writeErr  error
	batchErr  error
	written   int64
	batchSize int64
}

func (m *mockDataWriter) Write(ctx context.Context, data []PointData) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	atomic.AddInt64(&m.written, int64(len(data)))
	return nil
}

func (m *mockDataWriter) WriteBatch(ctx context.Context, batch [][]PointData) error {
	if m.batchErr != nil {
		return m.batchErr
	}
	for _, b := range batch {
		atomic.AddInt64(&m.batchSize, int64(len(b)))
	}
	return nil
}

func (m *mockDataWriter) Close() error { return nil }

func TestDataBuffer_NewCoverage(t *testing.T) {
	buf := NewDataBuffer(
		WithMaxSize(1000),
		WithFlushInterval(1*time.Second),
		WithFlushThreshold(100),
		WithMaxRetryCount(3),
		WithRetryDelay(100*time.Millisecond),
		WithEnableMetrics(true),
		WithEnableCompression(false),
	)
	require.NotNil(t, buf)
	assert.False(t, buf.IsRunning())
	assert.False(t, buf.IsClosed())
}

func TestDataBuffer_StartStopCoverage(t *testing.T) {
	buf := NewDataBuffer(
		WithMaxSize(1000),
		WithFlushInterval(1*time.Second),
		WithFlushThreshold(100),
	)
	writer := &mockDataWriter{}
	buf.SetWriter(writer)

	require.NoError(t, buf.Start())
	assert.True(t, buf.IsRunning())

	err := buf.Start()
	assert.Error(t, err)

	require.NoError(t, buf.Stop())
	assert.False(t, buf.IsRunning())

	err = buf.Stop()
	assert.Error(t, err)
}

func TestDataBuffer_WriteNotRunningCoverage(t *testing.T) {
	buf := NewDataBuffer(WithMaxSize(1000))
	writer := &mockDataWriter{}
	buf.SetWriter(writer)
	require.NoError(t, buf.Start())
	require.NoError(t, buf.Stop())
	err := buf.Write(PointData{PointID: "d1", Value: 42.0})
	assert.Error(t, err)
}

func TestDataBuffer_WriteAndFlushCoverage(t *testing.T) {
	buf := NewDataBuffer(
		WithMaxSize(1000),
		WithFlushInterval(10*time.Second),
		WithFlushThreshold(100),
	)
	writer := &mockDataWriter{}
	buf.SetWriter(writer)
	require.NoError(t, buf.Start())
	defer buf.Stop()

	err := buf.Write(PointData{PointID: "d1", Value: 42.0})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 1, buf.GetCurrentSize())

	require.NoError(t, buf.ForceFlush())
}

func TestDataBuffer_WriteBatchCoverage(t *testing.T) {
	buf := NewDataBuffer(
		WithMaxSize(1000),
		WithFlushInterval(10*time.Second),
		WithFlushThreshold(100),
	)
	writer := &mockDataWriter{}
	buf.SetWriter(writer)
	require.NoError(t, buf.Start())
	defer buf.Stop()

	data := []PointData{
		{PointID: "d1", Value: 1.0},
		{PointID: "d2", Value: 2.0},
	}
	err := buf.WriteBatch(data)
	require.NoError(t, err)
}

func TestDataBuffer_ForceFlushNoWriterCoverage(t *testing.T) {
	buf := NewDataBuffer(WithMaxSize(1000), WithFlushInterval(1*time.Second))
	writer := &mockDataWriter{}
	buf.SetWriter(writer)
	require.NoError(t, buf.Start())
	defer buf.Stop()

	buf.Write(PointData{PointID: "d1", Value: 1.0})
	err := buf.ForceFlush()
	assert.NoError(t, err)
}

func TestDataBuffer_ClearCoverage(t *testing.T) {
	buf := NewDataBuffer(WithMaxSize(1000), WithFlushInterval(10*time.Second))
	writer := &mockDataWriter{}
	buf.SetWriter(writer)
	require.NoError(t, buf.Start())
	defer buf.Stop()

	buf.Write(PointData{PointID: "d1", Value: 1.0})
	time.Sleep(100 * time.Millisecond)
	buf.Clear()
	assert.Equal(t, 0, buf.GetCurrentSize())
}

func TestDataBuffer_GetMetricsCoverage(t *testing.T) {
	buf := NewDataBuffer(
		WithMaxSize(1000),
		WithFlushInterval(1*time.Second),
		WithFlushThreshold(100),
		WithEnableMetrics(true),
	)
	writer := &mockDataWriter{}
	buf.SetWriter(writer)
	require.NoError(t, buf.Start())
	defer buf.Stop()

	buf.Write(PointData{PointID: "d1", Value: 1.0})
	metrics := buf.GetMetrics()
	assert.NotNil(t, metrics)
}

func TestBatchWriter_WriteCoverage(t *testing.T) {
	writer := &mockDataWriter{}
	bw := NewBatchWriter(writer, WithBatchSize(10), WithBatchTimeout(1*time.Second))
	ctx := context.Background()

	err := bw.Write(ctx, []PointData{{PointID: "d1", Value: 1.0}})
	require.NoError(t, err)

	require.NoError(t, bw.Close())
}

func TestBatchWriter_WriteBatchCoverage(t *testing.T) {
	writer := &mockDataWriter{}
	bw := NewBatchWriter(writer, WithBatchSize(10), WithBatchTimeout(1*time.Second))
	ctx := context.Background()

	data := [][]PointData{
		{{PointID: "d1", Value: 1.0}},
		{{PointID: "d2", Value: 2.0}},
	}
	err := bw.WriteBatch(ctx, data)
	require.NoError(t, err)

	require.NoError(t, bw.Close())
}

func TestBatchWriter_WriteErrorCoverage(t *testing.T) {
	writer := &mockDataWriter{batchErr: errors.New("write error")}
	bw := NewBatchWriter(writer, WithBatchSize(10), WithBatchTimeout(1*time.Second))
	ctx := context.Background()

	bw.Write(ctx, []PointData{{PointID: "d1", Value: 1.0}})
	err := bw.Close()
	assert.NoError(t, err)
}

func TestRetryWriter_WriteCoverage(t *testing.T) {
	writer := &mockDataWriter{}
	rw := NewRetryWriter(writer, WithWriterMaxRetry(3), WithWriterRetryDelay(10*time.Millisecond))
	ctx := context.Background()

	err := rw.Write(ctx, []PointData{{PointID: "d1", Value: 1.0}})
	require.NoError(t, err)

	require.NoError(t, rw.Close())
}

func TestRetryWriter_WriteBatchCoverage(t *testing.T) {
	writer := &mockDataWriter{}
	rw := NewRetryWriter(writer, WithWriterMaxRetry(3), WithWriterRetryDelay(10*time.Millisecond))
	ctx := context.Background()

	data := [][]PointData{{{PointID: "d1", Value: 1.0}}}
	err := rw.WriteBatch(ctx, data)
	require.NoError(t, err)

	require.NoError(t, rw.Close())
}

func TestRetryWriter_WriteBatchRetryCoverage(t *testing.T) {
	writer := &mockDataWriter{batchErr: errors.New("batch error")}
	rw := NewRetryWriter(writer, WithWriterMaxRetry(2), WithWriterRetryDelay(10*time.Millisecond))
	ctx := context.Background()

	data := [][]PointData{{{PointID: "d1", Value: 1.0}}}
	err := rw.WriteBatch(ctx, data)
	assert.Error(t, err)

	rw.Close()
}

func TestWorkerPool_NewCoverage(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(5), WithMinWorkers(2), WithPoolTaskQueueSize(10))
	require.NotNil(t, pool)
	assert.False(t, pool.IsRunning())
	assert.False(t, pool.IsClosed())
}

func TestWorkerPool_OptionsCoverage(t *testing.T) {
	pool := NewWorkerPool(
		WithMaxWorkers(10),
		WithMinWorkers(2),
		WithPoolTaskQueueSize(20),
		WithIdleTimeout(30*time.Second),
		WithMaxIdleWorkers(3),
	)
	require.NotNil(t, pool)
}

func TestWorkerPool_StartStopCoverage(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(5), WithMinWorkers(1), WithPoolTaskQueueSize(10))
	require.NoError(t, pool.Start())
	assert.True(t, pool.IsRunning())

	err := pool.Start()
	assert.Error(t, err)

	require.NoError(t, pool.GracefulShutdown(5 * time.Second))
	assert.False(t, pool.IsRunning())
}

func TestWorkerPool_StopNotRunningCoverage(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(5), WithMinWorkers(1), WithPoolTaskQueueSize(10))
	err := pool.GracefulShutdown(5 * time.Second)
	assert.NoError(t, err)
}

func TestWorkerPool_SubmitCoverage(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(5), WithMinWorkers(1), WithPoolTaskQueueSize(10))
	require.NoError(t, pool.Start())
	defer pool.GracefulShutdown(5 * time.Second)

	executed := int64(0)
	err := pool.Submit(context.Background(), "task-1", 1, func(ctx context.Context) error {
		atomic.AddInt64(&executed, 1)
		return nil
	})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int64(1), atomic.LoadInt64(&executed))
}

func TestWorkerPool_SubmitNotRunningCoverage(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(5), WithMinWorkers(1), WithPoolTaskQueueSize(10))
	err := pool.Submit(context.Background(), "task-1", 1, func(ctx context.Context) error {
		return nil
	})
	assert.Error(t, err)
}

func TestWorkerPool_SubmitAndWaitCoverage(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(5), WithMinWorkers(1), WithPoolTaskQueueSize(10))
	require.NoError(t, pool.Start())
	defer pool.GracefulShutdown(5 * time.Second)

	err := pool.SubmitAndWait(context.Background(), "task-wait", 1, func(ctx context.Context) error {
		return nil
	})
	require.NoError(t, err)
}

func TestWorkerPool_SubmitAndWait_ErrorCoverage(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(5), WithMinWorkers(1), WithPoolTaskQueueSize(10))
	require.NoError(t, pool.Start())
	defer pool.GracefulShutdown(5 * time.Second)

	err := pool.SubmitAndWait(context.Background(), "task-wait", 1, func(ctx context.Context) error {
		return errors.New("task error")
	})
	assert.Error(t, err)
}

func TestWorkerPool_SetSizeCoverage(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(5), WithMinWorkers(1), WithPoolTaskQueueSize(10))
	require.NoError(t, pool.Start())
	defer pool.GracefulShutdown(5 * time.Second)

	require.NoError(t, pool.SetSize(2, 10))
	metrics := pool.GetMetrics()
	assert.Equal(t, int32(2), metrics.TotalWorkers)
}

func TestWorkerPool_SetSize_InvalidCoverage(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(5), WithMinWorkers(1), WithPoolTaskQueueSize(10))
	require.NoError(t, pool.Start())
	defer pool.GracefulShutdown(5 * time.Second)

	err := pool.SetSize(0, 0)
	assert.Error(t, err)
}

func TestWorkerPool_GetMetricsCoverage(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(5), WithMinWorkers(1), WithPoolTaskQueueSize(10))
	require.NoError(t, pool.Start())
	defer pool.GracefulShutdown(5 * time.Second)

	metrics := pool.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, int32(1), metrics.TotalWorkers)
}

func TestWorkerPool_GetQueueSizeCoverage(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(5), WithMinWorkers(1), WithPoolTaskQueueSize(10))
	require.NoError(t, pool.Start())
	defer pool.GracefulShutdown(5 * time.Second)

	size := pool.GetQueueSize()
	assert.GreaterOrEqual(t, size, 0)
}

func TestCollectorPriorityQueueCoverage(t *testing.T) {
	pq := NewPriorityQueue()
	pq.Push(&TaskWrapper{ID: "t1", Priority: 5})
	pq.Push(&TaskWrapper{ID: "t2", Priority: 10})
	pq.Push(&TaskWrapper{ID: "t3", Priority: 3})

	assert.Equal(t, 3, pq.Len())

	task := pq.Pop()
	require.NotNil(t, task)
	assert.Equal(t, "t2", task.ID)

	task = pq.Pop()
	require.NotNil(t, task)
	assert.Equal(t, "t1", task.ID)

	task = pq.Pop()
	require.NotNil(t, task)
	assert.Equal(t, "t3", task.ID)

	task = pq.Pop()
	assert.Nil(t, task)
}

func newTestScheduler(t *testing.T) (*Scheduler, *WorkerPool) {
	t.Helper()
	pool := NewWorkerPool(WithMaxWorkers(5), WithMinWorkers(1), WithPoolTaskQueueSize(10))
	require.NoError(t, pool.Start())
	sched := NewScheduler(pool, WithMaxConcurrentTasks(10), WithTaskQueueSize(100), WithScheduleInterval(1*time.Second))
	return sched, pool
}

func TestScheduler_NewCoverage(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(5), WithMinWorkers(1), WithPoolTaskQueueSize(10))
	sched := NewScheduler(pool, WithMaxConcurrentTasks(10), WithTaskQueueSize(100), WithScheduleInterval(1*time.Second))
	require.NotNil(t, sched)
	assert.False(t, sched.IsRunning())
	assert.False(t, sched.IsClosed())
}

func TestScheduler_StartStopCoverage(t *testing.T) {
	sched, pool := newTestScheduler(t)
	defer pool.GracefulShutdown(5 * time.Second)

	require.NoError(t, sched.Start())
	assert.True(t, sched.IsRunning())

	err := sched.Start()
	assert.Error(t, err)

	require.NoError(t, sched.Stop())
	assert.False(t, sched.IsRunning())

	err = sched.Stop()
	assert.Error(t, err)
}

func TestScheduler_AddRemoveTaskCoverage(t *testing.T) {
	sched, pool := newTestScheduler(t)
	defer pool.GracefulShutdown(5 * time.Second)
	require.NoError(t, sched.Start())
	defer sched.Stop()

	task := &Task{
		ID:       "t1",
		Name:     "Test Task",
		Type:     TaskTypePeriodic,
		Interval: 1 * time.Hour,
	}
	require.NoError(t, sched.AddTask(task))
	assert.Equal(t, 1, sched.GetTaskCount())

	retrieved, err := sched.GetTask("t1")
	require.NoError(t, err)
	assert.Equal(t, "Test Task", retrieved.Name)

	_, err = sched.GetTask("nonexistent")
	assert.Error(t, err)

	allTasks := sched.GetAllTasks()
	assert.Len(t, allTasks, 1)

	require.NoError(t, sched.RemoveTask("t1"))
	assert.Equal(t, 0, sched.GetTaskCount())

	err = sched.RemoveTask("nonexistent")
	assert.Error(t, err)
}

func TestScheduler_AddInvalidTaskCoverage(t *testing.T) {
	sched, pool := newTestScheduler(t)
	defer pool.GracefulShutdown(5 * time.Second)
	require.NoError(t, sched.Start())
	defer sched.Stop()

	err := sched.AddTask(nil)
	assert.Error(t, err)

	err = sched.AddTask(&Task{ID: ""})
	assert.Error(t, err)
}

func TestScheduler_AddDuplicateTaskCoverage(t *testing.T) {
	sched, pool := newTestScheduler(t)
	defer pool.GracefulShutdown(5 * time.Second)
	require.NoError(t, sched.Start())
	defer sched.Stop()

	task := &Task{ID: "t1", Type: TaskTypePeriodic, Interval: 1 * time.Hour}
	require.NoError(t, sched.AddTask(task))
	err := sched.AddTask(task)
	assert.Error(t, err)
}

func TestScheduler_PauseResumeTaskCoverage(t *testing.T) {
	sched, pool := newTestScheduler(t)
	defer pool.GracefulShutdown(5 * time.Second)
	require.NoError(t, sched.Start())
	defer sched.Stop()

	task := &Task{ID: "t1", Type: TaskTypePeriodic, Interval: 1 * time.Hour}
	require.NoError(t, sched.AddTask(task))

	require.NoError(t, sched.PauseTask("t1"))
	t1, _ := sched.GetTask("t1")
	assert.Equal(t, TaskStatusCancelled, t1.Status)

	require.NoError(t, sched.ResumeTask("t1"))
	t1, _ = sched.GetTask("t1")
	assert.Equal(t, TaskStatusPending, t1.Status)

	err := sched.PauseTask("nonexistent")
	assert.Error(t, err)

	err = sched.ResumeTask("nonexistent")
	assert.Error(t, err)
}

func TestScheduler_UpdateTaskIntervalCoverage(t *testing.T) {
	sched, pool := newTestScheduler(t)
	defer pool.GracefulShutdown(5 * time.Second)
	require.NoError(t, sched.Start())
	defer sched.Stop()

	task := &Task{ID: "t1", Type: TaskTypePeriodic, Interval: 1 * time.Hour}
	require.NoError(t, sched.AddTask(task))

	require.NoError(t, sched.UpdateTaskInterval("t1", 30*time.Minute))
	t1, _ := sched.GetTask("t1")
	assert.Equal(t, 30*time.Minute, t1.Interval)

	err := sched.UpdateTaskInterval("nonexistent", 30*time.Minute)
	assert.Error(t, err)
}

func TestScheduler_RegisterCollectorCoverage(t *testing.T) {
	sched, pool := newTestScheduler(t)
	defer pool.GracefulShutdown(5 * time.Second)
	require.NoError(t, sched.Start())
	defer sched.Stop()

	collector := &mockCollector{}
	require.NoError(t, sched.RegisterCollector("c1", collector))
	assert.Equal(t, 1, sched.GetCollectorCount())

	sched.UnregisterCollector("c1")
	assert.Equal(t, 0, sched.GetCollectorCount())
}

func TestScheduler_SetEventHandlerCoverage(t *testing.T) {
	sched, pool := newTestScheduler(t)
	defer pool.GracefulShutdown(5 * time.Second)
	require.NoError(t, sched.Start())
	defer sched.Stop()

	handler := &mockEventHandler{}
	sched.SetEventHandler(handler)
}

func TestScheduler_TriggerEventCoverage(t *testing.T) {
	sched, pool := newTestScheduler(t)
	defer pool.GracefulShutdown(5 * time.Second)
	require.NoError(t, sched.Start())
	defer sched.Stop()

	handler := &mockEventHandler{}
	sched.SetEventHandler(handler)

	sched.TriggerEvent("test_event", "task-1", map[string]interface{}{"key": "value"})
}

func TestScheduler_GetMetricsCoverage(t *testing.T) {
	sched, pool := newTestScheduler(t)
	defer pool.GracefulShutdown(5 * time.Second)
	require.NoError(t, sched.Start())
	defer sched.Stop()

	metrics := sched.GetMetrics()
	assert.NotNil(t, metrics)
}

func TestScheduler_PeriodicTaskCoverage(t *testing.T) {
	sched, pool := newTestScheduler(t)
	defer pool.GracefulShutdown(5 * time.Second)
	require.NoError(t, sched.Start())
	defer sched.Stop()

	task := &Task{
		ID:       "periodic1",
		Type:     TaskTypePeriodic,
		Interval: 1 * time.Hour,
	}
	require.NoError(t, sched.AddTask(task))
}

func TestScheduler_OnceTaskCoverage(t *testing.T) {
	sched, pool := newTestScheduler(t)
	defer pool.GracefulShutdown(5 * time.Second)
	require.NoError(t, sched.Start())
	defer sched.Stop()

	task := &Task{
		ID:   "once1",
		Type: TaskTypeOnce,
	}
	require.NoError(t, sched.AddTask(task))
}

func TestPointData_StructCoverage(t *testing.T) {
	now := time.Now()
	pd := PointData{
		PointID:    "d1",
		Value:      25.5,
		Quality:    100,
		Timestamp:  now,
		Attributes: map[string]interface{}{"location": "room1"},
	}
	assert.Equal(t, "d1", pd.PointID)
	assert.Equal(t, 25.5, pd.Value)
}

func TestCollectResult_StructCoverage(t *testing.T) {
	result := CollectResult{
		Data: []PointData{
			{PointID: "d1", Value: 1.0},
			{PointID: "d2", Value: 2.0},
		},
		Error: nil,
	}
	assert.Len(t, result.Data, 2)
}

func TestCollectorMetrics_StructCoverage(t *testing.T) {
	metrics := CollectorMetrics{
		TotalCollects:   1000,
		FailedCollects:  10,
		LastCollectTime: time.Now(),
		AverageDuration: 50 * time.Millisecond,
	}
	assert.Equal(t, int64(1000), metrics.TotalCollects)
}

func TestTaskType_ConstantsCoverage(t *testing.T) {
	assert.Equal(t, TaskType(0), TaskTypePeriodic)
	assert.Equal(t, TaskType(1), TaskTypeEvent)
	assert.Equal(t, TaskType(2), TaskTypeOnce)
}

func TestTaskStatus_ConstantsCoverage(t *testing.T) {
	assert.Equal(t, TaskStatus(0), TaskStatusPending)
	assert.Equal(t, TaskStatus(1), TaskStatusRunning)
	assert.Equal(t, TaskStatus(2), TaskStatusCompleted)
	assert.Equal(t, TaskStatus(3), TaskStatusFailed)
}

type mockCollector struct{}

func (m *mockCollector) Initialize(ctx context.Context, config *CollectorConfig) error { return nil }
func (m *mockCollector) Start(ctx context.Context) error                              { return nil }
func (m *mockCollector) Stop(ctx context.Context) error                               { return nil }
func (m *mockCollector) Collect(ctx context.Context) (*CollectResult, error) {
	return &CollectResult{Data: []PointData{}}, nil
}
func (m *mockCollector) GetStatus() CollectorStatus    { return StatusRunning }
func (m *mockCollector) GetConfig() *CollectorConfig    { return &CollectorConfig{} }
func (m *mockCollector) GetMetrics() *CollectorMetrics  { return &CollectorMetrics{} }
func (m *mockCollector) HealthCheck(ctx context.Context) error { return nil }

type mockEventHandler struct{}

func (m *mockEventHandler) OnTaskComplete(result *TaskResult)        {}
func (m *mockEventHandler) OnTaskFailed(result *TaskResult)         {}
func (m *mockEventHandler) OnCollectorError(collectorID string, err error) {}
