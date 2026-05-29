# NEM 新能源监控系统 - 核心功能测试方案

| 项目 | 内容 |
|------|------|
| 文档版本 | v1.0 |
| 编写日期 | 2026-05-29 |
| 适用范围 | NEM 核心功能全模块测试 |
| 项目代号 | nem-core-features |

---

## 1. 测试概述

### 1.1 测试目标

1. **功能验证**：验证告警规则管理、统计报表、数据导出等核心功能模块的正确性与完整性
2. **质量保障**：确保代码覆盖率达到 ≥80% 的目标，消除 0% 覆盖率的包
3. **缺陷修复**：定位并修复已知缺陷（DryRun 无限循环、GORM autoCreateTime 覆盖、Flaky 测试）
4. **性能达标**：API 响应时间 P95 < 200ms，报表生成 < 10s（10 万条数据内）
5. **安全合规**：通过输入验证、SQL 注入防护、XSS 防护等安全测试

### 1.2 测试范围

| 层级 | 模块 | 包路径 | 当前状态 |
|------|------|--------|----------|
| 领域层 | 实体与业务规则 | `internal/domain/entity` | 部分覆盖 |
| 领域层 | 仓储接口 | `internal/domain/repository` | 接口定义，无测试 |
| 领域层 | 缓存接口 | `internal/domain/cache` | 无测试 |
| 领域层 | 日志接口 | `internal/domain/logger` | 无测试 |
| 应用层 | 业务服务 | `internal/application/service` | 大部分覆盖 |
| 应用层 | 告警 Harness | `internal/application/service` (alarm_harness) | 已覆盖 |
| 基础设施层 | 持久化仓储 | `internal/infrastructure/persistence` | 部分覆盖 |
| 基础设施层 | Redis 缓存 | `internal/infrastructure/cache` | 已覆盖 |
| 基础设施层 | 配置管理 | `internal/infrastructure/config` | 已覆盖 |
| 基础设施层 | 消息队列 | `internal/infrastructure/mq` | 已覆盖 |
| 基础设施层 | 日志实现 | `internal/infrastructure/logger` | 已覆盖 |
| 接口层 | HTTP Handler | `internal/api/handler` | 部分覆盖 |
| 接口层 | DTO | `internal/api/dto` | 无测试 |
| 核心包 | 告警引擎 | `pkg/alarm/*` | 大部分覆盖 |
| 核心包 | AI 模块 | `pkg/ai/*` | 大部分覆盖 |
| 核心包 | 协议解析 | `pkg/protocol/*` | 已覆盖 |
| 核心包 | 数据采集 | `pkg/collector` | 已覆盖 |
| 核心包 | 计算引擎 | `pkg/compute/*` | 已覆盖 |
| 核心包 | 存储引擎 | `pkg/storage/*` | 已覆盖 |
| 核心包 | 监控告警 | `pkg/monitoring/*` | 已覆盖 |
| 核心包 | Harness 框架 | `pkg/harness` | 已覆盖 |
| 核心包 | 导出功能 | `pkg/export` | 已覆盖 |
| 核心包 | 认证授权 | `pkg/auth` | 已覆盖 |
| 核心包 | 缓存封装 | `pkg/cache` | 已覆盖 |
| 核心包 | 大数据模块 | `pkg/bigdata` | 已覆盖 |
| 核心包 | WebSocket | `pkg/websocket` | 已覆盖 |
| 核心包 | Nacos 集成 | `pkg/nacos` | 部分覆盖 |
| 核心包 | 错误处理 | `pkg/errors` | 已覆盖 |
| 核心包 | 反馈模块 | `pkg/feedback` | 已覆盖 |
| 核心包 | 数据处理 | `pkg/processor` | 已覆盖 |
| 核心包 | 统计计算 | `pkg/statistics/*` | 已覆盖 |
| 核心包 | 配置加载 | `pkg/config` | 部分覆盖 |
| 前端 | Vue 组件 | `web/src/*` | 需补充 |
| E2E | 端到端流程 | `web/e2e/*` | 基础覆盖 |

### 1.3 测试排除范围

- 移动端响应式适配测试（后续迭代）
- 第三方系统集成测试（ERP/CRM 对接）
- AI 模型训练精度验证（非软件测试范畴）
- ClickHouse / Doris 生产环境性能测试（需独立环境）

---

## 2. 测试策略

### 2.1 测试金字塔

```
                    ┌─────────┐
                    │  E2E 测试 │  ← 少量，覆盖核心业务流程
                   ┌┴─────────┴┐
                   │ 集成测试    │  ← 适度，验证组件间协作
                  ┌┴───────────┴┐
                  │   单元测试    │  ← 大量，覆盖所有业务逻辑
                 └──────────────┘
```

| 测试类型 | 占比 | 目标 | 执行频率 |
|----------|------|------|----------|
| 单元测试 | 70% | 函数/方法级别逻辑验证 | 每次提交 |
| 集成测试 | 20% | 模块间交互与数据库操作验证 | 每次合并 |
| E2E 测试 | 10% | 端到端业务流程验证 | 每日构建 |
| 性能测试 | 专项 | 响应时间与吞吐量验证 | 每周 / 版本发布前 |
| 安全测试 | 专项 | 安全漏洞与合规验证 | 每月 / 版本发布前 |

### 2.2 单元测试策略

- **原则**：每个导出函数/方法至少一个测试用例，包含正常路径和异常路径
- **隔离**：使用 `testify/mock` 和手写 Mock（`tests/helpers/`）隔离外部依赖
- **命名**：`Test<函数名>_<场景>_<预期结果>`，如 `TestCreateAlarmRule_ValidInput_Success`
- **表驱动**：优先使用表驱动测试（table-driven tests）覆盖多场景
- **边界值**：必须覆盖零值、空值、极大值、边界条件

### 2.3 集成测试策略

- **数据库集成**：使用 SQLite 内存数据库模拟 PostgreSQL，关键路径补充 PostgreSQL 真实测试
- **缓存集成**：使用 miniredis 模拟 Redis 服务端
- **消息队列集成**：使用 `tests/helpers/mock_kafka.go` 模拟 Kafka
- **API 集成**：使用 `httptest` 包模拟 HTTP 请求/响应

### 2.4 E2E 测试策略

- **工具**：Playwright（Chromium 优先）
- **覆盖场景**：登录认证、告警管理、电站监控、数据导出、配置管理
- **环境**：CI 中启动完整后端服务 + 前端构建产物

### 2.5 性能测试策略

- **Benchmark**：Go 标准 `testing.B` 基准测试，覆盖核心算法和热点路径
- **压力测试**：k6 / 自定义脚本模拟并发请求
- **内存泄漏**：pprof 长时间运行检测

### 2.6 安全测试策略

- **静态扫描**：Trivy 容器扫描 + Gosec 代码安全扫描
- **动态测试**：SQL 注入、XSS、CSRF、越权访问等 OWASP Top 10 场景
- **认证授权**：JWT 令牌安全、密码策略、权限边界

---

## 3. 测试环境

### 3.1 环境矩阵

| 环境 | 用途 | 数据库 | 缓存 | 消息队列 |
|------|------|--------|------|----------|
| 本地开发 | 日常开发测试 | SQLite 内存 / 本地 PostgreSQL | miniredis / 本地 Redis | MockKafka |
| CI 流水线 | 自动化测试 | GitHub Actions PostgreSQL 15 服务容器 | GitHub Actions Redis 7 服务容器 | Mock（无 Kafka 容器） |
| 集成测试环境 | 全链路集成 | PostgreSQL 15 (Docker) | Redis 7 (Docker) | Kafka (Docker) |
| 性能测试环境 | 压力基准 | 独立 PostgreSQL 实例 | 独立 Redis 集群 | 独立 Kafka 集群 |

### 3.2 测试框架与工具

| 类别 | 工具 | 版本 | 用途 |
|------|------|------|------|
| Go 测试框架 | `testing` (标准库) | Go 1.24+ | 测试入口与 Benchmark |
| 断言库 | `testify/assert` | v1.11.1 | 流式断言 |
| Mock 库 | `testify/mock` | v1.11.1 | 接口 Mock |
| 内存数据库 | `gorm.io/driver/sqlite` | v1.6.0 | 持久化层单元测试 |
| 内存 Redis | `github.com/alicebob/miniredis/v2` | v2.38.0 | 缓存层单元测试 |
| HTTP 测试 | `net/http/httptest` | 标准库 | Handler 层测试 |
| 前端单元测试 | Vitest | - | Vue 组件与工具函数测试 |
| E2E 测试 | Playwright | - | 端到端流程测试 |
| 代码检查 | golangci-lint | latest | 静态代码分析 |
| 安全扫描 | Trivy / Gosec | latest | 安全漏洞扫描 |
| 性能分析 | pprof | 标准库 | CPU / 内存性能分析 |
| 覆盖率 | `go test -cover` | 标准库 | 覆盖率收集 |
| 覆盖率上传 | Codecov | - | 覆盖率可视化与趋势追踪 |

### 3.3 测试配置

测试配置文件位于 `configs/config-test.yaml`，关键配置项：

```yaml
server:
  mode: release          # 测试环境使用 release 模式
database:
  type: postgres
  host: ${DB_HOST:postgres}
  port: ${DB_PORT:5432}
  dbname: ${DB_NAME:nem_system_test}
redis:
  addrs:
    - ${REDIS_HOST:redis}:${REDIS_PORT:6379}
kafka:
  brokers:
    - ${KAFKA_HOST:kafka}:${KAFKA_PORT:9092}
  topic_prefix: nem_test
auth:
  jwt:
    access_expire: 7200
    refresh_expire: 604800
  password:
    min_length: 8
    require_uppercase: true
    require_lowercase: true
    require_digit: true
```

### 3.4 测试辅助工具

项目已有的测试辅助工具位于 `tests/helpers/`：

| 文件 | 功能 |
|------|------|
| `mock_db.go` | 内存模拟数据库，支持 User/Role/Permission/Region/Station/Device/Alarm/Point 实体 CRUD |
| `mock_redis.go` | 内存模拟 Redis，支持 String/List/Set/Hash 操作及 TTL |
| `mock_kafka.go` | 内存模拟 Kafka，支持 Produce/Consume/Subscribe |
| `test_utils.go` | 测试工具函数（JWT/密码管理器/HTTP 请求构造/断言辅助/实体工厂） |

`tests/test_config.go` 提供：

| 组件 | 功能 |
|------|------|
| `TestConfig` | 测试环境配置结构体 |
| `TestDatabase` | 测试数据库连接与迁移管理 |
| `TestSuite` | 测试套件（DB + JWT + Password + Config） |
| `SetupTestUser/Role/Permission` | 测试数据工厂方法 |
| `GenerateTestToken` | 测试 JWT Token 生成 |

---

## 4. 单元测试方案

### 4.1 领域层（Domain Layer）

#### 4.1.1 实体测试 `internal/domain/entity`

| 实体 | 测试文件 | 关键测试点 | 目标覆盖率 |
|------|----------|-----------|-----------|
| User | `user_test.go` | 创建/状态变更/密码验证/角色分配 | ≥90% |
| Role | `role_test.go` | 创建/权限管理/状态变更 | ≥90% |
| Permission | `permission_test.go` | 创建/资源类型/操作类型 | ≥90% |
| Region | `region_test.go` | 创建/层级关系/子区域 | ≥85% |
| Station | `station_test.go` | 创建/类型枚举/状态管理 | ≥85% |
| Device | `device_test.go` | 创建/类型枚举/状态管理/关联 | ≥85% |
| Point | `point_test.go` | 创建/类型枚举/数据类型 | ≥85% |
| Alarm | `alarm_test.go` | 创建/级别/状态流转/确认/清除 | ≥90% |
| AlarmRule | `alarm_rule_test.go` | 创建/规则类型/触发条件/阈值 | ≥90% |
| OperationLog | `operation_log_test.go` | 创建/操作类型/审计追踪 | ≥85% |
| NotificationConfig | `notification_config_test.go` | 创建/渠道类型/配置验证 | ≥85% |
| SystemConfig | `system_config_test.go` | 创建/配置键值/类型验证 | ≥85% |
| QA | `qa_test.go` | 会话创建/消息角色/状态流转 | ≥80% |

**已知缺陷关注点**：

- `qa.go` 中 `CreatedAt` 使用 `gorm:"autoCreateTime"` 标签，在测试中手动设置 `CreatedAt` 时会被 GORM 覆盖。测试方案：
  - 使用 `gorm:"<-:create"` 替代或测试中通过 GORM 创建后验证时间字段非零
  - 添加专项测试验证 `autoCreateTime` 在 SQLite 和 PostgreSQL 下的行为一致性

#### 4.1.2 仓储接口测试 `internal/domain/repository`

当前仓储接口仅有定义，无直接测试。测试策略：

- 仓储接口测试通过具体实现（`internal/infrastructure/persistence`）间接覆盖
- 接口契约测试：验证所有实现均满足接口语义

#### 4.1.3 缓存/日志接口 `internal/domain/cache`, `internal/domain/logger`

- 缓存接口通过 `internal/infrastructure/cache` 实现测试覆盖
- 日志接口通过 `internal/infrastructure/logger` 实现测试覆盖

### 4.2 应用层（Application Layer）

#### 4.2.1 服务层测试 `internal/application/service`

| 服务 | 测试文件 | 关键测试点 | 目标覆盖率 |
|------|----------|-----------|-----------|
| AlarmService | `alarm_service_test.go` | 告警创建/查询/确认/清除/统计 | ≥85% |
| AlarmRuleService | `alarm_rule_service_test.go` | 规则 CRUD/条件验证/启用禁用 | ≥85% |
| AlarmServiceWithHarness | `alarm_service_with_harness_test.go` | Harness 集成/约束验证/快照 | ≥80% |
| AlarmHarness | `alarm_harness_test.go` | 告警约束/验证/监控 | ≥80% |
| DeviceService | `device_service_test.go` | 设备 CRUD/状态管理/批量操作 | ≥85% |
| DeviceHarness | `device_harness_test.go` | 设备约束/验证 | ≥80% |
| UserService | `user_service_test.go` | 用户 CRUD/认证/角色分配 | ≥85% |
| AuthService | `auth_service_test.go` | 登录/注册/Token 刷新/密码重置 | ≥90% |
| PermissionService | `permission_service_test.go` | 权限校验/角色权限/资源访问控制 | ≥90% |
| StationService | `station_service_test.go` | 电站 CRUD/类型管理/统计 | ≥85% |
| RegionService | `region_service_test.go` | 区域 CRUD/层级管理 | ≥85% |
| ConfigService | `config_service_test.go` | 配置 CRUD/热更新/验证 | ≥85% |
| CacheService | `cache_service_test.go` | 缓存读写/失效/预热 | ≥80% |
| ExportService | `export_service_test.go` | Excel/CSV 导出/大数据量 | ≥80% |
| ReportService | `report_service_test.go` | 报表生成/多维度统计 | ≥80% |
| ForecastService | `forecast_service_test.go` | 预测计算/精度评估 | ≥80% |
| FaultService | `fault_service_test.go` | 故障检测/分类/评估 | ≥80% |
| QAService | `qa_service_test.go` | 问答会话/意图识别/对话管理 | ≥80% |
| ModelService | `model_service_test.go` | 模型版本管理/部署 | ≥80% |
| EdgeService | `edge_service_test.go` | 边缘节点管理/心跳/同步 | ≥80% |
| NotificationConfigService | `notification_config_service_test.go` | 通知配置 CRUD/渠道验证 | ≥80% |
| OperationLogService | `operation_log_service_test.go` | 操作日志记录/查询/审计 | ≥80% |
| AuditService | `audit_service_test.go` | 审计追踪/合规检查 | ≥80% |
| PointService | `point_service_test.go` | 采集点管理/数据类型 | ≥80% |

**测试模式**：

```go
func TestAlarmRuleService_Create_ValidInput_Success(t *testing.T) {
    mockRepo := new(mock.AlarmRuleRepository)
    svc := NewAlarmRuleService(mockRepo)
    mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

    rule, err := svc.CreateAlarmRule(ctx, "Test Rule", AlarmRuleTypeLimit, AlarmLevelWarning, "value > 90")

    assert.NoError(t, err)
    assert.NotNil(t, rule)
    mockRepo.AssertCalled(t, "Create", ctx, mock.AnythingOfType("*entity.AlarmRule"))
}
```

### 4.3 基础设施层（Infrastructure Layer）

#### 4.3.1 持久化仓储测试 `internal/infrastructure/persistence`

| 仓储 | 测试文件 | 测试策略 | 目标覆盖率 |
|------|----------|----------|-----------|
| Database | `database_test.go` | SQLite 内存数据库，Ping/HealthCheck/Close | ≥85% |
| Migration | `migration_test.go` | SQLite 内存数据库，迁移执行/回滚 | ≥80% |
| UserRepository | `user_repository_test.go` | SQLite 内存数据库，CRUD/查询/关联 | ≥85% |
| AlarmRuleRepository | `alarm_rule_repository_test.go` | SQLite 内存数据库，CRUD/条件查询 | ≥85% |
| DeviceRepository | `device_repository_test.go` | SQLite 内存数据库，CRUD/关联查询 | ≥85% |
| QARepository | `qa_repository_test.go` | SQLite 内存数据库，CRUD/会话消息 | ≥80% |
| SystemConfigRepository | `system_config_repository_test.go` | SQLite 内存数据库，CRUD/键值查询 | ≥80% |
| OperationLogRepository | `operation_log_repository_test.go` | SQLite 内存数据库，CRUD/时间范围查询 | ≥80% |
| NotificationConfigRepository | `notification_config_repository_test.go` | SQLite 内存数据库，CRUD | ≥80% |
| 其他仓储 | `additional_repository_test.go` | SQLite 内存数据库 | ≥75% |

**SQLite vs PostgreSQL 差异注意**：

- SQLite 不支持 `RETURNING` 子句，需验证 GORM 在两种数据库下的行为
- SQLite 的 `autoCreateTime` 行为与 PostgreSQL 可能不一致，需专项测试
- 事务隔离级别差异：SQLite 默认 SERIALIZABLE，PostgreSQL 默认 READ COMMITTED

#### 4.3.2 缓存层测试 `internal/infrastructure/cache`

| 组件 | 测试文件 | 测试策略 | 目标覆盖率 |
|------|----------|----------|-----------|
| RedisClient | `redis_test.go` | miniredis 内存 Redis，连接/Ping/Get/Set/Delete | ≥85% |
| Middleware | `middleware_test.go` | 缓存中间件，命中/未命中/过期 | ≥80% |

#### 4.3.3 消息队列测试 `internal/infrastructure/mq`

| 组件 | 测试文件 | 测试策略 | 目标覆盖率 |
|------|----------|----------|-----------|
| KafkaClient | `kafka_test.go` | MockKafka，生产/消费/订阅/错误处理 | ≥80% |

#### 4.3.4 配置层测试 `internal/infrastructure/config`

| 组件 | 测试文件 | 测试策略 | 目标覆盖率 |
|------|----------|----------|-----------|
| Config | `config_test.go` | Viper 配置加载/环境变量覆盖/默认值 | ≥85% |
| RedisConfig | `redis_config_test.go` | Redis 配置解析/验证 | ≥85% |

### 4.4 接口层（API Handler Layer）

#### 4.4.1 Handler 测试 `internal/api/handler`

| Handler | 测试文件 | 关键测试点 | 目标覆盖率 |
|---------|----------|-----------|-----------|
| AlarmHandler | `alarm_handler_test.go` | 告警列表/详情/确认/清除/统计 | ≥85% |
| AlarmRuleHandler | `alarm_rule_handler_test.go` | 规则 CRUD/参数验证/错误处理 | ≥85% |
| AuthHandler | `auth_handler_test.go` | 登录/注册/Token 刷新/权限校验 | ≥90% |
| UserHandler | `user_handler_test.go` | 用户 CRUD/角色分配/密码修改 | ≥85% |
| StationHandler | `station_handler_test.go` | 电站 CRUD/类型筛选/统计 | ≥85% |
| DeviceHandler | `device_handler_test.go` | 设备 CRUD/状态管理/批量操作 | ≥85% |
| EdgeHandler | `edge_handler_test.go` | 边缘节点管理/心跳/同步 | ≥80% |
| ConfigHandler | `config_handler_test.go` | 配置 CRUD/热更新 | ≥80% |
| ExportHandler | `export_handler_test.go` | Excel/CSV 导出/参数验证 | ≥80% |
| ReportHandler | `report_handler_test.go` | 报表生成/多维度查询 | ≥80% |
| ForecastHandler | `forecast_handler_test.go` | 预测查询/精度评估 | ≥80% |
| FaultHandler | `fault_handler_test.go` | 故障检测/分类查询 | ≥80% |
| QAHandler | `qa_handler_test.go` | 问答会话/消息发送 | ≥80% |
| ModelHandler | `model_handler_test.go` | 模型版本管理 | ≥80% |
| NotificationConfigHandler | `notification_config_handler_test.go` | 通知配置 CRUD | ≥80% |
| OperationLogHandler | `operation_log_handler_test.go` | 操作日志查询 | ≥80% |

#### 4.4.2 DTO 测试 `internal/api/dto`

当前 `request.go` 和 `response.go` 无测试，需补充：

| DTO | 测试文件 | 关键测试点 | 目标覆盖率 |
|-----|----------|-----------|-----------|
| Request DTOs | `dto_test.go`（新增） | 参数绑定/验证标签/默认值 | ≥80% |
| Response DTOs | `dto_test.go`（新增） | 序列化/字段映射/空值处理 | ≥80% |

### 4.5 核心包（pkg）单元测试

#### 4.5.1 告警引擎 `pkg/alarm`

| 子包 | 测试文件 | 关键测试点 | 目标覆盖率 |
|------|----------|-----------|-----------|
| detector | `detector_test.go` | 阈值检测/趋势检测/复合条件 | ≥85% |
| rule | `rule_test.go` | DSL 解析/规则引擎/版本管理 | ≥85% |
| notifier | `notifier_test.go` | 邮件/短信/内部通知/模板渲染 | ≥80% |
| aggregator | `aggregator_test.go` | 告警聚合/去重/时间窗口 | ≥85% |
| dedup | `dedup_test.go` | 去重策略/指纹计算/过期清理 | ≥85% |
| state | `state_machine_test.go` | 状态流转/转换验证/非法转换 | ≥90% |
| storage | `storage_test.go` | 告警存储/查询/分页 | ≥80% |

#### 4.5.2 AI 模块 `pkg/ai`

| 子包 | 测试文件 | 关键测试点 | 目标覆盖率 |
|------|----------|-----------|-----------|
| config | `config_test.go` | AI 配置加载/助手配置 | ≥80% |
| datacollector | `datacollector_test.go` | 数据采集/验证/清洗/导入 | ≥80% |
| forecast | `forecast_test.go` | 光伏/风电/储能预测/精度评估 | ≥80% |
| fault | `fault_test.go` | 故障检测/分类/评估/RUL 预测 | ≥80% |
| inference | `service_test.go` | 推理服务/模型管理/缓存 | ≥80% |
| knowledge | `knowledge_test.go` | RAG/向量检索/嵌入 | ≥80% |
| operation | `operation_test.go` | 操作解析/确认/执行 | ≥80% |
| qa | `qa_test.go` | 意图识别/对话管理/生成 | ≥80% |
| service | `service_test.go` | AI 服务适配/上下文管理 | ≥80% |
| features | `features_test.go` | 特征提取/管道/时间/滞后 | ≥80% |
| edge | `edge_test.go` | 边缘推理/模型服务/心跳 | ≥80% |

#### 4.5.3 协议解析 `pkg/protocol`

| 子包 | 测试文件 | 关键测试点 | 目标覆盖率 |
|------|----------|-----------|-----------|
| modbus | `modbus_test.go` | RTU/TCP/ASCII 帧解析/CRC/功能码 | ≥85% |
| iec104 | `iec104_test.go` | ASDU 解析/连接管理/类型映射 | ≥85% |
| iec61850 | `iec61850_test.go` | MMS 客户端/报告/控制/采样值/模型 | ≥80% |

#### 4.5.4 其他核心包

| 包 | 测试文件 | 关键测试点 | 目标覆盖率 |
|----|----------|-----------|-----------|
| `pkg/compute/formula` | `formula_test.go` | 公式解析/执行/函数/运算符 | ≥85% |
| `pkg/compute/rule` | `rule_test.go` | 规则引擎/触发器/调度 | ≥80% |
| `pkg/collector` | `collector_test.go` | 数据采集/缓冲池/调度 | ≥80% |
| `pkg/storage/*` | 各子包测试 | 压缩/索引/分区/生命周期/查询/分片/时序 | ≥80% |
| `pkg/monitoring/*` | 各子包测试 | 告警/仪表盘/健康/指标/追踪 | ≥80% |
| `pkg/harness` | 各组件测试 | 验证器/校验器/约束/监控/快照 | ≥85% |
| `pkg/export` | `excel_test.go`, `csv_test.go` | Excel/CSV 生成/大数据量/编码 | ≥80% |
| `pkg/auth` | `auth_test.go` | JWT 生成/验证/刷新/密码哈希 | ≥90% |
| `pkg/cache` | `cache_test.go` | 缓存封装/策略/TTL | ≥80% |
| `pkg/errors` | `errors_test.go` | 错误码/错误链/响应格式 | ≥85% |
| `pkg/processor` | `processor_test.go` | 变更检测/过滤/管道/质量/缩放/验证 | ≥80% |
| `pkg/statistics/*` | 各子包测试 | 统计计算/调度/分布式 | ≥80% |
| `pkg/websocket` | `websocket_test.go` | WebSocket Hub/实时推送 | ≥80% |
| `pkg/bigdata` | `bigdata_test.go` | 大数据服务/分析/摄入 | ≥75% |
| `pkg/feedback` | `feedback_test.go` | 反馈收集/验证 | ≥80% |
| `pkg/nacos` | `options_test.go` | 配置选项/健康检查 | ≥75% |
| `pkg/config` | `example_test.go` | 配置加载/监听 | ≥70% |

### 4.6 零覆盖率包补充计划

以下包当前测试覆盖率为 0%，需优先补充：

| 包路径 | 优先级 | 预估工作量 | 补充策略 |
|--------|--------|-----------|----------|
| `internal/api/dto` | P1 | 2d | 参数验证测试 + 序列化测试 |
| `internal/domain/cache` | P2 | 1d | 接口契约测试 |
| `internal/domain/logger` | P2 | 1d | 接口契约测试 |
| `internal/domain/repository` (未覆盖方法) | P2 | 2d | 通过实现层间接测试 |
| `pkg/ai/edge` (未覆盖方法) | P2 | 1d | 补充 model_server/sync_manager 测试 |
| `pkg/storage/lifecycle` (DryRun 路径) | P1 | 1d | 补充 DryRun 模式测试，修复无限循环 |
| `pkg/monitoring/error_monitor` | P3 | 0.5d | 错误监控逻辑测试 |
| `pkg/skills/bridge` | P3 | 0.5d | 桥接逻辑测试 |
| `cmd/*` (各入口 main) | P3 | 1d | 启动/关闭流程测试 |

---

## 5. 集成测试方案

### 5.1 数据库集成测试

#### 5.1.1 SQLite 内存数据库测试（快速反馈）

适用场景：日常开发、CI 快速测试

```go
func setupTestDB(t *testing.T) *Database {
    t.Helper()
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
    require.NoError(t, err)
    // ... 配置连接池
    t.Cleanup(func() { database.Close() })
    return database
}
```

#### 5.1.2 PostgreSQL 真实数据库测试（完整验证）

适用场景：合并请求、版本发布前

| 测试场景 | 测试内容 | 验证点 |
|----------|----------|--------|
| 迁移测试 | 执行全部迁移脚本 | 迁移幂等性、无数据丢失 |
| CRUD 全链路 | 实体创建到查询 | 数据完整性、关联正确 |
| 事务测试 | 并发事务操作 | 隔离级别、死锁处理 |
| 索引验证 | 大数据量查询 | 查询计划使用索引 |
| autoCreateTime 验证 | 实体创建/更新 | 时间字段自动填充、不被覆盖 |

**autoCreateTime 已知缺陷专项测试**：

```go
func TestQAEntity_AutoCreateTime_NotOverwritten(t *testing.T) {
    db := setupTestDB(t)
    session := &entity.QASession{
        ID:     uuid.New().String(),
        UserID: uuid.New().String(),
        Title:  "Test Session",
    }
    require.NoError(t, db.Create(session).Error)

    var found entity.QASession
    require.NoError(t, db.First(&found, "id = ?", session.ID).Error)
    assert.False(t, found.CreatedAt.IsZero(), "CreatedAt should be auto-filled")

    // 验证更新时 CreatedAt 不被覆盖
    originalCreatedAt := found.CreatedAt
    found.Title = "Updated Title"
    time.Sleep(10 * time.Millisecond)
    require.NoError(t, db.Save(&found).Error)

    var updated entity.QASession
    require.NoError(t, db.First(&updated, "id = ?", session.ID).Error)
    assert.Equal(t, originalCreatedAt, updated.CreatedAt, "CreatedAt should not be overwritten on update")
}
```

### 5.2 消息队列集成测试

| 测试场景 | 测试内容 | 验证点 |
|----------|----------|--------|
| 生产消息 | 告警事件发布 | 消息格式正确、分区分配 |
| 消费消息 | 告警事件消费 | 消费顺序、偏移量管理 |
| 重试机制 | 消费失败重试 | 重试次数、延迟、死信队列 |
| 顺序保证 | 同一设备告警 | 分区内顺序消费 |

### 5.3 缓存集成测试

| 测试场景 | 测试内容 | 验证点 |
|----------|----------|--------|
| 缓存命中 | 重复查询 | 响应时间缩短、DB 查询减少 |
| 缓存失效 | 数据更新后 | 缓存及时清除、下次查询刷新 |
| 缓存穿透 | 查询不存在的数据 | 空值缓存、布隆过滤器 |
| 缓存雪崩 | 大量缓存同时过期 | 随机 TTL、互斥锁 |
| 分布式锁 | Redis 锁实现 | 加锁/解锁/续期/竞争 |

### 5.4 API 集成测试

| 测试场景 | API | 验证点 |
|----------|-----|--------|
| 认证流程 | POST /api/v1/auth/login | Token 返回、过期处理 |
| 告警规则 CRUD | /api/v1/alarm-rules/* | 完整 CRUD、参数验证 |
| 电站管理 | /api/v1/stations/* | CRUD、类型筛选、统计 |
| 数据导出 | /api/v1/export/* | Excel/CSV 文件下载 |
| 权限控制 | 各 API | 未授权访问拒绝、越权拒绝 |

### 5.5 DryRun 无限循环缺陷修复验证

**缺陷描述**：`pkg/storage/lifecycle/cleanup.go` 中 `executeCleanup` 方法在 DryRun 模式下，当查询结果不为空且 `task.DryRun == true` 时，不执行删除操作，但循环继续查询相同的记录，导致无限循环。

**根因分析**：DryRun 模式下跳过删除（第 516 行 `if !task.DryRun`），但未在 DryRun 模式下模拟"已处理"效果，导致下一轮循环再次查到相同记录。

**修复验证测试**：

```go
func TestDataCleaner_DryRun_NoInfiniteLoop(t *testing.T) {
    db := setupTestDBWithMigrate(t)
    // 插入测试数据
    for i := 0; i < 10; i++ {
        db.Table("metrics").Create(map[string]interface{}{"id": i, "data": "test"})
    }

    cleaner := NewDataCleaner(db, DefaultCleanupConfig())
    policy := CleanupPolicy{
        ID:            "dry-run-test",
        DataType:      "metrics",
        RetentionDays: 0,
        DryRun:        true,
        BatchSize:     5,
    }

    done := make(chan struct{})
    go func() {
        cleaner.ExecuteCleanup(context.Background(), policy)
        close(done)
    }()

    select {
    case <-done:
        // 正常完成
    case <-time.After(10 * time.Second):
        t.Fatal("DryRun mode caused infinite loop")
    }
}
```

---

## 6. 性能测试方案

### 6.1 Go Benchmark 基准测试

#### 6.1.1 核心算法基准

| 模块 | 基准测试 | 目标 |
|------|----------|------|
| `pkg/compute/formula` | 公式解析与执行 | 解析 < 1μs/op，执行 < 100ns/op |
| `pkg/protocol/modbus` | Modbus 帧解析 | < 500ns/op |
| `pkg/protocol/iec104` | IEC104 ASDU 解析 | < 1μs/op |
| `pkg/alarm/detector` | 告警检测 | < 10μs/op |
| `pkg/storage/compression` | 数据压缩 | 压缩率 > 50%，速度 > 100MB/s |
| `pkg/processor` | 数据处理管道 | < 5μs/op |
| `pkg/ai/forecast` | 预测计算 | < 100ms/op |
| `pkg/export/csv` | CSV 导出 | 10 万行 < 2s |
| `pkg/export/excel` | Excel 导出 | 10 万行 < 5s |

#### 6.1.2 数据库操作基准

| 操作 | 基准测试 | 目标 |
|------|----------|------|
| 单条插入 | `BenchmarkDB_Insert` | < 1ms/op |
| 批量插入 (1000) | `BenchmarkDB_BatchInsert` | < 100ms/op |
| 单条查询 | `BenchmarkDB_Query` | < 500μs/op |
| 复杂关联查询 | `BenchmarkDB_JoinQuery` | < 5ms/op |
| 分页查询 | `BenchmarkDB_Pagination` | < 2ms/op |

### 6.2 API 压力测试

#### 6.2.1 测试场景

| 场景 | 并发数 | 持续时间 | 目标 |
|------|--------|----------|------|
| 常规负载 | 100 | 5min | P95 < 200ms，错误率 < 0.1% |
| 峰值负载 | 500 | 3min | P95 < 500ms，错误率 < 1% |
| 极限负载 | 1000 | 1min | 系统不崩溃，优雅降级 |
| 长时间运行 | 50 | 30min | 无内存泄漏，性能无衰减 |

#### 6.2.2 关键 API 性能指标

| API | 目标 P50 | 目标 P95 | 目标 P99 | 目标 QPS |
|-----|----------|----------|----------|----------|
| GET /api/v1/stations | < 50ms | < 100ms | < 200ms | ≥ 500 |
| GET /api/v1/alarms | < 50ms | < 100ms | < 200ms | ≥ 500 |
| POST /api/v1/auth/login | < 100ms | < 200ms | < 500ms | ≥ 200 |
| GET /api/v1/reports | < 200ms | < 500ms | < 1s | ≥ 100 |
| GET /api/v1/export/csv | < 500ms | < 2s | < 5s | ≥ 50 |
| GET /api/v1/export/excel | < 1s | < 5s | < 10s | ≥ 20 |
| WebSocket 连接 | < 100ms | - | - | ≥ 1000 连接 |

### 6.3 内存泄漏检测

| 测试 | 方法 | 验证点 |
|------|------|--------|
| 长时间运行 | pprof heap profile 30min | 堆内存无持续增长 |
| 连接池泄漏 | 监控数据库连接数 | 连接数稳定，无泄漏 |
| Goroutine 泄漏 | runtime.NumGoroutine() | Goroutine 数量稳定 |
| 缓存膨胀 | 监控 Redis 内存使用 | 缓存大小可控，有淘汰机制 |

### 6.4 性能测试执行脚本

项目已有性能测试脚本：

| 脚本 | 位置 | 用途 |
|------|------|------|
| `scripts/performance/benchmark.sh` | 基准测试执行 | Go Benchmark 批量运行 |
| `scripts/performance/load_test.js` | k6 压力测试 | API 负载测试 |
| `scripts/performance/run_perf_tests.sh` | 综合性能测试 | 一键执行全部性能测试 |
| `tests/performance/` | Go 性能测试代码 | Benchmark + 压力 + 内存泄漏 |

---

## 7. 安全测试方案

### 7.1 输入验证测试

| 测试场景 | 测试方法 | 验证点 |
|----------|----------|--------|
| 必填参数缺失 | 发送缺少必填字段的请求 | 返回 400 Bad Request |
| 参数类型错误 | 发送错误类型的参数 | 返回 400 Bad Request |
| 参数长度越界 | 发送超长字符串 | 返回 400 Bad Request |
| 非法字符注入 | 发送含特殊字符的参数 | 正确处理或拒绝 |
| 枚举值越界 | 发送不在枚举范围内的值 | 返回 400 Bad Request |
| 数值范围越界 | 发送超出范围的数值 | 返回 400 Bad Request |

### 7.2 SQL 注入测试

| 测试场景 | 注入向量 | 预期行为 |
|----------|----------|----------|
| 用户名注入 | `' OR '1'='1` | 参数化查询，无注入 |
| 搜索条件注入 | `"; DROP TABLE users;--` | 参数化查询，无注入 |
| 排序字段注入 | `name; DROP TABLE users` | 白名单校验，拒绝 |
| ID 参数注入 | `1 OR 1=1` | 参数化查询，无注入 |
| 模糊查询注入 | `%'; DELETE FROM alarms;--` | 参数化查询，无注入 |

**验证方法**：确认所有 Repository 层使用 GORM 参数化查询，无字符串拼接 SQL。

### 7.3 XSS 防护测试

| 测试场景 | 注入向量 | 预期行为 |
|----------|----------|----------|
| 告警描述 XSS | `<script>alert('xss')</script>` | HTML 转义或过滤 |
| 用户名 XSS | `<img src=x onerror=alert(1)>` | HTML 转义或过滤 |
| 电站名称 XSS | `"><script>document.cookie</script>` | HTML 转义或过滤 |
| URL 参数 XSS | `?name=<script>alert(1)</script>` | URL 编码处理 |

### 7.4 认证授权安全测试

| 测试场景 | 测试方法 | 预期行为 |
|----------|----------|----------|
| 无 Token 访问 | 不带 Authorization 头请求受保护 API | 返回 401 Unauthorized |
| 过期 Token | 使用过期 JWT 访问 | 返回 401 Unauthorized |
| 伪造 Token | 使用错误密钥签名的 JWT | 返回 401 Unauthorized |
| 越权访问 | 普通用户访问管理员 API | 返回 403 Forbidden |
| 水平越权 | 用户 A 访问用户 B 的数据 | 返回 403 Forbidden |
| 暴力破解 | 连续 5 次错误密码 | 账户锁定 30 分钟 |
| Token 刷新 | 使用过期 Access Token + 有效 Refresh Token | 成功刷新 |

### 7.5 CSRF 防护测试

| 测试场景 | 测试方法 | 预期行为 |
|----------|----------|----------|
| 跨域请求 | 从不同 Origin 发起请求 | CORS 策略拒绝 |
| 缺少 CSRF Token | POST 请求无 CSRF Token | 拒绝请求 |

### 7.6 敏感数据保护测试

| 测试场景 | 测试方法 | 预期行为 |
|----------|----------|----------|
| 密码存储 | 检查数据库密码字段 | bcrypt 哈希，非明文 |
| JWT 密钥 | 检查配置文件 | 环境变量注入，非硬编码 |
| 日志脱敏 | 检查日志输出 | 无密码/Token 等敏感信息 |
| 错误信息 | 触发数据库错误 | 不暴露 SQL 语句和堆栈 |

### 7.7 安全扫描工具

| 工具 | 扫描类型 | CI 集成 | 频率 |
|------|----------|---------|------|
| Trivy | 容器镜像漏洞扫描 | `.github/workflows/ci.yml` | 每次提交 |
| Gosec | Go 代码安全扫描 | 手动执行 | 每周 |
| golangci-lint | 代码质量检查 | `.github/workflows/ci.yml` | 每次提交 |
| `.gitleaks.toml` | 密钥泄漏检测 | Git pre-commit hook | 每次提交 |

---

## 8. 测试数据管理

### 8.1 测试数据工厂

项目已实现测试数据工厂方法，位于 `tests/helpers/test_utils.go`：

| 工厂方法 | 用途 |
|----------|------|
| `CreateTestUser(username, password)` | 创建测试用户实体 |
| `CreateTestRole(code, name)` | 创建测试角色实体 |
| `CreateTestPermission(code, name, resourceType, action)` | 创建测试权限实体 |
| `CreateTestRegion(code, name, parentID)` | 创建测试区域实体 |
| `CreateTestStation(code, name, stationType, subRegionID)` | 创建测试电站实体 |
| `CreateTestDevice(code, name, deviceType, stationID)` | 创建测试设备实体 |
| `CreateTestAlarm(pointID, deviceID, stationID, alarmType, level)` | 创建测试告警实体 |
| `CreateTestPoint(code, name, pointType, deviceID, stationID)` | 创建测试采集点实体 |

### 8.2 测试数据 Fixtures

#### 8.2.1 MockDB 种子数据

`tests/helpers/mock_db.go` 中 `SeedTestData()` 方法提供基础种子数据：

| 实体 | ID | 数据 |
|------|-----|------|
| 用户 | user-001 | testuser / hashedpassword123 |
| 角色 | role-001 | admin / 管理员 |
| 权限 | perm-001 | user:read / 查看用户 |
| 区域 | region-001 | EAST / 华东区域 |
| 电站 | station-001 | PV_001 / 测试光伏电站 |
| 设备 | device-001 | INV_001 / 1号逆变器 |
| 告警 | alarm-001 | 测试告警 |

#### 8.2.2 数据库种子数据

`tests/test_config.go` 中 `TestSuite` 提供数据库级别的种子数据方法：

| 方法 | 用途 |
|------|------|
| `CreateUser(username, password)` | 在数据库中创建用户 |
| `CreateRole(code, name)` | 在数据库中创建角色 |
| `CreatePermission(code, name, resourceType, action)` | 在数据库中创建权限 |

### 8.3 测试数据生命周期

```
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│  Setup   │ →  │  Execute │ →  │  Assert  │ →  │ Teardown │
│ 准备数据  │    │ 执行测试  │    │ 验证结果  │    │ 清理数据  │
└──────────┘    └──────────┘    └──────────┘    └──────────┘
```

| 阶段 | 操作 | 工具 |
|------|------|------|
| Setup | 创建测试数据 / 迁移数据库 | `setupTestDB()`, `SeedTestData()` |
| Execute | 执行被测方法 | `go test` |
| Assert | 验证结果 | `testify/assert`, `testify/require` |
| Teardown | 清理数据 / 关闭连接 | `t.Cleanup()`, `db.Cleanup()`, `mock.Clear()` |

### 8.4 测试数据隔离策略

| 策略 | 适用场景 | 实现方式 |
|------|----------|----------|
| 内存数据库 | 单元测试 | SQLite `:memory:`，每次测试新建 |
| 事务回滚 | 集成测试 | `db.Transaction()` + 回滚 |
| 数据库清理 | E2E 测试 | `TRUNCATE TABLE ... CASCADE` |
| 命名空间隔离 | 并行测试 | 每个测试使用唯一 ID 前缀 |

### 8.5 测试数据合规要求

- 测试数据不包含真实用户信息
- 密码使用固定测试密码（`test-secret-key-for-unit-testing`）
- JWT 密钥使用测试专用密钥
- 不在日志中输出敏感测试数据

---

## 9. 缺陷管理

### 9.1 缺陷严重程度定义

| 等级 | 名称 | 定义 | 响应时间 | 修复时限 |
|------|------|------|----------|----------|
| P0 | 阻塞 (Blocker) | 系统崩溃、数据丢失、安全漏洞 | 立即 | 4 小时内 |
| P1 | 严重 (Critical) | 核心功能不可用、性能严重下降 | 2 小时内 | 24 小时内 |
| P2 | 一般 (Major) | 非核心功能异常、体验受损 | 1 个工作日 | 3 个工作日内 |
| P3 | 轻微 (Minor) | UI 瑕疵、文案错误、体验优化 | 3 个工作日 | 下版本修复 |
| P4 | 建议 (Suggestion) | 优化建议、改进意见 | 版本规划时 | 评估后决定 |

### 9.2 已知缺陷清单

| 缺陷ID | 严重程度 | 模块 | 描述 | 状态 | 修复方案 |
|--------|----------|------|------|------|----------|
| BUG-001 | P1 | `pkg/storage/lifecycle` | DryRun 模式下 `executeCleanup` 无限循环 | 待修复 | DryRun 模式下模拟"已处理"效果，使用 offset 或标记已扫描记录 |
| BUG-002 | P2 | `internal/domain/entity` | GORM `autoCreateTime` 标签在手动设置 `CreatedAt` 时被覆盖 | 待修复 | 测试中验证行为一致性；生产代码改用 `<-:create` 标签 |
| BUG-003 | P2 | 多模块 | Flaky 测试：部分测试因时序依赖或资源竞争偶发失败 | 待修复 | 添加超时保护、使用确定性等待、隔离共享状态 |

### 9.3 缺陷工作流

```
┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐
│  新建   │ →  │  确认   │ →  │  修复中  │ →  │  验证   │ →  │  关闭   │
│  New    │    │ Confirm │    │ Fixing  │    │ Verify  │    │ Closed  │
└─────────┘    └─────────┘    └─────────┘    └─────────┘    └─────────┘
     │                                            │
     │         ┌─────────┐                        │
     └────────→│  退回   │←───────────────────────┘
               │ Reopen  │   验证失败时退回
               └─────────┘
```

| 状态 | 操作人 | 说明 |
|------|--------|------|
| 新建 (New) | 测试人员 | 发现缺陷并提交 |
| 确认 (Confirmed) | 开发负责人 | 确认缺陷有效，分配优先级 |
| 修复中 (Fixing) | 开发人员 | 正在修复 |
| 验证 (Verifying) | 测试人员 | 验证修复结果 |
| 关闭 (Closed) | 测试人员 | 验证通过，关闭缺陷 |
| 退回 (Reopened) | 测试人员 | 验证失败，退回修复 |

### 9.4 缺陷追踪

- **追踪工具**：GitHub Issues
- **标签体系**：`bug`, `P0`/`P1`/`P2`/`P3`, `module:*`, `regression`
- **关联要求**：每个缺陷关联到具体的测试用例和代码变更
- **回归验证**：修复后必须通过原有测试 + 新增回归测试

### 9.5 Flaky 测试治理

| 策略 | 具体措施 |
|------|----------|
| 识别 | CI 中标记偶发失败的测试，收集失败日志 |
| 隔离 | 使用 `t.Parallel()` 标记可并行测试，串行运行有状态测试 |
| 确定性 | 消除时间依赖（用固定时钟）、网络依赖（用 Mock）、随机数依赖（用固定种子） |
| 超时保护 | 所有测试设置 `go test -timeout 5m`，单个测试 `t.Timeout()` |
| 重试机制 | CI 中对标记为 flaky 的测试允许 1 次重试 |
| 根因修复 | 优先修复 flaky 测试而非跳过 |

---

## 10. 测试覆盖率目标

### 10.1 总体覆盖率目标

| 指标 | 当前值 | 目标值 | 达标标准 |
|------|--------|--------|----------|
| 总体行覆盖率 | ~70% | ≥80% | CI 门禁检查 |
| 包覆盖率达标率 | 64% (41/64) | ≥95% (61/64) | 0% 包清零 |
| P0/P1 模块覆盖率 | ~85% | ≥90% | 核心模块强制 |
| 新增代码覆盖率 | - | ≥85% | PR 合并条件 |

### 10.2 各包覆盖率目标

#### 10.2.1 领域层 `internal/domain`

| 包 | 当前覆盖率 | 目标覆盖率 | 优先级 |
|----|-----------|-----------|--------|
| `entity` | ~80% | ≥90% | P1 |
| `repository` | 0% | ≥80%（通过实现层） | P2 |
| `cache` | 0% | ≥75%（通过实现层） | P2 |
| `logger` | 0% | ≥75%（通过实现层） | P2 |

#### 10.2.2 应用层 `internal/application/service`

| 包 | 当前覆盖率 | 目标覆盖率 | 优先级 |
|----|-----------|-----------|--------|
| `service` (全部) | ~85% | ≥85% | P1 |

#### 10.2.3 基础设施层 `internal/infrastructure`

| 包 | 当前覆盖率 | 目标覆盖率 | 优先级 |
|----|-----------|-----------|--------|
| `persistence` | ~75% | ≥85% | P1 |
| `cache` | ~85% | ≥85% | P1 |
| `config` | ~85% | ≥85% | P1 |
| `mq` | ~80% | ≥80% | P2 |
| `logger` | ~80% | ≥80% | P2 |

#### 10.2.4 接口层 `internal/api`

| 包 | 当前覆盖率 | 目标覆盖率 | 优先级 |
|----|-----------|-----------|--------|
| `handler` | ~70% | ≥85% | P1 |
| `dto` | 0% | ≥80% | P1 |

#### 10.2.5 核心包 `pkg`

| 包 | 当前覆盖率 | 目标覆盖率 | 优先级 |
|----|-----------|-----------|--------|
| `alarm/*` | ~85% | ≥85% | P1 |
| `ai/*` | ~80% | ≥80% | P1 |
| `protocol/*` | ~85% | ≥85% | P1 |
| `compute/*` | ~80% | ≥80% | P1 |
| `storage/*` | ~80% | ≥80% | P1 |
| `monitoring/*` | ~80% | ≥80% | P2 |
| `harness` | ~85% | ≥85% | P1 |
| `auth` | ~90% | ≥90% | P1 |
| `export` | ~80% | ≥80% | P1 |
| `collector` | ~80% | ≥80% | P2 |
| `processor` | ~80% | ≥80% | P2 |
| `statistics/*` | ~80% | ≥80% | P2 |
| `cache` | ~80% | ≥80% | P2 |
| `errors` | ~85% | ≥85% | P2 |
| `websocket` | ~80% | ≥80% | P2 |
| `bigdata` | ~75% | ≥75% | P3 |
| `feedback` | ~80% | ≥80% | P3 |
| `nacos` | ~70% | ≥75% | P3 |
| `config` | ~60% | ≥70% | P3 |
| `skills` | 0% | ≥70% | P3 |

### 10.3 覆盖率检查机制

| 检查点 | 阈值 | 执行方式 |
|--------|------|----------|
| CI 流水线总覆盖率 | ≥70% | `.github/workflows/ci.yml` 自动检查 |
| Makefile 覆盖率阈值 | ≥80% | `make test-coverage-check` |
| PR 覆盖率检查 | 新增代码 ≥85% | Codecov PR Comment |
| 覆盖率趋势 | 不下降 | Codecov 趋势图 |

---

## 11. 测试自动化

### 11.1 CI/CD 集成

#### 11.1.1 GitHub Actions 工作流

| 工作流 | 触发条件 | 执行内容 | 文件 |
|--------|----------|----------|------|
| CI | push/PR to main/develop | 后端测试 + 前端测试 + E2E + 代码质量 | `.github/workflows/ci.yml` |
| Test Coverage | push/PR to main/develop | 覆盖率报告 + Codecov 上传 | `.github/workflows/test-coverage.yml` |
| Security | 定期/schedule | 安全扫描 | `.github/workflows/security.yml` |
| CD | tag push | 构建 + 部署 | `.github/workflows/cd.yml` |
| Release | tag push | 发布 | `.github/workflows/release.yml` |

#### 11.1.2 CI 流水线阶段

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  Backend     │    │  Frontend    │    │  Code        │    │  E2E         │
│  Tests       │    │  Tests       │    │  Quality     │    │  Tests       │
│              │    │              │    │              │    │              │
│  go test     │    │  npm ci      │    │  golangci-   │    │  Playwright  │
│  -race       │    │  npm test    │    │  lint        │    │  Chromium    │
│  -cover      │    │  npm build   │    │  Trivy       │    │              │
└──────┬───────┘    └──────┬───────┘    └──────┬───────┘    └──────┬───────┘
       │                   │                   │                   │
       └───────────────────┴───────────────────┘                   │
                           │                                       │
                    ┌──────▼───────┐                               │
                    │  Coverage    │                               │
                    │  Check       │                               │
                    │  ≥70%        │                               │
                    └──────┬───────┘                               │
                           │                                       │
                    ┌──────▼───────────────────────────────────────▼───┐
                    │              E2E Tests (needs: backend+frontend) │
                    └──────────────────────────────────────────────────┘
```

#### 11.1.3 质量门禁

| 门禁 | 条件 | 阻断级别 |
|------|------|----------|
| 单元测试通过 | 0 失败 | 阻断合并 |
| 覆盖率达标 | 总覆盖率 ≥70%，新增代码 ≥85% | 阻断合并 |
| Lint 通过 | 0 error | 阻断合并 |
| 安全扫描 | 无 CRITICAL/HIGH 漏洞 | 阻断合并 |
| E2E 通过 | 核心流程 0 失败 | 阻断合并 |

### 11.2 自动化测试报告

| 报告类型 | 生成方式 | 存储位置 | 保留时间 |
|----------|----------|----------|----------|
| Go 覆盖率报告 | `go tool cover -html` | GitHub Artifacts | 30 天 |
| Codecov 趋势 | Codecov 自动 | Codecov 平台 | 永久 |
| Playwright 报告 | `npx playwright show-report` | GitHub Artifacts | 30 天 |
| Lint 报告 | golangci-lint | GitHub Actions Log | 90 天 |
| 安全扫描报告 | Trivy | GitHub Actions Log | 90 天 |

### 11.3 Git Hooks 自动化

项目已配置 Git Hooks（`scripts/git-hooks/`）：

| Hook | 功能 |
|------|------|
| `pre-commit` | 提交前运行 `go fmt`、`go vet`、lint |
| `pre-push` | 推送前运行单元测试 |
| `commit-msg` | 提交消息格式检查（Conventional Commits） |

### 11.4 定时任务

| 任务 | 频率 | 内容 |
|------|------|------|
| 每日全量测试 | 每日凌晨 | 运行全部测试套件 + 覆盖率报告 |
| 每周安全扫描 | 每周一 | Trivy + Gosec 全量扫描 |
| 每周性能基准 | 每周五 | 运行 Benchmark 并对比历史数据 |

---

## 12. 测试里程碑

### 12.1 里程碑计划

| 阶段 | 时间 | 目标 | 交付物 |
|------|------|------|--------|
| M1: 基础设施准备 | 第 1 周 | 测试环境搭建、Mock 完善、CI 配置 | 测试环境就绪、CI 流水线可用 |
| M2: 已知缺陷修复 | 第 2 周 | 修复 DryRun 无限循环、autoCreateTime 覆盖、Flaky 测试 | 3 个已知缺陷关闭 |
| M3: 零覆盖率包补充 | 第 3 周 | 补充 0% 覆盖率包的测试 | 0% 包数量从 23 降至 ≤5 |
| M4: 核心模块深化 | 第 4 周 | 告警/认证/电站等核心模块覆盖率提升至 ≥85% | 核心模块覆盖率达标 |
| M5: 集成测试完善 | 第 5 周 | 数据库/缓存/MQ/API 集成测试 | 集成测试套件完整 |
| M6: 性能/安全测试 | 第 6 周 | Benchmark + 压力测试 + 安全扫描 | 性能基线建立、安全报告 |
| M7: 覆盖率达标 | 第 7 周 | 总覆盖率 ≥80%，所有包 ≥70% | 覆盖率达标报告 |
| M8: 验收与发布 | 第 8 周 | 全量回归测试 + 验收 | 测试总结报告、发布就绪 |

### 12.2 里程碑验收标准

| 里程碑 | 验收标准 |
|--------|----------|
| M1 | CI 流水线绿色通过，测试环境可正常执行测试 |
| M2 | BUG-001/002/003 状态为 Closed，回归测试通过 |
| M3 | 0% 覆盖率包数量 ≤5，新增测试用例 ≥50 |
| M4 | 核心模块（alarm/auth/station/device）覆盖率 ≥85% |
| M5 | 集成测试覆盖全部核心 API，数据库集成测试通过 |
| M6 | 性能基线建立，P95 响应时间达标，无 HIGH 以上安全漏洞 |
| M7 | 总覆盖率 ≥80%，CI 覆盖率门禁通过 |
| M8 | 全量测试通过，测试报告输出，发布评审通过 |

### 12.3 风险与应对

| 风险 | 影响 | 概率 | 应对措施 |
|------|------|------|----------|
| PostgreSQL 与 SQLite 行为差异 | 集成测试不可靠 | 中 | 关键路径增加 PostgreSQL 真实测试 |
| Flaky 测试影响 CI 稳定性 | CI 频繁红灯 | 中 | 优先修复 Flaky 测试，添加重试机制 |
| 大数据量测试耗时过长 | CI 执行时间超标 | 低 | 大数据量测试放入定时任务，PR 只跑小数据量 |
| 第三方服务不可用 | 集成测试失败 | 低 | 使用 Mock 隔离，关键路径添加降级测试 |
| 覆盖率目标无法达成 | 发布延期 | 低 | 优先补充核心模块，非核心模块适当降低目标 |

---

## 附录 A：测试用例矩阵

### A.1 告警规则管理测试矩阵

| 用例ID | 场景 | 前置条件 | 操作 | 预期结果 | 优先级 |
|--------|------|----------|------|----------|--------|
| AR-001 | 创建告警规则-正常 | 用户已登录且有权限 | 提交有效规则数据 | 规则创建成功，返回 201 | P1 |
| AR-002 | 创建告警规则-名称重复 | 已存在同名规则 | 提交重复名称 | 返回 409 Conflict | P1 |
| AR-003 | 创建告警规则-参数缺失 | 用户已登录 | 缺少必填字段 | 返回 400 Bad Request | P1 |
| AR-004 | 查询告警规则-列表 | 存在多条规则 | 请求规则列表 | 返回分页数据 | P1 |
| AR-005 | 查询告警规则-按类型筛选 | 存在不同类型规则 | 按类型筛选 | 返回匹配规则 | P2 |
| AR-006 | 更新告警规则-正常 | 规则已存在 | 更新阈值 | 更新成功 | P1 |
| AR-007 | 更新告警规则-不存在 | 规则 ID 不存在 | 更新不存在的规则 | 返回 404 Not Found | P2 |
| AR-008 | 删除告警规则-正常 | 规则已存在 | 删除规则 | 删除成功 | P1 |
| AR-009 | 启用/禁用规则 | 规则已存在 | 切换状态 | 状态变更成功 | P1 |
| AR-010 | 规则触发告警 | 规则已启用，数据超阈值 | 采集数据超阈值 | 生成告警记录 | P1 |

### A.2 数据导出测试矩阵

| 用例ID | 场景 | 前置条件 | 操作 | 预期结果 | 优先级 |
|--------|------|----------|------|----------|--------|
| EX-001 | CSV 导出-正常 | 存在统计数据 | 选择 CSV 格式导出 | 下载 CSV 文件，数据正确 | P1 |
| EX-002 | Excel 导出-正常 | 存在统计数据 | 选择 Excel 格式导出 | 下载 Excel 文件，数据正确 | P1 |
| EX-003 | 导出-空数据 | 无匹配数据 | 导出空数据集 | 返回空文件或提示 | P2 |
| EX-004 | 导出-大数据量 | 10 万条数据 | 导出全量数据 | 10s 内完成，数据完整 | P1 |
| EX-005 | 导出-无权限 | 用户无导出权限 | 请求导出 | 返回 403 Forbidden | P1 |
| EX-006 | CSV 导出-特殊字符 | 数据含逗号/引号 | 导出含特殊字符数据 | CSV 格式正确，无注入 | P2 |
| EX-007 | Excel 导出-多 Sheet | 多维度数据 | 导出多维度报表 | 多 Sheet 正确生成 | P2 |

### A.3 认证授权测试矩阵

| 用例ID | 场景 | 前置条件 | 操作 | 预期结果 | 优先级 |
|--------|------|----------|------|----------|--------|
| AU-001 | 登录-正常 | 用户已注册 | 正确用户名密码 | 返回 Token 对 | P0 |
| AU-002 | 登录-密码错误 | 用户已注册 | 错误密码 | 返回 401 | P0 |
| AU-003 | 登录-账户锁定 | 连续 5 次错误 | 第 6 次尝试 | 账户锁定 30 分钟 | P0 |
| AU-004 | Token 刷新 | Access Token 过期 | 使用 Refresh Token | 返回新 Token 对 | P0 |
| AU-005 | 越权访问 | 普通用户 | 访问管理员 API | 返回 403 | P1 |
| AU-006 | Token 伪造 | 无 | 使用伪造 Token | 返回 401 | P0 |

---

## 附录 B：测试命令速查

| 命令 | 用途 |
|------|------|
| `make test` | 运行全部测试（含覆盖率） |
| `make test-unit` | 运行单元测试 |
| `make test-integration` | 运行集成测试 |
| `make test-coverage` | 生成覆盖率报告 |
| `make test-coverage-check` | 检查覆盖率是否达标（≥80%） |
| `make lint` | 运行代码检查 |
| `make vet` | 运行 go vet |
| `go test -v -race -run TestXxx ./pkg/...` | 运行指定测试 |
| `go test -bench=BenchmarkXxx -benchmem ./...` | 运行基准测试 |
| `bash tests/coverage.sh` | 运行覆盖率脚本 |
| `bash tests/run_tests.sh` | 运行单元测试脚本 |
| `bash tests/integration_test.sh` | 运行集成测试脚本 |
| `bash scripts/performance/run_perf_tests.sh` | 运行性能测试 |

---

**文档维护**：本测试方案应根据项目进展持续更新，建议每个里程碑结束后进行一次审查和修订。
