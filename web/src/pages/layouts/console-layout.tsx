import type { ReactNode } from 'react'
import { Link, NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { 
  UploadCloud, 
  Image as ImageIcon, 
  FolderOpen, 
  KeySquare, 
  Blocks, 
  Settings, 
  LayoutDashboard, 
  Users, 
  UsersRound, 
  Database, 
  Files, 
  Globe,
  ScrollText,
  ShieldCheck,
  BarChart3
} from 'lucide-react'

import { Button } from '@/components/ui/button'
import { LanguageSwitcher } from '../../components/language-switcher'
import { ThemeSwitcher } from '../../components/theme-switcher'
import { useAuth } from '../../lib/auth-context'
import { getSiteConfig } from '../../lib/site-config'

export function ConsoleLayout() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { user, logout } = useAuth()
  const { data: config } = useQuery({ queryKey: ['site-config'], queryFn: getSiteConfig })
  const appName = config?.app_name?.trim() || t('appName')
  const logoSrc = config?.favicon_url?.trim() || '/favicon-default.svg'

  const onLogout = async () => {
    await logout()
    navigate('/login', { replace: true })
  }

  return (
    <div className="relative min-h-screen bg-background overflow-x-hidden text-foreground">
      <header className="sticky top-0 z-50 w-full border-b border-border/40 bg-background/60 backdrop-blur-xl supports-[backdrop-filter]:bg-background/60">
        <div className="mx-auto flex h-14 w-full max-w-[1400px] items-center justify-between px-6">
          <Link to="/console/upload" className="flex items-center gap-2 text-lg font-bold tracking-tight transition-opacity hover:opacity-80">
            <div className="flex h-8 w-8 items-center justify-center overflow-hidden rounded-lg shadow-sm">
              <img
                src={logoSrc}
                alt="logo"
                className="h-full w-full object-contain"
                onError={(e) => {
                  e.currentTarget.src = '/favicon-default.svg'
                }}
              />
            </div>
            <span className="bg-gradient-to-r from-primary to-primary/60 bg-clip-text text-transparent">
              {appName}
            </span>
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

      <div className="relative z-10 mx-auto flex flex-col md:grid w-full max-w-[1400px] md:grid-cols-[220px_1fr] lg:grid-cols-[240px_1fr] gap-6 md:gap-10 px-4 md:px-6 py-6 md:py-8">
        <aside className="w-full md:block">
          <div className="md:sticky md:top-24 space-y-1">
            <MobileNavRail>
              <ConsoleNavItem to="/console/upload" label={t('nav.upload')} icon={UploadCloud} />
              <ConsoleNavItem to="/console/images" label={t('nav.images')} icon={ImageIcon} />
              <ConsoleNavItem to="/console/albums" label={t('nav.albums')} icon={FolderOpen} />
              <ConsoleNavItem to="/console/api-tokens" label={t('nav.apiTokens')} icon={KeySquare} />
              <ConsoleNavItem to="/console/integrations" label={t('connections.title', { defaultValue: '接入' })} icon={Blocks} />
              <ConsoleNavItem to="/console/settings" label={t('nav.settings')} icon={Settings} />
            </MobileNavRail>
            {user?.role === 'admin' && (
              <div className="mt-4 pt-4 border-t border-border/40">
                <p className="px-3 pb-2 text-[11px] font-bold uppercase tracking-wider text-muted-foreground/60 hidden md:block">
                  {t('nav.admin')}
                </p>
                <MobileNavRail>
                  <ConsoleNavItem to="/console/admin" label={t('admin.navDashboard', { defaultValue: '概览' })} icon={LayoutDashboard} />
                  <ConsoleNavItem to="/console/admin/users" label={t('admin.navUsers')} icon={Users} />
                  <ConsoleNavItem to="/console/admin/groups" label={t('admin.navGroups')} icon={UsersRound} />
                  <ConsoleNavItem to="/console/admin/strategies" label={t('admin.navStrategies')} icon={Database} />
                  <ConsoleNavItem to="/console/admin/images" label={t('admin.navImages')} icon={Files} />
                  <ConsoleNavItem to="/console/admin/audit-logs" label={t('admin.navAuditLogs', { defaultValue: '审计日志' })} icon={ScrollText} />
                  <ConsoleNavItem to="/console/admin/site" label={t('admin.navSiteSettings')} icon={Globe} />
                  <ConsoleNavItem to="/console/admin/access" label={t('admin.navAccessSettings')} icon={ShieldCheck} />
                  <ConsoleNavItem to="/console/admin/compliance" label={t('admin.navComplianceSettings')} icon={ScrollText} />
                  <ConsoleNavItem to="/console/admin/analytics" label={t('admin.navAnalyticsSettings')} icon={BarChart3} />
                </MobileNavRail>
              </div>
            )}
          </div>
        </aside>
        <main className="min-w-0">
          <div className="min-h-[calc(100vh-8rem)] pt-2 md:pt-0">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  )
}

function MobileNavRail({ children }: { children: ReactNode }) {
  return (
    <div className="relative">
      <div className="flex overflow-x-auto md:flex-col gap-2 pb-2 md:pb-0 scrollbar-hide -mx-4 px-4 md:mx-0 md:px-0">
        {children}
      </div>
      <div className="pointer-events-none absolute inset-y-0 left-0 w-6 bg-linear-to-r from-background via-background/80 to-transparent md:hidden" />
      <div className="pointer-events-none absolute inset-y-0 right-0 w-6 bg-linear-to-l from-background via-background/80 to-transparent md:hidden" />
    </div>
  )
}

function ConsoleNavItem({ to, label, icon: Icon }: { to: string; label: string; icon: React.ElementType }) {
  return (
    <NavLink
      to={to}
      end={to === '/console/admin'}
      className={({ isActive }) =>
        [
          'group relative flex items-center gap-3 overflow-hidden rounded-xl px-3 py-2 text-[14px] font-medium transition-colors duration-200',
          isActive
            ? 'text-foreground bg-muted/50 font-semibold'
            : 'text-muted-foreground hover:bg-muted/30 hover:text-foreground',
        ].join(' ')
      }
    >
      <Icon className="size-4 shrink-0" />
      <span className="relative z-10">{label}</span>
    </NavLink>
  )
}
