package service

import (
	"github.com/google/wire"
)

// ServiceSet 服务层 Provider Set
// 包含所有服务的实现
var ServiceSet = wire.NewSet(
	NewAuthService,
	NewUserService,
	NewDeviceService,
	NewAlarmService,
	NewStationService,
	NewRegionService,
	NewPointService,
	NewPermissionService,
	NewAuditService,
	NewQAService,
	NewConfigService,
)
