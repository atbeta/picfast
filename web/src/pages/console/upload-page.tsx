import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { XIcon, CheckCircle2, AlertCircle } from 'lucide-react'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

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
  const [dismissedWelcome, setDismissedWelcome] = useState(false)

  const isNewUser = user && user.image_num === 0 && user.album_num === 0

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
    <section className="mx-auto max-w-4xl space-y-8 animate-in slide-in-from-bottom-4 fade-in duration-700">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <h1 className="text-2xl font-bold tracking-tight text-transparent bg-clip-text bg-gradient-to-r from-foreground to-foreground/70 dark:from-white dark:to-white/60">
          {t('page.upload.title')}
        </h1>

        {/* Strategy Info / Selector */}
        <div className="flex items-center gap-3 rounded-full border border-border/50 bg-card/50 backdrop-blur-md px-4 py-1.5 shadow-sm text-sm">
          <span className="text-muted-foreground font-medium">{t('upload.strategy', { defaultValue: 'Strategy:' })}</span>
          {strategies.length > 1 ? (
            <Select 
              value={selectedStrategyId?.toString() ?? ''}
              onValueChange={(val) => val !== null && onStrategyChange(Number(val))}
              items={Object.fromEntries(strategies.map(s => [s.id.toString(), `${s.name} (${s.strategy_type === 'local' ? t('admin.typeLocal', { defaultValue: 'Local' }) : 'S3'})`]))}
            >
              <SelectTrigger className="w-[140px] h-7 bg-transparent border-none shadow-none font-semibold text-foreground hover:bg-accent/50 focus:ring-0 px-2 py-0">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {strategies.map((s) => (
                  <SelectItem key={s.id} value={s.id.toString()}>
                    {s.name} ({s.strategy_type === 'local' ? (t('admin.typeLocal', { defaultValue: 'Local' })) : 'S3'})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          ) : strategies.length === 1 ? (
            <span className="font-semibold text-foreground">
              {strategies[0].name} <span className="text-muted-foreground/70 font-normal">({strategies[0].strategy_type === 'local' ? (t('admin.typeLocal', { defaultValue: 'Local' })) : 'S3'})</span>
            </span>
          ) : (
            <span className="text-muted-foreground/50">...</span>
          )}
        </div>
      </div>

      {/* New user welcome banner */}
      {isNewUser && !dismissedWelcome && (
        <div className="relative overflow-hidden rounded-2xl border border-primary/20 bg-gradient-to-r from-primary/10 via-primary/5 to-transparent p-5 shadow-sm animate-in fade-in zoom-in-95 duration-500">
          <div className="flex items-start gap-4">
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary/20 text-primary">
              <CheckCircle2 className="h-5 w-5" />
            </div>
            <div className="flex-1 pt-1">
              <p className="text-base font-semibold text-foreground">{t('auth.registerSuccess')}</p>
              <p className="mt-1 text-sm text-muted-foreground">{t('auth.registerSuccessDesc')}</p>
            </div>
            <button
              type="button"
              onClick={() => setDismissedWelcome(true)}
              className="shrink-0 rounded-full p-2 text-muted-foreground hover:bg-background/50 hover:text-foreground transition-colors"
            >
              <XIcon className="h-4 w-4" />
            </button>
          </div>
        </div>
      )}

      <div className="rounded-3xl border border-border/50 bg-card/30 backdrop-blur-xl p-2 shadow-xl shadow-black/5 dark:shadow-black/20 transition-all hover:bg-card/50">
        <UploadZone onFiles={handleFiles} disabled={busy} />
      </div>

      {uploading.length > 0 && (
        <div className="space-y-4 rounded-3xl border border-border/50 bg-card/50 backdrop-blur-xl p-6 shadow-xl shadow-black/5 dark:shadow-black/20 animate-in fade-in slide-in-from-bottom-4 duration-500">
          <div className="flex items-center gap-3 mb-2">
            <div className="h-2 w-2 rounded-full bg-primary animate-pulse" />
            <h3 className="text-sm font-semibold text-foreground tracking-tight">{t('upload.uploading', { defaultValue: '正在上传...' })}</h3>
          </div>
          {uploading.map((u) => (
            <div key={u.file.name} className="flex flex-col gap-2">
              <div className="flex items-center justify-between gap-4">
                <span className="min-w-0 flex-1 truncate text-sm font-medium text-foreground/80">{u.file.name}</span>
                <span className="text-xs font-bold text-primary w-10 text-right">{u.progress}%</span>
              </div>
              <div className="h-1.5 w-full overflow-hidden rounded-full bg-secondary/50">
                <div
                  className="h-full rounded-full bg-gradient-to-r from-primary/50 to-primary transition-all duration-300 ease-out"
                  style={{ width: `${u.progress}%` }}
                />
              </div>
            </div>
          ))}
        </div>
      )}

      {errors.length > 0 && (
        <div className="space-y-2 animate-in fade-in">
          {errors.map((msg, i) => (
            <div key={i} className="flex items-center gap-3 rounded-xl bg-destructive/10 px-5 py-4 text-sm text-destructive border border-destructive/20 shadow-sm">
              <AlertCircle className="h-5 w-5 shrink-0" />
              <p className="font-medium">{msg}</p>
            </div>
          ))}
        </div>
      )}

      {results.length > 0 && (
        <div className="space-y-6 animate-in fade-in slide-in-from-bottom-8 duration-500 delay-150 fill-mode-both">
          <div className="flex items-center gap-3">
            <h2 className="text-xl font-semibold tracking-tight">{t('upload.results')}</h2>
            <div className="h-px flex-1 bg-border/50" />
          </div>
          <div className="grid gap-6">
            {results.map((r) => (
              <UploadResultCard key={r.id} result={r} />
            ))}
          </div>
        </div>
      )}
    </section>
  )
}
