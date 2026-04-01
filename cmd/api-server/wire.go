//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/new-energy-monitoring/internal/api/handler"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/infrastructure/persistence"
)

// InitializeApp 初始化应用程序
// Wire 会自动生成这个函数的实现
func InitializeApp() (*App, error) {
	wire.Build(
		// 基础设施层
		NewConfig,
		NewLogger,
		NewDatabase,
		NewRedis,
		NewKafka,
		NewJWTManager,
		NewPasswordManager,

		// 仓储层
		persistence.RepositorySet,

		// 服务层
		service.ServiceSet,

		// 处理器层
		handler.HandlerSet,

		// HTTP 服务器
		NewHTTPServer,

		// 应用程序
		NewApp,
	)
	return nil, nil
}
