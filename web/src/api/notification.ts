import { get, post, put } from '@/utils/request'

export type NotificationType = 'email' | 'sms' | 'webhook' | 'wechat'

export interface NotificationConfig {
  id: string
  type: NotificationType
  name: string
  config: Record<string, any>
  enabled: boolean
}

export interface EmailConfig {
  smtp_host: string
  smtp_port: number
  username: string
  password?: string
  from: string
  use_tls: boolean
}

export interface SmsConfig {
  access_key: string
  secret_key: string
  sign_name: string
  template_id?: string
}

export interface WebhookConfig {
  url: string
  method: 'GET' | 'POST'
  headers?: Record<string, string>
  timeout?: number
}

export interface WechatConfig {
  corp_id: string
  corp_secret: string
  agent_id: string
}

export function getNotificationConfigs(): Promise<NotificationConfig[]> {
  return get('/api/v1/notification-configs')
}

export function getNotificationConfig(type: NotificationType): Promise<NotificationConfig> {
  return get(`/api/v1/notification-configs/${type}`)
}

export function updateNotificationConfig(type: NotificationType, config: Record<string, any>): Promise<NotificationConfig> {
  return put(`/api/v1/notification-configs/${type}`, config)
}

export function enableNotificationConfig(type: NotificationType): Promise<void> {
  return post(`/api/v1/notification-configs/${type}/enable`)
}

export function disableNotificationConfig(type: NotificationType): Promise<void> {
  return post(`/api/v1/notification-configs/${type}/disable`)
}

export function testNotificationConfig(type: NotificationType, testTarget?: string): Promise<{ success: boolean; message: string }> {
  return post(`/api/v1/notification-configs/${type}/test`, { test_target: testTarget })
}
