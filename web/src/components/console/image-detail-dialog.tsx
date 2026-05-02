import { Copy, Trash2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import type { Album, ImageItem } from '@/lib/console-api'
import { storageStrategyLabel } from '@/lib/storage-strategy'
import { formatFileSize } from '@/lib/upload'
import { LoadingState } from '@/components/page-states'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

function toRelative(url: string): string {
  try {
    return new URL(url).pathname
  } catch {
    return url
  }
}

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

  return (
    <Dialog open={!!image} onOpenChange={(open) => { if (!open) onClose() }}>
      <DialogContent className="max-h-[90vh] flex flex-col sm:max-w-3xl p-0 gap-0 overflow-hidden">
        <DialogHeader className="px-6 py-4 border-b border-border/50 bg-muted/10 shrink-0">
          <DialogTitle>{t('images.detailTitle', { defaultValue: '图片详情' })}</DialogTitle>
        </DialogHeader>

        {loading && <LoadingState compact className="py-6" />}

        {image && (
          <div className="flex-1 overflow-y-auto p-6 space-y-6 [&::-webkit-scrollbar]:w-2 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-border/50 hover:[&::-webkit-scrollbar-thumb]:bg-border">
            <div className="flex justify-center rounded-xl bg-muted/20 border border-border/40 p-4 relative group">
              <button
                type="button"
                onClick={() => onDelete(image.key)}
                className="absolute top-2 right-2 p-2 rounded-md text-muted-foreground/60 hover:text-destructive hover:bg-destructive/10 transition-colors opacity-0 group-hover:opacity-100 cursor-pointer"
                title={t('images.delete', { defaultValue: '删除' })}
              >
                <Trash2 className="size-4" />
              </button>
              <img
                src={toRelative(image.links?.url ?? image.url ?? '')}
                alt={image.key}
                className="max-h-[35vh] rounded-lg object-contain shadow-sm"
                onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }}
              />
            </div>

            <div className="grid sm:grid-cols-2 gap-6">
              <div className="space-y-4">
                <h4 className="text-sm font-semibold text-foreground border-b border-border/40 pb-2">{t('images.metadata', { defaultValue: '元数据' })}</h4>
                <div className="grid grid-cols-2 gap-y-3 gap-x-4 text-sm">
                  <div className="flex flex-col gap-1"><span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">Key</span> <span className="font-mono text-foreground break-all">{image.key}</span></div>
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
                    <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('images.album', { defaultValue: '相册' })}</span>
                    <Select
                      value={image.album_id?.toString() ?? 'none'}
                      onValueChange={(val) => val !== null && onChangeAlbum(val as string)}
                      items={{
                        none: t('albums.noAlbum', { defaultValue: '不指定' }),
                        ...Object.fromEntries(albums.map((a) => [a.id.toString(), a.name])),
                      }}
                    >
                      <SelectTrigger className="w-full h-6 px-2 py-0 bg-primary/5 hover:bg-primary/10 border-none shadow-none text-xs font-medium text-primary">
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
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
