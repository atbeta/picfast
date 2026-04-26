import api from './index'
import type { AlbumData, ApiResult, PaginatedData } from './types'

export function getAlbums(page = 1, pageSize = 20) {
  return api.get<ApiResult<PaginatedData<AlbumData>>>('/albums', { params: { page, page_size: pageSize } })
}

export function createAlbum(name: string, intro = '') {
  return api.post<ApiResult<AlbumData>>('/albums', { name, intro })
}

export function updateAlbum(id: number, data: { name?: string; intro?: string }) {
  return api.put<ApiResult<AlbumData>>(`/albums/${id}`, data)
}

export function deleteAlbum(id: number) {
  return api.delete(`/albums/${id}`)
}
