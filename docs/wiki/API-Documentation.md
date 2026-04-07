# API 文档

本文档详细介绍新能源监控系统的RESTful API接口。

## 基本信息

| 项目 | 说明 |
|------|------|
| 基础URL | `http://localhost:8080/api/v1` |
| 认证方式 | JWT Bearer Token |
| 数据格式 | JSON |
| 编码 | UTF-8 |

## 认证

### 登录获取Token

**请求**
```
POST /auth/login
```

**参数**
```json
{
  "username": "string",  // 用户名
  "password": "string"   // 密码
}
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expire": "2026-04-08T12:00:00Z",
    "user": {
      "id": "1",
      "username": "admin",
      "email": "admin@example.com",
      "roles": ["admin"]
    }
  }
}
```

### 使用Token

在请求头中添加：
```
Authorization: Bearer <token>
```

### 刷新Token

**请求**
```
POST /auth/refresh
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expire": "2026-04-08T14:00:00Z"
  }
}
```

---

## 电站管理

### 获取电站列表

**请求**
```
GET /stations
```

**参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认10 |
| name | string | 否 | 电站名称（模糊搜索） |
| type | string | 否 | 电站类型 |

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": "uuid",
        "name": "光伏电站A",
        "type": "solar",
        "capacity": 1000,
        "address": "北京市朝阳区",
        "longitude": 116.4,
        "latitude": 39.9,
        "status": "online",
        "created_at": "2026-04-01T00:00:00Z"
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 10
  }
}
```

### 获取电站详情

**请求**
```
GET /stations/{id}
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "uuid",
    "name": "光伏电站A",
    "type": "solar",
    "capacity": 1000,
    "address": "北京市朝阳区",
    "longitude": 116.4,
    "latitude": 39.9,
    "status": "online",
    "device_count": 50,
    "alarm_count": 2,
    "realtime_power": 850.5,
    "daily_energy": 5000.2,
    "created_at": "2026-04-01T00:00:00Z",
    "updated_at": "2026-04-07T12:00:00Z"
  }
}
```

### 创建电站

**请求**
```
POST /stations
```

**参数**
```json
{
  "name": "光伏电站A",
  "type": "solar",
  "capacity": 1000,
  "address": "北京市朝阳区",
  "longitude": 116.4,
  "latitude": 39.9
}
```

### 更新电站

**请求**
```
PUT /stations/{id}
```

### 删除电站

**请求**
```
DELETE /stations/{id}
```

---

## 设备管理

### 获取设备列表

**请求**
```
GET /devices
```

**参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| station_id | string | 否 | 电站ID |
| type | string | 否 | 设备类型 |
| status | string | 否 | 设备状态 |

### 获取设备详情

**请求**
```
GET /devices/{id}
```

### 创建设备

**请求**
```
POST /devices
```

**参数**
```json
{
  "name": "逆变器#1",
  "type": "inverter",
  "station_id": "uuid",
  "protocol": "modbus_tcp",
  "address": "192.168.1.100",
  "port": 502,
  "slave_id": 1
}
```

---

## 告警管理

### 获取告警列表

**请求**
```
GET /alarms
```

**参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| level | int | 否 | 告警级别(1-4) |
| status | string | 否 | 状态(active/acknowledged/cleared) |
| station_id | string | 否 | 电站ID |
| device_id | string | 否 | 设备ID |
| start_time | string | 否 | 开始时间 |
| end_time | string | 否 | 结束时间 |

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": "uuid",
        "level": 3,
        "level_name": "严重",
        "message": "逆变器温度超过80°C",
        "device_id": "uuid",
        "device_name": "逆变器#1",
        "station_id": "uuid",
        "station_name": "光伏电站A",
        "status": "active",
        "triggered_at": "2026-04-07T10:30:00Z",
        "acknowledged_at": null,
        "cleared_at": null
      }
    ],
    "total": 50
  }
}
```

### 确认告警

**请求**
```
PUT /alarms/{id}/acknowledge
```

**参数**
```json
{
  "note": "已派人现场检查"
}
```

### 清除告警

**请求**
```
PUT /alarms/{id}/clear
```

### 告警规则管理

**获取规则列表**
```
GET /alarm-rules
```

**创建规则**
```
POST /alarm-rules
```

**参数**
```json
{
  "name": "温度超限告警",
  "type": "limit",
  "level": 3,
  "condition": "temperature > 80",
  "threshold": 80,
  "duration": 60,
  "notify_channels": ["email", "sms"],
  "notify_users": ["user_id_1", "user_id_2"]
}
```

---

## 数据查询

### 查询历史数据

**请求**
```
GET /data/history
```

**参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| device_id | string | 是 | 设备ID |
| point_ids | string | 是 | 数据点ID列表(逗号分隔) |
| start_time | string | 是 | 开始时间 |
| end_time | string | 是 | 结束时间 |
| interval | string | 否 | 采样间隔(1m/5m/1h/1d) |

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "device_id": "uuid",
    "points": [
      {
        "point_id": "temperature",
        "point_name": "温度",
        "unit": "°C",
        "values": [
          {
            "time": "2026-04-07T10:00:00Z",
            "value": 75.5
          },
          {
            "time": "2026-04-07T10:01:00Z",
            "value": 76.2
          }
        ]
      }
    ]
  }
}
```

### 导出数据

**请求**
```
GET /data/export
```

**参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| format | string | 是 | 导出格式(excel/csv) |
| device_id | string | 是 | 设备ID |
| point_ids | string | 是 | 数据点ID列表 |
| start_time | string | 是 | 开始时间 |
| end_time | string | 是 | 结束时间 |

---

## 用户管理

### 获取用户列表

**请求**
```
GET /users
```

### 创建用户

**请求**
```
POST /users
```

**参数**
```json
{
  "username": "operator1",
  "password": "Password@123",
  "email": "operator1@example.com",
  "phone": "13800138000",
  "roles": ["operator"]
}
```

### 更新用户

**请求**
```
PUT /users/{id}
```

### 删除用户

**请求**
```
DELETE /users/{id}
```

---

## 系统配置

### 获取配置

**请求**
```
GET /configs
```

### 更新配置

**请求**
```
PUT /configs/{category}/{key}
```

**参数**
```json
{
  "value": "新配置值"
}
```

---

## 错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权/Token过期 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 409 | 资源冲突 |
| 500 | 服务器内部错误 |

**错误响应格式**
```json
{
  "code": 400,
  "message": "参数验证失败",
  "data": {
    "errors": [
      {
        "field": "name",
        "message": "名称不能为空"
      }
    ]
  }
}
```

---

## 速率限制

| 接口类型 | 限制 |
|----------|------|
| 认证接口 | 10次/分钟 |
| 查询接口 | 100次/分钟 |
| 写入接口 | 50次/分钟 |

超过限制返回 `429 Too Many Requests`

---

**最后更新**: 2026-04-07
