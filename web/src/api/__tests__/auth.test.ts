import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { setupServer } from 'msw/node'
import * as authApi from '../auth'
import type { LoginForm, LoginResponse, UserInfo } from '@/types'

// Mock request utils
vi.mock('@/utils/request', () => ({
  post: vi.fn(),
  get: vi.fn()
}))

import { post, get } from '@/utils/request'

// Setup MSW server
const server = setupServer()

describe('Auth API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    server.listen()
  })

  afterEach(() => {
    server.close()
    vi.restoreAllMocks()
  })

  describe('登录接口', () => {
    it('应该成功登录', async () => {
      const mockLoginResponse: LoginResponse = {
        token: 'test-token',
        refreshToken: 'test-refresh-token',
        expiresIn: 3600,
        user: {
          id: 1,
          username: 'testuser',
          nickname: 'Test User',
          email: 'test@example.com',
          phone: '13800138000',
          avatar: '',
          status: 1,
          roles: ['admin'],
          permissions: ['*'],
          createdAt: '2024-01-01',
          updatedAt: '2024-01-01'
        }
      }

      vi.mocked(post).mockResolvedValue(mockLoginResponse)

      const loginForm: LoginForm = {
        username: 'testuser',
        password: 'password123'
      }

      const result = await authApi.login(loginForm)

      expect(post).toHaveBeenCalledWith('/auth/login', loginForm)
      expect(result).toEqual(mockLoginResponse)
      expect(result.token).toBe('test-token')
      expect(result.user.username).toBe('testuser')
    })

    it('登录失败应该抛出错误', async () => {
      const mockError = new Error('用户名或密码错误')
      vi.mocked(post).mockRejectedValue(mockError)

      const loginForm: LoginForm = {
        username: 'wronguser',
        password: 'wrongpassword'
      }

      await expect(authApi.login(loginForm)).rejects.toThrow('用户名或密码错误')
      expect(post).toHaveBeenCalledWith('/auth/login', loginForm)
    })

    it('应该发送正确的登录数据', async () => {
      const mockResponse = {
        token: 'token',
        refreshToken: 'refresh',
        expiresIn: 3600,
        user: {} as any
      }

      vi.mocked(post).mockResolvedValue(mockResponse)

      const loginForm: LoginForm = {
        username: 'admin',
        password: 'admin123',
        captcha: '1234',
        uuid: 'uuid-123'
      }

      await authApi.login(loginForm)

      expect(post).toHaveBeenCalledWith('/auth/login', {
        username: 'admin',
        password: 'admin123',
        captcha: '1234',
        uuid: 'uuid-123'
      })
    })
  })

  describe('登出接口', () => {
    it('应该成功登出', async () => {
      vi.mocked(post).mockResolvedValue(undefined)

      await authApi.logout()

      expect(post).toHaveBeenCalledWith('/auth/logout')
    })

    it('登出失败应该抛出错误', async () => {
      const mockError = new Error('登出失败')
      vi.mocked(post).mockRejectedValue(mockError)

      await expect(authApi.logout()).rejects.toThrow('登出失败')
    })
  })

  describe('获取用户信息接口', () => {
    it('应该成功获取用户信息', async () => {
      const mockUserInfo: UserInfo = {
        id: 1,
        username: 'testuser',
        nickname: 'Test User',
        email: 'test@example.com',
        phone: '13800138000',
        avatar: '',
        status: 1,
        roles: ['admin', 'user'],
        permissions: ['user:read', 'user:write', 'user:delete'],
        createdAt: '2024-01-01',
        updatedAt: '2024-01-01'
      }

      vi.mocked(get).mockResolvedValue(mockUserInfo)

      const result = await authApi.getUserInfo()

      expect(get).toHaveBeenCalledWith('/auth/user-info')
      expect(result).toEqual(mockUserInfo)
      expect(result.username).toBe('testuser')
      expect(result.roles).toContain('admin')
      expect(result.permissions).toHaveLength(3)
    })

    it('获取用户信息失败应该抛出错误', async () => {
      const mockError = new Error('未授权')
      vi.mocked(get).mockRejectedValue(mockError)

      await expect(authApi.getUserInfo()).rejects.toThrow('未授权')
    })

    it('应该返回完整的用户信息', async () => {
      const mockUserInfo: UserInfo = {
        id: 1,
        username: 'admin',
        nickname: 'Administrator',
        email: 'admin@example.com',
        phone: '13900139000',
        avatar: 'https://example.com/avatar.png',
        status: 1,
        roles: ['admin'],
        permissions: ['*'],
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-02T00:00:00Z'
      }

      vi.mocked(get).mockResolvedValue(mockUserInfo)

      const result = await authApi.getUserInfo()

      expect(result.id).toBe(1)
      expect(result.username).toBe('admin')
      expect(result.nickname).toBe('Administrator')
      expect(result.email).toBe('admin@example.com')
      expect(result.phone).toBe('13900139000')
      expect(result.avatar).toBe('https://example.com/avatar.png')
      expect(result.status).toBe(1)
      expect(result.roles).toEqual(['admin'])
      expect(result.permissions).toEqual(['*'])
    })
  })

  describe('刷新Token接口', () => {
    it('应该成功刷新Token', async () => {
      const mockResponse = {
        token: 'new-token',
        refreshToken: 'new-refresh-token'
      }

      vi.mocked(post).mockResolvedValue(mockResponse)

      const result = await authApi.refreshToken('old-refresh-token')

      expect(post).toHaveBeenCalledWith('/auth/refresh', {
        refreshToken: 'old-refresh-token'
      })
      expect(result).toEqual(mockResponse)
      expect(result.token).toBe('new-token')
      expect(result.refreshToken).toBe('new-refresh-token')
    })

    it('刷新Token失败应该抛出错误', async () => {
      const mockError = new Error('刷新Token失败')
      vi.mocked(post).mockRejectedValue(mockError)

      await expect(authApi.refreshToken('invalid-token')).rejects.toThrow('刷新Token失败')
    })

    it('应该发送正确的refreshToken', async () => {
      const mockResponse = {
        token: 'new-token',
        refreshToken: 'new-refresh-token'
      }

      vi.mocked(post).mockResolvedValue(mockResponse)

      await authApi.refreshToken('test-refresh-token')

      expect(post).toHaveBeenCalledWith('/auth/refresh', {
        refreshToken: 'test-refresh-token'
      })
    })
  })

  describe('获取验证码接口', () => {
    it('应该成功获取验证码', async () => {
      const mockResponse = {
        uuid: 'uuid-123',
        captcha: 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=='
      }

      vi.mocked(get).mockResolvedValue(mockResponse)

      const result = await authApi.getCaptcha()

      expect(get).toHaveBeenCalledWith('/auth/captcha')
      expect(result).toEqual(mockResponse)
      expect(result.uuid).toBe('uuid-123')
      expect(result.captcha).toContain('data:image')
    })

    it('获取验证码失败应该抛出错误', async () => {
      const mockError = new Error('获取验证码失败')
      vi.mocked(get).mockRejectedValue(mockError)

      await expect(authApi.getCaptcha()).rejects.toThrow('获取验证码失败')
    })
  })

  describe('修改密码接口', () => {
    it('应该成功修改密码', async () => {
      vi.mocked(post).mockResolvedValue(undefined)

      const data = {
        oldPassword: 'oldpass123',
        newPassword: 'newpass456'
      }

      await authApi.changePassword(data)

      expect(post).toHaveBeenCalledWith('/auth/change-password', data)
    })

    it('修改密码失败应该抛出错误', async () => {
      const mockError = new Error('原密码错误')
      vi.mocked(post).mockRejectedValue(mockError)

      const data = {
        oldPassword: 'wrongpass',
        newPassword: 'newpass456'
      }

      await expect(authApi.changePassword(data)).rejects.toThrow('原密码错误')
    })

    it('应该发送正确的密码数据', async () => {
      vi.mocked(post).mockResolvedValue(undefined)

      const data = {
        oldPassword: 'OldPass123!',
        newPassword: 'NewPass456!'
      }

      await authApi.changePassword(data)

      expect(post).toHaveBeenCalledWith('/auth/change-password', {
        oldPassword: 'OldPass123!',
        newPassword: 'NewPass456!'
      })
    })
  })
})
