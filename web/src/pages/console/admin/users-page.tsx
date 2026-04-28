import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Ban, Unlock, Trash2, ChevronLeft, ChevronRight } from 'lucide-react'

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
      <h1 className="text-2xl font-bold tracking-tight">{t('admin.usersTitle')}</h1>

      <input
        value={keyword}
        onChange={(e) => { setKeyword(e.target.value); setPage(1) }}
        placeholder={t('admin.searchPlaceholder')}
        className="w-full max-w-xs rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary/20 transition-all"
      />

      {isLoading && (
        <div className="flex justify-center py-12"><div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" /></div>
      )}
      {error && <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">{t('admin.loadFailed')}</p>}
      {data && data.items.length === 0 && <p className="py-12 text-center text-sm text-muted-foreground">{t('admin.empty')}</p>}

      {data && data.items.length > 0 && (
        <>
          <div className="overflow-x-auto rounded-lg border border-border/50 bg-card shadow-sm">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border/50 text-left text-xs text-muted-foreground bg-muted/30">
                  <th className="pb-2 pr-3 pl-4 pt-2 font-medium">ID</th>
                  <th className="pb-2 pr-3 pt-2 font-medium">{t('admin.colEmail')}</th>
                  <th className="pb-2 pr-3 pt-2 font-medium">{t('admin.colName')}</th>
                  <th className="pb-2 pr-3 pt-2 font-medium">{t('admin.colRole')}</th>
                  <th className="pb-2 pr-3 pt-2 font-medium">{t('admin.colStatus')}</th>
                  <th className="pb-2 pr-3 pt-2 font-medium">{t('admin.colImages')}</th>
                  <th className="pb-2 pr-3 pt-2 font-medium">{t('admin.usedCapacity', { defaultValue: '已用容量' })}</th>
                  <th className="pb-2 pr-4 pt-2 font-medium">{t('admin.colActions')}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/50">
                {data.items.map((u) => (
                  <tr key={u.id} className="group hover:bg-muted/20 transition-colors">
                    <td className="py-2 pr-3 pl-4 text-muted-foreground">{u.id}</td>
                    <td className="py-2 pr-3 text-foreground">{u.email}</td>
                    <td className="py-2 pr-3 text-foreground">{u.name}</td>
                    <td className="py-2 pr-3">
                      <span className={['rounded px-1.5 py-0.5 text-xs font-medium', u.role === 'admin' ? 'bg-warning/10 text-warning' : 'bg-muted text-muted-foreground'].join(' ')}>
                        {u.role}
                      </span>
                    </td>
                    <td className="py-2 pr-3">
                      <span className={['rounded px-1.5 py-0.5 text-xs font-medium', u.status === 1 ? 'bg-success/10 text-success' : 'bg-destructive/10 text-destructive'].join(' ')}>
                        {u.status === 1 ? t('admin.active') : t('admin.frozen')}
                      </span>
                    </td>
                    <td className="py-2 pr-3 text-muted-foreground">{u.image_num}</td>
                    <td className="py-2 pr-3 text-muted-foreground">{formatFileSize(u.used_capacity || 0)} / {formatFileSize(u.capacity_bytes)}</td>
                    <td className="py-2 pr-4">
                      <div className="flex gap-1">
                        {u.role !== 'admin' && (
                          <button
                            type="button"
                            onClick={() => toggleStatus(u)}
                            disabled={saving === u.id}
                            title={u.status === 1 ? t('admin.freeze') : t('admin.activate')}
                            className="rounded-lg px-2 py-1.5 text-xs font-medium hover:bg-muted disabled:opacity-50 transition-colors cursor-pointer"
                          >
                            {u.status === 1 ? <Ban className="h-4 w-4" /> : <Unlock className="h-4 w-4" />}
                          </button>
                        )}
                        {u.role !== 'admin' && (
                          <button
                            type="button"
                            onClick={() => setDeleteTarget(u.id)}
                            disabled={deleting === u.id}
                            title={t('admin.delete')}
                            className="opacity-0 group-hover:opacity-100 transition-opacity rounded-lg p-1.5 text-destructive/70 hover:bg-destructive/10 hover:text-destructive disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
                          >
                            <Trash2 className="size-4" />
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
            <div className="flex items-center justify-between pt-4">
              <span className="text-xs text-muted-foreground">{t('admin.pagination', { total: data.total })}</span>
              <div className="flex gap-2">
                <button 
                  type="button" 
                  disabled={page <= 1} 
                  onClick={() => setPage((p) => p - 1)} 
                  title={t('admin.prev')}
                  className="inline-flex h-8 w-8 items-center justify-center rounded-lg border border-input bg-background shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
                >
                  <ChevronLeft className="size-4" />
                </button>
                <button 
                  type="button" 
                  disabled={page * 20 >= data.total} 
                  onClick={() => setPage((p) => p + 1)} 
                  title={t('admin.next')}
                  className="inline-flex h-8 w-8 items-center justify-center rounded-lg border border-input bg-background shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
                >
                  <ChevronRight className="size-4" />
                </button>
              </div>
            </div>
          )}
        </>
      )}

      <ConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}
        title={t('admin.confirmDeleteUser')}
        description={t('admin.deleteUserDescription')}
        confirmLabel={t('admin.delete')}
        onConfirm={onDelete}
        loading={!!deleting}
      />
    </section>
  )
}
