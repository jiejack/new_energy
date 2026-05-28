package service

import (
	"context"
	"testing"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAssetDepreciationService_GetDepreciationSummary_NoDate2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	depRepo := new(mockAssetDepRepoSvc)
	svc := NewAssetDepreciationService(depRepo, assetRepo)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	depRepo.On("GetDepreciationSummaryByPeriod", mock.Anything, "annual", (*time.Time)(nil), (*time.Time)(nil)).Return(3000.0, nil)

	result, err := svc.GetDepreciationSummary(context.Background(), "asset-001", "")
	assert.NoError(t, err)
	assert.Equal(t, 3000.0, result)
}

func TestAssetDocumentService_CreateDocument_WithAssetID2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetDocumentService(docRepo, assetRepo)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	docRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.AssetDocument")).Return(nil)

	req := &CreateDocumentRequest{AssetID: "asset-001", DocumentType: "manual", Title: "Test Doc", Description: "desc"}
	doc, err := svc.CreateDocument(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, doc)
}

func TestAssetDocumentService_CreateDocument_AssetNotFound2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetDocumentService(docRepo, assetRepo)

	assetRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	req := &CreateDocumentRequest{AssetID: "nonexistent", DocumentType: "manual", Title: "Test Doc"}
	doc, err := svc.CreateDocument(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, doc)
}

func TestAssetDocumentService_DeleteDocument_NotFound2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetDocumentService(docRepo, assetRepo)

	docRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	err := svc.DeleteDocument(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestAssetMaintenanceService_CreateMaintenanceRecord_WithAssetID2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	svc := NewAssetMaintenanceService(maintRepo, assetRepo)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	maintRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.AssetMaintenanceRecord")).Return(nil)

	req := &CreateMaintenanceRequest{AssetID: "asset-001", Description: "Test maintenance", Status: "scheduled"}
	record, err := svc.CreateMaintenanceRecord(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, record)
}

func TestAssetMaintenanceService_CreateMaintenanceRecord_AssetNotFound2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	svc := NewAssetMaintenanceService(maintRepo, assetRepo)

	assetRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	req := &CreateMaintenanceRequest{AssetID: "nonexistent", Description: "Test maintenance"}
	record, err := svc.CreateMaintenanceRecord(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, record)
}

func TestAssetMaintenanceService_DeleteMaintenanceRecord_NotFound2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	svc := NewAssetMaintenanceService(maintRepo, assetRepo)

	maintRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	err := svc.DeleteMaintenanceRecord(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestAssetMaintenanceService_GetMaintenanceCosts_WithDates2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	svc := NewAssetMaintenanceService(maintRepo, assetRepo)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	maintRepo.On("GetMaintenanceCostByAsset", mock.Anything, "asset-001", mock.Anything, mock.Anything).Return(5000.0, nil)

	result, err := svc.GetMaintenanceCosts(context.Background(), "asset-001", "2024-01-01", "2024-12-31")
	assert.NoError(t, err)
	assert.Equal(t, 5000.0, result)
}

func TestAssetMaintenanceService_GetMaintenanceCosts_AssetNotFound2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	svc := NewAssetMaintenanceService(maintRepo, assetRepo)

	assetRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	_, err := svc.GetMaintenanceCosts(context.Background(), "nonexistent", "", "")
	assert.Error(t, err)
}

func TestAssetMaintenanceService_GetMaintenanceCosts_InvalidDates2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	svc := NewAssetMaintenanceService(maintRepo, assetRepo)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	maintRepo.On("GetMaintenanceCostByAsset", mock.Anything, "asset-001", (*time.Time)(nil), (*time.Time)(nil)).Return(3000.0, nil)

	result, err := svc.GetMaintenanceCosts(context.Background(), "asset-001", "invalid", "invalid")
	assert.NoError(t, err)
	assert.Equal(t, 3000.0, result)
}

func TestAssetMaintenanceService_UpdateMaintenanceRecord_WithAssetID2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	svc := NewAssetMaintenanceService(maintRepo, assetRepo)

	existing := &entity.AssetMaintenanceRecord{ID: "maint-1", AssetID: "asset-001"}
	maintRepo.On("GetByID", mock.Anything, "maint-1").Return(existing, nil)
	assetRepo.On("GetByID", mock.Anything, "asset-002").Return(&entity.Asset{ID: "asset-002"}, nil)
	maintRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.AssetMaintenanceRecord")).Return(nil)

	req := &UpdateMaintenanceRequest{AssetID: "asset-002", Description: "Updated", Status: "completed"}
	record, err := svc.UpdateMaintenanceRecord(context.Background(), "maint-1", req)
	assert.NoError(t, err)
	assert.NotNil(t, record)
}

func TestAssetMaintenanceService_UpdateMaintenanceRecord_AssetNotFound2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	svc := NewAssetMaintenanceService(maintRepo, assetRepo)

	existing := &entity.AssetMaintenanceRecord{ID: "maint-1", AssetID: "asset-001"}
	maintRepo.On("GetByID", mock.Anything, "maint-1").Return(existing, nil)
	assetRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	req := &UpdateMaintenanceRequest{AssetID: "nonexistent"}
	record, err := svc.UpdateMaintenanceRecord(context.Background(), "maint-1", req)
	assert.Error(t, err)
	assert.Nil(t, record)
}

func TestAssetMaintenanceService_UpdateMaintenanceRecord_NotFound2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	maintRepo := new(mockAssetMaintRepoSvc)
	svc := NewAssetMaintenanceService(maintRepo, assetRepo)

	maintRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	req := &UpdateMaintenanceRequest{Description: "Test"}
	record, err := svc.UpdateMaintenanceRecord(context.Background(), "nonexistent", req)
	assert.Error(t, err)
	assert.Nil(t, record)
}

func TestAssetDocumentService_UpdateDocument_WithAssetID2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetDocumentService(docRepo, assetRepo)

	existing := &entity.AssetDocument{ID: "doc-1", AssetID: "asset-001"}
	docRepo.On("GetByID", mock.Anything, "doc-1").Return(existing, nil)
	assetRepo.On("GetByID", mock.Anything, "asset-002").Return(&entity.Asset{ID: "asset-002"}, nil)
	docRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.AssetDocument")).Return(nil)

	req := &UpdateDocumentRequest{AssetID: "asset-002", DocumentType: "manual", Title: "Updated Doc", Description: "Updated desc"}
	doc, err := svc.UpdateDocument(context.Background(), "doc-1", req)
	assert.NoError(t, err)
	assert.NotNil(t, doc)
}

func TestAssetDocumentService_UpdateDocument_AssetNotFound2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetDocumentService(docRepo, assetRepo)

	existing := &entity.AssetDocument{ID: "doc-1", AssetID: "asset-001"}
	docRepo.On("GetByID", mock.Anything, "doc-1").Return(existing, nil)
	assetRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	req := &UpdateDocumentRequest{AssetID: "nonexistent"}
	doc, err := svc.UpdateDocument(context.Background(), "doc-1", req)
	assert.Error(t, err)
	assert.Nil(t, doc)
}

func TestAssetDocumentService_UpdateDocument_NotFound2(t *testing.T) {
	assetRepo := new(mockAssetRepoSvc)
	docRepo := new(mockAssetDocRepoSvc)
	svc := NewAssetDocumentService(docRepo, assetRepo)

	docRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	req := &UpdateDocumentRequest{Title: "Test"}
	doc, err := svc.UpdateDocument(context.Background(), "nonexistent", req)
	assert.Error(t, err)
	assert.Nil(t, doc)
}
