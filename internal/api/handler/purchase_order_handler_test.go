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

type mockPurchaseOrderRepo struct {
	mock.Mock
}

func (m *mockPurchaseOrderRepo) Create(ctx context.Context, order *entity.PurchaseOrder) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *mockPurchaseOrderRepo) Update(ctx context.Context, order *entity.PurchaseOrder) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *mockPurchaseOrderRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockPurchaseOrderRepo) GetByID(ctx context.Context, id string) (*entity.PurchaseOrder, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.PurchaseOrder), args.Error(1)
}

func (m *mockPurchaseOrderRepo) List(ctx context.Context, supplierID *string, status *string, startDate, endDate *time.Time, offset, limit int) ([]*entity.PurchaseOrder, int64, error) {
	args := m.Called(ctx, supplierID, status, startDate, endDate, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.PurchaseOrder), args.Get(1).(int64), args.Error(2)
}

func (m *mockPurchaseOrderRepo) GetByCode(ctx context.Context, code string) (*entity.PurchaseOrder, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.PurchaseOrder), args.Error(1)
}

type mockSupplierRepo struct {
	mock.Mock
}

func (m *mockSupplierRepo) Create(ctx context.Context, supplier *entity.Supplier) error {
	args := m.Called(ctx, supplier)
	return args.Error(0)
}

func (m *mockSupplierRepo) Update(ctx context.Context, supplier *entity.Supplier) error {
	args := m.Called(ctx, supplier)
	return args.Error(0)
}

func (m *mockSupplierRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockSupplierRepo) GetByID(ctx context.Context, id string) (*entity.Supplier, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Supplier), args.Error(1)
}

func (m *mockSupplierRepo) GetByCode(ctx context.Context, code string) (*entity.Supplier, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Supplier), args.Error(1)
}

func (m *mockSupplierRepo) List(ctx context.Context, filter interface{}) ([]*entity.Supplier, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Supplier), args.Error(1)
}

func (m *mockSupplierRepo) Count(ctx context.Context, filter interface{}) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func setupPurchaseOrderHandler(poRepo *mockPurchaseOrderRepo, supplierRepo *mockSupplierRepo) (*PurchaseOrderHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := service.NewPurchaseOrderService(poRepo, supplierRepo)
	handler := NewPurchaseOrderHandler(svc)
	r := gin.New()
	return handler, r
}

func TestPurchaseOrderHandler_CreatePurchaseOrder_Success(t *testing.T) {
	poRepo := new(mockPurchaseOrderRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupPurchaseOrderHandler(poRepo, supplierRepo)

	r.POST("/purchase-orders", handler.CreatePurchaseOrder)

	supplierRepo.On("GetByID", mock.Anything, "supplier-001").Return(&entity.Supplier{ID: "supplier-001"}, nil)
	poRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.PurchaseOrder")).Return(nil)

	body := map[string]interface{}{
		"supplier_id": "supplier-001",
		"items": []map[string]interface{}{
			{"item_code": "IC001", "item_name": "Item1", "quantity": 10, "unit": "pcs", "unit_price": 100.0},
		},
		"notes": "test order",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/purchase-orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPurchaseOrderHandler_CreatePurchaseOrder_BadRequest(t *testing.T) {
	poRepo := new(mockPurchaseOrderRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupPurchaseOrderHandler(poRepo, supplierRepo)

	r.POST("/purchase-orders", handler.CreatePurchaseOrder)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/purchase-orders", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPurchaseOrderHandler_CreatePurchaseOrder_ServiceError(t *testing.T) {
	poRepo := new(mockPurchaseOrderRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupPurchaseOrderHandler(poRepo, supplierRepo)

	r.POST("/purchase-orders", handler.CreatePurchaseOrder)

	supplierRepo.On("GetByID", mock.Anything, "supplier-001").Return(nil, assert.AnError)

	body := map[string]interface{}{
		"supplier_id": "supplier-001",
		"items": []map[string]interface{}{
			{"item_code": "IC001", "item_name": "Item1", "quantity": 10, "unit": "pcs", "unit_price": 100.0},
		},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/purchase-orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPurchaseOrderHandler_GetPurchaseOrder_Success(t *testing.T) {
	poRepo := new(mockPurchaseOrderRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupPurchaseOrderHandler(poRepo, supplierRepo)

	r.GET("/purchase-orders/:id", handler.GetPurchaseOrder)

	order := &entity.PurchaseOrder{ID: "po-001", Code: "PO001", Status: "pending"}
	poRepo.On("GetByID", mock.Anything, "po-001").Return(order, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/purchase-orders/po-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPurchaseOrderHandler_GetPurchaseOrder_Error(t *testing.T) {
	poRepo := new(mockPurchaseOrderRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupPurchaseOrderHandler(poRepo, supplierRepo)

	r.GET("/purchase-orders/:id", handler.GetPurchaseOrder)

	poRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/purchase-orders/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPurchaseOrderHandler_DeletePurchaseOrder_Success(t *testing.T) {
	poRepo := new(mockPurchaseOrderRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupPurchaseOrderHandler(poRepo, supplierRepo)

	r.DELETE("/purchase-orders/:id", handler.DeletePurchaseOrder)

	order := &entity.PurchaseOrder{ID: "po-001", Status: "pending"}
	poRepo.On("GetByID", mock.Anything, "po-001").Return(order, nil)
	poRepo.On("Delete", mock.Anything, "po-001").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/purchase-orders/po-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPurchaseOrderHandler_DeletePurchaseOrder_Error(t *testing.T) {
	poRepo := new(mockPurchaseOrderRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupPurchaseOrderHandler(poRepo, supplierRepo)

	r.DELETE("/purchase-orders/:id", handler.DeletePurchaseOrder)

	poRepo.On("GetByID", mock.Anything, "po-001").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/purchase-orders/po-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPurchaseOrderHandler_ListPurchaseOrders_Success(t *testing.T) {
	poRepo := new(mockPurchaseOrderRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupPurchaseOrderHandler(poRepo, supplierRepo)

	r.GET("/purchase-orders", handler.ListPurchaseOrders)

	orders := []*entity.PurchaseOrder{{ID: "po-001"}}
	poRepo.On("List", mock.Anything, (*string)(nil), (*string)(nil), (*time.Time)(nil), (*time.Time)(nil), 0, 10).Return(orders, int64(1), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/purchase-orders", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPurchaseOrderHandler_ListPurchaseOrders_WithFilters(t *testing.T) {
	poRepo := new(mockPurchaseOrderRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupPurchaseOrderHandler(poRepo, supplierRepo)

	r.GET("/purchase-orders", handler.ListPurchaseOrders)

	supplierID := "supplier-001"
	status := "pending"
	poRepo.On("List", mock.Anything, &supplierID, &status, mock.Anything, mock.Anything, 0, 10).Return([]*entity.PurchaseOrder{}, int64(0), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/purchase-orders?supplier_id=supplier-001&status=pending&start_date=2024-01-01&end_date=2024-12-31", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPurchaseOrderHandler_UpdatePurchaseOrderStatus_Success(t *testing.T) {
	poRepo := new(mockPurchaseOrderRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupPurchaseOrderHandler(poRepo, supplierRepo)

	r.PUT("/purchase-orders/:id/status", handler.UpdatePurchaseOrderStatus)

	order := &entity.PurchaseOrder{ID: "po-001", Status: "pending"}
	poRepo.On("GetByID", mock.Anything, "po-001").Return(order, nil)
	poRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.PurchaseOrder")).Return(nil)

	body := map[string]interface{}{"status": "approved"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/purchase-orders/po-001/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPurchaseOrderHandler_UpdatePurchaseOrderStatus_BadRequest(t *testing.T) {
	poRepo := new(mockPurchaseOrderRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupPurchaseOrderHandler(poRepo, supplierRepo)

	r.PUT("/purchase-orders/:id/status", handler.UpdatePurchaseOrderStatus)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/purchase-orders/po-001/status", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPurchaseOrderHandler_UpdatePurchaseOrder_Success(t *testing.T) {
	poRepo := new(mockPurchaseOrderRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupPurchaseOrderHandler(poRepo, supplierRepo)

	r.PUT("/purchase-orders/:id", handler.UpdatePurchaseOrder)

	existingOrder := &entity.PurchaseOrder{ID: "po-001", SupplierID: "supplier-001", Status: "pending"}
	poRepo.On("GetByID", mock.Anything, "po-001").Return(existingOrder, nil)
	supplierRepo.On("GetByID", mock.Anything, "supplier-001").Return(&entity.Supplier{ID: "supplier-001"}, nil)
	poRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.PurchaseOrder")).Return(nil)

	body := map[string]interface{}{
		"supplier_id": "supplier-001",
		"items": []map[string]interface{}{
			{"item_code": "IC001", "item_name": "Item1", "quantity": 5, "unit": "pcs", "unit_price": 200.0},
		},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/purchase-orders/po-001", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPurchaseOrderHandler_UpdatePurchaseOrder_BadRequest(t *testing.T) {
	poRepo := new(mockPurchaseOrderRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupPurchaseOrderHandler(poRepo, supplierRepo)

	r.PUT("/purchase-orders/:id", handler.UpdatePurchaseOrder)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/purchase-orders/po-001", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNewPurchaseOrderHandler(t *testing.T) {
	poRepo := new(mockPurchaseOrderRepo)
	supplierRepo := new(mockSupplierRepo)
	svc := service.NewPurchaseOrderService(poRepo, supplierRepo)
	handler := NewPurchaseOrderHandler(svc)
	assert.NotNil(t, handler)
}

func TestPurchaseOrderHandler_ListPurchaseOrders_Error(t *testing.T) {
	poRepo := new(mockPurchaseOrderRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupPurchaseOrderHandler(poRepo, supplierRepo)

	r.GET("/purchase-orders", handler.ListPurchaseOrders)

	poRepo.On("List", mock.Anything, (*string)(nil), (*string)(nil), (*time.Time)(nil), (*time.Time)(nil), 0, 10).Return(([]*entity.PurchaseOrder)(nil), int64(0), assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/purchase-orders", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPurchaseOrderHandler_UpdatePurchaseOrderStatus_Error(t *testing.T) {
	poRepo := new(mockPurchaseOrderRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupPurchaseOrderHandler(poRepo, supplierRepo)

	r.PUT("/purchase-orders/:id/status", handler.UpdatePurchaseOrderStatus)

	poRepo.On("GetByID", mock.Anything, "po-001").Return(nil, assert.AnError)

	body := map[string]interface{}{"status": "approved"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/purchase-orders/po-001/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

var _ repository.PurchaseOrderRepository = (*mockPurchaseOrderRepo)(nil)
var _ repository.SupplierRepository = (*mockSupplierRepo)(nil)
