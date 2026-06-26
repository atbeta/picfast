import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { copyToClipboard } from '@/lib/clipboard'
import { themeConfigToCSS, themePresets } from '@/lib/theme-config'
import {
  parseThemePackage,
  presetDefaultFormFields,
  serializeThemePackage,
  themeConfigToFormFields,
  type ThemeFormImportFields,
  ThemePackageError,
} from '@/lib/theme-package'
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
  const { handleSubmit, register, setValue, watch, getValues, formState: { errors } } = state.form
  const [importText, setImportText] = useState('')
  const preset = watch('theme_preset')
  const mode = watch('theme_mode')
  const backgroundStyle = watch('theme_background_style')
  const logoShape = watch('theme_logo_shape')
  const themePrimary = watch('theme_primary')
  const themeAccent = watch('theme_accent')
  const themeRadius = watch('theme_radius')
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

  const applyThemeFields = (fields: ThemeFormImportFields) => {
    for (const [key, value] of Object.entries(fields) as [keyof ThemeFormImportFields, string][]) {
      setValue(key, value, { shouldDirty: true })
    }
  }

  const applyPreset = (id: string) => {
    applyThemeFields(presetDefaultFormFields(id))
  }

  const previewCSS = useMemo(() => {
    return themeConfigToCSS(themePayload({
      ...getValues(),
      theme_preset: preset,
      theme_mode: mode,
      theme_primary: themePrimary,
      theme_accent: themeAccent,
      theme_radius: themeRadius,
      theme_background_style: backgroundStyle,
      theme_logo_shape: logoShape,
    }))
  }, [backgroundStyle, getValues, logoShape, mode, preset, themeAccent, themePrimary, themeRadius])

  const exportCurrentTheme = () => serializeThemePackage(themePayload(getValues()), 'site')

  const copyThemeJSON = async () => {
    await copyToClipboard(exportCurrentTheme())
    toast.success(t('admin.themeCopied'))
  }

  const downloadThemeJSON = () => {
    const blob = new Blob([exportCurrentTheme()], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = `picfast-theme-${preset || 'default'}.json`
    anchor.click()
    URL.revokeObjectURL(url)
    toast.success(t('admin.themeExported'))
  }

  const importThemeJSON = () => {
    try {
      const { config } = parseThemePackage(importText)
      applyThemeFields(themeConfigToFormFields(config))
      toast.success(t('admin.themeImported'))
    } catch (err: unknown) {
      const message = err instanceof ThemePackageError ? err.message : t('admin.themeImportFailed')
      toast.error(message)
    }
  }

  const resetThemeToPreset = () => {
    applyPreset(preset || 'default')
    setImportText('')
    toast.success(t('admin.themeReset'))
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

      <div className="border-t border-border/40 pt-6">
        <div className="mb-4">
          <h3 className="text-sm font-semibold text-foreground">{t('admin.sectionThemePreview')}</h3>
          <p className="mt-0.5 text-xs text-muted-foreground">{t('admin.sectionThemePreviewDesc')}</p>
        </div>
        <div className="overflow-hidden rounded-xl border border-border/60 bg-background">
          <style>{previewCSS}</style>
          <div className="space-y-4 p-6">
            <div className="flex items-center gap-3">
              <div className="pf-site-logo h-10 w-10 rounded-xl border border-border bg-muted" />
              <div>
                <div className="text-sm font-semibold">{t('admin.themePreviewTitle')}</div>
                <div className="text-xs text-muted-foreground">{t('admin.themePreviewSubtitle')}</div>
              </div>
            </div>
            <div className="rounded-xl border border-dashed border-primary/30 bg-card/60 p-6 text-center text-sm text-muted-foreground">
              {t('admin.themePreviewUpload')}
            </div>
            <div className="flex flex-wrap gap-2">
              <span className="inline-flex h-9 items-center rounded-lg bg-primary px-4 text-sm font-medium text-primary-foreground">
                {t('admin.themePreviewPrimaryButton')}
              </span>
              <span className="inline-flex h-9 items-center rounded-lg border border-border bg-background px-4 text-sm font-medium">
                {t('admin.themePreviewSecondaryButton')}
              </span>
            </div>
          </div>
        </div>
      </div>

      <div className="border-t border-border/40 pt-6">
        <div className="mb-4">
          <h3 className="text-sm font-semibold text-foreground">{t('admin.sectionThemePackage')}</h3>
          <p className="mt-0.5 text-xs text-muted-foreground">{t('admin.sectionThemePackageDesc')}</p>
        </div>
        <div className="space-y-4">
          <div className="flex flex-wrap gap-2">
            <Button type="button" variant="outline" onClick={copyThemeJSON}>
              {t('admin.themeCopyJson')}
            </Button>
            <Button type="button" variant="outline" onClick={downloadThemeJSON}>
              {t('admin.themeExportJson')}
            </Button>
            <Button type="button" variant="outline" onClick={resetThemeToPreset}>
              {t('admin.themeResetPreset')}
            </Button>
          </div>
          <textarea
            value={importText}
            onChange={(event) => setImportText(event.target.value)}
            spellCheck={false}
            placeholder={t('admin.themeImportPlaceholder')}
            className={`${fieldTextareaCls} min-h-40 font-mono`}
          />
          <div className="flex flex-wrap gap-2">
            <Button type="button" onClick={importThemeJSON}>
              {t('admin.themeApplyImport')}
            </Button>
          </div>
          <p className="text-xs leading-5 text-muted-foreground">{t('admin.themeImportHint')}</p>
        </div>
      </div>
    </SettingsPageLayout>
  )
}
