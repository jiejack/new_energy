package harness

import "context"

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
