package harness

import (
	"context"
	"encoding/json"
	"reflect"
)

// Verifier 输出验证器接口
type Verifier interface {
	// Verify 验证实际输出是否符合预期
	Verify(ctx context.Context, expected, actual interface{}) (bool, error)

	// Snapshot 创建目标对象的快照
	Snapshot(ctx context.Context, target interface{}) ([]byte, error)
}

// DefaultVerifier 默认验证器实现
type DefaultVerifier struct{}

// NewDefaultVerifier 创建默认验证器
func NewDefaultVerifier() *DefaultVerifier {
	return &DefaultVerifier{}
}

// Verify 验证实际输出是否符合预期（使用深度比较）
func (v *DefaultVerifier) Verify(ctx context.Context, expected, actual interface{}) (bool, error) {
	return reflect.DeepEqual(expected, actual), nil
}

// Snapshot 创建目标对象的快照（序列化为 JSON）
func (v *DefaultVerifier) Snapshot(ctx context.Context, target interface{}) ([]byte, error) {
	return json.Marshal(target)
}
