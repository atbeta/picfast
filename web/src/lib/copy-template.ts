export type CopyFormat = 'url' | 'markdown' | 'html' | 'bbcode' | 'thumbnail' | 'custom'

export interface CopyLinkFields {
  url: string
  markdown: string
  html: string
  bbcode: string
  thumbnail_url?: string
}

export interface CopyTemplateContext {
  name: string
  extension: string
  width: number
  height: number
}

export interface CopyPreferences {
  format: CopyFormat
  template: string
}

const FORMAT_VALUES = new Set<CopyFormat>(['url', 'markdown', 'html', 'bbcode', 'thumbnail', 'custom'])

export function normalizeCopyFormat(value?: string | null): CopyFormat {
  const raw = (value || 'markdown').trim().toLowerCase()
  if (FORMAT_VALUES.has(raw as CopyFormat)) {
    return raw as CopyFormat
  }
  return 'markdown'
}

export function renderCopyTemplate(template: string, links: CopyLinkFields, ctx: CopyTemplateContext): string {
  const vars: Record<string, string> = {
    url: links.url,
    name: ctx.name,
    thumb: links.thumbnail_url || links.url,
    markdown: links.markdown,
    html: links.html,
    bbcode: links.bbcode,
    width: String(ctx.width || 0),
    height: String(ctx.height || 0),
    extension: ctx.extension,
  }
  return template.replace(/\{([a-z_]+)\}/gi, (_, key: string) => vars[key.toLowerCase()] ?? '')
}

export function resolveCopyText(
  links: CopyLinkFields,
  ctx: CopyTemplateContext,
  prefs: CopyPreferences,
): string {
  const format = normalizeCopyFormat(prefs.format)
  if (format === 'custom') {
    const template = prefs.template.trim()
    if (template) {
      return renderCopyTemplate(template, links, ctx)
    }
    return links.markdown
  }
  switch (format) {
    case 'url':
      return links.url
    case 'html':
      return links.html
    case 'bbcode':
      return links.bbcode
    case 'thumbnail':
      return links.thumbnail_url || links.url
    case 'markdown':
    default:
      return links.markdown
  }
}

export function batchCopyMarkdown(items: Array<{ links: CopyLinkFields; name: string }>): string {
  return items
    .map((item) => `![${item.name}](${item.links.url})`)
    .join('\n')
}

export function batchCopyHTML(items: Array<{ links: CopyLinkFields; name: string }>): string {
  return items
    .map((item) => `<img src="${item.links.url}" alt="${item.name.replace(/"/g, '&quot;')}" />`)
    .join('\n')
}
