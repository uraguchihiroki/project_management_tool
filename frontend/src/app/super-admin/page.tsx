'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useRequireSuperAdmin, useSuperAdmin } from '@/context/SuperAdminContext'
import { superAdminCreateOrganization, superAdminGetOrganizations } from '@/lib/api'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Shield, Building2, Plus, LogOut, Loader2, Calendar } from 'lucide-react'
import { format } from 'date-fns'
import { ja } from 'date-fns/locale'
import { useAuthFetchEnabled } from '@/hooks/useAuthFetchEnabled'

export default function SuperAdminPage() {
  const currentSuperAdmin = useRequireSuperAdmin()
  const authFetch = useAuthFetchEnabled()
  const { logout } = useSuperAdmin()
  const queryClient = useQueryClient()
  const router = useRouter()
  const [newOrgName, setNewOrgName] = useState('')
  const [newOrgAdminEmail, setNewOrgAdminEmail] = useState('')
  const [newOrgAdminName, setNewOrgAdminName] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [formError, setFormError] = useState('')

  const { data: orgs = [], isLoading } = useQuery({
    queryKey: ['super-admin', 'organizations'],
    queryFn: superAdminGetOrganizations,
    enabled: authFetch && !!currentSuperAdmin,
  })

  const createMutation = useMutation({
    mutationFn: (data: { name: string; admin_email: string; admin_name: string }) =>
      superAdminCreateOrganization(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['super-admin', 'organizations'] })
      setNewOrgName('')
      setNewOrgAdminEmail('')
      setNewOrgAdminName('')
      setShowForm(false)
      setFormError('')
    },
    onError: () => {
      setFormError('作成に失敗しました（名前が重複している可能性があります）')
    },
  })

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault()
    if (!newOrgName.trim()) return
    setFormError('')
    createMutation.mutate({
      name: newOrgName.trim(),
      admin_email: newOrgAdminEmail.trim(),
      admin_name: newOrgAdminName.trim(),
    })
  }

  if (!currentSuperAdmin) return null

  return (
    <div className="min-h-screen bg-gray-900">
      {/* ヘッダー */}
      <header className="bg-gray-800 border-b border-gray-700 px-6 py-4">
        <div className="max-w-4xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Shield className="w-5 h-5 text-purple-400" />
            <span className="font-bold text-white">スーパー管理者パネル</span>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-sm text-gray-400">{currentSuperAdmin.name}</span>
            <button
              onClick={logout}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-400 hover:text-white hover:bg-gray-700 rounded-lg transition-colors"
            >
              <LogOut className="w-4 h-4" />
              ログアウト
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-4xl mx-auto p-6">
        {/* 組織一覧 */}
        <div className="bg-gray-800 rounded-xl border border-gray-700 overflow-hidden">
          <div className="flex items-center justify-between px-6 py-4 border-b border-gray-700">
            <div className="flex items-center gap-2">
              <Building2 className="w-5 h-5 text-purple-400" />
              <h2 className="font-semibold text-white">会社・組織管理</h2>
              <span className="text-xs bg-gray-700 text-gray-300 px-2 py-0.5 rounded-full">
                {orgs.length}件
              </span>
            </div>
            <button
              onClick={() => setShowForm(!showForm)}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-purple-600 hover:bg-purple-700 text-white rounded-lg transition-colors"
            >
              <Plus className="w-4 h-4" />
              新規会社作成
            </button>
          </div>

          {/* 新規作成フォーム */}
          {showForm && (
            <div className="px-6 py-4 border-b border-gray-700 bg-gray-750">
              <form onSubmit={handleCreate} className="space-y-3">
                <div className="flex gap-3 flex-wrap">
                  <input
                    type="text"
                    value={newOrgName}
                    onChange={(e) => setNewOrgName(e.target.value)}
                    placeholder="会社名・組織名 *"
                    className="flex-1 min-w-[200px] px-4 py-2 bg-gray-700 border border-gray-600 text-white rounded-lg placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500"
                    autoFocus
                  />
                  <input
                    type="email"
                    value={newOrgAdminEmail}
                    onChange={(e) => setNewOrgAdminEmail(e.target.value)}
                    placeholder="管理者メール"
                    className="flex-1 min-w-[200px] px-4 py-2 bg-gray-700 border border-gray-600 text-white rounded-lg placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500"
                  />
                  <input
                    type="text"
                    value={newOrgAdminName}
                    onChange={(e) => setNewOrgAdminName(e.target.value)}
                    placeholder="管理者名"
                    className="flex-1 min-w-[150px] px-4 py-2 bg-gray-700 border border-gray-600 text-white rounded-lg placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500"
                  />
                </div>
                <div className="flex gap-2">
                  <button
                    type="submit"
                    disabled={createMutation.isPending || !newOrgName.trim()}
                    className="flex items-center gap-2 px-4 py-2 bg-purple-600 hover:bg-purple-700 disabled:opacity-50 text-white rounded-lg"
                  >
                    {createMutation.isPending && <Loader2 className="w-4 h-4 animate-spin" />}
                    作成
                  </button>
                  <button
                    type="button"
                    onClick={() => { setShowForm(false); setFormError('') }}
                    className="px-4 py-2 text-gray-400 hover:text-white hover:bg-gray-700 rounded-lg"
                  >
                    キャンセル
                  </button>
                </div>
              </form>
              {formError && <p className="mt-2 text-sm text-red-400">{formError}</p>}
            </div>
          )}

          {isLoading ? (
            <div className="flex items-center justify-center py-16">
              <Loader2 className="w-6 h-6 animate-spin text-gray-400" />
            </div>
          ) : orgs.length === 0 ? (
            <div className="py-16 text-center text-gray-500">
              <Building2 className="w-10 h-10 mx-auto mb-3 text-gray-600" />
              <p>会社・組織が登録されていません</p>
            </div>
          ) : (
            <ul className="divide-y divide-gray-700">
              {orgs.map((org) => (
                <li key={org.id} className="px-6 py-4 flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="w-9 h-9 rounded-lg bg-purple-900/40 flex items-center justify-center">
                      <Building2 className="w-5 h-5 text-purple-400" />
                    </div>
                    <div>
                      <p className="font-medium text-white">{org.name}</p>
                      <p className="text-xs text-gray-500 flex items-center gap-1 mt-0.5">
                        <Calendar className="w-3 h-3" />
                        {format(new Date(org.created_at), 'yyyy/MM/dd', { locale: ja })}
                      </p>
                    </div>
                  </div>
                  <span className="text-xs text-gray-600 font-mono">{org.id.slice(0, 8)}...</span>
                </li>
              ))}
            </ul>
          )}
        </div>
      </main>
    </div>
  )
}
