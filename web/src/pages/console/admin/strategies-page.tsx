import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Pencil, Trash2, Database } from 'lucide-react'

import {
  createAdminStrategy,
  deleteAdminStrategy,
  listAdminStrategies,
  updateAdminStrategy,
  type AdminStrategy,
} from '../../../lib/admin-api'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { EmptyState, LoadingState } from '@/components/page-states'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

interface StrategyForm {
  name: string
  type: string
  // local
  localRoot: string
  // s3
  s3Endpoint: string
  s3Region: string
  s3Bucket: string
  s3AccessKey: string
  s3SecretKey: string
  s3URL: string
}

function emptyForm(): StrategyForm {
  return {
    name: '',
    type: 'local',
    localRoot: '/data/images',
    s3Endpoint: '',
    s3Region: 'us-east-1',
    s3Bucket: '',
    s3AccessKey: '',
    s3SecretKey: '',
    s3URL: '',
  }
}

function formToConfigs(form: StrategyForm): Record<string, unknown> {
  if (form.type === 'local') {
    return { root: form.localRoot, url: '/i' }
  }
  return {
    endpoint: form.s3Endpoint,
    region: form.s3Region,
    bucket: form.s3Bucket,
    access_key: form.s3AccessKey,
    secret_key: form.s3SecretKey,
    url: form.s3URL,
  }
}

function strategyToForm(s: AdminStrategy): StrategyForm {
  const c = s.configs || {}
  return {
    name: s.name,
    type: s.strategy_type,
    localRoot: (c.root as string) || '',
    s3Endpoint: (c.endpoint as string) || '',
    s3Region: (c.region as string) || 'us-east-1',
    s3Bucket: (c.bucket as string) || '',
    s3AccessKey: (c.access_key as string) || '',
    s3SecretKey: (c.secret_key as string) || '',
    s3URL: (c.url as string) || '',
  }
}

export function AdminStrategiesPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()

  const { data: strategies, isLoading, error } = useQuery({
    queryKey: ['admin-strategies'],
    queryFn: listAdminStrategies,
  })

  // Modal state
  const [showModal, setShowModal] = useState(false)
  const [editing, setEditing] = useState<AdminStrategy | null>(null)
  const [form, setForm] = useState<StrategyForm>(emptyForm())
  const [saving, setSaving] = useState(false)

  const openCreate = () => {
    setEditing(null)
    setForm(emptyForm())
    setShowModal(true)
  }

  const openEdit = (s: AdminStrategy) => {
    setEditing(s)
    setForm(strategyToForm(s))
    setShowModal(true)
  }

  const handleSave = useCallback(async () => {
    if (!form.name.trim()) return
    setSaving(true)
    try {
      const configs = formToConfigs(form)
      if (editing) {
        await updateAdminStrategy(editing.id, { name: form.name.trim(), configs })
      } else {
        await createAdminStrategy({ name: form.name.trim(), strategy_type: form.type, configs })
      }
      setShowModal(false)
      await qc.invalidateQueries({ queryKey: ['admin-strategies'] })
    } catch (err: unknown) {
      toast.error((err as { response?: { data?: { message?: string } } })?.response?.data?.message || t('admin.saveFailed'))
    } finally {
      setSaving(false)
    }
  }, [form, editing, qc, t])

  // Delete
  const [deleting, setDeleting] = useState<number | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<number | null>(null)

  const onDelete = useCallback(async () => {
    if (deleteTarget === null) return
    setDeleting(deleteTarget)
    try {
      await deleteAdminStrategy(deleteTarget)
      setDeleteTarget(null)
      await qc.invalidateQueries({ queryKey: ['admin-strategies'] })
    } catch (err: unknown) {
      toast.error((err as { response?: { data?: { message?: string } } })?.response?.data?.message || t('admin.deleteFailed'))
    } finally {
      setDeleting(null)
    }
  }, [deleteTarget, qc, t])

  const update = <K extends keyof StrategyForm>(key: K, value: StrategyForm[K]) =>
    setForm((prev) => ({ ...prev, [key]: value }))

  const inputCls = 'h-10 w-full rounded-lg border border-input bg-background px-3 text-sm outline-none transition-colors duration-150 focus:border-primary focus:ring-2 focus:ring-primary/20'

  return (
    <section className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight">{t('admin.strategiesTitle')}</h1>
          <p className="text-sm text-muted-foreground">{t('admin.strategiesSubtitle', { defaultValue: '维护存储策略配置并统一上传落盘规则。' })}</p>
        </div>
        <button type="button" onClick={openCreate} className="h-10 rounded-lg bg-primary px-4 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors shadow-sm cursor-pointer">
          {t('admin.create')}
        </button>
      </div>

      {/* Strategy create/edit dialog */}
      <Dialog open={showModal} onOpenChange={setShowModal}>
        <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-[550px]">
          <DialogHeader>
            <DialogTitle>
              {editing ? t('admin.edit') : t('admin.create')}
            </DialogTitle>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.colName')}</label>
              <input value={form.name} onChange={(e) => update('name', e.target.value)} placeholder={t('admin.namePlaceholder')} className={inputCls} />
            </div>

            <div>
              <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.colType', { defaultValue: '类型' })}</label>
              <Select 
          value={form.type} 
          onValueChange={(val) => val !== null && update('type', val as string)} 
          disabled={!!editing}
          items={{
                  local: t('admin.typeLocal', { defaultValue: '本地存储' }),
                  s3: t('admin.typeS3', { defaultValue: 'S3 兼容存储' })
                }}
              >
                <SelectTrigger className="h-10 w-full bg-background border-input">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="local">{t('admin.typeLocal', { defaultValue: '本地存储' })}</SelectItem>
                  <SelectItem value="s3">{t('admin.typeS3', { defaultValue: 'S3 兼容存储' })}</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {form.type === 'local' && (
              <div>
                <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.localRoot', { defaultValue: '存储路径' })}</label>
                <input value={form.localRoot} onChange={(e) => update('localRoot', e.target.value)} placeholder="/data/images" className={inputCls} />
              </div>
            )}

            {form.type === 's3' && (
              <>
                <div className="rounded-lg border border-primary/20 bg-primary/10 px-3 py-2 text-xs text-primary">
                  <strong>Endpoint</strong> {t('admin.s3EndpointHint', { defaultValue: '是 S3 API 地址（上传用），' })}
                  <strong>URL</strong> {t('admin.s3URLHint', { defaultValue: '是图片公开访问地址（浏览用）。两者通常不同。' })}
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div className="col-span-2">
                    <label className="mb-1 block text-sm font-medium text-foreground">Endpoint</label>
                    <input value={form.s3Endpoint} onChange={(e) => update('s3Endpoint', e.target.value)} placeholder="https://<account-id>.r2.cloudflarestorage.com" className={inputCls} />
                  </div>
                  <div>
                    <label className="mb-1 block text-sm font-medium text-foreground">Region</label>
                    <input value={form.s3Region} onChange={(e) => update('s3Region', e.target.value)} placeholder="R2: auto" className={inputCls} />
                  </div>
                  <div>
                    <label className="mb-1 block text-sm font-medium text-foreground">Bucket</label>
                    <input value={form.s3Bucket} onChange={(e) => update('s3Bucket', e.target.value)} className={inputCls} />
                  </div>
                  <div>
                    <label className="mb-1 block text-sm font-medium text-foreground">Access Key</label>
                    <input value={form.s3AccessKey} onChange={(e) => update('s3AccessKey', e.target.value)} className={inputCls} />
                  </div>
                  <div>
                    <label className="mb-1 block text-sm font-medium text-foreground">Secret Key</label>
                    <input value={form.s3SecretKey} onChange={(e) => update('s3SecretKey', e.target.value)} type="password" className={inputCls} />
                  </div>
                  <div className="col-span-2">
                    <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.accessURL', { defaultValue: '访问 URL' })}</label>
                    <input value={form.s3URL} onChange={(e) => update('s3URL', e.target.value)} placeholder="https://pub-xxx.r2.dev" className={inputCls} />
                  </div>
                </div>
              </>
            )}
          </div>

          <div className="mt-5 flex justify-end gap-2 border-t border-border/60 pt-4">
            <button type="button" onClick={() => setShowModal(false)} className="h-10 rounded-lg border border-input bg-background px-4 text-sm hover:bg-accent hover:text-accent-foreground transition-colors cursor-pointer">
              {t('admin.cancel')}
            </button>
            <button type="button" onClick={handleSave} disabled={saving || !form.name.trim()} className="h-10 rounded-lg bg-primary px-4 text-sm text-primary-foreground disabled:opacity-50 hover:bg-primary/90 transition-colors cursor-pointer disabled:cursor-not-allowed">
              {saving ? '…' : t('admin.confirmSave')}
            </button>
          </div>
        </DialogContent>
      </Dialog>

      {isLoading && <LoadingState />}
      {error && <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">{t('admin.loadFailed')}</p>}
      {strategies && strategies.length === 0 && (
        <EmptyState
          icon={<Database className="size-6 text-muted-foreground" />}
          title={t('admin.empty')}
          description={t('admin.strategiesEmptyDesc', { defaultValue: '请先创建本地或 S3 策略，以启用分组绑定与上传能力。' })}
        />
      )}

      {strategies && strategies.length > 0 && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {strategies.map((s) => (
            <div key={s.id} className="group relative flex flex-col justify-between rounded-xl border border-border bg-card/80 p-5 shadow-sm transition-colors duration-150 hover:shadow-sm">
              <div>
                <div className="flex items-center justify-between mb-3">
                  <span className={['rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-wider', s.strategy_type === 'local' ? 'bg-primary/10 text-primary' : 'bg-info/10 text-info'].join(' ')}>
                    {s.strategy_type === 'local' ? (t('admin.typeLocal', { defaultValue: '本地存储' })) : 'S3 API'}
                  </span>
                  <span className="text-xs text-muted-foreground/60">{new Date(s.created_at).toLocaleDateString()}</span>
                </div>
                <h3 className="text-lg font-bold tracking-tight text-foreground truncate">{s.name}</h3>
                
                {/* Preview configs */}
                <div className="mt-4 space-y-2 text-xs text-muted-foreground">
                  {s.strategy_type === 'local' ? (
                    <div className="flex items-center gap-2 truncate">
                      <span className="font-medium text-foreground/70 w-12">Root:</span>
                      <code className="truncate rounded bg-muted/50 px-1.5 py-0.5">{(s.configs?.root as string) || '-'}</code>
                    </div>
                  ) : (
                    <>
                      <div className="flex items-center gap-2 truncate">
                        <span className="font-medium text-foreground/70 w-12">Bucket:</span>
                        <code className="truncate rounded bg-muted/50 px-1.5 py-0.5">{(s.configs?.bucket as string) || '-'}</code>
                      </div>
                      <div className="flex items-center gap-2 truncate">
                        <span className="font-medium text-foreground/70 w-12">URL:</span>
                        <code className="truncate rounded bg-muted/50 px-1.5 py-0.5">{(s.configs?.url as string) || '-'}</code>
                      </div>
                    </>
                  )}
                </div>
              </div>

              <div className="mt-6 flex items-center justify-end gap-2 border-t border-border/40 pt-4">
                <button type="button" onClick={() => openEdit(s)} title={t('admin.edit')} className="flex h-8 items-center gap-1.5 rounded-lg px-3 text-xs font-medium text-muted-foreground hover:bg-muted hover:text-foreground transition-colors cursor-pointer">
                  <Pencil className="size-3.5" />
                  {t('admin.edit')}
                </button>
                {s.id !== 1 && (
                  <button type="button" onClick={() => setDeleteTarget(s.id)} disabled={deleting === s.id} title={t('admin.delete')} className="flex h-8 items-center gap-1.5 rounded-lg px-3 text-xs font-medium text-destructive/70 hover:bg-destructive/10 hover:text-destructive transition-colors disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed">
                    <Trash2 className="size-3.5" />
                    {t('admin.delete')}
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      <ConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}
        title={t('admin.confirmDeleteStrategy')}
        description={t('admin.deleteStrategyDescription')}
        confirmLabel={t('admin.delete')}
        onConfirm={onDelete}
        loading={!!deleting}
      />
    </section>
  )
}
