import { get, post, put, del } from '@/utils/request'
import type { Station, PageQuery, PageResult, StationType, StationStatus } from '@/types'

/**
 * 获取电站列表
 */
export function getStationList(params: PageQuery & {
  keyword?: string
  regionId?: number
  type?: StationType
  status?: StationStatus
}): Promise<PageResult<Station>> {
  return get('/stations', params)
}

/**
 * 获取电站详情
 */
export function getStationDetail(id: number): Promise<Station> {
  return get(`/stations/${id}`)
}

/**
 * 创建电站
 */
export function createStation(data: Partial<Station>): Promise<Station> {
  return post('/stations', data)
}

/**
 * 更新电站
 */
export function updateStation(id: number, data: Partial<Station>): Promise<Station> {
  return put(`/stations/${id}`, data)
}

/**
 * 删除电站
 */
export function deleteStation(id: number): Promise<void> {
  return del(`/stations/${id}`)
}

/**
 * 批量删除电站
 */
export function batchDeleteStations(ids: number[]): Promise<void> {
  return post('/stations/batch-delete', { ids })
}

/**
 * 更新电站状态
 */
export function updateStationStatus(id: number, status: StationStatus): Promise<void> {
  return put(`/stations/${id}/status`, { status })
}

/**
 * 获取电站统计信息
 */
export function getStationStatistics(id: number): Promise<{
  deviceCount: number
  onlineDeviceCount: number
  offlineDeviceCount: number
  alarmCount: number
  power: number
  energy: number
}> {
  return get(`/stations/${id}/statistics`)
}

/**
 * 获取电站实时数据
 */
export function getStationRealtimeData(id: number): Promise<Record<string, any>> {
  return get(`/stations/${id}/realtime`)
}

/**
 * 获取所有电站（下拉选择用）
 */
export function getAllStations(): Promise<Station[]> {
  return get('/stations/all')
}
