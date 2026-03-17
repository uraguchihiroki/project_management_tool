'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Pencil, Trash2, X, Check } from 'lucide-react'
import type { Role } from '@/types'

const API = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'

async function fetchRoles(): Promise<Role[]> {
  const res = await fetch(`${API}/roles`)
  const json = await res.json()
  return json.data ?? []
}

export default function RolesPage() {
  const queryClient = useQueryClient()
  const { data: roles = [], isLoading } = useQuery({ queryKey: ['roles'], queryFn: fetchRoles })

  const [showForm, setShowForm] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [form, setForm] = useState({ name: '', level: 1, description: '' })
  const [error, setError] = useState('')

  const createMutation = useMutation({
    mutationFn: async (data: typeof form) => {
      const res = await fetch(`${API}/roles`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      })
      if (!res.ok) {
        const json = await res.json()
        throw new Error(json.message ?? '作成に失敗しました')
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] })
      setShowForm(false)
      setForm({ name: '', level: 1, description: '' })
      setError('')
    },
    onError: (e: Error) => setError(e.message),
  })

  const updateMutation = useMutation({
    mutationFn: async ({ id, data }: { id: number; data: typeof form }) => {
      const res = await fetch(`${API}/roles/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      })
      if (!res.ok) {
        const json = await res.json()
        throw new Error(json.message ?? '更新に失敗しました')
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] })
      setEditingId(null)
      setForm({ name: '', level: 1, description: '' })
      setError('')
    },
    onError: (e: Error) => setError(e.message),
  })

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => {
      await fetch(`${API}/roles/${id}`, { method: 'DELETE' })
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['roles'] }),
  })

  const startEdit = (role: Role) => {
    setEditingId(role.id)
    setForm({ name: role.name, level: role.level, description: role.description })
    setShowForm(false)
    setError('')
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!form.name.trim()) { setError('名前は必須です'); return }
    if (editingId !== null) {
      updateMutation.mutate({ id: editingId, data: form })
    } else {
      createMutation.mutate(form)
    }
  }

  return (
    <div className="max-w-3xl">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-bold text-gray-900">役職管理</h1>
          <p className="text-sm text-gray-500 mt-0.5">役職とヒエラルキーレベルを管理します</p>
        </div>
        {!showForm && editingId === null && (
          <button
            onClick={() => { setShowForm(true); setForm({ name: '', level: 1, description: '' }); setError('') }}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors"
          >
            <Plus className="w-4 h-4" />
            役職を追加
          </button>
        )}
      </div>

      {/* 追加フォーム */}
      {(showForm || editingId !== null) && (
        <div className="bg-white border border-gray-200 rounded-xl p-5 mb-6 shadow-sm">
          <h2 className="text-sm font-semibold text-gray-700 mb-4">
            {editingId !== null ? '役職を編集' : '新しい役職'}
          </h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">役職名 *</label>
                <input
                  type="text"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  placeholder="例: 部長"
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">
                  レベル *
                  <span className="ml-1 font-normal text-gray-400">（大きいほど上位）</span>
                </label>
                <input
                  type="number"
                  min={1}
                  max={99}
                  value={form.level}
                  onChange={(e) => setForm({ ...form, level: parseInt(e.target.value) || 1 })}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">説明</label>
              <input
                type="text"
                value={form.description}
                onChange={(e) => setForm({ ...form, description: e.target.value })}
                placeholder="例: 部門の責任者"
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            {error && <p className="text-sm text-red-500">{error}</p>}
            <div className="flex gap-2">
              <button
                type="submit"
                disabled={createMutation.isPending || updateMutation.isPending}
                className="flex items-center gap-1.5 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50 transition-colors"
              >
                <Check className="w-4 h-4" />
                {editingId !== null ? '更新' : '追加'}
              </button>
              <button
                type="button"
                onClick={() => { setShowForm(false); setEditingId(null); setError('') }}
                className="flex items-center gap-1.5 px-4 py-2 border border-gray-300 text-gray-600 rounded-lg text-sm hover:bg-gray-50 transition-colors"
              >
                <X className="w-4 h-4" />
                キャンセル
              </button>
            </div>
          </form>
        </div>
      )}

      {/* 役職一覧 */}
      <div className="bg-white border border-gray-200 rounded-xl shadow-sm overflow-hidden">
        {isLoading ? (
          <div className="p-8 text-center text-gray-400 text-sm">読み込み中...</div>
        ) : roles.length === 0 ? (
          <div className="p-8 text-center text-gray-400 text-sm">
            役職がまだありません。「役職を追加」から作成してください。
          </div>
        ) : (
          <table className="w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase tracking-wide">役職名</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase tracking-wide w-24">レベル</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase tracking-wide">説明</th>
                <th className="px-4 py-3 w-20"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {roles.map((role) => (
                <tr key={role.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-4 py-3">
                    <span className="font-medium text-gray-900 text-sm">{role.name}</span>
                  </td>
                  <td className="px-4 py-3">
                    <span className="inline-flex items-center justify-center w-8 h-8 rounded-full bg-blue-100 text-blue-700 text-xs font-bold">
                      {role.level}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-500">{role.description || '—'}</td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-1 justify-end">
                      <button
                        onClick={() => startEdit(role)}
                        className="p-1.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                        title="編集"
                      >
                        <Pencil className="w-3.5 h-3.5" />
                      </button>
                      <button
                        onClick={() => {
                          if (confirm(`「${role.name}」を削除しますか？`)) {
                            deleteMutation.mutate(role.id)
                          }
                        }}
                        className="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                        title="削除"
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}
