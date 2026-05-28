package service

import (
	"context"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCostEntryService_UpdateCostEntry_AlreadyApproved(t *testing.T) {
	entryRepo := new(mockCERepoSvc)
	catRepo := new(mockCCRepoSvc)
	svc := NewCostEntryService(entryRepo, catRepo)

	existing := &entity.CostEntry{ID: "entry-1", CostCategoryID: "cat-1", ApprovalStatus: "approved"}
	entryRepo.On("GetByID", mock.Anything, "entry-1").Return(existing, nil)

	entry := &entity.CostEntry{ID: "entry-1", CostCategoryID: "cat-1"}
	err := svc.UpdateCostEntry(context.Background(), entry)
	assert.Error(t, err)
}

func TestCostEntryService_UpdateCostEntry_CategoryChanged(t *testing.T) {
	entryRepo := new(mockCERepoSvc)
	catRepo := new(mockCCRepoSvc)
	svc := NewCostEntryService(entryRepo, catRepo)

	existing := &entity.CostEntry{ID: "entry-1", CostCategoryID: "cat-1", ApprovalStatus: "pending"}
	entryRepo.On("GetByID", mock.Anything, "entry-1").Return(existing, nil)
	catRepo.On("GetByID", mock.Anything, "cat-2").Return(&entity.CostCategory{ID: "cat-2"}, nil)
	entryRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.CostEntry")).Return(nil)

	entry := &entity.CostEntry{ID: "entry-1", CostCategoryID: "cat-2"}
	err := svc.UpdateCostEntry(context.Background(), entry)
	assert.NoError(t, err)
}

func TestCostEntryService_UpdateCostEntry_CategoryChanged_NotFound(t *testing.T) {
	entryRepo := new(mockCERepoSvc)
	catRepo := new(mockCCRepoSvc)
	svc := NewCostEntryService(entryRepo, catRepo)

	existing := &entity.CostEntry{ID: "entry-1", CostCategoryID: "cat-1", ApprovalStatus: "pending"}
	entryRepo.On("GetByID", mock.Anything, "entry-1").Return(existing, nil)
	catRepo.On("GetByID", mock.Anything, "cat-2").Return(nil, assert.AnError)

	entry := &entity.CostEntry{ID: "entry-1", CostCategoryID: "cat-2"}
	err := svc.UpdateCostEntry(context.Background(), entry)
	assert.Error(t, err)
}

func TestCostEntryService_DeleteCostEntry_AlreadyApproved(t *testing.T) {
	entryRepo := new(mockCERepoSvc)
	catRepo := new(mockCCRepoSvc)
	svc := NewCostEntryService(entryRepo, catRepo)

	entryRepo.On("GetByID", mock.Anything, "entry-1").Return(&entity.CostEntry{ID: "entry-1", ApprovalStatus: "approved"}, nil)

	err := svc.DeleteCostEntry(context.Background(), "entry-1")
	assert.Error(t, err)
}

func TestCostEntryService_DeleteCostEntry_NotFound(t *testing.T) {
	entryRepo := new(mockCERepoSvc)
	catRepo := new(mockCCRepoSvc)
	svc := NewCostEntryService(entryRepo, catRepo)

	entryRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	err := svc.DeleteCostEntry(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestCostEntryService_GetCostEntryByCode2(t *testing.T) {
	entryRepo := new(mockCERepoSvc)
	catRepo := new(mockCCRepoSvc)
	svc := NewCostEntryService(entryRepo, catRepo)

	entryRepo.On("GetByCode", mock.Anything, "CE001").Return(&entity.CostEntry{Code: "CE001"}, nil)

	entry, err := svc.GetCostEntryByCode(context.Background(), "CE001")
	assert.NoError(t, err)
	assert.NotNil(t, entry)
}

func TestCostEntryService_GetTotalByCategory_CategoryNotFound(t *testing.T) {
	entryRepo := new(mockCERepoSvc)
	catRepo := new(mockCCRepoSvc)
	svc := NewCostEntryService(entryRepo, catRepo)

	catRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	_, err := svc.GetTotalByCategory(context.Background(), "nonexistent", nil, nil)
	assert.Error(t, err)
}

func TestCostEntryService_ApproveCostEntry_NotPending(t *testing.T) {
	entryRepo := new(mockCERepoSvc)
	catRepo := new(mockCCRepoSvc)
	svc := NewCostEntryService(entryRepo, catRepo)

	entryRepo.On("GetByID", mock.Anything, "entry-1").Return(&entity.CostEntry{ID: "entry-1", ApprovalStatus: "approved"}, nil)

	err := svc.ApproveCostEntry(context.Background(), "entry-1", "admin-001")
	assert.Error(t, err)
}

func TestCostEntryService_ApproveCostEntry_NotFound(t *testing.T) {
	entryRepo := new(mockCERepoSvc)
	catRepo := new(mockCCRepoSvc)
	svc := NewCostEntryService(entryRepo, catRepo)

	entryRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	err := svc.ApproveCostEntry(context.Background(), "nonexistent", "admin-001")
	assert.Error(t, err)
}

func TestCostEntryService_RejectCostEntry_NotPending(t *testing.T) {
	entryRepo := new(mockCERepoSvc)
	catRepo := new(mockCCRepoSvc)
	svc := NewCostEntryService(entryRepo, catRepo)

	entryRepo.On("GetByID", mock.Anything, "entry-1").Return(&entity.CostEntry{ID: "entry-1", ApprovalStatus: "rejected"}, nil)

	err := svc.RejectCostEntry(context.Background(), "entry-1", "admin-001")
	assert.Error(t, err)
}

func TestCostEntryService_RejectCostEntry_NotFound(t *testing.T) {
	entryRepo := new(mockCERepoSvc)
	catRepo := new(mockCCRepoSvc)
	svc := NewCostEntryService(entryRepo, catRepo)

	entryRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	err := svc.RejectCostEntry(context.Background(), "nonexistent", "admin-001")
	assert.Error(t, err)
}

func TestCostCategoryService_CreateCostCategory_WithParent2(t *testing.T) {
	repo := new(mockCCRepoSvc)
	svc := NewCostCategoryService(repo)

	parentID := "parent-1"
	repo.On("GetByCode", mock.Anything, "CAT002").Return(nil, assert.AnError)
	repo.On("GetByID", mock.Anything, "parent-1").Return(&entity.CostCategory{ID: "parent-1"}, nil)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.CostCategory")).Return(nil)

	err := svc.CreateCostCategory(context.Background(), &entity.CostCategory{Code: "CAT002", ParentID: &parentID})
	assert.NoError(t, err)
}

func TestCostCategoryService_CreateCostCategory_ParentNotFound2(t *testing.T) {
	repo := new(mockCCRepoSvc)
	svc := NewCostCategoryService(repo)

	parentID := "nonexistent"
	repo.On("GetByCode", mock.Anything, "CAT003").Return(nil, assert.AnError)
	repo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	err := svc.CreateCostCategory(context.Background(), &entity.CostCategory{Code: "CAT003", ParentID: &parentID})
	assert.Error(t, err)
}

func TestCostCategoryService_UpdateCostCategory_CodeConflict2(t *testing.T) {
	repo := new(mockCCRepoSvc)
	svc := NewCostCategoryService(repo)

	existing := &entity.CostCategory{ID: "cat-1", Code: "CAT001"}
	repo.On("GetByID", mock.Anything, "cat-1").Return(existing, nil)
	repo.On("GetByCode", mock.Anything, "CAT002").Return(&entity.CostCategory{ID: "cat-2", Code: "CAT002"}, nil)

	err := svc.UpdateCostCategory(context.Background(), &entity.CostCategory{ID: "cat-1", Code: "CAT002"})
	assert.Error(t, err)
}

func TestCostCategoryService_UpdateCostCategory_ParentNotFound3(t *testing.T) {
	repo := new(mockCCRepoSvc)
	svc := NewCostCategoryService(repo)

	existing := &entity.CostCategory{ID: "cat-1", Code: "CAT001"}
	parentID := "nonexistent"
	repo.On("GetByID", mock.Anything, "cat-1").Return(existing, nil)
	repo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	err := svc.UpdateCostCategory(context.Background(), &entity.CostCategory{ID: "cat-1", Code: "CAT001", ParentID: &parentID})
	assert.Error(t, err)
}

func TestCostCategoryService_UpdateCostCategory_SelfReference2(t *testing.T) {
	repo := new(mockCCRepoSvc)
	svc := NewCostCategoryService(repo)

	existing := &entity.CostCategory{ID: "cat-1", Code: "CAT001"}
	repo.On("GetByID", mock.Anything, "cat-1").Return(existing, nil)

	err := svc.UpdateCostCategory(context.Background(), &entity.CostCategory{ID: "cat-1", Code: "CAT001", ParentID: strPtr("cat-1")})
	assert.Error(t, err)
}

func TestCostCategoryService_DeleteCostCategory_HasChildren2(t *testing.T) {
	repo := new(mockCCRepoSvc)
	svc := NewCostCategoryService(repo)

	repo.On("GetByID", mock.Anything, "cat-1").Return(&entity.CostCategory{ID: "cat-1"}, nil)
	repo.On("List", mock.Anything, mock.AnythingOfType("*string"), (*string)(nil)).Return([]*entity.CostCategory{{ID: "child-1"}}, nil)

	err := svc.DeleteCostCategory(context.Background(), "cat-1")
	assert.Error(t, err)
}

func TestCostCategoryService_DeleteCostCategory_Success3(t *testing.T) {
	repo := new(mockCCRepoSvc)
	svc := NewCostCategoryService(repo)

	repo.On("GetByID", mock.Anything, "cat-1").Return(&entity.CostCategory{ID: "cat-1"}, nil)
	repo.On("List", mock.Anything, mock.AnythingOfType("*string"), (*string)(nil)).Return([]*entity.CostCategory{}, nil)
	repo.On("Delete", mock.Anything, "cat-1").Return(nil)

	err := svc.DeleteCostCategory(context.Background(), "cat-1")
	assert.NoError(t, err)
}

func TestAssetService_UpdateAsset_WithAllFields2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	depRepo := new(mockAssetDepRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetService(assetRepo, maintRepo, depRepo, docRepo)

	existing := &entity.Asset{ID: "asset-001", Code: "ASSET001"}
	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(existing, nil)
	assetRepo.On("GetByCode", mock.Anything, "ASSET002").Return(nil, assert.AnError)
	assetRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.Asset")).Return(nil)

	req := &UpdateAssetRequest{
		Code: "ASSET002", Name: "Updated", Category: "solar",
		AssetType: "panel", Manufacturer: "Mfg", Model: "M1",
		SerialNumber: "SN123", PurchasePrice: 50000, ExpectedLife: 15,
		ResidualValue: 5000, Location: "Beijing", Status: "active",
		Description: "Updated asset", PurchaseDate: "2024-06-01",
	}
	asset, err := svc.UpdateAsset(context.Background(), "asset-001", req)
	require.NoError(t, err)
	assert.Equal(t, "Updated", asset.Name)
}

func TestAssetService_UpdateAsset_CodeConflict2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	depRepo := new(mockAssetDepRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetService(assetRepo, maintRepo, depRepo, docRepo)

	existing := &entity.Asset{ID: "asset-001", Code: "ASSET001"}
	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(existing, nil)
	assetRepo.On("GetByCode", mock.Anything, "ASSET002").Return(&entity.Asset{ID: "asset-002", Code: "ASSET002"}, nil)

	req := &UpdateAssetRequest{Code: "ASSET002"}
	asset, err := svc.UpdateAsset(context.Background(), "asset-001", req)
	assert.Error(t, err)
	assert.Nil(t, asset)
}

func TestAssetService_DeleteAsset_HasDepreciationRecords2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	depRepo := new(mockAssetDepRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetService(assetRepo, maintRepo, depRepo, docRepo)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	maintRepo.On("ListByAssetID", mock.Anything, "asset-001", (*string)(nil), (*string)(nil), 0, 1).Return([]*entity.AssetMaintenanceRecord{}, int64(0), nil)
	depRepo.On("ListByAssetID", mock.Anything, "asset-001", (*string)(nil), 0, 1).Return([]*entity.AssetDepreciationRecord{{ID: "dep-1"}}, int64(1), nil)

	err := svc.DeleteAsset(context.Background(), "asset-001")
	assert.Error(t, err)
}

func TestAssetService_DeleteAsset_HasDocuments2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	depRepo := new(mockAssetDepRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetService(assetRepo, maintRepo, depRepo, docRepo)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	maintRepo.On("ListByAssetID", mock.Anything, "asset-001", (*string)(nil), (*string)(nil), 0, 1).Return([]*entity.AssetMaintenanceRecord{}, int64(0), nil)
	depRepo.On("ListByAssetID", mock.Anything, "asset-001", (*string)(nil), 0, 1).Return([]*entity.AssetDepreciationRecord{}, int64(0), nil)
	docRepo.On("ListByAssetID", mock.Anything, "asset-001", (*string)(nil), 0, 1).Return([]*entity.AssetDocument{{ID: "doc-1"}}, int64(1), nil)

	err := svc.DeleteAsset(context.Background(), "asset-001")
	assert.Error(t, err)
}

func TestAssetService_CalculateDepreciation_WithExistingRecord2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	depRepo := new(mockAssetDepRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetService(assetRepo, maintRepo, depRepo, docRepo)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001", Cost: 100000, UsefulLife: 10, SalvageValue: 10000}, nil)
	depRepo.On("GetLatestByAssetID", mock.Anything, "asset-001").Return(&entity.AssetDepreciationRecord{BookValue: 55000, AccumulatedDepreciation: 45000}, nil)

	result, err := svc.CalculateDepreciation(context.Background(), "asset-001", "straight-line")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestAssetService_CalculateDepreciation_AssetNotDepreciable2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	depRepo := new(mockAssetDepRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetService(assetRepo, maintRepo, depRepo, docRepo)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001", UsefulLife: 0}, nil)

	result, err := svc.CalculateDepreciation(context.Background(), "asset-001", "straight-line")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAssetService_CalculateDepreciation_DecliningBalance_FullyDepreciated2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	depRepo := new(mockAssetDepRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetService(assetRepo, maintRepo, depRepo, docRepo)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001", Cost: 100000, UsefulLife: 10, SalvageValue: 10000}, nil)
	depRepo.On("GetLatestByAssetID", mock.Anything, "asset-001").Return(&entity.AssetDepreciationRecord{BookValue: 8000, AccumulatedDepreciation: 92000}, nil)

	result, err := svc.CalculateDepreciation(context.Background(), "asset-001", "declining-balance")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAssetService_CreateAsset_DefaultCode2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	depRepo := new(mockAssetDepRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetService(assetRepo, maintRepo, depRepo, docRepo)

	assetRepo.On("GetByCode", mock.Anything, "").Return(nil, assert.AnError)
	assetRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Asset")).Return(nil)

	req := &CreateAssetRequest{
		Name: "Test Asset", Category: "equipment",
		PurchasePrice: 100000, PurchaseDate: "2024-01-01", ExpectedLife: 10,
	}
	asset, err := svc.CreateAsset(context.Background(), req)
	require.NoError(t, err)
	assert.NotEmpty(t, asset.Code)
}

func TestAssetService_ListAssets_WithFilters2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	depRepo := new(mockAssetDepRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetService(assetRepo, maintRepo, depRepo, docRepo)

	status := "active"
	category := "solar"
	assetRepo.On("List", mock.Anything, (*string)(nil), &status, &category, 0, 10).Return([]*entity.Asset{{ID: "a1"}}, int64(1), nil)

	assets, total, err := svc.ListAssets(context.Background(), "", "solar", "active", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, assets, 1)
}


