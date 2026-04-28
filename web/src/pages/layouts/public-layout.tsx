import { Link, Outlet } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

import { LanguageSwitcher } from '../../components/language-switcher'
import { ThemeSwitcher } from '../../components/theme-switcher'
import { useAuth } from '../../lib/auth-context'

export function PublicLayout() {
  const { t } = useTranslation()
  const { isAuthenticated } = useAuth()

  return (
    <div className="min-h-screen bg-muted/50 text-foreground">
      <header className="border-b border-border bg-card/90 px-6 py-3 backdrop-blur">
        <div className="mx-auto flex w-full max-w-6xl items-center justify-between">
          <Link to="/" className="text-lg font-semibold">
            {t('appName')}
          </Link>
          <div className="flex items-center gap-3">
            <LanguageSwitcher />
            <ThemeSwitcher />
            {isAuthenticated ? (
              <Link
                to="/console"
                className="text-sm text-muted-foreground hover:text-foreground transition-colors"
              >
                {t('nav.console')}
              </Link>
            ) : (
              <>
                <Link to="/login" className="text-sm text-muted-foreground hover:text-foreground transition-colors">
                  {t('nav.login')}
                </Link>
                <Link to="/register" className="text-sm text-muted-foreground hover:text-foreground transition-colors">
                  {t('nav.register')}
                </Link>
              </>
            )}
          </div>
        </div>
      </header>
      <main className="mx-auto w-full max-w-6xl px-6 py-8">
        <Outlet />
      </main>
    </div>
  )
}
