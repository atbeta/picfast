import {
  mergeThemeConfig,
  presetById,
  type ThemeConfig,
  type ThemeMode,
} from '@/lib/theme-config'

export type ThemePackageScope = 'site' | 'user'

const SITE_PACKAGE_KEYS = ['preset', 'mode', 'tokens', 'public', 'custom_css'] as const
const USER_PACKAGE_KEYS = ['preset', 'mode', 'density', 'motion'] as const

const THEME_MODES = new Set<ThemeMode>(['light', 'dark', 'system'])
const BACKGROUND_STYLES = new Set(['soft', 'clean', 'image'])
const LOGO_SHAPES = new Set(['rounded', 'circle', 'square'])
const DENSITIES = new Set(['compact', 'comfortable', 'spacious'])
const MOTIONS = new Set(['none', 'subtle', 'playful'])

export class ThemePackageError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'ThemePackageError'
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function stripEmptyTokens(tokens?: ThemeConfig['tokens']): ThemeConfig['tokens'] | undefined {
  if (!tokens) return undefined
  const light = Object.fromEntries(
    Object.entries(tokens.light ?? {}).filter(([, value]) => typeof value === 'string' && value.trim()),
  )
  const dark = Object.fromEntries(
    Object.entries(tokens.dark ?? {}).filter(([, value]) => typeof value === 'string' && value.trim()),
  )
  if (Object.keys(light).length === 0 && Object.keys(dark).length === 0) return undefined
  return {
    ...(Object.keys(light).length > 0 ? { light } : {}),
    ...(Object.keys(dark).length > 0 ? { dark } : {}),
  }
}

function stripPublic(publicConfig?: ThemeConfig['public']): ThemeConfig['public'] | undefined {
  if (!publicConfig) return undefined
  const next = Object.fromEntries(
    Object.entries(publicConfig).filter(([, value]) => typeof value === 'string' && value.trim()),
  )
  return Object.keys(next).length > 0 ? next as ThemeConfig['public'] : undefined
}

// ── Export ──────────────────────────────────────────────────────────

/** Export only safe, shareable theme fields for the given scope. */
export function exportThemePackage(config: ThemeConfig, scope: ThemePackageScope): ThemeConfig {
  const merged = mergeThemeConfig(config)
  const customCSS = merged.custom_css?.trim() ?? ''

  if (scope === 'user') {
    const pkg: ThemeConfig = {
      preset: merged.preset || 'default',
      mode: merged.mode || 'system',
      public: {},
    }
    const pub = merged.public
    if (pub?.density && pub.density !== 'comfortable') {
      pkg.public = { ...pkg.public, density: pub.density }
    }
    if (pub?.motion && pub.motion !== 'subtle') {
      pkg.public = { ...pkg.public, motion: pub.motion }
    }
    if (!pkg.public?.density && !pkg.public?.motion) {
      delete pkg.public
    }
    return pkg
  }

  const tokens = stripEmptyTokens(merged.tokens)
  const publicConfig = stripPublic(merged.public)

  const pkg: ThemeConfig = {
    preset: merged.preset || 'default',
    mode: merged.mode || 'system',
  }
  if (tokens) pkg.tokens = tokens
  if (publicConfig) pkg.public = publicConfig
  if (customCSS) pkg.custom_css = customCSS
  return pkg
}

export function serializeThemePackage(config: ThemeConfig, scope: ThemePackageScope): string {
  return JSON.stringify(exportThemePackage(config, scope), null, 2)
}

// ── Parsing helpers ─────────────────────────────────────────────────

function assertSafeString(value: unknown, field: string, maxLen: number): string | undefined {
  if (value === undefined || value === null || value === '') return undefined
  if (typeof value !== 'string') {
    throw new ThemePackageError(`${field} must be a string`)
  }
  const trimmed = value.trim()
  if (trimmed.length > maxLen) {
    throw new ThemePackageError(`${field} is too long`)
  }
  if (/[;{}<>]/.test(trimmed) && field.includes('color')) {
    throw new ThemePackageError(`${field} contains invalid characters`)
  }
  return trimmed
}

function parseTokens(raw: unknown): ThemeConfig['tokens'] | undefined {
  if (raw === undefined || raw === null) return undefined
  if (!isRecord(raw)) throw new ThemePackageError('tokens must be an object')

  const modes = ['light', 'dark'] as const
  const next: NonNullable<ThemeConfig['tokens']> = {}
  for (const mode of modes) {
    const values = raw[mode]
    if (values === undefined || values === null) continue
    if (!isRecord(values)) throw new ThemePackageError(`tokens.${mode} must be an object`)
    const tokenSet: Record<string, string> = {}
    for (const [key, value] of Object.entries(values)) {
      const maxLen = key === 'radius' ? 48 : 160
      const parsed = assertSafeString(value, `tokens.${mode}.${key}`, maxLen)
      if (!parsed) continue
      if (key === 'radius') {
        if (!CSS.supports('border-radius', parsed)) {
          throw new ThemePackageError(`tokens.${mode}.radius is invalid`)
        }
      } else if (!CSS.supports('color', parsed) && !/^#([0-9a-fA-F]{3,8})$/.test(parsed)) {
        throw new ThemePackageError(`tokens.${mode}.${key} is invalid`)
      }
      tokenSet[key] = parsed
    }
    if (Object.keys(tokenSet).length > 0) {
      next[mode] = tokenSet
    }
  }
  return Object.keys(next).length > 0 ? next : undefined
}

function parsePublic(raw: unknown): ThemeConfig['public'] | undefined {
  if (raw === undefined || raw === null) return undefined
  if (!isRecord(raw)) throw new ThemePackageError('public must be an object')

  const next: ThemeConfig['public'] = {}
  const backgroundImage = assertSafeString(raw.background_image, 'public.background_image', 2048)
  if (backgroundImage) next.background_image = backgroundImage

  const backgroundStyle = assertSafeString(raw.background_style, 'public.background_style', 16)
  if (backgroundStyle) {
    if (!BACKGROUND_STYLES.has(backgroundStyle)) {
      throw new ThemePackageError('public.background_style is invalid')
    }
    next.background_style = backgroundStyle as NonNullable<ThemeConfig['public']>['background_style']
  }

  const logoShape = assertSafeString(raw.logo_shape, 'public.logo_shape', 16)
  if (logoShape) {
    if (!LOGO_SHAPES.has(logoShape)) {
      throw new ThemePackageError('public.logo_shape is invalid')
    }
    next.logo_shape = logoShape as NonNullable<ThemeConfig['public']>['logo_shape']
  }

  const semanticFields: Array<{ key: keyof NonNullable<ThemeConfig['public']>; allowed: Set<string> }> = [
    { key: 'density', allowed: DENSITIES },
    { key: 'motion', allowed: MOTIONS },
  ]
  for (const field of semanticFields) {
    const value = assertSafeString(raw[field.key as string], `public.${String(field.key)}`, 16)
    if (!value) continue
    if (!field.allowed.has(value)) {
      throw new ThemePackageError(`public.${String(field.key)} is invalid`)
    }
    next[field.key] = value as never
  }

  return Object.keys(next).length > 0 ? next : undefined
}

// ── Parse (scope-aware) ─────────────────────────────────────────────

export interface ParsedThemePackage {
  config: ThemeConfig
  scope: ThemePackageScope
}

/** Parse imported JSON into a sanitized ThemeConfig, auto-detecting scope. */
export function parseThemePackage(raw: string): ParsedThemePackage {
  let parsed: unknown
  try {
    parsed = JSON.parse(raw)
  } catch {
    throw new ThemePackageError('invalid JSON')
  }
  if (!isRecord(parsed)) {
    throw new ThemePackageError('theme package must be a JSON object')
  }

  const scope = (parsed.scope === 'user') ? 'user' as ThemePackageScope : 'site' as ThemePackageScope
  const keySet: Set<string> = scope === 'user'
    ? new Set(USER_PACKAGE_KEYS)
    : new Set(SITE_PACKAGE_KEYS)

  for (const key of Object.keys(parsed)) {
    if (key === 'scope') continue
    if (!keySet.has(key)) {
      if (scope === 'user') {
        throw new ThemePackageError(`field "${key}" is not allowed in user-scope theme package`)
      }
      throw new ThemePackageError(`unknown field: ${key}`)
    }
  }

  const preset = assertSafeString(parsed.preset, 'preset', 80) || 'default'
  presetById(preset)

  if (scope === 'user') {
    const modeRaw = assertSafeString(parsed.mode, 'mode', 16) || 'system'
    if (!THEME_MODES.has(modeRaw as ThemeMode)) {
      throw new ThemePackageError('mode is invalid')
    }

    const densityRaw = assertSafeString(parsed.density, 'density', 16)
    if (densityRaw && !DENSITIES.has(densityRaw)) {
      throw new ThemePackageError('density is invalid')
    }

    const motionRaw = assertSafeString(parsed.motion, 'motion', 16)
    if (motionRaw && !MOTIONS.has(motionRaw)) {
      throw new ThemePackageError('motion is invalid')
    }

    const config: ThemeConfig = {
      preset,
      mode: modeRaw as ThemeMode,
      public: {},
    }
    if (densityRaw) config.public = { ...config.public, density: densityRaw as NonNullable<ThemeConfig['public']>['density'] }
    if (motionRaw) config.public = { ...config.public, motion: motionRaw as NonNullable<ThemeConfig['public']>['motion'] }
    return { config, scope }
  }

  const modeRaw = assertSafeString(parsed.mode, 'mode', 16) || 'system'
  if (!THEME_MODES.has(modeRaw as ThemeMode)) {
    throw new ThemePackageError('mode is invalid')
  }

  const customCSS = assertSafeString(parsed.custom_css, 'custom_css', 20000)
  const tokens = parseTokens(parsed.tokens)
  const publicConfig = parsePublic(parsed.public)

  const config: ThemeConfig = { preset, mode: modeRaw as ThemeMode }
  if (tokens) config.tokens = tokens
  if (publicConfig) config.public = publicConfig
  if (customCSS) config.custom_css = customCSS
  return { config, scope }
}

// ── Admin form helpers (unchanged) ──────────────────────────────────

export interface ThemeFormImportFields {
  theme_preset: string
  theme_mode: string
  theme_primary: string
  theme_accent: string
  theme_radius: string
  theme_background_image: string
  theme_background_style: string
  theme_logo_shape: string
  theme_custom_css: string
}

/** Map imported theme config onto admin settings form fields. */
export function themeConfigToFormFields(config: ThemeConfig): ThemeFormImportFields {
  const merged = mergeThemeConfig(config)
  const light = config.tokens?.light ?? {}
  const preset = presetById(merged.preset)
  return {
    theme_preset: merged.preset || 'default',
    theme_mode: merged.mode || 'system',
    theme_primary: light.primary?.trim() || '',
    theme_accent: light.accent?.trim() || '',
    theme_radius: light.radius?.trim() || '',
    theme_background_image: merged.public?.background_image?.trim() || '',
    theme_background_style: merged.public?.background_style || preset.config.public?.background_style || 'soft',
    theme_logo_shape: merged.public?.logo_shape || preset.config.public?.logo_shape || 'rounded',
    theme_custom_css: config.custom_css?.trim() || '',
  }
}

/** Reset form theme fields to a built-in preset without token overrides. */
export function presetDefaultFormFields(presetID: string): ThemeFormImportFields {
  return themeConfigToFormFields({ preset: presetID })
}
