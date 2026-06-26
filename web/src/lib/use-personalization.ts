import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'

import { useAuth } from './auth-context'
import { getSiteConfig } from './site-config'
import {
  getEffectiveTheme,
  getEffectiveMode,
  getEffectiveDensity,
  getEffectiveMotion,
  getEffectiveCopyPreferences,
  getEffectiveWorkflow,
  DENSITY_VALUES,
  MOTION_VALUES,
  type PersonalizationState,
  type ThemeOverride,
} from './personalization'
import { defaultThemeConfig, mergeThemeConfig } from './theme-config'

function safeThemeOverride(raw: Record<string, unknown> | null | undefined): ThemeOverride | null {
  if (!raw) return null
  const o = raw as Record<string, unknown>
  const override: ThemeOverride = {}
  if (typeof o.preset === 'string' && o.preset.trim()) override.preset = o.preset.trim()
  if (o.mode === 'light' || o.mode === 'dark' || o.mode === 'system') override.mode = o.mode
  if (o.density === 'compact' || o.density === 'comfortable' || o.density === 'spacious') override.density = o.density
  if (o.motion === 'none' || o.motion === 'subtle' || o.motion === 'playful') override.motion = o.motion
  return Object.keys(override).length > 0 ? override : null
}

export function usePersonalization(): PersonalizationState {
  const { user } = useAuth()
  const { data: site } = useQuery({ queryKey: ['site-config'], queryFn: getSiteConfig })

  const userName = user?.name ?? ''
  const userEmail = user?.email ?? ''
  const userRole = user?.role ?? 'user'
  const userSettings = user?.settings as Record<string, unknown> | undefined

  return useMemo(() => {
    const themeOverride = safeThemeOverride(userSettings?.theme_override as Record<string, unknown> | undefined)
    const siteThemeConfig = site?.theme_config ?? null

    // ── theme ──
    const { config, preset } = getEffectiveTheme(siteThemeConfig, themeOverride)
    const merged = mergeThemeConfig(config)

    // ── mode ──
    const colorMode = getEffectiveMode(
      siteThemeConfig?.mode ?? defaultThemeConfig.mode,
      themeOverride,
      'system',
    )
    const density = getEffectiveDensity(siteThemeConfig, themeOverride)
    const motion = getEffectiveMotion(siteThemeConfig, themeOverride)

    // ── output ──
    const output = getEffectiveCopyPreferences(
      site?.default_copy_format,
      site?.copy_template,
      (userSettings?.default_copy_format as string),
      (userSettings?.copy_template as string),
    )

    // ── workflow ──
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
        preset,
        config: merged,
        tokens: {
          light: merged.tokens?.light ?? {},
          dark: merged.tokens?.dark ?? {},
        },
        customCSS: merged.custom_css ?? '',
        logoShape: merged.public?.logo_shape ?? defaultThemeConfig.public?.logo_shape ?? 'rounded',
        backgroundImage: merged.public?.background_image ?? '',
        backgroundStyle: merged.public?.background_style ?? defaultThemeConfig.public?.background_style ?? 'soft',
      },
      mode: {
        colorMode,
        density,
        motion,
        densityValue: DENSITY_VALUES[density] ?? 1,
        motionDuration: MOTION_VALUES[motion] ?? '150ms',
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
  }, [site, userSettings, userName, userEmail, userRole])
}
