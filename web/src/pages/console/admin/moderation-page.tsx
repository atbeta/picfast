import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { ChevronLeft, ChevronRight, Check, X, ShieldAlert } from 'lucide-react'

import {
  approveModerationImage,
  listPendingModerationImages,
  rejectModerationImage,
} from '../../../lib/admin-api'
import { extractErrorMessage } from '../../../lib/error-handler'
import { formatFileSize } from '../../../lib/upload'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { EmptyState, LoadingState } from '@/components/page-states'

export function AdminModerationPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()

  const [page, setPage] = useState(1)
  const pageSize = 20
  const [processingId, setProcessingId] = useState<number | null>(null)
  const [approveTarget, setApproveTarget] = useState<number | null>(null)
  const [rejectTarget, setRejectTarget] = useState<number | null>(null)

  const { data, isLoading, error } = useQuery({
    queryKey: ['admin-moderation-pending', page],
    queryFn: () => listPendingModerationImages({ page, page_size: pageSize }),
  })

  const totalPages = data ? (data.total_pages > 0 ? data.total_pages : Math.max(1, Math.ceil(data.total / pageSize))) : 1

  const refreshAfterAction = useCallback(async () => {
    await Promise.all([
      qc.invalidateQueries({ queryKey: ['admin-moderation-pending'] }),
      qc.invalidateQueries({ queryKey: ['admin-observability-summary'] }),
    ])
  }, [qc])

  const onApprove = useCallback(async () => {
    if (approveTarget === null) return
    setProcessingId(approveTarget)
    try {
      await approveModerationImage(approveTarget)
      toast.success(t('admin.moderationApproveSuccess', { defaultValue: '已通过审核' }))
      setApproveTarget(null)
      await refreshAfterAction()
    } catch (err: unknown) {
      toast.error(extractErrorMessage(err, t('admin.moderationApproveFailed', { defaultValue: '审核通过失败' })))
    } finally {
      setProcessingId(null)
    }
  }, [approveTarget, refreshAfterAction, t])

  const onReject = useCallback(async () => {
    if (rejectTarget === null) return
    setProcessingId(rejectTarget)
    try {
      await rejectModerationImage(rejectTarget)
      toast.success(t('admin.moderationRejectSuccess', { defaultValue: '已拒绝图片' }))
      setRejectTarget(null)
      await refreshAfterAction()
    } catch (err: unknown) {
      toast.error(extractErrorMessage(err, t('admin.moderationRejectFailed', { defaultValue: '拒绝失败' })))
    } finally {
      setProcessingId(null)
    }
  }, [rejectTarget, refreshAfterAction, t])

  return (
    <section className="space-y-6">
      <div className="space-y-1">
        <h1 className="text-2xl font-bold tracking-tight">{t('admin.moderationTitle', { defaultValue: '审核管理' })}</h1>
        <p className="text-sm text-muted-foreground">{t('admin.moderationSubtitle', { defaultValue: '集中处理待人工审核的图片。' })}</p>
      </div>

      {isLoading && <LoadingState />}
      {error && <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">{t('admin.loadFailed')}</p>}

      {data && data.items.length === 0 && (
        <EmptyState
          icon={<ShieldAlert className="size-6 text-muted-foreground" />}
          title={t('admin.empty')}
          description={t('admin.moderationEmptyDesc', { defaultValue: '当前没有待审核图片。' })}
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
                      {img.thumbnail_url ? (
                        <img src={img.thumbnail_url} alt="" className="h-10 w-10 rounded border border-border/50 object-cover" />
                      ) : (
                        <div className="flex h-10 w-10 items-center justify-center rounded border border-border/50 text-xs text-muted-foreground bg-muted/30">
                          {img.extension.toUpperCase()}
                        </div>
                      )}
                    </td>
                    <td className="max-w-[120px] truncate px-3 py-3 font-mono text-xs text-muted-foreground">{img.key}</td>
                    <td className="max-w-[180px] truncate px-3 py-3 text-foreground">{img.origin_name}</td>
                    <td className="whitespace-nowrap px-3 py-3 text-muted-foreground">{formatFileSize(img.size_bytes)}</td>
                    <td className="px-3 py-3">
                      <span className={['rounded px-1.5 py-0.5 text-xs font-medium', img.permission === 1 ? 'bg-primary/10 text-primary' : 'bg-warning/10 text-warning'].join(' ')}>
                        {img.permission === 1 ? t('images.public', { defaultValue: '公开' }) : t('images.private', { defaultValue: '私有' })}
                      </span>
                    </td>
                    <td className="whitespace-nowrap px-3 py-3 text-muted-foreground">{new Date(img.created_at).toLocaleDateString()}</td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex items-center justify-end gap-2">
                        <button
                          type="button"
                          onClick={() => setApproveTarget(img.id)}
                          disabled={processingId === img.id}
                          className="inline-flex h-8 items-center gap-1 rounded-md border border-success/30 bg-success/10 px-2.5 text-xs font-medium text-success hover:bg-success/20 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
                        >
                          <Check className="size-3.5" />
                          {t('admin.moderationApprove', { defaultValue: '通过' })}
                        </button>
                        <button
                          type="button"
                          onClick={() => setRejectTarget(img.id)}
                          disabled={processingId === img.id}
                          className="inline-flex h-8 items-center gap-1 rounded-md border border-destructive/30 bg-destructive/10 px-2.5 text-xs font-medium text-destructive hover:bg-destructive/20 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
                        >
                          <X className="size-3.5" />
                          {t('admin.moderationReject', { defaultValue: '拒绝' })}
                        </button>
                      </div>
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
                <button
                  type="button"
                  disabled={page <= 1}
                  onClick={() => setPage((p) => p - 1)}
                  title={t('admin.prev')}
                  className="inline-flex h-8 w-8 items-center justify-center rounded-lg border border-input bg-background shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
                >
                  <ChevronLeft className="size-4" />
                </button>
                <span className="inline-flex h-8 min-w-[56px] items-center justify-center rounded-lg border border-input bg-background px-2 text-xs text-muted-foreground">
                  {page} / {totalPages}
                </span>
                <button
                  type="button"
                  disabled={page >= totalPages}
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
        open={approveTarget !== null}
        onOpenChange={(open) => { if (!open) setApproveTarget(null) }}
        title={t('admin.moderationApproveConfirm', { defaultValue: '确认通过这张图片吗？' })}
        description={t('admin.moderationApproveDescription', { defaultValue: '通过后图片将对外按权限规则可见。' })}
        confirmLabel={t('admin.moderationApprove', { defaultValue: '通过' })}
        onConfirm={onApprove}
        loading={processingId !== null}
      />

      <ConfirmDialog
        open={rejectTarget !== null}
        onOpenChange={(open) => { if (!open) setRejectTarget(null) }}
        variant="destructive"
        title={t('admin.moderationRejectConfirm', { defaultValue: '确认拒绝这张图片吗？' })}
        description={t('admin.moderationRejectDescription', { defaultValue: '拒绝后图片将保持不可公开状态。' })}
        confirmLabel={t('admin.moderationReject', { defaultValue: '拒绝' })}
        onConfirm={onReject}
        loading={processingId !== null}
      />
    </section>
  )
}
