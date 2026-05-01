import { useTranslation } from 'react-i18next'
import { useState } from 'react'
import { PlugIcon, MonitorIcon, Copy } from 'lucide-react'
import { toast } from 'sonner'
import { useQuery } from '@tanstack/react-query'
import { getSiteConfig } from '../../lib/site-config'

function trimTrailingSlash(value: string): string {
  return value.replace(/\/+$/, '')
}

export function IntegrationsPage() {
  const { t } = useTranslation()
  const [feedback, setFeedback] = useState<'mcp' | 'sharex' | 'download' | null>(null)
  const { data: siteConfig } = useQuery({
    queryKey: ['site-config'],
    queryFn: getSiteConfig,
  })

  const apiURL = siteConfig?.base_url ? trimTrailingSlash(siteConfig.base_url) : window.location.origin
  const mcpEndpoint = `${apiURL}/mcp`
  const sharexConfigExample = `{
  "Version": "15.0.0",
  "Name": "PicFast",
  "DestinationType": "ImageUploader",
  "RequestMethod": "POST",
  "RequestURL": "${apiURL}/api/v1/sharex/upload",
  "Headers": {
    "Authorization": "Bearer <YOUR_API_TOKEN>"
  },
  "Body": "MultipartFormData",
  "FileFormName": "file",
  "URL": "{json:url}",
  "ThumbnailURL": "{json:thumbnail_url}"
}`

  const downloadShareXConfig = () => {
    const blob = new Blob([sharexConfigExample], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'picfast-sharex-config.sxcu'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    setFeedback('download')
  }

  const mcpConfigExample = `{
  "mcpServers": {
    "picfast": {
      "url": "${mcpEndpoint}",
      "headers": {
        "Authorization": "Bearer <YOUR_API_TOKEN>"
      }
    }
  }
}`

  const onCopy = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      toast.success(t('upload.copySuccess'))
      setFeedback(text === mcpConfigExample ? 'mcp' : 'sharex')
    } catch {
      toast.error(t('upload.copyError'))
    }
  }

  return (
    <section className="mx-auto max-w-4xl space-y-6 animate-in slide-in-from-bottom-4 fade-in duration-700">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight">{t('integrations.title', { defaultValue: '集成与工具' })}</h1>
      </div>

      <div className="grid gap-6 sm:grid-cols-2">
        {/* MCP Server card */}
        <div className="flex h-full flex-col rounded-xl border border-border bg-card p-6 shadow-sm">
          <div className="mb-4 flex items-center gap-3">
            <div className="rounded-lg bg-info/10 p-2">
              <PlugIcon className="size-5 text-info" />
            </div>
            <div className="space-y-1">
              <span className="inline-flex rounded-full bg-info/10 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.18em] text-info">
                MCP
              </span>
              <h3 className="text-base font-semibold">{t('integrations.mcpTitle', { defaultValue: 'MCP 服务器' })}</h3>
            </div>
          </div>
          <p className="mb-4 text-sm text-muted-foreground">{t('integrations.mcpDesc', { defaultValue: '通过 MCP 协议连接 AI 助手，直接在对话中上传和管理图片。' })}</p>

          <div className="flex-1 space-y-4">
            <div className="space-y-2">
              <label className="text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('integrations.mcpEndpoint')}</label>
              <div className="flex items-center gap-2 rounded-lg border border-border/50 bg-muted/50 px-3 py-2">
                <code className="min-w-0 flex-1 truncate text-xs text-foreground">{mcpEndpoint}</code>
                <button
                  type="button"
                  onClick={() => onCopy(mcpEndpoint)}
                  className="shrink-0 flex items-center justify-center rounded-lg bg-background/80 border border-border/50 w-8 h-8 hover:bg-background hover:text-foreground text-muted-foreground transition-all cursor-pointer"
                  title={t('upload.copy')}
                >
                  <Copy className="size-4" />
                </button>
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('integrations.mcpConfigExample', { defaultValue: '配置示例' })}</label>
              <pre className="overflow-x-auto rounded-lg bg-muted p-4 text-xs leading-relaxed text-muted-foreground relative group">
                <code>{mcpConfigExample}</code>
                <button type="button" onClick={() => onCopy(mcpConfigExample)} className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 flex items-center justify-center rounded-lg bg-background/80 backdrop-blur-sm border border-border/50 w-8 h-8 hover:bg-background hover:text-foreground text-muted-foreground transition-all cursor-pointer" title={t('upload.copy')}>
                  <Copy className="size-4" />
                </button>
              </pre>
              <p className="text-xs text-muted-foreground mt-2">
                {t('integrations.mcpConfigHint', { defaultValue: '这是通用 MCP JSON 示例。把 Bearer 后面的令牌替换成你刚创建的 API Token，客户端字段名如果有差异，只保留 URL 和 Authorization 头即可。' })}
              </p>
              {feedback === 'mcp' && (
                <p className="rounded-lg border border-info/20 bg-info/5 px-3 py-2 text-xs text-info">
                  {t('integrations.mcpCopied', { defaultValue: 'MCP 配置已复制，可以直接粘贴到客户端。' })}
                </p>
              )}
            </div>
          </div>

          <div className="mt-5 rounded-lg border border-info/20 bg-info/5 px-4 py-3 text-xs text-muted-foreground">
            <p className="font-medium text-foreground">{t('integrations.mcpFooterTitle', { defaultValue: '推荐流程' })}</p>
            <p className="mt-1">
              {t('integrations.mcpFooterHint', { defaultValue: '先创建 API 令牌，再复制 MCP 配置到 Cursor、Claude Desktop 或其他支持远程 MCP 的客户端。' })}
            </p>
          </div>
        </div>

        {/* ShareX card */}
        <div className="flex h-full flex-col rounded-xl border border-border bg-card p-6 shadow-sm">
          <div className="mb-4 flex items-center gap-3">
            <div className="rounded-lg bg-success/10 p-2">
              <MonitorIcon className="size-5 text-success" />
            </div>
            <div className="space-y-1">
              <span className="inline-flex rounded-full bg-success/10 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.18em] text-success">
                ShareX
              </span>
              <h3 className="text-base font-semibold">{t('integrations.sharexTitle', { defaultValue: 'ShareX 集成' })}</h3>
            </div>
          </div>
          <p className="mb-4 text-sm text-muted-foreground">{t('integrations.sharexDesc', { defaultValue: '自动上传截图到 PicFast。下载配置文件并导入 ShareX 即可使用。' })}</p>

          <div className="flex-1 space-y-4">
            <div className="space-y-2">
              <label className="text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('integrations.sharexConfigExample', { defaultValue: '配置示例' })}</label>
              <pre className="overflow-x-auto rounded-lg bg-muted p-4 text-xs leading-relaxed text-muted-foreground relative group">
                <code>{sharexConfigExample}</code>
                <button type="button" onClick={() => onCopy(sharexConfigExample)} className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 flex items-center justify-center rounded-lg bg-background/80 backdrop-blur-sm border border-border/50 w-8 h-8 hover:bg-background hover:text-foreground text-muted-foreground transition-all cursor-pointer" title={t('upload.copy')}>
                  <Copy className="size-4" />
                </button>
              </pre>
              <p className="text-xs text-muted-foreground mt-2">
                请将 <code>&lt;YOUR_API_TOKEN&gt;</code> 替换为您创建的 API 令牌，然后您可以直接复制或下载此配置。
              </p>
              {feedback === 'sharex' && (
                <p className="rounded-lg border border-success/20 bg-success/5 px-3 py-2 text-xs text-success">
                  {t('integrations.sharexCopied', { defaultValue: 'ShareX 配置已复制，你也可以直接下载 `.sxcu` 文件导入。' })}
                </p>
              )}
            </div>
            <div className="pt-2">
              <div className="flex flex-wrap items-center gap-3">
                <button
                  type="button"
                  onClick={downloadShareXConfig}
                  className="inline-flex items-center justify-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:opacity-90 transition-opacity cursor-pointer"
                >
                  {t('integrations.downloadConfig', { defaultValue: '下载配置文件' })}
                </button>
                <span className="rounded-full bg-muted px-3 py-1 text-[11px] font-medium text-muted-foreground">
                  {t('integrations.sharexFooterTag', { defaultValue: '导入后即可开始截图上传' })}
                </span>
              </div>
              <p className="mt-3 text-xs text-muted-foreground">
                {t('integrations.sharexFooterHint', { defaultValue: '下载后在 ShareX 中选择：目标 -> 自定义上传器设置 -> 导入 -> 从文件。' })}
              </p>
              {feedback === 'download' && (
                <p className="mt-3 rounded-lg border border-success/20 bg-success/5 px-3 py-2 text-xs text-success">
                  {t('integrations.sharexDownloaded', { defaultValue: '配置文件已开始下载，导入后就可以直接截图上传。' })}
                </p>
              )}
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
