import type {
  StationType,
  StationStatus,
  DeviceType,
  DeviceStatus,
  PointType,
  DataType,
  AlarmStatus,
  AlarmLevel
} from '@/types'

export interface EnumMapping<T extends string> {
  label: Record<T, string>
  tagType: Record<T, 'primary' | 'success' | 'warning' | 'danger' | 'info'>
  defaultLabel?: string
}

export interface NumberEnumMapping {
  label: Record<number, string>
  tagType?: Record<number, 'primary' | 'success' | 'warning' | 'danger' | 'info'>
  defaultLabel?: string
}

export function createEnumMapper<T extends string>(config: EnumMapping<T>) {
  return {
    getLabel: (value?: T): string => {
      return value ? config.label[value] : config.defaultLabel || '-'
    },
    
    getTagType: (value?: T): 'primary' | 'success' | 'warning' | 'danger' | 'info' | '' => {
      return value ? config.tagType[value] : ''
    }
  }
}

export function createNumberEnumMapper(config: NumberEnumMapping) {
  return {
    getLabel: (value?: number): string => {
      return value !== undefined && value !== null ? (config.label[value] || config.defaultLabel || '-') : (config.defaultLabel || '-')
    },
    
    getTagType: (value?: number): 'primary' | 'success' | 'warning' | 'danger' | 'info' | '' => {
      return value !== undefined && value !== null && config.tagType ? (config.tagType[value] || '') : ''
    }
  }
}

export const stationTypeMapper = createEnumMapper<StationType>({
  label: {
    solar: '光伏',
    wind: '风电',
    hydro: '水电',
    storage: '储能'
  },
  tagType: {
    solar: 'warning',
    wind: 'success',
    hydro: 'primary',
    storage: 'info'
  }
})

export const stationStatusMapper = createEnumMapper<StationStatus>({
  label: {
    online: '在线',
    offline: '离线',
    maintenance: '维护',
    fault: '故障'
  },
  tagType: {
    online: 'success',
    offline: 'info',
    maintenance: 'warning',
    fault: 'danger'
  }
})

export const deviceTypeMapper = createEnumMapper<DeviceType>({
  label: {
    inverter: '逆变器',
    meter: '电表',
    sensor: '传感器',
    controller: '控制器',
    combiner: '汇流箱'
  },
  tagType: {
    inverter: 'primary',
    meter: 'success',
    sensor: 'warning',
    controller: 'danger',
    combiner: 'info'
  }
})

export const deviceStatusMapper = createEnumMapper<DeviceStatus>({
  label: {
    online: '在线',
    offline: '离线',
    maintenance: '维护',
    fault: '故障'
  },
  tagType: {
    online: 'success',
    offline: 'info',
    maintenance: 'warning',
    fault: 'danger'
  }
})

export const pointTypeMapper = createEnumMapper<PointType>({
  label: {
    analog: '模拟量',
    digital: '数字量',
    pulse: '脉冲量'
  },
  tagType: {
    analog: 'primary',
    digital: 'success',
    pulse: 'warning'
  }
})

export const dataTypeMapper = createEnumMapper<DataType>({
  label: {
    float: '浮点数',
    int: '整数',
    bool: '布尔值',
    string: '字符串'
  },
  tagType: {
    float: 'primary',
    int: 'success',
    bool: 'warning',
    string: 'info'
  }
})

export const alarmStatusMapper = createEnumMapper<AlarmStatus>({
  label: {
    active: '活动',
    acknowledged: '已确认',
    resolved: '已解决'
  },
  tagType: {
    active: 'danger',
    acknowledged: 'warning',
    resolved: 'success'
  }
})

export const alarmLevelMapper = createEnumMapper<AlarmLevel>({
  label: {
    critical: '严重',
    major: '重要',
    minor: '次要',
    warning: '警告'
  },
  tagType: {
    critical: 'danger',
    major: 'warning',
    minor: 'info',
    warning: 'primary'
  }
})

export const regionLevelMapper = createNumberEnumMapper({
  label: {
    1: '省级',
    2: '市级',
    3: '区/县级',
    4: '乡镇/街道'
  },
  defaultLabel: '未知'
})
