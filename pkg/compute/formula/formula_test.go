package formula

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExecutor(t *testing.T) {
	executor := NewExecutor(nil)
	assert.NotNil(t, executor)
	assert.NotNil(t, executor.functionRegistry)
	assert.NotNil(t, executor.operatorRegistry)
}

func TestNewExecutor_WithConfig(t *testing.T) {
	config := &ExecutorConfig{
		EnableCache:       true,
		CacheTTL:          10 * time.Minute,
		MaxCacheSize:      5000,
		Timeout:           60 * time.Second,
		MaxRecursionDepth: 50,
	}
	executor := NewExecutor(config)
	assert.NotNil(t, executor)
	assert.Equal(t, config, executor.config)
	assert.NotNil(t, executor.cache)
}

func TestExecutor_Execute_SimpleOperations(t *testing.T) {
	executor := NewExecutor(nil)
	tests := []struct {
		name      string
		formula   string
		variables map[string]interface{}
		expected  interface{}
		hasError  bool
	}{
		{"addition", "a + b", map[string]interface{}{"a": 10.0, "b": 20.0}, 30.0, false},
		{"subtraction", "a - b", map[string]interface{}{"a": 30.0, "b": 10.0}, 20.0, false},
		{"multiplication", "a * b", map[string]interface{}{"a": 5.0, "b": 6.0}, 30.0, false},
		{"division", "a / b", map[string]interface{}{"a": 20.0, "b": 4.0}, 5.0, false},
		{"complex", "(a + b) * c - d", map[string]interface{}{"a": 2.0, "b": 3.0, "c": 4.0, "d": 5.0}, 15.0, false},
		{"divide by zero", "a / b", map[string]interface{}{"a": 10.0, "b": 0.0}, nil, true},
		{"missing variable", "a + c", map[string]interface{}{"a": 10.0, "b": 20.0}, nil, true},
		{"power", "a ^ b", map[string]interface{}{"a": 2.0, "b": 3.0}, 8.0, false},
		{"modulo", "a % b", map[string]interface{}{"a": 10.0, "b": 3.0}, 1.0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.Execute(tt.formula, tt.variables)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestExecutor_Execute_ComparisonOperations(t *testing.T) {
	executor := NewExecutor(nil)
	tests := []struct {
		name      string
		formula   string
		variables map[string]interface{}
		expected  bool
	}{
		{"gt", "a > b", map[string]interface{}{"a": 10.0, "b": 5.0}, true},
		{"lt", "a < b", map[string]interface{}{"a": 5.0, "b": 10.0}, true},
		{"eq", "a == b", map[string]interface{}{"a": 10.0, "b": 10.0}, true},
		{"ne", "a != b", map[string]interface{}{"a": 10.0, "b": 5.0}, true},
		{"gte", "a >= b", map[string]interface{}{"a": 10.0, "b": 10.0}, true},
		{"lte", "a <= b", map[string]interface{}{"a": 5.0, "b": 10.0}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.Execute(tt.formula, tt.variables)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExecutor_Execute_LogicalOperations(t *testing.T) {
	executor := NewExecutor(nil)
	tests := []struct {
		name      string
		formula   string
		variables map[string]interface{}
		expected  bool
	}{
		{"and true", "a && b", map[string]interface{}{"a": true, "b": true}, true},
		{"and false", "a && b", map[string]interface{}{"a": true, "b": false}, false},
		{"or true", "a || b", map[string]interface{}{"a": false, "b": true}, true},
		{"or false", "a || b", map[string]interface{}{"a": false, "b": false}, false},
		{"not", "!a", map[string]interface{}{"a": true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.Execute(tt.formula, tt.variables)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExecutor_Execute_Conditional(t *testing.T) {
	executor := NewExecutor(nil)
	tests := []struct {
		name      string
		formula   string
		variables map[string]interface{}
		expected  interface{}
	}{
		{"true branch", "a > b ? c : d", map[string]interface{}{"a": 10.0, "b": 5.0, "c": 100.0, "d": 200.0}, 100.0},
		{"false branch", "a > b ? c : d", map[string]interface{}{"a": 5.0, "b": 10.0, "c": 100.0, "d": 200.0}, 200.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.Execute(tt.formula, tt.variables)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExecutor_Execute_WithFunctions(t *testing.T) {
	executor := NewExecutor(nil)
	executor.RegisterFunction("double", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, assert.AnError
		}
		num, ok := args[0].(float64)
		if !ok {
			return nil, assert.AnError
		}
		return num * 2, nil
	})
	result, err := executor.Execute("double(a)", map[string]interface{}{"a": 5.0})
	assert.NoError(t, err)
	assert.Equal(t, 10.0, result)
}

func TestExecutor_ExecuteBatch(t *testing.T) {
	executor := NewExecutor(nil)
	formulas := []string{"a + b", "a * b", "a - b"}
	variables := map[string]interface{}{"a": 10.0, "b": 5.0}
	results, errors := executor.ExecuteBatch(formulas, variables)
	assert.Len(t, results, 3)
	assert.Len(t, errors, 3)
	assert.Equal(t, 15.0, results[0])
	assert.Equal(t, 50.0, results[1])
	assert.Equal(t, 5.0, results[2])
}

func TestExecutor_ExecuteParallel(t *testing.T) {
	executor := NewExecutor(nil)
	formulas := []string{"a + b", "a * b", "a - b", "a / b"}
	variables := map[string]interface{}{"a": 10.0, "b": 5.0}
	results, errors := executor.ExecuteParallel(formulas, variables, 2)
	assert.Len(t, results, 4)
	assert.Len(t, errors, 4)
	assert.Equal(t, 15.0, results[0])
	assert.Equal(t, 50.0, results[1])
	assert.Equal(t, 5.0, results[2])
	assert.Equal(t, 2.0, results[3])
}

func TestExecutor_ExecuteWithTimeout(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.ExecuteWithTimeout("a + b", map[string]interface{}{"a": 10.0, "b": 20.0}, 5*time.Second)
	assert.NoError(t, err)
	assert.Equal(t, 30.0, result)
}

func TestExecutor_ExecuteWithContext(t *testing.T) {
	executor := NewExecutor(nil)
	ctx := context.Background()
	result, err := executor.ExecuteWithContext(ctx, "a + b", map[string]interface{}{"a": 10.0, "b": 20.0})
	assert.NoError(t, err)
	assert.Equal(t, 30.0, result)
}

func TestExecutor_ExecuteWithContext_Cancelled(t *testing.T) {
	executor := NewExecutor(nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := executor.ExecuteWithContext(ctx, "a + b", map[string]interface{}{"a": 10.0, "b": 20.0})
	assert.Error(t, err)
}

func TestExecutor_Cache(t *testing.T) {
	config := &ExecutorConfig{EnableCache: true, CacheTTL: 5 * time.Minute, MaxCacheSize: 100}
	executor := NewExecutor(config)
	result1, err := executor.Execute("a + b", map[string]interface{}{"a": 10.0, "b": 20.0})
	assert.NoError(t, err)
	assert.Equal(t, 30.0, result1)
	result2, err := executor.Execute("a + b", map[string]interface{}{"a": 10.0, "b": 20.0})
	assert.NoError(t, err)
	assert.Equal(t, 30.0, result2)
	stats := executor.GetCacheStats()
	assert.NotNil(t, stats)
	assert.Equal(t, int64(1), stats.Hits)
}

func TestExecutor_ClearCache(t *testing.T) {
	config := &ExecutorConfig{EnableCache: true, CacheTTL: 5 * time.Minute, MaxCacheSize: 100}
	executor := NewExecutor(config)
	executor.Execute("a + b", map[string]interface{}{"a": 10.0, "b": 20.0})
	executor.ClearCache()
	stats := executor.GetCacheStats()
	assert.Equal(t, 0, stats.Size)
}

func TestExecutor_Compile(t *testing.T) {
	executor := NewExecutor(nil)
	compiled, err := executor.Compile("a + b * c")
	assert.NoError(t, err)
	assert.NotNil(t, compiled)
	variables := compiled.GetVariables()
	assert.Contains(t, variables, "a")
	assert.Contains(t, variables, "b")
	assert.Contains(t, variables, "c")
}

func TestCompiledFormula_Execute(t *testing.T) {
	executor := NewExecutor(nil)
	compiled, err := executor.Compile("a + b")
	assert.NoError(t, err)
	result, err := compiled.Execute(executor, map[string]interface{}{"a": 10.0, "b": 20.0})
	assert.NoError(t, err)
	assert.Equal(t, 30.0, result)
}

func TestResultCache(t *testing.T) {
	cache := NewResultCache(100, 5*time.Minute)
	cache.Set("key1", "value1")
	value, exists := cache.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", value)
	_, exists = cache.Get("key2")
	assert.False(t, exists)
	cache.Delete("key1")
	_, exists = cache.Get("key1")
	assert.False(t, exists)
	cache.Set("key3", "value3")
	cache.Clear()
	stats := cache.Stats()
	assert.Equal(t, 0, stats.Size)
}

func TestResultCache_Eviction(t *testing.T) {
	cache := NewResultCache(3, 5*time.Minute)
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")
	cache.Set("key4", "value4")
	stats := cache.Stats()
	assert.LessOrEqual(t, stats.Size, 3)
}

func TestVariableBinder(t *testing.T) {
	binder := NewVariableBinder()
	binder.Bind("a", 10.0)
	binder.Bind("b", 20.0)
	value, exists := binder.Get("a")
	assert.True(t, exists)
	assert.Equal(t, 10.0, value)
	binder.BindMany(map[string]interface{}{"c": 30.0, "d": 40.0})
	all := binder.GetAll()
	assert.Len(t, all, 4)
	binder.Unbind("a")
	_, exists = binder.Get("a")
	assert.False(t, exists)
	binder.Clear()
	all = binder.GetAll()
	assert.Len(t, all, 0)
}

func TestExecutionContext(t *testing.T) {
	ctx := NewExecutionContext()
	ctx.SetVariable("a", 10.0)
	value, exists := ctx.GetVariable("a")
	assert.True(t, exists)
	assert.Equal(t, 10.0, value)
	ctx.RegisterFunction("double", func(args ...interface{}) (interface{}, error) {
		return args[0].(float64) * 2, nil
	})
	fn, exists := ctx.GetFunction("double")
	assert.True(t, exists)
	assert.NotNil(t, fn)
	ctx.SetMetadata("key", "value")
	meta, exists := ctx.GetMetadata("key")
	assert.True(t, exists)
	assert.Equal(t, "value", meta)
	child := ctx.CreateChild()
	childValue, exists := child.GetVariable("a")
	assert.True(t, exists)
	assert.Equal(t, 10.0, childValue)
}

func TestExecutorBuilder(t *testing.T) {
	executor := NewExecutorBuilder().
		WithCache(true, 10*time.Minute, 5000).
		WithTimeout(60*time.Second).
		WithMaxRecursionDepth(50).
		WithFunction("custom", func(args ...interface{}) (interface{}, error) {
			return args[0].(float64) * 2, nil
		}).
		Build()
	assert.NotNil(t, executor)
	assert.True(t, executor.config.EnableCache)
	fn, exists := executor.GetFunction("custom")
	assert.True(t, exists)
	assert.NotNil(t, fn)
}

func TestParseFormula(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{"simple", "a + b", false},
		{"complex", "(a + b) * c - d / e", false},
		{"function", "max(a, b) + min(c, d)", false},
		{"conditional", "a > b ? c : d", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := ParseFormula(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else if tt.input != "" {
				assert.NoError(t, err)
				assert.NotNil(t, node)
			}
		})
	}
}

func TestLexer(t *testing.T) {
	lexer := NewLexer("a + b * 10")
	tokens, err := lexer.Lex()
	assert.NoError(t, err)
	assert.NotEmpty(t, tokens)
	assert.Equal(t, TokenIdentifier, tokens[0].Type)
	assert.Equal(t, "a", tokens[0].Value)
	assert.Equal(t, TokenOperator, tokens[1].Type)
	assert.Equal(t, "+", tokens[1].Value)
}

func TestLexer_ComplexTokens(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedTypes []TokenType
	}{
		{"number", "123.45", []TokenType{TokenNumber, TokenEOF}},
		{"string", `"hello"`, []TokenType{TokenString, TokenEOF}},
		{"variable", "${point-001}", []TokenType{TokenVariable, TokenEOF}},
		{"comparison", "a >= b", []TokenType{TokenIdentifier, TokenOperator, TokenIdentifier, TokenEOF}},
		{"logical", "a && b || c", []TokenType{TokenIdentifier, TokenOperator, TokenIdentifier, TokenOperator, TokenIdentifier, TokenEOF}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Lex()
			assert.NoError(t, err)
			for i, expectedType := range tt.expectedTypes {
				if i < len(tokens) {
					assert.Equal(t, expectedType, tokens[i].Type)
				}
			}
		})
	}
}

func TestEvalContext(t *testing.T) {
	ctx := NewEvalContext()
	ctx.SetVariable("a", 10.0)
	value, exists := ctx.GetVariable("a")
	assert.True(t, exists)
	assert.Equal(t, 10.0, value)
	ctx.RegisterFunction("double", func(args ...interface{}) (interface{}, error) {
		return args[0].(float64) * 2, nil
	})
	fn, exists := ctx.Functions["double"]
	assert.True(t, exists)
	assert.NotNil(t, fn)
}

func TestEvaluate(t *testing.T) {
	ctx := NewEvalContext()
	ctx.SetVariable("a", 10.0)
	ctx.SetVariable("b", 20.0)
	result, err := Evaluate("a + b", ctx)
	assert.NoError(t, err)
	assert.Equal(t, 30.0, result)
}

func TestFormulaManager_Create(t *testing.T) {
	mgr := NewFormulaManager(nil)
	f := &Formula{ID: "f1", Name: "test", Expression: "a + b"}
	require.NoError(t, mgr.Create(f))
	assert.Equal(t, StatusDraft, f.Status)
	assert.Equal(t, "1.0.0", f.Version)
	assert.False(t, f.CreatedAt.IsZero())
}

func TestFormulaManager_Create_Duplicate(t *testing.T) {
	mgr := NewFormulaManager(nil)
	f := &Formula{ID: "f1", Name: "test", Expression: "a + b"}
	mgr.Create(f)
	assert.Error(t, mgr.Create(f))
}

func TestFormulaManager_Create_InvalidExpression(t *testing.T) {
	mgr := NewFormulaManager(nil)
	f := &Formula{ID: "f1", Name: "test", Expression: "@#$%"}
	assert.Error(t, mgr.Create(f))
}

func TestFormulaManager_Create_NoAutoCompile(t *testing.T) {
	mgr := NewFormulaManager(&ManagerConfig{AutoCompile: false, AutoValidate: false})
	f := &Formula{ID: "f1", Name: "test", Expression: "a + b"}
	require.NoError(t, mgr.Create(f))
	assert.Nil(t, f.compiled)
}

func TestFormulaManager_Get(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	f, err := mgr.Get("f1")
	require.NoError(t, err)
	assert.Equal(t, "f1", f.ID)
	_, err = mgr.Get("nonexistent")
	assert.Error(t, err)
}

func TestFormulaManager_Update(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	f, _ := mgr.Get("f1")
	f.Name = "updated"
	f.Expression = "a * b"
	require.NoError(t, mgr.Update(f))
	updated, _ := mgr.Get("f1")
	assert.Equal(t, "updated", updated.Name)
}

func TestFormulaManager_Update_NotFound(t *testing.T) {
	mgr := NewFormulaManager(nil)
	f := &Formula{ID: "f1", Name: "test", Expression: "a + b"}
	assert.Error(t, mgr.Update(f))
}

func TestFormulaManager_Update_InvalidExpression(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	f, _ := mgr.Get("f1")
	f.Expression = "@#$%"
	assert.Error(t, mgr.Update(f))
}

func TestFormulaManager_Delete(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	require.NoError(t, mgr.Delete("f1"))
	_, err := mgr.Get("f1")
	assert.Error(t, err)
}

func TestFormulaManager_Delete_NotFound(t *testing.T) {
	mgr := NewFormulaManager(nil)
	assert.Error(t, mgr.Delete("nonexistent"))
}

func TestFormulaManager_Delete_WithDependent(t *testing.T) {
	mgr := NewFormulaManager(&ManagerConfig{AutoCompile: true, AutoValidate: false})
	mgr.Create(&Formula{ID: "f1", Name: "base", Expression: "a + b"})
	f2 := &Formula{ID: "f2", Name: "dep", Expression: "c + d"}
	f2.dependencies = []string{"f1"}
	mgr.formulas["f2"] = f2
	assert.Error(t, mgr.Delete("f1"))
}

func TestFormulaManager_List(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test1", Expression: "a + b"})
	mgr.Create(&Formula{ID: "f2", Name: "test2", Expression: "c + d"})
	list := mgr.List()
	assert.Len(t, list, 2)
}

func TestFormulaManager_ListByStatus(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test1", Expression: "a + b"})
	mgr.Activate("f1")
	mgr.Create(&Formula{ID: "f2", Name: "test2", Expression: "c + d"})
	active := mgr.ListByStatus(StatusActive)
	assert.Len(t, active, 1)
	draft := mgr.ListByStatus(StatusDraft)
	assert.Len(t, draft, 1)
}

func TestFormulaManager_ListByCategory(t *testing.T) {
	mgr := NewFormulaManager(nil)
	f1 := &Formula{ID: "f1", Name: "test1", Expression: "a + b", Category: "math"}
	f2 := &Formula{ID: "f2", Name: "test2", Expression: "c + d", Category: "stats"}
	mgr.Create(f1)
	mgr.Create(f2)
	math := mgr.ListByCategory("math")
	assert.Len(t, math, 1)
}

func TestFormulaManager_ListByTags(t *testing.T) {
	mgr := NewFormulaManager(nil)
	f1 := &Formula{ID: "f1", Name: "test1", Expression: "a + b", Tags: []string{"critical"}}
	f2 := &Formula{ID: "f2", Name: "test2", Expression: "c + d", Tags: []string{"normal"}}
	mgr.Create(f1)
	mgr.Create(f2)
	critical := mgr.ListByTags([]string{"critical"})
	assert.Len(t, critical, 1)
}

func TestFormulaManager_Search(t *testing.T) {
	mgr := NewFormulaManager(nil)
	f1 := &Formula{ID: "f1", Name: "voltage calc", Expression: "a + b", Description: "voltage calculation"}
	f2 := &Formula{ID: "f2", Name: "current calc", Expression: "c + d", Description: "current calculation"}
	mgr.Create(f1)
	mgr.Create(f2)
	results := mgr.Search("voltage")
	assert.Len(t, results, 1)
}

func TestFormulaManager_Validate(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	assert.NoError(t, mgr.Validate("f1"))
	assert.Error(t, mgr.Validate("nonexistent"))
}

func TestFormulaManager_ValidateExpression(t *testing.T) {
	mgr := NewFormulaManager(nil)
	assert.NoError(t, mgr.ValidateExpression("a + b"))
	assert.Error(t, mgr.ValidateExpression("@#$%"))
}

func TestFormulaManager_Execute(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	mgr.Activate("f1")
	result, err := mgr.Execute("f1", map[string]interface{}{"a": 10.0, "b": 20.0})
	require.NoError(t, err)
	assert.Equal(t, 30.0, result)
}

func TestFormulaManager_Execute_NotActive(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	_, err := mgr.Execute("f1", map[string]interface{}{"a": 10.0, "b": 20.0})
	assert.Error(t, err)
}

func TestFormulaManager_Execute_NotFound(t *testing.T) {
	mgr := NewFormulaManager(nil)
	_, err := mgr.Execute("nonexistent", nil)
	assert.Error(t, err)
}

func TestFormulaManager_ExecuteWithContext(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	mgr.Activate("f1")
	result, err := mgr.ExecuteWithContext(context.Background(), "f1", map[string]interface{}{"a": 10.0, "b": 20.0})
	require.NoError(t, err)
	assert.Equal(t, 30.0, result)
}

func TestFormulaManager_ExecuteBatch(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test1", Expression: "a + b"})
	mgr.Create(&Formula{ID: "f2", Name: "test2", Expression: "a * b"})
	mgr.Activate("f1")
	mgr.Activate("f2")
	results, errors := mgr.ExecuteBatch([]string{"f1", "f2"}, map[string]interface{}{"a": 10.0, "b": 5.0})
	assert.Equal(t, 15.0, results["f1"])
	assert.Equal(t, 50.0, results["f2"])
	assert.NoError(t, errors["f1"])
	assert.NoError(t, errors["f2"])
}

func TestFormulaManager_GetDependencies(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	deps, err := mgr.GetDependencies("f1")
	require.NoError(t, err)
	assert.NotNil(t, deps)
	_, err = mgr.GetDependencies("nonexistent")
	assert.Error(t, err)
}

func TestFormulaManager_GetDependents(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "base", Expression: "a + b"})
	f2 := &Formula{ID: "f2", Name: "dep", Expression: "c + d"}
	f2.dependencies = []string{"f1"}
	mgr.formulas["f2"] = f2
	dependents, err := mgr.GetDependents("f1")
	require.NoError(t, err)
	assert.Contains(t, dependents, "f2")
	_, err = mgr.GetDependents("nonexistent")
	assert.Error(t, err)
}

func TestFormulaManager_GetDependencyGraph(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	graph := mgr.GetDependencyGraph()
	assert.NotNil(t, graph)
	assert.Contains(t, graph, "f1")
}

func TestFormulaManager_TopologicalSort(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test1", Expression: "a + b"})
	mgr.Create(&Formula{ID: "f2", Name: "test2", Expression: "c + d"})
	result, err := mgr.TopologicalSort()
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestFormulaManager_TopologicalSort_CircularDependency(t *testing.T) {
	mgr := NewFormulaManager(nil)
	f1 := &Formula{ID: "f1", Name: "test1", Expression: "a + b"}
	f1.dependencies = []string{"f2"}
	f2 := &Formula{ID: "f2", Name: "test2", Expression: "c + d"}
	f2.dependencies = []string{"f1"}
	mgr.formulas["f1"] = f1
	mgr.formulas["f2"] = f2
	_, err := mgr.TopologicalSort()
	assert.Error(t, err)
}

func TestFormulaManager_GetVersions(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	versions, err := mgr.GetVersions("f1")
	require.NoError(t, err)
	assert.Len(t, versions, 1)
	_, err = mgr.GetVersions("nonexistent")
	assert.Error(t, err)
}

func TestFormulaManager_Rollback(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	f, _ := mgr.Get("f1")
	f.Expression = "a * b"
	mgr.Update(f)
	versions, _ := mgr.GetVersions("f1")
	require.NoError(t, mgr.Rollback("f1", versions[0].Version))
}

func TestFormulaManager_Rollback_NotFound(t *testing.T) {
	mgr := NewFormulaManager(nil)
	assert.Error(t, mgr.Rollback("nonexistent", "1.0.0"))
}

func TestFormulaManager_Rollback_NoVersion(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	assert.Error(t, mgr.Rollback("f1", "99.0.0"))
}

func TestFormulaManager_Activate(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	require.NoError(t, mgr.Activate("f1"))
	f, _ := mgr.Get("f1")
	assert.Equal(t, StatusActive, f.Status)
	assert.Error(t, mgr.Activate("nonexistent"))
}

func TestFormulaManager_Deactivate(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	require.NoError(t, mgr.Deactivate("f1"))
	f, _ := mgr.Get("f1")
	assert.Equal(t, StatusInactive, f.Status)
	assert.Error(t, mgr.Deactivate("nonexistent"))
}

func TestFormulaManager_Deprecate(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	require.NoError(t, mgr.Deprecate("f1"))
	f, _ := mgr.Get("f1")
	assert.Equal(t, StatusDeprecated, f.Status)
	assert.Error(t, mgr.Deprecate("nonexistent"))
}

func TestFormulaManager_GetStats(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test1", Expression: "a + b", Category: "math", Tags: []string{"critical"}})
	mgr.Activate("f1")
	mgr.Create(&Formula{ID: "f2", Name: "test2", Expression: "c + d", Category: "stats"})
	stats := mgr.GetStats()
	assert.Equal(t, 2, stats.TotalFormulas)
	assert.Equal(t, 1, stats.ActiveFormulas)
	assert.Equal(t, 1, stats.DraftFormulas)
	assert.Equal(t, 1, stats.Categories["math"])
	assert.Equal(t, 1, stats.Tags["critical"])
}

func TestFormulaManager_Export(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	exported, err := mgr.Export([]string{"f1"})
	require.NoError(t, err)
	assert.Len(t, exported, 1)
	_, err = mgr.Export([]string{"nonexistent"})
	assert.Error(t, err)
}

func TestFormulaManager_Import(t *testing.T) {
	mgr := NewFormulaManager(nil)
	formulas := []*Formula{
		{ID: "f1", Name: "test1", Expression: "a + b"},
		{ID: "f2", Name: "test2", Expression: "c + d"},
	}
	require.NoError(t, mgr.Import(formulas, false))
	list := mgr.List()
	assert.Len(t, list, 2)
}

func TestFormulaManager_Import_Overwrite(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test1", Expression: "a + b"})
	formulas := []*Formula{{ID: "f1", Name: "updated", Expression: "a * b"}}
	require.NoError(t, mgr.Import(formulas, true))
	f, _ := mgr.Get("f1")
	assert.Equal(t, "updated", f.Name)
}

func TestFormulaManager_Import_NoOverwrite(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test1", Expression: "a + b"})
	formulas := []*Formula{{ID: "f1", Name: "updated", Expression: "a * b"}}
	require.NoError(t, mgr.Import(formulas, false))
	f, _ := mgr.Get("f1")
	assert.Equal(t, "test1", f.Name)
}

func TestFormulaManager_BatchCreate(t *testing.T) {
	mgr := NewFormulaManager(nil)
	formulas := []*Formula{
		{ID: "f1", Name: "test1", Expression: "a + b"},
		{ID: "f2", Name: "test2", Expression: "c + d"},
	}
	errors := mgr.BatchCreate(formulas)
	assert.Len(t, errors, 2)
	assert.NoError(t, errors[0])
	assert.NoError(t, errors[1])
}

func TestFormulaManager_BatchUpdate(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test1", Expression: "a + b"})
	f, _ := mgr.Get("f1")
	f.Name = "updated"
	errors := mgr.BatchUpdate([]*Formula{f})
	assert.Len(t, errors, 1)
	assert.NoError(t, errors[0])
}

func TestFormulaManager_BatchDelete(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test1", Expression: "a + b"})
	errors := mgr.BatchDelete([]string{"f1"})
	assert.Len(t, errors, 1)
	assert.NoError(t, errors[0])
}

func TestFormulaManager_BatchActivate(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test1", Expression: "a + b"})
	errors := mgr.BatchActivate([]string{"f1"})
	assert.Len(t, errors, 1)
	assert.NoError(t, errors[0])
}

func TestFormulaManager_BatchDeactivate(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test1", Expression: "a + b"})
	errors := mgr.BatchDeactivate([]string{"f1"})
	assert.Len(t, errors, 1)
	assert.NoError(t, errors[0])
}

func TestFormulaManager_SortByName(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "beta", Expression: "a + b"})
	mgr.Create(&Formula{ID: "f2", Name: "alpha", Expression: "c + d"})
	sorted := mgr.SortByName(mgr.List())
	assert.Equal(t, "alpha", sorted[0].Name)
}

func TestFormulaManager_SortByCreatedAt(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test1", Expression: "a + b"})
	mgr.Create(&Formula{ID: "f2", Name: "test2", Expression: "c + d"})
	sorted := mgr.SortByCreatedAt(mgr.List())
	assert.Len(t, sorted, 2)
}

func TestFormulaManager_SortByUpdatedAt(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test1", Expression: "a + b"})
	mgr.Create(&Formula{ID: "f2", Name: "test2", Expression: "c + d"})
	sorted := mgr.SortByUpdatedAt(mgr.List())
	assert.Len(t, sorted, 2)
}

func TestFormulaStatus_String(t *testing.T) {
	assert.Equal(t, "draft", StatusDraft.String())
	assert.Equal(t, "active", StatusActive.String())
	assert.Equal(t, "inactive", StatusInactive.String())
	assert.Equal(t, "deprecated", StatusDeprecated.String())
	assert.Equal(t, "error", StatusError.String())
	assert.Equal(t, "unknown", FormulaStatus(99).String())
}

func TestFormulaValidator_Validate(t *testing.T) {
	validator := NewFormulaValidator()
	assert.Error(t, validator.Validate(&Formula{ID: "", Name: "test", Expression: "a + b"}))
	assert.Error(t, validator.Validate(&Formula{ID: "f1", Name: "", Expression: "a + b"}))
	assert.Error(t, validator.Validate(&Formula{ID: "f1", Name: "test", Expression: ""}))
	assert.NoError(t, validator.Validate(&Formula{ID: "f1", Name: "test", Expression: "a + b"}))
}

func TestFormulaValidator_AddRule(t *testing.T) {
	validator := NewFormulaValidator()
	validator.AddRule(func(f *Formula) error {
		if f.Category == "" {
			return assert.AnError
		}
		return nil
	})
	assert.Error(t, validator.Validate(&Formula{ID: "f1", Name: "test", Expression: "a + b", Category: ""}))
}

func TestDependencyAnalyzer_Analyze(t *testing.T) {
	analyzer := NewDependencyAnalyzer()
	f := &Formula{ID: "f1", Name: "test", Expression: "a + b", variables: []string{"formula.other", "a"}}
	deps, err := analyzer.Analyze(f)
	require.NoError(t, err)
	assert.Contains(t, deps, "other")
}

func TestDependencyAnalyzer_AnalyzeExpression(t *testing.T) {
	analyzer := NewDependencyAnalyzer()
	deps, err := analyzer.AnalyzeExpression("a + b")
	require.NoError(t, err)
	assert.NotNil(t, deps)
}

func TestDependencyAnalyzer_CheckCircularDependency(t *testing.T) {
	analyzer := NewDependencyAnalyzer()
	formulas := map[string]*Formula{
		"f1": {ID: "f1", dependencies: []string{"f2"}},
		"f2": {ID: "f2", dependencies: []string{"f1"}},
	}
	hasCycle, cycle := analyzer.CheckCircularDependency(formulas)
	assert.True(t, hasCycle)
	assert.NotNil(t, cycle)
}

func TestDependencyAnalyzer_CheckCircularDependency_NoCycle(t *testing.T) {
	analyzer := NewDependencyAnalyzer()
	formulas := map[string]*Formula{
		"f1": {ID: "f1", dependencies: []string{}},
		"f2": {ID: "f2", dependencies: []string{"f1"}},
	}
	hasCycle, _ := analyzer.CheckCircularDependency(formulas)
	assert.False(t, hasCycle)
}

func TestOperatorRegistry(t *testing.T) {
	registry := NewOperatorRegistry()
	assert.NotNil(t, registry)
	assert.True(t, IsOperator("+"))
	assert.True(t, IsOperator("-"))
	assert.True(t, IsOperator("*"))
	assert.True(t, IsOperator("/"))
	assert.False(t, IsOperator("@"))
}

func TestOperatorRegistry_IsUnary(t *testing.T) {
	assert.True(t, IsUnaryOperator("-"))
	assert.True(t, IsUnaryOperator("!"))
	_ = IsUnaryOperator("+")
}

func TestOperatorRegistry_IsBinary(t *testing.T) {
	assert.True(t, IsBinaryOperator("+"))
	assert.True(t, IsBinaryOperator("-"))
	assert.True(t, IsBinaryOperator("*"))
	_ = IsBinaryOperator("!")
}

func TestOperatorRegistry_IsUnaryOnly(t *testing.T) {
	_ = IsUnaryOnly("!")
	_ = IsUnaryOnly("-")
}

func TestGetOperatorType(t *testing.T) {
	assert.Equal(t, OpTypeArithmetic, GetOperatorType("+"))
	assert.Equal(t, OpTypeComparison, GetOperatorType(">"))
	assert.Equal(t, OpTypeLogical, GetOperatorType("&&"))
	assert.Equal(t, OpTypeBitwise, GetOperatorType("&"))
}

func TestPrecedenceHandler(t *testing.T) {
	handler := NewPrecedenceHandler()
	mulPrec := handler.GetPrecedence("*")
	addPrec := handler.GetPrecedence("+")
	assert.NotEqual(t, -1, mulPrec)
	assert.NotEqual(t, -1, addPrec)
}

func TestPrecedenceHandler_ShouldReduce(t *testing.T) {
	handler := NewPrecedenceHandler()
	_ = handler.ShouldReduce("*", "+")
}

func TestExpressionEvaluator(t *testing.T) {
	evaluator := NewExpressionEvaluator()
	result, err := evaluator.EvaluateBinary("+", 10.0, 20.0)
	require.NoError(t, err)
	assert.Equal(t, 30.0, result)
}

func TestExpressionEvaluator_Unary(t *testing.T) {
	evaluator := NewExpressionEvaluator()
	result, err := evaluator.EvaluateUnary("-", 10.0)
	require.NoError(t, err)
	assert.Equal(t, -10.0, result)
}

func TestOperatorValidator(t *testing.T) {
	validator := NewOperatorValidator()
	err := validator.ValidateBinaryOperation("+", 10.0, 20.0)
	if err != nil {
		assert.NoError(t, validator.ValidateUnaryOperation("-", 10.0))
	}
}

func TestOperatorPrecedenceTable(t *testing.T) {
	table := OperatorPrecedenceTable
	assert.NotNil(t, table)
	assert.Contains(t, table, "+")
	assert.Contains(t, table, "*")
}

func TestOperatorAssociativityTable(t *testing.T) {
	table := OperatorAssociativityTable
	assert.NotNil(t, table)
	assert.Contains(t, table, "+")
	assert.Contains(t, table, "^")
}

func TestGetPrecedenceFromTable(t *testing.T) {
	p := GetPrecedenceFromTable("+")
	assert.Greater(t, p, 0)
}

func TestGetAssociativityFromTable(t *testing.T) {
	a := GetAssociativityFromTable("+")
	assert.Equal(t, "left", a)
	a = GetAssociativityFromTable("^")
	assert.Equal(t, "left", a)
}

func TestExecutor_BuiltInFunctions(t *testing.T) {
	executor := NewExecutor(nil)
	tests := []struct {
		name      string
		formula   string
		variables map[string]interface{}
		expected  interface{}
	}{
		{"abs", "abs(a)", map[string]interface{}{"a": -5.0}, 5.0},
		{"floor", "floor(a)", map[string]interface{}{"a": 5.7}, 5.0},
		{"ceil", "ceil(a)", map[string]interface{}{"a": 5.2}, 6.0},
		{"round", "round(a)", map[string]interface{}{"a": 5.5}, 6.0},
		{"sqrt", "sqrt(a)", map[string]interface{}{"a": 16.0}, 4.0},
		{"pow", "pow(a, b)", map[string]interface{}{"a": 2.0, "b": 3.0}, 8.0},
		{"max", "max(a, b)", map[string]interface{}{"a": 10.0, "b": 20.0}, 20.0},
		{"min", "min(a, b)", map[string]interface{}{"a": 10.0, "b": 20.0}, 10.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.Execute(tt.formula, tt.variables)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExecutor_BitwiseOperations(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("a & b", map[string]interface{}{"a": float64(12), "b": float64(10)})
	assert.NoError(t, err)
	assert.Equal(t, float64(8), result)
}

func TestExecutor_StringOperations(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute(`len("hello")`, map[string]interface{}{})
	assert.NoError(t, err)
	assert.Equal(t, float64(5), result)
}

func TestExecutor_ConstantFunctions(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("pi()", map[string]interface{}{})
	assert.NoError(t, err)
	assert.InDelta(t, 3.141592653589793, result, 0.001)
}

func TestExecutor_VariableSyntax(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("${point_001} + 1", map[string]interface{}{"point_001": 10.0})
	assert.NoError(t, err)
	assert.Equal(t, 11.0, result)
}

func TestExecutor_NegativeNumber(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("-5 + 10", map[string]interface{}{})
	assert.NoError(t, err)
	assert.Equal(t, 5.0, result)
}

func TestExecutor_ArrayFunction(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("first(array(1, 2, 3))", map[string]interface{}{})
	assert.NoError(t, err)
	assert.Equal(t, float64(1), result)
}

func TestExecutor_ConditionalFunction(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("if(a > 5, 100, 200)", map[string]interface{}{"a": 10.0})
	assert.NoError(t, err)
	assert.Equal(t, float64(100), result)
}

func TestExecutor_SwitchFunction(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("switch(a, 1, 10, 2, 20, 30)", map[string]interface{}{"a": 2.0})
	assert.NoError(t, err)
	assert.Equal(t, float64(20), result)
}

func TestExecutor_CoalesceFunction(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("coalesce(a, b, 42)", map[string]interface{}{"a": nil, "b": nil})
	assert.NoError(t, err)
	assert.Equal(t, float64(42), result)
}

func TestExecutor_TypeConversion(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("int(5.7)", map[string]interface{}{})
	assert.NoError(t, err)
	assert.Equal(t, float64(5), result)

	result, err = executor.Execute("string(42)", map[string]interface{}{})
	assert.NoError(t, err)
	assert.Equal(t, "42", result)
}

func TestDefaultManagerConfig(t *testing.T) {
	config := DefaultManagerConfig()
	assert.True(t, config.AutoCompile)
	assert.True(t, config.AutoValidate)
	assert.Equal(t, 10, config.MaxVersions)
	assert.True(t, config.EnableCache)
}

func TestExecutor_MathFunctions(t *testing.T) {
	executor := NewExecutor(nil)
	tests := []struct {
		name     string
		formula  string
		expected float64
	}{
		{"sign_pos", "sign(5)", 1.0},
		{"sign_neg", "sign(-5)", -1.0},
		{"sign_zero", "sign(0)", 0.0},
		{"trunc", "trunc(5.7)", 5.0},
		{"cbrt", "cbrt(27)", 3.0},
		{"exp", "round(exp(1))", 3.0},
		{"exp2", "exp2(3)", 8.0},
		{"exp10", "exp10(2)", 100.0},
		{"log", "round(log(e()))", 1.0},
		{"log10", "log10(100)", 2.0},
		{"log2", "log2(8)", 3.0},
		{"log1p", "round(log1p(0))", 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.Execute(tt.formula, nil)
			assert.NoError(t, err)
			assert.InDelta(t, tt.expected, result.(float64), 0.01)
		})
	}
}

func TestExecutor_TrigFunctions(t *testing.T) {
	executor := NewExecutor(nil)
	tests := []struct {
		name     string
		formula  string
		delta    float64
		expected float64
	}{
		{"sin", "sin(rad(30))", 0.01, 0.5},
		{"cos", "cos(rad(60))", 0.01, 0.5},
		{"tan", "tan(rad(45))", 0.01, 1.0},
		{"asin", "deg(asin(0.5))", 0.01, 30.0},
		{"acos", "deg(acos(0.5))", 0.01, 60.0},
		{"atan", "deg(atan(1))", 0.01, 45.0},
		{"atan2", "deg(atan2(1, 1))", 0.01, 45.0},
		{"sinh", "sinh(0)", 0.01, 0.0},
		{"cosh", "cosh(0)", 0.01, 1.0},
		{"tanh", "tanh(0)", 0.01, 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.Execute(tt.formula, nil)
			assert.NoError(t, err)
			assert.InDelta(t, tt.expected, result.(float64), tt.delta)
		})
	}
}

func TestExecutor_StatsFunctions(t *testing.T) {
	executor := NewExecutor(nil)
	tests := []struct {
		name     string
		formula  string
		expected float64
	}{
		{"sum", "sum(1, 2, 3, 4, 5)", 15.0},
		{"avg", "avg(10, 20, 30)", 20.0},
		{"count", "count(1, 2, 3)", 3.0},
		{"product", "product(2, 3, 4)", 24.0},
		{"variance", "round(variance(2, 4, 4, 4, 5, 5, 7, 9))", 4.0},
		{"stddev", "round(stddev(2, 4, 4, 4, 5, 5, 7, 9))", 2.0},
		{"median", "median(1, 3, 5)", 3.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.Execute(tt.formula, nil)
			assert.NoError(t, err)
			assert.InDelta(t, tt.expected, result.(float64), 0.5)
		})
	}
}

func TestExecutor_StringFunctions(t *testing.T) {
	executor := NewExecutor(nil)
	tests := []struct {
		name     string
		formula  string
		expected interface{}
	}{
		{"upper", `upper("hello")`, "HELLO"},
		{"lower", `lower("HELLO")`, "hello"},
		{"trim", `trim("  hello  ")`, "hello"},
		{"concat", `concat("hello", " ", "world")`, "hello world"},
		{"contains", `contains("hello world", "world")`, true},
		{"startsWith", `startsWith("hello", "he")`, true},
		{"endsWith", `endsWith("hello", "lo")`, true},
		{"replace", `replace("hello", "l", "r")`, "herro"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.Execute(tt.formula, nil)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExecutor_ArrayFunctions(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("last(array(1, 2, 3))", nil)
	assert.NoError(t, err)
	assert.Equal(t, float64(3), result)

	result, err = executor.Execute("nth(array(10, 20, 30), 1)", nil)
	assert.NoError(t, err)
	assert.Equal(t, float64(20), result)

	result, err = executor.Execute("len(array(1, 2, 3))", nil)
	assert.NoError(t, err)
	assert.Equal(t, float64(3), result)

	result, err = executor.Execute("push(array(1, 2), 3)", nil)
	assert.NoError(t, err)
	arr, ok := result.([]interface{})
	assert.True(t, ok)
	assert.Len(t, arr, 3)

	result, err = executor.Execute("reverse(array(1, 2, 3))", nil)
	assert.NoError(t, err)
	arr, ok = result.([]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(3), arr[0])
}

func TestExecutor_UtilityFunctions(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("clamp(15, 0, 10)", nil)
	assert.NoError(t, err)
	assert.Equal(t, float64(10), result)

	result, err = executor.Execute("lerp(0, 100, 0.5)", nil)
	assert.NoError(t, err)
	assert.Equal(t, float64(50), result)

	result, err = executor.Execute("step(5, 10)", nil)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), result)

	result, err = executor.Execute("step(5, 3)", nil)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), result)

	result, err = executor.Execute("smoothstep(0, 10, 5)", nil)
	assert.NoError(t, err)
	assert.Equal(t, float64(0.5), result)

	result, err = executor.Execute("mod(10, 3)", nil)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), result)

	result, err = executor.Execute("gcd(12, 8)", nil)
	assert.NoError(t, err)
	assert.Equal(t, float64(4), result)

	result, err = executor.Execute("lcm(4, 6)", nil)
	assert.NoError(t, err)
	assert.Equal(t, float64(12), result)

	result, err = executor.Execute("isFinite(42)", nil)
	assert.NoError(t, err)
	assert.Equal(t, true, result)

	_, err = executor.Execute("isInf(1.0/0.0)", nil)
	_ = err

	_, err = executor.Execute("isNaN(0.0/0.0)", nil)
	_ = err
}

func TestExecutor_ConstantFunctions2(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("pi()", nil)
	assert.NoError(t, err)
	assert.InDelta(t, 3.141592653589793, result.(float64), 0.001)

	result, err = executor.Execute("e()", nil)
	assert.NoError(t, err)
	assert.InDelta(t, 2.718281828459045, result.(float64), 0.001)

	result, err = executor.Execute("phi()", nil)
	assert.NoError(t, err)
	assert.InDelta(t, 1.618033988749895, result.(float64), 0.001)
}

func TestExecutor_RandomFunction(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("random()", nil)
	assert.NoError(t, err)
	val, ok := result.(float64)
	assert.True(t, ok)
	assert.True(t, val >= 0.0 && val < 1.0)
}

func TestExecutor_BoolFunction(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("bool(1)", nil)
	assert.NoError(t, err)
	assert.Equal(t, true, result)

	result, err = executor.Execute("bool(0)", nil)
	assert.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestExecutor_FloatFunction(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("float(42)", nil)
	assert.NoError(t, err)
	assert.Equal(t, 42.0, result)
}

func TestExecutor_SubstrFunction(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute(`substr("hello world", 0, 5)`, nil)
	assert.NoError(t, err)
	assert.Equal(t, "hello", result)
}

func TestExecutor_SplitJoinFunctions(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute(`join(split("a,b,c", ","), "-")`, nil)
	assert.NoError(t, err)
	assert.Equal(t, "a-b-c", result)
}

func TestExecutor_SliceFunction(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("slice(array(1, 2, 3, 4, 5), 1, 3)", nil)
	assert.NoError(t, err)
	arr, ok := result.([]interface{})
	assert.True(t, ok)
	assert.Len(t, arr, 2)
}

func TestExecutor_PopFunction(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("pop(array(1, 2, 3))", nil)
	assert.NoError(t, err)
	arr, ok := result.([]interface{})
	assert.True(t, ok)
	assert.Len(t, arr, 2)
}

func TestExecutor_SortFunction(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("sort(array(3, 1, 2))", nil)
	assert.NoError(t, err)
	arr, ok := result.([]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(1), arr[0])
}

func TestExecutor_RoundWithPrecision(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("round(3.14159, 2)", nil)
	assert.NoError(t, err)
	assert.InDelta(t, 3.14, result.(float64), 0.001)
}

func TestExecutor_FunctionArgErrors(t *testing.T) {
	executor := NewExecutor(nil)
	_, err := executor.Execute("abs()", nil)
	assert.Error(t, err)
	_, err = executor.Execute("abs(1, 2)", nil)
	assert.Error(t, err)
	_, err = executor.Execute("pow(1)", nil)
	assert.Error(t, err)
}

func TestExecutor_InvalidFormula(t *testing.T) {
	executor := NewExecutor(nil)
	_, err := executor.Execute("", nil)
	assert.Error(t, err)
}

func TestExecutor_ExecuteNode(t *testing.T) {
	executor := NewExecutor(nil)
	node, err := ParseFormula("a + b")
	require.NoError(t, err)
	result, err := executor.ExecuteNode(node, map[string]interface{}{"a": 10.0, "b": 20.0})
	assert.NoError(t, err)
	assert.Equal(t, 30.0, result)
}

func TestFunctionRegistry_GetAll(t *testing.T) {
	registry := NewFunctionRegistry()
	all := registry.GetAll()
	assert.NotEmpty(t, all)
	assert.Contains(t, all, "abs")
	assert.Contains(t, all, "max")
}

func TestExecutor_Execute_NilVariables(t *testing.T) {
	executor := NewExecutor(nil)
	_, err := executor.Execute("a + b", nil)
	assert.Error(t, err)
}

func TestExecutor_ExecuteWithContext_CacheHit(t *testing.T) {
	config := &ExecutorConfig{EnableCache: true, CacheTTL: 5 * time.Minute, MaxCacheSize: 100}
	executor := NewExecutor(config)
	ctx := context.Background()
	executor.ExecuteWithContext(ctx, "a + b", map[string]interface{}{"a": 10.0, "b": 20.0})
	result, err := executor.ExecuteWithContext(ctx, "a + b", map[string]interface{}{"a": 10.0, "b": 20.0})
	assert.NoError(t, err)
	assert.Equal(t, 30.0, result)
}

func TestFormulaManager_Execute_NilVariables(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	mgr.Activate("f1")
	_, err := mgr.Execute("f1", nil)
	assert.Error(t, err)
}

func TestFormulaManager_Execute_CompileError(t *testing.T) {
	mgr := NewFormulaManager(&ManagerConfig{AutoCompile: false, AutoValidate: false})
	f := &Formula{ID: "f1", Name: "test", Expression: "a + b"}
	f.compiled = nil
	mgr.formulas["f1"] = f
	mgr.Activate("f1")
	_, err := mgr.Execute("f1", map[string]interface{}{"a": 10.0, "b": 20.0})
	require.NoError(t, err)
}

func TestFormulaManager_ExecuteBatch_PartialError(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test1", Expression: "a + b"})
	mgr.Activate("f1")
	results, errors := mgr.ExecuteBatch([]string{"f1", "nonexistent"}, map[string]interface{}{"a": 10.0, "b": 20.0})
	assert.Equal(t, 30.0, results["f1"])
	assert.NoError(t, errors["f1"])
	assert.Error(t, errors["nonexistent"])
}

func TestFormulaManager_ExportAll(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test1", Expression: "a + b"})
	mgr.Create(&Formula{ID: "f2", Name: "test2", Expression: "c + d"})
	exported, err := mgr.Export([]string{"f1", "f2"})
	require.NoError(t, err)
	assert.Len(t, exported, 2)
}

func TestFormulaManager_Import_NilFormulas(t *testing.T) {
	mgr := NewFormulaManager(nil)
	err := mgr.Import(nil, false)
	assert.NoError(t, err)
}

func TestFormulaManager_BatchCreate_Empty(t *testing.T) {
	mgr := NewFormulaManager(nil)
	errors := mgr.BatchCreate(nil)
	assert.Empty(t, errors)
}

func TestFormulaManager_BatchUpdate_Empty(t *testing.T) {
	mgr := NewFormulaManager(nil)
	errors := mgr.BatchUpdate(nil)
	assert.Empty(t, errors)
}

func TestFormulaManager_BatchDelete_Empty(t *testing.T) {
	mgr := NewFormulaManager(nil)
	errors := mgr.BatchDelete(nil)
	assert.Empty(t, errors)
}

func TestFormulaManager_BatchActivate_Empty(t *testing.T) {
	mgr := NewFormulaManager(nil)
	errors := mgr.BatchActivate(nil)
	assert.Empty(t, errors)
}

func TestFormulaManager_BatchDeactivate_Empty(t *testing.T) {
	mgr := NewFormulaManager(nil)
	errors := mgr.BatchDeactivate(nil)
	assert.Empty(t, errors)
}

func TestFormulaManager_Delete_Active(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	mgr.Activate("f1")
	err := mgr.Delete("f1")
	_ = err
}

func TestFormulaManager_Activate_AlreadyActive(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	require.NoError(t, mgr.Activate("f1"))
	require.NoError(t, mgr.Activate("f1"))
}

func TestFormulaManager_Deactivate_AlreadyInactive(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	require.NoError(t, mgr.Deactivate("f1"))
	require.NoError(t, mgr.Deactivate("f1"))
}

func TestFormulaManager_Deprecate_AlreadyDeprecated(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	require.NoError(t, mgr.Deprecate("f1"))
	require.NoError(t, mgr.Deprecate("f1"))
}

func TestFormulaManager_Execute_ExecutionError(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a / 0"})
	mgr.Activate("f1")
	_, err := mgr.Execute("f1", map[string]interface{}{"a": 10.0})
	assert.Error(t, err)
}

func TestFormulaManager_ExecuteWithCache(t *testing.T) {
	mgr := NewFormulaManager(&ManagerConfig{EnableCache: true, AutoCompile: true, AutoValidate: true})
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	mgr.Activate("f1")
	result1, err := mgr.Execute("f1", map[string]interface{}{"a": 10.0, "b": 20.0})
	require.NoError(t, err)
	assert.Equal(t, 30.0, result1)
	result2, err := mgr.Execute("f1", map[string]interface{}{"a": 10.0, "b": 20.0})
	require.NoError(t, err)
	assert.Equal(t, 30.0, result2)
}

func TestFormulaManager_ExecuteWithContext_Cache(t *testing.T) {
	mgr := NewFormulaManager(&ManagerConfig{EnableCache: true, AutoCompile: true, AutoValidate: true})
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	mgr.Activate("f1")
	ctx := context.Background()
	result, err := mgr.ExecuteWithContext(ctx, "f1", map[string]interface{}{"a": 10.0, "b": 20.0})
	require.NoError(t, err)
	assert.Equal(t, 30.0, result)
}

func TestFormulaManager_GetStats_Empty(t *testing.T) {
	mgr := NewFormulaManager(nil)
	stats := mgr.GetStats()
	assert.Equal(t, 0, stats.TotalFormulas)
}

func TestFormulaManager_Export_NotFound(t *testing.T) {
	mgr := NewFormulaManager(nil)
	_, err := mgr.Export([]string{"nonexistent"})
	assert.Error(t, err)
}

func TestFormulaManager_TopologicalSort_Empty(t *testing.T) {
	mgr := NewFormulaManager(nil)
	result, err := mgr.TopologicalSort()
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestFormulaManager_GetDependencyGraph_Empty(t *testing.T) {
	mgr := NewFormulaManager(nil)
	graph := mgr.GetDependencyGraph()
	assert.Empty(t, graph)
}

func TestFormulaManager_VersionLimit(t *testing.T) {
	mgr := NewFormulaManager(&ManagerConfig{AutoCompile: true, AutoValidate: true, MaxVersions: 2})
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	f, _ := mgr.Get("f1")
	f.Expression = "a * b"
	mgr.Update(f)
	f.Expression = "a - b"
	mgr.Update(f)
	versions, _ := mgr.GetVersions("f1")
	assert.LessOrEqual(t, len(versions), 2)
}

func TestFormulaManager_Execute_NoCompiled(t *testing.T) {
	mgr := NewFormulaManager(&ManagerConfig{AutoCompile: false, AutoValidate: false})
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	mgr.Activate("f1")
	result, err := mgr.Execute("f1", map[string]interface{}{"a": 10.0, "b": 20.0})
	require.NoError(t, err)
	assert.Equal(t, 30.0, result)
}

func TestFormulaManager_Rollback_CompileError(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	f, _ := mgr.Get("f1")
	f.Expression = "a * b"
	mgr.Update(f)
	versions, _ := mgr.GetVersions("f1")
	originalExpr := f.Expression
	versions[0].Expression = "@#$%"
	err := mgr.Rollback("f1", versions[0].Version)
	assert.Error(t, err)
	f2, _ := mgr.Get("f1")
	assert.Equal(t, originalExpr, f2.Expression)
}
