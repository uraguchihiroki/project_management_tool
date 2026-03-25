'use client'

import { createContext, useContext, useState, useEffect, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import type { SuperAdmin } from '@/types'
import { clearAuthSession, setAuthToken } from '@/lib/authToken'
import { resolveApiBaseURL } from '@/lib/api'

const SA_SESSION_KEY = 'currentSuperAdmin'

interface SuperAdminContextType {
  currentSuperAdmin: SuperAdmin | null
  login: (email: string) => Promise<{ ok: boolean; error?: string }>
  logout: () => void
}

const SuperAdminContext = createContext<SuperAdminContextType | null>(null)

export function SuperAdminProvider({ children }: { children: React.ReactNode }) {
  const [currentSuperAdmin, setCurrentSuperAdmin] = useState<SuperAdmin | null>(null)
  const router = useRouter()

  useEffect(() => {
    try {
      const stored = sessionStorage.getItem(SA_SESSION_KEY)
      if (stored) setCurrentSuperAdmin(JSON.parse(stored))
    } catch {
      // SSR環境では無視
    }
  }, [])

  const login = useCallback(async (email: string): Promise<{ ok: boolean; error?: string }> => {
    try {
      const res = await fetch(`${resolveApiBaseURL()}/super-admin/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email }),
      })
      if (!res.ok) return { ok: false, error: 'メールアドレスが見つかりません' }
      const json = await res.json()
      const admin: SuperAdmin = json.data
      const token: string | undefined = json.token
      if (!admin || !token) return { ok: false, error: 'ログインに失敗しました' }
      setAuthToken(token)
      sessionStorage.setItem(SA_SESSION_KEY, JSON.stringify(admin))
      setCurrentSuperAdmin(admin)
      return { ok: true }
    } catch {
      return { ok: false, error: 'ログインに失敗しました' }
    }
  }, [])

  const logout = useCallback(() => {
    clearAuthSession()
    setCurrentSuperAdmin(null)
    router.push('/super-admin/login')
  }, [router])

  return (
    <SuperAdminContext.Provider value={{ currentSuperAdmin, login, logout }}>
      {children}
    </SuperAdminContext.Provider>
  )
}

export function useSuperAdmin(): SuperAdminContextType {
  const ctx = useContext(SuperAdminContext)
  if (!ctx) throw new Error('useSuperAdmin must be used within SuperAdminProvider')
  return ctx
}

export function useRequireSuperAdmin(): SuperAdmin {
  const { currentSuperAdmin } = useSuperAdmin()
  const router = useRouter()

  useEffect(() => {
    const stored = sessionStorage.getItem(SA_SESSION_KEY)
    if (!stored) router.push('/super-admin/login')
  }, [currentSuperAdmin, router])

  return currentSuperAdmin as SuperAdmin
}
