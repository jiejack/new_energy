# NEM 核心功能代码审查报告

**项目名称**: NEM (New Energy Monitoring) 新能源监控系统  
**审查版本**: v1.0 (主干分支)  
**审查日期**: 2026-05-29  
**审查人员**: 代码审查团队  
**审查范围**: 后端核心功能模块（Go 代码）  

---

## 1. 审查概述

### 1.1 审查目标

本次代码审查旨在对 NEM 新能源监控系统的核心功能代码进行全面评估，识别潜在的缺陷、安全风险、性能瓶颈和可维护性问题，为后续迭代优化提供依据。

### 1.2 审查范围

| 模块 | 路径 | 说明 |
|------|------|------|
| API 处理层 | `internal/api/` | HTTP Handler、DTO 定义 |
| 应用服务层 | `internal/application/service/` | 业务逻辑服务 |
| 领域层 | `internal/domain/` | 实体、仓储接口、缓存 |
| 基础设施层 | `internal/infrastructure/` | 数据库、配置、消息队列 |
| 存储生命周期 | `pkg/storage/lifecycle/` | 数据清理、归档、备份 |
| 计算引擎 | `pkg/compute/` | 公式计算、规则引擎 |
| 认证授权 | `pkg/auth/` | JWT、密码管理 |
| 协议适配 | `pkg/protocol/` | Modbus、IEC104、IEC61850 |
| AI 模块 | `pkg/ai/` | 预测、故障检测、QA |
| 告警系统 | `pkg/alarm/` | 告警检测、通知、状态机 |

### 1.3 审查方法

- **静态分析**: 逐文件代码审查，识别逻辑缺陷和代码异味
- **架构审查**: 检查分层架构合规性和依赖方向
- **安全审查**: 检查输入验证、SQL 注入、认证授权
- **性能审查**: 检查数据库查询、缓存策略、并发处理
- **测试审查**: 评估测试覆盖率、测试质量和测试基础设施

### 1.4 代码规模统计

| 指标 | 数值 |
|------|------|
| Go 源文件总数 | ~200+ |
| 测试文件总数 | 168 |
| 核心包数量 | ~30 |
| 代码行数（估算） | ~50,000+ |

---

## 2. 代码质量评估

### 2.1 总体评分

| 维度 | 评分 (1-10) | 等级 | 说明 |
|------|-------------|------|------|
| **功能完整性** | 7.0 | B | 核心功能基本完整，部分模块存在占位代码 |
| **代码规范** | 6.0 | C+ | 命名基本规范，但存在不一致和硬编码问题 |
| **错误处理** | 5.0 | C- | 关键路径错误处理不充分，缺少恢复机制 |
| **安全性** | 5.5 | C | 存在 SQL 注入风险和输入验证缺失 |
| **性能** | 5.5 | C | 存在 N+1 查询和频繁 DB 访问问题 |
| **可维护性** | 5.0 | C- | 长函数过多，缺少文档，代码重复 |
| **测试覆盖** | 3.0 | D | 大量包零覆盖率，测试质量参差不齐 |
| **综合评分** | **5.3** | **C** | 需要重点改进 |

### 2.2 各模块评分

| 模块 | 评分 | 关键问题 |
|------|------|----------|
| `pkg/storage/lifecycle/` | 4.0 | DryRun 无限循环、无事务、SQL 注入 |
| `pkg/compute/formula/` | 6.5 | 缓存淘汰策略粗糙，并发安全可改进 |
| `pkg/compute/rule/` | 6.0 | Redis KEYS 命令性能问题，缓存一致性 |
| `pkg/auth/` | 7.0 | 密码策略较弱，JWT 实现基本合理 |
| `internal/api/handler/` | 5.5 | 输入验证不足，错误处理不一致 |
| `internal/application/service/` | 6.0 | 服务层较薄，部分逻辑应下沉 |
| `internal/domain/entity/` | 7.5 | 领域模型设计合理，业务方法封装良好 |
| `internal/infrastructure/persistence/` | 6.0 | 缺少事务管理，查询未优化 |
| `pkg/protocol/` | 6.5 | 协议实现较完整，错误处理需加强 |
| `pkg/alarm/` | 6.5 | 状态机设计合理，通知渠道待完善 |

---

## 3. 架构审查

### 3.1 分层架构合规性

项目采用了 DDD（领域驱动设计）分层架构，整体结构如下：

```
cmd/                    → 应用入口
internal/api/           → 接口层（Handler + DTO）
internal/application/   → 应用层（Service）
internal/domain/        → 领域层（Entity + Repository Interface）
internal/infrastructure/→ 基础设施层（Persistence + Config + MQ）
pkg/                    → 公共包（可被外部引用）
```

**合规性评估**:

| 检查项 | 结果 | 说明 |
|--------|------|------|
| 领域层不依赖基础设施层 | ⚠️ 部分违反 | `pkg/storage/lifecycle/` 直接使用 `*gorm.DB`，绕过仓储接口 |
| 应用层通过接口解耦 | ✅ 合规 | Service 层通过 Repository 接口访问数据 |
| Handler 层仅做参数转换 | ⚠️ 部分违反 | 部分 Handler 直接返回内部错误信息 |
| 依赖注入使用 Wire | ✅ 合规 | 使用 Google Wire 进行依赖注入 |
| 公共包独立可复用 | ⚠️ 部分违反 | `pkg/storage/lifecycle/` 直接依赖 `gorm.DB`，耦合度高 |

### 3.2 依赖方向问题

**问题 1**: `pkg/storage/lifecycle/cleanup.go` 直接依赖 `*gorm.DB`

```go
type DataCleaner struct {
    db *gorm.DB  // 应通过仓储接口抽象
}
```

**问题 2**: `pkg/storage/lifecycle/archive.go` 同样直接依赖 `*gorm.DB`，且与 `cleanup.go` 存在大量代码重复（策略管理、任务管理、Worker 模式等）。

**问题 3**: `pkg/compute/rule/cache.go` 直接依赖 `github.com/go-redis/redis/v8`，而非通过缓存接口抽象，导致与特定 Redis 客户端版本绑定。

### 3.3 架构建议

1. **引入仓储接口**: `pkg/storage/lifecycle/` 应定义仓储接口，由基础设施层实现，降低耦合
2. **消除代码重复**: `cleanup.go` 和 `archive.go` 共享大量相似逻辑，应提取公共基类或组合模式
3. **缓存接口抽象**: `pkg/compute/rule/cache.go` 应定义通用缓存接口，支持不同缓存后端

---

## 4. 安全审查

### 4.1 SQL 注入风险

| 严重级别 | 位置 | 描述 |
|----------|------|------|
| **🔴 严重** | `cleanup.go:711-712` | `CleanupByQuery` 方法直接接收 `whereClause` 参数拼接到查询中 |
| **🔴 严重** | `cleanup.go:726-727` | `CleanupByDate` 使用 `fmt.Sprintf` 拼接字段名 |
| **🔴 严重** | `cleanup.go:826-829` | `CleanupOrphanedRecords` 使用 `fmt.Sprintf` 拼接表名和字段名 |
| **🔴 严重** | `cleanup.go:865-874` | `CleanupDuplicates` 使用 `fmt.Sprintf` 拼接 SQL |
| **🔴 严重** | `cleanup.go:894` | `VacuumTable` 使用 `fmt.Sprintf` 拼接表名 |
| **🔴 严重** | `cleanup.go:908` | `ReindexTable` 使用 `fmt.Sprintf` 拼接表名 |
| **🔴 严重** | `cleanup.go:948` | `AnalyzeTable` 使用 `fmt.Sprintf` 拼接表名 |

**示例 - `CleanupByQuery` 方法**:

```go
func (dc *DataCleaner) CleanupByQuery(ctx context.Context, tableName string,
    whereClause string, args ...interface{}) (int64, error) {
    result := dc.db.Table(tableName).Where(whereClause, args...).Delete(nil)
    // whereClause 完全由调用方控制，存在 SQL 注入风险
}
```

**修复建议**: 
- 对所有动态表名和字段名进行白名单校验
- 禁止直接传入原始 SQL 子句，改用结构化查询构建器
- 添加表名/字段名的正则校验（仅允许字母、数字、下划线）

### 4.2 输入验证缺失

| 严重级别 | 位置 | 描述 |
|----------|------|------|
| **🟠 高** | `cleanup.go:176-202` | `CreatePolicy` 未验证 `DataType`（表名）是否合法 |
| **🟠 高** | `cleanup.go:251-285` | `TriggerCleanup` 未验证 `policyID` 格式 |
| **🟠 高** | `dto/request.go` | 多个请求 DTO 缺少字段长度限制和格式校验 |
| **🟡 中** | `auth_handler.go:74-76` | `Logout` 中 `userID` 为空字符串（TODO 未完成） |
| **🟡 中** | `alarm_handler.go:96-97` | `AcknowledgeAlarm` 中 `by` 参数为空字符串（TODO 未完成） |

**示例 - `CreatePolicy` 缺少验证**:

```go
func (dc *DataCleaner) CreatePolicy(ctx context.Context, policy *CleanupPolicy) error {
    // 缺少对 policy.DataType 的验证，可能导致访问任意数据库表
    // 缺少对 policy.Name 长度的验证
    // 缺少对 policy.RetentionDays 合理范围的验证
}
```

### 4.3 认证授权问题

| 严重级别 | 位置 | 描述 |
|----------|------|------|
| **🟠 高** | `auth_handler.go:74-76` | `Logout` 未从上下文获取用户 ID，空字符串可被利用 |
| **🟠 高** | `alarm_handler.go:96-97` | `AcknowledgeAlarm` 未获取操作人信息 |
| **🟡 中** | `jwt.go` | JWT 使用对称加密（HS256），建议支持非对称加密（RS256） |
| **🟡 中** | `jwt.go:115-122` | `RefreshAccessToken` 未检查 RefreshToken 是否已被撤销 |
| **🟡 中** | `password.go` | 密码策略未要求特殊字符，强度较弱 |

### 4.4 敏感信息泄露

| 严重级别 | 位置 | 描述 |
|----------|------|------|
| **🟡 中** | `auth_handler.go:50-53` | 登录失败时直接返回 `err.Error()`，可能泄露内部信息 |
| **🟡 中** | `alarm_handler.go:67-69` | 内部错误信息直接返回给客户端 |
| **🟢 低** | `dto/response.go` | `ErrorResponse.Timestamp` 始终为 0，未正确设置 |

---

## 5. 性能审查

### 5.1 数据库查询问题

| 严重级别 | 位置 | 描述 |
|----------|------|------|
| **🔴 严重** | `cleanup.go:566-572` | `checkTaskCancelled` 在循环中每次迭代都查询数据库 |
| **🟠 高** | `cleanup.go:697-707` | `updateMetrics` 执行 3 次独立 COUNT 查询，应合并为 1 次 |
| **🟠 高** | `cleanup.go:778-793` | `GetStorageStats` 执行 7 次独立查询（1 COUNT + 2 MIN/MAX + 5 时间段），应合并 |
| **🟠 高** | `archive.go:744-755` | `updateMetrics` 同样执行 3 次独立 COUNT 查询 |
| **🟡 中** | `repository.go:63-71` | `GetTree` 使用 Preload 嵌套加载，可能产生 N+1 查询 |
| **🟡 中** | `repository.go:127-137` | `GetWithDevices` 嵌套 Preload，数据量大时性能差 |

**示例 - `checkTaskCancelled` 循环内查询**:

```go
func (dc *DataCleaner) doCleanup(ctx context.Context, task *CleanupTask, policy *CleanupPolicy) error {
    for {
        if dc.checkTaskCancelled(task.ID) {  // 每次循环都查询 DB
            return fmt.Errorf("task cancelled")
        }
        // ... 处理逻辑
    }
}

func (dc *DataCleaner) checkTaskCancelled(taskID string) bool {
    var task CleanupTask
    if err := dc.db.Select("status").First(&task, "id = ?", taskID).Error; err != nil {
        return false  // 查询失败默认不取消，可能导致无法停止的任务
    }
    return task.Status == CleanupStatusCancelled
}
```

**修复建议**: 
- 使用内存缓存 + 定期刷新替代每次查询
- 引入取消信号通道（channel），避免轮询数据库
- 批次间隔中添加适当休眠，减少数据库压力

### 5.2 缓存策略问题

| 严重级别 | 位置 | 描述 |
|----------|------|------|
| **🟠 高** | `formula/executor.go:274-283` | `buildCacheKey` 使用 `fmt.Sprintf` 拼接变量值，map 遍历顺序不确定导致缓存键不稳定 |
| **🟠 高** | `rule/cache.go:542-554` | `RedisCache.Clear` 使用 `KEYS` 命令，在大数据量下会阻塞 Redis |
| **🟠 高** | `rule/cache.go:557-569` | `RedisCache.Invalidate` 同样使用 `KEYS` 命令 |
| **🟡 中** | `formula/executor.go:383-405` | `evict` 淘汰策略简单粗暴（随机删除一半），命中率低 |
| **🟡 中** | `rule/cache.go:471-487` | `LocalCache.cleanup` 协程无退出机制，goroutine 泄漏 |

**示例 - 缓存键不稳定**:

```go
func (e *Executor) buildCacheKey(formula string, variables map[string]interface{}) string {
    key := formula
    if variables != nil {
        for k, v := range variables {  // map 遍历顺序不确定
            key += fmt.Sprintf("|%s=%v", k, v)
        }
    }
    return key
}
```

**修复建议**: 对 map key 排序后再拼接，或使用确定性序列化（如 JSON + sort keys）。

### 5.3 并发问题

| 严重级别 | 位置 | 描述 |
|----------|------|------|
| **🟠 高** | `cleanup.go:454-465` | `executeCleanupTask` 中 `metrics` 更新存在竞态：先读 `CompletedTasks` 再写，与 `updateMetrics` 并发冲突 |
| **🟡 中** | `formula/executor.go:319-337` | `ResultCache.Get` 中 `stats.Misses` 在 RLock 下修改，存在数据竞争 |
| **🟡 中** | `rule/cache.go:363-382` | `LocalCache.Get` 在 RLock 下修改 `accessTime` 和 `accessCount`，存在数据竞争 |
| **🟡 中** | `rule/cache.go:471-487` | `LocalCache.cleanup` 启动 goroutine 但无停止机制 |

### 5.4 资源管理

| 严重级别 | 位置 | 描述 |
|----------|------|------|
| **🔴 严重** | `cleanup.go:516-517` | 删除操作未使用事务，查询和删除之间存在数据不一致窗口 |
| **🔴 严重** | `archive.go:583-593` | 归档后删除数据未使用事务，归档成功但删除失败会导致数据不一致 |
| **🟠 高** | `archive.go:596-607` | 逐条创建 `ArchiveRecord`，应使用批量插入 |
| **🟡 中** | `cleanup.go:585` | `logTask` 写入日志失败被静默忽略，可能丢失关键操作记录 |

---

## 6. 可维护性审查

### 6.1 代码复杂度

**长函数统计**:

| 文件 | 函数名 | 行数 | 建议上限 |
|------|--------|------|----------|
| `cleanup.go` | `doCleanup` | 72 | 50 |
| `cleanup.go` | `executeCleanupTask` | 70 | 50 |
| `archive.go` | `doArchive` | 100 | 50 |
| `archive.go` | `executeArchiveTask` | 80 | 50 |
| `archive.go` | `CompactArchives` | 130 | 50 |
| `archive.go` | `RestoreArchive` | 73 | 50 |
| `rule/cache.go` | 整个文件 | 676 | 应拆分 |
| `formula/executor.go` | 整个文件 | 767 | 应拆分 |

### 6.2 代码重复

| 重复模式 | 位置 | 说明 |
|----------|------|------|
| 策略 CRUD | `cleanup.go` / `archive.go` | CreatePolicy、UpdatePolicy、DeletePolicy、GetPolicy、ListPolicies 几乎完全相同 |
| 任务管理 | `cleanup.go` / `archive.go` | GetTask、ListTasks、TriggerCleanup/TriggerArchive 模式相同 |
| Worker 模式 | `cleanup.go` / `archive.go` | cleanupWorker / archiveWorker、schedulerWorker、metricsWorker 结构相同 |
| 指标更新 | `cleanup.go` / `archive.go` | updateMetrics 逻辑完全相同 |
| 失败处理 | `cleanup.go` / `archive.go` | failTask 逻辑相同 |

**重复代码量估算**: `cleanup.go` 和 `archive.go` 之间约 60% 的代码为重复逻辑。

### 6.3 硬编码值

| 位置 | 硬编码值 | 说明 |
|------|----------|------|
| `cleanup.go:526` | `deleted * 1024` | 假设每条记录 1KB，应可配置 |
| `cleanup.go:811` | `count * 1024` | 同上 |
| `cleanup.go:647` | `24 * time.Hour` | 日志清理间隔硬编码 |
| `cleanup.go:681` | `30 * time.Second` | 指标更新间隔硬编码 |
| `archive.go:381` | `1000` | 批量插入大小硬编码 |
| `rule/cache.go:472` | `1 * time.Minute` | 缓存清理间隔硬编码 |
| `database.go:35` | `logger.Info` | 生产环境日志级别应为 Warn |

### 6.4 命名不一致

| 问题 | 位置 | 说明 |
|------|------|------|
| 混合语言注释 | 全局 | 部分注释中文，部分英文 |
| DTO 命名不一致 | `dto/request.go` | `LoginRequest` 在 `dto` 和 `service` 中重复定义 |
| 错误码不一致 | `handler/` | 部分使用 HTTP 状态码，部分使用业务码 |
| 时间戳格式 | `dto/response.go` | `Timestamp int64` 但始终为 0，未统一使用 |

### 6.5 文档缺失

| 缺失项 | 说明 |
|--------|------|
| 包级文档 | 大部分包缺少 `doc.go` 或包注释 |
| 函数文档 | `pkg/storage/lifecycle/` 公开方法缺少详细文档 |
| 架构文档 | 缺少模块间交互的时序图和状态图 |
| API 文档 | Swagger 注解存在但部分不完整 |

---

## 7. 测试审查

### 7.1 测试覆盖率

| 模块 | 测试文件数 | 覆盖评估 | 说明 |
|------|-----------|----------|------|
| `internal/domain/entity/` | 14 | ✅ 较好 | 实体测试覆盖了核心业务逻辑 |
| `internal/application/service/` | 30+ | ⚠️ 一般 | 测试存在但部分为结构验证测试 |
| `internal/api/handler/` | 15+ | ⚠️ 一般 | 测试存在但缺少边界场景 |
| `internal/infrastructure/persistence/` | 8 | ⚠️ 一般 | 部分仓储有测试 |
| `pkg/storage/lifecycle/` | 4 | ❌ 差 | 仅测试结构体字段，无功能测试 |
| `pkg/compute/formula/` | 1 | ❌ 差 | 仅基础测试 |
| `pkg/compute/rule/` | 1 | ❌ 差 | 仅基础测试 |
| `pkg/auth/` | 1 | ❌ 差 | 仅基础测试 |
| `pkg/protocol/modbus/` | 2 | ⚠️ 一般 | 有基础和综合测试 |
| `pkg/ai/` | 10+ | ⚠️ 一般 | 各子模块有测试但覆盖不全 |

**零覆盖率包（关键模块）**:

以下核心包的测试文件仅包含结构体字段验证测试，实际业务逻辑覆盖率为 0%：

- `pkg/storage/lifecycle/cleanup.go` — 967 行代码，测试仅验证结构体字段
- `pkg/storage/lifecycle/archive.go` — 1137 行代码，测试仅验证结构体字段
- `pkg/storage/lifecycle/backup.go`
- `pkg/storage/lifecycle/tiered.go`
- `pkg/storage/query/executor.go`
- `pkg/storage/timeseries/` 多个文件
- `pkg/collector/` 核心采集逻辑
- `pkg/bigdata/` 大数据处理

### 7.2 测试质量评估

**问题 1: 测试仅验证结构体字段**

```go
// cleanup_test.go — 当前测试
func TestCleanupPolicy_Struct(t *testing.T) {
    policy := CleanupPolicy{
        ID: "cp1",
        Name: "Test Cleanup",
        // ...
    }
    assert.Equal(t, "cp1", policy.ID)  // 仅验证赋值
}
```

这种测试不验证任何业务逻辑，对代码质量的保障价值极低。

**问题 2: Mock 实现不完整**

`tests/helpers/mock_db.go` 中的 `MockDB` 实现了手动内存数据库，但：
- 未实现 `repository.UserRepository` 接口
- 未实现 `repository.AlarmRepository` 接口
- 缺少对错误场景的模拟
- 缺少并发安全测试

**问题 3: 缺少集成测试和端到端测试**

- 无数据库集成测试（使用真实或测试数据库）
- 无 API 端到端测试
- 无并发安全测试
- 无性能基准测试（`pkg/` 下大部分模块）

### 7.3 测试基础设施问题

| 问题 | 说明 |
|------|------|
| 无测试数据库 | 缺少 SQLite 内存数据库或 Testcontainers 集成 |
| Mock 不符合接口 | `MockDB` 未实现仓储接口，无法用于 Service 层测试 |
| 无测试夹具 | 缺少统一的测试数据工厂 |
| 无覆盖率门槛 | CI 中未设置覆盖率最低要求 |
| 不稳定测试 | `formula/cache` 测试可能因并发问题偶发失败 |

---

## 8. 缺陷清单

### 8.1 严重缺陷 (Critical)

| 编号 | 位置 | 描述 | 修复建议 |
|------|------|------|----------|
| C-01 | `cleanup.go:484-540` | **DryRun 模式无限循环**: 当 `DryRun=true` 时，`doCleanup` 的 for 循环中只累加 `totalDeleted` 但不实际删除记录，下一轮查询会返回相同的记录，导致无限循环 | DryRun 模式下应在查询后直接 break，仅统计不循环 |
| C-02 | `cleanup.go:38,54,65` | **GORM autoCreateTime 覆盖手动时间戳**: `CleanupPolicy.CreatedAt`、`CleanupTask.CreatedAt`、`CleanupLog.CreatedAt` 标记了 `gorm:"autoCreateTime"`，但代码中手动设置 `CreatedAt = time.Now()`，GORM 会忽略手动值 | 移除 `gorm:"autoCreateTime"` 标签或移除手动赋值，二选一 |
| C-03 | `cleanup.go:711-712` | **SQL 注入**: `CleanupByQuery` 接受原始 `whereClause` 参数 | 改用结构化查询构建器，禁止传入原始 SQL |
| C-04 | `cleanup.go:826-829` | **SQL 注入**: `CleanupOrphanedRecords` 使用 `fmt.Sprintf` 拼接表名 | 添加表名白名单校验 |
| C-05 | `cleanup.go:865-874` | **SQL 注入**: `CleanupDuplicates` 使用 `fmt.Sprintf` 拼接 SQL | 使用参数化查询或 ORM 方法 |
| C-06 | `cleanup.go:516-517` | **删除操作无事务**: 查询记录后删除，中间可能被其他事务修改 | 使用数据库事务包裹查询和删除操作 |
| C-07 | `archive.go:583-593` | **归档删除无事务**: 归档成功后删除源数据，若删除失败则数据不一致 | 使用事务确保归档和删除的原子性 |

### 8.2 高危缺陷 (High)

| 编号 | 位置 | 描述 | 修复建议 |
|------|------|------|----------|
| H-01 | `cleanup.go:566-572` | `checkTaskCancelled` 循环内频繁 DB 查询 | 使用内存状态 + channel 通知替代 DB 轮询 |
| H-02 | `cleanup.go:414` | `dc.db.Save(task)` 错误被忽略 | 检查并处理 Save 返回的错误 |
| H-03 | `cleanup.go:451` | `dc.db.Save(task)` 完成时错误被忽略 | 检查并处理 Save 返回的错误 |
| H-04 | `archive.go:440` | `da.db.Save(task)` 错误被忽略 | 检查并处理 Save 返回的错误 |
| H-05 | `archive.go:488` | `da.db.Save(task)` 完成时错误被忽略 | 检查并处理 Save 返回的错误 |
| H-06 | `auth_handler.go:74-76` | `Logout` 未获取用户 ID，空字符串可被利用 | 从 JWT 中间件上下文获取用户 ID |
| H-07 | `alarm_handler.go:96-97` | `AcknowledgeAlarm` 未获取操作人 | 从 JWT 中间件上下文获取操作人 |
| H-08 | `rule/cache.go:542-554` | `RedisCache.Clear` 使用 `KEYS` 命令 | 使用 `SCAN` 命令替代 `KEYS` |
| H-09 | `cleanup.go:176-202` | `CreatePolicy` 未验证 `DataType` 表名合法性 | 添加正则校验，仅允许字母数字下划线 |
| H-10 | `formula/executor.go:274-283` | `buildCacheKey` map 遍历顺序不确定 | 对 key 排序后再拼接 |
| H-11 | `cleanup.go:454-465` | `metrics` 并发更新竞态条件 | 使用 atomic 操作或统一由 metricsWorker 更新 |

### 8.3 中危缺陷 (Medium)

| 编号 | 位置 | 描述 | 修复建议 |
|------|------|------|----------|
| M-01 | `cleanup.go:585` | `logTask` 写入失败被静默忽略 | 至少记录日志 |
| M-02 | `archive.go:606` | 逐条创建 `ArchiveRecord` | 使用批量插入 |
| M-03 | `rule/cache.go:363-382` | `LocalCache.Get` 在 RLock 下修改字段 | 改用 atomic 或升级为写锁 |
| M-04 | `rule/cache.go:471-487` | `LocalCache.cleanup` goroutine 泄漏 | 添加 context 或 stop channel |
| M-05 | `formula/executor.go:319-337` | `ResultCache.Get` 在 RLock 下修改 stats | 使用 atomic 计数器 |
| M-06 | `dto/response.go:10` | `Response.Timestamp` 始终为 0 | 正确设置时间戳 |
| M-07 | `database.go:35` | 生产环境 GORM 日志级别为 Info | 应根据环境配置日志级别 |
| M-08 | `jwt.go:115-122` | RefreshToken 未检查是否被撤销 | 实现 Token 黑名单机制 |
| M-09 | `password.go` | 密码策略未要求特殊字符 | 增加特殊字符要求 |
| M-10 | `cleanup.go:697-707` | `updateMetrics` 3 次独立 COUNT 查询 | 合并为单次查询 |
| M-11 | `archive.go:744-755` | `updateMetrics` 3 次独立 COUNT 查询 | 合并为单次查询 |
| M-12 | `cleanup.go:778-793` | `GetStorageStats` 7 次独立查询 | 合并为少量查询 |

### 8.4 低危缺陷 (Low)

| 编号 | 位置 | 描述 | 修复建议 |
|------|------|------|----------|
| L-01 | `cleanup.go:962-967` | 自定义 `min` 函数，Go 1.21+ 已内置 | 使用内置 `min` 函数 |
| L-02 | `dto/request.go` | 多处缺少字段长度限制 | 添加 `max` 和 `min` 验证标签 |
| L-03 | `handler/` 多处 | 错误响应中 `Timestamp` 始终为 0 | 使用 `time.Now().UnixMilli()` |
| L-04 | `archive.go:896-897` | `defer gzReader.Close()` 在循环内 | 应在循环体内显式关闭 |
| L-05 | `cleanup.go` / `archive.go` | 中英文注释混用 | 统一注释语言 |

---

## 9. 重构建议

### 9.1 高优先级（P0 - 必须修复）

| 编号 | 建议 | 影响范围 | 预估工作量 |
|------|------|----------|-----------|
| R-01 | **修复 DryRun 无限循环** | `cleanup.go` | 0.5 天 |
| R-02 | **修复 GORM autoCreateTime 冲突** | `cleanup.go`, `archive.go` | 0.5 天 |
| R-03 | **修复 SQL 注入漏洞** | `cleanup.go` 6 处 | 2 天 |
| R-04 | **添加事务支持** | `cleanup.go`, `archive.go` | 1 天 |
| R-05 | **修复认证 TODO 项** | `auth_handler.go`, `alarm_handler.go` | 1 天 |

### 9.2 中优先级（P1 - 应当修复）

| 编号 | 建议 | 影响范围 | 预估工作量 |
|------|------|----------|-----------|
| R-06 | **提取 lifecycle 公共基类** | `cleanup.go`, `archive.go`, `backup.go` | 3 天 |
| R-07 | **优化 checkTaskCancelled** | `cleanup.go` | 1 天 |
| R-08 | **修复 DB Save 错误忽略** | `cleanup.go`, `archive.go` | 0.5 天 |
| R-09 | **替换 Redis KEYS 为 SCAN** | `rule/cache.go` | 1 天 |
| R-10 | **修复并发安全问题** | `rule/cache.go`, `formula/executor.go` | 2 天 |
| R-11 | **添加关键模块单元测试** | `pkg/storage/lifecycle/` | 5 天 |
| R-12 | **引入仓储接口抽象** | `pkg/storage/lifecycle/` | 2 天 |

### 9.3 低优先级（P2 - 建议改进）

| 编号 | 建议 | 影响范围 | 预估工作量 |
|------|------|----------|-----------|
| R-13 | **消除硬编码值** | 全局 | 2 天 |
| R-14 | **统一错误处理模式** | `handler/` | 2 天 |
| R-15 | **完善 Mock 和测试基础设施** | `tests/` | 3 天 |
| R-16 | **合并重复 COUNT 查询** | `cleanup.go`, `archive.go` | 1 天 |
| R-17 | **修复缓存键不稳定问题** | `formula/executor.go` | 0.5 天 |
| R-18 | **统一注释语言** | 全局 | 1 天 |
| R-19 | **添加包级文档** | 全局 | 2 天 |
| R-20 | **完善 DTO 验证规则** | `dto/request.go` | 1 天 |

---

## 10. 审查结论

### 10.1 总体评价

NEM 新能源监控系统的核心功能在架构设计上采用了合理的 DDD 分层模式，领域模型设计清晰，实体业务方法封装良好。但在代码实现层面存在以下突出问题：

1. **安全风险严重**: `pkg/storage/lifecycle/cleanup.go` 中存在 6 处 SQL 注入漏洞，且输入验证严重不足，在生产环境中可能被利用造成数据泄露或破坏。

2. **关键逻辑缺陷**: DryRun 模式下的无限循环 bug 会导致系统资源耗尽；GORM `autoCreateTime` 标签与手动赋值冲突导致时间戳可能不正确。

3. **数据一致性风险**: 数据清理和归档操作未使用数据库事务，在异常情况下可能导致数据不一致或数据丢失。

4. **性能隐患**: 循环内数据库查询、Redis `KEYS` 命令、缓存键不稳定等问题在大数据量场景下将严重影响系统性能。

5. **测试覆盖不足**: 核心模块（数据生命周期管理、计算引擎）的测试仅停留在结构体验证层面，实际业务逻辑覆盖率为 0%，无法有效保障代码质量。

### 10.2 行动项

| 优先级 | 行动项 | 负责人 | 截止日期 |
|--------|--------|--------|----------|
| P0 | 修复所有 SQL 注入漏洞 | 后端团队 | 立即 |
| P0 | 修复 DryRun 无限循环 | 后端团队 | 立即 |
| P0 | 修复 GORM autoCreateTime 冲突 | 后端团队 | 1 周内 |
| P0 | 为清理/归档操作添加事务 | 后端团队 | 1 周内 |
| P0 | 修复认证 TODO 项（空 userID/by） | 后端团队 | 1 周内 |
| P1 | 重构 lifecycle 模块，消除代码重复 | 后端团队 | 2 周内 |
| P1 | 优化 checkTaskCancelled 性能 | 后端团队 | 2 周内 |
| P1 | 修复并发安全问题 | 后端团队 | 2 周内 |
| P1 | 为核心模块补充单元测试 | 后端团队 | 4 周内 |
| P2 | 统一错误处理和输入验证 | 后端团队 | 6 周内 |
| P2 | 完善测试基础设施 | 后端团队 | 6 周内 |
| P2 | 消除硬编码和代码异味 | 后端团队 | 8 周内 |

### 10.3 风险评估

| 风险 | 可能性 | 影响 | 风险等级 |
|------|--------|------|----------|
| SQL 注入被利用 | 高 | 严重 | 🔴 极高 |
| DryRun 导致服务不可用 | 中 | 严重 | 🔴 极高 |
| 数据不一致/丢失 | 中 | 严重 | 🟠 高 |
| 性能瓶颈影响用户体验 | 高 | 中等 | 🟠 高 |
| 并发 bug 导致数据错误 | 中 | 中等 | 🟡 中 |
| 测试不足导致回归 | 高 | 中等 | 🟡 中 |

---

**审查完成日期**: 2026-05-29  
**下次审查建议**: P0 缺陷修复完成后进行复审
