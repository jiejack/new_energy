package harness

import "context"

// Verifier 输出验证器接口
type Verifier interface {
	// Verify 验证实际输出是否符合预期
	Verify(ctx context.Context, expected, actual interface{}) (bool, error)

	// Snapshot 创建目标对象的快照
	Snapshot(ctx context.Context, target interface{}) ([]byte, error)
}
