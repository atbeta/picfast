import api from './api'

// --- Types ---

export interface AuthTokens {
  access_token: string
  refresh_token: string
  expires_in: number
  token_type: string
}

export interface UserProfile {
  id: number
  email: string
  name: string
  role: 'admin' | 'user'
  status: number
  capacity_bytes: number
  used_bytes: number
  image_num: number
  album_num: number
  settings: {
    default_album: number
    default_strategy: number
    default_permission: number
  }
  email_verified: boolean
  created_at: string
}

interface ApiResponse<T> {
  status: boolean
  message: string
  data: T
}

// --- API helpers ---

export async function login(email: string, password: string): Promise<AuthTokens> {
  const res = await api.post<ApiResponse<AuthTokens>>('/auth/login', { email, password })
  return res.data.data
}

export async function register(email: string, password: string, name: string): Promise<AuthTokens> {
  const res = await api.post<ApiResponse<AuthTokens>>('/auth/register', { email, password, name })
  return res.data.data
}

export async function logout(): Promise<void> {
  try {
    await api.post('/auth/logout')
  } catch {
    // Swallow — we clear tokens locally regardless
  }
}

export async function getProfile(): Promise<UserProfile> {
  const res = await api.get<ApiResponse<UserProfile>>('/users/me')
  return res.data.data
}

export async function updateProfile(data: {
  name?: string
  password?: string
  settings?: Record<string, unknown>
}): Promise<UserProfile> {
  const res = await api.put<ApiResponse<UserProfile>>('/users/me', data)
  return res.data.data
}

// --- Token helpers ---

export function saveTokens(tokens: AuthTokens) {
  localStorage.setItem('token', tokens.access_token)
  localStorage.setItem('refresh_token', tokens.refresh_token)
}

export function clearTokens() {
  localStorage.removeItem('token')
  localStorage.removeItem('refresh_token')
}

export function hasToken(): boolean {
  return !!localStorage.getItem('token')
}
