import { useTranslation } from 'react-i18next'

import { fieldInputCls, useAdminSettingsForm } from './form'
import {
  SettingField,
  SettingsPageLayout,
} from './shared'

export function AdminComplianceSettingsPage() {
  const { t } = useTranslation()
  const state = useAdminSettingsForm()
  const { register, handleSubmit } = state.form

  const onSubmit = handleSubmit((form) => state.saveSettings({
    icp_number: form.icp_number,
    psb_number: form.psb_number,
  }))

  return (
    <SettingsPageLayout
      title={t('admin.complianceSettingsTitle')}
      description={t('admin.complianceSettingsDesc')}
      state={state}
      onSubmit={onSubmit}
    >
      <SettingField label={t('admin.icpNumber')} hint={t('admin.icpNumberDesc')}>
        <input {...register('icp_number')} className={fieldInputCls} />
      </SettingField>

      <SettingField label={t('admin.psbNumber')} hint={t('admin.psbNumberDesc')}>
        <input {...register('psb_number')} className={fieldInputCls} />
      </SettingField>
    </SettingsPageLayout>
  )
}
