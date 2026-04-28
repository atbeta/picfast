import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'

import { PlugIcon, MonitorIcon } from 'lucide-react'
import { createApiToken, deleteApiToken, listApiTokens } from '../../lib/console-api'
import { ConfirmDialog } from '@/components/confirm-dialog'

function isRealDate(s?: string): boolean {
  if (!s) return false
  const d = new Date(s)
  return !isNaN(d.getTime()) && d.getFullYear() > 1
}

function formatDate(s?: string): string {
  if (!s) return '-'
  return new Date(s).toLocaleString()
}

export function ApiTokensPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()

  const { data: tokens, isLoading, error } = useQuery({
    queryKey: ['api-tokens'],
    queryFn: listApiTokens,
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

  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">{t('page.apiTokens.title')}</h1>
        <button
          type="button"
          onClick={() => setShowCreate(true)}
          className="rounded-lg bg-zinc-900 px-3 py-1.5 text-sm text-white hover:bg-zinc-700 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-300"
        >
          {t('tokens.create')}
        </button>
      </div>

      {createdToken && (
        <div className="rounded-lg border border-amber-300 bg-amber-50 p-4 dark:border-amber-700 dark:bg-amber-900/20">
          <p className="text-sm font-medium text-amber-800 dark:text-amber-200">
            {t('tokens.createdWarning')}
          </p>
          <div className="mt-2 flex items-center gap-2">
            <code className="min-w-0 flex-1 truncate rounded bg-amber-100 px-2 py-1 text-xs dark:bg-amber-900/40">
              {createdToken.token}
            </code>
            <button
              type="button"
              onClick={() => onCopy(createdToken.token)}
              className="shrink-0 rounded bg-amber-200 px-2 py-1 text-xs hover:bg-amber-300 dark:bg-amber-800 dark:hover:bg-amber-700"
            >
              {t('tokens.copyToken')}
            </button>
          </div>
          <button
            type="button"
            onClick={() => setCreatedToken(null)}
            className="mt-2 text-xs text-amber-600 hover:underline dark:text-amber-400"
          >
            {t('tokens.dismiss')}
          </button>
        </div>
      )}

      {showCreate && (
        <div className="space-y-3 rounded-lg border border-zinc-200 bg-zinc-50 p-4 dark:border-zinc-700 dark:bg-zinc-800">
          <input
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            placeholder={t('tokens.namePlaceholder')}
            className="w-full rounded-md border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 dark:border-zinc-600 dark:bg-zinc-900"
          />
          <select
            value={newExpires}
            onChange={(e) => setNewExpires(e.target.value)}
            className="w-full rounded-md border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 dark:border-zinc-600 dark:bg-zinc-900"
          >
            <option value="">{t('tokens.noExpiry')}</option>
            <option value="30d">30 {t('tokens.days')}</option>
            <option value="90d">90 {t('tokens.days')}</option>
            <option value="1y">1 {t('tokens.year')}</option>
          </select>

          {/* Scope selection */}
          <div>
            <label className="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">{t('tokens.scopes', { defaultValue: '权限' })}</label>
            <div className="flex gap-4">
              <label className="flex items-center gap-1.5 text-sm text-zinc-700 dark:text-zinc-300">
                <input type="checkbox" checked={newScopes.includes('read')} onChange={() => toggleScope('read')} className="h-4 w-4 rounded border-zinc-300 dark:border-zinc-600" />
                read
              </label>
              <label className="flex items-center gap-1.5 text-sm text-zinc-700 dark:text-zinc-300">
                <input type="checkbox" checked={newScopes.includes('write')} onChange={() => toggleScope('write')} className="h-4 w-4 rounded border-zinc-300 dark:border-zinc-600" />
                write
              </label>
            </div>
          </div>

          <div className="flex gap-2">
            <button
              type="button"
              onClick={handleCreate}
              disabled={creating || !newName.trim() || newScopes.length === 0}
              className="rounded-md bg-zinc-900 px-3 py-1.5 text-sm text-white disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900"
            >
              {creating ? t('tokens.creating') : t('tokens.confirmCreate')}
            </button>
            <button
              type="button"
              onClick={() => { setShowCreate(false); setNewName(''); setNewExpires(''); setNewScopes(['read', 'write']) }}
              className="rounded-md border border-zinc-300 px-3 py-1.5 text-sm dark:border-zinc-600"
            >
              {t('tokens.cancel')}
            </button>
          </div>
        </div>
      )}

      {isLoading && (
        <div className="flex justify-center py-12">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-400 border-t-transparent" />
        </div>
      )}
      {error && (
        <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-400">
          {t('tokens.loadFailed')}
        </p>
      )}
      {tokens && tokens.length === 0 && (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div className="mb-3 rounded-full bg-muted p-3">
            <svg className="size-6 text-muted-foreground" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" d="M15.75 5.25a3 3 0 0 1 3 3m3 0a6 6 0 0 1-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 1 1 21.75 8.25Z" /></svg>
          </div>
          <p className="text-sm font-medium text-muted-foreground">{t('tokens.empty')}</p>
        </div>
      )}

      {tokens && tokens.length > 0 && (
        <div className="divide-y divide-zinc-100 dark:divide-zinc-800">
          {tokens.map((tk) => (
            <div key={tk.id} className="flex items-center justify-between py-3">
              <div className="min-w-0 flex-1">
                <p className="font-medium">{tk.name}</p>
                <div className="mt-0.5 text-xs text-zinc-400">
                  {t('tokens.createdAt', { date: formatDate(tk.created_at) })}
                  {isRealDate(tk.expires_at) && (
                    <span className="ml-2">{t('tokens.expires', { date: formatDate(tk.expires_at) })}</span>
                  )}
                  {!isRealDate(tk.expires_at) && (
                    <span className="ml-2">{t('tokens.noExpiry')}</span>
                  )}
                  {isRealDate(tk.last_used_at) && (
                    <span className="ml-2">{t('tokens.lastUsedAt', { defaultValue: '上次使用' })} {formatDate(tk.last_used_at)}</span>
                  )}
                </div>
                <div className="mt-1.5 flex gap-1.5">
                  {tk.scopes.map((scope) => (
                    <span key={scope} className="rounded bg-blue-100 px-1.5 py-0.5 text-xs text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">
                      {scope}
                    </span>
                  ))}
                </div>
              </div>
              <button
                type="button"
                onClick={() => setDeleteTarget(tk.id)}
                disabled={deleting === tk.id}
                className="rounded px-2 py-1 text-xs text-red-500 hover:bg-red-50 disabled:opacity-50 dark:hover:bg-red-900/20"
              >
                {t('tokens.delete')}
              </button>
            </div>
          ))}
        </div>
      )}

      {/* Integration cards */}
      <div className="border-t pt-6">
        <h2 className="mb-4 text-sm font-semibold">{t('integrations.title')}</h2>
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
                <code className="min-w-0 flex-1 truncate rounded bg-muted px-2 py-1 text-xs">{window.location.origin}/mcp</code>
                <button type="button" onClick={() => onCopy(`${window.location.origin}/mcp`)} className="shrink-0 rounded bg-muted px-2 py-1 text-xs hover:bg-muted/80">{t('upload.copy')}</button>
              </div>
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
                <code className="min-w-0 flex-1 truncate rounded bg-muted px-2 py-1 text-xs">{window.location.origin}/api/v1/sharex/config</code>
                <button type="button" onClick={() => onCopy(`${window.location.origin}/api/v1/sharex/config`)} className="shrink-0 rounded bg-muted px-2 py-1 text-xs hover:bg-muted/80">{t('upload.copy')}</button>
              </div>
            </div>
            <a
              href="/api/v1/sharex/config"
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
