import { useEffect } from 'react'

import { useQuery } from '@tanstack/react-query'

import { useAuth } from '@/lib/auth-context'
import { getSiteConfig } from '@/lib/site-config'
import { useTheme } from '@/lib/theme'

export function SiteThemeRuntime() {
  const { setTheme } = useTheme()
  const { user } = useAuth()
  const { data: site } = useQuery({ queryKey: ['site-config'], queryFn: getSiteConfig })

  const customCSS = (site?.theme_config as { custom_css?: string } | null)?.custom_css ?? ''

  useEffect(() => {
    if (!user) return
    const mode = (site?.theme_config as { mode?: 'light' | 'dark' | 'system' } | null)?.mode
    if (mode && mode !== 'system' && !window.localStorage.getItem('theme')) {
      setTheme(mode)
    }
  }, [user, site, setTheme])

  if (!customCSS) return null
  return <style id="picfast-site-theme">{customCSS}</style>
}
