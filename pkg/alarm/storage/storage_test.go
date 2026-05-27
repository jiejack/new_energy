package storage

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRedisStorage(t *testing.T) (*miniredis.Miniredis, *RedisAlertStorage) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	storage := NewRedisAlertStorage(RedisAlertStorageConfig{
		Client: client,
	})
	t.Cleanup(func() { client.Close() })
	return mr, storage
}

func TestNewRedisAlertStorage(t *testing.T) {
	_, storage := setupRedisStorage(t)
	require.NotNil(t, storage)
}

func TestNewRedisAlertStorage_DefaultPrefix(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	storage := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	assert.Equal(t, "nem:alarm", storage.keyPrefix)
}

func TestRedisAlertStorage_StoreActive_GetActive(t *testing.T) {
	_, storage := setupRedisStorage(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test Alarm", "msg")
	err := storage.StoreActive(ctx, alarm)
	require.NoError(t, err)
	assert.NotEmpty(t, alarm.ID)

	got, err := storage.GetActive(ctx, alarm.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "pt1", got.PointID)
	assert.Equal(t, "dev1", got.DeviceID)
	assert.Equal(t, entity.AlarmTypeLimit, got.Type)
}

func TestRedisAlertStorage_GetActive_NotFound(t *testing.T) {
	_, storage := setupRedisStorage(t)
	ctx := context.Background()

	got, err := storage.GetActive(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestRedisAlertStorage_GetActiveByPoint(t *testing.T) {
	_, storage := setupRedisStorage(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	storage.StoreActive(ctx, alarm)

	got, err := storage.GetActiveByPoint(ctx, "pt1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "pt1", got.PointID)
}

func TestRedisAlertStorage_GetActiveByPoint_NotFound(t *testing.T) {
	_, storage := setupRedisStorage(t)
	ctx := context.Background()

	got, err := storage.GetActiveByPoint(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestRedisAlertStorage_ListActive(t *testing.T) {
	_, storage := setupRedisStorage(t)
	ctx := context.Background()

	alarm1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm2 := entity.NewAlarm("pt2", "dev2", "st1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	storage.StoreActive(ctx, alarm1)
	storage.StoreActive(ctx, alarm2)

	alarms, err := storage.ListActive(ctx, ListActiveOptions{})
	require.NoError(t, err)
	assert.Equal(t, 2, len(alarms))
}

func TestRedisAlertStorage_ListActive_ByStation(t *testing.T) {
	_, storage := setupRedisStorage(t)
	ctx := context.Background()

	alarm1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm2 := entity.NewAlarm("pt2", "dev2", "st2", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	storage.StoreActive(ctx, alarm1)
	storage.StoreActive(ctx, alarm2)

	alarms, err := storage.ListActive(ctx, ListActiveOptions{StationID: "st1"})
	require.NoError(t, err)
	assert.Equal(t, 1, len(alarms))
}

func TestRedisAlertStorage_DeleteActive(t *testing.T) {
	_, storage := setupRedisStorage(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	storage.StoreActive(ctx, alarm)

	err := storage.DeleteActive(ctx, alarm.ID)
	require.NoError(t, err)

	got, err := storage.GetActive(ctx, alarm.ID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestRedisAlertStorage_DeleteActive_NotFound(t *testing.T) {
	_, storage := setupRedisStorage(t)
	ctx := context.Background()

	err := storage.DeleteActive(ctx, "nonexistent")
	require.NoError(t, err)
}

func TestRedisAlertStorage_CountActive(t *testing.T) {
	_, storage := setupRedisStorage(t)
	ctx := context.Background()

	alarm1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm2 := entity.NewAlarm("pt2", "dev2", "st1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	storage.StoreActive(ctx, alarm1)
	storage.StoreActive(ctx, alarm2)

	count, err := storage.CountActive(ctx, CountOptions{})
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestRedisAlertStorage_CountActive_ByStation(t *testing.T) {
	_, storage := setupRedisStorage(t)
	ctx := context.Background()

	alarm1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm2 := entity.NewAlarm("pt2", "dev2", "st2", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	storage.StoreActive(ctx, alarm1)
	storage.StoreActive(ctx, alarm2)

	count, err := storage.CountActive(ctx, CountOptions{StationID: "st1"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestRedisAlertStorage_UpdateStatus(t *testing.T) {
	_, storage := setupRedisStorage(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	storage.StoreActive(ctx, alarm)

	err := storage.UpdateStatus(ctx, alarm.ID, entity.AlarmStatusAcknowledged, UpdateStatusOptions{AcknowledgedBy: "admin"})
	require.NoError(t, err)

	got, _ := storage.GetActive(ctx, alarm.ID)
	assert.Equal(t, entity.AlarmStatusAcknowledged, got.Status)
	assert.Equal(t, "admin", got.AcknowledgedBy)
}

func TestRedisAlertStorage_Acknowledge(t *testing.T) {
	_, storage := setupRedisStorage(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	storage.StoreActive(ctx, alarm)

	err := storage.Acknowledge(ctx, alarm.ID, "admin")
	require.NoError(t, err)
}

func TestRedisAlertStorage_Clear(t *testing.T) {
	_, storage := setupRedisStorage(t)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	storage.StoreActive(ctx, alarm)

	err := storage.Clear(ctx, alarm.ID)
	require.NoError(t, err)
}

func TestRedisAlertStorage_HistoryNotSupported(t *testing.T) {
	_, storage := setupRedisStorage(t)
	ctx := context.Background()

	err := storage.StoreHistory(ctx, &entity.Alarm{})
	assert.Error(t, err)

	_, err = storage.GetHistory(ctx, "id")
	assert.Error(t, err)

	_, err = storage.ListHistory(ctx, ListHistoryOptions{})
	assert.Error(t, err)

	err = storage.DeleteHistory(ctx, "id")
	assert.Error(t, err)
}

func TestRedisAlertStorage_CountByLevel(t *testing.T) {
	_, storage := setupRedisStorage(t)
	ctx := context.Background()

	alarm1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm2 := entity.NewAlarm("pt2", "dev2", "st1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	storage.StoreActive(ctx, alarm1)
	storage.StoreActive(ctx, alarm2)

	counts, err := storage.CountByLevel(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts[entity.AlarmLevelWarning])
	assert.Equal(t, int64(1), counts[entity.AlarmLevelCritical])
}

func TestRedisAlertStorage_CountByType(t *testing.T) {
	_, storage := setupRedisStorage(t)
	ctx := context.Background()

	alarm1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	alarm2 := entity.NewAlarm("pt2", "dev2", "st1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	storage.StoreActive(ctx, alarm1)
	storage.StoreActive(ctx, alarm2)

	counts, err := storage.CountByType(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts[entity.AlarmTypeLimit])
	assert.Equal(t, int64(1), counts[entity.AlarmTypeStatus])
}

func TestNewPostgresAlertStorage(t *testing.T) {
	storage := NewPostgresAlertStorage(nil)
	require.NotNil(t, storage)
}

func TestNewHybridAlertStorage(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	redisStorage := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	pgStorage := NewPostgresAlertStorage(nil)
	hybrid := NewHybridAlertStorage(redisStorage, pgStorage)
	require.NotNil(t, hybrid)
}

func TestListActiveOptions_Struct(t *testing.T) {
	opts := ListActiveOptions{
		StationID: "st1",
		DeviceID:  "dev1",
		PointID:   "pt1",
		Limit:     10,
		Offset:    0,
	}
	assert.Equal(t, "st1", opts.StationID)
	assert.Equal(t, 10, opts.Limit)
}

func TestListHistoryOptions_Struct(t *testing.T) {
	opts := ListHistoryOptions{
		StationID: "st1",
		Limit:     20,
		Offset:    0,
		OrderBy:   "triggered_at",
		OrderDesc: true,
	}
	assert.Equal(t, "st1", opts.StationID)
	assert.True(t, opts.OrderDesc)
}

func TestCountOptions_Struct(t *testing.T) {
	opts := CountOptions{StationID: "st1"}
	assert.Equal(t, "st1", opts.StationID)
}

func TestUpdateStatusOptions_Struct(t *testing.T) {
	opts := UpdateStatusOptions{AcknowledgedBy: "admin", Reason: "resolved"}
	assert.Equal(t, "admin", opts.AcknowledgedBy)
}
