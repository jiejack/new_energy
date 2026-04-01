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

// MockDeviceService 设备服务Mock
type MockDeviceService struct {
	mock.Mock
}

func (m *MockDeviceService) CreateDevice(ctx interface{}, req interface{}) (*entity.Device, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}

func (m *MockDeviceService) UpdateDevice(ctx interface{}, id string, req interface{}) (*entity.Device, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}

func (m *MockDeviceService) DeleteDevice(ctx interface{}, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDeviceService) GetDevice(ctx interface{}, id string) (*entity.Device, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}

func (m *MockDeviceService) GetDeviceWithPoints(ctx interface{}, id string) (*entity.Device, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}

func (m *MockDeviceService) ListDevices(ctx interface{}, stationID *string, deviceType *entity.DeviceType) ([]*entity.Device, error) {
	args := m.Called(ctx, stationID, deviceType)
	return args.Get(0).([]*entity.Device), args.Error(1)
}

func (m *MockDeviceService) GetOnlineDevices(ctx interface{}, stationID string) ([]*entity.Device, error) {
	args := m.Called(ctx, stationID)
	return args.Get(0).([]*entity.Device), args.Error(1)
}

func (m *MockDeviceService) SetDeviceOnline(ctx interface{}, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDeviceService) SetDeviceOffline(ctx interface{}, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupDeviceTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestDeviceAPI_CreateDevice_Success(t *testing.T) {
	router := setupDeviceTestRouter()
	mockDeviceService := new(MockDeviceService)

	expectedDevice := &entity.Device{
		ID:           "device-001",
		Code:         "INV_001",
		Name:         "1号逆变器",
		Type:         entity.DeviceTypeInverter,
		StationID:    "station-001",
		Manufacturer: "华为",
		Model:        "SUN2000-100KTL",
		Status:       entity.DeviceStatusOffline,
	}

	mockDeviceService.On("CreateDevice", mock.Anything, mock.Anything).Return(expectedDevice, nil)

	router.POST("/api/v1/devices", func(c *gin.Context) {
		var req dto.CreateDeviceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "参数错误"})
			return
		}

		device, err := mockDeviceService.CreateDevice(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: "创建失败"})
			return
		}

		c.JSON(http.StatusCreated, dto.Response{
			Code:    0,
			Message: "success",
			Data: dto.DeviceResponse{
				ID:           device.ID,
				Code:         device.Code,
				Name:         device.Name,
				Type:         string(device.Type),
				StationID:    device.StationID,
				Manufacturer: device.Manufacturer,
				Model:        device.Model,
				Status:       int(device.Status),
			},
		})
	})

	createReq := dto.CreateDeviceRequest{
		Code:         "INV_001",
		Name:         "1号逆变器",
		Type:         "inverter",
		StationID:    "station-001",
		Manufacturer: "华为",
		Model:        "SUN2000-100KTL",
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/devices", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockDeviceService.AssertExpectations(t)
}

func TestDeviceAPI_GetDevice_Success(t *testing.T) {
	router := setupDeviceTestRouter()
	mockDeviceService := new(MockDeviceService)

	expectedDevice := &entity.Device{
		ID:           "device-001",
		Code:         "INV_001",
		Name:         "1号逆变器",
		Type:         entity.DeviceTypeInverter,
		StationID:    "station-001",
		Manufacturer: "华为",
		Status:       entity.DeviceStatusOnline,
	}

	mockDeviceService.On("GetDevice", mock.Anything, "device-001").Return(expectedDevice, nil)

	router.GET("/api/v1/devices/:id", func(c *gin.Context) {
		id := c.Param("id")
		device, err := mockDeviceService.GetDevice(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "设备不存在"})
			return
		}

		c.JSON(http.StatusOK, dto.Response{
			Code:    0,
			Message: "success",
			Data: dto.DeviceResponse{
				ID:        device.ID,
				Code:      device.Code,
				Name:      device.Name,
				Type:      string(device.Type),
				StationID: device.StationID,
				Status:    int(device.Status),
			},
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/devices/device-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockDeviceService.AssertExpectations(t)
}

func TestDeviceAPI_ListDevices_Success(t *testing.T) {
	router := setupDeviceTestRouter()
	mockDeviceService := new(MockDeviceService)

	expectedDevices := []*entity.Device{
		{ID: "device-001", Code: "INV_001", Name: "1号逆变器", Type: entity.DeviceTypeInverter, Status: entity.DeviceStatusOnline},
		{ID: "device-002", Code: "INV_002", Name: "2号逆变器", Type: entity.DeviceTypeInverter, Status: entity.DeviceStatusOnline},
		{ID: "device-003", Code: "METER_001", Name: "电表", Type: entity.DeviceTypeMeter, Status: entity.DeviceStatusOnline},
	}

	mockDeviceService.On("ListDevices", mock.Anything, (*string)(nil), (*entity.DeviceType)(nil)).Return(expectedDevices, nil)

	router.GET("/api/v1/devices", func(c *gin.Context) {
		var stationID *string
		var deviceType *entity.DeviceType

		if sid := c.Query("station_id"); sid != "" {
			stationID = &sid
		}
		if dt := c.Query("type"); dt != "" {
			t := entity.DeviceType(dt)
			deviceType = &t
		}

		devices, err := mockDeviceService.ListDevices(c.Request.Context(), stationID, deviceType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: "查询失败"})
			return
		}

		resp := make([]dto.DeviceResponse, len(devices))
		for i, d := range devices {
			resp[i] = dto.DeviceResponse{
				ID:     d.ID,
				Code:   d.Code,
				Name:   d.Name,
				Type:   string(d.Type),
				Status: int(d.Status),
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

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/devices", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.PagedResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, int64(3), resp.Total)

	mockDeviceService.AssertExpectations(t)
}

func TestDeviceAPI_UpdateDevice_Success(t *testing.T) {
	router := setupDeviceTestRouter()
	mockDeviceService := new(MockDeviceService)

	expectedDevice := &entity.Device{
		ID:           "device-001",
		Code:         "INV_001",
		Name:         "更新后的逆变器",
		Type:         entity.DeviceTypeInverter,
		StationID:    "station-001",
		Manufacturer: "华为",
		Model:        "SUN2000-200KTL",
	}

	mockDeviceService.On("UpdateDevice", mock.Anything, "device-001", mock.Anything).Return(expectedDevice, nil)

	router.PUT("/api/v1/devices/:id", func(c *gin.Context) {
		id := c.Param("id")
		var req dto.UpdateDeviceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "参数错误"})
			return
		}

		device, err := mockDeviceService.UpdateDevice(c.Request.Context(), id, &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: "更新失败"})
			return
		}

		c.JSON(http.StatusOK, dto.Response{
			Code:    0,
			Message: "success",
			Data: dto.DeviceResponse{
				ID:   device.ID,
				Name: device.Name,
				Model: device.Model,
			},
		})
	})

	updateReq := dto.UpdateDeviceRequest{
		Name:  "更新后的逆变器",
		Model: "SUN2000-200KTL",
	}
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/devices/device-001", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockDeviceService.AssertExpectations(t)
}

func TestDeviceAPI_DeleteDevice_Success(t *testing.T) {
	router := setupDeviceTestRouter()
	mockDeviceService := new(MockDeviceService)

	mockDeviceService.On("DeleteDevice", mock.Anything, "device-001").Return(nil)

	router.DELETE("/api/v1/devices/:id", func(c *gin.Context) {
		id := c.Param("id")
		err := mockDeviceService.DeleteDevice(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: "删除失败"})
			return
		}

		c.JSON(http.StatusNoContent, nil)
	})

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/devices/device-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	mockDeviceService.AssertExpectations(t)
}

func TestDeviceAPI_GetOnlineDevices_Success(t *testing.T) {
	router := setupDeviceTestRouter()
	mockDeviceService := new(MockDeviceService)

	expectedDevices := []*entity.Device{
		{ID: "device-001", Code: "INV_001", Name: "1号逆变器", Type: entity.DeviceTypeInverter, Status: entity.DeviceStatusOnline},
		{ID: "device-002", Code: "INV_002", Name: "2号逆变器", Type: entity.DeviceTypeInverter, Status: entity.DeviceStatusOnline},
	}

	mockDeviceService.On("GetOnlineDevices", mock.Anything, "station-001").Return(expectedDevices, nil)

	router.GET("/api/v1/devices/online", func(c *gin.Context) {
		stationID := c.Query("station_id")
		if stationID == "" {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "缺少电站ID"})
			return
		}

		devices, err := mockDeviceService.GetOnlineDevices(c.Request.Context(), stationID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: "查询失败"})
			return
		}

		resp := make([]dto.DeviceResponse, len(devices))
		for i, d := range devices {
			resp[i] = dto.DeviceResponse{
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
			Data:    resp,
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/devices/online?station_id=station-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockDeviceService.AssertExpectations(t)
}

func TestDeviceAPI_GetDeviceWithPoints_Success(t *testing.T) {
	router := setupDeviceTestRouter()
	mockDeviceService := new(MockDeviceService)

	expectedDevice := &entity.Device{
		ID:     "device-001",
		Code:   "INV_001",
		Name:   "1号逆变器",
		Type:   entity.DeviceTypeInverter,
		Status: entity.DeviceStatusOnline,
		Points: []*entity.Point{
			{ID: "point-001", Code: "P", Name: "有功功率", Type: entity.PointTypeYaoCe},
			{ID: "point-002", Code: "U", Name: "电压", Type: entity.PointTypeYaoCe},
		},
	}

	mockDeviceService.On("GetDeviceWithPoints", mock.Anything, "device-001").Return(expectedDevice, nil)

	router.GET("/api/v1/devices/:id/points", func(c *gin.Context) {
		id := c.Param("id")
		device, err := mockDeviceService.GetDeviceWithPoints(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "设备不存在"})
			return
		}

		points := make([]dto.PointBrief, len(device.Points))
		for i, p := range device.Points {
			points[i] = dto.PointBrief{
				ID:   p.ID,
				Code: p.Code,
				Name: p.Name,
				Type: string(p.Type),
			}
		}

		c.JSON(http.StatusOK, dto.Response{
			Code:    0,
			Message: "success",
			Data: dto.DeviceResponse{
				ID:     device.ID,
				Code:   device.Code,
				Name:   device.Name,
				Points: points,
			},
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/devices/device-001/points", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockDeviceService.AssertExpectations(t)
}

func TestDeviceAPI_ListDevices_WithFilter(t *testing.T) {
	router := setupDeviceTestRouter()
	mockDeviceService := new(MockDeviceService)

	deviceType := entity.DeviceTypeInverter
	stationID := "station-001"
	expectedDevices := []*entity.Device{
		{ID: "device-001", Code: "INV_001", Name: "1号逆变器", Type: entity.DeviceTypeInverter, StationID: "station-001"},
		{ID: "device-002", Code: "INV_002", Name: "2号逆变器", Type: entity.DeviceTypeInverter, StationID: "station-001"},
	}

	mockDeviceService.On("ListDevices", mock.Anything, &stationID, &deviceType).Return(expectedDevices, nil)

	router.GET("/api/v1/devices", func(c *gin.Context) {
		var stationID *string
		var deviceType *entity.DeviceType

		if sid := c.Query("station_id"); sid != "" {
			stationID = &sid
		}
		if dt := c.Query("type"); dt != "" {
			t := entity.DeviceType(dt)
			deviceType = &t
		}

		devices, err := mockDeviceService.ListDevices(c.Request.Context(), stationID, deviceType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: "查询失败"})
			return
		}

		resp := make([]dto.DeviceResponse, len(devices))
		for i, d := range devices {
			resp[i] = dto.DeviceResponse{
				ID:        d.ID,
				Code:      d.Code,
				Name:      d.Name,
				Type:      string(d.Type),
				StationID: d.StationID,
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

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/devices?station_id=station-001&type=inverter", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockDeviceService.AssertExpectations(t)
}
