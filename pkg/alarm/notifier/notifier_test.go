package notifier

import (
	"context"
	"testing"
	"time"
)

// TestNotificationBuilder 测试通知构建器
func TestNotificationBuilder(t *testing.T) {
	builder := NewNotificationBuilder()
	notification := builder.
		WithID("test-001").
		WithAlarmID("alarm-001").
		WithChannel(ChannelSMS).
		WithPriority(PriorityHigh).
		WithSubject("测试通知").
		WithContent("这是一条测试通知").
		AddRecipient(Recipient{
			UserID: "user-001",
			Name:   "张三",
			Phone:  "13800138000",
			Email:  "zhangsan@example.com",
		}).
		Build()

	if notification.ID != "test-001" {
		t.Errorf("expected ID test-001, got %s", notification.ID)
	}
	if notification.Channel != ChannelSMS {
		t.Errorf("expected channel sms, got %s", notification.Channel)
	}
	if len(notification.Recipients) != 1 {
		t.Errorf("expected 1 recipient, got %d", len(notification.Recipients))
	}
}

// TestTokenBucketRateLimiter 测试令牌桶限流器
func TestTokenBucketRateLimiter(t *testing.T) {
	limiter := NewTokenBucketRateLimiter(10, 5)

	// 测试允许
	for i := 0; i < 5; i++ {
		if !limiter.Allow("test-key") {
			t.Errorf("expected allow at iteration %d", i)
		}
	}

	// 测试拒绝
	if limiter.Allow("test-key") {
		t.Error("expected deny after burst")
	}

	// 测试重置
	limiter.Reset("test-key")
	if !limiter.Allow("test-key") {
		t.Error("expected allow after reset")
	}
}

// TestMemoryNotificationStore 测试内存通知存储
func TestMemoryNotificationStore(t *testing.T) {
	store := NewMemoryNotificationStore()
	ctx := context.Background()

	notification := &Notification{
		ID:       "test-001",
		AlarmID:  "alarm-001",
		Channel:  ChannelSMS,
		Priority: PriorityHigh,
		Status:   StatusPending,
		Subject:  "测试通知",
		Content:  "测试内容",
	}

	// 测试保存
	if err := store.Save(ctx, notification); err != nil {
		t.Errorf("failed to save notification: %v", err)
	}

	// 测试获取
	got, err := store.Get(ctx, "test-001")
	if err != nil {
		t.Errorf("failed to get notification: %v", err)
	}
	if got.ID != notification.ID {
		t.Errorf("expected ID %s, got %s", notification.ID, got.ID)
	}

	// 测试更新
	notification.Status = StatusSent
	if err := store.Update(ctx, notification); err != nil {
		t.Errorf("failed to update notification: %v", err)
	}

	// 测试按状态获取
	pending, err := store.GetByStatus(ctx, StatusSent, 10)
	if err != nil {
		t.Errorf("failed to get by status: %v", err)
	}
	if len(pending) != 1 {
		t.Errorf("expected 1 notification, got %d", len(pending))
	}

	// 测试删除
	if err := store.Delete(ctx, "test-001"); err != nil {
		t.Errorf("failed to delete notification: %v", err)
	}

	_, err = store.Get(ctx, "test-001")
	if err == nil {
		t.Error("expected error after delete")
	}
}

// TestMemorySilenceChecker 测试内存静默期检查器
func TestMemorySilenceChecker(t *testing.T) {
	checker := NewMemorySilenceChecker()

	// 测试不在静默期
	if checker.IsSilent("alarm-001") {
		t.Error("expected not in silence period")
	}

	// 开始静默期
	if err := checker.StartSilence("alarm-001", 5*time.Second); err != nil {
		t.Errorf("failed to start silence: %v", err)
	}

	// 测试在静默期
	if !checker.IsSilent("alarm-001") {
		t.Error("expected in silence period")
	}

	// 结束静默期
	if err := checker.EndSilence("alarm-001"); err != nil {
		t.Errorf("failed to end silence: %v", err)
	}

	// 测试不在静默期
	if checker.IsSilent("alarm-001") {
		t.Error("expected not in silence period after end")
	}
}

// TestTemplateManager 测试模板管理器
func TestTemplateManager(t *testing.T) {
	store := NewMemoryTemplateStore()
	mgr := NewTemplateManager(store)
	ctx := context.Background()

	tmpl := &NotificationTemplate{
		ID:          "test-template",
		Name:        "测试模板",
		Type:        TemplateTypeSMS,
		Channel:     ChannelSMS,
		Language:    LanguageZH,
		Content:     "告警：{{.Message}}，时间：{{.Time}}",
		Description: "测试模板",
		Variables: []TemplateVariable{
			{Name: "Message", Description: "告警消息", Required: true},
			{Name: "Time", Description: "时间", Required: true},
		},
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 测试创建
	if err := mgr.Create(ctx, tmpl); err != nil {
		t.Errorf("failed to create template: %v", err)
	}

	// 测试获取
	got, err := mgr.GetTemplate(ctx, "test-template")
	if err != nil {
		t.Errorf("failed to get template: %v", err)
	}
	if got.ID != tmpl.ID {
		t.Errorf("expected ID %s, got %s", tmpl.ID, got.ID)
	}

	// 测试渲染
	data := map[string]interface{}{
		"Message": "设备故障",
		"Time":    "2024-01-01 12:00:00",
	}
	rendered, err := mgr.Render("test-template", data)
	if err != nil {
		t.Errorf("failed to render template: %v", err)
	}
	expected := "告警：设备故障，时间：2024-01-01 12:00:00"
	if rendered != expected {
		t.Errorf("expected %s, got %s", expected, rendered)
	}

	// 测试验证
	if err := mgr.Validate("test-template", data); err != nil {
		t.Errorf("validation failed: %v", err)
	}

	// 测试缺少必需变量
	invalidData := map[string]interface{}{
		"Message": "设备故障",
	}
	if err := mgr.Validate("test-template", invalidData); err == nil {
		t.Error("expected validation error for missing required variable")
	}
}

// TestWebSocketHub 测试WebSocket中心
func TestWebSocketHub(t *testing.T) {
	hub := NewWebSocketHub()

	// 启动hub
	go hub.Run()
	defer func() {
		hub.mu.Lock()
		hub.running = false
		hub.mu.Unlock()
	}()

	// 等待启动
	time.Sleep(100 * time.Millisecond)

	if !hub.IsRunning() {
		t.Error("expected hub to be running")
	}

	// 注册客户端
	client := &WebSocketClient{
		ID:     "client-001",
		UserID: "user-001",
		Send:   make(chan []byte, 10),
	}
	hub.Register(client)

	// 等待注册
	time.Sleep(100 * time.Millisecond)

	// 检查用户在线
	if !hub.IsUserOnline("user-001") {
		t.Error("expected user to be online")
	}

	// 发送消息
	message := &InternalMessage{
		ID:      "msg-001",
		Title:   "测试消息",
		Content: "这是一条测试消息",
	}
	if err := hub.SendToUser("user-001", message); err != nil {
		t.Errorf("failed to send message: %v", err)
	}

	// 注销客户端
	hub.Unregister(client)

	// 等待注销
	time.Sleep(100 * time.Millisecond)

	// 检查用户离线
	if hub.IsUserOnline("user-001") {
		t.Error("expected user to be offline")
	}
}

// TestUnreadMessageManager 测试未读消息管理器
func TestUnreadMessageManager(t *testing.T) {
	store := NewMemoryUnreadMessageStore()
	mgr := NewUnreadMessageManager(store)
	ctx := context.Background()

	userID := "user-001"

	// 测试初始未读数
	count, err := mgr.Get(ctx, userID)
	if err != nil {
		t.Errorf("failed to get unread count: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 unread, got %d", count)
	}

	// 测试增加未读数
	if err := mgr.Increment(ctx, userID); err != nil {
		t.Errorf("failed to increment: %v", err)
	}
	count, _ = mgr.Get(ctx, userID)
	if count != 1 {
		t.Errorf("expected 1 unread, got %d", count)
	}

	// 测试减少未读数
	if err := mgr.Decrement(ctx, userID); err != nil {
		t.Errorf("failed to decrement: %v", err)
	}
	count, _ = mgr.Get(ctx, userID)
	if count != 0 {
		t.Errorf("expected 0 unread, got %d", count)
	}

	// 测试重置
	mgr.Increment(ctx, userID)
	mgr.Increment(ctx, userID)
	if err := mgr.Reset(ctx, userID); err != nil {
		t.Errorf("failed to reset: %v", err)
	}
	count, _ = mgr.Get(ctx, userID)
	if count != 0 {
		t.Errorf("expected 0 unread after reset, got %d", count)
	}
}

// TestNotificationScheduler 测试通知调度器
func TestNotificationScheduler(t *testing.T) {
	store := NewMemoryNotificationStore()
	templateStore := NewMemoryTemplateStore()
	templateMgr := NewTemplateManager(templateStore)
	silenceChecker := NewMemorySilenceChecker()

	config := &SchedulerConfig{
		QueueSize:    100,
		WorkerCount:  2,
		MaxRetries:   3,
		RetryDelay:   1 * time.Second,
		RetryBackoff: 2.0,
		SendTimeout:  5 * time.Second,
	}

	scheduler := NewNotificationScheduler(store, nil, templateMgr, silenceChecker, config)

	// 启动调度器
	if err := scheduler.Start(); err != nil {
		t.Errorf("failed to start scheduler: %v", err)
	}
	defer scheduler.Stop()

	// 检查运行状态
	if !scheduler.IsRunning() {
		t.Error("expected scheduler to be running")
	}

	// 调度通知
	notification := &Notification{
		ID:       "test-001",
		AlarmID:  "alarm-001",
		Channel:  ChannelInternal,
		Priority: PriorityHigh,
		Status:   StatusPending,
		Subject:  "测试通知",
		Content:  "测试内容",
		Recipients: []Recipient{
			{UserID: "user-001", Name: "张三"},
		},
	}

	ctx := context.Background()
	if err := scheduler.Schedule(ctx, notification); err != nil {
		t.Errorf("failed to schedule notification: %v", err)
	}

	// 检查队列统计
	stats := scheduler.GetQueueStats()
	if stats[PriorityHigh] != 1 {
		t.Errorf("expected 1 in high priority queue, got %d", stats[PriorityHigh])
	}
}
