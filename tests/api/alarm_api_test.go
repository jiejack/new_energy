package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/api/dto"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAlarmService 告警服务Mock
type MockAlarmService struct {
	mock.Mock
}

func (m *MockAlarmService) CreateAlarm(ctx interface{}, req interface{}) (*entity.Alarm, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Alarm), args.Error(1)
}

func (m *MockAlarmService) GetAlarm(ctx interface{}, id string) (*entity.Alarm, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Alarm), args.Error(1)
}

func (m *MockAlarmService) GetActiveAlarms(ctx interface{}, stationID *string, level *entity.AlarmLevel) ([]*entity.Alarm, error) {
	args := m.Called(ctx, stationID, level)
	return args.Get(0).([]*entity.Alarm), args.Error(1)
}

func (m *MockAlarmService) GetHistoryAlarms(ctx interface{}, stationID *string, startTime, endTime int64) ([]*entity.Alarm, error) {
	args := m.Called(ctx, stationID, startTime, endTime)
	return args.Get(0).([]*entity.Alarm), args.Error(1)
}

func (m *MockAlarmService) AcknowledgeAlarm(ctx interface{}, id, by string) error {
	args := m.Called(ctx, id, by)
	return args.Error(0)
}

func (m *MockAlarmService) ClearAlarm(ctx interface{}, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAlarmService) CountAlarmsByLevel(ctx interface{}, stationID *string) (map[entity.AlarmLevel]int64, error) {
	args := m.Called(ctx, stationID)
	return args.Get(0).(map[entity.AlarmLevel]int64), args.Error(1)
}

func setupAlarmTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestAlarmAPI_GetAlarm_Success(t *testing.T) {
	router := setupAlarmTestRouter()
	mockAlarmService := new(MockAlarmService)

	triggeredAt := time.Now()
	expectedAlarm := &entity.Alarm{
		ID:          "alarm-001",
		PointID:     "point-001",
		DeviceID:    "device-001",
		StationID:   "station-001",
		Type:        entity.AlarmTypeLimit,
		Level:       entity.AlarmLevelWarning,
		Title:       "电压高限告警",
		Message:     "电压超过上限阈值",
		Value:       450.0,
		Threshold:   400.0,
		Status:      entity.AlarmStatusActive,
		TriggeredAt: triggeredAt,
	}

	mockAlarmService.On("GetAlarm", mock.Anything, "alarm-001").Return(expectedAlarm, nil)

	router.GET("/api/v1/alarms/:id", func(c *gin.Context) {
		id := c.Param("id")
		alarm, err := mockAlarmService.GetAlarm(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "告警不存在"})
			return
		}

		c.JSON(http.StatusOK, dto.Response{
			Code:    0,
			Message: "success",
			Data: dto.AlarmResponse{
				ID:          alarm.ID,
				PointID:     alarm.PointID,
				DeviceID:    alarm.DeviceID,
				StationID:   alarm.StationID,
				Type:        string(alarm.Type),
				Level:       int(alarm.Level),
				Title:       alarm.Title,
				Message:     alarm.Message,
				Value:       alarm.Value,
				Threshold:   alarm.Threshold,
				Status:      int(alarm.Status),
				TriggeredAt: alarm.TriggeredAt,
			},
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/alarms/alarm-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockAlarmService.AssertExpectations(t)
}

func TestAlarmAPI_GetActiveAlarms_Success(t *testing.T) {
	router := setupAlarmTestRouter()
	mockAlarmService := new(MockAlarmService)

	triggeredAt := time.Now()
	expectedAlarms := []*entity.Alarm{
		{
			ID:          "alarm-001",
			PointID:     "point-001",
			DeviceID:    "device-001",
			StationID:   "station-001",
			Type:        entity.AlarmTypeLimit,
			Level:       entity.AlarmLevelWarning,
			Title:       "电压高限告警",
			Status:      entity.AlarmStatusActive,
			TriggeredAt: triggeredAt,
		},
		{
			ID:          "alarm-002",
			PointID:     "point-002",
			DeviceID:    "device-001",
			StationID:   "station-001",
			Type:        entity.AlarmTypeDevice,
			Level:       entity.AlarmLevelCritical,
			Title:       "设备故障告警",
			Status:      entity.AlarmStatusActive,
			TriggeredAt: triggeredAt,
		},
	}

	mockAlarmService.On("GetActiveAlarms", mock.Anything, (*string)(nil), (*entity.AlarmLevel)(nil)).Return(expectedAlarms, nil)

	router.GET("/api/v1/alarms", func(c *gin.Context) {
		var stationID *string
		var level *entity.AlarmLevel

		if sid := c.Query("station_id"); sid != "" {
			stationID = &sid
		}
		if l := c.Query("level"); l != "" {
			var alarmLevel entity.AlarmLevel
			switch l {
			case "1":
				alarmLevel = entity.AlarmLevelInfo
			case "2":
				alarmLevel = entity.AlarmLevelWarning
			case "3":
				alarmLevel = entity.AlarmLevelMajor
			case "4":
				alarmLevel = entity.AlarmLevelCritical
			}
			level = &alarmLevel
		}

		alarms, err := mockAlarmService.GetActiveAlarms(c.Request.Context(), stationID, level)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: "查询失败"})
			return
		}

		resp := make([]dto.AlarmResponse, len(alarms))
		for i, a := range alarms {
			resp[i] = dto.AlarmResponse{
				ID:          a.ID,
				PointID:     a.PointID,
				DeviceID:    a.DeviceID,
				StationID:   a.StationID,
				Type:        string(a.Type),
				Level:       int(a.Level),
				Title:       a.Title,
				Message:     a.Message,
				Status:      int(a.Status),
				TriggeredAt: a.TriggeredAt,
			}
		}

		c.JSON(http.StatusOK, dto.PagedResponse{
			Code:     0,
			Message:  "success",
			Data:     resp,
			Total:    int64(len(resp)),
			Page:     1,
			PageSize: 20,
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/alarms", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.PagedResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, int64(2), resp.Total)

	mockAlarmService.AssertExpectations(t)
}

func TestAlarmAPI_AcknowledgeAlarm_Success(t *testing.T) {
	router := setupAlarmTestRouter()
	mockAlarmService := new(MockAlarmService)

	mockAlarmService.On("AcknowledgeAlarm", mock.Anything, "alarm-001", "operator-001").Return(nil)

	router.PUT("/api/v1/alarms/:id/ack", func(c *gin.Context) {
		id := c.Param("id")
		var req dto.AckAlarmRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "参数错误"})
			return
		}

		err := mockAlarmService.AcknowledgeAlarm(c.Request.Context(), id, req.Operator)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: "确认失败"})
			return
		}

		c.JSON(http.StatusOK, dto.Response{
			Code:    0,
			Message: "success",
		})
	})

	ackReq := dto.AckAlarmRequest{
		Operator: "operator-001",
		Comment:  "已确认，正在处理",
	}
	body, _ := json.Marshal(ackReq)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/alarms/alarm-001/ack", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockAlarmService.AssertExpectations(t)
}

func TestAlarmAPI_ClearAlarm_Success(t *testing.T) {
	router := setupAlarmTestRouter()
	mockAlarmService := new(MockAlarmService)

	mockAlarmService.On("ClearAlarm", mock.Anything, "alarm-001").Return(nil)

	router.PUT("/api/v1/alarms/:id/clear", func(c *gin.Context) {
		id := c.Param("id")
		var req dto.ClearAlarmRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "参数错误"})
			return
		}

		err := mockAlarmService.ClearAlarm(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: "清除失败"})
			return
		}

		c.JSON(http.StatusOK, dto.Response{
			Code:    0,
			Message: "success",
		})
	})

	clearReq := dto.ClearAlarmRequest{
		Operator: "operator-001",
		Comment:  "问题已解决",
	}
	body, _ := json.Marshal(clearReq)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/alarms/alarm-001/clear", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockAlarmService.AssertExpectations(t)
}

func TestAlarmAPI_GetHistoryAlarms_Success(t *testing.T) {
	router := setupAlarmTestRouter()
	mockAlarmService := new(MockAlarmService)

	triggeredAt := time.Now()
	expectedAlarms := []*entity.Alarm{
		{
			ID:          "alarm-001",
			PointID:     "point-001",
			DeviceID:    "device-001",
			StationID:   "station-001",
			Type:        entity.AlarmTypeLimit,
			Level:       entity.AlarmLevelWarning,
			Title:       "历史告警1",
			Status:      entity.AlarmStatusCleared,
			TriggeredAt: triggeredAt,
		},
	}

	stationID := "station-001"
	mockAlarmService.On("GetHistoryAlarms", mock.Anything, &stationID, mock.AnythingOfType("int64"), mock.AnythingOfType("int64")).Return(expectedAlarms, nil)

	router.GET("/api/v1/alarms/history", func(c *gin.Context) {
		var req dto.ListAlarmsRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "参数错误"})
			return
		}

		var stationID *string
		if req.StationID != "" {
			stationID = &req.StationID
		}

		alarms, err := mockAlarmService.GetHistoryAlarms(c.Request.Context(), stationID, req.StartTime, req.EndTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: "查询失败"})
			return
		}

		resp := make([]dto.AlarmResponse, len(alarms))
		for i, a := range alarms {
			resp[i] = dto.AlarmResponse{
				ID:          a.ID,
				PointID:     a.PointID,
				DeviceID:    a.DeviceID,
				StationID:   a.StationID,
				Type:        string(a.Type),
				Level:       int(a.Level),
				Title:       a.Title,
				Status:      int(a.Status),
				TriggeredAt: a.TriggeredAt,
			}
		}

		c.JSON(http.StatusOK, dto.PagedResponse{
			Code:     0,
			Message:  "success",
			Data:     resp,
			Total:    int64(len(resp)),
			Page:     req.Page,
			PageSize: req.PageSize,
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/alarms/history?station_id=station-001&start_time=1709414400000&end_time=1709500800000", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockAlarmService.AssertExpectations(t)
}

func TestAlarmAPI_GetAlarmStatistics_Success(t *testing.T) {
	router := setupAlarmTestRouter()
	mockAlarmService := new(MockAlarmService)

	expectedCounts := map[entity.AlarmLevel]int64{
		entity.AlarmLevelInfo:     5,
		entity.AlarmLevelWarning:  10,
		entity.AlarmLevelMajor:    3,
		entity.AlarmLevelCritical: 1,
	}

	mockAlarmService.On("CountAlarmsByLevel", mock.Anything, (*string)(nil)).Return(expectedCounts, nil)

	router.GET("/api/v1/alarms/statistics", func(c *gin.Context) {
		var stationID *string
		if sid := c.Query("station_id"); sid != "" {
			stationID = &sid
		}

		counts, err := mockAlarmService.CountAlarmsByLevel(c.Request.Context(), stationID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: "查询失败"})
			return
		}

		byLevel := make(map[int]int64)
		for level, count := range counts {
			byLevel[int(level)] = count
		}

		var total int64
		for _, count := range counts {
			total += count
		}

		c.JSON(http.StatusOK, dto.Response{
			Code:    0,
			Message: "success",
			Data: dto.AlarmStatisticsResponse{
				Total:   total,
				Active:  total,
				ByLevel: byLevel,
			},
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/alarms/statistics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockAlarmService.AssertExpectations(t)
}

func TestAlarmAPI_GetActiveAlarms_WithFilter(t *testing.T) {
	router := setupAlarmTestRouter()
	mockAlarmService := new(MockAlarmService)

	stationID := "station-001"
	level := entity.AlarmLevelWarning

	triggeredAt := time.Now()
	expectedAlarms := []*entity.Alarm{
		{
			ID:          "alarm-001",
			PointID:     "point-001",
			StationID:   "station-001",
			Type:        entity.AlarmTypeLimit,
			Level:       entity.AlarmLevelWarning,
			Title:       "警告级别告警",
			Status:      entity.AlarmStatusActive,
			TriggeredAt: triggeredAt,
		},
	}

	mockAlarmService.On("GetActiveAlarms", mock.Anything, &stationID, &level).Return(expectedAlarms, nil)

	router.GET("/api/v1/alarms", func(c *gin.Context) {
		var stationID *string
		var level *entity.AlarmLevel

		if sid := c.Query("station_id"); sid != "" {
			stationID = &sid
		}
		if l := c.Query("level"); l != "" {
			var alarmLevel entity.AlarmLevel
			switch l {
			case "1":
				alarmLevel = entity.AlarmLevelInfo
			case "2":
				alarmLevel = entity.AlarmLevelWarning
			case "3":
				alarmLevel = entity.AlarmLevelMajor
			case "4":
				alarmLevel = entity.AlarmLevelCritical
			}
			level = &alarmLevel
		}

		alarms, err := mockAlarmService.GetActiveAlarms(c.Request.Context(), stationID, level)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: "查询失败"})
			return
		}

		resp := make([]dto.AlarmResponse, len(alarms))
		for i, a := range alarms {
			resp[i] = dto.AlarmResponse{
				ID:          a.ID,
				StationID:   a.StationID,
				Level:       int(a.Level),
				Title:       a.Title,
				Status:      int(a.Status),
				TriggeredAt: a.TriggeredAt,
			}
		}

		c.JSON(http.StatusOK, dto.PagedResponse{
			Code:     0,
			Message:  "success",
			Data:     resp,
			Total:    int64(len(resp)),
			Page:     1,
			PageSize: 20,
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/alarms?station_id=station-001&level=2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockAlarmService.AssertExpectations(t)
}

func TestAlarmAPI_GetAlarm_NotFound(t *testing.T) {
	router := setupAlarmTestRouter()
	mockAlarmService := new(MockAlarmService)

	mockAlarmService.On("GetAlarm", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	router.GET("/api/v1/alarms/:id", func(c *gin.Context) {
		id := c.Param("id")
		alarm, err := mockAlarmService.GetAlarm(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "告警不存在"})
			return
		}

		c.JSON(http.StatusOK, dto.Response{
			Code:    0,
			Message: "success",
			Data:    alarm,
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/alarms/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	mockAlarmService.AssertExpectations(t)
}
