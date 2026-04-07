import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { UserInfo, LoginForm } from '@/types'
import { login, logout, getUserInfo } from '@/api/auth'
import { getToken, setToken, removeToken, getRefreshToken, setRefreshToken, removeRefreshToken } from '@/utils/auth'

export const useUserStore = defineStore('user', () => {
  // 状态
  const token = ref<string>(getToken() || '')
  const refreshToken = ref<string>(getRefreshToken() || '')
  const userInfo = ref<UserInfo | null>(null)
  const roles = ref<string[]>([])
  const permissions = ref<string[]>([])

  // 计算属性
  const isLoggedIn = computed(() => !!token.value)
  const username = computed(() => userInfo.value?.username || '')
  const nickname = computed(() => userInfo.value?.nickname || '')
  const avatar = computed(() => userInfo.value?.avatar || '')

  /**
   * 登录
   */
  async function loginAction(loginForm: LoginForm) {
    try {
      const response = await login(loginForm)
      const { token: accessToken, refreshToken: refreshTokenValue, user } = response

      // 保存token
      token.value = accessToken
      refreshToken.value = refreshTokenValue
      setToken(accessToken)
      setRefreshToken(refreshTokenValue)

      // 保存用户信息
      userInfo.value = user
      roles.value = user.roles || []
      permissions.value = user.permissions || []

      return response
    } catch (error) {
      throw error
    }
  }

  /**
   * 获取用户信息
   */
  async function getUserInfoAction() {
    try {
      const response = await getUserInfo()
      const { roles: userRoles, permissions: userPermissions, ...info } = response

      userInfo.value = { ...info, roles: userRoles || [], permissions: userPermissions || [] }
      roles.value = userRoles || []
      permissions.value = userPermissions || []

      return {
        roles: userRoles,
        permissions: userPermissions,
      }
    } catch (error) {
      throw error
    }
  }

  /**
   * 登出
   */
  async function logoutAction() {
    try {
      await logout()
    } finally {
      // 清除状态
      resetState()
    }
  }

  /**
   * 重置状态
   */
  function resetState() {
    token.value = ''
    refreshToken.value = ''
    userInfo.value = null
    roles.value = []
    permissions.value = []
    removeToken()
    removeRefreshToken()
  }

  /**
   * 更新token
   */
  function updateToken(newToken: string, newRefreshToken: string) {
    token.value = newToken
    refreshToken.value = newRefreshToken
    setToken(newToken)
    setRefreshToken(newRefreshToken)
  }

  return {
    // 状态
    token,
    refreshToken,
    userInfo,
    roles,
    permissions,
    // 计算属性
    isLoggedIn,
    username,
    nickname,
    avatar,
    // 方法
    loginAction,
    getUserInfoAction,
    logoutAction,
    resetState,
    updateToken,
  }
})
