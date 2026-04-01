package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"gorm.io/gorm"
)

// AlertStorage 告警存储接口
type AlertStorage interface {
	// 实时存储操作
	StoreActive(ctx context.Context, alarm *entity.Alarm) error
	GetActive(ctx context.Context, id string) (*entity.Alarm, error)
	GetActiveByPoint(ctx context.Context, pointID string) (*entity.Alarm, error)
	ListActive(ctx context.Context, opts ListActiveOptions) ([]*entity.Alarm, error)
	DeleteActive(ctx context.Context, id string) error
	CountActive(ctx context.Context, opts CountOptions) (int64, error)

	// 历史存储操作
	StoreHistory(ctx context.Context, alarm *entity.Alarm) error
	GetHistory(ctx context.Context, id string) (*entity.Alarm, error)
	ListHistory(ctx context.Context, opts ListHistoryOptions) ([]*entity.Alarm, error)
	DeleteHistory(ctx context.Context, id string) error

	// 状态更新
	UpdateStatus(ctx context.Context, id string, status entity.AlarmStatus, opts UpdateStatusOptions) error
	Acknowledge(ctx context.Context, id string, by string) error
	Clear(ctx context.Context, id string) error

	// 统计
	CountByLevel(ctx context.Context, stationID *string) (map[entity.AlarmLevel]int64, error)
	CountByType(ctx context.Context, stationID *string) (map[entity.AlarmType]int64, error)
}

// ListActiveOptions 活跃告警列表选项
type ListActiveOptions struct {
	StationID string
	DeviceID  string
	PointID   string
	Level     *entity.AlarmLevel
	Type      *entity.AlarmType
	Status    *entity.AlarmStatus
	Limit     int
	Offset    int
}

// ListHistoryOptions 历史告警列表选项
type ListHistoryOptions struct {
	StationID string
	DeviceID  string
	PointID   string
	Level     *entity.AlarmLevel
	Type      *entity.AlarmType
	Status    *entity.AlarmStatus
	StartTime int64
	EndTime   int64
	Limit     int
	Offset    int
	OrderBy   string
	OrderDesc bool
}

// CountOptions 计数选项
type CountOptions struct {
	StationID string
	DeviceID  string
	Level     *entity.AlarmLevel
	Type      *entity.AlarmType
}

// UpdateStatusOptions 更新状态选项
type UpdateStatusOptions struct {
	AcknowledgedBy string
	Reason         string
}

// RedisAlertStorage Redis告警存储
type RedisAlertStorage struct {
	client    *redis.Client
	keyPrefix string
	mu        sync.RWMutex
}

// RedisAlertStorageConfig Redis告警存储配置
type RedisAlertStorageConfig struct {
	Client    *redis.Client
	KeyPrefix string
}

// NewRedisAlertStorage 创建Redis告警存储
func NewRedisAlertStorage(cfg RedisAlertStorageConfig) *RedisAlertStorage {
	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = "nem:alarm"
	}
	return &RedisAlertStorage{
		client:    cfg.Client,
		keyPrefix: cfg.KeyPrefix,
	}
}

// activeKey 生成活跃告警key
func (s *RedisAlertStorage) activeKey(id string) string {
	return fmt.Sprintf("%s:active:%s", s.keyPrefix, id)
}

// activeByPointKey 生成测点活跃告警key
func (s *RedisAlertStorage) activeByPointKey(pointID string) string {
	return fmt.Sprintf("%s:active:point:%s", s.keyPrefix, pointID)
}

// activeListKey 生成活跃告警列表key
func (s *RedisAlertStorage) activeListKey() string {
	return fmt.Sprintf("%s:active:list", s.keyPrefix)
}

// activeByStationKey 生成站点活跃告警集合key
func (s *RedisAlertStorage) activeByStationKey(stationID string) string {
	return fmt.Sprintf("%s:active:station:%s", s.keyPrefix, stationID)
}

// activeByDeviceKey 生成设备活跃告警集合key
func (s *RedisAlertStorage) activeByDeviceKey(deviceID string) string {
	return fmt.Sprintf("%s:active:device:%s", s.keyPrefix, deviceID)
}

// StoreActive 存储活跃告警
func (s *RedisAlertStorage) StoreActive(ctx context.Context, alarm *entity.Alarm) error {
	if alarm.ID == "" {
		alarm.ID = uuid.New().String()
	}

	data, err := json.Marshal(alarm)
	if err != nil {
		return fmt.Errorf("failed to marshal alarm: %w", err)
	}

	pipe := s.client.Pipeline()

	// 存储告警详情
	pipe.Set(ctx, s.activeKey(alarm.ID), data, 0)

	// 添加到活跃告警列表
	pipe.ZAdd(ctx, s.activeListKey(), &redis.Z{
		Score:  float64(alarm.TriggeredAt.Unix()),
		Member: alarm.ID,
	})

	// 添加到测点索引
	if alarm.PointID != "" {
		pipe.Set(ctx, s.activeByPointKey(alarm.PointID), alarm.ID, 0)
	}

	// 添加到站点索引
	if alarm.StationID != "" {
		pipe.SAdd(ctx, s.activeByStationKey(alarm.StationID), alarm.ID)
	}

	// 添加到设备索引
	if alarm.DeviceID != "" {
		pipe.SAdd(ctx, s.activeByDeviceKey(alarm.DeviceID), alarm.ID)
	}

	_, err = pipe.Exec(ctx)
	return err
}

// GetActive 获取活跃告警
func (s *RedisAlertStorage) GetActive(ctx context.Context, id string) (*entity.Alarm, error) {
	data, err := s.client.Get(ctx, s.activeKey(id)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var alarm entity.Alarm
	if err := json.Unmarshal(data, &alarm); err != nil {
		return nil, fmt.Errorf("failed to unmarshal alarm: %w", err)
	}

	return &alarm, nil
}

// GetActiveByPoint 通过测点获取活跃告警
func (s *RedisAlertStorage) GetActiveByPoint(ctx context.Context, pointID string) (*entity.Alarm, error) {
	alarmID, err := s.client.Get(ctx, s.activeByPointKey(pointID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	return s.GetActive(ctx, alarmID)
}

// ListActive 列出活跃告警
func (s *RedisAlertStorage) ListActive(ctx context.Context, opts ListActiveOptions) ([]*entity.Alarm, error) {
	var alarmIDs []string

	if opts.StationID != "" {
		// 从站点索引获取
		members, err := s.client.SMembers(ctx, s.activeByStationKey(opts.StationID)).Result()
		if err != nil {
			return nil, err
		}
		alarmIDs = members
	} else if opts.DeviceID != "" {
		// 从设备索引获取
		members, err := s.client.SMembers(ctx, s.activeByDeviceKey(opts.DeviceID)).Result()
		if err != nil {
			return nil, err
		}
		alarmIDs = members
	} else {
		// 从全局列表获取
		var start, stop int64
		if opts.Offset > 0 {
			start = int64(opts.Offset)
		}
		if opts.Limit > 0 {
			stop = start + int64(opts.Limit) - 1
		} else {
			stop = -1
		}

		members, err := s.client.ZRevRange(ctx, s.activeListKey(), start, stop).Result()
		if err != nil {
			return nil, err
		}
		alarmIDs = members
	}

	alarms := make([]*entity.Alarm, 0, len(alarmIDs))
	for _, id := range alarmIDs {
		alarm, err := s.GetActive(ctx, id)
		if err != nil {
			continue
		}
		if alarm == nil {
			continue
		}

		// 过滤条件
		if opts.Level != nil && alarm.Level != *opts.Level {
			continue
		}
		if opts.Type != nil && alarm.Type != *opts.Type {
			continue
		}
		if opts.Status != nil && alarm.Status != *opts.Status {
			continue
		}
		if opts.PointID != "" && alarm.PointID != opts.PointID {
			continue
		}

		alarms = append(alarms, alarm)
	}

	return alarms, nil
}

// DeleteActive 删除活跃告警
func (s *RedisAlertStorage) DeleteActive(ctx context.Context, id string) error {
	// 先获取告警信息
	alarm, err := s.GetActive(ctx, id)
	if err != nil {
		return err
	}
	if alarm == nil {
		return nil
	}

	pipe := s.client.Pipeline()

	// 删除告警详情
	pipe.Del(ctx, s.activeKey(id))

	// 从活跃列表删除
	pipe.ZRem(ctx, s.activeListKey(), id)

	// 从测点索引删除
	if alarm.PointID != "" {
		pipe.Del(ctx, s.activeByPointKey(alarm.PointID))
	}

	// 从站点索引删除
	if alarm.StationID != "" {
		pipe.SRem(ctx, s.activeByStationKey(alarm.StationID), id)
	}

	// 从设备索引删除
	if alarm.DeviceID != "" {
		pipe.SRem(ctx, s.activeByDeviceKey(alarm.DeviceID), id)
	}

	_, err = pipe.Exec(ctx)
	return err
}

// CountActive 计数活跃告警
func (s *RedisAlertStorage) CountActive(ctx context.Context, opts CountOptions) (int64, error) {
	if opts.StationID != "" {
		return s.client.SCard(ctx, s.activeByStationKey(opts.StationID)).Result()
	}
	if opts.DeviceID != "" {
		return s.client.SCard(ctx, s.activeByDeviceKey(opts.DeviceID)).Result()
	}
	return s.client.ZCard(ctx, s.activeListKey()).Result()
}

// UpdateStatus 更新状态
func (s *RedisAlertStorage) UpdateStatus(ctx context.Context, id string, status entity.AlarmStatus, opts UpdateStatusOptions) error {
	alarm, err := s.GetActive(ctx, id)
	if err != nil {
		return err
	}
	if alarm == nil {
		return errors.New("alarm not found")
	}

	alarm.Status = status
	now := time.Now()

	switch status {
	case entity.AlarmStatusAcknowledged:
		alarm.AcknowledgedAt = &now
		alarm.AcknowledgedBy = opts.AcknowledgedBy
	case entity.AlarmStatusCleared:
		alarm.ClearedAt = &now
	}

	return s.StoreActive(ctx, alarm)
}

// Acknowledge 确认告警
func (s *RedisAlertStorage) Acknowledge(ctx context.Context, id string, by string) error {
	return s.UpdateStatus(ctx, id, entity.AlarmStatusAcknowledged, UpdateStatusOptions{
		AcknowledgedBy: by,
	})
}

// Clear 清除告警
func (s *RedisAlertStorage) Clear(ctx context.Context, id string) error {
	return s.UpdateStatus(ctx, id, entity.AlarmStatusCleared, UpdateStatusOptions{})
}

// StoreHistory 存储历史告警（Redis不实现）
func (s *RedisAlertStorage) StoreHistory(ctx context.Context, alarm *entity.Alarm) error {
	return errors.New("redis storage does not support history storage")
}

// GetHistory 获取历史告警（Redis不实现）
func (s *RedisAlertStorage) GetHistory(ctx context.Context, id string) (*entity.Alarm, error) {
	return nil, errors.New("redis storage does not support history storage")
}

// ListHistory 列出历史告警（Redis不实现）
func (s *RedisAlertStorage) ListHistory(ctx context.Context, opts ListHistoryOptions) ([]*entity.Alarm, error) {
	return nil, errors.New("redis storage does not support history storage")
}

// DeleteHistory 删除历史告警（Redis不实现）
func (s *RedisAlertStorage) DeleteHistory(ctx context.Context, id string) error {
	return errors.New("redis storage does not support history storage")
}

// CountByLevel 按级别计数
func (s *RedisAlertStorage) CountByLevel(ctx context.Context, stationID *string) (map[entity.AlarmLevel]int64, error) {
	alarms, err := s.ListActive(ctx, ListActiveOptions{
		StationID: func() string { if stationID != nil { return *stationID }; return "" }(),
	})
	if err != nil {
		return nil, err
	}

	counts := make(map[entity.AlarmLevel]int64)
	for _, alarm := range alarms {
		counts[alarm.Level]++
	}
	return counts, nil
}

// CountByType 按类型计数
func (s *RedisAlertStorage) CountByType(ctx context.Context, stationID *string) (map[entity.AlarmType]int64, error) {
	alarms, err := s.ListActive(ctx, ListActiveOptions{
		StationID: func() string { if stationID != nil { return *stationID }; return "" }(),
	})
	if err != nil {
		return nil, err
	}

	counts := make(map[entity.AlarmType]int64)
	for _, alarm := range alarms {
		counts[alarm.Type]++
	}
	return counts, nil
}

// PostgresAlertStorage PostgreSQL告警存储
type PostgresAlertStorage struct {
	db *gorm.DB
}

// NewPostgresAlertStorage 创建PostgreSQL告警存储
func NewPostgresAlertStorage(db *gorm.DB) *PostgresAlertStorage {
	return &PostgresAlertStorage{db: db}
}

// StoreActive 存储活跃告警
func (s *PostgresAlertStorage) StoreActive(ctx context.Context, alarm *entity.Alarm) error {
	if alarm.ID == "" {
		alarm.ID = uuid.New().String()
	}
	return s.db.WithContext(ctx).Create(alarm).Error
}

// GetActive 获取活跃告警
func (s *PostgresAlertStorage) GetActive(ctx context.Context, id string) (*entity.Alarm, error) {
	var alarm entity.Alarm
	err := s.db.WithContext(ctx).First(&alarm, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &alarm, nil
}

// GetActiveByPoint 通过测点获取活跃告警
func (s *PostgresAlertStorage) GetActiveByPoint(ctx context.Context, pointID string) (*entity.Alarm, error) {
	var alarm entity.Alarm
	err := s.db.WithContext(ctx).
		Where("point_id = ? AND status IN ?", pointID, []entity.AlarmStatus{
			entity.AlarmStatusActive,
			entity.AlarmStatusAcknowledged,
			entity.AlarmStatusSuppressed,
		}).
		First(&alarm).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &alarm, nil
}

// ListActive 列出活跃告警
func (s *PostgresAlertStorage) ListActive(ctx context.Context, opts ListActiveOptions) ([]*entity.Alarm, error) {
	query := s.db.WithContext(ctx).Model(&entity.Alarm{}).
		Where("status IN ?", []entity.AlarmStatus{
			entity.AlarmStatusActive,
			entity.AlarmStatusAcknowledged,
			entity.AlarmStatusSuppressed,
		})

	if opts.StationID != "" {
		query = query.Where("station_id = ?", opts.StationID)
	}
	if opts.DeviceID != "" {
		query = query.Where("device_id = ?", opts.DeviceID)
	}
	if opts.PointID != "" {
		query = query.Where("point_id = ?", opts.PointID)
	}
	if opts.Level != nil {
		query = query.Where("level = ?", *opts.Level)
	}
	if opts.Type != nil {
		query = query.Where("type = ?", *opts.Type)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}

	query = query.Order("triggered_at DESC")

	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}

	var alarms []*entity.Alarm
	err := query.Find(&alarms).Error
	return alarms, err
}

// DeleteActive 删除活跃告警
func (s *PostgresAlertStorage) DeleteActive(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&entity.Alarm{}, "id = ?", id).Error
}

// CountActive 计数活跃告警
func (s *PostgresAlertStorage) CountActive(ctx context.Context, opts CountOptions) (int64, error) {
	query := s.db.WithContext(ctx).Model(&entity.Alarm{}).
		Where("status IN ?", []entity.AlarmStatus{
			entity.AlarmStatusActive,
			entity.AlarmStatusAcknowledged,
			entity.AlarmStatusSuppressed,
		})

	if opts.StationID != "" {
		query = query.Where("station_id = ?", opts.StationID)
	}
	if opts.DeviceID != "" {
		query = query.Where("device_id = ?", opts.DeviceID)
	}
	if opts.Level != nil {
		query = query.Where("level = ?", *opts.Level)
	}
	if opts.Type != nil {
		query = query.Where("type = ?", *opts.Type)
	}

	var count int64
	err := query.Count(&count).Error
	return count, err
}

// StoreHistory 存储历史告警
func (s *PostgresAlertStorage) StoreHistory(ctx context.Context, alarm *entity.Alarm) error {
	if alarm.ID == "" {
		alarm.ID = uuid.New().String()
	}
	return s.db.WithContext(ctx).Create(alarm).Error
}

// GetHistory 获取历史告警
func (s *PostgresAlertStorage) GetHistory(ctx context.Context, id string) (*entity.Alarm, error) {
	var alarm entity.Alarm
	err := s.db.WithContext(ctx).First(&alarm, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &alarm, nil
}

// ListHistory 列出历史告警
func (s *PostgresAlertStorage) ListHistory(ctx context.Context, opts ListHistoryOptions) ([]*entity.Alarm, error) {
	query := s.db.WithContext(ctx).Model(&entity.Alarm{})

	if opts.StationID != "" {
		query = query.Where("station_id = ?", opts.StationID)
	}
	if opts.DeviceID != "" {
		query = query.Where("device_id = ?", opts.DeviceID)
	}
	if opts.PointID != "" {
		query = query.Where("point_id = ?", opts.PointID)
	}
	if opts.Level != nil {
		query = query.Where("level = ?", *opts.Level)
	}
	if opts.Type != nil {
		query = query.Where("type = ?", *opts.Type)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.StartTime > 0 {
		query = query.Where("triggered_at >= ?", time.Unix(opts.StartTime, 0))
	}
	if opts.EndTime > 0 {
		query = query.Where("triggered_at <= ?", time.Unix(opts.EndTime, 0))
	}

	orderBy := "triggered_at"
	if opts.OrderBy != "" {
		orderBy = opts.OrderBy
	}
	if opts.OrderDesc {
		query = query.Order(orderBy + " DESC")
	} else {
		query = query.Order(orderBy + " ASC")
	}

	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}

	var alarms []*entity.Alarm
	err := query.Find(&alarms).Error
	return alarms, err
}

// DeleteHistory 删除历史告警
func (s *PostgresAlertStorage) DeleteHistory(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&entity.Alarm{}, "id = ?", id).Error
}

// UpdateStatus 更新状态
func (s *PostgresAlertStorage) UpdateStatus(ctx context.Context, id string, status entity.AlarmStatus, opts UpdateStatusOptions) error {
	updates := map[string]interface{}{
		"status": status,
	}

	now := time.Now()
	switch status {
	case entity.AlarmStatusAcknowledged:
		updates["acknowledged_at"] = &now
		updates["acknowledged_by"] = opts.AcknowledgedBy
	case entity.AlarmStatusCleared:
		updates["cleared_at"] = &now
	}

	return s.db.WithContext(ctx).Model(&entity.Alarm{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// Acknowledge 确认告警
func (s *PostgresAlertStorage) Acknowledge(ctx context.Context, id string, by string) error {
	return s.UpdateStatus(ctx, id, entity.AlarmStatusAcknowledged, UpdateStatusOptions{
		AcknowledgedBy: by,
	})
}

// Clear 清除告警
func (s *PostgresAlertStorage) Clear(ctx context.Context, id string) error {
	return s.UpdateStatus(ctx, id, entity.AlarmStatusCleared, UpdateStatusOptions{})
}

// CountByLevel 按级别计数
func (s *PostgresAlertStorage) CountByLevel(ctx context.Context, stationID *string) (map[entity.AlarmLevel]int64, error) {
	query := s.db.WithContext(ctx).Model(&entity.Alarm{}).
		Select("level, count(*) as count").
		Group("level")

	if stationID != nil {
		query = query.Where("station_id = ?", *stationID)
	}

	type countResult struct {
		Level entity.AlarmLevel
		Count int64
	}

	var results []countResult
	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	counts := make(map[entity.AlarmLevel]int64)
	for _, r := range results {
		counts[r.Level] = r.Count
	}
	return counts, nil
}

// CountByType 按类型计数
func (s *PostgresAlertStorage) CountByType(ctx context.Context, stationID *string) (map[entity.AlarmType]int64, error) {
	query := s.db.WithContext(ctx).Model(&entity.Alarm{}).
		Select("type, count(*) as count").
		Group("type")

	if stationID != nil {
		query = query.Where("station_id = ?", *stationID)
	}

	type countResult struct {
		Type  entity.AlarmType
		Count int64
	}

	var results []countResult
	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	counts := make(map[entity.AlarmType]int64)
	for _, r := range results {
		counts[r.Type] = r.Count
	}
	return counts, nil
}

// HybridAlertStorage 混合存储（Redis + PostgreSQL）
type HybridAlertStorage struct {
	redis    *RedisAlertStorage
	postgres *PostgresAlertStorage
}

// NewHybridAlertStorage 创建混合存储
func NewHybridAlertStorage(redis *RedisAlertStorage, postgres *PostgresAlertStorage) *HybridAlertStorage {
	return &HybridAlertStorage{
		redis:    redis,
		postgres: postgres,
	}
}

// StoreActive 存储活跃告警（同时存Redis和PostgreSQL）
func (s *HybridAlertStorage) StoreActive(ctx context.Context, alarm *entity.Alarm) error {
	// 先存PostgreSQL
	if err := s.postgres.StoreActive(ctx, alarm); err != nil {
		return fmt.Errorf("failed to store in postgres: %w", err)
	}

	// 再存Redis
	if err := s.redis.StoreActive(ctx, alarm); err != nil {
		// Redis失败不影响主流程，记录日志即可
		fmt.Printf("warning: failed to store in redis: %v\n", err)
	}

	return nil
}

// GetActive 获取活跃告警（优先从Redis获取）
func (s *HybridAlertStorage) GetActive(ctx context.Context, id string) (*entity.Alarm, error) {
	// 优先从Redis获取
	alarm, err := s.redis.GetActive(ctx, id)
	if err != nil {
		fmt.Printf("warning: failed to get from redis: %v\n", err)
	}
	if alarm != nil {
		return alarm, nil
	}

	// Redis没有，从PostgreSQL获取
	return s.postgres.GetActive(ctx, id)
}

// GetActiveByPoint 通过测点获取活跃告警
func (s *HybridAlertStorage) GetActiveByPoint(ctx context.Context, pointID string) (*entity.Alarm, error) {
	alarm, err := s.redis.GetActiveByPoint(ctx, pointID)
	if err != nil {
		fmt.Printf("warning: failed to get from redis: %v\n", err)
	}
	if alarm != nil {
		return alarm, nil
	}
	return s.postgres.GetActiveByPoint(ctx, pointID)
}

// ListActive 列出活跃告警
func (s *HybridAlertStorage) ListActive(ctx context.Context, opts ListActiveOptions) ([]*entity.Alarm, error) {
	// 优先从Redis获取
	alarms, err := s.redis.ListActive(ctx, opts)
	if err != nil {
		fmt.Printf("warning: failed to list from redis: %v\n", err)
	}
	if len(alarms) > 0 {
		return alarms, nil
	}
	return s.postgres.ListActive(ctx, opts)
}

// DeleteActive 删除活跃告警
func (s *HybridAlertStorage) DeleteActive(ctx context.Context, id string) error {
	// 先从PostgreSQL删除
	if err := s.postgres.DeleteActive(ctx, id); err != nil {
		return err
	}
	// 再从Redis删除
	return s.redis.DeleteActive(ctx, id)
}

// CountActive 计数活跃告警
func (s *HybridAlertStorage) CountActive(ctx context.Context, opts CountOptions) (int64, error) {
	count, err := s.redis.CountActive(ctx, opts)
	if err != nil {
		return s.postgres.CountActive(ctx, opts)
	}
	return count, nil
}

// StoreHistory 存储历史告警
func (s *HybridAlertStorage) StoreHistory(ctx context.Context, alarm *entity.Alarm) error {
	return s.postgres.StoreHistory(ctx, alarm)
}

// GetHistory 获取历史告警
func (s *HybridAlertStorage) GetHistory(ctx context.Context, id string) (*entity.Alarm, error) {
	return s.postgres.GetHistory(ctx, id)
}

// ListHistory 列出历史告警
func (s *HybridAlertStorage) ListHistory(ctx context.Context, opts ListHistoryOptions) ([]*entity.Alarm, error) {
	return s.postgres.ListHistory(ctx, opts)
}

// DeleteHistory 删除历史告警
func (s *HybridAlertStorage) DeleteHistory(ctx context.Context, id string) error {
	return s.postgres.DeleteHistory(ctx, id)
}

// UpdateStatus 更新状态
func (s *HybridAlertStorage) UpdateStatus(ctx context.Context, id string, status entity.AlarmStatus, opts UpdateStatusOptions) error {
	// 更新PostgreSQL
	if err := s.postgres.UpdateStatus(ctx, id, status, opts); err != nil {
		return err
	}
	// 更新Redis
	return s.redis.UpdateStatus(ctx, id, status, opts)
}

// Acknowledge 确认告警
func (s *HybridAlertStorage) Acknowledge(ctx context.Context, id string, by string) error {
	if err := s.postgres.Acknowledge(ctx, id, by); err != nil {
		return err
	}
	return s.redis.Acknowledge(ctx, id, by)
}

// Clear 清除告警
func (s *HybridAlertStorage) Clear(ctx context.Context, id string) error {
	if err := s.postgres.Clear(ctx, id); err != nil {
		return err
	}
	return s.redis.Clear(ctx, id)
}

// CountByLevel 按级别计数
func (s *HybridAlertStorage) CountByLevel(ctx context.Context, stationID *string) (map[entity.AlarmLevel]int64, error) {
	return s.postgres.CountByLevel(ctx, stationID)
}

// CountByType 按类型计数
func (s *HybridAlertStorage) CountByType(ctx context.Context, stationID *string) (map[entity.AlarmType]int64, error) {
	return s.postgres.CountByType(ctx, stationID)
}

// MoveToHistory 将告警移到历史
func (s *HybridAlertStorage) MoveToHistory(ctx context.Context, id string) error {
	// 获取活跃告警
	alarm, err := s.GetActive(ctx, id)
	if err != nil {
		return err
	}
	if alarm == nil {
		return errors.New("alarm not found")
	}

	// 从Redis删除活跃告警
	if err := s.redis.DeleteActive(ctx, id); err != nil {
		fmt.Printf("warning: failed to delete from redis: %v\n", err)
	}

	// PostgreSQL中的记录保留作为历史记录
	return nil
}

// SyncFromPostgres 从PostgreSQL同步活跃告警到Redis
func (s *HybridAlertStorage) SyncFromPostgres(ctx context.Context) error {
	alarms, err := s.postgres.ListActive(ctx, ListActiveOptions{})
	if err != nil {
		return err
	}

	for _, alarm := range alarms {
		if err := s.redis.StoreActive(ctx, alarm); err != nil {
			fmt.Printf("warning: failed to sync alarm %s to redis: %v\n", alarm.ID, err)
		}
	}

	return nil
}
