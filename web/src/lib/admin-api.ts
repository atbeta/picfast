import api, { type ApiResponse, type PaginatedData } from './api'
import type { ThemeConfig } from './theme-config'

// --- Shared ---

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
  used_capacity: number
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
  user_count: number
  image_count: number
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

export async function setAdminGroupStrategies(id: number, strategyIds: number[]): Promise<void> {
  await api.put(`/admin/groups/${id}/strategies`, { strategy_ids: strategyIds })
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
  key: string
  user_id: number | null
  origin_name: string
  size_bytes: number
  extension: string
  width: number
  height: number
  permission: number
  user_email: string | null
  uploaded_ip: string
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
  date_from?: string
  date_to?: string
  tag_ids?: string
}): Promise<PaginatedData<AdminImage>> {
  const res = await api.get<ApiResponse<PaginatedData<AdminImage>>>('/admin/images', { params })
  return res.data.data
}

export async function deleteAdminImage(id: number): Promise<void> {
  await api.delete(`/admin/images/${id}`)
}

// ============================================================
// Moderation (admin)
// ============================================================

export interface AdminModerationImage {
  id: number
  key: string
  origin_name: string
  size_bytes: number
  mimetype: string
  extension: string
  width: number
  height: number
  permission: number
  moderation_status: string
  url: string
  thumbnail_url: string
  created_at: string
}

export async function listPendingModerationImages(params?: {
  page?: number
  page_size?: number
}): Promise<PaginatedData<AdminModerationImage>> {
  const res = await api.get<ApiResponse<PaginatedData<AdminModerationImage>>>('/admin/moderation/pending', { params })
  return res.data.data
}

export async function approveModerationImage(id: number): Promise<void> {
  await api.post(`/admin/moderation/${id}/approve`)
}

export async function rejectModerationImage(id: number, reason?: string): Promise<void> {
  await api.post(`/admin/moderation/${id}/reject`, reason ? { reason } : {})
}

// ============================================================
// Settings
// ============================================================

export interface AdminSettings {
  app_name: string
  app_url: string
  site_description: string
  favicon_url: string
  allow_guest_upload: boolean
  guest_capacity_bytes: number
  allow_registration: boolean
  allow_oauth_registration: boolean
  allow_user_image_processing: boolean
  skip_image_processing: boolean
  require_email_verification: boolean
  email_verification_ready: boolean
  user_initial_capacity: number
  default_image_ttl: string
  guest_image_ttl: string
  moderation_mode: string
  footer_text_1: string
  footer_link_1: string
  footer_text_2: string
  footer_link_2: string
  show_powered_by: boolean
  guest_upload_notice_title: string
  guest_upload_notice_subtitle: string
  show_login_link: boolean
  analytics_provider: string
  analytics_config: Record<string, unknown>
  theme_config: ThemeConfig
  default_copy_format: string
  copy_template: string
}

export async function getAdminSettings(): Promise<AdminSettings> {
  const res = await api.get<ApiResponse<AdminSettings>>('/admin/settings')
  return res.data.data
}

export async function updateAdminSettings(data: Partial<AdminSettings>): Promise<AdminSettings> {
  const res = await api.put<ApiResponse<AdminSettings>>('/admin/settings', data)
  return res.data.data
}

export interface AdminAuditLog {
  id: number
  actor_user_id: number | null
  actor_email: string | null
  action: string
  resource_type: string
  resource_id: string | null
  resource_name: string | null
  details: Record<string, unknown>
  ip: string
  user_agent: string
  created_at: string
}

export async function listAdminAuditLogs(params?: {
  page?: number
  page_size?: number
  action?: string
  resource_type?: string
}): Promise<PaginatedData<AdminAuditLog>> {
  const res = await api.get<ApiResponse<PaginatedData<AdminAuditLog>>>('/admin/audit-logs', { params })
  return res.data.data
}

// ============================================================
// Observability
// ============================================================

export interface AdminHealthItem {
  healthy: boolean
  status?: 'healthy' | 'unhealthy' | 'disabled'
  ready?: boolean
  configured?: boolean
  path?: string
  error?: string
  warning?: string
}

export interface AdminStorageStrategyHealth {
  id: number
  name: string
  type: string
  healthy: boolean
  error?: string
  warning?: string
}

export interface AdminObservabilitySummary {
  generated_at: string
  uptime_seconds: number
  health: {
    database: AdminHealthItem
    uploads: AdminHealthItem
    thumbnails: AdminHealthItem
    mail: AdminHealthItem
  }
  runtime: {
    go_version: string
    goos: string
    goarch: string
    num_cpu: number
    goroutines: number
    memory_alloc_bytes: number
    memory_sys_bytes: number
  }
  database: {
    total_connections: number
    acquired_connections: number
    idle_connections: number
    max_connections: number
  }
  usage: {
    users_total: number
    images_total: number
    storage_bytes: number
    uploads_24h: number
    pending_moderation: number
    audit_logs_24h: number
  }
  storage_strategies: AdminStorageStrategyHealth[]
  config: {
    metrics_enabled: boolean
    pprof_enabled: boolean
    moderation_mode: string
    audit_upload_logs: boolean
  }
}

export interface MaintenanceSummaryStrategy {
  id: number
  name: string
  type: string
  healthy: boolean
  error?: string
  warning?: string
}

export interface MaintenanceSummary {
  generated_at: string
  risks: MaintenanceRisk[]
  storage: { disk: DiskInfo; strategies: MaintenanceSummaryStrategy[] }
  usage: Record<string, number>
  backup: BackupInfo
  database: { table: string; rows: number }[]
  phash_coverage: { total: number; with_phash: number }
  thumbnails: { on_disk: number; dir: string; error?: string }
}

export interface MaintenanceRisk {
  level: 'info' | 'warn' | 'error'
  code: string
  message: string
  count?: number
}

export interface DiskInfo {
  healthy: boolean
  path: string
  total_bytes: number
  free_bytes: number
}

export interface BackupInfo {
  status: 'ok' | 'no_backups' | 'no_storage'
  file?: string
  size?: number
  timestamp?: string
  path?: string
}

export async function getAdminObservabilitySummary(): Promise<AdminObservabilitySummary> {
  const res = await api.get<ApiResponse<AdminObservabilitySummary>>('/admin/observability/summary')
  return res.data.data
}

export async function getMaintenanceSummary(): Promise<MaintenanceSummary> {
  const res = await api.get<ApiResponse<MaintenanceSummary>>('/admin/maintenance/summary')
  return res.data.data
}

export async function cleanupExpiredImages(): Promise<string> {
  const res = await api.post<ApiResponse<string>>('/admin/maintenance/cleanup-expired')
  return res.data.message ?? 'done'
}

export async function recalcPHash(): Promise<string> {
  const res = await api.post<ApiResponse<string>>('/admin/maintenance/recalc-phash')
  return res.data.message ?? 'started'
}

