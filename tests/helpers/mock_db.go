package helpers

import (
	"context"
	"sync"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"gorm.io/gorm"
)

// MockDB 模拟数据库
type MockDB struct {
	mu       sync.RWMutex
	users    map[string]*entity.User
	roles    map[string]*entity.Role
	perms    map[string]*entity.Permission
	regions  map[string]*entity.Region
	stations map[string]*entity.Station
	devices  map[string]*entity.Device
	alarms   map[string]*entity.Alarm
	points   map[string]*entity.Point
}

// NewMockDB 创建模拟数据库
func NewMockDB() *MockDB {
	return &MockDB{
		users:    make(map[string]*entity.User),
		roles:    make(map[string]*entity.Role),
		perms:    make(map[string]*entity.Permission),
		regions:  make(map[string]*entity.Region),
		stations: make(map[string]*entity.Station),
		devices:  make(map[string]*entity.Device),
		alarms:   make(map[string]*entity.Alarm),
		points:   make(map[string]*entity.Point),
	}
}

// CreateUser 创建用户
func (m *MockDB) CreateUser(ctx context.Context, user *entity.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[user.ID] = user
	return nil
}

// GetUser 获取用户
func (m *MockDB) GetUser(ctx context.Context, id string) (*entity.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	user, ok := m.users[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return user, nil
}

// GetUserByUsername 根据用户名获取用户
func (m *MockDB) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

// UpdateUser 更新用户
func (m *MockDB) UpdateUser(ctx context.Context, user *entity.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[user.ID] = user
	return nil
}

// DeleteUser 删除用户
func (m *MockDB) DeleteUser(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.users, id)
	return nil
}

// ListUsers 列出用户
func (m *MockDB) ListUsers(ctx context.Context, status *entity.UserStatus, page, pageSize int) ([]*entity.User, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	users := make([]*entity.User, 0)
	for _, user := range m.users {
		if status == nil || user.Status == *status {
			users = append(users, user)
		}
	}
	return users, int64(len(users)), nil
}

// CreateRole 创建角色
func (m *MockDB) CreateRole(ctx context.Context, role *entity.Role) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.roles[role.ID] = role
	return nil
}

// GetRole 获取角色
func (m *MockDB) GetRole(ctx context.Context, id string) (*entity.Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	role, ok := m.roles[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return role, nil
}

// ListRoles 列出角色
func (m *MockDB) ListRoles(ctx context.Context) ([]*entity.Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	roles := make([]*entity.Role, 0)
	for _, role := range m.roles {
		roles = append(roles, role)
	}
	return roles, nil
}

// CreatePermission 创建权限
func (m *MockDB) CreatePermission(ctx context.Context, perm *entity.Permission) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.perms[perm.ID] = perm
	return nil
}

// GetPermission 获取权限
func (m *MockDB) GetPermission(ctx context.Context, id string) (*entity.Permission, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	perm, ok := m.perms[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return perm, nil
}

// CreateRegion 创建区域
func (m *MockDB) CreateRegion(ctx context.Context, region *entity.Region) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.regions[region.ID] = region
	return nil
}

// GetRegion 获取区域
func (m *MockDB) GetRegion(ctx context.Context, id string) (*entity.Region, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	region, ok := m.regions[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return region, nil
}

// ListRegions 列出区域
func (m *MockDB) ListRegions(ctx context.Context, parentID *string) ([]*entity.Region, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	regions := make([]*entity.Region, 0)
	for _, region := range m.regions {
		if parentID == nil {
			if region.ParentID == nil {
				regions = append(regions, region)
			}
		} else {
			if region.ParentID != nil && *region.ParentID == *parentID {
				regions = append(regions, region)
			}
		}
	}
	return regions, nil
}

// CreateStation 创建电站
func (m *MockDB) CreateStation(ctx context.Context, station *entity.Station) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stations[station.ID] = station
	return nil
}

// GetStation 获取电站
func (m *MockDB) GetStation(ctx context.Context, id string) (*entity.Station, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	station, ok := m.stations[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return station, nil
}

// ListStations 列出电站
func (m *MockDB) ListStations(ctx context.Context, subRegionID *string, stationType *entity.StationType) ([]*entity.Station, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	stations := make([]*entity.Station, 0)
	for _, station := range m.stations {
		match := true
		if subRegionID != nil && station.SubRegionID != *subRegionID {
			match = false
		}
		if stationType != nil && station.Type != *stationType {
			match = false
		}
		if match {
			stations = append(stations, station)
		}
	}
	return stations, nil
}

// CreateDevice 创建设备
func (m *MockDB) CreateDevice(ctx context.Context, device *entity.Device) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.devices[device.ID] = device
	return nil
}

// GetDevice 获取设备
func (m *MockDB) GetDevice(ctx context.Context, id string) (*entity.Device, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	device, ok := m.devices[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return device, nil
}

// ListDevices 列出设备
func (m *MockDB) ListDevices(ctx context.Context, stationID *string, deviceType *entity.DeviceType) ([]*entity.Device, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	devices := make([]*entity.Device, 0)
	for _, device := range m.devices {
		match := true
		if stationID != nil && device.StationID != *stationID {
			match = false
		}
		if deviceType != nil && device.Type != *deviceType {
			match = false
		}
		if match {
			devices = append(devices, device)
		}
	}
	return devices, nil
}

// CreateAlarm 创建告警
func (m *MockDB) CreateAlarm(ctx context.Context, alarm *entity.Alarm) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alarms[alarm.ID] = alarm
	return nil
}

// GetAlarm 获取告警
func (m *MockDB) GetAlarm(ctx context.Context, id string) (*entity.Alarm, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	alarm, ok := m.alarms[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return alarm, nil
}

// ListAlarms 列出告警
func (m *MockDB) ListAlarms(ctx context.Context, stationID *string, level *entity.AlarmLevel, status *entity.AlarmStatus) ([]*entity.Alarm, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	alarms := make([]*entity.Alarm, 0)
	for _, alarm := range m.alarms {
		match := true
		if stationID != nil && alarm.StationID != *stationID {
			match = false
		}
		if level != nil && alarm.Level != *level {
			match = false
		}
		if status != nil && alarm.Status != *status {
			match = false
		}
		if match {
			alarms = append(alarms, alarm)
		}
	}
	return alarms, nil
}

// CreatePoint 创建采集点
func (m *MockDB) CreatePoint(ctx context.Context, point *entity.Point) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.points[point.ID] = point
	return nil
}

// GetPoint 获取采集点
func (m *MockDB) GetPoint(ctx context.Context, id string) (*entity.Point, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	point, ok := m.points[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return point, nil
}

// ListPoints 列出采集点
func (m *MockDB) ListPoints(ctx context.Context, deviceID *string, pointType *entity.PointType) ([]*entity.Point, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	points := make([]*entity.Point, 0)
	for _, point := range m.points {
		match := true
		if deviceID != nil && point.DeviceID != *deviceID {
			match = false
		}
		if pointType != nil && point.Type != *pointType {
			match = false
		}
		if match {
			points = append(points, point)
		}
	}
	return points, nil
}

// Clear 清空所有数据
func (m *MockDB) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users = make(map[string]*entity.User)
	m.roles = make(map[string]*entity.Role)
	m.perms = make(map[string]*entity.Permission)
	m.regions = make(map[string]*entity.Region)
	m.stations = make(map[string]*entity.Station)
	m.devices = make(map[string]*entity.Device)
	m.alarms = make(map[string]*entity.Alarm)
	m.points = make(map[string]*entity.Point)
}

// SeedTestData 填充测试数据
func (m *MockDB) SeedTestData() {
	// 创建测试用户
	user := entity.NewUser("testuser", "hashedpassword123")
	user.ID = "user-001"
	user.Email = "test@example.com"
	user.RealName = "测试用户"
	m.users[user.ID] = user

	// 创建测试角色
	role := entity.NewRole("admin", "管理员")
	role.ID = "role-001"
	role.Description = "系统管理员"
	m.roles[role.ID] = role

	// 创建测试权限
	perm := entity.NewPermission("user:read", "查看用户")
	perm.ID = "perm-001"
	perm.ResourceType = entity.ResourceUser
	perm.Action = entity.ActionRead
	m.perms[perm.ID] = perm

	// 创建测试区域
	region := entity.NewRegion("EAST", "华东区域", nil, 1)
	region.ID = "region-001"
	m.regions[region.ID] = region

	// 创建测试电站
	station := entity.NewStation("PV_001", "测试光伏电站", entity.StationTypePV, "region-001")
	station.ID = "station-001"
	m.stations[station.ID] = station

	// 创建测试设备
	device := entity.NewDevice("INV_001", "1号逆变器", entity.DeviceTypeInverter, "station-001")
	device.ID = "device-001"
	m.devices[device.ID] = device

	// 创建测试告警
	alarm := entity.NewAlarm("point-001", "device-001", "station-001", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "测试告警", "这是一个测试告警")
	alarm.ID = "alarm-001"
	m.alarms[alarm.ID] = alarm
}
