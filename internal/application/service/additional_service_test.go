package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockWorkOrderRepo struct {
	mock.Mock
}

func (m *mockWorkOrderRepo) Create(ctx context.Context, wo *entity.WorkOrder) error {
	args := m.Called(ctx, wo)
	return args.Error(0)
}

func (m *mockWorkOrderRepo) Update(ctx context.Context, wo *entity.WorkOrder) error {
	args := m.Called(ctx, wo)
	return args.Error(0)
}

func (m *mockWorkOrderRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockWorkOrderRepo) GetByID(ctx context.Context, id string) (*entity.WorkOrder, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkOrder), args.Error(1)
}

func (m *mockWorkOrderRepo) List(ctx context.Context, filter interface{}) ([]*entity.WorkOrder, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WorkOrder), args.Error(1)
}

func (m *mockWorkOrderRepo) Count(ctx context.Context, filter interface{}) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

type mockDeviceRepoForWO struct {
	mock.Mock
}

func (m *mockDeviceRepoForWO) Create(ctx context.Context, device *entity.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *mockDeviceRepoForWO) Update(ctx context.Context, device *entity.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *mockDeviceRepoForWO) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockDeviceRepoForWO) GetByID(ctx context.Context, id string) (*entity.Device, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}

func (m *mockDeviceRepoForWO) GetByCode(ctx context.Context, code string) (*entity.Device, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}

func (m *mockDeviceRepoForWO) List(ctx context.Context, stationID *string, deviceType *entity.DeviceType) ([]*entity.Device, error) {
	args := m.Called(ctx, stationID, deviceType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Device), args.Error(1)
}

func (m *mockDeviceRepoForWO) GetOnlineDevices(ctx context.Context, stationID string) ([]*entity.Device, error) {
	args := m.Called(ctx, stationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Device), args.Error(1)
}

func (m *mockDeviceRepoForWO) GetWithPoints(ctx context.Context, id string) (*entity.Device, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}

func TestWorkOrderService_CreateWorkOrder_Success(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	devRepo := new(mockDeviceRepoForWO)
	svc := NewWorkOrderService(woRepo, devRepo)
	ctx := context.Background()

	devRepo.On("GetByID", ctx, "device-001").Return(&entity.Device{ID: "device-001"}, nil)
	woRepo.On("Create", ctx, mock.AnythingOfType("*entity.WorkOrder")).Return(nil)

	req := &CreateWorkOrderRequest{
		DeviceID:    "device-001",
		Type:        "maintenance",
		Title:       "Test WO",
		Description: "Test description",
		Priority:    "high",
		Assignee:    "admin",
		CreatedBy:   "admin",
	}
	wo, err := svc.CreateWorkOrder(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, wo)
	assert.Equal(t, "Test WO", wo.Title)
	assert.Equal(t, "open", wo.Status)
}

func TestWorkOrderService_CreateWorkOrder_DeviceNotFound(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	devRepo := new(mockDeviceRepoForWO)
	svc := NewWorkOrderService(woRepo, devRepo)
	ctx := context.Background()

	devRepo.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))

	req := &CreateWorkOrderRequest{
		DeviceID: "nonexistent",
		Type:     "maintenance",
		Title:    "Test WO",
	}
	_, err := svc.CreateWorkOrder(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "device not found")
}

func TestWorkOrderService_CreateWorkOrder_NoDevice(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	devRepo := new(mockDeviceRepoForWO)
	svc := NewWorkOrderService(woRepo, devRepo)
	ctx := context.Background()

	woRepo.On("Create", ctx, mock.AnythingOfType("*entity.WorkOrder")).Return(nil)

	req := &CreateWorkOrderRequest{
		Type:  "maintenance",
		Title: "Test WO",
	}
	wo, err := svc.CreateWorkOrder(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, wo)
}

func TestWorkOrderService_CreateWorkOrder_RepoError(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	devRepo := new(mockDeviceRepoForWO)
	svc := NewWorkOrderService(woRepo, devRepo)
	ctx := context.Background()

	woRepo.On("Create", ctx, mock.AnythingOfType("*entity.WorkOrder")).Return(errors.New("db error"))

	req := &CreateWorkOrderRequest{
		Type:  "maintenance",
		Title: "Test WO",
	}
	_, err := svc.CreateWorkOrder(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create work order")
}

func TestWorkOrderService_UpdateWorkOrder_Success(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	devRepo := new(mockDeviceRepoForWO)
	svc := NewWorkOrderService(woRepo, devRepo)
	ctx := context.Background()

	existing := &entity.WorkOrder{ID: "wo-001", Title: "Old Title", Status: "open"}
	woRepo.On("GetByID", ctx, "wo-001").Return(existing, nil)
	woRepo.On("Update", ctx, mock.AnythingOfType("*entity.WorkOrder")).Return(nil)

	req := &UpdateWorkOrderRequest{Title: "New Title", Status: "in_progress"}
	wo, err := svc.UpdateWorkOrder(ctx, "wo-001", req)
	assert.NoError(t, err)
	assert.Equal(t, "New Title", wo.Title)
	assert.Equal(t, "in_progress", wo.Status)
	assert.NotNil(t, wo.StartDate)
}

func TestWorkOrderService_UpdateWorkOrder_Completed(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	devRepo := new(mockDeviceRepoForWO)
	svc := NewWorkOrderService(woRepo, devRepo)
	ctx := context.Background()

	existing := &entity.WorkOrder{ID: "wo-001", Title: "Title", Status: "in_progress"}
	woRepo.On("GetByID", ctx, "wo-001").Return(existing, nil)
	woRepo.On("Update", ctx, mock.AnythingOfType("*entity.WorkOrder")).Return(nil)

	req := &UpdateWorkOrderRequest{Status: "completed"}
	wo, err := svc.UpdateWorkOrder(ctx, "wo-001", req)
	assert.NoError(t, err)
	assert.Equal(t, "completed", wo.Status)
	assert.NotNil(t, wo.EndDate)
}

func TestWorkOrderService_UpdateWorkOrder_NotFound(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	devRepo := new(mockDeviceRepoForWO)
	svc := NewWorkOrderService(woRepo, devRepo)
	ctx := context.Background()

	woRepo.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))

	req := &UpdateWorkOrderRequest{Title: "New Title"}
	_, err := svc.UpdateWorkOrder(ctx, "nonexistent", req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "work order not found")
}

func TestWorkOrderService_GetWorkOrder(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	devRepo := new(mockDeviceRepoForWO)
	svc := NewWorkOrderService(woRepo, devRepo)
	ctx := context.Background()

	wo := &entity.WorkOrder{ID: "wo-001", Title: "Test WO"}
	woRepo.On("GetByID", ctx, "wo-001").Return(wo, nil)

	found, err := svc.GetWorkOrder(ctx, "wo-001")
	assert.NoError(t, err)
	assert.Equal(t, "Test WO", found.Title)
}

func TestWorkOrderService_DeleteWorkOrder(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	devRepo := new(mockDeviceRepoForWO)
	svc := NewWorkOrderService(woRepo, devRepo)
	ctx := context.Background()

	woRepo.On("GetByID", ctx, "wo-001").Return(&entity.WorkOrder{ID: "wo-001", Title: "Test WO"}, nil)
	woRepo.On("Delete", ctx, "wo-001").Return(nil)

	err := svc.DeleteWorkOrder(ctx, "wo-001")
	assert.NoError(t, err)
}

func TestWorkOrderService_ListWorkOrders(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	devRepo := new(mockDeviceRepoForWO)
	svc := NewWorkOrderService(woRepo, devRepo)
	ctx := context.Background()

	wos := []*entity.WorkOrder{{ID: "wo-001", Title: "WO 1"}}
	woRepo.On("List", ctx, mock.Anything).Return(wos, nil)

	found, err := svc.ListWorkOrders(ctx, &WorkOrderFilter{})
	assert.NoError(t, err)
	assert.Len(t, found, 1)
}

type mockInventoryRepo struct {
	mock.Mock
}

func (m *mockInventoryRepo) Create(ctx context.Context, inv *entity.Inventory) error {
	args := m.Called(ctx, inv)
	return args.Error(0)
}

func (m *mockInventoryRepo) Update(ctx context.Context, inv *entity.Inventory) error {
	args := m.Called(ctx, inv)
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

func (m *mockInventoryTransRepo) Create(ctx context.Context, tx *entity.InventoryTransaction) error {
	args := m.Called(ctx, tx)
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

func (m *mockInventoryTransRepo) ListByReference(ctx context.Context, refID, refType string) ([]*entity.InventoryTransaction, error) {
	args := m.Called(ctx, refID, refType)
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

type mockSupplierRepo struct {
	mock.Mock
}

func (m *mockSupplierRepo) Create(ctx context.Context, sup *entity.Supplier) error {
	args := m.Called(ctx, sup)
	return args.Error(0)
}

func (m *mockSupplierRepo) Update(ctx context.Context, sup *entity.Supplier) error {
	args := m.Called(ctx, sup)
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

func TestInventoryService_CreateInventory(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	supRepo := new(mockSupplierRepo)
	svc := NewInventoryService(invRepo, txRepo, supRepo)
	ctx := context.Background()

	invRepo.On("Create", ctx, mock.AnythingOfType("*entity.Inventory")).Return(nil)

	inv := &entity.Inventory{
		Code:       "INV001",
		Name:       "Test Item",
		Quantity:   100,
		UnitPrice:  10.0,
	}
	err := svc.CreateInventory(ctx, inv)
	assert.NoError(t, err)
	assert.Equal(t, 1000.0, inv.TotalValue)
}

func TestInventoryService_UpdateInventory(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	supRepo := new(mockSupplierRepo)
	svc := NewInventoryService(invRepo, txRepo, supRepo)
	ctx := context.Background()

	invRepo.On("Update", ctx, mock.AnythingOfType("*entity.Inventory")).Return(nil)

	inv := &entity.Inventory{
		Code:       "INV001",
		Name:       "Test Item",
		Quantity:   50,
		UnitPrice:  20.0,
	}
	err := svc.UpdateInventory(ctx, inv)
	assert.NoError(t, err)
	assert.Equal(t, 1000.0, inv.TotalValue)
}

func TestInventoryService_DeleteInventory(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	supRepo := new(mockSupplierRepo)
	svc := NewInventoryService(invRepo, txRepo, supRepo)
	ctx := context.Background()

	invRepo.On("Delete", ctx, "inv-001").Return(nil)

	err := svc.DeleteInventory(ctx, "inv-001")
	assert.NoError(t, err)
}

func TestInventoryService_GetInventoryByID(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	supRepo := new(mockSupplierRepo)
	svc := NewInventoryService(invRepo, txRepo, supRepo)
	ctx := context.Background()

	inv := &entity.Inventory{ID: "inv-001", Name: "Test Item"}
	invRepo.On("GetByID", ctx, "inv-001").Return(inv, nil)

	found, err := svc.GetInventoryByID(ctx, "inv-001")
	assert.NoError(t, err)
	assert.Equal(t, "Test Item", found.Name)
}

func TestInventoryService_GetInventoryByCode(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	supRepo := new(mockSupplierRepo)
	svc := NewInventoryService(invRepo, txRepo, supRepo)
	ctx := context.Background()

	inv := &entity.Inventory{ID: "inv-001", Code: "INV001"}
	invRepo.On("GetByCode", ctx, "INV001").Return(inv, nil)

	found, err := svc.GetInventoryByCode(ctx, "INV001")
	assert.NoError(t, err)
	assert.Equal(t, "inv-001", found.ID)
}

func TestInventoryService_ListInventories(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	supRepo := new(mockSupplierRepo)
	svc := NewInventoryService(invRepo, txRepo, supRepo)
	ctx := context.Background()

	invs := []*entity.Inventory{{ID: "inv-001"}}
	invRepo.On("List", ctx, mock.Anything).Return(invs, nil)

	found, err := svc.ListInventories(ctx, &InventoryFilter{})
	assert.NoError(t, err)
	assert.Len(t, found, 1)
}

func TestInventoryService_CountInventories(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	supRepo := new(mockSupplierRepo)
	svc := NewInventoryService(invRepo, txRepo, supRepo)
	ctx := context.Background()

	invRepo.On("Count", ctx, mock.Anything).Return(int64(10), nil)

	count, err := svc.CountInventories(ctx, &InventoryFilter{})
	assert.NoError(t, err)
	assert.Equal(t, int64(10), count)
}

func TestInventoryService_GetLowStockItems(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	supRepo := new(mockSupplierRepo)
	svc := NewInventoryService(invRepo, txRepo, supRepo)
	ctx := context.Background()

	invs := []*entity.Inventory{{ID: "inv-001", Status: "low_stock"}}
	invRepo.On("GetLowStockItems", ctx).Return(invs, nil)

	found, err := svc.GetLowStockItems(ctx)
	assert.NoError(t, err)
	assert.Len(t, found, 1)
}

type mockForecastResultRepo struct {
	mock.Mock
}

func (m *mockForecastResultRepo) Create(ctx context.Context, result *entity.ForecastResult) error {
	args := m.Called(ctx, result)
	return args.Error(0)
}

func (m *mockForecastResultRepo) GetByID(ctx context.Context, id string) (*entity.ForecastResult, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ForecastResult), args.Error(1)
}

func (m *mockForecastResultRepo) ListByStation(ctx context.Context, stationID string, forecastType *entity.ForecastType, startTime, endTime *time.Time, page, pageSize int) ([]*entity.ForecastResult, int64, error) {
	args := m.Called(ctx, stationID, forecastType, startTime, endTime, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.ForecastResult), args.Get(1).(int64), args.Error(2)
}

func (m *mockForecastResultRepo) UpdateActualPower(ctx context.Context, id string, power float64) error {
	args := m.Called(ctx, id, power)
	return args.Error(0)
}

func (m *mockForecastResultRepo) GetAccuracyStats(ctx context.Context, stationID string, forecastType *entity.ForecastType, start, end time.Time) (*repository.ForecastAccuracyStats, error) {
	args := m.Called(ctx, stationID, forecastType, start, end)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.ForecastAccuracyStats), args.Error(1)
}

func TestForecastService_PowerForecast_ExistingResults(t *testing.T) {
	resultRepo := new(mockForecastResultRepo)
	svc := NewForecastService(nil, nil, resultRepo)
	ctx := context.Background()

	existing := []*entity.ForecastResult{
		{StationID: "s1", PredictedPower: 1000.0},
	}
	resultRepo.On("ListByStation", ctx, "s1", mock.Anything, mock.Anything, mock.Anything, 0, 100).Return(existing, int64(1), nil)

	found, err := svc.PowerForecast(ctx, "s1", entity.ForecastTypeShortTerm)
	assert.NoError(t, err)
	assert.Len(t, found, 1)
}

func TestForecastService_PowerForecast_NoExisting(t *testing.T) {
	resultRepo := new(mockForecastResultRepo)
	svc := NewForecastService(nil, nil, resultRepo)
	ctx := context.Background()

	resultRepo.On("ListByStation", ctx, "s1", mock.Anything, mock.Anything, mock.Anything, 0, 100).Return([]*entity.ForecastResult{}, int64(0), nil)
	resultRepo.On("Create", ctx, mock.AnythingOfType("*entity.ForecastResult")).Return(nil)

	found, err := svc.PowerForecast(ctx, "s1", entity.ForecastTypeShortTerm)
	assert.NoError(t, err)
	assert.Greater(t, len(found), 0)
}

func TestForecastService_PowerForecast_UltraShortTerm(t *testing.T) {
	resultRepo := new(mockForecastResultRepo)
	svc := NewForecastService(nil, nil, resultRepo)
	ctx := context.Background()

	resultRepo.On("ListByStation", ctx, "s1", mock.Anything, mock.Anything, mock.Anything, 0, 100).Return([]*entity.ForecastResult{}, int64(0), nil)
	resultRepo.On("Create", ctx, mock.AnythingOfType("*entity.ForecastResult")).Return(nil)

	found, err := svc.PowerForecast(ctx, "s1", entity.ForecastTypeUltraShortTerm)
	assert.NoError(t, err)
	assert.Greater(t, len(found), 0)
}

func TestForecastService_PowerForecast_MediumTerm(t *testing.T) {
	resultRepo := new(mockForecastResultRepo)
	svc := NewForecastService(nil, nil, resultRepo)
	ctx := context.Background()

	resultRepo.On("ListByStation", ctx, "s1", mock.Anything, mock.Anything, mock.Anything, 0, 100).Return([]*entity.ForecastResult{}, int64(0), nil)
	resultRepo.On("Create", ctx, mock.AnythingOfType("*entity.ForecastResult")).Return(nil)

	found, err := svc.PowerForecast(ctx, "s1", entity.ForecastTypeMediumTerm)
	assert.NoError(t, err)
	assert.Greater(t, len(found), 0)
}
