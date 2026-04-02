# 测试与部署实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 完成新能源监控系统的全面测试和成功部署，确保系统性能、安全性和稳定性达到预设标准

**Architecture:** 采用分层测试策略（单元测试→集成测试→系统测试→验收测试），使用GitHub Actions实现CI/CD流水线，Docker容器化部署，支持蓝绿发布和快速回滚

**Tech Stack:** 
- 测试: Go testing, Vitest, Playwright, k6
- CI/CD: GitHub Actions, Docker, Kubernetes
- 监控: Prometheus, Grafana, ELK Stack

---

## 项目概况

### 测试范围
- **后端**: Go 1.24, Gin, GORM, PostgreSQL, Redis
- **前端**: Vue 3, TypeScript, Element Plus
- **覆盖率目标**: 单元测试 >80%, 集成测试 >70%

### 部署环境
- **开发环境**: Docker Compose
- **测试环境**: Kubernetes (minikube)
- **生产环境**: Kubernetes (云服务)

---

## Phase 1: 测试计划制定与实施

### Task 1: 运行现有测试并生成覆盖率报告

**Files:**
- Create: `tests/coverage/coverage-report.html`
- Create: `tests/coverage/go-cover.out`
- Create: `tests/coverage/frontend-cover.json`

- [ ] **Step 1: 运行后端单元测试**

```bash
cd e:\ai_work\new-energy-monitoring
go test ./... -coverprofile=tests/coverage/go-cover.out -covermode=atomic -v
```

Expected: 所有测试通过，生成覆盖率文件

- [ ] **Step 2: 生成后端覆盖率HTML报告**

```bash
go tool cover -html=tests/coverage/go-cover.out -o tests/coverage/coverage-report.html
```

Expected: 生成HTML报告

- [ ] **Step 3: 运行前端单元测试**

```bash
cd web
npm run test:coverage
```

Expected: 所有测试通过，生成覆盖率报告

- [ ] **Step 4: 分析覆盖率报告**

检查覆盖率是否达到目标：
- 后端覆盖率 > 80%
- 前端覆盖率 > 70%

- [ ] **Step 5: 提交覆盖率报告**

```bash
git add tests/coverage/
git commit -m "test: add test coverage reports"
```

---

### Task 2: 补充缺失的单元测试

**Files:**
- Create: `internal/api/handler/health_handler_test.go`
- Create: `internal/infrastructure/cache/redis_test.go`
- Create: `web/src/api/__tests__/config.test.ts`

- [ ] **Step 1: 创建健康检查Handler测试**

创建文件 `internal/api/handler/health_handler_test.go`:

```go
package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHealthHandler_Health(t *testing.T) {
	handler := NewHealthHandler()
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)
	
	handler.Health(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "status")
	assert.Contains(t, w.Body.String(), "healthy")
}

func TestHealthHandler_Ready(t *testing.T) {
	handler := NewHealthHandler()
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/ready", nil)
	
	handler.Ready(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
}
```

- [ ] **Step 2: 创建Redis缓存测试**

创建文件 `internal/infrastructure/cache/redis_test.go`:

```go
package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func TestRedisCache_Get(t *testing.T) {
	mockClient := new(MockRedisClient)
	cache := &RedisCache{client: mockClient}
	
	ctx := context.Background()
	key := "test-key"
	expectedValue := "test-value"
	
	mockClient.On("Get", ctx, key).Return(expectedValue, nil)
	
	value, err := cache.Get(ctx, key)
	
	assert.NoError(t, err)
	assert.Equal(t, expectedValue, value)
	mockClient.AssertExpectations(t)
}

func TestRedisCache_Set(t *testing.T) {
	mockClient := new(MockRedisClient)
	cache := &RedisCache{client: mockClient}
	
	ctx := context.Background()
	key := "test-key"
	value := "test-value"
	expiration := time.Hour
	
	mockClient.On("Set", ctx, key, value, expiration).Return(nil)
	
	err := cache.Set(ctx, key, value, expiration)
	
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestRedisCache_Delete(t *testing.T) {
	mockClient := new(MockRedisClient)
	cache := &RedisCache{client: mockClient}
	
	ctx := context.Background()
	keys := []string{"key1", "key2"}
	
	mockClient.On("Del", ctx, keys).Return(nil)
	
	err := cache.Delete(ctx, keys...)
	
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}
```

- [ ] **Step 3: 创建前端API测试**

创建文件 `web/src/api/__tests__/config.test.ts`:

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { configApi } from '../config'
import request from '@/utils/request'

vi.mock('@/utils/request')

describe('Config API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getAllConfigs', () => {
    it('should fetch all configs successfully', async () => {
      const mockConfigs = {
        basic: { system_name: '新能源监控系统' },
        alarm: { default_level: 2 }
      }
      
      vi.mocked(request.get).mockResolvedValue({ data: mockConfigs })
      
      const result = await configApi.getAllConfigs()
      
      expect(request.get).toHaveBeenCalledWith('/api/v1/configs')
      expect(result.data).toEqual(mockConfigs)
    })
  })

  describe('updateConfig', () => {
    it('should update config successfully', async () => {
      const category = 'basic'
      const key = 'system_name'
      const value = '新系统名称'
      
      vi.mocked(request.put).mockResolvedValue({ data: { success: true } })
      
      const result = await configApi.updateConfig(category, key, value)
      
      expect(request.put).toHaveBeenCalledWith(
        `/api/v1/configs/${category}/${key}`,
        { value }
      )
      expect(result.data.success).toBe(true)
    })
  })
})
```

- [ ] **Step 4: 运行新增测试**

```bash
# 后端测试
go test ./internal/api/handler -run TestHealthHandler -v
go test ./internal/infrastructure/cache -run TestRedisCache -v

# 前端测试
cd web && npm run test
```

Expected: 所有新增测试通过

- [ ] **Step 5: 提交新增测试**

```bash
git add internal/api/handler/health_handler_test.go
git add internal/infrastructure/cache/redis_test.go
git add web/src/api/__tests__/config.test.ts
git commit -m "test: add missing unit tests for health handler, redis cache, and config api"
```

---

### Task 3: 创建集成测试

**Files:**
- Create: `tests/integration/api_integration_test.go`
- Create: `tests/integration/database_integration_test.go`
- Create: `tests/integration/setup.go`

- [ ] **Step 1: 创建集成测试设置文件**

创建文件 `tests/integration/setup.go`:

```go
package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/new-energy-monitoring/internal/infrastructure/config"
	"github.com/new-energy-monitoring/internal/infrastructure/persistence"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestEnvironment struct {
	PostgresContainer testcontainers.Container
	RedisContainer    testcontainers.Container
	DB                *persistence.Database
	Config            *config.Config
}

func SetupTestEnvironment(t *testing.T) *TestEnvironment {
	ctx := context.Background()
	
	// 启动PostgreSQL容器
	pgReq := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "test_db",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}
	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: pgReq,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}
	
	pgHost, _ := pgContainer.Host(ctx)
	pgPort, _ := pgContainer.MappedPort(ctx, "5432")
	
	// 启动Redis容器
	redisReq := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: redisReq,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start Redis container: %v", err)
	}
	
	redisHost, _ := redisContainer.Host(ctx)
	redisPort, _ := redisContainer.MappedPort(ctx, "6379")
	
	// 创建配置
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     pgHost,
			Port:     pgPort.Int(),
			User:     "test",
			Password: "test",
			Database: "test_db",
		},
		Redis: config.RedisConfig{
			Host: redisHost,
			Port: redisPort.Int(),
		},
	}
	
	// 连接数据库
	db, err := persistence.NewDatabase(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	
	return &TestEnvironment{
		PostgresContainer: pgContainer,
		RedisContainer:    redisContainer,
		DB:                db,
		Config:            cfg,
	}
}

func (env *TestEnvironment) Cleanup(t *testing.T) {
	ctx := context.Background()
	
	if env.DB != nil {
		env.DB.Close()
	}
	
	if env.PostgresContainer != nil {
		if err := env.PostgresContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate PostgreSQL container: %v", err)
		}
	}
	
	if env.RedisContainer != nil {
		if err := env.RedisContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate Redis container: %v", err)
		}
	}
}
```

- [ ] **Step 2: 创建API集成测试**

创建文件 `tests/integration/api_integration_test.go`:

```go
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/api/handler"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/infrastructure/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type APIIntegrationTestSuite struct {
	suite.Suite
	env      *TestEnvironment
	router   *gin.Engine
}

func (suite *APIIntegrationTestSuite) SetupSuite() {
	suite.env = SetupTestEnvironment(suite.T())
	
	// 初始化路由
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	
	// 初始化依赖
	alarmRepo := persistence.NewAlarmRepository(suite.env.DB)
	alarmService := service.NewAlarmService(alarmRepo)
	alarmHandler := handler.NewAlarmHandler(alarmService)
	
	// 注册路由
	api := suite.router.Group("/api/v1")
	{
		alarms := api.Group("/alarms")
		{
			alarms.GET("", alarmHandler.ListAlarms)
			alarms.GET("/:id", alarmHandler.GetAlarm)
			alarms.PUT("/:id/acknowledge", alarmHandler.AcknowledgeAlarm)
		}
	}
}

func (suite *APIIntegrationTestSuite) TearDownSuite() {
	suite.env.Cleanup(suite.T())
}

func (suite *APIIntegrationTestSuite) TestListAlarms() {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/alarms?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "data")
}

func (suite *APIIntegrationTestSuite) TestCreateAndGetAlarm() {
	// 创建告警
	createReq := map[string]interface{}{
		"device_id": "device-001",
		"level":     2,
		"message":   "测试告警",
	}
	body, _ := json.Marshal(createReq)
	
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alarms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
}

func TestAPIIntegrationSuite(t *testing.T) {
	suite.Run(t, new(APIIntegrationTestSuite))
}
```

- [ ] **Step 3: 创建数据库集成测试**

创建文件 `tests/integration/database_integration_test.go`:

```go
package integration

import (
	"context"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DatabaseIntegrationTestSuite struct {
	suite.Suite
	env *TestEnvironment
}

func (suite *DatabaseIntegrationTestSuite) SetupSuite() {
	suite.env = SetupTestEnvironment(suite.T())
}

func (suite *DatabaseIntegrationTestSuite) TearDownSuite() {
	suite.env.Cleanup(suite.T())
}

func (suite *DatabaseIntegrationTestSuite) TestAlarmCRUD() {
	ctx := context.Background()
	repo := persistence.NewAlarmRepository(suite.env.DB)
	
	// Create
	alarm := entity.NewAlarm(
		"device-001",
		entity.AlarmLevelWarning,
		"测试告警消息",
		"temperature > 80",
	)
	
	err := repo.Create(ctx, alarm)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), alarm.ID)
	
	// Read
	found, err := repo.GetByID(ctx, alarm.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), alarm.Message, found.Message)
	
	// Update
	found.Acknowledge("user-001")
	err = repo.Update(ctx, found)
	assert.NoError(suite.T(), err)
	
	// Delete
	err = repo.Delete(ctx, alarm.ID)
	assert.NoError(suite.T(), err)
}

func (suite *DatabaseIntegrationTestSuite) TestDatabaseHealth() {
	ctx := context.Background()
	
	err := suite.env.DB.Ping(ctx)
	assert.NoError(suite.T(), err)
	
	stats := suite.env.DB.GetStats()
	assert.NotNil(suite.T(), stats)
	assert.Greater(suite.T(), stats.OpenConnections, 0)
}

func TestDatabaseIntegrationSuite(t *testing.T) {
	suite.Run(t, new(DatabaseIntegrationTestSuite))
}
```

- [ ] **Step 4: 运行集成测试**

```bash
cd e:\ai_work\new-energy-monitoring
go test ./tests/integration/... -v -count=1
```

Expected: 所有集成测试通过

- [ ] **Step 5: 提交集成测试**

```bash
git add tests/integration/
git commit -m "test: add integration tests with testcontainers"
```

---

### Task 4: 创建E2E测试

**Files:**
- Create: `web/e2e/system-settings.spec.ts`
- Create: `web/e2e/alarm-management.spec.ts`
- Create: `web/e2e/auth-flow.spec.ts`

- [ ] **Step 1: 创建系统设置E2E测试**

创建文件 `web/e2e/system-settings.spec.ts`:

```typescript
import { test, expect } from '@playwright/test'

test.describe('System Settings', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
    await page.fill('input[name="username"]', 'admin')
    await page.fill('input[name="password"]', 'admin123')
    await page.click('button[type="submit"]')
    await page.waitForURL('/dashboard')
  })

  test('should display system settings page', async ({ page }) => {
    await page.goto('/system/settings')
    
    await expect(page.locator('h2')).toContainText('系统设置')
    await expect(page.locator('.el-tabs__item')).toHaveCount(3)
  })

  test('should update basic settings', async ({ page }) => {
    await page.goto('/system/settings')
    
    // 切换到基本设置标签
    await page.click('.el-tabs__item:has-text("基本设置")')
    
    // 修改系统名称
    await page.fill('input[placeholder="请输入系统名称"]', '测试系统名称')
    
    // 保存设置
    await page.click('button:has-text("保存设置")')
    
    // 验证保存成功
    await expect(page.locator('.el-message--success')).toBeVisible()
  })

  test('should update alarm settings', async ({ page }) => {
    await page.goto('/system/settings')
    
    // 切换到告警设置标签
    await page.click('.el-tabs__item:has-text("告警设置")')
    
    // 修改默认告警级别
    await page.click('.el-select:has-text("请选择告警级别")')
    await page.click('.el-select-dropdown__item:has-text("重要")')
    
    // 保存设置
    await page.click('button:has-text("保存设置")')
    
    // 验证保存成功
    await expect(page.locator('.el-message--success')).toBeVisible()
  })

  test('should change theme', async ({ page }) => {
    await page.goto('/system/settings')
    
    // 切换到显示设置标签
    await page.click('.el-tabs__item:has-text("显示设置")')
    
    // 切换主题
    await page.click('label:has-text("深色")')
    
    // 验证主题已切换
    await expect(page.locator('html')).toHaveClass(/dark/)
  })
})
```

- [ ] **Step 2: 创建告警管理E2E测试**

创建文件 `web/e2e/alarm-management.spec.ts`:

```typescript
import { test, expect } from '@playwright/test'

test.describe('Alarm Management', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
    await page.fill('input[name="username"]', 'admin')
    await page.fill('input[name="password"]', 'admin123')
    await page.click('button[type="submit"]')
    await page.waitForURL('/dashboard')
  })

  test('should display alarm list', async ({ page }) => {
    await page.goto('/alarm/list')
    
    await expect(page.locator('.el-table')).toBeVisible()
    await expect(page.locator('.el-table__row').first()).toBeVisible()
  })

  test('should filter alarms by level', async ({ page }) => {
    await page.goto('/alarm/list')
    
    // 选择告警级别过滤
    await page.click('.el-select:has-text("全部级别")')
    await page.click('.el-select-dropdown__item:has-text("紧急")')
    
    // 等待表格刷新
    await page.waitForTimeout(500)
    
    // 验证过滤结果
    const rows = await page.locator('.el-table__row').count()
    expect(rows).toBeGreaterThan(0)
  })

  test('should acknowledge alarm', async ({ page }) => {
    await page.goto('/alarm/list')
    
    // 点击第一条告警的确认按钮
    await page.locator('.el-table__row').first()
      .locator('button:has-text("确认")').click()
    
    // 确认操作
    await page.click('.el-message-box__btns button:has-text("确定")')
    
    // 验证确认成功
    await expect(page.locator('.el-message--success')).toBeVisible()
  })

  test('should export alarms', async ({ page }) => {
    await page.goto('/alarm/list')
    
    // 点击导出按钮
    const [download] = await Promise.all([
      page.waitForEvent('download'),
      page.click('button:has-text("导出")')
    ])
    
    // 验证下载文件
    expect(download.suggestedFilename()).toContain('.xlsx')
  })
})
```

- [ ] **Step 3: 创建认证流程E2E测试**

创建文件 `web/e2e/auth-flow.spec.ts`:

```typescript
import { test, expect } from '@playwright/test'

test.describe('Authentication Flow', () => {
  test('should login successfully', async ({ page }) => {
    await page.goto('/login')
    
    await page.fill('input[name="username"]', 'admin')
    await page.fill('input[name="password"]', 'admin123')
    await page.click('button[type="submit"]')
    
    await page.waitForURL('/dashboard')
    
    await expect(page.locator('.user-info')).toBeVisible()
  })

  test('should show error for invalid credentials', async ({ page }) => {
    await page.goto('/login')
    
    await page.fill('input[name="username"]', 'invalid')
    await page.fill('input[name="password"]', 'invalid')
    await page.click('button[type="submit"]')
    
    await expect(page.locator('.el-message--error')).toBeVisible()
  })

  test('should logout successfully', async ({ page }) => {
    // 先登录
    await page.goto('/login')
    await page.fill('input[name="username"]', 'admin')
    await page.fill('input[name="password"]', 'admin123')
    await page.click('button[type="submit"]')
    await page.waitForURL('/dashboard')
    
    // 点击退出
    await page.click('.user-dropdown')
    await page.click('button:has-text("退出登录")')
    
    // 验证跳转到登录页
    await page.waitForURL('/login')
  })

  test('should redirect to login for protected routes', async ({ page }) => {
    await page.goto('/system/settings')
    
    // 验证重定向到登录页
    await page.waitForURL('/login')
  })
})
```

- [ ] **Step 4: 运行E2E测试**

```bash
cd web
npx playwright test --reporter=html
```

Expected: 所有E2E测试通过

- [ ] **Step 5: 提交E2E测试**

```bash
git add web/e2e/
git commit -m "test: add E2E tests for system settings, alarm management, and auth flow"
```

---

### Task 5: 创建性能测试

**Files:**
- Create: `tests/performance/api_load_test.go`
- Create: `tests/performance/database_bench_test.go`
- Create: `tests/performance/frontend_perf_test.ts`

- [ ] **Step 1: 创建API负载测试**

创建文件 `tests/performance/api_load_test.go`:

```go
package performance

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func BenchmarkAPIListAlarms(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := setupTestRouter()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/alarms?page=1&page_size=10", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func TestAPIConcurrency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupTestRouter()
	
	concurrentRequests := 100
	var wg sync.WaitGroup
	errors := make(chan error, concurrentRequests)
	
	start := time.Now()
	
	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			req := httptest.NewRequest(http.MethodGet, "/api/v1/alarms", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				errors <- fmt.Errorf("unexpected status code: %d", w.Code)
			}
		}()
	}
	
	wg.Wait()
	elapsed := time.Since(start)
	
	close(errors)
	
	errorCount := 0
	for err := range errors {
		t.Logf("Error: %v", err)
		errorCount++
	}
	
	t.Logf("Concurrent requests: %d", concurrentRequests)
	t.Logf("Total time: %v", elapsed)
	t.Logf("Requests per second: %.2f", float64(concurrentRequests)/elapsed.Seconds())
	t.Logf("Errors: %d", errorCount)
	
	if errorCount > 0 {
		t.Errorf("Had %d errors in concurrent requests", errorCount)
	}
}

func setupTestRouter() *gin.Engine {
	router := gin.New()
	
	router.GET("/api/v1/alarms", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"data": []interface{}{},
			"total": 0,
		})
	})
	
	return router
}
```

- [ ] **Step 2: 创建数据库基准测试**

创建文件 `tests/performance/database_bench_test.go`:

```go
package performance

import (
	"context"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/infrastructure/persistence"
)

func BenchmarkDatabaseCreateAlarm(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()
	
	repo := persistence.NewAlarmRepository(db)
	ctx := context.Background()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		alarm := entity.NewAlarm(
			"device-001",
			entity.AlarmLevelWarning,
			"测试告警",
			"temperature > 80",
		)
		repo.Create(ctx, alarm)
	}
}

func BenchmarkDatabaseQueryAlarms(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()
	
	repo := persistence.NewAlarmRepository(db)
	ctx := context.Background()
	
	// 准备测试数据
	for i := 0; i < 1000; i++ {
		alarm := entity.NewAlarm(
			fmt.Sprintf("device-%d", i),
			entity.AlarmLevelWarning,
			fmt.Sprintf("告警消息 %d", i),
			"condition",
		)
		repo.Create(ctx, alarm)
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		repo.List(ctx, &entity.AlarmFilter{
			Page:     1,
			PageSize: 10,
		})
	}
}

func setupBenchmarkDB(b *testing.B) *persistence.Database {
	// 使用测试数据库配置
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Database: "benchmark_db",
		},
	}
	
	db, err := persistence.NewDatabase(cfg)
	if err != nil {
		b.Fatalf("Failed to connect to database: %v", err)
	}
	
	return db
}
```

- [ ] **Step 3: 创建前端性能测试**

创建文件 `web/tests/performance/frontend_perf_test.ts`:

```typescript
import { describe, it, expect } from 'vitest'
import { render, waitFor } from '@testing-library/vue'
import Dashboard from '@/views/dashboard/index.vue'

describe('Frontend Performance', () => {
  it('should render dashboard within 100ms', async () => {
    const start = performance.now()
    
    render(Dashboard)
    
    await waitFor(() => {
      expect(document.querySelector('.dashboard')).toBeTruthy()
    })
    
    const elapsed = performance.now() - start
    expect(elapsed).toBeLessThan(100)
  })

  it('should handle large data sets efficiently', async () => {
    const largeDataSet = Array.from({ length: 10000 }, (_, i) => ({
      id: i,
      name: `Item ${i}`,
      value: Math.random()
    }))
    
    const start = performance.now()
    
    // 模拟大数据渲染
    const { container } = render(DataTable, {
      props: { data: largeDataSet }
    })
    
    await waitFor(() => {
      expect(container.querySelectorAll('tr').length).toBeGreaterThan(0)
    })
    
    const elapsed = performance.now() - start
    expect(elapsed).toBeLessThan(1000) // 应该在1秒内完成
  })
})
```

- [ ] **Step 4: 运行性能测试**

```bash
# 后端性能测试
go test ./tests/performance/... -bench=. -benchmem

# 前端性能测试
cd web && npm run test:perf
```

Expected: 性能测试通过，响应时间符合预期

- [ ] **Step 5: 提交性能测试**

```bash
git add tests/performance/
git add web/tests/performance/
git commit -m "test: add performance and load tests"
```

---

## Phase 2: CI/CD流水线配置

### Task 6: 创建GitHub Actions工作流

**Files:**
- Create: `.github/workflows/ci.yml`
- Create: `.github/workflows/cd.yml`
- Create: `.github/workflows/test-coverage.yml`

- [ ] **Step 1: 创建CI工作流**

创建文件 `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  backend-test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: test_db
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      
      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.24'
    
    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-
    
    - name: Run tests
      run: go test ./... -v -coverprofile=coverage.out -covermode=atomic
      env:
        DB_HOST: localhost
        DB_PORT: 5432
        DB_USER: test
        DB_PASSWORD: test
        DB_NAME: test_db
        REDIS_HOST: localhost
        REDIS_PORT: 6379
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./coverage.out
        flags: backend

  frontend-test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Node.js
      uses: actions/setup-node@v4
      with:
        node-version: '20'
        cache: 'npm'
        cache-dependency-path: web/package-lock.json
    
    - name: Install dependencies
      working-directory: ./web
      run: npm ci
    
    - name: Run linter
      working-directory: ./web
      run: npm run lint
    
    - name: Run tests
      working-directory: ./web
      run: npm run test:coverage
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./web/coverage/coverage-final.json
        flags: frontend
    
    - name: Build
      working-directory: ./web
      run: npm run build

  e2e-test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Node.js
      uses: actions/setup-node@v4
      with:
        node-version: '20'
        cache: 'npm'
        cache-dependency-path: web/package-lock.json
    
    - name: Install dependencies
      working-directory: ./web
      run: npm ci
    
    - name: Install Playwright
      working-directory: ./web
      run: npx playwright install --with-deps
    
    - name: Run E2E tests
      working-directory: ./web
      run: npx playwright test
    
    - name: Upload test results
      if: always()
      uses: actions/upload-artifact@v3
      with:
        name: playwright-report
        path: web/playwright-report/
```

- [ ] **Step 2: 创建CD工作流**

创建文件 `.github/workflows/cd.yml`:

```yaml
name: CD

on:
  push:
    branches: [ main ]
    tags:
      - 'v*'

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3
    
    - name: Login to Docker Hub
      uses: docker/login-action@v3
      with:
        username: ${{ secrets.DOCKER_USERNAME }}
        password: ${{ secrets.DOCKER_PASSWORD }}
    
    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v5
      with:
        images: ${{ secrets.DOCKER_USERNAME }}/new-energy-monitoring
        tags: |
          type=ref,event=branch
          type=semver,pattern={{version}}
          type=semver,pattern={{major}}.{{minor}}
    
    - name: Build and push backend
      uses: docker/build-push-action@v5
      with:
        context: .
        file: ./Dockerfile.backend
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max
    
    - name: Build and push frontend
      uses: docker/build-push-action@v5
      with:
        context: ./web
        file: ./web/Dockerfile
        push: true
        tags: ${{ secrets.DOCKER_USERNAME }}/new-energy-monitoring-frontend:${{ github.sha }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

  deploy:
    needs: build-and-push
    runs-on: ubuntu-latest
    if: startsWith(github.ref, 'refs/tags/v')
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up kubectl
      uses: azure/setup-kubectl@v3
    
    - name: Configure kubectl
      run: |
        mkdir -p ~/.kube
        echo "${{ secrets.KUBE_CONFIG }}" | base64 -d > ~/.kube/config
    
    - name: Deploy to Kubernetes
      run: |
        kubectl set image deployment/new-energy-monitoring \
          backend=${{ secrets.DOCKER_USERNAME }}/new-energy-monitoring:${{ github.sha }} \
          frontend=${{ secrets.DOCKER_USERNAME }}/new-energy-monitoring-frontend:${{ github.sha }} \
          -n production
    
    - name: Wait for rollout
      run: |
        kubectl rollout status deployment/new-energy-monitoring -n production --timeout=300s
    
    - name: Verify deployment
      run: |
        kubectl get pods -n production
        kubectl get services -n production
```

- [ ] **Step 3: 创建测试覆盖率工作流**

创建文件 `.github/workflows/test-coverage.yml`:

```yaml
name: Test Coverage

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  coverage:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Generate backend coverage
      run: |
        go test ./... -coverprofile=coverage.out -covermode=atomic
        go tool cover -func=coverage.out
    
    - name: Generate frontend coverage
      working-directory: ./web
      run: npm run test:coverage
    
    - name: Upload to Codecov
      uses: codecov/codecov-action@v3
      with:
        files: ./coverage.out,./web/coverage/coverage-final.json
        fail_ci_if_error: true
        verbose: true
    
    - name: Check coverage threshold
      run: |
        # 后端覆盖率检查
        BACKEND_COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        echo "Backend coverage: ${BACKEND_COVERAGE}%"
        
        if (( $(echo "$BACKEND_COVERAGE < 80" | bc -l) )); then
          echo "Backend coverage is below 80%"
          exit 1
        fi
```

- [ ] **Step 4: 提交CI/CD配置**

```bash
git add .github/
git commit -m "ci: add GitHub Actions workflows for CI/CD"
```

---

### Task 7: 创建Docker配置

**Files:**
- Create: `Dockerfile.backend`
- Create: `web/Dockerfile`
- Create: `docker-compose.yml`
- Create: `docker-compose.prod.yml`
- Create: `.dockerignore`

- [ ] **Step 1: 创建后端Dockerfile**

创建文件 `Dockerfile.backend`:

```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git

# 复制go mod文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api-server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# 从builder复制可执行文件
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 运行应用
CMD ["./main"]
```

- [ ] **Step 2: 创建前端Dockerfile**

创建文件 `web/Dockerfile`:

```dockerfile
# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# 复制package文件
COPY package*.json ./
RUN npm ci

# 复制源代码
COPY . .

# 构建应用
RUN npm run build

# Runtime stage
FROM nginx:alpine

# 复制nginx配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 从builder复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 暴露端口
EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

# 启动nginx
CMD ["nginx", "-g", "daemon off;"]
```

- [ ] **Step 3: 创建docker-compose.yml**

创建文件 `docker-compose.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: nem-postgres
    environment:
      POSTGRES_USER: nem
      POSTGRES_PASSWORD: nem123
      POSTGRES_DB: nem_system
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U nem"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: nem-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  backend:
    build:
      context: .
      dockerfile: Dockerfile.backend
    container_name: nem-backend
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=nem
      - DB_PASSWORD=nem123
      - DB_NAME=nem_system
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: unless-stopped

  frontend:
    build:
      context: ./web
      dockerfile: Dockerfile
    container_name: nem-frontend
    ports:
      - "80:80"
    depends_on:
      - backend
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
```

- [ ] **Step 4: 创建.dockerignore**

创建文件 `.dockerignore`:

```
# Git
.git
.gitignore

# Documentation
*.md
docs/

# Tests
tests/
*_test.go
**/__tests__/
**/*.test.ts
**/*.spec.ts
coverage/
.nyc_output/

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Build artifacts
dist/
build/
*.exe
*.exe~
*.dll
*.so
*.dylib

# Dependencies
node_modules/
vendor/

# Environment
.env
.env.local
*.local.yaml

# Logs
*.log
logs/

# Temp
tmp/
temp/
```

- [ ] **Step 5: 提交Docker配置**

```bash
git add Dockerfile.backend web/Dockerfile docker-compose.yml .dockerignore
git commit -m "feat: add Docker configuration for development and production"
```

---

### Task 8: 创建Kubernetes部署配置

**Files:**
- Create: `k8s/namespace.yaml`
- Create: `k8s/configmap.yaml`
- Create: `k8s/secrets.yaml`
- Create: `k8s/deployment-backend.yaml`
- Create: `k8s/deployment-frontend.yaml`
- Create: `k8s/service.yaml`
- Create: `k8s/ingress.yaml`

- [ ] **Step 1: 创建命名空间**

创建文件 `k8s/namespace.yaml`:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: production
  labels:
    name: production
```

- [ ] **Step 2: 创建ConfigMap**

创建文件 `k8s/configmap.yaml`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: nem-config
  namespace: production
data:
  DB_HOST: "postgres-service"
  DB_PORT: "5432"
  DB_NAME: "nem_system"
  REDIS_HOST: "redis-service"
  REDIS_PORT: "6379"
  LOG_LEVEL: "info"
  SERVER_PORT: "8080"
```

- [ ] **Step 3: 创建Secrets**

创建文件 `k8s/secrets.yaml`:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: nem-secrets
  namespace: production
type: Opaque
stringData:
  DB_USER: "nem"
  DB_PASSWORD: "your-secure-password-here"
  JWT_SECRET: "your-jwt-secret-here"
  REDIS_PASSWORD: ""
```

- [ ] **Step 4: 创建后端Deployment**

创建文件 `k8s/deployment-backend.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nem-backend
  namespace: production
  labels:
    app: nem-backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nem-backend
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    metadata:
      labels:
        app: nem-backend
    spec:
      containers:
      - name: backend
        image: your-docker-username/new-energy-monitoring:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: nem-config
        - secretRef:
            name: nem-secrets
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - nem-backend
              topologyKey: kubernetes.io/hostname
```

- [ ] **Step 5: 创建Service**

创建文件 `k8s/service.yaml`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nem-backend-service
  namespace: production
spec:
  selector:
    app: nem-backend
  ports:
  - protocol: TCP
    port: 8080
    targetPort: 8080
  type: ClusterIP

---
apiVersion: v1
kind: Service
metadata:
  name: nem-frontend-service
  namespace: production
spec:
  selector:
    app: nem-frontend
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
  type: ClusterIP

---
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
  namespace: production
spec:
  selector:
    app: postgres
  ports:
  - protocol: TCP
    port: 5432
    targetPort: 5432
  type: ClusterIP

---
apiVersion: v1
kind: Service
metadata:
  name: redis-service
  namespace: production
spec:
  selector:
    app: redis
  ports:
  - protocol: TCP
    port: 6379
    targetPort: 6379
  type: ClusterIP
```

- [ ] **Step 6: 创建Ingress**

创建文件 `k8s/ingress.yaml`:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: nem-ingress
  namespace: production
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - your-domain.com
    secretName: nem-tls
  rules:
  - host: your-domain.com
    http:
      paths:
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: nem-backend-service
            port:
              number: 8080
      - path: /
        pathType: Prefix
        backend:
          service:
            name: nem-frontend-service
            port:
              number: 80
```

- [ ] **Step 7: 提交Kubernetes配置**

```bash
git add k8s/
git commit -m "feat: add Kubernetes deployment configuration"
```

---

## Phase 3: 部署文档与运维指南

### Task 9: 创建部署文档

**Files:**
- Create: `docs/deployment-guide.md`
- Create: `docs/operations-guide.md`
- Create: `docs/troubleshooting.md`

- [ ] **Step 1: 创建部署指南**

创建文件 `docs/deployment-guide.md`:

```markdown
# 部署指南

## 环境要求

### 开发环境
- Docker 20.10+
- Docker Compose 2.0+
- Node.js 20+
- Go 1.24+

### 生产环境
- Kubernetes 1.25+
- Helm 3.0+
- kubectl

## 快速开始

### 使用Docker Compose部署

1. 克隆项目
\`\`\`bash
git clone https://github.com/your-org/new-energy-monitoring.git
cd new-energy-monitoring
\`\`\`

2. 配置环境变量
\`\`\`bash
cp .env.example .env
# 编辑.env文件，设置必要的配置
\`\`\`

3. 启动服务
\`\`\`bash
docker-compose up -d
\`\`\`

4. 验证部署
\`\`\`bash
curl http://localhost:8080/health
\`\`\`

### 使用Kubernetes部署

1. 创建命名空间
\`\`\`bash
kubectl apply -f k8s/namespace.yaml
\`\`\`

2. 创建Secrets
\`\`\`bash
kubectl apply -f k8s/secrets.yaml
\`\`\`

3. 部署应用
\`\`\`bash
kubectl apply -f k8s/
\`\`\`

4. 验证部署
\`\`\`bash
kubectl get pods -n production
kubectl get services -n production
\`\`\`

## 配置说明

### 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| DB_HOST | 数据库主机 | localhost |
| DB_PORT | 数据库端口 | 5432 |
| DB_USER | 数据库用户 | nem |
| DB_PASSWORD | 数据库密码 | - |
| DB_NAME | 数据库名称 | nem_system |
| REDIS_HOST | Redis主机 | localhost |
| REDIS_PORT | Redis端口 | 6379 |
| JWT_SECRET | JWT密钥 | - |
| LOG_LEVEL | 日志级别 | info |

### 数据库迁移

\`\`\`bash
# 运行迁移
go run cmd/api-server/main.go migrate

# 回滚迁移
go run cmd/api-server/main.go migrate:rollback
\`\`\`

## 发布流程

### 版本发布

1. 创建版本标签
\`\`\`bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
\`\`\`

2. CI/CD自动构建和部署

3. 验证部署
\`\`\`bash
kubectl rollout status deployment/nem-backend -n production
\`\`\`

### 回滚操作

\`\`\`bash
# 查看历史版本
kubectl rollout history deployment/nem-backend -n production

# 回滚到上一版本
kubectl rollout undo deployment/nem-backend -n production

# 回滚到指定版本
kubectl rollout undo deployment/nem-backend -n production --to-revision=2
\`\`\`
```

- [ ] **Step 2: 创建运维指南**

创建文件 `docs/operations-guide.md`:

```markdown
# 运维指南

## 日常运维

### 健康检查

\`\`\`bash
# 后端健康检查
curl http://localhost:8080/health

# 数据库连接检查
curl http://localhost:8080/ready

# Kubernetes健康检查
kubectl get pods -n production
kubectl describe pod <pod-name> -n production
\`\`\`

### 日志查看

\`\`\`bash
# Docker日志
docker logs nem-backend -f

# Kubernetes日志
kubectl logs -f deployment/nem-backend -n production

# 多容器日志
kubectl logs -f -l app=nem-backend -n production --all-containers
\`\`\`

### 性能监控

\`\`\`bash
# 查看资源使用
kubectl top pods -n production
kubectl top nodes

# 查看事件
kubectl get events -n production --sort-by='.lastTimestamp'
\`\`\`

## 备份与恢复

### 数据库备份

\`\`\`bash
# 手动备份
kubectl exec -it <postgres-pod> -n production -- \
  pg_dump -U nem nem_system > backup_$(date +%Y%m%d).sql

# 自动备份脚本
./scripts/backup.sh
\`\`\`

### 数据库恢复

\`\`\`bash
# 恢复数据库
kubectl exec -i <postgres-pod> -n production -- \
  psql -U nem nem_system < backup_20240101.sql
\`\`\`

## 扩容与缩容

### 手动扩容

\`\`\`bash
# 扩容到5个副本
kubectl scale deployment/nem-backend -n production --replicas=5

# 自动扩容（HPA）
kubectl autoscale deployment nem-backend -n production \
  --cpu-percent=70 --min=3 --max=10
\`\`\`

### 配置更新

\`\`\`bash
# 更新ConfigMap
kubectl edit configmap nem-config -n production

# 重启Pod使配置生效
kubectl rollout restart deployment/nem-backend -n production
\`\`\`

## 故障处理

### 常见问题

1. **Pod无法启动**
   - 检查镜像是否存在
   - 检查资源限制
   - 查看事件日志

2. **数据库连接失败**
   - 检查数据库服务状态
   - 验证连接配置
   - 检查网络策略

3. **性能下降**
   - 检查资源使用情况
   - 查看慢查询日志
   - 分析性能指标

### 应急响应

1. **服务不可用**
   \`\`\`bash
   # 快速回滚
   kubectl rollout undo deployment/nem-backend -n production
   
   # 重启服务
   kubectl rollout restart deployment/nem-backend -n production
   \`\`\`

2. **数据丢失**
   \`\`\`bash
   # 从备份恢复
   ./scripts/restore.sh backup_20240101.sql
   \`\`\`
```

- [ ] **Step 3: 创建故障排查文档**

创建文件 `docs/troubleshooting.md`:

```markdown
# 故障排查指南

## 常见问题

### 1. 服务启动失败

**症状**: 服务无法启动，日志显示连接错误

**排查步骤**:
\`\`\`bash
# 1. 检查Pod状态
kubectl describe pod <pod-name> -n production

# 2. 查看日志
kubectl logs <pod-name> -n production

# 3. 检查配置
kubectl get configmap nem-config -n production -o yaml
kubectl get secret nem-secrets -n production -o yaml

# 4. 检查依赖服务
kubectl get pods -l app=postgres -n production
kubectl get pods -l app=redis -n production
\`\`\`

**解决方案**:
- 确保数据库和Redis服务正常运行
- 验证配置和密钥正确
- 检查网络连接

### 2. 数据库连接超时

**症状**: 应用日志显示数据库连接超时

**排查步骤**:
\`\`\`bash
# 1. 检查数据库服务
kubectl get svc postgres-service -n production
kubectl get endpoints postgres-service -n production

# 2. 测试连接
kubectl run -it --rm debug --image=postgres:15-alpine --restart=Never -- \
  psql -h postgres-service -U nem -d nem_system

# 3. 检查连接池
kubectl exec -it <backend-pod> -n production -- \
  curl localhost:8080/debug/pprof/goroutine?debug=1
\`\`\`

**解决方案**:
- 增加数据库连接池大小
- 检查数据库负载
- 优化查询性能

### 3. 内存泄漏

**症状**: 服务内存持续增长，最终OOM

**排查步骤**:
\`\`\`bash
# 1. 查看内存使用
kubectl top pods -n production

# 2. 获取内存profile
kubectl port-forward <pod-name> 6060:6060 -n production
curl http://localhost:6060/debug/pprof/heap > heap.out

# 3. 分析profile
go tool pprof heap.out
\`\`\`

**解决方案**:
- 修复内存泄漏代码
- 增加内存限制
- 优化数据结构

### 4. API响应慢

**症状**: API响应时间超过阈值

**排查步骤**:
\`\`\`bash
# 1. 检查资源使用
kubectl top pods -n production

# 2. 查看慢查询日志
kubectl logs <backend-pod> -n production | grep "slow query"

# 3. 分析性能指标
# 访问Prometheus/Grafana查看性能指标
\`\`\`

**解决方案**:
- 优化数据库查询
- 添加缓存
- 增加资源配额

## 监控告警

### 告警规则

1. **服务不可用**
   - 条件: 服务健康检查失败 > 3次
   - 处理: 自动重启Pod

2. **高CPU使用**
   - 条件: CPU使用率 > 80% 持续5分钟
   - 处理: 自动扩容

3. **高内存使用**
   - 条件: 内存使用率 > 85% 持续5分钟
   - 处理: 自动扩容

### 告警通知

配置告警通知渠道：
- 邮件通知
- 钉钉/企业微信
- SMS短信
```

- [ ] **Step 4: 提交文档**

```bash
git add docs/
git commit -m "docs: add deployment, operations, and troubleshooting guides"
```

---

### Task 10: 最终验证与发布

**Files:**
- Create: `scripts/verify-deployment.sh`
- Create: `scripts/smoke-test.sh`

- [ ] **Step 1: 创建部署验证脚本**

创建文件 `scripts/verify-deployment.sh`:

```bash
#!/bin/bash

echo "Starting deployment verification..."

# 检查Pod状态
echo "Checking pod status..."
kubectl get pods -n production

# 检查服务状态
echo "Checking service status..."
kubectl get services -n production

# 检查Ingress状态
echo "Checking ingress status..."
kubectl get ingress -n production

# 健康检查
echo "Running health checks..."
kubectl exec -it deployment/nem-backend -n production -- curl -s http://localhost:8080/health

# 数据库连接检查
echo "Checking database connection..."
kubectl exec -it deployment/nem-backend -n production -- curl -s http://localhost:8080/ready

echo "Deployment verification completed!"
```

- [ ] **Step 2: 创建冒烟测试脚本**

创建文件 `scripts/smoke-test.sh`:

```bash
#!/bin/bash

BASE_URL="http://localhost:8080"

echo "Running smoke tests..."

# 测试健康检查
echo "Testing health endpoint..."
curl -f ${BASE_URL}/health || exit 1

# 测试就绪检查
echo "Testing ready endpoint..."
curl -f ${BASE_URL}/ready || exit 1

# 测试API端点
echo "Testing API endpoints..."
curl -f ${BASE_URL}/api/v1/alarms || exit 1
curl -f ${BASE_URL}/api/v1/devices || exit 1
curl -f ${BASE_URL}/api/v1/stations || exit 1

# 测试配置端点
echo "Testing config endpoint..."
curl -f ${BASE_URL}/api/v1/configs || exit 1

echo "All smoke tests passed!"
```

- [ ] **Step 3: 运行完整测试套件**

```bash
# 运行所有测试
go test ./... -v -coverprofile=coverage.out
cd web && npm run test:all

# 运行集成测试
go test ./tests/integration/... -v

# 运行E2E测试
cd web && npx playwright test
```

Expected: 所有测试通过

- [ ] **Step 4: 生成最终报告**

```bash
# 生成测试覆盖率报告
go tool cover -html=coverage.out -o final-coverage.html

# 生成测试报告
go test ./... -json > test-report.json

# 提交最终报告
git add final-coverage.html test-report.json
git commit -m "test: add final test reports"
```

- [ ] **Step 5: 创建发布标签**

```bash
git tag -a v1.0.0 -m "Release v1.0.0 - Initial production release"
git push origin v1.0.0
```

---

## 验收标准

### 测试验收
- [ ] 单元测试覆盖率 > 80%
- [ ] 集成测试覆盖率 > 70%
- [ ] E2E测试全部通过
- [ ] 性能测试符合预期

### 部署验收
- [ ] CI/CD流水线正常运行
- [ ] Docker镜像构建成功
- [ ] Kubernetes部署成功
- [ ] 健康检查通过

### 文档验收
- [ ] 部署文档完整
- [ ] 运维指南完整
- [ ] 故障排查文档完整

### 安全验收
- [ ] 无高危漏洞
- [ ] 密钥管理安全
- [ ] 网络策略配置正确

---

## 总结

本计划涵盖了从测试到部署的完整流程，包括：
- 5个测试任务（单元测试、集成测试、E2E测试、性能测试）
- 3个CI/CD任务（GitHub Actions、Docker、Kubernetes）
- 2个文档任务（部署文档、运维指南）

**准备好开始执行了吗？**
