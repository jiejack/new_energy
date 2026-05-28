package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCostCategoryRepo struct {
	mock.Mock
}

func (m *mockCostCategoryRepo) Create(ctx context.Context, category *entity.CostCategory) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *mockCostCategoryRepo) Update(ctx context.Context, category *entity.CostCategory) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *mockCostCategoryRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockCostCategoryRepo) GetByID(ctx context.Context, id string) (*entity.CostCategory, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostCategory), args.Error(1)
}

func (m *mockCostCategoryRepo) GetByCode(ctx context.Context, code string) (*entity.CostCategory, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostCategory), args.Error(1)
}

func (m *mockCostCategoryRepo) List(ctx context.Context, parentID *string, status *string) ([]*entity.CostCategory, error) {
	args := m.Called(ctx, parentID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.CostCategory), args.Error(1)
}

func (m *mockCostCategoryRepo) GetTree(ctx context.Context) ([]*entity.CostCategory, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.CostCategory), args.Error(1)
}

type mockCostEntryRepo struct {
	mock.Mock
}

func (m *mockCostEntryRepo) Create(ctx context.Context, entry *entity.CostEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *mockCostEntryRepo) Update(ctx context.Context, entry *entity.CostEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *mockCostEntryRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockCostEntryRepo) GetByID(ctx context.Context, id string) (*entity.CostEntry, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostEntry), args.Error(1)
}

func (m *mockCostEntryRepo) GetByCode(ctx context.Context, code string) (*entity.CostEntry, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostEntry), args.Error(1)
}

func (m *mockCostEntryRepo) List(ctx context.Context, categoryID *string, startDate, endDate *time.Time, status *string, offset, limit int) ([]*entity.CostEntry, int64, error) {
	args := m.Called(ctx, categoryID, startDate, endDate, status, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.CostEntry), args.Get(1).(int64), args.Error(2)
}

func (m *mockCostEntryRepo) GetTotalByCategory(ctx context.Context, categoryID string, startDate, endDate *time.Time) (float64, error) {
	args := m.Called(ctx, categoryID, startDate, endDate)
	return args.Get(0).(float64), args.Error(1)
}

func (m *mockCostEntryRepo) GetTotalByPeriod(ctx context.Context, startDate, endDate *time.Time) (float64, error) {
	args := m.Called(ctx, startDate, endDate)
	return args.Get(0).(float64), args.Error(1)
}

type mockCostAllocationRepo struct {
	mock.Mock
}

func (m *mockCostAllocationRepo) Create(ctx context.Context, allocation *entity.CostAllocation) error {
	args := m.Called(ctx, allocation)
	return args.Error(0)
}

func (m *mockCostAllocationRepo) Update(ctx context.Context, allocation *entity.CostAllocation) error {
	args := m.Called(ctx, allocation)
	return args.Error(0)
}

func (m *mockCostAllocationRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockCostAllocationRepo) GetByID(ctx context.Context, id string) (*entity.CostAllocation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostAllocation), args.Error(1)
}

func (m *mockCostAllocationRepo) ListByCostEntryID(ctx context.Context, costEntryID string) ([]*entity.CostAllocation, error) {
	args := m.Called(ctx, costEntryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.CostAllocation), args.Error(1)
}

func (m *mockCostAllocationRepo) ListByAllocated(ctx context.Context, allocatedTo, allocatedID string) ([]*entity.CostAllocation, error) {
	args := m.Called(ctx, allocatedTo, allocatedID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.CostAllocation), args.Error(1)
}

func (m *mockCostAllocationRepo) GetTotalByAllocated(ctx context.Context, allocatedTo, allocatedID string, startDate, endDate *time.Time) (float64, error) {
	args := m.Called(ctx, allocatedTo, allocatedID, startDate, endDate)
	return args.Get(0).(float64), args.Error(1)
}

type mockCostReportRepo struct {
	mock.Mock
}

func (m *mockCostReportRepo) Create(ctx context.Context, report *entity.CostReport) error {
	args := m.Called(ctx, report)
	return args.Error(0)
}

func (m *mockCostReportRepo) Update(ctx context.Context, report *entity.CostReport) error {
	args := m.Called(ctx, report)
	return args.Error(0)
}

func (m *mockCostReportRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockCostReportRepo) GetByID(ctx context.Context, id string) (*entity.CostReport, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostReport), args.Error(1)
}

func (m *mockCostReportRepo) GetByCode(ctx context.Context, code string) (*entity.CostReport, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostReport), args.Error(1)
}

func (m *mockCostReportRepo) List(ctx context.Context, reportType *string, status *string, startDate, endDate *time.Time, offset, limit int) ([]*entity.CostReport, int64, error) {
	args := m.Called(ctx, reportType, status, startDate, endDate, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.CostReport), args.Get(1).(int64), args.Error(2)
}

func (m *mockCostReportRepo) GetByPeriod(ctx context.Context, reportType string, periodStart, periodEnd time.Time) (*entity.CostReport, error) {
	args := m.Called(ctx, reportType, periodStart, periodEnd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostReport), args.Error(1)
}

func setupCostCategoryHandler(repo *mockCostCategoryRepo) (*CostCategoryHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := service.NewCostCategoryService(repo)
	handler := NewCostCategoryHandler(svc)
	r := gin.New()
	return handler, r
}

func TestCostCategoryHandler_CreateCostCategory_Success(t *testing.T) {
	repo := new(mockCostCategoryRepo)
	handler, r := setupCostCategoryHandler(repo)

	r.POST("/cost-categories", handler.CreateCostCategory)

	repo.On("GetByCode", mock.Anything, "CC001").Return(nil, assert.AnError)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.CostCategory")).Return(nil)

	body := map[string]interface{}{
		"code": "CC001", "name": "Test Category", "type": "direct",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/cost-categories", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCostCategoryHandler_CreateCostCategory_BadRequest(t *testing.T) {
	repo := new(mockCostCategoryRepo)
	handler, r := setupCostCategoryHandler(repo)

	r.POST("/cost-categories", handler.CreateCostCategory)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/cost-categories", bytes.NewBufferString(`invalid`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCostCategoryHandler_GetCostCategoryByID_Success(t *testing.T) {
	repo := new(mockCostCategoryRepo)
	handler, r := setupCostCategoryHandler(repo)

	r.GET("/cost-categories/:id", handler.GetCostCategoryByID)

	repo.On("GetByID", mock.Anything, "cc-001").Return(&entity.CostCategory{ID: "cc-001"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-categories/cc-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostCategoryHandler_GetCostCategoryByID_Error(t *testing.T) {
	repo := new(mockCostCategoryRepo)
	handler, r := setupCostCategoryHandler(repo)

	r.GET("/cost-categories/:id", handler.GetCostCategoryByID)

	repo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-categories/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCostCategoryHandler_GetCostCategoryByCode_Success(t *testing.T) {
	repo := new(mockCostCategoryRepo)
	handler, r := setupCostCategoryHandler(repo)

	r.GET("/cost-categories/code/:code", handler.GetCostCategoryByCode)

	repo.On("GetByCode", mock.Anything, "CC001").Return(&entity.CostCategory{ID: "cc-001", Code: "CC001"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-categories/code/CC001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostCategoryHandler_GetCostCategoryByCode_Error(t *testing.T) {
	repo := new(mockCostCategoryRepo)
	handler, r := setupCostCategoryHandler(repo)

	r.GET("/cost-categories/code/:code", handler.GetCostCategoryByCode)

	repo.On("GetByCode", mock.Anything, "INVALID").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-categories/code/INVALID", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCostCategoryHandler_UpdateCostCategory_Success(t *testing.T) {
	repo := new(mockCostCategoryRepo)
	handler, r := setupCostCategoryHandler(repo)

	r.PUT("/cost-categories/:id", handler.UpdateCostCategory)

	existing := &entity.CostCategory{ID: "cc-001", Code: "CC001"}
	repo.On("GetByID", mock.Anything, "cc-001").Return(existing, nil)
	repo.On("GetByCode", mock.Anything, "CC001").Return(existing, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*entity.CostCategory")).Return(nil)

	body := map[string]interface{}{
		"code": "CC001", "name": "Updated Category", "type": "direct",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/cost-categories/cc-001", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostCategoryHandler_UpdateCostCategory_BadRequest(t *testing.T) {
	repo := new(mockCostCategoryRepo)
	handler, r := setupCostCategoryHandler(repo)

	r.PUT("/cost-categories/:id", handler.UpdateCostCategory)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/cost-categories/cc-001", bytes.NewBufferString(`invalid`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCostCategoryHandler_DeleteCostCategory_Success(t *testing.T) {
	repo := new(mockCostCategoryRepo)
	handler, r := setupCostCategoryHandler(repo)

	r.DELETE("/cost-categories/:id", handler.DeleteCostCategory)

	repo.On("GetByID", mock.Anything, "cc-001").Return(&entity.CostCategory{ID: "cc-001"}, nil)
	repo.On("List", mock.Anything, mock.AnythingOfType("*string"), (*string)(nil)).Return([]*entity.CostCategory{}, nil)
	repo.On("Delete", mock.Anything, "cc-001").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/cost-categories/cc-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostCategoryHandler_DeleteCostCategory_Error(t *testing.T) {
	repo := new(mockCostCategoryRepo)
	handler, r := setupCostCategoryHandler(repo)

	r.DELETE("/cost-categories/:id", handler.DeleteCostCategory)

	repo.On("GetByID", mock.Anything, "cc-001").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/cost-categories/cc-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCostCategoryHandler_ListCostCategories_Success(t *testing.T) {
	repo := new(mockCostCategoryRepo)
	handler, r := setupCostCategoryHandler(repo)

	r.GET("/cost-categories", handler.ListCostCategories)

	repo.On("List", mock.Anything, (*string)(nil), (*string)(nil)).Return([]*entity.CostCategory{{ID: "cc-001"}}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-categories", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostCategoryHandler_ListCostCategories_WithFilters(t *testing.T) {
	repo := new(mockCostCategoryRepo)
	handler, r := setupCostCategoryHandler(repo)

	r.GET("/cost-categories", handler.ListCostCategories)

	parentID := "parent-001"
	status := "active"
	repo.On("List", mock.Anything, &parentID, &status).Return([]*entity.CostCategory{{ID: "cc-001"}}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-categories?parent_id=parent-001&status=active", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostCategoryHandler_GetCostCategoryTree_Success(t *testing.T) {
	repo := new(mockCostCategoryRepo)
	handler, r := setupCostCategoryHandler(repo)

	r.GET("/cost-categories/tree", handler.GetCostCategoryTree)

	repo.On("GetTree", mock.Anything).Return([]*entity.CostCategory{{ID: "cc-001"}}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-categories/tree", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostCategoryHandler_GetCostCategoryTree_Error(t *testing.T) {
	repo := new(mockCostCategoryRepo)
	handler, r := setupCostCategoryHandler(repo)

	r.GET("/cost-categories/tree", handler.GetCostCategoryTree)

	repo.On("GetTree", mock.Anything).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-categories/tree", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestNewCostCategoryHandler(t *testing.T) {
	repo := new(mockCostCategoryRepo)
	svc := service.NewCostCategoryService(repo)
	handler := NewCostCategoryHandler(svc)
	assert.NotNil(t, handler)
}

var _ repository.CostCategoryRepository = (*mockCostCategoryRepo)(nil)
var _ repository.CostEntryRepository = (*mockCostEntryRepo)(nil)
var _ repository.CostAllocationRepository = (*mockCostAllocationRepo)(nil)
var _ repository.CostReportRepository = (*mockCostReportRepo)(nil)
