package persistence

import (
	"context"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

// DeviceRepository 设备仓储
type DeviceRepository struct {
	db *Database
}

// NewDeviceRepository 创建设备仓储
func NewDeviceRepository(db *Database) repository.DeviceRepository {
	return &DeviceRepository{db: db}
}

// Create 创建设备
func (r *DeviceRepository) Create(ctx context.Context, device *entity.Device) error {
	return r.db.WithContext(ctx).Create(device).Error
}

// Update 更新设备
func (r *DeviceRepository) Update(ctx context.Context, device *entity.Device) error {
	return r.db.WithContext(ctx).Save(device).Error
}

// Delete 删除设备
func (r *DeviceRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Device{}, "id = ?", id).Error
}

// GetByID 根据ID获取设备
func (r *DeviceRepository) GetByID(ctx context.Context, id string) (*entity.Device, error) {
	var device entity.Device
	err := r.db.WithContext(ctx).First(&device, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// GetByCode 根据编码获取设备
func (r *DeviceRepository) GetByCode(ctx context.Context, code string) (*entity.Device, error) {
	var device entity.Device
	err := r.db.WithContext(ctx).First(&device, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// List 获取设备列表
func (r *DeviceRepository) List(ctx context.Context, stationID *string, deviceType *entity.DeviceType) ([]*entity.Device, error) {
	var devices []*entity.Device
	query := r.db.WithContext(ctx)
	
	if stationID != nil {
		query = query.Where("station_id = ?", *stationID)
	}
	if deviceType != nil {
		query = query.Where("type = ?", *deviceType)
	}
	
	err := query.Find(&devices).Error
	return devices, err
}

// GetOnlineDevices 获取在线设备列表
func (r *DeviceRepository) GetOnlineDevices(ctx context.Context, stationID string) ([]*entity.Device, error) {
	var devices []*entity.Device
	err := r.db.WithContext(ctx).Where("station_id = ? AND status = ?", stationID, "online").Find(&devices).Error
	return devices, err
}

// GetWithPoints 获取设备及其采集点
func (r *DeviceRepository) GetWithPoints(ctx context.Context, id string) (*entity.Device, error) {
	var device entity.Device
	err := r.db.WithContext(ctx).Preload("Points").First(&device, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}
