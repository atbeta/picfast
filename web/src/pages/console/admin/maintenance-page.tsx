import { Fragment } from 'react'

import { Copy, DatabaseBackup, FileCheck2, HardDrive, RotateCcw, ShieldCheck, Terminal } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import { copyToClipboard } from '@/lib/clipboard'
import { cn } from '@/lib/cn'

const SERVICE_PLACEHOLDER = '<service>' as const

type CommandKey = 'doctor' | 'backup' | 'inspect' | 'restore'

/** Compose service name is deployment-specific; never assume `app`. */
const dockerCommands: Record<CommandKey, string> = {
  doctor: `docker compose exec ${SERVICE_PLACEHOLDER} picfast maintenance doctor --all --batch-size 500`,
  backup: `docker compose exec ${SERVICE_PLACEHOLDER} picfast maintenance backup --output /app/data/backups/picfast-backup.tar.gz`,
  inspect: `docker compose exec ${SERVICE_PLACEHOLDER} picfast maintenance inspect /app/data/backups/picfast-backup.tar.gz`,
  restore: `docker compose exec ${SERVICE_PLACEHOLDER} picfast maintenance restore /app/data/backups/picfast-backup.tar.gz --apply --force`,
}

export function AdminMaintenancePage() {
  const { t } = useTranslation()

  const onCopy = async (value: string) => {
    try {
      await copyToClipboard(value)
      toast.success(t('upload.copySuccess'))
    } catch {
      toast.error(t('upload.copyError'))
    }
  }

  const capabilityItems = [
    {
      icon: FileCheck2,
      title: t('admin.maintenanceDoctorTitle'),
      desc: t('admin.maintenanceDoctorDesc'),
      command: dockerCommands.doctor,
    },
    {
      icon: DatabaseBackup,
      title: t('admin.maintenanceBackupTitle'),
      desc: t('admin.maintenanceBackupDesc'),
      command: dockerCommands.backup,
    },
    {
      icon: ShieldCheck,
      title: t('admin.maintenanceInspectTitle'),
      desc: t('admin.maintenanceInspectDesc'),
      command: dockerCommands.inspect,
    },
    {
      icon: RotateCcw,
      title: t('admin.maintenanceRestoreTitle'),
      desc: t('admin.maintenanceRestoreDesc'),
      command: dockerCommands.restore,
    },
  ]

  const safetyItems = [
    t('admin.maintenanceSafetyInspect'),
    t('admin.maintenanceSafetyWindow'),
    t('admin.maintenanceSafetyForce'),
    t('admin.maintenanceSafetyStorage'),
  ]

  return (
    <section className="space-y-6 pb-8">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">{t('admin.maintenanceTitle')}</h1>
        <p className="mt-1 text-sm text-muted-foreground">{t('admin.maintenanceSubtitle')}</p>
      </div>

      <div className="space-y-3">
        <h2 className="text-base font-semibold text-foreground">{t('admin.maintenanceOpsNotesTitle')}</h2>
        <div className="grid gap-3 lg:grid-cols-2">
          <section
            className={cn(
              'rounded-xl border border-amber-500/20 bg-gradient-to-b from-amber-500/[0.08] to-transparent p-4',
              'dark:border-amber-400/15 dark:from-amber-500/[0.06]',
            )}
          >
            <div className="flex items-center gap-2 text-foreground">
              <ShieldCheck className="size-4 shrink-0 text-amber-600 dark:text-amber-400" />
              <span className="text-sm font-semibold">{t('admin.maintenanceSafetyTitle')}</span>
            </div>
            <ul className="mt-3 space-y-2 text-sm leading-snug text-muted-foreground">
              {safetyItems.map((item) => (
                <li key={item} className="flex gap-2.5">
                  <span className="mt-1.5 size-1 shrink-0 rounded-full bg-amber-500/80 dark:bg-amber-400/80" />
                  <span>{item}</span>
                </li>
              ))}
            </ul>
          </section>

          <section className="rounded-xl border border-border/60 bg-muted/20 p-4">
            <div className="flex items-center gap-2 text-foreground">
              <HardDrive className="size-4 shrink-0 text-muted-foreground" />
              <span className="text-sm font-semibold">{t('admin.maintenanceBackupLocationTitle')}</span>
            </div>
            <p className="mt-2 text-sm leading-snug text-muted-foreground">{t('admin.maintenanceBackupLocationDesc')}</p>
            <div className="mt-3 space-y-1.5 rounded-lg border border-border/50 bg-background/60 px-2.5 py-2 font-mono text-[11px] leading-snug text-foreground sm:text-xs">
              <div className="flex flex-wrap items-baseline gap-x-2 gap-y-0.5">
                <span className="shrink-0 text-muted-foreground">{t('admin.maintenancePathInContainer')}</span>
                <span className="min-w-0 break-all">/app/data/backups</span>
              </div>
              <div className="flex flex-wrap items-baseline gap-x-2 gap-y-0.5 border-t border-border/40 pt-1.5">
                <span className="shrink-0 text-muted-foreground">{t('admin.maintenancePathOnHost')}</span>
                <span className="min-w-0 break-all">./data/backups</span>
              </div>
            </div>
            <p className="mt-3 text-xs leading-snug text-muted-foreground">{t('admin.maintenanceDockerHint')}</p>
          </section>
        </div>
      </div>

      <div className="space-y-4 border-t border-border/40 pt-5">
        <div className="space-y-1.5">
          <div className="flex flex-wrap items-center gap-2">
            <h2 className="text-base font-semibold text-foreground">{t('admin.maintenanceWorkflowTitle')}</h2>
            <span className="inline-flex items-center gap-1 rounded-full border border-border/60 bg-muted/30 px-2 py-0.5 text-xs font-medium text-muted-foreground">
              <Terminal className="size-3.5" />
              {t('admin.maintenanceComposeBadge')}
            </span>
          </div>
          <p className="text-sm leading-snug text-muted-foreground">{t('admin.maintenanceCommandsDesc')}</p>
          <p className="text-xs leading-snug text-muted-foreground">{t('admin.maintenanceCommandHint')}</p>
        </div>

        <ol className="list-none space-y-3">
          {capabilityItems.map((item, index) => {
            const Icon = item.icon
            const step = index + 1
            const stepLabel = t('admin.maintenanceStep', { step })
            return (
              <li key={item.title}>
                <article className="overflow-hidden rounded-lg border border-border/50 bg-card/40 shadow-sm ring-1 ring-black/[0.03] dark:bg-card/25 dark:ring-white/[0.04]">
                  <div className="border-l-[3px] border-primary/70 px-3 py-3 sm:px-4">
                    <div className="flex min-w-0 gap-3">
                      <span className="sr-only">{stepLabel}</span>
                      <div className="flex size-8 shrink-0 items-center justify-center rounded-md bg-primary/12 text-primary" aria-hidden>
                        <Icon className="size-4" />
                      </div>
                      <div className="min-w-0 flex-1">
                        <h3 className="text-sm font-semibold leading-tight text-foreground">{item.title}</h3>
                        <p className="mt-1 text-xs leading-snug text-muted-foreground">{item.desc}</p>
                      </div>
                    </div>

                    <CommandTerminal value={item.command} onCopy={() => void onCopy(item.command)} />
                  </div>
                </article>
              </li>
            )
          })}
        </ol>
      </div>
    </section>
  )
}

function CommandTerminal({ value, onCopy }: { value: string; onCopy: () => void }) {
  const { t } = useTranslation()

  return (
    <div className="mt-2.5 flex min-w-0 items-stretch overflow-hidden rounded-md border border-zinc-700/45 bg-zinc-950 text-zinc-100 dark:border-zinc-600/35">
      <pre className="min-h-0 min-w-0 flex-1 overflow-x-auto px-2.5 py-2 font-mono text-[11px] leading-snug sm:text-xs">
        <code className="block whitespace-pre text-left">
          <CommandCode text={value} />
        </code>
      </pre>
      <div className="flex shrink-0 items-center border-l border-zinc-700/45 px-1 dark:border-zinc-600/35">
        <Button
          type="button"
          variant="secondary"
          size="icon-sm"
          className="border-0 bg-transparent text-zinc-300 shadow-none hover:bg-white/10 hover:text-white"
          onClick={onCopy}
          title={t('common.copy')}
        >
          <Copy className="size-4" />
          <span className="sr-only">{t('common.copy')}</span>
        </Button>
      </div>
    </div>
  )
}

function CommandCode({ text }: { text: string }) {
  const token = SERVICE_PLACEHOLDER
  if (!text.includes(token)) {
    return <span className="text-zinc-100">{text}</span>
  }
  const parts = text.split(token)
  return (
    <>
      {parts.map((part, i) => (
        <Fragment key={i}>
          <span className="text-zinc-100">{part}</span>
          {i < parts.length - 1 ? (
            <span className="rounded px-0.5 font-semibold text-amber-300" title={token}>
              {token}
            </span>
          ) : null}
        </Fragment>
      ))}
    </>
  )
}
