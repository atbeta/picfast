import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'

import { deleteImage, listImages } from '../../lib/console-api'
import { formatFileSize } from '../../lib/upload'

export function ImagesPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [page, setPage] = useState(1)
  const pageSize = 20

  const { data, isLoading, error } = useQuery({
    queryKey: ['images', page],
    queryFn: () => listImages(page, pageSize),
  })

  const [deleting, setDeleting] = useState<string | null>(null)

  const onDelete = useCallback(
    async (key: string) => {
      if (!window.confirm(t('images.confirmDelete'))) return
      setDeleting(key)
      try {
        await deleteImage(key)
        await qc.invalidateQueries({ queryKey: ['images'] })
      } catch (err: unknown) {
        alert(
          (err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
          t('images.deleteFailed'),
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
      <h1 className="text-xl font-semibold">{t('page.images.title')}</h1>

      {isLoading && (
        <div className="flex justify-center py-12">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-400 border-t-transparent" />
        </div>
      )}

      {error && (
        <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-400">
          {t('images.loadFailed')}
        </p>
      )}

      {data && data.items.length === 0 && (
        <p className="py-12 text-center text-sm text-zinc-400">{t('images.empty')}</p>
      )}

      {data && data.items.length > 0 && (
        <>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-zinc-200 text-left text-xs text-zinc-500 dark:border-zinc-700">
                  <th className="pb-2 pr-3 font-medium">{t('images.colPreview')}</th>
                  <th className="pb-2 pr-3 font-medium">{t('images.colName')}</th>
                  <th className="pb-2 pr-3 font-medium">{t('images.colSize')}</th>
                  <th className="pb-2 pr-3 font-medium">{t('images.colDate')}</th>
                  <th className="pb-2 font-medium">{t('images.colActions')}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-100 dark:divide-zinc-800">
                {data.items.map((img) => (
                  <tr key={img.id}>
                    <td className="py-2 pr-3">
                      {img.thumbnail_url ? (
                        <img src={img.thumbnail_url} alt="" className="h-10 w-10 rounded border border-zinc-200 object-cover dark:border-zinc-700" />
                      ) : (
                        <div className="flex h-10 w-10 items-center justify-center rounded border border-zinc-200 text-xs text-zinc-400 dark:border-zinc-700">
                          {img.extension.toUpperCase()}
                        </div>
                      )}
                    </td>
                    <td className="max-w-[200px] truncate py-2 pr-3">{img.origin_name}</td>
                    <td className="whitespace-nowrap py-2 pr-3 text-zinc-500">{formatFileSize(img.size_bytes)}</td>
                    <td className="whitespace-nowrap py-2 pr-3 text-zinc-500">
                      {new Date(img.created_at).toLocaleDateString()}
                    </td>
                    <td className="py-2">
                      <div className="flex gap-1">
                        <button
                          type="button"
                          onClick={() => onCopy(img.links.url)}
                          className="rounded px-2 py-1 text-xs hover:bg-zinc-100 dark:hover:bg-zinc-800"
                        >
                          {t('images.copyUrl')}
                        </button>
                        <button
                          type="button"
                          onClick={() => onCopy(img.links.markdown)}
                          className="rounded px-2 py-1 text-xs hover:bg-zinc-100 dark:hover:bg-zinc-800"
                        >
                          {t('images.copyMd')}
                        </button>
                        <button
                          type="button"
                          onClick={() => onDelete(img.key)}
                          disabled={deleting === img.key}
                          className="rounded px-2 py-1 text-xs text-red-500 hover:bg-red-50 disabled:opacity-50 dark:hover:bg-red-900/20"
                        >
                          {t('images.delete')}
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {data.total > pageSize && (
            <div className="flex items-center justify-between pt-2">
              <span className="text-xs text-zinc-500">
                {t('images.pagination', { total: data.total })}
              </span>
              <div className="flex gap-2">
                <button
                  type="button"
                  disabled={page <= 1}
                  onClick={() => setPage((p) => p - 1)}
                  className="rounded border border-zinc-300 px-3 py-1 text-xs disabled:opacity-40 dark:border-zinc-700"
                >
                  {t('images.prev')}
                </button>
                <button
                  type="button"
                  disabled={page * pageSize >= data.total}
                  onClick={() => setPage((p) => p + 1)}
                  className="rounded border border-zinc-300 px-3 py-1 text-xs disabled:opacity-40 dark:border-zinc-700"
                >
                  {t('images.next')}
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </section>
  )
}
