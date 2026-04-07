package harness

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockValidator 是一个用于测试的 Validator 实现
type mockValidator struct {
	shouldFail bool
	errors     []error
	warnings   []string
}

func (m *mockValidator) Validate(ctx context.Context, input interface{}) error {
	if m.shouldFail {
		return errors.New("validation failed")
	}
	return nil
}

func (m *mockValidator) ValidateAsync(ctx context.Context, input interface{}) (<-chan ValidationResult, error) {
	resultChan := make(chan ValidationResult, 1)
	go func() {
		defer close(resultChan)
		time.Sleep(10 * time.Millisecond) // 模拟异步处理
		resultChan <- ValidationResult{
			Valid:    !m.shouldFail,
			Errors:   m.errors,
			Warnings: m.warnings,
		}
	}()
	return resultChan, nil
}

func TestValidator_Validate(t *testing.T) {
	tests := []struct {
		name       string
		validator  Validator
		input      interface{}
		wantErr    bool
		errMessage string
	}{
		{
			name:      "成功验证",
			validator: &mockValidator{shouldFail: false},
			input:     struct{ Name string }{Name: "test"},
			wantErr:   false,
		},
		{
			name:       "验证失败",
			validator:  &mockValidator{shouldFail: true},
			input:      struct{ Name string }{Name: ""},
			wantErr:    true,
			errMessage: "validation failed",
		},
		{
			name:      "nil输入",
			validator: &mockValidator{shouldFail: false},
			input:     nil,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := tt.validator.Validate(ctx, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err.Error() != tt.errMessage {
				t.Errorf("Validate() error message = %v, want %v", err.Error(), tt.errMessage)
			}
		})
	}
}

func TestValidator_ValidateAsync(t *testing.T) {
	tests := []struct {
		name      string
		validator Validator
		input     interface{}
		wantValid bool
		errors    []error
		warnings  []string
	}{
		{
			name:      "异步验证成功",
			validator: &mockValidator{shouldFail: false, warnings: []string{"这是一个警告"}},
			input:     struct{ Name string }{Name: "test"},
			wantValid: true,
			warnings:  []string{"这是一个警告"},
		},
		{
			name:      "异步验证失败",
			validator: &mockValidator{shouldFail: true, errors: []error{errors.New("错误1"), errors.New("错误2")}},
			input:     struct{ Name string }{Name: ""},
			wantValid: false,
			errors:    []error{errors.New("错误1"), errors.New("错误2")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			resultChan, err := tt.validator.ValidateAsync(ctx, tt.input)
			if err != nil {
				t.Errorf("ValidateAsync() returned error: %v", err)
				return
			}

			select {
			case result := <-resultChan:
				if result.Valid != tt.wantValid {
					t.Errorf("ValidateAsync() result.Valid = %v, want %v", result.Valid, tt.wantValid)
				}

				if len(result.Errors) != len(tt.errors) {
					t.Errorf("ValidateAsync() result.Errors length = %v, want %v", len(result.Errors), len(tt.errors))
				}

				if len(result.Warnings) != len(tt.warnings) {
					t.Errorf("ValidateAsync() result.Warnings length = %v, want %v", len(result.Warnings), len(tt.warnings))
				}

				for i, w := range result.Warnings {
					if i < len(tt.warnings) && w != tt.warnings[i] {
						t.Errorf("ValidateAsync() result.Warnings[%d] = %v, want %v", i, w, tt.warnings[i])
					}
				}

			case <-ctx.Done():
				t.Error("ValidateAsync() timeout")
			}
		})
	}
}

func TestValidationResult_Structure(t *testing.T) {
	// 测试 ValidationResult 结构体的字段
	result := ValidationResult{
		Valid:    true,
		Errors:   []error{},
		Warnings: []string{"警告1", "警告2"},
	}

	if result.Valid != true {
		t.Errorf("ValidationResult.Valid = %v, want true", result.Valid)
	}

	if len(result.Errors) != 0 {
		t.Errorf("ValidationResult.Errors length = %v, want 0", len(result.Errors))
	}

	if len(result.Warnings) != 2 {
		t.Errorf("ValidationResult.Warnings length = %v, want 2", len(result.Warnings))
	}

	if result.Warnings[0] != "警告1" || result.Warnings[1] != "警告2" {
		t.Errorf("ValidationResult.Warnings content incorrect")
	}
}

func TestValidator_Context_Cancellation(t *testing.T) {
	validator := &mockValidator{shouldFail: false}
	ctx, cancel := context.WithCancel(context.Background())

	// 立即取消上下文
	cancel()

	resultChan, err := validator.ValidateAsync(ctx, struct{ Name string }{Name: "test"})
	if err != nil {
		t.Errorf("ValidateAsync() returned error: %v", err)
		return
	}

	// 即使上下文已取消，也应该能收到结果（因为结果已经在channel中）
	select {
	case result := <-resultChan:
		// 成功接收到结果
		if !result.Valid {
			t.Error("Expected valid result")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive result")
	}
}
