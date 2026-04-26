import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'

import { createApiToken, deleteApiToken, listApiTokens } from '../../lib/console-api'

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
  const [creating, setCreating] = useState(false)
  const [createdToken, setCreatedToken] = useState<{ name: string; token: string } | null>(null)

  const handleCreate = useCallback(async () => {
    if (!newName.trim()) return
    setCreating(true)
    try {
      const result = await createApiToken(
        newName.trim(),
        newExpires || undefined,
        ['read', 'write'],
      )
      if (result.token) {
        setCreatedToken({ name: result.name, token: result.token })
      }
      setShowCreate(false)
      setNewName('')
      setNewExpires('')
      await qc.invalidateQueries({ queryKey: ['api-tokens'] })
    } catch (err: unknown) {
      alert(
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
        t('tokens.createFailed'),
      )
    } finally {
      setCreating(false)
    }
  }, [newName, newExpires, qc, t])

  // Delete
  const [deleting, setDeleting] = useState<number | null>(null)

  const handleDelete = useCallback(
    async (id: number) => {
      if (!window.confirm(t('tokens.confirmDelete'))) return
      setDeleting(id)
      try {
        await deleteApiToken(id)
        await qc.invalidateQueries({ queryKey: ['api-tokens'] })
      } catch (err: unknown) {
        alert(
          (err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
          t('tokens.deleteFailed'),
        )
      } finally {
        setDeleting(null)
      }
    },
    [qc, t],
  )

  const onCopy = async (text: string) => {
    await navigator.clipboard.writeText(text)
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
          <div className="flex gap-2">
            <button
              type="button"
              onClick={handleCreate}
              disabled={creating || !newName.trim()}
              className="rounded-md bg-zinc-900 px-3 py-1.5 text-sm text-white disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900"
            >
              {creating ? t('tokens.creating') : t('tokens.confirmCreate')}
            </button>
            <button
              type="button"
              onClick={() => { setShowCreate(false); setNewName(''); setNewExpires('') }}
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
        <p className="py-12 text-center text-sm text-zinc-400">{t('tokens.empty')}</p>
      )}

      {tokens && tokens.length > 0 && (
        <div className="divide-y divide-zinc-100 dark:divide-zinc-800">
          {tokens.map((tk) => (
            <div key={tk.id} className="flex items-center justify-between py-3">
              <div className="min-w-0 flex-1">
                <p className="font-medium">{tk.name}</p>
                <div className="mt-0.5 flex flex-wrap gap-2 text-xs text-zinc-400">
                  <span>{tk.scopes.join(', ')}</span>
                  {tk.expires_at && <span>{t('tokens.expires', { date: new Date(tk.expires_at).toLocaleDateString() })}</span>}
                  <span>{t('tokens.createdAt', { date: new Date(tk.created_at).toLocaleDateString() })}</span>
                </div>
              </div>
              <button
                type="button"
                onClick={() => handleDelete(tk.id)}
                disabled={deleting === tk.id}
                className="rounded px-2 py-1 text-xs text-red-500 hover:bg-red-50 disabled:opacity-50 dark:hover:bg-red-900/20"
              >
                {t('tokens.delete')}
              </button>
            </div>
          ))}
        </div>
      )}
    </section>
  )
}
