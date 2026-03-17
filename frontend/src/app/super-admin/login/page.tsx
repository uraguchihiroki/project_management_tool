'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useSuperAdmin } from '@/context/SuperAdminContext'
import { Shield, Mail, Loader2 } from 'lucide-react'

export default function SuperAdminLoginPage() {
  const [email, setEmail] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const { login } = useSuperAdmin()
  const router = useRouter()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    const result = await login(email)
    setLoading(false)
    if (result.ok) {
      router.push('/super-admin')
    } else {
      setError(result.error ?? 'ログインに失敗しました')
    }
  }

  return (
    <div className="min-h-screen bg-gray-900 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="flex items-center justify-center gap-2 mb-8">
          <Shield className="w-8 h-8 text-purple-400" />
          <span className="text-2xl font-bold text-white">スーパー管理者</span>
        </div>

        <div className="bg-gray-800 rounded-xl border border-gray-700 p-8 shadow-2xl">
          <h1 className="text-lg font-semibold text-white mb-6">管理者ログイン</h1>

          {error && (
            <div className="mb-4 p-3 bg-red-900/40 border border-red-700 text-red-300 rounded-lg text-sm">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-1">
                メールアドレス
              </label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="superadmin@example.com"
                  required
                  className="w-full pl-10 pr-4 py-2.5 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full flex items-center justify-center gap-2 py-2.5 px-4 bg-purple-600 hover:bg-purple-700 disabled:opacity-50 text-white font-medium rounded-lg transition-colors"
            >
              {loading && <Loader2 className="w-4 h-4 animate-spin" />}
              ログイン
            </button>
          </form>
        </div>

        <p className="text-center text-gray-500 text-xs mt-6">
          通常ユーザーのログインは{' '}
          <a href="/login" className="text-gray-400 hover:text-white underline">こちら</a>
        </p>
      </div>
    </div>
  )
}
