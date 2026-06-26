import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { useQuery, useQueryClient } from '@tanstack/react-query'

import { getAdminSettings, updateAdminSettings, type AdminSettings } from '@/lib/admin-api'
import { extractErrorMessage } from '@/lib/error-handler'
import type { ThemeConfig } from '@/lib/theme-config'

export interface SettingsForm {
  app_name: string
  app_url: string
  site_description: string
  favicon_url: string
  allow_guest_upload: boolean
  guest_capacity_mb: number
  allow_registration: boolean
  allow_oauth_registration: boolean
  allow_user_image_processing: boolean
  skip_image_processing: boolean
  require_email_verification: boolean
  user_initial_capacity_mb: number
  default_image_ttl: string
  guest_image_ttl: string
  moderation_mode: string
  footer_text_1: string
  footer_link_1: string
  footer_text_2: string
  footer_link_2: string
  analytics_provider: string
  analytics_domain: string
  analytics_script_url: string
  analytics_website_id: string
  analytics_measurement_id: string
  analytics_site_id: string
  analytics_custom_script: string
  analytics_connect_src: string
  analytics_script_src: string
  theme_custom_css: string
}

export const fieldInputCls = 'h-10 w-full rounded-lg border border-input bg-background px-3 text-sm outline-none transition-colors duration-150 focus:border-primary focus:ring-2 focus:ring-primary/20'
export const fieldTextareaCls = 'min-h-24 w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none transition-colors duration-150 focus:border-primary focus:ring-2 focus:ring-primary/20'

const defaultValues: SettingsForm = {
  app_name: '',
  app_url: '',
  site_description: '',
  favicon_url: '',
  allow_guest_upload: false,
  guest_capacity_mb: 10240,
  allow_registration: false,
  allow_oauth_registration: false,
  allow_user_image_processing: true,
  skip_image_processing: false,
  require_email_verification: false,
  user_initial_capacity_mb: 500,
  default_image_ttl: '0',
  guest_image_ttl: '0',
  moderation_mode: 'disabled',
  footer_text_1: '',
  footer_link_1: '',
  footer_text_2: '',
  footer_link_2: '',
  analytics_provider: '',
  analytics_domain: '',
  analytics_script_url: '',
  analytics_website_id: '',
  analytics_measurement_id: '',
  analytics_site_id: '',
  analytics_custom_script: '',
  analytics_connect_src: '',
  analytics_script_src: '',
  theme_custom_css: '',
}

const ttlOptions = new Set(['0', '24h', '168h', '720h', '2160h'])

function normalizeTTL(v?: unknown): string {
  const raw = typeof v === 'string' ? v.trim().toLowerCase() : ''
  if (!raw || raw === '0' || raw === '0s') return '0'
  if (ttlOptions.has(raw)) return raw

  // Go `time.Duration` strings are usually returned as "24h0m0s".
  const hourOnly = /^(\d+)h0m0s$/.exec(raw)
  if (hourOnly) {
    const compact = `${hourOnly[1]}h`
    if (ttlOptions.has(compact)) return compact
  }

  return '0'
}

function settingString(value: unknown): string {
  return typeof value === 'string' ? value : ''
}

function settingsToForm(data: AdminSettings): SettingsForm {
  const rawTheme = data.theme_config ?? null
  return {
    app_name: data.app_name,
    app_url: data.app_url || '',
    site_description: data.site_description || '',
    favicon_url: data.favicon_url || '',
    allow_guest_upload: data.allow_guest_upload,
    guest_capacity_mb: Math.round(data.guest_capacity_bytes / 1024 / 1024),
    allow_registration: data.allow_registration,
    allow_oauth_registration: data.allow_oauth_registration,
    allow_user_image_processing: data.allow_user_image_processing,
    skip_image_processing: data.skip_image_processing,
    require_email_verification: data.require_email_verification,
    user_initial_capacity_mb: Math.round(data.user_initial_capacity / 1024 / 1024),
    default_image_ttl: normalizeTTL(data.default_image_ttl),
    guest_image_ttl: normalizeTTL(data.guest_image_ttl),
    moderation_mode: data.moderation_mode,
    footer_text_1: data.footer_text_1 || '',
    footer_link_1: data.footer_link_1 || '',
    footer_text_2: data.footer_text_2 || '',
    footer_link_2: data.footer_link_2 || '',
    analytics_provider: data.analytics_provider || '',
    analytics_domain: settingString(data.analytics_config?.domain),
    analytics_script_url: settingString(data.analytics_config?.script_url),
    analytics_website_id: settingString(data.analytics_config?.website_id),
    analytics_measurement_id: settingString(data.analytics_config?.measurement_id),
    analytics_site_id: settingString(data.analytics_config?.site_id),
    analytics_custom_script: settingString(data.analytics_config?.script),
    analytics_connect_src: Array.isArray(data.analytics_config?.connect_src) ? (data.analytics_config.connect_src as string[]).join(', ') : '',
    analytics_script_src: Array.isArray(data.analytics_config?.script_src) ? (data.analytics_config.script_src as string[]).join(', ') : '',
    theme_custom_css: rawTheme?.custom_css || '',
  }
}

export function analyticsPayload(form: SettingsForm): Record<string, unknown> {
  const overrides: Record<string, unknown> = {}
  const connectSrc = parseCommaList(form.analytics_connect_src)
  if (connectSrc.length > 0) overrides.connect_src = connectSrc
  const scriptSrc = parseCommaList(form.analytics_script_src)
  if (scriptSrc.length > 0) overrides.script_src = scriptSrc

  switch (form.analytics_provider) {
    case 'plausible':
      return {
        domain: form.analytics_domain.trim(),
        script_url: form.analytics_script_url.trim() || 'https://plausible.io/js/script.js',
        ...overrides,
      }
    case 'umami':
      return {
        script_url: form.analytics_script_url.trim(),
        website_id: form.analytics_website_id.trim(),
        ...overrides,
      }
    case 'ga4':
      return { measurement_id: form.analytics_measurement_id.trim(), ...overrides }
    case 'baidu':
      return { site_id: form.analytics_site_id.trim(), ...overrides }
    case 'custom':
      return { script: form.analytics_custom_script.trim(), ...overrides }
    default:
      return {}
  }
}

function parseCommaList(s: string): string[] {
  return s.split(',').map((item) => item.trim()).filter((item) => item.length > 0)
}

export function themePayload(form: SettingsForm): ThemeConfig {
  return {
    custom_css: form.theme_custom_css,
  }
}

export function useAdminSettingsForm() {
  const { t } = useTranslation()
  const qc = useQueryClient()

  const { data, isLoading, error } = useQuery({
    queryKey: ['admin-settings'],
    queryFn: getAdminSettings,
  })

  const form = useForm<SettingsForm>({
    defaultValues,
    values: data ? settingsToForm(data) : undefined,
  })

  const saveSettings = async (payload: Partial<AdminSettings>) => {
    try {
      await updateAdminSettings(payload)
      toast.success(t('admin.saved'))
      await qc.invalidateQueries({ queryKey: ['admin-settings'] })
      await qc.invalidateQueries({ queryKey: ['site-config'] })
    } catch (err: unknown) {
      toast.error(extractErrorMessage(err, t('admin.saveFailed')))
    }
  }

  return { data, isLoading, error, form, saveSettings }
}
