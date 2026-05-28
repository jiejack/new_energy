package rule

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuleEngine_StartStop(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	require.False(t, engine.IsRunning())

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

func TestRuleEngine_LoadRule(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{
		ID:       "rule-1",
		Name:     "Test",
		Type:     RuleTypeFormula,
		Formula:  "A + B",
		Enabled:  true,
		PointID:  "point-1",
	}

	err := engine.LoadRule(rule)
	require.NoError(t, err)
	assert.Equal(t, RuleStatusActive, rule.Status)
	assert.Equal(t, 1, rule.Version)
	assert.False(t, rule.CreateTime.IsZero())
}

func TestRuleEngine_LoadRule_NoID(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{Type: RuleTypeFormula, Formula: "A+B"}
	err := engine.LoadRule(rule)
	assert.Equal(t, ErrInvalidRule, err)
}

func TestRuleEngine_LoadRule_Duplicate(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{ID: "rule-1", Type: RuleTypeFormula, Formula: "A+B"}
	engine.LoadRule(rule)

	err := engine.LoadRule(rule)
	assert.Equal(t, ErrRuleExists, err)
}

func TestRuleEngine_LoadRule_InvalidType(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{ID: "rule-1", Type: ""}
	err := engine.LoadRule(rule)
	assert.Error(t, err)
}

func TestRuleEngine_LoadRule_FormulaNoFormula(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{ID: "rule-1", Type: RuleTypeFormula}
	err := engine.LoadRule(rule)
	assert.Error(t, err)
}

func TestRuleEngine_LoadRule_ExpressionNoExpression(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{ID: "rule-1", Type: RuleTypeExpression}
	err := engine.LoadRule(rule)
	assert.Error(t, err)
}

func TestRuleEngine_LoadRule_ScriptNoScript(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{ID: "rule-1", Type: RuleTypeScript}
	err := engine.LoadRule(rule)
	assert.Error(t, err)
}

func TestRuleEngine_UnloadRule(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{ID: "rule-1", Type: RuleTypeFormula, Formula: "A+B", PointID: "p1"}
	engine.LoadRule(rule)

	err := engine.UnloadRule("rule-1")
	require.NoError(t, err)
	assert.Equal(t, 0, engine.GetRuleCount())
}

func TestRuleEngine_UnloadRule_NotFound(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	err := engine.UnloadRule("nonexistent")
	assert.Equal(t, ErrRuleNotFound, err)
}

func TestRuleEngine_UpdateRule(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{ID: "rule-1", Type: RuleTypeFormula, Formula: "A+B", Enabled: true}
	engine.LoadRule(rule)

	updated := &Rule{ID: "rule-1", Type: RuleTypeFormula, Formula: "C+D", Enabled: true}
	err := engine.UpdateRule(updated)
	require.NoError(t, err)
	assert.Equal(t, 2, updated.Version)
}

func TestRuleEngine_UpdateRule_NotFound(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{ID: "nonexistent", Type: RuleTypeFormula, Formula: "A+B"}
	err := engine.UpdateRule(rule)
	assert.Equal(t, ErrRuleNotFound, err)
}

func TestRuleEngine_UpdateRule_Invalid(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{ID: "rule-1", Type: RuleTypeFormula, Formula: "A+B"}
	engine.LoadRule(rule)

	updated := &Rule{ID: "rule-1", Type: RuleTypeFormula, Formula: ""}
	err := engine.UpdateRule(updated)
	assert.Error(t, err)
}

func TestRuleEngine_GetRule(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{ID: "rule-1", Type: RuleTypeFormula, Formula: "A+B"}
	engine.LoadRule(rule)

	found, err := engine.GetRule("rule-1")
	require.NoError(t, err)
	assert.Equal(t, "rule-1", found.ID)
}

func TestRuleEngine_GetRule_NotFound(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	_, err := engine.GetRule("nonexistent")
	assert.Equal(t, ErrRuleNotFound, err)
}

func TestRuleEngine_GetRulesByPoint(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	engine.LoadRule(&Rule{ID: "r1", Type: RuleTypeFormula, Formula: "A", PointID: "p1", Priority: 1, Enabled: true})
	engine.LoadRule(&Rule{ID: "r2", Type: RuleTypeFormula, Formula: "B", PointID: "p1", Priority: 5, Enabled: true})

	rules := engine.GetRulesByPoint("p1")
	assert.Equal(t, 2, len(rules))
	assert.Equal(t, "r2", rules[0].ID)
}

func TestRuleEngine_GetRulesByType(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	engine.LoadRule(&Rule{ID: "r1", Type: RuleTypeFormula, Formula: "A"})
	engine.LoadRule(&Rule{ID: "r2", Type: RuleTypeExpression, Expression: "x > 0"})

	rules := engine.GetRulesByType(RuleTypeFormula)
	assert.Equal(t, 1, len(rules))
}

func TestRuleEngine_EnableDisableRule(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{ID: "rule-1", Type: RuleTypeFormula, Formula: "A+B", Enabled: true}
	engine.LoadRule(rule)

	err := engine.DisableRule("rule-1")
	require.NoError(t, err)
	found, _ := engine.GetRule("rule-1")
	assert.False(t, found.Enabled)
	assert.Equal(t, RuleStatusDisabled, found.Status)

	err = engine.EnableRule("rule-1")
	require.NoError(t, err)
	found, _ = engine.GetRule("rule-1")
	assert.True(t, found.Enabled)
	assert.Equal(t, RuleStatusActive, found.Status)
}

func TestRuleEngine_EnableRule_NotFound(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	err := engine.EnableRule("nonexistent")
	assert.Equal(t, ErrRuleNotFound, err)
}

func TestRuleEngine_DisableRule_NotFound(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	err := engine.DisableRule("nonexistent")
	assert.Equal(t, ErrRuleNotFound, err)
}

func TestRuleEngine_Execute_Disabled(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{ID: "rule-1", Type: RuleTypeFormula, Formula: "A+B", Enabled: false}
	engine.LoadRule(rule)

	_, err := engine.Execute(context.Background(), "rule-1")
	assert.Equal(t, ErrRuleDisabled, err)
}

func TestRuleEngine_Execute_NotFound(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	_, err := engine.Execute(context.Background(), "nonexistent")
	assert.Equal(t, ErrRuleNotFound, err)
}

func TestRuleEngine_Execute_Formula(t *testing.T) {
	provider := &mockDataProvider{}
	cache := NewComputeCache(&CacheConfig{EnableLocalCache: true, LocalCacheSize: 100, DefaultTTL: 5 * time.Minute}, nil)
	engine := NewRuleEngine(cache, provider)

	rule := &Rule{
		ID:       "formula-rule",
		Type:     RuleTypeFormula,
		Formula:  "input1",
		Enabled:  true,
		Inputs:   []RuleInput{{Name: "input1", PointID: "point-a", Required: true}},
		Timeout:  5 * time.Second,
	}
	engine.LoadRule(rule)

	exec, err := engine.Execute(context.Background(), "formula-rule")
	require.NoError(t, err)
	assert.Equal(t, "success", exec.Status)
	assert.Equal(t, 50.0, exec.Outputs["value"])
}

func TestRuleEngine_Execute_FormulaWithDefault(t *testing.T) {
	provider := &mockDataProvider{}
	engine := NewRuleEngine(nil, provider)

	rule := &Rule{
		ID:      "default-rule",
		Type:    RuleTypeFormula,
		Formula: "input1",
		Enabled: true,
		Inputs:  []RuleInput{{Name: "input1", Default: 42.0, Required: false}},
		Timeout: 5 * time.Second,
	}
	engine.LoadRule(rule)

	exec, err := engine.Execute(context.Background(), "default-rule")
	require.NoError(t, err)
	assert.Equal(t, "success", exec.Status)
	assert.Equal(t, 42.0, exec.Outputs["value"])
}

func TestRuleEngine_Execute_Expression(t *testing.T) {
	provider := &mockDataProvider{}
	engine := NewRuleEngine(nil, provider)

	rule := &Rule{
		ID:         "expr-rule",
		Type:       RuleTypeExpression,
		Expression: "value > 100",
		Enabled:    true,
		Inputs:     []RuleInput{{Name: "value", PointID: "p1", Required: true}},
		Timeout:    5 * time.Second,
	}
	engine.LoadRule(rule)

	exec, err := engine.Execute(context.Background(), "expr-rule")
	require.NoError(t, err)
	assert.Equal(t, "success", exec.Status)
}

func TestRuleEngine_Execute_Script(t *testing.T) {
	provider := &mockDataProvider{}
	engine := NewRuleEngine(nil, provider)

	rule := &Rule{
		ID:      "script-rule",
		Type:    RuleTypeScript,
		Script:  "return sum(inputs)",
		Enabled: true,
		Inputs: []RuleInput{
			{Name: "a", PointID: "p1", Required: true},
			{Name: "b", PointID: "p2", Required: true},
		},
		Timeout: 5 * time.Second,
	}
	engine.LoadRule(rule)

	exec, err := engine.Execute(context.Background(), "script-rule")
	require.NoError(t, err)
	assert.Equal(t, "success", exec.Status)
}

func TestRuleEngine_Execute_Aggregate(t *testing.T) {
	provider := &mockDataProvider{}
	engine := NewRuleEngine(nil, provider)

	rule := &Rule{
		ID:      "agg-rule",
		Type:    RuleTypeAggregate,
		Enabled: true,
		Inputs:  []RuleInput{{Name: "input1", PointID: "p1", Required: true}},
		Config:  map[string]interface{}{"aggregateFunc": "avg", "window": "5m"},
		Timeout: 5 * time.Second,
	}
	engine.LoadRule(rule)

	exec, err := engine.Execute(context.Background(), "agg-rule")
	require.NoError(t, err)
	assert.Equal(t, "success", exec.Status)
}

func TestRuleEngine_Execute_Aggregate_NoProvider(t *testing.T) {
	engine := NewRuleEngine(nil, nil)

	rule := &Rule{
		ID:      "agg-rule",
		Type:    RuleTypeAggregate,
		Enabled: true,
		Inputs:  []RuleInput{{Name: "input1", PointID: "p1", Required: true}},
		Timeout: 5 * time.Second,
	}
	engine.LoadRule(rule)

	exec, err := engine.Execute(context.Background(), "agg-rule")
	require.NoError(t, err)
	assert.Equal(t, "error", exec.Status)
}

func TestRuleEngine_Execute_Aggregate_NoInputPoint(t *testing.T) {
	provider := &mockDataProvider{}
	engine := NewRuleEngine(nil, provider)

	rule := &Rule{
		ID:      "agg-rule",
		Type:    RuleTypeAggregate,
		Enabled: true,
		Inputs:  []RuleInput{{Name: "input1", Required: true}},
		Timeout: 5 * time.Second,
	}
	engine.LoadRule(rule)

	exec, err := engine.Execute(context.Background(), "agg-rule")
	require.NoError(t, err)
	assert.Equal(t, "error", exec.Status)
}

func TestRuleEngine_Execute_Transform(t *testing.T) {
	provider := &mockDataProvider{}
	engine := NewRuleEngine(nil, provider)

	rule := &Rule{
		ID:      "transform-rule",
		Type:    RuleTypeTransform,
		Enabled: true,
		Inputs:  []RuleInput{{Name: "input1", PointID: "p1", Required: true}},
		Config:  map[string]interface{}{"scale": 2.0, "offset": 10.0},
		Timeout: 5 * time.Second,
	}
	engine.LoadRule(rule)

	exec, err := engine.Execute(context.Background(), "transform-rule")
	require.NoError(t, err)
	assert.Equal(t, "success", exec.Status)
	assert.Equal(t, 110.0, exec.Outputs["value"])
}

func TestRuleEngine_Execute_UnknownType(t *testing.T) {
	engine := NewRuleEngine(nil, &mockDataProvider{})
	rule := &Rule{ID: "unknown-rule", Type: RuleType("unknown"), Enabled: true, Timeout: 5 * time.Second}
	engine.LoadRule(rule)

	exec, err := engine.Execute(context.Background(), "unknown-rule")
	require.NoError(t, err)
	assert.Equal(t, "error", exec.Status)
}

func TestRuleEngine_Execute_RequiredInputFail(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	rule := &Rule{
		ID:      "req-rule",
		Type:    RuleTypeFormula,
		Formula: "input1",
		Enabled: true,
		Inputs:  []RuleInput{{Name: "input1", PointID: "p1", Required: true}},
		Timeout: 5 * time.Second,
	}
	engine.LoadRule(rule)

	exec, err := engine.Execute(context.Background(), "req-rule")
	require.NoError(t, err)
	assert.Equal(t, "error", exec.Status)
}

func TestRuleEngine_ExecuteForPoint(t *testing.T) {
	provider := &mockDataProvider{}
	engine := NewRuleEngine(nil, provider)
	engine.LoadRule(&Rule{ID: "r1", Type: RuleTypeFormula, Formula: "A", PointID: "p1", Enabled: true, Timeout: 5 * time.Second})

	execs, err := engine.ExecuteForPoint(context.Background(), "p1")
	require.NoError(t, err)
	assert.Equal(t, 1, len(execs))

	execs, err = engine.ExecuteForPoint(context.Background(), "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, execs)
}

func TestRuleEngine_ExecuteAll(t *testing.T) {
	provider := &mockDataProvider{}
	engine := NewRuleEngine(nil, provider)
	engine.LoadRule(&Rule{ID: "r1", Type: RuleTypeFormula, Formula: "A", Enabled: true, Timeout: 5 * time.Second})
	engine.LoadRule(&Rule{ID: "r2", Type: RuleTypeFormula, Formula: "B", Enabled: true, Timeout: 5 * time.Second})
	engine.LoadRule(&Rule{ID: "r3", Type: RuleTypeFormula, Formula: "C", Enabled: false, Timeout: 5 * time.Second})

	results, err := engine.ExecuteAll(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 2, len(results))
}

func TestRuleEngine_BatchExecute(t *testing.T) {
	provider := &mockDataProvider{}
	engine := NewRuleEngine(nil, provider)
	engine.LoadRule(&Rule{ID: "r1", Type: RuleTypeFormula, Formula: "A", Enabled: true, Timeout: 5 * time.Second})

	results, err := engine.BatchExecute(context.Background(), []string{"r1", "nonexistent"})
	require.NoError(t, err)
	assert.Equal(t, 2, len(results))
}

func TestRuleEngine_ClearCache(t *testing.T) {
	cache := NewComputeCache(&CacheConfig{EnableLocalCache: true, LocalCacheSize: 100, DefaultTTL: 5 * time.Minute}, nil)
	engine := NewRuleEngine(cache, nil)

	err := engine.ClearCache(context.Background())
	assert.NoError(t, err)

	engine2 := NewRuleEngine(nil, nil)
	err = engine2.ClearCache(context.Background())
	assert.NoError(t, err)
}

func TestRuleEngine_ReloadRules(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	engine.LoadRule(&Rule{ID: "r1", Type: RuleTypeFormula, Formula: "A"})

	newRules := []*Rule{
		{ID: "r2", Type: RuleTypeExpression, Expression: "x > 0", Enabled: true},
		{ID: "r3", Type: RuleTypeFormula, Formula: "B", Enabled: false},
	}
	err := engine.ReloadRules(newRules)
	require.NoError(t, err)
	assert.Equal(t, 2, engine.GetRuleCount())
}

func TestRuleEngine_ReloadRules_InvalidRule(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	newRules := []*Rule{
		{ID: "r1", Type: RuleTypeFormula, Formula: "A"},
		{ID: "r2", Type: RuleTypeFormula, Formula: ""},
	}
	err := engine.ReloadRules(newRules)
	require.NoError(t, err)
	assert.Equal(t, 1, engine.GetRuleCount())
}

func TestRuleEngine_ExportRules(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	engine.LoadRule(&Rule{ID: "r1", Type: RuleTypeFormula, Formula: "A"})
	engine.LoadRule(&Rule{ID: "r2", Type: RuleTypeExpression, Expression: "x > 0"})

	rules := engine.ExportRules()
	assert.Equal(t, 2, len(rules))
}

func TestRuleEngine_GetStats(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	engine.LoadRule(&Rule{ID: "r1", Type: RuleTypeFormula, Formula: "A", Enabled: true})
	engine.LoadRule(&Rule{ID: "r2", Type: RuleTypeFormula, Formula: "B", Enabled: false})

	stats := engine.GetStats()
	assert.Equal(t, int64(2), stats.TotalRules)
	assert.Equal(t, int64(1), stats.ActiveRules)
}

func TestRuleEngine_GetRuleCount(t *testing.T) {
	engine := NewRuleEngine(nil, nil)
	assert.Equal(t, 0, engine.GetRuleCount())
	engine.LoadRule(&Rule{ID: "r1", Type: RuleTypeFormula, Formula: "A"})
	assert.Equal(t, 1, engine.GetRuleCount())
}

func TestRuleEngine_OutputTransform(t *testing.T) {
	engine := NewRuleEngine(nil, &mockDataProvider{})
	rule := &Rule{
		ID:      "transform-out",
		Type:    RuleTypeFormula,
		Formula: "input1",
		Enabled: true,
		Inputs:  []RuleInput{{Name: "input1", PointID: "p1", Required: true}},
		Outputs: []RuleOutput{{Scale: 2.0, Offset: 5.0}},
		Timeout: 5 * time.Second,
	}
	engine.LoadRule(rule)

	exec, err := engine.Execute(context.Background(), "transform-out")
	require.NoError(t, err)
	assert.Equal(t, "success", exec.Status)
	assert.Equal(t, 105.0, exec.Outputs["value"])
}

func TestRuleEngine_Conditions(t *testing.T) {
	engine := NewRuleEngine(nil, &mockDataProvider{})
	rule := &Rule{
		ID:         "cond-rule",
		Type:       RuleTypeFormula,
		Formula:    "input1",
		Enabled:    true,
		Inputs:     []RuleInput{{Name: "input1", PointID: "p1", Required: true}},
		Conditions: []RuleCondition{{Expression: "value > 0", Type: "pre"}, {Expression: "value > 0", Type: "post"}},
		Timeout:    5 * time.Second,
	}
	engine.LoadRule(rule)

	exec, err := engine.Execute(context.Background(), "cond-rule")
	require.NoError(t, err)
	assert.Equal(t, "success", exec.Status)
}

func TestComputeCache_SetGet(t *testing.T) {
	cache := NewComputeCache(&CacheConfig{EnableLocalCache: true, LocalCacheSize: 100, DefaultTTL: 5 * time.Minute}, nil)
	ctx := context.Background()

	result := &ComputeResult{PointID: "p1", Value: 42.0, Timestamp: time.Now()}
	err := cache.Set(ctx, "key1", result)
	require.NoError(t, err)

	got, err := cache.Get(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, 42.0, got.Value)
}

func TestComputeCache_Get_NotFound(t *testing.T) {
	cache := NewComputeCache(&CacheConfig{EnableLocalCache: true, LocalCacheSize: 100, DefaultTTL: 5 * time.Minute}, nil)
	ctx := context.Background()

	_, err := cache.Get(ctx, "nonexistent")
	assert.Equal(t, ErrCacheNotFound, err)
}

func TestComputeCache_Delete(t *testing.T) {
	cache := NewComputeCache(&CacheConfig{EnableLocalCache: true, LocalCacheSize: 100, DefaultTTL: 5 * time.Minute}, nil)
	ctx := context.Background()

	cache.Set(ctx, "key1", &ComputeResult{Value: 1.0})
	err := cache.Delete(ctx, "key1")
	require.NoError(t, err)

	_, err = cache.Get(ctx, "key1")
	assert.Equal(t, ErrCacheNotFound, err)
}

func TestComputeCache_Clear(t *testing.T) {
	cache := NewComputeCache(&CacheConfig{EnableLocalCache: true, LocalCacheSize: 100, DefaultTTL: 5 * time.Minute}, nil)
	ctx := context.Background()

	cache.Set(ctx, "key1", &ComputeResult{Value: 1.0})
	cache.Set(ctx, "key2", &ComputeResult{Value: 2.0})

	err := cache.Clear(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, cache.GetSize())
}

func TestComputeCache_Invalidate(t *testing.T) {
	cache := NewComputeCache(&CacheConfig{EnableLocalCache: true, LocalCacheSize: 100, DefaultTTL: 5 * time.Minute}, nil)
	ctx := context.Background()

	cache.Set(ctx, "rule:p1", &ComputeResult{Value: 1.0})
	cache.Set(ctx, "rule:p2", &ComputeResult{Value: 2.0})

	err := cache.Invalidate(ctx, "*p1*")
	require.NoError(t, err)
}

func TestComputeCache_InvalidateByPoint(t *testing.T) {
	cache := NewComputeCache(&CacheConfig{EnableLocalCache: true, LocalCacheSize: 100, DefaultTTL: 5 * time.Minute}, nil)
	ctx := context.Background()

	cache.Set(ctx, "rule:p1", &ComputeResult{Value: 1.0})
	err := cache.InvalidateByPoint(ctx, "p1")
	require.NoError(t, err)
}

func TestComputeCache_InvalidateByDependency(t *testing.T) {
	cache := NewComputeCache(&CacheConfig{EnableLocalCache: true, LocalCacheSize: 100, DefaultTTL: 5 * time.Minute}, nil)
	ctx := context.Background()

	cache.Set(ctx, "rule:p1", &ComputeResult{Value: 1.0})
	err := cache.InvalidateByDependency(ctx, []string{"p1"})
	require.NoError(t, err)
}

func TestComputeCache_GetOrCompute(t *testing.T) {
	cache := NewComputeCache(&CacheConfig{EnableLocalCache: true, LocalCacheSize: 100, DefaultTTL: 5 * time.Minute}, nil)
	ctx := context.Background()

	computed := false
	result, err := cache.GetOrCompute(ctx, "key1", func() (*ComputeResult, error) {
		computed = true
		return &ComputeResult{Value: 99.0}, nil
	})
	require.NoError(t, err)
	assert.True(t, computed)
	assert.Equal(t, 99.0, result.Value)

	computed = false
	result, err = cache.GetOrCompute(ctx, "key1", func() (*ComputeResult, error) {
		computed = true
		return &ComputeResult{Value: 88.0}, nil
	})
	require.NoError(t, err)
	assert.False(t, computed)
	assert.Equal(t, 99.0, result.Value)
}

func TestComputeCache_GetOrCompute_Error(t *testing.T) {
	cache := NewComputeCache(&CacheConfig{EnableLocalCache: true, LocalCacheSize: 100, DefaultTTL: 5 * time.Minute}, nil)
	ctx := context.Background()

	_, err := cache.GetOrCompute(ctx, "key1", func() (*ComputeResult, error) {
		return nil, assert.AnError
	})
	assert.Error(t, err)
}

func TestComputeCache_BatchGet(t *testing.T) {
	cache := NewComputeCache(&CacheConfig{EnableLocalCache: true, LocalCacheSize: 100, DefaultTTL: 5 * time.Minute}, nil)
	ctx := context.Background()

	cache.Set(ctx, "key1", &ComputeResult{Value: 1.0})
	cache.Set(ctx, "key2", &ComputeResult{Value: 2.0})

	results, err := cache.BatchGet(ctx, []string{"key1", "key2", "key3"})
	require.NoError(t, err)
	assert.Equal(t, 2, len(results))
}

func TestComputeCache_BatchSet(t *testing.T) {
	cache := NewComputeCache(&CacheConfig{EnableLocalCache: true, LocalCacheSize: 100, DefaultTTL: 5 * time.Minute}, nil)
	ctx := context.Background()

	items := map[string]*ComputeResult{
		"key1": {Value: 1.0},
		"key2": {Value: 2.0},
	}
	err := cache.BatchSet(ctx, items)
	require.NoError(t, err)
}

func TestComputeCache_GetMetrics(t *testing.T) {
	cache := NewComputeCache(&CacheConfig{EnableLocalCache: true, LocalCacheSize: 100, DefaultTTL: 5 * time.Minute}, nil)
	ctx := context.Background()

	cache.Set(ctx, "key1", &ComputeResult{Value: 1.0})
	cache.Get(ctx, "key1")
	cache.Get(ctx, "nonexistent")

	metrics := cache.GetMetrics()
	assert.Equal(t, int64(1), metrics.Hits)
	assert.Equal(t, int64(1), metrics.Misses)
}

func TestComputeCache_IsEnabled(t *testing.T) {
	cache := NewComputeCache(&CacheConfig{EnableLocalCache: true, LocalCacheSize: 100, DefaultTTL: 5 * time.Minute}, nil)
	assert.True(t, cache.IsEnabled())

	cache2 := NewComputeCache(&CacheConfig{EnableLocalCache: false, EnableRedisCache: false}, nil)
	assert.False(t, cache2.IsEnabled())
}

func TestComputeCache_DefaultConfig(t *testing.T) {
	cache := NewComputeCache(nil, nil)
	assert.NotNil(t, cache)
	assert.True(t, cache.IsEnabled())
}

func TestComputeCache_SetWithTTL(t *testing.T) {
	cache := NewComputeCache(&CacheConfig{EnableLocalCache: true, LocalCacheSize: 100, DefaultTTL: 5 * time.Minute}, nil)
	ctx := context.Background()

	err := cache.SetWithTTL(ctx, "key1", &ComputeResult{Value: 1.0}, 10*time.Minute)
	require.NoError(t, err)
}

func TestLocalCache_Eviction(t *testing.T) {
	cache := NewLocalCache(3, 5*time.Minute, CachePolicyLRU)

	for i := 0; i < 5; i++ {
		cache.Set(fmt.Sprintf("key%d", i), &ComputeResult{Value: float64(i)})
	}

	assert.LessOrEqual(t, cache.size, 3)
}

func TestLocalCache_LFUPolicy(t *testing.T) {
	cache := NewLocalCache(3, 5*time.Minute, CachePolicyLFU)

	cache.Set("key1", &ComputeResult{Value: 1.0})
	cache.Set("key2", &ComputeResult{Value: 2.0})
	cache.Set("key3", &ComputeResult{Value: 3.0})

	cache.Get("key1")
	cache.Get("key1")
	cache.Get("key2")

	cache.Set("key4", &ComputeResult{Value: 4.0})
	assert.LessOrEqual(t, cache.size, 3)
}

func TestLocalCache_FIFOPolicy(t *testing.T) {
	cache := NewLocalCache(3, 5*time.Minute, CachePolicyFIFO)

	cache.Set("key1", &ComputeResult{Value: 1.0})
	cache.Set("key2", &ComputeResult{Value: 2.0})
	cache.Set("key3", &ComputeResult{Value: 3.0})
	cache.Set("key4", &ComputeResult{Value: 4.0})

	assert.LessOrEqual(t, cache.size, 3)
}

func TestLocalCache_Invalidate(t *testing.T) {
	cache := NewLocalCache(100, 5*time.Minute, CachePolicyLRU)
	cache.Set("abc-test-123", &ComputeResult{Value: 1.0})
	cache.Set("abc-other-456", &ComputeResult{Value: 2.0})
	cache.Set("xyz-test-789", &ComputeResult{Value: 3.0})

	cache.Invalidate("*test*")
}

func TestLocalCache_ExpiredItem(t *testing.T) {
	cache := NewLocalCache(100, 1*time.Millisecond, CachePolicyLRU)
	cache.Set("key1", &ComputeResult{Value: 1.0})

	time.Sleep(5 * time.Millisecond)
	_, exists := cache.Get("key1")
	assert.False(t, exists)
}

func TestLocalCache_Clear(t *testing.T) {
	cache := NewLocalCache(100, 5*time.Minute, CachePolicyLRU)
	cache.Set("key1", &ComputeResult{Value: 1.0})
	cache.Set("key2", &ComputeResult{Value: 2.0})

	cache.Clear()
	assert.Equal(t, 0, cache.size)
}

func TestComputeScheduler_AddTask(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	err := scheduler.AddTask(&ComputeTask{
		ID:       "task-1",
		Name:     "Test Task",
		Type:     TaskTypeInterval,
		Interval: 1 * time.Minute,
		Enabled:  true,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, scheduler.GetTaskCount())
}

func TestComputeScheduler_AddTask_Invalid(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	err := scheduler.AddTask(nil)
	assert.Equal(t, ErrInvalidTask, err)

	err = scheduler.AddTask(&ComputeTask{ID: ""})
	assert.Equal(t, ErrInvalidTask, err)
}

func TestComputeScheduler_AddTask_Duplicate(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	scheduler.AddTask(&ComputeTask{ID: "task-1", Type: TaskTypeInterval, Interval: time.Minute, Enabled: true})
	err := scheduler.AddTask(&ComputeTask{ID: "task-1", Type: TaskTypeInterval, Interval: time.Minute, Enabled: true})
	assert.Equal(t, ErrTaskExists, err)
}

func TestComputeScheduler_RemoveTask(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	scheduler.AddTask(&ComputeTask{ID: "task-1", Type: TaskTypeInterval, Interval: time.Minute, Enabled: true})
	err := scheduler.RemoveTask("task-1")
	require.NoError(t, err)
	assert.Equal(t, 0, scheduler.GetTaskCount())
}

func TestComputeScheduler_RemoveTask_NotFound(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	err := scheduler.RemoveTask("nonexistent")
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestComputeScheduler_GetTask(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	scheduler.AddTask(&ComputeTask{ID: "task-1", Name: "Test", Type: TaskTypeInterval, Interval: time.Minute, Enabled: true})

	task, err := scheduler.GetTask("task-1")
	require.NoError(t, err)
	assert.Equal(t, "Test", task.Name)

	_, err = scheduler.GetTask("nonexistent")
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestComputeScheduler_GetAllTasks(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	scheduler.AddTask(&ComputeTask{ID: "task-1", Type: TaskTypeInterval, Interval: time.Minute, Enabled: true})
	scheduler.AddTask(&ComputeTask{ID: "task-2", Type: TaskTypeCron, CronExpr: "@hourly", Enabled: true})

	tasks := scheduler.GetAllTasks()
	assert.Equal(t, 2, len(tasks))
}

func TestComputeScheduler_UpdateTask(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	scheduler.AddTask(&ComputeTask{ID: "task-1", Name: "Original", Type: TaskTypeInterval, Interval: time.Minute, Enabled: true})

	err := scheduler.UpdateTask(&ComputeTask{ID: "task-1", Name: "Updated", Type: TaskTypeInterval, Interval: 2 * time.Minute, Enabled: true})
	require.NoError(t, err)

	task, _ := scheduler.GetTask("task-1")
	assert.Equal(t, "Updated", task.Name)
}

func TestComputeScheduler_UpdateTask_NotFound(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	err := scheduler.UpdateTask(&ComputeTask{ID: "nonexistent", Type: TaskTypeInterval, Interval: time.Minute})
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestComputeScheduler_PauseResumeTask(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	scheduler.AddTask(&ComputeTask{ID: "task-1", Type: TaskTypeInterval, Interval: time.Minute, Enabled: true})

	err := scheduler.PauseTask("task-1")
	require.NoError(t, err)
	task, _ := scheduler.GetTask("task-1")
	assert.Equal(t, TaskStatusPaused, task.Status)
	assert.False(t, task.Enabled)

	err = scheduler.ResumeTask("task-1")
	require.NoError(t, err)
	task, _ = scheduler.GetTask("task-1")
	assert.Equal(t, TaskStatusPending, task.Status)
	assert.True(t, task.Enabled)
}

func TestComputeScheduler_PauseTask_NotFound(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	err := scheduler.PauseTask("nonexistent")
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestComputeScheduler_ResumeTask_NotFound(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	err := scheduler.ResumeTask("nonexistent")
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestComputeScheduler_TriggerTask(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)
	scheduler.AddTask(&ComputeTask{ID: "task-1", Type: TaskTypeInterval, Interval: time.Minute, Enabled: true, Timeout: 5 * time.Second})

	err := scheduler.TriggerTask("task-1")
	require.NoError(t, err)
}

func TestComputeScheduler_TriggerTask_NotFound(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	err := scheduler.TriggerTask("nonexistent")
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestComputeScheduler_StartStop(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	err := scheduler.Start()
	require.NoError(t, err)
	assert.True(t, scheduler.IsRunning())

	err = scheduler.Start()
	assert.Equal(t, ErrSchedulerRunning, err)

	err = scheduler.Stop()
	require.NoError(t, err)
	assert.False(t, scheduler.IsRunning())

	err = scheduler.Stop()
	assert.Equal(t, ErrSchedulerNotRunning, err)
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
	assert.Equal(t, 0, len(logs))
}

func TestComputeScheduler_CalculateNextRunTime(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	task := &ComputeTask{Type: TaskTypeInterval, Interval: 5 * time.Minute}
	next := scheduler.calculateNextRunTime(task)
	assert.True(t, next.After(time.Now()))

	task2 := &ComputeTask{Type: TaskTypeCron, CronExpr: "@every 1m"}
	next2 := scheduler.calculateNextRunTime(task2)
	assert.True(t, next2.After(time.Now()))

	task3 := &ComputeTask{Type: TaskTypeOnce, NextRunTime: time.Now().Add(time.Hour)}
	next3 := scheduler.calculateNextRunTime(task3)
	assert.Equal(t, task3.NextRunTime, next3)

	task4 := &ComputeTask{Type: TaskTypeOnce, NextRunTime: time.Time{}}
	next4 := scheduler.calculateNextRunTime(task4)
	assert.False(t, next4.IsZero())
}

func TestComputeScheduler_ParseCronNextTime(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)
	now := time.Now()

	tests := []struct {
		expr     string
		expected time.Duration
	}{
		{"@every 1m", time.Minute},
		{"@every 5m", 5 * time.Minute},
		{"@every 15m", 15 * time.Minute},
		{"@every 30m", 30 * time.Minute},
		{"@hourly", time.Hour},
		{"@daily", 24 * time.Hour},
		{"@weekly", 7 * 24 * time.Hour},
		{"unknown", time.Minute},
	}

	for _, tt := range tests {
		result := scheduler.parseCronNextTime(tt.expr, now)
		assert.True(t, result.Sub(now) >= tt.expected-time.Second)
	}
}

func TestComputeScheduler_DisabledTask(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	err := scheduler.AddTask(&ComputeTask{ID: "task-1", Type: TaskTypeInterval, Interval: time.Minute, Enabled: false})
	require.NoError(t, err)

	task, _ := scheduler.GetTask("task-1")
	assert.Equal(t, TaskStatusPaused, task.Status)
}

func TestTriggerManager_CreateTrigger(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	trigger := &Trigger{
		ID:       "trigger-1",
		Name:     "Test",
		Type:     TriggerTypeDataChange,
		Enabled:  true,
		PointIDs: []string{"p1"},
	}
	err := tm.CreateTrigger(trigger)
	require.NoError(t, err)
	assert.Equal(t, TriggerStatusActive, trigger.Status)
}

func TestTriggerManager_CreateTrigger_NoID(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	err := tm.CreateTrigger(&Trigger{Type: TriggerTypeDataChange})
	assert.Equal(t, ErrInvalidTrigger, err)
}

func TestTriggerManager_CreateTrigger_Duplicate(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	tm.CreateTrigger(&Trigger{ID: "t1", Type: TriggerTypeDataChange})
	err := tm.CreateTrigger(&Trigger{ID: "t1", Type: TriggerTypeDataChange})
	assert.Equal(t, ErrTriggerExists, err)
}

func TestTriggerManager_UpdateTrigger(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	tm.CreateTrigger(&Trigger{ID: "t1", Name: "Original", Type: TriggerTypeDataChange, PointIDs: []string{"p1"}})
	err := tm.UpdateTrigger(&Trigger{ID: "t1", Name: "Updated", Type: TriggerTypeEvent, PointIDs: []string{"p2"}})
	require.NoError(t, err)
}

func TestTriggerManager_UpdateTrigger_NotFound(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	err := tm.UpdateTrigger(&Trigger{ID: "nonexistent", Type: TriggerTypeDataChange})
	assert.Equal(t, ErrTriggerNotFound, err)
}

func TestTriggerManager_DeleteTrigger(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	tm.CreateTrigger(&Trigger{ID: "t1", Type: TriggerTypeDataChange, PointIDs: []string{"p1"}})
	err := tm.DeleteTrigger("t1")
	require.NoError(t, err)

	_, err = tm.GetTrigger("t1")
	assert.Equal(t, ErrTriggerNotFound, err)
}

func TestTriggerManager_DeleteTrigger_NotFound(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	err := tm.DeleteTrigger("nonexistent")
	assert.Equal(t, ErrTriggerNotFound, err)
}

func TestTriggerManager_GetTriggersByType(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	tm.CreateTrigger(&Trigger{ID: "t1", Type: TriggerTypeDataChange})
	tm.CreateTrigger(&Trigger{ID: "t2", Type: TriggerTypeEvent})

	triggers := tm.GetTriggersByType(TriggerTypeDataChange)
	assert.Equal(t, 1, len(triggers))
}

func TestTriggerManager_EnableDisableTrigger(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	tm.CreateTrigger(&Trigger{ID: "t1", Type: TriggerTypeDataChange, Enabled: true})

	err := tm.DisableTrigger("t1")
	require.NoError(t, err)
	trigger, _ := tm.GetTrigger("t1")
	assert.False(t, trigger.Enabled)
	assert.Equal(t, TriggerStatusDisabled, trigger.Status)

	err = tm.EnableTrigger("t1")
	require.NoError(t, err)
	trigger, _ = tm.GetTrigger("t1")
	assert.True(t, trigger.Enabled)
	assert.Equal(t, TriggerStatusActive, trigger.Status)
}

func TestTriggerManager_EnableTrigger_NotFound(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	err := tm.EnableTrigger("nonexistent")
	assert.Equal(t, ErrTriggerNotFound, err)
}

func TestTriggerManager_DisableTrigger_NotFound(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	err := tm.DisableTrigger("nonexistent")
	assert.Equal(t, ErrTriggerNotFound, err)
}

func TestTriggerManager_OnDataChange(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)
	tm.Start()
	defer tm.Stop()

	err := tm.OnDataChange(context.Background(), "p1", 10.0, 20.0)
	require.NoError(t, err)
}

func TestTriggerManager_OnEvent(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)
	tm.Start()
	defer tm.Stop()

	err := tm.OnEvent(context.Background(), "alarm", map[string]interface{}{"data": "test"})
	require.NoError(t, err)
}

func TestTriggerManager_OnCondition(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)
	tm.Start()
	defer tm.Stop()

	err := tm.OnCondition(context.Background(), "p1", "value > 100", 150.0)
	require.NoError(t, err)
}

func TestTriggerManager_CreateChain(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	err := tm.CreateChain(&TriggerChain{ID: "chain-1", Name: "Test Chain"})
	require.NoError(t, err)

	chain, err := tm.GetChain("chain-1")
	require.NoError(t, err)
	assert.Equal(t, "Test Chain", chain.Name)
}

func TestTriggerManager_CreateChain_NoID(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	err := tm.CreateChain(&TriggerChain{Name: "No ID"})
	assert.Error(t, err)
}

func TestTriggerManager_CreateChain_Duplicate(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	tm.CreateChain(&TriggerChain{ID: "chain-1"})
	err := tm.CreateChain(&TriggerChain{ID: "chain-1"})
	assert.Error(t, err)
}

func TestTriggerManager_GetChain_NotFound(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	_, err := tm.GetChain("nonexistent")
	assert.Error(t, err)
}

func TestTriggerManager_GetStats(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	tm.CreateTrigger(&Trigger{ID: "t1", Type: TriggerTypeDataChange, Enabled: true})
	tm.CreateTrigger(&Trigger{ID: "t2", Type: TriggerTypeEvent, Enabled: false})

	stats := tm.GetStats()
	assert.Equal(t, 2, stats["totalTriggers"])
}

func TestTriggerManager_StartStop(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	err := tm.Start()
	require.NoError(t, err)

	err = tm.Start()
	assert.Error(t, err)

	err = tm.Stop()
	require.NoError(t, err)

	err = tm.Stop()
	assert.Error(t, err)
}

func TestTriggerManager_EvaluateDataChange(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	threshold := 5.0
	trigger := &Trigger{
		Type: TriggerTypeDataChange,
		Condition: &TriggerCondition{
			ChangeThreshold: &threshold,
		},
	}

	event := &TriggerEvent{OldValue: 10.0, NewValue: 20.0}
	assert.True(t, tm.evaluateDataChange(trigger, event))

	event2 := &TriggerEvent{OldValue: 10.0, NewValue: 12.0}
	assert.False(t, tm.evaluateDataChange(trigger, event2))
}

func TestTriggerManager_EvaluateDataChange_Direction(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	trigger := &Trigger{
		Type: TriggerTypeDataChange,
		Condition: &TriggerCondition{
			Direction: "up",
		},
	}

	upEvent := &TriggerEvent{OldValue: 10.0, NewValue: 20.0}
	assert.True(t, tm.evaluateDataChange(trigger, upEvent))

	downEvent := &TriggerEvent{OldValue: 20.0, NewValue: 10.0}
	assert.False(t, tm.evaluateDataChange(trigger, downEvent))

	trigger.Condition.Direction = "down"
	assert.True(t, tm.evaluateDataChange(trigger, downEvent))
	assert.False(t, tm.evaluateDataChange(trigger, upEvent))
}

func TestTriggerManager_EvaluateDataChange_Percent(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	percent := 50.0
	trigger := &Trigger{
		Type: TriggerTypeDataChange,
		Condition: &TriggerCondition{
			ChangePercent: &percent,
		},
	}

	bigChange := &TriggerEvent{OldValue: 100.0, NewValue: 200.0}
	assert.True(t, tm.evaluateDataChange(trigger, bigChange))

	smallChange := &TriggerEvent{OldValue: 100.0, NewValue: 120.0}
	assert.False(t, tm.evaluateDataChange(trigger, smallChange))
}

func TestTriggerManager_EvaluateDataChange_NoCondition(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	trigger := &Trigger{Type: TriggerTypeDataChange}
	event := &TriggerEvent{OldValue: 10.0, NewValue: 20.0}
	assert.False(t, tm.evaluateDataChange(trigger, event))
}

func TestTriggerManager_EvaluateEvent(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	trigger := &Trigger{Type: TriggerTypeEvent}
	event := &TriggerEvent{Metadata: map[string]string{"eventType": "alarm"}}
	assert.True(t, tm.evaluateEvent(trigger, event))

	trigger.Condition = &TriggerCondition{EventType: "alarm"}
	assert.True(t, tm.evaluateEvent(trigger, event))

	trigger.Condition.EventType = "other"
	assert.False(t, tm.evaluateEvent(trigger, event))
}

func TestTriggerManager_EvaluateCondition(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	trigger := &Trigger{Type: TriggerTypeCondition, Condition: &TriggerCondition{Expression: "value > 100"}}
	event := &TriggerEvent{NewValue: 50.0}
	assert.True(t, tm.evaluateCondition(trigger, event))

	trigger.Condition = &TriggerCondition{Expression: ""}
	assert.False(t, tm.evaluateCondition(trigger, event))

	trigger.Condition = nil
	assert.False(t, tm.evaluateCondition(trigger, event))
}

func TestTriggerManager_EvaluateComposite(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	threshold := 5.0
	trigger := &Trigger{
		Type: TriggerTypeComposite,
		Condition: &TriggerCondition{
			LogicOperator: "and",
			SubConditions: []*TriggerCondition{
				{ChangeThreshold: &threshold},
			},
		},
	}
	event := &TriggerEvent{OldValue: 10.0, NewValue: 20.0}
	assert.True(t, tm.evaluateComposite(trigger, event))

	trigger.Condition.LogicOperator = "or"
	assert.True(t, tm.evaluateComposite(trigger, event))

	trigger.Condition.LogicOperator = "unknown"
	assert.False(t, tm.evaluateComposite(trigger, event))

	trigger.Condition.SubConditions = nil
	assert.False(t, tm.evaluateComposite(trigger, event))
}

func TestPointManager_CreatePoint(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	point := &ComputePoint{
		ID:   "point-1",
		Name: "Test Point",
		Type: PointTypeVirtual,
	}
	err := pm.CreatePoint(ctx, point)
	require.NoError(t, err)
	assert.Equal(t, PointStatusActive, point.Status)
}

func TestPointManager_CreatePoint_NoID(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	err := pm.CreatePoint(ctx, &ComputePoint{Name: "No ID"})
	assert.Equal(t, ErrInvalidPointConfig, err)
}

func TestPointManager_CreatePoint_Duplicate(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual})
	err := pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual})
	assert.Equal(t, ErrPointExists, err)
}

func TestPointManager_UpdatePoint(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual, Name: "Original"})
	err := pm.UpdatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual, Name: "Updated"})
	require.NoError(t, err)

	point, _ := pm.GetPoint(ctx, "p1")
	assert.Equal(t, "Updated", point.Name)
}

func TestPointManager_UpdatePoint_NotFound(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	err := pm.UpdatePoint(ctx, &ComputePoint{ID: "nonexistent", Type: PointTypeVirtual})
	assert.Equal(t, ErrPointNotFound, err)
}

func TestPointManager_DeletePoint(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual})
	err := pm.DeletePoint(ctx, "p1")
	require.NoError(t, err)

	_, err = pm.GetPoint(ctx, "p1")
	assert.Equal(t, ErrPointNotFound, err)
}

func TestPointManager_DeletePoint_WithDependents(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual})
	pm.CreatePoint(ctx, &ComputePoint{ID: "p2", Type: PointTypeVirtual, Dependencies: []string{"p1"}})

	err := pm.DeletePoint(ctx, "p1")
	assert.Error(t, err)
}

func TestPointManager_DeletePoint_NotFound(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	err := pm.DeletePoint(ctx, "nonexistent")
	assert.Equal(t, ErrPointNotFound, err)
}

func TestPointManager_GetPointsByType(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual})
	pm.CreatePoint(ctx, &ComputePoint{ID: "p2", Type: PointTypeDerived})

	points := pm.GetPointsByType(ctx, PointTypeVirtual)
	assert.Equal(t, 1, len(points))
}

func TestPointManager_GetPointsByStatus(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual})
	pm.CreatePoint(ctx, &ComputePoint{ID: "p2", Type: PointTypeVirtual, Status: PointStatusError})

	points := pm.GetPointsByStatus(ctx, PointStatusActive)
	assert.Equal(t, 1, len(points))
}

func TestPointManager_GetAllPoints(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual})
	pm.CreatePoint(ctx, &ComputePoint{ID: "p2", Type: PointTypeDerived})

	points := pm.GetAllPoints(ctx)
	assert.Equal(t, 2, len(points))
}

func TestPointManager_SetGetPointConfig(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual})

	config := &PointConfig{ComputeInterval: time.Minute, Timeout: 5 * time.Second}
	err := pm.SetPointConfig(ctx, "p1", config)
	require.NoError(t, err)

	got, err := pm.GetPointConfig(ctx, "p1")
	require.NoError(t, err)
	assert.Equal(t, time.Minute, got.ComputeInterval)
}

func TestPointManager_SetPointConfig_NotFound(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	err := pm.SetPointConfig(ctx, "nonexistent", &PointConfig{})
	assert.Equal(t, ErrPointNotFound, err)
}

func TestPointManager_GetPointConfig_NotFound(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	_, err := pm.GetPointConfig(ctx, "nonexistent")
	assert.Equal(t, ErrPointNotFound, err)
}

func TestPointManager_UpdatePointStatus(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual})
	err := pm.UpdatePointStatus(ctx, "p1", PointStatusError)
	require.NoError(t, err)

	point, _ := pm.GetPoint(ctx, "p1")
	assert.Equal(t, PointStatusError, point.Status)
}

func TestPointManager_UpdatePointStatus_NotFound(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	err := pm.UpdatePointStatus(ctx, "nonexistent", PointStatusError)
	assert.Equal(t, ErrPointNotFound, err)
}

func TestPointManager_UpdatePointValue(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual})
	err := pm.UpdatePointValue(ctx, "p1", 42.5, 100)
	require.NoError(t, err)

	point, _ := pm.GetPoint(ctx, "p1")
	assert.Equal(t, 42.5, point.Value)
	assert.Equal(t, 100, point.Quality)
}

func TestPointManager_UpdatePointValue_NotFound(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	err := pm.UpdatePointValue(ctx, "nonexistent", 42.5, 100)
	assert.Equal(t, ErrPointNotFound, err)
}

func TestPointManager_GetDependencies(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual})
	pm.CreatePoint(ctx, &ComputePoint{ID: "p2", Type: PointTypeVirtual, Dependencies: []string{"p1"}})

	dep, err := pm.GetDependencies(ctx, "p2")
	require.NoError(t, err)
	assert.Equal(t, []string{"p1"}, dep.DependsOn)
}

func TestPointManager_GetDependencies_NotFound(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	_, err := pm.GetDependencies(ctx, "nonexistent")
	assert.Equal(t, ErrPointNotFound, err)
}

func TestPointManager_GetComputeOrder(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual})
	pm.CreatePoint(ctx, &ComputePoint{ID: "p2", Type: PointTypeVirtual, Dependencies: []string{"p1"}})
	pm.CreatePoint(ctx, &ComputePoint{ID: "p3", Type: PointTypeVirtual, Dependencies: []string{"p1", "p2"}})

	order := pm.GetComputeOrder(ctx)
	assert.Greater(t, len(order), 0)
}

func TestPointManager_GetDependents(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual})
	pm.CreatePoint(ctx, &ComputePoint{ID: "p2", Type: PointTypeVirtual, Dependencies: []string{"p1"}})

	dependents := pm.GetDependents(ctx, "p1")
	assert.Equal(t, 1, len(dependents))
	assert.Contains(t, dependents, "p2")
}

func TestPointManager_Filter(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual, Tags: map[string]string{"env": "prod"}})
	pm.CreatePoint(ctx, &ComputePoint{ID: "p2", Type: PointTypeDerived, Tags: map[string]string{"env": "dev"}})

	filter := &PointFilter{Types: []PointType{PointTypeVirtual}}
	points := pm.Filter(ctx, filter)
	assert.Equal(t, 1, len(points))

	filter2 := &PointFilter{Tags: map[string]string{"env": "prod"}}
	points2 := pm.Filter(ctx, filter2)
	assert.Equal(t, 1, len(points2))

	filter3 := &PointFilter{IDs: []string{"p1"}}
	points3 := pm.Filter(ctx, filter3)
	assert.Equal(t, 1, len(points3))

	filter4 := &PointFilter{Status: []PointStatus{PointStatusActive}}
	points4 := pm.Filter(ctx, filter4)
	assert.Equal(t, 2, len(points4))
}

func TestPointManager_GetStats(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	pm.CreatePoint(ctx, &ComputePoint{ID: "p1", Type: PointTypeVirtual})
	pm.CreatePoint(ctx, &ComputePoint{ID: "p2", Type: PointTypeDerived, Status: PointStatusError})

	stats := pm.GetStats(ctx)
	assert.Equal(t, 2, stats.TotalPoints)
	assert.Equal(t, 1, stats.ActivePoints)
	assert.Equal(t, 1, stats.ErrorPoints)
}

func TestDependencyGraph_HasCycle(t *testing.T) {
	dg := NewDependencyGraph()
	dg.AddNode("A", []string{})
	dg.AddNode("B", []string{"A"})

	hasCycle := dg.HasCycle("C", []string{"B"})
	assert.False(t, hasCycle)

	hasCycle = dg.HasCycle("A", []string{"C"})
	assert.False(t, hasCycle)
}

func TestDependencyGraph_UpdateNode(t *testing.T) {
	dg := NewDependencyGraph()
	dg.AddNode("A", []string{})

	err := dg.UpdateNode("A", []string{"B"})
	require.NoError(t, err)

	err = dg.UpdateNode("nonexistent", []string{})
	assert.Error(t, err)
}

func TestDependencyGraph_RemoveNode(t *testing.T) {
	dg := NewDependencyGraph()
	dg.AddNode("A", []string{})
	dg.RemoveNode("A")

	order := dg.GetTopologicalOrder()
	assert.Equal(t, 0, len(order))
}

func TestDependencyGraph_AddNode_Duplicate(t *testing.T) {
	dg := NewDependencyGraph()
	dg.AddNode("A", []string{})
	err := dg.AddNode("A", []string{})
	assert.Error(t, err)
}

func TestLocalLock_AcquireRelease(t *testing.T) {
	lock := NewLocalLock()
	ctx := context.Background()

	acquired, err := lock.Acquire(ctx, "key1", 5*time.Second)
	require.NoError(t, err)
	assert.True(t, acquired)

	held, err := lock.IsHeld(ctx, "key1")
	require.NoError(t, err)
	assert.True(t, held)

	err = lock.Release(ctx, "key1")
	require.NoError(t, err)

	held, _ = lock.IsHeld(ctx, "key1")
	assert.False(t, held)
}

func TestLocalLock_Acquire_AlreadyHeld(t *testing.T) {
	lock := NewLocalLock()
	ctx := context.Background()

	lock.Acquire(ctx, "key1", 5*time.Second)
	acquired, err := lock.Acquire(ctx, "key1", 5*time.Second)
	require.NoError(t, err)
	assert.False(t, acquired)
}

func TestLocalLock_Acquire_Expired(t *testing.T) {
	lock := NewLocalLock()
	ctx := context.Background()

	lock.Acquire(ctx, "key1", 1*time.Millisecond)
	time.Sleep(5 * time.Millisecond)

	acquired, err := lock.Acquire(ctx, "key1", 5*time.Second)
	require.NoError(t, err)
	assert.True(t, acquired)
}

func TestLocalLock_IsHeld_Expired(t *testing.T) {
	lock := NewLocalLock()
	ctx := context.Background()

	lock.Acquire(ctx, "key1", 1*time.Millisecond)
	time.Sleep(5 * time.Millisecond)

	held, err := lock.IsHeld(ctx, "key1")
	require.NoError(t, err)
	assert.False(t, held)
}

func TestLocalLock_IsHeld_NotHeld(t *testing.T) {
	lock := NewLocalLock()
	ctx := context.Background()

	held, err := lock.IsHeld(ctx, "nonexistent")
	require.NoError(t, err)
	assert.False(t, held)
}

func TestLockManager_LocalLock(t *testing.T) {
	lm := NewLockManager(nil)
	ctx := context.Background()

	acquired, err := lm.Acquire(ctx, "key1", 5*time.Second)
	require.NoError(t, err)
	assert.True(t, acquired)

	held, err := lm.IsHeld(ctx, "key1")
	require.NoError(t, err)
	assert.True(t, held)

	err = lm.Release(ctx, "key1")
	require.NoError(t, err)
}

func TestLockManager_TryAcquire(t *testing.T) {
	lm := NewLockManager(nil)
	ctx := context.Background()

	acquired, err := lm.TryAcquire(ctx, "key1", 5*time.Second, 3, 10*time.Millisecond)
	require.NoError(t, err)
	assert.True(t, acquired)
}

func TestLockManager_WithLock(t *testing.T) {
	lm := NewLockManager(nil)
	ctx := context.Background()

	executed := false
	err := lm.WithLock(ctx, "key1", 5*time.Second, func() error {
		executed = true
		return nil
	})
	require.NoError(t, err)
	assert.True(t, executed)
}

func TestLockManager_WithLock_FailedAcquire(t *testing.T) {
	lm := NewLockManager(nil)
	ctx := context.Background()

	lm.Acquire(ctx, "key1", 5*time.Second)
	err := lm.WithLock(ctx, "key1", 5*time.Second, func() error {
		return nil
	})
	assert.Equal(t, ErrLockFailed, err)
}

func TestContainsPattern(t *testing.T) {
	assert.True(t, containsPattern("abc-test-123", "*test*"))
	assert.True(t, containsPattern("exact-match", "exact-match"))
	assert.True(t, containsPattern("anything", ""))
	assert.False(t, containsPattern("abc-def", "*xyz*"))
}

func TestContains(t *testing.T) {
	assert.True(t, contains("hello world", "world"))
	assert.True(t, contains("hello", "hello"))
	assert.False(t, contains("hello", "xyz"))
	assert.True(t, contains("abc", "bc"))
}

func TestCachePolicyConstants(t *testing.T) {
	assert.Equal(t, CachePolicy("lru"), CachePolicyLRU)
	assert.Equal(t, CachePolicy("lfu"), CachePolicyLFU)
	assert.Equal(t, CachePolicy("fifo"), CachePolicyFIFO)
	assert.Equal(t, CachePolicy("ttl"), CachePolicyTTL)
	assert.Equal(t, CachePolicy("none"), CachePolicyNone)
}

func TestRuleTypeConstants(t *testing.T) {
	assert.Equal(t, RuleType("formula"), RuleTypeFormula)
	assert.Equal(t, RuleType("expression"), RuleTypeExpression)
	assert.Equal(t, RuleType("script"), RuleTypeScript)
	assert.Equal(t, RuleType("aggregate"), RuleTypeAggregate)
	assert.Equal(t, RuleType("transform"), RuleTypeTransform)
}

func TestRuleStatusConstants(t *testing.T) {
	assert.Equal(t, RuleStatus("active"), RuleStatusActive)
	assert.Equal(t, RuleStatus("inactive"), RuleStatusInactive)
	assert.Equal(t, RuleStatus("error"), RuleStatusError)
	assert.Equal(t, RuleStatus("disabled"), RuleStatusDisabled)
}

func TestPointTypeConstants(t *testing.T) {
	assert.Equal(t, PointType("virtual"), PointTypeVirtual)
	assert.Equal(t, PointType("derived"), PointTypeDerived)
	assert.Equal(t, PointType("aggregate"), PointTypeAggregate)
	assert.Equal(t, PointType("statistic"), PointTypeStatistic)
}

func TestTaskTypeConstants(t *testing.T) {
	assert.Equal(t, TaskType("cron"), TaskTypeCron)
	assert.Equal(t, TaskType("interval"), TaskTypeInterval)
	assert.Equal(t, TaskType("once"), TaskTypeOnce)
}

func TestTriggerTypeConstants(t *testing.T) {
	assert.Equal(t, TriggerType("data_change"), TriggerTypeDataChange)
	assert.Equal(t, TriggerType("event"), TriggerTypeEvent)
	assert.Equal(t, TriggerType("condition"), TriggerTypeCondition)
	assert.Equal(t, TriggerType("composite"), TriggerTypeComposite)
}

func TestComputeScheduler_StartWithTasks(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	scheduler.AddTask(&ComputeTask{
		ID:          "task-1",
		Type:        TaskTypeInterval,
		Interval:    100 * time.Millisecond,
		Enabled:     true,
		PointIDs:    []string{"p1"},
		Timeout:     5 * time.Second,
		NextRunTime: time.Now().Add(-1 * time.Second),
	})

	err := scheduler.Start()
	require.NoError(t, err)

	time.Sleep(300 * time.Millisecond)

	err = scheduler.Stop()
	require.NoError(t, err)

	task, _ := scheduler.GetTask("task-1")
	assert.GreaterOrEqual(t, task.RunCount, int64(1))
}

func TestComputeScheduler_CronTask(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	scheduler.AddTask(&ComputeTask{
		ID:       "task-cron",
		Type:     TaskTypeCron,
		CronExpr: "@every 100ms",
		Enabled:  true,
		PointIDs: []string{"p1"},
		Timeout:  5 * time.Second,
	})

	task, _ := scheduler.GetTask("task-cron")
	task.NextRunTime = time.Now().Add(-1 * time.Second)

	scheduler.Start()
	time.Sleep(300 * time.Millisecond)
	scheduler.Stop()

	task, _ = scheduler.GetTask("task-cron")
	assert.GreaterOrEqual(t, task.RunCount, int64(1))
}

func TestComputeScheduler_OnceTask(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	scheduler.AddTask(&ComputeTask{
		ID:          "task-once",
		Type:        TaskTypeOnce,
		Enabled:     true,
		PointIDs:    []string{"p1"},
		Timeout:     5 * time.Second,
		NextRunTime: time.Now().Add(-1 * time.Second),
	})

	scheduler.Start()
	time.Sleep(200 * time.Millisecond)
	scheduler.Stop()
}

func TestComputeScheduler_WithLock(t *testing.T) {
	executor := &mockExecutor{}
	lock := NewLocalLock()
	scheduler := NewComputeScheduler(nil, executor, lock)

	scheduler.AddTask(&ComputeTask{
		ID:          "task-lock",
		Type:        TaskTypeInterval,
		Interval:    100 * time.Millisecond,
		Enabled:     true,
		PointIDs:    []string{"p1"},
		Timeout:     5 * time.Second,
		NextRunTime: time.Now().Add(-1 * time.Second),
	})

	scheduler.Start()
	time.Sleep(300 * time.Millisecond)
	scheduler.Stop()
}

func TestComputeScheduler_ErrorTask(t *testing.T) {
	executor := &errorMockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	scheduler.AddTask(&ComputeTask{
		ID:          "task-err",
		Type:        TaskTypeInterval,
		Interval:    100 * time.Millisecond,
		Enabled:     true,
		PointIDs:    []string{"p1"},
		Timeout:     5 * time.Second,
		NextRunTime: time.Now().Add(-1 * time.Second),
		MaxRetry:    2,
	})

	scheduler.Start()
	time.Sleep(300 * time.Millisecond)
	scheduler.Stop()

	task, _ := scheduler.GetTask("task-err")
	assert.GreaterOrEqual(t, task.FailCount, int64(1))
}

func TestComputeScheduler_GetLogsAfterRun(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	scheduler.AddTask(&ComputeTask{
		ID:          "task-log",
		Type:        TaskTypeInterval,
		Interval:    100 * time.Millisecond,
		Enabled:     true,
		PointIDs:    []string{"p1"},
		Timeout:     5 * time.Second,
		NextRunTime: time.Now().Add(-1 * time.Second),
	})

	scheduler.Start()
	time.Sleep(300 * time.Millisecond)
	scheduler.Stop()

	logs := scheduler.GetLogs("", 10)
	assert.GreaterOrEqual(t, len(logs), 1)

	logsByTask := scheduler.GetLogs("task-log", 10)
	assert.GreaterOrEqual(t, len(logsByTask), 1)
}

func TestComputeScheduler_MetricsAfterRun(t *testing.T) {
	executor := &mockExecutor{}
	scheduler := NewComputeScheduler(nil, executor, nil)

	scheduler.AddTask(&ComputeTask{
		ID:          "task-metrics",
		Type:        TaskTypeInterval,
		Interval:    100 * time.Millisecond,
		Enabled:     true,
		PointIDs:    []string{"p1"},
		Timeout:     5 * time.Second,
		NextRunTime: time.Now().Add(-1 * time.Second),
	})

	scheduler.Start()
	time.Sleep(300 * time.Millisecond)
	scheduler.Stop()

	metrics := scheduler.GetMetrics()
	assert.GreaterOrEqual(t, metrics.TotalTasks, int64(1))
	assert.GreaterOrEqual(t, metrics.ScheduledTasks, int64(1))
}

func TestTriggerManager_HandleDataChangeEvent(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)
	tm.Start()
	defer tm.Stop()

	threshold := 5.0
	tm.CreateTrigger(&Trigger{
		ID:       "dc-trigger",
		Type:     TriggerTypeDataChange,
		Enabled:  true,
		PointIDs: []string{"p1"},
		Condition: &TriggerCondition{
			ChangeThreshold: &threshold,
		},
	})

	tm.OnDataChange(context.Background(), "p1", 10.0, 20.0)
	time.Sleep(100 * time.Millisecond)

	trigger, _ := tm.GetTrigger("dc-trigger")
	assert.GreaterOrEqual(t, trigger.TriggerCount, int64(1))
}

func TestTriggerManager_HandleEventTrigger(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)
	tm.Start()
	defer tm.Stop()

	tm.CreateTrigger(&Trigger{
		ID:       "evt-trigger",
		Type:     TriggerTypeEvent,
		Enabled:  true,
		Condition: &TriggerCondition{
			EventType: "alarm",
		},
	})

	tm.OnEvent(context.Background(), "alarm", map[string]interface{}{"data": "test"})
	time.Sleep(100 * time.Millisecond)

	trigger, _ := tm.GetTrigger("evt-trigger")
	assert.GreaterOrEqual(t, trigger.TriggerCount, int64(1))
}

func TestTriggerManager_HandleConditionTrigger(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)
	tm.Start()
	defer tm.Stop()

	tm.CreateTrigger(&Trigger{
		ID:       "cond-trigger",
		Type:     TriggerTypeCondition,
		Enabled:  true,
		PointIDs: []string{"p1"},
		Condition: &TriggerCondition{
			Expression: "value > 100",
		},
	})

	tm.OnCondition(context.Background(), "p1", "value > 100", 150.0)
	time.Sleep(100 * time.Millisecond)

	trigger, _ := tm.GetTrigger("cond-trigger")
	assert.GreaterOrEqual(t, trigger.TriggerCount, int64(1))
}

func TestTriggerManager_CompositeTrigger(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)
	tm.Start()
	defer tm.Stop()

	threshold := 5.0
	tm.CreateTrigger(&Trigger{
		ID:       "comp-trigger",
		Type:     TriggerTypeComposite,
		Enabled:  true,
		PointIDs: []string{"p1"},
		Condition: &TriggerCondition{
			LogicOperator: "and",
			SubConditions: []*TriggerCondition{
				{ChangeThreshold: &threshold},
			},
		},
	})

	tm.OnDataChange(context.Background(), "p1", 10.0, 20.0)
	time.Sleep(100 * time.Millisecond)
}

func TestTriggerManager_CooldownTrigger(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)
	tm.Start()
	defer tm.Stop()

	threshold := 5.0
	tm.CreateTrigger(&Trigger{
		ID:       "cool-trigger",
		Type:     TriggerTypeDataChange,
		Enabled:  true,
		PointIDs: []string{"p1"},
		Condition: &TriggerCondition{
			ChangeThreshold: &threshold,
			CooldownTime:    1 * time.Hour,
		},
	})

	tm.OnDataChange(context.Background(), "p1", 10.0, 20.0)
	time.Sleep(100 * time.Millisecond)

	trigger, _ := tm.GetTrigger("cool-trigger")
	assert.GreaterOrEqual(t, trigger.TriggerCount, int64(1))

	tm.OnDataChange(context.Background(), "p1", 20.0, 30.0)
	time.Sleep(100 * time.Millisecond)

	trigger2, _ := tm.GetTrigger("cool-trigger")
	assert.Equal(t, trigger.TriggerCount, trigger2.TriggerCount)
}

func TestTriggerManager_TriggerWithActions(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)
	tm.Start()
	defer tm.Stop()

	threshold := 5.0
	tm.CreateTrigger(&Trigger{
		ID:       "action-trigger",
		Type:     TriggerTypeDataChange,
		Enabled:  true,
		PointIDs: []string{"p1"},
		Condition: &TriggerCondition{
			ChangeThreshold: &threshold,
		},
		Actions: []TriggerAction{
			{Type: "compute", Target: "rule-1", Params: map[string]interface{}{"ruleID": "rule-1"}},
			{Type: "notify", Target: "email", Params: map[string]interface{}{"channel": "email"}},
		},
	})

	tm.OnDataChange(context.Background(), "p1", 10.0, 20.0)
	time.Sleep(100 * time.Millisecond)
}

func TestTriggerManager_TriggerWithChain(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)
	tm.Start()
	defer tm.Stop()

	tm.CreateChain(&TriggerChain{
		ID:          "chain-1",
		Name:        "Test Chain",
		Triggers:    []string{"next-trigger"},
	})

	threshold := 5.0
	tm.CreateTrigger(&Trigger{
		ID:       "chain-trigger",
		Type:     TriggerTypeDataChange,
		Enabled:  true,
		PointIDs: []string{"p1"},
		Condition: &TriggerCondition{
			ChangeThreshold: &threshold,
		},
		ChainConfig: &TriggerChainConfig{
			ChainID:      "chain-1",
			NextTriggers: []string{"next-trigger"},
		},
	})

	tm.OnDataChange(context.Background(), "p1", 10.0, 20.0)
	time.Sleep(100 * time.Millisecond)
}

func TestTriggerManager_UnknownTriggerType(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)
	tm.Start()
	defer tm.Stop()

	tm.CreateTrigger(&Trigger{
		ID:       "unknown-trigger",
		Type:     TriggerType("unknown"),
		Enabled:  true,
		PointIDs: []string{"p1"},
	})

	tm.OnDataChange(context.Background(), "p1", 10.0, 20.0)
	time.Sleep(100 * time.Millisecond)
}

func TestTriggerManager_GetStats_AfterTriggers(t *testing.T) {
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	tm.CreateTrigger(&Trigger{ID: "t1", Type: TriggerTypeDataChange, Enabled: true})
	tm.CreateTrigger(&Trigger{ID: "t2", Type: TriggerTypeEvent, Enabled: false})

	stats := tm.GetStats()
	assert.NotNil(t, stats)
}

func TestPriorityQueueComprehensive(t *testing.T) {
	pq := NewPriorityQueue()
	pq.Push(&ComputeTask{ID: "low", Priority: 1})
	pq.Push(&ComputeTask{ID: "high", Priority: 10})
	pq.Push(&ComputeTask{ID: "mid", Priority: 5})

	assert.Equal(t, 3, pq.Len())

	task := pq.Pop()
	assert.Equal(t, "high", task.ID)
}

func TestPriorityQueue_PopEmpty(t *testing.T) {
	pq := NewPriorityQueue()
	task := pq.Pop()
	assert.Nil(t, task)
}

func TestPriorityQueue_SortByTime(t *testing.T) {
	pq := NewPriorityQueue()
	pq.Push(&ComputeTask{ID: "later", Priority: 5, NextRunTime: time.Now().Add(1 * time.Hour)})
	pq.Push(&ComputeTask{ID: "sooner", Priority: 5, NextRunTime: time.Now().Add(1 * time.Minute)})

	task := pq.Pop()
	assert.Equal(t, "sooner", task.ID)
}

type errorMockExecutor struct{}

func (m *errorMockExecutor) Execute(ctx context.Context, pointIDs []string) (map[string]*ComputeResult, error) {
	return nil, fmt.Errorf("execution error")
}
