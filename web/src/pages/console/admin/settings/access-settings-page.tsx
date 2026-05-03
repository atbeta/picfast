import { useTranslation } from 'react-i18next'
import { Controller } from 'react-hook-form'
import { AlertCircle } from 'lucide-react'

import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { fieldInputCls, useAdminSettingsForm } from './form'
import {
  SettingField,
  SettingsPageLayout,
} from './shared'

export function AdminAccessSettingsPage() {
  const { t } = useTranslation()
  const state = useAdminSettingsForm()
  const { data, form } = state
  const { register, control, handleSubmit } = form

  const onSubmit = handleSubmit((values) => state.saveSettings({
    allow_guest_upload: Boolean(values.allow_guest_upload),
    guest_capacity_bytes: values.guest_capacity_mb * 1024 * 1024,
    allow_registration: Boolean(values.allow_registration),
    allow_user_image_processing: Boolean(values.allow_user_image_processing),
    require_email_verification: Boolean(values.require_email_verification),
    user_initial_capacity: values.user_initial_capacity_mb * 1024 * 1024,
    default_image_ttl: values.default_image_ttl,
    guest_image_ttl: values.guest_image_ttl,
    moderation_mode: values.moderation_mode,
  }))

  return (
    <SettingsPageLayout
      title={t('admin.accessSettingsTitle')}
      description={t('admin.accessSettingsDesc')}
      state={state}
      onSubmit={onSubmit}
    >
      <SettingField label={t('admin.allowGuestUpload')}>
        <div className="flex h-11 items-center justify-end">
          <Controller
            name="allow_guest_upload"
            control={control}
            render={({ field }) => (
              <Switch checked={field.value} onCheckedChange={field.onChange} id="guestUpload" />
            )}
          />
        </div>
      </SettingField>

      <SettingField label={t('admin.guestCapacity')} hint={t('admin.guestCapacityDesc')}>
        <div className="flex items-center gap-3">
          <input
            type="number"
            min={0}
            step={1}
            {...register('guest_capacity_mb', { valueAsNumber: true, min: 0 })}
            className={`${fieldInputCls} w-40`}
          />
          <span className="text-sm text-muted-foreground">MB</span>
        </div>
      </SettingField>

      <SettingField label={t('admin.allowRegistration')}>
        <div className="flex h-11 items-center justify-end">
          <Controller
            name="allow_registration"
            control={control}
            render={({ field }) => (
              <Switch checked={field.value} onCheckedChange={field.onChange} id="registration" />
            )}
          />
        </div>
      </SettingField>

      <SettingField
        label={t('admin.allowUserImageProcessing', { defaultValue: '允许用户自定义图片处理' })}
        hint={t('admin.allowUserImageProcessingDesc', { defaultValue: '关闭后用户侧上传统一使用系统默认处理（质量85、去EXIF、保持原格式、水印关闭）。' })}
      >
        <div className="flex h-11 items-center justify-end">
          <Controller
            name="allow_user_image_processing"
            control={control}
            render={({ field }) => (
              <Switch checked={field.value} onCheckedChange={field.onChange} id="allowUserImageProcessing" />
            )}
          />
        </div>
      </SettingField>

      <SettingField label={t('admin.requireEmailVerification')} hint={t('admin.requireEmailVerificationDesc')}>
        <div className="flex h-11 items-center justify-end">
          <Controller
            name="require_email_verification"
            control={control}
            render={({ field }) => (
              <Switch
                checked={field.value}
                onCheckedChange={field.onChange}
                id="requireEmailVerification"
                disabled={!data?.email_verification_ready}
              />
            )}
          />
        </div>
      </SettingField>

      {data && !data.email_verification_ready && (
        <div className="flex items-start gap-3 rounded-xl bg-amber-500/10 px-4 py-3 text-sm text-amber-600 dark:bg-amber-500/10 dark:text-amber-400">
          <div className="mt-0.5 shrink-0">
            <AlertCircle className="size-4" />
          </div>
          <div className="space-y-1">
            <p className="font-medium">{t('admin.smtpNotConfiguredTitle')}</p>
            <p className="text-xs opacity-80">{t('admin.smtpNotConfiguredDesc')}</p>
          </div>
        </div>
      )}

      <SettingField label={t('admin.initialCapacity')} hint={t('admin.initialCapacityDesc')}>
        <div className="flex items-center gap-3">
          <input type="number" {...register('user_initial_capacity_mb', { valueAsNumber: true })} className={`${fieldInputCls} w-40`} />
          <span className="text-sm text-muted-foreground">MB</span>
        </div>
      </SettingField>

      <SettingField label={t('admin.defaultImageTTL')} hint={t('admin.defaultImageTTLDesc')}>
        <Controller
          name="default_image_ttl"
          control={control}
          render={({ field }) => (
            <Select
              value={field.value || '0'}
              items={{
                '0': t('admin.ttlNever'),
                '24h': t('admin.ttl1Day'),
                '168h': t('admin.ttl7Days'),
                '720h': t('admin.ttl30Days'),
                '2160h': t('admin.ttl90Days'),
              }}
              onValueChange={(val) => field.onChange(String(val))}
            >
              <SelectTrigger className="h-11 w-full bg-background border-input">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="0">{t('admin.ttlNever')}</SelectItem>
                <SelectItem value="24h">{t('admin.ttl1Day')}</SelectItem>
                <SelectItem value="168h">{t('admin.ttl7Days')}</SelectItem>
                <SelectItem value="720h">{t('admin.ttl30Days')}</SelectItem>
                <SelectItem value="2160h">{t('admin.ttl90Days')}</SelectItem>
              </SelectContent>
            </Select>
          )}
        />
      </SettingField>

      <SettingField label={t('admin.guestImageTTL')} hint={t('admin.guestImageTTLDesc')}>
        <Controller
          name="guest_image_ttl"
          control={control}
          render={({ field }) => (
            <Select
              value={field.value || '0'}
              items={{
                '0': t('admin.ttlGuestDefault'),
                '24h': t('admin.ttl1Day'),
                '168h': t('admin.ttl7Days'),
                '720h': t('admin.ttl30Days'),
                '2160h': t('admin.ttl90Days'),
              }}
              onValueChange={(val) => field.onChange(String(val))}
            >
              <SelectTrigger className="h-11 w-full bg-background border-input">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="0">{t('admin.ttlGuestDefault')}</SelectItem>
                <SelectItem value="24h">{t('admin.ttl1Day')}</SelectItem>
                <SelectItem value="168h">{t('admin.ttl7Days')}</SelectItem>
                <SelectItem value="720h">{t('admin.ttl30Days')}</SelectItem>
                <SelectItem value="2160h">{t('admin.ttl90Days')}</SelectItem>
              </SelectContent>
            </Select>
          )}
        />
      </SettingField>

      <SettingField label={t('admin.moderationMode')} hint={t('admin.moderationModeDesc')}>
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
    </SettingsPageLayout>
  )
}
