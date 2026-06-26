import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { useAuth } from '@/lib/auth-context'
import { usePersonalization } from '@/lib/use-personalization'
import { extractErrorMessage } from '@/lib/error-handler'
import { themePresets } from '@/lib/theme-config'
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

  if (!user) return null

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
    </div>
  )
}
