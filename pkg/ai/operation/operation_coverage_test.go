package operation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOperationParser_Parse_EmptyText(t *testing.T) {
	parser := NewOperationParser()
	_, err := parser.Parse(context.Background(), "")
	assert.Error(t, err)
}

func TestOperationParser_Parse_WhitespaceText(t *testing.T) {
	parser := NewOperationParser()
	_, err := parser.Parse(context.Background(), "   ")
	assert.Error(t, err)
}

func TestOperationParser_Parse_MultipleCommands(t *testing.T) {
	parser := NewOperationParser()
	result, err := parser.Parse(context.Background(), "启动设备DEV-001然后关闭设备DEV-002")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result.Operations), 1)
}

func TestOperationParser_Parse_StationTarget(t *testing.T) {
	parser := NewOperationParser()
	result, err := parser.Parse(context.Background(), "查询电站STATION-001的状态")
	require.NoError(t, err)
	assert.NotEmpty(t, result.Operations)
	assert.Equal(t, "station", result.Operations[0].TargetType)
}

func TestOperationParser_Parse_InverterTarget(t *testing.T) {
	parser := NewOperationParser()
	result, err := parser.Parse(context.Background(), "调整逆变器INV-001功率")
	require.NoError(t, err)
	assert.NotEmpty(t, result.Operations)
	assert.Equal(t, "device", result.Operations[0].TargetType)
}

func TestOperationParser_Parse_WithTimeParam(t *testing.T) {
	parser := NewOperationParser()
	result, err := parser.Parse(context.Background(), "设置设备DEV-001在14:30启动")
	require.NoError(t, err)
	assert.NotEmpty(t, result.Operations)
	_, hasTime := result.Operations[0].Parameters["time"]
	assert.True(t, hasTime)
}

func TestOperationParser_Parse_WithDelayParam(t *testing.T) {
	parser := NewOperationParser()
	result, err := parser.Parse(context.Background(), "延迟5秒启动设备DEV-001")
	require.NoError(t, err)
	assert.NotEmpty(t, result.Operations)
}

func TestOperationParser_Parse_WithUnitParam(t *testing.T) {
	parser := NewOperationParser()
	result, err := parser.Parse(context.Background(), "设置功率为500kW")
	require.NoError(t, err)
	assert.NotEmpty(t, result.Operations)
}

func TestOperationParser_Parse_NoTarget(t *testing.T) {
	parser := NewOperationParser()
	result, err := parser.Parse(context.Background(), "查询所有状态")
	require.NoError(t, err)
	assert.NotEmpty(t, result.Operations)
	assert.Empty(t, result.Operations[0].TargetID)
}

func TestOperationParser_ValidateOperation_Valid(t *testing.T) {
	parser := NewOperationParser()
	op := &ParsedOperation{
		Type:     OperationTypeQuery,
		Action:   "query",
		TargetID: "DEV-001",
	}
	err := parser.ValidateOperation(op)
	assert.NoError(t, err)
}

func TestOperationParser_ValidateOperation_NoType(t *testing.T) {
	parser := NewOperationParser()
	op := &ParsedOperation{Action: "query", TargetID: "DEV-001"}
	err := parser.ValidateOperation(op)
	assert.Error(t, err)
}

func TestOperationParser_ValidateOperation_NoTarget(t *testing.T) {
	parser := NewOperationParser()
	op := &ParsedOperation{Type: OperationTypeRemoteControl, Action: "control"}
	err := parser.ValidateOperation(op)
	assert.Error(t, err)
}

func TestOperationParser_ValidateOperation_QueryNoTarget(t *testing.T) {
	parser := NewOperationParser()
	op := &ParsedOperation{Type: OperationTypeQuery, Action: "query"}
	err := parser.ValidateOperation(op)
	assert.NoError(t, err)
}

func TestOperationParser_ValidateOperation_NoAction(t *testing.T) {
	parser := NewOperationParser()
	op := &ParsedOperation{Type: OperationTypeQuery, TargetID: "DEV-001"}
	err := parser.ValidateOperation(op)
	assert.Error(t, err)
}

func TestOperationParser_ValidateOperation_InvalidValue(t *testing.T) {
	parser := NewOperationParser()
	op := &ParsedOperation{
		Type:       OperationTypeSetPoint,
		Action:     "set",
		TargetID:   "PT-001",
		Parameters: map[string]interface{}{"value": "not_a_number"},
	}
	err := parser.ValidateOperation(op)
	assert.Error(t, err)
}

func TestSafetyChecker_RemoveProtectedTarget(t *testing.T) {
	checker := NewSafetyChecker()
	checker.AddProtectedTarget("TARGET-001")
	op := &ParsedOperation{
		TargetID:     "TARGET-001",
		OriginalText: "操作TARGET-001",
	}
	warnings := checker.Check(op)
	assert.NotEmpty(t, warnings)
	checker.RemoveProtectedTarget("TARGET-001")
	warnings2 := checker.Check(op)
	assert.Empty(t, warnings2)
}

func TestSafetyChecker_NoConstraints(t *testing.T) {
	checker := NewSafetyChecker()
	op := &ParsedOperation{
		TargetID:     "DEV-001",
		OriginalText: "查询设备",
	}
	warnings := checker.Check(op)
	assert.Empty(t, warnings)
}

func TestSafetyChecker_ValueOutOfRange(t *testing.T) {
	checker := NewSafetyChecker()
	op := &ParsedOperation{
		TargetID:     "DEV-001",
		OriginalText: "设置设备",
		Parameters:   map[string]interface{}{"value": 200.0},
		Constraints:  &OperationConstraints{MinValue: 0, MaxValue: 100},
	}
	warnings := checker.Check(op)
	assert.NotEmpty(t, warnings)
}

func TestOperationQueue_PushPop(t *testing.T) {
	q := NewOperationQueue(10)
	op := &ParsedOperation{ID: "q-1"}
	err := q.Push(op)
	require.NoError(t, err)
	assert.Equal(t, 1, q.Len())
	popped := q.Pop()
	assert.NotNil(t, popped)
	assert.Equal(t, "q-1", popped.ID)
	assert.Equal(t, 0, q.Len())
}

func TestOperationQueue_PopEmpty(t *testing.T) {
	q := NewOperationQueue(10)
	assert.Nil(t, q.Pop())
}

func TestOperationQueue_Peek(t *testing.T) {
	q := NewOperationQueue(10)
	q.Push(&ParsedOperation{ID: "q-1"})
	q.Push(&ParsedOperation{ID: "q-2"})
	peeked := q.Peek()
	assert.NotNil(t, peeked)
	assert.Equal(t, "q-1", peeked.ID)
	assert.Equal(t, 2, q.Len())
}

func TestOperationQueue_PeekEmpty(t *testing.T) {
	q := NewOperationQueue(10)
	assert.Nil(t, q.Peek())
}

func TestOperationQueue_PushFull(t *testing.T) {
	q := NewOperationQueue(2)
	q.Push(&ParsedOperation{ID: "1"})
	q.Push(&ParsedOperation{ID: "2"})
	err := q.Push(&ParsedOperation{ID: "3"})
	assert.Equal(t, ErrQueueFull, err)
}

func TestOperationQueue_Notify(t *testing.T) {
	q := NewOperationQueue(10)
	notifyCh := q.Notify()
	assert.NotNil(t, notifyCh)
	q.Push(&ParsedOperation{ID: "1"})
	select {
	case <-notifyCh:
	default:
		t.Error("expected notification")
	}
}

func TestDefaultExecutorConfig(t *testing.T) {
	cfg := DefaultExecutorConfig()
	assert.Equal(t, 1000, cfg.QueueCapacity)
	assert.Equal(t, 10, cfg.MaxWorkers)
	assert.Equal(t, 30*time.Second, cfg.DefaultTimeout)
	assert.Equal(t, 1*time.Second, cfg.RetryDelay)
	assert.Equal(t, 30*time.Second, cfg.MaxRetryDelay)
	assert.True(t, cfg.EnablePriority)
	assert.Equal(t, 10000, cfg.HistoryCapacity)
}

func TestOperationExecutor_SubmitShutdown(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	executor.Start(context.Background())
	executor.Stop()
	op := &ParsedOperation{ID: "shut-1", Type: OperationTypeQuery, Action: "query"}
	err := executor.Submit(context.Background(), op)
	assert.Equal(t, ErrExecutorShutdown, err)
}

func TestOperationExecutor_SubmitDuplicate(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	op := &ParsedOperation{ID: "dup-1", Type: OperationTypeQuery, Action: "query"}
	err := executor.Submit(context.Background(), op)
	require.NoError(t, err)
	err = executor.Submit(context.Background(), op)
	assert.Equal(t, ErrOperationAlreadyExists, err)
}

func TestOperationExecutor_GetStatusNotFound(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	_, err := executor.GetStatus("nonexistent")
	assert.Equal(t, ErrOperationNotFound, err)
}

func TestOperationExecutor_CancelNotFound(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	err := executor.Cancel("nonexistent")
	assert.Equal(t, ErrOperationNotFound, err)
}

func TestOperationExecutor_CancelPending(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	op := &ParsedOperation{ID: "cancel-1", Type: OperationTypeQuery, Action: "query"}
	executor.Submit(context.Background(), op)
	err := executor.Cancel("cancel-1")
	require.NoError(t, err)
	record, _ := executor.GetStatus("cancel-1")
	assert.Equal(t, StatusCancelled, record.Status)
}

func TestOperationExecutor_CancelWrongStatus(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	executor.RegisterHandler(OperationTypeQuery, &MockQueryHandler{})
	executor.Start(context.Background())
	defer executor.Stop()
	op := &ParsedOperation{ID: "cancel-2", Type: OperationTypeQuery, Action: "query", Constraints: &OperationConstraints{Timeout: 5 * time.Second}}
	executor.Submit(context.Background(), op)
	executor.WaitForCompletion(context.Background(), op.ID, 10*time.Second)
	err := executor.Cancel("cancel-2")
	assert.Error(t, err)
}

func TestOperationExecutor_RollbackNotFound(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	err := executor.Rollback(context.Background(), "nonexistent")
	assert.Equal(t, ErrOperationNotFound, err)
}

func TestOperationExecutor_RollbackNotAllowed(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	op := &ParsedOperation{ID: "rb-1", Type: OperationTypeQuery, Action: "query", Constraints: &OperationConstraints{AllowRollback: false}}
	executor.Submit(context.Background(), op)
	err := executor.Rollback(context.Background(), "rb-1")
	assert.Error(t, err)
}

func TestOperationExecutor_RollbackNotSuccess(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	op := &ParsedOperation{ID: "rb-2", Type: OperationTypeQuery, Action: "query", Constraints: &OperationConstraints{AllowRollback: true}}
	executor.Submit(context.Background(), op)
	err := executor.Rollback(context.Background(), "rb-2")
	assert.Error(t, err)
}

func TestOperationExecutor_GetHistory(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	executor.RegisterHandler(OperationTypeQuery, &MockQueryHandler{})
	executor.Start(context.Background())
	defer executor.Stop()
	op := &ParsedOperation{ID: "hist-1", Type: OperationTypeQuery, Action: "query", Constraints: &OperationConstraints{Timeout: 5 * time.Second}}
	executor.Submit(context.Background(), op)
	executor.WaitForCompletion(context.Background(), op.ID, 10*time.Second)
	history := executor.GetHistory(10)
	assert.NotEmpty(t, history)
}

func TestOperationExecutor_GetHistory_Empty(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	history := executor.GetHistory(10)
	assert.Empty(t, history)
}

func TestOperationExecutor_GetStats(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	stats := executor.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, 0, stats.QueueLength)
	assert.Equal(t, int64(0), stats.RunningCount)
}

func TestOperationExecutor_SubmitBatch(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	ops := []*ParsedOperation{
		{ID: "batch-1", Type: OperationTypeQuery, Action: "query"},
		{ID: "batch-2", Type: OperationTypeQuery, Action: "query"},
	}
	result, err := executor.SubmitBatch(context.Background(), ops)
	require.NoError(t, err)
	assert.Equal(t, 2, result.TotalCount)
	assert.Equal(t, 2, result.Success)
}

func TestOperationExecutor_SubmitBatch_WithFailure(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	executor.Submit(context.Background(), &ParsedOperation{ID: "batch-dup", Type: OperationTypeQuery, Action: "query"})
	ops := []*ParsedOperation{
		{ID: "batch-dup", Type: OperationTypeQuery, Action: "query"},
		{ID: "batch-ok", Type: OperationTypeQuery, Action: "query"},
	}
	result, err := executor.SubmitBatch(context.Background(), ops)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Failed)
	assert.Equal(t, 1, result.Success)
}

func TestOperationExecutor_GetQueueLength(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	assert.Equal(t, 0, executor.GetQueueLength())
}

func TestOperationExecutor_GetRunningCount(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	assert.Equal(t, int64(0), executor.GetRunningCount())
}

func TestConfirmationManager_CreateConfirmation_NoRequireConfirm(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	op := &ParsedOperation{
		ID:         "no-confirm-1",
		Type:       OperationTypeQuery,
		Action:     "query",
		Constraints: &OperationConstraints{RequireConfirm: false},
	}
	record, err := confirmer.CreateConfirmation(context.Background(), op)
	require.NoError(t, err)
	assert.Equal(t, ConfirmationStateConfirmed, record.State)
}

func TestConfirmationManager_CreateConfirmation_NoConstraints(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	op := &ParsedOperation{ID: "no-constraint-1", Type: OperationTypeQuery, Action: "query"}
	record, err := confirmer.CreateConfirmation(context.Background(), op)
	require.NoError(t, err)
	assert.Equal(t, ConfirmationStateConfirmed, record.State)
}

func TestConfirmationManager_FirstStepConfirm_NotFound(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	err := confirmer.FirstStepConfirm(context.Background(), "nonexistent", "123456", "user1", 1)
	assert.Equal(t, ErrConfirmationNotFound, err)
}

func TestConfirmationManager_FirstStepConfirm_Expired(t *testing.T) {
	cfg := &ConfirmationConfig{DefaultTimeout: 1 * time.Nanosecond, EnableTwoStep: true, CodeLength: 6, MaxPendingConfirm: 100, AuditLogCapacity: 1000}
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(cfg, executor)
	op := &ParsedOperation{ID: "expired-1", Type: OperationTypeRemoteControl, Action: "switch_on", Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 2}}
	record, err := confirmer.CreateConfirmation(context.Background(), op)
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)
	err = confirmer.FirstStepConfirm(context.Background(), record.ID, record.ConfirmCode, "user1", 2)
	assert.Equal(t, ErrConfirmationExpired, err)
}

func TestConfirmationManager_FirstStepConfirm_WrongCode(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	op := &ParsedOperation{ID: "wrong-code-1", Type: OperationTypeRemoteControl, Action: "switch_on", Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 2}}
	record, err := confirmer.CreateConfirmation(context.Background(), op)
	require.NoError(t, err)
	err = confirmer.FirstStepConfirm(context.Background(), record.ID, "wrong-code", "user1", 2)
	assert.Equal(t, ErrInvalidConfirmCode, err)
}

func TestConfirmationManager_FirstStepConfirm_InsufficientAuth(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	op := &ParsedOperation{ID: "insuff-auth-1", Type: OperationTypeRemoteControl, Action: "switch_on", Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 3}}
	record, err := confirmer.CreateConfirmation(context.Background(), op)
	require.NoError(t, err)
	err = confirmer.FirstStepConfirm(context.Background(), record.ID, record.ConfirmCode, "user1", 1)
	assert.Equal(t, ErrInsufficientAuthLevel, err)
}

func TestConfirmationManager_FirstStepConfirm_AlreadyUsed(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(&ConfirmationConfig{DefaultTimeout: 1 * time.Hour, EnableTwoStep: false, CodeLength: 6, MaxPendingConfirm: 100, AuditLogCapacity: 1000}, executor)
	op := &ParsedOperation{ID: "used-1", Type: OperationTypeRemoteControl, Action: "switch_on", Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1}}
	record, err := confirmer.CreateConfirmation(context.Background(), op)
	require.NoError(t, err)
	err = confirmer.FirstStepConfirm(context.Background(), record.ID, record.ConfirmCode, "user1", 1)
	require.NoError(t, err)
	err = confirmer.FirstStepConfirm(context.Background(), record.ID, record.ConfirmCode, "user1", 1)
	assert.Equal(t, ErrConfirmationAlreadyUsed, err)
}

func TestConfirmationManager_SecondStepConfirm_NotFound(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	err := confirmer.SecondStepConfirm(context.Background(), "nonexistent", "user2", 2)
	assert.Equal(t, ErrConfirmationNotFound, err)
}

func TestConfirmationManager_SecondStepConfirm_WrongState(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	op := &ParsedOperation{ID: "ws-1", Type: OperationTypeRemoteControl, Action: "switch_on", Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 2}}
	record, _ := confirmer.CreateConfirmation(context.Background(), op)
	err := confirmer.SecondStepConfirm(context.Background(), record.ID, "user2", 2)
	assert.Error(t, err)
}

func TestConfirmationManager_SecondStepConfirm_SameUser(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(&ConfirmationConfig{DefaultTimeout: 1 * time.Hour, EnableTwoStep: true, CodeLength: 6, MaxPendingConfirm: 100, AuditLogCapacity: 1000}, executor)
	op := &ParsedOperation{ID: "same-user-1", Type: OperationTypeRemoteControl, Action: "switch_on", Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 2}}
	record, _ := confirmer.CreateConfirmation(context.Background(), op)
	confirmer.FirstStepConfirm(context.Background(), record.ID, record.ConfirmCode, "user1", 2)
	err := confirmer.SecondStepConfirm(context.Background(), record.ID, "user1", 2)
	assert.Error(t, err)
}

func TestConfirmationManager_Reject(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	op := &ParsedOperation{ID: "reject-1", Type: OperationTypeRemoteControl, Action: "switch_on", Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 2}}
	record, _ := confirmer.CreateConfirmation(context.Background(), op)
	err := confirmer.Reject(context.Background(), record.ID, "admin", "too dangerous")
	require.NoError(t, err)
	updated, _ := confirmer.GetConfirmation(record.ID)
	assert.Equal(t, ConfirmationStateRejected, updated.State)
}

func TestConfirmationManager_Reject_NotFound(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	err := confirmer.Reject(context.Background(), "nonexistent", "admin", "reason")
	assert.Equal(t, ErrConfirmationNotFound, err)
}

func TestConfirmationManager_Cancel(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	op := &ParsedOperation{ID: "cancel-c-1", Type: OperationTypeRemoteControl, Action: "switch_on", Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 2}}
	record, _ := confirmer.CreateConfirmation(context.Background(), op)
	err := confirmer.Cancel(context.Background(), record.ID, "admin")
	require.NoError(t, err)
	updated, _ := confirmer.GetConfirmation(record.ID)
	assert.Equal(t, ConfirmationStateCancelled, updated.State)
}

func TestConfirmationManager_Cancel_NotFound(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	err := confirmer.Cancel(context.Background(), "nonexistent", "admin")
	assert.Equal(t, ErrConfirmationNotFound, err)
}

func TestConfirmationManager_IsConfirmed(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(&ConfirmationConfig{DefaultTimeout: 1 * time.Hour, EnableTwoStep: false, CodeLength: 6, MaxPendingConfirm: 100, AuditLogCapacity: 1000}, executor)
	op := &ParsedOperation{ID: "is-conf-1", Type: OperationTypeRemoteControl, Action: "switch_on", Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1}}
	record, _ := confirmer.CreateConfirmation(context.Background(), op)
	confirmed, _ := confirmer.IsConfirmed(record.ID)
	assert.False(t, confirmed)
	confirmer.FirstStepConfirm(context.Background(), record.ID, record.ConfirmCode, "user1", 1)
	confirmed, _ = confirmer.IsConfirmed(record.ID)
	assert.True(t, confirmed)
}

func TestConfirmationManager_Authorize(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(&ConfirmationConfig{DefaultTimeout: 1 * time.Hour, EnableTwoStep: false, CodeLength: 6, MaxPendingConfirm: 100, AuditLogCapacity: 1000}, executor)
	op := &ParsedOperation{ID: "auth-1", Type: OperationTypeRemoteControl, Action: "switch_on", Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1}}
	record, _ := confirmer.CreateConfirmation(context.Background(), op)
	confirmer.FirstStepConfirm(context.Background(), record.ID, record.ConfirmCode, "user1", 1)
	err := confirmer.Authorize(context.Background(), record.ID, "admin", 2)
	require.NoError(t, err)
}

func TestConfirmationManager_Authorize_NotConfirmed(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	op := &ParsedOperation{ID: "auth-nc-1", Type: OperationTypeRemoteControl, Action: "switch_on", Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 2}}
	record, _ := confirmer.CreateConfirmation(context.Background(), op)
	err := confirmer.Authorize(context.Background(), record.ID, "admin", 2)
	assert.Error(t, err)
}

func TestConfirmationManager_Authorize_InsufficientAuth(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(&ConfirmationConfig{DefaultTimeout: 1 * time.Hour, EnableTwoStep: false, CodeLength: 6, MaxPendingConfirm: 100, AuditLogCapacity: 1000}, executor)
	op := &ParsedOperation{ID: "auth-ia-1", Type: OperationTypeRemoteControl, Action: "switch_on", Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1}}
	record, _ := confirmer.CreateConfirmation(context.Background(), op)
	confirmer.FirstStepConfirm(context.Background(), record.ID, record.ConfirmCode, "user1", 1)
	err := confirmer.Authorize(context.Background(), record.ID, "admin", 0)
	assert.Equal(t, ErrInsufficientAuthLevel, err)
}

func TestConfirmationManager_GetConfirmationByOperation(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	op := &ParsedOperation{ID: "by-op-1", Type: OperationTypeRemoteControl, Action: "switch_on", Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 2}}
	confirmer.CreateConfirmation(context.Background(), op)
	record, err := confirmer.GetConfirmationByOperation("by-op-1")
	require.NoError(t, err)
	assert.Equal(t, "by-op-1", record.OperationID)
}

func TestConfirmationManager_GetConfirmationByOperation_NotFound(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	_, err := confirmer.GetConfirmationByOperation("nonexistent")
	assert.Equal(t, ErrConfirmationNotFound, err)
}

func TestConfirmationManager_CleanupExpired(t *testing.T) {
	cfg := &ConfirmationConfig{DefaultTimeout: 1 * time.Nanosecond, EnableTwoStep: true, CodeLength: 6, MaxPendingConfirm: 100, AuditLogCapacity: 1000}
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(cfg, executor)
	op := &ParsedOperation{ID: "cleanup-1", Type: OperationTypeRemoteControl, Action: "switch_on", Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 2}}
	confirmer.CreateConfirmation(context.Background(), op)
	time.Sleep(10 * time.Millisecond)
	count := confirmer.CleanupExpired()
	assert.Equal(t, 1, count)
}

func TestConfirmationManager_GetPendingConfirmations(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	op := &ParsedOperation{ID: "pending-1", Type: OperationTypeRemoteControl, Action: "switch_on", Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 2}}
	confirmer.CreateConfirmation(context.Background(), op)
	pending := confirmer.GetPendingConfirmations()
	assert.Len(t, pending, 1)
}

func TestConfirmationManager_GetStats(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	op := &ParsedOperation{ID: "stats-1", Type: OperationTypeRemoteControl, Action: "switch_on", Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 2}}
	confirmer.CreateConfirmation(context.Background(), op)
	stats := confirmer.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, 1, stats.ByState[ConfirmationStatePending])
}

func TestConfirmationManager_Rollback(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	executor.RegisterHandler(OperationTypeRemoteControl, &MockRemoteControlHandler{})
	executor.Start(context.Background())
	defer executor.Stop()
	confirmer := NewConfirmationManager(&ConfirmationConfig{DefaultTimeout: 1 * time.Hour, EnableTwoStep: false, CodeLength: 6, MaxPendingConfirm: 100, AuditLogCapacity: 1000}, executor)
	op := &ParsedOperation{ID: "rb-conf-1", Type: OperationTypeRemoteControl, Action: "switch_on", Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1, AllowRollback: true, Timeout: 5 * time.Second}}
	confirmer.CreateConfirmation(context.Background(), op)
	record, _ := confirmer.GetConfirmationByOperation("rb-conf-1")
	confirmer.FirstStepConfirm(context.Background(), record.ID, record.ConfirmCode, "user1", 1)
	executor.Submit(context.Background(), op)
	executor.WaitForCompletion(context.Background(), op.ID, 10*time.Second)
	err := confirmer.Rollback(context.Background(), "rb-conf-1", "admin")
	require.NoError(t, err)
}

func TestConfirmationManager_Rollback_NotFound(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	err := confirmer.Rollback(context.Background(), "nonexistent", "admin")
	assert.Equal(t, ErrConfirmationNotFound, err)
}

func TestConfirmationManager_Rollback_NotAllowed(t *testing.T) {
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	op := &ParsedOperation{ID: "rb-na-1", Type: OperationTypeRemoteControl, Action: "switch_on", Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1, AllowRollback: false}}
	confirmer.CreateConfirmation(context.Background(), op)
	err := confirmer.Rollback(context.Background(), "rb-na-1", "admin")
	assert.Error(t, err)
}

func TestOperationAPI_SubmitOperation_WithConstraints(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	api := NewOperationAPI(parser, executor, confirmer, &MockAuthChecker{})
	req := &OperationRequest{
		Text:    "查询设备DEV-001",
		UserID:  "user1",
		DryRun:  true,
		Constraints: &OperationConstraints{
			Timeout:    10 * time.Second,
			MaxRetries: 5,
			AuthLevel:  3,
		},
	}
	resp, err := api.SubmitOperation(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestOperationAPI_SubmitOperation_EmptyText(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	api := NewOperationAPI(parser, executor, confirmer, &MockAuthChecker{})
	req := &OperationRequest{Text: "", UserID: "user1"}
	resp, err := api.SubmitOperation(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, resp.Success)
}

func TestOperationAPI_ConfirmOperation(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(DefaultExecutorConfig())
	executor.RegisterHandler(OperationTypeQuery, &MockQueryHandler{})
	executor.Start(context.Background())
	defer executor.Stop()
	confirmer := NewConfirmationManager(&ConfirmationConfig{DefaultTimeout: 1 * time.Hour, EnableTwoStep: false, CodeLength: 6, MaxPendingConfirm: 100, AuditLogCapacity: 1000}, executor)
	api := NewOperationAPI(parser, executor, confirmer, &MockAuthChecker{})
	op := &ParsedOperation{ID: "api-conf-1", Type: OperationTypeQuery, Action: "query", TargetID: "DEV-001", Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1}}
	confirmRec, _ := confirmer.CreateConfirmation(context.Background(), op)
	req := &ConfirmRequest{ConfirmID: confirmRec.ID, ConfirmCode: confirmRec.ConfirmCode, UserID: "user1", Step: 1}
	resp, err := api.ConfirmOperation(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestOperationAPI_ConfirmOperation_NotFound(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	api := NewOperationAPI(parser, executor, confirmer, &MockAuthChecker{})
	req := &ConfirmRequest{ConfirmID: "nonexistent", ConfirmCode: "123456", UserID: "user1", Step: 1}
	resp, err := api.ConfirmOperation(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, resp.Success)
}

func TestOperationAPI_RejectOperation(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	api := NewOperationAPI(parser, executor, confirmer, &MockAuthChecker{})
	op := &ParsedOperation{ID: "api-rej-1", Type: OperationTypeRemoteControl, Action: "switch_on", Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 2}}
	confirmRec, _ := confirmer.CreateConfirmation(context.Background(), op)
	err := api.RejectOperation(context.Background(), confirmRec.ID, "admin", "dangerous")
	require.NoError(t, err)
}

func TestOperationAPI_CancelOperation(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(DefaultExecutorConfig())
	executor.RegisterHandler(OperationTypeRemoteControl, &MockRemoteControlHandler{})
	executor.Start(context.Background())
	defer executor.Stop()
	confirmer := NewConfirmationManager(&ConfirmationConfig{DefaultTimeout: 1 * time.Hour, EnableTwoStep: false, CodeLength: 6, MaxPendingConfirm: 100, AuditLogCapacity: 1000}, executor)
	api := NewOperationAPI(parser, executor, confirmer, &MockAuthChecker{})
	op := &ParsedOperation{ID: "api-cancel-1", Type: OperationTypeRemoteControl, Action: "switch_on", Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1, Timeout: 5 * time.Second}}
	confirmRec, _ := confirmer.CreateConfirmation(context.Background(), op)
	confirmer.FirstStepConfirm(context.Background(), confirmRec.ID, confirmRec.ConfirmCode, "user1", 1)
	executor.Submit(context.Background(), op)
	executor.WaitForCompletion(context.Background(), op.ID, 10*time.Second)
	err := api.CancelOperation(context.Background(), confirmRec.ID, "admin")
	_ = err
}

func TestOperationAPI_GetOperationHistory(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	api := NewOperationAPI(parser, executor, confirmer, &MockAuthChecker{})
	now := time.Now()
	resp, err := api.GetOperationHistory(context.Background(), &HistoryRequest{
		Limit:         10,
		Status:        "success",
		OperationType: "query",
		StartTime:     &now,
		EndTime:       &now,
	})
	require.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestOperationAPI_GetPendingConfirmations(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	api := NewOperationAPI(parser, executor, confirmer, &MockAuthChecker{})
	op := &ParsedOperation{ID: "api-pend-1", Type: OperationTypeRemoteControl, Action: "switch_on", Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 2}}
	confirmer.CreateConfirmation(context.Background(), op)
	pending, err := api.GetPendingConfirmations(context.Background())
	require.NoError(t, err)
	assert.Len(t, pending, 1)
}

func TestOperationAPI_GetAuditLogs(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	api := NewOperationAPI(parser, executor, confirmer, &MockAuthChecker{})
	logs, err := api.GetAuditLogs(context.Background(), "", 10)
	require.NoError(t, err)
	_ = logs
}

func TestOperationAPI_GetStats(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	api := NewOperationAPI(parser, executor, confirmer, &MockAuthChecker{})
	stats := api.GetStats(context.Background())
	assert.NotNil(t, stats)
	assert.NotNil(t, stats.Executor)
	assert.NotNil(t, stats.Confirmer)
}

func TestOperationAPI_ValidateRequest_EmptyText(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	api := NewOperationAPI(parser, executor, confirmer, &MockAuthChecker{})
	err := api.ValidateRequest(&OperationRequest{Text: "", UserID: "user1"})
	assert.Error(t, err)
}

func TestOperationAPI_ValidateRequest_EmptyUserID(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	api := NewOperationAPI(parser, executor, confirmer, &MockAuthChecker{})
	err := api.ValidateRequest(&OperationRequest{Text: "test", UserID: ""})
	assert.Error(t, err)
}

func TestOperationAPI_ValidateRequest_Valid(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(DefaultExecutorConfig())
	confirmer := NewConfirmationManager(DefaultConfirmationConfig(), executor)
	api := NewOperationAPI(parser, executor, confirmer, &MockAuthChecker{})
	err := api.ValidateRequest(&OperationRequest{Text: "test", UserID: "user1"})
	assert.NoError(t, err)
}

func TestDefaultConfirmationConfig(t *testing.T) {
	cfg := DefaultConfirmationConfig()
	assert.Equal(t, 5*time.Minute, cfg.DefaultTimeout)
	assert.True(t, cfg.EnableTwoStep)
	assert.Equal(t, 6, cfg.CodeLength)
	assert.Equal(t, 1000, cfg.MaxPendingConfirm)
	assert.Equal(t, 10000, cfg.AuditLogCapacity)
}

func TestOperationConstants(t *testing.T) {
	assert.Equal(t, OperationType("remote_control"), OperationTypeRemoteControl)
	assert.Equal(t, OperationType("setpoint"), OperationTypeSetPoint)
	assert.Equal(t, OperationType("adjust"), OperationTypeAdjust)
	assert.Equal(t, OperationType("query"), OperationTypeQuery)
	assert.Equal(t, OperationType("batch"), OperationTypeBatch)
	assert.Equal(t, OperationPriority(1), PriorityLow)
	assert.Equal(t, OperationPriority(5), PriorityNormal)
	assert.Equal(t, OperationPriority(8), PriorityHigh)
	assert.Equal(t, OperationPriority(10), PriorityCritical)
}

func TestOperationStatus_Constants(t *testing.T) {
	assert.Equal(t, OperationStatus("pending"), StatusPending)
	assert.Equal(t, OperationStatus("success"), StatusSuccess)
	assert.Equal(t, OperationStatus("failed"), StatusFailed)
	assert.Equal(t, OperationStatus("timeout"), StatusTimeout)
	assert.Equal(t, OperationStatus("cancelled"), StatusCancelled)
	assert.Equal(t, OperationStatus("rolledback"), StatusRolledBack)
}

func TestConfirmationState_Constants(t *testing.T) {
	assert.Equal(t, ConfirmationState("pending"), ConfirmationStatePending)
	assert.Equal(t, ConfirmationState("first_step_confirmed"), ConfirmationStateFirstStep)
	assert.Equal(t, ConfirmationState("confirmed"), ConfirmationStateConfirmed)
	assert.Equal(t, ConfirmationState("rejected"), ConfirmationStateRejected)
	assert.Equal(t, ConfirmationState("expired"), ConfirmationStateExpired)
	assert.Equal(t, ConfirmationState("cancelled"), ConfirmationStateCancelled)
}

func TestErrors(t *testing.T) {
	assert.True(t, errors.Is(ErrOperationNotFound, ErrOperationNotFound))
	assert.True(t, errors.Is(ErrOperationAlreadyExists, ErrOperationAlreadyExists))
	assert.True(t, errors.Is(ErrOperationTimeout, ErrOperationTimeout))
	assert.True(t, errors.Is(ErrOperationCancelled, ErrOperationCancelled))
	assert.True(t, errors.Is(ErrQueueFull, ErrQueueFull))
	assert.True(t, errors.Is(ErrExecutorShutdown, ErrExecutorShutdown))
	assert.True(t, errors.Is(ErrConfirmationNotFound, ErrConfirmationNotFound))
	assert.True(t, errors.Is(ErrConfirmationExpired, ErrConfirmationExpired))
	assert.True(t, errors.Is(ErrConfirmationAlreadyUsed, ErrConfirmationAlreadyUsed))
	assert.True(t, errors.Is(ErrInvalidConfirmCode, ErrInvalidConfirmCode))
	assert.True(t, errors.Is(ErrUnauthorized, ErrUnauthorized))
	assert.True(t, errors.Is(ErrInsufficientAuthLevel, ErrInsufficientAuthLevel))
}

type MockRemoteControlHandler struct{}

func (h *MockRemoteControlHandler) Handle(ctx context.Context, op *ParsedOperation) (interface{}, error) {
	return map[string]interface{}{"device_id": op.TargetID, "action": op.Action, "success": true}, nil
}

func (h *MockRemoteControlHandler) CanHandle(op *ParsedOperation) bool {
	return op.Type == OperationTypeRemoteControl
}

func (h *MockRemoteControlHandler) Rollback(ctx context.Context, op *ParsedOperation, result interface{}) error {
	return nil
}
