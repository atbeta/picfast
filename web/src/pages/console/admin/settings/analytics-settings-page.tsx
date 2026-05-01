import { useTranslation } from 'react-i18next'
import { Controller, useWatch } from 'react-hook-form'

import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import {
  analyticsPayload,
  fieldInputCls,
  fieldTextareaCls,
  useAdminSettingsForm,
} from './form'
import {
  SettingField,
  SettingsPageLayout,
} from './shared'

export function AdminAnalyticsSettingsPage() {
  const { t } = useTranslation()
  const state = useAdminSettingsForm()
  const { register, control, handleSubmit } = state.form
  const analyticsProvider = useWatch({ control, name: 'analytics_provider' }) || ''

  const onSubmit = handleSubmit((form) => state.saveSettings({
    analytics_provider: form.analytics_provider,
    analytics_config: analyticsPayload(form),
  }))

  return (
    <SettingsPageLayout
      title={t('admin.analyticsSettingsTitle')}
      description={t('admin.analyticsSettingsDesc')}
      state={state}
      onSubmit={onSubmit}
    >
      <SettingField label={t('admin.analyticsProvider')}>
        <Controller
          name="analytics_provider"
          control={control}
          render={({ field }) => (
            <Select
              value={field.value || ''}
              onValueChange={(val) => field.onChange(String(val))}
              items={{
                '': t('admin.analyticsDisabled'),
                plausible: 'Plausible',
                umami: 'Umami',
                ga4: 'Google Analytics 4',
                baidu: t('admin.analyticsBaidu'),
                custom: t('admin.analyticsCustom'),
              }}
            >
              <SelectTrigger className="h-11 w-full bg-background border-input">
                <SelectValue placeholder={t('admin.analyticsDisabled')} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">{t('admin.analyticsDisabled')}</SelectItem>
                <SelectItem value="plausible">Plausible</SelectItem>
                <SelectItem value="umami">Umami</SelectItem>
                <SelectItem value="ga4">Google Analytics 4</SelectItem>
                <SelectItem value="baidu">{t('admin.analyticsBaidu')}</SelectItem>
                <SelectItem value="custom">{t('admin.analyticsCustom')}</SelectItem>
              </SelectContent>
            </Select>
          )}
        />
      </SettingField>

      {analyticsProvider === 'plausible' && (
        <>
          <SettingField label={t('admin.analyticsDomain')}>
            <input {...register('analytics_domain')} placeholder="img.example.com" className={fieldInputCls} />
          </SettingField>
          <SettingField label={t('admin.analyticsScriptUrl')}>
            <input {...register('analytics_script_url')} placeholder="https://plausible.io/js/script.js" className={fieldInputCls} />
          </SettingField>
        </>
      )}

      {analyticsProvider === 'umami' && (
        <>
          <SettingField label={t('admin.analyticsScriptUrl')}>
            <input {...register('analytics_script_url')} placeholder="https://analytics.example.com/script.js" className={fieldInputCls} />
          </SettingField>
          <SettingField label={t('admin.analyticsWebsiteId')}>
            <input {...register('analytics_website_id')} className={fieldInputCls} />
          </SettingField>
        </>
      )}

      {analyticsProvider === 'ga4' && (
        <SettingField label={t('admin.analyticsMeasurementId')}>
          <input {...register('analytics_measurement_id')} placeholder="G-XXXXXXXXXX" className={fieldInputCls} />
        </SettingField>
      )}

      {analyticsProvider === 'baidu' && (
        <SettingField label={t('admin.analyticsSiteId')}>
          <input {...register('analytics_site_id')} className={fieldInputCls} />
        </SettingField>
      )}

      {analyticsProvider === 'custom' && (
        <SettingField label={t('admin.analyticsCustomScript')} hint={t('admin.analyticsCustomScriptDesc')}>
          <textarea {...register('analytics_custom_script')} className={fieldTextareaCls} />
        </SettingField>
      )}
    </SettingsPageLayout>
  )
}
