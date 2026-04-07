import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useConfigStore } from '../config'
import * as configApi from '@/api/config'
import type { AllConfigs } from '@/api/config'

// Mock API
vi.mock('@/api/config', () => ({
  getAllConfigs: vi.fn(),
  updateConfig: vi.fn(),
  batchUpdateConfig: vi.fn()
}))

describe('Config Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    // 重置 document.documentElement
    document.documentElement.classList.remove('dark')
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('状态初始化', () => {
    it('应该初始化默认配置', () => {
      const store = useConfigStore()

      expect(store.loading).toBe(false)
      expect(store.basicConfig).toEqual({
        systemName: '新能源监控系统',
        logo: '',
        language: 'zh-CN',
        timezone: 'Asia/Shanghai'
      })
      expect(store.alarmConfig).toEqual({
        defaultLevel: 'warning',
        soundEnabled: true,
        emailEnabled: false,
        smsEnabled: false,
        emailRecipients: '',
        smsRecipients: ''
      })
      expect(store.displayConfig).toEqual({
        theme: 'light',
        pageSize: 10,
        refreshInterval: 10,
        dateFormat: 'YYYY-MM-DD HH:mm:ss'
      })
    })
  })

  describe('loadConfigs', () => {
    it('应该成功加载所有配置', async () => {
      const mockConfigs = {
        basic: {
          systemName: '测试系统',
          logo: 'logo.png',
          language: 'en-US',
          timezone: 'UTC'
        },
        alarm: {
          defaultLevel: 'critical',
          soundEnabled: false,
          emailEnabled: true,
          smsEnabled: false,
          emailRecipients: 'admin@example.com',
          smsRecipients: ''
        },
        display: {
          theme: 'dark',
          pageSize: 20,
          refreshInterval: 5,
          dateFormat: 'YYYY-MM-DD'
        }
      }

      vi.mocked(configApi.getAllConfigs).mockResolvedValue(mockConfigs)

      const store = useConfigStore()
      const result = await store.loadConfigs()

      expect(configApi.getAllConfigs).toHaveBeenCalled()
      expect(store.basicConfig).toEqual(mockConfigs.basic)
      expect(store.alarmConfig).toEqual(mockConfigs.alarm)
      expect(store.displayConfig).toEqual(mockConfigs.display)
      expect(result).toEqual(mockConfigs)
    })

    it('加载配置时应该设置loading状态', async () => {
      const emptyConfigs: AllConfigs = {
        basic: { systemName: '', logo: '', language: '', timezone: '' },
        alarm: { defaultLevel: '', soundEnabled: false, emailEnabled: false, smsEnabled: false, emailRecipients: '', smsRecipients: '' },
        display: { theme: '', pageSize: 10, refreshInterval: 10, dateFormat: '' }
      }
      vi.mocked(configApi.getAllConfigs).mockResolvedValue(emptyConfigs)

      const store = useConfigStore()
      const promise = store.loadConfigs()

      expect(store.loading).toBe(true)

      await promise

      expect(store.loading).toBe(false)
    })

    it('加载配置失败应该抛出错误', async () => {
      const mockError = new Error('加载配置失败')
      vi.mocked(configApi.getAllConfigs).mockRejectedValue(mockError)

      const store = useConfigStore()

      await expect(store.loadConfigs()).rejects.toThrow('加载配置失败')
      expect(store.loading).toBe(false)
    })
  })

  describe('updateBasicConfig', () => {
    it('应该成功更新基本配置', async () => {
      vi.mocked(configApi.batchUpdateConfig).mockResolvedValue()

      const store = useConfigStore()
      const config = {
        systemName: '新系统名称',
        language: 'en-US'
      }

      await store.updateBasicConfig(config)

      expect(configApi.batchUpdateConfig).toHaveBeenCalledWith('basic', config)
      expect(store.basicConfig.systemName).toBe('新系统名称')
      expect(store.basicConfig.language).toBe('en-US')
    })

    it('更新失败应该抛出错误', async () => {
      const mockError = new Error('更新失败')
      vi.mocked(configApi.batchUpdateConfig).mockRejectedValue(mockError)

      const store = useConfigStore()

      await expect(store.updateBasicConfig({ systemName: 'test' })).rejects.toThrow('更新失败')
    })
  })

  describe('updateAlarmConfig', () => {
    it('应该成功更新告警配置', async () => {
      vi.mocked(configApi.batchUpdateConfig).mockResolvedValue()

      const store = useConfigStore()
      const config = {
        soundEnabled: false,
        emailEnabled: true
      }

      await store.updateAlarmConfig(config)

      expect(configApi.batchUpdateConfig).toHaveBeenCalledWith('alarm', config)
      expect(store.alarmConfig.soundEnabled).toBe(false)
      expect(store.alarmConfig.emailEnabled).toBe(true)
    })
  })

  describe('updateDisplayConfig', () => {
    it('应该成功更新显示配置', async () => {
      vi.mocked(configApi.batchUpdateConfig).mockResolvedValue()

      const store = useConfigStore()
      const config = {
        theme: 'dark',
        pageSize: 20
      }

      await store.updateDisplayConfig(config)

      expect(configApi.batchUpdateConfig).toHaveBeenCalledWith('display', config)
      expect(store.displayConfig.theme).toBe('dark')
      expect(store.displayConfig.pageSize).toBe(20)
    })

    it('更新主题时应该应用主题', async () => {
      vi.mocked(configApi.batchUpdateConfig).mockResolvedValue()

      const store = useConfigStore()

      await store.updateDisplayConfig({ theme: 'dark' })

      expect(document.documentElement.classList.contains('dark')).toBe(true)

      await store.updateDisplayConfig({ theme: 'light' })

      expect(document.documentElement.classList.contains('dark')).toBe(false)
    })
  })

  describe('updateSingleConfig', () => {
    it('应该成功更新单个配置项', async () => {
      vi.mocked(configApi.updateConfig).mockResolvedValue()

      const store = useConfigStore()

      await store.updateSingleConfig('basic', 'systemName', '新名称')

      expect(configApi.updateConfig).toHaveBeenCalledWith('basic', 'systemName', '新名称')
      expect(store.basicConfig.systemName).toBe('新名称')
    })

    it('应该更新告警配置项', async () => {
      vi.mocked(configApi.updateConfig).mockResolvedValue()

      const store = useConfigStore()

      await store.updateSingleConfig('alarm', 'soundEnabled', false)

      expect(configApi.updateConfig).toHaveBeenCalledWith('alarm', 'soundEnabled', false)
      expect(store.alarmConfig.soundEnabled).toBe(false)
    })

    it('应该更新显示配置项', async () => {
      vi.mocked(configApi.updateConfig).mockResolvedValue()

      const store = useConfigStore()

      await store.updateSingleConfig('display', 'pageSize', 20)

      expect(configApi.updateConfig).toHaveBeenCalledWith('display', 'pageSize', 20)
      expect(store.displayConfig.pageSize).toBe(20)
    })
  })

  describe('resetConfigs', () => {
    it('应该重置所有配置到默认值', () => {
      const store = useConfigStore()

      // 修改配置
      store.basicConfig.systemName = '测试系统'
      store.alarmConfig.soundEnabled = false
      store.displayConfig.theme = 'dark'

      // 重置
      store.resetConfigs()

      expect(store.basicConfig).toEqual({
        systemName: '新能源监控系统',
        logo: '',
        language: 'zh-CN',
        timezone: 'Asia/Shanghai'
      })
      expect(store.alarmConfig).toEqual({
        defaultLevel: 'warning',
        soundEnabled: true,
        emailEnabled: false,
        smsEnabled: false,
        emailRecipients: '',
        smsRecipients: ''
      })
      expect(store.displayConfig).toEqual({
        theme: 'light',
        pageSize: 10,
        refreshInterval: 10,
        dateFormat: 'YYYY-MM-DD HH:mm:ss'
      })
    })
  })
})
