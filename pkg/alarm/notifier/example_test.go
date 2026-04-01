package notifier

import (
	"context"
	"fmt"
	"time"
)

// ExampleUsage 使用示例
func ExampleUsage() {
	// 1. 创建模板管理器
	templateStore := NewMemoryTemplateStore()
	templateMgr := NewTemplateManager(templateStore)

	// 初始化内置模板
	InitBuiltInTemplates(templateStore)

	// 2. 创建通知存储
	notificationStore := NewMemoryNotificationStore()

	// 3. 创建静默期检查器
	silenceChecker := NewMemorySilenceChecker()

	// 4. 创建调度器
	config := &SchedulerConfig{
		QueueSize:    1000,
		WorkerCount:  10,
		MaxRetries:   3,
		RetryDelay:   5 * time.Second,
		RetryBackoff: 2.0,
		SendTimeout:  30 * time.Second,
	}

	scheduler := NewNotificationScheduler(
		notificationStore,
		nil, // logger
		templateMgr,
		silenceChecker,
		config,
	)

	// 5. 创建并注册通知器

	// 短信通知器
	smsConfig := &NotificationConfig{
		Enabled:    true,
		Channel:    ChannelSMS,
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: 5 * time.Second,
		RateLimit:  100,
		BurstLimit: 20,
		SMSConfig: &SMSConfig{
			Provider:     "aliyun",
			AccessKey:    "your-access-key",
			AccessSecret: "your-access-secret",
			SignName:     "新能源监控",
			Region:       "cn-hangzhou",
		},
	}
	smsNotifier, _ := NewSMSNotifier(smsConfig, templateMgr)
	scheduler.RegisterNotifier(ChannelSMS, smsNotifier)

	// 邮件通知器
	emailConfig := &NotificationConfig{
		Enabled:    true,
		Channel:    ChannelEmail,
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: 5 * time.Second,
		RateLimit:  50,
		BurstLimit: 10,
		EmailConfig: &EmailConfig{
			SMTPHost:    "smtp.example.com",
			SMTPPort:    465,
			Username:    "noreply@example.com",
			Password:    "password",
			FromName:    "新能源监控系统",
			FromAddress: "noreply@example.com",
			UseTLS:      true,
		},
	}
	emailNotifier, _ := NewEmailNotifier(emailConfig, templateMgr)
	scheduler.RegisterNotifier(ChannelEmail, emailNotifier)

	// 系统内消息通知器
	wsHub := NewWebSocketHub()
	go wsHub.Run()

	unreadStore := NewMemoryUnreadMessageStore()
	unreadMgr := NewUnreadMessageManager(unreadStore)

	internalConfig := &NotificationConfig{
		Enabled:    true,
		Channel:    ChannelInternal,
		Timeout:    10 * time.Second,
		MaxRetries: 2,
		RetryDelay: 1 * time.Second,
		RateLimit:  200,
		BurstLimit: 50,
	}

	// internalNotifier, _ := NewInternalNotifier(internalConfig, wsHub, unreadMgr, templateMgr)
	// scheduler.RegisterNotifier(ChannelInternal, internalNotifier)

	// 6. 启动调度器
	scheduler.Start()
	defer scheduler.Stop()

	// 7. 发送通知
	ctx := context.Background()

	// 使用构建器创建通知
	notification := NewNotificationBuilder().
		WithAlarmID("alarm-001").
		WithChannel(ChannelSMS).
		WithPriority(PriorityCritical).
		WithTemplate("alarm_critical", map[string]interface{}{
			"StationName": "光伏电站A",
			"DeviceName":  "逆变器01",
			"Message":     "设备离线",
			"TriggerTime": time.Now().Format("2006-01-02 15:04:05"),
		}).
		AddRecipient(Recipient{
			UserID: "user-001",
			Name:   "张三",
			Phone:  "13800138000",
		}).
		Build()

	// 调度通知
	if err := scheduler.Schedule(ctx, notification); err != nil {
		fmt.Printf("Failed to schedule notification: %v\n", err)
		return
	}

	fmt.Println("Notification scheduled successfully")

	// 8. 设置静默期（避免重复告警）
	silenceChecker.StartSilence("alarm-001", 30*time.Minute)

	// 9. 检查队列状态
	stats := scheduler.GetQueueStats()
	fmt.Printf("Queue stats: %+v\n", stats)
}

// ExampleMultiChannelNotification 多渠道通知示例
func ExampleMultiChannelNotification() {
	ctx := context.Background()
	notificationStore := NewMemoryNotificationStore()
	templateStore := NewMemoryTemplateStore()
	templateMgr := NewTemplateManager(templateStore)
	silenceChecker := NewMemorySilenceChecker()

	config := DefaultSchedulerConfig()
	scheduler := NewNotificationScheduler(notificationStore, nil, templateMgr, silenceChecker, config)
	scheduler.Start()
	defer scheduler.Stop()

	// 创建多渠道通知
	channels := []NotificationChannel{ChannelSMS, ChannelEmail, ChannelInternal}

	for _, channel := range channels {
		notification := NewNotificationBuilder().
			WithAlarmID("alarm-002").
			WithChannel(channel).
			WithPriority(PriorityHigh).
			WithSubject("设备告警").
			WithContent("光伏电站A - 逆变器01 发生故障").
			AddRecipient(Recipient{
				UserID: "user-001",
				Name:   "张三",
				Phone:  "13800138000",
				Email:  "zhangsan@example.com",
			}).
			Build()

		if err := scheduler.Schedule(ctx, notification); err != nil {
			fmt.Printf("Failed to schedule %s notification: %v\n", channel, err)
		}
	}

	fmt.Println("Multi-channel notifications scheduled")
}

// ExampleTemplateRendering 模板渲染示例
func ExampleTemplateRendering() {
	store := NewMemoryTemplateStore()
	mgr := NewTemplateManager(store)
	ctx := context.Background()

	// 创建自定义模板
	tmpl := &NotificationTemplate{
		ID:          "custom_alarm",
		Name:        "自定义告警模板",
		Type:        TemplateTypeEmail,
		Channel:     ChannelEmail,
		Language:    LanguageZH,
		Subject:     "【新能源监控】{{.Level}}级告警 - {{.StationName}}",
		Content:     "站点：{{.StationName}}\n设备：{{.DeviceName}}\n告警级别：{{.Level}}\n内容：{{.Message}}",
		HTMLContent: `<html><body><h2>{{.Level}}级告警</h2><p>站点：{{.StationName}}</p><p>设备：{{.DeviceName}}</p><p>内容：{{.Message}}</p></body></html>`,
		Variables: []TemplateVariable{
			{Name: "StationName", Description: "站点名称", Required: true},
			{Name: "DeviceName", Description: "设备名称", Required: true},
			{Name: "Level", Description: "告警级别", Required: true},
			{Name: "Message", Description: "告警消息", Required: true},
		},
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 保存模板
	if err := mgr.Create(ctx, tmpl); err != nil {
		fmt.Printf("Failed to create template: %v\n", err)
		return
	}

	// 渲染模板
	data := map[string]interface{}{
		"StationName": "光伏电站A",
		"DeviceName":  "逆变器01",
		"Level":       "严重",
		"Message":     "设备温度过高",
	}

	rendered, err := mgr.Render("custom_alarm", data)
	if err != nil {
		fmt.Printf("Failed to render template: %v\n", err)
		return
	}

	fmt.Println("Rendered content:")
	fmt.Println(rendered)

	// 预览模板（使用默认值）
	preview, err := mgr.Preview("custom_alarm", map[string]interface{}{
		"StationName": "测试站点",
	})
	if err != nil {
		fmt.Printf("Failed to preview template: %v\n", err)
		return
	}

	fmt.Println("\nPreview:")
	fmt.Println(preview)
}

// ExampleWebSocketIntegration WebSocket集成示例
func ExampleWebSocketIntegration() {
	// 创建WebSocket Hub
	hub := NewWebSocketHub()
	go hub.Run()

	// 模拟客户端连接
	client := &WebSocketClient{
		ID:     "client-001",
		UserID: "user-001",
		Send:   make(chan []byte, 10),
	}

	// 注册客户端
	hub.Register(client)

	// 模拟接收消息
	go func() {
		for msg := range client.Send {
			fmt.Printf("Received message: %s\n", string(msg))
		}
	}()

	// 发送消息
	message := &InternalMessage{
		ID:      "msg-001",
		AlarmID: "alarm-001",
		Title:   "设备告警",
		Content: "逆变器01 发生故障",
	}

	// 等待客户端注册
	time.Sleep(100 * time.Millisecond)

	if err := hub.SendToUser("user-001", message); err != nil {
		fmt.Printf("Failed to send message: %v\n", err)
		return
	}

	fmt.Println("Message sent via WebSocket")

	// 检查在线用户
	users := hub.GetOnlineUsers()
	fmt.Printf("Online users: %v\n", users)

	// 注销客户端
	hub.Unregister(client)
}

// ExampleRateLimiting 限流示例
func ExampleRateLimiting() {
	// 创建限流器：每分钟10个，突发5个
	limiter := NewTokenBucketRateLimiter(10, 5)

	// 突发发送
	successCount := 0
	for i := 0; i < 10; i++ {
		if limiter.Allow("user-001") {
			successCount++
		}
	}
	fmt.Printf("Burst test: %d succeeded (expected 5)\n", successCount)

	// 等待令牌恢复
	time.Sleep(6 * time.Second)

	// 再次尝试
	if limiter.Allow("user-001") {
		fmt.Println("After wait: succeeded")
	}

	// 重置
	limiter.Reset("user-001")
	if limiter.Allow("user-001") {
		fmt.Println("After reset: succeeded")
	}
}

// ExampleSilencePeriod 静默期示例
func ExampleSilencePeriod() {
	checker := NewMemorySilenceChecker()
	alarmID := "alarm-001"

	// 检查是否在静默期
	if !checker.IsSilent(alarmID) {
		fmt.Println("Not in silence period, can send notification")

		// 发送通知后设置静默期
		checker.StartSilence(alarmID, 30*time.Minute)
		fmt.Println("Silence period started")
	}

	// 再次检查
	if checker.IsSilent(alarmID) {
		fmt.Println("In silence period, skip notification")
	}

	// 手动结束静默期
	checker.EndSilence(alarmID)
	if !checker.IsSilent(alarmID) {
		fmt.Println("Silence period ended")
	}
}
