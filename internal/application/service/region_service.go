package service

import (
	"context"
	"fmt"
	
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type RegionService struct {
	regionRepo    repository.RegionRepository
	subRegionRepo repository.SubRegionRepository
}

func NewRegionService(regionRepo repository.RegionRepository, subRegionRepo repository.SubRegionRepository) *RegionService {
	return &RegionService{
		regionRepo:    regionRepo,
		subRegionRepo: subRegionRepo,
	}
}

func (s *RegionService) CreateRegion(ctx context.Context, req *CreateRegionRequest) (*entity.Region, error) {
	existing, _ := s.regionRepo.GetByCode(ctx, req.Code)
	if existing != nil {
		return nil, fmt.Errorf("region with code %s already exists", req.Code)
	}
	
	region := entity.NewRegion(req.Code, req.Name, req.ParentID, req.Level)
	region.Description = req.Description
	region.SortOrder = req.SortOrder
	
	if err := s.regionRepo.Create(ctx, region); err != nil {
		return nil, fmt.Errorf("failed to create region: %w", err)
	}
	
	return region, nil
}

func (s *RegionService) UpdateRegion(ctx context.Context, id string, req *UpdateRegionRequest) (*entity.Region, error) {
	region, err := s.regionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("region not found: %w", err)
	}
	
	if req.Name != "" {
		region.Name = req.Name
	}
	if req.Description != "" {
		region.Description = req.Description
	}
	if req.SortOrder > 0 {
		region.SortOrder = req.SortOrder
	}
	
	if err := s.regionRepo.Update(ctx, region); err != nil {
		return nil, fmt.Errorf("failed to update region: %w", err)
	}
	
	return region, nil
}

func (s *RegionService) DeleteRegion(ctx context.Context, id string) error {
	region, err := s.regionRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("region not found: %w", err)
	}
	
	subRegions, err := s.subRegionRepo.GetByRegionID(ctx, region.ID)
	if err == nil && len(subRegions) > 0 {
		return fmt.Errorf("cannot delete region with sub-regions")
	}
	
	return s.regionRepo.Delete(ctx, id)
}

func (s *RegionService) GetRegion(ctx context.Context, id string) (*entity.Region, error) {
	return s.regionRepo.GetByID(ctx, id)
}

func (s *RegionService) ListRegions(ctx context.Context, parentID *string) ([]*entity.Region, error) {
	return s.regionRepo.List(ctx, parentID)
}

func (s *RegionService) GetRegionTree(ctx context.Context) ([]*entity.Region, error) {
	return s.regionRepo.GetTree(ctx)
}

type CreateRegionRequest struct {
	Code        string  `json:"code" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	ParentID    *string `json:"parent_id"`
	Level       int     `json:"level"`
	SortOrder   int     `json:"sort_order"`
	Description string  `json:"description"`
}

type UpdateRegionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}
