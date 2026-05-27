package sharding

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShardStatus_String(t *testing.T) {
	tests := []struct {
		status ShardStatus
		exp    string
	}{
		{ShardStatusActive, "active"},
		{ShardStatusInactive, "inactive"},
		{ShardStatusMigrating, "migrating"},
		{ShardStatusRebalancing, "rebalancing"},
		{ShardStatus(99), "unknown"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.exp, tt.status.String())
	}
}

func TestTimeGranularity_String(t *testing.T) {
	tests := []struct {
		tg  TimeGranularity
		exp string
	}{
		{TimeGranularityHour, "hour"},
		{TimeGranularityDay, "day"},
		{TimeGranularityWeek, "week"},
		{TimeGranularityMonth, "month"},
		{TimeGranularityYear, "year"},
		{TimeGranularity(99), "unknown"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.exp, tt.tg.String())
	}
}

func TestNewHashSharding(t *testing.T) {
	hs, err := NewHashSharding(4)
	require.NoError(t, err)
	assert.Equal(t, "hash", hs.GetType())
	assert.Equal(t, 4, hs.GetShardCount())
}

func TestNewHashSharding_InvalidCount(t *testing.T) {
	_, err := NewHashSharding(0)
	assert.Equal(t, ErrInvalidShardCount, err)

	_, err = NewHashSharding(-1)
	assert.Equal(t, ErrInvalidShardCount, err)
}

func TestNewHashSharding_WithOptions(t *testing.T) {
	hs, err := NewHashSharding(4,
		WithVirtualNodes(100),
		WithHashFunc(md5Hash),
	)
	require.NoError(t, err)
	assert.Equal(t, 4, hs.GetShardCount())
}

func TestHashSharding_GetShard(t *testing.T) {
	hs, _ := NewHashSharding(4)
	key := &ShardKey{
		DeviceID: "device1",
		PointID:  "point1",
	}
	shardID, err := hs.GetShard(key)
	require.NoError(t, err)
	assert.True(t, shardID >= 0 && shardID < 4)
}

func TestHashSharding_GetShard_NilKey(t *testing.T) {
	hs, _ := NewHashSharding(4)
	_, err := hs.GetShard(nil)
	assert.Equal(t, ErrInvalidShardKey, err)
}

func TestHashSharding_GetShard_Consistent(t *testing.T) {
	hs, _ := NewHashSharding(4)
	key := &ShardKey{DeviceID: "device1", PointID: "point1"}

	id1, _ := hs.GetShard(key)
	id2, _ := hs.GetShard(key)
	assert.Equal(t, id1, id2)
}

func TestHashSharding_GetShard_Distributed(t *testing.T) {
	hs, _ := NewHashSharding(8)
	shardCounts := make(map[int]int)
	for i := 0; i < 100; i++ {
		key := &ShardKey{DeviceID: "device" + string(rune(i)), PointID: "point1"}
		id, err := hs.GetShard(key)
		require.NoError(t, err)
		shardCounts[id]++
	}
	assert.True(t, len(shardCounts) > 1)
}

func TestHashSharding_AddShard(t *testing.T) {
	hs, _ := NewHashSharding(2)
	shard := &Shard{ID: 10, Name: "shard-10", Status: ShardStatusActive}
	err := hs.AddShard(shard)
	require.NoError(t, err)
	assert.Equal(t, 3, hs.GetShardCount())
}

func TestHashSharding_AddShard_Duplicate(t *testing.T) {
	hs, _ := NewHashSharding(2)
	shard := &Shard{ID: 0, Name: "dup", Status: ShardStatusActive}
	err := hs.AddShard(shard)
	assert.Error(t, err)
}

func TestHashSharding_RemoveShard(t *testing.T) {
	hs, _ := NewHashSharding(4)
	err := hs.RemoveShard(0)
	require.NoError(t, err)
	assert.Equal(t, 3, hs.GetShardCount())
}

func TestHashSharding_RemoveShard_NotFound(t *testing.T) {
	hs, _ := NewHashSharding(4)
	err := hs.RemoveShard(999)
	assert.Equal(t, ErrShardNotFound, err)
}

func TestHashSharding_Rebalance(t *testing.T) {
	hs, _ := NewHashSharding(4)
	err := hs.Rebalance()
	require.NoError(t, err)
}

func TestHashSharding_GetShards(t *testing.T) {
	hs, _ := NewHashSharding(4)
	shards := hs.GetShards()
	assert.Equal(t, 4, len(shards))
}

func TestNewRangeSharding(t *testing.T) {
	rs := NewRangeSharding()
	assert.Equal(t, "range", rs.GetType())
	assert.Equal(t, 0, rs.GetShardCount())
}

func TestRangeSharding_AddRangeShard(t *testing.T) {
	rs := NewRangeSharding()
	shard := &Shard{ID: 0, Name: "shard-0", Status: ShardStatusActive}
	err := rs.AddRangeShard(shard, []byte("a"), []byte("m"))
	require.NoError(t, err)
	assert.Equal(t, 1, rs.GetShardCount())
}

func TestRangeSharding_AddRangeShard_Overlap(t *testing.T) {
	rs := NewRangeSharding()
	rs.AddRangeShard(&Shard{ID: 0, Name: "s0", Status: ShardStatusActive}, []byte("a"), []byte("m"))
	err := rs.AddRangeShard(&Shard{ID: 1, Name: "s1", Status: ShardStatusActive}, []byte("l"), []byte("z"))
	assert.Error(t, err)
}

func TestRangeSharding_GetShard(t *testing.T) {
	rs := NewRangeSharding()
	rs.AddRangeShard(&Shard{ID: 0, Name: "s0", Status: ShardStatusActive}, []byte("a"), []byte("m"))
	rs.AddRangeShard(&Shard{ID: 1, Name: "s1", Status: ShardStatusActive}, []byte("m"), []byte("z"))

	key := &ShardKey{DeviceID: "device1"}
	id, err := rs.GetShard(key)
	require.NoError(t, err)
	assert.True(t, id == 0 || id == 1)
}

func TestRangeSharding_GetShard_NilKey(t *testing.T) {
	rs := NewRangeSharding()
	_, err := rs.GetShard(nil)
	assert.Equal(t, ErrInvalidShardKey, err)
}

func TestRangeSharding_GetShard_NotFound(t *testing.T) {
	rs := NewRangeSharding()
	key := &ShardKey{DeviceID: "zzz"}
	_, err := rs.GetShard(key)
	assert.Equal(t, ErrShardNotFound, err)
}

func TestRangeSharding_AddShard_UseAddRange(t *testing.T) {
	rs := NewRangeSharding()
	err := rs.AddShard(&Shard{})
	assert.Error(t, err)
}

func TestRangeSharding_RemoveShard(t *testing.T) {
	rs := NewRangeSharding()
	rs.AddRangeShard(&Shard{ID: 0, Name: "s0", Status: ShardStatusActive}, []byte("a"), []byte("z"))
	err := rs.RemoveShard(0)
	require.NoError(t, err)
	assert.Equal(t, 0, rs.GetShardCount())
}

func TestRangeSharding_RemoveShard_NotFound(t *testing.T) {
	rs := NewRangeSharding()
	err := rs.RemoveShard(999)
	assert.Equal(t, ErrShardNotFound, err)
}

func TestRangeSharding_Rebalance(t *testing.T) {
	rs := NewRangeSharding()
	err := rs.Rebalance()
	require.NoError(t, err)
}

func TestNewTimeSharding(t *testing.T) {
	ts, err := NewTimeSharding(TimeGranularityDay)
	require.NoError(t, err)
	assert.Equal(t, "time", ts.GetType())
	assert.Equal(t, 0, ts.GetShardCount())
}

func TestNewTimeSharding_WithOptions(t *testing.T) {
	ts, err := NewTimeSharding(TimeGranularityHour,
		WithAutoCreate(false),
		WithMaxShards(100),
	)
	require.NoError(t, err)
	assert.Equal(t, 0, ts.GetShardCount())
}

func TestTimeSharding_GetShard_AutoCreate(t *testing.T) {
	ts, _ := NewTimeSharding(TimeGranularityDay)
	key := &ShardKey{
		DeviceID:  "device1",
		PointID:   "point1",
		Timestamp: time.Now(),
	}
	id, err := ts.GetShard(key)
	require.NoError(t, err)
	assert.True(t, id >= 0)
	assert.Equal(t, 1, ts.GetShardCount())
}

func TestTimeSharding_GetShard_NilKey(t *testing.T) {
	ts, _ := NewTimeSharding(TimeGranularityDay)
	_, err := ts.GetShard(nil)
	assert.Equal(t, ErrInvalidShardKey, err)
}

func TestTimeSharding_GetShard_AutoCreateDisabled(t *testing.T) {
	ts, _ := NewTimeSharding(TimeGranularityDay, WithAutoCreate(false))
	key := &ShardKey{Timestamp: time.Now()}
	_, err := ts.GetShard(key)
	assert.Equal(t, ErrShardNotFound, err)
}

func TestTimeSharding_AddShard(t *testing.T) {
	ts, _ := NewTimeSharding(TimeGranularityDay, WithAutoCreate(false))
	now := time.Now()
	shard := &Shard{
		ID:        0,
		Name:      "shard-test",
		StartTime: now,
		EndTime:   now.AddDate(0, 0, 1),
		Status:    ShardStatusActive,
	}
	err := ts.AddShard(shard)
	require.NoError(t, err)
	assert.Equal(t, 1, ts.GetShardCount())
}

func TestTimeSharding_AddShard_InvalidTime(t *testing.T) {
	ts, _ := NewTimeSharding(TimeGranularityDay)
	shard := &Shard{ID: 0, Name: "bad", Status: ShardStatusActive}
	err := ts.AddShard(shard)
	assert.Error(t, err)
}

func TestTimeSharding_AddShard_Overlap(t *testing.T) {
	ts, _ := NewTimeSharding(TimeGranularityDay, WithAutoCreate(false))
	now := time.Now()
	ts.AddShard(&Shard{ID: 0, Name: "s0", StartTime: now, EndTime: now.AddDate(0, 0, 2), Status: ShardStatusActive})
	err := ts.AddShard(&Shard{ID: 1, Name: "s1", StartTime: now.AddDate(0, 0, 1), EndTime: now.AddDate(0, 0, 3), Status: ShardStatusActive})
	assert.Error(t, err)
}

func TestTimeSharding_RemoveShard(t *testing.T) {
	ts, _ := NewTimeSharding(TimeGranularityDay)
	ts.GetShard(&ShardKey{Timestamp: time.Now()})
	shards := ts.GetShards()
	err := ts.RemoveShard(shards[0].ID)
	require.NoError(t, err)
}

func TestTimeSharding_RemoveShard_NotFound(t *testing.T) {
	ts, _ := NewTimeSharding(TimeGranularityDay)
	err := ts.RemoveShard(999)
	assert.Equal(t, ErrShardNotFound, err)
}

func TestTimeSharding_Rebalance(t *testing.T) {
	ts, _ := NewTimeSharding(TimeGranularityDay)
	ts.GetShard(&ShardKey{Timestamp: time.Now()})
	err := ts.Rebalance()
	require.NoError(t, err)
}

func TestTimeSharding_GetShardsByTimeRange(t *testing.T) {
	ts, _ := NewTimeSharding(TimeGranularityDay)
	now := time.Now()
	ts.GetShard(&ShardKey{Timestamp: now})
	ts.GetShard(&ShardKey{Timestamp: now.AddDate(0, 0, 1)})

	result := ts.GetShardsByTimeRange(now, now.AddDate(0, 0, 2))
	assert.Equal(t, 2, len(result))
}

func TestTimeSharding_MaxShards(t *testing.T) {
	ts, _ := NewTimeSharding(TimeGranularityHour, WithMaxShards(2))
	now := time.Now()
	ts.GetShard(&ShardKey{Timestamp: now})
	ts.GetShard(&ShardKey{Timestamp: now.Add(1 * time.Hour)})

	_, err := ts.GetShard(&ShardKey{Timestamp: now.Add(2 * time.Hour)})
	assert.Error(t, err)
}

func TestNewShardRouter(t *testing.T) {
	hs, _ := NewHashSharding(4)
	router := NewShardRouter(hs, 1000)
	require.NotNil(t, router)
}

func TestShardRouter_Route(t *testing.T) {
	hs, _ := NewHashSharding(4)
	router := NewShardRouter(hs, 1000)

	key := &ShardKey{DeviceID: "device1", PointID: "point1"}
	id, err := router.Route(key)
	require.NoError(t, err)
	assert.True(t, id >= 0 && id < 4)
}

func TestShardRouter_Route_NilKey(t *testing.T) {
	hs, _ := NewHashSharding(4)
	router := NewShardRouter(hs, 1000)
	_, err := router.Route(nil)
	assert.Equal(t, ErrInvalidShardKey, err)
}

func TestShardRouter_Route_CacheHit(t *testing.T) {
	hs, _ := NewHashSharding(4)
	router := NewShardRouter(hs, 1000)

	key := &ShardKey{DeviceID: "device1", PointID: "point1", Timestamp: time.Now()}
	id1, _ := router.Route(key)
	id2, _ := router.Route(key)
	assert.Equal(t, id1, id2)

	stats := router.GetStats()
	assert.True(t, stats["cacheHits"].(int64) > 0)
}

func TestShardRouter_RouteBatch(t *testing.T) {
	hs, _ := NewHashSharding(4)
	router := NewShardRouter(hs, 1000)

	keys := []*ShardKey{
		{DeviceID: "d1", PointID: "p1", Timestamp: time.Now()},
		{DeviceID: "d2", PointID: "p2", Timestamp: time.Now()},
		{DeviceID: "d3", PointID: "p3", Timestamp: time.Now()},
	}
	result, err := router.RouteBatch(keys)
	require.NoError(t, err)
	assert.True(t, len(result) > 0)
}

func TestShardRouter_Rebalance(t *testing.T) {
	hs, _ := NewHashSharding(4)
	router := NewShardRouter(hs, 1000)
	err := router.Rebalance()
	require.NoError(t, err)
}

func TestNewRebalancer(t *testing.T) {
	hs, _ := NewHashSharding(4)
	rb := NewRebalancer(hs, 0.3)
	require.NotNil(t, rb)
	assert.False(t, rb.IsInProgress())
}

func TestRebalancer_Analyze(t *testing.T) {
	hs, _ := NewHashSharding(4)
	rb := NewRebalancer(hs, 0.3)
	analysis, err := rb.Analyze()
	require.NoError(t, err)
	assert.Equal(t, 4, analysis.TotalShards)
	assert.False(t, analysis.NeedsRebalance)
}

func TestRebalancer_Analyze_NoShards(t *testing.T) {
	rs := NewRangeSharding()
	rb := NewRebalancer(rs, 0.3)
	_, err := rb.Analyze()
	assert.Equal(t, ErrInvalidShardCount, err)
}

func TestRebalancer_Execute(t *testing.T) {
	hs, _ := NewHashSharding(4)
	rb := NewRebalancer(hs, 0.3)
	err := rb.Execute()
	require.NoError(t, err)
}

func TestRebalancer_Execute_WithHandlers(t *testing.T) {
	hs, _ := NewHashSharding(4)
	rb := NewRebalancer(hs, 0.3)

	completeCalled := false
	rb.SetCompleteHandler(func(stats *RebalanceStats) {
		completeCalled = true
	})
	rb.SetMigrateHandler(func(shardID int, fromNode, toNode string) error {
		return nil
	})

	err := rb.Execute()
	require.NoError(t, err)
	assert.True(t, completeCalled)
}

func TestRebalancer_Execute_InProgress(t *testing.T) {
	hs, _ := NewHashSharding(4)
	rb := NewRebalancer(hs, 0.3)

	rb.mu.Lock()
	rb.inProgress = true
	rb.mu.Unlock()

	err := rb.Execute()
	assert.Equal(t, ErrRebalanceInProgress, err)
}

func TestNewShardManager(t *testing.T) {
	hs, _ := NewHashSharding(4)
	sm := NewShardManager(hs, 1000)
	require.NotNil(t, sm)
}

func TestShardManager_Route(t *testing.T) {
	hs, _ := NewHashSharding(4)
	sm := NewShardManager(hs, 1000)

	key := &ShardKey{DeviceID: "device1", PointID: "point1"}
	id, err := sm.Route(key)
	require.NoError(t, err)
	assert.True(t, id >= 0 && id < 4)
}

func TestShardManager_RegisterStrategy(t *testing.T) {
	hs, _ := NewHashSharding(4)
	sm := NewShardManager(hs, 1000)

	rs := NewRangeSharding()
	sm.RegisterStrategy("range", rs)

	strategy, ok := sm.GetStrategy("range")
	assert.True(t, ok)
	assert.Equal(t, rs, strategy)
}

func TestShardManager_RouteWithStrategy(t *testing.T) {
	hs, _ := NewHashSharding(4)
	sm := NewShardManager(hs, 1000)

	_, err := sm.RouteWithStrategy("nonexistent", &ShardKey{DeviceID: "d1"})
	assert.Error(t, err)
}

func TestShardManager_GetRouterStats(t *testing.T) {
	hs, _ := NewHashSharding(4)
	sm := NewShardManager(hs, 1000)
	stats := sm.GetRouterStats()
	assert.NotNil(t, stats)
}

func TestShardManager_Rebalance(t *testing.T) {
	hs, _ := NewHashSharding(4)
	sm := NewShardManager(hs, 1000)
	err := sm.Rebalance()
	require.NoError(t, err)
}

func TestShardManager_GetRebalanceStatus(t *testing.T) {
	hs, _ := NewHashSharding(4)
	sm := NewShardManager(hs, 1000)
	inProgress, analysis, err := sm.GetRebalanceStatus()
	require.NoError(t, err)
	assert.False(t, inProgress)
	assert.NotNil(t, analysis)
}

func TestShardManager_AddRemoveShard(t *testing.T) {
	hs, _ := NewHashSharding(4)
	sm := NewShardManager(hs, 1000)

	shard := &Shard{ID: 10, Name: "shard-10", Status: ShardStatusActive}
	err := sm.AddShard(shard)
	require.NoError(t, err)
	assert.Equal(t, 5, sm.GetShardCount())

	err = sm.RemoveShard(10)
	require.NoError(t, err)
	assert.Equal(t, 4, sm.GetShardCount())
}

func TestShardManager_GetShards(t *testing.T) {
	hs, _ := NewHashSharding(4)
	sm := NewShardManager(hs, 1000)
	shards := sm.GetShards()
	assert.Equal(t, 4, len(shards))
}

func TestRouterStats(t *testing.T) {
	stats := NewRouterStats()
	require.NotNil(t, stats)

	stats.RecordRoute(0, true, false)
	stats.RecordRoute(1, false, false)
	stats.RecordRoute(0, false, true)

	s := stats.GetStats()
	assert.Equal(t, int64(3), s["totalRoutes"])
	assert.Equal(t, int64(1), s["cacheHits"])
	assert.Equal(t, int64(2), s["cacheMisses"])
	assert.Equal(t, int64(1), s["routeErrors"])
}

func TestShardCache(t *testing.T) {
	cache := newShardCache(10)
	cache.Set("key1", 1)
	cache.Set("key2", 2)

	id, ok := cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, 1, id)

	_, ok = cache.Get("nonexistent")
	assert.False(t, ok)
}

func TestShardCache_Eviction(t *testing.T) {
	cache := newShardCache(4)
	cache.Set("key1", 1)
	cache.Set("key2", 2)
	cache.Set("key3", 3)
	cache.Set("key4", 4)
	cache.Set("key5", 5)
}

func TestShardKey_Struct(t *testing.T) {
	key := &ShardKey{
		DeviceID:  "dev1",
		PointID:   "pt1",
		Timestamp: time.Now(),
		Tags:      map[string]string{"env": "prod"},
	}
	assert.Equal(t, "dev1", key.DeviceID)
	assert.Equal(t, "pt1", key.PointID)
}

func TestRebalanceStats_Struct(t *testing.T) {
	stats := &RebalanceStats{
		StartTime:       time.Now(),
		EndTime:         time.Now().Add(time.Minute),
		ShardsMigrated:  2,
		DataTransferred: 1024,
	}
	assert.Equal(t, 2, stats.ShardsMigrated)
}

func TestBalanceAnalysis_Struct(t *testing.T) {
	analysis := &BalanceAnalysis{
		TotalShards:    4,
		AverageLoad:    1.0,
		MaxLoad:        1.0,
		MinLoad:        1.0,
		Imbalance:      0.0,
		NeedsRebalance: false,
	}
	assert.Equal(t, 4, analysis.TotalShards)
	assert.False(t, analysis.NeedsRebalance)
}
