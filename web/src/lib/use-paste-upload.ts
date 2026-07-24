import { useEffect } from 'react'

export function usePasteUpload(onFiles: (files: File[]) => void, disabled?: boolean) {
  useEffect(() => {
    const handler = (e: ClipboardEvent) => {
      if (disabled) return
      if (!e.clipboardData?.items) return

      const files: File[] = []
      for (let i = 0; i < e.clipboardData.items.length; i++) {
        const item = e.clipboardData.items[i]
        if (item.kind === 'file' && item.type.startsWith('image/')) {
          const file = item.getAsFile()
          if (file) files.push(file)
        }
      }
      if (files.length > 0) {
        e.preventDefault()
        onFiles(files)
      }
    }
    document.addEventListener('paste', handler)
    return () => document.removeEventListener('paste', handler)
  }, [onFiles, disabled])
}
