# 新能源监控系统 - Docker部署指南

## 概述

本项目提供了完整的Docker容器化部署方案，支持单节点快速部署和生产环境部署。

## 快速开始

### 前置要求

- Docker 20.10+
- Docker Compose 1.29+

### 单节点部署（推荐用于开发测试）

使用`ops/docker/`目录的`docker-compose.yml`进行快速部署：

```bash
# 克隆项目
git clone <repository-url>
cd new-energy-monitoring

# 构建并启动所有服务
docker-compose -f ops/docker/docker-compose.yml up -d

# 查看服务状态
docker-compose -f ops/docker/docker-compose.yml ps

# 查看日志
docker-compose -f ops/docker/docker-compose.yml logs -f
```

### 服务访问地址

部署完成后，可以通过以下地址访问服务：

| 服务 | 地址 | 说明 |
|------|------|------|
| 前端应用 | http://localhost | 新能源监控系统前端 |
| 后端API | http://localhost:8080 | 后端REST API |
| Swagger文档 | http://localhost:8080/swagger/index.html | API文档 |
| Prometheus | http://localhost:9090 | 监控指标 |
| Grafana | http://localhost:3001 | 监控仪表盘（默认账号：admin/admin） |

## 架构说明

### 单节点部署架构

```
┌─────────────────────────────────────────────────────┐
│                     用户浏览器                       │
└──────────────────────┬──────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────┐
│                   Nginx (前端)                       │
│              端口: 80 (Docker映射)                   │
└────────────┬───────────────────────────┬────────────┘
             │                           │
             ▼                           ▼
┌──────────────────────┐    ┌──────────────────────────┐
│   后端API服务        │    │   静态资源/Vue应用       │
│   端口: 8080         │    │   (由Nginx服务)          │
└──────────────────────┘    └──────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────┐
│              监控堆栈 (Prometheus + Grafana)         │
└─────────────────────────────────────────────────────┘
```

### 文件说明

- `ops/docker/Dockerfile`: 后端服务Docker镜像构建文件
- `ops/docker/Dockerfile.frontend`: 前端服务Docker镜像构建文件
- `ops/docker/docker-compose.yml`: 单节点部署配置
- `ops/docker/docker-compose.full.yml`: 生产环境完整部署配置（包含数据库、消息队列等）

## 配置说明

### 后端配置

后端配置文件位于`configs/`目录，可以通过Docker卷挂载进行修改：

```yaml
volumes:
  - ./configs:/app/configs
```

### 环境变量

可以在`docker-compose.yml`中配置以下环境变量：

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| TZ | Asia/Shanghai | 时区设置 |
| GIN_MODE | release | Gin运行模式（debug/release） |

## 高级部署

### 生产环境部署

生产环境建议使用`ops/docker/`目录下的完整配置，包含：

- PostgreSQL数据库
- Redis缓存
- Kafka消息队列
- 多个微服务（API Server、Collector、Alarm、Compute等）

部署步骤：

```bash
cd ops/docker

# 构建并启动所有服务
docker-compose -f docker-compose.full.yml up -d
```

### 自定义构建

如果需要自定义构建，可以分别构建镜像：

```bash
# 构建后端镜像
docker build -t nem-backend:latest -f ops/docker/Dockerfile .

# 构建前端镜像
docker build -t nem-frontend:latest -f ops/docker/Dockerfile.frontend .

# 启动服务
docker-compose -f ops/docker/docker-compose.yml up -d
```

## 故障排查

### 服务无法启动

```bash
# 查看服务状态
docker-compose ps

# 查看具体服务日志
docker-compose logs backend
docker-compose logs frontend
```

### 端口冲突

如果端口被占用，可以修改`docker-compose.yml`中的端口映射：

```yaml
ports:
  - "8081:8080"  # 将主机8081端口映射到容器8080端口
  - "8080:80"    # 将主机8080端口映射到容器80端口
```

### 数据持久化

所有重要数据都通过Docker卷进行持久化：

- `prometheus-data`: Prometheus监控数据
- `grafana-data`: Grafana配置和仪表盘

如需备份，可以使用以下命令：

```bash
# 备份卷数据
docker run --rm -v prometheus-data:/data -v $(pwd):/backup alpine tar czf /backup/prometheus-backup.tar.gz /data
```

## 停止和清理

```bash
# 停止服务
docker-compose -f ops/docker/docker-compose.yml down

# 停止完整服务栈
docker-compose -f ops/docker/docker-compose.full.yml down

# 停止服务并删除卷（谨慎使用）
docker-compose -f ops/docker/docker-compose.yml down -v

# 删除所有镜像
docker rmi $(docker images -q "nem-*")
```

## 技术支持

如有问题，请查看项目文档或提交Issue。