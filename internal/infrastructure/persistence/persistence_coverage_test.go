package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

func newTestDB(t *testing.T) (*Database, func()) {
	t.Helper()
	dialector := sqlite.Open(":memory:")
	cfg := DatabaseConfig{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
	}
	db, err := NewDatabaseWithDialector(dialector, cfg)
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

func setupAllTables(db *Database) error {
	return db.DB.AutoMigrate(
		&entity.Alarm{},
		&entity.AlarmRule{},
		&entity.Device{},
		&entity.Point{},
		&entity.User{},
		&entity.Role{},
		&entity.Permission{},
		&entity.UserRole{},
		&entity.RolePermission{},
		&entity.SystemConfig{},
		&entity.QASession{},
		&entity.QAMessage{},
		&entity.OperationLog{},
		&entity.NotificationConfig{},
		&entity.EnergyEfficiencyRecord{},
		&entity.EnergyEfficiencyAnalysis{},
		&entity.WorkOrder{},
		&entity.Inventory{},
		&entity.InventoryTransaction{},
		&entity.Supplier{},
		&entity.PurchaseOrder{},
		&entity.PurchaseOrderItem{},
		&entity.Receipt{},
		&entity.ReceiptItem{},
		&entity.CostCategory{},
		&entity.CostEntry{},
		&entity.CostAllocation{},
		&entity.CostReport{},
		&entity.Asset{},
		&entity.AssetMaintenanceRecord{},
		&entity.AssetDepreciationRecord{},
		&entity.AssetDocument{},
		&entity.ForecastResult{},
		&entity.FaultDetectionResult{},
		&entity.EdgeNode{},
		&entity.ModelVersion{},
		&entity.Station{},
		&entity.Region{},
		&entity.SubRegion{},
		&entity.CarbonEmissionFactor{},
		&entity.CarbonEmissionRecord{},
		&entity.CarbonEmissionSummary{},
		&entity.CarbonReductionTarget{},
		&entity.MaintenanceRecord{},
		&entity.SparePart{},
		&entity.DeviceDocument{},
	)
}

func TestNewDatabaseWithDialector_SQLite(t *testing.T) {
	dialector := sqlite.Open(":memory:")
	cfg := DatabaseConfig{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
	}
	db, err := NewDatabaseWithDialector(dialector, cfg)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	var result int
	err = db.DB.Raw("SELECT 1").Scan(&result).Error
	require.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestNewDatabaseWithDialector_InvalidDialector(t *testing.T) {
	cfg := DatabaseConfig{}
	db, err := NewDatabaseWithDialector(nil, cfg)
	require.Error(t, err)
	assert.Nil(t, db)
}

func TestDatabase_Ping(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	err := db.Ping(context.Background())
	require.NoError(t, err)
}

func TestDatabase_IsReady(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	assert.True(t, db.IsReady(context.Background()))
}

func TestDatabase_GetStats(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	stats := db.GetStats()
	require.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.MaxOpenConnections, 0)
}

func TestDatabase_HealthCheck(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	status, err := db.HealthCheck(context.Background())
	require.NoError(t, err)
	require.NotNil(t, status)
	assert.Equal(t, "healthy", status.Status)
	assert.NotEmpty(t, status.Details)
}

func TestDatabase_AutoMigrate(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	err := db.AutoMigrate()
	require.NoError(t, err)

	migrator := db.DB.Migrator()
	tables, err := migrator.GetTables()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(tables), 5)
}

func TestDatabase_Close(t *testing.T) {
	db, cleanup := newTestDB(t)
	cleanup()

	err := db.Ping(context.Background())
	assert.Error(t, err)
}

func TestMigrationManagerCoverage(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	mgr := NewMigrationManager(db)
	require.NotNil(t, mgr)

	ctx := context.Background()

	err := mgr.createMigrationsTable(ctx)
	require.NoError(t, err)

	applied, err := mgr.getAppliedMigrations(ctx)
	require.NoError(t, err)
	assert.Empty(t, applied)

	status, err := mgr.GetMigrationStatus(ctx)
	require.NoError(t, err)
	assert.Empty(t, status)

	isApplied, err := mgr.IsMigrationApplied(ctx, "v1")
	require.NoError(t, err)
	assert.False(t, isApplied)

	err = mgr.applyMigration(ctx, "v1", "CREATE TABLE test_migration (id INTEGER PRIMARY KEY)")
	require.NoError(t, err)

	isApplied2, err := mgr.IsMigrationApplied(ctx, "v1")
	require.NoError(t, err)
	assert.True(t, isApplied2)

	err = mgr.RollbackMigration(ctx, "v1")
	require.NoError(t, err)

	isApplied3, err := mgr.IsMigrationApplied(ctx, "v1")
	require.NoError(t, err)
	assert.False(t, isApplied3)
}

func TestMigrationManagerCoverage_ValidateMigration(t *testing.T) {
	mgr := &MigrationManager{}

	err := mgr.validateMigration("")
	assert.Error(t, err)

	err = mgr.validateMigration("  ")
	assert.Error(t, err)

	err = mgr.validateMigration("CREATE TABLE test (id INT)")
	assert.NoError(t, err)
}

func TestMigrationManagerCoverage_RollbackWithScript(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	mgr := NewMigrationManager(db)
	ctx := context.Background()

	mgr.createMigrationsTable(ctx)
	mgr.applyMigration(ctx, "v2", "CREATE TABLE test_rb (id INTEGER PRIMARY KEY)")

	err := mgr.RollbackMigrationWithScript(ctx, "v2", "DROP TABLE IF EXISTS test_rb")
	require.NoError(t, err)

	isApplied, _ := mgr.IsMigrationApplied(ctx, "v2")
	assert.False(t, isApplied)
}

func TestAlarmRuleRepository_CRUD(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	setupAllTables(db)

	repo := NewAlarmRuleRepository(db)
	ctx := context.Background()

	rule := entity.NewAlarmRule("test-rule", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "value > 100")
	rule.ID = uuid.New().String()
	rule.Description = "test description"
	pointID := "point-001"
	deviceID := "device-001"
	stationID := "station-001"
	rule.PointID = &pointID
	rule.DeviceID = &deviceID
	rule.StationID = &stationID
	rule.Threshold = 100.0
	rule.Duration = 60
	rule.NotifyChannels = []string{"email"}
	rule.NotifyUsers = []string{"admin"}

	t.Run("Create", func(t *testing.T) {
		err := repo.Create(ctx, rule)
		require.NoError(t, err)
	})

	t.Run("GetByID", func(t *testing.T) {
		found, err := repo.GetByID(ctx, rule.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, rule.Name, found.Name)
		assert.Equal(t, rule.Type, found.Type)
		assert.Equal(t, entity.AlarmRuleStatusEnabled, found.Status)
	})

	t.Run("GetByName", func(t *testing.T) {
		found, err := repo.GetByName(ctx, "test-rule")
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, rule.ID, found.ID)
	})

	t.Run("Update", func(t *testing.T) {
		rule.Description = "updated description"
		rule.Status = entity.AlarmRuleStatusDisabled
		err := repo.Update(ctx, rule)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, rule.ID)
		require.NoError(t, err)
		assert.Equal(t, "updated description", found.Description)
		assert.Equal(t, entity.AlarmRuleStatusDisabled, found.Status)
	})

	t.Run("List", func(t *testing.T) {
		query := &repository.AlarmRuleQuery{
			Page:     1,
			PageSize: 10,
		}
		results, total, err := repo.List(ctx, query)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(1))
		assert.Len(t, results, 1)
	})

	t.Run("ListByLevel", func(t *testing.T) {
		level := entity.AlarmLevelWarning
		query := &repository.AlarmRuleQuery{
			Page:     1,
			Level:    &level,
			PageSize: 10,
		}
		results, _, err := repo.List(ctx, query)
		require.NoError(t, err)
		assert.Len(t, results, 1)
	})

	t.Run("ListByStationID", func(t *testing.T) {
		query := &repository.AlarmRuleQuery{
			Page:      1,
			StationID: &stationID,
			PageSize:  10,
		}
		results, _, err := repo.List(ctx, query)
		require.NoError(t, err)
		assert.Len(t, results, 1)
	})

	t.Run("GetRulesByPointID", func(t *testing.T) {
		rules, err := repo.GetRulesByPointID(ctx, pointID)
		require.NoError(t, err)
		assert.Len(t, rules, 0)
	})

	t.Run("GetRulesByDeviceID", func(t *testing.T) {
		rules, err := repo.GetRulesByDeviceID(ctx, deviceID)
		require.NoError(t, err)
		assert.Len(t, rules, 0)
	})

	t.Run("GetRulesByStationID", func(t *testing.T) {
		rules, err := repo.GetRulesByStationID(ctx, stationID)
		require.NoError(t, err)
		assert.Len(t, rules, 0)
	})

	t.Run("GetEnabledRules", func(t *testing.T) {
		rules, err := repo.GetEnabledRules(ctx)
		require.NoError(t, err)
		assert.Len(t, rules, 0)
	})

	t.Run("Delete", func(t *testing.T) {
		err := repo.Delete(ctx, rule.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, rule.ID)
		assert.Error(t, err)
	})
}

func TestDeviceRepository_CRUD(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	setupAllTables(db)

	repo := NewDeviceRepository(db)
	ctx := context.Background()

	device := entity.NewDevice("DEV-001", "Test Device", entity.DeviceTypeInverter, "station-001")
	device.ID = uuid.New().String()
	device.Manufacturer = "Test Mfg"
	device.Model = "Model X"

	t.Run("Create", func(t *testing.T) {
		err := repo.Create(ctx, device)
		require.NoError(t, err)
	})

	t.Run("GetByID", func(t *testing.T) {
		found, err := repo.GetByID(ctx, device.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "DEV-001", found.Code)
		assert.Equal(t, entity.DeviceStatusOffline, found.Status)
	})

	t.Run("GetByCode", func(t *testing.T) {
		found, err := repo.GetByCode(ctx, "DEV-001")
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, device.ID, found.ID)
	})

	t.Run("Update", func(t *testing.T) {
		device.Name = "Updated Device"
		device.SetOnline()
		err := repo.Update(ctx, device)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, device.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Device", found.Name)
		assert.Equal(t, entity.DeviceStatusOnline, found.Status)
	})

	t.Run("List", func(t *testing.T) {
		devices, err := repo.List(ctx, nil, nil)
		require.NoError(t, err)
		assert.Len(t, devices, 1)
	})

	t.Run("ListByStation", func(t *testing.T) {
		stationID := "station-001"
		devices, err := repo.List(ctx, &stationID, nil)
		require.NoError(t, err)
		assert.Len(t, devices, 1)
	})

	t.Run("GetWithPoints", func(t *testing.T) {
		found, err := repo.GetWithPoints(ctx, device.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, device.ID, found.ID)
	})

	t.Run("GetOnlineDevices", func(t *testing.T) {
		online, err := repo.GetOnlineDevices(ctx, "station-001")
		require.NoError(t, err)
		assert.NotNil(t, online)
	})

	t.Run("Delete", func(t *testing.T) {
		err := repo.Delete(ctx, device.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, device.ID)
		assert.Error(t, err)
	})
}

func TestUserRepository_CRUD(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	setupAllTables(db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := entity.NewUser("testuser", "hashedpassword123")
	user.ID = uuid.New().String()

	t.Run("Create", func(t *testing.T) {
		err := repo.Create(ctx, user)
		require.NoError(t, err)
	})

	t.Run("GetByID", func(t *testing.T) {
		found, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "testuser", found.Username)
		assert.Equal(t, entity.UserStatusActive, found.Status)
	})

	t.Run("GetByUsername", func(t *testing.T) {
		found, err := repo.GetByUsername(ctx, "testuser")
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, user.ID, found.ID)
	})

	t.Run("Update", func(t *testing.T) {
		user.SetEmail("test@example.com")
		user.SetPhone("13800138000")
		user.SetRealName("Test User")
		err := repo.Update(ctx, user)
		require.NoError(t, err)

		found, _ := repo.GetByID(ctx, user.ID)
		assert.Equal(t, "test@example.com", found.Email)
	})

	t.Run("List", func(t *testing.T) {
		users, total, err := repo.List(ctx, nil, 1, 10)
		require.NoError(t, err)
		assert.Len(t, users, 1)
		assert.GreaterOrEqual(t, total, int64(1))
	})

	t.Run("Delete", func(t *testing.T) {
		err := repo.Delete(ctx, user.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, user.ID)
		assert.Error(t, err)
	})
}

func TestSystemConfigRepository_CRUD(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	setupAllTables(db)

	repo := NewSystemConfigRepository(db)
	ctx := context.Background()

	config := entity.NewSystemConfig("basic", "system_name", "新能源监控系统", entity.SystemConfigValueTypeString, "系统名称")

	t.Run("Create", func(t *testing.T) {
		err := repo.Create(ctx, config)
		require.NoError(t, err)
	})

	t.Run("GetByID", func(t *testing.T) {
		found, err := repo.GetByID(ctx, config.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "basic", found.Category)
		assert.Equal(t, "system_name", found.Key)
	})

	t.Run("GetByKey", func(t *testing.T) {
		found, err := repo.GetByKey(ctx, "basic", "system_name")
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, config.ID, found.ID)
	})

	t.Run("GetByKey_NotFound", func(t *testing.T) {
		_, err := repo.GetByKey(ctx, "nonexistent", "key")
		assert.Error(t, err)
	})

	t.Run("Update", func(t *testing.T) {
		config.SetValue("新系统名称", entity.SystemConfigValueTypeString)
		config.SetDescription("新的系统描述")
		err := repo.Update(ctx, config)
		require.NoError(t, err)

		found, _ := repo.GetByID(ctx, config.ID)
		assert.Equal(t, "新系统名称", found.Value)
	})

	t.Run("GetByCategory", func(t *testing.T) {
		configs, err := repo.GetByCategory(ctx, "basic")
		require.NoError(t, err)
		assert.Len(t, configs, 1)
	})

	t.Run("GetByCategory_Empty", func(t *testing.T) {
		configs, err := repo.GetByCategory(ctx, "nonexistent")
		require.NoError(t, err)
		assert.Empty(t, configs)
	})

	t.Run("GetAll", func(t *testing.T) {
		configs, err := repo.GetAll(ctx)
		require.NoError(t, err)
		assert.Len(t, configs, 1)
	})

	t.Run("ExistsByKey_True", func(t *testing.T) {
		exists, err := repo.ExistsByKey(ctx, "basic", "system_name")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("ExistsByKey_False", func(t *testing.T) {
		exists, err := repo.ExistsByKey(ctx, "basic", "nonexistent")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("List", func(t *testing.T) {
		filter := &entity.SystemConfigFilter{
			Page:     1,
			PageSize: 10,
		}
		configs, total, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(1))
		assert.NotEmpty(t, configs)
	})

	t.Run("BatchUpdate", func(t *testing.T) {
		config2 := entity.NewSystemConfig("basic", "version", "1.0.0", entity.SystemConfigValueTypeString, "")
		repo.Create(ctx, config2)

		err := repo.BatchUpdate(ctx, []*entity.SystemConfig{config, config2})
		require.NoError(t, err)
	})

	t.Run("Delete", func(t *testing.T) {
		err := repo.Delete(ctx, config.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, config.ID)
		assert.Error(t, err)
	})
}

func TestQARepository_CRUD(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	setupAllTables(db)

	repo := NewQARepository(db)
	ctx := context.Background()

	session := entity.NewQASession("user-001", "Test Session")
	session.ID = uuid.New().String()

	t.Run("CreateSession", func(t *testing.T) {
		err := repo.CreateSession(ctx, session)
		require.NoError(t, err)
	})

	t.Run("GetSessionByID", func(t *testing.T) {
		found, err := repo.GetSessionByID(ctx, session.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "Test Session", found.Title)
		assert.True(t, found.IsActive())
	})

	t.Run("GetSessionWithMessages", func(t *testing.T) {
		found, err := repo.GetSessionWithMessages(ctx, session.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, session.ID, found.ID)
	})

	t.Run("UpdateSession", func(t *testing.T) {
		session.Archive()
		err := repo.UpdateSession(ctx, session)
		require.NoError(t, err)

		found, _ := repo.GetSessionByID(ctx, session.ID)
		assert.True(t, found.IsArchived())
	})

	t.Run("ListSessionsByUserID", func(t *testing.T) {
		sessions, total, err := repo.ListSessionsByUserID(ctx, "user-001", nil, 1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(1))
		assert.NotEmpty(t, sessions)
	})

	t.Run("CreateMessage", func(t *testing.T) {
		msg := session.AddMessage(entity.QAMessageRoleUser, "Hello")
		err := repo.CreateMessage(ctx, msg)
		require.NoError(t, err)
	})

	t.Run("GetMessagesBySessionID", func(t *testing.T) {
		messages, total, err := repo.GetMessagesBySessionID(ctx, session.ID, 1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(1))
		assert.Len(t, messages, 1)
		assert.True(t, messages[0].IsUserMessage())
	})

	t.Run("GetRecentMessages", func(t *testing.T) {
		messages, err := repo.GetRecentMessages(ctx, session.ID, 10)
		require.NoError(t, err)
		assert.Len(t, messages, 1)
	})

	t.Run("DeleteMessagesBySessionID", func(t *testing.T) {
		err := repo.DeleteMessagesBySessionID(ctx, session.ID)
		require.NoError(t, err)

		messages, _, _ := repo.GetMessagesBySessionID(ctx, session.ID, 1, 10)
		assert.Empty(t, messages)
	})

	t.Run("DeleteSession", func(t *testing.T) {
		err := repo.DeleteSession(ctx, session.ID)
		require.NoError(t, err)

		_, err = repo.GetSessionByID(ctx, session.ID)
		assert.Error(t, err)
	})
}

func TestOperationLogRepository_CRUD(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	setupAllTables(db)

	repo := NewOperationLogRepository(db)
	ctx := context.Background()

	log := entity.NewOperationLog("user-001", "admin", entity.ActionLogin)
	log.SetResource(entity.ResourceSystemConfig, "config-001")
	log.SetDetails(entity.Details{"action": "login"})
	log.SetRequestInfo("127.0.0.1", "Mozilla/5.0")

	t.Run("Create", func(t *testing.T) {
		err := repo.Create(ctx, log)
		require.NoError(t, err)
		assert.NotEmpty(t, log.ID)
	})

	t.Run("GetByID", func(t *testing.T) {
		found, err := repo.GetByID(ctx, log.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "user-001", found.UserID)
		assert.Equal(t, entity.ActionLogin, found.Action)
	})

	t.Run("List", func(t *testing.T) {
		logs, total, err := repo.List(ctx, &repository.OperationLogQuery{
			UserID:   "user-001",
			Action:   entity.ActionLogin,
			Page:     1,
			PageSize: 10,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(1))
		assert.NotEmpty(t, logs)
	})

	t.Run("ListWithTimeRange", func(t *testing.T) {
		now := time.Now()
		logs, _, err := repo.List(ctx, &repository.OperationLogQuery{
			StartTime: now.Add(-1 * time.Hour).Unix(),
			EndTime:   now.Add(1 * time.Hour).Unix(),
			Page:      1,
			PageSize:  10,
		})
		require.NoError(t, err)
		assert.NotEmpty(t, logs)
	})

	t.Run("DeleteBefore", func(t *testing.T) {
		count, err := repo.DeleteBefore(ctx, time.Now().Add(1*time.Hour).Unix())
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(1))
	})
}

func TestNotificationConfigRepository_CRUD(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	setupAllTables(db)

	repo := NewNotificationConfigRepository(db)
	ctx := context.Background()

	cfg := &entity.NotificationConfig{
		ID:      uuid.New().String(),
		Type:    entity.NotificationTypeEmail,
		Name:    "Email Notification",
		Config:  entity.JSONMap{"smtp_host": "smtp.example.com", "smtp_port": 587},
		Enabled: true,
	}

	t.Run("Create", func(t *testing.T) {
		err := repo.Create(ctx, cfg)
		require.NoError(t, err)
	})

	t.Run("GetByType", func(t *testing.T) {
		found, err := repo.GetByType(ctx, entity.NotificationTypeEmail)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, cfg.ID, found.ID)
		assert.True(t, found.Enabled)
	})

	t.Run("GetByType_NotFound", func(t *testing.T) {
		_, err := repo.GetByType(ctx, entity.NotificationTypeSMS)
		assert.Error(t, err)
	})

	t.Run("Update", func(t *testing.T) {
		cfg.Enabled = false
		cfg.Config["smtp_port"] = 465
		err := repo.Update(ctx, cfg)
		require.NoError(t, err)

		found, _ := repo.GetByType(ctx, entity.NotificationTypeEmail)
		assert.False(t, found.Enabled)
	})

	t.Run("GetAll", func(t *testing.T) {
		configs, err := repo.GetAll(ctx)
		require.NoError(t, err)
		assert.Len(t, configs, 1)
	})
}

func TestEnergyEfficiencyRepository_CRUD(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	setupAllTables(db)

	repo := NewEnergyEfficiencyRepository(db)
	ctx := context.Background()

	record := entity.NewEnergyEfficiencyRecord(
		time.Now(),
		entity.EnergyEfficiencyTypeDevice,
		"device-001",
		"Device 1",
		100.0,
		85.0,
		"hourly",
	)
	record.ID = uuid.New().String()

	t.Run("CreateRecord", func(t *testing.T) {
		err := repo.CreateRecord(ctx, record)
		require.NoError(t, err)
	})

	t.Run("GetRecordByID", func(t *testing.T) {
		found, err := repo.GetRecordByID(ctx, record.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, entity.EnergyEfficiencyTypeDevice, found.Type)
		assert.InDelta(t, 0.85, found.Efficiency, 0.01)
	})

	t.Run("BatchCreateRecords", func(t *testing.T) {
		r2 := entity.NewEnergyEfficiencyRecord(time.Now(), entity.EnergyEfficiencyTypeStation, "station-001", "Station 1", 200.0, 180.0, "hourly")
		r2.ID = uuid.New().String()
		err := repo.BatchCreateRecords(ctx, []*entity.EnergyEfficiencyRecord{r2})
		require.NoError(t, err)
	})

	t.Run("ListRecords", func(t *testing.T) {
		eeType := entity.EnergyEfficiencyTypeDevice
		records, total, err := repo.ListRecords(ctx, &repository.EnergyEfficiencyQuery{
			Type:     &eeType,
			TargetID: strPtr("device-001"),
			Page:     1,
			PageSize: 10,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(1))
		assert.NotEmpty(t, records)
	})

	t.Run("GetRecordsByTimeRange", func(t *testing.T) {
		records, err := repo.GetRecordsByTimeRange(ctx, "device-001", entity.EnergyEfficiencyTypeDevice, time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour))
		require.NoError(t, err)
		assert.NotEmpty(t, records)
	})

	t.Run("CreateAnalysis", func(t *testing.T) {
		analysis := &entity.EnergyEfficiencyAnalysis{
			ID:             uuid.New().String(),
			AnalysisTime:   time.Now(),
			Type:           entity.EnergyEfficiencyTypeDevice,
			TargetID:       "device-001",
			TargetName:     "Device 1",
			TimeRangeStart: time.Now().Add(-24 * time.Hour),
			TimeRangeEnd:   time.Now(),
			AvgEfficiency:  0.88,
			MaxEfficiency:  0.95,
			MinEfficiency:  0.80,
			Trend:          "stable",
		}

		err := repo.CreateAnalysis(ctx, analysis)
		require.NoError(t, err)

		found, err := repo.GetAnalysisByID(ctx, analysis.ID)
		require.NoError(t, err)
		assert.InDelta(t, 0.88, found.AvgEfficiency, 0.01)
	})

	t.Run("ListAnalyses", func(t *testing.T) {
		eeType := entity.EnergyEfficiencyTypeDevice
		_, _, err := repo.ListAnalyses(ctx, &repository.EnergyEfficiencyAnalysisQuery{
			Type:     &eeType,
			Page:     1,
			PageSize: 10,
		})
		require.NoError(t, err)
	})

	t.Run("GetStatistics", func(t *testing.T) {
		stats, err := repo.GetStatistics(ctx, "device-001", entity.EnergyEfficiencyTypeDevice, "hourly", time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour))
		require.NoError(t, err)
		assert.NotNil(t, stats)
		assert.GreaterOrEqual(t, stats.TotalRecords, int64(1))
	})

	t.Run("GetBenchmark", func(t *testing.T) {
		benchmark, err := repo.GetBenchmark(ctx, entity.EnergyEfficiencyTypeDevice, "device-001")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, benchmark, 0.0)
	})
}

func TestAlarmRepository_CRUD(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	setupAllTables(db)

	repo := NewAlarmRepository(db)
	ctx := context.Background()

	alarm := entity.NewAlarm("point-001", "device-001", "station-001", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test Alarm", "Test message")
	alarm.ID = uuid.New().String()

	t.Run("Create", func(t *testing.T) {
		err := repo.Create(ctx, alarm)
		require.NoError(t, err)
	})

	t.Run("GetByID", func(t *testing.T) {
		found, err := repo.GetByID(ctx, alarm.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "Test Alarm", found.Title)
		assert.True(t, found.IsActive())
	})

	t.Run("GetActiveAlarms", func(t *testing.T) {
		alarms, err := repo.GetActiveAlarms(ctx, nil, nil)
		require.NoError(t, err)
		assert.Len(t, alarms, 1)
	})

	t.Run("Acknowledge", func(t *testing.T) {
		err := repo.Acknowledge(ctx, alarm.ID, "admin")
		require.NoError(t, err)

		found, _ := repo.GetByID(ctx, alarm.ID)
		assert.True(t, found.IsAcknowledged())
	})

	t.Run("Clear", func(t *testing.T) {
		err := repo.Clear(ctx, alarm.ID)
		require.NoError(t, err)

		found, _ := repo.GetByID(ctx, alarm.ID)
		assert.True(t, found.IsCleared())
	})

	t.Run("CountByLevel", func(t *testing.T) {
		counts, err := repo.CountByLevel(ctx, nil)
		require.NoError(t, err)
		assert.NotNil(t, counts)
	})
}

func TestProviderFunctions(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	assert.NotNil(t, NewAlarmRuleRepository(db))
	assert.NotNil(t, NewDeviceRepository(db))
	assert.NotNil(t, NewUserRepository(db))
	assert.NotNil(t, NewSystemConfigRepository(db))
	assert.NotNil(t, NewQARepository(db))
	assert.NotNil(t, NewOperationLogRepository(db))
	assert.NotNil(t, NewNotificationConfigRepository(db))
	assert.NotNil(t, NewEnergyEfficiencyRepository(db))
	assert.NotNil(t, NewWorkOrderRepository(db))
	assert.NotNil(t, NewInventoryRepository(db))
	assert.NotNil(t, NewInventoryTransactionRepository(db))
	assert.NotNil(t, NewSupplierRepository(db))
	assert.NotNil(t, NewPurchaseOrderRepository(db))
	assert.NotNil(t, NewReceiptRepository(db))
	assert.NotNil(t, NewCostCategoryRepository(db))
	assert.NotNil(t, NewCostEntryRepository(db))
	assert.NotNil(t, NewCostAllocationRepository(db))
	assert.NotNil(t, NewCostReportRepository(db))
	assert.NotNil(t, NewAssetRepository(db))
	assert.NotNil(t, NewAssetMaintenanceRepository(db))
	assert.NotNil(t, NewAssetDepreciationRepository(db))
	assert.NotNil(t, NewAssetDocumentRepository(db))
	assert.NotNil(t, NewForecastResultRepository(db))
	assert.NotNil(t, NewFaultDetectionResultRepository(db))
	assert.NotNil(t, NewEdgeNodeRepository(db))
	assert.NotNil(t, NewModelVersionRepository(db))
	assert.NotNil(t, NewStationRepository(db))
	assert.NotNil(t, NewRegionRepository(db))
	assert.NotNil(t, NewSubRegionRepository(db))
	assert.NotNil(t, NewPointRepository(db))
	assert.NotNil(t, NewAlarmRepository(db))
	assert.NotNil(t, NewReportRepository(db))
	assert.NotNil(t, NewRoleRepository(db))
	assert.NotNil(t, NewPermissionRepository(db))
	assert.NotNil(t, NewCarbonEmissionRepository(db))
}

func strPtr(s string) *string { return &s }
