package query

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryPlanOptimizer_OptimizeCoverage(t *testing.T) {
	optimizer := NewQueryPlanOptimizer()
	require.NotNil(t, optimizer)
	assert.Len(t, optimizer.rules, 4)

	req := &QueryRequest{
		Type:    QueryTypeSelect,
		Table:   "users",
		Fields:  []string{"id", "name"},
		Conditions: []QueryCondition{
			{Field: "status", Operator: "=", Value: "active"},
		},
		TimeRange: &TimeRange{Field: "created_at"},
		Joins:      []JoinClause{{Type: "INNER", Table: "orders"}},
		Aggregates: []AggregateField{{Field: "amount", Function: "SUM"}},
		OrderBy:    []OrderByField{{Field: "name"}},
	}

	plan, err := optimizer.Optimize(req)
	require.NoError(t, err)
	require.NotNil(t, plan)
	assert.NotEmpty(t, plan.ID)
	assert.NotEmpty(t, plan.Steps)
	assert.Greater(t, plan.EstimatedCost, float64(0))
}

func TestOptimizationRules_NamesCoverage(t *testing.T) {
	rules := []OptimizationRule{
		&IndexScanRule{},
		&PushDownFilterRule{},
		&JoinOrderRule{},
		&ParallelExecutionRule{},
	}
	for _, rule := range rules {
		assert.NotEmpty(t, rule.Name())
	}
}

func TestIndexScanRule_ApplyCoverage(t *testing.T) {
	rule := &IndexScanRule{}
	plan := &QueryPlan{
		Steps: []QueryStep{
			{Type: "scan", Conditions: []string{"status = active"}},
			{Type: "filter", Conditions: []string{"age > 18"}},
		},
	}
	result, err := rule.Apply(plan)
	require.NoError(t, err)
	assert.Equal(t, "index_scan", result.Steps[0].Type)
}

func TestPushDownFilterRule_ApplyCoverage(t *testing.T) {
	rule := &PushDownFilterRule{}
	plan := &QueryPlan{
		Steps: []QueryStep{
			{Type: "scan"},
			{Type: "filter", Conditions: []string{"a > 1"}},
		},
	}
	result, err := rule.Apply(plan)
	require.NoError(t, err)
	assert.Len(t, result.Steps, 1)
}

func TestJoinOrderRule_ApplyCoverage(t *testing.T) {
	rule := &JoinOrderRule{}
	plan := &QueryPlan{Steps: []QueryStep{{Type: "scan"}}}
	result, err := rule.Apply(plan)
	require.NoError(t, err)
	assert.Same(t, plan, result)
}

func TestParallelExecutionRule_Apply_LargeDataCoverage(t *testing.T) {
	rule := &ParallelExecutionRule{}
	plan := &QueryPlan{EstimatedRows: 20000, EstimatedCost: 10.0}
	result, err := rule.Apply(plan)
	require.NoError(t, err)
	assert.True(t, result.Parallel)
}

func TestParallelExecutor_ExecuteCoverage(t *testing.T) {
	executor := NewParallelExecutor(2)
	ctx := context.Background()

	tasks := []func() error{
		func() error { return nil },
		func() error { return nil },
	}

	err := executor.Execute(ctx, tasks)
	assert.NoError(t, err)
}

func TestResultStreamer_Stream_EmptyCoverage(t *testing.T) {
	streamer := NewResultStreamer(100)
	require.NotNil(t, streamer)
}

func TestQueryBuilder_FullChainCoverage(t *testing.T) {
	builder := NewQueryBuilder()
	req := builder.
		Table("users").
		Select("id", "name").
		Where("status", "=", "active").
		OrWhere("role", "!=", "admin").
		TimeRange("created_at", time.Now().Add(-24*time.Hour), time.Now()).
		Join("INNER", "orders", "o", JoinCondition{LeftField: "u.id", Operator: "=", RightField: "o.user_id"}).
		GroupBy("department").
		OrderBy("name", false).
		Limit(10).
		Offset(20).
		Aggregate("salary", "AVG", "avg_salary").
		Priority(PriorityHigh).
		Timeout(5 * time.Second).
		Build()

	require.NotNil(t, req)
	assert.Equal(t, "users", req.Table)
	assert.Contains(t, req.Fields, "id")
	assert.Len(t, req.Conditions, 2)
	assert.Len(t, req.Joins, 1)
	assert.Len(t, req.GroupBy, 1)
	assert.Len(t, req.OrderBy, 1)
	assert.Equal(t, 10, req.Limit)
	assert.Equal(t, 20, req.Offset)
	assert.Len(t, req.Aggregates, 1)
	assert.Equal(t, PriorityHigh, req.Priority)
	assert.Equal(t, 5*time.Second, req.Timeout)
}

func TestQueryBuilder_DefaultValuesCoverage(t *testing.T) {
	builder := NewQueryBuilder()
	req := builder.Build()
	require.NotNil(t, req)
	assert.Empty(t, req.Table)
	assert.Empty(t, req.Fields)
	assert.Empty(t, req.Conditions)
	assert.Equal(t, 0, req.Limit)
}

func TestQueryError_ErrorCoverage(t *testing.T) {
	err := &QueryError{Code: "TEST", Message: "test message"}
	assert.Contains(t, err.Error(), "test message")
	assert.Contains(t, err.Error(), "TEST")
}

func TestQueryRequestJSON_RoundTripCoverage(t *testing.T) {
	req := &QueryRequest{
		ID:       "test-id",
		Type:     QueryTypeSelect,
		Database: "db1",
		Table:    "users",
		Fields:   []string{"id", "name"},
		Priority: PriorityNormal,
		Limit:    10,
	}

	jsonStr, err := QueryRequestJSON(req)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonStr)

	parsed, err := ParseQueryRequestJSON(jsonStr)
	require.NoError(t, err)
	assert.Equal(t, req.ID, parsed.ID)
	assert.Equal(t, req.Type, parsed.Type)
	assert.Equal(t, req.Database, parsed.Database)
	assert.Equal(t, req.Table, parsed.Table)
	assert.Equal(t, req.Priority, parsed.Priority)
	assert.Equal(t, req.Limit, parsed.Limit)
}

func TestParseQueryRequestJSON_InvalidInputCoverage(t *testing.T) {
	_, err := ParseQueryRequestJSON("invalid json")
	assert.Error(t, err)
}

func TestDefaultCacheConfigValueCoverage(t *testing.T) {
	config := DefaultCacheConfig()
	assert.True(t, config.Enabled)
	assert.Equal(t, 5*time.Minute, config.DefaultTTL)
	assert.Equal(t, int64(100*1024*1024), config.MaxSize)
	assert.Equal(t, 10000, config.MaxEntries)
	assert.Equal(t, "LRU", config.EvictionPolicy)
}

func TestCacheEntry_IsExpiredCoverage(t *testing.T) {
	entry := &CacheEntry{ExpiresAt: time.Now().Add(1 * time.Hour)}
	assert.False(t, entry.IsExpired())

	expiredEntry := &CacheEntry{ExpiresAt: time.Now().Add(-1 * time.Second)}
	assert.True(t, expiredEntry.IsExpired())
}

func TestQueryCache_LocalOnlyCoverage(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	ctx := context.Background()

	req := &QueryRequest{
		Type:   QueryTypeSelect,
		Table:  "users",
		Fields: []string{"id"},
	}

	result := &QueryResult{
		QueryID: "q1",
		Status:  QueryStatusCompleted,
		Data:    []map[string]interface{}{{"id": int64(1)}},
		Total:   1,
	}

	err := cache.Set(ctx, req, result)
	require.NoError(t, err)

	got, status, err := cache.Get(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, CacheStatusHit, status)
	assert.Equal(t, result.Total, got.Total)

	stats := cache.GetStats()
	assert.GreaterOrEqual(t, stats.Hits, int64(1))
	assert.GreaterOrEqual(t, stats.TotalRequests, int64(1))
}

func TestQueryCache_DisabledCoverage(t *testing.T) {
	config := DefaultCacheConfig()
	config.Enabled = false
	cache := NewQueryCache(nil, config)
	ctx := context.Background()

	req := &QueryRequest{Table: "t"}
	got, status, err := cache.Get(ctx, req)
	require.NoError(t, err)
	assert.Nil(t, got)
	assert.Equal(t, CacheStatusMiss, status)
}

func TestQueryCache_DeleteCoverage(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	ctx := context.Background()

	req := &QueryRequest{
		Type:   QueryTypeSelect,
		Table:  "users",
		Fields: []string{"id"},
	}

	result := &QueryResult{QueryID: "q1", Data: []map[string]interface{}{{"id": 1}}}
	cache.Set(ctx, req, result)

	err := cache.Delete(ctx, req)
	require.NoError(t, err)

	_, status, _ := cache.Get(ctx, req)
	assert.Equal(t, CacheStatusMiss, status)
}

func TestQueryCache_ClearCoverage(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	ctx := context.Background()

	req := &QueryRequest{Table: "t"}
	result := &QueryResult{QueryID: "q1", Data: []map[string]interface{}{{}}}
	cache.Set(ctx, req, result)

	err := cache.Clear(ctx)
	require.NoError(t, err)

	stats := cache.GetStats()
	assert.Equal(t, int64(0), stats.EntryCount)
}

func TestQueryCache_GenerateKeyAndHashCoverage(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())

	req := &QueryRequest{
		Type:      QueryTypeSelect,
		Table:     "users",
		Fields:    []string{"id", "name"},
		Conditions: []QueryCondition{{Field: "status", Operator: "=", Value: "active"}},
	}

	key := cache.GenerateKey(req)
	assert.NotEmpty(t, key)
	assert.Contains(t, key, cache.config.KeyPrefix)

	hash := cache.GenerateHash(req)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64)
}

func TestQueryCache_ResetStatsCoverage(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	ctx := context.Background()

	req := &QueryRequest{Table: "t"}
	result := &QueryResult{QueryID: "q1", Data: []map[string]interface{}{{}}}
	cache.Set(ctx, req, result)
	cache.Get(ctx, req)

	cache.ResetStats()
	stats := cache.GetStats()
	assert.Equal(t, int64(0), stats.TotalRequests)
	assert.Equal(t, int64(0), stats.Hits)
}

func TestQueryCache_WarmupCoverage(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	ctx := context.Background()

	tasks := []*WarmupTask{
		{Name: "task1", Enabled: false},
		{Name: "task2", Enabled: true},
	}
	err := cache.Warmup(ctx, tasks)
	require.NoError(t, err)
}

func TestQueryCache_DeleteByTagsCoverage(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	ctx := context.Background()

	req1 := &QueryRequest{Type: QueryTypeSelect, Database: "db1", Table: "users"}
	req2 := &QueryRequest{Type: QueryTypeSelect, Database: "db1", Table: "orders"}

	cache.Set(ctx, req1, &QueryResult{QueryID: "q1"})
	cache.Set(ctx, req2, &QueryResult{QueryID: "q2"})

	err := cache.DeleteByTags(ctx, []string{"db:db1"})
	require.NoError(t, err)

	stats := cache.GetStats()
	assert.Equal(t, int64(0), stats.EntryCount)
}

func TestLRUList_PushFrontCoverage(t *testing.T) {
	lru := NewLRUList()
	lru.PushFront("key1")
	lru.PushFront("key2")
	lru.PushFront("key3")

	back, ok := lru.Back()
	assert.True(t, ok)
	assert.Equal(t, "key1", back)
}

func TestLRUList_DuplicateKeyCoverage(t *testing.T) {
	lru := NewLRUList()
	lru.PushFront("key1")
	lru.PushFront("key1")

	back, ok := lru.Back()
	assert.True(t, ok)
	assert.Equal(t, "key1", back)
}

func TestLRUList_MoveToFrontCoverage(t *testing.T) {
	lru := NewLRUList()
	lru.PushFront("key1")
	lru.PushFront("key2")
	lru.PushFront("key3")

	lru.MoveToFront("key1")
	back, _ := lru.Back()
	assert.Equal(t, "key2", back)
}

func TestLRUList_Back_EmptyCoverage(t *testing.T) {
	lru := NewLRUList()
	_, ok := lru.Back()
	assert.False(t, ok)
}

func TestLRUList_RemoveCoverage(t *testing.T) {
	lru := NewLRUList()
	lru.PushFront("key1")
	lru.PushFront("key2")

	lru.Remove("key1")
	back, ok := lru.Back()
	assert.True(t, ok)
	assert.Equal(t, "key2", back)
}

func TestCacheInvalidator_AddRemoveRuleCoverage(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	invalidator := NewCacheInvalidator(cache)

	rule := InvalidationRule{
		ID:       "rule1",
		Name:     "Test Rule",
		Trigger:  "event",
		Tables:   []string{"users", "*"},
		Tags:     []string{"table:users"},
		Enabled:  true,
	}
	invalidator.AddRule(rule)

	invalidator.RemoveRule("rule1")
	invalidator.Notify(InvalidationEvent{Type: "data_change", Table: "users"})
}

func TestCacheInvalidator_TimeBasedRuleCoverage(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	invalidator := NewCacheInvalidator(cache)

	rule := InvalidationRule{
		ID:        "time-rule",
		Trigger:   "time",
		Interval:  10 * time.Millisecond,
		Tags:      []string{"tag1"},
		Enabled:   true,
	}
	invalidator.AddRule(rule)

	time.Sleep(50 * time.Millisecond)
}

func TestCacheWarmer_AddTaskCoverage(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	warmer := NewCacheWarmer(cache, nil)

	warmer.AddTask(&WarmupTask{Name: "warmup1", Enabled: true})
	warmer.Stop()
}

func TestCacheKeyBuilder_BuildCoverage(t *testing.T) {
	builder := NewCacheKeyBuilder()
	key := builder.
		Add("prefix").
		AddInt(42).
		AddInt64(int64(100)).
		AddTime(time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)).
		Build()

	assert.NotEmpty(t, key)
	assert.Len(t, key, 32)
}

func TestMultiLevelCache_GetSetDeleteCoverage(t *testing.T) {
	config := MultiLevelCacheConfig{
		Levels: []CacheLevelConfig{
			{Name: "L1", Priority: 1, Config: DefaultCacheConfig()},
			{Name: "L2", Priority: 2, Config: DefaultCacheConfig()},
		},
	}
	mlc := NewMultiLevelCache(nil, config)
	ctx := context.Background()

	req := &QueryRequest{Type: QueryTypeSelect, Table: "users"}
	result := &QueryResult{QueryID: "q1", Data: []map[string]interface{}{{"id": 1}}}

	err := mlc.Set(ctx, req, result)
	require.NoError(t, err)

	got, status, err := mlc.Get(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, CacheStatusHit, status)
	assert.Equal(t, result.QueryID, got.QueryID)

	err = mlc.Delete(ctx, req)
	require.NoError(t, err)
}

func TestDefaultMonitorConfigValue(t *testing.T) {
	config := DefaultMonitorConfig()
	assert.True(t, config.Enabled)
	assert.Equal(t, 1*time.Second, config.SlowQueryThreshold)
	assert.Equal(t, 1000, config.MaxSlowQueries)
}

func TestDefaultRateLimiterConfigValue(t *testing.T) {
	config := DefaultRateLimiterConfig()
	assert.True(t, config.Enabled)
	assert.Equal(t, "token_bucket", config.Algorithm)
	assert.Equal(t, 100, config.DefaultRate)
	assert.Equal(t, 200, config.DefaultBurst)
}

func TestTokenBucket_TakeCoverage(t *testing.T) {
	bucket := NewTokenBucket(10, 1)
	require.NotNil(t, bucket)

	allowed := bucket.Take()
	assert.True(t, allowed)

	for i := 0; i < 9; i++ {
		bucket.Take()
	}

	allowed = bucket.Take()
	assert.False(t, allowed)
}

func TestSlidingWindow_AllowCoverage(t *testing.T) {
	window := NewSlidingWindow(5, time.Second)
	require.NotNil(t, window)

	for i := 0; i < 5; i++ {
		allowed := window.Allow()
		assert.True(t, allowed)
	}

	allowed := window.Allow()
	assert.False(t, allowed)
}

func TestAdaptiveRateLimiter_RecordSuccessFailureCoverage(t *testing.T) {
	adaptor := NewAdaptiveRateLimiter(100)
	require.NotNil(t, adaptor)

	for i := 0; i < 100; i++ {
		adaptor.RecordSuccess()
	}
	rate := adaptor.GetCurrentRate()
	assert.Greater(t, rate, 0)

	for i := 0; i < 200; i++ {
		adaptor.RecordFailure()
	}
	reducedRate := adaptor.GetCurrentRate()
	assert.LessOrEqual(t, reducedRate, rate)
}

func TestMetricsExporter_ExportPrometheusCoverage(t *testing.T) {
	monitor := NewQueryMonitor(DefaultMonitorConfig())
	exporter := NewMetricsExporter(monitor)
	output := exporter.ExportPrometheus()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "query_total")
}

func TestAlertManager_CheckAndAlertCoverage(t *testing.T) {
	alertManager := NewAlertManager(500 * time.Millisecond)
	req := &QueryRequest{ID: "q1", Type: QueryTypeSelect, Table: "users"}
	result := &QueryResult{QueryID: "q1", Status: QueryStatusFailed}
	alertManager.CheckAndAlert(req, result)
}

func TestPerformanceReporter_SubscribeNotifyCoverage(t *testing.T) {
	reporter := NewPerformanceReporter(1 * time.Minute)
	ch := reporter.Subscribe()
	assert.NotNil(t, ch)

	reporter.AddReport(&PerformanceReport{
		GeneratedAt: time.Now(),
		TotalQueries: 100,
	})

	latest := reporter.GetLatestReport()
	assert.NotNil(t, latest)
}

func TestQueryAnalyzer_AnalyzeCoverage(t *testing.T) {
	analyzer := NewQueryAnalyzer()
	req := &QueryRequest{ID: "q1", Type: QueryTypeSelect, Table: "users"}
	result := &QueryResult{QueryID: "q1", Status: QueryStatusCompleted}
	analyzer.Analyze(req, result)
}

func TestQueryAnalyzer_GetAllPatternsCoverage(t *testing.T) {
	analyzer := NewQueryAnalyzer()
	patterns := analyzer.GetAllPatterns()
	assert.NotNil(t, patterns)
}

func TestDistributedRateLimiter_Coverage(t *testing.T) {
	limiter := NewQueryRateLimiter(DefaultRateLimiterConfig())
	ctx := context.Background()

	result, err := limiter.Allow(ctx, "key1", QueryPriority(0))
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestRateLimitGroup_Coverage(t *testing.T) {
	limiter := NewQueryRateLimiter(DefaultRateLimiterConfig())
	ctx := context.Background()

	result, err := limiter.Allow(ctx, "g1:user1", QueryPriority(0))
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestRateLimitRuleManager_CRUDCoverage(t *testing.T) {
	limiter := NewQueryRateLimiter(DefaultRateLimiterConfig())
	manager := NewRateLimitRuleManager(limiter)
	require.NotNil(t, manager)

	rule := &RateLimitRule{
		ID:          "rule1",
		KeyPattern:  "api:*",
		Rate:        100,
		Burst:       200,
		Enabled:     true,
		Description: "API rate limit",
	}
	manager.AddRule(rule)

	retrieved, ok := manager.GetRule("rule1")
	assert.True(t, ok)
	assert.NotNil(t, retrieved)

	matched := manager.MatchRule("api:users:list")
	assert.NotNil(t, matched)

	manager.RemoveRule("rule1")
	retrieved2, _ := manager.GetRule("rule1")
	assert.Nil(t, retrieved2)
}
