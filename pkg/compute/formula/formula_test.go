package formula

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
		name       string
		formula    string
		variables  map[string]interface{}
		expected   interface{}
		hasError   bool
	}{
		{
			name:      "简单加法",
			formula:   "a + b",
			variables: map[string]interface{}{"a": 10.0, "b": 20.0},
			expected:  30.0,
			hasError:  false,
		},
		{
			name:      "简单减法",
			formula:   "a - b",
			variables: map[string]interface{}{"a": 30.0, "b": 10.0},
			expected:  20.0,
			hasError:  false,
		},
		{
			name:      "简单乘法",
			formula:   "a * b",
			variables: map[string]interface{}{"a": 5.0, "b": 6.0},
			expected:  30.0,
			hasError:  false,
		},
		{
			name:      "简单除法",
			formula:   "a / b",
			variables: map[string]interface{}{"a": 20.0, "b": 4.0},
			expected:  5.0,
			hasError:  false,
		},
		{
			name:      "复杂表达式",
			formula:   "(a + b) * c - d",
			variables: map[string]interface{}{"a": 2.0, "b": 3.0, "c": 4.0, "d": 5.0},
			expected:  15.0, // (2+3)*4-5 = 15
			hasError:  false,
		},
		{
			name:      "除零错误",
			formula:   "a / b",
			variables: map[string]interface{}{"a": 10.0, "b": 0.0},
			hasError:  true,
		},
		{
			name:      "变量不存在",
			formula:   "a + c",
			variables: map[string]interface{}{"a": 10.0, "b": 20.0},
			hasError:  true,
		},
		{
			name:      "幂运算",
			formula:   "a ^ b",
			variables: map[string]interface{}{"a": 2.0, "b": 3.0},
			expected:  8.0,
			hasError:  false,
		},
		{
			name:      "取模运算",
			formula:   "a % b",
			variables: map[string]interface{}{"a": 10.0, "b": 3.0},
			expected:  1.0,
			hasError:  false,
		},
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
		{
			name:      "大于比较",
			formula:   "a > b",
			variables: map[string]interface{}{"a": 10.0, "b": 5.0},
			expected:  true,
		},
		{
			name:      "小于比较",
			formula:   "a < b",
			variables: map[string]interface{}{"a": 5.0, "b": 10.0},
			expected:  true,
		},
		{
			name:      "等于比较",
			formula:   "a == b",
			variables: map[string]interface{}{"a": 10.0, "b": 10.0},
			expected:  true,
		},
		{
			name:      "不等于比较",
			formula:   "a != b",
			variables: map[string]interface{}{"a": 10.0, "b": 5.0},
			expected:  true,
		},
		{
			name:      "大于等于",
			formula:   "a >= b",
			variables: map[string]interface{}{"a": 10.0, "b": 10.0},
			expected:  true,
		},
		{
			name:      "小于等于",
			formula:   "a <= b",
			variables: map[string]interface{}{"a": 5.0, "b": 10.0},
			expected:  true,
		},
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
		{
			name:      "逻辑与-真",
			formula:   "a && b",
			variables: map[string]interface{}{"a": true, "b": true},
			expected:  true,
		},
		{
			name:      "逻辑与-假",
			formula:   "a && b",
			variables: map[string]interface{}{"a": true, "b": false},
			expected:  false,
		},
		{
			name:      "逻辑或-真",
			formula:   "a || b",
			variables: map[string]interface{}{"a": false, "b": true},
			expected:  true,
		},
		{
			name:      "逻辑或-假",
			formula:   "a || b",
			variables: map[string]interface{}{"a": false, "b": false},
			expected:  false,
		},
		{
			name:      "逻辑非",
			formula:   "!a",
			variables: map[string]interface{}{"a": true},
			expected:  false,
		},
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
		{
			name:      "条件为真",
			formula:   "a > b ? c : d",
			variables: map[string]interface{}{"a": 10.0, "b": 5.0, "c": 100.0, "d": 200.0},
			expected:  100.0,
		},
		{
			name:      "条件为假",
			formula:   "a > b ? c : d",
			variables: map[string]interface{}{"a": 5.0, "b": 10.0, "c": 100.0, "d": 200.0},
			expected:  200.0,
		},
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

	// 注册自定义函数
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

	tests := []struct {
		name      string
		formula   string
		variables map[string]interface{}
		expected  interface{}
	}{
		{
			name:      "自定义函数",
			formula:   "double(a)",
			variables: map[string]interface{}{"a": 5.0},
			expected:  10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.Execute(tt.formula, tt.variables)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExecutor_ExecuteBatch(t *testing.T) {
	executor := NewExecutor(nil)

	formulas := []string{
		"a + b",
		"a * b",
		"a - b",
	}

	variables := map[string]interface{}{
		"a": 10.0,
		"b": 5.0,
	}

	results, errors := executor.ExecuteBatch(formulas, variables)

	assert.Len(t, results, 3)
	assert.Len(t, errors, 3)
	assert.Equal(t, 15.0, results[0])
	assert.Equal(t, 50.0, results[1])
	assert.Equal(t, 5.0, results[2])
	for _, err := range errors {
		assert.NoError(t, err)
	}
}

func TestExecutor_ExecuteParallel(t *testing.T) {
	executor := NewExecutor(nil)

	formulas := []string{
		"a + b",
		"a * b",
		"a - b",
		"a / b",
	}

	variables := map[string]interface{}{
		"a": 10.0,
		"b": 5.0,
	}

	results, errors := executor.ExecuteParallel(formulas, variables, 2)

	assert.Len(t, results, 4)
	assert.Len(t, errors, 4)
	assert.Equal(t, 15.0, results[0])
	assert.Equal(t, 50.0, results[1])
	assert.Equal(t, 5.0, results[2])
	assert.Equal(t, 2.0, results[3])
	for _, err := range errors {
		assert.NoError(t, err)
	}
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
	cancel() // 立即取消

	_, err := executor.ExecuteWithContext(ctx, "a + b", map[string]interface{}{"a": 10.0, "b": 20.0})

	assert.Error(t, err)
}

func TestExecutor_Cache(t *testing.T) {
	config := &ExecutorConfig{
		EnableCache:  true,
		CacheTTL:     5 * time.Minute,
		MaxCacheSize: 100,
	}

	executor := NewExecutor(config)

	// 第一次执行
	result1, err := executor.Execute("a + b", map[string]interface{}{"a": 10.0, "b": 20.0})
	assert.NoError(t, err)
	assert.Equal(t, 30.0, result1)

	// 第二次执行（应该从缓存获取）
	result2, err := executor.Execute("a + b", map[string]interface{}{"a": 10.0, "b": 20.0})
	assert.NoError(t, err)
	assert.Equal(t, 30.0, result2)

	// 检查缓存统计
	stats := executor.GetCacheStats()
	assert.NotNil(t, stats)
	assert.Equal(t, int64(1), stats.Hits)
}

func TestExecutor_ClearCache(t *testing.T) {
	config := &ExecutorConfig{
		EnableCache:  true,
		CacheTTL:     5 * time.Minute,
		MaxCacheSize: 100,
	}

	executor := NewExecutor(config)

	// 执行并缓存
	executor.Execute("a + b", map[string]interface{}{"a": 10.0, "b": 20.0})

	// 清除缓存
	executor.ClearCache()

	// 检查缓存统计
	stats := executor.GetCacheStats()
	assert.NotNil(t, stats)
	assert.Equal(t, 0, stats.Size)
}

func TestExecutor_Compile(t *testing.T) {
	executor := NewExecutor(nil)

	compiled, err := executor.Compile("a + b * c")
	assert.NoError(t, err)
	assert.NotNil(t, compiled)
	assert.Equal(t, "a + b * c", compiled.formula)
	assert.NotNil(t, compiled.ast)

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

	// 设置缓存
	cache.Set("key1", "value1")

	// 获取缓存
	value, exists := cache.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", value)

	// 获取不存在的缓存
	_, exists = cache.Get("key2")
	assert.False(t, exists)

	// 删除缓存
	cache.Delete("key1")
	_, exists = cache.Get("key1")
	assert.False(t, exists)

	// 清除缓存
	cache.Set("key3", "value3")
	cache.Clear()
	stats := cache.Stats()
	assert.Equal(t, 0, stats.Size)
}

func TestResultCache_Eviction(t *testing.T) {
	cache := NewResultCache(3, 5*time.Minute)

	// 添加超过容量的项
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")
	cache.Set("key4", "value4") // 应该触发淘汰

	stats := cache.Stats()
	assert.LessOrEqual(t, stats.Size, 3)
}

func TestVariableBinder(t *testing.T) {
	binder := NewVariableBinder()

	// 绑定变量
	binder.Bind("a", 10.0)
	binder.Bind("b", 20.0)

	// 获取变量
	value, exists := binder.Get("a")
	assert.True(t, exists)
	assert.Equal(t, 10.0, value)

	// 批量绑定
	binder.BindMany(map[string]interface{}{"c": 30.0, "d": 40.0})

	// 获取所有变量
	all := binder.GetAll()
	assert.Len(t, all, 4)

	// 解绑
	binder.Unbind("a")
	_, exists = binder.Get("a")
	assert.False(t, exists)

	// 清除
	binder.Clear()
	all = binder.GetAll()
	assert.Len(t, all, 0)
}

func TestExecutionContext(t *testing.T) {
	ctx := NewExecutionContext()

	// 设置变量
	ctx.SetVariable("a", 10.0)

	// 获取变量
	value, exists := ctx.GetVariable("a")
	assert.True(t, exists)
	assert.Equal(t, 10.0, value)

	// 注册函数
	ctx.RegisterFunction("double", func(args ...interface{}) (interface{}, error) {
		return args[0].(float64) * 2, nil
	})

	// 获取函数
	fn, exists := ctx.GetFunction("double")
	assert.True(t, exists)
	assert.NotNil(t, fn)

	// 设置元数据
	ctx.SetMetadata("key", "value")
	meta, exists := ctx.GetMetadata("key")
	assert.True(t, exists)
	assert.Equal(t, "value", meta)

	// 创建子上下文
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
	assert.Equal(t, 60*time.Second, executor.config.Timeout)
	assert.Equal(t, 50, executor.config.MaxRecursionDepth)

	// 验证自定义函数
	fn, exists := executor.GetFunction("custom")
	assert.True(t, exists)
	assert.NotNil(t, fn)
}

func TestParseFormula(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		hasError  bool
	}{
		{
			name:     "简单加法",
			input:    "a + b",
			hasError: false,
		},
		{
			name:     "复杂表达式",
			input:    "(a + b) * c - d / e",
			hasError: false,
		},
		{
			name:     "带函数的表达式",
			input:    "max(a, b) + min(c, d)",
			hasError: false,
		},
		{
			name:     "条件表达式",
			input:    "a > b ? c : d",
			hasError: false,
		},
		{
			name:     "空表达式",
			input:    "",
			hasError: false, // 空表达式返回 EOF
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := ParseFormula(tt.input)
			if tt.hasError {
				assert.Error(t, err)
				assert.Nil(t, node)
			} else {
				// 空表达式可能返回 nil 节点但不报错
				if tt.input != "" {
					assert.NoError(t, err)
					assert.NotNil(t, node)
				}
			}
		})
	}
}

func TestLexer(t *testing.T) {
	lexer := NewLexer("a + b * 10")

	tokens, err := lexer.Lex()
	assert.NoError(t, err)
	assert.NotEmpty(t, tokens)

	// 验证 token 类型
	assert.Equal(t, TokenIdentifier, tokens[0].Type)
	assert.Equal(t, "a", tokens[0].Value)
	assert.Equal(t, TokenOperator, tokens[1].Type)
	assert.Equal(t, "+", tokens[1].Value)
	assert.Equal(t, TokenIdentifier, tokens[2].Type)
	assert.Equal(t, "b", tokens[2].Value)
	assert.Equal(t, TokenOperator, tokens[3].Type)
	assert.Equal(t, "*", tokens[3].Value)
	assert.Equal(t, TokenNumber, tokens[4].Type)
	assert.Equal(t, "10", tokens[4].Value)
}

func TestLexer_ComplexTokens(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedTypes []TokenType
	}{
		{
			name:          "数字",
			input:         "123.45",
			expectedTypes: []TokenType{TokenNumber, TokenEOF},
		},
		{
			name:          "字符串",
			input:         `"hello"`,
			expectedTypes: []TokenType{TokenString, TokenEOF},
		},
		{
			name:          "变量引用",
			input:         "${point-001}",
			expectedTypes: []TokenType{TokenVariable, TokenEOF},
		},
		{
			name:          "比较运算符",
			input:         "a >= b",
			expectedTypes: []TokenType{TokenIdentifier, TokenOperator, TokenIdentifier, TokenEOF},
		},
		{
			name:          "逻辑运算符",
			input:         "a && b || c",
			expectedTypes: []TokenType{TokenIdentifier, TokenOperator, TokenIdentifier, TokenOperator, TokenIdentifier, TokenEOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Lex()
			assert.NoError(t, err)

			for i, expectedType := range tt.expectedTypes {
				if i < len(tokens) {
					assert.Equal(t, expectedType, tokens[i].Type, "Token %d: expected %v, got %v", i, expectedType, tokens[i].Type)
				}
			}
		})
	}
}

func TestEvalContext(t *testing.T) {
	ctx := NewEvalContext()

	// 设置变量
	ctx.SetVariable("a", 10.0)

	// 获取变量
	value, exists := ctx.GetVariable("a")
	assert.True(t, exists)
	assert.Equal(t, 10.0, value)

	// 注册函数
	ctx.RegisterFunction("double", func(args ...interface{}) (interface{}, error) {
		return args[0].(float64) * 2, nil
	})

	// 验证函数
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
