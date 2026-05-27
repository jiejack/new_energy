package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkOrderRepository_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewWorkOrderRepository(db)
	ctx := context.Background()

	wo := &entity.WorkOrder{
		ID:          uuid.New().String(),
		DeviceID:    uuid.New().String(),
		Type:        "maintenance",
		Title:       "Test Work Order",
		Description: "Test description",
		Priority:    "high",
		Status:      "open",
		Assignee:    "admin",
		CreatedBy:   "admin",
	}
	err := repo.Create(ctx, wo)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, wo.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Test Work Order", found.Title)
	assert.Equal(t, "open", found.Status)

	wo.Status = "in_progress"
	err = repo.Update(ctx, wo)
	assert.NoError(t, err)

	found, err = repo.GetByID(ctx, wo.ID)
	assert.NoError(t, err)
	assert.Equal(t, "in_progress", found.Status)

	err = repo.Delete(ctx, wo.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(ctx, wo.ID)
	assert.Error(t, err)
}

func TestWorkOrderRepository_List(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewWorkOrderRepository(db)
	ctx := context.Background()

	deviceID := uuid.New().String()
	wo1 := &entity.WorkOrder{
		ID:        uuid.New().String(),
		DeviceID:  deviceID,
		Type:      "maintenance",
		Title:     "WO 1",
		Priority:  "high",
		Status:    "open",
		CreatedBy: "admin",
	}
	wo2 := &entity.WorkOrder{
		ID:        uuid.New().String(),
		DeviceID:  uuid.New().String(),
		Type:      "repair",
		Title:     "WO 2",
		Priority:  "low",
		Status:    "completed",
		CreatedBy: "operator",
	}
	err := repo.Create(ctx, wo1)
	require.NoError(t, err)
	err = repo.Create(ctx, wo2)
	require.NoError(t, err)

	wos, err := repo.List(ctx, nil)
	assert.NoError(t, err)
	assert.Len(t, wos, 2)

	wos, err = repo.List(ctx, &service.WorkOrderFilter{DeviceID: &deviceID})
	assert.NoError(t, err)
	assert.Len(t, wos, 1)
	assert.Equal(t, "WO 1", wos[0].Title)

	woType := "repair"
	wos, err = repo.List(ctx, &service.WorkOrderFilter{Type: &woType})
	assert.NoError(t, err)
	assert.Len(t, wos, 1)

	wos, err = repo.List(ctx, &service.WorkOrderFilter{Status: "open"})
	assert.NoError(t, err)
	assert.Len(t, wos, 1)

	priority := "high"
	wos, err = repo.List(ctx, &service.WorkOrderFilter{Priority: &priority})
	assert.NoError(t, err)
	assert.Len(t, wos, 1)
}

func TestWorkOrderRepository_Count(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewWorkOrderRepository(db)
	ctx := context.Background()

	wo := &entity.WorkOrder{
		ID:        uuid.New().String(),
		DeviceID:  uuid.New().String(),
		Type:      "maintenance",
		Title:     "WO 1",
		Priority:  "high",
		Status:    "open",
		CreatedBy: "admin",
	}
	err := repo.Create(ctx, wo)
	require.NoError(t, err)

	count, err := repo.Count(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	count, err = repo.Count(ctx, &service.WorkOrderFilter{Status: "open"})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	count, err = repo.Count(ctx, &service.WorkOrderFilter{Status: "completed"})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestInventoryRepository_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewInventoryRepository(db)
	ctx := context.Background()

	inv := &entity.Inventory{
		ID:          uuid.New().String(),
		Code:        "INV001",
		Name:        "Test Item",
		Type:        "spare_part",
		Unit:        "pcs",
		Quantity:    100,
		MinQuantity: 10,
		MaxQuantity: 500,
		Status:      "normal",
	}
	err := repo.Create(ctx, inv)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, inv.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Test Item", found.Name)

	found, err = repo.GetByCode(ctx, "INV001")
	assert.NoError(t, err)
	assert.Equal(t, inv.ID, found.ID)

	inv.Quantity = 50
	err = repo.Update(ctx, inv)
	assert.NoError(t, err)

	err = repo.Delete(ctx, inv.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(ctx, inv.ID)
	assert.Error(t, err)
}

func TestInventoryRepository_List(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewInventoryRepository(db)
	ctx := context.Background()

	inv := &entity.Inventory{
		ID:       uuid.New().String(),
		Code:     "INV001",
		Name:     "Test Item",
		Type:     "spare_part",
		Unit:     "pcs",
		Quantity: 100,
		Status:   "normal",
	}
	err := repo.Create(ctx, inv)
	require.NoError(t, err)

	invs, err := repo.List(ctx, nil)
	assert.NoError(t, err)
	assert.Len(t, invs, 1)

	count, err := repo.Count(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestInventoryRepository_UpdateQuantity(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewInventoryRepository(db)
	ctx := context.Background()

	inv := &entity.Inventory{
		ID:       uuid.New().String(),
		Code:     "INV001",
		Name:     "Test Item",
		Type:     "spare_part",
		Unit:     "pcs",
		Quantity: 100,
		Status:   "normal",
	}
	err := repo.Create(ctx, inv)
	require.NoError(t, err)

	err = repo.UpdateQuantity(ctx, inv.ID, 75)
	assert.NoError(t, err)

	found, err := repo.GetByID(ctx, inv.ID)
	assert.NoError(t, err)
	assert.Equal(t, 75.0, found.Quantity)
}

func TestInventoryRepository_GetLowStockItems(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewInventoryRepository(db)
	ctx := context.Background()

	inv := &entity.Inventory{
		ID:          uuid.New().String(),
		Code:        "INV001",
		Name:        "Low Stock Item",
		Type:        "spare_part",
		Unit:        "pcs",
		Quantity:    5,
		MinQuantity: 10,
		Status:      "low_stock",
	}
	err := repo.Create(ctx, inv)
	require.NoError(t, err)

	items, err := repo.GetLowStockItems(ctx)
	assert.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestSupplierRepository_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewSupplierRepository(db)
	ctx := context.Background()

	sup := &entity.Supplier{
		ID:           uuid.New().String(),
		Code:         "SUP001",
		Name:         "Test Supplier",
		ContactName:  "John",
		ContactPhone: "13800138000",
		ContactEmail: "john@supplier.com",
		Status:       "active",
	}
	err := repo.Create(ctx, sup)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, sup.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Test Supplier", found.Name)

	found, err = repo.GetByCode(ctx, "SUP001")
	assert.NoError(t, err)
	assert.Equal(t, sup.ID, found.ID)

	sup.Name = "Updated Supplier"
	err = repo.Update(ctx, sup)
	assert.NoError(t, err)

	err = repo.Delete(ctx, sup.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(ctx, sup.ID)
	assert.Error(t, err)
}

func TestSupplierRepository_List(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewSupplierRepository(db)
	ctx := context.Background()

	sup := &entity.Supplier{
		ID:     uuid.New().String(),
		Code:   "SUP001",
		Name:   "Test Supplier",
		Status: "active",
	}
	err := repo.Create(ctx, sup)
	require.NoError(t, err)

	sups, err := repo.List(ctx, nil)
	assert.NoError(t, err)
	assert.Len(t, sups, 1)

	count, err := repo.Count(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestEdgeNodeRepository_CRUD(t *testing.T) {
	t.Skip("Skipping: MapJSON type incompatible with SQLite")
}

func TestEdgeNodeRepository_List(t *testing.T) {
	t.Skip("Skipping: MapJSON type incompatible with SQLite")
}

func TestEdgeNodeRepository_UpdateStatus(t *testing.T) {
	t.Skip("Skipping: MapJSON type incompatible with SQLite")
}

func TestEdgeNodeRepository_UpdateHeartbeat(t *testing.T) {
	t.Skip("Skipping: MapJSON type incompatible with SQLite")
}

func TestForecastResultRepository_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewForecastResultRepository(db)
	ctx := context.Background()

	result := entity.NewForecastResult(
		"station-001",
		entity.ForecastTypeShortTerm,
		time.Now().Add(24*time.Hour),
		1500.0,
		"v1.0.0",
	)
	result.ID = uuid.New().String()
	err := repo.Create(ctx, result)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, result.ID)
	assert.NoError(t, err)
	assert.Equal(t, "station-001", found.StationID)
	assert.Equal(t, 1500.0, found.PredictedPower)
}

func TestForecastResultRepository_ListByStation(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewForecastResultRepository(db)
	ctx := context.Background()

	stationID := "station-001"
	r1 := entity.NewForecastResult(stationID, entity.ForecastTypeShortTerm, time.Now().Add(24*time.Hour), 1500.0, "v1.0.0")
	r1.ID = uuid.New().String()
	r2 := entity.NewForecastResult(stationID, entity.ForecastTypeUltraShortTerm, time.Now().Add(1*time.Hour), 800.0, "v1.0.0")
	r2.ID = uuid.New().String()
	err := repo.Create(ctx, r1)
	require.NoError(t, err)
	err = repo.Create(ctx, r2)
	require.NoError(t, err)

	results, total, err := repo.ListByStation(ctx, stationID, nil, nil, nil, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, results, 2)

	ft := entity.ForecastTypeShortTerm
	results, total, err = repo.ListByStation(ctx, stationID, &ft, nil, nil, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, results, 1)
}

func TestForecastResultRepository_UpdateActualPower(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewForecastResultRepository(db)
	ctx := context.Background()

	result := entity.NewForecastResult("station-001", entity.ForecastTypeShortTerm, time.Now().Add(24*time.Hour), 1500.0, "v1.0.0")
	result.ID = uuid.New().String()
	err := repo.Create(ctx, result)
	require.NoError(t, err)

	err = repo.UpdateActualPower(ctx, result.ID, 1450.0)
	assert.NoError(t, err)

	found, err := repo.GetByID(ctx, result.ID)
	assert.NoError(t, err)
	if assert.NotNil(t, found.ActualPower) {
		assert.InDelta(t, 1450.0, *found.ActualPower, 0.1)
	}
}

func TestForecastResultRepository_GetAccuracyStats(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewForecastResultRepository(db)
	ctx := context.Background()

	stationID := "station-001"
	r1 := entity.NewForecastResult(stationID, entity.ForecastTypeShortTerm, time.Now().Add(24*time.Hour), 1500.0, "v1.0.0")
	r1.ID = uuid.New().String()
	r1.SetActualPower(1450.0)
	acc1 := 96.7
	r1.Accuracy = &acc1
	err := repo.Create(ctx, r1)
	require.NoError(t, err)

	stats, err := repo.GetAccuracyStats(ctx, stationID, nil, time.Now().Add(-48*time.Hour), time.Now().Add(48*time.Hour))
	assert.NoError(t, err)
	assert.Equal(t, stationID, stats.StationID)
	assert.Equal(t, int64(1), stats.TotalPoints)
}

func TestForecastResultRepository_GetAccuracyStats_Empty(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewForecastResultRepository(db)
	ctx := context.Background()

	stats, err := repo.GetAccuracyStats(ctx, "nonexistent", nil, time.Now().Add(-48*time.Hour), time.Now().Add(48*time.Hour))
	assert.NoError(t, err)
	assert.Equal(t, int64(0), stats.TotalPoints)
}

func TestFaultDetectionResultRepository_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewFaultDetectionResultRepository(db)
	ctx := context.Background()

	result := entity.NewFaultDetectionResult(uuid.New().String(), "overheating", entity.FaultSeverityCritical, 0.95, "Device temperature exceeded threshold", "v1.0.0")
	result.ID = uuid.New().String()
	err := repo.Create(ctx, result)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, result.ID)
	assert.NoError(t, err)
	assert.Equal(t, "overheating", found.FaultType)

	err = repo.UpdateStatus(ctx, result.ID, entity.FaultDetectionStatusConfirmed)
	assert.NoError(t, err)

	err = repo.UpdateRootCause(ctx, result.ID, "Fan failure")
	assert.NoError(t, err)

	err = repo.LinkWorkOrder(ctx, result.ID, uuid.New().String())
	assert.NoError(t, err)
}

func TestFaultDetectionResultRepository_List(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewFaultDetectionResultRepository(db)
	ctx := context.Background()

	deviceID := uuid.New().String()
	r1 := entity.NewFaultDetectionResult(deviceID, "overheating", entity.FaultSeverityCritical, 0.95, "High temp", "v1.0.0")
	r1.ID = uuid.New().String()
	err := repo.Create(ctx, r1)
	require.NoError(t, err)

	results, total, err := repo.ListByDevice(ctx, deviceID, nil, nil, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, results, 1)

	severity := entity.FaultSeverityCritical
	results, total, err = repo.ListByDevice(ctx, deviceID, &severity, nil, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)

	counts, err := repo.CountBySeverity(ctx, nil)
	assert.NoError(t, err)
	assert.NotNil(t, counts)
}

func TestModelVersionRepository_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewModelVersionRepository(db)
	ctx := context.Background()

	mv := &entity.ModelVersion{
		ID:           uuid.New().String(),
		ModelName:    "power_forecast",
		Version:      "v1.0.0",
		Status:       entity.ModelStatusProd,
		ArtifactPath: "/models/v1.0.0.pkl",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := repo.Create(ctx, mv)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, mv.ID)
	assert.NoError(t, err)
	assert.Equal(t, "power_forecast", found.ModelName)

	prod, err := repo.GetProductionModel(ctx, "power_forecast")
	assert.NoError(t, err)
	assert.Equal(t, mv.ID, prod.ID)

	versions, err := repo.ListByModel(ctx, "power_forecast")
	assert.NoError(t, err)
	assert.Len(t, versions, 1)

	err = repo.UpdateStatus(ctx, mv.ID, entity.ModelStatusRetired)
	assert.NoError(t, err)
}

func TestCostCategoryRepository_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewCostCategoryRepository(db)
	ctx := context.Background()

	cat := &entity.CostCategory{
		ID:       uuid.New().String(),
		Code:     "CAT001",
		Name:     "Operations",
		Type:     "direct",
		ParentID: nil,
		Status:   "active",
	}
	err := repo.Create(ctx, cat)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, cat.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Operations", found.Name)

	found, err = repo.GetByCode(ctx, "CAT001")
	assert.NoError(t, err)
	assert.Equal(t, cat.ID, found.ID)

	cat.Name = "Updated Operations"
	err = repo.Update(ctx, cat)
	assert.NoError(t, err)

	cats, err := repo.List(ctx, nil, nil)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(cats), 1)

	tree, err := repo.GetTree(ctx)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(tree), 1)

	err = repo.Delete(ctx, cat.ID)
	assert.NoError(t, err)
}

func TestCostEntryRepository_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewCostEntryRepository(db)
	ctx := context.Background()

	catID := uuid.New().String()
	entry := &entity.CostEntry{
		ID:             uuid.New().String(),
		Code:           "CE001",
		Date:           time.Now(),
		CostCategoryID: catID,
		Amount:         5000.00,
		Currency:       "CNY",
		ApprovalStatus: "approved",
	}
	err := repo.Create(ctx, entry)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, entry.ID)
	assert.NoError(t, err)
	assert.Equal(t, 5000.00, found.Amount)

	found, err = repo.GetByCode(ctx, "CE001")
	assert.NoError(t, err)
	assert.Equal(t, entry.ID, found.ID)

	entries, total, err := repo.List(ctx, nil, nil, nil, nil, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, entries, 1)

	totalAmount, err := repo.GetTotalByCategory(ctx, catID, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, 5000.00, totalAmount)

	totalAmount, err = repo.GetTotalByPeriod(ctx, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, 5000.00, totalAmount)
}

func TestCostAllocationRepository_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewCostAllocationRepository(db)
	ctx := context.Background()

	alloc := &entity.CostAllocation{
		ID:          uuid.New().String(),
		CostEntryID: uuid.New().String(),
		AllocatedTo: "station",
		AllocatedID: uuid.New().String(),
		Amount:      2500.00,
		Percentage:  50.0,
	}
	err := repo.Create(ctx, alloc)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, alloc.ID)
	assert.NoError(t, err)
	assert.Equal(t, 2500.00, found.Amount)

	allocs, err := repo.ListByCostEntryID(ctx, alloc.CostEntryID)
	assert.NoError(t, err)
	assert.Len(t, allocs, 1)

	allocs, err = repo.ListByAllocated(ctx, "station", alloc.AllocatedID)
	assert.NoError(t, err)
	assert.Len(t, allocs, 1)

	total, err := repo.GetTotalByAllocated(ctx, "station", alloc.AllocatedID, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, 2500.00, total)
}

func TestCostReportRepository_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewCostReportRepository(db)
	ctx := context.Background()

	report := &entity.CostReport{
		ID:          uuid.New().String(),
		Code:        "CR001",
		Name:        "Monthly Report",
		ReportType:  "monthly",
		PeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
		TotalCost:   50000.00,
		Status:      "draft",
		GeneratedBy: "admin",
	}
	err := repo.Create(ctx, report)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, report.ID)
	assert.NoError(t, err)
	assert.Equal(t, "monthly", found.ReportType)

	found, err = repo.GetByCode(ctx, "CR001")
	assert.NoError(t, err)
	assert.Equal(t, report.ID, found.ID)

	reports, total, err := repo.List(ctx, nil, nil, nil, nil, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, reports, 1)

	found, err = repo.GetByPeriod(ctx, "monthly", report.PeriodStart, report.PeriodEnd)
	assert.NoError(t, err)
	assert.Equal(t, report.ID, found.ID)
}

func TestPurchaseOrderRepository_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewPurchaseOrderRepository(db)
	ctx := context.Background()

	supplierID := uuid.New().String()
	po := &entity.PurchaseOrder{
		ID:          uuid.New().String(),
		Code:        "PO001",
		SupplierID:  supplierID,
		OrderDate:   time.Now(),
		Status:      "draft",
		TotalAmount: 10000.00,
		CreatedBy:   "admin",
	}
	err := repo.Create(ctx, po)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, po.ID)
	assert.NoError(t, err)
	assert.Equal(t, "PO001", found.Code)

	found, err = repo.GetByCode(ctx, "PO001")
	assert.NoError(t, err)
	assert.Equal(t, po.ID, found.ID)

	pos, total, err := repo.List(ctx, &supplierID, nil, nil, nil, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, pos, 1)
}

func TestReceiptRepository_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewReceiptRepository(db)
	ctx := context.Background()

	poID := uuid.New().String()
	receipt := &entity.Receipt{
		ID:              uuid.New().String(),
		Code:            "RC001",
		PurchaseOrderID: poID,
		ReceiptDate:     time.Now(),
		ReceivedBy:      "admin",
		Status:          "draft",
	}
	err := repo.Create(ctx, receipt)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, receipt.ID)
	assert.NoError(t, err)
	assert.Equal(t, "RC001", found.Code)

	found, err = repo.GetByCode(ctx, "RC001")
	assert.NoError(t, err)
	assert.Equal(t, receipt.ID, found.ID)

	receipts, total, err := repo.List(ctx, &poID, nil, nil, nil, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	_ = receipts

	foundList, err := repo.GetByPurchaseOrderID(ctx, poID)
	assert.NoError(t, err)
	assert.Len(t, foundList, 1)
}

func TestInventoryTransactionRepository_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewInventoryTransactionRepository(db)
	ctx := context.Background()

	invID := uuid.New().String()
	tx := &entity.InventoryTransaction{
		ID:          uuid.New().String(),
		InventoryID: invID,
		Type:        "in",
		Quantity:    50,
		UnitPrice:   10.0,
		TotalAmount: 500.0,
		BeforeQty:   100,
		AfterQty:    150,
		OperatorID:  "admin",
	}
	err := repo.Create(ctx, tx)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, tx.ID)
	assert.NoError(t, err)
	assert.Equal(t, "in", found.Type)

	txs, err := repo.ListByInventoryID(ctx, invID)
	assert.NoError(t, err)
	assert.Len(t, txs, 1)

	refID := uuid.New().String()
	tx2 := &entity.InventoryTransaction{
		ID:            uuid.New().String(),
		InventoryID:   uuid.New().String(),
		Type:          "out",
		Quantity:      20,
		ReferenceID:   refID,
		ReferenceType: "work_order",
		OperatorID:    "admin",
	}
	err = repo.Create(ctx, tx2)
	require.NoError(t, err)

	txs, err = repo.ListByReference(ctx, refID, "work_order")
	assert.NoError(t, err)
	assert.Len(t, txs, 1)

	history, err := repo.GetTransactionHistory(ctx, invID, 10)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(history), 1)
}

func TestEnergyEfficiencyRepository(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewEnergyEfficiencyRepository(db)
	ctx := context.Background()

	records, total, err := repo.ListRecords(ctx, &repository.EnergyEfficiencyQuery{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, records, 0)

	analyses, total, err := repo.ListAnalyses(ctx, &repository.EnergyEfficiencyAnalysisQuery{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, analyses, 0)
}

func TestCarbonEmissionRepository(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewCarbonEmissionRepository(db)
	ctx := context.Background()

	factors, total, err := repo.ListFactors(ctx, &repository.CarbonEmissionFactorQuery{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, factors, 0)

	records, total, err := repo.ListRecords(ctx, &repository.CarbonEmissionRecordQuery{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, records, 0)

	summaries, total, err := repo.ListSummaries(ctx, &repository.CarbonEmissionSummaryQuery{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, summaries, 0)
}

func TestAlarmRuleRepository_GetByName_NotFound(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewAlarmRuleRepository(db)
	ctx := context.Background()

	_, err := repo.GetByName(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestAlarmRuleRepository_List_Pagination(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewAlarmRuleRepository(db)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		rule := entity.NewAlarmRule("Rule "+string(rune('A'+i)), entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "value > threshold")
		rule.ID = uuid.New().String()
		err := repo.Create(ctx, rule)
		require.NoError(t, err)
	}

	rules, total, err := repo.List(ctx, &repository.AlarmRuleQuery{Page: 1, PageSize: 3})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, rules, 3)

	rules, total, err = repo.List(ctx, &repository.AlarmRuleQuery{Page: 2, PageSize: 3})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, rules, 2)
}

func TestQARepository_GetSessionByID_NotFound(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewQARepository(db)
	ctx := context.Background()

	_, err := repo.GetSessionByID(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestQARepository_GetSessionWithMessages_NotFound(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewQARepository(db)
	ctx := context.Background()

	_, err := repo.GetSessionWithMessages(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestQARepository_GetRecentMessages_Limit(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewQARepository(db)
	ctx := context.Background()

	session := entity.NewQASession("user-001", "Test Session")
	err := repo.CreateSession(ctx, session)
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		msg := entity.NewQAMessage(session.ID, entity.QAMessageRoleUser, "Message "+string(rune('A'+i)))
		err = repo.CreateMessage(ctx, msg)
		require.NoError(t, err)
	}

	recent, err := repo.GetRecentMessages(ctx, session.ID, 3)
	assert.NoError(t, err)
	assert.Len(t, recent, 3)
}

func TestSystemConfigRepository_GetByKey_NotFound(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewSystemConfigRepository(db)
	ctx := context.Background()

	_, err := repo.GetByKey(ctx, "nonexistent", "nonexistent")
	assert.Error(t, err)
}

func TestSystemConfigRepository_Delete_NotFound(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewSystemConfigRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, "nonexistent")
	assert.NoError(t, err)
}
