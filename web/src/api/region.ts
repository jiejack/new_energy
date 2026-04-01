import { get, post, put, del } from '@/utils/request'
import type { Region, PageQuery, PageResult } from '@/types'

/**
 * 获取区域树
 */
export function getRegionTree(): Promise<Region[]> {
  return get('/regions/tree')
}

/**
 * 获取区域列表
 */
export function getRegionList(params: PageQuery & { keyword?: string; status?: number }): Promise<PageResult<Region>> {
  return get('/regions', params)
}

/**
 * 获取区域详情
 */
export function getRegionDetail(id: number): Promise<Region> {
  return get(`/regions/${id}`)
}

/**
 * 创建区域
 */
export function createRegion(data: Partial<Region>): Promise<Region> {
  return post('/regions', data)
}

/**
 * 更新区域
 */
export function updateRegion(id: number, data: Partial<Region>): Promise<Region> {
  return put(`/regions/${id}`, data)
}

/**
 * 删除区域
 */
export function deleteRegion(id: number): Promise<void> {
  return del(`/regions/${id}`)
}

/**
 * 获取子区域
 */
export function getChildRegions(parentId: number): Promise<Region[]> {
  return get(`/regions/${parentId}/children`)
}

/**
 * 更新区域状态
 */
export function updateRegionStatus(id: number, status: number): Promise<void> {
  return put(`/regions/${id}/status`, { status })
}
