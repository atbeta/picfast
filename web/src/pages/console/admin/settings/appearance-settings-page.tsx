import { useTranslation } from 'react-i18next'

import { fieldTextareaCls, themePayload, useAdminSettingsForm } from './form'
import { SettingsPageLayout } from './shared'

export function AdminAppearanceSettingsPage() {
  const { t } = useTranslation()
  const state = useAdminSettingsForm()
  const { handleSubmit, register } = state.form

  const onSubmit = handleSubmit((form) =>
    state.saveSettings({
      theme_config: themePayload(form),
    }),
  )

  return (
    <SettingsPageLayout
      title={t('admin.appearanceSettingsTitle')}
      description={t('admin.appearanceSettingsDesc')}
      state={state}
      onSubmit={onSubmit}
    >
      <div className="space-y-3">
        <div>
          <label
            htmlFor="theme_custom_css"
            className="block text-sm font-medium text-foreground"
          >
            {t('admin.themeCustomCss')}
          </label>
          <p className="mt-1 text-sm text-muted-foreground">
            {t('admin.themeCustomCssHint')}
          </p>
        </div>
        <textarea
          id="theme_custom_css"
          {...register('theme_custom_css')}
          spellCheck={false}
          className={`${fieldTextareaCls} min-h-80 font-mono`}
          placeholder={':root {\n  --primary: oklch(0.62 0.18 260);\n  --radius: 0.75rem;\n}'}
        />
        <p className="text-xs leading-5 text-muted-foreground">
          {t('admin.themeCustomCssScope')}
        </p>
      </div>
    </SettingsPageLayout>
  )
}
