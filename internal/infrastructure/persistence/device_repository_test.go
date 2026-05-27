package persistence

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDBWithMigrate(t *testing.T) *Database {
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
		&entity.EnergyEfficiencyRecord{},
		&entity.EnergyEfficiencyAnalysis{},
		&entity.CarbonEmissionFactor{},
		&entity.CarbonEmissionRecord{},
		&entity.CarbonEmissionSummary{},
	)
	require.NoError(t, err)
	return db
}

func createTestStation(t *testing.T, db *Database, ctx context.Context, code string) *entity.Station {
	t.Helper()
	station := entity.NewStation(code, "Test Station "+code, entity.StationTypePV, "sr-001")
	station.ID = uuid.New().String()
	err := db.WithContext(ctx).Create(station).Error
	require.NoError(t, err)
	return station
}

func TestDeviceRepository_Create(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewDeviceRepository(db)
	ctx := context.Background()

	station := createTestStation(t, db, ctx, "ST001")

	device := entity.NewDevice("DEV001", "Test Device", entity.DeviceTypeInverter, station.ID)
	device.ID = uuid.New().String()
	err := repo.Create(ctx, device)
	assert.NoError(t, err)
	assert.NotEmpty(t, device.ID)
}

func TestDeviceRepository_GetByID(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewDeviceRepository(db)
	ctx := context.Background()

	station := createTestStation(t, db, ctx, "ST001")

	device := entity.NewDevice("DEV001", "Test Device", entity.DeviceTypeInverter, station.ID)
	device.ID = uuid.New().String()
	err := repo.Create(ctx, device)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, device.ID)
	assert.NoError(t, err)
	assert.Equal(t, device.ID, found.ID)
	assert.Equal(t, "DEV001", found.Code)
	assert.Equal(t, "Test Device", found.Name)
	assert.Equal(t, entity.DeviceTypeInverter, found.Type)
}

func TestDeviceRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewDeviceRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestDeviceRepository_GetByCode(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewDeviceRepository(db)
	ctx := context.Background()

	station := createTestStation(t, db, ctx, "ST001")

	device := entity.NewDevice("DEV001", "Test Device", entity.DeviceTypeInverter, station.ID)
	device.ID = uuid.New().String()
	err := repo.Create(ctx, device)
	require.NoError(t, err)

	found, err := repo.GetByCode(ctx, "DEV001")
	assert.NoError(t, err)
	assert.Equal(t, device.ID, found.ID)
}

func TestDeviceRepository_GetByCode_NotFound(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewDeviceRepository(db)
	ctx := context.Background()

	_, err := repo.GetByCode(ctx, "NONEXISTENT")
	assert.Error(t, err)
}

func TestDeviceRepository_Update(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewDeviceRepository(db)
	ctx := context.Background()

	station := createTestStation(t, db, ctx, "ST001")

	device := entity.NewDevice("DEV001", "Test Device", entity.DeviceTypeInverter, station.ID)
	device.ID = uuid.New().String()
	err := repo.Create(ctx, device)
	require.NoError(t, err)

	device.Name = "Updated Device"
	err = repo.Update(ctx, device)
	assert.NoError(t, err)

	found, err := repo.GetByID(ctx, device.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Device", found.Name)
}

func TestDeviceRepository_Update_Status(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewDeviceRepository(db)
	ctx := context.Background()

	station := createTestStation(t, db, ctx, "ST001")

	device := entity.NewDevice("DEV001", "Test Device", entity.DeviceTypeInverter, station.ID)
	device.ID = uuid.New().String()
	err := repo.Create(ctx, device)
	require.NoError(t, err)

	device.SetOnline()
	err = repo.Update(ctx, device)
	assert.NoError(t, err)

	found, err := repo.GetByID(ctx, device.ID)
	assert.NoError(t, err)
	assert.Equal(t, entity.DeviceStatusOnline, found.Status)
}

func TestDeviceRepository_Delete(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewDeviceRepository(db)
	ctx := context.Background()

	station := createTestStation(t, db, ctx, "ST001")

	device := entity.NewDevice("DEV001", "Test Device", entity.DeviceTypeInverter, station.ID)
	device.ID = uuid.New().String()
	err := repo.Create(ctx, device)
	require.NoError(t, err)

	err = repo.Delete(ctx, device.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(ctx, device.ID)
	assert.Error(t, err)
}

func TestDeviceRepository_List(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewDeviceRepository(db)
	ctx := context.Background()

	station := createTestStation(t, db, ctx, "ST001")

	d1 := entity.NewDevice("DEV001", "Device 1", entity.DeviceTypeInverter, station.ID)
	d1.ID = uuid.New().String()
	d2 := entity.NewDevice("DEV002", "Device 2", entity.DeviceTypeMeter, station.ID)
	d2.ID = uuid.New().String()
	err := repo.Create(ctx, d1)
	require.NoError(t, err)
	err = repo.Create(ctx, d2)
	require.NoError(t, err)

	devices, err := repo.List(ctx, nil, nil)
	assert.NoError(t, err)
	assert.Len(t, devices, 2)

	stationID := station.ID
	devices, err = repo.List(ctx, &stationID, nil)
	assert.NoError(t, err)
	assert.Len(t, devices, 2)

	dt := entity.DeviceTypeInverter
	devices, err = repo.List(ctx, &stationID, &dt)
	assert.NoError(t, err)
	assert.Len(t, devices, 1)
	assert.Equal(t, entity.DeviceTypeInverter, devices[0].Type)
}

func TestDeviceRepository_GetOnlineDevices(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewDeviceRepository(db)
	ctx := context.Background()

	station := createTestStation(t, db, ctx, "ST001")

	d1 := entity.NewDevice("DEV001", "Device 1", entity.DeviceTypeInverter, station.ID)
	d1.ID = uuid.New().String()
	d1.Status = 1
	d2 := entity.NewDevice("DEV002", "Device 2", entity.DeviceTypeMeter, station.ID)
	d2.ID = uuid.New().String()
	err := repo.Create(ctx, d1)
	require.NoError(t, err)
	err = repo.Create(ctx, d2)
	require.NoError(t, err)

	_, err = repo.GetOnlineDevices(ctx, station.ID)
	assert.NoError(t, err)
}

func TestDeviceRepository_GetWithPoints(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewDeviceRepository(db)
	ctx := context.Background()

	station := createTestStation(t, db, ctx, "ST001")

	device := entity.NewDevice("DEV001", "Test Device", entity.DeviceTypeInverter, station.ID)
	device.ID = uuid.New().String()
	err := repo.Create(ctx, device)
	require.NoError(t, err)

	point := entity.NewPoint("P001", "Power Point", entity.PointTypeYaoCe)
	point.ID = uuid.New().String()
	point.DeviceID = device.ID
	err = db.WithContext(ctx).Create(point).Error
	require.NoError(t, err)

	found, err := repo.GetWithPoints(ctx, device.ID)
	assert.NoError(t, err)
	assert.Equal(t, device.ID, found.ID)
	if assert.Len(t, found.Points, 1) {
		assert.Equal(t, "P001", found.Points[0].Code)
	}
}

func TestDeviceRepository_GetWithPoints_NotFound(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewDeviceRepository(db)
	ctx := context.Background()

	_, err := repo.GetWithPoints(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestDeviceRepository_List_NoStationFilter(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewDeviceRepository(db)
	ctx := context.Background()

	station := createTestStation(t, db, ctx, "ST001")

	d1 := entity.NewDevice("DEV001", "Device 1", entity.DeviceTypeInverter, station.ID)
	d1.ID = uuid.New().String()
	err := repo.Create(ctx, d1)
	require.NoError(t, err)

	dt := entity.DeviceTypeInverter
	devices, err := repo.List(ctx, nil, &dt)
	assert.NoError(t, err)
	assert.Len(t, devices, 1)
}

func TestDeviceRepository_List_NonMatchingStation(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewDeviceRepository(db)
	ctx := context.Background()

	station := createTestStation(t, db, ctx, "ST001")

	d1 := entity.NewDevice("DEV001", "Device 1", entity.DeviceTypeInverter, station.ID)
	d1.ID = uuid.New().String()
	err := repo.Create(ctx, d1)
	require.NoError(t, err)

	fakeStationID := "nonexistent-station"
	devices, err := repo.List(ctx, &fakeStationID, nil)
	assert.NoError(t, err)
	assert.Len(t, devices, 0)
}
