# 测试技能文档

| 字段 | 值 |
|------|------|
| 文档版本 | v1.0.0 |
| 文档名称 | NEM 测试技能文档 |
| 适用项目 | 新能源监控系统（NEM） |
| 文档类型 | 技能规范 |
| 文档负责人 | NEM 技术负责人 |
| 关联文档 | 15_skills_configuration.md、skills/golang-development.md、skills/code-review-skills.md、09_comprehensive_test_plan.md |
| 最后更新 | 2026-06-17 |

---

## 目录

1. [测试技能概述](#1-测试技能概述)
2. [单元测试技能](#2-单元测试技能)
3. [集成测试技能](#3-集成测试技能)
4. [前端测试技能](#4-前端测试技能)
5. [性能测试技能](#5-性能测试技能)
6. [测试覆盖率技能](#6-测试覆盖率技能)
7. [测试最佳实践](#7-测试最佳实践)
8. [常用测试工具速查表](#8-常用测试工具速查表)

---

## 1. 测试技能概述

### 1.1 目的

本文档定义 NEM 项目的测试技能规范，覆盖单元测试、集成测试、前端测试、性能测试与覆盖率要求，为后端（Go）与前端（Vue 3）提供统一的测试方法论与工具链。

### 1.2 测试技术栈

| 类别 | 工具 | 用途 |
|------|------|------|
| Go 断言 | testify/assert | 流畅的断言 API |
| Go Mock | testify/mock | 接口 Mock 生成与验证 |
| Go 要求 | testify/require | 失败立即终止的断言 |
| Go HTTP 测试 | net/http/httptest | HTTP 请求/响应构造 |
| Gin 测试 | gin.TestMode | Gin 测试模式开关 |
| Redis 测试 | miniredis | 内存级 Redis 替身 |
| 数据库测试 | SQLite 内存数据库 | 无外部依赖的数据库测试 |
| 前端单元测试 | Vitest | Vue 3 组件单元测试 |
| 前端 E2E 测试 | Playwright | 浏览器端到端测试 |
| 性能测试 | k6 | HTTP/API 负载与压力测试 |

### 1.3 测试金字塔

NEM 遵循测试金字塔原则，测试数量由多到少、执行速度由快到慢：

```
            /\
           /  \        E2E 测试（少量，关键路径）
          /----\
         /      \      集成测试（适量，模块间协作）
        /--------\
       /          \    单元测试（大量，快速反馈）
      /____________\
```

---

## 2. 单元测试技能

### 2.1 Go testify 使用规范

#### assert 与 require 的选择

| 包 | 行为 | 适用场景 |
|----|------|---------|
| `testify/assert` | 断言失败记录错误，继续执行后续断言 | 需要收集多个失败信息 |
| `testify/require` | 断言失败立即终止当前测试 | 前置条件必须满足时（如初始化失败） |

```go
func TestUserService_Create(t *testing.T) {
    // 前置条件用 require：初始化失败则无需继续
    db := setupTestDB(t)
    require.NotNil(t, db, "测试数据库初始化失败")

    svc := NewUserService(db)
    user, err := svc.Create("alice", "alice@nem.io")

    // 结果断言用 assert：收集所有失败信息
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "alice", user.Name)
}
```

### 2.2 Mock 编写规范

使用 `testify/mock` 对接口进行 Mock，遵循以下规范：

```go
// 定义接口（在 domain 层）
type UserRepository interface {
    FindByID(ctx context.Context, id int64) (*User, error)
}

// 测试中 Mock
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) FindByID(ctx context.Context, id int64) (*User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

func TestUserService_Get(t *testing.T) {
    mockRepo := new(MockUserRepository)
    mockRepo.On("FindByID", mock.Anything, int64(1)).
        Return(&User{ID: 1, Name: "alice"}, nil)

    svc := NewUserService(mockRepo)
    user, err := svc.Get(context.Background(), 1)

    assert.NoError(t, err)
    assert.Equal(t, "alice", user.Name)
    mockRepo.AssertExpectations(t) // 验证 Mock 方法被如期调用
}
```

**Mock 规范要点**：
- Mock 对象定义在测试文件中，不污染生产代码。
- 使用 `mock.Anything` 谨慎，优先匹配具体参数以增强断言强度。
- 测试末尾调用 `AssertExpectations(t)` 验证预期调用是否发生。
- 项目已提供 Mock 辅助工具，位于 `tests/helpers/`（mock_db.go、mock_kafka.go、mock_redis.go）。

### 2.3 表驱动测试

表驱动测试是 Go 推荐的测试组织方式，适合多输入多输出场景：

```go
func TestAlarmRule_Evaluate(t *testing.T) {
    tests := []struct {
        name     string
        rule     *AlarmRule
        value    float64
        expected bool
    }{
        {"低于阈值触发", &AlarmRule{Field: "temp", Op: "<", Threshold: 10}, 5.0, true},
        {"等于阈值不触发", &AlarmRule{Field: "temp", Op: "<", Threshold: 10}, 10.0, false},
        {"高于阈值不触发", &AlarmRule{Field: "temp", Op: "<", Threshold: 10}, 15.0, false},
        {"负值边界", &AlarmRule{Field: "temp", Op: "<", Threshold: 0}, -1.0, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := tt.rule.Evaluate(tt.value)
            assert.Equal(t, tt.expected, got)
        })
    }
}
```

### 2.4 边界值测试

每个函数须覆盖以下边界值：

| 边界类型 | 示例 |
|---------|------|
| 零值 | 0、""、nil、空切片 |
| 极值 | int64 最大/最小值、空指针 |
| 越界 | 数组索引 -1 / len、负数 |
| 单元素 | 长度为 1 的集合 |
| 空集合 | 长度为 0 的切片 / map |
| 临界值 | 阈值等于、阈值 ±1 |

---

## 3. 集成测试技能

### 3.1 测试替身策略

NEM 集成测试使用轻量级替身，避免依赖外部基础设施：

| 组件 | 替身方案 | 说明 |
|------|---------|------|
| PostgreSQL | SQLite 内存数据库 | GORM 兼容，快速启动，测试结束即销毁 |
| Redis | miniredis | 内存级 Redis 实现，支持常用命令 |
| Kafka | Mock（`tests/helpers/mock_kafka.go`） | 接口级 Mock，验证消息发送 |
| HTTP | httptest.Server | 构造模拟 HTTP 服务端 |

### 3.2 Gin 测试模式

测试 HTTP Handler 时，启用 Gin 测试模式以抑制调试日志：

```go
func TestMain(m *testing.M) {
    gin.SetMode(gin.TestMode)
    os.Exit(m.Run())
}

func TestAlarmHandler_List(t *testing.T) {
    // 构造测试路由
    router := gin.New()
    handler := NewAlarmHandler(mockService)
    router.GET("/api/v1/alarms", handler.List)

    // 构造请求
    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/api/v1/alarms?page=1&size=10", nil)

    // 执行
    router.ServeHTTP(w, req)

    // 断言
    assert.Equal(t, http.StatusOK, w.Code)
}
```

### 3.3 集成测试组织

- 集成测试位于 `tests/api/` 与 `tests/integration_test.sh`。
- 每个集成测试应自包含：自行准备数据、执行、清理，不依赖其他测试的副作用。
- 运行脚本：`tests/run_tests.sh`、`tests/coverage.sh`。

---

## 4. 前端测试技能

### 4.1 Vitest 单元测试

前端单元测试使用 Vitest，配置位于 `web/vitest.config.ts`，测试 setup 位于 `web/src/test/setup.ts`。

```typescript
import { describe, it, expect } from 'vitest'
import { useCounter } from '@/composables/useCounter'

describe('useCounter', () => {
  it('应正确递增计数', () => {
    const { count, increment } = useCounter()
    expect(count.value).toBe(0)
    increment()
    expect(count.value).toBe(1)
  })
})
```

### 4.2 组件测试

- 使用 `@vue/test-utils` 挂载组件，验证渲染与交互。
- Mock API 调用，避免真实网络请求。
- 测试用户交互（点击、输入）触发的状态变化。

### 4.3 Playwright E2E 测试

E2E 测试位于 `web/e2e/`，配置位于 `web/playwright.config.ts`，覆盖关键业务路径：

| 测试文件 | 覆盖场景 |
|---------|---------|
| `auth.spec.ts` | 登录认证流程 |
| `dashboard.spec.ts` | 仪表盘展示 |
| `alarm.spec.ts` | 告警列表与详情 |
| `data.spec.ts` | 数据监控 |
| `config.spec.ts` | 配置管理 |

### 4.4 前端测试规范

- 测试文件与源文件同目录或 `__tests__` 下，命名 `*.test.ts` / `*.spec.ts`。
- 优先测试组件行为（用户视角），而非实现细节。
- E2E 测试聚焦关键路径，避免过度覆盖。

---

## 5. 性能测试技能

### 5.1 Go 基准测试

使用 Go 内置 `testing.B` 编写基准测试，位于 `tests/performance/`：

```go
func BenchmarkAlarmRule_Evaluate(b *testing.B) {
    rule := &AlarmRule{Field: "temp", Op: "<", Threshold: 10}
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        rule.Evaluate(5.0)
    }
}
```

### 5.2 k6 负载测试

使用 k6 进行 HTTP/API 负载与压力测试，脚本位于 `scripts/performance/load_test.js`：

```javascript
import http from 'k6/http'
import { check, sleep } from 'k6'

export const options = {
  stages: [
    { duration: '30s', target: 50 },   // 30 秒内升至 50 并发
    { duration: '1m', target: 50 },     // 维持 50 并发 1 分钟
    { duration: '30s', target: 0 },     // 30 秒内降至 0
  ],
}

export default function () {
  const res = http.get('http://localhost:8080/api/v1/alarms')
  check(res, { '状态码 200': (r) => r.status === 200 })
  sleep(1)
}
```

### 5.3 性能测试规范

- 基准测试使用 `b.ResetTimer()` 排除初始化耗时。
- 性能测试报告位于 `tests/performance/`，含 `report.go` 与 `IMPLEMENTATION_SUMMARY.md`。
- 性能基准参见 `docs/performance-benchmark.md`。
- 运行脚本：`tests/performance/run_benchmarks.sh`、`scripts/performance/benchmark.sh`。

---

## 6. 测试覆盖率技能

### 6.1 覆盖率目标

| 范围 | 覆盖率目标 |
|------|----------|
| 核心业务逻辑（domain / service） | ≥ 80% |
| Handler 层 | ≥ 70% |
| 基础设施层 | ≥ 60% |
| 整体 | ≥ 70% |

### 6.2 覆盖率统计

```bash
# Go 覆盖率统计
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# 使用项目脚本
./tests/coverage.sh
```

### 6.3 覆盖率使用原则

- 覆盖率是参考指标，不是唯一目标；高覆盖率不等于高质量。
- 优先覆盖核心业务逻辑与高频变更模块。
- 关注分支覆盖率与边界覆盖，而非仅行覆盖率。
- CI 流水线（`.github/workflows/test-coverage.yml`）自动统计覆盖率。

---

## 7. 测试最佳实践

### 7.1 AAA 模式

每个测试遵循 Arrange-Act-Assert 三段式结构：

```go
func TestUserService_Create(t *testing.T) {
    // Arrange（准备）
    svc := NewUserService(mockRepo)
    expectedName := "alice"

    // Act（执行）
    user, err := svc.Create(expectedName, "alice@nem.io")

    // Assert（断言）
    assert.NoError(t, err)
    assert.Equal(t, expectedName, user.Name)
}
```

### 7.2 测试金字塔

- **单元测试（底层，最多）**：快速、隔离、覆盖所有分支与边界。
- **集成测试（中层，适量）**：验证模块间协作，使用轻量替身。
- **E2E 测试（顶层，最少）**：覆盖关键用户路径，执行慢但真实。

### 7.3 测试隔离

- 每个测试独立运行，不依赖其他测试的执行顺序或副作用。
- 测试自行准备与清理数据，使用 setup/teardown 或表驱动自包含数据。
- 禁止共享可变状态；并发测试须无数据竞争（`go test -race`）。

### 7.4 其他实践

- 测试命名清晰表达意图：`Test<函数>_<场景>_<预期>`。
- 一个测试只验证一个行为点，失败原因明确。
- 避免测试中的逻辑分支（if/switch），保持线性。
- 测试代码同样需可读、可维护，遵循生产代码质量标准。

---

## 8. 常用测试工具速查表

### 8.1 Go 测试工具

| 工具 | 用途 | 命令 / 路径 |
|------|------|------------|
| `go test` | 运行测试 | `go test ./...` |
| `go test -race` | 竞态检测 | `go test -race ./...` |
| `go test -cover` | 覆盖率 | `go test -cover ./...` |
| `go test -bench` | 基准测试 | `go test -bench=. ./...` |
| testify | 断言与 Mock | `github.com/stretchr/testify` |
| httptest | HTTP 测试 | `net/http/httptest` |
| gin.TestMode | Gin 测试模式 | `gin.SetMode(gin.TestMode)` |
| miniredis | Redis 替身 | `github.com/alicebob/miniredis` |
| SQLite | 数据库替身 | `gorm.io/driver/sqlite`（内存模式） |

### 8.2 前端测试工具

| 工具 | 用途 | 命令 / 路径 |
|------|------|------------|
| Vitest | 单元测试 | `web/vitest.config.ts` |
| @vue/test-utils | 组件挂载 | Vue 3 官方测试工具 |
| Playwright | E2E 测试 | `web/playwright.config.ts` |
| `vitest run` | 运行单元测试 | 在 `web/` 下执行 |
| `npx playwright test` | 运行 E2E 测试 | 在 `web/` 下执行 |

### 8.3 性能测试工具

| 工具 | 用途 | 路径 |
|------|------|------|
| testing.B | Go 基准测试 | `tests/performance/*_bench_test.go` |
| k6 | HTTP 负载测试 | `scripts/performance/load_test.js` |
| benchmark.sh | 基准测试脚本 | `scripts/performance/benchmark.sh` |
| run_benchmarks.sh | 批量基准测试 | `tests/performance/run_benchmarks.sh` |

### 8.4 测试辅助资源

| 资源 | 路径 | 说明 |
|------|------|------|
| Mock 数据库 | `tests/helpers/mock_db.go` | 数据库 Mock 辅助 |
| Mock Kafka | `tests/helpers/mock_kafka.go` | Kafka Mock 辅助 |
| Mock Redis | `tests/helpers/mock_redis.go` | Redis Mock 辅助 |
| 测试工具集 | `tests/helpers/test_utils.go` | 通用测试工具 |
| 测试配置 | `tests/test_config.go` | 测试环境配置 |
| 运行脚本 | `tests/run_tests.sh` | 测试运行脚本 |
| 覆盖率脚本 | `tests/coverage.sh` | 覆盖率统计脚本 |

---

## 变更记录表

| 版本 | 日期 | 变更内容 | 变更人 |
|-----|------|---------|-------|
| v1.0.0 | 2026-06-17 | 初始创建：测试技能概述、单元测试（testify/Mock/表驱动/边界值）、集成测试、前端测试、性能测试、测试覆盖率、测试最佳实践、常用测试工具速查表 | AI Agent |
