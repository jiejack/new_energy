package harness

import (
	"context"
	"errors"
	"testing"
)

// mockConstraint 是一个用于测试的 Constraint 实现
type mockConstraint struct {
	checkResult bool
	checkError  error
	applyError  error
}

func (m *mockConstraint) Check(ctx context.Context, target interface{}) (bool, error) {
	if m.checkError != nil {
		return false, m.checkError
	}
	return m.checkResult, nil
}

func (m *mockConstraint) Apply(ctx context.Context, target interface{}) error {
	return m.applyError
}

// TestConstraint_Check 测试约束检查功能
func TestConstraint_Check(t *testing.T) {
	tests := []struct {
		name       string
		constraint Constraint
		target     interface{}
		wantResult bool
		wantErr    bool
		errMessage string
	}{
		{
			name:       "检查通过",
			constraint: &mockConstraint{checkResult: true},
			target:     struct{ Value int }{Value: 100},
			wantResult: true,
			wantErr:    false,
		},
		{
			name:       "检查不通过",
			constraint: &mockConstraint{checkResult: false},
			target:     struct{ Value int }{Value: 50},
			wantResult: false,
			wantErr:    false,
		},
		{
			name:       "检查返回错误",
			constraint: &mockConstraint{checkError: errors.New("check failed")},
			target:     struct{ Value int }{Value: 100},
			wantResult: false,
			wantErr:    true,
			errMessage: "check failed",
		},
		{
			name:       "nil目标对象",
			constraint: &mockConstraint{checkResult: true},
			target:     nil,
			wantResult: true,
			wantErr:    false,
		},
		{
			name:       "复杂对象检查",
			constraint: &mockConstraint{checkResult: true},
			target:     map[string]interface{}{"key": "value", "number": 42},
			wantResult: true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := tt.constraint.Check(ctx, tt.target)

			if (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != tt.wantResult {
				t.Errorf("Check() result = %v, want %v", result, tt.wantResult)
			}

			if tt.wantErr && err.Error() != tt.errMessage {
				t.Errorf("Check() error message = %v, want %v", err.Error(), tt.errMessage)
			}
		})
	}
}

// TestConstraint_Apply 测试约束应用功能
func TestConstraint_Apply(t *testing.T) {
	tests := []struct {
		name       string
		constraint Constraint
		target     interface{}
		wantErr    bool
		errMessage string
	}{
		{
			name:       "应用成功",
			constraint: &mockConstraint{applyError: nil},
			target:     struct{ Value int }{Value: 100},
			wantErr:    false,
		},
		{
			name:       "应用失败",
			constraint: &mockConstraint{applyError: errors.New("apply failed")},
			target:     struct{ Value int }{Value: 50},
			wantErr:    true,
			errMessage: "apply failed",
		},
		{
			name:       "nil目标对象应用",
			constraint: &mockConstraint{applyError: nil},
			target:     nil,
			wantErr:    false,
		},
		{
			name:       "复杂对象应用",
			constraint: &mockConstraint{applyError: nil},
			target:     map[string]interface{}{"key": "value", "number": 42},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := tt.constraint.Apply(ctx, tt.target)

			if (err != nil) != tt.wantErr {
				t.Errorf("Apply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err.Error() != tt.errMessage {
				t.Errorf("Apply() error message = %v, want %v", err.Error(), tt.errMessage)
			}
		})
	}
}

// TestConstraint_Context_Cancellation 测试上下文取消
func TestConstraint_Context_Cancellation(t *testing.T) {
	constraint := &mockConstraint{checkResult: true, applyError: nil}
	ctx, cancel := context.WithCancel(context.Background())

	// 立即取消上下文
	cancel()

	// 即使上下文已取消，Check 和 Apply 也应该能正常执行
	// 因为 mock 实现不检查上下文状态
	result, err := constraint.Check(ctx, struct{ Value int }{Value: 100})
	if err != nil {
		t.Errorf("Check() returned error: %v", err)
		return
	}
	if !result {
		t.Error("Check() result should be true")
	}

	err = constraint.Apply(ctx, struct{ Value int }{Value: 100})
	if err != nil {
		t.Errorf("Apply() returned error: %v", err)
	}
}

// TestConstraint_Interface_Compliance 测试接口合规性
func TestConstraint_Interface_Compliance(t *testing.T) {
	// 确保 mockConstraint 实现了 Constraint 接口
	var _ Constraint = (*mockConstraint)(nil)

	// 测试接口方法签名
	constraint := &mockConstraint{checkResult: true}

	ctx := context.Background()
	result, err := constraint.Check(ctx, nil)
	if err != nil {
		t.Errorf("Check() returned unexpected error: %v", err)
	}
	if !result {
		t.Error("Check() result should be true")
	}

	err = constraint.Apply(ctx, nil)
	if err != nil {
		t.Errorf("Apply() returned unexpected error: %v", err)
	}
}

// TestConstraint_MultipleChecks 测试多次检查
func TestConstraint_MultipleChecks(t *testing.T) {
	constraint := &mockConstraint{checkResult: true}
	ctx := context.Background()

	targets := []interface{}{
		struct{ Value int }{Value: 100},
		struct{ Value int }{Value: 200},
		struct{ Value int }{Value: 300},
	}

	for i, target := range targets {
		result, err := constraint.Check(ctx, target)
		if err != nil {
			t.Errorf("Check() for target %d returned error: %v", i, err)
			continue
		}
		if !result {
			t.Errorf("Check() for target %d should return true", i)
		}
	}
}

// TestConstraint_CheckAndApply 测试检查和应用组合
func TestConstraint_CheckAndApply(t *testing.T) {
	tests := []struct {
		name        string
		constraint  Constraint
		target      interface{}
		wantCheck   bool
		wantApplyOk bool
	}{
		{
			name:        "检查通过并应用成功",
			constraint:  &mockConstraint{checkResult: true, applyError: nil},
			target:      struct{ Value int }{Value: 100},
			wantCheck:   true,
			wantApplyOk: true,
		},
		{
			name:        "检查不通过但应用成功",
			constraint:  &mockConstraint{checkResult: false, applyError: nil},
			target:      struct{ Value int }{Value: 50},
			wantCheck:   false,
			wantApplyOk: true,
		},
		{
			name:        "检查通过但应用失败",
			constraint:  &mockConstraint{checkResult: true, applyError: errors.New("apply error")},
			target:      struct{ Value int }{Value: 100},
			wantCheck:   true,
			wantApplyOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// 先检查
			result, err := tt.constraint.Check(ctx, tt.target)
			if err != nil {
				t.Errorf("Check() returned error: %v", err)
				return
			}
			if result != tt.wantCheck {
				t.Errorf("Check() result = %v, want %v", result, tt.wantCheck)
			}

			// 再应用
			err = tt.constraint.Apply(ctx, tt.target)
			if (err == nil) != tt.wantApplyOk {
				t.Errorf("Apply() success = %v, want %v", err == nil, tt.wantApplyOk)
			}
		})
	}
}
