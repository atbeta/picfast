import { useEffect, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'

import { getSiteConfig } from '@/lib/site-config'
import { mergeThemeConfig, themeConfigToCSS } from '@/lib/theme-config'
import { useTheme } from '@/lib/theme'

export function SiteThemeRuntime() {
  const { setTheme } = useTheme()
  const { data: config } = useQuery({ queryKey: ['site-config'], queryFn: getSiteConfig })
  const themeConfig = useMemo(() => mergeThemeConfig(config?.theme_config), [config?.theme_config])
  const css = useMemo(() => themeConfigToCSS(themeConfig), [themeConfig])

  useEffect(() => {
    const root = document.documentElement
    const pub = themeConfig.public
    root.dataset.pfTheme = themeConfig.preset || 'default'
    root.dataset.pfThemeBackground = pub?.background_style || 'soft'
    root.dataset.pfThemeLogo = pub?.logo_shape || 'rounded'
    root.dataset.pfUploadStyle = pub?.upload_style || 'dashed'
    root.dataset.pfCardStyle = pub?.card_style || 'elevated'
    root.dataset.pfButtonStyle = pub?.button_style || 'default'
    root.dataset.pfDensity = pub?.density || 'comfortable'
    root.dataset.pfMotion = pub?.motion || 'subtle'
  }, [themeConfig])

  useEffect(() => {
    if (!themeConfig.mode || themeConfig.mode === 'system') return
    if (window.localStorage.getItem('theme')) return
    setTheme(themeConfig.mode)
  }, [setTheme, themeConfig.mode])

  return <style id="picfast-site-theme">{css}</style>
}
