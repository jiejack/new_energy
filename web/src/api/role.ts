import { get, post, put, del } from '@/utils/request'
import type { Role, PageQuery, PageResult } from '@/types'

/**
 * 获取角色列表
 */
export function getRoleList(params: PageQuery & { keyword?: string; status?: number }): Promise<PageResult<Role>> {
  return get('/roles', params)
}

/**
 * 获取所有角色（不分页）
 */
export function getAllRoles(): Promise<Role[]> {
  return get('/roles/all')
}

/**
 * 获取角色详情
 */
export function getRoleDetail(id: number): Promise<Role> {
  return get(`/roles/${id}`)
}

/**
 * 创建角色
 */
export function createRole(data: Partial<Role>): Promise<Role> {
  return post('/roles', data)
}

/**
 * 更新角色
 */
export function updateRole(id: number, data: Partial<Role>): Promise<Role> {
  return put(`/roles/${id}`, data)
}

/**
 * 删除角色
 */
export function deleteRole(id: number): Promise<void> {
  return del(`/roles/${id}`)
}

/**
 * 批量删除角色
 */
export function batchDeleteRoles(ids: number[]): Promise<void> {
  return post('/roles/batch-delete', { ids })
}

/**
 * 更新角色状态
 */
export function updateRoleStatus(id: number, status: number): Promise<void> {
  return put(`/roles/${id}/status`, { status })
}

/**
 * 分配权限
 */
export function assignPermissions(roleId: number, permissionIds: number[]): Promise<void> {
  return post(`/roles/${roleId}/permissions`, { permissionIds })
}

/**
 * 获取角色权限
 */
export function getRolePermissions(roleId: number): Promise<number[]> {
  return get(`/roles/${roleId}/permissions`)
}
