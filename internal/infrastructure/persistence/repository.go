package persistence

import (
	"context"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type RegionRepository struct {
	db *Database
}

func NewRegionRepository(db *Database) repository.RegionRepository {
	return &RegionRepository{db: db}
}

func (r *RegionRepository) Create(ctx context.Context, region *entity.Region) error {
	return r.db.WithContext(ctx).Create(region).Error
}

func (r *RegionRepository) Update(ctx context.Context, region *entity.Region) error {
	return r.db.WithContext(ctx).Save(region).Error
}

func (r *RegionRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Region{}, "id = ?", id).Error
}

func (r *RegionRepository) GetByID(ctx context.Context, id string) (*entity.Region, error) {
	var region entity.Region
	err := r.db.WithContext(ctx).First(&region, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &region, nil
}

func (r *RegionRepository) GetByCode(ctx context.Context, code string) (*entity.Region, error) {
	var region entity.Region
	err := r.db.WithContext(ctx).First(&region, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &region, nil
}

func (r *RegionRepository) List(ctx context.Context, parentID *string) ([]*entity.Region, error) {
	var regions []*entity.Region
	query := r.db.WithContext(ctx)
	
	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}
	
	err := query.Order("sort_order").Find(&regions).Error
	return regions, err
}

func (r *RegionRepository) GetTree(ctx context.Context) ([]*entity.Region, error) {
	var regions []*entity.Region
	err := r.db.WithContext(ctx).
		Preload("SubRegions").
		Preload("SubRegions.Stations").
		Where("parent_id IS NULL").
		Order("sort_order").
		Find(&regions).Error
	return regions, err
}

type StationRepository struct {
	db *Database
}

func NewStationRepository(db *Database) repository.StationRepository {
	return &StationRepository{db: db}
}

func (r *StationRepository) Create(ctx context.Context, station *entity.Station) error {
	return r.db.WithContext(ctx).Create(station).Error
}

func (r *StationRepository) Update(ctx context.Context, station *entity.Station) error {
	return r.db.WithContext(ctx).Save(station).Error
}

func (r *StationRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Station{}, "id = ?", id).Error
}

func (r *StationRepository) GetByID(ctx context.Context, id string) (*entity.Station, error) {
	var station entity.Station
	err := r.db.WithContext(ctx).First(&station, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &station, nil
}

func (r *StationRepository) GetByCode(ctx context.Context, code string) (*entity.Station, error) {
	var station entity.Station
	err := r.db.WithContext(ctx).First(&station, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &station, nil
}

func (r *StationRepository) List(ctx context.Context, subRegionID *string, stationType *entity.StationType) ([]*entity.Station, error) {
	var stations []*entity.Station
	query := r.db.WithContext(ctx)
	
	if subRegionID != nil {
		query = query.Where("sub_region_id = ?", *subRegionID)
	}
	if stationType != nil {
		query = query.Where("type = ?", *stationType)
	}
	
	err := query.Find(&stations).Error
	return stations, err
}

func (r *StationRepository) GetWithDevices(ctx context.Context, id string) (*entity.Station, error) {
	var station entity.Station
	err := r.db.WithContext(ctx).
		Preload("Devices").
		Preload("Devices.Points").
		First(&station, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &station, nil
}

type PointRepository struct {
	db *Database
}

func NewPointRepository(db *Database) repository.PointRepository {
	return &PointRepository{db: db}
}

func (r *PointRepository) Create(ctx context.Context, point *entity.Point) error {
	return r.db.WithContext(ctx).Create(point).Error
}

func (r *PointRepository) BatchCreate(ctx context.Context, points []*entity.Point) error {
	return r.db.WithContext(ctx).CreateInBatches(points, 100).Error
}

func (r *PointRepository) Update(ctx context.Context, point *entity.Point) error {
	return r.db.WithContext(ctx).Save(point).Error
}

func (r *PointRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Point{}, "id = ?", id).Error
}

func (r *PointRepository) GetByID(ctx context.Context, id string) (*entity.Point, error) {
	var point entity.Point
	err := r.db.WithContext(ctx).First(&point, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &point, nil
}

func (r *PointRepository) GetByCode(ctx context.Context, code string) (*entity.Point, error) {
	var point entity.Point
	err := r.db.WithContext(ctx).First(&point, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &point, nil
}

func (r *PointRepository) List(ctx context.Context, deviceID *string, pointType *entity.PointType) ([]*entity.Point, error) {
	var points []*entity.Point
	query := r.db.WithContext(ctx)
	
	if deviceID != nil {
		query = query.Where("device_id = ?", *deviceID)
	}
	if pointType != nil {
		query = query.Where("type = ?", *pointType)
	}
	
	err := query.Find(&points).Error
	return points, err
}

func (r *PointRepository) GetByStationID(ctx context.Context, stationID string) ([]*entity.Point, error) {
	var points []*entity.Point
	err := r.db.WithContext(ctx).
		Joins("JOIN devices ON devices.id = points.device_id").
		Where("devices.station_id = ?", stationID).
		Find(&points).Error
	return points, err
}

func (r *PointRepository) GetByProtocol(ctx context.Context, protocol string) ([]*entity.Point, error) {
	var points []*entity.Point
	err := r.db.WithContext(ctx).Where("protocol = ?", protocol).Find(&points).Error
	return points, err
}

type AlarmRepository struct {
	db *Database
}

func NewAlarmRepository(db *Database) repository.AlarmRepository {
	return &AlarmRepository{db: db}
}

func (r *AlarmRepository) Create(ctx context.Context, alarm *entity.Alarm) error {
	return r.db.WithContext(ctx).Create(alarm).Error
}

func (r *AlarmRepository) Update(ctx context.Context, alarm *entity.Alarm) error {
	return r.db.WithContext(ctx).Save(alarm).Error
}

func (r *AlarmRepository) GetByID(ctx context.Context, id string) (*entity.Alarm, error) {
	var alarm entity.Alarm
	err := r.db.WithContext(ctx).First(&alarm, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &alarm, nil
}

func (r *AlarmRepository) GetActiveAlarms(ctx context.Context, stationID *string, level *entity.AlarmLevel) ([]*entity.Alarm, error) {
	var alarms []*entity.Alarm
	query := r.db.WithContext(ctx).Where("status = ?", entity.AlarmStatusActive)
	
	if stationID != nil {
		query = query.Where("station_id = ?", *stationID)
	}
	if level != nil {
		query = query.Where("level = ?", *level)
	}
	
	err := query.Order("triggered_at DESC").Find(&alarms).Error
	return alarms, err
}

func (r *AlarmRepository) GetHistoryAlarms(ctx context.Context, stationID *string, startTime, endTime int64) ([]*entity.Alarm, error) {
	var alarms []*entity.Alarm
	query := r.db.WithContext(ctx).
		Where("triggered_at >= ?", time.Unix(startTime, 0)).
		Where("triggered_at <= ?", time.Unix(endTime, 0))
	
	if stationID != nil {
		query = query.Where("station_id = ?", *stationID)
	}
	
	err := query.Order("triggered_at DESC").Find(&alarms).Error
	return alarms, err
}

func (r *AlarmRepository) Acknowledge(ctx context.Context, id, by string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entity.Alarm{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":          entity.AlarmStatusAcknowledged,
			"acknowledged_at": &now,
			"acknowledged_by": by,
		}).Error
}

func (r *AlarmRepository) Clear(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entity.Alarm{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     entity.AlarmStatusCleared,
			"cleared_at": &now,
		}).Error
}

func (r *AlarmRepository) CountByLevel(ctx context.Context, stationID *string) (map[entity.AlarmLevel]int64, error) {
	type CountResult struct {
		Level entity.AlarmLevel
		Count int64
	}

	var results []CountResult
	query := r.db.WithContext(ctx).
		Model(&entity.Alarm{}).
		Select("level, count(*) as count").
		Where("status = ?", entity.AlarmStatusActive).
		Group("level")

	if stationID != nil {
		query = query.Where("station_id = ?", *stationID)
	}

	err := query.Scan(&results).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[entity.AlarmLevel]int64)
	for _, r := range results {
		counts[r.Level] = r.Count
	}

	return counts, nil
}

type SubRegionRepository struct {
	db *Database
}

func NewSubRegionRepository(db *Database) repository.SubRegionRepository {
	return &SubRegionRepository{db: db}
}

func (r *SubRegionRepository) Create(ctx context.Context, subRegion *entity.SubRegion) error {
	return r.db.WithContext(ctx).Create(subRegion).Error
}

func (r *SubRegionRepository) Update(ctx context.Context, subRegion *entity.SubRegion) error {
	return r.db.WithContext(ctx).Save(subRegion).Error
}

func (r *SubRegionRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.SubRegion{}, "id = ?", id).Error
}

func (r *SubRegionRepository) GetByID(ctx context.Context, id string) (*entity.SubRegion, error) {
	var subRegion entity.SubRegion
	err := r.db.WithContext(ctx).First(&subRegion, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &subRegion, nil
}

func (r *SubRegionRepository) GetByRegionID(ctx context.Context, regionID string) ([]*entity.SubRegion, error) {
	var subRegions []*entity.SubRegion
	err := r.db.WithContext(ctx).Where("region_id = ?", regionID).Find(&subRegions).Error
	return subRegions, err
}
