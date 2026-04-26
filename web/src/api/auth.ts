import api from './index'
import type { AuthTokens, ApiResult } from './types'

export function login(email: string, password: string) {
  return api.post<ApiResult<AuthTokens>>('/auth/login', { email, password })
}

export function register(email: string, password: string, name: string) {
  return api.post<ApiResult<AuthTokens>>('/auth/register', { email, password, name })
}

export function refreshToken(refreshToken: string) {
  return api.post<ApiResult<AuthTokens>>('/auth/refresh', { refresh_token: refreshToken })
}

export function logout() {
  return api.post('/auth/logout')
}
