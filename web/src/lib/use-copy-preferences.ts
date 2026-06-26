import { type CopyFormat, type CopyPreferences } from '@/lib/copy-template'
import { usePersonalization } from '@/lib/use-personalization'

export function useCopyPreferences(): CopyPreferences {
  const { output } = usePersonalization()
  return output
}

export function copyFormatLabelKey(format: CopyFormat): string {
  switch (format) {
    case 'url':
      return 'copy.formatUrl'
    case 'html':
      return 'copy.formatHtml'
    case 'bbcode':
      return 'copy.formatBbcode'
    case 'thumbnail':
      return 'copy.formatThumbnail'
    case 'custom':
      return 'copy.formatCustom'
    case 'markdown':
    default:
      return 'copy.formatMarkdown'
  }
}
