package index

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexType_String(t *testing.T) {
	tests := []struct {
		it  IndexType
		exp string
	}{
		{IndexTypeTime, "time"},
		{IndexTypeTag, "tag"},
		{IndexTypeComposite, "composite"},
		{IndexTypeBloom, "bloom"},
		{IndexTypeBitmap, "bitmap"},
		{IndexType(99), "unknown"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.exp, tt.it.String())
	}
}

func TestNewIndexStats(t *testing.T) {
	stats := NewIndexStats()
	require.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.TotalEntries)
}

func TestIndexStats_RecordQuery(t *testing.T) {
	stats := NewIndexStats()
	stats.RecordQuery(true)
	assert.Equal(t, int64(1), stats.QueryCount)
	assert.Equal(t, int64(1), stats.HitCount)
	assert.Equal(t, int64(0), stats.MissCount)

	stats.RecordQuery(false)
	assert.Equal(t, int64(2), stats.QueryCount)
	assert.Equal(t, int64(1), stats.MissCount)
}

func TestIndexStats_GetHitRate(t *testing.T) {
	stats := NewIndexStats()
	assert.Equal(t, float64(0), stats.GetHitRate())

	stats.RecordQuery(true)
	stats.RecordQuery(true)
	stats.RecordQuery(false)
	assert.InDelta(t, 0.667, stats.GetHitRate(), 0.01)
}

func TestIndexStats_RecordRebuild(t *testing.T) {
	stats := NewIndexStats()
	stats.RecordRebuild()
	assert.Equal(t, int64(1), stats.RebuildCount)
	assert.False(t, stats.LastRebuildAt.IsZero())
}

func TestIndexStats_Snapshot(t *testing.T) {
	stats := NewIndexStats()
	stats.RecordQuery(true)
	snap := stats.Snapshot()
	assert.NotNil(t, snap)
	assert.Equal(t, int64(1), snap["queryCount"])
}

func TestTimeIndex(t *testing.T) {
	idx := NewTimeIndex("test_time")
	require.NotNil(t, idx)
	assert.Equal(t, IndexTypeTime, idx.GetType())
	assert.Equal(t, "test_time", idx.GetName())
	assert.Equal(t, int64(0), idx.Count())

	entry := &IndexEntry{
		Key:       "key1",
		Value:     "value1",
		RowID:     1,
		Timestamp: time.Now(),
	}
	err := idx.Insert(entry)
	require.NoError(t, err)
	assert.Equal(t, int64(1), idx.Count())
	assert.True(t, idx.Size() > 0)

	err = idx.Insert(nil)
	assert.Equal(t, ErrInvalidIndexKey, err)
}

func TestTimeIndex_Range(t *testing.T) {
	idx := NewTimeIndex("test_time_range")

	now := time.Now()
	entry1 := &IndexEntry{Key: "k1", Value: "v1", Timestamp: now.Add(-2 * time.Hour)}
	entry2 := &IndexEntry{Key: "k2", Value: "v2", Timestamp: now.Add(-1 * time.Hour)}
	entry3 := &IndexEntry{Key: "k3", Value: "v3", Timestamp: now}

	idx.Insert(entry1)
	idx.Insert(entry2)
	idx.Insert(entry3)

	start := now.Add(-90 * time.Minute).Format(time.RFC3339)
	end := now.Add(30 * time.Minute).Format(time.RFC3339)
	results, err := idx.Range(start, end)
	require.NoError(t, err)
	assert.Equal(t, 2, len(results))
}

func TestTimeIndex_Range_InvalidTime(t *testing.T) {
	idx := NewTimeIndex("test_invalid")
	_, err := idx.Range("invalid", "invalid")
	assert.Error(t, err)
}

func TestTimeIndex_Lookup(t *testing.T) {
	idx := NewTimeIndex("test_lookup")
	_, err := idx.Lookup("key1")
	assert.Equal(t, ErrInvalidIndexKey, err)
}

func TestTimeIndex_Delete(t *testing.T) {
	idx := NewTimeIndex("test_delete")
	ts := time.Now()
	idx.Insert(&IndexEntry{Key: "k1", Value: "v1", Timestamp: ts})

	err := idx.Delete("k1")
	require.NoError(t, err)
}

func TestTimeIndex_Clear(t *testing.T) {
	idx := NewTimeIndex("test_clear")
	idx.Insert(&IndexEntry{Key: "k1", Value: "v1", Timestamp: time.Now()})
	assert.Equal(t, int64(1), idx.Count())

	err := idx.Clear()
	require.NoError(t, err)
	assert.Equal(t, int64(0), idx.Count())
}

func TestTimeIndex_Rebuild(t *testing.T) {
	idx := NewTimeIndex("test_rebuild")
	idx.Insert(&IndexEntry{Key: "k1", Value: "v1", Timestamp: time.Now()})

	entries := []*IndexEntry{
		{Key: "k2", Value: "v2", Timestamp: time.Now()},
		{Key: "k3", Value: "v3", Timestamp: time.Now()},
	}
	err := idx.Rebuild(entries)
	require.NoError(t, err)
	assert.Equal(t, int64(2), idx.Count())
}

func TestTimeIndex_GetEntriesByTimeRange(t *testing.T) {
	idx := NewTimeIndex("test_get_entries")
	now := time.Now()
	idx.Insert(&IndexEntry{Key: "k1", Value: "v1", Timestamp: now})

	results, err := idx.GetEntriesByTimeRange(now.Add(-time.Hour), now.Add(time.Hour))
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))
}

func TestTagIndex(t *testing.T) {
	idx := NewTagIndex("test_tag")
	require.NotNil(t, idx)
	assert.Equal(t, IndexTypeTag, idx.GetType())
	assert.Equal(t, "test_tag", idx.GetName())

	entry := &IndexEntry{Key: "device=sensor1", Value: "v1", Timestamp: time.Now()}
	err := idx.Insert(entry)
	require.NoError(t, err)
	assert.Equal(t, int64(1), idx.Count())

	err = idx.Insert(nil)
	assert.Equal(t, ErrInvalidIndexKey, err)

	err = idx.Insert(&IndexEntry{Key: "noequalsign", Value: "v1", Timestamp: time.Now()})
	require.NoError(t, err)
}

func TestTagIndex_Lookup(t *testing.T) {
	idx := NewTagIndex("test_tag_lookup")
	idx.Insert(&IndexEntry{Key: "device=sensor1", Value: "v1", Timestamp: time.Now()})
	idx.Insert(&IndexEntry{Key: "device=sensor2", Value: "v2", Timestamp: time.Now()})

	results, err := idx.Lookup("device=sensor1")
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))

	results, err = idx.Lookup("device=")
	require.NoError(t, err)
	assert.Equal(t, 2, len(results))

	_, err = idx.Lookup("nonexistent=value")
	assert.Equal(t, ErrIndexNotFound, err)

	_, err = idx.Lookup("")
	assert.Equal(t, ErrInvalidIndexKey, err)

	_, err = idx.Lookup("nonexistent")
	assert.Equal(t, ErrIndexNotFound, err)
}

func TestTagIndex_Range(t *testing.T) {
	idx := NewTagIndex("test_tag_range")
	_, err := idx.Range("a", "z")
	assert.Equal(t, ErrInvalidIndexType, err)
}

func TestTagIndex_Delete(t *testing.T) {
	idx := NewTagIndex("test_tag_delete")
	idx.Insert(&IndexEntry{Key: "device=sensor1", Value: "v1", Timestamp: time.Now()})

	err := idx.Delete("device=sensor1")
	require.NoError(t, err)
}

func TestTagIndex_Clear(t *testing.T) {
	idx := NewTagIndex("test_tag_clear")
	idx.Insert(&IndexEntry{Key: "device=sensor1", Value: "v1", Timestamp: time.Now()})
	err := idx.Clear()
	require.NoError(t, err)
	assert.Equal(t, int64(0), idx.Count())
}

func TestTagIndex_Rebuild(t *testing.T) {
	idx := NewTagIndex("test_tag_rebuild")
	entries := []*IndexEntry{
		{Key: "device=sensor1", Value: "v1", Timestamp: time.Now()},
	}
	err := idx.Rebuild(entries)
	require.NoError(t, err)
	assert.Equal(t, int64(1), idx.Count())
}

func TestTagIndex_GetTagKeys(t *testing.T) {
	idx := NewTagIndex("test_tag_keys")
	idx.Insert(&IndexEntry{Key: "device=sensor1", Value: "v1", Timestamp: time.Now()})
	idx.Insert(&IndexEntry{Key: "location=room1", Value: "v2", Timestamp: time.Now()})

	keys := idx.GetTagKeys()
	assert.Equal(t, 2, len(keys))
}

func TestTagIndex_GetTagValues(t *testing.T) {
	idx := NewTagIndex("test_tag_values")
	idx.Insert(&IndexEntry{Key: "device=sensor1", Value: "v1", Timestamp: time.Now()})
	idx.Insert(&IndexEntry{Key: "device=sensor2", Value: "v2", Timestamp: time.Now()})

	values := idx.GetTagValues("device")
	assert.Equal(t, 2, len(values))

	values = idx.GetTagValues("nonexistent")
	assert.Nil(t, values)
}

func TestCompositeIndex(t *testing.T) {
	idx := NewCompositeIndex("test_composite", []string{"key", "value"})
	require.NotNil(t, idx)
	assert.Equal(t, IndexTypeComposite, idx.GetType())
	assert.Equal(t, "test_composite", idx.GetName())
	assert.Equal(t, []string{"key", "value"}, idx.GetFields())

	err := idx.Insert(nil)
	assert.Equal(t, ErrInvalidIndexKey, err)

	entry := &IndexEntry{Key: "k1", Value: "v1", Timestamp: time.Now()}
	err = idx.Insert(entry)
	require.NoError(t, err)
	assert.Equal(t, int64(1), idx.Count())
}

func TestCompositeIndex_Lookup(t *testing.T) {
	idx := NewCompositeIndex("test_comp_lookup", []string{"key", "value"})
	idx.Insert(&IndexEntry{Key: "k1", Value: "v1", Timestamp: time.Now()})

	results, err := idx.Lookup("k1|v1")
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))

	_, err = idx.Lookup("nonexistent")
	assert.Equal(t, ErrIndexNotFound, err)
}

func TestCompositeIndex_Range(t *testing.T) {
	idx := NewCompositeIndex("test_comp_range", []string{"key", "value"})
	idx.Insert(&IndexEntry{Key: "a", Value: "1", Timestamp: time.Now()})
	idx.Insert(&IndexEntry{Key: "b", Value: "2", Timestamp: time.Now()})
	idx.Insert(&IndexEntry{Key: "c", Value: "3", Timestamp: time.Now()})

	results, err := idx.Range("a|1", "b|2")
	require.NoError(t, err)
	assert.Equal(t, 2, len(results))
}

func TestCompositeIndex_Delete(t *testing.T) {
	idx := NewCompositeIndex("test_comp_delete", []string{"key"})
	idx.Insert(&IndexEntry{Key: "k1", Value: "v1", Timestamp: time.Now()})
	err := idx.Delete("k1|v1")
	require.NoError(t, err)
}

func TestCompositeIndex_Clear(t *testing.T) {
	idx := NewCompositeIndex("test_comp_clear", []string{"key"})
	idx.Insert(&IndexEntry{Key: "k1", Value: "v1", Timestamp: time.Now()})
	err := idx.Clear()
	require.NoError(t, err)
	assert.Equal(t, int64(0), idx.Count())
}

func TestCompositeIndex_Rebuild(t *testing.T) {
	idx := NewCompositeIndex("test_comp_rebuild", []string{"key"})
	entries := []*IndexEntry{
		{Key: "k1", Value: "v1", Timestamp: time.Now()},
		{Key: "k2", Value: "v2", Timestamp: time.Now()},
	}
	err := idx.Rebuild(entries)
	require.NoError(t, err)
	assert.Equal(t, int64(2), idx.Count())
}

func TestBloomIndex(t *testing.T) {
	idx := NewBloomIndex("test_bloom", 1000, 0.01)
	require.NotNil(t, idx)
	assert.Equal(t, IndexTypeBloom, idx.GetType())
	assert.Equal(t, "test_bloom", idx.GetName())
	assert.True(t, idx.Size() > 0)

	err := idx.Insert(nil)
	assert.Equal(t, ErrInvalidIndexKey, err)

	err = idx.Insert(&IndexEntry{Key: "test_key", Value: "v1", Timestamp: time.Now()})
	require.NoError(t, err)
	assert.Equal(t, int64(1), idx.Count())
}

func TestBloomIndex_Lookup(t *testing.T) {
	idx := NewBloomIndex("test_bloom_lookup", 1000, 0.01)
	idx.Insert(&IndexEntry{Key: "existing_key", Value: "v1", Timestamp: time.Now()})

	results, err := idx.Lookup("existing_key")
	assert.NoError(t, err)
	assert.True(t, len(results) > 0)

	_, err = idx.Lookup("definitely_not_existing_key_xyz")
	assert.Equal(t, ErrIndexNotFound, err)
}

func TestBloomIndex_MayContain(t *testing.T) {
	idx := NewBloomIndex("test_bloom_may", 1000, 0.01)
	idx.Insert(&IndexEntry{Key: "key1", Value: "v1", Timestamp: time.Now()})

	assert.True(t, idx.MayContain("key1"))
}

func TestBloomIndex_Delete(t *testing.T) {
	idx := NewBloomIndex("test_bloom_delete", 1000, 0.01)
	err := idx.Delete("key1")
	assert.Equal(t, ErrInvalidIndexType, err)
}

func TestBloomIndex_Range(t *testing.T) {
	idx := NewBloomIndex("test_bloom_range", 1000, 0.01)
	_, err := idx.Range("a", "z")
	assert.Equal(t, ErrInvalidIndexType, err)
}

func TestBloomIndex_Clear(t *testing.T) {
	idx := NewBloomIndex("test_bloom_clear", 1000, 0.01)
	idx.Insert(&IndexEntry{Key: "key1", Value: "v1", Timestamp: time.Now()})
	err := idx.Clear()
	require.NoError(t, err)
	assert.Equal(t, int64(0), idx.Count())
}

func TestBloomIndex_Rebuild(t *testing.T) {
	idx := NewBloomIndex("test_bloom_rebuild", 1000, 0.01)
	entries := []*IndexEntry{
		{Key: "key1", Value: "v1", Timestamp: time.Now()},
		{Key: "key2", Value: "v2", Timestamp: time.Now()},
	}
	err := idx.Rebuild(entries)
	require.NoError(t, err)
	assert.Equal(t, int64(2), idx.Count())
}

func TestDefaultIndexConfig(t *testing.T) {
	cfg := DefaultIndexConfig()
	require.NotNil(t, cfg)
	assert.True(t, cfg.AutoRebuild)
	assert.Equal(t, time.Hour*24, cfg.RebuildInterval)
}

func TestIndexManager(t *testing.T) {
	mgr := NewIndexManager(nil)
	require.NotNil(t, mgr)

	idx, err := mgr.CreateIndex("time_idx", IndexTypeTime)
	require.NoError(t, err)
	assert.Equal(t, IndexTypeTime, idx.GetType())

	got, err := mgr.GetIndex("time_idx")
	require.NoError(t, err)
	assert.Equal(t, idx, got)

	_, err = mgr.GetIndex("nonexistent")
	assert.Equal(t, ErrIndexNotFound, err)

	_, err = mgr.CreateIndex("time_idx", IndexTypeTime)
	assert.Equal(t, ErrIndexAlreadyExists, err)

	names := mgr.ListIndexes()
	assert.Equal(t, 1, len(names))
}

func TestIndexManager_CompositeIndex(t *testing.T) {
	mgr := NewIndexManager(nil)
	idx, err := mgr.CreateIndex("comp_idx", IndexTypeComposite, []string{"field1", "field2"})
	require.NoError(t, err)
	assert.Equal(t, IndexTypeComposite, idx.GetType())
}

func TestIndexManager_BloomIndex(t *testing.T) {
	mgr := NewIndexManager(nil)
	idx, err := mgr.CreateIndex("bloom_idx", IndexTypeBloom, uint(5000), 0.05)
	require.NoError(t, err)
	assert.Equal(t, IndexTypeBloom, idx.GetType())
}

func TestIndexManager_InvalidType(t *testing.T) {
	mgr := NewIndexManager(nil)
	_, err := mgr.CreateIndex("invalid", IndexType(99))
	assert.Equal(t, ErrInvalidIndexType, err)
}

func TestIndexManager_DropIndex(t *testing.T) {
	mgr := NewIndexManager(nil)
	mgr.CreateIndex("to_drop", IndexTypeTime)

	err := mgr.DropIndex("to_drop")
	require.NoError(t, err)

	err = mgr.DropIndex("nonexistent")
	assert.Equal(t, ErrIndexNotFound, err)
}

func TestIndexManager_InsertToIndex(t *testing.T) {
	mgr := NewIndexManager(nil)
	mgr.CreateIndex("tag_idx", IndexTypeTag)

	err := mgr.InsertToIndex("tag_idx", &IndexEntry{Key: "device=sensor1", Value: "v1", Timestamp: time.Now()})
	require.NoError(t, err)

	err = mgr.InsertToIndex("nonexistent", &IndexEntry{Key: "k", Value: "v", Timestamp: time.Now()})
	assert.Equal(t, ErrIndexNotFound, err)
}

func TestIndexManager_LookupFromIndex(t *testing.T) {
	mgr := NewIndexManager(nil)
	mgr.CreateIndex("tag_idx", IndexTypeTag)
	mgr.InsertToIndex("tag_idx", &IndexEntry{Key: "device=sensor1", Value: "v1", Timestamp: time.Now()})

	results, err := mgr.LookupFromIndex("tag_idx", "device=sensor1")
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))

	_, err = mgr.LookupFromIndex("nonexistent", "key")
	assert.Equal(t, ErrIndexNotFound, err)
}

func TestIndexManager_RangeFromIndex(t *testing.T) {
	mgr := NewIndexManager(nil)
	mgr.CreateIndex("time_idx", IndexTypeTime)
	now := time.Now()
	mgr.InsertToIndex("time_idx", &IndexEntry{Key: "k1", Value: "v1", Timestamp: now})

	_, err := mgr.RangeFromIndex("time_idx", now.Add(-time.Hour).Format(time.RFC3339), now.Add(time.Hour).Format(time.RFC3339))
	require.NoError(t, err)
}

func TestIndexManager_RebuildIndex(t *testing.T) {
	mgr := NewIndexManager(nil)
	mgr.CreateIndex("tag_idx", IndexTypeTag)
	mgr.InsertToIndex("tag_idx", &IndexEntry{Key: "device=s1", Value: "v1", Timestamp: time.Now()})

	err := mgr.RebuildIndex("tag_idx", []*IndexEntry{
		{Key: "device=s2", Value: "v2", Timestamp: time.Now()},
	})
	require.NoError(t, err)
}

func TestIndexManager_RebuildAll(t *testing.T) {
	mgr := NewIndexManager(nil)
	mgr.CreateIndex("tag_idx", IndexTypeTag)
	mgr.CreateIndex("time_idx", IndexTypeTime)

	err := mgr.RebuildAll([]*IndexEntry{
		{Key: "device=s1", Value: "v1", Timestamp: time.Now()},
	})
	require.NoError(t, err)
}

func TestIndexManager_GetStats(t *testing.T) {
	mgr := NewIndexManager(nil)
	mgr.CreateIndex("tag_idx", IndexTypeTag)
	stats := mgr.GetStats()
	assert.NotNil(t, stats)
}

func TestIndexManager_GetTotalSize(t *testing.T) {
	mgr := NewIndexManager(nil)
	mgr.CreateIndex("tag_idx", IndexTypeTag)
	mgr.InsertToIndex("tag_idx", &IndexEntry{Key: "device=s1", Value: "v1", Timestamp: time.Now()})
	size := mgr.GetTotalSize()
	assert.True(t, size > 0)
}

func TestIndexManager_GetTotalCount(t *testing.T) {
	mgr := NewIndexManager(nil)
	mgr.CreateIndex("tag_idx", IndexTypeTag)
	mgr.InsertToIndex("tag_idx", &IndexEntry{Key: "device=s1", Value: "v1", Timestamp: time.Now()})
	count := mgr.GetTotalCount()
	assert.Equal(t, int64(1), count)
}

func TestIndexBuilder(t *testing.T) {
	mgr := NewIndexManager(nil)
	mgr.CreateIndex("tag_idx", IndexTypeTag)

	builder := NewIndexBuilder(mgr, 2)
	err := builder.Add(&IndexEntry{Key: "device=s1", Value: "v1", Timestamp: time.Now()})
	require.NoError(t, err)

	err = builder.Add(&IndexEntry{Key: "device=s2", Value: "v2", Timestamp: time.Now()})
	require.NoError(t, err)

	err = builder.Add(&IndexEntry{Key: "device=s3", Value: "v3", Timestamp: time.Now()})
	require.NoError(t, err)

	err = builder.Flush()
	require.NoError(t, err)
}

func TestIndexQuery(t *testing.T) {
	mgr := NewIndexManager(nil)
	mgr.CreateIndex("tag_idx", IndexTypeTag)
	mgr.InsertToIndex("tag_idx", &IndexEntry{Key: "device=s1", Value: "v1", Timestamp: time.Now()})

	q := NewIndexQuery(mgr)
	q.UseIndex("tag_idx")
	q.Where("device", "=", "device=s1")
	q.OrderBy("timestamp")
	q.Limit(10)
	q.Offset(0)

	results, err := q.Execute()
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))
}

func TestIndexQuery_NoIndex(t *testing.T) {
	mgr := NewIndexManager(nil)
	q := NewIndexQuery(mgr)
	_, err := q.Execute()
	assert.Error(t, err)
}

func TestIndexQuery_RangeFilter(t *testing.T) {
	mgr := NewIndexManager(nil)
	mgr.CreateIndex("time_idx", IndexTypeTime)
	now := time.Now()
	mgr.InsertToIndex("time_idx", &IndexEntry{Key: "k1", Value: "v1", Timestamp: now})

	q := NewIndexQuery(mgr)
	q.UseIndex("time_idx")
	q.Where("time", "range", [2]string{now.Add(-time.Hour).Format(time.RFC3339), now.Add(time.Hour).Format(time.RFC3339)})

	results, err := q.Execute()
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))
}
