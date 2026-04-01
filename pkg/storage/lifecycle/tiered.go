package lifecycle

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DataTier 数据层级
type DataTier int

const (
	TierHot  DataTier = iota // 热数据 - Redis/内存
	TierWarm                 // 温数据 - PostgreSQL
	TierCold                 // 冷数据 - 时序数据库/对象存储
)

func (t DataTier) String() string {
	switch t {
	case TierHot:
		return "hot"
	case TierWarm:
		return "warm"
	case TierCold:
		return "cold"
	default:
		return "unknown"
	}
}

// DataItem 数据项
type DataItem struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Tier      DataTier               `json:"tier"`
	HitCount  int64                  `json:"hit_count"`
	LastAccess int64                 `json:"last_access"`
	Size      int64                  `json:"size"`
}

// AccessRecord 访问记录
type AccessRecord struct {
	Key       string    `json:"key"`
	Timestamp time.Time `json:"timestamp"`
	Count     int64     `json:"count"`
}

// TierConfig 分层配置
type TierConfig struct {
	// 热数据配置
	HotTTL           time.Duration `yaml:"hot_ttl" json:"hot_ttl"`                       // 热数据保留时间
	HotMaxSize       int64         `yaml:"hot_max_size" json:"hot_max_size"`             // 热数据最大容量
	HotThreshold     int64         `yaml:"hot_threshold" json:"hot_threshold"`           // 热数据访问阈值

	// 温数据配置
	WarmTTL          time.Duration `yaml:"warm_ttl" json:"warm_ttl"`                     // 温数据保留时间
	WarmThreshold    int64         `yaml:"warm_threshold" json:"warm_threshold"`         // 温数据访问阈值

	// 冷数据配置
	ColdTTL          time.Duration `yaml:"cold_ttl" json:"cold_ttl"`                     // 冷数据保留时间

	// 迁移配置
	MigrationBatch   int           `yaml:"migration_batch" json:"migration_batch"`       // 每次迁移批次大小
	MigrationInterval time.Duration `yaml:"migration_interval" json:"migration_interval"` // 迁移检查间隔

	// 热度统计配置
	StatsWindow      time.Duration `yaml:"stats_window" json:"stats_window"`             // 统计窗口
	StatsInterval    time.Duration `yaml:"stats_interval" json:"stats_interval"`         // 统计间隔
}

// DefaultTierConfig 默认分层配置
func DefaultTierConfig() TierConfig {
	return TierConfig{
		HotTTL:           24 * time.Hour,
		HotMaxSize:       10 * 1024 * 1024 * 1024, // 10GB
		HotThreshold:     100,
		WarmTTL:          30 * 24 * time.Hour,
		WarmThreshold:    10,
		ColdTTL:          365 * 24 * time.Hour,
		MigrationBatch:   1000,
		MigrationInterval: 5 * time.Minute,
		StatsWindow:      1 * time.Hour,
		StatsInterval:    1 * time.Minute,
	}
}

// TieredStorage 分层存储
type TieredStorage struct {
	config    TierConfig
	redis     *redis.Client
	db        *gorm.DB
	logger    *zap.Logger

	// 热度统计
	hotStats  sync.Map // key -> *HotStats
	statsMu   sync.RWMutex

	// 迁移控制
	migrateCh chan string
	stopCh    chan struct{}
	wg        sync.WaitGroup

	// 指标
	metrics   *TierMetrics
}

// HotStats 热度统计
type HotStats struct {
	Key        string
	HitCount   int64
	LastAccess time.Time
	FirstSeen  time.Time
	WindowHits int64
	UpdatedAt  time.Time
}

// TierMetrics 分层指标
type TierMetrics struct {
	mu              sync.RWMutex
	HotDataCount    int64
	WarmDataCount   int64
	ColdDataCount   int64
	HotDataSize     int64
	WarmDataSize    int64
	ColdDataSize    int64
	MigrationCount  int64
	HitRate         float64
	MissRate        float64
	LastMigration   time.Time
}

// NewTieredStorage 创建分层存储
func NewTieredStorage(config TierConfig, redis *redis.Client, db *gorm.DB, logger *zap.Logger) *TieredStorage {
	return &TieredStorage{
		config:    config,
		redis:     redis,
		db:        db,
		logger:    logger,
		migrateCh: make(chan string, 10000),
		stopCh:    make(chan struct{}),
		metrics:   &TierMetrics{},
	}
}

// Start 启动分层存储
func (ts *TieredStorage) Start(ctx context.Context) error {
	ts.logger.Info("Starting tiered storage")

	// 启动迁移协程
	ts.wg.Add(1)
	go ts.migrationWorker(ctx)

	// 启动统计协程
	ts.wg.Add(1)
	go ts.statsWorker(ctx)

	// 启动指标更新协程
	ts.wg.Add(1)
	go ts.metricsWorker(ctx)

	ts.logger.Info("Tiered storage started")
	return nil
}

// Stop 停止分层存储
func (ts *TieredStorage) Stop() error {
	ts.logger.Info("Stopping tiered storage")
	close(ts.stopCh)
	ts.wg.Wait()
	ts.logger.Info("Tiered storage stopped")
	return nil
}

// Store 存储数据
func (ts *TieredStorage) Store(ctx context.Context, key string, data interface{}, ttl time.Duration) error {
	// 序列化数据
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// 存储到热数据层
	if err := ts.storeHot(ctx, key, dataBytes, ttl); err != nil {
		return fmt.Errorf("failed to store hot data: %w", err)
	}

	// 初始化热度统计
	ts.initHotStats(key)

	// 更新指标
	ts.metrics.mu.Lock()
	ts.metrics.HotDataCount++
	ts.metrics.HotDataSize += int64(len(dataBytes))
	ts.metrics.mu.Unlock()

	ts.logger.Debug("Data stored to hot tier",
		zap.String("key", key),
		zap.Int("size", len(dataBytes)),
	)

	return nil
}

// Get 获取数据
func (ts *TieredStorage) Get(ctx context.Context, key string, dest interface{}) error {
	// 先从热数据层获取
	data, err := ts.getHot(ctx, key)
	if err == nil {
		ts.recordHit(key, TierHot)
		return json.Unmarshal(data, dest)
	}

	// 从温数据层获取
	data, err = ts.getWarm(ctx, key)
	if err == nil {
		ts.recordHit(key, TierWarm)
		// 提升到热数据层
		go ts.promoteToHot(context.Background(), key, data)
		return json.Unmarshal(data, dest)
	}

	// 从冷数据层获取
	data, err = ts.getCold(ctx, key)
	if err == nil {
		ts.recordHit(key, TierCold)
		// 提升到热数据层
		go ts.promoteToHot(context.Background(), key, data)
		return json.Unmarshal(data, dest)
	}

	ts.recordMiss(key)
	return fmt.Errorf("data not found: %s", key)
}

// Delete 删除数据
func (ts *TieredStorage) Delete(ctx context.Context, key string) error {
	var errs []error

	// 从热数据层删除
	if err := ts.deleteHot(ctx, key); err != nil && err != redis.Nil {
		errs = append(errs, fmt.Errorf("hot delete: %w", err))
	}

	// 从温数据层删除
	if err := ts.deleteWarm(ctx, key); err != nil && err != gorm.ErrRecordNotFound {
		errs = append(errs, fmt.Errorf("warm delete: %w", err))
	}

	// 从冷数据层删除
	if err := ts.deleteCold(ctx, key); err != nil {
		errs = append(errs, fmt.Errorf("cold delete: %w", err))
	}

	// 删除热度统计
	ts.hotStats.Delete(key)

	if len(errs) > 0 {
		return fmt.Errorf("delete errors: %v", errs)
	}

	return nil
}

// GetTier 获取数据所在层级
func (ts *TieredStorage) GetTier(ctx context.Context, key string) (DataTier, error) {
	// 检查热数据层
	if exists, _ := ts.redis.Exists(ctx, key).Result(); exists > 0 {
		return TierHot, nil
	}

	// 检查温数据层
	var count int64
	if err := ts.db.Table("warm_data").Where("key = ?", key).Count(&count).Error; err != nil {
		return TierHot, err
	}
	if count > 0 {
		return TierWarm, nil
	}

	// 检查冷数据层
	if exists, _ := ts.checkColdExists(ctx, key); exists {
		return TierCold, nil
	}

	return TierHot, fmt.Errorf("data not found: %s", key)
}

// Migrate 迁移数据
func (ts *TieredStorage) Migrate(ctx context.Context, key string, fromTier, toTier DataTier) error {
	ts.logger.Info("Migrating data",
		zap.String("key", key),
		zap.String("from", fromTier.String()),
		zap.String("to", toTier.String()),
	)

	var data []byte
	var err error

	// 从源层级获取数据
	switch fromTier {
	case TierHot:
		data, err = ts.getHot(ctx, key)
	case TierWarm:
		data, err = ts.getWarm(ctx, key)
	case TierCold:
		data, err = ts.getCold(ctx, key)
	default:
		return fmt.Errorf("invalid source tier: %s", fromTier)
	}

	if err != nil {
		return fmt.Errorf("failed to get data from %s tier: %w", fromTier, err)
	}

	// 存储到目标层级
	switch toTier {
	case TierHot:
		err = ts.storeHot(ctx, key, data, ts.config.HotTTL)
	case TierWarm:
		err = ts.storeWarm(ctx, key, data)
	case TierCold:
		err = ts.storeCold(ctx, key, data)
	default:
		return fmt.Errorf("invalid target tier: %s", toTier)
	}

	if err != nil {
		return fmt.Errorf("failed to store data to %s tier: %w", toTier, err)
	}

	// 从源层级删除
	switch fromTier {
	case TierHot:
		ts.deleteHot(ctx, key)
	case TierWarm:
		ts.deleteWarm(ctx, key)
	case TierCold:
		ts.deleteCold(ctx, key)
	}

	// 更新指标
	ts.metrics.mu.Lock()
	ts.metrics.MigrationCount++
	ts.metrics.LastMigration = time.Now()
	ts.metrics.mu.Unlock()

	return nil
}

// GetMetrics 获取指标
func (ts *TieredStorage) GetMetrics() TierMetrics {
	ts.metrics.mu.RLock()
	defer ts.metrics.mu.RUnlock()
	return *ts.metrics
}

// 内部方法

func (ts *TieredStorage) storeHot(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	return ts.redis.Set(ctx, key, data, ttl).Err()
}

func (ts *TieredStorage) getHot(ctx context.Context, key string) ([]byte, error) {
	return ts.redis.Get(ctx, key).Bytes()
}

func (ts *TieredStorage) deleteHot(ctx context.Context, key string) error {
	return ts.redis.Del(ctx, key).Err()
}

func (ts *TieredStorage) storeWarm(ctx context.Context, key string, data []byte) error {
	return ts.db.Exec(`
		INSERT INTO warm_data (key, data, created_at, updated_at)
		VALUES (?, ?, NOW(), NOW())
		ON CONFLICT (key) DO UPDATE SET data = EXCLUDED.data, updated_at = NOW()
	`, key, data).Error
}

func (ts *TieredStorage) getWarm(ctx context.Context, key string) ([]byte, error) {
	var result struct {
		Data []byte
	}
	err := ts.db.Table("warm_data").Select("data").Where("key = ?", key).First(&result).Error
	return result.Data, err
}

func (ts *TieredStorage) deleteWarm(ctx context.Context, key string) error {
	return ts.db.Exec("DELETE FROM warm_data WHERE key = ?", key).Error
}

func (ts *TieredStorage) storeCold(ctx context.Context, key string, data []byte) error {
	// 存储到冷数据层（时序数据库或对象存储）
	// 这里简化为数据库存储，实际应该使用对象存储或时序数据库
	return ts.db.Exec(`
		INSERT INTO cold_data (key, data, created_at)
		VALUES (?, ?, NOW())
		ON CONFLICT (key) DO UPDATE SET data = EXCLUDED.data
	`, key, data).Error
}

func (ts *TieredStorage) getCold(ctx context.Context, key string) ([]byte, error) {
	var result struct {
		Data []byte
	}
	err := ts.db.Table("cold_data").Select("data").Where("key = ?", key).First(&result).Error
	return result.Data, err
}

func (ts *TieredStorage) deleteCold(ctx context.Context, key string) error {
	return ts.db.Exec("DELETE FROM cold_data WHERE key = ?", key).Error
}

func (ts *TieredStorage) checkColdExists(ctx context.Context, key string) (bool, error) {
	var count int64
	err := ts.db.Table("cold_data").Where("key = ?", key).Count(&count).Error
	return count > 0, err
}

func (ts *TieredStorage) promoteToHot(ctx context.Context, key string, data []byte) {
	if err := ts.storeHot(ctx, key, data, ts.config.HotTTL); err != nil {
		ts.logger.Error("Failed to promote data to hot tier",
			zap.String("key", key),
			zap.Error(err),
		)
	}
}

func (ts *TieredStorage) initHotStats(key string) {
	now := time.Now()
	ts.hotStats.Store(key, &HotStats{
		Key:        key,
		HitCount:   0,
		LastAccess: now,
		FirstSeen:  now,
		WindowHits: 0,
		UpdatedAt:  now,
	})
}

func (ts *TieredStorage) recordHit(key string, tier DataTier) {
	if stats, ok := ts.hotStats.Load(key); ok {
		s := stats.(*HotStats)
		s.HitCount++
		s.LastAccess = time.Now()
		s.WindowHits++
		s.UpdatedAt = time.Now()
	}

	// 更新命中率
	ts.metrics.mu.Lock()
	defer ts.metrics.mu.Unlock()
	total := ts.metrics.HitRate + ts.metrics.MissRate
	if total > 0 {
		ts.metrics.HitRate = (ts.metrics.HitRate + 1) / (total + 1)
		ts.metrics.MissRate = ts.metrics.MissRate / (total + 1)
	} else {
		ts.metrics.HitRate = 1
	}
}

func (ts *TieredStorage) recordMiss(key string) {
	ts.metrics.mu.Lock()
	defer ts.metrics.mu.Unlock()
	total := ts.metrics.HitRate + ts.metrics.MissRate
	if total > 0 {
		ts.metrics.HitRate = ts.metrics.HitRate / (total + 1)
		ts.metrics.MissRate = (ts.metrics.MissRate + 1) / (total + 1)
	} else {
		ts.metrics.MissRate = 1
	}
}

func (ts *TieredStorage) migrationWorker(ctx context.Context) {
	defer ts.wg.Done()

	ticker := time.NewTicker(ts.config.MigrationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ts.stopCh:
			return
		case <-ticker.C:
			ts.checkAndMigrate(ctx)
		case key := <-ts.migrateCh:
			ts.migrateKey(ctx, key)
		}
	}
}

func (ts *TieredStorage) checkAndMigrate(ctx context.Context) {
	// 检查热数据层是否需要迁移到温数据层
	ts.checkHotToWarmMigration(ctx)

	// 检查温数据层是否需要迁移到冷数据层
	ts.checkWarmToColdMigration(ctx)
}

func (ts *TieredStorage) checkHotToWarmMigration(ctx context.Context) {
	// 获取所有热数据键
	keys, err := ts.getHotKeys(ctx)
	if err != nil {
		ts.logger.Error("Failed to get hot keys", zap.Error(err))
		return
	}

	now := time.Now()
	for _, key := range keys {
		stats, ok := ts.hotStats.Load(key)
		if !ok {
			continue
		}

		s := stats.(*HotStats)

		// 检查是否需要迁移
		// 条件：访问次数低于阈值 且 超过保留时间
		if s.HitCount < ts.config.HotThreshold &&
			now.Sub(s.LastAccess) > ts.config.HotTTL {
			ts.migrateCh <- key
		}
	}
}

func (ts *TieredStorage) checkWarmToColdMigration(ctx context.Context) {
	// 获取需要迁移的温数据
	var keys []string
	err := ts.db.Table("warm_data").
		Select("key").
		Where("updated_at < ?", time.Now().Add(-ts.config.WarmTTL)).
		Limit(ts.config.MigrationBatch).
		Pluck("key", &keys).Error

	if err != nil {
		ts.logger.Error("Failed to get warm keys for migration", zap.Error(err))
		return
	}

	for _, key := range keys {
		ts.migrateCh <- key
	}
}

func (ts *TieredStorage) migrateKey(ctx context.Context, key string) {
	// 确定数据所在层级
	tier, err := ts.GetTier(ctx, key)
	if err != nil {
		ts.logger.Debug("Key not found for migration", zap.String("key", key))
		return
	}

	// 根据层级决定迁移目标
	var targetTier DataTier
	switch tier {
	case TierHot:
		targetTier = TierWarm
	case TierWarm:
		targetTier = TierCold
	default:
		return
	}

	// 执行迁移
	if err := ts.Migrate(ctx, key, tier, targetTier); err != nil {
		ts.logger.Error("Failed to migrate key",
			zap.String("key", key),
			zap.String("from", tier.String()),
			zap.String("to", targetTier.String()),
			zap.Error(err),
		)
	}
}

func (ts *TieredStorage) getHotKeys(ctx context.Context) ([]string, error) {
	// 使用 SCAN 命令获取所有键
	var keys []string
	iter := ts.redis.Scan(ctx, 0, "*", 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	return keys, iter.Err()
}

func (ts *TieredStorage) statsWorker(ctx context.Context) {
	defer ts.wg.Done()

	ticker := time.NewTicker(ts.config.StatsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ts.stopCh:
			return
		case <-ticker.C:
			ts.updateStats(ctx)
		}
	}
}

func (ts *TieredStorage) updateStats(ctx context.Context) {
	// 重置窗口统计
	ts.hotStats.Range(func(key, value interface{}) bool {
		stats := value.(*HotStats)
		stats.WindowHits = 0
		stats.UpdatedAt = time.Now()
		return true
	})
}

func (ts *TieredStorage) metricsWorker(ctx context.Context) {
	defer ts.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ts.stopCh:
			return
		case <-ticker.C:
			ts.updateMetrics(ctx)
		}
	}
}

func (ts *TieredStorage) updateMetrics(ctx context.Context) {
	// 更新热数据统计
	hotCount, _ := ts.redis.DBSize(ctx).Result()

	var warmCount, coldCount int64
	ts.db.Table("warm_data").Count(&warmCount)
	ts.db.Table("cold_data").Count(&coldCount)

	ts.metrics.mu.Lock()
	ts.metrics.HotDataCount = hotCount
	ts.metrics.WarmDataCount = warmCount
	ts.metrics.ColdDataCount = coldCount
	ts.metrics.mu.Unlock()
}

// WarmData 温数据表结构
type WarmData struct {
	Key       string    `gorm:"primaryKey;size:255"`
	Data      []byte    `gorm:"type:bytea"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// ColdData 冷数据表结构
type ColdData struct {
	Key       string    `gorm:"primaryKey;size:255"`
	Data      []byte    `gorm:"type:bytea"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// TableName 指定表名
func (WarmData) TableName() string {
	return "warm_data"
}

func (ColdData) TableName() string {
	return "cold_data"
}

// AutoMigrate 自动迁移表结构
func (ts *TieredStorage) AutoMigrate() error {
	return ts.db.AutoMigrate(&WarmData{}, &ColdData{})
}
