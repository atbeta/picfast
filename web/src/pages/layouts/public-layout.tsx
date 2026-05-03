import { Link, Outlet } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'

import { LanguageSwitcher } from '../../components/language-switcher'
import { ThemeSwitcher } from '../../components/theme-switcher'
import { useAuth } from '../../lib/auth-context'
import { getSiteConfig } from '../../lib/site-config'

export function PublicLayout() {
  const { t } = useTranslation()
  const { isAuthenticated } = useAuth()
  const { data: config } = useQuery({ queryKey: ['site-config'], queryFn: getSiteConfig })
  const appName = config?.app_name?.trim() || t('appName')
  const logoSrc = config?.favicon_url?.trim() || '/favicon-default.svg'

  return (
    <div className="relative flex-1 flex flex-col overflow-hidden">
      {/* Premium Background Glow Effect */}
      <div className="pointer-events-none fixed inset-0 z-0">
        <div className="absolute top-[-10%] left-[-10%] h-[40%] w-[40%] rounded-full bg-primary/20 blur-[120px]" />
        <div className="absolute bottom-[-10%] right-[-10%] h-[40%] w-[40%] rounded-full bg-info/20 blur-[120px]" />
      </div>

      <header className="sticky top-0 z-50 w-full border-b border-border/40 bg-background/60 backdrop-blur-xl supports-[backdrop-filter]:bg-background/60">
        <div className="mx-auto flex h-14 w-full max-w-6xl items-center justify-between px-6">
          <Link to="/" className="flex items-center gap-3 transition-opacity duration-150 hover:opacity-80">
            <div className="h-9 w-9 shrink-0 overflow-hidden rounded-xl shadow-sm">
              <img
                src={logoSrc}
                alt="logo"
                className="block size-full object-cover"
                onError={(e) => {
                  e.currentTarget.src = '/favicon-default.svg'
                }}
              />
            </div>
            <span className="text-xl font-bold tracking-tight bg-clip-text text-transparent bg-gradient-to-r from-foreground to-foreground/70 dark:from-white dark:to-white/60">
              {appName}
            </span>
          </Link>
          <div className="flex items-center gap-4">
            <LanguageSwitcher />
            <ThemeSwitcher />
            <div className="h-4 w-px bg-border/50 hidden sm:block" />
            {isAuthenticated ? (
              <Link
                to="/console"
                className="text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
              >
                {t('nav.console')}
              </Link>
            ) : (
              <div className="flex items-center gap-3">
                <Link to="/login" className="text-sm font-medium text-muted-foreground transition-colors hover:text-foreground">
                  {t('nav.login')}
                </Link>
                <Link to="/register" className="inline-flex h-8 items-center justify-center rounded-lg bg-primary px-4 text-sm font-medium text-primary-foreground shadow transition-colors hover:bg-primary/90">
                  {t('nav.register')}
                </Link>
              </div>
            )}
          </div>
        </div>
      </header>

      <main className="relative z-10 mx-auto w-full max-w-6xl px-6 py-12 md:py-20">
        <Outlet />
      </main>
    </div>
  )
}
