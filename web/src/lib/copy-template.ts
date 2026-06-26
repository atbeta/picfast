export interface CopyLinkFields {
  url: string
  markdown: string
  html: string
  bbcode: string
  thumbnail_url?: string
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
