import { useEffect, useMemo } from 'react'

import { usePersonalization } from '@/lib/use-personalization'
import { themeConfigToCSS } from '@/lib/theme-config'
import { useTheme } from '@/lib/theme'

export function SiteThemeRuntime() {
  const { setTheme } = useTheme()
  const { theme, mode } = usePersonalization()

  const css = useMemo(() => themeConfigToCSS(theme.config), [theme.config])

  useEffect(() => {
    const root = document.documentElement
    root.dataset.pfTheme = theme.preset || 'default'
    root.dataset.pfThemeBackground = theme.backgroundStyle || 'soft'
    root.dataset.pfThemeLogo = theme.logoShape || 'rounded'
    root.dataset.pfUploadStyle = theme.config.public?.upload_style || 'dashed'
    root.dataset.pfCardStyle = theme.config.public?.card_style || 'elevated'
    root.dataset.pfButtonStyle = theme.config.public?.button_style || 'default'
    root.dataset.pfDensity = mode.density || 'comfortable'
    root.dataset.pfMotion = mode.motion || 'subtle'

    root.style.setProperty('--pf-density', String(mode.densityValue))
    root.style.setProperty('--pf-motion-duration', mode.motionDuration)
  }, [theme, mode])

  useEffect(() => {
    if (!mode.colorMode || mode.colorMode === 'system') return
    if (window.localStorage.getItem('theme')) return
    setTheme(mode.colorMode)
  }, [setTheme, mode.colorMode])

  return <style id="picfast-site-theme">{css}</style>
}
