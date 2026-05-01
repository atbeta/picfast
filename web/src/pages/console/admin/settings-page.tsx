import type { ReactNode } from 'react'
import { useForm, Controller } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'

import { getAdminSettings, updateAdminSettings } from '../../../lib/admin-api'
import { useState } from 'react'
import { Switch } from '@/components/ui/switch'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { LoadingState } from '@/components/page-states'
import { HelpHint } from '@/components/help-hint'

interface SettingsForm {
  app_name: string
  app_url: string
  allow_guest_upload: boolean
  allow_registration: boolean
  require_email_verification: boolean
  user_initial_capacity_mb: number
  moderation_mode: string
}

const fieldInputCls = 'h-11 w-full rounded-lg border border-input bg-background px-4 text-sm outline-none transition-all focus:border-primary focus:ring-2 focus:ring-primary/20'

function SettingField({
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
      <div className="pt-1">
        <div className="flex items-center gap-2">
          <p className="text-sm font-medium text-foreground">{label}</p>
          {hint ? <HelpHint text={hint} /> : null}
        </div>
      </div>
      <div className="min-w-0">{children}</div>
    </div>
  )
}

export function AdminSettingsPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [success, setSuccess] = useState(false)
  const [errorMsg, setErrorMsg] = useState('')

  const { data, isLoading, error } = useQuery({
    queryKey: ['admin-settings'],
    queryFn: getAdminSettings,
  })

  const { register, control, handleSubmit, formState: { isSubmitting } } = useForm<SettingsForm>({
    values: data
      ? {
          app_name: data.app_name,
          app_url: data.app_url || '',
          allow_guest_upload: data.allow_guest_upload,
          allow_registration: data.allow_registration,
          require_email_verification: data.require_email_verification,
          user_initial_capacity_mb: Math.round(data.user_initial_capacity / 1024 / 1024),
          moderation_mode: data.moderation_mode,
        }
      : undefined,
  })

  const onSubmit = async (form: SettingsForm) => {
    setSuccess(false)
    setErrorMsg('')
    try {
      await updateAdminSettings({
        app_name: form.app_name,
        app_url: form.app_url,
        allow_guest_upload: form.allow_guest_upload,
        allow_registration: form.allow_registration,
        require_email_verification: form.require_email_verification,
        user_initial_capacity: form.user_initial_capacity_mb * 1024 * 1024,
        moderation_mode: form.moderation_mode,
      })
      setSuccess(true)
      await qc.invalidateQueries({ queryKey: ['admin-settings'] })
      await qc.invalidateQueries({ queryKey: ['site-config'] })
    } catch (err: unknown) {
      setErrorMsg(
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
        t('admin.saveFailed'),
      )
    }
  }

  return (
    <section className="space-y-6">
      <h1 className="text-2xl font-bold tracking-tight">{t('admin.settingsTitle')}</h1>

      {isLoading && <LoadingState />}
      {error && <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">{t('admin.loadFailed')}</p>}

      {data && (
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6 max-w-4xl">
          <div className="space-y-6 rounded-xl border border-border bg-card p-6 shadow-sm">
            <SettingField
              label={t('admin.appName')}
            >
              <input {...register('app_name')} className={fieldInputCls} />
            </SettingField>

            <SettingField
              label={t('admin.appUrl', { defaultValue: '站点地址' })}
              hint={t('admin.appUrlDesc', { defaultValue: '用于生成回调链接、ShareX 配置和邮箱验证链接。' })}
            >
              <input {...register('app_url')} placeholder="https://your-domain.com" className={fieldInputCls} />
            </SettingField>

            <SettingField label={t('admin.allowGuestUpload')}>
              <div className="flex h-11 items-center justify-end">
                <Controller
                  name="allow_guest_upload"
                  control={control}
                  render={({ field }) => (
                    <Switch
                      checked={field.value}
                      onCheckedChange={field.onChange}
                      id="guestUpload"
                    />
                  )}
                />
              </div>
            </SettingField>

            <SettingField label={t('admin.allowRegistration')}>
              <div className="flex h-11 items-center justify-end">
                <Controller
                  name="allow_registration"
                  control={control}
                  render={({ field }) => (
                    <Switch
                      checked={field.value}
                      onCheckedChange={field.onChange}
                      id="registration"
                    />
                  )}
                />
              </div>
            </SettingField>

            <SettingField
              label={t('admin.requireEmailVerification')}
              hint={t('admin.requireEmailVerificationDesc')}
            >
              <div className="space-y-3">
                <div className="flex h-11 items-center justify-end">
                  <Controller
                    name="require_email_verification"
                    control={control}
                    render={({ field }) => (
                      <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                        id="requireEmailVerification"
                        disabled={!data.email_verification_ready}
                      />
                    )}
                  />
                </div>
                <div className={`rounded-lg border px-4 py-3 text-xs ${data.email_verification_ready ? 'border-success/20 bg-success/5 text-success dark:border-success/30 dark:bg-success/10' : 'border-amber-200 bg-amber-50 text-amber-700 dark:border-amber-900/50 dark:bg-amber-950/20 dark:text-amber-300'}`}>
                  {data.email_verification_ready
                    ? t('admin.emailVerificationReady')
                    : t('admin.emailVerificationPending')}
                </div>
              </div>
            </SettingField>

            <SettingField
              label={t('admin.initialCapacity')}
              hint={t('admin.initialCapacityDesc', { defaultValue: '新注册用户默认可用的总容量。' })}
            >
              <div className="flex items-center gap-3">
                <input type="number" {...register('user_initial_capacity_mb', { valueAsNumber: true })} className={`${fieldInputCls} w-40`} />
                <span className="text-sm text-muted-foreground">MB</span>
              </div>
            </SettingField>

            <SettingField
              label={t('admin.moderationMode')}
              hint={t('admin.moderationModeDesc', { defaultValue: '控制上传内容是直接通过还是进入审核流程。' })}
            >
              <Controller
                name="moderation_mode"
                control={control}
                render={({ field }) => (
                  <Select
                    value={field.value || 'disabled'}
                    onValueChange={(val) => field.onChange(String(val))}
                    items={{
                      disabled: t('admin.modDisabled'),
                      manual: t('admin.modManual'),
                      auto: t('admin.modAuto'),
                    }}
                  >
                    <SelectTrigger className="h-11 w-full bg-background border-input">
                      <SelectValue placeholder={t('admin.modDisabled')} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="disabled">{t('admin.modDisabled')}</SelectItem>
                      <SelectItem value="manual">{t('admin.modManual')}</SelectItem>
                      <SelectItem value="auto">{t('admin.modAuto')}</SelectItem>
                    </SelectContent>
                  </Select>
                )}
              />
            </SettingField>
          </div>

          {success && <p className="rounded-lg bg-success/10 px-3 py-2 text-sm text-success">{t('admin.saved')}</p>}
          {errorMsg && <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">{errorMsg}</p>}

          <div className="flex justify-end pt-4">
            <button type="submit" disabled={isSubmitting} className="rounded-lg bg-primary px-6 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50 transition-colors shadow-sm cursor-pointer">
              {isSubmitting ? t('admin.saving') : t('admin.save')}
            </button>
          </div>
        </form>
      )}
    </section>
  )
}
