import api from './api'

// --- Shared ---

interface ApiResponse<T> {
  status: boolean
  message: string
  data: T
}

interface PaginatedData<T> {
  items: T[]
  total: number
  page: number
  size: number
}

// ============================================================
// Users
// ============================================================

export interface AdminUser {
  id: number
  email: string
  name: string
  role: string
  group_id: number | null
  capacity_bytes: number
  image_num: number
  album_num: number
  status: number
  email_verified: boolean
  settings: Record<string, unknown>
  created_at: string
}

export async function listAdminUsers(params?: {
  page?: number
  page_size?: number
  keyword?: string
  status?: number
}): Promise<PaginatedData<AdminUser>> {
  const res = await api.get<ApiResponse<PaginatedData<AdminUser>>>('/admin/users', { params })
  return res.data.data
}

export async function updateAdminUser(
  id: number,
  data: {
    name?: string
    password?: string
    group_id?: number | null
    capacity_bytes?: number
    status?: number
  },
): Promise<AdminUser> {
  const res = await api.put<ApiResponse<AdminUser>>(`/admin/users/${id}`, data)
  return res.data.data
}

export async function deleteAdminUser(id: number): Promise<void> {
  await api.delete(`/admin/users/${id}`)
}

// ============================================================
// Groups
// ============================================================

export interface AdminGroup {
  id: number
  name: string
  is_default: boolean
  is_guest: boolean
  configs: Record<string, unknown>
  strategy_ids: number[]
  created_at: string
  updated_at: string
}

export async function listAdminGroups(): Promise<AdminGroup[]> {
  const res = await api.get<ApiResponse<AdminGroup[]>>('/admin/groups')
  return res.data.data
}

export async function createAdminGroup(data: {
  name: string
  is_default?: boolean
  is_guest?: boolean
  configs?: Record<string, unknown>
}): Promise<AdminGroup> {
  const res = await api.post<ApiResponse<AdminGroup>>('/admin/groups', data)
  return res.data.data
}

export async function updateAdminGroup(
  id: number,
  data: { name: string; is_default: boolean; is_guest: boolean; configs?: Record<string, unknown> },
): Promise<AdminGroup> {
  const res = await api.put<ApiResponse<AdminGroup>>(`/admin/groups/${id}`, data)
  return res.data.data
}

export async function deleteAdminGroup(id: number): Promise<void> {
  await api.delete(`/admin/groups/${id}`)
}

// ============================================================
// Strategies
// ============================================================

export interface AdminStrategy {
  id: number
  name: string
  strategy_type: string
  configs: Record<string, unknown>
  created_at: string
  updated_at: string
}

export async function listAdminStrategies(): Promise<AdminStrategy[]> {
  const res = await api.get<ApiResponse<AdminStrategy[]>>('/admin/strategies')
  return res.data.data
}

export async function createAdminStrategy(data: {
  name: string
  strategy_type: string
  configs: Record<string, unknown>
}): Promise<AdminStrategy> {
  const res = await api.post<ApiResponse<AdminStrategy>>('/admin/strategies', data)
  return res.data.data
}

export async function updateAdminStrategy(
  id: number,
  data: { name?: string; strategy_type?: string; configs?: Record<string, unknown> },
): Promise<AdminStrategy> {
  const res = await api.put<ApiResponse<AdminStrategy>>(`/admin/strategies/${id}`, data)
  return res.data.data
}

export async function deleteAdminStrategy(id: number): Promise<void> {
  await api.delete(`/admin/strategies/${id}`)
}

// ============================================================
// Images (admin)
// ============================================================

export interface AdminImage {
  id: number
  user_id: number | null
  origin_name: string
  size_bytes: number
  extension: string
  width: number
  height: number
  permission: number
  user_email: string | null
  url: string
  thumbnail_url: string
  created_at: string
}

export async function listAdminImages(params?: {
  page?: number
  page_size?: number
  keyword?: string
  email?: string
  extension?: string
}): Promise<PaginatedData<AdminImage>> {
  const res = await api.get<ApiResponse<PaginatedData<AdminImage>>>('/admin/images', { params })
  return res.data.data
}

export async function deleteAdminImage(id: number): Promise<void> {
  await api.delete(`/admin/images/${id}`)
}

// ============================================================
// Settings
// ============================================================

export interface AdminSettings {
  app_name: string
  allow_guest_upload: boolean
  allow_registration: boolean
  user_initial_capacity: number
  moderation_mode: string
}

export async function getAdminSettings(): Promise<AdminSettings> {
  const res = await api.get<ApiResponse<AdminSettings>>('/admin/settings')
  return res.data.data
}

export async function updateAdminSettings(data: Partial<AdminSettings>): Promise<AdminSettings> {
  const res = await api.put<ApiResponse<AdminSettings>>('/admin/settings', data)
  return res.data.data
}
