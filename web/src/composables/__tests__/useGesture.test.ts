import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { ref } from 'vue'
import { mount } from '@vue/test-utils'
import { useGesture } from '../useGesture'

// 创建测试组件
const createTestComponent = (options: any = {}) => {
  return {
    template: '<div ref="element"></div>',
    setup() {
      const element = ref<HTMLElement | null>(null)
      const gesture = useGesture(element, options)
      return { element, ...gesture }
    }
  }
}

describe('useGesture', () => {
  let wrapper: any
  let element: HTMLElement

  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
    }
    vi.restoreAllMocks()
  })

  describe('初始化', () => {
    it('应该初始化手势状态', () => {
      wrapper = mount(createTestComponent())
      const { isSwiping, swipeDirection } = wrapper.vm

      expect(isSwiping).toBe(false)
      expect(swipeDirection).toEqual({
        left: false,
        right: false,
        up: false,
        down: false
      })
    })
  })

  describe('触摸事件处理', () => {
    beforeEach(() => {
      wrapper = mount(createTestComponent())
      element = wrapper.find('div').element
    })

    it('应该处理touchstart事件', async () => {
      const touchStartEvent = new TouchEvent('touchstart', {
        touches: [{ clientX: 100, clientY: 100 } as Touch]
      })

      element.dispatchEvent(touchStartEvent)

      expect(wrapper.vm.isSwiping).toBe(true)
      expect(wrapper.vm.swipeDirection).toEqual({
        left: false,
        right: false,
        up: false,
        down: false
      })
    })

    it('应该处理touchend事件', async () => {
      // 先触发touchstart
      const touchStartEvent = new TouchEvent('touchstart', {
        touches: [{ clientX: 100, clientY: 100 } as Touch]
      })
      element.dispatchEvent(touchStartEvent)

      // 再触发touchend
      const touchEndEvent = new TouchEvent('touchend')
      element.dispatchEvent(touchEndEvent)

      expect(wrapper.vm.isSwiping).toBe(false)
    })
  })

  describe('滑动检测', () => {
    it('应该检测向右滑动', async () => {
      const onSwipeRight = vi.fn()
      wrapper = mount(createTestComponent({ onSwipeRight }))
      element = wrapper.find('div').element

      // 触发touchstart
      const touchStartEvent = new TouchEvent('touchstart', {
        touches: [{ clientX: 100, clientY: 100 } as Touch]
      })
      element.dispatchEvent(touchStartEvent)

      // 触发touchmove（向右滑动超过阈值）
      const touchMoveEvent = new TouchEvent('touchmove', {
        touches: [{ clientX: 200, clientY: 100 } as Touch]
      })
      element.dispatchEvent(touchMoveEvent)

      expect(wrapper.vm.swipeDirection.right).toBe(true)
      expect(onSwipeRight).toHaveBeenCalled()
      expect(wrapper.vm.isSwiping).toBe(false)
    })

    it('应该检测向左滑动', async () => {
      const onSwipeLeft = vi.fn()
      wrapper = mount(createTestComponent({ onSwipeLeft }))
      element = wrapper.find('div').element

      // 触发touchstart
      const touchStartEvent = new TouchEvent('touchstart', {
        touches: [{ clientX: 200, clientY: 100 } as Touch]
      })
      element.dispatchEvent(touchStartEvent)

      // 触发touchmove（向左滑动超过阈值）
      const touchMoveEvent = new TouchEvent('touchmove', {
        touches: [{ clientX: 100, clientY: 100 } as Touch]
      })
      element.dispatchEvent(touchMoveEvent)

      expect(wrapper.vm.swipeDirection.left).toBe(true)
      expect(onSwipeLeft).toHaveBeenCalled()
    })

    it('应该检测向下滑动', async () => {
      const onSwipeDown = vi.fn()
      wrapper = mount(createTestComponent({ onSwipeDown }))
      element = wrapper.find('div').element

      // 触发touchstart
      const touchStartEvent = new TouchEvent('touchstart', {
        touches: [{ clientX: 100, clientY: 100 } as Touch]
      })
      element.dispatchEvent(touchStartEvent)

      // 触发touchmove（向下滑动超过阈值）
      const touchMoveEvent = new TouchEvent('touchmove', {
        touches: [{ clientX: 100, clientY: 200 } as Touch]
      })
      element.dispatchEvent(touchMoveEvent)

      expect(wrapper.vm.swipeDirection.down).toBe(true)
      expect(onSwipeDown).toHaveBeenCalled()
    })

    it('应该检测向上滑动', async () => {
      const onSwipeUp = vi.fn()
      wrapper = mount(createTestComponent({ onSwipeUp }))
      element = wrapper.find('div').element

      // 触发touchstart
      const touchStartEvent = new TouchEvent('touchstart', {
        touches: [{ clientX: 100, clientY: 200 } as Touch]
      })
      element.dispatchEvent(touchStartEvent)

      // 触发touchmove（向上滑动超过阈值）
      const touchMoveEvent = new TouchEvent('touchmove', {
        touches: [{ clientX: 100, clientY: 100 } as Touch]
      })
      element.dispatchEvent(touchMoveEvent)

      expect(wrapper.vm.swipeDirection.up).toBe(true)
      expect(onSwipeUp).toHaveBeenCalled()
    })
  })

  describe('阈值设置', () => {
    it('应该使用自定义阈值', async () => {
      const onSwipeRight = vi.fn()
      wrapper = mount(createTestComponent({ threshold: 100, onSwipeRight }))
      element = wrapper.find('div').element

      // 触发touchstart
      const touchStartEvent = new TouchEvent('touchstart', {
        touches: [{ clientX: 100, clientY: 100 } as Touch]
      })
      element.dispatchEvent(touchStartEvent)

      // 触发touchmove（滑动距离小于阈值）
      const touchMoveEvent = new TouchEvent('touchmove', {
        touches: [{ clientX: 150, clientY: 100 } as Touch]
      })
      element.dispatchEvent(touchMoveEvent)

      // 不应该触发回调
      expect(onSwipeRight).not.toHaveBeenCalled()

      // 触发touchmove（滑动距离大于阈值）
      const touchMoveEvent2 = new TouchEvent('touchmove', {
        touches: [{ clientX: 250, clientY: 100 } as Touch]
      })
      element.dispatchEvent(touchMoveEvent2)

      // 应该触发回调
      expect(onSwipeRight).toHaveBeenCalled()
    })

    it('应该使用默认阈值（50）', async () => {
      const onSwipeRight = vi.fn()
      wrapper = mount(createTestComponent({ onSwipeRight }))
      element = wrapper.find('div').element

      // 触发touchstart
      const touchStartEvent = new TouchEvent('touchstart', {
        touches: [{ clientX: 100, clientY: 100 } as Touch]
      })
      element.dispatchEvent(touchStartEvent)

      // 触发touchmove（滑动距离等于阈值）
      const touchMoveEvent = new TouchEvent('touchmove', {
        touches: [{ clientX: 151, clientY: 100 } as Touch]
      })
      element.dispatchEvent(touchMoveEvent)

      // 应该触发回调
      expect(onSwipeRight).toHaveBeenCalled()
    })
  })

  describe('事件监听器清理', () => {
    it('卸载时应该移除事件监听器', () => {
      const addEventListenerSpy = vi.spyOn(HTMLElement.prototype, 'addEventListener')
      const removeEventListenerSpy = vi.spyOn(HTMLElement.prototype, 'removeEventListener')

      wrapper = mount(createTestComponent())

      expect(addEventListenerSpy).toHaveBeenCalled()

      wrapper.unmount()

      expect(removeEventListenerSpy).toHaveBeenCalled()

      addEventListenerSpy.mockRestore()
      removeEventListenerSpy.mockRestore()
    })
  })

  describe('边界情况', () => {
    it('滑动距离小于阈值不应该触发', async () => {
      const onSwipeRight = vi.fn()
      wrapper = mount(createTestComponent({ onSwipeRight }))
      element = wrapper.find('div').element

      // 触发touchstart
      const touchStartEvent = new TouchEvent('touchstart', {
        touches: [{ clientX: 100, clientY: 100 } as Touch]
      })
      element.dispatchEvent(touchStartEvent)

      // 触发touchmove（滑动距离小于阈值）
      const touchMoveEvent = new TouchEvent('touchmove', {
        touches: [{ clientX: 120, clientY: 100 } as Touch]
      })
      element.dispatchEvent(touchMoveEvent)

      expect(onSwipeRight).not.toHaveBeenCalled()
      expect(wrapper.vm.isSwiping).toBe(true)
    })

    it('非滑动状态下的touchmove不应该处理', async () => {
      wrapper = mount(createTestComponent())
      element = wrapper.find('div').element

      // 直接触发touchmove（没有先触发touchstart）
      const touchMoveEvent = new TouchEvent('touchmove', {
        touches: [{ clientX: 200, clientY: 100 } as Touch]
      })
      element.dispatchEvent(touchMoveEvent)

      expect(wrapper.vm.isSwiping).toBe(false)
      expect(wrapper.vm.swipeDirection.right).toBe(false)
    })

    it('对角线滑动应该根据主要方向判断', async () => {
      const onSwipeRight = vi.fn()
      const onSwipeDown = vi.fn()
      wrapper = mount(createTestComponent({ onSwipeRight, onSwipeDown }))
      element = wrapper.find('div').element

      // 触发touchstart
      const touchStartEvent = new TouchEvent('touchstart', {
        touches: [{ clientX: 100, clientY: 100 } as Touch]
      })
      element.dispatchEvent(touchStartEvent)

      // 触发touchmove（横向滑动为主）
      const touchMoveEvent = new TouchEvent('touchmove', {
        touches: [{ clientX: 200, clientY: 130 } as Touch]
      })
      element.dispatchEvent(touchMoveEvent)

      // 应该触发横向滑动
      expect(onSwipeRight).toHaveBeenCalled()
      expect(onSwipeDown).not.toHaveBeenCalled()
    })
  })

  describe('多次滑动', () => {
    it('应该能够检测多次滑动', async () => {
      const onSwipeRight = vi.fn()
      wrapper = mount(createTestComponent({ onSwipeRight }))
      element = wrapper.find('div').element

      // 第一次滑动
      let touchStartEvent = new TouchEvent('touchstart', {
        touches: [{ clientX: 100, clientY: 100 } as Touch]
      })
      element.dispatchEvent(touchStartEvent)

      let touchMoveEvent = new TouchEvent('touchmove', {
        touches: [{ clientX: 200, clientY: 100 } as Touch]
      })
      element.dispatchEvent(touchMoveEvent)

      expect(onSwipeRight).toHaveBeenCalledTimes(1)

      // 第二次滑动
      touchStartEvent = new TouchEvent('touchstart', {
        touches: [{ clientX: 100, clientY: 100 } as Touch]
      })
      element.dispatchEvent(touchStartEvent)

      touchMoveEvent = new TouchEvent('touchmove', {
        touches: [{ clientX: 200, clientY: 100 } as Touch]
      })
      element.dispatchEvent(touchMoveEvent)

      expect(onSwipeRight).toHaveBeenCalledTimes(2)
    })
  })
})
