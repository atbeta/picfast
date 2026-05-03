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
import { extractErrorMessage } from '../../../lib/error-handler'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { EmptyState, LoadingState } from '@/components/page-states'
import { storageStrategyLabel, storageStrategyTypes, type StorageStrategyType } from '@/lib/storage-strategy'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

interface StrategyForm {
  name: string
  type: StorageStrategyType
  // local
  localRoot: string
  // s3
  s3Endpoint: string
  s3Region: string
  s3Bucket: string
  s3AccessKey: string
  s3SecretKey: string
  s3URL: string
  // kodo
  kodoAccessKey: string
  kodoSecretKey: string
  kodoBucket: string
  kodoDomain: string
  kodoZone: string
  kodoPrivate: boolean
  // oss
  ossEndpoint: string
  ossBucket: string
  ossAccessKey: string
  ossSecretKey: string
  ossURL: string
  // cos
  cosBucketURL: string
  cosSecretID: string
  cosSecretKey: string
  cosURL: string
  // webdav
  webdavEndpoint: string
  webdavUsername: string
  webdavPassword: string
  webdavURL: string
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
    kodoAccessKey: '',
    kodoSecretKey: '',
    kodoBucket: '',
    kodoDomain: '',
    kodoZone: 'z0',
    kodoPrivate: false,
    ossEndpoint: '',
    ossBucket: '',
    ossAccessKey: '',
    ossSecretKey: '',
    ossURL: '',
    cosBucketURL: '',
    cosSecretID: '',
    cosSecretKey: '',
    cosURL: '',
    webdavEndpoint: '',
    webdavUsername: '',
    webdavPassword: '',
    webdavURL: '',
  }
}

function formToConfigs(form: StrategyForm): Record<string, unknown> {
  switch (form.type) {
    case 'local':
      return { root: form.localRoot, url: '/i' }
    case 's3':
      return {
        endpoint: form.s3Endpoint,
        region: form.s3Region,
        bucket: form.s3Bucket,
        access_key: form.s3AccessKey,
        secret_key: form.s3SecretKey,
        url: form.s3URL,
      }
    case 'kodo':
      return {
        access_key: form.kodoAccessKey,
        secret_key: form.kodoSecretKey,
        bucket: form.kodoBucket,
        domain: form.kodoDomain,
        zone: form.kodoZone,
        private: form.kodoPrivate,
      }
    case 'oss':
      return {
        endpoint: form.ossEndpoint,
        bucket: form.ossBucket,
        access_key: form.ossAccessKey,
        secret_key: form.ossSecretKey,
        url: form.ossURL,
      }
    case 'cos':
      return {
        bucket_url: form.cosBucketURL,
        secret_id: form.cosSecretID,
        secret_key: form.cosSecretKey,
        url: form.cosURL,
      }
    case 'webdav':
      return {
        endpoint: form.webdavEndpoint,
        username: form.webdavUsername,
        password: form.webdavPassword,
        url: form.webdavURL,
      }
  }
}

function strategyToForm(s: AdminStrategy): StrategyForm {
  const base = emptyForm()
  const c = s.configs || {}
  return {
    ...base,
    name: s.name,
    type: isStrategyType(s.strategy_type) ? s.strategy_type : 'local',
    localRoot: (c.root as string) || '',
    s3Endpoint: (c.endpoint as string) || '',
    s3Region: (c.region as string) || 'us-east-1',
    s3Bucket: (c.bucket as string) || '',
    s3AccessKey: (c.access_key as string) || '',
    s3SecretKey: (c.secret_key as string) || '',
    s3URL: (c.url as string) || '',
    kodoAccessKey: (c.access_key as string) || '',
    kodoSecretKey: (c.secret_key as string) || '',
    kodoBucket: (c.bucket as string) || '',
    kodoDomain: (c.domain as string) || '',
    kodoZone: (c.zone as string) || 'z0',
    kodoPrivate: Boolean(c.private),
    ossEndpoint: (c.endpoint as string) || '',
    ossBucket: (c.bucket as string) || '',
    ossAccessKey: (c.access_key as string) || '',
    ossSecretKey: (c.secret_key as string) || '',
    ossURL: (c.url as string) || '',
    cosBucketURL: (c.bucket_url as string) || '',
    cosSecretID: (c.secret_id as string) || '',
    cosSecretKey: (c.secret_key as string) || '',
    cosURL: (c.url as string) || '',
    webdavEndpoint: (c.endpoint as string) || '',
    webdavUsername: (c.username as string) || '',
    webdavPassword: (c.password as string) || '',
    webdavURL: (c.url as string) || '',
  }
}

function isStrategyType(type: string): type is StorageStrategyType {
  return storageStrategyTypes.includes(type as StorageStrategyType)
}

function requiredFieldsComplete(form: StrategyForm): boolean {
  const has = (...values: string[]) => values.every((value) => value.trim().length > 0)
  switch (form.type) {
    case 'local':
      return has(form.localRoot)
    case 's3':
      return has(form.s3Endpoint, form.s3Bucket, form.s3AccessKey, form.s3SecretKey)
    case 'kodo':
      return has(form.kodoAccessKey, form.kodoSecretKey, form.kodoBucket, form.kodoDomain)
    case 'oss':
      return has(form.ossEndpoint, form.ossBucket, form.ossAccessKey, form.ossSecretKey)
    case 'cos':
      return has(form.cosBucketURL, form.cosSecretID, form.cosSecretKey)
    case 'webdav':
      return has(form.webdavEndpoint, form.webdavUsername, form.webdavPassword)
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
        await updateAdminStrategy(editing.id, { name: form.name.trim(), strategy_type: form.type, configs })
      } else {
        await createAdminStrategy({ name: form.name.trim(), strategy_type: form.type, configs })
      }
      setShowModal(false)
      await qc.invalidateQueries({ queryKey: ['admin-strategies'] })
    } catch (err: unknown) {
      toast.error(extractErrorMessage(err, t('admin.saveFailed')))
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
      toast.error(extractErrorMessage(err, t('admin.deleteFailed')))
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
        <Button size="lg" onClick={openCreate}>
          {t('admin.create')}
        </Button>
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
          onValueChange={(val) => typeof val === 'string' && isStrategyType(val) && update('type', val)}
          items={Object.fromEntries(storageStrategyTypes.map((type) => [type, storageStrategyLabel(t, type)]))}
              >
                <SelectTrigger className="h-10 w-full bg-background border-input">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {storageStrategyTypes.map((type) => (
                    <SelectItem key={type} value={type}>{storageStrategyLabel(t, type)}</SelectItem>
                  ))}
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
                  <strong>{t('admin.fieldEndpoint')}</strong> {t('admin.s3EndpointHint', { defaultValue: '是 S3 API 地址（上传用），' })}
                  <strong>{t('admin.fieldURL')}</strong> {t('admin.s3URLHint', { defaultValue: '是图片公开访问地址（浏览用）。两者通常不同。' })}
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div className="col-span-2">
                    <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.fieldEndpoint')}</label>
                    <input value={form.s3Endpoint} onChange={(e) => update('s3Endpoint', e.target.value)} placeholder="https://s3.<region>.amazonaws.com 或 https://s3.example.com" className={inputCls} />
                  </div>
                  <div>
                    <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.fieldRegion')}</label>
                    <input value={form.s3Region} onChange={(e) => update('s3Region', e.target.value)} placeholder="us-east-1 / cn-north-1 / auto" className={inputCls} />
                  </div>
                  <div>
                    <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.fieldBucket')}</label>
                    <input value={form.s3Bucket} onChange={(e) => update('s3Bucket', e.target.value)} className={inputCls} />
                  </div>
                  <div>
                    <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.fieldAccessKey')}</label>
                    <input value={form.s3AccessKey} onChange={(e) => update('s3AccessKey', e.target.value)} className={inputCls} />
                  </div>
                  <div>
                    <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.fieldSecretKey')}</label>
                    <input value={form.s3SecretKey} onChange={(e) => update('s3SecretKey', e.target.value)} type="password" className={inputCls} />
                  </div>
                  <div className="col-span-2">
                    <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.accessURL', { defaultValue: '访问 URL' })}</label>
                    <input value={form.s3URL} onChange={(e) => update('s3URL', e.target.value)} placeholder="https://img.example.com" className={inputCls} />
                  </div>
                </div>
              </>
            )}

            {form.type === 'kodo' && (
              <div className="grid grid-cols-2 gap-3">
                <div className="col-span-2 rounded-lg border border-primary/20 bg-primary/10 px-3 py-2 text-xs text-primary">
                  {t('admin.kodoHint', { defaultValue: 'Domain 是七牛融合 CDN 或对象下载域名，Zone 可填 z0/z1/z2/na0/as0。' })}
                </div>
                <div>
                  <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.fieldAccessKey')}</label>
                  <input value={form.kodoAccessKey} onChange={(e) => update('kodoAccessKey', e.target.value)} className={inputCls} />
                </div>
                <div>
                  <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.fieldSecretKey')}</label>
                  <input value={form.kodoSecretKey} onChange={(e) => update('kodoSecretKey', e.target.value)} type="password" className={inputCls} />
                </div>
                <div>
                  <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.fieldBucket')}</label>
                  <input value={form.kodoBucket} onChange={(e) => update('kodoBucket', e.target.value)} className={inputCls} />
                </div>
                <div>
                  <label className="mb-1 block text-sm font-medium text-foreground">Zone</label>
                  <input value={form.kodoZone} onChange={(e) => update('kodoZone', e.target.value)} placeholder="z0 华东 / z1 华北 / z2 华南 / na0 北美 / as0 东南亚" className={inputCls} />
                </div>
                <div className="col-span-2">
                  <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.fieldDomain')}</label>
                  <input value={form.kodoDomain} onChange={(e) => update('kodoDomain', e.target.value)} placeholder="https://cdn.example.com 或 https://<bucket-domain>" className={inputCls} />
                </div>
                <label className="col-span-2 flex items-center gap-2 text-sm text-foreground">
                  <input type="checkbox" checked={form.kodoPrivate} onChange={(e) => update('kodoPrivate', e.target.checked)} className="size-4 rounded border-input" />
                  {t('admin.privateBucket', { defaultValue: '私有空间（读取时生成签名 URL）' })}
                </label>
              </div>
            )}

            {form.type === 'oss' && (
              <div className="grid grid-cols-2 gap-3">
                <div className="col-span-2 rounded-lg border border-primary/20 bg-primary/10 px-3 py-2 text-xs text-primary">
                  {t('admin.ossHint', { defaultValue: 'Endpoint 是 OSS API 访问域名，用于上传读取；访问 URL 是 CDN、自定义域名或公开 Bucket 域名，用于生成图片链接。' })}
                </div>
                <div className="col-span-2">
                  <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.fieldEndpoint')}</label>
                  <input value={form.ossEndpoint} onChange={(e) => update('ossEndpoint', e.target.value)} placeholder="https://oss-cn-hongkong.aliyuncs.com" className={inputCls} />
                </div>
                <div>
                  <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.fieldBucket')}</label>
                  <input value={form.ossBucket} onChange={(e) => update('ossBucket', e.target.value)} className={inputCls} />
                </div>
                <div>
                  <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.accessURL', { defaultValue: '访问 URL' })}</label>
                  <input value={form.ossURL} onChange={(e) => update('ossURL', e.target.value)} placeholder="https://img.example.com 或 https://<bucket>.oss-cn-hongkong.aliyuncs.com" className={inputCls} />
                </div>
                <div>
                  <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.fieldAccessKey')}</label>
                  <input value={form.ossAccessKey} onChange={(e) => update('ossAccessKey', e.target.value)} className={inputCls} />
                </div>
                <div>
                  <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.fieldSecretKey')}</label>
                  <input value={form.ossSecretKey} onChange={(e) => update('ossSecretKey', e.target.value)} type="password" className={inputCls} />
                </div>
              </div>
            )}

            {form.type === 'cos' && (
              <div className="grid grid-cols-2 gap-3">
                <div className="col-span-2 rounded-lg border border-primary/20 bg-primary/10 px-3 py-2 text-xs text-primary">
                  {t('admin.cosHint', { defaultValue: 'Bucket URL 使用腾讯云 COS 存储桶访问地址，例如 https://bucket-1250000000.cos.ap-guangzhou.myqcloud.com。' })}
                </div>
                <div className="col-span-2">
                  <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.fieldBucketUrl')}</label>
                  <input value={form.cosBucketURL} onChange={(e) => update('cosBucketURL', e.target.value)} placeholder="https://bucket-1250000000.cos.ap-guangzhou.myqcloud.com" className={inputCls} />
                </div>
                <div>
                  <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.fieldSecretId')}</label>
                  <input value={form.cosSecretID} onChange={(e) => update('cosSecretID', e.target.value)} className={inputCls} />
                </div>
                <div>
                  <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.fieldSecretKey')}</label>
                  <input value={form.cosSecretKey} onChange={(e) => update('cosSecretKey', e.target.value)} type="password" className={inputCls} />
                </div>
                <div className="col-span-2">
                  <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.accessURL', { defaultValue: '访问 URL' })}</label>
                  <input value={form.cosURL} onChange={(e) => update('cosURL', e.target.value)} placeholder="https://img.example.com 或 https://<bucket-appid>.cos.<region>.myqcloud.com" className={inputCls} />
                </div>
              </div>
            )}

            {form.type === 'webdav' && (
              <div className="grid grid-cols-2 gap-3">
                <div className="col-span-2">
                  <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.fieldEndpoint')}</label>
                  <input value={form.webdavEndpoint} onChange={(e) => update('webdavEndpoint', e.target.value)} placeholder="https://dav.example.com/uploads" className={inputCls} />
                </div>
                <div>
                  <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.username', { defaultValue: '用户名' })}</label>
                  <input value={form.webdavUsername} onChange={(e) => update('webdavUsername', e.target.value)} className={inputCls} />
                </div>
                <div>
                  <label className="mb-1 block text-sm font-medium text-foreground">{t('auth.password')}</label>
                  <input value={form.webdavPassword} onChange={(e) => update('webdavPassword', e.target.value)} type="password" className={inputCls} />
                </div>
                <div className="col-span-2">
                  <label className="mb-1 block text-sm font-medium text-foreground">{t('admin.accessURL', { defaultValue: '访问 URL' })}</label>
                  <input value={form.webdavURL} onChange={(e) => update('webdavURL', e.target.value)} placeholder="https://cdn.example.com" className={inputCls} />
                </div>
              </div>
            )}
          </div>

          <div className="mt-5 flex justify-end gap-2 border-t border-border/60 pt-4">
            <Button variant="outline" size="lg" onClick={() => setShowModal(false)}>
              {t('admin.cancel')}
            </Button>
            <Button size="lg" onClick={handleSave} disabled={saving || !form.name.trim() || !requiredFieldsComplete(form)}>
              {saving ? '…' : t('admin.confirmSave')}
            </Button>
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
                    {storageStrategyLabel(t, s.strategy_type)}
                  </span>
                  <span className="text-xs text-muted-foreground/60">{new Date(s.created_at).toLocaleDateString()}</span>
                </div>
                <h3 className="text-lg font-bold tracking-tight text-foreground truncate">{s.name}</h3>
                
                {/* Preview configs */}
                <div className="mt-4 space-y-2 text-xs text-muted-foreground">
                  {s.strategy_type === 'local' ? (
                    <div className="flex items-center gap-2 truncate">
                      <span className="font-medium text-foreground/70 w-12">{t('admin.fieldRoot')}:</span>
                      <code className="truncate rounded bg-muted/50 px-1.5 py-0.5">{(s.configs?.root as string) || '-'}</code>
                    </div>
                  ) : s.strategy_type === 'kodo' ? (
                    <>
                      <div className="flex items-center gap-2 truncate">
                        <span className="font-medium text-foreground/70 w-12">{t('admin.fieldBucket')}:</span>
                        <code className="truncate rounded bg-muted/50 px-1.5 py-0.5">{(s.configs?.bucket as string) || '-'}</code>
                      </div>
                      <div className="flex items-center gap-2 truncate">
                        <span className="font-medium text-foreground/70 w-12">{t('admin.fieldDomain')}:</span>
                        <code className="truncate rounded bg-muted/50 px-1.5 py-0.5">{(s.configs?.domain as string) || '-'}</code>
                      </div>
                    </>
                  ) : s.strategy_type === 'cos' ? (
                    <>
                      <div className="flex items-center gap-2 truncate">
                        <span className="font-medium text-foreground/70 w-20">{t('admin.fieldBucketUrl')}:</span>
                        <code className="truncate rounded bg-muted/50 px-1.5 py-0.5">{(s.configs?.bucket_url as string) || '-'}</code>
                      </div>
                      <div className="flex items-center gap-2 truncate">
                        <span className="font-medium text-foreground/70 w-12">{t('admin.fieldURL')}:</span>
                        <code className="truncate rounded bg-muted/50 px-1.5 py-0.5">{(s.configs?.url as string) || '-'}</code>
                      </div>
                    </>
                  ) : s.strategy_type === 'webdav' ? (
                    <>
                      <div className="flex items-center gap-2 truncate">
                        <span className="font-medium text-foreground/70 w-20">{t('admin.fieldEndpoint')}:</span>
                        <code className="truncate rounded bg-muted/50 px-1.5 py-0.5">{(s.configs?.endpoint as string) || '-'}</code>
                      </div>
                      <div className="flex items-center gap-2 truncate">
                        <span className="font-medium text-foreground/70 w-12">{t('admin.fieldURL')}:</span>
                        <code className="truncate rounded bg-muted/50 px-1.5 py-0.5">{(s.configs?.url as string) || '-'}</code>
                      </div>
                    </>
                  ) : (
                    <>
                      <div className="flex items-center gap-2 truncate">
                        <span className="font-medium text-foreground/70 w-12">{t('admin.fieldBucket')}:</span>
                        <code className="truncate rounded bg-muted/50 px-1.5 py-0.5">{(s.configs?.bucket as string) || '-'}</code>
                      </div>
                      <div className="flex items-center gap-2 truncate">
                        <span className="font-medium text-foreground/70 w-12">{t('admin.fieldURL')}:</span>
                        <code className="truncate rounded bg-muted/50 px-1.5 py-0.5">{(s.configs?.url as string) || '-'}</code>
                      </div>
                    </>
                  )}
                </div>
              </div>

              <div className="mt-6 flex items-center justify-end gap-2 border-t border-border/40 pt-4">
                <Button variant="ghost" size="sm" onClick={() => openEdit(s)} title={t('admin.edit')}>
                  <Pencil className="size-3.5" />
                  {t('admin.edit')}
                </Button>
                {s.id !== 1 && (
                  <Button variant="ghost" size="sm" onClick={() => setDeleteTarget(s.id)} disabled={deleting === s.id} className="text-destructive/70 hover:text-destructive" title={t('admin.delete')}>
                    <Trash2 className="size-3.5" />
                    {t('admin.delete')}
                  </Button>
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
