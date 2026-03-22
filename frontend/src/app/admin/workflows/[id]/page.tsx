'use client'

import { useState, useEffect } from 'react'
import { use } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ChevronLeft, Plus, Trash2 } from 'lucide-react'
import { useAuth } from '@/context/AuthContext'
import { useAuthFetchEnabled } from '@/hooks/useAuthFetchEnabled'
import {
  createWorkflowStatus,
  deleteWorkflowApi,
  getWorkflow,
  getWorkflowStatuses,
  updateWorkflowMeta,
} from '@/lib/api'

export default function WorkflowDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params)
  const router = useRouter()
  const queryClient = useQueryClient()
  const authFetch = useAuthFetchEnabled()
  const { currentOrg } = useAuth()

  const { data: workflow, isLoading } = useQuery({
    queryKey: ['workflow', currentOrg?.id, id],
    queryFn: () => getWorkflow(id),
    enabled: authFetch && !!id && !!currentOrg?.id,
  })

  const orgMatches =
    !!workflow && !!currentOrg && workflow.organization_id === currentOrg.id

  const { data: statuses = [], isLoading: statusesLoading } = useQuery({
    queryKey: ['workflow', currentOrg?.id, id, 'statuses'],
    queryFn: () => getWorkflowStatuses(id),
    enabled: authFetch && !!id && !!workflow && orgMatches,
  })

  const [editing, setEditing] = useState(false)
  const [form, setForm] = useState({ name: '', description: '' })
  const [error, setError] = useState('')
  const [showAddStatus, setShowAddStatus] = useState(false)
  const [statusForm, setStatusForm] = useState({
    name: '',
    color: '#6B7280',
    type: 'issue' as 'issue' | 'project',
    order: '',
  })

  useEffect(() => {
    if (workflow) {
      setForm({ name: workflow.name, description: workflow.description ?? '' })
    }
  }, [workflow])

  const saveMutation = useMutation({
    mutationFn: () => updateWorkflowMeta(id, { name: form.name, description: form.description }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflow', currentOrg?.id, id] })
      queryClient.invalidateQueries({ queryKey: ['workflows'] })
      setEditing(false)
      setError('')
    },
    onError: (e: Error) => setError(e.message),
  })

  const deleteMutation = useMutation({
    mutationFn: () => deleteWorkflowApi(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflows'] })
      router.push('/admin/workflows')
    },
    onError: (e: Error) => setError(e.message),
  })

  const addStatusMutation = useMutation({
    mutationFn: () => {
      const orderStr = statusForm.order.trim()
      const orderParsed = orderStr === '' ? NaN : parseInt(orderStr, 10)
      const order =
        !Number.isNaN(orderParsed) && orderParsed > 0 ? orderParsed : undefined
      return createWorkflowStatus(id, {
        name: statusForm.name.trim(),
        color: statusForm.color || '#6B7280',
        type: statusForm.type,
        ...(order !== undefined ? { order } : {}),
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflow', currentOrg?.id, id, 'statuses'] })
      setStatusForm({ name: '', color: '#6B7280', type: 'issue', order: '' })
      setShowAddStatus(false)
      setError('')
    },
    onError: (e: Error) => setError(e.message),
  })

  if (!authFetch) {
    return <div className="p-8 text-gray-500">読み込み中...</div>
  }
  if (!currentOrg?.id) {
    return (
      <div className="max-w-3xl mx-auto p-6">
        <div className="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
          現在の組織が選択されていません。プロジェクト一覧に戻り、右上の組織から選択してください。
        </div>
      </div>
    )
  }
  if (isLoading) {
    return <div className="p-8 text-gray-500">読み込み中...</div>
  }
  if (!workflow) {
    return <div className="p-8 text-gray-500">ワークフローが見つかりません</div>
  }

  const orgMismatch = !orgMatches

  return (
    <div className="max-w-3xl mx-auto p-6">
      <Link
        href="/admin/workflows"
        className="inline-flex items-center gap-1 text-sm text-gray-600 hover:text-gray-900 mb-6"
      >
        <ChevronLeft className="w-4 h-4" />
        一覧へ
      </Link>

      <div className="bg-white rounded-xl border border-gray-200 p-6">
        {orgMismatch && (
          <div className="mb-4 rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
            このワークフローは<strong>現在選択中の組織</strong>に属していません。右上で組織を切り替えるか、
            <Link href="/admin/workflows" className="text-blue-700 underline">
              ワークフロー一覧
            </Link>
            から開き直してください。
          </div>
        )}
        <div className="flex justify-between items-start gap-4">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">{workflow.name}</h1>
            <p className="mt-2 text-gray-600 whitespace-pre-wrap">{workflow.description || '—'}</p>
          </div>
          <div className="flex gap-2">
            {!orgMismatch && !editing ? (
              <button
                type="button"
                onClick={() => {
                  setForm({ name: workflow.name, description: workflow.description ?? '' })
                  setEditing(true)
                }}
                className="px-3 py-1.5 text-sm border rounded-lg hover:bg-gray-50"
              >
                編集
              </button>
            ) : !orgMismatch && editing ? (
              <>
                <button
                  type="button"
                  onClick={() => saveMutation.mutate()}
                  disabled={saveMutation.isPending || !form.name.trim()}
                  className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
                >
                  保存
                </button>
                <button
                  type="button"
                  onClick={() => setEditing(false)}
                  className="px-3 py-1.5 text-sm border rounded-lg"
                >
                  取消
                </button>
              </>
            ) : null}
            {!orgMismatch && (
              <button
                type="button"
                onClick={() => {
                  if (confirm('このワークフローを削除しますか？')) deleteMutation.mutate()
                }}
                className="p-2 text-red-600 hover:bg-red-50 rounded-lg"
                title="削除"
              >
                <Trash2 className="w-5 h-5" />
              </button>
            )}
          </div>
        </div>

        {!orgMismatch && editing && (
          <div className="mt-6 space-y-4 border-t pt-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">名前</label>
              <input
                value={form.name}
                onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                className="w-full border rounded-lg px-3 py-2"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">説明</label>
              <textarea
                value={form.description}
                onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
                rows={4}
                className="w-full border rounded-lg px-3 py-2"
              />
            </div>
          </div>
        )}
      </div>

      <div className="mt-8 bg-white rounded-xl border border-gray-200 p-6">
        <div className="flex flex-wrap items-center justify-between gap-3 mb-4">
          <h2 className="text-lg font-semibold text-gray-900">ステータス</h2>
          {!orgMismatch && (
            <button
              type="button"
              onClick={() => {
                setShowAddStatus((v) => !v)
                setError('')
              }}
              className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              <Plus className="w-4 h-4" />
              ステータスを追加
            </button>
          )}
        </div>

        {orgMismatch && (
          <p className="text-sm text-gray-500">
            選択中の組織に属するワークフローのみ、ステータス一覧を表示・編集できます。
          </p>
        )}

        {!orgMismatch && showAddStatus && (
          <div className="mb-6 p-4 rounded-lg border border-gray-200 bg-gray-50 space-y-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">名前（必須）</label>
              <input
                value={statusForm.name}
                onChange={(e) => setStatusForm((f) => ({ ...f, name: e.target.value }))}
                className="w-full border rounded-lg px-3 py-2 bg-white"
                placeholder="例: レビュー待ち"
              />
            </div>
            <div className="flex flex-wrap gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">色</label>
                <input
                  type="color"
                  value={statusForm.color}
                  onChange={(e) => setStatusForm((f) => ({ ...f, color: e.target.value }))}
                  className="h-10 w-14 rounded border cursor-pointer"
                />
              </div>
              <div className="flex-1 min-w-[140px]">
                <label className="block text-sm font-medium text-gray-700 mb-1">種別</label>
                <select
                  value={statusForm.type}
                  onChange={(e) =>
                    setStatusForm((f) => ({
                      ...f,
                      type: e.target.value as 'issue' | 'project',
                    }))
                  }
                  className="w-full border rounded-lg px-3 py-2 bg-white"
                >
                  <option value="issue">issue</option>
                  <option value="project">project</option>
                </select>
              </div>
              <div className="flex-1 min-w-[100px]">
                <label className="block text-sm font-medium text-gray-700 mb-1">表示順（任意）</label>
                <input
                  type="number"
                  min={1}
                  value={statusForm.order}
                  onChange={(e) => setStatusForm((f) => ({ ...f, order: e.target.value }))}
                  placeholder="自動"
                  className="w-full border rounded-lg px-3 py-2 bg-white"
                />
              </div>
            </div>
            <div className="flex gap-2 pt-1">
              <button
                type="button"
                disabled={addStatusMutation.isPending || !statusForm.name.trim()}
                onClick={() => addStatusMutation.mutate()}
                className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
              >
                追加する
              </button>
              <button
                type="button"
                onClick={() => {
                  setShowAddStatus(false)
                  setStatusForm({ name: '', color: '#6B7280', type: 'issue', order: '' })
                }}
                className="px-3 py-1.5 text-sm border rounded-lg"
              >
                閉じる
              </button>
            </div>
          </div>
        )}

        {!orgMismatch && statusesLoading && (
          <p className="text-sm text-gray-500">ステータスを読み込み中...</p>
        )}
        {!orgMismatch && !statusesLoading && statuses.length === 0 && (
          <p className="text-sm text-gray-500">まだステータスがありません。「ステータスを追加」から作成できます。</p>
        )}
        {!orgMismatch && !statusesLoading && statuses.length > 0 && (
          <div className="overflow-x-auto rounded-lg border border-gray-200">
            <table className="min-w-full text-sm">
              <thead className="bg-gray-50 text-left text-gray-600">
                <tr>
                  <th className="px-3 py-2 font-medium">順</th>
                  <th className="px-3 py-2 font-medium">名前</th>
                  <th className="px-3 py-2 font-medium">色</th>
                  <th className="px-3 py-2 font-medium">種別</th>
                </tr>
              </thead>
              <tbody>
                {statuses.map((s) => (
                  <tr key={s.id} className="border-t border-gray-100">
                    <td className="px-3 py-2 text-gray-700">{s.order}</td>
                    <td className="px-3 py-2 font-medium text-gray-900">{s.name}</td>
                    <td className="px-3 py-2">
                      <span
                        className="inline-block w-6 h-6 rounded border border-gray-200 align-middle"
                        style={{ backgroundColor: s.color }}
                        title={s.color}
                      />
                      <span className="ml-2 text-gray-600 font-mono text-xs">{s.color}</span>
                    </td>
                    <td className="px-3 py-2 text-gray-700">{s.type ?? '—'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        <p className="mt-4 text-xs text-gray-500">
          追加時は同一ワークフロー内の遷移（全ペア）が再計算され、Issue のステータス変更と整合します。
        </p>

        {error && <p className="mt-4 text-sm text-red-600">{error}</p>}
      </div>
    </div>
  )
}
