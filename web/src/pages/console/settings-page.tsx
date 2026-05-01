import type { ReactNode } from 'react'
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { z } from 'zod/v4'
import { zodResolver } from '@hookform/resolvers/zod'

import { useAuth } from '../../lib/auth-context'
import { useTheme } from '../../lib/theme'
import { formatFileSize } from '../../lib/upload'
import { getStrategies, type Strategy } from '../../lib/console-api'
import { extractErrorMessage } from '../../lib/error-handler'
import { storageStrategyLabel } from '../../lib/storage-strategy'

import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { HelpHint } from '@/components/help-hint'

const profileSchema = z.object({
  name: z.string().min(1),
  password: z.string().optional(),
  defaultStrategy: z.number().optional(),
})
type ProfileForm = z.infer<typeof profileSchema>

const fieldInputCls = 'h-11 w-full rounded-lg border border-border/50 bg-background/50 px-4 text-sm outline-none transition-colors duration-150 placeholder:text-muted-foreground/50 focus:border-primary focus:ring-1 focus:ring-primary/20'
const fieldDisabledCls = 'h-11 w-full rounded-lg border border-border/50 bg-muted/50 px-4 text-sm text-muted-foreground cursor-not-allowed'

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

export function SettingsPage() {
  const { t } = useTranslation()
  const { user, updateProfile } = useAuth()

  const [saving, setSaving] = useState(false)
  const [success, setSuccess] = useState(false)
  const [errorMsg, setErrorMsg] = useState('')
  const [strategies, setStrategies] = useState<Strategy[]>([])
  const [defaultStrategy, setDefaultStrategy] = useState(0)

  useEffect(() => {
    getStrategies()
      .then((list) => {
        setStrategies(list)
        const settings = (user?.settings as Record<string, unknown>) || {}
        if (settings.default_strategy) {
          setDefaultStrategy(Number(settings.default_strategy))
        }
      })
      .catch(() => {})
  }, [user])

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
      const payload: { name?: string; password?: string; settings?: Record<string, unknown> } = { name: data.name }
      if (data.password && data.password.length >= 8) {
        payload.password = data.password
      }
      payload.settings = defaultStrategy && defaultStrategy !== 0
        ? { default_strategy: defaultStrategy }
        : { default_strategy: null }
      await updateProfile(payload)
      setSuccess(true)
    } catch (err: unknown) {
      setErrorMsg(extractErrorMessage(err, t('settings.saveFailed')))
    } finally {
      setSaving(false)
    }
  }

  if (!user) return null

  const usagePercent = user.capacity_bytes > 0 ? Math.round((user.used_bytes / user.capacity_bytes) * 100) : 0
  const isUnlimitedCapacity = user.capacity_bytes <= 0

  return (
    <section className="space-y-6">
      <h1 className="text-2xl font-bold tracking-tight">{t('page.settings.title')}</h1>

      <div className="max-w-4xl space-y-6 pb-8">
        
        {/* Section 1: Storage usage */}
        <div className="space-y-6">
          <div>
            <h2 className="text-base font-semibold tracking-tight text-foreground">{t('settings.storage', { defaultValue: '存储用量' })}</h2>
            <p className="text-sm text-muted-foreground">{t('settings.storageDesc', { defaultValue: '查看您当前账号的可用空间和已使用情况。' })}</p>
          </div>
          <div className="rounded-xl border border-border bg-card p-6 shadow-sm">
            <div className="flex items-center gap-4">
              <div className="h-2.5 flex-1 overflow-hidden rounded-full bg-muted">
                <div
                  className="h-full rounded-full bg-primary transition-[width] duration-200 ease-out"
                  style={{ width: `${Math.min(usagePercent, 100)}%` }}
                />
              </div>
              <span className="shrink-0 text-sm font-medium text-muted-foreground">
                {formatFileSize(user.used_bytes)} / {isUnlimitedCapacity ? t('settings.unlimitedCapacity', { defaultValue: '无限制' }) : formatFileSize(user.capacity_bytes)}
              </span>
            </div>
            <p className="mt-3 text-sm text-muted-foreground/80">
              {t('settings.stats', { images: user.image_num, albums: user.album_num })}
            </p>
          </div>
        </div>

        {/* Section 2: Profile form */}
        <div className="space-y-6">
          <div className="pt-4 border-t border-border/40">
            <h2 className="text-base font-semibold tracking-tight text-foreground">{t('settings.profile', { defaultValue: '个人资料' })}</h2>
            <p className="text-sm text-muted-foreground">{t('settings.profileDesc', { defaultValue: '管理您的基础信息与上传偏好。' })}</p>
          </div>
          <form onSubmit={handleSubmit(onSubmit)} className="rounded-xl border border-border bg-card p-6 shadow-sm">
            <div className="space-y-6">
              <SettingField
                label={t('settings.email')}
              >
                <input
                  id="email"
                  type="email"
                  value={user.email}
                  disabled
                  className={fieldDisabledCls}
                />
              </SettingField>

              <SettingField
                label={t('settings.name')}
              >
                <input
                  id="name"
                  type="text"
                  placeholder={t('settings.profileNamePlaceholder', { defaultValue: '输入您的昵称' })}
                  className={fieldInputCls}
                  {...register('name')}
                />
                {errors.name && <p className="mt-1.5 text-xs text-destructive">{t('auth.required')}</p>}
              </SettingField>

              <SettingField
                label={t('settings.newPassword')}
                hint={t('settings.passwordHint', { defaultValue: '不修改请留空' })}
              >
                <input
                  id="password"
                  type="password"
                  autoComplete="new-password"
                  placeholder={t('settings.profilePasswordPlaceholder', { defaultValue: '留空表示不修改' })}
                  className={fieldInputCls}
                  {...register('password')}
                />
                {errors.password && <p className="mt-1.5 text-xs text-destructive">{t('auth.passwordMin')}</p>}
              </SettingField>

              <SettingField
                label={t('settings.defaultStrategy', { defaultValue: '默认策略' })}
                hint={t('settings.defaultStrategyDesc', { defaultValue: '上传时默认选中的存储策略。' })}
              >
                <Select
                  value={defaultStrategy.toString()}
                  onValueChange={(val) => val !== null && setDefaultStrategy(Number(val))}
                  items={{
                    '0': t('settings.followGroupDefault', { defaultValue: '跟随分组默认' }),
                    ...Object.fromEntries(strategies.map(s => [s.id.toString(), `${s.name} (${storageStrategyLabel(t, s.strategy_type)})`]))
                  }}
                >
                  <SelectTrigger id="strategy" className="h-11 w-full bg-background border-input md:max-w-md">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="0">{t('settings.followGroupDefault', { defaultValue: '跟随分组默认' })}</SelectItem>
                    {strategies.map((s) => (
                      <SelectItem key={s.id} value={s.id.toString()}>
                        {s.name} ({storageStrategyLabel(t, s.strategy_type)})
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </SettingField>
            </div>

            <div className="pt-2">
              {success && (
                <p className="mt-6 mb-2 rounded-lg bg-success/10 px-4 py-2.5 text-sm font-medium text-success-foreground border border-success/20">
                  {t('settings.saved')}
                </p>
              )}
              {errorMsg && (
                <p className="mt-6 mb-2 rounded-lg bg-destructive/10 px-4 py-2.5 text-sm font-medium text-destructive border border-destructive/20">
                  {errorMsg}
                </p>
              )}

              <div className="flex justify-end mt-6">
                <button
                  type="submit"
                  disabled={saving}
                  className="rounded-lg bg-primary px-6 py-2 text-sm font-medium text-primary-foreground shadow-sm transition-opacity duration-150 hover:opacity-90 disabled:opacity-50 cursor-pointer"
                >
                  {saving ? t('settings.saving') : t('settings.save')}
                </button>
              </div>
            </div>
          </form>
        </div>

        {/* Section 3: Preferences */}
        <div className="space-y-6">
          <div className="pt-4 border-t border-border/40">
            <h2 className="text-base font-semibold tracking-tight text-foreground">{t('settings.preferences', { defaultValue: '偏好设置' })}</h2>
            <p className="text-sm text-muted-foreground">{t('settings.preferencesDesc', { defaultValue: '自定义您在控制台的显示语言和主题外观。' })}</p>
          </div>
          <div className="rounded-xl border border-border bg-card p-6 shadow-sm">
            <div className="flex flex-col gap-6 md:flex-row md:items-center md:gap-12">
              <LanguageSelector />
              <ThemeSelector />
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}

function LanguageSelector() {
  const { i18n, t } = useTranslation()
  return (
    <div className="flex items-center justify-between gap-3 text-sm">
      <span className="font-medium text-foreground">{t('common.language')}</span>
      <Select
          value={i18n.language}
          onValueChange={(val) => val !== null && void i18n.changeLanguage(val as string)}
          items={{
            'zh-CN': '中文',
          'en-US': 'English'
        }}
      >
        <SelectTrigger className="w-[120px] bg-background/50 border-border/50">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="zh-CN">中文</SelectItem>
          <SelectItem value="en-US">English</SelectItem>
        </SelectContent>
      </Select>
    </div>
  )
}

function ThemeSelector() {
  const { t } = useTranslation()
  const { theme, setTheme } = useTheme()
  return (
    <div className="flex items-center justify-between gap-3 text-sm">
      <span className="font-medium text-foreground">{t('common.theme')}</span>
      <Select
          value={theme}
          onValueChange={(val) => val !== null && setTheme(val as 'light' | 'dark' | 'system')}
          items={{
            light: t('settings.themeLight'),
          dark: t('settings.themeDark'),
          system: t('settings.themeSystem')
        }}
      >
        <SelectTrigger className="w-[140px] bg-background/50 border-border/50">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="light">{t('settings.themeLight')}</SelectItem>
          <SelectItem value="dark">{t('settings.themeDark')}</SelectItem>
          <SelectItem value="system">{t('settings.themeSystem')}</SelectItem>
        </SelectContent>
      </Select>
    </div>
  )
}
