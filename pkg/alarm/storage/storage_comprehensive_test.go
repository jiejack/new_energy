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

func setupSQLiteDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	db.AutoMigrate(&entity.Alarm{})
	return db
}

func TestRedisAlertStorage_ListActive_ByDevice(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	s := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	ctx := context.Background()

	a1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	a2 := entity.NewAlarm("pt2", "dev2", "st2", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	s.StoreActive(ctx, a1)
	s.StoreActive(ctx, a2)

	alarms, err := s.ListActive(ctx, ListActiveOptions{DeviceID: "dev1"})
	require.NoError(t, err)
	assert.Len(t, alarms, 1)
}

func TestRedisAlertStorage_ListActive_WithLevelFilter(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	s := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	ctx := context.Background()

	warn := entity.AlarmLevelWarning
	crit := entity.AlarmLevelCritical
	a1 := entity.NewAlarm("pt1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	a2 := entity.NewAlarm("pt2", "d2", "s1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	s.StoreActive(ctx, a1)
	s.StoreActive(ctx, a2)

	alarms, err := s.ListActive(ctx, ListActiveOptions{Level: &crit})
	require.NoError(t, err)
	assert.Len(t, alarms, 1)

	alarms, _ = s.ListActive(ctx, ListActiveOptions{Level: &warn})
	assert.Len(t, alarms, 1)
}

func TestRedisAlertStorage_ListActive_WithTypeFilter(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	s := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	ctx := context.Background()

	limitType := entity.AlarmTypeLimit
	a1 := entity.NewAlarm("pt1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	a2 := entity.NewAlarm("pt2", "d2", "s1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	s.StoreActive(ctx, a1)
	s.StoreActive(ctx, a2)

	alarms, err := s.ListActive(ctx, ListActiveOptions{Type: &limitType})
	require.NoError(t, err)
	assert.Len(t, alarms, 1)
}

func TestRedisAlertStorage_ListActive_WithStatusFilter(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	s := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	ctx := context.Background()

	activeStatus := entity.AlarmStatusActive
	a1 := entity.NewAlarm("pt1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	a2 := entity.NewAlarm("pt2", "d2", "s1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	s.StoreActive(ctx, a1)
	s.StoreActive(ctx, a2)

	alarms, err := s.ListActive(ctx, ListActiveOptions{Status: &activeStatus})
	require.NoError(t, err)
	assert.Len(t, alarms, 2)
}

func TestRedisAlertStorage_ListActive_WithPointFilter(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	s := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	ctx := context.Background()

	a1 := entity.NewAlarm("pt1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	a2 := entity.NewAlarm("pt2", "d2", "s1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	s.StoreActive(ctx, a1)
	s.StoreActive(ctx, a2)

	alarms, err := s.ListActive(ctx, ListActiveOptions{PointID: "pt1"})
	require.NoError(t, err)
	assert.Len(t, alarms, 1)
}

func TestRedisAlertStorage_ListActive_WithLimitOffset(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	s := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		a := entity.NewAlarm("pt1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A", "")
		s.StoreActive(ctx, a)
		time.Sleep(1 * time.Millisecond)
	}

	alarms, err := s.ListActive(ctx, ListActiveOptions{Limit: 2, Offset: 1})
	require.NoError(t, err)
	assert.Len(t, alarms, 2)
}

func TestRedisAlertStorage_CountActive_ByDevice(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	s := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	ctx := context.Background()

	a1 := entity.NewAlarm("pt1", "dev1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	a2 := entity.NewAlarm("pt2", "dev2", "s1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	s.StoreActive(ctx, a1)
	s.StoreActive(ctx, a2)

	count, err := s.CountActive(ctx, CountOptions{DeviceID: "dev1"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestRedisAlertStorage_CountByLevel_ByStation(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	s := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	ctx := context.Background()

	stationID := "s1"
	a1 := entity.NewAlarm("pt1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	a2 := entity.NewAlarm("pt2", "d2", "s2", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	s.StoreActive(ctx, a1)
	s.StoreActive(ctx, a2)

	counts, err := s.CountByLevel(ctx, &stationID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts[entity.AlarmLevelWarning])
}

func TestRedisAlertStorage_CountByType_ByStation(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	s := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	ctx := context.Background()

	stationID := "s1"
	a1 := entity.NewAlarm("pt1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	a2 := entity.NewAlarm("pt2", "d2", "s2", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	s.StoreActive(ctx, a1)
	s.StoreActive(ctx, a2)

	counts, err := s.CountByType(ctx, &stationID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts[entity.AlarmTypeLimit])
}

func TestRedisAlertStorage_UpdateStatus_Cleared(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	s := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	s.StoreActive(ctx, alarm)

	err := s.UpdateStatus(ctx, alarm.ID, entity.AlarmStatusCleared, UpdateStatusOptions{})
	require.NoError(t, err)

	got, _ := s.GetActive(ctx, alarm.ID)
	assert.Equal(t, entity.AlarmStatusCleared, got.Status)
	assert.NotNil(t, got.ClearedAt)
}

func TestRedisAlertStorage_UpdateStatus_NotFound(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	s := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})

	err := s.UpdateStatus(context.Background(), "nonexistent", entity.AlarmStatusAcknowledged, UpdateStatusOptions{})
	assert.Error(t, err)
}

func TestPostgresAlertStorage_StoreActive_GetActive(t *testing.T) {
	db := setupSQLiteDB(t)
	s := NewPostgresAlertStorage(db)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test Alarm", "message")
	err := s.StoreActive(ctx, alarm)
	require.NoError(t, err)
	assert.NotEmpty(t, alarm.ID)

	got, err := s.GetActive(ctx, alarm.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "pt1", got.PointID)
	assert.Equal(t, entity.AlarmTypeLimit, got.Type)
}

func TestPostgresAlertStorage_GetActive_NotFound(t *testing.T) {
	db := setupSQLiteDB(t)
	s := NewPostgresAlertStorage(db)
	got, err := s.GetActive(context.Background(), "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestPostgresAlertStorage_GetActiveByPoint(t *testing.T) {
	db := setupSQLiteDB(t)
	s := NewPostgresAlertStorage(db)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "")
	s.StoreActive(ctx, alarm)

	got, err := s.GetActiveByPoint(ctx, "pt1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "pt1", got.PointID)
}

func TestPostgresAlertStorage_GetActiveByPoint_NotFound(t *testing.T) {
	db := setupSQLiteDB(t)
	s := NewPostgresAlertStorage(db)
	got, err := s.GetActiveByPoint(context.Background(), "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestPostgresAlertStorage_ListActive(t *testing.T) {
	db := setupSQLiteDB(t)
	s := NewPostgresAlertStorage(db)
	ctx := context.Background()

	a1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	a2 := entity.NewAlarm("pt2", "dev2", "st1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	s.StoreActive(ctx, a1)
	s.StoreActive(ctx, a2)

	alarms, err := s.ListActive(ctx, ListActiveOptions{})
	require.NoError(t, err)
	assert.Len(t, alarms, 2)
}

func TestPostgresAlertStorage_ListActive_Filters(t *testing.T) {
	db := setupSQLiteDB(t)
	s := NewPostgresAlertStorage(db)
	ctx := context.Background()

	a1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	a2 := entity.NewAlarm("pt2", "dev2", "st2", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	s.StoreActive(ctx, a1)
	s.StoreActive(ctx, a2)

	levelCrit := entity.AlarmLevelCritical
	typeStatus := entity.AlarmTypeStatus

	byStation, err := s.ListActive(ctx, ListActiveOptions{StationID: "st1"})
	require.NoError(t, err)
	assert.Len(t, byStation, 1)

	byLevel, err := s.ListActive(ctx, ListActiveOptions{Level: &levelCrit})
	require.NoError(t, err)
	assert.Len(t, byLevel, 1)

	byType, err := s.ListActive(ctx, ListActiveOptions{Type: &typeStatus})
	require.NoError(t, err)
	assert.Len(t, byType, 1)

	byDevice, err := s.ListActive(ctx, ListActiveOptions{DeviceID: "dev1"})
	require.NoError(t, err)
	assert.Len(t, byDevice, 1)

	byPoint, err := s.ListActive(ctx, ListActiveOptions{PointID: "pt1"})
	require.NoError(t, err)
	assert.Len(t, byPoint, 1)

	withLimit, err := s.ListActive(ctx, ListActiveOptions{Limit: 1})
	require.NoError(t, err)
	assert.Len(t, withLimit, 1)

	withOffset, err := s.ListActive(ctx, ListActiveOptions{Offset: 1})
	require.NoError(t, err)
	assert.Len(t, withOffset, 1)

	activeStatus := entity.AlarmStatusActive
	byStatus, err := s.ListActive(ctx, ListActiveOptions{Status: &activeStatus})
	require.NoError(t, err)
	assert.Len(t, byStatus, 2)
}

func TestPostgresAlertStorage_DeleteActive(t *testing.T) {
	db := setupSQLiteDB(t)
	s := NewPostgresAlertStorage(db)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	s.StoreActive(ctx, alarm)

	err := s.DeleteActive(ctx, alarm.ID)
	require.NoError(t, err)

	got, err := s.GetActive(ctx, alarm.ID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestPostgresAlertStorage_CountActive(t *testing.T) {
	db := setupSQLiteDB(t)
	s := NewPostgresAlertStorage(db)
	ctx := context.Background()

	a1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	a2 := entity.NewAlarm("pt2", "dev2", "st1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	s.StoreActive(ctx, a1)
	s.StoreActive(ctx, a2)

	count, err := s.CountActive(ctx, CountOptions{})
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	countByStation, err := s.CountActive(ctx, CountOptions{StationID: "st1"})
	require.NoError(t, err)
	assert.Equal(t, int64(2), countByStation)

	countByDevice, err := s.CountActive(ctx, CountOptions{DeviceID: "dev1"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), countByDevice)

	warnLevel := entity.AlarmLevelWarning
	countByLevel, err := s.CountActive(ctx, CountOptions{Level: &warnLevel})
	require.NoError(t, err)
	assert.Equal(t, int64(1), countByLevel)

	limitType := entity.AlarmTypeLimit
	countByType, err := s.CountActive(ctx, CountOptions{Type: &limitType})
	require.NoError(t, err)
	assert.Equal(t, int64(1), countByType)
}

func TestPostgresAlertStorage_HistoryOperations(t *testing.T) {
	db := setupSQLiteDB(t)
	s := NewPostgresAlertStorage(db)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "History", "msg")
	err := s.StoreHistory(ctx, alarm)
	require.NoError(t, err)
	assert.NotEmpty(t, alarm.ID)

	got, err := s.GetHistory(ctx, alarm.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "pt1", got.PointID)

	histories, err := s.ListHistory(ctx, ListHistoryOptions{})
	require.NoError(t, err)
	assert.Len(t, histories, 1)

	err = s.DeleteHistory(ctx, alarm.ID)
	require.NoError(t, err)

	afterDelete, err := s.GetHistory(ctx, alarm.ID)
	require.NoError(t, err)
	assert.Nil(t, afterDelete)
}

func TestPostgresAlertStorage_ListHistory_Options(t *testing.T) {
	db := setupSQLiteDB(t)
	s := NewPostgresAlertStorage(db)
	ctx := context.Background()

	now := time.Now().Unix()
	a1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "H1", "")
	a2 := entity.NewAlarm("pt2", "dev2", "st2", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "H2", "")
	s.StoreHistory(ctx, a1)
	s.StoreHistory(ctx, a2)

	byStation, err := s.ListHistory(ctx, ListHistoryOptions{StationID: "st1"})
	require.NoError(t, err)
	assert.Len(t, byStation, 1)

	byTimeRange, err := s.ListHistory(ctx, ListHistoryOptions{
		StartTime: now - 86400,
		EndTime:   now + 3600,
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(byTimeRange), 2)

	ascOrder, err := s.ListHistory(ctx, ListHistoryOptions{OrderBy: "triggered_at", OrderDesc: false})
	require.NoError(t, err)
	assert.NotEmpty(t, ascOrder)

	limitResult, err := s.ListHistory(ctx, ListHistoryOptions{Limit: 1})
	require.NoError(t, err)
	assert.Len(t, limitResult, 1)
}

func TestPostgresAlertStorage_UpdateStatus_Acknowledged(t *testing.T) {
	db := setupSQLiteDB(t)
	s := NewPostgresAlertStorage(db)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	s.StoreActive(ctx, alarm)

	err := s.UpdateStatus(ctx, alarm.ID, entity.AlarmStatusAcknowledged, UpdateStatusOptions{AcknowledgedBy: "admin"})
	require.NoError(t, err)

	got, _ := s.GetActive(ctx, alarm.ID)
	assert.Equal(t, entity.AlarmStatusAcknowledged, got.Status)
	assert.Equal(t, "admin", got.AcknowledgedBy)
	assert.NotNil(t, got.AcknowledgedAt)
}

func TestPostgresAlertStorage_UpdateStatus_Cleared(t *testing.T) {
	db := setupSQLiteDB(t)
	s := NewPostgresAlertStorage(db)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	s.StoreActive(ctx, alarm)

	err := s.UpdateStatus(ctx, alarm.ID, entity.AlarmStatusCleared, UpdateStatusOptions{})
	require.NoError(t, err)

	got, _ := s.GetActive(ctx, alarm.ID)
	assert.Equal(t, entity.AlarmStatusCleared, got.Status)
	assert.NotNil(t, got.ClearedAt)
}

func TestPostgresAlertStorage_Acknowledge_Clear(t *testing.T) {
	db := setupSQLiteDB(t)
	s := NewPostgresAlertStorage(db)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	s.StoreActive(ctx, alarm)

	err := s.Acknowledge(ctx, alarm.ID, "operator")
	require.NoError(t, err)

	err = s.Clear(ctx, alarm.ID)
	require.NoError(t, err)
}

func TestPostgresAlertStorage_CountByLevel(t *testing.T) {
	db := setupSQLiteDB(t)
	s := NewPostgresAlertStorage(db)
	ctx := context.Background()

	a1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	a2 := entity.NewAlarm("pt2", "dev2", "st1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	s.StoreActive(ctx, a1)
	s.StoreActive(ctx, a2)

	counts, err := s.CountByLevel(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts[entity.AlarmLevelWarning])
	assert.Equal(t, int64(1), counts[entity.AlarmLevelCritical])

	stationID := "st1"
	countsWithStation, err := s.CountByLevel(ctx, &stationID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), countsWithStation[entity.AlarmLevelWarning])
}

func TestPostgresAlertStorage_CountByType(t *testing.T) {
	db := setupSQLiteDB(t)
	s := NewPostgresAlertStorage(db)
	ctx := context.Background()

	a1 := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", "")
	a2 := entity.NewAlarm("pt2", "dev2", "st1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "A2", "")
	s.StoreActive(ctx, a1)
	s.StoreActive(ctx, a2)

	counts, err := s.CountByType(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts[entity.AlarmTypeLimit])

	stationID := "st1"
	countsWithStation, err := s.CountByType(ctx, &stationID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), countsWithStation[entity.AlarmTypeLimit])
}

func TestHybridAlertStorage_StoreActive(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	db := setupSQLiteDB(t)

	rs := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	ps := NewPostgresAlertStorage(db)
	hs := NewHybridAlertStorage(rs, ps)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Hybrid", "msg")
	err := hs.StoreActive(ctx, alarm)
	require.NoError(t, err)

	got, err := hs.GetActive(ctx, alarm.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "pt1", got.PointID)
}

func TestHybridAlertStorage_GetActiveByPoint(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	db := setupSQLiteDB(t)

	rs := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	ps := NewPostgresAlertStorage(db)
	hs := NewHybridAlertStorage(rs, ps)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "H", "")
	hs.StoreActive(ctx, alarm)

	got, err := hs.GetActiveByPoint(ctx, "pt1")
	require.NoError(t, err)
	require.NotNil(t, got)
}

func TestHybridAlertStorage_ListActive(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	db := setupSQLiteDB(t)

	rs := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	ps := NewPostgresAlertStorage(db)
	hs := NewHybridAlertStorage(rs, ps)
	ctx := context.Background()

	a1 := entity.NewAlarm("pt1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "HA1", "")
	a2 := entity.NewAlarm("pt2", "d2", "s1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "HA2", "")
	hs.StoreActive(ctx, a1)
	hs.StoreActive(ctx, a2)

	alarms, err := hs.ListActive(ctx, ListActiveOptions{})
	require.NoError(t, err)
	assert.Len(t, alarms, 2)
}

func TestHybridAlertStorage_DeleteActive(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	db := setupSQLiteDB(t)

	rs := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	ps := NewPostgresAlertStorage(db)
	hs := NewHybridAlertStorage(rs, ps)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "HD", "msg")
	hs.StoreActive(ctx, alarm)

	err := hs.DeleteActive(ctx, alarm.ID)
	require.NoError(t, err)

	got, err := hs.GetActive(ctx, alarm.ID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestHybridAlertStorage_CountActive(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	db := setupSQLiteDB(t)

	rs := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	ps := NewPostgresAlertStorage(db)
	hs := NewHybridAlertStorage(rs, ps)
	ctx := context.Background()

	a1 := entity.NewAlarm("pt1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "HC1", "")
	a2 := entity.NewAlarm("pt2", "d2", "s1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "HC2", "")
	hs.StoreActive(ctx, a1)
	hs.StoreActive(ctx, a2)

	count, err := hs.CountActive(ctx, CountOptions{})
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestHybridAlertStorage_HistoryOperations(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	db := setupSQLiteDB(t)

	rs := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	ps := NewPostgresAlertStorage(db)
	hs := NewHybridAlertStorage(rs, ps)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "HH", "msg")
	err := hs.StoreHistory(ctx, alarm)
	require.NoError(t, err)

	got, err := hs.GetHistory(ctx, alarm.ID)
	require.NoError(t, err)
	require.NotNil(t, got)

	list, err := hs.ListHistory(ctx, ListHistoryOptions{})
	require.NoError(t, err)
	assert.Len(t, list, 1)

	err = hs.DeleteHistory(ctx, alarm.ID)
	require.NoError(t, err)
}

func TestHybridAlertStorage_UpdateStatus(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	db := setupSQLiteDB(t)

	rs := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	ps := NewPostgresAlertStorage(db)
	hs := NewHybridAlertStorage(rs, ps)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "HU", "msg")
	hs.StoreActive(ctx, alarm)

	err := hs.UpdateStatus(ctx, alarm.ID, entity.AlarmStatusAcknowledged, UpdateStatusOptions{AcknowledgedBy: "admin"})
	require.NoError(t, err)

	got, _ := hs.GetActive(ctx, alarm.ID)
	assert.Equal(t, entity.AlarmStatusAcknowledged, got.Status)
}

func TestHybridAlertStorage_Acknowledge_Clear(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	db := setupSQLiteDB(t)

	rs := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	ps := NewPostgresAlertStorage(db)
	hs := NewHybridAlertStorage(rs, ps)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "HU", "msg")
	hs.StoreActive(ctx, alarm)

	err := hs.Acknowledge(ctx, alarm.ID, "op")
	require.NoError(t, err)

	err = hs.Clear(ctx, alarm.ID)
	require.NoError(t, err)
}

func TestHybridAlertStorage_CountByLevel(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	db := setupSQLiteDB(t)

	rs := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	ps := NewPostgresAlertStorage(db)
	hs := NewHybridAlertStorage(rs, ps)
	ctx := context.Background()

	a1 := entity.NewAlarm("pt1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "HL1", "")
	a2 := entity.NewAlarm("pt2", "d2", "s1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "HL2", "")
	hs.StoreActive(ctx, a1)
	hs.StoreActive(ctx, a2)

	counts, err := hs.CountByLevel(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts[entity.AlarmLevelWarning])
	assert.Equal(t, int64(1), counts[entity.AlarmLevelCritical])
}

func TestHybridAlertStorage_CountByType(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	db := setupSQLiteDB(t)

	rs := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	ps := NewPostgresAlertStorage(db)
	hs := NewHybridAlertStorage(rs, ps)
	ctx := context.Background()

	a1 := entity.NewAlarm("pt1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "HT1", "")
	a2 := entity.NewAlarm("pt2", "d2", "s1", entity.AlarmTypeStatus, entity.AlarmLevelCritical, "HT2", "")
	hs.StoreActive(ctx, a1)
	hs.StoreActive(ctx, a2)

	counts, err := hs.CountByType(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts[entity.AlarmTypeLimit])
}

func TestHybridAlertStorage_MoveToHistory(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	db := setupSQLiteDB(t)

	rs := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	ps := NewPostgresAlertStorage(db)
	hs := NewHybridAlertStorage(rs, ps)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Move", "msg")
	hs.StoreActive(ctx, alarm)

	err := hs.MoveToHistory(ctx, alarm.ID)
	require.NoError(t, err)

	got, err := rs.GetActive(ctx, alarm.ID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestHybridAlertStorage_MoveToHistory_NotFound(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	db := setupSQLiteDB(t)

	rs := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	ps := NewPostgresAlertStorage(db)
	hs := NewHybridAlertStorage(rs, ps)

	err := hs.MoveToHistory(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestHybridAlertStorage_SyncFromPostgres(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()
	db := setupSQLiteDB(t)

	rs := NewRedisAlertStorage(RedisAlertStorageConfig{Client: client})
	ps := NewPostgresAlertStorage(db)
	hs := NewHybridAlertStorage(rs, ps)
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Sync", "msg")
	ps.StoreActive(ctx, alarm)

	err := hs.SyncFromPostgres(ctx)
	require.NoError(t, err)

	got, err := rs.GetActive(ctx, alarm.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
}
