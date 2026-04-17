# CI/CD 部署指南

本指南介绍了新能源监控系统的完整CI/CD流程，包括Docker分层打包、Kubernetes部署和自动化工作流。

## 目录

- [架构概述](#架构概述)
- [Docker 分层打包](#docker-分层打包)
- [Kubernetes 部署配置](#kubernetes-部署配置)
- [CI/CD 自动化工作流](#cicd-自动化工作流)
- [一键部署脚本](#一键部署脚本)
- [Makefile 使用](#makefile-使用)

## 架构概述

系统采用微服务架构，包含以下组件：

- **api-server**: API网关和主服务
- **collector**: 数据采集服务
- **alarm**: 告警服务
- **compute**: 计算服务
- **ai-service**: AI分析服务
- **scheduler**: 调度服务
- **frontend**: 前端应用
- **监控组件**: Prometheus + Grafana
- **数据存储**: PostgreSQL + Redis + Kafka

## Docker 分层打包

### Dockerfile 列表

项目包含以下分层构建的Dockerfile：

1. [Dockerfile](file:///workspace/ops/docker/Dockerfile) - API Server
2. [Dockerfile.collector](file:///workspace/ops/docker/Dockerfile.collector) - 采集服务
3. [Dockerfile.alarm](file:///workspace/ops/docker/Dockerfile.alarm) - 告警服务
4. [Dockerfile.compute](file:///workspace/ops/docker/Dockerfile.compute) - 计算服务
5. [Dockerfile.ai-service](file:///workspace/ops/docker/Dockerfile.ai-service) - AI服务
6. [Dockerfile.scheduler](file:///workspace/ops/docker/Dockerfile.scheduler) - 调度服务
7. [Dockerfile.frontend](file:///workspace/ops/docker/Dockerfile.frontend) - 前端应用

### 分层构建特点

所有Dockerfile都采用多阶段构建策略：

- **Builder阶段**: 使用完整的Go/Node环境编译代码
- **Runtime阶段**: 使用轻量级alpine镜像运行应用
- **优化**: 利用Docker缓存层加速构建
- **安全**: 运行时镜像最小化攻击面

### 构建示例

```bash
# 构建所有镜像
make docker-build-all

# 指定仓库和标签
DOCKER_REGISTRY=my-registry.com IMAGE_TAG=v1.0.0 make docker-build-all
```

## Kubernetes 部署配置

### K8s 配置文件

Kubernetes配置文件位于 [ops/k8s/](file:///workspace/ops/k8s/) 目录：

1. [01-namespace.yaml](file:///workspace/ops/k8s/01-namespace.yaml) - 命名空间
2. [02-configmap.yaml](file:///workspace/ops/k8s/02-configmap.yaml) - 配置映射
3. [03-secrets.yaml](file:///workspace/ops/k8s/03-secrets.yaml) - 密钥配置
4. [04-postgres.yaml](file:///workspace/ops/k8s/04-postgres.yaml) - PostgreSQL数据库
5. [05-redis.yaml](file:///workspace/ops/k8s/05-redis.yaml) - Redis缓存
6. [06-kafka.yaml](file:///workspace/ops/k8s/06-kafka.yaml) - Kafka消息队列
7. [07-api-server.yaml](file:///workspace/ops/k8s/07-api-server.yaml) - API Server（含HPA）
8. [08-microservices.yaml](file:///workspace/ops/k8s/08-microservices.yaml) - 微服务部署
9. [09-frontend-monitoring.yaml](file:///workspace/ops/k8s/09-frontend-monitoring.yaml) - 前端和监控

### 部署到Kubernetes

```bash
# 完整部署
make k8s-deploy

# 查看状态
make k8s-status

# 查看日志
make k8s-logs

# 删除部署
make k8s-delete
```

## CI/CD 自动化工作流

### GitHub Actions 工作流

项目包含完整的CI/CD工作流：

1. [deploy-pipeline.yml](file:///workspace/.github/workflows/deploy-pipeline.yml) - 完整部署流水线

### 工作流触发条件

- **Push到main/develop分支**: 自动触发测试和部署
- **Pull Request**: 运行测试验证
- **手动触发**: 通过GitHub UI选择环境部署

### 工作流阶段

1. **Test阶段**: 运行后端和前端测试
2. **Build & Push阶段**: 构建并推送Docker镜像
3. **Deploy阶段**: 部署到Kubernetes

### 配置GitHub Secrets

需要在GitHub仓库中配置以下Secrets：

```
DOCKER_REGISTRY      # Docker仓库地址
DOCKER_USERNAME      # Docker用户名
DOCKER_PASSWORD      # Docker密码
KUBE_CONFIG          # Kubernetes kubeconfig内容
```

## 一键部署脚本

### 使用 deploy.sh

项目提供了 [ops/scripts/deploy.sh](file:///workspace/ops/scripts/deploy.sh) 一键部署脚本：

```bash
# 基本用法
./ops/scripts/deploy.sh

# 部署完整微服务栈
./ops/scripts/deploy.sh --full

# 部署到Kubernetes
./ops/scripts/deploy.sh --mode k8s

# 指定仓库和标签
./ops/scripts/deploy.sh --registry my-registry.com --tag v1.0.0

# 仅构建镜像
./ops/scripts/deploy.sh --build-only

# 查看帮助
./ops/scripts/deploy.sh --help
```

### 脚本功能

- 自动检查依赖
- 支持Docker和K8s两种模式
- 支持完整栈或精简栈部署
- 彩色输出和进度提示
- 错误处理和回滚机制

## Makefile 使用

### 主要目标

项目 [Makefile](file:///workspace/Makefile) 提供以下主要目标：

#### Docker相关

```bash
make docker-build-all    # 构建所有Docker镜像
make docker-push-all     # 推送所有Docker镜像
make docker-up          # 启动精简Docker栈
make docker-down        # 停止精简Docker栈
make docker-full-up     # 启动完整微服务栈
make docker-full-down   # 停止完整微服务栈
make docker-logs        # 查看Docker日志
```

#### Kubernetes相关

```bash
make k8s-deploy         # 部署到Kubernetes
make k8s-delete         # 从Kubernetes删除
make k8s-logs           # 查看Kubernetes日志
make k8s-status         # 查看Kubernetes状态
```

#### 一键部署

```bash
make deploy-docker      # Docker完整部署（构建+推送+启动）
make deploy-k8s         # K8s完整部署
make deploy-all         # 完整部署（默认Docker）
```

#### 开发相关

```bash
make build              # 构建所有服务
make test               # 运行所有测试
make lint               # 运行代码检查
make fmt                # 格式化代码
make clean              # 清理构建产物
```

### 环境变量

```bash
# Docker仓库配置
export DOCKER_REGISTRY=my-registry.com
export IMAGE_TAG=v1.0.0

# 使用自定义配置
make docker-build-all
```

## 快速开始

### Docker快速部署

```bash
# 1. 克隆项目
git clone <repo-url>
cd new-energy-monitoring

# 2. 使用脚本一键部署
./ops/scripts/deploy.sh --full

# 3. 访问服务
# 前端: http://localhost
# API: http://localhost:8080
# Grafana: http://localhost:3000
```

### Kubernetes快速部署

```bash
# 1. 配置kubectl
export KUBE_CONFIG=<your-kubeconfig>

# 2. 部署到K8s
./ops/scripts/deploy.sh --mode k8s

# 3. 查看部署状态
make k8s-status
```

## 常见问题

### Docker构建慢

- 确保使用了国内镜像源
- 利用Docker缓存层
- 使用buildx多平台构建

### Kubernetes部署失败

- 检查节点资源是否充足
- 验证镜像拉取权限
- 查看Pod日志定位问题

### CI/CD流水线失败

- 检查GitHub Secrets配置
- 确认Docker仓库可访问
- 查看Actions日志定位错误

## 下一步

- 查看 [DEPLOYMENT.md](file:///workspace/docs/DEPLOYMENT.md) 了解更多部署详情
- 查看 [Kubernetes配置](file:///workspace/ops/k8s/) 自定义部署参数
- 配置 [GitHub Actions](file:///workspace/.github/workflows/) 实现自动化部署
