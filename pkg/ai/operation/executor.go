package operation

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrOperationNotFound      = errors.New("operation not found")
	ErrOperationAlreadyExists = errors.New("operation already exists")
	ErrOperationTimeout       = errors.New("operation timeout")
	ErrOperationCancelled     = errors.New("operation cancelled")
	ErrQueueFull              = errors.New("operation queue is full")
	ErrExecutorShutdown       = errors.New("executor is shutdown")
)

// OperationRecord 操作记录
type OperationRecord struct {
	Operation   *ParsedOperation `json:"operation"`
	Status      OperationStatus  `json:"status"`
	Progress    int              `json:"progress"` // 0-100
	Result      interface{}      `json:"result,omitempty"`
	Error       string           `json:"error,omitempty"`
	StartTime   *time.Time       `json:"start_time,omitempty"`
	EndTime     *time.Time       `json:"end_time,omitempty"`
	Duration    time.Duration    `json:"duration"`
	RetryCount  int              `json:"retry_count"`
	ExecutedBy  string           `json:"executed_by"`
	ConfirmedBy string           `json:"confirmed_by,omitempty"`
}

// OperationQueue 操作队列
type OperationQueue struct {
	items    []*ParsedOperation
	capacity int
	mu       sync.RWMutex
	notify   chan struct{}
}

// NewOperationQueue 创建操作队列
func NewOperationQueue(capacity int) *OperationQueue {
	return &OperationQueue{
		items:    make([]*ParsedOperation, 0, capacity),
		capacity: capacity,
		notify:   make(chan struct{}, 1),
	}
}

// Push 添加操作到队列
func (q *OperationQueue) Push(op *ParsedOperation) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) >= q.capacity {
		return ErrQueueFull
	}

	q.items = append(q.items, op)

	// 通知有新操作
	select {
	case q.notify <- struct{}{}:
	default:
	}

	return nil
}

// Pop 从队列取出操作
func (q *OperationQueue) Pop() *ParsedOperation {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	op := q.items[0]
	q.items = q.items[1:]
	return op
}

// Peek 查看队列头部操作
func (q *OperationQueue) Peek() *ParsedOperation {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if len(q.items) == 0 {
		return nil
	}

	return q.items[0]
}

// Len 获取队列长度
func (q *OperationQueue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.items)
}

// Notify 返回通知通道
func (q *OperationQueue) Notify() <-chan struct{} {
	return q.notify
}

// OperationHandler 操作处理器接口
type OperationHandler interface {
	// Handle 处理操作
	Handle(ctx context.Context, op *ParsedOperation) (interface{}, error)
	// CanHandle 判断是否能处理该操作
	CanHandle(op *ParsedOperation) bool
	// Rollback 回滚操作
	Rollback(ctx context.Context, op *ParsedOperation, result interface{}) error
}

// ExecutorConfig 执行器配置
type ExecutorConfig struct {
	QueueCapacity    int           `json:"queue_capacity"`
	MaxWorkers       int           `json:"max_workers"`
	DefaultTimeout   time.Duration `json:"default_timeout"`
	RetryDelay       time.Duration `json:"retry_delay"`
	MaxRetryDelay    time.Duration `json:"max_retry_delay"`
	EnablePriority   bool          `json:"enable_priority"`
	HistoryCapacity  int           `json:"history_capacity"`
}

// DefaultExecutorConfig 默认执行器配置
func DefaultExecutorConfig() *ExecutorConfig {
	return &ExecutorConfig{
		QueueCapacity:   1000,
		MaxWorkers:      10,
		DefaultTimeout:  30 * time.Second,
		RetryDelay:      1 * time.Second,
		MaxRetryDelay:   30 * time.Second,
		EnablePriority:  true,
		HistoryCapacity: 10000,
	}
}

// OperationExecutor 操作执行器
type OperationExecutor struct {
	config       *ExecutorConfig
	queue        *OperationQueue
	handlers     map[OperationType]OperationHandler
	records      map[string]*OperationRecord
	history      []*OperationRecord
	statusMap    map[string]OperationStatus
	mu           sync.RWMutex
	historyMu    sync.RWMutex
	running      int64
	shutdown     int64
	shutdownChan chan struct{}
	wg           sync.WaitGroup
}

// NewOperationExecutor 创建操作执行器
func NewOperationExecutor(config *ExecutorConfig) *OperationExecutor {
	if config == nil {
		config = DefaultExecutorConfig()
	}

	return &OperationExecutor{
		config:       config,
		queue:        NewOperationQueue(config.QueueCapacity),
		handlers:     make(map[OperationType]OperationHandler),
		records:      make(map[string]*OperationRecord),
		history:      make([]*OperationRecord, 0, config.HistoryCapacity),
		statusMap:    make(map[string]OperationStatus),
		shutdownChan: make(chan struct{}),
	}
}

// RegisterHandler 注册操作处理器
func (e *OperationExecutor) RegisterHandler(opType OperationType, handler OperationHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.handlers[opType] = handler
}

// Submit 提交操作
func (e *OperationExecutor) Submit(ctx context.Context, op *ParsedOperation) error {
	if atomic.LoadInt64(&e.shutdown) == 1 {
		return ErrExecutorShutdown
	}

	// 创建操作记录
	record := &OperationRecord{
		Operation: op,
		Status:    StatusPending,
		Progress:  0,
	}

	e.mu.Lock()
	if _, exists := e.records[op.ID]; exists {
		e.mu.Unlock()
		return ErrOperationAlreadyExists
	}
	e.records[op.ID] = record
	e.statusMap[op.ID] = StatusPending
	e.mu.Unlock()

	// 添加到队列
	if err := e.queue.Push(op); err != nil {
		e.mu.Lock()
		delete(e.records, op.ID)
		delete(e.statusMap, op.ID)
		e.mu.Unlock()
		return err
	}

	return nil
}

// Start 启动执行器
func (e *OperationExecutor) Start(ctx context.Context) {
	for i := 0; i < e.config.MaxWorkers; i++ {
		e.wg.Add(1)
		go e.worker(ctx)
	}
}

// Stop 停止执行器
func (e *OperationExecutor) Stop() {
	atomic.StoreInt64(&e.shutdown, 1)
	close(e.shutdownChan)
	e.wg.Wait()
}

// worker 工作协程
func (e *OperationExecutor) worker(ctx context.Context) {
	defer e.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-e.shutdownChan:
			return
		case <-e.queue.Notify():
			e.processQueue(ctx)
		}
	}
}

// processQueue 处理队列
func (e *OperationExecutor) processQueue(ctx context.Context) {
	for {
		if atomic.LoadInt64(&e.shutdown) == 1 {
			return
		}

		op := e.queue.Pop()
		if op == nil {
			return
		}

		e.executeOperation(ctx, op)
	}
}

// executeOperation 执行操作
func (e *OperationExecutor) executeOperation(ctx context.Context, op *ParsedOperation) {
	atomic.AddInt64(&e.running, 1)
	defer atomic.AddInt64(&e.running, -1)

	e.mu.RLock()
	record, exists := e.records[op.ID]
	e.mu.RUnlock()

	if !exists {
		return
	}

	// 更新状态为执行中
	e.updateStatus(op.ID, StatusExecuting, 0)
	now := time.Now()
	record.StartTime = &now

	// 获取处理器
	e.mu.RLock()
	handler, exists := e.handlers[op.Type]
	e.mu.RUnlock()

	if !exists {
		e.handleFailure(op.ID, record, errors.New("no handler registered for operation type"))
		return
	}

	// 设置超时
	timeout := e.config.DefaultTimeout
	if op.Constraints != nil && op.Constraints.Timeout > 0 {
		timeout = op.Constraints.Timeout
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 执行操作
	result, err := e.executeWithRetry(execCtx, handler, op, record)

	// 记录结束时间
	endTime := time.Now()
	record.EndTime = &endTime
	if record.StartTime != nil {
		record.Duration = endTime.Sub(*record.StartTime)
	}

	if err != nil {
		e.handleFailure(op.ID, record, err)
		return
	}

	// 成功
	record.Result = result
	e.updateStatus(op.ID, StatusSuccess, 100)
	e.addToHistory(record)
}

// executeWithRetry 带重试的执行
func (e *OperationExecutor) executeWithRetry(
	ctx context.Context,
	handler OperationHandler,
	op *ParsedOperation,
	record *OperationRecord,
) (interface{}, error) {
	maxRetries := 0
	if op.Constraints != nil {
		maxRetries = op.Constraints.MaxRetries
	}

	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			record.RetryCount = i
			e.updateStatus(op.ID, StatusExecuting, 0)

			// 计算延迟
			delay := e.config.RetryDelay * time.Duration(i)
			if delay > e.config.MaxRetryDelay {
				delay = e.config.MaxRetryDelay
			}

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		// 更新进度
		e.updateStatus(op.ID, StatusExecuting, 10+i*20)

		result, err := handler.Handle(ctx, op)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// 检查是否可重试
		if !e.isRetryable(err) {
			break
		}

		// 检查上下文
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}

	return nil, lastErr
}

// isRetryable 判断错误是否可重试
func (e *OperationExecutor) isRetryable(err error) bool {
	// 超时错误可重试
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// 网络错误可重试
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	return false
}

// handleFailure 处理失败
func (e *OperationExecutor) handleFailure(opID string, record *OperationRecord, err error) {
	record.Error = err.Error()

	// 判断错误类型
	if errors.Is(err, context.DeadlineExceeded) {
		e.updateStatus(opID, StatusTimeout, 0)
	} else if errors.Is(err, context.Canceled) {
		e.updateStatus(opID, StatusCancelled, 0)
	} else {
		e.updateStatus(opID, StatusFailed, 0)
	}

	e.addToHistory(record)
}

// updateStatus 更新状态
func (e *OperationExecutor) updateStatus(opID string, status OperationStatus, progress int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if record, exists := e.records[opID]; exists {
		record.Status = status
		record.Progress = progress
	}
	e.statusMap[opID] = status
}

// addToHistory 添加到历史记录
func (e *OperationExecutor) addToHistory(record *OperationRecord) {
	e.historyMu.Lock()
	defer e.historyMu.Unlock()

	e.history = append(e.history, record)

	// 限制历史记录数量
	if len(e.history) > e.config.HistoryCapacity {
		e.history = e.history[1:]
	}
}

// GetStatus 获取操作状态
func (e *OperationExecutor) GetStatus(opID string) (*OperationRecord, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	record, exists := e.records[opID]
	if !exists {
		return nil, ErrOperationNotFound
	}

	return record, nil
}

// Cancel 取消操作
func (e *OperationExecutor) Cancel(opID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	record, exists := e.records[opID]
	if !exists {
		return ErrOperationNotFound
	}

	// 只能取消待执行或已确认的操作
	if record.Status != StatusPending && record.Status != StatusConfirmed {
		return errors.New("cannot cancel operation in current status")
	}

	record.Status = StatusCancelled
	e.statusMap[opID] = StatusCancelled

	return nil
}

// Rollback 回滚操作
func (e *OperationExecutor) Rollback(ctx context.Context, opID string) error {
	e.mu.RLock()
	record, exists := e.records[opID]
	e.mu.RUnlock()

	if !exists {
		return ErrOperationNotFound
	}

	// 检查是否允许回滚
	if record.Operation.Constraints == nil || !record.Operation.Constraints.AllowRollback {
		return errors.New("rollback is not allowed for this operation")
	}

	// 只能回滚成功的操作
	if record.Status != StatusSuccess {
		return errors.New("can only rollback successful operations")
	}

	// 获取处理器
	e.mu.RLock()
	handler, exists := e.handlers[record.Operation.Type]
	e.mu.RUnlock()

	if !exists {
		return errors.New("no handler registered for operation type")
	}

	// 执行回滚
	if err := handler.Rollback(ctx, record.Operation, record.Result); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	// 更新状态
	e.updateStatus(opID, StatusRolledBack, 0)

	return nil
}

// GetHistory 获取历史记录
func (e *OperationExecutor) GetHistory(limit int) []*OperationRecord {
	e.historyMu.RLock()
	defer e.historyMu.RUnlock()

	if limit <= 0 || limit > len(e.history) {
		limit = len(e.history)
	}

	result := make([]*OperationRecord, limit)
	copy(result, e.history[len(e.history)-limit:])

	return result
}

// GetQueueLength 获取队列长度
func (e *OperationExecutor) GetQueueLength() int {
	return e.queue.Len()
}

// GetRunningCount 获取正在执行的操作数
func (e *OperationExecutor) GetRunningCount() int64 {
	return atomic.LoadInt64(&e.running)
}

// GetStats 获取统计信息
func (e *OperationExecutor) GetStats() *ExecutorStats {
	e.mu.RLock()
	defer e.mu.RUnlock()
	e.historyMu.RLock()
	defer e.historyMu.RUnlock()

	stats := &ExecutorStats{
		QueueLength:   e.queue.Len(),
		RunningCount:  atomic.LoadInt64(&e.running),
		TotalExecuted: len(e.history),
		ByStatus:      make(map[OperationStatus]int),
		ByType:        make(map[OperationType]int),
	}

	for _, record := range e.history {
		stats.ByStatus[record.Status]++
		stats.ByType[record.Operation.Type]++
	}

	return stats
}

// ExecutorStats 执行器统计信息
type ExecutorStats struct {
	QueueLength   int                        `json:"queue_length"`
	RunningCount  int64                      `json:"running_count"`
	TotalExecuted int                        `json:"total_executed"`
	ByStatus      map[OperationStatus]int    `json:"by_status"`
	ByType        map[OperationType]int      `json:"by_type"`
}

// BatchOperationResult 批量操作结果
type BatchOperationResult struct {
	BatchID    string                     `json:"batch_id"`
	TotalCount int                        `json:"total_count"`
	Success    int                        `json:"success"`
	Failed     int                        `json:"failed"`
	Results    map[string]*OperationRecord `json:"results"`
	StartTime  time.Time                  `json:"start_time"`
	EndTime    time.Time                  `json:"end_time"`
}

// SubmitBatch 批量提交操作
func (e *OperationExecutor) SubmitBatch(ctx context.Context, operations []*ParsedOperation) (*BatchOperationResult, error) {
	batchID := fmt.Sprintf("BATCH-%d", time.Now().UnixNano())
	result := &BatchOperationResult{
		BatchID:    batchID,
		TotalCount: len(operations),
		Results:    make(map[string]*OperationRecord),
		StartTime:  time.Now(),
	}

	for _, op := range operations {
		if err := e.Submit(ctx, op); err != nil {
			result.Failed++
			result.Results[op.ID] = &OperationRecord{
				Operation: op,
				Status:    StatusFailed,
				Error:     err.Error(),
			}
		} else {
			result.Success++
		}
	}

	result.EndTime = time.Now()
	return result, nil
}

// WaitForCompletion 等待操作完成
func (e *OperationExecutor) WaitForCompletion(ctx context.Context, opID string, timeout time.Duration) (*OperationRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			record, err := e.GetStatus(opID)
			if err != nil {
				return nil, err
			}

			// 检查是否完成
			if record.Status == StatusSuccess ||
				record.Status == StatusFailed ||
				record.Status == StatusTimeout ||
				record.Status == StatusCancelled ||
				record.Status == StatusRolledBack {
				return record, nil
			}
		}
	}
}
