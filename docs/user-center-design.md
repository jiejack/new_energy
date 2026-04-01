# 用户中心模块设计文档

## 1. 模块概述

### 1.1 功能定位

用户中心是新能源在线监控系统的核心安全模块，负责用户身份认证、权限管理、个人信息维护等功能，确保系统操作的安全性与可追溯性。

### 1.2 核心功能

- 用户注册与登录
- 身份认证（JWT）
- 角色与权限管理（RBAC）
- 用户信息维护
- 操作审计日志
- 多因素认证（可选）

---

## 2. 架构设计

### 2.1 模块架构

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              用户中心模块                                     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  认证服务   │  │  授权服务   │  │  用户服务   │  │  审计服务   │        │
│  │    Auth     │  │ Permission  │  │    User     │  │    Audit    │        │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘        │
│         │                │                │                │               │
│         ▼                ▼                ▼                ▼               │
│  ┌─────────────────────────────────────────────────────────────────────┐  │
│  │                         数据访问层                                    │  │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐                   │  │
│  │  │UserRepo │ │RoleRepo │ │PermRepo │ │LogRepo  │                   │  │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘                   │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 2.2 认证流程

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   用户登录   │ ──► │  身份验证   │ ──► │  生成Token  │ ──► │  返回Token  │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │  验证失败   │
                    │  记录日志   │
                    └─────────────┘
```

---

## 3. 数据模型

### 3.1 用户表 (users)

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | VARCHAR(36) | PRIMARY KEY | UUID主键 |
| username | VARCHAR(100) | UNIQUE, NOT NULL | 用户名 |
| password_hash | VARCHAR(256) | NOT NULL | 密码哈希 |
| email | VARCHAR(200) | UNIQUE | 邮箱 |
| phone | VARCHAR(50) | | 手机号 |
| real_name | VARCHAR(100) | | 真实姓名 |
| avatar | VARCHAR(500) | | 头像URL |
| status | INTEGER | DEFAULT 1 | 状态：1-正常，0-禁用 |
| last_login | TIMESTAMP | | 最后登录时间 |
| login_count | INTEGER | DEFAULT 0 | 登录次数 |
| created_at | TIMESTAMP | DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT NOW() | 更新时间 |

### 3.2 角色表 (roles)

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | VARCHAR(36) | PRIMARY KEY | UUID主键 |
| code | VARCHAR(50) | UNIQUE, NOT NULL | 角色编码 |
| name | VARCHAR(100) | NOT NULL | 角色名称 |
| description | TEXT | | 描述 |
| is_system | BOOLEAN | DEFAULT FALSE | 是否系统角色 |
| created_at | TIMESTAMP | DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT NOW() | 更新时间 |

### 3.3 权限表 (permissions)

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | VARCHAR(36) | PRIMARY KEY | UUID主键 |
| code | VARCHAR(100) | UNIQUE, NOT NULL | 权限编码 |
| name | VARCHAR(100) | NOT NULL | 权限名称 |
| resource_type | VARCHAR(50) | | 资源类型 |
| resource_id | VARCHAR(36) | | 资源ID |
| action | VARCHAR(50) | | 操作类型 |
| description | TEXT | | 描述 |

### 3.4 用户角色关联表 (user_roles)

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| user_id | VARCHAR(36) | FOREIGN KEY | 用户ID |
| role_id | VARCHAR(36) | FOREIGN KEY | 角色ID |
| created_at | TIMESTAMP | DEFAULT NOW() | 创建时间 |

### 3.5 角色权限关联表 (role_permissions)

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| role_id | VARCHAR(36) | FOREIGN KEY | 角色ID |
| permission_id | VARCHAR(36) | FOREIGN KEY | 权限ID |
| created_at | TIMESTAMP | DEFAULT NOW() | 创建时间 |

### 3.6 操作日志表 (operation_logs)

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | VARCHAR(36) | PRIMARY KEY | UUID主键 |
| user_id | VARCHAR(36) | FOREIGN KEY | 用户ID |
| username | VARCHAR(100) | | 用户名 |
| action | VARCHAR(100) | NOT NULL | 操作类型 |
| resource_type | VARCHAR(50) | | 资源类型 |
| resource_id | VARCHAR(36) | | 资源ID |
| details | JSONB | | 操作详情 |
| ip_address | VARCHAR(50) | | IP地址 |
| user_agent | VARCHAR(500) | | 用户代理 |
| created_at | TIMESTAMP | DEFAULT NOW() | 创建时间 |

---

## 4. API接口设计

### 4.1 认证接口

#### 4.1.1 用户登录

**POST** `/api/v1/auth/login`

请求体：
```json
{
  "username": "admin",
  "password": "password123"
}
```

响应：
```json
{
  "code": 0,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 7200,
    "token_type": "Bearer"
  }
}
```

#### 4.1.2 刷新Token

**POST** `/api/v1/auth/refresh`

#### 4.1.3 用户登出

**POST** `/api/v1/auth/logout`

### 4.2 用户管理接口

#### 4.2.1 获取用户列表

**GET** `/api/v1/users`

#### 4.2.2 创建用户

**POST** `/api/v1/users`

#### 4.2.3 更新用户

**PUT** `/api/v1/users/{id}`

#### 4.2.4 删除用户

**DELETE** `/api/v1/users/{id}`

#### 4.2.5 修改密码

**PUT** `/api/v1/users/{id}/password`

### 4.3 角色管理接口

#### 4.3.1 获取角色列表

**GET** `/api/v1/roles`

#### 4.3.2 创建角色

**POST** `/api/v1/roles`

#### 4.3.3 分配用户角色

**POST** `/api/v1/users/{id}/roles`

### 4.4 权限管理接口

#### 4.4.1 获取权限列表

**GET** `/api/v1/permissions`

#### 4.4.2 分配角色权限

**POST** `/api/v1/roles/{id}/permissions`

---

## 5. 权限模型设计

### 5.1 RBAC模型

```
用户(User) ──┬──► 角色(Role) ──┬──► 权限(Permission)
             │                  │
             │                  └──► 资源(Resource) + 操作(Action)
             │
             └──► 多对多关系
```

### 5.2 预置角色

| 角色编码 | 角色名称 | 说明 |
|----------|----------|------|
| super_admin | 超级管理员 | 系统最高权限 |
| admin | 系统管理员 | 系统管理权限 |
| operator | 运维人员 | 设备操作权限 |
| viewer | 查看人员 | 只读权限 |

### 5.3 权限编码规范

```
格式：{resource}:{action}

示例：
- station:create   创建厂站
- station:read     查看厂站
- station:update   更新厂站
- station:delete   删除厂站
- device:control   设备控制
- alarm:ack        告警确认
```

---

## 6. 安全设计

### 6.1 密码安全

- 使用bcrypt进行密码哈希
- 密码强度要求：至少8位，包含大小写字母和数字
- 密码修改需验证旧密码

### 6.2 Token安全

- JWT Token有效期：2小时
- Refresh Token有效期：7天
- Token存储于Redis，支持主动失效

### 6.3 登录安全

- 登录失败次数限制：5次
- 失败锁定时间：30分钟
- 异常登录告警

### 6.4 操作审计

- 记录所有敏感操作
- 日志包含用户、时间、IP、操作内容
- 日志保留期限：1年

---

## 7. 部署说明

### 7.1 初始化数据

系统启动时自动创建：
- 超级管理员账号
- 预置角色
- 基础权限

### 7.2 配置项

```yaml
auth:
  jwt:
    secret: your-jwt-secret-key
    access_expire: 7200      # 2小时
    refresh_expire: 604800   # 7天
  password:
    min_length: 8
    require_uppercase: true
    require_lowercase: true
    require_digit: true
  login:
    max_attempts: 5
    lock_duration: 1800      # 30分钟
```
