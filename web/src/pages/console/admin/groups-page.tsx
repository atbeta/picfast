import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Pencil, Trash2, FolderTree } from 'lucide-react'

import {
  createAdminGroup,
  deleteAdminGroup,
  listAdminGroups,
  listAdminStrategies,
  setAdminGroupStrategies,
  updateAdminGroup,
  type AdminGroup,
} from '../../../lib/admin-api'
import { extractErrorMessage } from '../../../lib/error-handler'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { EmptyState, LoadingState } from '@/components/page-states'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Switch } from '@/components/ui/switch'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

interface GroupForm {
  name: string
  is_default: boolean
  max_size: number
  extensions: string
  limit_per_day: number
  limit_per_month: number
  strategy_ids: number[]
  default_strategy_id: number
  image_save_quality: number
  image_save_format: string
  is_strip_exif: boolean
  is_enable_watermark: boolean
  watermark_text: string
  watermark_position: string
  watermark_font_size: number
  watermark_color: string
  watermark_opacity: number
  raw_configs: Record<string, unknown>
}

const commonExtensions = ['jpeg', 'jpg', 'png', 'gif', 'webp', 'bmp', 'svg', 'ico']

function parseExtensions(value: string): string[] {
  return Array.from(
    new Set(
      value
        .split(',')
        .map((item) => item.trim().toLowerCase())
        .filter(Boolean),
    ),
  )
}

function emptyForm(): GroupForm {
  return {
    name: '',
    is_default: false,
    max_size: 5,
    extensions: 'jpg,jpeg,png,gif,webp,bmp,svg',
    limit_per_day: 300,
    limit_per_month: 9999,
    strategy_ids: [],
    default_strategy_id: 0,
    image_save_quality: 85,
    image_save_format: '',
    is_strip_exif: true,
    is_enable_watermark: false,
    watermark_text: '',
    watermark_position: 'bottom-right',
    watermark_font_size: 28,
    watermark_color: '#FFFFFF',
    watermark_opacity: 0.6,
    raw_configs: {},
  }
}

function parseWatermarkConfig(raw: unknown): Record<string, unknown> {
  if (typeof raw === 'string') {
    try {
      const parsed = JSON.parse(raw)
      return typeof parsed === 'object' && parsed !== null ? (parsed as Record<string, unknown>) : {}
    } catch {
      return {}
    }
  }
  return typeof raw === 'object' && raw !== null ? (raw as Record<string, unknown>) : {}
}

function groupToForm(g: AdminGroup): GroupForm {
  const c = g.configs || {}
  const watermark = parseWatermarkConfig(c.watermark_configs)
  return {
    name: g.name,
    is_default: g.is_default,
    max_size: Math.round(((c.maximum_file_size as number) || 5242880) / 1048576),
    extensions: ((c.accepted_extensions as string[]) || []).join(','),
    limit_per_day: (c.limit_per_day as number) || 300,
    limit_per_month: (c.limit_per_month as number) || 9999,
    strategy_ids: (g.strategy_ids || []).map(Number),
    default_strategy_id: Number(c.default_strategy_id || 0),
    image_save_quality: Number(c.image_save_quality ?? 100),
    image_save_format: (c.image_save_format as string) || '',
    is_strip_exif: Boolean(c.is_strip_exif),
    is_enable_watermark: Boolean(c.is_enable_watermark),
    watermark_text: (watermark.text as string) || '',
    watermark_position: (watermark.position as string) || 'bottom-right',
    watermark_font_size: Number(watermark.font_size ?? 28),
    watermark_color: (watermark.color as string) || '#FFFFFF',
    watermark_opacity: Number(watermark.opacity ?? 0.6),
    raw_configs: c,
  }
}

function formToConfigs(form: GroupForm) {
  const normalizedDefaultStrategyID = form.strategy_ids.includes(form.default_strategy_id)
    ? form.default_strategy_id
    : 0
  const quality = Number.isFinite(form.image_save_quality) ? Math.max(1, Math.min(100, form.image_save_quality)) : 100
  const watermarkOpacity = Number.isFinite(form.watermark_opacity) ? Math.max(0, Math.min(1, form.watermark_opacity)) : 0.6
  const watermarkFontSize = Number.isFinite(form.watermark_font_size) ? Math.max(8, Math.min(200, form.watermark_font_size)) : 28

  return {
    ...(form.raw_configs || {}),
    maximum_file_size: form.max_size * 1048576,
    accepted_extensions: parseExtensions(form.extensions),
    default_strategy_id: normalizedDefaultStrategyID,
    limit_per_day: form.limit_per_day,
    limit_per_month: form.limit_per_month,
    image_save_quality: quality,
    image_save_format: form.image_save_format || '',
    is_strip_exif: form.is_strip_exif,
    is_enable_watermark: form.is_enable_watermark,
    watermark_configs: {
      text: form.watermark_text,
      position: form.watermark_position,
      font_size: watermarkFontSize,
      color: form.watermark_color,
      opacity: watermarkOpacity,
    },
  }
}

function formatSize(bytes: number): string {
  if (!bytes) return '-'
  const mb = bytes / 1048576
  return mb >= 1024 ? `${(mb / 1024).toFixed(1)} GB` : `${mb.toFixed(1)} MB`
}

export function AdminGroupsPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()

  const { data: groups, isLoading, error } = useQuery({
    queryKey: ['admin-groups'],
    queryFn: listAdminGroups,
  })

  const { data: allStrategies = [] } = useQuery({
    queryKey: ['admin-strategies'],
    queryFn: listAdminStrategies,
  })

  // Modal state
  const [showModal, setShowModal] = useState(false)
  const [editing, setEditing] = useState<AdminGroup | null>(null)
  const [form, setForm] = useState<GroupForm>(emptyForm())
  const [saving, setSaving] = useState(false)

  const openCreate = () => {
    setEditing(null)
    setForm(emptyForm())
    setShowModal(true)
  }

  const openEdit = (g: AdminGroup) => {
    setEditing(g)
    setForm(groupToForm(g))
    setShowModal(true)
  }

  const handleSave = useCallback(async () => {
    if (!form.name.trim()) return
    setSaving(true)
    try {
      const configs = formToConfigs(form)
      if (editing) {
        await updateAdminGroup(editing.id, {
          name: form.name.trim(),
          is_default: form.is_default,
          is_guest: editing.is_guest,
          configs,
        })
        await setAdminGroupStrategies(editing.id, form.strategy_ids)
      } else {
        await createAdminGroup({
          name: form.name.trim(),
          is_default: form.is_default,
          configs,
        })
      }
      setShowModal(false)
      await qc.invalidateQueries({ queryKey: ['admin-groups'] })
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
      await deleteAdminGroup(deleteTarget)
      setDeleteTarget(null)
      await qc.invalidateQueries({ queryKey: ['admin-groups'] })
    } catch (err: unknown) {
      toast.error(extractErrorMessage(err, t('admin.deleteFailed')))
    } finally {
      setDeleting(null)
    }
  }, [deleteTarget, qc, t])

  const update = <K extends keyof GroupForm>(key: K, value: GroupForm[K]) =>
    setForm((prev) => ({ ...prev, [key]: value }))

  const toggleStrategy = (id: number) => {
    setForm((prev) => {
      const exists = prev.strategy_ids.includes(id)
      const nextStrategyIDs = exists
        ? prev.strategy_ids.filter((x) => x !== id)
        : [...prev.strategy_ids, id]
      return {
        ...prev,
        strategy_ids: nextStrategyIDs,
        default_strategy_id: exists && prev.default_strategy_id === id ? 0 : prev.default_strategy_id,
      }
    })
  }

  const inputCls = 'h-10 w-full rounded-lg border border-input bg-background px-3 text-sm outline-none transition-colors duration-150 focus:border-primary focus:ring-2 focus:ring-primary/20'
  const chipCls = (checked: boolean) =>
    `h-8 rounded-lg px-3 text-xs font-medium transition-colors border cursor-pointer ${
      checked
        ? 'bg-primary/10 text-primary border-primary/40 shadow-sm'
        : 'bg-background text-muted-foreground border-border hover:border-primary/40 hover:text-foreground'
    }`

  const getStrategyName = (id: number) => {
    const s = allStrategies.find((x) => x.id === id)
    return s ? s.name : String(id)
  }

  return (
    <section className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight">{t('admin.groupsTitle')}</h1>
          <p className="text-sm text-muted-foreground">{t('admin.groupsSubtitle', { defaultValue: '配置分组策略、格式限制与上传配额。' })}</p>
        </div>
        <button type="button" onClick={openCreate} className="h-10 rounded-lg bg-primary px-4 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors shadow-sm cursor-pointer">
          {t('admin.create')}
        </button>
      </div>

      {/* Group create/edit dialog */}
      <Dialog open={showModal} onOpenChange={setShowModal}>
        <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-[600px] border-border bg-card">
          <DialogHeader>
            <DialogTitle className="text-xl">
              {editing ? t('admin.edit') : t('admin.create')}
            </DialogTitle>
          </DialogHeader>

          <div className="grid gap-6 pt-4 sm:grid-cols-2">
            <div className="space-y-4">
              <div className="space-y-3">
                <label className="text-sm font-medium text-foreground">{t('admin.colName')}</label>
                <input value={form.name} onChange={(e) => update('name', e.target.value)} placeholder={t('admin.namePlaceholder')} className={inputCls} />
              </div>

              <div className="flex h-10 items-center justify-between rounded-lg border border-border bg-background px-3">
                <label htmlFor="isDefault" className="text-sm font-medium text-foreground cursor-pointer">{t('admin.defaultGroup', { defaultValue: '默认分组' })}</label>
                <Switch id="isDefault" checked={form.is_default} onCheckedChange={(checked) => update('is_default', checked)} />
              </div>

              <div className="space-y-3">
                <label className="text-sm font-medium text-foreground">{t('admin.maxFileSize', { defaultValue: '最大文件' })}</label>
                <div className="flex items-center gap-2">
                  <input type="number" min={1} value={form.max_size} onChange={(e) => update('max_size', Number(e.target.value))} className={`${inputCls} w-32`} />
                  <span className="text-sm text-muted-foreground">MB</span>
                </div>
              </div>

              <div className="space-y-3">
                <label className="text-sm font-medium text-foreground">{t('admin.limitPerDay', { defaultValue: '每日上限' })}</label>
                <input type="number" min={0} value={form.limit_per_day} onChange={(e) => update('limit_per_day', Number(e.target.value))} className={inputCls} />
              </div>

              <div className="space-y-3">
                <label className="text-sm font-medium text-foreground">{t('admin.limitPerMonth', { defaultValue: '每月上限' })}</label>
                <input type="number" min={0} value={form.limit_per_month} onChange={(e) => update('limit_per_month', Number(e.target.value))} className={inputCls} />
              </div>

              <div className="space-y-3 rounded-lg border border-border bg-muted/20 p-3">
                <label className="text-sm font-medium text-foreground">{t('admin.imageProcessing', { defaultValue: '图片处理' })}</label>
                <div className="space-y-3">
                  <div>
                    <label className="mb-1 block text-sm text-muted-foreground">{t('admin.imageSaveQuality', { defaultValue: '压缩质量' })}</label>
                    <div className="flex items-center gap-2">
                      <input
                        type="number"
                        min={1}
                        max={100}
                        value={form.image_save_quality}
                        onChange={(e) => update('image_save_quality', Number(e.target.value))}
                        className={`${inputCls} w-24`}
                      />
                      <span className="text-sm text-muted-foreground">1 - 100</span>
                    </div>
                  </div>
                  <div>
                    <label className="mb-1 block text-sm text-muted-foreground">{t('admin.imageSaveFormat', { defaultValue: '转码格式' })}</label>
                    <Select
                      value={form.image_save_format || 'origin'}
                      onValueChange={(val) => {
                        const next = String(val)
                        update('image_save_format', next === 'origin' ? '' : next)
                      }}
                      items={{
                        origin: t('admin.keepOriginalFormat', { defaultValue: '保持原格式' }),
                        jpeg: 'JPEG',
                        png: 'PNG',
                        webp: 'WebP',
                      }}
                    >
                      <SelectTrigger className="h-10 w-full">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="origin">{t('admin.keepOriginalFormat', { defaultValue: '保持原格式' })}</SelectItem>
                        <SelectItem value="jpeg">JPEG</SelectItem>
                        <SelectItem value="png">PNG</SelectItem>
                        <SelectItem value="webp">WebP</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="flex h-10 items-center justify-between rounded-lg border border-border bg-background px-3">
                    <label className="text-sm text-foreground cursor-pointer">{t('admin.stripExif', { defaultValue: '去除 EXIF 元数据' })}</label>
                    <Switch checked={form.is_strip_exif} onCheckedChange={(checked) => update('is_strip_exif', checked)} />
                  </div>
                </div>
              </div>
            </div>

            <div className="space-y-6">
              <div className="space-y-3">
                <label className="text-sm font-medium text-foreground">{t('admin.extensions', { defaultValue: '允许格式' })}</label>
                <div className="flex flex-wrap gap-2">
                  {commonExtensions.map((ext) => {
                    const exts = parseExtensions(form.extensions)
                    const checked = exts.includes(ext)
                    const toggleExt = () => {
                      const newExts = checked ? exts.filter(e => e !== ext) : [...exts, ext]
                      update('extensions', newExts.join(','))
                    }
                    return (
                      <button
                        key={ext} 
                        type="button"
                        onClick={toggleExt}
                        className={chipCls(checked)}
                      >
                        {ext}
                      </button>
                    )
                  })}
                </div>
                <div className="space-y-2 rounded-lg border border-dashed border-border/70 bg-muted/20 px-3 py-3 mt-2">
                  <label className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
                    {t('admin.extensionsCustom', { defaultValue: '自定义格式' })}
                  </label>
                  <input
                    value={parseExtensions(form.extensions).filter((ext) => !commonExtensions.includes(ext)).join(', ')}
                    onChange={(e) => {
                      const customExtensions = parseExtensions(e.target.value)
                      const selectedCommon = parseExtensions(form.extensions).filter((ext) => commonExtensions.includes(ext))
                      update('extensions', [...selectedCommon, ...customExtensions].join(','))
                    }}
                    placeholder={t('admin.extensionsCustomPlaceholder', { defaultValue: '例如：avif,heic,jxl' })}
                    className={`${inputCls} placeholder:text-muted-foreground/50`}
                  />
                </div>
              </div>

              {/* Strategy binding */}
              <div className="space-y-3">
                <label className="block text-sm font-medium text-foreground">{t('admin.availableStrategies', { defaultValue: '可用策略' })}</label>
                {allStrategies.length === 0 ? (
                  <p className="text-sm text-muted-foreground">{t('admin.noStrategies', { defaultValue: '暂无策略，请先创建策略' })}</p>
                ) : (
                  <div className="flex flex-wrap gap-2">
                    {allStrategies.map((s) => {
                      const checked = form.strategy_ids.includes(s.id)
                      return (
                        <button
                          key={s.id}
                          type="button"
                          onClick={() => toggleStrategy(s.id)}
                          className={chipCls(checked)}
                        >
                          {s.name}
                        </button>
                      )
                    })}
                  </div>
                )}
              </div>
              <div className="space-y-3">
                <label className="block text-sm font-medium text-foreground">
                  {t('admin.groupDefaultStrategy', { defaultValue: '分组默认策略' })}
                </label>
                <Select
                  value={String(form.default_strategy_id || 0)}
                  items={{
                    '0': t('admin.groupDefaultStrategyAuto', { defaultValue: '自动（第一个可用策略）' }),
                    ...Object.fromEntries(
                      allStrategies
                        .filter((s) => form.strategy_ids.includes(s.id))
                        .map((s) => [String(s.id), s.name]),
                    ),
                  }}
                  onValueChange={(val) => update('default_strategy_id', Number(val))}
                >
                  <SelectTrigger className="h-10 w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="0">
                      {t('admin.groupDefaultStrategyAuto', { defaultValue: '自动（第一个可用策略）' })}
                    </SelectItem>
                    {allStrategies
                      .filter((s) => form.strategy_ids.includes(s.id))
                      .map((s) => (
                        <SelectItem key={s.id} value={String(s.id)}>
                          {s.name}
                        </SelectItem>
                      ))}
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">
                  {t('admin.groupDefaultStrategyHint', { defaultValue: '当用户设置为“跟随分组默认”且上传未显式指定策略时，将使用这里的策略。' })}
                </p>
              </div>

              <div className="space-y-3 rounded-lg border border-border bg-muted/20 p-3">
                <div className="flex h-10 items-center justify-between rounded-lg border border-border bg-background px-3">
                  <label className="text-sm font-medium text-foreground cursor-pointer">{t('admin.enableWatermark', { defaultValue: '启用文字水印' })}</label>
                  <Switch checked={form.is_enable_watermark} onCheckedChange={(checked) => update('is_enable_watermark', checked)} />
                </div>

                {form.is_enable_watermark && (
                  <div className="space-y-3">
                    <div>
                      <label className="mb-1 block text-sm text-muted-foreground">{t('admin.watermarkText', { defaultValue: '水印文本' })}</label>
                      <input
                        value={form.watermark_text}
                        onChange={(e) => update('watermark_text', e.target.value)}
                        placeholder={t('admin.watermarkTextPlaceholder', { defaultValue: '例如：© PicFast' })}
                        className={inputCls}
                      />
                    </div>
                    <div>
                      <label className="mb-1 block text-sm text-muted-foreground">{t('admin.watermarkPosition', { defaultValue: '位置' })}</label>
                      <Select
                        value={form.watermark_position}
                        onValueChange={(val) => update('watermark_position', String(val))}
                        items={{
                          'bottom-right': t('admin.watermarkBottomRight', { defaultValue: '右下角' }),
                          'bottom-left': t('admin.watermarkBottomLeft', { defaultValue: '左下角' }),
                          'top-right': t('admin.watermarkTopRight', { defaultValue: '右上角' }),
                          'top-left': t('admin.watermarkTopLeft', { defaultValue: '左上角' }),
                          center: t('admin.watermarkCenter', { defaultValue: '居中' }),
                        }}
                      >
                        <SelectTrigger className="h-10 w-full">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="bottom-right">{t('admin.watermarkBottomRight', { defaultValue: '右下角' })}</SelectItem>
                          <SelectItem value="bottom-left">{t('admin.watermarkBottomLeft', { defaultValue: '左下角' })}</SelectItem>
                          <SelectItem value="top-right">{t('admin.watermarkTopRight', { defaultValue: '右上角' })}</SelectItem>
                          <SelectItem value="top-left">{t('admin.watermarkTopLeft', { defaultValue: '左上角' })}</SelectItem>
                          <SelectItem value="center">{t('admin.watermarkCenter', { defaultValue: '居中' })}</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    <div className="grid grid-cols-2 gap-3">
                      <div>
                        <label className="mb-1 block text-sm text-muted-foreground">{t('admin.watermarkFontSize', { defaultValue: '字号' })}</label>
                        <input
                          type="number"
                          min={8}
                          max={200}
                          value={form.watermark_font_size}
                          onChange={(e) => update('watermark_font_size', Number(e.target.value))}
                          className={inputCls}
                        />
                      </div>
                      <div>
                        <label className="mb-1 block text-sm text-muted-foreground">{t('admin.watermarkOpacity', { defaultValue: '透明度' })}</label>
                        <input
                          type="number"
                          min={0}
                          max={1}
                          step={0.1}
                          value={form.watermark_opacity}
                          onChange={(e) => update('watermark_opacity', Number(e.target.value))}
                          className={inputCls}
                        />
                      </div>
                    </div>
                    <div>
                      <label className="mb-1 block text-sm text-muted-foreground">{t('admin.watermarkColor', { defaultValue: '颜色' })}</label>
                      <input
                        value={form.watermark_color}
                        onChange={(e) => update('watermark_color', e.target.value)}
                        placeholder="#FFFFFF"
                        className={inputCls}
                      />
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>

          <div className="mt-2 flex justify-end gap-3 border-t border-border pt-5">
            <button type="button" onClick={() => setShowModal(false)} className="h-10 rounded-lg border border-input bg-background px-4 text-sm font-medium hover:bg-accent hover:text-accent-foreground transition-colors cursor-pointer">
              {t('admin.cancel')}
            </button>
            <button type="button" onClick={handleSave} disabled={saving || !form.name.trim()} className="h-10 rounded-lg bg-primary px-4 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50 transition-colors shadow-sm cursor-pointer">
              {saving ? '…' : t('admin.confirmSave')}
            </button>
          </div>
        </DialogContent>
      </Dialog>

      {isLoading && <LoadingState />}
      {error && <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">{t('admin.loadFailed')}</p>}
      {groups && groups.length === 0 && (
        <EmptyState
          icon={<FolderTree className="size-6 text-muted-foreground" />}
          title={t('admin.empty')}
          description={t('admin.groupsEmptyDesc', { defaultValue: '创建分组后，可为不同用户配置容量、格式与策略权限。' })}
        />
      )}

      {groups && groups.length > 0 && (
        <div className="overflow-x-auto rounded-xl border border-border bg-card/80 shadow-sm">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border text-left text-xs text-muted-foreground bg-muted/50">
                <th className="px-4 py-3 font-medium">{t('admin.colName')}</th>
                <th className="px-4 py-3 font-medium">{t('admin.userCount', { defaultValue: '用户数' })}</th>
                <th className="px-4 py-3 font-medium">{t('admin.maxFileSize', { defaultValue: '最大文件' })}</th>
                <th className="px-4 py-3 font-medium">{t('admin.limitPerDay', { defaultValue: '每日上限' })}</th>
                <th className="px-4 py-3 font-medium">{t('admin.availableStrategies', { defaultValue: '策略' })}</th>
                <th className="px-4 py-3 font-medium">{t('admin.colActions')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {groups.map((g) => (
                <tr key={g.id} className="group hover:bg-muted/50 transition-colors">
                  <td className="px-4 py-3 font-medium">
                    {g.name}
                    {g.is_default && <span className="ml-2 rounded-lg bg-primary/10 px-1.5 py-0.5 text-xs text-primary">{t('admin.default')}</span>}
                    {g.is_guest && <span className="ml-2 rounded-lg bg-success/10 px-1.5 py-0.5 text-xs text-success">{t('admin.guest')}</span>}
                  </td>
                  <td className="px-4 py-3 text-muted-foreground">{(g as unknown as Record<string, unknown>).user_count as number ?? '-'}</td>
                  <td className="px-4 py-3 text-muted-foreground">{formatSize((g.configs?.maximum_file_size as number) || 0)}</td>
                  <td className="px-4 py-3 text-muted-foreground">{(g.configs?.limit_per_day as number) || '-'}</td>
                  <td className="px-4 py-3">
                    <div className="flex flex-wrap gap-1">
                      {(g.strategy_ids || []).length > 0 ? (
                        g.strategy_ids.map((id) => (
                          <span key={id} className="rounded-lg bg-primary/10 px-1.5 py-0.5 text-xs text-primary">
                            {getStrategyName(id)}
                          </span>
                        ))
                      ) : (
                        <span className="text-xs text-muted-foreground">-</span>
                      )}
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex gap-1">
                      <button type="button" onClick={() => openEdit(g)} title={t('admin.edit')} className="rounded-lg p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground transition-colors cursor-pointer">
                        <Pencil className="size-4" />
                      </button>
                      {!g.is_default && !g.is_guest && (
                        <button type="button" onClick={() => setDeleteTarget(g.id)} disabled={deleting === g.id} title={t('admin.delete')} className="rounded-lg p-1.5 text-destructive/70 hover:bg-destructive/10 hover:text-destructive disabled:opacity-50 transition-colors cursor-pointer">
                          <Trash2 className="size-4" />
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <ConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}
        title={t('admin.confirmDeleteGroup')}
        description={t('admin.deleteGroupDescription')}
        confirmLabel={t('admin.delete')}
        onConfirm={onDelete}
        loading={!!deleting}
      />
    </section>
  )
}
