package rule

import (
	"context"
	"testing"
	"time"
)

// TestDSL 测试DSL定义
func TestDSL(t *testing.T) {
	rule := NewRuleDSL("test-rule-001", "测试规则", "1.0")
	rule.Description = "这是一个测试规则"
	rule.SetComparisonCondition("point-001", OpGT, ThresholdTypeAbsolute, 100)

	if err := rule.Validate(); err != nil {
		t.Errorf("Rule validation failed: %v", err)
	}

	if rule.ID != "test-rule-001" {
		t.Errorf("Expected ID 'test-rule-001', got '%s'", rule.ID)
	}

	if rule.Condition.Comparison == nil {
		t.Error("Expected comparison condition to be set")
	}

	if rule.Condition.Comparison.Operator != OpGT {
		t.Errorf("Expected operator '>', got '%s'", rule.Condition.Comparison.Operator)
	}
}

// TestParser 测试解析器
func TestParser(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "简单比较条件",
			input:   "point-001 > 100",
			wantErr: false,
		},
		{
			name:    "窗口函数",
			input:   "avg(point-001, 5m) > 50",
			wantErr: false,
		},
		{
			name:    "百分比阈值",
			input:   "point-001 > percentage(80)",
			wantErr: false,
		},
		{
			name:    "逻辑AND",
			input:   "AND(point-001 > 100, point-002 < 50)",
			wantErr: false,
		},
		{
			name:    "逻辑OR",
			input:   "OR(point-001 > 100, point-002 < 50)",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && condition == nil {
				t.Error("Expected non-nil condition")
			}
		})
	}
}

// TestEngine 测试执行引擎
func TestEngine(t *testing.T) {
	// 创建模拟数据提供者
	provider := &mockDataProvider{
		values: map[string]float64{
			"point-001": 120.0,
			"point-002": 40.0,
		},
	}

	engine := NewEngine(provider)

	// 创建测试规则
	rule := NewRuleDSL("test-rule-001", "测试规则", "1.0")
	rule.SetComparisonCondition("point-001", OpGT, ThresholdTypeAbsolute, 100)

	if err := engine.AddRule(rule); err != nil {
		t.Fatalf("Failed to add rule: %v", err)
	}

	// 评估规则
	result, err := engine.Evaluate(context.Background(), "test-rule-001")
	if err != nil {
		t.Fatalf("Failed to evaluate rule: %v", err)
	}

	if !result.Triggered {
		t.Error("Expected rule to be triggered")
	}

	if result.Value != 120.0 {
		t.Errorf("Expected value 120.0, got %f", result.Value)
	}
}

// TestVersionManager 测试版本管理
func TestVersionManager(t *testing.T) {
	vm := NewVersionManager()

	rule := NewRuleDSL("test-rule-001", "测试规则", "1.0")
	rule.SetComparisonCondition("point-001", OpGT, ThresholdTypeAbsolute, 100)

	// 创建版本
	version, err := vm.CreateVersion(rule, "初始版本", "创建规则", "admin")
	if err != nil {
		t.Fatalf("Failed to create version: %v", err)
	}

	if version.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", version.Version)
	}

	// 获取活跃版本
	activeVersion, err := vm.GetActiveVersion("test-rule-001")
	if err != nil {
		t.Fatalf("Failed to get active version: %v", err)
	}

	if activeVersion.Version != "1.0" {
		t.Errorf("Expected active version '1.0', got '%s'", activeVersion.Version)
	}

	// 获取历史版本
	history, err := vm.GetHistory("test-rule-001")
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if len(history) != 1 {
		t.Errorf("Expected 1 version in history, got %d", len(history))
	}
}

// TestRuleManager 测试规则管理器
func TestRuleManager(t *testing.T) {
	// 创建模拟数据提供者
	provider := &mockDataProvider{
		values: map[string]float64{
			"point-001": 120.0,
		},
	}

	engine := NewEngine(provider)
	manager := NewRuleManager(engine, nil)

	rule := NewRuleDSL("test-rule-001", "测试规则", "1.0")
	rule.SetComparisonCondition("point-001", OpGT, ThresholdTypeAbsolute, 100)
	rule.CreatedBy = "admin"

	// 创建规则
	if err := manager.CreateRule(context.Background(), rule); err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	// 获取规则
	retrieved, err := manager.GetRule(context.Background(), "test-rule-001")
	if err != nil {
		t.Fatalf("Failed to get rule: %v", err)
	}

	if retrieved.ID != "test-rule-001" {
		t.Errorf("Expected ID 'test-rule-001', got '%s'", retrieved.ID)
	}

	// 列出规则
	rules, err := manager.ListRules(context.Background())
	if err != nil {
		t.Fatalf("Failed to list rules: %v", err)
	}

	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}

	// 删除规则
	if err := manager.DeleteRule(context.Background(), "test-rule-001"); err != nil {
		t.Fatalf("Failed to delete rule: %v", err)
	}

	// 验证删除
	_, err = manager.GetRule(context.Background(), "test-rule-001")
	if err == nil {
		t.Error("Expected error when getting deleted rule")
	}
}

// mockDataProvider 模拟数据提供者
type mockDataProvider struct {
	values map[string]float64
}

func (m *mockDataProvider) GetCurrentValue(ctx context.Context, pointID string) (*PointData, error) {
	if value, exists := m.values[pointID]; exists {
		return &PointData{
			PointID:   pointID,
			Value:     value,
			Timestamp: time.Now(),
		}, nil
	}
	return nil, nil
}

func (m *mockDataProvider) GetTimeSeries(ctx context.Context, pointID string, start, end time.Time) (*TimeSeriesData, error) {
	return &TimeSeriesData{
		PointID: pointID,
		Values:  []PointData{},
	}, nil
}

func (m *mockDataProvider) GetWindowData(ctx context.Context, pointID string, window time.Duration) (*TimeSeriesData, error) {
	if value, exists := m.values[pointID]; exists {
		return &TimeSeriesData{
			PointID: pointID,
			Values: []PointData{
				{
					PointID:   pointID,
					Value:     value,
					Timestamp: time.Now(),
				},
			},
		}, nil
	}
	return &TimeSeriesData{
		PointID: pointID,
		Values:  []PointData{},
	}, nil
}
