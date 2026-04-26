import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { UploadZone } from '../../components/upload-zone'
import { UploadResultCard } from '../../components/upload-result'
import { uploadImage } from '../../lib/upload'
import type { UploadResult } from '../../lib/upload'

interface UploadingFile {
  file: File
  progress: number
}

export function GuestUploadPage() {
  const { t } = useTranslation()
  const [results, setResults] = useState<UploadResult[]>([])
  const [uploading, setUploading] = useState<UploadingFile[]>([])
  const [errors, setErrors] = useState<string[]>([])
  const [busy, setBusy] = useState(false)

  const handleFiles = useCallback(async (files: File[]) => {
    setErrors([])
    setBusy(true)

    const newUploading = files.map((f) => ({ file: f, progress: 0 }))
    setUploading(newUploading)

    const newResults: UploadResult[] = []
    const newErrors: string[] = []

    for (let i = 0; i < files.length; i++) {
      try {
        const result = await uploadImage(files[i], {
          onProgress: (p) => {
            setUploading((prev) =>
              prev.map((u, j) => (j === i ? { ...u, progress: p } : u)),
            )
          },
        })
        newResults.push(result)
      } catch (err: unknown) {
        const msg =
          (err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
          t('upload.uploadFailed', { name: files[i].name })
        newErrors.push(msg)
      }
    }

    setResults((prev) => [...newResults, ...prev])
    setUploading([])
    setErrors(newErrors)
    setBusy(false)
  }, [t])

  return (
    <section className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-100">{t('page.guestUpload.title')}</h1>
        <p className="mt-1 text-sm text-zinc-500">{t('page.guestUpload.subtitle')}</p>
      </div>

      <UploadZone onFiles={handleFiles} disabled={busy} />

      {uploading.length > 0 && (
        <div className="space-y-2">
          {uploading.map((u) => (
            <div key={u.file.name} className="flex items-center gap-3">
              <span className="min-w-0 flex-1 truncate text-sm">{u.file.name}</span>
              <div className="h-1.5 w-32 overflow-hidden rounded-full bg-zinc-200 dark:bg-zinc-700">
                <div
                  className="h-full rounded-full bg-zinc-600 transition-all dark:bg-zinc-400"
                  style={{ width: `${u.progress}%` }}
                />
              </div>
              <span className="text-xs text-zinc-500">{u.progress}%</span>
            </div>
          ))}
        </div>
      )}

      {errors.length > 0 && (
        <div className="space-y-2">
          {errors.map((msg, i) => (
            <p key={i} className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-400">
              {msg}
            </p>
          ))}
        </div>
      )}

      {results.length > 0 && (
        <div className="space-y-4">
          <h2 className="text-lg font-medium">{t('upload.results')}</h2>
          {results.map((r) => (
            <UploadResultCard key={r.id} result={r} />
          ))}
        </div>
      )}
    </section>
  )
}
