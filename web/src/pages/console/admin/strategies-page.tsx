import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'

import { createAdminStrategy, deleteAdminStrategy, listAdminStrategies } from '../../../lib/admin-api'

export function AdminStrategiesPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()

  const { data: strategies, isLoading, error } = useQuery({
    queryKey: ['admin-strategies'],
    queryFn: listAdminStrategies,
  })

  // Create
  const [showCreate, setShowCreate] = useState(false)
  const [newName, setNewName] = useState('')
  const [newType, setNewType] = useState('local')
  const [creating, setCreating] = useState(false)

  const handleCreate = useCallback(async () => {
    if (!newName.trim()) return
    setCreating(true)
    try {
      await createAdminStrategy({
        name: newName.trim(),
        strategy_type: newType,
        configs: newType === 'local' ? { root: '', url: '' } : { endpoint: '', bucket: '', access_key: '', secret_key: '' },
      })
      setShowCreate(false)
      setNewName('')
      await qc.invalidateQueries({ queryKey: ['admin-strategies'] })
    } catch (err: unknown) {
      alert((err as { response?: { data?: { message?: string } } })?.response?.data?.message || t('admin.createFailed'))
    } finally {
      setCreating(false)
    }
  }, [newName, newType, qc, t])

  // Delete
  const [deleting, setDeleting] = useState<number | null>(null)

  const onDelete = useCallback(async (id: number) => {
    if (!window.confirm(t('admin.confirmDeleteStrategy'))) return
    setDeleting(id)
    try {
      await deleteAdminStrategy(id)
      await qc.invalidateQueries({ queryKey: ['admin-strategies'] })
    } catch (err: unknown) {
      alert((err as { response?: { data?: { message?: string } } })?.response?.data?.message || t('admin.deleteFailed'))
    } finally {
      setDeleting(null)
    }
  }, [qc, t])

  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">{t('admin.strategiesTitle')}</h1>
        <button type="button" onClick={() => setShowCreate(true)} className="rounded-lg bg-zinc-900 px-3 py-1.5 text-sm text-white hover:bg-zinc-700 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-300">
          {t('admin.create')}
        </button>
      </div>

      {showCreate && (
        <div className="space-y-3 rounded-lg border border-zinc-200 bg-zinc-50 p-4 dark:border-zinc-700 dark:bg-zinc-800">
          <input value={newName} onChange={(e) => setNewName(e.target.value)} placeholder={t('admin.namePlaceholder')} className="w-full rounded-md border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 dark:border-zinc-600 dark:bg-zinc-900" />
          <select value={newType} onChange={(e) => setNewType(e.target.value)} className="w-full rounded-md border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 dark:border-zinc-600 dark:bg-zinc-900">
            <option value="local">Local</option>
            <option value="s3">S3</option>
          </select>
          <div className="flex gap-2">
            <button type="button" onClick={handleCreate} disabled={creating || !newName.trim()} className="rounded-md bg-zinc-900 px-3 py-1.5 text-sm text-white disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900">{creating ? '…' : t('admin.confirmCreate')}</button>
            <button type="button" onClick={() => { setShowCreate(false); setNewName('') }} className="rounded-md border border-zinc-300 px-3 py-1.5 text-sm dark:border-zinc-600">{t('admin.cancel')}</button>
          </div>
        </div>
      )}

      {isLoading && <div className="flex justify-center py-12"><div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-400 border-t-transparent" /></div>}
      {error && <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-400">{t('admin.loadFailed')}</p>}
      {strategies && strategies.length === 0 && <p className="py-12 text-center text-sm text-zinc-400">{t('admin.empty')}</p>}

      {strategies && strategies.length > 0 && (
        <div className="divide-y divide-zinc-100 dark:divide-zinc-800">
          {strategies.map((s) => (
            <div key={s.id} className="flex items-center justify-between py-3">
              <div>
                <p className="font-medium">{s.name}</p>
                <p className="mt-0.5 text-xs text-zinc-400">
                  <span className={['rounded px-1.5 py-0.5', s.strategy_type === 'local' ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400' : 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400'].join(' ')}>
                    {s.strategy_type.toUpperCase()}
                  </span>
                  <span className="ml-2">{new Date(s.created_at).toLocaleDateString()}</span>
                </p>
              </div>
              <button type="button" onClick={() => onDelete(s.id)} disabled={deleting === s.id} className="rounded px-2 py-1 text-xs text-red-500 hover:bg-red-50 disabled:opacity-50 dark:hover:bg-red-900/20">{t('admin.delete')}</button>
            </div>
          ))}
        </div>
      )}
    </section>
  )
}
