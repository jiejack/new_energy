package notifier

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrSchedulerNotRunning   = errors.New("scheduler is not running")
	ErrNotificationNotFound  = errors.New("notification not found")
	ErrMaxRetriesExceeded    = errors.New("max retries exceeded")
	ErrSilencePeriodActive   = errors.New("silence period is active")
)

// NotificationScheduler 通知调度器
type NotificationScheduler struct {
	notifiers      map[NotificationChannel]Notifier
	store          NotificationStore
	logger         NotificationLogger
	templateMgr    TemplateManager
	silenceChecker SilenceChecker

	// 配置
	config         *SchedulerConfig

	// 队列
	priorityQueues map[NotificationPriority]chan *Notification
	queueSize      int

	// 控制
	running        bool
	stopChan       chan struct{}
	workerCount    int

	mu             sync.RWMutex
	wg             sync.WaitGroup
}

// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	// 队列配置
	QueueSize      int `json:"queue_size"`       // 每个优先级队列大小
	WorkerCount    int `json:"worker_count"`     // 工作协程数

	// 重试配置
	MaxRetries     int           `json:"max_retries"`     // 最大重试次数
	RetryDelay     time.Duration `json:"retry_delay"`     // 重试延迟
	RetryBackoff   float64       `json:"retry_backoff"`   // 重试退避因子

	// 超时配置
	SendTimeout    time.Duration `json:"send_timeout"`    // 发送超时

	// 批处理配置
	BatchSize      int           `json:"batch_size"`      // 批处理大小
	BatchTimeout   time.Duration `json:"batch_timeout"`   // 批处理超时

	// 限流配置
	GlobalRateLimit int          `json:"global_rate_limit"` // 全局限流
}

// DefaultSchedulerConfig 默认调度器配置
func DefaultSchedulerConfig() *SchedulerConfig {
	return &SchedulerConfig{
		QueueSize:       1000,
		WorkerCount:     10,
		MaxRetries:      3,
		RetryDelay:      5 * time.Second,
		RetryBackoff:    2.0,
		SendTimeout:     30 * time.Second,
		BatchSize:       10,
		BatchTimeout:    5 * time.Second,
		GlobalRateLimit: 100,
	}
}

// NewNotificationScheduler 创建通知调度器
func NewNotificationScheduler(
	store NotificationStore,
	logger NotificationLogger,
	templateMgr TemplateManager,
	silenceChecker SilenceChecker,
	config *SchedulerConfig,
) *NotificationScheduler {
	if config == nil {
		config = DefaultSchedulerConfig()
	}

	s := &NotificationScheduler{
		notifiers:      make(map[NotificationChannel]Notifier),
		store:          store,
		logger:         logger,
		templateMgr:    templateMgr,
		silenceChecker: silenceChecker,
		config:         config,
		priorityQueues: make(map[NotificationPriority]chan *Notification),
		queueSize:      config.QueueSize,
		stopChan:       make(chan struct{}),
		workerCount:    config.WorkerCount,
	}

	// 初始化优先级队列
	for priority := PriorityLow; priority <= PriorityCritical; priority++ {
		s.priorityQueues[priority] = make(chan *Notification, config.QueueSize)
	}

	return s
}

// RegisterNotifier 注册通知器
func (s *NotificationScheduler) RegisterNotifier(channel NotificationChannel, notifier Notifier) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notifiers[channel] = notifier
}

// UnregisterNotifier 注销通知器
func (s *NotificationScheduler) UnregisterNotifier(channel NotificationChannel) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.notifiers, channel)
}

// Start 启动调度器
func (s *NotificationScheduler) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.mu.Unlock()

	// 启动工作协程
	for i := 0; i < s.workerCount; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}

	// 启动重试协程
	s.wg.Add(1)
	go s.retryWorker()

	// 启动统计协程
	s.wg.Add(1)
	go s.statsWorker()

	return nil
}

// Stop 停止调度器
func (s *NotificationScheduler) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopChan)
	s.wg.Wait()

	// 关闭所有通知器
	for _, notifier := range s.notifiers {
		notifier.Close()
	}

	return nil
}

// IsRunning 是否运行中
func (s *NotificationScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// Schedule 调度通知
func (s *NotificationScheduler) Schedule(ctx context.Context, notification *Notification) error {
	if !s.IsRunning() {
		return ErrSchedulerNotRunning
	}

	// 检查静默期
	if s.silenceChecker != nil && s.silenceChecker.IsSilent(notification.AlarmID) {
		return ErrSilencePeriodActive
	}

	// 设置默认值
	if notification.Status == "" {
		notification.Status = StatusPending
	}
	if notification.MaxRetries == 0 {
		notification.MaxRetries = s.config.MaxRetries
	}
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = time.Now()
	}

	// 保存通知
	if err := s.store.Save(ctx, notification); err != nil {
		return fmt.Errorf("failed to save notification: %w", err)
	}

	// 根据优先级放入队列
	queue, ok := s.priorityQueues[notification.Priority]
	if !ok {
		queue = s.priorityQueues[PriorityNormal]
	}

	select {
	case queue <- notification:
		return nil
	default:
		return errors.New("notification queue is full")
	}
}

// ScheduleBatch 批量调度通知
func (s *NotificationScheduler) ScheduleBatch(ctx context.Context, notifications []*Notification) error {
	for _, notification := range notifications {
		if err := s.Schedule(ctx, notification); err != nil {
			return err
		}
	}
	return nil
}

// Cancel 取消通知
func (s *NotificationScheduler) Cancel(ctx context.Context, notificationID string) error {
	notification, err := s.store.Get(ctx, notificationID)
	if err != nil {
		return ErrNotificationNotFound
	}

	notification.Status = StatusCancelled
	return s.store.Update(ctx, notification)
}

// worker 工作协程
func (s *NotificationScheduler) worker(id int) {
	defer s.wg.Done()

	for {
		select {
		case <-s.stopChan:
			return
		default:
			// 从优先级队列获取通知（优先级从高到低）
			notification := s.getNextNotification()
			if notification == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// 处理通知
			s.processNotification(notification)
		}
	}
}

// getNextNotification 获取下一个通知
func (s *NotificationScheduler) getNextNotification() *Notification {
	// 按优先级从高到低检查队列
	priorities := []NotificationPriority{
		PriorityCritical,
		PriorityHigh,
		PriorityNormal,
		PriorityLow,
	}

	for _, priority := range priorities {
		select {
		case notification := <-s.priorityQueues[priority]:
			return notification
		default:
			continue
		}
	}

	return nil
}

// processNotification 处理通知
func (s *NotificationScheduler) processNotification(notification *Notification) {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.SendTimeout)
	defer cancel()

	// 更新状态为发送中
	notification.Status = StatusSending
	if err := s.store.Update(ctx, notification); err != nil {
		// 记录错误但继续
	}

	// 获取通知器
	s.mu.RLock()
	notifier, ok := s.notifiers[notification.Channel]
	s.mu.RUnlock()

	if !ok {
		s.handleSendError(notification, fmt.Errorf("no notifier for channel: %s", notification.Channel))
		return
	}

	// 发送通知
	result, err := notifier.Send(ctx, notification)
	if err != nil {
		s.handleSendError(notification, err)
		return
	}

	// 处理结果
	s.handleSendResult(notification, result)
}

// handleSendError 处理发送错误
func (s *NotificationScheduler) handleSendError(notification *Notification, err error) {
	notification.RetryCount++
	notification.ErrorMessage = err.Error()

	// 检查是否超过最大重试次数
	if notification.RetryCount >= notification.MaxRetries {
		notification.Status = StatusFailed
		s.store.Update(context.Background(), notification)

		// 记录日志
		if s.logger != nil {
			s.logger.Log(context.Background(), notification, &NotificationResult{
				NotificationID: notification.ID,
				Success:        false,
				Status:         StatusFailed,
				Message:        err.Error(),
				Error:          err,
			})
		}
		return
	}

	// 计算下次重试时间
	delay := s.calculateRetryDelay(notification.RetryCount)
	nextRetry := time.Now().Add(delay)
	notification.NextRetryAt = &nextRetry
	notification.Status = StatusPending

	s.store.Update(context.Background(), notification)
}

// handleSendResult 处理发送结果
func (s *NotificationScheduler) handleSendResult(notification *Notification, result *NotificationResult) {
	if result.Success {
		notification.Status = StatusSent
		notification.SentAt = result.DeliveredAt
		notification.DeliveredAt = result.DeliveredAt
	} else {
		notification.Status = StatusFailed
		notification.ErrorMessage = result.Message
	}

	s.store.Update(context.Background(), notification)

	// 记录日志
	if s.logger != nil {
		s.logger.Log(context.Background(), notification, result)
	}
}

// calculateRetryDelay 计算重试延迟
func (s *NotificationScheduler) calculateRetryDelay(retryCount int) time.Duration {
	delay := float64(s.config.RetryDelay)
	for i := 1; i < retryCount; i++ {
		delay *= s.config.RetryBackoff
	}
	return time.Duration(delay)
}

// retryWorker 重试工作协程
func (s *NotificationScheduler) retryWorker() {
	defer s.wg.Done()

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.processRetries()
		}
	}
}

// processRetries 处理重试
func (s *NotificationScheduler) processRetries() {
	ctx := context.Background()

	// 获取待重试的通知
	notifications, err := s.store.GetByStatus(ctx, StatusPending, 100)
	if err != nil {
		return
	}

	now := time.Now()
	for _, notification := range notifications {
		// 检查是否到达重试时间
		if notification.NextRetryAt != nil && notification.NextRetryAt.After(now) {
			continue
		}

		// 检查是否超过最大重试次数
		if notification.RetryCount >= notification.MaxRetries {
			notification.Status = StatusFailed
			s.store.Update(ctx, notification)
			continue
		}

		// 重新调度
		queue, ok := s.priorityQueues[notification.Priority]
		if !ok {
			queue = s.priorityQueues[PriorityNormal]
		}

		select {
		case queue <- notification:
		default:
			// 队列满，等待下次
		}
	}
}

// statsWorker 统计工作协程
func (s *NotificationScheduler) statsWorker() {
	defer s.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.collectStats()
		}
	}
}

// collectStats 收集统计信息
func (s *NotificationScheduler) collectStats() {
	// 收集队列长度
	for priority, queue := range s.priorityQueues {
		_ = priority
		_ = len(queue)
		// 可以发送到监控系统
	}
}

// GetQueueStats 获取队列统计
func (s *NotificationScheduler) GetQueueStats() map[NotificationPriority]int {
	stats := make(map[NotificationPriority]int)
	for priority, queue := range s.priorityQueues {
		stats[priority] = len(queue)
	}
	return stats
}

// MemoryNotificationStore 内存通知存储
type MemoryNotificationStore struct {
	notifications map[string]*Notification
	mu            sync.RWMutex
}

// NewMemoryNotificationStore 创建内存通知存储
func NewMemoryNotificationStore() *MemoryNotificationStore {
	return &MemoryNotificationStore{
		notifications: make(map[string]*Notification),
	}
}

// Save 保存通知
func (s *MemoryNotificationStore) Save(ctx context.Context, notification *Notification) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notifications[notification.ID] = notification
	return nil
}

// Update 更新通知
func (s *MemoryNotificationStore) Update(ctx context.Context, notification *Notification) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notifications[notification.ID] = notification
	return nil
}

// Get 获取通知
func (s *MemoryNotificationStore) Get(ctx context.Context, id string) (*Notification, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	notification, ok := s.notifications[id]
	if !ok {
		return nil, ErrNotificationNotFound
	}
	return notification, nil
}

// GetByAlarmID 根据告警ID获取通知列表
func (s *MemoryNotificationStore) GetByAlarmID(ctx context.Context, alarmID string) ([]*Notification, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Notification, 0)
	for _, notification := range s.notifications {
		if notification.AlarmID == alarmID {
			result = append(result, notification)
		}
	}
	return result, nil
}

// GetPending 获取待发送的通知
func (s *MemoryNotificationStore) GetPending(ctx context.Context, limit int) ([]*Notification, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Notification, 0)
	for _, notification := range s.notifications {
		if notification.Status == StatusPending {
			result = append(result, notification)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

// GetByStatus 根据状态获取通知列表
func (s *MemoryNotificationStore) GetByStatus(ctx context.Context, status NotificationStatus, limit int) ([]*Notification, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Notification, 0)
	for _, notification := range s.notifications {
		if notification.Status == status {
			result = append(result, notification)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

// Delete 删除通知
func (s *MemoryNotificationStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.notifications, id)
	return nil
}

// MemorySilenceChecker 内存静默期检查器
type MemorySilenceChecker struct {
	silences map[string]*silence
	mu       sync.RWMutex
}

type silence struct {
	startTime time.Time
	duration  time.Duration
}

// NewMemorySilenceChecker 创建内存静默期检查器
func NewMemorySilenceChecker() *MemorySilenceChecker {
	return &MemorySilenceChecker{
		silences: make(map[string]*silence),
	}
}

// IsSilent 是否处于静默期
func (c *MemorySilenceChecker) IsSilent(alarmID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	s, ok := c.silences[alarmID]
	if !ok {
		return false
	}

	return time.Since(s.startTime) < s.duration
}

// StartSilence 开始静默期
func (c *MemorySilenceChecker) StartSilence(alarmID string, duration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.silences[alarmID] = &silence{
		startTime: time.Now(),
		duration:  duration,
	}
	return nil
}

// EndSilence 结束静默期
func (c *MemorySilenceChecker) EndSilence(alarmID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.silences, alarmID)
	return nil
}

// NotificationBuilder 通知构建器
type NotificationBuilder struct {
	notification *Notification
}

// NewNotificationBuilder 创建通知构建器
func NewNotificationBuilder() *NotificationBuilder {
	return &NotificationBuilder{
		notification: &Notification{
			Recipients:   make([]Recipient, 0),
			Tags:         make(map[string]string),
			TemplateData: make(map[string]interface{}),
			Attachments:  make([]Attachment, 0),
			CreatedAt:    time.Now(),
		},
	}
}

// WithID 设置ID
func (b *NotificationBuilder) WithID(id string) *NotificationBuilder {
	b.notification.ID = id
	return b
}

// WithAlarmID 设置告警ID
func (b *NotificationBuilder) WithAlarmID(alarmID string) *NotificationBuilder {
	b.notification.AlarmID = alarmID
	return b
}

// WithChannel 设置渠道
func (b *NotificationBuilder) WithChannel(channel NotificationChannel) *NotificationBuilder {
	b.notification.Channel = channel
	return b
}

// WithPriority 设置优先级
func (b *NotificationBuilder) WithPriority(priority NotificationPriority) *NotificationBuilder {
	b.notification.Priority = priority
	return b
}

// WithSubject 设置主题
func (b *NotificationBuilder) WithSubject(subject string) *NotificationBuilder {
	b.notification.Subject = subject
	return b
}

// WithContent 设置内容
func (b *NotificationBuilder) WithContent(content string) *NotificationBuilder {
	b.notification.Content = content
	return b
}

// WithHTMLContent 设置HTML内容
func (b *NotificationBuilder) WithHTMLContent(htmlContent string) *NotificationBuilder {
	b.notification.HTMLContent = htmlContent
	return b
}

// WithTemplate 设置模板
func (b *NotificationBuilder) WithTemplate(templateID string, data map[string]interface{}) *NotificationBuilder {
	b.notification.TemplateID = templateID
	b.notification.TemplateData = data
	return b
}

// AddRecipient 添加接收者
func (b *NotificationBuilder) AddRecipient(recipient Recipient) *NotificationBuilder {
	b.notification.Recipients = append(b.notification.Recipients, recipient)
	return b
}

// AddTag 添加标签
func (b *NotificationBuilder) AddTag(key, value string) *NotificationBuilder {
	b.notification.Tags[key] = value
	return b
}

// AddAttachment 添加附件
func (b *NotificationBuilder) AddAttachment(attachment Attachment) *NotificationBuilder {
	b.notification.Attachments = append(b.notification.Attachments, attachment)
	return b
}

// Build 构建通知
func (b *NotificationBuilder) Build() *Notification {
	return b.notification
}
