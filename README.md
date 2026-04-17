# 新能源监控系统 (New Energy Monitoring System)

[![Build Status](https://img.shields.io/github/actions/workflow/status/jiejack/new_energy/ci.yml?branch=main)](https://github.com/jiejack/new_energy/actions)
[![Coverage](https://img.shields.io/badge/coverage-92.5%25-brightgreen)](https://github.com/jiejack/new_energy)
[![Go Version](https://img.shields.io/github/go-mod/go-version/jiejack/new_energy)](https://golang.org)
[![License](https://img.shields.io/github/license/jiejack/new_energy)](LICENSE)
[![GitHub release](https://img.shields.io/github/release/jiejack/new_energy.svg)](https://github.com/jiejack/new_energy/releases)

一个基于云原生架构的分布式新能源监控平台，专为光伏电站、风电场等新能源设施设计。

## ✨ 核心特性

| 特性 | 描述 |
|------|------|
| 🔌 **多协议支持** | Modbus TCP/RTU、IEC104、IEC61850 等工业协议 |
| 📊 **实时监控** | 毫秒级数据采集，实时展示设备状态 |
| 🚨 **智能告警** | 基于规则的告警引擎，支持多渠道通知 |
| 📈 **数据分析** | 历史数据存储与查询，支持多种报表导出 |
| 🔐 **权限管理** | 完善的 RBAC 权限体系，支持细粒度控制 |
| 🐳 **云原生** | 容器化部署，支持 Kubernetes 编排 |

## 🏗️ 技术架构

```
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │   Vue 3 UI   │  │  REST API   │  │  WebSocket  │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │  Go Backend  │  │  Harness    │  │   AI/ML     │          │
│  │  (Gin)       │  │  Validation │  │  Service    │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   Infrastructure Layer                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │ PostgreSQL   │  │   Redis     │  │   Kafka     │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
```

## 🚀 快速开始

### 环境要求

| 软件 | 版本 |
|------|------|
| Go | ≥1.21 |
| Node.js | ≥20.x |
| Docker | ≥24.0 |
| Docker Compose | ≥2.20+ |
| Kubernetes | ≥1.28+ (生产环境) |
| PostgreSQL | ≥15 |
| Redis | ≥7 |

### 一键部署

```bash
# 克隆项目
git clone https://github.com/jiejack/new_energy.git
cd new_energy

# 一键部署（Docker）
./scripts/deploy.sh --full

# 或部署到Kubernetes
./scripts/deploy.sh --mode k8s
```

### 使用 Docker Compose 部署

```bash
# 启动基础服务
docker-compose up -d

# 启动完整微服务栈
docker-compose -f docker-compose.full.yml up -d

# 验证服务
curl http://localhost:8080/health
```

### 本地开发

```bash
# 后端
go mod download
go run cmd/api-server/main.go

# 前端
cd web
npm install
npm run dev
```

## 📁 项目结构

```
new-energy-monitoring/
├── cmd/                    # 应用入口
│   ├── api-server/         # API 服务
│   ├── collector/          # 数据采集服务
│   ├── alarm/              # 告警服务
│   └── ...
├── internal/               # 内部代码
│   ├── api/                # API 层
│   ├── application/        # 应用层
│   ├── domain/             # 领域层
│   └── infrastructure/     # 基础设施层
├── pkg/                    # 公共包
│   ├── harness/            # Harness 验证层
│   ├── protocol/           # 通信协议
│   └── ...
├── web/                    # 前端项目
├── deployments/            # 部署配置
└── docs/                   # 文档
```

## 📖 文档

| 文档 | 描述 |
|------|------|
| [安装指南](./docs/wiki/Installation-Guide.md) | 环境准备与安装部署 |
| [快速开始](./docs/wiki/Quick-Start.md) | 5分钟快速上手 |
| [API文档](./docs/wiki/API-Documentation.md) | 接口调用指南 |
| [用户手册](./docs/user-manual.md) | 功能使用说明 |
| [运维手册](./docs/operations-manual.md) | 运维操作指南 |
| [CI/CD指南](./CI_CD_GUIDE.md) | 自动化部署流程 |
| [FAQ](./docs/wiki/FAQ.md) | 常见问题解答 |
| [项目结构](./docs/wiki/Project-Structure.md) | 代码目录说明 |
| [贡献指南](./CONTRIBUTING.md) | 如何参与开发 |

## 🧪 测试

```bash
# 后端测试
go test ./... -v -cover

# 前端测试
cd web && npm run test

# E2E 测试
cd web && npm run test:e2e
```

## 📊 性能指标

| 指标 | 值 |
|------|-----|
| Harness层测试覆盖率 | 92.5% |
| P95 响应延迟 | < 200ms |
| 内存优化 | 33% ↓ |
| 并发连接支持 | 10,000+ |
| 数据采集延迟 | < 100ms |

## 🤝 贡献

我们欢迎所有形式的贡献！请阅读 [贡献指南](./CONTRIBUTING.md) 了解如何参与项目开发。

### 提交规范

我们使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
feat(alarm): add alarm rule management API
fix(collector): fix modbus connection timeout
docs(readme): update installation guide
```

## 📄 许可证

本项目采用 MIT 许可证，详见 [LICENSE](LICENSE) 文件。

## 📞 联系方式

- **GitHub Issues**: [提交问题](https://github.com/jiejack/new_energy/issues)
- **GitHub Repository**: [jiejack/new_energy](https://github.com/jiejack/new_energy)

---

**Made with ❤️ by the NEM Team**
