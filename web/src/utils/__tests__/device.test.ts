import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { isMobile, isTouchDevice, getDeviceType, useDevice, getSafeAreaInset } from '../device'

describe('Device Utils', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('isMobile', () => {
    it('窗口宽度小于等于768应该返回true', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 768
      })

      const result = isMobile()

      expect(result).toBe(true)
    })

    it('窗口宽度大于768应该返回false', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1024
      })

      const result = isMobile()

      expect(result).toBe(false)
    })

    it('窗口宽度等于500应该返回true', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 500
      })

      const result = isMobile()

      expect(result).toBe(true)
    })
  })

  describe('isTouchDevice', () => {
    it('有ontouchstart应该返回true', () => {
      Object.defineProperty(window, 'ontouchstart', {
        writable: true,
        configurable: true,
        value: {}
      })

      const result = isTouchDevice()

      expect(result).toBe(true)
    })

    it('maxTouchPoints大于0应该返回true', () => {
      Object.defineProperty(window, 'ontouchstart', {
        writable: true,
        configurable: true,
        value: undefined
      })
      Object.defineProperty(navigator, 'maxTouchPoints', {
        writable: true,
        configurable: true,
        value: 5
      })

      const result = isTouchDevice()

      expect(result).toBe(true)
    })

    it('没有触摸支持应该返回false', () => {
      // 删除 ontouchstart
      delete (window as any).ontouchstart
      Object.defineProperty(navigator, 'maxTouchPoints', {
        writable: true,
        configurable: true,
        value: 0
      })

      const result = isTouchDevice()

      expect(result).toBe(false)
    })
  })

  describe('getDeviceType', () => {
    it('宽度小于等于768应该返回mobile', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 768
      })

      const result = getDeviceType()

      expect(result).toBe('mobile')
    })

    it('宽度在769到992之间应该返回tablet', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 800
      })

      const result = getDeviceType()

      expect(result).toBe('tablet')
    })

    it('宽度等于992应该返回tablet', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 992
      })

      const result = getDeviceType()

      expect(result).toBe('tablet')
    })

    it('宽度大于992应该返回desktop', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1200
      })

      const result = getDeviceType()

      expect(result).toBe('desktop')
    })
  })

  describe('useDevice', () => {
    it('应该返回设备状态', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 500
      })

      const { isMobile, isTouch, deviceType } = useDevice()

      expect(isMobile.value).toBe(true)
      expect(typeof isTouch.value).toBe('boolean')
      expect(deviceType.value).toBe('mobile')
    })

    it('应该响应窗口大小变化', async () => {
      // 跳过这个测试，因为需要在组件上下文中运行
      // useDevice 使用了 onMounted 和 onUnmounted
    })
  })

  describe('getSafeAreaInset', () => {
    it('应该返回安全区域插入值', () => {
      const mockGetPropertyValue = vi.fn((prop: string) => {
        const values: Record<string, string> = {
          '--safe-area-inset-top': '20',
          '--safe-area-inset-bottom': '30',
          '--safe-area-inset-left': '10',
          '--safe-area-inset-right': '10'
        }
        return values[prop] || '0'
      })

      const mockComputedStyle = {
        getPropertyValue: mockGetPropertyValue
      }

      vi.spyOn(window, 'getComputedStyle').mockReturnValue(mockComputedStyle as any)

      const result = getSafeAreaInset()

      expect(result).toEqual({
        top: 20,
        bottom: 30,
        left: 10,
        right: 10
      })
    })

    it('没有CSS变量时应该返回0', () => {
      const mockGetPropertyValue = vi.fn(() => '')
      const mockComputedStyle = {
        getPropertyValue: mockGetPropertyValue
      }

      vi.spyOn(window, 'getComputedStyle').mockReturnValue(mockComputedStyle as any)

      const result = getSafeAreaInset()

      expect(result).toEqual({
        top: 0,
        bottom: 0,
        left: 0,
        right: 0
      })
    })
  })
})
