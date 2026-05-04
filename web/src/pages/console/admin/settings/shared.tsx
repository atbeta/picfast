import type { ReactNode } from 'react'
import type { UseFormReturn } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { HelpHint } from '@/components/help-hint'
import { LoadingState } from '@/components/page-states'

import type { SettingsForm, useAdminSettingsForm } from './form'

export function SettingField({
  label,
  hint,
  children,
}: {
  label: string
  hint?: string
  children: ReactNode
}) {
  return (
    <div className="grid gap-3 md:grid-cols-[180px_minmax(0,1fr)] md:items-start md:gap-6">
      <div className="pt-2">
        <div className="flex items-center gap-2">
          <p className="text-sm font-medium text-foreground">{label}</p>
          {hint ? <HelpHint text={hint} /> : null}
        </div>
      </div>
      <div className="min-w-0">{children}</div>
    </div>
  )
}

export function SettingsPageLayout({
  title,
  description,
  state,
  onSubmit,
  children,
}: {
  title: string
  description: string
  state: ReturnType<typeof useAdminSettingsForm>
  onSubmit: ReturnType<UseFormReturn<SettingsForm>['handleSubmit']>
  children: ReactNode
}) {
  const { t } = useTranslation()
  const { data, isLoading, error, form } = state
  const { formState: { isSubmitting } } = form

  return (
    <section className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">{title}</h1>
        <p className="mt-1 text-sm text-muted-foreground">{description}</p>
      </div>

      {isLoading && <LoadingState />}
      {error && <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">{t('admin.loadFailed')}</p>}

      {data && (
        <form onSubmit={onSubmit} className="space-y-6 pb-8">
          <div className="space-y-6 rounded-xl border border-border/40 bg-card/60 p-6 shadow-sm">
            {children}
          </div>

          <div className="flex justify-end pt-4">
            <Button type="submit" size="lg" disabled={isSubmitting}>
              {isSubmitting ? t('admin.saving') : t('admin.save')}
            </Button>
          </div>
        </form>
      )}
    </section>
  )
}
