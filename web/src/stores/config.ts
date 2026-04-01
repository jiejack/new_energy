import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getAllConfigs, updateConfig, batchUpdateConfig } from '@/api/config'
import type { BasicConfig, AlarmConfig, DisplayConfig } from '@/api/config'

export const useConfigStore = defineStore('config', () => {
  // 状态
  const loading = ref(false)
  const basicConfig = ref<BasicConfig>({
    systemName: '新能源监控系统',
    logo: '',
    language: 'zh-CN',
    timezone: 'Asia/Shanghai',
  })

  const alarmConfig = ref<AlarmConfig>({
    defaultLevel: 'warning',
    soundEnabled: true,
    emailEnabled: false,
    smsEnabled: false,
    emailRecipients: '',
    smsRecipients: '',
  })

  const displayConfig = ref<DisplayConfig>({
    theme: 'light',
    pageSize: 10,
    refreshInterval: 10,
    dateFormat: 'YYYY-MM-DD HH:mm:ss',
  })

  /**
   * 加载所有配置
   */
  async function loadConfigs() {
    loading.value = true
    try {
      const configs = await getAllConfigs()
      if (configs.basic) {
        basicConfig.value = { ...basicConfig.value, ...configs.basic }
      }
      if (configs.alarm) {
        alarmConfig.value = { ...alarmConfig.value, ...configs.alarm }
      }
      if (configs.display) {
        displayConfig.value = { ...displayConfig.value, ...configs.display }
      }
      return configs
    } catch (error) {
      console.error('加载配置失败:', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  /**
   * 更新基本配置
   */
  async function updateBasicConfig(config: Partial<BasicConfig>) {
    try {
      await batchUpdateConfig('basic', config)
      Object.assign(basicConfig.value, config)
    } catch (error) {
      console.error('更新基本配置失败:', error)
      throw error
    }
  }

  /**
   * 更新告警配置
   */
  async function updateAlarmConfig(config: Partial<AlarmConfig>) {
    try {
      await batchUpdateConfig('alarm', config)
      Object.assign(alarmConfig.value, config)
    } catch (error) {
      console.error('更新告警配置失败:', error)
      throw error
    }
  }

  /**
   * 更新显示配置
   */
  async function updateDisplayConfig(config: Partial<DisplayConfig>) {
    try {
      await batchUpdateConfig('display', config)
      Object.assign(displayConfig.value, config)
      // 应用主题
      if (config.theme) {
        applyTheme(config.theme)
      }
    } catch (error) {
      console.error('更新显示配置失败:', error)
      throw error
    }
  }

  /**
   * 更新单个配置项
   */
  async function updateSingleConfig(category: string, key: string, value: any) {
    try {
      await updateConfig(category, key, value)
      // 更新本地状态
      if (category === 'basic' && key in basicConfig.value) {
        (basicConfig.value as any)[key] = value
      } else if (category === 'alarm' && key in alarmConfig.value) {
        (alarmConfig.value as any)[key] = value
      } else if (category === 'display' && key in displayConfig.value) {
        (displayConfig.value as any)[key] = value
      }
    } catch (error) {
      console.error('更新配置失败:', error)
      throw error
    }
  }

  /**
   * 应用主题
   */
  function applyTheme(theme: string) {
    if (theme === 'dark') {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }

  /**
   * 重置配置
   */
  function resetConfigs() {
    basicConfig.value = {
      systemName: '新能源监控系统',
      logo: '',
      language: 'zh-CN',
      timezone: 'Asia/Shanghai',
    }
    alarmConfig.value = {
      defaultLevel: 'warning',
      soundEnabled: true,
      emailEnabled: false,
      smsEnabled: false,
      emailRecipients: '',
      smsRecipients: '',
    }
    displayConfig.value = {
      theme: 'light',
      pageSize: 10,
      refreshInterval: 10,
      dateFormat: 'YYYY-MM-DD HH:mm:ss',
    }
  }

  return {
    // 状态
    loading,
    basicConfig,
    alarmConfig,
    displayConfig,
    // 方法
    loadConfigs,
    updateBasicConfig,
    updateAlarmConfig,
    updateDisplayConfig,
    updateSingleConfig,
    resetConfigs,
  }
})
