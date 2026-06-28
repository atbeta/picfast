import { useCallback, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'

import { ArrowLeft, RotateCw, Clock, AlertCircle, CheckCircle2, XCircle, Ban, ChevronDown, ChevronRight } from 'lucide-react'
import { getWebhook, listWebhookDeliveries, replayWebhookDelivery, type WebhookDelivery } from '../../lib/console-api'
import { extractErrorMessage } from '../../lib/error-handler'
import { LoadingState, EmptyState } from '@/components/page-states'
import { Button } from '@/components/ui/button'

const STATUS_FILTERS = ['all', 'pending', 'retrying', 'delivered', 'failed', 'dead'] as const
type StatusFilter = (typeof STATUS_FILTERS)[number]

function StatusIcon({ status }: { status: string }) {
  const cls = 'h-4 w-4'
  switch (status) {
    case 'delivered': return <CheckCircle2 className={`${cls} text-green-500`} />
    case 'retrying': return <Clock className={`${cls} text-yellow-500`} />
    case 'failed': return <XCircle className={`${cls} text-red-500`} />
    case 'dead': return <Ban className={`${cls} text-red-600`} />
    default: return <AlertCircle className={`${cls} text-muted-foreground`} />
  }
}

function formatRetryAt(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString()
}

export function WebhookDetailPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const { id } = useParams<{ id: string }>()
  const webhookId = Number(id)

  const { data: webhook, isLoading, error } = useQuery({
    queryKey: ['webhook', webhookId],
    queryFn: () => getWebhook(webhookId),
    enabled: !isNaN(webhookId),
  })

  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all')
  const { data: deliveriesData } = useQuery({
    queryKey: ['webhook-deliveries', webhookId, page],
    queryFn: () => listWebhookDeliveries(webhookId, page, 50),
    enabled: !isNaN(webhookId),
  })

  const [replayingId, setReplayingId] = useState<number | null>(null)
  const [expandedId, setExpandedId] = useState<number | null>(null)
  const handleReplay = useCallback(async (deliveryId: number) => {
    setReplayingId(deliveryId)
    try {
      await replayWebhookDelivery(deliveryId)
      toast.success(t('webhooks.deliveryReplaySuccess'))
      await qc.invalidateQueries({ queryKey: ['webhook-deliveries', webhookId] })
    } catch (err: unknown) {
      toast.error(extractErrorMessage(err, t('webhooks.deliveryReplayFailed')))
    } finally {
      setReplayingId(null)
    }
  }, [webhookId, qc, t])

  if (isLoading) return <LoadingState className="min-h-[40vh]" compact />
  if (error || !webhook) return <EmptyState title={t('webhooks.loadFailed')} />
  if (isNaN(webhookId)) return <EmptyState title={t('webhooks.loadFailed')} />

  const allDeliveries = deliveriesData?.items ?? []
  const total = deliveriesData?.total ?? 0
  const totalPages = Math.ceil(total / 50)

  const filteredDeliveries = statusFilter === 'all'
    ? allDeliveries
    : allDeliveries.filter(d => d.status === statusFilter)

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Link to="/console/webhooks" className="text-muted-foreground hover:text-foreground">
          <ArrowLeft className="h-5 w-5" />
        </Link>
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{webhook.name}</h1>
          <p className="text-sm text-muted-foreground mt-0.5">{webhook.url}</p>
        </div>
        <div className="ml-auto flex flex-wrap items-center gap-2">
          {webhook.events.map((ev) => (
            <span key={ev} className="inline-flex items-center rounded-md border border-border/60 px-2 py-0.5 text-xs font-mono text-muted-foreground">
              {ev}
            </span>
          ))}
        </div>
      </div>

      <div>
        <h2 className="text-lg font-semibold mb-3">{t('webhooks.deliveries')}</h2>

        <div className="flex items-center gap-1 mb-3">
          {STATUS_FILTERS.map((s) => (
            <button
              key={s}
              onClick={() => { setStatusFilter(s); setPage(1) }}
              className={`rounded-md px-2.5 py-1 text-xs font-medium transition-colors ${
                statusFilter === s
                  ? 'bg-primary/10 text-primary border border-primary/30'
                  : 'text-muted-foreground hover:text-foreground border border-transparent'
              }`}
            >
              {s === 'all' ? t('admin.auditActionAll', { defaultValue: 'All' }) : t(`webhooks.${s}`)}
            </button>
          ))}
        </div>

        {filteredDeliveries.length === 0 ? (
          <EmptyState title={t('webhooks.deliveriesEmpty', { defaultValue: '暂无投递记录' })} compact />
        ) : (
          <div className="space-y-2">
            <div className="text-xs text-muted-foreground mb-2">
              {statusFilter === 'all'
                ? t('webhooks.deliveriesCount', { count: total })
                : `${filteredDeliveries.length} ${t(`webhooks.${statusFilter}`)}`
              }
            </div>
            {filteredDeliveries.map((d: WebhookDelivery) => (
              <div key={d.id}>
                <div className="flex items-center gap-3 rounded-lg border border-border/60 bg-card p-3 text-sm">
                  <StatusIcon status={d.status} />
                  <span className="w-16 text-xs font-medium">
                    {t(`webhooks.${d.status}`)}
                  </span>
                  <span className="text-xs text-muted-foreground">
                    #{d.attempt}/{d.max_attempts}
                  </span>
                  {d.response_status && (
                    <span className={`text-xs font-mono ${d.response_status < 300 ? 'text-green-600' : 'text-red-600'}`}>
                      HTTP {d.response_status}
                    </span>
                  )}
                  {d.duration_ms != null && (
                    <span className="text-xs text-muted-foreground">{d.duration_ms}ms</span>
                  )}
                  {d.error_message && (
                    <span className="flex-1 truncate text-xs text-red-500" title={d.error_message}>
                      {d.error_message}
                    </span>
                  )}
                  {(d.status === 'retrying' || d.status === 'pending') && d.next_retry_at && (
                    <span className="text-xs text-muted-foreground shrink-0" title={formatRetryAt(d.next_retry_at)}>
                      {t('webhooks.deliveryNextRetry')}: {formatRetryAt(d.next_retry_at)}
                    </span>
                  )}
                  <span className="text-xs text-muted-foreground ml-auto shrink-0">
                    {new Date(d.created_at).toLocaleString()}
                  </span>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    onClick={() => handleReplay(d.id)}
                    disabled={replayingId === d.id}
                    title={t('webhooks.deliveryReplay')}
                  >
                    <RotateCw className={`h-3.5 w-3.5 ${replayingId === d.id ? 'animate-spin' : ''}`} />
                  </Button>
                  {d.response_body && (
                    <button
                      onClick={() => setExpandedId(expandedId === d.id ? null : d.id)}
                      className="text-muted-foreground hover:text-foreground"
                    >
                      {expandedId === d.id ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
                    </button>
                  )}
                </div>
                {expandedId === d.id && d.response_body && (
                  <pre className="mt-1 rounded-lg border border-border/40 bg-muted/30 p-3 text-xs font-mono text-muted-foreground whitespace-pre-wrap overflow-x-auto max-h-40 overflow-y-auto">
                    {d.response_body}
                  </pre>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2 mt-4">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page <= 1}
          >
            &laquo;
          </Button>
          <span className="text-sm text-muted-foreground px-2">
            {page} / {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            disabled={page >= totalPages}
          >
            &raquo;
          </Button>
        </div>
      )}
    </div>
  )
}
