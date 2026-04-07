package nacos

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"go.uber.org/zap"
)

// HealthChecker 健康检查器
type HealthChecker struct {
	registry       *Registry
	options        *Options
	heartbeatTasks map[string]*HeartbeatTask
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	logger         *zap.Logger
}

// HeartbeatTask 心跳任务
type HeartbeatTask struct {
	ServiceName string
	Instance    *ServiceInstance
	Interval    time.Duration
	ticker      *time.Ticker
	cancel      context.CancelFunc
}

// HealthStatus 健康状态
type HealthStatus struct {
	ServiceName string
	Instance    *ServiceInstance
	Healthy     bool
	LastCheck   time.Time
	Error       error
}

// NewHealthChecker 创建新的健康检查器
func NewHealthChecker(registry *Registry, opts ...Option) (*HealthChecker, error) {
	if registry == nil {
		return nil, fmt.Errorf("registry cannot be nil")
	}

	options := ApplyOptions(opts...)
	ctx, cancel := context.WithCancel(context.Background())

	return &HealthChecker{
		registry:       registry,
		options:        options,
		heartbeatTasks: make(map[string]*HeartbeatTask),
		ctx:            ctx,
		cancel:         cancel,
	}, nil
}

// StartHeartbeat 启动心跳上报
func (h *HealthChecker) StartHeartbeat(instance *ServiceInstance) error {
	if instance == nil {
		return fmt.Errorf("instance cannot be nil")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.heartbeatTasks[instance.ServiceName]; exists {
		return fmt.Errorf("heartbeat task already exists for service %s", instance.ServiceName)
	}

	ctx, cancel := context.WithCancel(h.ctx)
	interval := h.options.HeartbeatInterval
	if interval == 0 {
		interval = 5 * time.Second
	}

	task := &HeartbeatTask{
		ServiceName: instance.ServiceName,
		Instance:    instance,
		Interval:    interval,
		ticker:      time.NewTicker(interval),
		cancel:      cancel,
	}

	h.heartbeatTasks[instance.ServiceName] = task

	// 启动心跳协程
	go h.runHeartbeat(ctx, task)

	return nil
}

// StopHeartbeat 停止心跳上报
func (h *HealthChecker) StopHeartbeat(serviceName string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	task, exists := h.heartbeatTasks[serviceName]
	if !exists {
		return fmt.Errorf("heartbeat task not found for service %s", serviceName)
	}

	if task.ticker != nil {
		task.ticker.Stop()
	}
	if task.cancel != nil {
		task.cancel()
	}

	delete(h.heartbeatTasks, serviceName)

	return nil
}

// runHeartbeat 运行心跳上报
func (h *HealthChecker) runHeartbeat(ctx context.Context, task *HeartbeatTask) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-task.ticker.C:
			h.sendHeartbeat(task)
		}
	}
}

// sendHeartbeat 发送心跳
func (h *HealthChecker) sendHeartbeat(task *HeartbeatTask) {
	// 使用Nacos SDK的心跳机制
	// Nacos SDK会自动处理临时实例的心跳
	// 这里我们只需要确保实例仍然注册即可
	h.registry.mu.RLock()
	_, exists := h.registry.instances[task.ServiceName]
	h.registry.mu.RUnlock()

	if !exists {
		// 实例已被注销，停止心跳
		_ = h.StopHeartbeat(task.ServiceName)
		return
	}

	// 对于持久化实例，需要手动发送心跳
	// 注意：nacos-sdk-go v2 已移除 SendHeartbeat 方法
	// 持久化实例的心跳由 SDK 内部自动处理
	if !task.Instance.Ephemeral {
		// SDK v2 自动处理持久化实例心跳
		h.logger.Debug("persistent instance heartbeat handled by SDK", 
			zap.String("service", task.ServiceName))
	}
}

// CheckInstance 检查实例健康状态
func (h *HealthChecker) CheckInstance(serviceName string) (*HealthStatus, error) {
	h.registry.mu.RLock()
	instance, exists := h.registry.instances[serviceName]
	h.registry.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	// 查询实例状态
	instances, err := h.registry.Discover(serviceName, WithDiscoveryHealthyOnly(false))
	if err != nil {
		return &HealthStatus{
			ServiceName: serviceName,
			Instance:    instance,
			Healthy:     false,
			LastCheck:   time.Now(),
			Error:       err,
		}, nil
	}

	// 查找当前实例
	for _, inst := range instances {
		if inst.Ip == instance.Ip && inst.Port == instance.Port {
			return &HealthStatus{
				ServiceName: serviceName,
				Instance:    inst,
				Healthy:     inst.Healthy,
				LastCheck:   time.Now(),
				Error:       nil,
			}, nil
		}
	}

	// 实例不在服务列表中
	return &HealthStatus{
		ServiceName: serviceName,
		Instance:    instance,
		Healthy:     false,
		LastCheck:   time.Now(),
		Error:       fmt.Errorf("instance not found in service list"),
	}, nil
}

// CheckAllInstances 检查所有实例健康状态
func (h *HealthChecker) CheckAllInstances() ([]*HealthStatus, error) {
	h.registry.mu.RLock()
	serviceNames := make([]string, 0, len(h.registry.instances))
	for name := range h.registry.instances {
		serviceNames = append(serviceNames, name)
	}
	h.registry.mu.RUnlock()

	statuses := make([]*HealthStatus, 0, len(serviceNames))
	for _, name := range serviceNames {
		status, err := h.CheckInstance(name)
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// UpdateInstanceStatus 更新实例状态
func (h *HealthChecker) UpdateInstanceStatus(serviceName string, healthy bool) error {
	h.registry.mu.RLock()
	instance, exists := h.registry.instances[serviceName]
	h.registry.mu.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found", serviceName)
	}

	// 更新实例健康状态
	instance.Healthy = healthy

	// 重新注册实例以更新状态
	success, err := h.registry.namingCli.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          instance.Ip,
		Port:        instance.Port,
		ServiceName: instance.ServiceName,
		Weight:      instance.Weight,
		Enable:      instance.Enable,
		Healthy:     instance.Healthy,
		Metadata:    instance.Metadata,
		ClusterName: instance.ClusterName,
		GroupName:   instance.GroupName,
		Ephemeral:   instance.Ephemeral,
	})

	if err != nil {
		return fmt.Errorf("failed to update instance status: %w", err)
	}

	if !success {
		return fmt.Errorf("failed to update instance status: operation returned false")
	}

	return nil
}

// SetInstanceWeight 设置实例权重
func (h *HealthChecker) SetInstanceWeight(serviceName string, weight float64) error {
	h.registry.mu.RLock()
	instance, exists := h.registry.instances[serviceName]
	h.registry.mu.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found", serviceName)
	}

	// 更新实例权重
	instance.Weight = weight

	// 更新实例
	success, err := h.registry.namingCli.UpdateInstance(vo.UpdateInstanceParam{
		Ip:          instance.Ip,
		Port:        instance.Port,
		ServiceName: instance.ServiceName,
		Weight:      weight,
		Enable:      instance.Enable,
		Metadata:    instance.Metadata,
		ClusterName: instance.ClusterName,
		GroupName:   instance.GroupName,
		Ephemeral:   instance.Ephemeral,
	})

	if err != nil {
		return fmt.Errorf("failed to update instance weight: %w", err)
	}

	if !success {
		return fmt.Errorf("failed to update instance weight: operation returned false")
	}

	return nil
}

// SetInstanceMetadata 设置实例元数据
func (h *HealthChecker) SetInstanceMetadata(serviceName string, metadata map[string]string) error {
	h.registry.mu.RLock()
	instance, exists := h.registry.instances[serviceName]
	h.registry.mu.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found", serviceName)
	}

	// 合并元数据
	if instance.Metadata == nil {
		instance.Metadata = make(map[string]string)
	}
	for k, v := range metadata {
		instance.Metadata[k] = v
	}

	// 更新实例
	success, err := h.registry.namingCli.UpdateInstance(vo.UpdateInstanceParam{
		Ip:          instance.Ip,
		Port:        instance.Port,
		ServiceName: instance.ServiceName,
		Weight:      instance.Weight,
		Enable:      instance.Enable,
		Metadata:    instance.Metadata,
		ClusterName: instance.ClusterName,
		GroupName:   instance.GroupName,
		Ephemeral:   instance.Ephemeral,
	})

	if err != nil {
		return fmt.Errorf("failed to update instance metadata: %w", err)
	}

	if !success {
		return fmt.Errorf("failed to update instance metadata: operation returned false")
	}

	return nil
}

// Beat 心跳接口（用于自定义心跳逻辑）
func (h *HealthChecker) Beat(serviceName string) error {
	h.mu.RLock()
	task, exists := h.heartbeatTasks[serviceName]
	h.mu.RUnlock()

	if !exists {
		return fmt.Errorf("heartbeat task not found for service %s", serviceName)
	}

	// 手动触发一次心跳
	h.sendHeartbeat(task)

	return nil
}

// GetHeartbeatTasks 获取所有心跳任务
func (h *HealthChecker) GetHeartbeatTasks() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	tasks := make([]string, 0, len(h.heartbeatTasks))
	for name := range h.heartbeatTasks {
		tasks = append(tasks, name)
	}

	return tasks
}

// Close 关闭健康检查器
func (h *HealthChecker) Close() error {
	h.cancel()

	// 停止所有心跳任务
	h.mu.Lock()
	tasks := make([]*HeartbeatTask, 0, len(h.heartbeatTasks))
	for _, task := range h.heartbeatTasks {
		tasks = append(tasks, task)
	}
	h.mu.Unlock()

	for _, task := range tasks {
		_ = h.StopHeartbeat(task.ServiceName)
	}

	return nil
}

// HealthCheckOption 健康检查选项函数
type HealthCheckOption func(*HealthCheckOptions)

// HealthCheckOptions 健康检查选项
type HealthCheckOptions struct {
	Interval    time.Duration
	Timeout     time.Duration
	RetryCount  int
	RetryDelay  time.Duration
}

// WithHealthCheckInterval 设置健康检查间隔
func WithHealthCheckInterval(interval time.Duration) HealthCheckOption {
	return func(o *HealthCheckOptions) {
		o.Interval = interval
	}
}

// WithHealthCheckTimeout 设置健康检查超时
func WithHealthCheckTimeout(timeout time.Duration) HealthCheckOption {
	return func(o *HealthCheckOptions) {
		o.Timeout = timeout
	}
}

// WithHealthCheckRetry 设置健康检查重试
func WithHealthCheckRetry(count int, delay time.Duration) HealthCheckOption {
	return func(o *HealthCheckOptions) {
		o.RetryCount = count
		o.RetryDelay = delay
	}
}
