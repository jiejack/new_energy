import { describe, it, expect } from 'vitest'
import {
  stationTypeMapper,
  stationStatusMapper,
  deviceTypeMapper,
  deviceStatusMapper,
  pointTypeMapper,
  dataTypeMapper,
  alarmStatusMapper,
  alarmLevelMapper,
  regionLevelMapper
} from '../enums'

describe('enums', () => {
  describe('stationTypeMapper', () => {
    it('should return correct label for station types', () => {
      expect(stationTypeMapper.getLabel('solar')).toBe('光伏')
      expect(stationTypeMapper.getLabel('wind')).toBe('风电')
      expect(stationTypeMapper.getLabel('hydro')).toBe('水电')
      expect(stationTypeMapper.getLabel('storage')).toBe('储能')
    })

    it('should return default label for undefined', () => {
      expect(stationTypeMapper.getLabel()).toBe('-')
      expect(stationTypeMapper.getLabel(undefined as any)).toBe('-')
    })

    it('should return correct tag type for station types', () => {
      expect(stationTypeMapper.getTagType('solar')).toBe('warning')
      expect(stationTypeMapper.getTagType('wind')).toBe('success')
      expect(stationTypeMapper.getTagType('hydro')).toBe('primary')
      expect(stationTypeMapper.getTagType('storage')).toBe('info')
    })

    it('should return empty string for undefined tag type', () => {
      expect(stationTypeMapper.getTagType()).toBe('')
      expect(stationTypeMapper.getTagType(undefined as any)).toBe('')
    })
  })

  describe('stationStatusMapper', () => {
    it('should return correct label for station status', () => {
      expect(stationStatusMapper.getLabel('online')).toBe('在线')
      expect(stationStatusMapper.getLabel('offline')).toBe('离线')
      expect(stationStatusMapper.getLabel('maintenance')).toBe('维护')
      expect(stationStatusMapper.getLabel('fault')).toBe('故障')
    })

    it('should return correct tag type for station status', () => {
      expect(stationStatusMapper.getTagType('online')).toBe('success')
      expect(stationStatusMapper.getTagType('offline')).toBe('info')
      expect(stationStatusMapper.getTagType('maintenance')).toBe('warning')
      expect(stationStatusMapper.getTagType('fault')).toBe('danger')
    })
  })

  describe('deviceTypeMapper', () => {
    it('should return correct label for device types', () => {
      expect(deviceTypeMapper.getLabel('inverter')).toBe('逆变器')
      expect(deviceTypeMapper.getLabel('meter')).toBe('电表')
      expect(deviceTypeMapper.getLabel('sensor')).toBe('传感器')
      expect(deviceTypeMapper.getLabel('controller')).toBe('控制器')
      expect(deviceTypeMapper.getLabel('combiner')).toBe('汇流箱')
    })

    it('should return correct tag type for device types', () => {
      expect(deviceTypeMapper.getTagType('inverter')).toBe('primary')
      expect(deviceTypeMapper.getTagType('meter')).toBe('success')
      expect(deviceTypeMapper.getTagType('sensor')).toBe('warning')
      expect(deviceTypeMapper.getTagType('controller')).toBe('danger')
      expect(deviceTypeMapper.getTagType('combiner')).toBe('info')
    })
  })

  describe('deviceStatusMapper', () => {
    it('should return correct label for device status', () => {
      expect(deviceStatusMapper.getLabel('online')).toBe('在线')
      expect(deviceStatusMapper.getLabel('offline')).toBe('离线')
      expect(deviceStatusMapper.getLabel('maintenance')).toBe('维护')
      expect(deviceStatusMapper.getLabel('fault')).toBe('故障')
    })

    it('should return correct tag type for device status', () => {
      expect(deviceStatusMapper.getTagType('online')).toBe('success')
      expect(deviceStatusMapper.getTagType('offline')).toBe('info')
      expect(deviceStatusMapper.getTagType('maintenance')).toBe('warning')
      expect(deviceStatusMapper.getTagType('fault')).toBe('danger')
    })

    it('should have same mapping as station status', () => {
      const statuses = ['online', 'offline', 'maintenance', 'fault'] as const
      statuses.forEach(status => {
        expect(deviceStatusMapper.getLabel(status)).toBe(stationStatusMapper.getLabel(status))
        expect(deviceStatusMapper.getTagType(status)).toBe(stationStatusMapper.getTagType(status))
      })
    })
  })

  describe('pointTypeMapper', () => {
    it('should return correct label for point types', () => {
      expect(pointTypeMapper.getLabel('analog')).toBe('模拟量')
      expect(pointTypeMapper.getLabel('digital')).toBe('数字量')
      expect(pointTypeMapper.getLabel('pulse')).toBe('脉冲量')
    })

    it('should return correct tag type for point types', () => {
      expect(pointTypeMapper.getTagType('analog')).toBe('primary')
      expect(pointTypeMapper.getTagType('digital')).toBe('success')
      expect(pointTypeMapper.getTagType('pulse')).toBe('warning')
    })
  })

  describe('dataTypeMapper', () => {
    it('should return correct label for data types', () => {
      expect(dataTypeMapper.getLabel('float')).toBe('浮点数')
      expect(dataTypeMapper.getLabel('int')).toBe('整数')
      expect(dataTypeMapper.getLabel('bool')).toBe('布尔值')
      expect(dataTypeMapper.getLabel('string')).toBe('字符串')
    })

    it('should return correct tag type for data types', () => {
      expect(dataTypeMapper.getTagType('float')).toBe('primary')
      expect(dataTypeMapper.getTagType('int')).toBe('success')
      expect(dataTypeMapper.getTagType('bool')).toBe('warning')
      expect(dataTypeMapper.getTagType('string')).toBe('info')
    })
  })

  describe('alarmStatusMapper', () => {
    it('should return correct label for alarm status', () => {
      expect(alarmStatusMapper.getLabel('active')).toBe('活动')
      expect(alarmStatusMapper.getLabel('acknowledged')).toBe('已确认')
      expect(alarmStatusMapper.getLabel('resolved')).toBe('已解决')
    })

    it('should return correct tag type for alarm status', () => {
      expect(alarmStatusMapper.getTagType('active')).toBe('danger')
      expect(alarmStatusMapper.getTagType('acknowledged')).toBe('warning')
      expect(alarmStatusMapper.getTagType('resolved')).toBe('success')
    })
  })

  describe('alarmLevelMapper', () => {
    it('should return correct label for alarm levels', () => {
      expect(alarmLevelMapper.getLabel('critical')).toBe('严重')
      expect(alarmLevelMapper.getLabel('major')).toBe('重要')
      expect(alarmLevelMapper.getLabel('minor')).toBe('次要')
      expect(alarmLevelMapper.getLabel('warning')).toBe('警告')
    })

    it('should return correct tag type for alarm levels', () => {
      expect(alarmLevelMapper.getTagType('critical')).toBe('danger')
      expect(alarmLevelMapper.getTagType('major')).toBe('warning')
      expect(alarmLevelMapper.getTagType('minor')).toBe('info')
      expect(alarmLevelMapper.getTagType('warning')).toBe('primary')
    })
  })

  describe('regionLevelMapper', () => {
    it('should return correct label for region levels', () => {
      expect(regionLevelMapper.getLabel(1)).toBe('省级')
      expect(regionLevelMapper.getLabel(2)).toBe('市级')
      expect(regionLevelMapper.getLabel(3)).toBe('区/县级')
      expect(regionLevelMapper.getLabel(4)).toBe('乡镇/街道')
    })

    it('should return default label for unknown level', () => {
      expect(regionLevelMapper.getLabel(5)).toBe('未知')
      expect(regionLevelMapper.getLabel(0)).toBe('未知')
    })

    it('should return default label for undefined', () => {
      expect(regionLevelMapper.getLabel()).toBe('未知')
      expect(regionLevelMapper.getLabel(undefined as any)).toBe('未知')
    })
  })
})
