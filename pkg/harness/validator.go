package harness

import (
	"context"
	"errors"
)

// Validator 输入验证器接口
type Validator interface {
	Validate(ctx context.Context, input interface{}) error
	ValidateAsync(ctx context.Context, input interface{}) (<-chan ValidationResult, error)
}

// ValidationResult 验证结果
type ValidationResult struct {
	Valid    bool
	Errors   []error
	Warnings []string
}

// DefaultValidator 默认验证器实现
type DefaultValidator struct{}

// NewDefaultValidator 创建默认验证器
func NewDefaultValidator() *DefaultValidator {
	return &DefaultValidator{}
}

// Validate 执行同步验证
func (v *DefaultValidator) Validate(ctx context.Context, input interface{}) error {
	if input == nil {
		return errors.New("input cannot be nil")
	}
	return nil
}

// ValidateAsync 执行异步验证
func (v *DefaultValidator) ValidateAsync(ctx context.Context, input interface{}) (<-chan ValidationResult, error) {
	resultChan := make(chan ValidationResult, 1)
	go func() {
		defer close(resultChan)
		err := v.Validate(ctx, input)
		result := ValidationResult{
			Valid:    err == nil,
			Errors:   nil,
			Warnings: []string{},
		}
		if err != nil {
			result.Errors = []error{err}
		}
		resultChan <- result
	}()
	return resultChan, nil
}
