package operation

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestOperationParser 测试操作解析器
func TestOperationParser(t *testing.T) {
	parser := NewOperationParser()
	ctx := context.Background()

	tests := []struct {
		name     string
		text     string
		wantType OperationType
		wantErr  bool
	}{
		{
			name:     "遥控启动设备",
			text:     "启动设备DEV-001",
			wantType: OperationTypeRemoteControl,
			wantErr:  false,
		},
		{
			name:     "设点操作",
			text:     "设置测点POINT-001的值为100kW",
			wantType: OperationTypeSetPoint,
			wantErr:  false,
		},
		{
			name:     "调节操作",
			text:     "调整逆变器INV-001功率增加50kW",
			wantType: OperationTypeAdjust,
			wantErr:  false,
		},
		{
			name:     "查询操作",
			text:     "查询设备DEV-002的状态",
			wantType: OperationTypeQuery,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(ctx, tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(result.Operations) == 0 {
					t.Error("Parse() returned no operations")
					return
				}

				if result.Operations[0].Type != tt.wantType {
					t.Errorf("Parse() type = %v, want %v", result.Operations[0].Type, tt.wantType)
				}

				t.Logf("解析成功: %s -> %s (置信度: %.2f)", 
					tt.text, result.Operations[0].Description, result.Operations[0].Confidence)
			}
		})
	}
}

// TestOperationExecutor 测试操作执行器
func TestOperationExecutor(t *testing.T) {
	config := DefaultExecutorConfig()
	executor := NewOperationExecutor(config)

	// 注册模拟处理器
	executor.RegisterHandler(OperationTypeQuery, &MockQueryHandler{})

	ctx := context.Background()
	executor.Start(ctx)
	defer executor.Stop()

	// 提交操作
	op := &ParsedOperation{
		ID:         "test-op-001",
		Type:       OperationTypeQuery,
		Action:     "query",
		TargetID:   "DEV-001",
		TargetType: "device",
		Parameters: map[string]interface{}{
			"value": 100.0,
		},
		Constraints: &OperationConstraints{
			Timeout:     5 * time.Second,
			MaxRetries:  1,
		},
	}

	if err := executor.Submit(ctx, op); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	// 等待完成
	record, err := executor.WaitForCompletion(ctx, op.ID, 10*time.Second)
	if err != nil {
		t.Fatalf("WaitForCompletion() error = %v", err)
	}

	t.Logf("操作状态: %s, 耗时: %v", record.Status, record.Duration)

	if record.Status != StatusSuccess {
		t.Errorf("Expected status %s, got %s", StatusSuccess, record.Status)
	}
}

// TestConfirmationManager 测试确认管理器
func TestConfirmationManager(t *testing.T) {
	config := DefaultConfirmationConfig()
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(config, executor)

	ctx := context.Background()

	// 创建需要确认的操作
	op := &ParsedOperation{
		ID:         "test-op-002",
		Type:       OperationTypeRemoteControl,
		Action:     "switch_on",
		TargetID:   "DEV-001",
		TargetType: "device",
		Constraints: &OperationConstraints{
			RequireConfirm:  true,
			RequireAuth:     true,
			AuthLevel:       2,
			AllowRollback:   true,
		},
	}

	// 创建确认请求
	confirmRec, err := confirmer.CreateConfirmation(ctx, op)
	if err != nil {
		t.Fatalf("CreateConfirmation() error = %v", err)
	}

	t.Logf("确认ID: %s, 确认码: %s", confirmRec.ID, confirmRec.ConfirmCode)

	// 第一步确认
	err = confirmer.FirstStepConfirm(ctx, confirmRec.ID, confirmRec.ConfirmCode, "user-001", 2)
	if err != nil {
		t.Fatalf("FirstStepConfirm() error = %v", err)
	}

	// 检查状态
	confirmRec, _ = confirmer.GetConfirmation(confirmRec.ID)
	if confirmRec.State != ConfirmationStateFirstStep {
		t.Errorf("Expected state %s, got %s", ConfirmationStateFirstStep, confirmRec.State)
	}

	// 第二步确认
	err = confirmer.SecondStepConfirm(ctx, confirmRec.ID, "user-002", 2)
	if err != nil {
		t.Fatalf("SecondStepConfirm() error = %v", err)
	}

	// 检查最终状态
	confirmRec, _ = confirmer.GetConfirmation(confirmRec.ID)
	if confirmRec.State != ConfirmationStateConfirmed {
		t.Errorf("Expected state %s, got %s", ConfirmationStateConfirmed, confirmRec.State)
	}

	t.Log("两步确认流程测试通过")
}

// TestOperationAPI 测试操作API
func TestOperationAPI(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)

	// 注册处理器
	executor.RegisterHandler(OperationTypeQuery, &MockQueryHandler{})
	executor.RegisterHandler(OperationTypeSetPoint, &MockSetPointHandler{})

	ctx := context.Background()
	executor.Start(ctx)
	defer executor.Stop()

	api := NewOperationAPI(parser, executor, confirmer, &MockAuthChecker{})

	// 测试提交操作
	req := &OperationRequest{
		Text:      "查询设备DEV-001的状态",
		UserID:    "user-001",
		IPAddress: "192.168.1.100",
		UserAgent: "test-client",
	}

	response, err := api.SubmitOperation(ctx, req)
	if err != nil {
		t.Fatalf("SubmitOperation() error = %v", err)
	}

	if !response.Success {
		t.Errorf("SubmitOperation() failed: %s", response.Message)
	}

	t.Logf("操作提交成功: %s", response.Message)
	t.Logf("解析到 %d 个操作", len(response.Operations))

	// 测试查询状态
	if len(response.Operations) > 0 {
		statusReq := &StatusRequest{
			OperationID: response.Operations[0].ID,
		}

		// 等待操作完成
		time.Sleep(100 * time.Millisecond)

		statusResp, err := api.GetOperationStatus(ctx, statusReq)
		if err != nil {
			t.Fatalf("GetOperationStatus() error = %v", err)
		}

		t.Logf("操作状态: %s", statusResp.Status)
	}
}

// TestSafetyChecker 测试安全校验器
func TestSafetyChecker(t *testing.T) {
	checker := NewSafetyChecker()

	// 添加受保护目标
	checker.AddProtectedTarget("CRITICAL-001")

	op := &ParsedOperation{
		ID:           "test-op-003",
		Type:         OperationTypeRemoteControl,
		Action:       "shutdown",
		TargetID:     "CRITICAL-001",
		OriginalText: "关闭设备CRITICAL-001",
		Parameters: map[string]interface{}{
			"value": 200.0,
		},
		Constraints: &OperationConstraints{
			MinValue: 0,
			MaxValue: 100,
		},
	}

	warnings := checker.Check(op)

	if len(warnings) == 0 {
		t.Error("Expected safety warnings, got none")
	}

	for _, warning := range warnings {
		t.Logf("安全警告: %s", warning)
	}
}

// MockQueryHandler 模拟查询处理器
type MockQueryHandler struct{}

func (h *MockQueryHandler) Handle(ctx context.Context, op *ParsedOperation) (interface{}, error) {
	return map[string]interface{}{
		"device_id": op.TargetID,
		"status":    "online",
		"power":     op.Parameters["value"],
	}, nil
}

func (h *MockQueryHandler) CanHandle(op *ParsedOperation) bool {
	return op.Type == OperationTypeQuery
}

func (h *MockQueryHandler) Rollback(ctx context.Context, op *ParsedOperation, result interface{}) error {
	return nil
}

// MockSetPointHandler 模拟设点处理器
type MockSetPointHandler struct{}

func (h *MockSetPointHandler) Handle(ctx context.Context, op *ParsedOperation) (interface{}, error) {
	return map[string]interface{}{
		"point_id": op.TargetID,
		"value":    op.Parameters["value"],
		"success":  true,
	}, nil
}

func (h *MockSetPointHandler) CanHandle(op *ParsedOperation) bool {
	return op.Type == OperationTypeSetPoint
}

func (h *MockSetPointHandler) Rollback(ctx context.Context, op *ParsedOperation, result interface{}) error {
	// 模拟回滚
	fmt.Printf("回滚操作: %s\n", op.ID)
	return nil
}

// MockAuthChecker 模拟授权检查器
type MockAuthChecker struct{}

func (c *MockAuthChecker) CheckPermission(ctx context.Context, userID, permission string) (bool, error) {
	return true, nil
}

func (c *MockAuthChecker) GetAuthLevel(ctx context.Context, userID string) (int, error) {
	return 3, nil
}

func (c *MockAuthChecker) GetUserInfo(ctx context.Context, userID string) (*UserInfo, error) {
	return &UserInfo{
		ID:        userID,
		Username:  "test-user",
		AuthLevel: 3,
	}, nil
}

// ExampleOperationParser 示例：使用操作解析器
func ExampleOperationParser() {
	parser := NewOperationParser()
	ctx := context.Background()

	// 解析自然语言指令
	result, err := parser.Parse(ctx, "启动逆变器INV-001，功率设置为500kW")
	if err != nil {
		fmt.Printf("解析失败: %v\n", err)
		return
	}

	fmt.Printf("解析到 %d 个操作:\n", len(result.Operations))
	for i, op := range result.Operations {
		fmt.Printf("%d. %s (类型: %s, 置信度: %.2f)\n", 
			i+1, op.Description, op.Type, op.Confidence)
	}

	// 输出警告和建议
	if len(result.Warnings) > 0 {
		fmt.Println("\n警告:")
		for _, w := range result.Warnings {
			fmt.Printf("  - %s\n", w)
		}
	}

	if len(result.Suggestions) > 0 {
		fmt.Println("\n建议:")
		for _, s := range result.Suggestions {
			fmt.Printf("  - %s\n", s)
		}
	}
}

// ExampleOperationExecutor 示例：使用操作执行器
func ExampleOperationExecutor() {
	// 创建执行器
	config := &ExecutorConfig{
		QueueCapacity:   100,
		MaxWorkers:      5,
		DefaultTimeout:  30 * time.Second,
		RetryDelay:      1 * time.Second,
		MaxRetryDelay:   10 * time.Second,
		EnablePriority:  true,
		HistoryCapacity: 1000,
	}
	executor := NewOperationExecutor(config)

	// 注册处理器
	executor.RegisterHandler(OperationTypeSetPoint, &MockSetPointHandler{})

	// 启动执行器
	ctx := context.Background()
	executor.Start(ctx)
	defer executor.Stop()

	// 提交操作
	op := &ParsedOperation{
		ID:       "example-op-001",
		Type:     OperationTypeSetPoint,
		Priority: PriorityNormal,
		Action:   "set_value",
		TargetID: "POINT-001",
		Parameters: map[string]interface{}{
			"value": 100.0,
			"unit":  "kW",
		},
		Constraints: &OperationConstraints{
			Timeout:       10 * time.Second,
			MaxRetries:    3,
			AllowRollback: true,
		},
	}

	if err := executor.Submit(ctx, op); err != nil {
		fmt.Printf("提交失败: %v\n", err)
		return
	}

	// 等待完成
	record, err := executor.WaitForCompletion(ctx, op.ID, 15*time.Second)
	if err != nil {
		fmt.Printf("等待失败: %v\n", err)
		return
	}

	fmt.Printf("操作完成: 状态=%s, 耗时=%v\n", record.Status, record.Duration)

	// 获取统计信息
	stats := executor.GetStats()
	fmt.Printf("执行器统计: 队列=%d, 运行中=%d, 总计=%d\n", 
		stats.QueueLength, stats.RunningCount, stats.TotalExecuted)
}

// ExampleConfirmationManager 示例：使用确认管理器
func ExampleConfirmationManager() {
	// 创建确认管理器
	config := &ConfirmationConfig{
		DefaultTimeout:    10 * time.Minute,
		EnableTwoStep:     true,
		CodeLength:        6,
		MaxPendingConfirm: 100,
		AuditLogCapacity:  1000,
	}
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(config, executor)

	ctx := context.Background()

	// 创建需要确认的操作
	op := &ParsedOperation{
		ID:       "example-op-002",
		Type:     OperationTypeRemoteControl,
		Action:   "switch_off",
		TargetID: "DEV-001",
		Constraints: &OperationConstraints{
			RequireConfirm:  true,
			RequireAuth:     true,
			AuthLevel:       2,
			AllowRollback:   true,
		},
	}

	// 创建确认请求
	confirmRec, err := confirmer.CreateConfirmation(ctx, op)
	if err != nil {
		fmt.Printf("创建确认失败: %v\n", err)
		return
	}

	fmt.Printf("确认请求已创建:\n")
	fmt.Printf("  确认ID: %s\n", confirmRec.ID)
	fmt.Printf("  确认码: %s\n", confirmRec.ConfirmCode)
	fmt.Printf("  过期时间: %s\n", confirmRec.ExpiresAt.Format("2006-01-02 15:04:05"))

	// 第一步确认
	err = confirmer.FirstStepConfirm(ctx, confirmRec.ID, confirmRec.ConfirmCode, "operator-001", 2)
	if err != nil {
		fmt.Printf("第一步确认失败: %v\n", err)
		return
	}
	fmt.Println("第一步确认成功")

	// 第二步确认（需要不同用户）
	err = confirmer.SecondStepConfirm(ctx, confirmRec.ID, "supervisor-001", 3)
	if err != nil {
		fmt.Printf("第二步确认失败: %v\n", err)
		return
	}
	fmt.Println("第二步确认成功，操作已授权执行")

	// 查看审计日志
	logs := confirmer.GetAuditLogs(op.ID, 10)
	fmt.Printf("\n审计日志 (%d 条):\n", len(logs))
	for _, log := range logs {
		fmt.Printf("  [%s] %s - %s\n", 
			log.Timestamp.Format("15:04:05"), log.Action, log.UserID)
	}
}
