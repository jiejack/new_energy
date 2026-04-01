import { post, get } from '@/utils/request'
import type { LoginForm, LoginResponse, UserInfo } from '@/types'

/**
 * 登录
 */
export function login(data: LoginForm): Promise<LoginResponse> {
  return post('/auth/login', data)
}

/**
 * 登出
 */
export function logout(): Promise<void> {
  return post('/auth/logout')
}

/**
 * 获取用户信息
 */
export function getUserInfo(): Promise<UserInfo> {
  return get('/auth/user-info')
}

/**
 * 刷新Token
 */
export function refreshToken(refreshToken: string): Promise<{ token: string; refreshToken: string }> {
  return post('/auth/refresh', { refreshToken })
}

/**
 * 获取验证码
 */
export function getCaptcha(): Promise<{ uuid: string; captcha: string }> {
  return get('/auth/captcha')
}

/**
 * 修改密码
 */
export function changePassword(data: { oldPassword: string; newPassword: string }): Promise<void> {
  return post('/auth/change-password', data)
}
