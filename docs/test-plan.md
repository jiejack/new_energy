# 新能源监控系统测试计划

## 1. 测试概述

### 1.1 测试目标
- 验证系统功能完整性
- 确保代码质量符合生产标准
- 发现并修复潜在缺陷
- 验证系统性能指标

### 1.2 测试范围

| 模块 | 测试类型 | 覆盖目标 |
|------|----------|----------|
| 前端 | 单元测试 | ≥80% |
| 前端 | 组件测试 | 核心组件100% |
| 前端 | E2E测试 | 主流程100% |
| 后端 | 单元测试 | ≥80% |
| 后端 | 接口测试 | 100% |
| 后端 | 性能测试 | 关键接口 |

### 1.3 测试环境

| 环境 | 配置 |
|------|------|
| 开发环境 | 本地开发机 |
| 测试环境 | Docker Compose |
| CI环境 | GitHub Actions |

### 1.4 测试工具

| 工具 | 用途 |
|------|------|
| Vitest | 前端单元测试 |
| Playwright | E2E测试 |
| Go testing | 后端单元测试 |
| k6 | 后端性能测试 |
| Mockery | Mock工具 |

## 2. 前端测试计划

### 2.1 单元测试

#### 工具配置
- Vitest + Vue Test Utils
- 覆盖率目标: ≥80%

#### 测试范围
- 工具函数测试 (utils/)
- API层测试 (api/)
- Store测试 (stores/)
- 组合式函数测试 (composables/)

#### 测试用例
```typescript
// 示例: request.ts 单元测试
describe('request utils', () => {
  test('should add auth header', () => {
    // 测试Token添加
  })
  test('should handle 401 error', () => {
    // 测试401处理
  })
  test('should refresh token', () => {
    // 测试Token刷新
  })
})
```

### 2.2 组件测试

#### 测试范围
- 公共组件 (CrudTable, FormDialog)
- 业务组件 (StationList, AlarmList, StatCards)

#### 测试用例
```typescript
// 示例: CrudTable组件测试
describe('CrudTable', () => {
  test('should render table with data', () => {
    // 测试数据渲染
  })
  test('should handle pagination', () => {
    // 测试分页
  })
  test('should emit selection change', () => {
    // 测试选择事件
  })
})
```

### 2.3 E2E测试

#### 工具配置
- Playwright
- 测试浏览器: Chromium, Firefox, WebKit

#### 测试场景
1. **登录流程**
   - 正确用户名密码登录成功
   - 错误密码登录失败
   - Token过期自动刷新

2. **监控大屏**
   - 页面加载正常
   - 实时数据更新
   - 告警弹窗显示

3. **配置管理**
   - 区域CRUD操作
   - 电站CRUD操作
   - 设备CRUD操作

4. **数据查询**
   - 历史数据查询
   - 图表渲染
   - 数据导出

## 3. 后端测试计划

### 3.1 单元测试

#### 工具配置
- Go testing + testify
- 覆盖率目标: ≥80%

#### 测试范围
- 领域实体测试
- 应用服务测试
- 基础设施测试
- 协议解析测试

#### 测试用例
```go
// 示例: 设备服务测试
func TestDeviceService_Create(t *testing.T) {
    // 测试设备创建
}

func TestDeviceService_Update(t *testing.T) {
    // 测试设备更新
}

func TestDeviceService_Delete(t *testing.T) {
    // 测试设备删除
}
```

### 3.2 接口测试

#### 测试范围
- 认证接口
- 区域管理接口
- 电站管理接口
- 设备管理接口
- 告警接口
- 数据查询接口

#### 测试用例
```go
// 示例: 设备接口测试
func TestDeviceAPI_List(t *testing.T) {
    // 测试设备列表接口
}

func TestDeviceAPI_Create(t *testing.T) {
    // 测试设备创建接口
}
```

### 3.3 性能测试

#### 工具配置
- k6 + Go benchmark
- 目标指标:
  - API响应时间 P95 < 200ms
  - 并发支持 1000 QPS
  - 内存占用 < 500MB

#### 测试场景
1. **API压力测试**
   - 并发请求测试
   - 持续负载测试
   - 峰值负载测试

2. **数据库性能测试**
   - 批量插入测试
   - 复杂查询测试
   - 索引效率测试

3. **内存泄漏测试**
   - 长时间运行测试
   - 对象生命周期测试

## 4. 测试执行计划

### 4.1 测试阶段

| 阶段 | 内容 | 时间 |
|------|------|------|
| 阶段1 | 单元测试 | Day 1-2 |
| 阶段2 | 组件测试 | Day 2-3 |
| 阶段3 | 接口测试 | Day 3-4 |
| 阶段4 | E2E测试 | Day 4-5 |
| 阶段5 | 性能测试 | Day 5-6 |

### 4.2 缺陷管理

| 严重程度 | 处理时间 |
|----------|----------|
| P0 阻塞 | 立即修复 |
| P1 严重 | 24小时内 |
| P2 一般 | 3天内 |
| P3 轻微 | 下版本 |

### 4.3 测试报告

测试完成后生成以下报告:
- 测试覆盖率报告
- 缺陷统计报告
- 性能测试报告
- 测试总结报告
