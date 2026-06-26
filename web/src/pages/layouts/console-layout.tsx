import { useState } from 'react'
import { Link, NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import {
  UploadCloud,
  Image as ImageIcon,
  FolderOpen,
  KeySquare,
  Blocks,
  UserRound,
  LayoutDashboard,
  Users,
  UsersRound,
  Database,
  Files,
  Globe,
  ScrollText,
  ShieldCheck,
  ShieldAlert,
  Menu,
} from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogTitle } from '@/components/ui/dialog'
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
  const [mobileNavOpen, setMobileNavOpen] = useState(false)

  const onLogout = async () => {
    await logout()
    navigate('/login', { replace: true })
  }

  return (
    <div className="relative flex-1 flex flex-col overflow-x-hidden">
      <header className="sticky top-0 z-50 w-full border-b border-border/60 bg-card/80 backdrop-blur-xl supports-[backdrop-filter]:bg-card/80 shadow-[0_1px_2px_rgba(0,0,0,0.02)] dark:shadow-[0_1px_2px_rgba(0,0,0,0.2)]">
        <div className="mx-auto flex h-14 w-full max-w-[1400px] items-center justify-between px-4 md:px-6">
          <div className="flex min-w-0 items-center gap-2">
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              className="md:hidden"
              onClick={() => setMobileNavOpen(true)}
              title={t('nav.menu', { defaultValue: '菜单' })}
            >
              <Menu className="size-4" />
              <span className="sr-only">{t('nav.menu', { defaultValue: '菜单' })}</span>
            </Button>
            <Link to="/console/upload" className="flex min-w-0 items-center gap-2 text-lg font-bold tracking-tight transition-opacity hover:opacity-80">
            <div className="pf-site-logo h-7 w-7 shrink-0 overflow-hidden rounded-md shadow-sm border border-border/50 bg-background">
              <img
                src={logoSrc}
                alt="logo"
                className="block size-full object-cover"
                onError={(e) => {
                  e.currentTarget.src = '/favicon-default.svg'
                }}
              />
            </div>
            <span className="truncate text-[14px] font-semibold text-foreground sm:text-[15px]">
              {appName}
            </span>
          </Link>
          </div>
          <div className="flex items-center gap-3 md:gap-5">
            <div className="flex items-center gap-2">
              <LanguageSwitcher />
              <ThemeSwitcher />
            </div>
            <div className="h-4 w-px bg-border/80 hidden sm:block" />
            <div className="flex items-center gap-3">
              {user && (
                <span className="hidden max-w-[180px] truncate text-[13px] font-medium text-foreground sm:inline-block">
                  {user.name || user.email}
                </span>
              )}
              <Button variant="outline" size="sm" onClick={onLogout} className="h-8 px-3 text-xs rounded-md shadow-sm hover:bg-destructive/10 hover:text-destructive hover:border-destructive/30 transition-colors">
                {t('common.logout')}
              </Button>
            </div>
          </div>
        </div>
      </header>

      <Dialog open={mobileNavOpen} onOpenChange={setMobileNavOpen}>
        <DialogContent
          showCloseButton={false}
          className="left-0 top-0 h-dvh max-h-dvh w-[86vw] max-w-[320px] translate-x-0 translate-y-0 rounded-none border-r border-border/60 bg-card p-4 pt-5"
        >
          <DialogTitle className="sr-only">{t('nav.menu', { defaultValue: '菜单' })}</DialogTitle>
          <div className="flex h-full flex-col">
            <div className="mb-4 flex items-center justify-between">
              <span className="text-sm font-semibold text-foreground">{t('nav.menu', { defaultValue: '菜单' })}</span>
              <Button type="button" size="sm" variant="outline" onClick={() => setMobileNavOpen(false)}>
                {t('common.close', { defaultValue: '关闭' })}
              </Button>
            </div>
            <div className="space-y-2 overflow-y-auto pr-1">
              <ConsoleNavItem to="/console/upload" label={t('nav.upload')} icon={UploadCloud} onNavigate={() => setMobileNavOpen(false)} />
              <ConsoleNavItem to="/console/images" label={t('nav.images')} icon={ImageIcon} onNavigate={() => setMobileNavOpen(false)} />
              <ConsoleNavItem to="/console/albums" label={t('nav.albums')} icon={FolderOpen} onNavigate={() => setMobileNavOpen(false)} />
              <ConsoleNavItem to="/console/api-tokens" label={t('nav.apiTokens')} icon={KeySquare} onNavigate={() => setMobileNavOpen(false)} />
              <ConsoleNavItem to="/console/integrations" label={t('connections.title', { defaultValue: '接入' })} icon={Blocks} onNavigate={() => setMobileNavOpen(false)} />
              <ConsoleNavItem to="/console/account" label={t('nav.account')} icon={UserRound} onNavigate={() => setMobileNavOpen(false)} />
              {user?.role === 'admin' && (
                <>
                  <p className="mt-4 px-3 text-xs font-bold uppercase tracking-wider text-muted-foreground/80">
                    {t('nav.admin')}
                  </p>
                  <ConsoleNavItem to="/console/admin" label={t('admin.navDashboard', { defaultValue: '概览' })} icon={LayoutDashboard} onNavigate={() => setMobileNavOpen(false)} />
                  <ConsoleNavItem to="/console/admin/users" label={t('admin.navUsers')} icon={Users} onNavigate={() => setMobileNavOpen(false)} />
                  <ConsoleNavItem to="/console/admin/groups" label={t('admin.navGroups')} icon={UsersRound} onNavigate={() => setMobileNavOpen(false)} />
                  <ConsoleNavItem to="/console/admin/strategies" label={t('admin.navStrategies')} icon={Database} onNavigate={() => setMobileNavOpen(false)} />
                  <ConsoleNavItem to="/console/admin/images" label={t('admin.navImages')} icon={Files} onNavigate={() => setMobileNavOpen(false)} />
                  <ConsoleNavItem to="/console/admin/moderation" label={t('admin.navModeration', { defaultValue: '审核管理' })} icon={ShieldAlert} onNavigate={() => setMobileNavOpen(false)} />
                  <ConsoleNavItem to="/console/admin/audit-logs" label={t('admin.navAuditLogs', { defaultValue: '审计日志' })} icon={ScrollText} onNavigate={() => setMobileNavOpen(false)} />
                  <ConsoleNavItem to="/console/admin/site" label={t('admin.navSiteSettings')} icon={Globe} onNavigate={() => setMobileNavOpen(false)} />
                  <ConsoleNavItem to="/console/admin/access" label={t('admin.navAccessSettings')} icon={ShieldCheck} onNavigate={() => setMobileNavOpen(false)} />
                </>
              )}
            </div>
          </div>
        </DialogContent>
      </Dialog>

      <div className="pf-console-shell relative z-10 mx-auto flex flex-col md:grid w-full max-w-[1400px] md:grid-cols-[200px_1fr] gap-6 md:gap-8 px-4 md:px-6 py-6 md:py-8 flex-1">
        <aside className="hidden w-full md:block">
          <div className="md:sticky md:top-24 space-y-1">
            <div className="space-y-2">
              <ConsoleNavItem to="/console/upload" label={t('nav.upload')} icon={UploadCloud} />
              <ConsoleNavItem to="/console/images" label={t('nav.images')} icon={ImageIcon} />
              <ConsoleNavItem to="/console/albums" label={t('nav.albums')} icon={FolderOpen} />
              <ConsoleNavItem to="/console/api-tokens" label={t('nav.apiTokens')} icon={KeySquare} />
              <ConsoleNavItem to="/console/integrations" label={t('connections.title', { defaultValue: '接入' })} icon={Blocks} />
              <ConsoleNavItem to="/console/account" label={t('nav.account')} icon={UserRound} />
            </div>
            {user?.role === 'admin' && (
              <div className="mt-4 pt-4 border-t border-border/40">
                <p className="px-3 pb-2 text-xs font-bold uppercase tracking-wider text-muted-foreground/80 hidden md:block">
                  {t('nav.admin')}
                </p>
                <div className="space-y-2">
                  <ConsoleNavItem to="/console/admin" label={t('admin.navDashboard', { defaultValue: '概览' })} icon={LayoutDashboard} />
                  <ConsoleNavItem to="/console/admin/users" label={t('admin.navUsers')} icon={Users} />
                  <ConsoleNavItem to="/console/admin/groups" label={t('admin.navGroups')} icon={UsersRound} />
                  <ConsoleNavItem to="/console/admin/strategies" label={t('admin.navStrategies')} icon={Database} />
                  <ConsoleNavItem to="/console/admin/images" label={t('admin.navImages')} icon={Files} />
                  <ConsoleNavItem to="/console/admin/moderation" label={t('admin.navModeration', { defaultValue: '审核管理' })} icon={ShieldAlert} />
                  <ConsoleNavItem to="/console/admin/audit-logs" label={t('admin.navAuditLogs', { defaultValue: '审计日志' })} icon={ScrollText} />
                  <ConsoleNavItem to="/console/admin/site" label={t('admin.navSiteSettings')} icon={Globe} />
                  <ConsoleNavItem to="/console/admin/access" label={t('admin.navAccessSettings')} icon={ShieldCheck} />
                </div>
              </div>
            )}
          </div>
        </aside>
        <main className="min-w-0 flex flex-col">
          <div className="pf-console-content rounded-2xl border border-border/40 bg-card p-4 sm:p-6 md:p-8 shadow-sm flex-1">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  )
}

function ConsoleNavItem({
  to,
  label,
  icon: Icon,
  onNavigate,
}: {
  to: string
  label: string
  icon: React.ElementType
  onNavigate?: () => void
}) {
  return (
    <NavLink
      to={to}
      end={to === '/console/admin'}
      onClick={onNavigate}
      className={({ isActive }) =>
        [
          'group relative flex items-center gap-3 overflow-hidden rounded-xl px-3 py-2.5 text-[13px] font-medium transition-colors duration-200',
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
