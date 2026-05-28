package datacollector

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAPIHandler(t *testing.T) {
	handler := NewAPIHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.importer)
	assert.NotNil(t, handler.validator)
	assert.NotNil(t, handler.cleaner)
}

func TestAPIHandler_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewAPIHandler()
	handler.RegisterRoutes(r)

	routes := r.Routes()
	assert.NotEmpty(t, routes)

	pathMap := make(map[string]string)
	for _, route := range routes {
		pathMap[route.Method+":"+route.Path] = route.Path
	}
	assert.Contains(t, pathMap, "POST:/api/v1/data/import")
	assert.Contains(t, pathMap, "GET:/api/v1/data/import/:job_id")
	assert.Contains(t, pathMap, "POST:/api/v1/data/collect")
	assert.Contains(t, pathMap, "POST:/api/v1/data/quality/check")
	assert.Contains(t, pathMap, "POST:/api/v1/data/clean")
	assert.Contains(t, pathMap, "GET:/api/v1/data/query")
}

func TestAPIHandler_ImportData_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewAPIHandler()
	handler.RegisterRoutes(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/data/import", nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAPIHandler_ImportData_FileNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewAPIHandler()
	handler.RegisterRoutes(r)

	body := `{"file_path":"/nonexistent/file.csv","format":"csv","mapping":{"timestamp":"0","value":"1","metric":"2"}}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/data/import", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var resp ImportDataResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.JobID)
	assert.Equal(t, JobStatusPending, resp.Status)
}

func TestAPIHandler_GetImportJobStatus_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewAPIHandler()
	handler.RegisterRoutes(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/data/import/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAPIHandler_CollectData_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewAPIHandler()
	handler.RegisterRoutes(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/data/collect", nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAPIHandler_CollectData_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewAPIHandler()
	handler.RegisterRoutes(r)

	now := time.Now()
	points := []*DataPoint{
		{Timestamp: now, StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityUnknown},
		{Timestamp: now.Add(time.Hour), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 50.0, Quality: QualityUnknown},
	}
	bodyBytes, _ := json.Marshal(CollectDataRequest{Points: points})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/data/collect", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp CollectDataResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.ReceivedCount)
}

func TestAPIHandler_CheckDataQuality_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewAPIHandler()
	handler.RegisterRoutes(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/data/quality/check", nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAPIHandler_CheckDataQuality_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewAPIHandler()
	handler.RegisterRoutes(r)

	now := time.Now()
	points := []*DataPoint{
		{Timestamp: now, StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityUnknown},
		{Timestamp: now.Add(time.Hour), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 50.0, Quality: QualityUnknown},
	}
	bodyBytes, _ := json.Marshal(CheckDataQualityRequest{
		Points: points,
		Config: ValidationConfig{CheckMissing: true},
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/data/quality/check", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp CheckDataQualityResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.TotalPoints)
	assert.GreaterOrEqual(t, resp.GoodPoints, 0)
}

func TestAPIHandler_CleanData_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewAPIHandler()
	handler.RegisterRoutes(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/data/clean", nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAPIHandler_CleanData_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewAPIHandler()
	handler.RegisterRoutes(r)

	now := time.Now()
	points := []*DataPoint{
		{Timestamp: now, StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
		{Timestamp: now, StationID: "s1", DeviceID: "d1", Metric: "power", Value: 101.0, Quality: QualityGood},
		{Timestamp: now.Add(time.Hour), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 0, Quality: QualityMissing},
	}
	bodyBytes, _ := json.Marshal(CleanDataRequest{
		Points: points,
		Options: CleanOptions{
			Deduplicate:    true,
			RemoveOutliers: true,
			FillMissing:    true,
			OutlierMethod:  "iqr",
			FillMethod:     "linear",
		},
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/data/clean", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp CleanDataResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 3, resp.OriginalCount)
}

func TestAPIHandler_QueryData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewAPIHandler()
	handler.RegisterRoutes(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/data/query", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

func TestSlidingWindowCleaner_DeduplicateWithWindow(t *testing.T) {
	cleaner := NewSlidingWindowCleaner(5 * time.Minute)

	baseTime := time.Now()
	points := []*DataPoint{
		{Timestamp: baseTime, StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0},
		{Timestamp: baseTime.Add(1 * time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 101.0},
		{Timestamp: baseTime.Add(2 * time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 102.0},
		{Timestamp: baseTime.Add(10 * time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 103.0},
	}

	result, err := cleaner.DeduplicateWithWindow(context.Background(), points)
	assert.NoError(t, err)
	assert.LessOrEqual(t, len(result), len(points))
}

func TestSlidingWindowCleaner_DeduplicateWithWindow_Empty(t *testing.T) {
	cleaner := NewSlidingWindowCleaner(5 * time.Minute)

	result, err := cleaner.DeduplicateWithWindow(context.Background(), []*DataPoint{})
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestSlidingWindowCleaner_DeduplicateWithWindow_Single(t *testing.T) {
	cleaner := NewSlidingWindowCleaner(5 * time.Minute)

	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0},
	}

	result, err := cleaner.DeduplicateWithWindow(context.Background(), points)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestSimpleCleaner_FillMissing_Backward(t *testing.T) {
	cleaner := NewSimpleCleaner()

	baseTime := time.Now()
	points := []*DataPoint{
		{Timestamp: baseTime, StationID: "s1", DeviceID: "d1", Metric: "power", Value: 0, Quality: QualityMissing},
		{Timestamp: baseTime.Add(time.Hour), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 120.0, Quality: QualityGood},
	}

	result, err := cleaner.FillMissing(context.Background(), points, "backward")
	assert.NoError(t, err)
	assert.Equal(t, QualityGood, result[0].Quality)
	assert.Equal(t, 120.0, result[0].Value)
}

func TestSimpleCleaner_FillMissing_Forward(t *testing.T) {
	cleaner := NewSimpleCleaner()

	baseTime := time.Now()
	points := []*DataPoint{
		{Timestamp: baseTime, StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
		{Timestamp: baseTime.Add(time.Hour), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 0, Quality: QualityMissing},
	}

	result, err := cleaner.FillMissing(context.Background(), points, "forward")
	assert.NoError(t, err)
	assert.Equal(t, QualityGood, result[1].Quality)
	assert.Equal(t, 100.0, result[1].Value)
}

func TestSimpleCleaner_FillMissing_Linear_FirstMissing(t *testing.T) {
	cleaner := NewSimpleCleaner()

	baseTime := time.Now()
	points := []*DataPoint{
		{Timestamp: baseTime, StationID: "s1", DeviceID: "d1", Metric: "power", Value: 0, Quality: QualityMissing},
		{Timestamp: baseTime.Add(time.Hour), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
	}

	result, err := cleaner.FillMissing(context.Background(), points, "linear")
	assert.NoError(t, err)
	assert.Equal(t, QualityMissing, result[0].Quality)
}

func TestSimpleCleaner_FillMissing_Linear_LastMissing(t *testing.T) {
	cleaner := NewSimpleCleaner()

	baseTime := time.Now()
	points := []*DataPoint{
		{Timestamp: baseTime, StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
		{Timestamp: baseTime.Add(time.Hour), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 0, Quality: QualityMissing},
	}

	result, err := cleaner.FillMissing(context.Background(), points, "linear")
	assert.NoError(t, err)
	assert.Equal(t, QualityMissing, result[1].Quality)
}

func TestSimpleCleaner_FillMissing_Linear_AdjacentMissing(t *testing.T) {
	cleaner := NewSimpleCleaner()

	baseTime := time.Now()
	points := []*DataPoint{
		{Timestamp: baseTime, StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
		{Timestamp: baseTime.Add(time.Hour), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 0, Quality: QualityMissing},
		{Timestamp: baseTime.Add(2 * time.Hour), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 0, Quality: QualityMissing},
		{Timestamp: baseTime.Add(3 * time.Hour), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 200.0, Quality: QualityGood},
	}

	result, err := cleaner.FillMissing(context.Background(), points, "linear")
	assert.NoError(t, err)
	assert.Equal(t, QualityMissing, result[1].Quality)
	assert.Equal(t, QualityMissing, result[2].Quality)
}

func TestSimpleCleaner_FillMissing_Empty(t *testing.T) {
	cleaner := NewSimpleCleaner()

	result, err := cleaner.FillMissing(context.Background(), []*DataPoint{}, "linear")
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestSimpleCleaner_FillMissing_Single(t *testing.T) {
	cleaner := NewSimpleCleaner()

	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
	}

	result, err := cleaner.FillMissing(context.Background(), points, "linear")
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestSimpleCleaner_RemoveOutliers_3Sigma(t *testing.T) {
	cleaner := NewSimpleCleaner()

	baseTime := time.Now()
	var points []*DataPoint
	for i := 0; i < 20; i++ {
		points = append(points, &DataPoint{
			Timestamp: baseTime.Add(time.Duration(i) * time.Minute),
			StationID: "s1", DeviceID: "d1", Metric: "power",
			Value:   100.0 + float64(i%3),
			Quality: QualityGood,
		})
	}
	points = append(points, &DataPoint{
		Timestamp: baseTime.Add(20 * time.Minute),
		StationID: "s1", DeviceID: "d1", Metric: "power",
		Value:   100000.0,
		Quality: QualityGood,
	})

	result, err := cleaner.RemoveOutliers(context.Background(), points, "3sigma")
	assert.NoError(t, err)
	assert.Len(t, result, 21)

	outlierFound := false
	for _, p := range result {
		if p.Value == 100000.0 {
			assert.Equal(t, QualityBad, p.Quality)
			outlierFound = true
		}
	}
	assert.True(t, outlierFound)
}

func TestSimpleCleaner_RemoveOutliers_DefaultMethod(t *testing.T) {
	cleaner := NewSimpleCleaner()

	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
		{Timestamp: time.Now().Add(time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 101.0, Quality: QualityGood},
		{Timestamp: time.Now().Add(2 * time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 99.0, Quality: QualityGood},
		{Timestamp: time.Now().Add(3 * time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.5, Quality: QualityGood},
	}

	result, err := cleaner.RemoveOutliers(context.Background(), points, "unknown_method")
	assert.NoError(t, err)
	assert.Len(t, result, 4)
}

func TestSimpleCleaner_RemoveOutliers_FewPoints(t *testing.T) {
	cleaner := NewSimpleCleaner()

	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
	}

	result, err := cleaner.RemoveOutliers(context.Background(), points, "iqr")
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestSimpleCleaner_RemoveOutliers_MissingValues(t *testing.T) {
	cleaner := NewSimpleCleaner()

	baseTime := time.Now()
	points := []*DataPoint{
		{Timestamp: baseTime, StationID: "s1", DeviceID: "d1", Metric: "power", Value: math.NaN(), Quality: QualityMissing},
		{Timestamp: baseTime.Add(time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
		{Timestamp: baseTime.Add(2 * time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 101.0, Quality: QualityGood},
		{Timestamp: baseTime.Add(3 * time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 99.0, Quality: QualityGood},
		{Timestamp: baseTime.Add(4 * time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.5, Quality: QualityGood},
	}

	result, err := cleaner.RemoveOutliers(context.Background(), points, "iqr")
	assert.NoError(t, err)
	assert.Len(t, result, 5)
}

func TestSimpleCleaner_Deduplicate_Empty(t *testing.T) {
	cleaner := NewSimpleCleaner()

	result, err := cleaner.Deduplicate(context.Background(), []*DataPoint{})
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestSimpleCleaner_Deduplicate_Single(t *testing.T) {
	cleaner := NewSimpleCleaner()

	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0},
	}

	result, err := cleaner.Deduplicate(context.Background(), points)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestCSVImporter_ImportJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	data := []map[string]interface{}{
		{"timestamp": "2024-01-01T00:00:00Z", "station_id": "s1", "device_id": "d1", "metric": "power", "value": 100.0, "unit": "kW"},
		{"timestamp": "2024-01-01T01:00:00Z", "station_id": "s1", "device_id": "d1", "metric": "power", "value": 110.0, "unit": "kW"},
	}
	jsonData, _ := json.Marshal(data)
	tmpFile.Write(jsonData)
	tmpFile.Close()

	importer := NewCSVImporter()
	req := &BatchImportRequest{
		FilePath: tmpFile.Name(),
		Format:   FormatJSON,
		Mapping:  map[string]string{},
	}

	job, err := importer.Import(context.Background(), req)
	require.NoError(t, err)
	assert.NotEmpty(t, job.ID)

	time.Sleep(200 * time.Millisecond)

	status, err := importer.GetJobStatus(context.Background(), job.ID)
	require.NoError(t, err)
	assert.Equal(t, JobStatusCompleted, status.Status)
	assert.Equal(t, 2, status.SuccessRecords)
}

func TestCSVImporter_ImportJSON_InvalidFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("invalid json")
	tmpFile.Close()

	importer := NewCSVImporter()
	req := &BatchImportRequest{
		FilePath: tmpFile.Name(),
		Format:   FormatJSON,
	}

	job, err := importer.Import(context.Background(), req)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	status, err := importer.GetJobStatus(context.Background(), job.ID)
	require.NoError(t, err)
	assert.Equal(t, JobStatusFailed, status.Status)
}

func TestCSVImporter_Import_UnsupportedFormat(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("some data")
	tmpFile.Close()

	importer := NewCSVImporter()
	req := &BatchImportRequest{
		FilePath: tmpFile.Name(),
		Format:   DataFormat("xml"),
	}

	job, err := importer.Import(context.Background(), req)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	status, err := importer.GetJobStatus(context.Background(), job.ID)
	require.NoError(t, err)
	assert.Equal(t, JobStatusFailed, status.Status)
}

func TestCSVImporter_Import_WithDeduplicate(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.csv")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	writer := csv.NewWriter(tmpFile)
	writer.Write([]string{"timestamp", "station_id", "device_id", "metric", "value", "unit"})
	writer.Write([]string{"2024-01-01T00:00:00Z", "station-001", "device-001", "power", "100.0", "kW"})
	writer.Write([]string{"2024-01-01T00:00:00Z", "station-001", "device-001", "power", "101.0", "kW"})
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
			SkipHeader:  true,
			TimeFormat:  time.RFC3339,
			Deduplicate: true,
		},
	}

	job, err := importer.Import(context.Background(), req)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	status, err := importer.GetJobStatus(context.Background(), job.ID)
	require.NoError(t, err)
	assert.Equal(t, JobStatusCompleted, status.Status)
}

func TestCSVImporter_Import_WithFillMissing(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.csv")
	require.NoError(t, err)
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
			SkipHeader:  true,
			TimeFormat:  time.RFC3339,
			FillMissing: true,
			FillMethod:  "forward",
		},
	}

	job, err := importer.Import(context.Background(), req)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	status, err := importer.GetJobStatus(context.Background(), job.ID)
	require.NoError(t, err)
	assert.Equal(t, JobStatusCompleted, status.Status)
}

func TestCSVImporter_ParseTime_WithTimezone(t *testing.T) {
	importer := NewCSVImporter()

	t.Run("with timezone", func(t *testing.T) {
		parsed, err := importer.parseTime("2024-01-01T00:00:00Z", time.RFC3339, "Asia/Shanghai")
		assert.NoError(t, err)
		assert.False(t, parsed.IsZero())
	})

	t.Run("with invalid timezone", func(t *testing.T) {
		parsed, err := importer.parseTime("2024-01-01T00:00:00Z", time.RFC3339, "Invalid/Zone")
		assert.NoError(t, err)
		assert.False(t, parsed.IsZero())
	})

	t.Run("with empty format", func(t *testing.T) {
		parsed, err := importer.parseTime("2024-01-01T00:00:00Z", "", "")
		assert.NoError(t, err)
		assert.False(t, parsed.IsZero())
	})

	t.Run("with invalid time value", func(t *testing.T) {
		_, err := importer.parseTime("not-a-time", time.RFC3339, "")
		assert.Error(t, err)
	})
}

func TestCSVImporter_RecordToDataPoint_InvalidIndex(t *testing.T) {
	importer := NewCSVImporter()

	record := []string{"2024-01-01T00:00:00Z", "s1", "d1", "power", "100.0", "kW"}

	req := &BatchImportRequest{
		Mapping: map[string]string{
			"timestamp":  "0",
			"station_id": "1",
			"device_id":  "2",
			"metric":     "3",
			"value":      "4",
			"unit":       "5",
			"extra":      "999",
			"bad_index":  "abc",
			"neg_index":  "-1",
		},
	}

	point, err := importer.recordToDataPoint(record, req)
	assert.NoError(t, err)
	assert.NotNil(t, point)
	assert.Equal(t, "s1", point.StationID)
}

func TestCSVImporter_RecordToDataPoint_MissingRequired(t *testing.T) {
	importer := NewCSVImporter()

	record := []string{"s1", "d1"}

	req := &BatchImportRequest{
		Mapping: map[string]string{
			"station_id": "0",
			"device_id":  "1",
		},
	}

	point, err := importer.recordToDataPoint(record, req)
	assert.Error(t, err)
	assert.Nil(t, point)
}

func TestCSVImporter_CancelJob_Running(t *testing.T) {
	importer := NewCSVImporter()

	importer.jobsMutex.Lock()
	importer.jobs["running-job"] = &ImportJob{
		ID:     "running-job",
		Status: JobStatusRunning,
	}
	importer.jobsMutex.Unlock()

	err := importer.CancelJob(context.Background(), "running-job")
	assert.NoError(t, err)

	job, _ := importer.GetJobStatus(context.Background(), "running-job")
	assert.Equal(t, JobStatusPaused, job.Status)
}

func TestCSVImporter_CancelJob_Completed(t *testing.T) {
	importer := NewCSVImporter()

	importer.jobsMutex.Lock()
	importer.jobs["completed-job"] = &ImportJob{
		ID:     "completed-job",
		Status: JobStatusCompleted,
	}
	importer.jobsMutex.Unlock()

	err := importer.CancelJob(context.Background(), "completed-job")
	assert.NoError(t, err)

	job, _ := importer.GetJobStatus(context.Background(), "completed-job")
	assert.Equal(t, JobStatusCompleted, job.Status)
}

func TestBatchValidator_ValidateBatch_Empty(t *testing.T) {
	validator := NewBatchValidator(ValidationConfig{})

	result, err := validator.ValidateBatch(context.Background(), []*DataPoint{})
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestBatchValidator_CheckContinuity_TimeGap(t *testing.T) {
	validator := NewBatchValidator(ValidationConfig{
		CheckContinuity: true,
	})

	baseTime := time.Now()
	points := []*DataPoint{
		{Timestamp: baseTime, StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
		{Timestamp: baseTime.Add(5 * time.Hour), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 110.0, Quality: QualityGood},
	}

	result, err := validator.ValidateBatch(context.Background(), points)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, QualitySuspect, result[0].Quality)
	assert.Equal(t, QualitySuspect, result[1].Quality)
}

func TestBatchValidator_CheckContinuity_ValueJump(t *testing.T) {
	validator := NewBatchValidator(ValidationConfig{
		CheckContinuity: true,
	})

	baseTime := time.Now()
	points := []*DataPoint{
		{Timestamp: baseTime, StationID: "s1", DeviceID: "d1", Metric: "power", Value: 10.0, Quality: QualityGood},
		{Timestamp: baseTime.Add(time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 1000.0, Quality: QualityGood},
	}

	result, err := validator.ValidateBatch(context.Background(), points)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestBatchValidator_IsOutlierBatch_IQR(t *testing.T) {
	validator := NewBatchValidator(ValidationConfig{})

	assert.True(t, validator.isOutlierBatch(1000.0, 100.0, 10.0, 90.0, 110.0, "iqr"))
	assert.False(t, validator.isOutlierBatch(100.0, 100.0, 10.0, 90.0, 110.0, "iqr"))
}

func TestBatchValidator_IsOutlierBatch_Default(t *testing.T) {
	validator := NewBatchValidator(ValidationConfig{})

	assert.False(t, validator.isOutlierBatch(1000.0, 100.0, 10.0, 90.0, 110.0, "unknown"))
}

func TestSimpleValidator_IsOutlier_3Sigma(t *testing.T) {
	validator := NewSimpleValidator(ValidationConfig{
		CheckOutliers: true,
		OutlierMethod: "3sigma",
	})

	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityUnknown},
	}

	result, err := validator.Validate(context.Background(), points)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestSimpleValidator_IsOutlier_IQR(t *testing.T) {
	validator := NewSimpleValidator(ValidationConfig{
		CheckOutliers: true,
		OutlierMethod: "iqr",
	})

	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityUnknown},
	}

	result, err := validator.Validate(context.Background(), points)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestSimpleValidator_IsOutlier_Default(t *testing.T) {
	validator := NewSimpleValidator(ValidationConfig{
		CheckOutliers: true,
		OutlierMethod: "unknown",
	})

	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityUnknown},
	}

	result, err := validator.Validate(context.Background(), points)
	assert.NoError(t, err)
	assert.Equal(t, QualityGood, result[0].Quality)
}

func TestSimpleValidator_InfValue(t *testing.T) {
	validator := NewSimpleValidator(ValidationConfig{
		CheckMissing: true,
	})

	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power", Value: math.Inf(1), Quality: QualityUnknown},
	}

	result, err := validator.Validate(context.Background(), points)
	assert.NoError(t, err)
	assert.Equal(t, QualityMissing, result[0].Quality)
}

func TestSimpleValidator_OutlierOnBadQuality(t *testing.T) {
	validator := NewSimpleValidator(ValidationConfig{
		CheckOutliers: true,
		OutlierMethod: "3sigma",
	})

	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityBad},
	}

	result, err := validator.Validate(context.Background(), points)
	assert.NoError(t, err)
	assert.Equal(t, QualityBad, result[0].Quality)
}

func TestCalculateMeanAndStdDev_Empty(t *testing.T) {
	mean, stdDev := calculateMeanAndStdDev([]float64{})
	assert.Equal(t, 0.0, mean)
	assert.Equal(t, 0.0, stdDev)
}

func TestCalculateQuartiles_Empty(t *testing.T) {
	q1, q3 := calculateQuartiles([]float64{})
	assert.Equal(t, 0.0, q1)
	assert.Equal(t, 0.0, q3)
}

func TestBatchProcessor_ProcessLargeFile(t *testing.T) {
	bp := NewBatchProcessor(100)

	tmpFile, err := os.CreateTemp("", "test-*.csv")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	writer := csv.NewWriter(tmpFile)
	writer.Write([]string{"timestamp", "station_id", "device_id", "metric", "value", "unit"})
	writer.Write([]string{"2024-01-01T00:00:00Z", "s1", "d1", "power", "100.0", "kW"})
	writer.Flush()
	tmpFile.Close()

	req := &BatchImportRequest{
		FilePath: tmpFile.Name(),
		Format:   FormatCSV,
		Mapping: map[string]string{
			"timestamp": "0", "station_id": "1", "device_id": "2",
			"metric": "3", "value": "4", "unit": "5",
		},
		Options: ImportOptions{SkipHeader: true, TimeFormat: time.RFC3339},
	}

	job, err := bp.ProcessLargeFile(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, job)
}

func TestAdvancedCleaner_CleanPipeline_NoOpts(t *testing.T) {
	cleaner := NewAdvancedCleaner()

	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
	}

	result, err := cleaner.CleanPipeline(context.Background(), points, CleanOptions{})
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestAdvancedCleaner_CleanPipeline_3SigmaOutlier(t *testing.T) {
	cleaner := NewAdvancedCleaner()

	baseTime := time.Now()
	points := []*DataPoint{
		{Timestamp: baseTime, StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
		{Timestamp: baseTime.Add(time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 101.0, Quality: QualityGood},
		{Timestamp: baseTime.Add(2 * time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 99.0, Quality: QualityGood},
		{Timestamp: baseTime.Add(3 * time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.5, Quality: QualityGood},
		{Timestamp: baseTime.Add(4 * time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 10000.0, Quality: QualityGood},
	}

	opts := CleanOptions{
		RemoveOutliers: true,
		OutlierMethod:  "3sigma",
	}

	result, err := cleaner.CleanPipeline(context.Background(), points, opts)
	assert.NoError(t, err)
	assert.Len(t, result, 5)
}
