package index

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewIndex(t *testing.T) {
	config := IndexConfig{
		Type:      "btree",
		PageSize:  4096,
		CacheSize: 1000,
	}

	index, err := NewIndex(config)
	assert.NoError(t, err)
	assert.NotNil(t, index)
}

func TestIndex_Add(t *testing.T) {
	config := IndexConfig{
		Type:      "btree",
		PageSize:  4096,
		CacheSize: 1000,
	}

	index, err := NewIndex(config)
	assert.NoError(t, err)

	entry := &IndexEntry{
		Key:       "point001",
		Timestamp: time.Now(),
		Offset:    100,
		Size:      50,
	}

	err = index.Add(entry)
	assert.NoError(t, err)
}

func TestIndex_Get(t *testing.T) {
	config := IndexConfig{
		Type:      "btree",
		PageSize:  4096,
		CacheSize: 1000,
	}

	index, err := NewIndex(config)
	assert.NoError(t, err)

	now := time.Now()
	entry := &IndexEntry{
		Key:       "point001",
		Timestamp: now,
		Offset:    100,
		Size:      50,
	}

	index.Add(entry)

	got, err := index.Get("point001", now)
	assert.NoError(t, err)
	assert.Equal(t, entry.Offset, got.Offset)
}

func TestIndex_GetRange(t *testing.T) {
	config := IndexConfig{
		Type:      "btree",
		PageSize:  4096,
		CacheSize: 1000,
	}

	index, err := NewIndex(config)
	assert.NoError(t, err)

	now := time.Now()
	for i := 0; i < 10; i++ {
		entry := &IndexEntry{
			Key:       "point001",
			Timestamp: now.Add(time.Duration(i) * time.Hour),
			Offset:    int64(i * 100),
			Size:      50,
		}
		index.Add(entry)
	}

	start := now.Add(2 * time.Hour)
	end := now.Add(6 * time.Hour)

	entries, err := index.GetRange("point001", start, end)
	assert.NoError(t, err)
	assert.Len(t, entries, 5) // 2,3,4,5,6
}

func TestIndex_Delete(t *testing.T) {
	config := IndexConfig{
		Type:      "btree",
		PageSize:  4096,
		CacheSize: 1000,
	}

	index, err := NewIndex(config)
	assert.NoError(t, err)

	now := time.Now()
	entry := &IndexEntry{
		Key:       "point001",
		Timestamp: now,
		Offset:    100,
		Size:      50,
	}

	index.Add(entry)

	err = index.Delete("point001", now)
	assert.NoError(t, err)

	_, err = index.Get("point001", now)
	assert.Error(t, err)
}

func TestNewTimeSeriesIndex(t *testing.T) {
	config := TimeSeriesIndexConfig{
		PartitionDuration: 24 * time.Hour,
		RetentionDays:     30,
	}

	index := NewTimeSeriesIndex(config)

	assert.NotNil(t, index)
	assert.Equal(t, config.PartitionDuration, index.config.PartitionDuration)
}

func TestTimeSeriesIndex_AddPoint(t *testing.T) {
	config := TimeSeriesIndexConfig{
		PartitionDuration: 24 * time.Hour,
		RetentionDays:     30,
	}

	index := NewTimeSeriesIndex(config)

	now := time.Now()
	point := &TimeSeriesPoint{
		PointID:   "point001",
		Timestamp: now,
		Value:     100.5,
	}

	err := index.AddPoint(point)
	assert.NoError(t, err)
}

func TestTimeSeriesIndex_QueryPoints(t *testing.T) {
	config := TimeSeriesIndexConfig{
		PartitionDuration: 24 * time.Hour,
		RetentionDays:     30,
	}

	index := NewTimeSeriesIndex(config)

	now := time.Now()
	for i := 0; i < 10; i++ {
		point := &TimeSeriesPoint{
			PointID:   "point001",
			Timestamp: now.Add(time.Duration(i) * time.Hour),
			Value:     float64(i * 10),
		}
		index.AddPoint(point)
	}

	start := now.Add(2 * time.Hour)
	end := now.Add(6 * time.Hour)

	points, err := index.QueryPoints("point001", start, end)
	assert.NoError(t, err)
	assert.Len(t, points, 5)
}

func TestTimeSeriesIndex_GetLatestPoint(t *testing.T) {
	config := TimeSeriesIndexConfig{
		PartitionDuration: 24 * time.Hour,
		RetentionDays:     30,
	}

	index := NewTimeSeriesIndex(config)

	now := time.Now()
	for i := 0; i < 5; i++ {
		point := &TimeSeriesPoint{
			PointID:   "point001",
			Timestamp: now.Add(time.Duration(i) * time.Hour),
			Value:     float64(i * 10),
		}
		index.AddPoint(point)
	}

	latest, err := index.GetLatestPoint("point001")
	assert.NoError(t, err)
	assert.Equal(t, 40.0, latest.Value)
}

func TestTimeSeriesIndex_DeleteBefore(t *testing.T) {
	config := TimeSeriesIndexConfig{
		PartitionDuration: 24 * time.Hour,
		RetentionDays:     30,
	}

	index := NewTimeSeriesIndex(config)

	now := time.Now()
	for i := 0; i < 10; i++ {
		point := &TimeSeriesPoint{
			PointID:   "point001",
			Timestamp: now.Add(time.Duration(i-5) * 24 * time.Hour),
			Value:     float64(i * 10),
		}
		index.AddPoint(point)
	}

	// 删除3天前的数据
	cutoff := now.Add(-3 * 24 * time.Hour)
	err := index.DeleteBefore(cutoff)
	assert.NoError(t, err)
}

func TestIndexCache(t *testing.T) {
	cache := NewIndexCache(100)

	entry := &IndexEntry{
		Key:       "point001",
		Timestamp: time.Now(),
		Offset:    100,
		Size:      50,
	}

	// 添加到缓存
	cache.Set("point001", entry)

	// 从缓存获取
	got, found := cache.Get("point001")
	assert.True(t, found)
	assert.Equal(t, entry.Offset, got.Offset)

	// 不存在的键
	_, found = cache.Get("point999")
	assert.False(t, found)

	// 删除
	cache.Delete("point001")
	_, found = cache.Get("point001")
	assert.False(t, found)
}

func TestIndexCache_Eviction(t *testing.T) {
	cache := NewIndexCache(5)

	// 添加超过容量的条目
	for i := 0; i < 10; i++ {
		entry := &IndexEntry{
			Key:       string(rune('a' + i)),
			Timestamp: time.Now(),
			Offset:    int64(i),
		}
		cache.Set(string(rune('a'+i)), entry)
	}

	// 应该只保留最新的5个
	assert.LessOrEqual(t, cache.Size(), 5)
}
