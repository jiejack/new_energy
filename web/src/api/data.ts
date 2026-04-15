import { get, post } from '@/utils/request'
import type { DataQuery, PointData, PageQuery, PageResult } from '@/types'

/**
 * 查询历史数据
 */
export function queryHistoryData(params: DataQuery): Promise<PointData[]> {
  return post('/api/v1/data/query', params)
}

/**
 * 查询实时数据
 */
export function queryRealtimeData(pointIds: number[]): Promise<Record<number, {
  value: number
  quality: number
  timestamp: number
}>> {
  return post('/api/v1/data/realtime', { pointIds })
}

/**
 * 获取实时数据列表（分页）
 */
export function getRealtimeData(params: PageQuery): Promise<PageResult<{
  stationName: string
  deviceName: string
  pointName: string
  value: number
  unit: string
  updateTime: string
  status: string
}>> {
  return get('/api/v1/data/realtime-list', params)
}

/**
 * 获取统计数据
 */
export function getStatistics(params: {
  pointIds: number[]
  startTime: string
  endTime: string
  interval: 'hour' | 'day' | 'month' | 'year'
  aggregations: Array<'avg' | 'max' | 'min' | 'sum' | 'count'>
}): Promise<Array<{
  timestamp: string
  values: Record<string, number>
}>> {
  return post('/api/v1/data/statistics', params)
}

/**
 * 导出数据
 */
export function exportData(params: DataQuery & { format: 'csv' | 'excel' }): Promise<Blob> {
  return post('/api/v1/data/export', params, { responseType: 'blob' })
}

/**
 * 获取电站发电量统计
 */
export function getStationPowerStatistics(params: {
  stationId: number
  startTime: string
  endTime: string
  interval: 'hour' | 'day' | 'month' | 'year'
}): Promise<Array<{
  timestamp: string
  power: number
  energy: number
}>> {
  return get(`/api/v1/stations/${params.stationId}/power-statistics`, params)
}

/**
 * 获取设备运行统计
 */
export function getDeviceOperationStatistics(params: {
  deviceId: number
  startTime: string
  endTime: string
}): Promise<{
  onlineTime: number
  offlineTime: number
  maintenanceTime: number
  faultTime: number
  availability: number
}> {
  return get(`/api/v1/devices/${params.deviceId}/operation-statistics`, params)
}

/**
 * 获取对比数据
 */
export function getComparisonData(params: {
  pointIds: number[]
  startTime: string
  endTime: string
  compareType: 'previous' | 'year' | 'custom'
  compareStartTime?: string
  compareEndTime?: string
}): Promise<{
  current: PointData[]
  compare: PointData[]
}> {
  return post('/api/v1/data/comparison', params)
}

/**
 * 获取聚合数据
 */
export function getAggregatedData(params: {
  pointId: number
  startTime: string
  endTime: string
  aggregation: 'avg' | 'max' | 'min' | 'sum'
  groupBy: 'hour' | 'day' | 'month'
}): Promise<Array<{
  time: string
  value: number
}>> {
  return get('/api/v1/data/aggregated', params)
}
