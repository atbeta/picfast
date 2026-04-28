import { Link, NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { ImagePlus } from 'lucide-react'

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
    <div className="relative min-h-screen bg-background overflow-x-hidden text-foreground">
      {/* Subtle Background Elements */}
      <div className="pointer-events-none fixed inset-0 z-0">
        <div className="absolute top-0 left-[20%] h-[30%] w-[30%] rounded-full bg-primary/10 blur-[100px]" />
      </div>

      <header className="sticky top-0 z-50 w-full border-b border-border/40 bg-background/60 backdrop-blur-xl supports-[backdrop-filter]:bg-background/60">
        <div className="mx-auto flex h-14 w-full max-w-[1400px] items-center justify-between px-6">
          <Link to="/console/upload" className="flex items-center gap-2 text-lg font-bold tracking-tight transition-opacity hover:opacity-80">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground shadow-sm">
              <ImagePlus className="h-5 w-5" />
            </div>
            {t('appName')}
          </Link>
          <div className="flex items-center gap-4">
            <LanguageSwitcher />
            <ThemeSwitcher />
            <div className="h-4 w-px bg-border/50 hidden sm:block" />
            {user && (
              <span className="text-sm font-medium text-muted-foreground">{user.name || user.email}</span>
            )}
            <Button variant="outline" size="sm" onClick={onLogout} className="rounded-full shadow-sm hover:bg-destructive/10 hover:text-destructive hover:border-destructive/30 transition-colors">
              {t('common.logout')}
            </Button>
          </div>
        </div>
      </header>

      <div className="relative z-10 mx-auto flex flex-col md:grid w-full max-w-[1400px] md:grid-cols-[240px_1fr] gap-6 md:gap-8 px-4 md:px-6 py-6 md:py-8">
        <aside className="w-full md:block">
          <div className="md:sticky md:top-24 space-y-1">
            <div className="flex overflow-x-auto md:flex-col gap-2 pb-2 md:pb-0 scrollbar-hide -mx-4 px-4 md:mx-0 md:px-0">
              <ConsoleNavItem to="/console/upload" label={t('nav.upload')} />
              <ConsoleNavItem to="/console/images" label={t('nav.images')} />
              <ConsoleNavItem to="/console/albums" label={t('nav.albums')} />
              <ConsoleNavItem to="/console/api-tokens" label={t('nav.apiTokens')} />
              <ConsoleNavItem to="/console/settings" label={t('nav.settings')} />
            </div>
            {user?.role === 'admin' && (
              <div className="pt-2 md:pt-4">
                <p className="px-4 pb-2 text-xs font-bold uppercase tracking-wider text-muted-foreground/70 hidden md:block">{t('nav.admin')}</p>
                <div className="flex overflow-x-auto md:flex-col gap-2 pb-2 md:pb-0 scrollbar-hide -mx-4 px-4 md:mx-0 md:px-0">
                  <ConsoleNavItem to="/console/admin" label={t('admin.navDashboard', { defaultValue: '概览' })} />
                  <ConsoleNavItem to="/console/admin/users" label={t('admin.navUsers')} />
                  <ConsoleNavItem to="/console/admin/groups" label={t('admin.navGroups')} />
                  <ConsoleNavItem to="/console/admin/strategies" label={t('admin.navStrategies')} />
                  <ConsoleNavItem to="/console/admin/images" label={t('admin.navImages')} />
                  <ConsoleNavItem to="/console/admin/settings" label={t('admin.navSettings')} />
                </div>
              </div>
            )}
          </div>
        </aside>
        <main className="min-w-0">
          <div className="min-h-[calc(100vh-8rem)] rounded-2xl border border-border/50 bg-card/40 backdrop-blur-md p-4 sm:p-6 md:p-8 shadow-sm">
            <Outlet />
          </div>
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
          'group relative flex items-center rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 overflow-hidden',
          isActive
            ? 'bg-primary/10 text-primary'
            : 'text-muted-foreground hover:bg-muted/50 hover:text-foreground',
        ].join(' ')
      }
    >
      {({ isActive }) => (
        <>
          {isActive && (
            <span className="absolute left-0 top-1/2 h-1/2 w-1 -translate-y-1/2 rounded-r-full bg-primary" />
          )}
          <span className="relative z-10">{label}</span>
        </>
      )}
    </NavLink>
  )
}
