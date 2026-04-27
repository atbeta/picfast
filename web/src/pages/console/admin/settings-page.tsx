import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'

import { getAdminSettings, updateAdminSettings } from '../../../lib/admin-api'
import { useState } from 'react'

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

  const { register, handleSubmit, formState: { isSubmitting } } = useForm<SettingsForm>({
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
    <section className="space-y-4">
      <h1 className="text-xl font-semibold">{t('admin.settingsTitle')}</h1>

      {isLoading && <div className="flex justify-center py-12"><div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-400 border-t-transparent" /></div>}
      {error && <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-400">{t('admin.loadFailed')}</p>}

      {data && (
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div>
            <label className="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">{t('admin.appName')}</label>
            <input {...register('app_name')} className="w-full max-w-sm rounded-lg border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 dark:border-zinc-700 dark:bg-zinc-800" />
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">{t('admin.appUrl', { defaultValue: '站点地址' })}</label>
            <input {...register('app_url')} placeholder="https://your-domain.com" className="w-full max-w-sm rounded-lg border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 dark:border-zinc-700 dark:bg-zinc-800" />
          </div>

          <div className="flex items-center gap-2">
            <input type="checkbox" id="guestUpload" {...register('allow_guest_upload')} className="h-4 w-4 rounded border-zinc-300 dark:border-zinc-600" />
            <label htmlFor="guestUpload" className="text-sm text-zinc-700 dark:text-zinc-300">{t('admin.allowGuestUpload')}</label>
          </div>

          <div className="flex items-center gap-2">
            <input type="checkbox" id="registration" {...register('allow_registration')} className="h-4 w-4 rounded border-zinc-300 dark:border-zinc-600" />
            <label htmlFor="registration" className="text-sm text-zinc-700 dark:text-zinc-300">{t('admin.allowRegistration')}</label>
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">{t('admin.initialCapacity')}</label>
            <div className="flex items-center gap-2">
              <input type="number" {...register('user_initial_capacity_mb', { valueAsNumber: true })} className="w-32 rounded-lg border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 dark:border-zinc-700 dark:bg-zinc-800" />
              <span className="text-sm text-zinc-500 dark:text-zinc-400">MB</span>
            </div>
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">{t('admin.moderationMode')}</label>
            <select {...register('moderation_mode')} className="rounded-md border border-zinc-300 bg-white px-3 py-2 text-sm dark:border-zinc-700 dark:bg-zinc-900">
              <option value="disabled">{t('admin.modDisabled')}</option>
              <option value="manual">{t('admin.modManual')}</option>
              <option value="auto">{t('admin.modAuto')}</option>
            </select>
          </div>

          {success && <p className="rounded-lg bg-green-50 px-3 py-2 text-sm text-green-600 dark:bg-green-900/20 dark:text-green-400">{t('admin.saved')}</p>}
          {errorMsg && <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-400">{errorMsg}</p>}

          <button type="submit" disabled={isSubmitting} className="rounded-lg bg-zinc-900 px-4 py-2 text-sm font-medium text-white hover:bg-zinc-700 disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-300">
            {isSubmitting ? t('admin.saving') : t('admin.save')}
          </button>
        </form>
      )}
    </section>
  )
}
