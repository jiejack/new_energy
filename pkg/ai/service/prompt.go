package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"
)

// PromptTemplate Prompt模板结构
type PromptTemplate struct {
	// 模板ID
	ID string `json:"id" yaml:"id"`
	
	// 模板名称
	Name string `json:"name" yaml:"name"`
	
	// 模板描述
	Description string `json:"description" yaml:"description"`
	
	// 模板内容
	Content string `json:"content" yaml:"content"`
	
	// 模板类型
	Type string `json:"type" yaml:"type"`
	
	// 变量定义
	Variables []TemplateVariable `json:"variables" yaml:"variables"`
	
	// 示例
	Examples []TemplateExample `json:"examples" yaml:"examples"`
	
	// 标签
	Tags []string `json:"tags" yaml:"tags"`
	
	// 版本号
	Version string `json:"version" yaml:"version"`
	
	// 创建时间
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`
	
	// 更新时间
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at"`
	
	// 创建者
	CreatedBy string `json:"created_by" yaml:"created_by"`
	
	// 是否启用
	Enabled bool `json:"enabled" yaml:"enabled"`
	
	// 元数据
	Metadata map[string]interface{} `json:"metadata" yaml:"metadata"`
}

// TemplateVariable 模板变量定义
type TemplateVariable struct {
	// 变量名
	Name string `json:"name" yaml:"name"`
	
	// 变量描述
	Description string `json:"description" yaml:"description"`
	
	// 变量类型 (string, number, boolean, array, object)
	Type string `json:"type" yaml:"type"`
	
	// 是否必填
	Required bool `json:"required" yaml:"required"`
	
	// 默认值
	Default interface{} `json:"default" yaml:"default"`
	
	// 验证规则
	Validation string `json:"validation" yaml:"validation"`
	
	// 枚举值
	Enum []string `json:"enum" yaml:"enum"`
	
	// 最小值 (用于number类型)
	Min *float64 `json:"min" yaml:"min"`
	
	// 最大值 (用于number类型)
	Max *float64 `json:"max" yaml:"max"`
	
	// 最小长度 (用于string类型)
	MinLength *int `json:"min_length" yaml:"min_length"`
	
	// 最大长度 (用于string类型)
	MaxLength *int `json:"max_length" yaml:"max_length"`
}

// TemplateExample 模板示例
type TemplateExample struct {
	// 示例名称
	Name string `json:"name" yaml:"name"`
	
	// 输入变量值
	Input map[string]interface{} `json:"input" yaml:"input"`
	
	// 预期输出
	Output string `json:"output" yaml:"output"`
	
	// 说明
	Note string `json:"note" yaml:"note"`
}

// TemplateVersion 模板版本
type TemplateVersion struct {
	// 模板ID
	TemplateID string `json:"template_id" yaml:"template_id"`
	
	// 版本号
	Version string `json:"version" yaml:"version"`
	
	// 模板内容
	Content string `json:"content" yaml:"content"`
	
	// 变更说明
	ChangeLog string `json:"change_log" yaml:"change_log"`
	
	// 创建时间
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`
	
	// 创建者
	CreatedBy string `json:"created_by" yaml:"created_by"`
}

// TemplateFilter 模板过滤条件
type TemplateFilter struct {
	// 名称模糊匹配
	Name string `json:"name" yaml:"name"`
	
	// 类型
	Type string `json:"type" yaml:"type"`
	
	// 标签
	Tags []string `json:"tags" yaml:"tags"`
	
	// 是否启用
	Enabled *bool `json:"enabled" yaml:"enabled"`
	
	// 创建者
	CreatedBy string `json:"created_by" yaml:"created_by"`
	
	// 创建时间范围
	CreatedAfter *time.Time `json:"created_after" yaml:"created_after"`
	CreatedBefore *time.Time `json:"created_before" yaml:"created_before"`
}

// RenderOptions 渲染选项
type RenderOptions struct {
	// 是否严格模式（缺少变量时报错）
	Strict bool `json:"strict" yaml:"strict"`
	
	// 缺失变量处理方式 (error, empty, default)
	MissingVariableHandling string `json:"missing_variable_handling" yaml:"missing_variable_handling"`
	
	// 自定义分隔符
	Delimiters []string `json:"delimiters" yaml:"delimiters"`
	
	// 转义函数
	EscapeFunc func(string) string `json:"-"`
	
	// 额外上下文
	Context map[string]interface{} `json:"context" yaml:"context"`
}

// TemplateManager 模板管理器
type TemplateManager struct {
	templates    map[string]*PromptTemplate
	versions     map[string][]*TemplateVersion
	mu           sync.RWMutex
	storage      TemplateStorage
	cache        map[string]string
	cacheEnabled bool
}

// TemplateStorage 模板存储接口
type TemplateStorage interface {
	// Save 保存模板
	Save(ctx context.Context, tmpl *PromptTemplate) error
	
	// Load 加载模板
	Load(ctx context.Context, id string) (*PromptTemplate, error)
	
	// LoadAll 加载所有模板
	LoadAll(ctx context.Context) ([]*PromptTemplate, error)
	
	// Delete 删除模板
	Delete(ctx context.Context, id string) error
	
	// SaveVersion 保存版本
	SaveVersion(ctx context.Context, version *TemplateVersion) error
	
	// LoadVersions 加载版本历史
	LoadVersions(ctx context.Context, templateID string) ([]*TemplateVersion, error)
}

// NewTemplateManager 创建模板管理器
func NewTemplateManager(storage TemplateStorage, enableCache bool) *TemplateManager {
	return &TemplateManager{
		templates:    make(map[string]*PromptTemplate),
		versions:     make(map[string][]*TemplateVersion),
		storage:      storage,
		cache:        make(map[string]string),
		cacheEnabled: enableCache,
	}
}

// Create 创建模板
func (m *TemplateManager) Create(ctx context.Context, tmpl *PromptTemplate) error {
	if err := m.validateTemplate(tmpl); err != nil {
		return fmt.Errorf("validate template: %w", err)
	}
	
	tmpl.ID = generateTemplateID()
	tmpl.CreatedAt = time.Now()
	tmpl.UpdatedAt = time.Now()
	tmpl.Version = "1.0.0"
	
	if tmpl.Enabled {
		tmpl.Enabled = true
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if err := m.storage.Save(ctx, tmpl); err != nil {
		return fmt.Errorf("save template: %w", err)
	}
	
	m.templates[tmpl.ID] = tmpl
	
	// 保存初始版本
	version := &TemplateVersion{
		TemplateID: tmpl.ID,
		Version:    tmpl.Version,
		Content:    tmpl.Content,
		ChangeLog:  "Initial version",
		CreatedAt:  tmpl.CreatedAt,
		CreatedBy:  tmpl.CreatedBy,
	}
	
	if err := m.storage.SaveVersion(ctx, version); err != nil {
		// 记录错误但不阻止创建
		fmt.Printf("warning: failed to save initial version: %v\n", err)
	}
	
	m.versions[tmpl.ID] = []*TemplateVersion{version}
	
	return nil
}

// Get 获取模板
func (m *TemplateManager) Get(ctx context.Context, id string) (*PromptTemplate, error) {
	m.mu.RLock()
	tmpl, ok := m.templates[id]
	m.mu.RUnlock()
	
	if ok {
		return tmpl, nil
	}
	
	// 从存储加载
	tmpl, err := m.storage.Load(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load template: %w", err)
	}
	
	m.mu.Lock()
	m.templates[id] = tmpl
	m.mu.Unlock()
	
	return tmpl, nil
}

// GetByName 通过名称获取模板
func (m *TemplateManager) GetByName(ctx context.Context, name string) (*PromptTemplate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, tmpl := range m.templates {
		if tmpl.Name == name {
			return tmpl, nil
		}
	}
	
	return nil, fmt.Errorf("template not found: %s", name)
}

// Update 更新模板
func (m *TemplateManager) Update(ctx context.Context, tmpl *PromptTemplate) error {
	if err := m.validateTemplate(tmpl); err != nil {
		return fmt.Errorf("validate template: %w", err)
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	existing, ok := m.templates[tmpl.ID]
	if !ok {
		return fmt.Errorf("template not found: %s", tmpl.ID)
	}
	
	// 更新版本
	newVersion := incrementVersion(existing.Version)
	oldVersion := existing.Version
	
	tmpl.UpdatedAt = time.Now()
	tmpl.Version = newVersion
	tmpl.CreatedAt = existing.CreatedAt
	tmpl.CreatedBy = existing.CreatedBy
	
	if err := m.storage.Save(ctx, tmpl); err != nil {
		return fmt.Errorf("save template: %w", err)
	}
	
	m.templates[tmpl.ID] = tmpl
	
	// 清除缓存
	delete(m.cache, tmpl.ID)
	
	// 保存版本历史
	version := &TemplateVersion{
		TemplateID: tmpl.ID,
		Version:    newVersion,
		Content:    tmpl.Content,
		ChangeLog:  fmt.Sprintf("Updated from version %s", oldVersion),
		CreatedAt:  tmpl.UpdatedAt,
		CreatedBy:  tmpl.CreatedBy,
	}
	
	if err := m.storage.SaveVersion(ctx, version); err != nil {
		fmt.Printf("warning: failed to save version: %v\n", err)
	}
	
	m.versions[tmpl.ID] = append(m.versions[tmpl.ID], version)
	
	return nil
}

// Delete 删除模板
func (m *TemplateManager) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, ok := m.templates[id]; !ok {
		return fmt.Errorf("template not found: %s", id)
	}
	
	if err := m.storage.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete template: %w", err)
	}
	
	delete(m.templates, id)
	delete(m.versions, id)
	delete(m.cache, id)
	
	return nil
}

// List 列出模板
func (m *TemplateManager) List(ctx context.Context, filter *TemplateFilter) ([]*PromptTemplate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var result []*PromptTemplate
	
	for _, tmpl := range m.templates {
		if filter != nil && !m.matchFilter(tmpl, filter) {
			continue
		}
		result = append(result, tmpl)
	}
	
	// 按更新时间排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].UpdatedAt.After(result[j].UpdatedAt)
	})
	
	return result, nil
}

// Render 渲染模板
func (m *TemplateManager) Render(ctx context.Context, id string, variables map[string]interface{}, opts *RenderOptions) (string, error) {
	tmpl, err := m.Get(ctx, id)
	if err != nil {
		return "", err
	}
	
	return m.RenderTemplate(tmpl, variables, opts)
}

// RenderTemplate 渲染指定模板
func (m *TemplateManager) RenderTemplate(tmpl *PromptTemplate, variables map[string]interface{}, opts *RenderOptions) (string, error) {
	if opts == nil {
		opts = &RenderOptions{
			Strict:                  false,
			MissingVariableHandling: "default",
		}
	}
	
	// 验证变量
	if err := m.validateVariables(tmpl, variables, opts); err != nil {
		return "", err
	}
	
	// 准备变量
	mergedVars := m.mergeVariables(tmpl, variables)
	
	// 渲染模板
	content := tmpl.Content
	
	// 使用Go模板引擎
	t, err := template.New("prompt").Parse(content)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}
	
	var buf bytes.Buffer
	if err := t.Execute(&buf, mergedVars); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	
	result := buf.String()
	
	// 应用转义函数
	if opts.EscapeFunc != nil {
		result = opts.EscapeFunc(result)
	}
	
	return result, nil
}

// GetVersions 获取模板版本历史
func (m *TemplateManager) GetVersions(ctx context.Context, templateID string) ([]*TemplateVersion, error) {
	m.mu.RLock()
	versions, ok := m.versions[templateID]
	m.mu.RUnlock()
	
	if ok {
		return versions, nil
	}
	
	// 从存储加载
	versions, err := m.storage.LoadVersions(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("load versions: %w", err)
	}
	
	m.mu.Lock()
	m.versions[templateID] = versions
	m.mu.Unlock()
	
	return versions, nil
}

// Rollback 回滚到指定版本
func (m *TemplateManager) Rollback(ctx context.Context, templateID, version string) error {
	versions, err := m.GetVersions(ctx, templateID)
	if err != nil {
		return err
	}
	
	var targetVersion *TemplateVersion
	for _, v := range versions {
		if v.Version == version {
			targetVersion = v
			break
		}
	}
	
	if targetVersion == nil {
		return fmt.Errorf("version not found: %s", version)
	}
	
	tmpl, err := m.Get(ctx, templateID)
	if err != nil {
		return err
	}
	
	tmpl.Content = targetVersion.Content
	tmpl.Version = incrementVersion(tmpl.Version)
	tmpl.UpdatedAt = time.Now()
	
	return m.Update(ctx, tmpl)
}

// Duplicate 复制模板
func (m *TemplateManager) Duplicate(ctx context.Context, id string, newName string) (*PromptTemplate, error) {
	tmpl, err := m.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	
	newTmpl := &PromptTemplate{
		Name:        newName,
		Description: tmpl.Description,
		Content:     tmpl.Content,
		Type:        tmpl.Type,
		Variables:   tmpl.Variables,
		Examples:    tmpl.Examples,
		Tags:        tmpl.Tags,
		Enabled:     true,
		Metadata:    tmpl.Metadata,
		CreatedBy:   tmpl.CreatedBy,
	}
	
	if err := m.Create(ctx, newTmpl); err != nil {
		return nil, err
	}
	
	return newTmpl, nil
}

// Export 导出模板
func (m *TemplateManager) Export(ctx context.Context, ids []string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var templates []*PromptTemplate
	for _, id := range ids {
		tmpl, ok := m.templates[id]
		if !ok {
			return nil, fmt.Errorf("template not found: %s", id)
		}
		templates = append(templates, tmpl)
	}
	
	return json.MarshalIndent(templates, "", "  ")
}

// Import 导入模板
func (m *TemplateManager) Import(ctx context.Context, data []byte, overwrite bool) error {
	var templates []*PromptTemplate
	if err := json.Unmarshal(data, &templates); err != nil {
		return fmt.Errorf("unmarshal templates: %w", err)
	}
	
	for _, tmpl := range templates {
		existing, err := m.GetByName(ctx, tmpl.Name)
		if err == nil {
			if overwrite {
				tmpl.ID = existing.ID
				if err := m.Update(ctx, tmpl); err != nil {
					return fmt.Errorf("update template %s: %w", tmpl.Name, err)
				}
			}
			continue
		}
		
		if err := m.Create(ctx, tmpl); err != nil {
			return fmt.Errorf("create template %s: %w", tmpl.Name, err)
		}
	}
	
	return nil
}

// validateTemplate 验证模板
func (m *TemplateManager) validateTemplate(tmpl *PromptTemplate) error {
	if tmpl.Name == "" {
		return errors.New("template name is required")
	}
	
	if tmpl.Content == "" {
		return errors.New("template content is required")
	}
	
	// 验证模板语法
	_, err := template.New("validate").Parse(tmpl.Content)
	if err != nil {
		return fmt.Errorf("invalid template syntax: %w", err)
	}
	
	// 验证变量定义
	for _, v := range tmpl.Variables {
		if v.Name == "" {
			return errors.New("variable name is required")
		}
		if v.Type == "" {
			v.Type = "string"
		}
	}
	
	return nil
}

// validateVariables 验证变量
func (m *TemplateManager) validateVariables(tmpl *PromptTemplate, variables map[string]interface{}, opts *RenderOptions) error {
	for _, v := range tmpl.Variables {
		value, exists := variables[v.Name]
		
		if !exists {
			if v.Required && v.Default == nil {
				if opts.Strict {
					return fmt.Errorf("required variable missing: %s", v.Name)
				}
				continue
			}
			continue
		}
		
		// 类型检查
		if err := m.validateVariableType(v, value); err != nil {
			return err
		}
		
		// 验证规则
		if v.Validation != "" {
			matched, err := regexp.MatchString(v.Validation, fmt.Sprintf("%v", value))
			if err != nil {
				return fmt.Errorf("invalid validation pattern: %w", err)
			}
			if !matched {
				return fmt.Errorf("variable %s does not match validation pattern", v.Name)
			}
		}
		
		// 枚举检查
		if len(v.Enum) > 0 {
			strValue := fmt.Sprintf("%v", value)
			found := false
			for _, e := range v.Enum {
				if e == strValue {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("variable %s must be one of: %v", v.Name, v.Enum)
			}
		}
	}
	
	return nil
}

// validateVariableType 验证变量类型
func (m *TemplateManager) validateVariableType(v TemplateVariable, value interface{}) error {
	switch v.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("variable %s must be string", v.Name)
		}
	case "number":
		switch value.(type) {
		case int, int64, float64:
		default:
			return fmt.Errorf("variable %s must be number", v.Name)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("variable %s must be boolean", v.Name)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("variable %s must be array", v.Name)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("variable %s must be object", v.Name)
		}
	}
	return nil
}

// mergeVariables 合并变量
func (m *TemplateManager) mergeVariables(tmpl *PromptTemplate, variables map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	// 添加默认值
	for _, v := range tmpl.Variables {
		if v.Default != nil {
			result[v.Name] = v.Default
		}
	}
	
	// 覆盖传入的变量
	for k, v := range variables {
		result[k] = v
	}
	
	return result
}

// matchFilter 匹配过滤条件
func (m *TemplateManager) matchFilter(tmpl *PromptTemplate, filter *TemplateFilter) bool {
	if filter.Name != "" && !strings.Contains(tmpl.Name, filter.Name) {
		return false
	}
	
	if filter.Type != "" && tmpl.Type != filter.Type {
		return false
	}
	
	if len(filter.Tags) > 0 {
		matched := false
		for _, tag := range filter.Tags {
			for _, t := range tmpl.Tags {
				if t == tag {
					matched = true
					break
				}
			}
		}
		if !matched {
			return false
		}
	}
	
	if filter.Enabled != nil && tmpl.Enabled != *filter.Enabled {
		return false
	}
	
	if filter.CreatedBy != "" && tmpl.CreatedBy != filter.CreatedBy {
		return false
	}
	
	if filter.CreatedAfter != nil && tmpl.CreatedAt.Before(*filter.CreatedAfter) {
		return false
	}
	
	if filter.CreatedBefore != nil && tmpl.CreatedAt.After(*filter.CreatedBefore) {
		return false
	}
	
	return true
}

// MemoryTemplateStorage 内存模板存储
type MemoryTemplateStorage struct {
	templates map[string]*PromptTemplate
	versions  map[string][]*TemplateVersion
	mu        sync.RWMutex
}

// NewMemoryTemplateStorage 创建内存存储
func NewMemoryTemplateStorage() *MemoryTemplateStorage {
	return &MemoryTemplateStorage{
		templates: make(map[string]*PromptTemplate),
		versions:  make(map[string][]*TemplateVersion),
	}
}

// Save 保存模板
func (s *MemoryTemplateStorage) Save(ctx context.Context, tmpl *PromptTemplate) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.templates[tmpl.ID] = tmpl
	return nil
}

// Load 加载模板
func (s *MemoryTemplateStorage) Load(ctx context.Context, id string) (*PromptTemplate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tmpl, ok := s.templates[id]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", id)
	}
	return tmpl, nil
}

// LoadAll 加载所有模板
func (s *MemoryTemplateStorage) LoadAll(ctx context.Context) ([]*PromptTemplate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var templates []*PromptTemplate
	for _, tmpl := range s.templates {
		templates = append(templates, tmpl)
	}
	return templates, nil
}

// Delete 删除模板
func (s *MemoryTemplateStorage) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.templates, id)
	delete(s.versions, id)
	return nil
}

// SaveVersion 保存版本
func (s *MemoryTemplateStorage) SaveVersion(ctx context.Context, version *TemplateVersion) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.versions[version.TemplateID] = append(s.versions[version.TemplateID], version)
	return nil
}

// LoadVersions 加载版本历史
func (s *MemoryTemplateStorage) LoadVersions(ctx context.Context, templateID string) ([]*TemplateVersion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	versions, ok := s.versions[templateID]
	if !ok {
		return nil, fmt.Errorf("versions not found: %s", templateID)
	}
	return versions, nil
}

// 辅助函数
func generateTemplateID() string {
	return fmt.Sprintf("tmpl_%d", time.Now().UnixNano())
}

func incrementVersion(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return "1.0.1"
	}
	
	minor := 0
	fmt.Sscanf(parts[1], "%d", &minor)
	return fmt.Sprintf("%s.%d.%s", parts[0], minor+1, parts[2])
}

// 预定义模板类型
const (
	TemplateTypeChat      = "chat"
	TemplateTypeCompletion = "completion"
	TemplateTypeSystem    = "system"
	TemplateTypeFunction  = "function"
)

// 内置模板
var BuiltInTemplates = []*PromptTemplate{
	{
		ID:          "builtin-chat-default",
		Name:        "default-chat",
		Description: "Default chat prompt template",
		Content:     "You are a helpful AI assistant for a new energy monitoring system. Please help the user with their questions about energy monitoring, device management, and data analysis.\n\nUser: {{.query}}\n\nAssistant:",
		Type:        TemplateTypeChat,
		Variables: []TemplateVariable{
			{Name: "query", Type: "string", Required: true, Description: "User query"},
		},
		Enabled: true,
	},
	{
		ID:          "builtin-alarm-analysis",
		Name:        "alarm-analysis",
		Description: "Alarm analysis prompt template",
		Content:     "Analyze the following alarm data and provide insights:\n\nAlarm Details:\n- Device: {{.device_name}}\n- Type: {{.alarm_type}}\n- Severity: {{.severity}}\n- Time: {{.alarm_time}}\n- Description: {{.description}}\n\nPlease provide:\n1. Root cause analysis\n2. Recommended actions\n3. Prevention measures",
		Type:        TemplateTypeChat,
		Variables: []TemplateVariable{
			{Name: "device_name", Type: "string", Required: true},
			{Name: "alarm_type", Type: "string", Required: true},
			{Name: "severity", Type: "string", Required: true},
			{Name: "alarm_time", Type: "string", Required: true},
			{Name: "description", Type: "string", Required: true},
		},
		Enabled: true,
	},
	{
		ID:          "builtin-report-generation",
		Name:        "report-generation",
		Description: "Report generation prompt template",
		Content:     "Generate a comprehensive energy monitoring report based on the following data:\n\nStation: {{.station_name}}\nPeriod: {{.period}}\n\nData Summary:\n{{.data_summary}}\n\nPlease include:\n1. Executive Summary\n2. Key Metrics Analysis\n3. Trend Analysis\n4. Recommendations\n5. Conclusion",
		Type:        TemplateTypeCompletion,
		Variables: []TemplateVariable{
			{Name: "station_name", Type: "string", Required: true},
			{Name: "period", Type: "string", Required: true},
			{Name: "data_summary", Type: "string", Required: true},
		},
		Enabled: true,
	},
}
