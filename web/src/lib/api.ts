import axios from 'axios'

export const ACCESS_TOKEN_KEY = 'token'
export const REFRESH_TOKEN_KEY = 'refresh_token'

export interface ApiResponse<T> {
  status: boolean
  message: string
  data: T
}

export interface PaginatedData<T> {
  items: T[]
  total: number
  page: number
  size: number
}

const api = axios.create({
  baseURL: '/api/v1',
  headers: { 'Content-Type': 'application/json' },
})

function syncAuthCookie(token: string) {
  document.cookie = `picfast_token=${token}; path=/; max-age=2592000`
}

api.interceptors.request.use((config) => {
  const token = localStorage.getItem(ACCESS_TOKEN_KEY)
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
    // Keep cookie in sync so private image/thumbnail URLs can authenticate.
    syncAuthCookie(token)
  }
  return config
})

let isRefreshing = false
let subscribers: Array<(token: string) => void> = []

function flushSubscribers(token: string) {
  subscribers.forEach((callback) => callback(token))
  subscribers = []
}

function subscribe(callback: (token: string) => void) {
  subscribers.push(callback)
}

api.interceptors.response.use(
  (res) => res,
  async (err) => {
    const originalRequest = err.config
    const isAuthEndpoint = originalRequest?.url?.startsWith('/auth/login') || originalRequest?.url?.startsWith('/auth/register')
    if (err.response?.status === 401 && originalRequest && !originalRequest._retry && !isAuthEndpoint) {
      const refreshToken = localStorage.getItem(REFRESH_TOKEN_KEY)
      if (!refreshToken) {
        localStorage.removeItem(ACCESS_TOKEN_KEY)
        localStorage.removeItem(REFRESH_TOKEN_KEY)
        window.location.href = '/login'
        return Promise.reject(err)
      }

      originalRequest._retry = true
      if (isRefreshing) {
        return new Promise((resolve) => {
          subscribe((nextToken) => {
            originalRequest.headers.Authorization = `Bearer ${nextToken}`
            resolve(api(originalRequest))
          })
        })
      }

      isRefreshing = true
      try {
        const res = await api.post('/auth/refresh', { refresh_token: refreshToken })
        const { access_token, refresh_token } = res.data.data
        localStorage.setItem(ACCESS_TOKEN_KEY, access_token)
        localStorage.setItem(REFRESH_TOKEN_KEY, refresh_token)
        syncAuthCookie(access_token)
        api.defaults.headers.common.Authorization = `Bearer ${access_token}`
        flushSubscribers(access_token)
        originalRequest.headers.Authorization = `Bearer ${access_token}`
        return api(originalRequest)
      } catch (refreshError) {
        localStorage.removeItem(ACCESS_TOKEN_KEY)
        localStorage.removeItem(REFRESH_TOKEN_KEY)
        window.location.href = '/login'
        return Promise.reject(refreshError)
      } finally {
        isRefreshing = false
      }
    }

    return Promise.reject(err)
  },
)

export default api
