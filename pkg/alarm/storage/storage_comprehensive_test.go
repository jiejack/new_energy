package storage

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupRedisStorageForComprehensive(t *testing.T) (*miniredis.Miniredis, *RedisAlertStorage) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	storage := NewRedisAlertStorage(RedisAlertStorageConfig{
		Client: client,
	})
	t.Cleanup(func() { client.Close() })
	return mr, storage
}

func setupPostgresStorageForComprehensive(t *testing.T) *PostgresAlertStorage {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	err = db.AutoMigrate(&entity.Alarm{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { sqlDB.Close() })
	return NewPostgresAlertStorage(db)
}

func setupHybridStorageForComprehensive(t *testing.T) (*miniredis.Miniredis, *HybridAlertStorage) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { client.Close() })

	redisStorage := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	pgStorage := setupPostgresStorageForComprehensive(t)
	hybrid := NewHybridAlertStorage(redisStorage, pgStorage)
	return mr, hybrid
}

func TestRedisAlertStorage_StoreActive_NoPointID(t *testing.T) {
	_, storage := setupRedisStorageForComprehensive(t)
	ctx := context.Background()

	alarm := &entity.Alarm{
		DeviceID:    "dev1",
		StationID:   "st1",
		Type:        entity.AlarmTypeLimit,
		Level:       entity.AlarmLevelWarning,
		Title:       "No Point",
		Status:      entity.AlarmStatusActive,
		TriggeredAt: time.Now(),
	}
	err := storage.StoreActive(ctx, alarm)
	require.NoError(t, err)
	assert.NotEmpty(t, alarm.ID)
}

func TestRedisAlertStorage_StoreActive_NoStationID(t *testing.T) {
	_, storage := setupRedisStorageForComprehensive(t)
	ctx := context.Background()

	alarm := &entity.Alarm{
		PointID:     "pt1",
		DeviceID:    "dev1",
		Type:        entity.AlarmTypeLimit,
		Level:       entity.AlarmLevelWarning,
		Title:       "No Station",
		Status:      entity.AlarmStatusActive,
		TriggeredAt: time.Now(),
	}
	err := storage.StoreActive(ctx, alarm)
	require.NoError(t, err)
}

func TestRedisAlertStorage_StoreActive_NoDeviceID(t *testing.T) {
	_, storage := setupRedisStorageForComprehensive(t)
	ctx := context.Background()

	alarm := &entity.Alarm{
		PointID:     "pt1",
		StationID:   "st1",
		Type:        entity.AlarmTypeLimit,
		Level:       entity.AlarmLevelWarning,
		Title:       "No Device",
		Status:      entity.AlarmStatusActive,
		TriggeredAt: time.Now(),
	}
	err := storage.StoreActive(ctx, alarm)
	require.NoError(t, err)
}

func TestRedisAlertStorage_StoreActive_WithID(t *testing.T) {
	_, storage := setupRedisStorageForComprehensive(t)
	ctx := context.Background()

	alarm := &entity.Alarm{
		ID:          "custom-id",
		PointID:     "pt1",
		DeviceID:    "dev1",
		StationID:   "st1",
		Type:        entity.AlarmTypeLimit,
		Level:       entity.AlarmLevelWarning,
		Title:       "Custom ID",
		Status:      entity.AlarmStatusActive,
		TriggeredAt: time.Now(),
	}
	err := storage.StoreActive(ctx, alarm)
	require.NoError(t, err)
	assert.Equal(t, "custom-id", alarm.ID)
}

func TestRedisAlertStorage_ListActive_ByDevice(t *testing.T) {
	_, storage := setupRedisStorageForComprehensive(t)
	ctx := context.Background()

	alarm1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm2 := entity.NewAlarm("pt2", "dev2", "st1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	storage.StoreActive(ctx, alarm1)
	storage.StoreActive(ctx, alarm2)

	alarms, err := storage.ListActive(ctx, ListActiveOptions{DeviceID: "dev1"})
	require.NoError(t, err)
	assert.Equal(t, 1, len(alarms))
	assert.Equal(t, "dev1", alarms[0].DeviceID)
}

func TestRedisAlertStorage_ListActive_WithLevelFilter(t *testing.T) {
	_, storage := setupRedisStorageForComprehensive(t)
	ctx := context.Background()

	alarm1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm2 := entity.NewAlarm("pt2", "dev2", "st1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	storage.StoreActive(ctx, alarm1)
	storage.StoreActive(ctx, alarm2)

	level := entity.AlarmLevelCritical
	alarms, err := storage.ListActive(ctx, ListActiveOptions{Level: &level})
	require.NoError(t, err)
	assert.Equal(t, 1, len(alarms))
	assert.Equal(t, entity.AlarmLevelCritical, alarms[0].Level)
}

func TestRedisAlertStorage_ListActive_WithTypeFilter(t *testing.T) {
	_, storage := setupRedisStorageForComprehensive(t)
	ctx := context.Background()

	alarm1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm2 := entity.NewAlarm("pt2", "dev2", "st1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	storage.StoreActive(ctx, alarm1)
	storage.StoreActive(ctx, alarm2)

	alarmType := entity.AlarmTypeLimit
	alarms, err := storage.ListActive(ctx, ListActiveOptions{Type: &alarmType})
	require.NoError(t, err)
	assert.Equal(t, 1, len(alarms))
	assert.Equal(t, entity.AlarmTypeLimit, alarms[0].Type)
}

func TestRedisAlertStorage_ListActive_WithStatusFilter(t *testing.T) {
	_, storage := setupRedisStorageForComprehensive(t)
	ctx := context.Background()

	alarm1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	storage.StoreActive(ctx, alarm1)

	status := entity.AlarmStatusActive
	alarms, err := storage.ListActive(ctx, ListActiveOptions{Status: &status})
	require.NoError(t, err)
	assert.Equal(t, 1, len(alarms))
}

func TestRedisAlertStorage_ListActive_WithPointIDFilter(t *testing.T) {
	_, storage := setupRedisStorageForComprehensive(t)
	ctx := context.Background()

	alarm1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm2 := entity.NewAlarm("pt2", "dev2", "st1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	storage.StoreActive(ctx, alarm1)
	storage.StoreActive(ctx, alarm2)

	alarms, err := storage.ListActive(ctx, ListActiveOptions{PointID: "pt1"})
	require.NoError(t, err)
	assert.Equal(t, 1, len(alarms))
	assert.Equal(t, "pt1", alarms[0].PointID)
}

func TestRedisAlertStorage_ListActive_WithLimit(t *testing.T) {
	_, storage := setupRedisStorageForComprehensive(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A", "")
		storage.StoreActive(ctx, alarm)
	}

	alarms, err := storage.ListActive(ctx, ListActiveOptions{Limit: 3})
	require.NoError(t, err)
	assert.LessOrEqual(t, len(alarms), 3)
}

func TestRedisAlertStorage_ListActive_WithOffset(t *testing.T) {
	_, storage := setupRedisStorageForComprehensive(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A", "")
		storage.StoreActive(ctx, alarm)
	}

	alarms, err := storage.ListActive(ctx, ListActiveOptions{Offset: 2})
	require.NoError(t, err)
	assert.LessOrEqual(t, len(alarms), 3)
}

func TestRedisAlertStorage_CountActive_ByDevice(t *testing.T) {
	_, storage := setupRedisStorageForComprehensive(t)
	ctx := context.Background()

	alarm1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm2 := entity.NewAlarm("pt2", "dev2", "st1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	storage.StoreActive(ctx, alarm1)
	storage.StoreActive(ctx, alarm2)

	count, err := storage.CountActive(ctx, CountOptions{DeviceID: "dev1"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestRedisAlertStorage_UpdateStatus_NotFound(t *testing.T) {
	_, storage := setupRedisStorageForComprehensive(t)
	ctx := context.Background()

	err := storage.UpdateStatus(ctx, "nonexistent", entity.AlarmStatusAcknowledged, UpdateStatusOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRedisAlertStorage_UpdateStatus_Cleared(t *testing.T) {
	_, storage := setupRedisStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	storage.StoreActive(ctx, alarm)

	err := storage.UpdateStatus(ctx, alarm.ID, entity.AlarmStatusCleared, UpdateStatusOptions{})
	require.NoError(t, err)

	got, _ := storage.GetActive(ctx, alarm.ID)
	assert.Equal(t, entity.AlarmStatusCleared, got.Status)
	assert.NotNil(t, got.ClearedAt)
}

func TestRedisAlertStorage_CountByLevel_WithStationID(t *testing.T) {
	_, storage := setupRedisStorageForComprehensive(t)
	ctx := context.Background()

	alarm1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm2 := entity.NewAlarm("pt2", "dev2", "st2", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	storage.StoreActive(ctx, alarm1)
	storage.StoreActive(ctx, alarm2)

	stationID := "st1"
	counts, err := storage.CountByLevel(ctx, &stationID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts[entity.AlarmLevelWarning])
}

func TestRedisAlertStorage_CountByType_WithStationID(t *testing.T) {
	_, storage := setupRedisStorageForComprehensive(t)
	ctx := context.Background()

	alarm1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm2 := entity.NewAlarm("pt2", "dev2", "st2", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	storage.StoreActive(ctx, alarm1)
	storage.StoreActive(ctx, alarm2)

	stationID := "st1"
	counts, err := storage.CountByType(ctx, &stationID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts[entity.AlarmTypeLimit])
}

func TestRedisAlertStorage_CustomKeyPrefix(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	storage := NewRedisAlertStorage(RedisAlertStorageConfig{
		Client:    client,
		KeyPrefix: "custom:prefix",
	})
	assert.Equal(t, "custom:prefix", storage.keyPrefix)
	assert.Equal(t, "custom:prefix:active:test-id", storage.activeKey("test-id"))
	assert.Equal(t, "custom:prefix:active:point:pt1", storage.activeByPointKey("pt1"))
	assert.Equal(t, "custom:prefix:active:list", storage.activeListKey())
	assert.Equal(t, "custom:prefix:active:station:st1", storage.activeByStationKey("st1"))
	assert.Equal(t, "custom:prefix:active:device:dev1", storage.activeByDeviceKey("dev1"))
}

func TestRedisAlertStorage_KeyHelpers(t *testing.T) {
	_, storage := setupRedisStorageForComprehensive(t)

	assert.Equal(t, "nem:alarm:active:test-id", storage.activeKey("test-id"))
	assert.Equal(t, "nem:alarm:active:point:pt1", storage.activeByPointKey("pt1"))
	assert.Equal(t, "nem:alarm:active:list", storage.activeListKey())
	assert.Equal(t, "nem:alarm:active:station:st1", storage.activeByStationKey("st1"))
	assert.Equal(t, "nem:alarm:active:device:dev1", storage.activeByDeviceKey("dev1"))
}

func TestPostgresAlertStorage_StoreActive_GetActive(t *testing.T) {
	storage := setupPostgresStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	err := storage.StoreActive(ctx, alarm)
	require.NoError(t, err)
	assert.NotEmpty(t, alarm.ID)

	got, err := storage.GetActive(ctx, alarm.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "pt1", got.PointID)
}

func TestPostgresAlertStorage_GetActive_NotFound(t *testing.T) {
	storage := setupPostgresStorageForComprehensive(t)
	ctx := context.Background()

	got, err := storage.GetActive(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestPostgresAlertStorage_GetActiveByPoint(t *testing.T) {
	storage := setupPostgresStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	storage.StoreActive(ctx, alarm)

	got, err := storage.GetActiveByPoint(ctx, "pt1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "pt1", got.PointID)
}

func TestPostgresAlertStorage_GetActiveByPoint_NotFound(t *testing.T) {
	storage := setupPostgresStorageForComprehensive(t)
	ctx := context.Background()

	got, err := storage.GetActiveByPoint(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestPostgresAlertStorage_ListActive(t *testing.T) {
	storage := setupPostgresStorageForComprehensive(t)
	ctx := context.Background()

	alarm1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm2 := entity.NewAlarm("pt2", "dev2", "st1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	storage.StoreActive(ctx, alarm1)
	storage.StoreActive(ctx, alarm2)

	alarms, err := storage.ListActive(ctx, ListActiveOptions{})
	require.NoError(t, err)
	assert.Equal(t, 2, len(alarms))
}

func TestPostgresAlertStorage_ListActive_WithFilters(t *testing.T) {
	storage := setupPostgresStorageForComprehensive(t)
	ctx := context.Background()

	alarm1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm2 := entity.NewAlarm("pt2", "dev2", "st2", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	storage.StoreActive(ctx, alarm1)
	storage.StoreActive(ctx, alarm2)

	alarms, err := storage.ListActive(ctx, ListActiveOptions{StationID: "st1"})
	require.NoError(t, err)
	assert.Equal(t, 1, len(alarms))

	alarms, err = storage.ListActive(ctx, ListActiveOptions{DeviceID: "dev1"})
	require.NoError(t, err)
	assert.Equal(t, 1, len(alarms))

	alarms, err = storage.ListActive(ctx, ListActiveOptions{PointID: "pt1"})
	require.NoError(t, err)
	assert.Equal(t, 1, len(alarms))

	level := entity.AlarmLevelWarning
	alarms, err = storage.ListActive(ctx, ListActiveOptions{Level: &level})
	require.NoError(t, err)
	assert.Equal(t, 1, len(alarms))

	alarmType := entity.AlarmTypeLimit
	alarms, err = storage.ListActive(ctx, ListActiveOptions{Type: &alarmType})
	require.NoError(t, err)
	assert.Equal(t, 1, len(alarms))

	status := entity.AlarmStatusActive
	alarms, err = storage.ListActive(ctx, ListActiveOptions{Status: &status})
	require.NoError(t, err)
	assert.Equal(t, 2, len(alarms))

	alarms, err = storage.ListActive(ctx, ListActiveOptions{Limit: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, len(alarms))

	alarms, err = storage.ListActive(ctx, ListActiveOptions{Offset: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, len(alarms))
}

func TestPostgresAlertStorage_DeleteActive(t *testing.T) {
	storage := setupPostgresStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	storage.StoreActive(ctx, alarm)

	err := storage.DeleteActive(ctx, alarm.ID)
	require.NoError(t, err)

	got, err := storage.GetActive(ctx, alarm.ID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestPostgresAlertStorage_CountActive(t *testing.T) {
	storage := setupPostgresStorageForComprehensive(t)
	ctx := context.Background()

	alarm1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm2 := entity.NewAlarm("pt2", "dev2", "st1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	storage.StoreActive(ctx, alarm1)
	storage.StoreActive(ctx, alarm2)

	count, err := storage.CountActive(ctx, CountOptions{})
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	count, err = storage.CountActive(ctx, CountOptions{StationID: "st1"})
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	count, err = storage.CountActive(ctx, CountOptions{DeviceID: "dev1"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	level := entity.AlarmLevelWarning
	count, err = storage.CountActive(ctx, CountOptions{Level: &level})
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	alarmType := entity.AlarmTypeLimit
	count, err = storage.CountActive(ctx, CountOptions{Type: &alarmType})
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestPostgresAlertStorage_StoreHistory_GetHistory(t *testing.T) {
	storage := setupPostgresStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	alarm.Status = entity.AlarmStatusCleared
	err := storage.StoreHistory(ctx, alarm)
	require.NoError(t, err)

	got, err := storage.GetHistory(ctx, alarm.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "pt1", got.PointID)
}

func TestPostgresAlertStorage_GetHistory_NotFound(t *testing.T) {
	storage := setupPostgresStorageForComprehensive(t)
	ctx := context.Background()

	got, err := storage.GetHistory(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestPostgresAlertStorage_ListHistory(t *testing.T) {
	storage := setupPostgresStorageForComprehensive(t)
	ctx := context.Background()

	alarm1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm1.Status = entity.AlarmStatusCleared
	alarm2 := entity.NewAlarm("pt2", "dev2", "st2", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	alarm2.Status = entity.AlarmStatusCleared
	storage.StoreHistory(ctx, alarm1)
	storage.StoreHistory(ctx, alarm2)

	alarms, err := storage.ListHistory(ctx, ListHistoryOptions{})
	require.NoError(t, err)
	assert.Equal(t, 2, len(alarms))
}

func TestPostgresAlertStorage_ListHistory_WithFilters(t *testing.T) {
	storage := setupPostgresStorageForComprehensive(t)
	ctx := context.Background()

	alarm1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm1.Status = entity.AlarmStatusCleared
	alarm2 := entity.NewAlarm("pt2", "dev2", "st2", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	alarm2.Status = entity.AlarmStatusCleared
	storage.StoreHistory(ctx, alarm1)
	storage.StoreHistory(ctx, alarm2)

	alarms, err := storage.ListHistory(ctx, ListHistoryOptions{StationID: "st1"})
	require.NoError(t, err)
	assert.Equal(t, 1, len(alarms))

	alarms, err = storage.ListHistory(ctx, ListHistoryOptions{DeviceID: "dev1"})
	require.NoError(t, err)
	assert.Equal(t, 1, len(alarms))

	alarms, err = storage.ListHistory(ctx, ListHistoryOptions{PointID: "pt1"})
	require.NoError(t, err)
	assert.Equal(t, 1, len(alarms))

	level := entity.AlarmLevelWarning
	alarms, err = storage.ListHistory(ctx, ListHistoryOptions{Level: &level})
	require.NoError(t, err)
	assert.Equal(t, 1, len(alarms))

	alarmType := entity.AlarmTypeLimit
	alarms, err = storage.ListHistory(ctx, ListHistoryOptions{Type: &alarmType})
	require.NoError(t, err)
	assert.Equal(t, 1, len(alarms))

	status := entity.AlarmStatusCleared
	alarms, err = storage.ListHistory(ctx, ListHistoryOptions{Status: &status})
	require.NoError(t, err)
	assert.Equal(t, 2, len(alarms))

	now := time.Now()
	startTime := now.Add(-1 * time.Hour).Unix()
	endTime := now.Add(1 * time.Hour).Unix()
	alarms, err = storage.ListHistory(ctx, ListHistoryOptions{StartTime: startTime, EndTime: endTime})
	require.NoError(t, err)
	assert.Equal(t, 2, len(alarms))

	alarms, err = storage.ListHistory(ctx, ListHistoryOptions{OrderBy: "triggered_at", OrderDesc: true})
	require.NoError(t, err)
	assert.Equal(t, 2, len(alarms))

	alarms, err = storage.ListHistory(ctx, ListHistoryOptions{OrderBy: "triggered_at", OrderDesc: false})
	require.NoError(t, err)
	assert.Equal(t, 2, len(alarms))

	alarms, err = storage.ListHistory(ctx, ListHistoryOptions{Limit: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, len(alarms))

	alarms, err = storage.ListHistory(ctx, ListHistoryOptions{Offset: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, len(alarms))
}

func TestPostgresAlertStorage_DeleteHistory(t *testing.T) {
	storage := setupPostgresStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	storage.StoreHistory(ctx, alarm)

	err := storage.DeleteHistory(ctx, alarm.ID)
	require.NoError(t, err)

	got, err := storage.GetHistory(ctx, alarm.ID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestPostgresAlertStorage_UpdateStatus(t *testing.T) {
	storage := setupPostgresStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	storage.StoreActive(ctx, alarm)

	err := storage.UpdateStatus(ctx, alarm.ID, entity.AlarmStatusAcknowledged, UpdateStatusOptions{AcknowledgedBy: "admin"})
	require.NoError(t, err)

	got, _ := storage.GetActive(ctx, alarm.ID)
	assert.Equal(t, entity.AlarmStatusAcknowledged, got.Status)
}

func TestPostgresAlertStorage_UpdateStatus_Cleared(t *testing.T) {
	storage := setupPostgresStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	storage.StoreActive(ctx, alarm)

	err := storage.UpdateStatus(ctx, alarm.ID, entity.AlarmStatusCleared, UpdateStatusOptions{})
	require.NoError(t, err)
}

func TestPostgresAlertStorage_Acknowledge(t *testing.T) {
	storage := setupPostgresStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	storage.StoreActive(ctx, alarm)

	err := storage.Acknowledge(ctx, alarm.ID, "admin")
	require.NoError(t, err)
}

func TestPostgresAlertStorage_Clear(t *testing.T) {
	storage := setupPostgresStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	storage.StoreActive(ctx, alarm)

	err := storage.Clear(ctx, alarm.ID)
	require.NoError(t, err)
}

func TestPostgresAlertStorage_CountByLevel(t *testing.T) {
	storage := setupPostgresStorageForComprehensive(t)
	ctx := context.Background()

	alarm1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm2 := entity.NewAlarm("pt2", "dev2", "st1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	storage.StoreActive(ctx, alarm1)
	storage.StoreActive(ctx, alarm2)

	counts, err := storage.CountByLevel(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts[entity.AlarmLevelWarning])
	assert.Equal(t, int64(1), counts[entity.AlarmLevelCritical])

	stationID := "st1"
	counts, err = storage.CountByLevel(ctx, &stationID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), counts[entity.AlarmLevelWarning]+counts[entity.AlarmLevelCritical])
}

func TestPostgresAlertStorage_CountByType(t *testing.T) {
	storage := setupPostgresStorageForComprehensive(t)
	ctx := context.Background()

	alarm1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm2 := entity.NewAlarm("pt2", "dev2", "st1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	storage.StoreActive(ctx, alarm1)
	storage.StoreActive(ctx, alarm2)

	counts, err := storage.CountByType(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts[entity.AlarmTypeLimit])
	assert.Equal(t, int64(1), counts[entity.AlarmTypeStatus])

	stationID := "st1"
	counts, err = storage.CountByType(ctx, &stationID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), counts[entity.AlarmTypeLimit]+counts[entity.AlarmTypeStatus])
}

func TestHybridAlertStorage_StoreActive(t *testing.T) {
	_, hybrid := setupHybridStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	err := hybrid.StoreActive(ctx, alarm)
	require.NoError(t, err)
	assert.NotEmpty(t, alarm.ID)
}

func TestHybridAlertStorage_GetActive(t *testing.T) {
	_, hybrid := setupHybridStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	hybrid.StoreActive(ctx, alarm)

	got, err := hybrid.GetActive(ctx, alarm.ID)
	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "pt1", got.PointID)
}

func TestHybridAlertStorage_GetActiveByPoint(t *testing.T) {
	_, hybrid := setupHybridStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	hybrid.StoreActive(ctx, alarm)

	got, err := hybrid.GetActiveByPoint(ctx, "pt1")
	require.NoError(t, err)
	assert.NotNil(t, got)
}

func TestHybridAlertStorage_ListActive(t *testing.T) {
	_, hybrid := setupHybridStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	hybrid.StoreActive(ctx, alarm)

	alarms, err := hybrid.ListActive(ctx, ListActiveOptions{})
	require.NoError(t, err)
	assert.Equal(t, 1, len(alarms))
}

func TestHybridAlertStorage_CountActive(t *testing.T) {
	_, hybrid := setupHybridStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	hybrid.StoreActive(ctx, alarm)

	count, err := hybrid.CountActive(ctx, CountOptions{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestHybridAlertStorage_DeleteActive(t *testing.T) {
	_, hybrid := setupHybridStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	hybrid.StoreActive(ctx, alarm)

	err := hybrid.DeleteActive(ctx, alarm.ID)
	require.NoError(t, err)
}

func TestHybridAlertStorage_StoreHistory(t *testing.T) {
	_, hybrid := setupHybridStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm.Status = entity.AlarmStatusCleared
	err := hybrid.StoreHistory(ctx, alarm)
	require.NoError(t, err)
}

func TestHybridAlertStorage_GetHistory(t *testing.T) {
	_, hybrid := setupHybridStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm.Status = entity.AlarmStatusCleared
	hybrid.StoreHistory(ctx, alarm)

	got, err := hybrid.GetHistory(ctx, alarm.ID)
	require.NoError(t, err)
	assert.NotNil(t, got)
}

func TestHybridAlertStorage_ListHistory(t *testing.T) {
	_, hybrid := setupHybridStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm.Status = entity.AlarmStatusCleared
	hybrid.StoreHistory(ctx, alarm)

	alarms, err := hybrid.ListHistory(ctx, ListHistoryOptions{})
	require.NoError(t, err)
	assert.Equal(t, 1, len(alarms))
}

func TestHybridAlertStorage_DeleteHistory(t *testing.T) {
	_, hybrid := setupHybridStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm.Status = entity.AlarmStatusCleared
	hybrid.StoreHistory(ctx, alarm)

	err := hybrid.DeleteHistory(ctx, alarm.ID)
	require.NoError(t, err)
}

func TestHybridAlertStorage_UpdateStatus(t *testing.T) {
	_, hybrid := setupHybridStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	hybrid.StoreActive(ctx, alarm)

	err := hybrid.UpdateStatus(ctx, alarm.ID, entity.AlarmStatusAcknowledged, UpdateStatusOptions{AcknowledgedBy: "admin"})
	require.NoError(t, err)
}

func TestHybridAlertStorage_Acknowledge(t *testing.T) {
	_, hybrid := setupHybridStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	hybrid.StoreActive(ctx, alarm)

	err := hybrid.Acknowledge(ctx, alarm.ID, "admin")
	require.NoError(t, err)
}

func TestHybridAlertStorage_Clear(t *testing.T) {
	_, hybrid := setupHybridStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	hybrid.StoreActive(ctx, alarm)

	err := hybrid.Clear(ctx, alarm.ID)
	require.NoError(t, err)
}

func TestHybridAlertStorage_CountByLevel(t *testing.T) {
	_, hybrid := setupHybridStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	hybrid.StoreActive(ctx, alarm)

	counts, err := hybrid.CountByLevel(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts[entity.AlarmLevelWarning])
}

func TestHybridAlertStorage_CountByType(t *testing.T) {
	_, hybrid := setupHybridStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	hybrid.StoreActive(ctx, alarm)

	counts, err := hybrid.CountByType(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts[entity.AlarmTypeLimit])
}

func TestHybridAlertStorage_MoveToHistory(t *testing.T) {
	_, hybrid := setupHybridStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	hybrid.StoreActive(ctx, alarm)

	err := hybrid.MoveToHistory(ctx, alarm.ID)
	require.NoError(t, err)
}

func TestHybridAlertStorage_MoveToHistory_NotFound(t *testing.T) {
	_, hybrid := setupHybridStorageForComprehensive(t)
	ctx := context.Background()

	err := hybrid.MoveToHistory(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHybridAlertStorage_SyncFromPostgres(t *testing.T) {
	_, hybrid := setupHybridStorageForComprehensive(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	hybrid.postgres.StoreActive(ctx, alarm)

	err := hybrid.SyncFromPostgres(ctx)
	require.NoError(t, err)
}
