import { Link, NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { LanguageSwitcher } from '../../components/language-switcher'
import { ThemeSwitcher } from '../../components/theme-switcher'
import { useAuth } from '../../lib/auth-context'

export function ConsoleLayout() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { user, logout } = useAuth()

  const onLogout = async () => {
    await logout()
    navigate('/login', { replace: true })
  }

  return (
    <div className="min-h-screen bg-muted/50">
      <header className="border-b border-border bg-card px-6 py-3">
        <div className="mx-auto flex w-full max-w-7xl items-center justify-between">
          <Link to="/console/upload" className="text-lg font-semibold text-foreground">
            {t('appName')}
          </Link>
          <div className="flex items-center gap-3">
            <LanguageSwitcher />
            <ThemeSwitcher />
            {user && (
              <span className="text-sm text-muted-foreground">{user.name || user.email}</span>
            )}
            <Button variant="outline" size="sm" onClick={onLogout}>
              {t('common.logout')}
            </Button>
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
          {user?.role === 'admin' && (
            <>
              <div className="my-2 border-t border-border" />
              <p className="px-3 pb-1 text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('nav.admin')}</p>
              <ConsoleNavItem to="/console/admin" label={t('admin.navDashboard', { defaultValue: '概览' })} />
              <ConsoleNavItem to="/console/admin/users" label={t('admin.navUsers')} />
              <ConsoleNavItem to="/console/admin/groups" label={t('admin.navGroups')} />
              <ConsoleNavItem to="/console/admin/strategies" label={t('admin.navStrategies')} />
              <ConsoleNavItem to="/console/admin/images" label={t('admin.navImages')} />
              <ConsoleNavItem to="/console/admin/settings" label={t('admin.navSettings')} />
            </>
          )}
        </aside>
        <main className="rounded-xl border border-border bg-card p-6 text-card-foreground">
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
      end={to === '/console/admin'}
      className={({ isActive }) =>
        [
          'block rounded-md px-3 py-2 text-sm transition-colors',
          isActive
            ? 'bg-primary text-primary-foreground'
            : 'text-muted-foreground hover:bg-muted hover:text-foreground',
        ].join(' ')
      }
    >
      {label}
    </NavLink>
  )
}
