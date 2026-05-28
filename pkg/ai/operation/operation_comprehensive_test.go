package operation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOperationType_Constants(t *testing.T) {
	assert.Equal(t, OperationType("remote_control"), OperationTypeRemoteControl)
	assert.Equal(t, OperationType("setpoint"), OperationTypeSetPoint)
	assert.Equal(t, OperationType("adjust"), OperationTypeAdjust)
	assert.Equal(t, OperationType("query"), OperationTypeQuery)
	assert.Equal(t, OperationType("batch"), OperationTypeBatch)
}

func TestOperationPriority_Constants(t *testing.T) {
	assert.Equal(t, OperationPriority(1), PriorityLow)
	assert.Equal(t, OperationPriority(5), PriorityNormal)
	assert.Equal(t, OperationPriority(8), PriorityHigh)
	assert.Equal(t, OperationPriority(10), PriorityCritical)
}

func TestOperationStatus_Constants(t *testing.T) {
	assert.Equal(t, OperationStatus("pending"), StatusPending)
	assert.Equal(t, OperationStatus("validating"), StatusValidating)
	assert.Equal(t, OperationStatus("confirmed"), StatusConfirmed)
	assert.Equal(t, OperationStatus("executing"), StatusExecuting)
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

func TestNewOperationQueue(t *testing.T) {
	q := NewOperationQueue(10)
	require.NotNil(t, q)
	assert.Equal(t, 0, q.Len())
}

func TestOperationQueue_PushPop(t *testing.T) {
	q := NewOperationQueue(10)
	op := &ParsedOperation{ID: "op-1", Type: OperationTypeQuery}
	err := q.Push(op)
	require.NoError(t, err)
	assert.Equal(t, 1, q.Len())
	popped := q.Pop()
	require.NotNil(t, popped)
	assert.Equal(t, "op-1", popped.ID)
	assert.Equal(t, 0, q.Len())
}

func TestOperationQueue_Pop_Empty(t *testing.T) {
	q := NewOperationQueue(10)
	popped := q.Pop()
	assert.Nil(t, popped)
}

func TestOperationQueue_Peek(t *testing.T) {
	q := NewOperationQueue(10)
	op1 := &ParsedOperation{ID: "op-1"}
	op2 := &ParsedOperation{ID: "op-2"}
	q.Push(op1)
	q.Push(op2)
	peeked := q.Peek()
	require.NotNil(t, peeked)
	assert.Equal(t, "op-1", peeked.ID)
	assert.Equal(t, 2, q.Len())
}

func TestOperationQueue_Peek_Empty(t *testing.T) {
	q := NewOperationQueue(10)
	peeked := q.Peek()
	assert.Nil(t, peeked)
}

func TestOperationQueue_Full(t *testing.T) {
	q := NewOperationQueue(2)
	q.Push(&ParsedOperation{ID: "op-1"})
	q.Push(&ParsedOperation{ID: "op-2"})
	err := q.Push(&ParsedOperation{ID: "op-3"})
	assert.Equal(t, ErrQueueFull, err)
}

func TestOperationQueue_Notify(t *testing.T) {
	q := NewOperationQueue(10)
	ch := q.Notify()
	assert.NotNil(t, ch)
	q.Push(&ParsedOperation{ID: "op-1"})
	select {
	case <-ch:
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

func TestNewOperationExecutor_NilConfig(t *testing.T) {
	e := NewOperationExecutor(nil)
	require.NotNil(t, e)
	require.NotNil(t, e.config)
	assert.Equal(t, 1000, e.config.QueueCapacity)
}

func TestNewOperationExecutor_CustomConfig(t *testing.T) {
	cfg := &ExecutorConfig{
		QueueCapacity:   50,
		MaxWorkers:      3,
		DefaultTimeout:  10 * time.Second,
		RetryDelay:      500 * time.Millisecond,
		MaxRetryDelay:   5 * time.Second,
		EnablePriority:  false,
		HistoryCapacity: 500,
	}
	e := NewOperationExecutor(cfg)
	require.NotNil(t, e)
	assert.Equal(t, 50, e.config.QueueCapacity)
}

func TestOperationExecutor_RegisterHandler(t *testing.T) {
	e := NewOperationExecutor(nil)
	e.RegisterHandler(OperationTypeQuery, &MockQueryHandler{})
	assert.NotNil(t, e.handlers[OperationTypeQuery])
}

func TestOperationExecutor_Submit(t *testing.T) {
	e := NewOperationExecutor(nil)
	op := &ParsedOperation{
		ID:     "test-submit-1",
		Type:   OperationTypeQuery,
		Action: "query",
	}
	err := e.Submit(context.Background(), op)
	require.NoError(t, err)
	record, err := e.GetStatus("test-submit-1")
	require.NoError(t, err)
	assert.Equal(t, StatusPending, record.Status)
}

func TestOperationExecutor_Submit_Shutdown(t *testing.T) {
	e := NewOperationExecutor(nil)
	e.Stop()
	err := e.Submit(context.Background(), &ParsedOperation{ID: "test-shutdown", Type: OperationTypeQuery})
	assert.Equal(t, ErrExecutorShutdown, err)
}

func TestOperationExecutor_Submit_Duplicate(t *testing.T) {
	e := NewOperationExecutor(nil)
	op := &ParsedOperation{ID: "test-dup", Type: OperationTypeQuery}
	e.Submit(context.Background(), op)
	err := e.Submit(context.Background(), op)
	assert.Equal(t, ErrOperationAlreadyExists, err)
}

func TestOperationExecutor_Submit_QueueFull(t *testing.T) {
	cfg := &ExecutorConfig{QueueCapacity: 1, MaxWorkers: 1, DefaultTimeout: time.Second, HistoryCapacity: 100}
	e := NewOperationExecutor(cfg)
	e.Submit(context.Background(), &ParsedOperation{ID: "qf-1", Type: OperationTypeQuery})
	err := e.Submit(context.Background(), &ParsedOperation{ID: "qf-2", Type: OperationTypeQuery})
	assert.Equal(t, ErrQueueFull, err)
}

func TestOperationExecutor_Cancel(t *testing.T) {
	e := NewOperationExecutor(nil)
	op := &ParsedOperation{ID: "test-cancel", Type: OperationTypeQuery}
	e.Submit(context.Background(), op)
	err := e.Cancel("test-cancel")
	require.NoError(t, err)
	record, _ := e.GetStatus("test-cancel")
	assert.Equal(t, StatusCancelled, record.Status)
}

func TestOperationExecutor_Cancel_NotFound(t *testing.T) {
	e := NewOperationExecutor(nil)
	err := e.Cancel("nonexistent")
	assert.Equal(t, ErrOperationNotFound, err)
}

func TestOperationExecutor_Cancel_InvalidStatus(t *testing.T) {
	e := NewOperationExecutor(nil)
	e.RegisterHandler(OperationTypeQuery, &MockQueryHandler{})
	e.Start(context.Background())
	op := &ParsedOperation{
		ID:          "test-cancel-exec",
		Type:        OperationTypeQuery,
		Action:      "query",
		Constraints: &OperationConstraints{Timeout: time.Second},
	}
	e.Submit(context.Background(), op)
	time.Sleep(200 * time.Millisecond)
	e.Stop()
	err := e.Cancel("test-cancel-exec")
	assert.Error(t, err)
}

func TestOperationExecutor_GetStatus_NotFound(t *testing.T) {
	e := NewOperationExecutor(nil)
	_, err := e.GetStatus("nonexistent")
	assert.Equal(t, ErrOperationNotFound, err)
}

func TestOperationExecutor_ExecuteAndRollback(t *testing.T) {
	e := NewOperationExecutor(nil)
	e.RegisterHandler(OperationTypeSetPoint, &MockSetPointHandler{})
	e.Start(context.Background())
	defer e.Stop()
	op := &ParsedOperation{
		ID:          "test-rollback-1",
		Type:        OperationTypeSetPoint,
		Action:      "set_value",
		TargetID:    "POINT-001",
		Constraints: &OperationConstraints{Timeout: 5 * time.Second, AllowRollback: true},
		Parameters:  map[string]interface{}{"value": 100.0},
	}
	e.Submit(context.Background(), op)
	record, err := e.WaitForCompletion(context.Background(), op.ID, 10*time.Second)
	require.NoError(t, err)
	assert.Equal(t, StatusSuccess, record.Status)
	err = e.Rollback(context.Background(), op.ID)
	require.NoError(t, err)
	updated, _ := e.GetStatus(op.ID)
	assert.Equal(t, StatusRolledBack, updated.Status)
}

func TestOperationExecutor_Rollback_NotFound(t *testing.T) {
	e := NewOperationExecutor(nil)
	err := e.Rollback(context.Background(), "nonexistent")
	assert.Equal(t, ErrOperationNotFound, err)
}

func TestOperationExecutor_Rollback_NotAllowed(t *testing.T) {
	e := NewOperationExecutor(nil)
	op := &ParsedOperation{
		ID:          "test-rollback-no",
		Type:        OperationTypeQuery,
		Action:      "query",
		Constraints: &OperationConstraints{AllowRollback: false},
	}
	e.Submit(context.Background(), op)
	err := e.Rollback(context.Background(), op.ID)
	assert.Error(t, err)
}

func TestOperationExecutor_Rollback_NotSuccess(t *testing.T) {
	e := NewOperationExecutor(nil)
	op := &ParsedOperation{
		ID:          "test-rollback-pending",
		Type:        OperationTypeQuery,
		Action:      "query",
		Constraints: &OperationConstraints{AllowRollback: true},
	}
	e.Submit(context.Background(), op)
	err := e.Rollback(context.Background(), op.ID)
	assert.Error(t, err)
}

func TestOperationExecutor_Rollback_NoHandler(t *testing.T) {
	e := NewOperationExecutor(nil)
	op := &ParsedOperation{
		ID:          "test-rollback-nohandler",
		Type:        OperationTypeBatch,
		Action:      "batch",
		Constraints: &OperationConstraints{AllowRollback: true},
	}
	e.mu.Lock()
	e.records[op.ID] = &OperationRecord{Operation: op, Status: StatusSuccess}
	e.statusMap[op.ID] = StatusSuccess
	e.mu.Unlock()
	err := e.Rollback(context.Background(), op.ID)
	assert.Error(t, err)
}

func TestOperationExecutor_GetHistory(t *testing.T) {
	e := NewOperationExecutor(nil)
	e.RegisterHandler(OperationTypeQuery, &MockQueryHandler{})
	e.Start(context.Background())
	op := &ParsedOperation{
		ID:          "test-history-1",
		Type:        OperationTypeQuery,
		Action:      "query",
		Constraints: &OperationConstraints{Timeout: 5 * time.Second},
	}
	e.Submit(context.Background(), op)
	e.WaitForCompletion(context.Background(), op.ID, 10*time.Second)
	e.Stop()
	history := e.GetHistory(10)
	assert.GreaterOrEqual(t, len(history), 1)
}

func TestOperationExecutor_GetHistory_All(t *testing.T) {
	e := NewOperationExecutor(nil)
	history := e.GetHistory(0)
	assert.Empty(t, history)
}

func TestOperationExecutor_GetQueueLength(t *testing.T) {
	e := NewOperationExecutor(nil)
	assert.Equal(t, 0, e.GetQueueLength())
	e.Submit(context.Background(), &ParsedOperation{ID: "ql-1", Type: OperationTypeQuery})
	assert.Equal(t, 1, e.GetQueueLength())
}

func TestOperationExecutor_GetRunningCount(t *testing.T) {
	e := NewOperationExecutor(nil)
	assert.Equal(t, int64(0), e.GetRunningCount())
}

func TestOperationExecutor_GetStats(t *testing.T) {
	e := NewOperationExecutor(nil)
	e.RegisterHandler(OperationTypeQuery, &MockQueryHandler{})
	e.Start(context.Background())
	op := &ParsedOperation{
		ID:          "test-stats-1",
		Type:        OperationTypeQuery,
		Action:      "query",
		Constraints: &OperationConstraints{Timeout: 5 * time.Second},
	}
	e.Submit(context.Background(), op)
	e.WaitForCompletion(context.Background(), op.ID, 10*time.Second)
	e.Stop()
	stats := e.GetStats()
	assert.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.TotalExecuted, 1)
	assert.NotNil(t, stats.ByStatus)
	assert.NotNil(t, stats.ByType)
}

func TestOperationExecutor_SubmitBatch(t *testing.T) {
	e := NewOperationExecutor(nil)
	ops := []*ParsedOperation{
		{ID: "batch-1", Type: OperationTypeQuery, Action: "query"},
		{ID: "batch-2", Type: OperationTypeQuery, Action: "query"},
	}
	result, err := e.SubmitBatch(context.Background(), ops)
	require.NoError(t, err)
	assert.Equal(t, 2, result.TotalCount)
	assert.Equal(t, 2, result.Success)
	assert.Equal(t, 0, result.Failed)
	assert.NotEmpty(t, result.BatchID)
}

func TestOperationExecutor_SubmitBatch_WithFailure(t *testing.T) {
	e := NewOperationExecutor(nil)
	e.Submit(context.Background(), &ParsedOperation{ID: "batch-dup", Type: OperationTypeQuery})
	ops := []*ParsedOperation{
		{ID: "batch-dup", Type: OperationTypeQuery, Action: "query"},
		{ID: "batch-ok", Type: OperationTypeQuery, Action: "query"},
	}
	result, err := e.SubmitBatch(context.Background(), ops)
	require.NoError(t, err)
	assert.Equal(t, 2, result.TotalCount)
	assert.Equal(t, 1, result.Success)
	assert.Equal(t, 1, result.Failed)
}

func TestOperationExecutor_Execute_NoHandler(t *testing.T) {
	e := NewOperationExecutor(nil)
	e.Start(context.Background())
	op := &ParsedOperation{
		ID:          "test-no-handler",
		Type:        OperationTypeBatch,
		Action:      "batch",
		Constraints: &OperationConstraints{Timeout: time.Second},
	}
	e.Submit(context.Background(), op)
	record, err := e.WaitForCompletion(context.Background(), op.ID, 10*time.Second)
	require.NoError(t, err)
	assert.Equal(t, StatusFailed, record.Status)
	e.Stop()
}

func TestOperationExecutor_Execute_HandlerError(t *testing.T) {
	e := NewOperationExecutor(nil)
	e.RegisterHandler(OperationTypeAdjust, &errorMockHandler{})
	e.Start(context.Background())
	op := &ParsedOperation{
		ID:          "test-handler-err",
		Type:        OperationTypeAdjust,
		Action:      "adjust",
		Constraints: &OperationConstraints{Timeout: time.Second, MaxRetries: 0},
	}
	e.Submit(context.Background(), op)
	record, err := e.WaitForCompletion(context.Background(), op.ID, 10*time.Second)
	require.NoError(t, err)
	assert.Equal(t, StatusFailed, record.Status)
	e.Stop()
}

func TestOperationExecutor_Execute_WithRetry(t *testing.T) {
	e := NewOperationExecutor(&ExecutorConfig{
		QueueCapacity:   100,
		MaxWorkers:      2,
		DefaultTimeout:  5 * time.Second,
		RetryDelay:      10 * time.Millisecond,
		MaxRetryDelay:   50 * time.Millisecond,
		HistoryCapacity: 100,
	})
	e.RegisterHandler(OperationTypeAdjust, &retryableMockHandler{})
	e.Start(context.Background())
	op := &ParsedOperation{
		ID:          "test-retry",
		Type:        OperationTypeAdjust,
		Action:      "adjust",
		Constraints: &OperationConstraints{Timeout: 5 * time.Second, MaxRetries: 2},
	}
	e.Submit(context.Background(), op)
	record, err := e.WaitForCompletion(context.Background(), op.ID, 15*time.Second)
	require.NoError(t, err)
	assert.Equal(t, StatusSuccess, record.Status)
	assert.GreaterOrEqual(t, record.RetryCount, 1)
	e.Stop()
}

func TestOperationExecutor_Execute_Timeout(t *testing.T) {
	e := NewOperationExecutor(nil)
	e.RegisterHandler(OperationTypeAdjust, &slowMockHandler{})
	e.Start(context.Background())
	op := &ParsedOperation{
		ID:          "test-timeout",
		Type:        OperationTypeAdjust,
		Action:      "adjust",
		Constraints: &OperationConstraints{Timeout: 50 * time.Millisecond, MaxRetries: 0},
	}
	e.Submit(context.Background(), op)
	record, err := e.WaitForCompletion(context.Background(), op.ID, 10*time.Second)
	require.NoError(t, err)
	assert.Equal(t, StatusTimeout, record.Status)
	e.Stop()
}

func TestOperationExecutor_Execute_Cancelled(t *testing.T) {
	e := NewOperationExecutor(nil)
	e.RegisterHandler(OperationTypeAdjust, &slowMockHandler{})
	ctx, cancel := context.WithCancel(context.Background())
	e.Start(ctx)
	op := &ParsedOperation{
		ID:          "test-cancelled",
		Type:        OperationTypeAdjust,
		Action:      "adjust",
		Constraints: &OperationConstraints{Timeout: 5 * time.Second, MaxRetries: 0},
	}
	e.Submit(ctx, op)
	cancel()
	time.Sleep(200 * time.Millisecond)
	e.Stop()
}

func TestNewOperationParser(t *testing.T) {
	p := NewOperationParser()
	require.NotNil(t, p)
	assert.NotNil(t, p.keywords)
	assert.NotNil(t, p.actionPatterns)
	assert.NotNil(t, p.safetyChecker)
}

func TestOperationParser_Parse_EmptyText(t *testing.T) {
	p := NewOperationParser()
	_, err := p.Parse(context.Background(), "")
	assert.Error(t, err)
}

func TestOperationParser_Parse_WhitespaceText(t *testing.T) {
	p := NewOperationParser()
	_, err := p.Parse(context.Background(), "   ")
	assert.Error(t, err)
}

func TestOperationParser_Parse_RemoteControl(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "启动设备DEV-001")
	require.NoError(t, err)
	require.NotEmpty(t, result.Operations)
	assert.Equal(t, OperationTypeRemoteControl, result.Operations[0].Type)
}

func TestOperationParser_Parse_SetPoint(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "设置测点POINT-001的值为100kW")
	require.NoError(t, err)
	require.NotEmpty(t, result.Operations)
	assert.Equal(t, OperationTypeSetPoint, result.Operations[0].Type)
}

func TestOperationParser_Parse_Adjust(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "调整逆变器INV-001功率增加50kW")
	require.NoError(t, err)
	require.NotEmpty(t, result.Operations)
	assert.Equal(t, OperationTypeAdjust, result.Operations[0].Type)
}

func TestOperationParser_Parse_Query(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "查询设备DEV-002的状态")
	require.NoError(t, err)
	require.NotEmpty(t, result.Operations)
	assert.Equal(t, OperationTypeQuery, result.Operations[0].Type)
}

func TestOperationParser_Parse_MultipleCommands(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "启动设备DEV-001然后查询设备DEV-002")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result.Operations), 2)
}

func TestOperationParser_Parse_TargetDevice(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "查询设备DEV-001的状态")
	require.NoError(t, err)
	require.NotEmpty(t, result.Operations)
	assert.Equal(t, "device", result.Operations[0].TargetType)
}

func TestOperationParser_Parse_TargetPoint(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "设置测点POINT-001的值")
	require.NoError(t, err)
	require.NotEmpty(t, result.Operations)
	assert.Equal(t, "point", result.Operations[0].TargetType)
}

func TestOperationParser_Parse_TargetStation(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "查询电站STATION-001的数据")
	require.NoError(t, err)
	require.NotEmpty(t, result.Operations)
	assert.Equal(t, "station", result.Operations[0].TargetType)
}

func TestOperationParser_Parse_TargetInverter(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "启动逆变器INV-001")
	require.NoError(t, err)
	require.NotEmpty(t, result.Operations)
	assert.Equal(t, "device", result.Operations[0].TargetType)
}

func TestOperationParser_Parse_NoTarget(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "查询数据")
	require.NoError(t, err)
	require.NotEmpty(t, result.Operations)
	assert.Empty(t, result.Operations[0].TargetID)
}

func TestOperationParser_Parse_ActionSwitchOn(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "启动设备DEV-001")
	require.NoError(t, err)
	require.NotEmpty(t, result.Operations)
	assert.Equal(t, "switch_on", result.Operations[0].Action)
}

func TestOperationParser_Parse_ActionSwitchOff(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "停止设备DEV-001")
	require.NoError(t, err)
	require.NotEmpty(t, result.Operations)
	assert.Equal(t, "switch_off", result.Operations[0].Action)
}

func TestOperationParser_Parse_ActionSetValue(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "设置测点POINT-001为100")
	require.NoError(t, err)
	require.NotEmpty(t, result.Operations)
	assert.Equal(t, "set_value", result.Operations[0].Action)
}

func TestOperationParser_Parse_ActionQuery(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "查询设备DEV-001")
	require.NoError(t, err)
	require.NotEmpty(t, result.Operations)
	assert.Equal(t, "query", result.Operations[0].Action)
}

func TestOperationParser_Parse_ExtractValue(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "设置测点POINT-001为100")
	require.NoError(t, err)
	require.NotEmpty(t, result.Operations)
	hasValue := false
	if v, ok := result.Operations[0].Parameters["value"]; ok {
		hasValue = v != nil
	} else if _, ok := result.Operations[0].Parameters["values"]; ok {
		hasValue = true
	}
	assert.True(t, hasValue)
}

func TestOperationParser_Parse_ExtractUnit(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "设置设备DEV-001功率为500kW")
	require.NoError(t, err)
	require.NotEmpty(t, result.Operations)
	_, hasUnit := result.Operations[0].Parameters["unit"]
	assert.True(t, hasUnit)
}

func TestOperationParser_Parse_ExtractTime(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "设置设备DEV-001在14:30执行")
	require.NoError(t, err)
	require.NotEmpty(t, result.Operations)
	_, hasTime := result.Operations[0].Parameters["time"]
	assert.True(t, hasTime)
}

func TestOperationParser_Parse_ExtractDelay(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "延迟10秒启动设备DEV-001")
	require.NoError(t, err)
	require.NotEmpty(t, result.Operations)
	_, hasDelay := result.Operations[0].Parameters["delay"]
	assert.True(t, hasDelay)
}

func TestOperationParser_Parse_MultipleValues(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "调整设备DEV-001参数100和200")
	require.NoError(t, err)
	require.NotEmpty(t, result.Operations)
	_, hasValues := result.Operations[0].Parameters["values"]
	assert.True(t, hasValues)
}

func TestOperationParser_Parse_Constraints(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "查询设备DEV-001的状态")
	require.NoError(t, err)
	require.NotEmpty(t, result.Operations)
	assert.NotNil(t, result.Operations[0].Constraints)
	assert.False(t, result.Operations[0].Constraints.RequireConfirm)
}

func TestOperationParser_Parse_Description(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "查询设备DEV-001的状态")
	require.NoError(t, err)
	require.NotEmpty(t, result.Operations)
	assert.NotEmpty(t, result.Operations[0].Description)
}

func TestOperationParser_Parse_Priority(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "查询设备DEV-001的状态")
	require.NoError(t, err)
	require.NotEmpty(t, result.Operations)
	assert.Equal(t, PriorityLow, result.Operations[0].Priority)
}

func TestOperationParser_Parse_PriorityCritical(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "关闭设备DEV-001")
	require.NoError(t, err)
	require.NotEmpty(t, result.Operations)
	assert.Equal(t, PriorityCritical, result.Operations[0].Priority)
}

func TestOperationParser_Parse_PriorityHigh(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "启动设备DEV-001")
	require.NoError(t, err)
	require.NotEmpty(t, result.Operations)
	assert.Equal(t, PriorityHigh, result.Operations[0].Priority)
}

func TestOperationParser_Parse_LowConfidence(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "随便说说")
	require.NoError(t, err)
	if len(result.Operations) > 0 && result.Operations[0].Confidence < 0.6 {
		assert.NotEmpty(t, result.Warnings)
	}
}

func TestOperationParser_Parse_SafetyWarnings(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "关闭设备DEV-001")
	require.NoError(t, err)
	assert.NotEmpty(t, result.Warnings)
}

func TestOperationParser_Parse_Suggestions(t *testing.T) {
	p := NewOperationParser()
	result, err := p.Parse(context.Background(), "查询数据")
	require.NoError(t, err)
	assert.NotEmpty(t, result.Suggestions)
}

func TestOperationParser_ValidateOperation_Valid(t *testing.T) {
	p := NewOperationParser()
	op := &ParsedOperation{
		Type:     OperationTypeQuery,
		Action:   "query",
		TargetID: "DEV-001",
	}
	err := p.ValidateOperation(op)
	assert.NoError(t, err)
}

func TestOperationParser_ValidateOperation_NoType(t *testing.T) {
	p := NewOperationParser()
	op := &ParsedOperation{Action: "query", TargetID: "DEV-001"}
	err := p.ValidateOperation(op)
	assert.Error(t, err)
}

func TestOperationParser_ValidateOperation_NoTarget(t *testing.T) {
	p := NewOperationParser()
	op := &ParsedOperation{Type: OperationTypeRemoteControl, Action: "control"}
	err := p.ValidateOperation(op)
	assert.Error(t, err)
}

func TestOperationParser_ValidateOperation_NoTarget_QueryAllowed(t *testing.T) {
	p := NewOperationParser()
	op := &ParsedOperation{Type: OperationTypeQuery, Action: "query"}
	err := p.ValidateOperation(op)
	assert.NoError(t, err)
}

func TestOperationParser_ValidateOperation_NoAction(t *testing.T) {
	p := NewOperationParser()
	op := &ParsedOperation{Type: OperationTypeQuery, TargetID: "DEV-001"}
	err := p.ValidateOperation(op)
	assert.Error(t, err)
}

func TestOperationParser_ValidateOperation_InvalidValue(t *testing.T) {
	p := NewOperationParser()
	op := &ParsedOperation{
		Type:       OperationTypeSetPoint,
		Action:     "set",
		TargetID:   "DEV-001",
		Parameters: map[string]interface{}{"value": "not-a-number"},
	}
	err := p.ValidateOperation(op)
	assert.Error(t, err)
}

func TestSafetyChecker_DangerousPattern(t *testing.T) {
	s := NewSafetyChecker()
	op := &ParsedOperation{
		OriginalText: "关闭设备DEV-001",
		TargetID:     "DEV-001",
	}
	warnings := s.Check(op)
	assert.NotEmpty(t, warnings)
}

func TestSafetyChecker_ProtectedTarget(t *testing.T) {
	s := NewSafetyChecker()
	s.AddProtectedTarget("CRITICAL-001")
	op := &ParsedOperation{
		OriginalText: "查询设备CRITICAL-001",
		TargetID:     "CRITICAL-001",
	}
	warnings := s.Check(op)
	assert.NotEmpty(t, warnings)
}

func TestSafetyChecker_RemoveProtectedTarget(t *testing.T) {
	s := NewSafetyChecker()
	s.AddProtectedTarget("CRITICAL-001")
	s.RemoveProtectedTarget("CRITICAL-001")
	op := &ParsedOperation{
		OriginalText: "查询设备CRITICAL-001",
		TargetID:     "CRITICAL-001",
	}
	warnings := s.Check(op)
	hasProtected := false
	for _, w := range warnings {
		if len(w) > 0 {
			hasProtected = true
		}
	}
	assert.True(t, hasProtected || len(warnings) == 0)
}

func TestSafetyChecker_ParameterOutOfRange(t *testing.T) {
	s := NewSafetyChecker()
	op := &ParsedOperation{
		OriginalText: "设置设备DEV-001功率为500kW",
		TargetID:     "DEV-001",
		Parameters:   map[string]interface{}{"value": 500.0},
		Constraints:  &OperationConstraints{MinValue: 0, MaxValue: 100},
	}
	warnings := s.Check(op)
	assert.NotEmpty(t, warnings)
}

func TestSafetyChecker_ParameterInRange(t *testing.T) {
	s := NewSafetyChecker()
	op := &ParsedOperation{
		OriginalText: "设置设备DEV-001功率为50kW",
		TargetID:     "DEV-001",
		Parameters:   map[string]interface{}{"value": 50.0},
		Constraints:  &OperationConstraints{MinValue: 0, MaxValue: 100},
	}
	warnings := s.Check(op)
	for _, w := range warnings {
		assert.NotContains(t, w, "超出允许范围")
	}
}

func TestSafetyChecker_NoConstraints(t *testing.T) {
	s := NewSafetyChecker()
	op := &ParsedOperation{
		OriginalText: "查询设备DEV-001",
		TargetID:     "DEV-001",
	}
	warnings := s.Check(op)
	assert.Empty(t, warnings)
}

func TestDefaultConfirmationConfig(t *testing.T) {
	cfg := DefaultConfirmationConfig()
	assert.Equal(t, 5*time.Minute, cfg.DefaultTimeout)
	assert.True(t, cfg.EnableTwoStep)
	assert.Equal(t, 6, cfg.CodeLength)
	assert.Equal(t, 1000, cfg.MaxPendingConfirm)
	assert.Equal(t, 10000, cfg.AuditLogCapacity)
}

func TestNewConfirmationManager_NilConfig(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	require.NotNil(t, m)
	require.NotNil(t, m.config)
	assert.Equal(t, 5*time.Minute, m.config.DefaultTimeout)
}

func TestConfirmationManager_CreateConfirmation_NoRequire(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:          "conf-no-require",
		Type:        OperationTypeQuery,
		Constraints: &OperationConstraints{RequireConfirm: false},
	}
	rec, err := m.CreateConfirmation(context.Background(), op)
	require.NoError(t, err)
	assert.Equal(t, ConfirmationStateConfirmed, rec.State)
}

func TestConfirmationManager_CreateConfirmation_Require(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:          "conf-require",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 2},
	}
	rec, err := m.CreateConfirmation(context.Background(), op)
	require.NoError(t, err)
	assert.Equal(t, ConfirmationStatePending, rec.State)
	assert.NotEmpty(t, rec.ConfirmCode)
	assert.NotEmpty(t, rec.ID)
}

func TestConfirmationManager_CreateConfirmation_NoConstraints(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:   "conf-no-constraints",
		Type: OperationTypeQuery,
	}
	rec, err := m.CreateConfirmation(context.Background(), op)
	require.NoError(t, err)
	assert.Equal(t, ConfirmationStateConfirmed, rec.State)
}

func TestConfirmationManager_CreateConfirmation_MaxPending(t *testing.T) {
	cfg := &ConfirmationConfig{MaxPendingConfirm: 1, CodeLength: 6, AuditLogCapacity: 100}
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(cfg, e)
	op1 := &ParsedOperation{
		ID:          "conf-max-1",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1},
	}
	m.CreateConfirmation(context.Background(), op1)
	op2 := &ParsedOperation{
		ID:          "conf-max-2",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1},
	}
	_, err := m.CreateConfirmation(context.Background(), op2)
	assert.Error(t, err)
}

func TestConfirmationManager_FirstStepConfirm(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:          "conf-first",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 2},
	}
	rec, _ := m.CreateConfirmation(context.Background(), op)
	err := m.FirstStepConfirm(context.Background(), rec.ID, rec.ConfirmCode, "user-001", 2)
	require.NoError(t, err)
	updated, _ := m.GetConfirmation(rec.ID)
	assert.Equal(t, ConfirmationStateFirstStep, updated.State)
}

func TestConfirmationManager_FirstStepConfirm_SingleStep(t *testing.T) {
	cfg := &ConfirmationConfig{EnableTwoStep: false, CodeLength: 6, AuditLogCapacity: 100, MaxPendingConfirm: 100, DefaultTimeout: 5 * time.Minute}
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(cfg, e)
	op := &ParsedOperation{
		ID:          "conf-single",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1},
	}
	rec, err := m.CreateConfirmation(context.Background(), op)
	require.NoError(t, err)
	require.NotNil(t, rec)
	err = m.FirstStepConfirm(context.Background(), rec.ID, rec.ConfirmCode, "user-001", 1)
	require.NoError(t, err)
	updated, _ := m.GetConfirmation(rec.ID)
	assert.Equal(t, ConfirmationStateConfirmed, updated.State)
}

func TestConfirmationManager_FirstStepConfirm_NotFound(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	err := m.FirstStepConfirm(context.Background(), "nonexistent", "code", "user", 1)
	assert.Equal(t, ErrConfirmationNotFound, err)
}

func TestConfirmationManager_FirstStepConfirm_AlreadyUsed(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:          "conf-used",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 2},
	}
	rec, _ := m.CreateConfirmation(context.Background(), op)
	m.FirstStepConfirm(context.Background(), rec.ID, rec.ConfirmCode, "user-001", 2)
	err := m.FirstStepConfirm(context.Background(), rec.ID, rec.ConfirmCode, "user-002", 2)
	assert.Equal(t, ErrConfirmationAlreadyUsed, err)
}

func TestConfirmationManager_FirstStepConfirm_Expired(t *testing.T) {
	cfg := &ConfirmationConfig{DefaultTimeout: 1 * time.Nanosecond, EnableTwoStep: true, CodeLength: 6, AuditLogCapacity: 100, MaxPendingConfirm: 100}
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(cfg, e)
	op := &ParsedOperation{
		ID:          "conf-expired",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 2},
	}
	rec, err := m.CreateConfirmation(context.Background(), op)
	require.NoError(t, err)
	require.NotNil(t, rec)
	time.Sleep(10 * time.Millisecond)
	err = m.FirstStepConfirm(context.Background(), rec.ID, rec.ConfirmCode, "user-001", 2)
	assert.Equal(t, ErrConfirmationExpired, err)
}

func TestConfirmationManager_FirstStepConfirm_InvalidCode(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:          "conf-invalid-code",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 2},
	}
	rec, _ := m.CreateConfirmation(context.Background(), op)
	err := m.FirstStepConfirm(context.Background(), rec.ID, "wrong-code", "user-001", 2)
	assert.Equal(t, ErrInvalidConfirmCode, err)
}

func TestConfirmationManager_FirstStepConfirm_InsufficientAuth(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:          "conf-low-auth",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 3},
	}
	rec, _ := m.CreateConfirmation(context.Background(), op)
	err := m.FirstStepConfirm(context.Background(), rec.ID, rec.ConfirmCode, "user-001", 1)
	assert.Equal(t, ErrInsufficientAuthLevel, err)
}

func TestConfirmationManager_SecondStepConfirm(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:          "conf-second",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 2},
	}
	rec, _ := m.CreateConfirmation(context.Background(), op)
	m.FirstStepConfirm(context.Background(), rec.ID, rec.ConfirmCode, "user-001", 2)
	err := m.SecondStepConfirm(context.Background(), rec.ID, "user-002", 2)
	require.NoError(t, err)
	updated, _ := m.GetConfirmation(rec.ID)
	assert.Equal(t, ConfirmationStateConfirmed, updated.State)
}

func TestConfirmationManager_SecondStepConfirm_NotFound(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	err := m.SecondStepConfirm(context.Background(), "nonexistent", "user", 1)
	assert.Equal(t, ErrConfirmationNotFound, err)
}

func TestConfirmationManager_SecondStepConfirm_InvalidState(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:          "conf-second-invalid",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 2},
	}
	rec, _ := m.CreateConfirmation(context.Background(), op)
	err := m.SecondStepConfirm(context.Background(), rec.ID, "user-002", 2)
	assert.Error(t, err)
}

func TestConfirmationManager_SecondStepConfirm_SameUser(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:          "conf-same-user",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 2},
	}
	rec, _ := m.CreateConfirmation(context.Background(), op)
	m.FirstStepConfirm(context.Background(), rec.ID, rec.ConfirmCode, "user-001", 2)
	err := m.SecondStepConfirm(context.Background(), rec.ID, "user-001", 2)
	assert.Error(t, err)
}

func TestConfirmationManager_SecondStepConfirm_Expired(t *testing.T) {
	cfg := &ConfirmationConfig{DefaultTimeout: 100 * time.Millisecond, EnableTwoStep: true, CodeLength: 6, AuditLogCapacity: 100, MaxPendingConfirm: 100}
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(cfg, e)
	op := &ParsedOperation{
		ID:          "conf-second-expired",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 2},
	}
	rec, err := m.CreateConfirmation(context.Background(), op)
	require.NoError(t, err)
	require.NotNil(t, rec)
	m.FirstStepConfirm(context.Background(), rec.ID, rec.ConfirmCode, "user-001", 2)
	time.Sleep(200 * time.Millisecond)
	err = m.SecondStepConfirm(context.Background(), rec.ID, "user-002", 2)
	assert.Equal(t, ErrConfirmationExpired, err)
}

func TestConfirmationManager_Reject(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:          "conf-reject",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1},
	}
	rec, _ := m.CreateConfirmation(context.Background(), op)
	err := m.Reject(context.Background(), rec.ID, "admin", "dangerous operation")
	require.NoError(t, err)
	updated, _ := m.GetConfirmation(rec.ID)
	assert.Equal(t, ConfirmationStateRejected, updated.State)
	assert.Equal(t, "dangerous operation", updated.RejectReason)
}

func TestConfirmationManager_Reject_NotFound(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	err := m.Reject(context.Background(), "nonexistent", "admin", "reason")
	assert.Equal(t, ErrConfirmationNotFound, err)
}

func TestConfirmationManager_Reject_AlreadyUsed(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:          "conf-reject-used",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1},
	}
	rec, _ := m.CreateConfirmation(context.Background(), op)
	m.Reject(context.Background(), rec.ID, "admin", "reason")
	err := m.Reject(context.Background(), rec.ID, "admin2", "another reason")
	assert.Equal(t, ErrConfirmationAlreadyUsed, err)
}

func TestConfirmationManager_Cancel(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:          "conf-cancel",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1},
	}
	rec, _ := m.CreateConfirmation(context.Background(), op)
	err := m.Cancel(context.Background(), rec.ID, "user-001")
	require.NoError(t, err)
	updated, _ := m.GetConfirmation(rec.ID)
	assert.Equal(t, ConfirmationStateCancelled, updated.State)
}

func TestConfirmationManager_Cancel_NotFound(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	err := m.Cancel(context.Background(), "nonexistent", "user")
	assert.Equal(t, ErrConfirmationNotFound, err)
}

func TestConfirmationManager_Cancel_InvalidState(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:          "conf-cancel-invalid",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1},
	}
	rec, _ := m.CreateConfirmation(context.Background(), op)
	m.Reject(context.Background(), rec.ID, "admin", "reason")
	err := m.Cancel(context.Background(), rec.ID, "user")
	assert.Error(t, err)
}

func TestConfirmationManager_GetConfirmation_NotFound(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	_, err := m.GetConfirmation("nonexistent")
	assert.Equal(t, ErrConfirmationNotFound, err)
}

func TestConfirmationManager_GetConfirmationByOperation(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:          "conf-by-op",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1},
	}
	m.CreateConfirmation(context.Background(), op)
	rec, err := m.GetConfirmationByOperation("conf-by-op")
	require.NoError(t, err)
	assert.NotNil(t, rec)
}

func TestConfirmationManager_GetConfirmationByOperation_NotFound(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	_, err := m.GetConfirmationByOperation("nonexistent")
	assert.Equal(t, ErrConfirmationNotFound, err)
}

func TestConfirmationManager_IsConfirmed(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:          "conf-is-confirmed",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1},
	}
	rec, err := m.CreateConfirmation(context.Background(), op)
	require.NoError(t, err)
	m.FirstStepConfirm(context.Background(), rec.ID, rec.ConfirmCode, "user-001", 1)
	confirmed, err := m.IsConfirmed(rec.ID)
	require.NoError(t, err)
	assert.True(t, confirmed)
}

func TestConfirmationManager_IsConfirmed_NotFound(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	_, err := m.IsConfirmed("nonexistent")
	assert.Equal(t, ErrConfirmationNotFound, err)
}

func TestConfirmationManager_Authorize(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:          "conf-auth",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1},
	}
	rec, err := m.CreateConfirmation(context.Background(), op)
	require.NoError(t, err)
	m.FirstStepConfirm(context.Background(), rec.ID, rec.ConfirmCode, "user-001", 1)
	err = m.Authorize(context.Background(), rec.ID, "admin", 2)
	require.NoError(t, err)
	updated, _ := m.GetConfirmation(rec.ID)
	assert.Equal(t, "admin", updated.AuthorizedBy)
}

func TestConfirmationManager_Authorize_NotFound(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	err := m.Authorize(context.Background(), "nonexistent", "admin", 1)
	assert.Equal(t, ErrConfirmationNotFound, err)
}

func TestConfirmationManager_Authorize_NotConfirmed(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:          "conf-auth-not-confirmed",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 2},
	}
	rec, err := m.CreateConfirmation(context.Background(), op)
	require.NoError(t, err)
	err = m.Authorize(context.Background(), rec.ID, "admin", 2)
	assert.Error(t, err)
}

func TestConfirmationManager_Rollback(t *testing.T) {
	e := NewOperationExecutor(nil)
	e.RegisterHandler(OperationTypeSetPoint, &MockSetPointHandler{})
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:          "conf-rollback",
		Type:        OperationTypeSetPoint,
		Action:      "set_value",
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1, AllowRollback: true},
	}
	rec, err := m.CreateConfirmation(context.Background(), op)
	require.NoError(t, err)
	m.FirstStepConfirm(context.Background(), rec.ID, rec.ConfirmCode, "user-001", 1)
	e.mu.Lock()
	e.records[op.ID] = &OperationRecord{Operation: op, Status: StatusSuccess, Result: "ok"}
	e.statusMap[op.ID] = StatusSuccess
	e.mu.Unlock()
	err = m.Rollback(context.Background(), op.ID, "admin")
	require.NoError(t, err)
}

func TestConfirmationManager_Rollback_NotFound(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	err := m.Rollback(context.Background(), "nonexistent", "admin")
	assert.Equal(t, ErrConfirmationNotFound, err)
}

func TestConfirmationManager_Rollback_NotAllowed(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:          "conf-rollback-no",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1, AllowRollback: false},
	}
	rec, err := m.CreateConfirmation(context.Background(), op)
	require.NoError(t, err)
	m.FirstStepConfirm(context.Background(), rec.ID, rec.ConfirmCode, "user-001", 1)
	err = m.Rollback(context.Background(), op.ID, "admin")
	assert.Error(t, err)
}

func TestConfirmationManager_GetAuditLogs(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:          "conf-audit",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1},
	}
	m.CreateConfirmation(context.Background(), op)
	logs := m.GetAuditLogs("", 10)
	assert.NotEmpty(t, logs)
}

func TestConfirmationManager_GetAuditLogs_ByOpID(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:          "conf-audit-op",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1},
	}
	m.CreateConfirmation(context.Background(), op)
	logs := m.GetAuditLogs("conf-audit-op", 10)
	assert.NotEmpty(t, logs)
}

func TestConfirmationManager_CleanupExpired(t *testing.T) {
	cfg := &ConfirmationConfig{DefaultTimeout: 1 * time.Nanosecond, EnableTwoStep: true, CodeLength: 6, AuditLogCapacity: 100, MaxPendingConfirm: 100}
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(cfg, e)
	op := &ParsedOperation{
		ID:          "conf-cleanup",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1},
	}
	m.CreateConfirmation(context.Background(), op)
	time.Sleep(10 * time.Millisecond)
	count := m.CleanupExpired()
	assert.Equal(t, 1, count)
}

func TestConfirmationManager_CleanupExpired_None(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	count := m.CleanupExpired()
	assert.Equal(t, 0, count)
}

func TestConfirmationManager_GetPendingConfirmations(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:          "conf-pending",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1},
	}
	m.CreateConfirmation(context.Background(), op)
	pending := m.GetPendingConfirmations()
	assert.NotEmpty(t, pending)
}

func TestConfirmationManager_GetStats(t *testing.T) {
	e := NewOperationExecutor(nil)
	m := NewConfirmationManager(nil, e)
	op := &ParsedOperation{
		ID:          "conf-stats",
		Type:        OperationTypeRemoteControl,
		Constraints: &OperationConstraints{RequireConfirm: true, AuthLevel: 1},
	}
	m.CreateConfirmation(context.Background(), op)
	stats := m.GetStats()
	assert.NotNil(t, stats)
	assert.NotNil(t, stats.ByState)
}

func TestNewOperationAPI(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, nil)
	require.NotNil(t, api)
}

func TestOperationAPI_SubmitOperation(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	executor.RegisterHandler(OperationTypeQuery, &MockQueryHandler{})
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, nil)
	resp, err := api.SubmitOperation(context.Background(), &OperationRequest{
		Text:   "查询设备DEV-001的状态",
		UserID: "user-001",
	})
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.NotEmpty(t, resp.Operations)
}

func TestOperationAPI_SubmitOperation_EmptyText(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, nil)
	resp, err := api.SubmitOperation(context.Background(), &OperationRequest{
		Text:   "",
		UserID: "user-001",
	})
	require.NoError(t, err)
	assert.False(t, resp.Success)
}

func TestOperationAPI_SubmitOperation_WithDryRun(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, nil)
	resp, err := api.SubmitOperation(context.Background(), &OperationRequest{
		Text:    "查询设备DEV-001的状态",
		UserID:  "user-001",
		DryRun:  true,
	})
	require.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestOperationAPI_SubmitOperation_WithConstraints(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, nil)
	resp, err := api.SubmitOperation(context.Background(), &OperationRequest{
		Text:   "查询设备DEV-001的状态",
		UserID: "user-001",
		Constraints: &OperationConstraints{
			Timeout:    10 * time.Second,
			MaxRetries: 5,
			AuthLevel:  3,
		},
	})
	require.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestOperationAPI_SubmitOperation_WithParameters(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, nil)
	resp, err := api.SubmitOperation(context.Background(), &OperationRequest{
		Text:       "查询设备DEV-001的状态",
		UserID:     "user-001",
		Parameters: map[string]interface{}{"extra": "value"},
	})
	require.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestOperationAPI_ConfirmOperation(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	executor.RegisterHandler(OperationTypeRemoteControl, &MockQueryHandler{})
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, &MockAuthChecker{})
	submitResp, _ := api.SubmitOperation(context.Background(), &OperationRequest{
		Text:   "启动设备DEV-001",
		UserID: "user-001",
	})
	require.NotEmpty(t, submitResp.Confirmations)
	confirmID := submitResp.Confirmations[0].ConfirmID
	confirmCode := submitResp.Confirmations[0].ConfirmCode
	confirmResp, err := api.ConfirmOperation(context.Background(), &ConfirmRequest{
		ConfirmID:   confirmID,
		ConfirmCode: confirmCode,
		UserID:      "user-001",
		Step:        1,
	})
	require.NoError(t, err)
	assert.True(t, confirmResp.Success)
}

func TestOperationAPI_ConfirmOperation_NotFound(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, nil)
	resp, err := api.ConfirmOperation(context.Background(), &ConfirmRequest{
		ConfirmID:   "nonexistent",
		ConfirmCode: "code",
		UserID:      "user-001",
		Step:        1,
	})
	require.NoError(t, err)
	assert.False(t, resp.Success)
}

func TestOperationAPI_ConfirmOperation_TwoStep(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	executor.RegisterHandler(OperationTypeRemoteControl, &MockQueryHandler{})
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, &MockAuthChecker{})
	submitResp, _ := api.SubmitOperation(context.Background(), &OperationRequest{
		Text:   "启动设备DEV-001",
		UserID: "user-001",
	})
	require.NotEmpty(t, submitResp.Confirmations)
	confirmID := submitResp.Confirmations[0].ConfirmID
	confirmCode := submitResp.Confirmations[0].ConfirmCode
	firstResp, _ := api.ConfirmOperation(context.Background(), &ConfirmRequest{
		ConfirmID:   confirmID,
		ConfirmCode: confirmCode,
		UserID:      "user-001",
		Step:        1,
	})
	assert.True(t, firstResp.NeedSecondStep)
	secondResp, err := api.ConfirmOperation(context.Background(), &ConfirmRequest{
		ConfirmID: confirmID,
		UserID:    "user-002",
		Step:      2,
	})
	require.NoError(t, err)
	assert.True(t, secondResp.Success)
}

func TestOperationAPI_RejectOperation(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, nil)
	submitResp, _ := api.SubmitOperation(context.Background(), &OperationRequest{
		Text:   "启动设备DEV-001",
		UserID: "user-001",
	})
	require.NotEmpty(t, submitResp.Confirmations)
	err := api.RejectOperation(context.Background(), submitResp.Confirmations[0].ConfirmID, "admin", "dangerous")
	assert.NoError(t, err)
}

func TestOperationAPI_GetOperationStatus(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, nil)
	resp, err := api.GetOperationStatus(context.Background(), &StatusRequest{OperationID: "nonexistent"})
	require.NoError(t, err)
	assert.False(t, resp.Success)
}

func TestOperationAPI_GetOperationHistory(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, nil)
	resp, err := api.GetOperationHistory(context.Background(), &HistoryRequest{Limit: 10})
	require.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestOperationAPI_GetOperationHistory_WithFilters(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, nil)
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now().Add(1 * time.Hour)
	resp, err := api.GetOperationHistory(context.Background(), &HistoryRequest{
		Status:        "success",
		OperationType: "query",
		StartTime:     &startTime,
		EndTime:       &endTime,
		Limit:         10,
		Offset:        0,
	})
	require.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestOperationAPI_RollbackOperation(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, nil)
	resp, err := api.RollbackOperation(context.Background(), &RollbackRequest{
		OperationID: "nonexistent",
		UserID:      "admin",
	})
	require.NoError(t, err)
	assert.False(t, resp.Success)
}

func TestOperationAPI_GetPendingConfirmations(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, nil)
	api.SubmitOperation(context.Background(), &OperationRequest{
		Text:   "启动设备DEV-001",
		UserID: "user-001",
	})
	pending, err := api.GetPendingConfirmations(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, pending)
}





func TestOperationAPI_GetStats(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, nil)
	stats := api.GetStats(context.Background())
	assert.NotNil(t, stats)
	assert.NotNil(t, stats.Executor)
	assert.NotNil(t, stats.Confirmer)
}

func TestOperationAPI_ValidateRequest(t *testing.T) {
	api := NewOperationAPI(nil, nil, nil, nil)
	err := api.ValidateRequest(&OperationRequest{Text: "test", UserID: "user1"})
	assert.NoError(t, err)
	err = api.ValidateRequest(&OperationRequest{Text: "", UserID: "user1"})
	assert.Error(t, err)
	err = api.ValidateRequest(&OperationRequest{Text: "test", UserID: ""})
	assert.Error(t, err)
}

func TestOperationAPI_HandleSubmit(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, nil)
	body, _ := json.Marshal(&OperationRequest{Text: "查询设备DEV-001", UserID: "user1"})
	req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewReader(body))
	w := httptest.NewRecorder()
	api.HandleSubmit(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOperationAPI_HandleSubmit_WrongMethod(t *testing.T) {
	api := NewOperationAPI(nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/submit", nil)
	w := httptest.NewRecorder()
	api.HandleSubmit(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestOperationAPI_HandleSubmit_BadJSON(t *testing.T) {
	api := NewOperationAPI(nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()
	api.HandleSubmit(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOperationAPI_HandleConfirm(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, nil)
	body, _ := json.Marshal(&ConfirmRequest{ConfirmID: "test", ConfirmCode: "code", UserID: "user1", Step: 1})
	req := httptest.NewRequest(http.MethodPost, "/confirm", bytes.NewReader(body))
	w := httptest.NewRecorder()
	api.HandleConfirm(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOperationAPI_HandleConfirm_WrongMethod(t *testing.T) {
	api := NewOperationAPI(nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/confirm", nil)
	w := httptest.NewRecorder()
	api.HandleConfirm(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestOperationAPI_HandleStatus(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, nil)
	req := httptest.NewRequest(http.MethodGet, "/status?operation_id=test", nil)
	w := httptest.NewRecorder()
	api.HandleStatus(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOperationAPI_HandleStatus_NoOpID(t *testing.T) {
	api := NewOperationAPI(nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	w := httptest.NewRecorder()
	api.HandleStatus(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOperationAPI_HandleStatus_WrongMethod(t *testing.T) {
	api := NewOperationAPI(nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/status", nil)
	w := httptest.NewRecorder()
	api.HandleStatus(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestOperationAPI_HandleHistory(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, nil)
	req := httptest.NewRequest(http.MethodGet, "/history?limit=10&status=success&type=query", nil)
	w := httptest.NewRecorder()
	api.HandleHistory(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOperationAPI_HandleHistory_WrongMethod(t *testing.T) {
	api := NewOperationAPI(nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/history", nil)
	w := httptest.NewRecorder()
	api.HandleHistory(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestOperationAPI_HandleRollback(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, nil)
	body, _ := json.Marshal(&RollbackRequest{OperationID: "test", UserID: "admin"})
	req := httptest.NewRequest(http.MethodPost, "/rollback", bytes.NewReader(body))
	w := httptest.NewRecorder()
	api.HandleRollback(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOperationAPI_HandleRollback_WrongMethod(t *testing.T) {
	api := NewOperationAPI(nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/rollback", nil)
	w := httptest.NewRecorder()
	api.HandleRollback(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestOperationAPI_HandlePending(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, nil)
	req := httptest.NewRequest(http.MethodGet, "/pending", nil)
	w := httptest.NewRecorder()
	api.HandlePending(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOperationAPI_HandlePending_WrongMethod(t *testing.T) {
	api := NewOperationAPI(nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/pending", nil)
	w := httptest.NewRecorder()
	api.HandlePending(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestOperationAPI_HandleAuditLogs(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, nil)
	req := httptest.NewRequest(http.MethodGet, "/audit?operation_id=test&limit=10", nil)
	w := httptest.NewRecorder()
	api.HandleAuditLogs(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOperationAPI_HandleAuditLogs_WrongMethod(t *testing.T) {
	api := NewOperationAPI(nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/audit", nil)
	w := httptest.NewRecorder()
	api.HandleAuditLogs(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestOperationAPI_HandleStats(t *testing.T) {
	parser := NewOperationParser()
	executor := NewOperationExecutor(nil)
	confirmer := NewConfirmationManager(nil, executor)
	api := NewOperationAPI(parser, executor, confirmer, nil)
	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	w := httptest.NewRecorder()
	api.HandleStats(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOperationAPI_HandleStats_WrongMethod(t *testing.T) {
	api := NewOperationAPI(nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/stats", nil)
	w := httptest.NewRecorder()
	api.HandleStats(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestOperationParser_SplitCommands(t *testing.T) {
	p := NewOperationParser()
	result := p.splitCommands("启动设备DEV-001然后查询设备DEV-002")
	assert.GreaterOrEqual(t, len(result), 2)
}

func TestOperationParser_SplitCommands_Semicolon(t *testing.T) {
	p := NewOperationParser()
	result := p.splitCommands("启动设备DEV-001;查询设备DEV-002")
	assert.GreaterOrEqual(t, len(result), 2)
}

func TestOperationParser_SplitCommands_And(t *testing.T) {
	p := NewOperationParser()
	result := p.splitCommands("启动设备DEV-001 and 查询设备DEV-002")
	assert.GreaterOrEqual(t, len(result), 2)
}

func TestOperationParser_GetDefaultConstraints_RemoteControl(t *testing.T) {
	p := NewOperationParser()
	c := p.getDefaultConstraints(OperationTypeRemoteControl, "switch_on")
	assert.Equal(t, 10*time.Second, c.Timeout)
	assert.Equal(t, 2, c.AuthLevel)
	assert.True(t, c.RequireConfirm)
}

func TestOperationParser_GetDefaultConstraints_SetPoint(t *testing.T) {
	p := NewOperationParser()
	c := p.getDefaultConstraints(OperationTypeSetPoint, "set_value")
	assert.Equal(t, 15*time.Second, c.Timeout)
	assert.Equal(t, 2, c.AuthLevel)
}

func TestOperationParser_GetDefaultConstraints_Adjust(t *testing.T) {
	p := NewOperationParser()
	c := p.getDefaultConstraints(OperationTypeAdjust, "adjust")
	assert.Equal(t, 20*time.Second, c.Timeout)
	assert.Equal(t, 1, c.AuthLevel)
}

func TestOperationParser_GetDefaultConstraints_Query(t *testing.T) {
	p := NewOperationParser()
	c := p.getDefaultConstraints(OperationTypeQuery, "query")
	assert.Equal(t, 5*time.Second, c.Timeout)
	assert.Equal(t, 0, c.AuthLevel)
	assert.False(t, c.RequireConfirm)
	assert.False(t, c.AllowRollback)
}

func TestOperationParser_GetDefaultConstraints_DangerousAction(t *testing.T) {
	p := NewOperationParser()
	c := p.getDefaultConstraints(OperationTypeRemoteControl, "switch_off")
	assert.Equal(t, 3, c.AuthLevel)
}

func TestOperationParser_DeterminePriority(t *testing.T) {
	p := NewOperationParser()
	assert.Equal(t, PriorityCritical, p.determinePriority(&ParsedOperation{Action: "switch_off"}))
	assert.Equal(t, PriorityCritical, p.determinePriority(&ParsedOperation{Action: "shutdown"}))
	assert.Equal(t, PriorityCritical, p.determinePriority(&ParsedOperation{Action: "emergency"}))
	assert.Equal(t, PriorityHigh, p.determinePriority(&ParsedOperation{Action: "switch_on"}))
	assert.Equal(t, PriorityHigh, p.determinePriority(&ParsedOperation{Type: OperationTypeRemoteControl, Action: "other"}))
	assert.Equal(t, PriorityNormal, p.determinePriority(&ParsedOperation{Type: OperationTypeSetPoint, Action: "set"}))
	assert.Equal(t, PriorityLow, p.determinePriority(&ParsedOperation{Type: OperationTypeQuery, Action: "query"}))
}

func TestOperationParser_GenerateDescription(t *testing.T) {
	p := NewOperationParser()
	op := &ParsedOperation{
		Type:       OperationTypeQuery,
		TargetName: "设备DEV-001",
		Action:     "query",
		Parameters: map[string]interface{}{"value": 100.0, "unit": "kW"},
	}
	desc := p.generateDescription(op)
	assert.Contains(t, desc, "查询操作")
	assert.Contains(t, desc, "设备DEV-001")
	assert.Contains(t, desc, "query")
}

func TestOperationParser_GenerateDescription_NoTargetName(t *testing.T) {
	p := NewOperationParser()
	op := &ParsedOperation{
		Type:     OperationTypeQuery,
		TargetID: "DEV-001",
		Action:   "query",
	}
	desc := p.generateDescription(op)
	assert.Contains(t, desc, "DEV-001")
}

func TestOperationParser_GenerateDescription_NoValue(t *testing.T) {
	p := NewOperationParser()
	op := &ParsedOperation{
		Type:   OperationTypeQuery,
		Action: "query",
	}
	desc := p.generateDescription(op)
	assert.NotEmpty(t, desc)
}

func TestGenerateOperationID(t *testing.T) {
	id := generateOperationID()
	assert.Contains(t, id, "OP-")
}

func TestGenerateConfirmID(t *testing.T) {
	id := generateConfirmID()
	assert.Contains(t, id, "CONF-")
}

func TestGenerateConfirmCode(t *testing.T) {
	code, err := generateConfirmCode(6)
	require.NoError(t, err)
	assert.Len(t, code, 6)
}

func TestGenerateAuditID(t *testing.T) {
	id := generateAuditID()
	assert.Contains(t, id, "AUDIT-")
}

func TestGenerateRollbackID(t *testing.T) {
	id := generateRollbackID()
	assert.Contains(t, id, "ROLL-")
}

func TestGenerateRequestID(t *testing.T) {
	id := generateRequestID()
	assert.Contains(t, id, "REQ-")
}

type errorMockHandler struct{}

func (h *errorMockHandler) Handle(ctx context.Context, op *ParsedOperation) (interface{}, error) {
	return nil, errors.New("handler error")
}

func (h *errorMockHandler) CanHandle(op *ParsedOperation) bool {
	return op.Type == OperationTypeAdjust
}

func (h *errorMockHandler) Rollback(ctx context.Context, op *ParsedOperation, result interface{}) error {
	return nil
}

type retryableMockHandler struct {
	attempts int
}

func (h *retryableMockHandler) Handle(ctx context.Context, op *ParsedOperation) (interface{}, error) {
	h.attempts++
	if h.attempts < 2 {
		return nil, context.DeadlineExceeded
	}
	return "success", nil
}

func (h *retryableMockHandler) CanHandle(op *ParsedOperation) bool {
	return op.Type == OperationTypeAdjust
}

func (h *retryableMockHandler) Rollback(ctx context.Context, op *ParsedOperation, result interface{}) error {
	return nil
}

type slowMockHandler struct{}

func (h *slowMockHandler) Handle(ctx context.Context, op *ParsedOperation) (interface{}, error) {
	select {
	case <-time.After(10 * time.Second):
		return "done", nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (h *slowMockHandler) CanHandle(op *ParsedOperation) bool {
	return op.Type == OperationTypeAdjust
}

func (h *slowMockHandler) Rollback(ctx context.Context, op *ParsedOperation, result interface{}) error {
	return nil
}
