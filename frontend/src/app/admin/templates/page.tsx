'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Pencil, Trash2, X, Check, GitBranch } from 'lucide-react'
import { getTemplates, getProjects, createTemplate, updateTemplate, deleteTemplate } from '@/lib/api'
import type { IssueTemplate, Project, Workflow } from '@/types'
import { PRIORITY_LABELS, PRIORITY_COLORS, type Priority } from '@/types'
import { SortableList, DragHandle } from '@/components/SortableList'

const API = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'

async function fetchWorkflows(): Promise<Workflow[]> {
  const res = await fetch(`${API}/workflows`)
  const json = await res.json()
  return json.data ?? []
}

const emptyForm = {
  project_id: '',
  name: '',
  description: '',
  body: '',
  default_priority: 'medium' as Priority,
  workflow_id: '' as string,
}

export default function TemplatesPage() {
  const queryClient = useQueryClient()
  const { data: templates = [], isLoading } = useQuery({ queryKey: ['templates'], queryFn: getTemplates })
  const { data: projects = [] } = useQuery<Project[]>({ queryKey: ['projects'], queryFn: () => getProjects() })
  const { data: workflows = [] } = useQuery({ queryKey: ['workflows'], queryFn: fetchWorkflows })

  const [showForm, setShowForm] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [form, setForm] = useState(emptyForm)
  const [error, setError] = useState('')

  const createMutation = useMutation({
    mutationFn: (data: typeof emptyForm) =>
      createTemplate({
        project_id: data.project_id,
        name: data.name,
        description: data.description,
        body: data.body,
        default_priority: data.default_priority,
        workflow_id: data.workflow_id ? Number(data.workflow_id) : undefined,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['templates'] })
      setShowForm(false)
      setForm(emptyForm)
      setError('')
    },
    onError: () => setError('作成に失敗しました'),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: typeof emptyForm }) =>
      updateTemplate(id, {
        name: data.name,
        description: data.description,
        body: data.body,
        default_priority: data.default_priority,
        workflow_id: data.workflow_id ? Number(data.workflow_id) : null,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['templates'] })
      setEditingId(null)
      setForm(emptyForm)
      setError('')
    },
    onError: () => setError('更新に失敗しました'),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => deleteTemplate(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['templates'] }),
  })

  const [reorderPending, setReorderPending] = useState<string | null>(null)
  const reorderMutation = useMutation({
    mutationFn: async ({ projectId, ids }: { projectId: string; ids: number[] }) => {
      const res = await fetch(`${API}/projects/${projectId}/templates/reorder`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ids }),
      })
      if (!res.ok) throw new Error('並び替えに失敗しました')
    },
    onMutate: ({ projectId }) => setReorderPending(projectId),
    onSettled: () => setReorderPending(null),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['templates'] }),
  })

  const startEdit = (tmpl: IssueTemplate) => {
    setEditingId(tmpl.id)
    setForm({
      project_id: tmpl.project_id,
      name: tmpl.name,
      description: tmpl.description,
      body: tmpl.body,
      default_priority: tmpl.default_priority,
      workflow_id: tmpl.workflow_id ? String(tmpl.workflow_id) : '',
    })
    setShowForm(false)
    setError('')
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!form.name.trim()) { setError('名前は必須です'); return }
    if (editingId !== null) {
      updateMutation.mutate({ id: editingId, data: form })
    } else {
      if (!form.project_id) { setError('プロジェクトを選択してください'); return }
      createMutation.mutate(form)
    }
  }

  // ワークフローは組織に属さないため全件表示
  const filteredWorkflows = workflows

  const getProjectName = (id: string) => {
    const p = projects.find((p: Project) => p.id === id)
    return p ? `${p.key} - ${p.name}` : id
  }

  // プロジェクトごとにグループ化
  const grouped = templates.reduce<Record<string, IssueTemplate[]>>((acc, t) => {
    if (!acc[t.project_id]) acc[t.project_id] = []
    acc[t.project_id].push(t)
    return acc
  }, {})

  return (
    <div className="max-w-3xl">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Issueテンプレート管理</h1>
          <p className="text-sm text-gray-500 mt-0.5">Issue作成時に選択できるテンプレートを定義します</p>
        </div>
        {!showForm && editingId === null && (
          <button
            onClick={() => {
              setShowForm(true)
              setForm({ ...emptyForm, project_id: projects[0]?.id ?? '' })
              setError('')
            }}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors"
          >
            <Plus className="w-4 h-4" />
            テンプレートを追加
          </button>
        )}
      </div>

      {/* フォーム */}
      {(showForm || editingId !== null) && (
        <div className="bg-white border border-gray-200 rounded-xl p-5 mb-6 shadow-sm">
          <h2 className="text-sm font-semibold text-gray-700 mb-4">
            {editingId !== null ? 'テンプレートを編集' : '新しいテンプレート'}
          </h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            {editingId === null && (
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">プロジェクト *</label>
                <select
                  value={form.project_id}
                  onChange={(e) => setForm({ ...form, project_id: e.target.value, workflow_id: '' })}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">プロジェクトを選択...</option>
                  {projects.map((p: Project) => (
                    <option key={p.id} value={p.id}>{p.key} - {p.name}</option>
                  ))}
                </select>
              </div>
            )}

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">テンプレート名 *</label>
                <input
                  type="text"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  placeholder="例: バグ報告"
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">デフォルト優先度</label>
                <select
                  value={form.default_priority}
                  onChange={(e) => setForm({ ...form, default_priority: e.target.value as Priority })}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  {Object.entries(PRIORITY_LABELS).map(([v, l]) => (
                    <option key={v} value={v}>{l}</option>
                  ))}
                </select>
              </div>
            </div>

            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">説明</label>
              <input
                type="text"
                value={form.description}
                onChange={(e) => setForm({ ...form, description: e.target.value })}
                placeholder="例: バグを報告するときに使用"
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">本文テンプレート</label>
              <textarea
                value={form.body}
                onChange={(e) => setForm({ ...form, body: e.target.value })}
                placeholder={'例:\n## 概要\n\n## 再現手順\n\n## 期待結果'}
                rows={6}
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none font-mono"
              />
            </div>

            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">
                承認ワークフロー
                <span className="ml-1 font-normal text-gray-400">（任意）</span>
              </label>
              <select
                value={form.workflow_id}
                onChange={(e) => setForm({ ...form, workflow_id: e.target.value })}
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">承認フローなし</option>
                {filteredWorkflows.map((w: Workflow) => (
                  <option key={w.id} value={w.id}>{w.name}</option>
                ))}
              </select>
              {form.project_id && filteredWorkflows.length === 0 && (
                <p className="text-xs text-gray-400 mt-1">このプロジェクトにはワークフローがありません</p>
              )}
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

      {/* テンプレート一覧 */}
      {isLoading ? (
        <div className="bg-white border border-gray-200 rounded-xl p-8 text-center text-gray-400 text-sm">読み込み中...</div>
      ) : templates.length === 0 ? (
        <div className="bg-white border border-gray-200 rounded-xl p-8 text-center text-gray-400 text-sm">
          テンプレートがまだありません。「テンプレートを追加」から作成してください。
        </div>
      ) : (
        <div className="space-y-4">
          {Object.entries(grouped).map(([projectId, tmpls]) => (
            <div key={projectId}>
              <p className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-2 px-1">
                {getProjectName(projectId)}
              </p>
              <div className="bg-white border border-gray-200 rounded-xl shadow-sm overflow-hidden">
                <SortableList
                  items={tmpls}
                  itemId={(t) => String(t.id)}
                  onReorder={(ids) => reorderMutation.mutate({ projectId, ids: ids.map(Number) })}
                  disabled={reorderPending === projectId}
                  renderItem={(tmpl, { handleProps, setNodeRef, style }) => (
                    <div ref={setNodeRef} style={style} className="flex items-start gap-3 px-4 py-3 hover:bg-gray-50 transition-colors border-t border-gray-100 first:border-t-0">
                      <DragHandle handleProps={handleProps} className="pt-0.5" />
                      <div className="flex-1 min-w-0 pt-0.5">
                        <div className="flex items-center gap-2 flex-wrap">
                          <span className="font-medium text-gray-900 text-sm">{tmpl.name}</span>
                          <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${PRIORITY_COLORS[tmpl.default_priority]}`}>
                            {PRIORITY_LABELS[tmpl.default_priority]}
                          </span>
                          {tmpl.workflow && (
                            <span className="flex items-center gap-1 text-xs text-blue-600 bg-blue-50 px-2 py-0.5 rounded-full">
                              <GitBranch className="w-3 h-3" />
                              {tmpl.workflow.name}
                            </span>
                          )}
                        </div>
                        {tmpl.description && (
                          <p className="text-xs text-gray-400 mt-0.5">{tmpl.description}</p>
                        )}
                        {tmpl.body && (
                          <p className="text-xs text-gray-300 mt-1 font-mono truncate">{tmpl.body.split('\n')[0]}</p>
                        )}
                      </div>
                      <div className="flex items-center gap-1 flex-shrink-0">
                        <button
                          onClick={() => startEdit(tmpl)}
                          className="p-1.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                          title="編集"
                        >
                          <Pencil className="w-3.5 h-3.5" />
                        </button>
                        <button
                          onClick={() => {
                            if (confirm(`「${tmpl.name}」を削除しますか？`)) {
                              deleteMutation.mutate(tmpl.id)
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
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
