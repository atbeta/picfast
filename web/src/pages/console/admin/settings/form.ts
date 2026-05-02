import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'

import { getAdminSettings, updateAdminSettings, type AdminSettings } from '@/lib/admin-api'
import { extractErrorMessage } from '@/lib/error-handler'

export interface SettingsForm {
  app_name: string
  app_url: string
  site_description: string
  favicon_url: string
  allow_guest_upload: boolean
  guest_capacity_mb: number
  allow_registration: boolean
  require_email_verification: boolean
  user_initial_capacity_mb: number
  default_image_ttl: string
  moderation_mode: string
  icp_number: string
  icp_link: string
  psb_number: string
  psb_link: string
  analytics_provider: string
  analytics_domain: string
  analytics_script_url: string
  analytics_website_id: string
  analytics_measurement_id: string
  analytics_site_id: string
  analytics_custom_script: string
}

export const fieldInputCls = 'h-11 w-full rounded-lg border border-input bg-background px-4 text-sm outline-none transition-colors duration-150 focus:border-primary focus:ring-2 focus:ring-primary/20'
export const fieldTextareaCls = 'min-h-28 w-full rounded-lg border border-input bg-background px-4 py-3 text-sm outline-none transition-colors duration-150 focus:border-primary focus:ring-2 focus:ring-primary/20'

const defaultValues: SettingsForm = {
  app_name: '',
  app_url: '',
  site_description: '',
  favicon_url: '',
  allow_guest_upload: false,
  guest_capacity_mb: 10240,
  allow_registration: false,
  require_email_verification: false,
  user_initial_capacity_mb: 500,
  default_image_ttl: '0',
  moderation_mode: 'disabled',
  icp_number: '',
  icp_link: '',
  psb_number: '',
  psb_link: '',
  analytics_provider: '',
  analytics_domain: '',
  analytics_script_url: '',
  analytics_website_id: '',
  analytics_measurement_id: '',
  analytics_site_id: '',
  analytics_custom_script: '',
}

function normalizeTTL(v?: string): string {
  const val = (v || '').trim().toLowerCase()
  if (!val || val === '0' || val === '0s') return '0'
  return v || '0'
}

function settingString(value: unknown): string {
  return typeof value === 'string' ? value : ''
}

function settingsToForm(data: AdminSettings): SettingsForm {
  return {
    app_name: data.app_name,
    app_url: data.app_url || '',
    site_description: data.site_description || '',
    favicon_url: data.favicon_url || '',
    allow_guest_upload: data.allow_guest_upload,
    guest_capacity_mb: Math.round(data.guest_capacity_bytes / 1024 / 1024),
    allow_registration: data.allow_registration,
    require_email_verification: data.require_email_verification,
    user_initial_capacity_mb: Math.round(data.user_initial_capacity / 1024 / 1024),
    default_image_ttl: normalizeTTL(data.default_image_ttl),
    moderation_mode: data.moderation_mode,
    icp_number: data.icp_number || '',
    icp_link: data.icp_link || '',
    psb_number: data.psb_number || '',
    psb_link: data.psb_link || '',
    analytics_provider: data.analytics_provider || '',
    analytics_domain: settingString(data.analytics_config?.domain),
    analytics_script_url: settingString(data.analytics_config?.script_url),
    analytics_website_id: settingString(data.analytics_config?.website_id),
    analytics_measurement_id: settingString(data.analytics_config?.measurement_id),
    analytics_site_id: settingString(data.analytics_config?.site_id),
    analytics_custom_script: settingString(data.analytics_config?.script),
  }
}

export function analyticsPayload(form: SettingsForm): Record<string, unknown> {
  switch (form.analytics_provider) {
    case 'plausible':
      return {
        domain: form.analytics_domain.trim(),
        script_url: form.analytics_script_url.trim() || 'https://plausible.io/js/script.js',
      }
    case 'umami':
      return {
        script_url: form.analytics_script_url.trim(),
        website_id: form.analytics_website_id.trim(),
      }
    case 'ga4':
      return { measurement_id: form.analytics_measurement_id.trim() }
    case 'baidu':
      return { site_id: form.analytics_site_id.trim() }
    case 'custom':
      return { script: form.analytics_custom_script.trim() }
    default:
      return {}
  }
}

export function useAdminSettingsForm() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [success, setSuccess] = useState(false)
  const [errorMsg, setErrorMsg] = useState('')

  const { data, isLoading, error } = useQuery({
    queryKey: ['admin-settings'],
    queryFn: getAdminSettings,
  })

  const form = useForm<SettingsForm>({
    defaultValues,
    values: data ? settingsToForm(data) : undefined,
  })

  const saveSettings = async (payload: Partial<AdminSettings>) => {
    setSuccess(false)
    setErrorMsg('')
    try {
      await updateAdminSettings(payload)
      setSuccess(true)
      await qc.invalidateQueries({ queryKey: ['admin-settings'] })
      await qc.invalidateQueries({ queryKey: ['site-config'] })
    } catch (err: unknown) {
      setErrorMsg(extractErrorMessage(err, t('admin.saveFailed')))
    }
  }

  return { data, isLoading, error, form, success, errorMsg, saveSettings }
}
