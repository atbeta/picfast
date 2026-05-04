import api, { type ApiResponse } from './api'
import { logError } from './error-handler'

// --- Types ---

export interface AuthTokens {
  access_token: string
  refresh_token: string
  expires_in: number
  token_type: string
}

export interface RegisterResult {
  requires_email_verification: boolean
  verification_email_sent: boolean
  tokens?: AuthTokens
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
    default_album?: number
    default_strategy?: number
    default_permission?: number
    image_processing?: {
      image_save_quality?: number
      image_save_format?: string
      is_strip_exif?: boolean
      is_enable_watermark?: boolean
      watermark_configs?: {
        text?: string
        position?: string
        font_size?: number
        color?: string
        opacity?: number
      }
    }
  }
  email_verified: boolean
  created_at: string
}

// --- API helpers ---

export async function login(email: string, password: string): Promise<AuthTokens> {
  const res = await api.post<ApiResponse<AuthTokens>>('/auth/login', { email, password })
  return res.data.data
}

export async function register(email: string, password: string, name: string, language?: string): Promise<RegisterResult> {
  const res = await api.post<ApiResponse<RegisterResult>>('/auth/register', { email, password, name, language })
  return res.data.data
}

export async function createSetupAdmin(email: string, password: string, name: string): Promise<AuthTokens> {
  const res = await api.post<ApiResponse<AuthTokens>>('/setup/admin', { email, password, name })
  return res.data.data
}

export async function resendVerification(email: string, language?: string): Promise<void> {
  await api.post('/auth/resend-verification', { email, language })
}

export async function forgotPassword(email: string, language?: string): Promise<number | null> {
  const res = await api.post('/auth/forgot-password', { email, language })
  const cooldown = Number.parseInt(String(res.headers['x-cooldown-seconds'] ?? ''), 10)
  return Number.isFinite(cooldown) && cooldown > 0 ? cooldown : null
}

export async function resetPassword(token: string, newPassword: string): Promise<void> {
  await api.post('/auth/reset-password', { token, new_password: newPassword })
}

export async function verifyEmail(token: string): Promise<void> {
  await api.post('/auth/verify-email', { token })
}

export async function logout(): Promise<void> {
  try {
    await api.post('/auth/logout')
  } catch (err: unknown) {
    logError('auth.logout', err)
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
  document.cookie = `picfast_token=${tokens.access_token}; path=/; max-age=${tokens.expires_in}`
}

export function clearTokens() {
  localStorage.removeItem('token')
  localStorage.removeItem('refresh_token')
  document.cookie = 'picfast_token=; path=/; max-age=0'
}

export function syncTokenToCookie() {
  const token = localStorage.getItem('token')
  if (token && !document.cookie.includes('picfast_token=')) {
    // 30 days default if we don't have the exact expires_in
    document.cookie = `picfast_token=${token}; path=/; max-age=2592000`
  }
}

export function hasToken(): boolean {
  return !!localStorage.getItem('token')
}
