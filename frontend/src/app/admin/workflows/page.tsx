'use client'

import { useState } from 'react'
import Link from 'next/link'
import axios from 'axios'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Pencil, Trash2, X, Check, ChevronRight, GitBranch } from 'lucide-react'
import type { Workflow } from '@/types'
import { SortableDndProvider, SortableList, DragHandle } from '@/components/SortableList'
import { useAuth } from '@/context/AuthContext'
import {
  getWorkflows,
  createWorkflow,
  updateWorkflowMeta,
  deleteWorkflowApi,
  reorderWorkflowsApi,
} from '@/lib/api'
import { useAuthFetchEnabled } from '@/hooks/useAuthFetchEnabled'

export default function WorkflowsPage() {
  const queryClient = useQueryClient()
  const { currentOrg } = useAuth()
  const authFetch = useAuthFetchEnabled()
  const { data: workflows = [], isLoading } = useQuery({
    queryKey: ['workflows', currentOrg?.id],
    queryFn: () => getWorkflows(),
    enabled: authFetch && !!currentOrg?.id,
  })

  const [showForm, setShowForm] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [form, setForm] = useState({ name: '', description: '' })
  const [error, setError] = useState('')

  const createMutation = useMutation({
    mutationFn: async (data: typeof form) => {
      if (!currentOrg?.id) {
        throw new Error('組織が選択されていません。プロジェクト一覧で組織を選び直してください。')
      }
      return createWorkflow({
        organization_id: currentOrg.id,
        name: data.name.trim(),
        description: data.description ?? '',
      })
    },
    onSuccess: async (created) => {
      const key = ['workflows', currentOrg?.id] as const
      // 一覧が即座に更新されるようキャッシュに追加（invalidate だけだと再取得が遅い／取りこぼす環境がある）
      queryClient.setQueryData<Workflow[]>(key, (old) => {
        const list = old ?? []
        if (list.some((w) => w.id === created.id)) return list
        return [...list, created]
      })
      await queryClient.invalidateQueries({ queryKey: key })
      setShowForm(false)
      setForm({ name: '', description: '' })
      setError('')
    },
    onError: (e: unknown) => {
      const msg = axios.isAxiosError(e)
        ? (e.response?.data as { message?: string } | undefined)?.message
        : undefined
      setError(typeof msg === 'string' ? msg : e instanceof Error ? e.message : '作成に失敗しました')
    },
  })

  const updateMutation = useMutation({
    mutationFn: async ({ id, data }: { id: number; data: { name: string; description: string } }) => {
      await updateWorkflowMeta(id, data)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflows'] })
      setEditingId(null)
      setForm({ name: '', description: '' })
      setError('')
    },
    onError: (e: Error) => setError(e.message),
  })

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => {
      await deleteWorkflowApi(id)
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['workflows'] }),
  })

  const [reorderPending, setReorderPending] = useState(false)
  const reorderMutation = useMutation({
    mutationFn: async (ids: number[]) => {
      await reorderWorkflowsApi(ids)
    },
    onMutate: () => setReorderPending(true),
    onSettled: () => setReorderPending(false),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['workflows'] }),
  })

  const startEdit = (wf: Workflow) => {
    setEditingId(wf.id)
    setForm({ name: wf.name, description: wf.description ?? '' })
    setShowForm(false)
    setError('')
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!form.name.trim()) { setError('名前は必須です'); return }
    if (editingId !== null) {
      updateMutation.mutate({ id: editingId, data: { name: form.name, description: form.description } })
    } else {
      createMutation.mutate(form)
    }
  }

  return (
    <div className="max-w-3xl">
      {!currentOrg?.id && (
        <div className="mb-4 rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
          現在の組織が選択されていません。プロジェクト一覧に戻り、右上の組織から選択してください。
        </div>
      )}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-bold text-gray-900">ワークフロー管理</h1>
          <p className="text-sm text-gray-500 mt-0.5">承認フローを定義します</p>
        </div>
        {!showForm && editingId === null && (
          <button
            type="button"
            disabled={!currentOrg?.id}
            title={!currentOrg?.id ? '組織を選択してください' : undefined}
            onClick={() => {
              setShowForm(true)
              setForm({ name: '', description: '' })
              setError('')
            }}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Plus className="w-4 h-4" />
            ワークフローを追加
          </button>
        )}
      </div>

      {/* 追加/編集フォーム */}
      {(showForm || editingId !== null) && (
        <div className="bg-white border border-gray-200 rounded-xl p-5 mb-6 shadow-sm">
          <h2 className="text-sm font-semibold text-gray-700 mb-4">
            {editingId !== null ? 'ワークフローを編集' : '新しいワークフロー'}
          </h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">ワークフロー名 *</label>
              <input
                type="text"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                placeholder="例: 通常承認フロー"
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">説明</label>
              <input
                type="text"
                value={form.description}
                onChange={(e) => setForm({ ...form, description: e.target.value })}
                placeholder="例: 一般的な業務申請に使用"
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

      {/* ワークフロー一覧 */}
      {isLoading ? (
        <div className="bg-white border border-gray-200 rounded-xl p-8 text-center text-gray-400 text-sm">
          読み込み中...
        </div>
      ) : workflows.length === 0 ? (
        <div className="bg-white border border-gray-200 rounded-xl p-8 text-center text-gray-400 text-sm">
          ワークフローがまだありません。「ワークフローを追加」から作成してください。
        </div>
      ) : (
        <div className="bg-white border border-gray-200 rounded-xl shadow-sm overflow-hidden">
          <SortableDndProvider
            items={workflows}
            itemId={(w) => String(w.id)}
            onReorder={(ids) => reorderMutation.mutate(ids.map(Number))}
            disabled={reorderPending}
          >
            <SortableList
              items={workflows}
              itemId={(w) => String(w.id)}
              onReorder={(ids) => reorderMutation.mutate(ids.map(Number))}
              disabled={reorderPending}
              renderItem={(wf, { handleProps, setNodeRef, style }) => (
                <div ref={setNodeRef} style={style} className="flex items-center gap-3 px-4 py-3 hover:bg-gray-50 transition-colors border-t border-gray-100 first:border-t-0">
                  <DragHandle handleProps={handleProps} />
                  <GitBranch className="w-4 h-4 text-blue-500 flex-shrink-0" />
                  <div className="flex-1 min-w-0">
                    <p className="font-medium text-gray-900 text-sm">{wf.name}</p>
                    {wf.description && (
                      <p className="text-xs text-gray-400 truncate">{wf.description}</p>
                    )}
                  </div>
                  <div className="flex items-center gap-1 flex-shrink-0">
                    <Link
                      href={`/admin/workflows/${wf.id}`}
                      className="flex items-center gap-1 px-2.5 py-1.5 text-xs text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                    >
                      ステップ編集
                      <ChevronRight className="w-3 h-3" />
                    </Link>
                    <button
                      onClick={() => startEdit(wf)}
                      className="p-1.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                      title="編集"
                    >
                      <Pencil className="w-3.5 h-3.5" />
                    </button>
                    <button
                      onClick={() => {
                        if (confirm(`「${wf.name}」を削除しますか？ステップもすべて削除されます。`)) {
                          deleteMutation.mutate(wf.id)
                        }
                      }}
                      className="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                      title="削除"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  </div>
                </div>
              )}
            />
          </SortableDndProvider>
        </div>
      )}
    </div>
  )
}
