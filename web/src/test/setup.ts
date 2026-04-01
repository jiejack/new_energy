import { config } from '@vue/test-utils'
import { vi } from 'vitest'

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

// 配置全局组件
config.global.stubs = {
  'el-button': true,
  'el-input': true,
  'el-table': true,
  'el-table-column': true,
  'el-pagination': true,
  'el-message': true,
  'el-message-box': true
}

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn()
}
global.localStorage = localStorageMock as any

// Mock window.location
const originalLocation = window.location
delete (window as any).location
window.location = {
  ...originalLocation,
  href: ''
} as any
