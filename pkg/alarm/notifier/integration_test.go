package notifier

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestIntegration 集成测试
func TestIntegration(t *testing.T) {
	// 1. 初始化所有组件
	templateStore := NewMemoryTemplateStore()
	templateMgr := NewTemplateManager(templateStore)
	notificationStore := NewMemoryNotificationStore()
	silenceChecker := NewMemorySilenceChecker()
	wsHub := NewWebSocketHub()
	unreadStore := NewMemoryUnreadMessageStore()

	// 初始化内置模板
	if err := InitBuiltInTemplates(templateStore); err != nil {
		t.Fatalf("Failed to init built-in templates: %v", err)
	}

	// 启动WebSocket Hub
	go wsHub.Run()
	time.Sleep(100 * time.Millisecond)

	// 2. 创建调度器
	config := &SchedulerConfig{
		QueueSize:    100,
		WorkerCount:  2,
		MaxRetries:   3,
		RetryDelay:   1 * time.Second,
		RetryBackoff: 2.0,
		SendTimeout:  5 * time.Second,
	}

	scheduler := NewNotificationScheduler(
		notificationStore,
		nil,
		templateMgr,
		silenceChecker,
		config,
	)

	// 3. 创建并注册通知器
	internalConfig := &NotificationConfig{
		Enabled:    true,
		Channel:    ChannelInternal,
		Timeout:    10 * time.Second,
		MaxRetries: 2,
		RetryDelay: 1 * time.Second,
		RateLimit:  100,
		BurstLimit: 20,
	}

	internalNotifier, err := NewInternalNotifier(internalConfig, wsHub, nil, templateMgr)
	if err != nil {
		t.Fatalf("Failed to create internal notifier: %v", err)
	}
	scheduler.RegisterNotifier(ChannelInternal, internalNotifier)

	// 4. 启动调度器
	if err := scheduler.Start(); err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}
	defer scheduler.Stop()

	// 5. 模拟客户端连接
	client := &WebSocketClient{
		ID:     "client-001",
		UserID: "user-001",
		Send:   make(chan []byte, 10),
	}
	wsHub.Register(client)
	time.Sleep(100 * time.Millisecond)

	// 6. 创建并发送通知
	ctx := context.Background()
	notification := NewNotificationBuilder().
		WithAlarmID("alarm-001").
		WithChannel(ChannelInternal).
		WithPriority(PriorityHigh).
		WithSubject("测试告警").
		WithContent("这是一条测试告警消息").
		AddRecipient(Recipient{
			UserID: "user-001",
			Name:   "张三",
		}).
		Build()

	if err := scheduler.Schedule(ctx, notification); err != nil {
		t.Fatalf("Failed to schedule notification: %v", err)
	}

	// 7. 等待处理
	time.Sleep(500 * time.Millisecond)

	// 8. 验证通知状态
	got, err := notificationStore.Get(ctx, notification.ID)
	if err != nil {
		t.Fatalf("Failed to get notification: %v", err)
	}

	if got.Status != StatusSent && got.Status != StatusSending {
		t.Errorf("Expected status sent or sending, got %s", got.Status)
	}

	// 9. 测试静默期
	if !silenceChecker.IsSilent("alarm-001") {
		t.Log("Alarm not in silence period")
	}

	// 10. 测试未读消息管理
	if err := unreadStore.Increment(ctx, "user-001"); err != nil {
		t.Errorf("Failed to increment unread count: %v", err)
	}

	count, err := unreadStore.Get(ctx, "user-001")
	if err != nil {
		t.Errorf("Failed to get unread count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 unread, got %d", count)
	}

	// 11. 清理
	wsHub.Unregister(client)
}

// TestMultiChannelIntegration 多渠道集成测试
func TestMultiChannelIntegration(t *testing.T) {
	// 初始化组件
	templateStore := NewMemoryTemplateStore()
	templateMgr := NewTemplateManager(templateStore)
	notificationStore := NewMemoryNotificationStore()
	silenceChecker := NewMemorySilenceChecker()

	if err := InitBuiltInTemplates(templateStore); err != nil {
		t.Fatalf("Failed to init templates: %v", err)
	}

	// 创建调度器
	config := DefaultSchedulerConfig()
	scheduler := NewNotificationScheduler(notificationStore, nil, templateMgr, silenceChecker, config)

	// 创建模拟通知器
	mockNotifier := &MockNotifier{
		channel: ChannelSMS,
		results: make([]*NotificationResult, 0),
	}
	scheduler.RegisterNotifier(ChannelSMS, mockNotifier)

	// 启动调度器
	if err := scheduler.Start(); err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}
	defer scheduler.Stop()

	// 发送多渠道通知
	ctx := context.Background()
	channels := []NotificationChannel{ChannelSMS, ChannelEmail}

	for i, channel := range channels {
		notification := NewNotificationBuilder().
			WithID(fmt.Sprintf("notif-%d", i)).
			WithAlarmID(fmt.Sprintf("alarm-%d", i)).
			WithChannel(channel).
			WithPriority(PriorityNormal).
			WithSubject("测试通知").
			WithContent("测试内容").
			AddRecipient(Recipient{
				UserID: "user-001",
				Name:   "张三",
				Phone:  "13800138000",
				Email:  "test@example.com",
			}).
			Build()

		if err := scheduler.Schedule(ctx, notification); err != nil {
			t.Errorf("Failed to schedule notification for channel %s: %v", channel, err)
		}
	}

	// 等待处理
	time.Sleep(500 * time.Millisecond)

	// 验证队列状态
	stats := scheduler.GetQueueStats()
	t.Logf("Queue stats: %+v", stats)
}

// MockNotifier 模拟通知器
type MockNotifier struct {
	channel NotificationChannel
	results []*NotificationResult
}

func (m *MockNotifier) Channel() NotificationChannel {
	return m.channel
}

func (m *MockNotifier) Send(ctx context.Context, notification *Notification) (*NotificationResult, error) {
	now := time.Now()
	result := &NotificationResult{
		NotificationID: notification.ID,
		Success:        true,
		Status:         StatusSent,
		Message:        "mock send success",
		DeliveredAt:    &now,
	}
	m.results = append(m.results, result)
	return result, nil
}

func (m *MockNotifier) SendBatch(ctx context.Context, notifications []*Notification) ([]*NotificationResult, error) {
	results := make([]*NotificationResult, len(notifications))
	for i, n := range notifications {
		result, _ := m.Send(ctx, n)
		results[i] = result
	}
	return results, nil
}

func (m *MockNotifier) Validate(notification *Notification) error {
	return nil
}

func (m *MockNotifier) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *MockNotifier) Close() error {
	return nil
}

// TestTemplateIntegration 模板集成测试
func TestTemplateIntegration(t *testing.T) {
	store := NewMemoryTemplateStore()
	mgr := NewTemplateManager(store)
	ctx := context.Background()

	// 创建模板
	tmpl := &NotificationTemplate{
		ID:       "test-template_zh-CN",
		Name:     "测试模板",
		Type:     TemplateTypeSMS,
		Channel:  ChannelSMS,
		Language: LanguageZH,
		Content:  "告警：{{.Station}} - {{.Device}} - {{.Message}}",
		Variables: []TemplateVariable{
			{Name: "Station", Description: "站点", Required: true},
			{Name: "Device", Description: "设备", Required: true},
			{Name: "Message", Description: "消息", Required: true},
		},
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := mgr.Create(ctx, tmpl); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// 测试多语言
	tmplEN := &NotificationTemplate{
		ID:       "test-template_en-US",
		Name:     "Test Template",
		Type:     TemplateTypeSMS,
		Channel:  ChannelSMS,
		Language: LanguageEN,
		Content:  "Alert: {{.Station}} - {{.Device}} - {{.Message}}",
		Variables: []TemplateVariable{
			{Name: "Station", Description: "Station", Required: true},
			{Name: "Device", Description: "Device", Required: true},
			{Name: "Message", Description: "Message", Required: true},
		},
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := mgr.Create(ctx, tmplEN); err != nil {
		t.Fatalf("Failed to create EN template: %v", err)
	}

	// 渲染模板
	data := map[string]interface{}{
		"Station": "光伏电站A",
		"Device":  "逆变器01",
		"Message": "温度过高",
	}

	rendered, err := mgr.RenderWithLanguage("test-template", LanguageZH, data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	expected := "告警：光伏电站A - 逆变器01 - 温度过高"
	if rendered != expected {
		t.Errorf("Expected %s, got %s", expected, rendered)
	}

	renderedEN, err := mgr.RenderWithLanguage("test-template", LanguageEN, data)
	if err != nil {
		t.Fatalf("Failed to render EN template: %v", err)
	}

	expectedEN := "Alert: 光伏电站A - 逆变器01 - 温度过高"
	if renderedEN != expectedEN {
		t.Errorf("Expected %s, got %s", expectedEN, renderedEN)
	}
}

// TestRateLimitingIntegration 限流集成测试
func TestRateLimitingIntegration(t *testing.T) {
	// 创建限流器
	limiter := NewTokenBucketRateLimiter(10, 5)

	// 测试突发
	successCount := 0
	for i := 0; i < 10; i++ {
		if limiter.Allow("test-key") {
			successCount++
		}
	}

	if successCount != 5 {
		t.Errorf("Expected 5 successful requests, got %d", successCount)
	}

	// 测试等待
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	if err := limiter.Wait(ctx, "test-key"); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}
	elapsed := time.Since(start)

	// 应该等待至少1秒（10/分钟 = 1/6秒，但突发后需要等待）
	if elapsed < 100*time.Millisecond {
		t.Errorf("Wait time too short: %v", elapsed)
	}

	// 测试重置
	limiter.Reset("test-key")
	if !limiter.Allow("test-key") {
		t.Error("Expected allow after reset")
	}
}

// TestSilencePeriodIntegration 静默期集成测试
func TestSilencePeriodIntegration(t *testing.T) {
	checker := NewMemorySilenceChecker()
	alarmID := "test-alarm-001"

	// 初始不在静默期
	if checker.IsSilent(alarmID) {
		t.Error("Should not be in silence period initially")
	}

	// 开始静默期
	if err := checker.StartSilence(alarmID, 2*time.Second); err != nil {
		t.Fatalf("Failed to start silence: %v", err)
	}

	// 检查静默期
	if !checker.IsSilent(alarmID) {
		t.Error("Should be in silence period")
	}

	// 等待静默期结束
	time.Sleep(3 * time.Second)

	// 检查静默期已结束
	if checker.IsSilent(alarmID) {
		t.Error("Should not be in silence period after duration")
	}

	// 手动结束静默期
	checker.StartSilence(alarmID, 1*time.Hour)
	if !checker.IsSilent(alarmID) {
		t.Error("Should be in silence period")
	}

	if err := checker.EndSilence(alarmID); err != nil {
		t.Fatalf("Failed to end silence: %v", err)
	}

	if checker.IsSilent(alarmID) {
		t.Error("Should not be in silence period after end")
	}
}
