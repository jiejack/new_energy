import axios from 'axios'
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse, AxiosError } from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getToken, getRefreshToken, setToken, removeToken, removeRefreshToken } from './auth'
import type { ApiResponse } from '@/types'

// 创建axios实例
const service: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json;charset=UTF-8',
  },
})

// 是否正在刷新token
let isRefreshing = false
// 重试请求队列
let retryQueue: Array<(token: string) => void> = []

/**
 * 请求拦截器
 */
service.interceptors.request.use(
  (config) => {
    // 添加token
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }

    // 添加租户ID（如果需要）
    const tenantId = localStorage.getItem('tenantId')
    if (tenantId) {
      config.headers['X-Tenant-Id'] = tenantId
    }

    return config
  },
  (error) => {
    console.error('Request error:', error)
    return Promise.reject(error)
  }
)

/**
 * 响应拦截器
 */
service.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>): any => {
    const { code, message, data, total, page, page_size } = response.data

    // 根据code判断请求是否成功
    if (code === 200 || code === 0) {
      // 返回完整数据结构，包含分页信息
      return {
        list: data || [],
        total: total || 0,
        page: page || 1,
        pageSize: page_size || 10
      }
    }

    // 处理业务错误
    handleBusinessError(code, message)
    return Promise.reject(new Error(message || '请求失败'))
  },
  async (error: AxiosError<ApiResponse>) => {
    const { response } = error

    if (!response) {
      ElMessage.error('网络连接失败，请检查网络')
      return Promise.reject(error)
    }

    const { status, data } = response

    // 处理401未授权错误
    if (status === 401) {
      return handleUnauthorized(error)
    }

    // 处理其他HTTP错误
    handleHttpError(status, data?.message)
    return Promise.reject(error)
  }
)

/**
 * 处理业务错误
 */
function handleBusinessError(code: number, message: string) {
  switch (code) {
    case 401:
      ElMessage.error('登录已过期，请重新登录')
      // 跳转到登录页
      window.location.href = '/login'
      break
    case 403:
      ElMessage.error('没有权限访问')
      break
    case 404:
      ElMessage.error('请求的资源不存在')
      break
    case 500:
      ElMessage.error('服务器内部错误')
      break
    default:
      ElMessage.error(message || '请求失败')
  }
}

/**
 * 处理HTTP错误
 */
function handleHttpError(status: number, message?: string) {
  switch (status) {
    case 400:
      ElMessage.error(message || '请求参数错误')
      break
    case 401:
      ElMessage.error('未授权，请登录')
      break
    case 403:
      ElMessage.error('拒绝访问')
      break
    case 404:
      ElMessage.error('请求地址不存在')
      break
    case 408:
      ElMessage.error('请求超时')
      break
    case 500:
      ElMessage.error('服务器内部错误')
      break
    case 501:
      ElMessage.error('服务未实现')
      break
    case 502:
      ElMessage.error('网关错误')
      break
    case 503:
      ElMessage.error('服务不可用')
      break
    case 504:
      ElMessage.error('网关超时')
      break
    case 505:
      ElMessage.error('HTTP版本不受支持')
      break
    default:
      ElMessage.error(message || '请求失败')
  }
}

/**
 * 处理401未授权错误
 */
async function handleUnauthorized(error: AxiosError): Promise<any> {
  const refreshToken = getRefreshToken()

  if (!refreshToken) {
    // 没有refreshToken，直接跳转登录页
    clearAuthAndRedirect()
    return Promise.reject(error)
  }

  if (isRefreshing) {
    // 正在刷新token，将请求加入队列
    return new Promise((resolve) => {
      retryQueue.push((token: string) => {
        if (error.config) {
          error.config.headers.Authorization = `Bearer ${token}`
          resolve(service(error.config))
        }
      })
    })
  }

  isRefreshing = true

  try {
    // 刷新token
    const response = await axios.post('/api/auth/refresh', {
      refreshToken,
    })

    const { token: newToken, refreshToken: newRefreshToken } = response.data.data

    // 保存新token
    setToken(newToken)
    localStorage.setItem('nem_refresh_token', newRefreshToken)

    // 重试队列中的请求
    retryQueue.forEach((callback) => callback(newToken))
    retryQueue = []

    // 重试当前请求
    if (error.config) {
      error.config.headers.Authorization = `Bearer ${newToken}`
      return service(error.config)
    }
  } catch (refreshError) {
    // 刷新token失败，清除认证信息并跳转登录页
    clearAuthAndRedirect()
    return Promise.reject(refreshError)
  } finally {
    isRefreshing = false
  }
}

/**
 * 清除认证信息并跳转登录页
 */
function clearAuthAndRedirect() {
  removeToken()
  removeRefreshToken()
  ElMessageBox.confirm('登录已过期，请重新登录', '提示', {
    confirmButtonText: '重新登录',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(() => {
    window.location.href = '/login'
  })
}

/**
 * 通用请求方法
 */
export function request<T = any>(config: AxiosRequestConfig): Promise<T> {
  return service.request<any, T>(config)
}

/**
 * GET请求
 */
export function get<T = any>(url: string, params?: any, config?: AxiosRequestConfig): Promise<T> {
  return service.get(url, { params, ...config })
}

/**
 * POST请求
 */
export function post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
  return service.post(url, data, config)
}

/**
 * PUT请求
 */
export function put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
  return service.put(url, data, config)
}

/**
 * DELETE请求
 */
export function del<T = any>(url: string, params?: any, config?: AxiosRequestConfig): Promise<T> {
  return service.delete(url, { params, ...config })
}

/**
 * 文件上传
 */
export function upload<T = any>(url: string, file: File, onProgress?: (progress: number) => void): Promise<T> {
  const formData = new FormData()
  formData.append('file', file)

  return service.post(url, formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
    onUploadProgress: (progressEvent) => {
      if (progressEvent.total && onProgress) {
        const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total)
        onProgress(progress)
      }
    },
  })
}

/**
 * 文件下载
 */
export function download(url: string, params?: any, filename?: string): Promise<void> {
  return service
    .get(url, {
      params,
      responseType: 'blob',
    })
    .then((response: any) => {
      const blob = new Blob([response])
      const link = document.createElement('a')
      link.href = URL.createObjectURL(blob)
      link.download = filename || 'download'
      link.click()
      URL.revokeObjectURL(link.href)
    })
}

export default service
