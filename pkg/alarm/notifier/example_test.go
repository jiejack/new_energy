package notifier

import (
	"context"
	"fmt"
	"time"
)

// Example 使用示例
func Example() {
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
