const TokenKey = 'nem_token'
const RefreshTokenKey = 'nem_refresh_token'

/**
 * 获取Token
 */
export function getToken(): string | null {
  return localStorage.getItem(TokenKey)
}

/**
 * 设置Token
 */
export function setToken(token: string): void {
  localStorage.setItem(TokenKey, token)
}

/**
 * 移除Token
 */
export function removeToken(): void {
  localStorage.removeItem(TokenKey)
}

/**
 * 获取RefreshToken
 */
export function getRefreshToken(): string | null {
  return localStorage.getItem(RefreshTokenKey)
}

/**
 * 设置RefreshToken
 */
export function setRefreshToken(token: string): void {
  localStorage.setItem(RefreshTokenKey, token)
}

/**
 * 移除RefreshToken
 */
export function removeRefreshToken(): void {
  localStorage.removeItem(RefreshTokenKey)
}

/**
 * 清除所有认证信息
 */
export function clearAuth(): void {
  removeToken()
  removeRefreshToken()
}
