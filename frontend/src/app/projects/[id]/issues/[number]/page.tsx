'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  getProject, getIssue, createComment, updateIssue,
  getApprovals, approveStep, rejectStep,
  getIssueEvents,
} from '@/lib/api'
import { useState, use } from 'react'
import Link from 'next/link'
import { Circle, MessageSquare, Send, CheckCircle, XCircle, Clock, History, Users } from 'lucide-react'
import { PRIORITY_LABELS, PRIORITY_COLORS, type Priority } from '@/types'
import type { Status, Comment, IssueApproval, IssueEvent } from '@/types'
import { format } from 'date-fns'
import { ja } from 'date-fns/locale'
import { useRequireAuth } from '@/context/AuthContext'
import Header from '@/components/Header'
import { useAuthFetchEnabled } from '@/hooks/useAuthFetchEnabled'

const APPROVAL_STATUS_LABELS: Record<string, string> = {
  pending: '未承認',
  approved: '承認済み',
  rejected: '却下',
}

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

const ApprovalStatusIcon = ({ status }: { status: string }) => {
  if (status === 'approved') return <CheckCircle className="w-5 h-5 text-green-500" />
  if (status === 'rejected') return <XCircle className="w-5 h-5 text-red-500" />
  return <Clock className="w-5 h-5 text-gray-400" />
}

export default function IssuePage({ params }: { params: Promise<{ id: string; number: string }> }) {
  const { id, number } = use(params)
  const currentUser = useRequireAuth()
  const authFetch = useAuthFetchEnabled()
  const queryClient = useQueryClient()
  const [comment, setComment] = useState('')
  const [editingStatus, setEditingStatus] = useState(false)
  const [approvalComment, setApprovalComment] = useState<Record<string, string>>({})

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

  const { data: approvals = [] } = useQuery({
    queryKey: ['approvals', issue?.id],
    queryFn: () => getApprovals(issue!.id),
    enabled: authFetch && !!issue?.id,
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

  const approveMutation = useMutation({
    mutationFn: ({ approvalId, comment }: { approvalId: string; comment: string }) =>
      approveStep(approvalId, currentUser!.id, comment),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['approvals', issue?.id] })
      queryClient.invalidateQueries({ queryKey: ['issue', id, number] })
      queryClient.invalidateQueries({ queryKey: ['issue-events', issue?.id] })
    },
  })

  const rejectMutation = useMutation({
    mutationFn: ({ approvalId, comment }: { approvalId: string; comment: string }) =>
      rejectStep(approvalId, currentUser!.id, comment),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['approvals', issue?.id] })
      queryClient.invalidateQueries({ queryKey: ['issue-events', issue?.id] })
    },
  })

  if (!currentUser) return null
  if (isLoading) return <div className="flex items-center justify-center h-screen text-gray-500">読み込み中...</div>
  if (!issue) return <div className="flex items-center justify-center h-screen text-gray-500">Issueが見つかりません</div>

  const statusById = new Map<string, string>((project?.statuses ?? []).map((s) => [s.id, s.name]))

  // 承認ステップをorder順にソート
  const sortedApprovals = [...approvals].sort(
    (a, b) => (a.workflow_step?.order ?? 0) - (b.workflow_step?.order ?? 0)
  )

  // 現在のアクティブなステップ（最初のpendingステップ）
  const activeApproval = sortedApprovals.find((a) => a.status === 'pending')

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
          {/* Main Content */}
          <div className="col-span-2 space-y-6">
            {/* Issue Title & Description */}
            <div className="bg-white rounded-xl border border-gray-200 p-6">
              <h1 className="text-2xl font-bold text-gray-900">{issue.title}</h1>
              {issue.description ? (
                <p className="mt-4 text-gray-700 leading-relaxed whitespace-pre-wrap">{issue.description}</p>
              ) : (
                <p className="mt-4 text-gray-400 italic">説明はありません</p>
              )}
            </div>

            {/* インプリント（履歴） */}
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

            {/* Approval Steps */}
            {sortedApprovals.length > 0 && (
              <div className="bg-white rounded-xl border border-gray-200 p-6">
                <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
                  <CheckCircle className="w-5 h-5 text-blue-500" />
                  承認フロー
                </h2>
                <div className="space-y-4">
                  {sortedApprovals.map((approval: IssueApproval, idx) => {
                    const step = approval.workflow_step
                    const isActive = activeApproval?.id === approval.id

                    return (
                      <div
                        key={approval.id}
                        className={`rounded-lg border p-4 transition-colors ${
                          isActive
                            ? 'border-blue-300 bg-blue-50'
                            : approval.status === 'approved'
                            ? 'border-green-200 bg-green-50'
                            : approval.status === 'rejected'
                            ? 'border-red-200 bg-red-50'
                            : 'border-gray-200 bg-gray-50'
                        }`}
                      >
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-3">
                            <div className="flex items-center justify-center w-7 h-7 rounded-full bg-white border text-xs font-bold text-gray-500">
                              {step?.order}
                            </div>
                            <div>
                              <p className="text-sm font-semibold text-gray-900">{step?.status?.name ?? step?.status_id}</p>
                              <p className="text-xs text-gray-500">閾値: {step?.threshold ?? 10}</p>
                            </div>
                          </div>
                          <div className="flex items-center gap-2">
                            <ApprovalStatusIcon status={approval.status} />
                            <span className={`text-xs font-medium ${
                              approval.status === 'approved' ? 'text-green-600'
                              : approval.status === 'rejected' ? 'text-red-600'
                              : 'text-gray-500'
                            }`}>
                              {APPROVAL_STATUS_LABELS[approval.status]}
                            </span>
                          </div>
                        </div>

                        {/* 承認者情報 */}
                        {approval.approver && (
                          <div className="mt-2 text-xs text-gray-500 flex items-center gap-2">
                            <span>{approval.approver.name}</span>
                            {approval.acted_at && (
                              <span>{format(new Date(approval.acted_at), 'yyyy/MM/dd HH:mm', { locale: ja })}</span>
                            )}
                            {approval.comment && <span className="italic">「{approval.comment}」</span>}
                          </div>
                        )}

                        {/* アクティブなステップの承認・却下ボタン */}
                        {isActive && (
                          <div className="mt-3 space-y-2">
                            <textarea
                              value={approvalComment[approval.id] ?? ''}
                              onChange={(e) =>
                                setApprovalComment((prev) => ({ ...prev, [approval.id]: e.target.value }))
                              }
                              placeholder="コメント（任意）"
                              rows={2}
                              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none bg-white"
                            />
                            <div className="flex gap-2">
                              <button
                                onClick={() =>
                                  approveMutation.mutate({
                                    approvalId: approval.id,
                                    comment: approvalComment[approval.id] ?? '',
                                  })
                                }
                                disabled={approveMutation.isPending}
                                className="flex items-center gap-1.5 px-4 py-2 bg-green-600 text-white text-sm rounded-lg hover:bg-green-700 disabled:opacity-50 font-medium"
                              >
                                <CheckCircle className="w-4 h-4" />
                                承認
                              </button>
                              <button
                                onClick={() =>
                                  rejectMutation.mutate({
                                    approvalId: approval.id,
                                    comment: approvalComment[approval.id] ?? '',
                                  })
                                }
                                disabled={rejectMutation.isPending}
                                className="flex items-center gap-1.5 px-4 py-2 bg-red-600 text-white text-sm rounded-lg hover:bg-red-700 disabled:opacity-50 font-medium"
                              >
                                <XCircle className="w-4 h-4" />
                                却下
                              </button>
                            </div>
                            {(approveMutation.isError || rejectMutation.isError) && (
                              <p className="text-xs text-red-600">
                                {String((approveMutation.error || rejectMutation.error) ?? '操作に失敗しました')}
                              </p>
                            )}
                          </div>
                        )}
                      </div>
                    )
                  })}
                </div>

                {/* 全ステップ承認完了メッセージ */}
                {sortedApprovals.length > 0 && sortedApprovals.every((a) => a.status === 'approved') && (
                  <div className="mt-4 p-3 bg-green-100 rounded-lg text-sm text-green-700 font-medium text-center">
                    すべての承認ステップが完了しました
                  </div>
                )}
                {sortedApprovals.some((a) => a.status === 'rejected') && (
                  <div className="mt-4 p-3 bg-red-100 rounded-lg text-sm text-red-700 font-medium text-center">
                    承認フローが却下されました
                  </div>
                )}
              </div>
            )}

            {/* Comments */}
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

              {/* Comment Input */}
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

          {/* Sidebar */}
          <div className="space-y-4">
            {/* Status */}
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

            {/* Groups */}
            {(issue.groups?.length ?? 0) > 0 && (
              <div className="bg-white rounded-xl border border-gray-200 p-4">
                <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-3 flex items-center gap-1">
                  <Users className="w-3.5 h-3.5" />
                  グループ
                </h3>
                <div className="flex flex-wrap gap-2" data-testid="issue-groups">
                  {(issue.groups ?? []).map((g) => (
                    <span
                      key={g.id}
                      className="text-xs px-2 py-1 rounded-md bg-indigo-50 text-indigo-800 border border-indigo-100"
                    >
                      {g.name}
                    </span>
                  ))}
                </div>
              </div>
            )}

            {/* Details */}
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

            {/* Approval Summary */}
            {sortedApprovals.length > 0 && (
              <div className="bg-white rounded-xl border border-gray-200 p-4">
                <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-3">承認進捗</h3>
                <div className="flex items-center gap-2">
                  <div className="flex-1 bg-gray-200 rounded-full h-2">
                    <div
                      className="bg-green-500 h-2 rounded-full transition-all"
                      style={{
                        width: `${(sortedApprovals.filter((a) => a.status === 'approved').length / sortedApprovals.length) * 100}%`,
                      }}
                    />
                  </div>
                  <span className="text-xs text-gray-500 whitespace-nowrap">
                    {sortedApprovals.filter((a) => a.status === 'approved').length} / {sortedApprovals.length}
                  </span>
                </div>
                {sortedApprovals.some((a) => a.status === 'rejected') && (
                  <p className="mt-2 text-xs text-red-600 font-medium">却下あり</p>
                )}
              </div>
            )}
          </div>
        </div>
      </main>
    </div>
  )
}
