// Package partition 提供时序数据分区策略实现
// 支持时间分区、范围分区、列表分区等多种分区方式
package partition

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

// 定义错误类型
var (
	ErrInvalidPartitionCount = errors.New("invalid partition count")
	ErrInvalidPartitionKey   = errors.New("invalid partition key")
	ErrPartitionNotFound     = errors.New("partition not found")
	ErrPartitionExists       = errors.New("partition already exists")
	ErrInvalidTimeRange      = errors.New("invalid time range")
	ErrAutoCreateDisabled    = errors.New("auto create disabled")
)

// Partition 表示一个分区
type Partition struct {
	ID          int                    // 分区ID
	Name        string                 // 分区名称
	StartKey    []byte                 // 起始键（范围分区）
	EndKey      []byte                 // 结束键（范围分区）
	StartTime   time.Time              // 起始时间（时间分区）
	EndTime     time.Time              // 结束时间（时间分区）
	Values      []string               // 值列表（列表分区）
	Status      PartitionStatus        // 状态
	Size        int64                  // 数据大小
	RowCount    int64                  // 行数
	Metadata    map[string]interface{} // 元数据
	CreatedAt   time.Time              // 创建时间
	LastAccess  time.Time              // 最后访问时间
}

// PartitionStatus 分区状态
type PartitionStatus int

const (
	PartitionStatusActive PartitionStatus = iota
	PartitionStatusInactive
	PartitionStatusReadOnly
	PartitionStatusArchived
	PartitionStatusDropping
)

// String 返回分区状态的字符串表示
func (s PartitionStatus) String() string {
	switch s {
	case PartitionStatusActive:
		return "active"
	case PartitionStatusInactive:
		return "inactive"
	case PartitionStatusReadOnly:
		return "read_only"
	case PartitionStatusArchived:
		return "archived"
	case PartitionStatusDropping:
		return "dropping"
	default:
		return "unknown"
	}
}

// PartitionKey 分区键
type PartitionKey struct {
	DeviceID  string    // 设备ID
	PointID   string    // 测点ID
	Timestamp time.Time // 时间戳
	Value     string    // 值（列表分区）
	Tags      map[string]string
}

// PartitionStrategy 分区策略接口
type PartitionStrategy interface {
	// GetPartition 根据分区键获取分区
	GetPartition(key *PartitionKey) (*Partition, error)

	// GetPartitions 获取所有分区
	GetPartitions() []*Partition

	// CreatePartition 创建分区
	CreatePartition(partition *Partition) error

	// DropPartition 删除分区
	DropPartition(partitionID int) error

	// GetPartitionCount 获取分区数量
	GetPartitionCount() int

	// GetType 获取分区类型
	GetType() string

	// GetActivePartitions 获取活跃分区
	GetActivePartitions() []*Partition
}

// TimePartition 时间分区策略
type TimePartition struct {
	mu            sync.RWMutex
	partitions    []*Partition
	granularity   TimeGranularity
	autoCreate    bool
	retentionDays int // 数据保留天数
	maxPartitions int
	timeRanges    []TimeRange
}

// TimeGranularity 时间粒度
type TimeGranularity int

const (
	TimeGranularityHour TimeGranularity = iota
	TimeGranularityDay
	TimeGranularityMonth
	TimeGranularityYear
)

// String 返回时间粒度的字符串表示
func (tg TimeGranularity) String() string {
	switch tg {
	case TimeGranularityHour:
		return "hour"
	case TimeGranularityDay:
		return "day"
	case TimeGranularityMonth:
		return "month"
	case TimeGranularityYear:
		return "year"
	default:
		return "unknown"
	}
}

// TimeRange 时间范围
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// NewTimePartition 创建时间分区策略
func NewTimePartition(granularity TimeGranularity, opts ...TimePartitionOption) (*TimePartition, error) {
	tp := &TimePartition{
		granularity:   granularity,
		autoCreate:    true,
		retentionDays: 365,
		maxPartitions: 400,
		partitions:    make([]*Partition, 0),
		timeRanges:    make([]TimeRange, 0),
	}

	for _, opt := range opts {
		opt(tp)
	}

	return tp, nil
}

// TimePartitionOption 时间分区选项
type TimePartitionOption func(*TimePartition)

// WithAutoCreatePartition 设置自动创建分区
func WithAutoCreatePartition(auto bool) TimePartitionOption {
	return func(tp *TimePartition) {
		tp.autoCreate = auto
	}
}

// WithRetentionDays 设置数据保留天数
func WithRetentionDays(days int) TimePartitionOption {
	return func(tp *TimePartition) {
		tp.retentionDays = days
	}
}

// WithMaxPartitions 设置最大分区数
func WithMaxPartitions(max int) TimePartitionOption {
	return func(tp *TimePartition) {
		tp.maxPartitions = max
	}
}

// GetPartition 根据分区键获取分区
func (tp *TimePartition) GetPartition(key *PartitionKey) (*Partition, error) {
	if key == nil {
		return nil, ErrInvalidPartitionKey
	}

	tp.mu.Lock()
	defer tp.mu.Unlock()

	// 计算分区时间
	partitionTime := tp.truncateTime(key.Timestamp)

	// 查找现有分区
	for i, partition := range tp.partitions {
		if partition.Status == PartitionStatusDropping {
			continue
		}

		r := tp.timeRanges[i]
		if !partitionTime.Before(r.Start) && partitionTime.Before(r.End) {
			// 更新最后访问时间
			partition.LastAccess = time.Now()
			return partition, nil
		}
	}

	// 自动创建分区
	if tp.autoCreate {
		return tp.createPartitionForTime(partitionTime)
	}

	return nil, ErrPartitionNotFound
}

// truncateTime 截断时间到指定粒度
func (tp *TimePartition) truncateTime(t time.Time) time.Time {
	switch tp.granularity {
	case TimeGranularityHour:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
	case TimeGranularityDay:
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	case TimeGranularityMonth:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	case TimeGranularityYear:
		return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
	default:
		return t
	}
}

// createPartitionForTime 为指定时间创建分区
func (tp *TimePartition) createPartitionForTime(t time.Time) (*Partition, error) {
	if len(tp.partitions) >= tp.maxPartitions {
		return nil, fmt.Errorf("max partitions limit reached: %d", tp.maxPartitions)
	}

	start := tp.truncateTime(t)
	end := tp.getNextTime(start)

	partitionID := len(tp.partitions)
	partition := &Partition{
		ID:         partitionID,
		Name:       tp.generatePartitionName(start),
		StartTime:  start,
		EndTime:    end,
		Status:     PartitionStatusActive,
		CreatedAt:  time.Now(),
		LastAccess: time.Now(),
		Metadata: map[string]interface{}{
			"type":      "time",
			"startTime": start,
			"endTime":   end,
		},
	}

	tp.partitions = append(tp.partitions, partition)
	tp.timeRanges = append(tp.timeRanges, TimeRange{Start: start, End: end})

	return partition, nil
}

// generatePartitionName 生成分区名称
func (tp *TimePartition) generatePartitionName(t time.Time) string {
	switch tp.granularity {
	case TimeGranularityHour:
		return fmt.Sprintf("p%s", t.Format("20060102_15"))
	case TimeGranularityDay:
		return fmt.Sprintf("p%s", t.Format("20060102"))
	case TimeGranularityMonth:
		return fmt.Sprintf("p%s", t.Format("200601"))
	case TimeGranularityYear:
		return fmt.Sprintf("p%s", t.Format("2006"))
	default:
		return fmt.Sprintf("p%s", t.Format("20060102"))
	}
}

// getNextTime 获取下一个时间点
func (tp *TimePartition) getNextTime(t time.Time) time.Time {
	switch tp.granularity {
	case TimeGranularityHour:
		return t.Add(time.Hour)
	case TimeGranularityDay:
		return t.AddDate(0, 0, 1)
	case TimeGranularityMonth:
		return t.AddDate(0, 1, 0)
	case TimeGranularityYear:
		return t.AddDate(1, 0, 0)
	default:
		return t
	}
}

// GetPartitions 获取所有分区
func (tp *TimePartition) GetPartitions() []*Partition {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	result := make([]*Partition, len(tp.partitions))
	copy(result, tp.partitions)
	return result
}

// CreatePartition 创建分区
func (tp *TimePartition) CreatePartition(partition *Partition) error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	// 检查时间范围是否有效
	if partition.StartTime.IsZero() || partition.EndTime.IsZero() {
		return ErrInvalidTimeRange
	}

	// 检查时间范围是否重叠
	for i, r := range tp.timeRanges {
		if partition.StartTime.Before(r.End) && partition.EndTime.After(r.Start) {
			return fmt.Errorf("time range overlaps with partition %s", tp.partitions[i].Name)
		}
	}

	partition.ID = len(tp.partitions)
	partition.CreatedAt = time.Now()
	partition.LastAccess = time.Now()

	tp.partitions = append(tp.partitions, partition)
	tp.timeRanges = append(tp.timeRanges, TimeRange{
		Start: partition.StartTime,
		End:   partition.EndTime,
	})

	return nil
}

// DropPartition 删除分区
func (tp *TimePartition) DropPartition(partitionID int) error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	for i, p := range tp.partitions {
		if p.ID == partitionID {
			// 标记为删除中
			p.Status = PartitionStatusDropping
			// 从列表中移除
			tp.partitions = append(tp.partitions[:i], tp.partitions[i+1:]...)
			tp.timeRanges = append(tp.timeRanges[:i], tp.timeRanges[i+1:]...)
			return nil
		}
	}

	return ErrPartitionNotFound
}

// GetPartitionCount 获取分区数量
func (tp *TimePartition) GetPartitionCount() int {
	tp.mu.RLock()
	defer tp.mu.RUnlock()
	return len(tp.partitions)
}

// GetType 获取分区类型
func (tp *TimePartition) GetType() string {
	return "time"
}

// GetActivePartitions 获取活跃分区
func (tp *TimePartition) GetActivePartitions() []*Partition {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	result := make([]*Partition, 0)
	for _, p := range tp.partitions {
		if p.Status == PartitionStatusActive {
			result = append(result, p)
		}
	}
	return result
}

// GetPartitionsByTimeRange 根据时间范围获取分区
func (tp *TimePartition) GetPartitionsByTimeRange(start, end time.Time) []*Partition {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	result := make([]*Partition, 0)
	for i, partition := range tp.partitions {
		r := tp.timeRanges[i]
		if partition.Status != PartitionStatusDropping &&
			!r.End.Before(start) && !r.Start.After(end) {
			result = append(result, partition)
		}
	}
	return result
}

// PurgeOldPartitions 清理过期分区
func (tp *TimePartition) PurgeOldPartitions() ([]*Partition, error) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if tp.retentionDays <= 0 {
		return nil, nil
	}

	cutoff := time.Now().AddDate(0, 0, -tp.retentionDays)
	purged := make([]*Partition, 0)

	newPartitions := make([]*Partition, 0)
	newTimeRanges := make([]TimeRange, 0)

	for i, partition := range tp.partitions {
		if partition.EndTime.Before(cutoff) {
			purged = append(purged, partition)
		} else {
			newPartitions = append(newPartitions, partition)
			newTimeRanges = append(newTimeRanges, tp.timeRanges[i])
		}
	}

	tp.partitions = newPartitions
	tp.timeRanges = newTimeRanges

	return purged, nil
}

// RangePartition 范围分区策略
type RangePartition struct {
	mu         sync.RWMutex
	partitions []*Partition
	ranges     []KeyRange
}

// KeyRange 键范围
type KeyRange struct {
	Start []byte
	End   []byte
}

// NewRangePartition 创建范围分区策略
func NewRangePartition() *RangePartition {
	return &RangePartition{
		partitions: make([]*Partition, 0),
		ranges:     make([]KeyRange, 0),
	}
}

// GetPartition 根据分区键获取分区
func (rp *RangePartition) GetPartition(key *PartitionKey) (*Partition, error) {
	if key == nil {
		return nil, ErrInvalidPartitionKey
	}

	rp.mu.RLock()
	defer rp.mu.RUnlock()

	// 构建范围键
	rangeKey := rp.buildRangeKey(key)

	// 查找匹配的范围
	for i, partition := range rp.partitions {
		if partition.Status == PartitionStatusDropping {
			continue
		}

		r := rp.ranges[i]
		if bytesCompare(rangeKey, r.Start) >= 0 && bytesCompare(rangeKey, r.End) < 0 {
			partition.LastAccess = time.Now()
			return partition, nil
		}
	}

	return nil, ErrPartitionNotFound
}

// buildRangeKey 构建范围键
func (rp *RangePartition) buildRangeKey(key *PartitionKey) []byte {
	// 使用设备ID作为范围键
	return []byte(key.DeviceID)
}

// bytesCompare 字节比较
func bytesCompare(a, b []byte) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	for i := 0; i < minLen; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}

// CreateRangePartition 创建范围分区
func (rp *RangePartition) CreateRangePartition(partition *Partition, start, end []byte) error {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	// 检查范围是否重叠
	for i, r := range rp.ranges {
		if bytesCompare(end, r.Start) > 0 && bytesCompare(start, r.End) < 0 {
			return fmt.Errorf("range overlaps with partition %s", rp.partitions[i].Name)
		}
	}

	partition.ID = len(rp.partitions)
	partition.StartKey = start
	partition.EndKey = end
	partition.CreatedAt = time.Now()
	partition.LastAccess = time.Now()

	rp.partitions = append(rp.partitions, partition)
	rp.ranges = append(rp.ranges, KeyRange{Start: start, End: end})

	return nil
}

// GetPartitions 获取所有分区
func (rp *RangePartition) GetPartitions() []*Partition {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	result := make([]*Partition, len(rp.partitions))
	copy(result, rp.partitions)
	return result
}

// CreatePartition 创建分区
func (rp *RangePartition) CreatePartition(partition *Partition) error {
	return fmt.Errorf("use CreateRangePartition for range partition")
}

// DropPartition 删除分区
func (rp *RangePartition) DropPartition(partitionID int) error {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	for i, p := range rp.partitions {
		if p.ID == partitionID {
			rp.partitions = append(rp.partitions[:i], rp.partitions[i+1:]...)
			rp.ranges = append(rp.ranges[:i], rp.ranges[i+1:]...)
			return nil
		}
	}

	return ErrPartitionNotFound
}

// GetPartitionCount 获取分区数量
func (rp *RangePartition) GetPartitionCount() int {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	return len(rp.partitions)
}

// GetType 获取分区类型
func (rp *RangePartition) GetType() string {
	return "range"
}

// GetActivePartitions 获取活跃分区
func (rp *RangePartition) GetActivePartitions() []*Partition {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	result := make([]*Partition, 0)
	for _, p := range rp.partitions {
		if p.Status == PartitionStatusActive {
			result = append(result, p)
		}
	}
	return result
}

// ListPartition 列表分区策略
type ListPartition struct {
	mu         sync.RWMutex
	partitions []*Partition
	valueMaps  []map[string]bool // 值到分区的映射
}

// NewListPartition 创建列表分区策略
func NewListPartition() *ListPartition {
	return &ListPartition{
		partitions: make([]*Partition, 0),
		valueMaps:  make([]map[string]bool, 0),
	}
}

// GetPartition 根据分区键获取分区
func (lp *ListPartition) GetPartition(key *PartitionKey) (*Partition, error) {
	if key == nil {
		return nil, ErrInvalidPartitionKey
	}

	lp.mu.RLock()
	defer lp.mu.RUnlock()

	// 查找值所属的分区
	for i, partition := range lp.partitions {
		if partition.Status == PartitionStatusDropping {
			continue
		}

		valueMap := lp.valueMaps[i]
		if valueMap[key.Value] {
			partition.LastAccess = time.Now()
			return partition, nil
		}
	}

	return nil, ErrPartitionNotFound
}

// CreateListPartition 创建列表分区
func (lp *ListPartition) CreateListPartition(partition *Partition, values []string) error {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	// 检查值是否已被其他分区使用
	for i, valueMap := range lp.valueMaps {
		for _, v := range values {
			if valueMap[v] {
				return fmt.Errorf("value %s already exists in partition %s", v, lp.partitions[i].Name)
			}
		}
	}

	partition.ID = len(lp.partitions)
	partition.Values = values
	partition.CreatedAt = time.Now()
	partition.LastAccess = time.Now()

	// 创建值映射
	valueMap := make(map[string]bool)
	for _, v := range values {
		valueMap[v] = true
	}

	lp.partitions = append(lp.partitions, partition)
	lp.valueMaps = append(lp.valueMaps, valueMap)

	return nil
}

// GetPartitions 获取所有分区
func (lp *ListPartition) GetPartitions() []*Partition {
	lp.mu.RLock()
	defer lp.mu.RUnlock()

	result := make([]*Partition, len(lp.partitions))
	copy(result, lp.partitions)
	return result
}

// CreatePartition 创建分区
func (lp *ListPartition) CreatePartition(partition *Partition) error {
	return fmt.Errorf("use CreateListPartition for list partition")
}

// DropPartition 删除分区
func (lp *ListPartition) DropPartition(partitionID int) error {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	for i, p := range lp.partitions {
		if p.ID == partitionID {
			lp.partitions = append(lp.partitions[:i], lp.partitions[i+1:]...)
			lp.valueMaps = append(lp.valueMaps[:i], lp.valueMaps[i+1:]...)
			return nil
		}
	}

	return ErrPartitionNotFound
}

// GetPartitionCount 获取分区数量
func (lp *ListPartition) GetPartitionCount() int {
	lp.mu.RLock()
	defer lp.mu.RUnlock()
	return len(lp.partitions)
}

// GetType 获取分区类型
func (lp *ListPartition) GetType() string {
	return "list"
}

// GetActivePartitions 获取活跃分区
func (lp *ListPartition) GetActivePartitions() []*Partition {
	lp.mu.RLock()
	defer lp.mu.RUnlock()

	result := make([]*Partition, 0)
	for _, p := range lp.partitions {
		if p.Status == PartitionStatusActive {
			result = append(result, p)
		}
	}
	return result
}

// AddValuesToPartition 向分区添加值
func (lp *ListPartition) AddValuesToPartition(partitionID int, values []string) error {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	for i, p := range lp.partitions {
		if p.ID == partitionID {
			// 检查值是否已被其他分区使用
			for j, valueMap := range lp.valueMaps {
				if i == j {
					continue
				}
				for _, v := range values {
					if valueMap[v] {
						return fmt.Errorf("value %s already exists in partition %s", v, lp.partitions[j].Name)
					}
				}
			}

			// 添加值
			p.Values = append(p.Values, values...)
			for _, v := range values {
				lp.valueMaps[i][v] = true
			}
			return nil
		}
	}

	return ErrPartitionNotFound
}

// PartitionManager 分区管理器
type PartitionManager struct {
	mu            sync.RWMutex
	strategies    map[string]PartitionStrategy
	defaultStrategy PartitionStrategy
	scheduler     *PartitionScheduler
	stats         *PartitionStats
}

// PartitionStats 分区统计
type PartitionStats struct {
	mu              sync.RWMutex
	totalPartitions int
	activePartitions int
	totalSize       int64
	totalRows       int64
	partitionCounts map[string]int
}

// NewPartitionStats 创建分区统计
func NewPartitionStats() *PartitionStats {
	return &PartitionStats{
		partitionCounts: make(map[string]int),
	}
}

// Update 更新统计
func (s *PartitionStats) Update(strategyType string, partitions []*Partition) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalPartitions = 0
	s.activePartitions = 0
	s.totalSize = 0
	s.totalRows = 0

	for _, p := range partitions {
		s.totalPartitions++
		if p.Status == PartitionStatusActive {
			s.activePartitions++
		}
		s.totalSize += p.Size
		s.totalRows += p.RowCount
	}

	s.partitionCounts[strategyType] = len(partitions)
}

// GetStats 获取统计信息
func (s *PartitionStats) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"totalPartitions":  s.totalPartitions,
		"activePartitions": s.activePartitions,
		"totalSize":        s.totalSize,
		"totalRows":        s.totalRows,
		"partitionCounts":  s.partitionCounts,
	}
}

// NewPartitionManager 创建分区管理器
func NewPartitionManager(defaultStrategy PartitionStrategy) *PartitionManager {
	pm := &PartitionManager{
		strategies:      make(map[string]PartitionStrategy),
		defaultStrategy: defaultStrategy,
		stats:           NewPartitionStats(),
	}
	pm.strategies["default"] = defaultStrategy
	pm.scheduler = NewPartitionScheduler(pm)
	return pm
}

// RegisterStrategy 注册分区策略
func (pm *PartitionManager) RegisterStrategy(name string, strategy PartitionStrategy) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.strategies[name] = strategy
}

// GetStrategy 获取分区策略
func (pm *PartitionManager) GetStrategy(name string) (PartitionStrategy, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	s, ok := pm.strategies[name]
	return s, ok
}

// GetPartition 获取分区
func (pm *PartitionManager) GetPartition(key *PartitionKey) (*Partition, error) {
	return pm.defaultStrategy.GetPartition(key)
}

// GetPartitionWithStrategy 使用指定策略获取分区
func (pm *PartitionManager) GetPartitionWithStrategy(strategyName string, key *PartitionKey) (*Partition, error) {
	strategy, ok := pm.GetStrategy(strategyName)
	if !ok {
		return nil, fmt.Errorf("strategy %s not found", strategyName)
	}
	return strategy.GetPartition(key)
}

// CreatePartition 创建分区
func (pm *PartitionManager) CreatePartition(partition *Partition) error {
	return pm.defaultStrategy.CreatePartition(partition)
}

// DropPartition 删除分区
func (pm *PartitionManager) DropPartition(partitionID int) error {
	return pm.defaultStrategy.DropPartition(partitionID)
}

// GetPartitions 获取所有分区
func (pm *PartitionManager) GetPartitions() []*Partition {
	return pm.defaultStrategy.GetPartitions()
}

// GetActivePartitions 获取活跃分区
func (pm *PartitionManager) GetActivePartitions() []*Partition {
	return pm.defaultStrategy.GetActivePartitions()
}

// GetStats 获取统计信息
func (pm *PartitionManager) GetStats() map[string]interface{} {
	return pm.stats.GetStats()
}

// StartScheduler 启动调度器
func (pm *PartitionManager) StartScheduler() {
	pm.scheduler.Start()
}

// StopScheduler 停止调度器
func (pm *PartitionManager) StopScheduler() {
	pm.scheduler.Stop()
}

// PartitionScheduler 分区调度器
type PartitionScheduler struct {
	mu          sync.Mutex
	manager     *PartitionManager
	running     bool
	stopChan    chan struct{}
	tasks       []ScheduledTask
}

// ScheduledTask 调度任务
type ScheduledTask struct {
	Name     string
	Interval time.Duration
	Handler  func() error
	LastRun  time.Time
}

// NewPartitionScheduler 创建分区调度器
func NewPartitionScheduler(manager *PartitionManager) *PartitionScheduler {
	return &PartitionScheduler{
		manager:  manager,
		stopChan: make(chan struct{}),
		tasks:    make([]ScheduledTask, 0),
	}
}

// Start 启动调度器
func (s *PartitionScheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	// 添加默认任务
	s.addDefaultTasks()

	go s.run()
}

// addDefaultTasks 添加默认任务
func (s *PartitionScheduler) addDefaultTasks() {
	// 清理过期分区任务
	s.tasks = append(s.tasks, ScheduledTask{
		Name:     "purge_old_partitions",
		Interval: time.Hour,
		Handler:  s.purgeOldPartitions,
	})

	// 更新统计任务
	s.tasks = append(s.tasks, ScheduledTask{
		Name:     "update_stats",
		Interval: time.Minute * 5,
		Handler:  s.updateStats,
	})
}

// run 运行调度器
func (s *PartitionScheduler) run() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.executeTasks()
		case <-s.stopChan:
			return
		}
	}
}

// executeTasks 执行任务
func (s *PartitionScheduler) executeTasks() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for i := range s.tasks {
		task := &s.tasks[i]
		if now.Sub(task.LastRun) >= task.Interval {
			task.Handler()
			task.LastRun = now
		}
	}
}

// purgeOldPartitions 清理过期分区
func (s *PartitionScheduler) purgeOldPartitions() error {
	// 检查是否是时间分区
	if tp, ok := s.manager.defaultStrategy.(*TimePartition); ok {
		_, err := tp.PurgeOldPartitions()
		return err
	}
	return nil
}

// updateStats 更新统计
func (s *PartitionScheduler) updateStats() error {
	partitions := s.manager.GetPartitions()
	s.manager.stats.Update(s.manager.defaultStrategy.GetType(), partitions)
	return nil
}

// Stop 停止调度器
func (s *PartitionScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false
	close(s.stopChan)
}

// AddTask 添加调度任务
func (s *PartitionScheduler) AddTask(name string, interval time.Duration, handler func() error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks = append(s.tasks, ScheduledTask{
		Name:     name,
		Interval: interval,
		Handler:  handler,
	})
}

// AutoPartitionCreator 自动分区创建器
type AutoPartitionCreator struct {
	mu           sync.Mutex
	strategy     *TimePartition
	lookAhead    time.Duration // 提前创建时间
	checkInterval time.Duration
	running      bool
	stopChan     chan struct{}
}

// NewAutoPartitionCreator 创建自动分区创建器
func NewAutoPartitionCreator(strategy *TimePartition, lookAhead time.Duration) *AutoPartitionCreator {
	return &AutoPartitionCreator{
		strategy:      strategy,
		lookAhead:     lookAhead,
		checkInterval: time.Hour,
		stopChan:      make(chan struct{}),
	}
}

// Start 启动自动创建
func (c *AutoPartitionCreator) Start() {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return
	}
	c.running = true
	c.mu.Unlock()

	go c.run()
}

// run 运行自动创建
func (c *AutoPartitionCreator) run() {
	ticker := time.NewTicker(c.checkInterval)
	defer ticker.Stop()

	// 立即执行一次
	c.createAhead()

	for {
		select {
		case <-ticker.C:
			c.createAhead()
		case <-c.stopChan:
			return
		}
	}
}

// createAhead 提前创建分区
func (c *AutoPartitionCreator) createAhead() {
	now := time.Now()
	future := now.Add(c.lookAhead)

	// 检查未来时间点是否已有分区
	key := &PartitionKey{
		Timestamp: future,
	}

	_, err := c.strategy.GetPartition(key)
	if err == ErrPartitionNotFound {
		// 分区不存在，尝试创建
		c.strategy.GetPartition(key)
	}
}

// Stop 停止自动创建
func (c *AutoPartitionCreator) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return
	}

	c.running = false
	close(c.stopChan)
}

// PartitionPruner 分区清理器
type PartitionPruner struct {
	mu          sync.Mutex
	strategy    *TimePartition
	retention   time.Duration
	checkInterval time.Duration
	running     bool
	stopChan    chan struct{}
	onPurge     func(partitions []*Partition)
}

// NewPartitionPruner 创建分区清理器
func NewPartitionPruner(strategy *TimePartition, retention time.Duration) *PartitionPruner {
	return &PartitionPruner{
		strategy:      strategy,
		retention:     retention,
		checkInterval: time.Hour * 6,
		stopChan:      make(chan struct{}),
	}
}

// SetPurgeHandler 设置清理处理函数
func (p *PartitionPruner) SetPurgeHandler(handler func(partitions []*Partition)) {
	p.onPurge = handler
}

// Start 启动清理器
func (p *PartitionPruner) Start() {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return
	}
	p.running = true
	p.mu.Unlock()

	go p.run()
}

// run 运行清理器
func (p *PartitionPruner) run() {
	ticker := time.NewTicker(p.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.prune()
		case <-p.stopChan:
			return
		}
	}
}

// prune 执行清理
func (p *PartitionPruner) prune() {
	partitions, err := p.strategy.PurgeOldPartitions()
	if err != nil {
		return
	}

	if p.onPurge != nil && len(partitions) > 0 {
		p.onPurge(partitions)
	}
}

// Stop 停止清理器
func (p *PartitionPruner) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return
	}

	p.running = false
	close(p.stopChan)
}

// PartitionInfo 分区信息
type PartitionInfo struct {
	ID          int
	Name        string
	Type        string
	Status      string
	Size        int64
	RowCount    int64
	StartTime   time.Time
	EndTime     time.Time
	CreatedAt   time.Time
	LastAccess  time.Time
}

// GetPartitionInfo 获取分区信息
func GetPartitionInfo(p *Partition, partitionType string) *PartitionInfo {
	return &PartitionInfo{
		ID:         p.ID,
		Name:       p.Name,
		Type:       partitionType,
		Status:     p.Status.String(),
		Size:       p.Size,
		RowCount:   p.RowCount,
		StartTime:  p.StartTime,
		EndTime:    p.EndTime,
		CreatedAt:  p.CreatedAt,
		LastAccess: p.LastAccess,
	}
}

// PartitionList 分区列表（用于排序）
type PartitionList []*Partition

// Len 实现sort.Interface
func (l PartitionList) Len() int {
	return len(l)
}

// Less 实现sort.Interface
func (l PartitionList) Less(i, j int) bool {
	return l[i].StartTime.Before(l[j].StartTime)
}

// Swap 实现sort.Interface
func (l PartitionList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// SortPartitionsByTime 按时间排序分区
func SortPartitionsByTime(partitions []*Partition) []*Partition {
	result := make([]*Partition, len(partitions))
	copy(result, partitions)
	sort.Sort(PartitionList(result))
	return result
}

// SortPartitionsBySize 按大小排序分区
func SortPartitionsBySize(partitions []*Partition, desc bool) []*Partition {
	result := make([]*Partition, len(partitions))
	copy(result, partitions)

	sort.Slice(result, func(i, j int) bool {
		if desc {
			return result[i].Size > result[j].Size
		}
		return result[i].Size < result[j].Size
	})

	return result
}

// SortPartitionsByAccess 按访问时间排序分区
func SortPartitionsByAccess(partitions []*Partition, desc bool) []*Partition {
	result := make([]*Partition, len(partitions))
	copy(result, partitions)

	sort.Slice(result, func(i, j int) bool {
		if desc {
			return result[i].LastAccess.After(result[j].LastAccess)
		}
		return result[i].LastAccess.Before(result[j].LastAccess)
	})

	return result
}
