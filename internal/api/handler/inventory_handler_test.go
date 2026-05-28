package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupInventoryHandler(invRepo *mockInventoryRepo, invTransRepo *mockInventoryTransRepo, supplierRepo *mockSupplierRepo) (*InventoryHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := service.NewInventoryService(invRepo, invTransRepo, supplierRepo)
	handler := NewInventoryHandler(svc)
	r := gin.New()
	return handler, r
}

func TestInventoryHandler_CreateInventory_Success(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupInventoryHandler(invRepo, invTransRepo, supplierRepo)

	r.POST("/inventories", handler.CreateInventory)

	invRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Inventory")).Return(nil)

	body := map[string]interface{}{
		"code": "INV001", "name": "Test Item", "type": "raw_material",
		"unit": "pcs", "quantity": 100, "unit_price": 10.0,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/inventories", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInventoryHandler_CreateInventory_BadRequest(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupInventoryHandler(invRepo, invTransRepo, supplierRepo)

	r.POST("/inventories", handler.CreateInventory)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/inventories", bytes.NewBufferString(`invalid json`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInventoryHandler_CreateInventory_ServiceError(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupInventoryHandler(invRepo, invTransRepo, supplierRepo)

	r.POST("/inventories", handler.CreateInventory)

	invRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Inventory")).Return(assert.AnError)

	body := map[string]interface{}{
		"code": "INV001", "name": "Test Item", "type": "raw_material",
		"unit": "pcs", "quantity": 100, "unit_price": 10.0,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/inventories", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestInventoryHandler_GetInventory_Success(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupInventoryHandler(invRepo, invTransRepo, supplierRepo)

	r.GET("/inventories/:id", handler.GetInventory)

	invRepo.On("GetByID", mock.Anything, "inv-001").Return(&entity.Inventory{ID: "inv-001"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/inventories/inv-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInventoryHandler_GetInventory_NotFound(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupInventoryHandler(invRepo, invTransRepo, supplierRepo)

	r.GET("/inventories/:id", handler.GetInventory)

	invRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/inventories/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestInventoryHandler_UpdateInventory_Success(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupInventoryHandler(invRepo, invTransRepo, supplierRepo)

	r.PUT("/inventories/:id", handler.UpdateInventory)

	invRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.Inventory")).Return(nil)

	body := map[string]interface{}{
		"code": "INV001", "name": "Updated Item", "type": "raw_material",
		"unit": "pcs", "quantity": 200, "unit_price": 15.0,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/inventories/inv-001", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInventoryHandler_UpdateInventory_BadRequest(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupInventoryHandler(invRepo, invTransRepo, supplierRepo)

	r.PUT("/inventories/:id", handler.UpdateInventory)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/inventories/inv-001", bytes.NewBufferString(`invalid`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInventoryHandler_DeleteInventory_Success(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupInventoryHandler(invRepo, invTransRepo, supplierRepo)

	r.DELETE("/inventories/:id", handler.DeleteInventory)

	invRepo.On("Delete", mock.Anything, "inv-001").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/inventories/inv-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInventoryHandler_DeleteInventory_Error(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupInventoryHandler(invRepo, invTransRepo, supplierRepo)

	r.DELETE("/inventories/:id", handler.DeleteInventory)

	invRepo.On("Delete", mock.Anything, "inv-001").Return(assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/inventories/inv-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestInventoryHandler_ListInventories_Success(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupInventoryHandler(invRepo, invTransRepo, supplierRepo)

	r.GET("/inventories", handler.ListInventories)

	invRepo.On("List", mock.Anything, mock.AnythingOfType("*service.InventoryFilter")).Return([]*entity.Inventory{{ID: "inv-001"}}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/inventories", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInventoryHandler_ListInventories_Error(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupInventoryHandler(invRepo, invTransRepo, supplierRepo)

	r.GET("/inventories", handler.ListInventories)

	invRepo.On("List", mock.Anything, mock.AnythingOfType("*service.InventoryFilter")).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/inventories", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestInventoryHandler_GetLowStockItems_Success(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupInventoryHandler(invRepo, invTransRepo, supplierRepo)

	r.GET("/inventories/low-stock", handler.GetLowStockItems)

	invRepo.On("GetLowStockItems", mock.Anything).Return([]*entity.Inventory{{ID: "inv-001"}}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/inventories/low-stock", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInventoryHandler_GetLowStockItems_Error(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupInventoryHandler(invRepo, invTransRepo, supplierRepo)

	r.GET("/inventories/low-stock", handler.GetLowStockItems)

	invRepo.On("GetLowStockItems", mock.Anything).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/inventories/low-stock", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestInventoryHandler_ProcessTransaction_Success(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupInventoryHandler(invRepo, invTransRepo, supplierRepo)

	r.POST("/inventories/transactions", handler.ProcessTransaction)

	inv := &entity.Inventory{ID: "inv-001", Quantity: 100, UnitPrice: 10.0}
	invRepo.On("GetByID", mock.Anything, "inv-001").Return(inv, nil)
	invTransRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.InventoryTransaction")).Return(nil)
	invRepo.On("UpdateQuantity", mock.Anything, "inv-001", mock.AnythingOfType("float64")).Return(nil)
	invRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.Inventory")).Return(nil)

	body := map[string]interface{}{
		"InventoryID": "inv-001", "Type": "in", "Quantity": 50, "UnitPrice": 10.0,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/inventories/transactions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInventoryHandler_ProcessTransaction_BadRequest(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupInventoryHandler(invRepo, invTransRepo, supplierRepo)

	r.POST("/inventories/transactions", handler.ProcessTransaction)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/inventories/transactions", bytes.NewBufferString(`invalid`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInventoryHandler_GetTransactions_Success(t *testing.T) {
	invTransRepo := new(mockInventoryTransRepo)
	invRepo := new(mockInventoryRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupInventoryHandler(invRepo, invTransRepo, supplierRepo)

	r.GET("/inventories/:inventory_id/transactions", handler.GetTransactions)

	invTransRepo.On("ListByInventoryID", mock.Anything, "inv-001").Return([]*entity.InventoryTransaction{{ID: "trans-001"}}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/inventories/inv-001/transactions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInventoryHandler_GetTransactions_Error(t *testing.T) {
	invTransRepo := new(mockInventoryTransRepo)
	invRepo := new(mockInventoryRepo)
	supplierRepo := new(mockSupplierRepo)
	handler, r := setupInventoryHandler(invRepo, invTransRepo, supplierRepo)

	r.GET("/inventories/:inventory_id/transactions", handler.GetTransactions)

	invTransRepo.On("ListByInventoryID", mock.Anything, "inv-001").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/inventories/inv-001/transactions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestNewInventoryHandler(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	invTransRepo := new(mockInventoryTransRepo)
	supplierRepo := new(mockSupplierRepo)
	svc := service.NewInventoryService(invRepo, invTransRepo, supplierRepo)
	handler := NewInventoryHandler(svc)
	assert.NotNil(t, handler)
}

var _ context.Context = nil
