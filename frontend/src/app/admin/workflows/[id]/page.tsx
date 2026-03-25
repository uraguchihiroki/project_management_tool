'use client'

import { useState, useEffect } from 'react'
import { use } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ChevronLeft, Pencil, Plus, Trash2, X } from 'lucide-react'
import { SortableDndProvider, SortableList, SortableTbody, DragHandle } from '@/components/SortableList'
import { useAuth } from '@/context/AuthContext'
import { useAuthFetchEnabled } from '@/hooks/useAuthFetchEnabled'
import type { Status, WorkflowTransition } from '@/types'
import {
  createWorkflowTransition,
  createWorkflowStatus,
  deleteWorkflowTransition,
  deleteStatus,
  deleteWorkflowApi,
  getWorkflow,
  getWorkflowStatuses,
  getWorkflowTransitions,
  reorderWorkflowStatuses,
  reorderWorkflowTransitions,
  updateStatus,
  updateWorkflowMeta,
  updateWorkflowTransition,
} from '@/lib/api'

type StatusDialogMode = 'create' | 'edit'
type TransitionInvalidReason = 'same' | 'duplicate'
type InvalidTransitionStrap = {
  clientId: string
  from_status_id: string
  to_status_id: string
}

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
  const { data: transitions = [], isLoading: transitionsLoading } = useQuery({
    queryKey: ['workflow', currentOrg?.id, id, 'transitions'],
    queryFn: () => getWorkflowTransitions(id),
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
  const [transitionDrafts, setTransitionDrafts] = useState<
    Record<number, { from_status_id: string; to_status_id: string }>
  >({})
  const [transitionBusyId, setTransitionBusyId] = useState<number | null>(null)
  const [transitionError, setTransitionError] = useState('')
  const [addingTransition, setAddingTransition] = useState(false)
  const [invalidTransitionStraps, setInvalidTransitionStraps] = useState<InvalidTransitionStrap[]>([])

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
    mutationFn: (data: { name: string; color: string; display_order?: number }) => createWorkflowStatus(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflow', currentOrg?.id, id, 'statuses'] })
      setStatusDialogOpen(false)
      setStatusDialogError('')
    },
    onError: (e: Error) => setStatusDialogError(e.message),
  })

  const updateStatusMutation = useMutation({
    mutationFn: ({
      statusId,
      data,
    }: {
      statusId: string
      data: { name: string; color: string; display_order: number }
    }) => updateStatus(statusId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflow', currentOrg?.id, id, 'statuses'] })
      setStatusDialogOpen(false)
      setStatusDialogError('')
    },
    onError: (e: Error) => setStatusDialogError(e.message),
  })

  const deleteStatusMutation = useMutation({
    mutationFn: (statusId: string) => deleteStatus(statusId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflow', currentOrg?.id, id, 'statuses'] })
      setStatusDialogError('')
      setError('')
    },
    onError: (e: Error) => setError(e.message),
  })

  const createTransitionMutation = useMutation({
    mutationFn: (data: { from_status_id: string; to_status_id: string }) => createWorkflowTransition(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflow', currentOrg?.id, id, 'transitions'] })
      setTransitionError('')
    },
    onError: (e: Error) => setTransitionError(e.message),
  })

  const deleteTransitionMutation = useMutation({
    mutationFn: (transitionId: number) => deleteWorkflowTransition(id, transitionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflow', currentOrg?.id, id, 'transitions'] })
      setTransitionError('')
    },
    onError: (e: Error) => setTransitionError(e.message),
  })

  const reorderStatusesMutation = useMutation({
    mutationFn: (statusIds: string[]) => reorderWorkflowStatuses(id, statusIds),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflow', currentOrg?.id, id, 'statuses'] })
    },
    onError: (e: Error) => setError(e.message),
  })

  const reorderTransitionsMutation = useMutation({
    mutationFn: (transitionIds: number[]) => reorderWorkflowTransitions(id, transitionIds),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflow', currentOrg?.id, id, 'transitions'] })
    },
    onError: (e: Error) => setTransitionError(e.message),
  })

  const updateTransitionMutation = useMutation({
    mutationFn: ({
      transitionId,
      data,
    }: {
      transitionId: number
      data: { from_status_id: string; to_status_id: string }
    }) => updateWorkflowTransition(id, transitionId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflow', currentOrg?.id, id, 'transitions'] })
      setTransitionError('')
    },
    onError: (e: Error) => setTransitionError(e.message),
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
      order: String(status.display_order),
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
        ...(orderStr !== '' ? { display_order: orderParsed } : {}),
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
        display_order: Number.isNaN(orderParsed) ? 1 : orderParsed,
      },
    })
  }

  const statusesByOrder = [...statuses].sort((a, b) => a.display_order - b.display_order)

  const statusReferencedInTransition = (statusId: string) =>
    transitions.some((t) => t.from_status_id === statusId || t.to_status_id === statusId)
  const initialFromStatusId = statusesByOrder[0]?.id
  const initialToStatusId = statusesByOrder[1]?.id

  const transitionReasonForPersistedRow = (
    rowId: number,
    fromStatusId: string,
    toStatusId: string
  ): TransitionInvalidReason | null => {
    if (fromStatusId === toStatusId) return 'same'
    for (const t of transitions) {
      if (t.id === rowId) continue
      const d = transitionDrafts[t.id]
      const ff = d?.from_status_id ?? t.from_status_id
      const tt = d?.to_status_id ?? t.to_status_id
      if (ff === fromStatusId && tt === toStatusId) return 'duplicate'
    }
    for (const s of invalidTransitionStraps) {
      if (s.from_status_id === fromStatusId && s.to_status_id === toStatusId) return 'duplicate'
    }
    return null
  }

  const transitionReasonForNewPair = (
    fromStatusId: string,
    toStatusId: string,
    excludeInvalidClientId?: string
  ): TransitionInvalidReason | null => {
    if (fromStatusId === toStatusId) return 'same'
    for (const t of transitions) {
      const d = transitionDrafts[t.id]
      const ff = d?.from_status_id ?? t.from_status_id
      const tt = d?.to_status_id ?? t.to_status_id
      if (ff === fromStatusId && tt === toStatusId) return 'duplicate'
    }
    for (const s of invalidTransitionStraps) {
      if (excludeInvalidClientId && s.clientId === excludeInvalidClientId) continue
      if (s.from_status_id === fromStatusId && s.to_status_id === toStatusId) return 'duplicate'
    }
    return null
  }

  const getRowTransition = (t: WorkflowTransition) => {
    const d = transitionDrafts[t.id]
    const from_status_id = d?.from_status_id ?? t.from_status_id
    const to_status_id = d?.to_status_id ?? t.to_status_id
    const reason = transitionReasonForPersistedRow(t.id, from_status_id, to_status_id)
    return { from_status_id, to_status_id, reason }
  }

  const invalidStrapReason = (s: InvalidTransitionStrap) =>
    transitionReasonForNewPair(s.from_status_id, s.to_status_id, s.clientId)

  const transitionStrapPending =
    createTransitionMutation.isPending ||
    deleteTransitionMutation.isPending ||
    updateTransitionMutation.isPending

  const addTransitionStrap = async () => {
    if (!initialFromStatusId || !initialToStatusId || addingTransition) return
    const reason = transitionReasonForNewPair(initialFromStatusId, initialToStatusId)
    if (reason) {
      const clientId = `invalid-${Date.now()}-${Math.random().toString(36).slice(2)}`
      setInvalidTransitionStraps((prev) => [
        ...prev,
        { clientId, from_status_id: initialFromStatusId, to_status_id: initialToStatusId },
      ])
      return
    }
    setAddingTransition(true)
    try {
      await createTransitionMutation.mutateAsync({
        from_status_id: initialFromStatusId,
        to_status_id: initialToStatusId,
      })
    } catch {
      // onError で transitionError を表示するため、ここでは握りつぶす
    } finally {
      setAddingTransition(false)
    }
  }

  const deleteTransitionStrap = async (transitionId: number) => {
    setTransitionBusyId(transitionId)
    try {
      await deleteTransitionMutation.mutateAsync(transitionId)
      setTransitionDrafts((prev) => {
        const next = { ...prev }
        delete next[transitionId]
        return next
      })
    } finally {
      setTransitionBusyId(null)
    }
  }

  const changeTransitionStrap = async (
    transitionId: number,
    field: 'from_status_id' | 'to_status_id',
    value: string
  ) => {
    const base = transitions.find((t) => t.id === transitionId)
    if (!base) return
    const cur = getRowTransition(base)
    const nextFrom = field === 'from_status_id' ? value : cur.from_status_id
    const nextTo = field === 'to_status_id' ? value : cur.to_status_id
    setTransitionDrafts((prev) => ({
      ...prev,
      [transitionId]: { from_status_id: nextFrom, to_status_id: nextTo },
    }))
    const reason = transitionReasonForPersistedRow(transitionId, nextFrom, nextTo)
    if (reason) return

    setTransitionBusyId(transitionId)
    try {
      await updateTransitionMutation.mutateAsync({
        transitionId,
        data: { from_status_id: nextFrom, to_status_id: nextTo },
      })
      setTransitionDrafts((prev) => {
        const next = { ...prev }
        delete next[transitionId]
        return next
      })
    } catch {
      // onError で transitionError を表示するため、ここでは握りつぶす
    } finally {
      setTransitionBusyId(null)
    }
  }

  const removeInvalidTransitionStrap = (clientId: string) => {
    setInvalidTransitionStraps((prev) => prev.filter((s) => s.clientId !== clientId))
  }

  const changeInvalidTransitionStrap = async (
    clientId: string,
    field: 'from_status_id' | 'to_status_id',
    value: string
  ) => {
    const current = invalidTransitionStraps.find((s) => s.clientId === clientId)
    if (!current) return
    const nextFrom = field === 'from_status_id' ? value : current.from_status_id
    const nextTo = field === 'to_status_id' ? value : current.to_status_id
    setInvalidTransitionStraps((prev) =>
      prev.map((s) => (s.clientId === clientId ? { ...s, from_status_id: nextFrom, to_status_id: nextTo } : s))
    )
    const reason = transitionReasonForNewPair(nextFrom, nextTo, clientId)
    if (reason) return
    try {
      await createTransitionMutation.mutateAsync({ from_status_id: nextFrom, to_status_id: nextTo })
      setInvalidTransitionStraps((prev) => prev.filter((s) => s.clientId !== clientId))
    } catch {
      // onError で transitionError を表示するため、ここでは握りつぶす
    }
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
  const atStatusDeleteFloor = statuses.length <= 2
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
          <h2 className="text-lg font-semibold text-gray-900">ステータス遷移設定</h2>
          {!orgMismatch && (
            <button
              type="button"
              onClick={() => void addTransitionStrap()}
              disabled={addingTransition || statusesByOrder.length < 2}
              title={
                statusesByOrder.length < 2
                  ? 'ステータスが2つ以上あるときのみ遷移を追加できます'
                  : undefined
              }
              className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
            >
              <Plus className="w-4 h-4" />
              ストラップを追加
            </button>
          )}
        </div>
        {orgMismatch && (
          <p className="text-sm text-gray-500">
            選択中の組織に属するワークフローのみ、遷移設定を表示・編集できます。
          </p>
        )}
        {!orgMismatch && (statusesLoading || transitionsLoading) && (
          <p className="text-sm text-gray-500">遷移設定を読み込み中...</p>
        )}
        {!orgMismatch && !statusesLoading && !transitionsLoading && statusesByOrder.length < 2 && (
          <p className="text-sm text-gray-500">
            ステータスが2つ以上あるときのみ遷移を追加できます。まず「ステータスを追加」でステータスを用意してください。
          </p>
        )}
        {!orgMismatch &&
          !transitionsLoading &&
          transitions.length === 0 &&
          statusesByOrder.length >= 2 && (
          <p className="text-sm text-gray-500">遷移はまだありません。「ストラップを追加」から作成できます。</p>
        )}
        {!orgMismatch && !transitionsLoading && transitions.length > 0 && (
          <div className="space-y-2">
            <SortableDndProvider
              items={transitions}
              itemId={(t) => String(t.id)}
              onReorder={(ids) =>
                reorderTransitionsMutation.mutate(ids.map((x) => Number(x)))
              }
              disabled={transitionStrapPending || reorderTransitionsMutation.isPending}
            >
              <SortableList
                items={transitions}
                itemId={(t) => String(t.id)}
                onReorder={() => {}}
                renderItem={(t, props) => {
                  const row = getRowTransition(t)
                  const busy = transitionBusyId === t.id
                  return (
                    <div
                      key={t.id}
                      ref={props.setNodeRef}
                      style={props.style}
                      className={`flex flex-wrap items-center gap-2 rounded-lg border px-3 py-2 ${row.reason ? 'border-amber-300 bg-amber-50' : 'border-gray-200'}`}
                    >
                      <DragHandle handleProps={props.handleProps} />
                      <select
                        value={row.from_status_id}
                        onChange={(e) => void changeTransitionStrap(t.id, 'from_status_id', e.target.value)}
                        disabled={busy || transitionStrapPending}
                        className="min-w-44 border rounded-lg px-2 py-1.5 text-sm"
                      >
                        {statusesByOrder.map((s) => (
                          <option key={s.id} value={s.id}>
                            {s.display_order}. {s.name}
                          </option>
                        ))}
                      </select>
                      <span className="text-gray-500">→</span>
                      <select
                        value={row.to_status_id}
                        onChange={(e) => void changeTransitionStrap(t.id, 'to_status_id', e.target.value)}
                        disabled={busy || transitionStrapPending}
                        className="min-w-44 border rounded-lg px-2 py-1.5 text-sm"
                      >
                        {statusesByOrder.map((s) => (
                          <option key={s.id} value={s.id}>
                            {s.display_order}. {s.name}
                          </option>
                        ))}
                      </select>
                      {row.reason === 'same' && (
                        <span className="text-xs text-amber-700 border border-amber-300 rounded px-2 py-0.5">
                          無効（遷移前後が同一）
                        </span>
                      )}
                      {row.reason === 'duplicate' && (
                        <span className="text-xs text-amber-700 border border-amber-300 rounded px-2 py-0.5">
                          無効（重複）
                        </span>
                      )}
                      <button
                        type="button"
                        onClick={() => void deleteTransitionStrap(t.id)}
                        disabled={busy || transitionStrapPending}
                        className="ml-auto p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded disabled:opacity-50"
                        title="削除"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  )
                }}
              />
            </SortableDndProvider>
            {invalidTransitionStraps.map((s) => {
              const ir = invalidStrapReason(s)
              return (
                <div
                  key={s.clientId}
                  className="flex flex-wrap items-center gap-2 rounded-lg border px-3 py-2 border-amber-300 bg-amber-50"
                >
                  <select
                    value={s.from_status_id}
                    onChange={(e) => void changeInvalidTransitionStrap(s.clientId, 'from_status_id', e.target.value)}
                    disabled={transitionStrapPending}
                    className="min-w-44 border rounded-lg px-2 py-1.5 text-sm"
                  >
                    {statusesByOrder.map((st) => (
                      <option key={st.id} value={st.id}>
                        {st.display_order}. {st.name}
                      </option>
                    ))}
                  </select>
                  <span className="text-gray-500">→</span>
                  <select
                    value={s.to_status_id}
                    onChange={(e) => void changeInvalidTransitionStrap(s.clientId, 'to_status_id', e.target.value)}
                    disabled={transitionStrapPending}
                    className="min-w-44 border rounded-lg px-2 py-1.5 text-sm"
                  >
                    {statusesByOrder.map((st) => (
                      <option key={st.id} value={st.id}>
                        {st.display_order}. {st.name}
                      </option>
                    ))}
                  </select>
                  {ir === 'same' && (
                    <span className="text-xs text-amber-700 border border-amber-300 rounded px-2 py-0.5">
                      無効（遷移前後が同一）
                    </span>
                  )}
                  {ir === 'duplicate' && (
                    <span className="text-xs text-amber-700 border border-amber-300 rounded px-2 py-0.5">
                      無効（重複）
                    </span>
                  )}
                  <button
                    type="button"
                    onClick={() => removeInvalidTransitionStrap(s.clientId)}
                    disabled={transitionStrapPending}
                    className="ml-auto p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded disabled:opacity-50"
                    title="削除"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              )
            })}
          </div>
        )}
        {!orgMismatch &&
          !transitionsLoading &&
          transitions.length === 0 &&
          invalidTransitionStraps.length > 0 && (
            <div className="space-y-2">
              {invalidTransitionStraps.map((s) => {
                const ir = invalidStrapReason(s)
                return (
                  <div
                    key={s.clientId}
                    className="flex flex-wrap items-center gap-2 rounded-lg border px-3 py-2 border-amber-300 bg-amber-50"
                  >
                    <select
                      value={s.from_status_id}
                      onChange={(e) => void changeInvalidTransitionStrap(s.clientId, 'from_status_id', e.target.value)}
                      disabled={transitionStrapPending}
                      className="min-w-44 border rounded-lg px-2 py-1.5 text-sm"
                    >
                      {statusesByOrder.map((st) => (
                        <option key={st.id} value={st.id}>
                          {st.display_order}. {st.name}
                        </option>
                      ))}
                    </select>
                    <span className="text-gray-500">→</span>
                    <select
                      value={s.to_status_id}
                      onChange={(e) => void changeInvalidTransitionStrap(s.clientId, 'to_status_id', e.target.value)}
                      disabled={transitionStrapPending}
                      className="min-w-44 border rounded-lg px-2 py-1.5 text-sm"
                    >
                      {statusesByOrder.map((st) => (
                        <option key={st.id} value={st.id}>
                          {st.display_order}. {st.name}
                        </option>
                      ))}
                    </select>
                    {ir === 'same' && (
                      <span className="text-xs text-amber-700 border border-amber-300 rounded px-2 py-0.5">
                        無効（遷移前後が同一）
                      </span>
                    )}
                    {ir === 'duplicate' && (
                      <span className="text-xs text-amber-700 border border-amber-300 rounded px-2 py-0.5">
                        無効（重複）
                      </span>
                    )}
                    <button
                      type="button"
                      onClick={() => removeInvalidTransitionStrap(s.clientId)}
                      disabled={transitionStrapPending}
                      className="ml-auto p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded disabled:opacity-50"
                      title="削除"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                )
              })}
            </div>
          )}
        {transitionError && <p className="mt-4 text-sm text-red-600">{transitionError}</p>}
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
            <SortableDndProvider
              items={statusesByOrder}
              itemId={(s) => s.id}
              onReorder={(ids) => reorderStatusesMutation.mutate(ids)}
              disabled={
                reorderStatusesMutation.isPending ||
                deleteStatusMutation.isPending ||
                statusesLoading
              }
            >
              <table className="min-w-full text-sm">
                <thead className="bg-gray-50 text-left text-gray-600">
                  <tr>
                    <th className="w-10 px-2 py-2" aria-hidden />
                    <th className="px-3 py-2 font-medium">順</th>
                    <th className="px-3 py-2 font-medium">名前</th>
                    <th className="px-3 py-2 font-medium">色</th>
                    <th className="px-3 py-2 font-medium w-28">操作</th>
                  </tr>
                </thead>
                <SortableTbody
                  items={statusesByOrder}
                  itemId={(s) => s.id}
                  disabled={
                    reorderStatusesMutation.isPending ||
                    deleteStatusMutation.isPending ||
                    statusesLoading
                  }
                  tbodyClassName="divide-y divide-gray-100"
                  renderItem={(s, props) => (
                    <tr
                      ref={props.setNodeRef}
                      style={props.style}
                      key={s.id}
                      className="border-t border-gray-100 hover:bg-gray-50/80"
                    >
                      <td className="px-2 py-2 align-middle">
                        <DragHandle handleProps={props.handleProps} />
                      </td>
                      <td className="px-3 py-2 text-gray-700">{s.display_order}</td>
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
                          <div className="flex items-center gap-1">
                            <button
                              type="button"
                              onClick={() => openEditStatusDialog(s)}
                              className="p-1.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded"
                              title="編集"
                            >
                              <Pencil className="w-4 h-4" />
                            </button>
                            <button
                              type="button"
                              disabled={
                                deleteStatusMutation.isPending ||
                                atStatusDeleteFloor ||
                                statusReferencedInTransition(s.id)
                              }
                              onClick={() => {
                                if (statusReferencedInTransition(s.id)) {
                                  window.alert(
                                    'このステータスは許可遷移で使用されているため削除できません'
                                  )
                                  return
                                }
                                if (confirm(`「${s.name}」を本当に削除しますか？`)) {
                                  deleteStatusMutation.mutate(s.id)
                                }
                              }}
                              className="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded disabled:opacity-50"
                              title={
                                atStatusDeleteFloor
                                  ? 'ステータスはワークフロー内で最低2つ必要なため削除できません'
                                  : statusReferencedInTransition(s.id)
                                    ? 'このステータスは許可遷移で使用されているため削除できません'
                                    : '削除'
                              }
                            >
                              <Trash2 className="w-4 h-4" />
                            </button>
                          </div>
                        )}
                      </td>
                    </tr>
                  )}
                />
              </table>
            </SortableDndProvider>
          </div>
        )}

        <p className="mt-4 text-xs text-gray-500">
          Issue のステータス変更は、上記の「ステータス遷移設定」で許可した遷移に沿う必要があります。ステータスはワークフロー内で最低2つ必要です。
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
