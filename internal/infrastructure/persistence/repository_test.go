package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlarmRuleRepository_RealDB_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewAlarmRuleRepository(db)
	ctx := context.Background()

	rule := entity.NewAlarmRule("Test Rule", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "value > threshold")
	rule.ID = uuid.New().String()
	rule.Threshold = 85.0
	rule.Duration = 60

	err := repo.Create(ctx, rule)
	require.NoError(t, err)
	assert.NotEmpty(t, rule.ID)

	found, err := repo.GetByID(ctx, rule.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Test Rule", found.Name)
	assert.Equal(t, entity.AlarmRuleTypeLimit, found.Type)
	assert.Equal(t, entity.AlarmLevelWarning, found.Level)
	assert.Equal(t, 85.0, found.Threshold)

	found, err = repo.GetByName(ctx, "Test Rule")
	assert.NoError(t, err)
	assert.Equal(t, rule.ID, found.ID)

	rule.Threshold = 90.0
	err = repo.Update(ctx, rule)
	assert.NoError(t, err)

	found, err = repo.GetByID(ctx, rule.ID)
	assert.NoError(t, err)
	assert.Equal(t, 90.0, found.Threshold)

	err = repo.Delete(ctx, rule.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(ctx, rule.ID)
	assert.Error(t, err)
}

func TestAlarmRuleRepository_RealDB_List(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewAlarmRuleRepository(db)
	ctx := context.Background()

	r1 := entity.NewAlarmRule("Rule 1", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "value > threshold")
	r1.ID = uuid.New().String()
	r2 := entity.NewAlarmRule("Rule 2", entity.AlarmRuleTypeTrend, entity.AlarmLevelMajor, "value < threshold")
	r2.ID = uuid.New().String()
	err := repo.Create(ctx, r1)
	require.NoError(t, err)
	err = repo.Create(ctx, r2)
	require.NoError(t, err)
	err = db.DB.Exec("UPDATE alarm_rules SET status = 0 WHERE id = ?", r2.ID).Error
	require.NoError(t, err)

	rules, total, err := repo.List(ctx, &repository.AlarmRuleQuery{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, rules, 2)

	rt := entity.AlarmRuleTypeLimit
	rules, total, err = repo.List(ctx, &repository.AlarmRuleQuery{Page: 1, PageSize: 10, Type: &rt})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, rules, 1)
	assert.Equal(t, entity.AlarmRuleTypeLimit, rules[0].Type)

	st := entity.AlarmRuleStatusEnabled
	rules, total, err = repo.List(ctx, &repository.AlarmRuleQuery{Page: 1, PageSize: 10, Status: &st})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
}

func TestAlarmRuleRepository_RealDB_GetEnabledRules(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewAlarmRuleRepository(db)
	ctx := context.Background()

	r1 := entity.NewAlarmRule("Enabled Rule", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "value > threshold")
	r1.ID = uuid.New().String()
	r2 := entity.NewAlarmRule("Disabled Rule", entity.AlarmRuleTypeTrend, entity.AlarmLevelMajor, "value < threshold")
	r2.ID = uuid.New().String()
	err := repo.Create(ctx, r1)
	require.NoError(t, err)
	err = repo.Create(ctx, r2)
	require.NoError(t, err)
	err = db.DB.Exec("UPDATE alarm_rules SET status = 0 WHERE id = ?", r2.ID).Error
	require.NoError(t, err)

	rules, err := repo.GetEnabledRules(ctx)
	assert.NoError(t, err)
	assert.Len(t, rules, 1)
	assert.Equal(t, "Enabled Rule", rules[0].Name)
}

func TestAlarmRuleRepository_RealDB_GetRulesByPointID(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewAlarmRuleRepository(db)
	ctx := context.Background()

	pointID := uuid.New().String()
	r1 := entity.NewAlarmRule("Point Rule", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "value > threshold")
	r1.ID = uuid.New().String()
	r1.PointID = &pointID
	err := repo.Create(ctx, r1)
	require.NoError(t, err)

	rules, err := repo.GetRulesByPointID(ctx, pointID)
	assert.NoError(t, err)
	assert.Len(t, rules, 1)

	rules, err = repo.GetRulesByPointID(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.Len(t, rules, 0)
}

func TestAlarmRuleRepository_RealDB_GetRulesByDeviceID(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewAlarmRuleRepository(db)
	ctx := context.Background()

	deviceID := uuid.New().String()
	r1 := entity.NewAlarmRule("Device Rule", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "value > threshold")
	r1.ID = uuid.New().String()
	r1.DeviceID = &deviceID
	err := repo.Create(ctx, r1)
	require.NoError(t, err)

	rules, err := repo.GetRulesByDeviceID(ctx, deviceID)
	assert.NoError(t, err)
	assert.Len(t, rules, 1)
}

func TestAlarmRuleRepository_RealDB_GetRulesByStationID(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewAlarmRuleRepository(db)
	ctx := context.Background()

	stationID := uuid.New().String()
	r1 := entity.NewAlarmRule("Station Rule", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "value > threshold")
	r1.ID = uuid.New().String()
	r1.StationID = &stationID
	err := repo.Create(ctx, r1)
	require.NoError(t, err)

	rules, err := repo.GetRulesByStationID(ctx, stationID)
	assert.NoError(t, err)
	assert.Len(t, rules, 1)
}

func TestQARepository_RealDB_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewQARepository(db)
	ctx := context.Background()

	session := entity.NewQASession("user-001", "Test Session")
	err := repo.CreateSession(ctx, session)
	require.NoError(t, err)
	assert.NotEmpty(t, session.ID)

	found, err := repo.GetSessionByID(ctx, session.ID)
	assert.NoError(t, err)
	assert.Equal(t, "user-001", found.UserID)
	assert.Equal(t, "Test Session", found.Title)

	session.Title = "Updated Title"
	err = repo.UpdateSession(ctx, session)
	assert.NoError(t, err)

	err = repo.DeleteSession(ctx, session.ID)
	assert.NoError(t, err)

	_, err = repo.GetSessionByID(ctx, session.ID)
	assert.Error(t, err)
}

func TestQARepository_RealDB_Messages(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewQARepository(db)
	ctx := context.Background()

	session := entity.NewQASession("user-001", "Test Session")
	err := repo.CreateSession(ctx, session)
	require.NoError(t, err)

	msg1 := entity.NewQAMessage(session.ID, entity.QAMessageRoleUser, "Hello")
	msg2 := entity.NewQAMessage(session.ID, entity.QAMessageRoleAssistant, "Hi there!")
	err = repo.CreateMessage(ctx, msg1)
	require.NoError(t, err)
	err = repo.CreateMessage(ctx, msg2)
	require.NoError(t, err)

	found, err := repo.GetSessionWithMessages(ctx, session.ID)
	assert.NoError(t, err)
	assert.Len(t, found.Messages, 2)

	messages, total, err := repo.GetMessagesBySessionID(ctx, session.ID, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, messages, 2)

	recent, err := repo.GetRecentMessages(ctx, session.ID, 5)
	assert.NoError(t, err)
	assert.Len(t, recent, 2)

	err = repo.DeleteMessagesBySessionID(ctx, session.ID)
	assert.NoError(t, err)

	messages, total, err = repo.GetMessagesBySessionID(ctx, session.ID, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, messages, 0)
}

func TestQARepository_RealDB_ListSessionsByUserID(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewQARepository(db)
	ctx := context.Background()

	s1 := entity.NewQASession("user-001", "Session 1")
	s2 := entity.NewQASession("user-001", "Session 2")
	err := repo.CreateSession(ctx, s1)
	require.NoError(t, err)
	err = repo.CreateSession(ctx, s2)
	require.NoError(t, err)

	sessions, total, err := repo.ListSessionsByUserID(ctx, "user-001", nil, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, sessions, 2)

	active := entity.QASessionStatusActive
	sessions, total, err = repo.ListSessionsByUserID(ctx, "user-001", &active, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
}

func TestSystemConfigRepository_RealDB_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewSystemConfigRepository(db)
	ctx := context.Background()

	config := entity.NewSystemConfig("basic", "system_name", "NEM System", entity.SystemConfigValueTypeString, "System name")
	err := repo.Create(ctx, config)
	require.NoError(t, err)
	assert.NotEmpty(t, config.ID)

	found, err := repo.GetByID(ctx, config.ID)
	assert.NoError(t, err)
	assert.Equal(t, "basic", found.Category)
	assert.Equal(t, "system_name", found.Key)

	found, err = repo.GetByKey(ctx, "basic", "system_name")
	assert.NoError(t, err)
	assert.Equal(t, "NEM System", found.Value)

	config.SetValue("NEM System V2", entity.SystemConfigValueTypeString)
	err = repo.Update(ctx, config)
	assert.NoError(t, err)

	found, err = repo.GetByKey(ctx, "basic", "system_name")
	assert.NoError(t, err)
	assert.Equal(t, "NEM System V2", found.Value)

	err = repo.Delete(ctx, config.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(ctx, config.ID)
	assert.Error(t, err)
}

func TestSystemConfigRepository_RealDB_GetByCategory(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewSystemConfigRepository(db)
	ctx := context.Background()

	c1 := entity.NewSystemConfig("basic", "key1", "val1", entity.SystemConfigValueTypeString, "")
	c2 := entity.NewSystemConfig("basic", "key2", "val2", entity.SystemConfigValueTypeString, "")
	c3 := entity.NewSystemConfig("alarm", "key3", "val3", entity.SystemConfigValueTypeString, "")
	err := repo.Create(ctx, c1)
	require.NoError(t, err)
	err = repo.Create(ctx, c2)
	require.NoError(t, err)
	err = repo.Create(ctx, c3)
	require.NoError(t, err)

	configs, err := repo.GetByCategory(ctx, "basic")
	assert.NoError(t, err)
	assert.Len(t, configs, 2)
}

func TestSystemConfigRepository_RealDB_GetAll(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewSystemConfigRepository(db)
	ctx := context.Background()

	c1 := entity.NewSystemConfig("basic", "key1", "val1", entity.SystemConfigValueTypeString, "")
	c2 := entity.NewSystemConfig("alarm", "key2", "val2", entity.SystemConfigValueTypeString, "")
	err := repo.Create(ctx, c1)
	require.NoError(t, err)
	err = repo.Create(ctx, c2)
	require.NoError(t, err)

	configs, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, configs, 2)
}

func TestSystemConfigRepository_RealDB_List(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewSystemConfigRepository(db)
	ctx := context.Background()

	c1 := entity.NewSystemConfig("basic", "system_name", "val1", entity.SystemConfigValueTypeString, "")
	c2 := entity.NewSystemConfig("basic", "system_version", "val2", entity.SystemConfigValueTypeString, "")
	err := repo.Create(ctx, c1)
	require.NoError(t, err)
	err = repo.Create(ctx, c2)
	require.NoError(t, err)

	configs, total, err := repo.List(ctx, &entity.SystemConfigFilter{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, configs, 2)

	cat := "basic"
	configs, total, err = repo.List(ctx, &entity.SystemConfigFilter{Category: &cat, Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
}

func TestSystemConfigRepository_RealDB_BatchUpdate(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewSystemConfigRepository(db)
	ctx := context.Background()

	c1 := entity.NewSystemConfig("basic", "key1", "val1", entity.SystemConfigValueTypeString, "")
	c2 := entity.NewSystemConfig("basic", "key2", "val2", entity.SystemConfigValueTypeString, "")
	err := repo.Create(ctx, c1)
	require.NoError(t, err)
	err = repo.Create(ctx, c2)
	require.NoError(t, err)

	c1.Value = "updated1"
	c2.Value = "updated2"
	err = repo.BatchUpdate(ctx, []*entity.SystemConfig{c1, c2})
	assert.NoError(t, err)

	found, err := repo.GetByKey(ctx, "basic", "key1")
	assert.NoError(t, err)
	assert.Equal(t, "updated1", found.Value)
}

func TestSystemConfigRepository_RealDB_ExistsByKey(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewSystemConfigRepository(db)
	ctx := context.Background()

	c1 := entity.NewSystemConfig("basic", "key1", "val1", entity.SystemConfigValueTypeString, "")
	err := repo.Create(ctx, c1)
	require.NoError(t, err)

	exists, err := repo.ExistsByKey(ctx, "basic", "key1")
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.ExistsByKey(ctx, "basic", "nonexistent")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestNotificationConfigRepository_RealDB(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewNotificationConfigRepository(db)
	ctx := context.Background()

	nc := &entity.NotificationConfig{
		ID:        uuid.New().String(),
		Type:      entity.NotificationTypeEmail,
		Name:      "Email Config",
		Config:    entity.JSONMap{"smtp_host": "smtp.example.com", "smtp_port": 587},
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := repo.Create(ctx, nc)
	require.NoError(t, err)

	found, err := repo.GetByType(ctx, entity.NotificationTypeEmail)
	assert.NoError(t, err)
	assert.Equal(t, "Email Config", found.Name)
	assert.True(t, found.Enabled)

	configs, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, configs, 1)

	nc.Name = "Updated Email Config"
	nc.Enabled = false
	err = repo.Update(ctx, nc)
	assert.NoError(t, err)

	found, err = repo.GetByType(ctx, entity.NotificationTypeEmail)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Email Config", found.Name)
	assert.False(t, found.Enabled)
}

func TestOperationLogRepository_RealDB(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewOperationLogRepository(db)
	ctx := context.Background()

	log := entity.NewOperationLog("user-001", "admin", entity.ActionLogin)
	log.ID = uuid.New().String()
	log.SetResource("system", "sys-001")
	log.SetDetails(entity.Details{"ip": "192.168.1.1"})
	log.SetRequestInfo("192.168.1.1", "Mozilla/5.0")

	err := repo.Create(ctx, log)
	require.NoError(t, err)
	assert.NotEmpty(t, log.ID)

	found, err := repo.GetByID(ctx, log.ID)
	assert.NoError(t, err)
	assert.Equal(t, "user-001", found.UserID)
	assert.Equal(t, "admin", found.Username)
	assert.Equal(t, entity.ActionLogin, found.Action)

	logs, total, err := repo.List(ctx, &repository.OperationLogQuery{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, logs, 1)

	logs, total, err = repo.List(ctx, &repository.OperationLogQuery{Page: 1, PageSize: 10, UserID: "user-001"})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)

	logs, total, err = repo.List(ctx, &repository.OperationLogQuery{Page: 1, PageSize: 10, UserID: "nonexistent"})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)

	deleted, err := repo.DeleteBefore(ctx, time.Now().Add(1*time.Hour).Unix())
	assert.NoError(t, err)
	assert.Equal(t, int64(1), deleted)
}

func TestReportRepository_RealDB(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewReportRepository(db)
	ctx := context.Background()

	now := time.Now()
	stats, err := repo.GetStationPowerStats(ctx, "station-001", now, now)
	assert.NoError(t, err)
	assert.Equal(t, "station-001", stats.StationID)

	alarmStats, err := repo.GetStationAlarmStats(ctx, "station-001", now, now)
	assert.NoError(t, err)
	assert.Equal(t, "station-001", alarmStats.StationID)

	onlineStats, err := repo.GetStationOnlineStats(ctx, "station-001", now, now)
	assert.NoError(t, err)
	assert.Equal(t, "station-001", onlineStats.StationID)

	allPower, err := repo.GetAllStationPowerStats(ctx, now, now)
	assert.NoError(t, err)
	assert.NotNil(t, allPower)

	allAlarm, err := repo.GetAllStationAlarmStats(ctx, now, now)
	assert.NoError(t, err)
	assert.NotNil(t, allAlarm)

	allOnline, err := repo.GetAllStationOnlineStats(ctx, now, now)
	assert.NoError(t, err)
	assert.NotNil(t, allOnline)
}

func TestRegionStationPointAlarm_RealDB(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	ctx := context.Background()

	regionRepo := NewRegionRepository(db)
	subRegionRepo := NewSubRegionRepository(db)
	stationRepo := NewStationRepository(db)
	pointRepo := NewPointRepository(db)
	alarmRepo := NewAlarmRepository(db)

	region := entity.NewRegion("EAST", "East Region", nil, 1)
	region.ID = uuid.New().String()
	err := regionRepo.Create(ctx, region)
	require.NoError(t, err)

	subRegion := entity.NewSubRegion("EAST_SH", "Shanghai", region.ID)
	subRegion.ID = uuid.New().String()
	err = subRegionRepo.Create(ctx, subRegion)
	require.NoError(t, err)

	station := entity.NewStation("PV001", "PV Station", entity.StationTypePV, subRegion.ID)
	station.ID = uuid.New().String()
	err = stationRepo.Create(ctx, station)
	require.NoError(t, err)

	foundStation, err := stationRepo.GetByCode(ctx, "PV001")
	assert.NoError(t, err)
	assert.Equal(t, station.ID, foundStation.ID)

	stations, err := stationRepo.List(ctx, &subRegion.ID, nil)
	assert.NoError(t, err)
	assert.Len(t, stations, 1)

	pvType := entity.StationTypePV
	stations, err = stationRepo.List(ctx, &subRegion.ID, &pvType)
	assert.NoError(t, err)
	assert.Len(t, stations, 1)

	point := entity.NewPoint("P001", "Power", entity.PointTypeYaoCe)
	point.ID = uuid.New().String()
	point.DeviceID = uuid.New().String()
	err = pointRepo.Create(ctx, point)
	require.NoError(t, err)

	foundPoint, err := pointRepo.GetByCode(ctx, "P001")
	assert.NoError(t, err)
	assert.Equal(t, point.ID, foundPoint.ID)

	points, err := pointRepo.List(ctx, nil, nil)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(points), 1)

	alarm := entity.NewAlarm(point.ID, point.DeviceID, station.ID, entity.AlarmTypeLimit, entity.AlarmLevelWarning, "High Temp", "Temperature exceeded")
	alarm.ID = uuid.New().String()
	err = alarmRepo.Create(ctx, alarm)
	require.NoError(t, err)

	foundAlarm, err := alarmRepo.GetByID(ctx, alarm.ID)
	assert.NoError(t, err)
	assert.Equal(t, "High Temp", foundAlarm.Title)

	alarms, err := alarmRepo.GetActiveAlarms(ctx, &station.ID, nil)
	assert.NoError(t, err)
	assert.Len(t, alarms, 1)

	err = alarmRepo.Acknowledge(ctx, alarm.ID, "admin")
	assert.NoError(t, err)

	err = alarmRepo.Clear(ctx, alarm.ID)
	assert.NoError(t, err)

	counts, err := alarmRepo.CountByLevel(ctx, nil)
	assert.NoError(t, err)
	assert.NotNil(t, counts)

	regions, err := regionRepo.List(ctx, nil)
	assert.NoError(t, err)
	assert.Len(t, regions, 1)

	tree, err := regionRepo.GetTree(ctx)
	assert.NoError(t, err)
	assert.Len(t, tree, 1)
}
