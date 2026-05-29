package rule

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuleEngine_StartStop(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	err := engine.Start()
	require.NoError(t, err)
	assert.True(t, engine.IsRunning())

	err = engine.Start()
	assert.Error(t, err)

	err = engine.Stop()
	require.NoError(t, err)
	assert.False(t, engine.IsRunning())

	err = engine.Stop()
	assert.Error(t, err)
}

func TestRuleEngine_Execute_FormulaRule(t *testing.T) {
	provider := &mockDataProvider{}
	engine := NewRuleEngine(nil, provider)
	rule := &Rule{
		ID:      "r1",
		Name:    "formula rule",
		Type:    RuleTypeFormula,
		Enabled: true,
		PointID: "p1",
		Formula: "42.5",
		Inputs: []RuleInput{
			{Name: "x", PointID: "pa", Required: true},
		},
		Timeout: 5 * time.Second,
	}
	require.NoError(t, engine.LoadRule(rule))

	exec, err := engine.Execute(context.Background(), "r1")
	require.NoError(t, err)
	assert.Equal(t, "success", exec.Status)
}

func TestRuleEngine_Execute_ExpressionRule(t *testing.T) {
	provider := &mockDataProvider{}
	engine := NewRuleEngine(nil, provider)
	rule := &Rule{
		ID:         "r2",
		Name:       "expression rule",
		Type:       RuleTypeExpression,
		Enabled:    true,
		PointID:    "p1",
		Expression: "value > 100",
		Inputs:     []RuleInput{{Name: "value", PointID: "pa"}},
		Timeout:    5 * time.Second,
	}
	require.NoError(t, engine.LoadRule(rule))

	exec, err := engine.Execute(context.Background(), "r2")
	require.NoError(t, err)
	assert.Equal(t, "success", exec.Status)
}

func TestRuleEngine_Execute_ScriptRule(t *testing.T) {
	provider := &mockDataProvider{}
	engine := NewRuleEngine(nil, provider)
	rule := &Rule{
		ID:      "r3",
		Name:    "script rule",
		Type:    RuleTypeScript,
		Enabled: true,
		PointID: "p1",
		Script:  "return x + y",
		Inputs: []RuleInput{
			{Name: "x", PointID: "pa"},
			{Name: "y", PointID: "pb"},
		},
		Timeout: 5 * time.Second,
	}
	require.NoError(t, engine.LoadRule(rule))

	exec, err := engine.Execute(context.Background(), "r3")
	require.NoError(t, err)
	assert.Equal(t, "success", exec.Status)
}

func TestRuleEngine_Execute_AggregateRule(t *testing.T) {
	provider := &mockDataProvider{}
	engine := NewRuleEngine(nil, provider)
	rule := &Rule{
		ID:      "r4",
		Name:    "aggregate rule",
		Type:    RuleTypeAggregate,
		Enabled: true,
		PointID: "p1",
		Inputs:  []RuleInput{{Name: "val", PointID: "pa"}},
		Config: map[string]interface{}{
			"aggregateFunc": "avg",
			"window":        "5m",
		},
		Timeout: 5 * time.Second,
	}
	require.NoError(t, engine.LoadRule(rule))

	exec, err := engine.Execute(context.Background(), "r4")
	require.NoError(t, err)
	assert.Equal(t, "success", exec.Status)
}

func TestRuleEngine_Execute_TransformRule(t *testing.T) {
	provider := &mockDataProvider{}
	engine := NewRuleEngine(nil, provider)
	rule := &Rule{
		ID:      "r5",
		Name:    "transform rule",
		Type:    RuleTypeTransform,
		Enabled: true,
		PointID: "p1",
		Inputs:  []RuleInput{{Name: "val", PointID: "pa"}},
		Config: map[string]interface{}{
			"scale":  2.0,
			"offset": 10.0,
		},
		Timeout: 5 * time.Second,
	}
	require.NoError(t, engine.LoadRule(rule))

	exec, err := engine.Execute(context.Background(), "r5")
	require.NoError(t, err)
	assert.Equal(t, "success", exec.Status)
}

func TestRuleEngine_Execute_DisabledRule(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{
		ID:      "r-disabled",
		Type:    RuleTypeFormula,
		Enabled: false,
		Formula: "1",
		Timeout: 5 * time.Second,
	}
	require.NoError(t, engine.LoadRule(rule))

	_, err := engine.Execute(context.Background(), "r-disabled")
	assert.Equal(t, ErrRuleDisabled, err)
}

func TestRuleEngine_Execute_NotFound(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	_, err := engine.Execute(context.Background(), "nonexistent")
	assert.Equal(t, ErrRuleNotFound, err)
}

func TestRuleEngine_Execute_RequiredInputMissing(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{
		ID:      "r-req",
		Type:    RuleTypeFormula,
		Enabled: true,
		Formula: "x",
		Inputs:  []RuleInput{{Name: "x", PointID: "missing", Required: true}},
		Timeout: 5 * time.Second,
	}
	require.NoError(t, engine.LoadRule(rule))

	exec, err := engine.Execute(context.Background(), "r-req")
	require.NoError(t, err)
	assert.Equal(t, "error", exec.Status)
}

func TestRuleEngine_Execute_OptionalInputDefault(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{
		ID:      "r-opt",
		Type:    RuleTypeFormula,
		Enabled: true,
		Formula: "x",
		Inputs:  []RuleInput{{Name: "x", PointID: "", Default: 99.0, Required: false}},
		Timeout: 5 * time.Second,
	}
	require.NoError(t, engine.LoadRule(rule))

	exec, err := engine.Execute(context.Background(), "r-opt")
	require.NoError(t, err)
	assert.Equal(t, "success", exec.Status)
	assert.Equal(t, 99.0, exec.Inputs["x"])
}

func TestRuleEngine_Execute_NoDataProvider(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{
		ID:      "r-nodp",
		Type:    RuleTypeFormula,
		Enabled: true,
		Formula: "x",
		Inputs:  []RuleInput{{Name: "x", PointID: "some-point", Required: true}},
		Timeout: 5 * time.Second,
	}
	require.NoError(t, engine.LoadRule(rule))

	exec, err := engine.Execute(context.Background(), "r-nodp")
	require.NoError(t, err)
	assert.Equal(t, "error", exec.Status)
}

func TestRuleEngine_Execute_WithCache(t *testing.T) {
	cache := NewComputeCache(&CacheConfig{
		EnableLocalCache: true,
		LocalCacheSize:   100,
		DefaultTTL:       5 * time.Minute,
		Policy:           CachePolicyLRU,
	}, nil)
	provider := &mockDataProvider{}
	engine := NewRuleEngine(cache, provider)

	rule := &Rule{
		ID:      "r-cached",
		Type:    RuleTypeFormula,
		Enabled: true,
		PointID: "p1",
		Formula: "42.5",
		Inputs:  []RuleInput{{Name: "x", PointID: "pa", Required: true}},
		Timeout: 5 * time.Second,
	}
	require.NoError(t, engine.LoadRule(rule))

	exec1, err := engine.Execute(context.Background(), "r-cached")
	require.NoError(t, err)
	assert.Equal(t, "success", exec1.Status)

	exec2, err := engine.Execute(context.Background(), "r-cached")
	require.NoError(t, err)
	assert.Equal(t, "success", exec2.Status)
}

func TestRuleEngine_LoadRule_Duplicate(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{ID: "dup", Type: RuleTypeFormula, Formula: "1", Timeout: 5 * time.Second}
	require.NoError(t, engine.LoadRule(rule))
	err := engine.LoadRule(rule)
	assert.Equal(t, ErrRuleExists, err)
}

func TestRuleEngine_LoadRule_Invalid(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	err := engine.LoadRule(&Rule{ID: ""})
	assert.Equal(t, ErrInvalidRule, err)

	err = engine.LoadRule(&Rule{ID: "r", Type: RuleTypeFormula, Formula: ""})
	assert.Error(t, err)

	err = engine.LoadRule(&Rule{ID: "r", Type: RuleTypeExpression, Expression: ""})
	assert.Error(t, err)

	err = engine.LoadRule(&Rule{ID: "r", Type: RuleTypeScript, Script: ""})
	assert.Error(t, err)
}

func TestRuleEngine_UpdateRule(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{ID: "r-upd", Type: RuleTypeFormula, Formula: "1", Enabled: true, Timeout: 5 * time.Second}
	require.NoError(t, engine.LoadRule(rule))

	updated := &Rule{ID: "r-upd", Type: RuleTypeFormula, Formula: "2", Enabled: true, Timeout: 5 * time.Second}
	require.NoError(t, engine.UpdateRule(updated))

	retrieved, err := engine.GetRule("r-upd")
	require.NoError(t, err)
	assert.Equal(t, 2, retrieved.Version)
}

func TestRuleEngine_UpdateRule_NotFound(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	err := engine.UpdateRule(&Rule{ID: "nonexistent", Type: RuleTypeFormula, Formula: "1"})
	assert.Equal(t, ErrRuleNotFound, err)
}

func TestRuleEngine_UnloadRule(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{ID: "r-unload", Type: RuleTypeFormula, Formula: "1", PointID: "p1", Enabled: true, Timeout: 5 * time.Second}
	require.NoError(t, engine.LoadRule(rule))
	require.NoError(t, engine.UnloadRule("r-unload"))
	_, err := engine.GetRule("r-unload")
	assert.Equal(t, ErrRuleNotFound, err)
}

func TestRuleEngine_UnloadRule_NotFound(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	err := engine.UnloadRule("nonexistent")
	assert.Equal(t, ErrRuleNotFound, err)
}

func TestRuleEngine_EnableDisableRule(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{ID: "r-en", Type: RuleTypeFormula, Formula: "1", Enabled: true, Timeout: 5 * time.Second}
	require.NoError(t, engine.LoadRule(rule))

	require.NoError(t, engine.DisableRule("r-en"))
	r, _ := engine.GetRule("r-en")
	assert.False(t, r.Enabled)
	assert.Equal(t, RuleStatusDisabled, r.Status)

	require.NoError(t, engine.EnableRule("r-en"))
	r, _ = engine.GetRule("r-en")
	assert.True(t, r.Enabled)
	assert.Equal(t, RuleStatusActive, r.Status)
}

func TestRuleEngine_EnableDisable_NotFound(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	assert.Equal(t, ErrRuleNotFound, engine.EnableRule("nonexistent"))
	assert.Equal(t, ErrRuleNotFound, engine.DisableRule("nonexistent"))
}

func TestRuleEngine_GetRulesByPoint(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	r1 := &Rule{ID: "r1", Type: RuleTypeFormula, Formula: "1", PointID: "p1", Enabled: true, Priority: 5, Timeout: 5 * time.Second}
	r2 := &Rule{ID: "r2", Type: RuleTypeFormula, Formula: "2", PointID: "p1", Enabled: true, Priority: 10, Timeout: 5 * time.Second}
	require.NoError(t, engine.LoadRule(r1))
	require.NoError(t, engine.LoadRule(r2))

	rules := engine.GetRulesByPoint("p1")
	assert.Len(t, rules, 2)
	assert.Equal(t, "r2", rules[0].ID)
}

func TestRuleEngine_GetRulesByType(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	r1 := &Rule{ID: "r1", Type: RuleTypeFormula, Formula: "1", Enabled: true, Timeout: 5 * time.Second}
	r2 := &Rule{ID: "r2", Type: RuleTypeExpression, Expression: "1>0", Enabled: true, Timeout: 5 * time.Second}
	require.NoError(t, engine.LoadRule(r1))
	require.NoError(t, engine.LoadRule(r2))

	rules := engine.GetRulesByType(RuleTypeFormula)
	assert.Len(t, rules, 1)
}

func TestRuleEngine_ExecuteForPoint(t *testing.T) {
	provider := &mockDataProvider{}
	engine := NewRuleEngine(nil, provider)
	r1 := &Rule{ID: "r1", Type: RuleTypeFormula, Formula: "1", PointID: "p1", Enabled: true, Inputs: []RuleInput{{Name: "x", PointID: "pa"}}, Timeout: 5 * time.Second}
	r2 := &Rule{ID: "r2", Type: RuleTypeFormula, Formula: "2", PointID: "p1", Enabled: false, Timeout: 5 * time.Second}
	require.NoError(t, engine.LoadRule(r1))
	require.NoError(t, engine.LoadRule(r2))

	execs, err := engine.ExecuteForPoint(context.Background(), "p1")
	require.NoError(t, err)
	assert.Len(t, execs, 1)

	execs, err = engine.ExecuteForPoint(context.Background(), "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, execs)
}

func TestRuleEngine_ExecuteAll(t *testing.T) {
	provider := &mockDataProvider{}
	engine := NewRuleEngine(nil, provider)
	r1 := &Rule{ID: "r1", Type: RuleTypeFormula, Formula: "1", Enabled: true, Inputs: []RuleInput{{Name: "x", PointID: "pa"}}, Timeout: 5 * time.Second}
	r2 := &Rule{ID: "r2", Type: RuleTypeFormula, Formula: "2", Enabled: true, Inputs: []RuleInput{{Name: "y", PointID: "pb"}}, Timeout: 5 * time.Second}
	require.NoError(t, engine.LoadRule(r1))
	require.NoError(t, engine.LoadRule(r2))

	results, err := engine.ExecuteAll(context.Background())
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestRuleEngine_BatchExecute(t *testing.T) {
	provider := &mockDataProvider{}
	engine := NewRuleEngine(nil, provider)
	r1 := &Rule{ID: "r1", Type: RuleTypeFormula, Formula: "1", Enabled: true, Inputs: []RuleInput{{Name: "x", PointID: "pa"}}, Timeout: 5 * time.Second}
	require.NoError(t, engine.LoadRule(r1))

	results, err := engine.BatchExecute(context.Background(), []string{"r1", "nonexistent"})
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "success", results["r1"].Status)
	assert.Equal(t, "error", results["nonexistent"].Status)
}

func TestRuleEngine_ReloadRules(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	r1 := &Rule{ID: "r1", Type: RuleTypeFormula, Formula: "1", Enabled: true, Timeout: 5 * time.Second}
	require.NoError(t, engine.LoadRule(r1))

	newRules := []*Rule{
		{ID: "r2", Type: RuleTypeFormula, Formula: "2", Enabled: true, Timeout: 5 * time.Second},
		{ID: "r3", Type: RuleTypeFormula, Formula: "3", Enabled: false, Timeout: 5 * time.Second},
	}
	require.NoError(t, engine.ReloadRules(newRules))
	assert.Equal(t, 2, engine.GetRuleCount())
}

func TestRuleEngine_ExportRules(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	r1 := &Rule{ID: "r1", Type: RuleTypeFormula, Formula: "1", Enabled: true, Timeout: 5 * time.Second}
	require.NoError(t, engine.LoadRule(r1))

	exported := engine.ExportRules()
	assert.Len(t, exported, 1)
}

func TestRuleEngine_ClearCache(t *testing.T) {
	cache := NewComputeCache(&CacheConfig{EnableLocalCache: true, LocalCacheSize: 100, DefaultTTL: 5 * time.Minute, Policy: CachePolicyLRU}, nil)
	engine := NewRuleEngine(cache, nil)
	require.NoError(t, engine.ClearCache(context.Background()))

	engine2 := NewRuleEngine(nil, nil)
	require.NoError(t, engine2.ClearCache(context.Background()))
}

func TestRuleEngine_GetStats(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	stats := engine.GetStats()
	assert.Equal(t, int64(0), stats.TotalRules)
}

func TestRuleEngine_OutputTransform(t *testing.T) {
	provider := &mockDataProvider{}
	engine := NewRuleEngine(nil, provider)
	rule := &Rule{
		ID:      "r-trans",
		Type:    RuleTypeFormula,
		Enabled: true,
		Formula: "42.5",
		Inputs:  []RuleInput{{Name: "x", PointID: "pa", Required: true}},
		Outputs: []RuleOutput{{Scale: 2.0, Offset: 10.0}},
		Timeout: 5 * time.Second,
	}
	require.NoError(t, engine.LoadRule(rule))

	exec, err := engine.Execute(context.Background(), "r-trans")
	require.NoError(t, err)
	assert.Equal(t, "success", exec.Status)
}

func TestRuleEngine_AggregateRule_NoInputPoint(t *testing.T) {
	engine := NewRuleEngine(nil, &mockDataProvider{})
	rule := &Rule{
		ID:      "r-agg-noinput",
		Type:    RuleTypeAggregate,
		Enabled: true,
		Inputs:  []RuleInput{{Name: "val", PointID: ""}},
		Config:  map[string]interface{}{"aggregateFunc": "avg"},
		Timeout: 5 * time.Second,
	}
	require.NoError(t, engine.LoadRule(rule))

	exec, err := engine.Execute(context.Background(), "r-agg-noinput")
	require.NoError(t, err)
	assert.Equal(t, "error", exec.Status)
}

func TestRuleEngine_AggregateRule_NoDataProvider(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{
		ID:      "r-agg-nodp",
		Type:    RuleTypeAggregate,
		Enabled: true,
		Inputs:  []RuleInput{{Name: "val", PointID: "pa"}},
		Timeout: 5 * time.Second,
	}
	require.NoError(t, engine.LoadRule(rule))

	exec, err := engine.Execute(context.Background(), "r-agg-nodp")
	require.NoError(t, err)
	assert.Equal(t, "error", exec.Status)
}

func TestComputeCache_LocalOnly(t *testing.T) {
	config := &CacheConfig{
		EnableLocalCache: true,
		EnableRedisCache: false,
		LocalCacheSize:   100,
		DefaultTTL:       5 * time.Minute,
		Policy:           CachePolicyLRU,
	}
	cache := NewComputeCache(config, nil)
	ctx := context.Background()

	result := &ComputeResult{PointID: "p1", Value: 100.0, Timestamp: time.Now()}
	require.NoError(t, cache.Set(ctx, "k1", result))

	retrieved, err := cache.Get(ctx, "k1")
	require.NoError(t, err)
	assert.Equal(t, 100.0, retrieved.Value)

	_, err = cache.Get(ctx, "nonexistent")
	assert.Equal(t, ErrCacheNotFound, err)

	require.NoError(t, cache.Delete(ctx, "k1"))
	_, err = cache.Get(ctx, "k1")
	assert.Equal(t, ErrCacheNotFound, err)

	require.NoError(t, cache.Clear(ctx))
	assert.Equal(t, 0, cache.GetSize())
}

func TestComputeCache_Invalidate(t *testing.T) {
	config := &CacheConfig{
		EnableLocalCache: true,
		EnableRedisCache: false,
		LocalCacheSize:   100,
		DefaultTTL:       5 * time.Minute,
		Policy:           CachePolicyLRU,
	}
	cache := NewComputeCache(config, nil)
	ctx := context.Background()

	cache.Set(ctx, "rule:p1", &ComputeResult{PointID: "p1", Value: 1.0, Timestamp: time.Now()})
	cache.Set(ctx, "rule:p2", &ComputeResult{PointID: "p2", Value: 2.0, Timestamp: time.Now()})

	require.NoError(t, cache.Invalidate(ctx, "*p1*"))
	_, err := cache.Get(ctx, "rule:p1")
	assert.Equal(t, ErrCacheNotFound, err)

	_, err = cache.Get(ctx, "rule:p2")
	assert.NoError(t, err)
}

func TestComputeCache_InvalidateByPoint(t *testing.T) {
	config := &CacheConfig{EnableLocalCache: true, LocalCacheSize: 100, DefaultTTL: 5 * time.Minute, Policy: CachePolicyLRU}
	cache := NewComputeCache(config, nil)
	ctx := context.Background()

	cache.Set(ctx, "rule:pointA", &ComputeResult{PointID: "pointA", Value: 1.0, Timestamp: time.Now()})
	require.NoError(t, cache.InvalidateByPoint(ctx, "pointA"))
	_, err := cache.Get(ctx, "rule:pointA")
	assert.Equal(t, ErrCacheNotFound, err)
}

func TestComputeCache_InvalidateByDependency(t *testing.T) {
	config := &CacheConfig{EnableLocalCache: true, LocalCacheSize: 100, DefaultTTL: 5 * time.Minute, Policy: CachePolicyLRU}
	cache := NewComputeCache(config, nil)
	ctx := context.Background()

	cache.Set(ctx, "rule:dep1", &ComputeResult{PointID: "dep1", Value: 1.0, Timestamp: time.Now()})
	require.NoError(t, cache.InvalidateByDependency(ctx, []string{"dep1"}))
}

func TestComputeCache_GetOrCompute(t *testing.T) {
	config := &CacheConfig{EnableLocalCache: true, LocalCacheSize: 100, DefaultTTL: 5 * time.Minute, Policy: CachePolicyLRU}
	cache := NewComputeCache(config, nil)
	ctx := context.Background()

	computed := false
	result, err := cache.GetOrCompute(ctx, "k1", func() (*ComputeResult, error) {
		computed = true
		return &ComputeResult{PointID: "p1", Value: 42.0, Timestamp: time.Now()}, nil
	})
	require.NoError(t, err)
	assert.True(t, computed)
	assert.Equal(t, 42.0, result.Value)

	computed = false
	result, err = cache.GetOrCompute(ctx, "k1", func() (*ComputeResult, error) {
		computed = true
		return &ComputeResult{PointID: "p1", Value: 99.0, Timestamp: time.Now()}, nil
	})
	require.NoError(t, err)
	assert.False(t, computed)
	assert.Equal(t, 42.0, result.Value)

	_, err = cache.GetOrCompute(ctx, "k2", func() (*ComputeResult, error) {
		return nil, errors.New("compute error")
	})
	assert.Error(t, err)
}

func TestComputeCache_BatchOperations(t *testing.T) {
	config := &CacheConfig{EnableLocalCache: true, LocalCacheSize: 100, DefaultTTL: 5 * time.Minute, Policy: CachePolicyLRU}
	cache := NewComputeCache(config, nil)
	ctx := context.Background()

	items := map[string]*ComputeResult{
		"k1": {PointID: "p1", Value: 1.0, Timestamp: time.Now()},
		"k2": {PointID: "p2", Value: 2.0, Timestamp: time.Now()},
	}
	require.NoError(t, cache.BatchSet(ctx, items))

	results, err := cache.BatchGet(ctx, []string{"k1", "k2", "k3"})
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestComputeCache_IsEnabled(t *testing.T) {
	cache := NewComputeCache(&CacheConfig{EnableLocalCache: true, LocalCacheSize: 100, DefaultTTL: 5 * time.Minute, Policy: CachePolicyLRU}, nil)
	assert.True(t, cache.IsEnabled())

	cache2 := NewComputeCache(&CacheConfig{EnableLocalCache: false, EnableRedisCache: false}, nil)
	assert.False(t, cache2.IsEnabled())
}

func TestComputeCache_GetMetrics(t *testing.T) {
	config := &CacheConfig{EnableLocalCache: true, LocalCacheSize: 100, DefaultTTL: 5 * time.Minute, Policy: CachePolicyLRU}
	cache := NewComputeCache(config, nil)
	ctx := context.Background()

	cache.Set(ctx, "k1", &ComputeResult{PointID: "p1", Value: 1.0, Timestamp: time.Now()})
	cache.Get(ctx, "k1")
	cache.Get(ctx, "nonexistent")

	metrics := cache.GetMetrics()
	assert.Equal(t, int64(1), metrics.Hits)
	assert.Equal(t, int64(1), metrics.Misses)
	assert.Equal(t, int64(1), metrics.LocalHits)
}

func TestComputeCache_WithRedis(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	config := &CacheConfig{
		EnableLocalCache: true,
		EnableRedisCache: true,
		LocalCacheSize:   100,
		DefaultTTL:       5 * time.Minute,
		Policy:           CachePolicyLRU,
		RedisKeyPrefix:   "test:",
	}
	cache := NewComputeCache(config, client)
	ctx := context.Background()

	result := &ComputeResult{PointID: "p1", Value: 100.0, Timestamp: time.Now()}
	require.NoError(t, cache.Set(ctx, "k1", result))

	retrieved, err := cache.Get(ctx, "k1")
	require.NoError(t, err)
	assert.Equal(t, 100.0, retrieved.Value)

	require.NoError(t, cache.Delete(ctx, "k1"))
	require.NoError(t, cache.Clear(ctx))
	require.NoError(t, cache.Invalidate(ctx, "*"))
}

func TestComputeCache_SetWithTTL(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	config := &CacheConfig{
		EnableLocalCache: true,
		EnableRedisCache: true,
		LocalCacheSize:   100,
		DefaultTTL:       5 * time.Minute,
		Policy:           CachePolicyLRU,
		RedisKeyPrefix:   "test:",
	}
	cache := NewComputeCache(config, client)
	ctx := context.Background()

	result := &ComputeResult{PointID: "p1", Value: 100.0, Timestamp: time.Now()}
	require.NoError(t, cache.SetWithTTL(ctx, "k1", result, 10*time.Minute))
}

func TestLocalCache_EvictionPolicies(t *testing.T) {
	t.Run("LRU", func(t *testing.T) {
		cache := NewLocalCache(3, 5*time.Minute, CachePolicyLRU)
		cache.Set("k1", &ComputeResult{Value: 1, Timestamp: time.Now()})
		cache.Set("k2", &ComputeResult{Value: 2, Timestamp: time.Now()})
		cache.Set("k3", &ComputeResult{Value: 3, Timestamp: time.Now()})
		cache.Get("k1")
		cache.Get("k2")
		cache.Set("k4", &ComputeResult{Value: 4, Timestamp: time.Now()})
	})

	t.Run("LFU", func(t *testing.T) {
		cache := NewLocalCache(3, 5*time.Minute, CachePolicyLFU)
		cache.Set("k1", &ComputeResult{Value: 1, Timestamp: time.Now()})
		cache.Set("k2", &ComputeResult{Value: 2, Timestamp: time.Now()})
		cache.Set("k3", &ComputeResult{Value: 3, Timestamp: time.Now()})
		cache.Get("k1")
		cache.Get("k1")
		cache.Get("k2")
		cache.Set("k4", &ComputeResult{Value: 4, Timestamp: time.Now()})
	})

	t.Run("FIFO", func(t *testing.T) {
		cache := NewLocalCache(3, 5*time.Minute, CachePolicyFIFO)
		cache.Set("k1", &ComputeResult{Value: 1, Timestamp: time.Now()})
		cache.Set("k2", &ComputeResult{Value: 2, Timestamp: time.Now()})
		cache.Set("k3", &ComputeResult{Value: 3, Timestamp: time.Now()})
		cache.Set("k4", &ComputeResult{Value: 4, Timestamp: time.Now()})
	})
}

func TestLocalCache_Invalidate(t *testing.T) {
	cache := NewLocalCache(100, 5*time.Minute, CachePolicyLRU)
	cache.Set("rule:p1", &ComputeResult{Value: 1, Timestamp: time.Now()})
	cache.Set("rule:p2", &ComputeResult{Value: 2, Timestamp: time.Now()})
	cache.Set("other:p3", &ComputeResult{Value: 3, Timestamp: time.Now()})

	cache.Invalidate("*p1*")
	_, exists := cache.Get("rule:p1")
	assert.False(t, exists)
	_, exists = cache.Get("rule:p2")
	assert.True(t, exists)
}

func TestLocalCache_ExpiredItem(t *testing.T) {
	cache := NewLocalCache(100, 1*time.Nanosecond, CachePolicyLRU)
	cache.Set("k1", &ComputeResult{Value: 1, Timestamp: time.Now()})
	time.Sleep(10 * time.Millisecond)
	_, exists := cache.Get("k1")
	assert.False(t, exists)
}

func TestRedisDistributedLock_WithMiniredis(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	lock := NewRedisDistributedLock(client)
	ctx := context.Background()

	acquired, err := lock.Acquire(ctx, "lock1", 10*time.Second)
	require.NoError(t, err)
	assert.True(t, acquired)

	held, err := lock.IsHeld(ctx, "lock1")
	require.NoError(t, err)
	assert.True(t, held)

	acquired2, err := lock.Acquire(ctx, "lock1", 10*time.Second)
	require.NoError(t, err)
	assert.False(t, acquired2)

	err = lock.Release(ctx, "lock1")
	require.NoError(t, err)

	held, err = lock.IsHeld(ctx, "lock1")
	require.NoError(t, err)
	assert.False(t, held)
}

func TestRedisDistributedLock_TryAcquire(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	lock := NewRedisDistributedLock(client)
	ctx := context.Background()

	acquired, err := lock.TryAcquire(ctx, "lock1", 10*time.Second, 3, 10*time.Millisecond)
	require.NoError(t, err)
	assert.True(t, acquired)
}

func TestRedisDistributedLock_Extend(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	lock := NewRedisDistributedLock(client)
	ctx := context.Background()

	_, _ = lock.Acquire(ctx, "lock1", 10*time.Second)
	err = lock.Extend(ctx, "lock1", 30*time.Second)
	require.NoError(t, err)
}

func TestRedisDistributedLock_ReleaseNotHeld(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	lock := NewRedisDistributedLock(client)
	ctx := context.Background()

	err = lock.Release(ctx, "lock-not-held")
	assert.Equal(t, ErrLockNotHeld, err)
}

func TestRedisDistributedLock_ExtendNotHeld(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	lock := NewRedisDistributedLock(client)
	ctx := context.Background()

	err = lock.Extend(ctx, "lock-not-held", 30*time.Second)
	assert.Equal(t, ErrLockNotHeld, err)
}

func TestLocalLock_ExpiredLock(t *testing.T) {
	lock := NewLocalLock()
	ctx := context.Background()

	acquired, err := lock.Acquire(ctx, "lock1", 1*time.Nanosecond)
	require.NoError(t, err)
	assert.True(t, acquired)

	time.Sleep(10 * time.Millisecond)
	held, err := lock.IsHeld(ctx, "lock1")
	require.NoError(t, err)
	assert.False(t, held)

	acquired2, err := lock.Acquire(ctx, "lock1", 5*time.Second)
	require.NoError(t, err)
	assert.True(t, acquired2)
}

func TestLocalLock_ContextCancellation(t *testing.T) {
	lock := NewLocalLock()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := lock.Acquire(ctx, "lock1", 5*time.Second)
	if err != nil {
		assert.Equal(t, context.Canceled, err)
	}

	err = lock.Release(ctx, "lock1")
	if err != nil {
		assert.Equal(t, context.Canceled, err)
	}

	_, err = lock.IsHeld(ctx, "lock1")
	if err != nil {
		assert.Equal(t, context.Canceled, err)
	}
}

func TestLockManager_LocalMode(t *testing.T) {
	lm := NewLockManager(nil)
	ctx := context.Background()

	acquired, err := lm.Acquire(ctx, "lock1", 5*time.Second)
	require.NoError(t, err)
	assert.True(t, acquired)

	held, err := lm.IsHeld(ctx, "lock1")
	require.NoError(t, err)
	assert.True(t, held)

	require.NoError(t, lm.Release(ctx, "lock1"))
}

func TestLockManager_TryAcquire(t *testing.T) {
	lm := NewLockManager(nil)
	ctx := context.Background()

	acquired, err := lm.TryAcquire(ctx, "lock1", 5*time.Second, 3, 10*time.Millisecond)
	require.NoError(t, err)
	assert.True(t, acquired)
}

func TestLockManager_WithLock(t *testing.T) {
	lm := NewLockManager(nil)
	ctx := context.Background()

	executed := false
	err := lm.WithLock(ctx, "lock1", 5*time.Second, func() error {
		executed = true
		return nil
	})
	require.NoError(t, err)
	assert.True(t, executed)
}

func TestLockManager_WithLock_FailedAcquire(t *testing.T) {
	lm := NewLockManager(nil)
	ctx := context.Background()

	acquired, _ := lm.Acquire(ctx, "lock1", 5*time.Second)
	require.True(t, acquired)

	err := lm.WithLock(ctx, "lock1", 5*time.Second, func() error {
		return nil
	})
	assert.Equal(t, ErrLockFailed, err)
}

func TestLockManager_WithLock_FunctionError(t *testing.T) {
	lm := NewLockManager(nil)
	ctx := context.Background()

	err := lm.WithLock(ctx, "lock1", 5*time.Second, func() error {
		return errors.New("function error")
	})
	assert.Error(t, err)
}

func TestLockManager_RedisMode(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	lm := NewLockManager(client)
	ctx := context.Background()

	acquired, err := lm.Acquire(ctx, "lock1", 5*time.Second)
	require.NoError(t, err)
	assert.True(t, acquired)

	held, err := lm.IsHeld(ctx, "lock1")
	require.NoError(t, err)
	assert.True(t, held)

	require.NoError(t, lm.Release(ctx, "lock1"))
}

func TestPointManager_FullCRUD(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	p1 := &ComputePoint{ID: "p1", Name: "Point 1", Type: PointTypeVirtual, Dependencies: []string{}}
	require.NoError(t, pm.CreatePoint(ctx, p1))

	p2 := &ComputePoint{ID: "p2", Name: "Point 2", Type: PointTypeDerived, Dependencies: []string{"p1"}}
	require.NoError(t, pm.CreatePoint(ctx, p2))

	retrieved, err := pm.GetPoint(ctx, "p1")
	require.NoError(t, err)
	assert.Equal(t, "Point 1", retrieved.Name)

	_, err = pm.GetPoint(ctx, "nonexistent")
	assert.Equal(t, ErrPointNotFound, err)

	p1Updated := &ComputePoint{ID: "p1", Name: "Updated", Type: PointTypeVirtual, Dependencies: []string{}}
	require.NoError(t, pm.UpdatePoint(ctx, p1Updated))
	updated, _ := pm.GetPoint(ctx, "p1")
	assert.Equal(t, "Updated", updated.Name)

	err = pm.UpdatePoint(ctx, &ComputePoint{ID: "nonexistent"})
	assert.Equal(t, ErrPointNotFound, err)

	err = pm.DeletePoint(ctx, "p2")
	require.NoError(t, err)

	err = pm.DeletePoint(ctx, "p1")
	require.NoError(t, err)

	err = pm.DeletePoint(ctx, "nonexistent")
	assert.Equal(t, ErrPointNotFound, err)
}

func TestPointManager_DeleteWithDependents(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	p1 := &ComputePoint{ID: "p1", Type: PointTypeVirtual, Dependencies: []string{}}
	p2 := &ComputePoint{ID: "p2", Type: PointTypeDerived, Dependencies: []string{"p1"}}
	require.NoError(t, pm.CreatePoint(ctx, p1))
	require.NoError(t, pm.CreatePoint(ctx, p2))

	err := pm.DeletePoint(ctx, "p1")
	assert.Error(t, err)
}

func TestPointManager_CreateInvalid(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	err := pm.CreatePoint(ctx, &ComputePoint{ID: ""})
	assert.Equal(t, ErrInvalidPointConfig, err)

	err = pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual, Dependencies: []string{}})
	require.NoError(t, err)
	err = pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual, Dependencies: []string{}})
	assert.Equal(t, ErrPointExists, err)
}

func TestPointManager_GetByTypeAndStatus(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual, Status: PointStatusActive, Dependencies: []string{}})
	pm.CreatePoint(ctx, &ComputePoint{ID: "p2", Type: PointTypeDerived, Status: PointStatusActive, Dependencies: []string{}})
	pm.CreatePoint(ctx, &ComputePoint{ID: "p3", Type: PointTypeVirtual, Status: PointStatusError, Dependencies: []string{}})

	virtual := pm.GetPointsByType(ctx, PointTypeVirtual)
	assert.Len(t, virtual, 2)

	active := pm.GetPointsByStatus(ctx, PointStatusActive)
	assert.Len(t, active, 2)

	all := pm.GetAllPoints(ctx)
	assert.Len(t, all, 3)
}

func TestPointManager_ConfigOperations(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual, Dependencies: []string{}})

	config := &PointConfig{ComputeInterval: 1 * time.Second, Timeout: 5 * time.Second}
	require.NoError(t, pm.SetPointConfig(ctx, "p1", config))

	retrieved, err := pm.GetPointConfig(ctx, "p1")
	require.NoError(t, err)
	assert.Equal(t, 1*time.Second, retrieved.ComputeInterval)

	_, err = pm.GetPointConfig(ctx, "nonexistent")
	assert.Equal(t, ErrPointNotFound, err)

	err = pm.SetPointConfig(ctx, "nonexistent", config)
	assert.Equal(t, ErrPointNotFound, err)
}

func TestPointManager_StatusAndValue(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual, Status: PointStatusActive, Dependencies: []string{}})

	require.NoError(t, pm.UpdatePointStatus(ctx, "p1", PointStatusError))
	p, _ := pm.GetPoint(ctx, "p1")
	assert.Equal(t, PointStatusError, p.Status)

	require.NoError(t, pm.UpdatePointValue(ctx, "p1", 42.5, 100))
	p, _ = pm.GetPoint(ctx, "p1")
	assert.Equal(t, 42.5, p.Value)
	assert.Equal(t, 100, p.Quality)

	err := pm.UpdatePointStatus(ctx, "nonexistent", PointStatusActive)
	assert.Equal(t, ErrPointNotFound, err)

	err = pm.UpdatePointValue(ctx, "nonexistent", 1.0, 100)
	assert.Equal(t, ErrPointNotFound, err)
}

func TestPointManager_Dependencies(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual, Dependencies: []string{}})
	pm.CreatePoint(ctx, &ComputePoint{ID: "p2", Type: PointTypeDerived, Dependencies: []string{"p1"}})

	dep, err := pm.GetDependencies(ctx, "p2")
	require.NoError(t, err)
	assert.Equal(t, "p2", dep.PointID)
	assert.Contains(t, dep.DependsOn, "p1")

	_, err = pm.GetDependencies(ctx, "nonexistent")
	assert.Equal(t, ErrPointNotFound, err)

	dependents := pm.GetDependents(ctx, "p1")
	assert.Contains(t, dependents, "p2")

	order := pm.GetComputeOrder(ctx)
	assert.NotEmpty(t, order)
}

func TestPointManager_Filter(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual, Status: PointStatusActive, Tags: map[string]string{"env": "prod"}, Dependencies: []string{}})
	pm.CreatePoint(ctx, &ComputePoint{ID: "p2", Type: PointTypeDerived, Status: PointStatusInactive, Tags: map[string]string{"env": "dev"}, Dependencies: []string{}})
	pm.CreatePoint(ctx, &ComputePoint{ID: "p3", Type: PointTypeVirtual, Status: PointStatusActive, Tags: map[string]string{"env": "prod"}, Dependencies: []string{}})

	filtered := pm.Filter(ctx, &PointFilter{Types: []PointType{PointTypeVirtual}})
	assert.Len(t, filtered, 2)

	filtered = pm.Filter(ctx, &PointFilter{Status: []PointStatus{PointStatusActive}})
	assert.Len(t, filtered, 2)

	filtered = pm.Filter(ctx, &PointFilter{Tags: map[string]string{"env": "prod"}})
	assert.Len(t, filtered, 2)

	filtered = pm.Filter(ctx, &PointFilter{IDs: []string{"p1", "p3"}})
	assert.Len(t, filtered, 2)

	filtered = pm.Filter(ctx, &PointFilter{Types: []PointType{PointTypeAggregate}})
	assert.Len(t, filtered, 0)
}

func TestPointManager_Stats(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual, Status: PointStatusActive, Dependencies: []string{}})
	pm.CreatePoint(ctx, &ComputePoint{ID: "p2", Type: PointTypeDerived, Status: PointStatusError, Dependencies: []string{}})

	stats := pm.GetStats(ctx)
	assert.Equal(t, 2, stats.TotalPoints)
	assert.Equal(t, 1, stats.ActivePoints)
	assert.Equal(t, 1, stats.ErrorPoints)
	assert.Equal(t, 1, stats.ByType[PointTypeVirtual])
}

func TestDependencyGraph_CycleDetection(t *testing.T) {
	dg := NewDependencyGraph()
	require.NoError(t, dg.AddNode("A", []string{}))
	require.NoError(t, dg.AddNode("B", []string{"A"}))

	hasCycle := dg.HasCycle("C", []string{"C"})
	assert.True(t, hasCycle)

	hasCycle = dg.HasCycle("D", []string{"A"})
	assert.False(t, hasCycle)
}

func TestDependencyGraph_UpdateNode(t *testing.T) {
	dg := NewDependencyGraph()
	require.NoError(t, dg.AddNode("A", []string{}))

	err := dg.UpdateNode("nonexistent", []string{})
	assert.Error(t, err)

	require.NoError(t, dg.UpdateNode("A", []string{"B"}))
}

func TestDependencyGraph_DuplicateNode(t *testing.T) {
	dg := NewDependencyGraph()
	require.NoError(t, dg.AddNode("A", []string{}))
	err := dg.AddNode("A", []string{})
	assert.Error(t, err)
}

func TestDependencyGraph_RemoveNode(t *testing.T) {
	dg := NewDependencyGraph()
	dg.AddNode("A", []string{})
	dg.RemoveNode("A")
	dg.RemoveNode("nonexistent")
}

func TestComputeScheduler_BasicOperations(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	require.NoError(t, scheduler.Start())
	assert.True(t, scheduler.IsRunning())

	err := scheduler.Start()
	assert.Error(t, err)

	task := &ComputeTask{
		ID:       "t1",
		Name:     "Task 1",
		Type:     TaskTypeInterval,
		Interval: 1 * time.Hour,
		Enabled:  true,
		Timeout:  5 * time.Second,
	}
	require.NoError(t, scheduler.AddTask(task))
	assert.Equal(t, 1, scheduler.GetTaskCount())

	retrieved, err := scheduler.GetTask("t1")
	require.NoError(t, err)
	assert.Equal(t, "Task 1", retrieved.Name)

	_, err = scheduler.GetTask("nonexistent")
	assert.Equal(t, ErrTaskNotFound, err)

	allTasks := scheduler.GetAllTasks()
	assert.Len(t, allTasks, 1)

	require.NoError(t, scheduler.Stop())
	assert.False(t, scheduler.IsRunning())

	err = scheduler.Stop()
	assert.Error(t, err)
}

func TestComputeScheduler_AddInvalidTask(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	err := scheduler.AddTask(nil)
	assert.Equal(t, ErrInvalidTask, err)

	err = scheduler.AddTask(&ComputeTask{ID: ""})
	assert.Equal(t, ErrInvalidTask, err)
}

func TestComputeScheduler_AddDuplicateTask(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	task := &ComputeTask{ID: "t1", Type: TaskTypeInterval, Interval: 1 * time.Hour, Enabled: true, Timeout: 5 * time.Second}
	require.NoError(t, scheduler.AddTask(task))
	err := scheduler.AddTask(task)
	assert.Equal(t, ErrTaskExists, err)
}

func TestComputeScheduler_RemoveTask(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	task := &ComputeTask{ID: "t1", Type: TaskTypeInterval, Interval: 1 * time.Hour, Enabled: true, Timeout: 5 * time.Second}
	require.NoError(t, scheduler.AddTask(task))
	require.NoError(t, scheduler.RemoveTask("t1"))
	assert.Equal(t, 0, scheduler.GetTaskCount())

	err := scheduler.RemoveTask("nonexistent")
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestComputeScheduler_UpdateTask(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	task := &ComputeTask{ID: "t1", Name: "Original", Type: TaskTypeInterval, Interval: 1 * time.Hour, Enabled: true, Timeout: 5 * time.Second}
	require.NoError(t, scheduler.AddTask(task))

	updated := &ComputeTask{ID: "t1", Name: "Updated", Type: TaskTypeInterval, Interval: 30 * time.Minute, Enabled: true, Timeout: 10 * time.Second}
	require.NoError(t, scheduler.UpdateTask(updated))

	retrieved, _ := scheduler.GetTask("t1")
	assert.Equal(t, "Updated", retrieved.Name)

	err := scheduler.UpdateTask(&ComputeTask{ID: "nonexistent"})
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestComputeScheduler_PauseResumeTask(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	task := &ComputeTask{ID: "t1", Type: TaskTypeInterval, Interval: 1 * time.Hour, Enabled: true, Timeout: 5 * time.Second}
	require.NoError(t, scheduler.AddTask(task))

	require.NoError(t, scheduler.PauseTask("t1"))
	t1, _ := scheduler.GetTask("t1")
	assert.Equal(t, TaskStatusPaused, t1.Status)
	assert.False(t, t1.Enabled)

	require.NoError(t, scheduler.ResumeTask("t1"))
	t1, _ = scheduler.GetTask("t1")
	assert.Equal(t, TaskStatusPending, t1.Status)
	assert.True(t, t1.Enabled)

	err := scheduler.PauseTask("nonexistent")
	assert.Equal(t, ErrTaskNotFound, err)

	err = scheduler.ResumeTask("nonexistent")
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestComputeScheduler_TriggerTask(t *testing.T) {
	executor := &mockExecutor{}
	config := &SchedulerConfig{MaxConcurrentTasks: 1, TaskQueueSize: 10, ScheduleInterval: 1 * time.Hour}
	scheduler := NewComputeScheduler(config, executor, nil)
	require.NoError(t, scheduler.Start())
	defer scheduler.Stop()

	task := &ComputeTask{ID: "t1", Type: TaskTypeInterval, Interval: 1 * time.Hour, Enabled: true, Timeout: 5 * time.Second, PointIDs: []string{"p1"}}
	require.NoError(t, scheduler.AddTask(task))

	err := scheduler.TriggerTask("t1")
	assert.NoError(t, err)

	err = scheduler.TriggerTask("nonexistent")
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestComputeScheduler_CronTask(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	task := &ComputeTask{ID: "t-cron", Type: TaskTypeCron, CronExpr: "@hourly", Enabled: true, Timeout: 5 * time.Second}
	require.NoError(t, scheduler.AddTask(task))
	t1, _ := scheduler.GetTask("t-cron")
	assert.False(t, t1.NextRunTime.IsZero())
}

func TestComputeScheduler_OnceTask(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	task := &ComputeTask{ID: "t-once", Type: TaskTypeOnce, Enabled: true, Timeout: 5 * time.Second}
	require.NoError(t, scheduler.AddTask(task))
}

func TestComputeScheduler_DisabledTask(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	task := &ComputeTask{ID: "t-disabled", Type: TaskTypeInterval, Interval: 1 * time.Hour, Enabled: false, Timeout: 5 * time.Second}
	require.NoError(t, scheduler.AddTask(task))
	t1, _ := scheduler.GetTask("t-disabled")
	assert.Equal(t, TaskStatusPaused, t1.Status)
}

func TestComputeScheduler_GetMetrics(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	metrics := scheduler.GetMetrics()
	assert.Equal(t, int64(0), metrics.TotalTasks)
}

func TestComputeScheduler_GetLogs(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	logs := scheduler.GetLogs("", 10)
	assert.Empty(t, logs)
}

func TestComputeScheduler_ParseCronNextTime(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	now := time.Now()
	tests := []struct {
		expr     string
		expected time.Duration
	}{
		{"@every 1m", 1 * time.Minute},
		{"@every 5m", 5 * time.Minute},
		{"@every 15m", 15 * time.Minute},
		{"@every 30m", 30 * time.Minute},
		{"@hourly", 1 * time.Hour},
		{"@daily", 24 * time.Hour},
		{"@weekly", 7 * 24 * time.Hour},
		{"unknown", 1 * time.Minute},
	}
	for _, tt := range tests {
		next := scheduler.parseCronNextTime(tt.expr, now)
		assert.WithinDuration(t, now.Add(tt.expected), next, time.Second, "expr: %s", tt.expr)
	}
}

func TestPriorityQueue_Operations(t *testing.T) {
	pq := NewPriorityQueue()

	assert.Equal(t, 0, pq.Len())

	pq.Push(&ComputeTask{ID: "t1", Priority: 5, NextRunTime: time.Now().Add(1 * time.Hour)})
	pq.Push(&ComputeTask{ID: "t2", Priority: 10, NextRunTime: time.Now().Add(2 * time.Hour)})
	pq.Push(&ComputeTask{ID: "t3", Priority: 3, NextRunTime: time.Now()})
	assert.Equal(t, 3, pq.Len())

	first := pq.Pop()
	assert.Equal(t, "t2", first.ID)

	second := pq.Pop()
	assert.Equal(t, "t1", second.ID)

	third := pq.Pop()
	assert.Equal(t, "t3", third.ID)

	nilTask := pq.Pop()
	assert.Nil(t, nilTask)
}

func TestTriggerManager_FullCRUD(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	trigger := &Trigger{
		ID:       "tr1",
		Name:     "Trigger 1",
		Type:     TriggerTypeDataChange,
		Enabled:  true,
		PointIDs: []string{"p1"},
		Condition: &TriggerCondition{
			ChangeThreshold: floatPtr(10.0),
		},
	}
	require.NoError(t, tm.CreateTrigger(trigger))

	retrieved, err := tm.GetTrigger("tr1")
	require.NoError(t, err)
	assert.Equal(t, "Trigger 1", retrieved.Name)

	_, err = tm.GetTrigger("nonexistent")
	assert.Equal(t, ErrTriggerNotFound, err)

	err = tm.CreateTrigger(&Trigger{ID: ""})
	assert.Equal(t, ErrInvalidTrigger, err)

	err = tm.CreateTrigger(trigger)
	assert.Equal(t, ErrTriggerExists, err)
}

func TestTriggerManager_UpdateTrigger(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	trigger := &Trigger{ID: "tr1", Name: "Original", Type: TriggerTypeDataChange, Enabled: true, PointIDs: []string{"p1"}}
	require.NoError(t, tm.CreateTrigger(trigger))

	updated := &Trigger{ID: "tr1", Name: "Updated", Type: TriggerTypeEvent, Enabled: true, PointIDs: []string{"p2"}}
	require.NoError(t, tm.UpdateTrigger(updated))

	err := tm.UpdateTrigger(&Trigger{ID: "nonexistent"})
	assert.Equal(t, ErrTriggerNotFound, err)
}

func TestTriggerManager_DeleteTrigger(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	trigger := &Trigger{ID: "tr1", Type: TriggerTypeDataChange, Enabled: true, PointIDs: []string{"p1"}}
	require.NoError(t, tm.CreateTrigger(trigger))
	require.NoError(t, tm.DeleteTrigger("tr1"))

	err := tm.DeleteTrigger("nonexistent")
	assert.Equal(t, ErrTriggerNotFound, err)
}

func TestTriggerManager_GetByPointAndType(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	tm.CreateTrigger(&Trigger{ID: "tr1", Type: TriggerTypeDataChange, Enabled: true, PointIDs: []string{"p1"}, Priority: 5})
	tm.CreateTrigger(&Trigger{ID: "tr2", Type: TriggerTypeDataChange, Enabled: true, PointIDs: []string{"p1"}, Priority: 10})
	tm.CreateTrigger(&Trigger{ID: "tr3", Type: TriggerTypeEvent, Enabled: true, PointIDs: []string{"p2"}})

	byPoint := tm.GetTriggersByPoint("p1")
	assert.Len(t, byPoint, 2)
	assert.Equal(t, "tr2", byPoint[0].ID)

	byType := tm.GetTriggersByType(TriggerTypeDataChange)
	assert.Len(t, byType, 2)
}

func TestTriggerManager_EnableDisable(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	tm.CreateTrigger(&Trigger{ID: "tr1", Type: TriggerTypeDataChange, Enabled: true})

	require.NoError(t, tm.DisableTrigger("tr1"))
	tr1, _ := tm.GetTrigger("tr1")
	assert.False(t, tr1.Enabled)
	assert.Equal(t, TriggerStatusDisabled, tr1.Status)

	require.NoError(t, tm.EnableTrigger("tr1"))
	tr1, _ = tm.GetTrigger("tr1")
	assert.True(t, tr1.Enabled)
	assert.Equal(t, TriggerStatusActive, tr1.Status)

	err := tm.EnableTrigger("nonexistent")
	assert.Equal(t, ErrTriggerNotFound, err)

	err = tm.DisableTrigger("nonexistent")
	assert.Equal(t, ErrTriggerNotFound, err)
}

func TestTriggerManager_StartStop(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	require.NoError(t, tm.Start())
	err := tm.Start()
	assert.Error(t, err)

	require.NoError(t, tm.Stop())
	err = tm.Stop()
	assert.Error(t, err)
}

func TestTriggerManager_OnDataChange(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)
	require.NoError(t, tm.Start())
	defer tm.Stop()

	tm.CreateTrigger(&Trigger{
		ID:       "tr1",
		Type:     TriggerTypeDataChange,
		Enabled:  true,
		PointIDs: []string{"p1"},
		Condition: &TriggerCondition{
			ChangeThreshold: floatPtr(5.0),
		},
	})

	err := tm.OnDataChange(context.Background(), "p1", 10.0, 20.0)
	assert.NoError(t, err)
}

func TestTriggerManager_OnEvent(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)
	require.NoError(t, tm.Start())
	defer tm.Stop()

	tm.CreateTrigger(&Trigger{
		ID:      "tr1",
		Type:    TriggerTypeEvent,
		Enabled: true,
		Condition: &TriggerCondition{
			EventType: "alert",
		},
	})

	err := tm.OnEvent(context.Background(), "alert", map[string]interface{}{"data": "test"})
	assert.NoError(t, err)
}

func TestTriggerManager_OnCondition(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)
	require.NoError(t, tm.Start())
	defer tm.Stop()

	tm.CreateTrigger(&Trigger{
		ID:       "tr1",
		Type:     TriggerTypeCondition,
		Enabled:  true,
		PointIDs: []string{"p1"},
		Condition: &TriggerCondition{
			Expression: "value > 100",
		},
	})

	err := tm.OnCondition(context.Background(), "p1", "value > 100", 150.0)
	assert.NoError(t, err)
}

func TestTriggerManager_EvaluateDataChange(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	trigger := &Trigger{
		Type: TriggerTypeDataChange,
		Condition: &TriggerCondition{
			ChangeThreshold: floatPtr(5.0),
		},
	}

	assert.True(t, tm.evaluateDataChange(trigger, &TriggerEvent{OldValue: 10.0, NewValue: 20.0}))
	assert.False(t, tm.evaluateDataChange(trigger, &TriggerEvent{OldValue: 10.0, NewValue: 12.0}))
	assert.False(t, tm.evaluateDataChange(trigger, &TriggerEvent{OldValue: 10.0, NewValue: 6.0}))

	upTrigger := &Trigger{
		Type: TriggerTypeDataChange,
		Condition: &TriggerCondition{
			Direction: "up",
		},
	}
	assert.True(t, tm.evaluateDataChange(upTrigger, &TriggerEvent{OldValue: 10.0, NewValue: 20.0}))
	assert.False(t, tm.evaluateDataChange(upTrigger, &TriggerEvent{OldValue: 20.0, NewValue: 10.0}))

	downTrigger := &Trigger{
		Type: TriggerTypeDataChange,
		Condition: &TriggerCondition{
			Direction: "down",
		},
	}
	assert.True(t, tm.evaluateDataChange(downTrigger, &TriggerEvent{OldValue: 20.0, NewValue: 10.0}))
	assert.False(t, tm.evaluateDataChange(downTrigger, &TriggerEvent{OldValue: 10.0, NewValue: 20.0}))

	noCondTrigger := &Trigger{Type: TriggerTypeDataChange}
	assert.False(t, tm.evaluateDataChange(noCondTrigger, &TriggerEvent{}))
}

func TestTriggerManager_EvaluateDataChange_Percent(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	trigger := &Trigger{
		Type: TriggerTypeDataChange,
		Condition: &TriggerCondition{
			ChangePercent: floatPtr(50.0),
		},
	}

	assert.True(t, tm.evaluateDataChange(trigger, &TriggerEvent{OldValue: 100.0, NewValue: 200.0}))
	assert.False(t, tm.evaluateDataChange(trigger, &TriggerEvent{OldValue: 100.0, NewValue: 120.0}))
}

func TestTriggerManager_EvaluateEvent(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	trigger := &Trigger{
		Type: TriggerTypeEvent,
		Condition: &TriggerCondition{
			EventType: "alert",
		},
	}

	assert.True(t, tm.evaluateEvent(trigger, &TriggerEvent{Metadata: map[string]string{"eventType": "alert"}}))
	assert.False(t, tm.evaluateEvent(trigger, &TriggerEvent{Metadata: map[string]string{"eventType": "info"}}))

	noCondTrigger := &Trigger{Type: TriggerTypeEvent}
	assert.True(t, tm.evaluateEvent(noCondTrigger, &TriggerEvent{}))
}

func TestTriggerManager_EvaluateEvent_WithData(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	trigger := &Trigger{
		Type: TriggerTypeEvent,
		Condition: &TriggerCondition{
			EventType: "alert",
			EventData: "critical",
		},
	}

	assert.True(t, tm.evaluateEvent(trigger, &TriggerEvent{
		Metadata: map[string]string{"eventType": "alert"},
		Payload:  map[string]interface{}{"data": "critical"},
	}))

	assert.False(t, tm.evaluateEvent(trigger, &TriggerEvent{
		Metadata: map[string]string{"eventType": "alert"},
		Payload:  map[string]interface{}{"data": "warning"},
	}))
}

func TestTriggerManager_EvaluateCondition(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	trigger := &Trigger{
		Type: TriggerTypeCondition,
		Condition: &TriggerCondition{
			Expression: "value > 100",
		},
	}

	assert.True(t, tm.evaluateCondition(trigger, &TriggerEvent{NewValue: 150.0}))
	assert.False(t, tm.evaluateCondition(trigger, &TriggerEvent{NewValue: -5.0}))

	noCondTrigger := &Trigger{Type: TriggerTypeCondition}
	assert.False(t, tm.evaluateCondition(noCondTrigger, &TriggerEvent{}))
}

func TestTriggerManager_EvaluateComposite(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	threshold := floatPtr(5.0)

	andTrigger := &Trigger{
		Type: TriggerTypeComposite,
		Condition: &TriggerCondition{
			LogicOperator: "and",
			SubConditions: []*TriggerCondition{
				{ChangeThreshold: threshold},
				{Direction: "up"},
			},
		},
	}
	assert.True(t, tm.evaluateComposite(andTrigger, &TriggerEvent{OldValue: 10.0, NewValue: 20.0}))
	assert.False(t, tm.evaluateComposite(andTrigger, &TriggerEvent{OldValue: 20.0, NewValue: 10.0}))

	orTrigger := &Trigger{
		Type: TriggerTypeComposite,
		Condition: &TriggerCondition{
			LogicOperator: "or",
			SubConditions: []*TriggerCondition{
				{ChangeThreshold: threshold},
				{Direction: "down"},
			},
		},
	}
	assert.True(t, tm.evaluateComposite(orTrigger, &TriggerEvent{OldValue: 10.0, NewValue: 20.0}))

	noCondTrigger := &Trigger{Type: TriggerTypeComposite}
	assert.False(t, tm.evaluateComposite(noCondTrigger, &TriggerEvent{}))

	unknownOpTrigger := &Trigger{
		Type: TriggerTypeComposite,
		Condition: &TriggerCondition{
			LogicOperator: "xor",
			SubConditions: []*TriggerCondition{{ChangeThreshold: threshold}},
		},
	}
	assert.False(t, tm.evaluateComposite(unknownOpTrigger, &TriggerEvent{OldValue: 10.0, NewValue: 20.0}))
}

func TestTriggerManager_ChainOperations(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	err := tm.CreateChain(&TriggerChain{ID: "", Name: "Invalid"})
	assert.Error(t, err)

	require.NoError(t, tm.CreateChain(&TriggerChain{ID: "chain1", Name: "Chain 1"}))

	err = tm.CreateChain(&TriggerChain{ID: "chain1", Name: "Duplicate"})
	assert.Error(t, err)

	chain, err := tm.GetChain("chain1")
	require.NoError(t, err)
	assert.Equal(t, "Chain 1", chain.Name)

	_, err = tm.GetChain("nonexistent")
	assert.Error(t, err)
}

func TestTriggerManager_GetStats(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	tm.CreateTrigger(&Trigger{ID: "tr1", Type: TriggerTypeDataChange, Enabled: true})
	tm.CreateTrigger(&Trigger{ID: "tr2", Type: TriggerTypeEvent, Enabled: true})

	stats := tm.GetStats()
	assert.Equal(t, 2, stats["totalTriggers"])
}

func TestTriggerManager_ExecuteActions(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	trigger := &Trigger{
		ID:      "tr1",
		Enabled: true,
		Actions: []TriggerAction{
			{Type: "compute", Target: "p1", Timeout: 5 * time.Second},
			{Type: "notify", Target: "admin"},
			{Type: "script", Target: "cleanup"},
		},
	}

	result := &TriggerResult{Context: make(map[string]interface{})}
	tm.executeActions(trigger, result)
	assert.True(t, result.Executed)
}

func TestTriggerManager_ExecuteAction_FailedSync(t *testing.T) {
	executor := &mockFailingExecutor{}
	tm := NewTriggerManager(executor)

	trigger := &Trigger{
		ID:      "tr1",
		Enabled: true,
		Actions: []TriggerAction{
			{Type: "compute", Target: "p1", Timeout: 5 * time.Second},
		},
	}

	result := &TriggerResult{Context: make(map[string]interface{})}
	tm.executeActions(trigger, result)
	assert.False(t, result.Executed)
}

func TestTriggerManager_ExecuteChain(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	tm.CreateTrigger(&Trigger{
		ID:       "tr1",
		Type:     TriggerTypeDataChange,
		Enabled:  true,
		PointIDs: []string{"p1"},
		Condition: &TriggerCondition{
			ChangeThreshold: floatPtr(5.0),
		},
		ChainConfig: &TriggerChainConfig{
			NextTriggers: []string{"tr2"},
		},
	})

	tm.CreateTrigger(&Trigger{
		ID:       "tr2",
		Type:     TriggerTypeDataChange,
		Enabled:  true,
		PointIDs: []string{"p1"},
		Condition: &TriggerCondition{
			ChangeThreshold: floatPtr(1.0),
		},
	})

	result := &TriggerResult{Executed: true}
	event := &TriggerEvent{OldValue: 10.0, NewValue: 20.0}
	tm.executeChain(&Trigger{ChainConfig: &TriggerChainConfig{NextTriggers: []string{"tr2"}}}, event, result)
}

func TestTriggerManager_ExecuteChain_StopOnSuccess(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	trigger := &Trigger{
		ChainConfig: &TriggerChainConfig{
			NextTriggers:  []string{"tr2"},
			StopOnSuccess: true,
		},
	}
	result := &TriggerResult{Executed: true}
	event := &TriggerEvent{}
	tm.executeChain(trigger, event, result)
}

func TestTriggerManager_ExecuteChain_StopOnFailure(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	trigger := &Trigger{
		ChainConfig: &TriggerChainConfig{
			NextTriggers:  []string{"tr2"},
			StopOnFailure: true,
		},
	}
	result := &TriggerResult{Executed: false}
	event := &TriggerEvent{}
	tm.executeChain(trigger, event, result)
}

type mockFailingExecutor struct{}

func (m *mockFailingExecutor) Execute(ctx context.Context, pointIDs []string) (map[string]*ComputeResult, error) {
	return nil, errors.New("execution failed")
}

func TestContainsPattern(t *testing.T) {
	assert.True(t, containsPattern("rule:p1", "*p1*"))
	assert.True(t, containsPattern("exact", "exact"))
	assert.True(t, containsPattern("any", ""))
	assert.False(t, containsPattern("rule:p1", "*p2*"))
}
