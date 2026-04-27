import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'

import {
  createAdminStrategy,
  deleteAdminStrategy,
  listAdminStrategies,
  updateAdminStrategy,
  type AdminStrategy,
} from '../../../lib/admin-api'

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
      alert((err as { response?: { data?: { message?: string } } })?.response?.data?.message || t('admin.saveFailed'))
    } finally {
      setSaving(false)
    }
  }, [form, editing, qc, t])

  // Delete
  const [deleting, setDeleting] = useState<number | null>(null)
  const onDelete = useCallback(async (id: number) => {
    if (!window.confirm(t('admin.confirmDeleteStrategy'))) return
    setDeleting(id)
    try {
      await deleteAdminStrategy(id)
      await qc.invalidateQueries({ queryKey: ['admin-strategies'] })
    } catch (err: unknown) {
      alert((err as { response?: { data?: { message?: string } } })?.response?.data?.message || t('admin.deleteFailed'))
    } finally {
      setDeleting(null)
    }
  }, [qc, t])

  const update = <K extends keyof StrategyForm>(key: K, value: StrategyForm[K]) =>
    setForm((prev) => ({ ...prev, [key]: value }))

  const inputCls = 'w-full rounded-md border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 dark:border-zinc-600 dark:bg-zinc-900'

  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">{t('admin.strategiesTitle')}</h1>
        <button type="button" onClick={openCreate} className="rounded-lg bg-zinc-900 px-3 py-1.5 text-sm text-white hover:bg-zinc-700 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-300">
          {t('admin.create')}
        </button>
      </div>

      {/* Modal overlay */}
      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40" onClick={() => setShowModal(false)}>
          <div className="w-full max-w-[550px] rounded-xl bg-white p-6 shadow-xl text-zinc-900 dark:bg-zinc-800 dark:text-zinc-100" onClick={(e) => e.stopPropagation()}>
            <h2 className="mb-4 text-lg font-semibold">
              {editing ? t('admin.edit') : t('admin.create')}
            </h2>

            <div className="space-y-3">
              <div>
                <label className="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">{t('admin.colName')}</label>
                <input value={form.name} onChange={(e) => update('name', e.target.value)} placeholder={t('admin.namePlaceholder')} className={inputCls} />
              </div>

              <div>
                <label className="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">{t('admin.colType', { defaultValue: '类型' })}</label>
                <select value={form.type} onChange={(e) => update('type', e.target.value)} disabled={!!editing} className={inputCls}>
                  <option value="local">{t('admin.typeLocal', { defaultValue: '本地存储' })}</option>
                  <option value="s3">{t('admin.typeS3', { defaultValue: 'S3 兼容存储' })}</option>
                </select>
              </div>

              {form.type === 'local' && (
                <div>
                  <label className="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">{t('admin.localRoot', { defaultValue: '存储路径' })}</label>
                  <input value={form.localRoot} onChange={(e) => update('localRoot', e.target.value)} placeholder="/data/images" className={inputCls} />
                </div>
              )}

              {form.type === 's3' && (
                <>
                  <div className="rounded-lg border border-blue-200 bg-blue-50 px-3 py-2 text-xs text-blue-700 dark:border-blue-800 dark:bg-blue-900/20 dark:text-blue-300">
                    <strong>Endpoint</strong> {t('admin.s3EndpointHint', { defaultValue: '是 S3 API 地址（上传用），' })}
                    <strong>URL</strong> {t('admin.s3URLHint', { defaultValue: '是图片公开访问地址（浏览用）。两者通常不同。' })}
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div className="col-span-2">
                      <label className="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">Endpoint</label>
                      <input value={form.s3Endpoint} onChange={(e) => update('s3Endpoint', e.target.value)} placeholder="https://<account-id>.r2.cloudflarestorage.com" className={inputCls} />
                    </div>
                    <div>
                      <label className="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">Region</label>
                      <input value={form.s3Region} onChange={(e) => update('s3Region', e.target.value)} placeholder="R2: auto" className={inputCls} />
                    </div>
                    <div>
                      <label className="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">Bucket</label>
                      <input value={form.s3Bucket} onChange={(e) => update('s3Bucket', e.target.value)} className={inputCls} />
                    </div>
                    <div>
                      <label className="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">Access Key</label>
                      <input value={form.s3AccessKey} onChange={(e) => update('s3AccessKey', e.target.value)} className={inputCls} />
                    </div>
                    <div>
                      <label className="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">Secret Key</label>
                      <input value={form.s3SecretKey} onChange={(e) => update('s3SecretKey', e.target.value)} type="password" className={inputCls} />
                    </div>
                    <div className="col-span-2">
                      <label className="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">{t('admin.accessURL', { defaultValue: '访问 URL' })}</label>
                      <input value={form.s3URL} onChange={(e) => update('s3URL', e.target.value)} placeholder="https://pub-xxx.r2.dev" className={inputCls} />
                    </div>
                  </div>
                </>
              )}
            </div>

            <div className="mt-5 flex justify-end gap-2">
              <button type="button" onClick={() => setShowModal(false)} className="rounded-md border border-zinc-300 px-3 py-1.5 text-sm dark:border-zinc-600">
                {t('admin.cancel')}
              </button>
              <button type="button" onClick={handleSave} disabled={saving || !form.name.trim()} className="rounded-md bg-zinc-900 px-3 py-1.5 text-sm text-white disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900">
                {saving ? '…' : t('admin.confirmSave')}
              </button>
            </div>
          </div>
        </div>
      )}

      {isLoading && <div className="flex justify-center py-12"><div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-400 border-t-transparent" /></div>}
      {error && <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-400">{t('admin.loadFailed')}</p>}
      {strategies && strategies.length === 0 && <p className="py-12 text-center text-sm text-zinc-400">{t('admin.empty')}</p>}

      {strategies && strategies.length > 0 && (
        <div className="divide-y divide-zinc-100 dark:divide-zinc-800">
          {strategies.map((s) => (
            <div key={s.id} className="flex items-center justify-between py-3">
              <div>
                <p className="font-medium">{s.name}</p>
                <p className="mt-0.5 text-xs text-zinc-400">
                  <span className={['rounded px-1.5 py-0.5', s.strategy_type === 'local' ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400' : 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400'].join(' ')}>
                    {s.strategy_type === 'local' ? (t('admin.typeLocal', { defaultValue: '本地存储' })) : 'S3'}
                  </span>
                  <span className="ml-2">{new Date(s.created_at).toLocaleDateString()}</span>
                </p>
              </div>
              <div className="flex gap-1">
                <button type="button" onClick={() => openEdit(s)} className="rounded px-2 py-1 text-xs hover:bg-zinc-100 dark:hover:bg-zinc-800">{t('admin.edit')}</button>
                <button type="button" onClick={() => onDelete(s.id)} disabled={deleting === s.id} className="rounded px-2 py-1 text-xs text-red-500 hover:bg-red-50 disabled:opacity-50 dark:hover:bg-red-900/20">{t('admin.delete')}</button>
              </div>
            </div>
          ))}
        </div>
      )}
    </section>
  )
}
