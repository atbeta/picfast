import api from './index'

export function login(email: string, password: string) {
  return api.post('/auth/login', { email, password })
}

export function register(email: string, password: string, name: string) {
  return api.post('/auth/register', { email, password, name })
}

export function refreshToken(refreshToken: string) {
  return api.post('/auth/refresh', { refresh_token: refreshToken })
}

export function logout() {
  return api.post('/auth/logout')
}
