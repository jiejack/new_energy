package notifier

import (
	"context"
	"net/mail"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationBuilder_AllFields(t *testing.T) {
	notification := NewNotificationBuilder().
		WithID("test-001").
		WithAlarmID("alarm-001").
		WithChannel(ChannelEmail).
		WithPriority(PriorityCritical).
		WithSubject("测试通知").
		WithContent("测试内容").
		WithHTMLContent("<p>HTML内容</p>").
		WithTemplate("tmpl-001", map[string]interface{}{"key": "value"}).
		AddRecipient(Recipient{UserID: "user-001", Name: "张三", Phone: "13800138000", Email: "test@example.com"}).
		AddTag("env", "prod").
		AddAttachment(Attachment{Name: "file.txt", Content: []byte("hello"), MimeType: "text/plain"}).
		Build()

	assert.Equal(t, "test-001", notification.ID)
	assert.Equal(t, "alarm-001", notification.AlarmID)
	assert.Equal(t, ChannelEmail, notification.Channel)
	assert.Equal(t, PriorityCritical, notification.Priority)
	assert.Equal(t, "测试通知", notification.Subject)
	assert.Equal(t, "测试内容", notification.Content)
	assert.Equal(t, "<p>HTML内容</p>", notification.HTMLContent)
	assert.Equal(t, "tmpl-001", notification.TemplateID)
	assert.Len(t, notification.Recipients, 1)
	assert.Equal(t, "prod", notification.Tags["env"])
	assert.Len(t, notification.Attachments, 1)
}

func TestNotificationBuilder_Defaults(t *testing.T) {
	notification := NewNotificationBuilder().Build()
	assert.NotNil(t, notification)
	assert.NotNil(t, notification.Recipients)
	assert.NotNil(t, notification.Tags)
	assert.NotNil(t, notification.TemplateData)
	assert.NotNil(t, notification.Attachments)
	assert.False(t, notification.CreatedAt.IsZero())
}

func TestTokenBucketRateLimiter_Basic(t *testing.T) {
	limiter := NewTokenBucketRateLimiter(10, 3)
	assert.True(t, limiter.Allow("key1"))
	assert.True(t, limiter.Allow("key1"))
	assert.True(t, limiter.Allow("key1"))
	assert.False(t, limiter.Allow("key1"))
}

func TestTokenBucketRateLimiter_DifferentKeys(t *testing.T) {
	limiter := NewTokenBucketRateLimiter(10, 2)
	assert.True(t, limiter.Allow("key1"))
	assert.True(t, limiter.Allow("key1"))
	assert.False(t, limiter.Allow("key1"))
	assert.True(t, limiter.Allow("key2"))
}

func TestTokenBucketRateLimiter_Reset(t *testing.T) {
	limiter := NewTokenBucketRateLimiter(10, 2)
	limiter.Allow("key1")
	limiter.Allow("key1")
	limiter.Reset("key1")
	assert.True(t, limiter.Allow("key1"))
}

func TestTokenBucketRateLimiter_Wait(t *testing.T) {
	limiter := NewTokenBucketRateLimiter(1000, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := limiter.Wait(ctx, "wait-key")
	assert.NoError(t, err)
}

func TestTokenBucketRateLimiter_Wait_Cancelled(t *testing.T) {
	limiter := NewTokenBucketRateLimiter(1, 0)
	limiter.Allow("wait-key2")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := limiter.Wait(ctx, "wait-key2")
	assert.Error(t, err)
}

func TestMemoryNotificationStore_All(t *testing.T) {
	store := NewMemoryNotificationStore()
	ctx := context.Background()

	n1 := &Notification{ID: "n1", AlarmID: "a1", Channel: ChannelSMS, Priority: PriorityHigh, Status: StatusPending, Subject: "test1"}
	n2 := &Notification{ID: "n2", AlarmID: "a1", Channel: ChannelEmail, Priority: PriorityNormal, Status: StatusPending, Subject: "test2"}
	n3 := &Notification{ID: "n3", AlarmID: "a2", Channel: ChannelSMS, Priority: PriorityLow, Status: StatusSent, Subject: "test3"}

	require.NoError(t, store.Save(ctx, n1))
	require.NoError(t, store.Save(ctx, n2))
	require.NoError(t, store.Save(ctx, n3))

	got, err := store.Get(ctx, "n1")
	require.NoError(t, err)
	assert.Equal(t, "n1", got.ID)

	_, err = store.Get(ctx, "nonexistent")
	assert.Error(t, err)

	byAlarm, err := store.GetByAlarmID(ctx, "a1")
	require.NoError(t, err)
	assert.Len(t, byAlarm, 2)

	pending, err := store.GetPending(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, pending, 2)

	byStatus, err := store.GetByStatus(ctx, StatusSent, 10)
	require.NoError(t, err)
	assert.Len(t, byStatus, 1)

	n1.Status = StatusSent
	require.NoError(t, store.Update(ctx, n1))

	require.NoError(t, store.Delete(ctx, "n3"))
	_, err = store.Get(ctx, "n3")
	assert.Error(t, err)
}

func TestMemorySilenceChecker_Expired(t *testing.T) {
	checker := NewMemorySilenceChecker()
	assert.False(t, checker.IsSilent("alarm-001"))
	require.NoError(t, checker.StartSilence("alarm-001", 100*time.Millisecond))
	assert.True(t, checker.IsSilent("alarm-001"))
	time.Sleep(150 * time.Millisecond)
	assert.False(t, checker.IsSilent("alarm-001"))
}

func TestMemorySilenceChecker_EndNonExistent(t *testing.T) {
	checker := NewMemorySilenceChecker()
	err := checker.EndSilence("nonexistent")
	assert.NoError(t, err)
}

func TestTemplateManager_CRUD(t *testing.T) {
	store := NewMemoryTemplateStore()
	mgr := NewTemplateManager(store)
	ctx := context.Background()

	tmpl := &NotificationTemplate{
		ID:       "tmpl-001",
		Name:     "测试模板",
		Type:     TemplateTypeSMS,
		Channel:  ChannelSMS,
		Language: LanguageZH,
		Content:  "告警：{{.Message}}",
		Variables: []TemplateVariable{
			{Name: "Message", Description: "消息", Required: true},
		},
		Enabled: true,
	}

	require.NoError(t, mgr.Create(ctx, tmpl))

	got, err := mgr.GetTemplate(ctx, "tmpl-001")
	require.NoError(t, err)
	assert.Equal(t, "tmpl-001", got.ID)

	content, err := mgr.Get(ctx, "tmpl-001")
	require.NoError(t, err)
	assert.Equal(t, "告警：{{.Message}}", content)

	tmpl.Content = "更新：{{.Message}}"
	require.NoError(t, mgr.Update(ctx, tmpl))

	require.NoError(t, mgr.Delete(ctx, "tmpl-001"))
	_, err = mgr.GetTemplate(ctx, "tmpl-001")
	assert.Error(t, err)
}

func TestTemplateManager_Render(t *testing.T) {
	store := NewMemoryTemplateStore()
	mgr := NewTemplateManager(store)
	ctx := context.Background()

	tmpl := &NotificationTemplate{
		ID:       "render-tmpl",
		Name:     "渲染测试",
		Type:     TemplateTypeSMS,
		Channel:  ChannelSMS,
		Language: LanguageZH,
		Content:  "告警：{{.Message}}，时间：{{.Time}}",
		Variables: []TemplateVariable{
			{Name: "Message", Description: "消息", Required: true},
			{Name: "Time", Description: "时间", Required: true},
		},
		Enabled: true,
	}
	require.NoError(t, mgr.Create(ctx, tmpl))

	rendered, err := mgr.Render("render-tmpl", map[string]interface{}{
		"Message": "设备故障",
		"Time":    "2024-01-01",
	})
	require.NoError(t, err)
	assert.Equal(t, "告警：设备故障，时间：2024-01-01", rendered)
}

func TestTemplateManager_RenderWithLanguage_CachingBehavior(t *testing.T) {
	store := NewMemoryTemplateStore()
	mgr := NewTemplateManager(store)
	ctx := context.Background()

	tmplZH := &NotificationTemplate{
		ID:       "lang-tmpl",
		Name:     "中文模板",
		Type:     TemplateTypeSMS,
		Channel:  ChannelSMS,
		Language: LanguageZH,
		Content:  "中文：{{.Message}}",
		Variables: []TemplateVariable{
			{Name: "Message", Description: "消息", Required: true},
		},
		Enabled: true,
	}
	require.NoError(t, mgr.Create(ctx, tmplZH))

	rendered, err := mgr.RenderWithLanguage("lang-tmpl", LanguageZH, map[string]interface{}{"Message": "测试"})
	require.NoError(t, err)
	assert.Contains(t, rendered, "中文")

	rendered2, err := mgr.RenderWithLanguage("lang-tmpl", LanguageEN, map[string]interface{}{"Message": "test"})
	require.NoError(t, err)
	assert.Contains(t, rendered2, "中文", "RenderWithLanguage returns cached template regardless of language parameter")
}

func TestTemplateManager_RenderWithLanguage_NotFound(t *testing.T) {
	store := NewMemoryTemplateStore()
	mgr := NewTemplateManager(store)
	_, err := mgr.RenderWithLanguage("nonexistent", LanguageZH, map[string]interface{}{"Message": "test"})
	assert.Error(t, err)
}

func TestTemplateManager_Preview(t *testing.T) {
	store := NewMemoryTemplateStore()
	mgr := NewTemplateManager(store)
	ctx := context.Background()

	tmpl := &NotificationTemplate{
		ID:       "preview-tmpl",
		Name:     "预览测试",
		Type:     TemplateTypeSMS,
		Channel:  ChannelSMS,
		Language: LanguageZH,
		Content:  "告警：{{.Message}}，级别：{{.Level}}",
		Variables: []TemplateVariable{
			{Name: "Message", Description: "消息", Required: true},
			{Name: "Level", Description: "级别", Required: false, Default: "高"},
		},
		Enabled: true,
	}
	require.NoError(t, mgr.Create(ctx, tmpl))

	preview, err := mgr.Preview("preview-tmpl", map[string]interface{}{"Message": "测试"})
	require.NoError(t, err)
	assert.Contains(t, preview, "测试")
	assert.Contains(t, preview, "高")
}

func TestTemplateManager_Preview_NotInCache(t *testing.T) {
	store := NewMemoryTemplateStore()
	mgr := NewTemplateManager(store)
	_, err := mgr.Preview("nonexistent", map[string]interface{}{})
	assert.Error(t, err)
}

func TestTemplateManager_Validate(t *testing.T) {
	store := NewMemoryTemplateStore()
	mgr := NewTemplateManager(store)
	ctx := context.Background()

	tmpl := &NotificationTemplate{
		ID:       "validate-tmpl",
		Name:     "验证测试",
		Type:     TemplateTypeSMS,
		Channel:  ChannelSMS,
		Language: LanguageZH,
		Content:  "{{.A}} {{.B}}",
		Variables: []TemplateVariable{
			{Name: "A", Description: "A", Required: true},
			{Name: "B", Description: "B", Required: false},
		},
		Enabled: true,
	}
	require.NoError(t, mgr.Create(ctx, tmpl))

	assert.NoError(t, mgr.Validate("validate-tmpl", map[string]interface{}{"A": "1", "B": "2"}))
	assert.NoError(t, mgr.Validate("validate-tmpl", map[string]interface{}{"A": "1"}))
	assert.Error(t, mgr.Validate("validate-tmpl", map[string]interface{}{}))
	assert.Error(t, mgr.Validate("nonexistent", map[string]interface{}{}))
}

func TestTemplateManager_validateTemplate_Invalid(t *testing.T) {
	store := NewMemoryTemplateStore()
	mgr := NewTemplateManager(store)
	ctx := context.Background()

	assert.Error(t, mgr.Create(ctx, &NotificationTemplate{ID: "", Name: "test", Content: "content"}))
	assert.Error(t, mgr.Create(ctx, &NotificationTemplate{ID: "id", Name: "", Content: "content"}))
	assert.Error(t, mgr.Create(ctx, &NotificationTemplate{ID: "id", Name: "test", Content: ""}))
	assert.Error(t, mgr.Create(ctx, &NotificationTemplate{ID: "id", Name: "test", Content: "{{.Invalid"}))
}

func TestTemplateManager_validateTemplate_HTMLInvalid(t *testing.T) {
	store := NewMemoryTemplateStore()
	mgr := NewTemplateManager(store)
	ctx := context.Background()

	assert.Error(t, mgr.Create(ctx, &NotificationTemplate{
		ID:          "id",
		Name:        "test",
		Content:     "ok",
		HTMLContent: "{{.Invalid",
	}))
}

func TestMemoryTemplateStore_List(t *testing.T) {
	store := NewMemoryTemplateStore()
	ctx := context.Background()

	store.Save(ctx, &NotificationTemplate{ID: "1", Type: TemplateTypeSMS, Channel: ChannelSMS, Language: LanguageZH, Enabled: true})
	store.Save(ctx, &NotificationTemplate{ID: "2", Type: TemplateTypeEmail, Channel: ChannelEmail, Language: LanguageEN, Enabled: false})
	store.Save(ctx, &NotificationTemplate{ID: "3", Type: TemplateTypeSMS, Channel: ChannelSMS, Language: LanguageEN, Enabled: true, Tags: []string{"critical"}})

	results, total, err := store.List(ctx, &TemplateQuery{Type: TemplateTypeSMS})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, results, 2)

	results, total, err = store.List(ctx, &TemplateQuery{Channel: ChannelEmail})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)

	results, total, err = store.List(ctx, &TemplateQuery{Language: LanguageEN})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)

	results, total, err = store.List(ctx, &TemplateQuery{Enabled: boolPtr(true)})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)

	results, total, err = store.List(ctx, &TemplateQuery{Tags: []string{"critical"}})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
}

func TestMemoryTemplateStore_List_Pagination(t *testing.T) {
	store := NewMemoryTemplateStore()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		store.Save(ctx, &NotificationTemplate{ID: string(rune('a' + i)), Type: TemplateTypeSMS, Channel: ChannelSMS, Language: LanguageZH, Enabled: true})
	}

	results, total, err := store.List(ctx, &TemplateQuery{Page: 1, PageSize: 2})
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, results, 2)

	results, total, err = store.List(ctx, &TemplateQuery{Page: 3, PageSize: 2})
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, results, 1)

	results, total, err = store.List(ctx, &TemplateQuery{Page: 10, PageSize: 2})
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, results, 0)
}

func TestMemoryTemplateStore_GetByIDAndLanguage(t *testing.T) {
	store := NewMemoryTemplateStore()
	ctx := context.Background()

	store.Save(ctx, &NotificationTemplate{ID: "tmpl", Language: LanguageZH, Content: "中文"})
	store.Save(ctx, &NotificationTemplate{ID: "tmpl_en-US", Language: LanguageEN, Content: "English"})

	tmpl, err := store.GetByIDAndLanguage(ctx, "tmpl", LanguageEN)
	require.NoError(t, err)
	assert.Equal(t, "English", tmpl.Content)

	tmpl, err = store.GetByIDAndLanguage(ctx, "tmpl", LanguageZH)
	require.NoError(t, err)
	assert.Equal(t, "中文", tmpl.Content)

	_, err = store.GetByIDAndLanguage(ctx, "nonexistent", LanguageZH)
	assert.Error(t, err)
}

func TestMemoryTemplateStore_Delete(t *testing.T) {
	store := NewMemoryTemplateStore()
	ctx := context.Background()

	store.Save(ctx, &NotificationTemplate{ID: "del-tmpl", Content: "test"})
	require.NoError(t, store.Delete(ctx, "del-tmpl"))
	_, err := store.Get(ctx, "del-tmpl")
	assert.Error(t, err)
}

func TestInitBuiltInTemplates(t *testing.T) {
	store := NewMemoryTemplateStore()
	err := InitBuiltInTemplates(store)
	require.NoError(t, err)

	ctx := context.Background()
	tmpl, err := store.Get(ctx, "alarm_critical")
	require.NoError(t, err)
	assert.Equal(t, "严重告警通知", tmpl.Name)

	tmpl, err = store.Get(ctx, "alarm_email")
	require.NoError(t, err)
	assert.Equal(t, "告警邮件通知", tmpl.Name)

	tmpl, err = store.Get(ctx, "alarm_internal")
	require.NoError(t, err)
	assert.Equal(t, "系统内告警通知", tmpl.Name)
}

func TestSMSNotifier_New(t *testing.T) {
	_, err := NewSMSNotifier(nil, nil)
	assert.Equal(t, ErrSMSConfigInvalid, err)

	_, err = NewSMSNotifier(&NotificationConfig{}, nil)
	assert.Equal(t, ErrSMSConfigInvalid, err)

	notifier, err := NewSMSNotifier(&NotificationConfig{
		SMSConfig:  &SMSConfig{Provider: "aliyun", AccessKey: "key", AccessSecret: "secret"},
		RateLimit:  100,
		BurstLimit: 10,
		Timeout:    5 * time.Second,
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, ChannelSMS, notifier.Channel())
}

func TestSMSNotifier_Validate(t *testing.T) {
	notifier, _ := NewSMSNotifier(&NotificationConfig{
		SMSConfig:  &SMSConfig{Provider: "aliyun"},
		RateLimit:  100,
		BurstLimit: 10,
		Timeout:    5 * time.Second,
	}, nil)

	assert.Error(t, notifier.Validate(nil))
	assert.Error(t, notifier.Validate(&Notification{}))
	assert.Error(t, notifier.Validate(&Notification{Recipients: []Recipient{{Name: "test"}}}))
	assert.Error(t, notifier.Validate(&Notification{Recipients: []Recipient{{Name: "test", Phone: "123"}}}))
	assert.NoError(t, notifier.Validate(&Notification{Recipients: []Recipient{{Name: "test", Phone: "13800138000"}}}))
}

func TestSMSNotifier_HealthCheck(t *testing.T) {
	notifier, _ := NewSMSNotifier(&NotificationConfig{
		SMSConfig:  &SMSConfig{Provider: "aliyun", AccessKey: "key", AccessSecret: "secret"},
		RateLimit:  100,
		BurstLimit: 10,
		Timeout:    5 * time.Second,
	}, nil)
	assert.NoError(t, notifier.HealthCheck(context.Background()))

	notifier2, _ := NewSMSNotifier(&NotificationConfig{
		SMSConfig:  &SMSConfig{Provider: "aliyun", AccessKey: "", AccessSecret: ""},
		RateLimit:  100,
		BurstLimit: 10,
		Timeout:    5 * time.Second,
	}, nil)
	assert.Error(t, notifier2.HealthCheck(context.Background()))
}

func TestSMSNotifier_Close(t *testing.T) {
	notifier, _ := NewSMSNotifier(&NotificationConfig{
		SMSConfig:  &SMSConfig{Provider: "aliyun"},
		RateLimit:  100,
		BurstLimit: 10,
		Timeout:    5 * time.Second,
	}, nil)
	assert.NoError(t, notifier.Close())
}

func TestSMSNotifier_Send_UnsupportedProvider(t *testing.T) {
	notifier, _ := NewSMSNotifier(&NotificationConfig{
		SMSConfig:  &SMSConfig{Provider: "unknown"},
		RateLimit:  100,
		BurstLimit: 10,
		Timeout:    5 * time.Second,
	}, nil)
	_, err := notifier.Send(context.Background(), &Notification{
		ID:         "test",
		AlarmID:    "alarm",
		Recipients: []Recipient{{Name: "test", Phone: "13800138000"}},
		Content:    "test",
	})
	assert.Error(t, err)
}

func TestSMSNotifier_SendBatch(t *testing.T) {
	notifier, _ := NewSMSNotifier(&NotificationConfig{
		SMSConfig:  &SMSConfig{Provider: "unknown"},
		RateLimit:  100,
		BurstLimit: 10,
		Timeout:    5 * time.Second,
	}, nil)
	results, err := notifier.SendBatch(context.Background(), []*Notification{
		{ID: "n1", Recipients: []Recipient{{Phone: "13800138000"}}, Content: "test"},
	})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.False(t, results[0].Success)
}

func TestSMSNotifier_Send_RateLimitExceeded(t *testing.T) {
	notifier, _ := NewSMSNotifier(&NotificationConfig{
		SMSConfig:  &SMSConfig{Provider: "aliyun"},
		RateLimit:  100,
		BurstLimit: 0,
		Timeout:    5 * time.Second,
	}, nil)
	notifier.rateLimiter.Allow("sms:alarm1")
	_, err := notifier.Send(context.Background(), &Notification{
		ID:         "test",
		AlarmID:    "alarm1",
		Recipients: []Recipient{{Phone: "13800138000"}},
		Content:    "test",
	})
	assert.Error(t, err)
}

func TestIsValidPhone(t *testing.T) {
	assert.True(t, isValidPhone("13800138000"))
	assert.True(t, isValidPhone("15012345678"))
	assert.False(t, isValidPhone("23800138000"))
	assert.False(t, isValidPhone("1380013800"))
	assert.False(t, isValidPhone("138001380001"))
	assert.False(t, isValidPhone("1380013800a"))
	assert.False(t, isValidPhone(""))
}

func TestIsValidEmail(t *testing.T) {
	assert.True(t, isValidEmail("test@example.com"))
	assert.True(t, isValidEmail("user.name@domain.org"))
	assert.False(t, isValidEmail("invalid"))
	assert.False(t, isValidEmail("@domain.com"))
	assert.False(t, isValidEmail("user@"))
	assert.False(t, isValidEmail("user@domain"))
	assert.False(t, isValidEmail(""))
}

func TestEmailNotifier_New(t *testing.T) {
	_, err := NewEmailNotifier(nil, nil)
	assert.Equal(t, ErrEmailConfigInvalid, err)

	_, err = NewEmailNotifier(&NotificationConfig{}, nil)
	assert.Equal(t, ErrEmailConfigInvalid, err)

	notifier, err := NewEmailNotifier(&NotificationConfig{
		EmailConfig: &EmailConfig{SMTPHost: "smtp.example.com", SMTPPort: 587},
		RateLimit:   100,
		BurstLimit:  10,
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, ChannelEmail, notifier.Channel())
}

func TestEmailNotifier_Validate(t *testing.T) {
	notifier, _ := NewEmailNotifier(&NotificationConfig{
		EmailConfig: &EmailConfig{SMTPHost: "smtp.example.com", SMTPPort: 587},
		RateLimit:   100,
		BurstLimit:  10,
	}, nil)

	assert.Error(t, notifier.Validate(nil))
	assert.Error(t, notifier.Validate(&Notification{}))
	assert.Error(t, notifier.Validate(&Notification{Recipients: []Recipient{{Name: "test"}}}))
	assert.Error(t, notifier.Validate(&Notification{Recipients: []Recipient{{Name: "test", Email: "invalid"}}}))
	assert.NoError(t, notifier.Validate(&Notification{Recipients: []Recipient{{Name: "test", Email: "test@example.com"}}}))
}

func TestEmailNotifier_Close(t *testing.T) {
	notifier, _ := NewEmailNotifier(&NotificationConfig{
		EmailConfig: &EmailConfig{SMTPHost: "smtp.example.com", SMTPPort: 587},
		RateLimit:   100,
		BurstLimit:  10,
	}, nil)
	assert.NoError(t, notifier.Close())
}

func TestEmailNotifier_prepareEmail(t *testing.T) {
	notifier, _ := NewEmailNotifier(&NotificationConfig{
		EmailConfig: &EmailConfig{SMTPHost: "smtp.example.com", SMTPPort: 587, FromName: "Test", FromAddress: "test@example.com"},
		RateLimit:   100,
		BurstLimit:  10,
	}, nil)

	email, err := notifier.prepareEmail(&Notification{
		Subject:    "测试",
		Content:    "内容",
		Priority:   PriorityHigh,
		AlarmID:    "alarm1",
		Recipients: []Recipient{{Name: "用户", Email: "user@example.com"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "测试", email.Subject)
	assert.Equal(t, "内容", email.TextContent)
	assert.Len(t, email.To, 1)
	assert.Equal(t, "2", email.Headers["X-Priority"])
}

func TestEmailNotifier_prepareEmail_HTMLContent(t *testing.T) {
	notifier, _ := NewEmailNotifier(&NotificationConfig{
		EmailConfig: &EmailConfig{SMTPHost: "smtp.example.com", SMTPPort: 587, FromName: "Test", FromAddress: "test@example.com"},
		RateLimit:   100,
		BurstLimit:  10,
	}, nil)

	email, err := notifier.prepareEmail(&Notification{
		Subject:     "测试",
		HTMLContent: "<p>HTML</p>",
		Priority:    PriorityCritical,
		AlarmID:     "alarm1",
		Recipients:  []Recipient{{Name: "用户", Email: "user@example.com"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "<p>HTML</p>", email.HTMLContent)
	assert.Equal(t, "1", email.Headers["X-Priority"])
}

func TestEmailNotifier_prepareEmail_WithAttachments(t *testing.T) {
	notifier, _ := NewEmailNotifier(&NotificationConfig{
		EmailConfig: &EmailConfig{SMTPHost: "smtp.example.com", SMTPPort: 587, FromName: "Test", FromAddress: "test@example.com"},
		RateLimit:   100,
		BurstLimit:  10,
	}, nil)

	email, err := notifier.prepareEmail(&Notification{
		Subject:    "测试",
		Content:    "内容",
		Priority:   PriorityNormal,
		AlarmID:    "alarm1",
		Recipients: []Recipient{{Name: "用户", Email: "user@example.com"}},
		Attachments: []Attachment{{Name: "file.txt", Content: []byte("hello"), MimeType: "text/plain"}},
	})
	require.NoError(t, err)
	assert.Len(t, email.Attachments, 1)
}

func TestEmailNotifier_getPriorityHeader(t *testing.T) {
	notifier, _ := NewEmailNotifier(&NotificationConfig{
		EmailConfig: &EmailConfig{SMTPHost: "smtp.example.com", SMTPPort: 587},
		RateLimit:   100,
		BurstLimit:  10,
	}, nil)

	assert.Equal(t, "1", notifier.getPriorityHeader(PriorityCritical))
	assert.Equal(t, "2", notifier.getPriorityHeader(PriorityHigh))
	assert.Equal(t, "3", notifier.getPriorityHeader(PriorityNormal))
	assert.Equal(t, "4", notifier.getPriorityHeader(PriorityLow))
}

func TestEmailNotifier_buildEmailContent_PlainText(t *testing.T) {
	notifier, _ := NewEmailNotifier(&NotificationConfig{
		EmailConfig: &EmailConfig{SMTPHost: "smtp.example.com", SMTPPort: 587, FromName: "Test", FromAddress: "test@example.com"},
		RateLimit:   100,
		BurstLimit:  10,
	}, nil)

	email := &EmailMessage{
		From:        mail.Address{Name: "Test", Address: "test@example.com"},
		To:          []mail.Address{{Name: "User", Address: "user@example.com"}},
		Subject:     "测试",
		TextContent: "内容",
		Headers:     map[string]string{},
	}
	content, err := notifier.buildEmailContent(email)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Content-Type: text/plain")
}

func TestEmailNotifier_buildEmailContent_HTML(t *testing.T) {
	notifier, _ := NewEmailNotifier(&NotificationConfig{
		EmailConfig: &EmailConfig{SMTPHost: "smtp.example.com", SMTPPort: 587, FromName: "Test", FromAddress: "test@example.com"},
		RateLimit:   100,
		BurstLimit:  10,
	}, nil)

	email := &EmailMessage{
		From:        mail.Address{Name: "Test", Address: "test@example.com"},
		To:          []mail.Address{{Name: "User", Address: "user@example.com"}},
		Subject:     "测试",
		HTMLContent: "<p>HTML</p>",
		Headers:     map[string]string{},
	}
	content, err := notifier.buildEmailContent(email)
	require.NoError(t, err)
	assert.Contains(t, string(content), "multipart/alternative")
}

func TestEmailNotifier_buildEmailContent_WithAttachments(t *testing.T) {
	notifier, _ := NewEmailNotifier(&NotificationConfig{
		EmailConfig: &EmailConfig{SMTPHost: "smtp.example.com", SMTPPort: 587, FromName: "Test", FromAddress: "test@example.com"},
		RateLimit:   100,
		BurstLimit:  10,
	}, nil)

	email := &EmailMessage{
		From:        mail.Address{Name: "Test", Address: "test@example.com"},
		To:          []mail.Address{{Name: "User", Address: "user@example.com"}},
		Subject:     "测试",
		TextContent: "内容",
		Headers:     map[string]string{},
		Attachments: []Attachment{{Name: "file.txt", Content: []byte("hello"), MimeType: "text/plain"}},
	}
	content, err := notifier.buildEmailContent(email)
	require.NoError(t, err)
	assert.Contains(t, string(content), "multipart/mixed")
}

func TestEmailNotifier_RenderEmailTemplate_NoTemplateMgr(t *testing.T) {
	notifier, _ := NewEmailNotifier(&NotificationConfig{
		EmailConfig: &EmailConfig{SMTPHost: "smtp.example.com", SMTPPort: 587},
		RateLimit:   100,
		BurstLimit:  10,
	}, nil)
	_, err := notifier.RenderEmailTemplate("nonexistent", map[string]interface{}{})
	assert.Error(t, err)
}

func TestBatchEmailSender(t *testing.T) {
	notifier, _ := NewEmailNotifier(&NotificationConfig{
		EmailConfig: &EmailConfig{SMTPHost: "smtp.example.com", SMTPPort: 587},
		RateLimit:   100,
		BurstLimit:  10,
	}, nil)
	sender := NewBatchEmailSender(notifier, 2, 5*time.Second)
	assert.NotNil(t, sender)

	results, err := sender.SendBatch(context.Background(), []*Notification{
		{ID: "n1", Recipients: []Recipient{{Email: "test@example.com"}}, Subject: "test"},
	})
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestInternalNotifier_New(t *testing.T) {
	_, err := NewInternalNotifier(nil, nil, nil, nil)
	assert.Equal(t, ErrInternalConfigInvalid, err)

	notifier, err := NewInternalNotifier(&NotificationConfig{
		RateLimit:  100,
		BurstLimit: 10,
	}, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, ChannelInternal, notifier.Channel())
}

func TestInternalNotifier_Validate(t *testing.T) {
	notifier, _ := NewInternalNotifier(&NotificationConfig{RateLimit: 100, BurstLimit: 10}, nil, nil, nil)

	assert.Error(t, notifier.Validate(nil))
	assert.Error(t, notifier.Validate(&Notification{}))
	assert.Error(t, notifier.Validate(&Notification{Recipients: []Recipient{{Name: "test"}}}))
	assert.NoError(t, notifier.Validate(&Notification{Recipients: []Recipient{{UserID: "user-001", Name: "test"}}}))
}

func TestInternalNotifier_Send(t *testing.T) {
	notifier, _ := NewInternalNotifier(&NotificationConfig{RateLimit: 100, BurstLimit: 10}, nil, nil, nil)

	result, err := notifier.Send(context.Background(), &Notification{
		ID:         "test",
		AlarmID:    "alarm1",
		Subject:    "测试",
		Content:    "内容",
		Recipients: []Recipient{{UserID: "user-001", Name: "test"}},
	})
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, StatusSent, result.Status)
}

func TestInternalNotifier_Send_WithMessageCenter(t *testing.T) {
	hub := NewWebSocketHub()
	go hub.Run()
	defer func() {
		hub.mu.Lock()
		hub.running = false
		hub.mu.Unlock()
	}()
	time.Sleep(50 * time.Millisecond)

	mc := &mockMessageCenter{}
	notifier, _ := NewInternalNotifier(&NotificationConfig{RateLimit: 100, BurstLimit: 10}, hub, mc, nil)

	result, err := notifier.Send(context.Background(), &Notification{
		ID:         "test",
		AlarmID:    "alarm1",
		Subject:    "测试",
		Content:    "内容",
		Recipients: []Recipient{{UserID: "user-001", Name: "test"}},
	})
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestInternalNotifier_SendBatch(t *testing.T) {
	notifier, _ := NewInternalNotifier(&NotificationConfig{RateLimit: 100, BurstLimit: 10}, nil, nil, nil)

	results, err := notifier.SendBatch(context.Background(), []*Notification{
		{ID: "n1", AlarmID: "a1", Recipients: []Recipient{{UserID: "u1"}}, Content: "test"},
		{ID: "n2", AlarmID: "a2", Recipients: []Recipient{{UserID: "u2"}}, Content: "test"},
	})
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestInternalNotifier_HealthCheck(t *testing.T) {
	notifier, _ := NewInternalNotifier(&NotificationConfig{RateLimit: 100, BurstLimit: 10}, nil, nil, nil)
	assert.NoError(t, notifier.HealthCheck(context.Background()))

	hub := NewWebSocketHub()
	notifier2, _ := NewInternalNotifier(&NotificationConfig{RateLimit: 100, BurstLimit: 10}, hub, nil, nil)
	assert.Error(t, notifier2.HealthCheck(context.Background()))
}

func TestInternalNotifier_Close(t *testing.T) {
	notifier, _ := NewInternalNotifier(&NotificationConfig{RateLimit: 100, BurstLimit: 10}, nil, nil, nil)
	assert.NoError(t, notifier.Close())
}

func TestInternalNotifier_prepareMessage(t *testing.T) {
	notifier, _ := NewInternalNotifier(&NotificationConfig{RateLimit: 100, BurstLimit: 10}, nil, nil, nil)

	msg, err := notifier.prepareMessage(&Notification{
		ID:         "test",
		AlarmID:    "alarm1",
		Subject:    "标题",
		Content:    "内容",
		Priority:   PriorityHigh,
		Recipients: []Recipient{{UserID: "u1"}, {UserID: "u2"}},
		HTMLContent: "<p>HTML</p>",
		Tags:       map[string]string{"env": "prod"},
	})
	require.NoError(t, err)
	assert.Equal(t, "test", msg.ID)
	assert.Equal(t, "标题", msg.Title)
	assert.Equal(t, "<p>HTML</p>", msg.HTMLContent)
	assert.Equal(t, []string{"u1", "u2"}, msg.Recipients)
	assert.Equal(t, "prod", msg.Tags["env"])
}

func TestWebSocketHub_GetOnlineUsers(t *testing.T) {
	hub := NewWebSocketHub()
	go hub.Run()
	defer func() {
		hub.mu.Lock()
		hub.running = false
		hub.mu.Unlock()
	}()
	time.Sleep(50 * time.Millisecond)

	client1 := &WebSocketClient{ID: "c1", UserID: "user-001", Send: make(chan []byte, 10)}
	client2 := &WebSocketClient{ID: "c2", UserID: "user-002", Send: make(chan []byte, 10)}
	hub.Register(client1)
	hub.Register(client2)
	time.Sleep(100 * time.Millisecond)

	users := hub.GetOnlineUsers()
	assert.Len(t, users, 2)
}

func TestWebSocketHub_Broadcast(t *testing.T) {
	hub := NewWebSocketHub()
	go hub.Run()
	defer func() {
		hub.mu.Lock()
		hub.running = false
		hub.mu.Unlock()
	}()
	time.Sleep(50 * time.Millisecond)

	client := &WebSocketClient{ID: "c1", UserID: "user-001", Send: make(chan []byte, 10)}
	hub.Register(client)
	time.Sleep(100 * time.Millisecond)

	hub.Broadcast(&WebSocketMessage{Type: "test", Payload: "hello"})
	time.Sleep(100 * time.Millisecond)

	select {
	case msg := <-client.Send:
		assert.Contains(t, string(msg), "test")
	default:
		t.Error("expected to receive broadcast message")
	}
}

func TestWebSocketHub_SendToUser_NotConnected(t *testing.T) {
	hub := NewWebSocketHub()
	err := hub.SendToUser("nonexistent", &InternalMessage{ID: "msg1", Title: "test", Content: "test"})
	assert.Equal(t, ErrWebSocketNotConnected, err)
}

func TestNotificationScheduler_RegisterUnregister(t *testing.T) {
	store := NewMemoryNotificationStore()
	scheduler := NewNotificationScheduler(store, nil, nil, nil, &SchedulerConfig{QueueSize: 100, WorkerCount: 1})

	smsNotifier, _ := NewSMSNotifier(&NotificationConfig{
		SMSConfig:  &SMSConfig{Provider: "aliyun"},
		RateLimit:  100,
		BurstLimit: 10,
		Timeout:    5 * time.Second,
	}, nil)

	scheduler.RegisterNotifier(ChannelSMS, smsNotifier)
	scheduler.UnregisterNotifier(ChannelSMS)
}

func TestNotificationScheduler_Schedule_NotRunning(t *testing.T) {
	store := NewMemoryNotificationStore()
	scheduler := NewNotificationScheduler(store, nil, nil, nil, &SchedulerConfig{QueueSize: 100, WorkerCount: 1})

	err := scheduler.Schedule(context.Background(), &Notification{ID: "n1", AlarmID: "a1"})
	assert.Equal(t, ErrSchedulerNotRunning, err)
}

func TestNotificationScheduler_Schedule_SilencePeriod(t *testing.T) {
	store := NewMemoryNotificationStore()
	silenceChecker := NewMemorySilenceChecker()
	scheduler := NewNotificationScheduler(store, nil, nil, silenceChecker, &SchedulerConfig{QueueSize: 100, WorkerCount: 1})
	scheduler.Start()
	defer scheduler.Stop()

	silenceChecker.StartSilence("alarm1", 1*time.Hour)
	err := scheduler.Schedule(context.Background(), &Notification{ID: "n1", AlarmID: "alarm1"})
	assert.Equal(t, ErrSilencePeriodActive, err)
}

func TestNotificationScheduler_Cancel(t *testing.T) {
	store := NewMemoryNotificationStore()
	scheduler := NewNotificationScheduler(store, nil, nil, nil, &SchedulerConfig{QueueSize: 100, WorkerCount: 1})
	scheduler.Start()
	defer scheduler.Stop()

	scheduler.Schedule(context.Background(), &Notification{ID: "n1", AlarmID: "a1", Priority: PriorityNormal})
	err := scheduler.Cancel(context.Background(), "n1")
	require.NoError(t, err)

	got, _ := store.Get(context.Background(), "n1")
	assert.Equal(t, StatusCancelled, got.Status)
}

func TestNotificationScheduler_Cancel_NotFound(t *testing.T) {
	store := NewMemoryNotificationStore()
	scheduler := NewNotificationScheduler(store, nil, nil, nil, &SchedulerConfig{QueueSize: 100, WorkerCount: 1})
	err := scheduler.Cancel(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestNotificationScheduler_ScheduleBatch(t *testing.T) {
	store := NewMemoryNotificationStore()
	scheduler := NewNotificationScheduler(store, nil, nil, nil, &SchedulerConfig{QueueSize: 100, WorkerCount: 1})
	scheduler.Start()
	defer scheduler.Stop()

	err := scheduler.ScheduleBatch(context.Background(), []*Notification{
		{ID: "n1", AlarmID: "a1", Priority: PriorityNormal},
		{ID: "n2", AlarmID: "a2", Priority: PriorityHigh},
	})
	require.NoError(t, err)
}

func TestNotificationScheduler_calculateRetryDelay(t *testing.T) {
	scheduler := NewNotificationScheduler(nil, nil, nil, nil, &SchedulerConfig{
		RetryDelay:   5 * time.Second,
		RetryBackoff: 2.0,
	})

	delay1 := scheduler.calculateRetryDelay(1)
	assert.Equal(t, 5*time.Second, delay1)

	delay2 := scheduler.calculateRetryDelay(2)
	assert.Equal(t, 10*time.Second, delay2)

	delay3 := scheduler.calculateRetryDelay(3)
	assert.Equal(t, 20*time.Second, delay3)
}

func TestNotificationScheduler_DefaultConfig(t *testing.T) {
	config := DefaultSchedulerConfig()
	assert.Equal(t, 1000, config.QueueSize)
	assert.Equal(t, 10, config.WorkerCount)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 2.0, config.RetryBackoff)
}

func TestNotificationScheduler_NilConfig(t *testing.T) {
	store := NewMemoryNotificationStore()
	scheduler := NewNotificationScheduler(store, nil, nil, nil, nil)
	assert.NotNil(t, scheduler)
	assert.Equal(t, 1000, scheduler.queueSize)
}

func TestNotificationScheduler_StartTwice(t *testing.T) {
	store := NewMemoryNotificationStore()
	scheduler := NewNotificationScheduler(store, nil, nil, nil, &SchedulerConfig{QueueSize: 100, WorkerCount: 1})
	require.NoError(t, scheduler.Start())
	defer scheduler.Stop()
	require.NoError(t, scheduler.Start())
}

func TestNotificationScheduler_StopTwice(t *testing.T) {
	store := NewMemoryNotificationStore()
	scheduler := NewNotificationScheduler(store, nil, nil, nil, &SchedulerConfig{QueueSize: 100, WorkerCount: 1})
	require.NoError(t, scheduler.Start())
	require.NoError(t, scheduler.Stop())
	require.NoError(t, scheduler.Stop())
}

func TestHmacSha256(t *testing.T) {
	result := hmacSha256([]byte("key"), "data")
	assert.NotNil(t, result)
	assert.Len(t, result, 32)
}

func TestHmacSha256Hex(t *testing.T) {
	result := hmacSha256Hex([]byte("key"), "data")
	assert.NotEmpty(t, result)
}

func boolPtr(b bool) *bool {
	return &b
}

type mockMessageCenter struct{}

func (m *mockMessageCenter) Save(ctx context.Context, message *InternalMessage) error               { return nil }
func (m *mockMessageCenter) Get(ctx context.Context, id string) (*InternalMessage, error)            { return nil, nil }
func (m *mockMessageCenter) GetByUser(ctx context.Context, userID string, unreadOnly bool, page, pageSize int) ([]*InternalMessage, int64, error) {
	return nil, 0, nil
}
func (m *mockMessageCenter) MarkAsRead(ctx context.Context, messageID string, userID string) error   { return nil }
func (m *mockMessageCenter) MarkAllAsRead(ctx context.Context, userID string) error                  { return nil }
func (m *mockMessageCenter) Delete(ctx context.Context, messageID string, userID string) error       { return nil }
func (m *mockMessageCenter) GetUnreadCount(ctx context.Context, userID string) (int64, error)        { return 0, nil }

func TestUnreadMessageManager_All(t *testing.T) {
	store := NewMemoryUnreadMessageStore()
	mgr := NewUnreadMessageManager(store)
	ctx := context.Background()

	count, err := mgr.Get(ctx, "user-001")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	require.NoError(t, mgr.Increment(ctx, "user-001"))
	require.NoError(t, mgr.Increment(ctx, "user-001"))
	require.NoError(t, mgr.Increment(ctx, "user-001"))

	count, err = mgr.Get(ctx, "user-001")
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	require.NoError(t, mgr.Decrement(ctx, "user-001"))
	count, err = mgr.Get(ctx, "user-001")
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	require.NoError(t, mgr.Reset(ctx, "user-001"))
	count, err = mgr.Get(ctx, "user-001")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestMemoryUnreadMessageStore(t *testing.T) {
	store := NewMemoryUnreadMessageStore()
	ctx := context.Background()

	count, err := store.Get(ctx, "user-001")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	require.NoError(t, store.Increment(ctx, "user-001"))
	require.NoError(t, store.Increment(ctx, "user-001"))
	count, _ = store.Get(ctx, "user-001")
	assert.Equal(t, int64(2), count)

	require.NoError(t, store.Decrement(ctx, "user-001"))
	count, _ = store.Get(ctx, "user-001")
	assert.Equal(t, int64(1), count)

	require.NoError(t, store.Reset(ctx, "user-001"))
	count, _ = store.Get(ctx, "user-001")
	assert.Equal(t, int64(0), count)
}

func TestMemoryUnreadMessageStore_DecrementZero(t *testing.T) {
	store := NewMemoryUnreadMessageStore()
	ctx := context.Background()
	require.NoError(t, store.Decrement(ctx, "user-001"))
	count, _ := store.Get(ctx, "user-001")
	assert.Equal(t, int64(0), count)
}

func TestNotificationScheduler_ProcessNotification_NoNotifier(t *testing.T) {
	store := NewMemoryNotificationStore()
	scheduler := NewNotificationScheduler(store, nil, nil, nil, &SchedulerConfig{QueueSize: 100, WorkerCount: 1, SendTimeout: 5 * time.Second})

	notification := &Notification{
		ID:         "n1",
		AlarmID:    "a1",
		Channel:    ChannelSMS,
		Priority:   PriorityNormal,
		MaxRetries: 3,
	}
	scheduler.processNotification(notification)
	assert.Equal(t, StatusPending, notification.Status)
	assert.Equal(t, 1, notification.RetryCount)
}

func TestNotificationScheduler_ProcessNotification_WithNotifier(t *testing.T) {
	store := NewMemoryNotificationStore()
	scheduler := NewNotificationScheduler(store, nil, nil, nil, &SchedulerConfig{QueueSize: 100, WorkerCount: 1, SendTimeout: 5 * time.Second})

	internalNotifier, _ := NewInternalNotifier(&NotificationConfig{RateLimit: 100, BurstLimit: 10}, nil, nil, nil)
	scheduler.RegisterNotifier(ChannelInternal, internalNotifier)

	notification := &Notification{
		ID:         "n1",
		AlarmID:    "a1",
		Channel:    ChannelInternal,
		Priority:   PriorityNormal,
		MaxRetries: 3,
		Recipients: []Recipient{{UserID: "u1"}},
		Content:    "test",
	}
	store.Save(context.Background(), notification)
	scheduler.processNotification(notification)
	assert.Equal(t, StatusSent, notification.Status)
}

func TestNotificationScheduler_HandleSendResult_Success(t *testing.T) {
	store := NewMemoryNotificationStore()
	scheduler := NewNotificationScheduler(store, nil, nil, nil, &SchedulerConfig{QueueSize: 100, WorkerCount: 1})

	notification := &Notification{ID: "n1", AlarmID: "a1", MaxRetries: 3}
	now := time.Now()
	result := &NotificationResult{
		NotificationID: "n1",
		Success:        true,
		Status:         StatusSent,
		DeliveredAt:    &now,
	}
	scheduler.handleSendResult(notification, result)
	assert.Equal(t, StatusSent, notification.Status)
	assert.NotNil(t, notification.SentAt)
}

func TestNotificationScheduler_HandleSendResult_Failed(t *testing.T) {
	store := NewMemoryNotificationStore()
	scheduler := NewNotificationScheduler(store, nil, nil, nil, &SchedulerConfig{QueueSize: 100, WorkerCount: 1})

	notification := &Notification{ID: "n1", AlarmID: "a1", MaxRetries: 3}
	result := &NotificationResult{
		NotificationID: "n1",
		Success:        false,
		Status:         StatusFailed,
		Message:        "send failed",
	}
	scheduler.handleSendResult(notification, result)
	assert.Equal(t, StatusFailed, notification.Status)
	assert.Equal(t, "send failed", notification.ErrorMessage)
}

func TestNotificationScheduler_HandleSendError_MaxRetries(t *testing.T) {
	store := NewMemoryNotificationStore()
	scheduler := NewNotificationScheduler(store, nil, nil, nil, &SchedulerConfig{QueueSize: 100, WorkerCount: 1})

	notification := &Notification{ID: "n1", AlarmID: "a1", MaxRetries: 2, RetryCount: 2}
	scheduler.handleSendError(notification, assert.AnError)
	assert.Equal(t, StatusFailed, notification.Status)
}

func TestNotificationScheduler_HandleSendError_Retryable(t *testing.T) {
	store := NewMemoryNotificationStore()
	scheduler := NewNotificationScheduler(store, nil, nil, nil, &SchedulerConfig{
		QueueSize:    100,
		WorkerCount:  1,
		RetryDelay:   5 * time.Second,
		RetryBackoff: 2.0,
	})

	notification := &Notification{ID: "n1", AlarmID: "a1", MaxRetries: 3, RetryCount: 0}
	scheduler.handleSendError(notification, assert.AnError)
	assert.Equal(t, StatusPending, notification.Status)
	assert.NotNil(t, notification.NextRetryAt)
	assert.Equal(t, 1, notification.RetryCount)
}

func TestNotificationScheduler_HandleSendError_WithLogger(t *testing.T) {
	store := NewMemoryNotificationStore()
	logger := &mockNotificationLogger{}
	scheduler := NewNotificationScheduler(store, logger, nil, nil, &SchedulerConfig{QueueSize: 100, WorkerCount: 1})

	notification := &Notification{ID: "n1", AlarmID: "a1", MaxRetries: 2, RetryCount: 2}
	scheduler.handleSendError(notification, assert.AnError)
	assert.Equal(t, StatusFailed, notification.Status)
}

func TestNotificationScheduler_GetQueueStats(t *testing.T) {
	store := NewMemoryNotificationStore()
	scheduler := NewNotificationScheduler(store, nil, nil, nil, &SchedulerConfig{QueueSize: 100, WorkerCount: 1})

	stats := scheduler.GetQueueStats()
	assert.Len(t, stats, 4)
}

func TestNotificationScheduler_Schedule_Defaults(t *testing.T) {
	store := NewMemoryNotificationStore()
	scheduler := NewNotificationScheduler(store, nil, nil, nil, &SchedulerConfig{QueueSize: 100, WorkerCount: 1, MaxRetries: 5})
	scheduler.Start()
	defer scheduler.Stop()

	notification := &Notification{ID: "n1", AlarmID: "a1", Priority: PriorityNormal}
	require.NoError(t, scheduler.Schedule(context.Background(), notification))
	assert.Equal(t, StatusPending, notification.Status)
	assert.Equal(t, 5, notification.MaxRetries)
	assert.False(t, notification.CreatedAt.IsZero())
}

func TestNotificationScheduler_Schedule_QueueFull(t *testing.T) {
	store := NewMemoryNotificationStore()
	scheduler := NewNotificationScheduler(store, nil, nil, nil, &SchedulerConfig{QueueSize: 1, WorkerCount: 0})
	scheduler.Start()
	defer scheduler.Stop()

	scheduler.priorityQueues[PriorityNormal] <- &Notification{ID: "filler"}
	err := scheduler.Schedule(context.Background(), &Notification{ID: "n1", AlarmID: "a1", Priority: PriorityNormal})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "queue is full")
}

func TestTemplateManager_GetTemplate_FromStore(t *testing.T) {
	store := NewMemoryTemplateStore()
	mgr := NewTemplateManager(store)
	ctx := context.Background()

	tmpl := &NotificationTemplate{
		ID:       "store-tmpl",
		Name:     "存储模板",
		Type:     TemplateTypeSMS,
		Channel:  ChannelSMS,
		Language: LanguageZH,
		Content:  "内容：{{.Msg}}",
		Variables: []TemplateVariable{
			{Name: "Msg", Description: "消息", Required: true},
		},
		Enabled: true,
	}
	require.NoError(t, store.Save(ctx, tmpl))

	got, err := mgr.GetTemplate(ctx, "store-tmpl")
	require.NoError(t, err)
	assert.Equal(t, "store-tmpl", got.ID)

	got2, err := mgr.GetTemplate(ctx, "store-tmpl")
	require.NoError(t, err)
	assert.Equal(t, "store-tmpl", got2.ID)
}

func TestTemplateManager_GetTemplate_NotFound(t *testing.T) {
	store := NewMemoryTemplateStore()
	mgr := NewTemplateManager(store)
	_, err := mgr.GetTemplate(context.Background(), "nonexistent")
	assert.Equal(t, ErrTemplateNotFound, err)
}

func TestTemplateManager_Render_MissingRequiredVariable(t *testing.T) {
	store := NewMemoryTemplateStore()
	mgr := NewTemplateManager(store)
	ctx := context.Background()

	tmpl := &NotificationTemplate{
		ID:       "required-tmpl",
		Name:     "必填变量模板",
		Type:     TemplateTypeSMS,
		Channel:  ChannelSMS,
		Language: LanguageZH,
		Content:  "{{.A}} {{.B}}",
		Variables: []TemplateVariable{
			{Name: "A", Description: "A", Required: true},
			{Name: "B", Description: "B", Required: true},
		},
		Enabled: true,
	}
	require.NoError(t, mgr.Create(ctx, tmpl))

	_, err := mgr.Render("required-tmpl", map[string]interface{}{"A": "1"})
	assert.Error(t, err)
}

func TestTemplateManager_Preview_DefaultValues(t *testing.T) {
	store := NewMemoryTemplateStore()
	mgr := NewTemplateManager(store)
	ctx := context.Background()

	tmpl := &NotificationTemplate{
		ID:       "preview-default-tmpl",
		Name:     "预览默认值",
		Type:     TemplateTypeSMS,
		Channel:  ChannelSMS,
		Language: LanguageZH,
		Content:  "{{.A}} {{.B}}",
		Variables: []TemplateVariable{
			{Name: "A", Description: "A", Required: true},
			{Name: "B", Description: "B", Required: false},
		},
		Enabled: true,
	}
	require.NoError(t, mgr.Create(ctx, tmpl))

	preview, err := mgr.Preview("preview-default-tmpl", map[string]interface{}{"A": "val"})
	require.NoError(t, err)
	assert.Contains(t, preview, "val")
	assert.Contains(t, preview, "{{B}}")
}

func TestTemplateManager_Delete_StoreError(t *testing.T) {
	store := &errorTemplateStore{}
	mgr := NewTemplateManager(store)
	err := mgr.Delete(context.Background(), "tmpl-001")
	assert.Error(t, err)
}

func TestTemplateManager_Update_StoreError(t *testing.T) {
	store := &errorTemplateStore{}
	mgr := NewTemplateManager(store)
	tmpl := &NotificationTemplate{ID: "t1", Name: "test", Content: "ok"}
	err := mgr.Update(context.Background(), tmpl)
	assert.Error(t, err)
}

func TestTemplateManager_Create_StoreError(t *testing.T) {
	store := &errorTemplateStore{}
	mgr := NewTemplateManager(store)
	tmpl := &NotificationTemplate{ID: "t1", Name: "test", Content: "ok"}
	err := mgr.Create(context.Background(), tmpl)
	assert.Error(t, err)
}

func TestSMSNotifier_Send_ValidationError(t *testing.T) {
	notifier, _ := NewSMSNotifier(&NotificationConfig{
		SMSConfig:  &SMSConfig{Provider: "aliyun"},
		RateLimit:  100,
		BurstLimit: 10,
		Timeout:    5 * time.Second,
	}, nil)
	_, err := notifier.Send(context.Background(), &Notification{
		ID:      "test",
		AlarmID: "alarm",
	})
	assert.Error(t, err)
}

func TestEmailNotifier_Send_ValidationError(t *testing.T) {
	notifier, _ := NewEmailNotifier(&NotificationConfig{
		EmailConfig: &EmailConfig{SMTPHost: "smtp.example.com", SMTPPort: 587},
		RateLimit:   100,
		BurstLimit:  10,
	}, nil)
	_, err := notifier.Send(context.Background(), &Notification{
		ID:      "test",
		AlarmID: "alarm",
	})
	assert.Error(t, err)
}

func TestEmailNotifier_RenderEmailTemplate_WithTemplateMgr(t *testing.T) {
	store := NewMemoryTemplateStore()
	mgr := NewTemplateManager(store)
	ctx := context.Background()

	tmpl := &NotificationTemplate{
		ID:       "email-tmpl",
		Name:     "邮件模板",
		Type:     TemplateTypeEmail,
		Channel:  ChannelEmail,
		Language: LanguageZH,
		Content:  "内容：{{.Message}}",
		Variables: []TemplateVariable{
			{Name: "Message", Description: "消息", Required: true},
		},
		Enabled: true,
	}
	require.NoError(t, mgr.Create(ctx, tmpl))

	notifier, _ := NewEmailNotifier(&NotificationConfig{
		EmailConfig: &EmailConfig{SMTPHost: "smtp.example.com", SMTPPort: 587},
		RateLimit:   100,
		BurstLimit:  10,
	}, mgr)

	emailMsg, err := notifier.RenderEmailTemplate("email-tmpl", map[string]interface{}{"Message": "测试"})
	require.NoError(t, err)
	assert.Contains(t, emailMsg.HTMLContent, "测试")
}

func TestInternalNotifier_Send_WithWebSocketHub(t *testing.T) {
	hub := NewWebSocketHub()
	go hub.Run()
	defer func() {
		hub.mu.Lock()
		hub.running = false
		hub.mu.Unlock()
	}()
	time.Sleep(50 * time.Millisecond)

	client := &WebSocketClient{ID: "c1", UserID: "user-001", Send: make(chan []byte, 10)}
	hub.Register(client)
	time.Sleep(100 * time.Millisecond)

	notifier, _ := NewInternalNotifier(&NotificationConfig{RateLimit: 100, BurstLimit: 10}, hub, nil, nil)
	result, err := notifier.Send(context.Background(), &Notification{
		ID:         "test",
		AlarmID:    "alarm1",
		Subject:    "测试",
		Content:    "内容",
		Recipients: []Recipient{{UserID: "user-001", Name: "test"}},
	})
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestInternalNotifier_Send_RateLimitExceeded(t *testing.T) {
	notifier, _ := NewInternalNotifier(&NotificationConfig{RateLimit: 100, BurstLimit: 0}, nil, nil, nil)
	notifier.rateLimiter.Allow("internal:alarm1")
	_, err := notifier.Send(context.Background(), &Notification{
		ID:         "test",
		AlarmID:    "alarm1",
		Recipients: []Recipient{{UserID: "u1"}},
		Content:    "test",
	})
	assert.Error(t, err)
}

func TestWebSocketHub_Unregister(t *testing.T) {
	hub := NewWebSocketHub()
	go hub.Run()
	defer func() {
		hub.mu.Lock()
		hub.running = false
		hub.mu.Unlock()
	}()
	time.Sleep(50 * time.Millisecond)

	client := &WebSocketClient{ID: "c1", UserID: "user-001", Send: make(chan []byte, 10)}
	hub.Register(client)
	time.Sleep(100 * time.Millisecond)
	assert.Len(t, hub.GetOnlineUsers(), 1)

	hub.Unregister(client)
	time.Sleep(100 * time.Millisecond)
	assert.Len(t, hub.GetOnlineUsers(), 0)
}

func TestWebSocketHub_SendToUser_Connected(t *testing.T) {
	hub := NewWebSocketHub()
	go hub.Run()
	defer func() {
		hub.mu.Lock()
		hub.running = false
		hub.mu.Unlock()
	}()
	time.Sleep(50 * time.Millisecond)

	client := &WebSocketClient{ID: "c1", UserID: "user-001", Send: make(chan []byte, 10)}
	hub.Register(client)
	time.Sleep(100 * time.Millisecond)

	err := hub.SendToUser("user-001", &InternalMessage{ID: "msg1", Title: "test", Content: "hello"})
	require.NoError(t, err)

	select {
	case msg := <-client.Send:
		assert.Contains(t, string(msg), "msg1")
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for message")
	}
}

type errorTemplateStore struct{}

func (s *errorTemplateStore) Save(ctx context.Context, tmpl *NotificationTemplate) error {
	return assert.AnError
}
func (s *errorTemplateStore) Get(ctx context.Context, id string) (*NotificationTemplate, error) {
	return nil, assert.AnError
}
func (s *errorTemplateStore) GetByIDAndLanguage(ctx context.Context, id string, language Language) (*NotificationTemplate, error) {
	return nil, assert.AnError
}
func (s *errorTemplateStore) Delete(ctx context.Context, id string) error {
	return assert.AnError
}
func (s *errorTemplateStore) List(ctx context.Context, query *TemplateQuery) ([]*NotificationTemplate, int64, error) {
	return nil, 0, assert.AnError
}

type mockNotificationLogger struct{}

func (l *mockNotificationLogger) Log(ctx context.Context, notification *Notification, result *NotificationResult) error {
	return nil
}
func (l *mockNotificationLogger) Query(ctx context.Context, query *NotificationLogQuery) ([]*NotificationLog, int64, error) {
	return nil, 0, nil
}

func TestWebSocketHub_IsUserOnline(t *testing.T) {
	hub := NewWebSocketHub()
	go hub.Run()
	defer func() {
		hub.mu.Lock()
		hub.running = false
		hub.mu.Unlock()
	}()
	time.Sleep(50 * time.Millisecond)

	assert.False(t, hub.IsUserOnline("user-001"))

	client := &WebSocketClient{ID: "c1", UserID: "user-001", Send: make(chan []byte, 10)}
	hub.Register(client)
	time.Sleep(100 * time.Millisecond)
	assert.True(t, hub.IsUserOnline("user-001"))

	hub.Unregister(client)
	time.Sleep(100 * time.Millisecond)
	assert.False(t, hub.IsUserOnline("user-001"))
}

func TestTemplateManager_List(t *testing.T) {
	store := NewMemoryTemplateStore()
	mgr := NewTemplateManager(store)
	ctx := context.Background()

	mgr.Create(ctx, &NotificationTemplate{ID: "1", Name: "t1", Type: TemplateTypeSMS, Channel: ChannelSMS, Language: LanguageZH, Content: "c1", Enabled: true})
	mgr.Create(ctx, &NotificationTemplate{ID: "2", Name: "t2", Type: TemplateTypeEmail, Channel: ChannelEmail, Language: LanguageEN, Content: "c2", Enabled: true})

	results, total, err := mgr.List(ctx, &TemplateQuery{Type: TemplateTypeSMS})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, results, 1)
}

func TestSMSNotifier_prepareContent_WithTemplate(t *testing.T) {
	store := NewMemoryTemplateStore()
	mgr := NewTemplateManager(store)
	ctx := context.Background()

	mgr.Create(ctx, &NotificationTemplate{
		ID:       "sms-tmpl",
		Name:     "SMS模板",
		Type:     TemplateTypeSMS,
		Channel:  ChannelSMS,
		Language: LanguageZH,
		Content:  "告警：{{.Message}}",
		Variables: []TemplateVariable{
			{Name: "Message", Description: "消息", Required: true},
		},
		Enabled: true,
	})

	notifier, _ := NewSMSNotifier(&NotificationConfig{
		SMSConfig:  &SMSConfig{Provider: "aliyun"},
		RateLimit:  100,
		BurstLimit: 10,
		Timeout:    5 * time.Second,
	}, mgr)

	content, err := notifier.prepareContent(&Notification{
		TemplateID:   "sms-tmpl",
		TemplateData: map[string]interface{}{"Message": "设备故障"},
	})
	require.NoError(t, err)
	assert.Contains(t, content, "设备故障")
}

func TestSMSNotifier_prepareContent_NoTemplate(t *testing.T) {
	notifier, _ := NewSMSNotifier(&NotificationConfig{
		SMSConfig:  &SMSConfig{Provider: "aliyun"},
		RateLimit:  100,
		BurstLimit: 10,
		Timeout:    5 * time.Second,
	}, nil)

	content, err := notifier.prepareContent(&Notification{
		Content: "直接内容",
	})
	require.NoError(t, err)
	assert.Equal(t, "直接内容", content)
}

func TestInternalNotifier_prepareMessage_NoHTMLContent(t *testing.T) {
	notifier, _ := NewInternalNotifier(&NotificationConfig{RateLimit: 100, BurstLimit: 10}, nil, nil, nil)

	msg, err := notifier.prepareMessage(&Notification{
		ID:         "test",
		AlarmID:    "alarm1",
		Subject:    "标题",
		Content:    "内容",
		Priority:   PriorityNormal,
		Recipients: []Recipient{{UserID: "u1"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "内容", msg.Content)
	assert.Empty(t, msg.HTMLContent)
}

func TestInternalNotifier_prepareMessage_NoTags(t *testing.T) {
	notifier, _ := NewInternalNotifier(&NotificationConfig{RateLimit: 100, BurstLimit: 10}, nil, nil, nil)

	msg, err := notifier.prepareMessage(&Notification{
		ID:         "test",
		AlarmID:    "alarm1",
		Subject:    "标题",
		Content:    "内容",
		Priority:   PriorityLow,
		Recipients: []Recipient{{UserID: "u1"}},
	})
	require.NoError(t, err)
	assert.Nil(t, msg.Tags)
}

func TestInternalNotifier_HealthCheck_WithRunningHub(t *testing.T) {
	hub := NewWebSocketHub()
	go hub.Run()
	defer func() {
		hub.mu.Lock()
		hub.running = false
		hub.mu.Unlock()
	}()
	time.Sleep(50 * time.Millisecond)

	notifier, _ := NewInternalNotifier(&NotificationConfig{RateLimit: 100, BurstLimit: 10}, hub, nil, nil)
	assert.NoError(t, notifier.HealthCheck(context.Background()))
}

func TestEmailNotifier_prepareEmail_WithCC(t *testing.T) {
	notifier, _ := NewEmailNotifier(&NotificationConfig{
		EmailConfig: &EmailConfig{SMTPHost: "smtp.example.com", SMTPPort: 587, FromName: "Test", FromAddress: "test@example.com"},
		RateLimit:   100,
		BurstLimit:  10,
	}, nil)

	email, err := notifier.prepareEmail(&Notification{
		Subject:    "测试",
		Content:    "内容",
		Priority:   PriorityNormal,
		AlarmID:    "alarm1",
		Recipients: []Recipient{{Name: "用户", Email: "user@example.com"}, {Name: "CC用户", Email: "cc@example.com"}},
	})
	require.NoError(t, err)
	assert.Len(t, email.To, 2)
}

func TestEmailNotifier_prepareEmail_NoContent(t *testing.T) {
	notifier, _ := NewEmailNotifier(&NotificationConfig{
		EmailConfig: &EmailConfig{SMTPHost: "smtp.example.com", SMTPPort: 587, FromName: "Test", FromAddress: "test@example.com"},
		RateLimit:   100,
		BurstLimit:  10,
	}, nil)

	email, err := notifier.prepareEmail(&Notification{
		Subject:     "测试",
		HTMLContent: "<p>HTML</p>",
		Priority:    PriorityNormal,
		AlarmID:     "alarm1",
		Recipients:  []Recipient{{Name: "用户", Email: "user@example.com"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "<p>HTML</p>", email.HTMLContent)
}

func TestNotificationScheduler_ProcessRetries(t *testing.T) {
	store := NewMemoryNotificationStore()
	scheduler := NewNotificationScheduler(store, nil, nil, nil, &SchedulerConfig{
		QueueSize:    100,
		WorkerCount:  1,
		RetryDelay:   1 * time.Second,
		RetryBackoff: 2.0,
		MaxRetries:   3,
	})
	scheduler.Start()
	defer scheduler.Stop()

	pastTime := time.Now().Add(-1 * time.Second)
	notification := &Notification{
		ID:          "n1",
		AlarmID:     "a1",
		Channel:     ChannelInternal,
		Priority:    PriorityNormal,
		Status:      StatusPending,
		MaxRetries:  3,
		RetryCount:  1,
		NextRetryAt: &pastTime,
	}
	store.Save(context.Background(), notification)
	scheduler.processRetries()
}

func TestNotificationScheduler_ProcessRetries_MaxRetries(t *testing.T) {
	store := NewMemoryNotificationStore()
	scheduler := NewNotificationScheduler(store, nil, nil, nil, &SchedulerConfig{
		QueueSize:    100,
		WorkerCount:  1,
		RetryDelay:   1 * time.Second,
		RetryBackoff: 2.0,
		MaxRetries:   2,
	})
	scheduler.Start()
	defer scheduler.Stop()

	pastTime := time.Now().Add(-1 * time.Second)
	notification := &Notification{
		ID:          "n1",
		AlarmID:     "a1",
		Channel:     ChannelInternal,
		Priority:    PriorityNormal,
		Status:      StatusPending,
		MaxRetries:  2,
		RetryCount:  2,
		NextRetryAt: &pastTime,
	}
	store.Save(context.Background(), notification)
	scheduler.processRetries()

	got, _ := store.Get(context.Background(), "n1")
	assert.Equal(t, StatusFailed, got.Status)
}

func TestNotificationScheduler_ProcessRetries_FutureRetry(t *testing.T) {
	store := NewMemoryNotificationStore()
	scheduler := NewNotificationScheduler(store, nil, nil, nil, &SchedulerConfig{
		QueueSize:    100,
		WorkerCount:  1,
		RetryDelay:   1 * time.Second,
		RetryBackoff: 2.0,
		MaxRetries:   3,
	})
	scheduler.Start()
	defer scheduler.Stop()

	futureTime := time.Now().Add(1 * time.Hour)
	notification := &Notification{
		ID:          "n1",
		AlarmID:     "a1",
		Channel:     ChannelInternal,
		Priority:    PriorityNormal,
		Status:      StatusPending,
		MaxRetries:  3,
		RetryCount:  1,
		NextRetryAt: &futureTime,
	}
	store.Save(context.Background(), notification)
	scheduler.processRetries()

	got, _ := store.Get(context.Background(), "n1")
	assert.Equal(t, StatusPending, got.Status)
}

func TestNotificationScheduler_HandleSendResult_WithLogger(t *testing.T) {
	store := NewMemoryNotificationStore()
	logger := &mockNotificationLogger{}
	scheduler := NewNotificationScheduler(store, logger, nil, nil, &SchedulerConfig{QueueSize: 100, WorkerCount: 1})

	notification := &Notification{ID: "n1", AlarmID: "a1", MaxRetries: 3}
	now := time.Now()
	result := &NotificationResult{
		NotificationID: "n1",
		Success:        true,
		Status:         StatusSent,
		DeliveredAt:    &now,
	}
	scheduler.handleSendResult(notification, result)
	assert.Equal(t, StatusSent, notification.Status)
}

func TestEmailNotifier_buildEmailContent_WithCC(t *testing.T) {
	notifier, _ := NewEmailNotifier(&NotificationConfig{
		EmailConfig: &EmailConfig{SMTPHost: "smtp.example.com", SMTPPort: 587, FromName: "Test", FromAddress: "test@example.com"},
		RateLimit:   100,
		BurstLimit:  10,
	}, nil)

	email := &EmailMessage{
		From:        mail.Address{Name: "Test", Address: "test@example.com"},
		To:          []mail.Address{{Name: "User", Address: "user@example.com"}},
		CC:          []mail.Address{{Name: "CC", Address: "cc@example.com"}},
		Subject:     "测试",
		TextContent: "内容",
		Headers:     map[string]string{},
	}
	content, err := notifier.buildEmailContent(email)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Cc:")
}

func TestEmailNotifier_buildEmailContent_HTMLAndAttachments(t *testing.T) {
	notifier, _ := NewEmailNotifier(&NotificationConfig{
		EmailConfig: &EmailConfig{SMTPHost: "smtp.example.com", SMTPPort: 587, FromName: "Test", FromAddress: "test@example.com"},
		RateLimit:   100,
		BurstLimit:  10,
	}, nil)

	email := &EmailMessage{
		From:        mail.Address{Name: "Test", Address: "test@example.com"},
		To:          []mail.Address{{Name: "User", Address: "user@example.com"}},
		Subject:     "测试",
		HTMLContent: "<p>HTML</p>",
		TextContent: "文本",
		Headers:     map[string]string{},
		Attachments: []Attachment{{Name: "file.txt", Content: []byte("hello"), MimeType: "text/plain"}},
	}
	content, err := notifier.buildEmailContent(email)
	require.NoError(t, err)
	assert.Contains(t, string(content), "multipart/mixed")
}

func TestEmailNotifier_RenderEmailTemplate_Cached(t *testing.T) {
	store := NewMemoryTemplateStore()
	mgr := NewTemplateManager(store)
	ctx := context.Background()

	mgr.Create(ctx, &NotificationTemplate{
		ID:       "cached-email-tmpl",
		Name:     "缓存邮件模板",
		Type:     TemplateTypeEmail,
		Channel:  ChannelEmail,
		Language: LanguageZH,
		Content:  "内容：{{.Message}}",
		Variables: []TemplateVariable{
			{Name: "Message", Description: "消息", Required: true},
		},
		Enabled: true,
	})

	notifier, _ := NewEmailNotifier(&NotificationConfig{
		EmailConfig: &EmailConfig{SMTPHost: "smtp.example.com", SMTPPort: 587},
		RateLimit:   100,
		BurstLimit:  10,
	}, mgr)

	emailMsg1, err := notifier.RenderEmailTemplate("cached-email-tmpl", map[string]interface{}{"Message": "第一次"})
	require.NoError(t, err)
	assert.Contains(t, emailMsg1.HTMLContent, "第一次")

	emailMsg2, err := notifier.RenderEmailTemplate("cached-email-tmpl", map[string]interface{}{"Message": "第二次"})
	require.NoError(t, err)
	assert.Contains(t, emailMsg2.HTMLContent, "第二次")
}

func TestEmailNotifier_RenderEmailTemplate_NotFound(t *testing.T) {
	store := NewMemoryTemplateStore()
	mgr := NewTemplateManager(store)

	notifier, _ := NewEmailNotifier(&NotificationConfig{
		EmailConfig: &EmailConfig{SMTPHost: "smtp.example.com", SMTPPort: 587},
		RateLimit:   100,
		BurstLimit:  10,
	}, mgr)

	_, err := notifier.RenderEmailTemplate("nonexistent", map[string]interface{}{})
	assert.Error(t, err)
}

func TestSMSNotifier_Send_AliyunProvider(t *testing.T) {
	notifier, _ := NewSMSNotifier(&NotificationConfig{
		SMSConfig:  &SMSConfig{Provider: "aliyun", AccessKey: "key", AccessSecret: "secret", SignName: "test", Region: "cn-hangzhou"},
		RateLimit:  100,
		BurstLimit: 10,
		Timeout:    5 * time.Second,
	}, nil)

	_, err := notifier.Send(context.Background(), &Notification{
		ID:         "test",
		AlarmID:    "alarm",
		Recipients: []Recipient{{Name: "test", Phone: "13800138000"}},
		Content:    "test",
	})
	assert.Error(t, err)
}

func TestSMSNotifier_Send_TencentProvider(t *testing.T) {
	notifier, _ := NewSMSNotifier(&NotificationConfig{
		SMSConfig:  &SMSConfig{Provider: "tencent", AccessKey: "key", AccessSecret: "secret", SignName: "test", Region: "ap-guangzhou"},
		RateLimit:  100,
		BurstLimit: 10,
		Timeout:    5 * time.Second,
	}, nil)

	_, err := notifier.Send(context.Background(), &Notification{
		ID:         "test",
		AlarmID:    "alarm",
		Recipients: []Recipient{{Name: "test", Phone: "13800138000"}},
		Content:    "test",
	})
	assert.Error(t, err)
}
