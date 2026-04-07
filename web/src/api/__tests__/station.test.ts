import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import * as stationApi from '../station'
import type { StationType, StationStatus } from '@/types'

// Mock request utils
vi.mock('@/utils/request', () => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  del: vi.fn()
}))

import { get, post, put, del } from '@/utils/request'

describe('Station API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('获取电站列表', () => {
    it('应该成功获取电站列表', async () => {
      const mockResponse = {
        list: [
          { id: 1, name: '电站1' },
          { id: 2, name: '电站2' }
        ],
        total: 2,
        page: 1,
        pageSize: 10
      }

      vi.mocked(get).mockResolvedValue(mockResponse)

      const params = { page: 1, pageSize: 10 }
      const result = await stationApi.getStationList(params)

      expect(get).toHaveBeenCalledWith('/stations', params)
      expect(result).toEqual(mockResponse)
    })

    it('应该支持关键字搜索', async () => {
      const mockResponse = {
        list: [{ id: 1, name: '测试电站' }],
        total: 1,
        page: 1,
        pageSize: 10
      }

      vi.mocked(get).mockResolvedValue(mockResponse)

      const params = { page: 1, pageSize: 10, keyword: '测试' }
      await stationApi.getStationList(params)

      expect(get).toHaveBeenCalledWith('/stations', params)
    })

    it('应该支持区域筛选', async () => {
      const mockResponse = {
        list: [{ id: 1, name: '电站1', regionId: 1 }],
        total: 1,
        page: 1,
        pageSize: 10
      }

      vi.mocked(get).mockResolvedValue(mockResponse)

      const params = { page: 1, pageSize: 10, regionId: 1 }
      await stationApi.getStationList(params)

      expect(get).toHaveBeenCalledWith('/stations', params)
    })

    it('应该支持类型和状态筛选', async () => {
      const mockResponse = {
        list: [{ id: 1, name: '电站1', type: 'solar', status: 'online' }],
        total: 1,
        page: 1,
        pageSize: 10
      }

      vi.mocked(get).mockResolvedValue(mockResponse)

      const params = { 
        page: 1, 
        pageSize: 10, 
        type: 'solar' as StationType, 
        status: 'online' as StationStatus 
      }
      await stationApi.getStationList(params)

      expect(get).toHaveBeenCalledWith('/stations', params)
    })
  })

  describe('获取电站详情', () => {
    it('应该成功获取电站详情', async () => {
      const mockStation = {
        id: 1,
        name: '测试电站',
        type: 'solar',
        status: 'online'
      }

      vi.mocked(get).mockResolvedValue(mockStation)

      const result = await stationApi.getStationDetail(1)

      expect(get).toHaveBeenCalledWith('/stations/1')
      expect(result).toEqual(mockStation)
    })
  })

  describe('创建电站', () => {
    it('应该成功创建电站', async () => {
      const mockStation = {
        id: 1,
        name: '新电站',
        type: 'solar'
      }

      vi.mocked(post).mockResolvedValue(mockStation)

      const data = {
        name: '新电站',
        type: 'solar' as const,
        regionId: 1
      }

      const result = await stationApi.createStation(data)

      expect(post).toHaveBeenCalledWith('/stations', data)
      expect(result).toEqual(mockStation)
    })
  })

  describe('更新电站', () => {
    it('应该成功更新电站', async () => {
      const mockStation = {
        id: 1,
        name: '更新电站',
        type: 'solar'
      }

      vi.mocked(put).mockResolvedValue(mockStation)

      const data = { name: '更新电站' }
      const result = await stationApi.updateStation(1, data)

      expect(put).toHaveBeenCalledWith('/stations/1', data)
      expect(result).toEqual(mockStation)
    })
  })

  describe('删除电站', () => {
    it('应该成功删除电站', async () => {
      vi.mocked(del).mockResolvedValue(undefined)

      await stationApi.deleteStation(1)

      expect(del).toHaveBeenCalledWith('/stations/1')
    })
  })

  describe('批量删除电站', () => {
    it('应该成功批量删除电站', async () => {
      vi.mocked(post).mockResolvedValue(undefined)

      const ids = [1, 2, 3]
      await stationApi.batchDeleteStations(ids)

      expect(post).toHaveBeenCalledWith('/stations/batch-delete', { ids })
    })
  })

  describe('更新电站状态', () => {
    it('应该成功更新电站状态', async () => {
      vi.mocked(put).mockResolvedValue(undefined)

      await stationApi.updateStationStatus(1, 'offline')

      expect(put).toHaveBeenCalledWith('/stations/1/status', { status: 'offline' })
    })
  })

  describe('获取电站统计信息', () => {
    it('应该成功获取电站统计信息', async () => {
      const mockStats = {
        deviceCount: 10,
        onlineDeviceCount: 8,
        offlineDeviceCount: 2,
        alarmCount: 3,
        power: 100.5,
        energy: 1000.2
      }

      vi.mocked(get).mockResolvedValue(mockStats)

      const result = await stationApi.getStationStatistics(1)

      expect(get).toHaveBeenCalledWith('/stations/1/statistics')
      expect(result).toEqual(mockStats)
    })
  })

  describe('获取电站实时数据', () => {
    it('应该成功获取电站实时数据', async () => {
      const mockData = {
        power: 100.5,
        energy: 1000.2,
        temperature: 25.3
      }

      vi.mocked(get).mockResolvedValue(mockData)

      const result = await stationApi.getStationRealtimeData(1)

      expect(get).toHaveBeenCalledWith('/stations/1/realtime')
      expect(result).toEqual(mockData)
    })
  })

  describe('获取所有电站', () => {
    it('应该成功获取所有电站', async () => {
      const mockStations = [
        { id: 1, name: '电站1' },
        { id: 2, name: '电站2' }
      ]

      vi.mocked(get).mockResolvedValue(mockStations)

      const result = await stationApi.getAllStations()

      expect(get).toHaveBeenCalledWith('/stations/all')
      expect(result).toEqual(mockStations)
    })
  })
})
