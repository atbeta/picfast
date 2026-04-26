import api from './index'
import type { ImageData, ImageListItem, ApiResult, PaginatedData } from './types'

export function uploadImage(file: File, params?: Record<string, string>, onProgress?: (percent: number) => void) {
	const form = new FormData()
	form.append('file', file)
	if (params) {
		for (const [k, v] of Object.entries(params)) {
			form.append(k, v)
		}
	}
	return api.post<ApiResult<ImageData>>('/images', form, {
		headers: { 'Content-Type': 'multipart/form-data' },
		onUploadProgress: onProgress
			? (e) => {
				const total = e.total || e.loaded || 1
				onProgress(Math.round((e.loaded / total) * 100))
			}
			: undefined,
	})
}

export function getImages(page = 1, pageSize = 20) {
	return api.get<ApiResult<PaginatedData<ImageListItem>>>('/images', { params: { page, page_size: pageSize } })
}

export function getImage(key: string) {
	return api.get<ApiResult<ImageData>>(`/images/${key}`)
}

export function deleteImage(key: string) {
	return api.delete(`/images/${key}`)
}

export function updateImage(key: string, data: { album_id?: number; permission?: number }) {
	return api.patch<ApiResult<ImageData>>(`/images/${key}`, data)
}
