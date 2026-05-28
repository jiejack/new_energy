package query

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDefaultMonitorConfig(t *testing.T) {
	cfg := DefaultMonitorConfig()
	assert.True(t, cfg.Enabled)
	assert.Equal(t, 1*time.Second, cfg.SlowQueryThreshold)
	assert.Equal(t, 1000, cfg.MaxSlowQueries)
	assert.Equal(t, 10000, cfg.MaxQueryHistory)
	assert.True(t, cfg.LogSlowQueries)
	assert.True(t, cfg.EnableProfiling)
	assert.True(t, cfg.CollectQueryPlan)
	assert.Equal(t, 5*time.Second, cfg.AlertThreshold)
}

func TestNewQueryMonitor(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)
	require.NotNil(t, monitor)
}

func TestQueryMonitor_StartStop(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	cfg.StatsInterval = 100 * time.Millisecond
	cfg.ReportInterval = 100 * time.Millisecond
	monitor := NewQueryMonitor(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	monitor.Start(ctx)
	monitor.Stop()
}

func TestQueryMonitor_Start_AlreadyRunning(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	monitor.Start(ctx)
	monitor.Start(ctx)
	monitor.Stop()
}

func TestQueryMonitor_RecordQuery(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{
		QueryID:       "q1",
		Status:        QueryStatusCompleted,
		Data:          []map[string]interface{}{{"id": 1}},
		ExecutionTime: 50 * time.Millisecond,
	}

	monitor.RecordQuery(req, result, nil)

	stats := monitor.GetStats()
	assert.Equal(t, int64(1), stats.TotalQueries)
	assert.Equal(t, int64(1), stats.SuccessQueries)
	assert.Equal(t, int64(0), stats.FailedQueries)
}

func TestQueryMonitor_RecordQuery_WithError(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{
		QueryID:       "q1",
		Status:        QueryStatusFailed,
		ExecutionTime: 50 * time.Millisecond,
	}

	monitor.RecordQuery(req, result, ErrQueryTimeout)

	stats := monitor.GetStats()
	assert.Equal(t, int64(1), stats.TotalQueries)
	assert.Equal(t, int64(1), stats.FailedQueries)
}

func TestQueryMonitor_RecordQuery_SlowQuery(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	cfg.SlowQueryThreshold = 10 * time.Millisecond
	monitor := NewQueryMonitor(cfg)

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{
		QueryID:       "q1",
		Status:        QueryStatusCompleted,
		Data:          []map[string]interface{}{{"id": 1}},
		ExecutionTime: 100 * time.Millisecond,
	}

	monitor.RecordQuery(req, result, nil)

	stats := monitor.GetStats()
	assert.Equal(t, int64(1), stats.SlowQueries)

	slowQueries := monitor.GetSlowQueries(10)
	assert.Equal(t, 1, len(slowQueries))
}

func TestQueryMonitor_RecordQuery_Cached(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{
		QueryID:       "q1",
		Status:        QueryStatusCompleted,
		Data:          []map[string]interface{}{{"id": 1}},
		ExecutionTime: 5 * time.Millisecond,
		Cached:        true,
	}

	monitor.RecordQuery(req, result, nil)

	stats := monitor.GetStats()
	assert.Equal(t, int64(1), stats.CachedQueries)
}

func TestQueryMonitor_RecordQuery_Disabled(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.Enabled = false
	monitor := NewQueryMonitor(cfg)

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{QueryID: "q1", Status: QueryStatusCompleted}

	monitor.RecordQuery(req, result, nil)

	stats := monitor.GetStats()
	assert.Equal(t, int64(0), stats.TotalQueries)
}

func TestQueryMonitor_GetHistory(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{
		QueryID:       "q1",
		Status:        QueryStatusCompleted,
		Data:          []map[string]interface{}{{"id": 1}},
		ExecutionTime: 5 * time.Millisecond,
	}

	monitor.RecordQuery(req, result, nil)

	history := monitor.GetHistory(10)
	assert.Equal(t, 1, len(history))
}

func TestQueryMonitor_GenerateReport(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{
		QueryID:       "q1",
		Status:        QueryStatusCompleted,
		Data:          []map[string]interface{}{{"id": 1}},
		ExecutionTime: 5 * time.Millisecond,
	}

	monitor.RecordQuery(req, result, nil)

	report := monitor.GenerateReport()
	require.NotNil(t, report)
	assert.Equal(t, int64(1), report.TotalQueries)
	assert.GreaterOrEqual(t, report.HealthScore, 0.0)
	assert.LessOrEqual(t, report.HealthScore, 100.0)
}

func TestQueryMonitor_GenerateReport_WithFailures(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{
		QueryID:       "q1",
		Status:        QueryStatusFailed,
		ExecutionTime: 5 * time.Millisecond,
	}

	monitor.RecordQuery(req, result, ErrQueryTimeout)

	report := monitor.GenerateReport()
	require.NotNil(t, report)
	assert.Less(t, report.HealthScore, 100.0)
}

func TestQueryMonitor_MultipleQueries(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)

	for i := 0; i < 5; i++ {
		req := NewQueryBuilder().Table("devices").Build()
		result := &QueryResult{
			QueryID:       "q" + string(rune('0'+i)),
			Status:        QueryStatusCompleted,
			Data:          []map[string]interface{}{{"id": i}},
			ExecutionTime: time.Duration(i+1) * 10 * time.Millisecond,
		}
		monitor.RecordQuery(req, result, nil)
	}

	stats := monitor.GetStats()
	assert.Equal(t, int64(5), stats.TotalQueries)
	assert.Equal(t, int64(5), stats.SuccessQueries)
	assert.Greater(t, stats.QueriesPerSecond, 0.0)
}

func TestQueryMonitor_AlertThreshold(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	cfg.AlertThreshold = 50 * time.Millisecond
	monitor := NewQueryMonitor(cfg)

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{
		QueryID:       "q1",
		Status:        QueryStatusCompleted,
		Data:          []map[string]interface{}{{"id": 1}},
		ExecutionTime: 100 * time.Millisecond,
	}

	monitor.RecordQuery(req, result, nil)

	alerts := monitor.alertManager.GetAlerts(10)
	assert.Equal(t, 1, len(alerts))
	assert.Equal(t, "slow_query", alerts[0].Type)
	assert.Equal(t, "warning", alerts[0].Level)
}

func TestQueryMonitor_Middleware(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)
	mw := NewQueryMonitorMiddleware(monitor)
	require.NotNil(t, mw)

	ctx := context.Background()
	req := NewQueryBuilder().Table("devices").Build()

	ctx = mw.Before(ctx, req)
	assert.NotNil(t, ctx.Value("query_start_time"))

	result := &QueryResult{
		QueryID:       "q1",
		Status:        QueryStatusCompleted,
		ExecutionTime: 5 * time.Millisecond,
	}
	mw.After(ctx, req, result, nil)
}

func TestMetricsExporter(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{
		QueryID:       "q1",
		Status:        QueryStatusCompleted,
		Data:          []map[string]interface{}{{"id": 1}},
		ExecutionTime: 5 * time.Millisecond,
	}
	monitor.RecordQuery(req, result, nil)

	exporter := NewMetricsExporter(monitor)
	metrics := exporter.ExportPrometheus()
	assert.Contains(t, metrics, "query_total")
	assert.Contains(t, metrics, "query_success_total")
	assert.Contains(t, metrics, "query_failed_total")
	assert.Contains(t, metrics, "query_slow_total")
	assert.Contains(t, metrics, "query_duration_avg")
	assert.Contains(t, metrics, "query_qps")
}

func TestSlowQueryLog_WithFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := tmpDir + "/slow_queries.log"
	log := NewSlowQueryLog(5, filePath)
	require.NotNil(t, log)

	entry := &SlowQueryEntry{
		QueryID:       "q1",
		QueryType:     QueryTypeSelect,
		QueryText:     "SELECT * FROM devices",
		Database:      "testdb",
		Table:         "devices",
		ExecutionTime: 2 * time.Second,
		RowsReturned:  100,
		Timestamp:     time.Now(),
	}

	log.Add(entry)

	queries := log.Get(10)
	assert.Equal(t, 1, len(queries))
}

func TestSlowQueryLog_Overflow(t *testing.T) {
	log := NewSlowQueryLog(2, "")

	for i := 0; i < 5; i++ {
		log.Add(&SlowQueryEntry{
			QueryID:       "q" + string(rune('0'+i)),
			ExecutionTime: time.Second,
		})
	}

	queries := log.Get(10)
	assert.Equal(t, 2, len(queries))
}

func TestSlowQueryLog_GetNegative(t *testing.T) {
	log := NewSlowQueryLog(5, "")
	log.Add(&SlowQueryEntry{QueryID: "q1", ExecutionTime: time.Second})

	queries := log.Get(-1)
	assert.Equal(t, 1, len(queries))
}

func TestQueryHistory_Overflow(t *testing.T) {
	history := NewQueryHistory(3)

	for i := 0; i < 5; i++ {
		history.Add(&QueryHistoryEntry{
			QueryID:       "q" + string(rune('0'+i)),
			ExecutionTime: time.Duration(i+1) * 10 * time.Millisecond,
		})
	}

	entries := history.Get(10)
	assert.Equal(t, 3, len(entries))
}

func TestQueryHistory_GetExecutionTimes(t *testing.T) {
	history := NewQueryHistory(10)

	for i := 0; i < 3; i++ {
		history.Add(&QueryHistoryEntry{
			QueryID:       "q" + string(rune('0'+i)),
			ExecutionTime: time.Duration(i+1) * 10 * time.Millisecond,
		})
	}

	times := history.GetExecutionTimes()
	assert.Equal(t, 3, len(times))
}

func TestQueryAnalyzer(t *testing.T) {
	analyzer := NewQueryAnalyzer()
	require.NotNil(t, analyzer)

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{
		QueryID:       "q1",
		ExecutionTime: 600 * time.Millisecond,
	}

	analyzer.Analyze(req, result)

	patterns := analyzer.GetAllPatterns()
	assert.Greater(t, len(patterns), 0)
}

func TestQueryAnalyzer_GetPattern(t *testing.T) {
	analyzer := NewQueryAnalyzer()

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{QueryID: "q1", ExecutionTime: 100 * time.Millisecond}

	analyzer.Analyze(req, result)

	patterns := analyzer.GetAllPatterns()
	assert.Greater(t, len(patterns), 0)

	for key, p := range patterns {
		assert.Equal(t, int64(1), p.Count)
		_, ok := analyzer.GetPattern(key)
		assert.True(t, ok)
		break
	}

	_, ok := analyzer.GetPattern("nonexistent")
	assert.False(t, ok)
}

func TestPerformanceReporter(t *testing.T) {
	reporter := NewPerformanceReporter(1 * time.Second)
	require.NotNil(t, reporter)

	report := &PerformanceReport{
		GeneratedAt:  time.Now(),
		TotalQueries: 100,
	}

	reporter.AddReport(report)

	latest := reporter.GetLatestReport()
	require.NotNil(t, latest)
	assert.Equal(t, int64(100), latest.TotalQueries)
}

func TestPerformanceReporter_Empty(t *testing.T) {
	reporter := NewPerformanceReporter(1 * time.Second)
	latest := reporter.GetLatestReport()
	assert.Nil(t, latest)
}

func TestPerformanceReporter_Subscribe(t *testing.T) {
	reporter := NewPerformanceReporter(1 * time.Second)
	ch := reporter.Subscribe()
	require.NotNil(t, ch)

	report := &PerformanceReport{GeneratedAt: time.Now()}
	reporter.Notify(report)

	select {
	case r := <-ch:
		assert.NotNil(t, r)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for report")
	}
}

func TestAlertManager(t *testing.T) {
	am := NewAlertManager(1 * time.Second)
	require.NotNil(t, am)

	ch := am.Subscribe()
	require.NotNil(t, ch)
}

func TestAlertManager_GetAlerts(t *testing.T) {
	am := NewAlertManager(1 * time.Second)

	alerts := am.GetAlerts(10)
	assert.Equal(t, 0, len(alerts))
}

func TestQueryMonitor_RecordQuery_NilResult(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)

	req := NewQueryBuilder().Table("devices").Build()
	monitor.RecordQuery(req, nil, ErrQueryTimeout)

	stats := monitor.GetStats()
	assert.Equal(t, int64(1), stats.TotalQueries)
	assert.Equal(t, int64(1), stats.FailedQueries)
}

func TestQueryMonitor_WithTableStats(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)

	req1 := NewQueryBuilder().Table("devices").Build()
	result1 := &QueryResult{
		QueryID:       "q1",
		Status:        QueryStatusCompleted,
		Data:          []map[string]interface{}{{"id": 1}},
		ExecutionTime: 5 * time.Millisecond,
	}
	monitor.RecordQuery(req1, result1, nil)

	req2 := NewQueryBuilder().Table("sensors").Build()
	result2 := &QueryResult{
		QueryID:       "q2",
		Status:        QueryStatusCompleted,
		Data:          []map[string]interface{}{{"id": 2}},
		ExecutionTime: 10 * time.Millisecond,
	}
	monitor.RecordQuery(req2, result2, nil)

	report := monitor.GenerateReport()
	require.NotNil(t, report)
	assert.Greater(t, len(report.TopTables), 0)
}

func TestQueryMonitor_WithErrors(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{
		QueryID:       "q1",
		Status:        QueryStatusFailed,
		ExecutionTime: 5 * time.Millisecond,
	}
	monitor.RecordQuery(req, result, ErrQueryTimeout)

	report := monitor.GenerateReport()
	require.NotNil(t, report)
	assert.Greater(t, len(report.TopErrors), 0)
}

func TestQueryMonitor_HealthScore_LowCacheRate(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)

	for i := 0; i < 10; i++ {
		req := NewQueryBuilder().Table("devices").Build()
		result := &QueryResult{
			QueryID:       "q" + string(rune('0'+i)),
			Status:        QueryStatusCompleted,
			Data:          []map[string]interface{}{{"id": i}},
			ExecutionTime: 5 * time.Millisecond,
			Cached:        false,
		}
		monitor.RecordQuery(req, result, nil)
	}

	report := monitor.GenerateReport()
	assert.Greater(t, len(report.Recommendations), 0)
}

func TestQueryMonitor_SlowQueryRecommendations(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	cfg.SlowQueryThreshold = 1 * time.Millisecond
	monitor := NewQueryMonitor(cfg)

	for i := 0; i < 20; i++ {
		req := NewQueryBuilder().Table("devices").Build()
		result := &QueryResult{
			QueryID:       "q" + string(rune('0'+i)),
			Status:        QueryStatusCompleted,
			Data:          []map[string]interface{}{{"id": i}},
			ExecutionTime: 10 * time.Millisecond,
		}
		monitor.RecordQuery(req, result, nil)
	}

	report := monitor.GenerateReport()
	assert.Greater(t, len(report.Recommendations), 0)
}

func TestQueryMonitor_HighAvgLatencyRecommendations(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{
		QueryID:       "q1",
		Status:        QueryStatusCompleted,
		Data:          []map[string]interface{}{{"id": 1}},
		ExecutionTime: 600 * time.Millisecond,
	}
	monitor.RecordQuery(req, result, nil)

	report := monitor.GenerateReport()
	assert.NotNil(t, report)
}

func TestQueryMonitor_HighErrorRateRecommendations(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)

	for i := 0; i < 10; i++ {
		req := NewQueryBuilder().Table("devices").Build()
		result := &QueryResult{
			QueryID:       "q" + string(rune('0'+i)),
			Status:        QueryStatusFailed,
			ExecutionTime: 5 * time.Millisecond,
		}
		monitor.RecordQuery(req, result, ErrQueryTimeout)
	}

	report := monitor.GenerateReport()
	assert.NotNil(t, report)
}

func TestQueryExecutor_NewWithDefaults(t *testing.T) {
	executor := NewQueryExecutor(nil, ExecutorConfig{})
	require.NotNil(t, executor)
	assert.Equal(t, 10, executor.config.MaxParallelQueries)
	assert.Equal(t, 100000, executor.config.MaxResultRows)
	assert.Equal(t, 30*time.Second, executor.config.DefaultTimeout)
	assert.Equal(t, 1*time.Second, executor.config.SlowQueryThreshold)
	assert.Equal(t, 1000, executor.config.StreamBatchSize)
}

func TestQueryExecutor_GetStats(t *testing.T) {
	executor := NewQueryExecutor(nil, ExecutorConfig{})
	stats := executor.GetStats()
	require.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.TotalQueries)
}

func TestQueryExecutor_CompareValues(t *testing.T) {
	executor := NewQueryExecutor(nil, ExecutorConfig{})

	assert.Equal(t, -1, executor.compareValues(1, 2))
	assert.Equal(t, 0, executor.compareValues(5, 5))
	assert.Equal(t, 1, executor.compareValues(10, 5))

	assert.Equal(t, -1, executor.compareValues(int64(1), int64(2)))
	assert.Equal(t, 0, executor.compareValues(int64(5), int64(5)))

	assert.Equal(t, -1, executor.compareValues(1.0, 2.0))
	assert.Equal(t, 0, executor.compareValues(5.0, 5.0))

	assert.Equal(t, -1, executor.compareValues("a", "b"))
	assert.Equal(t, 0, executor.compareValues("x", "x"))

	t1 := time.Now()
	t2 := t1.Add(time.Hour)
	assert.Equal(t, -1, executor.compareValues(t1, t2))
	assert.Equal(t, 1, executor.compareValues(t2, t1))

	assert.Equal(t, 0, executor.compareValues(1, "string"))
}

func TestQueryExecutor_SortAndPaginate(t *testing.T) {
	executor := NewQueryExecutor(nil, ExecutorConfig{})

	data := []map[string]interface{}{
		{"name": "c", "value": 3},
		{"name": "a", "value": 1},
		{"name": "b", "value": 2},
	}

	req := &QueryRequest{
		OrderBy: []OrderByField{{Field: "value", Desc: false}},
		Limit:   2,
		Offset:  0,
	}

	result := executor.sortAndPaginate(data, req)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, 1, result[0]["value"])
	assert.Equal(t, 2, result[1]["value"])
}

func TestQueryExecutor_SortAndPaginate_Desc(t *testing.T) {
	executor := NewQueryExecutor(nil, ExecutorConfig{})

	data := []map[string]interface{}{
		{"name": "a", "value": 1},
		{"name": "b", "value": 2},
		{"name": "c", "value": 3},
	}

	req := &QueryRequest{
		OrderBy: []OrderByField{{Field: "value", Desc: true}},
	}

	result := executor.sortAndPaginate(data, req)
	assert.Equal(t, 3, len(result))
	assert.Equal(t, 3, result[0]["value"])
}

func TestQueryExecutor_SortAndPaginate_OffsetBeyond(t *testing.T) {
	executor := NewQueryExecutor(nil, ExecutorConfig{})

	data := []map[string]interface{}{
		{"name": "a", "value": 1},
	}

	req := &QueryRequest{
		Offset: 100,
	}

	result := executor.sortAndPaginate(data, req)
	assert.Equal(t, 0, len(result))
}

func TestQueryExecutor_ExtractFields(t *testing.T) {
	executor := NewQueryExecutor(nil, ExecutorConfig{})

	data := []map[string]interface{}{
		{"id": 1, "name": "test"},
	}

	fields := executor.extractFields(data)
	assert.Equal(t, 2, len(fields))
}

func TestQueryExecutor_ExtractFields_Empty(t *testing.T) {
	executor := NewQueryExecutor(nil, ExecutorConfig{})

	fields := executor.extractFields([]map[string]interface{}{})
	assert.Equal(t, 0, len(fields))
}

func TestQueryExecutor_CompareRows_MissingFields(t *testing.T) {
	executor := NewQueryExecutor(nil, ExecutorConfig{})

	a := map[string]interface{}{"value": 1}
	b := map[string]interface{}{}

	result := executor.compareRows(a, b, []OrderByField{{Field: "value", Desc: false}})
	assert.Equal(t, -1, result)

	result = executor.compareRows(b, a, []OrderByField{{Field: "value", Desc: false}})
	assert.Equal(t, 1, result)
}

func TestQueryExecutor_CompareRows_BothMissing(t *testing.T) {
	executor := NewQueryExecutor(nil, ExecutorConfig{})

	a := map[string]interface{}{}
	b := map[string]interface{}{}

	result := executor.compareRows(a, b, []OrderByField{{Field: "value", Desc: false}})
	assert.Equal(t, 0, result)
}

func TestQueryExecutor_BuildAggregateField(t *testing.T) {
	executor := NewQueryExecutor(nil, ExecutorConfig{})

	agg := AggregateField{Field: "power", Function: "SUM", Alias: "total"}
	result := executor.buildAggregateField(agg)
	assert.Equal(t, "SUM(power) AS total", result)

	aggDistinct := AggregateField{Field: "id", Function: "COUNT", Alias: "", Distinct: true}
	result = executor.buildAggregateField(aggDistinct)
	assert.Equal(t, "COUNT(DISTINCT id)", result)
}

func TestQueryExecutor_DecomposeQuery_NoTimeRange(t *testing.T) {
	executor := NewQueryExecutor(nil, ExecutorConfig{})

	req := NewQueryBuilder().Table("devices").Build()
	plan := &QueryPlan{Parallel: true}

	subQueries, err := executor.decomposeQuery(req, plan)
	require.NoError(t, err)
	assert.Equal(t, 1, len(subQueries))
}

func TestQueryExecutor_DecomposeQuery_InvalidInterval(t *testing.T) {
	executor := NewQueryExecutor(nil, ExecutorConfig{})

	req := NewQueryBuilder().
		Table("devices").
		TimeRange("ts", time.Now().Add(-time.Hour), time.Now()).
		Build()
	req.TimeRange.Interval = "invalid"

	plan := &QueryPlan{Parallel: true}

	_, err := executor.decomposeQuery(req, plan)
	assert.Error(t, err)
}

func TestQueryExecutor_DecomposeQuery_ValidInterval(t *testing.T) {
	executor := NewQueryExecutor(nil, ExecutorConfig{})

	now := time.Now()
	req := NewQueryBuilder().
		Table("devices").
		TimeRange("ts", now.Add(-2*time.Hour), now).
		Build()
	req.TimeRange.Interval = "1h"

	plan := &QueryPlan{Parallel: true}

	subQueries, err := executor.decomposeQuery(req, plan)
	require.NoError(t, err)
	assert.Equal(t, 2, len(subQueries))
}

func TestQueryPlanOptimizer_ComplexQuery(t *testing.T) {
	optimizer := NewQueryPlanOptimizer()

	req := NewQueryBuilder().
		Table("devices").
		Where("status", "=", "active").
		Join("INNER", "stations", "s", JoinCondition{
			LeftField:  "devices.station_id",
			Operator:   "=",
			RightField: "s.id",
		}).
		Aggregate("power", "SUM", "total_power").
		GroupBy("station_id").
		OrderBy("power", true).
		Build()

	plan, err := optimizer.Optimize(req)
	require.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Greater(t, len(plan.Steps), 0)
	assert.Greater(t, plan.EstimatedCost, 0.0)
}

func TestQueryPlanOptimizer_ParallelExecution(t *testing.T) {
	optimizer := NewQueryPlanOptimizer()

	req := NewQueryBuilder().Table("devices").Build()
	req.Conditions = make([]QueryCondition, 200)

	plan, err := optimizer.Optimize(req)
	require.NoError(t, err)
	assert.NotNil(t, plan)
}

func TestParallelExecutor_WithErrors(t *testing.T) {
	pe := NewParallelExecutor(4)

	tasks := []func() error{
		func() error { return ErrQueryTimeout },
		func() error { return nil },
	}

	pe.Execute(context.Background(), tasks)
}

func TestParallelExecutor_Cancelled(t *testing.T) {
	pe := NewParallelExecutor(4)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tasks := []func() error{
		func() error { return nil },
	}

	err := pe.Execute(ctx, tasks)
	assert.Error(t, err)
}

func TestQueryCache_SetWithCustomTTL(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	ctx := context.Background()

	req := NewQueryBuilder().Table("devices").Build()
	req.Options = map[string]interface{}{
		"cache_ttl": 10 * time.Minute,
	}
	result := &QueryResult{QueryID: "test", Status: QueryStatusCompleted}

	err := cache.Set(ctx, req, result)
	require.NoError(t, err)
}

func TestQueryCache_DeleteNonExistent(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	ctx := context.Background()

	req := NewQueryBuilder().Table("nonexistent").Build()
	err := cache.Delete(ctx, req)
	require.NoError(t, err)
}

func TestQueryCache_DeleteByTags_NonExistent(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	ctx := context.Background()

	err := cache.DeleteByTags(ctx, []string{"nonexistent"})
	require.NoError(t, err)
}

func TestQueryCache_EvictionBySize(t *testing.T) {
	cfg := DefaultCacheConfig()
	cfg.MaxEntries = 100
	cfg.MaxSize = 100
	cache := NewQueryCache(nil, cfg)
	ctx := context.Background()

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{
		QueryID: "test",
		Status:  QueryStatusCompleted,
		Data:    []map[string]interface{}{{"id": 1, "name": "a very long name to increase size"}},
	}

	cache.Set(ctx, req, result)
}

func TestQueryCache_CleanupExpired(t *testing.T) {
	cfg := DefaultCacheConfig()
	cfg.DefaultTTL = 1 * time.Nanosecond
	cache := NewQueryCache(nil, cfg)
	ctx := context.Background()

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{QueryID: "test", Status: QueryStatusCompleted}
	cache.Set(ctx, req, result)

	time.Sleep(10 * time.Millisecond)
	cache.cleanupExpired()

	stats := cache.GetStats()
	assert.Equal(t, int64(0), stats.EntryCount)
}

func TestCacheInvalidator_TimeBased(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	invalidator := NewCacheInvalidator(cache)
	require.NotNil(t, invalidator)

	invalidator.AddRule(InvalidationRule{
		ID:        "time_rule",
		Name:      "time rule",
		Trigger:   "time",
		Interval:  1 * time.Hour,
		Tags:      []string{"table:devices"},
		Enabled:   true,
	})

	invalidator.RemoveRule("time_rule")
}

func TestCacheInvalidator_DisabledRule(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	invalidator := NewCacheInvalidator(cache)

	invalidator.AddRule(InvalidationRule{
		ID:      "disabled_rule",
		Name:    "disabled",
		Trigger: "event",
		Tables:  []string{"devices"},
		Enabled: false,
	})

	invalidator.Notify(InvalidationEvent{
		Type:  "update",
		Table: "devices",
	})
}

func TestCacheInvalidator_WildcardTable(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	invalidator := NewCacheInvalidator(cache)

	ctx := context.Background()
	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{QueryID: "test", Status: QueryStatusCompleted}
	cache.Set(ctx, req, result)

	invalidator.AddRule(InvalidationRule{
		ID:      "wildcard_rule",
		Name:    "wildcard",
		Trigger: "event",
		Tables:  []string{"*"},
		Enabled: true,
		Tags:    []string{"table:devices"},
	})

	invalidator.Notify(InvalidationEvent{
		Type:  "update",
		Table: "any_table",
	})
}

func TestQueryCache_WithRedis(t *testing.T) {
	cfg := DefaultCacheConfig()
	cache := NewQueryCache(nil, cfg)
	ctx := context.Background()

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{QueryID: "test", Status: QueryStatusCompleted}
	cache.Set(ctx, req, result)

	got, status, err := cache.Get(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, CacheStatusHit, status)
	require.NotNil(t, got)
}

func TestCacheWarmer_StartAlreadyRunning(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	warmer := NewCacheWarmer(cache, nil)

	warmer.Start(context.Background())
	warmer.Start(context.Background())
	warmer.Stop()
}

func TestCacheWarmer_StartWithDisabledTask(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	warmer := NewCacheWarmer(cache, nil)

	warmer.AddTask(&WarmupTask{
		Name:    "disabled",
		Query:   NewQueryBuilder().Table("devices").Build(),
		Enabled: false,
	})

	warmer.Start(context.Background())
	warmer.Stop()
}

func TestQueryMonitor_Percentiles(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)

	for i := 0; i < 20; i++ {
		req := NewQueryBuilder().Table("devices").Build()
		result := &QueryResult{
			QueryID:       "q" + string(rune('0'+i)),
			Status:        QueryStatusCompleted,
			Data:          []map[string]interface{}{{"id": i}},
			ExecutionTime: time.Duration(i+1) * 10 * time.Millisecond,
		}
		monitor.RecordQuery(req, result, nil)
	}

	stats := monitor.GetStats()
	assert.Greater(t, stats.P95ExecutionTime, time.Duration(0))
	assert.Greater(t, stats.P99ExecutionTime, time.Duration(0))
}

func TestRateLimitStatus_Constants(t *testing.T) {
	assert.Equal(t, RateLimitStatus("allowed"), RateLimitStatusAllowed)
	assert.Equal(t, RateLimitStatus("limited"), RateLimitStatusLimited)
	assert.Equal(t, RateLimitStatus("waiting"), RateLimitStatusWaiting)
}

func TestQueryMonitor_RecordQuery_NilResultWithNilError(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)

	req := NewQueryBuilder().Table("devices").Build()
	monitor.RecordQuery(req, nil, nil)

	stats := monitor.GetStats()
	assert.Equal(t, int64(1), stats.TotalQueries)
	assert.Equal(t, int64(1), stats.SuccessQueries)
}

func TestQueryMonitor_TableSlowQueryRecommendations(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	cfg.SlowQueryThreshold = 1 * time.Millisecond
	monitor := NewQueryMonitor(cfg)

	for i := 0; i < 110; i++ {
		req := NewQueryBuilder().Table("devices").Build()
		result := &QueryResult{
			QueryID:       "q",
			Status:        QueryStatusCompleted,
			Data:          []map[string]interface{}{{"id": i}},
			ExecutionTime: 10 * time.Millisecond,
		}
		monitor.RecordQuery(req, result, nil)
	}

	report := monitor.GenerateReport()
	assert.NotNil(t, report)
}

func TestQueryExecutor_UpdateStats(t *testing.T) {
	executor := NewQueryExecutor(nil, ExecutorConfig{})

	result := &QueryResult{
		QueryID: "q1",
		Data:    []map[string]interface{}{{"id": 1}},
	}
	executor.updateStats(result, 100*time.Millisecond, nil)

	stats := executor.GetStats()
	assert.Equal(t, int64(1), stats.TotalQueries)
	assert.Equal(t, int64(1), stats.SuccessQueries)
	assert.Equal(t, int64(1), stats.TotalRows)
}

func TestQueryExecutor_UpdateStats_WithError(t *testing.T) {
	executor := NewQueryExecutor(nil, ExecutorConfig{})

	result := &QueryResult{QueryID: "q1"}
	executor.updateStats(result, 100*time.Millisecond, ErrQueryTimeout)

	stats := executor.GetStats()
	assert.Equal(t, int64(1), stats.TotalQueries)
	assert.Equal(t, int64(1), stats.FailedQueries)
}

func TestQueryExecutor_CompareRows_DescOrder(t *testing.T) {
	executor := NewQueryExecutor(nil, ExecutorConfig{})

	a := map[string]interface{}{"value": 1}
	b := map[string]interface{}{"value": 2}

	result := executor.compareRows(a, b, []OrderByField{{Field: "value", Desc: true}})
	assert.Equal(t, 1, result)
}

func TestQueryMonitor_WithPriorityStats(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)

	req := NewQueryBuilder().Table("devices").Priority(PriorityHigh).Build()
	result := &QueryResult{
		QueryID:       "q1",
		Status:        QueryStatusCompleted,
		Data:          []map[string]interface{}{{"id": 1}},
		ExecutionTime: 5 * time.Millisecond,
	}
	monitor.RecordQuery(req, result, nil)

	stats := monitor.GetStats()
	assert.NotNil(t, stats.ByPriority)
	_, ok := stats.ByPriority[PriorityHigh]
	assert.True(t, ok)
}

func TestQueryMonitor_RecordQuery_LongError(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{QueryID: "q1", Status: QueryStatusFailed}

	longErr := &QueryError{
		Code:    "VERY_LONG_ERROR_CODE_THAT_EXCEEDS_100_CHARS_WHICH_SHOULD_BE_TRUNCATED_IN_THE_STATS_STORAGE",
		Message: "This is a very long error message that should be truncated when stored in the statistics",
	}
	monitor.RecordQuery(req, result, longErr)

	stats := monitor.GetStats()
	assert.Equal(t, int64(1), stats.FailedQueries)
}

func TestQueryMonitor_RecordQuery_MinExecutionTime(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)

	req1 := NewQueryBuilder().Table("devices").Build()
	result1 := &QueryResult{
		QueryID:       "q1",
		Status:        QueryStatusCompleted,
		Data:          []map[string]interface{}{{"id": 1}},
		ExecutionTime: 100 * time.Millisecond,
	}
	monitor.RecordQuery(req1, result1, nil)

	req2 := NewQueryBuilder().Table("devices").Build()
	result2 := &QueryResult{
		QueryID:       "q2",
		Status:        QueryStatusCompleted,
		Data:          []map[string]interface{}{{"id": 2}},
		ExecutionTime: 10 * time.Millisecond,
	}
	monitor.RecordQuery(req2, result2, nil)

	stats := monitor.GetStats()
	assert.Less(t, stats.MinExecutionTime, 100*time.Millisecond)
}

func TestQueryMonitor_RecordQuery_NilResultHistory(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)

	req := NewQueryBuilder().Table("devices").Build()
	monitor.RecordQuery(req, nil, nil)

	history := monitor.GetHistory(10)
	assert.Equal(t, 0, len(history))
}

func TestQueryMonitor_SlowQueryLog_GetLargeLimit(t *testing.T) {
	log := NewSlowQueryLog(5, "")
	log.Add(&SlowQueryEntry{QueryID: "q1", ExecutionTime: time.Second})

	queries := log.Get(100)
	assert.Equal(t, 1, len(queries))
}

func TestQueryHistory_GetNegativeLimit(t *testing.T) {
	history := NewQueryHistory(10)
	history.Add(&QueryHistoryEntry{QueryID: "q1", ExecutionTime: time.Millisecond})

	entries := history.Get(-1)
	assert.Equal(t, 1, len(entries))
}

func TestAlertManager_GetAlerts_Empty(t *testing.T) {
	am := NewAlertManager(1 * time.Second)
	alerts := am.GetAlerts(0)
	assert.Equal(t, 0, len(alerts))
}

func TestPerformanceReporter_MultipleReports(t *testing.T) {
	reporter := NewPerformanceReporter(1 * time.Second)

	for i := 0; i < 15; i++ {
		reporter.AddReport(&PerformanceReport{
			GeneratedAt:  time.Now(),
			TotalQueries: int64(i),
		})
	}

	latest := reporter.GetLatestReport()
	require.NotNil(t, latest)
}

func TestCacheKeyBuilder_Empty(t *testing.T) {
	b := NewCacheKeyBuilder()
	key := b.Build()
	assert.NotEmpty(t, key)
}

func TestMultiLevelCache_Miss(t *testing.T) {
	cfg := MultiLevelCacheConfig{
		Levels: []CacheLevelConfig{
			{Name: "L1", Priority: 1, Config: DefaultCacheConfig()},
		},
	}
	mlc := NewMultiLevelCache(nil, cfg)

	ctx := context.Background()
	req := NewQueryBuilder().Table("nonexistent").Build()
	got, status, err := mlc.Get(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, CacheStatusMiss, status)
	assert.Nil(t, got)
}

func TestRateLimitRuleManager_DisabledRule(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	limiter := NewQueryRateLimiter(cfg)
	mgr := NewRateLimitRuleManager(limiter)

	mgr.AddRule(&RateLimitRule{
		ID:         "disabled",
		Name:       "Disabled Rule",
		KeyPattern: "*",
		Enabled:    false,
	})

	matched := mgr.MatchRule("any_key")
	assert.Nil(t, matched)
}

func TestQueryRateLimiter_AllowAndWait(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	cfg.DefaultBurst = 100
	cfg.MaxWaitTime = 100 * time.Millisecond
	limiter := NewQueryRateLimiter(cfg)

	ctx := context.Background()
	result, err := limiter.AllowAndWait(ctx, "key1", PriorityNormal)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

func TestQueryRateLimiter_AllowAndWait_NoMaxWait(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	cfg.DefaultBurst = 100
	cfg.MaxWaitTime = 0
	limiter := NewQueryRateLimiter(cfg)

	ctx := context.Background()
	result, err := limiter.AllowAndWait(ctx, "key1", PriorityNormal)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

func TestDistributedRateLimiter_FirstRequest(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	cfg.DefaultBurst = 100
	limiter := NewQueryRateLimiter(cfg)

	mockRedis := &mockRedisRateLimitClient{count: 0}
	distLimiter := NewDistributedRateLimiter(limiter, mockRedis, DistributedRateLimitConfig{
		GlobalRate: 100,
		KeyPrefix:  "ratelimit:",
	})

	ctx := context.Background()
	result, err := distLimiter.Allow(ctx, "key1", PriorityNormal)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

func TestQueryMonitor_buildQueryText(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)

	req := &QueryRequest{
		Fields: []string{"id", "name"},
		Table:  "devices",
		Conditions: []QueryCondition{
			{Field: "status", Operator: "=", Value: "active"},
		},
		GroupBy: []string{"status"},
		OrderBy: []OrderByField{{Field: "id", Desc: true}},
		Limit:   10,
	}

	text := monitor.buildQueryText(req)
	assert.Contains(t, text, "SELECT")
	assert.Contains(t, text, "devices")
	assert.Contains(t, text, "WHERE")
	assert.Contains(t, text, "GROUP BY")
	assert.Contains(t, text, "ORDER BY")
	assert.Contains(t, text, "LIMIT")
}

func TestQueryExecutor_Execute_WithSQLite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	type Device struct {
		ID     uint   `gorm:"primarykey"`
		Name   string `gorm:"column:name"`
		Status string `gorm:"column:status"`
		Power  float64 `gorm:"column:power"`
	}
	require.NoError(t, db.AutoMigrate(&Device{}))

	db.Create(&Device{Name: "dev1", Status: "active", Power: 100.5})
	db.Create(&Device{Name: "dev2", Status: "inactive", Power: 50.0})
	db.Create(&Device{Name: "dev3", Status: "active", Power: 200.0})

	executor := NewQueryExecutor(db, ExecutorConfig{})
	ctx := context.Background()

	req := NewQueryBuilder().Table("devices").Build()
	result, err := executor.Execute(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, QueryStatusCompleted, result.Status)
	assert.Equal(t, 3, len(result.Data))

	req2 := NewQueryBuilder().
		Table("devices").
		Select("name", "power").
		Where("status", "=", "active").
		OrderBy("power", true).
		Build()
	result2, err := executor.Execute(ctx, req2)
	require.NoError(t, err)
	assert.Equal(t, QueryStatusCompleted, result2.Status)
	assert.Equal(t, 2, len(result2.Data))
	assert.Equal(t, 200.0, result2.Data[0]["power"])

	req3 := NewQueryBuilder().
		Table("devices").
		Where("status", "=", "active").
		Build()
	result3, err := executor.Execute(ctx, req3)
	require.NoError(t, err)
	assert.Equal(t, QueryStatusCompleted, result3.Status)
	assert.Equal(t, 2, len(result3.Data))

	req4 := NewQueryBuilder().
		Table("devices").
		Where("status", "=", "nonexistent").
		Build()
	result4, err := executor.Execute(ctx, req4)
	require.NoError(t, err)
	assert.Equal(t, 0, len(result4.Data))
}

func TestQueryExecutor_Execute_WithQueryPlan_SQLite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	type Device struct {
		ID     uint   `gorm:"primarykey"`
		Name   string `gorm:"column:name"`
	}
	require.NoError(t, db.AutoMigrate(&Device{}))
	db.Create(&Device{Name: "dev1"})

	executor := NewQueryExecutor(db, ExecutorConfig{
		EnableQueryPlan: true,
	})
	ctx := context.Background()

	req := NewQueryBuilder().Table("devices").Build()
	result, err := executor.Execute(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, QueryStatusCompleted, result.Status)
}

func TestQueryExecutor_Execute_ParallelEnabled_SQLite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	type Device struct {
		ID     uint   `gorm:"primarykey"`
		Name   string `gorm:"column:name"`
	}
	require.NoError(t, db.AutoMigrate(&Device{}))
	db.Create(&Device{Name: "dev1"})

	executor := NewQueryExecutor(db, ExecutorConfig{
		EnableParallel:  true,
		EnableQueryPlan: true,
	})
	ctx := context.Background()

	req := NewQueryBuilder().Table("devices").Build()
	result, err := executor.Execute(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, QueryStatusCompleted, result.Status)
}

func TestQueryExecutor_StreamExecute_SQLite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	type Device struct {
		ID   uint   `gorm:"primarykey"`
		Name string `gorm:"column:name"`
	}
	require.NoError(t, db.AutoMigrate(&Device{}))
	db.Create(&Device{Name: "dev1"})
	db.Create(&Device{Name: "dev2"})

	executor := NewQueryExecutor(db, ExecutorConfig{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := NewQueryBuilder().Table("devices").Build()
	stream, err := executor.StreamExecute(ctx, req)
	require.NoError(t, err)

	count := 0
	for row := range stream {
		if row.Error != nil {
			t.Fatalf("stream error: %v", row.Error)
		}
		count++
	}
	assert.Equal(t, 2, count)
}

func TestQueryExecutor_Execute_LimitOffset_SQLite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	type Device struct {
		ID   uint   `gorm:"primarykey"`
		Name string `gorm:"column:name"`
	}
	require.NoError(t, db.AutoMigrate(&Device{}))
	for i := 0; i < 10; i++ {
		db.Create(&Device{Name: "dev" + string(rune('0'+i))})
	}

	executor := NewQueryExecutor(db, ExecutorConfig{})
	ctx := context.Background()

	req := NewQueryBuilder().
		Table("devices").
		Limit(3).
		Offset(2).
		Build()
	result, err := executor.Execute(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, 3, len(result.Data))
}

func TestAdaptiveRateLimiter_Record(t *testing.T) {
	limiter := NewAdaptiveRateLimiter(100)

	for i := 0; i < 20; i++ {
		limiter.RecordSuccess()
	}
	assert.Equal(t, 100, limiter.GetCurrentRate())

	limiter2 := NewAdaptiveRateLimiter(100)
	for i := 0; i < 20; i++ {
		limiter2.RecordFailure()
	}
	assert.Equal(t, 100, limiter2.GetCurrentRate())
}

func TestRateLimitMiddleware_Execute(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	cfg.DefaultRate = 1000
	cfg.DefaultBurst = 100
	limiter := NewQueryRateLimiter(cfg)
	mw := NewRateLimitMiddleware(limiter)

	ctx := context.Background()
	req := NewQueryBuilder().Table("devices").Build()
	result, err := mw.Process(ctx, req)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

func TestRateLimitGroup_Basic(t *testing.T) {
	group := NewRateLimitGroup("test", DefaultRateLimiterConfig())

	ctx := context.Background()
	result, err := group.Allow(ctx, "read", "key1", PriorityNormal)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

func TestDistributedRateLimiter_WithRedis(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	cfg.DefaultBurst = 100
	limiter := NewQueryRateLimiter(cfg)

	mockRedis := &mockRedisRateLimitClient{count: 50}
	distLimiter := NewDistributedRateLimiter(limiter, mockRedis, DistributedRateLimitConfig{
		GlobalRate: 100,
		KeyPrefix:  "ratelimit:",
	})

	ctx := context.Background()
	result, err := distLimiter.Allow(ctx, "key1", PriorityNormal)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

func TestQueryMonitor_StatsCollector(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	cfg.StatsInterval = 50 * time.Millisecond
	cfg.ReportInterval = 50 * time.Millisecond
	monitor := NewQueryMonitor(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	monitor.Start(ctx)

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{
		QueryID:       "q1",
		Status:        QueryStatusCompleted,
		Data:          []map[string]interface{}{{"id": 1}},
		ExecutionTime: 5 * time.Millisecond,
	}
	monitor.RecordQuery(req, result, nil)

	time.Sleep(200 * time.Millisecond)
	monitor.Stop()
}

func TestQueryMonitor_buildQueryText_EmptyConditions(t *testing.T) {
	cfg := DefaultMonitorConfig()
	cfg.LogFilePath = ""
	monitor := NewQueryMonitor(cfg)

	req := &QueryRequest{
		Table: "devices",
	}

	text := monitor.buildQueryText(req)
	assert.Contains(t, text, "devices")
}

func TestCacheInvalidator_TimeBasedInvalidation(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	invalidator := NewCacheInvalidator(cache)

	invalidator.AddRule(InvalidationRule{
		ID:       "time_rule",
		Name:     "time rule",
		Trigger:  "time",
		Interval: 100 * time.Millisecond,
		Tags:     []string{"table:devices"},
		Enabled:  true,
	})

	ctx := context.Background()
	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{QueryID: "test", Status: QueryStatusCompleted}
	cache.Set(ctx, req, result)

	invalidator.Notify(InvalidationEvent{
		Type:  "time",
		Table: "devices",
	})
}
