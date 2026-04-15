package datacollector

import (
	"context"
	"encoding/csv"
	"math"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDataPointValidation(t *testing.T) {
	tests := []struct {
		name        string
		point       *DataPoint
		config      ValidationConfig
		expectedQuality DataQuality
	}{
		{
			name: "good data point",
			point: &DataPoint{
				Timestamp: time.Now(),
				StationID: "station-001",
				DeviceID:  "device-001",
				Metric:    "power",
				Value:     100.0,
				Unit:      "kW",
				Quality:   QualityUnknown,
			},
			config: ValidationConfig{
				CheckMissing: true,
			},
			expectedQuality: QualityGood,
		},
		{
			name: "missing value",
			point: &DataPoint{
				Timestamp: time.Now(),
				StationID: "station-001",
				DeviceID:  "device-001",
				Metric:    "power",
				Value:     math.NaN(),
				Unit:      "kW",
				Quality:   QualityUnknown,
			},
			config: ValidationConfig{
				CheckMissing: true,
			},
			expectedQuality: QualityMissing,
		},
		{
			name: "out of range",
			point: &DataPoint{
				Timestamp: time.Now(),
				StationID: "station-001",
				DeviceID:  "device-001",
				Metric:    "power",
				Value:     1000.0,
				Unit:      "kW",
				Quality:   QualityUnknown,
			},
			config: ValidationConfig{
				CheckRange: true,
				RangeConstraints: map[string][2]float64{
					"power": {0, 500},
				},
			},
			expectedQuality: QualityBad,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewSimpleValidator(tt.config)
			result, err := validator.Validate(context.Background(), []*DataPoint{tt.point})
			
			assert.NoError(t, err)
			assert.Len(t, result, 1)
			assert.Equal(t, tt.expectedQuality, result[0].Quality)
		})
	}
}

func TestDeduplication(t *testing.T) {
	cleaner := NewSimpleCleaner()
	
	baseTime := time.Now()
	points := []*DataPoint{
		{
			Timestamp: baseTime,
			StationID: "station-001",
			DeviceID:  "device-001",
			Metric:    "power",
			Value:     100.0,
		},
		{
			Timestamp: baseTime,
			StationID: "station-001",
			DeviceID:  "device-001",
			Metric:    "power",
			Value:     101.0,
		},
		{
			Timestamp: baseTime.Add(time.Hour),
			StationID: "station-001",
			DeviceID:  "device-001",
			Metric:    "power",
			Value:     110.0,
		},
	}
	
	result, err := cleaner.Deduplicate(context.Background(), points)
	
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestFillMissing(t *testing.T) {
	cleaner := NewSimpleCleaner()
	
	baseTime := time.Now()
	points := []*DataPoint{
		{
			Timestamp: baseTime,
			StationID: "station-001",
			DeviceID:  "device-001",
			Metric:    "power",
			Value:     100.0,
			Quality:   QualityGood,
		},
		{
			Timestamp: baseTime.Add(time.Hour),
			StationID: "station-001",
			DeviceID:  "device-001",
			Metric:    "power",
			Value:     0,
			Quality:   QualityMissing,
		},
		{
			Timestamp: baseTime.Add(2 * time.Hour),
			StationID: "station-001",
			DeviceID:  "device-001",
			Metric:    "power",
			Value:     120.0,
			Quality:   QualityGood,
		},
	}
	
	result, err := cleaner.FillMissing(context.Background(), points, "linear")
	
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, QualityGood, result[1].Quality)
	assert.InDelta(t, 110.0, result[1].Value, 0.01)
}

func TestRemoveOutliers(t *testing.T) {
	cleaner := NewSimpleCleaner()
	
	baseTime := time.Now()
	points := []*DataPoint{
		{
			Timestamp: baseTime,
			StationID: "station-001",
			DeviceID:  "device-001",
			Metric:    "power",
			Value:     100.0,
			Quality:   QualityGood,
		},
		{
			Timestamp: baseTime.Add(time.Hour),
			StationID: "station-001",
			DeviceID:  "device-001",
			Metric:    "power",
			Value:     101.0,
			Quality:   QualityGood,
		},
		{
			Timestamp: baseTime.Add(2 * time.Hour),
			StationID: "station-001",
			DeviceID:  "device-001",
			Metric:    "power",
			Value:     99.0,
			Quality:   QualityGood,
		},
		{
			Timestamp: baseTime.Add(3 * time.Hour),
			StationID: "station-001",
			DeviceID:  "device-001",
			Metric:    "power",
			Value:     200.0,
			Quality:   QualityGood,
		},
		{
			Timestamp: baseTime.Add(4 * time.Hour),
			StationID: "station-001",
			DeviceID:  "device-001",
			Metric:    "power",
			Value:     1000.0,
			Quality:   QualityGood,
		},
	}
	
	result, err := cleaner.RemoveOutliers(context.Background(), points, "iqr")
	
	assert.NoError(t, err)
	assert.Len(t, result, 5)
}

func TestCSVImport(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.csv")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	
	writer := csv.NewWriter(tmpFile)
	writer.Write([]string{"timestamp", "station_id", "device_id", "metric", "value", "unit"})
	writer.Write([]string{"2024-01-01T00:00:00Z", "station-001", "device-001", "power", "100.0", "kW"})
	writer.Write([]string{"2024-01-01T01:00:00Z", "station-001", "device-001", "power", "110.0", "kW"})
	writer.Flush()
	tmpFile.Close()
	
	importer := NewCSVImporter()
	
	req := &BatchImportRequest{
		FilePath: tmpFile.Name(),
		Format:   FormatCSV,
		Mapping: map[string]string{
			"timestamp":  "0",
			"station_id": "1",
			"device_id":  "2",
			"metric":     "3",
			"value":      "4",
			"unit":       "5",
		},
		Options: ImportOptions{
			SkipHeader: true,
			TimeFormat: time.RFC3339,
		},
	}
	
	job, err := importer.Import(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, job)
	
	time.Sleep(100 * time.Millisecond)
	
	jobStatus, err := importer.GetJobStatus(context.Background(), job.ID)
	assert.NoError(t, err)
	assert.NotNil(t, jobStatus)
}

func TestBatchValidator(t *testing.T) {
	validator := NewBatchValidator(ValidationConfig{
		CheckOutliers:   true,
		CheckContinuity: true,
		OutlierMethod:   "3sigma",
	})
	
	baseTime := time.Now()
	points := []*DataPoint{
		{Timestamp: baseTime, StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityUnknown},
		{Timestamp: baseTime.Add(time.Hour), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 101.0, Quality: QualityUnknown},
		{Timestamp: baseTime.Add(2 * time.Hour), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 99.0, Quality: QualityUnknown},
		{Timestamp: baseTime.Add(3 * time.Hour), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 200.0, Quality: QualityUnknown},
		{Timestamp: baseTime.Add(4 * time.Hour), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 1000.0, Quality: QualityUnknown},
	}
	
	result, err := validator.ValidateBatch(context.Background(), points)
	
	assert.NoError(t, err)
	assert.Len(t, result, 5)
}

func TestAdvancedCleanerPipeline(t *testing.T) {
	cleaner := NewAdvancedCleaner()
	
	baseTime := time.Now()
	points := []*DataPoint{
		{Timestamp: baseTime, StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
		{Timestamp: baseTime, StationID: "s1", DeviceID: "d1", Metric: "power", Value: 101.0, Quality: QualityGood},
		{Timestamp: baseTime.Add(time.Hour), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 0, Quality: QualityMissing},
		{Timestamp: baseTime.Add(2 * time.Hour), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 1000.0, Quality: QualityGood},
	}
	
	opts := CleanOptions{
		Deduplicate:    true,
		RemoveOutliers: true,
		FillMissing:    true,
		OutlierMethod:  "iqr",
		FillMethod:     "forward",
	}
	
	result, err := cleaner.CleanPipeline(context.Background(), points, opts)
	
	assert.NoError(t, err)
	assert.Len(t, result, 3)
}

func TestMeanAndStdDev(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5}
	mean, stdDev := calculateMeanAndStdDev(values)
	
	assert.InDelta(t, 3.0, mean, 0.001)
	assert.InDelta(t, 1.414, stdDev, 0.001)
}

func TestQuartiles(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	q1, q3 := calculateQuartiles(values)
	
	assert.InDelta(t, 3.0, q1, 0.001)
	assert.InDelta(t, 8.0, q3, 0.001)
}

func TestSlidingWindowDeduplication(t *testing.T) {
	t.Skip("SlidingWindowDeduplication needs refinement")
}
