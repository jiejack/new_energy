package service

import (
	"context"
	"testing"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPORepoExtra struct {
	mock.Mock
}

func (m *mockPORepoExtra) Create(ctx context.Context, order *entity.PurchaseOrder) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *mockPORepoExtra) Update(ctx context.Context, order *entity.PurchaseOrder) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *mockPORepoExtra) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockPORepoExtra) GetByID(ctx context.Context, id string) (*entity.PurchaseOrder, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.PurchaseOrder), args.Error(1)
}

func (m *mockPORepoExtra) List(ctx context.Context, supplierID *string, status *string, startDate, endDate *time.Time, offset, limit int) ([]*entity.PurchaseOrder, int64, error) {
	args := m.Called(ctx, supplierID, status, startDate, endDate, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.PurchaseOrder), args.Get(1).(int64), args.Error(2)
}

func (m *mockPORepoExtra) GetByCode(ctx context.Context, code string) (*entity.PurchaseOrder, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.PurchaseOrder), args.Error(1)
}

type mockReceiptRepoExtra struct {
	mock.Mock
}

func (m *mockReceiptRepoExtra) Create(ctx context.Context, receipt *entity.Receipt) error {
	args := m.Called(ctx, receipt)
	return args.Error(0)
}

func (m *mockReceiptRepoExtra) Update(ctx context.Context, receipt *entity.Receipt) error {
	args := m.Called(ctx, receipt)
	return args.Error(0)
}

func (m *mockReceiptRepoExtra) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockReceiptRepoExtra) GetByID(ctx context.Context, id string) (*entity.Receipt, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Receipt), args.Error(1)
}

func (m *mockReceiptRepoExtra) List(ctx context.Context, purchaseOrderID *string, status *string, startDate, endDate *time.Time, offset, limit int) ([]*entity.Receipt, int64, error) {
	args := m.Called(ctx, purchaseOrderID, status, startDate, endDate, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.Receipt), args.Get(1).(int64), args.Error(2)
}

func (m *mockReceiptRepoExtra) GetByCode(ctx context.Context, code string) (*entity.Receipt, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Receipt), args.Error(1)
}

func (m *mockReceiptRepoExtra) GetByPurchaseOrderID(ctx context.Context, purchaseOrderID string) ([]*entity.Receipt, error) {
	args := m.Called(ctx, purchaseOrderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Receipt), args.Error(1)
}

type mockFaultRepoExtra struct {
	mock.Mock
}

func (m *mockFaultRepoExtra) Create(ctx context.Context, result *entity.FaultDetectionResult) error {
	args := m.Called(ctx, result)
	return args.Error(0)
}

func (m *mockFaultRepoExtra) GetByID(ctx context.Context, id string) (*entity.FaultDetectionResult, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.FaultDetectionResult), args.Error(1)
}

func (m *mockFaultRepoExtra) ListByDevice(ctx context.Context, deviceID string, severity *entity.FaultSeverity, status *entity.FaultDetectionStatus, offset, limit int) ([]*entity.FaultDetectionResult, int64, error) {
	args := m.Called(ctx, deviceID, severity, status, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.FaultDetectionResult), args.Get(1).(int64), args.Error(2)
}

func (m *mockFaultRepoExtra) ListByStation(ctx context.Context, stationID string, severity *entity.FaultSeverity, offset, limit int) ([]*entity.FaultDetectionResult, int64, error) {
	args := m.Called(ctx, stationID, severity, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.FaultDetectionResult), args.Get(1).(int64), args.Error(2)
}

func (m *mockFaultRepoExtra) UpdateStatus(ctx context.Context, id string, status entity.FaultDetectionStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *mockFaultRepoExtra) UpdateRootCause(ctx context.Context, id, rootCause string) error {
	args := m.Called(ctx, id, rootCause)
	return args.Error(0)
}

func (m *mockFaultRepoExtra) LinkWorkOrder(ctx context.Context, id, workOrderID string) error {
	args := m.Called(ctx, id, workOrderID)
	return args.Error(0)
}

func (m *mockFaultRepoExtra) CountBySeverity(ctx context.Context, deviceID *string) (map[entity.FaultSeverity]int64, error) {
	args := m.Called(ctx, deviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[entity.FaultSeverity]int64), args.Error(1)
}

type mockWOCreatorExtra struct {
	mock.Mock
}

func (m *mockWOCreatorExtra) CreateFromFaultDetection(ctx context.Context, detection *entity.FaultDetectionResult) (string, error) {
	args := m.Called(ctx, detection)
	return args.String(0), args.Error(1)
}

type mockAlarmRepoExtra struct {
	mock.Mock
}

func (m *mockAlarmRepoExtra) Create(ctx context.Context, alarm *entity.Alarm) error {
	args := m.Called(ctx, alarm)
	return args.Error(0)
}

func (m *mockAlarmRepoExtra) GetByID(ctx context.Context, id string) (*entity.Alarm, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Alarm), args.Error(1)
}

func (m *mockAlarmRepoExtra) Update(ctx context.Context, alarm *entity.Alarm) error {
	args := m.Called(ctx, alarm)
	return args.Error(0)
}

func (m *mockAlarmRepoExtra) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockAlarmRepoExtra) List(ctx context.Context, stationID *string, level *entity.AlarmLevel, status *entity.AlarmStatus, startTime, endTime int64, offset, limit int) ([]*entity.Alarm, int64, error) {
	args := m.Called(ctx, stationID, level, status, startTime, endTime, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.Alarm), args.Get(1).(int64), args.Error(2)
}

func (m *mockAlarmRepoExtra) GetActiveAlarms(ctx context.Context, stationID *string, level *entity.AlarmLevel) ([]*entity.Alarm, error) {
	args := m.Called(ctx, stationID, level)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Alarm), args.Error(1)
}

func (m *mockAlarmRepoExtra) GetHistoryAlarms(ctx context.Context, stationID *string, startTime, endTime int64) ([]*entity.Alarm, error) {
	args := m.Called(ctx, stationID, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Alarm), args.Error(1)
}

func (m *mockAlarmRepoExtra) Acknowledge(ctx context.Context, id, by string) error {
	args := m.Called(ctx, id, by)
	return args.Error(0)
}

func (m *mockAlarmRepoExtra) Clear(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockAlarmRepoExtra) CountByLevel(ctx context.Context, stationID *string) (map[entity.AlarmLevel]int64, error) {
	args := m.Called(ctx, stationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[entity.AlarmLevel]int64), args.Error(1)
}

func TestNewAlarmHarnessWithComponents_Extra(t *testing.T) {
	ah := NewAlarmHarnessWithComponents(nil)
	assert.NotNil(t, ah)
}

func TestAlarmServiceWithHarness_GetAlarm_Extra(t *testing.T) {
	repo := new(mockAlarmRepoExtra)
	svc := NewAlarmServiceWithHarness(repo)
	ctx := context.Background()

	alarm := &entity.Alarm{ID: "a-001", Title: "Test Alarm"}
	repo.On("GetByID", ctx, "a-001").Return(alarm, nil)

	result, err := svc.GetAlarm(ctx, "a-001")
	assert.NoError(t, err)
	assert.Equal(t, "a-001", result.ID)
}

func TestAlarmServiceWithHarness_GetActiveAlarms_Extra(t *testing.T) {
	repo := new(mockAlarmRepoExtra)
	svc := NewAlarmServiceWithHarness(repo)
	ctx := context.Background()

	alarms := []*entity.Alarm{{ID: "a-001"}}
	repo.On("GetActiveAlarms", ctx, (*string)(nil), (*entity.AlarmLevel)(nil)).Return(alarms, nil)

	result, err := svc.GetActiveAlarms(ctx, nil, nil)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestAlarmServiceWithHarness_CountAlarmsByLevel_Extra(t *testing.T) {
	repo := new(mockAlarmRepoExtra)
	svc := NewAlarmServiceWithHarness(repo)
	ctx := context.Background()

	counts := map[entity.AlarmLevel]int64{entity.AlarmLevelWarning: 5}
	repo.On("CountByLevel", ctx, (*string)(nil)).Return(counts, nil)

	result, err := svc.CountAlarmsByLevel(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), result[entity.AlarmLevelWarning])
}

func TestAlarmServiceWithHarness_AcknowledgeAlarm_NotActive_Extra(t *testing.T) {
	repo := new(mockAlarmRepoExtra)
	svc := NewAlarmServiceWithHarness(repo)
	ctx := context.Background()

	clearedAlarm := &entity.Alarm{ID: "a-001", Status: entity.AlarmStatusCleared}
	repo.On("GetByID", ctx, "a-001").Return(clearedAlarm, nil)

	err := svc.AcknowledgeAlarm(ctx, "a-001", "admin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in active state")
}

func TestAlarmServiceWithHarness_ClearAlarm_NotActiveOrAcked_Extra(t *testing.T) {
	repo := new(mockAlarmRepoExtra)
	svc := NewAlarmServiceWithHarness(repo)
	ctx := context.Background()

	pendingAlarm := &entity.Alarm{ID: "a-001", Status: entity.AlarmStatusSuppressed}
	repo.On("GetByID", ctx, "a-001").Return(pendingAlarm, nil)

	err := svc.ClearAlarm(ctx, "a-001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be cleared")
}

func TestAlarmServiceWithHarness_ClearAlarm_Acked_Extra(t *testing.T) {
	repo := new(mockAlarmRepoExtra)
	svc := NewAlarmServiceWithHarness(repo)
	ctx := context.Background()

	ackedAlarm := &entity.Alarm{ID: "a-001", Status: entity.AlarmStatusAcknowledged}
	repo.On("GetByID", ctx, "a-001").Return(ackedAlarm, nil)
	repo.On("Clear", ctx, "a-001").Return(nil)

	err := svc.ClearAlarm(ctx, "a-001")
	assert.NoError(t, err)
}

func TestAlarmServiceWithHarness_VerifyAlarmState_Extra(t *testing.T) {
	repo := new(mockAlarmRepoExtra)
	svc := NewAlarmServiceWithHarness(repo)
	ctx := context.Background()

	alarm := &entity.Alarm{ID: "a-001", Status: entity.AlarmStatusActive}
	repo.On("GetByID", ctx, "a-001").Return(alarm, nil)

	match, err := svc.VerifyAlarmState(ctx, "a-001", entity.AlarmStatusActive)
	assert.NoError(t, err)
	assert.True(t, match)
}

func TestAlarmServiceWithHarness_VerifyAlarmState_Mismatch_Extra(t *testing.T) {
	repo := new(mockAlarmRepoExtra)
	svc := NewAlarmServiceWithHarness(repo)
	ctx := context.Background()

	alarm := &entity.Alarm{ID: "a-001", Status: entity.AlarmStatusActive}
	repo.On("GetByID", ctx, "a-001").Return(alarm, nil)

	match, err := svc.VerifyAlarmState(ctx, "a-001", entity.AlarmStatusCleared)
	assert.NoError(t, err)
	assert.False(t, match)
}

func TestAlarmServiceWithHarness_VerifyAlarmState_Error_Extra(t *testing.T) {
	repo := new(mockAlarmRepoExtra)
	svc := NewAlarmServiceWithHarness(repo)
	ctx := context.Background()

	repo.On("GetByID", ctx, "nonexistent").Return(nil, assert.AnError)

	match, err := svc.VerifyAlarmState(ctx, "nonexistent", entity.AlarmStatusActive)
	assert.Error(t, err)
	assert.False(t, match)
}

func TestNewAlarmServiceWithHarnessComponents_Extra(t *testing.T) {
	repo := new(mockAlarmRepoExtra)
	svc := NewAlarmServiceWithHarnessComponents(repo, nil)
	assert.NotNil(t, svc)
}

func TestFaultWorkOrderBridge_ShouldAutoCreateWorkOrder_Extra(t *testing.T) {
	bridge := NewFaultWorkOrderBridge(nil, nil)

	assert.True(t, bridge.ShouldAutoCreateWorkOrder(entity.FaultSeverityFatal))
	assert.True(t, bridge.ShouldAutoCreateWorkOrder(entity.FaultSeverityCritical))
	assert.False(t, bridge.ShouldAutoCreateWorkOrder(entity.FaultSeverityWarning))
	assert.False(t, bridge.ShouldAutoCreateWorkOrder(entity.FaultSeverityInfo))
}

func TestFaultWorkOrderBridge_GetWorkOrderSLA_Extra(t *testing.T) {
	bridge := NewFaultWorkOrderBridge(nil, nil)

	assert.Equal(t, "30min", bridge.GetWorkOrderSLA(entity.FaultSeverityFatal))
	assert.Equal(t, "1h", bridge.GetWorkOrderSLA(entity.FaultSeverityCritical))
	assert.Equal(t, "4h", bridge.GetWorkOrderSLA(entity.FaultSeverityWarning))
	assert.Equal(t, "24h", bridge.GetWorkOrderSLA(entity.FaultSeverityInfo))
}

func TestFaultWorkOrderBridge_CreateWorkOrderFromDetection_InfoSeverity_Extra(t *testing.T) {
	faultRepo := new(mockFaultRepoExtra)
	creator := new(mockWOCreatorExtra)
	bridge := NewFaultWorkOrderBridge(faultRepo, creator)
	ctx := context.Background()

	detection := &entity.FaultDetectionResult{ID: "d-001", Severity: entity.FaultSeverityInfo}
	faultRepo.On("GetByID", ctx, "d-001").Return(detection, nil)

	woID, err := bridge.CreateWorkOrderFromDetection(ctx, "d-001")
	assert.NoError(t, err)
	assert.Empty(t, woID)
}

func TestFaultWorkOrderBridge_CreateWorkOrderFromDetection_Critical_Extra(t *testing.T) {
	faultRepo := new(mockFaultRepoExtra)
	creator := new(mockWOCreatorExtra)
	bridge := NewFaultWorkOrderBridge(faultRepo, creator)
	ctx := context.Background()

	detection := &entity.FaultDetectionResult{ID: "d-001", Severity: entity.FaultSeverityCritical}
	faultRepo.On("GetByID", ctx, "d-001").Return(detection, nil)
	creator.On("CreateFromFaultDetection", ctx, detection).Return("wo-001", nil)
	faultRepo.On("LinkWorkOrder", ctx, "d-001", "wo-001").Return(nil)

	woID, err := bridge.CreateWorkOrderFromDetection(ctx, "d-001")
	assert.NoError(t, err)
	assert.Equal(t, "wo-001", woID)
}

func TestFaultWorkOrderBridge_CreateWorkOrderFromDetection_NotFound_Extra(t *testing.T) {
	faultRepo := new(mockFaultRepoExtra)
	creator := new(mockWOCreatorExtra)
	bridge := NewFaultWorkOrderBridge(faultRepo, creator)
	ctx := context.Background()

	faultRepo.On("GetByID", ctx, "nonexistent").Return(nil, assert.AnError)

	woID, err := bridge.CreateWorkOrderFromDetection(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Empty(t, woID)
}

func TestFaultWorkOrderBridge_CreateWorkOrderFromDetection_LinkError_Extra(t *testing.T) {
	faultRepo := new(mockFaultRepoExtra)
	creator := new(mockWOCreatorExtra)
	bridge := NewFaultWorkOrderBridge(faultRepo, creator)
	ctx := context.Background()

	detection := &entity.FaultDetectionResult{ID: "d-001", Severity: entity.FaultSeverityCritical}
	faultRepo.On("GetByID", ctx, "d-001").Return(detection, nil)
	creator.On("CreateFromFaultDetection", ctx, detection).Return("wo-001", nil)
	faultRepo.On("LinkWorkOrder", ctx, "d-001", "wo-001").Return(assert.AnError)

	woID, err := bridge.CreateWorkOrderFromDetection(ctx, "d-001")
	assert.Error(t, err)
	assert.Empty(t, woID)
}

func TestInventoryService_ProcessInventoryTransaction_In_Extra(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	supRepo := new(mockSupplierRepo)
	svc := NewInventoryService(invRepo, txRepo, supRepo)
	ctx := context.Background()

	inv := &entity.Inventory{ID: "inv-001", Quantity: 100, UnitPrice: 10.0}
	invRepo.On("GetByID", ctx, "inv-001").Return(inv, nil)
	txRepo.On("Create", ctx, mock.AnythingOfType("*entity.InventoryTransaction")).Return(nil)
	invRepo.On("UpdateQuantity", ctx, "inv-001", 150.0).Return(nil)
	invRepo.On("Update", ctx, mock.AnythingOfType("*entity.Inventory")).Return(nil)

	req := &InventoryTransactionRequest{
		InventoryID: "inv-001", Type: "in", Quantity: 50, UnitPrice: 10.0,
		ReferenceID: "ref-001", ReferenceType: "purchase", OperatorID: "user-001",
	}
	err := svc.ProcessInventoryTransaction(ctx, req)
	assert.NoError(t, err)
}

func TestInventoryService_ProcessInventoryTransaction_Out_Extra(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	supRepo := new(mockSupplierRepo)
	svc := NewInventoryService(invRepo, txRepo, supRepo)
	ctx := context.Background()

	inv := &entity.Inventory{ID: "inv-001", Quantity: 100, UnitPrice: 10.0}
	invRepo.On("GetByID", ctx, "inv-001").Return(inv, nil)
	txRepo.On("Create", ctx, mock.AnythingOfType("*entity.InventoryTransaction")).Return(nil)
	invRepo.On("UpdateQuantity", ctx, "inv-001", 60.0).Return(nil)
	invRepo.On("Update", ctx, mock.AnythingOfType("*entity.Inventory")).Return(nil)

	req := &InventoryTransactionRequest{
		InventoryID: "inv-001", Type: "out", Quantity: 40, UnitPrice: 10.0,
	}
	err := svc.ProcessInventoryTransaction(ctx, req)
	assert.NoError(t, err)
}

func TestInventoryService_ProcessInventoryTransaction_OutInsufficient_Extra(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	supRepo := new(mockSupplierRepo)
	svc := NewInventoryService(invRepo, txRepo, supRepo)
	ctx := context.Background()

	inv := &entity.Inventory{ID: "inv-001", Quantity: 10, UnitPrice: 10.0}
	invRepo.On("GetByID", ctx, "inv-001").Return(inv, nil)

	req := &InventoryTransactionRequest{
		InventoryID: "inv-001", Type: "out", Quantity: 50, UnitPrice: 10.0,
	}
	err := svc.ProcessInventoryTransaction(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient")
}

func TestInventoryService_ProcessInventoryTransaction_Adjustment_Extra(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	supRepo := new(mockSupplierRepo)
	svc := NewInventoryService(invRepo, txRepo, supRepo)
	ctx := context.Background()

	inv := &entity.Inventory{ID: "inv-001", Quantity: 100, UnitPrice: 10.0}
	invRepo.On("GetByID", ctx, "inv-001").Return(inv, nil)
	txRepo.On("Create", ctx, mock.AnythingOfType("*entity.InventoryTransaction")).Return(nil)
	invRepo.On("UpdateQuantity", ctx, "inv-001", 80.0).Return(nil)
	invRepo.On("Update", ctx, mock.AnythingOfType("*entity.Inventory")).Return(nil)

	req := &InventoryTransactionRequest{
		InventoryID: "inv-001", Type: "adjustment", Quantity: 80, UnitPrice: 10.0,
	}
	err := svc.ProcessInventoryTransaction(ctx, req)
	assert.NoError(t, err)
}

func TestInventoryService_ProcessInventoryTransaction_InvalidType_Extra(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	supRepo := new(mockSupplierRepo)
	svc := NewInventoryService(invRepo, txRepo, supRepo)
	ctx := context.Background()

	inv := &entity.Inventory{ID: "inv-001", Quantity: 100, UnitPrice: 10.0}
	invRepo.On("GetByID", ctx, "inv-001").Return(inv, nil)

	req := &InventoryTransactionRequest{
		InventoryID: "inv-001", Type: "invalid", Quantity: 50, UnitPrice: 10.0,
	}
	err := svc.ProcessInventoryTransaction(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid transaction type")
}

func TestInventoryService_ProcessInventoryTransaction_NotFound_Extra(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	supRepo := new(mockSupplierRepo)
	svc := NewInventoryService(invRepo, txRepo, supRepo)
	ctx := context.Background()

	invRepo.On("GetByID", ctx, "inv-999").Return(nil, assert.AnError)

	req := &InventoryTransactionRequest{InventoryID: "inv-999", Type: "in", Quantity: 50}
	err := svc.ProcessInventoryTransaction(ctx, req)
	assert.Error(t, err)
}

func TestInventoryService_GetInventoryTransactions_Extra(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	supRepo := new(mockSupplierRepo)
	svc := NewInventoryService(invRepo, txRepo, supRepo)
	ctx := context.Background()

	txs := []*entity.InventoryTransaction{{ID: "tx-001"}}
	txRepo.On("ListByInventoryID", ctx, "inv-001").Return(txs, nil)

	result, err := svc.GetInventoryTransactions(ctx, "inv-001")
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestInventoryService_GetTransactionByID_Extra(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	supRepo := new(mockSupplierRepo)
	svc := NewInventoryService(invRepo, txRepo, supRepo)
	ctx := context.Background()

	tx := &entity.InventoryTransaction{ID: "tx-001"}
	txRepo.On("GetByID", ctx, "tx-001").Return(tx, nil)

	result, err := svc.GetTransactionByID(ctx, "tx-001")
	assert.NoError(t, err)
	assert.Equal(t, "tx-001", result.ID)
}

func TestInventoryService_GetTransactionsByReference_Extra(t *testing.T) {
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	supRepo := new(mockSupplierRepo)
	svc := NewInventoryService(invRepo, txRepo, supRepo)
	ctx := context.Background()

	txs := []*entity.InventoryTransaction{{ID: "tx-001"}}
	txRepo.On("ListByReference", ctx, "ref-001", "purchase").Return(txs, nil)

	result, err := svc.GetTransactionsByReference(ctx, "ref-001", "purchase")
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestPurchaseOrderService_Create_Success_Extra(t *testing.T) {
	poRepo := new(mockPORepoExtra)
	supRepo := new(mockSupplierRepo)
	svc := NewPurchaseOrderService(poRepo, supRepo)
	ctx := context.Background()

	supRepo.On("GetByID", ctx, "sup-001").Return(&entity.Supplier{ID: "sup-001"}, nil)
	poRepo.On("Create", ctx, mock.AnythingOfType("*entity.PurchaseOrder")).Return(nil)

	order := &entity.PurchaseOrder{
		SupplierID: "sup-001",
		Items: []*entity.PurchaseOrderItem{
			{Quantity: 10, UnitPrice: 100.0},
			{Quantity: 5, UnitPrice: 200.0},
		},
	}
	err := svc.CreatePurchaseOrder(ctx, order)
	assert.NoError(t, err)
	assert.Equal(t, "pending", order.Status)
	assert.InDelta(t, 2000.0, order.TotalAmount, 0.01)
}

func TestPurchaseOrderService_Create_SupplierNotFound_Extra(t *testing.T) {
	poRepo := new(mockPORepoExtra)
	supRepo := new(mockSupplierRepo)
	svc := NewPurchaseOrderService(poRepo, supRepo)
	ctx := context.Background()

	supRepo.On("GetByID", ctx, "sup-999").Return(nil, assert.AnError)

	order := &entity.PurchaseOrder{SupplierID: "sup-999"}
	err := svc.CreatePurchaseOrder(ctx, order)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "供应商不存在")
}

func TestPurchaseOrderService_Update_Success_Extra(t *testing.T) {
	poRepo := new(mockPORepoExtra)
	supRepo := new(mockSupplierRepo)
	svc := NewPurchaseOrderService(poRepo, supRepo)
	ctx := context.Background()

	existing := &entity.PurchaseOrder{ID: "po-001", SupplierID: "sup-001"}
	poRepo.On("GetByID", ctx, "po-001").Return(existing, nil)
	poRepo.On("Update", ctx, mock.AnythingOfType("*entity.PurchaseOrder")).Return(nil)

	order := &entity.PurchaseOrder{
		ID: "po-001", SupplierID: "sup-001",
		Items: []*entity.PurchaseOrderItem{{Quantity: 10, UnitPrice: 50.0}},
	}
	err := svc.UpdatePurchaseOrder(ctx, order)
	assert.NoError(t, err)
	assert.InDelta(t, 500.0, order.TotalAmount, 0.01)
}

func TestPurchaseOrderService_Update_NotFound_Extra(t *testing.T) {
	poRepo := new(mockPORepoExtra)
	supRepo := new(mockSupplierRepo)
	svc := NewPurchaseOrderService(poRepo, supRepo)
	ctx := context.Background()

	poRepo.On("GetByID", ctx, "po-999").Return(nil, assert.AnError)

	order := &entity.PurchaseOrder{ID: "po-999"}
	err := svc.UpdatePurchaseOrder(ctx, order)
	assert.Error(t, err)
}

func TestPurchaseOrderService_Update_SupplierChanged_Extra(t *testing.T) {
	poRepo := new(mockPORepoExtra)
	supRepo := new(mockSupplierRepo)
	svc := NewPurchaseOrderService(poRepo, supRepo)
	ctx := context.Background()

	existing := &entity.PurchaseOrder{ID: "po-001", SupplierID: "sup-001"}
	poRepo.On("GetByID", ctx, "po-001").Return(existing, nil)
	supRepo.On("GetByID", ctx, "sup-002").Return(&entity.Supplier{ID: "sup-002"}, nil)
	poRepo.On("Update", ctx, mock.AnythingOfType("*entity.PurchaseOrder")).Return(nil)

	order := &entity.PurchaseOrder{ID: "po-001", SupplierID: "sup-002", Items: []*entity.PurchaseOrderItem{}}
	err := svc.UpdatePurchaseOrder(ctx, order)
	assert.NoError(t, err)
}

func TestPurchaseOrderService_Update_SupplierNotFound_Extra(t *testing.T) {
	poRepo := new(mockPORepoExtra)
	supRepo := new(mockSupplierRepo)
	svc := NewPurchaseOrderService(poRepo, supRepo)
	ctx := context.Background()

	existing := &entity.PurchaseOrder{ID: "po-001", SupplierID: "sup-001"}
	poRepo.On("GetByID", ctx, "po-001").Return(existing, nil)
	supRepo.On("GetByID", ctx, "sup-999").Return(nil, assert.AnError)

	order := &entity.PurchaseOrder{ID: "po-001", SupplierID: "sup-999", Items: []*entity.PurchaseOrderItem{}}
	err := svc.UpdatePurchaseOrder(ctx, order)
	assert.Error(t, err)
}

func TestPurchaseOrderService_Delete_Success_Extra(t *testing.T) {
	poRepo := new(mockPORepoExtra)
	supRepo := new(mockSupplierRepo)
	svc := NewPurchaseOrderService(poRepo, supRepo)
	ctx := context.Background()

	existing := &entity.PurchaseOrder{ID: "po-001", Status: "pending"}
	poRepo.On("GetByID", ctx, "po-001").Return(existing, nil)
	poRepo.On("Delete", ctx, "po-001").Return(nil)

	err := svc.DeletePurchaseOrder(ctx, "po-001")
	assert.NoError(t, err)
}

func TestPurchaseOrderService_Delete_NotPending_Extra(t *testing.T) {
	poRepo := new(mockPORepoExtra)
	supRepo := new(mockSupplierRepo)
	svc := NewPurchaseOrderService(poRepo, supRepo)
	ctx := context.Background()

	existing := &entity.PurchaseOrder{ID: "po-001", Status: "approved"}
	poRepo.On("GetByID", ctx, "po-001").Return(existing, nil)

	err := svc.DeletePurchaseOrder(ctx, "po-001")
	assert.Error(t, err)
}

func TestPurchaseOrderService_Delete_NotFound_Extra(t *testing.T) {
	poRepo := new(mockPORepoExtra)
	supRepo := new(mockSupplierRepo)
	svc := NewPurchaseOrderService(poRepo, supRepo)
	ctx := context.Background()

	poRepo.On("GetByID", ctx, "po-999").Return(nil, assert.AnError)

	err := svc.DeletePurchaseOrder(ctx, "po-999")
	assert.Error(t, err)
}

func TestPurchaseOrderService_GetByID_Extra(t *testing.T) {
	poRepo := new(mockPORepoExtra)
	supRepo := new(mockSupplierRepo)
	svc := NewPurchaseOrderService(poRepo, supRepo)
	ctx := context.Background()

	po := &entity.PurchaseOrder{ID: "po-001"}
	poRepo.On("GetByID", ctx, "po-001").Return(po, nil)

	result, err := svc.GetPurchaseOrderByID(ctx, "po-001")
	assert.NoError(t, err)
	assert.Equal(t, "po-001", result.ID)
}

func TestPurchaseOrderService_List_Extra(t *testing.T) {
	poRepo := new(mockPORepoExtra)
	supRepo := new(mockSupplierRepo)
	svc := NewPurchaseOrderService(poRepo, supRepo)
	ctx := context.Background()

	pos := []*entity.PurchaseOrder{{ID: "po-001"}}
	poRepo.On("List", ctx, (*string)(nil), (*string)(nil), (*time.Time)(nil), (*time.Time)(nil), 0, 10).Return(pos, int64(1), nil)

	result, count, err := svc.ListPurchaseOrders(ctx, nil, nil, nil, nil, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, int64(1), count)
}

func TestPurchaseOrderService_GetByCode_Extra(t *testing.T) {
	poRepo := new(mockPORepoExtra)
	supRepo := new(mockSupplierRepo)
	svc := NewPurchaseOrderService(poRepo, supRepo)
	ctx := context.Background()

	po := &entity.PurchaseOrder{ID: "po-001", Code: "PO20240101001"}
	poRepo.On("GetByCode", ctx, "PO20240101001").Return(po, nil)

	result, err := svc.GetPurchaseOrderByCode(ctx, "PO20240101001")
	assert.NoError(t, err)
	assert.Equal(t, "po-001", result.ID)
}

func TestPurchaseOrderService_UpdateStatus_Success_Extra(t *testing.T) {
	poRepo := new(mockPORepoExtra)
	supRepo := new(mockSupplierRepo)
	svc := NewPurchaseOrderService(poRepo, supRepo)
	ctx := context.Background()

	order := &entity.PurchaseOrder{ID: "po-001", Status: "pending"}
	poRepo.On("GetByID", ctx, "po-001").Return(order, nil)
	poRepo.On("Update", ctx, mock.AnythingOfType("*entity.PurchaseOrder")).Return(nil)

	err := svc.UpdatePurchaseOrderStatus(ctx, "po-001", "approved")
	assert.NoError(t, err)
}

func TestPurchaseOrderService_UpdateStatus_InvalidStatus_Extra(t *testing.T) {
	poRepo := new(mockPORepoExtra)
	supRepo := new(mockSupplierRepo)
	svc := NewPurchaseOrderService(poRepo, supRepo)
	ctx := context.Background()

	order := &entity.PurchaseOrder{ID: "po-001", Status: "pending"}
	poRepo.On("GetByID", ctx, "po-001").Return(order, nil)

	err := svc.UpdatePurchaseOrderStatus(ctx, "po-001", "invalid_status")
	assert.Error(t, err)
}

func TestPurchaseOrderService_UpdateStatus_NotFound_Extra(t *testing.T) {
	poRepo := new(mockPORepoExtra)
	supRepo := new(mockSupplierRepo)
	svc := NewPurchaseOrderService(poRepo, supRepo)
	ctx := context.Background()

	poRepo.On("GetByID", ctx, "po-999").Return(nil, assert.AnError)

	err := svc.UpdatePurchaseOrderStatus(ctx, "po-999", "approved")
	assert.Error(t, err)
}

func TestReceiptService_Create_Success_Extra(t *testing.T) {
	rcptRepo := new(mockReceiptRepoExtra)
	poRepo := new(mockPORepoExtra)
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	svc := NewReceiptService(rcptRepo, poRepo, invRepo, txRepo)
	ctx := context.Background()

	poRepo.On("GetByID", ctx, "po-001").Return(&entity.PurchaseOrder{ID: "po-001"}, nil)
	rcptRepo.On("Create", ctx, mock.AnythingOfType("*entity.Receipt")).Return(nil)

	receipt := &entity.Receipt{
		PurchaseOrderID: "po-001",
		Items: []*entity.ReceiptItem{
			{Quantity: 10, UnitPrice: 100.0},
		},
	}
	err := svc.CreateReceipt(ctx, receipt)
	assert.NoError(t, err)
	assert.Equal(t, "pending", receipt.Status)
	assert.Equal(t, 10, receipt.TotalItems)
}

func TestReceiptService_Create_PONotFound_Extra(t *testing.T) {
	rcptRepo := new(mockReceiptRepoExtra)
	poRepo := new(mockPORepoExtra)
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	svc := NewReceiptService(rcptRepo, poRepo, invRepo, txRepo)
	ctx := context.Background()

	poRepo.On("GetByID", ctx, "po-999").Return(nil, assert.AnError)

	receipt := &entity.Receipt{PurchaseOrderID: "po-999"}
	err := svc.CreateReceipt(ctx, receipt)
	assert.Error(t, err)
}

func TestReceiptService_Update_Success_Extra(t *testing.T) {
	rcptRepo := new(mockReceiptRepoExtra)
	poRepo := new(mockPORepoExtra)
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	svc := NewReceiptService(rcptRepo, poRepo, invRepo, txRepo)
	ctx := context.Background()

	existing := &entity.Receipt{ID: "rc-001", PurchaseOrderID: "po-001"}
	rcptRepo.On("GetByID", ctx, "rc-001").Return(existing, nil)
	rcptRepo.On("Update", ctx, mock.AnythingOfType("*entity.Receipt")).Return(nil)

	receipt := &entity.Receipt{ID: "rc-001", PurchaseOrderID: "po-001", Items: []*entity.ReceiptItem{{Quantity: 5}}}
	err := svc.UpdateReceipt(ctx, receipt)
	assert.NoError(t, err)
}

func TestReceiptService_Update_NotFound_Extra(t *testing.T) {
	rcptRepo := new(mockReceiptRepoExtra)
	poRepo := new(mockPORepoExtra)
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	svc := NewReceiptService(rcptRepo, poRepo, invRepo, txRepo)
	ctx := context.Background()

	rcptRepo.On("GetByID", ctx, "rc-999").Return(nil, assert.AnError)

	receipt := &entity.Receipt{ID: "rc-999"}
	err := svc.UpdateReceipt(ctx, receipt)
	assert.Error(t, err)
}

func TestReceiptService_Update_POChanged_Extra(t *testing.T) {
	rcptRepo := new(mockReceiptRepoExtra)
	poRepo := new(mockPORepoExtra)
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	svc := NewReceiptService(rcptRepo, poRepo, invRepo, txRepo)
	ctx := context.Background()

	existing := &entity.Receipt{ID: "rc-001", PurchaseOrderID: "po-001"}
	rcptRepo.On("GetByID", ctx, "rc-001").Return(existing, nil)
	poRepo.On("GetByID", ctx, "po-002").Return(&entity.PurchaseOrder{ID: "po-002"}, nil)
	rcptRepo.On("Update", ctx, mock.AnythingOfType("*entity.Receipt")).Return(nil)

	receipt := &entity.Receipt{ID: "rc-001", PurchaseOrderID: "po-002", Items: []*entity.ReceiptItem{}}
	err := svc.UpdateReceipt(ctx, receipt)
	assert.NoError(t, err)
}

func TestReceiptService_Update_POChangedNotFound_Extra(t *testing.T) {
	rcptRepo := new(mockReceiptRepoExtra)
	poRepo := new(mockPORepoExtra)
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	svc := NewReceiptService(rcptRepo, poRepo, invRepo, txRepo)
	ctx := context.Background()

	existing := &entity.Receipt{ID: "rc-001", PurchaseOrderID: "po-001"}
	rcptRepo.On("GetByID", ctx, "rc-001").Return(existing, nil)
	poRepo.On("GetByID", ctx, "po-999").Return(nil, assert.AnError)

	receipt := &entity.Receipt{ID: "rc-001", PurchaseOrderID: "po-999", Items: []*entity.ReceiptItem{}}
	err := svc.UpdateReceipt(ctx, receipt)
	assert.Error(t, err)
}

func TestReceiptService_Delete_Success_Extra(t *testing.T) {
	rcptRepo := new(mockReceiptRepoExtra)
	poRepo := new(mockPORepoExtra)
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	svc := NewReceiptService(rcptRepo, poRepo, invRepo, txRepo)
	ctx := context.Background()

	existing := &entity.Receipt{ID: "rc-001", Status: "pending"}
	rcptRepo.On("GetByID", ctx, "rc-001").Return(existing, nil)
	rcptRepo.On("Delete", ctx, "rc-001").Return(nil)

	err := svc.DeleteReceipt(ctx, "rc-001")
	assert.NoError(t, err)
}

func TestReceiptService_Delete_NotPending_Extra(t *testing.T) {
	rcptRepo := new(mockReceiptRepoExtra)
	poRepo := new(mockPORepoExtra)
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	svc := NewReceiptService(rcptRepo, poRepo, invRepo, txRepo)
	ctx := context.Background()

	existing := &entity.Receipt{ID: "rc-001", Status: "completed"}
	rcptRepo.On("GetByID", ctx, "rc-001").Return(existing, nil)

	err := svc.DeleteReceipt(ctx, "rc-001")
	assert.Error(t, err)
}

func TestReceiptService_Delete_NotFound_Extra(t *testing.T) {
	rcptRepo := new(mockReceiptRepoExtra)
	poRepo := new(mockPORepoExtra)
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	svc := NewReceiptService(rcptRepo, poRepo, invRepo, txRepo)
	ctx := context.Background()

	rcptRepo.On("GetByID", ctx, "rc-999").Return(nil, assert.AnError)

	err := svc.DeleteReceipt(ctx, "rc-999")
	assert.Error(t, err)
}

func TestReceiptService_GetByID_Extra(t *testing.T) {
	rcptRepo := new(mockReceiptRepoExtra)
	poRepo := new(mockPORepoExtra)
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	svc := NewReceiptService(rcptRepo, poRepo, invRepo, txRepo)
	ctx := context.Background()

	rcpt := &entity.Receipt{ID: "rc-001"}
	rcptRepo.On("GetByID", ctx, "rc-001").Return(rcpt, nil)

	result, err := svc.GetReceiptByID(ctx, "rc-001")
	assert.NoError(t, err)
	assert.Equal(t, "rc-001", result.ID)
}

func TestReceiptService_List_Extra(t *testing.T) {
	rcptRepo := new(mockReceiptRepoExtra)
	poRepo := new(mockPORepoExtra)
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	svc := NewReceiptService(rcptRepo, poRepo, invRepo, txRepo)
	ctx := context.Background()

	rcpts := []*entity.Receipt{{ID: "rc-001"}}
	rcptRepo.On("List", ctx, (*string)(nil), (*string)(nil), (*time.Time)(nil), (*time.Time)(nil), 0, 10).Return(rcpts, int64(1), nil)

	result, count, err := svc.ListReceipts(ctx, nil, nil, nil, nil, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, int64(1), count)
}

func TestReceiptService_GetByCode_Extra(t *testing.T) {
	rcptRepo := new(mockReceiptRepoExtra)
	poRepo := new(mockPORepoExtra)
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	svc := NewReceiptService(rcptRepo, poRepo, invRepo, txRepo)
	ctx := context.Background()

	rcpt := &entity.Receipt{ID: "rc-001", Code: "RC20240101001"}
	rcptRepo.On("GetByCode", ctx, "RC20240101001").Return(rcpt, nil)

	result, err := svc.GetReceiptByCode(ctx, "RC20240101001")
	assert.NoError(t, err)
	assert.Equal(t, "rc-001", result.ID)
}

func TestReceiptService_GetByPurchaseOrderID_Extra(t *testing.T) {
	rcptRepo := new(mockReceiptRepoExtra)
	poRepo := new(mockPORepoExtra)
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	svc := NewReceiptService(rcptRepo, poRepo, invRepo, txRepo)
	ctx := context.Background()

	rcpts := []*entity.Receipt{{ID: "rc-001"}}
	rcptRepo.On("GetByPurchaseOrderID", ctx, "po-001").Return(rcpts, nil)

	result, err := svc.GetReceiptsByPurchaseOrderID(ctx, "po-001")
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestReceiptService_UpdateStatus_Completed_Extra(t *testing.T) {
	rcptRepo := new(mockReceiptRepoExtra)
	poRepo := new(mockPORepoExtra)
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	svc := NewReceiptService(rcptRepo, poRepo, invRepo, txRepo)
	ctx := context.Background()

	inventory := &entity.Inventory{ID: "inv-001", Quantity: 100, UnitPrice: 10.0}
	receipt := &entity.Receipt{
		ID: "rc-001", PurchaseOrderID: "po-001", Status: "pending",
		Items: []*entity.ReceiptItem{
			{InventoryID: "inv-001", Quantity: 10, UnitPrice: 10.0},
		},
	}
	rcptRepo.On("GetByID", ctx, "rc-001").Return(receipt, nil)
	invRepo.On("GetByID", ctx, "inv-001").Return(inventory, nil)
	invRepo.On("Update", ctx, mock.AnythingOfType("*entity.Inventory")).Return(nil)
	txRepo.On("Create", ctx, mock.AnythingOfType("*entity.InventoryTransaction")).Return(nil)
	poRepo.On("GetByID", ctx, "po-001").Return(&entity.PurchaseOrder{ID: "po-001"}, nil)
	poRepo.On("Update", ctx, mock.AnythingOfType("*entity.PurchaseOrder")).Return(nil)
	rcptRepo.On("Update", ctx, mock.AnythingOfType("*entity.Receipt")).Return(nil)

	err := svc.UpdateReceiptStatus(ctx, "rc-001", "completed")
	assert.NoError(t, err)
}

func TestReceiptService_UpdateStatus_InvalidStatus_Extra(t *testing.T) {
	rcptRepo := new(mockReceiptRepoExtra)
	poRepo := new(mockPORepoExtra)
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	svc := NewReceiptService(rcptRepo, poRepo, invRepo, txRepo)
	ctx := context.Background()

	receipt := &entity.Receipt{ID: "rc-001", Status: "pending"}
	rcptRepo.On("GetByID", ctx, "rc-001").Return(receipt, nil)

	err := svc.UpdateReceiptStatus(ctx, "rc-001", "invalid")
	assert.Error(t, err)
}

func TestReceiptService_UpdateStatus_NotFound_Extra(t *testing.T) {
	rcptRepo := new(mockReceiptRepoExtra)
	poRepo := new(mockPORepoExtra)
	invRepo := new(mockInventoryRepo)
	txRepo := new(mockInventoryTransRepo)
	svc := NewReceiptService(rcptRepo, poRepo, invRepo, txRepo)
	ctx := context.Background()

	rcptRepo.On("GetByID", ctx, "rc-999").Return(nil, assert.AnError)

	err := svc.UpdateReceiptStatus(ctx, "rc-999", "completed")
	assert.Error(t, err)
}

func TestWorkOrderService_GetWorkOrdersByDevice_Extra(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	devRepo := new(mockDeviceRepoForWO)
	svc := NewWorkOrderService(woRepo, devRepo)
	ctx := context.Background()

	wos := []*entity.WorkOrder{{ID: "wo-001", DeviceID: "dev-001"}}
	woRepo.On("List", ctx, mock.Anything).Return(wos, nil)

	result, err := svc.GetWorkOrdersByDevice(ctx, "dev-001")
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestWorkOrderService_GetWorkOrderStats_Extra(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	devRepo := new(mockDeviceRepoForWO)
	svc := NewWorkOrderService(woRepo, devRepo)
	ctx := context.Background()

	woRepo.On("Count", ctx, mock.Anything).Return(int64(100), nil)
	woRepo.On("Count", ctx, mock.Anything).Return(int64(30), nil)
	woRepo.On("Count", ctx, mock.Anything).Return(int64(20), nil)
	woRepo.On("Count", ctx, mock.Anything).Return(int64(50), nil)

	stats, err := svc.GetWorkOrderStats(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestWorkOrderService_GetWorkOrderStats_Error_Extra(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	devRepo := new(mockDeviceRepoForWO)
	svc := NewWorkOrderService(woRepo, devRepo)
	ctx := context.Background()

	woRepo.On("Count", ctx, mock.Anything).Return(int64(0), assert.AnError)

	stats, err := svc.GetWorkOrderStats(ctx)
	assert.Error(t, err)
	assert.Nil(t, stats)
}

func TestDeviceHarness_ValidateCreateDevice_InvalidCode_Extra(t *testing.T) {
	dh := NewDeviceHarness()
	ctx := context.Background()

	req := &CreateDeviceRequest{
		Code: "ab", Name: "Test Device", Type: entity.DeviceTypeInverter,
		RatedPower: 100, RatedVoltage: 220, RatedCurrent: 10,
	}
	err := dh.ValidateCreateDevice(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "device code")
}

func TestDeviceHarness_ValidateCreateDevice_EmptyName_Extra(t *testing.T) {
	dh := NewDeviceHarness()
	ctx := context.Background()

	req := &CreateDeviceRequest{
		Code: "DEV001", Name: "   ", Type: entity.DeviceTypeInverter,
		RatedPower: 100, RatedVoltage: 220, RatedCurrent: 10,
	}
	err := dh.ValidateCreateDevice(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "device name")
}

func TestDeviceHarness_ValidateCreateDevice_InvalidType_Extra(t *testing.T) {
	dh := NewDeviceHarness()
	ctx := context.Background()

	req := &CreateDeviceRequest{
		Code: "DEV001", Name: "Test Device", Type: entity.DeviceType("invalid"),
		RatedPower: 100, RatedVoltage: 220, RatedCurrent: 10,
	}
	err := dh.ValidateCreateDevice(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid device type")
}

func TestDeviceHarness_ValidateCreateDevice_NegativePower_Extra(t *testing.T) {
	dh := NewDeviceHarness()
	ctx := context.Background()

	req := &CreateDeviceRequest{
		Code: "DEV001", Name: "Test Device", Type: entity.DeviceTypeInverter,
		RatedPower: -100, RatedVoltage: 220, RatedCurrent: 10,
	}
	err := dh.ValidateCreateDevice(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rated power")
}

func TestDeviceHarness_ValidateCreateDevice_InvalidProtocol_Extra(t *testing.T) {
	dh := NewDeviceHarness()
	ctx := context.Background()

	req := &CreateDeviceRequest{
		Code: "DEV001", Name: "Test Device", Type: entity.DeviceTypeInverter,
		RatedPower: 100, RatedVoltage: 220, RatedCurrent: 10,
		Protocol: "invalid-proto", IPAddress: "192.168.1.1",
	}
	err := dh.ValidateCreateDevice(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported protocol")
}

func TestDeviceHarness_ValidateCreateDevice_InvalidIP_Extra(t *testing.T) {
	dh := NewDeviceHarness()
	ctx := context.Background()

	req := &CreateDeviceRequest{
		Code: "DEV001", Name: "Test Device", Type: entity.DeviceTypeInverter,
		RatedPower: 100, RatedVoltage: 220, RatedCurrent: 10,
		Protocol: "modbus-tcp", IPAddress: "invalid-ip",
	}
	err := dh.ValidateCreateDevice(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid IP address")
}

func TestDeviceHarness_ValidateCreateDevice_InvalidPort_Extra(t *testing.T) {
	dh := NewDeviceHarness()
	ctx := context.Background()

	req := &CreateDeviceRequest{
		Code: "DEV001", Name: "Test Device", Type: entity.DeviceTypeInverter,
		RatedPower: 100, RatedVoltage: 220, RatedCurrent: 10,
		Protocol: "modbus-tcp", IPAddress: "192.168.1.1", Port: 99999,
	}
	err := dh.ValidateCreateDevice(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port")
}

func TestDeviceHarness_ValidateCreateDevice_InvalidSlaveID_Extra(t *testing.T) {
	dh := NewDeviceHarness()
	ctx := context.Background()

	req := &CreateDeviceRequest{
		Code: "DEV001", Name: "Test Device", Type: entity.DeviceTypeInverter,
		RatedPower: 100, RatedVoltage: 220, RatedCurrent: 10,
		Protocol: "modbus-tcp", IPAddress: "192.168.1.1", Port: 502, SlaveID: 300,
	}
	err := dh.ValidateCreateDevice(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "slave ID")
}

func TestDeviceHarness_ValidateUpdateDevice_EmptyID_Extra(t *testing.T) {
	dh := NewDeviceHarness()
	ctx := context.Background()

	req := &UpdateDeviceRequest{}
	err := dh.ValidateUpdateDevice(ctx, "", req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "device ID")
}

func TestDeviceHarness_ValidateDeviceStatus_Invalid_Extra(t *testing.T) {
	dh := NewDeviceHarness()
	ctx := context.Background()

	err := dh.ValidateDeviceStatus(ctx, entity.DeviceStatus(99))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid device status")
}

func TestDeviceHarness_ValidateDeviceQuery_InvalidType_Extra(t *testing.T) {
	dh := NewDeviceHarness()
	ctx := context.Background()

	invalidType := entity.DeviceType("invalid")
	err := dh.ValidateDeviceQuery(ctx, nil, &invalidType)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid device type")
}

func TestAlarmHarness_ValidateCreateAlarm_InvalidLevel_Extra(t *testing.T) {
	ah := NewAlarmHarness()
	ctx := context.Background()

	req := &CreateAlarmRequest{Level: entity.AlarmLevel(99), Type: entity.AlarmTypeLimit, Title: "Test"}
	err := ah.ValidateCreateAlarm(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid alarm level")
}

func TestAlarmHarness_ValidateCreateAlarm_InvalidType_Extra(t *testing.T) {
	ah := NewAlarmHarness()
	ctx := context.Background()

	req := &CreateAlarmRequest{Level: entity.AlarmLevelWarning, Type: entity.AlarmType("invalid"), Title: "Test"}
	err := ah.ValidateCreateAlarm(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid alarm type")
}

func TestAlarmHarness_ValidateCreateAlarm_EmptyTitle_Extra(t *testing.T) {
	ah := NewAlarmHarness()
	ctx := context.Background()

	req := &CreateAlarmRequest{Level: entity.AlarmLevelWarning, Type: entity.AlarmTypeLimit, Title: ""}
	err := ah.ValidateCreateAlarm(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title cannot be empty")
}

func TestAlarmHarness_ValidateCreateAlarm_LimitAlarmNoDeviation_Extra(t *testing.T) {
	ah := NewAlarmHarness()
	ctx := context.Background()

	req := &CreateAlarmRequest{
		Level: entity.AlarmLevelWarning, Type: entity.AlarmTypeLimit,
		Title: "Test", Value: 100.0, Threshold: 100.0,
	}
	err := ah.ValidateCreateAlarm(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deviate")
}

func TestAlarmHarness_ValidateAcknowledgeAlarm_EmptyID_Extra(t *testing.T) {
	ah := NewAlarmHarness()
	ctx := context.Background()

	err := ah.ValidateAcknowledgeAlarm(ctx, "", "admin")
	assert.Error(t, err)
}

func TestAlarmHarness_ValidateAcknowledgeAlarm_EmptyOperator_Extra(t *testing.T) {
	ah := NewAlarmHarness()
	ctx := context.Background()

	err := ah.ValidateAcknowledgeAlarm(ctx, "alarm-001", "")
	assert.Error(t, err)
}

func TestAlarmHarness_ValidateClearAlarm_EmptyID_Extra(t *testing.T) {
	ah := NewAlarmHarness()
	ctx := context.Background()

	err := ah.ValidateClearAlarm(ctx, "")
	assert.Error(t, err)
}

func TestAlarmHarness_ValidateAlarmQuery_InvalidTimeRange_Extra(t *testing.T) {
	ah := NewAlarmHarness()
	ctx := context.Background()

	err := ah.ValidateAlarmQuery(ctx, nil, 2000, 1000)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "start time")
}

func TestAlarmHarness_GetHarness_Extra(t *testing.T) {
	ah := NewAlarmHarness()
	h := ah.GetHarness()
	assert.NotNil(t, h)
}
