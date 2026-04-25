import api from './index'

// Users
export function adminGetUsers(params?: Record<string, string>) {
  return api.get('/admin/users', { params })
}
export function adminGetUser(id: number) {
  return api.get(`/admin/users/${id}`)
}
export function adminUpdateUser(id: number, data: Record<string, unknown>) {
  return api.put(`/admin/users/${id}`, data)
}
export function adminDeleteUser(id: number) {
  return api.delete(`/admin/users/${id}`)
}

// Groups
export function adminGetGroups() {
  return api.get('/admin/groups')
}
export function adminCreateGroup(data: Record<string, unknown>) {
  return api.post('/admin/groups', data)
}
export function adminUpdateGroup(id: number, data: Record<string, unknown>) {
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
  return api.get('/admin/strategies')
}
export function adminCreateStrategy(data: Record<string, unknown>) {
  return api.post('/admin/strategies', data)
}
export function adminUpdateStrategy(id: number, data: Record<string, unknown>) {
  return api.put(`/admin/strategies/${id}`, data)
}
export function adminDeleteStrategy(id: number) {
  return api.delete(`/admin/strategies/${id}`)
}

// Images
export function adminGetImages(params?: Record<string, string>) {
  return api.get('/admin/images', { params })
}
export function adminDeleteImage(id: number) {
  return api.delete(`/admin/images/${id}`)
}

// Settings
export function adminGetSettings() {
  return api.get('/admin/settings')
}
export function adminUpdateSettings(data: Record<string, unknown>) {
  return api.put('/admin/settings', data)
}
