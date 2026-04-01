import { ref, onMounted, onUnmounted } from 'vue'
import { debounce } from 'lodash-es'

/**
 * 检测是否为移动设备
 */
export function isMobile(): boolean {
  return window.innerWidth <= 768
}

/**
 * 检测是否为触摸设备
 */
export function isTouchDevice(): boolean {
  return 'ontouchstart' in window || navigator.maxTouchPoints > 0
}

/**
 * 获取设备类型
 */
export function getDeviceType(): 'mobile' | 'tablet' | 'desktop' {
  const width = window.innerWidth
  if (width <= 768) return 'mobile'
  if (width <= 992) return 'tablet'
  return 'desktop'
}

/**
 * 响应式设备检测Hook
 */
export function useDevice() {
  const isMobileDevice = ref(isMobile())
  const isTouch = ref(isTouchDevice())
  const deviceType = ref(getDeviceType())

  const handleResize = debounce(() => {
    isMobileDevice.value = isMobile()
    deviceType.value = getDeviceType()
  }, 100)

  onMounted(() => {
    window.addEventListener('resize', handleResize)
  })

  onUnmounted(() => {
    window.removeEventListener('resize', handleResize)
  })

  return {
    isMobile: isMobileDevice,
    isTouch,
    deviceType
  }
}

/**
 * 获取安全区域插入值
 */
export function getSafeAreaInset(): { top: number; bottom: number; left: number; right: number } {
  const style = getComputedStyle(document.documentElement)
  return {
    top: parseInt(style.getPropertyValue('--safe-area-inset-top') || '0'),
    bottom: parseInt(style.getPropertyValue('--safe-area-inset-bottom') || '0'),
    left: parseInt(style.getPropertyValue('--safe-area-inset-left') || '0'),
    right: parseInt(style.getPropertyValue('--safe-area-inset-right') || '0')
  }
}
