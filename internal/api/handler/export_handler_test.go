package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockExportService 导出服务Mock
type MockExportService struct {
	mock.Mock
}

func (m *MockExportService) Export(ctx context.Context, req *service.ExportRequest) (*service.ExportResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.ExportResult), args.Error(1)
}

func init() {
	gin.SetMode(gin.TestMode)
}

func TestExportHandler_Export(t *testing.T) {
	t.Run("成功导出告警数据Excel格式", func(t *testing.T) {
		mockService := new(MockExportService)
		handler := NewExportHandler(mockService)

		req := map[string]interface{}{
			"type":   "alarm",
			"format": "excel",
		}

		expectedResult := &service.ExportResult{
			Buffer:      bytes.NewBuffer([]byte("test excel data")),
			Filename:    "alarms_20240101120000.xlsx",
			ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		}

		mockService.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(expectedResult, nil)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/export", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Export(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment; filename=")
		mockService.AssertExpectations(t)
	})

	t.Run("成功导出设备数据CSV格式", func(t *testing.T) {
		mockService := new(MockExportService)
		handler := NewExportHandler(mockService)

		req := map[string]interface{}{
			"type":   "device",
			"format": "csv",
		}

		expectedResult := &service.ExportResult{
			Buffer:      bytes.NewBuffer([]byte("test csv data")),
			Filename:    "devices_20240101120000.csv",
			ContentType: "text/csv",
		}

		mockService.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(expectedResult, nil)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/export", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Export(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
		mockService.AssertExpectations(t)
	})

	t.Run("无效的请求参数", func(t *testing.T) {
		mockService := new(MockExportService)
		handler := NewExportHandler(mockService)

		body := []byte(`{"invalid": "data"}`)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/export", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Export(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("服务返回错误", func(t *testing.T) {
		mockService := new(MockExportService)
		handler := NewExportHandler(mockService)

		req := map[string]interface{}{
			"type":   "alarm",
			"format": "excel",
		}

		mockService.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(nil, errors.New("database error"))

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/export", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Export(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestExportHandler_ExportAlarms(t *testing.T) {
	t.Run("成功导出告警数据", func(t *testing.T) {
		mockService := new(MockExportService)
		handler := NewExportHandler(mockService)

		expectedResult := &service.ExportResult{
			Buffer:      bytes.NewBuffer([]byte("test data")),
			Filename:    "alarms_20240101120000.xlsx",
			ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		}

		mockService.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(expectedResult, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/export/alarms?format=excel", nil)

		handler.ExportAlarms(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("使用默认格式", func(t *testing.T) {
		mockService := new(MockExportService)
		handler := NewExportHandler(mockService)

		expectedResult := &service.ExportResult{
			Buffer:      bytes.NewBuffer([]byte("test data")),
			Filename:    "alarms_20240101120000.xlsx",
			ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		}

		mockService.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(expectedResult, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/export/alarms", nil)

		handler.ExportAlarms(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("服务返回错误", func(t *testing.T) {
		mockService := new(MockExportService)
		handler := NewExportHandler(mockService)

		mockService.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(nil, errors.New("database error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/export/alarms", nil)

		handler.ExportAlarms(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestExportHandler_ExportDevices(t *testing.T) {
	t.Run("成功导出设备数据", func(t *testing.T) {
		mockService := new(MockExportService)
		handler := NewExportHandler(mockService)

		expectedResult := &service.ExportResult{
			Buffer:      bytes.NewBuffer([]byte("test data")),
			Filename:    "devices_20240101120000.xlsx",
			ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		}

		mockService.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(expectedResult, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/export/devices?format=excel&station_id=station-001", nil)

		handler.ExportDevices(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("带设备类型过滤", func(t *testing.T) {
		mockService := new(MockExportService)
		handler := NewExportHandler(mockService)

		expectedResult := &service.ExportResult{
			Buffer:      bytes.NewBuffer([]byte("test data")),
			Filename:    "devices_20240101120000.xlsx",
			ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		}

		mockService.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(expectedResult, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/export/devices?type=inverter", nil)

		handler.ExportDevices(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("服务返回错误", func(t *testing.T) {
		mockService := new(MockExportService)
		handler := NewExportHandler(mockService)

		mockService.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(nil, errors.New("database error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/export/devices", nil)

		handler.ExportDevices(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestExportHandler_ExportStations(t *testing.T) {
	t.Run("成功导出厂站数据", func(t *testing.T) {
		mockService := new(MockExportService)
		handler := NewExportHandler(mockService)

		expectedResult := &service.ExportResult{
			Buffer:      bytes.NewBuffer([]byte("test data")),
			Filename:    "stations_20240101120000.xlsx",
			ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		}

		mockService.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(expectedResult, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/export/stations?format=excel", nil)

		handler.ExportStations(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("带区域过滤", func(t *testing.T) {
		mockService := new(MockExportService)
		handler := NewExportHandler(mockService)

		expectedResult := &service.ExportResult{
			Buffer:      bytes.NewBuffer([]byte("test data")),
			Filename:    "stations_20240101120000.xlsx",
			ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		}

		mockService.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(expectedResult, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/export/stations?sub_region_id=region-001", nil)

		handler.ExportStations(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("带厂站类型过滤", func(t *testing.T) {
		mockService := new(MockExportService)
		handler := NewExportHandler(mockService)

		expectedResult := &service.ExportResult{
			Buffer:      bytes.NewBuffer([]byte("test data")),
			Filename:    "stations_20240101120000.xlsx",
			ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		}

		mockService.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(expectedResult, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/export/stations?type=pv", nil)

		handler.ExportStations(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("服务返回错误", func(t *testing.T) {
		mockService := new(MockExportService)
		handler := NewExportHandler(mockService)

		mockService.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(nil, errors.New("database error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/export/stations", nil)

		handler.ExportStations(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}
