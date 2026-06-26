import { useEffect, useMemo } from 'react'

import { usePersonalization } from '@/lib/use-personalization'
import { themeConfigToCSS } from '@/lib/theme-config'
import { useTheme } from '@/lib/theme'

const LOGO_RADIUS: Record<string, string> = {
  circle: '9999px',
  square: 'calc(var(--radius) * 0.35)',
  rounded: '0.75rem',
}

export function SiteThemeRuntime() {
  const { setTheme } = useTheme()
  const { theme, mode } = usePersonalization()

  const css = useMemo(() => themeConfigToCSS(theme.config), [theme.config])

  useEffect(() => {
    const root = document.documentElement
    root.dataset.pfTheme = theme.preset || 'default'
    root.dataset.pfThemeBackground = theme.backgroundStyle || 'soft'
    root.dataset.pfThemeLogo = theme.logoShape || 'rounded'
    root.dataset.pfDensity = mode.density || 'comfortable'
    root.dataset.pfMotion = mode.motion || 'subtle'

    root.style.setProperty('--pf-density', String(mode.densityValue))
    root.style.setProperty('--pf-motion-duration', mode.motionDuration)
    root.style.setProperty('--pf-public-glow-opacity', theme.backgroundStyle === 'image' ? '0.45' : '1')
    root.style.setProperty('--pf-logo-radius', LOGO_RADIUS[theme.logoShape] ?? '0.75rem')
  }, [theme, mode])

  useEffect(() => {
    if (!mode.colorMode || mode.colorMode === 'system') return
    if (window.localStorage.getItem('theme')) return
    setTheme(mode.colorMode)
  }, [setTheme, mode.colorMode])

  return <style id="picfast-site-theme">{css}</style>
}
