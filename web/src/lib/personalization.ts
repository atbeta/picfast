import type { CopyPreferences } from './copy-template'
import { normalizeCopyFormat } from './copy-template'

export interface PersonalizationState {
  theme: { customCSS: string }
  mode: { colorMode: 'light' | 'dark' | 'system'; language: string }
  output: CopyPreferences
  workflow: PersonalizationWorkflow
  account: PersonalizationAccount
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

function pickNumber(serverVal: unknown, localVal: string | null): number | null {
  if (serverVal != null) return Number(serverVal)
  if (localVal != null) return Number(localVal)
  return null
}

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

// ── LocalStorage → server migration (upload defaults) ─────────────────
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
