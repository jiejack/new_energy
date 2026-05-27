package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCronParser(t *testing.T) {
	p := NewCronParser()
	require.NotNil(t, p)
}

func TestCronParser_WithLocation(t *testing.T) {
	p := NewCronParser().WithLocation(time.UTC)
	require.NotNil(t, p)
}

func TestCronParser_Parse_Predefined(t *testing.T) {
	p := NewCronParser()

	tests := []string{"@yearly", "@annually", "@monthly", "@weekly", "@daily", "@midnight", "@hourly", "@minutely"}
	for _, input := range tests {
		expr, err := p.Parse(input)
		require.NoError(t, err, "failed to parse %s", input)
		require.NotNil(t, expr, "nil expression for %s", input)
	}
}

func TestCronParser_Parse_Every(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("@every 1h30m")
	require.NoError(t, err)
	require.NotNil(t, expr)
}

func TestCronParser_Parse_InvalidFields(t *testing.T) {
	p := NewCronParser()

	_, err := p.Parse("")
	assert.Error(t, err)

	_, err = p.Parse("1 2 3")
	assert.Error(t, err)

	_, err = p.Parse("1 2 3 4 5 6 7 8")
	assert.Error(t, err)
}

func TestCronParser_Parse_InvalidField(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 abc 10 * * *")
	assert.Error(t, err)
}

func TestCronParser_IsValid(t *testing.T) {
	p := NewCronParser()
	assert.True(t, p.IsValid("0 30 10 * * *"))
	assert.True(t, p.IsValid("@daily"))
	assert.False(t, p.IsValid("invalid"))
	assert.False(t, p.IsValid(""))
}

func TestCronExpression_Next(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 30 10 * * *")
	require.NoError(t, err)

	now := time.Date(2024, 1, 1, 9, 0, 0, 0, time.Local)
	next := expr.Next(now)
	require.NotNil(t, next)
	assert.Equal(t, 30, next.Minute())
	assert.Equal(t, 10, next.Hour())
}

func TestCronExpression_Prev(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 12 * * *")
	require.NoError(t, err)

	now := time.Date(2024, 1, 2, 15, 0, 0, 0, time.Local)
	prev := expr.Prev(now)
	require.NotNil(t, prev)
	assert.True(t, prev.Before(now))
}

func TestCronExpression_GetNextN(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 * * * *")
	require.NoError(t, err)

	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
	nexts := expr.GetNextN(now, 3)
	assert.Equal(t, 3, len(nexts))
}

func TestCronExpression_Validate(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 30 10 * * *")
	require.NoError(t, err)
	err = expr.Validate()
	assert.NoError(t, err)
}

func TestCronExpression_String(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 30 10 * * *")
	require.NoError(t, err)
	str := expr.String()
	assert.Contains(t, str, "30")
}

func TestCronExpression_GetDescription(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 30 10 * * *")
	require.NoError(t, err)
	desc := expr.GetDescription()
	assert.NotEmpty(t, desc)
}

func TestCronParserBuilder(t *testing.T) {
	builder := NewCronParserBuilder()
	require.NotNil(t, builder)
	builder.WithLocation(time.UTC)
	p := builder.Build()
	require.NotNil(t, p)
}

func TestDefaultExecutorConfig(t *testing.T) {
	cfg := DefaultExecutorConfig()
	assert.Equal(t, 100, cfg.MaxConcurrency)
	assert.Equal(t, 10000, cfg.QueueSize)
	assert.Equal(t, 30*time.Second, cfg.DefaultTimeout)
}

func TestNewTaskExecutor(t *testing.T) {
	executor := NewTaskExecutor(nil)
	require.NotNil(t, executor)
	assert.False(t, executor.IsRunning())
	assert.Equal(t, 0, executor.GetActiveCount())
	assert.Equal(t, 0, executor.GetQueueSize())
}

func TestTaskExecutor_StartStop(t *testing.T) {
	executor := NewTaskExecutor(nil)

	err := executor.Start()
	require.NoError(t, err)
	assert.True(t, executor.IsRunning())

	err = executor.Stop()
	require.NoError(t, err)
	assert.False(t, executor.IsRunning())
}

func TestTaskExecutor_Start_AlreadyRunning(t *testing.T) {
	executor := NewTaskExecutor(nil)
	executor.Start()

	err := executor.Start()
	assert.Equal(t, ErrExecutorRunning, err)

	executor.Stop()
}

func TestTaskExecutor_Stop_NotRunning(t *testing.T) {
	executor := NewTaskExecutor(nil)
	err := executor.Stop()
	assert.Equal(t, ErrExecutorNotRunning, err)
}

func TestTaskExecutor_RegisterHandler(t *testing.T) {
	executor := NewTaskExecutor(nil)
	executor.Start()
	defer executor.Stop()

	handler := func(ctx *ExecutionContext) (*ExecutionResult, error) {
		return &ExecutionResult{Success: true}, nil
	}
	executor.RegisterHandler("test_task", handler)
	executor.UnregisterHandler("test_task")
}

func TestTaskExecutor_Execute_NotRunning(t *testing.T) {
	executor := NewTaskExecutor(nil)
	_, err := executor.Execute(context.Background(), &ExecutionContext{TaskType: "test_task"})
	assert.Equal(t, ErrExecutorNotRunning, err)
}

func TestTaskExecutor_ExecuteSync_NotRunning(t *testing.T) {
	executor := NewTaskExecutor(nil)
	_, err := executor.ExecuteSync(context.Background(), &ExecutionContext{TaskType: "test_task"}, nil)
	assert.Equal(t, ErrExecutorNotRunning, err)
}

func TestTaskExecutor_GetRecords(t *testing.T) {
	executor := NewTaskExecutor(nil)
	records := executor.GetRecords("", 10)
	assert.Equal(t, 0, len(records))
}

func TestNewStatisticsTaskExecutor(t *testing.T) {
	executor := NewStatisticsTaskExecutor(nil)
	require.NotNil(t, executor)
}

func TestNewExecutorBuilder(t *testing.T) {
	builder := NewExecutorBuilder()
	require.NotNil(t, builder)

	builder.WithMaxConcurrency(5)
	builder.WithQueueSize(500)
	builder.WithDefaultTimeout(10 * time.Second)
	builder.WithWorkerCount(4)
	builder.WithPanicRecovery(true)
	builder.WithMetrics(true)

	executor := builder.Build()
	require.NotNil(t, executor)
}

func TestExecutionResult_Struct(t *testing.T) {
	result := &ExecutionResult{
		TaskID:    "t1",
		Success:   true,
		Data:      map[string]interface{}{"key": "value"},
		Error:     "",
		StartTime: time.Now(),
		EndTime:   time.Now(),
		Duration:  100 * time.Millisecond,
	}
	assert.Equal(t, "t1", result.TaskID)
	assert.True(t, result.Success)
}

func TestExecutionContext_Struct(t *testing.T) {
	ctx := &ExecutionContext{
		TaskID:     "t1",
		TaskName:   "test_task",
		TaskType:   "statistics",
		Config:     map[string]interface{}{"key": "value"},
		Timeout:    30 * time.Second,
		RetryCount: 3,
	}
	assert.Equal(t, "t1", ctx.TaskID)
	assert.Equal(t, "statistics", ctx.TaskType)
}

func TestRecord_Struct(t *testing.T) {
	record := &Record{
		Key:       "metric_1",
		Value:     42.0,
		Timestamp: time.Now().Unix(),
		Tags:      map[string]string{"device": "dev1"},
	}
	assert.Equal(t, "metric_1", record.Key)
}

func TestDefaultMonitorConfig(t *testing.T) {
	cfg := DefaultMonitorConfig()
	assert.Equal(t, 10000, cfg.MaxRecords)
	assert.Equal(t, 24*time.Hour, cfg.RecordRetention)
	assert.Equal(t, 10*time.Second, cfg.MetricsInterval)
	assert.Equal(t, 30*time.Second, cfg.AlertCheckInterval)
}

func TestNewTaskMonitor(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	require.NotNil(t, monitor)
	assert.False(t, monitor.IsRunning())
}

func TestTaskMonitor_StartStop(t *testing.T) {
	monitor := NewTaskMonitor(nil)

	err := monitor.Start()
	require.NoError(t, err)
	assert.True(t, monitor.IsRunning())

	err = monitor.Stop()
	require.NoError(t, err)
	assert.False(t, monitor.IsRunning())
}

func TestTaskMonitor_RecordExecution(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	monitor.Start()
	defer monitor.Stop()

	monitor.RecordExecution("t1", time.Now(), 100*time.Millisecond, true, nil)
}

func TestTaskMonitor_AddAlertRule(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	monitor.Start()
	defer monitor.Stop()

	rule := &AlertRule{
		Name:     "High Failure Rate",
		Type:     "failure",
		Severity: "critical",
		Enabled:  true,
	}
	monitor.AddAlertRule(rule)
	monitor.RemoveAlertRule(rule.ID)
}

func TestTaskMonitor_AddAlertHandler(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	monitor.Start()
	defer monitor.Stop()

	handler := func(alert *Alert) error { return nil }
	monitor.AddAlertHandler(handler)
}

func TestTaskMonitor_ResolveAlert(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	monitor.Start()
	defer monitor.Stop()

	alert := &Alert{
		ID:       "a1",
		TaskID:   "t1",
		Severity: "warning",
		Message:  "Test alert",
	}
	monitor.alerts["a1"] = alert
	monitor.ResolveAlert("a1")
}

func TestTaskMonitor_GetAlerts(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	alerts := monitor.GetAlerts("")
	assert.Equal(t, 0, len(alerts))
}

func TestTaskMonitor_GetTaskMetrics(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	monitor.Start()
	defer monitor.Stop()

	monitor.RecordExecution("t1", time.Now(), 100*time.Millisecond, true, nil)

	metrics, err := monitor.GetTaskMetrics("t1")
	require.NoError(t, err)
	require.NotNil(t, metrics)
}

func TestTaskMonitor_GetAllMetrics(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	metrics := monitor.GetAllMetrics()
	assert.NotNil(t, metrics)
}

func TestTaskMonitor_GetSummary(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	monitor.Start()
	defer monitor.Stop()

	summary := monitor.GetSummary()
	assert.NotNil(t, summary)
}

func TestTaskMonitor_GetRecords(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	records := monitor.GetRecords("", 10)
	assert.Equal(t, 0, len(records))
}

func TestTaskMonitor_ExportMetrics(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	data, err := monitor.ExportMetrics()
	require.NoError(t, err)
	assert.NotNil(t, data)
}

func TestNewMonitorBuilder(t *testing.T) {
	builder := NewMonitorBuilder()
	require.NotNil(t, builder)

	builder.WithMaxRecords(500)
	builder.WithRecordRetention(12 * time.Hour)
	builder.WithMetricsInterval(30 * time.Second)

	monitor := builder.Build()
	require.NotNil(t, monitor)
}

func TestDefaultDistributedSchedulerConfig(t *testing.T) {
	cfg := DefaultDistributedSchedulerConfig()
	assert.Equal(t, 5*time.Second, cfg.HeartbeatInterval)
	assert.Equal(t, 15*time.Second, cfg.HeartbeatTimeout)
	assert.Equal(t, 100, cfg.MaxConcurrentTasks)
	assert.Equal(t, 30*time.Second, cfg.LockTTL)
}

func TestAlertRule_Struct(t *testing.T) {
	rule := AlertRule{
		Name:     "Test",
		Type:     "failure",
		Severity: "critical",
		Enabled:  true,
	}
	assert.Equal(t, "Test", rule.Name)
	assert.True(t, rule.Enabled)
}

func TestAlert_Struct(t *testing.T) {
	alert := Alert{
		ID:        "a1",
		TaskID:    "t1",
		Severity:  "warning",
		Message:   "Test message",
		StartTime: time.Now(),
	}
	assert.Equal(t, "a1", alert.ID)
}

func TestTaskMetrics_Struct(t *testing.T) {
	metrics := &TaskMetrics{
		TaskID:          "t1",
		TaskType:        "stats",
		TotalExecutions: 100,
		SuccessCount:    95,
		FailureCount:    5,
		AverageDuration: 200 * time.Millisecond,
		MinDuration:     50 * time.Millisecond,
		MaxDuration:     5 * time.Second,
		SuccessRate:     0.95,
	}
	assert.Equal(t, "t1", metrics.TaskID)
	assert.Equal(t, int64(100), metrics.TotalExecutions)
}

func TestMonitorSummary_Struct(t *testing.T) {
	summary := &MonitorSummary{
		TotalTasks:         10,
		ActiveAlerts:       2,
		OverallSuccessRate: 0.95,
		AverageLatency:     200 * time.Millisecond,
	}
	assert.Equal(t, int64(10), summary.TotalTasks)
	assert.Equal(t, 2, summary.ActiveAlerts)
}

func TestDistributedTask_Struct(t *testing.T) {
	task := &DistributedTask{
		ID:            "dt1",
		Name:          "Test",
		CronExpr:      "0 * * * * *",
		Config:        map[string]interface{}{"key": "value"},
		Enabled:       true,
		AssignedNode:  "node1",
		LastRunTime:   time.Now(),
		NextRunTime:   time.Now().Add(time.Hour),
		RunCount:      5,
		LastError:     "",
	}
	assert.Equal(t, "dt1", task.ID)
	assert.Equal(t, "node1", task.AssignedNode)
}

func TestClusterNode_Struct(t *testing.T) {
	node := &ClusterNode{
		NodeID:        "node1",
		Address:       "10.0.0.1:8080",
		State:         "active",
		LastHeartbeat: time.Now(),
		TaskCount:     5,
	}
	assert.Equal(t, "node1", node.NodeID)
}

func TestTaskShard_Struct(t *testing.T) {
	shard := &TaskShard{
		TaskID:       "dt1",
		ShardIndex:   0,
		TotalShards:  2,
		AssignedNode: "node1",
		Status:       "running",
	}
	assert.Equal(t, "dt1", shard.TaskID)
}

func TestErrors(t *testing.T) {
	assert.True(t, errors.Is(ErrExecutorNotRunning, ErrExecutorNotRunning))
	assert.True(t, errors.Is(ErrExecutorRunning, ErrExecutorRunning))
	assert.True(t, errors.Is(ErrTaskTimeout, ErrTaskTimeout))
	assert.True(t, errors.Is(ErrTaskCancelled, ErrTaskCancelled))
	assert.True(t, errors.Is(ErrMaxConcurrency, ErrMaxConcurrency))
	assert.True(t, errors.Is(ErrInvalidTaskHandler, ErrInvalidTaskHandler))
	assert.True(t, errors.Is(ErrTaskHandlerPanic, ErrTaskHandlerPanic))
	assert.True(t, errors.Is(ErrInvalidCronExpression, ErrInvalidCronExpression))
	assert.True(t, errors.Is(ErrInvalidField, ErrInvalidField))
	assert.True(t, errors.Is(ErrUnsupportedCharacter, ErrUnsupportedCharacter))
}
