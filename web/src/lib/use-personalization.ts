import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'

import { useAuth } from './auth-context'
import { getSiteConfig } from './site-config'
import {
  getEffectiveCopyPreferences,
  getEffectiveWorkflow,
  type PersonalizationState,
} from './personalization'

export function usePersonalization(): PersonalizationState {
  const { user } = useAuth()
  const { data: site } = useQuery({ queryKey: ['site-config'], queryFn: getSiteConfig })

  const userName = user?.name ?? ''
  const userEmail = user?.email ?? ''
  const userRole = user?.role ?? 'user'
  const userSettings = user?.settings as Record<string, unknown> | undefined
  const siteTheme = site?.theme_config as { custom_css?: string; mode?: 'light' | 'dark' | 'system' } | null

  return useMemo(() => {
    const output = getEffectiveCopyPreferences(
      site?.default_copy_format,
      site?.copy_template,
      (userSettings?.default_copy_format as string),
      (userSettings?.copy_template as string),
    )

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
      output,
      workflow,
      account: {
        name: userName,
        email: userEmail,
        role: userRole,
      },
    }
  }, [site, siteTheme?.custom_css, siteTheme?.mode, userSettings, userName, userEmail, userRole])
}
