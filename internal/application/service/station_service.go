package service

import (
	"context"
	"fmt"
	
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type StationService struct {
	stationRepo repository.StationRepository
	deviceRepo  repository.DeviceRepository
	pointRepo   repository.PointRepository
}

func NewStationService(
	stationRepo repository.StationRepository,
	deviceRepo repository.DeviceRepository,
	pointRepo repository.PointRepository,
) *StationService {
	return &StationService{
		stationRepo: stationRepo,
		deviceRepo:  deviceRepo,
		pointRepo:   pointRepo,
	}
}

func (s *StationService) CreateStation(ctx context.Context, req *CreateStationRequest) (*entity.Station, error) {
	existing, _ := s.stationRepo.GetByCode(ctx, req.Code)
	if existing != nil {
		return nil, fmt.Errorf("station with code %s already exists", req.Code)
	}
	
	station := entity.NewStation(req.Code, req.Name, req.Type, req.SubRegionID)
	
	if req.Capacity > 0 {
		station.SetCapacity(req.Capacity, req.VoltageLevel)
	}
	
	if req.Longitude != 0 && req.Latitude != 0 {
		station.SetLocation(req.Longitude, req.Latitude, req.Address)
	}
	
	if err := s.stationRepo.Create(ctx, station); err != nil {
		return nil, fmt.Errorf("failed to create station: %w", err)
	}
	
	return station, nil
}

func (s *StationService) UpdateStation(ctx context.Context, id string, req *UpdateStationRequest) (*entity.Station, error) {
	station, err := s.stationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("station not found: %w", err)
	}
	
	if req.Name != "" {
		station.Name = req.Name
	}
	if req.Capacity > 0 {
		station.Capacity = req.Capacity
	}
	if req.VoltageLevel != "" {
		station.VoltageLevel = req.VoltageLevel
	}
	if req.Address != "" {
		station.Address = req.Address
	}
	
	if err := s.stationRepo.Update(ctx, station); err != nil {
		return nil, fmt.Errorf("failed to update station: %w", err)
	}
	
	return station, nil
}

func (s *StationService) DeleteStation(ctx context.Context, id string) error {
	station, err := s.stationRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("station not found: %w", err)
	}
	
	devices, _ := s.deviceRepo.List(ctx, &station.ID, nil)
	if len(devices) > 0 {
		return fmt.Errorf("cannot delete station with devices")
	}
	
	return s.stationRepo.Delete(ctx, id)
}

func (s *StationService) GetStation(ctx context.Context, id string) (*entity.Station, error) {
	return s.stationRepo.GetByID(ctx, id)
}

func (s *StationService) GetStationWithDevices(ctx context.Context, id string) (*entity.Station, error) {
	return s.stationRepo.GetWithDevices(ctx, id)
}

func (s *StationService) ListStations(ctx context.Context, subRegionID *string, stationType *entity.StationType) ([]*entity.Station, error) {
	return s.stationRepo.List(ctx, subRegionID, stationType)
}

func (s *StationService) GetStationPoints(ctx context.Context, id string) ([]*entity.Point, error) {
	return s.pointRepo.GetByStationID(ctx, id)
}

type CreateStationRequest struct {
	Code         string             `json:"code" binding:"required"`
	Name         string             `json:"name" binding:"required"`
	Type         entity.StationType `json:"type" binding:"required"`
	SubRegionID  string             `json:"sub_region_id" binding:"required"`
	Capacity     float64            `json:"capacity"`
	VoltageLevel string             `json:"voltage_level"`
	Longitude    float64            `json:"longitude"`
	Latitude     float64            `json:"latitude"`
	Address      string             `json:"address"`
}

type UpdateStationRequest struct {
	Name         string  `json:"name"`
	Capacity     float64 `json:"capacity"`
	VoltageLevel string  `json:"voltage_level"`
	Address      string  `json:"address"`
}
