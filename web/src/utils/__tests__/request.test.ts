import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import axios from 'axios'

// Mock axios
vi.mock('axios')

// Mock auth utils
vi.mock('../auth', () => ({
  getToken: vi.fn(),
  setToken: vi.fn(),
  removeToken: vi.fn(),
  getRefreshToken: vi.fn(),
  setRefreshToken: vi.fn(),
  removeRefreshToken: vi.fn()
}))

// Mock Element Plus
vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn()
  },
  ElMessageBox: {
    confirm: vi.fn(),
    alert: vi.fn(),
    prompt: vi.fn()
  }
}))

describe('request.ts', () => {
  let mockAxiosInstance: any

  beforeEach(async () => {
    vi.clearAllMocks()

    // 创建 mock axios 实例
    mockAxiosInstance = {
      interceptors: {
        request: {
          use: vi.fn()
        },
        response: {
          use: vi.fn()
        }
      },
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
      request: vi.fn()
    }

    vi.mocked(axios.create).mockReturnValue(mockAxiosInstance)

    // 重新导入 request 模块以触发 axios.create
    vi.resetModules()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('Axios实例创建', () => {
    it('应该创建axios实例', async () => {
      await import('../request')

      expect(axios.create).toHaveBeenCalled()
    })

    it('应该配置正确的参数', async () => {
      await import('../request')

      expect(axios.create).toHaveBeenCalledWith(
        expect.objectContaining({
          timeout: 30000,
          headers: {
            'Content-Type': 'application/json;charset=UTF-8'
          }
        })
      )
    })
  })

  describe('请求拦截器', () => {
    it('应该注册请求拦截器', async () => {
      await import('../request')

      expect(mockAxiosInstance.interceptors.request.use).toHaveBeenCalled()
    })
  })

  describe('响应拦截器', () => {
    it('应该注册响应拦截器', async () => {
      await import('../request')

      expect(mockAxiosInstance.interceptors.response.use).toHaveBeenCalled()
    })
  })

  describe('请求方法', () => {
    it('get方法应该发送GET请求', async () => {
      const mockData = { id: 1 }
      mockAxiosInstance.get.mockResolvedValue(mockData)

      const { get } = await import('../request')
      const result = await get('/test', { id: 1 })

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/test', {
        params: { id: 1 }
      })
      expect(result).toEqual(mockData)
    })

    it('post方法应该发送POST请求', async () => {
      const mockData = { id: 1 }
      mockAxiosInstance.post.mockResolvedValue(mockData)

      const { post } = await import('../request')
      const result = await post('/test', { name: 'test' })

      expect(mockAxiosInstance.post).toHaveBeenCalledWith('/test', { name: 'test' }, undefined)
      expect(result).toEqual(mockData)
    })

    it('put方法应该发送PUT请求', async () => {
      const mockData = { id: 1 }
      mockAxiosInstance.put.mockResolvedValue(mockData)

      const { put } = await import('../request')
      const result = await put('/test/1', { name: 'updated' })

      expect(mockAxiosInstance.put).toHaveBeenCalledWith('/test/1', { name: 'updated' }, undefined)
      expect(result).toEqual(mockData)
    })

    it('del方法应该发送DELETE请求', async () => {
      mockAxiosInstance.delete.mockResolvedValue(undefined)

      const { del } = await import('../request')
      await del('/test/1')

      expect(mockAxiosInstance.delete).toHaveBeenCalledWith('/test/1', { params: undefined })
    })
  })

  describe('文件上传', () => {
    it('应该成功上传文件', async () => {
      const mockResponse = { url: 'http://example.com/file.pdf' }
      mockAxiosInstance.post.mockResolvedValue(mockResponse)

      const { upload } = await import('../request')
      const file = new File(['test'], 'test.txt', { type: 'text/plain' })
      const result = await upload('/upload', file)

      expect(mockAxiosInstance.post).toHaveBeenCalledWith(
        '/upload',
        expect.any(FormData),
        expect.objectContaining({
          headers: {
            'Content-Type': 'multipart/form-data'
          }
        })
      )
      expect(result).toEqual(mockResponse)
    })
  })
})
