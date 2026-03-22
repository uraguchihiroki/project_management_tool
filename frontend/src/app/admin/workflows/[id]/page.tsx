'use client'

import { useState, useEffect } from 'react'
import { use } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ChevronLeft, Trash2 } from 'lucide-react'
import { useAuthFetchEnabled } from '@/hooks/useAuthFetchEnabled'
import { deleteWorkflowApi, getWorkflow, updateWorkflowMeta } from '@/lib/api'

export default function WorkflowDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params)
  const router = useRouter()
  const queryClient = useQueryClient()
  const authFetch = useAuthFetchEnabled()

  const { data: workflow, isLoading } = useQuery({
    queryKey: ['workflow', id],
    queryFn: () => getWorkflow(id),
    enabled: authFetch && !!id,
  })

  const [editing, setEditing] = useState(false)
  const [form, setForm] = useState({ name: '', description: '' })
  const [error, setError] = useState('')

  useEffect(() => {
    if (workflow) {
      setForm({ name: workflow.name, description: workflow.description ?? '' })
    }
  }, [workflow])

  const saveMutation = useMutation({
    mutationFn: () => updateWorkflowMeta(id, { name: form.name, description: form.description }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflow', id] })
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

  if (!authFetch || isLoading) {
    return <div className="p-8 text-gray-500">読み込み中...</div>
  }
  if (!workflow) {
    return <div className="p-8 text-gray-500">ワークフローが見つかりません</div>
  }

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
        <div className="flex justify-between items-start gap-4">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">{workflow.name}</h1>
            <p className="mt-2 text-gray-600 whitespace-pre-wrap">{workflow.description || '—'}</p>
          </div>
          <div className="flex gap-2">
            {!editing ? (
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
            ) : (
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
            )}
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
          </div>
        </div>

        {editing && (
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

        <p className="mt-6 text-sm text-gray-500 border-t pt-4">
          ステータス列と遷移はプロジェクト作成時や各ワークフローに紐づくステータスから管理されます（承認ステップは廃止済みです）。
        </p>

        {error && <p className="mt-4 text-sm text-red-600">{error}</p>}
      </div>
    </div>
  )
}
