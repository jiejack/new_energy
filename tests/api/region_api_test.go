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

// MockRegionService 区域服务Mock
type MockRegionService struct {
	mock.Mock
}

func (m *MockRegionService) CreateRegion(ctx interface{}, req interface{}) (*entity.Region, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Region), args.Error(1)
}

func (m *MockRegionService) UpdateRegion(ctx interface{}, id string, req interface{}) (*entity.Region, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Region), args.Error(1)
}

func (m *MockRegionService) DeleteRegion(ctx interface{}, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRegionService) GetRegion(ctx interface{}, id string) (*entity.Region, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Region), args.Error(1)
}

func (m *MockRegionService) ListRegions(ctx interface{}, parentID *string) ([]*entity.Region, error) {
	args := m.Called(ctx, parentID)
	return args.Get(0).([]*entity.Region), args.Error(1)
}

func (m *MockRegionService) GetRegionTree(ctx interface{}) ([]*entity.Region, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entity.Region), args.Error(1)
}

func setupRegionTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestRegionAPI_CreateRegion_Success(t *testing.T) {
	router := setupRegionTestRouter()
	mockRegionService := new(MockRegionService)

	expectedRegion := &entity.Region{
		ID:          "region-001",
		Code:        "EAST",
		Name:        "华东区域",
		Level:       1,
		Description: "华东地区",
	}

	mockRegionService.On("CreateRegion", mock.Anything, mock.Anything).Return(expectedRegion, nil)

	router.POST("/api/v1/regions", func(c *gin.Context) {
		var req dto.CreateRegionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "参数错误"})
			return
		}

		region, err := mockRegionService.CreateRegion(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: "创建失败"})
			return
		}

		c.JSON(http.StatusCreated, dto.Response{
			Code:    0,
			Message: "success",
			Data: dto.RegionResponse{
				ID:          region.ID,
				Code:        region.Code,
				Name:        region.Name,
				Level:       region.Level,
				Description: region.Description,
			},
		})
	})

	createReq := dto.CreateRegionRequest{
		Code:        "EAST",
		Name:        "华东区域",
		Level:       1,
		Description: "华东地区",
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/regions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockRegionService.AssertExpectations(t)
}

func TestRegionAPI_GetRegion_Success(t *testing.T) {
	router := setupRegionTestRouter()
	mockRegionService := new(MockRegionService)

	expectedRegion := &entity.Region{
		ID:          "region-001",
		Code:        "EAST",
		Name:        "华东区域",
		Level:       1,
		Description: "华东地区",
	}

	mockRegionService.On("GetRegion", mock.Anything, "region-001").Return(expectedRegion, nil)

	router.GET("/api/v1/regions/:id", func(c *gin.Context) {
		id := c.Param("id")
		region, err := mockRegionService.GetRegion(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "区域不存在"})
			return
		}

		c.JSON(http.StatusOK, dto.Response{
			Code:    0,
			Message: "success",
			Data: dto.RegionResponse{
				ID:          region.ID,
				Code:        region.Code,
				Name:        region.Name,
				Level:       region.Level,
				Description: region.Description,
			},
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/regions/region-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockRegionService.AssertExpectations(t)
}

func TestRegionAPI_GetRegion_NotFound(t *testing.T) {
	router := setupRegionTestRouter()
	mockRegionService := new(MockRegionService)

	mockRegionService.On("GetRegion", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	router.GET("/api/v1/regions/:id", func(c *gin.Context) {
		id := c.Param("id")
		region, err := mockRegionService.GetRegion(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "区域不存在"})
			return
		}

		c.JSON(http.StatusOK, dto.Response{
			Code:    0,
			Message: "success",
			Data:    region,
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/regions/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	mockRegionService.AssertExpectations(t)
}

func TestRegionAPI_ListRegions_Success(t *testing.T) {
	router := setupRegionTestRouter()
	mockRegionService := new(MockRegionService)

	expectedRegions := []*entity.Region{
		{ID: "region-001", Code: "EAST", Name: "华东区域", Level: 1},
		{ID: "region-002", Code: "NORTH", Name: "华北区域", Level: 1},
		{ID: "region-003", Code: "SOUTH", Name: "华南区域", Level: 1},
	}

	mockRegionService.On("ListRegions", mock.Anything, (*string)(nil)).Return(expectedRegions, nil)

	router.GET("/api/v1/regions", func(c *gin.Context) {
		var parentID *string
		if pid := c.Query("parent_id"); pid != "" {
			parentID = &pid
		}

		regions, err := mockRegionService.ListRegions(c.Request.Context(), parentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: "查询失败"})
			return
		}

		resp := make([]dto.RegionResponse, len(regions))
		for i, r := range regions {
			resp[i] = dto.RegionResponse{
				ID:    r.ID,
				Code:  r.Code,
				Name:  r.Name,
				Level: r.Level,
			}
		}

		c.JSON(http.StatusOK, dto.Response{
			Code:    0,
			Message: "success",
			Data:    resp,
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/regions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockRegionService.AssertExpectations(t)
}

func TestRegionAPI_UpdateRegion_Success(t *testing.T) {
	router := setupRegionTestRouter()
	mockRegionService := new(MockRegionService)

	expectedRegion := &entity.Region{
		ID:          "region-001",
		Code:        "EAST",
		Name:        "华东区域(已更新)",
		Level:       1,
		Description: "更新后的描述",
	}

	mockRegionService.On("UpdateRegion", mock.Anything, "region-001", mock.Anything).Return(expectedRegion, nil)

	router.PUT("/api/v1/regions/:id", func(c *gin.Context) {
		id := c.Param("id")
		var req dto.UpdateRegionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "参数错误"})
			return
		}

		region, err := mockRegionService.UpdateRegion(c.Request.Context(), id, &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: "更新失败"})
			return
		}

		c.JSON(http.StatusOK, dto.Response{
			Code:    0,
			Message: "success",
			Data: dto.RegionResponse{
				ID:          region.ID,
				Code:        region.Code,
				Name:        region.Name,
				Description: region.Description,
			},
		})
	})

	updateReq := dto.UpdateRegionRequest{
		Name:        "华东区域(已更新)",
		Description: "更新后的描述",
	}
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/regions/region-001", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockRegionService.AssertExpectations(t)
}

func TestRegionAPI_DeleteRegion_Success(t *testing.T) {
	router := setupRegionTestRouter()
	mockRegionService := new(MockRegionService)

	mockRegionService.On("DeleteRegion", mock.Anything, "region-001").Return(nil)

	router.DELETE("/api/v1/regions/:id", func(c *gin.Context) {
		id := c.Param("id")
		err := mockRegionService.DeleteRegion(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: "删除失败"})
			return
		}

		c.JSON(http.StatusNoContent, nil)
	})

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/regions/region-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	mockRegionService.AssertExpectations(t)
}

func TestRegionAPI_GetRegionTree_Success(t *testing.T) {
	router := setupRegionTestRouter()
	mockRegionService := new(MockRegionService)

	childRegion := &entity.Region{
		ID:       "region-002",
		Code:     "SH",
		Name:     "上海",
		Level:    2,
		ParentID: strPtr("region-001"),
	}

	expectedRegions := []*entity.Region{
		{
			ID:        "region-001",
			Code:      "EAST",
			Name:      "华东区域",
			Level:     1,
			SubRegions: []*entity.Region{childRegion},
		},
	}

	mockRegionService.On("GetRegionTree", mock.Anything).Return(expectedRegions, nil)

	router.GET("/api/v1/regions/tree", func(c *gin.Context) {
		regions, err := mockRegionService.GetRegionTree(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: "查询失败"})
			return
		}

		c.JSON(http.StatusOK, dto.Response{
			Code:    0,
			Message: "success",
			Data:    regions,
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/regions/tree", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockRegionService.AssertExpectations(t)
}

func TestRegionAPI_CreateRegion_ValidationError(t *testing.T) {
	router := setupRegionTestRouter()

	router.POST("/api/v1/regions", func(c *gin.Context) {
		var req dto.CreateRegionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "参数验证失败"})
			return
		}

		// 验证必填字段
		if req.Code == "" || req.Name == "" {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "编码和名称为必填项"})
			return
		}

		c.JSON(http.StatusCreated, dto.Response{Code: 0, Message: "success"})
	})

	// 测试缺少必填字段
	createReq := dto.CreateRegionRequest{
		Description: "只有描述",
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/regions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func strPtr(s string) *string {
	return &s
}
