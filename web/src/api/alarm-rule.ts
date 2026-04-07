import { get, post, del } from '@/utils/request'
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

export interface UpdateAlarmRuleRequest extends CreateAlarmRuleRequest {
  id: string
}

export function getAlarmRuleList(params?: PageQuery & {
  type?: string
  level?: number
  status?: number
}): Promise<PageResult<AlarmRule>> {
  return get('/alarm-rules', params)
}

export function getAlarmRule(id: string): Promise<AlarmRule> {
  return get(`/alarm-rules/${id}`)
}

export function createAlarmRule(data: CreateAlarmRuleRequest): Promise<AlarmRule> {
  return post('/alarm-rules', data)
}

export function updateAlarmRule(id: string, data: UpdateAlarmRuleRequest): Promise<AlarmRule> {
  return post(`/alarm-rules/${id}`, { ...data, _method: 'PUT' })
}

export function deleteAlarmRule(id: string): Promise<void> {
  return del(`/alarm-rules/${id}`)
}
