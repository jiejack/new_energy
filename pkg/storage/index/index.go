// Package index 提供时序数据索引实现
// 支持时间索引、标签索引、复合索引等多种索引类型
package index

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// 定义错误类型
var (
	ErrIndexNotFound      = errors.New("index not found")
	ErrInvalidIndexKey    = errors.New("invalid index key")
	ErrIndexAlreadyExists = errors.New("index already exists")
	ErrInvalidIndexType   = errors.New("invalid index type")
	ErrIndexCorrupted     = errors.New("index corrupted")
)

// IndexType 索引类型
type IndexType int

const (
	IndexTypeTime IndexType = iota
	IndexTypeTag
	IndexTypeComposite
	IndexTypeBloom
	IndexTypeBitmap
)

// String 返回索引类型的字符串表示
func (t IndexType) String() string {
	switch t {
	case IndexTypeTime:
		return "time"
	case IndexTypeTag:
		return "tag"
	case IndexTypeComposite:
		return "composite"
	case IndexTypeBloom:
		return "bloom"
	case IndexTypeBitmap:
		return "bitmap"
	default:
		return "unknown"
	}
}

// IndexEntry 索引条目
type IndexEntry struct {
	Key       string    // 索引键
	Value     string    // 索引值
	RowID     int64     // 行ID
	Timestamp time.Time // 时间戳
	Size      int64     // 数据大小
}

// IndexStats 索引统计
type IndexStats struct {
	mu            sync.RWMutex
	TotalEntries  int64 // 总条目数
	TotalSize     int64 // 总大小
	IndexCount    int64 // 索引数量
	QueryCount    int64 // 查询次数
	HitCount      int64 // 命中次数
	MissCount     int64 // 未命中次数
	RebuildCount  int64 // 重建次数
	LastRebuildAt time.Time
}

// NewIndexStats 创建索引统计
func NewIndexStats() *IndexStats {
	return &IndexStats{}
}

// RecordQuery 记录查询
func (s *IndexStats) RecordQuery(hit bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.QueryCount++
	if hit {
		s.HitCount++
	} else {
		s.MissCount++
	}
}

// RecordRebuild 记录重建
func (s *IndexStats) RecordRebuild() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.RebuildCount++
	s.LastRebuildAt = time.Now()
}

// GetHitRate 获取命中率
func (s *IndexStats) GetHitRate() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.QueryCount == 0 {
		return 0
	}
	return float64(s.HitCount) / float64(s.QueryCount)
}

// Snapshot 获取统计快照
func (s *IndexStats) Snapshot() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"totalEntries":  s.TotalEntries,
		"totalSize":     s.TotalSize,
		"indexCount":    s.IndexCount,
		"queryCount":    s.QueryCount,
		"hitCount":      s.HitCount,
		"missCount":     s.MissCount,
		"hitRate":       s.GetHitRate(),
		"rebuildCount":  s.RebuildCount,
		"lastRebuildAt": s.LastRebuildAt,
	}
}

// Index 索引接口
type Index interface {
	// Insert 插入条目
	Insert(entry *IndexEntry) error

	// Delete 删除条目
	Delete(key string) error

	// Lookup 查找条目
	Lookup(key string) ([]*IndexEntry, error)

	// Range 范围查找
	Range(start, end string) ([]*IndexEntry, error)

	// GetType 获取索引类型
	GetType() IndexType

	// GetName 获取索引名称
	GetName() string

	// Size 获取索引大小
	Size() int64

	// Count 获取条目数量
	Count() int64

	// Clear 清空索引
	Clear() error

	// Rebuild 重建索引
	Rebuild(entries []*IndexEntry) error
}

// TimeIndex 时间索引
type TimeIndex struct {
	mu       sync.RWMutex
	name     string
	entries  map[int64][]*IndexEntry // 时间戳(秒) -> 条目列表
	sorted   []int64                 // 排序的时间戳
	stats    *IndexStats
}

// NewTimeIndex 创建时间索引
func NewTimeIndex(name string) *TimeIndex {
	return &TimeIndex{
		name:    name,
		entries: make(map[int64][]*IndexEntry),
		sorted:  make([]int64, 0),
		stats:   NewIndexStats(),
	}
}

// Insert 插入条目
func (idx *TimeIndex) Insert(entry *IndexEntry) error {
	if entry == nil {
		return ErrInvalidIndexKey
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	ts := entry.Timestamp.Unix()

	// 检查是否已存在该时间戳
	if _, exists := idx.entries[ts]; !exists {
		// 插入到排序列表
		pos := sort.Search(len(idx.sorted), func(i int) bool {
			return idx.sorted[i] >= ts
		})
		idx.sorted = append(idx.sorted, 0)
		copy(idx.sorted[pos+1:], idx.sorted[pos:])
		idx.sorted[pos] = ts
	}

	idx.entries[ts] = append(idx.entries[ts], entry)
	atomic.AddInt64(&idx.stats.TotalEntries, 1)
	atomic.AddInt64(&idx.stats.TotalSize, int64(len(entry.Key)+len(entry.Value)))

	return nil
}

// Delete 删除条目
func (idx *TimeIndex) Delete(key string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	for ts, entries := range idx.entries {
		newEntries := make([]*IndexEntry, 0)
		for _, e := range entries {
			if e.Key != key {
				newEntries = append(newEntries, e)
			}
		}

		if len(newEntries) == 0 {
			delete(idx.entries, ts)
			// 从排序列表中移除
			pos := sort.Search(len(idx.sorted), func(i int) bool {
				return idx.sorted[i] >= ts
			})
			if pos < len(idx.sorted) && idx.sorted[pos] == ts {
				idx.sorted = append(idx.sorted[:pos], idx.sorted[pos+1:]...)
			}
		} else {
			idx.entries[ts] = newEntries
		}
	}

	return nil
}

// Lookup 查找条目
func (idx *TimeIndex) Lookup(key string) ([]*IndexEntry, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	// 时间索引不支持精确键查找
	idx.stats.RecordQuery(false)
	return nil, ErrInvalidIndexKey
}

// Range 范围查找
func (idx *TimeIndex) Range(start, end string) ([]*IndexEntry, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	// 解析时间范围
	startTime, err := time.Parse(time.RFC3339, start)
	if err != nil {
		return nil, fmt.Errorf("invalid start time: %v", err)
	}

	endTime, err := time.Parse(time.RFC3339, end)
	if err != nil {
		return nil, fmt.Errorf("invalid end time: %v", err)
	}

	startTs := startTime.Unix()
	endTs := endTime.Unix()

	// 二分查找起始位置
	startPos := sort.Search(len(idx.sorted), func(i int) bool {
		return idx.sorted[i] >= startTs
	})

	// 二分查找结束位置
	endPos := sort.Search(len(idx.sorted), func(i int) bool {
		return idx.sorted[i] > endTs
	})

	result := make([]*IndexEntry, 0)
	for i := startPos; i < endPos && i < len(idx.sorted); i++ {
		ts := idx.sorted[i]
		result = append(result, idx.entries[ts]...)
	}

	idx.stats.RecordQuery(len(result) > 0)
	return result, nil
}

// GetType 获取索引类型
func (idx *TimeIndex) GetType() IndexType {
	return IndexTypeTime
}

// GetName 获取索引名称
func (idx *TimeIndex) GetName() string {
	return idx.name
}

// Size 获取索引大小
func (idx *TimeIndex) Size() int64 {
	return atomic.LoadInt64(&idx.stats.TotalSize)
}

// Count 获取条目数量
func (idx *TimeIndex) Count() int64 {
	return atomic.LoadInt64(&idx.stats.TotalEntries)
}

// Clear 清空索引
func (idx *TimeIndex) Clear() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.entries = make(map[int64][]*IndexEntry)
	idx.sorted = make([]int64, 0)
	atomic.StoreInt64(&idx.stats.TotalEntries, 0)
	atomic.StoreInt64(&idx.stats.TotalSize, 0)

	return nil
}

// Rebuild 重建索引
func (idx *TimeIndex) Rebuild(entries []*IndexEntry) error {
	idx.Clear()
	idx.stats.RecordRebuild()

	for _, entry := range entries {
		if err := idx.Insert(entry); err != nil {
			return err
		}
	}

	return nil
}

// GetEntriesByTimeRange 根据时间范围获取条目
func (idx *TimeIndex) GetEntriesByTimeRange(start, end time.Time) ([]*IndexEntry, error) {
	return idx.Range(start.Format(time.RFC3339), end.Format(time.RFC3339))
}

// TagIndex 标签索引
type TagIndex struct {
	mu      sync.RWMutex
	name    string
	entries map[string]map[string][]*IndexEntry // tagKey -> tagValue -> 条目列表
	stats   *IndexStats
}

// NewTagIndex 创建标签索引
func NewTagIndex(name string) *TagIndex {
	return &TagIndex{
		name:    name,
		entries: make(map[string]map[string][]*IndexEntry),
		stats:   NewIndexStats(),
	}
}

// Insert 插入条目
func (idx *TagIndex) Insert(entry *IndexEntry) error {
	if entry == nil {
		return ErrInvalidIndexKey
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	// 解析标签键值对
	tagKey, tagValue := parseTagKey(entry.Key)
	if tagKey == "" {
		return ErrInvalidIndexKey
	}

	// 确保map存在
	if idx.entries[tagKey] == nil {
		idx.entries[tagKey] = make(map[string][]*IndexEntry)
	}

	idx.entries[tagKey][tagValue] = append(idx.entries[tagKey][tagValue], entry)
	atomic.AddInt64(&idx.stats.TotalEntries, 1)
	atomic.AddInt64(&idx.stats.TotalSize, int64(len(entry.Key)+len(entry.Value)))

	return nil
}

// parseTagKey 解析标签键
func parseTagKey(key string) (string, string) {
	for i := 0; i < len(key); i++ {
		if key[i] == '=' {
			return key[:i], key[i+1:]
		}
	}
	return key, ""
}

// Delete 删除条目
func (idx *TagIndex) Delete(key string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	tagKey, tagValue := parseTagKey(key)
	if tagKey == "" {
		return ErrInvalidIndexKey
	}

	if values, ok := idx.entries[tagKey]; ok {
		if tagValue == "" {
			// 删除整个标签键
			delete(idx.entries, tagKey)
		} else {
			// 删除特定标签值
			delete(values, tagValue)
			if len(values) == 0 {
				delete(idx.entries, tagKey)
			}
		}
	}

	return nil
}

// Lookup 查找条目
func (idx *TagIndex) Lookup(key string) ([]*IndexEntry, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	tagKey, tagValue := parseTagKey(key)
	if tagKey == "" {
		idx.stats.RecordQuery(false)
		return nil, ErrInvalidIndexKey
	}

	values, ok := idx.entries[tagKey]
	if !ok {
		idx.stats.RecordQuery(false)
		return nil, ErrIndexNotFound
	}

	if tagValue == "" {
		// 返回该标签键下的所有条目
		result := make([]*IndexEntry, 0)
		for _, entries := range values {
			result = append(result, entries...)
		}
		idx.stats.RecordQuery(len(result) > 0)
		return result, nil
	}

	entries, ok := values[tagValue]
	if !ok {
		idx.stats.RecordQuery(false)
		return nil, ErrIndexNotFound
	}

	result := make([]*IndexEntry, len(entries))
	copy(result, entries)
	idx.stats.RecordQuery(true)
	return result, nil
}

// Range 范围查找
func (idx *TagIndex) Range(start, end string) ([]*IndexEntry, error) {
	// 标签索引不支持范围查找
	return nil, ErrInvalidIndexType
}

// GetType 获取索引类型
func (idx *TagIndex) GetType() IndexType {
	return IndexTypeTag
}

// GetName 获取索引名称
func (idx *TagIndex) GetName() string {
	return idx.name
}

// Size 获取索引大小
func (idx *TagIndex) Size() int64 {
	return atomic.LoadInt64(&idx.stats.TotalSize)
}

// Count 获取条目数量
func (idx *TagIndex) Count() int64 {
	return atomic.LoadInt64(&idx.stats.TotalEntries)
}

// Clear 清空索引
func (idx *TagIndex) Clear() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.entries = make(map[string]map[string][]*IndexEntry)
	atomic.StoreInt64(&idx.stats.TotalEntries, 0)
	atomic.StoreInt64(&idx.stats.TotalSize, 0)

	return nil
}

// Rebuild 重建索引
func (idx *TagIndex) Rebuild(entries []*IndexEntry) error {
	idx.Clear()
	idx.stats.RecordRebuild()

	for _, entry := range entries {
		if err := idx.Insert(entry); err != nil {
			return err
		}
	}

	return nil
}

// GetTagKeys 获取所有标签键
func (idx *TagIndex) GetTagKeys() []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	keys := make([]string, 0, len(idx.entries))
	for k := range idx.entries {
		keys = append(keys, k)
	}
	return keys
}

// GetTagValues 获取标签值
func (idx *TagIndex) GetTagValues(tagKey string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	values, ok := idx.entries[tagKey]
	if !ok {
		return nil
	}

	result := make([]string, 0, len(values))
	for v := range values {
		result = append(result, v)
	}
	return result
}

// CompositeIndex 复合索引
type CompositeIndex struct {
	mu      sync.RWMutex
	name    string
	fields  []string // 索引字段
	entries map[string][]*IndexEntry
	stats   *IndexStats
}

// NewCompositeIndex 创建复合索引
func NewCompositeIndex(name string, fields []string) *CompositeIndex {
	return &CompositeIndex{
		name:    name,
		fields:  fields,
		entries: make(map[string][]*IndexEntry),
		stats:   NewIndexStats(),
	}
}

// buildCompositeKey 构建复合键
func (idx *CompositeIndex) buildCompositeKey(entry *IndexEntry) string {
	// 使用Key和Value组合
	return fmt.Sprintf("%s|%s", entry.Key, entry.Value)
}

// Insert 插入条目
func (idx *CompositeIndex) Insert(entry *IndexEntry) error {
	if entry == nil {
		return ErrInvalidIndexKey
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	key := idx.buildCompositeKey(entry)
	idx.entries[key] = append(idx.entries[key], entry)
	atomic.AddInt64(&idx.stats.TotalEntries, 1)
	atomic.AddInt64(&idx.stats.TotalSize, int64(len(key)))

	return nil
}

// Delete 删除条目
func (idx *CompositeIndex) Delete(key string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	delete(idx.entries, key)
	return nil
}

// Lookup 查找条目
func (idx *CompositeIndex) Lookup(key string) ([]*IndexEntry, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	entries, ok := idx.entries[key]
	if !ok {
		idx.stats.RecordQuery(false)
		return nil, ErrIndexNotFound
	}

	result := make([]*IndexEntry, len(entries))
	copy(result, entries)
	idx.stats.RecordQuery(true)
	return result, nil
}

// Range 范围查找
func (idx *CompositeIndex) Range(start, end string) ([]*IndexEntry, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	result := make([]*IndexEntry, 0)

	// 收集范围内的条目
	for key, entries := range idx.entries {
		if key >= start && key <= end {
			result = append(result, entries...)
		}
	}

	idx.stats.RecordQuery(len(result) > 0)
	return result, nil
}

// GetType 获取索引类型
func (idx *CompositeIndex) GetType() IndexType {
	return IndexTypeComposite
}

// GetName 获取索引名称
func (idx *CompositeIndex) GetName() string {
	return idx.name
}

// Size 获取索引大小
func (idx *CompositeIndex) Size() int64 {
	return atomic.LoadInt64(&idx.stats.TotalSize)
}

// Count 获取条目数量
func (idx *CompositeIndex) Count() int64 {
	return atomic.LoadInt64(&idx.stats.TotalEntries)
}

// Clear 清空索引
func (idx *CompositeIndex) Clear() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.entries = make(map[string][]*IndexEntry)
	atomic.StoreInt64(&idx.stats.TotalEntries, 0)
	atomic.StoreInt64(&idx.stats.TotalSize, 0)

	return nil
}

// Rebuild 重建索引
func (idx *CompositeIndex) Rebuild(entries []*IndexEntry) error {
	idx.Clear()
	idx.stats.RecordRebuild()

	for _, entry := range entries {
		if err := idx.Insert(entry); err != nil {
			return err
		}
	}

	return nil
}

// GetFields 获取索引字段
func (idx *CompositeIndex) GetFields() []string {
	return idx.fields
}

// BloomIndex 布隆过滤器索引
type BloomIndex struct {
	mu        sync.RWMutex
	name      string
	bitArray  []uint64
	size      uint
	hashCount uint
	stats     *IndexStats
}

// NewBloomIndex 创建布隆过滤器索引
func NewBloomIndex(name string, expectedItems uint, falsePositiveRate float64) *BloomIndex {
	// 计算最优参数
	size := optimalBloomSize(expectedItems, falsePositiveRate)
	hashCount := optimalHashCount(size, expectedItems)

	return &BloomIndex{
		name:      name,
		bitArray:  make([]uint64, (size+63)/64),
		size:      size,
		hashCount: hashCount,
		stats:     NewIndexStats(),
	}
}

// optimalBloomSize 计算最优布隆过滤器大小
func optimalBloomSize(n uint, p float64) uint {
	// m = -n * ln(p) / (ln(2)^2)
	m := float64(n) * -ln(p) / (ln(2) * ln(2))
	return uint(m)
}

// optimalHashCount 计算最优哈希函数数量
func optimalHashCount(m, n uint) uint {
	// k = m/n * ln(2)
	k := float64(m) / float64(n) * ln(2)
	return uint(k)
}

func ln(x float64) float64 {
	// 简化的自然对数计算
	if x <= 0 {
		return 0
	}
	// 使用近似值
	result := 0.0
	for i := 1; i <= 100; i++ {
		result += pow(-1.0, float64(i+1)) * pow(x-1, float64(i)) / float64(i)
	}
	return result
}

func pow(x, y float64) float64 {
	result := 1.0
	for i := 0; i < int(y); i++ {
		result *= x
	}
	return result
}

// hash 计算哈希值
func (idx *BloomIndex) hash(data []byte, seed uint) uint {
	h := uint(0)
	for _, b := range data {
		h = h*31 + uint(b) + seed
	}
	return h % idx.size
}

// Insert 插入条目
func (idx *BloomIndex) Insert(entry *IndexEntry) error {
	if entry == nil {
		return ErrInvalidIndexKey
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	data := []byte(entry.Key)

	for i := uint(0); i < idx.hashCount; i++ {
		pos := idx.hash(data, i)
		arrayIdx := pos / 64
		bitIdx := pos % 64
		idx.bitArray[arrayIdx] |= 1 << bitIdx
	}

	atomic.AddInt64(&idx.stats.TotalEntries, 1)
	return nil
}

// Delete 删除条目
func (idx *BloomIndex) Delete(key string) error {
	// 布隆过滤器不支持删除
	return ErrInvalidIndexType
}

// Lookup 查找条目
func (idx *BloomIndex) Lookup(key string) ([]*IndexEntry, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	data := []byte(key)
	mayExist := true

	for i := uint(0); i < idx.hashCount && mayExist; i++ {
		pos := idx.hash(data, i)
		arrayIdx := pos / 64
		bitIdx := pos % 64
		if idx.bitArray[arrayIdx]&(1<<bitIdx) == 0 {
			mayExist = false
		}
	}

	idx.stats.RecordQuery(mayExist)

	if mayExist {
		// 可能存在，返回空条目表示需要进一步检查
		return []*IndexEntry{{Key: key}}, nil
	}

	return nil, ErrIndexNotFound
}

// Range 范围查找
func (idx *BloomIndex) Range(start, end string) ([]*IndexEntry, error) {
	return nil, ErrInvalidIndexType
}

// GetType 获取索引类型
func (idx *BloomIndex) GetType() IndexType {
	return IndexTypeBloom
}

// GetName 获取索引名称
func (idx *BloomIndex) GetName() string {
	return idx.name
}

// Size 获取索引大小
func (idx *BloomIndex) Size() int64 {
	return int64(len(idx.bitArray) * 8)
}

// Count 获取条目数量
func (idx *BloomIndex) Count() int64 {
	return atomic.LoadInt64(&idx.stats.TotalEntries)
}

// Clear 清空索引
func (idx *BloomIndex) Clear() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	for i := range idx.bitArray {
		idx.bitArray[i] = 0
	}
	atomic.StoreInt64(&idx.stats.TotalEntries, 0)

	return nil
}

// Rebuild 重建索引
func (idx *BloomIndex) Rebuild(entries []*IndexEntry) error {
	idx.Clear()
	idx.stats.RecordRebuild()

	for _, entry := range entries {
		if err := idx.Insert(entry); err != nil {
			return err
		}
	}

	return nil
}

// MayContain 检查是否可能包含
func (idx *BloomIndex) MayContain(key string) bool {
	entries, err := idx.Lookup(key)
	return err == nil && len(entries) > 0
}

// IndexManager 索引管理器
type IndexManager struct {
	mu       sync.RWMutex
	indexes  map[string]Index
	stats    *IndexStats
	config   *IndexConfig
}

// IndexConfig 索引配置
type IndexConfig struct {
	AutoRebuild     bool
	RebuildInterval time.Duration
	MaxIndexSize    int64
}

// DefaultIndexConfig 默认索引配置
func DefaultIndexConfig() *IndexConfig {
	return &IndexConfig{
		AutoRebuild:     true,
		RebuildInterval: time.Hour * 24,
		MaxIndexSize:    1024 * 1024 * 1024, // 1GB
	}
}

// NewIndexManager 创建索引管理器
func NewIndexManager(config *IndexConfig) *IndexManager {
	if config == nil {
		config = DefaultIndexConfig()
	}

	return &IndexManager{
		indexes: make(map[string]Index),
		stats:   NewIndexStats(),
		config:  config,
	}
}

// CreateIndex 创建索引
func (m *IndexManager) CreateIndex(name string, indexType IndexType, opts ...interface{}) (Index, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.indexes[name]; exists {
		return nil, ErrIndexAlreadyExists
	}

	var index Index
	switch indexType {
	case IndexTypeTime:
		index = NewTimeIndex(name)
	case IndexTypeTag:
		index = NewTagIndex(name)
	case IndexTypeComposite:
		if len(opts) > 0 {
			if fields, ok := opts[0].([]string); ok {
				index = NewCompositeIndex(name, fields)
			}
		} else {
			index = NewCompositeIndex(name, []string{"default"})
		}
	case IndexTypeBloom:
		expectedItems := uint(10000)
		falsePositiveRate := 0.01
		if len(opts) >= 2 {
			if n, ok := opts[0].(uint); ok {
				expectedItems = n
			}
			if p, ok := opts[1].(float64); ok {
				falsePositiveRate = p
			}
		}
		index = NewBloomIndex(name, expectedItems, falsePositiveRate)
	default:
		return nil, ErrInvalidIndexType
	}

	m.indexes[name] = index
	atomic.AddInt64(&m.stats.IndexCount, 1)

	return index, nil
}

// GetIndex 获取索引
func (m *IndexManager) GetIndex(name string) (Index, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	index, ok := m.indexes[name]
	if !ok {
		return nil, ErrIndexNotFound
	}

	return index, nil
}

// DropIndex 删除索引
func (m *IndexManager) DropIndex(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.indexes[name]; !ok {
		return ErrIndexNotFound
	}

	delete(m.indexes, name)
	atomic.AddInt64(&m.stats.IndexCount, -1)

	return nil
}

// ListIndexes 列出所有索引
func (m *IndexManager) ListIndexes() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.indexes))
	for name := range m.indexes {
		names = append(names, name)
	}
	return names
}

// InsertToIndex 向索引插入条目
func (m *IndexManager) InsertToIndex(name string, entry *IndexEntry) error {
	index, err := m.GetIndex(name)
	if err != nil {
		return err
	}

	return index.Insert(entry)
}

// LookupFromIndex 从索引查找条目
func (m *IndexManager) LookupFromIndex(name string, key string) ([]*IndexEntry, error) {
	index, err := m.GetIndex(name)
	if err != nil {
		return nil, err
	}

	return index.Lookup(key)
}

// RangeFromIndex 从索引范围查找
func (m *IndexManager) RangeFromIndex(name string, start, end string) ([]*IndexEntry, error) {
	index, err := m.GetIndex(name)
	if err != nil {
		return nil, err
	}

	return index.Range(start, end)
}

// RebuildIndex 重建索引
func (m *IndexManager) RebuildIndex(name string, entries []*IndexEntry) error {
	index, err := m.GetIndex(name)
	if err != nil {
		return err
	}

	m.stats.RecordRebuild()
	return index.Rebuild(entries)
}

// RebuildAll 重建所有索引
func (m *IndexManager) RebuildAll(entries []*IndexEntry) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, index := range m.indexes {
		if err := index.Rebuild(entries); err != nil {
			return fmt.Errorf("failed to rebuild index %s: %v", name, err)
		}
	}

	m.stats.RecordRebuild()
	return nil
}

// GetStats 获取统计信息
func (m *IndexManager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := m.stats.Snapshot()

	// 收集各索引统计
	indexStats := make(map[string]interface{})
	for name, index := range m.indexes {
		indexStats[name] = map[string]interface{}{
			"type":  index.GetType().String(),
			"size":  index.Size(),
			"count": index.Count(),
		}
	}
	stats["indexes"] = indexStats

	return stats
}

// GetTotalSize 获取总大小
func (m *IndexManager) GetTotalSize() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var total int64
	for _, index := range m.indexes {
		total += index.Size()
	}
	return total
}

// GetTotalCount 获取总条目数
func (m *IndexManager) GetTotalCount() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var total int64
	for _, index := range m.indexes {
		total += index.Count()
	}
	return total
}

// IndexBuilder 索引构建器
type IndexBuilder struct {
	manager *IndexManager
	batch   []*IndexEntry
	size    int
}

// NewIndexBuilder 创建索引构建器
func NewIndexBuilder(manager *IndexManager, batchSize int) *IndexBuilder {
	return &IndexBuilder{
		manager: manager,
		batch:   make([]*IndexEntry, 0, batchSize),
		size:    batchSize,
	}
}

// Add 添加条目
func (b *IndexBuilder) Add(entry *IndexEntry) error {
	b.batch = append(b.batch, entry)

	if len(b.batch) >= b.size {
		return b.Flush()
	}

	return nil
}

// Flush 刷新批次
func (b *IndexBuilder) Flush() error {
	if len(b.batch) == 0 {
		return nil
	}

	// 批量插入到所有索引
	for _, index := range b.manager.indexes {
		for _, entry := range b.batch {
			if err := index.Insert(entry); err != nil {
				return err
			}
		}
	}

	b.batch = b.batch[:0]
	return nil
}

// IndexQuery 索引查询
type IndexQuery struct {
	manager    *IndexManager
	indexNames []string
	filters    []QueryFilter
	orderBy    string
	limit      int
	offset     int
}

// QueryFilter 查询过滤器
type QueryFilter struct {
	Field    string
	Operator string
	Value    interface{}
}

// NewIndexQuery 创建索引查询
func NewIndexQuery(manager *IndexManager) *IndexQuery {
	return &IndexQuery{
		manager:    manager,
		indexNames: make([]string, 0),
		filters:    make([]QueryFilter, 0),
		limit:      1000,
	}
}

// UseIndex 使用索引
func (q *IndexQuery) UseIndex(name string) *IndexQuery {
	q.indexNames = append(q.indexNames, name)
	return q
}

// Where 添加过滤条件
func (q *IndexQuery) Where(field string, op string, value interface{}) *IndexQuery {
	q.filters = append(q.filters, QueryFilter{
		Field:    field,
		Operator: op,
		Value:    value,
	})
	return q
}

// OrderBy 设置排序
func (q *IndexQuery) OrderBy(field string) *IndexQuery {
	q.orderBy = field
	return q
}

// Limit 设置限制
func (q *IndexQuery) Limit(limit int) *IndexQuery {
	q.limit = limit
	return q
}

// Offset 设置偏移
func (q *IndexQuery) Offset(offset int) *IndexQuery {
	q.offset = offset
	return q
}

// Execute 执行查询
func (q *IndexQuery) Execute() ([]*IndexEntry, error) {
	if len(q.indexNames) == 0 {
		return nil, errors.New("no index specified")
	}

	// 从第一个索引获取结果
	index, err := q.manager.GetIndex(q.indexNames[0])
	if err != nil {
		return nil, err
	}

	var result []*IndexEntry

	// 根据过滤器类型执行查询
	for _, filter := range q.filters {
		switch filter.Operator {
		case "=", "==":
			entries, err := index.Lookup(fmt.Sprintf("%v", filter.Value))
			if err != nil {
				continue
			}
			result = append(result, entries...)
		case "range":
			if values, ok := filter.Value.([2]string); ok {
				entries, err := index.Range(values[0], values[1])
				if err != nil {
					continue
				}
				result = append(result, entries...)
			}
		}
	}

	// 应用偏移和限制
	if q.offset > 0 && q.offset < len(result) {
		result = result[q.offset:]
	}
	if q.limit > 0 && len(result) > q.limit {
		result = result[:q.limit]
	}

	return result, nil
}
