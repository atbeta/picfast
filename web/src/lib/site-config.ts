import api from './api'

export interface SiteConfig {
  app_name: string
  allow_guest_upload: boolean
  allow_registration: boolean
  require_email_verification: boolean
  base_url: string
}

let cachedConfig: SiteConfig | null = null

export async function getSiteConfig(): Promise<SiteConfig> {
  if (cachedConfig) return cachedConfig
  const res = await api.get<{ status: boolean; data: SiteConfig }>('/config')
  cachedConfig = res.data.data
  return cachedConfig
}

export function clearSiteConfigCache() {
  cachedConfig = null
}
