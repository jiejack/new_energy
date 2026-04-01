import { get, post, put, del } from '@/utils/request'
import type { Point, PageQuery, PageResult, PointType, DataType } from '@/types'

/**
 * 获取采集点列表
 */
export function getPointList(params: PageQuery & {
  keyword?: string
  deviceId?: number
  type?: PointType
  dataType?: DataType
}): Promise<PageResult<Point>> {
  return get('/points', params)
}

/**
 * 获取采集点详情
 */
export function getPointDetail(id: number): Promise<Point> {
  return get(`/points/${id}`)
}

/**
 * 创建采集点
 */
export function createPoint(data: Partial<Point>): Promise<Point> {
  return post('/points', data)
}

/**
 * 更新采集点
 */
export function updatePoint(id: number, data: Partial<Point>): Promise<Point> {
  return put(`/points/${id}`, data)
}

/**
 * 删除采集点
 */
export function deletePoint(id: number): Promise<void> {
  return del(`/points/${id}`)
}

/**
 * 批量删除采集点
 */
export function batchDeletePoints(ids: number[]): Promise<void> {
  return post('/points/batch-delete', { ids })
}

/**
 * 获取采集点实时数据
 */
export function getPointRealtimeData(id: number): Promise<{
  value: number
  quality: number
  timestamp: number
}> {
  return get(`/points/${id}/realtime`)
}

/**
 * 获取采集点历史数据
 */
export function getPointHistoryData(id: number, params: {
  startTime: string
  endTime: string
  interval?: number
  aggregation?: 'avg' | 'max' | 'min' | 'sum'
}): Promise<any[]> {
  return get(`/points/${id}/history`, params)
}

/**
 * 获取设备下所有采集点
 */
export function getPointsByDevice(deviceId: number): Promise<Point[]> {
  return get(`/devices/${deviceId}/points`)
}

/**
 * 批量获取采集点实时数据
 */
export function getBatchRealtimeData(pointIds: number[]): Promise<Record<number, {
  value: number
  quality: number
  timestamp: number
}>> {
  return post('/points/batch-realtime', { pointIds })
}
