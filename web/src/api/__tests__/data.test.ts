import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import * as dataApi from '../data'

// Mock request utils
vi.mock('@/utils/request', () => ({
  get: vi.fn(),
  post: vi.fn()
}))

import { get, post } from '@/utils/request'

describe('Data API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('查询历史数据', () => {
    it('应该成功查询历史数据', async () => {
      const mockData = [
        { pointId: 1, timestamp: '2024-01-01 00:00:00', value: 25.5 },
        { pointId: 1, timestamp: '2024-01-01 01:00:00', value: 26.0 }
      ]

      vi.mocked(post).mockResolvedValue(mockData)

      const params = {
        pointIds: [1],
        startTime: '2024-01-01',
        endTime: '2024-01-02'
      }
      const result = await dataApi.queryHistoryData(params)

      expect(post).toHaveBeenCalledWith('/data/query', params)
      expect(result).toEqual(mockData)
    })
  })

  describe('查询实时数据', () => {
    it('应该成功查询实时数据', async () => {
      const mockData = {
        1: { value: 25.5, quality: 1, timestamp: Date.now() },
        2: { value: 220, quality: 1, timestamp: Date.now() }
      }

      vi.mocked(post).mockResolvedValue(mockData)

      const pointIds = [1, 2]
      const result = await dataApi.queryRealtimeData(pointIds)

      expect(post).toHaveBeenCalledWith('/data/realtime', { pointIds })
      expect(result).toEqual(mockData)
    })
  })

  describe('获取统计数据', () => {
    it('应该成功获取统计数据', async () => {
      const mockData = [
        { timestamp: '2024-01-01', values: { avg: 25, max: 30, min: 20 } },
        { timestamp: '2024-01-02', values: { avg: 26, max: 31, min: 21 } }
      ]

      vi.mocked(post).mockResolvedValue(mockData)

      const params = {
        pointIds: [1],
        startTime: '2024-01-01',
        endTime: '2024-01-31',
        interval: 'day' as const,
        aggregations: ['avg', 'max', 'min'] as Array<'avg' | 'max' | 'min'>
      }
      const result = await dataApi.getStatistics(params)

      expect(post).toHaveBeenCalledWith('/data/statistics', params)
      expect(result).toEqual(mockData)
    })
  })

  describe('导出数据', () => {
    it('应该成功导出CSV数据', async () => {
      const mockBlob = new Blob(['test,data'], { type: 'text/csv' })
      vi.mocked(post).mockResolvedValue(mockBlob)

      const params = {
        pointIds: [1],
        startTime: '2024-01-01',
        endTime: '2024-01-02',
        format: 'csv' as const
      }
      const result = await dataApi.exportData(params)

      expect(post).toHaveBeenCalledWith('/data/export', params, { responseType: 'blob' })
      expect(result).toBeInstanceOf(Blob)
    })

    it('应该成功导出Excel数据', async () => {
      const mockBlob = new Blob(['test'], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
      vi.mocked(post).mockResolvedValue(mockBlob)

      const params = {
        pointIds: [1],
        startTime: '2024-01-01',
        endTime: '2024-01-02',
        format: 'excel' as const
      }
      const result = await dataApi.exportData(params)

      expect(post).toHaveBeenCalledWith('/data/export', params, { responseType: 'blob' })
      expect(result).toBeInstanceOf(Blob)
    })
  })

  describe('获取电站发电量统计', () => {
    it('应该成功获取电站发电量统计', async () => {
      const mockData = [
        { timestamp: '2024-01-01', power: 100.5, energy: 1000 },
        { timestamp: '2024-01-02', power: 105.2, energy: 1050 }
      ]

      vi.mocked(get).mockResolvedValue(mockData)

      const params = {
        stationId: 1,
        startTime: '2024-01-01',
        endTime: '2024-01-31',
        interval: 'day' as const
      }
      const result = await dataApi.getStationPowerStatistics(params)

      expect(get).toHaveBeenCalledWith('/stations/1/power-statistics', params)
      expect(result).toEqual(mockData)
    })
  })

  describe('获取设备运行统计', () => {
    it('应该成功获取设备运行统计', async () => {
      const mockData = {
        onlineTime: 720,
        offlineTime: 10,
        maintenanceTime: 5,
        faultTime: 5,
        availability: 0.98
      }

      vi.mocked(get).mockResolvedValue(mockData)

      const params = {
        deviceId: 1,
        startTime: '2024-01-01',
        endTime: '2024-01-31'
      }
      const result = await dataApi.getDeviceOperationStatistics(params)

      expect(get).toHaveBeenCalledWith('/devices/1/operation-statistics', params)
      expect(result).toEqual(mockData)
    })
  })

  describe('获取对比数据', () => {
    it('应该成功获取对比数据', async () => {
      const mockData = {
        current: [
          { pointId: 1, timestamp: '2024-01-01', value: 25 }
        ],
        compare: [
          { pointId: 1, timestamp: '2023-01-01', value: 23 }
        ]
      }

      vi.mocked(post).mockResolvedValue(mockData)

      const params = {
        pointIds: [1],
        startTime: '2024-01-01',
        endTime: '2024-01-31',
        compareType: 'year' as const
      }
      const result = await dataApi.getComparisonData(params)

      expect(post).toHaveBeenCalledWith('/data/comparison', params)
      expect(result).toEqual(mockData)
    })

    it('应该支持自定义对比时间段', async () => {
      const mockData = {
        current: [
          { pointId: 1, timestamp: '2024-01-01', value: 25 }
        ],
        compare: [
          { pointId: 1, timestamp: '2023-06-01', value: 22 }
        ]
      }

      vi.mocked(post).mockResolvedValue(mockData)

      const params = {
        pointIds: [1],
        startTime: '2024-01-01',
        endTime: '2024-01-31',
        compareType: 'custom' as const,
        compareStartTime: '2023-06-01',
        compareEndTime: '2023-06-30'
      }
      const result = await dataApi.getComparisonData(params)

      expect(post).toHaveBeenCalledWith('/data/comparison', params)
      expect(result).toEqual(mockData)
    })
  })

  describe('获取聚合数据', () => {
    it('应该成功获取聚合数据', async () => {
      const mockData = [
        { time: '2024-01-01 00:00', value: 25.5 },
        { time: '2024-01-01 01:00', value: 26.0 }
      ]

      vi.mocked(get).mockResolvedValue(mockData)

      const params = {
        pointId: 1,
        startTime: '2024-01-01',
        endTime: '2024-01-02',
        aggregation: 'avg' as const,
        groupBy: 'hour' as const
      }
      const result = await dataApi.getAggregatedData(params)

      expect(get).toHaveBeenCalledWith('/data/aggregated', params)
      expect(result).toEqual(mockData)
    })
  })
})
