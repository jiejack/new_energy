import { get, post, put, del } from '@/utils/request'
import type { UserInfo, PageQuery, PageResult } from '@/types'

/**
 * 获取用户列表
 */
export function getUserList(params: PageQuery & { keyword?: string; status?: number }): Promise<PageResult<UserInfo>> {
  return get('/users', params)
}

/**
 * 获取用户详情
 */
export function getUserDetail(id: number): Promise<UserInfo> {
  return get(`/users/${id}`)
}

/**
 * 创建用户
 */
export function createUser(data: Partial<UserInfo>): Promise<UserInfo> {
  return post('/users', data)
}

/**
 * 更新用户
 */
export function updateUser(id: number, data: Partial<UserInfo>): Promise<UserInfo> {
  return put(`/users/${id}`, data)
}

/**
 * 删除用户
 */
export function deleteUser(id: number): Promise<void> {
  return del(`/users/${id}`)
}

/**
 * 批量删除用户
 */
export function batchDeleteUsers(ids: number[]): Promise<void> {
  return post('/users/batch-delete', { ids })
}

/**
 * 重置密码
 */
export function resetPassword(id: number): Promise<void> {
  return post(`/users/${id}/reset-password`)
}

/**
 * 更新用户状态
 */
export function updateUserStatus(id: number, status: number): Promise<void> {
  return put(`/users/${id}/status`, { status })
}

/**
 * 分配角色
 */
export function assignRoles(userId: number, roleIds: number[]): Promise<void> {
  return post(`/users/${userId}/roles`, { roleIds })
}

/**
 * 获取用户角色
 */
export function getUserRoles(userId: number): Promise<number[]> {
  return get(`/users/${userId}/roles`)
}
