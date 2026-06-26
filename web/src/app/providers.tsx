import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'

import '../i18n'
import { ThemeProvider } from '../lib/theme'
import { AuthProvider } from '../lib/auth-context'
import { SiteThemeRuntime } from '../components/site-theme-runtime'
import { Toaster } from '../components/ui/sonner'

const queryClient = new QueryClient()

export function AppProviders({ children }: { children: React.ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        <BrowserRouter>
          <AuthProvider>
            <SiteThemeRuntime />
            {children}
            <Toaster />
          </AuthProvider>
        </BrowserRouter>
      </ThemeProvider>
    </QueryClientProvider>
  )
}
