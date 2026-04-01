package rule

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
	ErrSchedulerRunning    = errors.New("scheduler is already running")
	ErrSchedulerNotRunning = errors.New("scheduler is not running")
	ErrTaskNotFound        = errors.New("task not found")
	ErrTaskExists          = errors.New("task already exists")
	ErrInvalidTask         = errors.New("invalid task")
	ErrLockFailed          = errors.New("failed to acquire distributed lock")
)

// Prometheus指标
var (
	schedulerTasksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compute_scheduler_tasks_total",
		Help: "Total number of compute tasks scheduled",
	}, []string{"status"})

	schedulerTasksActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "compute_scheduler_tasks_active",
		Help: "Number of active compute tasks",
	})

	schedulerTaskDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "compute_scheduler_task_duration_seconds",
		Help:    "Compute task execution duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"task_type"})

	schedulerLockWait = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "compute_scheduler_lock_wait_seconds",
		Help:    "Time waiting for distributed lock",
		Buckets: prometheus.DefBuckets,
	})
)

// TaskType 任务类型
type TaskType string

const (
	TaskTypeCron     TaskType = "cron"     // Cron表达式
	TaskTypeInterval TaskType = "interval" // 固定间隔
	TaskTypeOnce     TaskType = "once"     // 一次性任务
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"   // 待执行
	TaskStatusRunning   TaskStatus = "running"   // 执行中
	TaskStatusCompleted TaskStatus = "completed" // 已完成
	TaskStatusFailed    TaskStatus = "failed"    // 失败
	TaskStatusCancelled TaskStatus = "cancelled" // 已取消
	TaskStatusPaused    TaskStatus = "paused"    // 已暂停
)

// ComputeTask 计算任务
type ComputeTask struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Type         TaskType               `json:"type"`
	Status       TaskStatus             `json:"status"`
	PointIDs     []string               `json:"pointIds"`     // 要计算的计算点ID列表
	CronExpr     string                 `json:"cronExpr"`     // Cron表达式
	Interval     time.Duration          `json:"interval"`     // 执行间隔
	Timeout      time.Duration          `json:"timeout"`      // 超时时间
	MaxRetry     int                    `json:"maxRetry"`     // 最大重试次数
	RetryDelay   time.Duration          `json:"retryDelay"`   // 重试延迟
	Priority     int                    `json:"priority"`     // 优先级
	Enabled      bool                   `json:"enabled"`      // 是否启用
	CreateTime   time.Time              `json:"createTime"`   // 创建时间
	UpdateTime   time.Time              `json:"updateTime"`   // 更新时间
	NextRunTime  time.Time              `json:"nextRunTime"`  // 下次执行时间
	LastRunTime  time.Time              `json:"lastRunTime"`  // 上次执行时间
	LastStatus   TaskStatus             `json:"lastStatus"`   // 上次执行状态
	LastError    string                 `json:"lastError"`    // 上次错误信息
	RunCount     int64                  `json:"runCount"`     // 执行次数
	SuccessCount int64                  `json:"successCount"` // 成功次数
	FailCount    int64                  `json:"failCount"`    // 失败次数
	Config       map[string]interface{} `json:"config"`       // 配置参数
}

// TaskExecutionLog 任务执行日志
type TaskExecutionLog struct {
	TaskID      string        `json:"taskId"`
	StartTime   time.Time     `json:"startTime"`
	EndTime     time.Time     `json:"endTime"`
	Duration    time.Duration `json:"duration"`
	Status      TaskStatus    `json:"status"`
	Error       string        `json:"error"`
	Results     map[string]*ComputeResult `json:"results"`
	TriggerType string        `json:"triggerType"` // manual, scheduled, event
}

// DistributedLock 分布式锁接口
type DistributedLock interface {
	Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Release(ctx context.Context, key string) error
	IsHeld(ctx context.Context, key string) (bool, error)
}

// ComputeExecutor 计算执行器接口
type ComputeExecutor interface {
	Execute(ctx context.Context, pointIDs []string) (map[string]*ComputeResult, error)
}

// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	MaxConcurrentTasks int           `json:"maxConcurrentTasks"` // 最大并发任务数
	TaskQueueSize      int           `json:"taskQueueSize"`      // 任务队列大小
	ScheduleInterval   time.Duration `json:"scheduleInterval"`   // 调度间隔
	LockTTL            time.Duration `json:"lockTTL"`            // 锁过期时间
	EnableMetrics      bool          `json:"enableMetrics"`      // 启用指标
	LogRetention       time.Duration `json:"logRetention"`       // 日志保留时间
	MaxLogSize         int           `json:"maxLogSize"`         // 最大日志数量
}

// ComputeScheduler 计算调度器
type ComputeScheduler struct {
	config     SchedulerConfig
	running    int32
	closed     int32

	// 任务管理
	tasks      map[string]*ComputeTask
	taskMutex  sync.RWMutex

	// 任务队列
	taskQueue  chan *ComputeTask
	priorityQueue *PriorityQueue

	// 执行器
	executor   ComputeExecutor

	// 分布式锁
	lock       DistributedLock

	// 执行日志
	logs       []*TaskExecutionLog
	logMutex   sync.RWMutex

	// 指标
	metrics    SchedulerMetrics
	metricsMutex sync.RWMutex

	// 控制
	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup

	// 日志
	logger     *zap.Logger
}

// SchedulerMetrics 调度器指标
type SchedulerMetrics struct {
	TotalTasks       int64
	ActiveTasks      int64
	CompletedTasks   int64
	FailedTasks      int64
	ScheduledTasks   int64
	CancelledTasks   int64
	AverageLatency   int64
	LastScheduleTime time.Time
}

// NewComputeScheduler 创建计算调度器
func NewComputeScheduler(config *SchedulerConfig, executor ComputeExecutor, lock DistributedLock) *ComputeScheduler {
	if config == nil {
		config = &SchedulerConfig{
			MaxConcurrentTasks: 100,
			TaskQueueSize:      10000,
			ScheduleInterval:   100 * time.Millisecond,
			LockTTL:            30 * time.Second,
			EnableMetrics:      true,
			LogRetention:       24 * time.Hour,
			MaxLogSize:         10000,
		}
	}

	s := &ComputeScheduler{
		config:        *config,
		tasks:         make(map[string]*ComputeTask),
		executor:      executor,
		lock:          lock,
		logs:          make([]*TaskExecutionLog, 0),
		priorityQueue: NewPriorityQueue(),
		logger:        logger.Named("compute-scheduler"),
	}

	s.taskQueue = make(chan *ComputeTask, s.config.TaskQueueSize)
	s.ctx, s.cancelFunc = context.WithCancel(context.Background())

	return s
}

// Start 启动调度器
func (s *ComputeScheduler) Start() error {
	if atomic.LoadInt32(&s.running) == 1 {
		return ErrSchedulerRunning
	}

	atomic.StoreInt32(&s.running, 1)
	atomic.StoreInt32(&s.closed, 0)

	s.logger.Info("Starting compute scheduler",
		zap.Int("maxConcurrentTasks", s.config.MaxConcurrentTasks),
		zap.Int("taskQueueSize", s.config.TaskQueueSize))

	// 启动任务调度器
	s.wg.Add(1)
	go s.schedule()

	// 启动任务执行器
	for i := 0; i < s.config.MaxConcurrentTasks; i++ {
		s.wg.Add(1)
		go s.executeWorker(i)
	}

	// 启动日志清理器
	s.wg.Add(1)
	go s.cleanupLogs()

	// 启动指标收集器
	if s.config.EnableMetrics {
		s.wg.Add(1)
		go s.collectMetrics()
	}

	return nil
}

// Stop 停止调度器
func (s *ComputeScheduler) Stop() error {
	if atomic.LoadInt32(&s.running) == 0 {
		return ErrSchedulerNotRunning
	}

	s.logger.Info("Stopping compute scheduler")

	atomic.StoreInt32(&s.running, 0)
	atomic.StoreInt32(&s.closed, 1)

	// 取消上下文
	s.cancelFunc()

	// 等待所有协程完成
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("Compute scheduler stopped successfully")
		return nil
	case <-time.After(30 * time.Second):
		s.logger.Warn("Compute scheduler stop timeout")
		return errors.New("stop timeout")
	}
}

// AddTask 添加任务
func (s *ComputeScheduler) AddTask(task *ComputeTask) error {
	if task == nil || task.ID == "" {
		return ErrInvalidTask
	}

	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	if _, exists := s.tasks[task.ID]; exists {
		return ErrTaskExists
	}

	// 设置默认值
	if task.CreateTime.IsZero() {
		task.CreateTime = time.Now()
	}
	task.UpdateTime = time.Now()
	if task.Status == "" {
		task.Status = TaskStatusPending
	}
	if task.Enabled {
		task.Status = TaskStatusPending
	} else {
		task.Status = TaskStatusPaused
	}

	// 计算下次执行时间
	if task.Enabled {
		task.NextRunTime = s.calculateNextRunTime(task)
	}

	s.tasks[task.ID] = task
	atomic.AddInt64(&s.metrics.TotalTasks, 1)

	s.logger.Info("Task added",
		zap.String("taskID", task.ID),
		zap.String("taskName", task.Name),
		zap.String("taskType", string(task.Type)),
		zap.Int("priority", task.Priority))

	return nil
}

// RemoveTask 移除任务
func (s *ComputeScheduler) RemoveTask(taskID string) error {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return ErrTaskNotFound
	}

	task.Status = TaskStatusCancelled
	delete(s.tasks, taskID)

	atomic.AddInt64(&s.metrics.CancelledTasks, 1)

	s.logger.Info("Task removed", zap.String("taskID", taskID))

	return nil
}

// GetTask 获取任务
func (s *ComputeScheduler) GetTask(taskID string) (*ComputeTask, error) {
	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, ErrTaskNotFound
	}

	return task, nil
}

// GetAllTasks 获取所有任务
func (s *ComputeScheduler) GetAllTasks() []*ComputeTask {
	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()

	tasks := make([]*ComputeTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// UpdateTask 更新任务
func (s *ComputeScheduler) UpdateTask(task *ComputeTask) error {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	existing, exists := s.tasks[task.ID]
	if !exists {
		return ErrTaskNotFound
	}

	// 保留创建时间和统计信息
	task.CreateTime = existing.CreateTime
	task.RunCount = existing.RunCount
	task.SuccessCount = existing.SuccessCount
	task.FailCount = existing.FailCount
	task.UpdateTime = time.Now()

	// 重新计算下次执行时间
	if task.Enabled {
		task.NextRunTime = s.calculateNextRunTime(task)
	}

	s.tasks[task.ID] = task

	s.logger.Info("Task updated", zap.String("taskID", task.ID))

	return nil
}

// PauseTask 暂停任务
func (s *ComputeScheduler) PauseTask(taskID string) error {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return ErrTaskNotFound
	}

	task.Status = TaskStatusPaused
	task.Enabled = false
	task.UpdateTime = time.Now()

	s.logger.Info("Task paused", zap.String("taskID", taskID))

	return nil
}

// ResumeTask 恢复任务
func (s *ComputeScheduler) ResumeTask(taskID string) error {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return ErrTaskNotFound
	}

	task.Status = TaskStatusPending
	task.Enabled = true
	task.UpdateTime = time.Now()
	task.NextRunTime = s.calculateNextRunTime(task)

	s.logger.Info("Task resumed", zap.String("taskID", taskID))

	return nil
}

// TriggerTask 手动触发任务
func (s *ComputeScheduler) TriggerTask(taskID string) error {
	s.taskMutex.RLock()
	task, exists := s.tasks[taskID]
	s.taskMutex.RUnlock()

	if !exists {
		return ErrTaskNotFound
	}

	// 直接提交到执行队列
	select {
	case s.taskQueue <- task:
		s.logger.Info("Task triggered manually", zap.String("taskID", taskID))
		return nil
	default:
		return errors.New("task queue is full")
	}
}

// schedule 任务调度
func (s *ComputeScheduler) schedule() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.ScheduleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.scheduleTasks()

		case <-s.ctx.Done():
			return
		}
	}
}

// scheduleTasks 调度任务
func (s *ComputeScheduler) scheduleTasks() {
	now := time.Now()

	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()

	for _, task := range s.tasks {
		// 跳过非活跃任务
		if task.Status != TaskStatusPending || !task.Enabled {
			continue
		}

		// 检查是否到达执行时间
		if now.Before(task.NextRunTime) {
			continue
		}

		// 提交任务到队列
		select {
		case s.taskQueue <- task:
			task.Status = TaskStatusRunning
			atomic.AddInt64(&s.metrics.ScheduledTasks, 1)
		default:
			s.logger.Warn("Task queue is full", zap.String("taskID", task.ID))
		}
	}
}

// executeWorker 任务执行工作器
func (s *ComputeScheduler) executeWorker(workerID int) {
	defer s.wg.Done()

	for {
		select {
		case task := <-s.taskQueue:
			if task == nil {
				continue
			}
			s.executeTask(task, "scheduled")

		case <-s.ctx.Done():
			return
		}
	}
}

// executeTask 执行任务
func (s *ComputeScheduler) executeTask(task *ComputeTask, triggerType string) {
	startTime := time.Now()

	// 尝试获取分布式锁
	lockKey := fmt.Sprintf("compute:task:%s", task.ID)
	if s.lock != nil {
		acquired, err := s.lock.Acquire(s.ctx, lockKey, s.config.LockTTL)
		if err != nil {
			s.logger.Error("Failed to acquire lock",
				zap.String("taskID", task.ID),
				zap.Error(err))
			s.handleTaskError(task, err, startTime, triggerType)
			return
		}
		if !acquired {
			s.logger.Debug("Task already running on another instance",
				zap.String("taskID", task.ID))
			return
		}
		defer s.lock.Release(s.ctx, lockKey)
	}

	// 创建任务上下文
	ctx, cancel := context.WithTimeout(s.ctx, task.Timeout)
	defer cancel()

	// 执行计算
	results, err := s.executor.Execute(ctx, task.PointIDs)

	duration := time.Since(startTime)

	// 更新任务状态
	s.taskMutex.Lock()
	if t, ok := s.tasks[task.ID]; ok {
		t.LastRunTime = startTime
		t.NextRunTime = s.calculateNextRunTime(task)
		t.Status = TaskStatusPending
		t.RunCount++
		if err != nil {
			t.FailCount++
			t.LastStatus = TaskStatusFailed
			t.LastError = err.Error()
		} else {
			t.SuccessCount++
			t.LastStatus = TaskStatusCompleted
			t.LastError = ""
		}
	}
	s.taskMutex.Unlock()

	// 记录执行日志
	log := &TaskExecutionLog{
		TaskID:      task.ID,
		StartTime:   startTime,
		EndTime:     time.Now(),
		Duration:    duration,
		Status:      TaskStatusCompleted,
		Results:     results,
		TriggerType: triggerType,
	}

	if err != nil {
		log.Status = TaskStatusFailed
		log.Error = err.Error()
		s.handleTaskError(task, err, startTime, triggerType)
	} else {
		atomic.AddInt64(&s.metrics.CompletedTasks, 1)
	}

	// 保存日志
	s.addLog(log)

	// 更新指标
	if s.config.EnableMetrics {
		schedulerTasksTotal.WithLabelValues(string(log.Status)).Inc()
		schedulerTaskDuration.WithLabelValues(string(task.Type)).Observe(duration.Seconds())
	}
}

// handleTaskError 处理任务错误
func (s *ComputeScheduler) handleTaskError(task *ComputeTask, err error, startTime time.Time, triggerType string) {
	atomic.AddInt64(&s.metrics.FailedTasks, 1)

	s.logger.Error("Task execution failed",
		zap.String("taskID", task.ID),
		zap.Error(err))

	// 重试逻辑
	if task.MaxRetry > 0 {
		go s.retryTask(task, err, triggerType)
	}
}

// retryTask 重试任务
func (s *ComputeScheduler) retryTask(task *ComputeTask, lastError error, triggerType string) {
	retryCount := 0
	maxRetry := task.MaxRetry

	for retryCount < maxRetry {
		retryCount++

		s.logger.Info("Retrying task",
			zap.String("taskID", task.ID),
			zap.Int("attempt", retryCount),
			zap.Int("maxRetry", maxRetry))

		// 等待一段时间后重试
		time.Sleep(task.RetryDelay)

		// 检查任务是否还存在
		s.taskMutex.RLock()
		_, exists := s.tasks[task.ID]
		s.taskMutex.RUnlock()

		if !exists {
			return
		}

		// 重新执行任务
		s.executeTask(task, "retry")

		// 检查是否成功
		s.taskMutex.RLock()
		currentTask, exists := s.tasks[task.ID]
		s.taskMutex.RUnlock()

		if exists && currentTask.LastStatus == TaskStatusCompleted {
			return
		}
	}

	s.logger.Error("Task retry exhausted",
		zap.String("taskID", task.ID),
		zap.Int("attempts", retryCount),
		zap.Error(lastError))
}

// calculateNextRunTime 计算下次执行时间
func (s *ComputeScheduler) calculateNextRunTime(task *ComputeTask) time.Time {
	now := time.Now()

	switch task.Type {
	case TaskTypeCron:
		// 解析Cron表达式并计算下次执行时间
		return s.parseCronNextTime(task.CronExpr, now)

	case TaskTypeInterval:
		return now.Add(task.Interval)

	case TaskTypeOnce:
		if task.NextRunTime.IsZero() {
			return now
		}
		return task.NextRunTime

	default:
		return now.Add(task.Interval)
	}
}

// parseCronNextTime 解析Cron表达式并计算下次执行时间
func (s *ComputeScheduler) parseCronNextTime(cronExpr string, from time.Time) time.Time {
	// 简化实现：解析基本的Cron表达式
	// 格式: "秒 分 时 日 月 周"
	// 实际项目中可以使用 github.com/robfig/cron 等库

	// 这里提供一个简化版本
	// 支持: "@every 1h", "@hourly", "@daily" 等预定义表达式
	switch cronExpr {
	case "@every 1m":
		return from.Add(time.Minute)
	case "@every 5m":
		return from.Add(5 * time.Minute)
	case "@every 15m":
		return from.Add(15 * time.Minute)
	case "@every 30m":
		return from.Add(30 * time.Minute)
	case "@hourly":
		return from.Add(time.Hour)
	case "@daily":
		return from.Add(24 * time.Hour)
	case "@weekly":
		return from.Add(7 * 24 * time.Hour)
	default:
		// 默认1分钟
		return from.Add(time.Minute)
	}
}

// addLog 添加执行日志
func (s *ComputeScheduler) addLog(log *TaskExecutionLog) {
	s.logMutex.Lock()
	defer s.logMutex.Unlock()

	s.logs = append(s.logs, log)

	// 限制日志数量
	if len(s.logs) > s.config.MaxLogSize {
		s.logs = s.logs[1:]
	}
}

// GetLogs 获取执行日志
func (s *ComputeScheduler) GetLogs(taskID string, limit int) []*TaskExecutionLog {
	s.logMutex.RLock()
	defer s.logMutex.RUnlock()

	logs := make([]*TaskExecutionLog, 0)
	count := 0

	for i := len(s.logs) - 1; i >= 0 && count < limit; i-- {
		if taskID == "" || s.logs[i].TaskID == taskID {
			logs = append(logs, s.logs[i])
			count++
		}
	}

	return logs
}

// cleanupLogs 清理过期日志
func (s *ComputeScheduler) cleanupLogs() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.logMutex.Lock()
			now := time.Now()
			validLogs := make([]*TaskExecutionLog, 0)
			for _, log := range s.logs {
				if now.Sub(log.StartTime) < s.config.LogRetention {
					validLogs = append(validLogs, log)
				}
			}
			s.logs = validLogs
			s.logMutex.Unlock()

		case <-s.ctx.Done():
			return
		}
	}
}

// collectMetrics 收集指标
func (s *ComputeScheduler) collectMetrics() {
	defer s.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.taskMutex.RLock()
			activeCount := int64(len(s.tasks))
			s.taskMutex.RUnlock()

			schedulerTasksActive.Set(float64(activeCount))

			s.metricsMutex.Lock()
			s.metrics.ActiveTasks = activeCount
			s.metrics.LastScheduleTime = time.Now()
			s.metricsMutex.Unlock()

		case <-s.ctx.Done():
			return
		}
	}
}

// GetMetrics 获取指标
func (s *ComputeScheduler) GetMetrics() SchedulerMetrics {
	s.metricsMutex.RLock()
	defer s.metricsMutex.RUnlock()

	return SchedulerMetrics{
		TotalTasks:       atomic.LoadInt64(&s.metrics.TotalTasks),
		ActiveTasks:      atomic.LoadInt64(&s.metrics.ActiveTasks),
		CompletedTasks:   atomic.LoadInt64(&s.metrics.CompletedTasks),
		FailedTasks:      atomic.LoadInt64(&s.metrics.FailedTasks),
		ScheduledTasks:   atomic.LoadInt64(&s.metrics.ScheduledTasks),
		CancelledTasks:   atomic.LoadInt64(&s.metrics.CancelledTasks),
		AverageLatency:   s.metrics.AverageLatency,
		LastScheduleTime: s.metrics.LastScheduleTime,
	}
}

// IsRunning 检查是否运行中
func (s *ComputeScheduler) IsRunning() bool {
	return atomic.LoadInt32(&s.running) == 1
}

// GetTaskCount 获取任务数量
func (s *ComputeScheduler) GetTaskCount() int {
	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()
	return len(s.tasks)
}

// PriorityQueue 优先级队列
type PriorityQueue struct {
	items []*ComputeTask
	mu    sync.Mutex
}

// NewPriorityQueue 创建优先级队列
func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{
		items: make([]*ComputeTask, 0),
	}
}

// Push 添加任务
func (pq *PriorityQueue) Push(task *ComputeTask) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	pq.items = append(pq.items, task)
	pq.sort()
}

// Pop 取出任务
func (pq *PriorityQueue) Pop() *ComputeTask {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if len(pq.items) == 0 {
		return nil
	}

	task := pq.items[0]
	pq.items = pq.items[1:]
	return task
}

// sort 排序
func (pq *PriorityQueue) sort() {
	// 按优先级和下次执行时间排序
	for i := 0; i < len(pq.items)-1; i++ {
		for j := i + 1; j < len(pq.items); j++ {
			if pq.items[i].Priority < pq.items[j].Priority ||
				(pq.items[i].Priority == pq.items[j].Priority &&
					pq.items[i].NextRunTime.After(pq.items[j].NextRunTime)) {
				pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
			}
		}
	}
}

// Len 获取长度
func (pq *PriorityQueue) Len() int {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	return len(pq.items)
}
