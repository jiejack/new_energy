package harness

import (
	"context"
	"errors"
)

// Harness 主入口，整合所有组件
type Harness struct {
	validator  Validator
	verifier   Verifier
	constraint Constraint
	monitor    Monitor
	snapshot   *SnapshotManager
}

// NewHarness 创建新的 Harness 实例
func NewHarness() *Harness {
	return &Harness{
		validator:  NewDefaultValidator(),
		verifier:   NewDefaultVerifier(),
		constraint: NewDefaultConstraint(),
		monitor:    NewDefaultMonitor(),
		snapshot:   NewSnapshotManager(),
	}
}

// NewHarnessWithComponents 使用自定义组件创建 Harness 实例
func NewHarnessWithComponents(
	validator Validator,
	verifier Verifier,
	constraint Constraint,
	monitor Monitor,
) *Harness {
	return &Harness{
		validator:  validator,
		verifier:   verifier,
		constraint: constraint,
		monitor:    monitor,
		snapshot:   NewSnapshotManager(),
	}
}

// Validate 执行验证（包括输入验证和约束检查）
func (h *Harness) Validate(ctx context.Context, input interface{}) error {
	// 1. 执行输入验证
	if err := h.validator.Validate(ctx, input); err != nil {
		return err
	}

	// 2. 执行约束检查
	if ok, err := h.constraint.Check(ctx, input); !ok || err != nil {
		if err != nil {
			return err
		}
		return errors.New("constraint check failed")
	}

	return nil
}

// ValidateAsync 执行异步验证
func (h *Harness) ValidateAsync(ctx context.Context, input interface{}) (<-chan ValidationResult, error) {
	return h.validator.ValidateAsync(ctx, input)
}

// Verify 执行输出验证
func (h *Harness) Verify(ctx context.Context, expected, actual interface{}) (bool, error) {
	return h.verifier.Verify(ctx, expected, actual)
}

// Snapshot 创建快照
func (h *Harness) Snapshot(ctx context.Context, target interface{}) ([]byte, error) {
	return h.verifier.Snapshot(ctx, target)
}

// SaveSnapshot 保存快照到管理器
func (h *Harness) SaveSnapshot(id string, data []byte) error {
	return h.snapshot.Save(id, data)
}

// LoadSnapshot 从管理器加载快照
func (h *Harness) LoadSnapshot(id string) (*Snapshot, error) {
	return h.snapshot.Load(id)
}

// CompareSnapshot 比较快照
func (h *Harness) CompareSnapshot(id string, data []byte) (bool, error) {
	return h.snapshot.Compare(id, data)
}

// ApplyConstraint 应用约束条件
func (h *Harness) ApplyConstraint(ctx context.Context, target interface{}) error {
	return h.constraint.Apply(ctx, target)
}

// CheckConstraint 检查约束条件
func (h *Harness) CheckConstraint(ctx context.Context, target interface{}) (bool, error) {
	return h.constraint.Check(ctx, target)
}

// RecordMetric 记录指标
func (h *Harness) RecordMetric(ctx context.Context, metric string, value float64) error {
	return h.monitor.Record(ctx, metric, value)
}

// GetMetrics 获取指标
func (h *Harness) GetMetrics(ctx context.Context, pattern string) ([]Metric, error) {
	return h.monitor.GetMetrics(ctx, pattern)
}

// Execute 执行完整的测试流程
func (h *Harness) Execute(ctx context.Context, input interface{}, expectedOutput interface{}, actualOutput interface{}) error {
	// 1. 验证输入
	if err := h.Validate(ctx, input); err != nil {
		return err
	}

	// 2. 验证输出
	if ok, err := h.Verify(ctx, expectedOutput, actualOutput); !ok || err != nil {
		if err != nil {
			return err
		}
		return errors.New("output verification failed")
	}

	return nil
}

// ExecuteWithSnapshot 执行完整的测试流程并创建快照
func (h *Harness) ExecuteWithSnapshot(ctx context.Context, snapshotID string, input interface{}, expectedOutput interface{}, actualOutput interface{}) error {
	// 1. 执行验证
	if err := h.Execute(ctx, input, expectedOutput, actualOutput); err != nil {
		return err
	}

	// 2. 创建快照
	snapshot, err := h.Snapshot(ctx, actualOutput)
	if err != nil {
		return err
	}

	// 3. 保存快照
	if err := h.SaveSnapshot(snapshotID, snapshot); err != nil {
		return err
	}

	return nil
}

// GetValidator 获取验证器
func (h *Harness) GetValidator() Validator {
	return h.validator
}

// GetVerifier 获取验证器
func (h *Harness) GetVerifier() Verifier {
	return h.verifier
}

// GetConstraint 获取约束
func (h *Harness) GetConstraint() Constraint {
	return h.constraint
}

// GetMonitor 获取监控器
func (h *Harness) GetMonitor() Monitor {
	return h.monitor
}

// GetSnapshotManager 获取快照管理器
func (h *Harness) GetSnapshotManager() *SnapshotManager {
	return h.snapshot
}
