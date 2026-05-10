import { Suspense, lazy, type ReactNode } from 'react'
import { Navigate, Outlet, Route, Routes, useLocation } from 'react-router-dom'

import { useQuery } from '@tanstack/react-query'

import { LoadingState } from '../components/page-states'
import { SiteFooter } from '../components/site-footer'
import { SiteMetadata } from '../components/site-metadata'
import { useAuth } from '../lib/auth-context'
import { getSiteConfig, type SiteConfig } from '../lib/site-config'
import { ConsoleLayout } from '../pages/layouts/console-layout'
import { PublicLayout } from '../pages/layouts/public-layout'

const AlbumsPage = lazy(async () => ({ default: (await import('../pages/console/albums-page')).AlbumsPage }))
const ApiTokensPage = lazy(async () => ({ default: (await import('../pages/console/api-tokens-page')).ApiTokensPage }))
const ImagesPage = lazy(async () => ({ default: (await import('../pages/console/images-page')).ImagesPage }))
const SettingsPage = lazy(async () => ({ default: (await import('../pages/console/settings-page')).SettingsPage }))
const IntegrationsPage = lazy(async () => ({ default: (await import('../pages/console/integrations-page')).IntegrationsPage }))
const UploadPage = lazy(async () => ({ default: (await import('../pages/console/upload-page')).UploadPage }))
const AdminUsersPage = lazy(async () => ({ default: (await import('../pages/console/admin/users-page')).AdminUsersPage }))
const AdminGroupsPage = lazy(async () => ({ default: (await import('../pages/console/admin/groups-page')).AdminGroupsPage }))
const AdminStrategiesPage = lazy(async () => ({ default: (await import('../pages/console/admin/strategies-page')).AdminStrategiesPage }))
const AdminImagesPage = lazy(async () => ({ default: (await import('../pages/console/admin/images-page')).AdminImagesPage }))
const AdminModerationPage = lazy(async () => ({ default: (await import('../pages/console/admin/moderation-page')).AdminModerationPage }))
const AdminMaintenancePage = lazy(async () => ({ default: (await import('../pages/console/admin/maintenance-page')).AdminMaintenancePage }))
const AdminSiteSettingsPage = lazy(async () => ({ default: (await import('../pages/console/admin/settings/site-settings-page')).AdminSiteSettingsPage }))
const AdminAccessSettingsPage = lazy(async () => ({ default: (await import('../pages/console/admin/settings/access-settings-page')).AdminAccessSettingsPage }))
const AdminAnalyticsSettingsPage = lazy(async () => ({ default: (await import('../pages/console/admin/settings/analytics-settings-page')).AdminAnalyticsSettingsPage }))
const AdminAppearanceSettingsPage = lazy(async () => ({ default: (await import('../pages/console/admin/settings/appearance-settings-page')).AdminAppearanceSettingsPage }))
const AdminDashboardPage = lazy(async () => ({ default: (await import('../pages/console/admin/dashboard-page')).AdminDashboardPage }))
const AdminAuditLogsPage = lazy(async () => ({ default: (await import('../pages/console/admin/audit-logs-page')).AdminAuditLogsPage }))
const GuestUploadPage = lazy(async () => ({ default: (await import('../pages/public/guest-upload-page')).GuestUploadPage }))
const LoginPage = lazy(async () => ({ default: (await import('../pages/public/login-page')).LoginPage }))
const ForgotPasswordPage = lazy(async () => ({ default: (await import('../pages/public/forgot-password-page')).ForgotPasswordPage }))
const RegisterPage = lazy(async () => ({ default: (await import('../pages/public/register-page')).RegisterPage }))
const ResetPasswordPage = lazy(async () => ({ default: (await import('../pages/public/reset-password-page')).ResetPasswordPage }))
const SetupPage = lazy(async () => ({ default: (await import('../pages/public/setup-page')).SetupPage }))
const VerifyEmailPage = lazy(async () => ({ default: (await import('../pages/public/verify-email-page')).VerifyEmailPage }))

function PageLoader() {
  return <LoadingState className="min-h-[40vh]" compact />
}

function LazyPage({ children }: { children: ReactNode }) {
  return <Suspense fallback={<PageLoader />}>{children}</Suspense>
}

function RequireAuth() {
  const { isAuthenticated, isLoading } = useAuth()
  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    )
  }
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }
  return <Outlet />
}

function RequireAdmin() {
  const { user, isLoading } = useAuth()
  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    )
  }
  if (user?.role !== 'admin') {
    return <Navigate to="/console" replace />
  }
  return <Outlet />
}

function PublicRoutes() {
  const location = useLocation()
  const { data: config, isLoading } = useQuery<SiteConfig>({
    queryKey: ['site-config'],
    queryFn: getSiteConfig,
  })

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    )
  }

  if (config?.setup_required) {
    return (
      <div className="flex min-h-screen flex-col bg-background text-foreground">
        <SiteMetadata config={config} />
        <div className="flex flex-1 items-center px-4 py-10">
          <Routes>
            <Route path="/setup" element={<LazyPage><SetupPage /></LazyPage>} />
            <Route path="*" element={<Navigate to="/setup" replace />} />
          </Routes>
        </div>
        <SiteFooter config={config} />
      </div>
    )
  }

  if (location.pathname === '/setup') {
    return <Navigate to="/login" replace />
  }

  return (
    <div className="flex min-h-screen flex-col bg-background text-foreground">
      {config && <SiteMetadata config={config} />}
      <div className="flex flex-1 flex-col">
        <Routes>
          <Route element={<PublicLayout />}>
            <Route
              path="/"
              element={
                config?.allow_guest_upload ? (
                  <LazyPage>
                    <GuestUploadPage />
                  </LazyPage>
                ) : (
                  <Navigate to="/login" replace />
                )
              }
            />
            <Route path="/login" element={<LazyPage><LoginPage /></LazyPage>} />
            <Route path="/forgot-password" element={<LazyPage><ForgotPasswordPage /></LazyPage>} />
            <Route path="/reset-password" element={<LazyPage><ResetPasswordPage /></LazyPage>} />
            <Route path="/verify-email" element={<LazyPage><VerifyEmailPage /></LazyPage>} />
            <Route
              path="/register"
              element={
                config?.allow_registration ? (
                  <LazyPage>
                    <RegisterPage />
                  </LazyPage>
                ) : (
                  <Navigate to="/login" replace />
                )
              }
            />
          </Route>

          <Route element={<RequireAuth />}>
            <Route path="/console" element={<ConsoleLayout />}>
              <Route path="upload" element={<LazyPage><UploadPage /></LazyPage>} />
              <Route path="images" element={<LazyPage><ImagesPage /></LazyPage>} />
              <Route path="albums" element={<LazyPage><AlbumsPage /></LazyPage>} />
              <Route path="api-tokens" element={<LazyPage><ApiTokensPage /></LazyPage>} />
              <Route path="integrations" element={<LazyPage><IntegrationsPage /></LazyPage>} />
              <Route path="settings" element={<LazyPage><SettingsPage /></LazyPage>} />
              <Route path="admin" element={<RequireAdmin />}>
                <Route index element={<LazyPage><AdminDashboardPage /></LazyPage>} />
                <Route path="users" element={<LazyPage><AdminUsersPage /></LazyPage>} />
                <Route path="groups" element={<LazyPage><AdminGroupsPage /></LazyPage>} />
                <Route path="strategies" element={<LazyPage><AdminStrategiesPage /></LazyPage>} />
                <Route path="images" element={<LazyPage><AdminImagesPage /></LazyPage>} />
                <Route path="moderation" element={<LazyPage><AdminModerationPage /></LazyPage>} />
                <Route path="audit-logs" element={<LazyPage><AdminAuditLogsPage /></LazyPage>} />
                <Route path="maintenance" element={<LazyPage><AdminMaintenancePage /></LazyPage>} />
                <Route path="settings" element={<Navigate to="/console/admin/site" replace />} />
                <Route path="site" element={<LazyPage><AdminSiteSettingsPage /></LazyPage>} />
                <Route path="access" element={<LazyPage><AdminAccessSettingsPage /></LazyPage>} />
                <Route path="seo" element={<Navigate to="/console/admin/site" replace />} />
                <Route path="compliance" element={<Navigate to="/console/admin/site" replace />} />
                <Route path="analytics" element={<LazyPage><AdminAnalyticsSettingsPage /></LazyPage>} />
                <Route path="appearance" element={<LazyPage><AdminAppearanceSettingsPage /></LazyPage>} />
              </Route>
              <Route index element={<Navigate to="/console/upload" replace />} />
            </Route>
          </Route>

          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </div>
      {config && <SiteFooter config={config} />}
    </div>
  )
}

export function AppRouter() {
  return <PublicRoutes />
}
