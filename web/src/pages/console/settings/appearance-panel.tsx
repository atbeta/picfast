import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { useAuth } from '@/lib/auth-context'
import { usePersonalization } from '@/lib/use-personalization'
import { extractErrorMessage } from '@/lib/error-handler'
import { themePresets } from '@/lib/theme-config'
import { serializeThemePackage, parseThemePackage, ThemePackageError } from '@/lib/theme-package'
import { copyToClipboard } from '@/lib/clipboard'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

export function AppearancePanel() {
  const { t } = useTranslation()
  const { user, updateProfile } = useAuth()
  const { theme, mode } = usePersonalization()
  const [saving, setSaving] = useState(false)

  const currentSettings = (user?.settings as Record<string, unknown>) ?? {}
  const currentOverride = (currentSettings.theme_override as Record<string, unknown>) ?? {}

  const [preset, setPreset] = useState(String(currentOverride.preset || theme.preset))
  const [colorMode, setColorMode] = useState(String(currentOverride.mode || mode.colorMode || 'system'))
  const [density, setDensity] = useState(String(currentOverride.density || mode.density || 'comfortable'))
  const [motion, setMotion] = useState(String(currentOverride.motion || mode.motion || 'subtle'))
  const [importText, setImportText] = useState('')

  if (!user) return null

  const buildThemeConfig = () => ({
    preset,
    mode: colorMode as 'light' | 'dark' | 'system',
    public: {
      density: density as 'compact' | 'comfortable' | 'spacious',
      motion: motion as 'none' | 'subtle' | 'playful',
    },
  })

  const exportCurrentTheme = () => serializeThemePackage(buildThemeConfig(), 'user')
  const copyThemeJSON = async () => {
    await copyToClipboard(exportCurrentTheme())
    toast.success(t('admin.themeCopied'))
  }
  const downloadThemeJSON = () => {
    const blob = new Blob([exportCurrentTheme()], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = `picfast-theme-${preset}-user.json`
    anchor.click()
    URL.revokeObjectURL(url)
    toast.success(t('admin.themeExported'))
  }
  const importThemeJSON = () => {
    try {
      const { config } = parseThemePackage(importText)
      if (config.preset) setPreset(config.preset)
      if (config.mode) setColorMode(config.mode)
      if (config.public?.density) setDensity(config.public.density)
      if (config.public?.motion) setMotion(config.public.motion)
      setImportText('')
      toast.success(t('admin.themeImported'))
    } catch (err: unknown) {
      const message = err instanceof ThemePackageError ? err.message : t('admin.themeImportFailed')
      toast.error(message)
    }
  }

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSaving(true)
    try {
      const override: Record<string, string> = {}
      if (preset && preset !== theme.preset) override.preset = preset
      if (colorMode && colorMode !== 'system') override.mode = colorMode
      if (density && density !== 'comfortable') override.density = density
      if (motion && motion !== 'subtle') override.motion = motion

      await updateProfile({
        settings: {
          ...currentSettings,
          theme_override: Object.keys(override).length > 0 ? override : undefined,
        },
      })
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
          {t('settings.appearance', { defaultValue: '外观偏好' })}
        </h2>
        <p className="text-sm text-muted-foreground">
          {t('settings.appearanceDesc', { defaultValue: '选择你喜欢的主题预设和交互风格，覆盖站点默认。' })}
        </p>
      </div>

      <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
        {themePresets.map((item) => (
          <button
            key={item.id}
            type="button"
            onClick={() => setPreset(item.id)}
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

      <form onSubmit={onSubmit}>
        <div className="rounded-xl border border-border bg-card p-6 shadow-sm space-y-6">
          <div className="grid gap-3 md:grid-cols-[180px_minmax(0,1fr)] md:items-start md:gap-6">
            <div className="pt-1">
              <p className="text-sm font-medium text-foreground">{t('admin.themeMode')}</p>
            </div>
            <Select value={colorMode} onValueChange={(val) => setColorMode(String(val ?? 'system'))} items={{ system: t('settings.themeSystem'), light: t('settings.themeLight'), dark: t('settings.themeDark') }}>
              <SelectTrigger className="h-10 w-full sm:w-56 bg-background border-input"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="system">{t('settings.themeSystem')}</SelectItem>
                <SelectItem value="light">{t('settings.themeLight')}</SelectItem>
                <SelectItem value="dark">{t('settings.themeDark')}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="grid gap-3 md:grid-cols-[180px_minmax(0,1fr)] md:items-start md:gap-6">
            <div className="pt-1">
              <p className="text-sm font-medium text-foreground">{t('admin.themeDensity', { defaultValue: '界面密度' })}</p>
            </div>
            <Select value={density} onValueChange={(val) => setDensity(String(val ?? 'comfortable'))} items={{ compact: t('admin.themeDensityCompact', { defaultValue: '紧凑' }), comfortable: t('admin.themeDensityComfortable', { defaultValue: '舒适' }), spacious: t('admin.themeDensitySpacious', { defaultValue: '宽松' }) }}>
              <SelectTrigger className="h-10 w-full sm:w-56 bg-background border-input"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="compact">{t('admin.themeDensityCompact', { defaultValue: '紧凑' })}</SelectItem>
                <SelectItem value="comfortable">{t('admin.themeDensityComfortable', { defaultValue: '舒适' })}</SelectItem>
                <SelectItem value="spacious">{t('admin.themeDensitySpacious', { defaultValue: '宽松' })}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="grid gap-3 md:grid-cols-[180px_minmax(0,1fr)] md:items-start md:gap-6">
            <div className="pt-1">
              <p className="text-sm font-medium text-foreground">{t('admin.themeMotion', { defaultValue: '动画效果' })}</p>
            </div>
            <Select value={motion} onValueChange={(val) => setMotion(String(val ?? 'subtle'))} items={{ none: t('admin.themeMotionNone', { defaultValue: '关闭' }), subtle: t('admin.themeMotionSubtle', { defaultValue: '微妙' }), playful: t('admin.themeMotionPlayful', { defaultValue: '活泼' }) }}>
              <SelectTrigger className="h-10 w-full sm:w-56 bg-background border-input"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="none">{t('admin.themeMotionNone', { defaultValue: '关闭' })}</SelectItem>
                <SelectItem value="subtle">{t('admin.themeMotionSubtle', { defaultValue: '微妙' })}</SelectItem>
                <SelectItem value="playful">{t('admin.themeMotionPlayful', { defaultValue: '活泼' })}</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
        <div className="flex justify-end mt-6">
          <Button type="submit" size="lg" disabled={saving}>
            {saving ? t('settings.saving') : t('settings.save')}
          </Button>
        </div>
      </form>

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
          </div>
          <textarea
            value={importText}
            onChange={(event) => setImportText(event.target.value)}
            spellCheck={false}
            placeholder={t('admin.themeImportPlaceholder')}
            className="min-h-24 w-full rounded-lg border border-input bg-background px-3 py-2 text-sm font-mono outline-none transition-colors duration-150 focus:border-primary focus:ring-2 focus:ring-primary/20"
          />
          <div className="flex flex-wrap gap-2">
            <Button type="button" onClick={importThemeJSON}>
              {t('admin.themeApplyImport')}
            </Button>
          </div>
          <p className="text-xs leading-5 text-muted-foreground">{t('admin.themeImportHint')}</p>
        </div>
      </div>
    </div>
  )
}
