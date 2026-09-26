import { useCallback, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ImagePlus, Upload } from 'lucide-react'

import { useIsCoarsePointer } from '@/lib/use-media-query'

interface UploadZoneProps {
  onFiles: (files: File[]) => void
  disabled?: boolean
  className?: string
}

const IMAGE_EXTENSIONS = new Set(['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp', 'svg', 'ico', 'tif', 'tiff'])

function isImageFile(file: File): boolean {
  const ext = file.name.split('.').pop()?.toLowerCase() ?? ''
  return IMAGE_EXTENSIONS.has(ext) || file.type.startsWith('image/')
}

export function UploadZone({ onFiles, disabled, className = '' }: UploadZoneProps) {
  const { t } = useTranslation()
  const isTouch = useIsCoarsePointer()
  const inputRef = useRef<HTMLInputElement>(null)
  const [dragging, setDragging] = useState(false)

  const handleFiles = useCallback(
    (fileList: FileList | null) => {
      if (!fileList?.length) return
      const images = Array.from(fileList).filter(isImageFile)
      if (images.length) onFiles(images)
    },
    [onFiles],
  )

  const onDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault()
      setDragging(false)
      if (!disabled) handleFiles(e.dataTransfer.files)
    },
    [disabled, handleFiles],
  )

  const onDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setDragging(true)
  }, [])

  const onDragLeave = useCallback(() => setDragging(false), [])

  const onClick = () => {
    if (!disabled) inputRef.current?.click()
  }

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={onClick}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault()
          onClick()
        }
      }}
      onDrop={onDrop}
      onDragOver={onDragOver}
      onDragLeave={onDragLeave}
      className={[
        'relative flex h-full min-h-[240px] cursor-pointer flex-col items-center justify-center overflow-hidden rounded-xl border-2 border-dashed border-border p-6 transition-colors duration-150 group hover:border-primary/50 sm:min-h-0 sm:p-12',
        dragging
          ? 'bg-primary/5 dark:bg-primary/10 border-primary'
          : 'bg-muted/30 hover:bg-muted/50 dark:bg-muted/10 dark:hover:bg-muted/20',
        disabled && 'pointer-events-none opacity-50',
        className
      ].join(' ')}
    >
      {/* Inner Glow Effect */}
      <div className={[
        "absolute inset-0 bg-gradient-to-tr from-primary/10 via-transparent to-info/10 opacity-0 transition-opacity duration-150 rounded-xl",
        dragging ? "opacity-100" : "group-hover:opacity-100"
      ].join(' ')} />

      <div className={`relative z-10 flex flex-col items-center ${isTouch ? 'text-center' : ''}`}>
        <div className={[
          "mb-5 flex h-16 w-16 items-center justify-center rounded-full border border-border/50 bg-background shadow-sm transition-shadow duration-150 sm:mb-6 sm:h-20 sm:w-20",
          dragging ? "shadow-primary/20 shadow-md" : "group-hover:shadow-sm"
        ].join(' ')}>
          {isTouch ? (
            <ImagePlus className="h-7 w-7 text-primary/80 sm:h-8 sm:w-8" />
          ) : (
            <Upload className={[
              "h-8 w-8 transition-colors duration-300",
              dragging ? "text-primary" : "text-muted-foreground group-hover:text-primary/80"
            ].join(' ')} />
          )}
        </div>
        <h3 className="text-lg font-semibold tracking-tight text-foreground sm:text-xl">
          {isTouch ? t('upload.mobileHint') : t('upload.dropHint')}
        </h3>
        <p className="mt-2 max-w-xs text-sm text-muted-foreground/80">
          {t('upload.dropFormats')}
        </p>
        {!isTouch && (
          <p className="mt-1 text-xs text-muted-foreground/60">
            {t('upload.pasteHint')}
          </p>
        )}
        {isTouch && (
          <span className="mt-5 inline-flex h-11 items-center justify-center rounded-xl bg-primary px-6 text-sm font-semibold text-primary-foreground shadow-sm">
            {t('upload.choosePhotos')}
          </span>
        )}
      </div>

      <input
        ref={inputRef}
        type="file"
        accept="image/*"
        multiple
        className="hidden"
        onChange={(e) => {
          handleFiles(e.target.files)
          e.target.value = ''
        }}
      />
    </div>
  )
}
