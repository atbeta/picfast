import { Link, NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

import { LanguageSwitcher } from '../../components/language-switcher'
import { ThemeSwitcher } from '../../components/theme-switcher'
import { ACCESS_TOKEN_KEY, REFRESH_TOKEN_KEY } from '../../lib/api'

export function ConsoleLayout() {
  const { t } = useTranslation()
  const navigate = useNavigate()

  const onLogout = () => {
    localStorage.removeItem(ACCESS_TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
    navigate('/login')
  }

  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-zinc-950">
      <header className="border-b border-zinc-200 bg-white px-6 py-3 dark:border-zinc-800 dark:bg-zinc-950">
        <div className="mx-auto flex w-full max-w-7xl items-center justify-between">
          <Link to="/console/upload" className="text-lg font-semibold text-zinc-900 dark:text-zinc-100">
            {t('appName')}
          </Link>
          <div className="flex items-center gap-3">
            <LanguageSwitcher />
            <ThemeSwitcher />
            <button
              type="button"
              onClick={onLogout}
              className="rounded-md border border-zinc-300 px-3 py-1.5 text-sm hover:bg-zinc-100 dark:border-zinc-700 dark:hover:bg-zinc-900"
            >
              {t('common.logout')}
            </button>
          </div>
        </div>
      </header>

      <div className="mx-auto grid w-full max-w-7xl grid-cols-[220px_1fr] gap-6 px-6 py-6">
        <aside className="space-y-1">
          <ConsoleNavItem to="/console/upload" label={t('nav.upload')} />
          <ConsoleNavItem to="/console/images" label={t('nav.images')} />
          <ConsoleNavItem to="/console/albums" label={t('nav.albums')} />
          <ConsoleNavItem to="/console/api-tokens" label={t('nav.apiTokens')} />
          <ConsoleNavItem to="/console/settings" label={t('nav.settings')} />
        </aside>
        <main className="rounded-xl border border-zinc-200 bg-white p-6 dark:border-zinc-800 dark:bg-zinc-900">
          <Outlet />
        </main>
      </div>
    </div>
  )
}

function ConsoleNavItem({ to, label }: { to: string; label: string }) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        [
          'block rounded-md px-3 py-2 text-sm transition',
          isActive
            ? 'bg-zinc-900 text-white dark:bg-zinc-100 dark:text-zinc-900'
            : 'text-zinc-700 hover:bg-zinc-100 dark:text-zinc-300 dark:hover:bg-zinc-800',
        ].join(' ')
      }
    >
      {label}
    </NavLink>
  )
}
