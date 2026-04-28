import { Link, Outlet } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { ImagePlus } from 'lucide-react'

import { LanguageSwitcher } from '../../components/language-switcher'
import { ThemeSwitcher } from '../../components/theme-switcher'
import { useAuth } from '../../lib/auth-context'

export function PublicLayout() {
  const { t } = useTranslation()
  const { isAuthenticated } = useAuth()

  return (
    <div className="relative min-h-screen bg-background text-foreground overflow-hidden">
      {/* Premium Background Glow Effect */}
      <div className="pointer-events-none fixed inset-0 z-0">
        <div className="absolute top-[-10%] left-[-10%] h-[40%] w-[40%] rounded-full bg-primary/20 blur-[120px]" />
        <div className="absolute bottom-[-10%] right-[-10%] h-[40%] w-[40%] rounded-full bg-info/20 blur-[120px]" />
      </div>

      <header className="sticky top-0 z-50 w-full border-b border-border/40 bg-background/60 backdrop-blur-xl supports-[backdrop-filter]:bg-background/60">
        <div className="mx-auto flex h-14 w-full max-w-6xl items-center justify-between px-6">
          <Link to="/" className="flex items-center gap-3 transition-opacity hover:opacity-80 active:scale-95">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-sm">
              <ImagePlus className="h-5 w-5" />
            </div>
            <span className="text-xl font-bold tracking-tight bg-clip-text text-transparent bg-gradient-to-r from-foreground to-foreground/70 dark:from-white dark:to-white/60">
              {t('appName')}
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
                <Link to="/register" className="inline-flex h-8 items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground shadow transition-colors hover:bg-primary/90">
                  {t('nav.register')}
                </Link>
              </div>
            )}
          </div>
        </div>
      </header>

      <main className="relative z-10 mx-auto w-full max-w-6xl px-6 py-12 md:py-20 animate-in fade-in duration-500">
        <Outlet />
      </main>
    </div>
  )
}
