import { useTranslation } from 'react-i18next'
import { useState, useMemo } from 'react'
import { PlugIcon, MonitorIcon, ImagePlus, Code, Copy, ChevronDown, KeyRound, Eye, EyeOff } from 'lucide-react'
import { toast } from 'sonner'
import { useQuery } from '@tanstack/react-query'
import { getSiteConfig } from '../../lib/site-config'
import { createApiToken } from '../../lib/console-api'
import { copyToClipboard } from '../../lib/clipboard'
import { Button } from '@/components/ui/button'
import { extractErrorMessage } from '../../lib/error-handler'

function trimTrailingSlash(value: string): string {
  return value.replace(/\/+$/, '')
}

type CardKey = 'mcp' | 'sharex' | 'flat' | 'vscode'

export function IntegrationsPage() {
  const { t } = useTranslation()
  const [expanded, setExpanded] = useState<CardKey | null>(null)
  const [copied, setCopied] = useState<string | null>(null)
  const [activeToken, setActiveToken] = useState('')
  const [showToken, setShowToken] = useState(false)
  const [inputToken, setInputToken] = useState('')
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [newTokenName, setNewTokenName] = useState('')
  const [creating, setCreating] = useState(false)
  const { data: siteConfig } = useQuery({
    queryKey: ['site-config'],
    queryFn: getSiteConfig,
  })

  const apiURL = siteConfig?.base_url ? trimTrailingSlash(siteConfig.base_url) : window.location.origin

  const tk = activeToken || '<YOUR_API_TOKEN>'

  const handleApplyToken = () => {
    const trimmed = inputToken.trim()
    if (!trimmed) return
    setActiveToken(trimmed)
    setInputToken('')
    toast.success(t('connections.tokenApplied'))
  }

  const handleCreateToken = async () => {
    if (!newTokenName.trim()) return
    setCreating(true)
    try {
      const result = await createApiToken(newTokenName.trim(), 'never', ['read', 'write'])
      if (result.token) {
        setActiveToken(result.token)
        setInputToken('')
        setShowCreateForm(false)
        setNewTokenName('')
        toast.success(t('connections.tokenCreated'))
      }
    } catch (err: unknown) {
      toast.error(extractErrorMessage(err, t('tokens.createFailed')))
    } finally {
      setCreating(false)
    }
  }

  const configs = useMemo(() => ({
    mcp: `{
  "mcpServers": {
    "picfast": {
      "command": "npx",
      "args": ["-y", "@picfast/mcp"],
      "env": {
        "PICFAST_BASE_URL": "${apiURL}",
        "PICFAST_API_TOKEN": "${tk}"
      }
    }
  }
}`,
    sharex: `{
  "Version": "15.0.0",
  "Name": "PicFast",
  "DestinationType": "ImageUploader",
  "RequestMethod": "POST",
  "RequestURL": "${apiURL}/api/v1/sharex/upload",
  "Headers": {
    "Authorization": "Bearer ${tk}"
  },
  "Body": "MultipartFormData",
  "FileFormName": "file",
  "URL": "{json:url}",
  "ThumbnailURL": "{json:thumbnail_url}"
}`,
    flatPicgo: `{
  "picBed": {
    "uploader": "picgo-plugin-web-uploader",
    "picgo-plugin-web-uploader": {
      "url": "${apiURL}/api/v1/flat/upload",
      "paramName": "file",
      "jsonPath": "url",
      "customHeader": "{\\"Authorization\\": \\"Bearer ${tk}\\"}"
    }
  }
}`,
    flatUPic: `URL: ${apiURL}/api/v1/flat/upload
Method: POST
File Field: file
Header: Authorization: Bearer ${tk}
URL Path: ["url"]
Domain: ${apiURL}`,
    flatDropshare: `Upload URL: ${apiURL}/api/v1/flat/upload
Method: POST
Content Type: multipart/form-data
Form Field: file
Header: {"Authorization": "Bearer ${tk}"}
Response Content Type: JSON
URL to file: %url%`,
    vscodePasteImage: `// settings.json
{
  "pasteImage.path": "${apiURL}/api/v1/flat/upload",
  "pasteImage.prefix": "",
  "pasteImage.forceUpload": true,
  "pasteImage.namePrefix": "",
  "pasteImage.suffix": ".png",
  "pasteImage.insertType": "markdown",
  "pasteImage.customHeader": {
    "Authorization": "Bearer ${tk}"
  },
  "pasteImage.responseType": "json",
  "pasteImage.responsePath": "$.url"
}`,
    vscodeImageSnippets: `// settings.json — Image Snippets extension
{
  "imageSnippets.uploadUrl": "${apiURL}/api/v1/flat/upload",
  "imageSnippets.uploadHeader": {
    "Authorization": "Bearer ${tk}"
  },
  "imageSnippets.uploadFieldName": "file",
  "imageSnippets.responseUrlPath": "url"
}`,
  }), [apiURL, tk])

  const onCopy = async (text: string, key: string) => {
    try {
      await copyToClipboard(text)
      toast.success(t('upload.copySuccess'))
      setCopied(key)
      setTimeout(() => setCopied(null), 2500)
    } catch {
      toast.error(t('upload.copyError'))
    }
  }

  const downloadShareXConfig = () => {
    const blob = new Blob([configs.sharex], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'picfast-sharex-config.sxcu'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    setCopied('download')
    setTimeout(() => setCopied(null), 2500)
  }

  const cards: { key: CardKey; icon: React.ReactNode; colorClass: string }[] = [
    { key: 'mcp', icon: <PlugIcon className="size-5" />, colorClass: 'bg-info/10 text-info' },
    { key: 'sharex', icon: <MonitorIcon className="size-5" />, colorClass: 'bg-success/10 text-success' },
    { key: 'flat', icon: <ImagePlus className="size-5" />, colorClass: 'bg-warning/10 text-warning' },
    { key: 'vscode', icon: <Code className="size-5" />, colorClass: 'bg-purple-500/10 text-purple-500' },
  ]

  const tokenApplied = activeToken && activeToken !== '<YOUR_API_TOKEN>'

  return (
    <section className="space-y-6">
      <div className="flex items-center justify-between border-b border-border/40 pb-3">
        <h1 className="text-2xl font-bold tracking-tight">{t('connections.title', { defaultValue: '接入' })}</h1>
      </div>

      {/* Token section */}
      <div className="rounded-xl border border-border bg-card p-5 shadow-sm space-y-4">
        <div className="flex items-center gap-2">
          <KeyRound className="size-5 text-primary" />
          <h2 className="text-base font-semibold">{t('connections.tokenTitle', { defaultValue: 'API 令牌' })}</h2>
        </div>

        {tokenApplied ? (
          <div className="space-y-3">
            <div className="flex items-center gap-2 rounded-lg bg-success/10 border border-success/20 px-3 py-2">
              <div className="size-2 rounded-full bg-success shrink-0" />
              <span className="text-sm text-foreground">{t('connections.tokenActive', { defaultValue: '令牌已填入，以下配置中的令牌可直接复制使用。' })}</span>
              <Button
                variant="ghost"
                size="icon-sm"
                onClick={() => setShowToken(!showToken)}
                className="ml-auto"
              >
                {showToken ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
              </Button>
            </div>
            <div className="flex items-center gap-2">
              <code className="flex-1 bg-muted rounded-md px-3 py-2 text-sm font-mono break-all select-all overflow-hidden text-ellipsis">
                {showToken ? activeToken : 'img_' + '•'.repeat(16) + activeToken.slice(-4)}
              </code>
              <Button
                variant="link"
                size="xs"
                onClick={() => { setActiveToken(''); setShowToken(false) }}
                className="text-muted-foreground hover:text-destructive whitespace-nowrap"
              >
                {t('connections.tokenClear', { defaultValue: '清除' })}
              </Button>
            </div>
          </div>
        ) : (
          <div className="space-y-3">
            <p className="text-sm text-muted-foreground">{t('connections.tokenHint', { defaultValue: '填入已保存的 API 令牌，或创建一个新的。令牌只会显示一次，请妥善保存。' })}</p>

            <div className="flex flex-col sm:flex-row gap-2">
              <div className="flex-1 flex items-center gap-2">
                <input
                  type="text"
                  value={inputToken}
                  onChange={(e) => setInputToken(e.target.value)}
                  onKeyDown={(e) => { if (e.key === 'Enter') handleApplyToken() }}
                  placeholder="img_..."
                  className="flex-1 rounded-md border border-border bg-background px-3 py-2 text-sm font-mono placeholder:text-muted-foreground/50 focus:outline-none focus:ring-2 focus:ring-primary/50"
                />
                <Button
                  size="sm"
                  onClick={handleApplyToken}
                  disabled={!inputToken.trim()}
                >
                  {t('connections.tokenApply', { defaultValue: '应用' })}
                </Button>
              </div>
            </div>

            <div className="relative">
              <div className="absolute inset-x-0 top-1/2 border-t border-border/50" />
              <div className="relative flex justify-center">
                <span className="bg-card px-3 text-xs text-muted-foreground">{t('connections.tokenOr', { defaultValue: '或' })}</span>
              </div>
            </div>

            {!showCreateForm ? (
              <Button
                variant="outline"
                onClick={() => setShowCreateForm(true)}
              >
                <PlusIcon className="size-4" />
                {t('connections.tokenCreateNew', { defaultValue: '创建新令牌' })}
              </Button>
            ) : (
              <div className="rounded-lg border border-border p-4 space-y-3">
                <div className="flex items-center gap-2">
                  <input
                    type="text"
                    value={newTokenName}
                    onChange={(e) => setNewTokenName(e.target.value)}
                    placeholder={t('connections.tokenNamePlaceholder', { defaultValue: '令牌名称，如"ShareX"' })}
                    className="flex-1 rounded-md border border-border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground/50 focus:outline-none focus:ring-2 focus:ring-primary/50"
                    onKeyDown={(e) => { if (e.key === 'Enter') handleCreateToken() }}
                  />
                  <Button
                    size="sm"
                    onClick={handleCreateToken}
                    disabled={creating || !newTokenName.trim()}
                  >
                    {creating ? t('tokens.creating', { defaultValue: '创建中…' }) : t('tokens.confirmCreate', { defaultValue: '创建' })}
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => { setShowCreateForm(false); setNewTokenName('') }}
                  >
                    {t('dialog.cancel', { defaultValue: '取消' })}
                  </Button>
                </div>
                <p className="text-xs text-muted-foreground">{t('connections.tokenCreateNote', { defaultValue: '将创建一个读写权限且永不过期的令牌。创建后令牌只会显示一次。' })}</p>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Card grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {cards.map((card) => (
          <button
            key={card.key}
            type="button"
            onClick={() => setExpanded(expanded === card.key ? null : card.key)}
            className={`flex items-center gap-3 rounded-xl border bg-card p-4 shadow-sm transition-colors text-left cursor-pointer ${expanded === card.key ? 'border-primary/50' : 'border-border hover:border-primary/30'}`}
          >
            <div className={`rounded-lg ${card.colorClass} p-2`}>
              {card.icon}
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <h2 className="text-base font-semibold truncate">{t(`connections.${card.key}Title`)}</h2>
              </div>
              <p className="text-sm text-muted-foreground truncate">{t(`connections.${card.key}Desc`)}</p>
            </div>
            <ChevronDown className={`size-4 text-muted-foreground shrink-0 transition-transform ${expanded === card.key ? 'rotate-180' : ''}`} />
          </button>
        ))}
      </div>

      {/* Expanded sections */}
      {expanded === 'mcp' && (
        <div className="rounded-xl border border-border bg-card p-6 shadow-sm space-y-4">
          <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">{t('connections.mcpConfigExample', { defaultValue: '配置示例' })}</label>
          <div className="relative group">
            <pre className="overflow-x-auto rounded-lg bg-muted/50 border border-border/50 p-4 text-sm leading-relaxed text-muted-foreground">
              <code>{configs.mcp}</code>
            </pre>
            <button type="button" onClick={() => onCopy(configs.mcp, 'mcp')} className="absolute top-3 right-3 flex h-8 w-8 items-center justify-center rounded-md border border-border/50 bg-background backdrop-blur-sm opacity-0 shadow-sm transition-opacity duration-150 group-hover:opacity-100 hover:border-primary hover:bg-primary hover:text-primary-foreground cursor-pointer" title={t('upload.copy')}>
              <Copy className="size-4" />
            </button>
          </div>
          {copied === 'mcp' && (
            <p className="rounded-lg border border-info/20 bg-info/5 px-3 py-2 text-sm text-info">
              {t('connections.mcpCopied', { defaultValue: 'MCP 配置已复制，可以直接粘贴到客户端。' })}
            </p>
          )}
          <div className="rounded-lg border border-info/20 bg-info/5 px-4 py-3 text-sm text-muted-foreground">
            <p>{t('connections.mcpFooterHint', { defaultValue: '填入 API 令牌后，复制 MCP 配置到 Cursor、Claude Desktop 或其他支持 MCP 的客户端。' })}</p>
          </div>
        </div>
      )}

      {expanded === 'sharex' && (
        <div className="rounded-xl border border-border bg-card p-6 shadow-sm space-y-4">
          <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">{t('connections.sharexConfigExample', { defaultValue: '配置示例' })}</label>
          <div className="relative group">
            <pre className="overflow-x-auto rounded-lg bg-muted/50 border border-border/50 p-4 text-sm leading-relaxed text-muted-foreground">
              <code>{configs.sharex}</code>
            </pre>
            <button type="button" onClick={() => onCopy(configs.sharex, 'sharex')} className="absolute top-3 right-3 flex h-8 w-8 items-center justify-center rounded-md border border-border/50 bg-background backdrop-blur-sm opacity-0 shadow-sm transition-opacity duration-150 group-hover:opacity-100 hover:border-primary hover:bg-primary hover:text-primary-foreground cursor-pointer" title={t('upload.copy')}>
              <Copy className="size-4" />
            </button>
          </div>
          {copied === 'sharex' && (
            <p className="rounded-lg border border-success/20 bg-success/5 px-3 py-2 text-sm text-success">
              {t('connections.sharexCopied', { defaultValue: 'ShareX 配置已复制，也可以直接下载 `.sxcu` 文件导入。' })}
            </p>
          )}
          <Button size="lg" onClick={downloadShareXConfig}>
            {t('connections.downloadConfig', { defaultValue: '下载配置文件' })}
          </Button>
          <p className="text-sm text-muted-foreground">
            {t('connections.sharexFooterHint', { defaultValue: '下载后在 ShareX 中选择：目标 → 自定义上传器设置 → 导入 → 从文件。' })}
          </p>
          {copied === 'download' && (
            <p className="rounded-lg border border-success/20 bg-success/5 px-3 py-2 text-sm text-success">
              {t('connections.sharexDownloaded', { defaultValue: '配置文件已开始下载，导入后可以直接截图上传。' })}
            </p>
          )}
        </div>
      )}

      {expanded === 'flat' && (
        <div className="rounded-xl border border-border bg-card p-6 shadow-sm space-y-5">
          <div className="rounded-lg border border-warning/20 bg-warning/5 px-4 py-3 text-sm text-muted-foreground">
            <p>{t('connections.flatEndpointHint', { defaultValue: '以下工具均使用相同的上传接口 /api/v1/flat/upload，响应为扁平 JSON 格式。' })}</p>
          </div>

          <div className="space-y-2">
            <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">PicGo / PicList</label>
            <ol className="text-sm text-muted-foreground list-decimal list-inside space-y-1">
              <li>{t('connections.flatPicgoStep1', { defaultValue: '安装 PicGo / PicList，在插件设置中搜索并安装「web-uploader」插件。' })}</li>
              <li>{t('connections.flatPicgoStep2', { defaultValue: '在图床设置中选择「自定义 Web 图床」，填入下方配置。' })}</li>
            </ol>
            <div className="relative group">
              <pre className="overflow-x-auto rounded-lg bg-muted/50 border border-border/50 p-4 text-sm leading-relaxed text-muted-foreground">
                <code>{configs.flatPicgo}</code>
              </pre>
              <button type="button" onClick={() => onCopy(configs.flatPicgo, 'flat-picgo')} className="absolute top-3 right-3 flex h-8 w-8 items-center justify-center rounded-md border border-border/50 bg-background backdrop-blur-sm opacity-0 shadow-sm transition-opacity duration-150 group-hover:opacity-100 hover:border-primary hover:bg-primary hover:text-primary-foreground cursor-pointer" title={t('upload.copy')}>
                <Copy className="size-4" />
              </button>
            </div>
            {copied === 'flat-picgo' && (
              <p className="rounded-lg border border-warning/20 bg-warning/5 px-3 py-2 text-sm text-warning">
                {t('connections.flatCopied', { defaultValue: '配置已复制，粘贴到插件设置中即可使用。' })}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">uPic</label>
            <p className="text-sm text-muted-foreground">{t('connections.flatUPicHint', { defaultValue: '在 uPic 中选择「自定义」图床，按以下配置填写：' })}</p>
            <div className="relative group">
              <pre className="overflow-x-auto rounded-lg bg-muted/50 border border-border/50 p-4 text-sm leading-relaxed text-muted-foreground">
                <code>{configs.flatUPic}</code>
              </pre>
              <button type="button" onClick={() => onCopy(configs.flatUPic, 'flat-upic')} className="absolute top-3 right-3 flex h-8 w-8 items-center justify-center rounded-md border border-border/50 bg-background backdrop-blur-sm opacity-0 shadow-sm transition-opacity duration-150 group-hover:opacity-100 hover:border-primary hover:bg-primary hover:text-primary-foreground cursor-pointer" title={t('upload.copy')}>
                <Copy className="size-4" />
              </button>
            </div>
            {copied === 'flat-upic' && (
              <p className="rounded-lg border border-warning/20 bg-warning/5 px-3 py-2 text-sm text-warning">
                {t('connections.flatCopied', { defaultValue: '配置已复制，粘贴到插件设置中即可使用。' })}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Dropshare</label>
            <p className="text-sm text-muted-foreground">{t('connections.flatDropshareHint', { defaultValue: '在 Dropshare 中选择 Custom API 连接，按以下配置填写：' })}</p>
            <div className="relative group">
              <pre className="overflow-x-auto rounded-lg bg-muted/50 border border-border/50 p-4 text-sm leading-relaxed text-muted-foreground">
                <code>{configs.flatDropshare}</code>
              </pre>
              <button type="button" onClick={() => onCopy(configs.flatDropshare, 'flat-dropshare')} className="absolute top-3 right-3 flex h-8 w-8 items-center justify-center rounded-md border border-border/50 bg-background backdrop-blur-sm opacity-0 shadow-sm transition-opacity duration-150 group-hover:opacity-100 hover:border-primary hover:bg-primary hover:text-primary-foreground cursor-pointer" title={t('upload.copy')}>
                <Copy className="size-4" />
              </button>
            </div>
            {copied === 'flat-dropshare' && (
              <p className="rounded-lg border border-warning/20 bg-warning/5 px-3 py-2 text-sm text-warning">
                {t('connections.flatCopied', { defaultValue: '配置已复制，粘贴到插件设置中即可使用。' })}
              </p>
            )}
          </div>

          <div className="rounded-lg border border-warning/20 bg-warning/5 px-4 py-3 text-sm text-muted-foreground">
            <p>{t('connections.flatObsidianHint', { defaultValue: 'Obsidian 用户：安装社区插件「Image auto upload」并选择 PicGo 作为上传服务后，即可自动上传图片。' })}</p>
          </div>
        </div>
      )}

      {expanded === 'vscode' && (
        <div className="rounded-xl border border-border bg-card p-6 shadow-sm space-y-5">
          <div className="rounded-lg border border-purple-500/20 bg-purple-500/5 px-4 py-3 text-sm text-muted-foreground">
            <p className="font-semibold text-foreground">{t('connections.vscodeStepsTitle', { defaultValue: '推荐扩展' })}</p>
            <ul className="mt-1 list-disc list-inside space-y-0.5">
              <li>{t('connections.vscodeExtPasteImage', { defaultValue: 'Paste Image — 支持自定义上传 URL 和响应解析，截图粘贴即上传。' })}</li>
              <li>{t('connections.vscodeExtImageSnippets', { defaultValue: 'Image Snippets — 支持自定义上传端点，粘贴或拖放图片自动上传。' })}</li>
            </ul>
          </div>

          <div className="space-y-2">
            <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">{t('connections.vscodeConfigExampleLabel', { defaultValue: 'Paste Image 配置' })}</label>
            <div className="relative group">
              <pre className="overflow-x-auto rounded-lg bg-muted/50 border border-border/50 p-4 text-sm leading-relaxed text-muted-foreground">
                <code>{configs.vscodePasteImage}</code>
              </pre>
              <button type="button" onClick={() => onCopy(configs.vscodePasteImage, 'vscode-paste')} className="absolute top-3 right-3 flex h-8 w-8 items-center justify-center rounded-md border border-border/50 bg-background backdrop-blur-sm opacity-0 shadow-sm transition-opacity duration-150 group-hover:opacity-100 hover:border-primary hover:bg-primary hover:text-primary-foreground cursor-pointer" title={t('upload.copy')}>
                <Copy className="size-4" />
              </button>
            </div>
            {copied === 'vscode-paste' && (
              <p className="rounded-lg border border-purple-500/20 bg-purple-500/5 px-3 py-2 text-sm text-purple-600">
                {t('connections.vscodeCopied', { defaultValue: 'VS Code 配置已复制，粘贴到 settings.json 中即可使用。' })}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">{t('connections.vscodeImageSnippetsTitle', { defaultValue: 'Image Snippets 配置' })}</label>
            <div className="relative group">
              <pre className="overflow-x-auto rounded-lg bg-muted/50 border border-border/50 p-4 text-sm leading-relaxed text-muted-foreground">
                <code>{configs.vscodeImageSnippets}</code>
              </pre>
              <button type="button" onClick={() => onCopy(configs.vscodeImageSnippets, 'vscode-snippets')} className="absolute top-3 right-3 flex h-8 w-8 items-center justify-center rounded-md border border-border/50 bg-background backdrop-blur-sm opacity-0 shadow-sm transition-opacity duration-150 group-hover:opacity-100 hover:border-primary hover:bg-primary hover:text-primary-foreground cursor-pointer" title={t('upload.copy')}>
                <Copy className="size-4" />
              </button>
            </div>
            {copied === 'vscode-snippets' && (
              <p className="rounded-lg border border-purple-500/20 bg-purple-500/5 px-3 py-2 text-sm text-purple-600">
                {t('connections.vscodeCopied', { defaultValue: 'VS Code 配置已复制，粘贴到 settings.json 中即可使用。' })}
              </p>
            )}
          </div>
        </div>
      )}
    </section>
  )
}

function PlusIcon({ className }: { className?: string }) {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={className}>
      <path d="M5 12h14" /><path d="M12 5v14" />
    </svg>
  )
}