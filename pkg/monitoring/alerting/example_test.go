package alerting

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

// ExampleUsage 示例：如何使用告警系统
func ExampleUsage() {
	ctx := context.Background()

	// 1. 创建规则管理器
	metricProvider := &MockMetricProvider{}
	ruleManager := NewRuleManager(metricProvider)

	// 2. 添加系统资源告警规则
	systemRules := NewSystemResourceAlertRules()
	cpuRule := systemRules.CPUUsageRule(80.0, 5*time.Minute)
	memoryRule := systemRules.MemoryUsageRule(85.0, 5*time.Minute)
	diskRule := systemRules.DiskUsageRule(90.0, 5*time.Minute)

	ruleManager.AddRule(cpuRule)
	ruleManager.AddRule(memoryRule)
	ruleManager.AddRule(diskRule)

	// 3. 添加服务健康告警规则
	serviceRules := NewServiceHealthAlertRules()
	apiDownRule := serviceRules.ServiceDownRule("api-server")
	apiResponseTimeRule := serviceRules.ServiceResponseTimeRule("api-server", 1000)
	apiErrorRateRule := serviceRules.ServiceErrorRateRule("api-server", 5.0)

	ruleManager.AddRule(apiDownRule)
	ruleManager.AddRule(apiResponseTimeRule)
	ruleManager.AddRule(apiErrorRateRule)

	// 4. 添加业务告警规则
	businessRules := NewBusinessAlertRules()
	stationOfflineRule := businessRules.StationOfflineRule()
	powerAnomalyRule := businessRules.PowerGenerationAnomalyRule(50.0)
	deviceFaultRule := businessRules.DeviceFaultRule()

	ruleManager.AddRule(stationOfflineRule)
	ruleManager.AddRule(powerAnomalyRule)
	ruleManager.AddRule(deviceFaultRule)

	// 5. 创建告警聚合器
	aggregatorConfig := DefaultAggregatorConfig()
	aggregatorConfig.Strategy = StrategyBySource
	aggregatorConfig.WindowDuration = 5 * time.Minute
	aggregator := NewAlertAggregator(aggregatorConfig)

	// 6. 启动聚合器
	aggregator.Start(ctx)
	defer aggregator.Stop()

	// 7. 创建告警通知器
	notifier := NewAlertNotifier(nil)

	// 8. 注册通知渠道
	emailChannel := NewEmailChannel(&EmailConfig{
		SMTPHost:    "smtp.example.com",
		SMTPPort:    587,
		Username:    "alert@example.com",
		Password:    "password",
		FromName:    "告警系统",
		FromAddress: "alert@example.com",
		UseTLS:      true,
	})
	notifier.RegisterChannel(emailChannel)

	smsChannel := NewSMSChannel(&SMSConfig{
		Provider:     "aliyun",
		AccessKey:    "access_key",
		AccessSecret: "access_secret",
		SignName:     "新能源监控",
	})
	notifier.RegisterChannel(smsChannel)

	dingTalkChannel := NewDingTalkChannel(&DingTalkConfig{
		WebhookURL: "https://oapi.dingtalk.com/robot/send?access_token=xxx",
		Secret:     "secret",
	})
	notifier.RegisterChannel(dingTalkChannel)

	wechatChannel := NewWeChatChannel(&WeChatConfig{
		CorpID:  "corp_id",
		AgentID: "agent_id",
		Secret:  "secret",
	})
	notifier.RegisterChannel(wechatChannel)

	// 9. 添加通知模板
	templateManager := notifier.templateManager
	templateManager.AddTemplate(&NotificationTemplate{
		ID:              "alert_email_template",
		Name:            "告警邮件模板",
		Channel:         "email",
		SubjectTemplate: "【告警】{{.alert.Title}}",
		ContentTemplate: `告警详情：
规则名称: {{.alert.RuleName}}
告警级别: {{.alert.Severity}}
告警类别: {{.alert.Category}}
告警内容: {{.alert.Message}}
当前值: {{.alert.Value}}
阈值: {{.alert.Threshold}}
触发时间: {{.alert.TriggeredAt.Format "2006-01-02 15:04:05"}}
来源: {{.alert.Source}}
`,
	})

	// 10. 配置升级规则
	escalationEngine := notifier.escalationEngine
	escalationEngine.AddRule(&EscalationRule{
		ID:           "critical_escalation",
		Name:         "严重告警升级规则",
		AlertMatcher: map[string]string{"severity": "critical"},
		Levels: []EscalationLevel{
			{
				Level:      1,
				After:      5 * time.Minute,
				Channels:   []string{"email"},
				Recipients: []Recipient{{Email: "admin@example.com", Name: "管理员"}},
			},
			{
				Level:      2,
				After:      15 * time.Minute,
				Channels:   []string{"email", "sms"},
				Recipients: []Recipient{{Email: "manager@example.com", Name: "经理", Phone: "13800138000"}},
			},
			{
				Level:      3,
				After:      30 * time.Minute,
				Channels:   []string{"email", "sms", "dingtalk"},
				Recipients: []Recipient{{Email: "director@example.com", Name: "总监", Phone: "13900139000"}},
			},
		},
		Enabled: true,
	})

	// 11. 配置限流
	rateLimiter := notifier.rateLimiter
	rateLimiter.SetLimit("email", 100, 1000)  // 每分钟100封，每小时1000封
	rateLimiter.SetLimit("sms", 50, 500)      // 每分钟50条，每小时500条

	// 12. 模拟告警流程
	// 创建告警
	alert := &AlertInstance{
		ID:          "alert_001",
		RuleID:      cpuRule.ID,
		RuleName:    cpuRule.Name,
		Category:    CategorySystemResource,
		Severity:    SeverityWarning,
		Title:       "CPU使用率告警",
		Message:     "服务器 server-1 的CPU使用率达到85.5%，超过阈值80%",
		Value:       85.5,
		Threshold:   80.0,
		TriggeredAt: time.Now(),
		Source:      "server-1",
		Labels: map[string]string{
			"hostname": "server-1",
			"env":      "production",
		},
	}

	// 聚合告警
	group, isNew, err := aggregator.Aggregate(ctx, alert)
	if err != nil {
		log.Printf("聚合告警失败: %v", err)
	} else {
		if isNew {
			fmt.Printf("创建新的告警分组: %s\n", group.GroupKey)
		} else {
			fmt.Printf("告警添加到现有分组: %s (总数: %d)\n", group.GroupKey, group.Count)
		}
	}

	// 发送通知
	recipients := []Recipient{
		{Email: "admin@example.com", Name: "管理员", Phone: "13800138000"},
		{Email: "ops@example.com", Name: "运维人员"},
	}

	err = notifier.NotifyAlert(ctx, alert, []string{"email", "sms"}, recipients)
	if err != nil {
		log.Printf("发送通知失败: %v", err)
	} else {
		fmt.Println("通知发送成功")
	}

	// 13. 创建静默规则
	silenceManager := aggregator.silenceManager
	silenceManager.AddSilence(&Silence{
		ID:        "silence_001",
		Matchers:  map[string]string{"hostname": "server-2"},
		StartTime: time.Now(),
		EndTime:   time.Now().Add(2 * time.Hour),
		Reason:    "计划维护",
		CreatedBy: "admin",
	})

	// 14. 创建抑制规则
	suppressionEngine := aggregator.suppressionEngine
	suppressionEngine.AddRule(&SuppressionRule{
		ID:             "suppression_001",
		Name:           "抑制低优先级告警",
		SourceMatchers: map[string]string{"severity": "critical"},
		TargetMatchers: map[string]string{"severity": "info"},
		Enabled:        true,
	})

	// 15. 评估规则
	results, err := ruleManager.EvaluateAll(ctx)
	if err != nil {
		log.Printf("评估规则失败: %v", err)
	} else {
		for _, result := range results {
			if result.Triggered {
				fmt.Printf("规则 %s 触发: 值=%.2f, 阈值=%.2f\n",
					result.RuleName, result.Value, result.Threshold)
			}
		}
	}

	// 16. 获取统计信息
	stats := aggregator.GetStats()
	fmt.Printf("聚合器统计: 分组数=%d, 告警数=%d, 平均分组大小=%.2f\n",
		stats.TotalGroups, stats.TotalAlerts, stats.AvgGroupSize)

	// 保持运行
	fmt.Println("告警系统运行中...")
	time.Sleep(1 * time.Minute)
}

// ExampleAPIUsage 示例：如何使用告警API
func ExampleAPIUsage() {
	// 创建组件
	metricProvider := &MockMetricProvider{}
	ruleManager := NewRuleManager(metricProvider)
	aggregator := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)

	// 创建API
	api := NewAlertAPI(ruleManager, aggregator, notifier, nil)

	// 注册路由
	mux := http.NewServeMux()
	api.RegisterRoutes(mux)

	// 启动HTTP服务器
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	fmt.Println("告警API服务启动在 :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// ExampleCustomRule 示例：创建自定义告警规则
func ExampleCustomRule() {
	// 创建自定义告警规则
	customRule := &AlertRule{
		ID:          "custom_temperature_alert",
		Name:        "设备温度告警",
		Description: "设备温度超过阈值",
		Category:    CategoryBusiness,
		Severity:    SeverityWarning,
		Enabled:     true,
		Condition: AlertCondition{
			MetricName:       "device_temperature_celsius",
			Operator:         OpGT,
			Threshold:        60.0,
			Duration:         3 * time.Minute,
			Aggregation:      AggMax,
			AggregationWindow: 1 * time.Minute,
			LabelFilters: map[string]string{
				"device_type": "inverter",
			},
		},
		NotifyChannels: []string{"email", "dingtalk"},
		NotifyTemplate: "alert_email_template",
		SilenceDuration: 10 * time.Minute,
		Tags: map[string]string{
			"type": "temperature",
			"device": "inverter",
		},
	}

	// 添加到规则管理器
	metricProvider := &MockMetricProvider{}
	ruleManager := NewRuleManager(metricProvider)
	ruleManager.AddRule(customRule)

	fmt.Printf("自定义规则已添加: %s\n", customRule.Name)
}
