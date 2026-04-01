package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/new-energy-monitoring/internal/infrastructure/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var (
	ErrExecutorNotRunning   = errors.New("executor is not running")
	ErrExecutorRunning      = errors.New("executor is already running")
	ErrTaskTimeout          = errors.New("task execution timeout")
	ErrTaskCancelled        = errors.New("task cancelled")
	ErrMaxConcurrency       = errors.New("max concurrency reached")
	ErrInvalidTaskHandler   = errors.New("invalid task handler")
	ErrTaskHandlerPanic     = errors.New("task handler panic")
)

// Prometheus指标
var (
	executorTasksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "statistics_executor_tasks_total",
		Help: "Total number of executed tasks",
	}, []string{"status", "type"})

	executorTasksActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "statistics_executor_tasks_active",
		Help: "Number of active tasks being executed",
	})

	executorTaskDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "statistics_executor_task_duration_seconds",
		Help:    "Task execution duration in seconds",
		Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60, 120, 300},
	}, []string{"type"})

	executorConcurrency = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "statistics_executor_concurrency",
		Help: "Current concurrency level",
	})

	executorQueueSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "statistics_executor_queue_size",
		Help: "Current queue size",
	})
)

// TaskHandler 任务处理函数
type TaskHandler func(ctx *ExecutionContext) (*ExecutionResult, error)

// ExecutionContext 任务执行上下文
type ExecutionContext struct {
	TaskID       string                 `json:"taskId"`
	TaskName     string                 `json:"taskName"`
	TaskType     string                 `json:"taskType"`
	ShardIndex   int                    `json:"shardIndex"`
	TotalShards  int                    `json:"totalShards"`
	Timeout      time.Duration          `json:"timeout"`
	MaxRetry     int                    `json:"maxRetry"`
	RetryCount   int                    `json:"retryCount"`
	Config       map[string]interface{} `json:"config"`
	Labels       map[string]string      `json:"labels"`
	StartTime    time.Time              `json:"startTime"`
	Deadline     time.Time              `json:"deadline"`
	TraceID      string                 `json:"traceId"`
	SpanID       string                 `json:"spanId"`
	ParentSpanID string                 `json:"parentSpanId"`
	
	// 内部字段
	ctx          context.Context
	cancel       context.CancelFunc
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	TaskID      string                 `json:"taskId"`
	Success     bool                   `json:"success"`
	Data        interface{}            `json:"data"`
	Error       string                 `json:"error"`
	Metrics     map[string]interface{} `json:"metrics"`
	Records     []Record               `json:"records"`
	Duration    time.Duration          `json:"duration"`
	StartTime   time.Time              `json:"startTime"`
	EndTime     time.Time              `json:"endTime"`
	ProcessedAt int64                  `json:"processedAt"`
}

// Record 执行记录
type Record struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	Timestamp int64       `json:"timestamp"`
	Tags      map[string]string `json:"tags"`
}

// TaskExecutor 任务执行器接口
type TaskExecutor interface {
	Execute(ctx context.Context, execCtx *ExecutionContext) (*ExecutionResult, error)
}

// ExecutorConfig 执行器配置
type ExecutorConfig struct {
	MaxConcurrency    int           `json:"maxConcurrency"`    // 最大并发数
	QueueSize         int           `json:"queueSize"`         // 队列大小
	DefaultTimeout    time.Duration `json:"defaultTimeout"`    // 默认超时
	EnableMetrics     bool          `json:"enableMetrics"`     // 启用指标
	PanicRecovery     bool          `json:"panicRecovery"`     // Panic恢复
	WorkerCount       int           `json:"workerCount"`       // 工作协程数
	ShutdownTimeout   time.Duration `json:"shutdownTimeout"`   // 关闭超时
}

// DefaultExecutorConfig 默认执行器配置
func DefaultExecutorConfig() *ExecutorConfig {
	return &ExecutorConfig{
		MaxConcurrency:  100,
		QueueSize:       10000,
		DefaultTimeout:  30 * time.Second,
		EnableMetrics:   true,
		PanicRecovery:   true,
		WorkerCount:     10,
		ShutdownTimeout: 30 * time.Second,
	}
}

// TaskExecutorImpl 任务执行器实现
type TaskExecutorImpl struct {
	config     *ExecutorConfig
	running    int32
	closed     int32

	// 任务处理器
	handlers   map[string]TaskHandler
	handlerMu  sync.RWMutex

	// 任务队列
	taskQueue  chan *executionRequest

	// 并发控制
	semaphore  chan struct{}
	active     int32

	// 执行记录
	records    []*ExecutionResult
	recordMu   sync.RWMutex
	maxRecords int

	// 控制
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup

	// 日志
	logger     *zap.Logger
}

// executionRequest 执行请求
type executionRequest struct {
	ctx       context.Context
	execCtx   *ExecutionContext
	handler   TaskHandler
	resultCh  chan *executionResponse
}

// executionResponse 执行响应
type executionResponse struct {
	result *ExecutionResult
	err    error
}

// NewTaskExecutor 创建任务执行器
func NewTaskExecutor(config *ExecutorConfig) *TaskExecutorImpl {
	if config == nil {
		config = DefaultExecutorConfig()
	}

	e := &TaskExecutorImpl{
		config:    config,
		handlers:  make(map[string]TaskHandler),
		taskQueue: make(chan *executionRequest, config.QueueSize),
		semaphore: make(chan struct{}, config.MaxConcurrency),
		records:   make([]*ExecutionResult, 0),
		maxRecords: 10000,
		logger:    logger.Named("task-executor"),
	}

	e.ctx, e.cancel = context.WithCancel(context.Background())

	return e
}

// Start 启动执行器
func (e *TaskExecutorImpl) Start() error {
	if atomic.LoadInt32(&e.running) == 1 {
		return ErrExecutorRunning
	}

	atomic.StoreInt32(&e.running, 1)
	atomic.StoreInt32(&e.closed, 0)

	e.logger.Info("Starting task executor",
		zap.Int("maxConcurrency", e.config.MaxConcurrency),
		zap.Int("workerCount", e.config.WorkerCount))

	// 启动工作协程
	for i := 0; i < e.config.WorkerCount; i++ {
		e.wg.Add(1)
		go e.worker(i)
	}

	// 启动指标收集
	if e.config.EnableMetrics {
		e.wg.Add(1)
		go e.collectMetrics()
	}

	return nil
}

// Stop 停止执行器
func (e *TaskExecutorImpl) Stop() error {
	if atomic.LoadInt32(&e.running) == 0 {
		return ErrExecutorNotRunning
	}

	e.logger.Info("Stopping task executor")

	atomic.StoreInt32(&e.running, 0)
	atomic.StoreInt32(&e.closed, 1)

	// 取消上下文
	e.cancel()

	// 等待所有协程完成
	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		e.logger.Info("Task executor stopped successfully")
		return nil
	case <-time.After(e.config.ShutdownTimeout):
		e.logger.Warn("Task executor stop timeout")
		return errors.New("stop timeout")
	}
}

// RegisterHandler 注册任务处理器
func (e *TaskExecutorImpl) RegisterHandler(taskType string, handler TaskHandler) error {
	if handler == nil {
		return ErrInvalidTaskHandler
	}

	e.handlerMu.Lock()
	defer e.handlerMu.Unlock()

	e.handlers[taskType] = handler
	e.logger.Info("Task handler registered", zap.String("taskType", taskType))
	return nil
}

// UnregisterHandler 注销任务处理器
func (e *TaskExecutorImpl) UnregisterHandler(taskType string) {
	e.handlerMu.Lock()
	defer e.handlerMu.Unlock()

	delete(e.handlers, taskType)
	e.logger.Info("Task handler unregistered", zap.String("taskType", taskType))
}

// Execute 执行任务
func (e *TaskExecutorImpl) Execute(ctx context.Context, execCtx *ExecutionContext) (*ExecutionResult, error) {
	if atomic.LoadInt32(&e.running) == 0 {
		return nil, ErrExecutorNotRunning
	}

	// 获取处理器
	e.handlerMu.RLock()
	handler, ok := e.handlers[execCtx.TaskType]
	e.handlerMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("no handler for task type: %s", execCtx.TaskType)
	}

	// 设置默认超时
	if execCtx.Timeout == 0 {
		execCtx.Timeout = e.config.DefaultTimeout
	}

	// 设置开始时间
	execCtx.StartTime = time.Now()
	execCtx.Deadline = execCtx.StartTime.Add(execCtx.Timeout)

	// 创建带超时的上下文
	childCtx, cancel := context.WithTimeout(ctx, execCtx.Timeout)
	execCtx.ctx = childCtx
	execCtx.cancel = cancel
	defer cancel()

	// 创建执行请求
	resultCh := make(chan *executionResponse, 1)
	req := &executionRequest{
		ctx:      childCtx,
		execCtx:  execCtx,
		handler:  handler,
		resultCh: resultCh,
	}

	// 发送到队列
	select {
	case e.taskQueue <- req:
		executorQueueSize.Set(float64(len(e.taskQueue)))
	case <-ctx.Done():
		return nil, ErrTaskCancelled
	default:
		return nil, ErrMaxConcurrency
	}

	// 等待结果
	select {
	case resp := <-resultCh:
		return resp.result, resp.err
	case <-ctx.Done():
		return nil, ErrTaskCancelled
	case <-childCtx.Done():
		return nil, ErrTaskTimeout
	}
}

// ExecuteSync 同步执行任务
func (e *TaskExecutorImpl) ExecuteSync(ctx context.Context, execCtx *ExecutionContext, handler TaskHandler) (*ExecutionResult, error) {
	if atomic.LoadInt32(&e.running) == 0 {
		return nil, ErrExecutorNotRunning
	}

	// 设置默认超时
	if execCtx.Timeout == 0 {
		execCtx.Timeout = e.config.DefaultTimeout
	}

	// 设置开始时间
	execCtx.StartTime = time.Now()
	execCtx.Deadline = execCtx.StartTime.Add(execCtx.Timeout)

	// 创建带超时的上下文
	childCtx, cancel := context.WithTimeout(ctx, execCtx.Timeout)
	execCtx.ctx = childCtx
	execCtx.cancel = cancel
	defer cancel()

	// 执行任务
	return e.executeHandler(childCtx, execCtx, handler)
}

// worker 工作协程
func (e *TaskExecutorImpl) worker(id int) {
	defer e.wg.Done()

	for {
		select {
		case req := <-e.taskQueue:
			if req == nil {
				continue
			}
			e.processRequest(req)

		case <-e.ctx.Done():
			return
		}
	}
}

// processRequest 处理执行请求
func (e *TaskExecutorImpl) processRequest(req *executionRequest) {
	// 获取信号量
	select {
	case e.semaphore <- struct{}{}:
		atomic.AddInt32(&e.active, 1)
		executorTasksActive.Inc()
		executorConcurrency.Set(float64(atomic.LoadInt32(&e.active)))
		defer func() {
			<-e.semaphore
			atomic.AddInt32(&e.active, -1)
			executorTasksActive.Dec()
			executorConcurrency.Set(float64(atomic.LoadInt32(&e.active)))
		}()
	case <-req.ctx.Done():
		req.resultCh <- &executionResponse{
			result: nil,
			err:    ErrTaskTimeout,
		}
		return
	}

	// 执行任务
	result, err := e.executeHandler(req.ctx, req.execCtx, req.handler)

	// 发送结果
	req.resultCh <- &executionResponse{
		result: result,
		err:    err,
	}
}

// executeHandler 执行处理器
func (e *TaskExecutorImpl) executeHandler(ctx context.Context, execCtx *ExecutionContext, handler TaskHandler) (*ExecutionResult, error) {
	startTime := time.Now()
	var result *ExecutionResult
	var err error

	// Panic恢复
	if e.config.PanicRecovery {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%w: %v", ErrTaskHandlerPanic, r)
				e.logger.Error("Task handler panic",
					zap.String("taskID", execCtx.TaskID),
					zap.Any("panic", r))
			}
		}()
	}

	// 执行处理器
	result, err = handler(execCtx)

	duration := time.Since(startTime)

	// 构建结果
	if result == nil {
		result = &ExecutionResult{
			TaskID:    execCtx.TaskID,
			StartTime: startTime,
			EndTime:   time.Now(),
			Duration:  duration,
		}
	} else {
		result.TaskID = execCtx.TaskID
		result.StartTime = startTime
		result.EndTime = time.Now()
		result.Duration = duration
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
	} else {
		result.Success = true
	}

	// 记录结果
	e.addRecord(result)

	// 更新指标
	if e.config.EnableMetrics {
		status := "success"
		if err != nil {
			status = "failed"
		}
		executorTasksTotal.WithLabelValues(status, execCtx.TaskType).Inc()
		executorTaskDuration.WithLabelValues(execCtx.TaskType).Observe(duration.Seconds())
	}

	return result, err
}

// addRecord 添加执行记录
func (e *TaskExecutorImpl) addRecord(result *ExecutionResult) {
	e.recordMu.Lock()
	defer e.recordMu.Unlock()

	e.records = append(e.records, result)

	// 限制记录数量
	if len(e.records) > e.maxRecords {
		e.records = e.records[1:]
	}
}

// GetRecords 获取执行记录
func (e *TaskExecutorImpl) GetRecords(taskID string, limit int) []*ExecutionResult {
	e.recordMu.RLock()
	defer e.recordMu.RUnlock()

	records := make([]*ExecutionResult, 0)
	count := 0

	for i := len(e.records) - 1; i >= 0 && count < limit; i-- {
		if taskID == "" || e.records[i].TaskID == taskID {
			records = append(records, e.records[i])
			count++
		}
	}

	return records
}

// collectMetrics 收集指标
func (e *TaskExecutorImpl) collectMetrics() {
	defer e.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			executorQueueSize.Set(float64(len(e.taskQueue)))
			executorConcurrency.Set(float64(atomic.LoadInt32(&e.active)))

		case <-e.ctx.Done():
			return
		}
	}
}

// IsRunning 检查是否运行中
func (e *TaskExecutorImpl) IsRunning() bool {
	return atomic.LoadInt32(&e.running) == 1
}

// GetActiveCount 获取活跃任务数
func (e *TaskExecutorImpl) GetActiveCount() int {
	return int(atomic.LoadInt32(&e.active))
}

// GetQueueSize 获取队列大小
func (e *TaskExecutorImpl) GetQueueSize() int {
	return len(e.taskQueue)
}

// StatisticsTaskExecutor 统计任务执行器
type StatisticsTaskExecutor struct {
	executor   *TaskExecutorImpl
	logger     *zap.Logger
}

// NewStatisticsTaskExecutor 创建统计任务执行器
func NewStatisticsTaskExecutor(config *ExecutorConfig) *StatisticsTaskExecutor {
	executor := NewTaskExecutor(config)
	
	ste := &StatisticsTaskExecutor{
		executor: executor,
		logger:   logger.Named("statistics-executor"),
	}

	// 注册默认处理器
	ste.registerDefaultHandlers()

	return ste
}

// registerDefaultHandlers 注册默认处理器
func (ste *StatisticsTaskExecutor) registerDefaultHandlers() {
	// 数据聚合任务
	ste.executor.RegisterHandler("aggregation", ste.handleAggregationTask)
	
	// 数据统计任务
	ste.executor.RegisterHandler("statistics", ste.handleStatisticsTask)
	
	// 数据清理任务
	ste.executor.RegisterHandler("cleanup", ste.handleCleanupTask)
	
	// 报表生成任务
	ste.executor.RegisterHandler("report", ste.handleReportTask)
	
	// 数据同步任务
	ste.executor.RegisterHandler("sync", ste.handleSyncTask)
}

// Start 启动执行器
func (ste *StatisticsTaskExecutor) Start() error {
	return ste.executor.Start()
}

// Stop 停止执行器
func (ste *StatisticsTaskExecutor) Stop() error {
	return ste.executor.Stop()
}

// Execute 执行任务
func (ste *StatisticsTaskExecutor) Execute(ctx context.Context, execCtx *ExecutionContext) (*ExecutionResult, error) {
	return ste.executor.Execute(ctx, execCtx)
}

// handleAggregationTask 处理聚合任务
func (ste *StatisticsTaskExecutor) handleAggregationTask(ctx *ExecutionContext) (*ExecutionResult, error) {
	ste.logger.Info("Executing aggregation task",
		zap.String("taskID", ctx.TaskID),
		zap.Int("shardIndex", ctx.ShardIndex))

	// 获取配置参数
	timeRange, _ := ctx.Config["timeRange"].(string)
	granularity, _ := ctx.Config["granularity"].(string)
	pointIDs, _ := ctx.Config["pointIds"].([]interface{})

	// 模拟聚合计算
	result := &ExecutionResult{
		Success:     true,
		ProcessedAt: time.Now().Unix(),
		Metrics: map[string]interface{}{
			"timeRange":   timeRange,
			"granularity": granularity,
			"pointCount":  len(pointIDs),
		},
		Records: make([]Record, 0),
	}

	// 检查上下文是否取消
	select {
	case <-ctx.ctx.Done():
		return nil, ctx.ctx.Err()
	default:
	}

	return result, nil
}

// handleStatisticsTask 处理统计任务
func (ste *StatisticsTaskExecutor) handleStatisticsTask(ctx *ExecutionContext) (*ExecutionResult, error) {
	ste.logger.Info("Executing statistics task",
		zap.String("taskID", ctx.TaskID),
		zap.Int("shardIndex", ctx.ShardIndex))

	// 获取配置参数
	statType, _ := ctx.Config["statType"].(string)
	timeRange, _ := ctx.Config["timeRange"].(string)

	// 模拟统计计算
	result := &ExecutionResult{
		Success:     true,
		ProcessedAt: time.Now().Unix(),
		Metrics: map[string]interface{}{
			"statType":  statType,
			"timeRange": timeRange,
		},
	}

	return result, nil
}

// handleCleanupTask 处理清理任务
func (ste *StatisticsTaskExecutor) handleCleanupTask(ctx *ExecutionContext) (*ExecutionResult, error) {
	ste.logger.Info("Executing cleanup task",
		zap.String("taskID", ctx.TaskID))

	// 获取配置参数
	retentionDays, _ := ctx.Config["retentionDays"].(int)
	tableName, _ := ctx.Config["tableName"].(string)

	// 模拟数据清理
	result := &ExecutionResult{
		Success:     true,
		ProcessedAt: time.Now().Unix(),
		Metrics: map[string]interface{}{
			"retentionDays": retentionDays,
			"tableName":     tableName,
		},
	}

	return result, nil
}

// handleReportTask 处理报表任务
func (ste *StatisticsTaskExecutor) handleReportTask(ctx *ExecutionContext) (*ExecutionResult, error) {
	ste.logger.Info("Executing report task",
		zap.String("taskID", ctx.TaskID))

	// 获取配置参数
	reportType, _ := ctx.Config["reportType"].(string)
	format, _ := ctx.Config["format"].(string)

	// 模拟报表生成
	result := &ExecutionResult{
		Success:     true,
		ProcessedAt: time.Now().Unix(),
		Metrics: map[string]interface{}{
			"reportType": reportType,
			"format":     format,
		},
	}

	return result, nil
}

// handleSyncTask 处理同步任务
func (ste *StatisticsTaskExecutor) handleSyncTask(ctx *ExecutionContext) (*ExecutionResult, error) {
	ste.logger.Info("Executing sync task",
		zap.String("taskID", ctx.TaskID))

	// 获取配置参数
	source, _ := ctx.Config["source"].(string)
	target, _ := ctx.Config["target"].(string)

	// 模拟数据同步
	result := &ExecutionResult{
		Success:     true,
		ProcessedAt: time.Now().Unix(),
		Metrics: map[string]interface{}{
			"source": source,
			"target": target,
		},
	}

	return result, nil
}

// RegisterHandler 注册任务处理器
func (ste *StatisticsTaskExecutor) RegisterHandler(taskType string, handler TaskHandler) error {
	return ste.executor.RegisterHandler(taskType, handler)
}

// GetExecutor 获取底层执行器
func (ste *StatisticsTaskExecutor) GetExecutor() *TaskExecutorImpl {
	return ste.executor
}

// ExecutorBuilder 执行器构建器
type ExecutorBuilder struct {
	config *ExecutorConfig
}

// NewExecutorBuilder 创建执行器构建器
func NewExecutorBuilder() *ExecutorBuilder {
	return &ExecutorBuilder{
		config: DefaultExecutorConfig(),
	}
}

// WithMaxConcurrency 设置最大并发数
func (b *ExecutorBuilder) WithMaxConcurrency(n int) *ExecutorBuilder {
	b.config.MaxConcurrency = n
	return b
}

// WithQueueSize 设置队列大小
func (b *ExecutorBuilder) WithQueueSize(n int) *ExecutorBuilder {
	b.config.QueueSize = n
	return b
}

// WithDefaultTimeout 设置默认超时
func (b *ExecutorBuilder) WithDefaultTimeout(d time.Duration) *ExecutorBuilder {
	b.config.DefaultTimeout = d
	return b
}

// WithWorkerCount 设置工作协程数
func (b *ExecutorBuilder) WithWorkerCount(n int) *ExecutorBuilder {
	b.config.WorkerCount = n
	return b
}

// WithPanicRecovery 设置Panic恢复
func (b *ExecutorBuilder) WithPanicRecovery(enable bool) *ExecutorBuilder {
	b.config.PanicRecovery = enable
	return b
}

// WithMetrics 设置指标
func (b *ExecutorBuilder) WithMetrics(enable bool) *ExecutorBuilder {
	b.config.EnableMetrics = enable
	return b
}

// Build 构建执行器
func (b *ExecutorBuilder) Build() *TaskExecutorImpl {
	return NewTaskExecutor(b.config)
}

// BuildStatistics 构建统计执行器
func (b *ExecutorBuilder) BuildStatistics() *StatisticsTaskExecutor {
	return NewStatisticsTaskExecutor(b.config)
}
