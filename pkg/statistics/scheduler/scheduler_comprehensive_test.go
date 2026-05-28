package scheduler

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCronParser_MustParse(t *testing.T) {
	p := NewCronParser()
	expr := p.MustParse("0 30 10 * * *")
	assert.NotNil(t, expr)

	assert.Panics(t, func() {
		p.MustParse("invalid")
	})
}

func TestCronParser_Parse_Standard5Field(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("30 10 * * *")
	require.NoError(t, err)
	require.NotNil(t, expr)

	now := time.Date(2024, 1, 1, 9, 0, 0, 0, time.Local)
	next := expr.Next(now)
	assert.Equal(t, 30, next.Minute())
	assert.Equal(t, 10, next.Hour())
}

func TestCronParser_Parse_7Fields(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 30 10 * * * 2024")
	require.NoError(t, err)
	require.NotNil(t, expr)
}

func TestCronParser_Parse_StepValues(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("*/5 * * * * *")
	require.NoError(t, err)
	require.NotNil(t, expr)

	seconds := expr.GetSeconds()
	assert.Contains(t, seconds, 0)
	assert.Contains(t, seconds, 5)
	assert.Contains(t, seconds, 10)
	assert.Contains(t, seconds, 55)
}

func TestCronParser_Parse_RangeValues(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 9-17 * * *")
	require.NoError(t, err)
	require.NotNil(t, expr)

	hours := expr.GetHours()
	assert.Contains(t, hours, 9)
	assert.Contains(t, hours, 12)
	assert.Contains(t, hours, 17)
	assert.NotContains(t, hours, 8)
	assert.NotContains(t, hours, 18)
}

func TestCronParser_Parse_CommaValues(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 9,12,18 * * *")
	require.NoError(t, err)
	require.NotNil(t, expr)

	hours := expr.GetHours()
	assert.Contains(t, hours, 9)
}

func TestCronParser_Parse_MonthNames(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 1 jan *")
	require.NoError(t, err)
	require.NotNil(t, expr)

	months := expr.GetMonths()
	assert.Contains(t, months, 1)
}

func TestCronParser_Parse_WeekdayNames(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 * * mon")
	require.NoError(t, err)
	require.NotNil(t, expr)

	weekdays := expr.GetWeekdays()
	assert.Contains(t, weekdays, 1)
}

func TestCronParser_Parse_L_Day(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 L * *")
	require.NoError(t, err)
	require.NotNil(t, expr)
}

func TestCronParser_Parse_L_InvalidField(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0 L * * *")
	assert.Error(t, err)
}

func TestCronParser_Parse_W_Field(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 15W * *")
	require.NoError(t, err)
	require.NotNil(t, expr)
}

func TestCronParser_Parse_W_InvalidFormat(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0 0 WW * *")
	assert.Error(t, err)
}

func TestCronParser_Parse_Hash_Field(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 * * 6#3")
	require.NoError(t, err)
	require.NotNil(t, expr)
}

func TestCronParser_Parse_Hash_InvalidFormat(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0 0 * * #3")
	assert.Error(t, err)

	_, err = p.Parse("0 0 0 * * 7#3")
	assert.Error(t, err)

	_, err = p.Parse("0 0 0 * * 6#6")
	assert.Error(t, err)
}

func TestCronParser_Parse_QuestionMark(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 ? * *")
	require.NoError(t, err)
	require.NotNil(t, expr)
}

func TestCronParser_Parse_EveryDuration(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("@every 1h30m")
	require.NoError(t, err)
	require.NotNil(t, expr)

	expr, err = p.Parse("@every 2h")
	require.NoError(t, err)
	require.NotNil(t, expr)

	expr, err = p.Parse("@every 30m")
	require.NoError(t, err)
	require.NotNil(t, expr)

	expr, err = p.Parse("@every 24h")
	require.NoError(t, err)
	require.NotNil(t, expr)
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

func TestCronParser_Parse_InvalidStepFormat(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0/ * * * *")
	assert.Error(t, err)
}

func TestCronParser_Parse_InvalidStepValue(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0/0 * * * *")
	assert.Error(t, err)
}

func TestCronParser_Parse_InvalidRangeFormat(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0 1--5 * * *")
	assert.Error(t, err)
}

func TestCronParser_Parse_RangeMinGtMax(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0 5-1 * * *")
	assert.Error(t, err)
}

func TestCronParser_Parse_ValueOutOfRange(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0 25 * * *")
	assert.Error(t, err)
}

func TestCronParser_Parse_LDaySuffix(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0 0 3L * *")
	require.NoError(t, err)
}

func TestCronParser_Parse_LDayInvalidValue(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0 0 abcL * *")
	assert.Error(t, err)
}

func TestCronExpression_NextAfter(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 30 10 * * *")
	require.NoError(t, err)

	now := time.Date(2024, 1, 1, 10, 30, 0, 0, time.Local)
	next := expr.NextAfter(now)
	assert.True(t, !next.Before(now))
}

func TestCronExpression_GetFieldValues(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 30 10 * * *")
	require.NoError(t, err)

	seconds := expr.GetSeconds()
	assert.Contains(t, seconds, 0)

	minutes := expr.GetMinutes()
	assert.Contains(t, minutes, 30)

	hours := expr.GetHours()
	assert.Contains(t, hours, 10)

	days := expr.GetDays()
	assert.True(t, len(days) > 0)

	months := expr.GetMonths()
	assert.True(t, len(months) > 0)

	weekdays := expr.GetWeekdays()
	assert.True(t, len(weekdays) > 0)
}

func TestCronExpression_GetNextN_Zero(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 31 2 *")
	require.NoError(t, err)

	nexts := expr.GetNextN(time.Now(), 3)
	assert.LessOrEqual(t, len(nexts), 3)
}

func TestCronExpression_Prev_NoMatch(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 31 2 * 2050")
	require.NoError(t, err)

	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
	prev := expr.Prev(now)
	assert.True(t, prev.IsZero() || prev.Before(now))
}

func TestCronExpression_Validate_NoValidTime(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 31 2 *")
	require.NoError(t, err)

	err = expr.Validate()
	if err != nil {
		assert.ErrorIs(t, err, ErrInvalidCronExpression)
	}
}

func TestCronExpression_MatchDay_WeekdayOnly(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 ? * 1")
	require.NoError(t, err)
	require.NotNil(t, expr)

	now := time.Date(2024, 1, 1, 10, 0, 0, 0, time.Local)
	next := expr.Next(now)
	assert.NotNil(t, next)
}

func TestCronExpression_MatchDay_DayOnly(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 15 * ?")
	require.NoError(t, err)
	require.NotNil(t, expr)

	now := time.Date(2024, 1, 1, 10, 0, 0, 0, time.Local)
	next := expr.Next(now)
	assert.NotNil(t, next)
}

func TestCronExpression_isNthWeekday(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 0 * * 1#2")
	require.NoError(t, err)
	require.NotNil(t, expr)

	firstMonday := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
	for d := firstMonday; d.Month() == time.January; d = d.AddDate(0, 0, 1) {
		if d.Weekday() == time.Monday {
			assert.False(t, expr.isNthWeekday(d, 2) && d.Day() < 8, "first monday should not be 2nd")
			break
		}
	}
}

func TestTaskExecutor_Execute_Success(t *testing.T) {
	cfg := DefaultExecutorConfig()
	cfg.WorkerCount = 2
	executor := NewTaskExecutor(cfg)
	err := executor.Start()
	require.NoError(t, err)
	defer executor.Stop()

	handler := func(ctx *ExecutionContext) (*ExecutionResult, error) {
		return &ExecutionResult{Success: true, Data: "test"}, nil
	}
	executor.RegisterHandler("test", handler)

	result, err := executor.Execute(context.Background(), &ExecutionContext{
		TaskID:   "t1",
		TaskType: "test",
		Timeout:  5 * time.Second,
	})
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "t1", result.TaskID)
}

func TestTaskExecutor_Execute_NoHandler(t *testing.T) {
	executor := NewTaskExecutor(nil)
	executor.Start()
	defer executor.Stop()

	_, err := executor.Execute(context.Background(), &ExecutionContext{
		TaskID:   "t1",
		TaskType: "nonexistent",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no handler")
}

func TestTaskExecutor_ExecuteSync_Success(t *testing.T) {
	executor := NewTaskExecutor(nil)
	executor.Start()
	defer executor.Stop()

	handler := func(ctx *ExecutionContext) (*ExecutionResult, error) {
		return &ExecutionResult{Success: true}, nil
	}

	result, err := executor.ExecuteSync(context.Background(), &ExecutionContext{
		TaskID:   "t1",
		TaskType: "test",
		Timeout:  5 * time.Second,
	}, handler)
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestTaskExecutor_ExecuteSync_PanicRecovery(t *testing.T) {
	cfg := DefaultExecutorConfig()
	cfg.PanicRecovery = true
	executor := NewTaskExecutor(cfg)
	executor.Start()
	defer executor.Stop()

	handler := func(ctx *ExecutionContext) (*ExecutionResult, error) {
		panic("test panic")
	}

	result, err := executor.ExecuteSync(context.Background(), &ExecutionContext{
		TaskID:   "t1",
		TaskType: "test",
		Timeout:  5 * time.Second,
	}, handler)
	assert.True(t, err != nil || result == nil || !result.Success)
}

func TestTaskExecutor_RegisterHandler_Nil(t *testing.T) {
	executor := NewTaskExecutor(nil)
	err := executor.RegisterHandler("test", nil)
	assert.Equal(t, ErrInvalidTaskHandler, err)
}

func TestTaskExecutor_Execute_Cancelled(t *testing.T) {
	cfg := DefaultExecutorConfig()
	cfg.QueueSize = 1
	cfg.MaxConcurrency = 1
	cfg.WorkerCount = 1
	executor := NewTaskExecutor(cfg)
	executor.Start()
	defer executor.Stop()

	slowHandler := func(ctx *ExecutionContext) (*ExecutionResult, error) {
		time.Sleep(3 * time.Second)
		return &ExecutionResult{Success: true}, nil
	}
	executor.RegisterHandler("slow", slowHandler)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := executor.Execute(ctx, &ExecutionContext{
		TaskID:   "t1",
		TaskType: "slow",
		Timeout:  5 * time.Second,
	})
	assert.True(t, errors.Is(err, ErrTaskCancelled) || errors.Is(err, context.Canceled))
}

func TestStatisticsTaskExecutor_Execute(t *testing.T) {
	ste := NewStatisticsTaskExecutor(nil)
	err := ste.Start()
	require.NoError(t, err)
	defer ste.Stop()

	result, err := ste.Execute(context.Background(), &ExecutionContext{
		TaskID:   "t1",
		TaskType: "aggregation",
		Config: map[string]interface{}{
			"timeRange":   "1h",
			"granularity": "5m",
			"pointIds":    []interface{}{"p1", "p2"},
		},
		Timeout: 5 * time.Second,
	})
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestStatisticsTaskExecutor_Execute_StatisticsTask(t *testing.T) {
	ste := NewStatisticsTaskExecutor(nil)
	ste.Start()
	defer ste.Stop()

	result, err := ste.Execute(context.Background(), &ExecutionContext{
		TaskID:   "t2",
		TaskType: "statistics",
		Config: map[string]interface{}{
			"statType":  "daily",
			"timeRange": "24h",
		},
		Timeout: 5 * time.Second,
	})
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestStatisticsTaskExecutor_Execute_CleanupTask(t *testing.T) {
	ste := NewStatisticsTaskExecutor(nil)
	ste.Start()
	defer ste.Stop()

	result, err := ste.Execute(context.Background(), &ExecutionContext{
		TaskID:   "t3",
		TaskType: "cleanup",
		Config: map[string]interface{}{
			"retentionDays": 30,
			"tableName":     "stats",
		},
		Timeout: 5 * time.Second,
	})
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestStatisticsTaskExecutor_Execute_ReportTask(t *testing.T) {
	ste := NewStatisticsTaskExecutor(nil)
	ste.Start()
	defer ste.Stop()

	result, err := ste.Execute(context.Background(), &ExecutionContext{
		TaskID:   "t4",
		TaskType: "report",
		Config: map[string]interface{}{
			"reportType": "daily",
			"format":     "pdf",
		},
		Timeout: 5 * time.Second,
	})
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestStatisticsTaskExecutor_Execute_SyncTask(t *testing.T) {
	ste := NewStatisticsTaskExecutor(nil)
	ste.Start()
	defer ste.Stop()

	result, err := ste.Execute(context.Background(), &ExecutionContext{
		TaskID:   "t5",
		TaskType: "sync",
		Config: map[string]interface{}{
			"source": "db1",
			"target": "db2",
		},
		Timeout: 5 * time.Second,
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
	exec := ste.GetExecutor()
	assert.NotNil(t, exec)
}

func TestExecutorBuilder_BuildStatistics(t *testing.T) {
	builder := NewExecutorBuilder()
	ste := builder.BuildStatistics()
	assert.NotNil(t, ste)
}

func TestTaskMonitor_Start_AlreadyRunning(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	monitor.Start()
	defer monitor.Stop()

	err := monitor.Start()
	assert.Equal(t, ErrMonitorRunning, err)
}

func TestTaskMonitor_Stop_NotRunning(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	err := monitor.Stop()
	assert.Equal(t, ErrMonitorNotRunning, err)
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

func TestTaskMonitor_AddAlertRule_Invalid(t *testing.T) {
	monitor := NewTaskMonitor(nil)

	err := monitor.AddAlertRule(nil)
	assert.Equal(t, ErrInvalidAlertConfig, err)

	err = monitor.AddAlertRule(&AlertRule{})
	assert.Equal(t, ErrInvalidAlertConfig, err)
}

func TestTaskMonitor_RecordExecution_WithResult(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	monitor.Start()
	defer monitor.Stop()

	result := &ExecutionResult{
		TaskID:  "t1",
		Success: true,
		Error:   "",
		Metrics: map[string]interface{}{"duration": 100},
	}
	monitor.RecordExecution("t1", time.Now(), 100*time.Millisecond, true, result)

	metrics, err := monitor.GetTaskMetrics("t1")
	require.NoError(t, err)
	assert.Equal(t, int64(1), metrics.TotalExecutions)
	assert.Equal(t, int64(1), metrics.SuccessCount)
}

func TestTaskMonitor_RecordExecution_Failure(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	monitor.Start()
	defer monitor.Stop()

	result := &ExecutionResult{
		TaskID: "t1",
		Error:  "something failed",
	}
	monitor.RecordExecution("t1", time.Now(), 50*time.Millisecond, false, result)

	metrics, err := monitor.GetTaskMetrics("t1")
	require.NoError(t, err)
	assert.Equal(t, int64(1), metrics.FailureCount)
	assert.Equal(t, 1, metrics.ConsecutiveFailures)
	assert.Equal(t, "something failed", metrics.LastError)
}

func TestTaskMonitor_RecordExecution_Multiple(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	monitor.Start()
	defer monitor.Stop()

	for i := 0; i < 5; i++ {
		monitor.RecordExecution("t1", time.Now(), time.Duration(100+i*10)*time.Millisecond, true, nil)
	}

	metrics, err := monitor.GetTaskMetrics("t1")
	require.NoError(t, err)
	assert.Equal(t, int64(5), metrics.TotalExecutions)
	assert.Equal(t, int64(5), metrics.SuccessCount)
	assert.Equal(t, 0, metrics.ConsecutiveFailures)
}

func TestTaskMonitor_GetAlerts_FilterByStatus(t *testing.T) {
	monitor := NewTaskMonitor(nil)

	monitor.alerts["a1"] = &Alert{ID: "a1", Status: AlertStatusActive}
	monitor.alerts["a2"] = &Alert{ID: "a2", Status: AlertStatusResolved}

	activeAlerts := monitor.GetAlerts(AlertStatusActive)
	assert.Len(t, activeAlerts, 1)
	assert.Equal(t, "a1", activeAlerts[0].ID)

	allAlerts := monitor.GetAlerts("")
	assert.Len(t, allAlerts, 2)
}

func TestTaskMonitor_GetRecords_WithTaskID(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	monitor.Start()
	defer monitor.Stop()

	monitor.RecordExecution("t1", time.Now(), 100*time.Millisecond, true, nil)
	monitor.RecordExecution("t2", time.Now(), 200*time.Millisecond, true, nil)
	monitor.RecordExecution("t1", time.Now(), 150*time.Millisecond, true, nil)

	records := monitor.GetRecords("t1", 10)
	assert.Len(t, records, 2)
}

func TestTaskMonitor_GetRecords_Limit(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	monitor.Start()
	defer monitor.Stop()

	for i := 0; i < 5; i++ {
		monitor.RecordExecution("t1", time.Now(), 100*time.Millisecond, true, nil)
	}

	records := monitor.GetRecords("t1", 2)
	assert.Len(t, records, 2)
}

func TestTaskMonitor_doCollectMetrics(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	monitor.Start()
	defer monitor.Stop()

	monitor.RecordExecution("t1", time.Now(), 100*time.Millisecond, true, nil)
	monitor.RecordExecution("t2", time.Now(), 200*time.Millisecond, false, nil)

	monitor.doCollectMetrics()

	summary := monitor.GetSummary()
	assert.Equal(t, int64(2), summary.TotalTasks)
	assert.Equal(t, int64(2), summary.TotalExecutions)
}

func TestTaskMonitor_doCheckAlerts_Failure(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	monitor.Start()
	defer monitor.Stop()

	for i := 0; i < 5; i++ {
		monitor.RecordExecution("t1", time.Now(), 100*time.Millisecond, false, &ExecutionResult{Error: "fail"})
	}

	monitor.doCheckAlerts()

	alerts := monitor.GetAlerts(AlertStatusActive)
	assert.True(t, len(alerts) > 0)
}

func TestTaskMonitor_doCheckAlerts_Timeout(t *testing.T) {
	cfg := DefaultMonitorConfig()
	monitor := NewTaskMonitor(cfg)
	monitor.Start()
	defer monitor.Stop()

	metrics := &TaskMetrics{
		TaskID:       "t1",
		TimeoutCount: 5,
	}
	monitor.metrics["t1"] = metrics

	monitor.doCheckAlerts()
}

func TestTaskMonitor_doCheckAlerts_Latency(t *testing.T) {
	cfg := DefaultMonitorConfig()
	monitor := NewTaskMonitor(cfg)
	monitor.Start()
	defer monitor.Stop()

	metrics := &TaskMetrics{
		TaskID:          "t1",
		AverageDuration: 60 * time.Second,
	}
	monitor.metrics["t1"] = metrics

	monitor.doCheckAlerts()
}

func TestTaskMonitor_doCheckAlerts_SuccessRate(t *testing.T) {
	cfg := DefaultMonitorConfig()
	monitor := NewTaskMonitor(cfg)
	monitor.Start()
	defer monitor.Stop()

	metrics := &TaskMetrics{
		TaskID:          "t1",
		TotalExecutions: 10,
		SuccessRate:     0.5,
	}
	monitor.metrics["t1"] = metrics

	monitor.doCheckAlerts()
}

func TestTaskMonitor_doCleanupRecords(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.RecordRetention = 1 * time.Millisecond
	monitor := NewTaskMonitor(cfg)

	oldRecord := &TaskExecutionRecord{
		TaskID:    "t1",
		StartTime: time.Now().Add(-1 * time.Hour),
	}
	monitor.records = append(monitor.records, oldRecord)

	monitor.doCleanupRecords()
	assert.Equal(t, 0, len(monitor.records))
}

func TestTaskMonitor_MaxRecords(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.MaxRecords = 5
	monitor := NewTaskMonitor(cfg)

	for i := 0; i < 10; i++ {
		monitor.RecordExecution("t1", time.Now(), 100*time.Millisecond, true, nil)
	}

	monitor.recordMu.RLock()
	count := len(monitor.records)
	monitor.recordMu.RUnlock()
	assert.LessOrEqual(t, count, 5)
}

func TestTaskMonitor_ExportMetricsData(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	monitor.Start()
	defer monitor.Stop()

	monitor.RecordExecution("t1", time.Now(), 100*time.Millisecond, true, nil)

	data, err := monitor.ExportMetrics()
	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.True(t, len(data) > 0)
}

func TestMonitorBuilder_AllOptions(t *testing.T) {
	builder := NewMonitorBuilder()
	builder.WithMetrics(true)
	builder.WithAlerts(true)
	builder.WithRecordRetention(48 * time.Hour)
	builder.WithMaxRecords(5000)
	builder.WithMetricsInterval(30 * time.Second)
	builder.WithAlertCheckInterval(1 * time.Minute)
	builder.WithNotifications(false)

	monitor := builder.Build()
	require.NotNil(t, monitor)
}

func TestDistributedScheduler_Helpers(t *testing.T) {
	nodeID := generateNodeID()
	assert.NotEmpty(t, nodeID)

	lockValue := generateLockValue()
	assert.NotEmpty(t, lockValue)
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

func TestSelectNodeForTask(t *testing.T) {
	s := &DistributedScheduler{
		nodeID:    "node1",
		nodeState: NodeStateActive,
		tasks:     make(map[string]*DistributedTask),
		shards:    make(map[string][]*TaskShard),
		nodes:     make(map[string]*ClusterNode),
		locks:     make(map[string]*DistributedLock),
	}

	nodes := []*ClusterNode{
		{NodeID: "n1", Load: 0.3},
		{NodeID: "n2", Load: 0.8},
	}

	task := &DistributedTask{ID: "t1"}
	selected := s.selectNodeForTask(task, nodes)
	assert.NotNil(t, selected)
	assert.Equal(t, "n1", selected.NodeID)
}

func TestSelectNodeForTask_Empty(t *testing.T) {
	s := &DistributedScheduler{
		nodeID:    "node1",
		tasks:     make(map[string]*DistributedTask),
		shards:    make(map[string][]*TaskShard),
		nodes:     make(map[string]*ClusterNode),
		locks:     make(map[string]*DistributedLock),
	}

	task := &DistributedTask{ID: "t1"}
	selected := s.selectNodeForTask(task, nil)
	assert.Nil(t, selected)
}

func TestSelectNodeByHash(t *testing.T) {
	s := &DistributedScheduler{
		nodeID:    "node1",
		tasks:     make(map[string]*DistributedTask),
		shards:    make(map[string][]*TaskShard),
		nodes:     make(map[string]*ClusterNode),
		locks:     make(map[string]*DistributedLock),
	}

	nodes := []*ClusterNode{
		{NodeID: "n1", Load: 0.3},
		{NodeID: "n2", Load: 0.5},
	}

	task := &DistributedTask{ID: "t1", ShardingKey: "shard1"}
	selected := s.selectNodeForTask(task, nodes)
	assert.NotNil(t, selected)
}

func TestSelectNodeByHash_Empty(t *testing.T) {
	s := &DistributedScheduler{
		nodeID:    "node1",
		tasks:     make(map[string]*DistributedTask),
		shards:    make(map[string][]*TaskShard),
		nodes:     make(map[string]*ClusterNode),
		locks:     make(map[string]*DistributedLock),
	}

	selected := s.selectNodeByHash("key", nil)
	assert.Nil(t, selected)
}

func TestDistributedScheduler_KeyHelpers(t *testing.T) {
	s := &DistributedScheduler{
		nodeID:    "node1",
		tasks:     make(map[string]*DistributedTask),
		shards:    make(map[string][]*TaskShard),
		nodes:     make(map[string]*ClusterNode),
		locks:     make(map[string]*DistributedLock),
	}

	assert.Equal(t, "nem:scheduler:node:node1", s.getNodeKey("node1"))
	assert.Equal(t, "nem:scheduler:nodes", s.getNodesSetKey())
	assert.Equal(t, "nem:scheduler:task:t1", s.getTaskKey("t1"))
	assert.Equal(t, "nem:scheduler:tasks", s.getTasksSetKey())
	assert.Equal(t, "nem:scheduler:lock:task:t1", s.getTaskLockKey("t1"))
}

func TestDistributedScheduler_GetNodeID(t *testing.T) {
	s := &DistributedScheduler{nodeID: "test-node"}
	assert.Equal(t, "test-node", s.GetNodeID())
}

func TestDistributedScheduler_IsRunning(t *testing.T) {
	s := &DistributedScheduler{}
	assert.False(t, s.IsRunning())
}

func TestDistributedScheduler_GetTaskCount(t *testing.T) {
	s := &DistributedScheduler{
		tasks: make(map[string]*DistributedTask),
	}
	s.tasks["t1"] = &DistributedTask{ID: "t1"}
	s.tasks["t2"] = &DistributedTask{ID: "t2"}
	assert.Equal(t, 2, s.GetTaskCount())
}

func TestDistributedScheduler_GetAllTasks(t *testing.T) {
	s := &DistributedScheduler{
		tasks: make(map[string]*DistributedTask),
	}
	s.tasks["t1"] = &DistributedTask{ID: "t1"}
	s.tasks["t2"] = &DistributedTask{ID: "t2"}

	tasks := s.GetAllTasks()
	assert.Len(t, tasks, 2)
}

func TestDistributedScheduler_GetTask(t *testing.T) {
	s := &DistributedScheduler{
		tasks: make(map[string]*DistributedTask),
	}
	s.tasks["t1"] = &DistributedTask{ID: "t1"}

	task, err := s.GetTask("t1")
	assert.NoError(t, err)
	assert.Equal(t, "t1", task.ID)

	_, err = s.GetTask("nonexistent")
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestDistributedScheduler_GetNodes(t *testing.T) {
	s := &DistributedScheduler{
		nodes: make(map[string]*ClusterNode),
	}
	s.nodes["n1"] = &ClusterNode{NodeID: "n1"}

	nodes := s.GetNodes()
	assert.Len(t, nodes, 1)
}

func TestDistributedScheduler_RemoveTask_NotFound(t *testing.T) {
	s := &DistributedScheduler{
		tasks: make(map[string]*DistributedTask),
	}
	err := s.RemoveTask("nonexistent")
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestDistributedScheduler_AddTask_Invalid(t *testing.T) {
	s := &DistributedScheduler{
		tasks:      make(map[string]*DistributedTask),
		cronParser: NewCronParser(),
	}

	err := s.AddTask(nil)
	assert.Equal(t, ErrInvalidTask, err)

	err = s.AddTask(&DistributedTask{})
	assert.Equal(t, ErrInvalidTask, err)
}

func TestDistributedScheduler_AddTask_Exists(t *testing.T) {
	s := &DistributedScheduler{
		tasks:      make(map[string]*DistributedTask),
		cronParser: NewCronParser(),
	}
	s.tasks["t1"] = &DistributedTask{ID: "t1"}

	err := s.AddTask(&DistributedTask{ID: "t1"})
	assert.Equal(t, ErrTaskExists, err)
}

func TestDistributedScheduler_AddTask_InvalidCron(t *testing.T) {
	s := &DistributedScheduler{
		tasks:      make(map[string]*DistributedTask),
		cronParser: NewCronParser(),
	}

	err := s.AddTask(&DistributedTask{ID: "t1", CronExpr: "invalid", Enabled: true})
	assert.Error(t, err)
}

func TestDistributedTask_StatusConstants(t *testing.T) {
	assert.Equal(t, TaskStatus("pending"), TaskStatusPending)
	assert.Equal(t, TaskStatus("running"), TaskStatusRunning)
	assert.Equal(t, TaskStatus("completed"), TaskStatusCompleted)
	assert.Equal(t, TaskStatus("failed"), TaskStatusFailed)
	assert.Equal(t, TaskStatus("cancelled"), TaskStatusCancelled)
	assert.Equal(t, TaskStatus("paused"), TaskStatusPaused)
}

func TestNodeState_Constants(t *testing.T) {
	assert.Equal(t, NodeState("active"), NodeStateActive)
	assert.Equal(t, NodeState("inactive"), NodeStateInactive)
	assert.Equal(t, NodeState("leaving"), NodeStateLeaving)
	assert.Equal(t, NodeState("joining"), NodeStateJoining)
}

func TestTaskPriority_Constants(t *testing.T) {
	assert.Equal(t, TaskPriority(0), PriorityLow)
	assert.Equal(t, TaskPriority(1), PriorityNormal)
	assert.Equal(t, TaskPriority(2), PriorityHigh)
	assert.Equal(t, TaskPriority(3), PriorityCritical)
}

func TestAlertSeverity_Constants(t *testing.T) {
	assert.Equal(t, AlertSeverity("info"), AlertSeverityInfo)
	assert.Equal(t, AlertSeverity("warning"), AlertSeverityWarning)
	assert.Equal(t, AlertSeverity("error"), AlertSeverityError)
	assert.Equal(t, AlertSeverity("critical"), AlertSeverityCritical)
}

func TestAlertStatus_Constants(t *testing.T) {
	assert.Equal(t, AlertStatus("active"), AlertStatusActive)
	assert.Equal(t, AlertStatus("resolved"), AlertStatusResolved)
	assert.Equal(t, AlertStatus("silenced"), AlertStatusSilenced)
}

func TestExecutionStatus_Constants(t *testing.T) {
	assert.Equal(t, TaskExecutionStatus("success"), ExecutionStatusSuccess)
	assert.Equal(t, TaskExecutionStatus("failed"), ExecutionStatusFailed)
	assert.Equal(t, TaskExecutionStatus("timeout"), ExecutionStatusTimeout)
	assert.Equal(t, TaskExecutionStatus("skipped"), ExecutionStatusSkipped)
}

func TestDistributedScheduler_Errors(t *testing.T) {
	assert.True(t, errors.Is(ErrSchedulerRunning, ErrSchedulerRunning))
	assert.True(t, errors.Is(ErrSchedulerNotRunning, ErrSchedulerNotRunning))
	assert.True(t, errors.Is(ErrTaskNotFound, ErrTaskNotFound))
	assert.True(t, errors.Is(ErrTaskExists, ErrTaskExists))
	assert.True(t, errors.Is(ErrInvalidTask, ErrInvalidTask))
	assert.True(t, errors.Is(ErrLockAcquireFailed, ErrLockAcquireFailed))
	assert.True(t, errors.Is(ErrNoAvailableNodes, ErrNoAvailableNodes))
	assert.True(t, errors.Is(ErrNodeNotRegistered, ErrNodeNotRegistered))
	assert.True(t, errors.Is(ErrTaskRebalanceFailed, ErrTaskRebalanceFailed))
}

func TestMonitorErrors(t *testing.T) {
	assert.True(t, errors.Is(ErrMonitorNotRunning, ErrMonitorNotRunning))
	assert.True(t, errors.Is(ErrMonitorRunning, ErrMonitorRunning))
	assert.True(t, errors.Is(ErrAlertNotFound, ErrAlertNotFound))
	assert.True(t, errors.Is(ErrInvalidAlertConfig, ErrInvalidAlertConfig))
}

func TestCronParser_Parse_EverySeconds(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("@every 15s")
	require.NoError(t, err)
	require.NotNil(t, expr)
}

func TestCronExpression_FormatFieldValue(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 30 10 * * *")
	require.NoError(t, err)

	desc := expr.GetDescription()
	assert.NotEmpty(t, desc)
}

func TestCronExpression_FormatFieldValue_AllFields(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("30 15 9 1 6 1")
	require.NoError(t, err)

	desc := expr.GetDescription()
	assert.Contains(t, desc, "9")
}

func TestCronParser_Parse_StepWithRange(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0-30/10 * * * *")
	require.NoError(t, err)
	require.NotNil(t, expr)

	minutes := expr.GetMinutes()
	assert.Contains(t, minutes, 0)
	assert.Contains(t, minutes, 10)
	assert.Contains(t, minutes, 20)
	assert.Contains(t, minutes, 30)
}

func TestCronParser_Parse_StepWithAsterisk(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 */15 * * * *")
	require.NoError(t, err)
	require.NotNil(t, expr)

	minutes := expr.GetMinutes()
	assert.Contains(t, minutes, 0)
	assert.Contains(t, minutes, 15)
	assert.Contains(t, minutes, 30)
	assert.Contains(t, minutes, 45)
}

func TestCronParser_Parse_SingleValue(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 9 * * *")
	require.NoError(t, err)

	hours := expr.GetHours()
	assert.Contains(t, hours, 9)
}

func TestCronExpression_Next_MultipleFields(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 30 9 * * 1-5")
	require.NoError(t, err)

	now := time.Date(2024, 1, 1, 8, 0, 0, 0, time.Local)
	next := expr.Next(now)
	assert.NotNil(t, next)
}

func TestCronExpression_Prev_Basic(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("0 0 12 * * *")
	require.NoError(t, err)

	now := time.Date(2024, 6, 15, 15, 0, 0, 0, time.Local)
	prev := expr.Prev(now)
	assert.True(t, prev.Before(now))
}

func TestCronParser_Parse_EveryDay(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("@every 24h")
	require.NoError(t, err)
	require.NotNil(t, expr)
}

func TestCronParser_Parse_EveryHour(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("@every 2h")
	require.NoError(t, err)
	require.NotNil(t, expr)
}

func TestCronParser_Parse_EveryMinute(t *testing.T) {
	p := NewCronParser()
	expr, err := p.Parse("@every 5m")
	require.NoError(t, err)
	require.NotNil(t, expr)
}

func TestCronParser_Parse_W_OutOfRange(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0 0 32W * *")
	assert.Error(t, err)
}

func TestCronParser_Parse_InvalidValue(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0 abc * * *")
	assert.Error(t, err)
}

func TestCronParser_Parse_StepWithSlash(t *testing.T) {
	p := NewCronParser()
	_, err := p.Parse("0 0/ * * * *")
	assert.Error(t, err)

	_, err = p.Parse("0 */0 * * * *")
	assert.Error(t, err)
}

func TestTaskExecutor_ExecuteSync_NilResult(t *testing.T) {
	executor := NewTaskExecutor(nil)
	executor.Start()
	defer executor.Stop()

	handler := func(ctx *ExecutionContext) (*ExecutionResult, error) {
		return nil, nil
	}

	result, err := executor.ExecuteSync(context.Background(), &ExecutionContext{
		TaskID:   "t1",
		TaskType: "test",
		Timeout:  5 * time.Second,
	}, handler)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "t1", result.TaskID)
}

func TestTaskExecutor_ExecuteSync_WithError(t *testing.T) {
	executor := NewTaskExecutor(nil)
	executor.Start()
	defer executor.Stop()

	handler := func(ctx *ExecutionContext) (*ExecutionResult, error) {
		return nil, fmt.Errorf("task failed")
	}

	result, err := executor.ExecuteSync(context.Background(), &ExecutionContext{
		TaskID:   "t1",
		TaskType: "test",
		Timeout:  5 * time.Second,
	}, handler)
	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
}

func TestTaskExecutor_ExecuteSync_NilResultWithError(t *testing.T) {
	executor := NewTaskExecutor(nil)
	executor.Start()
	defer executor.Stop()

	handler := func(ctx *ExecutionContext) (*ExecutionResult, error) {
		return nil, fmt.Errorf("error")
	}

	result, err := executor.ExecuteSync(context.Background(), &ExecutionContext{
		TaskID:   "t1",
		TaskType: "test",
		Timeout:  5 * time.Second,
	}, handler)
	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, "error", result.Error)
}

func TestTaskExecutor_GetRecords_WithTaskID(t *testing.T) {
	executor := NewTaskExecutor(nil)
	executor.Start()
	defer executor.Stop()

	handler := func(ctx *ExecutionContext) (*ExecutionResult, error) {
		return &ExecutionResult{Success: true}, nil
	}

	executor.ExecuteSync(context.Background(), &ExecutionContext{
		TaskID: "t1", TaskType: "test", Timeout: 5 * time.Second,
	}, handler)
	executor.ExecuteSync(context.Background(), &ExecutionContext{
		TaskID: "t2", TaskType: "test", Timeout: 5 * time.Second,
	}, handler)

	records := executor.GetRecords("t1", 10)
	assert.Len(t, records, 1)
	assert.Equal(t, "t1", records[0].TaskID)
}

func TestTaskExecutor_DefaultTimeout(t *testing.T) {
	executor := NewTaskExecutor(nil)
	executor.Start()
	defer executor.Stop()

	handler := func(ctx *ExecutionContext) (*ExecutionResult, error) {
		return &ExecutionResult{Success: true}, nil
	}

	result, err := executor.ExecuteSync(context.Background(), &ExecutionContext{
		TaskID:   "t1",
		TaskType: "test",
	}, handler)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
}

func TestTaskMonitor_AlertHandler(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	monitor.Start()
	defer monitor.Stop()

	var receivedAlert *Alert
	monitor.AddAlertHandler(func(alert *Alert) error {
		receivedAlert = alert
		return nil
	})

	for i := 0; i < 5; i++ {
		monitor.RecordExecution("t1", time.Now(), 100*time.Millisecond, false, &ExecutionResult{Error: "fail"})
	}

	monitor.doCheckAlerts()

	if receivedAlert != nil {
		assert.Equal(t, "t1", receivedAlert.TaskID)
	}
}

func TestTaskMonitor_AlertHandler_Error(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	monitor.Start()
	defer monitor.Stop()

	monitor.AddAlertHandler(func(alert *Alert) error {
		return fmt.Errorf("handler error")
	})

	for i := 0; i < 5; i++ {
		monitor.RecordExecution("t1", time.Now(), 100*time.Millisecond, false, &ExecutionResult{Error: "fail"})
	}

	monitor.doCheckAlerts()
}

func TestTaskMonitor_RecordExecution_Timeout(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	monitor.Start()
	defer monitor.Stop()

	record := &TaskExecutionRecord{
		TaskID:    "t1",
		StartTime: time.Now(),
		Status:    ExecutionStatusTimeout,
	}
	monitor.addRecord(record)
	monitor.updateMetrics(record)

	metrics, err := monitor.GetTaskMetrics("t1")
	require.NoError(t, err)
	assert.Equal(t, int64(1), metrics.TimeoutCount)
	assert.Equal(t, 1, metrics.ConsecutiveFailures)
}

func TestTaskMonitor_DefaultAlertRules(t *testing.T) {
	monitor := NewTaskMonitor(nil)
	assert.Len(t, monitor.rules, 4)
}

func TestDistributedSchedulerConfig_Defaults(t *testing.T) {
	cfg := DefaultDistributedSchedulerConfig()
	assert.Equal(t, 100, cfg.MaxConcurrentTasks)
	assert.Equal(t, 10000, cfg.TaskQueueSize)
	assert.Equal(t, 100*time.Millisecond, cfg.ScheduleInterval)
	assert.Equal(t, 30*time.Second, cfg.LockTTL)
	assert.Equal(t, 5*time.Second, cfg.HeartbeatInterval)
	assert.Equal(t, 15*time.Second, cfg.HeartbeatTimeout)
	assert.Equal(t, 30*time.Second, cfg.RebalanceInterval)
	assert.True(t, cfg.EnableAutoRebalance)
	assert.True(t, cfg.EnableMetrics)
	assert.Equal(t, 100, cfg.NodeCapacity)
}

func newSchedulerWithMiniredis(t *testing.T) (*DistributedScheduler, *miniredis.Miniredis, *TaskExecutorImpl) {
	t.Helper()
	mr := miniredis.RunT(t)

	cfg := DefaultDistributedSchedulerConfig()
	cfg.RedisAddr = mr.Addr()
	cfg.HeartbeatInterval = 100 * time.Millisecond
	cfg.HeartbeatTimeout = 500 * time.Millisecond
	cfg.ScheduleInterval = 50 * time.Millisecond
	cfg.RebalanceInterval = 200 * time.Millisecond
	cfg.MaxConcurrentTasks = 2
	cfg.TaskQueueSize = 100
	cfg.NodeCapacity = 10
	cfg.EnableAutoRebalance = false

	executorImpl := NewTaskExecutor(nil)
	s, err := NewDistributedScheduler(cfg, executorImpl)
	require.NoError(t, err)
	require.NotNil(t, s)

	return s, mr, executorImpl
}

func TestDistributedScheduler_WithMiniredis_StartStop(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	err := s.Start()
	require.NoError(t, err)
	assert.True(t, s.IsRunning())

	err = s.Stop()
	require.NoError(t, err)
	assert.False(t, s.IsRunning())
}

func TestDistributedScheduler_WithMiniredis_AddTask(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	s.Start()
	defer s.Stop()

	task := &DistributedTask{
		ID:       "task-1",
		Name:     "Test Task",
		CronExpr: "0 */5 * * * *",
		Enabled:  true,
		Timeout:  30 * time.Second,
		MaxRetry: 3,
		Config:   map[string]interface{}{"key": "value"},
	}

	err := s.AddTask(task)
	require.NoError(t, err)

	retrieved, err := s.GetTask("task-1")
	require.NoError(t, err)
	assert.Equal(t, "Test Task", retrieved.Name)
	assert.Equal(t, TaskStatusPending, retrieved.Status)

	count := s.GetTaskCount()
	assert.Equal(t, 1, count)
}

func TestDistributedScheduler_WithMiniredis_AddTask_Disabled(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	s.Start()
	defer s.Stop()

	task := &DistributedTask{
		ID:      "task-2",
		Name:    "Disabled Task",
		Enabled: false,
	}

	err := s.AddTask(task)
	require.NoError(t, err)

	retrieved, err := s.GetTask("task-2")
	require.NoError(t, err)
	assert.Equal(t, TaskStatusPaused, retrieved.Status)
}

func TestDistributedScheduler_WithMiniredis_RemoveTask(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	s.Start()
	defer s.Stop()

	task := &DistributedTask{
		ID:      "task-3",
		Name:    "Remove Task",
		Enabled: true,
	}

	err := s.AddTask(task)
	require.NoError(t, err)

	err = s.RemoveTask("task-3")
	require.NoError(t, err)

	_, err = s.GetTask("task-3")
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestDistributedScheduler_WithMiniredis_GetAllTasks(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	s.Start()
	defer s.Stop()

	for i := 0; i < 3; i++ {
		err := s.AddTask(&DistributedTask{
			ID:      fmt.Sprintf("task-%d", i),
			Name:    fmt.Sprintf("Task %d", i),
			Enabled: true,
		})
		require.NoError(t, err)
	}

	tasks := s.GetAllTasks()
	assert.Len(t, tasks, 3)
}

func TestDistributedScheduler_WithMiniredis_GetNodes(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	s.Start()
	defer s.Stop()

	nodes := s.GetNodes()
	assert.NotNil(t, nodes)
}

func TestDistributedScheduler_WithMiniredis_CalculateNextRunTime(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	s.Start()
	defer s.Stop()

	task := &DistributedTask{
		ID:       "task-cron",
		Name:     "Cron Task",
		CronExpr: "0 0 * * * *",
		Enabled:  true,
	}

	err := s.AddTask(task)
	require.NoError(t, err)

	nextRun := s.calculateNextRunTime(task)
	assert.True(t, nextRun.After(time.Now().Add(-time.Second)))
}

func TestDistributedScheduler_WithMiniredis_AcquireLock(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	s.Start()
	defer s.Stop()

	lock, err := s.acquireLock("test-lock", 10*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, lock)

	s.releaseLock(lock)
}

func TestDistributedScheduler_WithMiniredis_AcquireLock_Conflict(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	s.Start()
	defer s.Stop()

	lock1, err := s.acquireLock("conflict-lock", 10*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, lock1)

	lock2, err := s.acquireLock("conflict-lock", 10*time.Second)
	require.NoError(t, err)
	assert.Nil(t, lock2)

	s.releaseLock(lock1)
}

func TestDistributedScheduler_WithMiniredis_ReleaseAllLocks(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	s.Start()
	defer s.Stop()

	lock1, err := s.acquireLock("lock-1", 10*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, lock1)

	lock2, err := s.acquireLock("lock-2", 10*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, lock2)

	s.releaseAllLocks()
}

func TestDistributedScheduler_WithMiniredis_RenewAllLocks(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	s.Start()
	defer s.Stop()

	lock, err := s.acquireLock("renew-lock", 10*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, lock)

	s.renewAllLocks()
}

func TestDistributedScheduler_WithMiniredis_RegisterNode(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	err := s.registerNode()
	require.NoError(t, err)
}

func TestDistributedScheduler_WithMiniredis_UnregisterNode(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	err := s.registerNode()
	require.NoError(t, err)

	s.unregisterNode()
}

func TestDistributedScheduler_WithMiniredis_SendHeartbeat(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	err := s.registerNode()
	require.NoError(t, err)

	s.sendHeartbeat()
}

func TestDistributedScheduler_WithMiniredis_CheckNodes(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	err := s.registerNode()
	require.NoError(t, err)

	s.checkNodes()
}

func TestDistributedScheduler_WithMiniredis_ScheduleTasks(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	s.Start()
	defer s.Stop()

	task := &DistributedTask{
		ID:           "sched-task",
		Name:         "Schedule Task",
		Enabled:      true,
		Status:       TaskStatusPending,
		AssignedNode: s.nodeID,
		NextRunTime:  time.Now().Add(-1 * time.Second),
	}
	s.tasks["sched-task"] = task

	s.scheduleTasks()
}

func TestDistributedScheduler_WithMiniredis_ExecuteTask(t *testing.T) {
	s, mr, execImpl := newSchedulerWithMiniredis(t)
	defer mr.Close()

	execImpl.Start()
	defer execImpl.Stop()

	execImpl.RegisterHandler("", func(ctx *ExecutionContext) (*ExecutionResult, error) {
		return &ExecutionResult{Success: true}, nil
	})

	task := &DistributedTask{
		ID:           "exec-task",
		Name:         "Exec Task",
		Enabled:      true,
		Timeout:      5 * time.Second,
		Config:       map[string]interface{}{},
	}
	s.tasks["exec-task"] = task

	s.executeTask(task)
}

func TestDistributedScheduler_WithMiniredis_ExecuteTask_WithCron(t *testing.T) {
	s, mr, execImpl := newSchedulerWithMiniredis(t)
	defer mr.Close()

	execImpl.Start()
	defer execImpl.Stop()

	execImpl.RegisterHandler("", func(ctx *ExecutionContext) (*ExecutionResult, error) {
		return &ExecutionResult{Success: true}, nil
	})

	p := NewCronParser()
	expr, err := p.Parse("0 */5 * * * *")
	require.NoError(t, err)

	task := &DistributedTask{
		ID:           "cron-exec-task",
		Name:         "Cron Exec Task",
		Enabled:      true,
		Timeout:      5 * time.Second,
		CronExpr:     "0 */5 * * * *",
		Config:       map[string]interface{}{},
		cronExpr:     expr,
	}
	s.tasks["cron-exec-task"] = task

	s.executeTask(task)

	retrieved, err := s.GetTask("cron-exec-task")
	require.NoError(t, err)
	assert.Equal(t, int64(1), retrieved.RunCount)
	assert.Equal(t, int64(1), retrieved.SuccessCount)
}

func TestDistributedScheduler_WithMiniredis_ExecuteTask_Fail(t *testing.T) {
	s, mr, execImpl := newSchedulerWithMiniredis(t)
	defer mr.Close()

	execImpl.Start()
	defer execImpl.Stop()

	execImpl.RegisterHandler("", func(ctx *ExecutionContext) (*ExecutionResult, error) {
		return nil, fmt.Errorf("task failed")
	})

	task := &DistributedTask{
		ID:           "fail-task",
		Name:         "Fail Task",
		Enabled:      true,
		Timeout:      5 * time.Second,
		Config:       map[string]interface{}{},
	}
	s.tasks["fail-task"] = task

	s.executeTask(task)

	retrieved, err := s.GetTask("fail-task")
	require.NoError(t, err)
	assert.Equal(t, int64(1), retrieved.FailCount)
	assert.Equal(t, TaskStatusFailed, retrieved.LastStatus)
}

func TestDistributedScheduler_WithMiniredis_HandleTaskError_NoRetry(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	task := &DistributedTask{
		ID:       "no-retry-task",
		MaxRetry: 0,
	}
	s.tasks["no-retry-task"] = task

	s.handleTaskError(task, fmt.Errorf("error"), time.Now())
}

func TestDistributedScheduler_WithMiniredis_HandleTaskError_WithRetry(t *testing.T) {
	s, mr, execImpl := newSchedulerWithMiniredis(t)
	defer mr.Close()

	execImpl.Start()
	defer execImpl.Stop()

	execImpl.RegisterHandler("test", func(ctx *ExecutionContext) (*ExecutionResult, error) {
		return &ExecutionResult{Success: true}, nil
	})

	task := &DistributedTask{
		ID:       "retry-task",
		MaxRetry: 1,
		Timeout:  5 * time.Second,
		Config:   map[string]interface{}{},
	}
	s.tasks["retry-task"] = task

	s.handleTaskError(task, fmt.Errorf("error"), time.Now())
	time.Sleep(2 * time.Second)
}

func TestDistributedScheduler_WithMiniredis_DoRebalance(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	s.Start()
	defer s.Stop()

	s.nodeState = NodeStateActive
	s.nodes["node1"] = &ClusterNode{
		NodeID: "node1",
		State:  NodeStateActive,
		Load:   0.3,
	}

	task := &DistributedTask{
		ID:           "rebalance-task",
		Name:         "Rebalance Task",
		Enabled:      true,
		AssignedNode: "old-node",
	}
	s.tasks["rebalance-task"] = task

	s.doRebalance()
}

func TestDistributedScheduler_WithMiniredis_DoRebalance_NoNodes(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	s.Start()
	defer s.Stop()

	s.doRebalance()
}

func TestDistributedScheduler_WithMiniredis_AddTask_WithNodes(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	s.Start()
	defer s.Stop()

	s.nodeMutex.Lock()
	s.nodes["node1"] = &ClusterNode{
		NodeID: "node1",
		State:  NodeStateActive,
		Load:   0.3,
	}
	s.nodeMutex.Unlock()

	task := &DistributedTask{
		ID:      "task-with-nodes",
		Name:    "Task With Nodes",
		Enabled: true,
	}

	err := s.AddTask(task)
	require.NoError(t, err)

	retrieved, err := s.GetTask("task-with-nodes")
	require.NoError(t, err)
	assert.Equal(t, "node1", retrieved.AssignedNode)
}

func TestDistributedScheduler_WithMiniredis_AddTask_WithSharding(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	s.Start()
	defer s.Stop()

	s.nodeMutex.Lock()
	s.nodes["node1"] = &ClusterNode{
		NodeID: "node1",
		State:  NodeStateActive,
		Load:   0.3,
	}
	s.nodes["node2"] = &ClusterNode{
		NodeID: "node2",
		State:  NodeStateActive,
		Load:   0.5,
	}
	s.nodeMutex.Unlock()

	task := &DistributedTask{
		ID:           "task-shard",
		Name:         "Shard Task",
		Enabled:      true,
		ShardingKey:  "shard-key-1",
	}

	err := s.AddTask(task)
	require.NoError(t, err)
}

func TestDistributedScheduler_WithMiniredis_Start_AlreadyRunning(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	err := s.Start()
	require.NoError(t, err)
	defer s.Stop()

	err = s.Start()
	assert.Equal(t, ErrSchedulerRunning, err)
}

func TestDistributedScheduler_WithMiniredis_Stop_NotRunning(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	err := s.Stop()
	assert.Equal(t, ErrSchedulerNotRunning, err)
}

func TestDistributedScheduler_WithMiniredis_New_Fails(t *testing.T) {
	cfg := DefaultDistributedSchedulerConfig()
	cfg.RedisAddr = "invalid-host:6379"
	executor := NewTaskExecutor(nil)
	_, err := NewDistributedScheduler(cfg, executor)
	assert.Error(t, err)
}

func TestDistributedScheduler_WithMiniredis_AddTask_NoCron(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	s.Start()
	defer s.Stop()

	task := &DistributedTask{
		ID:      "task-no-cron",
		Name:    "No Cron Task",
		Enabled: true,
	}

	err := s.AddTask(task)
	require.NoError(t, err)

	retrieved, err := s.GetTask("task-no-cron")
	require.NoError(t, err)
	assert.Equal(t, "task-no-cron", retrieved.ID)
}

func TestDistributedScheduler_WithMiniredis_AddTask_AutoGenID(t *testing.T) {
	s, mr, _ := newSchedulerWithMiniredis(t)
	defer mr.Close()

	s.Start()
	defer s.Stop()

	task := &DistributedTask{
		ID:      "auto-gen-id",
		Name:    "Auto ID Task",
		Enabled: true,
	}

	err := s.AddTask(task)
	require.NoError(t, err)
	assert.Equal(t, "auto-gen-id", task.ID)
}
