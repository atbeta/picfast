import api from './api'

export interface SiteConfig {
  app_name: string
  allow_guest_upload: boolean
  allow_registration: boolean
  require_email_verification: boolean
  base_url: string
}

export async function getSiteConfig(): Promise<SiteConfig> {
  const res = await api.get<{ status: boolean; data: SiteConfig }>('/config')
  return res.data.data
}
