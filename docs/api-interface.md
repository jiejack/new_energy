# API接口文档

## 1. 概述

### 1.1 基础信息

- **Base URL**: `http://localhost:8080/api/v1`
- **认证方式**: JWT Bearer Token
- **数据格式**: JSON
- **编码**: UTF-8

### 1.2 通用响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {},
  "timestamp": 1709500800000
}
```

### 1.3 错误码定义

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 2. 区域管理API

### 2.1 获取区域列表

**GET** `/regions`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| parent_id | string | 否 | 父区域ID |

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
      "sort_order": 1
    }
  ]
}
```

### 2.2 创建区域

**POST** `/regions`

**请求体**:

```json
{
  "code": "EAST_SH",
  "name": "上海子区域",
  "parent_id": "region-001",
  "level": 2
}
```

### 2.3 更新区域

**PUT** `/regions/{id}`

### 2.4 删除区域

**DELETE** `/regions/{id}`

---

## 3. 厂站管理API

### 3.1 获取厂站列表

**GET** `/stations`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sub_region_id | string | 否 | 子区域ID |
| type | string | 否 | 厂站类型 |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

**响应示例**:

```json
{
  "code": 0,
  "data": {
    "total": 100,
    "items": [
      {
        "id": "station-001",
        "code": "PV_SH_001",
        "name": "上海光伏电站1号",
        "type": "pv",
        "capacity": 50.0,
        "status": 1
      }
    ]
  }
}
```

### 3.2 创建厂站

**POST** `/stations`

**请求体**:

```json
{
  "code": "PV_SH_002",
  "name": "上海光伏电站2号",
  "type": "pv",
  "sub_region_id": "region-002",
  "capacity": 100.0,
  "voltage_level": "35kV"
}
```

### 3.3 获取厂站详情

**GET** `/stations/{id}`

### 3.4 获取厂站设备

**GET** `/stations/{id}/devices`

### 3.5 获取厂站采集点

**GET** `/stations/{id}/points`

---

## 4. 设备管理API

### 4.1 获取设备列表

**GET** `/devices`

### 4.2 创建设备

**POST** `/devices`

**请求体**:

```json
{
  "code": "INV_001",
  "name": "1号逆变器",
  "type": "inverter",
  "station_id": "station-001",
  "protocol": "modbus",
  "ip_address": "192.168.1.101",
  "port": 502,
  "slave_id": 1
}
```

### 4.3 获取设备详情

**GET** `/devices/{id}`

### 4.4 获取设备采集点

**GET** `/devices/{id}/points`

---

## 5. 采集点管理API

### 5.1 获取采集点列表

**GET** `/points`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| device_id | string | 否 | 设备ID |
| type | string | 否 | 点类型 |

### 5.2 创建采集点

**POST** `/points`

**请求体**:

```json
{
  "code": "INV_001_P",
  "name": "有功功率",
  "type": "yaoc",
  "device_id": "device-001",
  "unit": "kW",
  "address": 40001,
  "scan_interval": 1000,
  "is_alarm": true,
  "alarm_high": 95
}
```

### 5.3 批量创建采集点

**POST** `/points/batch`

---

## 6. 告警管理API

### 6.1 获取告警列表

**GET** `/alarms`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| station_id | string | 否 | 厂站ID |
| level | int | 否 | 告警级别 |
| status | int | 否 | 状态 |

**响应示例**:

```json
{
  "code": 0,
  "data": {
    "total": 50,
    "items": [
      {
        "id": "alarm-001",
        "type": "limit",
        "level": 3,
        "title": "逆变器温度过高",
        "value": 87.5,
        "threshold": 85.0,
        "status": 1,
        "triggered_at": "2024-03-01T10:30:00Z"
      }
    ]
  }
}
```

### 6.2 确认告警

**PUT** `/alarms/{id}/ack`

### 6.3 清除告警

**PUT** `/alarms/{id}/clear`

### 6.4 获取告警统计

**GET** `/alarms/statistics`

---

## 7. 数据查询API

### 7.1 获取实时数据

**GET** `/data/realtime`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| point_ids | string | 是 | 采集点ID列表 |

**响应示例**:

```json
{
  "code": 0,
  "data": [
    {
      "point_id": "point-001",
      "value": 85.6,
      "quality": 192,
      "timestamp": 1709500800000
    }
  ]
}
```

### 7.2 获取历史数据

**GET** `/data/history`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| point_id | string | 是 | 采集点ID |
| start_time | int64 | 是 | 开始时间戳 |
| end_time | int64 | 是 | 结束时间戳 |

---

## 8. 控制操作API

### 8.1 遥控操作

**POST** `/control/operate`

**请求体**:

```json
{
  "point_id": "point-001",
  "value": 1,
  "operator": "user-001",
  "reason": "设备检修"
}
```

### 8.2 参数设置

**POST** `/control/setpoint`

---

## 9. AI服务API

### 9.1 智能问答

**POST** `/ai/qa`

**请求体**:

```json
{
  "question": "1号逆变器为什么报警？",
  "session_id": "session-001"
}
```

**响应示例**:

```json
{
  "code": 0,
  "data": {
    "answer": "1号逆变器温度为87.5°C，已超过85°C阈值。",
    "confidence": 0.95
  }
}
```

### 9.2 智能配置

**POST** `/ai/config/suggest`

---

## 10. WebSocket接口

### 10.1 实时数据推送

**连接地址**: `ws://localhost:8080/ws/realtime`

**订阅消息**:

```json
{
  "action": "subscribe",
  "point_ids": ["point-001"]
}
```

### 10.2 告警推送

**连接地址**: `ws://localhost:8080/ws/alarm`

---

## 11. 数据类型定义

### 11.1 PointType 采集点类型

| 值 | 说明 |
|------|------|
| yaoxin | 遥信点 |
| yaoc | 遥测点 |
| yaokong | 遥控点 |
| setpoint | 设置点 |
| diandu | 电度点 |

### 11.2 StationType 厂站类型

| 值 | 说明 |
|------|------|
| pv | 光伏电站 |
| wind | 风电场 |
| ess | 储能站 |
| hybrid | 混合电站 |

### 11.3 DeviceType 设备类型

| 值 | 说明 |
|------|------|
| inverter | 逆变器 |
| meter | 电表 |
| transformer | 变压器 |
| switch | 开关 |

### 11.4 AlarmLevel 告警级别

| 值 | 说明 |
|------|------|
| 1 | 提示 |
| 2 | 一般 |
| 3 | 重要 |
| 4 | 紧急 |
