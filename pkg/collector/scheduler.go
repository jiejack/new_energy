package collector

import (
	"context"
	"errors"
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
)

// Prometheus指标
var (
	schedulerTasksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "collector_scheduler_tasks_total",
		Help: "Total number of tasks scheduled",
	}, []string{"type", "status"})

	schedulerTasksActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "collector_scheduler_tasks_active",
		Help: "Number of active tasks",
	})

	schedulerTaskDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "collector_scheduler_task_duration_seconds",
		Help:    "Task execution duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"task_type"})
)

// SchedulerMetrics 调度器指标
type SchedulerMetrics struct {
	TotalTasks       int64 // 总任务数
	ActiveTasks      int64 // 活跃任务数
	CompletedTasks   int64 // 已完成任务数
	FailedTasks      int64 // 失败任务数
	ScheduledTasks   int64 // 已调度任务数
	CancelledTasks   int64 // 已取消任务数
	AverageLatency   int64 // 平均延迟(纳秒)
	LastScheduleTime time.Time
}

// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	MaxConcurrentTasks int           // 最大并发任务数
	TaskQueueSize      int           // 任务队列大小
	ScheduleInterval   time.Duration // 调度间隔
	EventBufferSize    int           // 事件缓冲区大小
	EnableMetrics      bool          // 是否启用指标
}

// SchedulerOption 调度器配置选项
type SchedulerOption func(*Scheduler)

// WithMaxConcurrentTasks 设置最大并发任务数
func WithMaxConcurrentTasks(n int) SchedulerOption {
	return func(s *Scheduler) {
		if n > 0 {
			s.config.MaxConcurrentTasks = n
		}
	}
}

// WithTaskQueueSize 设置任务队列大小
func WithTaskQueueSize(n int) SchedulerOption {
	return func(s *Scheduler) {
		if n > 0 {
			s.config.TaskQueueSize = n
		}
	}
}

// WithScheduleInterval 设置调度间隔
func WithScheduleInterval(d time.Duration) SchedulerOption {
	return func(s *Scheduler) {
		if d > 0 {
			s.config.ScheduleInterval = d
		}
	}
}

// WithEventBufferSize 设置事件缓冲区大小
func WithEventBufferSize(n int) SchedulerOption {
	return func(s *Scheduler) {
		if n > 0 {
			s.config.EventBufferSize = n
		}
	}
}

// WithSchedulerEnableMetrics 设置是否启用指标
func WithSchedulerEnableMetrics(enable bool) SchedulerOption {
	return func(s *Scheduler) {
		s.config.EnableMetrics = enable
	}
}

// EventTrigger 事件触发器
type EventTrigger struct {
	EventID   string
	TaskID    string
	Timestamp time.Time
	Payload   interface{}
}

// Scheduler 采集任务调度器
type Scheduler struct {
	config     SchedulerConfig
	running    int32
	closed     int32

	// 任务管理
	tasks      map[string]*Task
	taskMutex  sync.RWMutex

	// 任务队列
	taskQueue  chan *Task
	eventQueue chan *EventTrigger

	// 协程池
	pool       *WorkerPool

	// 采集器管理
	collectors map[string]Collector
	collectorMutex sync.RWMutex

	// 事件处理器
	eventHandler EventHandler

	// 指标
	metrics      SchedulerMetrics
	metricsMutex sync.RWMutex

	// 控制
	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup

	// 日志
	logger *zap.Logger
}

// NewScheduler 创建调度器
func NewScheduler(pool *WorkerPool, opts ...SchedulerOption) *Scheduler {
	// 默认配置
	s := &Scheduler{
		config: SchedulerConfig{
			MaxConcurrentTasks: 1000,
			TaskQueueSize:      100000,
			ScheduleInterval:   100 * time.Millisecond,
			EventBufferSize:    10000,
			EnableMetrics:      true,
		},
		tasks:      make(map[string]*Task),
		collectors: make(map[string]Collector),
		pool:       pool,
		logger:     logger.Named("scheduler"),
	}

	// 应用配置选项
	for _, opt := range opts {
		opt(s)
	}

	// 创建任务队列
	s.taskQueue = make(chan *Task, s.config.TaskQueueSize)
	s.eventQueue = make(chan *EventTrigger, s.config.EventBufferSize)

	// 创建上下文
	s.ctx, s.cancelFunc = context.WithCancel(context.Background())

	return s
}

// Start 启动调度器
func (s *Scheduler) Start() error {
	if atomic.LoadInt32(&s.running) == 1 {
		return ErrSchedulerRunning
	}

	atomic.StoreInt32(&s.running, 1)
	atomic.StoreInt32(&s.closed, 0)

	s.logger.Info("Starting scheduler",
		zap.Int("maxConcurrentTasks", s.config.MaxConcurrentTasks),
		zap.Int("taskQueueSize", s.config.TaskQueueSize))

	// 启动任务调度器
	s.wg.Add(1)
	go s.schedule()

	// 启动事件处理器
	s.wg.Add(1)
	go s.handleEvents()

	// 启动指标收集器
	if s.config.EnableMetrics {
		s.wg.Add(1)
		go s.collectMetrics()
	}

	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() error {
	if atomic.LoadInt32(&s.running) == 0 {
		return ErrSchedulerNotRunning
	}

	s.logger.Info("Stopping scheduler")

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
		s.logger.Info("Scheduler stopped successfully")
		return nil
	case <-time.After(30 * time.Second):
		s.logger.Warn("Scheduler stop timeout")
		return errors.New("stop timeout")
	}
}

// AddTask 添加任务
func (s *Scheduler) AddTask(task *Task) error {
	if task == nil || task.ID == "" {
		return ErrInvalidTask
	}

	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	if _, exists := s.tasks[task.ID]; exists {
		return ErrTaskExists
	}

	// 设置任务创建时间
	if task.CreateTime.IsZero() {
		task.CreateTime = time.Now()
	}

	// 设置下次执行时间
	if task.Type == TaskTypePeriodic && task.Interval > 0 {
		task.NextRunTime = task.CreateTime.Add(task.Interval)
	}

	// 设置默认状态
	if task.Status == TaskStatusCancelled {
		task.Status = TaskStatusPending
	}

	s.tasks[task.ID] = task
	atomic.AddInt64(&s.metrics.TotalTasks, 1)

	s.logger.Info("Task added",
		zap.String("taskID", task.ID),
		zap.String("taskName", task.Name),
		zap.String("taskType", task.Type.String()),
		zap.Int("priority", task.Priority))

	return nil
}

// RemoveTask 移除任务
func (s *Scheduler) RemoveTask(taskID string) error {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return ErrTaskNotFound
	}

	// 设置任务状态为已取消
	task.Status = TaskStatusCancelled
	delete(s.tasks, taskID)

	atomic.AddInt64(&s.metrics.CancelledTasks, 1)

	s.logger.Info("Task removed", zap.String("taskID", taskID))

	return nil
}

// GetTask 获取任务
func (s *Scheduler) GetTask(taskID string) (*Task, error) {
	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, ErrTaskNotFound
	}

	return task, nil
}

// GetAllTasks 获取所有任务
func (s *Scheduler) GetAllTasks() []*Task {
	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// RegisterCollector 注册采集器
func (s *Scheduler) RegisterCollector(collectorID string, collector Collector) error {
	s.collectorMutex.Lock()
	defer s.collectorMutex.Unlock()

	s.collectors[collectorID] = collector

	s.logger.Info("Collector registered", zap.String("collectorID", collectorID))

	return nil
}

// UnregisterCollector 注销采集器
func (s *Scheduler) UnregisterCollector(collectorID string) error {
	s.collectorMutex.Lock()
	defer s.collectorMutex.Unlock()

	delete(s.collectors, collectorID)

	s.logger.Info("Collector unregistered", zap.String("collectorID", collectorID))

	return nil
}

// SetEventHandler 设置事件处理器
func (s *Scheduler) SetEventHandler(handler EventHandler) {
	s.eventHandler = handler
}

// TriggerEvent 触发事件
func (s *Scheduler) TriggerEvent(eventID, taskID string, payload interface{}) error {
	trigger := &EventTrigger{
		EventID:   eventID,
		TaskID:    taskID,
		Timestamp: time.Now(),
		Payload:   payload,
	}

	select {
	case s.eventQueue <- trigger:
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	default:
		return errors.New("event queue is full")
	}
}

// schedule 任务调度
func (s *Scheduler) schedule() {
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
func (s *Scheduler) scheduleTasks() {
	now := time.Now()

	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()

	for _, task := range s.tasks {
		// 跳过非活跃任务
		if task.Status != TaskStatusPending {
			continue
		}

		// 检查是否到达执行时间
		if task.Type == TaskTypePeriodic {
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

	// 处理任务队列
	s.processTaskQueue()
}

// processTaskQueue 处理任务队列
func (s *Scheduler) processTaskQueue() {
	for {
		select {
		case task := <-s.taskQueue:
			if task == nil {
				continue
			}

			// 提交到协程池执行
			go s.executeTask(task)

		default:
			return
		}
	}
}

// executeTask 执行任务
func (s *Scheduler) executeTask(task *Task) {
	startTime := time.Now()

	// 获取采集器
	s.collectorMutex.RLock()
	collector, exists := s.collectors[task.CollectorID]
	s.collectorMutex.RUnlock()

	if !exists {
		s.handleTaskError(task, errors.New("collector not found"))
		return
	}

	// 创建任务上下文
	ctx, cancel := context.WithTimeout(s.ctx, task.Timeout)
	defer cancel()

	// 执行采集
	result, err := collector.Collect(ctx)

	duration := time.Since(startTime)

	// 更新任务状态
	s.taskMutex.Lock()
	if t, ok := s.tasks[task.ID]; ok {
		t.LastRunTime = startTime
		if t.Type == TaskTypePeriodic {
			t.NextRunTime = startTime.Add(t.Interval)
		}
		t.Status = TaskStatusPending
	}
	s.taskMutex.Unlock()

	// 创建任务结果
	taskResult := &TaskResult{
		TaskID:    task.ID,
		Success:   err == nil,
		Error:     err,
		Result:    result,
		StartTime: startTime,
		EndTime:   time.Now(),
		Duration:  duration,
	}

	// 处理结果
	if err != nil {
		s.handleTaskError(task, err)
		if s.eventHandler != nil {
			s.eventHandler.OnTaskFailed(taskResult)
		}
	} else {
		atomic.AddInt64(&s.metrics.CompletedTasks, 1)
		if s.eventHandler != nil {
			s.eventHandler.OnTaskComplete(taskResult)
		}
	}

	// 更新指标
	if s.config.EnableMetrics {
		schedulerTasksTotal.WithLabelValues(task.Type.String(), "success").Inc()
		schedulerTaskDuration.WithLabelValues(task.Type.String()).Observe(duration.Seconds())
	}
}

// handleTaskError 处理任务错误
func (s *Scheduler) handleTaskError(task *Task, err error) {
	atomic.AddInt64(&s.metrics.FailedTasks, 1)

	s.logger.Error("Task execution failed",
		zap.String("taskID", task.ID),
		zap.Error(err))

	if s.config.EnableMetrics {
		schedulerTasksTotal.WithLabelValues(task.Type.String(), "failed").Inc()
	}

	// 重试逻辑
	if task.MaxRetry > 0 {
		go s.retryTask(task, err)
	}
}

// retryTask 重试任务
func (s *Scheduler) retryTask(task *Task, lastError error) {
	retryCount := 0
	maxRetry := task.MaxRetry

	for retryCount < maxRetry {
		retryCount++

		s.logger.Info("Retrying task",
			zap.String("taskID", task.ID),
			zap.Int("attempt", retryCount),
			zap.Int("maxRetry", maxRetry))

		// 等待一段时间后重试
		time.Sleep(time.Duration(retryCount) * time.Second)

		// 检查任务是否还存在
		s.taskMutex.RLock()
		_, exists := s.tasks[task.ID]
		s.taskMutex.RUnlock()

		if !exists {
			return
		}

		// 重新执行任务
		s.executeTask(task)

		// 检查是否成功
		s.taskMutex.RLock()
		currentTask, exists := s.tasks[task.ID]
		s.taskMutex.RUnlock()

		if exists && currentTask.Status == TaskStatusPending {
			return
		}
	}

	s.logger.Error("Task retry exhausted",
		zap.String("taskID", task.ID),
		zap.Int("attempts", retryCount),
		zap.Error(lastError))
}

// handleEvents 处理事件
func (s *Scheduler) handleEvents() {
	defer s.wg.Done()

	for {
		select {
		case event := <-s.eventQueue:
			s.processEvent(event)

		case <-s.ctx.Done():
			return
		}
	}
}

// processEvent 处理事件
func (s *Scheduler) processEvent(event *EventTrigger) {
	s.taskMutex.RLock()
	task, exists := s.tasks[event.TaskID]
	s.taskMutex.RUnlock()

	if !exists {
		s.logger.Warn("Event triggered for non-existent task",
			zap.String("eventID", event.EventID),
			zap.String("taskID", event.TaskID))
		return
	}

	// 执行事件触发任务
	if task.Type == TaskTypeEvent {
		task.Status = TaskStatusRunning
		s.executeTask(task)
	}
}

// collectMetrics 收集指标
func (s *Scheduler) collectMetrics() {
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
func (s *Scheduler) GetMetrics() SchedulerMetrics {
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
func (s *Scheduler) IsRunning() bool {
	return atomic.LoadInt32(&s.running) == 1
}

// IsClosed 检查是否已关闭
func (s *Scheduler) IsClosed() bool {
	return atomic.LoadInt32(&s.closed) == 1
}

// GetTaskCount 获取任务数量
func (s *Scheduler) GetTaskCount() int {
	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()
	return len(s.tasks)
}

// GetCollectorCount 获取采集器数量
func (s *Scheduler) GetCollectorCount() int {
	s.collectorMutex.RLock()
	defer s.collectorMutex.RUnlock()
	return len(s.collectors)
}

// PauseTask 暂停任务
func (s *Scheduler) PauseTask(taskID string) error {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return ErrTaskNotFound
	}

	task.Status = TaskStatusCancelled

	s.logger.Info("Task paused", zap.String("taskID", taskID))

	return nil
}

// ResumeTask 恢复任务
func (s *Scheduler) ResumeTask(taskID string) error {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return ErrTaskNotFound
	}

	task.Status = TaskStatusPending

	s.logger.Info("Task resumed", zap.String("taskID", taskID))

	return nil
}

// UpdateTaskInterval 更新任务间隔
func (s *Scheduler) UpdateTaskInterval(taskID string, interval time.Duration) error {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return ErrTaskNotFound
	}

	task.Interval = interval
	if task.Type == TaskTypePeriodic {
		task.NextRunTime = time.Now().Add(interval)
	}

	s.logger.Info("Task interval updated",
		zap.String("taskID", taskID),
		zap.Duration("interval", interval))

	return nil
}
