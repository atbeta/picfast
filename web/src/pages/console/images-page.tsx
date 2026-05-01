import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Trash2, Copy, Image as ImageIcon, ChevronLeft, ChevronRight } from 'lucide-react'

import { deleteImage, getImage, listImages, updateImage } from '../../lib/console-api'
import type { ImageItem } from '../../lib/console-api'
import { formatFileSize } from '../../lib/upload'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { EmptyState, LoadingState } from '@/components/page-states'
import { Switch } from '@/components/ui/switch'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

function toRelative(url: string): string {
  try { return new URL(url).pathname }
  catch { return url }
}

export function ImagesPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [page, setPage] = useState(1)
  const pageSize = 20

  const { data, isLoading, error } = useQuery({
    queryKey: ['images', page],
    queryFn: () => listImages(page, pageSize),
  })

  // Detail modal
  const [detail, setDetail] = useState<ImageItem | null>(null)
  const [detailLoading, setDetailLoading] = useState(false)

  const showDetail = async (img: ImageItem) => {
    setDetailLoading(true)
    try {
      const full = await getImage(img.key)
      setDetail(full)
    } catch {
      setDetail(img)
    } finally {
      setDetailLoading(false)
    }
  }

  // Permission toggle
  const togglePermission = async () => {
    if (!detail) return
    try {
      const newPerm = detail.permission === 1 ? 0 : 1
      await updateImage(detail.key, { permission: newPerm })
      setDetail({ ...detail, permission: newPerm })
      await qc.invalidateQueries({ queryKey: ['images'] })
    } catch {
      // silently fail
    }
  }

  // Delete confirmation
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)
  const [deleteLoading, setDeleteLoading] = useState(false)

  const confirmDelete = async () => {
    if (!deleteTarget) return
    setDeleteLoading(true)
    try {
      await deleteImage(deleteTarget)
      setDetail(null)
      setDeleteTarget(null)
      await qc.invalidateQueries({ queryKey: ['images'] })
    } catch (err: unknown) {
      toast.error((err as { response?: { data?: { message?: string } } })?.response?.data?.message || t('images.deleteFailed'))
    } finally {
      setDeleteLoading(false)
    }
  }

  // Batch mode
  const [batchMode, setBatchMode] = useState(false)
  const [selectedKeys, setSelectedKeys] = useState<Set<string>>(new Set())
  const [batchDeleting, setBatchDeleting] = useState(false)
  const [showBatchConfirm, setShowBatchConfirm] = useState(false)

  const toggleSelect = (key: string) => {
    setSelectedKeys((prev) => {
      const next = new Set(prev)
      if (next.has(key)) next.delete(key)
      else next.add(key)
      return next
    })
  }

  const toggleSelectAll = () => {
    if (!data) return
    if (selectedKeys.size === data.items.length) {
      setSelectedKeys(new Set())
    } else {
      setSelectedKeys(new Set(data.items.map((img) => img.key)))
    }
  }

  const exitBatch = () => {
    setBatchMode(false)
    setSelectedKeys(new Set())
  }

  const batchDelete = async () => {
    setBatchDeleting(true)
    let success = 0
    let failed = 0
    for (const key of selectedKeys) {
      try { await deleteImage(key); success++ } catch { failed++ }
    }
    setBatchDeleting(false)
    setShowBatchConfirm(false)
    setSelectedKeys(new Set())
    if (failed === 0) {
      toast.success(t('images.batchDeleteSuccess', { defaultValue: `成功删除 ${success} 张图片` }))
    } else {
      toast.error(t('images.batchDeletePartial', { defaultValue: `删除完成：成功 ${success} 张，失败 ${failed} 张` }))
    }
    await qc.invalidateQueries({ queryKey: ['images'] })
  }

  const onCopy = async (text: string) => {
    await navigator.clipboard.writeText(text)
    toast.success(t('upload.copied'))
  }

  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight">{t('page.images.title')}</h1>
        <div className="flex items-center gap-3">
          {data && <span className="text-sm font-medium text-muted-foreground">{t('images.pagination', { total: data.total })}</span>}
          {!batchMode ? (
            <button type="button" onClick={() => setBatchMode(true)} className="rounded-lg border border-border/50 bg-background px-3 py-1.5 text-sm font-medium shadow-sm transition-colors hover:bg-muted hover:text-foreground">
              {t('images.batchManage', { defaultValue: '批量管理' })}
            </button>
          ) : (
            <button type="button" onClick={exitBatch} className="rounded-lg border border-border/50 bg-background px-3 py-1.5 text-sm font-medium shadow-sm transition-colors hover:bg-muted hover:text-foreground">
              {t('images.exitBatch', { defaultValue: '退出管理' })}
            </button>
          )}
        </div>
      </div>

      {/* Batch bar */}
      {batchMode && (
        <div className="flex items-center justify-between rounded-xl bg-primary/10 border border-primary/20 px-4 py-3">
          <label className="flex items-center gap-2 text-sm font-medium cursor-pointer group">
            <Switch checked={data ? selectedKeys.size === data.items.length && data.items.length > 0 : false} onCheckedChange={toggleSelectAll} />
            <span className="text-primary group-hover:text-primary/80 transition-colors">{t('images.selectAll', { defaultValue: '全选' })} ({selectedKeys.size} / {data?.items.length ?? 0})</span>
          </label>
          <button type="button" onClick={() => setShowBatchConfirm(true)} disabled={selectedKeys.size === 0 || batchDeleting} className="flex items-center gap-1.5 rounded-lg bg-destructive px-3 py-1.5 text-xs font-medium text-destructive-foreground shadow-sm transition-colors hover:opacity-90 disabled:opacity-50 cursor-pointer">
            <Trash2 className="size-3.5" />
            {batchDeleting ? '…' : t('images.delete')}
          </button>
        </div>
      )}

      {isLoading && (
        <LoadingState />
      )}
      {error && (
        <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {t('images.loadFailed')}
        </p>
      )}
      {data && data.items.length === 0 && (
        <EmptyState
          icon={<ImageIcon className="size-6 text-muted-foreground" />}
          title={t('images.empty')}
          description={t('images.emptyDesc', { defaultValue: '上传后的图片会出现在这里。' })}
        />
      )}

      {data && data.items.length > 0 && (
        <>
          {/* Grid view */}
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
            {data.items.map((img) => (
              <div
                key={img.id}
                className="group cursor-pointer overflow-hidden rounded-xl border border-border/50 bg-card transition-all duration-300 hover:shadow-md hover:-translate-y-1 hover:border-primary/30"
                onClick={() => batchMode ? toggleSelect(img.key) : showDetail(img)}
              >
                <div className="relative aspect-square flex items-center justify-center overflow-hidden bg-muted/30">
                  {img.thumbnail_url || img.links?.thumbnail_url ? (
                    <>
                      <img
                        src={toRelative(img.thumbnail_url || img.links?.thumbnail_url || '')}
                        alt=""
                        className="h-full w-full object-cover transition-transform duration-500 group-hover:scale-110"
                        loading="lazy"
                        onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }}
                      />
                      <div className="absolute inset-0 bg-black/0 transition-colors duration-300 group-hover:bg-black/10" />
                    </>
                  ) : (
                    <span className="text-xs font-medium text-muted-foreground transition-transform duration-500 group-hover:scale-110">{img.extension.toUpperCase()}</span>
                  )}
                  {batchMode && (
                    <div className="absolute left-2 top-2 z-10" onClick={(e) => e.stopPropagation()}>
                      <Switch checked={selectedKeys.has(img.key)} onCheckedChange={() => toggleSelect(img.key)} />
                    </div>
                  )}
                  {img.permission === 0 && (
                    <div className="absolute right-2 top-2 rounded-lg bg-amber-500/90 px-2 py-0.5 text-[10px] font-medium text-white backdrop-blur-sm shadow-sm">
                      {t('images.private', { defaultValue: '私有' })}
                    </div>
                  )}
                </div>
                <div className="p-3">
                  <p className="truncate text-sm font-medium text-foreground transition-colors group-hover:text-primary">{img.origin_name}</p>
                  <div className="mt-1.5 flex items-center justify-between">
                    <span className="rounded-lg bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">{formatFileSize(img.size_bytes)}</span>
                    {img.strategy_name && (
                      <span className="rounded-lg bg-primary/10 px-1.5 py-0.5 text-[10px] font-medium text-primary">{img.strategy_name}</span>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>

          {data.total > pageSize && (
            <div className="flex items-center justify-between pt-2">
              <span className="text-xs text-muted-foreground">
                {t('images.pagination', { total: data.total })}
              </span>
              <div className="flex gap-2">
                <button type="button" disabled={page <= 1} onClick={() => setPage((p) => p - 1)} title={t('images.prev')} className="inline-flex h-8 w-8 items-center justify-center rounded-lg border border-input bg-background text-xs font-medium shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed">
                  <ChevronLeft className="size-4" />
                </button>
                <button type="button" disabled={page * pageSize >= data.total} onClick={() => setPage((p) => p + 1)} title={t('images.next')} className="inline-flex h-8 w-8 items-center justify-center rounded-lg border border-input bg-background text-xs font-medium shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed">
                  <ChevronRight className="size-4" />
                </button>
              </div>
            </div>
          )}
        </>
      )}

      {/* Detail dialog */}
      <Dialog open={!!detail} onOpenChange={(open) => { if (!open) setDetail(null) }}>
        <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-xl">
          <DialogHeader>
            <DialogTitle>{t('images.detailTitle', { defaultValue: '图片详情' })}</DialogTitle>
          </DialogHeader>

          {detailLoading && <LoadingState compact className="py-6" />}

          {detail && (
            <>
              {/* Preview */}
              <div className="flex justify-center rounded-xl bg-muted/30 border border-border/50 p-4">
                <img
                  src={toRelative(detail.links?.url ?? detail.url ?? '')}
                  alt={detail.key}
                  className="max-h-[40vh] rounded-lg object-contain shadow-sm"
                  onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }}
                />
              </div>

              {/* Metadata */}
              <div className="grid grid-cols-2 gap-4 text-sm rounded-xl border border-border/50 bg-card p-4">
                <div className="flex flex-col gap-1"><span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">Key</span> <span className="font-mono text-foreground break-all">{detail.key}</span></div>
                <div className="flex flex-col gap-1"><span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('images.colName')}</span> <span className="text-foreground truncate" title={detail.origin_name}>{detail.origin_name}</span></div>
                <div className="flex flex-col gap-1"><span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('images.colSize')}</span> <span className="text-foreground">{formatFileSize(detail.size_bytes)}</span></div>
                <div className="flex flex-col gap-1"><span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('images.type', { defaultValue: '类型' })}</span> <span className="text-foreground">{detail.mimetype}</span></div>
                <div className="flex flex-col gap-1"><span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('images.dimensions', { defaultValue: '尺寸' })}</span> <span className="text-foreground">{detail.width}x{detail.height}</span></div>
                <div className="flex flex-col gap-1 items-start">
                  <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('images.permission', { defaultValue: '权限' })}</span>
                  <button
                    type="button"
                    onClick={togglePermission}
                    className={['rounded-lg px-2 py-0.5 text-[10px] font-medium tracking-wide uppercase transition-colors cursor-pointer', detail.permission === 1 ? 'bg-success/10 text-success hover:bg-success/20' : 'bg-warning/10 text-warning hover:bg-warning/20'].join(' ')}
                  >
                    {detail.permission === 1 ? (t('images.public', { defaultValue: '公开' })) : (t('images.private', { defaultValue: '私有' }))}
                  </button>
                </div>
                {detail.strategy_name && (
                  <div className="col-span-2 flex flex-col gap-1">
                    <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('images.strategy', { defaultValue: '存储策略' })}</span>
                    <span className="self-start rounded-lg bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
                      {detail.strategy_name} ({detail.strategy_type === 'local' ? t('admin.typeLocal', { defaultValue: '本地' }) : 'S3'})
                    </span>
                  </div>
                )}
              </div>

              {/* Links */}
              {detail.links && (
                <div className="space-y-2">
                  {Object.entries(detail.links).map(([fmt, val]) => (
                    <div key={fmt} className="group flex items-center gap-3 rounded-xl bg-muted/40 px-4 py-2.5 transition-colors hover:bg-muted/60 border border-transparent hover:border-border/50">
                      <span className="shrink-0 w-20 text-xs font-semibold tracking-wider text-muted-foreground uppercase">{fmt}</span>
                      <code className="min-w-0 flex-1 truncate text-xs font-medium text-foreground bg-background/50 px-2 py-1.5 rounded-lg border border-border/30">{val}</code>
                      <button type="button" onClick={() => onCopy(val)} className="shrink-0 flex items-center justify-center h-7 w-7 rounded-lg bg-background border border-border/50 text-muted-foreground hover:bg-primary hover:text-primary-foreground hover:border-primary transition-all duration-200 shadow-sm cursor-pointer" title={t('upload.copy')}>
                        <Copy className="size-3.5" />
                      </button>
                    </div>
                  ))}
                </div>
              )}

              <div className="flex justify-end pt-4 border-t border-border/50">
                <button type="button" onClick={() => setDeleteTarget(detail.key)} className="flex items-center gap-1.5 rounded-lg px-3 py-2 text-xs font-medium text-destructive transition-colors hover:bg-destructive/10 cursor-pointer">
                  <Trash2 className="size-4" />
                  {t('images.delete')}
                </button>
              </div>
            </>
          )}
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}
        title={t('images.confirmDelete')}
        description={t('images.deleteDescription')}
        confirmLabel={t('images.delete')}
        onConfirm={confirmDelete}
        loading={deleteLoading}
      />

      <ConfirmDialog
        open={showBatchConfirm}
        onOpenChange={setShowBatchConfirm}
        title={t('images.confirmBatchDelete', { defaultValue: `确定要删除选中的 ${selectedKeys.size} 张图片吗？此操作不可撤销。` })}
        description={t('images.batchDeleteDescription', { count: selectedKeys.size })}
        confirmLabel={t('images.delete')}
        onConfirm={batchDelete}
        loading={batchDeleting}
      />
    </section>
  )
}
