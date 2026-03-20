'use client'

import { useState, useEffect } from 'react'
import { use } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2, X, ChevronLeft, Pencil, GitBranch, Check } from 'lucide-react'
import type { Workflow, Status } from '@/types'
import { SortableDndProvider, SortableList, DragHandle } from '@/components/SortableList'
import { useAuth } from '@/context/AuthContext'

const API = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'

async function fetchWorkflow(id: string): Promise<Workflow> {
  const res = await fetch(`${API}/workflows/${id}`)
  const json = await res.json()
  return json.data
}

async function updateWorkflow(id: string, data: { name: string; description: string }) {
  const res = await fetch(`${API}/workflows/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error('更新に失敗しました')
}

async function fetchOrgStatuses(orgId: string): Promise<Status[]> {
  const res = await fetch(`${API}/organizations/${orgId}/statuses?type=issue&exclude_system=1`)
  const json = await res.json()
  const data: Status[] = json.data ?? []
  return data.filter((s) => !s.project_id && s.organization_id)
}

const emptyStep: { status_id: string; description: string; threshold: number } = {
  status_id: '',
  description: '',
  threshold: 10,
}

const isUserStep = (s: { status?: { status_key?: string } }) =>
  s.status?.status_key !== 'sts_start' && s.status?.status_key !== 'sts_goal'

export default function WorkflowDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params)
  const router = useRouter()
  const queryClient = useQueryClient()
  const { currentOrg } = useAuth()

  const { data: workflow, isLoading } = useQuery({
    queryKey: ['workflow', id],
    queryFn: () => fetchWorkflow(id),
  })

  const { data: statuses = [] } = useQuery({
    queryKey: ['org-statuses', currentOrg?.id],
    queryFn: () => fetchOrgStatuses(currentOrg!.id),
    enabled: !!currentOrg?.id,
  })

  const [showAddForm, setShowAddForm] = useState(false)
  const [stepForm, setStepForm] = useState(emptyStep)
  const [error, setError] = useState('')
  const [editingWorkflow, setEditingWorkflow] = useState(false)
  const [workflowForm, setWorkflowForm] = useState({ name: '', description: '' })

  const updateWorkflowMutation = useMutation({
    mutationFn: (data: { name: string; description: string }) => updateWorkflow(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflow', id] })
      queryClient.invalidateQueries({ queryKey: ['workflows'] })
      setEditingWorkflow(false)
      setError('')
    },
    onError: (e: Error) => setError(e.message),
  })

  const addStepMutation = useMutation({
    mutationFn: async (data: typeof emptyStep) => {
      if (!data.status_id) throw new Error('ステータスは必須です')
      if (data.threshold < 1) throw new Error('閾値は1以上で指定してください')
      const body = {
        status_id: data.status_id,
        description: data.description,
        threshold: data.threshold,
        approval_objects: [],
      }
      const res = await fetch(`${API}/workflows/${id}/steps`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
      if (!res.ok) {
        const json = await res.json().catch(() => ({}))
        throw new Error(json.message ?? '追加に失敗しました')
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflow', id] })
      setShowAddForm(false)
      setStepForm(emptyStep)
      setError('')
    },
    onError: (e: Error) => setError(e.message),
  })

  const deleteStepMutation = useMutation({
    mutationFn: async (stepId: number) => {
      await fetch(`${API}/workflows/${id}/steps/${stepId}`, { method: 'DELETE' })
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['workflow', id] }),
  })

  const deleteWorkflowMutation = useMutation({
    mutationFn: async () => {
      const res = await fetch(`${API}/workflows/${id}`, { method: 'DELETE' })
      if (!res.ok) throw new Error('ワークフロー削除に失敗しました')
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflows'] })
      router.push('/admin/workflows')
    },
  })

  const [reorderPending, setReorderPending] = useState(false)
  const [localStepOrder, setLocalStepOrder] = useState<number[]>([])
  const [hasReorderChanges, setHasReorderChanges] = useState(false)
  const reorderStepsMutation = useMutation({
    mutationFn: async (ids: number[]) => {
      const res = await fetch(`${API}/workflows/${id}/steps/reorder`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ids }),
      })
      if (!res.ok) throw new Error('並び替えに失敗しました')
    },
    onMutate: () => setReorderPending(true),
    onSettled: () => setReorderPending(false),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflow', id] })
      setHasReorderChanges(false)
    },
  })

  const allSteps = workflow?.steps ?? []
  const userSteps = [...allSteps.filter(isUserStep)].sort((a, b) => (a.order ?? 0) - (b.order ?? 0))

  useEffect(() => {
    if (userSteps.length > 0 && !hasReorderChanges) {
      setLocalStepOrder(userSteps.map((s) => s.id))
    }
  }, [workflow?.id, userSteps.length, hasReorderChanges])

  const orderedUserSteps =
    localStepOrder.length === userSteps.length
      ? localStepOrder
          .map((id) => userSteps.find((s) => s.id === id))
          .filter((s): s is NonNullable<typeof s> => s != null)
      : userSteps

  useEffect(() => {
    if (workflow) {
      setWorkflowForm({ name: workflow.name, description: workflow.description ?? '' })
    }
  }, [workflow])

  if (isLoading) {
    return <div className="text-gray-400 text-sm">読み込み中...</div>
  }
  if (!workflow) {
    return <div className="text-red-500 text-sm">ワークフローが見つかりません</div>
  }

  const renderStepForm = (onSubmit: (data: typeof emptyStep) => void, loading: boolean) => (
    <form
      onSubmit={(e) => { e.preventDefault(); onSubmit(stepForm) }}
      className="space-y-3"
    >
      <div className="grid grid-cols-12 gap-3">
        <div className="col-span-4">
          <label className="block text-xs font-medium text-gray-500 mb-1">ステータス *</label>
          <select
            value={stepForm.status_id}
            onChange={(e) => setStepForm((prev) => ({ ...prev, status_id: e.target.value }))}
            className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">選択</option>
            {statuses.map((s) => (
              <option key={s.id} value={s.id}>{s.name}</option>
            ))}
          </select>
        </div>
        <div className="col-span-2">
          <label className="block text-xs font-medium text-gray-500 mb-1">閾値</label>
          <input
            type="number"
            min={1}
            max={99999}
            value={stepForm.threshold}
            onChange={(e) => setStepForm((prev) => ({ ...prev, threshold: parseInt(e.target.value) || 1 }))}
            className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        <div className="col-span-2 flex gap-1.5 items-end">
          <button
            type="submit"
            disabled={loading}
            className="flex items-center gap-1 px-3 py-2 bg-blue-600 text-white rounded-lg text-xs font-medium hover:bg-blue-700 disabled:opacity-50 transition-colors"
          >
            <Plus className="w-3.5 h-3.5" />
            追加
          </button>
          <button
            type="button"
            onClick={() => { setShowAddForm(false); setStepForm(emptyStep); setError('') }}
            className="flex items-center gap-1 px-2.5 py-2 border border-gray-300 text-gray-600 rounded-lg text-xs hover:bg-gray-50 transition-colors"
          >
            <X className="w-3.5 h-3.5" />
          </button>
        </div>
      </div>
      <div>
        <label className="block text-xs font-medium text-gray-500 mb-1">説明</label>
        <input
          type="text"
          value={stepForm.description}
          onChange={(e) => setStepForm((prev) => ({ ...prev, description: e.target.value }))}
          placeholder="ステップの説明（任意）"
          className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>
      <p className="text-xs text-gray-400">追加後、ステップをクリックして承認オブジェクトを設定できます。</p>
    </form>
  )

  const handleWorkflowSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!workflowForm.name.trim()) { setError('ワークフロー名は必須です'); return }
    updateWorkflowMutation.mutate(workflowForm)
  }

  return (
    <div className="max-w-3xl">
      {/* ヘッダー */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <Link
            href="/admin/workflows"
            className="flex items-center gap-1 text-sm text-gray-500 hover:text-gray-700 transition-colors"
          >
            <ChevronLeft className="w-4 h-4" />
            ワークフロー一覧
          </Link>
          <span className="text-gray-300">/</span>
          {editingWorkflow ? (
            <form onSubmit={handleWorkflowSubmit} className="flex items-center gap-2 flex-1">
              <input
                type="text"
                value={workflowForm.name}
                onChange={(e) => setWorkflowForm((p) => ({ ...p, name: e.target.value }))}
                className="border border-gray-300 rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="ワークフロー名"
              />
              <button
                type="submit"
                disabled={updateWorkflowMutation.isPending}
                className="px-3 py-1.5 bg-blue-600 text-white rounded-lg text-xs font-medium hover:bg-blue-700 disabled:opacity-50"
              >
                保存
              </button>
              <button
                type="button"
                onClick={() => { setEditingWorkflow(false); setWorkflowForm({ name: workflow.name, description: workflow.description ?? '' }); setError('') }}
                className="px-3 py-1.5 border border-gray-300 text-gray-600 rounded-lg text-xs hover:bg-gray-50"
              >
                キャンセル
              </button>
            </form>
          ) : (
            <>
              <GitBranch className="w-5 h-5 text-blue-500 flex-shrink-0" />
              <h1 className="text-xl font-bold text-gray-900">{workflow.name}</h1>
              <button
                onClick={() => setEditingWorkflow(true)}
                className="p-1.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                title="ワークフローを編集"
              >
                <Pencil className="w-4 h-4" />
              </button>
            </>
          )}
        </div>
      </div>

      {error && <p className="text-sm text-red-500 mb-4">{error}</p>}

      {!editingWorkflow && workflow.description && (
        <p className="text-sm text-gray-500 mb-6">{workflow.description}</p>
      )}
      {editingWorkflow && (
        <div className="mb-6">
          <label className="block text-xs font-medium text-gray-600 mb-1">説明</label>
          <input
            type="text"
            value={workflowForm.description}
            onChange={(e) => setWorkflowForm((p) => ({ ...p, description: e.target.value }))}
            placeholder="例: 一般的な業務申請に使用"
            className="w-full max-w-md border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
      )}

      {/* ステップ一覧 */}
      <div className="bg-white border border-gray-200 rounded-xl shadow-sm overflow-hidden mb-4">
        <div className="flex items-center justify-between px-4 py-3 border-b border-gray-100">
          <h2 className="text-sm font-semibold text-gray-700">承認ステップ</h2>
          <div className="flex items-center gap-3">
            {userSteps.length > 1 && (
              <>
                <button
                  onClick={() => reorderStepsMutation.mutate(localStepOrder)}
                  disabled={!hasReorderChanges || reorderPending}
                  className="flex items-center gap-1 px-3 py-1.5 bg-blue-600 text-white rounded-lg text-xs font-medium hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <Check className="w-3.5 h-3.5" />
                  保存
                </button>
                <button
                  onClick={() => {
                    setHasReorderChanges(false)
                    setLocalStepOrder(userSteps.map((s) => s.id))
                  }}
                  disabled={!hasReorderChanges}
                  className="px-3 py-1.5 border border-gray-300 text-gray-600 rounded-lg text-xs hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  キャンセル
                </button>
              </>
            )}
            <span className="text-xs text-gray-400">{userSteps.length} ステップ</span>
          </div>
        </div>

        {userSteps.length === 0 ? (
          <div className="p-8 text-center text-gray-400 text-sm">
            ステップがまだありません。「ステップを追加」から作成してください。
          </div>
        ) : (
          <div className="divide-y divide-gray-100">
            <SortableDndProvider
              items={orderedUserSteps}
              itemId={(s) => String(s.id)}
              onReorder={(ids) => {
                setLocalStepOrder(ids.map(Number))
                setHasReorderChanges(true)
              }}
              disabled={reorderPending}
            >
              <SortableList
                items={orderedUserSteps}
                itemId={(s) => String(s.id)}
                onReorder={(ids) => {
                  setLocalStepOrder(ids.map(Number))
                  setHasReorderChanges(true)
                }}
                disabled={reorderPending}
                renderItem={(s, { handleProps, setNodeRef, style }) => (
                  <div ref={setNodeRef} style={style} className="flex items-center gap-3 px-4 py-3 hover:bg-gray-50 transition-colors">
                    <DragHandle handleProps={handleProps} />
                    <span className="w-6 h-6 rounded-full bg-blue-100 text-blue-700 text-xs font-bold flex items-center justify-center flex-shrink-0">
                      {orderedUserSteps.findIndex((x) => x.id === s.id) + 1}
                    </span>
                    <div className="flex-1 min-w-0">
                      <p className="font-medium text-gray-900 text-sm">
                        {s.status?.name ?? s.status_id}
                      </p>
                      <p className="text-xs text-gray-400">
                        {s.next_status_id && (
                          <>
                            閾値 {s.threshold ?? 10} 点
                            {(s.approval_objects?.length ?? 0) > 0 && (
                              <span className="ml-2">（承認オブジェクト {(s.approval_objects?.length ?? 0)} 件）</span>
                            )}
                            {s.next_status && (
                              <span className="ml-2">
                                → <span
                                  className="inline-block w-2 h-2 rounded-full mr-0.5"
                                  style={{ backgroundColor: s.next_status.color }}
                                />
                                {s.next_status.name}
                              </span>
                            )}
                          </>
                        )}
                      </p>
                    </div>
                    <div className="flex items-center gap-1 flex-shrink-0">
                      <Link
                        href={`/admin/workflows/${id}/steps/${s.id}`}
                        className="p-1.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                        title="編集"
                      >
                        <Pencil className="w-3.5 h-3.5" />
                      </Link>
                      <button
                        onClick={() => {
                          if (userSteps.length === 1) {
                            if (confirm('このステップを削除するとワークフローも削除されます。ワークフローを削除しますか？')) {
                              deleteWorkflowMutation.mutate()
                            }
                          } else if (confirm(`「${s.status?.name ?? s.status_id}」を削除しますか？`)) {
                            deleteStepMutation.mutate(s.id)
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

        {/* ステップ追加フォーム */}
        {showAddForm && (
          <div className="px-4 py-3 border-t border-gray-100 bg-gray-50">
            {error && <p className="text-xs text-red-500 mb-2">{error}</p>}
            {renderStepForm(
              (data) => {
                if (!data.status_id) { setError('ステータスは必須です'); return }
                addStepMutation.mutate(data)
              },
              addStepMutation.isPending
            )}
          </div>
        )}
      </div>

      {!showAddForm && (
        <button
          onClick={() => { setShowAddForm(true); setStepForm(emptyStep); setError('') }}
          className="flex items-center gap-2 px-4 py-2 border border-dashed border-gray-300 text-gray-500 rounded-lg text-sm hover:border-blue-400 hover:text-blue-600 transition-colors w-full justify-center"
        >
          <Plus className="w-4 h-4" />
          ステップを追加
        </button>
      )}
    </div>
  )
}
