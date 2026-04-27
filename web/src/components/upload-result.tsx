import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { formatFileSize } from '../lib/upload'

interface UploadResultLike {
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
      className="shrink-0 rounded px-2 py-1 text-xs text-zinc-500 hover:bg-zinc-100 hover:text-zinc-700 dark:text-zinc-400 dark:hover:bg-zinc-800 dark:hover:text-zinc-200"
    >
      {copied ? t('upload.copied') : t('upload.copy')}
    </button>
  )
}

function CopyRow({ item }: { item: CopyItem }) {
  return (
    <div className="flex items-center gap-2 rounded-md bg-zinc-50 px-3 py-2 dark:bg-zinc-800">
      <span className="shrink-0 text-xs font-medium text-zinc-500 dark:text-zinc-400">{item.label}</span>
      <code className="min-w-0 flex-1 truncate text-xs text-zinc-700 dark:text-zinc-300">{item.value}</code>
      <CopyButton text={item.value} />
    </div>
  )
}

export function UploadResultCard({ result }: UploadResultCardProps) {
  const items: CopyItem[] = [
    { label: 'URL', value: result.links.url },
    { label: 'Markdown', value: result.links.markdown },
    { label: 'HTML', value: result.links.html },
    { label: 'BBCode', value: result.links.bbcode },
  ]

  return (
    <div className="rounded-lg border border-zinc-200 bg-white p-4 dark:border-zinc-800 dark:bg-zinc-900">
      <div className="flex items-start gap-4">
        {result.links.thumbnail_url && (
          <img
            src={result.links.thumbnail_url}
            alt={result.origin_name}
            className="h-20 w-20 shrink-0 rounded-md border border-zinc-200 object-cover dark:border-zinc-700"
          />
        )}
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-medium text-zinc-700 dark:text-zinc-300">{result.origin_name}</p>
          <p className="mt-0.5 text-xs text-zinc-500">
            {result.width}×{result.height} · {formatFileSize(result.size_bytes)} · {result.extension.toUpperCase()}
          </p>
        </div>
      </div>
      <div className="mt-4 space-y-2">
        {items.map((item) => (
          <CopyRow key={item.label} item={item} />
        ))}
      </div>
    </div>
  )
}
