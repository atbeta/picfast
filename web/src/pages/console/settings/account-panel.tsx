import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { z } from 'zod/v4'
import { zodResolver } from '@hookform/resolvers/zod'

import { useAuth } from '@/lib/auth-context'
import { formatFileSize } from '@/lib/upload'
import { extractErrorMessage } from '@/lib/error-handler'
import { Button } from '@/components/ui/button'
import { HelpHint } from '@/components/help-hint'

const profileSchema = z.object({
  name: z.string().min(1),
  password: z.string().optional(),
})
type ProfileForm = z.infer<typeof profileSchema>

const fieldInputCls = 'h-10 w-full rounded-lg border border-border/50 bg-background/50 px-4 text-sm outline-none transition-colors duration-150 placeholder:text-muted-foreground/50 focus:border-primary focus:ring-1 focus:ring-primary/20'
const fieldDisabledCls = 'h-10 w-full rounded-lg border border-border/50 bg-muted/50 px-4 text-sm text-muted-foreground cursor-not-allowed'

export function AccountPanel() {
  const { t } = useTranslation()
  const { user, updateProfile } = useAuth()
  const [saving, setSaving] = useState(false)

  const { register, handleSubmit, formState: { errors } } = useForm<ProfileForm>({
    resolver: zodResolver(profileSchema),
    defaultValues: { name: user?.name ?? '', password: '' },
  })

  if (!user) return null

  const usagePercent = user.capacity_bytes > 0 ? Math.round((user.used_bytes / user.capacity_bytes) * 100) : 0
  const isUnlimitedCapacity = user.capacity_bytes <= 0

  const onSubmit = async (data: ProfileForm) => {
    setSaving(true)
    try {
      const payload: { name?: string; password?: string } = { name: data.name }
      if (data.password && data.password.length >= 8) {
        payload.password = data.password
      }
      await updateProfile(payload)
      toast.success(t('settings.saved'))
    } catch (err: unknown) {
      toast.error(extractErrorMessage(err, t('settings.saveFailed')))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-base font-semibold tracking-tight text-foreground">
          {t('settings.storage', { defaultValue: '存储用量' })}
        </h2>
        <p className="text-sm text-muted-foreground">
          {t('settings.storageDesc', { defaultValue: '查看您当前账号的可用空间和已使用情况。' })}
        </p>
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

      <div>
        <h2 className="text-base font-semibold tracking-tight text-foreground">
          {t('settings.profile', { defaultValue: '个人资料' })}
        </h2>
        <p className="text-sm text-muted-foreground">
          {t('settings.profileDesc', { defaultValue: '管理您的基础信息。' })}
        </p>
      </div>
      <form onSubmit={handleSubmit(onSubmit)}>
        <div className="rounded-xl border border-border bg-card p-6 shadow-sm">
          <div className="space-y-6">
            <div className="grid gap-3 md:grid-cols-[180px_minmax(0,1fr)] md:items-start md:gap-6">
              <div className="pt-1">
                <p className="text-sm font-medium text-foreground">{t('settings.email')}</p>
              </div>
              <input value={user.email} disabled className={fieldDisabledCls} />
            </div>

            <div className="grid gap-3 md:grid-cols-[180px_minmax(0,1fr)] md:items-start md:gap-6">
              <div className="pt-1">
                <p className="text-sm font-medium text-foreground">{t('settings.name')}</p>
              </div>
              <div className="min-w-0">
                <input
                  type="text"
                  placeholder={t('settings.profileNamePlaceholder', { defaultValue: '输入您的昵称' })}
                  className={fieldInputCls}
                  {...register('name')}
                />
                {errors.name && <p className="mt-1.5 text-xs text-destructive">{t('auth.required')}</p>}
              </div>
            </div>

            <div className="grid gap-3 md:grid-cols-[180px_minmax(0,1fr)] md:items-start md:gap-6">
              <div className="pt-1 flex items-center gap-2">
                <p className="text-sm font-medium text-foreground">{t('settings.newPassword')}</p>
                <HelpHint text={t('settings.passwordHint', { defaultValue: '不修改请留空' })} />
              </div>
              <div className="min-w-0">
                <input
                  type="password"
                  autoComplete="new-password"
                  placeholder={t('settings.profilePasswordPlaceholder', { defaultValue: '留空表示不修改' })}
                  className={fieldInputCls}
                  {...register('password')}
                />
                {errors.password && <p className="mt-1.5 text-xs text-destructive">{t('auth.passwordMin')}</p>}
              </div>
            </div>
          </div>
        </div>

        <div className="flex justify-end mt-6">
          <Button type="submit" size="lg" disabled={saving}>
            {saving ? t('settings.saving') : t('settings.save')}
          </Button>
        </div>
      </form>
    </div>
  )
}
