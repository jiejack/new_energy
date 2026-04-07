# 新能源在线监控系统

## 项目简介

新能源在线监控系统是一个面向光伏、风电等新能源电站的综合监控平台，提供设备数据采集、实时监控、告警管理、智能分析等功能。

## 技术栈

### 后端
- **语言**: Go 1.24
- **框架**: Gin
- **ORM**: GORM
- **数据库**: PostgreSQL 18.3
- **缓存**: Redis 8.6.2
- **消息队列**: Kafka 4.2.0, NATS
- **注册中心**: Nacos
- **流处理**: Flink 2.2

### 前端
- **框架**: Vue 3 + TypeScript
- **UI组件**: Element Plus
- **状态管理**: Pinia
- **构建工具**: Vite

### 基础设施
- **容器化**: Docker, Docker Compose
- **编排**: Kubernetes
- **MQTT**: EMQX 6.2.0
- **对象存储**: RustFS

## 项目结构

```
new-energy-monitoring/
├── cmd/                    # 应用入口
│   ├── api-server/        # API服务
│   ├── collector/         # 数据采集服务
│   ├── alarm/             # 告警服务
│   ├── compute/           # 计算服务
│   ├── ai/                # AI服务
│   └── scheduler/         # 调度服务
├── internal/              # 内部代码
│   ├── api/              # API层
│   ├── application/      # 应用层
│   ├── domain/           # 领域层
│   └── infrastructure/   # 基础设施层
├── pkg/                   # 公共包
│   ├── collector/        # 采集器
│   ├── protocol/         # 协议实现
│   ├── storage/          # 存储组件
│   └── ...
├── web/                   # 前端代码
├── configs/               # 配置文件
├── scripts/               # 脚本文件
├── deployments/           # 部署配置
│   ├── kubernetes/       # K8s配置
│   └── docker/           # Docker配置
└── docs/                  # 文档
```

## 快速开始

### 环境要求

- Go 1.24+
- Node.js 18+
- Docker & Docker Compose
- PostgreSQL 18+
- Redis 8+

### 本地开发

1. **克隆项目**
```bash
git clone <repository-url>
cd new-energy-monitoring
```

2. **启动基础设施**
```bash
# 使用 Docker Compose 启动所有服务
docker-compose up -d postgres redis kafka nacos emqx rustfs flink
```

3. **运行数据库迁移**
```bash
# 执行迁移脚本
docker exec -i nem-postgres psql -U postgres -d nem_system < scripts/migrations/001_init_schema.sql
```

4. **启动后端服务**
```bash
go run ./cmd/api-server/main.go
```

5. **启动前端服务**
```bash
cd web
npm install
npm run dev
```

### 访问服务

| 服务 | 地址 | 说明 |
|------|------|------|
| 前端应用 | http://localhost:3001 | Vue 前端 |
| API服务 | http://localhost:8080 | 后端 API |
| Swagger文档 | http://localhost:8080/swagger | API 文档 |
| EMQX Dashboard | http://localhost:18083 | MQTT管理 |
| Nacos控制台 | http://localhost:8848/nacos | 服务注册 |
| Flink Dashboard | http://localhost:8081 | 流处理管理 |

默认账号:
- EMQX: admin / admin123
- Nacos: admin / admin123

## 服务组件

| 组件 | 版本 | 端口 | 说明 |
|------|------|------|------|
| PostgreSQL | 18.3-alpine | 5432 | 关系数据库 |
| Redis | 8.6.2-alpine | 6379 | 缓存服务 |
| Kafka | 4.2.0 | 9092 | 消息队列 |
| NATS | alpine | 4222 | 消息队列 |
| Nacos | v0.8.2 | 8848 | 注册中心 |
| EMQX | 6.2.0 | 1883, 18083 | MQTT代理 |
| RustFS | latest | 9000, 9001 | 对象存储 |
| Flink | 2.2-java21 | 8081 | 流处理 |

## API 文档

启动后端服务后，访问 http://localhost:8080/swagger/index.html 查看完整的 API 文档。

主要 API 模块:
- `/api/v1/stations` - 电站管理
- `/api/v1/devices` - 设备管理
- `/api/v1/points` - 采集点管理
- `/api/v1/alarms` - 告警管理
- `/api/v1/data` - 数据查询
- `/api/v1/ai` - AI智能问答
- `/api/v1/auth` - 认证授权

## 部署

### Docker Compose 部署

```bash
# 构建并启动所有服务
docker-compose up -d --build

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f backend
```

### Kubernetes 部署

```bash
# 创建命名空间
kubectl apply -f k8s/namespace.yaml

# 部署所有服务
kubectl apply -f k8s/

# 查看部署状态
kubectl get all -n new-energy-monitoring
```

详细部署文档请参考 [部署指南](docs/deployment-guide.md)。

## 测试

### 后端测试
```bash
# 运行所有测试
go test ./... -v

# 运行带覆盖率的测试
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### 前端测试
```bash
cd web
npm test
npm run test:coverage
```

## 文档

### 用户文档

- [用户手册](docs/user-manual.md) - 系统使用指南，包括功能介绍、操作步骤和常见问题
- [运维手册](docs/operations-manual.md) - 系统运维指南，包括部署、监控、故障排查等

### 开发文档

- [开发指南](docs/developer-guide.md) - 开发环境搭建、代码规范、测试指南和提交规范
- [系统架构](docs/system-architecture.md) - 系统架构设计和技术选型
- [API 文档](docs/api-documentation.md) - API 接口文档
- [数据库设计](docs/database-design.md) - 数据库表结构设计

### 部署文档

- [部署指南](docs/deployment-guide.md) - 详细部署步骤和配置说明
- [故障排查](docs/troubleshooting.md) - 常见问题和解决方案

## 许可证

MIT License
