package service

import (
	"context"
	"fmt"
	
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type DeviceService struct {
	deviceRepo repository.DeviceRepository
	pointRepo  repository.PointRepository
}

func NewDeviceService(deviceRepo repository.DeviceRepository, pointRepo repository.PointRepository) *DeviceService {
	return &DeviceService{
		deviceRepo: deviceRepo,
		pointRepo:  pointRepo,
	}
}

func (s *DeviceService) CreateDevice(ctx context.Context, req *CreateDeviceRequest) (*entity.Device, error) {
	existing, _ := s.deviceRepo.GetByCode(ctx, req.Code)
	if existing != nil {
		return nil, fmt.Errorf("device with code %s already exists", req.Code)
	}
	
	device := entity.NewDevice(req.Code, req.Name, req.Type, req.StationID)
	device.Manufacturer = req.Manufacturer
	device.Model = req.Model
	device.SerialNumber = req.SerialNumber
	device.RatedPower = req.RatedPower
	device.RatedVoltage = req.RatedVoltage
	device.RatedCurrent = req.RatedCurrent
	
	if req.Protocol != "" && req.IPAddress != "" {
		device.SetCommunication(req.Protocol, req.IPAddress, req.Port, req.SlaveID)
	}
	
	if err := s.deviceRepo.Create(ctx, device); err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}
	
	return device, nil
}

func (s *DeviceService) UpdateDevice(ctx context.Context, id string, req *UpdateDeviceRequest) (*entity.Device, error) {
	device, err := s.deviceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}
	
	if req.Name != "" {
		device.Name = req.Name
	}
	if req.Manufacturer != "" {
		device.Manufacturer = req.Manufacturer
	}
	if req.Model != "" {
		device.Model = req.Model
	}
	
	if err := s.deviceRepo.Update(ctx, device); err != nil {
		return nil, fmt.Errorf("failed to update device: %w", err)
	}
	
	return device, nil
}

func (s *DeviceService) DeleteDevice(ctx context.Context, id string) error {
	device, err := s.deviceRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("device not found: %w", err)
	}
	
	points, _ := s.pointRepo.List(ctx, &device.ID, nil)
	if len(points) > 0 {
		return fmt.Errorf("cannot delete device with points")
	}
	
	return s.deviceRepo.Delete(ctx, id)
}

func (s *DeviceService) GetDevice(ctx context.Context, id string) (*entity.Device, error) {
	return s.deviceRepo.GetByID(ctx, id)
}

func (s *DeviceService) GetDeviceWithPoints(ctx context.Context, id string) (*entity.Device, error) {
	return s.deviceRepo.GetWithPoints(ctx, id)
}

func (s *DeviceService) ListDevices(ctx context.Context, stationID *string, deviceType *entity.DeviceType) ([]*entity.Device, error) {
	return s.deviceRepo.List(ctx, stationID, deviceType)
}

func (s *DeviceService) GetOnlineDevices(ctx context.Context, stationID string) ([]*entity.Device, error) {
	return s.deviceRepo.GetOnlineDevices(ctx, stationID)
}

func (s *DeviceService) SetDeviceOnline(ctx context.Context, id string) error {
	device, err := s.deviceRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	device.SetOnline()
	return s.deviceRepo.Update(ctx, device)
}

func (s *DeviceService) SetDeviceOffline(ctx context.Context, id string) error {
	device, err := s.deviceRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	device.SetOffline()
	return s.deviceRepo.Update(ctx, device)
}

type CreateDeviceRequest struct {
	Code          string            `json:"code" binding:"required"`
	Name          string            `json:"name" binding:"required"`
	Type          entity.DeviceType `json:"type" binding:"required"`
	StationID     string            `json:"station_id" binding:"required"`
	Manufacturer  string            `json:"manufacturer"`
	Model         string            `json:"model"`
	SerialNumber  string            `json:"serial_number"`
	RatedPower    float64           `json:"rated_power"`
	RatedVoltage  float64           `json:"rated_voltage"`
	RatedCurrent  float64           `json:"rated_current"`
	Protocol      string            `json:"protocol"`
	IPAddress     string            `json:"ip_address"`
	Port          int               `json:"port"`
	SlaveID       int               `json:"slave_id"`
}

type UpdateDeviceRequest struct {
	Name          string  `json:"name"`
	Manufacturer  string  `json:"manufacturer"`
	Model         string  `json:"model"`
	RatedPower    float64 `json:"rated_power"`
	RatedVoltage  float64 `json:"rated_voltage"`
	RatedCurrent  float64 `json:"rated_current"`
}
