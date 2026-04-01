package notifier

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"text/template"
	"time"
)

var (
	ErrTemplateNotFound   = errors.New("template not found")
	ErrTemplateInvalid    = errors.New("template is invalid")
	ErrTemplateVariableMissing = errors.New("template variable missing")
)

// TemplateType 模板类型
type TemplateType string

const (
	TemplateTypeSMS     TemplateType = "sms"
	TemplateTypeEmail   TemplateType = "email"
	TemplateTypeInternal TemplateType = "internal"
	TemplateTypeWeChat  TemplateType = "wechat"
	TemplateTypeDingTalk TemplateType = "dingtalk"
)

// Language 语言
type Language string

const (
	LanguageZH Language = "zh-CN"
	LanguageEN Language = "en-US"
)

// NotificationTemplate 通知模板
type NotificationTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        TemplateType           `json:"type"`
	Channel     NotificationChannel    `json:"channel"`
	Language    Language               `json:"language"`

	// 模板内容
	Subject     string                 `json:"subject,omitempty"`     // 邮件主题
	Content     string                 `json:"content"`               // 文本内容
	HTMLContent string                 `json:"html_content,omitempty"` // HTML内容

	// 变量定义
	Variables   []TemplateVariable     `json:"variables"`

	// 元数据
	Description string                 `json:"description"`
	Tags        []string               `json:"tags,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   string                 `json:"created_by"`
	Enabled     bool                   `json:"enabled"`

	// 第三方平台模板ID
	ExternalID  string                 `json:"external_id,omitempty"` // 阿里云/腾讯云模板ID
}

// TemplateVariable 模板变量
type TemplateVariable struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Default     string `json:"default,omitempty"`
	Type        string `json:"type"` // string, number, date, etc.
}

// TemplateManager 模板管理器
type TemplateManager interface {
	// Create 创建模板
	Create(ctx context.Context, tmpl *NotificationTemplate) error

	// Update 更新模板
	Update(ctx context.Context, tmpl *NotificationTemplate) error

	// Delete 删除模板
	Delete(ctx context.Context, id string) error

	// Get 获取模板
	Get(ctx context.Context, id string) (string, error)

	// GetTemplate 获取模板详情
	GetTemplate(ctx context.Context, id string) (*NotificationTemplate, error)

	// List 列出模板
	List(ctx context.Context, query *TemplateQuery) ([]*NotificationTemplate, int64, error)

	// Render 渲染模板
	Render(templateID string, data map[string]interface{}) (string, error)

	// RenderWithLanguage 使用指定语言渲染模板
	RenderWithLanguage(templateID string, language Language, data map[string]interface{}) (string, error)

	// Preview 预览模板
	Preview(templateID string, data map[string]interface{}) (string, error)

	// Validate 验证模板
	Validate(templateID string, data map[string]interface{}) error
}

// TemplateQuery 模板查询
type TemplateQuery struct {
	Type     TemplateType
	Channel  NotificationChannel
	Language Language
	Tags     []string
	Enabled  *bool
	Page     int
	PageSize int
}

// TemplateStore 模板存储接口
type TemplateStore interface {
	// Save 保存模板
	Save(ctx context.Context, tmpl *NotificationTemplate) error

	// Get 获取模板
	Get(ctx context.Context, id string) (*NotificationTemplate, error)

	// GetByIDAndLanguage 根据ID和语言获取模板
	GetByIDAndLanguage(ctx context.Context, id string, language Language) (*NotificationTemplate, error)

	// Delete 删除模板
	Delete(ctx context.Context, id string) error

	// List 列出模板
	List(ctx context.Context, query *TemplateQuery) ([]*NotificationTemplate, int64, error)
}

// DefaultTemplateManager 默认模板管理器
type DefaultTemplateManager struct {
	store       TemplateStore
	cache       map[string]*NotificationTemplate
	mu          sync.RWMutex
}

// NewTemplateManager 创建模板管理器
func NewTemplateManager(store TemplateStore) *DefaultTemplateManager {
	return &DefaultTemplateManager{
		store: store,
		cache: make(map[string]*NotificationTemplate),
	}
}

// Create 创建模板
func (m *DefaultTemplateManager) Create(ctx context.Context, tmpl *NotificationTemplate) error {
	if err := m.validateTemplate(tmpl); err != nil {
		return err
	}

	tmpl.CreatedAt = time.Now()
	tmpl.UpdatedAt = time.Now()

	if err := m.store.Save(ctx, tmpl); err != nil {
		return err
	}

	// 更新缓存
	m.mu.Lock()
	m.cache[tmpl.ID] = tmpl
	m.mu.Unlock()

	return nil
}

// Update 更新模板
func (m *DefaultTemplateManager) Update(ctx context.Context, tmpl *NotificationTemplate) error {
	if err := m.validateTemplate(tmpl); err != nil {
		return err
	}

	tmpl.UpdatedAt = time.Now()

	if err := m.store.Save(ctx, tmpl); err != nil {
		return err
	}

	// 更新缓存
	m.mu.Lock()
	m.cache[tmpl.ID] = tmpl
	m.mu.Unlock()

	return nil
}

// Delete 删除模板
func (m *DefaultTemplateManager) Delete(ctx context.Context, id string) error {
	if err := m.store.Delete(ctx, id); err != nil {
		return err
	}

	// 删除缓存
	m.mu.Lock()
	delete(m.cache, id)
	m.mu.Unlock()

	return nil
}

// Get 获取模板内容
func (m *DefaultTemplateManager) Get(ctx context.Context, id string) (string, error) {
	tmpl, err := m.GetTemplate(ctx, id)
	if err != nil {
		return "", err
	}
	return tmpl.Content, nil
}

// GetTemplate 获取模板详情
func (m *DefaultTemplateManager) GetTemplate(ctx context.Context, id string) (*NotificationTemplate, error) {
	// 先从缓存获取
	m.mu.RLock()
	tmpl, exists := m.cache[id]
	m.mu.RUnlock()

	if exists {
		return tmpl, nil
	}

	// 从存储获取
	tmpl, err := m.store.Get(ctx, id)
	if err != nil {
		return nil, ErrTemplateNotFound
	}

	// 更新缓存
	m.mu.Lock()
	m.cache[id] = tmpl
	m.mu.Unlock()

	return tmpl, nil
}

// List 列出模板
func (m *DefaultTemplateManager) List(ctx context.Context, query *TemplateQuery) ([]*NotificationTemplate, int64, error) {
	return m.store.List(ctx, query)
}

// Render 渲染模板
func (m *DefaultTemplateManager) Render(templateID string, data map[string]interface{}) (string, error) {
	return m.RenderWithLanguage(templateID, LanguageZH, data)
}

// RenderWithLanguage 使用指定语言渲染模板
func (m *DefaultTemplateManager) RenderWithLanguage(templateID string, language Language, data map[string]interface{}) (string, error) {
	// 获取模板
	m.mu.RLock()
	tmpl, exists := m.cache[templateID]
	m.mu.RUnlock()

	if !exists {
		// 从存储获取
		var err error
		tmpl, err = m.store.GetByIDAndLanguage(context.Background(), templateID, language)
		if err != nil {
			return "", ErrTemplateNotFound
		}

		// 更新缓存
		m.mu.Lock()
		m.cache[templateID] = tmpl
		m.mu.Unlock()
	}

	// 验证数据
	if err := m.Validate(templateID, data); err != nil {
		return "", err
	}

	// 渲染模板
	return m.renderTemplate(tmpl, data)
}

// Preview 预览模板
func (m *DefaultTemplateManager) Preview(templateID string, data map[string]interface{}) (string, error) {
	// 获取模板
	m.mu.RLock()
	tmpl, exists := m.cache[templateID]
	m.mu.RUnlock()

	if !exists {
		return "", ErrTemplateNotFound
	}

	// 使用默认值填充缺失的变量
	previewData := make(map[string]interface{})
	for k, v := range data {
		previewData[k] = v
	}

	for _, variable := range tmpl.Variables {
		if _, ok := previewData[variable.Name]; !ok {
			if variable.Default != "" {
				previewData[variable.Name] = variable.Default
			} else {
				previewData[variable.Name] = fmt.Sprintf("{{%s}}", variable.Name)
			}
		}
	}

	return m.renderTemplate(tmpl, previewData)
}

// Validate 验证模板数据
func (m *DefaultTemplateManager) Validate(templateID string, data map[string]interface{}) error {
	// 获取模板
	m.mu.RLock()
	tmpl, exists := m.cache[templateID]
	m.mu.RUnlock()

	if !exists {
		return ErrTemplateNotFound
	}

	// 检查必需变量
	for _, variable := range tmpl.Variables {
		if variable.Required {
			if _, ok := data[variable.Name]; !ok {
				return fmt.Errorf("%w: %s", ErrTemplateVariableMissing, variable.Name)
			}
		}
	}

	return nil
}

// validateTemplate 验证模板
func (m *DefaultTemplateManager) validateTemplate(tmpl *NotificationTemplate) error {
	if tmpl.ID == "" {
		return errors.New("template id is required")
	}

	if tmpl.Name == "" {
		return errors.New("template name is required")
	}

	if tmpl.Content == "" {
		return errors.New("template content is required")
	}

	// 验证模板语法
	if _, err := template.New("test").Parse(tmpl.Content); err != nil {
		return fmt.Errorf("invalid template syntax: %w", err)
	}

	// 如果有HTML内容，验证HTML模板语法
	if tmpl.HTMLContent != "" {
		if _, err := template.New("test").Parse(tmpl.HTMLContent); err != nil {
			return fmt.Errorf("invalid html template syntax: %w", err)
		}
	}

	return nil
}

// renderTemplate 渲染模板
func (m *DefaultTemplateManager) renderTemplate(tmpl *NotificationTemplate, data map[string]interface{}) (string, error) {
	// 解析模板
	t, err := template.New(tmpl.ID).Parse(tmpl.Content)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// 渲染模板
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return buf.String(), nil
}

// MemoryTemplateStore 内存模板存储
type MemoryTemplateStore struct {
	templates map[string]*NotificationTemplate
	mu        sync.RWMutex
}

// NewMemoryTemplateStore 创建内存模板存储
func NewMemoryTemplateStore() *MemoryTemplateStore {
	return &MemoryTemplateStore{
		templates: make(map[string]*NotificationTemplate),
	}
}

// Save 保存模板
func (s *MemoryTemplateStore) Save(ctx context.Context, tmpl *NotificationTemplate) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.templates[tmpl.ID] = tmpl
	return nil
}

// Get 获取模板
func (s *MemoryTemplateStore) Get(ctx context.Context, id string) (*NotificationTemplate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tmpl, ok := s.templates[id]
	if !ok {
		return nil, ErrTemplateNotFound
	}
	return tmpl, nil
}

// GetByIDAndLanguage 根据ID和语言获取模板
func (s *MemoryTemplateStore) GetByIDAndLanguage(ctx context.Context, id string, language Language) (*NotificationTemplate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 先尝试获取指定语言的模板
	key := fmt.Sprintf("%s_%s", id, language)
	if tmpl, ok := s.templates[key]; ok {
		return tmpl, nil
	}

	// 回退到默认模板
	tmpl, ok := s.templates[id]
	if !ok {
		return nil, ErrTemplateNotFound
	}
	return tmpl, nil
}

// Delete 删除模板
func (s *MemoryTemplateStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.templates, id)
	return nil
}

// List 列出模板
func (s *MemoryTemplateStore) List(ctx context.Context, query *TemplateQuery) ([]*NotificationTemplate, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*NotificationTemplate, 0)
	for _, tmpl := range s.templates {
		// 过滤条件
		if query.Type != "" && tmpl.Type != query.Type {
			continue
		}
		if query.Channel != "" && tmpl.Channel != query.Channel {
			continue
		}
		if query.Language != "" && tmpl.Language != query.Language {
			continue
		}
		if query.Enabled != nil && tmpl.Enabled != *query.Enabled {
			continue
		}
		if len(query.Tags) > 0 {
			matched := false
			for _, tag := range query.Tags {
				for _, t := range tmpl.Tags {
					if t == tag {
						matched = true
						break
					}
				}
				if matched {
					break
				}
			}
			if !matched {
				continue
			}
		}
		result = append(result, tmpl)
	}

	// 分页
	total := int64(len(result))
	if query.Page > 0 && query.PageSize > 0 {
		start := (query.Page - 1) * query.PageSize
		end := start + query.PageSize
		if start >= len(result) {
			return []*NotificationTemplate{}, total, nil
		}
		if end > len(result) {
			end = len(result)
		}
		result = result[start:end]
	}

	return result, total, nil
}

// BuiltInTemplates 内置模板
var BuiltInTemplates = []*NotificationTemplate{
	{
		ID:          "alarm_critical",
		Name:        "严重告警通知",
		Type:        TemplateTypeSMS,
		Channel:     ChannelSMS,
		Language:    LanguageZH,
		Content:     "【新能源监控】严重告警：{{.StationName}} - {{.DeviceName}}，告警内容：{{.Message}}，触发时间：{{.TriggerTime}}",
		Description: "严重告警短信通知模板",
		Variables: []TemplateVariable{
			{Name: "StationName", Description: "站点名称", Required: true},
			{Name: "DeviceName", Description: "设备名称", Required: true},
			{Name: "Message", Description: "告警消息", Required: true},
			{Name: "TriggerTime", Description: "触发时间", Required: true},
		},
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		ID:          "alarm_email",
		Name:        "告警邮件通知",
		Type:        TemplateTypeEmail,
		Channel:     ChannelEmail,
		Language:    LanguageZH,
		Subject:     "【新能源监控】告警通知 - {{.AlarmLevel}}",
		Content:     "站点：{{.StationName}}\n设备：{{.DeviceName}}\n告警级别：{{.AlarmLevel}}\n告警内容：{{.Message}}\n触发时间：{{.TriggerTime}}",
		HTMLContent: `<html><body><h2>告警通知</h2><p>站点：{{.StationName}}</p><p>设备：{{.DeviceName}}</p><p>告警级别：{{.AlarmLevel}}</p><p>告警内容：{{.Message}}</p><p>触发时间：{{.TriggerTime}}</p></body></html>`,
		Description: "告警邮件通知模板",
		Variables: []TemplateVariable{
			{Name: "StationName", Description: "站点名称", Required: true},
			{Name: "DeviceName", Description: "设备名称", Required: true},
			{Name: "AlarmLevel", Description: "告警级别", Required: true},
			{Name: "Message", Description: "告警消息", Required: true},
			{Name: "TriggerTime", Description: "触发时间", Required: true},
		},
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		ID:          "alarm_internal",
		Name:        "系统内告警通知",
		Type:        TemplateTypeInternal,
		Channel:     ChannelInternal,
		Language:    LanguageZH,
		Content:     "【{{.AlarmLevel}}】{{.StationName}} - {{.DeviceName}}：{{.Message}}",
		Description: "系统内告警通知模板",
		Variables: []TemplateVariable{
			{Name: "AlarmLevel", Description: "告警级别", Required: true},
			{Name: "StationName", Description: "站点名称", Required: true},
			{Name: "DeviceName", Description: "设备名称", Required: true},
			{Name: "Message", Description: "告警消息", Required: true},
		},
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
}

// InitBuiltInTemplates 初始化内置模板
func InitBuiltInTemplates(store TemplateStore) error {
	ctx := context.Background()
	for _, tmpl := range BuiltInTemplates {
		if err := store.Save(ctx, tmpl); err != nil {
			return fmt.Errorf("failed to save template %s: %w", tmpl.ID, err)
		}
	}
	return nil
}
