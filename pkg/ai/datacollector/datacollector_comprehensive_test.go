package datacollector

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataPoint_Struct(t *testing.T) {
	dp := &DataPoint{
		Timestamp: time.Now(),
		StationID: "station-1",
		DeviceID:  "dev-1",
		Metric:    "power",
		Value:     100.0,
		Unit:      "kW",
		Quality:   QualityGood,
		Metadata:  map[string]interface{}{"source": "test"},
	}
	assert.Equal(t, "station-1", dp.StationID)
	assert.Equal(t, 100.0, dp.Value)
	assert.Equal(t, QualityGood, dp.Quality)
}

func TestDataQuality_Constants(t *testing.T) {
	assert.Equal(t, DataQuality(0), QualityUnknown)
	assert.Equal(t, DataQuality(1), QualityGood)
	assert.Equal(t, DataQuality(2), QualitySuspect)
	assert.Equal(t, DataQuality(3), QualityBad)
	assert.Equal(t, DataQuality(4), QualityMissing)
}

func TestDataFormat_Constants(t *testing.T) {
	assert.Equal(t, DataFormat("csv"), FormatCSV)
	assert.Equal(t, DataFormat("json"), FormatJSON)
	assert.Equal(t, DataFormat("parquet"), FormatParquet)
	assert.Equal(t, DataFormat("excel"), FormatExcel)
}

func TestJobStatus_Constants(t *testing.T) {
	assert.Equal(t, JobStatus("pending"), JobStatusPending)
	assert.Equal(t, JobStatus("running"), JobStatusRunning)
	assert.Equal(t, JobStatus("completed"), JobStatusCompleted)
	assert.Equal(t, JobStatus("failed"), JobStatusFailed)
	assert.Equal(t, JobStatus("paused"), JobStatusPaused)
}

func TestBatchImportRequest_Struct(t *testing.T) {
	req := &BatchImportRequest{
		FilePath: "/data/test.csv",
		Format:   FormatCSV,
		Mapping:  map[string]string{"timestamp": "time"},
		Validation: ValidationConfig{
			CheckMissing:  true,
			CheckOutliers: true,
		},
		Options: ImportOptions{
			BatchSize:  1000,
			SkipHeader: true,
		},
	}
	assert.Equal(t, "/data/test.csv", req.FilePath)
	assert.Equal(t, FormatCSV, req.Format)
}

func TestValidationConfig_Struct(t *testing.T) {
	cfg := ValidationConfig{
		CheckMissing:    true,
		CheckOutliers:   true,
		CheckRange:      true,
		CheckContinuity: false,
		RangeConstraints: map[string][2]float64{"temperature": {0, 100}},
		OutlierMethod:   "zscore",
	}
	assert.True(t, cfg.CheckMissing)
	assert.Equal(t, "zscore", cfg.OutlierMethod)
}

func TestImportOptions_Struct(t *testing.T) {
	opts := ImportOptions{
		BatchSize:   500,
		SkipHeader:  true,
		Delimiter:   ",",
		TimeFormat:  "2006-01-02 15:04:05",
		TimeZone:    "Asia/Shanghai",
		Deduplicate: true,
		FillMissing: false,
	}
	assert.Equal(t, 500, opts.BatchSize)
	assert.True(t, opts.SkipHeader)
}

func TestImportJob_Struct(t *testing.T) {
	now := time.Now()
	job := &ImportJob{
		ID:             "job-1",
		FilePath:       "/data/test.csv",
		Format:         FormatCSV,
		Status:         JobStatusRunning,
		TotalRecords:   1000,
		SuccessRecords: 500,
		FailedRecords:  10,
		StartedAt:      &now,
		CreatedAt:      now,
	}
	assert.Equal(t, "job-1", job.ID)
	assert.Equal(t, JobStatusRunning, job.Status)
}

func TestDataQualityReport_Struct(t *testing.T) {
	report := &DataQualityReport{
		ID:               1,
		StationID:        "station-1",
		TotalPoints:      1000,
		GoodPoints:       950,
		SuspectPoints:    30,
		BadPoints:        10,
		MissingPoints:    10,
		CompletenessRate: 0.99,
		AccuracyRate:     0.95,
	}
	assert.Equal(t, "station-1", report.StationID)
	assert.Equal(t, 0.99, report.CompletenessRate)
}

func TestNewCSVImporter(t *testing.T) {
	importer := NewCSVImporter()
	assert.NotNil(t, importer)
	assert.NotNil(t, importer.validator)
	assert.NotNil(t, importer.cleaner)
}

func TestCSVImporter_Import(t *testing.T) {
	importer := NewCSVImporter()
	req := &BatchImportRequest{
		FilePath: "/nonexistent/file.csv",
		Format:   FormatCSV,
	}

	job, err := importer.Import(context.Background(), req)
	require.NoError(t, err)
	assert.NotEmpty(t, job.ID)
	assert.Equal(t, JobStatusPending, job.Status)

	time.Sleep(100 * time.Millisecond)

	status, err := importer.GetJobStatus(context.Background(), job.ID)
	require.NoError(t, err)
	assert.True(t, status.Status == JobStatusFailed || status.Status == JobStatusCompleted)
}

func TestCSVImporter_GetJobStatus_NotFound(t *testing.T) {
	importer := NewCSVImporter()
	_, err := importer.GetJobStatus(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestCSVImporter_CancelJob(t *testing.T) {
	importer := NewCSVImporter()
	req := &BatchImportRequest{
		FilePath: "/nonexistent/file.csv",
		Format:   FormatCSV,
	}

	job, _ := importer.Import(context.Background(), req)
	err := importer.CancelJob(context.Background(), job.ID)
	assert.NoError(t, err)
}

func TestCSVImporter_CancelJob_NotFound(t *testing.T) {
	importer := NewCSVImporter()
	err := importer.CancelJob(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestSimpleValidator(t *testing.T) {
	validator := NewSimpleValidator(ValidationConfig{
		CheckMissing:  true,
		CheckOutliers: true,
		CheckRange:    true,
		RangeConstraints: map[string][2]float64{
			"temperature": {0, 100},
		},
	})
	assert.NotNil(t, validator)

	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "temperature", Value: 50.0, Quality: QualityGood},
	}

	validated, err := validator.Validate(context.Background(), points)
	require.NoError(t, err)
	assert.NotEmpty(t, validated)
}

func TestSimpleValidator_MissingData(t *testing.T) {
	validator := NewSimpleValidator(ValidationConfig{
		CheckMissing: true,
	})

	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "temperature", Value: math.NaN(), Quality: QualityGood},
	}

	validated, err := validator.Validate(context.Background(), points)
	require.NoError(t, err)
	for _, p := range validated {
		if p.Metric == "temperature" {
			assert.Equal(t, QualityMissing, p.Quality)
		}
	}
}

func TestSimpleValidator_OutOfRange(t *testing.T) {
	validator := NewSimpleValidator(ValidationConfig{
		CheckRange: true,
		RangeConstraints: map[string][2]float64{
			"temperature": {0, 100},
		},
	})

	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "temperature", Value: 200.0, Quality: QualityGood},
	}

	validated, err := validator.Validate(context.Background(), points)
	require.NoError(t, err)
	for _, p := range validated {
		if p.Metric == "temperature" {
			assert.Equal(t, QualityBad, p.Quality)
		}
	}
}

func TestSimpleValidator_GenerateReport(t *testing.T) {
	validator := NewSimpleValidator(ValidationConfig{
		CheckMissing: true,
	})

	report, err := validator.GenerateReport(context.Background(), "station-1", "dev-1", "power", time.Now().Add(-1*time.Hour), time.Now())
	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "station-1", report.StationID)
}

func TestSimpleCleaner(t *testing.T) {
	cleaner := NewSimpleCleaner()
	assert.NotNil(t, cleaner)

	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
	}

	deduped, err := cleaner.Deduplicate(context.Background(), points)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(deduped), len(points))
}

func TestSimpleCleaner_FillMissing(t *testing.T) {
	cleaner := NewSimpleCleaner()

	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
		{Timestamp: time.Now().Add(2 * time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 200.0, Quality: QualityGood},
	}

	filled, err := cleaner.FillMissing(context.Background(), points, "linear")
	require.NoError(t, err)
	assert.NotNil(t, filled)
}

func TestSimpleCleaner_RemoveOutliers(t *testing.T) {
	cleaner := NewSimpleCleaner()

	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
		{Timestamp: time.Now().Add(1 * time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 10000.0, Quality: QualityGood},
		{Timestamp: time.Now().Add(2 * time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 105.0, Quality: QualityGood},
	}

	cleaned, err := cleaner.RemoveOutliers(context.Background(), points, "zscore")
	require.NoError(t, err)
	assert.NotNil(t, cleaned)
}

func TestBatchProcessor(t *testing.T) {
	bp := NewBatchProcessor(100)
	assert.NotNil(t, bp)

	req := &BatchImportRequest{
		FilePath: "/nonexistent/file.csv",
		Format:   FormatCSV,
	}

	job, err := bp.ProcessLargeFile(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, job)
	assert.Equal(t, JobStatusPending, job.Status)
}
