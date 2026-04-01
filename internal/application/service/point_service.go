package service

import (
	"context"
	"fmt"
	
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type PointService struct {
	pointRepo repository.PointRepository
}

func NewPointService(pointRepo repository.PointRepository) *PointService {
	return &PointService{pointRepo: pointRepo}
}

func (s *PointService) CreatePoint(ctx context.Context, req *CreatePointRequest) (*entity.Point, error) {
	existing, _ := s.pointRepo.GetByCode(ctx, req.Code)
	if existing != nil {
		return nil, fmt.Errorf("point with code %s already exists", req.Code)
	}
	
	point := entity.NewPoint(req.Code, req.Name, req.Type)
	point.DeviceID = req.DeviceID
	point.StationID = req.StationID
	point.Unit = req.Unit
	point.Precision = req.Precision
	point.Protocol = req.Protocol
	point.Address = req.Address
	point.DataFormat = req.DataFormat
	point.ScanInterval = req.ScanInterval
	point.Deadband = req.Deadband
	
	if req.MinValue != 0 || req.MaxValue != 0 {
		point.SetRange(req.MinValue, req.MaxValue)
	}
	
	if req.IsAlarm {
		point.SetAlarmThreshold(req.AlarmHigh, req.AlarmLow)
	}
	
	if err := s.pointRepo.Create(ctx, point); err != nil {
		return nil, fmt.Errorf("failed to create point: %w", err)
	}
	
	return point, nil
}

func (s *PointService) BatchCreatePoints(ctx context.Context, reqs []*CreatePointRequest) error {
	points := make([]*entity.Point, 0, len(reqs))
	for _, req := range reqs {
		point := entity.NewPoint(req.Code, req.Name, req.Type)
		point.DeviceID = req.DeviceID
		point.StationID = req.StationID
		point.Unit = req.Unit
		point.Precision = req.Precision
		point.Protocol = req.Protocol
		point.Address = req.Address
		point.ScanInterval = req.ScanInterval
		
		if req.MinValue != 0 || req.MaxValue != 0 {
			point.SetRange(req.MinValue, req.MaxValue)
		}
		
		if req.IsAlarm {
			point.SetAlarmThreshold(req.AlarmHigh, req.AlarmLow)
		}
		
		points = append(points, point)
	}
	
	return s.pointRepo.BatchCreate(ctx, points)
}

func (s *PointService) UpdatePoint(ctx context.Context, id string, req *UpdatePointRequest) (*entity.Point, error) {
	point, err := s.pointRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("point not found: %w", err)
	}
	
	if req.Name != "" {
		point.Name = req.Name
	}
	if req.Unit != "" {
		point.Unit = req.Unit
	}
	if req.ScanInterval > 0 {
		point.ScanInterval = req.ScanInterval
	}
	if req.MinValue != 0 || req.MaxValue != 0 {
		point.SetRange(req.MinValue, req.MaxValue)
	}
	
	if err := s.pointRepo.Update(ctx, point); err != nil {
		return nil, fmt.Errorf("failed to update point: %w", err)
	}
	
	return point, nil
}

func (s *PointService) DeletePoint(ctx context.Context, id string) error {
	return s.pointRepo.Delete(ctx, id)
}

func (s *PointService) GetPoint(ctx context.Context, id string) (*entity.Point, error) {
	return s.pointRepo.GetByID(ctx, id)
}

func (s *PointService) ListPoints(ctx context.Context, deviceID *string, pointType *entity.PointType) ([]*entity.Point, error) {
	return s.pointRepo.List(ctx, deviceID, pointType)
}

func (s *PointService) GetPointsByStation(ctx context.Context, stationID string) ([]*entity.Point, error) {
	return s.pointRepo.GetByStationID(ctx, stationID)
}

func (s *PointService) GetPointsByProtocol(ctx context.Context, protocol string) ([]*entity.Point, error) {
	return s.pointRepo.GetByProtocol(ctx, protocol)
}

type CreatePointRequest struct {
	Code         string           `json:"code" binding:"required"`
	Name         string           `json:"name" binding:"required"`
	Type         entity.PointType `json:"type" binding:"required"`
	DeviceID     string           `json:"device_id"`
	StationID    string           `json:"station_id"`
	Unit         string           `json:"unit"`
	Precision    int              `json:"precision"`
	MinValue     float64          `json:"min_value"`
	MaxValue     float64          `json:"max_value"`
	Protocol     string           `json:"protocol"`
	Address      int              `json:"address"`
	DataFormat   string           `json:"data_format"`
	ScanInterval int              `json:"scan_interval"`
	Deadband     float64          `json:"deadband"`
	IsAlarm      bool             `json:"is_alarm"`
	AlarmHigh    float64          `json:"alarm_high"`
	AlarmLow     float64          `json:"alarm_low"`
}

type UpdatePointRequest struct {
	Name         string  `json:"name"`
	Unit         string  `json:"unit"`
	Precision    int     `json:"precision"`
	MinValue     float64 `json:"min_value"`
	MaxValue     float64 `json:"max_value"`
	ScanInterval int     `json:"scan_interval"`
	Deadband     float64 `json:"deadband"`
}
