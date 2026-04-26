import type { components } from './types.gen'

// Re-export commonly used types
export type AuthTokens = components['schemas']['AuthResponse']
export type ImageData = components['schemas']['ImageResponse']
export type ImageListItem = components['schemas']['ImageListItem']
export type ImageLinks = components['schemas']['ImageLinks']
export type AlbumData = components['schemas']['AlbumResponse']
export type UserProfile = components['schemas']['UserProfileResponse']
export type AdminUser = components['schemas']['AdminUserResponse']
export type AdminGroup = components['schemas']['AdminGroupResponse']
export type AdminStrategy = components['schemas']['AdminStrategyResponse']
export type Settings = components['schemas']['SettingsResponse']

// API response wrapper
export interface ApiResult<T> {
  status: boolean
  message: string
  data: T
}

// Paginated data
export interface PaginatedData<T> {
  items: T[]
  total: number
  page: number
  size: number
}
