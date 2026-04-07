import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import {
  getToken,
  setToken,
  removeToken,
  getRefreshToken,
  setRefreshToken,
  removeRefreshToken,
  clearAuth
} from '../auth'

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn()
}
global.localStorage = localStorageMock as any

describe('Auth Utils', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('Token管理', () => {
    describe('getToken', () => {
      it('应该从localStorage获取token', () => {
        const mockToken = 'test-token-123'
        localStorageMock.getItem.mockReturnValue(mockToken)

        const result = getToken()

        expect(localStorageMock.getItem).toHaveBeenCalledWith('nem_token')
        expect(result).toBe(mockToken)
      })

      it('没有token时应该返回null', () => {
        localStorageMock.getItem.mockReturnValue(null)

        const result = getToken()

        expect(result).toBeNull()
      })
    })

    describe('setToken', () => {
      it('应该保存token到localStorage', () => {
        const token = 'new-token-456'

        setToken(token)

        expect(localStorageMock.setItem).toHaveBeenCalledWith('nem_token', token)
      })
    })

    describe('removeToken', () => {
      it('应该从localStorage移除token', () => {
        removeToken()

        expect(localStorageMock.removeItem).toHaveBeenCalledWith('nem_token')
      })
    })
  })

  describe('RefreshToken管理', () => {
    describe('getRefreshToken', () => {
      it('应该从localStorage获取refreshToken', () => {
        const mockRefreshToken = 'refresh-token-123'
        localStorageMock.getItem.mockReturnValue(mockRefreshToken)

        const result = getRefreshToken()

        expect(localStorageMock.getItem).toHaveBeenCalledWith('nem_refresh_token')
        expect(result).toBe(mockRefreshToken)
      })

      it('没有refreshToken时应该返回null', () => {
        localStorageMock.getItem.mockReturnValue(null)

        const result = getRefreshToken()

        expect(result).toBeNull()
      })
    })

    describe('setRefreshToken', () => {
      it('应该保存refreshToken到localStorage', () => {
        const refreshToken = 'new-refresh-token-456'

        setRefreshToken(refreshToken)

        expect(localStorageMock.setItem).toHaveBeenCalledWith('nem_refresh_token', refreshToken)
      })
    })

    describe('removeRefreshToken', () => {
      it('应该从localStorage移除refreshToken', () => {
        removeRefreshToken()

        expect(localStorageMock.removeItem).toHaveBeenCalledWith('nem_refresh_token')
      })
    })
  })

  describe('clearAuth', () => {
    it('应该清除所有认证信息', () => {
      clearAuth()

      expect(localStorageMock.removeItem).toHaveBeenCalledWith('nem_token')
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('nem_refresh_token')
    })

    it('应该调用removeToken和removeRefreshToken', () => {
      clearAuth()

      expect(localStorageMock.removeItem).toHaveBeenCalledTimes(2)
    })
  })
})
