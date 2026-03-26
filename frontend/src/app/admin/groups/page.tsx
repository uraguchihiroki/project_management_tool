'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Pencil, Trash2, X, Check } from 'lucide-react'
import type { Group } from '@/types'
import { SortableDndProvider, SortableTbody, DragHandle } from '@/components/SortableList'

import { useAuth } from '@/context/AuthContext'
import { useAuthFetchEnabled } from '@/hooks/useAuthFetchEnabled'
import { resolveApiBaseURL } from '@/lib/api'

async function fetchGroups(orgId: string): Promise<Group[]> {
  if (!orgId) return []
  const res = await fetch(`${resolveApiBaseURL()}/organizations/${orgId}/groups`)
  if (!res.ok) return []
  const json = await res.json()
  return json.data ?? []
}

export default function GroupsPage() {
  const { currentOrg } = useAuth()
  const authFetch = useAuthFetchEnabled()
  const queryClient = useQueryClient()
  const { data: groups = [], isLoading } = useQuery({
    queryKey: ['groups', currentOrg?.id],
    queryFn: () => fetchGroups(currentOrg?.id ?? ''),
    enabled: authFetch && !!currentOrg?.id,
  })

  const [showForm, setShowForm] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [form, setForm] = useState({ name: '' })
  const [error, setError] = useState('')

  const createMutation = useMutation({
    mutationFn: async (data: typeof form) => {
      const orgId = currentOrg?.id
      if (!orgId) throw new Error('組織が選択されていません')
      const res = await fetch(`${resolveApiBaseURL()}/organizations/${orgId}/groups`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: data.name }),
      })
      const json = await res.json().catch(() => ({}))
      if (!res.ok) {
        const msg = json.message ?? '作成に失敗しました'
        if (res.status === 404) {
          throw new Error(`${msg}（API エンドポイントが見つかりません。バックエンドが起動しているか確認してください）`)
        }
        throw new Error(msg)
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['groups'] })
      setShowForm(false)
      setForm({ name: '' })
      setError('')
    },
    onError: (e: Error) => setError(e.message),
  })

  const updateMutation = useMutation({
    mutationFn: async ({ id, data }: { id: string; data: typeof form }) => {
      const orgId = currentOrg?.id
      if (!orgId) throw new Error('組織が選択されていません')
      const res = await fetch(`${resolveApiBaseURL()}/organizations/${orgId}/groups/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: data.name }),
      })
      const json = await res.json().catch(() => ({}))
      if (!res.ok) {
        const msg = json.message ?? '更新に失敗しました'
        if (res.status === 404) {
          throw new Error(`${msg}（API エンドポイントが見つかりません）`)
        }
        throw new Error(msg)
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['groups'] })
      setEditingId(null)
      setForm({ name: '' })
      setError('')
    },
    onError: (e: Error) => setError(e.message),
  })

  const deleteMutation = useMutation({
    mutationFn: async (id: string) => {
      const orgId = currentOrg?.id
      if (!orgId) throw new Error('組織が選択されていません')
      const res = await fetch(`${resolveApiBaseURL()}/organizations/${orgId}/groups/${id}`, { method: 'DELETE' })
      if (!res.ok) {
        const json = await res.json().catch(() => ({}))
        const msg = json.message ?? '削除に失敗しました'
        if (res.status === 404) throw new Error(`${msg}（API エンドポイントが見つかりません）`)
        throw new Error(msg)
      }
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['groups'] }),
    onError: (e: Error) => setError(e.message),
  })

  const [reorderPending, setReorderPending] = useState(false)
  const reorderMutation = useMutation({
    mutationFn: async (ids: string[]) => {
      const orgId = currentOrg?.id
      if (!orgId) throw new Error('組織が選択されていません')
      const res = await fetch(`${resolveApiBaseURL()}/organizations/${orgId}/groups/reorder`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ids }),
      })
      if (!res.ok) throw new Error('並び替えに失敗しました')
    },
    onMutate: () => setReorderPending(true),
    onSettled: () => setReorderPending(false),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['groups'] }),
  })

  const startEdit = (group: Group) => {
    setEditingId(group.id)
    setForm({ name: group.name })
    setShowForm(false)
    setError('')
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!form.name.trim()) {
      setError('グループ名は必須です')
      return
    }
    if (form.name.length > 200) {
      setError('グループ名は200文字以内で指定してください')
      return
    }
    if (editingId !== null) {
      updateMutation.mutate({ id: editingId, data: form })
    } else {
      createMutation.mutate(form)
    }
  }

  if (!currentOrg?.id) {
    return <div className="text-sm text-gray-500">組織を選択してください</div>
  }

  return (
    <div className="max-w-3xl">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-bold text-gray-900">グループ管理</h1>
          <p className="text-sm text-gray-500 mt-0.5">グループ（開発部、営業部、委員会など）を管理します</p>
        </div>
        {!showForm && editingId === null && (
          <button
            onClick={() => {
              setShowForm(true)
              setForm({ name: '' })
              setError('')
            }}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors"
          >
            <Plus className="w-4 h-4" />
            グループを追加
          </button>
        )}
      </div>

      {(showForm || editingId !== null) && (
        <div className="bg-white border border-gray-200 rounded-xl p-5 mb-6 shadow-sm">
          <h2 className="text-sm font-semibold text-gray-700 mb-4">{editingId !== null ? 'グループを編集' : '新しいグループ'}</h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">グループ名 *</label>
              <input
                type="text"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                placeholder="例: 開発部、予算委員会（200文字以内）"
                maxLength={200}
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
        ) : groups.length === 0 ? (
          <div className="p-8 text-center text-gray-400 text-sm">
            グループがまだありません。「グループを追加」から作成してください。
          </div>
        ) : (
          <SortableDndProvider
            items={groups}
            itemId={(g) => g.id}
            onReorder={(ids) => reorderMutation.mutate(ids)}
            disabled={reorderPending}
          >
            <table className="w-full">
              <thead className="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th className="w-10 px-2 py-3"></th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase tracking-wide">グループ名</th>
                  <th className="px-4 py-3 w-20"></th>
                </tr>
              </thead>
              <SortableTbody
                items={groups}
                itemId={(g) => g.id}
                disabled={reorderPending}
                tbodyClassName="divide-y divide-gray-100"
                renderItem={(group, { handleProps, setNodeRef, style }) => (
                  <tr ref={setNodeRef} style={style} className="hover:bg-gray-50 transition-colors">
                    <td className="px-2 py-3">
                      <DragHandle handleProps={handleProps} />
                    </td>
                    <td className="px-4 py-3">
                      <span className="font-medium text-gray-900 text-sm">{group.name}</span>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-1 justify-end">
                        <button
                          onClick={() => startEdit(group)}
                          className="p-1.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                          title="編集"
                        >
                          <Pencil className="w-3.5 h-3.5" />
                        </button>
                        <button
                          onClick={() => {
                            if (confirm(`「${group.name}」を削除しますか？`)) {
                              deleteMutation.mutate(group.id)
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
                )}
              />
            </table>
          </SortableDndProvider>
        )}
      </div>
    </div>
  )
}

