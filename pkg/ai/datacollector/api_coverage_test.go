package datacollector

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAPIHandler(t *testing.T) {
	h := NewAPIHandler()
	assert.NotNil(t, h)
	assert.NotNil(t, h.importer)
	assert.NotNil(t, h.validator)
	assert.NotNil(t, h.cleaner)
}

func TestAPIHandler_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewAPIHandler()
	h.RegisterRoutes(r)
}

func TestAPIHandler_ImportData_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = createJSONRequest(map[string]interface{}{
		"file_path": "/tmp/test.csv",
		"format":    "csv",
		"mapping":   map[string]string{"timestamp": "0", "value": "1"},
	})

	h := NewAPIHandler()
	h.ImportData(c)

	assert.Equal(t, http.StatusAccepted, w.Code)
	var resp ImportDataResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.JobID)
	assert.Equal(t, JobStatusPending, resp.Status)
}

func TestAPIHandler_ImportData_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = createJSONRequest("invalid json")

	h := NewAPIHandler()
	h.ImportData(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAPIHandler_GetImportJobStatus_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "job_id", Value: "nonexistent"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/datacollector/import/nonexistent", nil)

	h := NewAPIHandler()
	h.GetImportJobStatus(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAPIHandler_CollectData_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Unit: "kW", Quality: QualityGood},
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = createJSONRequest(map[string]interface{}{
		"points": points,
	})

	h := NewAPIHandler()
	h.CollectData(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp CollectDataResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.ReceivedCount)
	assert.Equal(t, 1, resp.SuccessCount)
}

func TestAPIHandler_CollectData_MissingValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 0, Unit: "kW", Quality: QualityMissing},
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = createJSONRequest(map[string]interface{}{
		"points": points,
	})

	h := NewAPIHandler()
	h.CollectData(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp CollectDataResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 1, resp.FailedCount)
}

func TestAPIHandler_CollectData_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = createJSONRequest(nil)

	h := NewAPIHandler()
	h.CollectData(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAPIHandler_CheckDataQuality(t *testing.T) {
	gin.SetMode(gin.TestMode)
	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 0, Quality: QualityMissing},
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = createJSONRequest(map[string]interface{}{
		"points":    points,
		"station_id": "s1",
		"device_id":  "d1",
		"metric":     "power",
	})

	h := NewAPIHandler()
	h.CheckDataQuality(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp CheckDataQualityResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.TotalPoints)
	assert.Greater(t, resp.GoodPoints, 0)
}

func TestAPIHandler_CheckDataQuality_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = createJSONRequest(nil)

	h := NewAPIHandler()
	h.CheckDataQuality(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAPIHandler_CleanData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = createJSONRequest(map[string]interface{}{
		"points": points,
		"options": CleanOptions{Deduplicate: true, RemoveOutliers: true, FillMissing: true},
	})

	h := NewAPIHandler()
	h.CleanData(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp CleanDataResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.OriginalCount)
}

func TestAPIHandler_CleanData_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = createJSONRequest(nil)

	h := NewAPIHandler()
	h.CleanData(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAPIHandler_QueryData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	h := NewAPIHandler()
	h.QueryData(c)

	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

func TestSlidingWindowCleaner_New(t *testing.T) {
	swc := NewSlidingWindowCleaner(5 * time.Minute)
	assert.NotNil(t, swc)
	assert.Equal(t, 5*time.Minute, swc.windowSize)
}

func TestSlidingWindowCleaner_DeduplicateWithWindow_Empty(t *testing.T) {
	swc := NewSlidingWindowCleaner(5 * time.Minute)
	result, err := swc.DeduplicateWithWindow(context.Background(), []*DataPoint{})
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestSlidingWindowCleaner_DeduplicateWithWindow_Single(t *testing.T) {
	swc := NewSlidingWindowCleaner(5 * time.Minute)
	point := &DataPoint{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power"}
	result, err := swc.DeduplicateWithWindow(context.Background(), []*DataPoint{point})
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestSlidingWindowCleaner_DeduplicateWithWindow_DuplicatesInWindow(t *testing.T) {
	swc := NewSlidingWindowCleaner(10 * time.Minute)
	baseTime := time.Now()
	points := []*DataPoint{
		{Timestamp: baseTime, StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0},
		{Timestamp: baseTime.Add(1 * time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 101.0},
		{Timestamp: baseTime.Add(15 * time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 102.0},
	}
	result, err := swc.DeduplicateWithWindow(context.Background(), points)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(result), len(points))
}

func TestAdvancedCleaner_CleanPipeline_NoOptions(t *testing.T) {
	ac := NewAdvancedCleaner()
	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
	}
	result, err := ac.CleanPipeline(context.Background(), points, CleanOptions{})
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestAdvancedCleaner_CleanPipeline_AllOptions(t *testing.T) {
	ac := NewAdvancedCleaner()
	baseTime := time.Now()
	points := []*DataPoint{
		{Timestamp: baseTime, StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
		{Timestamp: baseTime, StationID: "s1", DeviceID: "d1", Metric: "power", Value: 101.0, Quality: QualityGood},
		{Timestamp: baseTime.Add(1 * time.Hour), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 0, Quality: QualityMissing},
		{Timestamp: baseTime.Add(2 * time.Hour), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 120.0, Quality: QualityGood},
	}
	opts := CleanOptions{
		Deduplicate:    true,
		RemoveOutliers: true,
		FillMissing:    true,
		OutlierMethod:  "iqr",
		FillMethod:     "linear",
	}
	result, err := ac.CleanPipeline(context.Background(), points, opts)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestSimpleCleaner_FillMissing_Backward(t *testing.T) {
	cleaner := NewSimpleCleaner()
	baseTime := time.Now()
	points := []*DataPoint{
		{Timestamp: baseTime, StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
		{Timestamp: baseTime.Add(2 * time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 0, Quality: QualityMissing},
		{Timestamp: baseTime.Add(4 * time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 200.0, Quality: QualityGood},
	}
	result, err := cleaner.FillMissing(context.Background(), points, "backward")
	require.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, QualityGood, result[1].Quality)
	assert.InDelta(t, 200.0, result[1].Value, 0.01)
}

func TestSimpleCleaner_FillMissing_SinglePoint(t *testing.T) {
	cleaner := NewSimpleCleaner()
	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
	}
	result, err := cleaner.FillMissing(context.Background(), points, "linear")
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestSimpleCleaner_RemoveOutliers_DefaultMethod(t *testing.T) {
	cleaner := NewSimpleCleaner()
	points := []*DataPoint{
		{Timestamp: time.Now(), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
		{Timestamp: time.Now().Add(1 * time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 200.0, Quality: QualityGood},
	}
	result, err := cleaner.RemoveOutliers(context.Background(), points, "unknown")
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestBatchValidator_ValidateBatch_Empty(t *testing.T) {
	bv := NewBatchValidator(ValidationConfig{CheckContinuity: true})
	result, err := bv.ValidateBatch(context.Background(), []*DataPoint{})
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestBatchValidator_CheckContinuity_TimeGap(t *testing.T) {
	bv := NewBatchValidator(ValidationConfig{CheckContinuity: true})
	baseTime := time.Now()
	points := []*DataPoint{
		{Timestamp: baseTime, StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
		{Timestamp: baseTime.Add(3 * time.Hour), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 110.0, Quality: QualityGood},
	}
	result, err := bv.ValidateBatch(context.Background(), points)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, QualitySuspect, result[0].Quality)
	assert.Equal(t, QualitySuspect, result[1].Quality)
}

func TestBatchValidator_CheckContinuity_ValueJump(t *testing.T) {
	bv := NewBatchValidator(ValidationConfig{CheckContinuity: true})
	baseTime := time.Now()
	points := []*DataPoint{
		{Timestamp: baseTime, StationID: "s1", DeviceID: "d1", Metric: "power", Value: 100.0, Quality: QualityGood},
		{Timestamp: baseTime.Add(1 * time.Minute), StationID: "s1", DeviceID: "d1", Metric: "power", Value: 500.0, Quality: QualityGood},
	}
	result, err := bv.ValidateBatch(context.Background(), points)
	require.NoError(t, err)
	assert.Equal(t, QualityGood, result[1].Quality)
}

func TestDataCollectorInterface_Compliance(t *testing.T) {
	var _ DataCollector = (*mockDataCollector)(nil)
	var _ DataImporter = (*CSVImporter)(nil)
	var _ DataValidator = (*SimpleValidator)(nil)
	var _ DataCleaner = (*SimpleCleaner)(nil)
}

type mockDataCollector struct{}

func (m *mockDataCollector) Collect(ctx context.Context) (<-chan *DataPoint, error) { return nil, nil }
func (m *mockDataCollector) Close() error                                         { return nil }

func TestImportDataRequest_Struct(t *testing.T) {
	req := ImportDataRequest{
		FilePath: "/data/test.csv",
		Format:   FormatCSV,
		Mapping: map[string]string{"timestamp": "0"},
		Validation: ValidationConfig{CheckMissing: true},
		Options: ImportOptions{SkipHeader: true},
	}
	assert.Equal(t, "/data/test.csv", req.FilePath)
}

func TestImportDataResponse_Struct(t *testing.T) {
	resp := ImportDataResponse{JobID: "job-1", Status: JobStatusRunning, Message: "processing"}
	assert.Equal(t, "job-1", resp.JobID)
}

func TestCollectDataRequest_Struct(t *testing.T) {
	req := CollectDataRequest{Points: []*DataPoint{{Metric: "power"}}}
	assert.Len(t, req.Points, 1)
}

func TestCollectDataResponse_Struct(t *testing.T) {
	resp := CollectDataResponse{ReceivedCount: 10, SuccessCount: 8, FailedCount: 2}
	assert.Equal(t, 8, resp.SuccessCount)
}

func TestCheckDataQualityRequest_Struct(t *testing.T) {
	req := CheckDataQualityRequest{
		Config:    ValidationConfig{},
		StationID: "s1",
		Metric:    "power",
	}
	assert.Equal(t, "s1", req.StationID)
}

func TestCheckDataQualityResponse_Struct(t *testing.T) {
	resp := CheckDataQualityResponse{
		TotalPoints:   100,
		GoodPoints:    90,
		SuspectPoints: 5,
		BadPoints:     5,
		QualityScore:  0.925,
	}
	assert.InDelta(t, 0.925, resp.QualityScore, 0.001)
}

func TestCleanDataRequest_Struct(t *testing.T) {
	req := CleanDataRequest{
		Points:  []*DataPoint{{Metric: "power"}},
		Options: CleanOptions{Deduplicate: true},
	}
	assert.True(t, req.Options.Deduplicate)
}

func TestCleanDataResponse_Struct(t *testing.T) {
	resp := CleanDataResponse{
		OriginalCount: 10,
		CleanedCount:  8,
		CleanedPoints: []*DataPoint{{Metric: "power"}},
	}
	assert.Equal(t, 10, resp.OriginalCount)
}

func TestQueryDataRequest_Struct(t *testing.T) {
	req := QueryDataRequest{
		StationID: "s1",
		DeviceID:  "d1",
		Metric:    "power",
		Limit:     100,
	}
	assert.Equal(t, 100, req.Limit)
}

func TestCleanOptions_Struct(t *testing.T) {
	opts := CleanOptions{
		Deduplicate:    true,
		RemoveOutliers: true,
		FillMissing:    true,
		OutlierMethod:  "zscore",
		FillMethod:     "backward",
	}
	assert.Equal(t, "backward", opts.FillMethod)
}

func createJSONRequest(data interface{}) *http.Request {
	body, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}
