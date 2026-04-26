import api from './api'

// --- Types ---

export interface UploadResult {
  id: number
  key: string
  origin_name: string
  size_bytes: number
  mimetype: string
  extension: string
  width: number
  height: number
  md5: string
  sha1: string
  permission: number
  album_id: number | null
  moderation_status: string
  links: UploadLinks
  created_at: string
}

export interface UploadLinks {
  url: string
  html: string
  bbcode: string
  markdown: string
  thumbnail_url: string
}

interface ApiResponse<T> {
  status: boolean
  message: string
  data: T
}

// --- API ---

export async function uploadImage(
  file: File,
  options?: { onProgress?: (percent: number) => void },
): Promise<UploadResult> {
  const form = new FormData()
  form.append('file', file)

  const res = await api.post<ApiResponse<UploadResult>>('/upload', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress: (e) => {
      if (e.total) {
        options?.onProgress?.(Math.round((e.loaded * 100) / e.total))
      }
    },
  })
  return res.data.data
}

// --- Helpers ---

export function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}
