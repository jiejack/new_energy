package notifier

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrInternalConfigInvalid = errors.New("internal notification config is invalid")
	ErrInternalRecipientEmpty = errors.New("internal notification recipient is empty")
	ErrInternalSendFailed    = errors.New("internal notification send failed")
	ErrWebSocketNotConnected = errors.New("websocket not connected")
)

// InternalNotifier 系统内消息通知器
type InternalNotifier struct {
	config        *NotificationConfig
	wsHub         *WebSocketHub
	messageCenter MessageCenter
	rateLimiter   RateLimiter
	templateMgr   TemplateManager
	mu            sync.RWMutex
}

// NewInternalNotifier 创建系统内消息通知器
func NewInternalNotifier(config *NotificationConfig, wsHub *WebSocketHub, messageCenter MessageCenter, templateMgr TemplateManager) (*InternalNotifier, error) {
	if config == nil {
		return nil, ErrInternalConfigInvalid
	}

	return &InternalNotifier{
		config:        config,
		wsHub:         wsHub,
		messageCenter: messageCenter,
		rateLimiter:   NewTokenBucketRateLimiter(config.RateLimit, config.BurstLimit),
		templateMgr:   templateMgr,
	}, nil
}

// Channel 返回通知渠道类型
func (i *InternalNotifier) Channel() NotificationChannel {
	return ChannelInternal
}

// Send 发送系统内消息通知
func (i *InternalNotifier) Send(ctx context.Context, notification *Notification) (*NotificationResult, error) {
	if err := i.Validate(notification); err != nil {
		return nil, err
	}

	// 限流检查
	key := fmt.Sprintf("internal:%s", notification.AlarmID)
	if !i.rateLimiter.Allow(key) {
		return &NotificationResult{
			NotificationID: notification.ID,
			Success:        false,
			Status:         StatusFailed,
			Message:        "rate limit exceeded",
			Error:          ErrSMSRateLimitExceeded,
		}, ErrSMSRateLimitExceeded
	}

	// 准备消息内容
	message, err := i.prepareMessage(notification)
	if err != nil {
		return nil, err
	}

	// 保存到消息中心
	if i.messageCenter != nil {
		if err := i.messageCenter.Save(ctx, message); err != nil {
			return nil, fmt.Errorf("failed to save message: %w", err)
		}
	}

	// 通过WebSocket实时推送
	if i.wsHub != nil {
		for _, recipient := range notification.Recipients {
			if recipient.UserID != "" {
				if err := i.wsHub.SendToUser(recipient.UserID, message); err != nil {
					// WebSocket推送失败不影响消息保存
					// 记录错误但继续
				}
			}
		}
	}

	now := time.Now()
	return &NotificationResult{
		NotificationID: notification.ID,
		Success:        true,
		Status:         StatusSent,
		Message:        "internal notification sent successfully",
		DeliveredAt:    &now,
	}, nil
}

// SendBatch 批量发送系统内消息通知
func (i *InternalNotifier) SendBatch(ctx context.Context, notifications []*Notification) ([]*NotificationResult, error) {
	results := make([]*NotificationResult, len(notifications))

	var wg sync.WaitGroup
	var mu sync.Mutex

	for idx, notification := range notifications {
		wg.Add(1)
		go func(index int, notif *Notification) {
			defer wg.Done()
			result, err := i.Send(ctx, notif)
			mu.Lock()
			if err != nil {
				results[index] = &NotificationResult{
					NotificationID: notif.ID,
					Success:        false,
					Status:         StatusFailed,
					Message:        err.Error(),
					Error:          err,
				}
			} else {
				results[index] = result
			}
			mu.Unlock()
		}(idx, notification)
	}

	wg.Wait()
	return results, nil
}

// Validate 验证系统内消息通知
func (i *InternalNotifier) Validate(notification *Notification) error {
	if notification == nil {
		return errors.New("notification is nil")
	}

	if len(notification.Recipients) == 0 {
		return ErrInternalRecipientEmpty
	}

	// 验证用户ID
	for _, r := range notification.Recipients {
		if r.UserID == "" {
			return fmt.Errorf("recipient %s has no user id", r.Name)
		}
	}

	return nil
}

// HealthCheck 健康检查
func (i *InternalNotifier) HealthCheck(ctx context.Context) error {
	if i.wsHub != nil {
		if !i.wsHub.IsRunning() {
			return errors.New("websocket hub is not running")
		}
	}
	return nil
}

// Close 关闭系统内消息通知器
func (i *InternalNotifier) Close() error {
	return nil
}

// prepareMessage 准备消息内容
func (i *InternalNotifier) prepareMessage(notification *Notification) (*InternalMessage, error) {
	message := &InternalMessage{
		ID:          notification.ID,
		AlarmID:     notification.AlarmID,
		Type:        "alarm",
		Title:       notification.Subject,
		Content:     notification.Content,
		Priority:    int(notification.Priority),
		CreatedAt:   time.Now(),
		Read:        false,
		Tags:        notification.Tags,
	}

	// 如果有模板，使用模板渲染
	if notification.TemplateID != "" && i.templateMgr != nil {
		rendered, err := i.templateMgr.Render(notification.TemplateID, notification.TemplateData)
		if err != nil {
			return nil, err
		}
		message.Content = rendered
	}

	// 如果有HTML内容，使用HTML内容
	if notification.HTMLContent != "" {
		message.HTMLContent = notification.HTMLContent
	}

	// 设置接收者
	for _, r := range notification.Recipients {
		message.Recipients = append(message.Recipients, r.UserID)
	}

	return message, nil
}

// InternalMessage 系统内消息
type InternalMessage struct {
	ID           string                 `json:"id"`
	AlarmID      string                 `json:"alarm_id"`
	Type         string                 `json:"type"`
	Title        string                 `json:"title"`
	Content      string                 `json:"content"`
	HTMLContent  string                 `json:"html_content,omitempty"`
	Priority     int                    `json:"priority"`
	Recipients   []string               `json:"recipients"`
	Read         bool                   `json:"read"`
	ReadAt       *time.Time             `json:"read_at,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	Tags         map[string]string      `json:"tags,omitempty"`
	Extra        map[string]interface{} `json:"extra,omitempty"`
}

// MessageCenter 消息中心接口
type MessageCenter interface {
	// Save 保存消息
	Save(ctx context.Context, message *InternalMessage) error

	// Get 获取消息
	Get(ctx context.Context, id string) (*InternalMessage, error)

	// GetByUser 获取用户消息列表
	GetByUser(ctx context.Context, userID string, unreadOnly bool, page, pageSize int) ([]*InternalMessage, int64, error)

	// MarkAsRead 标记为已读
	MarkAsRead(ctx context.Context, messageID string, userID string) error

	// MarkAllAsRead 标记所有消息为已读
	MarkAllAsRead(ctx context.Context, userID string) error

	// Delete 删除消息
	Delete(ctx context.Context, messageID string, userID string) error

	// GetUnreadCount 获取未读消息数
	GetUnreadCount(ctx context.Context, userID string) (int64, error)
}

// WebSocketHub WebSocket中心
type WebSocketHub struct {
	clients     map[string]map[*WebSocketClient]bool
	broadcast   chan *WebSocketMessage
	register    chan *WebSocketClient
	unregister  chan *WebSocketClient
	send        chan *UserMessage
	running     bool
	mu          sync.RWMutex
}

// WebSocketClient WebSocket客户端
type WebSocketClient struct {
	ID     string
	UserID string
	Send   chan []byte
}

// WebSocketMessage WebSocket消息
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// UserMessage 用户消息
type UserMessage struct {
	UserID  string
	Message *InternalMessage
}

// NewWebSocketHub 创建WebSocket中心
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[string]map[*WebSocketClient]bool),
		broadcast:  make(chan *WebSocketMessage, 256),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
		send:       make(chan *UserMessage, 256),
		running:    false,
	}
}

// Run 运行WebSocket中心
func (h *WebSocketHub) Run() {
	h.mu.Lock()
	h.running = true
	h.mu.Unlock()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; !ok {
				h.clients[client.UserID] = make(map[*WebSocketClient]bool)
			}
			h.clients[client.UserID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.UserID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.clients, client.UserID)
					}
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			data, err := json.Marshal(message)
			if err != nil {
				h.mu.RUnlock()
				continue
			}

			for _, clients := range h.clients {
				for client := range clients {
					select {
					case client.Send <- data:
					default:
						close(client.Send)
						delete(clients, client)
					}
				}
			}
			h.mu.RUnlock()

		case userMessage := <-h.send:
			h.mu.RLock()
			clients, ok := h.clients[userMessage.UserID]
			if !ok {
				h.mu.RUnlock()
				continue
			}

			wsMessage := &WebSocketMessage{
				Type:    "notification",
				Payload: userMessage.Message,
			}
			data, err := json.Marshal(wsMessage)
			if err != nil {
				h.mu.RUnlock()
				continue
			}

			for client := range clients {
				select {
				case client.Send <- data:
				default:
					close(client.Send)
					delete(clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// IsRunning 是否运行中
func (h *WebSocketHub) IsRunning() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.running
}

// Register 注册客户端
func (h *WebSocketHub) Register(client *WebSocketClient) {
	h.register <- client
}

// Unregister 注销客户端
func (h *WebSocketHub) Unregister(client *WebSocketClient) {
	h.unregister <- client
}

// Broadcast 广播消息
func (h *WebSocketHub) Broadcast(message *WebSocketMessage) {
	h.broadcast <- message
}

// SendToUser 发送消息给指定用户
func (h *WebSocketHub) SendToUser(userID string, message *InternalMessage) error {
	h.mu.RLock()
	_, ok := h.clients[userID]
	h.mu.RUnlock()

	if !ok {
		return ErrWebSocketNotConnected
	}

	h.send <- &UserMessage{
		UserID:  userID,
		Message: message,
	}

	return nil
}

// GetOnlineUsers 获取在线用户列表
func (h *WebSocketHub) GetOnlineUsers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]string, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	return users
}

// IsUserOnline 检查用户是否在线
func (h *WebSocketHub) IsUserOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.clients[userID]
	return ok && len(clients) > 0
}

// UnreadMessageManager 未读消息管理器
type UnreadMessageManager struct {
	store       UnreadMessageStore
	mu          sync.RWMutex
}

// UnreadMessageStore 未读消息存储接口
type UnreadMessageStore interface {
	// Increment 增加未读数
	Increment(ctx context.Context, userID string) error

	// Decrement 减少未读数
	Decrement(ctx context.Context, userID string) error

	// Get 获取未读数
	Get(ctx context.Context, userID string) (int64, error)

	// Reset 重置未读数
	Reset(ctx context.Context, userID string) error
}

// NewUnreadMessageManager 创建未读消息管理器
func NewUnreadMessageManager(store UnreadMessageStore) *UnreadMessageManager {
	return &UnreadMessageManager{
		store: store,
	}
}

// Increment 增加未读数
func (m *UnreadMessageManager) Increment(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.store.Increment(ctx, userID)
}

// Decrement 减少未读数
func (m *UnreadMessageManager) Decrement(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.store.Decrement(ctx, userID)
}

// Get 获取未读数
func (m *UnreadMessageManager) Get(ctx context.Context, userID string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.store.Get(ctx, userID)
}

// Reset 重置未读数
func (m *UnreadMessageManager) Reset(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.store.Reset(ctx, userID)
}

// MemoryUnreadMessageStore 内存未读消息存储
type MemoryUnreadMessageStore struct {
	counts map[string]int64
	mu     sync.RWMutex
}

// NewMemoryUnreadMessageStore 创建内存未读消息存储
func NewMemoryUnreadMessageStore() *MemoryUnreadMessageStore {
	return &MemoryUnreadMessageStore{
		counts: make(map[string]int64),
	}
}

// Increment 增加未读数
func (s *MemoryUnreadMessageStore) Increment(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counts[userID]++
	return nil
}

// Decrement 减少未读数
func (s *MemoryUnreadMessageStore) Decrement(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.counts[userID] > 0 {
		s.counts[userID]--
	}
	return nil
}

// Get 获取未读数
func (s *MemoryUnreadMessageStore) Get(ctx context.Context, userID string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.counts[userID], nil
}

// Reset 重置未读数
func (s *MemoryUnreadMessageStore) Reset(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counts[userID] = 0
	return nil
}
