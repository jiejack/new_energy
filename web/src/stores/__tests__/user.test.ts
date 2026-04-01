import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useUserStore } from '../user'
import * as authApi from '@/api/auth'
import * as authUtils from '@/utils/auth'
import type { LoginForm, LoginResponse, UserInfo } from '@/types'

// Mock API
vi.mock('@/api/auth', () => ({
  login: vi.fn(),
  logout: vi.fn(),
  getUserInfo: vi.fn()
}))

// Mock auth utils
vi.mock('@/utils/auth', () => ({
  getToken: vi.fn(),
  setToken: vi.fn(),
  removeToken: vi.fn(),
  getRefreshToken: vi.fn(),
  setRefreshToken: vi.fn(),
  removeRefreshToken: vi.fn()
}))

describe('User Store', () => {
  beforeEach(() => {
    // 创建新的 pinia 实例
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('用户状态初始化', () => {
    it('应该初始化为空状态', () => {
      const store = useUserStore()

      expect(store.token).toBe('')
      expect(store.refreshToken).toBe('')
      expect(store.userInfo).toBeNull()
      expect(store.roles).toEqual([])
      expect(store.permissions).toEqual([])
    })

    it('应该从localStorage读取token', () => {
      const mockToken = 'existing-token'
      const mockRefreshToken = 'existing-refresh-token'

      vi.mocked(authUtils.getToken).mockReturnValue(mockToken)
      vi.mocked(authUtils.getRefreshToken).mockReturnValue(mockRefreshToken)

      const store = useUserStore()

      expect(store.token).toBe(mockToken)
      expect(store.refreshToken).toBe(mockRefreshToken)
    })

    it('isLoggedIn计算属性应该正确返回登录状态', () => {
      vi.mocked(authUtils.getToken).mockReturnValue(null)
      const store = useUserStore()

      expect(store.isLoggedIn).toBe(false)

      store.token = 'test-token'

      expect(store.isLoggedIn).toBe(true)
    })

    it('username计算属性应该返回用户名', () => {
      const store = useUserStore()

      expect(store.username).toBe('')

      store.userInfo = {
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

      expect(store.username).toBe('testuser')
    })

    it('nickname计算属性应该返回昵称', () => {
      const store = useUserStore()

      expect(store.nickname).toBe('')

      store.userInfo = {
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

      expect(store.nickname).toBe('Test User')
    })
  })

  describe('登录action', () => {
    it('应该成功登录并保存token', async () => {
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

      vi.mocked(authApi.login).mockResolvedValue(mockLoginResponse)

      const store = useUserStore()
      const loginForm: LoginForm = {
        username: 'testuser',
        password: 'password123'
      }

      const result = await store.loginAction(loginForm)

      expect(authApi.login).toHaveBeenCalledWith(loginForm)
      expect(store.token).toBe(mockLoginResponse.token)
      expect(store.refreshToken).toBe(mockLoginResponse.refreshToken)
      expect(store.userInfo).toEqual(mockLoginResponse.user)
      expect(store.roles).toEqual(mockLoginResponse.user.roles)
      expect(store.permissions).toEqual(mockLoginResponse.user.permissions)
      expect(authUtils.setToken).toHaveBeenCalledWith(mockLoginResponse.token)
      expect(authUtils.setRefreshToken).toHaveBeenCalledWith(mockLoginResponse.refreshToken)
      expect(result).toEqual(mockLoginResponse)
    })

    it('登录失败应该抛出错误', async () => {
      const mockError = new Error('登录失败')
      vi.mocked(authApi.login).mockRejectedValue(mockError)

      const store = useUserStore()
      const loginForm: LoginForm = {
        username: 'testuser',
        password: 'wrongpassword'
      }

      await expect(store.loginAction(loginForm)).rejects.toThrow('登录失败')
    })
  })

  describe('登出action', () => {
    it('应该成功登出并清除状态', async () => {
      vi.mocked(authApi.logout).mockResolvedValue()

      const store = useUserStore()
      store.token = 'test-token'
      store.refreshToken = 'test-refresh-token'
      store.userInfo = {
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
      store.roles = ['admin']
      store.permissions = ['*']

      await store.logoutAction()

      expect(authApi.logout).toHaveBeenCalled()
      expect(store.token).toBe('')
      expect(store.refreshToken).toBe('')
      expect(store.userInfo).toBeNull()
      expect(store.roles).toEqual([])
      expect(store.permissions).toEqual([])
      expect(authUtils.removeToken).toHaveBeenCalled()
      expect(authUtils.removeRefreshToken).toHaveBeenCalled()
    })

    it('登出API失败时仍应清除状态', async () => {
      const mockError = new Error('登出失败')
      vi.mocked(authApi.logout).mockRejectedValue(mockError)

      const store = useUserStore()
      store.token = 'test-token'

      try {
        await store.logoutAction()
      } catch (error) {
        // 预期会抛出错误
      }

      expect(store.token).toBe('')
      expect(authUtils.removeToken).toHaveBeenCalled()
    })
  })

  describe('获取用户信息action', () => {
    it('应该成功获取用户信息', async () => {
      const mockUserInfo: UserInfo = {
        id: 1,
        username: 'testuser',
        nickname: 'Test User',
        email: 'test@example.com',
        phone: '13800138000',
        avatar: '',
        status: 1,
        roles: ['admin'],
        permissions: ['user:read', 'user:write'],
        createdAt: '2024-01-01',
        updatedAt: '2024-01-01'
      }

      vi.mocked(authApi.getUserInfo).mockResolvedValue(mockUserInfo)

      const store = useUserStore()
      const result = await store.getUserInfoAction()

      expect(authApi.getUserInfo).toHaveBeenCalled()
      // userInfo 不包含 roles 和 permissions
      expect(store.userInfo).toEqual({
        id: 1,
        username: 'testuser',
        nickname: 'Test User',
        email: 'test@example.com',
        phone: '13800138000',
        avatar: '',
        status: 1,
        createdAt: '2024-01-01',
        updatedAt: '2024-01-01'
      })
      expect(store.roles).toEqual(mockUserInfo.roles)
      expect(store.permissions).toEqual(mockUserInfo.permissions)
      expect(result).toEqual({
        roles: mockUserInfo.roles,
        permissions: mockUserInfo.permissions
      })
    })

    it('获取用户信息失败应该抛出错误', async () => {
      const mockError = new Error('获取用户信息失败')
      vi.mocked(authApi.getUserInfo).mockRejectedValue(mockError)

      const store = useUserStore()

      await expect(store.getUserInfoAction()).rejects.toThrow('获取用户信息失败')
    })
  })

  describe('Token管理', () => {
    it('应该更新token', () => {
      const store = useUserStore()
      const newToken = 'new-token'
      const newRefreshToken = 'new-refresh-token'

      store.updateToken(newToken, newRefreshToken)

      expect(store.token).toBe(newToken)
      expect(store.refreshToken).toBe(newRefreshToken)
      expect(authUtils.setToken).toHaveBeenCalledWith(newToken)
      expect(authUtils.setRefreshToken).toHaveBeenCalledWith(newRefreshToken)
    })

    it('应该重置状态', () => {
      const store = useUserStore()
      store.token = 'test-token'
      store.refreshToken = 'test-refresh-token'
      store.userInfo = {
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
      store.roles = ['admin']
      store.permissions = ['*']

      store.resetState()

      expect(store.token).toBe('')
      expect(store.refreshToken).toBe('')
      expect(store.userInfo).toBeNull()
      expect(store.roles).toEqual([])
      expect(store.permissions).toEqual([])
      expect(authUtils.removeToken).toHaveBeenCalled()
      expect(authUtils.removeRefreshToken).toHaveBeenCalled()
    })
  })
})
