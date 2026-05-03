import { useState, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useSearchParams } from 'react-router-dom'
import { toast } from 'sonner'
import { Trash2, Image as ImageIcon, ChevronLeft, ChevronRight, Check, Download } from 'lucide-react'

import { deleteImage, getImage, listImages, updateImage, listAlbums } from '@/lib/console-api'
import type { ImageItem, Album } from '@/lib/console-api'
import { extractErrorMessage, logError } from '@/lib/error-handler'
import { formatFileSize } from '@/lib/upload'
import { ImageDetailDialog } from '@/components/console/image-detail-dialog'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { EmptyState, LoadingState } from '@/components/page-states'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Button } from '@/components/ui/button'

function toRelative(url: string): string {
  try { return new URL(url).pathname }
  catch { return url }
}

export function ImagesPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [searchParams, setSearchParams] = useSearchParams()
  const albumIdParam = searchParams.get('album_id')
  const albumId = albumIdParam ? Number(albumIdParam) : null
  const [page, setPage] = useState(1)
  const pageSize = 60

  // Reset page when album changes
  // Remove synchronous setState from useEffect to prevent cascading renders
  const prevAlbumIdRef = useRef(albumId)
  if (albumId !== prevAlbumIdRef.current) {
    setPage(1)
    prevAlbumIdRef.current = albumId
  }

  const { data, isLoading, error } = useQuery({
    queryKey: ['images', page, albumId],
    queryFn: () => listImages(page, pageSize, albumId),
  })
  const totalPages = data ? (data.total_pages > 0 ? data.total_pages : Math.max(1, Math.ceil(data.total / pageSize))) : 1

  // Detail modal
  const [detail, setDetail] = useState<ImageItem | null>(null)
  const [detailLoading, setDetailLoading] = useState(false)
  const [albums, setAlbums] = useState<Album[]>([])

  useEffect(() => {
    listAlbums(1, 100)
      .then((res) => setAlbums(res.items))
      .catch((err: unknown) => logError('images.loadAlbums', err))
  }, [])

  const showDetail = async (img: ImageItem) => {
    setDetailLoading(true)
    try {
      const full = await getImage(img.key)
      setDetail(full)
    } catch (err: unknown) {
      logError('images.loadDetail', err)
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
    } catch (err: unknown) {
      toast.error(extractErrorMessage(err, t('images.updateFailed', { defaultValue: '更新失败' })))
    }
  }

  const changeAlbum = async (albumId: string) => {
    if (!detail) return
    try {
      const id = albumId === 'none' ? null : Number(albumId)
      await updateImage(detail.key, { album_id: id ?? undefined })
      setDetail({ ...detail, album_id: id })
      await qc.invalidateQueries({ queryKey: ['images'] })
      toast.success(t('images.updateSuccess', { defaultValue: '更新成功' }))
    } catch (err: unknown) {
      toast.error(extractErrorMessage(err, t('images.updateFailed', { defaultValue: '更新失败' })))
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
      toast.error(extractErrorMessage(err, t('images.deleteFailed')))
    } finally {
      setDeleteLoading(false)
    }
  }

  // Batch mode
  const [batchMode, setBatchMode] = useState(false)
  const [selectedKeys, setSelectedKeys] = useState<Set<string>>(new Set())
  const [batchProcessing, setBatchProcessing] = useState(false)
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
    setBatchProcessing(true)
    let success = 0
    let failed = 0
    for (const key of selectedKeys) {
      try {
        await deleteImage(key)
        success++
      } catch (err: unknown) {
        logError('images.batchDelete', err)
        failed++
      }
    }
    setBatchProcessing(false)
    setShowBatchConfirm(false)
    setSelectedKeys(new Set())
    if (failed === 0) {
      toast.success(t('images.batchDeleteSuccess', { defaultValue: `成功删除 ${success} 张图片` }))
    } else {
      toast.error(t('images.batchDeletePartial', { defaultValue: `删除完成：成功 ${success} 张，失败 ${failed} 张` }))
    }
    await qc.invalidateQueries({ queryKey: ['images'] })
  }

  const batchDownload = async () => {
    if (!data) return
    setBatchProcessing(true)
    let success = 0
    for (const key of selectedKeys) {
      const img = data.items.find(i => i.key === key)
      if (img) {
        try {
          const url = toRelative(img.links?.url ?? img.url ?? '')
          const res = await fetch(url)
          const blob = await res.blob()
          const a = document.createElement('a')
          a.href = URL.createObjectURL(blob)
          a.download = img.origin_name || key
          a.click()
          URL.revokeObjectURL(a.href)
          success++
          await new Promise(resolve => setTimeout(resolve, 300)) // delay between downloads
        } catch (e) {
          console.error('Download failed for', key, e)
        }
      }
    }
    setBatchProcessing(false)
    setSelectedKeys(new Set())
    toast.success(t('images.batchDownloadSuccess', { defaultValue: `成功下载 ${success} 张图片` }))
  }

  const batchChangeAlbum = async (val: string) => {
    if (val === 'none') return
    setBatchProcessing(true)
    const targetAlbumId = Number(val)
    let success = 0
    let failed = 0
    for (const key of selectedKeys) {
      try {
        await updateImage(key, { album_id: targetAlbumId ?? undefined })
        success++
      } catch (err: unknown) {
        logError('images.batchMove', err)
        failed++
      }
    }
    setBatchProcessing(false)
    setSelectedKeys(new Set())
    if (failed === 0) {
      toast.success(t('images.batchMoveSuccess', { defaultValue: `成功移动 ${success} 张图片` }))
    } else {
      toast.error(t('images.batchMovePartial', { defaultValue: `移动完成：成功 ${success} 张，失败 ${failed} 张` }))
    }
    await qc.invalidateQueries({ queryKey: ['images'] })
  }

  const onCopy = async (text: string) => {
    await navigator.clipboard.writeText(text)
    toast.success(t('upload.copied'))
  }

  const currentAlbum = albumId ? albums.find(a => a.id === albumId) : null

  return (
    <section className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="flex flex-col gap-1.5">
          <h1 className="text-2xl font-bold tracking-tight">
            {t('page.images.title')}
          </h1>
          {currentAlbum && (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <span>{t('albums.filteredBy', { defaultValue: '相册筛选：' })}</span>
              <span className="rounded-md bg-primary/10 px-2 py-0.5 text-primary font-medium">{currentAlbum.name}</span>
            </div>
          )}
        </div>
        <div className="flex items-center gap-3">
            {data && <span className="text-sm font-medium text-muted-foreground">{t('images.pagination', { total: data.total })}</span>}
            {currentAlbum && (
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  searchParams.delete('album_id')
                  setSearchParams(searchParams)
                }}
              >
                {t('albums.backToAll', { defaultValue: '返回全部' })}
              </Button>
            )}
            {!batchMode ? (
            <Button type="button" variant="outline" onClick={() => setBatchMode(true)}>
              {t('images.batchManage', { defaultValue: '批量管理' })}
            </Button>
          ) : (
            <Button type="button" variant="outline" onClick={exitBatch}>
              {t('images.exitBatch', { defaultValue: '退出管理' })}
            </Button>
          )}
        </div>
      </div>

      {/* Batch bar */}
      {batchMode && (
        <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-primary/20 bg-primary/[0.08] px-3 py-2.5">
          <button
            type="button"
            onClick={toggleSelectAll}
            className="group inline-flex items-center gap-3 text-sm font-medium text-primary transition-colors hover:text-primary/80"
          >
            <span className={`inline-flex size-6 items-center justify-center rounded-full border transition-colors ${data && selectedKeys.size === data.items.length && data.items.length > 0 ? 'border-primary bg-primary text-primary-foreground' : 'border-primary/40 bg-background text-transparent'}`}>
              <Check className="size-3.5" />
            </span>
            <span>{t('images.selectAll', { defaultValue: '全选' })} ({selectedKeys.size} / {data?.items.length ?? 0})</span>
          </button>
          <div className="flex flex-wrap items-center gap-2">
            <Select
              value="none"
              items={{
                none: t('images.batchMove', { defaultValue: '批量移动至...' }),
                ...Object.fromEntries(albums.map((a) => [a.id.toString(), a.name])),
              }}
              onValueChange={(val) => val !== null && batchChangeAlbum(val as string)}
            >
              <SelectTrigger className="h-8 w-[160px] text-xs font-medium border-primary/20 bg-background hover:bg-muted" disabled={selectedKeys.size === 0 || batchProcessing}>
                <SelectValue placeholder={t('images.batchMove', { defaultValue: '批量移动至...' })} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="none" disabled>{t('images.batchMove', { defaultValue: '批量移动至...' })}</SelectItem>
                {albums.map(a => (
                  <SelectItem key={a.id} value={a.id.toString()}>{a.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button type="button" size="sm" variant="outline" onClick={batchDownload} disabled={selectedKeys.size === 0 || batchProcessing}>
              <Download className="size-3.5" />
              {t('images.batchDownload', { defaultValue: '下载' })}
            </Button>
            <Button type="button" size="sm" variant="destructive" onClick={() => setShowBatchConfirm(true)} disabled={selectedKeys.size === 0 || batchProcessing}>
              <Trash2 className="size-3.5" />
              {batchProcessing ? '…' : t('images.delete')}
            </Button>
          </div>
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
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 md:gap-5 lg:grid-cols-5 xl:grid-cols-6">
            {data.items.map((img) => (
              <div
                key={img.id}
                className={`group cursor-pointer overflow-hidden rounded-xl border bg-card transition-colors duration-150 hover:shadow-sm hover:border-primary/30 flex flex-col ${batchMode && selectedKeys.has(img.key) ? 'border-primary/70 bg-primary/[0.04] ring-2 ring-primary/25' : 'border-border/50'}`}
                onClick={() => batchMode ? toggleSelect(img.key) : showDetail(img)}
              >
                <div className="relative aspect-[4/3] flex items-center justify-center overflow-hidden bg-muted/30">
                  {img.thumbnail_url || img.links?.thumbnail_url ? (
                    <>
                      <img
                        src={toRelative(img.thumbnail_url || img.links?.thumbnail_url || '')}
                        alt=""
                        className="h-full w-full object-cover"
                        loading="lazy"
                        onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }}
                      />
                      <div className="absolute inset-0 bg-black/0 transition-colors duration-150 group-hover:bg-black/10" />
                    </>
                  ) : (
                    <span className="text-xs font-medium text-muted-foreground">{img.extension.toUpperCase()}</span>
                  )}
                  {batchMode && (
                    <div className="absolute right-2 top-2 z-10">
                      <button
                        type="button"
                        onClick={(e) => {
                          e.stopPropagation()
                          toggleSelect(img.key)
                        }}
                        aria-label={selectedKeys.has(img.key) ? t('images.exitBatch', { defaultValue: '取消选中' }) : t('images.selectAll', { defaultValue: '选中' })}
                        className={`flex size-5 items-center justify-center rounded-full border shadow-sm transition-all ${
                          selectedKeys.has(img.key)
                            ? 'border-primary bg-primary text-primary-foreground'
                            : 'border-white/50 bg-black/20 text-white backdrop-blur-md hover:bg-black/40'
                        }`}
                      >
                        <Check className={`size-3.5 ${!selectedKeys.has(img.key) ? 'opacity-0' : ''}`} />
                      </button>
                    </div>
                  )}
                  {img.permission === 0 && (
                    <div className="absolute right-2 top-2 rounded-lg bg-amber-500/90 px-2 py-0.5 text-[10px] font-medium text-white backdrop-blur-sm shadow-sm">
                      {t('images.private', { defaultValue: '私有' })}
                    </div>
                  )}
                  {img.moderation_status === 'pending' && (
                    <div className="absolute left-2 top-2 rounded-lg bg-warning/90 px-2 py-0.5 text-[10px] font-medium text-white backdrop-blur-sm shadow-sm">
                      {t('images.moderationPending', { defaultValue: '待审核' })}
                    </div>
                  )}
                  {img.moderation_status === 'rejected' && (
                    <div className="absolute left-2 top-2 rounded-lg bg-destructive/90 px-2 py-0.5 text-[10px] font-medium text-white backdrop-blur-sm shadow-sm">
                      {t('images.moderationRejected', { defaultValue: '审核拒绝' })}
                    </div>
                  )}
                  {!batchMode && (
                    <button
                      type="button"
                      className="absolute right-2 top-2 z-10 inline-flex size-7 items-center justify-center rounded-full bg-black/40 text-white/80 opacity-0 backdrop-blur-sm transition-all duration-200 hover:bg-destructive hover:text-white group-hover:opacity-100"
                      onClick={(e) => {
                        e.stopPropagation()
                        setDeleteTarget(img.key)
                      }}
                      title={t('images.delete', { defaultValue: '删除' })}
                    >
                      <Trash2 className="size-3.5" />
                    </button>
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
                <Button type="button" variant="outline" size="icon" disabled={page <= 1} onClick={() => setPage((p) => p - 1)} title={t('images.prev')}>
                  <ChevronLeft className="size-4" />
                </Button>
                <span className="inline-flex h-8 min-w-[56px] items-center justify-center rounded-lg border border-input bg-background px-2 text-xs text-muted-foreground">
                  {page} / {totalPages}
                </span>
                <Button type="button" variant="outline" size="icon" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)} title={t('images.next')}>
                  <ChevronRight className="size-4" />
                </Button>
              </div>
            </div>
          )}
        </>
      )}

      <ImageDetailDialog
        image={detail}
        loading={detailLoading}
        albums={albums}
        onClose={() => setDetail(null)}
        onTogglePermission={togglePermission}
        onChangeAlbum={changeAlbum}
        onCopy={onCopy}
        onDelete={setDeleteTarget}
      />

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
        loading={batchProcessing}
      />
    </section>
  )
}
