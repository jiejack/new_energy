import { get, post, put, del } from '@/utils/request'
import type { Permission } from '@/types'

/**
 * 获取权限树
 */
export function getPermissionTree(): Promise<Permission[]> {
  return get('/permissions/tree')
}

/**
 * 获取权限列表
 */
export function getPermissionList(params?: { type?: string; status?: number }): Promise<Permission[]> {
  return get('/permissions', params)
}

/**
 * 获取权限详情
 */
export function getPermissionDetail(id: number): Promise<Permission> {
  return get(`/permissions/${id}`)
}

/**
 * 创建权限
 */
export function createPermission(data: Partial<Permission>): Promise<Permission> {
  return post('/permissions', data)
}

/**
 * 更新权限
 */
export function updatePermission(id: number, data: Partial<Permission>): Promise<Permission> {
  return put(`/permissions/${id}`, data)
}

/**
 * 删除权限
 */
export function deletePermission(id: number): Promise<void> {
  return del(`/permissions/${id}`)
}

/**
 * 更新权限状态
 */
export function updatePermissionStatus(id: number, status: number): Promise<void> {
  return put(`/permissions/${id}/status`, { status })
}

/**
 * 获取用户权限
 */
export function getUserPermissions(): Promise<Permission[]> {
  return get('/user/permissions')
}
