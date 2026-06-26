export type ThemeMode = 'light' | 'dark' | 'system'

export interface ThemeTokenSet {
  background?: string
  foreground?: string
  card?: string
  cardForeground?: string
  primary?: string
  primaryForeground?: string
  secondary?: string
  secondaryForeground?: string
  muted?: string
  mutedForeground?: string
  accent?: string
  accentForeground?: string
  destructive?: string
  destructiveForeground?: string
  success?: string
  successForeground?: string
  warning?: string
  warningForeground?: string
  info?: string
  infoForeground?: string
  border?: string
  input?: string
  ring?: string
  popover?: string
  popoverForeground?: string
  sidebar?: string
  sidebarForeground?: string
  sidebarPrimary?: string
  sidebarPrimaryForeground?: string
  sidebarAccent?: string
  sidebarAccentForeground?: string
  sidebarBorder?: string
  sidebarRing?: string
  radius?: string
}

export interface ThemeConfig {
  preset?: string
  mode?: ThemeMode
  tokens?: {
    light?: ThemeTokenSet
    dark?: ThemeTokenSet
  }
  public?: {
    background_image?: string
    background_style?: 'soft' | 'clean' | 'image'
    logo_shape?: 'rounded' | 'circle' | 'square'
    density?: 'compact' | 'comfortable' | 'spacious'
    motion?: 'none' | 'subtle' | 'playful'
  }
  custom_css?: string
}

export interface ThemePreset {
  id: string
  nameKey: string
  descriptionKey: string
  config: ThemeConfig
}

export const defaultThemeConfig: ThemeConfig = {
  preset: 'default',
  mode: 'system',
  tokens: {},
  public: {
    background_style: 'soft',
    logo_shape: 'rounded',
  },
  custom_css: '',
}

export const themePresets: ThemePreset[] = [
  {
    id: 'default',
    nameKey: 'admin.themePresetDefault',
    descriptionKey: 'admin.themePresetDefaultDesc',
    config: defaultThemeConfig,
  },
  {
    id: 'moe',
    nameKey: 'admin.themePresetMoe',
    descriptionKey: 'admin.themePresetMoeDesc',
    config: {
      preset: 'moe',
      mode: 'system',
      tokens: {
        light: {
          background: 'oklch(0.985 0.018 340)',
          foreground: 'oklch(0.24 0.035 330)',
          card: 'oklch(1 0.012 340 / 86%)',
          primary: 'oklch(0.68 0.18 345)',
          primaryForeground: 'oklch(0.99 0.01 340)',
          secondary: 'oklch(0.94 0.04 330)',
          muted: 'oklch(0.95 0.026 330)',
          accent: 'oklch(0.9 0.07 20)',
          destructive: 'oklch(0.55 0.22 15)',
          destructiveForeground: 'oklch(0.99 0.01 340)',
          success: 'oklch(0.58 0.16 150)',
          successForeground: 'oklch(0.99 0.01 340)',
          warning: 'oklch(0.75 0.14 75)',
          warningForeground: 'oklch(0.2 0.02 80)',
          info: 'oklch(0.68 0.18 345)',
          infoForeground: 'oklch(0.99 0.01 340)',
          border: 'oklch(0.88 0.04 335)',
          ring: 'oklch(0.68 0.18 345)',
          popover: 'oklch(1 0.012 340 / 86%)',
          popoverForeground: 'oklch(0.24 0.035 330)',
          sidebarPrimary: 'oklch(0.68 0.18 345)',
          sidebarPrimaryForeground: 'oklch(0.99 0.01 340)',
          sidebarRing: 'oklch(0.68 0.18 345)',
          radius: '1rem',
        },
        dark: {
          background: 'oklch(0.16 0.035 320)',
          foreground: 'oklch(0.96 0.02 330)',
          card: 'oklch(0.2 0.04 320 / 88%)',
          primary: 'oklch(0.74 0.18 345)',
          secondary: 'oklch(0.27 0.05 315)',
          muted: 'oklch(0.26 0.04 315)',
          accent: 'oklch(0.36 0.09 20)',
          destructive: 'oklch(0.57 0.24 15)',
          destructiveForeground: 'oklch(0.99 0.01 340)',
          success: 'oklch(0.55 0.15 150)',
          successForeground: 'oklch(0.99 0.01 340)',
          warning: 'oklch(0.68 0.14 75)',
          warningForeground: 'oklch(0.15 0.02 80)',
          info: 'oklch(0.74 0.18 345)',
          infoForeground: 'oklch(0.99 0.01 340)',
          border: 'oklch(0.38 0.06 320)',
          ring: 'oklch(0.74 0.18 345)',
          popover: 'oklch(0.2 0.04 320 / 88%)',
          popoverForeground: 'oklch(0.96 0.02 330)',
          sidebarPrimary: 'oklch(0.74 0.18 345)',
          sidebarPrimaryForeground: 'oklch(0.99 0.01 340)',
          sidebarRing: 'oklch(0.74 0.18 345)',
          radius: '1rem',
        },
      },
      public: { background_style: 'soft', logo_shape: 'circle', density: 'comfortable', motion: 'subtle' },
    },
  },
  {
    id: 'cyber',
    nameKey: 'admin.themePresetCyber',
    descriptionKey: 'admin.themePresetCyberDesc',
    config: {
      preset: 'cyber',
      mode: 'dark',
      tokens: {
        light: {
          background: 'oklch(0.97 0.02 235)',
          foreground: 'oklch(0.18 0.04 255)',
          card: 'oklch(0.99 0.015 235 / 88%)',
          primary: 'oklch(0.62 0.23 210)',
          secondary: 'oklch(0.92 0.055 190)',
          muted: 'oklch(0.93 0.03 240)',
          accent: 'oklch(0.72 0.22 320)',
          destructive: 'oklch(0.6 0.26 20)',
          destructiveForeground: 'oklch(0.99 0 0)',
          success: 'oklch(0.6 0.18 155)',
          successForeground: 'oklch(0.99 0 0)',
          warning: 'oklch(0.72 0.16 80)',
          warningForeground: 'oklch(0.18 0.02 80)',
          info: 'oklch(0.62 0.23 210)',
          infoForeground: 'oklch(0.99 0 0)',
          border: 'oklch(0.82 0.055 230)',
          ring: 'oklch(0.72 0.22 320)',
          popover: 'oklch(0.99 0.015 235 / 88%)',
          popoverForeground: 'oklch(0.18 0.04 255)',
          sidebarPrimary: 'oklch(0.62 0.23 210)',
          sidebarPrimaryForeground: 'oklch(0.985 0 0)',
          sidebarRing: 'oklch(0.72 0.22 320)',
          radius: '0.5rem',
        },
        dark: {
          background: 'oklch(0.12 0.045 250)',
          foreground: 'oklch(0.94 0.04 205)',
          card: 'oklch(0.17 0.055 255 / 88%)',
          primary: 'oklch(0.73 0.22 205)',
          secondary: 'oklch(0.23 0.075 260)',
          muted: 'oklch(0.22 0.055 260)',
          accent: 'oklch(0.75 0.22 320)',
          destructive: 'oklch(0.62 0.27 22)',
          destructiveForeground: 'oklch(0.99 0 0)',
          success: 'oklch(0.62 0.2 155)',
          successForeground: 'oklch(0.99 0 0)',
          warning: 'oklch(0.7 0.17 80)',
          warningForeground: 'oklch(0.15 0.02 80)',
          info: 'oklch(0.73 0.22 205)',
          infoForeground: 'oklch(0.99 0 0)',
          border: 'oklch(0.34 0.09 250)',
          ring: 'oklch(0.75 0.22 320)',
          popover: 'oklch(0.17 0.055 255 / 88%)',
          popoverForeground: 'oklch(0.94 0.04 205)',
          sidebarPrimary: 'oklch(0.73 0.22 205)',
          sidebarPrimaryForeground: 'oklch(0.985 0 0)',
          sidebarRing: 'oklch(0.75 0.22 320)',
          radius: '0.5rem',
        },
      },
      public: { background_style: 'soft', logo_shape: 'square', density: 'compact', motion: 'subtle' },
    },
  },
  {
    id: 'pixel',
    nameKey: 'admin.themePresetPixel',
    descriptionKey: 'admin.themePresetPixelDesc',
    config: {
      preset: 'pixel',
      mode: 'system',
      tokens: {
        light: {
          background: 'oklch(0.97 0.025 85)',
          foreground: 'oklch(0.2 0.03 80)',
          card: 'oklch(1 0.018 85)',
          primary: 'oklch(0.58 0.18 145)',
          secondary: 'oklch(0.91 0.045 95)',
          muted: 'oklch(0.92 0.03 90)',
          accent: 'oklch(0.78 0.16 75)',
          destructive: 'oklch(0.5 0.22 15)',
          destructiveForeground: 'oklch(0.99 0 0)',
          success: 'oklch(0.6 0.2 140)',
          successForeground: 'oklch(0.99 0 0)',
          warning: 'oklch(0.72 0.14 75)',
          warningForeground: 'oklch(0.2 0.02 80)',
          info: 'oklch(0.58 0.18 145)',
          infoForeground: 'oklch(0.98 0 0)',
          border: 'oklch(0.62 0.04 85)',
          ring: 'oklch(0.58 0.18 145)',
          popover: 'oklch(1 0.018 85)',
          popoverForeground: 'oklch(0.2 0.03 80)',
          sidebarPrimary: 'oklch(0.58 0.18 145)',
          sidebarPrimaryForeground: 'oklch(0.98 0 0)',
          sidebarRing: 'oklch(0.58 0.18 145)',
          radius: '0.125rem',
        },
        dark: {
          background: 'oklch(0.16 0.025 90)',
          foreground: 'oklch(0.93 0.025 90)',
          card: 'oklch(0.2 0.03 90)',
          primary: 'oklch(0.7 0.17 145)',
          secondary: 'oklch(0.28 0.035 90)',
          muted: 'oklch(0.26 0.03 90)',
          accent: 'oklch(0.72 0.15 75)',
          destructive: 'oklch(0.52 0.24 18)',
          destructiveForeground: 'oklch(0.99 0 0)',
          success: 'oklch(0.62 0.19 140)',
          successForeground: 'oklch(0.99 0 0)',
          warning: 'oklch(0.68 0.13 75)',
          warningForeground: 'oklch(0.15 0.02 80)',
          info: 'oklch(0.7 0.17 145)',
          infoForeground: 'oklch(0.98 0 0)',
          border: 'oklch(0.42 0.035 90)',
          ring: 'oklch(0.7 0.17 145)',
          popover: 'oklch(0.2 0.03 90)',
          popoverForeground: 'oklch(0.93 0.025 90)',
          sidebarPrimary: 'oklch(0.7 0.17 145)',
          sidebarPrimaryForeground: 'oklch(0.98 0 0)',
          sidebarRing: 'oklch(0.7 0.17 145)',
          radius: '0.125rem',
        },
      },
      public: { background_style: 'clean', logo_shape: 'square', density: 'compact', motion: 'none' },
    },
  },
  {
    id: 'terminal',
    nameKey: 'admin.themePresetTerminal',
    descriptionKey: 'admin.themePresetTerminalDesc',
    config: {
      preset: 'terminal',
      mode: 'dark',
      tokens: {
        light: {
          background: 'oklch(0.965 0.01 145)',
          foreground: 'oklch(0.18 0.04 145)',
          card: 'oklch(0.99 0.01 145)',
          primary: 'oklch(0.52 0.18 145)',
          secondary: 'oklch(0.9 0.035 145)',
          muted: 'oklch(0.91 0.025 145)',
          accent: 'oklch(0.78 0.13 110)',
          destructive: 'oklch(0.5 0.2 10)',
          destructiveForeground: 'oklch(0.99 0 0)',
          success: 'oklch(0.55 0.2 140)',
          successForeground: 'oklch(0.99 0 0)',
          warning: 'oklch(0.7 0.14 80)',
          warningForeground: 'oklch(0.2 0.02 80)',
          info: 'oklch(0.52 0.18 145)',
          infoForeground: 'oklch(0.98 0 0)',
          border: 'oklch(0.75 0.04 145)',
          ring: 'oklch(0.52 0.18 145)',
          popover: 'oklch(0.99 0.01 145)',
          popoverForeground: 'oklch(0.18 0.04 145)',
          sidebarPrimary: 'oklch(0.52 0.18 145)',
          sidebarPrimaryForeground: 'oklch(0.98 0 0)',
          sidebarRing: 'oklch(0.52 0.18 145)',
          radius: '0.25rem',
        },
        dark: {
          background: 'oklch(0.11 0.025 145)',
          foreground: 'oklch(0.9 0.13 145)',
          card: 'oklch(0.15 0.035 145)',
          primary: 'oklch(0.75 0.19 145)',
          secondary: 'oklch(0.21 0.045 145)',
          muted: 'oklch(0.19 0.035 145)',
          accent: 'oklch(0.76 0.16 110)',
          destructive: 'oklch(0.52 0.22 12)',
          destructiveForeground: 'oklch(0.99 0 0)',
          success: 'oklch(0.65 0.2 140)',
          successForeground: 'oklch(0.99 0 0)',
          warning: 'oklch(0.65 0.13 75)',
          warningForeground: 'oklch(0.15 0.02 80)',
          info: 'oklch(0.75 0.19 145)',
          infoForeground: 'oklch(0.98 0 0)',
          border: 'oklch(0.35 0.065 145)',
          ring: 'oklch(0.75 0.19 145)',
          popover: 'oklch(0.15 0.035 145)',
          popoverForeground: 'oklch(0.9 0.13 145)',
          sidebarPrimary: 'oklch(0.75 0.19 145)',
          sidebarPrimaryForeground: 'oklch(0.98 0 0)',
          sidebarRing: 'oklch(0.75 0.19 145)',
          radius: '0.25rem',
        },
      },
      public: { background_style: 'clean', logo_shape: 'square', density: 'compact', motion: 'none' },
    },
  },
  {
    id: 'fresh',
    nameKey: 'admin.themePresetFresh',
    descriptionKey: 'admin.themePresetFreshDesc',
    config: {
      preset: 'fresh',
      mode: 'system',
      tokens: {
        light: {
          background: 'oklch(0.985 0.018 185)',
          foreground: 'oklch(0.2 0.035 205)',
          card: 'oklch(1 0.01 180 / 88%)',
          primary: 'oklch(0.58 0.16 190)',
          secondary: 'oklch(0.93 0.045 175)',
          muted: 'oklch(0.94 0.025 185)',
          accent: 'oklch(0.84 0.09 155)',
          destructive: 'oklch(0.58 0.22 15)',
          destructiveForeground: 'oklch(0.99 0 0)',
          success: 'oklch(0.6 0.16 155)',
          successForeground: 'oklch(0.99 0 0)',
          warning: 'oklch(0.74 0.14 80)',
          warningForeground: 'oklch(0.2 0.02 80)',
          info: 'oklch(0.58 0.16 190)',
          infoForeground: 'oklch(0.98 0 0)',
          border: 'oklch(0.84 0.04 185)',
          ring: 'oklch(0.58 0.16 190)',
          popover: 'oklch(1 0.01 180 / 88%)',
          popoverForeground: 'oklch(0.2 0.035 205)',
          sidebarPrimary: 'oklch(0.58 0.16 190)',
          sidebarPrimaryForeground: 'oklch(0.98 0 0)',
          sidebarRing: 'oklch(0.58 0.16 190)',
          radius: '0.875rem',
        },
        dark: {
          background: 'oklch(0.13 0.035 205)',
          foreground: 'oklch(0.95 0.025 180)',
          card: 'oklch(0.18 0.04 205 / 88%)',
          primary: 'oklch(0.7 0.15 190)',
          secondary: 'oklch(0.25 0.045 200)',
          muted: 'oklch(0.23 0.035 200)',
          accent: 'oklch(0.67 0.12 155)',
          destructive: 'oklch(0.6 0.24 17)',
          destructiveForeground: 'oklch(0.99 0 0)',
          success: 'oklch(0.55 0.15 155)',
          successForeground: 'oklch(0.99 0 0)',
          warning: 'oklch(0.68 0.13 75)',
          warningForeground: 'oklch(0.15 0.02 80)',
          info: 'oklch(0.7 0.15 190)',
          infoForeground: 'oklch(0.98 0 0)',
          border: 'oklch(0.35 0.045 200)',
          ring: 'oklch(0.7 0.15 190)',
          popover: 'oklch(0.18 0.04 205 / 88%)',
          popoverForeground: 'oklch(0.95 0.025 180)',
          sidebarPrimary: 'oklch(0.7 0.15 190)',
          sidebarPrimaryForeground: 'oklch(0.98 0 0)',
          sidebarRing: 'oklch(0.7 0.15 190)',
          radius: '0.875rem',
        },
      },
      public: { background_style: 'soft', logo_shape: 'rounded', density: 'comfortable', motion: 'subtle' },
    },
  },
]

export function presetById(id?: string) {
  return themePresets.find((preset) => preset.id === id) ?? themePresets[0]
}

export function mergeThemeConfig(config?: ThemeConfig | null): ThemeConfig {
  const preset = presetById(config?.preset)
  return {
    ...defaultThemeConfig,
    ...preset.config,
    ...config,
    tokens: {
      light: {
        ...(preset.config.tokens?.light ?? {}),
        ...(config?.tokens?.light ?? {}),
      },
      dark: {
        ...(preset.config.tokens?.dark ?? {}),
        ...(config?.tokens?.dark ?? {}),
      },
    },
    public: {
      ...(defaultThemeConfig.public ?? {}),
      ...(preset.config.public ?? {}),
      ...(config?.public ?? {}),
    },
    custom_css: config?.custom_css ?? '',
  }
}

const tokenToCssVar: Record<keyof ThemeTokenSet, string> = {
  background: '--background',
  foreground: '--foreground',
  card: '--card',
  cardForeground: '--card-foreground',
  primary: '--primary',
  primaryForeground: '--primary-foreground',
  secondary: '--secondary',
  secondaryForeground: '--secondary-foreground',
  muted: '--muted',
  mutedForeground: '--muted-foreground',
  accent: '--accent',
  accentForeground: '--accent-foreground',
  destructive: '--destructive',
  destructiveForeground: '--destructive-foreground',
  success: '--success',
  successForeground: '--success-foreground',
  warning: '--warning',
  warningForeground: '--warning-foreground',
  info: '--info',
  infoForeground: '--info-foreground',
  border: '--border',
  input: '--input',
  ring: '--ring',
  popover: '--popover',
  popoverForeground: '--popover-foreground',
  sidebar: '--sidebar',
  sidebarForeground: '--sidebar-foreground',
  sidebarPrimary: '--sidebar-primary',
  sidebarPrimaryForeground: '--sidebar-primary-foreground',
  sidebarAccent: '--sidebar-accent',
  sidebarAccentForeground: '--sidebar-accent-foreground',
  sidebarBorder: '--sidebar-border',
  sidebarRing: '--sidebar-ring',
  radius: '--radius',
}

function cssDeclarations(tokens?: ThemeTokenSet): string {
  if (!tokens) return ''
  return Object.entries(tokens)
    .map(([key, value]) => {
      const cssVar = tokenToCssVar[key as keyof ThemeTokenSet]
      if (!cssVar || typeof value !== 'string' || !value.trim()) return ''
      if (/[;{}<>]/.test(value)) return ''
      return `${cssVar}: ${value.trim()};`
    })
    .filter(Boolean)
    .join('\n')
}

function cssURL(value?: string): string {
  const raw = value?.trim()
  if (!raw) return 'none'
  return `url("${raw.replace(/["\\]/g, '\\$&')}")`
}

export function themeConfigToCSS(config?: ThemeConfig | null): string {
  const merged = mergeThemeConfig(config)
  const light = cssDeclarations(merged.tokens?.light)
  const dark = cssDeclarations(merged.tokens?.dark)
  const publicVars = `
--pf-public-bg-image: ${cssURL(merged.public?.background_image)};
--pf-public-bg-opacity: ${merged.public?.background_style === 'image' ? '0.18' : '0'};
`

  return `
:root {
${light}
${publicVars}
}
.dark {
${dark}
}
${merged.custom_css ?? ''}
`
}
