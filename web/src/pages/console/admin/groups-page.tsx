import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'

import { createAdminGroup, deleteAdminGroup, listAdminGroups, updateAdminGroup } from '../../../lib/admin-api'

export function AdminGroupsPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()

  const { data: groups, isLoading, error } = useQuery({
    queryKey: ['admin-groups'],
    queryFn: listAdminGroups,
  })

  // Create
  const [showCreate, setShowCreate] = useState(false)
  const [newName, setNewName] = useState('')
  const [creating, setCreating] = useState(false)

  const handleCreate = useCallback(async () => {
    if (!newName.trim()) return
    setCreating(true)
    try {
      await createAdminGroup({ name: newName.trim() })
      setShowCreate(false)
      setNewName('')
      await qc.invalidateQueries({ queryKey: ['admin-groups'] })
    } catch (err: unknown) {
      alert((err as { response?: { data?: { message?: string } } })?.response?.data?.message || t('admin.createFailed'))
    } finally {
      setCreating(false)
    }
  }, [newName, qc, t])

  // Edit
  const [editId, setEditId] = useState<number | null>(null)
  const [editName, setEditName] = useState('')
  const [saving, setSaving] = useState(false)

  const startEdit = (g: { id: number; name: string }) => {
    setEditId(g.id)
    setEditName(g.name)
  }

  const handleUpdate = useCallback(async () => {
    if (!editId || !editName.trim()) return
    setSaving(true)
    try {
      const g = groups?.find((x) => x.id === editId)
      await updateAdminGroup(editId, { name: editName.trim(), is_default: g?.is_default ?? false, is_guest: g?.is_guest ?? false })
      setEditId(null)
      await qc.invalidateQueries({ queryKey: ['admin-groups'] })
    } catch (err: unknown) {
      alert((err as { response?: { data?: { message?: string } } })?.response?.data?.message || t('admin.saveFailed'))
    } finally {
      setSaving(false)
    }
  }, [editId, editName, groups, qc, t])

  // Delete
  const [deleting, setDeleting] = useState<number | null>(null)

  const onDelete = useCallback(async (id: number) => {
    if (!window.confirm(t('admin.confirmDeleteGroup'))) return
    setDeleting(id)
    try {
      await deleteAdminGroup(id)
      await qc.invalidateQueries({ queryKey: ['admin-groups'] })
    } catch (err: unknown) {
      alert((err as { response?: { data?: { message?: string } } })?.response?.data?.message || t('admin.deleteFailed'))
    } finally {
      setDeleting(null)
    }
  }, [qc, t])

  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">{t('admin.groupsTitle')}</h1>
        <button type="button" onClick={() => setShowCreate(true)} className="rounded-lg bg-zinc-900 px-3 py-1.5 text-sm text-white hover:bg-zinc-700 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-300">
          {t('admin.create')}
        </button>
      </div>

      {showCreate && (
        <div className="flex gap-2">
          <input value={newName} onChange={(e) => setNewName(e.target.value)} placeholder={t('admin.namePlaceholder')} className="flex-1 rounded-md border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 dark:border-zinc-700 dark:bg-zinc-800" />
          <button type="button" onClick={handleCreate} disabled={creating || !newName.trim()} className="rounded-md bg-zinc-900 px-3 py-1.5 text-sm text-white disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900">{creating ? '…' : t('admin.confirmCreate')}</button>
          <button type="button" onClick={() => { setShowCreate(false); setNewName('') }} className="rounded-md border border-zinc-300 px-3 py-1.5 text-sm dark:border-zinc-600">{t('admin.cancel')}</button>
        </div>
      )}

      {isLoading && <div className="flex justify-center py-12"><div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-400 border-t-transparent" /></div>}
      {error && <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-400">{t('admin.loadFailed')}</p>}
      {groups && groups.length === 0 && <p className="py-12 text-center text-sm text-zinc-400">{t('admin.empty')}</p>}

      {groups && groups.length > 0 && (
        <div className="divide-y divide-zinc-100 dark:divide-zinc-800">
          {groups.map((g) =>
            editId === g.id ? (
              <div key={g.id} className="flex items-center gap-2 py-3">
                <input value={editName} onChange={(e) => setEditName(e.target.value)} className="flex-1 rounded-md border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 dark:border-zinc-700 dark:bg-zinc-800" />
                <button type="button" onClick={handleUpdate} disabled={saving} className="rounded-md bg-zinc-900 px-3 py-1.5 text-sm text-white disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900">{t('admin.confirmSave')}</button>
                <button type="button" onClick={() => setEditId(null)} className="rounded-md border border-zinc-300 px-3 py-1.5 text-sm dark:border-zinc-600">{t('admin.cancel')}</button>
              </div>
            ) : (
              <div key={g.id} className="flex items-center justify-between py-3">
                <div>
                  <p className="font-medium">
                    {g.name}
                    {g.is_default && <span className="ml-2 rounded bg-blue-100 px-1.5 py-0.5 text-xs text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">{t('admin.default')}</span>}
                    {g.is_guest && <span className="ml-2 rounded bg-green-100 px-1.5 py-0.5 text-xs text-green-700 dark:bg-green-900/30 dark:text-green-400">{t('admin.guest')}</span>}
                  </p>
                  <p className="mt-0.5 text-xs text-zinc-400">{t('admin.strategiesCount', { count: g.strategy_ids?.length ?? 0 })} · {new Date(g.created_at).toLocaleDateString()}</p>
                </div>
                <div className="flex gap-1">
                  <button type="button" onClick={() => startEdit(g)} className="rounded px-2 py-1 text-xs hover:bg-zinc-100 dark:hover:bg-zinc-800">{t('admin.edit')}</button>
                  {!g.is_default && !g.is_guest && (
                    <button type="button" onClick={() => onDelete(g.id)} disabled={deleting === g.id} className="rounded px-2 py-1 text-xs text-red-500 hover:bg-red-50 disabled:opacity-50 dark:hover:bg-red-900/20">{t('admin.delete')}</button>
                  )}
                </div>
              </div>
            ),
          )}
        </div>
      )}
    </section>
  )
}
