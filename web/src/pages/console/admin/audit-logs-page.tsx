import { Fragment, useMemo, useState } from 'react'
import type { TFunction } from 'i18next'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'

import { listAdminAuditLogs } from '../../../lib/admin-api'
import { Button } from '@/components/ui/button'
import { LoadingState } from '@/components/page-states'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

function auditActionLabel(action: string, t: TFunction): string {
  if (!action) return ''
  const key = `admin.audit_action_${action.replace(/\./g, '_')}`
  return t(key, { defaultValue: action })
}

const AUDIT_ACTION_FILTER_CODES = [
  '',
  'image.upload',
  'image.delete',
  'image.update',
  'image.batch_delete',
  'api_token.create',
  'api_token.delete',
  'admin.auth.login.success',
  'admin.auth.login.failed',
  'admin.settings.update',
  'admin.image.delete',
  'admin.user.update',
  'admin.user.delete',
  'admin.group.create',
  'admin.group.update',
  'admin.group.delete',
  'admin.group.set_strategies',
  'admin.strategy.create',
  'admin.strategy.update',
  'admin.strategy.delete',
  'admin.moderation.approve',
  'admin.moderation.reject',
  'album.create',
  'album.update',
  'album.delete',
  'oauth.login',
  'oauth.link',
  'oauth.unlink',
  'image.tag_add',
  'image.tag_remove',
  'tag.create',
  'webhook.create',
  'webhook.update',
  'webhook.delete',
  'user.register',
  'user.password_change',
  'user.password_reset',
] as const

const AUDIT_RESOURCE_FILTER_CODES = [
  '',
  'image',
  'user',
  'group',
  'strategy',
  'api_token',
  'auth',
  'site_settings',
  'user_identity',
  'album',
  'tag',
  'webhook',
] as const

export function AdminAuditLogsPage() {
  const { t } = useTranslation()
  const [page, setPage] = useState(1)
  const [action, setAction] = useState('')
  const [resourceType, setResourceType] = useState('')
  const [expandedId, setExpandedId] = useState<number | null>(null)
  const pageSize = 20

  const actionSelectItems = useMemo(
    () => ({
      '': t('admin.auditActionAll', { defaultValue: '全部操作' }),
      ...Object.fromEntries(
        AUDIT_ACTION_FILTER_CODES.filter(Boolean).map((a) => [a, auditActionLabel(a, t)] as const),
      ),
    }),
    [t],
  )

  const resourceSelectItems = useMemo(
    () => ({
      '': t('admin.auditResourceAll', { defaultValue: '全部资源' }),
      image: t('admin.auditResourceImage', { defaultValue: '图片' }),
      user: t('admin.auditResourceUser', { defaultValue: '用户' }),
      group: t('admin.auditResourceGroup', { defaultValue: '分组' }),
      strategy: t('admin.auditResourceStrategy', { defaultValue: '存储策略' }),
      api_token: t('admin.auditResourceApiToken', { defaultValue: 'API 令牌' }),
      auth: t('admin.auditResourceAuth', { defaultValue: '认证' }),
      site_settings: t('admin.auditResourceSiteSettings', { defaultValue: '站点设置' }),
      user_identity: t('admin.auditResourceUserIdentity', { defaultValue: '用户身份' }),
      album: t('admin.auditResourceAlbum', { defaultValue: '相册' }),
      tag: t('admin.auditResourceTag', { defaultValue: '标签' }),
      webhook: t('admin.auditResourceWebhook', { defaultValue: 'Webhook' }),
    }),
    [t],
  )

  const { data, isLoading, error } = useQuery({
    queryKey: ['admin-audit-logs', page, action, resourceType],
    queryFn: () => listAdminAuditLogs({ page, page_size: pageSize, action, resource_type: resourceType }),
  })
  const totalPages = data ? (data.total_pages > 0 ? data.total_pages : Math.max(1, Math.ceil(data.total / pageSize))) : 1

  return (
    <section className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight">{t('admin.auditLogsTitle', { defaultValue: '审计日志' })}</h1>
          <p className="text-sm text-muted-foreground">{t('admin.auditLogsSubtitle', { defaultValue: '记录关键管理操作，便于审计追踪与问题回溯。' })}</p>
        </div>
      </div>

      <div className="rounded-xl border border-border/50 bg-card p-4">
        <div className="flex flex-wrap items-center gap-3">
          <div className="min-w-[220px] flex-1 sm:max-w-[320px]">
            <div className="flex items-center gap-2">
              <label className="shrink-0 text-xs font-medium text-muted-foreground">
              {t('admin.auditAction', { defaultValue: '操作' })}
              </label>
              <Select
                value={action}
                items={actionSelectItems}
                onValueChange={(val) => {
                  setPage(1)
                  setAction(String(val))
                }}
              >
                <SelectTrigger className="h-10 min-w-0 flex-1">
                  <SelectValue placeholder={t('admin.auditActionAll', { defaultValue: '全部操作' })} />
                </SelectTrigger>
                <SelectContent>
                  {AUDIT_ACTION_FILTER_CODES.map((item) => (
                    <SelectItem key={item || 'all'} value={item}>
                      {item ? auditActionLabel(item, t) : t('admin.auditActionAll', { defaultValue: '全部操作' })}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="min-w-[220px] flex-1 sm:max-w-[320px]">
            <div className="flex items-center gap-2">
              <label className="shrink-0 text-xs font-medium text-muted-foreground">
                {t('admin.auditResource', { defaultValue: '资源' })}
              </label>
              <Select
                value={resourceType}
                items={resourceSelectItems}
                onValueChange={(val) => {
                  setPage(1)
                  setResourceType(String(val))
                }}
              >
                <SelectTrigger className="h-10 min-w-0 flex-1">
                  <SelectValue placeholder={t('admin.auditResourceAll', { defaultValue: '全部资源' })} />
                </SelectTrigger>
                <SelectContent>
                  {AUDIT_RESOURCE_FILTER_CODES.map((item) => (
                    <SelectItem key={item || 'all'} value={item}>
                      {resourceSelectItems[item]}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="ms-auto flex h-10 items-center gap-2">
            {(action || resourceType) ? (
              <>
                <span className="rounded-full border border-primary/30 bg-primary/10 px-2 py-0.5 text-[11px] text-primary">
                  {t('admin.auditFiltered', { defaultValue: '已筛选' })}
                </span>
                <Button
                  variant="link"
                  size="xs"
                  className="text-muted-foreground hover:text-foreground"
                  onClick={() => {
                    setPage(1)
                    setAction('')
                    setResourceType('')
                  }}
                >
                  {t('admin.auditClearFilters', { defaultValue: '清空筛选' })}
                </Button>
              </>
            ) : null}
          </div>
        </div>
      </div>

      {isLoading && <LoadingState />}
      {error && <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">{t('admin.loadFailed')}</p>}

      {data && (
        <div className="space-y-4">
          <div className="space-y-3 md:hidden">
            {data.items.map((item) => (
              <div key={item.id} className="rounded-xl border border-border/50 bg-card/80 p-4 shadow-sm">
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0 space-y-1">
                    <div className="text-xs text-muted-foreground">{new Date(item.created_at).toLocaleString()}</div>
                    <div className="text-sm font-medium text-foreground">
                      {item.actor_email ??
                        (item.details?.guest === true ? t('admin.auditActorGuest', { defaultValue: '游客' }) : '—')}
                    </div>
                  </div>
                  <span className="rounded bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary" title={item.action}>
                    {auditActionLabel(item.action, t)}
                  </span>
                </div>

                <div className="mt-3 grid grid-cols-1 gap-2 text-xs">
                  <div className="rounded-lg bg-muted/20 px-2.5 py-2">
                    <p className="text-muted-foreground">{t('admin.auditResource', { defaultValue: '资源' })}</p>
                    <p className="mt-1 text-foreground">{item.resource_type}</p>
                    <p className="mt-1 truncate text-muted-foreground">{item.resource_name || item.resource_id || '-'}</p>
                  </div>
                  <div className="rounded-lg bg-muted/20 px-2.5 py-2">
                    <p className="text-muted-foreground">IP</p>
                    <p className="mt-1 font-mono text-foreground">{item.ip || '-'}</p>
                  </div>
                </div>

                <Button
                  variant="link"
                  size="xs"
                  className="mt-2 px-0 text-[11px]"
                  onClick={() => setExpandedId((prev) => (prev === item.id ? null : item.id))}
                >
                  {expandedId === item.id
                    ? t('admin.auditHideDetails', { defaultValue: '收起详情' })
                    : t('admin.auditViewDetails', { defaultValue: '查看详情' })}
                </Button>

                {expandedId === item.id && (
                  <pre className="mt-2 max-h-72 overflow-auto rounded-lg border border-border/40 bg-background p-3 text-xs">
                    {JSON.stringify(item.details ?? {}, null, 2)}
                  </pre>
                )}
              </div>
            ))}
            {data.items.length === 0 && (
              <div className="rounded-xl border border-border/50 bg-card/80 px-4 py-10 text-center text-sm text-muted-foreground">
                {t('admin.empty')}
              </div>
            )}
          </div>

          <div className="hidden overflow-x-auto rounded-xl border border-border/50 bg-card/80 md:block">
            <table className="min-w-full text-sm">
              <thead className="bg-muted/40 text-left">
                <tr>
                  <th className="px-4 py-3 font-medium">{t('common.createdAt')}</th>
                  <th className="px-4 py-3 font-medium">{t('admin.auditColActor', { defaultValue: '操作者' })}</th>
                  <th className="px-4 py-3 font-medium">{t('admin.auditAction', { defaultValue: '动作' })}</th>
                  <th className="px-4 py-3 font-medium">{t('admin.auditResource', { defaultValue: '资源' })}</th>
                  <th className="px-4 py-3 font-medium">IP</th>
                </tr>
              </thead>
              <tbody>
                {data.items.map((item) => (
                  <Fragment key={item.id}>
                    <tr className="border-t border-border/40 align-top">
                      <td className="px-4 py-3 text-muted-foreground">{new Date(item.created_at).toLocaleString()}</td>
                      <td className="px-4 py-3">
                        {item.actor_email ??
                          (item.details?.guest === true ? t('admin.auditActorGuest', { defaultValue: '游客' }) : '—')}
                      </td>
                      <td className="px-4 py-3">
                        <span className="rounded bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary" title={item.action}>
                          {auditActionLabel(item.action, t)}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <div className="text-xs text-muted-foreground">{item.resource_type}</div>
                        <div className="font-medium">{item.resource_name || item.resource_id || '-'}</div>
                      </td>
                      <td className="px-4 py-3 font-mono text-xs text-muted-foreground">
                        <div>{item.ip || '-'}</div>
                        <Button
                          variant="link"
                          size="xs"
                          className="mt-1 text-[11px]"
                          onClick={() => setExpandedId((prev) => (prev === item.id ? null : item.id))}
                        >
                          {expandedId === item.id
                            ? t('admin.auditHideDetails', { defaultValue: '收起详情' })
                            : t('admin.auditViewDetails', { defaultValue: '查看详情' })}
                        </Button>
                      </td>
                    </tr>
                    {expandedId === item.id && (
                      <tr className="border-t border-border/30 bg-muted/20">
                        <td colSpan={5} className="px-4 py-3">
                          <pre className="max-h-72 overflow-auto rounded-lg border border-border/40 bg-background p-3 text-xs">
                            {JSON.stringify(item.details ?? {}, null, 2)}
                          </pre>
                        </td>
                      </tr>
                    )}
                  </Fragment>
                ))}
                {data.items.length === 0 && (
                  <tr>
                    <td colSpan={5} className="px-4 py-10 text-center text-sm text-muted-foreground">
                      {t('admin.empty')}
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          {data.total > pageSize && (
            <div className="flex items-center justify-between">
              <span className="text-xs text-muted-foreground">{t('admin.pagination', { total: data.total })}</span>
              <div className="flex gap-2">
                <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
                  {t('admin.prev')}
                </Button>
                <span className="inline-flex h-8 min-w-[56px] items-center justify-center rounded-lg border border-input bg-background px-2 text-xs text-muted-foreground">
                  {page} / {totalPages}
                </span>
                <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
                  {t('admin.next')}
                </Button>
              </div>
            </div>
          )}
        </div>
      )}
    </section>
  )
}
