import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Trash2, ChevronLeft, ChevronRight, Image as ImageIcon } from 'lucide-react'

import { deleteAdminImage, listAdminImages } from '../../../lib/admin-api'
import { formatFileSize } from '../../../lib/upload'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { EmptyState, LoadingState } from '@/components/page-states'

export function AdminImagesPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [emailFilter, setEmailFilter] = useState('')
  const [searchType, setSearchType] = useState<'name' | 'email'>('name')

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
      <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4">
        <h1 className="text-2xl font-bold tracking-tight">{t('admin.imagesTitle', { defaultValue: '图片管理' })}</h1>
        
        <div className="flex items-center rounded-lg border border-input bg-background shadow-sm focus-within:ring-1 focus-within:ring-primary/20 focus-within:border-primary transition-all overflow-hidden h-9">
          <select 
            className="h-full bg-muted/30 px-3 py-1 text-sm outline-none border-r border-input cursor-pointer"
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
          description={t('admin.imagesEmptyDesc', { defaultValue: '站点产生图片后，这里会汇总显示所有上传内容。' })}
        />
      )}

      {data && data.items.length > 0 && (
        <>
          <div className="overflow-x-auto rounded-lg border border-border/50 bg-card shadow-sm">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border/50 text-left text-xs text-muted-foreground bg-muted/30">
                  <th className="pb-2 pr-3 pl-4 pt-2 font-medium">{t('admin.colPreview')}</th>
                  <th className="pb-2 pr-3 pt-2 font-medium">Key</th>
                  <th className="pb-2 pr-3 pt-2 font-medium">{t('admin.colName')}</th>
                  <th className="pb-2 pr-3 pt-2 font-medium">{t('admin.colUploader')}</th>
                  <th className="pb-2 pr-3 pt-2 font-medium">{t('admin.colSize')}</th>
                  <th className="pb-2 pr-3 pt-2 font-medium">{t('images.permission', { defaultValue: '权限' })}</th>
                  <th className="pb-2 pr-3 pt-2 font-medium">{t('admin.colDate')}</th>
                  <th className="pb-2 pr-4 pt-2 font-medium text-right">{t('admin.colActions')}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/50">
                {data.items.map((img) => (
                  <tr key={img.id} className="group hover:bg-muted/50 transition-colors">
                    <td className="py-2 pr-3 pl-4">
                      {img.thumbnail_url ? (
                        <img src={img.thumbnail_url} alt="" className="h-10 w-10 rounded border border-border/50 object-cover" />
                      ) : (
                        <div className="flex h-10 w-10 items-center justify-center rounded border border-border/50 text-xs text-muted-foreground bg-muted/30">{img.extension.toUpperCase()}</div>
                      )}
                    </td>
                    <td className="max-w-[120px] truncate py-2 pr-3 font-mono text-xs text-muted-foreground">{img.key}</td>
                    <td className="max-w-[140px] truncate py-2 pr-3 text-foreground">{img.origin_name}</td>
                    <td className="py-2 pr-3 text-muted-foreground">{img.user_email ?? '—'}</td>
                    <td className="whitespace-nowrap py-2 pr-3 text-muted-foreground">{formatFileSize(img.size_bytes)}</td>
                    <td className="py-2 pr-3">
                      <span className={['rounded px-1.5 py-0.5 text-xs font-medium', img.permission === 1 ? 'bg-primary/10 text-primary' : 'bg-warning/10 text-warning'].join(' ')}>
                        {img.permission === 1 ? (t('images.public', { defaultValue: '公开' })) : (t('images.private', { defaultValue: '私有' }))}
                      </span>
                    </td>
                    <td className="whitespace-nowrap py-2 pr-3 text-muted-foreground">{new Date(img.created_at).toLocaleDateString()}</td>
                    <td className="py-2 pr-4 text-right">
                      <button 
                        type="button" 
                        onClick={() => setDeleteTarget(img.id)} 
                        disabled={deleting === img.id} 
                        className="transition-opacity rounded-lg p-1.5 text-destructive/70 hover:bg-destructive/10 hover:text-destructive disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
                        title={t('admin.delete')}
                      >
                        <Trash2 className="size-4" />
                      </button>
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
        title={t('admin.confirmDeleteImage')}
        description={t('admin.deleteImageDescription')}
        confirmLabel={t('admin.delete')}
        onConfirm={onDelete}
        loading={!!deleting}
      />
    </section>
  )
}
