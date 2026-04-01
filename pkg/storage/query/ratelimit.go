package query

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// RateLimitStatus 限流状态
type RateLimitStatus string

const (
	RateLimitStatusAllowed RateLimitStatus = "allowed"
	RateLimitStatusLimited RateLimitStatus = "limited"
	RateLimitStatusWaiting RateLimitStatus = "waiting"
)

// RateLimitResult 限流结果
type RateLimitResult struct {
	Status       RateLimitStatus `json:"status"`
	Allowed      bool            `json:"allowed"`
	RetryAfter   time.Duration   `json:"retry_after"`
	Limit        int             `json:"limit"`
	Remaining    int             `json:"remaining"`
	ResetAt      time.Time       `json:"reset_at"`
	Key          string          `json:"key"`
	Priority     QueryPriority   `json:"priority"`
	WaitPosition int             `json:"wait_position"`
}

// RateLimiterConfig 限流器配置
type RateLimiterConfig struct {
	Enabled           bool          `json:"enabled"`
	Algorithm         string        `json:"algorithm"` // token_bucket, sliding_window, leaky_bucket
	DefaultRate       int           `json:"default_rate"`        // 每秒请求数
	DefaultBurst      int           `json:"default_burst"`       // 突发容量
	MaxWaitTime       time.Duration `json:"max_wait_time"`       // 最大等待时间
	PriorityQueues    int           `json:"priority_queues"`     // 优先级队列数
	EnableAdaptive    bool          `json:"enable_adaptive"`     // 自适应限流
	CleanupInterval   time.Duration `json:"cleanup_interval"`    // 清理间隔
	KeyExpiry         time.Duration `json:"key_expiry"`          // 键过期时间
}

// DefaultRateLimiterConfig 默认限流器配置
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		Enabled:          true,
		Algorithm:        "token_bucket",
		DefaultRate:      100,
		DefaultBurst:     200,
		MaxWaitTime:      30 * time.Second,
		PriorityQueues:   4,
		EnableAdaptive:   true,
		CleanupInterval:  1 * time.Minute,
		KeyExpiry:        10 * time.Minute,
	}
}

// QueryRateLimiter 查询限流器
type QueryRateLimiter struct {
	config       RateLimiterConfig
	tokenBuckets map[string]*TokenBucket
	slidingWindows map[string]*SlidingWindow
	priorityQueues map[QueryPriority]*PriorityQueue
	stats        *RateLimiterStats
	adaptor      *AdaptiveRateLimiter
	mu           sync.RWMutex
}

// RateLimiterStats 限流器统计
type RateLimiterStats struct {
	TotalRequests   int64         `json:"total_requests"`
	AllowedRequests int64         `json:"allowed_requests"`
	LimitedRequests int64         `json:"limited_requests"`
	CurrentQueued   int64         `json:"current_queued"`
	AvgWaitTime     time.Duration `json:"avg_wait_time"`
	ActiveKeys      int64         `json:"active_keys"`
	LastCleanup     time.Time     `json:"last_cleanup"`
}

// NewQueryRateLimiter 创建查询限流器
func NewQueryRateLimiter(config RateLimiterConfig) *QueryRateLimiter {
	limiter := &QueryRateLimiter{
		config:         config,
		tokenBuckets:   make(map[string]*TokenBucket),
		slidingWindows: make(map[string]*SlidingWindow),
		priorityQueues: make(map[QueryPriority]*PriorityQueue),
		stats:          &RateLimiterStats{LastCleanup: time.Now()},
	}

	// 初始化优先级队列
	priorities := []QueryPriority{PriorityCritical, PriorityHigh, PriorityNormal, PriorityLow}
	for _, p := range priorities {
		limiter.priorityQueues[p] = NewPriorityQueue(int(p))
	}

	// 初始化自适应限流器
	if config.EnableAdaptive {
		limiter.adaptor = NewAdaptiveRateLimiter(config.DefaultRate)
	}

	// 启动后台清理
	go limiter.backgroundCleanup()

	return limiter
}

// Allow 检查是否允许请求
func (r *QueryRateLimiter) Allow(ctx context.Context, key string, priority QueryPriority) (*RateLimitResult, error) {
	if !r.config.Enabled {
		return &RateLimitResult{
			Status:   RateLimitStatusAllowed,
			Allowed:  true,
			Priority: priority,
		}, nil
	}

	atomic.AddInt64(&r.stats.TotalRequests, 1)

	// 根据算法选择限流方式
	var result *RateLimitResult
	var err error

	switch r.config.Algorithm {
	case "token_bucket":
		result, err = r.checkTokenBucket(key)
	case "sliding_window":
		result, err = r.checkSlidingWindow(key)
	case "leaky_bucket":
		result, err = r.checkLeakyBucket(key)
	default:
		result, err = r.checkTokenBucket(key)
	}

	if err != nil {
		return nil, err
	}

	result.Key = key
	result.Priority = priority

	// 更新统计
	if result.Allowed {
		atomic.AddInt64(&r.stats.AllowedRequests, 1)
	} else {
		atomic.AddInt64(&r.stats.LimitedRequests, 1)
	}

	return result, nil
}

// AllowAndWait 检查是否允许请求，如果不允许则等待
func (r *QueryRateLimiter) AllowAndWait(ctx context.Context, key string, priority QueryPriority) (*RateLimitResult, error) {
	result, err := r.Allow(ctx, key, priority)
	if err != nil {
		return nil, err
	}

	if result.Allowed {
		return result, nil
	}

	// 加入优先级队列等待
	if r.config.MaxWaitTime > 0 {
		return r.waitInQueue(ctx, key, priority)
	}

	return result, nil
}

// waitInQueue 在队列中等待
func (r *QueryRateLimiter) waitInQueue(ctx context.Context, key string, priority QueryPriority) (*RateLimitResult, error) {
	queue, ok := r.priorityQueues[priority]
	if !ok {
		queue = r.priorityQueues[PriorityNormal]
	}

	// 加入队列
	queue.Add(key)
	atomic.AddInt64(&r.stats.CurrentQueued, 1)

	defer func() {
		queue.Remove(key)
		atomic.AddInt64(&r.stats.CurrentQueued, -1)
	}()

	// 等待轮到执行
	waitChan := make(chan struct{})
	go func() {
		for {
			if queue.IsNext(key) {
				close(waitChan)
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	select {
	case <-ctx.Done():
		return &RateLimitResult{
			Status:   RateLimitStatusLimited,
			Allowed:  false,
			Key:      key,
			Priority: priority,
		}, ctx.Err()
	case <-time.After(r.config.MaxWaitTime):
		return &RateLimitResult{
			Status:     RateLimitStatusLimited,
			Allowed:    false,
			RetryAfter: r.config.MaxWaitTime,
			Key:        key,
			Priority:   priority,
		}, errors.New("wait timeout")
	case <-waitChan:
		// 轮到执行，再次检查
		return r.Allow(ctx, key, priority)
	}
}

// checkTokenBucket 令牌桶算法检查
func (r *QueryRateLimiter) checkTokenBucket(key string) (*RateLimitResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	bucket, exists := r.tokenBuckets[key]
	if !exists {
		bucket = NewTokenBucket(r.config.DefaultRate, r.config.DefaultBurst)
		r.tokenBuckets[key] = bucket
	}

	allowed := bucket.Take()
	remaining := bucket.Tokens()

	return &RateLimitResult{
		Status:    boolToStatus(allowed),
		Allowed:   allowed,
		Limit:     r.config.DefaultBurst,
		Remaining: int(remaining),
		ResetAt:   time.Now().Add(time.Second),
	}, nil
}

// checkSlidingWindow 滑动窗口算法检查
func (r *QueryRateLimiter) checkSlidingWindow(key string) (*RateLimitResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	window, exists := r.slidingWindows[key]
	if !exists {
		window = NewSlidingWindow(r.config.DefaultRate, time.Second)
		r.slidingWindows[key] = window
	}

	allowed := window.Allow()
	count := window.Count()

	return &RateLimitResult{
		Status:    boolToStatus(allowed),
		Allowed:   allowed,
		Limit:     r.config.DefaultRate,
		Remaining: max(0, r.config.DefaultRate-count),
		ResetAt:   time.Now().Add(time.Second),
	}, nil
}

// checkLeakyBucket 漏桶算法检查
func (r *QueryRateLimiter) checkLeakyBucket(key string) (*RateLimitResult, error) {
	// 漏桶算法使用滑动窗口实现
	return r.checkSlidingWindow(key)
}

// SetRate 设置速率
func (r *QueryRateLimiter) SetRate(key string, rate int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if bucket, exists := r.tokenBuckets[key]; exists {
		bucket.SetRate(rate)
	}

	if window, exists := r.slidingWindows[key]; exists {
		window.SetLimit(rate)
	}
}

// GetStats 获取统计信息
func (r *QueryRateLimiter) GetStats() *RateLimiterStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return &RateLimiterStats{
		TotalRequests:   atomic.LoadInt64(&r.stats.TotalRequests),
		AllowedRequests: atomic.LoadInt64(&r.stats.AllowedRequests),
		LimitedRequests: atomic.LoadInt64(&r.stats.LimitedRequests),
		CurrentQueued:   atomic.LoadInt64(&r.stats.CurrentQueued),
		AvgWaitTime:     r.stats.AvgWaitTime,
		ActiveKeys:      int64(len(r.tokenBuckets) + len(r.slidingWindows)),
		LastCleanup:     r.stats.LastCleanup,
	}
}

// Reset 重置限流器
func (r *QueryRateLimiter) Reset(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.tokenBuckets, key)
	delete(r.slidingWindows, key)
}

// backgroundCleanup 后台清理
func (r *QueryRateLimiter) backgroundCleanup() {
	ticker := time.NewTicker(r.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		r.cleanup()
	}
}

// cleanup 清理过期条目
func (r *QueryRateLimiter) cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	// 清理过期的令牌桶
	for key, bucket := range r.tokenBuckets {
		if now.Sub(bucket.LastRefill()) > r.config.KeyExpiry {
			delete(r.tokenBuckets, key)
		}
	}

	// 清理过期的滑动窗口
	for key, window := range r.slidingWindows {
		if now.Sub(window.LastAccess()) > r.config.KeyExpiry {
			delete(r.slidingWindows, key)
		}
	}

	r.stats.LastCleanup = now
}

// TokenBucket 令牌桶
type TokenBucket struct {
	rate       float64   // 每秒放入令牌数
	capacity   float64   // 桶容量
	tokens     float64   // 当前令牌数
	lastRefill time.Time // 上次填充时间
	mu         sync.Mutex
}

// NewTokenBucket 创建令牌桶
func NewTokenBucket(rate, capacity int) *TokenBucket {
	return &TokenBucket{
		rate:       float64(rate),
		capacity:   float64(capacity),
		tokens:     float64(capacity),
		lastRefill: time.Now(),
	}
}

// Take 取出一个令牌
func (b *TokenBucket) Take() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.refill()

	if b.tokens >= 1 {
		b.tokens--
		return true
	}

	return false
}

// TakeN 取出N个令牌
func (b *TokenBucket) TakeN(n int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.refill()

	if b.tokens >= float64(n) {
		b.tokens -= float64(n)
		return true
	}

	return false
}

// Tokens 获取当前令牌数
func (b *TokenBucket) Tokens() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.refill()
	return b.tokens
}

// refill 填充令牌
func (b *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.lastRefill = now

	b.tokens = min(b.capacity, b.tokens+elapsed*b.rate)
}

// SetRate 设置速率
func (b *TokenBucket) SetRate(rate int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.rate = float64(rate)
}

// LastRefill 获取上次填充时间
func (b *TokenBucket) LastRefill() time.Time {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.lastRefill
}

// SlidingWindow 滑动窗口
type SlidingWindow struct {
	limit      int
	windowSize time.Duration
	timestamps []time.Time
	lastAccess time.Time
	mu         sync.Mutex
}

// NewSlidingWindow 创建滑动窗口
func NewSlidingWindow(limit int, windowSize time.Duration) *SlidingWindow {
	return &SlidingWindow{
		limit:      limit,
		windowSize: windowSize,
		timestamps: make([]time.Time, 0),
		lastAccess: time.Now(),
	}
}

// Allow 检查是否允许
func (w *SlidingWindow) Allow() bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.lastAccess = time.Now()
	w.cleanup()

	if len(w.timestamps) < w.limit {
		w.timestamps = append(w.timestamps, time.Now())
		return true
	}

	return false
}

// Count 获取当前计数
func (w *SlidingWindow) Count() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.cleanup()
	return len(w.timestamps)
}

// cleanup 清理过期时间戳
func (w *SlidingWindow) cleanup() {
	now := time.Now()
	cutoff := now.Add(-w.windowSize)

	validIdx := 0
	for _, ts := range w.timestamps {
		if ts.After(cutoff) {
			w.timestamps[validIdx] = ts
			validIdx++
		}
	}
	w.timestamps = w.timestamps[:validIdx]
}

// SetLimit 设置限制
func (w *SlidingWindow) SetLimit(limit int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.limit = limit
}

// LastAccess 获取最后访问时间
func (w *SlidingWindow) LastAccess() time.Time {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.lastAccess
}

// PriorityQueue 优先级队列
type PriorityQueue struct {
	capacity  int
	items     map[string]int
	positions []string
	mu        sync.Mutex
}

// NewPriorityQueue 创建优先级队列
func NewPriorityQueue(capacity int) *PriorityQueue {
	return &PriorityQueue{
		capacity:  capacity,
		items:     make(map[string]int),
		positions: make([]string, 0),
	}
}

// Add 添加元素
func (q *PriorityQueue) Add(key string) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.items[key]; exists {
		return q.items[key]
	}

	position := len(q.positions)
	q.items[key] = position
	q.positions = append(q.positions, key)

	return position
}

// Remove 移除元素
func (q *PriorityQueue) Remove(key string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if pos, exists := q.items[key]; exists {
		// 移除元素
		q.positions = append(q.positions[:pos], q.positions[pos+1:]...)
		delete(q.items, key)

		// 更新位置
		for i := pos; i < len(q.positions); i++ {
			q.items[q.positions[i]] = i
		}
	}
}

// IsNext 检查是否是下一个
func (q *PriorityQueue) IsNext(key string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.positions) == 0 {
		return false
	}

	return q.positions[0] == key
}

// Size 获取队列大小
func (q *PriorityQueue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.positions)
}

// AdaptiveRateLimiter 自适应限流器
type AdaptiveRateLimiter struct {
	baseRate      int
	currentRate   int
	minRate       int
	maxRate       int
	successCount  int64
	failureCount  int64
	adjustFactor  float64
	lastAdjust    time.Time
	mu            sync.RWMutex
}

// NewAdaptiveRateLimiter 创建自适应限流器
func NewAdaptiveRateLimiter(baseRate int) *AdaptiveRateLimiter {
	return &AdaptiveRateLimiter{
		baseRate:     baseRate,
		currentRate:  baseRate,
		minRate:      baseRate / 10,
		maxRate:      baseRate * 10,
		adjustFactor: 1.0,
		lastAdjust:   time.Now(),
	}
}

// RecordSuccess 记录成功
func (a *AdaptiveRateLimiter) RecordSuccess() {
	atomic.AddInt64(&a.successCount, 1)
	a.maybeAdjust()
}

// RecordFailure 记录失败
func (a *AdaptiveRateLimiter) RecordFailure() {
	atomic.AddInt64(&a.failureCount, 1)
	a.maybeAdjust()
}

// maybeAdjust 可能调整速率
func (a *AdaptiveRateLimiter) maybeAdjust() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 每100个请求调整一次
	total := atomic.LoadInt64(&a.successCount) + atomic.LoadInt64(&a.failureCount)
	if total%100 != 0 {
		return
	}

	// 计算成功率
	successRate := float64(atomic.LoadInt64(&a.successCount)) / float64(total)

	// 根据成功率调整
	if successRate > 0.95 {
		// 成功率高，增加速率
		newRate := int(float64(a.currentRate) * 1.1)
		if newRate < a.maxRate {
			a.currentRate = newRate
		} else {
			a.currentRate = a.maxRate
		}
	} else if successRate < 0.8 {
		// 成功率低，降低速率
		newRate := int(float64(a.currentRate) * 0.9)
		if newRate > a.minRate {
			a.currentRate = newRate
		} else {
			a.currentRate = a.minRate
		}
	}

	a.lastAdjust = time.Now()
}

// GetCurrentRate 获取当前速率
func (a *AdaptiveRateLimiter) GetCurrentRate() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.currentRate
}

// RateLimitMiddleware 限流中间件
type RateLimitMiddleware struct {
	limiter  *QueryRateLimiter
	keyFunc  func(*QueryRequest) string
	onLimit  func(*QueryRequest, *RateLimitResult) error
}

// NewRateLimitMiddleware 创建限流中间件
func NewRateLimitMiddleware(limiter *QueryRateLimiter) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: limiter,
		keyFunc: DefaultKeyFunc,
		onLimit: DefaultOnLimitFunc,
	}
}

// SetKeyFunc 设置键函数
func (m *RateLimitMiddleware) SetKeyFunc(fn func(*QueryRequest) string) {
	m.keyFunc = fn
}

// SetOnLimit 设置限流回调
func (m *RateLimitMiddleware) SetOnLimit(fn func(*QueryRequest, *RateLimitResult) error) {
	m.onLimit = fn
}

// Process 处理请求
func (m *RateLimitMiddleware) Process(ctx context.Context, req *QueryRequest) (*RateLimitResult, error) {
	key := m.keyFunc(req)
	priority := req.Priority
	if priority == 0 {
		priority = PriorityNormal
	}

	result, err := m.limiter.AllowAndWait(ctx, key, priority)
	if err != nil {
		return nil, err
	}

	if !result.Allowed {
		if m.onLimit != nil {
			return result, m.onLimit(req, result)
		}
		return result, ErrQueryRateLimited
	}

	return result, nil
}

// DefaultKeyFunc 默认键函数
func DefaultKeyFunc(req *QueryRequest) string {
	return fmt.Sprintf("query:%s:%s", req.Database, req.Table)
}

// DefaultOnLimitFunc 默认限流回调
func DefaultOnLimitFunc(req *QueryRequest, result *RateLimitResult) error {
	return &QueryError{
		Code:    "RATE_LIMITED",
		Message: fmt.Sprintf("rate limited, retry after %v", result.RetryAfter),
	}
}

// DistributedRateLimiter 分布式限流器
type DistributedRateLimiter struct {
	local    *QueryRateLimiter
	redis    RedisRateLimitClient
	key      string
	config   DistributedRateLimitConfig
}

// RedisRateLimitClient Redis限流客户端接口
type RedisRateLimitClient interface {
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
}

// DistributedRateLimitConfig 分布式限流配置
type DistributedRateLimitConfig struct {
	GlobalRate  int           `json:"global_rate"`
	LocalRate   int           `json:"local_rate"`
	SyncPeriod  time.Duration `json:"sync_period"`
	KeyPrefix   string        `json:"key_prefix"`
}

// NewDistributedRateLimiter 创建分布式限流器
func NewDistributedRateLimiter(local *QueryRateLimiter, redis RedisRateLimitClient, config DistributedRateLimitConfig) *DistributedRateLimiter {
	return &DistributedRateLimiter{
		local:  local,
		redis:  redis,
		config: config,
	}
}

// Allow 分布式限流检查
func (d *DistributedRateLimiter) Allow(ctx context.Context, key string, priority QueryPriority) (*RateLimitResult, error) {
	// 先检查本地限流
	localResult, err := d.local.Allow(ctx, key, priority)
	if err != nil {
		return nil, err
	}

	if !localResult.Allowed {
		return localResult, nil
	}

	// 检查全局限流
	globalKey := d.config.KeyPrefix + key
	count, err := d.redis.Incr(ctx, globalKey)
	if err != nil {
		// Redis错误时降级到本地限流
		return localResult, nil
	}

	// 设置过期时间
	if count == 1 {
		d.redis.Expire(ctx, globalKey, time.Second)
	}

	if int(count) > d.config.GlobalRate {
		// 超过全局限制
		return &RateLimitResult{
			Status:     RateLimitStatusLimited,
			Allowed:    false,
			Limit:      d.config.GlobalRate,
			Remaining:  0,
			RetryAfter: time.Second,
			Key:        key,
			Priority:   priority,
		}, nil
	}

	return &RateLimitResult{
		Status:    RateLimitStatusAllowed,
		Allowed:   true,
		Limit:     d.config.GlobalRate,
		Remaining: d.config.GlobalRate - int(count),
		Key:       key,
		Priority:  priority,
	}, nil
}

// RateLimitGroup 限流组
type RateLimitGroup struct {
	name     string
	limiters map[string]*QueryRateLimiter
	defaults RateLimiterConfig
	mu       sync.RWMutex
}

// NewRateLimitGroup 创建限流组
func NewRateLimitGroup(name string, defaults RateLimiterConfig) *RateLimitGroup {
	return &RateLimitGroup{
		name:     name,
		limiters: make(map[string]*QueryRateLimiter),
		defaults: defaults,
	}
}

// Get 获取限流器
func (g *RateLimitGroup) Get(key string) *QueryRateLimiter {
	g.mu.RLock()
	limiter, exists := g.limiters[key]
	g.mu.RUnlock()

	if exists {
		return limiter
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// 双重检查
	if limiter, exists = g.limiters[key]; exists {
		return limiter
	}

	limiter = NewQueryRateLimiter(g.defaults)
	g.limiters[key] = limiter

	return limiter
}

// Allow 组内限流检查
func (g *RateLimitGroup) Allow(ctx context.Context, groupKey, key string, priority QueryPriority) (*RateLimitResult, error) {
	limiter := g.Get(groupKey)
	return limiter.Allow(ctx, key, priority)
}

// RateLimitRule 限流规则
type RateLimitRule struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	KeyPattern  string        `json:"key_pattern"`
	Rate        int           `json:"rate"`
	Burst       int           `json:"burst"`
	Priority    QueryPriority `json:"priority"`
	Enabled     bool          `json:"enabled"`
	Description string        `json:"description"`
}

// RateLimitRuleManager 限流规则管理器
type RateLimitRuleManager struct {
	rules   map[string]*RateLimitRule
	limiter *QueryRateLimiter
	mu      sync.RWMutex
}

// NewRateLimitRuleManager 创建限流规则管理器
func NewRateLimitRuleManager(limiter *QueryRateLimiter) *RateLimitRuleManager {
	return &RateLimitRuleManager{
		rules:   make(map[string]*RateLimitRule),
		limiter: limiter,
	}
}

// AddRule 添加规则
func (m *RateLimitRuleManager) AddRule(rule *RateLimitRule) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rules[rule.ID] = rule
}

// RemoveRule 移除规则
func (m *RateLimitRuleManager) RemoveRule(ruleID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rules, ruleID)
}

// GetRule 获取规则
func (m *RateLimitRuleManager) GetRule(ruleID string) (*RateLimitRule, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	rule, exists := m.rules[ruleID]
	return rule, exists
}

// MatchRule 匹配规则
func (m *RateLimitRuleManager) MatchRule(key string) *RateLimitRule {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, rule := range m.rules {
		if !rule.Enabled {
			continue
		}
		if matchPattern(rule.KeyPattern, key) {
			return rule
		}
	}

	return nil
}

// matchPattern 匹配模式
func matchPattern(pattern, key string) bool {
	// 简单的通配符匹配
	if pattern == "*" {
		return true
	}
	if pattern == key {
		return true
	}
	// 前缀匹配
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// 辅助函数
func boolToStatus(allowed bool) RateLimitStatus {
	if allowed {
		return RateLimitStatusAllowed
	}
	return RateLimitStatusLimited
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
