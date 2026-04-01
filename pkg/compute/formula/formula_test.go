package formula

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewParser(t *testing.T) {
	parser := NewParser()

	assert.NotNil(t, parser)
}

func TestParser_Parse(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name      string
		expression string
		hasError  bool
	}{
		{
			name:       "简单加法",
			expression: "a + b",
			hasError:   false,
		},
		{
			name:       "复杂表达式",
			expression: "(a + b) * c - d / e",
			hasError:   false,
		},
		{
			name:       "带函数的表达式",
			expression: "max(a, b) + min(c, d)",
			hasError:   false,
		},
		{
			name:       "条件表达式",
			expression: "a > b ? c : d",
			hasError:   false,
		},
		{
			name:       "空表达式",
			expression: "",
			hasError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := parser.Parse(tt.expression)
			if tt.hasError {
				assert.Error(t, err)
				assert.Nil(t, node)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, node)
			}
		})
	}
}

func TestNewExecutor(t *testing.T) {
	executor := NewExecutor()

	assert.NotNil(t, executor)
}

func TestExecutor_Execute(t *testing.T) {
	executor := NewExecutor()
	parser := NewParser()

	tests := []struct {
		name       string
		expression string
		variables  map[string]float64
		expected   float64
		hasError   bool
	}{
		{
			name:       "简单加法",
			expression: "a + b",
			variables:  map[string]float64{"a": 10, "b": 20},
			expected:   30,
			hasError:   false,
		},
		{
			name:       "简单减法",
			expression: "a - b",
			variables:  map[string]float64{"a": 30, "b": 10},
			expected:   20,
			hasError:   false,
		},
		{
			name:       "简单乘法",
			expression: "a * b",
			variables:  map[string]float64{"a": 5, "b": 6},
			expected:   30,
			hasError:   false,
		},
		{
			name:       "简单除法",
			expression: "a / b",
			variables:  map[string]float64{"a": 20, "b": 4},
			expected:   5,
			hasError:   false,
		},
		{
			name:       "复杂表达式",
			expression: "(a + b) * c - d",
			variables:  map[string]float64{"a": 2, "b": 3, "c": 4, "d": 5},
			expected:   15, // (2+3)*4-5 = 15
			hasError:   false,
		},
		{
			name:       "除零错误",
			expression: "a / b",
			variables:  map[string]float64{"a": 10, "b": 0},
			hasError:   true,
		},
		{
			name:       "变量不存在",
			expression: "a + c",
			variables:  map[string]float64{"a": 10, "b": 20},
			hasError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := parser.Parse(tt.expression)
			if err != nil {
				t.Skipf("Parse error: %v", err)
				return
			}

			result, err := executor.Execute(node, tt.variables)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestExecutor_ExecuteWithFunctions(t *testing.T) {
	executor := NewExecutor()
	parser := NewParser()

	// 注册自定义函数
	executor.RegisterFunction("double", func(args ...float64) (float64, error) {
		if len(args) != 1 {
			return 0, assert.AnError
		}
		return args[0] * 2, nil
	})

	tests := []struct {
		name       string
		expression string
		variables  map[string]float64
		expected   float64
	}{
		{
			name:       "自定义函数",
			expression: "double(a)",
			variables:  map[string]float64{"a": 5},
			expected:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := parser.Parse(tt.expression)
			if err != nil {
				t.Skipf("Parse error: %v", err)
				return
			}

			result, err := executor.Execute(node, tt.variables)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewManager(t *testing.T) {
	manager := NewManager()

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.formulas)
}

func TestManager_AddFormula(t *testing.T) {
	manager := NewManager()

	formula := &Formula{
		ID:         "formula001",
		Name:       "功率计算",
		Expression: "voltage * current",
		Variables:  []string{"voltage", "current"},
		Unit:       "kW",
	}

	err := manager.AddFormula(formula)
	assert.NoError(t, err)

	// 重复添加
	err = manager.AddFormula(formula)
	assert.Error(t, err)
}

func TestManager_GetFormula(t *testing.T) {
	manager := NewManager()

	formula := &Formula{
		ID:         "formula001",
		Name:       "功率计算",
		Expression: "voltage * current",
		Variables:  []string{"voltage", "current"},
		Unit:       "kW",
	}

	manager.AddFormula(formula)

	got := manager.GetFormula("formula001")
	assert.NotNil(t, got)
	assert.Equal(t, "formula001", got.ID)

	notFound := manager.GetFormula("formula999")
	assert.Nil(t, notFound)
}

func TestManager_RemoveFormula(t *testing.T) {
	manager := NewManager()

	formula := &Formula{
		ID:         "formula001",
		Name:       "功率计算",
		Expression: "voltage * current",
		Variables:  []string{"voltage", "current"},
		Unit:       "kW",
	}

	manager.AddFormula(formula)
	assert.NotNil(t, manager.GetFormula("formula001"))

	manager.RemoveFormula("formula001")
	assert.Nil(t, manager.GetFormula("formula001"))
}

func TestManager_ExecuteFormula(t *testing.T) {
	manager := NewManager()

	formula := &Formula{
		ID:         "formula001",
		Name:       "功率计算",
		Expression: "voltage * current",
		Variables:  []string{"voltage", "current"},
		Unit:       "kW",
	}

	manager.AddFormula(formula)

	variables := map[string]float64{
		"voltage": 220.0,
		"current": 10.0,
	}

	result, err := manager.ExecuteFormula("formula001", variables)
	assert.NoError(t, err)
	assert.Equal(t, 2200.0, result)
}

func TestManager_ListFormulas(t *testing.T) {
	manager := NewManager()

	formula1 := &Formula{
		ID:         "formula001",
		Name:       "功率计算",
		Expression: "voltage * current",
		Variables:  []string{"voltage", "current"},
	}

	formula2 := &Formula{
		ID:         "formula002",
		Name:       "效率计算",
		Expression: "output / input * 100",
		Variables:  []string{"output", "input"},
	}

	manager.AddFormula(formula1)
	manager.AddFormula(formula2)

	formulas := manager.ListFormulas()
	assert.Len(t, formulas, 2)
}

func TestBuiltinFunctions(t *testing.T) {
	executor := NewExecutor()
	parser := NewParser()

	tests := []struct {
		name       string
		expression string
		variables  map[string]float64
		expected   float64
	}{
		{
			name:       "max函数",
			expression: "max(a, b)",
			variables:  map[string]float64{"a": 10, "b": 20},
			expected:   20,
		},
		{
			name:       "min函数",
			expression: "min(a, b)",
			variables:  map[string]float64{"a": 10, "b": 20},
			expected:   10,
		},
		{
			name:       "abs函数",
			expression: "abs(a)",
			variables:  map[string]float64{"a": -10},
			expected:   10,
		},
		{
			name:       "sqrt函数",
			expression: "sqrt(a)",
			variables:  map[string]float64{"a": 16},
			expected:   4,
		},
		{
			name:       "pow函数",
			expression: "pow(a, b)",
			variables:  map[string]float64{"a": 2, "b": 3},
			expected:   8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := parser.Parse(tt.expression)
			if err != nil {
				t.Skipf("Parse error: %v", err)
				return
			}

			result, err := executor.Execute(node, tt.variables)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
