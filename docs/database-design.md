# 数据库设计文档

## 1. 概述

### 1.1 数据库选型

| 数据库 | 用途 | 说明 |
|--------|------|------|
| PostgreSQL | 关系数据存储 | 配置数据、业务数据 |
| Redis | 缓存存储 | 实时数据、会话数据 |
| Doris/ClickHouse | 时序数据存储 | 历史数据、统计数据 |

---

## 2. PostgreSQL 表结构设计

### 2.1 区域表 (regions)

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | VARCHAR(36) | PRIMARY KEY | UUID主键 |
| code | VARCHAR(100) | UNIQUE, NOT NULL | 区域编码 |
| name | VARCHAR(200) | NOT NULL | 区域名称 |
| parent_id | VARCHAR(36) | FOREIGN KEY | 父区域ID |
| level | INTEGER | DEFAULT 1 | 层级 |
| sort_order | INTEGER | DEFAULT 0 | 排序号 |
| description | TEXT | | 描述 |
| created_at | TIMESTAMP | DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT NOW() | 更新时间 |

**索引设计**:
- `idx_regions_parent_id` ON (parent_id)
- `idx_regions_level` ON (level)

### 2.2 子区域表 (sub_regions)

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | VARCHAR(36) | PRIMARY KEY | UUID主键 |
| code | VARCHAR(100) | UNIQUE, NOT NULL | 子区域编码 |
| name | VARCHAR(200) | NOT NULL | 子区域名称 |
| region_id | VARCHAR(36) | FOREIGN KEY, NOT NULL | 所属区域ID |
| description | TEXT | | 描述 |
| created_at | TIMESTAMP | DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT NOW() | 更新时间 |

### 2.3 厂站表 (stations)

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | VARCHAR(36) | PRIMARY KEY | UUID主键 |
| code | VARCHAR(100) | UNIQUE, NOT NULL | 厂站编码 |
| name | VARCHAR(200) | NOT NULL | 厂站名称 |
| type | VARCHAR(50) | NOT NULL | 厂站类型 |
| sub_region_id | VARCHAR(36) | FOREIGN KEY | 所属子区域ID |
| capacity | DOUBLE PRECISION | | 装机容量(MW) |
| voltage_level | VARCHAR(50) | | 电压等级 |
| longitude | DOUBLE PRECISION | | 经度 |
| latitude | DOUBLE PRECISION | | 纬度 |
| address | VARCHAR(500) | | 地址 |
| status | INTEGER | DEFAULT 1 | 状态 |
| created_at | TIMESTAMP | DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT NOW() | 更新时间 |

**索引设计**:
- `idx_stations_sub_region_id` ON (sub_region_id)
- `idx_stations_type` ON (type)
- `idx_stations_status` ON (status)

### 2.4 设备表 (devices)

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | VARCHAR(36) | PRIMARY KEY | UUID主键 |
| code | VARCHAR(100) | UNIQUE, NOT NULL | 设备编码 |
| name | VARCHAR(200) | NOT NULL | 设备名称 |
| type | VARCHAR(50) | NOT NULL | 设备类型 |
| station_id | VARCHAR(36) | FOREIGN KEY | 所属厂站ID |
| manufacturer | VARCHAR(100) | | 制造商 |
| model | VARCHAR(100) | | 型号 |
| rated_power | DOUBLE PRECISION | | 额定功率(kW) |
| protocol | VARCHAR(50) | | 通信协议 |
| ip_address | VARCHAR(50) | | IP地址 |
| port | INTEGER | | 端口号 |
| slave_id | INTEGER | | 从站地址 |
| status | INTEGER | DEFAULT 0 | 状态 |
| last_online | TIMESTAMP | | 最后在线时间 |
| created_at | TIMESTAMP | DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT NOW() | 更新时间 |

**索引设计**:
- `idx_devices_station_id` ON (station_id)
- `idx_devices_type` ON (type)
- `idx_devices_protocol` ON (protocol)

### 2.5 采集点表 (points)

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | VARCHAR(36) | PRIMARY KEY | UUID主键 |
| code | VARCHAR(100) | UNIQUE, NOT NULL | 采集点编码 |
| name | VARCHAR(200) | NOT NULL | 采集点名称 |
| type | VARCHAR(20) | NOT NULL | 点类型 |
| device_id | VARCHAR(36) | FOREIGN KEY | 所属设备ID |
| station_id | VARCHAR(36) | FOREIGN KEY | 所属厂站ID |
| unit | VARCHAR(20) | | 单位 |
| precision | INTEGER | | 精度 |
| min_value | DOUBLE PRECISION | | 最小值 |
| max_value | DOUBLE PRECISION | | 最大值 |
| protocol | VARCHAR(50) | | 通信协议 |
| address | INTEGER | | 寄存器地址 |
| scan_interval | INTEGER | | 采集间隔(ms) |
| deadband | DOUBLE PRECISION | | 死区值 |
| is_alarm | BOOLEAN | DEFAULT FALSE | 是否告警 |
| alarm_high | DOUBLE PRECISION | | 高限值 |
| alarm_low | DOUBLE PRECISION | | 低限值 |
| status | INTEGER | DEFAULT 1 | 状态 |
| created_at | TIMESTAMP | DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT NOW() | 更新时间 |

**索引设计**:
- `idx_points_device_id` ON (device_id)
- `idx_points_station_id` ON (station_id)
- `idx_points_type` ON (type)
- `idx_points_protocol` ON (protocol)

### 2.6 告警表 (alarms)

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | VARCHAR(36) | PRIMARY KEY | UUID主键 |
| point_id | VARCHAR(36) | FOREIGN KEY | 关联采集点ID |
| device_id | VARCHAR(36) | FOREIGN KEY | 关联设备ID |
| station_id | VARCHAR(36) | FOREIGN KEY | 关联厂站ID |
| type | VARCHAR(20) | NOT NULL | 告警类型 |
| level | INTEGER | NOT NULL | 告警级别 |
| title | VARCHAR(200) | NOT NULL | 告警标题 |
| message | TEXT | | 告警详情 |
| value | DOUBLE PRECISION | | 触发值 |
| threshold | DOUBLE PRECISION | | 阈值 |
| status | INTEGER | DEFAULT 1 | 状态 |
| triggered_at | TIMESTAMP | NOT NULL | 触发时间 |
| acknowledged_at | TIMESTAMP | | 确认时间 |
| cleared_at | TIMESTAMP | | 清除时间 |
| acknowledged_by | VARCHAR(100) | | 确认人 |
| created_at | TIMESTAMP | DEFAULT NOW() | 创建时间 |

**索引设计**:
- `idx_alarms_point_id` ON (point_id)
- `idx_alarms_station_id` ON (station_id)
- `idx_alarms_status` ON (status)
- `idx_alarms_level` ON (level)
- `idx_alarms_triggered_at` ON (triggered_at)

---

## 3. Redis 数据结构设计

### 3.1 Key设计规范

| Key模式 | 类型 | TTL | 说明 |
|---------|------|-----|------|
| `nem:realtime:{point_id}` | Hash | 1天 | 实时数据 |
| `nem:device:status:{device_id}` | String | 永久 | 设备状态 |
| `nem:alarm:active:{alarm_id}` | Hash | 7天 | 活动告警 |
| `nem:alarm:count` | Hash | 永久 | 告警计数 |

### 3.2 实时数据结构

```
Key: nem:realtime:{point_id}
Type: Hash
Fields:
  - value: 采集值
  - quality: 质量码
  - timestamp: 时间戳
TTL: 1天
```

---

## 4. Doris 时序数据表设计

### 4.1 历史数据表

```sql
CREATE TABLE IF NOT EXISTS nem_history_data (
    point_id VARCHAR(36) NOT NULL COMMENT '采集点ID',
    station_id VARCHAR(36) NOT NULL COMMENT '厂站ID',
    value DOUBLE COMMENT '采集值',
    quality INT COMMENT '质量码',
    timestamp DATETIME NOT NULL COMMENT '时间戳'
) ENGINE=OLAP
DUPLICATE KEY(point_id, timestamp)
PARTITION BY RANGE(timestamp) ()
DISTRIBUTED BY HASH(point_id) BUCKETS 10
PROPERTIES (
    "dynamic_partition.enable" = "true",
    "dynamic_partition.time_unit" = "DAY",
    "dynamic_partition.start" = "-30",
    "dynamic_partition.end" = "3"
);
```

### 4.2 统计数据表

```sql
CREATE TABLE IF NOT EXISTS nem_statistics_data (
    station_id VARCHAR(36) NOT NULL,
    dimension VARCHAR(100) NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    metric_value DOUBLE,
    period_type VARCHAR(20) NOT NULL,
    period_start DATETIME NOT NULL
) ENGINE=OLAP
AGGREGATE KEY(station_id, dimension, metric_name, period_type, period_start)
PARTITION BY RANGE(period_start) ()
DISTRIBUTED BY HASH(station_id) BUCKETS 10;
```

---

## 5. 数据迁移策略

### 5.1 版本迁移

使用数据库迁移脚本管理版本:

```
scripts/migrations/
├── 001_init_schema.sql
├── 002_add_alarm_rules.sql
└── 003_add_compute_points.sql
```

### 5.2 数据备份

- PostgreSQL: 每日全量备份 + WAL增量备份
- Redis: RDB快照 + AOF日志
- Doris: 冷数据归档到对象存储

---

## 6. 性能优化策略

### 6.1 索引优化

- 高频查询字段建立索引
- 组合索引遵循最左前缀原则
- 定期分析索引使用情况

### 6.2 分区策略

- 时序数据按天分区
- 历史数据自动归档
- 冷热数据分离存储

### 6.3 查询优化

- 避免全表扫描
- 使用覆盖索引
- 分页查询限制返回条数
