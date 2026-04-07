package harness

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestNewHarness 测试创建 Harness 实例
func TestNewHarness(t *testing.T) {
	h := NewHarness()
	if h == nil {
		t.Fatal("NewHarness() returned nil")
	}

	if h.validator == nil {
		t.Error("validator is nil")
	}

	if h.verifier == nil {
		t.Error("verifier is nil")
	}

	if h.constraint == nil {
		t.Error("constraint is nil")
	}

	if h.monitor == nil {
		t.Error("monitor is nil")
	}

	if h.snapshot == nil {
		t.Error("snapshot is nil")
	}
}

// TestNewHarnessWithComponents 测试使用自定义组件创建 Harness
func TestNewHarnessWithComponents(t *testing.T) {
	validator := NewDefaultValidator()
	verifier := NewDefaultVerifier()
	constraint := NewDefaultConstraint()
	monitor := NewDefaultMonitor()

	h := NewHarnessWithComponents(validator, verifier, constraint, monitor)
	if h == nil {
		t.Fatal("NewHarnessWithComponents() returned nil")
	}

	if h.validator != validator {
		t.Error("validator not set correctly")
	}

	if h.verifier != verifier {
		t.Error("verifier not set correctly")
	}

	if h.constraint != constraint {
		t.Error("constraint not set correctly")
	}

	if h.monitor != monitor {
		t.Error("monitor not set correctly")
	}
}

// TestHarness_Validate 测试验证功能
func TestHarness_Validate(t *testing.T) {
	h := NewHarness()
	ctx := context.Background()

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "有效输入",
			input:   struct{ Name string }{Name: "test"},
			wantErr: false,
		},
		{
			name:    "nil输入",
			input:   nil,
			wantErr: true,
		},
		{
			name:    "map输入",
			input:   map[string]int{"value": 100},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := h.Validate(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestHarness_ValidateAsync 测试异步验证功能
func TestHarness_ValidateAsync(t *testing.T) {
	h := NewHarness()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	resultChan, err := h.ValidateAsync(ctx, struct{ Name string }{Name: "test"})
	if err != nil {
		t.Fatalf("ValidateAsync() returned error: %v", err)
	}

	select {
	case result := <-resultChan:
		if !result.Valid {
			t.Error("Expected valid result")
		}
	case <-ctx.Done():
		t.Error("ValidateAsync() timeout")
	}
}

// TestHarness_Verify 测试输出验证功能
func TestHarness_Verify(t *testing.T) {
	h := NewHarness()
	ctx := context.Background()

	tests := []struct {
		name     string
		expected interface{}
		actual   interface{}
		want     bool
		wantErr  bool
	}{
		{
			name:     "相同的值",
			expected: 100,
			actual:   100,
			want:     true,
			wantErr:  false,
		},
		{
			name:     "不同的值",
			expected: 100,
			actual:   200,
			want:     false,
			wantErr:  false,
		},
		{
			name:     "相同的map",
			expected: map[string]int{"key": 1},
			actual:   map[string]int{"key": 1},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "nil值",
			expected: nil,
			actual:   nil,
			want:     true,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.Verify(ctx, tt.expected, tt.actual)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.want {
				t.Errorf("Verify() result = %v, want %v", result, tt.want)
			}
		})
	}
}

// TestHarness_Snapshot 测试快照功能
func TestHarness_Snapshot(t *testing.T) {
	h := NewHarness()
	ctx := context.Background()

	target := struct {
		Name  string
		Value int
	}{
		Name:  "test",
		Value: 100,
	}

	snapshot, err := h.Snapshot(ctx, target)
	if err != nil {
		t.Fatalf("Snapshot() returned error: %v", err)
	}

	if len(snapshot) == 0 {
		t.Error("Snapshot() returned empty data")
	}
}

// TestHarness_SnapshotManager 测试快照管理功能
func TestHarness_SnapshotManager(t *testing.T) {
	h := NewHarness()
	ctx := context.Background()

	// 创建快照
	target := map[string]int{"value": 100}
	snapshot, err := h.Snapshot(ctx, target)
	if err != nil {
		t.Fatalf("Snapshot() returned error: %v", err)
	}

	// 保存快照
	err = h.SaveSnapshot("test-snapshot", snapshot)
	if err != nil {
		t.Fatalf("SaveSnapshot() returned error: %v", err)
	}

	// 加载快照
	loaded, err := h.LoadSnapshot("test-snapshot")
	if err != nil {
		t.Fatalf("LoadSnapshot() returned error: %v", err)
	}

	if loaded.ID != "test-snapshot" {
		t.Errorf("LoadSnapshot() ID = %v, want test-snapshot", loaded.ID)
	}

	// 比较快照
	match, err := h.CompareSnapshot("test-snapshot", snapshot)
	if err != nil {
		t.Fatalf("CompareSnapshot() returned error: %v", err)
	}

	if !match {
		t.Error("CompareSnapshot() expected match")
	}
}

// TestHarness_Constraint 测试约束功能
func TestHarness_Constraint(t *testing.T) {
	h := NewHarness()
	ctx := context.Background()

	target := struct{ Value int }{Value: 100}

	// 检查约束
	ok, err := h.CheckConstraint(ctx, target)
	if err != nil {
		t.Fatalf("CheckConstraint() returned error: %v", err)
	}
	if !ok {
		t.Error("CheckConstraint() expected true")
	}

	// 应用约束
	err = h.ApplyConstraint(ctx, target)
	if err != nil {
		t.Fatalf("ApplyConstraint() returned error: %v", err)
	}
}

// TestHarness_Monitor 测试监控功能
func TestHarness_Monitor(t *testing.T) {
	h := NewHarness()
	ctx := context.Background()

	// 记录指标
	err := h.RecordMetric(ctx, "cpu_usage", 75.5)
	if err != nil {
		t.Fatalf("RecordMetric() returned error: %v", err)
	}

	err = h.RecordMetric(ctx, "memory_usage", 60.0)
	if err != nil {
		t.Fatalf("RecordMetric() returned error: %v", err)
	}

	// 获取指标
	metrics, err := h.GetMetrics(ctx, "")
	if err != nil {
		t.Fatalf("GetMetrics() returned error: %v", err)
	}

	if len(metrics) != 2 {
		t.Errorf("GetMetrics() returned %d metrics, want 2", len(metrics))
	}

	// 按模式过滤
	cpuMetrics, err := h.GetMetrics(ctx, "cpu")
	if err != nil {
		t.Fatalf("GetMetrics() with pattern returned error: %v", err)
	}

	if len(cpuMetrics) != 1 {
		t.Errorf("GetMetrics() with pattern returned %d metrics, want 1", len(cpuMetrics))
	}
}

// TestHarness_Execute 测试完整执行流程
func TestHarness_Execute(t *testing.T) {
	h := NewHarness()
	ctx := context.Background()

	tests := []struct {
		name           string
		input          interface{}
		expectedOutput interface{}
		actualOutput   interface{}
		wantErr        bool
	}{
		{
			name:           "成功执行",
			input:          struct{ Value int }{Value: 100},
			expectedOutput: 200,
			actualOutput:   200,
			wantErr:        false,
		},
		{
			name:           "输出不匹配",
			input:          struct{ Value int }{Value: 100},
			expectedOutput: 200,
			actualOutput:   300,
			wantErr:        true,
		},
		{
			name:           "nil输入",
			input:          nil,
			expectedOutput: 200,
			actualOutput:   200,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := h.Execute(ctx, tt.input, tt.expectedOutput, tt.actualOutput)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestHarness_ExecuteWithSnapshot 测试带快照的完整执行流程
func TestHarness_ExecuteWithSnapshot(t *testing.T) {
	h := NewHarness()
	ctx := context.Background()

	input := struct{ Value int }{Value: 100}
	expectedOutput := 200
	actualOutput := 200

	err := h.ExecuteWithSnapshot(ctx, "test-snapshot", input, expectedOutput, actualOutput)
	if err != nil {
		t.Fatalf("ExecuteWithSnapshot() returned error: %v", err)
	}

	// 验证快照已保存
	snapshot, err := h.LoadSnapshot("test-snapshot")
	if err != nil {
		t.Fatalf("LoadSnapshot() returned error: %v", err)
	}

	if snapshot == nil {
		t.Error("Snapshot not saved")
	}
}

// TestHarness_Getters 测试获取器方法
func TestHarness_Getters(t *testing.T) {
	h := NewHarness()

	if h.GetValidator() == nil {
		t.Error("GetValidator() returned nil")
	}

	if h.GetVerifier() == nil {
		t.Error("GetVerifier() returned nil")
	}

	if h.GetConstraint() == nil {
		t.Error("GetConstraint() returned nil")
	}

	if h.GetMonitor() == nil {
		t.Error("GetMonitor() returned nil")
	}

	if h.GetSnapshotManager() == nil {
		t.Error("GetSnapshotManager() returned nil")
	}
}

// TestHarness_Context_Cancellation 测试上下文取消
func TestHarness_Context_Cancellation(t *testing.T) {
	h := NewHarness()
	ctx, cancel := context.WithCancel(context.Background())

	// 立即取消上下文
	cancel()

	// 即使上下文已取消，也应该能正常执行（因为默认实现不检查上下文）
	err := h.Validate(ctx, struct{ Name string }{Name: "test"})
	if err != nil {
		t.Errorf("Validate() with cancelled context returned error: %v", err)
	}

	result, err := h.Verify(ctx, 100, 100)
	if err != nil {
		t.Errorf("Verify() with cancelled context returned error: %v", err)
	}
	if !result {
		t.Error("Verify() should return true")
	}
}

// TestHarness_Integration 测试集成场景
func TestHarness_Integration(t *testing.T) {
	h := NewHarness()
	ctx := context.Background()

	// 模拟一个完整的测试场景
	input := map[string]interface{}{
		"device_id": "device-001",
		"value":     100.5,
	}

	// 1. 验证输入
	if err := h.Validate(ctx, input); err != nil {
		t.Fatalf("Validate() failed: %v", err)
	}

	// 2. 记录指标
	if err := h.RecordMetric(ctx, "test.metric", 100.5); err != nil {
		t.Fatalf("RecordMetric() failed: %v", err)
	}

	// 3. 执行业务逻辑（模拟）
	expectedOutput := map[string]interface{}{
		"status":  "ok",
		"message": "success",
	}
	actualOutput := map[string]interface{}{
		"status":  "ok",
		"message": "success",
	}

	// 4. 验证输出
	if ok, err := h.Verify(ctx, expectedOutput, actualOutput); !ok || err != nil {
		t.Fatalf("Verify() failed: ok=%v, err=%v", ok, err)
	}

	// 5. 创建并保存快照
	snapshot, err := h.Snapshot(ctx, actualOutput)
	if err != nil {
		t.Fatalf("Snapshot() failed: %v", err)
	}

	if err := h.SaveSnapshot("integration-test", snapshot); err != nil {
		t.Fatalf("SaveSnapshot() failed: %v", err)
	}

	// 6. 验证快照
	match, err := h.CompareSnapshot("integration-test", snapshot)
	if err != nil {
		t.Fatalf("CompareSnapshot() failed: %v", err)
	}

	if !match {
		t.Error("Snapshot comparison failed")
	}

	// 7. 获取指标
	metrics, err := h.GetMetrics(ctx, "test")
	if err != nil {
		t.Fatalf("GetMetrics() failed: %v", err)
	}

	if len(metrics) == 0 {
		t.Error("No metrics found")
	}
}

// TestHarness_CustomComponents 测试自定义组件
func TestHarness_CustomComponents(t *testing.T) {
	// 创建自定义验证器
	customValidator := &mockValidator{shouldFail: false}

	// 创建自定义验证器
	customVerifier := &mockVerifier{verifyResult: true}

	// 创建自定义约束
	customConstraint := &mockConstraint{checkResult: true}

	// 创建自定义监控器
	customMonitor := newMockMonitor()

	// 使用自定义组件创建 Harness
	h := NewHarnessWithComponents(customValidator, customVerifier, customConstraint, customMonitor)

	ctx := context.Background()

	// 测试自定义组件是否正常工作
	err := h.Validate(ctx, struct{ Name string }{Name: "test"})
	if err != nil {
		t.Errorf("Validate() with custom validator failed: %v", err)
	}

	result, err := h.Verify(ctx, "expected", "actual")
	if err != nil {
		t.Errorf("Verify() with custom verifier failed: %v", err)
	}
	if !result {
		t.Error("Verify() should return true")
	}

	ok, err := h.CheckConstraint(ctx, struct{ Value int }{Value: 100})
	if err != nil {
		t.Errorf("CheckConstraint() with custom constraint failed: %v", err)
	}
	if !ok {
		t.Error("CheckConstraint() should return true")
	}

	err = h.RecordMetric(ctx, "custom.metric", 50.0)
	if err != nil {
		t.Errorf("RecordMetric() with custom monitor failed: %v", err)
	}
}

// TestHarness_ErrorHandling 测试错误处理
func TestHarness_ErrorHandling(t *testing.T) {
	// 创建会失败的组件
	failingValidator := &mockValidator{shouldFail: true}
	failingVerifier := &mockVerifier{verifyError: errors.New("verification error")}
	failingConstraint := &mockConstraint{checkError: errors.New("constraint error")}
	failingMonitor := &mockMonitor{err: errors.New("monitor error")}

	t.Run("验证器错误", func(t *testing.T) {
		h := NewHarnessWithComponents(failingValidator, NewDefaultVerifier(), NewDefaultConstraint(), NewDefaultMonitor())
		ctx := context.Background()

		err := h.Validate(ctx, struct{ Name string }{Name: "test"})
		if err == nil {
			t.Error("Expected validation error")
		}
	})

	t.Run("验证器错误", func(t *testing.T) {
		h := NewHarnessWithComponents(NewDefaultValidator(), failingVerifier, NewDefaultConstraint(), NewDefaultMonitor())
		ctx := context.Background()

		_, err := h.Verify(ctx, "expected", "actual")
		if err == nil {
			t.Error("Expected verification error")
		}
	})

	t.Run("约束错误", func(t *testing.T) {
		h := NewHarnessWithComponents(NewDefaultValidator(), NewDefaultVerifier(), failingConstraint, NewDefaultMonitor())
		ctx := context.Background()

		_, err := h.CheckConstraint(ctx, struct{ Value int }{Value: 100})
		if err == nil {
			t.Error("Expected constraint error")
		}
	})

	t.Run("监控器错误", func(t *testing.T) {
		h := NewHarnessWithComponents(NewDefaultValidator(), NewDefaultVerifier(), NewDefaultConstraint(), failingMonitor)
		ctx := context.Background()

		err := h.RecordMetric(ctx, "test", 100.0)
		if err == nil {
			t.Error("Expected monitor error")
		}

		_, err = h.GetMetrics(ctx, "")
		if err == nil {
			t.Error("Expected monitor error")
		}
	})
}

// TestHarness_ConcurrentAccess 测试并发访问
func TestHarness_ConcurrentAccess(t *testing.T) {
	h := NewHarness()
	ctx := context.Background()

	// 并发执行多个操作
	done := make(chan bool)

	// 并发验证
	for i := 0; i < 5; i++ {
		go func(id int) {
			defer func() { done <- true }()
			input := map[string]int{"id": id}
			_ = h.Validate(ctx, input)
		}(i)
	}

	// 并发记录指标
	for i := 0; i < 5; i++ {
		go func(id int) {
			defer func() { done <- true }()
			_ = h.RecordMetric(ctx, "concurrent_metric", float64(id))
		}(i)
	}

	// 并发验证输出
	for i := 0; i < 5; i++ {
		go func(id int) {
			defer func() { done <- true }()
			_, _ = h.Verify(ctx, id, id)
		}(i)
	}

	// 等待所有操作完成
	for i := 0; i < 15; i++ {
		<-done
	}

	// 验证指标已记录
	metrics, err := h.GetMetrics(ctx, "concurrent_metric")
	if err != nil {
		t.Fatalf("GetMetrics() failed: %v", err)
	}

	if len(metrics) != 5 {
		t.Errorf("Expected 5 metrics, got %d", len(metrics))
	}
}
