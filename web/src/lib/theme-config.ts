export type ThemeMode = 'light' | 'dark' | 'system'

export interface ThemeConfig {
  mode?: ThemeMode
  custom_css?: string
}
