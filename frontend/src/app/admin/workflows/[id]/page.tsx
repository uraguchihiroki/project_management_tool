'use client'

import { useState, useEffect } from 'react'
import { use } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ChevronLeft, Pencil, Plus, Trash2, X } from 'lucide-react'
import { useAuth } from '@/context/AuthContext'
import { useAuthFetchEnabled } from '@/hooks/useAuthFetchEnabled'
import type { Status } from '@/types'
import {
  createWorkflowStatus,
  deleteWorkflowApi,
  getWorkflow,
  getWorkflowStatuses,
  updateStatus,
  updateWorkflowMeta,
} from '@/lib/api'

type StatusDialogMode = 'create' | 'edit'

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

  const orgMatches = !!workflow && !!currentOrg && workflow.organization_id === currentOrg.id

  const { data: statuses = [], isLoading: statusesLoading } = useQuery({
    queryKey: ['workflow', currentOrg?.id, id, 'statuses'],
    queryFn: () => getWorkflowStatuses(id),
    enabled: authFetch && !!id && !!workflow && orgMatches,
  })

  const [editing, setEditing] = useState(false)
  const [form, setForm] = useState({ name: '', description: '' })
  const [error, setError] = useState('')

  const [statusDialogOpen, setStatusDialogOpen] = useState(false)
  const [statusDialogMode, setStatusDialogMode] = useState<StatusDialogMode>('create')
  const [statusDialogStatusId, setStatusDialogStatusId] = useState<string | null>(null)
  const [statusDialogForm, setStatusDialogForm] = useState({
    name: '',
    color: '#6B7280',
    order: '',
  })
  const [statusDialogError, setStatusDialogError] = useState('')

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
    mutationFn: (data: { name: string; color: string; order?: number }) => createWorkflowStatus(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflow', currentOrg?.id, id, 'statuses'] })
      setStatusDialogOpen(false)
      setStatusDialogError('')
    },
    onError: (e: Error) => setStatusDialogError(e.message),
  })

  const updateStatusMutation = useMutation({
    mutationFn: ({ statusId, data }: { statusId: string; data: { name: string; color: string; order: number } }) =>
      updateStatus(statusId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflow', currentOrg?.id, id, 'statuses'] })
      setStatusDialogOpen(false)
      setStatusDialogError('')
    },
    onError: (e: Error) => setStatusDialogError(e.message),
  })

  const openCreateStatusDialog = () => {
    setStatusDialogMode('create')
    setStatusDialogStatusId(null)
    setStatusDialogForm({ name: '', color: '#6B7280', order: '' })
    setStatusDialogError('')
    setStatusDialogOpen(true)
  }

  const openEditStatusDialog = (status: Status) => {
    setStatusDialogMode('edit')
    setStatusDialogStatusId(status.id)
    setStatusDialogForm({
      name: status.name,
      color: status.color,
      order: String(status.order),
    })
    setStatusDialogError('')
    setStatusDialogOpen(true)
  }

  const closeStatusDialog = () => {
    if (addStatusMutation.isPending || updateStatusMutation.isPending) return
    setStatusDialogOpen(false)
    setStatusDialogError('')
  }

  const submitStatusDialog = () => {
    const name = statusDialogForm.name.trim()
    const color = statusDialogForm.color.trim()
    const orderStr = statusDialogForm.order.trim()
    const orderParsed = orderStr === '' ? NaN : parseInt(orderStr, 10)

    if (!name) {
      setStatusDialogError('ステータス名は必須です')
      return
    }
    if (!/^#[0-9A-Fa-f]{6}$/.test(color)) {
      setStatusDialogError('色は#RRGGBB形式で指定してください')
      return
    }
    if (orderStr !== '' && (Number.isNaN(orderParsed) || orderParsed <= 0)) {
      setStatusDialogError('表示順は1以上の整数で指定してください')
      return
    }

    if (statusDialogMode === 'create') {
      addStatusMutation.mutate({
        name,
        color,
        ...(orderStr !== '' ? { order: orderParsed } : {}),
      })
      return
    }

    if (!statusDialogStatusId) {
      setStatusDialogError('編集対象のステータスが特定できません')
      return
    }
    updateStatusMutation.mutate({
      statusId: statusDialogStatusId,
      data: {
        name,
        color,
        order: Number.isNaN(orderParsed) ? 1 : orderParsed,
      },
    })
  }

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
  const statusDialogTitle = statusDialogMode === 'create' ? 'ステータスを追加' : 'ステータスを編集'
  const statusDialogSaving = addStatusMutation.isPending || updateStatusMutation.isPending

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
            {!orgMismatch && currentOrg && (
              <p className="mt-1 text-sm text-gray-500">
                選択中の組織: <span className="font-medium text-gray-800">{currentOrg.name}</span>
              </p>
            )}
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
              onClick={openCreateStatusDialog}
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
                  <th className="px-3 py-2 font-medium w-20">操作</th>
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
                    <td className="px-3 py-2">
                      {s.status_key === 'sts_start' || s.status_key === 'sts_goal' ? (
                        <span className="text-xs text-gray-400">システム</span>
                      ) : (
                        <button
                          type="button"
                          onClick={() => openEditStatusDialog(s)}
                          className="p-1.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded"
                          title="編集"
                        >
                          <Pencil className="w-4 h-4" />
                        </button>
                      )}
                    </td>
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

      {!orgMismatch && statusDialogOpen && (
        <div className="fixed inset-0 z-50 bg-black/30 flex items-center justify-center p-4">
          <div className="w-full max-w-lg bg-white rounded-xl shadow-xl border border-gray-200">
            <div className="flex items-center justify-between px-5 py-4 border-b">
              <h3 className="text-base font-semibold text-gray-900">{statusDialogTitle}</h3>
              <button
                type="button"
                onClick={closeStatusDialog}
                className="p-1.5 rounded hover:bg-gray-100 text-gray-500"
                aria-label="閉じる"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
            <div className="p-5 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">名前（必須）</label>
                <input
                  value={statusDialogForm.name}
                  onChange={(e) => setStatusDialogForm((f) => ({ ...f, name: e.target.value }))}
                  className="w-full border rounded-lg px-3 py-2"
                  placeholder="例: レビュー待ち"
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">色</label>
                  <input
                    type="color"
                    value={statusDialogForm.color}
                    onChange={(e) => setStatusDialogForm((f) => ({ ...f, color: e.target.value }))}
                    className="h-10 w-14 rounded border cursor-pointer"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    表示順{statusDialogMode === 'create' ? '（任意）' : '（必須）'}
                  </label>
                  <input
                    type="number"
                    min={1}
                    value={statusDialogForm.order}
                    onChange={(e) => setStatusDialogForm((f) => ({ ...f, order: e.target.value }))}
                    placeholder={statusDialogMode === 'create' ? '自動' : '1'}
                    className="w-full border rounded-lg px-3 py-2"
                  />
                </div>
              </div>
              {statusDialogError && (
                <p className="text-sm text-red-600 bg-red-50 border border-red-200 rounded px-3 py-2">
                  {statusDialogError}
                </p>
              )}
            </div>
            <div className="px-5 py-4 border-t flex gap-2 justify-end">
              <button
                type="button"
                onClick={closeStatusDialog}
                className="px-3 py-1.5 text-sm border rounded-lg"
              >
                キャンセル
              </button>
              <button
                type="button"
                disabled={statusDialogSaving}
                onClick={submitStatusDialog}
                className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
              >
                {statusDialogMode === 'create' ? '追加する' : '更新する'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
