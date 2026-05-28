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

type mockReceiptRepo struct {
	mock.Mock
}

func (m *mockReceiptRepo) Create(ctx context.Context, receipt *entity.Receipt) error {
	args := m.Called(ctx, receipt)
	return args.Error(0)
}

func (m *mockReceiptRepo) Update(ctx context.Context, receipt *entity.Receipt) error {
	args := m.Called(ctx, receipt)
	return args.Error(0)
}

func (m *mockReceiptRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockReceiptRepo) GetByID(ctx context.Context, id string) (*entity.Receipt, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Receipt), args.Error(1)
}

func (m *mockReceiptRepo) List(ctx context.Context, purchaseOrderID *string, status *string, startDate, endDate *time.Time, offset, limit int) ([]*entity.Receipt, int64, error) {
	args := m.Called(ctx, purchaseOrderID, status, startDate, endDate, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.Receipt), args.Get(1).(int64), args.Error(2)
}

func (m *mockReceiptRepo) GetByCode(ctx context.Context, code string) (*entity.Receipt, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Receipt), args.Error(1)
}

func (m *mockReceiptRepo) GetByPurchaseOrderID(ctx context.Context, purchaseOrderID string) ([]*entity.Receipt, error) {
	args := m.Called(ctx, purchaseOrderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Receipt), args.Error(1)
}

type mockInventoryRepo struct {
	mock.Mock
}

func (m *mockInventoryRepo) Create(ctx context.Context, inventory *entity.Inventory) error {
	args := m.Called(ctx, inventory)
	return args.Error(0)
}

func (m *mockInventoryRepo) Update(ctx context.Context, inventory *entity.Inventory) error {
	args := m.Called(ctx, inventory)
	return args.Error(0)
}

func (m *mockInventoryRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockInventoryRepo) GetByID(ctx context.Context, id string) (*entity.Inventory, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Inventory), args.Error(1)
}

func (m *mockInventoryRepo) GetByCode(ctx context.Context, code string) (*entity.Inventory, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Inventory), args.Error(1)
}

func (m *mockInventoryRepo) List(ctx context.Context, filter interface{}) ([]*entity.Inventory, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Inventory), args.Error(1)
}

func (m *mockInventoryRepo) Count(ctx context.Context, filter interface{}) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockInventoryRepo) UpdateQuantity(ctx context.Context, id string, quantity float64) error {
	args := m.Called(ctx, id, quantity)
	return args.Error(0)
}

func (m *mockInventoryRepo) GetLowStockItems(ctx context.Context) ([]*entity.Inventory, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Inventory), args.Error(1)
}

type mockInventoryTransRepo struct {
	mock.Mock
}

func (m *mockInventoryTransRepo) Create(ctx context.Context, transaction *entity.InventoryTransaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *mockInventoryTransRepo) GetByID(ctx context.Context, id string) (*entity.InventoryTransaction, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.InventoryTransaction), args.Error(1)
}

func (m *mockInventoryTransRepo) ListByInventoryID(ctx context.Context, inventoryID string) ([]*entity.InventoryTransaction, error) {
	args := m.Called(ctx, inventoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.InventoryTransaction), args.Error(1)
}

func (m *mockInventoryTransRepo) ListByReference(ctx context.Context, referenceID string, referenceType string) ([]*entity.InventoryTransaction, error) {
	args := m.Called(ctx, referenceID, referenceType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.InventoryTransaction), args.Error(1)
}

func (m *mockInventoryTransRepo) GetTransactionHistory(ctx context.Context, inventoryID string, limit int) ([]*entity.InventoryTransaction, error) {
	args := m.Called(ctx, inventoryID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.InventoryTransaction), args.Error(1)
}

func setupReceiptHandler(receiptRepo *mockReceiptRepo, poRepo *mockPurchaseOrderRepo, invRepo *mockInventoryRepo, invTransRepo *mockInventoryTransRepo) (*ReceiptHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := service.NewReceiptService(receiptRepo, poRepo, invRepo, invTransRepo)
	handler := NewReceiptHandler(svc)
	r := gin.New()
	return handler, r
}

func TestReceiptHandler_CreateReceipt_Success(t *testing.T) {
	receiptRepo := new(mockReceiptRepo)
	poRepo := new(mockPurchaseOrderRepo)
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	handler, r := setupReceiptHandler(receiptRepo, poRepo, invRepo, invTransRepo)

	r.POST("/receipts", handler.CreateReceipt)

	poRepo.On("GetByID", mock.Anything, "po-001").Return(&entity.PurchaseOrder{ID: "po-001"}, nil)
	receiptRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Receipt")).Return(nil)

	body := map[string]interface{}{
		"purchase_order_id": "po-001",
		"items": []map[string]interface{}{
			{"purchase_order_item_id": "poi-001", "inventory_id": "inv-001", "quantity": 10, "unit": "pcs", "unit_price": 100.0},
		},
		"notes": "test receipt",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/receipts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReceiptHandler_CreateReceipt_BadRequest(t *testing.T) {
	receiptRepo := new(mockReceiptRepo)
	poRepo := new(mockPurchaseOrderRepo)
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	handler, r := setupReceiptHandler(receiptRepo, poRepo, invRepo, invTransRepo)

	r.POST("/receipts", handler.CreateReceipt)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/receipts", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReceiptHandler_GetReceipt_Success(t *testing.T) {
	receiptRepo := new(mockReceiptRepo)
	poRepo := new(mockPurchaseOrderRepo)
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	handler, r := setupReceiptHandler(receiptRepo, poRepo, invRepo, invTransRepo)

	r.GET("/receipts/:id", handler.GetReceipt)

	receipt := &entity.Receipt{ID: "rc-001", Code: "RC001"}
	receiptRepo.On("GetByID", mock.Anything, "rc-001").Return(receipt, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/receipts/rc-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReceiptHandler_GetReceipt_Error(t *testing.T) {
	receiptRepo := new(mockReceiptRepo)
	poRepo := new(mockPurchaseOrderRepo)
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	handler, r := setupReceiptHandler(receiptRepo, poRepo, invRepo, invTransRepo)

	r.GET("/receipts/:id", handler.GetReceipt)

	receiptRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/receipts/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestReceiptHandler_DeleteReceipt_Success(t *testing.T) {
	receiptRepo := new(mockReceiptRepo)
	poRepo := new(mockPurchaseOrderRepo)
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	handler, r := setupReceiptHandler(receiptRepo, poRepo, invRepo, invTransRepo)

	r.DELETE("/receipts/:id", handler.DeleteReceipt)

	existing := &entity.Receipt{ID: "rc-001", Status: "pending"}
	receiptRepo.On("GetByID", mock.Anything, "rc-001").Return(existing, nil)
	receiptRepo.On("Delete", mock.Anything, "rc-001").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/receipts/rc-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReceiptHandler_DeleteReceipt_Error(t *testing.T) {
	receiptRepo := new(mockReceiptRepo)
	poRepo := new(mockPurchaseOrderRepo)
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	handler, r := setupReceiptHandler(receiptRepo, poRepo, invRepo, invTransRepo)

	r.DELETE("/receipts/:id", handler.DeleteReceipt)

	receiptRepo.On("GetByID", mock.Anything, "rc-001").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/receipts/rc-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestReceiptHandler_ListReceipts_Success(t *testing.T) {
	receiptRepo := new(mockReceiptRepo)
	poRepo := new(mockPurchaseOrderRepo)
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	handler, r := setupReceiptHandler(receiptRepo, poRepo, invRepo, invTransRepo)

	r.GET("/receipts", handler.ListReceipts)

	receipts := []*entity.Receipt{{ID: "rc-001"}}
	receiptRepo.On("List", mock.Anything, (*string)(nil), (*string)(nil), (*time.Time)(nil), (*time.Time)(nil), 0, 10).Return(receipts, int64(1), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/receipts", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReceiptHandler_UpdateReceiptStatus_Success(t *testing.T) {
	receiptRepo := new(mockReceiptRepo)
	poRepo := new(mockPurchaseOrderRepo)
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	handler, r := setupReceiptHandler(receiptRepo, poRepo, invRepo, invTransRepo)

	r.PUT("/receipts/:id/status", handler.UpdateReceiptStatus)

	existing := &entity.Receipt{ID: "rc-001", Status: "pending"}
	receiptRepo.On("GetByID", mock.Anything, "rc-001").Return(existing, nil)
	receiptRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.Receipt")).Return(nil)

	body := map[string]interface{}{"status": "cancelled"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/receipts/rc-001/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReceiptHandler_UpdateReceiptStatus_BadRequest(t *testing.T) {
	receiptRepo := new(mockReceiptRepo)
	poRepo := new(mockPurchaseOrderRepo)
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	handler, r := setupReceiptHandler(receiptRepo, poRepo, invRepo, invTransRepo)

	r.PUT("/receipts/:id/status", handler.UpdateReceiptStatus)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/receipts/rc-001/status", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNewReceiptHandler(t *testing.T) {
	receiptRepo := new(mockReceiptRepo)
	poRepo := new(mockPurchaseOrderRepo)
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	svc := service.NewReceiptService(receiptRepo, poRepo, invRepo, invTransRepo)
	handler := NewReceiptHandler(svc)
	assert.NotNil(t, handler)
}

func TestReceiptHandler_UpdateReceipt_Success(t *testing.T) {
	receiptRepo := new(mockReceiptRepo)
	poRepo := new(mockPurchaseOrderRepo)
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	handler, r := setupReceiptHandler(receiptRepo, poRepo, invRepo, invTransRepo)

	r.PUT("/receipts/:id", handler.UpdateReceipt)

	existing := &entity.Receipt{ID: "rc-001", PurchaseOrderID: "po-001", Status: "pending"}
	receiptRepo.On("GetByID", mock.Anything, "rc-001").Return(existing, nil)
	receiptRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.Receipt")).Return(nil)

	body := map[string]interface{}{
		"purchase_order_id": "po-001",
		"items": []map[string]interface{}{
			{"purchase_order_item_id": "poi-001", "inventory_id": "inv-001", "quantity": 5, "unit": "pcs", "unit_price": 200.0},
		},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/receipts/rc-001", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReceiptHandler_UpdateReceipt_BadRequest(t *testing.T) {
	receiptRepo := new(mockReceiptRepo)
	poRepo := new(mockPurchaseOrderRepo)
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	handler, r := setupReceiptHandler(receiptRepo, poRepo, invRepo, invTransRepo)

	r.PUT("/receipts/:id", handler.UpdateReceipt)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/receipts/rc-001", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReceiptHandler_ListReceipts_Error(t *testing.T) {
	receiptRepo := new(mockReceiptRepo)
	poRepo := new(mockPurchaseOrderRepo)
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	handler, r := setupReceiptHandler(receiptRepo, poRepo, invRepo, invTransRepo)

	r.GET("/receipts", handler.ListReceipts)

	receiptRepo.On("List", mock.Anything, (*string)(nil), (*string)(nil), (*time.Time)(nil), (*time.Time)(nil), 0, 10).Return(([]*entity.Receipt)(nil), int64(0), assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/receipts", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

var _ repository.ReceiptRepository = (*mockReceiptRepo)(nil)
var _ repository.InventoryRepository = (*mockInventoryRepo)(nil)
var _ repository.InventoryTransactionRepository = (*mockInventoryTransRepo)(nil)
