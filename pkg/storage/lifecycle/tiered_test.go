package lifecycle

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDataTier_String(t *testing.T) {
	assert.Equal(t, "hot", TierHot.String())
	assert.Equal(t, "warm", TierWarm.String())
	assert.Equal(t, "cold", TierCold.String())
	assert.Equal(t, "unknown", DataTier(99).String())
}

func TestDefaultTierConfig(t *testing.T) {
	cfg := DefaultTierConfig()
	assert.Equal(t, 24*time.Hour, cfg.HotTTL)
	assert.Equal(t, int64(10*1024*1024*1024), cfg.HotMaxSize)
	assert.Equal(t, int64(100), cfg.HotThreshold)
	assert.Equal(t, 30*24*time.Hour, cfg.WarmTTL)
	assert.Equal(t, int64(10), cfg.WarmThreshold)
	assert.Equal(t, 365*24*time.Hour, cfg.ColdTTL)
	assert.Equal(t, 1000, cfg.MigrationBatch)
	assert.Equal(t, 5*time.Minute, cfg.MigrationInterval)
	assert.Equal(t, 1*time.Hour, cfg.StatsWindow)
	assert.Equal(t, 1*time.Minute, cfg.StatsInterval)
}

func TestDataItem_Struct(t *testing.T) {
	item := DataItem{
		ID:        "item1",
		Type:      "metric",
		Timestamp: 1700000000,
		Data:      map[string]interface{}{"value": 42.5},
		Tier:      TierHot,
		HitCount:  10,
		LastAccess: 1700000100,
		Size:      256,
	}
	assert.Equal(t, "item1", item.ID)
	assert.Equal(t, TierHot, item.Tier)
	assert.Equal(t, int64(10), item.HitCount)
}

func TestAccessRecord_Struct(t *testing.T) {
	record := AccessRecord{
		Key:       "test_key",
		Timestamp: time.Now(),
		Count:     5,
	}
	assert.Equal(t, "test_key", record.Key)
	assert.Equal(t, int64(5), record.Count)
}

func TestTierConfig_Struct(t *testing.T) {
	cfg := TierConfig{
		HotTTL:           1 * time.Hour,
		HotMaxSize:       1024,
		HotThreshold:     50,
		WarmTTL:          24 * time.Hour,
		WarmThreshold:    10,
		ColdTTL:          365 * 24 * time.Hour,
		MigrationBatch:   500,
		MigrationInterval: 10 * time.Minute,
		StatsWindow:      30 * time.Minute,
		StatsInterval:    5 * time.Minute,
	}
	assert.Equal(t, 1*time.Hour, cfg.HotTTL)
	assert.Equal(t, int64(1024), cfg.HotMaxSize)
}

func TestHotStats_Struct(t *testing.T) {
	stats := HotStats{
		Key:        "key1",
		HitCount:   100,
		LastAccess: time.Now(),
		FirstSeen:  time.Now().Add(-1 * time.Hour),
		WindowHits: 10,
		UpdatedAt:  time.Now(),
	}
	assert.Equal(t, "key1", stats.Key)
	assert.Equal(t, int64(100), stats.HitCount)
}

func TestTierMetrics_Struct(t *testing.T) {
	metrics := &TierMetrics{
		HotDataCount:   100,
		WarmDataCount:  500,
		ColdDataCount:  1000,
		HotDataSize:    1024 * 1024,
		WarmDataSize:   10 * 1024 * 1024,
		ColdDataSize:   100 * 1024 * 1024,
		MigrationCount: 50,
		HitRate:        0.85,
		MissRate:       0.15,
	}
	assert.Equal(t, int64(100), metrics.HotDataCount)
	assert.InDelta(t, 0.85, metrics.HitRate, 0.01)
}
