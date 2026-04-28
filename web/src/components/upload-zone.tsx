import { useCallback, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Upload } from 'lucide-react'

interface UploadZoneProps {
  onFiles: (files: File[]) => void
  disabled?: boolean
}

const IMAGE_EXTENSIONS = new Set(['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp', 'svg', 'ico', 'tif', 'tiff'])

function isImageFile(file: File): boolean {
  const ext = file.name.split('.').pop()?.toLowerCase() ?? ''
  return IMAGE_EXTENSIONS.has(ext) || file.type.startsWith('image/')
}

export function UploadZone({ onFiles, disabled }: UploadZoneProps) {
  const { t } = useTranslation()
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
      onDrop={onDrop}
      onDragOver={onDragOver}
      onDragLeave={onDragLeave}
      className={[
        'relative flex cursor-pointer flex-col items-center justify-center rounded-xl p-12 transition-all duration-300 overflow-hidden group',
        dragging
          ? 'bg-primary/5 dark:bg-primary/10'
          : 'bg-muted/30 hover:bg-muted/50 dark:bg-muted/10 dark:hover:bg-muted/20',
        disabled && 'pointer-events-none opacity-50',
      ].join(' ')}
    >
      {/* Animated Border */}
      <div className={[
        "absolute inset-0 rounded-xl border-2 border-dashed transition-colors duration-300",
        dragging ? "border-primary scale-[0.99]" : "border-border group-hover:border-primary/50"
      ].join(' ')} />

      {/* Inner Glow Effect */}
      <div className={[
        "absolute inset-0 bg-gradient-to-tr from-primary/10 via-transparent to-info/10 opacity-0 transition-opacity duration-500",
        dragging ? "opacity-100" : "group-hover:opacity-100"
      ].join(' ')} />

      <div className="relative z-10 flex flex-col items-center">
        <div className={[
          "mb-6 flex h-20 w-20 items-center justify-center rounded-full bg-background shadow-sm transition-transform duration-300 border border-border/50",
          dragging ? "scale-110 shadow-primary/20 shadow-lg" : "group-hover:scale-105 group-hover:shadow-md"
        ].join(' ')}>
          <Upload className={[
            "h-8 w-8 transition-colors duration-300",
            dragging ? "text-primary" : "text-muted-foreground group-hover:text-primary/80"
          ].join(' ')} />
        </div>
        <h3 className="text-xl font-semibold tracking-tight text-foreground">
          {t('upload.dropHint')}
        </h3>
        <p className="mt-2 text-sm text-muted-foreground/80 max-w-xs text-center">
          {t('upload.dropFormats')}
        </p>
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
