package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/api/dto"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStationService 电站服务Mock
type MockStationService struct {
	mock.Mock
}

func (m *MockStationService) CreateStation(ctx interface{}, req interface{}) (*entity.Station, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Station), args.Error(1)
}

func (m *MockStationService) UpdateStation(ctx interface{}, id string, req interface{}) (*entity.Station, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Station), args.Error(1)
}

func (m *MockStationService) DeleteStation(ctx interface{}, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStationService) GetStation(ctx interface{}, id string) (*entity.Station, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Station), args.Error(1)
}

func (m *MockStationService) GetStationWithDevices(ctx interface{}, id string) (*entity.Station, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Station), args.Error(1)
}

func (m *MockStationService) ListStations(ctx interface{}, subRegionID *string, stationType *entity.StationType) ([]*entity.Station, error) {
	args := m.Called(ctx, subRegionID, stationType)
	return args.Get(0).([]*entity.Station), args.Error(1)
}

func (m *MockStationService) GetStationPoints(ctx interface{}, id string) ([]*entity.Point, error) {
	args := m.Called(ctx, id)
	return args.Get(0).([]*entity.Point), args.Error(1)
}

func setupStationTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestStationAPI_CreateStation_Success(t *testing.T) {
	router := setupStationTestRouter()
	mockStationService := new(MockStationService)

	expectedStation := &entity.Station{
		ID:          "station-001",
		Code:        "PV_001",
		Name:        "测试光伏电站",
		Type:        entity.StationTypePV,
		SubRegionID: "region-001",
		Capacity:    100.0,
		Status:      entity.StationStatusActive,
	}

	mockStationService.On("CreateStation", mock.Anything, mock.Anything).Return(expectedStation, nil)

	router.POST("/api/v1/stations", func(c *gin.Context) {
		var req dto.CreateStationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "参数错误"})
			return
		}

		station, err := mockStationService.CreateStation(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: "创建失败"})
			return
		}

		c.JSON(http.StatusCreated, dto.Response{
			Code:    0,
			Message: "success",
			Data: dto.StationResponse{
				ID:          station.ID,
				Code:        station.Code,
				Name:        station.Name,
				Type:        string(station.Type),
				SubRegionID: station.SubRegionID,
				Capacity:    station.Capacity,
				Status:      int(station.Status),
			},
		})
	})

	createReq := dto.CreateStationRequest{
		Code:        "PV_001",
		Name:        "测试光伏电站",
		Type:        "pv",
		SubRegionID: "region-001",
		Capacity:    100.0,
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/stations", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockStationService.AssertExpectations(t)
}

func TestStationAPI_GetStation_Success(t *testing.T) {
	router := setupStationTestRouter()
	mockStationService := new(MockStationService)

	expectedStation := &entity.Station{
		ID:          "station-001",
		Code:        "PV_001",
		Name:        "测试光伏电站",
		Type:        entity.StationTypePV,
		SubRegionID: "region-001",
		Capacity:    100.0,
		Status:      entity.StationStatusActive,
	}

	mockStationService.On("GetStation", mock.Anything, "station-001").Return(expectedStation, nil)

	router.GET("/api/v1/stations/:id", func(c *gin.Context) {
		id := c.Param("id")
		station, err := mockStationService.GetStation(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "电站不存在"})
			return
		}

		c.JSON(http.StatusOK, dto.Response{
			Code:    0,
			Message: "success",
			Data: dto.StationResponse{
				ID:          station.ID,
				Code:        station.Code,
				Name:        station.Name,
				Type:        string(station.Type),
				SubRegionID: station.SubRegionID,
				Capacity:    station.Capacity,
				Status:      int(station.Status),
			},
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/stations/station-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockStationService.AssertExpectations(t)
}

func TestStationAPI_ListStations_Success(t *testing.T) {
	router := setupStationTestRouter()
	mockStationService := new(MockStationService)

	expectedStations := []*entity.Station{
		{ID: "station-001", Code: "PV_001", Name: "光伏电站1", Type: entity.StationTypePV, Status: entity.StationStatusActive},
		{ID: "station-002", Code: "WIND_001", Name: "风电场1", Type: entity.StationTypeWind, Status: entity.StationStatusActive},
	}

	mockStationService.On("ListStations", mock.Anything, (*string)(nil), (*entity.StationType)(nil)).Return(expectedStations, nil)

	router.GET("/api/v1/stations", func(c *gin.Context) {
		var subRegionID *string
		var stationType *entity.StationType

		if srid := c.Query("sub_region_id"); srid != "" {
			subRegionID = &srid
		}
		if st := c.Query("type"); st != "" {
			t := entity.StationType(st)
			stationType = &t
		}

		stations, err := mockStationService.ListStations(c.Request.Context(), subRegionID, stationType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: "查询失败"})
			return
		}

		resp := make([]dto.StationResponse, len(stations))
		for i, s := range stations {
			resp[i] = dto.StationResponse{
				ID:     s.ID,
				Code:   s.Code,
				Name:   s.Name,
				Type:   string(s.Type),
				Status: int(s.Status),
			}
		}

		c.JSON(http.StatusOK, dto.PagedResponse{
			Code:      0,
			Message:   "success",
			Data:      resp,
			Total:     int64(len(resp)),
			Page:      1,
			PageSize:  20,
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/stations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.PagedResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, int64(2), resp.Total)

	mockStationService.AssertExpectations(t)
}

func TestStationAPI_UpdateStation_Success(t *testing.T) {
	router := setupStationTestRouter()
	mockStationService := new(MockStationService)

	expectedStation := &entity.Station{
		ID:          "station-001",
		Code:        "PV_001",
		Name:        "更新后的电站名称",
		Type:        entity.StationTypePV,
		SubRegionID: "region-001",
		Capacity:    200.0,
	}

	mockStationService.On("UpdateStation", mock.Anything, "station-001", mock.Anything).Return(expectedStation, nil)

	router.PUT("/api/v1/stations/:id", func(c *gin.Context) {
		id := c.Param("id")
		var req dto.UpdateStationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "参数错误"})
			return
		}

		station, err := mockStationService.UpdateStation(c.Request.Context(), id, &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: "更新失败"})
			return
		}

		c.JSON(http.StatusOK, dto.Response{
			Code:    0,
			Message: "success",
			Data: dto.StationResponse{
				ID:       station.ID,
				Name:     station.Name,
				Capacity: station.Capacity,
			},
		})
	})

	updateReq := dto.UpdateStationRequest{
		Name:     "更新后的电站名称",
		Capacity: 200.0,
	}
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/stations/station-001", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockStationService.AssertExpectations(t)
}

func TestStationAPI_DeleteStation_Success(t *testing.T) {
	router := setupStationTestRouter()
	mockStationService := new(MockStationService)

	mockStationService.On("DeleteStation", mock.Anything, "station-001").Return(nil)

	router.DELETE("/api/v1/stations/:id", func(c *gin.Context) {
		id := c.Param("id")
		err := mockStationService.DeleteStation(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: "删除失败"})
			return
		}

		c.JSON(http.StatusNoContent, nil)
	})

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/stations/station-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	mockStationService.AssertExpectations(t)
}

func TestStationAPI_GetStationWithDevices_Success(t *testing.T) {
	router := setupStationTestRouter()
	mockStationService := new(MockStationService)

	expectedStation := &entity.Station{
		ID:     "station-001",
		Code:   "PV_001",
		Name:   "测试光伏电站",
		Type:   entity.StationTypePV,
		Status: entity.StationStatusActive,
		Devices: []*entity.Device{
			{ID: "device-001", Code: "INV_001", Name: "1号逆变器", Type: entity.DeviceTypeInverter, Status: entity.DeviceStatusOnline},
			{ID: "device-002", Code: "METER_001", Name: "电表", Type: entity.DeviceTypeMeter, Status: entity.DeviceStatusOnline},
		},
	}

	mockStationService.On("GetStationWithDevices", mock.Anything, "station-001").Return(expectedStation, nil)

	router.GET("/api/v1/stations/:id/devices", func(c *gin.Context) {
		id := c.Param("id")
		station, err := mockStationService.GetStationWithDevices(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "电站不存在"})
			return
		}

		devices := make([]dto.DeviceBrief, len(station.Devices))
		for i, d := range station.Devices {
			devices[i] = dto.DeviceBrief{
				ID:     d.ID,
				Code:   d.Code,
				Name:   d.Name,
				Type:   string(d.Type),
				Status: int(d.Status),
			}
		}

		c.JSON(http.StatusOK, dto.Response{
			Code:    0,
			Message: "success",
			Data: dto.StationResponse{
				ID:      station.ID,
				Code:    station.Code,
				Name:    station.Name,
				Devices: devices,
			},
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/stations/station-001/devices", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockStationService.AssertExpectations(t)
}

func TestStationAPI_ListStations_WithFilter(t *testing.T) {
	router := setupStationTestRouter()
	mockStationService := new(MockStationService)

	stationType := entity.StationTypePV
	expectedStations := []*entity.Station{
		{ID: "station-001", Code: "PV_001", Name: "光伏电站1", Type: entity.StationTypePV},
		{ID: "station-002", Code: "PV_002", Name: "光伏电站2", Type: entity.StationTypePV},
	}

	mockStationService.On("ListStations", mock.Anything, mock.Anything, &stationType).Return(expectedStations, nil)

	router.GET("/api/v1/stations", func(c *gin.Context) {
		var subRegionID *string
		var stationType *entity.StationType

		if srid := c.Query("sub_region_id"); srid != "" {
			subRegionID = &srid
		}
		if st := c.Query("type"); st != "" {
			t := entity.StationType(st)
			stationType = &t
		}

		stations, err := mockStationService.ListStations(c.Request.Context(), subRegionID, stationType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: "查询失败"})
			return
		}

		resp := make([]dto.StationResponse, len(stations))
		for i, s := range stations {
			resp[i] = dto.StationResponse{
				ID:   s.ID,
				Code: s.Code,
				Name: s.Name,
				Type: string(s.Type),
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

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/stations?type=pv", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockStationService.AssertExpectations(t)
}
