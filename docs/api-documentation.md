# 新能源监控系统 API 技术文档

## 概述

本文档描述了新能源监控系统的RESTful API接口规范。所有API均遵循统一的请求/响应格式。

### 基础信息

- **基础URL**: `http://localhost:8080/api/v1`
- **认证方式**: JWT Bearer Token
- **内容类型**: `application/json`
- **字符编码**: UTF-8

### 通用响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {},
  "timestamp": 1712345678
}
```

### 分页响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": [],
  "timestamp": 1712345678,
  "page": 1,
  "pageSize": 20,
  "total": 100
}
```

### 错误码说明

| 错误码 | 说明 |
|-------|------|
| 0 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 认证接口

### 用户登录

**POST** `/auth/login`

请求体:
```json
{
  "username": "admin",
  "password": "password123"
}
```

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 7200,
    "user": {
      "id": "1",
      "username": "admin",
      "nickname": "系统管理员",
      "role": "admin"
    }
  }
}
```

### 用户登出

**POST** `/auth/logout`

Headers: `Authorization: Bearer {token}`

---

## 区域管理接口

### 获取区域列表

**GET** `/regions`

查询参数:
- `parent_id` (可选): 父区域ID

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "region_001",
      "name": "华北区域",
      "code": "HB",
      "parent_id": null,
      "level": 1,
      "status": 1,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### 创建区域

**POST** `/regions`

请求体:
```json
{
  "name": "华北区域",
  "code": "HB",
  "parent_id": null,
  "description": "华北地区电站"
}
```

### 获取区域详情

**GET** `/regions/:id`

### 更新区域

**PUT** `/regions/:id`

### 删除区域

**DELETE** `/regions/:id`

---

## 厂站管理接口

### 获取厂站列表

**GET** `/stations`

查询参数:
- `sub_region_id` (可选): 子区域ID
- `type` (可选): 厂站类型 (solar/wind/storage)
- `status` (可选): 状态
- `page` (可选): 页码，默认1
- `page_size` (可选): 每页数量，默认20

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "station_001",
      "code": "BJ-CY-001",
      "name": "北京朝阳光伏电站",
      "type": "solar",
      "sub_region_id": "region_001",
      "capacity": 5000,
      "voltage_level": "35kV",
      "longitude": 116.4074,
      "latitude": 39.9042,
      "address": "北京市朝阳区XXX路XXX号",
      "status": 1,
      "commission_date": "2024-01-01T00:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "page": 1,
  "pageSize": 20,
  "total": 6
}
```

### 创建厂站

**POST** `/stations`

请求体:
```json
{
  "code": "BJ-CY-001",
  "name": "北京朝阳光伏电站",
  "type": "solar",
  "sub_region_id": "region_001",
  "capacity": 5000,
  "voltage_level": "35kV",
  "longitude": 116.4074,
  "latitude": 39.9042,
  "address": "北京市朝阳区XXX路XXX号"
}
```

### 获取厂站详情

**GET** `/stations/:id`

### 更新厂站

**PUT** `/stations/:id`

### 删除厂站

**DELETE** `/stations/:id`

### 获取厂站统计

**GET** `/stations/:id/statistics`

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "device_count": 25,
    "online_device_count": 24,
    "offline_device_count": 1,
    "alarm_count": 5,
    "power": 4500,
    "energy": 28500
  }
}
```

---

## 设备管理接口

### 获取设备列表

**GET** `/devices`

查询参数:
- `station_id` (可选): 厂站ID
- `type` (可选): 设备类型
- `status` (可选): 状态
- `page` (可选): 页码
- `page_size` (可选): 每页数量

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "device_001",
      "code": "INV-BJ-001",
      "name": "逆变器 #01",
      "type": "inverter",
      "station_id": "station_001",
      "manufacturer": "华为",
      "model": "SUN2000-100KTL",
      "serial_number": "SN2101012345",
      "rated_power": 100,
      "rated_voltage": 380,
      "rated_current": 150,
      "protocol": "modbus",
      "ip_address": "192.168.1.101",
      "port": 502,
      "slave_id": 1,
      "status": 1,
      "last_online": "2024-01-01T00:00:00Z",
      "install_date": "2024-01-01T00:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "page": 1,
  "pageSize": 20,
  "total": 8
}
```

### 创建设备

**POST** `/devices`

### 获取设备详情

**GET** `/devices/:id`

### 更新设备

**PUT** `/devices/:id`

### 删除设备

**DELETE** `/devices/:id`

---

## 采集点管理接口

### 获取采集点列表

**GET** `/points`

查询参数:
- `device_id` (可选): 设备ID
- `type` (可选): 点类型
- `status` (可选): 状态

### 创建采集点

**POST** `/points`

### 获取采集点详情

**GET** `/points/:id`

### 更新采集点

**PUT** `/points/:id`

### 删除采集点

**DELETE** `/points/:id`

---

## 告警管理接口

### 获取告警列表

**GET** `/alarms`

查询参数:
- `station_id` (可选): 厂站ID
- `device_id` (可选): 设备ID
- `level` (可选): 告警级别 (1-4)
- `status` (可选): 状态
- `type` (可选): 告警类型
- `start_time` (可选): 开始时间戳
- `end_time` (可选): 结束时间戳
- `page` (可选): 页码
- `page_size` (可选): 每页数量

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "alarm_001",
      "point_id": "point_001",
      "device_id": "device_001",
      "station_id": "station_001",
      "type": "limit",
      "level": 3,
      "title": "逆变器温度过高",
      "message": "逆变器#01温度为87.5°C，已超过85°C阈值",
      "value": 87.5,
      "threshold": 85.0,
      "status": 1,
      "triggered_at": "2024-01-01T00:00:00Z"
    }
  ],
  "page": 1,
  "pageSize": 20,
  "total": 8
}
```

### 获取告警详情

**GET** `/alarms/:id`

### 确认告警

**PUT** `/alarms/:id/ack`

请求体:
```json
{
  "ack_note": "已确认，正在处理"
}
```

### 清除告警

**PUT** `/alarms/:id/clear`

请求体:
```json
{
  "clear_note": "问题已解决"
}
```

### 获取告警统计

**GET** `/alarms/statistics`

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 156,
    "active": 23,
    "acknowledged": 45,
    "cleared": 88,
    "by_level": {
      "1": 45,
      "2": 78,
      "3": 33
    },
    "by_type": {
      "limit": 89,
      "communication": 42,
      "quality": 25
    }
  }
}
```

---

## 告警规则管理接口

### 获取告警规则列表

**GET** `/alarm-rules`

查询参数:
- `type` (可选): 规则类型
- `level` (可选): 告警级别
- `status` (可选): 状态
- `page` (可选): 页码
- `page_size` (可选): 每页数量

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "rule_001",
      "name": "温度超限告警",
      "description": "设备温度超过设定阈值时触发",
      "type": "limit",
      "level": 3,
      "condition": "temperature > threshold",
      "threshold": 80,
      "duration": 60,
      "point_id": null,
      "device_id": null,
      "station_id": null,
      "notify_channels": ["email", "sms"],
      "notify_users": ["admin", "operator"],
      "status": 1,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "page": 1,
  "pageSize": 10,
  "total": 3
}
```

### 创建告警规则

**POST** `/alarm-rules`

请求体:
```json
{
  "name": "温度超限告警",
  "description": "设备温度超过设定阈值时触发",
  "type": "limit",
  "level": 3,
  "condition": "temperature > threshold",
  "threshold": 80,
  "duration": 60,
  "notify_channels": ["email", "sms"],
  "notify_users": ["admin", "operator"]
}
```

### 获取告警规则详情

**GET** `/alarm-rules/:id`

### 更新告警规则

**PUT** `/alarm-rules/:id`

### 删除告警规则

**DELETE** `/alarm-rules/:id`

---

## 通知配置管理接口

### 获取通知配置列表

**GET** `/notification-configs`

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "notif_001",
      "type": "email",
      "name": "邮件通知",
      "config": {
        "smtp_host": "smtp.example.com",
        "smtp_port": 465,
        "username": "alert@example.com",
        "from": "alert@example.com",
        "use_tls": true
      },
      "enabled": false
    },
    {
      "id": "notif_002",
      "type": "sms",
      "name": "短信通知",
      "config": {
        "access_key": "",
        "secret_key": "",
        "sign_name": "新能源监控"
      },
      "enabled": false
    },
    {
      "id": "notif_003",
      "type": "webhook",
      "name": "Webhook通知",
      "config": {
        "url": "",
        "method": "POST"
      },
      "enabled": false
    },
    {
      "id": "notif_004",
      "type": "wechat",
      "name": "微信通知",
      "config": {
        "corp_id": "",
        "agent_id": ""
      },
      "enabled": false
    }
  ]
}
```

### 获取通知配置详情

**GET** `/notification-configs/:type`

### 更新通知配置

**PUT** `/notification-configs/:type`

请求体:
```json
{
  "smtp_host": "smtp.example.com",
  "smtp_port": 465,
  "username": "alert@example.com",
  "password": "password",
  "from": "alert@example.com",
  "use_tls": true
}
```

### 启用通知配置

**POST** `/notification-configs/:type/enable`

### 禁用通知配置

**POST** `/notification-configs/:type/disable`

### 测试通知配置

**POST** `/notification-configs/:type/test`

请求体:
```json
{
  "test_target": "test@example.com"
}
```

---

## 报表管理接口

### 生成报表

**GET** `/reports`

查询参数:
- `type` (可选): 报表类型 (daily/weekly/monthly)
- `start_time` (可选): 开始时间
- `end_time` (可选): 结束时间
- `station_id` (可选): 电站ID

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "type": "daily",
    "start_time": "2024-03-01",
    "end_time": "2024-03-31",
    "stations": [
      {
        "station_id": "station_001",
        "station_name": "光伏电站A",
        "total_power": 125000,
        "yoy_change": 12.5,
        "mom_change": 5.2,
        "alarm_count": 15,
        "online_rate": 99.5
      }
    ],
    "summary": {
      "total_power": 214000,
      "total_alarms": 23,
      "avg_online_rate": 98.85
    }
  }
}
```

### 导出报表

**GET** `/reports/export`

查询参数:
- `type` (可选): 报表类型
- `format` (可选): 导出格式 (excel/csv)

---

## 操作日志接口

### 获取操作日志列表

**GET** `/operation-logs`

查询参数:
- `user_id` (可选): 用户ID
- `action` (可选): 操作类型
- `start_time` (可选): 开始时间
- `end_time` (可选): 结束时间
- `page` (可选): 页码
- `page_size` (可选): 每页数量

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "log_001",
      "user_id": "user_001",
      "username": "admin",
      "method": "POST",
      "path": "/api/v1/stations",
      "action": "create",
      "resource": "station",
      "resource_id": "station_new",
      "request_ip": "192.168.1.100",
      "status": 200,
      "duration": 45,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "page": 1,
  "pageSize": 20,
  "total": 3
}
```

---

## 数据查询接口

### 获取实时数据

**GET** `/data/realtime`

查询参数:
- `point_ids`: 采集点ID列表，逗号分隔

### 获取历史数据

**GET** `/data/history`

查询参数:
- `point_id`: 采集点ID
- `start_time`: 开始时间戳
- `end_time`: 结束时间戳
- `interval` (可选): 采样间隔(秒)

### 获取统计数据

**GET** `/data/statistics`

查询参数:
- `station_id` (可选): 厂站ID
- `type` (可选): 统计类型
- `start_time` (可选): 开始时间戳
- `end_time` (可选): 结束时间戳

---

## 控制操作接口

### 遥控操作

**POST** `/control/operate`

请求体:
```json
{
  "point_id": "point_001",
  "value": 1,
  "operator": "admin"
}
```

### 参数设置

**POST** `/control/setpoint`

请求体:
```json
{
  "point_id": "point_001",
  "value": 100,
  "operator": "admin"
}
```

---

## AI服务接口

### AI问答

**POST** `/ai/qa`

请求体:
```json
{
  "question": "如何优化光伏电站的发电效率？",
  "context": {}
}
```

### AI配置建议

**POST** `/ai/config/suggest`

请求体:
```json
{
  "station_id": "station_001",
  "config_type": "alarm"
}
```

---

## 用户管理接口

### 获取用户列表

**GET** `/users`

### 创建用户

**POST** `/users`

请求体:
```json
{
  "username": "operator",
  "password": "password123",
  "nickname": "操作员",
  "email": "operator@example.com",
  "phone": "13800138001",
  "role": "operator"
}
```

### 获取用户详情

**GET** `/users/:id`

### 更新用户

**PUT** `/users/:id`

### 删除用户

**DELETE** `/users/:id`

### 修改密码

**PUT** `/users/:id/password`

请求体:
```json
{
  "old_password": "oldpass123",
  "new_password": "newpass456"
}
```

---

## 个人中心接口

### 获取个人信息

**GET** `/profile`

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "1",
    "username": "admin",
    "nickname": "系统管理员",
    "email": "admin@example.com",
    "phone": "13800138000",
    "avatar": "https://api.dicebear.com/7.x/avataaars/svg?seed=admin",
    "role": "admin",
    "status": 1,
    "create_time": "2024-01-01 00:00:00"
  }
}
```

### 更新个人信息

**PUT** `/profile`

### 获取偏好设置

**GET** `/profile/preferences`

### 更新偏好设置

**PUT** `/profile/preferences`

请求体:
```json
{
  "theme": "dark",
  "language": "zh-CN",
  "timezone": "Asia/Shanghai",
  "notify_enabled": true,
  "notify_types": ["alarm", "system"]
}
```

### 上传头像

**POST** `/profile/avatar`

---

## WebSocket 接口

### 连接地址

`ws://localhost:8080/ws`

### 消息格式

订阅实时功率:
```json
{
  "type": "subscribe-power"
}
```

订阅告警:
```json
{
  "type": "subscribe-alarm"
}
```

实时功率数据:
```json
{
  "type": "realtime-power",
  "payload": {
    "station1": 1250.5,
    "station2": 890.3,
    "station3": 2100.8
  },
  "timestamp": 1712345678
}
```

告警数据:
```json
{
  "type": "alarm",
  "payload": {
    "id": 1712345678000,
    "level": "warning",
    "title": "测试告警",
    "message": "这是一条测试告警信息",
    "time": "2024-04-05 12:00:00"
  }
}
```

---

## Harness 验证接口

### 概述

Harness 是一个统一的验证框架，提供输入验证、输出验证、约束检查、监控和快照功能。所有 Harness API 均为内部服务接口，不直接暴露给外部客户端。

### 基础验证

**POST** `/harness/validate`

请求体:
```json
{
  "input": {
    "type": "alarm",
    "data": {
      "level": 3,
      "title": "温度告警",
      "value": 85.5,
      "threshold": 80.0
    }
  }
}
```

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "valid": true,
    "errors": []
  }
}
```

### 异步验证

**POST** `/harness/validate/async`

请求体:
```json
{
  "input": {
    "type": "device",
    "data": {
      "code": "INV-001",
      "name": "逆变器01",
      "type": "inverter"
    }
  }
}
```

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "task_id": "task_12345",
    "status": "processing"
  }
}
```

### 输出验证

**POST** `/harness/verify`

请求体:
```json
{
  "expected": {
    "id": "alarm_001",
    "level": 3,
    "title": "温度告警"
  },
  "actual": {
    "id": "alarm_001",
    "level": 3,
    "title": "温度告警"
  }
}
```

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "match": true,
    "differences": []
  }
}
```

### 快照管理

#### 创建快照

**POST** `/harness/snapshot`

请求体:
```json
{
  "target": {
    "type": "alarm",
    "id": "alarm_001"
  }
}
```

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "snapshot_id": "snap_12345",
    "created_at": "2024-04-07T12:00:00Z"
  }
}
```

#### 加载快照

**GET** `/harness/snapshot/:id`

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "snap_12345",
    "data": "...",
    "created_at": "2024-04-07T12:00:00Z"
  }
}
```

#### 比较快照

**POST** `/harness/snapshot/:id/compare`

请求体:
```json
{
  "data": "..."
}
```

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "match": true,
    "differences": []
  }
}
```

### 约束检查

**POST** `/harness/constraint/check`

请求体:
```json
{
  "target": {
    "type": "device",
    "data": {
      "rated_power": 100,
      "rated_voltage": 380
    }
  }
}
```

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "satisfied": true,
    "violations": []
  }
}
```

### 指标记录

**POST** `/harness/metrics/record`

请求体:
```json
{
  "metric": "device.validation.count",
  "value": 1.0,
  "tags": {
    "device_type": "inverter",
    "station_id": "station_001"
  }
}
```

### 获取指标

**GET** `/harness/metrics`

查询参数:
- `pattern` (可选): 指标名称模式

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "name": "device.validation.count",
      "value": 156.0,
      "timestamp": 1712345678
    }
  ]
}
```

### 完整测试流程

**POST** `/harness/execute`

请求体:
```json
{
  "input": {
    "type": "alarm",
    "data": {
      "level": 3,
      "title": "温度告警"
    }
  },
  "expected_output": {
    "status": 1,
    "level": 3
  },
  "actual_output": {
    "status": 1,
    "level": 3
  }
}
```

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "input_valid": true,
    "output_match": true,
    "errors": []
  }
}
```

---

## 告警模块 Harness 集成接口

### 概述

告警模块集成了 Harness 验证框架，在创建、确认、清除告警时自动执行验证，确保数据质量和业务规则合规性。

### 验证创建告警

**POST** `/alarms/harness/validate-create`

请求体:
```json
{
  "point_id": "point_001",
  "device_id": "device_001",
  "station_id": "station_001",
  "type": "limit",
  "level": 3,
  "title": "逆变器温度过高",
  "message": "逆变器#01温度为87.5°C，已超过85°C阈值",
  "value": 87.5,
  "threshold": 85.0
}
```

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "valid": true,
    "errors": []
  }
}
```

### 验证确认告警

**POST** `/alarms/:id/harness/validate-ack`

请求体:
```json
{
  "operator": "admin"
}
```

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "valid": true,
    "errors": []
  }
}
```

### 验证清除告警

**POST** `/alarms/:id/harness/validate-clear`

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "valid": true,
    "errors": []
  }
}
```

### 验证告警查询

**POST** `/alarms/harness/validate-query`

请求体:
```json
{
  "station_id": "station_001",
  "start_time": 1712345678,
  "end_time": 1712432078
}
```

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "valid": true,
    "errors": []
  }
}
```

### 验证告警输出

**POST** `/alarms/:id/harness/verify-output`

请求体:
```json
{
  "expected": {
    "id": "alarm_001",
    "level": 3,
    "title": "温度告警",
    "status": 1
  }
}
```

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "match": true,
    "differences": []
  }
}
```

### 创建告警快照

**POST** `/alarms/:id/harness/snapshot`

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "snapshot_id": "alarm_snap_12345",
    "created_at": "2024-04-07T12:00:00Z"
  }
}
```

### 验证告警状态

**GET** `/alarms/:id/harness/verify-state`

查询参数:
- `expected_status`: 预期状态 (1=活动, 2=已确认, 3=已清除)

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "match": true,
    "actual_status": 1,
    "expected_status": 1
  }
}
```

---

## 设备模块 Harness 集成接口

### 概述

设备模块集成了 Harness 验证框架，在创建、更新设备时自动执行验证，确保设备参数的有效性和一致性。

### 验证创建设备

**POST** `/devices/harness/validate-create`

请求体:
```json
{
  "code": "INV-BJ-001",
  "name": "逆变器 #01",
  "type": "inverter",
  "station_id": "station_001",
  "manufacturer": "华为",
  "model": "SUN2000-100KTL",
  "serial_number": "SN2101012345",
  "rated_power": 100,
  "rated_voltage": 380,
  "rated_current": 150,
  "protocol": "modbus-tcp",
  "ip_address": "192.168.1.101",
  "port": 502,
  "slave_id": 1
}
```

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "valid": true,
    "errors": []
  }
}
```

### 验证更新设备

**POST** `/devices/:id/harness/validate-update`

请求体:
```json
{
  "rated_power": 120,
  "rated_voltage": 380,
  "rated_current": 180
}
```

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "valid": true,
    "errors": []
  }
}
```

### 验证设备状态

**POST** `/devices/:id/harness/validate-status`

请求体:
```json
{
  "status": 1
}
```

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "valid": true,
    "errors": []
  }
}
```

### 验证设备查询

**POST** `/devices/harness/validate-query`

请求体:
```json
{
  "station_id": "station_001",
  "type": "inverter"
}
```

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "valid": true,
    "errors": []
  }
}
```

### 验证通信参数

**POST** `/devices/harness/validate-comm`

请求体:
```json
{
  "protocol": "modbus-tcp",
  "ip_address": "192.168.1.101",
  "port": 502,
  "slave_id": 1
}
```

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "valid": true,
    "errors": []
  }
}
```

### 验证设备输出

**POST** `/devices/:id/harness/verify-output`

请求体:
```json
{
  "expected": {
    "id": "device_001",
    "code": "INV-BJ-001",
    "name": "逆变器 #01",
    "status": 1
  }
}
```

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "match": true,
    "differences": []
  }
}
```

### 创建设备快照

**POST** `/devices/:id/harness/snapshot`

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "snapshot_id": "device_snap_12345",
    "created_at": "2024-04-07T12:00:00Z"
  }
}
```

### 记录设备指标

**POST** `/devices/:id/harness/metrics/record`

请求体:
```json
{
  "metric": "device.validation.count",
  "value": 1.0
}
```

### 获取设备指标

**GET** `/devices/:id/harness/metrics`

查询参数:
- `pattern` (可选): 指标名称模式

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "name": "device.validation.count",
      "value": 45.0,
      "timestamp": 1712345678
    },
    {
      "name": "device.comm.success_rate",
      "value": 99.5,
      "timestamp": 1712345678
    }
  ]
}
```

---

## Swagger 文档

访问 Swagger UI: `http://localhost:8080/swagger/index.html`

---

## 版本历史

| 版本 | 日期 | 说明 |
|-----|------|------|
| 1.1.0 | 2026-04-07 | 添加 Harness 验证框架集成 API |
| 1.0.0 | 2024-04-05 | 初始版本 |
