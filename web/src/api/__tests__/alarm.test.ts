import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import * as alarmApi from '../alarm'
import * as alarmRuleApi from '../alarm-rule'

// Mock request utils
vi.mock('@/utils/request', () => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  del: vi.fn()
}))

import { get, post, put, del } from '@/utils/request'

describe('Alarm API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('获取告警列表', () => {
    it('应该成功获取告警列表', async () => {
      const mockResponse = {
        list: [
          { id: 1, level: 'critical', message: '告警1' },
          { id: 2, level: 'major', message: '告警2' }
        ],
        total: 2,
        page: 1,
        pageSize: 10
      }

      vi.mocked(get).mockResolvedValue(mockResponse)

      const params = { page: 1, pageSize: 10 }
      const result = await alarmApi.getAlarmList(params)

      expect(get).toHaveBeenCalledWith('/alarms', params)
      expect(result).toEqual(mockResponse)
    })

    it('应该支持多条件筛选', async () => {
      const mockResponse = {
        list: [{ id: 1, level: 'critical' }],
        total: 1,
        page: 1,
        pageSize: 10
      }

      vi.mocked(get).mockResolvedValue(mockResponse)

      const params = {
        page: 1,
        pageSize: 10,
        level: 'critical' as const,
        status: 'active' as const,
        sourceType: 'device',
        sourceId: 1,
        startTime: '2024-01-01',
        endTime: '2024-01-31'
      }
      await alarmApi.getAlarmList(params)

      expect(get).toHaveBeenCalledWith('/alarms', params)
    })
  })

  describe('获取告警详情', () => {
    it('应该成功获取告警详情', async () => {
      const mockAlarm = {
        id: 1,
        level: 'critical',
        message: '测试告警',
        status: 'active'
      }

      vi.mocked(get).mockResolvedValue(mockAlarm)

      const result = await alarmApi.getAlarmDetail(1)

      expect(get).toHaveBeenCalledWith('/alarms/1')
      expect(result).toEqual(mockAlarm)
    })
  })

  describe('确认告警', () => {
    it('应该成功确认告警', async () => {
      vi.mocked(put).mockResolvedValue(undefined)

      await alarmApi.acknowledgeAlarm(1)

      expect(put).toHaveBeenCalledWith('/alarms/1/acknowledge')
    })
  })

  describe('批量确认告警', () => {
    it('应该成功批量确认告警', async () => {
      vi.mocked(post).mockResolvedValue(undefined)

      const ids = [1, 2, 3]
      await alarmApi.batchAcknowledgeAlarms(ids)

      expect(post).toHaveBeenCalledWith('/alarms/batch-acknowledge', { ids })
    })
  })

  describe('解决告警', () => {
    it('应该成功解决告警', async () => {
      vi.mocked(put).mockResolvedValue(undefined)

      await alarmApi.resolveAlarm(1)

      expect(put).toHaveBeenCalledWith('/alarms/1/resolve')
    })
  })

  describe('批量解决告警', () => {
    it('应该成功批量解决告警', async () => {
      vi.mocked(post).mockResolvedValue(undefined)

      const ids = [1, 2, 3]
      await alarmApi.batchResolveAlarms(ids)

      expect(post).toHaveBeenCalledWith('/alarms/batch-resolve', { ids })
    })
  })

  describe('获取告警统计', () => {
    it('应该成功获取告警统计', async () => {
      const mockStats = {
        total: 100,
        critical: 10,
        major: 20,
        minor: 30,
        warning: 40
      }

      vi.mocked(get).mockResolvedValue(mockStats)

      const result = await alarmApi.getAlarmStatistics()

      expect(get).toHaveBeenCalledWith('/alarms/statistics')
      expect(result).toEqual(mockStats)
    })
  })

  describe('获取告警趋势', () => {
    it('应该成功获取告警趋势', async () => {
      const mockTrend = [
        { time: '2024-01-01', count: 10 },
        { time: '2024-01-02', count: 15 }
      ]

      vi.mocked(get).mockResolvedValue(mockTrend)

      const params = {
        startTime: '2024-01-01',
        endTime: '2024-01-31',
        interval: 'day' as const
      }
      const result = await alarmApi.getAlarmTrend(params)

      expect(get).toHaveBeenCalledWith('/alarms/trend', params)
      expect(result).toEqual(mockTrend)
    })
  })

  describe('获取告警分布', () => {
    it('应该成功获取告警分布', async () => {
      const mockDistribution = [
        { name: 'critical', count: 10 },
        { name: 'major', count: 20 }
      ]

      vi.mocked(get).mockResolvedValue(mockDistribution)

      const params = {
        startTime: '2024-01-01',
        endTime: '2024-01-31',
        groupBy: 'level' as const
      }
      const result = await alarmApi.getAlarmDistribution(params)

      expect(get).toHaveBeenCalledWith('/alarms/distribution', params)
      expect(result).toEqual(mockDistribution)
    })
  })

  describe('告警规则管理', () => {
    it('应该成功创建告警规则', async () => {
      const mockRule = { id: '1', name: '规则1' }
      vi.mocked(post).mockResolvedValue(mockRule)

      const data = { name: '规则1', condition: 'value > 100' }
      const result = await alarmRuleApi.createAlarmRule(data as any)

      expect(post).toHaveBeenCalledWith('/alarm-rules', data)
      expect(result).toEqual(mockRule)
    })

    it('应该成功更新告警规则', async () => {
      const mockRule = { id: '1', name: '更新规则' }
      vi.mocked(post).mockResolvedValue(mockRule)

      const data = { name: '更新规则' }
      const result = await alarmRuleApi.updateAlarmRule('1', data as any)

      expect(post).toHaveBeenCalledWith('/alarm-rules/1', { ...data, _method: 'PUT' })
      expect(result).toEqual(mockRule)
    })

    it('应该成功删除告警规则', async () => {
      vi.mocked(del).mockResolvedValue(undefined)

      await alarmRuleApi.deleteAlarmRule('1')

      expect(del).toHaveBeenCalledWith('/alarm-rules/1')
    })

    it('应该成功获取告警规则列表', async () => {
      const mockResponse = {
        list: [{ id: '1', name: '规则1' }],
        total: 1,
        page: 1,
        pageSize: 10
      }

      vi.mocked(get).mockResolvedValue(mockResponse)

      const params = { page: 1, pageSize: 10 }
      const result = await alarmRuleApi.getAlarmRuleList(params)

      expect(get).toHaveBeenCalledWith('/alarm-rules', params)
      expect(result).toEqual(mockResponse)
    })
  })
})
