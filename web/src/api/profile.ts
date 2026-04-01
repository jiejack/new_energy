import { get, put } from '@/utils/request'
import { upload } from '@/utils/request'
import type { UserInfo } from '@/types'

/**
 * 获取个人信息
 */
export function getProfile(): Promise<UserInfo> {
  return get('/profile')
}

/**
 * 更新个人信息
 */
export function updateProfile(data: Partial<UserInfo>): Promise<UserInfo> {
  return put('/profile', data)
}

/**
 * 修改密码
 */
export function changePassword(data: {
  oldPassword: string
  newPassword: string
  confirmPassword: string
}): Promise<void> {
  return put('/profile/password', data)
}

/**
 * 上传头像
 */
export function uploadAvatar(file: File, onProgress?: (progress: number) => void): Promise<{ url: string }> {
  return upload('/profile/avatar', file, onProgress)
}

/**
 * 更新偏好设置
 */
export function updatePreferences(data: Record<string, any>): Promise<void> {
  return put('/profile/preferences', data)
}

/**
 * 获取偏好设置
 */
export function getPreferences(): Promise<Record<string, any>> {
  return get('/profile/preferences')
}
