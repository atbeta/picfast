import { useQuery } from '@tanstack/react-query'

import { useAuth } from '@/lib/auth-context'
import { normalizeCopyFormat, type CopyFormat, type CopyPreferences } from '@/lib/copy-template'
import { getSiteConfig } from '@/lib/site-config'

export function useCopyPreferences(): CopyPreferences {
  const { user } = useAuth()
  const { data: site } = useQuery({ queryKey: ['site-config'], queryFn: getSiteConfig })

  const userSettings = user?.settings as {
    default_copy_format?: string
    copy_template?: string
  } | undefined

  const format = normalizeCopyFormat(userSettings?.default_copy_format || site?.default_copy_format)
  const template = (userSettings?.copy_template ?? site?.copy_template ?? '').trim()

  return { format, template }
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
