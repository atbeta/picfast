import { Navigate, Outlet, Route, Routes } from 'react-router-dom'

import { ACCESS_TOKEN_KEY } from '../lib/api'
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
  const token = localStorage.getItem(ACCESS_TOKEN_KEY)
  if (!token) {
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
