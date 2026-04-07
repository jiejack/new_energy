import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import * as userApi from '../user'

// Mock request utils
vi.mock('@/utils/request', () => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  del: vi.fn()
}))

import { get, post, put, del } from '@/utils/request'

describe('User API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('获取用户列表', () => {
    it('应该成功获取用户列表', async () => {
      const mockResponse = {
        list: [
          { id: 1, username: 'user1' },
          { id: 2, username: 'user2' }
        ],
        total: 2,
        page: 1,
        pageSize: 10
      }

      vi.mocked(get).mockResolvedValue(mockResponse)

      const params = { page: 1, pageSize: 10 }
      const result = await userApi.getUserList(params)

      expect(get).toHaveBeenCalledWith('/users', params)
      expect(result).toEqual(mockResponse)
    })

    it('应该支持关键字搜索', async () => {
      const mockResponse = {
        list: [{ id: 1, username: 'test' }],
        total: 1,
        page: 1,
        pageSize: 10
      }

      vi.mocked(get).mockResolvedValue(mockResponse)

      const params = { page: 1, pageSize: 10, keyword: 'test' }
      await userApi.getUserList(params)

      expect(get).toHaveBeenCalledWith('/users', params)
    })

    it('应该支持状态筛选', async () => {
      const mockResponse = {
        list: [{ id: 1, username: 'active', status: 1 }],
        total: 1,
        page: 1,
        pageSize: 10
      }

      vi.mocked(get).mockResolvedValue(mockResponse)

      const params = { page: 1, pageSize: 10, status: 1 }
      await userApi.getUserList(params)

      expect(get).toHaveBeenCalledWith('/users', params)
    })
  })

  describe('获取用户详情', () => {
    it('应该成功获取用户详情', async () => {
      const mockUser = {
        id: 1,
        username: 'testuser',
        nickname: 'Test User',
        email: 'test@example.com'
      }

      vi.mocked(get).mockResolvedValue(mockUser)

      const result = await userApi.getUserDetail(1)

      expect(get).toHaveBeenCalledWith('/users/1')
      expect(result).toEqual(mockUser)
    })
  })

  describe('创建用户', () => {
    it('应该成功创建用户', async () => {
      const mockUser = {
        id: 1,
        username: 'newuser',
        nickname: 'New User',
        email: 'new@example.com'
      }

      vi.mocked(post).mockResolvedValue(mockUser)

      const data = {
        username: 'newuser',
        nickname: 'New User',
        email: 'new@example.com',
        password: 'password123'
      }

      const result = await userApi.createUser(data)

      expect(post).toHaveBeenCalledWith('/users', data)
      expect(result).toEqual(mockUser)
    })
  })

  describe('更新用户', () => {
    it('应该成功更新用户', async () => {
      const mockUser = {
        id: 1,
        username: 'updateduser',
        nickname: 'Updated User'
      }

      vi.mocked(put).mockResolvedValue(mockUser)

      const data = { nickname: 'Updated User' }
      const result = await userApi.updateUser(1, data)

      expect(put).toHaveBeenCalledWith('/users/1', data)
      expect(result).toEqual(mockUser)
    })
  })

  describe('删除用户', () => {
    it('应该成功删除用户', async () => {
      vi.mocked(del).mockResolvedValue(undefined)

      await userApi.deleteUser(1)

      expect(del).toHaveBeenCalledWith('/users/1')
    })
  })

  describe('批量删除用户', () => {
    it('应该成功批量删除用户', async () => {
      vi.mocked(post).mockResolvedValue(undefined)

      const ids = [1, 2, 3]
      await userApi.batchDeleteUsers(ids)

      expect(post).toHaveBeenCalledWith('/users/batch-delete', { ids })
    })
  })

  describe('重置密码', () => {
    it('应该成功重置密码', async () => {
      vi.mocked(post).mockResolvedValue(undefined)

      await userApi.resetPassword(1)

      expect(post).toHaveBeenCalledWith('/users/1/reset-password')
    })
  })

  describe('更新用户状态', () => {
    it('应该成功更新用户状态', async () => {
      vi.mocked(put).mockResolvedValue(undefined)

      await userApi.updateUserStatus(1, 0)

      expect(put).toHaveBeenCalledWith('/users/1/status', { status: 0 })
    })
  })

  describe('分配角色', () => {
    it('应该成功分配角色', async () => {
      vi.mocked(post).mockResolvedValue(undefined)

      const roleIds = [1, 2, 3]
      await userApi.assignRoles(1, roleIds)

      expect(post).toHaveBeenCalledWith('/users/1/roles', { roleIds })
    })
  })

  describe('获取用户角色', () => {
    it('应该成功获取用户角色', async () => {
      const mockRoleIds = [1, 2, 3]
      vi.mocked(get).mockResolvedValue(mockRoleIds)

      const result = await userApi.getUserRoles(1)

      expect(get).toHaveBeenCalledWith('/users/1/roles')
      expect(result).toEqual(mockRoleIds)
    })
  })
})
