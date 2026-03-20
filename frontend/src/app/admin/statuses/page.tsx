'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Pencil, Trash2, X, Check } from 'lucide-react'
import type { Status } from '@/types'
import { useAuth } from '@/context/AuthContext'
import { useAuthFetchEnabled } from '@/hooks/useAuthFetchEnabled'

const API = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'

async function fetchOrgStatuses(orgId: string): Promise<Status[]> {
  const res = await fetch(`${API}/organizations/${orgId}/statuses?exclude_system=1`)
  const json = await res.json()
  const data: Status[] = json.data ?? []
  return data.filter((s) => !s.project_id && s.organization_id)
}

export default function StatusesPage() {
  const { currentOrg } = useAuth()
  const authFetch = useAuthFetchEnabled()
  const queryClient = useQueryClient()
  const { data: statuses = [], isLoading } = useQuery({
    queryKey: ['org-statuses-admin', currentOrg?.id],
    queryFn: () => fetchOrgStatuses(currentOrg?.id ?? ''),
    enabled: authFetch && !!currentOrg?.id,
  })

  const [showForm, setShowForm] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [form, setForm] = useState({ name: '', color: '#6B7280', type: 'issue' as 'issue' | 'project', order: 1 })
  const [error, setError] = useState('')

  const createMutation = useMutation({
    mutationFn: async (data: typeof form) => {
      const res = await fetch(`${API}/organizations/${currentOrg!.id}/statuses`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: data.name,
          color: data.color,
          type: data.type,
          order: data.order,
        }),
      })
      if (!res.ok) {
        const json = await res.json()
        throw new Error(json.message ?? '作成に失敗しました')
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['org-statuses-admin'] })
      queryClient.invalidateQueries({ queryKey: ['org-statuses'] })
      setShowForm(false)
      setForm({ name: '', color: '#6B7280', type: 'issue', order: 1 })
      setError('')
    },
    onError: (e: Error) => setError(e.message),
  })

  const updateMutation = useMutation({
    mutationFn: async ({ id, data }: { id: string; data: { name: string; color: string; order: number } }) => {
      const res = await fetch(`${API}/statuses/${id}`, {
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
      queryClient.invalidateQueries({ queryKey: ['org-statuses-admin'] })
      queryClient.invalidateQueries({ queryKey: ['org-statuses'] })
      setEditingId(null)
      setForm({ name: '', color: '#6B7280', type: 'issue', order: 1 })
      setError('')
    },
    onError: (e: Error) => setError(e.message),
  })

  const deleteMutation = useMutation({
    mutationFn: async (id: string) => {
      const res = await fetch(`${API}/statuses/${id}`, { method: 'DELETE' })
      if (!res.ok) {
        const json = await res.json()
        throw new Error(json.message ?? '削除に失敗しました')
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['org-statuses-admin'] })
      queryClient.invalidateQueries({ queryKey: ['org-statuses'] })
      setError('')
    },
    onError: (e: Error) => setError(e.message),
  })

  const startEdit = (status: Status) => {
    setEditingId(status.id)
    setForm({
      name: status.name,
      color: status.color,
      type: (status.type as 'issue' | 'project') || 'issue',
      order: status.order,
    })
    setShowForm(false)
    setError('')
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!form.name.trim()) {
      setError('ステータス名は必須です')
      return
    }
    if (!/^#[0-9A-Fa-f]{6}$/.test(form.color)) {
      setError('色は#RRGGBB形式で指定してください')
      return
    }
    if (editingId) {
      updateMutation.mutate({
        id: editingId,
        data: { name: form.name, color: form.color, order: form.order },
      })
    } else {
      createMutation.mutate(form)
    }
  }

  if (!currentOrg) {
    return (
      <div className="p-8 text-center text-gray-500 text-sm">
        組織を選択してください
      </div>
    )
  }

  return (
    <div className="max-w-3xl">
      {error && !showForm && !editingId && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700">
          {error}
        </div>
      )}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-bold text-gray-900">ステータス管理</h1>
          <p className="text-sm text-gray-500 mt-0.5">
            組織のステータス（Issue用・プロジェクト用）を管理します
          </p>
        </div>
        {!showForm && !editingId && (
          <button
            onClick={() => {
              setShowForm(true)
              setForm({ name: '', color: '#6B7280', type: 'issue', order: statuses.length + 1 })
              setError('')
            }}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors"
          >
            <Plus className="w-4 h-4" />
            ステータスを追加
          </button>
        )}
      </div>

      {(showForm || editingId) && (
        <div className="bg-white border border-gray-200 rounded-xl p-5 mb-6 shadow-sm">
          <h2 className="text-sm font-semibold text-gray-700 mb-4">
            {editingId ? 'ステータスを編集' : '新しいステータス'}
          </h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">ステータス名 *</label>
                <input
                  type="text"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  placeholder="例: 未着手"
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">色 (#RRGGBB) *</label>
                <div className="flex gap-2">
                  <input
                    type="color"
                    value={form.color}
                    onChange={(e) => setForm({ ...form, color: e.target.value })}
                    className="w-10 h-10 rounded border border-gray-300 cursor-pointer"
                  />
                  <input
                    type="text"
                    value={form.color}
                    onChange={(e) => setForm({ ...form, color: e.target.value })}
                    placeholder="#6B7280"
                    className="flex-1 border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              {!editingId && (
                <div>
                  <label className="block text-xs font-medium text-gray-600 mb-1">タイプ</label>
                  <select
                    value={form.type}
                    onChange={(e) => setForm({ ...form, type: e.target.value as 'issue' | 'project' })}
                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="issue">Issue</option>
                    <option value="project">プロジェクト</option>
                  </select>
                </div>
              )}
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">並び順</label>
                <input
                  type="number"
                  min={1}
                  value={form.order}
                  onChange={(e) => setForm({ ...form, order: parseInt(e.target.value) || 1 })}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>
            {error && <p className="text-sm text-red-500">{error}</p>}
            <div className="flex gap-2">
              <button
                type="submit"
                disabled={createMutation.isPending || updateMutation.isPending}
                className="flex items-center gap-1.5 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50 transition-colors"
              >
                <Check className="w-4 h-4" />
                {editingId ? '更新' : '追加'}
              </button>
              <button
                type="button"
                onClick={() => {
                  setShowForm(false)
                  setEditingId(null)
                  setError('')
                }}
                className="flex items-center gap-1.5 px-4 py-2 border border-gray-300 text-gray-600 rounded-lg text-sm hover:bg-gray-50 transition-colors"
              >
                <X className="w-4 h-4" />
                キャンセル
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="bg-white border border-gray-200 rounded-xl shadow-sm overflow-hidden">
        {isLoading ? (
          <div className="p-8 text-center text-gray-400 text-sm">読み込み中...</div>
        ) : statuses.length === 0 ? (
          <div className="p-8 text-center text-gray-400 text-sm">
            組織直下のステータスがまだありません。「ステータスを追加」から作成してください。
          </div>
        ) : (
          <table className="w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase tracking-wide">ステータス名</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase tracking-wide w-24">タイプ</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase tracking-wide w-20">色</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase tracking-wide w-16">順</th>
                <th className="px-4 py-3 w-20"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {statuses.map((s) => (
                <tr key={s.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-4 py-3">
                    <span className="font-medium text-gray-900 text-sm">{s.name}</span>
                  </td>
                  <td className="px-4 py-3">
                    <span className="text-xs text-gray-600">
                      {s.type === 'project' ? 'プロジェクト' : 'Issue'}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className="inline-block w-6 h-6 rounded border border-gray-200"
                      style={{ backgroundColor: s.color }}
                      title={s.color}
                    />
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-500">{s.order}</td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-1 justify-end">
                      {(s.status_key !== 'sts_start' && s.status_key !== 'sts_goal') ? (
                        <>
                          <button
                            onClick={() => startEdit(s)}
                            className="p-1.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                            title="編集"
                          >
                            <Pencil className="w-3.5 h-3.5" />
                          </button>
                          <button
                            onClick={() => {
                              if (confirm(`「${s.name}」を削除しますか？`)) {
                                deleteMutation.mutate(s.id)
                              }
                            }}
                            className="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                            title="削除"
                          >
                            <Trash2 className="w-3.5 h-3.5" />
                          </button>
                        </>
                      ) : (
                        <span className="text-xs text-gray-400">システム</span>
                      )}
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
