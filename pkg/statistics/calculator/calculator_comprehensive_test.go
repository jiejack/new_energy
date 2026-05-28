package calculator

import (
	"context"
	"fmt"
	"testing"
	"time"

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
	require.NotNil(t, db)
	err = db.AutoMigrate(&StatisticsData{}, &StatisticsTask{})
	require.NoError(t, err)
	return db
}

func setupPostgreSQLStorage(t *testing.T) *PostgreSQLStorage {
	t.Helper()
	db := setupSQLiteDB(t)
	return &PostgreSQLStorage{
		db:     db,
		config: StorageConfig{},
	}
}

func setupPostgreSQLStorageWithCompression(t *testing.T) *PostgreSQLStorage {
	t.Helper()
	db := setupSQLiteDB(t)
	return &PostgreSQLStorage{
		db:       db,
		config:   StorageConfig{CompressionEnabled: true, CompressionDays: 30},
		compress: NewDataCompressor(30),
	}
}

func setupPostgreSQLStorageWithArchive(t *testing.T) *PostgreSQLStorage {
	t.Helper()
	db := setupSQLiteDB(t)
	return &PostgreSQLStorage{
		db:      db,
		config:  StorageConfig{ArchiveEnabled: true, ArchiveDays: 90},
		archive: NewDataArchiver(90),
	}
}

func TestPostgreSQLStorage_Save(t *testing.T) {
	storage := setupPostgreSQLStorage(t)
	defer storage.Close()

	ctx := context.Background()
	data := &StatisticsData{
		TaskID:         "task-001",
		Dimension:      "station",
		DimensionValue: "station-001",
		MetricName:     "daily_generation",
		MetricValue:    1000.0,
		PeriodType:     PeriodTypeDay,
		PeriodStart:    time.Now().Add(-24 * time.Hour),
		PeriodEnd:      time.Now(),
	}

	err := storage.Save(ctx, data)
	assert.NoError(t, err)
	assert.NotEmpty(t, data.ID)
	assert.False(t, data.CreatedAt.IsZero())
}

func TestPostgreSQLStorage_Save_WithExistingID(t *testing.T) {
	storage := setupPostgreSQLStorage(t)
	defer storage.Close()

	ctx := context.Background()
	data := &StatisticsData{
		ID:             "custom-id-001",
		TaskID:         "task-001",
		Dimension:      "station",
		DimensionValue: "station-001",
		MetricName:     "daily_generation",
		MetricValue:    1000.0,
		PeriodType:     PeriodTypeDay,
		PeriodStart:    time.Now().Add(-24 * time.Hour),
		PeriodEnd:      time.Now(),
		CreatedAt:      time.Now().Add(-1 * time.Hour),
	}

	err := storage.Save(ctx, data)
	assert.NoError(t, err)
	assert.Equal(t, "custom-id-001", data.ID)
}

func TestPostgreSQLStorage_SaveBatch(t *testing.T) {
	storage := setupPostgreSQLStorage(t)
	defer storage.Close()

	ctx := context.Background()
	now := time.Now()
	data := []*StatisticsData{
		{
			TaskID: "task-001", Dimension: "station", DimensionValue: "s1",
			MetricName: "metric1", MetricValue: 100.0,
			PeriodType: PeriodTypeDay, PeriodStart: now, PeriodEnd: now,
		},
		{
			TaskID: "task-001", Dimension: "station", DimensionValue: "s2",
			MetricName: "metric2", MetricValue: 200.0,
			PeriodType: PeriodTypeDay, PeriodStart: now, PeriodEnd: now,
		},
	}

	err := storage.SaveBatch(ctx, data)
	assert.NoError(t, err)
	for _, d := range data {
		assert.NotEmpty(t, d.ID)
		assert.False(t, d.CreatedAt.IsZero())
	}
}

func TestPostgreSQLStorage_SaveBatch_Empty(t *testing.T) {
	storage := setupPostgreSQLStorage(t)
	defer storage.Close()

	ctx := context.Background()
	err := storage.SaveBatch(ctx, []*StatisticsData{})
	assert.NoError(t, err)
}

func TestPostgreSQLStorage_Query(t *testing.T) {
	storage := setupPostgreSQLStorage(t)
	defer storage.Close()

	ctx := context.Background()
	now := time.Now()

	storage.Save(ctx, &StatisticsData{
		TaskID: "task-001", Dimension: "station", DimensionValue: "s1",
		MetricName: "generation", MetricValue: 100.0,
		PeriodType: PeriodTypeDay, PeriodStart: now.Add(-2 * time.Hour), PeriodEnd: now,
	})
	storage.Save(ctx, &StatisticsData{
		TaskID: "task-001", Dimension: "station", DimensionValue: "s1",
		MetricName: "efficiency", MetricValue: 95.0,
		PeriodType: PeriodTypeDay, PeriodStart: now.Add(-1 * time.Hour), PeriodEnd: now,
	})
	storage.Save(ctx, &StatisticsData{
		TaskID: "task-002", Dimension: "device", DimensionValue: "d1",
		MetricName: "power", MetricValue: 50.0,
		PeriodType: PeriodTypeHour, PeriodStart: now, PeriodEnd: now,
	})

	results, err := storage.Query(ctx, &StatisticsQuery{
		TaskID:    "task-001",
		Dimension: "station",
	})
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	results2, err := storage.Query(ctx, &StatisticsQuery{
		MetricName: "power",
	})
	assert.NoError(t, err)
	assert.Len(t, results2, 1)
	assert.Equal(t, 50.0, results2[0].MetricValue)
}

func TestPostgreSQLStorage_Query_WithAllFilters(t *testing.T) {
	storage := setupPostgreSQLStorage(t)
	defer storage.Close()

	ctx := context.Background()
	now := time.Now()

	storage.Save(ctx, &StatisticsData{
		TaskID: "task-001", Dimension: "station", DimensionValue: "s1",
		MetricName: "generation", MetricValue: 100.0,
		PeriodType: PeriodTypeDay, PeriodStart: now.Add(-24 * time.Hour), PeriodEnd: now,
	})

	results, err := storage.Query(ctx, &StatisticsQuery{
		TaskID:         "task-001",
		Dimension:      "station",
		DimensionValue: "s1",
		MetricName:     "generation",
		PeriodType:     PeriodTypeDay,
		PeriodStart:    now.Add(-48 * time.Hour),
		PeriodEnd:      now,
		OrderBy:        "metric_value",
		OrderDesc:      true,
		Limit:          10,
		Offset:         0,
	})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestPostgreSQLStorage_Query_OrderByDefault(t *testing.T) {
	storage := setupPostgreSQLStorage(t)
	defer storage.Close()

	ctx := context.Background()
	now := time.Now()

	storage.Save(ctx, &StatisticsData{
		TaskID: "task-001", Dimension: "station", DimensionValue: "s1",
		MetricName: "gen", MetricValue: 100.0,
		PeriodType: PeriodTypeDay, PeriodStart: now.Add(-2 * time.Hour), PeriodEnd: now,
	})

	results, err := storage.Query(ctx, &StatisticsQuery{})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestPostgreSQLStorage_QueryLatest(t *testing.T) {
	storage := setupPostgreSQLStorage(t)
	defer storage.Close()

	ctx := context.Background()
	now := time.Now()

	storage.Save(ctx, &StatisticsData{
		Dimension: "station", DimensionValue: "s1",
		MetricName: "generation", MetricValue: 100.0,
		PeriodType: PeriodTypeDay, PeriodStart: now.Add(-2 * time.Hour), PeriodEnd: now,
	})
	storage.Save(ctx, &StatisticsData{
		Dimension: "station", DimensionValue: "s1",
		MetricName: "generation", MetricValue: 200.0,
		PeriodType: PeriodTypeDay, PeriodStart: now.Add(-1 * time.Hour), PeriodEnd: now,
	})

	result, err := storage.QueryLatest(ctx, "station", "s1", "generation")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200.0, result.MetricValue)
}

func TestPostgreSQLStorage_QueryLatest_NotFound(t *testing.T) {
	storage := setupPostgreSQLStorage(t)
	defer storage.Close()

	ctx := context.Background()
	result, err := storage.QueryLatest(ctx, "station", "s1", "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestPostgreSQLStorage_SaveTimeSeries(t *testing.T) {
	storage := setupPostgreSQLStorage(t)
	defer storage.Close()

	db, _ := storage.db.DB()
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS time_series_data (
		id TEXT PRIMARY KEY, point_id TEXT, point_code TEXT,
		timestamp DATETIME, value REAL, quality INTEGER, created_at DATETIME
	)`)
	require.NoError(t, err)

	ctx := context.Background()
	data := &TimeSeriesData{
		PointID:   "point-001",
		PointCode: "P001",
		Unit:      "kW",
		Data: []TimeSeriesPoint{
			{Timestamp: time.Now().Add(-2 * time.Hour), Value: 100.0, Quality: 1},
			{Timestamp: time.Now().Add(-1 * time.Hour), Value: 150.0, Quality: 1},
		},
	}

	err = storage.SaveTimeSeries(ctx, data)
	assert.NoError(t, err)
}

func TestPostgreSQLStorage_SaveTimeSeries_Empty(t *testing.T) {
	storage := setupPostgreSQLStorage(t)
	defer storage.Close()

	ctx := context.Background()
	data := &TimeSeriesData{
		PointID: "point-001",
		Data:    []TimeSeriesPoint{},
	}

	err := storage.SaveTimeSeries(ctx, data)
	assert.NoError(t, err)
}

func TestPostgreSQLStorage_QueryTimeSeries(t *testing.T) {
	storage := setupPostgreSQLStorage(t)
	defer storage.Close()

	db, _ := storage.db.DB()
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS time_series_data (
		id TEXT PRIMARY KEY, point_id TEXT, point_code TEXT,
		timestamp DATETIME, value REAL, quality INTEGER, created_at DATETIME
	)`)
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now()
	tsData := &TimeSeriesData{
		PointID:   "point-001",
		PointCode: "P001",
		Unit:      "kW",
		Data: []TimeSeriesPoint{
			{Timestamp: now.Add(-2 * time.Hour), Value: 100.0, Quality: 1},
			{Timestamp: now.Add(-1 * time.Hour), Value: 150.0, Quality: 1},
		},
	}
	err = storage.SaveTimeSeries(ctx, tsData)
	require.NoError(t, err)

	result, err := storage.QueryTimeSeries(ctx, "point-001", now.Add(-3*time.Hour), now)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "point-001", result.PointID)
	assert.Len(t, result.Data, 2)
}

func TestPostgreSQLStorage_QueryTimeSeries_Empty(t *testing.T) {
	storage := setupPostgreSQLStorage(t)
	defer storage.Close()

	db, _ := storage.db.DB()
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS time_series_data (
		id TEXT PRIMARY KEY, point_id TEXT, point_code TEXT,
		timestamp DATETIME, value REAL, quality INTEGER, created_at DATETIME
	)`)
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now()
	result, err := storage.QueryTimeSeries(ctx, "point-999", now.Add(-3*time.Hour), now)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Data)
}

func TestPostgreSQLStorage_SaveTask(t *testing.T) {
	storage := setupPostgreSQLStorage(t)
	defer storage.Close()

	ctx := context.Background()
	task := &StatisticsTask{
		Name:           "Test Task",
		TaskType:       "daily",
		CronExpression: "0 0 * * *",
		Config:         `{"key": "value"}`,
		Enabled:        true,
	}

	err := storage.SaveTask(ctx, task)
	assert.NoError(t, err)
	assert.NotEmpty(t, task.ID)
	assert.False(t, task.CreatedAt.IsZero())
	assert.False(t, task.UpdatedAt.IsZero())
}

func TestPostgreSQLStorage_GetTask(t *testing.T) {
	storage := setupPostgreSQLStorage(t)
	defer storage.Close()

	ctx := context.Background()
	task := &StatisticsTask{
		Name:           "Test Task",
		TaskType:       "daily",
		CronExpression: "0 0 * * *",
		Enabled:        true,
	}
	storage.SaveTask(ctx, task)

	result, err := storage.GetTask(ctx, task.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Task", result.Name)
	assert.Equal(t, "daily", result.TaskType)
}

func TestPostgreSQLStorage_GetTask_NotFound(t *testing.T) {
	storage := setupPostgreSQLStorage(t)
	defer storage.Close()

	ctx := context.Background()
	result, err := storage.GetTask(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestPostgreSQLStorage_ListTasks(t *testing.T) {
	storage := setupPostgreSQLStorage(t)
	defer storage.Close()

	ctx := context.Background()
	storage.SaveTask(ctx, &StatisticsTask{Name: "Task1", TaskType: "daily", CronExpression: "0 0 * * *", Enabled: true})
	storage.SaveTask(ctx, &StatisticsTask{Name: "Task2", TaskType: "hourly", CronExpression: "0 * * * *", Enabled: false})
	storage.SaveTask(ctx, &StatisticsTask{Name: "Task3", TaskType: "daily", CronExpression: "0 0 * * *", Enabled: true})

	tasks, err := storage.ListTasks(ctx, nil)
	assert.NoError(t, err)
	assert.Len(t, tasks, 3)

	enabled := true
	enabledTasks, err := storage.ListTasks(ctx, &enabled)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(enabledTasks), 2)

	disabled := false
	disabledTasks, err := storage.ListTasks(ctx, &disabled)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(disabledTasks), 0)
}

func TestPostgreSQLStorage_UpdateTaskRunTime(t *testing.T) {
	storage := setupPostgreSQLStorage(t)
	defer storage.Close()

	ctx := context.Background()
	task := &StatisticsTask{Name: "Task1", TaskType: "daily", CronExpression: "0 0 * * *", Enabled: true}
	storage.SaveTask(ctx, task)

	lastRun := time.Now().Add(-1 * time.Hour)
	nextRun := time.Now().Add(23 * time.Hour)
	err := storage.UpdateTaskRunTime(ctx, task.ID, lastRun, nextRun)
	assert.NoError(t, err)

	updated, _ := storage.GetTask(ctx, task.ID)
	assert.NotNil(t, updated)
	assert.NotNil(t, updated.LastRun)
}

func TestPostgreSQLStorage_CompressData_NotEnabled(t *testing.T) {
	storage := setupPostgreSQLStorage(t)
	defer storage.Close()

	ctx := context.Background()
	err := storage.CompressData(ctx, time.Now())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "compression not enabled")
}

func TestPostgreSQLStorage_CompressData_Enabled(t *testing.T) {
	storage := setupPostgreSQLStorageWithCompression(t)
	defer storage.Close()

	ctx := context.Background()
	now := time.Now()
	before := now.Add(-1 * time.Hour)

	storage.Save(ctx, &StatisticsData{
		Dimension: "station", DimensionValue: "s1",
		MetricName: "generation", MetricValue: 100.0,
		PeriodType: PeriodTypeDay, PeriodStart: now.Add(-48 * time.Hour), PeriodEnd: before,
	})

	err := storage.CompressData(ctx, before)
	assert.NoError(t, err)
}

func TestPostgreSQLStorage_CompressData_NoDataToCompress(t *testing.T) {
	storage := setupPostgreSQLStorageWithCompression(t)
	defer storage.Close()

	ctx := context.Background()
	err := storage.CompressData(ctx, time.Now().Add(-1*time.Hour))
	assert.NoError(t, err)
}

func TestPostgreSQLStorage_ArchiveData_NotEnabled(t *testing.T) {
	storage := setupPostgreSQLStorage(t)
	defer storage.Close()

	ctx := context.Background()
	err := storage.ArchiveData(ctx, time.Now())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "archive not enabled")
}

func TestPostgreSQLStorage_ArchiveData_Enabled(t *testing.T) {
	storage := setupPostgreSQLStorageWithArchive(t)
	defer storage.Close()

	ctx := context.Background()
	now := time.Now()
	before := now.Add(-1 * time.Hour)

	storage.Save(ctx, &StatisticsData{
		Dimension: "station", DimensionValue: "s1",
		MetricName: "generation", MetricValue: 100.0,
		PeriodType: PeriodTypeDay, PeriodStart: now.Add(-48 * time.Hour), PeriodEnd: before,
	})

	err := storage.ArchiveData(ctx, before)
	assert.NoError(t, err)
}

func TestPostgreSQLStorage_Ping(t *testing.T) {
	storage := setupPostgreSQLStorage(t)
	defer storage.Close()

	ctx := context.Background()
	err := storage.Ping(ctx)
	assert.NoError(t, err)
}

func TestPostgreSQLStorage_Close(t *testing.T) {
	storage := setupPostgreSQLStorage(t)
	err := storage.Close()
	assert.NoError(t, err)
}

func TestDataCompressor_New(t *testing.T) {
	compressor := NewDataCompressor(30)
	assert.NotNil(t, compressor)
	assert.Equal(t, 30, compressor.compressionDays)
}

func TestDataCompressor_Compress(t *testing.T) {
	db := setupSQLiteDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	compressor := NewDataCompressor(30)
	ctx := context.Background()
	now := time.Now()
	before := now.Add(-1 * time.Hour)

	db.Create(&StatisticsData{
		ID: "sd1", Dimension: "station", DimensionValue: "s1",
		MetricName: "generation", MetricValue: 100.0,
		PeriodType: PeriodTypeDay, PeriodStart: now.Add(-48 * time.Hour), PeriodEnd: before,
		CreatedAt: now.Add(-48 * time.Hour),
	})
	db.Create(&StatisticsData{
		ID: "sd2", Dimension: "station", DimensionValue: "s1",
		MetricName: "generation", MetricValue: 200.0,
		PeriodType: PeriodTypeDay, PeriodStart: now.Add(-24 * time.Hour), PeriodEnd: before,
		CreatedAt: now.Add(-24 * time.Hour),
	})

	err := compressor.Compress(ctx, db, before)
	assert.NoError(t, err)
}

func TestDataCompressor_Compress_EmptyData(t *testing.T) {
	db := setupSQLiteDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	compressor := NewDataCompressor(30)
	ctx := context.Background()

	err := compressor.Compress(ctx, db, time.Now())
	assert.NoError(t, err)
}

func TestDataArchiver_New(t *testing.T) {
	archiver := NewDataArchiver(90)
	assert.NotNil(t, archiver)
	assert.Equal(t, 90, archiver.archiveDays)
}

func TestDataArchiver_Archive(t *testing.T) {
	db := setupSQLiteDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	archiver := NewDataArchiver(90)
	ctx := context.Background()
	now := time.Now()
	before := now.Add(-1 * time.Hour)

	db.Create(&StatisticsData{
		ID: "sd1", Dimension: "station", DimensionValue: "s1",
		MetricName: "generation", MetricValue: 100.0,
		PeriodType: PeriodTypeDay, PeriodStart: now.Add(-48 * time.Hour), PeriodEnd: before,
		CreatedAt: now.Add(-48 * time.Hour),
	})

	err := archiver.Archive(ctx, db, before)
	assert.NoError(t, err)
}

func TestTimeSeriesDBStorage_All(t *testing.T) {
	pgStorage := setupPostgreSQLStorage(t)
	defer pgStorage.Close()

	tsStorage, err := NewTimeSeriesDBStorage(pgStorage, StorageConfig{})
	require.NoError(t, err)
	require.NotNil(t, tsStorage)

	ctx := context.Background()

	t.Run("Save", func(t *testing.T) {
		data := &StatisticsData{
			TaskID: "task-001", Dimension: "station", DimensionValue: "s1",
			MetricName: "generation", MetricValue: 100.0,
			PeriodType: PeriodTypeDay, PeriodStart: time.Now(), PeriodEnd: time.Now(),
		}
		err := tsStorage.Save(ctx, data)
		assert.NoError(t, err)
	})

	t.Run("SaveBatch", func(t *testing.T) {
		data := []*StatisticsData{
			{TaskID: "task-002", Dimension: "device", DimensionValue: "d1",
				MetricName: "power", MetricValue: 50.0,
				PeriodType: PeriodTypeHour, PeriodStart: time.Now(), PeriodEnd: time.Now()},
		}
		err := tsStorage.SaveBatch(ctx, data)
		assert.NoError(t, err)
	})

	t.Run("Query", func(t *testing.T) {
		results, err := tsStorage.Query(ctx, &StatisticsQuery{TaskID: "task-001"})
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
	})

	t.Run("QueryLatest", func(t *testing.T) {
		result, err := tsStorage.QueryLatest(ctx, "station", "s1", "generation")
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("SaveTimeSeries_Empty", func(t *testing.T) {
		err := tsStorage.SaveTimeSeries(ctx, &TimeSeriesData{PointID: "p1", Data: []TimeSeriesPoint{}})
		assert.NoError(t, err)
	})

	t.Run("SaveTask", func(t *testing.T) {
		task := &StatisticsTask{Name: "TS Task", TaskType: "daily", CronExpression: "0 0 * * *", Enabled: true}
		err := tsStorage.SaveTask(ctx, task)
		assert.NoError(t, err)
	})

	t.Run("GetTask", func(t *testing.T) {
		task := &StatisticsTask{Name: "TS Task2", TaskType: "daily", CronExpression: "0 0 * * *", Enabled: true}
		tsStorage.SaveTask(ctx, task)
		result, err := tsStorage.GetTask(ctx, task.ID)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("ListTasks", func(t *testing.T) {
		tasks, err := tsStorage.ListTasks(ctx, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, tasks)
	})

	t.Run("UpdateTaskRunTime", func(t *testing.T) {
		task := &StatisticsTask{Name: "TS Task3", TaskType: "daily", CronExpression: "0 0 * * *", Enabled: true}
		tsStorage.SaveTask(ctx, task)
		err := tsStorage.UpdateTaskRunTime(ctx, task.ID, time.Now(), time.Now().Add(24*time.Hour))
		assert.NoError(t, err)
	})

	t.Run("CompressData_NotEnabled", func(t *testing.T) {
		err := tsStorage.CompressData(ctx, time.Now())
		assert.Error(t, err)
	})

	t.Run("ArchiveData_NotEnabled", func(t *testing.T) {
		err := tsStorage.ArchiveData(ctx, time.Now())
		assert.Error(t, err)
	})

	t.Run("Ping", func(t *testing.T) {
		err := tsStorage.Ping(ctx)
		assert.NoError(t, err)
	})

	t.Run("Close", func(t *testing.T) {
		err := tsStorage.Close()
		assert.NoError(t, err)
	})
}

func TestSqrt(t *testing.T) {
	assert.Equal(t, 0.0, sqrt(-1))
	assert.Equal(t, 0.0, sqrt(0))
	assert.InDelta(t, 2.0, sqrt(4), 0.001)
	assert.InDelta(t, 3.0, sqrt(9), 0.001)
	assert.InDelta(t, 1.414, sqrt(2), 0.01)
}

func TestStatisticsResult_ToStatisticsData_NilMetadata(t *testing.T) {
	result := &StatisticsResult{
		Dimension:      "station",
		DimensionValue: "s1",
		Metrics: map[string]float64{
			"generation": 1000.0,
		},
		Metadata:    nil,
		PeriodStart: time.Now(),
		PeriodEnd:   time.Now(),
		PeriodType:  PeriodTypeDay,
	}

	data := result.ToStatisticsData("task-001")
	assert.Len(t, data, 1)
	assert.Equal(t, "", data[0].Metadata)
}

func TestStatisticsResult_ToStatisticsData_WithMetadata(t *testing.T) {
	result := &StatisticsResult{
		Dimension:      "station",
		DimensionValue: "s1",
		Metrics: map[string]float64{
			"generation": 1000.0,
			"efficiency": 95.5,
		},
		Metadata: map[string]interface{}{
			"station_name": "Test Station",
		},
		PeriodStart: time.Now(),
		PeriodEnd:   time.Now(),
		PeriodType:  PeriodTypeDay,
	}

	data := result.ToStatisticsData("task-001")
	assert.Len(t, data, 2)
	for _, d := range data {
		assert.NotEmpty(t, d.Metadata)
	}
}

func TestCalculateAggregated_SingleValue(t *testing.T) {
	stats := CalculateAggregated([]float64{42.0})
	assert.Equal(t, 42.0, stats.Sum)
	assert.Equal(t, 42.0, stats.Avg)
	assert.Equal(t, 42.0, stats.Min)
	assert.Equal(t, 42.0, stats.Max)
	assert.Equal(t, int64(1), stats.Count)
	assert.InDelta(t, 0.0, stats.Variance, 0.001)
	assert.InDelta(t, 0.0, stats.StdDev, 0.001)
}

func TestCalculateAggregated_VarianceAndStdDev(t *testing.T) {
	values := []float64{2.0, 4.0, 4.0, 4.0, 5.0, 5.0, 7.0, 9.0}
	stats := CalculateAggregated(values)
	assert.Equal(t, 40.0, stats.Sum)
	assert.InDelta(t, 5.0, stats.Avg, 0.001)
	assert.InDelta(t, 4.0, stats.Variance, 0.001)
	assert.InDelta(t, 2.0, stats.StdDev, 0.001)
}

func TestStationCalculator_CalculateStationStatistics(t *testing.T) {
	now := time.Now()
	lastOnline := now.Add(-1 * time.Hour)
	acknowledgedAt := now.Add(-30 * time.Minute)

	provider := &MockDataProvider{
		stations: []StationInfo{
			{ID: "station-001", Code: "ST001", Name: "Test Station", Type: "solar", Capacity: 1000.0, Status: 1},
		},
		devices: []DeviceInfo{
			{ID: "device-001", Code: "D001", Name: "Dev1", Type: "inverter", StationID: "station-001", RatedPower: 100.0, Status: 1, LastOnline: &lastOnline},
			{ID: "device-002", Code: "D002", Name: "Dev2", Type: "inverter", StationID: "station-001", RatedPower: 100.0, Status: 2},
		},
		alarms: []AlarmInfo{
			{ID: "alarm-001", StationID: "station-001", DeviceID: "device-001", Type: "fault", Level: 3, Status: 2, TriggeredAt: now.Add(-2 * time.Hour), AcknowledgedAt: &acknowledgedAt},
		},
		points: []PointInfo{
			{ID: "point-001", Code: "P001", Name: "Generation", Type: "generation", DeviceID: "device-001", Unit: "kWh"},
			{ID: "point-eff-001", Code: "system_eff", Name: "Efficiency", Type: "efficiency", DeviceID: "device-001", Unit: "%"},
		},
		timeSeriesData: map[string][]TimeSeriesPoint{
			"point-001": {
				{Timestamp: now.Add(-2 * time.Hour), Value: 100.0, Quality: 1},
				{Timestamp: now.Add(-1 * time.Hour), Value: 200.0, Quality: 1},
			},
			"point-eff-001": {
				{Timestamp: now.Add(-1 * time.Hour), Value: 95.0, Quality: 1},
			},
		},
	}

	storage := &MockStatisticsStorage{}
	config := StationCalculatorConfig{
		ParallelWorkers: 2,
		DataProvider:    provider,
	}

	calc := NewStationCalculator(config, storage)
	start := now.Add(-24 * time.Hour)
	end := now

	stats, err := calc.CalculateStationStatistics(context.Background(), "station-001", PeriodTypeDay, start, end)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "station-001", stats.StationID)
	assert.Equal(t, "ST001", stats.StationCode)
	assert.Equal(t, "Test Station", stats.StationName)
	assert.Equal(t, 2, stats.DeviceCount)
	assert.Equal(t, 1, stats.OnlineDeviceCount)
}

func TestStationCalculator_CalculateStationStatistics_WithCache(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		stations: []StationInfo{
			{ID: "station-001", Code: "ST001", Name: "Test Station", Type: "solar", Capacity: 1000.0, Status: 1},
		},
		devices: []DeviceInfo{},
		points:  []PointInfo{},
		timeSeriesData: map[string][]TimeSeriesPoint{},
	}

	storage := &MockStatisticsStorage{}
	config := StationCalculatorConfig{
		CacheEnabled: true,
		CacheTTL:     5 * time.Minute,
		DataProvider: provider,
	}

	calc := NewStationCalculator(config, storage)
	start := now.Add(-24 * time.Hour)
	end := now

	stats1, err := calc.CalculateStationStatistics(context.Background(), "station-001", PeriodTypeDay, start, end)
	assert.NoError(t, err)
	assert.NotNil(t, stats1)

	stats2, err := calc.CalculateStationStatistics(context.Background(), "station-001", PeriodTypeDay, start, end)
	assert.NoError(t, err)
	assert.NotNil(t, stats2)
	assert.Equal(t, stats1.StationID, stats2.StationID)
}

func TestStationCalculator_CalculateStationStatistics_MonthlyYearly(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		stations: []StationInfo{
			{ID: "station-001", Code: "ST001", Name: "Test Station", Type: "solar", Capacity: 1000.0, Status: 1},
		},
		devices: []DeviceInfo{},
		points:  []PointInfo{},
		timeSeriesData: map[string][]TimeSeriesPoint{},
	}

	storage := &MockStatisticsStorage{}
	config := StationCalculatorConfig{DataProvider: provider}
	calc := NewStationCalculator(config, storage)

	stats, err := calc.CalculateStationStatistics(context.Background(), "station-001", PeriodTypeMonth, now.Add(-720*time.Hour), now)
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	stats2, err := calc.CalculateStationStatistics(context.Background(), "station-001", PeriodTypeYear, now.Add(-8760*time.Hour), now)
	assert.NoError(t, err)
	assert.NotNil(t, stats2)
}

func TestStationCalculator_CalculateStationStatistics_GetStationError(t *testing.T) {
	provider := &ErrorDataProvider{}
	storage := &MockStatisticsStorage{}
	config := StationCalculatorConfig{DataProvider: provider}
	calc := NewStationCalculator(config, storage)

	now := time.Now()
	_, err := calc.CalculateStationStatistics(context.Background(), "nonexistent", PeriodTypeDay, now, now)
	assert.Error(t, err)
}

type ErrorDataProvider struct{}

func (e *ErrorDataProvider) GetTimeSeriesData(ctx context.Context, pointIDs []string, start, end time.Time) (map[string][]TimeSeriesPoint, error) {
	return nil, fmt.Errorf("error")
}

func (e *ErrorDataProvider) GetDevices(ctx context.Context, stationID string) ([]DeviceInfo, error) {
	return nil, fmt.Errorf("error")
}

func (e *ErrorDataProvider) GetAllDevices(ctx context.Context) ([]DeviceInfo, error) {
	return nil, fmt.Errorf("error")
}

func (e *ErrorDataProvider) GetStation(ctx context.Context, stationID string) (*StationInfo, error) {
	return nil, fmt.Errorf("station not found")
}

func (e *ErrorDataProvider) GetAllStations(ctx context.Context) ([]StationInfo, error) {
	return nil, fmt.Errorf("error")
}

func (e *ErrorDataProvider) GetAlarms(ctx context.Context, stationID string, start, end time.Time) ([]AlarmInfo, error) {
	return nil, fmt.Errorf("error")
}

func (e *ErrorDataProvider) GetPoints(ctx context.Context, stationID string, pointType string) ([]PointInfo, error) {
	return nil, fmt.Errorf("error")
}

func TestStationCalculator_CalculateGeneration_NoPoints(t *testing.T) {
	provider := &MockDataProvider{
		stations: []StationInfo{{ID: "station-001", Code: "ST001", Name: "Test", Type: "solar", Capacity: 1000.0}},
		points:   []PointInfo{},
		timeSeriesData: map[string][]TimeSeriesPoint{},
	}
	storage := &MockStatisticsStorage{}
	config := StationCalculatorConfig{DataProvider: provider}
	calc := NewStationCalculator(config, storage)

	now := time.Now()
	stats, err := calc.CalculateGeneration(context.Background(), "station-001", PeriodTypeDay, now, now)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 0.0, stats.TotalGeneration)
}

func TestStationCalculator_CalculateGeneration_WithCapacity(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		stations: []StationInfo{
			{ID: "station-001", Code: "ST001", Name: "Test", Type: "solar", Capacity: 1000.0},
		},
		points: []PointInfo{
			{ID: "point-001", Code: "P001", Name: "Gen", Type: "generation", DeviceID: "d1", Unit: "kWh"},
		},
		timeSeriesData: map[string][]TimeSeriesPoint{
			"point-001": {
				{Timestamp: now.Add(-2 * time.Hour), Value: 100.0, Quality: 1},
				{Timestamp: now.Add(-1 * time.Hour), Value: 200.0, Quality: 1},
			},
		},
	}
	storage := &MockStatisticsStorage{}
	config := StationCalculatorConfig{DataProvider: provider}
	calc := NewStationCalculator(config, storage)

	stats, err := calc.CalculateGeneration(context.Background(), "station-001", PeriodTypeDay, now.Add(-24*time.Hour), now)
	assert.NoError(t, err)
	assert.True(t, stats.GenerationHours > 0)
	assert.True(t, stats.CapacityFactor > 0)
}

func TestStationCalculator_CalculateDeviceRunRate_NoDevices(t *testing.T) {
	provider := &MockDataProvider{devices: []DeviceInfo{}}
	storage := &MockStatisticsStorage{}
	config := StationCalculatorConfig{DataProvider: provider}
	calc := NewStationCalculator(config, storage)

	now := time.Now()
	stats, err := calc.CalculateDeviceRunRate(context.Background(), "station-001", now, now)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 0, stats.TotalDevices)
}

func TestStationCalculator_CalculateDeviceRunRate_WithOffline(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "d1", Code: "D1", Name: "Dev1", Type: "inverter", StationID: "s1", Status: 0},
		},
	}
	storage := &MockStatisticsStorage{}
	config := StationCalculatorConfig{DataProvider: provider}
	calc := NewStationCalculator(config, storage)

	stats, err := calc.CalculateDeviceRunRate(context.Background(), "s1", now, now)
	assert.NoError(t, err)
	assert.Equal(t, 1, stats.TotalDevices)
	assert.Equal(t, 0, stats.OnlineDevices)
}

func TestStationCalculator_CalculateAlarmStats_NoAlarms(t *testing.T) {
	provider := &MockDataProvider{alarms: []AlarmInfo{}}
	storage := &MockStatisticsStorage{}
	config := StationCalculatorConfig{DataProvider: provider}
	calc := NewStationCalculator(config, storage)

	now := time.Now()
	stats, err := calc.CalculateAlarmStats(context.Background(), "s1", now, now)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), stats.TotalCount)
}

func TestStationCalculator_CalculateAlarmStats_AllLevels(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		alarms: []AlarmInfo{
			{ID: "a1", StationID: "s1", Level: 1, Status: 1, TriggeredAt: now.Add(-1 * time.Hour)},
			{ID: "a2", StationID: "s1", Level: 4, Status: 3, TriggeredAt: now.Add(-2 * time.Hour), ClearedAt: &now},
		},
	}
	storage := &MockStatisticsStorage{}
	config := StationCalculatorConfig{DataProvider: provider}
	calc := NewStationCalculator(config, storage)

	stats, err := calc.CalculateAlarmStats(context.Background(), "s1", now.Add(-3*time.Hour), now)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), stats.InfoCount)
	assert.Equal(t, int64(1), stats.CriticalCount)
	assert.True(t, stats.AvgClearTime > 0)
}

func TestStationCalculator_CalculateEfficiency_DefaultMethod(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		stations: []StationInfo{
			{ID: "s1", Code: "ST001", Name: "Test", Type: "solar", Capacity: 1000.0},
		},
		points: []PointInfo{
			{ID: "input-001", Code: "IN", Name: "Input", Type: "input_power", DeviceID: "d1", Unit: "kW"},
			{ID: "output-001", Code: "OUT", Name: "Output", Type: "output_power", DeviceID: "d1", Unit: "kW"},
		},
		timeSeriesData: map[string][]TimeSeriesPoint{
			"input-001": {
				{Timestamp: now.Add(-1 * time.Hour), Value: 1000.0, Quality: 1},
			},
			"output-001": {
				{Timestamp: now.Add(-1 * time.Hour), Value: 950.0, Quality: 1},
			},
		},
	}
	storage := &MockStatisticsStorage{}
	config := StationCalculatorConfig{DataProvider: provider}
	calc := NewStationCalculator(config, storage)

	stats, err := calc.CalculateEfficiency(context.Background(), "s1", now.Add(-2*time.Hour), now)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.True(t, stats.SystemEfficiency > 0)
}

func TestStationCalculator_CalculateEfficiency_DefaultMethod_NoInputPoints(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		stations: []StationInfo{{ID: "s1", Code: "ST001", Name: "Test", Type: "solar", Capacity: 1000.0}},
		points:   []PointInfo{},
		timeSeriesData: map[string][]TimeSeriesPoint{},
	}
	storage := &MockStatisticsStorage{}
	config := StationCalculatorConfig{DataProvider: provider}
	calc := NewStationCalculator(config, storage)

	stats, err := calc.CalculateEfficiency(context.Background(), "s1", now, now)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 0.0, stats.SystemEfficiency)
}

func TestStationCalculator_CalculateEfficiency_DefaultMethod_EmptyOutput(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		stations: []StationInfo{{ID: "s1", Code: "ST001", Name: "Test", Type: "solar", Capacity: 1000.0}},
		points: []PointInfo{
			{ID: "input-001", Code: "IN", Name: "Input", Type: "input_power", DeviceID: "d1", Unit: "kW"},
		},
		timeSeriesData: map[string][]TimeSeriesPoint{},
	}
	storage := &MockStatisticsStorage{}
	config := StationCalculatorConfig{DataProvider: provider}
	calc := NewStationCalculator(config, storage)

	stats, err := calc.CalculateEfficiency(context.Background(), "s1", now, now)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestStationCalculator_CalculateEfficiency_WithInverterEff(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		stations: []StationInfo{{ID: "s1", Code: "ST001", Name: "Test", Type: "solar", Capacity: 1000.0}},
		points: []PointInfo{
			{ID: "eff-001", Code: "inverter_eff", Name: "Inverter Eff", Type: "efficiency", DeviceID: "d1", Unit: "%"},
			{ID: "eff-002", Code: "transformer_eff", Name: "Transformer Eff", Type: "efficiency", DeviceID: "d1", Unit: "%"},
		},
		timeSeriesData: map[string][]TimeSeriesPoint{
			"eff-001": {
				{Timestamp: now.Add(-1 * time.Hour), Value: 98.0, Quality: 1},
			},
			"eff-002": {
				{Timestamp: now.Add(-1 * time.Hour), Value: 99.0, Quality: 1},
			},
		},
	}
	storage := &MockStatisticsStorage{}
	config := StationCalculatorConfig{DataProvider: provider}
	calc := NewStationCalculator(config, storage)

	stats, err := calc.CalculateEfficiency(context.Background(), "s1", now.Add(-2*time.Hour), now)
	assert.NoError(t, err)
	assert.True(t, stats.InverterEfficiency > 0)
	assert.True(t, stats.TransformerEfficiency > 0)
}

func TestStationCalculator_CalculateAllStations(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		stations: []StationInfo{
			{ID: "s1", Code: "ST001", Name: "Station1", Type: "solar", Capacity: 1000.0},
			{ID: "s2", Code: "ST002", Name: "Station2", Type: "wind", Capacity: 2000.0},
		},
		devices: []DeviceInfo{},
		points:  []PointInfo{},
		timeSeriesData: map[string][]TimeSeriesPoint{},
	}
	storage := &MockStatisticsStorage{}
	config := StationCalculatorConfig{
		ParallelWorkers: 2,
		DataProvider:    provider,
	}
	calc := NewStationCalculator(config, storage)

	results, err := calc.CalculateAllStations(context.Background(), PeriodTypeDay, now.Add(-24*time.Hour), now)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestStationCalculator_CalculateAllStations_DefaultWorkers(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		stations: []StationInfo{
			{ID: "s1", Code: "ST001", Name: "Station1", Type: "solar", Capacity: 1000.0},
		},
		devices: []DeviceInfo{},
		points:  []PointInfo{},
		timeSeriesData: map[string][]TimeSeriesPoint{},
	}
	storage := &MockStatisticsStorage{}
	config := StationCalculatorConfig{
		ParallelWorkers: 0,
		DataProvider:    provider,
	}
	calc := NewStationCalculator(config, storage)

	results, err := calc.CalculateAllStations(context.Background(), PeriodTypeDay, now.Add(-24*time.Hour), now)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestStationCalculator_SaveStatistics(t *testing.T) {
	now := time.Now()
	storage := &MockStatisticsStorage{}
	config := StationCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewStationCalculator(config, storage)

	stats := &StationStatistics{
		StationID:         "s1",
		StationCode:       "ST001",
		StationName:       "Test Station",
		StationType:       "solar",
		Capacity:          1000.0,
		DailyGeneration:   500.0,
		MonthlyGeneration: 15000.0,
		YearlyGeneration:  180000.0,
		DeviceCount:       10,
		OnlineDeviceCount: 8,
		DeviceRunRate:     80.0,
		TotalAlarmCount:   5,
		ActiveAlarmCount:  2,
		CriticalAlarmCount: 1,
		Efficiency:        95.0,
		PerformanceRatio:  0.95,
		EquivalentHours:   1200.0,
		PeriodStart:       now.Add(-24 * time.Hour),
		PeriodEnd:         now,
	}

	err := calc.SaveStatistics(context.Background(), "task-001", stats)
	assert.NoError(t, err)
	assert.NotEmpty(t, storage.data)
}

func TestStatisticsCache_Expiry(t *testing.T) {
	cache := NewStatisticsCache(10 * time.Millisecond)
	cache.Set("key1", "value1")

	value, ok := cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", value)

	time.Sleep(20 * time.Millisecond)
	_, ok = cache.Get("key1")
	assert.False(t, ok)
}

func TestStatisticsCache_Concurrent(t *testing.T) {
	cache := NewStatisticsCache(5 * time.Minute)
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			cache.Set(fmt.Sprintf("key%d", i), i)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			cache.Get(fmt.Sprintf("key%d", i))
		}
		done <- true
	}()

	<-done
	<-done
}

func TestMathHelpers(t *testing.T) {
	assert.Equal(t, 5.0, mathAbs(5.0))
	assert.Equal(t, 5.0, mathAbs(-5.0))
	assert.Equal(t, 0.0, mathAbs(0.0))

	assert.InDelta(t, 8.0, mathPow(2, 3), 0.001)
	assert.InDelta(t, 1.0, mathPow(5, 0), 0.001)

	assert.InDelta(t, 2.0, mathSqrt(4), 0.001)
	assert.InDelta(t, 3.0, mathSqrt(9), 0.001)
}

func TestContains_EdgeCases(t *testing.T) {
	assert.True(t, contains("abc", "abc"))
	assert.True(t, contains("abc", "a"))
	assert.True(t, contains("abc", "c"))
	assert.False(t, contains("ab", "abc"))
	assert.True(t, contains("", ""))
}

func TestDeviceCalculator_CalculateDeviceStatus(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "device-001", Code: "D001", Name: "Dev1", Type: "inverter", StationID: "station-001", Status: 1},
		},
		alarms: []AlarmInfo{
			{ID: "a1", StationID: "station-001", DeviceID: "device-001", Type: "device", Level: 3, Status: 1, TriggeredAt: now.Add(-1 * time.Hour)},
		},
	}
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{DataProvider: provider}
	calc := NewDeviceCalculator(config, storage)

	stats, err := calc.CalculateDeviceStatus(context.Background(), "device-001", now.Add(-24*time.Hour), now)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "device-001", stats.DeviceID)
	assert.Equal(t, 1, stats.FaultCount)
	assert.True(t, stats.OnlineDuration > 0)
	assert.True(t, stats.Availability > 0)
}

func TestDeviceCalculator_CalculateDeviceStatus_OfflineDevice(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "device-001", Code: "D001", Name: "Dev1", Type: "inverter", StationID: "station-001", Status: 0},
		},
	}
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{DataProvider: provider}
	calc := NewDeviceCalculator(config, storage)

	stats, err := calc.CalculateDeviceStatus(context.Background(), "device-001", now.Add(-24*time.Hour), now)
	assert.NoError(t, err)
	assert.True(t, stats.OfflineDuration > 0)
}

func TestDeviceCalculator_CalculateDeviceStatus_FaultDevice(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "device-001", Code: "D001", Name: "Dev1", Type: "inverter", StationID: "station-001", Status: 2},
		},
	}
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{DataProvider: provider}
	calc := NewDeviceCalculator(config, storage)

	stats, err := calc.CalculateDeviceStatus(context.Background(), "device-001", now.Add(-24*time.Hour), now)
	assert.NoError(t, err)
	assert.True(t, stats.FaultDuration > 0)
}

func TestDeviceCalculator_CalculateDeviceStatus_MaintainDevice(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "device-001", Code: "D001", Name: "Dev1", Type: "inverter", StationID: "station-001", Status: 3},
		},
	}
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{DataProvider: provider}
	calc := NewDeviceCalculator(config, storage)

	stats, err := calc.CalculateDeviceStatus(context.Background(), "device-001", now.Add(-24*time.Hour), now)
	assert.NoError(t, err)
	assert.True(t, stats.MaintainDuration > 0)
}

func TestDeviceCalculator_CalculateDeviceStatus_NotFound(t *testing.T) {
	provider := &MockDataProvider{devices: []DeviceInfo{}}
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{DataProvider: provider}
	calc := NewDeviceCalculator(config, storage)

	now := time.Now()
	_, err := calc.CalculateDeviceStatus(context.Background(), "nonexistent", now, now)
	assert.Error(t, err)
}

func TestDeviceCalculator_CalculateDeviceAvailability(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "device-001", Code: "D001", Name: "Dev1", Type: "inverter", StationID: "station-001", Status: 1},
		},
	}
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{DataProvider: provider}
	calc := NewDeviceCalculator(config, storage)

	stats, err := calc.CalculateDeviceAvailability(context.Background(), "device-001", now.Add(-24*time.Hour), now)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "device-001", stats.DeviceID)
	assert.True(t, stats.TotalHours > 0)
	assert.True(t, stats.AvailableHours > 0)
	assert.True(t, stats.Availability > 0)
}

func TestDeviceCalculator_CalculateDeviceAvailability_WithFaults(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "device-001", Code: "D001", Name: "Dev1", Type: "inverter", StationID: "station-001", Status: 2},
		},
		alarms: []AlarmInfo{
			{ID: "a1", StationID: "station-001", DeviceID: "device-001", Type: "device", Level: 3, Status: 1, TriggeredAt: now.Add(-1 * time.Hour)},
		},
	}
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{DataProvider: provider}
	calc := NewDeviceCalculator(config, storage)

	stats, err := calc.CalculateDeviceAvailability(context.Background(), "device-001", now.Add(-24*time.Hour), now)
	assert.NoError(t, err)
	assert.True(t, stats.UnplannedOutage > 0)
	assert.Equal(t, 1, stats.OutageCount)
	assert.True(t, stats.AvgOutageDuration > 0)
}

func TestDeviceCalculator_CalculateAllDeviceTypes(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "d1", Code: "D001", Name: "Inv1", Type: "inverter", StationID: "s1", Status: 1},
			{ID: "d2", Code: "D002", Name: "Met1", Type: "meter", StationID: "s1", Status: 1},
		},
	}
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{
		ParallelWorkers: 2,
		DataProvider:    provider,
	}
	calc := NewDeviceCalculator(config, storage)

	results, err := calc.CalculateAllDeviceTypes(context.Background(), "s1", now.Add(-24*time.Hour), now)
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestDeviceCalculator_CalculateAllDeviceTypes_DefaultWorkers(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "d1", Code: "D001", Name: "Inv1", Type: "inverter", StationID: "s1", Status: 1},
		},
	}
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{
		ParallelWorkers: 0,
		DataProvider:    provider,
	}
	calc := NewDeviceCalculator(config, storage)

	results, err := calc.CalculateAllDeviceTypes(context.Background(), "s1", now.Add(-24*time.Hour), now)
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestDeviceCalculator_CalculateDeviceTypeStatistics_NoStationID(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "d1", Code: "D001", Name: "Inv1", Type: "inverter", StationID: "s1", Status: 1},
		},
	}
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{DataProvider: provider}
	calc := NewDeviceCalculator(config, storage)

	stats, err := calc.CalculateDeviceTypeStatistics(context.Background(), "inverter", "", now.Add(-24*time.Hour), now)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 1, stats.TotalDevices)
}

func TestDeviceCalculator_CalculateDeviceTypeStatistics_WithCache(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "d1", Code: "D001", Name: "Inv1", Type: "inverter", StationID: "s1", Status: 1},
		},
	}
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{
		CacheEnabled: true,
		CacheTTL:     5 * time.Minute,
		DataProvider: provider,
	}
	calc := NewDeviceCalculator(config, storage)

	stats1, err := calc.CalculateDeviceTypeStatistics(context.Background(), "inverter", "s1", now.Add(-24*time.Hour), now)
	assert.NoError(t, err)

	stats2, err := calc.CalculateDeviceTypeStatistics(context.Background(), "inverter", "s1", now.Add(-24*time.Hour), now)
	assert.NoError(t, err)
	assert.Equal(t, stats1.TotalDevices, stats2.TotalDevices)
}

func TestDeviceCalculator_CalculateDeviceTypeStatistics_NoMatchingDevices(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "d1", Code: "D001", Name: "Inv1", Type: "inverter", StationID: "s1", Status: 1},
		},
	}
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{DataProvider: provider}
	calc := NewDeviceCalculator(config, storage)

	stats, err := calc.CalculateDeviceTypeStatistics(context.Background(), "meter", "s1", now.Add(-24*time.Hour), now)
	assert.NoError(t, err)
	assert.Equal(t, 0, stats.TotalDevices)
}

func TestDeviceCalculator_CalculateDeviceTypeStatistics_AllStatuses(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "d1", Code: "D001", Name: "Inv1", Type: "inverter", StationID: "s1", Status: 1},
			{ID: "d2", Code: "D002", Name: "Inv2", Type: "inverter", StationID: "s1", Status: 0},
			{ID: "d3", Code: "D003", Name: "Inv3", Type: "inverter", StationID: "s1", Status: 2},
			{ID: "d4", Code: "D004", Name: "Inv4", Type: "inverter", StationID: "s1", Status: 3},
		},
	}
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{DataProvider: provider}
	calc := NewDeviceCalculator(config, storage)

	stats, err := calc.CalculateDeviceTypeStatistics(context.Background(), "inverter", "s1", now.Add(-24*time.Hour), now)
	assert.NoError(t, err)
	assert.Equal(t, 4, stats.TotalDevices)
	assert.Equal(t, 1, stats.OnlineDevices)
	assert.Equal(t, 1, stats.OfflineDevices)
	assert.Equal(t, 1, stats.FaultDevices)
	assert.Equal(t, 1, stats.MaintainDevices)
}

func TestDeviceCalculator_SaveDeviceTypeStatistics(t *testing.T) {
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewDeviceCalculator(config, storage)

	now := time.Now()
	stats := &DeviceTypeStatistics{
		DeviceType: "inverter", StationID: "s1",
		TotalDevices: 10, OnlineDevices: 8, OfflineDevices: 1, FaultDevices: 1, MaintainDevices: 0,
		OnlineRate: 80.0, OfflineRate: 10.0, FaultRate: 10.0, MaintainRate: 0.0,
		AvgEfficiency: 95.0, Availability: 80.0,
		PeriodStart: now.Add(-24 * time.Hour), PeriodEnd: now,
	}

	err := calc.SaveDeviceTypeStatistics(context.Background(), "task-001", stats)
	assert.NoError(t, err)
	assert.NotEmpty(t, storage.data)
}

func TestDeviceCalculator_SaveDevicePerformanceStatistics(t *testing.T) {
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewDeviceCalculator(config, storage)

	now := time.Now()
	stats := &DevicePerformanceStatistics{
		DeviceID: "d1", DeviceCode: "D001", DeviceName: "Dev1", DeviceType: "inverter", StationID: "s1",
		AvgPower: 80.0, MaxPower: 100.0, MinPower: 20.0, TotalEnergy: 1920.0,
		AvgEfficiency: 95.0, MaxEfficiency: 98.0, MinEfficiency: 90.0,
		RuntimeHours: 24.0, LoadFactor: 80.0, CapacityFactor: 80.0,
		RatedPower: 100.0, DataPoints: 1440, DataQuality: 100.0,
		PeriodStart: now.Add(-24 * time.Hour), PeriodEnd: now,
	}

	err := calc.SaveDevicePerformanceStatistics(context.Background(), "task-001", stats)
	assert.NoError(t, err)
	assert.NotEmpty(t, storage.data)
}

func TestDeviceCalculator_SaveDeviceFaultStatistics(t *testing.T) {
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewDeviceCalculator(config, storage)

	now := time.Now()
	stats := &DeviceFaultStatistics{
		DeviceID: "d1", DeviceCode: "D001", DeviceName: "Dev1", DeviceType: "inverter", StationID: "s1",
		TotalFaultCount: 3, CriticalFaultCount: 1, MajorFaultCount: 1, MinorFaultCount: 1,
		TotalFaultDuration: 10.0, AvgFaultDuration: 3.33, MaxFaultDuration: 5.0,
		MTBF: 100.0, MTTR: 3.33, FailureRate: 1.25,
		FaultTypeDistribution: map[string]int64{"overheat": 2, "overvoltage": 1},
		PeriodStart: now.Add(-24 * time.Hour), PeriodEnd: now,
	}

	err := calc.SaveDeviceFaultStatistics(context.Background(), "task-001", stats)
	assert.NoError(t, err)
	assert.NotEmpty(t, storage.data)
}

func TestDeviceCalculator_SaveDeviceAvailabilityStatistics(t *testing.T) {
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewDeviceCalculator(config, storage)

	now := time.Now()
	stats := &DeviceAvailabilityStatistics{
		DeviceID: "d1", DeviceCode: "D001", DeviceName: "Dev1", DeviceType: "inverter", StationID: "s1",
		TotalHours: 720.0, AvailableHours: 680.0, UnavailableHours: 40.0,
		PlannedOutage: 20.0, UnplannedOutage: 20.0,
		Availability: 94.4, ServiceFactor: 94.4, OperationalRate: 97.1,
		OutageCount: 5, AvgOutageDuration: 4.0,
		PeriodStart: now.Add(-30 * 24 * time.Hour), PeriodEnd: now,
	}

	err := calc.SaveDeviceAvailabilityStatistics(context.Background(), "task-001", stats)
	assert.NoError(t, err)
	assert.NotEmpty(t, storage.data)
}

func TestCustomCalculator_Calculate(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		timeSeriesData: map[string][]TimeSeriesPoint{
			"point-001": {
				{Timestamp: now.Add(-1 * time.Hour), Value: 100.0, Quality: 1},
				{Timestamp: now, Value: 200.0, Quality: 1},
			},
		},
	}
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: provider}
	calc := NewCustomCalculator(config, storage)

	customConfig := &CustomStatisticsConfig{
		ID:          "config-001",
		Name:        "Test Config",
		PeriodType:  PeriodTypeDay,
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
		Metrics: []CustomMetric{
			{Name: "avg_value", Aggregation: AggregationAvg, PointIDs: []string{"point-001"}},
		},
	}
	calc.RegisterConfig(customConfig)

	results, err := calc.Calculate(context.Background(), "config-001")
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, "config-001", results.ConfigID)
}

func TestCustomCalculator_Calculate_NotFound(t *testing.T) {
	provider := &MockDataProvider{}
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: provider}
	calc := NewCustomCalculator(config, storage)

	_, err := calc.Calculate(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestCustomCalculator_CalculateWithConfig(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		timeSeriesData: map[string][]TimeSeriesPoint{
			"point-001": {
				{Timestamp: now.Add(-1 * time.Hour), Value: 100.0, Quality: 1},
				{Timestamp: now, Value: 200.0, Quality: 1},
			},
		},
	}
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: provider}
	calc := NewCustomCalculator(config, storage)

	customConfig := &CustomStatisticsConfig{
		ID:          "config-001",
		Name:        "Test Config",
		PeriodType:  PeriodTypeDay,
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
		Metrics: []CustomMetric{
			{Name: "sum_value", Aggregation: AggregationSum, PointIDs: []string{"point-001"}},
			{Name: "min_value", Aggregation: AggregationMin, PointIDs: []string{"point-001"}},
			{Name: "max_value", Aggregation: AggregationMax, PointIDs: []string{"point-001"}},
			{Name: "count_value", Aggregation: AggregationCount, PointIDs: []string{"point-001"}},
			{Name: "first_value", Aggregation: AggregationFirst, PointIDs: []string{"point-001"}},
			{Name: "last_value", Aggregation: AggregationLast, PointIDs: []string{"point-001"}},
		},
	}

	results, err := calc.CalculateWithConfig(context.Background(), customConfig)
	assert.NoError(t, err)
	assert.NotNil(t, results)
}

func TestCustomCalculator_CalculateWithConfig_WithCache(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		timeSeriesData: map[string][]TimeSeriesPoint{
			"point-001": {{Timestamp: now, Value: 100.0, Quality: 1}},
		},
	}
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{
		CacheEnabled: true,
		CacheTTL:     5 * time.Minute,
		DataProvider: provider,
	}
	calc := NewCustomCalculator(config, storage)

	customConfig := &CustomStatisticsConfig{
		ID: "config-001", Name: "Test", PeriodType: PeriodTypeDay,
		PeriodStart: now.Add(-24 * time.Hour), PeriodEnd: now,
		Metrics: []CustomMetric{{Name: "val", Aggregation: AggregationAvg, PointIDs: []string{"point-001"}}},
	}

	results1, err := calc.CalculateWithConfig(context.Background(), customConfig)
	assert.NoError(t, err)

	results2, err := calc.CalculateWithConfig(context.Background(), customConfig)
	assert.NoError(t, err)
	assert.Equal(t, results1.ConfigID, results2.ConfigID)
}

func TestCustomCalculator_CalculateWithConfig_NoPointIDs(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{}
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: provider}
	calc := NewCustomCalculator(config, storage)

	customConfig := &CustomStatisticsConfig{
		ID: "config-001", Name: "Test", PeriodType: PeriodTypeDay,
		PeriodStart: now, PeriodEnd: now,
		Metrics: []CustomMetric{{Name: "val", Aggregation: AggregationAvg}},
	}

	results, err := calc.CalculateWithConfig(context.Background(), customConfig)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Empty(t, results.Results)
}

func TestCustomCalculator_CalculateWithConfig_WithGrouping(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		timeSeriesData: map[string][]TimeSeriesPoint{
			"point-001": {
				{Timestamp: now.Add(-2 * time.Hour), Value: 100.0, Quality: 1},
				{Timestamp: now.Add(-1 * time.Hour), Value: 200.0, Quality: 1},
			},
		},
	}
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: provider}
	calc := NewCustomCalculator(config, storage)

	customConfig := &CustomStatisticsConfig{
		ID: "config-001", Name: "Test", PeriodType: PeriodTypeDay,
		PeriodStart: now.Add(-24 * time.Hour), PeriodEnd: now,
		Metrics: []CustomMetric{{Name: "avg_val", Aggregation: AggregationAvg, PointIDs: []string{"point-001"}}},
		GroupBy: []GroupByField{{Field: "point_id", Alias: "point"}},
	}

	results, err := calc.CalculateWithConfig(context.Background(), customConfig)
	assert.NoError(t, err)
	assert.NotEmpty(t, results.Results)
}

func TestCustomCalculator_CalculateWithConfig_WithSorting(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		timeSeriesData: map[string][]TimeSeriesPoint{
			"point-001": {{Timestamp: now, Value: 100.0, Quality: 1}},
			"point-002": {{Timestamp: now, Value: 200.0, Quality: 1}},
		},
	}
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: provider}
	calc := NewCustomCalculator(config, storage)

	customConfig := &CustomStatisticsConfig{
		ID: "config-001", Name: "Test", PeriodType: PeriodTypeDay,
		PeriodStart: now.Add(-24 * time.Hour), PeriodEnd: now,
		Metrics: []CustomMetric{{Name: "avg_val", Aggregation: AggregationAvg, PointIDs: []string{"point-001", "point-002"}}},
		GroupBy: []GroupByField{{Field: "point_id", Alias: "point"}},
		OrderBy: []OrderByField{{Field: "avg_val", Desc: true}},
	}

	results, err := calc.CalculateWithConfig(context.Background(), customConfig)
	assert.NoError(t, err)
	assert.NotEmpty(t, results.Results)
}

func TestCustomCalculator_CalculateWithConfig_WithPagination(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		timeSeriesData: map[string][]TimeSeriesPoint{
			"point-001": {{Timestamp: now, Value: 100.0, Quality: 1}},
			"point-002": {{Timestamp: now, Value: 200.0, Quality: 1}},
			"point-003": {{Timestamp: now, Value: 300.0, Quality: 1}},
		},
	}
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: provider}
	calc := NewCustomCalculator(config, storage)

	customConfig := &CustomStatisticsConfig{
		ID: "config-001", Name: "Test", PeriodType: PeriodTypeDay,
		PeriodStart: now.Add(-24 * time.Hour), PeriodEnd: now,
		Metrics: []CustomMetric{{Name: "avg_val", Aggregation: AggregationAvg, PointIDs: []string{"point-001", "point-002", "point-003"}}},
		GroupBy: []GroupByField{{Field: "point_id", Alias: "point"}},
		Limit:  2,
		Offset: 1,
	}

	results, err := calc.CalculateWithConfig(context.Background(), customConfig)
	assert.NoError(t, err)
	assert.LessOrEqual(t, len(results.Results), 2)
	assert.Equal(t, 3, results.Total)
}

func TestCustomCalculator_CalculateWithConfig_OffsetBeyondResults(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		timeSeriesData: map[string][]TimeSeriesPoint{
			"point-001": {{Timestamp: now, Value: 100.0, Quality: 1}},
		},
	}
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: provider}
	calc := NewCustomCalculator(config, storage)

	customConfig := &CustomStatisticsConfig{
		ID: "config-001", Name: "Test", PeriodType: PeriodTypeDay,
		PeriodStart: now.Add(-24 * time.Hour), PeriodEnd: now,
		Metrics: []CustomMetric{{Name: "avg_val", Aggregation: AggregationAvg, PointIDs: []string{"point-001"}}},
		GroupBy: []GroupByField{{Field: "point_id", Alias: "point"}},
		Offset: 100,
	}

	results, err := calc.CalculateWithConfig(context.Background(), customConfig)
	assert.NoError(t, err)
	assert.Empty(t, results.Results)
}

func TestCustomCalculator_ApplyFilters(t *testing.T) {
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewCustomCalculator(config, storage)

	data := []map[string]interface{}{
		{"point_id": "p1", "value": 100.0},
		{"point_id": "p2", "value": 200.0},
		{"point_id": "p3", "value": 50.0},
	}

	t.Run("eq filter", func(t *testing.T) {
		filtered := calc.applyFilters(data, FilterGroup{
			Conditions: []FilterCondition{{Field: "point_id", Operator: "eq", Value: "p1"}},
		})
		assert.Len(t, filtered, 1)
	})

	t.Run("ne filter", func(t *testing.T) {
		filtered := calc.applyFilters(data, FilterGroup{
			Conditions: []FilterCondition{{Field: "point_id", Operator: "ne", Value: "p1"}},
		})
		assert.Len(t, filtered, 2)
	})

	t.Run("gt filter", func(t *testing.T) {
		filtered := calc.applyFilters(data, FilterGroup{
			Conditions: []FilterCondition{{Field: "value", Operator: "gt", Value: 80.0}},
		})
		assert.Len(t, filtered, 2)
	})

	t.Run("lt filter", func(t *testing.T) {
		filtered := calc.applyFilters(data, FilterGroup{
			Conditions: []FilterCondition{{Field: "value", Operator: "lt", Value: 150.0}},
		})
		assert.Len(t, filtered, 2)
	})

	t.Run("gte filter", func(t *testing.T) {
		filtered := calc.applyFilters(data, FilterGroup{
			Conditions: []FilterCondition{{Field: "value", Operator: "gte", Value: 100.0}},
		})
		assert.Len(t, filtered, 2)
	})

	t.Run("lte filter", func(t *testing.T) {
		filtered := calc.applyFilters(data, FilterGroup{
			Conditions: []FilterCondition{{Field: "value", Operator: "lte", Value: 100.0}},
		})
		assert.Len(t, filtered, 2)
	})

	t.Run("in filter", func(t *testing.T) {
		filtered := calc.applyFilters(data, FilterGroup{
			Conditions: []FilterCondition{{Field: "point_id", Operator: "in", Value: []interface{}{"p1", "p2"}}},
		})
		assert.Len(t, filtered, 2)
	})

	t.Run("not_in filter", func(t *testing.T) {
		filtered := calc.applyFilters(data, FilterGroup{
			Conditions: []FilterCondition{{Field: "point_id", Operator: "not_in", Value: []interface{}{"p1"}}},
		})
		assert.Len(t, filtered, 2)
	})

	t.Run("like filter", func(t *testing.T) {
		filtered := calc.applyFilters(data, FilterGroup{
			Conditions: []FilterCondition{{Field: "point_id", Operator: "like", Value: "p"}},
		})
		assert.Len(t, filtered, 3)
	})

	t.Run("or logic", func(t *testing.T) {
		filtered := calc.applyFilters(data, FilterGroup{
			Logic: "or",
			Conditions: []FilterCondition{
				{Field: "value", Operator: "gt", Value: 150.0},
				{Field: "value", Operator: "lt", Value: 80.0},
			},
		})
		assert.Len(t, filtered, 2)
	})

	t.Run("and logic", func(t *testing.T) {
		filtered := calc.applyFilters(data, FilterGroup{
			Logic: "and",
			Conditions: []FilterCondition{
				{Field: "value", Operator: "gte", Value: 50.0},
				{Field: "value", Operator: "lte", Value: 100.0},
			},
		})
		assert.Len(t, filtered, 2)
	})

	t.Run("nested filter groups", func(t *testing.T) {
		filtered := calc.applyFilters(data, FilterGroup{
			Logic: "and",
			Conditions: []FilterCondition{{Field: "value", Operator: "gt", Value: 0.0}},
			Groups: []FilterGroup{
				{Logic: "or", Conditions: []FilterCondition{
					{Field: "point_id", Operator: "eq", Value: "p1"},
					{Field: "point_id", Operator: "eq", Value: "p3"},
				}},
			},
		})
		assert.Len(t, filtered, 2)
	})

	t.Run("empty filter", func(t *testing.T) {
		filtered := calc.applyFilters(data, FilterGroup{})
		assert.Len(t, filtered, 3)
	})

	t.Run("missing field", func(t *testing.T) {
		filtered := calc.applyFilters(data, FilterGroup{
			Conditions: []FilterCondition{{Field: "nonexistent", Operator: "eq", Value: "x"}},
		})
		assert.Len(t, filtered, 0)
	})
}

func TestCustomCalculator_EvaluateCondition_UnknownOperator(t *testing.T) {
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewCustomCalculator(config, storage)

	record := map[string]interface{}{"value": 100.0}
	result := calc.evaluateCondition(record, FilterCondition{Field: "value", Operator: "unknown", Value: 100.0})
	assert.False(t, result)
}

func TestCustomCalculator_EvaluateCondition_InNotSlice(t *testing.T) {
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewCustomCalculator(config, storage)

	record := map[string]interface{}{"value": 100.0}
	result := calc.evaluateCondition(record, FilterCondition{Field: "value", Operator: "in", Value: "not_a_slice"})
	assert.False(t, result)
}

func TestCustomCalculator_EvaluateCondition_NotInNotSlice(t *testing.T) {
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewCustomCalculator(config, storage)

	record := map[string]interface{}{"value": 100.0}
	result := calc.evaluateCondition(record, FilterCondition{Field: "value", Operator: "not_in", Value: "not_a_slice"})
	assert.True(t, result)
}

func TestCustomCalculator_CalculateMetric_AllAggregations(t *testing.T) {
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewCustomCalculator(config, storage)

	data := []map[string]interface{}{
		{"value": 10.0},
		{"value": 20.0},
		{"value": 30.0},
		{"value": 40.0},
		{"value": 50.0},
	}

	t.Run("sum", func(t *testing.T) {
		result := calc.calculateMetric(data, CustomMetric{Aggregation: AggregationSum, ScaleFactor: 1.0})
		assert.Equal(t, 150.0, result)
	})

	t.Run("avg", func(t *testing.T) {
		result := calc.calculateMetric(data, CustomMetric{Aggregation: AggregationAvg, ScaleFactor: 1.0})
		assert.Equal(t, 30.0, result)
	})

	t.Run("min", func(t *testing.T) {
		result := calc.calculateMetric(data, CustomMetric{Aggregation: AggregationMin, ScaleFactor: 1.0})
		assert.Equal(t, 10.0, result)
	})

	t.Run("max", func(t *testing.T) {
		result := calc.calculateMetric(data, CustomMetric{Aggregation: AggregationMax, ScaleFactor: 1.0})
		assert.Equal(t, 50.0, result)
	})

	t.Run("count", func(t *testing.T) {
		result := calc.calculateMetric(data, CustomMetric{Aggregation: AggregationCount, ScaleFactor: 1.0})
		assert.Equal(t, 5.0, result)
	})

	t.Run("first", func(t *testing.T) {
		result := calc.calculateMetric(data, CustomMetric{Aggregation: AggregationFirst, ScaleFactor: 1.0})
		assert.Equal(t, 10.0, result)
	})

	t.Run("last", func(t *testing.T) {
		result := calc.calculateMetric(data, CustomMetric{Aggregation: AggregationLast, ScaleFactor: 1.0})
		assert.Equal(t, 50.0, result)
	})

	t.Run("stddev", func(t *testing.T) {
		result := calc.calculateMetric(data, CustomMetric{Aggregation: AggregationStdDev, ScaleFactor: 1.0})
		assert.True(t, result > 0)
	})

	t.Run("variance", func(t *testing.T) {
		result := calc.calculateMetric(data, CustomMetric{Aggregation: AggregationVariance, ScaleFactor: 1.0})
		assert.True(t, result > 0)
	})

	t.Run("median", func(t *testing.T) {
		result := calc.calculateMetric(data, CustomMetric{Aggregation: AggregationMedian, ScaleFactor: 1.0})
		assert.Equal(t, 30.0, result)
	})

	t.Run("p95", func(t *testing.T) {
		result := calc.calculateMetric(data, CustomMetric{Aggregation: AggregationP95, ScaleFactor: 1.0})
		assert.True(t, result >= 40.0)
	})

	t.Run("p99", func(t *testing.T) {
		result := calc.calculateMetric(data, CustomMetric{Aggregation: AggregationP99, ScaleFactor: 1.0})
		assert.True(t, result >= 40.0)
	})

	t.Run("default", func(t *testing.T) {
		result := calc.calculateMetric(data, CustomMetric{Aggregation: "unknown", ScaleFactor: 1.0})
		assert.Equal(t, 30.0, result)
	})
}

func TestCustomCalculator_CalculateMetric_EmptyData(t *testing.T) {
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewCustomCalculator(config, storage)

	result := calc.calculateMetric([]map[string]interface{}{}, CustomMetric{Aggregation: AggregationAvg, ScaleFactor: 1.0})
	assert.Equal(t, 0.0, result)
}

func TestCustomCalculator_CalculateMetric_NoFloatValues(t *testing.T) {
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewCustomCalculator(config, storage)

	data := []map[string]interface{}{
		{"value": "not_a_number"},
		{"value": "also_not"},
	}

	result := calc.calculateMetric(data, CustomMetric{Aggregation: AggregationAvg, ScaleFactor: 1.0})
	assert.Equal(t, 0.0, result)
}

func TestCustomCalculator_CalculateMetric_WithScaleAndOffset(t *testing.T) {
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewCustomCalculator(config, storage)

	data := []map[string]interface{}{
		{"value": 100.0},
	}

	result := calc.calculateMetric(data, CustomMetric{
		Aggregation: AggregationAvg,
		ScaleFactor: 2.0,
		Offset:      10.0,
	})
	assert.Equal(t, 210.0, result)
}

func TestCustomCalculator_GenerateGroupKey_WithTimeFormat(t *testing.T) {
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewCustomCalculator(config, storage)

	now := time.Now()
	record := map[string]interface{}{
		"timestamp": now,
		"station":   "s1",
	}

	key := calc.generateGroupKey(record, []GroupByField{
		{Field: "timestamp", Alias: "hour", TimeFormat: "2006-01-02"},
		{Field: "station", Alias: "station"},
	})
	assert.Contains(t, key, now.Format("2006-01-02"))
	assert.Contains(t, key, "s1")
}

func TestCustomCalculator_ParseGroupKey(t *testing.T) {
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewCustomCalculator(config, storage)

	dimensions := calc.parseGroupKey("s1|inverter", []GroupByField{
		{Field: "station", Alias: "station"},
		{Field: "device_type"},
	})
	assert.Equal(t, "s1", dimensions["station"])
	assert.Equal(t, "inverter", dimensions["device_type"])
}

func TestCustomCalculator_ParseGroupKey_MoreFieldsThanValues(t *testing.T) {
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewCustomCalculator(config, storage)

	dimensions := calc.parseGroupKey("s1", []GroupByField{
		{Field: "station", Alias: "station"},
		{Field: "device_type", Alias: "type"},
	})
	assert.Equal(t, "s1", dimensions["station"])
	_, ok := dimensions["type"]
	assert.False(t, ok)
}

func TestCustomCalculator_ApplySorting_Empty(t *testing.T) {
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewCustomCalculator(config, storage)

	results := []*CustomStatisticsResult{{Metrics: map[string]float64{"val": 1.0}}}
	sorted := calc.applySorting(results, nil)
	assert.Len(t, sorted, 1)
}

func TestCustomCalculator_CompareResults(t *testing.T) {
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewCustomCalculator(config, storage)

	a := &CustomStatisticsResult{
		Metrics:    map[string]float64{"val": 10.0},
		Dimensions: map[string]string{"name": "a"},
	}
	b := &CustomStatisticsResult{
		Metrics:    map[string]float64{"val": 20.0},
		Dimensions: map[string]string{"name": "b"},
	}

	t.Run("ascending", func(t *testing.T) {
		result := calc.compareResults(a, b, []OrderByField{{Field: "val", Desc: false}})
		assert.Equal(t, -1, result)
	})

	t.Run("descending", func(t *testing.T) {
		result := calc.compareResults(a, b, []OrderByField{{Field: "val", Desc: true}})
		assert.Equal(t, 1, result)
	})

	t.Run("equal values", func(t *testing.T) {
		a2 := &CustomStatisticsResult{Metrics: map[string]float64{"val": 10.0}}
		b2 := &CustomStatisticsResult{Metrics: map[string]float64{"val": 10.0}}
		result := calc.compareResults(a2, b2, []OrderByField{{Field: "val"}})
		assert.Equal(t, 0, result)
	})

	t.Run("dimension value", func(t *testing.T) {
		result := calc.compareResults(a, b, []OrderByField{{Field: "name"}})
		assert.Equal(t, 0, result)
	})
}

func TestCustomCalculator_ApplyPagination(t *testing.T) {
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewCustomCalculator(config, storage)

	results := make([]*CustomStatisticsResult, 5)
	for i := range results {
		results[i] = &CustomStatisticsResult{Metrics: map[string]float64{"val": float64(i)}}
	}

	t.Run("limit only", func(t *testing.T) {
		paged := calc.applyPagination(results, 3, 0)
		assert.Len(t, paged, 3)
	})

	t.Run("offset only", func(t *testing.T) {
		paged := calc.applyPagination(results, 0, 2)
		assert.Len(t, paged, 3)
	})

	t.Run("limit and offset", func(t *testing.T) {
		paged := calc.applyPagination(results, 2, 1)
		assert.Len(t, paged, 2)
	})

	t.Run("offset beyond", func(t *testing.T) {
		paged := calc.applyPagination(results, 0, 100)
		assert.Empty(t, paged)
	})
}

func TestCustomCalculator_CalculateSummary(t *testing.T) {
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewCustomCalculator(config, storage)

	results := []*CustomStatisticsResult{
		{Metrics: map[string]float64{"val": 10.0}},
		{Metrics: map[string]float64{"val": 20.0}},
	}

	summary := calc.calculateSummary(results)
	assert.NotNil(t, summary)
	assert.Contains(t, summary, "val")
	assert.Equal(t, 30.0, summary["val"].Sum)
}

func TestCustomCalculator_CalculateSummary_Empty(t *testing.T) {
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewCustomCalculator(config, storage)

	summary := calc.calculateSummary([]*CustomStatisticsResult{})
	assert.Empty(t, summary)
}

func TestCustomCalculator_CalculateMultiDimension(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		timeSeriesData: map[string][]TimeSeriesPoint{
			"point-001": {{Timestamp: now, Value: 100.0, Quality: 1}},
		},
	}
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: provider}
	calc := NewCustomCalculator(config, storage)

	customConfig := &CustomStatisticsConfig{
		ID: "config-001", Name: "Test", PeriodType: PeriodTypeDay,
		PeriodStart: now.Add(-24 * time.Hour), PeriodEnd: now,
		Metrics: []CustomMetric{{Name: "val", Aggregation: AggregationAvg, PointIDs: []string{"point-001"}}},
	}

	results, err := calc.CalculateMultiDimension(context.Background(), customConfig)
	assert.NoError(t, err)
	assert.NotNil(t, results)
}

func TestCustomCalculator_CalculateTimeSeries(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		timeSeriesData: map[string][]TimeSeriesPoint{
			"point-001": {{Timestamp: now, Value: 100.0, Quality: 1}},
		},
	}
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: provider}
	calc := NewCustomCalculator(config, storage)

	customConfig := &CustomStatisticsConfig{
		ID: "config-001", Name: "Test", PeriodType: PeriodTypeDay,
		PeriodStart: now.Add(-3 * time.Hour),
		PeriodEnd:   now,
		Metrics: []CustomMetric{{Name: "val", Aggregation: AggregationAvg, PointIDs: []string{"point-001"}}},
	}

	results, err := calc.CalculateTimeSeries(context.Background(), customConfig, time.Hour)
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestCustomCalculator_CalculateTimeSeries_Error(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{}
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: provider}
	calc := NewCustomCalculator(config, storage)

	customConfig := &CustomStatisticsConfig{
		ID: "config-001", Name: "Test", PeriodType: PeriodTypeDay,
		PeriodStart: now.Add(-3 * time.Hour),
		PeriodEnd:   now,
		Metrics:     []CustomMetric{{Name: "val", Aggregation: AggregationAvg}},
	}

	results, err := calc.CalculateTimeSeries(context.Background(), customConfig, time.Hour)
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestCustomCalculator_SaveResults(t *testing.T) {
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewCustomCalculator(config, storage)

	now := time.Now()
	results := &CustomStatisticsResults{
		ConfigID:   "config-001",
		ConfigName: "Test Config",
		Results: []*CustomStatisticsResult{
			{
				ConfigID:    "config-001",
				ConfigName:  "Test Config",
				PeriodStart: now.Add(-24 * time.Hour),
				PeriodEnd:   now,
				Dimensions:  map[string]string{"station": "s1"},
				Metrics:     map[string]float64{"generation": 1000.0},
				Metadata:    map[string]interface{}{"quality": 100.0},
			},
		},
	}

	err := calc.SaveResults(context.Background(), "task-001", results)
	assert.NoError(t, err)
	assert.NotEmpty(t, storage.data)
}

func TestCustomCalculator_SaveResults_EmptyDimensions(t *testing.T) {
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewCustomCalculator(config, storage)

	now := time.Now()
	results := &CustomStatisticsResults{
		ConfigID:   "config-001",
		ConfigName: "Test Config",
		Results: []*CustomStatisticsResult{
			{
				ConfigID:    "config-001",
				PeriodStart: now.Add(-24 * time.Hour),
				PeriodEnd:   now,
				Dimensions:  map[string]string{},
				Metrics:     map[string]float64{"generation": 1000.0},
			},
		},
	}

	err := calc.SaveResults(context.Background(), "task-001", results)
	assert.NoError(t, err)
}

func TestCustomCalculator_SaveResults_NoMetadata(t *testing.T) {
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: &MockDataProvider{}}
	calc := NewCustomCalculator(config, storage)

	now := time.Now()
	results := &CustomStatisticsResults{
		ConfigID:   "config-001",
		ConfigName: "Test Config",
		Results: []*CustomStatisticsResult{
			{
				ConfigID:    "config-001",
				PeriodStart: now.Add(-24 * time.Hour),
				PeriodEnd:   now,
				Dimensions:  map[string]string{"station": "s1"},
				Metrics:     map[string]float64{"generation": 1000.0},
			},
		},
	}

	err := calc.SaveResults(context.Background(), "task-001", results)
	assert.NoError(t, err)
}

func TestCustomCalculator_CreatePresetConfig_AllTypes(t *testing.T) {
	provider := &MockDataProvider{}
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: provider}
	calc := NewCustomCalculator(config, storage)

	t.Run("generation_by_hour", func(t *testing.T) {
		cfg, err := calc.CreatePresetConfig("generation_by_hour", map[string]interface{}{"station_id": "s1"})
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, PeriodTypeHour, cfg.PeriodType)
	})

	t.Run("generation_by_hour_no_station", func(t *testing.T) {
		cfg, err := calc.CreatePresetConfig("generation_by_hour", map[string]interface{}{})
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
	})

	t.Run("device_status_summary", func(t *testing.T) {
		cfg, err := calc.CreatePresetConfig("device_status_summary", nil)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, PeriodTypeDay, cfg.PeriodType)
	})

	t.Run("alarm_statistics", func(t *testing.T) {
		cfg, err := calc.CreatePresetConfig("alarm_statistics", nil)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
	})

	t.Run("efficiency_analysis", func(t *testing.T) {
		cfg, err := calc.CreatePresetConfig("efficiency_analysis", nil)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
	})

	t.Run("unknown_preset", func(t *testing.T) {
		_, err := calc.CreatePresetConfig("unknown", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown preset type")
	})
}

func TestCustomCalculator_GetConfig_NotFound(t *testing.T) {
	provider := &MockDataProvider{}
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: provider}
	calc := NewCustomCalculator(config, storage)

	_, err := calc.GetConfig("nonexistent")
	assert.Error(t, err)
}

func TestCustomCalculator_RegisterConfig_AutoID(t *testing.T) {
	provider := &MockDataProvider{}
	storage := &MockStatisticsStorage{}
	config := CustomCalculatorConfig{DataProvider: provider}
	calc := NewCustomCalculator(config, storage)

	cfg := &CustomStatisticsConfig{Name: "Test", PeriodType: PeriodTypeDay}
	err := calc.RegisterConfig(cfg)
	assert.NoError(t, err)
	assert.NotEmpty(t, cfg.ID)
}

func TestHelperFunctions_EdgeCases(t *testing.T) {
	t.Run("avgValues empty", func(t *testing.T) {
		assert.Equal(t, 0.0, avgValues([]float64{}))
	})

	t.Run("minValues empty", func(t *testing.T) {
		assert.Equal(t, 0.0, minValues([]float64{}))
	})

	t.Run("maxValues empty", func(t *testing.T) {
		assert.Equal(t, 0.0, maxValues([]float64{}))
	})

	t.Run("medianValues empty", func(t *testing.T) {
		assert.Equal(t, 0.0, medianValues([]float64{}))
	})

	t.Run("percentileValues empty", func(t *testing.T) {
		assert.Equal(t, 0.0, percentileValues([]float64{}, 95))
	})

	t.Run("toFloat64 float32", func(t *testing.T) {
		val, ok := toFloat64(float32(10.5))
		assert.True(t, ok)
		assert.InDelta(t, 10.5, val, 0.01)
	})

	t.Run("toFloat64 int32", func(t *testing.T) {
		val, ok := toFloat64(int32(10))
		assert.True(t, ok)
		assert.Equal(t, 10.0, val)
	})

	t.Run("splitString no separator", func(t *testing.T) {
		result := splitString("abc", "|")
		assert.Len(t, result, 1)
		assert.Equal(t, "abc", result[0])
	})
}

func TestDeviceCalculator_CalculateDevicePerformance_NoPoints(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "device-001", Code: "D001", Name: "Dev1", Type: "inverter", StationID: "station-001", RatedPower: 100.0, Status: 1},
		},
		points:         []PointInfo{},
		timeSeriesData: map[string][]TimeSeriesPoint{},
	}
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{DataProvider: provider}
	calc := NewDeviceCalculator(config, storage)

	stats, err := calc.CalculateDevicePerformance(context.Background(), "device-001", now.Add(-3*time.Hour), now)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 0.0, stats.AvgPower)
}

func TestDeviceCalculator_CalculateDevicePerformance_WithEfficiency(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "device-001", Code: "D001", Name: "Dev1", Type: "inverter", StationID: "station-001", RatedPower: 100.0, Status: 1},
		},
		points: []PointInfo{
			{ID: "power-001", Code: "P001", Name: "Power", Type: "power", DeviceID: "device-001", Unit: "kW"},
			{ID: "eff-001", Code: "E001", Name: "Efficiency", Type: "efficiency", DeviceID: "device-001", Unit: "%"},
		},
		timeSeriesData: map[string][]TimeSeriesPoint{
			"power-001": {
				{Timestamp: now.Add(-2 * time.Hour), Value: 50.0, Quality: 1},
				{Timestamp: now.Add(-1 * time.Hour), Value: 80.0, Quality: 1},
			},
			"eff-001": {
				{Timestamp: now.Add(-2 * time.Hour), Value: 95.0, Quality: 1},
				{Timestamp: now.Add(-1 * time.Hour), Value: 98.0, Quality: 1},
			},
		},
	}
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{DataProvider: provider}
	calc := NewDeviceCalculator(config, storage)

	stats, err := calc.CalculateDevicePerformance(context.Background(), "device-001", now.Add(-3*time.Hour), now)
	assert.NoError(t, err)
	assert.True(t, stats.AvgEfficiency > 0)
	assert.True(t, stats.MaxEfficiency > 0)
	assert.True(t, stats.MinEfficiency > 0)
}

func TestDeviceCalculator_CalculateDevicePerformance_NotFound(t *testing.T) {
	provider := &MockDataProvider{devices: []DeviceInfo{}}
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{DataProvider: provider}
	calc := NewDeviceCalculator(config, storage)

	now := time.Now()
	_, err := calc.CalculateDevicePerformance(context.Background(), "nonexistent", now, now)
	assert.Error(t, err)
}

func TestDeviceCalculator_CalculateDeviceFaultStats_NoFaults(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "device-001", Code: "D001", Name: "Dev1", Type: "inverter", StationID: "station-001", Status: 1},
		},
		alarms: []AlarmInfo{},
	}
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{DataProvider: provider}
	calc := NewDeviceCalculator(config, storage)

	stats, err := calc.CalculateDeviceFaultStats(context.Background(), "device-001", now.Add(-24*time.Hour), now)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), stats.TotalFaultCount)
	assert.True(t, stats.MTBF > 0)
}

func TestDeviceCalculator_CalculateDeviceFaultStats_NotFound(t *testing.T) {
	provider := &MockDataProvider{devices: []DeviceInfo{}}
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{DataProvider: provider}
	calc := NewDeviceCalculator(config, storage)

	now := time.Now()
	_, err := calc.CalculateDeviceFaultStats(context.Background(), "nonexistent", now, now)
	assert.Error(t, err)
}

func TestDeviceCalculator_CalculateDeviceTypeStatistics_WithPerformance(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "d1", Code: "D001", Name: "Inv1", Type: "inverter", StationID: "s1", RatedPower: 100.0, Status: 1},
		},
		points: []PointInfo{
			{ID: "power-001", Code: "P001", Name: "Power", Type: "power", DeviceID: "d1", Unit: "kW"},
		},
		timeSeriesData: map[string][]TimeSeriesPoint{
			"power-001": {
				{Timestamp: now.Add(-1 * time.Hour), Value: 80.0, Quality: 1},
				{Timestamp: now, Value: 90.0, Quality: 1},
			},
		},
		alarms: []AlarmInfo{},
	}
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{DataProvider: provider}
	calc := NewDeviceCalculator(config, storage)

	stats, err := calc.CalculateDeviceTypeStatistics(context.Background(), "inverter", "s1", now.Add(-24*time.Hour), now)
	assert.NoError(t, err)
	assert.True(t, stats.AvgPowerOutput > 0)
	assert.True(t, stats.TotalPowerOutput > 0)
}

func TestDeviceCalculator_CalculateDeviceTypeStatistics_WithFaults(t *testing.T) {
	now := time.Now()
	clearedAt := now.Add(-30 * time.Minute)
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "d1", Code: "D001", Name: "Inv1", Type: "inverter", StationID: "s1", Status: 1},
		},
		points: []PointInfo{},
		alarms: []AlarmInfo{
			{ID: "a1", StationID: "s1", DeviceID: "d1", Type: "fault", Level: 4, Status: 3, TriggeredAt: now.Add(-2 * time.Hour), ClearedAt: &clearedAt},
		},
	}
	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{DataProvider: provider}
	calc := NewDeviceCalculator(config, storage)

	stats, err := calc.CalculateDeviceTypeStatistics(context.Background(), "inverter", "s1", now.Add(-24*time.Hour), now)
	assert.NoError(t, err)
	assert.True(t, stats.FaultCount > 0)
}

func TestNewPostgreSQLStorage_InvalidConnection(t *testing.T) {
	config := StorageConfig{
		Host:     "invalid-host",
		Port:     5432,
		User:     "invalid",
		Password: "invalid",
		DBName:   "invalid",
		SSLMode:  "disable",
	}

	_, err := NewPostgreSQLStorage(config)
	assert.Error(t, err)
}

func TestNewTimeSeriesDBStorage(t *testing.T) {
	pgStorage := setupPostgreSQLStorage(t)
	defer pgStorage.Close()

	tsStorage, err := NewTimeSeriesDBStorage(pgStorage, StorageConfig{TSDBEnabled: true})
	require.NoError(t, err)
	assert.NotNil(t, tsStorage)
	assert.Equal(t, pgStorage, tsStorage.pgStorage)
}

func TestGenerateUUID(t *testing.T) {
	id1 := generateUUID()
	id2 := generateUUID()
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
}

func TestStatisticsData_TableName(t *testing.T) {
	data := &StatisticsData{}
	assert.Equal(t, "statistics_data", data.TableName())
}

func TestStatisticsTask_TableName(t *testing.T) {
	task := &StatisticsTask{}
	assert.Equal(t, "statistics_tasks", task.TableName())
}

func TestPeriodTypes(t *testing.T) {
	assert.Equal(t, PeriodType("minute"), PeriodTypeMinute)
	assert.Equal(t, PeriodType("hour"), PeriodTypeHour)
	assert.Equal(t, PeriodType("day"), PeriodTypeDay)
	assert.Equal(t, PeriodType("month"), PeriodTypeMonth)
	assert.Equal(t, PeriodType("year"), PeriodTypeYear)
	assert.Equal(t, PeriodType("custom"), PeriodTypeCustom)
}

func TestAggregationTypes(t *testing.T) {
	assert.Equal(t, AggregationType("sum"), AggregationSum)
	assert.Equal(t, AggregationType("avg"), AggregationAvg)
	assert.Equal(t, AggregationType("min"), AggregationMin)
	assert.Equal(t, AggregationType("max"), AggregationMax)
	assert.Equal(t, AggregationType("count"), AggregationCount)
	assert.Equal(t, AggregationType("first"), AggregationFirst)
	assert.Equal(t, AggregationType("last"), AggregationLast)
	assert.Equal(t, AggregationType("stddev"), AggregationStdDev)
	assert.Equal(t, AggregationType("variance"), AggregationVariance)
	assert.Equal(t, AggregationType("median"), AggregationMedian)
	assert.Equal(t, AggregationType("p95"), AggregationP95)
	assert.Equal(t, AggregationType("p99"), AggregationP99)
}

func TestDimensionTypes(t *testing.T) {
	assert.Equal(t, DimensionType("station"), DimensionTypeStation)
	assert.Equal(t, DimensionType("device_type"), DimensionTypeDeviceType)
	assert.Equal(t, DimensionType("region"), DimensionTypeRegion)
	assert.Equal(t, DimensionType("time"), DimensionTypeTime)
	assert.Equal(t, DimensionType("custom"), DimensionTypeCustom)
}
