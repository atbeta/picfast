import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient, keepPreviousData } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Trash2, ChevronLeft, ChevronRight, Image as ImageIcon } from 'lucide-react'

import { deleteAdminImage, listAdminImages } from '../../../lib/admin-api'
import { extractErrorMessage } from '../../../lib/error-handler'
import { formatFileSize } from '../../../lib/upload'
import { Button } from '@/components/ui/button'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { EmptyState, LoadingState } from '@/components/page-states'

export function AdminImagesPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [page, setPage] = useState(1)
  const pageSize = 20
  const [keyword, setKeyword] = useState('')
  const [emailFilter, setEmailFilter] = useState('')
  const [searchType, setSearchType] = useState<'name' | 'email'>('name')

  const { data, isLoading, error } = useQuery({
    queryKey: ['admin-images', page, keyword, emailFilter],
    queryFn: () => listAdminImages({ page, page_size: pageSize, keyword: keyword || undefined, email: emailFilter || undefined }),
    placeholderData: keepPreviousData,
  })
  const totalPages = data ? (data.total_pages > 0 ? data.total_pages : Math.max(1, Math.ceil(data.total / pageSize))) : 1

  const [deleting, setDeleting] = useState<number | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<number | null>(null)

  const onDelete = useCallback(async () => {
    if (deleteTarget === null) return
    setDeleting(deleteTarget)
    try {
      // If this is the only item on a non-first page, jump back first so the
      // refetched data lands on a page with content instead of an empty one.
      const willEmptyPage = data && data.items.length === 1 && page > 1
      if (willEmptyPage) {
        setPage((p) => p - 1)
      }
      await deleteAdminImage(deleteTarget)
      setDeleteTarget(null)
      await qc.invalidateQueries({ queryKey: ['admin-images'] })
    } catch (err: unknown) {
      toast.error(extractErrorMessage(err, t('admin.deleteFailed')))
    } finally {
      setDeleting(null)
    }
  }, [deleteTarget, data, page, qc, t])

  return (
    <section className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight">{t('admin.imagesTitle', { defaultValue: '图片管理' })}</h1>
          <p className="text-sm text-muted-foreground">{t('admin.imagesSubtitle', { defaultValue: '按文件名或上传者检索全站图片资产。' })}</p>
        </div>
        
        <div className="flex h-10 items-center overflow-hidden rounded-lg border border-input bg-background shadow-sm transition-colors duration-150 focus-within:border-primary focus-within:ring-1 focus-within:ring-primary/20">
          <select 
            className="h-full border-r border-input bg-muted/30 px-3 text-sm outline-none cursor-pointer"
            value={searchType}
            onChange={(e) => {
              setSearchType(e.target.value as 'name' | 'email');
              setKeyword('');
              setEmailFilter('');
              setPage(1);
            }}
          >
            <option value="name">{t('admin.searchName', { defaultValue: '搜文件名' })}</option>
            <option value="email">{t('admin.searchEmail', { defaultValue: '搜用户邮箱' })}</option>
          </select>
          <input 
            value={searchType === 'name' ? keyword : emailFilter} 
            onChange={(e) => { 
              if (searchType === 'name') setKeyword(e.target.value);
              else setEmailFilter(e.target.value);
              setPage(1);
            }} 
            placeholder={searchType === 'name' ? t('admin.searchName') : t('admin.searchEmail')} 
            className="h-full w-48 sm:w-64 bg-transparent px-3 py-1 text-sm outline-none" 
          />
        </div>
      </div>

      {isLoading && <LoadingState />}
      {error && <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">{t('admin.loadFailed')}</p>}
      {data && data.items.length === 0 && (
        <EmptyState
          icon={<ImageIcon className="size-6 text-muted-foreground" />}
          title={t('admin.empty')}
          description={t('admin.imagesEmptyDesc', { defaultValue: '产生上传数据后，此处将汇总显示全站图片内容。' })}
        />
      )}

      {data && data.items.length > 0 && (
        <>
          <div className="overflow-x-auto rounded-xl border border-border/50 bg-card/80 shadow-sm">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border/50 bg-muted/35 text-left text-xs text-muted-foreground">
                    <th className="px-4 py-3 font-medium whitespace-nowrap">{t('admin.colPreview')}</th>
                    <th className="px-3 py-3 font-medium whitespace-nowrap">{t('admin.imageKey')}</th>
                    <th className="px-3 py-3 font-medium whitespace-nowrap">{t('admin.colName')}</th>
                    <th className="px-3 py-3 font-medium whitespace-nowrap">{t('admin.colUploader')}</th>
                    <th className="px-3 py-3 font-medium whitespace-nowrap">{t('admin.colSize')}</th>
                    <th className="px-3 py-3 font-medium whitespace-nowrap">{t('images.permission', { defaultValue: '权限' })}</th>
                    <th className="px-3 py-3 font-medium whitespace-nowrap">{t('admin.colDate')}</th>
                    <th className="px-4 py-3 font-medium text-right whitespace-nowrap">{t('admin.colActions')}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/50">
                {data.items.map((img) => (
                  <tr key={img.id} className="group hover:bg-muted/50 transition-colors">
                    <td className="px-4 py-3">
                      {(img.thumbnail_url || img.extension === 'svg' || img.extension === 'ico') ? (
                        <img src={img.thumbnail_url || img.url} alt="" className="h-10 w-10 rounded border border-border/50 object-cover" />
                      ) : (
                        <div className="flex h-10 w-10 items-center justify-center rounded border border-border/50 text-xs text-muted-foreground bg-muted/30">{img.extension.toUpperCase()}</div>
                      )}
                    </td>
                    <td className="max-w-[120px] truncate px-3 py-3 font-mono text-xs text-muted-foreground">{img.key}</td>
                    <td className="max-w-[140px] truncate px-3 py-3 text-foreground">{img.origin_name}</td>
                    <td className="px-3 py-3 text-muted-foreground">{img.user_email || img.uploaded_ip || '—'}</td>
                    <td className="whitespace-nowrap px-3 py-3 text-muted-foreground">{formatFileSize(img.size_bytes)}</td>
                    <td className="px-3 py-3">
                      <span className={['rounded px-1.5 py-0.5 text-xs font-medium', img.permission === 1 ? 'bg-primary/10 text-primary' : 'bg-warning/10 text-warning'].join(' ')}>
                        {img.permission === 1 ? (t('images.public', { defaultValue: '公开' })) : (t('images.private', { defaultValue: '私有' }))}
                      </span>
                    </td>
                    <td className="whitespace-nowrap px-3 py-3 text-muted-foreground">{new Date(img.created_at).toLocaleDateString()}</td>
                    <td className="px-4 py-3 text-right">
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        onClick={() => setDeleteTarget(img.id)}
                        disabled={deleting === img.id}
                        className="text-destructive/70 hover:text-destructive"
                        title={t('admin.delete')}
                      >
                        <Trash2 className="size-4" />
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {data.total > pageSize && (
            <div className="flex items-center justify-between pt-4">
              <span className="text-xs text-muted-foreground">{t('admin.pagination', { total: data.total })}</span>
              <div className="flex gap-2">
                <Button variant="outline" size="icon" disabled={page <= 1} onClick={() => setPage((p) => p - 1)} title={t('admin.prev')}>
                  <ChevronLeft className="size-4" />
                </Button>
                <span className="inline-flex h-8 min-w-[56px] items-center justify-center rounded-lg border border-input bg-background px-2 text-xs text-muted-foreground">
                  {page} / {totalPages}
                </span>
                <Button variant="outline" size="icon" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)} title={t('admin.next')}>
                  <ChevronRight className="size-4" />
                </Button>
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
