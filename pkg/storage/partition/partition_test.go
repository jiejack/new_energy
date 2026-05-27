package partition

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPartitionStatus_String(t *testing.T) {
	tests := []struct {
		status PartitionStatus
		exp    string
	}{
		{PartitionStatusActive, "active"},
		{PartitionStatusInactive, "inactive"},
		{PartitionStatusReadOnly, "read_only"},
		{PartitionStatusArchived, "archived"},
		{PartitionStatusDropping, "dropping"},
		{PartitionStatus(99), "unknown"},
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
		{TimeGranularityMonth, "month"},
		{TimeGranularityYear, "year"},
		{TimeGranularity(99), "unknown"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.exp, tt.tg.String())
	}
}

func TestNewTimePartition(t *testing.T) {
	tp, err := NewTimePartition(TimeGranularityDay)
	require.NoError(t, err)
	assert.Equal(t, "time", tp.GetType())
	assert.Equal(t, 0, tp.GetPartitionCount())
}

func TestNewTimePartition_WithOptions(t *testing.T) {
	tp, err := NewTimePartition(TimeGranularityHour,
		WithAutoCreatePartition(false),
		WithRetentionDays(30),
		WithMaxPartitions(100),
	)
	require.NoError(t, err)
	assert.Equal(t, 0, tp.GetPartitionCount())
}

func TestTimePartition_GetPartition_AutoCreate(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityDay)
	key := &PartitionKey{
		DeviceID:  "device1",
		PointID:   "point1",
		Timestamp: time.Now(),
	}
	partition, err := tp.GetPartition(key)
	require.NoError(t, err)
	assert.NotNil(t, partition)
	assert.Equal(t, PartitionStatusActive, partition.Status)
	assert.Equal(t, 1, tp.GetPartitionCount())
}

func TestTimePartition_GetPartition_NilKey(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityDay)
	_, err := tp.GetPartition(nil)
	assert.Equal(t, ErrInvalidPartitionKey, err)
}

func TestTimePartition_GetPartition_AutoCreateDisabled(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityDay, WithAutoCreatePartition(false))
	key := &PartitionKey{Timestamp: time.Now()}
	_, err := tp.GetPartition(key)
	assert.Equal(t, ErrPartitionNotFound, err)
}

func TestTimePartition_GetPartition_SameDay(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityDay)
	now := time.Now()

	key1 := &PartitionKey{Timestamp: now}
	p1, err := tp.GetPartition(key1)
	require.NoError(t, err)

	key2 := &PartitionKey{Timestamp: now.Add(1 * time.Hour)}
	p2, err := tp.GetPartition(key2)
	require.NoError(t, err)

	assert.Equal(t, p1.ID, p2.ID)
	assert.Equal(t, 1, tp.GetPartitionCount())
}

func TestTimePartition_GetPartition_DifferentDays(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityDay)
	now := time.Now()

	key1 := &PartitionKey{Timestamp: now}
	_, err := tp.GetPartition(key1)
	require.NoError(t, err)

	key2 := &PartitionKey{Timestamp: now.AddDate(0, 0, 1)}
	_, err = tp.GetPartition(key2)
	require.NoError(t, err)

	assert.Equal(t, 2, tp.GetPartitionCount())
}

func TestTimePartition_CreatePartition(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityDay, WithAutoCreatePartition(false))
	now := time.Now()
	p := &Partition{
		Name:      "p_test",
		StartTime: now,
		EndTime:   now.AddDate(0, 0, 1),
		Status:    PartitionStatusActive,
	}
	err := tp.CreatePartition(p)
	require.NoError(t, err)
	assert.Equal(t, 1, tp.GetPartitionCount())
}

func TestTimePartition_CreatePartition_InvalidTimeRange(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityDay)
	p := &Partition{
		Name:   "p_test",
		Status: PartitionStatusActive,
	}
	err := tp.CreatePartition(p)
	assert.Equal(t, ErrInvalidTimeRange, err)
}

func TestTimePartition_CreatePartition_Overlap(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityDay, WithAutoCreatePartition(false))
	now := time.Now()
	p1 := &Partition{Name: "p1", StartTime: now, EndTime: now.AddDate(0, 0, 2), Status: PartitionStatusActive}
	tp.CreatePartition(p1)

	p2 := &Partition{Name: "p2", StartTime: now.AddDate(0, 0, 1), EndTime: now.AddDate(0, 0, 3), Status: PartitionStatusActive}
	err := tp.CreatePartition(p2)
	assert.Error(t, err)
}

func TestTimePartition_DropPartition(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityDay)
	key := &PartitionKey{Timestamp: time.Now()}
	p, err := tp.GetPartition(key)
	require.NoError(t, err)

	err = tp.DropPartition(p.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, tp.GetPartitionCount())
}

func TestTimePartition_DropPartition_NotFound(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityDay)
	err := tp.DropPartition(999)
	assert.Equal(t, ErrPartitionNotFound, err)
}

func TestTimePartition_GetActivePartitions(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityDay)
	key := &PartitionKey{Timestamp: time.Now()}
	tp.GetPartition(key)

	active := tp.GetActivePartitions()
	assert.Equal(t, 1, len(active))
}

func TestTimePartition_GetPartitionsByTimeRange(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityDay)
	now := time.Now()
	tp.GetPartition(&PartitionKey{Timestamp: now})
	tp.GetPartition(&PartitionKey{Timestamp: now.AddDate(0, 0, 1)})

	result := tp.GetPartitionsByTimeRange(now, now.AddDate(0, 0, 2))
	assert.Equal(t, 2, len(result))
}

func TestTimePartition_PurgeOldPartitions(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityDay, WithRetentionDays(1))
	now := time.Now()

	tp.GetPartition(&PartitionKey{Timestamp: now})
	tp.GetPartition(&PartitionKey{Timestamp: now.AddDate(0, 0, -5)})

	purged, err := tp.PurgeOldPartitions()
	require.NoError(t, err)
	assert.True(t, len(purged) > 0)
}

func TestTimePartition_PurgeOldPartitions_NoRetention(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityDay, WithRetentionDays(0))
	purged, err := tp.PurgeOldPartitions()
	require.NoError(t, err)
	assert.Nil(t, purged)
}

func TestTimePartition_MaxPartitions(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityHour, WithMaxPartitions(2))
	now := time.Now()

	tp.GetPartition(&PartitionKey{Timestamp: now})
	tp.GetPartition(&PartitionKey{Timestamp: now.Add(1 * time.Hour)})

	_, err := tp.GetPartition(&PartitionKey{Timestamp: now.Add(2 * time.Hour)})
	assert.Error(t, err)
}

func TestTimePartition_GranularityHour(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityHour)
	now := time.Now()
	p, err := tp.GetPartition(&PartitionKey{Timestamp: now})
	require.NoError(t, err)
	assert.Contains(t, p.Name, "p")
}

func TestTimePartition_GranularityMonth(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityMonth)
	now := time.Now()
	p, err := tp.GetPartition(&PartitionKey{Timestamp: now})
	require.NoError(t, err)
	assert.Contains(t, p.Name, "p")
}

func TestTimePartition_GranularityYear(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityYear)
	now := time.Now()
	p, err := tp.GetPartition(&PartitionKey{Timestamp: now})
	require.NoError(t, err)
	assert.Contains(t, p.Name, "p")
}

func TestNewRangePartition(t *testing.T) {
	rp := NewRangePartition()
	assert.Equal(t, "range", rp.GetType())
	assert.Equal(t, 0, rp.GetPartitionCount())
}

func TestRangePartition_CreateRangePartition(t *testing.T) {
	rp := NewRangePartition()
	p := &Partition{
		Name:   "range_p1",
		Status: PartitionStatusActive,
	}
	err := rp.CreateRangePartition(p, []byte("a"), []byte("m"))
	require.NoError(t, err)
	assert.Equal(t, 1, rp.GetPartitionCount())
}

func TestRangePartition_CreateRangePartition_Overlap(t *testing.T) {
	rp := NewRangePartition()
	p1 := &Partition{Name: "p1", Status: PartitionStatusActive}
	rp.CreateRangePartition(p1, []byte("a"), []byte("m"))

	p2 := &Partition{Name: "p2", Status: PartitionStatusActive}
	err := rp.CreateRangePartition(p2, []byte("l"), []byte("z"))
	assert.Error(t, err)
}

func TestRangePartition_GetPartition(t *testing.T) {
	rp := NewRangePartition()
	p1 := &Partition{Name: "p1", Status: PartitionStatusActive}
	rp.CreateRangePartition(p1, []byte("a"), []byte("m"))

	p2 := &Partition{Name: "p2", Status: PartitionStatusActive}
	rp.CreateRangePartition(p2, []byte("m"), []byte("z"))

	key := &PartitionKey{DeviceID: "device1"}
	result, err := rp.GetPartition(key)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestRangePartition_GetPartition_NilKey(t *testing.T) {
	rp := NewRangePartition()
	_, err := rp.GetPartition(nil)
	assert.Equal(t, ErrInvalidPartitionKey, err)
}

func TestRangePartition_GetPartition_NotFound(t *testing.T) {
	rp := NewRangePartition()
	key := &PartitionKey{DeviceID: "zzz"}
	_, err := rp.GetPartition(key)
	assert.Equal(t, ErrPartitionNotFound, err)
}

func TestRangePartition_CreatePartition_UseCreateRange(t *testing.T) {
	rp := NewRangePartition()
	err := rp.CreatePartition(&Partition{})
	assert.Error(t, err)
}

func TestRangePartition_DropPartition(t *testing.T) {
	rp := NewRangePartition()
	p := &Partition{Name: "p1", Status: PartitionStatusActive}
	rp.CreateRangePartition(p, []byte("a"), []byte("z"))

	err := rp.DropPartition(0)
	require.NoError(t, err)
	assert.Equal(t, 0, rp.GetPartitionCount())
}

func TestRangePartition_DropPartition_NotFound(t *testing.T) {
	rp := NewRangePartition()
	err := rp.DropPartition(999)
	assert.Equal(t, ErrPartitionNotFound, err)
}

func TestRangePartition_GetActivePartitions(t *testing.T) {
	rp := NewRangePartition()
	p := &Partition{Name: "p1", Status: PartitionStatusActive}
	rp.CreateRangePartition(p, []byte("a"), []byte("z"))

	active := rp.GetActivePartitions()
	assert.Equal(t, 1, len(active))
}

func TestNewListPartition(t *testing.T) {
	lp := NewListPartition()
	assert.Equal(t, "list", lp.GetType())
	assert.Equal(t, 0, lp.GetPartitionCount())
}

func TestListPartition_CreateListPartition(t *testing.T) {
	lp := NewListPartition()
	p := &Partition{Name: "east", Status: PartitionStatusActive}
	err := lp.CreateListPartition(p, []string{"shanghai", "beijing"})
	require.NoError(t, err)
	assert.Equal(t, 1, lp.GetPartitionCount())
}

func TestListPartition_CreateListPartition_DuplicateValue(t *testing.T) {
	lp := NewListPartition()
	p1 := &Partition{Name: "east", Status: PartitionStatusActive}
	lp.CreateListPartition(p1, []string{"shanghai"})

	p2 := &Partition{Name: "west", Status: PartitionStatusActive}
	err := lp.CreateListPartition(p2, []string{"shanghai"})
	assert.Error(t, err)
}

func TestListPartition_GetPartition(t *testing.T) {
	lp := NewListPartition()
	p1 := &Partition{Name: "east", Status: PartitionStatusActive}
	lp.CreateListPartition(p1, []string{"shanghai", "beijing"})

	key := &PartitionKey{Value: "shanghai"}
	result, err := lp.GetPartition(key)
	require.NoError(t, err)
	assert.Equal(t, "east", result.Name)
}

func TestListPartition_GetPartition_NilKey(t *testing.T) {
	lp := NewListPartition()
	_, err := lp.GetPartition(nil)
	assert.Equal(t, ErrInvalidPartitionKey, err)
}

func TestListPartition_GetPartition_NotFound(t *testing.T) {
	lp := NewListPartition()
	key := &PartitionKey{Value: "nonexistent"}
	_, err := lp.GetPartition(key)
	assert.Equal(t, ErrPartitionNotFound, err)
}

func TestListPartition_CreatePartition_UseCreateList(t *testing.T) {
	lp := NewListPartition()
	err := lp.CreatePartition(&Partition{})
	assert.Error(t, err)
}

func TestListPartition_DropPartition(t *testing.T) {
	lp := NewListPartition()
	p := &Partition{Name: "east", Status: PartitionStatusActive}
	lp.CreateListPartition(p, []string{"shanghai"})

	err := lp.DropPartition(0)
	require.NoError(t, err)
	assert.Equal(t, 0, lp.GetPartitionCount())
}

func TestListPartition_DropPartition_NotFound(t *testing.T) {
	lp := NewListPartition()
	err := lp.DropPartition(999)
	assert.Equal(t, ErrPartitionNotFound, err)
}

func TestListPartition_AddValuesToPartition(t *testing.T) {
	lp := NewListPartition()
	p := &Partition{Name: "east", Status: PartitionStatusActive}
	lp.CreateListPartition(p, []string{"shanghai"})

	err := lp.AddValuesToPartition(0, []string{"hangzhou"})
	require.NoError(t, err)

	key := &PartitionKey{Value: "hangzhou"}
	result, err := lp.GetPartition(key)
	require.NoError(t, err)
	assert.Equal(t, "east", result.Name)
}

func TestListPartition_AddValuesToPartition_Duplicate(t *testing.T) {
	lp := NewListPartition()
	p1 := &Partition{Name: "east", Status: PartitionStatusActive}
	lp.CreateListPartition(p1, []string{"shanghai"})

	p2 := &Partition{Name: "west", Status: PartitionStatusActive}
	lp.CreateListPartition(p2, []string{"chengdu"})

	err := lp.AddValuesToPartition(1, []string{"shanghai"})
	assert.Error(t, err)
}

func TestListPartition_AddValuesToPartition_NotFound(t *testing.T) {
	lp := NewListPartition()
	err := lp.AddValuesToPartition(999, []string{"test"})
	assert.Equal(t, ErrPartitionNotFound, err)
}

func TestListPartition_GetActivePartitions(t *testing.T) {
	lp := NewListPartition()
	p := &Partition{Name: "east", Status: PartitionStatusActive}
	lp.CreateListPartition(p, []string{"shanghai"})

	active := lp.GetActivePartitions()
	assert.Equal(t, 1, len(active))
}

func TestPartitionManager(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityDay)
	pm := NewPartitionManager(tp)
	require.NotNil(t, pm)

	strategy, ok := pm.GetStrategy("default")
	assert.True(t, ok)
	assert.Equal(t, tp, strategy)

	_, ok = pm.GetStrategy("nonexistent")
	assert.False(t, ok)
}

func TestPartitionManager_RegisterStrategy(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityDay)
	pm := NewPartitionManager(tp)

	rp := NewRangePartition()
	pm.RegisterStrategy("range", rp)

	strategy, ok := pm.GetStrategy("range")
	assert.True(t, ok)
	assert.Equal(t, rp, strategy)
}

func TestPartitionManager_GetPartition(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityDay)
	pm := NewPartitionManager(tp)

	key := &PartitionKey{Timestamp: time.Now()}
	p, err := pm.GetPartition(key)
	require.NoError(t, err)
	assert.NotNil(t, p)
}

func TestPartitionManager_GetPartitionWithStrategy(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityDay)
	pm := NewPartitionManager(tp)

	rp := NewRangePartition()
	pm.RegisterStrategy("range", rp)

	_, err := pm.GetPartitionWithStrategy("nonexistent", &PartitionKey{Timestamp: time.Now()})
	assert.Error(t, err)
}

func TestPartitionManager_Stats(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityDay)
	pm := NewPartitionManager(tp)

	pm.GetPartition(&PartitionKey{Timestamp: time.Now()})
	stats := pm.GetStats()
	assert.NotNil(t, stats)
}

func TestPartitionManager_Scheduler(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityDay)
	pm := NewPartitionManager(tp)

	pm.StartScheduler()
	time.Sleep(100 * time.Millisecond)
	pm.StopScheduler()
}

func TestPartitionStats(t *testing.T) {
	stats := NewPartitionStats()
	require.NotNil(t, stats)

	stats.Update("time", []*Partition{
		{Status: PartitionStatusActive, Size: 100, RowCount: 50},
		{Status: PartitionStatusInactive, Size: 200, RowCount: 100},
	})

	s := stats.GetStats()
	assert.Equal(t, 2, s["totalPartitions"])
	assert.Equal(t, 1, s["activePartitions"])
	assert.Equal(t, int64(300), s["totalSize"])
	assert.Equal(t, int64(150), s["totalRows"])
}

func TestAutoPartitionCreator(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityDay)
	creator := NewAutoPartitionCreator(tp, 7*24*time.Hour)
	require.NotNil(t, creator)

	creator.Start()
	time.Sleep(100 * time.Millisecond)
	creator.Stop()
}

func TestPartitionPruner(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityDay, WithRetentionDays(30))
	pruner := NewPartitionPruner(tp, 30*24*time.Hour)
	require.NotNil(t, pruner)

	pruner.SetPurgeHandler(func(partitions []*Partition) {
		_ = partitions
	})

	pruner.Start()
	time.Sleep(100 * time.Millisecond)
	pruner.Stop()
}

func TestGetPartitionInfo(t *testing.T) {
	p := &Partition{
		ID:         1,
		Name:       "p_test",
		Status:     PartitionStatusActive,
		Size:       1024,
		RowCount:   500,
		StartTime:  time.Now(),
		EndTime:    time.Now().AddDate(0, 0, 1),
		CreatedAt:  time.Now(),
		LastAccess: time.Now(),
	}
	info := GetPartitionInfo(p, "time")
	assert.Equal(t, 1, info.ID)
	assert.Equal(t, "p_test", info.Name)
	assert.Equal(t, "time", info.Type)
	assert.Equal(t, "active", info.Status)
	assert.Equal(t, int64(1024), info.Size)
	assert.Equal(t, int64(500), info.RowCount)
}

func TestSortPartitionsByTime(t *testing.T) {
	now := time.Now()
	partitions := []*Partition{
		{StartTime: now.Add(2 * time.Hour)},
		{StartTime: now},
		{StartTime: now.Add(1 * time.Hour)},
	}
	sorted := SortPartitionsByTime(partitions)
	assert.True(t, sorted[0].StartTime.Before(sorted[1].StartTime))
	assert.True(t, sorted[1].StartTime.Before(sorted[2].StartTime))
}

func TestSortPartitionsBySize(t *testing.T) {
	partitions := []*Partition{
		{Size: 300},
		{Size: 100},
		{Size: 200},
	}
	sorted := SortPartitionsBySize(partitions, false)
	assert.Equal(t, int64(100), sorted[0].Size)
	assert.Equal(t, int64(300), sorted[2].Size)

	sortedDesc := SortPartitionsBySize(partitions, true)
	assert.Equal(t, int64(300), sortedDesc[0].Size)
}

func TestSortPartitionsByAccess(t *testing.T) {
	now := time.Now()
	partitions := []*Partition{
		{LastAccess: now.Add(2 * time.Hour)},
		{LastAccess: now},
		{LastAccess: now.Add(1 * time.Hour)},
	}
	sorted := SortPartitionsByAccess(partitions, false)
	assert.True(t, sorted[0].LastAccess.Before(sorted[1].LastAccess))

	sortedDesc := SortPartitionsByAccess(partitions, true)
	assert.True(t, sortedDesc[0].LastAccess.After(sortedDesc[1].LastAccess))
}

func TestPartitionKey_Struct(t *testing.T) {
	key := &PartitionKey{
		DeviceID:  "dev1",
		PointID:   "pt1",
		Timestamp: time.Now(),
		Value:     "test",
		Tags:      map[string]string{"env": "prod"},
	}
	assert.Equal(t, "dev1", key.DeviceID)
	assert.Equal(t, "pt1", key.PointID)
	assert.Equal(t, "test", key.Value)
}

func TestScheduledTask(t *testing.T) {
	task := ScheduledTask{
		Name:     "test_task",
		Interval: time.Hour,
		Handler:  func() error { return nil },
	}
	assert.Equal(t, "test_task", task.Name)
	assert.Equal(t, time.Hour, task.Interval)
}

func TestPartitionScheduler_AddTask(t *testing.T) {
	tp, _ := NewTimePartition(TimeGranularityDay)
	pm := NewPartitionManager(tp)
	scheduler := NewPartitionScheduler(pm)

	scheduler.AddTask("custom_task", time.Minute, func() error { return nil })
	assert.Equal(t, 1, len(scheduler.tasks))

	scheduler.Start()
	time.Sleep(100 * time.Millisecond)
	scheduler.Stop()
}
