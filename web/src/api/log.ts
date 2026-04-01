import { get, post } from '@/utils/request'
import type { OperationLog, PageQuery, PageResult } from '@/types'

/**
 * 获取操作日志列表
 */
export function getLogList(params: PageQuery & {
  userId?: number
  username?: string
  operation?: string
  status?: number
  startTime?: string
  endTime?: string
}): Promise<PageResult<OperationLog>> {
  return get('/logs', params)
}

/**
 * 获取日志详情
 */
export function getLogDetail(id: number): Promise<OperationLog> {
  return get(`/logs/${id}`)
}

/**
 * 删除日志
 */
export function deleteLog(id: number): Promise<void> {
  return post(`/logs/${id}/delete`)
}

/**
 * 批量删除日志
 */
export function batchDeleteLogs(ids: number[]): Promise<void> {
  return post('/logs/batch-delete', { ids })
}

/**
 * 清空日志
 */
export function clearLogs(beforeDate?: string): Promise<void> {
  return post('/logs/clear', { beforeDate })
}

/**
 * 导出日志
 */
export function exportLogs(params: {
  userId?: number
  username?: string
  operation?: string
  status?: number
  startTime?: string
  endTime?: string
}): Promise<void> {
  return post('/logs/export', params)
}
