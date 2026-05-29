package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCronParser_MustParse(t *testing.T) {
	p := NewCronParser()
	expr := p.MustParse("0 30 10 * * *")
	assert.NotNil(t, expr)
}

func TestCronParser_MustParse_Panic(t *testing.T) {
	p := NewCronParser()
	assert.Panics(t, func() {
		p.MustParse("invalid")
	})
}

func TestCronParser_Parse_MonthNames(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 1 jan *")
	require.NoError(t, err)
	assert.NotNil(t, expr)
}

func TestCronParser_Parse_WeekdayNames(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 * * mon")
	require.NoError(t, err)
	assert.NotNil(t, expr)
}

func TestCronParser_Parse_L_Day(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 L * *")
	require.NoError(t, err)
	assert.NotNil(t, expr)
}

func TestCronParser_Parse_L_InvalidField(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 L 0 * * *")
	assert.Error(t, err)
}

func TestCronParser_Parse_W_Field(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 15W * *")
	require.NoError(t, err)
	assert.NotNil(t, expr)
}

func TestCronParser_Parse_W_InvalidFormat(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0 0 AW * *")
	assert.Error(t, err)
}

func TestCronParser_Parse_NthWeekday(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 * * 5#3")
	require.NoError(t, err)
	assert.NotNil(t, expr)
}

func TestCronParser_Parse_NthWeekday_InvalidWeekday(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0 0 * * 7#3")
	assert.Error(t, err)
}

func TestCronParser_Parse_NthWeekday_InvalidNth(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0 0 * * 5#6")
	assert.Error(t, err)
}

func TestCronParser_Parse_EveryMinute(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("@every 5m")
	require.NoError(t, err)
	assert.NotNil(t, expr)
}

func TestCronParser_Parse_EveryHour(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("@every 2h")
	require.NoError(t, err)
	assert.NotNil(t, expr)
}

func TestCronParser_Parse_EveryDay(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("@every 24h")
	require.NoError(t, err)
	assert.NotNil(t, expr)
}

func TestCronParser_Parse_EverySecond(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("@every 30s")
	require.NoError(t, err)
	assert.NotNil(t, expr)
}

func TestCronParser_Parse_EveryInvalidDuration(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("@every invalid")
	assert.Error(t, err)
}

func TestCronParser_Parse_EveryTooShort(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("@every 500ms")
	assert.Error(t, err)
}

func TestCronParser_Parse_Step(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 */5 * * * *")
	require.NoError(t, err)
	assert.NotNil(t, expr)
}

func TestCronParser_Parse_Range(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 9-17 * * *")
	require.NoError(t, err)
	assert.NotNil(t, expr)
}

func TestCronParser_Parse_RangeInvalid(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0 17-9 * * *")
	assert.Error(t, err)
}

func TestCronParser_Parse_InvalidStepFormat(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0/0 * * * *")
	assert.Error(t, err)
}

func TestCronParser_Parse_Comma(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 9,12,18 * * *")
	require.NoError(t, err)
	assert.NotNil(t, expr)
}

func TestCronExpression_NextAfter(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 12 * * *")
	require.NoError(t, err)
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.Local)
	next := expr.NextAfter(now)
	assert.True(t, next.After(now) || next.Equal(now))
}

func TestCronExpression_GetSeconds(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 30 10 * * *")
	require.NoError(t, err)
	secs := expr.GetSeconds()
	assert.Contains(t, secs, 0)
}

func TestCronExpression_GetMinutes(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 30 10 * * *")
	require.NoError(t, err)
	mins := expr.GetMinutes()
	assert.Contains(t, mins, 30)
}

func TestCronExpression_GetHours(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 30 10 * * *")
	require.NoError(t, err)
	hours := expr.GetHours()
	assert.Contains(t, hours, 10)
}

func TestCronExpression_GetDays(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 15 * *")
	require.NoError(t, err)
	days := expr.GetDays()
	assert.Contains(t, days, 15)
}

func TestCronExpression_GetMonths(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 1 6 *")
	require.NoError(t, err)
	months := expr.GetMonths()
	assert.Contains(t, months, 6)
}

func TestCronExpression_GetWeekdays(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 * * 1")
	require.NoError(t, err)
	weekdays := expr.GetWeekdays()
	assert.Contains(t, weekdays, 1)
}

func TestCov_CronExpression_GetDescription_Full(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("30 15 10 25 12 5")
	require.NoError(t, err)
	desc := expr.GetDescription()
	assert.NotEmpty(t, desc)
}

func TestTaskExecutor_RegisterHandler_Nil(t *testing.T) {
	executor := NewTaskExecutor(nil)
	err := executor.RegisterHandler("test", nil)
	assert.Equal(t, ErrInvalidTaskHandler, err)
}

func TestTaskExecutor_Execute_NoHandler(t *testing.T) {
	executor := NewTaskExecutor(nil)
	executor.Start()
	defer executor.Stop()
	_, err := executor.Execute(context.Background(), &ExecutionContext{TaskType: "nonexistent"})
	assert.Error(t, err)
}

func TestTaskExecutor_Execute_Success(t *testing.T) {
	cfg := DefaultExecutorConfig()
	cfg.WorkerCount = 1
	executor := NewTaskExecutor(cfg)
	executor.RegisterHandler("test", func(ctx *ExecutionContext) (*ExecutionResult, error) {
		return &ExecutionResult{Success: true, Data: "ok"}, nil
	})
	executor.Start()
	defer executor.Stop()
	result, err := executor.Execute(context.Background(), &ExecutionContext{
		TaskID:   "t1",
		TaskType: "test",
		Timeout:  5 * time.Second,
	})
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestTaskExecutor_Execute_HandlerError(t *testing.T) {
	cfg := DefaultExecutorConfig()
	cfg.WorkerCount = 1
	executor := NewTaskExecutor(cfg)
	executor.RegisterHandler("fail", func(ctx *ExecutionContext) (*ExecutionResult, error) {
		return nil, assert.AnError
	})
	executor.Start()
	defer executor.Stop()
	result, err := executor.Execute(context.Background(), &ExecutionContext{
		TaskID:   "t2",
		TaskType: "fail",
		Timeout:  5 * time.Second,
	})
	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
}

func TestTaskExecutor_ExecuteSync_Success(t *testing.T) {
	cfg := DefaultExecutorConfig()
	cfg.WorkerCount = 1
	executor := NewTaskExecutor(cfg)
	executor.Start()
	defer executor.Stop()
	handler := func(ctx *ExecutionContext) (*ExecutionResult, error) {
		return &ExecutionResult{Success: true}, nil
	}
	result, err := executor.ExecuteSync(context.Background(), &ExecutionContext{
		TaskID:   "sync-1",
		TaskType: "sync",
		Timeout:  5 * time.Second,
	}, handler)
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestTaskExecutor_ExecuteSync_PanicRecovery(t *testing.T) {
	cfg := DefaultExecutorConfig()
	cfg.WorkerCount = 1
	cfg.PanicRecovery = true
	executor := NewTaskExecutor(cfg)
	executor.Start()
	defer executor.Stop()
	handler := func(ctx *ExecutionContext) (*ExecutionResult, error) {
		panic("test panic")
	}
	result, err := executor.ExecuteSync(context.Background(), &ExecutionContext{
		TaskID:   "panic-1",
		TaskType: "panic",
		Timeout:  5 * time.Second,
	}, handler)
	_ = err
	if result != nil {
		assert.False(t, result.Success)
	}
}

func TestTaskExecutor_Execute_Cancelled(t *testing.T) {
	cfg := DefaultExecutorConfig()
	cfg.QueueSize = 1
	cfg.WorkerCount = 1
	executor := NewTaskExecutor(cfg)
	executor.Start()
	defer executor.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := executor.Execute(ctx, &ExecutionContext{TaskType: "test", Timeout: 5 * time.Second})
	assert.Error(t, err)
}

func TestStatisticsTaskExecutor_StartStop(t *testing.T) {
	ste := NewStatisticsTaskExecutor(nil)
	err := ste.Start()
	require.NoError(t, err)
	err = ste.Stop()
	require.NoError(t, err)
}

func TestStatisticsTaskExecutor_Execute(t *testing.T) {
	cfg := DefaultExecutorConfig()
	cfg.WorkerCount = 1
	ste := NewStatisticsTaskExecutor(cfg)
	ste.Start()
	defer ste.Stop()
	result, err := ste.Execute(context.Background(), &ExecutionContext{
		TaskID:   "agg-1",
		TaskType: "aggregation",
		Timeout:  5 * time.Second,
		Config:   map[string]interface{}{"timeRange": "1h", "granularity": "5m"},
	})
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestStatisticsTaskExecutor_Execute_Statistics(t *testing.T) {
	cfg := DefaultExecutorConfig()
	cfg.WorkerCount = 1
	ste := NewStatisticsTaskExecutor(cfg)
	ste.Start()
	defer ste.Stop()
	result, err := ste.Execute(context.Background(), &ExecutionContext{
		TaskID:   "stat-1",
		TaskType: "statistics",
		Timeout:  5 * time.Second,
		Config:   map[string]interface{}{"statType": "avg", "timeRange": "24h"},
	})
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestStatisticsTaskExecutor_Execute_Cleanup(t *testing.T) {
	cfg := DefaultExecutorConfig()
	cfg.WorkerCount = 1
	ste := NewStatisticsTaskExecutor(cfg)
	ste.Start()
	defer ste.Stop()
	result, err := ste.Execute(context.Background(), &ExecutionContext{
		TaskID:   "clean-1",
		TaskType: "cleanup",
		Timeout:  5 * time.Second,
		Config:   map[string]interface{}{"retentionDays": 30, "tableName": "stats"},
	})
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestStatisticsTaskExecutor_Execute_Report(t *testing.T) {
	cfg := DefaultExecutorConfig()
	cfg.WorkerCount = 1
	ste := NewStatisticsTaskExecutor(cfg)
	ste.Start()
	defer ste.Stop()
	result, err := ste.Execute(context.Background(), &ExecutionContext{
		TaskID:   "report-1",
		TaskType: "report",
		Timeout:  5 * time.Second,
		Config:   map[string]interface{}{"reportType": "daily", "format": "pdf"},
	})
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestStatisticsTaskExecutor_Execute_Sync(t *testing.T) {
	cfg := DefaultExecutorConfig()
	cfg.WorkerCount = 1
	ste := NewStatisticsTaskExecutor(cfg)
	ste.Start()
	defer ste.Stop()
	result, err := ste.Execute(context.Background(), &ExecutionContext{
		TaskID:   "sync-1",
		TaskType: "sync",
		Timeout:  5 * time.Second,
		Config:   map[string]interface{}{"source": "db1", "target": "db2"},
	})
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestStatisticsTaskExecutor_RegisterHandler(t *testing.T) {
	ste := NewStatisticsTaskExecutor(nil)
	err := ste.RegisterHandler("custom", func(ctx *ExecutionContext) (*ExecutionResult, error) {
		return &ExecutionResult{Success: true}, nil
	})
	assert.NoError(t, err)
}

func TestStatisticsTaskExecutor_GetExecutor(t *testing.T) {
	ste := NewStatisticsTaskExecutor(nil)
	ex := ste.GetExecutor()
	assert.NotNil(t, ex)
}

func TestExecutorBuilder_BuildStatistics(t *testing.T) {
	ste := NewExecutorBuilder().
		WithMaxConcurrency(5).
		WithQueueSize(100).
		WithDefaultTimeout(10 * time.Second).
		WithWorkerCount(2).
		WithPanicRecovery(true).
		WithMetrics(false).
		BuildStatistics()
	require.NotNil(t, ste)
}

func TestTaskMonitor_Start_AlreadyRunning(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	monitor.Start()
	err := monitor.Start()
	assert.Equal(t, ErrMonitorRunning, err)
	monitor.Stop()
}

func TestTaskMonitor_Stop_NotRunning(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	err := monitor.Stop()
	assert.Equal(t, ErrMonitorNotRunning, err)
}

func TestTaskMonitor_RecordExecution_Failure(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	monitor.Start()
	defer monitor.Stop()
	monitor.RecordExecution("t1", time.Now(), 100*time.Millisecond, false, &ExecutionResult{
		TaskID: "t1",
		Error:  "something failed",
	})
	metrics, err := monitor.GetTaskMetrics("t1")
	require.NoError(t, err)
	assert.Equal(t, int64(1), metrics.FailureCount)
	assert.Equal(t, 1, metrics.ConsecutiveFailures)
}

func TestTaskMonitor_RecordExecution_Timeout(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	monitor.Start()
	defer monitor.Stop()
	record := &TaskExecutionRecord{
		TaskID:    "t2",
		StartTime: time.Now(),
		Duration:  5 * time.Second,
		Status:    ExecutionStatusTimeout,
	}
	monitor.addRecord(record)
	monitor.updateMetrics(record)
	metrics, err := monitor.GetTaskMetrics("t2")
	require.NoError(t, err)
	assert.Equal(t, int64(1), metrics.TimeoutCount)
}

func TestTaskMonitor_AddAlertRule_Nil(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	err := monitor.AddAlertRule(nil)
	assert.Equal(t, ErrInvalidAlertConfig, err)
}

func TestTaskMonitor_AddAlertRule_EmptyID(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	err := monitor.AddAlertRule(&AlertRule{Name: "test"})
	assert.Equal(t, ErrInvalidAlertConfig, err)
}

func TestTaskMonitor_ResolveAlert_NotFound(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	err := monitor.ResolveAlert("nonexistent")
	assert.Equal(t, ErrAlertNotFound, err)
}

func TestTaskMonitor_GetTaskMetrics_NotFound(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	_, err := monitor.GetTaskMetrics("nonexistent")
	assert.Error(t, err)
}

func TestTaskMonitor_GetAlerts_WithStatus(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	monitor.alerts["a1"] = &Alert{ID: "a1", Status: AlertStatusActive}
	monitor.alerts["a2"] = &Alert{ID: "a2", Status: AlertStatusResolved}
	active := monitor.GetAlerts(AlertStatusActive)
	assert.Len(t, active, 1)
}

func TestCov_TaskMonitor_ExportMetrics(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	monitor.RecordExecution("t1", time.Now(), 100*time.Millisecond, true, nil)
	data, err := monitor.ExportMetrics()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestMonitorBuilder_AllOptions(t *testing.T) {
	monitor := NewMonitorBuilder().
		WithMetrics(true).
		WithAlerts(true).
		WithRecordRetention(48 * time.Hour).
		WithMaxRecords(5000).
		WithMetricsInterval(30 * time.Second).
		WithAlertCheckInterval(1 * time.Minute).
		WithNotifications(false).
		Build()
	require.NotNil(t, monitor)
}

func TestMonitorBuilder_WithMetrics_Disabled(t *testing.T) {
	monitor := NewMonitorBuilder().
		WithMetrics(false).
		WithAlerts(false).
		Build()
	require.NotNil(t, monitor)
	monitor.Start()
	time.Sleep(50 * time.Millisecond)
	monitor.Stop()
}

func TestDistributedSchedulerConfig_Defaults(t *testing.T) {
	cfg := DefaultDistributedSchedulerConfig()
	assert.Equal(t, 100*time.Millisecond, cfg.ScheduleInterval)
	assert.Equal(t, 30*time.Second, cfg.RebalanceInterval)
	assert.True(t, cfg.EnableAutoRebalance)
	assert.True(t, cfg.EnableMetrics)
	assert.Equal(t, 100, cfg.NodeCapacity)
}

func TestGenerateNodeID(t *testing.T) {
	id := generateNodeID()
	assert.NotEmpty(t, id)
}

func TestGenerateLockValue(t *testing.T) {
	val := generateLockValue()
	assert.NotEmpty(t, val)
}

func TestSortNodesByLoad(t *testing.T) {
	nodes := []*ClusterNode{
		{NodeID: "n1", Load: 0.8},
		{NodeID: "n2", Load: 0.3},
		{NodeID: "n3", Load: 0.5},
	}
	sortNodesByLoad(nodes)
	assert.Equal(t, 0.3, nodes[0].Load)
	assert.Equal(t, 0.5, nodes[1].Load)
	assert.Equal(t, 0.8, nodes[2].Load)
}

func TestTaskExecutor_Execute_DefaultTimeout(t *testing.T) {
	cfg := DefaultExecutorConfig()
	cfg.WorkerCount = 1
	executor := NewTaskExecutor(cfg)
	executor.RegisterHandler("test", func(ctx *ExecutionContext) (*ExecutionResult, error) {
		return &ExecutionResult{Success: true}, nil
	})
	executor.Start()
	defer executor.Stop()
	result, err := executor.Execute(context.Background(), &ExecutionContext{
		TaskID:   "t1",
		TaskType: "test",
	})
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestTaskExecutor_GetRecords_WithTaskID(t *testing.T) {
	cfg := DefaultExecutorConfig()
	cfg.WorkerCount = 1
	executor := NewTaskExecutor(cfg)
	executor.RegisterHandler("test", func(ctx *ExecutionContext) (*ExecutionResult, error) {
		return &ExecutionResult{Success: true}, nil
	})
	executor.Start()
	defer executor.Stop()
	executor.Execute(context.Background(), &ExecutionContext{TaskID: "r1", TaskType: "test", Timeout: 5 * time.Second})
	time.Sleep(100 * time.Millisecond)
	records := executor.GetRecords("r1", 10)
	assert.NotEmpty(t, records)
}

func TestCronExpression_Prev_Complex(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 9 * * 1-5")
	require.NoError(t, err)
	now := time.Date(2024, 6, 15, 14, 0, 0, 0, time.Local)
	prev := expr.Prev(now)
	assert.True(t, prev.Before(now))
}

func TestCronExpression_Next_WithWeekday(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 * * 1")
	require.NoError(t, err)
	now := time.Date(2024, 1, 1, 1, 0, 0, 0, time.Local)
	next := expr.Next(now)
	assert.Equal(t, time.Monday, next.Weekday())
}

func TestCronParser_Parse_QuestionMark(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 ? * *")
	require.NoError(t, err)
	assert.NotNil(t, expr)
}

func TestCronParser_Parse_SevenFields(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 * * * 2024")
	require.NoError(t, err)
	assert.NotNil(t, expr)
}

func TestCronParser_Parse_InvalidValue(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0 0 32 * *")
	assert.Error(t, err)
}

func TestCronParser_Parse_InvalidStepValue(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0/abc * * * *")
	assert.Error(t, err)
}

func TestCronParser_Parse_StepFormat(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0/ * * * *")
	assert.Error(t, err)
}

func TestCronParser_Parse_LDaySuffix(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0 0 5L * *")
	require.NoError(t, err)
}

func TestCronParser_Parse_LDaySuffix_Invalid(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0 0 abcL * *")
	assert.Error(t, err)
}

func TestCronParser_Parse_W_OutOfRange(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0 0 32W * *")
	assert.Error(t, err)
}

func TestCronParser_Parse_NthWeekday_InvalidFormat(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0 0 * * abc#3")
	assert.Error(t, err)
}

func TestCronParser_Parse_NthWeekday_InvalidNthFormat(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0 0 * * 5#abc")
	assert.Error(t, err)
}

func TestCov_CronExpression_GetNextN_Zero(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 31 2 *")
	require.NoError(t, err)
	nexts := expr.GetNextN(time.Now(), 3)
	_ = nexts
}
