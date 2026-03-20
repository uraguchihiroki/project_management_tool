'use client'

import { createContext, useContext, useState, useEffect, useLayoutEffect, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import type { User, Organization } from '@/types'
import { clearAuthSession, setAuthToken } from '@/lib/authToken'
import { adminLogin, createUser, getUserOrganizations, setUserAdmin, switchOrganization } from '@/lib/api'

const SESSION_KEY = 'currentUser'
const ORG_KEY = 'currentOrg'

interface AuthContextType {
  currentUser: User | null
  currentOrg: Organization | null
  login: (email: string, asAdmin?: boolean) => Promise<{ ok: boolean; error?: string }>
  logout: () => void
  register: (name: string, email: string, asAdmin?: boolean) => Promise<{ ok: boolean; error?: string }>
  selectOrg: (org: Organization) => Promise<void>
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

  // ペイント前に sessionStorage を同期（E2E の addInitScript や Hydration 直後の currentOrg 欠落を防ぐ）
  useLayoutEffect(() => {
    try {
      const stored = sessionStorage.getItem(SESSION_KEY)
      if (stored) setCurrentUser(JSON.parse(stored))
      const storedOrg = sessionStorage.getItem(ORG_KEY)
      if (storedOrg) setCurrentOrg(JSON.parse(storedOrg))
      const t = sessionStorage.getItem('authToken')
      if (t) setAuthToken(t)
    } catch {
      // ignore
    }
  }, [])

  const selectOrg = useCallback(async (org: Organization) => {
    try {
      const payload = await switchOrganization(org.id)
      setAuthToken(payload.token)
      sessionStorage.setItem(SESSION_KEY, JSON.stringify(payload.user))
      sessionStorage.setItem(ORG_KEY, JSON.stringify(org))
      setCurrentUser(payload.user)
      setCurrentOrg(org)
    } catch {
      sessionStorage.setItem(ORG_KEY, JSON.stringify(org))
      setCurrentOrg(org)
    }
  }, [])

  // ログイン後に組織を取得し、1件なら自動選択・複数なら選択画面へ・0件ならエラー
  const handleOrgSelection = useCallback(
    async (userId: string): Promise<{ dest: string; error?: string; syncedUser?: User }> => {
      try {
        // JWT 必須 API のため axios（Authorization 付与）を使う。生の fetch だとトークンが付かず 401 になる。
        const orgs: Organization[] = await getUserOrganizations(userId)
        if (orgs.length === 1) {
          sessionStorage.setItem(ORG_KEY, JSON.stringify(orgs[0]))
          setCurrentOrg(orgs[0])
          try {
            const payload = await switchOrganization(orgs[0].id)
            setAuthToken(payload.token)
            return { dest: '/projects', syncedUser: payload.user }
          } catch {
            return { dest: '/projects' }
          }
        }
        if (orgs.length > 1) {
          return { dest: '/select-org' }
        }
        return { dest: '/login', error: '所属組織がありません。管理者に連絡してください。' }
      } catch {
        return {
          dest: '/login',
          error: '所属組織の取得に失敗しました。バックエンドが起動しているか確認してください。',
        }
      }
    },
    []
  )

  const login = useCallback(async (email: string, asAdmin?: boolean): Promise<{ ok: boolean; error?: string }> => {
    try {
      const payload = await adminLogin(email)
      const found = payload.user
      const token = payload.token
      if (!found || !token) return { ok: false, error: 'メールアドレスが見つかりません' }
      const user = { ...found, is_admin: asAdmin ?? found.is_admin }
      // JWT のみ先にセット（getUserOrganizations に必要）。currentUser は組織が決まってから。
      // 先に setCurrentUser すると /login の useEffect が /projects へ飛ばし、Turbopack が先に /projects をコンパイルして競合する。
      setAuthToken(token)
      const { dest, error, syncedUser } = await handleOrgSelection(found.id)
      if (error) {
        setAuthToken(null)
        return { ok: false, error }
      }
      const userForSession = syncedUser
        ? { ...syncedUser, is_admin: asAdmin ?? syncedUser.is_admin }
        : user
      sessionStorage.setItem(SESSION_KEY, JSON.stringify(userForSession))
      setCurrentUser(userForSession)
      router.push(dest)
      return { ok: true }
    } catch (e: unknown) {
      const msg =
        typeof e === 'object' && e !== null && 'response' in e
          ? (e as { response?: { data?: { message?: string } } }).response?.data?.message
          : undefined
      if (msg) return { ok: false, error: String(msg) }
      return { ok: false, error: 'ログインに失敗しました（メールが未登録、またはAPIに接続できません）' }
    }
  }, [handleOrgSelection, router])

  const logout = useCallback(() => {
    clearAuthSession()
    setCurrentUser(null)
    setCurrentOrg(null)
    router.push('/login')
  }, [router])

  const register = useCallback(async (name: string, email: string, asAdmin?: boolean): Promise<{ ok: boolean; error?: string }> => {
    try {
      await createUser({ name, email })
      const payload = await adminLogin(email)
      if (!payload.token || !payload.user) {
        return { ok: false, error: '登録後のログインに失敗しました' }
      }
      setAuthToken(payload.token)
      if (asAdmin !== undefined) {
        await setUserAdmin(payload.user.id, asAdmin)
      }
      const user = { ...payload.user, is_admin: asAdmin ?? payload.user.is_admin }
      const { dest, error, syncedUser } = await handleOrgSelection(payload.user.id)
      if (error) {
        setAuthToken(null)
        return { ok: false, error }
      }
      const userForSession = syncedUser ? { ...syncedUser, is_admin: asAdmin ?? syncedUser.is_admin } : user
      sessionStorage.setItem(SESSION_KEY, JSON.stringify(userForSession))
      setCurrentUser(userForSession)
      router.push(dest)
      return { ok: true }
    } catch (e: unknown) {
      const msg =
        typeof e === 'object' && e !== null && 'response' in e
          ? (e as { response?: { data?: { message?: string } } }).response?.data?.message
          : undefined
      return { ok: false, error: msg ? String(msg) : '登録に失敗しました' }
    }
  }, [handleOrgSelection, router])

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
