'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getProject, getIssues, createIssue, getUsers } from '@/lib/api'
import { useState, use } from 'react'
import Link from 'next/link'
import { Plus, Circle } from 'lucide-react'
import { PRIORITY_LABELS, PRIORITY_COLORS, type Priority } from '@/types'
import type { Issue, Status } from '@/types'
import { formatDistanceToNow } from 'date-fns'
import { ja } from 'date-fns/locale'
import { useRequireAuth } from '@/context/AuthContext'
import Header from '@/components/Header'

export default function ProjectPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params)
  const currentUser = useRequireAuth()
  const queryClient = useQueryClient()
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({
    title: '',
    description: '',
    status_id: '',
    priority: 'medium' as Priority,
    assignee_id: '',
  })

  const { data: project, isLoading: projectLoading } = useQuery({
    queryKey: ['project', id],
    queryFn: () => getProject(id),
  })
  const { data: issues = [], isLoading: issuesLoading } = useQuery({
    queryKey: ['issues', id],
    queryFn: () => getIssues(id),
  })

  const { data: users = [] } = useQuery({
    queryKey: ['users'],
    queryFn: getUsers,
  })

  const createMutation = useMutation({
    mutationFn: (data: Parameters<typeof createIssue>[1]) => createIssue(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['issues', id] })
      setShowForm(false)
      setForm({ title: '', description: '', status_id: '', priority: 'medium', assignee_id: '' })
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!currentUser) return
    const statusId = form.status_id || project?.statuses?.[0]?.id
    if (!statusId) return alert('ステータスが取得できませんでした')
    createMutation.mutate({
      title: form.title,
      description: form.description || undefined,
      status_id: statusId,
      priority: form.priority,
      assignee_id: form.assignee_id || undefined,
      reporter_id: currentUser.id,
    })
  }

  if (!currentUser) return null
  if (projectLoading) return <div className="flex items-center justify-center h-screen text-gray-500">読み込み中...</div>
  if (!project) return <div className="flex items-center justify-center h-screen text-gray-500">プロジェクトが見つかりません</div>

  const statusGroups = (project.statuses || []).reduce((acc: Record<string, Issue[]>, status: Status) => {
    acc[status.id] = issues.filter((i: Issue) => i.status_id === status.id)
    return acc
  }, {} as Record<string, Issue[]>)

  return (
    <div className="min-h-screen bg-gray-50">
      <Header
        backHref="/projects"
        title={
          <div>
            <div className="flex items-center gap-2">
              <span className="text-xs font-mono font-bold text-blue-600 bg-blue-50 px-2 py-0.5 rounded">
                {project.key}
              </span>
              <h1 className="text-lg font-bold text-gray-900">{project.name}</h1>
            </div>
            {project.description && (
              <p className="text-sm text-gray-500">{project.description}</p>
            )}
          </div>
        }
        actions={
          <button
            onClick={() => setShowForm(true)}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors"
          >
            <Plus className="w-4 h-4" />
            Issue作成
          </button>
        }
      />

      {/* Board */}
      <main className="p-6 overflow-x-auto">
        {issuesLoading ? (
          <div className="text-center py-16 text-gray-500">読み込み中...</div>
        ) : (
          <div className="flex gap-4 min-w-max">
            {(project.statuses || []).map((status: Status) => (
              <div key={status.id} className="w-72 flex-shrink-0">
                <div className="flex items-center gap-2 mb-3 px-1">
                  <Circle className="w-3 h-3 flex-shrink-0" style={{ color: status.color, fill: status.color }} />
                  <span className="text-sm font-semibold text-gray-700">{status.name}</span>
                  <span className="ml-auto text-xs text-gray-400 bg-gray-100 px-2 py-0.5 rounded-full">
                    {statusGroups[status.id]?.length || 0}
                  </span>
                </div>
                <div className="space-y-2">
                  {(statusGroups[status.id] || []).map((issue: Issue) => (
                    <Link
                      key={issue.id}
                      href={`/projects/${id}/issues/${issue.number}`}
                      className="block bg-white rounded-lg border border-gray-200 p-4 hover:border-blue-300 hover:shadow-sm transition-all"
                    >
                      <div className="flex items-start justify-between gap-2">
                        <span className="text-xs text-gray-400 font-mono">{project.key}-{issue.number}</span>
                        <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${PRIORITY_COLORS[issue.priority as Priority]}`}>
                          {PRIORITY_LABELS[issue.priority as Priority]}
                        </span>
                      </div>
                      <p className="mt-1 text-sm font-medium text-gray-900 line-clamp-2">{issue.title}</p>
                      <div className="mt-3 flex items-center justify-between">
                        <span className="text-xs text-gray-400">
                          {issue.assignee ? issue.assignee.name : '未割り当て'}
                        </span>
                        <span className="text-xs text-gray-400">
                          {formatDistanceToNow(new Date(issue.updated_at), { locale: ja, addSuffix: true })}
                        </span>
                      </div>
                    </Link>
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}
      </main>

      {/* Create Issue Modal */}
      {showForm && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-xl w-full max-w-lg p-6 shadow-xl max-h-[90vh] overflow-y-auto">
            <h3 className="text-lg font-bold text-gray-900 mb-4">Issue作成</h3>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  タイトル <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  value={form.title}
                  onChange={(e) => setForm({ ...form, title: e.target.value })}
                  placeholder="Issueのタイトルを入力"
                  required
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">説明</label>
                <textarea
                  value={form.description}
                  onChange={(e) => setForm({ ...form, description: e.target.value })}
                  placeholder="詳細な説明（任意）"
                  rows={4}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">ステータス</label>
                  <select
                    value={form.status_id}
                    onChange={(e) => setForm({ ...form, status_id: e.target.value })}
                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    {(project.statuses || []).map((s: Status) => (
                      <option key={s.id} value={s.id}>{s.name}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">優先度</label>
                  <select
                    value={form.priority}
                    onChange={(e) => setForm({ ...form, priority: e.target.value as Priority })}
                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    {Object.entries(PRIORITY_LABELS).map(([value, label]) => (
                      <option key={value} value={value}>{label}</option>
                    ))}
                  </select>
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">担当者</label>
                <select
                  value={form.assignee_id}
                  onChange={(e) => setForm({ ...form, assignee_id: e.target.value })}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">未割り当て</option>
                  {users.map((u) => (
                    <option key={u.id} value={u.id}>{u.name}</option>
                  ))}
                </select>
              </div>
              <div className="flex gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setShowForm(false)}
                  className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg text-sm hover:bg-gray-50"
                >
                  キャンセル
                </button>
                <button
                  type="submit"
                  disabled={createMutation.isPending}
                  className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50"
                >
                  {createMutation.isPending ? '作成中...' : '作成する'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
