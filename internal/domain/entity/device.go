package entity

import (
	"time"
)

type DeviceType string
type DeviceStatus int

const (
	DeviceTypeInverter    DeviceType = "inverter"
	DeviceTypeMeter       DeviceType = "meter"
	DeviceTypeTransformer DeviceType = "transformer"
	DeviceTypeSwitch      DeviceType = "switch"
	DeviceTypeWeather     DeviceType = "weather"
	DeviceTypeESS         DeviceType = "ess"
	DeviceTypePCS         DeviceType = "pcs"
	DeviceTypeBMS         DeviceType = "bms"
)

const (
	DeviceStatusOffline    DeviceStatus = 0
	DeviceStatusOnline     DeviceStatus = 1
	DeviceStatusFault      DeviceStatus = 2
	DeviceStatusMaintain   DeviceStatus = 3
)

type Device struct {
	ID           string       `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Code         string       `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name         string       `json:"name" gorm:"type:varchar(200);not null"`
	Type         DeviceType   `json:"type" gorm:"type:varchar(50);not null"`
	StationID    string       `json:"station_id" gorm:"type:varchar(36);index;not null"`
	
	Manufacturer string       `json:"manufacturer" gorm:"type:varchar(100)"`
	Model        string       `json:"model" gorm:"type:varchar(100)"`
	SerialNumber string       `json:"serial_number" gorm:"type:varchar(100)"`
	
	RatedPower   float64      `json:"rated_power"`
	RatedVoltage float64      `json:"rated_voltage"`
	RatedCurrent float64      `json:"rated_current"`
	
	Protocol     string       `json:"protocol" gorm:"type:varchar(50)"`
	IPAddress    string       `json:"ip_address" gorm:"type:varchar(50)"`
	Port         int          `json:"port"`
	SlaveID      int          `json:"slave_id"`
	
	Status       DeviceStatus `json:"status" gorm:"default:0"`
	LastOnline   *time.Time   `json:"last_online"`
	
	InstallDate  *time.Time   `json:"install_date"`
	WarrantyDate *time.Time   `json:"warranty_date"`
	CommissionDate *time.Time `json:"commission_date"`
	DecommissionDate *time.Time `json:"decommission_date"`
	
	LifecycleStatus string    `json:"lifecycle_status" gorm:"type:varchar(50);default:'in_service'"` // in_service, maintenance, decommissioned, retired
	
	Description  string       `json:"description" gorm:"type:text"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	
	Points       []*Point     `json:"points" gorm:"foreignKey:DeviceID"`
	MaintenanceRecords []*MaintenanceRecord `json:"maintenance_records" gorm:"foreignKey:DeviceID"`
	SpareParts   []*SparePart `json:"spare_parts" gorm:"foreignKey:DeviceID"`
	Documents    []*DeviceDocument `json:"documents" gorm:"foreignKey:DeviceID"`
}

func (d *Device) TableName() string {
	return "devices"
}

func NewDevice(code, name string, deviceType DeviceType, stationID string) *Device {
	return &Device{
		Code:      code,
		Name:      name,
		Type:      deviceType,
		StationID: stationID,
		Status:    DeviceStatusOffline,
	}
}

func (d *Device) SetOnline() {
	now := time.Now()
	d.Status = DeviceStatusOnline
	d.LastOnline = &now
}

func (d *Device) SetOffline() {
	d.Status = DeviceStatusOffline
}

func (d *Device) SetFault() {
	d.Status = DeviceStatusFault
}

func (d *Device) SetMaintain() {
	d.Status = DeviceStatusMaintain
}

func (d *Device) IsOnline() bool {
	return d.Status == DeviceStatusOnline
}

func (d *Device) SetCommunication(protocol, ip string, port, slaveID int) {
	d.Protocol = protocol
	d.IPAddress = ip
	d.Port = port
	d.SlaveID = slaveID
}

// MaintenanceRecord 设备维护记录
type MaintenanceRecord struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	DeviceID    string    `json:"device_id" gorm:"type:varchar(36);index;not null"`
	MaintenanceType string `json:"maintenance_type" gorm:"type:varchar(100);not null"` // preventive, corrective, predictive
	Description string    `json:"description" gorm:"type:text"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Status      string    `json:"status" gorm:"type:varchar(50);default:'pending'"` // pending, in_progress, completed, cancelled
	Technician  string    `json:"technician" gorm:"type:varchar(100)"`
	Cost        float64   `json:"cost"`
	Notes       string    `json:"notes" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (m *MaintenanceRecord) TableName() string {
	return "maintenance_records"
}

// SparePart 备件信息
type SparePart struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	DeviceID    string    `json:"device_id" gorm:"type:varchar(36);index"`
	Code        string    `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name        string    `json:"name" gorm:"type:varchar(200);not null"`
	Description string    `json:"description" gorm:"type:text"`
	Quantity    int       `json:"quantity" gorm:"default:0"`
	MinStock    int       `json:"min_stock" gorm:"default:1"`
	UnitPrice   float64   `json:"unit_price"`
	Supplier    string    `json:"supplier" gorm:"type:varchar(100)"`
	Location    string    `json:"location" gorm:"type:varchar(200)"`
	Status      string    `json:"status" gorm:"type:varchar(50);default:'active'"` // active, deprecated, obsolete
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *SparePart) TableName() string {
	return "spare_parts"
}

// DeviceDocument 设备文档
type DeviceDocument struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	DeviceID    string    `json:"device_id" gorm:"type:varchar(36);index;not null"`
	Title       string    `json:"title" gorm:"type:varchar(200);not null"`
	Type        string    `json:"type" gorm:"type:varchar(100)"` // manual, datasheet, certificate, etc.
	FilePath    string    `json:"file_path" gorm:"type:varchar(500)"`
	URL         string    `json:"url" gorm:"type:varchar(500)"`
	Description string    `json:"description" gorm:"type:text"`
	UploadDate  time.Time `json:"upload_date"`
	Version     string    `json:"version" gorm:"type:varchar(50)"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (d *DeviceDocument) TableName() string {
	return "device_documents"
}

// WorkOrder 工单信息
type WorkOrder struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	DeviceID    string    `json:"device_id" gorm:"type:varchar(36);index"`
	Type        string    `json:"type" gorm:"type:varchar(100);not null"` // maintenance, repair, inspection
	Title       string    `json:"title" gorm:"type:varchar(200);not null"`
	Description string    `json:"description" gorm:"type:text"`
	Priority    string    `json:"priority" gorm:"type:varchar(50);default:'medium'"` // low, medium, high, urgent
	Status      string    `json:"status" gorm:"type:varchar(50);default:'open'"` // open, in_progress, completed, cancelled
	Assignee    string    `json:"assignee" gorm:"type:varchar(100)"`
	DueDate     *time.Time `json:"due_date"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	CreatedBy   string    `json:"created_by" gorm:"type:varchar(100)"`
	Notes       string    `json:"notes" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (w *WorkOrder) TableName() string {
	return "work_orders"
}

// Inventory 库存信息
type Inventory struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Code        string    `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name        string    `json:"name" gorm:"type:varchar(200);not null"`
	Type        string    `json:"type" gorm:"type:varchar(100);index"` // raw_material, spare_part, tool, consumable
	Spec        string    `json:"spec" gorm:"type:varchar(200)"`
	Unit        string    `json:"unit" gorm:"type:varchar(50);not null"`
	Quantity    float64   `json:"quantity" gorm:"default:0"`
	MinQuantity float64   `json:"min_quantity" gorm:"default:0"`
	MaxQuantity float64   `json:"max_quantity" gorm:"default:1000"`
	Location    string    `json:"location" gorm:"type:varchar(200)"`
	Status      string    `json:"status" gorm:"type:varchar(50);default:'normal'"` // normal, low_stock, out_of_stock, expired
	SupplierID  string    `json:"supplier_id" gorm:"type:varchar(36);index"`
	UnitPrice   float64   `json:"unit_price" gorm:"default:0"`
	TotalValue  float64   `json:"total_value" gorm:"default:0"`
	LastInStock *time.Time `json:"last_in_stock"`
	LastOutStock *time.Time `json:"last_out_stock"`
	ExpiryDate  *time.Time `json:"expiry_date"`
	Description string    `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (i *Inventory) TableName() string {
	return "inventory"
}

// InventoryTransaction 库存交易记录
type InventoryTransaction struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	InventoryID string    `json:"inventory_id" gorm:"type:varchar(36);index;not null"`
	Type        string    `json:"type" gorm:"type:varchar(50);not null"` // in, out, adjustment
	Quantity    float64   `json:"quantity" gorm:"not null"`
	UnitPrice   float64   `json:"unit_price" gorm:"default:0"`
	TotalAmount float64   `json:"total_amount" gorm:"default:0"`
	BeforeQty   float64   `json:"before_qty"`
	AfterQty    float64   `json:"after_qty"`
	ReferenceID string    `json:"reference_id" gorm:"type:varchar(36);index"` // purchase_order_id, work_order_id, etc.
	ReferenceType string    `json:"reference_type" gorm:"type:varchar(100)"` // purchase_order, work_order, etc.
	OperatorID  string    `json:"operator_id" gorm:"type:varchar(36);not null"`
	Notes       string    `json:"notes" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
}

func (it *InventoryTransaction) TableName() string {
	return "inventory_transactions"
}

// Supplier 供应商信息
type Supplier struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Code        string    `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name        string    `json:"name" gorm:"type:varchar(200);not null"`
	ContactName string    `json:"contact_name" gorm:"type:varchar(100)"`
	ContactPhone string    `json:"contact_phone" gorm:"type:varchar(50)"`
	ContactEmail string    `json:"contact_email" gorm:"type:varchar(100)"`
	Address     string    `json:"address" gorm:"type:text"`
	TaxID       string    `json:"tax_id" gorm:"type:varchar(100)"`
	BankInfo    string    `json:"bank_info" gorm:"type:text"`
	Status      string    `json:"status" gorm:"type:varchar(50);default:'active'"` // active, inactive, blacklisted
	Description string    `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *Supplier) TableName() string {
	return "suppliers"
}

// PurchaseOrder 采购订单
type PurchaseOrder struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Code        string    `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	SupplierID  string    `json:"supplier_id" gorm:"type:varchar(36);index;not null"`
	OrderDate   time.Time `json:"order_date" gorm:"not null"`
	ExpectedDate *time.Time `json:"expected_date"`
	ActualDate  *time.Time `json:"actual_date"`
	Status      string    `json:"status" gorm:"type:varchar(50);default:'draft'"` // draft, pending, approved, ordered, received, cancelled
	TotalAmount float64   `json:"total_amount" gorm:"default:0"`
	TaxAmount   float64   `json:"tax_amount" gorm:"default:0"`
	GrandTotal  float64   `json:"grand_total" gorm:"default:0"`
	CreatedBy   string    `json:"created_by" gorm:"type:varchar(100);not null"`
	ApprovedBy  *string   `json:"approved_by" gorm:"type:varchar(100)"`
	ApprovedAt  *time.Time `json:"approved_at"`
	Notes       string    `json:"notes" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Items       []*PurchaseOrderItem `json:"items" gorm:"foreignKey:PurchaseOrderID"`
	Supplier    *Supplier          `json:"supplier" gorm:"foreignKey:SupplierID"`
}

func (po *PurchaseOrder) TableName() string {
	return "purchase_orders"
}

// PurchaseOrderItem 采购订单项
type PurchaseOrderItem struct {
	ID              string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	PurchaseOrderID string    `json:"purchase_order_id" gorm:"type:varchar(36);index;not null"`
	InventoryID     *string   `json:"inventory_id" gorm:"type:varchar(36);index"`
	ItemCode        string    `json:"item_code" gorm:"type:varchar(100);not null"`
	ItemName        string    `json:"item_name" gorm:"type:varchar(200);not null"`
	Specification   string    `json:"specification" gorm:"type:varchar(200)"`
	Quantity        float64   `json:"quantity" gorm:"not null"`
	Unit            string    `json:"unit" gorm:"type:varchar(50);not null"`
	UnitPrice       float64   `json:"unit_price" gorm:"not null"`
	Subtotal        float64   `json:"subtotal" gorm:"not null"`
	TaxRate         float64   `json:"tax_rate" gorm:"default:0"`
	TaxAmount       float64   `json:"tax_amount" gorm:"default:0"`
	TotalAmount     float64   `json:"total_amount" gorm:"not null"`
	ReceivedQuantity float64   `json:"received_quantity" gorm:"default:0"`
	Status          string    `json:"status" gorm:"type:varchar(50);default:'pending'"` // pending, partially_received, received, cancelled
	Notes           string    `json:"notes" gorm:"type:text"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	Inventory       *Inventory `json:"inventory" gorm:"foreignKey:InventoryID"`
}

func (poi *PurchaseOrderItem) TableName() string {
	return "purchase_order_items"
}

// Receipt 收货单
type Receipt struct {
	ID              string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Code            string    `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	PurchaseOrderID string    `json:"purchase_order_id" gorm:"type:varchar(36);index;not null"`
	ReceiptDate     time.Time `json:"receipt_date" gorm:"not null"`
	ReceivedBy      string    `json:"received_by" gorm:"type:varchar(100);not null"`
	Status          string    `json:"status" gorm:"type:varchar(50);default:'draft'"` // draft, completed, cancelled
	TotalItems      int       `json:"total_items" gorm:"default:0"`
	Notes           string    `json:"notes" gorm:"type:text"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	Items           []*ReceiptItem `json:"items" gorm:"foreignKey:ReceiptID"`
	PurchaseOrder   *PurchaseOrder `json:"purchase_order" gorm:"foreignKey:PurchaseOrderID"`
}

func (r *Receipt) TableName() string {
	return "receipts"
}

// ReceiptItem 收货单项
type ReceiptItem struct {
	ID              string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	ReceiptID       string    `json:"receipt_id" gorm:"type:varchar(36);index;not null"`
	PurchaseOrderItemID string `json:"purchase_order_item_id" gorm:"type:varchar(36);index;not null"`
	InventoryID     string    `json:"inventory_id" gorm:"type:varchar(36);index;not null"`
	Quantity        float64   `json:"quantity" gorm:"not null"`
	Unit            string    `json:"unit" gorm:"type:varchar(50);not null"`
	UnitPrice       float64   `json:"unit_price" gorm:"not null"`
	TotalAmount     float64   `json:"total_amount" gorm:"not null"`
	Status          string    `json:"status" gorm:"type:varchar(50);default:'received'"` // received, damaged, returned
	Notes           string    `json:"notes" gorm:"type:text"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	PurchaseOrderItem *PurchaseOrderItem `json:"purchase_order_item" gorm:"foreignKey:PurchaseOrderItemID"`
	Inventory         *Inventory         `json:"inventory" gorm:"foreignKey:InventoryID"`
}

func (ri *ReceiptItem) TableName() string {
	return "receipt_items"
}

// CostCategory 成本类别
type CostCategory struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Code        string    `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name        string    `json:"name" gorm:"type:varchar(200);not null"`
	ParentID    *string   `json:"parent_id" gorm:"type:varchar(36);index"`
	Type        string    `json:"type" gorm:"type:varchar(100);not null"` // direct, indirect, fixed, variable
	Description string    `json:"description" gorm:"type:text"`
	Status      string    `json:"status" gorm:"type:varchar(50);default:'active'"` // active, inactive
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (cc *CostCategory) TableName() string {
	return "cost_categories"
}

// CostEntry 成本条目
type CostEntry struct {
	ID            string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Code          string    `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Date          time.Time `json:"date" gorm:"not null"`
	CostCategoryID string    `json:"cost_category_id" gorm:"type:varchar(36);index;not null"`
	Amount        float64   `json:"amount" gorm:"not null"`
	Currency      string    `json:"currency" gorm:"type:varchar(10);default:'CNY'"`
	Description   string    `json:"description" gorm:"type:text"`
	ReferenceID   string    `json:"reference_id" gorm:"type:varchar(36);index"` // purchase_order_id, work_order_id, etc.
	ReferenceType string    `json:"reference_type" gorm:"type:varchar(100)"` // purchase_order, work_order, etc.
	DepartmentID  string    `json:"department_id" gorm:"type:varchar(36);index"`
	ProjectID     string    `json:"project_id" gorm:"type:varchar(36);index"`
	DeviceID      string    `json:"device_id" gorm:"type:varchar(36);index"`
	StationID     string    `json:"station_id" gorm:"type:varchar(36);index"`
	ApprovalStatus string    `json:"approval_status" gorm:"type:varchar(50);default:'pending'"` // pending, approved, rejected
	ApprovedBy    *string   `json:"approved_by" gorm:"type:varchar(100)"`
	ApprovedAt    *time.Time `json:"approved_at"`
	Notes         string    `json:"notes" gorm:"type:text"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	CostCategory  *CostCategory `json:"cost_category" gorm:"foreignKey:CostCategoryID"`
}

func (ce *CostEntry) TableName() string {
	return "cost_entries"
}

// CostAllocation 成本分配
type CostAllocation struct {
	ID            string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	CostEntryID   string    `json:"cost_entry_id" gorm:"type:varchar(36);index;not null"`
	AllocatedTo   string    `json:"allocated_to" gorm:"type:varchar(100);not null"` // project, device, station, department
	AllocatedID   string    `json:"allocated_id" gorm:"type:varchar(36);index;not null"`
	Amount        float64   `json:"amount" gorm:"not null"`
	Percentage    float64   `json:"percentage" gorm:"not null"`
	Description   string    `json:"description" gorm:"type:text"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	CostEntry     *CostEntry `json:"cost_entry" gorm:"foreignKey:CostEntryID"`
}

func (ca *CostAllocation) TableName() string {
	return "cost_allocations"
}

// CostReport 成本报表
type CostReport struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Code        string    `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name        string    `json:"name" gorm:"type:varchar(200);not null"`
	ReportType  string    `json:"report_type" gorm:"type:varchar(100);not null"` // monthly, quarterly, annual, project
	PeriodStart time.Time `json:"period_start" gorm:"not null"`
	PeriodEnd   time.Time `json:"period_end" gorm:"not null"`
	TotalCost   float64   `json:"total_cost" gorm:"default:0"`
	Currency    string    `json:"currency" gorm:"type:varchar(10);default:'CNY'"`
	Status      string    `json:"status" gorm:"type:varchar(50);default:'draft'"` // draft, generated, approved
	GeneratedBy string    `json:"generated_by" gorm:"type:varchar(100);not null"`
	ApprovedBy  *string   `json:"approved_by" gorm:"type:varchar(100)"`
	ApprovedAt  *time.Time `json:"approved_at"`
	Notes       string    `json:"notes" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (cr *CostReport) TableName() string {
	return "cost_reports"
}

// Asset 资产信息
type Asset struct {
	ID              string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Code            string    `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name            string    `json:"name" gorm:"type:varchar(200);not null"`
	AssetType       string    `json:"asset_type" gorm:"type:varchar(100);index;not null"` // device, equipment, building, land, etc.
	Category        string    `json:"category" gorm:"type:varchar(100);index"`
	Manufacturer    string    `json:"manufacturer" gorm:"type:varchar(100)"`
	Model           string    `json:"model" gorm:"type:varchar(100)"`
	SerialNumber    string    `json:"serial_number" gorm:"type:varchar(100)"`
	PurchaseDate    time.Time `json:"purchase_date"`
	InstallDate     *time.Time `json:"install_date"`
	WarrantyEndDate *time.Time `json:"warranty_end_date"`
	DecommissionDate *time.Time `json:"decommission_date"`
	Cost            float64   `json:"cost" gorm:"type:decimal(10,2)"`
	CurrentValue    float64   `json:"current_value" gorm:"type:decimal(10,2)"`
	Location        string    `json:"location" gorm:"type:varchar(200)"`
	Status          string    `json:"status" gorm:"type:varchar(50);default:'active'"` // active, maintenance, decommissioned, disposed
	UsageStatus     string    `json:"usage_status" gorm:"type:varchar(50);default:'in_use'"` // in_use, idle, stored
	DepreciationMethod string  `json:"depreciation_method" gorm:"type:varchar(100)"` // straight_line, declining_balance, sum_of_years, units_of_production
	UsefulLife      int       `json:"useful_life"` // in years
	SalvageValue    float64   `json:"salvage_value" gorm:"type:decimal(10,2)"`
	DepartmentID    string    `json:"department_id" gorm:"type:varchar(36);index"`
	ResponsiblePerson string   `json:"responsible_person" gorm:"type:varchar(100)"`
	Description     string    `json:"description" gorm:"type:text"`
	Notes           string    `json:"notes" gorm:"type:text"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	MaintenanceRecords []*AssetMaintenanceRecord `json:"maintenance_records" gorm:"foreignKey:AssetID"`
	DepreciationRecords []*AssetDepreciationRecord `json:"depreciation_records" gorm:"foreignKey:AssetID"`
	Documents       []*AssetDocument `json:"documents" gorm:"foreignKey:AssetID"`
}

func (a *Asset) TableName() string {
	return "assets"
}

// AssetMaintenanceRecord 资产维护记录
type AssetMaintenanceRecord struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	AssetID     string    `json:"asset_id" gorm:"type:varchar(36);index;not null"`
	MaintenanceType string `json:"maintenance_type" gorm:"type:varchar(100);not null"` // preventive, corrective, predictive
	Title       string    `json:"title" gorm:"type:varchar(200);not null"`
	Description string    `json:"description" gorm:"type:text"`
	StartDate   time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	Status      string    `json:"status" gorm:"type:varchar(50);default:'pending'"` // pending, in_progress, completed, cancelled
	Cost        float64   `json:"cost" gorm:"type:decimal(10,2)"`
	Technician  string    `json:"technician" gorm:"type:varchar(100)"`
	Notes       string    `json:"notes" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Asset       *Asset    `json:"asset" gorm:"foreignKey:AssetID"`
}

func (amr *AssetMaintenanceRecord) TableName() string {
	return "asset_maintenance_records"
}

// AssetDepreciationRecord 资产折旧记录
type AssetDepreciationRecord struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	AssetID     string    `json:"asset_id" gorm:"type:varchar(36);index;not null"`
	DepreciationDate time.Time `json:"depreciation_date"`
	Period      string    `json:"period" gorm:"type:varchar(50);not null"` // monthly, quarterly, annual
	DepreciationAmount float64   `json:"depreciation_amount" gorm:"type:decimal(10,2);not null"`
	AccumulatedDepreciation float64 `json:"accumulated_depreciation" gorm:"type:decimal(10,2);not null"`
	BookValue   float64   `json:"book_value" gorm:"type:decimal(10,2);not null"`
	Notes       string    `json:"notes" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Asset       *Asset    `json:"asset" gorm:"foreignKey:AssetID"`
}

func (adr *AssetDepreciationRecord) TableName() string {
	return "asset_depreciation_records"
}

// AssetDocument 资产文档
type AssetDocument struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	AssetID     string    `json:"asset_id" gorm:"type:varchar(36);index;not null"`
	Title       string    `json:"title" gorm:"type:varchar(200);not null"`
	Type        string    `json:"type" gorm:"type:varchar(100)"` // manual, datasheet, certificate, invoice, etc.
	FilePath    string    `json:"file_path" gorm:"type:varchar(500)"`
	URL         string    `json:"url" gorm:"type:varchar(500)"`
	Description string    `json:"description" gorm:"type:text"`
	UploadDate  time.Time `json:"upload_date"`
	Version     string    `json:"version" gorm:"type:varchar(50)"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Asset       *Asset    `json:"asset" gorm:"foreignKey:AssetID"`
}

func (ad *AssetDocument) TableName() string {
	return "asset_documents"
}
