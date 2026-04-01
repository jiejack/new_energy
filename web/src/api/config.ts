import { get, put } from '@/utils/request'

/**
 * 配置项类型定义
 */
export interface ConfigItem {
  category: string
  key: string
  value: string
  description?: string
  createdAt?: string
  updatedAt?: string
}

/**
 * 配置分类
 */
export type ConfigCategory = 'basic' | 'alarm' | 'display'

/**
 * 基本设置
 */
export interface BasicConfig {
  systemName: string
  logo: string
  language: string
  timezone: string
}

/**
 * 告警设置
 */
export interface AlarmConfig {
  defaultLevel: string
  soundEnabled: boolean
  emailEnabled: boolean
  smsEnabled: boolean
  emailRecipients: string
  smsRecipients: string
}

/**
 * 显示设置
 */
export interface DisplayConfig {
  theme: string
  pageSize: number
  refreshInterval: number
  dateFormat: string
}

/**
 * 所有配置
 */
export interface AllConfigs {
  basic: BasicConfig
  alarm: AlarmConfig
  display: DisplayConfig
}

/**
 * 获取所有配置
 */
export function getAllConfigs(): Promise<AllConfigs> {
  return get('/v1/configs')
}

/**
 * 获取指定分类的配置
 */
export function getConfigsByCategory(category: ConfigCategory): Promise<Record<string, any>> {
  return get(`/v1/configs/${category}`)
}

/**
 * 更新配置
 */
export function updateConfig(category: string, key: string, value: any): Promise<void> {
  return put(`/v1/configs/${category}/${key}`, { value })
}

/**
 * 批量更新配置
 */
export function batchUpdateConfig(category: string, configs: Record<string, any>): Promise<void> {
  return put(`/v1/configs/${category}`, configs)
}
