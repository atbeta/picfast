import { useTranslation } from 'react-i18next'

import { fieldInputCls, fieldTextareaCls, useAdminSettingsForm } from './form'
import {
  SettingField,
  SettingsPageLayout,
} from './shared'

export function AdminSiteSettingsPage() {
  const { t } = useTranslation()
  const state = useAdminSettingsForm()
  const { register, handleSubmit } = state.form

  const onSubmit = handleSubmit((form) => state.saveSettings({
    app_name: form.app_name,
    app_url: form.app_url,
    site_description: form.site_description,
  }))

  return (
    <SettingsPageLayout
      title={t('admin.siteSettingsTitle')}
      description={t('admin.siteSettingsDesc')}
      state={state}
      onSubmit={onSubmit}
    >
      <SettingField label={t('admin.appName')}>
        <input {...register('app_name')} className={fieldInputCls} />
      </SettingField>

      <SettingField label={t('admin.appUrl')} hint={t('admin.appUrlDesc')}>
        <input {...register('app_url')} placeholder="https://your-domain.com" className={fieldInputCls} />
      </SettingField>

      <SettingField label={t('admin.siteDescription')} hint={t('admin.siteDescriptionDesc')}>
        <textarea
          {...register('site_description')}
          placeholder={t('admin.siteDescriptionPlaceholder')}
          className={fieldTextareaCls}
        />
      </SettingField>

    </SettingsPageLayout>
  )
}
