package persistence

import (
	"context"
	"testing"
	"testing/fstest"
	"time"

	"github.com/google/uuid"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCompTestDB(t *testing.T) *Database {
	t.Helper()
	db := setupTestDB(t)
	err := db.AutoMigrate()
	require.NoError(t, err)
	err = db.DB.AutoMigrate(
		&entity.User{},
		&entity.Role{},
		&entity.Permission{},
		&entity.UserRole{},
		&entity.RolePermission{},
		&entity.NotificationConfig{},
		&entity.OperationLog{},
		&entity.WorkOrder{},
		&entity.Inventory{},
		&entity.InventoryTransaction{},
		&entity.Supplier{},
		&entity.PurchaseOrder{},
		&entity.PurchaseOrderItem{},
		&entity.Receipt{},
		&entity.ReceiptItem{},
		&entity.EdgeNode{},
		&entity.ForecastResult{},
		&entity.FaultDetectionResult{},
		&entity.ModelVersion{},
		&entity.CostCategory{},
		&entity.CostEntry{},
		&entity.CostAllocation{},
		&entity.CostReport{},
		&entity.Asset{},
		&entity.AssetMaintenanceRecord{},
		&entity.AssetDepreciationRecord{},
		&entity.AssetDocument{},
		&entity.EnergyEfficiencyRecord{},
		&entity.EnergyEfficiencyAnalysis{},
		&entity.CarbonEmissionFactor{},
		&entity.CarbonEmissionRecord{},
		&entity.CarbonEmissionSummary{},
		&entity.CarbonReductionTarget{},
	)
	require.NoError(t, err)
	return db
}

func TestComp_InventoryRepository_Update(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewInventoryRepository(db)

	inv := &entity.Inventory{
		ID:        uuid.New().String(),
		Code:      "INV-UPD-COMP",
		Name:      "Original",
		Type:      "raw_material",
		Unit:      "kg",
		Quantity:  50,
		Status:    "normal",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.Create(ctx, inv)

	inv.Name = "Updated"
	inv.Quantity = 75
	err := repo.Update(ctx, inv)
	require.NoError(t, err)

	got, _ := repo.GetByID(ctx, inv.ID)
	assert.Equal(t, "Updated", got.Name)
	assert.Equal(t, 75.0, got.Quantity)
}

func TestComp_InventoryRepository_GetByCode(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewInventoryRepository(db)

	inv := &entity.Inventory{
		ID:        uuid.New().String(),
		Code:      "INV-CODE-COMP",
		Name:      "Code Item",
		Type:      "spare_part",
		Unit:      "pcs",
		Quantity:  10,
		Status:    "normal",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.Create(ctx, inv)

	got, err := repo.GetByCode(ctx, "INV-CODE-COMP")
	require.NoError(t, err)
	assert.Equal(t, inv.ID, got.ID)
}

func TestComp_InventoryRepository_List(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewInventoryRepository(db)

	for i := 0; i < 3; i++ {
		inv := &entity.Inventory{
			ID:        uuid.New().String(),
			Code:      "INV-LIST-COMP-" + string(rune('A'+i)),
			Name:      "Item " + string(rune('A'+i)),
			Type:      "spare_part",
			Unit:      "pcs",
			Quantity:  10,
			Status:    "normal",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo.Create(ctx, inv)
	}

	items, err := repo.List(ctx, nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(items), 3)
}

func TestComp_InventoryRepository_Count(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewInventoryRepository(db)

	inv := &entity.Inventory{
		ID:        uuid.New().String(),
		Code:      "INV-COUNT-COMP",
		Name:      "Count Item",
		Type:      "consumable",
		Unit:      "pcs",
		Quantity:  10,
		Status:    "normal",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.Create(ctx, inv)

	count, err := repo.Count(ctx, nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(1))
}

func TestComp_InventoryRepository_UpdateQuantity(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewInventoryRepository(db)

	inv := &entity.Inventory{
		ID:        uuid.New().String(),
		Code:      "INV-QTY-COMP",
		Name:      "Qty Item",
		Type:      "spare_part",
		Unit:      "pcs",
		Quantity:  50,
		Status:    "normal",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.Create(ctx, inv)

	err := repo.UpdateQuantity(ctx, inv.ID, 75.0)
	require.NoError(t, err)

	got, _ := repo.GetByID(ctx, inv.ID)
	assert.Equal(t, 75.0, got.Quantity)
}

func TestComp_InventoryRepository_GetLowStockItems(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewInventoryRepository(db)

	inv := &entity.Inventory{
		ID:          uuid.New().String(),
		Code:        "INV-LOW-COMP",
		Name:        "Low Stock Item",
		Type:        "spare_part",
		Unit:        "pcs",
		Quantity:    5,
		MinQuantity: 10,
		Status:      "low_stock",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	repo.Create(ctx, inv)

	items, err := repo.GetLowStockItems(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(items), 1)
}

func TestComp_InventoryTransactionRepository_ListByInventoryID(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewInventoryTransactionRepository(db)

	invID := uuid.New().String()
	for i := 0; i < 3; i++ {
		tx := &entity.InventoryTransaction{
			ID:          uuid.New().String(),
			InventoryID: invID,
			Type:        "in",
			Quantity:    10,
			OperatorID:  uuid.New().String(),
			CreatedAt:   time.Now(),
		}
		repo.Create(ctx, tx)
	}

	txs, err := repo.ListByInventoryID(ctx, invID)
	require.NoError(t, err)
	assert.Equal(t, 3, len(txs))
}

func TestComp_InventoryTransactionRepository_ListByReference(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewInventoryTransactionRepository(db)

	refID := "ref-comp-001"
	tx := &entity.InventoryTransaction{
		ID:            uuid.New().String(),
		InventoryID:   uuid.New().String(),
		Type:          "out",
		Quantity:      20,
		ReferenceID:   refID,
		ReferenceType: "work_order",
		OperatorID:    uuid.New().String(),
		CreatedAt:     time.Now(),
	}
	repo.Create(ctx, tx)

	txs, err := repo.ListByReference(ctx, refID, "work_order")
	require.NoError(t, err)
	assert.Equal(t, 1, len(txs))
}

func TestComp_InventoryTransactionRepository_GetTransactionHistory(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewInventoryTransactionRepository(db)

	invID := uuid.New().String()
	for i := 0; i < 5; i++ {
		tx := &entity.InventoryTransaction{
			ID:          uuid.New().String(),
			InventoryID: invID,
			Type:        "in",
			Quantity:    10,
			OperatorID:  uuid.New().String(),
			CreatedAt:   time.Now(),
		}
		repo.Create(ctx, tx)
	}

	txs, err := repo.GetTransactionHistory(ctx, invID, 3)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(txs), 3)
}

func TestComp_SupplierRepository_Update(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewSupplierRepository(db)

	sup := &entity.Supplier{
		ID:        uuid.New().String(),
		Code:      "SUP-UPD-COMP",
		Name:      "Original",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.Create(ctx, sup)

	sup.Name = "Updated"
	err := repo.Update(ctx, sup)
	require.NoError(t, err)

	got, _ := repo.GetByID(ctx, sup.ID)
	assert.Equal(t, "Updated", got.Name)
}

func TestComp_SupplierRepository_GetByCode(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewSupplierRepository(db)

	sup := &entity.Supplier{
		ID:        uuid.New().String(),
		Code:      "SUP-CODE-COMP",
		Name:      "Code Supplier",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.Create(ctx, sup)

	got, err := repo.GetByCode(ctx, "SUP-CODE-COMP")
	require.NoError(t, err)
	assert.Equal(t, sup.ID, got.ID)
}

func TestComp_SupplierRepository_List(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewSupplierRepository(db)

	for i := 0; i < 3; i++ {
		sup := &entity.Supplier{
			ID:        uuid.New().String(),
			Code:      "SUP-LIST-COMP-" + string(rune('A'+i)),
			Name:      "Supplier " + string(rune('A'+i)),
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo.Create(ctx, sup)
	}

	suppliers, err := repo.List(ctx, nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(suppliers), 3)
}

func TestComp_SupplierRepository_Count(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewSupplierRepository(db)

	sup := &entity.Supplier{
		ID:        uuid.New().String(),
		Code:      "SUP-COUNT-COMP",
		Name:      "Count Supplier",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.Create(ctx, sup)

	count, err := repo.Count(ctx, nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(1))
}

func TestComp_PurchaseOrderRepository_Update(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewPurchaseOrderRepository(db)

	po := &entity.PurchaseOrder{
		ID:         uuid.New().String(),
		Code:       "PO-UPD-COMP",
		SupplierID: uuid.New().String(),
		OrderDate:  time.Now(),
		Status:     "draft",
		CreatedBy:  "admin",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	repo.Create(ctx, po)

	po.Status = "approved"
	err := repo.Update(ctx, po)
	require.NoError(t, err)

	got, _ := repo.GetByID(ctx, po.ID)
	assert.Equal(t, "approved", got.Status)
}

func TestComp_PurchaseOrderRepository_List(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewPurchaseOrderRepository(db)

	supplierID := uuid.New().String()
	for i := 0; i < 3; i++ {
		po := &entity.PurchaseOrder{
			ID:         uuid.New().String(),
			Code:       "PO-LIST-COMP-" + string(rune('A'+i)),
			SupplierID: supplierID,
			OrderDate:  time.Now(),
			Status:     "draft",
			CreatedBy:  "admin",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		repo.Create(ctx, po)
	}

	orders, count, err := repo.List(ctx, &supplierID, nil, nil, nil, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
	assert.Equal(t, 3, len(orders))
}

func TestComp_PurchaseOrderRepository_GetByCode(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewPurchaseOrderRepository(db)

	po := &entity.PurchaseOrder{
		ID:         uuid.New().String(),
		Code:       "PO-CODE-COMP",
		SupplierID: uuid.New().String(),
		OrderDate:  time.Now(),
		Status:     "draft",
		CreatedBy:  "admin",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	repo.Create(ctx, po)

	got, err := repo.GetByCode(ctx, "PO-CODE-COMP")
	require.NoError(t, err)
	assert.Equal(t, po.ID, got.ID)
}

func TestComp_ReceiptRepository_Update(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewReceiptRepository(db)

	rcpt := &entity.Receipt{
		ID:              uuid.New().String(),
		Code:            "RCPT-UPD-COMP",
		PurchaseOrderID: uuid.New().String(),
		ReceiptDate:     time.Now(),
		ReceivedBy:      "user1",
		Status:          "draft",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	repo.Create(ctx, rcpt)

	rcpt.Status = "completed"
	err := repo.Update(ctx, rcpt)
	require.NoError(t, err)

	got, _ := repo.GetByID(ctx, rcpt.ID)
	assert.Equal(t, "completed", got.Status)
}

func TestComp_ReceiptRepository_List(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewReceiptRepository(db)

	poID := uuid.New().String()
	for i := 0; i < 3; i++ {
		rcpt := &entity.Receipt{
			ID:              uuid.New().String(),
			Code:            "RCPT-LIST-COMP-" + string(rune('A'+i)),
			PurchaseOrderID: poID,
			ReceiptDate:     time.Now(),
			ReceivedBy:      "user1",
			Status:          "draft",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		repo.Create(ctx, rcpt)
	}

	receipts, count, err := repo.List(ctx, &poID, nil, nil, nil, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
	assert.Equal(t, 3, len(receipts))
}

func TestComp_ReceiptRepository_GetByCode(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewReceiptRepository(db)

	rcpt := &entity.Receipt{
		ID:              uuid.New().String(),
		Code:            "RCPT-CODE-COMP",
		PurchaseOrderID: uuid.New().String(),
		ReceiptDate:     time.Now(),
		ReceivedBy:      "user1",
		Status:          "draft",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	repo.Create(ctx, rcpt)

	got, err := repo.GetByCode(ctx, "RCPT-CODE-COMP")
	require.NoError(t, err)
	assert.Equal(t, rcpt.ID, got.ID)
}

func TestComp_ReceiptRepository_GetByPurchaseOrderID(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewReceiptRepository(db)

	poID := uuid.New().String()
	rcpt := &entity.Receipt{
		ID:              uuid.New().String(),
		Code:            "RCPT-PO-COMP",
		PurchaseOrderID: poID,
		ReceiptDate:     time.Now(),
		ReceivedBy:      "user1",
		Status:          "completed",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	repo.Create(ctx, rcpt)

	receipts, err := repo.GetByPurchaseOrderID(ctx, poID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(receipts))
}

func TestComp_CostEntryRepository_GetTotalByCategory(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	catRepo := NewCostCategoryRepository(db)
	repo := NewCostEntryRepository(db)

	cat := &entity.CostCategory{
		ID:        uuid.New().String(),
		Code:      "CE-TOTAL-CAT-COMP",
		Name:      "Category",
		Type:      "direct",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	catRepo.Create(ctx, cat)

	for i := 0; i < 3; i++ {
		entry := &entity.CostEntry{
			ID:             uuid.New().String(),
			Code:           "CE-TOTAL-COMP-" + string(rune('A'+i)),
			Date:           time.Now(),
			CostCategoryID: cat.ID,
			Amount:         100.0,
			Currency:       "CNY",
			ApprovalStatus: "approved",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		repo.Create(ctx, entry)
	}

	total, err := repo.GetTotalByCategory(ctx, cat.ID, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 300.0, total)
}

func TestComp_CostEntryRepository_GetTotalByPeriod(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	catRepo := NewCostCategoryRepository(db)
	repo := NewCostEntryRepository(db)

	cat := &entity.CostCategory{
		ID:        uuid.New().String(),
		Code:      "CE-PERIOD-CAT-COMP",
		Name:      "Category",
		Type:      "direct",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	catRepo.Create(ctx, cat)

	entry := &entity.CostEntry{
		ID:             uuid.New().String(),
		Code:           "CE-PERIOD-COMP",
		Date:           time.Now(),
		CostCategoryID: cat.ID,
		Amount:         500.0,
		Currency:       "CNY",
		ApprovalStatus: "approved",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	repo.Create(ctx, entry)

	start := time.Now().Add(-24 * time.Hour)
	end := time.Now().Add(24 * time.Hour)
	total, err := repo.GetTotalByPeriod(ctx, &start, &end)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 500.0)
}

func TestComp_CostAllocationRepository_ListByAllocated(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	catRepo := NewCostCategoryRepository(db)
	entryRepo := NewCostEntryRepository(db)
	repo := NewCostAllocationRepository(db)

	cat := &entity.CostCategory{
		ID:        uuid.New().String(),
		Code:      "CA-ALLOC-CAT-COMP",
		Name:      "Category",
		Type:      "direct",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	catRepo.Create(ctx, cat)

	entry := &entity.CostEntry{
		ID:             uuid.New().String(),
		Code:           "CA-ALLOC-ENTRY-COMP",
		Date:           time.Now(),
		CostCategoryID: cat.ID,
		Amount:         1000.0,
		Currency:       "CNY",
		ApprovalStatus: "approved",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	entryRepo.Create(ctx, entry)

	allocatedID := uuid.New().String()
	alloc := &entity.CostAllocation{
		ID:          uuid.New().String(),
		CostEntryID: entry.ID,
		AllocatedTo: "department",
		AllocatedID: allocatedID,
		Amount:      600.0,
		Percentage:  60.0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	repo.Create(ctx, alloc)

	allocs, err := repo.ListByAllocated(ctx, "department", allocatedID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(allocs))
}

func TestComp_CostAllocationRepository_GetTotalByAllocated(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	catRepo := NewCostCategoryRepository(db)
	entryRepo := NewCostEntryRepository(db)
	repo := NewCostAllocationRepository(db)

	cat := &entity.CostCategory{
		ID:        uuid.New().String(),
		Code:      "CA-TOTAL-CAT-COMP",
		Name:      "Category",
		Type:      "direct",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	catRepo.Create(ctx, cat)

	entry := &entity.CostEntry{
		ID:             uuid.New().String(),
		Code:           "CA-TOTAL-ENTRY-COMP",
		Date:           time.Now(),
		CostCategoryID: cat.ID,
		Amount:         1000.0,
		Currency:       "CNY",
		ApprovalStatus: "approved",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	entryRepo.Create(ctx, entry)

	allocatedID := uuid.New().String()
	alloc := &entity.CostAllocation{
		ID:          uuid.New().String(),
		CostEntryID: entry.ID,
		AllocatedTo: "device",
		AllocatedID: allocatedID,
		Amount:      700.0,
		Percentage:  70.0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	repo.Create(ctx, alloc)

	total, err := repo.GetTotalByAllocated(ctx, "device", allocatedID, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 700.0, total)
}

func TestComp_CostReportRepository_Update(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewCostReportRepository(db)

	report := &entity.CostReport{
		ID:          uuid.New().String(),
		Code:        "CR-UPD-COMP",
		Name:        "Original",
		ReportType:  "monthly",
		PeriodStart: time.Now().AddDate(0, -1, 0),
		PeriodEnd:   time.Now(),
		TotalCost:   1000.0,
		Currency:    "CNY",
		Status:      "draft",
		GeneratedBy: "admin",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	repo.Create(ctx, report)

	report.Status = "generated"
	err := repo.Update(ctx, report)
	require.NoError(t, err)

	got, _ := repo.GetByID(ctx, report.ID)
	assert.Equal(t, "generated", got.Status)
}

func TestComp_CostReportRepository_GetByPeriod(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewCostReportRepository(db)

	periodStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	report := &entity.CostReport{
		ID:          uuid.New().String(),
		Code:        "CR-PERIOD-COMP",
		Name:        "Period Report",
		ReportType:  "monthly",
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		TotalCost:   5000.0,
		Currency:    "CNY",
		Status:      "generated",
		GeneratedBy: "admin",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	repo.Create(ctx, report)

	got, err := repo.GetByPeriod(ctx, "monthly", periodStart, periodEnd)
	require.NoError(t, err)
	assert.Equal(t, "CR-PERIOD-COMP", got.Code)
}

func TestComp_AssetRepository_GetByLocation(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewAssetRepository(db)

	location := "Building B"
	asset := &entity.Asset{
		ID:           uuid.New().String(),
		Code:         "ASSET-LOC-COMP",
		Name:         "Location Asset",
		AssetType:    "device",
		PurchaseDate: time.Now(),
		Cost:         5000,
		Location:     location,
		Status:       "active",
		UsageStatus:  "in_use",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	repo.Create(ctx, asset)

	assets, err := repo.GetByLocation(ctx, location)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(assets), 1)
}

func TestComp_AssetRepository_GetByDepartment(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewAssetRepository(db)

	deptID := uuid.New().String()
	asset := &entity.Asset{
		ID:           uuid.New().String(),
		Code:         "ASSET-DEPT-COMP",
		Name:         "Dept Asset",
		AssetType:    "equipment",
		PurchaseDate: time.Now(),
		Cost:         5000,
		DepartmentID: deptID,
		Status:       "active",
		UsageStatus:  "in_use",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	repo.Create(ctx, asset)

	assets, err := repo.GetByDepartment(ctx, deptID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(assets), 1)
}

func TestComp_AssetRepository_GetByResponsiblePerson(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewAssetRepository(db)

	person := "user-comp-test"
	asset := &entity.Asset{
		ID:                uuid.New().String(),
		Code:              "ASSET-RESP-COMP",
		Name:              "Resp Asset",
		AssetType:         "device",
		PurchaseDate:      time.Now(),
		Cost:              5000,
		ResponsiblePerson: person,
		Status:            "active",
		UsageStatus:       "in_use",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	repo.Create(ctx, asset)

	assets, err := repo.GetByResponsiblePerson(ctx, person)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(assets), 1)
}

func TestComp_AssetRepository_GetDepreciatingAssets(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewAssetRepository(db)

	asset := &entity.Asset{
		ID:                uuid.New().String(),
		Code:              "ASSET-DEP-COMP",
		Name:              "Dep Asset",
		AssetType:         "equipment",
		PurchaseDate:      time.Now(),
		Cost:              50000,
		Status:            "active",
		UsageStatus:       "in_use",
		DepreciationMethod: "straight_line",
		UsefulLife:        10,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	repo.Create(ctx, asset)

	assets, err := repo.GetDepreciatingAssets(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(assets), 1)
}

func TestComp_AssetRepository_GetAssetsNearWarrantyEnd(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewAssetRepository(db)

	warrantyEnd := time.Now().AddDate(0, 0, 15)
	asset := &entity.Asset{
		ID:              uuid.New().String(),
		Code:            "ASSET-WAR-COMP",
		Name:            "Warranty Asset",
		AssetType:       "device",
		PurchaseDate:    time.Now().AddDate(-1, 0, 0),
		WarrantyEndDate: &warrantyEnd,
		Cost:            5000,
		Status:          "active",
		UsageStatus:     "in_use",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	repo.Create(ctx, asset)

	assets, err := repo.GetAssetsNearWarrantyEnd(ctx, 30)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(assets), 1)
}

func TestComp_AssetMaintenanceRepository_Update(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	assetRepo := NewAssetRepository(db)
	repo := NewAssetMaintenanceRepository(db)

	asset := &entity.Asset{
		ID:           uuid.New().String(),
		Code:         "ASSET-MAINT-UPD-COMP",
		Name:         "Maint Asset",
		AssetType:    "device",
		PurchaseDate: time.Now(),
		Cost:         5000,
		Status:       "active",
		UsageStatus:  "in_use",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	assetRepo.Create(ctx, asset)

	record := &entity.AssetMaintenanceRecord{
		ID:              uuid.New().String(),
		AssetID:         asset.ID,
		MaintenanceType: "corrective",
		Title:           "Fix Issue",
		StartDate:       time.Now(),
		Status:          "pending",
		Cost:            300.0,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	repo.Create(ctx, record)

	record.Status = "completed"
	err := repo.Update(ctx, record)
	require.NoError(t, err)
}

func TestComp_AssetMaintenanceRepository_ListByAssetID(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	assetRepo := NewAssetRepository(db)
	repo := NewAssetMaintenanceRepository(db)

	asset := &entity.Asset{
		ID:           uuid.New().String(),
		Code:         "ASSET-MAINT-LIST-COMP",
		Name:         "Maint Asset",
		AssetType:    "device",
		PurchaseDate: time.Now(),
		Cost:         5000,
		Status:       "active",
		UsageStatus:  "in_use",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	assetRepo.Create(ctx, asset)

	for i := 0; i < 3; i++ {
		record := &entity.AssetMaintenanceRecord{
			ID:              uuid.New().String(),
			AssetID:         asset.ID,
			MaintenanceType: "preventive",
			Title:           "Check " + string(rune('A'+i)),
			StartDate:       time.Now(),
			Status:          "completed",
			Cost:            100.0,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		repo.Create(ctx, record)
	}

	records, count, err := repo.ListByAssetID(ctx, asset.ID, nil, nil, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
	assert.Equal(t, 3, len(records))
}

func TestComp_AssetMaintenanceRepository_ListByStatus(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	assetRepo := NewAssetRepository(db)
	repo := NewAssetMaintenanceRepository(db)

	asset := &entity.Asset{
		ID:           uuid.New().String(),
		Code:         "ASSET-MAINT-STATUS-COMP",
		Name:         "Maint Asset",
		AssetType:    "device",
		PurchaseDate: time.Now(),
		Cost:         5000,
		Status:       "active",
		UsageStatus:  "in_use",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	assetRepo.Create(ctx, asset)

	record := &entity.AssetMaintenanceRecord{
		ID:              uuid.New().String(),
		AssetID:         asset.ID,
		MaintenanceType: "preventive",
		Title:           "Status Check",
		StartDate:       time.Now(),
		Status:          "in_progress",
		Cost:            200.0,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	repo.Create(ctx, record)

	records, count, err := repo.ListByStatus(ctx, "in_progress", 0, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(1))
	assert.GreaterOrEqual(t, len(records), 1)
}

func TestComp_AssetMaintenanceRepository_GetMaintenanceCostByAsset(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	assetRepo := NewAssetRepository(db)
	repo := NewAssetMaintenanceRepository(db)

	asset := &entity.Asset{
		ID:           uuid.New().String(),
		Code:         "ASSET-MAINT-COST-COMP",
		Name:         "Maint Asset",
		AssetType:    "device",
		PurchaseDate: time.Now(),
		Cost:         5000,
		Status:       "active",
		UsageStatus:  "in_use",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	assetRepo.Create(ctx, asset)

	for i := 0; i < 2; i++ {
		record := &entity.AssetMaintenanceRecord{
			ID:              uuid.New().String(),
			AssetID:         asset.ID,
			MaintenanceType: "preventive",
			Title:           "Cost Check",
			StartDate:       time.Now(),
			Status:          "completed",
			Cost:            250.0,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		repo.Create(ctx, record)
	}

	cost, err := repo.GetMaintenanceCostByAsset(ctx, asset.ID, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 500.0, cost)
}

func TestComp_AssetDepreciationRepository_Update(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	assetRepo := NewAssetRepository(db)
	repo := NewAssetDepreciationRepository(db)

	asset := &entity.Asset{
		ID:           uuid.New().String(),
		Code:         "ASSET-DEPR-UPD-COMP",
		Name:         "Depr Asset",
		AssetType:    "equipment",
		PurchaseDate: time.Now(),
		Cost:         100000,
		Status:       "active",
		UsageStatus:  "in_use",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	assetRepo.Create(ctx, asset)

	record := &entity.AssetDepreciationRecord{
		ID:                      uuid.New().String(),
		AssetID:                 asset.ID,
		DepreciationDate:        time.Now(),
		Period:                  "monthly",
		DepreciationAmount:      800.0,
		AccumulatedDepreciation: 800.0,
		BookValue:               99200.0,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}
	repo.Create(ctx, record)

	record.DepreciationAmount = 900.0
	err := repo.Update(ctx, record)
	require.NoError(t, err)
}

func TestComp_AssetDepreciationRepository_ListByAssetID(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	assetRepo := NewAssetRepository(db)
	repo := NewAssetDepreciationRepository(db)

	asset := &entity.Asset{
		ID:           uuid.New().String(),
		Code:         "ASSET-DEPR-LIST-COMP",
		Name:         "Depr Asset",
		AssetType:    "equipment",
		PurchaseDate: time.Now(),
		Cost:         100000,
		Status:       "active",
		UsageStatus:  "in_use",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	assetRepo.Create(ctx, asset)

	for i := 0; i < 3; i++ {
		record := &entity.AssetDepreciationRecord{
			ID:                      uuid.New().String(),
			AssetID:                 asset.ID,
			DepreciationDate:        time.Now().AddDate(0, -i, 0),
			Period:                  "monthly",
			DepreciationAmount:      800.0,
			AccumulatedDepreciation: float64(i+1) * 800.0,
			BookValue:               100000 - float64(i+1)*800.0,
			CreatedAt:               time.Now(),
			UpdatedAt:               time.Now(),
		}
		repo.Create(ctx, record)
	}

	records, count, err := repo.ListByAssetID(ctx, asset.ID, nil, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
	assert.Equal(t, 3, len(records))
}

func TestComp_AssetDepreciationRepository_GetLatestByAssetID(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	assetRepo := NewAssetRepository(db)
	repo := NewAssetDepreciationRepository(db)

	asset := &entity.Asset{
		ID:           uuid.New().String(),
		Code:         "ASSET-DEPR-LATEST-COMP",
		Name:         "Depr Asset",
		AssetType:    "equipment",
		PurchaseDate: time.Now(),
		Cost:         100000,
		Status:       "active",
		UsageStatus:  "in_use",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	assetRepo.Create(ctx, asset)

	record := &entity.AssetDepreciationRecord{
		ID:                      uuid.New().String(),
		AssetID:                 asset.ID,
		DepreciationDate:        time.Now(),
		Period:                  "monthly",
		DepreciationAmount:      800.0,
		AccumulatedDepreciation: 800.0,
		BookValue:               99200.0,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}
	repo.Create(ctx, record)

	got, err := repo.GetLatestByAssetID(ctx, asset.ID)
	require.NoError(t, err)
	assert.Equal(t, 800.0, got.DepreciationAmount)
}

func TestComp_AssetDepreciationRepository_GetDepreciationSummaryByPeriod(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	assetRepo := NewAssetRepository(db)
	repo := NewAssetDepreciationRepository(db)

	asset := &entity.Asset{
		ID:           uuid.New().String(),
		Code:         "ASSET-DEPR-SUM-COMP",
		Name:         "Depr Asset",
		AssetType:    "equipment",
		PurchaseDate: time.Now(),
		Cost:         100000,
		Status:       "active",
		UsageStatus:  "in_use",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	assetRepo.Create(ctx, asset)

	for i := 0; i < 2; i++ {
		record := &entity.AssetDepreciationRecord{
			ID:                      uuid.New().String(),
			AssetID:                 asset.ID,
			DepreciationDate:        time.Now(),
			Period:                  "monthly",
			DepreciationAmount:      800.0,
			AccumulatedDepreciation: float64(i+1) * 800.0,
			BookValue:               100000 - float64(i+1)*800.0,
			CreatedAt:               time.Now(),
			UpdatedAt:               time.Now(),
		}
		repo.Create(ctx, record)
	}

	total, err := repo.GetDepreciationSummaryByPeriod(ctx, "monthly", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 1600.0, total)
}

func TestComp_AssetDocumentRepository_Update(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	assetRepo := NewAssetRepository(db)
	repo := NewAssetDocumentRepository(db)

	asset := &entity.Asset{
		ID:           uuid.New().String(),
		Code:         "ASSET-DOC-UPD-COMP",
		Name:         "Doc Asset",
		AssetType:    "device",
		PurchaseDate: time.Now(),
		Cost:         5000,
		Status:       "active",
		UsageStatus:  "in_use",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	assetRepo.Create(ctx, asset)

	doc := &entity.AssetDocument{
		ID:         uuid.New().String(),
		AssetID:    asset.ID,
		Title:      "Original Doc",
		Type:       "datasheet",
		UploadDate: time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	repo.Create(ctx, doc)

	doc.Title = "Updated Doc"
	err := repo.Update(ctx, doc)
	require.NoError(t, err)
}

func TestComp_AssetDocumentRepository_ListByAssetID(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	assetRepo := NewAssetRepository(db)
	repo := NewAssetDocumentRepository(db)

	asset := &entity.Asset{
		ID:           uuid.New().String(),
		Code:         "ASSET-DOC-LIST-COMP",
		Name:         "Doc Asset",
		AssetType:    "device",
		PurchaseDate: time.Now(),
		Cost:         5000,
		Status:       "active",
		UsageStatus:  "in_use",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	assetRepo.Create(ctx, asset)

	for i := 0; i < 3; i++ {
		doc := &entity.AssetDocument{
			ID:         uuid.New().String(),
			AssetID:    asset.ID,
			Title:      "Doc " + string(rune('A'+i)),
			Type:       "manual",
			UploadDate: time.Now(),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		repo.Create(ctx, doc)
	}

	docs, count, err := repo.ListByAssetID(ctx, asset.ID, nil, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
	assert.Equal(t, 3, len(docs))
}

func TestComp_AssetDocumentRepository_GetByType(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	assetRepo := NewAssetRepository(db)
	repo := NewAssetDocumentRepository(db)

	asset := &entity.Asset{
		ID:           uuid.New().String(),
		Code:         "ASSET-DOC-TYPE-COMP",
		Name:         "Doc Asset",
		AssetType:    "device",
		PurchaseDate: time.Now(),
		Cost:         5000,
		Status:       "active",
		UsageStatus:  "in_use",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	assetRepo.Create(ctx, asset)

	doc := &entity.AssetDocument{
		ID:         uuid.New().String(),
		AssetID:    asset.ID,
		Title:      "Invoice Doc",
		Type:       "invoice",
		UploadDate: time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	repo.Create(ctx, doc)

	docs, count, err := repo.GetByType(ctx, "invoice", 0, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(1))
	assert.GreaterOrEqual(t, len(docs), 1)
}

func createEdgeNodeDirectly(t *testing.T, db *Database, node *entity.EdgeNode) {
	t.Helper()
	err := db.WithContext(context.Background()).Omit("model_versions").Create(node).Error
	if err != nil {
		t.Fatalf("failed to create edge node: %v", err)
	}
}

func TestComp_EdgeNodeRepository_CRUD(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewEdgeNodeRepository(db)

	node := &entity.EdgeNode{
		ID:        uuid.New().String(),
		Name:      "Edge Node Comp",
		StationID: uuid.New().String(),
		IPAddress: "192.168.1.100",
		Status:    entity.EdgeNodeOnline,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	createEdgeNodeDirectly(t, db, node)

	got, err := repo.GetByID(ctx, node.ID)
	require.NoError(t, err)
	assert.Equal(t, "Edge Node Comp", got.Name)

	node.Name = "Updated Node"
	err = db.WithContext(ctx).Omit("model_versions").Save(node).Error
	require.NoError(t, err)

	got, _ = repo.GetByID(ctx, node.ID)
	assert.Equal(t, "Updated Node", got.Name)

	err = repo.Delete(ctx, node.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, node.ID)
	assert.Error(t, err)
}

func TestComp_EdgeNodeRepository_List(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewEdgeNodeRepository(db)

	stationID := uuid.New().String()
	for i := 0; i < 3; i++ {
		node := &entity.EdgeNode{
			ID:        uuid.New().String(),
			Name:      "Node " + string(rune('A'+i)),
			StationID: stationID,
			IPAddress: "192.168.1." + string(rune('1'+i)),
			Status:    entity.EdgeNodeOnline,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		createEdgeNodeDirectly(t, db, node)
	}

	nodes, err := repo.List(ctx, &stationID, nil)
	require.NoError(t, err)
	assert.Equal(t, 3, len(nodes))
}

func TestComp_EdgeNodeRepository_UpdateStatus(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewEdgeNodeRepository(db)

	node := &entity.EdgeNode{
		ID:        uuid.New().String(),
		Name:      "Status Node",
		StationID: uuid.New().String(),
		IPAddress: "192.168.1.103",
		Status:    entity.EdgeNodeOffline,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	createEdgeNodeDirectly(t, db, node)

	err := repo.UpdateStatus(ctx, node.ID, entity.EdgeNodeOnline)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, node.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.EdgeNodeOnline, got.Status)
}

func TestComp_EdgeNodeRepository_UpdateHeartbeat(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewEdgeNodeRepository(db)

	node := &entity.EdgeNode{
		ID:        uuid.New().String(),
		Name:      "Heartbeat Node",
		StationID: uuid.New().String(),
		IPAddress: "192.168.1.104",
		Status:    entity.EdgeNodeOffline,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	createEdgeNodeDirectly(t, db, node)

	err := repo.UpdateHeartbeat(ctx, node.ID, 45.5, 60.2)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, node.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.EdgeNodeOnline, got.Status)
}

func TestComp_EnergyEfficiencyRepository_CreateAndGetRecord(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewEnergyEfficiencyRepository(db)

	record := entity.NewEnergyEfficiencyRecord(
		time.Now(),
		entity.EnergyEfficiencyTypeDevice,
		"target-comp-1",
		"Device 1",
		1000.0,
		900.0,
		"daily",
	)
	record.ID = uuid.New().String()

	err := repo.CreateRecord(ctx, record)
	require.NoError(t, err)

	got, err := repo.GetRecordByID(ctx, record.ID)
	require.NoError(t, err)
	assert.Equal(t, "target-comp-1", got.TargetID)
}

func TestComp_EnergyEfficiencyRepository_BatchCreateRecords(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewEnergyEfficiencyRepository(db)

	var records []*entity.EnergyEfficiencyRecord
	for i := 0; i < 3; i++ {
		r := entity.NewEnergyEfficiencyRecord(
			time.Now().Add(time.Duration(i)*time.Hour),
			entity.EnergyEfficiencyTypeStation,
			"target-batch-comp",
			"Station Batch",
			1000.0,
			850.0,
			"daily",
		)
		r.ID = uuid.New().String()
		records = append(records, r)
	}

	err := repo.BatchCreateRecords(ctx, records)
	require.NoError(t, err)
}

func TestComp_EnergyEfficiencyRepository_ListRecords(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewEnergyEfficiencyRepository(db)

	eeType := entity.EnergyEfficiencyTypeDevice
	targetID := "target-list-comp"
	for i := 0; i < 3; i++ {
		r := entity.NewEnergyEfficiencyRecord(
			time.Now().Add(time.Duration(i)*time.Hour),
			eeType,
			targetID,
			"Device List",
			1000.0,
			900.0,
			"daily",
		)
		r.ID = uuid.New().String()
		repo.CreateRecord(ctx, r)
	}

	query := &repository.EnergyEfficiencyQuery{
		Page:     1,
		PageSize: 10,
		Type:     &eeType,
		TargetID: &targetID,
	}

	records, total, err := repo.ListRecords(ctx, query)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Equal(t, 3, len(records))
}

func TestComp_EnergyEfficiencyRepository_GetRecordsByTimeRange(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewEnergyEfficiencyRepository(db)

	now := time.Now()
	r := entity.NewEnergyEfficiencyRecord(
		now,
		entity.EnergyEfficiencyTypeDevice,
		"target-range-comp",
		"Device Range",
		1000.0,
		900.0,
		"daily",
	)
	r.ID = uuid.New().String()
	repo.CreateRecord(ctx, r)

	records, err := repo.GetRecordsByTimeRange(ctx, "target-range-comp", entity.EnergyEfficiencyTypeDevice, now.Add(-1*time.Hour), now.Add(1*time.Hour))
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(records), 1)
}

func TestComp_EnergyEfficiencyRepository_Analysis(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewEnergyEfficiencyRepository(db)

	analysis := &entity.EnergyEfficiencyAnalysis{
		ID:               uuid.New().String(),
		AnalysisTime:     time.Now(),
		Type:             entity.EnergyEfficiencyTypeDevice,
		TargetID:         "target-analysis-comp",
		TargetName:       "Device Analysis",
		TimeRangeStart:   time.Now().Add(-24 * time.Hour),
		TimeRangeEnd:     time.Now(),
		AvgEfficiency:    0.90,
		MaxEfficiency:    0.95,
		MinEfficiency:    0.85,
		StdDevEfficiency: 0.03,
		Trend:            "stable",
		CreatedAt:        time.Now(),
	}

	err := repo.CreateAnalysis(ctx, analysis)
	require.NoError(t, err)

	got, err := repo.GetAnalysisByID(ctx, analysis.ID)
	require.NoError(t, err)
	assert.Equal(t, "target-analysis-comp", got.TargetID)

	eeType := entity.EnergyEfficiencyTypeDevice
	query := &repository.EnergyEfficiencyAnalysisQuery{
		Page:     1,
		PageSize: 10,
		Type:     &eeType,
		TargetID: &analysis.TargetID,
	}

	_, total, err := repo.ListAnalyses(ctx, query)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1))

	gotLatest, err := repo.GetLatestAnalysis(ctx, "target-analysis-comp", entity.EnergyEfficiencyTypeDevice)
	require.NoError(t, err)
	assert.Equal(t, 0.90, gotLatest.AvgEfficiency)
}

func TestComp_EnergyEfficiencyRepository_GetStatistics(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewEnergyEfficiencyRepository(db)

	targetID := "target-stats-comp"
	for i := 0; i < 3; i++ {
		r := entity.NewEnergyEfficiencyRecord(
			time.Now().Add(time.Duration(i)*time.Hour),
			entity.EnergyEfficiencyTypeDevice,
			targetID,
			"Device Stats",
			1000.0,
			900.0+float64(i)*10,
			"daily",
		)
		r.ID = uuid.New().String()
		repo.CreateRecord(ctx, r)
	}

	stats, err := repo.GetStatistics(ctx, targetID, entity.EnergyEfficiencyTypeDevice, "daily", time.Now().Add(-24*time.Hour), time.Now().Add(24*time.Hour))
	require.NoError(t, err)
	assert.Equal(t, int64(3), stats.TotalRecords)
	assert.Greater(t, stats.AvgEfficiency, 0.0)
}

func TestComp_EnergyEfficiencyRepository_GetBenchmark(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewEnergyEfficiencyRepository(db)

	r := entity.NewEnergyEfficiencyRecord(
		time.Now(),
		entity.EnergyEfficiencyTypeDevice,
		"target-other-comp",
		"Other Device",
		1000.0,
		880.0,
		"daily",
	)
	r.ID = uuid.New().String()
	repo.CreateRecord(ctx, r)

	benchmark, err := repo.GetBenchmark(ctx, entity.EnergyEfficiencyTypeDevice, "target-bench-comp")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, benchmark, 0.0)
}

func TestComp_CarbonEmissionRepository_FactorCRUD(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewCarbonEmissionRepository(db)

	factor := entity.NewCarbonEmissionFactor(
		"Electricity Factor Comp",
		"CEF-COMP-001",
		entity.CarbonEmissionScope2,
		"National Grid",
		0.5810,
		"tCO2e/MWh",
		"2024",
		time.Now(),
	)
	factor.ID = uuid.New().String()

	err := repo.CreateFactor(ctx, factor)
	require.NoError(t, err)

	got, err := repo.GetFactorByID(ctx, factor.ID)
	require.NoError(t, err)
	assert.Equal(t, "CEF-COMP-001", got.Code)

	byCode, err := repo.GetFactorByCode(ctx, "CEF-COMP-001")
	require.NoError(t, err)
	assert.Equal(t, factor.ID, byCode.ID)

	factor.Name = "Updated Factor"
	err = repo.UpdateFactor(ctx, factor)
	require.NoError(t, err)

	scope := entity.CarbonEmissionScope2
	query := &repository.CarbonEmissionFactorQuery{
		Page:     1,
		PageSize: 10,
		Scope:    &scope,
	}

	_, total, err := repo.ListFactors(ctx, query)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1))

	activeFactors, err := repo.GetActiveFactors(ctx, &scope)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(activeFactors), 1)
}

func TestComp_CarbonEmissionRepository_RecordCRUD(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewCarbonEmissionRepository(db)

	record := entity.NewCarbonEmissionRecord(
		time.Now(),
		entity.CarbonEmissionScope2,
		"target-comp-1",
		"Station 1",
		"factor-1",
		"CEF-REC-COMP",
		0.5810,
		1000.0,
		"MWh",
		"monthly",
	)
	record.ID = uuid.New().String()

	err := repo.CreateRecord(ctx, record)
	require.NoError(t, err)

	got, err := repo.GetRecordByID(ctx, record.ID)
	require.NoError(t, err)
	assert.Equal(t, "target-comp-1", got.TargetID)

	record.Status = entity.CarbonEmissionStatusApproved
	err = repo.UpdateRecord(ctx, record)
	require.NoError(t, err)

	var batchRecords []*entity.CarbonEmissionRecord
	for i := 0; i < 3; i++ {
		r := entity.NewCarbonEmissionRecord(
			time.Now().Add(time.Duration(i)*time.Hour),
			entity.CarbonEmissionScope1,
			"target-batch-comp",
			"Station Batch",
			"factor-1",
			"CEF-BATCH-COMP",
			2.16,
			500.0,
			"MWh",
			"monthly",
		)
		r.ID = uuid.New().String()
		batchRecords = append(batchRecords, r)
	}
	err = repo.BatchCreateRecords(ctx, batchRecords)
	require.NoError(t, err)
}

func TestComp_CarbonEmissionRepository_ListRecords(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewCarbonEmissionRepository(db)

	scope := entity.CarbonEmissionScope2
	targetID := "target-list-comp-rec"
	for i := 0; i < 3; i++ {
		r := entity.NewCarbonEmissionRecord(
			time.Now().Add(time.Duration(i)*time.Hour),
			scope,
			targetID,
			"Station List",
			"factor-1",
			"CEF-LIST-COMP-REC",
			0.5810,
			1000.0,
			"MWh",
			"monthly",
		)
		r.ID = uuid.New().String()
		repo.CreateRecord(ctx, r)
	}

	query := &repository.CarbonEmissionRecordQuery{
		Page:     1,
		PageSize: 10,
		Scope:    &scope,
		TargetID: &targetID,
	}

	records, total, err := repo.ListRecords(ctx, query)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Equal(t, 3, len(records))
}

func TestComp_CarbonEmissionRepository_GetRecordsByTimeRange(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewCarbonEmissionRepository(db)

	now := time.Now()
	r := entity.NewCarbonEmissionRecord(
		now,
		entity.CarbonEmissionScope1,
		"target-range-comp",
		"Station Range",
		"factor-1",
		"CEF-RANGE-COMP",
		2.16,
		500.0,
		"MWh",
		"monthly",
	)
	r.ID = uuid.New().String()
	repo.CreateRecord(ctx, r)

	records, err := repo.GetRecordsByTimeRange(ctx, "target-range-comp", nil, now.Add(-1*time.Hour), now.Add(1*time.Hour))
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(records), 1)
}

func TestComp_CarbonEmissionRepository_GetTotalEmissionByScope(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewCarbonEmissionRepository(db)

	r := entity.NewCarbonEmissionRecord(
		time.Now(),
		entity.CarbonEmissionScope2,
		"target-total-comp",
		"Station Total",
		"factor-1",
		"CEF-TOTAL-COMP",
		0.5810,
		1000.0,
		"MWh",
		"monthly",
	)
	r.ID = uuid.New().String()
	repo.CreateRecord(ctx, r)

	total, err := repo.GetTotalEmissionByScope(ctx, "target-total-comp", entity.CarbonEmissionScope2, "monthly", time.Now().Add(-24*time.Hour), time.Now().Add(24*time.Hour))
	require.NoError(t, err)
	assert.Greater(t, total, 0.0)
}

func TestComp_CarbonEmissionRepository_SummaryCRUD(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewCarbonEmissionRepository(db)

	summary := &entity.CarbonEmissionSummary{
		ID:             uuid.New().String(),
		SummaryTime:    time.Now(),
		TargetID:       "target-summary-comp",
		TargetName:     "Station Summary",
		Period:         "monthly",
		PeriodStart:    time.Now().AddDate(0, -1, 0),
		PeriodEnd:      time.Now(),
		Scope1Emission: 100.0,
		Scope2Emission: 500.0,
		Scope3Emission: 50.0,
		TotalEmission:  650.0,
		Status:         entity.CarbonEmissionStatusDraft,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := repo.CreateSummary(ctx, summary)
	require.NoError(t, err)

	got, err := repo.GetSummaryByID(ctx, summary.ID)
	require.NoError(t, err)
	assert.Equal(t, "target-summary-comp", got.TargetID)

	summary.TotalEmission = 700.0
	err = repo.UpdateSummary(ctx, summary)
	require.NoError(t, err)

	got, _ = repo.GetSummaryByID(ctx, summary.ID)
	assert.Equal(t, 700.0, got.TotalEmission)

	gotLatest, err := repo.GetLatestSummary(ctx, "target-summary-comp", "monthly")
	require.NoError(t, err)
	assert.Equal(t, 700.0, gotLatest.TotalEmission)
}

func TestComp_CarbonEmissionRepository_ListSummaries(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewCarbonEmissionRepository(db)

	targetID := "target-list-comp-sum"
	for i := 0; i < 3; i++ {
		summary := &entity.CarbonEmissionSummary{
			ID:             uuid.New().String(),
			SummaryTime:    time.Now().Add(time.Duration(i) * time.Hour),
			TargetID:       targetID,
			TargetName:     "Station List",
			Period:         "monthly",
			PeriodStart:    time.Now().AddDate(0, -1, 0),
			PeriodEnd:      time.Now(),
			TotalEmission:  500.0,
			Status:         entity.CarbonEmissionStatusDraft,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		repo.CreateSummary(ctx, summary)
	}

	query := &repository.CarbonEmissionSummaryQuery{
		Page:     1,
		PageSize: 10,
		TargetID: &targetID,
	}

	summaries, total, err := repo.ListSummaries(ctx, query)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Equal(t, 3, len(summaries))
}

func TestComp_CarbonEmissionRepository_TargetCRUD(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewCarbonEmissionRepository(db)

	target := &entity.CarbonReductionTarget{
		ID:              uuid.New().String(),
		Name:            "2025 Reduction Target Comp",
		TargetID:        "target-reduction-comp",
		TargetName:      "Station Reduction",
		BaseYear:        2023,
		BaseEmission:    1000.0,
		TargetYear:      2025,
		TargetReduction: 0.30,
		TargetEmission:  700.0,
		CurrentProgress: 0.10,
		CurrentEmission: 900.0,
		StartDate:       time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:         time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
		Status:          "active",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err := repo.CreateTarget(ctx, target)
	require.NoError(t, err)

	got, err := repo.GetTargetByID(ctx, target.ID)
	require.NoError(t, err)
	assert.Equal(t, "target-reduction-comp", got.TargetID)

	target.CurrentProgress = 0.15
	err = repo.UpdateTarget(ctx, target)
	require.NoError(t, err)

	got, _ = repo.GetTargetByID(ctx, target.ID)
	assert.Equal(t, 0.15, got.CurrentProgress)

	targets, err := repo.GetActiveTargets(ctx, "target-reduction-comp")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(targets), 1)
}

func TestComp_CarbonEmissionRepository_ListTargets(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewCarbonEmissionRepository(db)

	targetID := "target-list-comp-red"
	for i := 0; i < 3; i++ {
		target := &entity.CarbonReductionTarget{
			ID:              uuid.New().String(),
			Name:            "Target Comp " + string(rune('A'+i)),
			TargetID:        targetID,
			TargetName:      "Station",
			BaseYear:        2023,
			BaseEmission:    1000.0,
			TargetYear:      2025 + i,
			TargetReduction: 0.30,
			TargetEmission:  700.0,
			StartDate:       time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:         time.Date(2025+i, 12, 31, 0, 0, 0, 0, time.UTC),
			Status:          "active",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		repo.CreateTarget(ctx, target)
	}

	query := &repository.CarbonReductionTargetQuery{
		Page:     1,
		PageSize: 10,
		TargetID: &targetID,
	}

	targets, total, err := repo.ListTargets(ctx, query)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Equal(t, 3, len(targets))
}

func TestComp_ReportRepository_All(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewReportRepository(db)

	now := time.Now()
	start := now.Add(-24 * time.Hour)

	powerStats, err := repo.GetStationPowerStats(ctx, "station-1", start, now)
	require.NoError(t, err)
	assert.Equal(t, "station-1", powerStats.StationID)

	alarmStats, err := repo.GetStationAlarmStats(ctx, "station-1", start, now)
	require.NoError(t, err)
	assert.Equal(t, "station-1", alarmStats.StationID)

	onlineStats, err := repo.GetStationOnlineStats(ctx, "station-1", start, now)
	require.NoError(t, err)
	assert.Equal(t, "station-1", onlineStats.StationID)

	allPowerStats, err := repo.GetAllStationPowerStats(ctx, start, now)
	require.NoError(t, err)
	assert.NotNil(t, allPowerStats)

	allAlarmStats, err := repo.GetAllStationAlarmStats(ctx, start, now)
	require.NoError(t, err)
	assert.NotNil(t, allAlarmStats)

	allOnlineStats, err := repo.GetAllStationOnlineStats(ctx, start, now)
	require.NoError(t, err)
	assert.NotNil(t, allOnlineStats)
}

func TestComp_RegionRepository_CRUD(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewRegionRepository(db)

	region := &entity.Region{
		ID:        uuid.New().String(),
		Code:      "REG-COMP-001",
		Name:      "Test Region",
		SortOrder: 1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, region)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, region.ID)
	require.NoError(t, err)
	assert.Equal(t, "REG-COMP-001", got.Code)

	byCode, err := repo.GetByCode(ctx, "REG-COMP-001")
	require.NoError(t, err)
	assert.Equal(t, region.ID, byCode.ID)

	region.Name = "Updated Region"
	err = repo.Update(ctx, region)
	require.NoError(t, err)

	got, _ = repo.GetByID(ctx, region.ID)
	assert.Equal(t, "Updated Region", got.Name)

	regions, err := repo.List(ctx, nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(regions), 1)

	tree, err := repo.GetTree(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(tree), 1)

	err = repo.Delete(ctx, region.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, region.ID)
	assert.Error(t, err)
}

func TestComp_RegionRepository_ListWithParent(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewRegionRepository(db)

	parent := &entity.Region{
		ID:        uuid.New().String(),
		Code:      "REG-PARENT-COMP",
		Name:      "Parent Region",
		SortOrder: 1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.Create(ctx, parent)

	child := &entity.Region{
		ID:        uuid.New().String(),
		ParentID:  &parent.ID,
		Code:      "REG-CHILD-COMP",
		Name:      "Child Region",
		SortOrder: 2,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.Create(ctx, child)

	regions, err := repo.List(ctx, &parent.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(regions), 1)
}

func TestComp_StationRepository_CRUD(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewStationRepository(db)

	station := &entity.Station{
		ID:           uuid.New().String(),
		Code:         "STA-COMP-001",
		Name:         "Test Station",
		Type:         entity.StationTypePV,
		SubRegionID:  uuid.New().String(),
		Status:       entity.StationStatusActive,
		Capacity: 100.0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(ctx, station)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, station.ID)
	require.NoError(t, err)
	assert.Equal(t, "STA-COMP-001", got.Code)

	byCode, err := repo.GetByCode(ctx, "STA-COMP-001")
	require.NoError(t, err)
	assert.Equal(t, station.ID, byCode.ID)

	station.Name = "Updated Station"
	err = repo.Update(ctx, station)
	require.NoError(t, err)

	got, _ = repo.GetByID(ctx, station.ID)
	assert.Equal(t, "Updated Station", got.Name)

	stations, err := repo.List(ctx, nil, nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(stations), 1)

	withDevices, err := repo.GetWithDevices(ctx, station.ID)
	require.NoError(t, err)
	assert.Equal(t, station.ID, withDevices.ID)

	err = repo.Delete(ctx, station.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, station.ID)
	assert.Error(t, err)
}

func TestComp_StationRepository_ListWithFilters(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewStationRepository(db)

	subRegionID := uuid.New().String()
	stationType := entity.StationTypePV
	station := &entity.Station{
		ID:           uuid.New().String(),
		Code:         "STA-FILTER-COMP",
		Name:         "Filter Station",
		Type:         stationType,
		SubRegionID:  subRegionID,
		Status:       entity.StationStatusActive,
		Capacity: 100.0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	repo.Create(ctx, station)

	stations, err := repo.List(ctx, &subRegionID, &stationType)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(stations), 1)
}

func TestComp_PointRepository_CRUD(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	stationRepo := NewStationRepository(db)
	deviceRepo := NewDeviceRepository(db)
	repo := NewPointRepository(db)

	station := &entity.Station{
		ID:           uuid.New().String(),
		Code:         "STA-PT-COMP",
		Name:         "Point Station",
		Type:         entity.StationTypePV,
		SubRegionID:  uuid.New().String(),
		Status:       entity.StationStatusActive,
		Capacity: 100.0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	stationRepo.Create(ctx, station)

	device := &entity.Device{
		ID:        uuid.New().String(),
		StationID: station.ID,
		Code:      "DEV-PT-COMP",
		Name:      "Point Device",
		Type:      entity.DeviceTypeInverter,
		Status:    entity.DeviceStatusOnline,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	deviceRepo.Create(ctx, device)

	point := &entity.Point{
		ID:        uuid.New().String(),
		DeviceID:  device.ID,
		Code:      "PT-COMP-001",
		Name:      "Test Point",
		Type:      entity.PointTypeYaoCe,
		Unit:      "kW",
		Protocol:  "modbus",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, point)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, point.ID)
	require.NoError(t, err)
	assert.Equal(t, "PT-COMP-001", got.Code)

	byCode, err := repo.GetByCode(ctx, "PT-COMP-001")
	require.NoError(t, err)
	assert.Equal(t, point.ID, byCode.ID)

	point.Name = "Updated Point"
	err = repo.Update(ctx, point)
	require.NoError(t, err)

	got, _ = repo.GetByID(ctx, point.ID)
	assert.Equal(t, "Updated Point", got.Name)

	points, err := repo.List(ctx, &device.ID, nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(points), 1)

	ptType := entity.PointTypeYaoCe
	points, err = repo.List(ctx, &device.ID, &ptType)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(points), 1)

	stationPoints, err := repo.GetByStationID(ctx, station.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(stationPoints), 1)

	protocolPoints, err := repo.GetByProtocol(ctx, "modbus")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(protocolPoints), 1)

	var batchPoints []*entity.Point
	for i := 0; i < 3; i++ {
		batchPoints = append(batchPoints, &entity.Point{
			ID:        uuid.New().String(),
			DeviceID:  device.ID,
			Code:      "PT-BATCH-COMP-" + string(rune('A'+i)),
			Name:      "Batch Point " + string(rune('A'+i)),
			Type:      entity.PointTypeYaoXin,
			Protocol:  "opcua",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}
	err = repo.BatchCreate(ctx, batchPoints)
	require.NoError(t, err)

	err = repo.Delete(ctx, point.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, point.ID)
	assert.Error(t, err)
}

func TestComp_AlarmRepository_CRUD(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewAlarmRepository(db)

	now := time.Now()
	alarm := &entity.Alarm{
		ID:          uuid.New().String(),
		StationID:   uuid.New().String(),
		DeviceID:    uuid.New().String(),
		PointID:     uuid.New().String(),
		Type:        entity.AlarmTypeLimit,
		Level:       entity.AlarmLevelCritical,
		Status:      entity.AlarmStatusActive,
		Title:       "Test Alarm",
		Message:     "Test alarm",
		TriggeredAt: now,
		CreatedAt:   now,
	}

	err := repo.Create(ctx, alarm)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, alarm.ID)
	require.NoError(t, err)
	assert.Equal(t, "Test alarm", got.Message)

	alarm.Message = "Updated alarm"
	err = repo.Update(ctx, alarm)
	require.NoError(t, err)

	got, _ = repo.GetByID(ctx, alarm.ID)
	assert.Equal(t, "Updated alarm", got.Message)

	activeAlarms, err := repo.GetActiveAlarms(ctx, nil, nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(activeAlarms), 1)

	stationID := alarm.StationID
	activeAlarms, err = repo.GetActiveAlarms(ctx, &stationID, nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(activeAlarms), 1)

	level := entity.AlarmLevelCritical
	activeAlarms, err = repo.GetActiveAlarms(ctx, &stationID, &level)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(activeAlarms), 1)

	startTime := now.Add(-1 * time.Hour).Unix()
	endTime := now.Add(1 * time.Hour).Unix()
	historyAlarms, err := repo.GetHistoryAlarms(ctx, nil, startTime, endTime)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(historyAlarms), 1)

	historyAlarms, err = repo.GetHistoryAlarms(ctx, &stationID, startTime, endTime)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(historyAlarms), 1)

	err = repo.Acknowledge(ctx, alarm.ID, "admin")
	require.NoError(t, err)

	err = repo.Clear(ctx, alarm.ID)
	require.NoError(t, err)

	counts, err := repo.CountByLevel(ctx, nil)
	require.NoError(t, err)
	assert.NotNil(t, counts)

	counts, err = repo.CountByLevel(ctx, &stationID)
	require.NoError(t, err)
	assert.NotNil(t, counts)
}

func TestComp_SubRegionRepository_CRUD(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	regionRepo := NewRegionRepository(db)
	repo := NewSubRegionRepository(db)

	region := &entity.Region{
		ID:        uuid.New().String(),
		Code:      "REG-SUB-COMP",
		Name:      "Sub Region Parent",
		SortOrder: 1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	regionRepo.Create(ctx, region)

	subRegion := &entity.SubRegion{
		ID:          uuid.New().String(),
		RegionID:    region.ID,
		Code:        "SUB-COMP-001",
		Name:        "Test SubRegion",
		Description: "test sub region",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := repo.Create(ctx, subRegion)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, subRegion.ID)
	require.NoError(t, err)
	assert.Equal(t, "SUB-COMP-001", got.Code)

	subRegion.Name = "Updated SubRegion"
	err = repo.Update(ctx, subRegion)
	require.NoError(t, err)

	got, _ = repo.GetByID(ctx, subRegion.ID)
	assert.Equal(t, "Updated SubRegion", got.Name)

	subRegions, err := repo.GetByRegionID(ctx, region.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(subRegions), 1)

	err = repo.Delete(ctx, subRegion.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, subRegion.ID)
	assert.Error(t, err)
}

func TestComp_MigrationManager(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	mgr := NewMigrationManager(db)

	migrationsFS := fstest.MapFS{
		"migrations/000_init.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE IF NOT EXISTS init_table (id TEXT PRIMARY KEY)"),
		},
	}

	err := mgr.RunMigrations(ctx, migrationsFS)
	require.NoError(t, err)

	status, err := mgr.GetMigrationStatus(ctx)
	require.NoError(t, err)
	assert.NotNil(t, status)

	isApplied, err := mgr.IsMigrationApplied(ctx, "000_init")
	require.NoError(t, err)
	assert.True(t, isApplied)

	isApplied, err = mgr.IsMigrationApplied(ctx, "nonexistent")
	require.NoError(t, err)
	assert.False(t, isApplied)

	statusSummary, err := mgr.GetMigrationStatusSummary(ctx, migrationsFS)
	require.NoError(t, err)
	assert.NotNil(t, statusSummary)

	pending, err := mgr.GetPendingMigrations(ctx, migrationsFS)
	require.NoError(t, err)
	assert.Equal(t, 0, len(pending))

	err = mgr.RollbackMigration(ctx, "nonexistent")
	require.NoError(t, err)
}

func TestComp_MigrationManager_WithFS(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	mgr := NewMigrationManager(db)

	migrationsFS := fstest.MapFS{
		"migrations/001_create_test.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE IF NOT EXISTS test_table (id TEXT PRIMARY KEY)"),
		},
		"migrations/002_add_column.sql": &fstest.MapFile{
			Data: []byte("ALTER TABLE test_table ADD COLUMN name TEXT"),
		},
	}

	err := mgr.RunMigrations(ctx, migrationsFS)
	require.NoError(t, err)

	err = mgr.RunMigrations(ctx, migrationsFS)
	require.NoError(t, err)

	status, err := mgr.GetMigrationStatusSummary(ctx, migrationsFS)
	require.NoError(t, err)
	assert.Equal(t, 2, status.Total)
	assert.Equal(t, 2, status.Applied)
	assert.Equal(t, 0, status.Pending)
	assert.NotNil(t, status.LastApplied)

	isApplied, err := mgr.IsMigrationApplied(ctx, "001_create_test")
	require.NoError(t, err)
	assert.True(t, isApplied)

	pending, err := mgr.GetPendingMigrations(ctx, migrationsFS)
	require.NoError(t, err)
	assert.Equal(t, 0, len(pending))

	err = mgr.ApplySpecificMigration(ctx, migrationsFS, "001_create_test")
	assert.Error(t, err)

	err = mgr.RollbackMigrationWithScript(ctx, "002_add_column", "ALTER TABLE test_table DROP COLUMN name")
	require.NoError(t, err)
}

func TestComp_MigrationManager_RunWithLimit(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	mgr := NewMigrationManager(db)

	migrationsFS := fstest.MapFS{
		"migrations/003_test_a.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE IF NOT EXISTS test_a (id TEXT PRIMARY KEY)"),
		},
		"migrations/004_test_b.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE IF NOT EXISTS test_b (id TEXT PRIMARY KEY)"),
		},
	}

	err := mgr.RunMigrationsWithLimit(ctx, migrationsFS, 1)
	require.NoError(t, err)

	isApplied, _ := mgr.IsMigrationApplied(ctx, "003_test_a")
	assert.True(t, isApplied)

	isApplied, _ = mgr.IsMigrationApplied(ctx, "004_test_b")
	assert.False(t, isApplied)

	err = mgr.RunMigrationsWithLimit(ctx, migrationsFS, 1)
	require.NoError(t, err)

	isApplied, _ = mgr.IsMigrationApplied(ctx, "004_test_b")
	assert.True(t, isApplied)
}

func TestComp_MigrationManager_ApplySpecific(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	mgr := NewMigrationManager(db)

	migrationsFS := fstest.MapFS{
		"migrations/005_test_c.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE IF NOT EXISTS test_c (id TEXT PRIMARY KEY)"),
		},
	}

	err := mgr.ApplySpecificMigration(ctx, migrationsFS, "005_test_c")
	require.NoError(t, err)

	err = mgr.ApplySpecificMigration(ctx, migrationsFS, "005_test_c")
	assert.Error(t, err)
}

func TestComp_MigrationManager_ValidationError(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	mgr := NewMigrationManager(db)

	migrationsFS := fstest.MapFS{
		"migrations/006_empty.sql": &fstest.MapFile{
			Data: []byte("  "),
		},
	}

	err := mgr.RunMigrations(ctx, migrationsFS)
	assert.Error(t, err)
}

func TestComp_WorkOrderRepository_ListWithFilter(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewWorkOrderRepository(db)

	deviceID := "device-filter-comp"
	priority := "high"
	status := "open"
	assignee := "user1"
	createdBy := "admin"

	wo := &entity.WorkOrder{
		ID:        uuid.New().String(),
		DeviceID:  deviceID,
		Type:      "maintenance",
		Title:     "Filtered WO",
		Priority:  priority,
		Status:    status,
		Assignee:  assignee,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.Create(ctx, wo)

	filter := &service.WorkOrderFilter{
		DeviceID:  &deviceID,
		Priority:  &priority,
		Status:    status,
		Assignee:  &assignee,
		CreatedBy: &createdBy,
	}

	results, err := repo.List(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))

	count, err := repo.Count(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestComp_OperationLogRepository_List(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewOperationLogRepository(db)

	for i := 0; i < 5; i++ {
		log := &entity.OperationLog{
			ID:           uuid.New().String(),
			UserID:       "user-op-comp",
			Username:     "testuser",
			Action:       "create",
			ResourceType: "device",
			IPAddress:    "127.0.0.1",
			CreatedAt:    time.Now(),
		}
		repo.Create(ctx, log)
	}

	logs, total, err := repo.List(ctx, &repository.OperationLogQuery{
		Page:     1,
		PageSize: 3,
		UserID:   "user-op-comp",
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(5))
	assert.LessOrEqual(t, len(logs), 3)
}

func TestComp_SystemConfigRepository_GetByCategory(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewSystemConfigRepository(db)

	config := &entity.SystemConfig{
		ID:        uuid.New().String(),
		Key:       "sys.comp.test",
		Value:     "test_value",
		Category:  "test_category",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.Create(ctx, config)

	configs, err := repo.GetByCategory(ctx, "test_category")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(configs), 1)
}

func TestComp_ForecastResultRepository_ListByStation(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewForecastResultRepository(db)

	stationID := uuid.New().String()
	for i := 0; i < 3; i++ {
		result := &entity.ForecastResult{
			ID:             uuid.New().String(),
			StationID:      stationID,
			ForecastType:   entity.ForecastTypeShortTerm,
			TargetTime:     time.Now().Add(time.Duration(i) * time.Hour),
			PredictedPower: 100.0 + float64(i)*10,
			ModelVersion:   "v1.0.0",
			CreatedAt:      time.Now(),
		}
		repo.Create(ctx, result)
	}

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now().Add(4 * time.Hour)
	results, total, err := repo.ListByStation(ctx, stationID, nil, &start, &end, 0, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, int(total), 1)
	assert.GreaterOrEqual(t, len(results), 1)
}

func TestComp_ForecastResultRepository_UpdateActualPower(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewForecastResultRepository(db)

	result := &entity.ForecastResult{
		ID:             uuid.New().String(),
		StationID:      uuid.New().String(),
		ForecastType:   entity.ForecastTypeShortTerm,
		TargetTime:     time.Now(),
		PredictedPower: 95.0,
		ModelVersion:   "v1.0.0",
		CreatedAt:      time.Now(),
	}
	repo.Create(ctx, result)

	actualPower := 100.5
	err := repo.UpdateActualPower(ctx, result.ID, actualPower)
	require.NoError(t, err)

	got, _ := repo.GetByID(ctx, result.ID)
	assert.NotNil(t, got.ActualPower)
	assert.Equal(t, actualPower, *got.ActualPower)
}

func TestComp_FaultDetectionResultRepository_ListByDevice(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewFaultDetectionResultRepository(db)

	deviceID := uuid.New().String()
	for i := 0; i < 3; i++ {
		result := &entity.FaultDetectionResult{
			ID:           uuid.New().String(),
			DeviceID:     deviceID,
			FaultType:    "overheating",
			Severity:     entity.FaultSeverityCritical,
			Confidence:   0.85,
			Status:       entity.FaultDetectionStatusPending,
			ModelVersion: "v1.0.0",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		repo.Create(ctx, result)
	}

	results, total, err := repo.ListByDevice(ctx, deviceID, nil, nil, 0, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, int(total), 1)
	assert.GreaterOrEqual(t, len(results), 1)
}

func TestComp_FaultDetectionResultRepository_ListByStation(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	stationRepo := NewStationRepository(db)
	deviceRepo := NewDeviceRepository(db)
	repo := NewFaultDetectionResultRepository(db)

	station := &entity.Station{
		ID:          uuid.New().String(),
		Code:        "STA-FD-COMP",
		Name:        "FD Station",
		Type:        entity.StationTypePV,
		SubRegionID: uuid.New().String(),
		Status:      entity.StationStatusActive,
		Capacity:    100.0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	stationRepo.Create(ctx, station)

	device := &entity.Device{
		ID:        uuid.New().String(),
		StationID: station.ID,
		Code:      "DEV-FD-COMP",
		Name:      "FD Device",
		Type:      entity.DeviceTypeInverter,
		Status:    entity.DeviceStatusOnline,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	deviceRepo.Create(ctx, device)

	result := &entity.FaultDetectionResult{
		ID:           uuid.New().String(),
		DeviceID:     device.ID,
		FaultType:    "degradation",
		Severity:     entity.FaultSeverityWarning,
		Confidence:   0.75,
		Status:       entity.FaultDetectionStatusPending,
		ModelVersion: "v1.0.0",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	repo.Create(ctx, result)

	results, total, err := repo.ListByStation(ctx, station.ID, nil, 0, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, int(total), 1)
	assert.GreaterOrEqual(t, len(results), 1)
}

func TestComp_FaultDetectionResultRepository_UpdateStatus(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewFaultDetectionResultRepository(db)

	result := &entity.FaultDetectionResult{
		ID:           uuid.New().String(),
		DeviceID:     uuid.New().String(),
		FaultType:    "vibration",
		Severity:     entity.FaultSeverityInfo,
		Confidence:   0.60,
		Status:       entity.FaultDetectionStatusPending,
		ModelVersion: "v1.0.0",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	repo.Create(ctx, result)

	err := repo.UpdateStatus(ctx, result.ID, entity.FaultDetectionStatusConfirmed)
	require.NoError(t, err)
}

func TestComp_FaultDetectionResultRepository_UpdateRootCause(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewFaultDetectionResultRepository(db)

	result := &entity.FaultDetectionResult{
		ID:           uuid.New().String(),
		DeviceID:     uuid.New().String(),
		FaultType:    "overheating",
		Severity:     entity.FaultSeverityCritical,
		Confidence:   0.90,
		Status:       entity.FaultDetectionStatusPending,
		ModelVersion: "v1.0.0",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	repo.Create(ctx, result)

	err := repo.UpdateRootCause(ctx, result.ID, "Dust buildup in cooling system")
	require.NoError(t, err)
}

func TestComp_FaultDetectionResultRepository_LinkWorkOrder(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewFaultDetectionResultRepository(db)

	result := &entity.FaultDetectionResult{
		ID:           uuid.New().String(),
		DeviceID:     uuid.New().String(),
		FaultType:    "electrical",
		Severity:     entity.FaultSeverityCritical,
		Confidence:   0.95,
		Status:       entity.FaultDetectionStatusPending,
		ModelVersion: "v1.0.0",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	repo.Create(ctx, result)

	woID := uuid.New().String()
	err := repo.LinkWorkOrder(ctx, result.ID, woID)
	require.NoError(t, err)
}

func TestComp_FaultDetectionResultRepository_CountBySeverity(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewFaultDetectionResultRepository(db)

	deviceID := uuid.New().String()
	for _, sev := range []entity.FaultSeverity{entity.FaultSeverityInfo, entity.FaultSeverityWarning, entity.FaultSeverityCritical} {
		result := &entity.FaultDetectionResult{
			ID:           uuid.New().String(),
			DeviceID:     deviceID,
			FaultType:    "test",
			Severity:     sev,
			Confidence:   0.80,
			Status:       entity.FaultDetectionStatusPending,
			ModelVersion: "v1.0.0",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		repo.Create(ctx, result)
	}

	counts, err := repo.CountBySeverity(ctx, &deviceID)
	require.NoError(t, err)
	assert.NotNil(t, counts)
}

func TestComp_ModelVersionRepository_GetProductionModel(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewModelVersionRepository(db)

	accuracy := 0.95
	mv := &entity.ModelVersion{
		ID:           uuid.New().String(),
		ModelName:    "solar_prod_comp",
		Version:      "v1.0.0",
		Accuracy:     &accuracy,
		Status:       entity.ModelStatusProd,
		ArtifactPath: "/models/v1.onnx",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	repo.Create(ctx, mv)

	got, err := repo.GetProductionModel(ctx, "solar_prod_comp")
	require.NoError(t, err)
	assert.Equal(t, "v1.0.0", got.Version)
}

func TestComp_ModelVersionRepository_ListByModel(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewModelVersionRepository(db)

	accuracy := 0.95
	for i := 0; i < 3; i++ {
		mv := &entity.ModelVersion{
			ID:           uuid.New().String(),
			ModelName:    "solar_list_comp",
			Version:      "v1." + string(rune('0'+i)) + ".0",
			Accuracy:     &accuracy,
			Status:       entity.ModelStatusStaging,
			ArtifactPath: "/models/a" + string(rune('0'+i)) + ".onnx",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		repo.Create(ctx, mv)
	}

	results, err := repo.ListByModel(ctx, "solar_list_comp")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 3)
}

func TestComp_ModelVersionRepository_UpdateStatus(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewModelVersionRepository(db)

	accuracy := 0.95
	mv := &entity.ModelVersion{
		ID:           uuid.New().String(),
		ModelName:    "solar_status_comp",
		Version:      "v2.0.0",
		Accuracy:     &accuracy,
		Status:       entity.ModelStatusStaging,
		ArtifactPath: "/models/v2.onnx",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	repo.Create(ctx, mv)

	err := repo.UpdateStatus(ctx, mv.ID, entity.ModelStatusProd)
	require.NoError(t, err)
}

func TestComp_ForecastResultRepository_GetAccuracyStats(t *testing.T) {
	db := setupCompTestDB(t)
	ctx := context.Background()
	repo := NewForecastResultRepository(db)

	stationID := uuid.New().String()
	actualPower := 100.0
	accuracy := 0.95
	for i := 0; i < 3; i++ {
		result := &entity.ForecastResult{
			ID:             uuid.New().String(),
			StationID:      stationID,
			ForecastType:   entity.ForecastTypeShortTerm,
			TargetTime:     time.Now().Add(time.Duration(i) * time.Hour),
			PredictedPower: 100.0 + float64(i)*5,
			ActualPower:    &actualPower,
			Accuracy:       &accuracy,
			ModelVersion:   "v1.0.0",
			CreatedAt:      time.Now(),
		}
		repo.Create(ctx, result)
	}

	stats, err := repo.GetAccuracyStats(ctx, stationID, nil, time.Now().Add(-1*time.Hour), time.Now().Add(4*time.Hour))
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.TotalPoints, int64(1))
}
