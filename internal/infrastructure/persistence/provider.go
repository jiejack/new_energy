package persistence

import (
	"github.com/google/wire"
)

// RepositorySet 仓储层 Provider Set
// 包含所有仓储的实现
var RepositorySet = wire.NewSet(
	NewUserRepository,
	NewRegionRepository,
	NewStationRepository,
	NewDeviceRepository,
	NewPointRepository,
	NewAlarmRepository,
	NewAlarmRuleRepository,
	NewRoleRepository,
	NewPermissionRepository,
	NewOperationLogRepository,
	NewQARepository,
	NewSystemConfigRepository,
	NewNotificationConfigRepository,
	NewEnergyEfficiencyRepository,
)
