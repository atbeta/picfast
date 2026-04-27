import { useEffect, useState } from 'react'
import { Navigate, Outlet, Route, Routes } from 'react-router-dom'

import { useAuth } from '../lib/auth-context'
import { getSiteConfig, type SiteConfig } from '../lib/site-config'
import { ConsoleLayout } from '../pages/layouts/console-layout'
import { PublicLayout } from '../pages/layouts/public-layout'
import { AlbumsPage } from '../pages/console/albums-page'
import { ApiTokensPage } from '../pages/console/api-tokens-page'
import { ImagesPage } from '../pages/console/images-page'
import { SettingsPage } from '../pages/console/settings-page'
import { UploadPage } from '../pages/console/upload-page'
import { AdminUsersPage } from '../pages/console/admin/users-page'
import { AdminGroupsPage } from '../pages/console/admin/groups-page'
import { AdminStrategiesPage } from '../pages/console/admin/strategies-page'
import { AdminImagesPage } from '../pages/console/admin/images-page'
import { AdminSettingsPage } from '../pages/console/admin/settings-page'
import { AdminDashboardPage } from '../pages/console/admin/dashboard-page'
import { GuestUploadPage } from '../pages/public/guest-upload-page'
import { LoginPage } from '../pages/public/login-page'
import { RegisterPage } from '../pages/public/register-page'

function RequireAuth() {
  const { isAuthenticated, isLoading } = useAuth()
  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-400 border-t-transparent" />
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
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-400 border-t-transparent" />
      </div>
    )
  }
  if (user?.role !== 'admin') {
    return <Navigate to="/console" replace />
  }
  return <Outlet />
}

function PublicRoutes() {
  const [config, setConfig] = useState<SiteConfig | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    getSiteConfig()
      .then(setConfig)
      .catch(() => setConfig(null))
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-400 border-t-transparent" />
      </div>
    )
  }

  return (
    <Routes>
      <Route element={<PublicLayout />}>
        <Route
          path="/"
          element={
            config?.allow_guest_upload ? (
              <GuestUploadPage />
            ) : (
              <Navigate to="/login" replace />
            )
          }
        />
        <Route path="/login" element={<LoginPage />} />
        <Route
          path="/register"
          element={
            config?.allow_registration ? (
              <RegisterPage />
            ) : (
              <Navigate to="/login" replace />
            )
          }
        />
      </Route>

      <Route element={<RequireAuth />}>
        <Route path="/console" element={<ConsoleLayout />}>
          <Route path="upload" element={<UploadPage />} />
          <Route path="images" element={<ImagesPage />} />
          <Route path="albums" element={<AlbumsPage />} />
          <Route path="api-tokens" element={<ApiTokensPage />} />
          <Route path="settings" element={<SettingsPage />} />
          <Route path="admin" element={<RequireAdmin />}>
            <Route index element={<AdminDashboardPage />} />
            <Route path="users" element={<AdminUsersPage />} />
            <Route path="groups" element={<AdminGroupsPage />} />
            <Route path="strategies" element={<AdminStrategiesPage />} />
            <Route path="images" element={<AdminImagesPage />} />
            <Route path="settings" element={<AdminSettingsPage />} />
            <Route index element={<Navigate to="/console/admin/users" replace />} />
          </Route>
          <Route index element={<Navigate to="/console/upload" replace />} />
        </Route>
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export function AppRouter() {
  return <PublicRoutes />
}
