import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useLoadingStore = defineStore('loading', () => {
  // 全局加载状态
  const globalLoading = ref(false)
  const loadingText = ref('加载中...')
  
  // 页面级加载状态
  const pageLoading = ref(false)
  const pageLoadingText = ref('页面加载中...')
  
  // 操作加载状态 - 用于单个操作
  const actionLoadings = ref<Record<string, boolean>>({})
  
  // 计算属性
  const isLoading = computed(() => globalLoading.value || pageLoading.value)
  
  // 设置全局加载
  function setGlobalLoading(loading: boolean, text = '加载中...') {
    globalLoading.value = loading
    loadingText.value = text
  }
  
  // 设置页面加载
  function setPageLoading(loading: boolean, text = '页面加载中...') {
    pageLoading.value = loading
    pageLoadingText.value = text
  }
  
  // 设置操作加载
  function setActionLoading(actionId: string, loading: boolean) {
    actionLoadings.value[actionId] = loading
  }
  
  // 获取操作加载状态
  function getActionLoading(actionId: string): boolean {
    return actionLoadings.value[actionId] || false
  }
  
  // 清除所有加载状态
  function clearAll() {
    globalLoading.value = false
    pageLoading.value = false
    actionLoadings.value = {}
  }
  
  return {
    globalLoading,
    loadingText,
    pageLoading,
    pageLoadingText,
    actionLoadings,
    isLoading,
    setGlobalLoading,
    setPageLoading,
    setActionLoading,
    getActionLoading,
    clearAll,
  }
})
