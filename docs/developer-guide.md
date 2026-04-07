# 新能源监控系统 - 开发指南

## 文档信息

| 项目 | 内容 |
|------|------|
| 项目名称 | 新能源在线监控系统 |
| 文档版本 | v1.0.0 |
| 编写日期 | 2026-04-07 |
| 文档状态 | 正式发布 |
| 维护团队 | 开发团队 |

---

## 目录

1. [开发环境搭建](#1-开发环境搭建)
2. [代码规范](#2-代码规范)
3. [测试指南](#3-测试指南)
4. [提交规范](#4-提交规范)
5. [开发流程](#5-开发流程)
6. [最佳实践](#6-最佳实践)

---

## 1. 开发环境搭建

### 1.1 系统要求

#### 操作系统

| 操作系统 | 版本要求 | 架构 |
|----------|----------|------|
| Windows | 10/11 | x86_64 |
| macOS | 12+ | x86_64/ARM64 |
| Linux | Ubuntu 20.04+/CentOS 7+ | x86_64 |

#### 硬件要求

| 组件 | 最低配置 | 推荐配置 |
|------|----------|----------|
| CPU | 4核 | 8核+ |
| 内存 | 8GB | 16GB+ |
| 磁盘 | 50GB SSD | 100GB+ SSD |

### 1.2 必需软件安装

#### Go 环境

**版本要求**: Go 1.24+

**安装步骤**:

```bash
# macOS
brew install go

# Linux
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Windows
# 下载安装包: https://go.dev/dl/
# 运行安装程序并按照提示完成安装
```

**验证安装**:

```bash
go version
# 输出: go version go1.24.x darwin/amd64
```

**环境变量配置**:

```bash
# ~/.bashrc 或 ~/.zshrc
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
export GO111MODULE=on
export GOPROXY=https://goproxy.cn,direct
```

#### Node.js 环境

**版本要求**: Node.js 18+

**安装步骤**:

```bash
# macOS
brew install node

# Linux
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# Windows
# 下载安装包: https://nodejs.org/
# 运行安装程序并按照提示完成安装
```

**验证安装**:

```bash
node --version
# 输出: v18.x.x
npm --version
# 输出: 9.x.x
```

#### Docker 环境

**版本要求**: Docker 24.0+, Docker Compose 2.20+

**安装步骤**:

```bash
# macOS
brew install docker docker-compose

# Linux
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Windows
# 下载 Docker Desktop: https://www.docker.com/products/docker-desktop
# 运行安装程序并按照提示完成安装
```

**验证安装**:

```bash
docker --version
# 输出: Docker version 24.x.x
docker-compose --version
# 输出: Docker Compose version v2.20.x
```

#### 数据库客户端

**PostgreSQL 客户端**:

```bash
# macOS
brew install postgresql

# Linux
sudo apt-get install postgresql-client

# Windows
# 下载安装包: https://www.postgresql.org/download/windows/
```

**Redis 客户端**:

```bash
# macOS
brew install redis

# Linux
sudo apt-get install redis-tools

# Windows
# 使用 WSL 或下载 Windows 版本
```

### 1.3 开发工具推荐

#### IDE/编辑器

**推荐**: Visual Studio Code

**必装插件**:

| 插件 | 用途 |
|------|------|
| Go | Go 语言支持 |
| Vue - Official | Vue 3 支持 |
| TypeScript Vue Plugin (Volar) | TypeScript 支持 |
| ESLint | 代码检查 |
| Prettier | 代码格式化 |
| GitLens | Git 增强 |
| Docker | Docker 支持 |
| REST Client | API 测试 |
| Thunder Client | API 测试 |
| Database Client | 数据库管理 |

**Go 开发插件**:

| 插件 | 用途 |
|------|------|
| Go | Go 语言支持 |
| Go Nightly | Go 最新特性 |
| gopls | Go 语言服务器 |
| Go Test Explorer | 测试管理 |
| Go Coverage | 覆盖率显示 |

**前端开发插件**:

| 插件 | 用途 |
|------|------|
| Vue - Official | Vue 3 支持 |
| Volar | Vue Language Features |
| ESLint | 代码检查 |
| Prettier | 代码格式化 |
| Auto Close Tag | 自动闭合标签 |
| Auto Rename Tag | 自动重命名标签 |

#### 其他工具

| 工具 | 用途 |
|------|------|
| Postman | API 测试 |
| DBeaver | 数据库管理 |
| RedisInsight | Redis 管理 |
| k9s | Kubernetes 管理 |
| Lens | Kubernetes IDE |

### 1.4 项目配置

#### 克隆项目

```bash
git clone <repository-url>
cd new-energy-monitoring
```

#### 安装依赖

**后端依赖**:

```bash
# 下载 Go 依赖
go mod download
go mod tidy
```

**前端依赖**:

```bash
cd web
npm install
```

#### 配置开发环境

```bash
# 复制配置文件
cp configs/config-dev.yaml configs/config.yaml

# 编辑配置文件
vim configs/config.yaml
```

**关键配置项**:

```yaml
server:
  mode: debug  # 开发模式

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: nem_system

redis:
  addrs:
    - localhost:6379

logging:
  level: debug  # 日志级别
```

#### 启动开发服务

**启动基础设施**:

```bash
# 启动 PostgreSQL, Redis, Kafka 等
docker-compose up -d postgres redis kafka

# 查看服务状态
docker-compose ps
```

**运行数据库迁移**:

```bash
# 执行迁移脚本
make migrate-up
# 或手动执行
docker exec -i nem-postgres psql -U postgres -d nem_system < scripts/migrations/001_init_schema.sql
```

**启动后端服务**:

```bash
# 方式1: 使用 Makefile
make run-api

# 方式2: 直接运行
go run ./cmd/api-server/main.go

# 方式3: 编译后运行
make build
./bin/api-server
```

**启动前端服务**:

```bash
cd web
npm run dev
```

**访问服务**:

- 前端应用: http://localhost:3001
- API 服务: http://localhost:8080
- Swagger 文档: http://localhost:8080/swagger/index.html

### 1.5 常见问题

#### Go 依赖下载慢

```bash
# 设置代理
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn
```

#### npm 依赖下载慢

```bash
# 设置淘宝镜像
npm config set registry https://registry.npmmirror.com
```

#### Docker 启动失败

```bash
# 检查 Docker 服务状态
docker info

# 重启 Docker 服务
# macOS/Windows: 重启 Docker Desktop
# Linux: sudo systemctl restart docker
```

---

## 2. 代码规范

### 2.1 Go 代码规范

#### 项目结构

```
new-energy-monitoring/
├── cmd/                    # 应用入口
│   ├── api-server/        # API 服务入口
│   ├── collector/         # 采集服务入口
│   └── ...
├── internal/              # 内部代码
│   ├── api/              # API 层 (Handler)
│   ├── application/      # 应用层 (Service)
│   ├── domain/           # 领域层 (Entity, Repository)
│   └── infrastructure/   # 基础设施层
├── pkg/                   # 公共包
│   ├── collector/        # 采集器
│   ├── protocol/         # 协议实现
│   └── ...
├── configs/               # 配置文件
├── scripts/               # 脚本文件
└── tests/                 # 测试文件
```

#### 命名规范

**包命名**:

- 全小写，不使用下划线或驼峰
- 简短、有意义
- 避免使用常见名如 `util`, `common`

```go
// 好的命名
package collector
package alarm
package storage

// 不好的命名
package collectorService
package alarm_utils
package common
```

**文件命名**:

- 全小写，使用下划线分隔
- 文件名应描述其内容

```go
// 好的命名
alarm_service.go
alarm_service_test.go
alarm_rule_repository.go

// 不好的命名
AlarmService.go
alarmService.go
```

**变量命名**:

- 驼峰命名法
- 导出变量首字母大写
- 私有变量首字母小写
- 缩写词保持一致大小写

```go
// 好的命名
var deviceCount int
var DeviceCount int  // 导出
var httpClient *http.Client
var url string
var id string

// 不好的命名
var device_count int
var devicecount int
var HTTPClient *http.Client
var URL string
```

**函数命名**:

- 驼峰命名法
- 函数名应描述其行为
- 返回布尔值的函数以 `Is`, `Has`, `Can` 开头

```go
// 好的命名
func CreateDevice(ctx context.Context, device *Device) error
func GetDeviceByID(ctx context.Context, id string) (*Device, error)
func IsDeviceOnline(deviceID string) bool
func HasPermission(userID, permission string) bool

// 不好的命名
func create_device(ctx context.Context, device *Device) error
func device_get(id string) (*Device, error)
func check_online(deviceID string) bool
```

**接口命名**:

- 单方法接口以 `er` 结尾
- 接口名应描述其行为

```go
// 好的命名
type Collector interface {
    Collect(ctx context.Context) ([]DataPoint, error)
}

type Notifier interface {
    Notify(ctx context.Context, alarm *Alarm) error
}

type Repository interface {
    Save(ctx context.Context, entity interface{}) error
    FindByID(ctx context.Context, id string) (interface{}, error)
}

// 不好的命名
type CollectorInterface interface { ... }
type ICollector interface { ... }
```

#### 代码风格

**导入分组**:

```go
import (
    // 标准库
    "context"
    "fmt"
    "time"

    // 第三方库
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    // 项目内部包
    "new-energy-monitoring/internal/domain/entity"
    "new-energy-monitoring/pkg/errors"
)
```

**错误处理**:

```go
// 好的错误处理
func (s *Service) GetDevice(ctx context.Context, id string) (*Device, error) {
    device, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to find device: %w", err)
    }
    return device, nil
}

// 不好的错误处理
func (s *Service) GetDevice(ctx context.Context, id string) (*Device, error) {
    device, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, err  // 缺少上下文信息
    }
    return device, nil
}
```

**注释规范**:

```go
// Package collector 提供数据采集功能。
//
// 该包实现了多种工业协议的数据采集，包括 IEC104、Modbus、IEC61850 等。
package collector

// Collector 定义了数据采集器的接口。
//
// 实现该接口可以支持新的采集协议。
type Collector interface {
    // Connect 连接到设备。
    //
    // ctx 用于控制连接超时。
    // 返回连接错误或 nil。
    Connect(ctx context.Context) error

    // Collect 采集数据。
    //
    // 返回采集的数据点列表或错误。
    Collect(ctx context.Context) ([]DataPoint, error)
}

// NewModbusCollector 创建 Modbus 采集器。
//
// 参数:
//   - config: Modbus 配置
//
// 返回:
//   - *ModbusCollector: 采集器实例
//   - error: 创建错误
func NewModbusCollector(config *ModbusConfig) (*ModbusCollector, error) {
    // 实现...
}
```

**Context 使用**:

```go
// 好的实践
func (s *Service) CreateDevice(ctx context.Context, device *Device) error {
    // 使用 context 控制超时
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    return s.repo.Save(ctx, device)
}

// 不好的实践
func (s *Service) CreateDevice(device *Device) error {
    // 缺少 context，无法控制超时
    return s.repo.Save(device)
}
```

**结构体初始化**:

```go
// 好的实践
device := &Device{
    ID:     "device-001",
    Name:   "逆变器1",
    Type:   DeviceTypeInverter,
    Status: DeviceStatusOnline,
}

// 不好的实践
device := &Device{"device-001", "逆变器1", DeviceTypeInverter, DeviceStatusOnline}
```

#### 测试规范

**测试文件命名**:

```
源文件: device_service.go
测试文件: device_service_test.go
```

**测试函数命名**:

```go
// 单元测试
func TestDeviceService_Create(t *testing.T) {}
func TestDeviceService_Create_InvalidInput(t *testing.T) {}
func TestDeviceService_Create_DuplicateID(t *testing.T) {}

// 基准测试
func BenchmarkDeviceService_Create(b *testing.B) {}

// 示例测试
func ExampleDeviceService_Create() {}
```

**测试结构**:

```go
func TestDeviceService_Create(t *testing.T) {
    // Arrange (准备)
    mockRepo := &MockDeviceRepository{}
    service := NewDeviceService(mockRepo)
    device := &Device{
        ID:   "device-001",
        Name: "逆变器1",
    }

    // Act (执行)
    err := service.Create(context.Background(), device)

    // Assert (断言)
    assert.NoError(t, err)
    assert.Equal(t, "device-001", device.ID)
}
```

**表驱动测试**:

```go
func TestDeviceService_Validate(t *testing.T) {
    tests := []struct {
        name    string
        device  *Device
        wantErr bool
    }{
        {
            name: "valid device",
            device: &Device{
                ID:   "device-001",
                Name: "逆变器1",
            },
            wantErr: false,
        },
        {
            name: "empty id",
            device: &Device{
                ID:   "",
                Name: "逆变器1",
            },
            wantErr: true,
        },
        {
            name: "empty name",
            device: &Device{
                ID:   "device-001",
                Name: "",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.device.Validate()
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 2.2 前端代码规范

#### 项目结构

```
web/
├── src/
│   ├── api/              # API 接口
│   ├── assets/           # 静态资源
│   ├── components/       # 公共组件
│   ├── composables/      # 组合式函数
│   ├── directives/       # 自定义指令
│   ├── layouts/          # 布局组件
│   ├── plugins/          # 插件
│   ├── router/           # 路由配置
│   ├── stores/           # 状态管理
│   ├── styles/           # 样式文件
│   ├── types/            # 类型定义
│   ├── utils/            # 工具函数
│   └── views/            # 页面组件
├── public/               # 公共资源
├── tests/                # 测试文件
└── e2e/                  # E2E 测试
```

#### 命名规范

**文件命名**:

- 组件文件: PascalCase
- 其他文件: camelCase

```typescript
// 组件文件
DeviceList.vue
AlarmRuleForm.vue

// 其他文件
deviceService.ts
alarmUtils.ts
```

**组件命名**:

```vue
<!-- 好的命名 -->
<template>
  <DeviceList />
  <AlarmRuleForm />
  <StationMap />
</template>

<!-- 不好的命名 -->
<template>
  <device-list />
  <alarmRuleForm />
  <station_map />
</template>
```

**变量命名**:

```typescript
// 好的命名
const deviceList = ref<Device[]>([])
const isLoading = ref(false)
const totalCount = ref(0)

// 不好的命名
const DeviceList = ref([])
const loading = ref(false)
const count = ref(0)
```

**函数命名**:

```typescript
// 好的命名
const fetchDeviceList = async () => { }
const handleDeviceCreate = () => { }
const validateDeviceForm = () => { }

// 不好的命名
const get_data = async () => { }
const create = () => { }
const check = () => { }
```

#### 代码风格

**Vue 组件结构**:

```vue
<template>
  <!-- 模板内容 -->
</template>

<script setup lang="ts">
// 导入
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

// 类型定义
interface Props {
  deviceId: string
}

// Props
const props = defineProps<Props>()

// 响应式数据
const device = ref<Device | null>(null)
const isLoading = ref(false)

// 计算属性
const deviceName = computed(() => device.value?.name || '')

// 方法
const fetchDevice = async () => {
  isLoading.value = true
  try {
    device.value = await deviceApi.getById(props.deviceId)
  } finally {
    isLoading.value = false
  }
}

// 生命周期
onMounted(() => {
  fetchDevice()
})
</script>

<style scoped>
/* 样式 */
</style>
```

**TypeScript 类型定义**:

```typescript
// 好的类型定义
interface Device {
  id: string
  name: string
  type: DeviceType
  status: DeviceStatus
  createdAt: Date
  updatedAt: Date
}

enum DeviceType {
  Inverter = 'inverter',
  Meter = 'meter',
  Combiner = 'combiner',
}

enum DeviceStatus {
  Online = 'online',
  Offline = 'offline',
  Fault = 'fault',
}

// API 响应类型
interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

interface PageResponse<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}
```

**组合式函数**:

```typescript
// composables/useDevice.ts
import { ref } from 'vue'
import { deviceApi } from '@/api/device'

export function useDevice() {
  const device = ref<Device | null>(null)
  const isLoading = ref(false)
  const error = ref<Error | null>(null)

  const fetchDevice = async (id: string) => {
    isLoading.value = true
    error.value = null
    try {
      device.value = await deviceApi.getById(id)
    } catch (e) {
      error.value = e as Error
    } finally {
      isLoading.value = false
    }
  }

  return {
    device,
    isLoading,
    error,
    fetchDevice,
  }
}
```

**API 请求封装**:

```typescript
// api/device.ts
import request from '@/utils/request'

export const deviceApi = {
  list(params: DeviceQueryParams) {
    return request.get<PageResponse<Device>>('/devices', { params })
  },

  getById(id: string) {
    return request.get<Device>(`/devices/${id}`)
  },

  create(data: CreateDeviceRequest) {
    return request.post<Device>('/devices', data)
  },

  update(id: string, data: UpdateDeviceRequest) {
    return request.put<Device>(`/devices/${id}`, data)
  },

  delete(id: string) {
    return request.delete(`/devices/${id}`)
  },
}
```

#### 样式规范

**使用 SCSS**:

```scss
// 好的样式组织
.device-list {
  padding: 20px;

  &__header {
    display: flex;
    justify-content: space-between;
    margin-bottom: 20px;
  }

  &__table {
    width: 100%;
  }

  &__pagination {
    margin-top: 20px;
    text-align: right;
  }
}

// 使用 CSS 变量
.status-badge {
  padding: 4px 8px;
  border-radius: 4px;

  &--online {
    background-color: var(--el-color-success-light);
    color: var(--el-color-success);
  }

  &--offline {
    background-color: var(--el-color-danger-light);
    color: var(--el-color-danger);
  }
}
```

#### 测试规范

**单元测试**:

```typescript
// __tests__/deviceService.test.ts
import { describe, it, expect, vi } from 'vitest'
import { deviceApi } from '@/api/device'

describe('deviceApi', () => {
  it('should fetch device list', async () => {
    const mockResponse = {
      list: [{ id: '1', name: 'Device 1' }],
      total: 1,
    }

    vi.spyOn(request, 'get').mockResolvedValue(mockResponse)

    const result = await deviceApi.list({ page: 1, pageSize: 10 })

    expect(result).toEqual(mockResponse)
    expect(request.get).toHaveBeenCalledWith('/devices', {
      params: { page: 1, pageSize: 10 },
    })
  })
})
```

**组件测试**:

```typescript
// __tests__/DeviceList.test.ts
import { mount } from '@vue/test-utils'
import DeviceList from '@/views/device/DeviceList.vue'

describe('DeviceList', () => {
  it('should render device list', () => {
    const wrapper = mount(DeviceList, {
      global: {
        mocks: {
          $route: { query: {} },
        },
      },
    })

    expect(wrapper.find('.device-list').exists()).toBe(true)
  })
})
```

---

## 3. 测试指南

### 3.1 后端测试

#### 单元测试

**运行单元测试**:

```bash
# 运行所有测试
make test
# 或
go test ./... -v

# 运行特定包的测试
go test ./internal/application/service -v

# 运行特定测试函数
go test -run TestDeviceService_Create -v

# 运行匹配模式的测试
go test -run "TestDevice" -v
```

**生成覆盖率报告**:

```bash
# 生成覆盖率报告
make test-coverage

# 检查覆盖率阈值
make test-coverage-check

# 查看覆盖率详情
go tool cover -func=coverage.out

# 生成 HTML 报告
go tool cover -html=coverage.out -o coverage.html
```

**测试覆盖率要求**:

| 模块 | 最低覆盖率 |
|------|-----------|
| 领域层 (domain) | 80% |
| 应用层 (application) | 80% |
| 基础设施层 (infrastructure) | 70% |
| API 层 (api) | 70% |
| 整体覆盖率 | 80% |

#### 集成测试

**运行集成测试**:

```bash
# 运行集成测试
make test-integration
# 或
bash tests/integration_test.sh
```

**集成测试配置**:

```go
// tests/test_config.go
func SetupTestDB(t *testing.T) *gorm.DB {
    dsn := "host=localhost port=5432 user=postgres password=postgres dbname=nem_test sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed to connect database: %v", err)
    }

    // 运行迁移
    db.AutoMigrate(&entity.Device{}, &entity.Station{})

    // 清理数据
    t.Cleanup(func() {
        db.Exec("TRUNCATE devices, stations CASCADE")
        sqlDB, _ := db.DB()
        sqlDB.Close()
    })

    return db
}
```

#### 性能测试

**运行性能测试**:

```bash
# 运行基准测试
go test -bench=. -benchmem

# 运行特定基准测试
go test -bench=BenchmarkDeviceService -benchmem

# 运行性能测试脚本
bash tests/performance/run_benchmarks.sh
```

**基准测试示例**:

```go
func BenchmarkDeviceService_Create(b *testing.B) {
    service := setupBenchmarkService(b)
    device := &Device{
        ID:   "device-001",
        Name: "逆变器1",
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.Create(context.Background(), device)
    }
}
```

#### Mock 测试

**使用 Mockery 生成 Mock**:

```bash
# 安装 Mockery
go install github.com/vektra/mockery/v2@latest

# 生成 Mock
mockery --name=DeviceRepository --dir=internal/domain/repository --output=tests/mocks
```

**使用 Mock**:

```go
func TestDeviceService_Create(t *testing.T) {
    // 创建 Mock
    mockRepo := &MockDeviceRepository{}
    mockRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

    // 使用 Mock
    service := NewDeviceService(mockRepo)
    err := service.Create(context.Background(), &Device{})

    // 断言
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

### 3.2 前端测试

#### 单元测试

**运行单元测试**:

```bash
cd web

# 运行所有测试
npm test

# 运行特定文件
npm test deviceService.test.ts

# 运行监视模式
npm run test:watch

# 生成覆盖率报告
npm run test:coverage
```

**测试覆盖率要求**:

| 模块 | 最低覆盖率 |
|------|-----------|
| 工具函数 (utils) | 80% |
| API 层 (api) | 80% |
| Store (stores) | 80% |
| 组合式函数 (composables) | 80% |
| 整体覆盖率 | 80% |

#### 组件测试

**运行组件测试**:

```bash
cd web
npm run test:components
```

**组件测试示例**:

```typescript
import { mount } from '@vue/test-utils'
import DeviceList from '@/views/device/DeviceList.vue'

describe('DeviceList', () => {
  it('should render device list', async () => {
    const wrapper = mount(DeviceList)

    // 等待数据加载
    await wrapper.vm.$nextTick()

    // 断言
    expect(wrapper.find('.device-list').exists()).toBe(true)
  })

  it('should handle device creation', async () => {
    const wrapper = mount(DeviceList)
    const createButton = wrapper.find('.create-button')

    await createButton.trigger('click')

    expect(wrapper.find('.device-form').exists()).toBe(true)
  })
})
```

#### E2E 测试

**运行 E2E 测试**:

```bash
cd web

# 运行所有 E2E 测试
npm run test:e2e

# 运行特定测试文件
npm run test:e2e -- auth.spec.ts

# 运行 UI 模式
npm run test:e2e:ui
```

**E2E 测试示例**:

```typescript
// e2e/auth.spec.ts
import { test, expect } from '@playwright/test'

test.describe('Authentication', () => {
  test('should login successfully', async ({ page }) => {
    await page.goto('/login')

    await page.fill('input[name="username"]', 'admin')
    await page.fill('input[name="password"]', 'admin123')
    await page.click('button[type="submit"]')

    await expect(page).toHaveURL('/dashboard')
  })

  test('should show error with invalid credentials', async ({ page }) => {
    await page.goto('/login')

    await page.fill('input[name="username"]', 'admin')
    await page.fill('input[name="password"]', 'wrong')
    await page.click('button[type="submit"]')

    await expect(page.locator('.error-message')).toBeVisible()
  })
})
```

---

## 4. 提交规范

### 4.1 Git 工作流

#### 分支命名

```
main                # 主分支，生产环境代码
develop             # 开发分支，开发环境代码
feature/xxx         # 功能分支
bugfix/xxx          # Bug 修复分支
hotfix/xxx          # 紧急修复分支
release/v1.0.0      # 发布分支
```

**示例**:

```
feature/device-management
feature/alarm-notification
bugfix/device-list-pagination
hotfix/auth-token-expire
release/v1.2.0
```

#### 分支策略

```
main (生产)
  ↑
  └── release/v1.0.0 (发布)
        ↑
        └── develop (开发)
              ↑
              ├── feature/device-management
              ├── feature/alarm-notification
              └── bugfix/device-list-pagination
```

**工作流程**:

1. 从 `develop` 创建功能分支
2. 开发完成后提交 Pull Request
3. 代码审查通过后合并到 `develop`
4. 发布时从 `develop` 创建 `release` 分支
5. 测试通过后合并到 `main` 并打标签

### 4.2 Commit 规范

#### Commit Message 格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

#### Type 类型

| Type | 说明 | 示例 |
|------|------|------|
| feat | 新功能 | feat(device): 添加设备批量导入功能 |
| fix | Bug 修复 | fix(auth): 修复 Token 过期时间计算错误 |
| docs | 文档更新 | docs(api): 更新 API 文档 |
| style | 代码格式 | style(lint): 修复代码格式问题 |
| refactor | 重构 | refactor(service): 重构设备服务代码 |
| perf | 性能优化 | perf(query): 优化设备查询性能 |
| test | 测试 | test(device): 添加设备服务单元测试 |
| chore | 构建/工具 | chore(docker): 更新 Docker 配置 |
| revert | 回滚 | revert: 回滚设备导入功能 |

#### Scope 范围

| Scope | 说明 |
|-------|------|
| api | API 相关 |
| device | 设备模块 |
| alarm | 告警模块 |
| auth | 认证授权 |
| collector | 采集模块 |
| compute | 计算模块 |
| ui | 前端 UI |
| db | 数据库 |
| config | 配置 |

#### Subject 主题

- 简短描述，不超过 50 个字符
- 使用祈使句，首字母小写
- 不以句号结尾

**好的示例**:

```
feat(device): 添加设备批量导入功能
fix(auth): 修复 Token 过期时间计算错误
docs(api): 更新设备 API 文档
```

**不好的示例**:

```
添加设备批量导入功能
Fix: 修复了 Token 过期时间计算错误。
update docs
```

#### Body 正文

- 详细描述改动内容
- 解释为什么做这个改动
- 可以分多行

```
feat(device): 添加设备批量导入功能

- 支持 Excel 和 CSV 格式导入
- 支持数据验证和错误提示
- 支持导入进度显示

Closes #123
```

#### Footer 页脚

- 关联 Issue
- Breaking Changes
- 其他备注

```
BREAKING CHANGE: 设备 API 返回格式变更

Closes #123
Related #124
```

#### 完整示例

```
feat(device): 添加设备批量导入功能

- 支持 Excel 和 CSV 格式导入
- 支持数据验证和错误提示
- 支持导入进度显示
- 添加导入历史记录

实现细节:
1. 使用 excelize 库解析 Excel 文件
2. 使用 encoding/csv 解析 CSV 文件
3. 使用 goroutine 并发处理数据
4. 使用 WebSocket 推送导入进度

性能:
- 1000 条数据导入耗时约 2 秒
- 内存占用峰值约 50MB

Closes #123
Related #124
```

### 4.3 Pull Request 规范

#### PR 标题

格式: `[Type] Brief description`

```
[Feature] 添加设备批量导入功能
[Fix] 修复 Token 过期时间计算错误
[Refactor] 重构设备服务代码
```

#### PR 描述模板

```markdown
## 变更类型

- [ ] 新功能 (Feature)
- [ ] Bug 修复 (Bug Fix)
- [ ] 重构 (Refactor)
- [ ] 性能优化 (Performance)
- [ ] 文档更新 (Documentation)
- [ ] 测试 (Test)

## 变更说明

<!-- 详细描述本次变更的内容 -->

## 变更原因

<!-- 说明为什么需要这个变更 -->

## 测试情况

- [ ] 单元测试已通过
- [ ] 集成测试已通过
- [ ] 手动测试已完成

## 影响范围

<!-- 说明这个变更会影响哪些模块 -->

## 检查清单

- [ ] 代码符合规范
- [ ] 已添加必要的注释
- [ ] 已更新相关文档
- [ ] 已添加测试用例
- [ ] 无新增警告
- [ ] 测试覆盖率达标

## 关联 Issue

Closes #123
Related #124

## 截图

<!-- 如果是 UI 变更，请提供截图 -->
```

#### Code Review 规范

**审查要点**:

1. **代码质量**
   - 代码是否符合规范
   - 是否有明显的 Bug
   - 是否有性能问题
   - 是否有安全隐患

2. **设计合理性**
   - 是否符合架构设计
   - 是否有过度设计
   - 是否有重复代码

3. **测试完整性**
   - 是否有足够的测试
   - 测试覆盖率是否达标
   - 测试用例是否合理

4. **文档完整性**
   - 是否有必要的注释
   - 是否更新了 API 文档
   - 是否更新了用户文档

**审查流程**:

1. 至少需要 1 位审查者批准
2. 所有 CI 检查必须通过
3. 解决所有审查意见
4. Squash 合并到目标分支

---

## 5. 开发流程

### 5.1 功能开发流程

```
1. 领取任务
   ↓
2. 创建分支
   git checkout -b feature/xxx
   ↓
3. 编写代码
   - 编写实现代码
   - 编写单元测试
   - 编写文档
   ↓
4. 本地测试
   make test
   make lint
   ↓
5. 提交代码
   git add .
   git commit -m "feat(xxx): xxx"
   git push origin feature/xxx
   ↓
6. 创建 PR
   - 填写 PR 模板
   - 关联 Issue
   ↓
7. Code Review
   - 修复审查意见
   - 通过 CI 检查
   ↓
8. 合并代码
   - Squash 合并
   - 删除分支
```

### 5.2 Bug 修复流程

```
1. 确认 Bug
   - 复现 Bug
   - 记录复现步骤
   ↓
2. 创建 Issue
   - 描述 Bug 现象
   - 说明复现步骤
   - 标注优先级
   ↓
3. 创建分支
   git checkout -b bugfix/xxx
   ↓
4. 修复 Bug
   - 定位问题原因
   - 编写修复代码
   - 编写回归测试
   ↓
5. 本地测试
   make test
   make lint
   ↓
6. 提交代码
   git commit -m "fix(xxx): xxx"
   ↓
7. 创建 PR
   - 说明修复内容
   - 关联 Issue
   ↓
8. Code Review
   ↓
9. 合并代码
```

### 5.3 发布流程

```
1. 创建发布分支
   git checkout -b release/v1.0.0 develop
   ↓
2. 版本号更新
   - 更新 package.json
   - 更新 go.mod
   ↓
3. 测试验证
   - 运行所有测试
   - 执行性能测试
   - 进行安全扫描
   ↓
4. 文档更新
   - 更新 CHANGELOG
   - 更新用户文档
   - 更新 API 文档
   ↓
5. 合并到 main
   git checkout main
   git merge release/v1.0.0
   ↓
6. 打标签
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ↓
7. 构建部署
   - 构建 Docker 镜像
   - 推送到镜像仓库
   - 部署到生产环境
   ↓
8. 合并回 develop
   git checkout develop
   git merge release/v1.0.0
   ↓
9. 删除发布分支
   git branch -d release/v1.0.0
```

---

## 6. 最佳实践

### 6.1 代码质量

#### 使用 Linter

**Go 代码检查**:

```bash
# 安装 golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 运行检查
make lint
# 或
golangci-lint run
```

**前端代码检查**:

```bash
cd web

# 运行 ESLint
npm run lint

# 自动修复
npm run lint:fix
```

#### 代码审查清单

**Go 代码审查**:

- [ ] 是否遵循 Go 代码规范
- [ ] 是否正确处理错误
- [ ] 是否使用 context 控制超时
- [ ] 是否避免全局变量
- [ ] 是否有并发安全问题
- [ ] 是否有资源泄漏
- [ ] 是否有性能问题
- [ ] 是否有足够的测试

**前端代码审查**:

- [ ] 是否遵循 TypeScript 规范
- [ ] 是否正确使用 Vue 3 Composition API
- [ ] 是否正确处理异步操作
- [ ] 是否避免内存泄漏
- [ ] 是否有性能问题
- [ ] 是否有可访问性问题
- [ ] 是否有足够的测试

### 6.2 性能优化

#### 后端性能优化

**数据库优化**:

```go
// 使用索引
// bad
db.Where("station_id = ?", stationID).Find(&devices)

// good
db.Where("station_id = ?", stationID).
    Index("idx_devices_station_id").
    Find(&devices)

// 批量操作
// bad
for _, device := range devices {
    db.Create(&device)
}

// good
db.CreateInBatches(devices, 100)

// 预加载关联
// bad
db.Find(&stations)
for _, station := range stations {
    db.Where("station_id = ?", station.ID).Find(&station.Devices)
}

// good
db.Preload("Devices").Find(&stations)
```

**缓存优化**:

```go
// 使用 Redis 缓存
func (s *Service) GetDevice(ctx context.Context, id string) (*Device, error) {
    // 先从缓存获取
    cached, err := s.redis.Get(ctx, "device:"+id).Result()
    if err == nil {
        var device Device
        json.Unmarshal([]byte(cached), &device)
        return &device, nil
    }

    // 从数据库获取
    device, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // 写入缓存
    data, _ := json.Marshal(device)
    s.redis.Set(ctx, "device:"+id, data, time.Hour)

    return device, nil
}
```

#### 前端性能优化

**组件优化**:

```vue
<script setup lang="ts">
// 使用 computed 缓存计算结果
const filteredDevices = computed(() => {
  return devices.value.filter(d => d.status === 'online')
})

// 使用 shallowRef 减少响应式开销
const largeList = shallowRef<Device[]>([])

// 使用 v-memo 缓存渲染
</script>

<template>
  <div v-for="device in filteredDevices" :key="device.id" v-memo="[device.status]">
    {{ device.name }}
  </div>
</template>
```

**懒加载**:

```typescript
// 路由懒加载
const routes = [
  {
    path: '/devices',
    component: () => import('@/views/device/DeviceList.vue'),
  },
]

// 组件懒加载
const HeavyComponent = defineAsyncComponent(() =>
  import('@/components/HeavyComponent.vue')
)
```

### 6.3 安全最佳实践

#### 后端安全

**输入验证**:

```go
// 使用 validator 验证输入
type CreateDeviceRequest struct {
    Name     string `json:"name" validate:"required,min=1,max=100"`
    Type     string `json:"type" validate:"required,oneof=inverter meter combiner"`
    Capacity float64 `json:"capacity" validate:"required,gt=0"`
}

func (h *Handler) CreateDevice(c *gin.Context) {
    var req CreateDeviceRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    if err := validate.Struct(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 处理请求...
}
```

**SQL 注入防护**:

```go
// 使用参数化查询
// bad
query := fmt.Sprintf("SELECT * FROM devices WHERE id = '%s'", id)
db.Raw(query).Scan(&devices)

// good
db.Where("id = ?", id).Find(&devices)
```

**XSS 防护**:

```go
// 对用户输入进行转义
import "html"

func sanitize(input string) string {
    return html.EscapeString(input)
}
```

#### 前端安全

**XSS 防护**:

```vue
<template>
  <!-- 使用 v-text 而不是 v-html -->
  <div v-text="userInput"></div>

  <!-- 如果必须使用 v-html，先进行转义 -->
  <div v-html="sanitize(userInput)"></div>
</template>

<script setup lang="ts">
import DOMPurify from 'dompurify'

const sanitize = (html: string) => {
  return DOMPurify.sanitize(html)
}
</script>
```

**CSRF 防护**:

```typescript
// 在请求中添加 CSRF Token
request.interceptors.request.use(config => {
  const token = getCsrfToken()
  if (token) {
    config.headers['X-CSRF-Token'] = token
  }
  return config
})
```

### 6.4 文档规范

#### 代码注释

**Go 注释**:

```go
// Package collector 提供数据采集功能。
//
// 该包实现了多种工业协议的数据采集，包括 IEC104、Modbus、IEC61850 等。
// 使用示例:
//
//	collector := NewModbusCollector(config)
//	if err := collector.Connect(ctx); err != nil {
//	    log.Fatal(err)
//	}
//	data, err := collector.Collect(ctx)
package collector

// Device 表示一个设备实体。
type Device struct {
    // ID 是设备的唯一标识符
    ID string `json:"id"`

    // Name 是设备的名称
    Name string `json:"name"`

    // Status 表示设备的当前状态
    // 可选值: online, offline, fault
    Status DeviceStatus `json:"status"`
}
```

**TypeScript 注释**:

```typescript
/**
 * 设备服务
 *
 * 提供设备的增删改查功能
 *
 * @example
 * ```typescript
 * const device = await deviceApi.getById('device-001')
 * console.log(device.name)
 * ```
 */
export const deviceApi = {
  /**
   * 获取设备列表
   *
   * @param params - 查询参数
   * @returns 设备列表和总数
   */
  list(params: DeviceQueryParams): Promise<PageResponse<Device>> {
    return request.get('/devices', { params })
  },

  /**
   * 根据 ID 获取设备
   *
   * @param id - 设备 ID
   * @returns 设备信息
   * @throws {NotFoundError} 设备不存在
   */
  getById(id: string): Promise<Device> {
    return request.get(`/devices/${id}`)
  },
}
```

#### API 文档

**使用 Swagger**:

```go
// CreateDevice godoc
// @Summary 创建设备
// @Description 创建一个新的设备
// @Tags 设备管理
// @Accept json
// @Produce json
// @Param device body CreateDeviceRequest true "设备信息"
// @Success 201 {object} Device
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /devices [post]
// @Security BearerAuth
func (h *Handler) CreateDevice(c *gin.Context) {
    // 实现...
}
```

**生成 Swagger 文档**:

```bash
# 生成文档
make swagger

# 查看 Swagger UI
make swagger-serve
```

---

## 附录

### A. 常用命令速查

#### Go 命令

```bash
# 运行测试
go test ./... -v

# 运行测试并生成覆盖率
go test -coverprofile=coverage.out ./...

# 查看覆盖率
go tool cover -func=coverage.out

# 生成 HTML 覆盖率报告
go tool cover -html=coverage.out

# 格式化代码
go fmt ./...

# 静态检查
go vet ./...

# 运行 linter
golangci-lint run

# 下载依赖
go mod download

# 整理依赖
go mod tidy

# 查看依赖图
go mod graph

# 运行程序
go run ./cmd/api-server

# 编译程序
go build -o bin/api-server ./cmd/api-server
```

#### Make 命令

```bash
# 查看所有可用命令
make help

# 构建所有服务
make build

# 运行测试
make test

# 运行单元测试
make test-unit

# 运行集成测试
make test-integration

# 生成覆盖率报告
make test-coverage

# 检查覆盖率阈值
make test-coverage-check

# 运行 linter
make lint

# 格式化代码
make fmt

# 清理构建产物
make clean

# 构建 Docker 镜像
make docker-build

# 启动 Docker 容器
make docker-up

# 停止 Docker 容器
make docker-down

# 生成 Wire 代码
make wire

# 生成 Swagger 文档
make swagger
```

#### 前端命令

```bash
# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 构建生产版本
npm run build

# 运行测试
npm test

# 运行测试并生成覆盖率
npm run test:coverage

# 运行 E2E 测试
npm run test:e2e

# 运行 linter
npm run lint

# 自动修复 linter 问题
npm run lint:fix

# 格式化代码
npm run format
```

### B. 故障排查

#### Go 开发问题

**问题1: 依赖下载失败**

```bash
# 设置代理
go env -w GOPROXY=https://goproxy.cn,direct

# 清理缓存
go clean -modcache

# 重新下载
go mod download
```

**问题2: 测试失败**

```bash
# 查看详细输出
go test -v -run TestName

# 检查测试环境
go test -v -count=1 ./...

# 清理测试缓存
go clean -testcache
```

#### 前端开发问题

**问题1: 依赖安装失败**

```bash
# 清理缓存
npm cache clean --force

# 删除 node_modules
rm -rf node_modules package-lock.json

# 重新安装
npm install
```

**问题2: 构建失败**

```bash
# 清理构建缓存
npm run clean

# 检查 Node 版本
node --version

# 检查依赖版本
npm outdated
```

---

## 文档维护

本文档由开发团队维护，如有问题或建议，请联系：

- 邮箱：dev@example.com
- 文档仓库：https://github.com/new-energy-monitoring/docs

**最后更新**: 2026-04-07
