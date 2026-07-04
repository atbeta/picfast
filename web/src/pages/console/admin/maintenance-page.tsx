import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { ShieldCheck, AlertTriangle, Info, CheckCircle2, HardDrive, Clock, XCircle } from 'lucide-react'
import { getMaintenanceSummary, type MaintenanceSummary, type MaintenanceRisk } from '../../../lib/admin-api'
import { LoadingState } from '@/components/page-states'

function formatBytes(bytes: number): string {
  if (!bytes || bytes <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0
  let v = bytes
  while (v >= 1024 && i < units.length - 1) { v /= 1024; i++ }
  return `${v.toFixed(1)} ${units[i]}`
}

function RiskIcon({ level }: { level: string }) {
  const cls = 'size-5 shrink-0'
  switch (level) {
    case 'error': return <XCircle className={`${cls} text-red-500`} />
    case 'warn': return <AlertTriangle className={`${cls} text-yellow-500`} />
    default: return <Info className={`${cls} text-blue-500`} />
  }
}

export function MaintenancePage() {
  const { t } = useTranslation()
  const { data, isLoading, error } = useQuery<MaintenanceSummary>({
    queryKey: ['maintenance-summary'],
    queryFn: getMaintenanceSummary,
    refetchInterval: 30000,
  })

  if (isLoading) return <LoadingState className="min-h-[40vh]" compact />
  if (error || !data) return (
    <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
      <ShieldCheck className="size-10 mb-3 opacity-30" />
      <p className="text-sm">{t('webhooks.loadFailed', { defaultValue: '加载失败' })}</p>
    </div>
  )

  const risks = data.risks ?? []
  const errorCount = risks.filter((r: MaintenanceRisk) => r.level === 'error').length
  const warnCount = risks.filter((r: MaintenanceRisk) => r.level === 'warn').length
  const statusIcon = errorCount > 0 ? <XCircle className="size-8 text-red-500" />
    : warnCount > 0 ? <AlertTriangle className="size-8 text-yellow-500" />
    : <CheckCircle2 className="size-8 text-green-500" />

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{t('admin.maintenanceTitle', { defaultValue: '系统维护' })}</h1>
          <p className="text-sm text-muted-foreground mt-1">{t('admin.maintenanceDesc', { defaultValue: '数据完整性、存储用量、风险巡检' })}</p>
        </div>
        <div className="flex items-center gap-2 rounded-xl border bg-card px-4 py-3">
          {statusIcon}
          <div>
            <p className="text-sm font-semibold">{errorCount > 0 ? `${errorCount} 个错误` : warnCount > 0 ? `${warnCount} 个提醒` : '系统正常'}</p>
            <p className="text-xs text-muted-foreground">{new Date(data.generated_at).toLocaleString()}</p>
          </div>
        </div>
      </div>

      {/* Risks */}
      {risks.length > 0 && (
        <div className="space-y-2">
          <h2 className="flex items-center gap-2 text-sm font-semibold"><AlertTriangle className="size-4" />{t('admin.risks', { defaultValue: '风险与提醒' })}</h2>
          {risks.map((r: MaintenanceRisk, i: number) => (
            <div key={i} className={`flex items-start gap-3 rounded-lg border p-3 text-sm ${
              r.level === 'error' ? 'border-red-500/30 bg-red-500/5' :
              r.level === 'warn' ? 'border-yellow-500/30 bg-yellow-500/5' :
              'border-blue-500/30 bg-blue-500/5'
            }`}>
              <RiskIcon level={r.level} />
              <div className="flex-1">
                <p className="text-foreground">{r.message}</p>
                {r.count != null && <p className="text-xs text-muted-foreground mt-0.5">数量: {r.count}</p>}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Storage */}
      <div>
        <h2 className="flex items-center gap-2 text-sm font-semibold mb-3"><HardDrive className="size-4" />{t('admin.storageHealth', { defaultValue: '存储状况' })}</h2>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          {data.storage?.disk && (
            <div className="rounded-xl border bg-card p-4">
              <p className="text-xs text-muted-foreground uppercase tracking-wider">{t('admin.diskUsage', { defaultValue: '磁盘用量' })}</p>
              <div className="mt-2">
                <p className="text-2xl font-bold">{formatBytes(data.storage.disk.free_bytes)}</p>
                <p className="text-xs text-muted-foreground">可用 / {formatBytes(data.storage.disk.total_bytes)} 总量</p>
              </div>
              <div className="mt-3 h-2 rounded-full bg-muted overflow-hidden">
                <div className="h-full rounded-full bg-primary transition-all" style={{
                  width: `${data.storage.disk.total_bytes > 0 ? 100 - (data.storage.disk.free_bytes / data.storage.disk.total_bytes * 100) : 0}%`
                }} />
              </div>
            </div>
          )}
          <div className="rounded-xl border bg-card p-4">
            <p className="text-xs text-muted-foreground uppercase tracking-wider">{t('admin.imagesTotal', { defaultValue: '图片总数' })}</p>
            <p className="text-2xl font-bold mt-2">{data.usage?.images_total ?? '-'}</p>
            <p className="text-xs text-muted-foreground">{formatBytes(data.usage?.storage_bytes ?? 0)} 已用</p>
          </div>
          {data.storage?.strategies && (
            <div className="rounded-xl border bg-card p-4">
              <p className="text-xs text-muted-foreground uppercase tracking-wider">{t('admin.strategiesHealth', { defaultValue: '存储策略' })}</p>
              <div className="mt-2 space-y-1">
                {data.storage.strategies.map((s: Record<string, unknown>, i: number) => (
                  <div key={i} className="flex items-center gap-2 text-sm">
                    {s.healthy ? <CheckCircle2 className="size-3.5 text-green-500" /> : <XCircle className="size-3.5 text-red-500" />}
                    <span>{s.name as string}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Backup & History */}
      <div>
        <h2 className="flex items-center gap-2 text-sm font-semibold mb-3"><Clock className="size-4" />{t('admin.backup', { defaultValue: '备份状况' })}</h2>
        <div className="rounded-xl border bg-card p-4">
          {data.backup?.status === 'ok' ? (
            <div className="flex items-center gap-3">
              <CheckCircle2 className="size-5 text-green-500" />
              <div>
                <p className="text-sm font-medium">{data.backup.file}</p>
                <p className="text-xs text-muted-foreground">
                  {formatBytes(data.backup.size ?? 0)} · {data.backup.timestamp ? new Date(data.backup.timestamp).toLocaleString() : '-'}
                </p>
              </div>
            </div>
          ) : (
            <div className="flex items-center gap-3">
              <AlertTriangle className="size-5 text-yellow-500" />
              <p className="text-sm text-muted-foreground">
                {data.backup?.status === 'no_storage' ? t('admin.noStorageDir', { defaultValue: '存储目录未配置' }) : t('admin.noBackups', { defaultValue: '尚未创建备份，建议通过 CLI 执行备份' })}
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
