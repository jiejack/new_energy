import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import * as deviceApi from '../device'

// Mock request utils
vi.mock('@/utils/request', () => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  del: vi.fn()
}))

import { get, post, put, del } from '@/utils/request'

describe('Device API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('获取设备列表', () => {
    it('应该成功获取设备列表', async () => {
      const mockResponse = {
        list: [
          { id: 1, name: '设备1' },
          { id: 2, name: '设备2' }
        ],
        total: 2,
        page: 1,
        pageSize: 10
      }

      vi.mocked(get).mockResolvedValue(mockResponse)

      const params = { page: 1, pageSize: 10 }
      const result = await deviceApi.getDeviceList(params)

      expect(get).toHaveBeenCalledWith('/devices', params)
      expect(result).toEqual(mockResponse)
    })

    it('应该支持多条件筛选', async () => {
      const mockResponse = {
        list: [{ id: 1, name: '设备1' }],
        total: 1,
        page: 1,
        pageSize: 10
      }

      vi.mocked(get).mockResolvedValue(mockResponse)

      const params = {
        page: 1,
        pageSize: 10,
        keyword: '设备',
        stationId: 1,
        type: 'inverter' as const,
        status: 'online' as const
      }
      await deviceApi.getDeviceList(params)

      expect(get).toHaveBeenCalledWith('/devices', params)
    })
  })

  describe('获取设备详情', () => {
    it('应该成功获取设备详情', async () => {
      const mockDevice = {
        id: 1,
        name: '测试设备',
        type: 'inverter',
        status: 'online'
      }

      vi.mocked(get).mockResolvedValue(mockDevice)

      const result = await deviceApi.getDeviceDetail(1)

      expect(get).toHaveBeenCalledWith('/devices/1')
      expect(result).toEqual(mockDevice)
    })
  })

  describe('创建设备', () => {
    it('应该成功创建设备', async () => {
      const mockDevice = {
        id: 1,
        name: '新设备',
        type: 'inverter'
      }

      vi.mocked(post).mockResolvedValue(mockDevice)

      const data = {
        name: '新设备',
        type: 'inverter' as const,
        stationId: 1
      }

      const result = await deviceApi.createDevice(data)

      expect(post).toHaveBeenCalledWith('/devices', data)
      expect(result).toEqual(mockDevice)
    })
  })

  describe('更新设备', () => {
    it('应该成功更新设备', async () => {
      const mockDevice = {
        id: 1,
        name: '更新设备',
        type: 'inverter'
      }

      vi.mocked(put).mockResolvedValue(mockDevice)

      const data = { name: '更新设备' }
      const result = await deviceApi.updateDevice(1, data)

      expect(put).toHaveBeenCalledWith('/devices/1', data)
      expect(result).toEqual(mockDevice)
    })
  })

  describe('删除设备', () => {
    it('应该成功删除设备', async () => {
      vi.mocked(del).mockResolvedValue(undefined)

      await deviceApi.deleteDevice(1)

      expect(del).toHaveBeenCalledWith('/devices/1')
    })
  })

  describe('批量删除设备', () => {
    it('应该成功批量删除设备', async () => {
      vi.mocked(post).mockResolvedValue(undefined)

      const ids = [1, 2, 3]
      await deviceApi.batchDeleteDevices(ids)

      expect(post).toHaveBeenCalledWith('/devices/batch-delete', { ids })
    })
  })

  describe('更新设备状态', () => {
    it('应该成功更新设备状态', async () => {
      vi.mocked(put).mockResolvedValue(undefined)

      await deviceApi.updateDeviceStatus(1, 'offline')

      expect(put).toHaveBeenCalledWith('/devices/1/status', { status: 'offline' })
    })
  })

  describe('获取设备采集点', () => {
    it('应该成功获取设备采集点', async () => {
      const mockPoints = [
        { id: 1, name: '温度' },
        { id: 2, name: '电压' }
      ]

      vi.mocked(get).mockResolvedValue(mockPoints)

      const result = await deviceApi.getDevicePoints(1)

      expect(get).toHaveBeenCalledWith('/devices/1/points')
      expect(result).toEqual(mockPoints)
    })
  })

  describe('获取设备实时数据', () => {
    it('应该成功获取设备实时数据', async () => {
      const mockData = {
        temperature: 25.5,
        voltage: 220,
        current: 10
      }

      vi.mocked(get).mockResolvedValue(mockData)

      const result = await deviceApi.getDeviceRealtimeData(1)

      expect(get).toHaveBeenCalledWith('/devices/1/realtime')
      expect(result).toEqual(mockData)
    })
  })

  describe('获取设备历史数据', () => {
    it('应该成功获取设备历史数据', async () => {
      const mockData = [
        { timestamp: '2024-01-01 00:00:00', value: 25.5 },
        { timestamp: '2024-01-01 01:00:00', value: 26.0 }
      ]

      vi.mocked(get).mockResolvedValue(mockData)

      const params = {
        startTime: '2024-01-01',
        endTime: '2024-01-02',
        pointIds: [1, 2]
      }
      const result = await deviceApi.getDeviceHistoryData(1, params)

      expect(get).toHaveBeenCalledWith('/devices/1/history', params)
      expect(result).toEqual(mockData)
    })
  })

  describe('获取所有设备', () => {
    it('应该成功获取所有设备', async () => {
      const mockDevices = [
        { id: 1, name: '设备1' },
        { id: 2, name: '设备2' }
      ]

      vi.mocked(get).mockResolvedValue(mockDevices)

      const result = await deviceApi.getAllDevices()

      expect(get).toHaveBeenCalledWith('/devices/all', {})
      expect(result).toEqual(mockDevices)
    })

    it('应该支持按电站筛选', async () => {
      const mockDevices = [
        { id: 1, name: '设备1', stationId: 1 }
      ]

      vi.mocked(get).mockResolvedValue(mockDevices)

      const result = await deviceApi.getAllDevices(1)

      expect(get).toHaveBeenCalledWith('/devices/all', { stationId: 1 })
      expect(result).toEqual(mockDevices)
    })
  })
})
