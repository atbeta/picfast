import { useState, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useSearchParams } from 'react-router-dom'
import { toast } from 'sonner'
import { Trash2, Copy, Image as ImageIcon, ChevronLeft, ChevronRight, Check, Download } from 'lucide-react'

import { deleteImage, getImage, listImages, updateImage, listAlbums } from '../../lib/console-api'
import type { ImageItem, Album } from '../../lib/console-api'
import { formatFileSize } from '../../lib/upload'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { EmptyState, LoadingState } from '@/components/page-states'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

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
  const pageSize = 20

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

  // Detail modal
  const [detail, setDetail] = useState<ImageItem | null>(null)
  const [detailLoading, setDetailLoading] = useState(false)
  const [albums, setAlbums] = useState<Album[]>([])

  useEffect(() => {
    listAlbums(1, 100).then(res => setAlbums(res.items)).catch(() => {})
  }, [])

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

  const changeAlbum = async (albumId: string) => {
    if (!detail) return
    try {
      const id = albumId === 'none' ? null : Number(albumId)
      await updateImage(detail.key, { album_id: id ?? undefined })
      setDetail({ ...detail, album_id: id })
      await qc.invalidateQueries({ queryKey: ['images'] })
      toast.success(t('images.updateSuccess', { defaultValue: '更新成功' }))
    } catch {
      toast.error(t('images.updateFailed', { defaultValue: '更新失败' }))
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
      try { await deleteImage(key); success++ } catch { failed++ }
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
      } catch {
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
              <button
                type="button"
                onClick={() => {
                  searchParams.delete('album_id')
                  setSearchParams(searchParams)
                }}
                className="rounded-lg border border-border/50 bg-background px-3 py-1.5 text-sm font-medium shadow-sm transition-colors hover:bg-muted hover:text-foreground"
              >
                {t('albums.backToAll', { defaultValue: '返回全部' })}
              </button>
            )}
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
            <button type="button" onClick={batchDownload} disabled={selectedKeys.size === 0 || batchProcessing} className="flex h-8 items-center gap-1.5 rounded-lg border border-primary/20 bg-background px-3 text-xs font-medium text-primary shadow-sm transition-colors hover:bg-primary/10 disabled:opacity-50 cursor-pointer">
              <Download className="size-3.5" />
              {t('images.batchDownload', { defaultValue: '下载' })}
            </button>
            <button type="button" onClick={() => setShowBatchConfirm(true)} disabled={selectedKeys.size === 0 || batchProcessing} className="flex h-8 items-center gap-1.5 rounded-lg border border-destructive/30 bg-destructive/10 px-3 text-xs font-medium text-destructive shadow-sm transition-colors hover:bg-destructive/15 disabled:opacity-50 cursor-pointer">
              <Trash2 className="size-3.5" />
              {batchProcessing ? '…' : t('images.delete')}
            </button>
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
                className={`group cursor-pointer overflow-hidden rounded-xl border bg-card transition-colors duration-150 hover:shadow-sm hover:border-primary/30 ${batchMode && selectedKeys.has(img.key) ? 'border-primary/70 bg-primary/[0.04] ring-2 ring-primary/25' : 'border-border/50'}`}
                onClick={() => batchMode ? toggleSelect(img.key) : showDetail(img)}
              >
                <div className="relative aspect-square flex items-center justify-center overflow-hidden bg-muted/30">
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
                    <button
                      type="button"
                      className={`absolute right-2 top-2 z-10 inline-flex size-7 items-center justify-center rounded-full border shadow-sm backdrop-blur-sm transition-colors duration-150 ${selectedKeys.has(img.key) ? 'border-primary bg-primary text-primary-foreground' : 'border-white/70 bg-black/35 text-transparent hover:border-white hover:bg-black/45'}`}
                      onClick={(e) => {
                        e.stopPropagation()
                        toggleSelect(img.key)
                      }}
                      aria-label={selectedKeys.has(img.key) ? t('images.exitBatch', { defaultValue: '取消选中' }) : t('images.selectAll', { defaultValue: '选中' })}
                    >
                      <Check className="size-4" />
                    </button>
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
        <DialogContent className="max-h-[90vh] flex flex-col sm:max-w-2xl p-0 gap-0 overflow-hidden">
          <DialogHeader className="px-6 py-4 border-b border-border/50 bg-muted/10 shrink-0">
            <DialogTitle>{t('images.detailTitle', { defaultValue: '图片详情' })}</DialogTitle>
          </DialogHeader>

          {detailLoading && <LoadingState compact className="py-6" />}

          {detail && (
            <div className="flex-1 overflow-y-auto p-6 space-y-6 [&::-webkit-scrollbar]:w-2 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-border/50 hover:[&::-webkit-scrollbar-thumb]:bg-border">
              {/* Preview */}
              <div className="flex justify-center rounded-xl bg-muted/20 border border-border/40 p-4">
                <img
                  src={toRelative(detail.links?.url ?? detail.url ?? '')}
                  alt={detail.key}
                  className="max-h-[35vh] rounded-lg object-contain shadow-sm"
                  onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }}
                />
              </div>

              <div className="grid sm:grid-cols-2 gap-6">
                {/* Metadata */}
                <div className="space-y-4">
                  <h4 className="text-sm font-semibold text-foreground border-b border-border/40 pb-2">{t('images.metadata', { defaultValue: '元数据' })}</h4>
                  <div className="grid grid-cols-2 gap-y-3 gap-x-4 text-sm">
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
                        className={['rounded-lg px-2 py-0.5 text-[10px] font-medium tracking-wide uppercase transition-colors cursor-pointer', detail.permission === 1 ? 'bg-primary/10 text-primary hover:bg-primary/20' : 'bg-warning/10 text-warning hover:bg-warning/20'].join(' ')}
                      >
                        {detail.permission === 1 ? (t('images.public', { defaultValue: '公开' })) : (t('images.private', { defaultValue: '私有' }))}
                      </button>
                    </div>
                    <div className="flex flex-col gap-1 items-start">
                      <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('images.album', { defaultValue: '相册' })}</span>
                      <Select 
                        value={detail.album_id?.toString() ?? 'none'}
                        onValueChange={(val) => val !== null && changeAlbum(val as string)}
                        items={{
                          'none': t('albums.noAlbum', { defaultValue: '不指定' }),
                          ...Object.fromEntries(albums.map(a => [a.id.toString(), a.name]))
                        }}
                      >
                        <SelectTrigger className="w-full h-6 px-2 py-0 bg-primary/5 hover:bg-primary/10 border-none shadow-none text-xs font-medium text-primary">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="none">{t('albums.noAlbum', { defaultValue: '不指定' })}</SelectItem>
                          {albums.map(a => (
                            <SelectItem key={a.id} value={a.id.toString()}>{a.name}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
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
                </div>

                {/* Links */}
                <div className="space-y-4">
                  <h4 className="text-sm font-semibold text-foreground border-b border-border/40 pb-2">{t('images.links', { defaultValue: '链接' })}</h4>
                  {detail.links && (
                    <div className="space-y-2">
                      {Object.entries(detail.links).map(([fmt, val]) => (
                        <div key={fmt} className="group flex items-center gap-3 rounded-lg bg-muted/30 px-3 py-2 transition-colors hover:bg-muted/50 border border-border/40 hover:border-primary/30">
                          <span className="shrink-0 w-14 text-[10px] font-bold tracking-wider text-muted-foreground uppercase">{fmt}</span>
                          <code className="min-w-0 flex-1 truncate text-xs font-medium text-foreground bg-background/50 px-2 py-1 rounded-md border border-border/30">{val}</code>
                          <button type="button" onClick={() => onCopy(val)} className="shrink-0 flex h-6 w-6 items-center justify-center rounded-md border border-border/50 bg-background text-muted-foreground shadow-sm transition-colors duration-150 hover:border-primary hover:bg-primary hover:text-primary-foreground cursor-pointer" title={t('upload.copy')}>
                            <Copy className="size-3" />
                          </button>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>

              <div className="flex justify-end pt-4 mt-6">
                <button type="button" onClick={() => setDeleteTarget(detail.key)} className="flex items-center gap-1.5 rounded-lg px-4 py-2 text-sm font-medium text-destructive bg-destructive/10 transition-colors hover:bg-destructive/20 cursor-pointer">
                  <Trash2 className="size-4" />
                  {t('images.delete')}
                </button>
              </div>
            </div>
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
        loading={batchProcessing}
      />
    </section>
  )
}
