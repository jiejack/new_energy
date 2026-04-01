package scheduler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/new-energy-monitoring/internal/infrastructure/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var (
	ErrSchedulerRunning      = errors.New("scheduler is already running")
	ErrSchedulerNotRunning   = errors.New("scheduler is not running")
	ErrTaskNotFound          = errors.New("task not found")
	ErrTaskExists            = errors.New("task already exists")
	ErrInvalidTask           = errors.New("invalid task")
	ErrLockAcquireFailed     = errors.New("failed to acquire distributed lock")
	ErrNoAvailableNodes      = errors.New("no available nodes in cluster")
	ErrNodeNotRegistered     = errors.New("node not registered")
	ErrTaskRebalanceFailed   = errors.New("task rebalance failed")
)

// Prometheus指标
var (
	distributedTasksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "statistics_scheduler_distributed_tasks_total",
		Help: "Total number of distributed tasks",
	}, []string{"status"})

	distributedTasksActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "statistics_scheduler_distributed_tasks_active",
		Help: "Number of active distributed tasks",
	})

	distributedLockWait = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "statistics_scheduler_distributed_lock_wait_seconds",
		Help:    "Time waiting for distributed lock",
		Buckets: prometheus.DefBuckets,
	})

	distributedNodeCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "statistics_scheduler_distributed_node_count",
		Help: "Number of active nodes in cluster",
	})

	distributedRebalanceTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "statistics_scheduler_distributed_rebalance_total",
		Help: "Total number of task rebalances",
	}, []string{"status"})
)

// NodeState 节点状态
type NodeState string

const (
	NodeStateActive    NodeState = "active"
	NodeStateInactive  NodeState = "inactive"
	NodeStateLeaving   NodeState = "leaving"
	NodeStateJoining   NodeState = "joining"
)

// TaskPriority 任务优先级
type TaskPriority int

const (
	PriorityLow      TaskPriority = 0
	PriorityNormal   TaskPriority = 1
	PriorityHigh     TaskPriority = 2
	PriorityCritical TaskPriority = 3
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
	TaskStatusPaused    TaskStatus = "paused"
)

// ClusterNode 集群节点
type ClusterNode struct {
	NodeID        string            `json:"nodeId"`
	Address       string            `json:"address"`
	State         NodeState         `json:"state"`
	LastHeartbeat time.Time         `json:"lastHeartbeat"`
	TaskCount     int               `json:"taskCount"`
	Load          float64           `json:"load"`
	Capacity      int               `json:"capacity"`
	Labels        map[string]string `json:"labels"`
	StartTime     time.Time         `json:"startTime"`
	Version       string            `json:"version"`
}

// DistributedTask 分布式任务
type DistributedTask struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	CronExpr       string            `json:"cronExpr"`
	Priority       TaskPriority      `json:"priority"`
	Status         TaskStatus        `json:"status"`
	ShardingKey    string            `json:"shardingKey"`
	ShardingCount  int               `json:"shardingCount"`
	ShardingIndex  int               `json:"shardingIndex"`
	AssignedNode   string            `json:"assignedNode"`
	Timeout        time.Duration     `json:"timeout"`
	MaxRetry       int               `json:"maxRetry"`
	RetryCount     int               `json:"retryCount"`
	CreateTime     time.Time         `json:"createTime"`
	UpdateTime     time.Time         `json:"updateTime"`
	NextRunTime    time.Time         `json:"nextRunTime"`
	LastRunTime    time.Time         `json:"lastRunTime"`
	LastStatus     TaskStatus        `json:"lastStatus"`
	LastError      string            `json:"lastError"`
	RunCount       int64             `json:"runCount"`
	SuccessCount   int64             `json:"successCount"`
	FailCount      int64             `json:"failCount"`
	Enabled        bool              `json:"enabled"`
	Config         map[string]interface{} `json:"config"`
	Labels         map[string]string `json:"labels"`
	cronExpr       *CronExpression   `json:"-"`
}

// TaskShard 任务分片
type TaskShard struct {
	TaskID       string    `json:"taskId"`
	ShardIndex   int       `json:"shardIndex"`
	TotalShards  int       `json:"totalShards"`
	AssignedNode string    `json:"assignedNode"`
	Status       TaskStatus `json:"status"`
	CreateTime   time.Time `json:"createTime"`
}

// DistributedLock 分布式锁
type DistributedLock struct {
	client    *redis.Client
	key       string
	value     string
	ttl       time.Duration
	renewChan chan struct{}
	stopChan  chan struct{}
	mu        sync.Mutex
}

// DistributedSchedulerConfig 分布式调度器配置
type DistributedSchedulerConfig struct {
	NodeID              string        `json:"nodeId"`
	Address             string        `json:"address"`
	RedisAddr           string        `json:"redisAddr"`
	RedisPassword       string        `json:"redisPassword"`
	RedisDB             int           `json:"redisDB"`
	MaxConcurrentTasks  int           `json:"maxConcurrentTasks"`
	TaskQueueSize       int           `json:"taskQueueSize"`
	ScheduleInterval    time.Duration `json:"scheduleInterval"`
	LockTTL             time.Duration `json:"lockTTL"`
	HeartbeatInterval   time.Duration `json:"heartbeatInterval"`
	HeartbeatTimeout    time.Duration `json:"heartbeatTimeout"`
	RebalanceInterval   time.Duration `json:"rebalanceInterval"`
	EnableAutoRebalance bool          `json:"enableAutoRebalance"`
	EnableMetrics       bool          `json:"enableMetrics"`
	NodeCapacity        int           `json:"nodeCapacity"`
}

// DefaultDistributedSchedulerConfig 默认配置
func DefaultDistributedSchedulerConfig() *DistributedSchedulerConfig {
	return &DistributedSchedulerConfig{
		MaxConcurrentTasks:  100,
		TaskQueueSize:       10000,
		ScheduleInterval:    100 * time.Millisecond,
		LockTTL:             30 * time.Second,
		HeartbeatInterval:   5 * time.Second,
		HeartbeatTimeout:    15 * time.Second,
		RebalanceInterval:   30 * time.Second,
		EnableAutoRebalance: true,
		EnableMetrics:       true,
		NodeCapacity:        100,
	}
}

// DistributedScheduler 分布式任务调度器
type DistributedScheduler struct {
	config     *DistributedSchedulerConfig
	redis      *redis.Client
	running    int32
	closed     int32

	// 本节点信息
	nodeID     string
	nodeState  NodeState

	// 任务管理
	tasks      map[string]*DistributedTask
	taskMutex  sync.RWMutex

	// 分片管理
	shards     map[string][]*TaskShard
	shardMutex sync.RWMutex

	// 节点管理
	nodes      map[string]*ClusterNode
	nodeMutex  sync.RWMutex

	// 任务队列
	taskQueue  chan *DistributedTask

	// 分布式锁
	locks      map[string]*DistributedLock
	lockMutex  sync.Mutex

	// Cron解析器
	cronParser *CronParser

	// 执行器
	executor   TaskExecutor

	// 监控器
	monitor    *TaskMonitor

	// 控制
	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup

	// 日志
	logger     *zap.Logger
}

// NewDistributedScheduler 创建分布式调度器
func NewDistributedScheduler(config *DistributedSchedulerConfig, executor TaskExecutor) (*DistributedScheduler, error) {
	if config == nil {
		config = DefaultDistributedSchedulerConfig()
	}

	// 创建Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	// 生成节点ID
	nodeID := config.NodeID
	if nodeID == "" {
		nodeID = generateNodeID()
	}

	s := &DistributedScheduler{
		config:     config,
		redis:      redisClient,
		nodeID:     nodeID,
		nodeState:  NodeStateJoining,
		tasks:      make(map[string]*DistributedTask),
		shards:     make(map[string][]*TaskShard),
		nodes:      make(map[string]*ClusterNode),
		locks:      make(map[string]*DistributedLock),
		cronParser: NewCronParser(),
		executor:   executor,
		logger:     logger.Named("distributed-scheduler"),
	}

	s.taskQueue = make(chan *DistributedTask, config.TaskQueueSize)
	s.ctx, s.cancelFunc = context.WithCancel(context.Background())

	return s, nil
}

// Start 启动调度器
func (s *DistributedScheduler) Start() error {
	if atomic.LoadInt32(&s.running) == 1 {
		return ErrSchedulerRunning
	}

	atomic.StoreInt32(&s.running, 1)
	atomic.StoreInt32(&s.closed, 0)

	s.logger.Info("Starting distributed scheduler",
		zap.String("nodeID", s.nodeID),
		zap.Int("maxConcurrentTasks", s.config.MaxConcurrentTasks))

	// 注册节点
	if err := s.registerNode(); err != nil {
		return fmt.Errorf("failed to register node: %w", err)
	}

	// 启动心跳
	s.wg.Add(1)
	go s.heartbeat()

	// 启动任务调度
	s.wg.Add(1)
	go s.schedule()

	// 启动任务执行器
	for i := 0; i < s.config.MaxConcurrentTasks; i++ {
		s.wg.Add(1)
		go s.executeWorker(i)
	}

	// 启动节点监控
	s.wg.Add(1)
	go s.monitorNodes()

	// 启动任务重新分配
	if s.config.EnableAutoRebalance {
		s.wg.Add(1)
		go s.rebalanceTasks()
	}

	// 启动锁续期
	s.wg.Add(1)
	go s.renewLocks()

	// 节点状态更新为活跃
	s.nodeState = NodeStateActive

	return nil
}

// Stop 停止调度器
func (s *DistributedScheduler) Stop() error {
	if atomic.LoadInt32(&s.running) == 0 {
		return ErrSchedulerNotRunning
	}

	s.logger.Info("Stopping distributed scheduler")

	// 节点状态更新为离开
	s.nodeState = NodeStateLeaving

	// 释放所有锁
	s.releaseAllLocks()

	// 注销节点
	s.unregisterNode()

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
		s.logger.Info("Distributed scheduler stopped successfully")
		return nil
	case <-time.After(30 * time.Second):
		s.logger.Warn("Distributed scheduler stop timeout")
		return errors.New("stop timeout")
	}
}

// registerNode 注册节点
func (s *DistributedScheduler) registerNode() error {
	ctx := context.Background()
	nodeKey := s.getNodeKey(s.nodeID)

	node := &ClusterNode{
		NodeID:        s.nodeID,
		Address:       s.config.Address,
		State:         NodeStateJoining,
		LastHeartbeat: time.Now(),
		TaskCount:     0,
		Load:          0,
		Capacity:      s.config.NodeCapacity,
		Labels:        make(map[string]string),
		StartTime:     time.Now(),
	}

	data, err := json.Marshal(node)
	if err != nil {
		return err
	}

	// 存储节点信息
	if err := s.redis.Set(ctx, nodeKey, data, s.config.HeartbeatTimeout*2).Err(); err != nil {
		return err
	}

	// 添加到节点集合
	if err := s.redis.SAdd(ctx, s.getNodesSetKey(), s.nodeID).Err(); err != nil {
		return err
	}

	s.logger.Info("Node registered", zap.String("nodeID", s.nodeID))
	return nil
}

// unregisterNode 注销节点
func (s *DistributedScheduler) unregisterNode() {
	ctx := context.Background()

	// 从节点集合移除
	s.redis.SRem(ctx, s.getNodesSetKey(), s.nodeID)

	// 删除节点信息
	s.redis.Del(ctx, s.getNodeKey(s.nodeID))

	s.logger.Info("Node unregistered", zap.String("nodeID", s.nodeID))
}

// heartbeat 心跳
func (s *DistributedScheduler) heartbeat() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.sendHeartbeat()

		case <-s.ctx.Done():
			return
		}
	}
}

// sendHeartbeat 发送心跳
func (s *DistributedScheduler) sendHeartbeat() {
	ctx := context.Background()
	nodeKey := s.getNodeKey(s.nodeID)

	// 获取当前任务数
	s.taskMutex.RLock()
	taskCount := len(s.tasks)
	s.taskMutex.RUnlock()

	// 更新节点信息
	node := &ClusterNode{
		NodeID:        s.nodeID,
		Address:       s.config.Address,
		State:         s.nodeState,
		LastHeartbeat: time.Now(),
		TaskCount:     taskCount,
		Load:          float64(taskCount) / float64(s.config.NodeCapacity),
		Capacity:      s.config.NodeCapacity,
		Labels:        make(map[string]string),
		StartTime:     time.Now(),
	}

	data, err := json.Marshal(node)
	if err != nil {
		s.logger.Error("Failed to marshal node info", zap.Error(err))
		return
	}

	// 更新节点信息并延长过期时间
	if err := s.redis.Set(ctx, nodeKey, data, s.config.HeartbeatTimeout*2).Err(); err != nil {
		s.logger.Error("Failed to send heartbeat", zap.Error(err))
	}
}

// monitorNodes 监控集群节点
func (s *DistributedScheduler) monitorNodes() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkNodes()

		case <-s.ctx.Done():
			return
		}
	}
}

// checkNodes 检查节点状态
func (s *DistributedScheduler) checkNodes() {
	ctx := context.Background()

	// 获取所有节点
	nodeIDs, err := s.redis.SMembers(ctx, s.getNodesSetKey()).Result()
	if err != nil {
		s.logger.Error("Failed to get nodes", zap.Error(err))
		return
	}

	s.nodeMutex.Lock()
	defer s.nodeMutex.Unlock()

	activeCount := 0
	now := time.Now()

	for _, nodeID := range nodeIDs {
		nodeKey := s.getNodeKey(nodeID)
		data, err := s.redis.Get(ctx, nodeKey).Bytes()
		if err != nil {
			// 节点不存在，从集合移除
			s.redis.SRem(ctx, s.getNodesSetKey(), nodeID)
			delete(s.nodes, nodeID)
			continue
		}

		var node ClusterNode
		if err := json.Unmarshal(data, &node); err != nil {
			continue
		}

		// 检查心跳超时
		if now.Sub(node.LastHeartbeat) > s.config.HeartbeatTimeout {
			// 节点超时，标记为不活跃
			node.State = NodeStateInactive
			s.redis.SRem(ctx, s.getNodesSetKey(), nodeID)
			delete(s.nodes, nodeID)

			s.logger.Warn("Node timeout",
				zap.String("nodeID", nodeID),
				zap.Time("lastHeartbeat", node.LastHeartbeat))
		} else {
			s.nodes[nodeID] = &node
			if node.State == NodeStateActive {
				activeCount++
			}
		}
	}

	// 更新指标
	distributedNodeCount.Set(float64(activeCount))
}

// schedule 任务调度
func (s *DistributedScheduler) schedule() {
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
func (s *DistributedScheduler) scheduleTasks() {
	now := time.Now()

	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()

	for _, task := range s.tasks {
		// 跳过非活跃任务
		if task.Status != TaskStatusPending || !task.Enabled {
			continue
		}

		// 检查是否是分配给本节点的任务
		if task.AssignedNode != "" && task.AssignedNode != s.nodeID {
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
		default:
			s.logger.Warn("Task queue is full", zap.String("taskID", task.ID))
		}
	}
}

// executeWorker 任务执行工作器
func (s *DistributedScheduler) executeWorker(workerID int) {
	defer s.wg.Done()

	for {
		select {
		case task := <-s.taskQueue:
			if task == nil {
				continue
			}
			s.executeTask(task)

		case <-s.ctx.Done():
			return
		}
	}
}

// executeTask 执行任务
func (s *DistributedScheduler) executeTask(task *DistributedTask) {
	startTime := time.Now()

	// 尝试获取分布式锁
	lockKey := s.getTaskLockKey(task.ID)
	lock, err := s.acquireLock(lockKey, s.config.LockTTL)
	if err != nil {
		s.logger.Error("Failed to acquire lock",
			zap.String("taskID", task.ID),
			zap.Error(err))
		s.handleTaskError(task, err, startTime)
		return
	}

	if lock == nil {
		s.logger.Debug("Task already running on another node",
			zap.String("taskID", task.ID))
		return
	}
	defer s.releaseLock(lock)

	// 创建任务上下文
	ctx, cancel := context.WithTimeout(s.ctx, task.Timeout)
	defer cancel()

	// 执行任务
	result, err := s.executor.Execute(ctx, &ExecutionContext{
		TaskID:      task.ID,
		ShardIndex:  task.ShardingIndex,
		TotalShards: task.ShardingCount,
		Config:      task.Config,
	})

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

	// 记录执行结果
	if s.monitor != nil {
		s.monitor.RecordExecution(task.ID, startTime, duration, err == nil, result)
	}

	if err != nil {
		s.handleTaskError(task, err, startTime)
	}

	// 更新指标
	if s.config.EnableMetrics {
		status := "success"
		if err != nil {
			status = "failed"
		}
		distributedTasksTotal.WithLabelValues(status).Inc()
	}
}

// handleTaskError 处理任务错误
func (s *DistributedScheduler) handleTaskError(task *DistributedTask, err error, startTime time.Time) {
	s.logger.Error("Task execution failed",
		zap.String("taskID", task.ID),
		zap.Error(err))

	// 重试逻辑
	if task.MaxRetry > 0 && task.RetryCount < task.MaxRetry {
		go s.retryTask(task)
	}
}

// retryTask 重试任务
func (s *DistributedScheduler) retryTask(task *DistributedTask) {
	s.taskMutex.Lock()
	task.RetryCount++
	s.taskMutex.Unlock()

	s.logger.Info("Retrying task",
		zap.String("taskID", task.ID),
		zap.Int("retryCount", task.RetryCount),
		zap.Int("maxRetry", task.MaxRetry))

	// 等待一段时间后重试
	time.Sleep(time.Duration(task.RetryCount) * time.Second)

	// 检查任务是否还存在
	s.taskMutex.RLock()
	_, exists := s.tasks[task.ID]
	s.taskMutex.RUnlock()

	if !exists {
		return
	}

	// 重新执行任务
	s.executeTask(task)
}

// calculateNextRunTime 计算下次执行时间
func (s *DistributedScheduler) calculateNextRunTime(task *DistributedTask) time.Time {
	if task.cronExpr == nil && task.CronExpr != "" {
		expr, err := s.cronParser.Parse(task.CronExpr)
		if err != nil {
			s.logger.Error("Failed to parse cron expression",
				zap.String("taskID", task.ID),
				zap.String("cronExpr", task.CronExpr),
				zap.Error(err))
			return time.Now().Add(time.Minute)
		}
		task.cronExpr = expr
	}

	if task.cronExpr != nil {
		return task.cronExpr.Next(time.Now())
	}

	return time.Now().Add(time.Minute)
}

// acquireLock 获取分布式锁
func (s *DistributedScheduler) acquireLock(key string, ttl time.Duration) (*DistributedLock, error) {
	start := time.Now()

	ctx := context.Background()
	value := generateLockValue()

	// 使用SET NX EX原子操作
	acquired, err := s.redis.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return nil, err
	}

	if !acquired {
		distributedLockWait.Observe(time.Since(start).Seconds())
		return nil, nil
	}

	lock := &DistributedLock{
		client:    s.redis,
		key:       key,
		value:     value,
		ttl:       ttl,
		renewChan: make(chan struct{}, 1),
		stopChan:  make(chan struct{}),
	}

	s.lockMutex.Lock()
	s.locks[key] = lock
	s.lockMutex.Unlock()

	distributedLockWait.Observe(time.Since(start).Seconds())
	return lock, nil
}

// releaseLock 释放锁
func (s *DistributedScheduler) releaseLock(lock *DistributedLock) {
	lock.mu.Lock()
	defer lock.mu.Unlock()

	// 停止续期
	close(lock.stopChan)

	// 使用Lua脚本确保只释放自己持有的锁
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	ctx := context.Background()
	s.redis.Eval(ctx, script, []string{lock.key}, lock.value)

	s.lockMutex.Lock()
	delete(s.locks, lock.key)
	s.lockMutex.Unlock()
}

// releaseAllLocks 释放所有锁
func (s *DistributedScheduler) releaseAllLocks() {
	s.lockMutex.Lock()
	locks := make([]*DistributedLock, 0, len(s.locks))
	for _, lock := range s.locks {
		locks = append(locks, lock)
	}
	s.lockMutex.Unlock()

	for _, lock := range locks {
		s.releaseLock(lock)
	}
}

// renewLocks 锁续期
func (s *DistributedScheduler) renewLocks() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.LockTTL / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.renewAllLocks()

		case <-s.ctx.Done():
			return
		}
	}
}

// renewAllLocks 续期所有锁
func (s *DistributedScheduler) renewAllLocks() {
	s.lockMutex.Lock()
	locks := make([]*DistributedLock, 0, len(s.locks))
	for _, lock := range s.locks {
		locks = append(locks, lock)
	}
	s.lockMutex.Unlock()

	ctx := context.Background()
	for _, lock := range locks {
		script := `
			if redis.call("get", KEYS[1]) == ARGV[1] then
				return redis.call("expire", KEYS[1], ARGV[2])
			else
				return 0
			end
		`
		s.redis.Eval(ctx, script, []string{lock.key}, lock.value, int(lock.ttl.Seconds()))
	}
}

// rebalanceTasks 任务重新分配
func (s *DistributedScheduler) rebalanceTasks() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.RebalanceInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.doRebalance()

		case <-s.ctx.Done():
			return
		}
	}
}

// doRebalance 执行任务重新分配
func (s *DistributedScheduler) doRebalance() {
	ctx := context.Background()

	// 获取分布式锁，确保只有一个节点执行重新分配
	lockKey := "nem:scheduler:rebalance:lock"
	lock, err := s.acquireLock(lockKey, 10*time.Second)
	if err != nil || lock == nil {
		return
	}
	defer s.releaseLock(lock)

	// 获取所有活跃节点
	s.nodeMutex.RLock()
	activeNodes := make([]*ClusterNode, 0)
	for _, node := range s.nodes {
		if node.State == NodeStateActive {
			activeNodes = append(activeNodes, node)
		}
	}
	s.nodeMutex.RUnlock()

	if len(activeNodes) == 0 {
		return
	}

	// 获取所有任务
	s.taskMutex.RLock()
	allTasks := make([]*DistributedTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		allTasks = append(allTasks, task)
	}
	s.taskMutex.RUnlock()

	// 按负载排序节点
	sortNodesByLoad(activeNodes)

	// 重新分配任务
	reassigned := 0
	for _, task := range allTasks {
		// 计算任务应该分配的节点
		targetNode := s.selectNodeForTask(task, activeNodes)
		if targetNode == nil {
			continue
		}

		// 如果任务已分配且节点正常，跳过
		if task.AssignedNode == targetNode.NodeID {
			continue
		}

		// 更新任务分配
		taskKey := s.getTaskKey(task.ID)
		task.AssignedNode = targetNode.NodeID
		task.UpdateTime = time.Now()

		data, err := json.Marshal(task)
		if err != nil {
			continue
		}

		if err := s.redis.Set(ctx, taskKey, data, 0).Err(); err != nil {
			continue
		}

		reassigned++
	}

	if reassigned > 0 {
		s.logger.Info("Tasks rebalanced",
			zap.Int("reassigned", reassigned),
			zap.Int("totalTasks", len(allTasks)),
			zap.Int("activeNodes", len(activeNodes)))
		distributedRebalanceTotal.WithLabelValues("success").Inc()
	}
}

// selectNodeForTask 为任务选择节点
func (s *DistributedScheduler) selectNodeForTask(task *DistributedTask, nodes []*ClusterNode) *ClusterNode {
	if len(nodes) == 0 {
		return nil
	}

	// 如果任务有分片，使用一致性哈希
	if task.ShardingKey != "" {
		return s.selectNodeByHash(task.ShardingKey, nodes)
	}

	// 否则选择负载最低的节点
	return nodes[0]
}

// selectNodeByHash 使用一致性哈希选择节点
func (s *DistributedScheduler) selectNodeByHash(key string, nodes []*ClusterNode) *ClusterNode {
	if len(nodes) == 0 {
		return nil
	}

	// 计算哈希值
	hash := sha256.Sum256([]byte(key))
	hashInt := 0
	for _, b := range hash[:4] {
		hashInt = hashInt*256 + int(b)
	}

	// 选择节点
	index := hashInt % len(nodes)
	return nodes[index]
}

// AddTask 添加任务
func (s *DistributedScheduler) AddTask(task *DistributedTask) error {
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

	// 解析Cron表达式
	if task.CronExpr != "" {
		expr, err := s.cronParser.Parse(task.CronExpr)
		if err != nil {
			return fmt.Errorf("invalid cron expression: %w", err)
		}
		task.cronExpr = expr
	}

	// 计算下次执行时间
	if task.Enabled {
		task.NextRunTime = s.calculateNextRunTime(task)
	}

	// 分配节点
	s.nodeMutex.RLock()
	activeNodes := make([]*ClusterNode, 0)
	for _, node := range s.nodes {
		if node.State == NodeStateActive {
			activeNodes = append(activeNodes, node)
		}
	}
	s.nodeMutex.RUnlock()

	if len(activeNodes) > 0 {
		targetNode := s.selectNodeForTask(task, activeNodes)
		if targetNode != nil {
			task.AssignedNode = targetNode.NodeID
		}
	}

	// 保存到本地
	s.tasks[task.ID] = task

	// 保存到Redis
	ctx := context.Background()
	taskKey := s.getTaskKey(task.ID)
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	if err := s.redis.Set(ctx, taskKey, data, 0).Err(); err != nil {
		delete(s.tasks, task.ID)
		return err
	}

	// 添加到任务集合
	s.redis.SAdd(ctx, s.getTasksSetKey(), task.ID)

	s.logger.Info("Task added",
		zap.String("taskID", task.ID),
		zap.String("name", task.Name),
		zap.String("assignedNode", task.AssignedNode))

	return nil
}

// RemoveTask 移除任务
func (s *DistributedScheduler) RemoveTask(taskID string) error {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return ErrTaskNotFound
	}

	task.Status = TaskStatusCancelled
	delete(s.tasks, taskID)

	// 从Redis删除
	ctx := context.Background()
	s.redis.Del(ctx, s.getTaskKey(taskID))
	s.redis.SRem(ctx, s.getTasksSetKey(), taskID)

	s.logger.Info("Task removed", zap.String("taskID", taskID))
	return nil
}

// GetTask 获取任务
func (s *DistributedScheduler) GetTask(taskID string) (*DistributedTask, error) {
	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, ErrTaskNotFound
	}

	return task, nil
}

// GetAllTasks 获取所有任务
func (s *DistributedScheduler) GetAllTasks() []*DistributedTask {
	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()

	tasks := make([]*DistributedTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// GetNodes 获取所有节点
func (s *DistributedScheduler) GetNodes() []*ClusterNode {
	s.nodeMutex.RLock()
	defer s.nodeMutex.RUnlock()

	nodes := make([]*ClusterNode, 0, len(s.nodes))
	for _, node := range s.nodes {
		nodes = append(nodes, node)
	}

	return nodes
}

// GetNodeID 获取当前节点ID
func (s *DistributedScheduler) GetNodeID() string {
	return s.nodeID
}

// IsRunning 检查是否运行中
func (s *DistributedScheduler) IsRunning() bool {
	return atomic.LoadInt32(&s.running) == 1
}

// GetTaskCount 获取任务数量
func (s *DistributedScheduler) GetTaskCount() int {
	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()
	return len(s.tasks)
}

// Redis键名辅助方法
func (s *DistributedScheduler) getNodeKey(nodeID string) string {
	return fmt.Sprintf("nem:scheduler:node:%s", nodeID)
}

func (s *DistributedScheduler) getNodesSetKey() string {
	return "nem:scheduler:nodes"
}

func (s *DistributedScheduler) getTaskKey(taskID string) string {
	return fmt.Sprintf("nem:scheduler:task:%s", taskID)
}

func (s *DistributedScheduler) getTasksSetKey() string {
	return "nem:scheduler:tasks"
}

func (s *DistributedScheduler) getTaskLockKey(taskID string) string {
	return fmt.Sprintf("nem:scheduler:lock:task:%s", taskID)
}

// 辅助函数
func generateNodeID() string {
	timestamp := time.Now().UnixNano()
	random := rand.Int63()
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d-%d", timestamp, random)))
	return hex.EncodeToString(hash[:8])
}

func generateLockValue() string {
	timestamp := time.Now().UnixNano()
	random := rand.Int63()
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d-%d-%d", timestamp, random, time.Now().Nanosecond())))
	return hex.EncodeToString(hash[:16])
}

func sortNodesByLoad(nodes []*ClusterNode) {
	// 简单冒泡排序
	for i := 0; i < len(nodes)-1; i++ {
		for j := i + 1; j < len(nodes); j++ {
			if nodes[i].Load > nodes[j].Load {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}
}
