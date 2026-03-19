'use client'

import { useState, useCallback, useEffect } from 'react'
import { use } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ChevronLeft, Check, Plus, Trash2 } from 'lucide-react'
import type { WorkflowStep, Status, Role, User, ApprovalObject } from '@/types'
import { getWorkflowStep, updateWorkflowStep } from '@/lib/api'
import { useAuth } from '@/context/AuthContext'

const API = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'

async function fetchOrgStatuses(orgId: string): Promise<Status[]> {
  const res = await fetch(`${API}/organizations/${orgId}/statuses?type=issue`)
  const json = await res.json()
  const data: Status[] = json.data ?? []
  return data.filter((s) =>
    !s.project_id && (s.organization_id || s.status_key === 'sts_start' || s.status_key === 'sts_goal')
  )
}

async function fetchRoles(orgId?: string): Promise<Role[]> {
  const url = orgId ? `${API}/roles?org_id=${orgId}` : `${API}/roles`
  const res = await fetch(url)
  const json = await res.json()
  return json.data ?? []
}

async function fetchUsers(orgId: string): Promise<User[]> {
  const res = await fetch(`${API}/admin/users?org_id=${orgId}`)
  const json = await res.json()
  return json.data ?? []
}

const emptyApprovalObject = (): ApprovalObject & { _new?: boolean } => ({
  id: 0,
  workflow_step_id: 0,
  order: 0,
  type: 'role',
  points: 1,
  exclude_reporter: false,
  exclude_assignee: false,
  _new: true,
})

export default function StepEditPage({
  params,
}: {
  params: Promise<{ id: string; stepId: string }>
}) {
  const { id, stepId } = use(params)
  const router = useRouter()
  const queryClient = useQueryClient()
  const { currentOrg } = useAuth()

  const { data: step, isLoading } = useQuery({
    queryKey: ['workflow-step', id, stepId],
    queryFn: () => getWorkflowStep(id, stepId),
    enabled: !!id && !!stepId,
  })

  const { data: statuses = [] } = useQuery({
    queryKey: ['org-statuses', currentOrg?.id],
    queryFn: () => fetchOrgStatuses(currentOrg!.id),
    enabled: !!currentOrg?.id,
  })

  const { data: roles = [] } = useQuery({
    queryKey: ['roles', currentOrg?.id],
    queryFn: () => fetchRoles(currentOrg?.id),
    enabled: !!currentOrg?.id,
  })

  const { data: users = [] } = useQuery({
    queryKey: ['admin-users', currentOrg?.id],
    queryFn: () => fetchUsers(currentOrg!.id),
    enabled: !!currentOrg?.id,
  })

  const [form, setForm] = useState({
    status_id: '',
    next_status_id: '',
    description: '',
    threshold: 10,
  })
  const [approvalObjects, setApprovalObjects] = useState<(ApprovalObject & { _new?: boolean })[]>([])
  const [error, setError] = useState('')
  const [initialized, setInitialized] = useState(false)

  const updateFormFromStep = useCallback((s: WorkflowStep) => {
    setForm({
      status_id: s.status_id ?? '',
      next_status_id: s.next_status_id ?? '',
      description: s.description ?? '',
      threshold: s.threshold ?? 10,
    })
    setApprovalObjects((s.approval_objects ?? []).map((ao) => ({ ...ao, _new: false })))
  }, [])

  useEffect(() => {
    if (step && !initialized) {
      updateFormFromStep(step)
      setInitialized(true)
    }
  }, [step, initialized, updateFormFromStep])

  type StepFormData = typeof form
  type ApprovalObjectsData = typeof approvalObjects

  const updateMutation = useMutation({
    mutationFn: async ({
      formData,
      approvalObjectsData,
    }: {
      formData: StepFormData
      approvalObjectsData: ApprovalObjectsData
    }) => {
      if (!formData.status_id) throw new Error('ステータスは必須です')
      if (formData.threshold < 1) throw new Error('閾値は1以上で指定してください')
      const payload = {
        status_id: formData.status_id,
        next_status_id: formData.next_status_id || undefined,
        description: formData.description,
        threshold: formData.threshold,
        approval_objects: approvalObjectsData
          .filter((ao) => (ao.type === 'role' && ao.role_id) || (ao.type === 'user' && ao.user_id))
          .map((ao) => ({
            type: ao.type,
            role_id: ao.role_id != null && ao.role_id !== '' ? Number(ao.role_id) : undefined,
            role_operator: ao.role_operator ?? 'gte',
            user_id: ao.user_id || undefined,
            points: Math.max(1, Number(ao.points) || 1),
            exclude_reporter: ao.exclude_reporter,
            exclude_assignee: ao.exclude_assignee,
          })),
      }
      return updateWorkflowStep(id, stepId, payload)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflow', id] })
      queryClient.invalidateQueries({ queryKey: ['workflow-step', id, stepId] })
      setError('')
      router.push(`/admin/workflows/${id}`)
    },
    onError: (e: Error & { response?: { data?: { message?: string } } }) => {
      const msg = e.response?.data?.message ?? e.message
      setError(msg)
    },
  })

  const isGoal = !form.next_status_id

  if (isLoading || !step) {
    return (
      <div className="max-w-2xl">
        <div className="text-gray-400 text-sm">読み込み中...</div>
      </div>
    )
  }

  return (
    <div className="max-w-2xl">
      <div className="flex items-center gap-3 mb-6">
          <Link
            href={`/admin/workflows/${id}`}
            className="flex items-center gap-1 text-sm text-gray-500 hover:text-gray-700 transition-colors"
          >
            <ChevronLeft className="w-4 h-4" />
            ワークフローに戻る
          </Link>
          <span className="text-gray-300">/</span>
          <h1 className="text-xl font-bold text-gray-900">ステップを編集: {step.status?.name ?? step.status_id}</h1>
        </div>

        <form
          onSubmit={(e) => {
            e.preventDefault()
            updateMutation.mutate({ formData: form, approvalObjectsData: approvalObjects })
          }}
          className="space-y-6"
        >
          {error && (
            <p className="text-sm text-red-500 bg-red-50 px-3 py-2 rounded-lg">{error}</p>
          )}

          <div className="bg-white border border-gray-200 rounded-xl p-6 shadow-sm space-y-4">
            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">ステータス *</label>
              <select
                value={form.status_id}
                onChange={(e) => setForm((prev) => ({ ...prev, status_id: e.target.value }))}
                disabled={step.status?.status_key === 'sts_start' || step.status?.status_key === 'sts_goal'}
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:cursor-not-allowed"
              >
                <option value="">選択</option>
                {statuses.map((s) => (
                  <option key={s.id} value={s.id}>{s.name}</option>
                ))}
              </select>
              {(step.status?.status_key === 'sts_start' || step.status?.status_key === 'sts_goal') && (
                <p className="text-xs text-gray-500 mt-1">システムステータスは変更できません</p>
              )}
            </div>

            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">承認後ステータス</label>
              <select
                value={form.next_status_id}
                onChange={(e) => setForm((prev) => ({ ...prev, next_status_id: e.target.value }))}
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">（なし・ゴール）</option>
                {statuses.map((s) => (
                  <option key={s.id} value={s.id}>{s.name}</option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">ステップの説明</label>
              <textarea
                value={form.description}
                onChange={(e) => setForm((prev) => ({ ...prev, description: e.target.value }))}
                placeholder="ステップの説明を入力"
                rows={2}
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            {!isGoal && (
              <>
                <div>
                  <label className="block text-xs font-medium text-gray-600 mb-1">閾値（点数合計がこれ以上で遷移）</label>
                  <input
                    type="number"
                    min={1}
                    max={99999}
                    value={form.threshold}
                    onChange={(e) =>
                      setForm((prev) => ({ ...prev, threshold: parseInt(e.target.value) || 1 }))
                    }
                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>

                <div>
                  <div className="flex items-center justify-between mb-2">
                    <label className="block text-xs font-medium text-gray-600">承認オブジェクト</label>
                    <button
                      type="button"
                      onClick={() =>
                        setApprovalObjects((prev) => [...prev, emptyApprovalObject()])
                      }
                      className="flex items-center gap-1 text-xs text-blue-600 hover:text-blue-700"
                    >
                      <Plus className="w-3.5 h-3.5" />
                      追加
                    </button>
                  </div>
                  <div className="space-y-3">
                    {approvalObjects.map((ao, idx) => (
                      <div
                        key={ao._new ? `new-${idx}` : ao.id}
                        className="p-3 border border-gray-200 rounded-lg bg-gray-50 space-y-2"
                      >
                        <div className="flex justify-between items-start">
                          <span className="text-xs font-medium text-gray-500">#{idx + 1}</span>
                          <button
                            type="button"
                            onClick={() =>
                              setApprovalObjects((prev) => prev.filter((_, i) => i !== idx))
                            }
                            className="p-1 text-gray-400 hover:text-red-600"
                          >
                            <Trash2 className="w-3.5 h-3.5" />
                          </button>
                        </div>
                        <div className="grid grid-cols-2 gap-2">
                          <div>
                            <label className="block text-xs text-gray-500 mb-0.5">種類</label>
                            <select
                              value={ao.type}
                              onChange={(e) =>
                                setApprovalObjects((prev) => {
                                  const next = [...prev]
                                  next[idx] = {
                                    ...next[idx],
                                    type: e.target.value as 'role' | 'user',
                                    role_id: undefined,
                                    user_id: undefined,
                                  }
                                  return next
                                })
                              }
                              className="w-full border border-gray-300 rounded px-2 py-1 text-xs"
                            >
                              <option value="role">役職</option>
                              <option value="user">人</option>
                            </select>
                          </div>
                          <div>
                            <label className="block text-xs text-gray-500 mb-0.5">点数</label>
                            <input
                              type="number"
                              min={1}
                              value={ao.points}
                              onChange={(e) =>
                                setApprovalObjects((prev) => {
                                  const next = [...prev]
                                  next[idx] = { ...next[idx], points: parseInt(e.target.value) || 1 }
                                  return next
                                })
                              }
                              className="w-full border border-gray-300 rounded px-2 py-1 text-xs"
                            />
                          </div>
                        </div>
                        {ao.type === 'role' && (
                          <div className="grid grid-cols-2 gap-2">
                            <div>
                              <label className="block text-xs text-gray-500 mb-0.5">役職</label>
                              <select
                                value={ao.role_id ?? ''}
                                onChange={(e) =>
                                  setApprovalObjects((prev) => {
                                    const next = [...prev]
                                    next[idx] = {
                                      ...next[idx],
                                      role_id: e.target.value ? Number(e.target.value) : undefined,
                                    }
                                    return next
                                  })
                                }
                                className="w-full border border-gray-300 rounded px-2 py-1 text-xs"
                              >
                                <option value="">選択</option>
                                {roles.map((r) => (
                                  <option key={r.id} value={r.id}>
                                    {r.name} (Lv{r.level})
                                  </option>
                                ))}
                              </select>
                            </div>
                            <div>
                              <label className="block text-xs text-gray-500 mb-0.5">比較</label>
                              <select
                                value={ao.role_operator ?? 'gte'}
                                onChange={(e) =>
                                  setApprovalObjects((prev) => {
                                    const next = [...prev]
                                    next[idx] = {
                                      ...next[idx],
                                      role_operator: e.target.value as 'eq' | 'gte',
                                    }
                                    return next
                                  })
                                }
                                className="w-full border border-gray-300 rounded px-2 py-1 text-xs"
                              >
                                <option value="eq">イコール</option>
                                <option value="gte">以上</option>
                              </select>
                            </div>
                          </div>
                        )}
                        {ao.type === 'user' && (
                          <div>
                            <label className="block text-xs text-gray-500 mb-0.5">ユーザー</label>
                            <select
                              value={ao.user_id ?? ''}
                              onChange={(e) =>
                                setApprovalObjects((prev) => {
                                  const next = [...prev]
                                  next[idx] = {
                                    ...next[idx],
                                    user_id: e.target.value || undefined,
                                  }
                                  return next
                                })
                              }
                              className="w-full border border-gray-300 rounded px-2 py-1 text-xs"
                            >
                              <option value="">選択</option>
                              {users.map((u) => (
                                <option key={u.id} value={u.id}>
                                  {u.name} ({u.email})
                                </option>
                              ))}
                            </select>
                          </div>
                        )}
                        <div className="flex gap-4 pt-1">
                          <label className="flex items-center gap-1 text-xs cursor-pointer">
                            <input
                              type="checkbox"
                              checked={ao.exclude_reporter}
                              onChange={(e) =>
                                setApprovalObjects((prev) => {
                                  const next = [...prev]
                                  next[idx] = { ...next[idx], exclude_reporter: e.target.checked }
                                  return next
                                })
                              }
                              className="rounded"
                            />
                            起票者を除外
                          </label>
                          <label className="flex items-center gap-1 text-xs cursor-pointer">
                            <input
                              type="checkbox"
                              checked={ao.exclude_assignee}
                              onChange={(e) =>
                                setApprovalObjects((prev) => {
                                  const next = [...prev]
                                  next[idx] = { ...next[idx], exclude_assignee: e.target.checked }
                                  return next
                                })
                              }
                              className="rounded"
                            />
                            担当者を除外
                          </label>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </>
            )}
          </div>

          <div className="flex gap-2">
            <button
              type="submit"
              disabled={updateMutation.isPending}
              className="flex items-center gap-1 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50"
            >
              <Check className="w-4 h-4" />
              保存
            </button>
            <Link
              href={`/admin/workflows/${id}`}
              className="px-4 py-2 border border-gray-300 text-gray-600 rounded-lg text-sm hover:bg-gray-50"
            >
              キャンセル
            </Link>
          </div>
        </form>
    </div>
  )
}
