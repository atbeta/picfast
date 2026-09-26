import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { formatFileSize } from '../lib/upload'
import { Copy, Check, ExternalLink, ChevronDown } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { copyToClipboard } from '@/lib/clipboard'

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
    await copyToClipboard(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={copy}
      title={copied ? t('upload.copied') : t('upload.copy')}
      className="shrink-0 border border-border/50 bg-background/50 hover:border-primary hover:bg-primary hover:text-primary-foreground"
    >
      {copied ? <Check className="size-4" /> : <Copy className="size-4" />}
    </Button>
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
  const [primaryCopied, setPrimaryCopied] = useState(false)
  const [urlCopied, setUrlCopied] = useState(false)
  const [showAll, setShowAll] = useState(false)

  // Desktop keeps the original primary action (Markdown); phones additionally
  // get a one-tap "copy link" action up front.
  const copyMarkdown = async () => {
    await copyToClipboard(result.links.markdown)
    setPrimaryCopied(true)
    setTimeout(() => setPrimaryCopied(false), 2000)
  }

  const copyUrl = async () => {
    await copyToClipboard(result.links.url)
    setUrlCopied(true)
    setTimeout(() => setUrlCopied(false), 2000)
  }

  const items: CopyItem[] = [
    { label: 'URL', value: result.links.url },
    { label: 'Markdown', value: result.links.markdown },
    { label: 'HTML', value: result.links.html },
    { label: 'BBCode', value: result.links.bbcode },
  ]
  if (result.links.thumbnail_url) {
    items.push({ label: 'Thumbnail URL', value: result.links.thumbnail_url })
  }

  return (
    <div className="group overflow-hidden rounded-2xl border border-border/50 bg-card/60 shadow-sm transition-colors duration-150 hover:border-border/80">
      <div className="flex flex-col gap-4 p-4 sm:flex-row sm:items-start sm:gap-6 sm:p-6">
        {/* Image Preview */}
        <div className="relative w-full shrink-0 sm:w-auto">
          <div className="absolute inset-0 rounded-xl bg-gradient-to-tr from-primary/20 to-transparent opacity-0 transition-opacity duration-150 group-hover:opacity-100 blur-md" />
          <div className="relative h-44 w-full overflow-hidden rounded-xl border border-border/60 bg-muted/30 shadow-sm sm:h-40 sm:w-40">
            {(result.links.thumbnail_url || result.extension === 'svg' || result.extension === 'ico') ? (
              <img
                src={result.links.thumbnail_url || result.links.url}
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
          <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h3 className="break-all text-base font-bold tracking-tight text-foreground sm:truncate sm:text-lg" title={result.origin_name}>
                {result.origin_name}
              </h3>
              <div className="mt-1 flex flex-wrap items-center gap-1.5 text-xs font-medium text-muted-foreground sm:flex-nowrap sm:gap-2">
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
            <div className="flex items-center gap-2">
              <Button
                type="button"
                size="sm"
                onClick={copyUrl}
                className="h-10 flex-1 gap-2 sm:hidden"
              >
                {urlCopied ? <Check className="size-4" /> : <Copy className="size-4" />}
                {urlCopied ? t('upload.copied') : t('upload.copyLink')}
              </Button>
              <Button
                type="button"
                size="sm"
                onClick={copyMarkdown}
                className="h-10 flex-1 gap-2 sm:h-7 sm:flex-none"
              >
                {primaryCopied ? <Check className="size-4" /> : <Copy className="size-4" />}
                {primaryCopied ? t('upload.copied') : 'Markdown'}
              </Button>
              <a
                href={result.links.url}
                target="_blank"
                rel="noreferrer"
                className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border border-border/50 bg-background/50 text-primary shadow-sm transition-colors duration-150 hover:bg-primary hover:text-primary-foreground sm:h-8 sm:w-8 sm:rounded-full sm:border-0 sm:bg-primary/10"
                title={t('upload.openOriginal')}
              >
                <ExternalLink className="h-4 w-4" />
              </a>
            </div>
          </div>

          <div className={`space-y-2 ${showAll ? 'block' : 'hidden sm:block'}`}>
            {items.map((item) => (
              <CopyRow key={item.label} item={item} />
            ))}
          </div>

          <button
            type="button"
            onClick={() => setShowAll((v) => !v)}
            className="mt-3 inline-flex h-8 items-center gap-1 self-start text-xs font-semibold text-muted-foreground transition-colors hover:text-foreground sm:hidden"
          >
            <ChevronDown className={`size-4 transition-transform ${showAll ? 'rotate-180' : ''}`} />
            {showAll ? t('upload.lessFormats') : t('upload.moreFormats')}
          </button>
        </div>
      </div>
    </div>
  )
}
