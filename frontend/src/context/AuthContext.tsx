'use client'

import { createContext, useContext, useState, useEffect, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import type { User } from '@/types'

const SESSION_KEY = 'currentUser'

interface AuthContextType {
  currentUser: User | null
  login: (email: string, asAdmin?: boolean) => Promise<{ ok: boolean; error?: string }>
  logout: () => void
  register: (name: string, email: string, asAdmin?: boolean) => Promise<{ ok: boolean; error?: string }>
}

const AuthContext = createContext<AuthContextType | null>(null)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [currentUser, setCurrentUser] = useState<User | null>(null)
  const router = useRouter()

  // 初期化時にsessionStorageから復元
  useEffect(() => {
    try {
      const stored = sessionStorage.getItem(SESSION_KEY)
      if (stored) setCurrentUser(JSON.parse(stored))
    } catch {
      // sessionStorageが使えない環境（SSR）では無視
    }
  }, [])

  const login = useCallback(async (email: string, asAdmin?: boolean): Promise<{ ok: boolean; error?: string }> => {
    try {
      const base = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'
      const res = await fetch(`${base}/users`)
      if (!res.ok) throw new Error('APIエラー')
      const json = await res.json()
      const users: User[] = json.data ?? []
      const found = users.find((u) => u.email === email)
      if (!found) return { ok: false, error: 'メールアドレスが見つかりません' }
      // DBのis_adminも更新する
      await fetch(`${base}/users/${found.id}/admin`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ is_admin: asAdmin ?? false }),
      })
      const user = { ...found, is_admin: asAdmin ?? false }
      sessionStorage.setItem(SESSION_KEY, JSON.stringify(user))
      setCurrentUser(user)
      return { ok: true }
    } catch {
      return { ok: false, error: 'ログインに失敗しました' }
    }
  }, [])

  const logout = useCallback(() => {
    sessionStorage.removeItem(SESSION_KEY)
    setCurrentUser(null)
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
      // チェックボックスの状態でDBのis_adminも更新する
      if (asAdmin !== undefined) {
        await fetch(`${base}/users/${created.id}/admin`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ is_admin: asAdmin }),
        })
      }
      const user = { ...created, is_admin: asAdmin ?? created.is_admin }
      sessionStorage.setItem(SESSION_KEY, JSON.stringify(user))
      setCurrentUser(user)
      return { ok: true }
    } catch {
      return { ok: false, error: '登録に失敗しました' }
    }
  }, [])

  return (
    <AuthContext.Provider value={{ currentUser, login, logout, register }}>
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
      // sessionStorage読み込み完了後にnullなら未ログイン
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
