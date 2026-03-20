'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/context/AuthContext'
import { FolderKanban } from 'lucide-react'

export default function LoginPage() {
  const { currentUser, login } = useAuth()
  const router = useRouter()
  const [email, setEmail] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [asAdmin, setAsAdmin] = useState(false)

  useEffect(() => {
    if (currentUser) router.push('/projects')
  }, [currentUser, router])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    const result = await login(email, asAdmin)

    setLoading(false)

    if (result.ok) {
      router.push('/projects')
    } else {
      setError(result.error ?? 'エラーが発生しました')
    }
  }

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="w-full max-w-sm">
        <div className="flex items-center justify-center gap-2 mb-8">
          <FolderKanban className="w-8 h-8 text-blue-600" />
          <span className="text-2xl font-bold text-gray-900">ProjectHub</span>
        </div>

        <div className="bg-white rounded-xl border border-gray-200 p-6 shadow-sm">
          <h1 className="text-lg font-semibold text-gray-900 mb-6">ログイン</h1>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                メールアドレス <span className="text-red-500">*</span>
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="例: taro@example.com"
                required
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <label className="flex items-center gap-2 cursor-pointer select-none">
              <input
                type="checkbox"
                checked={asAdmin}
                onChange={(e) => setAsAdmin(e.target.checked)}
                className="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500 cursor-pointer"
              />
              <span className="text-sm text-gray-600">管理者としてログイン</span>
            </label>

            {error && (
              <p data-testid="login-error" className="text-sm text-red-500 bg-red-50 px-3 py-2 rounded-lg">
                {error}
              </p>
            )}

            <button
              type="submit"
              disabled={loading}
              className="w-full py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50 transition-colors"
            >
              {loading ? '処理中...' : 'ログイン'}
            </button>
          </form>

          <p className="mt-4 text-center text-xs text-gray-400">
            アカウントがない場合は組織の管理者に連絡してください
          </p>
        </div>
      </div>
    </div>
  )
}
