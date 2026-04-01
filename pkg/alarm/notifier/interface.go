package notifier

import (
	"context"
	"time"
)

// NotificationChannel 通知渠道类型
type NotificationChannel string

const (
	ChannelSMS     NotificationChannel = "sms"
	ChannelEmail   NotificationChannel = "email"
	ChannelWebhook NotificationChannel = "webhook"
	ChannelInternal NotificationChannel = "internal"
	ChannelWeChat  NotificationChannel = "wechat"
	ChannelDingTalk NotificationChannel = "dingtalk"
)

// NotificationPriority 通知优先级
type NotificationPriority int

const (
	PriorityLow      NotificationPriority = 1
	PriorityNormal   NotificationPriority = 2
	PriorityHigh     NotificationPriority = 3
	PriorityCritical NotificationPriority = 4
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

// Notification 通知消息
type Notification struct {
	ID           string                 `json:"id"`
	AlarmID      string                 `json:"alarm_id"`
	Channel      NotificationChannel    `json:"channel"`
	Priority     NotificationPriority   `json:"priority"`
	Status       NotificationStatus     `json:"status"`

	// 接收者信息
	Recipients   []Recipient            `json:"recipients"`

	// 通知内容
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
	UserID    string `json:"user_id,omitempty"`
	Name      string `json:"name,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Email     string `json:"email,omitempty"`
	OpenID    string `json:"open_id,omitempty"`    // 微信OpenID
	DingTalkID string `json:"dingtalk_id,omitempty"` // 钉钉ID
}

// Attachment 附件
type Attachment struct {
	Name     string `json:"name"`
	Content  []byte `json:"content"`
	MimeType string `json:"mime_type"`
}

// NotificationResult 通知结果
type NotificationResult struct {
	NotificationID string            `json:"notification_id"`
	Success        bool              `json:"success"`
	Status         NotificationStatus `json:"status"`
	Message        string            `json:"message,omitempty"`
	ExternalID     string            `json:"external_id,omitempty"` // 第三方平台返回的ID
	DeliveredAt    *time.Time        `json:"delivered_at,omitempty"`
	Error          error             `json:"error,omitempty"`
}

// NotificationConfig 通知配置
type NotificationConfig struct {
	// 基础配置
	Enabled      bool              `json:"enabled"`
	Channel      NotificationChannel `json:"channel"`
	Timeout      time.Duration     `json:"timeout"`
	MaxRetries   int               `json:"max_retries"`
	RetryDelay   time.Duration     `json:"retry_delay"`

	// 限流配置
	RateLimit    int               `json:"rate_limit"`     // 每分钟最大发送数
	BurstLimit   int               `json:"burst_limit"`    // 突发最大数

	// 静默期配置
	SilencePeriod time.Duration    `json:"silence_period"` // 静默期时长
	SilenceStart  *time.Time       `json:"silence_start"`  // 静默期开始时间

	// 渠道特定配置
	SMSConfig    *SMSConfig        `json:"sms_config,omitempty"`
	EmailConfig  *EmailConfig      `json:"email_config,omitempty"`
	WebhookConfig *WebhookConfig   `json:"webhook_config,omitempty"`
}

// SMSConfig 短信配置
type SMSConfig struct {
	Provider     string `json:"provider"`      // aliyun, tencent
	AccessKey    string `json:"access_key"`
	AccessSecret string `json:"access_secret"`
	SignName     string `json:"sign_name"`
	Region       string `json:"region"`
	Endpoint     string `json:"endpoint"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	FromName     string `json:"from_name"`
	FromAddress  string `json:"from_address"`
	UseTLS       bool   `json:"use_tls"`
}

// WebhookConfig Webhook配置
type WebhookConfig struct {
	URL          string            `json:"url"`
	Method       string            `json:"method"`
	Headers      map[string]string `json:"headers"`
	Timeout      time.Duration     `json:"timeout"`
}

// Notifier 通知器接口
type Notifier interface {
	// Channel 返回通知渠道类型
	Channel() NotificationChannel

	// Send 发送通知
	Send(ctx context.Context, notification *Notification) (*NotificationResult, error)

	// SendBatch 批量发送通知
	SendBatch(ctx context.Context, notifications []*Notification) ([]*NotificationResult, error)

	// Validate 验证通知内容
	Validate(notification *Notification) error

	// HealthCheck 健康检查
	HealthCheck(ctx context.Context) error

	// Close 关闭通知器
	Close() error
}

// NotifierFactory 通知器工厂接口
type NotifierFactory interface {
	// Create 创建通知器
	Create(config *NotificationConfig) (Notifier, error)

	// Supports 是否支持指定渠道
	Supports(channel NotificationChannel) bool
}

// NotificationStore 通知存储接口
type NotificationStore interface {
	// Save 保存通知
	Save(ctx context.Context, notification *Notification) error

	// Update 更新通知
	Update(ctx context.Context, notification *Notification) error

	// Get 获取通知
	Get(ctx context.Context, id string) (*Notification, error)

	// GetByAlarmID 根据告警ID获取通知列表
	GetByAlarmID(ctx context.Context, alarmID string) ([]*Notification, error)

	// GetPending 获取待发送的通知
	GetPending(ctx context.Context, limit int) ([]*Notification, error)

	// GetByStatus 根据状态获取通知列表
	GetByStatus(ctx context.Context, status NotificationStatus, limit int) ([]*Notification, error)

	// Delete 删除通知
	Delete(ctx context.Context, id string) error
}

// NotificationLogger 通知日志接口
type NotificationLogger interface {
	// Log 记录通知日志
	Log(ctx context.Context, notification *Notification, result *NotificationResult) error

	// Query 查询通知日志
	Query(ctx context.Context, query *NotificationLogQuery) ([]*NotificationLog, int64, error)
}

// NotificationLogQuery 通知日志查询
type NotificationLogQuery struct {
	AlarmID    string
	Channel    NotificationChannel
	Status     NotificationStatus
	StartTime  *time.Time
	EndTime    *time.Time
	Page       int
	PageSize   int
}

// NotificationLog 通知日志
type NotificationLog struct {
	ID             string              `json:"id"`
	NotificationID string              `json:"notification_id"`
	AlarmID        string              `json:"alarm_id"`
	Channel        NotificationChannel `json:"channel"`
	Status         NotificationStatus  `json:"status"`
	Recipients     string              `json:"recipients"` // JSON格式
	Subject        string              `json:"subject"`
	Content        string              `json:"content"`
	ErrorMessage   string              `json:"error_message,omitempty"`
	ExternalID     string              `json:"external_id,omitempty"`
	RetryCount     int                 `json:"retry_count"`
	CreatedAt      time.Time           `json:"created_at"`
	SentAt         *time.Time          `json:"sent_at,omitempty"`
	Duration       int64               `json:"duration"` // 发送耗时(毫秒)
}

// RateLimiter 限流器接口
type RateLimiter interface {
	// Allow 是否允许发送
	Allow(key string) bool

	// Wait 等待直到可以发送
	Wait(ctx context.Context, key string) error

	// Reset 重置限流器
	Reset(key string)
}

// SilenceChecker 静默期检查器接口
type SilenceChecker interface {
	// IsSilent 是否处于静默期
	IsSilent(alarmID string) bool

	// StartSilence 开始静默期
	StartSilence(alarmID string, duration time.Duration) error

	// EndSilence 结束静默期
	EndSilence(alarmID string) error
}
