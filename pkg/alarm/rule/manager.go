package rule

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// RuleManager 规则管理器
type RuleManager struct {
	engine         *Engine
	versionManager *VersionManager
	rules          map[string]*RuleDSL
	mu             sync.RWMutex
	storage        RuleStorage
}

// RuleStorage 规则存储接口
type RuleStorage interface {
	Save(ctx context.Context, rule *RuleDSL) error
	Load(ctx context.Context, ruleID string) (*RuleDSL, error)
	LoadAll(ctx context.Context) ([]*RuleDSL, error)
	Delete(ctx context.Context, ruleID string) error
}

// NewRuleManager 创建规则管理器
func NewRuleManager(engine *Engine, storage RuleStorage) *RuleManager {
	return &RuleManager{
		engine:         engine,
		versionManager: NewVersionManager(),
		rules:          make(map[string]*RuleDSL),
		storage:        storage,
	}
}

// CreateRule 创建规则
func (rm *RuleManager) CreateRule(ctx context.Context, rule *RuleDSL) error {
	if err := rule.Validate(); err != nil {
		return fmt.Errorf("invalid rule: %w", err)
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	// 检查规则是否已存在
	if _, exists := rm.rules[rule.ID]; exists {
		return fmt.Errorf("rule %s already exists", rule.ID)
	}

	// 设置创建时间
	now := time.Now()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	// 保存到存储
	if rm.storage != nil {
		if err := rm.storage.Save(ctx, rule); err != nil {
			return fmt.Errorf("failed to save rule: %w", err)
		}
	}

	// 添加到内存
	rm.rules[rule.ID] = rule

	// 添加到引擎
	if rm.engine != nil {
		if err := rm.engine.AddRule(rule); err != nil {
			return fmt.Errorf("failed to add rule to engine: %w", err)
		}
	}

	// 创建版本
	if _, err := rm.versionManager.CreateVersion(rule, "Initial version", "Rule created", rule.CreatedBy); err != nil {
		return fmt.Errorf("failed to create version: %w", err)
	}

	return nil
}

// GetRule 获取规则
func (rm *RuleManager) GetRule(ctx context.Context, ruleID string) (*RuleDSL, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rule, exists := rm.rules[ruleID]
	if !exists {
		return nil, fmt.Errorf("rule %s not found", ruleID)
	}

	return rule, nil
}

// UpdateRule 更新规则
func (rm *RuleManager) UpdateRule(ctx context.Context, rule *RuleDSL) error {
	if err := rule.Validate(); err != nil {
		return fmt.Errorf("invalid rule: %w", err)
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	// 检查规则是否存在
	oldRule, exists := rm.rules[rule.ID]
	if !exists {
		return fmt.Errorf("rule %s not found", rule.ID)
	}

	// 保留创建信息
	rule.CreatedAt = oldRule.CreatedAt
	rule.CreatedBy = oldRule.CreatedBy
	rule.UpdatedAt = time.Now()

	// 保存到存储
	if rm.storage != nil {
		if err := rm.storage.Save(ctx, rule); err != nil {
			return fmt.Errorf("failed to save rule: %w", err)
		}
	}

	// 更新内存
	rm.rules[rule.ID] = rule

	// 更新引擎
	if rm.engine != nil {
		rm.engine.RemoveRule(rule.ID)
		if err := rm.engine.AddRule(rule); err != nil {
			return fmt.Errorf("failed to update rule in engine: %w", err)
		}
	}

	// 创建新版本
	if _, err := rm.versionManager.CreateVersion(rule, "Rule updated", "Rule updated", rule.UpdatedBy); err != nil {
		return fmt.Errorf("failed to create version: %w", err)
	}

	return nil
}

// DeleteRule 删除规则
func (rm *RuleManager) DeleteRule(ctx context.Context, ruleID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// 检查规则是否存在
	if _, exists := rm.rules[ruleID]; !exists {
		return fmt.Errorf("rule %s not found", ruleID)
	}

	// 从存储删除
	if rm.storage != nil {
		if err := rm.storage.Delete(ctx, ruleID); err != nil {
			return fmt.Errorf("failed to delete rule from storage: %w", err)
		}
	}

	// 从内存删除
	delete(rm.rules, ruleID)

	// 从引擎删除
	if rm.engine != nil {
		rm.engine.RemoveRule(ruleID)
	}

	return nil
}

// ListRules 列出所有规则
func (rm *RuleManager) ListRules(ctx context.Context) ([]*RuleDSL, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rules := make([]*RuleDSL, 0, len(rm.rules))
	for _, rule := range rm.rules {
		rules = append(rules, rule)
	}

	// 按优先级排序
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority
	})

	return rules, nil
}

// ListRulesByTag 根据标签列出规则
func (rm *RuleManager) ListRulesByTag(ctx context.Context, tag string) ([]*RuleDSL, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rules := make([]*RuleDSL, 0)
	for _, rule := range rm.rules {
		for _, t := range rule.Tags {
			if t == tag {
				rules = append(rules, rule)
				break
			}
		}
	}

	// 按优先级排序
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority
	})

	return rules, nil
}

// SearchRules 搜索规则
func (rm *RuleManager) SearchRules(ctx context.Context, query string) ([]*RuleDSL, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	query = strings.ToLower(query)
	rules := make([]*RuleDSL, 0)

	for _, rule := range rm.rules {
		// 搜索名称、描述、ID
		if strings.Contains(strings.ToLower(rule.Name), query) ||
			strings.Contains(strings.ToLower(rule.Description), query) ||
			strings.Contains(strings.ToLower(rule.ID), query) {
			rules = append(rules, rule)
			continue
		}

		// 搜索标签
		for _, tag := range rule.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				rules = append(rules, rule)
				break
			}
		}
	}

	return rules, nil
}

// EnableRule 启用规则
func (rm *RuleManager) EnableRule(ctx context.Context, ruleID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rule, exists := rm.rules[ruleID]
	if !exists {
		return fmt.Errorf("rule %s not found", ruleID)
	}

	rule.Enabled = true
	rule.UpdatedAt = time.Now()

	// 保存到存储
	if rm.storage != nil {
		if err := rm.storage.Save(ctx, rule); err != nil {
			return fmt.Errorf("failed to save rule: %w", err)
		}
	}

	return nil
}

// DisableRule 禁用规则
func (rm *RuleManager) DisableRule(ctx context.Context, ruleID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rule, exists := rm.rules[ruleID]
	if !exists {
		return fmt.Errorf("rule %s not found", ruleID)
	}

	rule.Enabled = false
	rule.UpdatedAt = time.Now()

	// 保存到存储
	if rm.storage != nil {
		if err := rm.storage.Save(ctx, rule); err != nil {
			return fmt.Errorf("failed to save rule: %w", err)
		}
	}

	return nil
}

// ValidateRule 验证规则
func (rm *RuleManager) ValidateRule(rule *RuleDSL) error {
	return rule.Validate()
}

// ImportRules 从JSON导入规则
func (rm *RuleManager) ImportRules(ctx context.Context, jsonData []byte) ([]*RuleDSL, error) {
	var rules []*RuleDSL
	if err := json.Unmarshal(jsonData, &rules); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
	}

	imported := make([]*RuleDSL, 0, len(rules))
	for _, rule := range rules {
		if err := rm.CreateRule(ctx, rule); err != nil {
			return nil, fmt.Errorf("failed to import rule %s: %w", rule.ID, err)
		}
		imported = append(imported, rule)
	}

	return imported, nil
}

// ExportRules 导出规则为JSON
func (rm *RuleManager) ExportRules(ctx context.Context, ruleIDs []string) ([]byte, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rules := make([]*RuleDSL, 0, len(ruleIDs))
	for _, id := range ruleIDs {
		rule, exists := rm.rules[id]
		if !exists {
			return nil, fmt.Errorf("rule %s not found", id)
		}
		rules = append(rules, rule)
	}

	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rules: %w", err)
	}

	return data, nil
}

// ExportAllRules 导出所有规则
func (rm *RuleManager) ExportAllRules(ctx context.Context) ([]byte, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rules := make([]*RuleDSL, 0, len(rm.rules))
	for _, rule := range rm.rules {
		rules = append(rules, rule)
	}

	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rules: %w", err)
	}

	return data, nil
}

// ImportFromFile 从文件导入规则
func (rm *RuleManager) ImportFromFile(ctx context.Context, filePath string) ([]*RuleDSL, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return rm.ImportRules(ctx, data)
}

// ExportToFile 导出规则到文件
func (rm *RuleManager) ExportToFile(ctx context.Context, filePath string, ruleIDs []string) error {
	data, err := rm.ExportRules(ctx, ruleIDs)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// CloneRule 克隆规则
func (rm *RuleManager) CloneRule(ctx context.Context, ruleID, newID, newName string) (*RuleDSL, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// 获取原规则
	original, exists := rm.rules[ruleID]
	if !exists {
		return nil, fmt.Errorf("rule %s not found", ruleID)
	}

	// 检查新ID是否已存在
	if _, exists := rm.rules[newID]; exists {
		return nil, fmt.Errorf("rule %s already exists", newID)
	}

	// 克隆规则
	cloned := *original
	cloned.ID = newID
	cloned.Name = newName
	cloned.Version = "1.0"
	cloned.CreatedAt = time.Now()
	cloned.UpdatedAt = time.Now()

	// 保存到存储
	if rm.storage != nil {
		if err := rm.storage.Save(ctx, &cloned); err != nil {
			return nil, fmt.Errorf("failed to save cloned rule: %w", err)
		}
	}

	// 添加到内存
	rm.rules[newID] = &cloned

	// 添加到引擎
	if rm.engine != nil {
		if err := rm.engine.AddRule(&cloned); err != nil {
			return nil, fmt.Errorf("failed to add cloned rule to engine: %w", err)
		}
	}

	// 创建版本
	if _, err := rm.versionManager.CreateVersion(&cloned, "Cloned from "+ruleID, "Rule cloned", cloned.CreatedBy); err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	return &cloned, nil
}

// GetRuleStatistics 获取规则统计信息
func (rm *RuleManager) GetRuleStatistics(ctx context.Context) (*RuleStatistics, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	stats := &RuleStatistics{
		Total:   len(rm.rules),
		Enabled: 0,
		Disabled: 0,
		ByPriority: make(map[int]int),
		ByTag:     make(map[string]int),
	}

	for _, rule := range rm.rules {
		if rule.Enabled {
			stats.Enabled++
		} else {
			stats.Disabled++
		}

		stats.ByPriority[rule.Priority]++

		for _, tag := range rule.Tags {
			stats.ByTag[tag]++
		}
	}

	return stats, nil
}

// RuleStatistics 规则统计信息
type RuleStatistics struct {
	Total       int            `json:"total"`
	Enabled     int            `json:"enabled"`
	Disabled    int            `json:"disabled"`
	ByPriority  map[int]int    `json:"by_priority"`
	ByTag       map[string]int `json:"by_tag"`
}

// GetVersionManager 获取版本管理器
func (rm *RuleManager) GetVersionManager() *VersionManager {
	return rm.versionManager
}

// GetEngine 获取执行引擎
func (rm *RuleManager) GetEngine() *Engine {
	return rm.engine
}

// LoadRules 从存储加载规则
func (rm *RuleManager) LoadRules(ctx context.Context) error {
	if rm.storage == nil {
		return nil
	}

	rules, err := rm.storage.LoadAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	for _, rule := range rules {
		rm.rules[rule.ID] = rule
		if rm.engine != nil {
			if err := rm.engine.AddRule(rule); err != nil {
				return fmt.Errorf("failed to add rule %s to engine: %w", rule.ID, err)
			}
		}
	}

	return nil
}

// BatchCreateRules 批量创建规则
func (rm *RuleManager) BatchCreateRules(ctx context.Context, rules []*RuleDSL) ([]*RuleDSL, []error) {
	created := make([]*RuleDSL, 0, len(rules))
	errors := make([]error, 0)

	for _, rule := range rules {
		if err := rm.CreateRule(ctx, rule); err != nil {
			errors = append(errors, fmt.Errorf("failed to create rule %s: %w", rule.ID, err))
		} else {
			created = append(created, rule)
		}
	}

	return created, errors
}

// BatchDeleteRules 批量删除规则
func (rm *RuleManager) BatchDeleteRules(ctx context.Context, ruleIDs []string) []error {
	errors := make([]error, 0)

	for _, id := range ruleIDs {
		if err := rm.DeleteRule(ctx, id); err != nil {
			errors = append(errors, fmt.Errorf("failed to delete rule %s: %w", id, err))
		}
	}

	return errors
}

// BatchEnableRules 批量启用规则
func (rm *RuleManager) BatchEnableRules(ctx context.Context, ruleIDs []string) []error {
	errors := make([]error, 0)

	for _, id := range ruleIDs {
		if err := rm.EnableRule(ctx, id); err != nil {
			errors = append(errors, fmt.Errorf("failed to enable rule %s: %w", id, err))
		}
	}

	return errors
}

// BatchDisableRules 批量禁用规则
func (rm *RuleManager) BatchDisableRules(ctx context.Context, ruleIDs []string) []error {
	errors := make([]error, 0)

	for _, id := range ruleIDs {
		if err := rm.DisableRule(ctx, id); err != nil {
			errors = append(errors, fmt.Errorf("failed to disable rule %s: %w", id, err))
		}
	}

	return errors
}

// GetRulesByPriority 按优先级范围获取规则
func (rm *RuleManager) GetRulesByPriority(ctx context.Context, minPriority, maxPriority int) ([]*RuleDSL, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rules := make([]*RuleDSL, 0)
	for _, rule := range rm.rules {
		if rule.Priority >= minPriority && rule.Priority <= maxPriority {
			rules = append(rules, rule)
		}
	}

	// 按优先级排序
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority
	})

	return rules, nil
}

// GetEnabledRules 获取所有启用的规则
func (rm *RuleManager) GetEnabledRules(ctx context.Context) ([]*RuleDSL, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rules := make([]*RuleDSL, 0)
	for _, rule := range rm.rules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}

	// 按优先级排序
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority
	})

	return rules, nil
}

// GetDisabledRules 获取所有禁用的规则
func (rm *RuleManager) GetDisabledRules(ctx context.Context) ([]*RuleDSL, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rules := make([]*RuleDSL, 0)
	for _, rule := range rm.rules {
		if !rule.Enabled {
			rules = append(rules, rule)
		}
	}

	return rules, nil
}

// CountRules 统计规则数量
func (rm *RuleManager) CountRules(ctx context.Context) int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return len(rm.rules)
}

// CountEnabledRules 统计启用的规则数量
func (rm *RuleManager) CountEnabledRules(ctx context.Context) int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	count := 0
	for _, rule := range rm.rules {
		if rule.Enabled {
			count++
		}
	}
	return count
}
