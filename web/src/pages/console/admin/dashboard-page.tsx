import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'

import {
  getAdminObservabilitySummary,
  getAdminSettings,
  listAdminUsers,
  listAdminImages,
  listAdminGroups,
  listAdminStrategies,
  type AdminHealthItem,
} from '../../../lib/admin-api'

function formatBytes(bytes?: number): string {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let value = bytes
  let unit = 0
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024
    unit += 1
  }
  return `${value.toFixed(value >= 10 || unit === 0 ? 0 : 1)} ${units[unit]}`
}

function formatDuration(seconds?: number): string {
  if (!seconds) return '0m'
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  if (days > 0) return `${days}d ${hours}h`
  if (hours > 0) return `${hours}h ${minutes}m`
  return `${minutes}m`
}

function HealthBadge({ item }: { item?: AdminHealthItem }) {
  const { t } = useTranslation()
  if (item?.status === 'disabled') {
    return (
      <span className="inline-flex rounded-full bg-muted px-2.5 py-1 text-xs font-medium text-muted-foreground">
        {t('admin.statusNotEnabled', { defaultValue: '未启用' })}
      </span>
    )
  }
  const healthy = item?.healthy === true
  return (
    <span className={`inline-flex rounded-full px-2.5 py-1 text-xs font-medium ${healthy ? 'bg-success text-white' : 'bg-destructive/10 text-destructive'}`}>
      {healthy ? t('admin.statusHealthy', { defaultValue: '正常' }) : t('admin.statusUnhealthy', { defaultValue: '异常' })}
    </span>
  )
}

export function AdminDashboardPage() {
  const { t } = useTranslation()

  const { data: usersData } = useQuery({ queryKey: ['admin-users-stats'], queryFn: () => listAdminUsers({ page: 1, page_size: 1 }) })
  const { data: imagesData } = useQuery({ queryKey: ['admin-images-stats'], queryFn: () => listAdminImages({ page: 1, page_size: 1 }) })
  const { data: groups } = useQuery({ queryKey: ['admin-groups-stats'], queryFn: listAdminGroups })
  const { data: strategies } = useQuery({ queryKey: ['admin-strategies-stats'], queryFn: listAdminStrategies })
  const { data: settings } = useQuery({ queryKey: ['admin-dashboard-settings'], queryFn: getAdminSettings })
  const { data: observability } = useQuery({
    queryKey: ['admin-observability-summary'],
    queryFn: getAdminObservabilitySummary,
    refetchInterval: 30_000,
  })

  const stats = [
    { label: t('admin.statUsers', { defaultValue: '用户数' }), value: usersData?.total ?? 0 },
    { label: t('admin.statImages', { defaultValue: '图片数' }), value: imagesData?.total ?? 0 },
    { label: t('admin.statGroups', { defaultValue: '分组数' }), value: Array.isArray(groups) ? groups.length : 0 },
    { label: t('admin.statStrategies', { defaultValue: '存储策略' }), value: Array.isArray(strategies) ? strategies.length : 0 },
  ]

  const siteModes = settings
    ? [
        {
          label: t('admin.allowGuestUpload'),
          value: settings.allow_guest_upload ? t('admin.statusEnabled', { defaultValue: '已开启' }) : t('admin.statusDisabled', { defaultValue: '已关闭' }),
          tone: settings.allow_guest_upload ? 'text-white border-transparent bg-success' : 'text-muted-foreground border-transparent bg-muted',
        },
        {
          label: t('admin.allowRegistration'),
          value: settings.allow_registration ? t('admin.statusEnabled', { defaultValue: '已开启' }) : t('admin.statusDisabled', { defaultValue: '已关闭' }),
          tone: settings.allow_registration ? 'text-white border-transparent bg-success' : 'text-muted-foreground border-transparent bg-muted',
        },
        {
          label: t('admin.requireEmailVerification'),
          value: settings.require_email_verification ? t('admin.statusEnabled', { defaultValue: '已开启' }) : t('admin.statusDisabled', { defaultValue: '已关闭' }),
          tone: settings.require_email_verification ? 'text-white border-transparent bg-success' : 'text-muted-foreground border-transparent bg-muted',
        },
        {
          label: t('admin.moderationMode'),
          value:
            settings.moderation_mode === 'manual'
              ? t('admin.modManual')
              : settings.moderation_mode === 'auto'
                ? t('admin.modAuto')
                : t('admin.modDisabled'),
          tone:
            settings.moderation_mode === 'manual'
              ? 'text-white border-transparent bg-amber-500'
              : settings.moderation_mode === 'auto'
                ? 'text-white border-transparent bg-info'
                : 'text-muted-foreground border-transparent bg-muted',
        },
      ]
    : []

  return (
    <section className="space-y-8">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <h1 className="text-2xl md:text-3xl font-bold tracking-tight">{t('admin.dashboardTitle', { defaultValue: '概览' })}</h1>
      </div>

      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        {stats.map((s) => (
          <div key={s.label} className="group flex flex-col justify-between overflow-hidden rounded-xl border border-border/40 bg-card/60 p-5 shadow-sm transition-all duration-200 hover:border-primary/40 hover:bg-card/80">
            <div className="text-sm font-medium text-muted-foreground">{s.label}</div>
            <div className="mt-3 text-4xl font-bold tracking-tight text-foreground">{s.value}</div>
          </div>
        ))}
      </div>

      {siteModes.length > 0 && (
        <div className="rounded-xl border border-border/40 bg-card/40 overflow-hidden shadow-sm">
          <div className="border-b border-border/40 bg-muted/20 px-6 py-4">
            <h2 className="text-lg font-semibold tracking-tight text-foreground">{t('admin.dashboardSiteState', { defaultValue: '站点状态' })}</h2>
            <p className="mt-1 text-sm text-muted-foreground">{t('admin.dashboardSiteStateDesc', { defaultValue: '快速确认当前注册、游客上传、邮箱验证与审核策略。' })}</p>
          </div>
          <div className="grid gap-px bg-border/40 sm:grid-cols-2 xl:grid-cols-4">
            {siteModes.map((item) => (
              <div key={item.label} className="bg-card/60 p-6 flex flex-col justify-between">
                <div className="text-sm font-medium text-muted-foreground">{item.label}</div>
                <div className="mt-4 flex items-center">
                  <span className={`inline-flex items-center rounded-md px-2.5 py-1 text-xs font-medium ${item.tone}`}>
                    {item.value}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {observability && (
        <div className="rounded-xl border border-border/40 bg-card/40 overflow-hidden shadow-sm">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between border-b border-border/40 bg-muted/20 px-6 py-4">
            <div>
              <h2 className="text-lg font-semibold tracking-tight text-foreground">{t('admin.observabilityTitle', { defaultValue: '系统状态' })}</h2>
              <p className="mt-1 text-sm text-muted-foreground">
                {t('admin.observabilityDesc', { defaultValue: '轻量展示服务健康、运行时和最近使用情况，详细指标仍建议接入 Prometheus / Grafana。' })}
              </p>
            </div>
            <div className="text-[11px] font-medium tracking-wider uppercase text-muted-foreground bg-muted/50 px-2.5 py-1 rounded-md">
              {t('admin.observabilityUpdatedAt', { defaultValue: '更新于' })} {new Date(observability.generated_at).toLocaleTimeString()}
            </div>
          </div>

          <div className="grid gap-px bg-border/40 lg:grid-cols-3">
            <div className="bg-card/60 p-6">
              <div className="mb-5 text-sm font-semibold text-foreground flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-primary/70"></span>
                {t('admin.observabilityHealth', { defaultValue: '健康检查' })}
              </div>
              <div className="space-y-4 text-sm">
                <div className="flex items-center justify-between gap-3">
                  <span className="text-muted-foreground">{t('admin.observabilityDatabase', { defaultValue: '数据库' })}</span>
                  <HealthBadge item={observability.health.database} />
                </div>
                <div className="flex items-center justify-between gap-3">
                  <span className="text-muted-foreground">{t('admin.observabilityUploads', { defaultValue: '上传目录' })}</span>
                  <HealthBadge item={observability.health.uploads} />
                </div>
                <div className="flex items-center justify-between gap-3">
                  <span className="text-muted-foreground">{t('admin.observabilityThumbnails', { defaultValue: '缩略图目录' })}</span>
                  <HealthBadge item={observability.health.thumbnails} />
                </div>
                <div className="flex items-center justify-between gap-3">
                  <span className="text-muted-foreground">{t('admin.observabilityMail', { defaultValue: '邮件服务' })}</span>
                  <HealthBadge item={observability.health.mail} />
                </div>
              </div>
            </div>

            <div className="bg-card/60 p-6">
              <div className="mb-5 text-sm font-semibold text-foreground flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-primary/70"></span>
                {t('admin.observabilityRuntime', { defaultValue: '运行时' })}
              </div>
              <dl className="grid grid-cols-2 gap-x-4 gap-y-4 text-sm">
                <div>
                  <dt className="text-xs text-muted-foreground mb-1">{t('admin.observabilityUptime', { defaultValue: '运行时间' })}</dt>
                  <dd className="font-medium text-foreground">{formatDuration(observability.uptime_seconds)}</dd>
                </div>
                <div>
                  <dt className="text-xs text-muted-foreground mb-1">{t('admin.observabilitySystem', { defaultValue: '系统' })}</dt>
                  <dd className="font-medium text-foreground">{observability.runtime.goos}/{observability.runtime.goarch}</dd>
                </div>
                <div>
                  <dt className="text-xs text-muted-foreground mb-1">{t('admin.observabilityCpu', { defaultValue: 'CPU' })}</dt>
                  <dd className="font-medium text-foreground">{observability.runtime.num_cpu}</dd>
                </div>
                <div>
                  <dt className="text-xs text-muted-foreground mb-1">{t('admin.observabilityGoroutines', { defaultValue: 'Goroutines' })}</dt>
                  <dd className="font-medium text-foreground">{observability.runtime.goroutines}</dd>
                </div>
                <div>
                  <dt className="text-xs text-muted-foreground mb-1">{t('admin.observabilityMemory', { defaultValue: '进程内存' })}</dt>
                  <dd className="font-medium text-foreground">{formatBytes(observability.runtime.memory_alloc_bytes)}</dd>
                </div>
                <div>
                  <dt className="text-xs text-muted-foreground mb-1">Go</dt>
                  <dd className="font-medium text-foreground">{observability.runtime.go_version}</dd>
                </div>
              </dl>
            </div>

            <div className="bg-card/60 p-6">
              <div className="mb-5 text-sm font-semibold text-foreground flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-primary/70"></span>
                {t('admin.observabilityUsage', { defaultValue: '最近使用' })}
              </div>
              <dl className="grid grid-cols-2 gap-x-4 gap-y-4 text-sm">
                <div>
                  <dt className="text-xs text-muted-foreground mb-1">{t('admin.observabilityStorageBytes', { defaultValue: '存储用量' })}</dt>
                  <dd className="font-medium text-foreground">{formatBytes(observability.usage.storage_bytes)}</dd>
                </div>
                <div>
                  <dt className="text-xs text-muted-foreground mb-1">{t('admin.observabilityUploads24h', { defaultValue: '24h 上传' })}</dt>
                  <dd className="font-medium text-foreground">{observability.usage.uploads_24h}</dd>
                </div>
                <div>
                  <dt className="text-xs text-muted-foreground mb-1">{t('admin.observabilityPendingModeration', { defaultValue: '待审核' })}</dt>
                  <dd className="font-medium text-foreground">
                    <Link to="/console/admin/moderation" className="hover:text-primary transition-colors">
                      {observability.usage.pending_moderation}
                    </Link>
                  </dd>
                </div>
                <div>
                  <dt className="text-xs text-muted-foreground mb-1">{t('admin.observabilityAudit24h', { defaultValue: '24h 审计' })}</dt>
                  <dd className="font-medium text-foreground">{observability.usage.audit_logs_24h}</dd>
                </div>
                <div>
                  <dt className="text-xs text-muted-foreground mb-1">{t('admin.observabilityDbPool', { defaultValue: 'DB 连接' })}</dt>
                  <dd className="font-medium text-foreground">{observability.database.acquired_connections}/{observability.database.max_connections}</dd>
                </div>
                <div>
                  <dt className="text-xs text-muted-foreground mb-1">{t('admin.observabilityPprof', { defaultValue: 'pprof' })}</dt>
                  <dd className="font-medium text-foreground">{observability.config.pprof_enabled ? t('admin.statusEnabled', { defaultValue: '已开启' }) : t('admin.statusDisabled', { defaultValue: '已关闭' })}</dd>
                </div>
              </dl>
            </div>
          </div>

          <div className="border-t border-border/40 bg-card/60 p-6">
            <div className="mb-4 text-sm font-semibold text-foreground flex items-center gap-2">
              <span className="w-1.5 h-1.5 rounded-full bg-primary/70"></span>
              {t('admin.observabilityStorageStrategies', { defaultValue: '存储策略健康' })}
            </div>
            <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
              {observability.storage_strategies.map((strategy) => (
                <div key={strategy.id} className="rounded-lg border border-border/40 bg-card/40 p-4 transition-colors hover:border-primary/30">
                  <div className="flex items-center justify-between gap-3">
                    <div className="min-w-0">
                      <div className="truncate text-sm font-medium text-foreground">{strategy.name}</div>
                      <div className="mt-0.5 text-xs text-muted-foreground font-mono">{strategy.type}</div>
                    </div>
                    <span className={`shrink-0 rounded-md px-2 py-0.5 text-[11px] font-medium uppercase tracking-wide ${strategy.healthy ? 'bg-success/15 text-success border border-success/20' : 'bg-destructive/15 text-destructive border border-destructive/20'}`}>
                      {strategy.healthy
                        ? strategy.warning
                          ? t('admin.statusLimited', { defaultValue: '受限' })
                          : t('admin.statusHealthy', { defaultValue: '健康' })
                        : t('admin.statusUnhealthy', { defaultValue: '异常' })}
                    </span>
                  </div>
                  {strategy.warning && <div className="mt-3 truncate text-xs text-amber-500/90" title={strategy.warning}>{strategy.warning}</div>}
                  {strategy.error && <div className="mt-3 truncate text-xs text-destructive/90" title={strategy.error}>{strategy.error}</div>}
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </section>
  )
}
