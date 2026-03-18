'use client'

import { useState } from 'react'
import { use } from 'react'
import Link from 'next/link'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2, X, Check, ArrowUp, ArrowDown, ChevronLeft } from 'lucide-react'
import type { Workflow, Status } from '@/types'

const API = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'

async function fetchWorkflow(id: string): Promise<Workflow> {
  const res = await fetch(`${API}/workflows/${id}`)
  const json = await res.json()
  return json.data
}

async function fetchOrgStatuses(orgId: string): Promise<Status[]> {
  const res = await fetch(`${API}/organizations/${orgId}/statuses`)
  const json = await res.json()
  return json.data ?? []
}

const emptyStep = { name: '', required_level: 1, status_id: '' }

export default function WorkflowDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params)
  const queryClient = useQueryClient()

  const { data: workflow, isLoading } = useQuery({
    queryKey: ['workflow', id],
    queryFn: () => fetchWorkflow(id),
  })

  const { data: statuses = [] } = useQuery({
    queryKey: ['org-statuses', workflow?.organization_id],
    queryFn: () => fetchOrgStatuses(workflow!.organization_id),
    enabled: !!workflow?.organization_id,
  })

  const [showAddForm, setShowAddForm] = useState(false)
  const [editingStepId, setEditingStepId] = useState<number | null>(null)
  const [stepForm, setStepForm] = useState(emptyStep)
  const [error, setError] = useState('')

  const validateStepForm = (data: typeof emptyStep) => {
    if (data.required_level < 0 || data.required_level > 9999) {
      throw new Error('必要レベルは0～9999の範囲で指定してください')
    }
  }

  const addStepMutation = useMutation({
    mutationFn: async (data: typeof emptyStep) => {
      validateStepForm(data)
      const body: Record<string, unknown> = {
        name: data.name,
        required_level: data.required_level,
      }
      if (data.status_id) body.status_id = data.status_id
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

  const updateStepMutation = useMutation({
    mutationFn: async ({ stepId, data, order }: { stepId: number; data: typeof emptyStep; order: number }) => {
      validateStepForm(data)
      if (order < 0 || order > 9999) throw new Error('表示順は0～9999の範囲で指定してください')
      const body: Record<string, unknown> = {
        name: data.name,
        required_level: data.required_level,
        order,
      }
      if (data.status_id) body.status_id = data.status_id
      const res = await fetch(`${API}/workflows/${id}/steps/${stepId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
      if (!res.ok) {
        const json = await res.json().catch(() => ({}))
        throw new Error(json.message ?? '更新に失敗しました')
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflow', id] })
      setEditingStepId(null)
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

  const moveStepMutation = useMutation({
    mutationFn: async ({ stepId, name, requiredLevel, statusId, newOrder }: {
      stepId: number; name: string; requiredLevel: number; statusId?: string; newOrder: number
    }) => {
      const body: Record<string, unknown> = { name, required_level: requiredLevel, order: newOrder }
      if (statusId) body.status_id = statusId
      await fetch(`${API}/workflows/${id}/steps/${stepId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['workflow', id] }),
  })

  const handleMoveStep = async (index: number, direction: 'up' | 'down') => {
    const steps = workflow?.steps ?? []
    const step = steps[index]
    const swapIndex = direction === 'up' ? index - 1 : index + 1
    const swapStep = steps[swapIndex]
    if (!step || !swapStep) return

    await moveStepMutation.mutateAsync({
      stepId: step.id, name: step.name, requiredLevel: step.required_level,
      statusId: step.status_id, newOrder: swapStep.order,
    })
    await moveStepMutation.mutateAsync({
      stepId: swapStep.id, name: swapStep.name, requiredLevel: swapStep.required_level,
      statusId: swapStep.status_id, newOrder: step.order,
    })
  }

  const steps = workflow?.steps ?? []

  if (isLoading) {
    return <div className="text-gray-400 text-sm">読み込み中...</div>
  }
  if (!workflow) {
    return <div className="text-red-500 text-sm">ワークフローが見つかりません</div>
  }

  const StepForm = ({ onSubmit, loading }: { onSubmit: (data: typeof emptyStep) => void; loading: boolean }) => (
    <form
      onSubmit={(e) => { e.preventDefault(); onSubmit(stepForm) }}
      className="grid grid-cols-12 gap-3 items-end"
    >
      <div className="col-span-4">
        <label className="block text-xs font-medium text-gray-500 mb-1">ステップ名 *</label>
        <input
          type="text"
          value={stepForm.name}
          onChange={(e) => setStepForm({ ...stepForm, name: e.target.value })}
          placeholder="例: 上司承認"
          className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>
      <div className="col-span-3">
        <label className="block text-xs font-medium text-gray-500 mb-1">
          必要Lv（0～9999、以上） *
        </label>
        <input
          type="number"
          min={0}
          max={9999}
          value={stepForm.required_level}
          onChange={(e) => setStepForm({ ...stepForm, required_level: parseInt(e.target.value) || 0 })}
          placeholder="1"
          className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>
      <div className="col-span-3">
        <label className="block text-xs font-medium text-gray-500 mb-1">承認後ステータス</label>
        <select
          value={stepForm.status_id}
          onChange={(e) => setStepForm({ ...stepForm, status_id: e.target.value })}
          className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option value="">（変更なし）</option>
          {statuses.map((s) => (
            <option key={s.id} value={s.id}>{s.name}</option>
          ))}
        </select>
      </div>
      <div className="col-span-2 flex gap-1.5">
        <button
          type="submit"
          disabled={loading}
          className="flex items-center gap-1 px-3 py-2 bg-blue-600 text-white rounded-lg text-xs font-medium hover:bg-blue-700 disabled:opacity-50 transition-colors"
        >
          <Check className="w-3.5 h-3.5" />
          保存
        </button>
        <button
          type="button"
          onClick={() => { setShowAddForm(false); setEditingStepId(null); setStepForm(emptyStep); setError('') }}
          className="flex items-center gap-1 px-2.5 py-2 border border-gray-300 text-gray-600 rounded-lg text-xs hover:bg-gray-50 transition-colors"
        >
          <X className="w-3.5 h-3.5" />
        </button>
      </div>
    </form>
  )

  return (
    <div className="max-w-3xl">
      {/* ヘッダー */}
      <div className="flex items-center gap-3 mb-6">
        <Link
          href="/admin/workflows"
          className="flex items-center gap-1 text-sm text-gray-500 hover:text-gray-700 transition-colors"
        >
          <ChevronLeft className="w-4 h-4" />
          ワークフロー一覧
        </Link>
        <span className="text-gray-300">/</span>
        <h1 className="text-xl font-bold text-gray-900">{workflow.name}</h1>
      </div>

      {workflow.description && (
        <p className="text-sm text-gray-500 mb-6">{workflow.description}</p>
      )}

      {/* ステップ一覧 */}
      <div className="bg-white border border-gray-200 rounded-xl shadow-sm overflow-hidden mb-4">
        <div className="flex items-center justify-between px-4 py-3 border-b border-gray-100">
          <h2 className="text-sm font-semibold text-gray-700">承認ステップ</h2>
          <span className="text-xs text-gray-400">{steps.length} ステップ</span>
        </div>

        {steps.length === 0 ? (
          <div className="p-8 text-center text-gray-400 text-sm">
            ステップがまだありません。「ステップを追加」から作成してください。
          </div>
        ) : (
          <div className="divide-y divide-gray-100">
            {steps.map((step, idx) => (
              <div key={step.id}>
                {editingStepId === step.id ? (
                  <div className="px-4 py-3 bg-blue-50">
                    {error && <p className="text-xs text-red-500 mb-2">{error}</p>}
                    <StepForm
                      onSubmit={(data) => {
                        if (!data.name.trim()) { setError('ステップ名は必須です'); return }
                        updateStepMutation.mutate({ stepId: step.id, data, order: step.order })
                      }}
                      loading={updateStepMutation.isPending}
                    />
                  </div>
                ) : (
                  <div className="flex items-center gap-3 px-4 py-3 hover:bg-gray-50 transition-colors">
                    {/* 並び替えボタン */}
                    <div className="flex flex-col gap-0.5 flex-shrink-0">
                      <button
                        onClick={() => handleMoveStep(idx, 'up')}
                        disabled={idx === 0 || moveStepMutation.isPending}
                        className="p-0.5 text-gray-300 hover:text-gray-500 disabled:opacity-20 transition-colors"
                      >
                        <ArrowUp className="w-3 h-3" />
                      </button>
                      <button
                        onClick={() => handleMoveStep(idx, 'down')}
                        disabled={idx === steps.length - 1 || moveStepMutation.isPending}
                        className="p-0.5 text-gray-300 hover:text-gray-500 disabled:opacity-20 transition-colors"
                      >
                        <ArrowDown className="w-3 h-3" />
                      </button>
                    </div>

                    {/* ステップ番号 */}
                    <span className="w-6 h-6 rounded-full bg-blue-100 text-blue-700 text-xs font-bold flex items-center justify-center flex-shrink-0">
                      {step.order}
                    </span>

                    {/* ステップ情報 */}
                    <div className="flex-1 min-w-0">
                      <p className="font-medium text-gray-900 text-sm">{step.name}</p>
                      <p className="text-xs text-gray-400">
                        Level {step.required_level} 以上が承認可能
                        {step.status && (
                          <span className="ml-2">
                            → <span
                              className="inline-block w-2 h-2 rounded-full mr-0.5"
                              style={{ backgroundColor: step.status.color }}
                            />
                            {step.status.name}
                          </span>
                        )}
                      </p>
                    </div>

                    {/* 操作ボタン */}
                    <div className="flex items-center gap-1 flex-shrink-0">
                      <button
                        onClick={() => {
                          setEditingStepId(step.id)
                          setStepForm({
                            name: step.name,
                            required_level: step.required_level,
                            status_id: step.status_id ?? '',
                          })
                          setShowAddForm(false)
                          setError('')
                        }}
                        className="p-1.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                        title="編集"
                      >
                        <Plus className="w-3.5 h-3.5 rotate-45" />
                      </button>
                      <button
                        onClick={() => {
                          if (confirm(`「${step.name}」を削除しますか？`)) {
                            deleteStepMutation.mutate(step.id)
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
              </div>
            ))}
          </div>
        )}

        {/* ステップ追加フォーム */}
        {showAddForm && (
          <div className="px-4 py-3 border-t border-gray-100 bg-gray-50">
            {error && <p className="text-xs text-red-500 mb-2">{error}</p>}
            <StepForm
              onSubmit={(data) => {
                if (!data.name.trim()) { setError('ステップ名は必須です'); return }
                addStepMutation.mutate(data)
              }}
              loading={addStepMutation.isPending}
            />
          </div>
        )}
      </div>

      {!showAddForm && editingStepId === null && (
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
