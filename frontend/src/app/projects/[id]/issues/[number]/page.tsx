'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getProject, getIssue, createComment, updateIssue, getIssueEvents } from '@/lib/api'
import { useState, use } from 'react'
import Link from 'next/link'
import { Circle, MessageSquare, Send, History } from 'lucide-react'
import { PRIORITY_LABELS, PRIORITY_COLORS, type Priority } from '@/types'
import type { Status, Comment, IssueEvent } from '@/types'
import { format } from 'date-fns'
import { ja } from 'date-fns/locale'
import { useRequireAuth } from '@/context/AuthContext'
import Header from '@/components/Header'
import { useAuthFetchEnabled } from '@/hooks/useAuthFetchEnabled'

const ISSUE_EVENT_TYPE_LABELS: Record<string, string> = {
  'issue.status_changed': 'ステータス変更',
  'issue.assignee_changed': '担当者変更',
}

function labelForIssueEvent(ev: IssueEvent, statusById: Map<string, string>): string {
  const base = ISSUE_EVENT_TYPE_LABELS[ev.event_type] ?? ev.event_type
  if (ev.event_type === 'issue.status_changed' && (ev.from_status_id || ev.to_status_id)) {
    const from = ev.from_status_id ? statusById.get(ev.from_status_id) ?? '—' : '—'
    const to = ev.to_status_id ? statusById.get(ev.to_status_id) ?? '—' : '—'
    return `${base}: ${from} → ${to}`
  }
  return base
}

export default function IssuePage({ params }: { params: Promise<{ id: string; number: string }> }) {
  const { id, number } = use(params)
  const currentUser = useRequireAuth()
  const authFetch = useAuthFetchEnabled()
  const queryClient = useQueryClient()
  const [comment, setComment] = useState('')
  const [editingStatus, setEditingStatus] = useState(false)

  const { data: project } = useQuery({
    queryKey: ['project', id],
    queryFn: () => getProject(id),
    enabled: authFetch && !!id,
  })

  const { data: issue, isLoading } = useQuery({
    queryKey: ['issue', id, number],
    queryFn: () => getIssue(id, Number(number)),
    enabled: authFetch && !!id && !!number,
  })

  const { data: issueEvents = [] } = useQuery({
    queryKey: ['issue-events', issue?.id],
    queryFn: () => getIssueEvents(issue!.id),
    enabled: authFetch && !!issue?.id,
  })

  const commentMutation = useMutation({
    mutationFn: (body: string) =>
      createComment(issue!.id, { author_id: currentUser?.id ?? '', body }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['issue', id, number] })
      setComment('')
    },
  })

  const updateStatusMutation = useMutation({
    mutationFn: (statusId: string) =>
      updateIssue(id, Number(number), { status_id: statusId }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['issue', id, number] })
      queryClient.invalidateQueries({ queryKey: ['issue-events', issue?.id] })
      setEditingStatus(false)
    },
  })

  if (!currentUser) return null
  if (isLoading) return <div className="flex items-center justify-center h-screen text-gray-500">読み込み中...</div>
  if (!issue) return <div className="flex items-center justify-center h-screen text-gray-500">Issueが見つかりません</div>

  const statusById = new Map<string, string>((project?.statuses ?? []).map((s) => [s.id, s.name]))

  return (
    <div className="min-h-screen bg-gray-50">
      <Header
        backHref={`/projects/${id}`}
        title={
          <div className="flex items-center gap-2 text-sm text-gray-500">
            <Link href="/projects" className="hover:text-blue-600">{project?.name}</Link>
            <span>/</span>
            <span className="font-mono">{project?.key}-{issue.number}</span>
          </div>
        }
      />

      <main className="max-w-5xl mx-auto px-6 py-8">
        <div className="grid grid-cols-3 gap-6">
          <div className="col-span-2 space-y-6">
            <div className="bg-white rounded-xl border border-gray-200 p-6">
              <h1 className="text-2xl font-bold text-gray-900">{issue.title}</h1>
              {issue.description ? (
                <p className="mt-4 text-gray-700 leading-relaxed whitespace-pre-wrap">{issue.description}</p>
              ) : (
                <p className="mt-4 text-gray-400 italic">説明はありません</p>
              )}
            </div>

            <div
              className="bg-white rounded-xl border border-gray-200 p-6"
              data-testid="issue-imprint-timeline"
            >
              <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
                <History className="w-5 h-5 text-gray-600" />
                履歴
              </h2>
              {issueEvents.length === 0 ? (
                <p className="text-sm text-gray-400">まだ記録がありません（ステータスや担当の変更などがここに表示されます）</p>
              ) : (
                <ul className="space-y-3 border-l-2 border-gray-200 ml-2 pl-4">
                  {issueEvents.map((ev: IssueEvent) => (
                    <li key={ev.id} className="relative">
                      <span className="absolute -left-[21px] top-1.5 w-2 h-2 rounded-full bg-blue-500 border-2 border-white" />
                      <p className="text-sm text-gray-900">{labelForIssueEvent(ev, statusById)}</p>
                      <p className="text-xs text-gray-500 mt-0.5">
                        {ev.actor?.name ?? '—'} ·{' '}
                        {format(new Date(ev.occurred_at), 'yyyy/MM/dd HH:mm:ss', { locale: ja })}
                      </p>
                    </li>
                  ))}
                </ul>
              )}
            </div>

            <div className="bg-white rounded-xl border border-gray-200 p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
                <MessageSquare className="w-5 h-5" />
                コメント ({issue.comments?.length || 0})
              </h2>
              <div className="space-y-4">
                {(issue.comments || []).map((c: Comment) => (
                  <div key={c.id} className="flex gap-3">
                    <div className="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center flex-shrink-0">
                      <span className="text-xs font-bold text-blue-600">
                        {c.author?.name?.[0]?.toUpperCase()}
                      </span>
                    </div>
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-medium text-gray-900">{c.author?.name}</span>
                        <span className="text-xs text-gray-400">
                          {format(new Date(c.created_at), 'yyyy/MM/dd HH:mm', { locale: ja })}
                        </span>
                      </div>
                      <p className="mt-1 text-sm text-gray-700 whitespace-pre-wrap">{c.body}</p>
                    </div>
                  </div>
                ))}
              </div>

              <div className="mt-6 flex gap-3">
                <div className="w-8 h-8 rounded-full bg-gray-100 flex items-center justify-center flex-shrink-0">
                  <span className="text-xs font-bold text-gray-500">
                    {currentUser.name[0]?.toUpperCase()}
                  </span>
                </div>
                <div className="flex-1 flex gap-2">
                  <textarea
                    value={comment}
                    onChange={(e) => setComment(e.target.value)}
                    placeholder="コメントを入力..."
                    rows={3}
                    className="flex-1 border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
                  />
                  <button
                    onClick={() => comment.trim() && commentMutation.mutate(comment)}
                    disabled={!comment.trim() || commentMutation.isPending}
                    className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 self-end"
                  >
                    <Send className="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <div className="bg-white rounded-xl border border-gray-200 p-4">
              <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-3">ステータス</h3>
              {editingStatus ? (
                <div className="space-y-1">
                  {(project?.statuses || []).map((s: Status) => (
                    <button
                      key={s.id}
                      onClick={() => updateStatusMutation.mutate(s.id)}
                      className="w-full flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-gray-50 text-left"
                    >
                      <Circle className="w-3 h-3 flex-shrink-0" style={{ color: s.color, fill: s.color }} />
                      <span className="text-sm text-gray-700">{s.name}</span>
                    </button>
                  ))}
                  <button onClick={() => setEditingStatus(false)} className="w-full text-xs text-gray-400 hover:text-gray-600 mt-1">
                    キャンセル
                  </button>
                </div>
              ) : (
                <button
                  onClick={() => setEditingStatus(true)}
                  className="flex items-center gap-2 px-3 py-2 rounded-lg border border-gray-200 hover:border-blue-300 w-full"
                >
                  <Circle
                    className="w-3 h-3 flex-shrink-0"
                    style={{ color: issue.status?.color, fill: issue.status?.color }}
                  />
                  <span className="text-sm text-gray-700">{issue.status?.name}</span>
                </button>
              )}
            </div>

            <div className="bg-white rounded-xl border border-gray-200 p-4 space-y-3">
              <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wide">詳細</h3>
              <div>
                <p className="text-xs text-gray-400 mb-1">優先度</p>
                <span className={`text-xs px-2 py-1 rounded-full font-medium ${PRIORITY_COLORS[issue.priority as Priority]}`}>
                  {PRIORITY_LABELS[issue.priority as Priority]}
                </span>
              </div>
              <div>
                <p className="text-xs text-gray-400 mb-1">担当者</p>
                <p className="text-sm text-gray-700">{issue.assignee?.name || '未割り当て'}</p>
              </div>
              <div>
                <p className="text-xs text-gray-400 mb-1">起票者</p>
                <p className="text-sm text-gray-700">{issue.reporter?.name}</p>
              </div>
              {issue.due_date && (
                <div>
                  <p className="text-xs text-gray-400 mb-1">期日</p>
                  <p className="text-sm text-gray-700">
                    {format(new Date(issue.due_date), 'yyyy/MM/dd', { locale: ja })}
                  </p>
                </div>
              )}
              <div>
                <p className="text-xs text-gray-400 mb-1">作成日時</p>
                <p className="text-sm text-gray-700">
                  {format(new Date(issue.created_at), 'yyyy/MM/dd HH:mm', { locale: ja })}
                </p>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  )
}
