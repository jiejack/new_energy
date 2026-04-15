package rule

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrPointNotFound      = errors.New("compute point not found")
	ErrPointExists        = errors.New("compute point already exists")
	ErrInvalidPointConfig = errors.New("invalid compute point configuration")
	ErrCircularDependency = errors.New("circular dependency detected")
)

// PointType 计算点类型
type PointType string

const (
	PointTypeVirtual   PointType = "virtual"   // 虚拟计算点
	PointTypeDerived   PointType = "derived"   // 派生计算点
	PointTypeAggregate PointType = "aggregate" // 聚合计算点
	PointTypeStatistic PointType = "statistic" // 统计计算点
)

// PointStatus 计算点状态
type PointStatus string

const (
	PointStatusActive   PointStatus = "active"   // 活跃
	PointStatusInactive PointStatus = "inactive" // 不活跃
	PointStatusError    PointStatus = "error"    // 错误
	PointStatusDisabled PointStatus = "disabled" // 禁用
)

// ComputePoint 计算点结构
type ComputePoint struct {
	ID          string                 `json:"id"`          // 计算点ID
	Name        string                 `json:"name"`        // 计算点名称
	Description string                 `json:"description"` // 描述
	Type        PointType              `json:"type"`        // 类型
	Status      PointStatus            `json:"status"`      // 状态
	Formula     string                 `json:"formula"`     // 计算公式
	Dependencies []string              `json:"dependencies"` // 依赖的计算点或测点ID
	Unit        string                 `json:"unit"`        // 单位
	Precision   int                    `json:"precision"`   // 精度
	Range       *ValueRange            `json:"range"`       // 数值范围
	Config      map[string]interface{} `json:"config"`      // 配置参数
	Tags        map[string]string      `json:"tags"`        // 标签
	CreateTime  time.Time              `json:"createTime"`  // 创建时间
	UpdateTime  time.Time              `json:"updateTime"`  // 更新时间
	LastCompute time.Time              `json:"lastCompute"` // 最后计算时间
	Value       float64                `json:"value"`       // 当前值
	Quality     int                    `json:"quality"`     // 数据质量
	Error       string                 `json:"error"`       // 错误信息
}

// ValueRange 数值范围
type ValueRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// PointConfig 计算点配置
type PointConfig struct {
	ComputeInterval time.Duration         `json:"computeInterval"` // 计算间隔
	Timeout         time.Duration         `json:"timeout"`         // 超时时间
	RetryCount      int                   `json:"retryCount"`      // 重试次数
	EnableCache     bool                  `json:"enableCache"`     // 是否启用缓存
	CacheTTL        time.Duration         `json:"cacheTTL"`        // 缓存时间
	EnableAlert     bool                  `json:"enableAlert"`     // 是否启用告警
	AlertRules      []string              `json:"alertRules"`      // 告警规则ID列表
	Extra           map[string]interface{} `json:"extra"`           // 扩展配置
}

// PointDependency 计算点依赖关系
type PointDependency struct {
	PointID      string   `json:"pointId"`      // 计算点ID
	DependsOn    []string `json:"dependsOn"`    // 依赖的点ID列表
	Level        int      `json:"level"`        // 依赖层级
	ComputeOrder int      `json:"computeOrder"` // 计算顺序
}

// PointManager 计算点管理器
type PointManager struct {
	points      map[string]*ComputePoint
	configs     map[string]*PointConfig
	dependencies map[string]*PointDependency
	mu          sync.RWMutex
	depGraph    *DependencyGraph
}

// NewPointManager 创建计算点管理器
func NewPointManager() *PointManager {
	return &PointManager{
		points:      make(map[string]*ComputePoint),
		configs:     make(map[string]*PointConfig),
		dependencies: make(map[string]*PointDependency),
		depGraph:    NewDependencyGraph(),
	}
}

// CreatePoint 创建计算点
func (pm *PointManager) CreatePoint(ctx context.Context, point *ComputePoint) error {
	if point.ID == "" {
		return ErrInvalidPointConfig
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.points[point.ID]; exists {
		return ErrPointExists
	}

	// 设置默认值
	if point.Status == "" {
		point.Status = PointStatusActive
	}
	if point.CreateTime.IsZero() {
		point.CreateTime = time.Now()
	}
	point.UpdateTime = time.Now()

	// 验证依赖关系
	if err := pm.validateDependencies(point); err != nil {
		return err
	}

	// 添加到依赖图
	if err := pm.depGraph.AddNode(point.ID, point.Dependencies); err != nil {
		return err
	}

	pm.points[point.ID] = point

	// 构建依赖关系
	pm.buildDependency(point.ID, point.Dependencies)

	return nil
}

// UpdatePoint 更新计算点
func (pm *PointManager) UpdatePoint(ctx context.Context, point *ComputePoint) error {
	if point.ID == "" {
		return ErrInvalidPointConfig
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	existing, exists := pm.points[point.ID]
	if !exists {
		return ErrPointNotFound
	}

	// 验证依赖关系
	if err := pm.validateDependencies(point); err != nil {
		return err
	}

	// 更新依赖图
	if err := pm.depGraph.UpdateNode(point.ID, point.Dependencies); err != nil {
		return err
	}

	// 保留创建时间
	point.CreateTime = existing.CreateTime
	point.UpdateTime = time.Now()

	pm.points[point.ID] = point

	// 重新构建依赖关系
	pm.buildDependency(point.ID, point.Dependencies)

	return nil
}

// DeletePoint 删除计算点
func (pm *PointManager) DeletePoint(ctx context.Context, pointID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.points[pointID]; !exists {
		return ErrPointNotFound
	}

	// 检查是否有其他点依赖此点
	dependents := pm.getDependents(pointID)
	if len(dependents) > 0 {
		return fmt.Errorf("cannot delete point %s: other points depend on it: %v", pointID, dependents)
	}

	// 从依赖图中移除
	pm.depGraph.RemoveNode(pointID)

	delete(pm.points, pointID)
	delete(pm.configs, pointID)
	delete(pm.dependencies, pointID)

	return nil
}

// GetPoint 获取计算点
func (pm *PointManager) GetPoint(ctx context.Context, pointID string) (*ComputePoint, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	point, exists := pm.points[pointID]
	if !exists {
		return nil, ErrPointNotFound
	}

	return point, nil
}

// GetPointsByType 按类型获取计算点
func (pm *PointManager) GetPointsByType(ctx context.Context, pointType PointType) []*ComputePoint {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	points := make([]*ComputePoint, 0)
	for _, point := range pm.points {
		if point.Type == pointType {
			points = append(points, point)
		}
	}

	return points
}

// GetPointsByStatus 按状态获取计算点
func (pm *PointManager) GetPointsByStatus(ctx context.Context, status PointStatus) []*ComputePoint {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	points := make([]*ComputePoint, 0)
	for _, point := range pm.points {
		if point.Status == status {
			points = append(points, point)
		}
	}

	return points
}

// GetAllPoints 获取所有计算点
func (pm *PointManager) GetAllPoints(ctx context.Context) []*ComputePoint {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	points := make([]*ComputePoint, 0, len(pm.points))
	for _, point := range pm.points {
		points = append(points, point)
	}

	return points
}

// SetPointConfig 设置计算点配置
func (pm *PointManager) SetPointConfig(ctx context.Context, pointID string, config *PointConfig) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.points[pointID]; !exists {
		return ErrPointNotFound
	}

	pm.configs[pointID] = config
	return nil
}

// GetPointConfig 获取计算点配置
func (pm *PointManager) GetPointConfig(ctx context.Context, pointID string) (*PointConfig, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	config, exists := pm.configs[pointID]
	if !exists {
		return nil, ErrPointNotFound
	}

	return config, nil
}

// UpdatePointStatus 更新计算点状态
func (pm *PointManager) UpdatePointStatus(ctx context.Context, pointID string, status PointStatus) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	point, exists := pm.points[pointID]
	if !exists {
		return ErrPointNotFound
	}

	point.Status = status
	point.UpdateTime = time.Now()

	return nil
}

// UpdatePointValue 更新计算点值
func (pm *PointManager) UpdatePointValue(ctx context.Context, pointID string, value float64, quality int) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	point, exists := pm.points[pointID]
	if !exists {
		return ErrPointNotFound
	}

	point.Value = value
	point.Quality = quality
	point.LastCompute = time.Now()
	point.UpdateTime = time.Now()

	return nil
}

// GetDependencies 获取计算点依赖关系
func (pm *PointManager) GetDependencies(ctx context.Context, pointID string) (*PointDependency, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	dep, exists := pm.dependencies[pointID]
	if !exists {
		return nil, ErrPointNotFound
	}

	return dep, nil
}

// GetComputeOrder 获取计算顺序
func (pm *PointManager) GetComputeOrder(ctx context.Context) [][]string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return pm.depGraph.GetTopologicalOrder()
}

// GetDependents 获取依赖此点的所有点
func (pm *PointManager) GetDependents(ctx context.Context, pointID string) []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return pm.getDependents(pointID)
}

// validateDependencies 验证依赖关系
func (pm *PointManager) validateDependencies(point *ComputePoint) error {
	for _, depID := range point.Dependencies {
		// 检查依赖点是否存在（可以是计算点或测点）
		if _, exists := pm.points[depID]; !exists {
			// 如果不是计算点，可能是测点，这里不做限制
			continue
		}
	}

	// 检查循环依赖
	if pm.depGraph.HasCycle(point.ID, point.Dependencies) {
		return ErrCircularDependency
	}

	return nil
}

// buildDependency 构建依赖关系
func (pm *PointManager) buildDependency(pointID string, dependencies []string) {
	dep := &PointDependency{
		PointID:   pointID,
		DependsOn: dependencies,
	}

	// 计算依赖层级
	level := 0
	for _, depID := range dependencies {
		if parentDep, exists := pm.dependencies[depID]; exists {
			if parentDep.Level+1 > level {
				level = parentDep.Level + 1
			}
		}
	}
	dep.Level = level

	pm.dependencies[pointID] = dep
}

// getDependents 获取依赖此点的所有点
func (pm *PointManager) getDependents(pointID string) []string {
	dependents := make([]string, 0)
	for id, dep := range pm.dependencies {
		for _, depID := range dep.DependsOn {
			if depID == pointID {
				dependents = append(dependents, id)
				break
			}
		}
	}
	return dependents
}

// DependencyGraph 依赖图
type DependencyGraph struct {
	nodes map[string][]string // 节点及其依赖
	mu    sync.RWMutex
}

// NewDependencyGraph 创建依赖图
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		nodes: make(map[string][]string),
	}
}

// AddNode 添加节点
func (dg *DependencyGraph) AddNode(nodeID string, dependencies []string) error {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	if _, exists := dg.nodes[nodeID]; exists {
		return fmt.Errorf("node already exists: %s", nodeID)
	}

	dg.nodes[nodeID] = dependencies
	return nil
}

// UpdateNode 更新节点
func (dg *DependencyGraph) UpdateNode(nodeID string, dependencies []string) error {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	if _, exists := dg.nodes[nodeID]; !exists {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	dg.nodes[nodeID] = dependencies
	return nil
}

// RemoveNode 移除节点
func (dg *DependencyGraph) RemoveNode(nodeID string) {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	delete(dg.nodes, nodeID)
}

// HasCycle 检查是否有循环依赖
func (dg *DependencyGraph) HasCycle(nodeID string, dependencies []string) bool {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	// 创建临时图副本，避免修改原数据
	tempNodes := make(map[string][]string)
	for k, v := range dg.nodes {
		tempNodes[k] = v
	}
	tempNodes[nodeID] = dependencies

	return dg.hasCycleDFSWithGraph(nodeID, visited, recStack, tempNodes)
}

// hasCycleDFSWithGraph 使用给定的图进行深度优先搜索检测循环
func (dg *DependencyGraph) hasCycleDFSWithGraph(nodeID string, visited, recStack map[string]bool, nodes map[string][]string) bool {
	visited[nodeID] = true
	recStack[nodeID] = true

	for _, depID := range nodes[nodeID] {
		if !visited[depID] {
			if dg.hasCycleDFSWithGraph(depID, visited, recStack, nodes) {
				return true
			}
		} else if recStack[depID] {
			return true
		}
	}

	recStack[nodeID] = false
	return false
}



// GetTopologicalOrder 获取拓扑排序
func (dg *DependencyGraph) GetTopologicalOrder() [][]string {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	// 计算入度
	inDegree := make(map[string]int)
	for nodeID := range dg.nodes {
		if _, exists := inDegree[nodeID]; !exists {
			inDegree[nodeID] = 0
		}
		for _, depID := range dg.nodes[nodeID] {
			if _, exists := dg.nodes[depID]; exists {
				inDegree[nodeID]++
			}
		}
	}

	// 按层级分组
	result := make([][]string, 0)
	processed := make(map[string]bool)

	for len(processed) < len(dg.nodes) {
		level := make([]string, 0)

		for nodeID := range dg.nodes {
			if processed[nodeID] {
				continue
			}

			// 检查所有依赖是否已处理
			allDepsProcessed := true
			for _, depID := range dg.nodes[nodeID] {
				if _, exists := dg.nodes[depID]; exists && !processed[depID] {
					allDepsProcessed = false
					break
				}
			}

			if allDepsProcessed {
				level = append(level, nodeID)
			}
		}

		if len(level) == 0 {
			break // 避免死循环
		}

		for _, nodeID := range level {
			processed[nodeID] = true
		}

		result = append(result, level)
	}

	return result
}

// PointFilter 计算点过滤器
type PointFilter struct {
	Types   []PointType
	Status  []PointStatus
	Tags    map[string]string
	IDs     []string
}

// Filter 过滤计算点
func (pm *PointManager) Filter(ctx context.Context, filter *PointFilter) []*ComputePoint {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make([]*ComputePoint, 0)

	for _, point := range pm.points {
		if !pm.matchFilter(point, filter) {
			continue
		}
		result = append(result, point)
	}

	return result
}

// matchFilter 匹配过滤器
func (pm *PointManager) matchFilter(point *ComputePoint, filter *PointFilter) bool {
	// 类型过滤
	if len(filter.Types) > 0 {
		matched := false
		for _, t := range filter.Types {
			if point.Type == t {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// 状态过滤
	if len(filter.Status) > 0 {
		matched := false
		for _, s := range filter.Status {
			if point.Status == s {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// ID过滤
	if len(filter.IDs) > 0 {
		matched := false
		for _, id := range filter.IDs {
			if point.ID == id {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// 标签过滤
	for key, value := range filter.Tags {
		if point.Tags[key] != value {
			return false
		}
	}

	return true
}

// PointStats 计算点统计
type PointStats struct {
	TotalPoints     int            `json:"totalPoints"`     // 总计算点数
	ActivePoints    int            `json:"activePoints"`    // 活跃计算点数
	InactivePoints  int            `json:"inactivePoints"`  // 不活跃计算点数
	ErrorPoints     int            `json:"errorPoints"`     // 错误计算点数
	DisabledPoints  int            `json:"disabledPoints"`  // 禁用计算点数
	ByType          map[PointType]int `json:"byType"`       // 按类型统计
	AverageComputeTime time.Duration `json:"averageComputeTime"` // 平均计算时间
}

// GetStats 获取统计信息
func (pm *PointManager) GetStats(ctx context.Context) *PointStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	stats := &PointStats{
		ByType: make(map[PointType]int),
	}

	for _, point := range pm.points {
		stats.TotalPoints++

		switch point.Status {
		case PointStatusActive:
			stats.ActivePoints++
		case PointStatusInactive:
			stats.InactivePoints++
		case PointStatusError:
			stats.ErrorPoints++
		case PointStatusDisabled:
			stats.DisabledPoints++
		}

		stats.ByType[point.Type]++
	}

	return stats
}
