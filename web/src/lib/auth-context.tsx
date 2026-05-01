import { createContext, useContext, useEffect, useMemo, useState, useCallback } from 'react'
import type { RegisterResult, UserProfile } from './auth'
import * as authApi from './auth'

interface AuthContextValue {
  user: UserProfile | null
  isLoading: boolean
  isAuthenticated: boolean
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string, name: string) => Promise<RegisterResult>
  logout: () => Promise<void>
  updateProfile: (data: { name?: string; password?: string }) => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<UserProfile | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  useEffect(() => {
    if (!authApi.hasToken()) return
    let mounted = true
    const fetchProfile = async () => {
      setIsLoading(true)
      try {
        const profile = await authApi.getProfile()
        if (mounted) setUser(profile)
      } catch {
        authApi.clearTokens()
        if (mounted) setUser(null)
      } finally {
        if (mounted) setIsLoading(false)
      }
    }
    fetchProfile()
    return () => { mounted = false }
  }, [])

  const login = useCallback(async (email: string, password: string) => {
    const tokens = await authApi.login(email, password)
    authApi.saveTokens(tokens)
    const profile = await authApi.getProfile()
    setUser(profile)
  }, [])

  const register = useCallback(async (email: string, password: string, name: string) => {
    const result = await authApi.register(email, password, name)
    if (result.tokens) {
      authApi.saveTokens(result.tokens)
      const profile = await authApi.getProfile()
      setUser(profile)
    } else {
      authApi.clearTokens()
      setUser(null)
    }
    return result
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

// eslint-disable-next-line react-refresh/only-export-components
export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
