import { Link, Outlet } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

import { LanguageSwitcher } from '../../components/language-switcher'
import { ThemeSwitcher } from '../../components/theme-switcher'
import { useAuth } from '../../lib/auth-context'

export function PublicLayout() {
  const { t } = useTranslation()
  const { isAuthenticated } = useAuth()

  return (
    <div className="min-h-screen bg-zinc-50 text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100">
      <header className="border-b border-zinc-200 bg-white/90 px-6 py-3 backdrop-blur dark:border-zinc-800 dark:bg-zinc-950/90">
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
                className="text-sm text-zinc-700 hover:text-zinc-950 dark:text-zinc-300"
              >
                {t('nav.console')}
              </Link>
            ) : (
              <>
                <Link to="/login" className="text-sm text-zinc-700 hover:text-zinc-950 dark:text-zinc-300">
                  {t('nav.login')}
                </Link>
                <Link to="/register" className="text-sm text-zinc-700 hover:text-zinc-950 dark:text-zinc-300">
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
