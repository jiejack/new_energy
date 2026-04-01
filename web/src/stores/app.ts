import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAppStore = defineStore('app', () => {
  // 状态
  const sidebar = ref({
    opened: localStorage.getItem('sidebarStatus') !== 'closed',
    withoutAnimation: false,
  })
  const device = ref<'desktop' | 'mobile'>('desktop')
  const size = ref<'default' | 'small' | 'large'>(
    (localStorage.getItem('size') as 'default' | 'small' | 'large') || 'default'
  )
  const theme = ref<'light' | 'dark'>(
    (localStorage.getItem('theme') as 'light' | 'dark') || 'light'
  )
  const loading = ref(false)

  // 计算属性
  const sidebarOpened = computed(() => sidebar.value.opened)

  /**
   * 切换侧边栏
   */
  function toggleSidebar(withoutAnimation = false) {
    sidebar.value.opened = !sidebar.value.opened
    sidebar.value.withoutAnimation = withoutAnimation
    if (sidebar.value.opened) {
      localStorage.setItem('sidebarStatus', 'opened')
    } else {
      localStorage.setItem('sidebarStatus', 'closed')
    }
  }

  /**
   * 关闭侧边栏
   */
  function closeSidebar(withoutAnimation = false) {
    sidebar.value.opened = false
    sidebar.value.withoutAnimation = withoutAnimation
    localStorage.setItem('sidebarStatus', 'closed')
  }

  /**
   * 切换设备类型
   */
  function toggleDevice(val: 'desktop' | 'mobile') {
    device.value = val
  }

  /**
   * 设置元素大小
   */
  function setSize(val: 'default' | 'small' | 'large') {
    size.value = val
    localStorage.setItem('size', val)
  }

  /**
   * 切换主题
   */
  function toggleTheme() {
    theme.value = theme.value === 'light' ? 'dark' : 'light'
    localStorage.setItem('theme', theme.value)
    document.documentElement.setAttribute('data-theme', theme.value)
  }

  /**
   * 设置加载状态
   */
  function setLoading(val: boolean) {
    loading.value = val
  }

  return {
    // 状态
    sidebar,
    device,
    size,
    theme,
    loading,
    // 计算属性
    sidebarOpened,
    // 方法
    toggleSidebar,
    closeSidebar,
    toggleDevice,
    setSize,
    toggleTheme,
    setLoading,
  }
})
