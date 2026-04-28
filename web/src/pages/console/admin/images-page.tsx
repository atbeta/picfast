import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'

import { deleteAdminImage, listAdminImages } from '../../../lib/admin-api'
import { formatFileSize } from '../../../lib/upload'
import { ConfirmDialog } from '@/components/confirm-dialog'

export function AdminImagesPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [emailFilter, setEmailFilter] = useState('')

  const { data, isLoading, error } = useQuery({
    queryKey: ['admin-images', page, keyword, emailFilter],
    queryFn: () => listAdminImages({ page, page_size: 20, keyword: keyword || undefined, email: emailFilter || undefined }),
  })

  const [deleting, setDeleting] = useState<number | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<number | null>(null)

  const onDelete = useCallback(async () => {
    if (deleteTarget === null) return
    setDeleting(deleteTarget)
    try {
      await deleteAdminImage(deleteTarget)
      setDeleteTarget(null)
      await qc.invalidateQueries({ queryKey: ['admin-images'] })
    } catch (err: unknown) {
      toast.error((err as { response?: { data?: { message?: string } } })?.response?.data?.message || t('admin.deleteFailed'))
    } finally {
      setDeleting(null)
    }
  }, [deleteTarget, qc, t])

  return (
    <section className="space-y-4">
      <h1 className="text-xl font-semibold">{t('admin.imagesTitle')}</h1>

      <div className="flex flex-wrap gap-2">
        <input value={keyword} onChange={(e) => { setKeyword(e.target.value); setPage(1) }} placeholder={t('admin.searchName')} className="w-48 rounded-md border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 dark:border-zinc-700 dark:bg-zinc-800" />
        <input value={emailFilter} onChange={(e) => { setEmailFilter(e.target.value); setPage(1) }} placeholder={t('admin.searchEmail')} className="w-48 rounded-md border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 dark:border-zinc-700 dark:bg-zinc-800" />
      </div>

      {isLoading && <div className="flex justify-center py-12"><div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-400 border-t-transparent" /></div>}
      {error && <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-400">{t('admin.loadFailed')}</p>}
      {data && data.items.length === 0 && <p className="py-12 text-center text-sm text-zinc-400">{t('admin.empty')}</p>}

      {data && data.items.length > 0 && (
        <>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-zinc-200 text-left text-xs text-zinc-500 dark:border-zinc-700">
                  <th className="pb-2 pr-3 font-medium">{t('admin.colPreview')}</th>
                  <th className="pb-2 pr-3 font-medium">Key</th>
                  <th className="pb-2 pr-3 font-medium">{t('admin.colName')}</th>
                  <th className="pb-2 pr-3 font-medium">{t('admin.colUploader')}</th>
                  <th className="pb-2 pr-3 font-medium">{t('admin.colSize')}</th>
                  <th className="pb-2 pr-3 font-medium">{t('images.permission', { defaultValue: '权限' })}</th>
                  <th className="pb-2 pr-3 font-medium">{t('admin.colDate')}</th>
                  <th className="pb-2 font-medium">{t('admin.colActions')}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-100 dark:divide-zinc-800">
                {data.items.map((img) => (
                  <tr key={img.id}>
                    <td className="py-2 pr-3">
                      {img.thumbnail_url ? (
                        <img src={img.thumbnail_url} alt="" className="h-10 w-10 rounded border border-zinc-200 object-cover dark:border-zinc-700" />
                      ) : (
                        <div className="flex h-10 w-10 items-center justify-center rounded border border-zinc-200 text-xs text-zinc-400 dark:border-zinc-700">{img.extension.toUpperCase()}</div>
                      )}
                    </td>
                    <td className="max-w-[120px] truncate py-2 pr-3 font-mono text-xs text-zinc-500">{img.key}</td>
                    <td className="max-w-[140px] truncate py-2 pr-3">{img.origin_name}</td>
                    <td className="py-2 pr-3 text-zinc-500">{img.user_email ?? '—'}</td>
                    <td className="whitespace-nowrap py-2 pr-3 text-zinc-500">{formatFileSize(img.size_bytes)}</td>
                    <td className="py-2 pr-3">
                      <span className={['rounded px-1.5 py-0.5 text-xs', img.permission === 1 ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'].join(' ')}>
                        {img.permission === 1 ? (t('images.public', { defaultValue: '公开' })) : (t('images.private', { defaultValue: '私有' }))}
                      </span>
                    </td>
                    <td className="whitespace-nowrap py-2 pr-3 text-zinc-500">{new Date(img.created_at).toLocaleDateString()}</td>
                    <td className="py-2">
                      <button type="button" onClick={() => setDeleteTarget(img.id)} disabled={deleting === img.id} className="rounded px-2 py-1 text-xs text-red-500 hover:bg-red-50 disabled:opacity-50 dark:hover:bg-red-900/20">{t('admin.delete')}</button>
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
        title={t('admin.confirmDeleteImage')}
        description={t('admin.deleteImageDescription')}
        confirmLabel={t('admin.delete')}
        onConfirm={onDelete}
        loading={!!deleting}
      />
    </section>
  )
}
