# API接口文档

## 文档信息

| 项目 | 内容 |
|------|------|
| 项目名称 | 新能源在线监控系统 |
| 文档版本 | v1.0.0 |
| 编写日期 | 2024-03-01 |
| Base URL | `http://localhost:8080/api/v1` |

---

## 1. 概述

### 1.1 基础信息

- **协议**: HTTP/HTTPS
- **数据格式**: JSON
- **字符编码**: UTF-8
- **认证方式**: JWT Bearer Token
- **API版本**: v1

### 1.2 通用响应格式

#### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": {},
  "timestamp": 1709500800000
}
```

#### 错误响应

```json
{
  "code": 400,
  "message": "请求参数错误",
  "data": null,
  "timestamp": 1709500800000
}
```

### 1.3 错误码定义

| 错误码 | HTTP状态码 | 说明 |
|--------|------------|------|
| 0 | 200 | 成功 |
| 400 | 400 | 请求参数错误 |
| 401 | 401 | 未授权，Token无效或已过期 |
| 403 | 403 | 无权限访问该资源 |
| 404 | 404 | 资源不存在 |
| 409 | 409 | 资源冲突 |
| 422 | 422 | 参数验证失败 |
| 429 | 429 | 请求频率超限 |
| 500 | 500 | 服务器内部错误 |
| 503 | 503 | 服务暂时不可用 |

### 1.4 分页参数

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码，从1开始 |
| page_size | int | 否 | 20 | 每页数量，最大100 |

#### 分页响应格式

```json
{
  "code": 0,
  "data": {
    "total": 100,
    "page": 1,
    "page_size": 20,
    "items": []
  }
}
```

---

## 2. 认证授权

### 2.1 认证方式

系统采用JWT (JSON Web Token) 进行身份认证。

#### 请求头设置

```
Authorization: Bearer <access_token>
```

### 2.2 获取Token

#### 用户登录

**POST** `/auth/login`

**请求体**:

```json
{
  "username": "admin",
  "password": "Admin@123"
}
```

**响应示例**:

```json
{
  "code": 0,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 7200
  }
}
```

### 2.3 刷新Token

**POST** `/auth/refresh`

**请求体**:

```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### 2.4 退出登录

**POST** `/auth/logout`

**请求头**: 需要携带有效的Access Token

---

## 3. 区域管理API

### 3.1 获取区域列表

**GET** `/regions`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| parent_id | string | 否 | 父区域ID，获取子区域列表 |

**响应示例**:

```json
{
  "code": 0,
  "data": [
    {
      "id": "region-001",
      "code": "EAST",
      "name": "华东区域",
      "level": 1,
      "sort_order": 1,
      "parent_id": null,
      "description": "华东区域",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    },
    {
      "id": "region-002",
      "code": "NORTH",
      "name": "华北区域",
      "level": 1,
      "sort_order": 2,
      "parent_id": null,
      "description": "华北区域",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### 3.2 获取区域详情

**GET** `/regions/{id}`

**路径参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 区域ID |

**响应示例**:

```json
{
  "code": 0,
  "data": {
    "id": "region-001",
    "code": "EAST",
    "name": "华东区域",
    "level": 1,
    "sort_order": 1,
    "parent_id": null,
    "description": "华东区域",
    "children": [
      {
        "id": "sub-region-001",
        "code": "EAST_SH",
        "name": "上海",
        "level": 2
      }
    ],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### 3.3 创建区域

**POST** `/regions`

**请求体**:

```json
{
  "code": "EAST_SH",
  "name": "上海子区域",
  "parent_id": "region-001",
  "level": 2,
  "sort_order": 1,
  "description": "上海市区域"
}
```

**响应示例**:

```json
{
  "code": 0,
  "data": {
    "id": "region-003",
    "code": "EAST_SH",
    "name": "上海子区域",
    "level": 2,
    "sort_order": 1,
    "parent_id": "region-001"
  },
  "message": "创建成功"
}
```

### 3.4 更新区域

**PUT** `/regions/{id}`

**请求体**:

```json
{
  "name": "上海区域",
  "sort_order": 2,
  "description": "上海市区域（更新）"
}
```

### 3.5 删除区域

**DELETE** `/regions/{id}`

**响应示例**:

```json
{
  "code": 0,
  "message": "删除成功"
}
```

---

## 4. 厂站管理API

### 4.1 获取厂站列表

**GET** `/stations`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sub_region_id | string | 否 | 子区域ID |
| type | string | 否 | 厂站类型：pv/wind/ess/hybrid |
| status | int | 否 | 状态：0-停用，1-启用 |
| keyword | string | 否 | 搜索关键字 |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

**响应示例**:

```json
{
  "code": 0,
  "data": {
    "total": 100,
    "page": 1,
    "page_size": 20,
    "items": [
      {
        "id": "station-001",
        "code": "PV_SH_001",
        "name": "上海光伏电站1号",
        "type": "pv",
        "sub_region_id": "sub-region-001",
        "capacity": 50.0,
        "voltage_level": "35kV",
        "longitude": 121.4737,
        "latitude": 31.2304,
        "address": "上海市浦东新区",
        "status": 1,
        "commission_date": "2023-01-01T00:00:00Z",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

### 4.2 获取厂站详情

**GET** `/stations/{id}`

**响应示例**:

```json
{
  "code": 0,
  "data": {
    "id": "station-001",
    "code": "PV_SH_001",
    "name": "上海光伏电站1号",
    "type": "pv",
    "sub_region_id": "sub-region-001",
    "capacity": 50.0,
    "voltage_level": "35kV",
    "longitude": 121.4737,
    "latitude": 31.2304,
    "address": "上海市浦东新区",
    "status": 1,
    "statistics": {
      "device_count": 120,
      "point_count": 5000,
      "online_devices": 118,
      "active_alarms": 3
    },
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### 4.3 创建厂站

**POST** `/stations`

**请求体**:

```json
{
  "code": "PV_SH_002",
  "name": "上海光伏电站2号",
  "type": "pv",
  "sub_region_id": "sub-region-001",
  "capacity": 100.0,
  "voltage_level": "35kV",
  "longitude": 121.4737,
  "latitude": 31.2304,
  "address": "上海市浦东新区",
  "commission_date": "2024-01-01T00:00:00Z",
  "description": "100MW光伏电站"
}
```

### 4.4 更新厂站

**PUT** `/stations/{id}`

### 4.5 删除厂站

**DELETE** `/stations/{id}`

### 4.6 获取厂站设备列表

**GET** `/stations/{id}/devices`

**响应示例**:

```json
{
  "code": 0,
  "data": {
    "total": 120,
    "items": [
      {
        "id": "device-001",
        "code": "INV_001",
        "name": "1号逆变器",
        "type": "inverter",
        "status": 1,
        "last_online": "2024-03-01T10:30:00Z"
      }
    ]
  }
}
```

### 4.7 获取厂站采集点列表

**GET** `/stations/{id}/points`

---

## 5. 设备管理API

### 5.1 获取设备列表

**GET** `/devices`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| station_id | string | 否 | 厂站ID |
| type | string | 否 | 设备类型 |
| status | int | 否 | 状态 |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

**响应示例**:

```json
{
  "code": 0,
  "data": {
    "total": 120,
    "items": [
      {
        "id": "device-001",
        "code": "INV_001",
        "name": "1号逆变器",
        "type": "inverter",
        "station_id": "station-001",
        "manufacturer": "华为",
        "model": "SUN2000-100KTL",
        "rated_power": 100.0,
        "protocol": "modbus",
        "ip_address": "192.168.1.101",
        "port": 502,
        "slave_id": 1,
        "status": 1,
        "last_online": "2024-03-01T10:30:00Z",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

### 5.2 获取设备详情

**GET** `/devices/{id}`

### 5.3 创建设备

**POST** `/devices`

**请求体**:

```json
{
  "code": "INV_002",
  "name": "2号逆变器",
  "type": "inverter",
  "station_id": "station-001",
  "manufacturer": "华为",
  "model": "SUN2000-100KTL",
  "rated_power": 100.0,
  "protocol": "modbus",
  "ip_address": "192.168.1.102",
  "port": 502,
  "slave_id": 2,
  "description": "100kW组串式逆变器"
}
```

### 5.4 更新设备

**PUT** `/devices/{id}`

### 5.5 删除设备

**DELETE** `/devices/{id}`

### 5.6 获取设备采集点列表

**GET** `/devices/{id}/points`

---

## 6. 采集点管理API

### 6.1 获取采集点列表

**GET** `/points`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| device_id | string | 否 | 设备ID |
| station_id | string | 否 | 厂站ID |
| type | string | 否 | 点类型：yaoxin/yaoc/yaokong/setpoint/diandu |
| status | int | 否 | 状态 |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

**响应示例**:

```json
{
  "code": 0,
  "data": {
    "total": 5000,
    "items": [
      {
        "id": "point-001",
        "code": "INV_001_P",
        "name": "有功功率",
        "type": "yaoc",
        "device_id": "device-001",
        "station_id": "station-001",
        "unit": "kW",
        "precision": 2,
        "min_value": 0,
        "max_value": 100,
        "protocol": "modbus",
        "address": 40001,
        "scan_interval": 1000,
        "deadband": 0.1,
        "is_alarm": true,
        "alarm_high": 95,
        "alarm_low": 0,
        "status": 1,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

### 6.2 获取采集点详情

**GET** `/points/{id}`

### 6.3 创建采集点

**POST** `/points`

**请求体**:

```json
{
  "code": "INV_001_U",
  "name": "电压",
  "type": "yaoc",
  "device_id": "device-001",
  "station_id": "station-001",
  "unit": "V",
  "precision": 1,
  "min_value": 0,
  "max_value": 500,
  "protocol": "modbus",
  "address": 40002,
  "scan_interval": 1000,
  "deadband": 0.5,
  "is_alarm": true,
  "alarm_high": 450,
  "alarm_low": 350
}
```

### 6.4 批量创建采集点

**POST** `/points/batch`

**请求体**:

```json
{
  "device_id": "device-001",
  "points": [
    {
      "code": "INV_001_P",
      "name": "有功功率",
      "type": "yaoc",
      "unit": "kW",
      "address": 40001
    },
    {
      "code": "INV_001_Q",
      "name": "无功功率",
      "type": "yaoc",
      "unit": "kVar",
      "address": 40002
    }
  ]
}
```

### 6.5 更新采集点

**PUT** `/points/{id}`

### 6.6 删除采集点

**DELETE** `/points/{id}`

---

## 7. 告警管理API

### 7.1 获取告警列表

**GET** `/alarms`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| station_id | string | 否 | 厂站ID |
| device_id | string | 否 | 设备ID |
| level | int | 否 | 告警级别：1-提示，2-一般，3-重要，4-紧急 |
| status | int | 否 | 状态：1-活动，2-已确认，3-已清除 |
| start_time | int64 | 否 | 开始时间戳 |
| end_time | int64 | 否 | 结束时间戳 |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

**响应示例**:

```json
{
  "code": 0,
  "data": {
    "total": 50,
    "items": [
      {
        "id": "alarm-001",
        "point_id": "point-001",
        "device_id": "device-001",
        "station_id": "station-001",
        "type": "limit",
        "level": 3,
        "title": "逆变器温度过高",
        "message": "1号逆变器温度为87.5°C，已超过85°C阈值",
        "value": 87.5,
        "threshold": 85.0,
        "status": 1,
        "triggered_at": "2024-03-01T10:30:00Z",
        "acknowledged_at": null,
        "cleared_at": null,
        "acknowledged_by": null,
        "created_at": "2024-03-01T10:30:00Z"
      }
    ]
  }
}
```

### 7.2 获取告警详情

**GET** `/alarms/{id}`

### 7.3 确认告警

**PUT** `/alarms/{id}/ack`

**请求体**:

```json
{
  "acknowledged_by": "user-001",
  "comment": "已安排现场检查"
}
```

**响应示例**:

```json
{
  "code": 0,
  "data": {
    "id": "alarm-001",
    "status": 2,
    "acknowledged_at": "2024-03-01T10:35:00Z",
    "acknowledged_by": "user-001"
  },
  "message": "告警已确认"
}
```

### 7.4 清除告警

**PUT** `/alarms/{id}/clear`

**请求体**:

```json
{
  "comment": "故障已排除"
}
```

### 7.5 批量确认告警

**PUT** `/alarms/batch/ack`

**请求体**:

```json
{
  "alarm_ids": ["alarm-001", "alarm-002", "alarm-003"],
  "acknowledged_by": "user-001",
  "comment": "批量确认"
}
```

### 7.6 获取告警统计

**GET** `/alarms/statistics`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| station_id | string | 否 | 厂站ID |
| start_time | int64 | 否 | 开始时间戳 |
| end_time | int64 | 否 | 结束时间戳 |

**响应示例**:

```json
{
  "code": 0,
  "data": {
    "total": 150,
    "by_level": {
      "1": 50,
      "2": 60,
      "3": 30,
      "4": 10
    },
    "by_status": {
      "active": 20,
      "acknowledged": 30,
      "cleared": 100
    },
    "by_type": {
      "limit": 80,
      "communication": 40,
      "device": 30
    }
  }
}
```

---

## 8. 数据查询API

### 8.1 获取实时数据

**GET** `/data/realtime`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| point_ids | string | 是 | 采集点ID列表，逗号分隔 |

**响应示例**:

```json
{
  "code": 0,
  "data": [
    {
      "point_id": "point-001",
      "point_code": "INV_001_P",
      "point_name": "有功功率",
      "value": 85.6,
      "quality": 192,
      "quality_desc": "正常",
      "timestamp": 1709500800000,
      "updated_at": "2024-03-01T10:30:00Z"
    },
    {
      "point_id": "point-002",
      "point_code": "INV_001_U",
      "point_name": "电压",
      "value": 380.5,
      "quality": 192,
      "quality_desc": "正常",
      "timestamp": 1709500800000,
      "updated_at": "2024-03-01T10:30:00Z"
    }
  ]
}
```

### 8.2 获取历史数据

**GET** `/data/history`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| point_id | string | 是 | 采集点ID |
| start_time | int64 | 是 | 开始时间戳（毫秒） |
| end_time | int64 | 是 | 结束时间戳（毫秒） |
| interval | string | 否 | 聚合间隔：1m/5m/15m/1h/1d |
| aggregate | string | 否 | 聚合方式：avg/max/min/sum |

**响应示例**:

```json
{
  "code": 0,
  "data": {
    "point_id": "point-001",
    "point_name": "有功功率",
    "unit": "kW",
    "start_time": 1709407200000,
    "end_time": 1709493600000,
    "interval": "1h",
    "values": [
      {
        "timestamp": 1709407200000,
        "value": 85.6,
        "quality": 192
      },
      {
        "timestamp": 1709410800000,
        "value": 87.2,
        "quality": 192
      }
    ]
  }
}
```

### 8.3 批量获取历史数据

**POST** `/data/history/batch`

**请求体**:

```json
{
  "point_ids": ["point-001", "point-002"],
  "start_time": 1709407200000,
  "end_time": 1709493600000,
  "interval": "1h",
  "aggregate": "avg"
}
```

### 8.4 获取统计数据

**GET** `/data/statistics`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| station_id | string | 是 | 厂站ID |
| metric | string | 是 | 指标：generation/consumption/efficiency |
| period | string | 是 | 周期：day/month/year |
| start_time | int64 | 是 | 开始时间戳 |
| end_time | int64 | 是 | 结束时间戳 |

**响应示例**:

```json
{
  "code": 0,
  "data": {
    "station_id": "station-001",
    "station_name": "上海光伏电站1号",
    "metric": "generation",
    "period": "day",
    "values": [
      {
        "period_start": "2024-03-01T00:00:00Z",
        "period_end": "2024-03-01T23:59:59Z",
        "value": 125.6,
        "unit": "MWh"
      }
    ]
  }
}
```

---

## 9. 控制操作API

### 9.1 遥控操作

**POST** `/control/operate`

**请求体**:

```json
{
  "point_id": "point-003",
  "value": 1,
  "operator": "user-001",
  "reason": "设备检修",
  "confirmation_required": true
}
```

**响应示例**:

```json
{
  "code": 0,
  "data": {
    "operation_id": "op-001",
    "point_id": "point-003",
    "point_name": "开关1",
    "value": 1,
    "value_desc": "合闸",
    "status": "pending",
    "confirmation_required": true,
    "created_at": "2024-03-01T10:30:00Z"
  },
  "message": "操作已提交，等待确认"
}
```

### 9.2 确认控制操作

**POST** `/control/operate/{operation_id}/confirm`

**请求体**:

```json
{
  "confirmed": true,
  "confirmer": "user-002",
  "comment": "确认执行"
}
```

### 9.3 参数设置

**POST** `/control/setpoint`

**请求体**:

```json
{
  "point_id": "point-004",
  "value": 50.0,
  "operator": "user-001",
  "reason": "调整功率设定值"
}
```

### 9.4 获取操作记录

**GET** `/control/operations`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| station_id | string | 否 | 厂站ID |
| operator | string | 否 | 操作人 |
| start_time | int64 | 否 | 开始时间戳 |
| end_time | int64 | 否 | 结束时间戳 |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

---

## 10. AI服务API

### 10.1 智能问答

**POST** `/ai/qa`

**请求体**:

```json
{
  "question": "1号逆变器为什么报警？",
  "session_id": "session-001",
  "context": {
    "station_id": "station-001"
  }
}
```

**响应示例**:

```json
{
  "code": 0,
  "data": {
    "answer": "1号逆变器温度为87.5°C，已超过85°C阈值，触发了高温告警。建议检查散热系统是否正常工作，清理散热片灰尘，确保通风良好。",
    "confidence": 0.95,
    "sources": [
      {
        "type": "alarm",
        "id": "alarm-001",
        "title": "逆变器温度过高"
      }
    ],
    "suggestions": [
      "检查散热风扇是否正常运转",
      "清理逆变器散热片",
      "检查环境温度是否过高"
    ],
    "session_id": "session-001"
  }
}
```

### 10.2 智能配置建议

**POST** `/ai/config/suggest`

**请求体**:

```json
{
  "device_type": "inverter",
  "manufacturer": "华为",
  "model": "SUN2000-100KTL",
  "protocol": "modbus"
}
```

**响应示例**:

```json
{
  "code": 0,
  "data": {
    "suggested_points": [
      {
        "code": "P",
        "name": "有功功率",
        "type": "yaoc",
        "unit": "kW",
        "address": 40001,
        "description": "逆变器输出有功功率"
      },
      {
        "code": "U",
        "name": "电压",
        "type": "yaoc",
        "unit": "V",
        "address": 40002,
        "description": "逆变器输出电压"
      }
    ],
    "confidence": 0.92
  }
}
```

### 10.3 故障诊断

**POST** `/ai/diagnosis`

**请求体**:

```json
{
  "device_id": "device-001",
  "symptoms": ["温度过高", "功率下降"]
}
```

**响应示例**:

```json
{
  "code": 0,
  "data": {
    "diagnosis": [
      {
        "possible_cause": "散热系统故障",
        "probability": 0.75,
        "suggestions": [
          "检查散热风扇是否正常运转",
          "检查散热片是否堵塞"
        ]
      },
      {
        "possible_cause": "环境温度过高",
        "probability": 0.15,
        "suggestions": [
          "检查机房通风情况",
          "考虑增加空调设备"
        ]
      }
    ]
  }
}
```

---

## 11. 用户管理API

### 11.1 获取用户列表

**GET** `/users`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| status | int | 否 | 状态 |
| role_id | string | 否 | 角色ID |
| keyword | string | 否 | 搜索关键字 |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

**响应示例**:

```json
{
  "code": 0,
  "data": {
    "total": 50,
    "items": [
      {
        "id": "user-001",
        "username": "admin",
        "email": "admin@example.com",
        "phone": "13800138000",
        "real_name": "管理员",
        "status": 1,
        "roles": [
          {
            "id": "role-001",
            "name": "系统管理员"
          }
        ],
        "last_login": "2024-03-01T10:00:00Z",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

### 11.2 创建用户

**POST** `/users`

**请求体**:

```json
{
  "username": "operator01",
  "password": "Operator@123",
  "email": "operator@example.com",
  "phone": "13800138001",
  "real_name": "操作员01",
  "role_ids": ["role-002"]
}
```

### 11.3 更新用户

**PUT** `/users/{id}`

### 11.4 删除用户

**DELETE** `/users/{id}`

### 11.5 修改密码

**PUT** `/users/{id}/password`

**请求体**:

```json
{
  "old_password": "OldPassword@123",
  "new_password": "NewPassword@456"
}
```

---

## 12. 角色权限API

### 12.1 获取角色列表

**GET** `/roles`

**响应示例**:

```json
{
  "code": 0,
  "data": [
    {
      "id": "role-001",
      "name": "系统管理员",
      "description": "拥有所有权限",
      "permissions": ["*"],
      "user_count": 2,
      "created_at": "2024-01-01T00:00:00Z"
    },
    {
      "id": "role-002",
      "name": "运维人员",
      "description": "负责日常运维工作",
      "permissions": ["stations:read", "devices:read", "alarms:*"],
      "user_count": 10,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### 12.2 创建角色

**POST** `/roles`

**请求体**:

```json
{
  "name": "巡检员",
  "description": "负责设备巡检",
  "permissions": [
    "stations:read",
    "devices:read",
    "points:read",
    "alarms:read"
  ]
}
```

### 12.3 更新角色

**PUT** `/roles/{id}`

### 12.4 删除角色

**DELETE** `/roles/{id}`

### 12.5 获取权限列表

**GET** `/permissions`

**响应示例**:

```json
{
  "code": 0,
  "data": [
    {
      "resource": "stations",
      "actions": ["read", "write", "delete"],
      "description": "厂站管理"
    },
    {
      "resource": "devices",
      "actions": ["read", "write", "delete"],
      "description": "设备管理"
    },
    {
      "resource": "alarms",
      "actions": ["read", "ack", "clear"],
      "description": "告警管理"
    }
  ]
}
```

---

## 13. WebSocket接口

### 13.1 实时数据推送

**连接地址**: `ws://localhost:8080/ws/realtime`

**认证方式**: URL参数 `?token=<access_token>`

#### 订阅消息

```json
{
  "action": "subscribe",
  "point_ids": ["point-001", "point-002"]
}
```

#### 取消订阅

```json
{
  "action": "unsubscribe",
  "point_ids": ["point-001"]
}
```

#### 推送消息格式

```json
{
  "type": "data",
  "data": {
    "point_id": "point-001",
    "value": 85.6,
    "quality": 192,
    "timestamp": 1709500800000
  }
}
```

### 13.2 告警推送

**连接地址**: `ws://localhost:8080/ws/alarm`

#### 订阅消息

```json
{
  "action": "subscribe",
  "filters": {
    "station_ids": ["station-001"],
    "levels": [3, 4]
  }
}
```

#### 推送消息格式

```json
{
  "type": "alarm",
  "data": {
    "id": "alarm-001",
    "level": 3,
    "title": "逆变器温度过高",
    "message": "1号逆变器温度为87.5°C",
    "station_id": "station-001",
    "triggered_at": "2024-03-01T10:30:00Z"
  }
}
```

---

## 14. 数据类型定义

### 14.1 PointType 采集点类型

| 值 | 名称 | 说明 |
|------|------|------|
| yaoxin | 遥信 | 开关量状态信号 |
| yaoc | 遥测 | 模拟量测量值 |
| yaokong | 遥控 | 远程控制点 |
| setpoint | 遥调 | 参数设置点 |
| diandu | 电度 | 累计电量值 |

### 14.2 StationType 厂站类型

| 值 | 名称 | 说明 |
|------|------|------|
| pv | 光伏电站 | 太阳能光伏发电站 |
| wind | 风电场 | 风力发电场 |
| ess | 储能站 | 储能电站 |
| hybrid | 混合电站 | 光储、风储等混合电站 |

### 14.3 DeviceType 设备类型

| 值 | 名称 | 说明 |
|------|------|------|
| inverter | 逆变器 | 光伏逆变器 |
| meter | 电表 | 电能计量表 |
| transformer | 变压器 | 电力变压器 |
| switch | 开关 | 断路器、隔离开关 |
| combiner | 汇流箱 | 光伏汇流箱 |
| weather | 气象站 | 气象监测设备 |

### 14.4 AlarmLevel 告警级别

| 值 | 名称 | 颜色 | 说明 |
|------|------|------|------|
| 1 | 提示 | 蓝色 | 一般提示信息 |
| 2 | 一般 | 黄色 | 需要关注的问题 |
| 3 | 重要 | 橙色 | 需要及时处理 |
| 4 | 紧急 | 红色 | 需要立即处理 |

### 14.5 AlarmStatus 告警状态

| 值 | 名称 | 说明 |
|------|------|------|
| 1 | 活动 | 告警正在发生 |
| 2 | 已确认 | 告警已被确认 |
| 3 | 已清除 | 告警已清除 |

### 14.6 QualityCode 质量码

| 值 | 说明 |
|------|------|
| 192 | 正常 |
| 0 | 无效 |
| 64 | 可疑 |
| 128 | 旧数据 |

---

## 15. 调用示例

### 15.1 cURL示例

#### 登录获取Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Admin@123"}'
```

#### 获取厂站列表

```bash
curl -X GET "http://localhost:8080/api/v1/stations?type=pv&page=1&page_size=20" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### 创建设备

```bash
curl -X POST http://localhost:8080/api/v1/devices \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "code": "INV_001",
    "name": "1号逆变器",
    "type": "inverter",
    "station_id": "station-001",
    "protocol": "modbus",
    "ip_address": "192.168.1.101",
    "port": 502
  }'
```

### 15.2 JavaScript示例

```javascript
// 使用axios调用API
const axios = require('axios');

const api = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
  timeout: 10000,
  headers: {
    'Authorization': `Bearer ${token}`
  }
});

// 获取厂站列表
async function getStations() {
  const response = await api.get('/stations', {
    params: {
      type: 'pv',
      page: 1,
      page_size: 20
    }
  });
  return response.data;
}

// 获取实时数据
async function getRealtimeData(pointIds) {
  const response = await api.get('/data/realtime', {
    params: {
      point_ids: pointIds.join(',')
    }
  });
  return response.data;
}
```

### 15.3 Go示例

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

type Client struct {
    baseURL    string
    token      string
    httpClient *http.Client
}

func NewClient(baseURL, token string) *Client {
    return &Client{
        baseURL: baseURL,
        token:   token,
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

func (c *Client) GetStations(ctx context.Context, stationType string, page, pageSize int) (*StationListResponse, error) {
    url := fmt.Sprintf("%s/stations?type=%s&page=%d&page_size=%d", 
        c.baseURL, stationType, page, pageSize)
    
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", "Bearer "+c.token)
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    
    var result StationListResponse
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, err
    }
    
    return &result, nil
}

type StationListResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    struct {
        Total    int       `json:"total"`
        Items    []Station `json:"items"`
    } `json:"data"`
}

type Station struct {
    ID       string  `json:"id"`
    Code     string  `json:"code"`
    Name     string  `json:"name"`
    Type     string  `json:"type"`
    Capacity float64 `json:"capacity"`
    Status   int     `json:"status"`
}
```

### 15.4 Python示例

```python
import requests

class NEMClient:
    def __init__(self, base_url, token):
        self.base_url = base_url
        self.token = token
        self.headers = {
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        }
    
    def get_stations(self, station_type=None, page=1, page_size=20):
        """获取厂站列表"""
        params = {
            'page': page,
            'page_size': page_size
        }
        if station_type:
            params['type'] = station_type
        
        response = requests.get(
            f'{self.base_url}/stations',
            headers=self.headers,
            params=params
        )
        return response.json()
    
    def get_realtime_data(self, point_ids):
        """获取实时数据"""
        response = requests.get(
            f'{self.base_url}/data/realtime',
            headers=self.headers,
            params={'point_ids': ','.join(point_ids)}
        )
        return response.json()
    
    def create_alarm_ack(self, alarm_id, acknowledged_by, comment=''):
        """确认告警"""
        response = requests.put(
            f'{self.base_url}/alarms/{alarm_id}/ack',
            headers=self.headers,
            json={
                'acknowledged_by': acknowledged_by,
                'comment': comment
            }
        )
        return response.json()

# 使用示例
client = NEMClient('http://localhost:8080/api/v1', 'your-token')

# 获取光伏电站列表
stations = client.get_stations(station_type='pv')
print(stations)

# 获取实时数据
realtime = client.get_realtime_data(['point-001', 'point-002'])
print(realtime)
```

---

## 16. 附录

### 16.1 Swagger文档

访问地址：`http://localhost:8080/swagger/index.html`

### 16.2 Postman集合

可导入Postman集合进行API测试，集合文件位于：`docs/postman_collection.json`

### 16.3 变更记录

| 版本 | 日期 | 变更内容 | 变更人 |
|------|------|----------|--------|
| v1.0.0 | 2024-03-01 | 初始版本 | API团队 |
