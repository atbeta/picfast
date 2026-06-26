import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useSearchParams } from 'react-router-dom'
import { toast } from 'sonner'

import { useAuth } from '@/lib/auth-context'
import { getOAuthIdentities, unlinkOAuth, type OAuthIdentity } from '@/lib/auth'
import { getSiteConfig } from '@/lib/site-config'
import { extractErrorMessage } from '@/lib/error-handler'
import { Button } from '@/components/ui/button'

import { AccountPanel } from './settings/account-panel'
import { WorkflowPanel } from './settings/workflow-panel'

type SettingsTab = 'account' | 'upload' | 'accounts'

const TABS: { id: SettingsTab; labelKey: string }[] = [
  { id: 'account', labelKey: 'settings.tabAccount' },
  { id: 'upload', labelKey: 'settings.tabUpload' },
  { id: 'accounts', labelKey: 'settings.tabConnectedAccounts' },
]

export function SettingsPage() {
  const { t } = useTranslation()
  const { user } = useAuth()
  const [searchParams] = useSearchParams()
  const oauthError = searchParams.get('oauth_error')

  const [activeTab, setActiveTab] = useState<SettingsTab>('account')
  const [oauthIdentities, setOauthIdentities] = useState<OAuthIdentity[]>([])
  const [unlinkingProvider, setUnlinkingProvider] = useState<string | null>(null)
  const [siteConfig, setSiteConfig] = useState<Awaited<ReturnType<typeof getSiteConfig>> | null>(null)

  useEffect(() => {
    if (oauthError) {
      const url = new URL(window.location.href)
      url.searchParams.delete('oauth_error')
      window.history.replaceState({}, '', url.toString())
    }
  }, [oauthError])

  useEffect(() => {
    getOAuthIdentities().then(setOauthIdentities).catch(() => {})
    getSiteConfig().then(setSiteConfig).catch(() => {})
  }, [])

  const handleUnlink = async (provider: string) => {
    if (!window.confirm(t('settings.confirmUnlinkProvider', { provider }))) return
    setUnlinkingProvider(provider)
    try {
      await unlinkOAuth(provider)
      setOauthIdentities((prev) => prev.filter((i) => i.provider !== provider))
      toast.success(t('settings.unlinkSuccess'))
    } catch (err: unknown) {
      const msg = extractErrorMessage(err, t('settings.unlinkFailed'))
      if (msg.includes('lock out') || msg.includes('set a password')) {
        toast.error(t('settings.oauthError.lockoutPrevented', { defaultValue: t('settings.unlinkFailed') }))
      } else {
        toast.error(msg)
      }
    } finally {
      setUnlinkingProvider(null)
    }
  }

  if (!user) return null

  const linkedProviders = new Set(oauthIdentities.map((i) => i.provider))
  const oauthProviders = siteConfig?.oauth_providers ?? []
  const unlinkedProviders = oauthProviders.filter((p) => !linkedProviders.has(p.id))

  const visibleTabs = TABS.filter(tab => {
    if (tab.id === 'accounts' && oauthProviders.length === 0 && oauthIdentities.length === 0) return false
    return true
  })

  return (
    <section className="space-y-6">
      <h1 className="text-2xl font-bold tracking-tight">{t('page.settings.title')}</h1>

      {oauthError && (
        <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {t(`settings.oauthError.${oauthError}`, { defaultValue: t('settings.oauthError.generic') })}
        </p>
      )}

      <div className="flex flex-wrap gap-1 border-b border-border/40 pb-0">
        {visibleTabs.map((tab) => (
          <button
            key={tab.id}
            type="button"
            onClick={() => setActiveTab(tab.id)}
            className={[
              'px-4 py-2.5 text-sm font-medium transition-colors border-b-2 -mb-px',
              activeTab === tab.id
                ? 'border-primary text-foreground'
                : 'border-transparent text-muted-foreground hover:text-foreground',
            ].join(' ')}
          >
            {t(tab.labelKey, { defaultValue: tab.id })}
          </button>
        ))}
      </div>

      <div className="pb-8">
        {activeTab === 'account' && <AccountPanel />}
        {activeTab === 'upload' && <WorkflowPanel />}
        {activeTab === 'accounts' && (
          <div className="space-y-6">
            <div>
              <h2 className="text-base font-semibold tracking-tight text-foreground">
                {t('settings.connectedAccounts')}
              </h2>
              <p className="text-sm text-muted-foreground">
                {t('settings.connectedAccountsDesc')}
              </p>
            </div>
            <div className="rounded-xl border border-border bg-card p-6 shadow-sm">
              <div className="space-y-4">
                {oauthIdentities.map((identity) => (
                  <div key={identity.provider} className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-foreground">{identity.provider}</p>
                      <p className="text-xs text-muted-foreground">{identity.email}</p>
                    </div>
                    <Button
                      size="sm"
                      variant="outline"
                      disabled={unlinkingProvider === identity.provider}
                      onClick={() => handleUnlink(identity.provider)}
                    >
                      {unlinkingProvider === identity.provider ? t('settings.saving') : t('settings.unlinkProvider')}
                    </Button>
                  </div>
                ))}
                {unlinkedProviders.map((p) => (
                  <div key={p.id} className="flex items-center justify-between">
                    <p className="text-sm font-medium text-foreground">{p.display_name}</p>
                    <a
                      href={`/api/v1/auth/oauth/${p.id}/link`}
                      className="inline-flex items-center justify-center rounded-lg border border-input bg-background px-3 py-1.5 text-xs font-medium text-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
                    >
                      {t('settings.connectProvider', { provider: p.display_name })}
                    </a>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </div>
    </section>
  )
}
