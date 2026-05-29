# NEM 核心功能迭代任务列表

**项目名称**: NEM (New Energy Monitoring) 新能源监控系统  
**文档版本**: v1.0  
**编制日期**: 2026-05-29  
**规划周期**: 24 周 (6 个月)  
**基准文档**: [05_code_review.md](05_code_review.md)、[04_test_plan.md](04_test_plan.md)

---

## 1. 任务总览

### 1.1 任务统计

| 统计维度 | 数量 |
|----------|------|
| **任务总数** | 55 |
| **P0 (紧急)** | 6 |
| **P1 (高)** | 14 |
| **P2 (中)** | 22 |
| **P3 (低)** | 13 |
| **总预估工时** | ~1,420 小时 |

### 1.2 阶段分布

| 阶段 | 名称 | 周期 | 任务数 | 预估工时 | 优先级范围 |
|------|------|------|--------|----------|-----------|
| Phase 1 | 紧急修复与基础加固 | Week 1-4 | 6 | 180h | P0 |
| Phase 2 | 测试覆盖与质量提升 | Week 5-8 | 25 | 420h | P0-P1 |
| Phase 3 | 功能完善与架构优化 | Week 9-14 | 17 | 440h | P1-P2 |
| Phase 4 | 性能优化与安全加固 | Week 15-20 | 4 | 200h | P1-P2 |
| Phase 5 | 高级特性与生产就绪 | Week 21-24 | 3 | 180h | P2-P3 |

### 1.3 类型分布

| 类型 | 数量 | 说明 |
|------|------|------|
| BUG | 5 | 缺陷修复 |
| SEC | 3 | 安全加固 |
| TEST | 25 | 测试补充 |
| FEAT | 14 | 功能实现 |
| REFACTOR | 5 | 代码重构 |
| PERF | 3 | 性能优化 |
| DOC | 2 | 文档编写 |

### 1.4 依赖关系概览

```
Phase 1 (BUG/SEC) ──→ Phase 2 (TEST) ──→ Phase 3 (FEAT/REFACTOR) ──→ Phase 4 (PERF/SEC) ──→ Phase 5 (FEAT/DOC)
     │                    │                      │                          │                       │
     └─ BUG-001~005      └─ TEST-001~025        └─ FEAT-001~012           └─ PERF-001~003        └─ FEAT-013~014
     └─ SEC-001~002                             └─ REFACTOR-001~005       └─ SEC-003              └─ DOC-001~002
```

---

## 2. Phase 1：紧急修复与基础加固 (Week 1-4)

> **目标**: 修复所有 P0 级别的关键缺陷和安全漏洞，消除系统最严重的风险点。  
> **准入条件**: 无  
> **准出条件**: 所有 P0 任务验收通过，CI 流水线绿灯

### 2.1 任务清单

| ID | 标题 | 描述 | 优先级 | 预估工时 | 依赖 | 验收标准 | 状态 |
|----|------|------|--------|----------|------|----------|------|
| BUG-001 | 修复 DryRun 无限循环 | `cleanup.go` 中 `doCleanup` 方法在 DryRun 模式下，`totalDeleted` 仅做累加但不实际删除记录，导致下一轮循环查询到相同记录，形成无限循环。需在 DryRun 模式下使用 `break` 跳出循环或记录已扫描 ID 避免重复查询 | P0 | 8h | 无 | 1) DryRun 模式下循环正常终止；2) 单元测试覆盖 DryRun 场景；3) 集成测试验证大批量数据下无死循环 | 🔴 待开始 |
| BUG-002 | 修复 GORM autoCreateTime 覆盖 | `CleanupPolicy`、`CleanupTask`、`CleanupLog` 结构体中 `CreatedAt` 字段同时标注 `gorm:"autoCreateTime"` 并在代码中手动赋值 `time.Now()`，导致 GORM 自动管理与手动赋值冲突。需移除手动赋值，统一由 GORM 管理 | P0 | 4h | 无 | 1) 所有 `autoCreateTime` 字段移除手动赋值；2) 数据库写入时间与实际时间一致；3) 单元测试验证时间戳正确性 | 🔴 待开始 |
| SEC-001 | 修复 SQL 注入漏洞 (6处) | 以下 6 个方法存在 SQL 注入风险：1) `CleanupByQuery` - `whereClause` 直接拼接；2) `CleanupByDate` - `dateField` 通过 `fmt.Sprintf` 拼接；3) `CleanupOrphanedRecords` - `childTable/parentTable/foreignKey` 直接拼接；4) `CleanupDuplicates` - `tableName/fieldList` 直接拼接；5) `VacuumTable` - `tableName` 直接拼接；6) `ReindexTable`/`AnalyzeTable` - `tableName` 直接拼接。需对表名/字段名使用白名单校验，对查询条件使用参数化查询 | P0 | 24h | 无 | 1) 所有动态 SQL 使用参数化查询或白名单校验；2) 表名/字段名仅允许字母数字下划线；3) 安全扫描工具无 SQL 注入告警；4) 添加注入攻击防御单元测试 | 🔴 待开始 |
| SEC-002 | 添加 API Handler 输入验证 | 所有 API Handler 缺少统一的输入验证中间件。需：1) 添加 `binding` 标签验证；2) 实现 `SanitizeMiddleware` 防止 XSS；3) 添加请求体大小限制；4) 对路径参数进行格式校验（UUID 格式等）；5) 统一错误响应格式 | P0 | 40h | 无 | 1) 所有 Handler 使用 `binding` 标签；2) 中间件拦截非法输入返回 400；3) 路径参数 UUID 格式校验；4) 请求体大小限制 1MB；5) 验证测试覆盖所有 Handler | 🔴 待开始 |
| BUG-003 | 添加事务支持 | `doCleanup` 中的删除操作、`CleanupOrphanedRecords`、`CleanupDuplicates` 等方法缺少数据库事务。删除/归档操作若中途失败会导致数据不一致。需使用 GORM 事务包裹关键操作 | P0 | 16h | BUG-001 | 1) 删除操作在事务内执行；2) 失败时自动回滚；3) 集成测试验证事务回滚场景；4) 长事务添加超时控制 | 🔴 待开始 |
| BUG-004 | 修复 checkTaskCancelled 频繁 DB 查询 | `checkTaskCancelled` 在 `doCleanup` 的每次循环迭代中都查询数据库检查任务状态，当批次量大时产生大量无效查询。需引入内存状态缓存或使用 channel 通知机制替代轮询 | P0 | 12h | 无 | 1) 取消检查不再每次迭代查询 DB；2) 使用 channel 或内存标记实现取消通知；3) 压测验证 DB 查询次数显著下降；4) 取消响应延迟 < 1s | 🔴 待开始 |
| BUG-005 | 修复 WebSocket 并发问题 | `Hub.Run()` 中 `broadcast` 分支在 `RLock` 下修改 map（`delete(h.clients, client)`），违反 Go map 并发安全规则；`sendHeartbeat` 同样存在此问题。`BroadcastToStation`/`BroadcastToUser` 直接操作 map 绕过 channel，与 `Run()` 存在数据竞争。需统一通过 channel 操作或使用 `Lock` 替代 `RLock` | P0 | 16h | 无 | 1) 所有 map 修改操作使用写锁；2) 广播操作统一通过 channel；3) `-race` 检测无数据竞争；4) 并发压测 1000+ 连接无 panic | 🔴 待开始 |

### 2.2 Phase 1 风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| BUG-001 修复引入新逻辑缺陷 | 高 | 中 | 编写充分的 DryRun 场景单元测试和集成测试 |
| SEC-001 白名单校验影响现有 API | 中 | 低 | 逐步上线，先在测试环境验证所有 API 调用 |
| BUG-005 WebSocket 重构影响在线连接 | 高 | 中 | 实现优雅迁移，新旧逻辑兼容过渡 |

---

## 3. Phase 2：测试覆盖与质量提升 (Week 5-8)

> **目标**: 将零覆盖率包的测试覆盖率提升至 60% 以上，修复不稳定测试，建立测试基础设施。  
> **准入条件**: Phase 1 全部完成  
> **准出条件**: 整体测试覆盖率 ≥ 50%，零覆盖率包清零，CI 流水线稳定

### 3.1 零覆盖包测试任务

| ID | 包路径 | 当前覆盖率 | 目标覆盖率 | 优先级 | 预估工时 | 依赖 | 验收标准 | 状态 |
|----|--------|-----------|-----------|--------|----------|------|----------|------|
| TEST-001 | `internal/api/handler/asset_handler.go` | 0% | ≥60% | P1 | 12h | SEC-002 | CRUD + 折旧计算测试通过 | 🔴 待开始 |
| TEST-002 | `internal/api/handler/asset_depreciation_handler.go` | 0% | ≥60% | P1 | 8h | SEC-002 | CRUD + 汇总查询测试通过 | 🔴 待开始 |
| TEST-003 | `internal/api/handler/asset_document_handler.go` | 0% | ≥60% | P1 | 8h | SEC-002 | CRUD + 文件关联测试通过 | 🔴 待开始 |
| TEST-004 | `internal/api/handler/asset_maintenance_handler.go` | 0% | ≥60% | P1 | 8h | SEC-002 | CRUD + 维护成本统计测试通过 | 🔴 待开始 |
| TEST-005 | `internal/api/handler/cost_entry_handler.go` | 0% | ≥60% | P1 | 10h | SEC-002 | CRUD + 审批流程 + 统计测试通过 | 🔴 待开始 |
| TEST-006 | `internal/api/handler/cost_category_handler.go` | 0% | ≥60% | P1 | 6h | SEC-002 | CRUD + 树形结构测试通过 | 🔴 待开始 |
| TEST-007 | `internal/api/handler/cost_allocation_handler.go` | 0% | ≥60% | P1 | 8h | SEC-002 | CRUD + 分配统计测试通过 | 🔴 待开始 |
| TEST-008 | `internal/api/handler/cost_report_handler.go` | 0% | ≥60% | P1 | 10h | SEC-002 | CRUD + 生成/审批/拒绝流程测试通过 | 🔴 待开始 |
| TEST-009 | `internal/api/handler/inventory_handler.go` | 0% | ≥60% | P1 | 10h | SEC-002 | CRUD + 低库存 + 交易测试通过 | 🔴 待开始 |
| TEST-010 | `internal/api/handler/purchase_order_handler.go` | 0% | ≥60% | P1 | 10h | SEC-002 | CRUD + 状态流转测试通过 | 🔴 待开始 |
| TEST-011 | `internal/api/handler/receipt_handler.go` | 0% | ≥60% | P1 | 8h | SEC-002 | CRUD + 状态流转测试通过 | 🔴 待开始 |
| TEST-012 | `internal/api/handler/work_order_handler.go` | 0% | ≥60% | P1 | 10h | SEC-002 | CRUD + 统计测试通过 | 🔴 待开始 |
| TEST-013 | `internal/api/handler/energy_efficiency_handler.go` | 0% | ≥60% | P1 | 12h | SEC-002 | CRUD + 趋势/统计/对比/分析测试通过 | 🔴 待开始 |
| TEST-014 | `internal/application/service/asset_service.go` | 0% | ≥60% | P1 | 12h | BUG-003 | CRUD + 折旧计算业务逻辑测试通过 | 🔴 待开始 |
| TEST-015 | `internal/application/service/cost_entry_service.go` | 0% | ≥60% | P1 | 10h | BUG-003 | 成本条目全流程测试通过 | 🔴 待开始 |
| TEST-016 | `internal/application/service/cost_category_service.go` | 0% | ≥60% | P1 | 6h | 无 | 类别管理 + 树形构建测试通过 | 🔴 待开始 |
| TEST-017 | `internal/application/service/cost_allocation_service.go` | 0% | ≥60% | P1 | 8h | 无 | 分配逻辑 + 统计测试通过 | 🔴 待开始 |
| TEST-018 | `internal/application/service/cost_report_service.go` | 0% | ≥60% | P1 | 10h | 无 | 报表生成 + 审批测试通过 | 🔴 待开始 |
| TEST-019 | `internal/application/service/inventory_service.go` | 0% | ≥60% | P1 | 10h | BUG-003 | 库存管理 + 交易事务测试通过 | 🔴 待开始 |
| TEST-020 | `internal/application/service/purchase_order_service.go` | 0% | ≥60% | P1 | 10h | BUG-003 | 采购订单全流程测试通过 | 🔴 待开始 |
| TEST-021 | `internal/application/service/work_order_service.go` | 0% | ≥60% | P1 | 10h | BUG-003 | 工单全生命周期测试通过 | 🔴 待开始 |
| TEST-022 | `internal/application/service/carbon_emission_service.go` | 0% | ≥60% | P1 | 12h | 无 | 碳排放记录 + 分析 + 统计测试通过 | 🔴 待开始 |
| TEST-023 | `internal/application/service/energy_efficiency_service.go` | 0% | ≥60% | P1 | 12h | 无 | 能效记录 + 分析 + 趋势测试通过 | 🔴 待开始 |

### 3.2 测试质量与基础设施

| ID | 标题 | 描述 | 优先级 | 预估工时 | 依赖 | 验收标准 | 状态 |
|----|------|------|--------|----------|------|----------|------|
| TEST-024 | 修复不稳定测试 | 1) `formula/manager.go` 缓存测试因时间依赖导致偶发失败；2) `rule/scheduler.go` 触发器测试因 goroutine 调度不确定性导致 flaky。需引入确定性时间源和同步机制 | P0 | 16h | 无 | 1) CI 连续 20 次运行无 flaky 失败；2) 消除 `time.Sleep` 硬编码等待；3) 使用 `testify/assert` 的 Eventually 机制 | 🔴 待开始 |
| TEST-025 | 完善测试基础设施 | 1) 创建共享测试 fixture（常用实体工厂方法）；2) 封装通用 test helper（DB 初始化、Mock 构造、HTTP 请求构造）；3) 添加 `golden file` 测试模式用于复杂响应验证；4) 统一 mock 生成规范 | P1 | 24h | 无 | 1) `tests/helpers/` 包含通用 helper；2) 各 handler 测试使用统一 fixture；3) 文档化 mock 使用规范 | 🔴 待开始 |

### 3.3 Phase 2 里程碑

| 里程碑 | 时间节点 | 交付物 |
|--------|----------|--------|
| M2.1 | Week 5 | TEST-024、TEST-025 完成，测试基础设施就绪 |
| M2.2 | Week 6 | TEST-001~TEST-013 (Handler 层) 全部完成 |
| M2.3 | Week 7 | TEST-014~TEST-023 (Service 层) 全部完成 |
| M2.4 | Week 8 | 整体覆盖率验证 ≥ 50%，CI 稳定运行 |

---

## 4. Phase 3：功能完善与架构优化 (Week 9-14)

> **目标**: 完成空 Handler 实现，实现认证授权体系，重构长函数提升可维护性。  
> **准入条件**: Phase 2 全部完成  
> **准出条件**: 所有 Handler 功能完整，JWT 认证 + RBAC 授权上线，长函数重构完成

### 4.1 功能实现任务

| ID | 标题 | 描述 | 优先级 | 预估工时 | 依赖 | 验收标准 | 状态 |
|----|------|------|--------|----------|------|----------|------|
| FEAT-001 | 完善 CarbonEmission Handler | `carbon_emission_handler.go.tmp` 为临时文件，需重命名为 `.go` 并补充缺失的碳排放因子管理接口（CRUD + 版本管理 + 生效/失效） | P1 | 16h | SEC-002, TEST-022 | 1) 文件重命名并编译通过；2) 因子管理 CRUD 接口可用；3) 接口测试覆盖 ≥ 80% | 🔴 待开始 |
| FEAT-002 | 完善 AuthHandler Logout | `auth_handler.go` 中 `Logout` 方法的 `userID` 硬编码为空字符串，需从 JWT 中间件上下文中获取实际用户 ID | P1 | 4h | FEAT-011 | 1) Logout 正确获取 userID；2) Token 黑名单机制生效；3) 集成测试验证登出后 Token 失效 | 🔴 待开始 |
| FEAT-003 | 统一 Handler 错误响应格式 | 当前 Handler 存在三种错误响应格式：`dto.ErrorResponse`、`gin.H{"error": ...}`、`gin.H{"message": ...}`，需统一为 `dto.ErrorResponse` | P1 | 16h | SEC-002 | 1) 所有 Handler 使用 `dto.ErrorResponse`；2) 错误码体系一致；3) API 文档更新 | 🔴 待开始 |
| FEAT-004 | 完善 Asset 模块缺失接口 | 补充资产模块缺失的业务接口：1) 资产转移接口；2) 资产盘点接口；3) 资产报废流程 | P2 | 20h | TEST-001 | 1) 新接口 Handler + Service + Repository 实现；2) 单元测试覆盖 ≥ 70%；3) Swagger 文档更新 | 🔴 待开始 |
| FEAT-005 | 完善 Cost 模块缺失接口 | 补充成本模块缺失接口：1) 成本预算管理；2) 成本预警规则；3) 成本趋势分析 | P2 | 20h | TEST-005~008 | 1) 新接口完整实现；2) 测试覆盖 ≥ 70%；3) Swagger 文档更新 | 🔴 待开始 |
| FEAT-006 | 完善 Inventory 模块缺失接口 | 补充库存模块缺失接口：1) 库存预警规则配置；2) 批量导入/导出；3) 库存盘点 | P2 | 16h | TEST-009 | 1) 新接口完整实现；2) 测试覆盖 ≥ 70%；3) Swagger 文档更新 | 🔴 待开始 |
| FEAT-007 | 完善 WorkOrder 模块缺失接口 | 补充工单模块缺失接口：1) 工单分配；2) 工单转派；3) 工单评价；4) SLA 管理 | P2 | 16h | TEST-012 | 1) 新接口完整实现；2) 工单状态流转完整；3) 测试覆盖 ≥ 70% | 🔴 待开始 |
| FEAT-008 | 完善 PurchaseOrder 模块缺失接口 | 补充采购模块缺失接口：1) 采购审批流程；2) 采购退货；3) 供应商评价 | P2 | 12h | TEST-010 | 1) 新接口完整实现；2) 审批流程测试通过；3) Swagger 文档更新 | 🔴 待开始 |
| FEAT-009 | 完善 Receipt 模块缺失接口 | 补充收货模块缺失接口：1) 质检流程；2) 部分收货；3) 收货异常处理 | P2 | 12h | TEST-011 | 1) 新接口完整实现；2) 质检流程测试通过；3) Swagger 文档更新 | 🔴 待开始 |
| FEAT-010 | 完善 EnergyEfficiency 模块缺失接口 | 补充能效模块缺失接口：1) 能效基准线管理；2) 能效预警规则；3) 能效报告自动生成 | P2 | 16h | TEST-013 | 1) 新接口完整实现；2) 基准线计算测试通过；3) Swagger 文档更新 | 🔴 待开始 |

### 4.2 认证授权实现

| ID | 标题 | 描述 | 优先级 | 预估工时 | 依赖 | 验收标准 | 状态 |
|----|------|------|--------|----------|------|----------|------|
| FEAT-011 | 实现 JWT 认证中间件 | 当前 `pkg/auth/jwt.go` 已实现 JWT 生成/解析，但缺少 Gin 中间件集成。需：1) 实现 `AuthMiddleware` 从 Header 提取 Token 并验证；2) 将用户信息注入 Gin Context；3) 实现 Token 黑名单（Redis）；4) 添加登录频率限制 | P1 | 24h | BUG-005 | 1) 所有受保护 API 需携带有效 Token；2) 无效/过期 Token 返回 401；3) Token 黑名单即时生效；4) 登录失败 5 次锁定 15 分钟 | 🔴 待开始 |
| FEAT-012 | 实现 RBAC 授权 | 当前 `permission_service.go` 已有权限模型，但缺少授权中间件。需：1) 实现 `RequirePermission` 中间件；2) 实现 `RequireRole` 中间件；3) 角色权限缓存；4) 权限变更实时生效 | P1 | 24h | FEAT-011 | 1) 无权限访问返回 403；2) 角色权限缓存命中率 ≥ 90%；3) 权限变更 5s 内生效；4) 授权中间件测试覆盖 ≥ 80% | 🔴 待开始 |

### 4.3 代码重构任务

| ID | 标题 | 描述 | 优先级 | 预估工时 | 依赖 | 验收标准 | 状态 |
|----|------|------|--------|----------|------|----------|------|
| REFACTOR-001 | 重构 doCleanup 长函数 | `cleanup.go:doCleanup` 函数超过 70 行，职责过多（查询、删除、日志、进度更新）。需拆分为：1) `queryBatch` - 查询批次；2) `deleteBatch` - 删除批次；3) `updateProgress` - 更新进度；4) `handleDryRun` - DryRun 逻辑 | P1 | 8h | BUG-001, BUG-003 | 1) 每个函数 ≤ 30 行；2) 单一职责；3) 原有测试全部通过；4) 新增拆分后函数的单元测试 | 🔴 待开始 |
| REFACTOR-002 | 重构 executeCleanupTask | `cleanup.go:executeCleanupTask` 包含状态更新、重试逻辑、指标计算，超过 60 行。需拆分为：1) `markTaskRunning` - 状态更新；2) `retryCleanup` - 重试逻辑；3) `completeTask` - 完成处理；4) `updateMetrics` - 指标更新 | P1 | 8h | BUG-001 | 1) 每个函数 ≤ 30 行；2) 重试逻辑可配置；3) 原有测试全部通过 | 🔴 待开始 |
| REFACTOR-003 | 重构 Hub.Run 事件循环 | `websocket/hub.go:Run` 方法混合处理注册/注销/广播/心跳，需拆分为独立 handler 方法，并统一通过 channel 操作 | P1 | 12h | BUG-005 | 1) 事件处理逻辑拆分到独立方法；2) 无 map 并发修改；3) `-race` 检测通过；4) WebSocket 集成测试通过 | 🔴 待开始 |
| REFACTOR-004 | 提取 Handler 通用逻辑 | 所有 Handler 存在重复的分页解析、ID 提取、错误响应代码。需提取：1) `ParsePagination` - 分页参数解析；2) `GetIDParam` - 路径 ID 提取；3) `RespondError`/`RespondSuccess` - 统一响应 | P2 | 12h | FEAT-003 | 1) 通用函数提取到 `handler/common.go`；2) 所有 Handler 使用通用函数；3) 代码行数减少 ≥ 15%；4) 测试全部通过 | 🔴 待开始 |
| REFACTOR-005 | 重构 CostEntryHandler 长方法 | `cost_entry_handler.go:ListCostEntries` 超过 50 行，参数解析逻辑复杂。需提取参数解析为独立方法，使用结构体绑定替代手动解析 | P2 | 6h | FEAT-003 | 1) 方法行数 ≤ 30 行；2) 使用 `ShouldBindQuery` 替代手动解析；3) 测试全部通过 | 🔴 待开始 |

### 4.4 Phase 3 里程碑

| 里程碑 | 时间节点 | 交付物 |
|--------|----------|--------|
| M3.1 | Week 9-10 | FEAT-011、FEAT-012 认证授权体系上线 |
| M3.2 | Week 11 | FEAT-001~003 Handler 补全与统一 |
| M3.3 | Week 12-13 | FEAT-004~010 各模块缺失接口实现 |
| M3.4 | Week 14 | REFACTOR-001~005 重构完成，代码质量验证 |

---

## 5. Phase 4：性能优化与安全加固 (Week 15-20)

> **目标**: 解决性能瓶颈，实现缓存策略，完成安全审计与加固。  
> **准入条件**: Phase 3 全部完成  
> **准出条件**: P95 延迟降低 50%，安全审计无高危漏洞

### 5.1 性能优化任务

| ID | 标题 | 描述 | 优先级 | 预估工时 | 依赖 | 验收标准 | 状态 |
|----|------|------|--------|----------|------|----------|------|
| PERF-001 | 实现高频查询数据缓存 | 以下场景需引入缓存：1) 站点信息查询（`station_service`）- 热点数据；2) 设备信息查询（`device_service`）- 高频读取；3) 告警规则查询（`alarm_rule_service`）- 规则引擎频繁读取；4) 系统配置查询（`config_service`）- 全局配置；5) 权限信息查询（`permission_service`）- 每次请求校验。使用 Redis 缓存 + 主动失效策略 | P1 | 60h | FEAT-012 | 1) 缓存命中率 ≥ 85%；2) P95 查询延迟降低 ≥ 50%；3) 缓存失效延迟 < 5s；4) 缓存穿透/雪崩防护；5) 缓存监控指标接入 Prometheus | 🔴 待开始 |
| PERF-002 | 优化数据库查询 | 1) `GetStorageStats` 执行 7 次独立 COUNT 查询，需合并为单条 SQL；2) `checkScheduledCleanups` 对每个策略执行 3 次查询，需批量查询优化；3) `updateMetrics` 执行 3 次 COUNT 查询，需合并；4) 添加缺失的数据库索引；5) 慢查询日志阈值配置 | P1 | 40h | BUG-004 | 1) `GetStorageStats` 查询次数 7→1；2) 慢查询（>100ms）数量减少 ≥ 80%；3) 关键查询添加 `EXPLAIN` 验证索引使用；4) 慢查询监控告警配置 | 🔴 待开始 |
| PERF-003 | 数据库连接池优化 | 1) 配置 GORM 连接池参数（`MaxIdleConns`、`MaxOpenConns`、`ConnMaxLifetime`）；2) 实现读写分离支持；3) 连接池监控指标暴露；4) 连接泄漏检测 | P2 | 20h | 无 | 1) 连接池参数根据负载调优；2) 连接池指标接入 Grafana；3) 压测下无连接泄漏；4) 读写分离配置文档 | 🔴 待开始 |

### 5.2 安全加固任务

| ID | 标题 | 描述 | 优先级 | 预估工时 | 依赖 | 验收标准 | 状态 |
|----|------|------|--------|----------|------|----------|------|
| SEC-003 | 安全审计与加固 | 1) 执行 OWASP Top 10 全面审计；2) 添加速率限制中间件；3) 添加 CORS 安全配置；4) 敏感数据加密存储（数据库字段级加密）；5) 审计日志完善；6) 依赖漏洞扫描（`govulncheck`）；7) 密码策略增强（最小长度、复杂度、历史检查） | P1 | 80h | SEC-001, SEC-002, FEAT-011 | 1) OWASP 审计报告无高危/严重漏洞；2) 速率限制配置生效；3) CORS 仅允许合法域名；4) `govulncheck` 无已知漏洞；5) 密码策略满足等保要求 | 🔴 待开始 |

### 5.3 Phase 4 里程碑

| 里程碑 | 时间节点 | 交付物 |
|--------|----------|--------|
| M4.1 | Week 15-16 | PERF-001 缓存体系上线 |
| M4.2 | Week 17-18 | PERF-002、PERF-003 数据库优化完成 |
| M4.3 | Week 19-20 | SEC-003 安全审计完成，漏洞修复验证 |

---

## 6. Phase 5：高级特性与生产就绪 (Week 21-24)

> **目标**: 实现高级监控特性，完善部署自动化，补充生产级文档。  
> **准入条件**: Phase 4 全部完成  
> **准出条件**: 生产环境就绪，文档完备，系统可交付

### 6.1 高级特性任务

| ID | 标题 | 描述 | 优先级 | 预估工时 | 依赖 | 验收标准 | 状态 |
|----|------|------|--------|----------|------|----------|------|
| FEAT-013 | 高级监控仪表盘 | 1) 实时数据流监控大屏（WebSocket 推送）；2) 站点概览仪表盘（发电量、收益、告警统计）；3) 设备健康度仪表盘（在线率、故障率、维护计划）；4) 自定义仪表盘布局（用户可配置）；5) 数据导出（PDF/Excel） | P2 | 60h | PERF-001, BUG-005 | 1) 仪表盘数据实时刷新延迟 < 2s；2) 支持 ≥ 100 个站点同时监控；3) 自定义布局保存/加载正常；4) 数据导出功能完整 | 🔴 待开始 |
| FEAT-014 | 生产部署自动化 | 1) Helm Chart 完善（含资源限制、HPA、PDB）；2) 蓝绿部署/金丝雀发布策略；3) 数据库迁移自动化（零停机）；4) 配置管理（Nacos 集成）；5) 健康检查端点完善；6) 优雅关闭处理 | P2 | 40h | SEC-003 | 1) Helm 部署一键完成；2) 金丝雀发布可按比例切流；3) 数据库迁移零停机；4) 健康检查覆盖所有依赖；5) 优雅关闭无数据丢失 | 🔴 待开始 |

### 6.2 文档任务

| ID | 标题 | 描述 | 优先级 | 预估工时 | 依赖 | 验收标准 | 状态 |
|----|------|------|--------|----------|------|----------|------|
| DOC-001 | API 文档完善 | 1) Swagger 注解补全（所有 Handler）；2) 请求/响应示例；3) 错误码文档；4) 认证授权说明；5) API 变更日志 | P3 | 40h | FEAT-011, FEAT-012 | 1) Swagger UI 可正常访问且文档完整；2) 所有 API 有请求/响应示例；3) 错误码文档覆盖所有业务错误 | 🔴 待开始 |
| DOC-002 | 运维手册 | 1) 部署架构图；2) 配置项说明；3) 监控告警配置指南；4) 故障排查手册；5) 备份恢复流程；6) 容量规划指南；7) 安全加固指南 | P3 | 40h | FEAT-014, SEC-003 | 1) 运维手册覆盖所有生产场景；2) 故障排查手册包含常见问题解决方案；3) 新运维人员可独立完成部署 | 🔴 待开始 |

### 6.3 Phase 5 里程碑

| 里程碑 | 时间节点 | 交付物 |
|--------|----------|--------|
| M5.1 | Week 21-22 | FEAT-013 监控仪表盘上线 |
| M5.2 | Week 23 | FEAT-014 部署自动化完成 |
| M5.3 | Week 24 | DOC-001、DOC-002 文档交付，项目验收 |

---

## 7. 关键路径分析

### 7.1 关键路径

```
BUG-001 (DryRun修复) → BUG-003 (事务支持) → REFACTOR-001 (doCleanup重构)
                                              ↓
SEC-001 (SQL注入) → SEC-002 (输入验证) → TEST-001~013 (Handler测试) → FEAT-001~010 (功能完善)
                                        ↓
BUG-005 (WebSocket) → FEAT-011 (JWT认证) → FEAT-012 (RBAC授权) → PERF-001 (缓存) → SEC-003 (安全审计)
```

### 7.2 可并行任务

| 并行组 | 任务 | 说明 |
|--------|------|------|
| 并行组 A | BUG-001, BUG-002, BUG-004, BUG-005 | Phase 1 内部无依赖的 Bug 修复 |
| 并行组 B | SEC-001, SEC-002 | 安全修复可并行推进 |
| 并行组 C | TEST-001~013 | Handler 层测试可并行编写 |
| 并行组 D | TEST-014~023 | Service 层测试可并行编写 |
| 并行组 E | FEAT-004~010 | 各模块功能实现可并行 |

---

## 8. 资源规划

### 8.1 团队配置建议

| 角色 | 人数 | 职责范围 |
|------|------|----------|
| 后端开发 (高级) | 2 | Phase 1 Bug/安全修复、Phase 3 认证授权、Phase 4 性能优化 |
| 后端开发 (中级) | 3 | Phase 2 测试编写、Phase 3 功能实现 |
| DevOps 工程师 | 1 | Phase 4 连接池/监控、Phase 5 部署自动化 |
| QA 工程师 | 1 | 全程测试验证、安全审计辅助 |
| 技术文档 | 1 | Phase 5 文档编写 |

### 8.2 工时分布

```
Phase 1: ████████░░░░░░░░░░░░  180h (12.7%)
Phase 2: ████████████████░░░░  420h (29.6%)
Phase 3: ████████████████░░░░  440h (31.0%)
Phase 4: ████████░░░░░░░░░░░░  200h (14.1%)
Phase 5: ████████░░░░░░░░░░░░  180h (12.7%)
         总计: 1,420 小时
```

---

## 9. 风险管理

### 9.1 风险登记册

| 编号 | 风险描述 | 影响 | 概率 | 缓解策略 | 负责人 |
|------|----------|------|------|----------|--------|
| R-001 | DryRun 修复可能影响现有清理逻辑 | 高 | 中 | 充分的 DryRun 场景测试，灰度发布 | 后端高级 |
| R-002 | SQL 注入修复可能破坏现有 API 兼容性 | 中 | 低 | 白名单校验宽松过渡期，API 兼容性测试 | 后端高级 |
| R-003 | 测试覆盖率目标可能无法按时达成 | 中 | 中 | 优先覆盖核心路径，非核心路径降低目标 | QA + 后端 |
| R-004 | RBAC 实现复杂度超出预期 | 高 | 中 | 先实现基础角色权限，细粒度权限迭代推进 | 后端高级 |
| R-005 | 缓存引入一致性问题 | 高 | 中 | 主动失效 + TTL 兜底，缓存监控告警 | 后端高级 |
| R-006 | WebSocket 重构影响在线用户 | 高 | 低 | 优雅迁移方案，新旧逻辑兼容过渡期 | 后端高级 |
| R-007 | 安全审计发现新的高危漏洞 | 高 | 中 | 预留 20% 缓冲工时处理审计发现 | 全团队 |

### 9.2 应急预案

| 场景 | 触发条件 | 应急措施 |
|------|----------|----------|
| Phase 1 延期 | Week 3 未完成 P0 任务 | 增加高级开发资源，暂停 Phase 2 非关键任务 |
| 测试覆盖率不达标 | Week 7 覆盖率 < 40% | 降低非核心包覆盖率目标至 40%，核心包维持 60% |
| 安全漏洞批量发现 | SEC-003 审计发现 > 5 个高危 | 紧急修复高危漏洞，调整 Phase 5 时间线 |
| 性能优化效果不达预期 | P95 延迟降低 < 30% | 引入专业 DBA 顾问，调整缓存策略 |

---

## 10. 验收标准总览

### 10.1 各阶段准出标准

| 阶段 | 准出标准 |
|------|----------|
| Phase 1 | 1) 所有 P0 Bug 修复并验证；2) SQL 注入漏洞清零；3) 输入验证中间件上线；4) CI 流水线绿灯 |
| Phase 2 | 1) 零覆盖率包清零；2) 整体覆盖率 ≥ 50%；3) Flaky 测试率 < 1%；4) 测试基础设施文档化 |
| Phase 3 | 1) 所有 Handler 功能完整；2) JWT + RBAC 认证授权上线；3) 长函数重构完成（≤ 30 行）；4) 错误响应格式统一 |
| Phase 4 | 1) P95 延迟降低 ≥ 50%；2) 缓存命中率 ≥ 85%；3) 安全审计无高危漏洞；4) 连接池监控就绪 |
| Phase 5 | 1) 监控仪表盘上线；2) 部署自动化完成；3) API 文档完整；4) 运维手册交付；5) 生产环境验证通过 |

### 10.2 最终交付标准

| 维度 | 目标值 | 当前值 | 差距 |
|------|--------|--------|------|
| 综合代码质量评分 | ≥ 8.0 | 5.3 | +2.7 |
| 测试覆盖率 | ≥ 60% | ~15% | +45% |
| P0/P1 缺陷数 | 0 | 6 | -6 |
| 安全漏洞（高危+） | 0 | 7 | -7 |
| API 功能完整率 | 100% | ~75% | +25% |
| P95 接口延迟 | < 200ms | ~500ms | -300ms |

---

## 附录 A：任务状态说明

| 状态标识 | 含义 |
|----------|------|
| 🔴 待开始 | 任务尚未启动 |
| 🟡 进行中 | 任务正在执行 |
| 🟢 已完成 | 任务已完成并通过验收 |
| 🔵 已阻塞 | 任务因依赖或外部原因阻塞 |
| ⚪ 已取消 | 任务经评估后取消 |

## 附录 B：优先级定义

| 优先级 | 定义 | 响应时间 |
|--------|------|----------|
| P0 | 紧急：系统崩溃、数据丢失、安全漏洞 | 立即修复 (24h 内) |
| P1 | 高：功能不可用、性能严重下降 | 本迭代内修复 |
| P2 | 中：功能受限、体验不佳 | 下个迭代修复 |
| P3 | 低：优化建议、文档完善 | 排期处理 |

## 附录 C：关联文档

| 文档 | 路径 | 说明 |
|------|------|------|
| 需求分析 | [01_requirements_analysis.md](01_requirements_analysis.md) | 项目需求基线 |
| 需求评审 | [02_requirements_review.md](02_requirements_review.md) | 需求评审记录 |
| 实施计划 | [03_implementation_plan.md](03_implementation_plan.md) | 原始实施计划 |
| 测试计划 | [04_test_plan.md](04_test_plan.md) | 测试策略与计划 |
| 代码审查 | [05_code_review.md](05_code_review.md) | 代码审查报告（本任务列表基准） |
| 安全审计 | [security-audit-report.md](security-audit-report.md) | 安全审计报告 |
| 系统架构 | [system-architecture-v2.md](system-architecture-v2.md) | 系统架构文档 |
