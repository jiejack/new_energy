import { get, post, put } from '@/utils/request'
import type { Alarm, PageQuery, PageResult, AlarmLevel, AlarmStatus } from '@/types'

/**
 * 获取告警列表
 */
export function getAlarmList(params: PageQuery & {
  keyword?: string
  level?: AlarmLevel
  status?: AlarmStatus
  sourceType?: string
  sourceId?: number
  startTime?: string
  endTime?: string
}): Promise<PageResult<Alarm>> {
  return get('/alarms', params)
}

/**
 * 获取告警详情
 */
export function getAlarmDetail(id: number): Promise<Alarm> {
  return get(`/alarms/${id}`)
}

/**
 * 确认告警
 */
export function acknowledgeAlarm(id: number): Promise<void> {
  return put(`/alarms/${id}/acknowledge`)
}

/**
 * 批量确认告警
 */
export function batchAcknowledgeAlarms(ids: number[]): Promise<void> {
  return post('/alarms/batch-acknowledge', { ids })
}

/**
 * 解决告警
 */
export function resolveAlarm(id: number): Promise<void> {
  return put(`/alarms/${id}/resolve`)
}

/**
 * 批量解决告警
 */
export function batchResolveAlarms(ids: number[]): Promise<void> {
  return post('/alarms/batch-resolve', { ids })
}

/**
 * 获取当前告警统计
 */
export function getAlarmStatistics(): Promise<{
  total: number
  critical: number
  major: number
  minor: number
  warning: number
}> {
  return get('/alarms/statistics')
}

/**
 * 获取告警趋势
 */
export function getAlarmTrend(params: {
  startTime: string
  endTime: string
  interval?: 'hour' | 'day' | 'month'
}): Promise<Array<{
  time: string
  count: number
}>> {
  return get('/alarms/trend', params)
}

/**
 * 获取告警分布
 */
export function getAlarmDistribution(params: {
  startTime: string
  endTime: string
  groupBy: 'level' | 'source' | 'type'
}): Promise<Array<{
  name: string
  count: number
}>> {
  return get('/alarms/distribution', params)
}
