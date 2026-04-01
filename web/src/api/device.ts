import { get, post, put, del } from '@/utils/request'
import type { Device, PageQuery, PageResult, DeviceType, DeviceStatus } from '@/types'

/**
 * 获取设备列表
 */
export function getDeviceList(params: PageQuery & {
  keyword?: string
  stationId?: number
  type?: DeviceType
  status?: DeviceStatus
}): Promise<PageResult<Device>> {
  return get('/devices', params)
}

/**
 * 获取设备详情
 */
export function getDeviceDetail(id: number): Promise<Device> {
  return get(`/devices/${id}`)
}

/**
 * 创建设备
 */
export function createDevice(data: Partial<Device>): Promise<Device> {
  return post('/devices', data)
}

/**
 * 更新设备
 */
export function updateDevice(id: number, data: Partial<Device>): Promise<Device> {
  return put(`/devices/${id}`, data)
}

/**
 * 删除设备
 */
export function deleteDevice(id: number): Promise<void> {
  return del(`/devices/${id}`)
}

/**
 * 批量删除设备
 */
export function batchDeleteDevices(ids: number[]): Promise<void> {
  return post('/devices/batch-delete', { ids })
}

/**
 * 更新设备状态
 */
export function updateDeviceStatus(id: number, status: DeviceStatus): Promise<void> {
  return put(`/devices/${id}/status`, { status })
}

/**
 * 获取设备采集点
 */
export function getDevicePoints(deviceId: number): Promise<any[]> {
  return get(`/devices/${deviceId}/points`)
}

/**
 * 获取设备实时数据
 */
export function getDeviceRealtimeData(id: number): Promise<Record<string, any>> {
  return get(`/devices/${id}/realtime`)
}

/**
 * 获取设备历史数据
 */
export function getDeviceHistoryData(id: number, params: {
  startTime: string
  endTime: string
  pointIds?: number[]
}): Promise<any[]> {
  return get(`/devices/${id}/history`, params)
}

/**
 * 获取所有设备（下拉选择用）
 */
export function getAllDevices(stationId?: number): Promise<Device[]> {
  return get('/devices/all', { stationId })
}
