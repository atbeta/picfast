import type { ThemeConfig, ThemeMode, ThemeTokenSet } from './theme-config'
import { defaultThemeConfig, mergeThemeConfig } from './theme-config'
import type { CopyPreferences } from './copy-template'
import { normalizeCopyFormat } from './copy-template'

// ── Theme override (user-scoped, persisted in users.settings.theme_override) ─
export interface ThemeOverride {
  preset?: string
  mode?: ThemeMode
  density?: 'compact' | 'comfortable' | 'spacious'
  motion?: 'none' | 'subtle' | 'playful'
}

export const DENSITY_VALUES: Record<string, number> = {
  compact: 0.75,
  comfortable: 1,
  spacious: 1.25,
}

export const MOTION_VALUES: Record<string, string> = {
  none: '0ms',
  subtle: '150ms',
  playful: '300ms',
}

// ── Personalization state (the single source of truth) ─────────────
export interface PersonalizationState {
  theme: PersonalizationTheme
  mode: PersonalizationMode
  output: CopyPreferences
  workflow: PersonalizationWorkflow
  account: PersonalizationAccount
}

export interface PersonalizationTheme {
  preset: string
  config: ThemeConfig
  tokens: {
    light: ThemeTokenSet
    dark: ThemeTokenSet
  }
  customCSS: string
  logoShape: 'rounded' | 'circle' | 'square'
  backgroundImage: string
  backgroundStyle: 'soft' | 'clean' | 'image'
}

export interface PersonalizationMode {
  colorMode: ThemeMode
  density: 'compact' | 'comfortable' | 'spacious'
  motion: 'none' | 'subtle' | 'playful'
  densityValue: number
  motionDuration: string
  language: string
}

export interface PersonalizationWorkflow {
  defaultStrategy: number | null
  defaultAlbum: number | null
  defaultPermission: number | null
  allowUserImageProcessing: boolean
  skipImageProcessing: boolean
  imageProcessing: Record<string, unknown> | null
}

export interface PersonalizationAccount {
  name: string
  email: string
  role: string
}

// ── Effective theme computation ────────────────────────────────────
export function getEffectiveTheme(
  siteThemeConfig: ThemeConfig | null | undefined,
  userOverride: ThemeOverride | null | undefined,
): { config: ThemeConfig; preset: string } {
  const merged = mergeThemeConfig(siteThemeConfig ?? undefined)

  if (userOverride?.preset) {
    return {
      preset: userOverride.preset,
      config: mergeThemeConfig({ ...siteThemeConfig ?? {}, preset: userOverride.preset }),
    }
  }

  return {
    preset: merged.preset ?? defaultThemeConfig.preset ?? 'default',
    config: merged,
  }
}

// ── Mode computation ───────────────────────────────────────────────
export function getEffectiveMode(
  siteThemeMode: ThemeMode | undefined,
  userOverride: ThemeOverride | null | undefined,
  fallbackMode: ThemeMode,
): ThemeMode {
  return userOverride?.mode ?? siteThemeMode ?? fallbackMode
}

export function getEffectiveDensity(
  siteThemeConfig: ThemeConfig | null | undefined,
  userOverride: ThemeOverride | null | undefined,
): 'compact' | 'comfortable' | 'spacious' {
  const merged = mergeThemeConfig(siteThemeConfig ?? undefined)
  const siteDefault = merged.public?.density ?? 'comfortable'
  return userOverride?.density ?? siteDefault
}

export function getEffectiveMotion(
  siteThemeConfig: ThemeConfig | null | undefined,
  userOverride: ThemeOverride | null | undefined,
): 'none' | 'subtle' | 'playful' {
  const merged = mergeThemeConfig(siteThemeConfig ?? undefined)
  const siteDefault = merged.public?.motion ?? 'subtle'
  return userOverride?.motion ?? siteDefault
}

// ── Copy preferences computation ───────────────────────────────────
export function getEffectiveCopyPreferences(
  siteCopyFormat: string | null | undefined,
  siteCopyTemplate: string | null | undefined,
  userCopyFormat: string | null | undefined,
  userCopyTemplate: string | null | undefined,
): CopyPreferences {
  return {
    format: normalizeCopyFormat(userCopyFormat ?? siteCopyFormat),
    template: (userCopyTemplate ?? siteCopyTemplate ?? '').trim(),
  }
}

// ── Workflow computation ───────────────────────────────────────────
export function getEffectiveWorkflow(
  userSettings: Record<string, unknown> | null | undefined,
  siteAllowUserImageProcessing: boolean,
  siteSkipImageProcessing: boolean,
  localStorageStrategy: string | null,
  localStorageAlbum: string | null,
  localStoragePermission: string | null,
): PersonalizationWorkflow {
  const s = (userSettings as Record<string, unknown>) ?? {}

  return {
    defaultStrategy: pickNumber(s.default_strategy, localStorageStrategy),
    defaultAlbum: pickNumber(s.default_album, localStorageAlbum),
    defaultPermission: pickNumber(s.default_permission, localStoragePermission) ?? 1,
    allowUserImageProcessing: siteAllowUserImageProcessing,
    skipImageProcessing: siteSkipImageProcessing,
    imageProcessing: (s.image_processing as Record<string, unknown>) ?? null,
  }
}

function pickNumber(serverVal: unknown, localVal: string | null): number | null {
  if (serverVal != null) return Number(serverVal)
  if (localVal != null) return Number(localVal)
  return null
}

// ── LocalStorage → server migration (v1 one-way) ────────────────────
const MIGRATION_FLAG_KEY = 'pf-migrated-v1'

const LOCAL_PREF_KEYS = [
  'default_strategy_id',
  'default_album_id',
  'default_permission',
] as const

export interface MigrationResult {
  migrated: boolean
  reason: 'already_migrated' | 'no_local_prefs' | 'migrated' | 'failed'
}

export async function migrateLocalStorageToServer(
  updateProfile: (data: { settings?: Record<string, unknown> }) => Promise<void>,
  currentServerSettings: Record<string, unknown> | null | undefined,
): Promise<MigrationResult> {
  if (localStorage.getItem(MIGRATION_FLAG_KEY) === '1') {
    return { migrated: false, reason: 'already_migrated' }
  }

  const localPrefs: Record<string, unknown> = {}
  for (const key of LOCAL_PREF_KEYS) {
    const val = localStorage.getItem(key)
    if (val != null) localPrefs[key] = val
  }

  if (Object.keys(localPrefs).length === 0) {
    localStorage.setItem(MIGRATION_FLAG_KEY, '1')
    return { migrated: false, reason: 'no_local_prefs' }
  }

  const merged: Record<string, unknown> = { ...(currentServerSettings ?? {}) }
  if (localPrefs.default_strategy_id) {
    merged.default_strategy = Number(localPrefs.default_strategy_id)
  }
  if (localPrefs.default_album_id) {
    merged.default_album = Number(localPrefs.default_album_id)
  }
  if (localPrefs.default_permission) {
    merged.default_permission = Number(localPrefs.default_permission)
  }

  try {
    await updateProfile({ settings: merged })
    for (const key of LOCAL_PREF_KEYS) {
      localStorage.removeItem(key)
    }
    localStorage.setItem(MIGRATION_FLAG_KEY, '1')
    return { migrated: true, reason: 'migrated' }
  } catch {
    return { migrated: false, reason: 'failed' }
  }
}

export function buildThemeOverridePayload(
  override: ThemeOverride | null,
): Record<string, unknown> | undefined {
  if (!override) return undefined
  const payload: Record<string, unknown> = {}
  if (override.preset) payload.preset = override.preset
  if (override.mode) payload.mode = override.mode
  if (override.density) payload.density = override.density
  if (override.motion) payload.motion = override.motion
  return Object.keys(payload).length > 0 ? payload : undefined
}
