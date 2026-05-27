package dedup

import (
	"context"
	"testing"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestAlarm(pointID, deviceID, stationID string) *entity.Alarm {
	return entity.NewAlarm(pointID, deviceID, stationID, entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test Alarm", "Test message")
}

func TestDefaultDeduplicationConfig(t *testing.T) {
	cfg := DefaultDeduplicationConfig()
	assert.Equal(t, 5*time.Minute, cfg.WindowDuration)
	assert.Equal(t, 100000, cfg.MaxCacheSize)
	assert.Equal(t, 1*time.Minute, cfg.CleanupInterval)
	assert.False(t, cfg.IncludeValue)
	assert.False(t, cfg.IncludeThreshold)
}

func TestNewDeduplicator(t *testing.T) {
	d := NewDeduplicator(DefaultDeduplicationConfig())
	require.NotNil(t, d)
}

func TestNewDeduplicator_Defaults(t *testing.T) {
	d := NewDeduplicator(DeduplicationConfig{})
	require.NotNil(t, d)
	assert.Equal(t, 5*time.Minute, d.GetWindowDuration())
}

func TestNewDeduplicatorWithDistributed(t *testing.T) {
	mock := &mockDistributedCache{data: make(map[string]*cacheEntry)}
	d := NewDeduplicatorWithDistributed(DefaultDeduplicationConfig(), mock)
	require.NotNil(t, d)
}

func TestDeduplicator_Check_NewAlarm(t *testing.T) {
	d := NewDeduplicator(DefaultDeduplicationConfig())
	ctx := context.Background()
	alarm := newTestAlarm("pt1", "dev1", "st1")

	result := d.Check(ctx, alarm)
	assert.False(t, result.IsDuplicate)
	assert.NotEmpty(t, result.Fingerprint)
	assert.Equal(t, 1, result.OccurrenceCount)
}

func TestDeduplicator_Check_DuplicateAlarm(t *testing.T) {
	d := NewDeduplicator(DefaultDeduplicationConfig())
	ctx := context.Background()
	alarm := newTestAlarm("pt1", "dev1", "st1")

	d.Check(ctx, alarm)
	result := d.Check(ctx, alarm)
	assert.True(t, result.IsDuplicate)
	assert.Equal(t, 2, result.OccurrenceCount)
}

func TestDeduplicator_Check_DifferentAlarms(t *testing.T) {
	d := NewDeduplicator(DefaultDeduplicationConfig())
	ctx := context.Background()

	alarm1 := newTestAlarm("pt1", "dev1", "st1")
	alarm2 := newTestAlarm("pt2", "dev1", "st1")

	r1 := d.Check(ctx, alarm1)
	r2 := d.Check(ctx, alarm2)
	assert.False(t, r1.IsDuplicate)
	assert.False(t, r2.IsDuplicate)
	assert.NotEqual(t, r1.Fingerprint, r2.Fingerprint)
}

func TestDeduplicator_IsDuplicate(t *testing.T) {
	d := NewDeduplicator(DefaultDeduplicationConfig())
	ctx := context.Background()
	alarm := newTestAlarm("pt1", "dev1", "st1")

	assert.False(t, d.IsDuplicate(ctx, alarm))
	assert.True(t, d.IsDuplicate(ctx, alarm))
}

func TestDeduplicator_MarkAsSeen(t *testing.T) {
	d := NewDeduplicator(DefaultDeduplicationConfig())
	ctx := context.Background()
	alarm := newTestAlarm("pt1", "dev1", "st1")

	d.MarkAsSeen(ctx, alarm)
	assert.True(t, d.IsDuplicate(ctx, alarm))
}

func TestDeduplicator_Remove(t *testing.T) {
	d := NewDeduplicator(DefaultDeduplicationConfig())
	ctx := context.Background()
	alarm := newTestAlarm("pt1", "dev1", "st1")

	result := d.Check(ctx, alarm)
	err := d.Remove(ctx, result.Fingerprint)
	require.NoError(t, err)

	assert.False(t, d.IsDuplicate(ctx, alarm))
}

func TestDeduplicator_Cleanup(t *testing.T) {
	cfg := DefaultDeduplicationConfig()
	cfg.WindowDuration = 1 * time.Nanosecond
	d := NewDeduplicator(cfg)
	ctx := context.Background()
	alarm := newTestAlarm("pt1", "dev1", "st1")

	d.Check(ctx, alarm)
	time.Sleep(10 * time.Millisecond)
	d.Cleanup()

	assert.False(t, d.IsDuplicate(ctx, alarm))
}

func TestDeduplicator_StartCleanup(t *testing.T) {
	cfg := DefaultDeduplicationConfig()
	cfg.CleanupInterval = 10 * time.Millisecond
	cfg.WindowDuration = 1 * time.Nanosecond
	d := NewDeduplicator(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go d.StartCleanup(ctx)
	alarm := newTestAlarm("pt1", "dev1", "st1")
	d.Check(ctx, alarm)
	time.Sleep(50 * time.Millisecond)
}

func TestDeduplicator_GetStats(t *testing.T) {
	d := NewDeduplicator(DefaultDeduplicationConfig())
	ctx := context.Background()
	alarm := newTestAlarm("pt1", "dev1", "st1")

	d.Check(ctx, alarm)
	d.Check(ctx, alarm)

	stats := d.GetStats()
	assert.Equal(t, int64(2), stats.TotalProcessed)
	assert.Equal(t, int64(1), stats.DuplicatesFound)
	assert.Equal(t, int64(1), stats.UniqueAlarms)
	assert.Equal(t, int64(1), stats.CacheHits)
	assert.Equal(t, int64(1), stats.CacheMisses)
}

func TestDeduplicator_Reset(t *testing.T) {
	d := NewDeduplicator(DefaultDeduplicationConfig())
	ctx := context.Background()
	alarm := newTestAlarm("pt1", "dev1", "st1")

	d.Check(ctx, alarm)
	d.Reset()

	stats := d.GetStats()
	assert.Equal(t, int64(0), stats.TotalProcessed)
}

func TestDeduplicator_SetWindowDuration(t *testing.T) {
	d := NewDeduplicator(DefaultDeduplicationConfig())
	d.SetWindowDuration(10 * time.Minute)
	assert.Equal(t, 10*time.Minute, d.GetWindowDuration())
}

func TestDeduplicator_BatchCheck(t *testing.T) {
	d := NewDeduplicator(DefaultDeduplicationConfig())
	ctx := context.Background()

	alarms := []*entity.Alarm{
		newTestAlarm("pt1", "dev1", "st1"),
		newTestAlarm("pt2", "dev1", "st1"),
		newTestAlarm("pt1", "dev1", "st1"),
	}

	results := d.BatchCheck(ctx, alarms)
	assert.Equal(t, 3, len(results))
	assert.False(t, results[0].IsDuplicate)
	assert.False(t, results[1].IsDuplicate)
	assert.True(t, results[2].IsDuplicate)
}

func TestDeduplicator_FilterDuplicates(t *testing.T) {
	d := NewDeduplicator(DefaultDeduplicationConfig())
	ctx := context.Background()

	alarms := []*entity.Alarm{
		newTestAlarm("pt1", "dev1", "st1"),
		newTestAlarm("pt2", "dev1", "st1"),
		newTestAlarm("pt1", "dev1", "st1"),
	}

	unique := d.FilterDuplicates(ctx, alarms)
	assert.Equal(t, 2, len(unique))
}

func TestDeduplicator_GetFingerprintInfo(t *testing.T) {
	d := NewDeduplicator(DefaultDeduplicationConfig())
	ctx := context.Background()
	alarm := newTestAlarm("pt1", "dev1", "st1")

	result := d.Check(ctx, alarm)
	entry, found := d.GetFingerprintInfo(ctx, result.Fingerprint)
	assert.True(t, found)
	assert.NotNil(t, entry)

	_, found = d.GetFingerprintInfo(ctx, "nonexistent")
	assert.False(t, found)
}

func TestDeduplicator_UpdateFingerprintTTL(t *testing.T) {
	d := NewDeduplicator(DefaultDeduplicationConfig())
	ctx := context.Background()
	alarm := newTestAlarm("pt1", "dev1", "st1")

	result := d.Check(ctx, alarm)
	err := d.UpdateFingerprintTTL(ctx, result.Fingerprint, 10*time.Minute)
	require.NoError(t, err)

	err = d.UpdateFingerprintTTL(ctx, "nonexistent", 10*time.Minute)
	assert.Error(t, err)
}

func TestGenerateFingerprintFromFields(t *testing.T) {
	fp := GenerateFingerprintFromFields("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning)
	assert.NotEmpty(t, fp)
}

func TestGenerateFingerprint(t *testing.T) {
	d := NewDeduplicator(DeduplicationConfig{IncludeValue: true, IncludeThreshold: true})
	alarm := newTestAlarm("pt1", "dev1", "st1")
	alarm.Value = 42.5
	alarm.Threshold = 50.0

	fp := d.GenerateFingerprint(alarm)
	assert.NotEmpty(t, fp)
}

func TestDeduplicator_MaxCacheSize(t *testing.T) {
	cfg := DefaultDeduplicationConfig()
	cfg.MaxCacheSize = 3
	cfg.WindowDuration = 1 * time.Hour
	d := NewDeduplicator(cfg)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		alarm := newTestAlarm("pt"+string(rune('0'+i)), "dev1", "st1")
		d.Check(ctx, alarm)
	}

	stats := d.GetStats()
	assert.True(t, stats.CurrentCacheSize <= 3)
}

func TestDeduplicator_WithDistributedCache(t *testing.T) {
	mock := &mockDistributedCache{data: make(map[string]*cacheEntry)}
	d := NewDeduplicatorWithDistributed(DefaultDeduplicationConfig(), mock)
	ctx := context.Background()
	alarm := newTestAlarm("pt1", "dev1", "st1")

	result := d.Check(ctx, alarm)
	assert.False(t, result.IsDuplicate)

	result = d.Check(ctx, alarm)
	assert.True(t, result.IsDuplicate)
}

func TestDeduplicator_Distributed_Remove(t *testing.T) {
	mock := &mockDistributedCache{data: make(map[string]*cacheEntry)}
	d := NewDeduplicatorWithDistributed(DefaultDeduplicationConfig(), mock)
	ctx := context.Background()
	alarm := newTestAlarm("pt1", "dev1", "st1")

	result := d.Check(ctx, alarm)
	err := d.Remove(ctx, result.Fingerprint)
	require.NoError(t, err)
}

func TestDeduplicator_Distributed_GetFingerprintInfo(t *testing.T) {
	mock := &mockDistributedCache{data: make(map[string]*cacheEntry)}
	d := NewDeduplicatorWithDistributed(DefaultDeduplicationConfig(), mock)
	ctx := context.Background()
	alarm := newTestAlarm("pt1", "dev1", "st1")

	result := d.Check(ctx, alarm)
	entry, found := d.GetFingerprintInfo(ctx, result.Fingerprint)
	assert.True(t, found)
	assert.NotNil(t, entry)
}

func TestDeduplicator_Distributed_UpdateFingerprintTTL(t *testing.T) {
	mock := &mockDistributedCache{data: make(map[string]*cacheEntry)}
	d := NewDeduplicatorWithDistributed(DefaultDeduplicationConfig(), mock)
	ctx := context.Background()
	alarm := newTestAlarm("pt1", "dev1", "st1")

	result := d.Check(ctx, alarm)
	err := d.UpdateFingerprintTTL(ctx, result.Fingerprint, 10*time.Minute)
	require.NoError(t, err)
}

func TestDeduplicationResult_Struct(t *testing.T) {
	r := &DeduplicationResult{
		IsDuplicate:     true,
		Fingerprint:     "abc",
		OccurrenceCount: 3,
	}
	assert.True(t, r.IsDuplicate)
	assert.Equal(t, 3, r.OccurrenceCount)
}

type mockDistributedCache struct {
	data map[string]*cacheEntry
}

func (m *mockDistributedCache) Get(ctx context.Context, key string) (*cacheEntry, error) {
	entry, ok := m.data[key]
	if !ok {
		return nil, nil
	}
	return entry, nil
}

func (m *mockDistributedCache) Set(ctx context.Context, key string, entry *cacheEntry, ttl time.Duration) error {
	m.data[key] = entry
	return nil
}

func (m *mockDistributedCache) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *mockDistributedCache) Exists(ctx context.Context, key string) (bool, error) {
	_, ok := m.data[key]
	return ok, nil
}
