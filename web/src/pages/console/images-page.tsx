import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'

import { deleteImage, getImage, listImages, updateImage } from '../../lib/console-api'
import type { ImageItem } from '../../lib/console-api'
import { formatFileSize } from '../../lib/upload'
import { ConfirmDialog } from '@/components/confirm-dialog'
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
        <h1 className="text-xl font-semibold">{t('page.images.title')}</h1>
        <div className="flex items-center gap-3">
          {data && <span className="text-sm text-zinc-400">{t('images.pagination', { total: data.total })}</span>}
          {!batchMode ? (
            <button type="button" onClick={() => setBatchMode(true)} className="rounded px-2 py-1 text-xs hover:bg-zinc-100 dark:hover:bg-zinc-800">
              {t('images.batchManage', { defaultValue: '批量管理' })}
            </button>
          ) : (
            <button type="button" onClick={exitBatch} className="rounded px-2 py-1 text-xs hover:bg-zinc-100 dark:hover:bg-zinc-800">
              {t('images.exitBatch', { defaultValue: '退出管理' })}
            </button>
          )}
        </div>
      </div>

      {/* Batch bar */}
      {batchMode && (
        <div className="flex items-center justify-between rounded-lg bg-blue-50 px-4 py-3 dark:bg-blue-900/20">
          <label className="flex items-center gap-2 text-sm">
            <input type="checkbox" checked={data ? selectedKeys.size === data.items.length && data.items.length > 0 : false} onChange={toggleSelectAll} className="h-4 w-4 rounded border-zinc-300 dark:border-zinc-600" />
            {t('images.selectAll', { defaultValue: '全选' })} ({selectedKeys.size} / {data?.items.length ?? 0})
          </label>
          <button type="button" onClick={() => setShowBatchConfirm(true)} disabled={selectedKeys.size === 0 || batchDeleting} className="rounded bg-red-500 px-3 py-1 text-xs text-white disabled:opacity-50">
            {batchDeleting ? '…' : t('images.delete')}
          </button>
        </div>
      )}

      {isLoading && (
        <div className="flex justify-center py-12">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-400 border-t-transparent" />
        </div>
      )}
      {error && (
        <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-400">
          {t('images.loadFailed')}
        </p>
      )}
      {data && data.items.length === 0 && (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div className="mb-3 rounded-full bg-muted p-3">
            <svg className="size-6 text-muted-foreground" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" d="m2.25 15.75 5.159-5.159a2.25 2.25 0 0 1 3.182 0l5.159 5.159m-1.5-1.5 1.409-1.409a2.25 2.25 0 0 1 3.182 0l2.909 2.909m-18 3.75h16.5a1.5 1.5 0 0 0 1.5-1.5V6a1.5 1.5 0 0 0-1.5-1.5H3.75A1.5 1.5 0 0 0 2.25 6v12a1.5 1.5 0 0 0 1.5 1.5Zm10.5-11.25h.008v.008h-.008V8.25Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Z" /></svg>
          </div>
          <p className="text-sm font-medium text-muted-foreground">{t('images.empty')}</p>
        </div>
      )}

      {data && data.items.length > 0 && (
        <>
          {/* Grid view */}
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
            {data.items.map((img) => (
              <div
                key={img.id}
                className="group cursor-pointer overflow-hidden rounded-lg border border-zinc-200 bg-white transition-shadow hover:shadow-md dark:border-zinc-700 dark:bg-zinc-800"
                onClick={() => batchMode ? toggleSelect(img.key) : showDetail(img)}
              >
                <div className="relative aspect-square flex items-center justify-center overflow-hidden bg-slate-50 dark:bg-zinc-900">
                  {img.thumbnail_url || img.links?.thumbnail_url ? (
                    <img
                      src={toRelative(img.thumbnail_url || img.links?.thumbnail_url || '')}
                      alt=""
                      className="h-full w-full object-cover"
                      loading="lazy"
                      onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }}
                    />
                  ) : (
                    <span className="text-xs text-zinc-400">{img.extension.toUpperCase()}</span>
                  )}
                  {batchMode && (
                    <div className="absolute left-2 top-2 z-10" onClick={(e) => e.stopPropagation()}>
                      <input type="checkbox" checked={selectedKeys.has(img.key)} onChange={() => toggleSelect(img.key)} className="h-4 w-4 rounded border-zinc-300 dark:border-zinc-600" />
                    </div>
                  )}
                  {img.permission === 0 && (
                    <div className="absolute right-2 top-2 rounded bg-amber-500/80 px-1.5 py-0.5 text-[10px] text-white">
                      {t('images.private', { defaultValue: '私有' })}
                    </div>
                  )}
                </div>
                <div className="px-2.5 py-2">
                  <p className="max-w-[120px] truncate text-xs font-medium">{img.origin_name}</p>
                  <div className="mt-0.5 flex items-center justify-between">
                    <span className="text-[10px] text-zinc-400">{formatFileSize(img.size_bytes)}</span>
                    {img.strategy_name && (
                      <span className="rounded bg-blue-100 px-1 py-0.5 text-[10px] text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">{img.strategy_name}</span>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>

          {data.total > pageSize && (
            <div className="flex items-center justify-between pt-2">
              <span className="text-xs text-zinc-500">
                {t('images.pagination', { total: data.total })}
              </span>
              <div className="flex gap-2">
                <button type="button" disabled={page <= 1} onClick={() => setPage((p) => p - 1)} className="rounded border border-zinc-300 px-3 py-1 text-xs disabled:opacity-40 dark:border-zinc-700">{t('images.prev')}</button>
                <button type="button" disabled={page * pageSize >= data.total} onClick={() => setPage((p) => p + 1)} className="rounded border border-zinc-300 px-3 py-1 text-xs disabled:opacity-40 dark:border-zinc-700">{t('images.next')}</button>
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

          {detailLoading && <div className="flex justify-center py-4"><div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-400 border-t-transparent" /></div>}

          {detail && (
            <>
              {/* Preview */}
              <div className="flex justify-center rounded-lg bg-slate-50 p-3 dark:bg-zinc-900">
                <img
                  src={toRelative(detail.links?.url ?? detail.url ?? '')}
                  alt={detail.key}
                  className="max-h-72 object-contain"
                  onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }}
                />
              </div>

              {/* Metadata */}
              <div className="grid grid-cols-2 gap-2 text-sm">
                <div><span className="text-zinc-400">Key:</span> <span className="break-all">{detail.key}</span></div>
                <div><span className="text-zinc-400">{t('images.colName')}:</span> {detail.origin_name}</div>
                <div><span className="text-zinc-400">{t('images.colSize')}:</span> {formatFileSize(detail.size_bytes)}</div>
                <div><span className="text-zinc-400">{t('images.type', { defaultValue: '类型' })}:</span> {detail.mimetype}</div>
                <div><span className="text-zinc-400">{t('images.dimensions', { defaultValue: '尺寸' })}:</span> {detail.width}x{detail.height}</div>
                <div className="flex items-center gap-2">
                  <span className="text-zinc-400">{t('images.permission', { defaultValue: '权限' })}:</span>
                  <button
                    type="button"
                    onClick={togglePermission}
                    className={['rounded px-2 py-0.5 text-xs', detail.permission === 1 ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'].join(' ')}
                  >
                    {detail.permission === 1 ? (t('images.public', { defaultValue: '公开' })) : (t('images.private', { defaultValue: '私有' }))}
                  </button>
                </div>
                {detail.strategy_name && (
                  <div className="col-span-2">
                    <span className="text-zinc-400">{t('images.strategy', { defaultValue: '存储策略' })}:</span>
                    <span className="ml-1 rounded bg-blue-100 px-1.5 py-0.5 text-xs text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">
                      {detail.strategy_name} ({detail.strategy_type === 'local' ? t('admin.typeLocal', { defaultValue: '本地' }) : 'S3'})
                    </span>
                  </div>
                )}
              </div>

              {/* Links */}
              {detail.links && (
                <div className="space-y-2">
                  {Object.entries(detail.links).map(([fmt, val]) => (
                    <div key={fmt} className="flex items-center gap-2">
                      <input value={val} readOnly className="min-w-0 flex-1 rounded border border-zinc-200 bg-zinc-50 px-2 py-1 text-xs dark:border-zinc-700 dark:bg-zinc-900" />
                      <button type="button" onClick={() => onCopy(val)} className="shrink-0 rounded bg-zinc-100 px-2 py-1 text-xs hover:bg-zinc-200 dark:bg-zinc-700 dark:hover:bg-zinc-600">{fmt}</button>
                    </div>
                  ))}
                </div>
              )}

              <div className="flex justify-end">
                <button type="button" onClick={() => setDeleteTarget(detail.key)} className="rounded px-2 py-1 text-xs text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20">{t('images.delete')}</button>
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
