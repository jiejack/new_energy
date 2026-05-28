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

type mockCCRepoSvc struct {
	mock.Mock
}

func (m *mockCCRepoSvc) Create(ctx context.Context, category *entity.CostCategory) error {
	return m.Called(ctx, category).Error(0)
}

func (m *mockCCRepoSvc) Update(ctx context.Context, category *entity.CostCategory) error {
	return m.Called(ctx, category).Error(0)
}

func (m *mockCCRepoSvc) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockCCRepoSvc) GetByID(ctx context.Context, id string) (*entity.CostCategory, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostCategory), args.Error(1)
}

func (m *mockCCRepoSvc) GetByCode(ctx context.Context, code string) (*entity.CostCategory, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostCategory), args.Error(1)
}

func (m *mockCCRepoSvc) List(ctx context.Context, parentID *string, status *string) ([]*entity.CostCategory, error) {
	args := m.Called(ctx, parentID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.CostCategory), args.Error(1)
}

func (m *mockCCRepoSvc) GetTree(ctx context.Context) ([]*entity.CostCategory, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.CostCategory), args.Error(1)
}

type mockCERepoSvc struct {
	mock.Mock
}

func (m *mockCERepoSvc) Create(ctx context.Context, entry *entity.CostEntry) error {
	return m.Called(ctx, entry).Error(0)
}

func (m *mockCERepoSvc) Update(ctx context.Context, entry *entity.CostEntry) error {
	return m.Called(ctx, entry).Error(0)
}

func (m *mockCERepoSvc) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockCERepoSvc) GetByID(ctx context.Context, id string) (*entity.CostEntry, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostEntry), args.Error(1)
}

func (m *mockCERepoSvc) GetByCode(ctx context.Context, code string) (*entity.CostEntry, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostEntry), args.Error(1)
}

func (m *mockCERepoSvc) List(ctx context.Context, categoryID *string, startDate, endDate *time.Time, status *string, offset, limit int) ([]*entity.CostEntry, int64, error) {
	args := m.Called(ctx, categoryID, startDate, endDate, status, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.CostEntry), args.Get(1).(int64), args.Error(2)
}

func (m *mockCERepoSvc) GetTotalByCategory(ctx context.Context, categoryID string, startDate, endDate *time.Time) (float64, error) {
	args := m.Called(ctx, categoryID, startDate, endDate)
	return args.Get(0).(float64), args.Error(1)
}

func (m *mockCERepoSvc) GetTotalByPeriod(ctx context.Context, startDate, endDate *time.Time) (float64, error) {
	args := m.Called(ctx, startDate, endDate)
	return args.Get(0).(float64), args.Error(1)
}

type mockCARepoSvc struct {
	mock.Mock
}

func (m *mockCARepoSvc) Create(ctx context.Context, allocation *entity.CostAllocation) error {
	return m.Called(ctx, allocation).Error(0)
}

func (m *mockCARepoSvc) Update(ctx context.Context, allocation *entity.CostAllocation) error {
	return m.Called(ctx, allocation).Error(0)
}

func (m *mockCARepoSvc) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockCARepoSvc) GetByID(ctx context.Context, id string) (*entity.CostAllocation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostAllocation), args.Error(1)
}

func (m *mockCARepoSvc) ListByCostEntryID(ctx context.Context, costEntryID string) ([]*entity.CostAllocation, error) {
	args := m.Called(ctx, costEntryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.CostAllocation), args.Error(1)
}

func (m *mockCARepoSvc) ListByAllocated(ctx context.Context, allocatedTo, allocatedID string) ([]*entity.CostAllocation, error) {
	args := m.Called(ctx, allocatedTo, allocatedID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.CostAllocation), args.Error(1)
}

func (m *mockCARepoSvc) GetTotalByAllocated(ctx context.Context, allocatedTo, allocatedID string, startDate, endDate *time.Time) (float64, error) {
	args := m.Called(ctx, allocatedTo, allocatedID, startDate, endDate)
	return args.Get(0).(float64), args.Error(1)
}

type mockCRRepoSvc struct {
	mock.Mock
}

func (m *mockCRRepoSvc) Create(ctx context.Context, report *entity.CostReport) error {
	return m.Called(ctx, report).Error(0)
}

func (m *mockCRRepoSvc) Update(ctx context.Context, report *entity.CostReport) error {
	return m.Called(ctx, report).Error(0)
}

func (m *mockCRRepoSvc) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockCRRepoSvc) GetByID(ctx context.Context, id string) (*entity.CostReport, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostReport), args.Error(1)
}

func (m *mockCRRepoSvc) GetByCode(ctx context.Context, code string) (*entity.CostReport, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostReport), args.Error(1)
}

func (m *mockCRRepoSvc) List(ctx context.Context, reportType *string, status *string, startDate, endDate *time.Time, offset, limit int) ([]*entity.CostReport, int64, error) {
	args := m.Called(ctx, reportType, status, startDate, endDate, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.CostReport), args.Get(1).(int64), args.Error(2)
}

func (m *mockCRRepoSvc) GetByPeriod(ctx context.Context, reportType string, periodStart, periodEnd time.Time) (*entity.CostReport, error) {
	args := m.Called(ctx, reportType, periodStart, periodEnd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostReport), args.Error(1)
}

var _ repository.CostCategoryRepository = (*mockCCRepoSvc)(nil)
var _ repository.CostEntryRepository = (*mockCERepoSvc)(nil)
var _ repository.CostAllocationRepository = (*mockCARepoSvc)(nil)
var _ repository.CostReportRepository = (*mockCRRepoSvc)(nil)

func TestCostCategoryService_CreateCostCategory_Success(t *testing.T) {
	repo := new(mockCCRepoSvc)
	svc := NewCostCategoryService(repo)

	repo.On("GetByCode", mock.Anything, "CC001").Return(nil, assert.AnError)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.CostCategory")).Return(nil)

	cat := &entity.CostCategory{Code: "CC001", Name: "Test Category", Type: "direct", Status: "active"}
	err := svc.CreateCostCategory(context.Background(), cat)
	assert.NoError(t, err)
}

func TestCostCategoryService_CreateCostCategory_DuplicateCode(t *testing.T) {
	repo := new(mockCCRepoSvc)
	svc := NewCostCategoryService(repo)

	repo.On("GetByCode", mock.Anything, "CC001").Return(&entity.CostCategory{Code: "CC001"}, nil)

	cat := &entity.CostCategory{Code: "CC001", Name: "Test Category", Type: "direct", Status: "active"}
	err := svc.CreateCostCategory(context.Background(), cat)
	assert.Error(t, err)
}

func TestCostCategoryService_GetCostCategoryByID_Success(t *testing.T) {
	repo := new(mockCCRepoSvc)
	svc := NewCostCategoryService(repo)

	repo.On("GetByID", mock.Anything, "cc-001").Return(&entity.CostCategory{ID: "cc-001"}, nil)

	cat, err := svc.GetCostCategoryByID(context.Background(), "cc-001")
	assert.NoError(t, err)
	assert.NotNil(t, cat)
}

func TestCostCategoryService_UpdateCostCategory_Success(t *testing.T) {
	repo := new(mockCCRepoSvc)
	svc := NewCostCategoryService(repo)

	existing := &entity.CostCategory{ID: "cc-001", Code: "CC001"}
	repo.On("GetByID", mock.Anything, "cc-001").Return(existing, nil)
	repo.On("GetByCode", mock.Anything, "CC001").Return(existing, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*entity.CostCategory")).Return(nil)

	cat := &entity.CostCategory{ID: "cc-001", Code: "CC001", Name: "Updated Category"}
	err := svc.UpdateCostCategory(context.Background(), cat)
	assert.NoError(t, err)
}

func TestCostCategoryService_DeleteCostCategory_Success(t *testing.T) {
	repo := new(mockCCRepoSvc)
	svc := NewCostCategoryService(repo)

	repo.On("GetByID", mock.Anything, "cc-001").Return(&entity.CostCategory{ID: "cc-001"}, nil)
	repo.On("List", mock.Anything, mock.AnythingOfType("*string"), (*string)(nil)).Return([]*entity.CostCategory{}, nil)
	repo.On("Delete", mock.Anything, "cc-001").Return(nil)

	err := svc.DeleteCostCategory(context.Background(), "cc-001")
	assert.NoError(t, err)
}

func TestCostCategoryService_ListCostCategories_Success(t *testing.T) {
	repo := new(mockCCRepoSvc)
	svc := NewCostCategoryService(repo)

	repo.On("List", mock.Anything, (*string)(nil), (*string)(nil)).Return([]*entity.CostCategory{{ID: "cc-001"}}, nil)

	cats, err := svc.ListCostCategories(context.Background(), nil, nil)
	assert.NoError(t, err)
	assert.Len(t, cats, 1)
}

func TestCostCategoryService_GetCostCategoryTree_Success(t *testing.T) {
	repo := new(mockCCRepoSvc)
	svc := NewCostCategoryService(repo)

	repo.On("GetTree", mock.Anything).Return([]*entity.CostCategory{{ID: "cc-001"}}, nil)

	cats, err := svc.GetCostCategoryTree(context.Background())
	assert.NoError(t, err)
	assert.Len(t, cats, 1)
}

func TestCostEntryService_CreateCostEntry_Success(t *testing.T) {
	ceRepo := new(mockCERepoSvc)
	ccRepo := new(mockCCRepoSvc)
	svc := NewCostEntryService(ceRepo, ccRepo)

	ccRepo.On("GetByID", mock.Anything, "cc-001").Return(&entity.CostCategory{ID: "cc-001"}, nil)
	ceRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.CostEntry")).Return(nil)

	entry := &entity.CostEntry{CostCategoryID: "cc-001", Amount: 1000, ApprovalStatus: "pending"}
	err := svc.CreateCostEntry(context.Background(), entry)
	assert.NoError(t, err)
}

func TestCostEntryService_CreateCostEntry_CategoryNotFound(t *testing.T) {
	ceRepo := new(mockCERepoSvc)
	ccRepo := new(mockCCRepoSvc)
	svc := NewCostEntryService(ceRepo, ccRepo)

	ccRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	entry := &entity.CostEntry{CostCategoryID: "nonexistent", Amount: 1000}
	err := svc.CreateCostEntry(context.Background(), entry)
	assert.Error(t, err)
}

func TestCostEntryService_GetCostEntryByID_Success(t *testing.T) {
	ceRepo := new(mockCERepoSvc)
	ccRepo := new(mockCCRepoSvc)
	svc := NewCostEntryService(ceRepo, ccRepo)

	ceRepo.On("GetByID", mock.Anything, "ce-001").Return(&entity.CostEntry{ID: "ce-001"}, nil)

	entry, err := svc.GetCostEntryByID(context.Background(), "ce-001")
	assert.NoError(t, err)
	assert.NotNil(t, entry)
}

func TestCostEntryService_UpdateCostEntry_Success(t *testing.T) {
	ceRepo := new(mockCERepoSvc)
	ccRepo := new(mockCCRepoSvc)
	svc := NewCostEntryService(ceRepo, ccRepo)

	existing := &entity.CostEntry{ID: "ce-001", CostCategoryID: "cc-001", ApprovalStatus: "pending"}
	ceRepo.On("GetByID", mock.Anything, "ce-001").Return(existing, nil)
	ccRepo.On("GetByID", mock.Anything, "cc-001").Return(&entity.CostCategory{ID: "cc-001"}, nil)
	ceRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.CostEntry")).Return(nil)

	entry := &entity.CostEntry{ID: "ce-001", CostCategoryID: "cc-001", Amount: 2000}
	err := svc.UpdateCostEntry(context.Background(), entry)
	assert.NoError(t, err)
}

func TestCostEntryService_DeleteCostEntry_Success(t *testing.T) {
	ceRepo := new(mockCERepoSvc)
	ccRepo := new(mockCCRepoSvc)
	svc := NewCostEntryService(ceRepo, ccRepo)

	ceRepo.On("GetByID", mock.Anything, "ce-001").Return(&entity.CostEntry{ID: "ce-001", ApprovalStatus: "pending"}, nil)
	ceRepo.On("Delete", mock.Anything, "ce-001").Return(nil)

	err := svc.DeleteCostEntry(context.Background(), "ce-001")
	assert.NoError(t, err)
}

func TestCostEntryService_ListCostEntries_Success(t *testing.T) {
	ceRepo := new(mockCERepoSvc)
	ccRepo := new(mockCCRepoSvc)
	svc := NewCostEntryService(ceRepo, ccRepo)

	ceRepo.On("List", mock.Anything, (*string)(nil), (*time.Time)(nil), (*time.Time)(nil), (*string)(nil), 0, 10).Return([]*entity.CostEntry{{ID: "ce-001"}}, int64(1), nil)

	entries, total, err := svc.ListCostEntries(context.Background(), nil, nil, nil, nil, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, entries, 1)
}

func TestCostEntryService_ApproveCostEntry_Success(t *testing.T) {
	ceRepo := new(mockCERepoSvc)
	ccRepo := new(mockCCRepoSvc)
	svc := NewCostEntryService(ceRepo, ccRepo)

	ceRepo.On("GetByID", mock.Anything, "ce-001").Return(&entity.CostEntry{ID: "ce-001", ApprovalStatus: "pending"}, nil)
	ceRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.CostEntry")).Return(nil)

	err := svc.ApproveCostEntry(context.Background(), "ce-001", "admin")
	assert.NoError(t, err)
}

func TestCostEntryService_RejectCostEntry_Success(t *testing.T) {
	ceRepo := new(mockCERepoSvc)
	ccRepo := new(mockCCRepoSvc)
	svc := NewCostEntryService(ceRepo, ccRepo)

	ceRepo.On("GetByID", mock.Anything, "ce-001").Return(&entity.CostEntry{ID: "ce-001", ApprovalStatus: "pending"}, nil)
	ceRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.CostEntry")).Return(nil)

	err := svc.RejectCostEntry(context.Background(), "ce-001", "admin")
	assert.NoError(t, err)
}

func TestCostEntryService_GetTotalByCategory_Success(t *testing.T) {
	ceRepo := new(mockCERepoSvc)
	ccRepo := new(mockCCRepoSvc)
	svc := NewCostEntryService(ceRepo, ccRepo)

	ccRepo.On("GetByID", mock.Anything, "cc-001").Return(&entity.CostCategory{ID: "cc-001"}, nil)
	ceRepo.On("GetTotalByCategory", mock.Anything, "cc-001", (*time.Time)(nil), (*time.Time)(nil)).Return(1000.0, nil)

	total, err := svc.GetTotalByCategory(context.Background(), "cc-001", nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1000.0, total)
}

func TestCostEntryService_GetTotalByPeriod_Success(t *testing.T) {
	ceRepo := new(mockCERepoSvc)
	ccRepo := new(mockCCRepoSvc)
	svc := NewCostEntryService(ceRepo, ccRepo)

	ceRepo.On("GetTotalByPeriod", mock.Anything, (*time.Time)(nil), (*time.Time)(nil)).Return(5000.0, nil)

	total, err := svc.GetTotalByPeriod(context.Background(), nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, 5000.0, total)
}

func TestCostAllocationService_CreateCostAllocation_Success(t *testing.T) {
	caRepo := new(mockCARepoSvc)
	ceRepo := new(mockCERepoSvc)
	svc := NewCostAllocationService(caRepo, ceRepo)

	ceRepo.On("GetByID", mock.Anything, "ce-001").Return(&entity.CostEntry{ID: "ce-001", ApprovalStatus: "approved", Amount: 10000}, nil)
	caRepo.On("ListByCostEntryID", mock.Anything, "ce-001").Return([]*entity.CostAllocation{}, nil)
	caRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.CostAllocation")).Return(nil)

	alloc := &entity.CostAllocation{CostEntryID: "ce-001", AllocatedTo: "project", AllocatedID: "proj-001", Amount: 5000, Percentage: 50}
	err := svc.CreateCostAllocation(context.Background(), alloc)
	assert.NoError(t, err)
}

func TestCostAllocationService_GetCostAllocationByID_Success(t *testing.T) {
	caRepo := new(mockCARepoSvc)
	ceRepo := new(mockCERepoSvc)
	svc := NewCostAllocationService(caRepo, ceRepo)

	caRepo.On("GetByID", mock.Anything, "ca-001").Return(&entity.CostAllocation{ID: "ca-001"}, nil)

	alloc, err := svc.GetCostAllocationByID(context.Background(), "ca-001")
	assert.NoError(t, err)
	assert.NotNil(t, alloc)
}

func TestCostAllocationService_DeleteCostAllocation_Success(t *testing.T) {
	caRepo := new(mockCARepoSvc)
	ceRepo := new(mockCERepoSvc)
	svc := NewCostAllocationService(caRepo, ceRepo)

	caRepo.On("GetByID", mock.Anything, "ca-001").Return(&entity.CostAllocation{ID: "ca-001"}, nil)
	caRepo.On("Delete", mock.Anything, "ca-001").Return(nil)

	err := svc.DeleteCostAllocation(context.Background(), "ca-001")
	assert.NoError(t, err)
}

func TestCostAllocationService_ListCostAllocationsByCostEntryID_Success(t *testing.T) {
	caRepo := new(mockCARepoSvc)
	ceRepo := new(mockCERepoSvc)
	svc := NewCostAllocationService(caRepo, ceRepo)

	ceRepo.On("GetByID", mock.Anything, "ce-001").Return(&entity.CostEntry{ID: "ce-001"}, nil)
	caRepo.On("ListByCostEntryID", mock.Anything, "ce-001").Return([]*entity.CostAllocation{{ID: "ca-001"}}, nil)

	allocs, err := svc.ListCostAllocationsByCostEntryID(context.Background(), "ce-001")
	assert.NoError(t, err)
	assert.Len(t, allocs, 1)
}

func TestCostAllocationService_GetTotalByAllocated_Success(t *testing.T) {
	caRepo := new(mockCARepoSvc)
	ceRepo := new(mockCERepoSvc)
	svc := NewCostAllocationService(caRepo, ceRepo)

	caRepo.On("GetTotalByAllocated", mock.Anything, "project", "proj-001", (*time.Time)(nil), (*time.Time)(nil)).Return(5000.0, nil)

	total, err := svc.GetTotalByAllocated(context.Background(), "project", "proj-001", nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, 5000.0, total)
}

func TestCostReportService_CreateCostReport_Success(t *testing.T) {
	crRepo := new(mockCRRepoSvc)
	ceRepo := new(mockCERepoSvc)
	svc := NewCostReportService(crRepo, ceRepo)

	crRepo.On("GetByPeriod", mock.Anything, "monthly", mock.Anything, mock.Anything).Return(nil, assert.AnError)
	crRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.CostReport")).Return(nil)

	report := &entity.CostReport{Name: "Monthly Report", ReportType: "monthly", Status: "draft"}
	err := svc.CreateCostReport(context.Background(), report)
	assert.NoError(t, err)
}

func TestCostReportService_GetCostReportByID_Success(t *testing.T) {
	crRepo := new(mockCRRepoSvc)
	ceRepo := new(mockCERepoSvc)
	svc := NewCostReportService(crRepo, ceRepo)

	crRepo.On("GetByID", mock.Anything, "cr-001").Return(&entity.CostReport{ID: "cr-001"}, nil)

	report, err := svc.GetCostReportByID(context.Background(), "cr-001")
	assert.NoError(t, err)
	assert.NotNil(t, report)
}

func TestCostReportService_DeleteCostReport_Success(t *testing.T) {
	crRepo := new(mockCRRepoSvc)
	ceRepo := new(mockCERepoSvc)
	svc := NewCostReportService(crRepo, ceRepo)

	crRepo.On("GetByID", mock.Anything, "cr-001").Return(&entity.CostReport{ID: "cr-001", Status: "draft"}, nil)
	crRepo.On("Delete", mock.Anything, "cr-001").Return(nil)

	err := svc.DeleteCostReport(context.Background(), "cr-001")
	assert.NoError(t, err)
}

func TestCostReportService_ListCostReports_Success(t *testing.T) {
	crRepo := new(mockCRRepoSvc)
	ceRepo := new(mockCERepoSvc)
	svc := NewCostReportService(crRepo, ceRepo)

	crRepo.On("List", mock.Anything, (*string)(nil), (*string)(nil), (*time.Time)(nil), (*time.Time)(nil), 0, 10).Return([]*entity.CostReport{{ID: "cr-001"}}, int64(1), nil)

	reports, total, err := svc.ListCostReports(context.Background(), nil, nil, nil, nil, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, reports, 1)
}

func TestCostReportService_GenerateCostReport_Success(t *testing.T) {
	crRepo := new(mockCRRepoSvc)
	ceRepo := new(mockCERepoSvc)
	svc := NewCostReportService(crRepo, ceRepo)

	now := time.Now()
	existing := &entity.CostReport{ID: "cr-001", Status: "draft", PeriodStart: now, PeriodEnd: now.AddDate(0, 1, 0)}
	crRepo.On("GetByID", mock.Anything, "cr-001").Return(existing, nil)
	ceRepo.On("GetTotalByPeriod", mock.Anything, mock.Anything, mock.Anything).Return(5000.0, nil)
	crRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.CostReport")).Return(nil)

	err := svc.GenerateCostReport(context.Background(), "cr-001", "admin")
	assert.NoError(t, err)
}

func TestCostReportService_ApproveCostReport_Success(t *testing.T) {
	crRepo := new(mockCRRepoSvc)
	ceRepo := new(mockCERepoSvc)
	svc := NewCostReportService(crRepo, ceRepo)

	crRepo.On("GetByID", mock.Anything, "cr-001").Return(&entity.CostReport{ID: "cr-001", Status: "generated"}, nil)
	crRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.CostReport")).Return(nil)

	err := svc.ApproveCostReport(context.Background(), "cr-001", "admin")
	assert.NoError(t, err)
}

func TestCostReportService_RejectCostReport_Success(t *testing.T) {
	crRepo := new(mockCRRepoSvc)
	ceRepo := new(mockCERepoSvc)
	svc := NewCostReportService(crRepo, ceRepo)

	crRepo.On("GetByID", mock.Anything, "cr-001").Return(&entity.CostReport{ID: "cr-001", Status: "generated"}, nil)
	crRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.CostReport")).Return(nil)

	err := svc.RejectCostReport(context.Background(), "cr-001", "admin")
	assert.NoError(t, err)
}
