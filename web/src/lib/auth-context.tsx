import { createContext, useContext, useEffect, useMemo, useState, useCallback } from 'react'
import type { UserProfile } from './auth'
import * as authApi from './auth'

interface AuthContextValue {
  user: UserProfile | null
  isLoading: boolean
  isAuthenticated: boolean
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string, name: string) => Promise<void>
  logout: () => Promise<void>
  updateProfile: (data: { name?: string; password?: string }) => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<UserProfile | null>(null)
  const [isLoading, setIsLoading] = useState(() => authApi.hasToken())

  useEffect(() => {
    if (!authApi.hasToken()) return
    setIsLoading(true)
    authApi
      .getProfile()
      .then(setUser)
      .catch(() => {
        authApi.clearTokens()
        setUser(null)
      })
      .finally(() => setIsLoading(false))
  }, [])

  const login = useCallback(async (email: string, password: string) => {
    const tokens = await authApi.login(email, password)
    authApi.saveTokens(tokens)
    const profile = await authApi.getProfile()
    setUser(profile)
  }, [])

  const register = useCallback(async (email: string, password: string, name: string) => {
    const tokens = await authApi.register(email, password, name)
    authApi.saveTokens(tokens)
    const profile = await authApi.getProfile()
    setUser(profile)
  }, [])

  const logout = useCallback(async () => {
    await authApi.logout()
    authApi.clearTokens()
    setUser(null)
  }, [])

  const updateProfile = useCallback(async (data: { name?: string; password?: string }) => {
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

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
