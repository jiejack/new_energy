package service

import (
	"context"
	"testing"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAssetRepoSvc struct {
	mock.Mock
}

func (m *mockAssetRepoSvc) Create(ctx context.Context, asset *entity.Asset) error {
	return m.Called(ctx, asset).Error(0)
}

func (m *mockAssetRepoSvc) Update(ctx context.Context, asset *entity.Asset) error {
	return m.Called(ctx, asset).Error(0)
}

func (m *mockAssetRepoSvc) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockAssetRepoSvc) GetByID(ctx context.Context, id string) (*entity.Asset, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Asset), args.Error(1)
}

func (m *mockAssetRepoSvc) GetByCode(ctx context.Context, code string) (*entity.Asset, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Asset), args.Error(1)
}

func (m *mockAssetRepoSvc) List(ctx context.Context, assetType *string, status *string, category *string, offset, limit int) ([]*entity.Asset, int64, error) {
	args := m.Called(ctx, assetType, status, category, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.Asset), args.Get(1).(int64), args.Error(2)
}

func (m *mockAssetRepoSvc) GetByLocation(ctx context.Context, location string) ([]*entity.Asset, error) {
	args := m.Called(ctx, location)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Asset), args.Error(1)
}

func (m *mockAssetRepoSvc) GetByDepartment(ctx context.Context, departmentID string) ([]*entity.Asset, error) {
	args := m.Called(ctx, departmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Asset), args.Error(1)
}

func (m *mockAssetRepoSvc) GetByResponsiblePerson(ctx context.Context, person string) ([]*entity.Asset, error) {
	args := m.Called(ctx, person)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Asset), args.Error(1)
}

func (m *mockAssetRepoSvc) GetDepreciatingAssets(ctx context.Context) ([]*entity.Asset, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Asset), args.Error(1)
}

func (m *mockAssetRepoSvc) GetAssetsNearWarrantyEnd(ctx context.Context, days int) ([]*entity.Asset, error) {
	args := m.Called(ctx, days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Asset), args.Error(1)
}

type mockAssetMaintRepoSvc struct {
	mock.Mock
}

func (m *mockAssetMaintRepoSvc) Create(ctx context.Context, record *entity.AssetMaintenanceRecord) error {
	return m.Called(ctx, record).Error(0)
}

func (m *mockAssetMaintRepoSvc) Update(ctx context.Context, record *entity.AssetMaintenanceRecord) error {
	return m.Called(ctx, record).Error(0)
}

func (m *mockAssetMaintRepoSvc) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockAssetMaintRepoSvc) GetByID(ctx context.Context, id string) (*entity.AssetMaintenanceRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AssetMaintenanceRecord), args.Error(1)
}

func (m *mockAssetMaintRepoSvc) ListByAssetID(ctx context.Context, assetID string, status *string, maintenanceType *string, offset, limit int) ([]*entity.AssetMaintenanceRecord, int64, error) {
	args := m.Called(ctx, assetID, status, maintenanceType, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.AssetMaintenanceRecord), args.Get(1).(int64), args.Error(2)
}

func (m *mockAssetMaintRepoSvc) ListByStatus(ctx context.Context, status string, offset, limit int) ([]*entity.AssetMaintenanceRecord, int64, error) {
	args := m.Called(ctx, status, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.AssetMaintenanceRecord), args.Get(1).(int64), args.Error(2)
}

func (m *mockAssetMaintRepoSvc) GetMaintenanceCostByAsset(ctx context.Context, assetID string, startDate, endDate *time.Time) (float64, error) {
	args := m.Called(ctx, assetID, startDate, endDate)
	return args.Get(0).(float64), args.Error(1)
}

type mockAssetDepRepoSvc struct {
	mock.Mock
}

func (m *mockAssetDepRepoSvc) Create(ctx context.Context, record *entity.AssetDepreciationRecord) error {
	return m.Called(ctx, record).Error(0)
}

func (m *mockAssetDepRepoSvc) Update(ctx context.Context, record *entity.AssetDepreciationRecord) error {
	return m.Called(ctx, record).Error(0)
}

func (m *mockAssetDepRepoSvc) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockAssetDepRepoSvc) GetByID(ctx context.Context, id string) (*entity.AssetDepreciationRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AssetDepreciationRecord), args.Error(1)
}

func (m *mockAssetDepRepoSvc) ListByAssetID(ctx context.Context, assetID string, period *string, offset, limit int) ([]*entity.AssetDepreciationRecord, int64, error) {
	args := m.Called(ctx, assetID, period, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.AssetDepreciationRecord), args.Get(1).(int64), args.Error(2)
}

func (m *mockAssetDepRepoSvc) GetLatestByAssetID(ctx context.Context, assetID string) (*entity.AssetDepreciationRecord, error) {
	args := m.Called(ctx, assetID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AssetDepreciationRecord), args.Error(1)
}

func (m *mockAssetDepRepoSvc) GetDepreciationSummaryByPeriod(ctx context.Context, period string, startDate, endDate *time.Time) (float64, error) {
	args := m.Called(ctx, period, startDate, endDate)
	return args.Get(0).(float64), args.Error(1)
}

type mockAssetDocRepoSvc struct {
	mock.Mock
}

func (m *mockAssetDocRepoSvc) Create(ctx context.Context, document *entity.AssetDocument) error {
	return m.Called(ctx, document).Error(0)
}

func (m *mockAssetDocRepoSvc) Update(ctx context.Context, document *entity.AssetDocument) error {
	return m.Called(ctx, document).Error(0)
}

func (m *mockAssetDocRepoSvc) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockAssetDocRepoSvc) GetByID(ctx context.Context, id string) (*entity.AssetDocument, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AssetDocument), args.Error(1)
}

func (m *mockAssetDocRepoSvc) ListByAssetID(ctx context.Context, assetID string, documentType *string, offset, limit int) ([]*entity.AssetDocument, int64, error) {
	args := m.Called(ctx, assetID, documentType, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.AssetDocument), args.Get(1).(int64), args.Error(2)
}

func (m *mockAssetDocRepoSvc) GetByType(ctx context.Context, documentType string, offset, limit int) ([]*entity.AssetDocument, int64, error) {
	args := m.Called(ctx, documentType, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.AssetDocument), args.Get(1).(int64), args.Error(2)
}

var _ repository.AssetRepository = (*mockAssetRepoSvc)(nil)
var _ repository.AssetMaintenanceRepository = (*mockAssetMaintRepoSvc)(nil)
var _ repository.AssetDepreciationRepository = (*mockAssetDepRepoSvc)(nil)
var _ repository.AssetDocumentRepository = (*mockAssetDocRepoSvc)(nil)

func TestAssetService_CreateAsset_Success(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	depRepo := new(mockAssetDepRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetService(assetRepo, maintRepo, depRepo, docRepo)

	assetRepo.On("GetByCode", mock.Anything, "ASSET001").Return(nil, assert.AnError)
	assetRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Asset")).Return(nil)

	req := &CreateAssetRequest{
		Code: "ASSET001", Name: "Test Asset", Category: "equipment",
		PurchasePrice: 100000, PurchaseDate: "2024-01-01", ExpectedLife: 10,
	}
	asset, err := svc.CreateAsset(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, asset)
	assert.Equal(t, "ASSET001", asset.Code)
}

func TestAssetService_CreateAsset_DuplicateCode(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	depRepo := new(mockAssetDepRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetService(assetRepo, maintRepo, depRepo, docRepo)

	assetRepo.On("GetByCode", mock.Anything, "ASSET001").Return(&entity.Asset{Code: "ASSET001"}, nil)

	req := &CreateAssetRequest{
		Code: "ASSET001", Name: "Test Asset", Category: "equipment",
		PurchasePrice: 100000, PurchaseDate: "2024-01-01", ExpectedLife: 10,
	}
	asset, err := svc.CreateAsset(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, asset)
}

func TestAssetService_GetAsset_Success(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	depRepo := new(mockAssetDepRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetService(assetRepo, maintRepo, depRepo, docRepo)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)

	asset, err := svc.GetAsset(context.Background(), "asset-001")
	assert.NoError(t, err)
	assert.NotNil(t, asset)
}

func TestAssetService_UpdateAsset_Success(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	depRepo := new(mockAssetDepRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetService(assetRepo, maintRepo, depRepo, docRepo)

	existing := &entity.Asset{ID: "asset-001", Code: "ASSET001"}
	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(existing, nil)
	assetRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.Asset")).Return(nil)

	req := &UpdateAssetRequest{Name: "Updated Asset"}
	asset, err := svc.UpdateAsset(context.Background(), "asset-001", req)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Asset", asset.Name)
}

func TestAssetService_UpdateAsset_NotFound(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	depRepo := new(mockAssetDepRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetService(assetRepo, maintRepo, depRepo, docRepo)

	assetRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	req := &UpdateAssetRequest{Name: "Updated"}
	asset, err := svc.UpdateAsset(context.Background(), "nonexistent", req)
	assert.Error(t, err)
	assert.Nil(t, asset)
}

func TestAssetService_DeleteAsset_Success(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	depRepo := new(mockAssetDepRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetService(assetRepo, maintRepo, depRepo, docRepo)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	maintRepo.On("ListByAssetID", mock.Anything, "asset-001", (*string)(nil), (*string)(nil), 0, 1).Return([]*entity.AssetMaintenanceRecord{}, int64(0), nil)
	depRepo.On("ListByAssetID", mock.Anything, "asset-001", (*string)(nil), 0, 1).Return([]*entity.AssetDepreciationRecord{}, int64(0), nil)
	docRepo.On("ListByAssetID", mock.Anything, "asset-001", (*string)(nil), 0, 1).Return([]*entity.AssetDocument{}, int64(0), nil)
	assetRepo.On("Delete", mock.Anything, "asset-001").Return(nil)

	err := svc.DeleteAsset(context.Background(), "asset-001")
	assert.NoError(t, err)
}

func TestAssetService_DeleteAsset_HasMaintenanceRecords(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	depRepo := new(mockAssetDepRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetService(assetRepo, maintRepo, depRepo, docRepo)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	maintRepo.On("ListByAssetID", mock.Anything, "asset-001", (*string)(nil), (*string)(nil), 0, 1).Return([]*entity.AssetMaintenanceRecord{{ID: "m-001"}}, int64(1), nil)

	err := svc.DeleteAsset(context.Background(), "asset-001")
	assert.Error(t, err)
}

func TestAssetService_ListAssets_Success(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	depRepo := new(mockAssetDepRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetService(assetRepo, maintRepo, depRepo, docRepo)

	assetRepo.On("List", mock.Anything, (*string)(nil), (*string)(nil), (*string)(nil), 0, 10).Return([]*entity.Asset{{ID: "asset-001"}}, int64(1), nil)

	assets, total, err := svc.ListAssets(context.Background(), "", "", "", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, assets, 1)
}

func TestAssetService_CalculateDepreciation_StraightLine(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	depRepo := new(mockAssetDepRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetService(assetRepo, maintRepo, depRepo, docRepo)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001", Cost: 100000, UsefulLife: 10, SalvageValue: 10000}, nil)
	depRepo.On("GetLatestByAssetID", mock.Anything, "asset-001").Return(nil, assert.AnError)

	result, err := svc.CalculateDepreciation(context.Background(), "asset-001", "straight-line")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 9000.0, result["annual_depreciation"])
}

func TestAssetService_CalculateDepreciation_DecliningBalance(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	depRepo := new(mockAssetDepRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetService(assetRepo, maintRepo, depRepo, docRepo)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001", Cost: 100000, UsefulLife: 10, SalvageValue: 10000}, nil)
	depRepo.On("GetLatestByAssetID", mock.Anything, "asset-001").Return(nil, assert.AnError)

	result, err := svc.CalculateDepreciation(context.Background(), "asset-001", "declining-balance")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestAssetService_CalculateDepreciation_UnsupportedMethod(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	depRepo := new(mockAssetDepRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetService(assetRepo, maintRepo, depRepo, docRepo)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001", Cost: 100000, UsefulLife: 10, SalvageValue: 10000}, nil)
	depRepo.On("GetLatestByAssetID", mock.Anything, "asset-001").Return(nil, assert.AnError)

	result, err := svc.CalculateDepreciation(context.Background(), "asset-001", "invalid")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAssetDepreciationService_CreateDepreciationRecord_Success(t *testing.T) {
	depRepo := new(mockAssetDepRepoSvc)
	assetRepo := new(mockAssetRepoSvc)
	svc := NewAssetDepreciationService(depRepo, assetRepo)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	depRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.AssetDepreciationRecord")).Return(nil)

	req := &CreateDepreciationRequest{
		AssetID: "asset-001", DepreciationMethod: "straight-line",
		Year: 2024, Amount: 9000, AccumulatedAmount: 9000, BookValue: 91000,
	}
	record, err := svc.CreateDepreciationRecord(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, record)
}

func TestAssetDepreciationService_CreateDepreciationRecord_AssetNotFound(t *testing.T) {
	depRepo := new(mockAssetDepRepoSvc)
	assetRepo := new(mockAssetRepoSvc)
	svc := NewAssetDepreciationService(depRepo, assetRepo)

	assetRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	req := &CreateDepreciationRequest{AssetID: "nonexistent", DepreciationMethod: "straight-line", Year: 2024, Amount: 9000, AccumulatedAmount: 9000, BookValue: 91000}
	record, err := svc.CreateDepreciationRecord(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, record)
}

func TestAssetDepreciationService_GetDepreciationRecord_Success(t *testing.T) {
	depRepo := new(mockAssetDepRepoSvc)
	assetRepo := new(mockAssetRepoSvc)
	svc := NewAssetDepreciationService(depRepo, assetRepo)

	depRepo.On("GetByID", mock.Anything, "dep-001").Return(&entity.AssetDepreciationRecord{ID: "dep-001"}, nil)

	record, err := svc.GetDepreciationRecord(context.Background(), "dep-001")
	assert.NoError(t, err)
	assert.NotNil(t, record)
}

func TestAssetDepreciationService_UpdateDepreciationRecord_Success(t *testing.T) {
	depRepo := new(mockAssetDepRepoSvc)
	assetRepo := new(mockAssetRepoSvc)
	svc := NewAssetDepreciationService(depRepo, assetRepo)

	existing := &entity.AssetDepreciationRecord{ID: "dep-001", AssetID: "asset-001", Period: "annual"}
	depRepo.On("GetByID", mock.Anything, "dep-001").Return(existing, nil)
	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	depRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.AssetDepreciationRecord")).Return(nil)

	req := &UpdateDepreciationRequest{AssetID: "asset-001", Amount: 9500, AccumulatedAmount: 9500, BookValue: 90500}
	record, err := svc.UpdateDepreciationRecord(context.Background(), "dep-001", req)
	assert.NoError(t, err)
	assert.NotNil(t, record)
}

func TestAssetDepreciationService_DeleteDepreciationRecord_Success(t *testing.T) {
	depRepo := new(mockAssetDepRepoSvc)
	assetRepo := new(mockAssetRepoSvc)
	svc := NewAssetDepreciationService(depRepo, assetRepo)

	depRepo.On("GetByID", mock.Anything, "dep-001").Return(&entity.AssetDepreciationRecord{ID: "dep-001"}, nil)
	depRepo.On("Delete", mock.Anything, "dep-001").Return(nil)

	err := svc.DeleteDepreciationRecord(context.Background(), "dep-001")
	assert.NoError(t, err)
}

func TestAssetDepreciationService_ListDepreciationRecords_Success(t *testing.T) {
	depRepo := new(mockAssetDepRepoSvc)
	assetRepo := new(mockAssetRepoSvc)
	svc := NewAssetDepreciationService(depRepo, assetRepo)

	depRepo.On("ListByAssetID", mock.Anything, "", (*string)(nil), 0, 10).Return([]*entity.AssetDepreciationRecord{{ID: "dep-001"}}, int64(1), nil)

	records, total, err := svc.ListDepreciationRecords(context.Background(), "", "", 0, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, records, 1)
}

func TestAssetDepreciationService_GetDepreciationSummary_Success(t *testing.T) {
	depRepo := new(mockAssetDepRepoSvc)
	assetRepo := new(mockAssetRepoSvc)
	svc := NewAssetDepreciationService(depRepo, assetRepo)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	depRepo.On("GetDepreciationSummaryByPeriod", mock.Anything, "annual", (*time.Time)(nil), mock.Anything).Return(9000.0, nil)

	summary, err := svc.GetDepreciationSummary(context.Background(), "asset-001", "")
	assert.NoError(t, err)
	assert.Equal(t, 9000.0, summary)
}

func TestAssetMaintenanceService_CreateMaintenanceRecord_Success(t *testing.T) {
	maintRepo := new(mockAssetMaintRepoSvc)
	assetRepo := new(mockAssetRepoSvc)
	svc := NewAssetMaintenanceService(maintRepo, assetRepo)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	maintRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.AssetMaintenanceRecord")).Return(nil)

	req := &CreateMaintenanceRequest{
		AssetID: "asset-001", MaintenanceType: "preventive",
		MaintenanceDate: "2024-01-15", Cost: 5000,
	}
	record, err := svc.CreateMaintenanceRecord(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, record)
}

func TestAssetMaintenanceService_CreateMaintenanceRecord_AssetNotFound(t *testing.T) {
	maintRepo := new(mockAssetMaintRepoSvc)
	assetRepo := new(mockAssetRepoSvc)
	svc := NewAssetMaintenanceService(maintRepo, assetRepo)

	assetRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	req := &CreateMaintenanceRequest{AssetID: "nonexistent", MaintenanceType: "preventive", MaintenanceDate: "2024-01-15"}
	record, err := svc.CreateMaintenanceRecord(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, record)
}

func TestAssetMaintenanceService_GetMaintenanceRecord_Success(t *testing.T) {
	maintRepo := new(mockAssetMaintRepoSvc)
	assetRepo := new(mockAssetRepoSvc)
	svc := NewAssetMaintenanceService(maintRepo, assetRepo)

	maintRepo.On("GetByID", mock.Anything, "maint-001").Return(&entity.AssetMaintenanceRecord{ID: "maint-001"}, nil)

	record, err := svc.GetMaintenanceRecord(context.Background(), "maint-001")
	assert.NoError(t, err)
	assert.NotNil(t, record)
}

func TestAssetMaintenanceService_UpdateMaintenanceRecord_Success(t *testing.T) {
	maintRepo := new(mockAssetMaintRepoSvc)
	assetRepo := new(mockAssetRepoSvc)
	svc := NewAssetMaintenanceService(maintRepo, assetRepo)

	existing := &entity.AssetMaintenanceRecord{ID: "maint-001", AssetID: "asset-001"}
	maintRepo.On("GetByID", mock.Anything, "maint-001").Return(existing, nil)
	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	maintRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.AssetMaintenanceRecord")).Return(nil)

	req := &UpdateMaintenanceRequest{AssetID: "asset-001", Status: "completed"}
	record, err := svc.UpdateMaintenanceRecord(context.Background(), "maint-001", req)
	assert.NoError(t, err)
	assert.NotNil(t, record)
}

func TestAssetMaintenanceService_DeleteMaintenanceRecord_Success(t *testing.T) {
	maintRepo := new(mockAssetMaintRepoSvc)
	assetRepo := new(mockAssetRepoSvc)
	svc := NewAssetMaintenanceService(maintRepo, assetRepo)

	maintRepo.On("GetByID", mock.Anything, "maint-001").Return(&entity.AssetMaintenanceRecord{ID: "maint-001"}, nil)
	maintRepo.On("Delete", mock.Anything, "maint-001").Return(nil)

	err := svc.DeleteMaintenanceRecord(context.Background(), "maint-001")
	assert.NoError(t, err)
}

func TestAssetMaintenanceService_ListMaintenanceRecords_Success(t *testing.T) {
	maintRepo := new(mockAssetMaintRepoSvc)
	assetRepo := new(mockAssetRepoSvc)
	svc := NewAssetMaintenanceService(maintRepo, assetRepo)

	maintRepo.On("ListByAssetID", mock.Anything, "", (*string)(nil), (*string)(nil), 0, 10).Return([]*entity.AssetMaintenanceRecord{{ID: "maint-001"}}, int64(1), nil)

	records, total, err := svc.ListMaintenanceRecords(context.Background(), "", "", "", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, records, 1)
}

func TestAssetMaintenanceService_GetMaintenanceCosts_Success(t *testing.T) {
	maintRepo := new(mockAssetMaintRepoSvc)
	assetRepo := new(mockAssetRepoSvc)
	svc := NewAssetMaintenanceService(maintRepo, assetRepo)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	maintRepo.On("GetMaintenanceCostByAsset", mock.Anything, "asset-001", (*time.Time)(nil), (*time.Time)(nil)).Return(5000.0, nil)

	cost, err := svc.GetMaintenanceCosts(context.Background(), "asset-001", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 5000.0, cost)
}

func TestAssetDocumentService_CreateDocument_Success(t *testing.T) {
	docRepo := new(mockAssetDocRepoSvc)
	assetRepo := new(mockAssetRepoSvc)
	svc := NewAssetDocumentService(docRepo, assetRepo)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	docRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.AssetDocument")).Return(nil)

	req := &CreateDocumentRequest{
		AssetID: "asset-001", DocumentType: "manual", Title: "Test Manual",
		FilePath: "/docs/manual.pdf", UploadDate: "2024-01-01",
	}
	doc, err := svc.CreateDocument(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, doc)
}

func TestAssetDocumentService_CreateDocument_AssetNotFound(t *testing.T) {
	docRepo := new(mockAssetDocRepoSvc)
	assetRepo := new(mockAssetRepoSvc)
	svc := NewAssetDocumentService(docRepo, assetRepo)

	assetRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	req := &CreateDocumentRequest{AssetID: "nonexistent", DocumentType: "manual", Title: "Test", FilePath: "/docs/test.pdf"}
	doc, err := svc.CreateDocument(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, doc)
}

func TestAssetDocumentService_GetDocument_Success(t *testing.T) {
	docRepo := new(mockAssetDocRepoSvc)
	assetRepo := new(mockAssetRepoSvc)
	svc := NewAssetDocumentService(docRepo, assetRepo)

	docRepo.On("GetByID", mock.Anything, "doc-001").Return(&entity.AssetDocument{ID: "doc-001"}, nil)

	doc, err := svc.GetDocument(context.Background(), "doc-001")
	assert.NoError(t, err)
	assert.NotNil(t, doc)
}

func TestAssetDocumentService_UpdateDocument_Success(t *testing.T) {
	docRepo := new(mockAssetDocRepoSvc)
	assetRepo := new(mockAssetRepoSvc)
	svc := NewAssetDocumentService(docRepo, assetRepo)

	existing := &entity.AssetDocument{ID: "doc-001", AssetID: "asset-001"}
	docRepo.On("GetByID", mock.Anything, "doc-001").Return(existing, nil)
	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	docRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.AssetDocument")).Return(nil)

	req := &UpdateDocumentRequest{AssetID: "asset-001", Title: "Updated Manual"}
	doc, err := svc.UpdateDocument(context.Background(), "doc-001", req)
	assert.NoError(t, err)
	assert.NotNil(t, doc)
}

func TestAssetDocumentService_DeleteDocument_Success(t *testing.T) {
	docRepo := new(mockAssetDocRepoSvc)
	assetRepo := new(mockAssetRepoSvc)
	svc := NewAssetDocumentService(docRepo, assetRepo)

	docRepo.On("GetByID", mock.Anything, "doc-001").Return(&entity.AssetDocument{ID: "doc-001"}, nil)
	docRepo.On("Delete", mock.Anything, "doc-001").Return(nil)

	err := svc.DeleteDocument(context.Background(), "doc-001")
	assert.NoError(t, err)
}

func TestAssetDocumentService_ListDocuments_Success(t *testing.T) {
	docRepo := new(mockAssetDocRepoSvc)
	assetRepo := new(mockAssetRepoSvc)
	svc := NewAssetDocumentService(docRepo, assetRepo)

	docRepo.On("ListByAssetID", mock.Anything, "", (*string)(nil), 0, 10).Return([]*entity.AssetDocument{{ID: "doc-001"}}, int64(1), nil)

	docs, total, err := svc.ListDocuments(context.Background(), "", "", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, docs, 1)
}
