import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'

import { deleteAdminUser, listAdminUsers, updateAdminUser } from '../../../lib/admin-api'
import { formatFileSize } from '../../../lib/upload'
import { ConfirmDialog } from '@/components/confirm-dialog'

export function AdminUsersPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')

  const { data, isLoading, error } = useQuery({
    queryKey: ['admin-users', page, keyword],
    queryFn: () => listAdminUsers({ page, page_size: 20, keyword: keyword || undefined }),
  })

  const [saving, setSaving] = useState<number | null>(null)
  const [deleting, setDeleting] = useState<number | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<number | null>(null)

  const toggleStatus = useCallback(
    async (user: { id: number; status: number }) => {
      setSaving(user.id)
      try {
        await updateAdminUser(user.id, { status: user.status === 1 ? 0 : 1 })
        await qc.invalidateQueries({ queryKey: ['admin-users'] })
      } catch (err: unknown) {
        toast.error((err as { response?: { data?: { message?: string } } })?.response?.data?.message || t('admin.saveFailed'))
      } finally {
        setSaving(null)
      }
    },
    [qc, t],
  )

  const onDelete = useCallback(
    async () => {
      if (deleteTarget === null) return
      setDeleting(deleteTarget)
      try {
        await deleteAdminUser(deleteTarget)
        setDeleteTarget(null)
        await qc.invalidateQueries({ queryKey: ['admin-users'] })
      } catch (err: unknown) {
        toast.error((err as { response?: { data?: { message?: string } } })?.response?.data?.message || t('admin.deleteFailed'))
      } finally {
        setDeleting(null)
      }
    },
    [deleteTarget, qc, t],
  )

  return (
    <section className="space-y-4">
      <h1 className="text-xl font-semibold">{t('admin.usersTitle')}</h1>

      <input
        value={keyword}
        onChange={(e) => { setKeyword(e.target.value); setPage(1) }}
        placeholder={t('admin.searchPlaceholder')}
        className="w-full max-w-xs rounded-md border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 dark:border-zinc-700 dark:bg-zinc-800"
      />

      {isLoading && (
        <div className="flex justify-center py-12"><div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-400 border-t-transparent" /></div>
      )}
      {error && <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-400">{t('admin.loadFailed')}</p>}
      {data && data.items.length === 0 && <p className="py-12 text-center text-sm text-zinc-400">{t('admin.empty')}</p>}

      {data && data.items.length > 0 && (
        <>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-zinc-200 text-left text-xs text-zinc-500 dark:border-zinc-700">
                  <th className="pb-2 pr-3 font-medium">ID</th>
                  <th className="pb-2 pr-3 font-medium">{t('admin.colEmail')}</th>
                  <th className="pb-2 pr-3 font-medium">{t('admin.colName')}</th>
                  <th className="pb-2 pr-3 font-medium">{t('admin.colRole')}</th>
                  <th className="pb-2 pr-3 font-medium">{t('admin.colStatus')}</th>
                  <th className="pb-2 pr-3 font-medium">{t('admin.colImages')}</th>
                  <th className="pb-2 pr-3 font-medium">{t('admin.usedCapacity', { defaultValue: '已用容量' })}</th>
                  <th className="pb-2 font-medium">{t('admin.colActions')}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-100 dark:divide-zinc-800">
                {data.items.map((u) => (
                  <tr key={u.id}>
                    <td className="py-2 pr-3 text-zinc-500">{u.id}</td>
                    <td className="py-2 pr-3">{u.email}</td>
                    <td className="py-2 pr-3">{u.name}</td>
                    <td className="py-2 pr-3">
                      <span className={['rounded px-1.5 py-0.5 text-xs', u.role === 'admin' ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400' : 'bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-400'].join(' ')}>
                        {u.role}
                      </span>
                    </td>
                    <td className="py-2 pr-3">
                      <span className={['rounded px-1.5 py-0.5 text-xs', u.status === 1 ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'].join(' ')}>
                        {u.status === 1 ? t('admin.active') : t('admin.frozen')}
                      </span>
                    </td>
                    <td className="py-2 pr-3 text-zinc-500">{u.image_num}</td>
                    <td className="py-2 pr-3 text-zinc-500">{formatFileSize(u.used_capacity || 0)} / {formatFileSize(u.capacity_bytes)}</td>
                    <td className="py-2">
                      <div className="flex gap-1">
                        {u.role !== 'admin' && (
                          <button
                            type="button"
                            onClick={() => toggleStatus(u)}
                            disabled={saving === u.id}
                            className="rounded px-2 py-1 text-xs hover:bg-zinc-100 disabled:opacity-50 dark:hover:bg-zinc-800"
                          >
                            {u.status === 1 ? t('admin.freeze') : t('admin.activate')}
                          </button>
                        )}
                        {u.role !== 'admin' && (
                          <button
                            type="button"
                            onClick={() => setDeleteTarget(u.id)}
                            disabled={deleting === u.id}
                            className="rounded px-2 py-1 text-xs text-red-500 hover:bg-red-50 disabled:opacity-50 dark:hover:bg-red-900/20"
                          >
                            {t('admin.delete')}
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {data.total > 20 && (
            <div className="flex items-center justify-between pt-2">
              <span className="text-xs text-zinc-500">{t('admin.pagination', { total: data.total })}</span>
              <div className="flex gap-2">
                <button type="button" disabled={page <= 1} onClick={() => setPage((p) => p - 1)} className="rounded border border-zinc-300 px-3 py-1 text-xs disabled:opacity-40 dark:border-zinc-700">{t('admin.prev')}</button>
                <button type="button" disabled={page * 20 >= data.total} onClick={() => setPage((p) => p + 1)} className="rounded border border-zinc-300 px-3 py-1 text-xs disabled:opacity-40 dark:border-zinc-700">{t('admin.next')}</button>
              </div>
            </div>
          )}
        </>
      )}

      <ConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}
        title={t('admin.confirmDeleteUser')}
        destructive
        confirmLabel={t('admin.delete')}
        onConfirm={onDelete}
        loading={!!deleting}
      />
    </section>
  )
}
