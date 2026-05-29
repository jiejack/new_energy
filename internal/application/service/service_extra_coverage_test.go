package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCovAssetRepo struct {
	getByIDFn func(ctx context.Context, id string) (*entity.Asset, error)
	createFn  func(ctx context.Context, a *entity.Asset) error
	updateFn  func(ctx context.Context, a *entity.Asset) error
	deleteFn  func(ctx context.Context, id string) error
}

func (m *mockCovAssetRepo) Create(ctx context.Context, a *entity.Asset) error {
	if m.createFn != nil {
		return m.createFn(ctx, a)
	}
	return nil
}
func (m *mockCovAssetRepo) Update(ctx context.Context, a *entity.Asset) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, a)
	}
	return nil
}
func (m *mockCovAssetRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}
func (m *mockCovAssetRepo) GetByID(ctx context.Context, id string) (*entity.Asset, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return &entity.Asset{ID: id}, nil
}
func (m *mockCovAssetRepo) GetByCode(ctx context.Context, code string) (*entity.Asset, error) { return nil, errors.New("not found") }
func (m *mockCovAssetRepo) List(ctx context.Context, assetType *string, status *string, category *string, offset, limit int) ([]*entity.Asset, int64, error) { return nil, 0, nil }
func (m *mockCovAssetRepo) GetByLocation(ctx context.Context, location string) ([]*entity.Asset, error) { return nil, nil }
func (m *mockCovAssetRepo) GetByDepartment(ctx context.Context, departmentID string) ([]*entity.Asset, error) { return nil, nil }
func (m *mockCovAssetRepo) GetByResponsiblePerson(ctx context.Context, person string) ([]*entity.Asset, error) { return nil, nil }
func (m *mockCovAssetRepo) GetDepreciatingAssets(ctx context.Context) ([]*entity.Asset, error) { return nil, nil }
func (m *mockCovAssetRepo) GetAssetsNearWarrantyEnd(ctx context.Context, days int) ([]*entity.Asset, error) { return nil, nil }

type mockCovMaintenanceRepo struct {
	getByIDFn func(ctx context.Context, id string) (*entity.AssetMaintenanceRecord, error)
	createFn  func(ctx context.Context, r *entity.AssetMaintenanceRecord) error
	updateFn  func(ctx context.Context, r *entity.AssetMaintenanceRecord) error
	deleteFn  func(ctx context.Context, id string) error
	listFn    func(ctx context.Context, assetID string, status *string, mt *string, offset, limit int) ([]*entity.AssetMaintenanceRecord, int64, error)
	costFn    func(ctx context.Context, assetID string, start, end *time.Time) (float64, error)
}

func (m *mockCovMaintenanceRepo) Create(ctx context.Context, r *entity.AssetMaintenanceRecord) error {
	if m.createFn != nil { return m.createFn(ctx, r) }
	return nil
}
func (m *mockCovMaintenanceRepo) Update(ctx context.Context, r *entity.AssetMaintenanceRecord) error {
	if m.updateFn != nil { return m.updateFn(ctx, r) }
	return nil
}
func (m *mockCovMaintenanceRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil { return m.deleteFn(ctx, id) }
	return nil
}
func (m *mockCovMaintenanceRepo) GetByID(ctx context.Context, id string) (*entity.AssetMaintenanceRecord, error) {
	if m.getByIDFn != nil { return m.getByIDFn(ctx, id) }
	return nil, errors.New("not found")
}
func (m *mockCovMaintenanceRepo) ListByAssetID(ctx context.Context, assetID string, status *string, mt *string, offset, limit int) ([]*entity.AssetMaintenanceRecord, int64, error) {
	if m.listFn != nil { return m.listFn(ctx, assetID, status, mt, offset, limit) }
	return nil, 0, nil
}
func (m *mockCovMaintenanceRepo) ListByStatus(ctx context.Context, status string, offset, limit int) ([]*entity.AssetMaintenanceRecord, int64, error) { return nil, 0, nil }
func (m *mockCovMaintenanceRepo) GetMaintenanceCostByAsset(ctx context.Context, assetID string, start, end *time.Time) (float64, error) {
	if m.costFn != nil { return m.costFn(ctx, assetID, start, end) }
	return 0, nil
}

type mockCovDepreciationRepo struct {
	getByIDFn func(ctx context.Context, id string) (*entity.AssetDepreciationRecord, error)
	createFn  func(ctx context.Context, r *entity.AssetDepreciationRecord) error
	updateFn  func(ctx context.Context, r *entity.AssetDepreciationRecord) error
	deleteFn  func(ctx context.Context, id string) error
	listFn    func(ctx context.Context, assetID string, period *string, offset, limit int) ([]*entity.AssetDepreciationRecord, int64, error)
	summaryFn func(ctx context.Context, period string, start, end *time.Time) (float64, error)
}

func (m *mockCovDepreciationRepo) Create(ctx context.Context, r *entity.AssetDepreciationRecord) error {
	if m.createFn != nil { return m.createFn(ctx, r) }
	return nil
}
func (m *mockCovDepreciationRepo) Update(ctx context.Context, r *entity.AssetDepreciationRecord) error {
	if m.updateFn != nil { return m.updateFn(ctx, r) }
	return nil
}
func (m *mockCovDepreciationRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil { return m.deleteFn(ctx, id) }
	return nil
}
func (m *mockCovDepreciationRepo) GetByID(ctx context.Context, id string) (*entity.AssetDepreciationRecord, error) {
	if m.getByIDFn != nil { return m.getByIDFn(ctx, id) }
	return nil, errors.New("not found")
}
func (m *mockCovDepreciationRepo) ListByAssetID(ctx context.Context, assetID string, period *string, offset, limit int) ([]*entity.AssetDepreciationRecord, int64, error) {
	if m.listFn != nil { return m.listFn(ctx, assetID, period, offset, limit) }
	return nil, 0, nil
}
func (m *mockCovDepreciationRepo) GetLatestByAssetID(ctx context.Context, assetID string) (*entity.AssetDepreciationRecord, error) { return nil, nil }
func (m *mockCovDepreciationRepo) GetDepreciationSummaryByPeriod(ctx context.Context, period string, start, end *time.Time) (float64, error) {
	if m.summaryFn != nil { return m.summaryFn(ctx, period, start, end) }
	return 100.0, nil
}

type mockCovDocumentRepo struct {
	getByIDFn func(ctx context.Context, id string) (*entity.AssetDocument, error)
	createFn  func(ctx context.Context, d *entity.AssetDocument) error
	updateFn  func(ctx context.Context, d *entity.AssetDocument) error
	deleteFn  func(ctx context.Context, id string) error
	listFn    func(ctx context.Context, assetID string, dt *string, offset, limit int) ([]*entity.AssetDocument, int64, error)
}

func (m *mockCovDocumentRepo) Create(ctx context.Context, d *entity.AssetDocument) error {
	if m.createFn != nil { return m.createFn(ctx, d) }
	return nil
}
func (m *mockCovDocumentRepo) Update(ctx context.Context, d *entity.AssetDocument) error {
	if m.updateFn != nil { return m.updateFn(ctx, d) }
	return nil
}
func (m *mockCovDocumentRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil { return m.deleteFn(ctx, id) }
	return nil
}
func (m *mockCovDocumentRepo) GetByID(ctx context.Context, id string) (*entity.AssetDocument, error) {
	if m.getByIDFn != nil { return m.getByIDFn(ctx, id) }
	return nil, errors.New("not found")
}
func (m *mockCovDocumentRepo) ListByAssetID(ctx context.Context, assetID string, dt *string, offset, limit int) ([]*entity.AssetDocument, int64, error) {
	if m.listFn != nil { return m.listFn(ctx, assetID, dt, offset, limit) }
	return nil, 0, nil
}
func (m *mockCovDocumentRepo) GetByType(ctx context.Context, documentType string, offset, limit int) ([]*entity.AssetDocument, int64, error) { return nil, 0, nil }

type mockCovInventoryRepo struct {
	getByIDFn   func(ctx context.Context, id string) (*entity.Inventory, error)
	getByCodeFn func(ctx context.Context, code string) (*entity.Inventory, error)
	createFn    func(ctx context.Context, i *entity.Inventory) error
	updateFn    func(ctx context.Context, i *entity.Inventory) error
	deleteFn    func(ctx context.Context, id string) error
	lowStockFn  func(ctx context.Context) ([]*entity.Inventory, error)
}

func (m *mockCovInventoryRepo) Create(ctx context.Context, i *entity.Inventory) error {
	if m.createFn != nil { return m.createFn(ctx, i) }
	return nil
}
func (m *mockCovInventoryRepo) Update(ctx context.Context, i *entity.Inventory) error {
	if m.updateFn != nil { return m.updateFn(ctx, i) }
	return nil
}
func (m *mockCovInventoryRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil { return m.deleteFn(ctx, id) }
	return nil
}
func (m *mockCovInventoryRepo) GetByID(ctx context.Context, id string) (*entity.Inventory, error) {
	if m.getByIDFn != nil { return m.getByIDFn(ctx, id) }
	return nil, errors.New("not found")
}
func (m *mockCovInventoryRepo) GetByCode(ctx context.Context, code string) (*entity.Inventory, error) {
	if m.getByCodeFn != nil { return m.getByCodeFn(ctx, code) }
	return nil, errors.New("not found")
}
func (m *mockCovInventoryRepo) List(ctx context.Context, filter interface{}) ([]*entity.Inventory, error) { return nil, nil }
func (m *mockCovInventoryRepo) Count(ctx context.Context, filter interface{}) (int64, error) { return 0, nil }
func (m *mockCovInventoryRepo) UpdateQuantity(ctx context.Context, id string, quantity float64) error { return nil }
func (m *mockCovInventoryRepo) GetLowStockItems(ctx context.Context) ([]*entity.Inventory, error) {
	if m.lowStockFn != nil { return m.lowStockFn(ctx) }
	return nil, nil
}

type mockCovSupplierRepo struct {
	getByIDFn   func(ctx context.Context, id string) (*entity.Supplier, error)
	getByCodeFn func(ctx context.Context, code string) (*entity.Supplier, error)
	createFn    func(ctx context.Context, s *entity.Supplier) error
	updateFn    func(ctx context.Context, s *entity.Supplier) error
	deleteFn    func(ctx context.Context, id string) error
}

func (m *mockCovSupplierRepo) Create(ctx context.Context, s *entity.Supplier) error {
	if m.createFn != nil { return m.createFn(ctx, s) }
	return nil
}
func (m *mockCovSupplierRepo) Update(ctx context.Context, s *entity.Supplier) error {
	if m.updateFn != nil { return m.updateFn(ctx, s) }
	return nil
}
func (m *mockCovSupplierRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil { return m.deleteFn(ctx, id) }
	return nil
}
func (m *mockCovSupplierRepo) GetByID(ctx context.Context, id string) (*entity.Supplier, error) {
	if m.getByIDFn != nil { return m.getByIDFn(ctx, id) }
	return nil, errors.New("not found")
}
func (m *mockCovSupplierRepo) GetByCode(ctx context.Context, code string) (*entity.Supplier, error) {
	if m.getByCodeFn != nil { return m.getByCodeFn(ctx, code) }
	return nil, errors.New("not found")
}
func (m *mockCovSupplierRepo) List(ctx context.Context, filter interface{}) ([]*entity.Supplier, error) { return nil, nil }
func (m *mockCovSupplierRepo) Count(ctx context.Context, filter interface{}) (int64, error) { return 0, nil }

type mockCovPurchaseOrderRepo struct {
	getByIDFn   func(ctx context.Context, id string) (*entity.PurchaseOrder, error)
	getByCodeFn func(ctx context.Context, code string) (*entity.PurchaseOrder, error)
	createFn    func(ctx context.Context, o *entity.PurchaseOrder) error
	updateFn    func(ctx context.Context, o *entity.PurchaseOrder) error
	deleteFn    func(ctx context.Context, id string) error
	listFn      func(ctx context.Context, supplierID *string, status *string, start, end *time.Time, offset, limit int) ([]*entity.PurchaseOrder, int64, error)
}

func (m *mockCovPurchaseOrderRepo) Create(ctx context.Context, o *entity.PurchaseOrder) error {
	if m.createFn != nil { return m.createFn(ctx, o) }
	return nil
}
func (m *mockCovPurchaseOrderRepo) Update(ctx context.Context, o *entity.PurchaseOrder) error {
	if m.updateFn != nil { return m.updateFn(ctx, o) }
	return nil
}
func (m *mockCovPurchaseOrderRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil { return m.deleteFn(ctx, id) }
	return nil
}
func (m *mockCovPurchaseOrderRepo) GetByID(ctx context.Context, id string) (*entity.PurchaseOrder, error) {
	if m.getByIDFn != nil { return m.getByIDFn(ctx, id) }
	return nil, errors.New("not found")
}
func (m *mockCovPurchaseOrderRepo) GetByCode(ctx context.Context, code string) (*entity.PurchaseOrder, error) {
	if m.getByCodeFn != nil { return m.getByCodeFn(ctx, code) }
	return nil, errors.New("not found")
}
func (m *mockCovPurchaseOrderRepo) List(ctx context.Context, supplierID *string, status *string, start, end *time.Time, offset, limit int) ([]*entity.PurchaseOrder, int64, error) {
	if m.listFn != nil { return m.listFn(ctx, supplierID, status, start, end, offset, limit) }
	return nil, 0, nil
}

type mockCovReceiptRepo struct {
	getByIDFn           func(ctx context.Context, id string) (*entity.Receipt, error)
	getByCodeFn         func(ctx context.Context, code string) (*entity.Receipt, error)
	getByPOIDFn         func(ctx context.Context, poID string) ([]*entity.Receipt, error)
	createFn            func(ctx context.Context, r *entity.Receipt) error
	updateFn            func(ctx context.Context, r *entity.Receipt) error
	deleteFn            func(ctx context.Context, id string) error
	listFn              func(ctx context.Context, poID *string, status *string, start, end *time.Time, offset, limit int) ([]*entity.Receipt, int64, error)
}

func (m *mockCovReceiptRepo) Create(ctx context.Context, r *entity.Receipt) error {
	if m.createFn != nil { return m.createFn(ctx, r) }
	return nil
}
func (m *mockCovReceiptRepo) Update(ctx context.Context, r *entity.Receipt) error {
	if m.updateFn != nil { return m.updateFn(ctx, r) }
	return nil
}
func (m *mockCovReceiptRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil { return m.deleteFn(ctx, id) }
	return nil
}
func (m *mockCovReceiptRepo) GetByID(ctx context.Context, id string) (*entity.Receipt, error) {
	if m.getByIDFn != nil { return m.getByIDFn(ctx, id) }
	return nil, errors.New("not found")
}
func (m *mockCovReceiptRepo) GetByCode(ctx context.Context, code string) (*entity.Receipt, error) {
	if m.getByCodeFn != nil { return m.getByCodeFn(ctx, code) }
	return nil, errors.New("not found")
}
func (m *mockCovReceiptRepo) GetByPurchaseOrderID(ctx context.Context, poID string) ([]*entity.Receipt, error) {
	if m.getByPOIDFn != nil { return m.getByPOIDFn(ctx, poID) }
	return nil, nil
}
func (m *mockCovReceiptRepo) List(ctx context.Context, poID *string, status *string, start, end *time.Time, offset, limit int) ([]*entity.Receipt, int64, error) {
	if m.listFn != nil { return m.listFn(ctx, poID, status, start, end, offset, limit) }
	return nil, 0, nil
}

type mockCovCostCategoryRepo struct {
	getByIDFn   func(ctx context.Context, id string) (*entity.CostCategory, error)
	getByCodeFn func(ctx context.Context, code string) (*entity.CostCategory, error)
	createFn    func(ctx context.Context, c *entity.CostCategory) error
	updateFn    func(ctx context.Context, c *entity.CostCategory) error
	deleteFn    func(ctx context.Context, id string) error
	listFn      func(ctx context.Context, parentID *string, status *string) ([]*entity.CostCategory, error)
	treeFn      func(ctx context.Context) ([]*entity.CostCategory, error)
}

func (m *mockCovCostCategoryRepo) Create(ctx context.Context, c *entity.CostCategory) error {
	if m.createFn != nil { return m.createFn(ctx, c) }
	return nil
}
func (m *mockCovCostCategoryRepo) Update(ctx context.Context, c *entity.CostCategory) error {
	if m.updateFn != nil { return m.updateFn(ctx, c) }
	return nil
}
func (m *mockCovCostCategoryRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil { return m.deleteFn(ctx, id) }
	return nil
}
func (m *mockCovCostCategoryRepo) GetByID(ctx context.Context, id string) (*entity.CostCategory, error) {
	if m.getByIDFn != nil { return m.getByIDFn(ctx, id) }
	return nil, errors.New("not found")
}
func (m *mockCovCostCategoryRepo) GetByCode(ctx context.Context, code string) (*entity.CostCategory, error) {
	if m.getByCodeFn != nil { return m.getByCodeFn(ctx, code) }
	return nil, errors.New("not found")
}
func (m *mockCovCostCategoryRepo) List(ctx context.Context, parentID *string, status *string) ([]*entity.CostCategory, error) {
	if m.listFn != nil { return m.listFn(ctx, parentID, status) }
	return nil, nil
}
func (m *mockCovCostCategoryRepo) GetTree(ctx context.Context) ([]*entity.CostCategory, error) {
	if m.treeFn != nil { return m.treeFn(ctx) }
	return nil, nil
}

type mockCovCostEntryRepo struct {
	getByIDFn     func(ctx context.Context, id string) (*entity.CostEntry, error)
	getByCodeFn   func(ctx context.Context, code string) (*entity.CostEntry, error)
	createFn      func(ctx context.Context, e *entity.CostEntry) error
	updateFn      func(ctx context.Context, e *entity.CostEntry) error
	deleteFn      func(ctx context.Context, id string) error
	listFn        func(ctx context.Context, catID *string, start, end *time.Time, status *string, offset, limit int) ([]*entity.CostEntry, int64, error)
	totalCatFn    func(ctx context.Context, catID string, start, end *time.Time) (float64, error)
	totalPeriodFn func(ctx context.Context, start, end *time.Time) (float64, error)
}

func (m *mockCovCostEntryRepo) Create(ctx context.Context, e *entity.CostEntry) error {
	if m.createFn != nil { return m.createFn(ctx, e) }
	return nil
}
func (m *mockCovCostEntryRepo) Update(ctx context.Context, e *entity.CostEntry) error {
	if m.updateFn != nil { return m.updateFn(ctx, e) }
	return nil
}
func (m *mockCovCostEntryRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil { return m.deleteFn(ctx, id) }
	return nil
}
func (m *mockCovCostEntryRepo) GetByID(ctx context.Context, id string) (*entity.CostEntry, error) {
	if m.getByIDFn != nil { return m.getByIDFn(ctx, id) }
	return nil, errors.New("not found")
}
func (m *mockCovCostEntryRepo) GetByCode(ctx context.Context, code string) (*entity.CostEntry, error) {
	if m.getByCodeFn != nil { return m.getByCodeFn(ctx, code) }
	return nil, errors.New("not found")
}
func (m *mockCovCostEntryRepo) List(ctx context.Context, catID *string, start, end *time.Time, status *string, offset, limit int) ([]*entity.CostEntry, int64, error) {
	if m.listFn != nil { return m.listFn(ctx, catID, start, end, status, offset, limit) }
	return nil, 0, nil
}
func (m *mockCovCostEntryRepo) GetTotalByCategory(ctx context.Context, catID string, start, end *time.Time) (float64, error) {
	if m.totalCatFn != nil { return m.totalCatFn(ctx, catID, start, end) }
	return 0, nil
}
func (m *mockCovCostEntryRepo) GetTotalByPeriod(ctx context.Context, start, end *time.Time) (float64, error) {
	if m.totalPeriodFn != nil { return m.totalPeriodFn(ctx, start, end) }
	return 0, nil
}

type mockCovCostAllocationRepo struct {
	getByIDFn    func(ctx context.Context, id string) (*entity.CostAllocation, error)
	createFn     func(ctx context.Context, a *entity.CostAllocation) error
	updateFn     func(ctx context.Context, a *entity.CostAllocation) error
	deleteFn     func(ctx context.Context, id string) error
	listByEntryFn func(ctx context.Context, entryID string) ([]*entity.CostAllocation, error)
	totalAllocFn func(ctx context.Context, allocatedTo, allocatedID string, start, end *time.Time) (float64, error)
}

func (m *mockCovCostAllocationRepo) Create(ctx context.Context, a *entity.CostAllocation) error {
	if m.createFn != nil { return m.createFn(ctx, a) }
	return nil
}
func (m *mockCovCostAllocationRepo) Update(ctx context.Context, a *entity.CostAllocation) error {
	if m.updateFn != nil { return m.updateFn(ctx, a) }
	return nil
}
func (m *mockCovCostAllocationRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil { return m.deleteFn(ctx, id) }
	return nil
}
func (m *mockCovCostAllocationRepo) GetByID(ctx context.Context, id string) (*entity.CostAllocation, error) {
	if m.getByIDFn != nil { return m.getByIDFn(ctx, id) }
	return nil, errors.New("not found")
}
func (m *mockCovCostAllocationRepo) ListByCostEntryID(ctx context.Context, entryID string) ([]*entity.CostAllocation, error) {
	if m.listByEntryFn != nil { return m.listByEntryFn(ctx, entryID) }
	return nil, nil
}
func (m *mockCovCostAllocationRepo) ListByAllocated(ctx context.Context, allocatedTo, allocatedID string) ([]*entity.CostAllocation, error) { return nil, nil }
func (m *mockCovCostAllocationRepo) GetTotalByAllocated(ctx context.Context, allocatedTo, allocatedID string, start, end *time.Time) (float64, error) {
	if m.totalAllocFn != nil { return m.totalAllocFn(ctx, allocatedTo, allocatedID, start, end) }
	return 0, nil
}

type mockCovCostReportRepo struct {
	getByIDFn    func(ctx context.Context, id string) (*entity.CostReport, error)
	getByCodeFn  func(ctx context.Context, code string) (*entity.CostReport, error)
	getByPeriodFn func(ctx context.Context, reportType string, start, end time.Time) (*entity.CostReport, error)
	createFn     func(ctx context.Context, r *entity.CostReport) error
	updateFn     func(ctx context.Context, r *entity.CostReport) error
	deleteFn     func(ctx context.Context, id string) error
	listFn       func(ctx context.Context, reportType *string, status *string, start, end *time.Time, offset, limit int) ([]*entity.CostReport, int64, error)
}

func (m *mockCovCostReportRepo) Create(ctx context.Context, r *entity.CostReport) error {
	if m.createFn != nil { return m.createFn(ctx, r) }
	return nil
}
func (m *mockCovCostReportRepo) Update(ctx context.Context, r *entity.CostReport) error {
	if m.updateFn != nil { return m.updateFn(ctx, r) }
	return nil
}
func (m *mockCovCostReportRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil { return m.deleteFn(ctx, id) }
	return nil
}
func (m *mockCovCostReportRepo) GetByID(ctx context.Context, id string) (*entity.CostReport, error) {
	if m.getByIDFn != nil { return m.getByIDFn(ctx, id) }
	return nil, errors.New("not found")
}
func (m *mockCovCostReportRepo) GetByCode(ctx context.Context, code string) (*entity.CostReport, error) {
	if m.getByCodeFn != nil { return m.getByCodeFn(ctx, code) }
	return nil, errors.New("not found")
}
func (m *mockCovCostReportRepo) GetByPeriod(ctx context.Context, reportType string, start, end time.Time) (*entity.CostReport, error) {
	if m.getByPeriodFn != nil { return m.getByPeriodFn(ctx, reportType, start, end) }
	return nil, errors.New("not found")
}
func (m *mockCovCostReportRepo) List(ctx context.Context, reportType *string, status *string, start, end *time.Time, offset, limit int) ([]*entity.CostReport, int64, error) {
	if m.listFn != nil { return m.listFn(ctx, reportType, status, start, end, offset, limit) }
	return nil, 0, nil
}

type mockCovCarbonEmissionRepo struct {
	factorByIDFn    func(ctx context.Context, id string) (*entity.CarbonEmissionFactor, error)
	factorByCodeFn  func(ctx context.Context, code string) (*entity.CarbonEmissionFactor, error)
	createFactorFn  func(ctx context.Context, f *entity.CarbonEmissionFactor) error
	updateFactorFn  func(ctx context.Context, f *entity.CarbonEmissionFactor) error
	recordByIDFn     func(ctx context.Context, id string) (*entity.CarbonEmissionRecord, error)
	createRecordFn   func(ctx context.Context, r *entity.CarbonEmissionRecord) error
	batchRecordsFn   func(ctx context.Context, records []*entity.CarbonEmissionRecord) error
	recordsByTimeFn  func(ctx context.Context, targetID string, scope *entity.CarbonEmissionScope, start, end time.Time) ([]*entity.CarbonEmissionRecord, error)
	summaryByIDFn    func(ctx context.Context, id string) (*entity.CarbonEmissionSummary, error)
	createSummaryFn  func(ctx context.Context, s *entity.CarbonEmissionSummary) error
	targetByIDFn     func(ctx context.Context, id string) (*entity.CarbonReductionTarget, error)
	createTargetFn   func(ctx context.Context, t *entity.CarbonReductionTarget) error
	updateTargetFn   func(ctx context.Context, t *entity.CarbonReductionTarget) error
	totalScopeFn     func(ctx context.Context, targetID string, scope entity.CarbonEmissionScope, period string, start, end time.Time) (float64, error)
	latestSummaryFn  func(ctx context.Context, targetID string, period string) (*entity.CarbonEmissionSummary, error)
}

func (m *mockCovCarbonEmissionRepo) GetFactorByID(ctx context.Context, id string) (*entity.CarbonEmissionFactor, error) {
	if m.factorByIDFn != nil { return m.factorByIDFn(ctx, id) }
	return nil, errors.New("not found")
}
func (m *mockCovCarbonEmissionRepo) GetFactorByCode(ctx context.Context, code string) (*entity.CarbonEmissionFactor, error) {
	if m.factorByCodeFn != nil { return m.factorByCodeFn(ctx, code) }
	return nil, errors.New("not found")
}
func (m *mockCovCarbonEmissionRepo) CreateFactor(ctx context.Context, f *entity.CarbonEmissionFactor) error {
	if m.createFactorFn != nil { return m.createFactorFn(ctx, f) }
	return nil
}
func (m *mockCovCarbonEmissionRepo) UpdateFactor(ctx context.Context, f *entity.CarbonEmissionFactor) error {
	if m.updateFactorFn != nil { return m.updateFactorFn(ctx, f) }
	return nil
}
func (m *mockCovCarbonEmissionRepo) ListFactors(ctx context.Context, q *repository.CarbonEmissionFactorQuery) ([]*entity.CarbonEmissionFactor, int64, error) { return nil, 0, nil }
func (m *mockCovCarbonEmissionRepo) GetActiveFactors(ctx context.Context, scope *entity.CarbonEmissionScope) ([]*entity.CarbonEmissionFactor, error) { return nil, nil }
func (m *mockCovCarbonEmissionRepo) GetRecordByID(ctx context.Context, id string) (*entity.CarbonEmissionRecord, error) {
	if m.recordByIDFn != nil { return m.recordByIDFn(ctx, id) }
	return nil, errors.New("not found")
}
func (m *mockCovCarbonEmissionRepo) CreateRecord(ctx context.Context, r *entity.CarbonEmissionRecord) error {
	if m.createRecordFn != nil { return m.createRecordFn(ctx, r) }
	return nil
}
func (m *mockCovCarbonEmissionRepo) BatchCreateRecords(ctx context.Context, records []*entity.CarbonEmissionRecord) error {
	if m.batchRecordsFn != nil { return m.batchRecordsFn(ctx, records) }
	return nil
}
func (m *mockCovCarbonEmissionRepo) UpdateRecord(ctx context.Context, r *entity.CarbonEmissionRecord) error { return nil }
func (m *mockCovCarbonEmissionRepo) ListRecords(ctx context.Context, q *repository.CarbonEmissionRecordQuery) ([]*entity.CarbonEmissionRecord, int64, error) { return nil, 0, nil }
func (m *mockCovCarbonEmissionRepo) GetRecordsByTimeRange(ctx context.Context, targetID string, scope *entity.CarbonEmissionScope, start, end time.Time) ([]*entity.CarbonEmissionRecord, error) {
	if m.recordsByTimeFn != nil { return m.recordsByTimeFn(ctx, targetID, scope, start, end) }
	return nil, nil
}
func (m *mockCovCarbonEmissionRepo) GetSummaryByID(ctx context.Context, id string) (*entity.CarbonEmissionSummary, error) {
	if m.summaryByIDFn != nil { return m.summaryByIDFn(ctx, id) }
	return nil, errors.New("not found")
}
func (m *mockCovCarbonEmissionRepo) CreateSummary(ctx context.Context, s *entity.CarbonEmissionSummary) error {
	if m.createSummaryFn != nil { return m.createSummaryFn(ctx, s) }
	return nil
}
func (m *mockCovCarbonEmissionRepo) UpdateSummary(ctx context.Context, s *entity.CarbonEmissionSummary) error { return nil }
func (m *mockCovCarbonEmissionRepo) GetLatestSummary(ctx context.Context, targetID string, period string) (*entity.CarbonEmissionSummary, error) {
	if m.latestSummaryFn != nil { return m.latestSummaryFn(ctx, targetID, period) }
	return nil, nil
}
func (m *mockCovCarbonEmissionRepo) ListSummaries(ctx context.Context, q *repository.CarbonEmissionSummaryQuery) ([]*entity.CarbonEmissionSummary, int64, error) { return nil, 0, nil }
func (m *mockCovCarbonEmissionRepo) GetTargetByID(ctx context.Context, id string) (*entity.CarbonReductionTarget, error) {
	if m.targetByIDFn != nil { return m.targetByIDFn(ctx, id) }
	return nil, errors.New("not found")
}
func (m *mockCovCarbonEmissionRepo) CreateTarget(ctx context.Context, t *entity.CarbonReductionTarget) error {
	if m.createTargetFn != nil { return m.createTargetFn(ctx, t) }
	return nil
}
func (m *mockCovCarbonEmissionRepo) UpdateTarget(ctx context.Context, t *entity.CarbonReductionTarget) error {
	if m.updateTargetFn != nil { return m.updateTargetFn(ctx, t) }
	return nil
}
func (m *mockCovCarbonEmissionRepo) ListTargets(ctx context.Context, q *repository.CarbonReductionTargetQuery) ([]*entity.CarbonReductionTarget, int64, error) { return nil, 0, nil }
func (m *mockCovCarbonEmissionRepo) GetActiveTargets(ctx context.Context, targetID string) ([]*entity.CarbonReductionTarget, error) { return nil, nil }
func (m *mockCovCarbonEmissionRepo) GetTotalEmissionByScope(ctx context.Context, targetID string, scope entity.CarbonEmissionScope, period string, start, end time.Time) (float64, error) {
	if m.totalScopeFn != nil { return m.totalScopeFn(ctx, targetID, scope, period, start, end) }
	return 0, nil
}

type mockCovFaultDetectionRepo struct {
	getByIDFn      func(ctx context.Context, id string) (*entity.FaultDetectionResult, error)
	createFn       func(ctx context.Context, r *entity.FaultDetectionResult) error
	linkWorkOrderFn func(ctx context.Context, detectionID, woID string) error
}

func (m *mockCovFaultDetectionRepo) Create(ctx context.Context, r *entity.FaultDetectionResult) error {
	if m.createFn != nil { return m.createFn(ctx, r) }
	return nil
}
func (m *mockCovFaultDetectionRepo) GetByID(ctx context.Context, id string) (*entity.FaultDetectionResult, error) {
	if m.getByIDFn != nil { return m.getByIDFn(ctx, id) }
	return nil, errors.New("not found")
}
func (m *mockCovFaultDetectionRepo) ListByDevice(ctx context.Context, deviceID string, severity *entity.FaultSeverity, status *entity.FaultDetectionStatus, offset, limit int) ([]*entity.FaultDetectionResult, int64, error) { return nil, 0, nil }
func (m *mockCovFaultDetectionRepo) ListByStation(ctx context.Context, stationID string, severity *entity.FaultSeverity, offset, limit int) ([]*entity.FaultDetectionResult, int64, error) { return nil, 0, nil }
func (m *mockCovFaultDetectionRepo) UpdateStatus(ctx context.Context, id string, status entity.FaultDetectionStatus) error { return nil }
func (m *mockCovFaultDetectionRepo) UpdateRootCause(ctx context.Context, id, rootCause string) error { return nil }
func (m *mockCovFaultDetectionRepo) LinkWorkOrder(ctx context.Context, detectionID, workOrderID string) error {
	if m.linkWorkOrderFn != nil { return m.linkWorkOrderFn(ctx, detectionID, workOrderID) }
	return nil
}
func (m *mockCovFaultDetectionRepo) CountBySeverity(ctx context.Context, deviceID *string) (map[entity.FaultSeverity]int64, error) { return nil, nil }

type mockCovForecastRepo struct {
	accuracyStatsFn func(ctx context.Context, stationID string, ft *entity.ForecastType, start, end time.Time) (*repository.ForecastAccuracyStats, error)
}

func (m *mockCovForecastRepo) Create(ctx context.Context, r *entity.ForecastResult) error { return nil }
func (m *mockCovForecastRepo) GetByID(ctx context.Context, id string) (*entity.ForecastResult, error) { return nil, nil }
func (m *mockCovForecastRepo) ListByStation(ctx context.Context, stationID string, ft *entity.ForecastType, start, end *time.Time, offset, limit int) ([]*entity.ForecastResult, int64, error) { return nil, 0, nil }
func (m *mockCovForecastRepo) UpdateActualPower(ctx context.Context, id string, actualPower float64) error { return nil }
func (m *mockCovForecastRepo) GetAccuracyStats(ctx context.Context, stationID string, ft *entity.ForecastType, start, end time.Time) (*repository.ForecastAccuracyStats, error) {
	if m.accuracyStatsFn != nil { return m.accuracyStatsFn(ctx, stationID, ft, start, end) }
	return nil, nil
}

type mockCovModelService struct{}

func (m *mockCovModelService) RegisterModel(ctx context.Context, modelName, version, artifactPath string, accuracy float64) (*entity.ModelVersion, error) { return nil, nil }
func (m *mockCovModelService) GetModel(ctx context.Context, id string) (*entity.ModelVersion, error) { return nil, nil }
func (m *mockCovModelService) GetProductionModel(ctx context.Context, modelName string) (*entity.ModelVersion, error) { return nil, nil }
func (m *mockCovModelService) ListModels(ctx context.Context, modelName string) ([]*entity.ModelVersion, error) { return nil, nil }
func (m *mockCovModelService) PromoteToProduction(ctx context.Context, id string) error { return nil }
func (m *mockCovModelService) RetireModel(ctx context.Context, id string) error { return nil }

type mockCovWorkOrderCreator struct {
	fn func(ctx context.Context, det *entity.FaultDetectionResult) (string, error)
}

func (m *mockCovWorkOrderCreator) CreateFromFaultDetection(ctx context.Context, det *entity.FaultDetectionResult) (string, error) {
	if m.fn != nil { return m.fn(ctx, det) }
	return "default-wo", nil
}

func covStrPtr(s string) *string { return &s }
func covFloatPtr(f float64) *float64 { return &f }
func covBoolPtr(b bool) *bool { return &b }

func TestAssetDepreciationService_Create(t *testing.T) {
	repo := &mockCovDepreciationRepo{}
	assetRepo := &mockCovAssetRepo{}
	svc := NewAssetDepreciationService(repo, assetRepo)

	req := &CreateDepreciationRequest{
		AssetID:             "asset-1",
		DepreciationMethod: "straight_line",
		Year:                2024,
		Amount:              1000.0,
		AccumulatedAmount:   5000.0,
		BookValue:          45000.0,
	}

	record, err := svc.CreateDepreciationRecord(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "asset-1", record.AssetID)
	assert.Equal(t, 1000.0, record.DepreciationAmount)
}

func TestAssetDepreciationService_CreateAssetNotFound(t *testing.T) {
	repo := &mockCovDepreciationRepo{}
	assetRepo := &mockCovAssetRepo{getByIDFn: func(_ context.Context, _ string) (*entity.Asset, error) { return nil, errors.New("not found") }}
	svc := NewAssetDepreciationService(repo, assetRepo)

	req := &CreateDepreciationRequest{
		AssetID:             "nonexistent",
		DepreciationMethod: "straight_line",
		Year:                2024,
		Amount:              1000.0,
		AccumulatedAmount:   5000.0,
		BookValue:          45000.0,
	}

	_, err := svc.CreateDepreciationRecord(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "资产不存在")
}

func TestAssetDepreciationService_Update(t *testing.T) {
	existing := &entity.AssetDepreciationRecord{ID: "dep-1", AssetID: "asset-1", Period: "annual"}
	repo := &mockCovDepreciationRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.AssetDepreciationRecord, error) { return existing, nil },
	}
	assetRepo := &mockCovAssetRepo{}
	svc := NewAssetDepreciationService(repo, assetRepo)

	req := &UpdateDepreciationRequest{Amount: 2000.0, AccumulatedAmount: 7000.0, BookValue: 43000.0}
	record, err := svc.UpdateDepreciationRecord(context.Background(), "dep-1", req)
	require.NoError(t, err)
	assert.Equal(t, 2000.0, record.DepreciationAmount)
}

func TestAssetDepreciationService_Delete(t *testing.T) {
	repo := &mockCovDepreciationRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.AssetDepreciationRecord, error) {
			return &entity.AssetDepreciationRecord{ID: "dep-1"}, nil
		},
	}
	svc := NewAssetDepreciationService(repo, &mockCovAssetRepo{})

	err := svc.DeleteDepreciationRecord(context.Background(), "dep-1")
	require.NoError(t, err)
}

func TestAssetDepreciationService_List(t *testing.T) {
	repo := &mockCovDepreciationRepo{
		listFn: func(_ context.Context, _ string, _ *string, _, _ int) ([]*entity.AssetDepreciationRecord, int64, error) {
			return []*entity.AssetDepreciationRecord{{ID: "d1"}, {ID: "d2"}}, 2, nil
		},
	}
	svc := NewAssetDepreciationService(repo, &mockCovAssetRepo{})

	records, total, err := svc.ListDepreciationRecords(context.Background(), "asset-1", "", 2024, 1, 10)
	require.NoError(t, err)
	assert.Len(t, records, 2)
	assert.Equal(t, int64(2), total)
}

func TestAssetDepreciationService_GetSummary(t *testing.T) {
	repo := &mockCovDepreciationRepo{
		summaryFn: func(_ context.Context, _ string, _, _ *time.Time) (float64, error) { return 15000.5, nil },
	}
	svc := NewAssetDepreciationService(repo, &mockCovAssetRepo{})

	summary, err := svc.GetDepreciationSummary(context.Background(), "asset-1", "")
	require.NoError(t, err)
	assert.Equal(t, 15000.5, summary)
}

func TestAssetDocumentService_Create(t *testing.T) {
	repo := &mockCovDocumentRepo{}
	assetRepo := &mockCovAssetRepo{}
	svc := NewAssetDocumentService(repo, assetRepo)

	req := &CreateDocumentRequest{
		AssetID:      "asset-1",
		DocumentType: "manual",
		Title:        "User Manual",
		FilePath:     "/docs/manual.pdf",
		Description:  "Equipment manual",
	}

	doc, err := svc.CreateDocument(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "asset-1", doc.AssetID)
	assert.Equal(t, "manual", doc.Type)
	assert.Equal(t, "User Manual", doc.Title)
}

func TestDocumentService_CreateWithUploadDate(t *testing.T) {
	repo := &mockCovDocumentRepo{}
	assetRepo := &mockCovAssetRepo{}
	svc := NewAssetDocumentService(repo, assetRepo)

	req := &CreateDocumentRequest{
		AssetID:      "asset-1",
		DocumentType: "certificate",
		Title:        "Cert Doc",
		FilePath:     "/cert.pdf",
		UploadDate:   "2024-01-15",
	}

	doc, err := svc.CreateDocument(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 15, doc.UploadDate.Day())
}

func TestAssetDocumentService_Update(t *testing.T) {
	existing := &entity.AssetDocument{ID: "doc-1", AssetID: "asset-1"}
	repo := &mockCovDocumentRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.AssetDocument, error) { return existing, nil },
	}
	assetRepo := &mockCovAssetRepo{}
	svc := NewAssetDocumentService(repo, assetRepo)

	req := &UpdateDocumentRequest{Title: "Updated Title", Description: "New desc"}
	doc, err := svc.UpdateDocument(context.Background(), "doc-1", req)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", doc.Title)
	assert.Equal(t, "New desc", doc.Description)
}

func TestAssetDocumentService_Delete(t *testing.T) {
	repo := &mockCovDocumentRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.AssetDocument, error) {
			return &entity.AssetDocument{ID: "doc-1"}, nil
		},
	}
	svc := NewAssetDocumentService(repo, &mockCovAssetRepo{})

	err := svc.DeleteDocument(context.Background(), "doc-1")
	require.NoError(t, err)
}

func TestAssetDocumentService_List(t *testing.T) {
	repo := &mockCovDocumentRepo{
		listFn: func(_ context.Context, _ string, _ *string, _, _ int) ([]*entity.AssetDocument, int64, error) {
			return []*entity.AssetDocument{{ID: "d1"}}, 1, nil
		},
	}
	svc := NewAssetDocumentService(repo, &mockCovAssetRepo{})

	docs, total, err := svc.ListDocuments(context.Background(), "asset-1", "", 1, 10)
	require.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.Equal(t, int64(1), total)
}

func TestAssetMaintenanceService_Create(t *testing.T) {
	repo := &mockCovMaintenanceRepo{}
	assetRepo := &mockCovAssetRepo{}
	svc := NewAssetMaintenanceService(repo, assetRepo)

	req := &CreateMaintenanceRequest{
		AssetID:          "asset-1",
		MaintenanceType:  "preventive",
		Description:      "Routine check",
		Cost:             500.0,
		MaintenanceDate:  "2024-06-15",
		Technician:       "John Doe",
		Status:           "pending",
	}

	record, err := svc.CreateMaintenanceRecord(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "preventive", record.MaintenanceType)
	assert.Equal(t, 500.0, record.Cost)
	assert.Equal(t, "pending", record.Status)
}

func TestAssetMaintenanceService_CreateDefaultStatus(t *testing.T) {
	repo := &mockCovMaintenanceRepo{}
	assetRepo := &mockCovAssetRepo{}
	svc := NewAssetMaintenanceService(repo, assetRepo)

	req := &CreateMaintenanceRequest{
		AssetID:         "asset-1",
		MaintenanceType: "corrective",
		MaintenanceDate: "2024-06-20",
	}

	record, err := svc.CreateMaintenanceRecord(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "pending", record.Status)
}

func TestAssetMaintenanceService_UpdateSetCompleted(t *testing.T) {
	existing := &entity.AssetMaintenanceRecord{ID: "maint-1", AssetID: "asset-1", Status: "in_progress"}
	repo := &mockCovMaintenanceRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.AssetMaintenanceRecord, error) { return existing, nil },
	}
	assetRepo := &mockCovAssetRepo{}
	svc := NewAssetMaintenanceService(repo, assetRepo)

	req := &UpdateMaintenanceRequest{Status: "completed"}
	record, err := svc.UpdateMaintenanceRecord(context.Background(), "maint-1", req)
	require.NoError(t, err)
	assert.Equal(t, "completed", record.Status)
	assert.NotNil(t, record.EndDate)
}

func TestAssetMaintenanceService_Delete(t *testing.T) {
	repo := &mockCovMaintenanceRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.AssetMaintenanceRecord, error) {
			return &entity.AssetMaintenanceRecord{ID: "m-1"}, nil
		},
	}
	svc := NewAssetMaintenanceService(repo, &mockCovAssetRepo{})

	err := svc.DeleteMaintenanceRecord(context.Background(), "m-1")
	require.NoError(t, err)
}

func TestAssetMaintenanceService_List(t *testing.T) {
	repo := &mockCovMaintenanceRepo{
		listFn: func(_ context.Context, _ string, _, _ *string, _, _ int) ([]*entity.AssetMaintenanceRecord, int64, error) {
			return []*entity.AssetMaintenanceRecord{{ID: "m1"}, {ID: "m2"}}, 2, nil
		},
	}
	svc := NewAssetMaintenanceService(repo, &mockCovAssetRepo{})

	records, total, err := svc.ListMaintenanceRecords(context.Background(), "asset-1", "", "", 1, 10)
	require.NoError(t, err)
	assert.Len(t, records, 2)
	assert.Equal(t, int64(2), total)
}

func TestAssetMaintenanceService_GetCosts(t *testing.T) {
	repo := &mockCovMaintenanceRepo{
		costFn: func(_ context.Context, _ string, _, _ *time.Time) (float64, error) { return 2500.75, nil },
	}
	svc := NewAssetMaintenanceService(repo, &mockCovAssetRepo{})

	cost, err := svc.GetMaintenanceCosts(context.Background(), "asset-1", "2024-01-01", "2024-12-31")
	require.NoError(t, err)
	assert.Equal(t, 2500.75, cost)
}

func TestInventoryService_Create(t *testing.T) {
	repo := &mockCovInventoryRepo{}
	transRepo := &mockCovTransRepo{}
	supplierRepo := &mockCovSupplierRepo{}
	svc := NewInventoryService(repo, transRepo, supplierRepo)

	item := &entity.Inventory{Code: "INV001", Name: "Test Item", Unit: "pcs", Quantity: 100}
	err := svc.CreateInventory(context.Background(), item)
	require.NoError(t, err)
}

func TestInventoryService_GetLowStock(t *testing.T) {
	repo := &mockCovInventoryRepo{
		lowStockFn: func(_ context.Context) ([]*entity.Inventory, error) {
			return []*entity.Inventory{{ID: "low-1", Code: "LOW001", Name: "Low Stock Item"}}, nil
		},
	}
	svc := NewInventoryService(repo, &mockCovTransRepo{}, &mockCovSupplierRepo{})

	items, err := svc.GetLowStockItems(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestPurchaseOrderService_Create(t *testing.T) {
	repo := &mockCovPurchaseOrderRepo{}
	supplierRepo := &mockCovSupplierRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.Supplier, error) {
			return &entity.Supplier{ID: "sup-1"}, nil
		},
	}
	svc := NewPurchaseOrderService(repo, supplierRepo)

	order := &entity.PurchaseOrder{Code: "PO001", SupplierID: "sup-1", TotalAmount: 5000.0, CreatedBy: "admin"}
	err := svc.CreatePurchaseOrder(context.Background(), order)
	require.NoError(t, err)
}

func TestPurchaseOrderService_List(t *testing.T) {
	repo := &mockCovPurchaseOrderRepo{
		listFn: func(_ context.Context, _, _ *string, _, _ *time.Time, _, _ int) ([]*entity.PurchaseOrder, int64, error) {
			return []*entity.PurchaseOrder{{ID: "p1"}, {ID: "p2"}}, 2, nil
		},
	}
	svc := NewPurchaseOrderService(repo, &mockCovSupplierRepo{})

	orders, total, err := svc.ListPurchaseOrders(context.Background(), nil, nil, nil, nil, 1, 10)
	require.NoError(t, err)
	assert.Len(t, orders, 2)
	assert.Equal(t, int64(2), total)
}

func TestPurchaseOrderService_UpdateStatus(t *testing.T) {
	existing := &entity.PurchaseOrder{ID: "po-1", Code: "PO001", Status: "pending"}
	repo := &mockCovPurchaseOrderRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.PurchaseOrder, error) { return existing, nil },
	}
	svc := NewPurchaseOrderService(repo, &mockCovSupplierRepo{})

	err := svc.UpdatePurchaseOrderStatus(context.Background(), "po-1", "approved")
	require.NoError(t, err)
	assert.Equal(t, "approved", existing.Status)
}

func TestReceiptService_Create(t *testing.T) {
	repo := &mockCovReceiptRepo{}
	poRepo := &mockCovPurchaseOrderRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.PurchaseOrder, error) {
			return &entity.PurchaseOrder{ID: "po-1"}, nil
		},
	}
	invRepo := &mockCovInventoryRepo{}
	transRepo := &mockCovTransRepo{}
	svc := NewReceiptService(repo, poRepo, invRepo, transRepo)

	receipt := &entity.Receipt{Code: "RCPT001", PurchaseOrderID: "po-1", ReceivedBy: "receiver"}
	err := svc.CreateReceipt(context.Background(), receipt)
	require.NoError(t, err)
}

func TestReceiptService_List(t *testing.T) {
	repo := &mockCovReceiptRepo{
		listFn: func(_ context.Context, _, _ *string, _, _ *time.Time, _, _ int) ([]*entity.Receipt, int64, error) {
			return []*entity.Receipt{{ID: "r1"}}, 1, nil
		},
	}
	svc := NewReceiptService(repo, &mockCovPurchaseOrderRepo{}, &mockCovInventoryRepo{}, &mockCovTransRepo{})

	receipts, total, err := svc.ListReceipts(context.Background(), nil, nil, nil, nil, 1, 10)
	require.NoError(t, err)
	assert.Len(t, receipts, 1)
	assert.Equal(t, int64(1), total)
}

func TestReceiptService_UpdateStatus(t *testing.T) {
	existing := &entity.Receipt{ID: "rcpt-1", Code: "RCPT001", Status: "draft"}
	repo := &mockCovReceiptRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.Receipt, error) { return existing, nil },
	}
	svc := NewReceiptService(repo, &mockCovPurchaseOrderRepo{}, &mockCovInventoryRepo{}, &mockCovTransRepo{})

	err := svc.UpdateReceiptStatus(context.Background(), "rcpt-1", "completed")
	require.NoError(t, err)
	assert.Equal(t, "completed", existing.Status)
}

func TestCostCategoryService_Create(t *testing.T) {
	repo := &mockCovCostCategoryRepo{
		getByCodeFn: func(_ context.Context, _ string) (*entity.CostCategory, error) { return nil, errors.New("not found") },
	}
	svc := NewCostCategoryService(repo)

	cat := &entity.CostCategory{Code: "CAT001", Name: "Direct Costs", Type: "direct"}
	err := svc.CreateCostCategory(context.Background(), cat)
	require.NoError(t, err)
}

func TestCostCategoryService_CreateDuplicateCode(t *testing.T) {
	repo := &mockCovCostCategoryRepo{
		getByCodeFn: func(_ context.Context, _ string) (*entity.CostCategory, error) {
			return &entity.CostCategory{ID: "existing"}, nil
		},
	}
	svc := NewCostCategoryService(repo)

	cat := &entity.CostCategory{Code: "CAT001", Name: "Dup"}
	err := svc.CreateCostCategory(context.Background(), cat)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "编码已存在")
}

func TestCostCategoryService_DeleteHasChildren(t *testing.T) {
	repo := &mockCovCostCategoryRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.CostCategory, error) {
			return &entity.CostCategory{ID: "parent-1"}, nil
		},
		listFn: func(_ context.Context, _, _ *string) ([]*entity.CostCategory, error) {
			return []*entity.CostCategory{{ID: "child-1"}}, nil
		},
	}
	svc := NewCostCategoryService(repo)

	err := svc.DeleteCostCategory(context.Background(), "parent-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "子类别")
}

func TestCostEntryService_Create(t *testing.T) {
	catRepo := &mockCovCostCategoryRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.CostCategory, error) {
			return &entity.CostCategory{ID: "cat-1"}, nil
		},
	}
	entryRepo := &mockCovCostEntryRepo{}
	svc := NewCostEntryService(entryRepo, catRepo)

	entry := &entity.CostEntry{CostCategoryID: "cat-1", Amount: 1000.0, Currency: "CNY", Description: "Test entry"}
	err := svc.CreateCostEntry(context.Background(), entry)
	require.NoError(t, err)
	assert.NotEmpty(t, entry.Code)
	assert.Equal(t, "pending", entry.ApprovalStatus)
}

func TestCostEntryService_Approve(t *testing.T) {
	existing := &entity.CostEntry{ID: "entry-1", ApprovalStatus: "pending"}
	repo := &mockCovCostEntryRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.CostEntry, error) { return existing, nil },
	}
	svc := NewCostEntryService(repo, &mockCovCostCategoryRepo{})

	err := svc.ApproveCostEntry(context.Background(), "entry-1", "approver")
	require.NoError(t, err)
	assert.Equal(t, "approved", existing.ApprovalStatus)
}

func TestCostEntryService_Reject(t *testing.T) {
	existing := &entity.CostEntry{ID: "entry-2", ApprovalStatus: "pending"}
	repo := &mockCovCostEntryRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.CostEntry, error) { return existing, nil },
	}
	svc := NewCostEntryService(repo, &mockCovCostCategoryRepo{})

	err := svc.RejectCostEntry(context.Background(), "entry-2", "approver")
	require.NoError(t, err)
	assert.Equal(t, "rejected", existing.ApprovalStatus)
}

func TestCostEntryService_ApproveNonPending(t *testing.T) {
	existing := &entity.CostEntry{ID: "entry-3", ApprovalStatus: "approved"}
	repo := &mockCovCostEntryRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.CostEntry, error) { return existing, nil },
	}
	svc := NewCostEntryService(repo, &mockCovCostCategoryRepo{})

	err := svc.ApproveCostEntry(context.Background(), "entry-3", "approver")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "不是待审批")
}

func TestCostAllocationService_Create(t *testing.T) {
	allocationRepo := &mockCovCostAllocationRepo{}
	entryRepo := &mockCovCostEntryRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.CostEntry, error) {
			return &entity.CostEntry{ID: "e-1", Amount: 1000.0, ApprovalStatus: "approved"}, nil
		},
	}
	svc := NewCostAllocationService(allocationRepo, entryRepo)

	allocation := &entity.CostAllocation{
		CostEntryID: "e-1", AllocatedTo: "project", AllocatedID: "proj-1",
		Amount: 300.0, Percentage: 30.0,
	}
	err := svc.CreateCostAllocation(context.Background(), allocation)
	require.NoError(t, err)
}

func TestCostAllocationService_CreateUnapprovedEntry(t *testing.T) {
	allocationRepo := &mockCovCostAllocationRepo{}
	entryRepo := &mockCovCostEntryRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.CostEntry, error) {
			return &entity.CostEntry{ID: "e-2", ApprovalStatus: "pending"}, nil
		},
	}
	svc := NewCostAllocationService(allocationRepo, entryRepo)

	allocation := &entity.CostAllocation{CostEntryID: "e-2", Amount: 100.0, Percentage: 10.0}
	err := svc.CreateCostAllocation(context.Background(), allocation)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "未审批")
}

func TestCostAllocationService_CreateZeroAmount(t *testing.T) {
	entryRepo := &mockCovCostEntryRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.CostEntry, error) {
			return &entity.CostEntry{ID: "e-x", Amount: 1000.0, ApprovalStatus: "approved"}, nil
		},
	}
	svc := NewCostAllocationService(&mockCovCostAllocationRepo{}, entryRepo)
	allocation := &entity.CostAllocation{CostEntryID: "e-x", Amount: 0, Percentage: 10.0}
	err := svc.CreateCostAllocation(context.Background(), allocation)
	require.Error(t, err)
}

func TestCostAllocationService_CreateOverBudget(t *testing.T) {
	allocationRepo := &mockCovCostAllocationRepo{
		listByEntryFn: func(_ context.Context, _ string) ([]*entity.CostAllocation, error) {
			return []*entity.CostAllocation{{Amount: 800.0}}, nil
		},
	}
	entryRepo := &mockCovCostEntryRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.CostEntry, error) {
			return &entity.CostEntry{ID: "e-3", Amount: 1000.0, ApprovalStatus: "approved"}, nil
		},
	}
	svc := NewCostAllocationService(allocationRepo, entryRepo)

	allocation := &entity.CostAllocation{CostEntryID: "e-3", Amount: 300.0, Percentage: 30.0}
	err := svc.CreateCostAllocation(context.Background(), allocation)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "累计分配金额不能超过")
}

func TestCostReportService_Create(t *testing.T) {
	reportRepo := &mockCovCostReportRepo{
		getByPeriodFn: func(_ context.Context, _ string, _, _ time.Time) (*entity.CostReport, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewCostReportService(reportRepo, &mockCovCostEntryRepo{})

	report := &entity.CostReport{
		Name: "Monthly Report", ReportType: "monthly",
		PeriodStart: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2024, 6, 30, 23, 59, 59, 0, time.UTC),
	}
	err := svc.CreateCostReport(context.Background(), report)
	require.NoError(t, err)
	assert.Equal(t, "draft", report.Status)
}

func TestCostReportService_Generate(t *testing.T) {
	existing := &entity.CostReport{ID: "r-1", Status: "draft"}
	reportRepo := &mockCovCostReportRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.CostReport, error) { return existing, nil },
	}
	entryRepo := &mockCovCostEntryRepo{
		totalPeriodFn: func(_ context.Context, _, _ *time.Time) (float64, error) { return 9999.99, nil },
	}
	svc := NewCostReportService(reportRepo, entryRepo)

	err := svc.GenerateCostReport(context.Background(), "r-1", "generator")
	require.NoError(t, err)
	assert.Equal(t, "generated", existing.Status)
	assert.Equal(t, 9999.99, existing.TotalCost)
}

func TestCostReportService_Approve(t *testing.T) {
	existing := &entity.CostReport{ID: "r-2", Status: "generated"}
	reportRepo := &mockCovCostReportRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.CostReport, error) { return existing, nil },
	}
	svc := NewCostReportService(reportRepo, &mockCovCostEntryRepo{})

	err := svc.ApproveCostReport(context.Background(), "r-2", "approver")
	require.NoError(t, err)
	assert.Equal(t, "approved", existing.Status)
}

func TestCostReportService_Reject(t *testing.T) {
	existing := &entity.CostReport{ID: "r-3", Status: "generated"}
	reportRepo := &mockCovCostReportRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.CostReport, error) { return existing, nil },
	}
	svc := NewCostReportService(reportRepo, &mockCovCostEntryRepo{})

	err := svc.RejectCostReport(context.Background(), "r-3", "approver")
	require.NoError(t, err)
	assert.Equal(t, "rejected", existing.Status)
}

func TestCarbonEmissionService_CreateFactor(t *testing.T) {
	repo := &mockCovCarbonEmissionRepo{
		factorByCodeFn: func(_ context.Context, _ string) (*entity.CarbonEmissionFactor, error) { return nil, errors.New("not found") },
	}
	svc := NewCarbonEmissionService(repo)

	req := &CreateCarbonEmissionFactorRequest{
		Name:        "Electricity Factor",
		Code:        "EF-ELEC-001",
		Scope:       entity.CarbonEmissionScope2,
		Source:      "IPCC",
		Value:       0.456,
		Unit:        "kgCO2/kWh",
		Version:     "v1.0",
		EffectiveAt: time.Now(),
	}

	factor, err := svc.CreateFactor(context.Background(), req, "admin")
	require.NoError(t, err)
	assert.Equal(t, "EF-ELEC-001", factor.Code)
	assert.True(t, factor.IsActive)
}

func TestCarbonEmissionService_CreateFactorDuplicate(t *testing.T) {
	repo := &mockCovCarbonEmissionRepo{
		factorByCodeFn: func(_ context.Context, _ string) (*entity.CarbonEmissionFactor, error) {
			return &entity.CarbonEmissionFactor{ID: "existing"}, nil
		},
	}
	svc := NewCarbonEmissionService(repo)

	req := &CreateCarbonEmissionFactorRequest{
		Name: "Dup Factor", Code: "DUP-001", Scope: entity.CarbonEmissionScope1,
		Source: "test", Value: 1.0, Unit: "t", Version: "v1", EffectiveAt: time.Now(),
	}

	_, err := svc.CreateFactor(context.Background(), req, "admin")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestCarbonEmissionService_UpdateFactor(t *testing.T) {
	existing := &entity.CarbonEmissionFactor{ID: "f-1", Code: "F001", Name: "Old Name"}
	repo := &mockCovCarbonEmissionRepo{
		factorByIDFn: func(_ context.Context, _ string) (*entity.CarbonEmissionFactor, error) { return existing, nil },
	}
	svc := NewCarbonEmissionService(repo)

	name := "Updated Name"
	req := &UpdateCarbonEmissionFactorRequest{Name: &name}
	factor, err := svc.UpdateFactor(context.Background(), "f-1", req, "admin")
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", factor.Name)
}

func TestCarbonEmissionService_CreateRecord(t *testing.T) {
	repo := &mockCovCarbonEmissionRepo{}
	svc := NewCarbonEmissionService(repo)

	req := &CreateCarbonEmissionRecordRequest{
		RecordTime:   time.Now(),
		Scope:        entity.CarbonEmissionScope2,
		TargetID:     "station-1",
		TargetName:   "Solar Station A",
		FactorCode:   "EF-ELEC-001",
		FactorValue:  0.456,
		ActivityData: 1000.0,
		ActivityUnit: "kWh",
		Period:       "2024-06",
	}

	record, err := svc.CreateRecord(context.Background(), req, "operator")
	require.NoError(t, err)
	assert.Equal(t, "station-1", record.TargetID)
	assert.Equal(t, float64(456.0), record.EmissionValue)
}

func TestCarbonEmissionService_BatchCreateRecords(t *testing.T) {
	repo := &mockCovCarbonEmissionRepo{}
	svc := NewCarbonEmissionService(repo)

	reqs := []*CreateCarbonEmissionRecordRequest{
		{
			RecordTime: time.Now(), Scope: entity.CarbonEmissionScope2, TargetID: "s1",
			TargetName: "S1", FactorCode: "EF001", FactorValue: 0.5, ActivityData: 100.0,
			ActivityUnit: "kWh", Period: "2024-06",
		},
		{
			RecordTime: time.Now(), Scope: entity.CarbonEmissionScope1, TargetID: "s2",
			TargetName: "S2", FactorCode: "EF002", FactorValue: 2.0, ActivityData: 50.0,
			ActivityUnit: "ton", Period: "2024-06",
		},
	}

	records, err := svc.BatchCreateRecords(context.Background(), reqs, "batch-op")
	require.NoError(t, err)
	assert.Len(t, records, 2)
}

func TestCarbonEmissionService_CreateSummary(t *testing.T) {
	repo := &mockCovCarbonEmissionRepo{
		totalScopeFn: func(_ context.Context, _ string, _ entity.CarbonEmissionScope, _ string, _, _ time.Time) (float64, error) {
			return 100.0, nil
		},
		latestSummaryFn: func(_ context.Context, _ string, _ string) (*entity.CarbonEmissionSummary, error) {
			return &entity.CarbonEmissionSummary{TotalEmission: 90.0}, nil
		},
	}
	svc := NewCarbonEmissionService(repo)

	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
	summary, err := svc.CreateSummary(context.Background(), "station-1", "Station A", "2024-06", start, end, "admin")
	require.NoError(t, err)
	assert.Equal(t, "station-1", summary.TargetID)
	assert.GreaterOrEqual(t, summary.TotalEmission, 300.0)
}

func TestCarbonEmissionService_CreateTarget(t *testing.T) {
	repo := &mockCovCarbonEmissionRepo{}
	svc := NewCarbonEmissionService(repo)

	req := &CreateCarbonReductionTargetRequest{
		Name:            "2030 Target",
		Description:     "Reduce emissions by 30%",
		TargetID:        "station-1",
		TargetName:      "Station A",
		BaseYear:        2024,
		BaseEmission:    1000.0,
		TargetYear:      2030,
		TargetReduction: 30.0,
		StartDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:         time.Date(2030, 12, 31, 0, 0, 0, 0, time.UTC),
	}

	target, err := svc.CreateTarget(context.Background(), req, "admin")
	require.NoError(t, err)
	assert.Equal(t, "2030 Target", target.Name)
	assert.Equal(t, "active", target.Status)
	assert.Less(t, target.TargetEmission, target.BaseEmission)
}

func TestCarbonEmissionService_UpdateTarget(t *testing.T) {
	existing := &entity.CarbonReductionTarget{ID: "t-1", Name: "Old Target", CurrentProgress: 25.0}
	repo := &mockCovCarbonEmissionRepo{
		targetByIDFn: func(_ context.Context, _ string) (*entity.CarbonReductionTarget, error) { return existing, nil },
	}
	svc := NewCarbonEmissionService(repo)

	progress := covFloatPtr(50.0)
	req := &UpdateCarbonReductionTargetRequest{CurrentProgress: progress}
	target, err := svc.UpdateTarget(context.Background(), "t-1", req, "admin")
	require.NoError(t, err)
	assert.Equal(t, 50.0, target.CurrentProgress)
}

func TestCarbonEmissionService_TrendData(t *testing.T) {
	repo := &mockCovCarbonEmissionRepo{
		recordsByTimeFn: func(_ context.Context, _ string, _ *entity.CarbonEmissionScope, start, end time.Time) ([]*entity.CarbonEmissionRecord, error) {
			return []*entity.CarbonEmissionRecord{
				{RecordTime: start, Scope: entity.CarbonEmissionScope1, EmissionValue: 10.0},
				{RecordTime: start, Scope: entity.CarbonEmissionScope2, EmissionValue: 20.0},
				{RecordTime: start.Add(24 * time.Hour), Scope: entity.CarbonEmissionScope1, EmissionValue: 11.0},
			}, nil
		},
	}
	svc := NewCarbonEmissionService(repo)

	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)
	data, err := svc.GetTrendData(context.Background(), "station-1", "daily", start, end)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(data), 1)
}

func TestDataLoopService_CollectFeedback(t *testing.T) {
	repo := &mockCovForecastRepo{
		accuracyStatsFn: func(_ context.Context, _ string, _ *entity.ForecastType, _, _ time.Time) (*repository.ForecastAccuracyStats, error) {
			return &repository.ForecastAccuracyStats{TotalPoints: 100, AvgAccuracy: 0.92, RMSE: 8.5}, nil
		},
	}
	svc := NewDataLoopService(repo, &mockCovModelService{})

	report, err := svc.CollectFeedback(context.Background(), "station-1", entity.ForecastTypeShortTerm, time.Now())
	require.NoError(t, err)
	assert.Equal(t, "station-1", report.StationID)
	assert.Equal(t, 100, report.TotalPoints)
}

func TestDataLoopService_EvaluateAndTrigger(t *testing.T) {
	repo := &mockCovForecastRepo{
		accuracyStatsFn: func(_ context.Context, _ string, _ *entity.ForecastType, _, _ time.Time) (*repository.ForecastAccuracyStats, error) {
			return &repository.ForecastAccuracyStats{AvgAccuracy: 0.70}, nil
		},
	}
	svc := NewDataLoopService(repo, &mockCovModelService{})

	decision, err := svc.EvaluateAndTrigger(context.Background(), "station-1")
	require.NoError(t, err)
	assert.True(t, decision.ShouldRetrain)
	assert.Contains(t, decision.Reason, "低于阈值")
}

func TestDataLoopService_EvaluateAndTrigger_NoStats(t *testing.T) {
	repo := &mockCovForecastRepo{
		accuracyStatsFn: func(_ context.Context, _ string, _ *entity.ForecastType, _, _ time.Time) (*repository.ForecastAccuracyStats, error) {
			return nil, errors.New("no data")
		},
	}
	svc := NewDataLoopService(repo, &mockCovModelService{})

	decision, err := svc.EvaluateAndTrigger(context.Background(), "station-1")
	require.NoError(t, err)
	assert.False(t, decision.ShouldRetrain)
	assert.Contains(t, decision.Reason, "无法获取准确率统计")
}

func TestDataLoopService_GetLoopStatus(t *testing.T) {
	repo := &mockCovForecastRepo{
		accuracyStatsFn: func(_ context.Context, _ string, _ *entity.ForecastType, _, _ time.Time) (*repository.ForecastAccuracyStats, error) {
			return &repository.ForecastAccuracyStats{AvgAccuracy: 0.88}, nil
		},
	}
	svc := NewDataLoopService(repo, &mockCovModelService{})

	status, err := svc.GetLoopStatus(context.Background(), "station-1")
	require.NoError(t, err)
	assert.Equal(t, "station-1", status.StationID)
	assert.Equal(t, 0.88, status.CurrentAccuracy)
}

func TestFaultWorkOrderBridge_CreateFromDetection(t *testing.T) {
	detection := &entity.FaultDetectionResult{ID: "det-1", Severity: entity.FaultSeverityCritical}
	faultRepo := &mockCovFaultDetectionRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.FaultDetectionResult, error) { return detection, nil },
		linkWorkOrderFn: func(_ context.Context, _, _ string) error { return nil },
	}
	creator := &mockCovWorkOrderCreator{fn: func(_ context.Context, _ *entity.FaultDetectionResult) (string, error) { return "wo-1", nil }}
	bridge := NewFaultWorkOrderBridge(faultRepo, creator)

	woID, err := bridge.CreateWorkOrderFromDetection(context.Background(), "det-1")
	require.NoError(t, err)
	assert.Equal(t, "wo-1", woID)
}

func TestFaultWorkOrderBridge_InfoSeveritySkipped(t *testing.T) {
	detection := &entity.FaultDetectionResult{ID: "det-info", Severity: entity.FaultSeverityInfo}
	faultRepo := &mockCovFaultDetectionRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.FaultDetectionResult, error) { return detection, nil },
	}
	bridge := NewFaultWorkOrderBridge(faultRepo, &mockCovWorkOrderCreator{})

	woID, err := bridge.CreateWorkOrderFromDetection(context.Background(), "det-info")
	require.NoError(t, err)
	assert.Empty(t, woID)
}

func TestFaultWorkOrderBridge_DetectionNotFound(t *testing.T) {
	faultRepo := &mockCovFaultDetectionRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.FaultDetectionResult, error) { return nil, errors.New("not found") },
	}
	bridge := NewFaultWorkOrderBridge(faultRepo, &mockCovWorkOrderCreator{})

	_, err := bridge.CreateWorkOrderFromDetection(context.Background(), "bad-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFaultWorkOrderBridge_ShouldAutoCreate(t *testing.T) {
	bridge := NewFaultWorkOrderBridge(nil, nil)

	assert.True(t, bridge.ShouldAutoCreateWorkOrder(entity.FaultSeverityFatal))
	assert.True(t, bridge.ShouldAutoCreateWorkOrder(entity.FaultSeverityCritical))
	assert.False(t, bridge.ShouldAutoCreateWorkOrder(entity.FaultSeverityWarning))
	assert.False(t, bridge.ShouldAutoCreateWorkOrder(entity.FaultSeverityInfo))
}

func TestFaultWorkOrderBridge_GetSLA(t *testing.T) {
	bridge := NewFaultWorkOrderBridge(nil, nil)

	assert.Equal(t, "30min", bridge.GetWorkOrderSLA(entity.FaultSeverityFatal))
	assert.Equal(t, "1h", bridge.GetWorkOrderSLA(entity.FaultSeverityCritical))
	assert.Equal(t, "4h", bridge.GetWorkOrderSLA(entity.FaultSeverityWarning))
	assert.Equal(t, "24h", bridge.GetWorkOrderSLA(entity.FaultSeverityInfo))
}

func TestRequestStructs(t *testing.T) {
	t.Run("CreateDepreciationRequest", func(t *testing.T) {
		r := &CreateDepreciationRequest{AssetID: "a1", DepreciationMethod: "sl", Year: 2024, Amount: 100, AccumulatedAmount: 200, BookValue: 900}
		assert.Equal(t, "a1", r.AssetID)
	})
	t.Run("UpdateDepreciationRequest", func(t *testing.T) {
		r := &UpdateDepreciationRequest{AssetID: "a2", DepreciationMethod: "db", Year: 2025, Amount: 200, AccumulatedAmount: 400, BookValue: 800}
		assert.Equal(t, "db", r.DepreciationMethod)
	})
	t.Run("CreateDocumentRequest", func(t *testing.T) {
		r := &CreateDocumentRequest{AssetID: "a1", DocumentType: "pdf", Title: "Doc", FilePath: "/path", UploadDate: "2024-01-01"}
		assert.Equal(t, "pdf", r.DocumentType)
	})
	t.Run("UpdateDocumentRequest", func(t *testing.T) {
		r := &UpdateDocumentRequest{AssetID: "a1", DocumentType: "txt", Title: "Upd", FilePath: "/new/path", UploadDate: "2024-06-01"}
		assert.Equal(t, "Upd", r.Title)
	})
	t.Run("CreateMaintenanceRequest", func(t *testing.T) {
		r := &CreateMaintenanceRequest{AssetID: "a1", MaintenanceType: "pm", Cost: 500, MaintenanceDate: "2024-03-15", Technician: "T1", Status: "done"}
		assert.Equal(t, "pm", r.MaintenanceType)
	})
	t.Run("UpdateMaintenanceRequest", func(t *testing.T) {
		r := &UpdateMaintenanceRequest{AssetID: "a1", MaintenanceType: "cm", Cost: 800, MaintenanceDate: "2024-04-15", Technician: "T2", Status: "open"}
		assert.Equal(t, "cm", r.MaintenanceType)
	})
	t.Run("CreateCarbonEmissionFactorRequest", func(t *testing.T) {
		r := &CreateCarbonEmissionFactorRequest{Name: "N", Code: "C", Scope: entity.CarbonEmissionScope1, Source: "S", Value: 1.0, Unit: "U", Version: "v1", EffectiveAt: time.Now()}
		assert.Equal(t, "N", r.Name)
	})
	t.Run("UpdateCarbonEmissionFactorRequest", func(t *testing.T) {
		n := "N2"
		s := entity.CarbonEmissionScope2
		v := covFloatPtr(2.0)
		b := covBoolPtr(true)
		r := &UpdateCarbonEmissionFactorRequest{Name: &n, Scope: &s, Value: v, IsActive: b}
		assert.Equal(t, "N2", *r.Name)
	})
	t.Run("CreateCarbonEmissionRecordRequest", func(t *testing.T) {
		r := &CreateCarbonEmissionRecordRequest{RecordTime: time.Now(), Scope: entity.CarbonEmissionScope3, TargetID: "t1", TargetName: "T1", FactorCode: "FC1", FactorValue: 1.5, ActivityData: 100.0, ActivityUnit: "kWh", Period: "2024-06"}
		assert.Equal(t, "T1", r.TargetName)
	})
	t.Run("QueryCarbonEmissionFactorsRequest", func(t *testing.T) {
		scope := entity.CarbonEmissionScope1
		active := true
		r := &QueryCarbonEmissionFactorsRequest{Page: 2, PageSize: 20, Scope: &scope, Source: covStrPtr("IPCC"), IsActive: &active}
		assert.Equal(t, 2, r.Page)
	})
	t.Run("QueryCarbonEmissionRecordsRequest", func(t *testing.T) {
		scope := entity.CarbonEmissionScope2
		status := entity.CarbonEmissionStatusApproved
		r := &QueryCarbonEmissionRecordsRequest{Page: 1, PageSize: 50, Scope: &scope, TargetID: covStrPtr("s1"), Period: covStrPtr("2024"), Status: &status}
		assert.Equal(t, 50, r.PageSize)
	})
	t.Run("CreateCarbonReductionTargetRequest", func(t *testing.T) {
		r := &CreateCarbonReductionTargetRequest{Name: "Target", TargetID: "t1", TargetName: "T1", BaseYear: 2024, BaseEmission: 1000, TargetYear: 2030, TargetReduction: 30, StartDate: time.Now(), EndDate: time.Now().AddDate(6, 0, 0)}
		assert.Equal(t, 30.0, r.TargetReduction)
	})
	t.Run("UpdateCarbonReductionTargetRequest", func(t *testing.T) {
		e := covFloatPtr(700.0)
		p := covFloatPtr(35.0)
		st := covStrPtr("active")
		r := &UpdateCarbonReductionTargetRequest{TargetEmission: e, CurrentProgress: p, Status: st}
		assert.Equal(t, 700.0, *r.TargetEmission)
	})
	t.Run("CarbonEmissionTrendData", func(t *testing.T) {
		now := time.Now()
		d := &CarbonEmissionTrendData{Time: now, Scope1Emission: 10, Scope2Emission: 20, Scope3Emission: 5, TotalEmission: 35}
		assert.Equal(t, 35.0, d.TotalEmission)
	})
}

type mockCovTransRepo struct{}

func (m *mockCovTransRepo) Create(ctx context.Context, t *entity.InventoryTransaction) error { return nil }
func (m *mockCovTransRepo) GetByID(ctx context.Context, id string) (*entity.InventoryTransaction, error) { return nil, errors.New("not found") }
func (m *mockCovTransRepo) ListByInventoryID(ctx context.Context, inventoryID string) ([]*entity.InventoryTransaction, error) { return nil, nil }
func (m *mockCovTransRepo) ListByReference(ctx context.Context, referenceID string, referenceType string) ([]*entity.InventoryTransaction, error) { return nil, nil }
func (m *mockCovTransRepo) GetTransactionHistory(ctx context.Context, inventoryID string, limit int) ([]*entity.InventoryTransaction, error) { return nil, nil }
