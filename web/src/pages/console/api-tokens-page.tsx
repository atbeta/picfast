import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'

import { PlugIcon, MonitorIcon, Copy, Trash2, KeyRound, Calendar, Clock, CheckCircle2, History } from 'lucide-react'
import { createApiToken, deleteApiToken, listApiTokens } from '../../lib/console-api'
import { getSiteConfig } from '../../lib/site-config'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'

function isRealDate(s?: string): boolean {
  if (!s) return false
  const d = new Date(s)
  return !isNaN(d.getTime()) && d.getFullYear() > 1
}

function formatDate(s?: string): string {
  if (!s) return '-'
  return new Date(s).toLocaleString()
}

function trimTrailingSlash(value: string): string {
  return value.replace(/\/+$/, '')
}

export function ApiTokensPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()

  const { data: tokens, isLoading, error } = useQuery({
    queryKey: ['api-tokens'],
    queryFn: listApiTokens,
  })
  const { data: siteConfig } = useQuery({
    queryKey: ['site-config'],
    queryFn: getSiteConfig,
  })

  // Create
  const [showCreate, setShowCreate] = useState(false)
  const [newName, setNewName] = useState('')
  const [newExpires, setNewExpires] = useState('')
  const [newScopes, setNewScopes] = useState<string[]>(['read', 'write'])
  const [creating, setCreating] = useState(false)
  const [createdToken, setCreatedToken] = useState<{ name: string; token: string } | null>(null)

  const handleCreate = useCallback(async () => {
    if (!newName.trim()) return
    setCreating(true)
    try {
      const result = await createApiToken(
        newName.trim(),
        newExpires || undefined,
        newScopes,
      )
      if (result.token) {
        setCreatedToken({ name: result.name, token: result.token })
      }
      setShowCreate(false)
      setNewName('')
      setNewExpires('')
      setNewScopes(['read', 'write'])
      await qc.invalidateQueries({ queryKey: ['api-tokens'] })
    } catch (err: unknown) {
      toast.error(
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
        t('tokens.createFailed'),
      )
    } finally {
      setCreating(false)
    }
  }, [newName, newExpires, newScopes, qc, t])

  // Delete
  const [deleting, setDeleting] = useState<number | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<number | null>(null)

  const handleDelete = useCallback(
    async () => {
      if (deleteTarget === null) return
      setDeleting(deleteTarget)
      try {
        await deleteApiToken(deleteTarget)
        setDeleteTarget(null)
        await qc.invalidateQueries({ queryKey: ['api-tokens'] })
      } catch (err: unknown) {
        toast.error(
          (err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
          t('tokens.deleteFailed'),
        )
      } finally {
        setDeleting(null)
      }
    },
    [deleteTarget, qc, t],
  )

  const onCopy = async (text: string) => {
    await navigator.clipboard.writeText(text)
    toast.success(t('upload.copied'))
  }

  const toggleScope = (scope: string) => {
    setNewScopes((prev) =>
      prev.includes(scope) ? prev.filter((s) => s !== scope) : [...prev, scope],
    )
  }

  const baseURL = trimTrailingSlash(siteConfig?.base_url || window.location.origin)
  const mcpEndpoint = `${baseURL}/mcp`
  const sharexConfigURL = `${baseURL}/api/v1/sharex/config`
  const mcpConfigExample = `{
  "mcpServers": {
    "picfast": {
      "url": "${mcpEndpoint}",
      "headers": {
        "Authorization": "Bearer ${createdToken?.token || '<YOUR_API_TOKEN>'}"
      }
    }
  }
}`

  return (
    <section className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight">{t('page.apiTokens.title')}</h1>
        <button
          type="button"
          onClick={() => setShowCreate(true)}
          className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow-sm transition-all hover:opacity-90 active:scale-95"
        >
          {t('tokens.create')}
        </button>
      </div>

      {createdToken && (
        <div className="rounded-xl border border-success/30 bg-success/10 p-5 shadow-sm">
          <p className="text-sm font-medium text-success-foreground">
            {t('tokens.createdWarning')}
          </p>
          <div className="mt-3 flex items-center gap-2">
            <code className="min-w-0 flex-1 truncate rounded-lg bg-background/50 backdrop-blur-sm px-3 py-1.5 text-xs font-mono border border-border/50">
              {createdToken.token}
            </code>
            <button
              type="button"
              onClick={() => onCopy(createdToken.token)}
              className="shrink-0 flex items-center justify-center rounded-lg bg-success text-success-foreground w-8 h-8 shadow-sm transition-all hover:opacity-90 active:scale-95 cursor-pointer"
              title={t('tokens.copyToken')}
            >
              <Copy className="size-4" />
            </button>
          </div>
          <button
            type="button"
            onClick={() => setCreatedToken(null)}
            className="mt-3 text-xs font-medium text-success-foreground/70 hover:text-success-foreground transition-colors"
          >
            {t('tokens.dismiss')}
          </button>
        </div>
      )}

      {showCreate && (
        <div className="space-y-4 rounded-xl border border-border/50 bg-card p-6 shadow-sm backdrop-blur-sm animate-in slide-in-from-top-2 fade-in duration-300">
          <h2 className="text-lg font-semibold tracking-tight">{t('tokens.create', { defaultValue: '创建令牌' })}</h2>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-1.5">
              <label className="text-sm font-medium text-foreground">{t('tokens.namePlaceholder', { defaultValue: '名称' })}</label>
              <input
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                placeholder={t('tokens.namePlaceholder')}
                className="w-full rounded-lg border border-border/50 bg-background/50 px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary/20 transition-all"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-sm font-medium text-foreground">{t('tokens.expires', { defaultValue: '过期时间' }).replace('{{date}}', '')}</label>
              <Select 
                value={newExpires} 
                onValueChange={(val) => val !== null && setNewExpires(val as string)}
                items={{
                  '': t('tokens.noExpiry'),
                  '30d': `30 ${t('tokens.days')}`,
                  '90d': `90 ${t('tokens.days')}`,
                  '1y': `1 ${t('tokens.year')}`
                }}
              >
                <SelectTrigger className="w-full bg-background/50 border-border/50">
                  <SelectValue placeholder={t('tokens.noExpiry')} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">{t('tokens.noExpiry')}</SelectItem>
                  <SelectItem value="30d">30 {t('tokens.days')}</SelectItem>
                  <SelectItem value="90d">90 {t('tokens.days')}</SelectItem>
                  <SelectItem value="1y">1 {t('tokens.year')}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* Scope selection */}
          <div className="space-y-2 pt-2">
            <label className="block text-sm font-medium text-foreground">{t('tokens.scopes', { defaultValue: '权限' })}</label>
            <div className="flex gap-6">
              <label className="flex items-center gap-2 cursor-pointer group">
                <Switch checked={newScopes.includes('read')} onCheckedChange={() => toggleScope('read')} />
                <span className="text-sm text-muted-foreground group-hover:text-foreground transition-colors">read</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer group">
                <Switch checked={newScopes.includes('write')} onCheckedChange={() => toggleScope('write')} />
                <span className="text-sm text-muted-foreground group-hover:text-foreground transition-colors">write</span>
              </label>
            </div>
          </div>

          <div className="flex items-center gap-3 pt-4 border-t border-border/50">
            <button
              type="button"
              onClick={handleCreate}
              disabled={creating || !newName.trim() || newScopes.length === 0}
              className="rounded-lg bg-primary px-5 py-2 text-sm font-medium text-primary-foreground shadow-sm transition-all hover:opacity-90 disabled:opacity-50 active:scale-95"
            >
              {creating ? t('tokens.creating') : t('tokens.confirmCreate')}
            </button>
            <button
              type="button"
              onClick={() => { setShowCreate(false); setNewName(''); setNewExpires(''); setNewScopes(['read', 'write']) }}
              className="rounded-lg border border-border/50 bg-background px-5 py-2 text-sm font-medium text-muted-foreground shadow-sm transition-all hover:bg-muted hover:text-foreground active:scale-95"
            >
              {t('tokens.cancel')}
            </button>
          </div>
        </div>
      )}

      {isLoading && (
        <div className="flex justify-center py-12">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
        </div>
      )}
      {error && (
        <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {t('tokens.loadFailed')}
        </p>
      )}
      {tokens && tokens.length === 0 && (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div className="mb-3 rounded-full bg-muted p-3">
            <KeyRound className="size-6 text-muted-foreground" />
          </div>
          <p className="text-sm font-medium text-muted-foreground">{t('tokens.empty')}</p>
        </div>
      )}

      {tokens && tokens.length > 0 && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {tokens.map((tk) => (
            <div key={tk.id} className="group relative flex flex-col justify-between rounded-xl border border-border/50 bg-card p-5 shadow-sm transition-all hover:shadow-md hover:-translate-y-1 hover:border-primary/30">
              <div>
                <div className="flex items-start justify-between">
                  <p className="font-semibold text-foreground">{tk.name}</p>
                  <button
                    type="button"
                    onClick={() => setDeleteTarget(tk.id)}
                    disabled={deleting === tk.id}
                    className="opacity-0 group-hover:opacity-100 transition-opacity rounded-lg p-1.5 text-destructive/70 hover:bg-destructive/10 hover:text-destructive disabled:opacity-50 cursor-pointer"
                    title={t('tokens.delete')}
                  >
                    <Trash2 className="size-4" />
                  </button>
                </div>
                <div className="mt-3 flex flex-wrap gap-2 text-xs text-muted-foreground">
                  <div className="flex items-center gap-1.5">
                    <Calendar className="size-3.5" />
                    <span>{formatDate(tk.created_at)}</span>
                  </div>
                  {isRealDate(tk.expires_at) ? (
                    <div className="flex items-center gap-1.5">
                      <Clock className="size-3.5 text-amber-500/70" />
                      <span className="text-amber-500/90">{t('tokens.expires', { date: formatDate(tk.expires_at) })}</span>
                    </div>
                  ) : (
                    <div className="flex items-center gap-1.5">
                      <CheckCircle2 className="size-3.5 text-success/70" />
                      <span className="text-success/90">{t('tokens.noExpiry')}</span>
                    </div>
                  )}
                  {isRealDate(tk.last_used_at) && (
                    <div className="flex items-center gap-1.5 w-full mt-1">
                      <History className="size-3.5" />
                      <span>{t('tokens.lastUsedAt', { defaultValue: '上次使用' })} {formatDate(tk.last_used_at)}</span>
                    </div>
                  )}
                </div>
              </div>
              <div className="mt-5 flex gap-2">
                {tk.scopes.map((scope) => (
                  <span key={scope} className="rounded-lg bg-primary/10 px-2 py-1 text-[10px] font-medium uppercase tracking-wider text-primary">
                    {scope}
                  </span>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Integration cards */}
      <div className="border-t border-border/50 pt-6 mt-8">
        <h2 className="mb-5 text-lg font-semibold tracking-tight">{t('integrations.title')}</h2>
        <div className="grid gap-4 sm:grid-cols-2">
          {/* MCP Server card */}
          <div className="rounded-xl border border-border bg-card p-4">
            <div className="mb-3 flex items-center gap-2">
              <div className="rounded-lg bg-info/10 p-1.5">
                <PlugIcon className="size-4 text-info" />
              </div>
              <h3 className="text-sm font-medium">{t('integrations.mcpTitle')}</h3>
            </div>
            <p className="text-xs text-muted-foreground">{t('integrations.mcpDesc')}</p>
            <div className="mt-3 space-y-2">
              <label className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">{t('integrations.mcpEndpoint')}</label>
              <div className="flex items-center gap-1.5">
                <code className="min-w-0 flex-1 truncate rounded bg-muted px-2 py-1.5 text-xs">{mcpEndpoint}</code>
                <button type="button" onClick={() => onCopy(mcpEndpoint)} className="shrink-0 flex items-center justify-center rounded bg-muted w-7 h-7 hover:bg-muted/80 hover:text-foreground text-muted-foreground transition-colors cursor-pointer" title={t('upload.copy')}>
                  <Copy className="size-3.5" />
                </button>
              </div>
            </div>
            <div className="mt-3 space-y-2">
              <label className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">{t('integrations.mcpConfigExample')}</label>
              <pre className="overflow-x-auto rounded bg-muted px-3 py-2 text-[11px] leading-5 text-muted-foreground">
                <code>{mcpConfigExample}</code>
              </pre>
              <p className="text-[11px] text-muted-foreground">{t('integrations.mcpConfigHint')}</p>
            </div>
            <div className="mt-3">
              <label className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">{t('integrations.quickStart')}</label>
              <ol className="mt-1 list-inside list-decimal text-xs text-muted-foreground space-y-0.5">
                <li>{t('integrations.step1')}</li>
                <li>{t('integrations.step2')}</li>
                <li>{t('integrations.step3')}</li>
              </ol>
            </div>
          </div>

          {/* ShareX card */}
          <div className="rounded-xl border border-border bg-card p-4">
            <div className="mb-3 flex items-center gap-2">
              <div className="rounded-lg bg-success/10 p-1.5">
                <MonitorIcon className="size-4 text-success" />
              </div>
              <h3 className="text-sm font-medium">{t('integrations.sharexTitle')}</h3>
            </div>
            <p className="text-xs text-muted-foreground">{t('integrations.sharexDesc')}</p>
            <div className="mt-3 space-y-2">
              <label className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">{t('integrations.sharexEndpoint')}</label>
              <div className="flex items-center gap-1.5">
                <code className="min-w-0 flex-1 truncate rounded bg-muted px-2 py-1.5 text-xs">{sharexConfigURL}</code>
                <button type="button" onClick={() => onCopy(sharexConfigURL)} className="shrink-0 flex items-center justify-center rounded bg-muted w-7 h-7 hover:bg-muted/80 hover:text-foreground text-muted-foreground transition-colors cursor-pointer" title={t('upload.copy')}>
                  <Copy className="size-3.5" />
                </button>
              </div>
            </div>
            <a
              href={sharexConfigURL}
              className="mt-3 inline-flex items-center gap-1.5 rounded-lg bg-primary px-3 py-1.5 text-xs font-medium text-primary-foreground hover:opacity-90"
            >
              {t('integrations.downloadConfig')}
            </a>
          </div>
        </div>
      </div>

      <ConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}
        title={t('tokens.confirmDelete')}
        description={t('tokens.deleteDescription')}
        confirmLabel={t('tokens.delete')}
        onConfirm={handleDelete}
        loading={!!deleting}
      />
    </section>
  )
}
