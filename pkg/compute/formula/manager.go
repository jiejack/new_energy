package formula

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// FormulaStatus 公式状态
type FormulaStatus int

const (
	StatusDraft FormulaStatus = iota
	StatusActive
	StatusInactive
	StatusDeprecated
	StatusError
)

func (s FormulaStatus) String() string {
	switch s {
	case StatusDraft:
		return "draft"
	case StatusActive:
		return "active"
	case StatusInactive:
		return "inactive"
	case StatusDeprecated:
		return "deprecated"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

// Formula 公式定义
type Formula struct {
	ID          string                 // 公式ID
	Name        string                 // 公式名称
	Description string                 // 描述
	Expression  string                 // 表达式
	Status      FormulaStatus          // 状态
	Version     string                 // 版本号
	Tags        []string               // 标签
	Category    string                 // 分类
	CreatedAt   time.Time              // 创建时间
	UpdatedAt   time.Time              // 更新时间
	CreatedBy   string                 // 创建者
	Metadata    map[string]interface{} // 元数据

	// 编译后的数据
	compiled    *CompiledFormula
	variables   []string
	functions   []string
	dependencies []string
}

// FormulaVersion 公式版本
type FormulaVersion struct {
	Version     string
	Expression  string
	ChangedAt   time.Time
	ChangedBy   string
	ChangeNote  string
}

// FormulaManager 公式管理器
type FormulaManager struct {
	formulas    map[string]*Formula
	versions    map[string][]*FormulaVersion
	executor    *Executor
	validator   *FormulaValidator
	dependency  *DependencyAnalyzer
	mu          sync.RWMutex
	config      *ManagerConfig
}

// ManagerConfig 管理器配置
type ManagerConfig struct {
	AutoCompile    bool // 自动编译
	AutoValidate   bool // 自动验证
	MaxVersions    int  // 最大版本数
	EnableCache    bool // 启用缓存
}

// DefaultManagerConfig 默认管理器配置
func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		AutoCompile:  true,
		AutoValidate: true,
		MaxVersions:  10,
		EnableCache:  true,
	}
}

// NewFormulaManager 创建公式管理器
func NewFormulaManager(config *ManagerConfig) *FormulaManager {
	if config == nil {
		config = DefaultManagerConfig()
	}

	manager := &FormulaManager{
		formulas:   make(map[string]*Formula),
		versions:   make(map[string][]*FormulaVersion),
		executor:   NewExecutor(nil),
		validator:  NewFormulaValidator(),
		dependency: NewDependencyAnalyzer(),
		config:     config,
	}

	return manager
}

// ==================== CRUD 操作 ====================

// Create 创建公式
func (m *FormulaManager) Create(formula *Formula) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查ID是否已存在
	if _, exists := m.formulas[formula.ID]; exists {
		return fmt.Errorf("formula with ID %s already exists", formula.ID)
	}

	// 设置默认值
	if formula.CreatedAt.IsZero() {
		formula.CreatedAt = time.Now()
	}
	if formula.UpdatedAt.IsZero() {
		formula.UpdatedAt = time.Now()
	}
	if formula.Status == 0 {
		formula.Status = StatusDraft
	}
	if formula.Version == "" {
		formula.Version = "1.0.0"
	}
	if formula.Metadata == nil {
		formula.Metadata = make(map[string]interface{})
	}

	// 自动编译
	if m.config.AutoCompile {
		compiled, err := m.executor.Compile(formula.Expression)
		if err != nil {
			return fmt.Errorf("compile error: %w", err)
		}
		formula.compiled = compiled
		formula.variables = compiled.GetVariables()
		formula.functions = compiled.GetFunctions()
	}

	// 自动验证
	if m.config.AutoValidate {
		if err := m.validator.Validate(formula); err != nil {
			return fmt.Errorf("validation error: %w", err)
		}
	}

	// 分析依赖
	deps, err := m.dependency.Analyze(formula)
	if err == nil {
		formula.dependencies = deps
	}

	m.formulas[formula.ID] = formula

	// 记录初始版本
	m.recordVersion(formula, "Initial version")

	return nil
}

// Get 获取公式
func (m *FormulaManager) Get(id string) (*Formula, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	formula, exists := m.formulas[id]
	if !exists {
		return nil, fmt.Errorf("formula not found: %s", id)
	}

	return formula, nil
}

// Update 更新公式
func (m *FormulaManager) Update(formula *Formula) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, exists := m.formulas[formula.ID]
	if !exists {
		return fmt.Errorf("formula not found: %s", formula.ID)
	}

	// 检查表达式是否变化
	expressionChanged := existing.Expression != formula.Expression

	// 更新时间
	formula.UpdatedAt = time.Now()
	if formula.CreatedAt.IsZero() {
		formula.CreatedAt = existing.CreatedAt
	}

	// 如果表达式变化，重新编译
	if expressionChanged && m.config.AutoCompile {
		compiled, err := m.executor.Compile(formula.Expression)
		if err != nil {
			return fmt.Errorf("compile error: %w", err)
		}
		formula.compiled = compiled
		formula.variables = compiled.GetVariables()
		formula.functions = compiled.GetFunctions()

		// 分析依赖
		deps, err := m.dependency.Analyze(formula)
		if err == nil {
			formula.dependencies = deps
		}
	} else {
		formula.compiled = existing.compiled
		formula.variables = existing.variables
		formula.functions = existing.functions
		formula.dependencies = existing.dependencies
	}

	// 验证
	if m.config.AutoValidate {
		if err := m.validator.Validate(formula); err != nil {
			return fmt.Errorf("validation error: %w", err)
		}
	}

	// 记录版本
	if expressionChanged {
		m.recordVersion(formula, "Expression updated")
	}

	m.formulas[formula.ID] = formula
	return nil
}

// Delete 删除公式
func (m *FormulaManager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.formulas[id]; !exists {
		return fmt.Errorf("formula not found: %s", id)
	}

	// 检查是否有其他公式依赖此公式
	for _, f := range m.formulas {
		for _, dep := range f.dependencies {
			if dep == id {
				return fmt.Errorf("cannot delete formula %s: formula %s depends on it", id, f.ID)
			}
		}
	}

	delete(m.formulas, id)
	delete(m.versions, id)

	return nil
}

// List 列出所有公式
func (m *FormulaManager) List() []*Formula {
	m.mu.RLock()
	defer m.mu.RUnlock()

	formulas := make([]*Formula, 0, len(m.formulas))
	for _, f := range m.formulas {
		formulas = append(formulas, f)
	}

	return formulas
}

// ListByStatus 按状态列出公式
func (m *FormulaManager) ListByStatus(status FormulaStatus) []*Formula {
	m.mu.RLock()
	defer m.mu.RUnlock()

	formulas := make([]*Formula, 0)
	for _, f := range m.formulas {
		if f.Status == status {
			formulas = append(formulas, f)
		}
	}

	return formulas
}

// ListByCategory 按分类列出公式
func (m *FormulaManager) ListByCategory(category string) []*Formula {
	m.mu.RLock()
	defer m.mu.RUnlock()

	formulas := make([]*Formula, 0)
	for _, f := range m.formulas {
		if f.Category == category {
			formulas = append(formulas, f)
		}
	}

	return formulas
}

// ListByTags 按标签列出公式
func (m *FormulaManager) ListByTags(tags []string) []*Formula {
	m.mu.RLock()
	defer m.mu.RUnlock()

	formulas := make([]*Formula, 0)
	for _, f := range m.formulas {
		if hasAnyTag(f.Tags, tags) {
			formulas = append(formulas, f)
		}
	}

	return formulas
}

// Search 搜索公式
func (m *FormulaManager) Search(query string) []*Formula {
	m.mu.RLock()
	defer m.mu.RUnlock()

	query = strings.ToLower(query)
	formulas := make([]*Formula, 0)

	for _, f := range m.formulas {
		if strings.Contains(strings.ToLower(f.Name), query) ||
			strings.Contains(strings.ToLower(f.Description), query) ||
			strings.Contains(strings.ToLower(f.Expression), query) {
			formulas = append(formulas, f)
		}
	}

	return formulas
}

// ==================== 公式验证 ====================

// Validate 验证公式
func (m *FormulaManager) Validate(id string) error {
	m.mu.RLock()
	formula, exists := m.formulas[id]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("formula not found: %s", id)
	}

	return m.validator.Validate(formula)
}

// ValidateExpression 验证表达式
func (m *FormulaManager) ValidateExpression(expression string) error {
	return m.validator.ValidateExpression(expression)
}

// ==================== 公式执行 ====================

// Execute 执行公式
func (m *FormulaManager) Execute(id string, variables map[string]interface{}) (interface{}, error) {
	m.mu.RLock()
	formula, exists := m.formulas[id]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("formula not found: %s", id)
	}

	if formula.Status != StatusActive {
		return nil, fmt.Errorf("formula %s is not active (status: %s)", id, formula.Status)
	}

	if formula.compiled == nil {
		compiled, err := m.executor.Compile(formula.Expression)
		if err != nil {
			return nil, fmt.Errorf("compile error: %w", err)
		}
		formula.compiled = compiled
	}

	return formula.compiled.Execute(m.executor, variables)
}

// ExecuteWithContext 带上下文执行公式
func (m *FormulaManager) ExecuteWithContext(ctx context.Context, id string, variables map[string]interface{}) (interface{}, error) {
	// 简化实现，实际应该支持上下文传递
	return m.Execute(id, variables)
}

// ExecuteBatch 批量执行公式
func (m *FormulaManager) ExecuteBatch(ids []string, variables map[string]interface{}) (map[string]interface{}, map[string]error) {
	results := make(map[string]interface{})
	errors := make(map[string]error)

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, id := range ids {
		wg.Add(1)
		go func(formulaID string) {
			defer wg.Done()
			result, err := m.Execute(formulaID, variables)
			mu.Lock()
			if err != nil {
				errors[formulaID] = err
			} else {
				results[formulaID] = result
			}
			mu.Unlock()
		}(id)
	}

	wg.Wait()
	return results, errors
}

// ==================== 依赖分析 ====================

// GetDependencies 获取公式依赖
func (m *FormulaManager) GetDependencies(id string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	formula, exists := m.formulas[id]
	if !exists {
		return nil, fmt.Errorf("formula not found: %s", id)
	}

	return formula.dependencies, nil
}

// GetDependents 获取依赖此公式的公式
func (m *FormulaManager) GetDependents(id string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.formulas[id]; !exists {
		return nil, fmt.Errorf("formula not found: %s", id)
	}

	dependents := make([]string, 0)
	for _, f := range m.formulas {
		for _, dep := range f.dependencies {
			if dep == id {
				dependents = append(dependents, f.ID)
				break
			}
		}
	}

	return dependents, nil
}

// GetDependencyGraph 获取依赖图
func (m *FormulaManager) GetDependencyGraph() map[string][]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	graph := make(map[string][]string)
	for id, f := range m.formulas {
		graph[id] = f.dependencies
	}

	return graph
}

// TopologicalSort 拓扑排序
func (m *FormulaManager) TopologicalSort() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 构建依赖图
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	for id := range m.formulas {
		inDegree[id] = 0
	}

	for id, f := range m.formulas {
		for _, dep := range f.dependencies {
			if _, exists := m.formulas[dep]; exists {
				graph[dep] = append(graph[dep], id)
				inDegree[id]++
			}
		}
	}

	// Kahn算法
	queue := make([]string, 0)
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	result := make([]string, 0)
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		result = append(result, id)

		for _, neighbor := range graph[id] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(result) != len(m.formulas) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	return result, nil
}

// ==================== 版本管理 ====================

// GetVersions 获取公式版本历史
func (m *FormulaManager) GetVersions(id string) ([]*FormulaVersion, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.formulas[id]; !exists {
		return nil, fmt.Errorf("formula not found: %s", id)
	}

	versions := m.versions[id]
	if versions == nil {
		return []*FormulaVersion{}, nil
	}

	return versions, nil
}

// Rollback 回滚到指定版本
func (m *FormulaManager) Rollback(id string, version string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	formula, exists := m.formulas[id]
	if !exists {
		return fmt.Errorf("formula not found: %s", id)
	}

	versions := m.versions[id]
	if versions == nil {
		return fmt.Errorf("no version history for formula: %s", id)
	}

	// 查找指定版本
	var targetVersion *FormulaVersion
	for _, v := range versions {
		if v.Version == version {
			targetVersion = v
			break
		}
	}

	if targetVersion == nil {
		return fmt.Errorf("version %s not found for formula: %s", version, id)
	}

	// 记录当前版本
	currentExpression := formula.Expression
	currentVersion := formula.Version

	// 回滚
	formula.Expression = targetVersion.Expression
	formula.Version = version
	formula.UpdatedAt = time.Now()

	// 重新编译
	compiled, err := m.executor.Compile(formula.Expression)
	if err != nil {
		// 恢复原表达式
		formula.Expression = currentExpression
		formula.Version = currentVersion
		return fmt.Errorf("compile error: %w", err)
	}
	formula.compiled = compiled
	formula.variables = compiled.GetVariables()
	formula.functions = compiled.GetFunctions()

	// 记录回滚版本
	m.recordVersion(formula, fmt.Sprintf("Rollback from version %s", currentVersion))

	return nil
}

// recordVersion 记录版本
func (m *FormulaManager) recordVersion(formula *Formula, note string) {
	version := &FormulaVersion{
		Version:    formula.Version,
		Expression: formula.Expression,
		ChangedAt:  time.Now(),
		ChangedBy:  formula.CreatedBy,
		ChangeNote: note,
	}

	if m.versions[formula.ID] == nil {
		m.versions[formula.ID] = make([]*FormulaVersion, 0)
	}

	m.versions[formula.ID] = append(m.versions[formula.ID], version)

	// 限制版本数量
	if len(m.versions[formula.ID]) > m.config.MaxVersions {
		m.versions[formula.ID] = m.versions[formula.ID][1:]
	}
}

// ==================== 状态管理 ====================

// Activate 激活公式
func (m *FormulaManager) Activate(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	formula, exists := m.formulas[id]
	if !exists {
		return fmt.Errorf("formula not found: %s", id)
	}

	formula.Status = StatusActive
	formula.UpdatedAt = time.Now()

	return nil
}

// Deactivate 停用公式
func (m *FormulaManager) Deactivate(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	formula, exists := m.formulas[id]
	if !exists {
		return fmt.Errorf("formula not found: %s", id)
	}

	formula.Status = StatusInactive
	formula.UpdatedAt = time.Now()

	return nil
}

// Deprecate 标记为废弃
func (m *FormulaManager) Deprecate(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	formula, exists := m.formulas[id]
	if !exists {
		return fmt.Errorf("formula not found: %s", id)
	}

	formula.Status = StatusDeprecated
	formula.UpdatedAt = time.Now()

	return nil
}

// ==================== 辅助函数 ====================

func hasAnyTag(formulaTags []string, searchTags []string) bool {
	for _, st := range searchTags {
		for _, ft := range formulaTags {
			if st == ft {
				return true
			}
		}
	}
	return false
}

// ==================== 公式验证器 ====================

// FormulaValidator 公式验证器
type FormulaValidator struct {
	rules []ValidationRule
}

// ValidationRule 验证规则
type ValidationRule func(*Formula) error

// NewFormulaValidator 创建公式验证器
func NewFormulaValidator() *FormulaValidator {
	validator := &FormulaValidator{
		rules: make([]ValidationRule, 0),
	}

	// 添加默认规则
	validator.AddRule(validator.validateID)
	validator.AddRule(validator.validateName)
	validator.AddRule(validator.validateExpression)
	validator.AddRule(validator.validateSyntax)

	return validator
}

// AddRule 添加验证规则
func (v *FormulaValidator) AddRule(rule ValidationRule) {
	v.rules = append(v.rules, rule)
}

// Validate 验证公式
func (v *FormulaValidator) Validate(formula *Formula) error {
	for _, rule := range v.rules {
		if err := rule(formula); err != nil {
			return err
		}
	}
	return nil
}

// ValidateExpression 验证表达式
func (v *FormulaValidator) ValidateExpression(expression string) error {
	_, err := ParseFormula(expression)
	return err
}

func (v *FormulaValidator) validateID(formula *Formula) error {
	if formula.ID == "" {
		return fmt.Errorf("formula ID cannot be empty")
	}
	return nil
}

func (v *FormulaValidator) validateName(formula *Formula) error {
	if formula.Name == "" {
		return fmt.Errorf("formula name cannot be empty")
	}
	return nil
}

func (v *FormulaValidator) validateExpression(formula *Formula) error {
	if formula.Expression == "" {
		return fmt.Errorf("formula expression cannot be empty")
	}
	return nil
}

func (v *FormulaValidator) validateSyntax(formula *Formula) error {
	_, err := ParseFormula(formula.Expression)
	if err != nil {
		return fmt.Errorf("syntax error: %w", err)
	}
	return nil
}

// ==================== 依赖分析器 ====================

// DependencyAnalyzer 依赖分析器
type DependencyAnalyzer struct {
	variablePattern string
}

// NewDependencyAnalyzer 创建依赖分析器
func NewDependencyAnalyzer() *DependencyAnalyzer {
	return &DependencyAnalyzer{
		variablePattern: `\$\{([^}]+)\}`,
	}
}

// Analyze 分析公式依赖
func (a *DependencyAnalyzer) Analyze(formula *Formula) ([]string, error) {
	// 从变量中提取依赖
	dependencies := make([]string, 0)

	// 变量可能是其他公式的ID
	for _, v := range formula.variables {
		// 检查变量是否是公式引用（以 formula. 开头）
		if strings.HasPrefix(v, "formula.") {
			dep := strings.TrimPrefix(v, "formula.")
			dependencies = append(dependencies, dep)
		}
	}

	return dependencies, nil
}

// AnalyzeExpression 分析表达式依赖
func (a *DependencyAnalyzer) AnalyzeExpression(expression string) ([]string, error) {
	// 解析表达式
	node, err := ParseFormula(expression)
	if err != nil {
		return nil, err
	}

	// 提取变量
	variables := extractVariables(node)

	// 提取依赖
	dependencies := make([]string, 0)
	for _, v := range variables {
		if strings.HasPrefix(v, "formula.") {
			dep := strings.TrimPrefix(v, "formula.")
			dependencies = append(dependencies, dep)
		}
	}

	return dependencies, nil
}

// CheckCircularDependency 检查循环依赖
func (a *DependencyAnalyzer) CheckCircularDependency(formulas map[string]*Formula) (bool, []string) {
	// 构建依赖图
	graph := make(map[string][]string)
	for id, f := range formulas {
		graph[id] = f.dependencies
	}

	// DFS检测环
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	cycle := make([]string, 0)

	var dfs func(string) bool
	dfs = func(node string) bool {
		visited[node] = true
		recStack[node] = true

		for _, neighbor := range graph[node] {
			if !visited[neighbor] {
				if dfs(neighbor) {
					cycle = append(cycle, neighbor)
					return true
				}
			} else if recStack[neighbor] {
				cycle = append(cycle, neighbor)
				return true
			}
		}

		recStack[node] = false
		return false
	}

	for node := range graph {
		if !visited[node] {
			if dfs(node) {
				return true, cycle
			}
		}
	}

	return false, nil
}

// ==================== 统计信息 ====================

// ManagerStats 管理器统计信息
type ManagerStats struct {
	TotalFormulas   int
	ActiveFormulas  int
	InactiveFormulas int
	DraftFormulas   int
	DeprecatedFormulas int
	Categories      map[string]int
	Tags            map[string]int
}

// GetStats 获取统计信息
func (m *FormulaManager) GetStats() *ManagerStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &ManagerStats{
		Categories: make(map[string]int),
		Tags:       make(map[string]int),
	}

	for _, f := range m.formulas {
		stats.TotalFormulas++

		switch f.Status {
		case StatusActive:
			stats.ActiveFormulas++
		case StatusInactive:
			stats.InactiveFormulas++
		case StatusDraft:
			stats.DraftFormulas++
		case StatusDeprecated:
			stats.DeprecatedFormulas++
		}

		if f.Category != "" {
			stats.Categories[f.Category]++
		}

		for _, tag := range f.Tags {
			stats.Tags[tag]++
		}
	}

	return stats
}

// ==================== 导入导出 ====================

// Export 导出公式
func (m *FormulaManager) Export(ids []string) ([]*Formula, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	formulas := make([]*Formula, 0, len(ids))
	for _, id := range ids {
		f, exists := m.formulas[id]
		if !exists {
			return nil, fmt.Errorf("formula not found: %s", id)
		}
		formulas = append(formulas, f)
	}

	return formulas, nil
}

// Import 导入公式
func (m *FormulaManager) Import(formulas []*Formula, overwrite bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, f := range formulas {
		existing, exists := m.formulas[f.ID]

		if exists && !overwrite {
			continue
		}

		if exists {
			// 保留创建信息
			f.CreatedAt = existing.CreatedAt
			f.CreatedBy = existing.CreatedBy
		} else {
			if f.CreatedAt.IsZero() {
				f.CreatedAt = time.Now()
			}
		}
		f.UpdatedAt = time.Now()

		// 编译
		compiled, err := m.executor.Compile(f.Expression)
		if err != nil {
			return fmt.Errorf("compile error for formula %s: %w", f.ID, err)
		}
		f.compiled = compiled
		f.variables = compiled.GetVariables()
		f.functions = compiled.GetFunctions()

		m.formulas[f.ID] = f
	}

	return nil
}

// ==================== 批量操作 ====================

// BatchCreate 批量创建
func (m *FormulaManager) BatchCreate(formulas []*Formula) []error {
	errors := make([]error, len(formulas))

	for i, f := range formulas {
		errors[i] = m.Create(f)
	}

	return errors
}

// BatchUpdate 批量更新
func (m *FormulaManager) BatchUpdate(formulas []*Formula) []error {
	errors := make([]error, len(formulas))

	for i, f := range formulas {
		errors[i] = m.Update(f)
	}

	return errors
}

// BatchDelete 批量删除
func (m *FormulaManager) BatchDelete(ids []string) []error {
	errors := make([]error, len(ids))

	for i, id := range ids {
		errors[i] = m.Delete(id)
	}

	return errors
}

// BatchActivate 批量激活
func (m *FormulaManager) BatchActivate(ids []string) []error {
	errors := make([]error, len(ids))

	for i, id := range ids {
		errors[i] = m.Activate(id)
	}

	return errors
}

// BatchDeactivate 批量停用
func (m *FormulaManager) BatchDeactivate(ids []string) []error {
	errors := make([]error, len(ids))

	for i, id := range ids {
		errors[i] = m.Deactivate(id)
	}

	return errors
}

// ==================== 排序 ====================

// SortByName 按名称排序
func (m *FormulaManager) SortByName(formulas []*Formula) []*Formula {
	sorted := make([]*Formula, len(formulas))
	copy(sorted, formulas)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	return sorted
}

// SortByCreatedAt 按创建时间排序
func (m *FormulaManager) SortByCreatedAt(formulas []*Formula) []*Formula {
	sorted := make([]*Formula, len(formulas))
	copy(sorted, formulas)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
	})

	return sorted
}

// SortByUpdatedAt 按更新时间排序
func (m *FormulaManager) SortByUpdatedAt(formulas []*Formula) []*Formula {
	sorted := make([]*Formula, len(formulas))
	copy(sorted, formulas)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].UpdatedAt.Before(sorted[j].UpdatedAt)
	})

	return sorted
}
