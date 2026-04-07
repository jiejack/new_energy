package service

import (
	"context"
	"errors"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRegionRepository 区域仓储Mock
type MockRegionRepository struct {
	mock.Mock
}

func (m *MockRegionRepository) Create(ctx context.Context, region *entity.Region) error {
	args := m.Called(ctx, region)
	return args.Error(0)
}

func (m *MockRegionRepository) Update(ctx context.Context, region *entity.Region) error {
	args := m.Called(ctx, region)
	return args.Error(0)
}

func (m *MockRegionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRegionRepository) GetByID(ctx context.Context, id string) (*entity.Region, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Region), args.Error(1)
}

func (m *MockRegionRepository) GetByCode(ctx context.Context, code string) (*entity.Region, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Region), args.Error(1)
}

func (m *MockRegionRepository) List(ctx context.Context, parentID *string) ([]*entity.Region, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Region), args.Error(1)
}

func (m *MockRegionRepository) GetTree(ctx context.Context) ([]*entity.Region, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Region), args.Error(1)
}

// MockSubRegionRepository 子区域仓储Mock
type MockSubRegionRepository struct {
	mock.Mock
}

func (m *MockSubRegionRepository) Create(ctx context.Context, subRegion *entity.SubRegion) error {
	args := m.Called(ctx, subRegion)
	return args.Error(0)
}

func (m *MockSubRegionRepository) Update(ctx context.Context, subRegion *entity.SubRegion) error {
	args := m.Called(ctx, subRegion)
	return args.Error(0)
}

func (m *MockSubRegionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSubRegionRepository) GetByID(ctx context.Context, id string) (*entity.SubRegion, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.SubRegion), args.Error(1)
}

func (m *MockSubRegionRepository) GetByRegionID(ctx context.Context, regionID string) ([]*entity.SubRegion, error) {
	args := m.Called(ctx, regionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.SubRegion), args.Error(1)
}

func TestRegionService_CreateRegion_Success(t *testing.T) {
	ctx := context.Background()

	mockRegionRepo := new(MockRegionRepository)
	mockSubRegionRepo := new(MockSubRegionRepository)
	service := NewRegionService(mockRegionRepo, mockSubRegionRepo)

	req := &CreateRegionRequest{
		Code:        "R001",
		Name:        "华北地区",
		ParentID:     nil,
		Level:        1,
		SortOrder:    1,
		Description: "华北地区描述",
	}

	mockRegionRepo.On("GetByCode", ctx, "R001").Return(nil, errors.New("not found"))
	mockRegionRepo.On("Create", ctx, mock.AnythingOfType("*entity.Region")).Return(nil)

	region, err := service.CreateRegion(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, region)
	assert.Equal(t, "R001", region.Code)
	assert.Equal(t, "华北地区", region.Name)
	mockRegionRepo.AssertExpectations(t)
}

func TestRegionService_CreateRegion_AlreadyExists(t *testing.T) {
	ctx := context.Background()

	mockRegionRepo := new(MockRegionRepository)
	mockSubRegionRepo := new(MockSubRegionRepository)
	service := NewRegionService(mockRegionRepo, mockSubRegionRepo)

	existingRegion := entity.NewRegion("R001", "已存在地区", nil, 1)
	req := &CreateRegionRequest{
		Code: "R001",
		Name: "华北地区",
	}

	mockRegionRepo.On("GetByCode", ctx, "R001").Return(existingRegion, nil)

	region, err := service.CreateRegion(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	assert.Nil(t, region)
	mockRegionRepo.AssertExpectations(t)
}

func TestRegionService_GetRegion_Success(t *testing.T) {
	ctx := context.Background()

	mockRegionRepo := new(MockRegionRepository)
	mockSubRegionRepo := new(MockSubRegionRepository)
	service := NewRegionService(mockRegionRepo, mockSubRegionRepo)

	expectedRegion := entity.NewRegion("R001", "华北地区", nil, 1)
	mockRegionRepo.On("GetByID", ctx, "region001").Return(expectedRegion, nil)

	region, err := service.GetRegion(ctx, "region001")

	assert.NoError(t, err)
	assert.NotNil(t, region)
	assert.Equal(t, "R001", region.Code)
	mockRegionRepo.AssertExpectations(t)
}

func TestRegionService_GetRegion_NotFound(t *testing.T) {
	ctx := context.Background()

	mockRegionRepo := new(MockRegionRepository)
	mockSubRegionRepo := new(MockSubRegionRepository)
	service := NewRegionService(mockRegionRepo, mockSubRegionRepo)

	mockRegionRepo.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))

	region, err := service.GetRegion(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, region)
	mockRegionRepo.AssertExpectations(t)
}

func TestRegionService_UpdateRegion_Success(t *testing.T) {
	ctx := context.Background()

	mockRegionRepo := new(MockRegionRepository)
	mockSubRegionRepo := new(MockSubRegionRepository)
	service := NewRegionService(mockRegionRepo, mockSubRegionRepo)

	existingRegion := entity.NewRegion("R001", "旧名称", nil, 1)
	req := &UpdateRegionRequest{
		Name:        "新名称",
		Description: "新描述",
		SortOrder:   2,
	}

	mockRegionRepo.On("GetByID", ctx, "region001").Return(existingRegion, nil)
	mockRegionRepo.On("Update", ctx, mock.AnythingOfType("*entity.Region")).Return(nil)

	region, err := service.UpdateRegion(ctx, "region001", req)

	assert.NoError(t, err)
	assert.NotNil(t, region)
	assert.Equal(t, "新名称", region.Name)
	mockRegionRepo.AssertExpectations(t)
}

func TestRegionService_UpdateRegion_NotFound(t *testing.T) {
	ctx := context.Background()

	mockRegionRepo := new(MockRegionRepository)
	mockSubRegionRepo := new(MockSubRegionRepository)
	service := NewRegionService(mockRegionRepo, mockSubRegionRepo)

	req := &UpdateRegionRequest{
		Name: "新名称",
	}

	mockRegionRepo.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))

	region, err := service.UpdateRegion(ctx, "nonexistent", req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Nil(t, region)
	mockRegionRepo.AssertExpectations(t)
}

func TestRegionService_DeleteRegion_Success(t *testing.T) {
	ctx := context.Background()

	mockRegionRepo := new(MockRegionRepository)
	mockSubRegionRepo := new(MockSubRegionRepository)
	service := NewRegionService(mockRegionRepo, mockSubRegionRepo)

	existingRegion := entity.NewRegion("R001", "华北地区", nil, 1)
	existingRegion.ID = "region001"

	mockRegionRepo.On("GetByID", ctx, "region001").Return(existingRegion, nil)
	mockSubRegionRepo.On("GetByRegionID", ctx, existingRegion.ID).Return([]*entity.SubRegion{}, nil)
	mockRegionRepo.On("Delete", ctx, "region001").Return(nil)

	err := service.DeleteRegion(ctx, "region001")

	assert.NoError(t, err)
	mockRegionRepo.AssertExpectations(t)
	mockSubRegionRepo.AssertExpectations(t)
}

func TestRegionService_DeleteRegion_WithSubRegions(t *testing.T) {
	ctx := context.Background()

	mockRegionRepo := new(MockRegionRepository)
	mockSubRegionRepo := new(MockSubRegionRepository)
	service := NewRegionService(mockRegionRepo, mockSubRegionRepo)

	existingRegion := entity.NewRegion("R001", "华北地区", nil, 1)
	existingRegion.ID = "region001"

	subRegions := []*entity.SubRegion{
		entity.NewSubRegion("SR001", "北京", "region001"),
	}

	mockRegionRepo.On("GetByID", ctx, "region001").Return(existingRegion, nil)
	mockSubRegionRepo.On("GetByRegionID", ctx, existingRegion.ID).Return(subRegions, nil)

	err := service.DeleteRegion(ctx, "region001")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete region with sub-regions")
	mockRegionRepo.AssertExpectations(t)
	mockSubRegionRepo.AssertExpectations(t)
}

func TestRegionService_ListRegions(t *testing.T) {
	ctx := context.Background()

	mockRegionRepo := new(MockRegionRepository)
	mockSubRegionRepo := new(MockSubRegionRepository)
	service := NewRegionService(mockRegionRepo, mockSubRegionRepo)

	parentID := "parent001"
	expectedRegions := []*entity.Region{
		entity.NewRegion("R001", "华北地区", &parentID, 1),
		entity.NewRegion("R002", "华东地区", &parentID, 1),
	}

	mockRegionRepo.On("List", ctx, &parentID).Return(expectedRegions, nil)

	regions, err := service.ListRegions(ctx, &parentID)

	assert.NoError(t, err)
	assert.Len(t, regions, 2)
	mockRegionRepo.AssertExpectations(t)
}

func TestRegionService_GetRegionTree(t *testing.T) {
	ctx := context.Background()

	mockRegionRepo := new(MockRegionRepository)
	mockSubRegionRepo := new(MockSubRegionRepository)
	service := NewRegionService(mockRegionRepo, mockSubRegionRepo)

	expectedRegions := []*entity.Region{
		entity.NewRegion("R001", "华北地区", nil, 1),
		entity.NewRegion("R002", "华东地区", nil, 1),
	}

	mockRegionRepo.On("GetTree", ctx).Return(expectedRegions, nil)

	regions, err := service.GetRegionTree(ctx)

	assert.NoError(t, err)
	assert.Len(t, regions, 2)
	mockRegionRepo.AssertExpectations(t)
}
