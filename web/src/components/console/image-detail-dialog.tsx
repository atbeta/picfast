import { Copy, Trash2, Activity, CheckCircle2, XCircle, Clock, SkipForward } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'

import type { Album, ImageItem } from '@/lib/console-api'
import { getPipelineStatus } from '@/lib/console-api'
import { storageStrategyLabel } from '@/lib/storage-strategy'
import { formatFileSize } from '@/lib/upload'
import { Button } from '@/components/ui/button'
import { LoadingState } from '@/components/page-states'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

interface ImageDetailDialogProps {
  image: ImageItem | null
  loading: boolean
  albums: Album[]
  onClose: () => void
  onTogglePermission: () => void
  onChangeAlbum: (albumId: string) => void
  onCopy: (text: string) => void
  onDelete: (key: string) => void
}

export function ImageDetailDialog({
  image,
  loading,
  albums,
  onClose,
  onTogglePermission,
  onChangeAlbum,
  onCopy,
  onDelete,
}: ImageDetailDialogProps) {
  const { t } = useTranslation()

  const { data: pipeline } = useQuery({
    queryKey: ['pipeline', image?.key],
    queryFn: () => getPipelineStatus(image!.key),
    enabled: !!image,
  })

  return (
    <Dialog open={!!image} onOpenChange={(open) => { if (!open) onClose() }}>
      <DialogContent className="flex h-[calc(100dvh-1rem)] max-h-[calc(100dvh-1rem)] flex-col gap-0 overflow-hidden p-0 sm:h-auto sm:max-h-[90vh] sm:max-w-3xl">
        <DialogHeader className="shrink-0 border-b border-border/50 bg-muted/10 px-4 py-3 sm:px-6 sm:py-4">
          <DialogTitle>{t('images.detailTitle', { defaultValue: '图片详情' })}</DialogTitle>
        </DialogHeader>

        {loading && <LoadingState compact className="py-6" />}

        {image && (
          <div className="flex-1 space-y-4 overflow-y-auto p-4 sm:space-y-6 sm:p-6 [&::-webkit-scrollbar]:w-2 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-border/50 hover:[&::-webkit-scrollbar-thumb]:bg-border">
            <div className="group relative flex justify-center rounded-xl border border-border/40 bg-muted/20 p-3 sm:p-4">
              <Button
                variant="ghost"
                size="icon-sm"
                onClick={() => onDelete(image.key)}
                className="absolute right-2 top-2 text-muted-foreground/70 hover:text-destructive sm:opacity-0 sm:group-hover:opacity-100"
                title={t('images.delete', { defaultValue: '删除' })}
              >
                <Trash2 className="size-4" />
              </Button>
              <img
                src={image.links?.url ?? image.url ?? ''}
                alt={image.key}
                className="max-h-[32vh] rounded-lg object-contain shadow-sm sm:max-h-[35vh]"
                onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }}
              />
            </div>

            <div className="grid gap-4 sm:grid-cols-2 sm:gap-6">
              <div className="space-y-4">
                <h4 className="text-sm font-semibold text-foreground border-b border-border/40 pb-2">{t('images.metadata', { defaultValue: '元数据' })}</h4>
                <div className="grid grid-cols-1 gap-x-4 gap-y-3 text-sm sm:grid-cols-2">
                  <div className="flex flex-col gap-1 sm:col-span-2"><span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">Key</span> <span className="font-mono text-foreground break-all">{image.key}</span></div>
                  <div className="flex flex-col gap-1"><span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('images.colName')}</span> <span className="text-foreground truncate" title={image.origin_name}>{image.origin_name}</span></div>
                  <div className="flex flex-col gap-1"><span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('images.colSize')}</span> <span className="text-foreground">{formatFileSize(image.size_bytes)}</span></div>
                  <div className="flex flex-col gap-1"><span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('images.type', { defaultValue: '类型' })}</span> <span className="text-foreground">{image.mimetype}</span></div>
                  <div className="flex flex-col gap-1"><span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('images.dimensions', { defaultValue: '尺寸' })}</span> <span className="text-foreground">{image.width}x{image.height}</span></div>
                  <div className="flex flex-col gap-1 items-start">
                    <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('images.permission', { defaultValue: '权限' })}</span>
                    <button
                      type="button"
                      onClick={onTogglePermission}
                      className={['rounded-lg px-2 py-0.5 text-[10px] font-medium tracking-wide uppercase transition-colors cursor-pointer', image.permission === 1 ? 'bg-primary/10 text-primary hover:bg-primary/20' : 'bg-warning/10 text-warning hover:bg-warning/20'].join(' ')}
                    >
                      {image.permission === 1 ? t('images.public', { defaultValue: '公开' }) : t('images.private', { defaultValue: '私有' })}
                    </button>
                  </div>
                  <div className="flex flex-col gap-1 items-start">
                    <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('images.moderationStatus', { defaultValue: '审核状态' })}</span>
                    <span className={['rounded-lg px-2 py-0.5 text-[10px] font-medium tracking-wide uppercase', image.moderation_status === 'approved' ? 'bg-success/10 text-success' : image.moderation_status === 'pending' ? 'bg-warning/10 text-warning' : 'bg-destructive/10 text-destructive'].join(' ')}>
                      {image.moderation_status === 'approved'
                        ? t('images.moderationApproved', { defaultValue: '已通过' })
                        : image.moderation_status === 'pending'
                          ? t('images.moderationPending', { defaultValue: '待审核' })
                          : t('images.moderationRejected', { defaultValue: '审核拒绝' })}
                    </span>
                  </div>
                  <div className="flex flex-col gap-1 items-start">
                    <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('images.album', { defaultValue: '相册' })}</span>
                    <Select
                      value={image.album_id?.toString() ?? 'none'}
                      onValueChange={(val) => val !== null && onChangeAlbum(val as string)}
                      items={{
                        none: t('albums.noAlbum', { defaultValue: '不指定' }),
                        ...Object.fromEntries(albums.map((a) => [a.id.toString(), a.name])),
                      }}
                    >
                      <SelectTrigger className="h-8 w-full border-none bg-primary/5 px-2 py-0 text-xs font-medium text-primary shadow-none hover:bg-primary/10">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="none">{t('albums.noAlbum', { defaultValue: '不指定' })}</SelectItem>
                        {albums.map((album) => (
                          <SelectItem key={album.id} value={album.id.toString()}>{album.name}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  {image.strategy_name && (
                    <div className="col-span-2 flex flex-col gap-1">
                      <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('images.strategy', { defaultValue: '存储策略' })}</span>
                      <span className="self-start rounded-lg bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
                        {image.strategy_name} ({storageStrategyLabel(t, image.strategy_type)})
                      </span>
                    </div>
                  )}
                </div>
              </div>

              <div className="space-y-4">
                <h4 className="text-sm font-semibold text-foreground border-b border-border/40 pb-2">{t('images.links', { defaultValue: '链接' })}</h4>
                {image.links && (
                  <div className="space-y-2">
                    {Object.entries(image.links).map(([fmt, val]) => (
                      <div key={fmt} className="group flex items-start gap-2 rounded-lg border border-border/40 bg-muted/30 px-2.5 py-2 transition-colors hover:border-primary/30 hover:bg-muted/50 sm:items-center sm:gap-3 sm:px-3">
                        <span className="w-12 shrink-0 pt-1 text-[10px] font-bold uppercase tracking-wider text-muted-foreground sm:w-14 sm:pt-0">{fmt}</span>
                        <code className="min-w-0 flex-1 break-all text-[11px] font-medium text-foreground bg-background/50 px-2 py-1 rounded-md border border-border/30 sm:text-xs">{val}</code>
                        <Button variant="ghost" size="icon-xs" onClick={() => onCopy(val)} className="mt-0.5 shrink-0 border border-border/50 hover:border-primary hover:bg-primary hover:text-primary-foreground sm:mt-0" title={t('upload.copy')}>
                          <Copy className="size-3" />
                        </Button>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>

            {pipeline && (() => {
              const stages = ['upload', 'processing', 'thumbnail', 'moderation'] as const
              const hasIssueOrPending = stages.some(key => {
                const status = pipeline[key] ?? 'pending'
                return status === 'pending' || status === 'failed' || status === 'rejected'
              })

              // 如果一切正常且没有在处理中，就隐藏整个流水线面板以节省空间
              if (!hasIssueOrPending) return null

              return (
                <div className="space-y-3 rounded-xl border border-border/40 bg-muted/10 p-4">
                  <h4 className="flex items-center gap-2 text-sm font-semibold text-foreground border-b border-border/40 pb-2">
                    <Activity className="size-4 text-primary" />
                    {t('images.pipeline', { defaultValue: '处理流水线' })}
                  </h4>
                  <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                    {stages.map((key) => {
                      const label = t(`images.pipeline${key.charAt(0).toUpperCase() + key.slice(1)}`)
                      const status = pipeline[key] ?? 'pending'
                      const isGood = status === 'completed' || status === 'approved'
                      const isBad = status === 'failed' || status === 'rejected'
                      const isSkipped = status === 'skipped'
                      const statusKey = status === 'completed' ? 'statusCompleted'
                        : status === 'approved' ? 'statusApproved'
                        : status === 'rejected' ? 'statusRejected'
                        : status === 'failed' ? 'statusFailed'
                        : status === 'skipped' ? 'statusSkipped'
                        : 'statusPending'
                      return (
                        <div key={key} className="flex flex-col items-center gap-1 rounded-lg border border-border/40 bg-background/50 p-3">
                          {isGood ? <CheckCircle2 className="size-5 text-green-500" /> :
                           isBad ? <XCircle className="size-5 text-red-500" /> :
                           isSkipped ? <SkipForward className="size-5 text-muted-foreground" /> :
                           <Clock className="size-5 text-yellow-500" />}
                          <span className="text-xs font-medium text-muted-foreground">{label}</span>
                          <span className={`text-xs font-semibold ${
                            isGood ? 'text-green-600' :
                            isBad ? 'text-red-600' :
                            'text-yellow-600'
                          }`}>
                            {t(`images.${statusKey}`)}
                          </span>
                        </div>
                      )
                    })}
                  </div>
                  <p className="text-xs text-muted-foreground text-right">
                    {pipeline.updated_at && new Date(pipeline.updated_at).toLocaleString()}
                  </p>
                </div>
              )
            })()}
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
