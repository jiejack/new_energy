package handler

import (
	"github.com/google/wire"
)

// HandlerSet 处理器层 Provider Set
// 包含所有 HTTP 处理器的实现
var HandlerSet = wire.NewSet(
	NewAuthHandler,
	NewUserHandler,
	NewDeviceHandler,
	NewAlarmHandler,
	NewAlarmRuleHandler,
	NewStationHandler,
	NewRegionHandler,
	NewPointHandler,
	NewQAHandler,
	NewConfigHandler,
	NewNotificationConfigHandler,
	NewExportHandler,
	NewOperationLogHandler,
	NewReportHandler,
	NewEnergyEfficiencyHandler,
	NewCarbonEmissionHandler,
)
