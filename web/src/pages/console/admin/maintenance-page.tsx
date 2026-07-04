import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { ShieldCheck, AlertTriangle, Info, CheckCircle2, HardDrive, Clock, XCircle, Layers, Activity, Image, Wrench, Terminal, Play } from 'lucide-react'
import { toast } from 'sonner'
import { getMaintenanceSummary, cleanupExpiredImages, type MaintenanceSummary, type MaintenanceRisk } from '../../../lib/admin-api'
import { extractErrorMessage } from '../../../lib/error-handler'
import { LoadingState } from '@/components/page-states'
import { Button } from '@/components/ui/button'

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
  const qc = useQueryClient()
  const { data, isLoading, error } = useQuery<MaintenanceSummary>({
    queryKey: ['maintenance-summary'],
    queryFn: getMaintenanceSummary,
    refetchInterval: 30000,
  })

  const [cleaning, setCleaning] = useState(false)
  const handleCleanup = async () => {
    setCleaning(true)
    try {
      const msg = await cleanupExpiredImages()
      toast.success(msg)
      qc.invalidateQueries({ queryKey: ['maintenance-summary'] })
    } catch (err: unknown) {
      toast.error(extractErrorMessage(err, 'cleanup failed'))
    } finally {
      setCleaning(false)
    }
  }

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

      {/* Database */}
      {data.database && data.database.length > 0 && (
        <div>
          <h2 className="flex items-center gap-2 text-sm font-semibold mb-3"><Layers className="size-4" />{t('admin.databaseStats', { defaultValue: '数据库' })}</h2>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-2">
            {data.database.map((t: Record<string, unknown>, i: number) => (
              <div key={i} className="rounded-lg border bg-card px-3 py-2 text-center">
                <p className="text-xs text-muted-foreground truncate">{t.table as string}</p>
                <p className="text-lg font-bold">{String(t.rows)}</p>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* pHash Coverage */}
      {data.phash_coverage && (
        <div>
          <h2 className="flex items-center gap-2 text-sm font-semibold mb-3"><Activity className="size-4" />{t('admin.phashCoverage', { defaultValue: '感知哈希覆盖' })}</h2>
          <div className="rounded-xl border bg-card p-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm">{data.phash_coverage.with_phash as number} / {data.phash_coverage.total as number}</span>
              <span className="text-xs text-muted-foreground">{data.phash_coverage.total as number > 0 ? Math.round((data.phash_coverage.with_phash as number / data.phash_coverage.total as number) * 100) : 0}%</span>
            </div>
            <div className="h-2 rounded-full bg-muted overflow-hidden">
              <div className="h-full rounded-full bg-green-500 transition-all" style={{
                width: `${data.phash_coverage.total as number > 0 ? (data.phash_coverage.with_phash as number / data.phash_coverage.total as number) * 100 : 0}%`
              }} />
            </div>
            {(data.phash_coverage.with_phash as number) < (data.phash_coverage.total as number) && (
              <p className="mt-2 text-xs text-muted-foreground">
                {t('admin.phashHint', { defaultValue: '历史图片尚未计算感知哈希，可通过 CLI 或维护工具补算' })}
              </p>
            )}
          </div>
        </div>
      )}

      {/* Thumbnails */}
      {data.thumbnails && (
        <div>
          <h2 className="flex items-center gap-2 text-sm font-semibold mb-3"><Image className="size-4" />{t('admin.thumbnails', { defaultValue: '缩略图' })}</h2>
          <div className="rounded-xl border bg-card p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium">{data.thumbnails.on_disk as number} {t('admin.thumbsOnDisk', { defaultValue: '个文件' })}</p>
                <p className="text-xs text-muted-foreground">{data.thumbnails.dir as string}</p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Tools */}
      <div>
        <h2 className="flex items-center gap-2 text-sm font-semibold mb-3"><Wrench className="size-4" />{t('admin.tools', { defaultValue: '维护工具' })}</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div className="rounded-xl border bg-card p-4 flex items-center justify-between">
            <div>
              <p className="text-sm font-medium">{t('admin.cleanExpired', { defaultValue: '清理过期图片' })}</p>
              <p className="text-xs text-muted-foreground">{t('admin.cleanExpiredDesc', { defaultValue: '删除所有已过期的图片及其文件' })}</p>
            </div>
            <Button size="sm" variant="outline" onClick={handleCleanup} disabled={cleaning}>
              {cleaning ? '...' : t('admin.run', { defaultValue: '执行' })}
            </Button>
          </div>
          <div className="rounded-xl border bg-card p-4 flex items-center justify-between">
            <div>
              <p className="text-sm font-medium">{t('admin.repairThumbnails', { defaultValue: '修复缩略图' })}</p>
              <p className="text-xs text-muted-foreground">{t('admin.repairThumbnailsDesc', { defaultValue: '重建所有丢失的缩略图文件' })}</p>
            </div>
            <Button size="sm" variant="outline" disabled>
              <Terminal className="size-3.5 mr-1" />CLI
            </Button>
          </div>
          <div className="rounded-xl border bg-card p-4 flex items-center justify-between">
            <div>
              <p className="text-sm font-medium">{t('admin.dataDoctor', { defaultValue: '数据完整性校验' })}</p>
              <p className="text-xs text-muted-foreground">{t('admin.dataDoctorDesc', { defaultValue: '校验存储对象与缩略图一致性（不修改数据）' })}</p>
            </div>
            <Button size="sm" variant="outline" disabled>
              <Terminal className="size-3.5 mr-1" />CLI
            </Button>
          </div>
          <div className="rounded-xl border bg-card p-4 flex items-center justify-between">
            <div>
              <p className="text-sm font-medium">{t('admin.recalcPhash', { defaultValue: '补算感知哈希' })}</p>
              <p className="text-xs text-muted-foreground">{t('admin.recalcPhashDesc', { defaultValue: '为无哈希的历史图片计算感知哈希' })}</p>
            </div>
            <Button size="sm" variant="outline" disabled>
              <Terminal className="size-3.5 mr-1" />CLI
            </Button>
          </div>
          <div className="rounded-xl border bg-card p-4 sm:col-span-2">
            <p className="text-sm font-medium flex items-center gap-2">
              <Play className="size-4 text-green-500" />
              {t('admin.backupNow', { defaultValue: '创建备份' })}
            </p>
            <p className="text-xs text-muted-foreground mt-1">{t('admin.backupNowDesc', { defaultValue: '创建完整备份（含数据库 + 文件），下载 tar.gz 归档' })}</p>
            <code className="mt-2 block rounded bg-muted px-2 py-1 text-xs font-mono">
              picfast maintenance backup -o backup.tar.gz
            </code>
          </div>
          <div className="rounded-xl border bg-card p-4">
            <p className="text-sm font-medium">{t('admin.logs', { defaultValue: '查看日志' })}</p>
            <p className="text-xs text-muted-foreground mt-1">{t('admin.logsHint', { defaultValue: '实时日志查看命令' })}</p>
            <code className="mt-2 block rounded bg-muted px-2 py-1 text-xs font-mono">journalctl -u picfast -f</code>
          </div>
          <div className="rounded-xl border bg-card p-4">
            <p className="text-sm font-medium">{t('admin.dockerLogs', { defaultValue: 'Docker 日志' })}</p>
            <p className="text-xs text-muted-foreground mt-1">{t('admin.dockerLogsHint', { defaultValue: '容器环境查看日志' })}</p>
            <code className="mt-2 block rounded bg-muted px-2 py-1 text-xs font-mono">docker logs picfast -f</code>
          </div>
        </div>
      </div>
    </div>
  )
}
