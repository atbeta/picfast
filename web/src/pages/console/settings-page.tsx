import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { z } from 'zod/v4'
import { zodResolver } from '@hookform/resolvers/zod'

import { useAuth } from '../../lib/auth-context'
import { useTheme } from '../../lib/theme'
import { formatFileSize } from '../../lib/upload'

const profileSchema = z.object({
  name: z.string().min(1),
  password: z.string().optional(),
})
type ProfileForm = z.infer<typeof profileSchema>

export function SettingsPage() {
  const { t } = useTranslation()
  const { user, updateProfile } = useAuth()

  const [saving, setSaving] = useState(false)
  const [success, setSuccess] = useState(false)
  const [errorMsg, setErrorMsg] = useState('')

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ProfileForm>({
    resolver: zodResolver(profileSchema),
    defaultValues: { name: user?.name ?? '', password: '' },
  })

  const onSubmit = async (data: ProfileForm) => {
    setSaving(true)
    setSuccess(false)
    setErrorMsg('')
    try {
      const payload: { name?: string; password?: string } = { name: data.name }
      if (data.password && data.password.length >= 8) {
        payload.password = data.password
      }
      await updateProfile(payload)
      setSuccess(true)
    } catch (err: unknown) {
      setErrorMsg(
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
        t('settings.saveFailed'),
      )
    } finally {
      setSaving(false)
    }
  }

  if (!user) return null

  const usagePercent = user.capacity_bytes > 0 ? Math.round((user.used_bytes / user.capacity_bytes) * 100) : 0

  return (
    <section className="space-y-8">
      <h1 className="text-xl font-semibold">{t('page.settings.title')}</h1>

      {/* Storage usage */}
      <div className="rounded-lg border border-zinc-200 p-4 dark:border-zinc-700">
        <h2 className="text-sm font-medium">{t('settings.storage')}</h2>
        <div className="mt-3 flex items-center gap-3">
          <div className="h-2 flex-1 overflow-hidden rounded-full bg-zinc-200 dark:bg-zinc-700">
            <div
              className="h-full rounded-full bg-zinc-600 dark:bg-zinc-400"
              style={{ width: `${Math.min(usagePercent, 100)}%` }}
            />
          </div>
          <span className="shrink-0 text-xs text-zinc-500">
            {formatFileSize(user.used_bytes)} / {formatFileSize(user.capacity_bytes)}
          </span>
        </div>
        <p className="mt-2 text-xs text-zinc-400">
          {t('settings.stats', { images: user.image_num, albums: user.album_num })}
        </p>
      </div>

      {/* Profile form */}
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div>
          <label htmlFor="email" className="mb-1 block text-sm font-medium">{t('settings.email')}</label>
          <input
            id="email"
            type="email"
            value={user.email}
            disabled
            className="w-full rounded-lg border border-zinc-200 bg-zinc-100 px-3 py-2 text-sm text-zinc-500 dark:border-zinc-700 dark:bg-zinc-800"
          />
        </div>

        <div>
          <label htmlFor="name" className="mb-1 block text-sm font-medium">{t('settings.name')}</label>
          <input
            id="name"
            type="text"
            className="w-full rounded-lg border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 dark:border-zinc-700 dark:bg-zinc-800 dark:focus:border-zinc-500"
            {...register('name')}
          />
          {errors.name && <p className="mt-1 text-xs text-red-500">{t('auth.required')}</p>}
        </div>

        <div>
          <label htmlFor="password" className="mb-1 block text-sm font-medium">{t('settings.newPassword')}</label>
          <input
            id="password"
            type="password"
            autoComplete="new-password"
            placeholder={t('settings.passwordHint')}
            className="w-full rounded-lg border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 dark:border-zinc-700 dark:bg-zinc-800 dark:focus:border-zinc-500"
            {...register('password')}
          />
          {errors.password && <p className="mt-1 text-xs text-red-500">{t('auth.passwordMin')}</p>}
        </div>

        {success && (
          <p className="rounded-lg bg-green-50 px-3 py-2 text-sm text-green-600 dark:bg-green-900/20 dark:text-green-400">
            {t('settings.saved')}
          </p>
        )}
        {errorMsg && (
          <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-400">
            {errorMsg}
          </p>
        )}

        <button
          type="submit"
          disabled={saving}
          className="rounded-lg bg-zinc-900 px-4 py-2 text-sm font-medium text-white transition hover:bg-zinc-700 disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-300"
        >
          {saving ? t('settings.saving') : t('settings.save')}
        </button>
      </form>

      {/* Preferences */}
      <div className="rounded-lg border border-zinc-200 p-4 dark:border-zinc-700">
        <h2 className="mb-3 text-sm font-medium">{t('settings.preferences')}</h2>
        <div className="flex flex-wrap gap-6">
          <LanguageSelector />
          <ThemeSelector />
        </div>
      </div>
    </section>
  )
}

function LanguageSelector() {
  const { i18n, t } = useTranslation()
  return (
    <label className="flex items-center gap-2 text-sm">
      <span className="text-zinc-600 dark:text-zinc-300">{t('common.language')}</span>
      <select
        value={i18n.language}
        onChange={(e) => void i18n.changeLanguage(e.target.value)}
        className="rounded-md border border-zinc-300 bg-white px-2 py-1 text-sm dark:border-zinc-700 dark:bg-zinc-900"
      >
        <option value="zh-CN">中文</option>
        <option value="en-US">English</option>
      </select>
    </label>
  )
}

function ThemeSelector() {
  const { t } = useTranslation()
  const { theme, setTheme } = useTheme()
  return (
    <label className="flex items-center gap-2 text-sm">
      <span className="text-zinc-600 dark:text-zinc-300">{t('common.theme')}</span>
      <select
        value={theme}
        onChange={(e) => setTheme(e.target.value as 'light' | 'dark' | 'system')}
        className="rounded-md border border-zinc-300 bg-white px-2 py-1 text-sm dark:border-zinc-700 dark:bg-zinc-900"
      >
        <option value="light">{t('settings.themeLight')}</option>
        <option value="dark">{t('settings.themeDark')}</option>
        <option value="system">{t('settings.themeSystem')}</option>
      </select>
    </label>
  )
}
