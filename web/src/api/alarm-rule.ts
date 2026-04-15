import { get, post, put, del } from '@/utils/request'
import type { PageQuery, PageResult } from '@/types'

export interface AlarmRule {
  id: string
  name: string
  description: string
  type: string
  level: number
  condition: string
  threshold: number
  duration: number
  point_id?: string
  device_id?: string
  station_id?: string
  notify_channels: string[]
  notify_users: string[]
  status: number
  created_at: string
  updated_at: string
}

export interface CreateAlarmRuleRequest {
  name: string
  description: string
  type: string
  level: number
  condition: string
  threshold: number
  duration: number
  point_id?: string
  device_id?: string
  station_id?: string
  notify_channels: string[]
  notify_users: string[]
}

export interface UpdateAlarmRuleRequest extends Partial<CreateAlarmRuleRequest> {
  id?: string
  status?: number
}

export function getAlarmRuleList(params?: PageQuery & {
  type?: string
  level?: number
  status?: number
  station_id?: string
}): Promise<PageResult<AlarmRule>> {
  return get('/api/v1/alarm-rules', params)
}

export function getAlarmRule(id: string): Promise<AlarmRule> {
  return get(`/api/v1/alarm-rules/${id}`)
}

export function createAlarmRule(data: CreateAlarmRuleRequest): Promise<AlarmRule> {
  return post('/api/v1/alarm-rules', data)
}

export function updateAlarmRule(id: string, data: UpdateAlarmRuleRequest): Promise<AlarmRule> {
  return put(`/api/v1/alarm-rules/${id}`, data)
}

export function deleteAlarmRule(id: string): Promise<void> {
  return del(`/api/v1/alarm-rules/${id}`)
}

export function enableAlarmRule(id: string): Promise<void> {
  return post(`/api/v1/alarm-rules/${id}/enable`)
}

export function disableAlarmRule(id: string): Promise<void> {
  return post(`/api/v1/alarm-rules/${id}/disable`)
}
