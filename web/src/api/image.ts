import api from './index'

export function uploadImage(file: File, params?: Record<string, string>) {
  const form = new FormData()
  form.append('file', file)
  if (params) {
    for (const [k, v] of Object.entries(params)) {
      form.append(k, v)
    }
  }
  return api.post('/images', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export function getImages(page = 1, pageSize = 20) {
  return api.get('/images', { params: { page, page_size: pageSize } })
}

export function getImage(key: string) {
  return api.get(`/images/${key}`)
}

export function deleteImage(key: string) {
  return api.delete(`/images/${key}`)
}

export function updateImage(key: string, data: { album_id?: number; permission?: number }) {
  return api.patch(`/images/${key}`, data)
}
