import { useEffect } from 'react'

import type { SiteConfig } from '@/lib/site-config'

function ensureMeta(selector: string, create: () => HTMLMetaElement): HTMLMetaElement {
  const existing = document.head.querySelector<HTMLMetaElement>(selector)
  if (existing) return existing
  const meta = create()
  document.head.appendChild(meta)
  return meta
}

function setNamedMeta(name: string, content: string) {
  const meta = ensureMeta(`meta[name="${name}"]`, () => {
    const el = document.createElement('meta')
    el.setAttribute('name', name)
    return el
  })
  meta.setAttribute('content', content)
}

function setPropertyMeta(property: string, content: string) {
  const meta = ensureMeta(`meta[property="${property}"]`, () => {
    const el = document.createElement('meta')
    el.setAttribute('property', property)
    return el
  })
  meta.setAttribute('content', content)
}

function setCanonical(url: string) {
  const existing = document.head.querySelector<HTMLLinkElement>('link[rel="canonical"]')
  const link = existing ?? document.createElement('link')
  link.setAttribute('rel', 'canonical')
  link.setAttribute('href', url)
  if (!existing) document.head.appendChild(link)
}

function removeAnalyticsScripts() {
  document.querySelectorAll('[data-picfast-analytics="true"]').forEach((node) => node.remove())
}

function appendScript(script: HTMLScriptElement) {
  script.dataset.picfastAnalytics = 'true'
  document.body.appendChild(script)
}

function injectAnalytics(config: SiteConfig) {
  removeAnalyticsScripts()
  const provider = config.analytics_provider
  const analytics = config.analytics_config || {}
  if (!provider) return

  if (provider === 'plausible') {
    const domain = String(analytics.domain || '').trim()
    if (!domain) return
    const script = document.createElement('script')
    script.defer = true
    script.src = String(analytics.script_url || 'https://plausible.io/js/script.js')
    script.setAttribute('data-domain', domain)
    appendScript(script)
    return
  }

  if (provider === 'umami') {
    const scriptURL = String(analytics.script_url || '').trim()
    const websiteID = String(analytics.website_id || '').trim()
    if (!scriptURL || !websiteID) return
    const script = document.createElement('script')
    script.defer = true
    script.src = scriptURL
    script.setAttribute('data-website-id', websiteID)
    appendScript(script)
    return
  }

  if (provider === 'ga4') {
    const measurementID = String(analytics.measurement_id || '').trim()
    if (!measurementID) return
    const loader = document.createElement('script')
    loader.async = true
    loader.src = `https://www.googletagmanager.com/gtag/js?id=${encodeURIComponent(measurementID)}`
    appendScript(loader)

    const inline = document.createElement('script')
    inline.text = `window.dataLayer=window.dataLayer||[];function gtag(){dataLayer.push(arguments);}gtag('js',new Date());gtag('config','${measurementID}');`
    appendScript(inline)
    return
  }

  if (provider === 'baidu') {
    const siteID = String(analytics.site_id || '').trim()
    if (!siteID) return
    const script = document.createElement('script')
    script.src = `https://hm.baidu.com/hm.js?${encodeURIComponent(siteID)}`
    appendScript(script)
    return
  }

  if (provider === 'custom') {
    const raw = String(analytics.script || '').trim()
    if (!raw) return
    const container = document.createElement('div')
    container.dataset.picfastAnalytics = 'true'
    container.innerHTML = raw
    container.querySelectorAll('script').forEach((oldScript) => {
      const script = document.createElement('script')
      Array.from(oldScript.attributes).forEach((attr) => script.setAttribute(attr.name, attr.value))
      script.text = oldScript.text
      oldScript.replaceWith(script)
    })
    document.body.appendChild(container)
  }
}

export function SiteMetadata({ config }: { config: SiteConfig }) {
  useEffect(() => {
    const title = config.app_name?.trim() || 'PicFast'
    const description = config.site_description?.trim() || 'PicFast is a modern self-hosted image hosting service.'
    const canonical = config.base_url?.trim() || window.location.origin

    document.title = title
    setNamedMeta('description', description)
    setPropertyMeta('og:title', title)
    setPropertyMeta('og:description', description)
    setPropertyMeta('og:type', 'website')
    setPropertyMeta('og:url', canonical)
    setNamedMeta('twitter:card', 'summary')
    setNamedMeta('twitter:title', title)
    setNamedMeta('twitter:description', description)
    setCanonical(canonical)
    injectAnalytics(config)

    return removeAnalyticsScripts
  }, [config])

  return null
}
