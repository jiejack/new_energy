package repository

import (
	"context"
	"time"
	
	"github.com/new-energy-monitoring/internal/domain/entity"
)

type RegionRepository interface {
	Create(ctx context.Context, region *entity.Region) error
	Update(ctx context.Context, region *entity.Region) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.Region, error)
	GetByCode(ctx context.Context, code string) (*entity.Region, error)
	List(ctx context.Context, parentID *string) ([]*entity.Region, error)
	GetTree(ctx context.Context) ([]*entity.Region, error)
}

type SubRegionRepository interface {
	Create(ctx context.Context, subRegion *entity.SubRegion) error
	Update(ctx context.Context, subRegion *entity.SubRegion) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.SubRegion, error)
	GetByRegionID(ctx context.Context, regionID string) ([]*entity.SubRegion, error)
}

type StationRepository interface {
	Create(ctx context.Context, station *entity.Station) error
	Update(ctx context.Context, station *entity.Station) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.Station, error)
	GetByCode(ctx context.Context, code string) (*entity.Station, error)
	List(ctx context.Context, subRegionID *string, stationType *entity.StationType) ([]*entity.Station, error)
	GetWithDevices(ctx context.Context, id string) (*entity.Station, error)
}

type DeviceRepository interface {
	Create(ctx context.Context, device *entity.Device) error
	Update(ctx context.Context, device *entity.Device) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.Device, error)
	GetByCode(ctx context.Context, code string) (*entity.Device, error)
	List(ctx context.Context, stationID *string, deviceType *entity.DeviceType) ([]*entity.Device, error)
	GetWithPoints(ctx context.Context, id string) (*entity.Device, error)
	GetOnlineDevices(ctx context.Context, stationID string) ([]*entity.Device, error)
}

type PointRepository interface {
	Create(ctx context.Context, point *entity.Point) error
	BatchCreate(ctx context.Context, points []*entity.Point) error
	Update(ctx context.Context, point *entity.Point) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.Point, error)
	GetByCode(ctx context.Context, code string) (*entity.Point, error)
	List(ctx context.Context, deviceID *string, pointType *entity.PointType) ([]*entity.Point, error)
	GetByStationID(ctx context.Context, stationID string) ([]*entity.Point, error)
	GetByProtocol(ctx context.Context, protocol string) ([]*entity.Point, error)
}

type AlarmRepository interface {
	Create(ctx context.Context, alarm *entity.Alarm) error
	Update(ctx context.Context, alarm *entity.Alarm) error
	GetByID(ctx context.Context, id string) (*entity.Alarm, error)
	GetActiveAlarms(ctx context.Context, stationID *string, level *entity.AlarmLevel) ([]*entity.Alarm, error)
	GetHistoryAlarms(ctx context.Context, stationID *string, startTime, endTime int64) ([]*entity.Alarm, error)
	Acknowledge(ctx context.Context, id, by string) error
	Clear(ctx context.Context, id string) error
	CountByLevel(ctx context.Context, stationID *string) (map[entity.AlarmLevel]int64, error)
}

type WorkOrderRepository interface {
	Create(ctx context.Context, workOrder *entity.WorkOrder) error
	Update(ctx context.Context, workOrder *entity.WorkOrder) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.WorkOrder, error)
	List(ctx context.Context, filter interface{}) ([]*entity.WorkOrder, error)
	Count(ctx context.Context, filter interface{}) (int64, error)
}

type InventoryRepository interface {
	Create(ctx context.Context, inventory *entity.Inventory) error
	Update(ctx context.Context, inventory *entity.Inventory) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.Inventory, error)
	GetByCode(ctx context.Context, code string) (*entity.Inventory, error)
	List(ctx context.Context, filter interface{}) ([]*entity.Inventory, error)
	Count(ctx context.Context, filter interface{}) (int64, error)
	UpdateQuantity(ctx context.Context, id string, quantity float64) error
	GetLowStockItems(ctx context.Context) ([]*entity.Inventory, error)
}

type InventoryTransactionRepository interface {
	Create(ctx context.Context, transaction *entity.InventoryTransaction) error
	GetByID(ctx context.Context, id string) (*entity.InventoryTransaction, error)
	ListByInventoryID(ctx context.Context, inventoryID string) ([]*entity.InventoryTransaction, error)
	ListByReference(ctx context.Context, referenceID string, referenceType string) ([]*entity.InventoryTransaction, error)
	GetTransactionHistory(ctx context.Context, inventoryID string, limit int) ([]*entity.InventoryTransaction, error)
}

type SupplierRepository interface {
	Create(ctx context.Context, supplier *entity.Supplier) error
	Update(ctx context.Context, supplier *entity.Supplier) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.Supplier, error)
	GetByCode(ctx context.Context, code string) (*entity.Supplier, error)
	List(ctx context.Context, filter interface{}) ([]*entity.Supplier, error)
	Count(ctx context.Context, filter interface{}) (int64, error)
}

type PurchaseOrderRepository interface {
	Create(ctx context.Context, order *entity.PurchaseOrder) error
	Update(ctx context.Context, order *entity.PurchaseOrder) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.PurchaseOrder, error)
	List(ctx context.Context, supplierID *string, status *string, startDate, endDate *time.Time, offset, limit int) ([]*entity.PurchaseOrder, int64, error)
	GetByCode(ctx context.Context, code string) (*entity.PurchaseOrder, error)
}

type ReceiptRepository interface {
	Create(ctx context.Context, receipt *entity.Receipt) error
	Update(ctx context.Context, receipt *entity.Receipt) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.Receipt, error)
	List(ctx context.Context, purchaseOrderID *string, status *string, startDate, endDate *time.Time, offset, limit int) ([]*entity.Receipt, int64, error)
	GetByCode(ctx context.Context, code string) (*entity.Receipt, error)
	GetByPurchaseOrderID(ctx context.Context, purchaseOrderID string) ([]*entity.Receipt, error)
}

type CostCategoryRepository interface {
	Create(ctx context.Context, category *entity.CostCategory) error
	Update(ctx context.Context, category *entity.CostCategory) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.CostCategory, error)
	GetByCode(ctx context.Context, code string) (*entity.CostCategory, error)
	List(ctx context.Context, parentID *string, status *string) ([]*entity.CostCategory, error)
	GetTree(ctx context.Context) ([]*entity.CostCategory, error)
}

type CostEntryRepository interface {
	Create(ctx context.Context, entry *entity.CostEntry) error
	Update(ctx context.Context, entry *entity.CostEntry) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.CostEntry, error)
	GetByCode(ctx context.Context, code string) (*entity.CostEntry, error)
	List(ctx context.Context, categoryID *string, startDate, endDate *time.Time, status *string, offset, limit int) ([]*entity.CostEntry, int64, error)
	GetTotalByCategory(ctx context.Context, categoryID string, startDate, endDate *time.Time) (float64, error)
	GetTotalByPeriod(ctx context.Context, startDate, endDate *time.Time) (float64, error)
}

type CostAllocationRepository interface {
	Create(ctx context.Context, allocation *entity.CostAllocation) error
	Update(ctx context.Context, allocation *entity.CostAllocation) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.CostAllocation, error)
	ListByCostEntryID(ctx context.Context, costEntryID string) ([]*entity.CostAllocation, error)
	ListByAllocated(ctx context.Context, allocatedTo, allocatedID string) ([]*entity.CostAllocation, error)
	GetTotalByAllocated(ctx context.Context, allocatedTo, allocatedID string, startDate, endDate *time.Time) (float64, error)
}

type CostReportRepository interface {
	Create(ctx context.Context, report *entity.CostReport) error
	Update(ctx context.Context, report *entity.CostReport) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.CostReport, error)
	GetByCode(ctx context.Context, code string) (*entity.CostReport, error)
	List(ctx context.Context, reportType *string, status *string, startDate, endDate *time.Time, offset, limit int) ([]*entity.CostReport, int64, error)
	GetByPeriod(ctx context.Context, reportType string, periodStart, periodEnd time.Time) (*entity.CostReport, error)
}

type AssetRepository interface {
	Create(ctx context.Context, asset *entity.Asset) error
	Update(ctx context.Context, asset *entity.Asset) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.Asset, error)
	GetByCode(ctx context.Context, code string) (*entity.Asset, error)
	List(ctx context.Context, assetType *string, status *string, category *string, offset, limit int) ([]*entity.Asset, int64, error)
	GetByLocation(ctx context.Context, location string) ([]*entity.Asset, error)
	GetByDepartment(ctx context.Context, departmentID string) ([]*entity.Asset, error)
	GetByResponsiblePerson(ctx context.Context, person string) ([]*entity.Asset, error)
	GetDepreciatingAssets(ctx context.Context) ([]*entity.Asset, error)
	GetAssetsNearWarrantyEnd(ctx context.Context, days int) ([]*entity.Asset, error)
}

type AssetMaintenanceRepository interface {
	Create(ctx context.Context, record *entity.AssetMaintenanceRecord) error
	Update(ctx context.Context, record *entity.AssetMaintenanceRecord) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.AssetMaintenanceRecord, error)
	ListByAssetID(ctx context.Context, assetID string, status *string, maintenanceType *string, offset, limit int) ([]*entity.AssetMaintenanceRecord, int64, error)
	ListByStatus(ctx context.Context, status string, offset, limit int) ([]*entity.AssetMaintenanceRecord, int64, error)
	GetMaintenanceCostByAsset(ctx context.Context, assetID string, startDate, endDate *time.Time) (float64, error)
}

type AssetDepreciationRepository interface {
	Create(ctx context.Context, record *entity.AssetDepreciationRecord) error
	Update(ctx context.Context, record *entity.AssetDepreciationRecord) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.AssetDepreciationRecord, error)
	ListByAssetID(ctx context.Context, assetID string, period *string, offset, limit int) ([]*entity.AssetDepreciationRecord, int64, error)
	GetLatestByAssetID(ctx context.Context, assetID string) (*entity.AssetDepreciationRecord, error)
	GetDepreciationSummaryByPeriod(ctx context.Context, period string, startDate, endDate *time.Time) (float64, error)
}

type AssetDocumentRepository interface {
	Create(ctx context.Context, document *entity.AssetDocument) error
	Update(ctx context.Context, document *entity.AssetDocument) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.AssetDocument, error)
	ListByAssetID(ctx context.Context, assetID string, documentType *string, offset, limit int) ([]*entity.AssetDocument, int64, error)
	GetByType(ctx context.Context, documentType string, offset, limit int) ([]*entity.AssetDocument, int64, error)
}
