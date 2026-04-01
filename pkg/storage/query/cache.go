package query

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
)

// CacheStatus 缓存状态
type CacheStatus string

const (
	CacheStatusHit   CacheStatus = "hit"
	CacheStatusMiss  CacheStatus = "miss"
	CacheStatusExpired CacheStatus = "expired"
	CacheStatusEvicted CacheStatus = "evicted"
)

// CacheEntry 缓存条目
type CacheEntry struct {
	Key         string                 `json:"key"`
	QueryHash   string                 `json:"query_hash"`
	Result      *QueryResult           `json:"result"`
	CreatedAt   time.Time              `json:"created_at"`
	ExpiresAt   time.Time              `json:"expires_at"`
	AccessCount int64                  `json:"access_count"`
	LastAccess  time.Time              `json:"last_access"`
	Size        int64                  `json:"size"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// IsExpired 检查是否过期
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Enabled          bool          `json:"enabled"`
	DefaultTTL       time.Duration `json:"default_ttl"`
	MaxSize          int64         `json:"max_size"`        // 最大缓存大小（字节）
	MaxEntries       int           `json:"max_entries"`     // 最大条目数
	EvictionPolicy   string        `json:"eviction_policy"` // LRU, LFU, FIFO
	KeyPrefix        string        `json:"key_prefix"`
	EnableStats      bool          `json:"enable_stats"`
	WarmupOnStart    bool          `json:"warmup_on_start"`
	CompressionLevel int           `json:"compression_level"`
}

// DefaultCacheConfig 默认缓存配置
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		Enabled:          true,
		DefaultTTL:       5 * time.Minute,
		MaxSize:          100 * 1024 * 1024, // 100MB
		MaxEntries:       10000,
		EvictionPolicy:   "LRU",
		KeyPrefix:        "nem:query:cache:",
		EnableStats:      true,
		WarmupOnStart:    false,
		CompressionLevel: 0,
	}
}

// QueryCache 查询缓存
type QueryCache struct {
	redis       *redis.Client
	config      CacheConfig
	stats       *CacheStats
	entries     map[string]*CacheEntry
	lruList     *LRUList
	tagIndex    map[string]map[string]struct{}
	mu          sync.RWMutex
	warmupTasks []*WarmupTask
}

// CacheStats 缓存统计
type CacheStats struct {
	TotalRequests  int64         `json:"total_requests"`
	Hits           int64         `json:"hits"`
	Misses         int64         `json:"misses"`
	Evictions      int64         `json:"evictions"`
	ExpiredCount   int64         `json:"expired_count"`
	TotalSize      int64         `json:"total_size"`
	EntryCount     int64         `json:"entry_count"`
	AvgAccessTime  time.Duration `json:"avg_access_time"`
	HitRate        float64       `json:"hit_rate"`
	MemoryUsage    int64         `json:"memory_usage"`
	LastReset      time.Time     `json:"last_reset"`
}

// NewQueryCache 创建查询缓存
func NewQueryCache(redisClient *redis.Client, config CacheConfig) *QueryCache {
	if config.KeyPrefix == "" {
		config.KeyPrefix = "nem:query:cache:"
	}

	cache := &QueryCache{
		redis:    redisClient,
		config:   config,
		stats:    &CacheStats{LastReset: time.Now()},
		entries:  make(map[string]*CacheEntry),
		lruList:  NewLRUList(),
		tagIndex: make(map[string]map[string]struct{}),
	}

	// 启动后台清理任务
	go cache.backgroundCleanup()

	return cache
}

// Get 获取缓存
func (c *QueryCache) Get(ctx context.Context, req *QueryRequest) (*QueryResult, CacheStatus, error) {
	if !c.config.Enabled {
		return nil, CacheStatusMiss, nil
	}

	key := c.GenerateKey(req)

	// 先检查本地缓存
	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()

	if exists {
		if entry.IsExpired() {
			c.deleteEntry(key)
			atomic.AddInt64(&c.stats.ExpiredCount, 1)
			return nil, CacheStatusExpired, nil
		}

		// 更新访问统计
		atomic.AddInt64(&entry.AccessCount, 1)
		entry.LastAccess = time.Now()
		c.lruList.MoveToFront(key)

		atomic.AddInt64(&c.stats.Hits, 1)
		atomic.AddInt64(&c.stats.TotalRequests, 1)
		c.updateHitRate()

		entry.Result.Cached = true
		return entry.Result, CacheStatusHit, nil
	}

	// 检查Redis缓存
	if c.redis != nil {
		result, err := c.getFromRedis(ctx, key)
		if err == nil && result != nil {
			// 回填本地缓存
			c.Set(ctx, req, result)

			atomic.AddInt64(&c.stats.Hits, 1)
			atomic.AddInt64(&c.stats.TotalRequests, 1)
			c.updateHitRate()

			result.Cached = true
			return result, CacheStatusHit, nil
		}
	}

	atomic.AddInt64(&c.stats.Misses, 1)
	atomic.AddInt64(&c.stats.TotalRequests, 1)
	c.updateHitRate()

	return nil, CacheStatusMiss, nil
}

// Set 设置缓存
func (c *QueryCache) Set(ctx context.Context, req *QueryRequest, result *QueryResult) error {
	if !c.config.Enabled {
		return nil
	}

	key := c.GenerateKey(req)
	ttl := c.config.DefaultTTL

	// 从请求选项获取自定义TTL
	if req.Options != nil {
		if ttlVal, ok := req.Options["cache_ttl"]; ok {
			if ttlDuration, ok := ttlVal.(time.Duration); ok {
				ttl = ttlDuration
			}
		}
	}

	// 计算结果大小
	size := c.calculateSize(result)

	// 检查是否需要淘汰
	if err := c.ensureCapacity(size); err != nil {
		return err
	}

	// 创建缓存条目
	entry := &CacheEntry{
		Key:        key,
		QueryHash:  c.GenerateHash(req),
		Result:     result,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(ttl),
		AccessCount: 0,
		LastAccess:  time.Now(),
		Size:       size,
		Tags:       c.extractTags(req),
	}

	// 存储到本地缓存
	c.mu.Lock()
	c.entries[key] = entry
	c.lruList.PushFront(key)
	c.updateTagIndex(key, entry.Tags)
	c.stats.TotalSize += size
	c.stats.EntryCount = int64(len(c.entries))
	c.mu.Unlock()

	// 存储到Redis
	if c.redis != nil {
		if err := c.setToRedis(ctx, key, entry, ttl); err != nil {
			// Redis存储失败不影响本地缓存
			fmt.Printf("warning: failed to set redis cache: %v\n", err)
		}
	}

	return nil
}

// Delete 删除缓存
func (c *QueryCache) Delete(ctx context.Context, req *QueryRequest) error {
	key := c.GenerateKey(req)
	return c.deleteEntry(key)
}

// DeleteByTags 按标签删除缓存
func (c *QueryCache) DeleteByTags(ctx context.Context, tags []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, tag := range tags {
		keys, exists := c.tagIndex[tag]
		if !exists {
			continue
		}

		for key := range keys {
			if entry, ok := c.entries[key]; ok {
				c.stats.TotalSize -= entry.Size
				delete(c.entries, key)
				c.lruList.Remove(key)
				atomic.AddInt64(&c.stats.Evictions, 1)
			}
		}

		delete(c.tagIndex, tag)
	}

	c.stats.EntryCount = int64(len(c.entries))
	return nil
}

// Clear 清空缓存
func (c *QueryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
	c.lruList = NewLRUList()
	c.tagIndex = make(map[string]map[string]struct{})
	c.stats.TotalSize = 0
	c.stats.EntryCount = 0

	if c.redis != nil {
		pattern := c.config.KeyPrefix + "*"
		iter := c.redis.Scan(ctx, 0, pattern, 0).Iterator()
		for iter.Next(ctx) {
			c.redis.Del(ctx, iter.Val())
		}
	}

	return nil
}

// GenerateKey 生成缓存键
func (c *QueryCache) GenerateKey(req *QueryRequest) string {
	hash := c.GenerateHash(req)
	return c.config.KeyPrefix + hash
}

// GenerateHash 生成查询哈希
func (c *QueryCache) GenerateHash(req *QueryRequest) string {
	// 构建规范化查询字符串
	normalized := struct {
		Type       string
		Database   string
		Table      string
		Fields     []string
		Conditions []QueryCondition
		OrderBy    []OrderByField
		GroupBy    []string
		Limit      int
		Offset     int
		TimeRange  *TimeRange
		Aggregates []AggregateField
	}{
		Type:       string(req.Type),
		Database:   req.Database,
		Table:      req.Table,
		Fields:     req.Fields,
		Conditions: req.Conditions,
		OrderBy:    req.OrderBy,
		GroupBy:    req.GroupBy,
		Limit:      req.Limit,
		Offset:     req.Offset,
		TimeRange:  req.TimeRange,
		Aggregates: req.Aggregates,
	}

	data, _ := json.Marshal(normalized)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// GetStats 获取统计信息
func (c *QueryCache) GetStats() *CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := &CacheStats{
		TotalRequests: atomic.LoadInt64(&c.stats.TotalRequests),
		Hits:          atomic.LoadInt64(&c.stats.Hits),
		Misses:        atomic.LoadInt64(&c.stats.Misses),
		Evictions:     atomic.LoadInt64(&c.stats.Evictions),
		ExpiredCount:  atomic.LoadInt64(&c.stats.ExpiredCount),
		TotalSize:     c.stats.TotalSize,
		EntryCount:    c.stats.EntryCount,
		HitRate:       c.stats.HitRate,
		MemoryUsage:   c.stats.TotalSize,
		LastReset:     c.stats.LastReset,
	}

	return stats
}

// ResetStats 重置统计
func (c *QueryCache) ResetStats() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.stats = &CacheStats{LastReset: time.Now()}
}

// Warmup 预热缓存
func (c *QueryCache) Warmup(ctx context.Context, tasks []*WarmupTask) error {
	c.warmupTasks = tasks

	for _, task := range tasks {
		if err := c.executeWarmupTask(ctx, task); err != nil {
			fmt.Printf("warmup task %s failed: %v\n", task.Name, err)
		}
	}

	return nil
}

// WarmupTask 预热任务
type WarmupTask struct {
	Name        string        `json:"name"`
	Query       *QueryRequest `json:"query"`
	TTL         time.Duration `json:"ttl"`
	RefreshInterval time.Duration `json:"refresh_interval"`
	Enabled     bool          `json:"enabled"`
}

// executeWarmupTask 执行预热任务
func (c *QueryCache) executeWarmupTask(ctx context.Context, task *WarmupTask) error {
	if !task.Enabled {
		return nil
	}

	// 这里应该执行查询并缓存结果
	// 简化实现，实际需要注入查询执行器
	return nil
}

// ensureCapacity 确保容量
func (c *QueryCache) ensureCapacity(newSize int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查条目数限制
	for len(c.entries) >= c.config.MaxEntries {
		if err := c.evictOne(); err != nil {
			return err
		}
	}

	// 检查大小限制
	for c.stats.TotalSize+newSize > c.config.MaxSize {
		if err := c.evictOne(); err != nil {
			return err
		}
	}

	return nil
}

// evictOne 淘汰一个条目
func (c *QueryCache) evictOne() error {
	key, ok := c.lruList.Back()
	if !ok {
		return errors.New("no entry to evict")
	}

	entry, exists := c.entries[key]
	if !exists {
		c.lruList.Remove(key)
		return nil
	}

	c.stats.TotalSize -= entry.Size
	delete(c.entries, key)
	c.lruList.Remove(key)
	c.removeFromTagIndex(key, entry.Tags)
	atomic.AddInt64(&c.stats.Evictions, 1)

	return nil
}

// deleteEntry 删除条目
func (c *QueryCache) deleteEntry(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil
	}

	c.stats.TotalSize -= entry.Size
	delete(c.entries, key)
	c.lruList.Remove(key)
	c.removeFromTagIndex(key, entry.Tags)
	c.stats.EntryCount = int64(len(c.entries))

	if c.redis != nil {
		c.redis.Del(context.Background(), key)
	}

	return nil
}

// updateTagIndex 更新标签索引
func (c *QueryCache) updateTagIndex(key string, tags []string) {
	for _, tag := range tags {
		if c.tagIndex[tag] == nil {
			c.tagIndex[tag] = make(map[string]struct{})
		}
		c.tagIndex[tag][key] = struct{}{}
	}
}

// removeFromTagIndex 从标签索引移除
func (c *QueryCache) removeFromTagIndex(key string, tags []string) {
	for _, tag := range tags {
		if keys, exists := c.tagIndex[tag]; exists {
			delete(keys, key)
			if len(keys) == 0 {
				delete(c.tagIndex, tag)
			}
		}
	}
}

// extractTags 提取标签
func (c *QueryCache) extractTags(req *QueryRequest) []string {
	tags := make([]string, 0)

	if req.Database != "" {
		tags = append(tags, "db:"+req.Database)
	}
	if req.Table != "" {
		tags = append(tags, "table:"+req.Table)
	}
	if req.Type != "" {
		tags = append(tags, "type:"+string(req.Type))
	}

	// 从条件中提取标签
	for _, cond := range req.Conditions {
		if cond.Field != "" {
			tags = append(tags, "field:"+cond.Field)
		}
	}

	return tags
}

// calculateSize 计算结果大小
func (c *QueryCache) calculateSize(result *QueryResult) int64 {
	data, err := json.Marshal(result)
	if err != nil {
		return 0
	}
	return int64(len(data))
}

// updateHitRate 更新命中率
func (c *QueryCache) updateHitRate() {
	total := atomic.LoadInt64(&c.stats.TotalRequests)
	if total > 0 {
		hits := atomic.LoadInt64(&c.stats.Hits)
		c.stats.HitRate = float64(hits) / float64(total)
	}
}

// getFromRedis 从Redis获取缓存
func (c *QueryCache) getFromRedis(ctx context.Context, key string) (*QueryResult, error) {
	data, err := c.redis.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	if entry.IsExpired() {
		c.redis.Del(ctx, key)
		return nil, nil
	}

	return entry.Result, nil
}

// setToRedis 设置Redis缓存
func (c *QueryCache) setToRedis(ctx context.Context, key string, entry *CacheEntry, ttl time.Duration) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return c.redis.Set(ctx, key, data, ttl).Err()
}

// backgroundCleanup 后台清理
func (c *QueryCache) backgroundCleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanupExpired()
	}
}

// cleanupExpired 清理过期条目
func (c *QueryCache) cleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, entry := range c.entries {
		if entry.IsExpired() {
			c.stats.TotalSize -= entry.Size
			delete(c.entries, key)
			c.lruList.Remove(key)
			c.removeFromTagIndex(key, entry.Tags)
			atomic.AddInt64(&c.stats.ExpiredCount, 1)
		}
	}

	c.stats.EntryCount = int64(len(c.entries))
}

// LRUList LRU列表
type LRUList struct {
	head *lruNode
	tail *lruNode
	keys map[string]*lruNode
	mu   sync.Mutex
}

type lruNode struct {
	key  string
	prev *lruNode
	next *lruNode
}

// NewLRUList 创建LRU列表
func NewLRUList() *LRUList {
	return &LRUList{
		keys: make(map[string]*lruNode),
	}
}

// PushFront 添加到头部
func (l *LRUList) PushFront(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, exists := l.keys[key]; exists {
		return
	}

	node := &lruNode{key: key}
	l.keys[key] = node

	if l.head == nil {
		l.head = node
		l.tail = node
		return
	}

	node.next = l.head
	l.head.prev = node
	l.head = node
}

// MoveToFront 移动到头部
func (l *LRUList) MoveToFront(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	node, exists := l.keys[key]
	if !exists {
		return
	}

	l.removeNode(node)
	l.addToFront(node)
}

// Back 获取尾部元素
func (l *LRUList) Back() (string, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.tail == nil {
		return "", false
	}
	return l.tail.key, true
}

// Remove 移除元素
func (l *LRUList) Remove(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	node, exists := l.keys[key]
	if !exists {
		return
	}

	l.removeNode(node)
	delete(l.keys, key)
}

// removeNode 移除节点
func (l *LRUList) removeNode(node *lruNode) {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		l.head = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		l.tail = node.prev
	}
}

// addToFront 添加到头部
func (l *LRUList) addToFront(node *lruNode) {
	node.prev = nil
	node.next = l.head

	if l.head != nil {
		l.head.prev = node
	}
	l.head = node

	if l.tail == nil {
		l.tail = node
	}
}

// CacheInvalidator 缓存失效器
type CacheInvalidator struct {
	cache   *QueryCache
	rules   []InvalidationRule
	notify  chan InvalidationEvent
	mu      sync.RWMutex
}

// InvalidationRule 失效规则
type InvalidationRule struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Trigger    string        `json:"trigger"` // time, event, dependency
	Interval   time.Duration `json:"interval"`
	Tables     []string      `json:"tables"`
	Tags       []string      `json:"tags"`
	Enabled    bool          `json:"enabled"`
	LastRun    time.Time     `json:"last_run"`
}

// InvalidationEvent 失效事件
type InvalidationEvent struct {
	Type      string    `json:"type"`
	Table     string    `json:"table"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

// NewCacheInvalidator 创建缓存失效器
func NewCacheInvalidator(cache *QueryCache) *CacheInvalidator {
	invalidator := &CacheInvalidator{
		cache:  cache,
		rules:  make([]InvalidationRule, 0),
		notify: make(chan InvalidationEvent, 100),
	}

	go invalidator.processEvents()

	return invalidator
}

// AddRule 添加失效规则
func (i *CacheInvalidator) AddRule(rule InvalidationRule) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.rules = append(i.rules, rule)

	// 启动定时失效
	if rule.Trigger == "time" && rule.Interval > 0 {
		go i.timeBasedInvalidation(rule)
	}
}

// RemoveRule 移除失效规则
func (i *CacheInvalidator) RemoveRule(ruleID string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	for idx, rule := range i.rules {
		if rule.ID == ruleID {
			i.rules = append(i.rules[:idx], i.rules[idx+1:]...)
			break
		}
	}
}

// Notify 通知失效事件
func (i *CacheInvalidator) Notify(event InvalidationEvent) {
	select {
	case i.notify <- event:
	default:
		// 通道满，丢弃事件
	}
}

// processEvents 处理失效事件
func (i *CacheInvalidator) processEvents() {
	for event := range i.notify {
		i.handleEvent(event)
	}
}

// handleEvent 处理事件
func (i *CacheInvalidator) handleEvent(event InvalidationEvent) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	for _, rule := range i.rules {
		if !rule.Enabled {
			continue
		}

		if rule.Trigger == "event" {
			// 检查表匹配
			for _, table := range rule.Tables {
				if table == event.Table || table == "*" {
					i.invalidateByRule(rule)
					break
				}
			}
		}
	}
}

// timeBasedInvalidation 定时失效
func (i *CacheInvalidator) timeBasedInvalidation(rule InvalidationRule) {
	ticker := time.NewTicker(rule.Interval)
	defer ticker.Stop()

	for range ticker.C {
		i.mu.RLock()
		if !rule.Enabled {
			i.mu.RUnlock()
			return
		}
		i.invalidateByRule(rule)
		i.mu.RUnlock()
	}
}

// invalidateByRule 按规则失效
func (i *CacheInvalidator) invalidateByRule(rule InvalidationRule) {
	if len(rule.Tags) > 0 {
		i.cache.DeleteByTags(context.Background(), rule.Tags)
	}
}

// CacheWarmer 缓存预热器
type CacheWarmer struct {
	cache    *QueryCache
	executor *QueryExecutor
	tasks    []*WarmupTask
	running  bool
	mu       sync.RWMutex
}

// NewCacheWarmer 创建缓存预热器
func NewCacheWarmer(cache *QueryCache, executor *QueryExecutor) *CacheWarmer {
	return &CacheWarmer{
		cache:    cache,
		executor: executor,
		tasks:    make([]*WarmupTask, 0),
	}
}

// AddTask 添加预热任务
func (w *CacheWarmer) AddTask(task *WarmupTask) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.tasks = append(w.tasks, task)
}

// Start 启动预热
func (w *CacheWarmer) Start(ctx context.Context) {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}
	w.running = true
	w.mu.Unlock()

	// 执行所有预热任务
	for _, task := range w.tasks {
		if !task.Enabled {
			continue
		}

		result, err := w.executor.Execute(ctx, task.Query)
		if err != nil {
			fmt.Printf("warmup task %s failed: %v\n", task.Name, err)
			continue
		}

		if task.TTL > 0 {
			if task.Query.Options == nil {
				task.Query.Options = make(map[string]interface{})
			}
			task.Query.Options["cache_ttl"] = task.TTL
		}

		w.cache.Set(ctx, task.Query, result)

		// 设置定时刷新
		if task.RefreshInterval > 0 {
			go w.scheduleRefresh(ctx, task)
		}
	}
}

// scheduleRefresh 定时刷新
func (w *CacheWarmer) scheduleRefresh(ctx context.Context, task *WarmupTask) {
	ticker := time.NewTicker(task.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			result, err := w.executor.Execute(ctx, task.Query)
			if err != nil {
				continue
			}
			w.cache.Set(ctx, task.Query, result)
		}
	}
}

// Stop 停止预热
func (w *CacheWarmer) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.running = false
}

// CacheKeyBuilder 缓存键构建器
type CacheKeyBuilder struct {
	parts []string
}

// NewCacheKeyBuilder 创建缓存键构建器
func NewCacheKeyBuilder() *CacheKeyBuilder {
	return &CacheKeyBuilder{
		parts: make([]string, 0),
	}
}

// Add 添加部分
func (b *CacheKeyBuilder) Add(part string) *CacheKeyBuilder {
	b.parts = append(b.parts, part)
	return b
}

// AddInt 添加整数
func (b *CacheKeyBuilder) AddInt(part int) *CacheKeyBuilder {
	b.parts = append(b.parts, fmt.Sprintf("%d", part))
	return b
}

// AddInt64 添加int64
func (b *CacheKeyBuilder) AddInt64(part int64) *CacheKeyBuilder {
	b.parts = append(b.parts, fmt.Sprintf("%d", part))
	return b
}

// AddTime 添加时间
func (b *CacheKeyBuilder) AddTime(part time.Time) *CacheKeyBuilder {
	b.parts = append(b.parts, part.Format(time.RFC3339))
	return b
}

// Build 构建键
func (b *CacheKeyBuilder) Build() string {
	data := []byte{}
	for _, part := range b.parts {
		data = append(data, []byte(part)...)
		data = append(data, ':')
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:16])
}

// MultiLevelCache 多级缓存
type MultiLevelCache struct {
	levels []*CacheLevel
	config MultiLevelCacheConfig
}

// CacheLevel 缓存层级
type CacheLevel struct {
	Name     string
	Cache    *QueryCache
	Priority int
}

// MultiLevelCacheConfig 多级缓存配置
type MultiLevelCacheConfig struct {
	Levels []CacheLevelConfig `json:"levels"`
}

// CacheLevelConfig 缓存层级配置
type CacheLevelConfig struct {
	Name     string        `json:"name"`
	Priority int           `json:"priority"`
	Config   CacheConfig   `json:"config"`
}

// NewMultiLevelCache 创建多级缓存
func NewMultiLevelCache(redis *redis.Client, config MultiLevelCacheConfig) *MultiLevelCache {
	mlc := &MultiLevelCache{
		levels: make([]*CacheLevel, 0),
		config: config,
	}

	for _, levelCfg := range config.Levels {
		level := &CacheLevel{
			Name:     levelCfg.Name,
			Cache:    NewQueryCache(redis, levelCfg.Config),
			Priority: levelCfg.Priority,
		}
		mlc.levels = append(mlc.levels, level)
	}

	return mlc
}

// Get 获取缓存（从高优先级到低优先级）
func (m *MultiLevelCache) Get(ctx context.Context, req *QueryRequest) (*QueryResult, CacheStatus, error) {
	for _, level := range m.levels {
		result, status, err := level.Cache.Get(ctx, req)
		if err == nil && result != nil {
			return result, status, nil
		}
	}
	return nil, CacheStatusMiss, nil
}

// Set 设置缓存（设置所有层级）
func (m *MultiLevelCache) Set(ctx context.Context, req *QueryRequest, result *QueryResult) error {
	for _, level := range m.levels {
		if err := level.Cache.Set(ctx, req, result); err != nil {
			fmt.Printf("failed to set cache level %s: %v\n", level.Name, err)
		}
	}
	return nil
}

// Delete 删除缓存（删除所有层级）
func (m *MultiLevelCache) Delete(ctx context.Context, req *QueryRequest) error {
	for _, level := range m.levels {
		level.Cache.Delete(ctx, req)
	}
	return nil
}
