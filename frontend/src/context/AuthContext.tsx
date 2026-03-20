'use client'

import { createContext, useContext, useState, useEffect, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import type { User, Organization } from '@/types'
import { setAuthToken } from '@/lib/authToken'

const SESSION_KEY = 'currentUser'
const ORG_KEY = 'currentOrg'

interface AuthContextType {
  currentUser: User | null
  currentOrg: Organization | null
  login: (email: string, asAdmin?: boolean) => Promise<{ ok: boolean; error?: string }>
  logout: () => void
  register: (name: string, email: string, asAdmin?: boolean) => Promise<{ ok: boolean; error?: string }>
  selectOrg: (org: Organization) => void
}

const AuthContext = createContext<AuthContextType | null>(null)

function getInitialUser(): User | null {
  if (typeof window === 'undefined') return null
  try {
    const stored = sessionStorage.getItem(SESSION_KEY)
    return stored ? JSON.parse(stored) : null
  } catch {
    return null
  }
}

function getInitialOrg(): Organization | null {
  if (typeof window === 'undefined') return null
  try {
    const stored = sessionStorage.getItem(ORG_KEY)
    return stored ? JSON.parse(stored) : null
  } catch {
    return null
  }
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [currentUser, setCurrentUser] = useState<User | null>(getInitialUser)
  const [currentOrg, setCurrentOrg] = useState<Organization | null>(getInitialOrg)
  const router = useRouter()

  useEffect(() => {
    try {
      const stored = sessionStorage.getItem(SESSION_KEY)
      if (stored) setCurrentUser(JSON.parse(stored))
      const storedOrg = sessionStorage.getItem(ORG_KEY)
      if (storedOrg) setCurrentOrg(JSON.parse(storedOrg))
    } catch {
      // SSR環境では無視
    }
  }, [])

  const selectOrg = useCallback((org: Organization) => {
    sessionStorage.setItem(ORG_KEY, JSON.stringify(org))
    setCurrentOrg(org)
  }, [])

  // ログイン後に組織を取得し、1件なら自動選択・複数なら選択画面へ・0件ならエラー
  const handleOrgSelection = useCallback(async (userId: string): Promise<{ dest: string; error?: string }> => {
    try {
      const base = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'
      const res = await fetch(`${base}/users/${userId}/organizations`)
      if (!res.ok) return { dest: '/projects' }
      const json = await res.json()
      const orgs: Organization[] = json.data ?? []
      if (orgs.length === 1) {
        sessionStorage.setItem(ORG_KEY, JSON.stringify(orgs[0]))
        setCurrentOrg(orgs[0])
        return { dest: '/projects' }
      }
      if (orgs.length > 1) {
        return { dest: '/select-org' }
      }
      return { dest: '/login', error: '所属組織がありません。管理者に連絡してください。' }
    } catch {
      return { dest: '/projects' }
    }
  }, [])

  const login = useCallback(async (email: string, asAdmin?: boolean): Promise<{ ok: boolean; error?: string }> => {
    try {
      const base = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'
      const res = await fetch(`${base}/admin/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email }),
      })
      if (!res.ok) throw new Error('APIエラー')
      const json = await res.json()
      const payload = json.data ?? {}
      const found = payload.user as User | undefined
      const token = payload.token as string | undefined
      if (!found || !token) return { ok: false, error: 'メールアドレスが見つかりません' }
      const user = { ...found, is_admin: asAdmin ?? found.is_admin }
      setAuthToken(token)
      sessionStorage.setItem(SESSION_KEY, JSON.stringify(user))
      setCurrentUser(user)
      const { dest, error } = await handleOrgSelection(found.id)
      if (error) return { ok: false, error }
      router.push(dest)
      return { ok: true }
    } catch {
      return { ok: false, error: 'ログインに失敗しました' }
    }
  }, [handleOrgSelection, router])

  const logout = useCallback(() => {
    setAuthToken(null)
    sessionStorage.removeItem(SESSION_KEY)
    sessionStorage.removeItem(ORG_KEY)
    setCurrentUser(null)
    setCurrentOrg(null)
    router.push('/login')
  }, [router])

  const register = useCallback(async (name: string, email: string, asAdmin?: boolean): Promise<{ ok: boolean; error?: string }> => {
    try {
      const base = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'
      const res = await fetch(`${base}/users`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, email }),
      })
      if (!res.ok) {
        const json = await res.json().catch(() => ({}))
        return { ok: false, error: json.message ?? 'ユーザー登録に失敗しました' }
      }
      const json = await res.json()
      const created: User = json.data
      if (asAdmin !== undefined) {
        await fetch(`${base}/users/${created.id}/admin`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ is_admin: asAdmin }),
        })
      }
      // 新規登録後はJWT発行を伴うログインフローへ委譲
      return await login(email, asAdmin)
    } catch {
      return { ok: false, error: '登録に失敗しました' }
    }
  }, [login])

  return (
    <AuthContext.Provider value={{ currentUser, currentOrg, login, logout, register, selectOrg }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthContextType {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}

// 未ログイン時に /login へリダイレクトするフック
export function useRequireAuth(): User {
  const { currentUser } = useAuth()
  const router = useRouter()

  useEffect(() => {
    if (currentUser === null) {
      const stored = sessionStorage.getItem(SESSION_KEY)
      if (!stored) router.push('/login')
    }
  }, [currentUser, router])

  return currentUser as User
}

// 管理者でない場合は /projects へリダイレクトするフック
export function useRequireAdmin(): User {
  const { currentUser } = useAuth()
  const router = useRouter()

  useEffect(() => {
    const stored = sessionStorage.getItem(SESSION_KEY)
    if (!stored) {
      router.push('/login')
      return
    }
    const user: User = JSON.parse(stored)
    if (!user.is_admin) {
      router.push('/projects')
    }
  }, [currentUser, router])

  return currentUser as User
}
