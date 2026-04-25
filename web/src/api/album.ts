import api from './index'

export function getAlbums(page = 1, pageSize = 20) {
  return api.get('/albums', { params: { page, page_size: pageSize } })
}

export function createAlbum(name: string, intro = '') {
  return api.post('/albums', { name, intro })
}

export function updateAlbum(id: number, data: { name?: string; intro?: string }) {
  return api.put(`/albums/${id}`, data)
}

export function deleteAlbum(id: number) {
  return api.delete(`/albums/${id}`)
}
