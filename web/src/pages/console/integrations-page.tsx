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
      "command": "npx",
      "args": ["-y", "@picfast/mcp"],
      "env": {
        "PICFAST_BASE_URL": "${apiURL}",
        "PICFAST_API_TOKEN": "<YOUR_API_TOKEN>"
      }
    }
  }
}`

  const onCopy = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      toast.success(t('upload.copySuccess'))
      setFeedback(text === sharexConfigExample ? 'sharex' : 'mcp')
    } catch {
      toast.error(t('upload.copyError'))
    }
  }

  return (
    <section className="space-y-6">
      <div className="flex items-center justify-between border-b border-border/40 pb-3">
        <h1 className="text-2xl font-bold tracking-tight">{t('integrations.title', { defaultValue: '集成与工具' })}</h1>
      </div>

      <div className="space-y-6">
        {/* MCP Server Section */}
        <section className="space-y-5">
          <div className="flex items-center gap-3">
            <div className="rounded-xl bg-info/10 p-2.5">
              <PlugIcon className="size-6 text-info" />
            </div>
            <div>
              <div className="flex items-center gap-2 mb-1">
                <h2 className="text-xl font-bold">{t('integrations.mcpTitle', { defaultValue: 'MCP 服务器' })}</h2>
                <span className="inline-flex rounded-full bg-info/10 px-2 py-0.5 text-[10px] font-bold tracking-wider text-info">
                  {t('integrations.mcpBadge', { defaultValue: 'AI 集成' })}
                </span>
              </div>
              <p className="text-sm text-muted-foreground">{t('integrations.mcpDesc', { defaultValue: '通过 MCP 协议连接 AI 助手，直接在对话中上传和管理图片。本地运行，零 token 开销。' })}</p>
            </div>
          </div>

          <div className="rounded-xl border border-border bg-card p-6 shadow-sm space-y-5">
            <div className="space-y-2">
              <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">{t('integrations.mcpConfigExample', { defaultValue: '配置示例' })}</label>
              <div className="relative group">
                <pre className="overflow-x-auto rounded-lg bg-muted/50 border border-border/50 p-4 text-sm leading-relaxed text-muted-foreground">
                  <code>{mcpConfigExample}</code>
                </pre>
                <button type="button" onClick={() => onCopy(mcpConfigExample)} className="absolute top-3 right-3 flex h-8 w-8 items-center justify-center rounded-md border border-border/50 bg-background backdrop-blur-sm opacity-0 shadow-sm transition-opacity duration-150 group-hover:opacity-100 hover:border-primary hover:bg-primary hover:text-primary-foreground cursor-pointer" title={t('upload.copy')}>
                  <Copy className="size-4" />
                </button>
              </div>
              <p className="text-sm text-muted-foreground mt-3">
                {t('integrations.mcpLocalHint', { defaultValue: '将 <YOUR_API_TOKEN> 替换为你创建的 API 令牌，然后将配置粘贴到 Cursor、Claude Desktop 或其他 MCP 客户端。' })}
              </p>
              {feedback === 'mcp' && (
                <p className="rounded-lg border border-info/20 bg-info/5 px-3 py-2 text-sm text-info">
                  {t('integrations.mcpCopied', { defaultValue: 'MCP 配置已复制，可以直接粘贴到客户端。' })}
                </p>
              )}
            </div>

            <div className="rounded-lg border border-info/20 bg-info/5 px-4 py-3 text-sm text-muted-foreground">
              <p className="font-semibold text-foreground">{t('integrations.mcpFooterTitle', { defaultValue: '推荐流程' })}</p>
              <p className="mt-1">
                {t('integrations.mcpFooterHint', { defaultValue: '先创建 API 令牌，再复制 MCP 配置到 Cursor、Claude Desktop 或其他支持 MCP 的客户端。' })}
              </p>
            </div>
          </div>
        </section>

        {/* ShareX Section */}
        <section className="space-y-5">
          <div className="flex items-center gap-3">
            <div className="rounded-xl bg-success/10 p-2.5">
              <MonitorIcon className="size-6 text-success" />
            </div>
            <div>
              <div className="flex items-center gap-2 mb-1">
                <h2 className="text-xl font-bold">{t('integrations.sharexTitle', { defaultValue: 'ShareX 集成' })}</h2>
                <span className="inline-flex rounded-full bg-success/10 px-2 py-0.5 text-[10px] font-bold tracking-wider text-success">
                  {t('integrations.sharexBadge', { defaultValue: '桌面工具' })}
                </span>
              </div>
              <p className="text-sm text-muted-foreground">{t('integrations.sharexDesc', { defaultValue: '自动上传截图到 PicFast。下载配置文件并导入 ShareX 即可使用。' })}</p>
            </div>
          </div>

          <div className="rounded-xl border border-border bg-card p-6 shadow-sm space-y-5">
            <div className="space-y-2">
              <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">{t('integrations.sharexConfigExample', { defaultValue: '配置示例' })}</label>
              <div className="relative group">
                <pre className="overflow-x-auto rounded-lg bg-muted/50 border border-border/50 p-4 text-sm leading-relaxed text-muted-foreground">
                  <code>{sharexConfigExample}</code>
                </pre>
                <button type="button" onClick={() => onCopy(sharexConfigExample)} className="absolute top-3 right-3 flex h-8 w-8 items-center justify-center rounded-md border border-border/50 bg-background backdrop-blur-sm opacity-0 shadow-sm transition-opacity duration-150 group-hover:opacity-100 hover:border-primary hover:bg-primary hover:text-primary-foreground cursor-pointer" title={t('upload.copy')}>
                  <Copy className="size-4" />
                </button>
              </div>
              <p className="text-sm text-muted-foreground mt-3">
                {t('integrations.sharexTokenHintPrefix', { defaultValue: '请将' })}{' '}
                <code className="bg-muted px-1.5 py-0.5 rounded text-xs">&lt;YOUR_API_TOKEN&gt;</code>{' '}
                {t('integrations.sharexTokenHintSuffix', { defaultValue: '替换为已创建的 API 令牌，随后可直接复制或下载该配置。' })}
              </p>
              {feedback === 'sharex' && (
                <p className="rounded-lg border border-success/20 bg-success/5 px-3 py-2 text-sm text-success mt-3">
                  {t('integrations.sharexCopied', { defaultValue: 'ShareX 配置已复制，你也可以直接下载 `.sxcu` 文件导入。' })}
                </p>
              )}
            </div>

            <div className="pt-2 border-t border-border/40">
              <div className="flex flex-wrap items-center gap-4 mt-4">
                <button
                  type="button"
                  onClick={downloadShareXConfig}
                  className="inline-flex items-center justify-center gap-2 rounded-lg bg-primary px-5 py-2.5 text-sm font-medium text-primary-foreground hover:opacity-90 transition-opacity cursor-pointer shadow-sm"
                >
                  {t('integrations.downloadConfig', { defaultValue: '下载配置文件' })}
                </button>
              </div>
              <p className="mt-4 text-sm text-muted-foreground">
                {t('integrations.sharexFooterHint', { defaultValue: '下载后在 ShareX 中选择：目标 -> 自定义上传器设置 -> 导入 -> 从文件。' })}
              </p>
              {feedback === 'download' && (
                <p className="mt-3 rounded-lg border border-success/20 bg-success/5 px-3 py-2 text-sm text-success">
                  {t('integrations.sharexDownloaded', { defaultValue: '配置文件已开始下载，导入后就可以直接截图上传。' })}
                </p>
              )}
            </div>
          </div>
        </section>
      </div>
    </section>
  )
}