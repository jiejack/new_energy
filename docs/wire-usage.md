# Wire 依赖注入框架使用说明

## 概述

本项目使用 Google Wire 实现依赖注入框架，提供编译时依赖检查和自动代码生成功能。

## 文件结构

```
new-energy-monitoring/
├── cmd/api-server/
│   ├── main.go              # 原始入口文件（保留用于兼容）
│   ├── main_wire.go         # 使用 Wire 的新入口文件
│   ├── wire.go              # Wire 配置文件
│   ├── wire_gen.go          # Wire 生成的代码（自动生成）
│   └── app.go               # App 结构体和基础设施 Provider
├── internal/
│   ├── api/handler/
│   │   └── provider.go      # 处理器层 Provider Set
│   ├── application/service/
│   │   └── provider.go      # 服务层 Provider Set
│   ├── infrastructure/persistence/
│   │   └── provider.go      # 仓储层 Provider Set
│   └── domain/
│       ├── cache/
│       │   └── cache.go     # 缓存接口定义
│       └── logger/
│           └── logger.go    # 日志接口定义
└── Makefile                 # 包含 Wire 生成脚本
```

## 使用方法

### 1. 生成依赖注入代码

```bash
# 生成 Wire 代码
make wire

# 或者直接运行
cd cmd/api-server && wire
```

### 2. 检查依赖完整性

```bash
# 检查依赖图是否完整
make wire-check

# 或者直接运行
cd cmd/api-server && wire check
```

### 3. 生成依赖图可视化

```bash
# 生成依赖图（DOT 格式）
make wire-graph

# 使用 Graphviz 可视化
dot -Tpng cmd/api-server/wire_graph.dot -o wire_graph.png
```

### 4. 构建和运行

```bash
# 构建项目（会自动运行 wire）
make build

# 运行 API 服务器（会自动运行 wire）
make run-api
```

## Provider Set 说明

### 基础设施层 Provider

在 `cmd/api-server/app.go` 中定义：

- `NewConfig`: 创建配置实例
- `NewLogger`: 创建日志实例
- `NewDatabase`: 创建数据库实例
- `NewRedis`: 创建 Redis 实例
- `NewKafka`: 创建 Kafka 实例
- `NewJWTManager`: 创建 JWT 管理器
- `NewPasswordManager`: 创建密码管理器
- `NewHTTPServer`: 创建 HTTP 服务器

### 仓储层 Provider Set

在 `internal/infrastructure/persistence/provider.go` 中定义：

```go
var repositorySet = wire.NewSet(
    NewUserRepository,
    NewRegionRepository,
    NewStationRepository,
    NewDeviceRepository,
    NewPointRepository,
    NewAlarmRepository,
    NewRoleRepository,
    NewPermissionRepository,
    NewOperationLogRepository,
    NewConfigRepository,
)
```

### 服务层 Provider Set

在 `internal/application/service/provider.go` 中定义：

```go
var serviceSet = wire.NewSet(
    NewAuthService,
    NewUserService,
    NewDeviceService,
    NewAlarmService,
    NewStationService,
    NewRegionService,
    NewPointService,
    NewPermissionService,
    NewAuditService,
)
```

### 处理器层 Provider Set

在 `internal/api/handler/provider.go` 中定义：

```go
var handlerSet = wire.NewSet(
    NewAuthHandler,
    NewUserHandler,
    NewDeviceHandler,
    NewAlarmHandler,
    NewStationHandler,
    NewRegionHandler,
    NewPointHandler,
)
```

## 添加新组件

### 1. 添加新的服务

在 `internal/application/service/` 中创建新的服务文件：

```go
// example_service.go
package service

type ExampleService struct {
    exampleRepo repository.ExampleRepository
    logger      logger.Logger
}

func NewExampleService(
    exampleRepo repository.ExampleRepository,
    logger logger.Logger,
) *ExampleService {
    return &ExampleService{
        exampleRepo: exampleRepo,
        logger:      logger,
    }
}
```

然后在 `provider.go` 中添加：

```go
var serviceSet = wire.NewSet(
    // ... 其他服务
    NewExampleService,  // 添加新服务
)
```

### 2. 添加新的处理器

在 `internal/api/handler/` 中创建新的处理器文件：

```go
// example_handler.go
package handler

type ExampleHandler struct {
    exampleService *service.ExampleService
}

func NewExampleHandler(exampleService *service.ExampleService) *ExampleHandler {
    return &ExampleHandler{
        exampleService: exampleService,
    }
}
```

然后在 `provider.go` 中添加：

```go
var handlerSet = wire.NewSet(
    // ... 其他处理器
    NewExampleHandler,  // 添加新处理器
)
```

### 3. 更新 Wire 配置

如果需要将新组件注入到其他组件中，需要在 `cmd/api-server/app.go` 中更新相应的 Provider 函数。

## Mock 注入支持

Wire 支持 Mock 注入，便于测试。在测试文件中：

```go
// example_service_test.go
package service_test

import (
    "testing"
    "github.com/google/wire"
    "github.com/stretchr/testify/mock"
)

// MockExampleRepository Mock 仓储
type MockExampleRepository struct {
    mock.Mock
}

// 实现 repository.ExampleRepository 接口...

// TestExampleService 测试示例
func TestExampleService(t *testing.T) {
    // 使用 Wire 创建测试依赖
    testSet := wire.NewSet(
        NewMockExampleRepository,
        NewExampleService,
    )
    
    // ... 测试代码
}
```

## 技术优势

1. **编译时检查**: Wire 在编译时检查依赖完整性，避免运行时错误
2. **自动代码生成**: 自动生成依赖注入代码，减少手动维护
3. **依赖图可视化**: 支持生成依赖图，便于理解和调试
4. **类型安全**: 完全类型安全，编译器会检查类型匹配
5. **性能优化**: 生成的代码性能接近手动编写的代码

## 常见问题

### 1. Wire 生成失败

**问题**: 运行 `make wire` 时提示依赖缺失

**解决**: 确保所有依赖都已正确导入，检查 `go.mod` 文件

### 2. 循环依赖

**问题**: Wire 检测到循环依赖

**解决**: 重构代码，使用接口解耦，或者使用 `wire.Bind` 绑定接口

### 3. 类型不匹配

**问题**: Wire 提示类型不匹配

**解决**: 检查 Provider 函数的返回类型和依赖注入目标的参数类型是否一致

## 参考资料

- [Wire 官方文档](https://github.com/google/wire)
- [Wire 用户指南](https://github.com/google/wire/blob/main/docs/guide.md)
- [Wire 最佳实践](https://github.com/google/wire/blob/main/docs/best-practices.md)
