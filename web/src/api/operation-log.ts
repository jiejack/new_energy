import { get, post, del } from '@/utils/request'
import type { PageQuery, PageResult } from '@/types'

export interface OperationLog {
  id: string
  user_id: string
  username: string
  method: string
  path: string
  action: string
  resource: string
  resource_id: string
  request_ip: string
  status: number
  duration: number
  created_at: string
}

export interface OperationLogQuery extends PageQuery {
  user_id?: string
  username?: string
  action?: string
  resource?: string
  start_time?: string
  end_time?: string
  status?: number
}

export function getOperationLogs(params: OperationLogQuery): Promise<PageResult<OperationLog>> {
  return get('/api/v1/operation-logs', params)
}

export function getOperationLog(id: string): Promise<OperationLog> {
  return get(`/api/v1/operation-logs/${id}`)
}

export function deleteOperationLog(id: string): Promise<void> {
  return del(`/api/v1/operation-logs/${id}`)
}

export function batchDeleteOperationLogs(ids: string[]): Promise<void> {
  return post('/api/v1/operation-logs/batch-delete', { ids })
}

export function clearOperationLogs(beforeDate?: string): Promise<void> {
  return post('/api/v1/operation-logs/clear', { before_date: beforeDate })
}

export function exportOperationLogs(params: OperationLogQuery): Promise<Blob> {
  return get('/api/v1/operation-logs/export', params, { responseType: 'blob' })
}
