import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { 
  ShieldCheck, AlertTriangle, Info, CheckCircle2, HardDrive, 
  Clock, XCircle, Layers, Activity, Image, Wrench, 
  Cloud, Server, CloudRain, Loader2, BookOpen 
} from 'lucide-react'
import { toast } from 'sonner'
import { getMaintenanceSummary, cleanupExpiredImages, recalcPHash, type MaintenanceSummary, type MaintenanceRisk, type MaintenanceSummaryStrategy } from '../../../lib/admin-api'
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

function StrategyIcon({ type }: { type: string }) {
  const cls = "size-5 shrink-0"
  switch (type?.toLowerCase()) {
    case 'local': return <HardDrive className={cls} />
    case 's3': return <Cloud className={cls} />
    case 'oss': return <Server className={cls} />
    case 'kodo': return <CloudRain className={cls} />
    default: return <Server className={cls} />
  }
}

function DonutChart({ percentage, colorClass }: { percentage: number, colorClass: string }) {
  const radius = 36
  const circumference = 2 * Math.PI * radius
  const offset = circumference - (percentage / 100) * circumference
  return (
    <div className="relative size-24 flex items-center justify-center shrink-0">
      <svg className="size-full -rotate-90 transform" viewBox="0 0 100 100">
        <circle
          className="text-muted/20"
          strokeWidth="8"
          stroke="currentColor"
          fill="transparent"
          r={radius}
          cx="50"
          cy="50"
        />
        <circle
          className={colorClass}
          strokeWidth="8"
          strokeDasharray={circumference}
          strokeDashoffset={offset}
          strokeLinecap="round"
          stroke="currentColor"
          fill="transparent"
          r={radius}
          cx="50"
          cy="50"
        />
      </svg>
      <div className="absolute inset-0 flex items-center justify-center text-sm font-bold">
        {Math.round(percentage)}%
      </div>
    </div>
  )
}

export function MaintenancePage() {
  const { t, i18n } = useTranslation()
  const qc = useQueryClient()

  const docsUrl = i18n.language?.startsWith('zh')
    ? 'https://picfast.dev/zh/docs/maintenance/'
    : 'https://picfast.dev/docs/maintenance/'
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

  const [recalcRunning, setRecalcRunning] = useState(false)
  const handleRecalc = async () => {
    setRecalcRunning(true)
    try {
      const msg = await recalcPHash()
      toast.success(msg)
      qc.invalidateQueries({ queryKey: ['maintenance-summary'] })
    } catch (err: unknown) {
      toast.error(extractErrorMessage(err, 'recalc failed'))
    } finally {
      setRecalcRunning(false)
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
    <div className="space-y-8 pb-10">
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold tracking-tight">{t('admin.maintenanceTitle', { defaultValue: '系统维护' })}</h1>
            <a
              href={docsUrl}
              target="_blank"
              rel="noreferrer"
              className="inline-flex items-center gap-1.5 rounded-lg border border-border/60 bg-muted/30 px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:border-primary/30 hover:text-primary"
            >
              <BookOpen className="size-3.5" />
              {t('admin.maintenanceDocs', { defaultValue: '维护文档' })}
            </a>
          </div>
          <p className="text-sm text-muted-foreground mt-1">{t('admin.maintenanceDesc', { defaultValue: '数据完整性、存储用量、风险巡检' })}</p>
        </div>
        <div className="flex items-center gap-3 rounded-xl border bg-card px-4 py-3 shadow-sm">
          {statusIcon}
          <div>
            <p className="text-sm font-semibold">{errorCount > 0 ? `${errorCount} 个错误` : warnCount > 0 ? `${warnCount} 个提醒` : '系统正常'}</p>
            <p className="text-xs text-muted-foreground">{new Date(data.generated_at).toLocaleString()}</p>
          </div>
        </div>
      </div>

      {/* Overview Dashboard */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {data.storage?.disk ? (
          <div className="rounded-xl border bg-card p-6 flex items-center gap-6 shadow-sm">
            <DonutChart 
              percentage={data.storage.disk.total_bytes > 0 ? (1 - data.storage.disk.free_bytes / data.storage.disk.total_bytes) * 100 : 0} 
              colorClass="text-primary" 
            />
            <div>
              <p className="text-xs text-muted-foreground uppercase tracking-wider font-medium">{t('admin.diskUsage', { defaultValue: '磁盘用量' })}</p>
              <div className="mt-2">
                <p className="text-3xl font-bold">{formatBytes(data.storage.disk.total_bytes - data.storage.disk.free_bytes)}</p>
                <p className="text-sm text-muted-foreground mt-1">
                  可用 {formatBytes(data.storage.disk.free_bytes)} / {formatBytes(data.storage.disk.total_bytes)} 总量
                </p>
              </div>
            </div>
          </div>
        ) : (
          <div className="rounded-xl border bg-card p-6 shadow-sm flex items-center justify-center text-muted-foreground">
            {t('admin.noDiskInfo', { defaultValue: '无磁盘信息' })}
          </div>
        )}
        
        <div className="rounded-xl border bg-card p-6 flex flex-col justify-center shadow-sm">
          <p className="text-xs text-muted-foreground uppercase tracking-wider font-medium">{t('admin.imagesTotal', { defaultValue: '图片总数' })}</p>
          <div className="mt-2 flex items-baseline gap-3">
            <p className="text-4xl font-bold">{data.usage?.images_total ?? '-'}</p>
            <p className="text-sm text-muted-foreground">/ {formatBytes(data.usage?.storage_bytes ?? 0)} 已用</p>
          </div>
          <div className="mt-4 h-1.5 w-full bg-muted rounded-full overflow-hidden">
            <div className="h-full bg-blue-500/50 rounded-full w-1/3" />
          </div>
        </div>
      </div>

      {/* Health & Risks Panel */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="space-y-3">
          <h2 className="flex items-center gap-2 text-sm font-semibold"><AlertTriangle className="size-4" />{t('admin.systemHealth', { defaultValue: '风险与提醒' })}</h2>
          <div className="rounded-xl border bg-card overflow-hidden shadow-sm h-[calc(100%-2rem)]">
            {risks.length === 0 ? (
               <div className="p-6 flex flex-col items-center justify-center h-full text-center bg-green-500/5">
                 <CheckCircle2 className="size-8 text-green-500 mb-2" />
                 <p className="text-sm font-medium">未发现明显系统风险</p>
                 <p className="text-xs text-muted-foreground mt-1">各项服务运行正常</p>
               </div>
            ) : (
               <div className="divide-y">
                 {risks.map((r: MaintenanceRisk, i: number) => (
                   <div key={i} className={`flex items-start gap-3 p-4 text-sm ${
                      r.level === 'error' ? 'bg-red-500/5' :
                      r.level === 'warn' ? 'bg-yellow-500/5' :
                      'bg-blue-500/5'
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
          </div>
        </div>

        <div className="space-y-3">
          <h2 className="flex items-center gap-2 text-sm font-semibold"><Clock className="size-4" />{t('admin.backup', { defaultValue: '备份状况' })}</h2>
          <div className="rounded-xl border bg-card p-6 shadow-sm h-[calc(100%-2rem)] flex items-center">
            {data.backup?.status === 'ok' ? (
              <div className="flex items-center gap-5 w-full">
                <div className="p-4 rounded-full bg-green-500/10 shrink-0">
                  <CheckCircle2 className="size-8 text-green-500" />
                </div>
                <div className="flex-1 overflow-hidden">
                  <p className="text-base font-medium truncate">{data.backup.file}</p>
                  <p className="text-sm text-muted-foreground mt-1">
                    {formatBytes(data.backup.size ?? 0)} · {data.backup.timestamp ? new Date(data.backup.timestamp).toLocaleString() : '-'}
                  </p>
                </div>
              </div>
            ) : (
              <div className="flex items-center gap-5 w-full">
                <div className="p-4 rounded-full bg-yellow-500/10 shrink-0">
                  <AlertTriangle className="size-8 text-yellow-500" />
                </div>
                <div>
                  <p className="text-base font-medium">无可用备份</p>
                  <p className="text-sm text-muted-foreground mt-1">
                    {data.backup?.status === 'no_storage' ? t('admin.noStorageDir', { defaultValue: '存储目录未配置' }) : t('admin.noBackups', { defaultValue: '尚未创建备份，建议通过 CLI 执行备份' })}
                  </p>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Storage Strategies List */}
      {data.storage?.strategies && data.storage.strategies.length > 0 && (
        <div className="space-y-3">
          <h2 className="flex items-center gap-2 text-sm font-semibold"><HardDrive className="size-4" />{t('admin.strategiesHealth', { defaultValue: '存储策略' })}</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {data.storage.strategies.map((s: MaintenanceSummaryStrategy, i: number) => (
              <div key={i} className="flex items-center gap-4 rounded-xl border bg-card p-5 shadow-sm hover:border-primary/20 transition-colors">
                <div className={`p-3 rounded-xl ${s.healthy ? 'bg-green-500/10 text-green-500' : 'bg-red-500/10 text-red-500'}`}>
                   <StrategyIcon type={s.type} />
                </div>
                <div className="flex-1 overflow-hidden">
                  <div className="flex items-center justify-between mb-1">
                    <p className="text-base font-medium truncate">{s.name}</p>
                    {s.healthy ? <CheckCircle2 className="size-4 text-green-500 shrink-0" /> : <XCircle className="size-4 text-red-500 shrink-0" />}
                  </div>
                  <div className="flex items-center gap-2 mt-1.5">
                    <span className="text-xs text-muted-foreground uppercase px-2 py-0.5 rounded-md bg-muted/50 border border-border/50 font-medium">
                      {s.type || 'Unknown'}
                    </span>
                    {s.error && <span className="text-[10px] text-red-500 truncate" title={s.error}>{s.error}</span>}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Database & Metadata Stats */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* pHash Coverage */}
        {data.phash_coverage && (
          <div className="space-y-3">
            <h2 className="flex items-center gap-2 text-sm font-semibold"><Activity className="size-4" />{t('admin.phashCoverage', { defaultValue: '感知哈希覆盖' })}</h2>
            <div className="rounded-xl border bg-card p-5 shadow-sm h-[calc(100%-2rem)] flex flex-col justify-center">
              <div className="flex items-center justify-between mb-3">
                <span className="text-sm font-medium">{data.phash_coverage.with_phash as number} / {data.phash_coverage.total as number}</span>
                <span className="text-xs font-bold text-muted-foreground bg-muted px-2 py-1 rounded-md">
                  {data.phash_coverage.total as number > 0 ? Math.round((data.phash_coverage.with_phash as number / data.phash_coverage.total as number) * 100) : 0}%
                </span>
              </div>
              <div className="h-2.5 rounded-full bg-muted overflow-hidden">
                <div className="h-full rounded-full bg-green-500 transition-all" style={{
                  width: `${data.phash_coverage.total as number > 0 ? (data.phash_coverage.with_phash as number / data.phash_coverage.total as number) * 100 : 0}%`
                }} />
              </div>
              {(data.phash_coverage.with_phash as number) < (data.phash_coverage.total as number) && (
                <p className="mt-4 text-xs text-muted-foreground bg-blue-500/5 p-2 rounded-lg border border-blue-500/10">
                  {t('admin.phashHint', { defaultValue: '历史图片尚未计算感知哈希，可通过 CLI 或维护工具补算' })}
                </p>
              )}
            </div>
          </div>
        )}

        {/* Thumbnails */}
        {data.thumbnails && (
          <div className="space-y-3">
            <h2 className="flex items-center gap-2 text-sm font-semibold"><Image className="size-4" />{t('admin.thumbnails', { defaultValue: '缩略图' })}</h2>
            <div className="rounded-xl border bg-card p-5 shadow-sm h-[calc(100%-2rem)] flex flex-col justify-center">
              <p className="text-3xl font-bold">{data.thumbnails.on_disk as number}</p>
              <p className="text-sm text-muted-foreground mt-1">{t('admin.thumbsOnDisk', { defaultValue: '个缓存文件' })}</p>
              <p className="text-xs text-muted-foreground mt-4 truncate bg-muted/50 p-2 rounded-lg border border-border/50" title={data.thumbnails.dir as string}>
                {data.thumbnails.dir as string}
              </p>
            </div>
          </div>
        )}

        {/* Database */}
        {data.database && data.database.length > 0 && (
          <div className="space-y-3 lg:col-span-1">
            <h2 className="flex items-center gap-2 text-sm font-semibold"><Layers className="size-4" />{t('admin.databaseStats', { defaultValue: '数据库' })}</h2>
            <div className="grid grid-cols-2 gap-3 h-[calc(100%-2rem)]">
              {data.database.slice(0, 4).map((row: Record<string, unknown>, i: number) => (
                <div key={i} className="rounded-xl border bg-card px-3 py-3 text-center shadow-sm flex flex-col justify-center">
                  <p className="text-lg font-bold">{String(row.rows)}</p>
                  <p className="text-xs text-muted-foreground truncate mt-1">{row.table as string}</p>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Tools */}
      <div className="pt-2">
        <h2 className="flex items-center gap-2 text-sm font-semibold mb-3"><Wrench className="size-4" />{t('admin.tools', { defaultValue: '维护工具' })}</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="rounded-xl border bg-card p-5 flex flex-col justify-between gap-6 shadow-sm">
            <div>
              <p className="text-base font-medium">{t('admin.cleanExpired', { defaultValue: '清理过期图片' })}</p>
              <p className="text-sm text-muted-foreground mt-1.5">{t('admin.cleanExpiredDesc', { defaultValue: '删除所有已过期的图片及其文件' })}</p>
            </div>
            <div className="flex items-center justify-between mt-auto pt-4 border-t border-border/50">
              <span className="text-xs text-muted-foreground flex items-center gap-1.5">
                <Clock className="size-3.5" /> 按需手动执行
              </span>
              <Button size="sm" onClick={handleCleanup} disabled={cleaning} className="w-24">
                {cleaning ? <Loader2 className="size-4 animate-spin" /> : t('admin.run', { defaultValue: '执行' })}
              </Button>
            </div>
          </div>
          
          <div className="rounded-xl border bg-card p-5 flex flex-col justify-between gap-6 shadow-sm">
            <div>
              <p className="text-base font-medium">{t('admin.recalcPhash', { defaultValue: '补算感知哈希' })}</p>
              <p className="text-sm text-muted-foreground mt-1.5">{t('admin.recalcPhashDesc', { defaultValue: '为缺少哈希的历史图片计算感知哈希' })}</p>
            </div>
            <div className="flex items-center justify-between mt-auto pt-4 border-t border-border/50">
               <span className="text-xs text-muted-foreground flex items-center gap-1.5">
                <Activity className="size-3.5" /> 后台异步任务
              </span>
              <Button size="sm" onClick={handleRecalc} disabled={recalcRunning} className="w-24">
                {recalcRunning ? <Loader2 className="size-4 animate-spin" /> : t('admin.run', { defaultValue: '执行' })}
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
