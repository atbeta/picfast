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
  total_pages: number
}

const api = axios.create({
  baseURL: '/api/v1',
  headers: { 'Content-Type': 'application/json' },
  withCredentials: true,
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem(ACCESS_TOKEN_KEY)
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

let isRefreshing = false
type RefreshSubscriber = {
  onSuccess: (token: string) => void
  onError: (error: unknown) => void
}

let subscribers: RefreshSubscriber[] = []

function flushSubscribers(token: string) {
  subscribers.forEach((subscriber) => subscriber.onSuccess(token))
  subscribers = []
}

function rejectSubscribers(error: unknown) {
  subscribers.forEach((subscriber) => subscriber.onError(error))
  subscribers = []
}

function subscribe(onSuccess: (token: string) => void, onError: (error: unknown) => void) {
  subscribers.push({ onSuccess, onError })
}

api.interceptors.response.use(
  (res) => res,
  async (err) => {
    const originalRequest = err.config
    const isAuthEndpoint =
      originalRequest?.url?.startsWith('/auth/login') ||
      originalRequest?.url?.startsWith('/auth/register') ||
      originalRequest?.url?.startsWith('/auth/refresh')
    if (err.response?.status === 401 && originalRequest && !originalRequest._retry && !isAuthEndpoint) {
      const refreshToken = localStorage.getItem(REFRESH_TOKEN_KEY)

      if (refreshToken) {
        originalRequest._retry = true
        if (isRefreshing) {
          return new Promise((resolve, reject) => {
            subscribe(
              (nextToken) => {
                originalRequest.headers.Authorization = `Bearer ${nextToken}`
                resolve(api(originalRequest))
              },
              (refreshError) => {
                reject(refreshError)
              },
            )
          })
        }

        isRefreshing = true
        try {
          const res = await api.post('/auth/refresh', { refresh_token: refreshToken })
          const { access_token, refresh_token } = res.data.data
          localStorage.setItem(ACCESS_TOKEN_KEY, access_token)
          localStorage.setItem(REFRESH_TOKEN_KEY, refresh_token)
          api.defaults.headers.common.Authorization = `Bearer ${access_token}`
          flushSubscribers(access_token)
          originalRequest.headers.Authorization = `Bearer ${access_token}`
          return api(originalRequest)
        } catch (refreshError) {
          localStorage.removeItem(ACCESS_TOKEN_KEY)
          localStorage.removeItem(REFRESH_TOKEN_KEY)
          rejectSubscribers(refreshError)
          window.location.href = '/login'
          return Promise.reject(refreshError)
        } finally {
          isRefreshing = false
        }
      }

      // No localStorage token — try cookie-based refresh (OAuth path)
      originalRequest._retry = true
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          subscribe(
            (nextToken) => {
              originalRequest.headers.Authorization = `Bearer ${nextToken}`
              resolve(api(originalRequest))
            },
            (refreshError) => {
              reject(refreshError)
            },
          )
        })
      }

      isRefreshing = true
      try {
        const res = await api.post('/auth/refresh')
        const { access_token, refresh_token } = res.data.data
        localStorage.setItem(ACCESS_TOKEN_KEY, access_token)
        localStorage.setItem(REFRESH_TOKEN_KEY, refresh_token)
        api.defaults.headers.common.Authorization = `Bearer ${access_token}`
        flushSubscribers(access_token)
        originalRequest.headers.Authorization = `Bearer ${access_token}`
        return api(originalRequest)
      } catch (refreshError) {
        rejectSubscribers(refreshError)
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
