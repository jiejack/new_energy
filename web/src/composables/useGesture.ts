import { ref, onMounted, onUnmounted, type Ref } from 'vue'

export interface SwipeDirection {
  left: boolean
  right: boolean
  up: boolean
  down: boolean
}

export interface GestureOptions {
  threshold?: number
  onSwipeLeft?: () => void
  onSwipeRight?: () => void
  onSwipeUp?: () => void
  onSwipeDown?: () => void
}

/**
 * 手势支持Hook
 */
export function useGesture(
  element: Ref<HTMLElement | null>,
  options: GestureOptions = {}
) {
  const {
    threshold = 50,
    onSwipeLeft,
    onSwipeRight,
    onSwipeUp,
    onSwipeDown
  } = options

  const startX = ref(0)
  const startY = ref(0)
  const isSwiping = ref(false)
  const swipeDirection = ref<SwipeDirection>({
    left: false,
    right: false,
    up: false,
    down: false
  })

  const handleTouchStart = (e: TouchEvent) => {
    startX.value = e.touches[0].clientX
    startY.value = e.touches[0].clientY
    isSwiping.value = true
    swipeDirection.value = { left: false, right: false, up: false, down: false }
  }

  const handleTouchMove = (e: TouchEvent) => {
    if (!isSwiping.value) return

    const currentX = e.touches[0].clientX
    const currentY = e.touches[0].clientY
    const diffX = currentX - startX.value
    const diffY = currentY - startY.value

    // 横向滑动
    if (Math.abs(diffX) > Math.abs(diffY)) {
      if (Math.abs(diffX) > threshold) {
        if (diffX > 0) {
          swipeDirection.value.right = true
          onSwipeRight?.()
        } else {
          swipeDirection.value.left = true
          onSwipeLeft?.()
        }
        isSwiping.value = false
      }
    } else {
      // 纵向滑动
      if (Math.abs(diffY) > threshold) {
        if (diffY > 0) {
          swipeDirection.value.down = true
          onSwipeDown?.()
        } else {
          swipeDirection.value.up = true
          onSwipeUp?.()
        }
        isSwiping.value = false
      }
    }
  }

  const handleTouchEnd = () => {
    isSwiping.value = false
  }

  onMounted(() => {
    if (element.value) {
      element.value.addEventListener('touchstart', handleTouchStart, { passive: true })
      element.value.addEventListener('touchmove', handleTouchMove, { passive: true })
      element.value.addEventListener('touchend', handleTouchEnd, { passive: true })
    }
  })

  onUnmounted(() => {
    if (element.value) {
      element.value.removeEventListener('touchstart', handleTouchStart)
      element.value.removeEventListener('touchmove', handleTouchMove)
      element.value.removeEventListener('touchend', handleTouchEnd)
    }
  })

  return {
    isSwiping,
    swipeDirection
  }
}
