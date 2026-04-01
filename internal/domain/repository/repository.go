package repository

import (
	"context"
	
	"github.com/new-energy-monitoring/internal/domain/entity"
)

type RegionRepository interface {
	Create(ctx context.Context, region *entity.Region) error
	Update(ctx context.Context, region *entity.Region) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.Region, error)
	GetByCode(ctx context.Context, code string) (*entity.Region, error)
	List(ctx context.Context, parentID *string) ([]*entity.Region, error)
	GetTree(ctx context.Context) ([]*entity.Region, error)
}

type SubRegionRepository interface {
	Create(ctx context.Context, subRegion *entity.SubRegion) error
	Update(ctx context.Context, subRegion *entity.SubRegion) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.SubRegion, error)
	GetByRegionID(ctx context.Context, regionID string) ([]*entity.SubRegion, error)
}

type StationRepository interface {
	Create(ctx context.Context, station *entity.Station) error
	Update(ctx context.Context, station *entity.Station) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.Station, error)
	GetByCode(ctx context.Context, code string) (*entity.Station, error)
	List(ctx context.Context, subRegionID *string, stationType *entity.StationType) ([]*entity.Station, error)
	GetWithDevices(ctx context.Context, id string) (*entity.Station, error)
}

type DeviceRepository interface {
	Create(ctx context.Context, device *entity.Device) error
	Update(ctx context.Context, device *entity.Device) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.Device, error)
	GetByCode(ctx context.Context, code string) (*entity.Device, error)
	List(ctx context.Context, stationID *string, deviceType *entity.DeviceType) ([]*entity.Device, error)
	GetWithPoints(ctx context.Context, id string) (*entity.Device, error)
	GetOnlineDevices(ctx context.Context, stationID string) ([]*entity.Device, error)
}

type PointRepository interface {
	Create(ctx context.Context, point *entity.Point) error
	BatchCreate(ctx context.Context, points []*entity.Point) error
	Update(ctx context.Context, point *entity.Point) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.Point, error)
	GetByCode(ctx context.Context, code string) (*entity.Point, error)
	List(ctx context.Context, deviceID *string, pointType *entity.PointType) ([]*entity.Point, error)
	GetByStationID(ctx context.Context, stationID string) ([]*entity.Point, error)
	GetByProtocol(ctx context.Context, protocol string) ([]*entity.Point, error)
}

type AlarmRepository interface {
	Create(ctx context.Context, alarm *entity.Alarm) error
	Update(ctx context.Context, alarm *entity.Alarm) error
	GetByID(ctx context.Context, id string) (*entity.Alarm, error)
	GetActiveAlarms(ctx context.Context, stationID *string, level *entity.AlarmLevel) ([]*entity.Alarm, error)
	GetHistoryAlarms(ctx context.Context, stationID *string, startTime, endTime int64) ([]*entity.Alarm, error)
	Acknowledge(ctx context.Context, id, by string) error
	Clear(ctx context.Context, id string) error
	CountByLevel(ctx context.Context, stationID *string) (map[entity.AlarmLevel]int64, error)
}
