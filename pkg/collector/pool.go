package collector

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/new-energy-monitoring/internal/infrastructure/logger"
	"go.uber.org/zap"
)

var (
	ErrPoolClosed    = errors.New("worker pool is closed")
	ErrPoolRunning   = errors.New("worker pool is already running")
	ErrPoolNotRunning = errors.New("worker pool is not running")
	ErrInvalidSize   = errors.New("invalid pool size")
)

// TaskFunc 任务函数类型
type TaskFunc func(ctx context.Context) error

// TaskWrapper 任务包装器
type TaskWrapper struct {
	ID       string
	Priority int
	Task     TaskFunc
	Ctx      context.Context
	ResultCh chan error
}

// PoolMetrics 协程池指标
type PoolMetrics struct {
	TotalTasks      int64 // 总任务数
	CompletedTasks  int64 // 已完成任务数
	FailedTasks     int64 // 失败任务数
	ActiveWorkers   int32 // 活跃工作协程数
	TotalWorkers    int32 // 总工作协程数
	PendingTasks    int64 // 待处理任务数
	AverageDuration int64 // 平均执行时间(纳秒)
}

// WorkerPool 协程池
type WorkerPool struct {
	// 配置
	maxWorkers     int
	minWorkers     int
	taskQueueSize  int
	idleTimeout    time.Duration
	maxIdleWorkers int

	// 状态
	running    int32
	closed     int32
	shutdownCh chan struct{}

	// 任务队列
	taskQueue chan *TaskWrapper
	priorityQueue *PriorityQueue

	// 工作协程管理
	workers    map[int]*worker
	workerPool sync.Pool
	workerID   int32

	// 指标
	metrics PoolMetrics
	metricsMutex sync.RWMutex

	// 控制
	wg         sync.WaitGroup
	ctx        context.Context
	cancelFunc context.CancelFunc

	// 日志
	logger *zap.Logger
}

// worker 工作协程
type worker struct {
	id         int
	pool       *WorkerPool
	taskCh     chan *TaskWrapper
	stopCh     chan struct{}
	lastActive time.Time
}

// PriorityQueue 优先级队列
type PriorityQueue struct {
	items []*TaskWrapper
	mutex sync.Mutex
}

// NewPriorityQueue 创建优先级队列
func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{
		items: make([]*TaskWrapper, 0),
	}
}

// Push 添加任务
func (pq *PriorityQueue) Push(item *TaskWrapper) {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()
	
	// 按优先级插入（优先级高的在前）
	inserted := false
	for i, v := range pq.items {
		if item.Priority > v.Priority {
			pq.items = append(pq.items[:i], append([]*TaskWrapper{item}, pq.items[i:]...)...)
			inserted = true
			break
		}
	}
	if !inserted {
		pq.items = append(pq.items, item)
	}
}

// Pop 取出任务
func (pq *PriorityQueue) Pop() *TaskWrapper {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()
	
	if len(pq.items) == 0 {
		return nil
	}
	
	item := pq.items[0]
	pq.items = pq.items[1:]
	return item
}

// Len 获取队列长度
func (pq *PriorityQueue) Len() int {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()
	return len(pq.items)
}

// PoolOption 协程池配置选项
type PoolOption func(*WorkerPool)

// WithMaxWorkers 设置最大工作协程数
func WithMaxWorkers(n int) PoolOption {
	return func(p *WorkerPool) {
		if n > 0 {
			p.maxWorkers = n
		}
	}
}

// WithMinWorkers 设置最小工作协程数
func WithMinWorkers(n int) PoolOption {
	return func(p *WorkerPool) {
		if n > 0 {
			p.minWorkers = n
		}
	}
}

// WithTaskQueueSize 设置任务队列大小
func WithTaskQueueSize(n int) PoolOption {
	return func(p *WorkerPool) {
		if n > 0 {
			p.taskQueueSize = n
		}
	}
}

// WithIdleTimeout 设置空闲超时时间
func WithIdleTimeout(d time.Duration) PoolOption {
	return func(p *WorkerPool) {
		if d > 0 {
			p.idleTimeout = d
		}
	}
}

// WithMaxIdleWorkers 设置最大空闲工作协程数
func WithMaxIdleWorkers(n int) PoolOption {
	return func(p *WorkerPool) {
		if n > 0 {
			p.maxIdleWorkers = n
		}
	}
}

// NewWorkerPool 创建协程池
func NewWorkerPool(opts ...PoolOption) *WorkerPool {
	// 默认配置
	p := &WorkerPool{
		maxWorkers:     1000,
		minWorkers:     10,
		taskQueueSize:  100000,
		idleTimeout:    30 * time.Second,
		maxIdleWorkers: 100,
		workers:        make(map[int]*worker),
		shutdownCh:     make(chan struct{}),
		logger:         logger.Named("worker-pool"),
	}

	// 应用配置选项
	for _, opt := range opts {
		opt(p)
	}

	// 创建任务队列
	p.taskQueue = make(chan *TaskWrapper, p.taskQueueSize)
	p.priorityQueue = NewPriorityQueue()

	// 创建上下文
	p.ctx, p.cancelFunc = context.WithCancel(context.Background())

	// 创建工作协程池
	p.workerPool = sync.Pool{
		New: func() interface{} {
			return &worker{
				taskCh: make(chan *TaskWrapper, 1),
				stopCh: make(chan struct{}),
			}
		},
	}

	return p
}

// Start 启动协程池
func (p *WorkerPool) Start() error {
	if atomic.LoadInt32(&p.running) == 1 {
		return ErrPoolRunning
	}

	atomic.StoreInt32(&p.running, 1)
	atomic.StoreInt32(&p.closed, 0)

	p.logger.Info("Starting worker pool",
		zap.Int("minWorkers", p.minWorkers),
		zap.Int("maxWorkers", p.maxWorkers),
		zap.Int("taskQueueSize", p.taskQueueSize))

	// 启动最小数量的工作协程
	for i := 0; i < p.minWorkers; i++ {
		p.addWorker()
	}

	// 启动任务分发器
	p.wg.Add(1)
	go p.dispatch()

	// 启动指标收集器
	p.wg.Add(1)
	go p.collectMetrics()

	// 启动空闲工作协程清理器
	p.wg.Add(1)
	go p.cleanIdleWorkers()

	return nil
}

// addWorker 添加工作协程
func (p *WorkerPool) addWorker() {
	if atomic.LoadInt32(&p.metrics.TotalWorkers) >= int32(p.maxWorkers) {
		return
	}

	workerID := int(atomic.AddInt32(&p.workerID, 1))
	w := &worker{
		id:         workerID,
		pool:       p,
		taskCh:     make(chan *TaskWrapper, 1),
		stopCh:     make(chan struct{}),
		lastActive: time.Now(),
	}

	p.workers[workerID] = w
	atomic.AddInt32(&p.metrics.TotalWorkers, 1)

	p.wg.Add(1)
	go w.run()
}

// run 工作协程运行
func (w *worker) run() {
	defer w.pool.wg.Done()

	for {
		select {
		case task := <-w.taskCh:
			if task == nil {
				return
			}

			w.lastActive = time.Now()
			atomic.AddInt32(&w.pool.metrics.ActiveWorkers, 1)

			// 执行任务
			startTime := time.Now()
			err := task.Task(task.Ctx)
			duration := time.Since(startTime)

			// 更新指标
			atomic.AddInt32(&w.pool.metrics.ActiveWorkers, -1)
			atomic.AddInt64(&w.pool.metrics.CompletedTasks, 1)
			
			if err != nil {
				atomic.AddInt64(&w.pool.metrics.FailedTasks, 1)
			}

			// 更新平均执行时间
			w.pool.updateAverageDuration(duration)

			// 发送结果
			if task.ResultCh != nil {
				select {
				case task.ResultCh <- err:
				case <-time.After(5 * time.Second):
					w.pool.logger.Warn("Failed to send task result", zap.String("taskID", task.ID))
				}
			}

		case <-w.stopCh:
			return

		case <-w.pool.ctx.Done():
			return
		}
	}
}

// dispatch 任务分发
func (p *WorkerPool) dispatch() {
	defer p.wg.Done()

	for {
		select {
		case task := <-p.taskQueue:
			p.dispatchTask(task)

		case <-p.ctx.Done():
			// 处理剩余任务
			for len(p.taskQueue) > 0 {
				task := <-p.taskQueue
				p.dispatchTask(task)
			}
			return

		case <-p.shutdownCh:
			return
		}
	}
}

// dispatchTask 分发任务到工作协程
func (p *WorkerPool) dispatchTask(task *TaskWrapper) {
	// 尝试找到空闲的工作协程
	for _, w := range p.workers {
		select {
		case w.taskCh <- task:
			return
		default:
			continue
		}
	}

	// 没有空闲工作协程，创建新的
	if atomic.LoadInt32(&p.metrics.TotalWorkers) < int32(p.maxWorkers) {
		p.addWorker()
		// 再次尝试分发
		for _, w := range p.workers {
			select {
			case w.taskCh <- task:
				return
			default:
				continue
			}
		}
	}

	// 放入优先级队列
	p.priorityQueue.Push(task)
	atomic.AddInt64(&p.metrics.PendingTasks, 1)
}

// collectMetrics 收集指标
func (p *WorkerPool) collectMetrics() {
	defer p.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.metricsMutex.RLock()
			p.logger.Info("Pool metrics",
				zap.Int64("totalTasks", p.metrics.TotalTasks),
				zap.Int64("completedTasks", p.metrics.CompletedTasks),
				zap.Int64("failedTasks", p.metrics.FailedTasks),
				zap.Int32("activeWorkers", p.metrics.ActiveWorkers),
				zap.Int32("totalWorkers", p.metrics.TotalWorkers),
				zap.Int64("pendingTasks", p.metrics.PendingTasks))
			p.metricsMutex.RUnlock()

		case <-p.ctx.Done():
			return
		}
	}
}

// cleanIdleWorkers 清理空闲工作协程
func (p *WorkerPool) cleanIdleWorkers() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.idleTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.cleanIdleWorkersOnce()

		case <-p.ctx.Done():
			return
		}
	}
}

// cleanIdleWorkersOnce 执行一次空闲工作协程清理
func (p *WorkerPool) cleanIdleWorkersOnce() {
	now := time.Now()
	idleWorkers := make([]*worker, 0)

	for _, w := range p.workers {
		if now.Sub(w.lastActive) > p.idleTimeout {
			idleWorkers = append(idleWorkers, w)
		}
	}

	// 保留最小数量的工作协程
	if len(p.workers)-len(idleWorkers) < p.minWorkers {
		return
	}

	// 清理空闲工作协程，但不超过最大空闲数量
	maxClean := len(idleWorkers) - p.maxIdleWorkers
	if maxClean <= 0 {
		return
	}

	for i := 0; i < maxClean && i < len(idleWorkers); i++ {
		w := idleWorkers[i]
		close(w.stopCh)
		delete(p.workers, w.id)
		atomic.AddInt32(&p.metrics.TotalWorkers, -1)
	}
}

// updateAverageDuration 更新平均执行时间
func (p *WorkerPool) updateAverageDuration(duration time.Duration) {
	p.metricsMutex.Lock()
	defer p.metricsMutex.Unlock()

	totalCompleted := atomic.LoadInt64(&p.metrics.CompletedTasks)
	if totalCompleted == 0 {
		p.metrics.AverageDuration = int64(duration)
	} else {
		// 使用移动平均
		oldAvg := p.metrics.AverageDuration
		newAvg := (oldAvg*(totalCompleted-1) + int64(duration)) / totalCompleted
		p.metrics.AverageDuration = newAvg
	}
}

// Submit 提交任务
func (p *WorkerPool) Submit(ctx context.Context, taskID string, priority int, task TaskFunc) error {
	if atomic.LoadInt32(&p.closed) == 1 {
		return ErrPoolClosed
	}

	if atomic.LoadInt32(&p.running) == 0 {
		return ErrPoolNotRunning
	}

	wrapper := &TaskWrapper{
		ID:       taskID,
		Priority: priority,
		Task:     task,
		Ctx:      ctx,
		ResultCh: make(chan error, 1),
	}

	atomic.AddInt64(&p.metrics.TotalTasks, 1)

	select {
	case p.taskQueue <- wrapper:
		return nil
	case <-ctx.Done():
		atomic.AddInt64(&p.metrics.FailedTasks, 1)
		return ctx.Err()
	case <-p.shutdownCh:
		return ErrPoolClosed
	}
}

// SubmitAndWait 提交任务并等待结果
func (p *WorkerPool) SubmitAndWait(ctx context.Context, taskID string, priority int, task TaskFunc) error {
	if atomic.LoadInt32(&p.closed) == 1 {
		return ErrPoolClosed
	}

	if atomic.LoadInt32(&p.running) == 0 {
		return ErrPoolNotRunning
	}

	resultCh := make(chan error, 1)
	wrapper := &TaskWrapper{
		ID:       taskID,
		Priority: priority,
		Task:     task,
		Ctx:      ctx,
		ResultCh: resultCh,
	}

	atomic.AddInt64(&p.metrics.TotalTasks, 1)

	select {
	case p.taskQueue <- wrapper:
		select {
		case err := <-resultCh:
			return err
		case <-ctx.Done():
			return ctx.Err()
		case <-p.shutdownCh:
			return ErrPoolClosed
		}
	case <-ctx.Done():
		atomic.AddInt64(&p.metrics.FailedTasks, 1)
		return ctx.Err()
	case <-p.shutdownCh:
		return ErrPoolClosed
	}
}

// SetSize 动态调整池大小
func (p *WorkerPool) SetSize(minWorkers, maxWorkers int) error {
	if minWorkers <= 0 || maxWorkers <= 0 || minWorkers > maxWorkers {
		return ErrInvalidSize
	}

	p.minWorkers = minWorkers
	p.maxWorkers = maxWorkers

	p.logger.Info("Pool size adjusted",
		zap.Int("minWorkers", minWorkers),
		zap.Int("maxWorkers", maxWorkers))

	// 如果当前工作协程数小于最小值，添加工作协程
	currentWorkers := int(atomic.LoadInt32(&p.metrics.TotalWorkers))
	if currentWorkers < minWorkers {
		for i := currentWorkers; i < minWorkers; i++ {
			p.addWorker()
		}
	}

	return nil
}

// GracefulShutdown 优雅关闭
func (p *WorkerPool) GracefulShutdown(timeout time.Duration) error {
	if atomic.LoadInt32(&p.closed) == 1 {
		return ErrPoolClosed
	}

	p.logger.Info("Starting graceful shutdown", zap.Duration("timeout", timeout))

	atomic.StoreInt32(&p.running, 0)
	atomic.StoreInt32(&p.closed, 1)

	// 关闭shutdown通道
	close(p.shutdownCh)

	// 取消上下文
	p.cancelFunc()

	// 等待所有工作协程完成或超时
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		p.logger.Info("Worker pool shutdown completed")
		return nil
	case <-time.After(timeout):
		p.logger.Warn("Worker pool shutdown timeout")
		return errors.New("shutdown timeout")
	}
}

// GetMetrics 获取指标
func (p *WorkerPool) GetMetrics() PoolMetrics {
	p.metricsMutex.RLock()
	defer p.metricsMutex.RUnlock()

	return PoolMetrics{
		TotalTasks:      atomic.LoadInt64(&p.metrics.TotalTasks),
		CompletedTasks:  atomic.LoadInt64(&p.metrics.CompletedTasks),
		FailedTasks:     atomic.LoadInt64(&p.metrics.FailedTasks),
		ActiveWorkers:   atomic.LoadInt32(&p.metrics.ActiveWorkers),
		TotalWorkers:    atomic.LoadInt32(&p.metrics.TotalWorkers),
		PendingTasks:    atomic.LoadInt64(&p.metrics.PendingTasks),
		AverageDuration: p.metrics.AverageDuration,
	}
}

// IsRunning 检查是否运行中
func (p *WorkerPool) IsRunning() bool {
	return atomic.LoadInt32(&p.running) == 1
}

// IsClosed 检查是否已关闭
func (p *WorkerPool) IsClosed() bool {
	return atomic.LoadInt32(&p.closed) == 1
}

// GetQueueSize 获取任务队列大小
func (p *WorkerPool) GetQueueSize() int {
	return len(p.taskQueue) + p.priorityQueue.Len()
}
