import api from './api'

export interface SiteConfig {
  app_name: string
  site_description: string
  favicon_url: string
  allow_guest_upload: boolean
  guest_capacity_bytes: number
  allow_registration: boolean
  allow_user_image_processing: boolean
  require_email_verification: boolean
  base_url: string
  icp_number: string
  icp_link: string
  psb_number: string
  psb_link: string
  analytics_provider: string
  analytics_config: Record<string, unknown>
  github_url: string
}

export async function getSiteConfig(): Promise<SiteConfig> {
  const res = await api.get<{ status: boolean; data: SiteConfig }>('/config')
  return res.data.data
}

export interface VersionInfo {
  version: string
  commit: string
  build_time: string
  github_url: string
}

export async function getVersionInfo(): Promise<VersionInfo> {
  const res = await api.get<{ status: boolean; data: VersionInfo }>('/version')
  return res.data.data
}
