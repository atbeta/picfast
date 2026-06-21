import { createContext, useContext, useEffect, useMemo, useState, useCallback } from 'react'
import type { RegisterResult, UserProfile } from './auth'
import * as authApi from './auth'
import { logError } from './error-handler'

interface AuthContextValue {
  user: UserProfile | null
  isLoading: boolean
  isAuthenticated: boolean
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string, name: string, language?: string) => Promise<RegisterResult>
  logout: () => Promise<void>
  updateProfile: (data: { name?: string; password?: string; settings?: Record<string, unknown> }) => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<UserProfile | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  
  useEffect(() => {
    let mounted = true
    const fetchProfile = async () => {
      if (!authApi.hasToken()) {
        // No localStorage token — still attempt profile fetch via cookie
        // (OAuth stores the access token in an HttpOnly cookie).
        try {
          const res = await fetch('/api/v1/users/me', { credentials: 'include' })
          if (res.ok) {
            const body = await res.json()
            if (body?.data && mounted) {
              setUser(body.data)
              setIsLoading(false)
              return
            }
          }
        } catch {
          // user is not authenticated — fall through
        }
        if (mounted) setIsLoading(false)
        return
      }
      setIsLoading(true)
      try {
        const profile = await authApi.getProfile()
        if (mounted) setUser(profile)
      } catch (err: unknown) {
        logError('auth.fetchProfile', err)
        if (authApi.hasToken()) {
          authApi.clearTokens()
        }
        if (mounted) setUser(null)
      } finally {
        if (mounted) setIsLoading(false)
      }
    }
    fetchProfile()
    return () => { mounted = false }
  }, [])

  const login = useCallback(async (email: string, password: string) => {
    setIsLoading(true)
    try {
      const tokens = await authApi.login(email, password)
      authApi.saveTokens(tokens)
      const profile = await authApi.getProfile()
      setUser(profile)
    } finally {
      setIsLoading(false)
    }
  }, [])

  const register = useCallback(async (email: string, password: string, name: string, language?: string) => {
    setIsLoading(true)
    try {
      const result = await authApi.register(email, password, name, language)
      if (result.tokens) {
        authApi.saveTokens(result.tokens)
        const profile = await authApi.getProfile()
        setUser(profile)
      } else {
        authApi.clearTokens()
        setUser(null)
      }
      return result
    } finally {
      setIsLoading(false)
    }
  }, [])
  const logout = useCallback(async () => {
    await authApi.logout()
    authApi.clearTokens()
    setUser(null)
  }, [])

  const updateProfile = useCallback(async (data: { name?: string; password?: string; settings?: Record<string, unknown> }) => {
    const profile = await authApi.updateProfile(data)
    setUser(profile)
  }, [])

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      isLoading,
      isAuthenticated: !!user,
      login,
      register,
      logout,
      updateProfile,
    }),
    [user, isLoading, login, register, logout, updateProfile],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

// eslint-disable-next-line react-refresh/only-export-components
export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
