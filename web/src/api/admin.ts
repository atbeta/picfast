import api from './index'
import type { AdminUser, AdminGroup, AdminStrategy, Settings, ApiResult, PaginatedData } from './types'

// Users
export function adminGetUsers(page = 1, pageSize = 20) {
  return api.get<ApiResult<PaginatedData<AdminUser>>>('/admin/users', { params: { page, page_size: pageSize } })
}
export function adminGetUser(id: number) {
  return api.get<ApiResult<AdminUser>>(`/admin/users/${id}`)
}
export function adminUpdateUser(id: number, data: Partial<Pick<AdminUser, 'name' | 'status' | 'capacity_bytes'>>) {
  return api.put(`/admin/users/${id}`, data)
}
export function adminDeleteUser(id: number) {
  return api.delete(`/admin/users/${id}`)
}

// Groups
export function adminGetGroups() {
  return api.get<ApiResult<AdminGroup[]>>('/admin/groups')
}
export function adminCreateGroup(data: { name: string; is_default?: boolean; is_guest?: boolean; configs?: object }) {
  return api.post<ApiResult<AdminGroup>>('/admin/groups', data)
}
export function adminUpdateGroup(id: number, data: Partial<Pick<AdminGroup, 'name' | 'is_default' | 'is_guest'>>) {
  return api.put(`/admin/groups/${id}`, data)
}
export function adminDeleteGroup(id: number) {
  return api.delete(`/admin/groups/${id}`)
}
export function adminSetGroupStrategies(id: number, strategyIds: number[]) {
  return api.put(`/admin/groups/${id}/strategies`, { strategy_ids: strategyIds })
}

// Strategies
export function adminGetStrategies() {
  return api.get<ApiResult<AdminStrategy[]>>('/admin/strategies')
}
export function adminCreateStrategy(data: { name: string; strategy_type: string; configs: object }) {
  return api.post<ApiResult<AdminStrategy>>('/admin/strategies', data)
}
export function adminUpdateStrategy(id: number, data: Partial<Pick<AdminStrategy, 'name' | 'strategy_type'>>) {
  return api.put(`/admin/strategies/${id}`, data)
}
export function adminDeleteStrategy(id: number) {
  return api.delete(`/admin/strategies/${id}`)
}

// Images
export function adminGetImages(page = 1, pageSize = 20) {
  return api.get<ApiResult<PaginatedData<unknown>>>('/admin/images', { params: { page, page_size: pageSize } })
}
export function adminDeleteImage(id: number) {
  return api.delete(`/admin/images/${id}`)
}

// Settings
export function adminGetSettings() {
  return api.get<ApiResult<Settings>>('/admin/settings')
}
export function adminUpdateSettings(data: Partial<Settings>) {
  return api.put('/admin/settings', data)
}
