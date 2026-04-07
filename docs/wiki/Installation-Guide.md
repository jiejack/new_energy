# 安装指南

本文档详细介绍新能源监控系统的安装部署流程。

## 环境要求

### 硬件要求

| 组件 | 最低配置 | 推荐配置 |
|------|----------|----------|
| CPU | 4核 | 8核+ |
| 内存 | 8GB | 16GB+ |
| 存储 | 50GB SSD | 200GB SSD |
| 网络 | 100Mbps | 1Gbps |

### 软件要求

| 软件 | 版本要求 | 用途 |
|------|----------|------|
| Go | ≥1.24 | 后端运行时 |
| Node.js | ≥20.x | 前端构建 |
| PostgreSQL | ≥15 | 主数据库 |
| Redis | ≥7 | 缓存服务 |
| Docker | ≥24.0 | 容器运行时 |
| Docker Compose | ≥2.0 | 容器编排 |

## 安装方式

### 方式一：Docker Compose（推荐）

适合开发测试和中小规模部署。

#### 1. 克隆项目

```bash
git clone https://github.com/jiejack/new_energy.git
cd new_energy
```

#### 2. 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env` 文件，配置以下关键参数：

```env
# 数据库配置
DB_HOST=postgres
DB_PORT=5432
DB_USER=nem
DB_PASSWORD=your_secure_password
DB_NAME=nem_system

# Redis配置
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT配置
JWT_SECRET=your_jwt_secret_key
JWT_EXPIRE=24h

# 服务端口
API_PORT=8080
WEB_PORT=80
```

#### 3. 启动服务

```bash
docker-compose up -d
```

#### 4. 验证安装

```bash
# 检查服务状态
docker-compose ps

# 健康检查
curl http://localhost:8080/health
```

### 方式二：Kubernetes部署

适合生产环境大规模部署。

#### 1. 前置条件

- Kubernetes集群 ≥1.25
- kubectl 已配置
- Helm 3.x 已安装

#### 2. 创建命名空间

```bash
kubectl create namespace nem-system
```

#### 3. 配置Secrets

```bash
kubectl create secret generic nem-secrets \
  --from-literal=DB_PASSWORD=your_password \
  --from-literal=JWT_SECRET=your_jwt_secret \
  -n nem-system
```

#### 4. Helm部署

```bash
cd deployments/kubernetes/helm

# 开发环境
helm install nem . -f values-dev.yaml -n nem-system

# 生产环境
helm install nem . -f values-prod.yaml -n nem-system
```

#### 5. 验证部署

```bash
kubectl get pods -n nem-system
kubectl get services -n nem-system
```

### 方式三：源码编译

适合开发调试。

#### 1. 安装依赖

```bash
# 后端依赖
go mod download

# 前端依赖
cd web
npm install
```

#### 2. 数据库初始化

```bash
# 创建数据库
createdb -U postgres nem_system

# 执行迁移
go run cmd/migrate/main.go
```

#### 3. 启动后端

```bash
go run cmd/api-server/main.go
```

#### 4. 启动前端

```bash
cd web
npm run dev
```

## 配置说明

### 数据库配置

| 参数 | 说明 | 默认值 |
|------|------|--------|
| DB_HOST | 数据库主机 | localhost |
| DB_PORT | 数据库端口 | 5432 |
| DB_USER | 数据库用户 | nem |
| DB_PASSWORD | 数据库密码 | - |
| DB_NAME | 数据库名称 | nem_system |
| DB_MAX_CONN | 最大连接数 | 100 |
| DB_IDLE_CONN | 空闲连接数 | 10 |

### Redis配置

| 参数 | 说明 | 默认值 |
|------|------|--------|
| REDIS_HOST | Redis主机 | localhost |
| REDIS_PORT | Redis端口 | 6379 |
| REDIS_PASSWORD | Redis密码 | - |
| REDIS_DB | 数据库索引 | 0 |

### 服务配置

| 参数 | 说明 | 默认值 |
|------|------|--------|
| API_PORT | API服务端口 | 8080 |
| WEB_PORT | Web服务端口 | 80 |
| LOG_LEVEL | 日志级别 | info |
| CORS_ORIGINS | CORS允许源 | * |

## 初始化数据

### 创建管理员账户

```bash
curl -X POST http://localhost:8080/api/v1/auth/init \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "Admin@123456",
    "email": "admin@example.com"
  }'
```

### 导入初始数据

```bash
# 导入电站数据
psql -U nem -d nem_system -f scripts/init/stations.sql

# 导入设备模板
psql -U nem -d nem_system -f scripts/init/device_templates.sql
```

## 验证安装

### 1. 检查服务状态

```bash
# 后端健康检查
curl http://localhost:8080/health

# 预期响应
{
  "status": "healthy",
  "version": "1.0.0",
  "timestamp": "2026-04-07T12:00:00Z"
}
```

### 2. 访问Web界面

打开浏览器访问 `http://localhost`，使用管理员账户登录。

### 3. 检查监控

```bash
# Prometheus指标
curl http://localhost:8080/metrics
```

## 常见问题

### Q: 数据库连接失败？

检查以下项：
1. PostgreSQL服务是否启动
2. 连接参数是否正确
3. 防火墙是否开放端口
4. 用户权限是否正确

### Q: 前端无法访问后端API？

检查以下项：
1. 后端服务是否启动
2. CORS配置是否正确
3. 网络是否通畅

### Q: Docker容器启动失败？

检查以下项：
1. Docker服务是否运行
2. 端口是否被占用
3. 资源是否充足

## 下一步

- [快速开始](./Quick-Start) - 了解系统基本使用
- [配置说明](./Configuration) - 详细配置参数
- [部署指南](./Deployment-Guide) - 生产环境部署

---

**最后更新**: 2026-04-07
