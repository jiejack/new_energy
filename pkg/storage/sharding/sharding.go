// Package sharding 提供时序数据分片策略实现
// 支持哈希分片、范围分片、时间分片等多种分片方式
package sharding

import (
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"sync"
	"time"
)

// 定义错误类型
var (
	ErrInvalidShardCount   = errors.New("invalid shard count")
	ErrInvalidShardKey     = errors.New("invalid shard key")
	ErrShardNotFound       = errors.New("shard not found")
	ErrRebalanceInProgress = errors.New("rebalance in progress")
	ErrInvalidConfig       = errors.New("invalid configuration")
)

// Shard 表示一个分片
type Shard struct {
	ID          int                    // 分片ID
	Name        string                 // 分片名称
	StartKey    []byte                 // 起始键（范围分片）
	EndKey      []byte                 // 结束键（范围分片）
	StartTime   time.Time              // 起始时间（时间分片）
	EndTime     time.Time              // 结束时间（时间分片）
	Weight      int                    // 权重
	Status      ShardStatus            // 状态
	Metadata    map[string]interface{} // 元数据
	NodeAddress string                 // 节点地址
}

// ShardStatus 分片状态
type ShardStatus int

const (
	ShardStatusActive ShardStatus = iota
	ShardStatusInactive
	ShardStatusMigrating
	ShardStatusRebalancing
)

// String 返回分片状态的字符串表示
func (s ShardStatus) String() string {
	switch s {
	case ShardStatusActive:
		return "active"
	case ShardStatusInactive:
		return "inactive"
	case ShardStatusMigrating:
		return "migrating"
	case ShardStatusRebalancing:
		return "rebalancing"
	default:
		return "unknown"
	}
}

// ShardKey 分片键
type ShardKey struct {
	DeviceID  string    // 设备ID
	PointID   string    // 测点ID
	Timestamp time.Time // 时间戳
	Tags      map[string]string
}

// ShardingStrategy 分片策略接口
type ShardingStrategy interface {
	// GetShard 根据分片键获取分片ID
	GetShard(key *ShardKey) (int, error)

	// GetShards 获取所有分片
	GetShards() []*Shard

	// AddShard 添加分片
	AddShard(shard *Shard) error

	// RemoveShard 移除分片
	RemoveShard(shardID int) error

	// GetShardCount 获取分片数量
	GetShardCount() int

	// GetType 获取分片类型
	GetType() string

	// Rebalance 重新平衡分片
	Rebalance() error
}

// HashSharding 哈希分片策略
type HashSharding struct {
	mu          sync.RWMutex
	shards      []*Shard
	shardCount  int
	hashFunc    HashFunc
	virtualNodes int // 虚拟节点数（一致性哈希）
}

// HashFunc 哈希函数类型
type HashFunc func([]byte) uint32

// NewHashSharding 创建哈希分片策略
func NewHashSharding(shardCount int, opts ...HashOption) (*HashSharding, error) {
	if shardCount <= 0 {
		return nil, ErrInvalidShardCount
	}

	hs := &HashSharding{
		shardCount:   shardCount,
		virtualNodes: 150, // 默认虚拟节点数
		hashFunc:     fnvHash,
	}

	for _, opt := range opts {
		opt(hs)
	}

	// 初始化分片
	hs.shards = make([]*Shard, shardCount)
	for i := 0; i < shardCount; i++ {
		hs.shards[i] = &Shard{
			ID:     i,
			Name:   fmt.Sprintf("shard-%d", i),
			Weight: 1,
			Status: ShardStatusActive,
			Metadata: map[string]interface{}{
				"type": "hash",
			},
		}
	}

	return hs, nil
}

// HashOption 哈希分片选项
type HashOption func(*HashSharding)

// WithVirtualNodes 设置虚拟节点数
func WithVirtualNodes(n int) HashOption {
	return func(hs *HashSharding) {
		hs.virtualNodes = n
	}
}

// WithHashFunc 设置哈希函数
func WithHashFunc(fn HashFunc) HashOption {
	return func(hs *HashSharding) {
		hs.hashFunc = fn
	}
}

// fnvHash FNV哈希函数
func fnvHash(data []byte) uint32 {
	h := fnv.New32a()
	h.Write(data)
	return h.Sum32()
}

// md5Hash MD5哈希函数
func md5Hash(data []byte) uint32 {
	h := md5.New()
	h.Write(data)
	sum := h.Sum(nil)
	return binary.BigEndian.Uint32(sum[:4])
}

// GetShard 根据分片键获取分片ID
func (hs *HashSharding) GetShard(key *ShardKey) (int, error) {
	if key == nil {
		return -1, ErrInvalidShardKey
	}

	hs.mu.RLock()
	defer hs.mu.RUnlock()

	// 构建哈希键
	hashKey := hs.buildHashKey(key)
	hash := hs.hashFunc(hashKey)

	// 使用一致性哈希算法
	if hs.virtualNodes > 0 {
		return hs.consistentHash(hash), nil
	}

	// 简单取模 - 安全转换：确保结果在int范围内
	/* #nosec G115 -- hash%uint32结果在uint32范围内，转换为int是安全的 */
	return int(hash % uint32(hs.shardCount)), nil
}

// buildHashKey 构建哈希键
func (hs *HashSharding) buildHashKey(key *ShardKey) []byte {
	// 使用设备ID和测点ID构建键
	data := fmt.Sprintf("%s:%s", key.DeviceID, key.PointID)
	return []byte(data)
}

// consistentHash 一致性哈希
func (hs *HashSharding) consistentHash(hash uint32) int {
	// 构建虚拟节点环
	type virtualNode struct {
		hash    uint32
		shardID int
	}

	nodes := make([]virtualNode, 0, hs.shardCount*hs.virtualNodes)
	for _, shard := range hs.shards {
		if shard.Status != ShardStatusActive {
			continue
		}
		for i := 0; i < hs.virtualNodes; i++ {
			nodeKey := fmt.Sprintf("%d:%d", shard.ID, i)
			nodeHash := hs.hashFunc([]byte(nodeKey))
			nodes = append(nodes, virtualNode{
				hash:    nodeHash,
				shardID: shard.ID,
			})
		}
	}

	if len(nodes) == 0 {
		return 0
	}

	// 排序
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].hash < nodes[j].hash
	})

	// 查找第一个大于等于hash的节点
	idx := sort.Search(len(nodes), func(i int) bool {
		return nodes[i].hash >= hash
	})

	if idx >= len(nodes) {
		idx = 0
	}

	return nodes[idx].shardID
}

// GetShards 获取所有分片
func (hs *HashSharding) GetShards() []*Shard {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	result := make([]*Shard, len(hs.shards))
	copy(result, hs.shards)
	return result
}

// AddShard 添加分片
func (hs *HashSharding) AddShard(shard *Shard) error {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	// 检查是否已存在
	for _, s := range hs.shards {
		if s.ID == shard.ID {
			return fmt.Errorf("shard %d already exists", shard.ID)
		}
	}

	hs.shards = append(hs.shards, shard)
	hs.shardCount++
	return nil
}

// RemoveShard 移除分片
func (hs *HashSharding) RemoveShard(shardID int) error {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	for i, s := range hs.shards {
		if s.ID == shardID {
			hs.shards = append(hs.shards[:i], hs.shards[i+1:]...)
			hs.shardCount--
			return nil
		}
	}

	return ErrShardNotFound
}

// GetShardCount 获取分片数量
func (hs *HashSharding) GetShardCount() int {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	return hs.shardCount
}

// GetType 获取分片类型
func (hs *HashSharding) GetType() string {
	return "hash"
}

// Rebalance 重新平衡分片
func (hs *HashSharding) Rebalance() error {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	// 哈希分片不需要手动重平衡
	// 虚拟节点会自动处理负载均衡
	return nil
}

// RangeSharding 范围分片策略
type RangeSharding struct {
	mu     sync.RWMutex
	shards []*Shard
	ranges []KeyRange
}

// KeyRange 键范围
type KeyRange struct {
	Start []byte
	End   []byte
}

// NewRangeSharding 创建范围分片策略
func NewRangeSharding() *RangeSharding {
	return &RangeSharding{
		shards: make([]*Shard, 0),
		ranges: make([]KeyRange, 0),
	}
}

// GetShard 根据分片键获取分片ID
func (rs *RangeSharding) GetShard(key *ShardKey) (int, error) {
	if key == nil {
		return -1, ErrInvalidShardKey
	}

	rs.mu.RLock()
	defer rs.mu.RUnlock()

	// 构建范围键
	rangeKey := rs.buildRangeKey(key)

	// 查找匹配的范围
	for i, shard := range rs.shards {
		if shard.Status != ShardStatusActive {
			continue
		}

		r := rs.ranges[i]
		if bytesCompare(rangeKey, r.Start) >= 0 && bytesCompare(rangeKey, r.End) < 0 {
			return shard.ID, nil
		}
	}

	return -1, ErrShardNotFound
}

// buildRangeKey 构建范围键
func (rs *RangeSharding) buildRangeKey(key *ShardKey) []byte {
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

// AddRangeShard 添加范围分片
func (rs *RangeSharding) AddRangeShard(shard *Shard, start, end []byte) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	// 检查范围是否重叠
	for i, r := range rs.ranges {
		if bytesCompare(end, r.Start) > 0 && bytesCompare(start, r.End) < 0 {
			return fmt.Errorf("range overlaps with shard %d", rs.shards[i].ID)
		}
	}

	rs.shards = append(rs.shards, shard)
	rs.ranges = append(rs.ranges, KeyRange{Start: start, End: end})
	return nil
}

// GetShards 获取所有分片
func (rs *RangeSharding) GetShards() []*Shard {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	result := make([]*Shard, len(rs.shards))
	copy(result, rs.shards)
	return result
}

// AddShard 添加分片
func (rs *RangeSharding) AddShard(shard *Shard) error {
	return fmt.Errorf("use AddRangeShard for range sharding")
}

// RemoveShard 移除分片
func (rs *RangeSharding) RemoveShard(shardID int) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	for i, s := range rs.shards {
		if s.ID == shardID {
			rs.shards = append(rs.shards[:i], rs.shards[i+1:]...)
			rs.ranges = append(rs.ranges[:i], rs.ranges[i+1:]...)
			return nil
		}
	}

	return ErrShardNotFound
}

// GetShardCount 获取分片数量
func (rs *RangeSharding) GetShardCount() int {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return len(rs.shards)
}

// GetType 获取分片类型
func (rs *RangeSharding) GetType() string {
	return "range"
}

// Rebalance 重新平衡分片
func (rs *RangeSharding) Rebalance() error {
	// 范围分片需要手动调整范围
	return nil
}

// TimeSharding 时间分片策略
type TimeSharding struct {
	mu           sync.RWMutex
	shards       []*Shard
	granularity  TimeGranularity
	timeRanges   []TimeRange
	autoCreate   bool
	maxShards    int
}

// TimeGranularity 时间粒度
type TimeGranularity int

const (
	TimeGranularityHour TimeGranularity = iota
	TimeGranularityDay
	TimeGranularityWeek
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
	case TimeGranularityWeek:
		return "week"
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

// NewTimeSharding 创建时间分片策略
func NewTimeSharding(granularity TimeGranularity, opts ...TimeOption) (*TimeSharding, error) {
	ts := &TimeSharding{
		granularity: granularity,
		autoCreate:  true,
		maxShards:   365, // 默认最大分片数
		shards:      make([]*Shard, 0),
		timeRanges:  make([]TimeRange, 0),
	}

	for _, opt := range opts {
		opt(ts)
	}

	return ts, nil
}

// TimeOption 时间分片选项
type TimeOption func(*TimeSharding)

// WithAutoCreate 设置自动创建分片
func WithAutoCreate(auto bool) TimeOption {
	return func(ts *TimeSharding) {
		ts.autoCreate = auto
	}
}

// WithMaxShards 设置最大分片数
func WithMaxShards(max int) TimeOption {
	return func(ts *TimeSharding) {
		ts.maxShards = max
	}
}

// GetShard 根据分片键获取分片ID
func (ts *TimeSharding) GetShard(key *ShardKey) (int, error) {
	if key == nil {
		return -1, ErrInvalidShardKey
	}

	ts.mu.Lock()
	defer ts.mu.Unlock()

	// 计算时间分片
	shardTime := ts.truncateTime(key.Timestamp)

	// 查找现有分片
	for i, shard := range ts.shards {
		if shard.Status != ShardStatusActive {
			continue
		}

		r := ts.timeRanges[i]
		if !shardTime.Before(r.Start) && shardTime.Before(r.End) {
			return shard.ID, nil
		}
	}

	// 自动创建分片
	if ts.autoCreate {
		return ts.createShardForTime(shardTime)
	}

	return -1, ErrShardNotFound
}

// truncateTime 截断时间到指定粒度
func (ts *TimeSharding) truncateTime(t time.Time) time.Time {
	switch ts.granularity {
	case TimeGranularityHour:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
	case TimeGranularityDay:
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	case TimeGranularityWeek:
		weekday := t.Weekday()
		if weekday == time.Sunday {
			weekday = 7
		}
		return time.Date(t.Year(), t.Month(), t.Day()-int(weekday)+1, 0, 0, 0, 0, t.Location())
	case TimeGranularityMonth:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	case TimeGranularityYear:
		return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
	default:
		return t
	}
}

// createShardForTime 为指定时间创建分片
func (ts *TimeSharding) createShardForTime(t time.Time) (int, error) {
	if len(ts.shards) >= ts.maxShards {
		return -1, fmt.Errorf("max shards limit reached: %d", ts.maxShards)
	}

	start := ts.truncateTime(t)
	end := ts.getNextTime(start)

	shardID := len(ts.shards)
	shard := &Shard{
		ID:        shardID,
		Name:      fmt.Sprintf("shard-%s", start.Format("20060102-150405")),
		StartTime: start,
		EndTime:   end,
		Weight:    1,
		Status:    ShardStatusActive,
		Metadata: map[string]interface{}{
			"type":      "time",
			"startTime": start,
			"endTime":   end,
		},
	}

	ts.shards = append(ts.shards, shard)
	ts.timeRanges = append(ts.timeRanges, TimeRange{Start: start, End: end})

	return shardID, nil
}

// getNextTime 获取下一个时间点
func (ts *TimeSharding) getNextTime(t time.Time) time.Time {
	switch ts.granularity {
	case TimeGranularityHour:
		return t.Add(time.Hour)
	case TimeGranularityDay:
		return t.AddDate(0, 0, 1)
	case TimeGranularityWeek:
		return t.AddDate(0, 0, 7)
	case TimeGranularityMonth:
		return t.AddDate(0, 1, 0)
	case TimeGranularityYear:
		return t.AddDate(1, 0, 0)
	default:
		return t
	}
}

// GetShards 获取所有分片
func (ts *TimeSharding) GetShards() []*Shard {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	result := make([]*Shard, len(ts.shards))
	copy(result, ts.shards)
	return result
}

// AddShard 添加分片
func (ts *TimeSharding) AddShard(shard *Shard) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// 检查时间范围是否有效
	if shard.StartTime.IsZero() || shard.EndTime.IsZero() {
		return fmt.Errorf("shard must have valid start and end time")
	}

	// 检查时间范围是否重叠
	for i, r := range ts.timeRanges {
		if shard.StartTime.Before(r.End) && shard.EndTime.After(r.Start) {
			return fmt.Errorf("time range overlaps with shard %d", ts.shards[i].ID)
		}
	}

	ts.shards = append(ts.shards, shard)
	ts.timeRanges = append(ts.timeRanges, TimeRange{
		Start: shard.StartTime,
		End:   shard.EndTime,
	})
	return nil
}

// RemoveShard 移除分片
func (ts *TimeSharding) RemoveShard(shardID int) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	for i, s := range ts.shards {
		if s.ID == shardID {
			ts.shards = append(ts.shards[:i], ts.shards[i+1:]...)
			ts.timeRanges = append(ts.timeRanges[:i], ts.timeRanges[i+1:]...)
			return nil
		}
	}

	return ErrShardNotFound
}

// GetShardCount 获取分片数量
func (ts *TimeSharding) GetShardCount() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return len(ts.shards)
}

// GetType 获取分片类型
func (ts *TimeSharding) GetType() string {
	return "time"
}

// Rebalance 重新平衡分片
func (ts *TimeSharding) Rebalance() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// 按时间排序分片
	sort.Slice(ts.shards, func(i, j int) bool {
		return ts.shards[i].StartTime.Before(ts.shards[j].StartTime)
	})

	// 重新生成时间范围
	ts.timeRanges = make([]TimeRange, len(ts.shards))
	for i, shard := range ts.shards {
		ts.timeRanges[i] = TimeRange{
			Start: shard.StartTime,
			End:   shard.EndTime,
		}
	}

	return nil
}

// GetShardsByTimeRange 根据时间范围获取分片
func (ts *TimeSharding) GetShardsByTimeRange(start, end time.Time) []*Shard {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	result := make([]*Shard, 0)
	for i, shard := range ts.shards {
		r := ts.timeRanges[i]
		if shard.Status == ShardStatusActive &&
			!r.End.Before(start) && !r.Start.After(end) {
			result = append(result, shard)
		}
	}
	return result
}

// ShardRouter 分片路由器
type ShardRouter struct {
	mu       sync.RWMutex
	strategy ShardingStrategy
	cache    *shardCache
	stats    *RouterStats
}

// shardCache 分片缓存
type shardCache struct {
	mu    sync.RWMutex
	items map[string]int
	max   int
}

// newShardCache 创建分片缓存
func newShardCache(max int) *shardCache {
	return &shardCache{
		items: make(map[string]int),
		max:   max,
	}
}

// Get 获取缓存的分片ID
func (c *shardCache) Get(key string) (int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	shardID, ok := c.items[key]
	return shardID, ok
}

// Set 设置缓存的分片ID
func (c *shardCache) Set(key string, shardID int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.items) >= c.max {
		// 简单的LRU：清空一半
		count := 0
		for k := range c.items {
			delete(c.items, k)
			count++
			if count >= c.max/2 {
				break
			}
		}
	}

	c.items[key] = shardID
}

// RouterStats 路由统计
type RouterStats struct {
	mu           sync.RWMutex
	totalRoutes  int64
	cacheHits    int64
	cacheMisses  int64
	routeErrors  int64
	shardRoutes  map[int]int64
}

// NewRouterStats 创建路由统计
func NewRouterStats() *RouterStats {
	return &RouterStats{
		shardRoutes: make(map[int]int64),
	}
}

// RecordRoute 记录路由
func (s *RouterStats) RecordRoute(shardID int, cacheHit bool, err bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalRoutes++
	if cacheHit {
		s.cacheHits++
	} else {
		s.cacheMisses++
	}
	if err {
		s.routeErrors++
	}
	s.shardRoutes[shardID]++
}

// GetStats 获取统计信息
func (s *RouterStats) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"totalRoutes":  s.totalRoutes,
		"cacheHits":    s.cacheHits,
		"cacheMisses":  s.cacheMisses,
		"routeErrors":  s.routeErrors,
		"hitRate":      float64(s.cacheHits) / float64(s.totalRoutes+1),
		"shardRoutes":  s.shardRoutes,
	}
}

// NewShardRouter 创建分片路由器
func NewShardRouter(strategy ShardingStrategy, cacheSize int) *ShardRouter {
	return &ShardRouter{
		strategy: strategy,
		cache:    newShardCache(cacheSize),
		stats:    NewRouterStats(),
	}
}

// Route 路由到分片
func (r *ShardRouter) Route(key *ShardKey) (int, error) {
	if key == nil {
		return -1, ErrInvalidShardKey
	}

	// 构建缓存键
	cacheKey := r.buildCacheKey(key)

	// 尝试从缓存获取
	if shardID, ok := r.cache.Get(cacheKey); ok {
		r.stats.RecordRoute(shardID, true, false)
		return shardID, nil
	}

	// 通过策略获取分片
	shardID, err := r.strategy.GetShard(key)
	if err != nil {
		r.stats.RecordRoute(-1, false, true)
		return -1, err
	}

	// 缓存结果
	r.cache.Set(cacheKey, shardID)
	r.stats.RecordRoute(shardID, false, false)

	return shardID, nil
}

// buildCacheKey 构建缓存键
func (r *ShardRouter) buildCacheKey(key *ShardKey) string {
	return fmt.Sprintf("%s:%s:%d", key.DeviceID, key.PointID, key.Timestamp.Unix())
}

// RouteBatch 批量路由
func (r *ShardRouter) RouteBatch(keys []*ShardKey) (map[int][]*ShardKey, error) {
	result := make(map[int][]*ShardKey)
	var lastErr error

	for _, key := range keys {
		shardID, err := r.Route(key)
		if err != nil {
			lastErr = err
			continue
		}
		result[shardID] = append(result[shardID], key)
	}

	return result, lastErr
}

// GetStats 获取路由统计
func (r *ShardRouter) GetStats() map[string]interface{} {
	return r.stats.GetStats()
}

// Rebalance 执行重平衡
func (r *ShardRouter) Rebalance() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 清空缓存
	r.cache = newShardCache(r.cache.max)

	return r.strategy.Rebalance()
}

// Rebalancer 分片重平衡器
type Rebalancer struct {
	mu            sync.Mutex
	strategy      ShardingStrategy
	threshold     float64 // 不平衡阈值
	inProgress    bool
	onMigrate     func(shardID int, fromNode, toNode string) error
	onComplete    func(stats *RebalanceStats)
}

// RebalanceStats 重平衡统计
type RebalanceStats struct {
	StartTime        time.Time
	EndTime          time.Time
	ShardsMigrated   int
	DataTransferred  int64
	Errors           []error
	SourceNode       string
	TargetNode       string
}

// NewRebalancer 创建重平衡器
func NewRebalancer(strategy ShardingStrategy, threshold float64) *Rebalancer {
	return &Rebalancer{
		strategy:  strategy,
		threshold: threshold,
	}
}

// SetMigrateHandler 设置迁移处理函数
func (rb *Rebalancer) SetMigrateHandler(fn func(shardID int, fromNode, toNode string) error) {
	rb.onMigrate = fn
}

// SetCompleteHandler 设置完成处理函数
func (rb *Rebalancer) SetCompleteHandler(fn func(stats *RebalanceStats)) {
	rb.onComplete = fn
}

// Analyze 分析分片平衡状态
func (rb *Rebalancer) Analyze() (*BalanceAnalysis, error) {
	shards := rb.strategy.GetShards()
	if len(shards) == 0 {
		return nil, ErrInvalidShardCount
	}

	// 计算每个分片的负载
	loads := make([]float64, len(shards))
	totalLoad := 0.0
	for i, shard := range shards {
		// 这里简化处理，实际应该从监控获取真实负载
		loads[i] = float64(shard.Weight)
		totalLoad += loads[i]
	}

	avgLoad := totalLoad / float64(len(shards))
	maxLoad := 0.0
	minLoad := math.MaxFloat64

	for _, load := range loads {
		if load > maxLoad {
			maxLoad = load
		}
		if load < minLoad {
			minLoad = load
		}
	}

	imbalance := 0.0
	if avgLoad > 0 {
		imbalance = (maxLoad - minLoad) / avgLoad
	}

	return &BalanceAnalysis{
		TotalShards:   len(shards),
		AverageLoad:   avgLoad,
		MaxLoad:       maxLoad,
		MinLoad:       minLoad,
		Imbalance:     imbalance,
		NeedsRebalance: imbalance > rb.threshold,
	}, nil
}

// BalanceAnalysis 平衡分析结果
type BalanceAnalysis struct {
	TotalShards    int
	AverageLoad    float64
	MaxLoad        float64
	MinLoad        float64
	Imbalance      float64
	NeedsRebalance bool
}

// Execute 执行重平衡
func (rb *Rebalancer) Execute() error {
	rb.mu.Lock()
	if rb.inProgress {
		rb.mu.Unlock()
		return ErrRebalanceInProgress
	}
	rb.inProgress = true
	rb.mu.Unlock()

	defer func() {
		rb.mu.Lock()
		rb.inProgress = false
		rb.mu.Unlock()
	}()

	stats := &RebalanceStats{
		StartTime: time.Now(),
	}

	// 分析当前状态
	analysis, err := rb.Analyze()
	if err != nil {
		stats.Errors = append(stats.Errors, err)
		if rb.onComplete != nil {
			rb.onComplete(stats)
		}
		return err
	}

	// 如果不需要重平衡，直接返回
	if !analysis.NeedsRebalance {
		stats.EndTime = time.Now()
		if rb.onComplete != nil {
			rb.onComplete(stats)
		}
		return nil
	}

	// 执行策略的重平衡
	err = rb.strategy.Rebalance()
	if err != nil {
		stats.Errors = append(stats.Errors, err)
	}

	stats.EndTime = time.Now()
	if rb.onComplete != nil {
		rb.onComplete(stats)
	}

	return err
}

// IsInProgress 检查是否正在重平衡
func (rb *Rebalancer) IsInProgress() bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.inProgress
}

// ShardManager 分片管理器
type ShardManager struct {
	mu        sync.RWMutex
	strategies map[string]ShardingStrategy
	router    *ShardRouter
	rebalancer *Rebalancer
}

// NewShardManager 创建分片管理器
func NewShardManager(defaultStrategy ShardingStrategy, cacheSize int) *ShardManager {
	sm := &ShardManager{
		strategies: make(map[string]ShardingStrategy),
	}
	sm.strategies["default"] = defaultStrategy
	sm.router = NewShardRouter(defaultStrategy, cacheSize)
	sm.rebalancer = NewRebalancer(defaultStrategy, 0.3)
	return sm
}

// RegisterStrategy 注册分片策略
func (sm *ShardManager) RegisterStrategy(name string, strategy ShardingStrategy) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.strategies[name] = strategy
}

// GetStrategy 获取分片策略
func (sm *ShardManager) GetStrategy(name string) (ShardingStrategy, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	s, ok := sm.strategies[name]
	return s, ok
}

// Route 路由到分片
func (sm *ShardManager) Route(key *ShardKey) (int, error) {
	return sm.router.Route(key)
}

// RouteWithStrategy 使用指定策略路由
func (sm *ShardManager) RouteWithStrategy(strategyName string, key *ShardKey) (int, error) {
	strategy, ok := sm.GetStrategy(strategyName)
	if !ok {
		return -1, fmt.Errorf("strategy %s not found", strategyName)
	}
	return strategy.GetShard(key)
}

// GetRouterStats 获取路由统计
func (sm *ShardManager) GetRouterStats() map[string]interface{} {
	return sm.router.GetStats()
}

// Rebalance 执行重平衡
func (sm *ShardManager) Rebalance() error {
	return sm.rebalancer.Execute()
}

// GetRebalanceStatus 获取重平衡状态
func (sm *ShardManager) GetRebalanceStatus() (bool, *BalanceAnalysis, error) {
	inProgress := sm.rebalancer.IsInProgress()
	analysis, err := sm.rebalancer.Analyze()
	return inProgress, analysis, err
}

// AddShard 添加分片
func (sm *ShardManager) AddShard(shard *Shard) error {
	return sm.strategies["default"].AddShard(shard)
}

// RemoveShard 移除分片
func (sm *ShardManager) RemoveShard(shardID int) error {
	return sm.strategies["default"].RemoveShard(shardID)
}

// GetShards 获取所有分片
func (sm *ShardManager) GetShards() []*Shard {
	return sm.strategies["default"].GetShards()
}

// GetShardCount 获取分片数量
func (sm *ShardManager) GetShardCount() int {
	return sm.strategies["default"].GetShardCount()
}
