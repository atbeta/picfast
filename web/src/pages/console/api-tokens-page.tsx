import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'

import { Calendar, Clock, CheckCircle2, History, Trash2, Copy, KeyRound, Plus } from 'lucide-react'
import { createApiToken, deleteApiToken, listApiTokens } from '../../lib/console-api'
import { extractErrorMessage } from '../../lib/error-handler'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { EmptyState, LoadingState } from '@/components/page-states'
import { Button } from '@/components/ui/button'
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

function isExpiringSoon(s?: string): boolean {
  if (!s || !isRealDate(s)) return false
  return new Date(s).getTime() - Date.now() < 7 * 24 * 60 * 60 * 1000
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
  const [newExpires, setNewExpires] = useState<string | undefined>(undefined)
  const [newScopes, setNewScopes] = useState<string[]>(['read', 'write'])
  const [creating, setCreating] = useState(false)
  const [createdToken, setCreatedToken] = useState<{ name: string; token: string } | null>(null)

  const handleCreate = useCallback(async () => {
    if (!newName.trim()) return
    setCreating(true)
    try {
      const result = await createApiToken(
        newName.trim(),
        newExpires,
        newScopes,
      )
      if (result.token) {
        setCreatedToken({ name: result.name, token: result.token })
      }
      setShowCreate(false)
      setNewName('')
      setNewExpires(undefined)
      setNewScopes(['read', 'write'])
      await qc.invalidateQueries({ queryKey: ['api-tokens'] })
    } catch (err: unknown) {
      toast.error(extractErrorMessage(err, t('tokens.createFailed')))
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
        toast.error(extractErrorMessage(err, t('tokens.deleteFailed')))
      } finally {
        setDeleting(null)
      }
    },
    [deleteTarget, qc, t],
  )

  const onCopy = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      toast.success(t('upload.copied'))
    } catch {
      toast.error(t('upload.copyError'))
    }
  }

  const toggleScope = (scope: string) => {
    setNewScopes((prev) =>
      prev.includes(scope) ? prev.filter((s) => s !== scope) : [...prev, scope],
    )
  }

  return (
    <section className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <h1 className="text-2xl font-bold tracking-tight">{t('page.apiTokens.title')}</h1>
        <Button size="lg" onClick={() => setShowCreate(true)}>
          <Plus className="size-4" />
          {t('tokens.create')}
        </Button>
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
            <Button
              size="icon"
              onClick={() => onCopy(createdToken.token)}
              className="shrink-0 bg-success text-success-foreground hover:opacity-90"
              title={t('tokens.copyToken')}
            >
              <Copy className="size-4" />
            </Button>
          </div>
          <Button
            variant="link"
            size="xs"
            onClick={() => setCreatedToken(null)}
            className="mt-3 text-success-foreground/70 hover:text-success-foreground"
          >
            {t('tokens.dismiss')}
          </Button>
        </div>
      )}

      {showCreate && (
        <div className="space-y-4 rounded-xl border border-border/50 bg-card p-6 shadow-sm backdrop-blur-sm">
          <h2 className="text-lg font-semibold tracking-tight">{t('tokens.create', { defaultValue: '创建令牌' })}</h2>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-1.5">
              <label className="text-sm font-medium text-foreground">{t('tokens.namePlaceholder', { defaultValue: '名称' })}</label>
              <input
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                placeholder={t('tokens.namePlaceholder')}
                className="w-full rounded-lg border border-border/50 bg-background/50 px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary/20 transition-colors duration-150"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-sm font-medium text-foreground">{t('tokens.expires', { defaultValue: '过期时间' }).replace('{{date}}', '')}</label>
              <Select
                value={newExpires ?? 'never'}
                onValueChange={(val) =>
                  setNewExpires(val === 'never' ? undefined : String(val))
                }
                items={{
                  never: t('tokens.noExpiry'),
                  '30d': `30 ${t('tokens.days')}`,
                  '90d': `90 ${t('tokens.days')}`,
                  '1y': `1 ${t('tokens.year')}`,
                }}
              >
                <SelectTrigger className="w-full bg-background/50 border-border/50">
                  <SelectValue placeholder={t('tokens.noExpiry')} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="never">{t('tokens.noExpiry')}</SelectItem>
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
            <Button
              size="lg"
              onClick={handleCreate}
              disabled={creating || !newName.trim() || newScopes.length === 0}
            >
              {creating ? t('tokens.creating') : t('tokens.confirmCreate')}
            </Button>
            <Button
              variant="outline"
              size="lg"
              onClick={() => { setShowCreate(false); setNewName(''); setNewExpires(undefined); setNewScopes(['read', 'write']) }}
            >
              {t('tokens.cancel')}
            </Button>
          </div>
        </div>
      )}

      {isLoading && (
        <LoadingState />
      )}
      {error && (
        <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {t('tokens.loadFailed')}
        </p>
      )}
      {tokens && tokens.length === 0 && (
        <EmptyState
          icon={<KeyRound className="size-6 text-muted-foreground" />}
          title={t('tokens.empty')}
          description={t('tokens.emptyDesc', { defaultValue: '创建令牌后，就可以连接 MCP、ShareX 或其他自动化工具。' })}
        />
      )}

      {tokens && tokens.length > 0 && (
        <>
          <div className="space-y-3 md:hidden">
            {tokens.map((tk) => (
              <div
                key={tk.id}
                className="rounded-xl border border-border/50 bg-card p-4 shadow-sm transition-colors hover:border-primary/20"
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0 space-y-2">
                    <div className="flex items-center gap-2">
                      <div className="rounded-lg bg-muted/60 p-2">
                        <KeyRound className="size-4 text-muted-foreground" />
                      </div>
                      <div className="min-w-0">
                        <p className="truncate text-sm font-semibold text-foreground">{tk.name}</p>
                        <p className="text-xs text-muted-foreground">
                          {t('tokens.namePlaceholder', { defaultValue: '名称' })}
                        </p>
                      </div>
                    </div>
                    <div className="flex flex-wrap gap-1.5">
                      {tk.scopes.map((scope) => (
                        <span
                          key={scope}
                          className="rounded-full bg-primary/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider text-primary"
                        >
                          {scope}
                        </span>
                      ))}
                    </div>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    onClick={() => setDeleteTarget(tk.id)}
                    disabled={deleting === tk.id}
                    className="shrink-0 text-destructive/70 hover:text-destructive"
                    title={t('tokens.delete')}
                  >
                    <Trash2 className="size-4" />
                  </Button>
                </div>

                <div className="mt-4 grid gap-3 rounded-lg border border-border/40 bg-muted/20 p-3 text-xs sm:grid-cols-2">
                  <div className="space-y-1">
                    <p className="font-medium text-muted-foreground">
                      {t('tokens.lastUsedAt', { defaultValue: '上次使用' }).replace('{{date}}', '')}
                    </p>
                    {isRealDate(tk.last_used_at) ? (
                      <div className="flex items-center gap-1.5 text-foreground">
                        <History className="size-3.5 text-muted-foreground" />
                        <span>{formatDate(tk.last_used_at)}</span>
                      </div>
                    ) : (
                      <span className="inline-flex rounded-full bg-muted px-2 py-0.5 text-[11px] font-medium text-muted-foreground">
                        {t('tokens.neverUsed', { defaultValue: '从未使用' })}
                      </span>
                    )}
                  </div>
                  <div className="space-y-1">
                    <p className="font-medium text-muted-foreground">
                      {t('tokens.expires', { defaultValue: '过期时间' }).replace('{{date}}', '')}
                    </p>
                    {isRealDate(tk.expires_at) ? (
                      <div className={`flex items-center gap-1.5 ${isExpiringSoon(tk.expires_at) ? 'text-destructive/90' : 'text-amber-500/90'}`}>
                        <Clock className="size-3.5" />
                        <span>{formatDate(tk.expires_at)}</span>
                      </div>
                    ) : (
                      <div className="flex items-center gap-1.5 text-success/90">
                        <CheckCircle2 className="size-3.5" />
                        <span>{t('tokens.noExpiry')}</span>
                      </div>
                    )}
                  </div>
                  <div className="space-y-1 sm:col-span-2">
                    <p className="font-medium text-muted-foreground">
                      {t('common.createdAt', { defaultValue: '创建时间' })}
                    </p>
                    <div className="flex items-center gap-1.5 text-foreground">
                      <Calendar className="size-3.5 text-muted-foreground" />
                      <span>{formatDate(tk.created_at)}</span>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>

          <div className="hidden overflow-hidden rounded-xl border border-border/50 bg-card shadow-sm md:block">
            <div className="overflow-x-auto">
            <table className="w-full text-left text-sm">
              <thead className="border-b border-border/50 bg-muted/50 text-muted-foreground">
                <tr>
                  <th className="px-4 py-3 font-medium">{t('tokens.namePlaceholder', { defaultValue: '名称' })}</th>
                  <th className="px-4 py-3 font-medium">{t('tokens.scopes', { defaultValue: '权限' })}</th>
                  <th className="px-4 py-3 font-medium">{t('tokens.lastUsedAt', { defaultValue: '上次使用' }).replace('{{date}}', '')}</th>
                  <th className="px-4 py-3 font-medium">{t('tokens.createdColumn', { defaultValue: '创建时间' })}</th>
                  <th className="px-4 py-3 font-medium">{t('tokens.expiresColumn', { defaultValue: '过期时间' })}</th>
                  <th className="px-4 py-3 font-medium">{t('common.actions', { defaultValue: '操作' })}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/50">
                {tokens.map((tk) => (
                  <tr key={tk.id} className="group hover:bg-muted/50 transition-colors">
                    <td className="px-4 py-3 font-medium text-foreground">
                      <div className="flex items-center gap-2">
                        <KeyRound className="size-4 text-muted-foreground" />
                        {tk.name}
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex gap-1.5 flex-wrap">
                        {tk.scopes.map((scope) => (
                          <span key={scope} className="rounded-full bg-primary/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider text-primary">
                            {scope}
                          </span>
                        ))}
                      </div>
                    </td>
                    <td className="px-4 py-3 text-muted-foreground">
                      {isRealDate(tk.last_used_at) ? (
                        <div className="flex items-center gap-1.5">
                          <History className="size-3.5" />
                          {formatDate(tk.last_used_at)}
                        </div>
                      ) : (
                        <span className="inline-flex rounded-full bg-muted px-2 py-0.5 text-[11px] font-medium text-muted-foreground">
                          {t('tokens.neverUsed', { defaultValue: '从未使用' })}
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-muted-foreground">
                      <div className="flex items-center gap-1.5">
                        <Calendar className="size-3.5" />
                        {formatDate(tk.created_at)}
                      </div>
                    </td>
                    <td className="px-4 py-3 text-muted-foreground">
                      {isRealDate(tk.expires_at) ? (
                        <div className={`flex items-center gap-1.5 ${isExpiringSoon(tk.expires_at) ? 'text-destructive/90' : 'text-amber-500/90'}`}>
                          <Clock className="size-3.5" />
                          {formatDate(tk.expires_at)}
                        </div>
                      ) : (
                        <div className="flex items-center gap-1.5 text-success/90">
                          <CheckCircle2 className="size-3.5" />
                          {t('tokens.noExpiry')}
                        </div>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        onClick={() => setDeleteTarget(tk.id)}
                        disabled={deleting === tk.id}
                        className="text-destructive/70 hover:text-destructive"
                        title={t('tokens.delete')}
                      >
                        <Trash2 className="size-4" />
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          </div>
        </>
      )}



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
