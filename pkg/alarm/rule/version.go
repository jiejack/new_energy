package rule

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// VersionStatus 版本状态
type VersionStatus string

const (
	VersionStatusActive   VersionStatus = "active"   // 活跃版本
	VersionStatusArchived VersionStatus = "archived" // 已归档
	VersionStatusRollback VersionStatus = "rollback" // 回滚版本
)

// RuleVersion 规则版本
type RuleVersion struct {
	ID          string        `json:"id"`                    // 版本ID
	RuleID      string        `json:"rule_id"`               // 规则ID
	Version     string        `json:"version"`               // 版本号
	Rule        *RuleDSL      `json:"rule"`                  // 规则内容
	Status      VersionStatus `json:"status"`                // 版本状态
	Description string        `json:"description,omitempty"` // 版本描述
	ChangeLog   string        `json:"change_log,omitempty"`  // 变更日志
	CreatedAt   time.Time     `json:"created_at"`            // 创建时间
	CreatedBy   string        `json:"created_by"`            // 创建人
	Tags        []string      `json:"tags,omitempty"`        // 标签
}

// NewRuleVersion 创建新的规则版本
func NewRuleVersion(rule *RuleDSL, description, changeLog, createdBy string) *RuleVersion {
	return &RuleVersion{
		ID:          generateVersionID(rule.ID, rule.Version),
		RuleID:      rule.ID,
		Version:     rule.Version,
		Rule:        rule,
		Status:      VersionStatusActive,
		Description: description,
		ChangeLog:   changeLog,
		CreatedAt:   time.Now(),
		CreatedBy:   createdBy,
		Tags:        []string{},
	}
}

// VersionManager 版本管理器
type VersionManager struct {
	versions map[string][]*RuleVersion // ruleID -> versions
	mu       sync.RWMutex
}

// NewVersionManager 创建版本管理器
func NewVersionManager() *VersionManager {
	return &VersionManager{
		versions: make(map[string][]*RuleVersion),
	}
}

// CreateVersion 创建版本
func (vm *VersionManager) CreateVersion(rule *RuleDSL, description, changeLog, createdBy string) (*RuleVersion, error) {
	if err := rule.Validate(); err != nil {
		return nil, fmt.Errorf("invalid rule: %w", err)
	}

	version := NewRuleVersion(rule, description, changeLog, createdBy)

	vm.mu.Lock()
	defer vm.mu.Unlock()

	// 检查版本是否已存在
	for _, v := range vm.versions[rule.ID] {
		if v.Version == rule.Version {
			return nil, fmt.Errorf("version %s already exists for rule %s", rule.Version, rule.ID)
		}
	}

	// 将之前的活跃版本归档
	for _, v := range vm.versions[rule.ID] {
		if v.Status == VersionStatusActive {
			v.Status = VersionStatusArchived
		}
	}

	// 添加新版本
	vm.versions[rule.ID] = append(vm.versions[rule.ID], version)

	return version, nil
}

// GetVersion 获取指定版本
func (vm *VersionManager) GetVersion(ruleID, version string) (*RuleVersion, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	versions, exists := vm.versions[ruleID]
	if !exists {
		return nil, fmt.Errorf("rule %s not found", ruleID)
	}

	for _, v := range versions {
		if v.Version == version {
			return v, nil
		}
	}

	return nil, fmt.Errorf("version %s not found for rule %s", version, ruleID)
}

// GetActiveVersion 获取活跃版本
func (vm *VersionManager) GetActiveVersion(ruleID string) (*RuleVersion, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	versions, exists := vm.versions[ruleID]
	if !exists {
		return nil, fmt.Errorf("rule %s not found", ruleID)
	}

	for _, v := range versions {
		if v.Status == VersionStatusActive {
			return v, nil
		}
	}

	return nil, fmt.Errorf("no active version found for rule %s", ruleID)
}

// GetHistory 获取历史版本
func (vm *VersionManager) GetHistory(ruleID string) ([]*RuleVersion, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	versions, exists := vm.versions[ruleID]
	if !exists {
		return nil, fmt.Errorf("rule %s not found", ruleID)
	}

	// 返回副本
	result := make([]*RuleVersion, len(versions))
	copy(result, versions)

	// 按创建时间倒序排序
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].CreatedAt.Before(result[j].CreatedAt) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result, nil
}

// Rollback 版本回滚
func (vm *VersionManager) Rollback(ruleID, targetVersion string) (*RuleVersion, error) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	versions, exists := vm.versions[ruleID]
	if !exists {
		return nil, fmt.Errorf("rule %s not found", ruleID)
	}

	// 查找目标版本
	var targetVer *RuleVersion
	for _, v := range versions {
		if v.Version == targetVersion {
			targetVer = v
			break
		}
	}

	if targetVer == nil {
		return nil, fmt.Errorf("version %s not found for rule %s", targetVersion, ruleID)
	}

	// 创建回滚版本
	newVersion := fmt.Sprintf("%s-rollback-%d", targetVersion, time.Now().Unix())
	rollbackRule := *targetVer.Rule
	rollbackRule.Version = newVersion
	rollbackRule.UpdatedAt = time.Now()

	rollback := &RuleVersion{
		ID:          generateVersionID(ruleID, newVersion),
		RuleID:      ruleID,
		Version:     newVersion,
		Rule:        &rollbackRule,
		Status:      VersionStatusActive,
		Description: fmt.Sprintf("Rollback to version %s", targetVersion),
		ChangeLog:   fmt.Sprintf("Rolled back from version %s", targetVersion),
		CreatedAt:   time.Now(),
		CreatedBy:   "system",
		Tags:        []string{"rollback"},
	}

	// 将之前的活跃版本归档
	for _, v := range versions {
		if v.Status == VersionStatusActive {
			v.Status = VersionStatusArchived
		}
	}

	// 添加回滚版本
	vm.versions[ruleID] = append(vm.versions[ruleID], rollback)

	return rollback, nil
}

// CompareVersions 版本对比
func (vm *VersionManager) CompareVersions(ruleID, version1, version2 string) (*VersionDiff, error) {
	v1, err := vm.GetVersion(ruleID, version1)
	if err != nil {
		return nil, fmt.Errorf("failed to get version %s: %w", version1, err)
	}

	v2, err := vm.GetVersion(ruleID, version2)
	if err != nil {
		return nil, fmt.Errorf("failed to get version %s: %w", version2, err)
	}

	return compareRules(v1.Rule, v2.Rule), nil
}

// DeleteVersion 删除版本
func (vm *VersionManager) DeleteVersion(ruleID, version string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	versions, exists := vm.versions[ruleID]
	if !exists {
		return fmt.Errorf("rule %s not found", ruleID)
	}

	for i, v := range versions {
		if v.Version == version {
			// 不能删除活跃版本
			if v.Status == VersionStatusActive {
				return fmt.Errorf("cannot delete active version %s", version)
			}
			// 删除版本
			vm.versions[ruleID] = append(versions[:i], versions[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("version %s not found for rule %s", version, ruleID)
}

// ListAllVersions 列出所有规则的所有版本
func (vm *VersionManager) ListAllVersions() map[string][]*RuleVersion {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	result := make(map[string][]*RuleVersion)
	for ruleID, versions := range vm.versions {
		result[ruleID] = make([]*RuleVersion, len(versions))
		copy(result[ruleID], versions)
	}
	return result
}

// VersionDiff 版本差异
type VersionDiff struct {
	Version1     string        `json:"version1"`
	Version2     string        `json:"version2"`
	HasChanges   bool          `json:"has_changes"`
	Changes      []FieldChange `json:"changes"`
	ConditionDiff string       `json:"condition_diff,omitempty"`
}

// FieldChange 字段变更
type FieldChange struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
}

// compareRules 对比两个规则
func compareRules(rule1, rule2 *RuleDSL) *VersionDiff {
	diff := &VersionDiff{
		Version1:   rule1.Version,
		Version2:   rule2.Version,
		HasChanges: false,
		Changes:    []FieldChange{},
	}

	// 对比基本字段
	if rule1.Name != rule2.Name {
		diff.HasChanges = true
		diff.Changes = append(diff.Changes, FieldChange{
			Field:    "name",
			OldValue: rule1.Name,
			NewValue: rule2.Name,
		})
	}

	if rule1.Description != rule2.Description {
		diff.HasChanges = true
		diff.Changes = append(diff.Changes, FieldChange{
			Field:    "description",
			OldValue: rule1.Description,
			NewValue: rule2.Description,
		})
	}

	if rule1.Priority != rule2.Priority {
		diff.HasChanges = true
		diff.Changes = append(diff.Changes, FieldChange{
			Field:    "priority",
			OldValue: rule1.Priority,
			NewValue: rule2.Priority,
		})
	}

	if rule1.Enabled != rule2.Enabled {
		diff.HasChanges = true
		diff.Changes = append(diff.Changes, FieldChange{
			Field:    "enabled",
			OldValue: rule1.Enabled,
			NewValue: rule2.Enabled,
		})
	}

	// 对比条件
	cond1 := rule1.Condition.String()
	cond2 := rule2.Condition.String()
	if cond1 != cond2 {
		diff.HasChanges = true
		diff.ConditionDiff = fmt.Sprintf("Condition changed from '%s' to '%s'", cond1, cond2)
	}

	// 对比标签
	if !equalStringSlices(rule1.Tags, rule2.Tags) {
		diff.HasChanges = true
		diff.Changes = append(diff.Changes, FieldChange{
			Field:    "tags",
			OldValue: rule1.Tags,
			NewValue: rule2.Tags,
		})
	}

	// 对比动作
	if len(rule1.Actions) != len(rule2.Actions) {
		diff.HasChanges = true
		diff.Changes = append(diff.Changes, FieldChange{
			Field:    "actions_count",
			OldValue: len(rule1.Actions),
			NewValue: len(rule2.Actions),
		})
	}

	return diff
}

// equalStringSlices 判断两个字符串切片是否相等
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// generateVersionID 生成版本ID
func generateVersionID(ruleID, version string) string {
	return fmt.Sprintf("%s-%s-%d", ruleID, version, time.Now().UnixNano())
}

// ExportVersion 导出版本为JSON
func (v *RuleVersion) ExportVersion() (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to export version: %w", err)
	}
	return string(data), nil
}

// ImportVersion 从JSON导入版本
func ImportVersion(jsonStr string) (*RuleVersion, error) {
	var version RuleVersion
	if err := json.Unmarshal([]byte(jsonStr), &version); err != nil {
		return nil, fmt.Errorf("failed to import version: %w", err)
	}
	return &version, nil
}

// ArchiveVersion 归档版本
func (vm *VersionManager) ArchiveVersion(ruleID, version string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	versions, exists := vm.versions[ruleID]
	if !exists {
		return fmt.Errorf("rule %s not found", ruleID)
	}

	for _, v := range versions {
		if v.Version == version {
			if v.Status == VersionStatusActive {
				return fmt.Errorf("cannot archive active version %s", version)
			}
			v.Status = VersionStatusArchived
			return nil
		}
	}

	return fmt.Errorf("version %s not found for rule %s", version, ruleID)
}

// GetVersionsByTag 根据标签获取版本
func (vm *VersionManager) GetVersionsByTag(ruleID, tag string) ([]*RuleVersion, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	versions, exists := vm.versions[ruleID]
	if !exists {
		return nil, fmt.Errorf("rule %s not found", ruleID)
	}

	result := make([]*RuleVersion, 0)
	for _, v := range versions {
		for _, t := range v.Tags {
			if t == tag {
				result = append(result, v)
				break
			}
		}
	}

	return result, nil
}

// AddTagToVersion 为版本添加标签
func (vm *VersionManager) AddTagToVersion(ruleID, version, tag string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	versions, exists := vm.versions[ruleID]
	if !exists {
		return fmt.Errorf("rule %s not found", ruleID)
	}

	for _, v := range versions {
		if v.Version == version {
			// 检查标签是否已存在
			for _, t := range v.Tags {
				if t == tag {
					return nil // 标签已存在
				}
			}
			v.Tags = append(v.Tags, tag)
			return nil
		}
	}

	return fmt.Errorf("version %s not found for rule %s", version, ruleID)
}

// RemoveTagFromVersion 从版本移除标签
func (vm *VersionManager) RemoveTagFromVersion(ruleID, version, tag string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	versions, exists := vm.versions[ruleID]
	if !exists {
		return fmt.Errorf("rule %s not found", ruleID)
	}

	for _, v := range versions {
		if v.Version == version {
			for i, t := range v.Tags {
				if t == tag {
					v.Tags = append(v.Tags[:i], v.Tags[i+1:]...)
					return nil
				}
			}
			return fmt.Errorf("tag %s not found in version %s", tag, version)
		}
	}

	return fmt.Errorf("version %s not found for rule %s", version, ruleID)
}

// GetVersionCount 获取版本数量
func (vm *VersionManager) GetVersionCount(ruleID string) int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	versions, exists := vm.versions[ruleID]
	if !exists {
		return 0
	}
	return len(versions)
}

// SearchVersions 搜索版本
func (vm *VersionManager) SearchVersions(ruleID string, query string) ([]*RuleVersion, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	versions, exists := vm.versions[ruleID]
	if !exists {
		return nil, fmt.Errorf("rule %s not found", ruleID)
	}

	result := make([]*RuleVersion, 0)
	query = strings.ToLower(query)

	for _, v := range versions {
		// 搜索描述、变更日志、标签
		if strings.Contains(strings.ToLower(v.Description), query) ||
			strings.Contains(strings.ToLower(v.ChangeLog), query) ||
			strings.Contains(strings.ToLower(v.Version), query) {
			result = append(result, v)
			continue
		}

		// 搜索标签
		for _, tag := range v.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				result = append(result, v)
				break
			}
		}
	}

	return result, nil
}
