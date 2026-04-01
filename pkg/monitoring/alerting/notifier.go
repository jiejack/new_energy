package alerting

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"sync"
	"time"
)

// AlertNotifier 告警通知器
type AlertNotifier struct {
	mu               sync.RWMutex
	channels         map[string]NotificationChannel
	templateManager  *NotificationTemplateManager
	escalationEngine *EscalationEngine
	silenceChecker   *SilenceChecker
	rateLimiter      *NotificationRateLimiter
	notificationStore NotificationStore
}

// NotificationChannel 通知渠道
type NotificationChannel interface {
	// Name 渠道名称
	Name() string

	// Send 发送通知
	Send(ctx context.Context, notification *Notification) (*SendResult, error)

	// SendBatch 批量发送
	SendBatch(ctx context.Context, notifications []*Notification) ([]*SendResult, error)

	// HealthCheck 健康检查
	HealthCheck(ctx context.Context) error

	// Close 关闭渠道
	Close() error
}

// Notification 通知消息
type Notification struct {
	ID           string                 `json:"id"`
	AlertID      string                 `json:"alert_id"`
	Channel      string                 `json:"channel"`
	Priority     NotificationPriority  `json:"priority"`
	Status       NotificationStatus     `json:"status"`

	// 接收者
	Recipients   []Recipient            `json:"recipients"`

	// 内容
	Subject      string                 `json:"subject"`
	Content      string                 `json:"content"`
	HTMLContent  string                 `json:"html_content,omitempty"`
	TemplateID   string                 `json:"template_id,omitempty"`
	TemplateData map[string]interface{} `json:"template_data,omitempty"`

	// 附加信息
	Tags         map[string]string      `json:"tags,omitempty"`
	Attachments  []Attachment           `json:"attachments,omitempty"`

	// 时间信息
	CreatedAt    time.Time              `json:"created_at"`
	SentAt       *time.Time             `json:"sent_at,omitempty"`
	DeliveredAt  *time.Time             `json:"delivered_at,omitempty"`

	// 重试信息
	RetryCount   int                    `json:"retry_count"`
	MaxRetries   int                    `json:"max_retries"`
	NextRetryAt  *time.Time             `json:"next_retry_at,omitempty"`

	// 错误信息
	ErrorMessage string                 `json:"error_message,omitempty"`
}

// Recipient 接收者
type Recipient struct {
	UserID     string `json:"user_id,omitempty"`
	Name       string `json:"name,omitempty"`
	Phone      string `json:"phone,omitempty"`
	Email      string `json:"email,omitempty"`
	WechatID   string `json:"wechat_id,omitempty"`
	DingTalkID string `json:"dingtalk_id,omitempty"`
}

// Attachment 附件
type Attachment struct {
	Name     string `json:"name"`
	Content  []byte `json:"content"`
	MimeType string `json:"mime_type"`
}

// NotificationPriority 通知优先级
type NotificationPriority string

const (
	PriorityLow      NotificationPriority = "low"
	PriorityNormal   NotificationPriority = "normal"
	PriorityHigh     NotificationPriority = "high"
	PriorityCritical NotificationPriority = "critical"
)

// NotificationStatus 通知状态
type NotificationStatus string

const (
	StatusPending   NotificationStatus = "pending"
	StatusSending   NotificationStatus = "sending"
	StatusSent      NotificationStatus = "sent"
	StatusFailed    NotificationStatus = "failed"
	StatusCancelled NotificationStatus = "cancelled"
)

// SendResult 发送结果
type SendResult struct {
	NotificationID string    `json:"notification_id"`
	Success        bool      `json:"success"`
	Status         NotificationStatus `json:"status"`
	Message        string    `json:"message,omitempty"`
	ExternalID     string    `json:"external_id,omitempty"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
	Error          error     `json:"error,omitempty"`
}

// NotificationStore 通知存储接口
type NotificationStore interface {
	Save(ctx context.Context, notification *Notification) error
	Update(ctx context.Context, notification *Notification) error
	Get(ctx context.Context, id string) (*Notification, error)
	GetPending(ctx context.Context, limit int) ([]*Notification, error)
	GetByAlertID(ctx context.Context, alertID string) ([]*Notification, error)
}

// NewAlertNotifier 创建告警通知器
func NewAlertNotifier(store NotificationStore) *AlertNotifier {
	return &AlertNotifier{
		channels:         make(map[string]NotificationChannel),
		templateManager:  NewNotificationTemplateManager(),
		escalationEngine: NewEscalationEngine(),
		silenceChecker:   NewSilenceChecker(),
		rateLimiter:      NewNotificationRateLimiter(),
		notificationStore: store,
	}
}

// RegisterChannel 注册通知渠道
func (an *AlertNotifier) RegisterChannel(channel NotificationChannel) {
	an.mu.Lock()
	defer an.mu.Unlock()
	an.channels[channel.Name()] = channel
}

// UnregisterChannel 注销通知渠道
func (an *AlertNotifier) UnregisterChannel(name string) {
	an.mu.Lock()
	defer an.mu.Unlock()
	if channel, exists := an.channels[name]; exists {
		_ = channel.Close()
		delete(an.channels, name)
	}
}

// SendNotification 发送通知
func (an *AlertNotifier) SendNotification(ctx context.Context, notification *Notification) (*SendResult, error) {
	// 检查静默期
	if an.silenceChecker.IsSilent(notification.AlertID) {
		return &SendResult{
			NotificationID: notification.ID,
			Success:        false,
			Status:         StatusCancelled,
			Message:        "notification is silenced",
		}, nil
	}

	// 检查限流
	if !an.rateLimiter.Allow(notification.Channel) {
		return &SendResult{
			NotificationID: notification.ID,
			Success:        false,
			Status:         StatusCancelled,
			Message:        "rate limit exceeded",
		}, nil
	}

	// 应用模板
	if notification.TemplateID != "" {
		if err := an.applyTemplate(notification); err != nil {
			return nil, fmt.Errorf("failed to apply template: %w", err)
		}
	}

	// 获取渠道
	an.mu.RLock()
	channel, exists := an.channels[notification.Channel]
	an.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("channel not found: %s", notification.Channel)
	}

	// 更新状态
	notification.Status = StatusSending
	if an.notificationStore != nil {
		_ = an.notificationStore.Update(ctx, notification)
	}

	// 发送通知
	result, err := channel.Send(ctx, notification)
	if err != nil {
		notification.Status = StatusFailed
		notification.ErrorMessage = err.Error()
		if an.notificationStore != nil {
			_ = an.notificationStore.Update(ctx, notification)
		}
		return result, err
	}

	// 更新通知状态
	notification.Status = result.Status
	notification.SentAt = result.DeliveredAt
	if an.notificationStore != nil {
		_ = an.notificationStore.Update(ctx, notification)
	}

	return result, nil
}

// SendBatch 批量发送通知
func (an *AlertNotifier) SendBatch(ctx context.Context, notifications []*Notification) ([]*SendResult, error) {
	results := make([]*SendResult, 0, len(notifications))

	// 按渠道分组
	groups := make(map[string][]*Notification)
	for _, notification := range notifications {
		groups[notification.Channel] = append(groups[notification.Channel], notification)
	}

	// 批量发送
	for channelName, channelNotifications := range groups {
		an.mu.RLock()
		channel, exists := an.channels[channelName]
		an.mu.RUnlock()

		if !exists {
			for _, notification := range channelNotifications {
				results = append(results, &SendResult{
					NotificationID: notification.ID,
					Success:        false,
					Status:         StatusFailed,
					Message:        fmt.Sprintf("channel not found: %s", channelName),
				})
			}
			continue
		}

		channelResults, err := channel.SendBatch(ctx, channelNotifications)
		if err != nil {
			for _, notification := range channelNotifications {
				results = append(results, &SendResult{
					NotificationID: notification.ID,
					Success:        false,
					Status:         StatusFailed,
					Error:          err,
				})
			}
			continue
		}

		results = append(results, channelResults...)
	}

	return results, nil
}

// applyTemplate 应用模板
func (an *AlertNotifier) applyTemplate(notification *Notification) error {
	tmpl, err := an.templateManager.GetTemplate(notification.TemplateID)
	if err != nil {
		return err
	}

	// 渲染主题
	if tmpl.SubjectTemplate != "" {
		subject, err := tmpl.RenderSubject(notification.TemplateData)
		if err != nil {
			return fmt.Errorf("failed to render subject: %w", err)
		}
		notification.Subject = subject
	}

	// 渲染内容
	if tmpl.ContentTemplate != "" {
		content, err := tmpl.RenderContent(notification.TemplateData)
		if err != nil {
			return fmt.Errorf("failed to render content: %w", err)
		}
		notification.Content = content
	}

	// 渲染HTML内容
	if tmpl.HTMLTemplate != "" {
		html, err := tmpl.RenderHTML(notification.TemplateData)
		if err != nil {
			return fmt.Errorf("failed to render html: %w", err)
		}
		notification.HTMLContent = html
	}

	return nil
}

// NotifyAlert 通知告警
func (an *AlertNotifier) NotifyAlert(ctx context.Context, alert *AlertInstance, channels []string, recipients []Recipient) error {
	for _, channelName := range channels {
		notification := &Notification{
			AlertID:    alert.ID,
			Channel:    channelName,
			Priority:   getPriorityFromSeverity(alert.Severity),
			Status:     StatusPending,
			Recipients: recipients,
			Subject:    alert.Title,
			Content:    alert.Message,
			TemplateData: map[string]interface{}{
				"alert": alert,
			},
			CreatedAt:  time.Now(),
			MaxRetries: 3,
		}

		// 保存通知
		if an.notificationStore != nil {
			if err := an.notificationStore.Save(ctx, notification); err != nil {
				continue
			}
		}

		// 发送通知
		_, err := an.SendNotification(ctx, notification)
		if err != nil {
			// 记录错误但继续发送其他渠道
			continue
		}
	}

	return nil
}

// getPriorityFromSeverity 从严重程度获取优先级
func getPriorityFromSeverity(severity AlertSeverity) NotificationPriority {
	switch severity {
	case SeverityEmergency:
		return PriorityCritical
	case SeverityCritical:
		return PriorityHigh
	case SeverityWarning:
		return PriorityNormal
	default:
		return PriorityLow
	}
}

// NotificationTemplateManager 通知模板管理器
type NotificationTemplateManager struct {
	mu        sync.RWMutex
	templates map[string]*NotificationTemplate
}

// NotificationTemplate 通知模板
type NotificationTemplate struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Channel         string `json:"channel"`
	SubjectTemplate string `json:"subject_template"`
	ContentTemplate string `json:"content_template"`
	HTMLTemplate    string `json:"html_template"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// NewNotificationTemplateManager 创建通知模板管理器
func NewNotificationTemplateManager() *NotificationTemplateManager {
	return &NotificationTemplateManager{
		templates: make(map[string]*NotificationTemplate),
	}
}

// AddTemplate 添加模板
func (ntm *NotificationTemplateManager) AddTemplate(template *NotificationTemplate) error {
	ntm.mu.Lock()
	defer ntm.mu.Unlock()

	if template.ID == "" {
		return fmt.Errorf("template id is required")
	}

	template.UpdatedAt = time.Now()
	if template.CreatedAt.IsZero() {
		template.CreatedAt = time.Now()
	}

	ntm.templates[template.ID] = template
	return nil
}

// GetTemplate 获取模板
func (ntm *NotificationTemplateManager) GetTemplate(templateID string) (*NotificationTemplate, error) {
	ntm.mu.RLock()
	defer ntm.mu.RUnlock()

	template, exists := ntm.templates[templateID]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	return template, nil
}

// RemoveTemplate 移除模板
func (ntm *NotificationTemplateManager) RemoveTemplate(templateID string) {
	ntm.mu.Lock()
	defer ntm.mu.Unlock()
	delete(ntm.templates, templateID)
}

// RenderSubject 渲染主题
func (t *NotificationTemplate) RenderSubject(data map[string]interface{}) (string, error) {
	return renderTemplate(t.SubjectTemplate, data)
}

// RenderContent 渲染内容
func (t *NotificationTemplate) RenderContent(data map[string]interface{}) (string, error) {
	return renderTemplate(t.ContentTemplate, data)
}

// RenderHTML 渲染HTML
func (t *NotificationTemplate) RenderHTML(data map[string]interface{}) (string, error) {
	return renderTemplate(t.HTMLTemplate, data)
}

// renderTemplate 渲染模板
func renderTemplate(templateStr string, data map[string]interface{}) (string, error) {
	if templateStr == "" {
		return "", nil
	}

	tmpl, err := template.New("notification").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// EscalationEngine 升级引擎
type EscalationEngine struct {
	mu     sync.RWMutex
	rules  map[string]*EscalationRule
}

// EscalationRule 升级规则
type EscalationRule struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	AlertMatcher map[string]string   `json:"alert_matcher"`
	Levels       []EscalationLevel   `json:"levels"`
	Enabled      bool                `json:"enabled"`
	CreatedAt    time.Time           `json:"created_at"`
}

// EscalationLevel 升级级别
type EscalationLevel struct {
	Level       int           `json:"level"`
	After       time.Duration `json:"after"`
	Channels    []string      `json:"channels"`
	Recipients  []Recipient   `json:"recipients"`
}

// NewEscalationEngine 创建升级引擎
func NewEscalationEngine() *EscalationEngine {
	return &EscalationEngine{
		rules: make(map[string]*EscalationRule),
	}
}

// AddRule 添加升级规则
func (ee *EscalationEngine) AddRule(rule *EscalationRule) error {
	ee.mu.Lock()
	defer ee.mu.Unlock()

	if rule.ID == "" {
		return fmt.Errorf("rule id is required")
	}

	ee.rules[rule.ID] = rule
	return nil
}

// RemoveRule 移除升级规则
func (ee *EscalationEngine) RemoveRule(ruleID string) {
	ee.mu.Lock()
	defer ee.mu.Unlock()
	delete(ee.rules, ruleID)
}

// CheckEscalation 检查升级
func (ee *EscalationEngine) CheckEscalation(alert *AlertInstance, triggeredAt time.Time) *EscalationLevel {
	ee.mu.RLock()
	defer ee.mu.RUnlock()

	elapsed := time.Since(triggeredAt)

	for _, rule := range ee.rules {
		if !rule.Enabled {
			continue
		}

		// 检查告警是否匹配
		if !ee.matchAlert(alert, rule.AlertMatcher) {
			continue
		}

		// 检查是否需要升级
		for i := len(rule.Levels) - 1; i >= 0; i-- {
			level := rule.Levels[i]
			if elapsed >= level.After {
				return &level
			}
		}
	}

	return nil
}

// matchAlert 匹配告警
func (ee *EscalationEngine) matchAlert(alert *AlertInstance, matchers map[string]string) bool {
	for key, value := range matchers {
		switch key {
		case "severity":
			if string(alert.Severity) != value {
				return false
			}
		case "category":
			if string(alert.Category) != value {
				return false
			}
		default:
			if alertValue, exists := alert.Labels[key]; !exists || alertValue != value {
				return false
			}
		}
	}
	return true
}

// SilenceChecker 静默检查器
type SilenceChecker struct {
	mu      sync.RWMutex
	silents map[string]*SilentPeriod
}

// SilentPeriod 静默期
type SilentPeriod struct {
	AlertID   string    `json:"alert_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Reason    string    `json:"reason"`
}

// NewSilenceChecker 创建静默检查器
func NewSilenceChecker() *SilenceChecker {
	return &SilenceChecker{
		silents: make(map[string]*SilentPeriod),
	}
}

// IsSilent 检查是否在静默期
func (sc *SilenceChecker) IsSilent(alertID string) bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	period, exists := sc.silents[alertID]
	if !exists {
		return false
	}

	now := time.Now()
	return now.After(period.StartTime) && now.Before(period.EndTime)
}

// StartSilence 开始静默期
func (sc *SilenceChecker) StartSilence(alertID string, duration time.Duration, reason string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	now := time.Now()
	sc.silents[alertID] = &SilentPeriod{
		AlertID:   alertID,
		StartTime: now,
		EndTime:   now.Add(duration),
		Reason:    reason,
	}
}

// EndSilence 结束静默期
func (sc *SilenceChecker) EndSilence(alertID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	delete(sc.silents, alertID)
}

// CleanupExpired 清理过期的静默期
func (sc *SilenceChecker) CleanupExpired() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	now := time.Now()
	for alertID, period := range sc.silents {
		if now.After(period.EndTime) {
			delete(sc.silents, alertID)
		}
	}
}

// NotificationRateLimiter 通知限流器
type NotificationRateLimiter struct {
	mu       sync.RWMutex
	limits   map[string]*ChannelLimit
}

// ChannelLimit 渠道限制
type ChannelLimit struct {
	Channel       string
	MaxPerMinute  int
	MaxPerHour    int
	CurrentMinute int
	CurrentHour   int
	LastReset     time.Time
}

// NewNotificationRateLimiter 创建通知限流器
func NewNotificationRateLimiter() *NotificationRateLimiter {
	return &NotificationRateLimiter{
		limits: make(map[string]*ChannelLimit),
	}
}

// SetLimit 设置限制
func (nrl *NotificationRateLimiter) SetLimit(channel string, maxPerMinute, maxPerHour int) {
	nrl.mu.Lock()
	defer nrl.mu.Unlock()

	nrl.limits[channel] = &ChannelLimit{
		Channel:      channel,
		MaxPerMinute: maxPerMinute,
		MaxPerHour:   maxPerHour,
		LastReset:    time.Now(),
	}
}

// Allow 检查是否允许发送
func (nrl *NotificationRateLimiter) Allow(channel string) bool {
	nrl.mu.Lock()
	defer nrl.mu.Unlock()

	limit, exists := nrl.limits[channel]
	if !exists {
		return true // 没有限制
	}

	// 检查是否需要重置计数
	now := time.Now()
	if now.Sub(limit.LastReset) >= time.Hour {
		limit.CurrentMinute = 0
		limit.CurrentHour = 0
		limit.LastReset = now
	} else if now.Sub(limit.LastReset) >= time.Minute {
		limit.CurrentMinute = 0
	}

	// 检查限制
	if limit.MaxPerMinute > 0 && limit.CurrentMinute >= limit.MaxPerMinute {
		return false
	}
	if limit.MaxPerHour > 0 && limit.CurrentHour >= limit.MaxPerHour {
		return false
	}

	// 更新计数
	limit.CurrentMinute++
	limit.CurrentHour++

	return true
}

// EmailChannel 邮件通知渠道
type EmailChannel struct {
	config *EmailConfig
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost    string
	SMTPPort    int
	Username    string
	Password    string
	FromName    string
	FromAddress string
	UseTLS      bool
}

// NewEmailChannel 创建邮件渠道
func NewEmailChannel(config *EmailConfig) *EmailChannel {
	return &EmailChannel{
		config: config,
	}
}

// Name 渠道名称
func (ec *EmailChannel) Name() string {
	return "email"
}

// Send 发送通知
func (ec *EmailChannel) Send(ctx context.Context, notification *Notification) (*SendResult, error) {
	// 这里应该实现实际的邮件发送逻辑
	// 为了示例，我们只是返回成功
	now := time.Now()
	return &SendResult{
		NotificationID: notification.ID,
		Success:        true,
		Status:         StatusSent,
		DeliveredAt:    &now,
	}, nil
}

// SendBatch 批量发送
func (ec *EmailChannel) SendBatch(ctx context.Context, notifications []*Notification) ([]*SendResult, error) {
	results := make([]*SendResult, 0, len(notifications))
	for _, notification := range notifications {
		result, err := ec.Send(ctx, notification)
		if err != nil {
			results = append(results, &SendResult{
				NotificationID: notification.ID,
				Success:        false,
				Status:         StatusFailed,
				Error:          err,
			})
			continue
		}
		results = append(results, result)
	}
	return results, nil
}

// HealthCheck 健康检查
func (ec *EmailChannel) HealthCheck(ctx context.Context) error {
	// 检查SMTP连接
	return nil
}

// Close 关闭渠道
func (ec *EmailChannel) Close() error {
	return nil
}

// SMSChannel 短信通知渠道
type SMSChannel struct {
	config *SMSConfig
}

// SMSConfig 短信配置
type SMSConfig struct {
	Provider     string
	AccessKey    string
	AccessSecret string
	SignName     string
	Region       string
}

// NewSMSChannel 创建短信渠道
func NewSMSChannel(config *SMSConfig) *SMSChannel {
	return &SMSChannel{
		config: config,
	}
}

// Name 渠道名称
func (sc *SMSChannel) Name() string {
	return "sms"
}

// Send 发送通知
func (sc *SMSChannel) Send(ctx context.Context, notification *Notification) (*SendResult, error) {
	// 这里应该实现实际的短信发送逻辑
	now := time.Now()
	return &SendResult{
		NotificationID: notification.ID,
		Success:        true,
		Status:         StatusSent,
		DeliveredAt:    &now,
	}, nil
}

// SendBatch 批量发送
func (sc *SMSChannel) SendBatch(ctx context.Context, notifications []*Notification) ([]*SendResult, error) {
	results := make([]*SendResult, 0, len(notifications))
	for _, notification := range notifications {
		result, err := sc.Send(ctx, notification)
		if err != nil {
			results = append(results, &SendResult{
				NotificationID: notification.ID,
				Success:        false,
				Status:         StatusFailed,
				Error:          err,
			})
			continue
		}
		results = append(results, result)
	}
	return results, nil
}

// HealthCheck 健康检查
func (sc *SMSChannel) HealthCheck(ctx context.Context) error {
	return nil
}

// Close 关闭渠道
func (sc *SMSChannel) Close() error {
	return nil
}

// DingTalkChannel 钉钉通知渠道
type DingTalkChannel struct {
	config *DingTalkConfig
}

// DingTalkConfig 钉钉配置
type DingTalkConfig struct {
	WebhookURL string
	Secret     string
}

// NewDingTalkChannel 创建钉钉渠道
func NewDingTalkChannel(config *DingTalkConfig) *DingTalkChannel {
	return &DingTalkChannel{
		config: config,
	}
}

// Name 渠道名称
func (dc *DingTalkChannel) Name() string {
	return "dingtalk"
}

// Send 发送通知
func (dc *DingTalkChannel) Send(ctx context.Context, notification *Notification) (*SendResult, error) {
	// 这里应该实现实际的钉钉发送逻辑
	now := time.Now()
	return &SendResult{
		NotificationID: notification.ID,
		Success:        true,
		Status:         StatusSent,
		DeliveredAt:    &now,
	}, nil
}

// SendBatch 批量发送
func (dc *DingTalkChannel) SendBatch(ctx context.Context, notifications []*Notification) ([]*SendResult, error) {
	results := make([]*SendResult, 0, len(notifications))
	for _, notification := range notifications {
		result, err := dc.Send(ctx, notification)
		if err != nil {
			results = append(results, &SendResult{
				NotificationID: notification.ID,
				Success:        false,
				Status:         StatusFailed,
				Error:          err,
			})
			continue
		}
		results = append(results, result)
	}
	return results, nil
}

// HealthCheck 健康检查
func (dc *DingTalkChannel) HealthCheck(ctx context.Context) error {
	return nil
}

// Close 关闭渠道
func (dc *DingTalkChannel) Close() error {
	return nil
}

// WeChatChannel 企业微信通知渠道
type WeChatChannel struct {
	config *WeChatConfig
}

// WeChatConfig 企业微信配置
type WeChatConfig struct {
	CorpID     string
	AgentID    string
	Secret     string
}

// NewWeChatChannel 创建企业微信渠道
func NewWeChatChannel(config *WeChatConfig) *WeChatChannel {
	return &WeChatChannel{
		config: config,
	}
}

// Name 渠道名称
func (wc *WeChatChannel) Name() string {
	return "wechat"
}

// Send 发送通知
func (wc *WeChatChannel) Send(ctx context.Context, notification *Notification) (*SendResult, error) {
	// 这里应该实现实际的企业微信发送逻辑
	now := time.Now()
	return &SendResult{
		NotificationID: notification.ID,
		Success:        true,
		Status:         StatusSent,
		DeliveredAt:    &now,
	}, nil
}

// SendBatch 批量发送
func (wc *WeChatChannel) SendBatch(ctx context.Context, notifications []*Notification) ([]*SendResult, error) {
	results := make([]*SendResult, 0, len(notifications))
	for _, notification := range notifications {
		result, err := wc.Send(ctx, notification)
		if err != nil {
			results = append(results, &SendResult{
				NotificationID: notification.ID,
				Success:        false,
				Status:         StatusFailed,
				Error:          err,
			})
			continue
		}
		results = append(results, result)
	}
	return results, nil
}

// HealthCheck 健康检查
func (wc *WeChatChannel) HealthCheck(ctx context.Context) error {
	return nil
}

// Close 关闭渠道
func (wc *WeChatChannel) Close() error {
	return nil
}
