import { Navigate, Outlet, Route, Routes } from 'react-router-dom'

import { useAuth } from '../lib/auth-context'
import { ConsoleLayout } from '../pages/layouts/console-layout'
import { PublicLayout } from '../pages/layouts/public-layout'
import { AlbumsPage } from '../pages/console/albums-page'
import { ApiTokensPage } from '../pages/console/api-tokens-page'
import { ImagesPage } from '../pages/console/images-page'
import { SettingsPage } from '../pages/console/settings-page'
import { UploadPage } from '../pages/console/upload-page'
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

export function AppRouter() {
  return (
    <Routes>
      <Route element={<PublicLayout />}>
        <Route path="/" element={<GuestUploadPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
      </Route>

      <Route element={<RequireAuth />}>
        <Route path="/console" element={<ConsoleLayout />}>
          <Route path="upload" element={<UploadPage />} />
          <Route path="images" element={<ImagesPage />} />
          <Route path="albums" element={<AlbumsPage />} />
          <Route path="api-tokens" element={<ApiTokensPage />} />
          <Route path="settings" element={<SettingsPage />} />
          <Route index element={<Navigate to="/console/upload" replace />} />
        </Route>
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
