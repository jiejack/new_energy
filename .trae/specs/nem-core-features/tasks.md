# 新能源监控系统 - 核心功能完善 - 实施计划

## [ ] Task 1: 告警规则管理 - 数据库持久化层完善
- **Priority**: P0
- **Depends On**: None
- **Description**:
  - 完善告警规则 Repository 层，实现真正的数据库操作
  - 创建数据库迁移脚本（如需要）
  - 实现完整的 CRUD 数据访问逻辑
- **Acceptance Criteria Addressed**: [AC-1, AC-2]
- **Test Requirements**:
  - `programmatic` TR-1.1: Repository 层单元测试通过率 100%
  - `programmatic` TR-1.2: 数据库迁移脚本执行成功
  - `programmatic` TR-1.3: CRUD 操作正确读写数据库
- **Notes**: 参考现有 device_harness.go 的模式实现
- **Files**:
  - Modify: [internal/infrastructure/persistence/alarm_rule_repository.go](file:///workspace/internal/infrastructure/persistence/alarm_rule_repository.go)
  - Verify: [internal/domain/repository/alarm_rule_repository.go](file:///workspace/internal/domain/repository/alarm_rule_repository.go)
  - Verify: [internal/domain/entity/alarm_rule.go](file:///workspace/internal/domain/entity/alarm_rule.go)

---

## [ ] Task 2: 告警规则管理 - Service 层业务逻辑完善
- **Priority**: P0
- **Depends On**: Task 1
- **Description**:
  - 完善告警规则 Service 层，实现完整的业务逻辑
  - 添加表单验证和错误处理
  - 实现规则启用/禁用功能
- **Acceptance Criteria Addressed**: [AC-1, AC-2]
- **Test Requirements**:
  - `programmatic` TR-2.1: Service 层单元测试通过率 100%
  - `programmatic` TR-2.2: 业务规则验证逻辑正确
  - `programmatic` TR-2.3: 错误处理完整且有意义
- **Files**:
  - Modify: [internal/application/service/alarm_rule_service.go](file:///workspace/internal/application/service/alarm_rule_service.go)
  - Add: [internal/application/service/alarm_rule_service_test.go](file:///workspace/internal/application/service/alarm_rule_service_test.go)

---

## [ ] Task 3: 告警规则管理 - Handler 层 API 完善
- **Priority**: P0
- **Depends On**: Task 2
- **Description**:
  - 完善告警规则 Handler 层，移除模拟数据
  - 正确连接 Service 层
  - 实现完整的 RESTful API
- **Acceptance Criteria Addressed**: [AC-1, AC-2]
- **Test Requirements**:
  - `programmatic` TR-3.1: API 集成测试通过
  - `programmatic` TR-3.2: 所有端点返回正确状态码
  - `programmatic` TR-3.3: 请求/响应格式符合 DTO 规范
- **Files**:
  - Modify: [internal/api/handler/alarm_rule_handler.go](file:///workspace/internal/api/handler/alarm_rule_handler.go)
  - Verify: [cmd/api-server/main.go](file:///workspace/cmd/api-server/main.go#L342-L347) (路由已存在)

---

## [ ] Task 4: 告警规则管理 - 前端页面完善
- **Priority**: P0
- **Depends On**: Task 3
- **Description**:
  - 完善告警规则前端页面
  - 正确调用后端 API
  - 实现完整的表单交互和数据展示
- **Acceptance Criteria Addressed**: [AC-1, AC-2, AC-5]
- **Test Requirements**:
  - `programmatic` TR-4.1: 前端组件测试通过
  - `human-judgement` TR-4.2: 用户交互流畅，无明显 bug
  - `programmatic` TR-4.3: API 调用正确，数据展示准确
- **Files**:
  - Modify: [web/src/views/alarm/rule/index.vue](file:///workspace/web/src/views/alarm/rule/index.vue)
  - Modify: [web/src/api/alarm.ts](file:///workspace/web/src/api/alarm.ts)

---

## [ ] Task 5: 统计报表 - Repository 层数据查询实现
- **Priority**: P1
- **Depends On**: None
- **Description**:
  - 实现统计报表数据查询 Repository
  - 支持按时间、电站、设备等多维度查询
  - 实现高效的数据聚合查询
- **Acceptance Criteria Addressed**: [AC-3]
- **Test Requirements**:
  - `programmatic` TR-5.1: Repository 层单元测试通过
  - `programmatic` TR-5.2: 查询性能满足要求（&lt; 5s）
  - `programmatic` TR-5.3: 数据聚合计算准确
- **Files**:
  - Create/Modify: Repository 层查询接口和实现

---

## [ ] Task 6: 统计报表 - Service 层业务逻辑实现
- **Priority**: P1
- **Depends On**: Task 5
- **Description**:
  - 实现统计报表 Service 层
  - 支持日报、周报、月报等报表类型
  - 计算同比、环比等关键指标
- **Acceptance Criteria Addressed**: [AC-3, AC-4]
- **Test Requirements**:
  - `programmatic` TR-6.1: Service 层单元测试通过率 100%
  - `programmatic` TR-6.2: 统计指标计算准确
  - `programmatic` TR-6.3: 支持多种报表类型
- **Files**:
  - Modify: [internal/application/service/report_service.go](file:///workspace/internal/application/service/report_service.go)

---

## [ ] Task 7: 数据导出 - Excel/CSV 导出功能实现
- **Priority**: P1
- **Depends On**: Task 6
- **Description**:
  - 完善 Excel 和 CSV 导出功能
  - 支持导出统计报表数据
  - 实现文件流式下载
- **Acceptance Criteria Addressed**: [AC-4]
- **Test Requirements**:
  - `programmatic` TR-7.1: 导出功能单元测试通过
  - `programmatic` TR-7.2: 导出文件格式正确
  - `programmatic` TR-7.3: 导出数据完整准确
- **Files**:
  - Modify: [pkg/export/excel.go](file:///workspace/pkg/export/excel.go)
  - Modify: [pkg/export/csv.go](file:///workspace/pkg/export/csv.go)
  - Modify: [internal/application/service/export_service.go](file:///workspace/internal/application/service/export_service.go)

---

## [ ] Task 8: 统计报表 - Handler 层 API 实现
- **Priority**: P1
- **Depends On**: Task 6, Task 7
- **Description**:
  - 实现统计报表和数据导出 API
  - 正确连接 Service 层
  - 实现文件下载响应
- **Acceptance Criteria Addressed**: [AC-3, AC-4]
- **Test Requirements**:
  - `programmatic` TR-8.1: API 集成测试通过
  - `programmatic` TR-8.2: 报表生成接口正常
  - `programmatic` TR-8.3: 导出接口返回正确文件
- **Files**:
  - Create/Modify: [internal/api/handler/report_handler.go](file:///workspace/internal/api/handler/)
  - Verify: [cmd/api-server/main.go](file:///workspace/cmd/api-server/main.go#L357-L359) (路由已存在)

---

## [ ] Task 9: 统计报表 - 前端页面完善
- **Priority**: P1
- **Depends On**: Task 8
- **Description**:
  - 完善统计报表前端页面
  - 集成图表可视化（ECharts）
  - 实现数据导出功能
- **Acceptance Criteria Addressed**: [AC-3, AC-4, AC-5]
- **Test Requirements**:
  - `programmatic` TR-9.1: 前端组件测试通过
  - `human-judgement` TR-9.2: 图表展示正确美观
  - `programmatic` TR-9.3: 导出功能正常工作
- **Files**:
  - Modify: [web/src/views/data/report/index.vue](file:///workspace/web/src/views/data/report/index.vue)

---

## [ ] Task 10: 其他未完成模块 - 前后端集成
- **Priority**: P1
- **Depends On**: Task 4, Task 9
- **Description**:
  - 检查并完善通知配置管理模块
  - 检查并完善权限管理模块
  - 检查并完善操作日志模块
  - 确保所有前端页面都正确连接后端 API
- **Acceptance Criteria Addressed**: [AC-5]
- **Test Requirements**:
  - `programmatic` TR-10.1: 所有 API 调用正常
  - `human-judgement` TR-10.2: 所有页面数据展示正确
  - `programmatic` TR-10.3: 无模拟数据残留
- **Files**:
  - Verify: [web/src/views/alarm/notification/index.vue](file:///workspace/web/src/views/alarm/notification/index.vue)
  - Verify: [web/src/views/system/permission/index.vue](file:///workspace/web/src/views/system/permission/index.vue)
  - Verify: [web/src/views/system/log/index.vue](file:///workspace/web/src/views/system/log/index.vue)
  - 以及其他相关文件

---

## [ ] Task 11: 完整测试与验证
- **Priority**: P0
- **Depends On**: Task 1-10
- **Description**:
  - 运行完整的测试套件
  - 进行集成测试和端到端测试
  - 验证代码覆盖率达标
- **Acceptance Criteria Addressed**: [AC-6]
- **Test Requirements**:
  - `programmatic` TR-11.1: 所有测试通过
  - `programmatic` TR-11.2: 代码覆盖率 ≥ 80%
  - `human-judgement` TR-11.3: 手动测试验证所有功能正常
