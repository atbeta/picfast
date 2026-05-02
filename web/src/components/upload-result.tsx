import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { formatFileSize } from '../lib/upload'
import { Copy, Check, ExternalLink } from 'lucide-react'

interface UploadResultLike {
  permission?: number
  moderation_status?: string
  origin_name: string
  size_bytes: number
  extension: string
  width: number
  height: number
  links: { url: string; html: string; bbcode: string; markdown: string; thumbnail_url: string }
}

interface UploadResultCardProps {
  result: UploadResultLike
}

interface CopyItem {
  label: string
  value: string
}

function CopyButton({ text }: { text: string }) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)

  const copy = async () => {
    await navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <button
          type="button"
          onClick={copy}
          title={copied ? t('upload.copied') : t('upload.copy')}
          className="shrink-0 flex h-8 w-8 items-center justify-center rounded-lg border border-border/50 bg-background/50 text-muted-foreground shadow-sm transition-colors duration-150 hover:border-primary hover:bg-primary hover:text-primary-foreground cursor-pointer"
        >
      {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
    </button>
  )
}

function CopyRow({ item }: { item: CopyItem }) {
  return (
    <div className="group flex items-center gap-3 rounded-lg border border-border/40 bg-muted/30 px-4 py-2.5 transition-colors duration-150 hover:border-primary/30 hover:bg-muted/50">
      <span className="shrink-0 w-20 text-xs font-semibold tracking-wider text-muted-foreground uppercase">{item.label}</span>
      <code className="min-w-0 flex-1 truncate text-sm font-medium text-foreground bg-background/50 px-2.5 py-1 rounded-lg border border-border/40 shadow-inner">{item.value}</code>
      <CopyButton text={item.value} />
    </div>
  )
}

export function UploadResultCard({ result }: UploadResultCardProps) {
  const { t } = useTranslation()
  const items: CopyItem[] = [
    { label: 'URL', value: result.links.url },
    { label: 'Markdown', value: result.links.markdown },
    { label: 'HTML', value: result.links.html },
    { label: 'BBCode', value: result.links.bbcode },
  ]

  return (
    <div className="group overflow-hidden rounded-2xl border border-border/50 bg-card/60 shadow-sm transition-colors duration-150 hover:border-border/80">
      <div className="flex flex-col sm:flex-row sm:items-start gap-6 p-6">
        {/* Image Preview */}
        <div className="relative shrink-0">
          <div className="absolute inset-0 rounded-xl bg-gradient-to-tr from-primary/20 to-transparent opacity-0 transition-opacity duration-150 group-hover:opacity-100 blur-md" />
          <div className="relative h-32 w-32 sm:h-40 sm:w-40 overflow-hidden rounded-xl border border-border/60 bg-muted/30 shadow-sm">
            {result.links.thumbnail_url ? (
              <img
                src={result.links.thumbnail_url}
                alt={result.origin_name}
                className="h-full w-full object-contain p-2"
                onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }}
              />
            ) : (
              <div className="flex h-full items-center justify-center text-muted-foreground">
                <span className="text-xs font-medium uppercase tracking-widest">{result.extension}</span>
              </div>
            )}
          </div>
        </div>

        {/* Details & Links */}
        <div className="min-w-0 flex-1 flex flex-col justify-center">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h3 className="truncate text-lg font-bold tracking-tight text-foreground" title={result.origin_name}>
                {result.origin_name}
              </h3>
              <div className="mt-1 flex items-center gap-2 text-xs font-medium text-muted-foreground">
                <span className="bg-muted px-2 py-0.5 rounded-full border border-border/50">{result.width}×{result.height}</span>
                <span className="bg-muted px-2 py-0.5 rounded-full border border-border/50">{formatFileSize(result.size_bytes)}</span>
                <span className="bg-primary/10 text-primary px-2 py-0.5 rounded-full border border-primary/20 uppercase tracking-wider">{result.extension}</span>
                {result.permission !== undefined && (
                  <span className={['px-2 py-0.5 rounded-full border', result.permission === 1 ? 'bg-primary/10 text-primary border-primary/20' : 'bg-warning/10 text-warning border-warning/20'].join(' ')}>
                    {result.permission === 1 ? t('images.public', { defaultValue: '公开' }) : t('images.private', { defaultValue: '私有' })}
                  </span>
                )}
                {result.moderation_status === 'pending' && (
                  <span className="bg-warning/10 text-warning px-2 py-0.5 rounded-full border border-warning/20">
                    {t('images.moderationPending', { defaultValue: '待审核' })}
                  </span>
                )}
                {result.moderation_status === 'rejected' && (
                  <span className="bg-destructive/10 text-destructive px-2 py-0.5 rounded-full border border-destructive/20">
                    {t('images.moderationRejected', { defaultValue: '审核拒绝' })}
                  </span>
                )}
              </div>
            </div>
            <a 
              href={result.links.url} 
              target="_blank" 
              rel="noreferrer"
              className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary/10 text-primary shadow-sm transition-colors duration-150 hover:bg-primary hover:text-primary-foreground cursor-pointer"
              title="Open Original"
            >
              <ExternalLink className="h-4 w-4" />
            </a>
          </div>

          <div className="space-y-2">
            {items.map((item) => (
              <CopyRow key={item.label} item={item} />
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
