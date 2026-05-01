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

function setFavicon(url: string) {
  const href = url.trim()
  if (!href) return
  const selectors = ['link[rel="icon"]', 'link[rel="shortcut icon"]', 'link[rel="apple-touch-icon"]']
  selectors.forEach((selector) => {
    const existing = document.head.querySelector<HTMLLinkElement>(selector)
    const link = existing ?? document.createElement('link')
    if (selector.includes('shortcut')) {
      link.setAttribute('rel', 'shortcut icon')
    } else if (selector.includes('apple-touch-icon')) {
      link.setAttribute('rel', 'apple-touch-icon')
    } else {
      link.setAttribute('rel', 'icon')
    }
    link.setAttribute('href', href)
    link.dataset.picfastFaviconFallbackBound = ''
    link.onerror = null
    link.addEventListener('error', () => {
      if (link.dataset.picfastFaviconFallbackBound === '1') return
      link.dataset.picfastFaviconFallbackBound = '1'
      link.setAttribute('href', '/favicon-default.svg')
    }, { once: true })
    if (!existing) document.head.appendChild(link)
  })
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
    const faviconURL = config.favicon_url?.trim() || ''

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
    if (faviconURL) setFavicon(faviconURL)
    injectAnalytics(config)

    return removeAnalyticsScripts
  }, [config])

  return null
}
