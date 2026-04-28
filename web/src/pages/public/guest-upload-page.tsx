import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertCircle } from 'lucide-react'

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
    <section className="mx-auto max-w-3xl space-y-8 animate-in slide-in-from-bottom-4 fade-in duration-700">
      <div className="text-center space-y-4">
        <h1 className="text-4xl md:text-5xl font-extrabold tracking-tight text-transparent bg-clip-text bg-gradient-to-br from-foreground to-foreground/70 dark:from-white dark:to-white/60">
          {t('page.guestUpload.title')}
        </h1>
        <p className="text-lg text-muted-foreground max-w-xl mx-auto">
          {t('page.guestUpload.subtitle')}
        </p>
      </div>

      <div className="rounded-2xl border border-border/50 bg-card/50 backdrop-blur-xl p-2 shadow-xl shadow-black/5 dark:shadow-black/20">
        <UploadZone onFiles={handleFiles} disabled={busy} />
      </div>

      {uploading.length > 0 && (
        <div className="space-y-3 rounded-xl border border-border/50 bg-card/50 backdrop-blur-sm p-4 shadow-sm">
          {uploading.map((u) => (
            <div key={u.file.name} className="flex items-center gap-4">
              <span className="min-w-0 flex-1 truncate text-sm font-medium">{u.file.name}</span>
              <div className="h-2 w-32 md:w-48 overflow-hidden rounded-full bg-secondary">
                <div
                  className="h-full rounded-full bg-primary transition-all duration-300 ease-out"
                  style={{ width: `${u.progress}%` }}
                />
              </div>
              <span className="text-xs font-medium text-muted-foreground w-8 text-right">{u.progress}%</span>
            </div>
          ))}
        </div>
      )}

      {errors.length > 0 && (
        <div className="space-y-2 animate-in fade-in">
          {errors.map((msg, i) => (
            <div key={i} className="flex items-center gap-3 rounded-xl bg-destructive/10 px-5 py-4 text-sm text-destructive border border-destructive/20 shadow-sm">
              <AlertCircle className="w-5 h-5 shrink-0" />
              <p className="font-medium">{msg}</p>
            </div>
          ))}
        </div>
      )}

      {results.length > 0 && (
        <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500 delay-150 fill-mode-both">
          <h2 className="text-xl font-semibold tracking-tight">{t('upload.results')}</h2>
          <div className="grid gap-4">
            {results.map((r) => (
              <UploadResultCard key={r.id} result={r} />
            ))}
          </div>
        </div>
      )}
    </section>
  )
}
