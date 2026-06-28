import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Link } from 'react-router-dom'

import { Webhook, Plus, Trash2, Pencil, Copy, RotateCw, Play } from 'lucide-react'
import {
  createWebhook, deleteWebhook, listWebhooks, updateWebhook,
  rotateWebhookSecret, testWebhook,
} from '../../lib/console-api'
import { extractErrorMessage } from '../../lib/error-handler'
import { copyToClipboard } from '../../lib/clipboard'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { EmptyState, LoadingState } from '@/components/page-states'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription,
} from '@/components/ui/dialog'

const ALL_EVENT_TYPES = [
  'image.uploaded',
  'image.processed',
  'image.deleted',
  'moderation.reviewed',
  'user.registered',
]

export function WebhooksPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()

  const { data: webhooks, isLoading, error } = useQuery({
    queryKey: ['webhooks'],
    queryFn: listWebhooks,
  })

  // Create / Edit dialog
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [formName, setFormName] = useState('')
  const [formUrl, setFormUrl] = useState('')
  const [formEvents, setFormEvents] = useState<string[]>(['image.uploaded'])
  const [submitting, setSubmitting] = useState(false)
  const [createdSecret, setCreatedSecret] = useState<string | null>(null)

  const openCreate = useCallback(() => {
    setEditingId(null)
    setFormName('')
    setFormUrl('')
    setFormEvents(['image.uploaded'])
    setCreatedSecret(null)
    setDialogOpen(true)
  }, [])

  const openEdit = useCallback((wh: { id: number; name: string; url: string; events: string[] }) => {
    setEditingId(wh.id)
    setFormName(wh.name)
    setFormUrl(wh.url)
    setFormEvents(wh.events.length ? wh.events : ['image.uploaded'])
    setCreatedSecret(null)
    setDialogOpen(true)
  }, [])

  const handleSubmit = useCallback(async () => {
    if (!formName.trim() || !formUrl.trim() || formEvents.length === 0) return
    setSubmitting(true)
    try {
      if (editingId) {
        await updateWebhook(editingId, {
          name: formName.trim(),
          url: formUrl.trim(),
          events: formEvents,
        })
        setDialogOpen(false)
        await qc.invalidateQueries({ queryKey: ['webhooks'] })
        toast.success(t('webhooks.confirmUpdate'))
      } else {
        const result = await createWebhook(formName.trim(), formUrl.trim(), formEvents)
        if (result.secret) {
          setCreatedSecret(result.secret)
        } else {
          setDialogOpen(false)
          await qc.invalidateQueries({ queryKey: ['webhooks'] })
        }
      }
    } catch (err: unknown) {
      toast.error(extractErrorMessage(err, editingId ? t('webhooks.updateFailed') : t('webhooks.createFailed')))
    } finally {
      setSubmitting(false)
    }
  }, [formName, formUrl, formEvents, editingId, qc, t])

  const handleToggleEnabled = useCallback(async (id: number, enabled: boolean) => {
    try {
      await updateWebhook(id, { enabled })
      await qc.invalidateQueries({ queryKey: ['webhooks'] })
    } catch (err: unknown) {
      toast.error(extractErrorMessage(err, t('webhooks.updateFailed')))
    }
  }, [qc, t])

  // Delete
  const [deleting, setDeleting] = useState<number | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<number | null>(null)

  const handleDelete = useCallback(async () => {
    if (deleteTarget === null) return
    setDeleting(deleteTarget)
    try {
      await deleteWebhook(deleteTarget)
      setDeleteTarget(null)
      await qc.invalidateQueries({ queryKey: ['webhooks'] })
    } catch (err: unknown) {
      toast.error(extractErrorMessage(err, t('webhooks.deleteFailed')))
    } finally {
      setDeleting(null)
    }
  }, [deleteTarget, qc, t])

  // Rotate secret
  const [rotatingId, setRotatingId] = useState<number | null>(null)
  const [rotateTarget, setRotateTarget] = useState<number | null>(null)
  const [newSecret, setNewSecret] = useState<string | null>(null)

  const handleRotate = useCallback(async () => {
    if (rotateTarget === null) return
    setRotatingId(rotateTarget)
    try {
      const result = await rotateWebhookSecret(rotateTarget)
      setNewSecret(result.secret)
      setRotateTarget(null)
      await qc.invalidateQueries({ queryKey: ['webhooks'] })
    } catch (err: unknown) {
      toast.error(extractErrorMessage(err, t('webhooks.rotateSecretFailed')))
    } finally {
      setRotatingId(null)
    }
  }, [rotateTarget, qc, t])

  // Test
  const [testingId, setTestingId] = useState<number | null>(null)
  const handleTest = useCallback(async (id: number) => {
    setTestingId(id)
    try {
      await testWebhook(id)
      toast.success(t('webhooks.testSuccess'))
    } catch (err: unknown) {
      toast.error(extractErrorMessage(err, t('webhooks.testFailed')))
    } finally {
      setTestingId(null)
    }
  }, [t])

  // Copy
  const onCopy = async (text: string) => {
    try {
      await copyToClipboard(text)
      toast.success(t('upload.copied'))
    } catch {
      toast.error(t('upload.copyError'))
    }
  }

  const toggleEvent = (ev: string) => {
    setFormEvents(prev =>
      prev.includes(ev) ? prev.filter(e => e !== ev) : [...prev, ev],
    )
  }

  if (isLoading) return <LoadingState className="min-h-[40vh]" compact />
  if (error) return <EmptyState title={t('webhooks.loadFailed')} icon={<Webhook className="size-6 text-muted-foreground" />} />

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{t('page.webhooks.title')}</h1>
          <p className="text-sm text-muted-foreground mt-1">{t('webhooks.emptyDesc')}</p>
        </div>
        <Button onClick={openCreate} size="sm">
          <Plus className="h-4 w-4 mr-1" />
          {t('webhooks.create')}
        </Button>
      </div>

      {(!webhooks || webhooks.length === 0) ? (
        <EmptyState
          title={t('webhooks.empty')}
          icon={<Webhook className="size-6 text-muted-foreground" />}
          action={
            <Button onClick={openCreate} size="sm" variant="outline">
              <Plus className="h-4 w-4 mr-1" />
              {t('webhooks.create')}
            </Button>
          }
        />
      ) : (
        <div className="space-y-3">
          {webhooks.map((wh) => (
            <div
              key={wh.id}
              className="flex flex-col sm:flex-row sm:items-center gap-3 rounded-xl border border-border/60 bg-card p-4"
            >
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <Link to={`/console/webhooks/${wh.id}`} className="font-medium text-sm hover:underline truncate">
                    {wh.name}
                  </Link>
                  <Switch
                    checked={wh.enabled}
                    onCheckedChange={(v) => handleToggleEnabled(wh.id, v)}
                  />
                </div>
                <p className="text-xs text-muted-foreground mt-1 truncate">{wh.url}</p>
                <div className="flex flex-wrap gap-1 mt-2">
                  {wh.events.map((ev) => (
                    <span key={ev} className="inline-flex items-center rounded-md border border-border/60 px-2 py-0.5 text-xs font-mono text-muted-foreground">
                      {ev}
                    </span>
                  ))}
                </div>
                <p className="text-xs text-muted-foreground mt-1">
                  {t('webhooks.createdAt', { date: new Date(wh.created_at).toLocaleString() })}
                </p>
              </div>
              <div className="flex items-center gap-1 shrink-0">
                <Button variant="ghost" size="icon-sm" onClick={() => handleTest(wh.id)} disabled={testingId === wh.id}>
                  <Play className={`h-4 w-4 ${testingId === wh.id ? 'animate-pulse' : ''}`} />
                </Button>
                <Button variant="ghost" size="icon-sm" onClick={() => setRotateTarget(wh.id)} disabled={rotatingId === wh.id}>
                  <RotateCw className={`h-4 w-4 ${rotatingId === wh.id ? 'animate-spin' : ''}`} />
                </Button>
                <Button variant="ghost" size="icon-sm" onClick={() => openEdit(wh)}>
                  <Pencil className="h-4 w-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  onClick={() => setDeleteTarget(wh.id)}
                  className="text-destructive hover:text-destructive"
                  disabled={deleting === wh.id}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create / Edit Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{editingId ? t('webhooks.edit') : t('webhooks.create')}</DialogTitle>
            <DialogDescription>
              {editingId ? '' : t('webhooks.emptyDesc')}
            </DialogDescription>
          </DialogHeader>

          {createdSecret ? (
            <div className="space-y-4">
              <div className="rounded-lg border border-yellow-500/30 bg-yellow-500/10 p-4">
                <p className="text-sm font-medium text-yellow-600 dark:text-yellow-400">
                  {t('webhooks.secretWarning')}
                </p>
                <div className="mt-2 flex items-center gap-2">
                  <code className="flex-1 break-all rounded bg-background px-3 py-2 text-sm font-mono border">
                    {createdSecret}
                  </code>
                  <Button size="icon-sm" variant="outline" onClick={() => onCopy(createdSecret)}>
                    <Copy className="h-4 w-4" />
                  </Button>
                </div>
              </div>
              <Button
                className="w-full"
                onClick={() => {
                  setCreatedSecret(null)
                  setDialogOpen(false)
                  qc.invalidateQueries({ queryKey: ['webhooks'] })
                }}
              >
                {t('webhooks.dismiss')}
              </Button>
            </div>
          ) : (
            <>
              <div className="space-y-4">
                <div>
                  <label className="text-sm font-medium">{t('webhooks.namePlaceholder')}</label>
                  <input
                    className="mt-1 w-full rounded-lg border border-input bg-transparent px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                    placeholder={t('webhooks.namePlaceholder')}
                    value={formName}
                    onChange={(e) => setFormName(e.target.value)}
                  />
                </div>
                <div>
                  <label className="text-sm font-medium">{t('webhooks.urlPlaceholder')}</label>
                  <input
                    className="mt-1 w-full rounded-lg border border-input bg-transparent px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                    placeholder="https://example.com/webhook"
                    value={formUrl}
                    onChange={(e) => setFormUrl(e.target.value)}
                  />
                </div>
                <div>
                  <label className="text-sm font-medium">{t('webhooks.eventsLabel')}</label>
                  <p className="text-xs text-muted-foreground mb-2">{t('webhooks.eventsHint')}</p>
                  <div className="flex flex-wrap gap-2">
                    {ALL_EVENT_TYPES.map((ev) => (
                      <button
                        key={ev}
                        type="button"
                        onClick={() => toggleEvent(ev)}
                        className={`inline-flex items-center rounded-lg border px-3 py-1.5 text-xs font-mono transition-colors ${
                          formEvents.includes(ev)
                            ? 'border-primary bg-primary/10 text-primary'
                            : 'border-border/60 text-muted-foreground hover:border-primary/50'
                        }`}
                      >
                        {ev}
                      </button>
                    ))}
                  </div>
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setDialogOpen(false)} disabled={submitting}>
                  {t('webhooks.cancel')}
                </Button>
                <Button
                  onClick={handleSubmit}
                  disabled={submitting || !formName.trim() || !formUrl.trim() || formEvents.length === 0}
                >
                  {submitting
                    ? (editingId ? t('webhooks.updating') : t('webhooks.creating'))
                    : (editingId ? t('webhooks.confirmUpdate') : t('webhooks.confirmCreate'))
                  }
                </Button>
              </DialogFooter>
            </>
          )}
        </DialogContent>
      </Dialog>

      {/* New secret after rotate */}
      {newSecret && (
        <Dialog open={!!newSecret} onOpenChange={() => setNewSecret(null)}>
          <DialogContent className="sm:max-w-lg">
            <DialogHeader>
              <DialogTitle>{t('webhooks.newSecret')}</DialogTitle>
              <DialogDescription>{t('webhooks.secretWarning')}</DialogDescription>
            </DialogHeader>
            <div className="space-y-4">
              <div className="flex items-center gap-2">
                <code className="flex-1 break-all rounded bg-muted px-3 py-2 text-sm font-mono">
                  {newSecret}
                </code>
                <Button size="icon-sm" variant="outline" onClick={() => onCopy(newSecret)}>
                  <Copy className="h-4 w-4" />
                </Button>
              </div>
              <Button className="w-full" onClick={() => setNewSecret(null)}>
                {t('webhooks.dismiss')}
              </Button>
            </div>
          </DialogContent>
        </Dialog>
      )}

      {/* Rotate secret confirm */}
      <ConfirmDialog
        open={rotateTarget !== null}
        onOpenChange={(v) => { if (!v) setRotateTarget(null) }}
        title={t('webhooks.rotateSecret')}
        description={t('webhooks.rotateSecretConfirm')}
        confirmLabel={t('webhooks.rotateSecret')}
        variant="warning"
        onConfirm={handleRotate}
        loading={rotatingId !== null}
      />

      {/* Delete confirm */}
      <ConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(v) => { if (!v) setDeleteTarget(null) }}
        title={t('webhooks.confirmDelete')}
        description={t('webhooks.deleteDescription')}
        confirmLabel={t('webhooks.delete')}
        variant="destructive"
        onConfirm={handleDelete}
        loading={deleting !== null}
      />
    </div>
  )
}
