import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAppStore } from '../app'

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn()
}
global.localStorage = localStorageMock as any

describe('App Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    // 重置 document.documentElement
    document.documentElement.removeAttribute('data-theme')
    document.documentElement.classList.remove('dark')
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('状态初始化', () => {
    it('应该初始化侧边栏状态为打开', () => {
      localStorageMock.getItem.mockReturnValue(null)
      const store = useAppStore()

      expect(store.sidebar.opened).toBe(true)
      expect(store.sidebar.withoutAnimation).toBe(false)
    })

    it('应该从localStorage读取侧边栏状态', () => {
      localStorageMock.getItem.mockReturnValue('closed')
      const store = useAppStore()

      expect(store.sidebar.opened).toBe(false)
    })

    it('应该初始化设备类型为desktop', () => {
      const store = useAppStore()

      expect(store.device).toBe('desktop')
    })

    it('应该从localStorage读取size', () => {
      localStorageMock.getItem.mockReturnValue('small')
      const store = useAppStore()

      expect(store.size).toBe('small')
    })

    it('应该从localStorage读取theme', () => {
      localStorageMock.getItem.mockReturnValue('dark')
      const store = useAppStore()

      expect(store.theme).toBe('dark')
    })

    it('应该初始化loading为false', () => {
      const store = useAppStore()

      expect(store.loading).toBe(false)
    })
  })

  describe('计算属性', () => {
    it('sidebarOpened应该返回侧边栏打开状态', () => {
      const store = useAppStore()

      expect(store.sidebarOpened).toBe(true)

      store.sidebar.opened = false

      expect(store.sidebarOpened).toBe(false)
    })
  })

  describe('toggleSidebar', () => {
    it('应该切换侧边栏状态', () => {
      const store = useAppStore()

      store.toggleSidebar()

      expect(store.sidebar.opened).toBe(false)
      expect(localStorageMock.setItem).toHaveBeenCalledWith('sidebarStatus', 'closed')

      store.toggleSidebar()

      expect(store.sidebar.opened).toBe(true)
      expect(localStorageMock.setItem).toHaveBeenCalledWith('sidebarStatus', 'opened')
    })

    it('应该设置动画标志', () => {
      const store = useAppStore()

      store.toggleSidebar(true)

      expect(store.sidebar.withoutAnimation).toBe(true)
    })
  })

  describe('closeSidebar', () => {
    it('应该关闭侧边栏', () => {
      const store = useAppStore()
      store.sidebar.opened = true

      store.closeSidebar()

      expect(store.sidebar.opened).toBe(false)
      expect(localStorageMock.setItem).toHaveBeenCalledWith('sidebarStatus', 'closed')
    })

    it('应该设置动画标志', () => {
      const store = useAppStore()

      store.closeSidebar(true)

      expect(store.sidebar.withoutAnimation).toBe(true)
    })
  })

  describe('toggleDevice', () => {
    it('应该切换设备类型', () => {
      const store = useAppStore()

      store.toggleDevice('mobile')

      expect(store.device).toBe('mobile')

      store.toggleDevice('desktop')

      expect(store.device).toBe('desktop')
    })
  })

  describe('setSize', () => {
    it('应该设置元素大小', () => {
      const store = useAppStore()

      store.setSize('large')

      expect(store.size).toBe('large')
      expect(localStorageMock.setItem).toHaveBeenCalledWith('size', 'large')
    })
  })

  describe('toggleTheme', () => {
    it('应该切换主题', () => {
      localStorageMock.getItem.mockReturnValue('light')
      const store = useAppStore()

      // 确保初始主题是 light
      expect(store.theme).toBe('light')

      store.toggleTheme()

      expect(store.theme).toBe('dark')
      expect(localStorageMock.setItem).toHaveBeenCalledWith('theme', 'dark')
      expect(document.documentElement.getAttribute('data-theme')).toBe('dark')

      store.toggleTheme()

      expect(store.theme).toBe('light')
      expect(localStorageMock.setItem).toHaveBeenCalledWith('theme', 'light')
      expect(document.documentElement.getAttribute('data-theme')).toBe('light')
    })
  })

  describe('setLoading', () => {
    it('应该设置加载状态', () => {
      const store = useAppStore()

      store.setLoading(true)

      expect(store.loading).toBe(true)

      store.setLoading(false)

      expect(store.loading).toBe(false)
    })
  })
})
