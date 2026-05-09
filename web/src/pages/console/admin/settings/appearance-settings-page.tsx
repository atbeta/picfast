import { useTranslation } from 'react-i18next'

import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { themePresets } from '@/lib/theme-config'
import { fieldInputCls, fieldTextareaCls, themePayload, useAdminSettingsForm } from './form'
import { SettingField, SettingsPageLayout } from './shared'

function validateCSSColor(value: string) {
  const raw = value.trim()
  if (!raw) return true
  if (/[;{}<>]/.test(raw)) return false
  return CSS.supports('color', raw)
}

function validateCSSRadius(value: string) {
  const raw = value.trim()
  if (!raw) return true
  if (/[;{}<>]/.test(raw)) return false
  return CSS.supports('border-radius', raw)
}

function validateOptionalHTTPURL(value: string) {
  const raw = value.trim()
  if (!raw) return true
  try {
    const url = new URL(raw)
    return url.protocol === 'http:' || url.protocol === 'https:'
  } catch {
    return false
  }
}

export function AdminAppearanceSettingsPage() {
  const { t } = useTranslation()
  const state = useAdminSettingsForm()
  const { handleSubmit, register, setValue, watch, formState: { errors } } = state.form
  const preset = watch('theme_preset')
  const mode = watch('theme_mode')
  const backgroundStyle = watch('theme_background_style')
  const logoShape = watch('theme_logo_shape')
  const modeItems = {
    system: t('settings.themeSystem'),
    light: t('settings.themeLight'),
    dark: t('settings.themeDark'),
  }
  const backgroundItems = {
    soft: t('admin.themeBackgroundSoft'),
    clean: t('admin.themeBackgroundClean'),
    image: t('admin.themeBackgroundImageMode'),
  }
  const logoItems = {
    rounded: t('admin.themeLogoRounded'),
    circle: t('admin.themeLogoCircle'),
    square: t('admin.themeLogoSquare'),
  }

  const applyPreset = (id: string) => {
    const next = themePresets.find((item) => item.id === id)
    setValue('theme_preset', id, { shouldDirty: true })
    setValue('theme_primary', '', { shouldDirty: true })
    setValue('theme_accent', '', { shouldDirty: true })
    setValue('theme_radius', '', { shouldDirty: true })
    setValue('theme_background_style', next?.config.public?.background_style || 'soft', { shouldDirty: true })
    setValue('theme_logo_shape', next?.config.public?.logo_shape || 'rounded', { shouldDirty: true })
  }

  const onSubmit = handleSubmit((form) => state.saveSettings({
    theme_config: themePayload(form),
  }))

  return (
    <SettingsPageLayout
      title={t('admin.appearanceSettingsTitle')}
      description={t('admin.appearanceSettingsDesc')}
      state={state}
      onSubmit={onSubmit}
    >
      <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
        {themePresets.map((item) => (
          <button
            key={item.id}
            type="button"
            onClick={() => applyPreset(item.id)}
            className={[
              'group rounded-lg border bg-background p-4 text-left transition-colors',
              preset === item.id ? 'border-primary ring-2 ring-primary/20' : 'border-border hover:border-primary/50',
            ].join(' ')}
          >
            <div className="mb-3 flex gap-1.5">
              <span className="h-5 w-5 rounded-full border border-border" style={{ background: item.config.tokens?.light?.primary || 'var(--primary)' }} />
              <span className="h-5 w-5 rounded-full border border-border" style={{ background: item.config.tokens?.light?.accent || 'var(--accent)' }} />
              <span className="h-5 w-5 rounded-full border border-border" style={{ background: item.config.tokens?.light?.background || 'var(--background)' }} />
            </div>
            <div className="text-sm font-semibold text-foreground">{t(item.nameKey)}</div>
            <div className="mt-1 text-xs leading-5 text-muted-foreground">{t(item.descriptionKey)}</div>
          </button>
        ))}
      </div>

      <div className="border-t border-border/40 pt-6">
        <div className="mb-4">
          <h3 className="text-sm font-semibold text-foreground">{t('admin.sectionThemeBasics')}</h3>
          <p className="mt-0.5 text-xs text-muted-foreground">{t('admin.sectionThemeBasicsDesc')}</p>
        </div>

        <div className="space-y-5">
          <SettingField label={t('admin.themeMode')} hint={t('admin.themeModeDesc')}>
            <Select items={modeItems} value={mode} onValueChange={(value) => setValue('theme_mode', String(value), { shouldDirty: true })}>
              <SelectTrigger className="w-full sm:w-56">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="system">{t('settings.themeSystem')}</SelectItem>
                <SelectItem value="light">{t('settings.themeLight')}</SelectItem>
                <SelectItem value="dark">{t('settings.themeDark')}</SelectItem>
              </SelectContent>
            </Select>
          </SettingField>

          <SettingField label={t('admin.themePrimary')} hint={t('admin.themeColorHint')}>
            <input
              {...register('theme_primary', {
                validate: (value) => validateCSSColor(value) || t('admin.themeInvalidColor'),
              })}
              placeholder="oklch(0.62 0.18 260) / #7c3aed"
              className={fieldInputCls}
            />
            {errors.theme_primary && <p className="mt-2 text-xs text-destructive">{errors.theme_primary.message}</p>}
          </SettingField>

          <SettingField label={t('admin.themeAccent')} hint={t('admin.themeColorHint')}>
            <input
              {...register('theme_accent', {
                validate: (value) => validateCSSColor(value) || t('admin.themeInvalidColor'),
              })}
              placeholder="oklch(0.9 0.07 20) / #fbcfe8"
              className={fieldInputCls}
            />
            {errors.theme_accent && <p className="mt-2 text-xs text-destructive">{errors.theme_accent.message}</p>}
          </SettingField>

          <SettingField label={t('admin.themeRadius')} hint={t('admin.themeRadiusHint')}>
            <input
              {...register('theme_radius', {
                validate: (value) => validateCSSRadius(value) || t('admin.themeInvalidRadius'),
              })}
              placeholder="0.75rem"
              className={fieldInputCls}
            />
            {errors.theme_radius && <p className="mt-2 text-xs text-destructive">{errors.theme_radius.message}</p>}
          </SettingField>
        </div>
      </div>

      <div className="border-t border-border/40 pt-6">
        <div className="mb-4">
          <h3 className="text-sm font-semibold text-foreground">{t('admin.sectionPublicSurface')}</h3>
          <p className="mt-0.5 text-xs text-muted-foreground">{t('admin.sectionPublicSurfaceDesc')}</p>
        </div>

        <div className="space-y-5">
          <SettingField label={t('admin.themeBackgroundImage')} hint={t('admin.themeBackgroundImageHint')}>
            <input
              {...register('theme_background_image', {
                validate: (value) => validateOptionalHTTPURL(value) || t('admin.themeInvalidURL'),
              })}
              placeholder="https://example.com/background.png"
              className={fieldInputCls}
            />
            {errors.theme_background_image && <p className="mt-2 text-xs text-destructive">{errors.theme_background_image.message}</p>}
          </SettingField>

          <SettingField label={t('admin.themeBackgroundStyle')}>
            <Select items={backgroundItems} value={backgroundStyle} onValueChange={(value) => setValue('theme_background_style', String(value), { shouldDirty: true })}>
              <SelectTrigger className="w-full sm:w-56">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="soft">{t('admin.themeBackgroundSoft')}</SelectItem>
                <SelectItem value="clean">{t('admin.themeBackgroundClean')}</SelectItem>
                <SelectItem value="image">{t('admin.themeBackgroundImageMode')}</SelectItem>
              </SelectContent>
            </Select>
          </SettingField>

          <SettingField label={t('admin.themeLogoShape')}>
            <Select items={logoItems} value={logoShape} onValueChange={(value) => setValue('theme_logo_shape', String(value), { shouldDirty: true })}>
              <SelectTrigger className="w-full sm:w-56">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="rounded">{t('admin.themeLogoRounded')}</SelectItem>
                <SelectItem value="circle">{t('admin.themeLogoCircle')}</SelectItem>
                <SelectItem value="square">{t('admin.themeLogoSquare')}</SelectItem>
              </SelectContent>
            </Select>
          </SettingField>
        </div>
      </div>

      <div className="border-t border-border/40 pt-6">
        <SettingField label={t('admin.themeCustomCss')} hint={t('admin.themeCustomCssHint')}>
          <div className="space-y-3">
            <textarea
              {...register('theme_custom_css')}
              spellCheck={false}
              className={`${fieldTextareaCls} min-h-48 font-mono`}
            />
            <p className="text-xs leading-5 text-muted-foreground">{t('admin.themeCustomCssScope')}</p>
          </div>
        </SettingField>
      </div>
    </SettingsPageLayout>
  )
}
