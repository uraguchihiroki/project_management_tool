'use client'

import { use, useState } from 'react'
import Link from 'next/link'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getProject, getProjectStatuses, updateProject, updateProjectStatus } from '@/lib/api'
import { ChevronLeft, Pencil, Check, X } from 'lucide-react'
import type { ProjectStatus } from '@/types'
import { useRequireAdmin } from '@/context/AuthContext'
import { useAuthFetchEnabled } from '@/hooks/useAuthFetchEnabled'

export default function AdminProjectEditPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params)
  const currentUser = useRequireAdmin()
  const authFetch = useAuthFetchEnabled()
  const queryClient = useQueryClient()

  const { data: project, isLoading } = useQuery({
    queryKey: ['project', id],
    queryFn: () => getProject(id),
    enabled: authFetch && !!id,
  })

  const { data: projectStatuses = [], isLoading: psLoading } = useQuery({
    queryKey: ['project-statuses', id],
    queryFn: () => getProjectStatuses(id),
    enabled: authFetch && !!id && !!project,
  })

  const updateMutation = useMutation({
    mutationFn: (data: Parameters<typeof updateProject>[1]) => updateProject(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project', id] })
      queryClient.invalidateQueries({ queryKey: ['projects'] })
    },
  })

  const [editingPsId, setEditingPsId] = useState<string | null>(null)
  const [psForm, setPsForm] = useState({ name: '', color: '#6B7280', order: 1 })
  const [psError, setPsError] = useState('')

  const updatePsMutation = useMutation({
    mutationFn: ({ statusId, data }: { statusId: string; data: { name: string; color: string; order: number } }) =>
      updateProjectStatus(id, statusId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-statuses', id] })
      queryClient.invalidateQueries({ queryKey: ['project', id] })
      setEditingPsId(null)
      setPsError('')
    },
    onError: (e: Error) => setPsError(e.message),
  })

  const startEditPs = (s: ProjectStatus) => {
    setEditingPsId(s.id)
    setPsForm({ name: s.name, color: s.color, order: s.order })
    setPsError('')
  }

  const handleSubmitPs = (e: React.FormEvent) => {
    e.preventDefault()
    if (!editingPsId) return
    if (!/^#[0-9A-Fa-f]{6}$/.test(psForm.color)) {
      setPsError('色は#RRGGBB形式で指定してください')
      return
    }
    updatePsMutation.mutate({
      statusId: editingPsId,
      data: { name: psForm.name.trim(), color: psForm.color, order: psForm.order },
    })
  }

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    if (!currentUser || !project) return
    const form = e.currentTarget
    const formData = new FormData(form)
    updateMutation.mutate({
      name: (formData.get('name') as string) || undefined,
      description: (formData.get('description') as string) || undefined,
      start_date: (formData.get('start_date') as string) || undefined,
      end_date: (formData.get('end_date') as string) || undefined,
    })
  }

  if (!currentUser) return null
  if (isLoading) return <div className="text-center py-16 text-gray-500">読み込み中...</div>
  if (!project) return <div className="text-center py-16 text-gray-500">プロジェクトが見つかりません</div>

  return (
    <div>
      <Link
        href="/admin/projects"
        className="inline-flex items-center gap-1 text-sm text-gray-500 hover:text-blue-600 mb-6"
      >
        <ChevronLeft className="w-4 h-4" />
        プロジェクト管理に戻る
      </Link>

      <h1 className="text-xl font-bold text-gray-900 mb-6">プロジェクト編集</h1>

      <form onSubmit={handleSubmit} className="max-w-xl space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">プロジェクトキー</label>
          <input
            type="text"
            value={project.key}
            disabled
            className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm bg-gray-50 text-gray-500"
          />
          <p className="mt-0.5 text-xs text-gray-400">キーは変更できません</p>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            プロジェクト名 <span className="text-red-500">*</span>
          </label>
          <input
            name="name"
            type="text"
            defaultValue={project.name}
            required
            className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">説明</label>
          <textarea
            name="description"
            defaultValue={project.description ?? ''}
            rows={3}
            className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">開始日</label>
            <input
              name="start_date"
              type="date"
              defaultValue={project.start_date ?? ''}
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">終了日</label>
            <input
              name="end_date"
              type="date"
              defaultValue={project.end_date ?? ''}
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
        </div>

        {updateMutation.isError && (
          <p className="text-sm text-red-500 bg-red-50 px-3 py-2 rounded-lg">
            {updateMutation.error instanceof Error ? updateMutation.error.message : '更新に失敗しました'}
          </p>
        )}

        <div className="flex gap-3 pt-2">
          <Link
            href="/admin/projects"
            className="px-4 py-2 border border-gray-300 text-gray-700 rounded-lg text-sm hover:bg-gray-50"
          >
            キャンセル
          </Link>
          <button
            type="submit"
            disabled={updateMutation.isPending}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50"
          >
            {updateMutation.isPending ? '保存中...' : '保存'}
          </button>
        </div>
      </form>

      <section className="mt-12 max-w-xl">
        <h2 className="text-lg font-semibold text-gray-900 mb-1">プロジェクト進行ステータス</h2>
        <p className="text-sm text-gray-500 mb-4">
          計画中・進行中・完了などの表示名・色・並び順を編集できます（Issue のカンバン列とは別です）。
        </p>

        {psLoading ? (
          <p className="text-sm text-gray-400">読み込み中...</p>
        ) : projectStatuses.length === 0 ? (
          <p className="text-sm text-gray-400">進行ステータスがありません</p>
        ) : (
          <div className="bg-white border border-gray-200 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th className="text-left px-4 py-2 font-medium text-gray-600">順</th>
                  <th className="text-left px-4 py-2 font-medium text-gray-600">名前</th>
                  <th className="text-left px-4 py-2 font-medium text-gray-600">色</th>
                  <th className="w-12"></th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {projectStatuses.map((s) => (
                  <tr key={s.id}>
                    {editingPsId === s.id ? (
                      <td colSpan={4} className="px-4 py-3 bg-gray-50">
                        <form onSubmit={handleSubmitPs} className="space-y-3">
                          {psError && <p className="text-sm text-red-600">{psError}</p>}
                          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                            <div>
                              <label className="block text-xs text-gray-500 mb-1">名前</label>
                              <input
                                value={psForm.name}
                                onChange={(e) => setPsForm((f) => ({ ...f, name: e.target.value }))}
                                className="w-full border rounded px-2 py-1.5 text-sm"
                                required
                              />
                            </div>
                            <div>
                              <label className="block text-xs text-gray-500 mb-1">色</label>
                              <div className="flex gap-2">
                                <input
                                  type="color"
                                  value={psForm.color}
                                  onChange={(e) => setPsForm((f) => ({ ...f, color: e.target.value }))}
                                  className="h-9 w-12 rounded border cursor-pointer"
                                />
                                <input
                                  value={psForm.color}
                                  onChange={(e) => setPsForm((f) => ({ ...f, color: e.target.value }))}
                                  className="flex-1 border rounded px-2 py-1.5 text-xs font-mono"
                                />
                              </div>
                            </div>
                            <div>
                              <label className="block text-xs text-gray-500 mb-1">並び</label>
                              <input
                                type="number"
                                min={1}
                                value={psForm.order}
                                onChange={(e) =>
                                  setPsForm((f) => ({ ...f, order: parseInt(e.target.value, 10) || 1 }))
                                }
                                className="w-full border rounded px-2 py-1.5 text-sm"
                              />
                            </div>
                          </div>
                          <div className="flex gap-2">
                            <button
                              type="submit"
                              disabled={updatePsMutation.isPending}
                              className="inline-flex items-center gap-1 px-3 py-1.5 bg-blue-600 text-white rounded text-xs font-medium disabled:opacity-50"
                            >
                              <Check className="w-3.5 h-3.5" />
                              更新
                            </button>
                            <button
                              type="button"
                              onClick={() => {
                                setEditingPsId(null)
                                setPsError('')
                              }}
                              className="inline-flex items-center gap-1 px-3 py-1.5 border rounded text-xs"
                            >
                              <X className="w-3.5 h-3.5" />
                              キャンセル
                            </button>
                          </div>
                        </form>
                      </td>
                    ) : (
                      <>
                        <td className="px-4 py-2 text-gray-600">{s.order}</td>
                        <td className="px-4 py-2 font-medium text-gray-900">{s.name}</td>
                        <td className="px-4 py-2">
                          <span
                            className="inline-block w-6 h-6 rounded border border-gray-200 align-middle"
                            style={{ backgroundColor: s.color }}
                            title={s.color}
                          />
                          <span className="ml-2 text-xs font-mono text-gray-500">{s.color}</span>
                        </td>
                        <td className="px-4 py-2">
                          <button
                            type="button"
                            onClick={() => startEditPs(s)}
                            className="p-1.5 text-gray-400 hover:text-blue-600 rounded"
                            title="編集"
                          >
                            <Pencil className="w-4 h-4" />
                          </button>
                        </td>
                      </>
                    )}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </div>
  )
}
