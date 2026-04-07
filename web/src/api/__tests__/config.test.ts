import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import * as configApi from '../config'

// Mock request utils
vi.mock('@/utils/request', () => ({
  get: vi.fn(),
  put: vi.fn()
}))

import { get, put } from '@/utils/request'

describe('Config API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('获取所有配置', () => {
    it('应该成功获取所有配置', async () => {
      const mockConfigs = {
        basic: {
          systemName: '新能源监控系统',
          logo: '',
          language: 'zh-CN',
          timezone: 'Asia/Shanghai'
        },
        alarm: {
          defaultLevel: 'warning',
          soundEnabled: true,
          emailEnabled: false,
          smsEnabled: false,
          emailRecipients: '',
          smsRecipients: ''
        },
        display: {
          theme: 'light',
          pageSize: 10,
          refreshInterval: 10,
          dateFormat: 'YYYY-MM-DD HH:mm:ss'
        }
      }

      vi.mocked(get).mockResolvedValue(mockConfigs)

      const result = await configApi.getAllConfigs()

      expect(get).toHaveBeenCalledWith('/v1/configs')
      expect(result).toEqual(mockConfigs)
    })
  })

  describe('获取指定分类的配置', () => {
    it('应该成功获取基本配置', async () => {
      const mockConfig = {
        systemName: '新能源监控系统',
        logo: '',
        language: 'zh-CN',
        timezone: 'Asia/Shanghai'
      }

      vi.mocked(get).mockResolvedValue(mockConfig)

      const result = await configApi.getConfigsByCategory('basic')

      expect(get).toHaveBeenCalledWith('/v1/configs/basic')
      expect(result).toEqual(mockConfig)
    })

    it('应该成功获取告警配置', async () => {
      const mockConfig = {
        defaultLevel: 'warning',
        soundEnabled: true,
        emailEnabled: false,
        smsEnabled: false,
        emailRecipients: '',
        smsRecipients: ''
      }

      vi.mocked(get).mockResolvedValue(mockConfig)

      const result = await configApi.getConfigsByCategory('alarm')

      expect(get).toHaveBeenCalledWith('/v1/configs/alarm')
      expect(result).toEqual(mockConfig)
    })

    it('应该成功获取显示配置', async () => {
      const mockConfig = {
        theme: 'light',
        pageSize: 10,
        refreshInterval: 10,
        dateFormat: 'YYYY-MM-DD HH:mm:ss'
      }

      vi.mocked(get).mockResolvedValue(mockConfig)

      const result = await configApi.getConfigsByCategory('display')

      expect(get).toHaveBeenCalledWith('/v1/configs/display')
      expect(result).toEqual(mockConfig)
    })
  })

  describe('更新配置', () => {
    it('应该成功更新单个配置项', async () => {
      vi.mocked(put).mockResolvedValue(undefined)

      await configApi.updateConfig('basic', 'systemName', '新系统名称')

      expect(put).toHaveBeenCalledWith('/v1/configs/basic/systemName', { value: '新系统名称' })
    })

    it('应该成功更新告警配置项', async () => {
      vi.mocked(put).mockResolvedValue(undefined)

      await configApi.updateConfig('alarm', 'soundEnabled', false)

      expect(put).toHaveBeenCalledWith('/v1/configs/alarm/soundEnabled', { value: false })
    })

    it('应该成功更新显示配置项', async () => {
      vi.mocked(put).mockResolvedValue(undefined)

      await configApi.updateConfig('display', 'theme', 'dark')

      expect(put).toHaveBeenCalledWith('/v1/configs/display/theme', { value: 'dark' })
    })
  })

  describe('批量更新配置', () => {
    it('应该成功批量更新基本配置', async () => {
      vi.mocked(put).mockResolvedValue(undefined)

      const configs = {
        systemName: '新系统名称',
        language: 'en-US'
      }

      await configApi.batchUpdateConfig('basic', configs)

      expect(put).toHaveBeenCalledWith('/v1/configs/basic', configs)
    })

    it('应该成功批量更新告警配置', async () => {
      vi.mocked(put).mockResolvedValue(undefined)

      const configs = {
        soundEnabled: false,
        emailEnabled: true,
        emailRecipients: 'admin@example.com'
      }

      await configApi.batchUpdateConfig('alarm', configs)

      expect(put).toHaveBeenCalledWith('/v1/configs/alarm', configs)
    })

    it('应该成功批量更新显示配置', async () => {
      vi.mocked(put).mockResolvedValue(undefined)

      const configs = {
        theme: 'dark',
        pageSize: 20,
        refreshInterval: 5
      }

      await configApi.batchUpdateConfig('display', configs)

      expect(put).toHaveBeenCalledWith('/v1/configs/display', configs)
    })
  })
})
