import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'

import { useAuth } from './auth-context'
import { getSiteConfig } from './site-config'
import { getEffectiveWorkflow, type PersonalizationState } from './personalization'

export function usePersonalization(): PersonalizationState {
  const { user } = useAuth()
  const { data: site } = useQuery({ queryKey: ['site-config'], queryFn: getSiteConfig })

  const userName = user?.name ?? ''
  const userEmail = user?.email ?? ''
  const userRole = user?.role ?? 'user'
  const userSettings = user?.settings as Record<string, unknown> | undefined
  const siteTheme = site?.theme_config as { custom_css?: string; mode?: 'light' | 'dark' | 'system' } | null

  return useMemo(() => {
    const workflow = getEffectiveWorkflow(
      userSettings,
      site?.allow_user_image_processing ?? true,
      site?.skip_image_processing ?? false,
      localStorage.getItem('default_strategy_id'),
      localStorage.getItem('default_album_id'),
      localStorage.getItem('default_permission'),
    )

    return {
      theme: {
        customCSS: siteTheme?.custom_css ?? '',
      },
      mode: {
        colorMode: siteTheme?.mode ?? 'system',
        language: (userSettings?.language as string) || localStorage.getItem('i18nextLng') || 'en-US',
      },
      workflow,
      account: {
        name: userName,
        email: userEmail,
        role: userRole,
      },
    }
  }, [site, siteTheme?.custom_css, siteTheme?.mode, userSettings, userName, userEmail, userRole])
}
