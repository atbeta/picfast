import api, { type ApiResponse, type PaginatedData } from './api'

// --- Shared types ---

// ============================================================
// Images
// ============================================================

export interface ImageItem {
  id: number
  key: string
  origin_name: string
  size_bytes: number
  mimetype: string
  extension: string
  width: number
  height: number
  permission: number
  album_id: number | null
  url: string
  thumbnail_url: string
  moderation_status: string
  strategy_id: number | null
  strategy_name: string
  strategy_type: string
  links: { url: string; html: string; bbcode: string; markdown: string; thumbnail_url: string }
  created_at: string
}

export async function listImages(page = 1, pageSize = 20, albumId?: number | null, keyword?: string): Promise<PaginatedData<ImageItem>> {
  const params: Record<string, string | number> = { page, page_size: pageSize }
  if (albumId) {
    params.album_id = albumId
  }
  if (keyword) {
    params.keyword = keyword
  }
  const res = await api.get<ApiResponse<PaginatedData<ImageItem>>>('/images', { params })
  return res.data.data
}

export async function deleteImage(key: string): Promise<void> {
  await api.delete(`/images/${encodeURIComponent(key)}`)
}

export async function batchDeleteImages(keys: string[]): Promise<{ deleted: number; failed: number }> {
  const res = await api.post<ApiResponse<{ deleted: number; failed: number }>>('/images/batch-delete', { keys })
  return res.data.data
}

export async function getImage(key: string): Promise<ImageItem> {
  const res = await api.get<ApiResponse<ImageItem>>(`/images/${encodeURIComponent(key)}`)
  return res.data.data
}

export async function updateImage(key: string, data: { album_id?: number; permission?: number }): Promise<ImageItem> {
  const res = await api.patch<ApiResponse<ImageItem>>(`/images/${encodeURIComponent(key)}`, data)
  return res.data.data
}

export async function uploadImageAuth(
  file: File,
  options?: { album_id?: number; strategy_id?: number; permission?: number; onProgress?: (pct: number) => void },
): Promise<ImageItem> {
  const form = new FormData()
  form.append('file', file)
  if (options?.album_id != null) form.append('album_id', String(options.album_id))
  if (options?.strategy_id != null) form.append('strategy_id', String(options.strategy_id))
  if (options?.permission != null) form.append('permission', String(options.permission))

  const res = await api.post<ApiResponse<ImageItem>>('/images', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress: (e) => {
      if (e.total) options?.onProgress?.(Math.round((e.loaded * 100) / e.total))
    },
  })
  return res.data.data
}

// ============================================================
// Albums
// ============================================================

export interface Album {
  id: number
  user_id: number
  name: string
  intro: string
  image_num: number
  created_at: string
  updated_at: string
  cover_md5?: string
}

export async function listAlbums(page = 1, pageSize = 20): Promise<PaginatedData<Album>> {
  const res = await api.get<ApiResponse<PaginatedData<Album>>>('/albums', {
    params: { page, page_size: pageSize },
  })
  return res.data.data
}

export async function createAlbum(name: string, intro?: string): Promise<Album> {
  const res = await api.post<ApiResponse<Album>>('/albums', { name, intro })
  return res.data.data
}

export async function updateAlbum(id: number, data: { name?: string; intro?: string }): Promise<Album> {
  const res = await api.put<ApiResponse<Album>>(`/albums/${id}`, data)
  return res.data.data
}

export async function deleteAlbum(id: number): Promise<void> {
  await api.delete(`/albums/${id}`)
}

// ============================================================
// API Tokens
// ============================================================

export interface ApiToken {
  id: number
  name: string
  token?: string
  scopes: string[]
  last_used_at?: string
  expires_at?: string
  created_at: string
}

export async function listApiTokens(): Promise<ApiToken[]> {
  const res = await api.get<ApiResponse<ApiToken[]>>('/api-tokens')
  return res.data.data
}

export async function createApiToken(name: string, expiresIn?: string, scopes?: string[]): Promise<ApiToken> {
  const body: Record<string, unknown> = { name }
  if (expiresIn) body.expires_in = expiresIn
  if (scopes) body.scopes = scopes
  const res = await api.post<ApiResponse<ApiToken>>('/api-tokens', body)
  return res.data.data
}

export async function deleteApiToken(id: number): Promise<void> {
  await api.delete(`/api-tokens/${id}`)
}

// ============================================================
// Strategies (user-facing)
// ============================================================

export interface Strategy {
  id: number
  name: string
  strategy_type: string
  configs: Record<string, unknown>
  created_at: string
  updated_at: string
}

export async function getStrategies(): Promise<Strategy[]> {
  const res = await api.get<ApiResponse<Strategy[]>>('/strategies')
  return res.data.data
}

// ============================================================
// Webhooks
// ============================================================

export interface WebhookItem {
  id: number
  name: string
  url: string
  events: string[]
  enabled: boolean
  secret?: string
  created_at: string
  updated_at: string
}

export interface WebhookDelivery {
  id: number
  webhook_id: number
  status: string
  attempt: number
  max_attempts: number
  next_retry_at: string
  response_status: number | null
  response_body: string
  error_message: string
  duration_ms: number | null
  created_at: string
  completed_at: string | null
}

export async function listWebhooks(): Promise<WebhookItem[]> {
  const res = await api.get<ApiResponse<WebhookItem[]>>('/webhooks')
  return res.data.data
}

export async function createWebhook(name: string, url: string, events: string[]): Promise<WebhookItem> {
  const res = await api.post<ApiResponse<WebhookItem>>('/webhooks', { name, url, events })
  return res.data.data
}

export async function getWebhook(id: number): Promise<WebhookItem> {
  const res = await api.get<ApiResponse<WebhookItem>>(`/webhooks/${id}`)
  return res.data.data
}

export async function updateWebhook(id: number, data: { name?: string; url?: string; events?: string[]; enabled?: boolean }): Promise<WebhookItem> {
  const res = await api.put<ApiResponse<WebhookItem>>(`/webhooks/${id}`, data)
  return res.data.data
}

export async function deleteWebhook(id: number): Promise<void> {
  await api.delete(`/webhooks/${id}`)
}

export async function rotateWebhookSecret(id: number): Promise<{ secret: string }> {
  const res = await api.post<ApiResponse<{ secret: string }>>(`/webhooks/${id}/rotate-secret`)
  return res.data.data
}

export async function testWebhook(id: number): Promise<void> {
  await api.post(`/webhooks/${id}/test`)
}

export async function listWebhookDeliveries(id: number, page = 1, pageSize = 20): Promise<PaginatedData<WebhookDelivery>> {
  const res = await api.get<ApiResponse<PaginatedData<WebhookDelivery>>>(`/webhooks/${id}/deliveries`, {
    params: { page, page_size: pageSize },
  })
  return res.data.data
}

export async function replayWebhookDelivery(deliveryId: number): Promise<void> {
  await api.post(`/webhook-deliveries/${deliveryId}/replay`)
}

// ============================================================
// Pipeline
// ============================================================

export interface PipelineStatus {
  upload: string
  processing: string
  thumbnail: string
  moderation: string
  updated_at: string
}

export async function getPipelineStatus(key: string): Promise<PipelineStatus> {
  const res = await api.get<ApiResponse<PipelineStatus>>(`/images/${encodeURIComponent(key)}/pipeline`)
  return res.data.data
}
