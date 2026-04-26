import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { UploadZone } from '../../components/upload-zone'
import { UploadResultCard } from '../../components/upload-result'
import { getStrategies, uploadImageAuth } from '../../lib/console-api'
import type { ImageItem, Strategy } from '../../lib/console-api'
import { useAuth } from '../../lib/auth-context'

interface UploadingFile {
  file: File
  progress: number
}

export function UploadPage() {
  const { t } = useTranslation()
  const { user } = useAuth()
  const [results, setResults] = useState<ImageItem[]>([])
  const [uploading, setUploading] = useState<UploadingFile[]>([])
  const [errors, setErrors] = useState<string[]>([])
  const [busy, setBusy] = useState(false)

  // Strategy selector
  const [strategies, setStrategies] = useState<Strategy[]>([])
  const [selectedStrategyId, setSelectedStrategyId] = useState<number | null>(null)

  useEffect(() => {
    getStrategies()
      .then((list) => {
        setStrategies(list)
        // Priority: user settings -> localStorage -> first available
        const userDefault = (user?.settings as Record<string, unknown>)?.default_strategy
        const userDefaultNum = userDefault ? Number(userDefault) : null
        const saved = localStorage.getItem('default_strategy_id')
        let initialId: number | null = null
        if (userDefaultNum && list.some((s) => s.id === userDefaultNum)) {
          initialId = userDefaultNum
        } else if (saved && list.some((s) => s.id === Number(saved))) {
          initialId = Number(saved)
        } else if (list.length > 0) {
          initialId = list[0].id
        }
        setSelectedStrategyId(initialId)
      })
      .catch(() => {})
  }, [user])

  const onStrategyChange = (id: number) => {
    setSelectedStrategyId(id)
    localStorage.setItem('default_strategy_id', String(id))
  }

  const handleFiles = useCallback(async (files: File[]) => {
    setErrors([])
    setBusy(true)
    setUploading(files.map((f) => ({ file: f, progress: 0 })))

    const newResults: ImageItem[] = []
    const newErrors: string[] = []

    for (let i = 0; i < files.length; i++) {
      try {
        const result = await uploadImageAuth(files[i], {
          strategy_id: selectedStrategyId ?? undefined,
          onProgress: (p) =>
            setUploading((prev) => prev.map((u, j) => (j === i ? { ...u, progress: p } : u))),
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
  }, [selectedStrategyId, t])

  return (
    <section className="space-y-6">
      <h1 className="text-xl font-semibold">{t('page.upload.title')}</h1>

      {/* Strategy selector */}
      {strategies.length > 1 && (
        <div className="flex items-center gap-3 text-sm text-zinc-500">
          <span>{t('upload.strategy', { defaultValue: '存储策略：' })}</span>
          <select
            value={selectedStrategyId ?? ''}
            onChange={(e) => onStrategyChange(Number(e.target.value))}
            className="rounded-md border border-zinc-300 bg-white px-2 py-1 text-sm dark:border-zinc-600 dark:bg-zinc-900"
          >
            {strategies.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name} ({s.strategy_type === 'local' ? (t('admin.typeLocal', { defaultValue: '本地' })) : 'S3'})
              </option>
            ))}
          </select>
        </div>
      )}
      {strategies.length === 1 && (
        <div className="text-sm text-zinc-500">
          {t('upload.currentStrategy', { defaultValue: '当前使用存储策略：' })}<span className="font-medium text-zinc-700 dark:text-zinc-300">{strategies[0].name}</span>
          ({strategies[0].strategy_type === 'local' ? (t('admin.typeLocal', { defaultValue: '本地' })) : 'S3'})
        </div>
      )}

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
