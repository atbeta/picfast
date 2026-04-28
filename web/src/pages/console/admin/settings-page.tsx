import { useForm, Controller } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'

import { getAdminSettings, updateAdminSettings } from '../../../lib/admin-api'
import { useState } from 'react'
import { Switch } from '@/components/ui/switch'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

interface SettingsForm {
  app_name: string
  app_url: string
  allow_guest_upload: boolean
  allow_registration: boolean
  user_initial_capacity_mb: number
  moderation_mode: string
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
        user_initial_capacity: form.user_initial_capacity_mb * 1024 * 1024,
        moderation_mode: form.moderation_mode,
      })
      setSuccess(true)
      await qc.invalidateQueries({ queryKey: ['admin-settings'] })
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

      {isLoading && <div className="flex justify-center py-12"><div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" /></div>}
      {error && <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">{t('admin.loadFailed')}</p>}

      {data && (
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6 max-w-2xl">
          <div className="space-y-4 rounded-xl border border-border bg-card p-6 shadow-sm">
            <div className="space-y-2">
              <label className="text-sm font-medium text-foreground">{t('admin.appName')}</label>
              <input {...register('app_name')} className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all" />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-foreground">{t('admin.appUrl', { defaultValue: '站点地址' })}</label>
              <input {...register('app_url')} placeholder="https://your-domain.com" className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all" />
            </div>

            <div className="flex items-center justify-between rounded-lg border border-border bg-background p-4 shadow-sm">
              <div className="space-y-0.5">
                <label htmlFor="guestUpload" className="text-sm font-medium text-foreground cursor-pointer">{t('admin.allowGuestUpload')}</label>
                <p className="text-xs text-muted-foreground">{t('admin.allowGuestUploadDesc', { defaultValue: '允许未登录访客上传图片' })}</p>
              </div>
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

            <div className="flex items-center justify-between rounded-lg border border-border bg-background p-4 shadow-sm">
              <div className="space-y-0.5">
                <label htmlFor="registration" className="text-sm font-medium text-foreground cursor-pointer">{t('admin.allowRegistration')}</label>
                <p className="text-xs text-muted-foreground">{t('admin.allowRegistrationDesc', { defaultValue: '开放新用户注册通道' })}</p>
              </div>
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

            <div className="space-y-2">
              <label className="text-sm font-medium text-foreground">{t('admin.initialCapacity')}</label>
              <div className="flex items-center gap-2">
                <input type="number" {...register('user_initial_capacity_mb', { valueAsNumber: true })} className="w-32 rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all" />
                <span className="text-sm text-muted-foreground">MB</span>
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-foreground">{t('admin.moderationMode')}</label>
              <Controller
                name="moderation_mode"
                control={control}
                render={({ field }) => (
                  <Select 
              value={field.value} 
              onValueChange={(val) => val !== null && field.onChange(val as string)}
              items={{
                      disabled: t('admin.modDisabled'),
                      manual: t('admin.modManual'),
                      auto: t('admin.modAuto')
                    }}
                  >
                    <SelectTrigger className="w-full bg-background border-input">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="disabled">{t('admin.modDisabled')}</SelectItem>
                      <SelectItem value="manual">{t('admin.modManual')}</SelectItem>
                      <SelectItem value="auto">{t('admin.modAuto')}</SelectItem>
                    </SelectContent>
                  </Select>
                )}
              />
            </div>
          </div>

          {success && <p className="rounded-lg bg-success/10 px-3 py-2 text-sm text-success">{t('admin.saved')}</p>}
          {errorMsg && <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">{errorMsg}</p>}

          <button type="submit" disabled={isSubmitting} className="rounded-lg bg-primary px-6 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50 transition-colors shadow-sm cursor-pointer">
            {isSubmitting ? t('admin.saving') : t('admin.save')}
          </button>
        </form>
      )}
    </section>
  )
}
