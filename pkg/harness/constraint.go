package harness

import (
	"context"
)

// Constraint 约束条件接口
type Constraint interface {
	// Check 检查目标对象是否满足约束条件
	Check(ctx context.Context, target interface{}) (bool, error)

	// Apply 对目标对象应用约束条件
	Apply(ctx context.Context, target interface{}) error
}

// DefaultConstraint 默认约束实现
type DefaultConstraint struct{}

// NewDefaultConstraint 创建默认约束
func NewDefaultConstraint() *DefaultConstraint {
	return &DefaultConstraint{}
}

// Check 检查目标对象是否满足约束条件（默认总是返回 true）
func (c *DefaultConstraint) Check(ctx context.Context, target interface{}) (bool, error) {
	return true, nil
}

// Apply 对目标对象应用约束条件（默认不做任何操作）
func (c *DefaultConstraint) Apply(ctx context.Context, target interface{}) error {
	return nil
}
