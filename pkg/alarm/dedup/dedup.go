package dedup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
)

// Fingerprint 告警指纹
type Fingerprint string

// DeduplicationResult 去重结果
type DeduplicationResult struct {
	IsDuplicate   bool        `json:"is_duplicate"`
	Fingerprint   Fingerprint `json:"fingerprint"`
	FirstSeenAt   time.Time   `json:"first_seen_at,omitempty"`
	LastSeenAt    time.Time   `json:"last_seen_at,omitempty"`
	OccurrenceCount int       `json:"occurrence_count"`
}

// DeduplicationStats 去重统计
type DeduplicationStats struct {
	TotalProcessed    int64 `json:"total_processed"`
	DuplicatesFound   int64 `json:"duplicates_found"`
	UniqueAlarms      int64 `json:"unique_alarms"`
	CacheHits         int64 `json:"cache_hits"`
	CacheMisses       int64 `json:"cache_misses"`
	CurrentCacheSize  int   `json:"current_cache_size"`
}

// DeduplicationConfig 去重配置
type DeduplicationConfig struct {
	// WindowDuration 去重窗口时长
	WindowDuration time.Duration
	// MaxCacheSize 最大缓存大小
	MaxCacheSize int
	// CleanupInterval 清理间隔
	CleanupInterval time.Duration
	// IncludeValue 是否在指纹中包含值
	IncludeValue bool
	// IncludeThreshold 是否在指纹中包含阈值
	IncludeThreshold bool
	// CustomFingerprintFields 自定义指纹字段
	CustomFingerprintFields []string
}

// DefaultDeduplicationConfig 默认去重配置
func DefaultDeduplicationConfig() DeduplicationConfig {
	return DeduplicationConfig{
		WindowDuration:    5 * time.Minute,
		MaxCacheSize:      100000,
		CleanupInterval:   1 * time.Minute,
		IncludeValue:      false,
		IncludeThreshold:  false,
	}
}

// cacheEntry 缓存条目
type cacheEntry struct {
	fingerprint     Fingerprint
	firstSeenAt     time.Time
	lastSeenAt      time.Time
	occurrenceCount int
	expireAt        time.Time
}

// Deduplicator 去重器
type Deduplicator struct {
	mu     sync.RWMutex
	config DeduplicationConfig
	cache  map[string]*cacheEntry
	stats  DeduplicationStats

	// 用于分布式场景的接口
	distributedCache DistributedCache
}

// DistributedCache 分布式缓存接口
type DistributedCache interface {
	Get(ctx context.Context, key string) (*cacheEntry, error)
	Set(ctx context.Context, key string, entry *cacheEntry, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// NewDeduplicator 创建去重器
func NewDeduplicator(config DeduplicationConfig) *Deduplicator {
	if config.WindowDuration <= 0 {
		config.WindowDuration = 5 * time.Minute
	}
	if config.MaxCacheSize <= 0 {
		config.MaxCacheSize = 100000
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = 1 * time.Minute
	}

	return &Deduplicator{
		config: config,
		cache:  make(map[string]*cacheEntry),
	}
}

// NewDeduplicatorWithDistributed 创建带分布式缓存的去重器
func NewDeduplicatorWithDistributed(config DeduplicationConfig, distributedCache DistributedCache) *Deduplicator {
	d := NewDeduplicator(config)
	d.distributedCache = distributedCache
	return d
}

// GenerateFingerprint 生成告警指纹
func (d *Deduplicator) GenerateFingerprint(alarm *entity.Alarm) Fingerprint {
	// 构建指纹数据
	data := fmt.Sprintf("%s|%s|%s|%s|%d",
		alarm.PointID,
		alarm.DeviceID,
		alarm.StationID,
		alarm.Type,
		alarm.Level,
	)

	if d.config.IncludeValue {
		data += fmt.Sprintf("|%f", alarm.Value)
	}
	if d.config.IncludeThreshold {
		data += fmt.Sprintf("|%f", alarm.Threshold)
	}

	// 添加自定义字段
	for range d.config.CustomFingerprintFields {
		// 可以根据需要从alarm中提取自定义字段
	}

	// 计算SHA256哈希
	hash := sha256.Sum256([]byte(data))
	return Fingerprint(hex.EncodeToString(hash[:]))
}

// GenerateFingerprintFromFields 从字段生成指纹
func GenerateFingerprintFromFields(pointID, deviceID, stationID string, alarmType entity.AlarmType, level entity.AlarmLevel) Fingerprint {
	data := fmt.Sprintf("%s|%s|%s|%s|%d", pointID, deviceID, stationID, alarmType, level)
	hash := sha256.Sum256([]byte(data))
	return Fingerprint(hex.EncodeToString(hash[:]))
}

// Check 检查告警是否重复
func (d *Deduplicator) Check(ctx context.Context, alarm *entity.Alarm) *DeduplicationResult {
	fingerprint := d.GenerateFingerprint(alarm)
	now := time.Now()

	// 如果有分布式缓存，优先使用
	if d.distributedCache != nil {
		return d.checkDistributed(ctx, fingerprint, now)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.stats.TotalProcessed++

	// 检查缓存
	entry, exists := d.cache[string(fingerprint)]
	if exists && entry.expireAt.After(now) {
		// 命中缓存，更新条目
		entry.lastSeenAt = now
		entry.occurrenceCount++
		d.stats.DuplicatesFound++
		d.stats.CacheHits++

		return &DeduplicationResult{
			IsDuplicate:     true,
			Fingerprint:     fingerprint,
			FirstSeenAt:     entry.firstSeenAt,
			LastSeenAt:      entry.lastSeenAt,
			OccurrenceCount: entry.occurrenceCount,
		}
	}

	// 未命中，创建新条目
	d.stats.CacheMisses++
	d.stats.UniqueAlarms++

	// 检查缓存大小
	if len(d.cache) >= d.config.MaxCacheSize {
		d.cleanupExpiredLocked(now)
		if len(d.cache) >= d.config.MaxCacheSize {
			// 强制清理最旧的条目
			d.evictOldestLocked()
		}
	}

	d.cache[string(fingerprint)] = &cacheEntry{
		fingerprint:     fingerprint,
		firstSeenAt:     now,
		lastSeenAt:      now,
		occurrenceCount: 1,
		expireAt:        now.Add(d.config.WindowDuration),
	}
	d.stats.CurrentCacheSize = len(d.cache)

	return &DeduplicationResult{
		IsDuplicate:     false,
		Fingerprint:     fingerprint,
		FirstSeenAt:     now,
		LastSeenAt:      now,
		OccurrenceCount: 1,
	}
}

// checkDistributed 使用分布式缓存检查
func (d *Deduplicator) checkDistributed(ctx context.Context, fingerprint Fingerprint, now time.Time) *DeduplicationResult {
	d.stats.TotalProcessed++

	entry, err := d.distributedCache.Get(ctx, string(fingerprint))
	if err == nil && entry != nil {
		// 命中缓存
		entry.lastSeenAt = now
		entry.occurrenceCount++
		_ = d.distributedCache.Set(ctx, string(fingerprint), entry, d.config.WindowDuration)

		d.mu.Lock()
		d.stats.DuplicatesFound++
		d.stats.CacheHits++
		d.mu.Unlock()

		return &DeduplicationResult{
			IsDuplicate:     true,
			Fingerprint:     fingerprint,
			FirstSeenAt:     entry.firstSeenAt,
			LastSeenAt:      entry.lastSeenAt,
			OccurrenceCount: entry.occurrenceCount,
		}
	}

	// 未命中，创建新条目
	d.mu.Lock()
	d.stats.CacheMisses++
	d.stats.UniqueAlarms++
	d.mu.Unlock()

	newEntry := &cacheEntry{
		fingerprint:     fingerprint,
		firstSeenAt:     now,
		lastSeenAt:      now,
		occurrenceCount: 1,
		expireAt:        now.Add(d.config.WindowDuration),
	}
	_ = d.distributedCache.Set(ctx, string(fingerprint), newEntry, d.config.WindowDuration)

	return &DeduplicationResult{
		IsDuplicate:     false,
		Fingerprint:     fingerprint,
		FirstSeenAt:     now,
		LastSeenAt:      now,
		OccurrenceCount: 1,
	}
}

// IsDuplicate 检查是否重复（简化接口）
func (d *Deduplicator) IsDuplicate(ctx context.Context, alarm *entity.Alarm) bool {
	result := d.Check(ctx, alarm)
	return result.IsDuplicate
}

// MarkAsSeen 标记告警为已见
func (d *Deduplicator) MarkAsSeen(ctx context.Context, alarm *entity.Alarm) {
	d.Check(ctx, alarm)
}

// Remove 从去重缓存中移除
func (d *Deduplicator) Remove(ctx context.Context, fingerprint Fingerprint) error {
	if d.distributedCache != nil {
		return d.distributedCache.Delete(ctx, string(fingerprint))
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.cache, string(fingerprint))
	d.stats.CurrentCacheSize = len(d.cache)
	return nil
}

// cleanupExpiredLocked 清理过期条目（需要持有锁）
func (d *Deduplicator) cleanupExpiredLocked(now time.Time) {
	for key, entry := range d.cache {
		if entry.expireAt.Before(now) {
			delete(d.cache, key)
		}
	}
	d.stats.CurrentCacheSize = len(d.cache)
}

// evictOldestLocked 清理最旧的条目（需要持有锁）
func (d *Deduplicator) evictOldestLocked() {
	if len(d.cache) == 0 {
		return
	}

	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, entry := range d.cache {
		if first || entry.firstSeenAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.firstSeenAt
			first = false
		}
	}

	if oldestKey != "" {
		delete(d.cache, oldestKey)
	}
	d.stats.CurrentCacheSize = len(d.cache)
}

// Cleanup 清理过期条目
func (d *Deduplicator) Cleanup() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cleanupExpiredLocked(time.Now())
}

// StartCleanup 启动定期清理
func (d *Deduplicator) StartCleanup(ctx context.Context) {
	ticker := time.NewTicker(d.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.Cleanup()
		}
	}
}

// GetStats 获取统计信息
func (d *Deduplicator) GetStats() DeduplicationStats {
	d.mu.RLock()
	defer d.mu.RUnlock()
	stats := d.stats
	stats.CurrentCacheSize = len(d.cache)
	return stats
}

// Reset 重置去重器
func (d *Deduplicator) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.cache = make(map[string]*cacheEntry)
	d.stats = DeduplicationStats{}
}

// SetWindowDuration 设置去重窗口时长
func (d *Deduplicator) SetWindowDuration(duration time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.config.WindowDuration = duration
}

// GetWindowDuration 获取去重窗口时长
func (d *Deduplicator) GetWindowDuration() time.Duration {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.config.WindowDuration
}

// BatchCheck 批量检查
func (d *Deduplicator) BatchCheck(ctx context.Context, alarms []*entity.Alarm) []*DeduplicationResult {
	results := make([]*DeduplicationResult, len(alarms))
	for i, alarm := range alarms {
		results[i] = d.Check(ctx, alarm)
	}
	return results
}

// FilterDuplicates 过滤重复告警
func (d *Deduplicator) FilterDuplicates(ctx context.Context, alarms []*entity.Alarm) []*entity.Alarm {
	var uniqueAlarms []*entity.Alarm
	for _, alarm := range alarms {
		result := d.Check(ctx, alarm)
		if !result.IsDuplicate {
			uniqueAlarms = append(uniqueAlarms, alarm)
		}
	}
	return uniqueAlarms
}

// GetFingerprintInfo 获取指纹信息
func (d *Deduplicator) GetFingerprintInfo(ctx context.Context, fingerprint Fingerprint) (*cacheEntry, bool) {
	if d.distributedCache != nil {
		entry, err := d.distributedCache.Get(ctx, string(fingerprint))
		if err != nil || entry == nil {
			return nil, false
		}
		return entry, true
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	entry, exists := d.cache[string(fingerprint)]
	if !exists {
		return nil, false
	}
	return entry, true
}

// UpdateFingerprintTTL 更新指纹TTL
func (d *Deduplicator) UpdateFingerprintTTL(ctx context.Context, fingerprint Fingerprint, ttl time.Duration) error {
	if d.distributedCache != nil {
		entry, err := d.distributedCache.Get(ctx, string(fingerprint))
		if err != nil {
			return err
		}
		if entry == nil {
			return fmt.Errorf("fingerprint not found: %s", fingerprint)
		}
		return d.distributedCache.Set(ctx, string(fingerprint), entry, ttl)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	entry, exists := d.cache[string(fingerprint)]
	if !exists {
		return fmt.Errorf("fingerprint not found: %s", fingerprint)
	}
	entry.expireAt = time.Now().Add(ttl)
	return nil
}
