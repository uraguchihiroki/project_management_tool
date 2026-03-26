'use client'

import { useState, useEffect, useMemo, useCallback } from 'react'
import { use } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import dagre from 'dagre'
import { ChevronLeft, Pencil, Plus, Trash2, X } from 'lucide-react'
import { SortableDndProvider, SortableList, SortableTbody, DragHandle } from '@/components/SortableList'
import { useAuth } from '@/context/AuthContext'
import { useWorkflowEditorUnsaved } from '@/context/AdminWorkflowEditorUnsavedContext'
import { useAuthFetchEnabled } from '@/hooks/useAuthFetchEnabled'
import type { Status, Workflow, WorkflowTransition } from '@/types'
import {
  deleteWorkflowApi,
  formatApiError,
  getWorkflow,
  getWorkflowStatuses,
  getWorkflowTransitions,
  saveWorkflowEditor,
} from '@/lib/api'
import {
  draftToPayload,
  transitionsFromDraft,
  snapshotFromServer,
  serializeEditorSnapshot,
  statusesFromDraft,
  type EditorDraftTransition,
  type EditorSnapshot,
} from '@/app/admin/workflows/[id]/workflowEditorSnapshot'

type StatusDialogMode = 'create' | 'edit'
type TransitionInvalidReason = 'same' | 'duplicate'
type TransitionDiagramNode = {
  id: string
  name: string
  color: string
  x: number
  y: number
  isEntry: boolean
  isTerminal: boolean
}
type TransitionDiagramEdge = {
  id: string
  from: string
  to: string
  invalid: boolean
}

/** API 由来の UUID 文字列の表記揺れ（大文字・小文字）で辺の端点とノード id が一致しないのを防ぐ */
function normUuid(id: string): string {
  return id.trim().toLowerCase()
}

function orient2(ax: number, ay: number, bx: number, by: number, cx: number, cy: number): number {
  return (bx - ax) * (cy - ay) - (by - ay) * (cx - ax)
}

function onSeg2(ax: number, ay: number, bx: number, by: number, cx: number, cy: number): boolean {
  return (
    cx >= Math.min(ax, bx) - 1e-9 &&
    cx <= Math.max(ax, bx) + 1e-9 &&
    cy >= Math.min(ay, by) - 1e-9 &&
    cy <= Math.max(ay, by) + 1e-9
  )
}

/** 線分 AB と線分 CD が交差（端点接触含む） */
function segmentsIntersect2(
  ax: number,
  ay: number,
  bx: number,
  by: number,
  cx: number,
  cy: number,
  dx: number,
  dy: number
): boolean {
  const o1 = orient2(ax, ay, bx, by, cx, cy)
  const o2 = orient2(ax, ay, bx, by, dx, dy)
  const o3 = orient2(cx, cy, dx, dy, ax, ay)
  const o4 = orient2(cx, cy, dx, dy, bx, by)
  if (o1 * o2 < 0 && o3 * o4 < 0) return true
  if (o1 === 0 && onSeg2(ax, ay, bx, by, cx, cy)) return true
  if (o2 === 0 && onSeg2(ax, ay, bx, by, dx, dy)) return true
  if (o3 === 0 && onSeg2(cx, cy, dx, dy, ax, ay)) return true
  if (o4 === 0 && onSeg2(cx, cy, dx, dy, bx, by)) return true
  return false
}

/** 開線分が拡張矩形と交わるか（他ノード貫通の検知用） */
function segmentHitsExpandedRect(
  ax: number,
  ay: number,
  bx: number,
  by: number,
  rx: number,
  ry: number,
  rw: number,
  rh: number
): boolean {
  const r2x = rx + rw
  const r2y = ry + rh
  const edges: [number, number, number, number][] = [
    [rx, ry, r2x, ry],
    [r2x, ry, r2x, r2y],
    [r2x, r2y, rx, r2y],
    [rx, r2y, rx, ry],
  ]
  for (const [x1, y1, x2, y2] of edges) {
    if (segmentsIntersect2(ax, ay, bx, by, x1, y1, x2, y2)) return true
  }
  return false
}

export default function WorkflowDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params)
  const router = useRouter()
  const searchParams = useSearchParams()
  const transitionDiagramDebugEnabled = searchParams.get('debug') === '1'
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

  const [error, setError] = useState('')
  const [editorBaseline, setEditorBaseline] = useState<EditorSnapshot | null>(null)
  const [editorDraft, setEditorDraft] = useState<EditorSnapshot | null>(null)

  const [statusDialogOpen, setStatusDialogOpen] = useState(false)
  const [statusDialogMode, setStatusDialogMode] = useState<StatusDialogMode>('create')
  const [statusDialogStatusId, setStatusDialogStatusId] = useState<string | null>(null)
  const [statusDialogForm, setStatusDialogForm] = useState({
    name: '',
    color: '#6B7280',
  })
  const [statusDialogError, setStatusDialogError] = useState('')
  const [transitionError, setTransitionError] = useState('')

  const isDraftDirty = useMemo(() => {
    if (!editorBaseline || !editorDraft) return false
    return serializeEditorSnapshot(editorBaseline) !== serializeEditorSnapshot(editorDraft)
  }, [editorBaseline, editorDraft])

  const { setWorkflowEditorUnsaved } = useWorkflowEditorUnsaved()
  useEffect(() => {
    setWorkflowEditorUnsaved(!!(isDraftDirty && orgMatches))
    return () => setWorkflowEditorUnsaved(false)
  }, [isDraftDirty, orgMatches, setWorkflowEditorUnsaved])

  useEffect(() => {
    setEditorBaseline(null)
    setEditorDraft(null)
  }, [id])

  useEffect(() => {
    if (!workflow || !orgMatches || statusesLoading || transitionsLoading) return
    if (isDraftDirty) return
    const snap = snapshotFromServer(workflow, statuses, transitions)
    setEditorBaseline(structuredClone(snap))
    setEditorDraft(structuredClone(snap))
  }, [workflow, statuses, transitions, orgMatches, statusesLoading, transitionsLoading, isDraftDirty])

  const saveEditorMutation = useMutation({
    mutationFn: () => {
      if (!editorDraft) throw new Error('ドラフトが未初期化です')
      return saveWorkflowEditor(id, draftToPayload(editorDraft))
    },
    onMutate: () => setError(''),
    onSuccess: async () => {
      await queryClient.refetchQueries({ queryKey: ['workflow', currentOrg?.id, id] })
      await queryClient.refetchQueries({ queryKey: ['workflow', currentOrg?.id, id, 'statuses'] })
      await queryClient.refetchQueries({ queryKey: ['workflow', currentOrg?.id, id, 'transitions'] })
      const wfRow = queryClient.getQueryData<Workflow>(['workflow', currentOrg?.id, id])
      const st = queryClient.getQueryData<Status[]>(['workflow', currentOrg?.id, id, 'statuses']) ?? []
      const tr =
        queryClient.getQueryData<WorkflowTransition[]>(['workflow', currentOrg?.id, id, 'transitions']) ?? []
      if (wfRow) {
        const snap = snapshotFromServer(wfRow, st, tr)
        setEditorBaseline(structuredClone(snap))
        setEditorDraft(structuredClone(snap))
      }
      queryClient.invalidateQueries({ queryKey: ['workflows'] })
      setError('')
      setTransitionError('')
    },
    onError: (e: unknown) => setError(formatApiError(e)),
  })

  const deleteMutation = useMutation({
    mutationFn: () => deleteWorkflowApi(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflows'] })
      router.push('/admin/workflows')
    },
    onError: (e: unknown) => setError(formatApiError(e)),
  })

  const openCreateStatusDialog = () => {
    setStatusDialogMode('create')
    setStatusDialogStatusId(null)
    setStatusDialogForm({ name: '', color: '#6B7280' })
    setStatusDialogError('')
    setStatusDialogOpen(true)
  }

  const openEditStatusDialog = (status: Status) => {
    setStatusDialogMode('edit')
    setStatusDialogStatusId(status.id)
    setStatusDialogForm({
      name: status.name,
      color: status.color,
    })
    setStatusDialogError('')
    setStatusDialogOpen(true)
  }

  const closeStatusDialog = () => {
    setStatusDialogOpen(false)
    setStatusDialogError('')
  }

  const submitStatusDialog = () => {
    const name = statusDialogForm.name.trim()
    const color = statusDialogForm.color.trim()

    if (!name) {
      setStatusDialogError('ステータス名は必須です')
      return
    }
    if (!/^#[0-9A-Fa-f]{6}$/.test(color)) {
      setStatusDialogError('色は#RRGGBB形式で指定してください')
      return
    }

    if (!editorDraft) return

    if (statusDialogMode === 'create') {
      const cid =
        typeof crypto !== 'undefined' && crypto.randomUUID
          ? crypto.randomUUID()
          : `new-${Date.now()}-${Math.random().toString(36).slice(2)}`
      setEditorDraft({
        ...editorDraft,
        statuses: [
          ...editorDraft.statuses,
          {
            client_id: cid,
            name,
            color,
            is_entry: false,
            is_terminal: false,
          },
        ],
      })
      setStatusDialogOpen(false)
      setStatusDialogError('')
      return
    }

    if (!statusDialogStatusId) {
      setStatusDialogError('編集対象のステータスが特定できません')
      return
    }
    const sid = normUuid(statusDialogStatusId)
    setEditorDraft({
      ...editorDraft,
      statuses: editorDraft.statuses.map((row) => {
        const rid = row.id ? normUuid(row.id) : normUuid(row.client_id ?? '')
        if (rid !== sid) return row
        return { ...row, name, color }
      }),
    })
    setStatusDialogOpen(false)
    setStatusDialogError('')
  }

  const visibleStatuses = useMemo(() => {
    if (!editorDraft) return []
    return statusesFromDraft(editorDraft)
  }, [editorDraft])

  const draftTransitionRows = useMemo(() => {
    if (!editorDraft) return []
    return transitionsFromDraft(editorDraft)
  }, [editorDraft])

  /** 出発/到着のいずれかが現在のステータス一覧に解決できない遷移（図では描画しない） */
  const unresolvedTransitionStrapCount = useMemo(() => {
    if (!editorDraft) return 0
    let n = 0
    for (const t of editorDraft.transitions) {
      const fromOk = visibleStatuses.some((s) => normUuid(s.id) === normUuid(t.from_ref))
      const toOk = visibleStatuses.some((s) => normUuid(s.id) === normUuid(t.to_ref))
      if (!fromOk || !toOk) n++
    }
    return n
  }, [editorDraft, visibleStatuses])

  const transitionDiagram = useMemo(() => {
    if (!editorDraft) {
      const nodeWidth = 170
      const nodeHeight = 54
      return {
        nodeWidth,
        nodeHeight,
        width: 420,
        height: 170,
        nodes: [] as TransitionDiagramNode[],
        nodeById: new Map<string, TransitionDiagramNode>(),
        edges: [] as TransitionDiagramEdge[],
      }
    }
    const visibleStatuses = statusesFromDraft(editorDraft)
    const nodeWidth = 170
    const nodeHeight = 54
    const paddingX = 36
    const paddingY = 30
    const minWidth = 420
    const minHeight = 170

    const statusIdByNorm = new Map<string, string>()
    for (const s of visibleStatuses) statusIdByNorm.set(normUuid(s.id), s.id)
    const canonStatusId = (raw: string) => statusIdByNorm.get(normUuid(raw)) ?? raw

    const persistedEdges: TransitionDiagramEdge[] = editorDraft.transitions.map((t) => {
      let invalid = normUuid(t.from_ref) === normUuid(t.to_ref)
      if (!invalid) {
        for (const u of editorDraft.transitions) {
          if (u.key === t.key) continue
          if (
            normUuid(u.from_ref) === normUuid(t.from_ref) &&
            normUuid(u.to_ref) === normUuid(t.to_ref)
          ) {
            invalid = true
            break
          }
        }
      }
      return {
        id: `draft-${t.key}`,
        from: canonStatusId(t.from_ref),
        to: canonStatusId(t.to_ref),
        invalid,
      }
    })
    const edgeCandidates = persistedEdges

    const visibleNodeIdsNorm = new Set(visibleStatuses.map((s) => normUuid(s.id)))
    const edges = edgeCandidates.filter(
      (e) => visibleNodeIdsNorm.has(normUuid(e.from)) && visibleNodeIdsNorm.has(normUuid(e.to))
    )

    const degreeById = new Map<string, number>()
    for (const s of visibleStatuses) degreeById.set(s.id, 0)
    for (const e of edges) {
      degreeById.set(e.from, (degreeById.get(e.from) ?? 0) + 1)
      degreeById.set(e.to, (degreeById.get(e.to) ?? 0) + 1)
    }

    const connectedStatuses = visibleStatuses.filter((s) => (degreeById.get(s.id) ?? 0) > 0)
    const disconnectedStatuses = visibleStatuses.filter((s) => (degreeById.get(s.id) ?? 0) === 0)

    const graph = new dagre.graphlib.Graph()
    graph.setGraph({
      rankdir: 'LR',
      ranksep: 100,
      nodesep: 52,
      marginx: paddingX,
      marginy: paddingY,
    })
    graph.setDefaultEdgeLabel(() => ({}))

    for (const s of connectedStatuses) {
      graph.setNode(s.id, { width: nodeWidth, height: nodeHeight })
    }
    for (const e of edges) {
      if (!graph.hasNode(e.from) || !graph.hasNode(e.to)) continue
      graph.setEdge(e.from, e.to)
    }
    if (connectedStatuses.length > 0) {
      dagre.layout(graph)
    }

    const connectedNodes: TransitionDiagramNode[] = connectedStatuses.map((s) => {
      const laid = graph.node(s.id) as { x: number; y: number } | undefined
      const defaultX = paddingX + nodeWidth / 2
      const defaultY = paddingY + nodeHeight / 2
      const centerX = laid?.x ?? defaultX
      const centerY = laid?.y ?? defaultY
      return {
        id: s.id,
        name: s.name,
        color: s.color,
        x: Math.round(centerX - nodeWidth / 2),
        y: Math.round(centerY - nodeHeight / 2),
        isEntry: s.is_entry === true,
        isTerminal: s.is_terminal === true,
      }
    })

    const connectedMaxX =
      connectedNodes.length > 0
        ? Math.max(...connectedNodes.map((n) => n.x + nodeWidth))
        : paddingX + nodeWidth
    const disconnectedStartX = connectedMaxX + 120
    const disconnectedNodes: TransitionDiagramNode[] = disconnectedStatuses.map((s, idx) => ({
      id: s.id,
      name: s.name,
      color: s.color,
      x: disconnectedStartX,
      y: paddingY + idx * (nodeHeight + 22),
      isEntry: s.is_entry === true,
      isTerminal: s.is_terminal === true,
    }))

    const nodes = [...connectedNodes, ...disconnectedNodes]
    const nodeById = new Map(nodes.map((n) => [n.id, n]))
    const nodeIdsNorm = new Set(nodes.map((n) => normUuid(n.id)))
    const drawnEdges = edges.filter(
      (e) => nodeIdsNorm.has(normUuid(e.from)) && nodeIdsNorm.has(normUuid(e.to))
    )

    // Curves can extend beyond node bounds. Reserve vertical padding and shift the whole drawing down if needed.
    const hasLoop = drawnEdges.some((e) => e.from === e.to)
    const hasReverseHorizontal = drawnEdges.some((e) => {
      const f = nodeById.get(e.from)
      const t = nodeById.get(e.to)
      if (!f || !t) return false
      const fromCx = f.x + nodeWidth / 2
      const toCx = t.x + nodeWidth / 2
      return toCx < fromCx
    })
    const extraTop = (hasReverseHorizontal ? 120 : 0) + (hasLoop ? 120 : 0) + 36
    const extraBottom = (hasReverseHorizontal ? 72 : 0) + 36

    for (const n of nodes) n.y += extraTop

    const width = Math.max(
      minWidth,
      nodes.length > 0 ? Math.max(...nodes.map((n) => n.x + nodeWidth)) + paddingX : minWidth
    )
    const height = Math.max(
      minHeight,
      nodes.length > 0 ? Math.max(...nodes.map((n) => n.y + nodeHeight)) + paddingY + extraBottom : minHeight
    )

    return {
      nodeWidth,
      nodeHeight,
      width,
      height,
      nodes,
      nodeById,
      edges: drawnEdges,
    }
  }, [editorDraft])

  const transitionDiagramDebugPayload = useMemo(() => {
    const label = (sid: string) =>
      visibleStatuses.find((s) => normUuid(s.id) === normUuid(sid))?.name ?? sid
    const edgesDrawn = transitionDiagram.edges.map((e) => ({
      id: e.id,
      from: e.from,
      to: e.to,
      fromLabel: label(e.from),
      toLabel: label(e.to),
      invalid: e.invalid,
      reverseEdgeExists: transitionDiagram.edges.some(
        (x) => normUuid(x.from) === normUuid(e.to) && normUuid(x.to) === normUuid(e.from)
      ),
    }))
    const forwardKeys = new Set(
      transitionDiagram.edges.map((e) => `${normUuid(e.from)}>${normUuid(e.to)}`)
    )
    const seenUndirected = new Set<string>()
    const bidirectionalPairsInDraw: string[] = []
    for (const e of transitionDiagram.edges) {
      const a = normUuid(e.from)
      const b = normUuid(e.to)
      const uKey = a < b ? `${a}|${b}` : `${b}|${a}`
      if (seenUndirected.has(uKey)) continue
      seenUndirected.add(uKey)
      if (forwardKeys.has(`${a}>${b}`) && forwardKeys.has(`${b}>${a}`)) {
        bidirectionalPairsInDraw.push(`${label(e.from)} ⟷ ${label(e.to)}`)
      }
    }
    return {
      visibleStatusIds: visibleStatuses.map((s) => s.id),
      visibleStatusKeys: visibleStatuses.map((s) => s.status_key),
      draftTransitions: (editorDraft?.transitions ?? []).map((t) => ({
        key: t.key,
        from: t.from_ref,
        to: t.to_ref,
        fromLabel: label(t.from_ref),
        toLabel: label(t.to_ref),
      })),
      bidirectionalPairsInDraw,
      edgesDrawn,
      unresolvedStrapCount: unresolvedTransitionStrapCount,
      strapsInDraft: editorDraft?.transitions.length ?? 0,
      edgesInDiagram: transitionDiagram.edges.length,
      nodesForDraw: transitionDiagram.nodes.map((n) => ({
        id: n.id,
        name: n.name,
        x: n.x,
        y: n.y,
      })),
      svg: { width: transitionDiagram.width, height: transitionDiagram.height },
    }
  }, [transitionDiagram, visibleStatuses, editorDraft?.transitions, unresolvedTransitionStrapCount])

  useEffect(() => {
    if (process.env.NODE_ENV !== 'development') return
    console.log('[WorkflowTransitionDiagram] draw payload', transitionDiagramDebugPayload)
  }, [transitionDiagramDebugPayload])

  const applyEntryStatus = (targetId: string) => {
    if (!editorDraft) return
    const tid = normUuid(targetId)
    setEditorDraft({
      ...editorDraft,
      statuses: editorDraft.statuses.map((s) => {
        const sid = normUuid(s.id ?? s.client_id ?? '')
        return { ...s, is_entry: sid === tid }
      }),
    })
  }

  const toggleTerminalStatus = (st: Status) => {
    if (!editorDraft) return
    const tid = normUuid(st.id)
    setEditorDraft({
      ...editorDraft,
      statuses: editorDraft.statuses.map((s) => {
        const sid = normUuid(s.id ?? s.client_id ?? '')
        if (sid !== tid) return s
        if (s.is_entry) return s
        return { ...s, is_terminal: !s.is_terminal }
      }),
    })
  }

  const entryRadioName = `workflow-${id}-entry`

  const statusReferencedInTransition = (statusId: string) => {
    const u = normUuid(statusId)
    if (!editorDraft) return false
    return editorDraft.transitions.some(
      (t) => normUuid(t.from_ref) === u || normUuid(t.to_ref) === u
    )
  }
  const initialFromStatusId = visibleStatuses[0]?.id
  const initialToStatusId = visibleStatuses[1]?.id

  const transitionReasonForDraftKey = (
    key: string,
    fromStatusId: string,
    toStatusId: string
  ): TransitionInvalidReason | null => {
    if (!editorDraft) return null
    if (normUuid(fromStatusId) === normUuid(toStatusId)) return 'same'
    for (const t of editorDraft.transitions) {
      if (t.key === key) continue
      if (
        normUuid(t.from_ref) === normUuid(fromStatusId) &&
        normUuid(t.to_ref) === normUuid(toStatusId)
      )
        return 'duplicate'
    }
    return null
  }

  const transitionStrapPending = saveEditorMutation.isPending

  const reorderDraftTransitions = (orderedKeys: string[]) => {
    if (!editorDraft) return
    const byKey = new Map(editorDraft.transitions.map((x) => [x.key, x]))
    const next = orderedKeys.map((k) => byKey.get(k)).filter(Boolean) as EditorDraftTransition[]
    setEditorDraft({ ...editorDraft, transitions: next })
  }

  const reorderDraftStatuses = (orderedIds: string[]) => {
    if (!editorDraft) return
    const byEff = new Map(
      editorDraft.statuses.map((s) => [normUuid(s.id ?? s.client_id ?? ''), s])
    )
    const next = orderedIds
      .map((oid) => byEff.get(normUuid(oid)))
      .filter(Boolean) as typeof editorDraft.statuses
    setEditorDraft({ ...editorDraft, statuses: next })
  }

  const addTransitionStrap = () => {
    if (!editorDraft || !initialFromStatusId || !initialToStatusId) return
    const key = `n-${Date.now()}-${Math.random().toString(36).slice(2)}`
    setEditorDraft({
      ...editorDraft,
      transitions: [
        ...editorDraft.transitions,
        {
          key,
          from_ref: normUuid(initialFromStatusId),
          to_ref: normUuid(initialToStatusId),
        },
      ],
    })
  }

  const deleteTransitionStrapByKey = (transitionKey: string) => {
    if (!editorDraft) return
    setEditorDraft({
      ...editorDraft,
      transitions: editorDraft.transitions.filter((t) => t.key !== transitionKey),
    })
  }

  const changeTransitionStrap = (
    transitionKey: string,
    field: 'from_status_id' | 'to_status_id',
    value: string
  ) => {
    if (!editorDraft) return
    const v = normUuid(value)
    setEditorDraft({
      ...editorDraft,
      transitions: editorDraft.transitions.map((t) => {
        if (t.key !== transitionKey) return t
        if (field === 'from_status_id') return { ...t, from_ref: v }
        return { ...t, to_ref: v }
      }),
    })
  }

  const removeStatusFromDraft = (statusId: string) => {
    if (!editorDraft) return
    const u = normUuid(statusId)
    const nextStatuses = editorDraft.statuses.filter(
      (s) => normUuid(s.id ?? s.client_id ?? '') !== u
    )
    const nextTrans = editorDraft.transitions.filter(
      (t) => normUuid(t.from_ref) !== u && normUuid(t.to_ref) !== u
    )
    setEditorDraft({ ...editorDraft, statuses: nextStatuses, transitions: nextTrans })
  }

  const tryNavigateAway = useCallback(
    (href: string) => {
      if (
        orgMatches &&
        isDraftDirty &&
        !window.confirm('保存していない変更があります。ページを離れますか？')
      ) {
        return
      }
      router.push(href)
    },
    [orgMatches, isDraftDirty, router]
  )

  useEffect(() => {
    const onBeforeUnload = (e: BeforeUnloadEvent) => {
      if (!isDraftDirty || !orgMatches) return
      e.preventDefault()
    }
    window.addEventListener('beforeunload', onBeforeUnload)
    return () => window.removeEventListener('beforeunload', onBeforeUnload)
  }, [isDraftDirty, orgMatches])

  const cancelEditorDraft = () => {
    if (!editorBaseline) return
    if (
      isDraftDirty &&
      !window.confirm('保存していない変更を破棄してよいですか？')
    ) {
      return
    }
    setEditorDraft(structuredClone(editorBaseline))
  }

  if (!authFetch) {
    return <div className="p-8 text-gray-500">読み込み中...</div>
  }
  if (!currentOrg?.id) {
    return (
      <div className="w-full max-w-screen-2xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
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
  const atStatusDeleteFloor = (editorDraft?.statuses.length ?? 0) <= 2
  const statusDialogTitle = statusDialogMode === 'create' ? 'ステータスを追加' : 'ステータスを編集'
  const statusDialogSaving = false

  return (
    <div className="w-full max-w-screen-2xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
      <button
        type="button"
        onClick={() => tryNavigateAway('/admin/workflows')}
        className="inline-flex items-center gap-1 text-sm text-gray-600 hover:text-gray-900 mb-6"
      >
        <ChevronLeft className="w-4 h-4" />
        一覧へ
      </button>

      <div className="bg-white rounded-xl border border-gray-200 p-6">
        {orgMismatch && (
          <div className="mb-4 rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
            このワークフローは<strong>現在選択中の組織</strong>に属していません。右上で組織を切り替えるか、
            <button
              type="button"
              onClick={() => tryNavigateAway('/admin/workflows')}
              className="text-blue-700 underline"
            >
              ワークフロー一覧
            </button>
            から開き直してください。
          </div>
        )}
        <div className="flex justify-between items-start gap-4">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">
              {editorDraft?.meta.name ?? workflow.name}
            </h1>
            {!orgMismatch && currentOrg && (
              <p className="mt-1 text-sm text-gray-500">
                選択中の組織: <span className="font-medium text-gray-800">{currentOrg.name}</span>
              </p>
            )}
            <p className="mt-2 text-xs text-amber-800 bg-amber-50 border border-amber-200 rounded px-2 py-1.5 inline-block">
              変更は「保存する」でサーバに反映されます（画面は自動では遷移しません）。
            </p>
          </div>
          <div className="flex flex-wrap gap-2 justify-end">
            {!orgMismatch && (
              <>
                <button
                  type="button"
                  onClick={() => saveEditorMutation.mutate()}
                  disabled={
                    saveEditorMutation.isPending ||
                    !isDraftDirty ||
                    !(editorDraft?.meta.name?.trim()) ||
                    visibleStatuses.length < 2
                  }
                  className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
                >
                  {saveEditorMutation.isPending ? '保存中…' : '保存する'}
                </button>
                <button
                  type="button"
                  onClick={cancelEditorDraft}
                  disabled={saveEditorMutation.isPending || !isDraftDirty}
                  className="px-3 py-1.5 text-sm border rounded-lg disabled:opacity-50"
                >
                  変更を破棄
                </button>
              </>
            )}
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

        {error && (
          <div
            className="mt-4 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-900"
            role="alert"
          >
            <p className="font-medium text-red-800">保存または削除に失敗しました</p>
            <p className="mt-1 whitespace-pre-wrap">{error}</p>
          </div>
        )}

        {!orgMismatch && (
          <div className="mt-6 space-y-4 border-t pt-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">ワークフロー名</label>
              <input
                value={editorDraft?.meta.name ?? ''}
                onChange={(e) =>
                  editorDraft &&
                  setEditorDraft({
                    ...editorDraft,
                    meta: { ...editorDraft.meta, name: e.target.value },
                  })
                }
                className="w-full border rounded-lg px-3 py-2"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">説明</label>
              <textarea
                value={editorDraft?.meta.description ?? ''}
                onChange={(e) =>
                  editorDraft &&
                  setEditorDraft({
                    ...editorDraft,
                    meta: { ...editorDraft.meta, description: e.target.value },
                  })
                }
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
              onClick={() => addTransitionStrap()}
              disabled={visibleStatuses.length < 2}
              title={
                visibleStatuses.length < 2
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
        {!orgMismatch && (
          <div className="mb-6 rounded-lg border border-gray-200 bg-gray-50/60 p-4">
            <h3 className="text-sm font-semibold text-gray-900">遷移図</h3>
            <p className="mt-1 text-xs text-gray-600">現在の許可遷移を図で表示しています。</p>
            {unresolvedTransitionStrapCount > 0 && (
              <div
                className="mt-3 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-xs text-amber-950"
                role="status"
              >
                <p className="font-medium text-amber-900">
                  参照切れの遷移が {unresolvedTransitionStrapCount} 件あります
                </p>
                <p className="mt-1 text-amber-900/95">
                  出発または到着のステータス ID が、このワークフローのステータス一覧に無いため遷移図には描画されません（下のストラップ一覧には出ます。選択が意図とずれる行がある場合は同じ理由です）。DB
                  の件数が多いときは、同一 <code className="rounded bg-amber-100/80 px-1">workflow_id</code>{' '}
                  で絞っているか、論理削除{' '}
                  <code className="rounded bg-amber-100/80 px-1">deleted_at</code> が付いた行を数えていないかを確認してください。
                </p>
              </div>
            )}
            {(statusesLoading || transitionsLoading) && (
              <p className="mt-3 text-sm text-gray-500">遷移図を読み込み中...</p>
            )}
            {!statusesLoading && !transitionsLoading && visibleStatuses.length < 2 && (
              <p className="mt-3 text-sm text-gray-500">ステータスが2つ以上あるときに遷移図を表示できます。</p>
            )}
            {!statusesLoading && !transitionsLoading && visibleStatuses.length >= 2 && (
              <div className="mt-3 overflow-x-auto">
                <svg
                  width="100%"
                  viewBox={`0 0 ${transitionDiagram.width} ${transitionDiagram.height}`}
                  preserveAspectRatio="xMinYMin meet"
                  className="h-auto min-w-[640px] xl:min-w-0"
                  role="img"
                  aria-label="ステータス遷移図"
                >
                  <defs>
                    <marker
                      id="transition-arrow"
                      markerWidth="10"
                      markerHeight="10"
                      refX="8"
                      refY="3"
                      orient="auto"
                      markerUnits="strokeWidth"
                    >
                      <path d="M0,0 L0,6 L9,3 z" fill="#94A3B8" />
                    </marker>
                    <marker
                      id="transition-arrow-invalid"
                      markerWidth="10"
                      markerHeight="10"
                      refX="8"
                      refY="3"
                      orient="auto"
                      markerUnits="strokeWidth"
                    >
                      <path d="M0,0 L0,6 L9,3 z" fill="#D97706" />
                    </marker>
                  </defs>

                  {/* SVG 描画順: 先に edges、後に nodes。後から描いたノードが手前（Z 上）になる。順序を入れ替えないこと。 */}
                  <g data-layer="edges">
                  {transitionDiagram.edges.map((edge, idx) => {
                    const from = transitionDiagram.nodeById.get(edge.from)
                    const to = transitionDiagram.nodeById.get(edge.to)
                    if (!from || !to) return null

                    const samePair = transitionDiagram.edges.filter(
                      (e) =>
                        (normUuid(e.from) === normUuid(edge.from) &&
                          normUuid(e.to) === normUuid(edge.to)) ||
                        (normUuid(e.from) === normUuid(edge.to) &&
                          normUuid(e.to) === normUuid(edge.from))
                    )
                    const samePairIndex = samePair.findIndex((e) => e.id === edge.id)
                    const pairSep = samePair.length > 1 ? 34 : 18
                    const pairOffset = (samePairIndex - (samePair.length - 1) / 2) * pairSep

                    let path = ''
                    if (edge.from === edge.to) {
                      // 自己ループのみ折れ線（曲線は使わない）
                      const sx = from.x + transitionDiagram.nodeWidth
                      const sy = from.y + transitionDiagram.nodeHeight / 2
                      const pad = 28 + (idx % 3) * 7
                      const lift = 38 + (idx % 2) * 10
                      const lx = from.x
                      path = `M ${sx} ${sy} L ${sx + pad} ${sy} L ${sx + pad} ${sy - lift} L ${lx} ${sy - lift} L ${lx} ${sy}`
                    } else {
                      const fromCx = from.x + transitionDiagram.nodeWidth / 2
                      const fromCy = from.y + transitionDiagram.nodeHeight / 2
                      const toCx = to.x + transitionDiagram.nodeWidth / 2
                      const toCy = to.y + transitionDiagram.nodeHeight / 2
                      const centerDx = toCx - fromCx
                      const centerDy = toCy - fromCy

                      const horizontalDominant = Math.abs(centerDx) >= Math.abs(centerDy)
                      const fromCandidates = [
                        { x: from.x + transitionDiagram.nodeWidth, y: fromCy, side: 'right' as const },
                        { x: from.x, y: fromCy, side: 'left' as const },
                        { x: fromCx, y: from.y, side: 'top' as const },
                        { x: fromCx, y: from.y + transitionDiagram.nodeHeight, side: 'bottom' as const },
                      ]
                      const toCandidates = [
                        { x: to.x + transitionDiagram.nodeWidth, y: toCy, side: 'right' as const },
                        { x: to.x, y: toCy, side: 'left' as const },
                        { x: toCx, y: to.y, side: 'top' as const },
                        { x: toCx, y: to.y + transitionDiagram.nodeHeight, side: 'bottom' as const },
                      ]

                      let best = {
                        sx: fromCandidates[0].x,
                        sy: fromCandidates[0].y,
                        ex: toCandidates[0].x,
                        ey: toCandidates[0].y,
                        score: Number.POSITIVE_INFINITY,
                      }
                      for (const s of fromCandidates) {
                        for (const t of toCandidates) {
                          const dx = t.x - s.x
                          const dy = t.y - s.y
                          const d2 = dx * dx + dy * dy
                          const sIsVertical = s.side === 'top' || s.side === 'bottom'
                          const tIsVertical = t.side === 'top' || t.side === 'bottom'
                          const orientationPenalty =
                            (horizontalDominant && (sIsVertical || tIsVertical)) ||
                            (!horizontalDominant && (!sIsVertical || !tIsVertical))
                              ? 8000
                              : 0
                          const score = d2 + orientationPenalty
                          if (score < best.score) best = { sx: s.x, sy: s.y, ex: t.x, ey: t.y, score }
                        }
                      }

                      const startX = best.sx
                      const startY = best.sy
                      const endX = best.ex
                      const endY = best.ey

                      const dx = endX - startX
                      const dy = endY - startY
                      const dist = Math.max(1, Math.hypot(dx, dy))
                      const ux = dx / dist
                      const uy = dy / dist
                      const nx = -uy
                      const ny = ux

                      // 直線＋法線方向の平行オフセット（双方向の分離＋他ノード貫通の回避）
                      const basePerp = samePair.length > 1 ? pairOffset : 0
                      const obstaclePad = 8
                      const avoidStep = 20
                      const nw = transitionDiagram.nodeWidth
                      const nh = transitionDiagram.nodeHeight
                      const avoidMaxK = Math.max(32, Math.ceil((4 * nh) / avoidStep))

                      const blockers = transitionDiagram.nodes.filter(
                        (n) =>
                          normUuid(n.id) !== normUuid(edge.from) &&
                          normUuid(n.id) !== normUuid(edge.to)
                      )

                      const hitsObstacle = (perpMag: number) => {
                        const px1 = startX + nx * perpMag
                        const py1 = startY + ny * perpMag
                        const px2 = endX + nx * perpMag
                        const py2 = endY + ny * perpMag
                        for (const n of blockers) {
                          const rx = n.x - obstaclePad
                          const ry = n.y - obstaclePad
                          const rw = nw + obstaclePad * 2
                          const rh = nh + obstaclePad * 2
                          if (segmentHitsExpandedRect(px1, py1, px2, py2, rx, ry, rw, rh)) {
                            return true
                          }
                        }
                        return false
                      }

                      let chosenPerp = basePerp
                      if (hitsObstacle(chosenPerp)) {
                        let bestPerp: number | null = null
                        let bestAbs = Number.POSITIVE_INFINITY
                        for (let k = 1; k <= avoidMaxK; k++) {
                          for (const s of [1, -1] as const) {
                            const tryPerp = basePerp + s * k * avoidStep
                            if (!hitsObstacle(tryPerp)) {
                              const ad = Math.abs(tryPerp - basePerp)
                              if (ad < bestAbs) {
                                bestAbs = ad
                                bestPerp = tryPerp
                              }
                            }
                          }
                        }
                        if (bestPerp !== null) chosenPerp = bestPerp
                      }

                      const stillBlocked = hitsObstacle(chosenPerp)
                      if (!stillBlocked) {
                        const sx1 = startX + nx * chosenPerp
                        const sy1 = startY + ny * chosenPerp
                        const ex1 = endX + nx * chosenPerp
                        const ey1 = endY + ny * chosenPerp
                        path = `M ${sx1} ${sy1} L ${ex1} ${ey1}`
                      } else {
                        // 諦めモード: 三次ベジェ（ノード矩形との交差は検査しない。Z 順でノードが手前）
                        const fromNodeCx = from.x + nw / 2
                        let anchor: (typeof transitionDiagram.nodes)[number] | null = null
                        let minAnchorCx = Number.POSITIVE_INFINITY
                        for (const n of transitionDiagram.nodes) {
                          if (
                            normUuid(n.id) === normUuid(edge.from) ||
                            normUuid(n.id) === normUuid(edge.to)
                          ) {
                            continue
                          }
                          const cx = n.x + nw / 2
                          if (cx > fromNodeCx && cx < minAnchorCx) {
                            minAnchorCx = cx
                            anchor = n
                          }
                        }
                        const midX = (startX + endX) / 2
                        const midY = (startY + endY) / 2
                        const bulgeSign =
                          (idx + Math.round(pairOffset / 17)) % 2 === 0 ? 1 : -1
                        const bulgeY = bulgeSign * nh
                        let c1x = startX + (endX - startX) * 0.33
                        let c1y = startY + (endY - startY) * 0.33 + bulgeY
                        let c2x = startX + (endX - startX) * 0.67
                        let c2y = startY + (endY - startY) * 0.67 + bulgeY
                        if (anchor) {
                          const ax = anchor.x + nw / 2
                          const ay = anchor.y + nh / 2
                          const pull = 0.22
                          c1x = c1x * (1 - pull) + ax * pull
                          c1y = c1y * (1 - pull) + ay * pull
                          c2x = c2x * (1 - pull) + ax * pull
                          c2y = c2y * (1 - pull) + ay * pull
                        } else {
                          const fx = midX
                          const fy = midY + bulgeY
                          const pull = 0.35
                          c1x = c1x * (1 - pull) + fx * pull
                          c1y = c1y * (1 - pull) + fy * pull
                          c2x = c2x * (1 - pull) + fx * pull
                          c2y = c2y * (1 - pull) + fy * pull
                        }
                        const r = (v: number) => Math.round(v * 10) / 10
                        path = `M ${r(startX)} ${r(startY)} C ${r(c1x)} ${r(c1y)} ${r(c2x)} ${r(c2y)} ${r(endX)} ${r(endY)}`
                      }
                    }

                    return (
                      <path
                        key={edge.id}
                        d={path}
                        fill="none"
                        stroke={edge.invalid ? '#D97706' : '#94A3B8'}
                        strokeWidth="2"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeDasharray={edge.invalid ? '5 4' : undefined}
                        markerEnd={edge.invalid ? 'url(#transition-arrow-invalid)' : 'url(#transition-arrow)'}
                      />
                    )
                  })}
                  </g>

                  <g data-layer="nodes">
                  {transitionDiagram.nodes.map((node) => (
                    <g key={node.id}>
                      <rect
                        x={node.x}
                        y={node.y}
                        width={transitionDiagram.nodeWidth}
                        height={transitionDiagram.nodeHeight}
                        rx="10"
                        fill="#ffffff"
                        stroke="#D1D5DB"
                      />
                      <circle cx={node.x + 16} cy={node.y + 16} r="6" fill={node.color} />
                      <text
                        x={node.x + 30}
                        y={node.y + 20}
                        fontSize="12"
                        fontWeight="600"
                        fill="#111827"
                        dominantBaseline="middle"
                      >
                        {node.name}
                      </text>
                      {node.isEntry && (
                        <text
                          x={node.x + transitionDiagram.nodeWidth - 8}
                          y={node.y + 12}
                          fontSize="9"
                          fontWeight="700"
                          fill="#0369A1"
                          textAnchor="end"
                        >
                          START
                        </text>
                      )}
                      {node.isTerminal && (
                        <text
                          x={node.x + transitionDiagram.nodeWidth - 8}
                          y={node.y + (node.isEntry ? 24 : 12)}
                          fontSize="9"
                          fontWeight="700"
                          fill="#15803D"
                          textAnchor="end"
                        >
                          GOAL
                        </text>
                      )}
                    </g>
                  ))}
                  </g>
                </svg>
                {transitionDiagram.edges.length === 0 && (
                  <p className="mt-2 text-xs text-gray-500">遷移はまだありません。下の「ストラップを追加」から作成できます。</p>
                )}
                {transitionDiagram.edges.some((e) => e.invalid) && (
                  <p className="mt-2 text-xs text-amber-700">
                    破線の矢印は未保存の無効な遷移候補（同一ステータスまたは重複）です。
                  </p>
                )}
                {transitionDiagramDebugEnabled && (
                  <pre className="mt-3 max-h-64 overflow-auto rounded border border-amber-200 bg-amber-50/80 p-3 text-left text-[11px] leading-snug text-gray-800">
                    {JSON.stringify(transitionDiagramDebugPayload, null, 2)}
                  </pre>
                )}
              </div>
            )}
          </div>
        )}
        {orgMismatch && (
          <p className="text-sm text-gray-500">
            選択中の組織に属するワークフローのみ、遷移設定を表示・編集できます。
          </p>
        )}
        {!orgMismatch && (statusesLoading || transitionsLoading) && (
          <p className="text-sm text-gray-500">遷移設定を読み込み中...</p>
        )}
        {!orgMismatch && !statusesLoading && !transitionsLoading && visibleStatuses.length < 2 && (
          <p className="text-sm text-gray-500">
            ステータスが2つ以上あるときのみ遷移を追加できます。まず「ステータスを追加」でステータスを用意してください。
          </p>
        )}
        {!orgMismatch &&
          !statusesLoading &&
          !transitionsLoading &&
          draftTransitionRows.length === 0 &&
          visibleStatuses.length >= 2 && (
          <p className="text-sm text-gray-500">遷移はまだありません。「ストラップを追加」から作成できます。</p>
        )}
        {!orgMismatch && !transitionsLoading && draftTransitionRows.length > 0 && (
          <div className="space-y-2">
            <SortableDndProvider
              items={editorDraft?.transitions ?? []}
              itemId={(t) => t.key}
              onReorder={reorderDraftTransitions}
              disabled={transitionStrapPending}
            >
              <SortableList
                items={editorDraft?.transitions ?? []}
                itemId={(t) => t.key}
                onReorder={() => {}}
                renderItem={(t, props) => {
                  const reason = transitionReasonForDraftKey(t.key, t.from_ref, t.to_ref)
                  const fromSt = visibleStatuses.find((x) => normUuid(x.id) === normUuid(t.from_ref))
                  const toSt = visibleStatuses.find((x) => normUuid(x.id) === normUuid(t.to_ref))
                  const fromVal = fromSt?.id ?? visibleStatuses[0]?.id ?? ''
                  const toVal = toSt?.id ?? visibleStatuses[0]?.id ?? ''
                  return (
                    <div
                      key={t.key}
                      ref={props.setNodeRef}
                      style={props.style}
                      className={`flex flex-wrap xl:flex-nowrap items-center gap-2 rounded-lg border px-3 py-2 ${reason ? 'border-amber-300 bg-amber-50' : 'border-gray-200'}`}
                    >
                      <DragHandle handleProps={props.handleProps} />
                      <select
                        value={fromVal}
                        onChange={(e) => changeTransitionStrap(t.key, 'from_status_id', e.target.value)}
                        disabled={transitionStrapPending}
                        className="min-w-44 border rounded-lg px-2 py-1.5 text-sm"
                      >
                        {visibleStatuses.map((s) => (
                          <option key={s.id} value={s.id}>
                            {s.name}
                          </option>
                        ))}
                      </select>
                      {fromSt?.is_entry && (
                        <span className="text-[10px] font-semibold text-sky-700 border border-sky-200 rounded px-1 py-0.5">
                          開始
                        </span>
                      )}
                      {fromSt?.is_terminal && (
                        <span className="text-[10px] font-semibold text-emerald-800 border border-emerald-200 rounded px-1 py-0.5">
                          終了
                        </span>
                      )}
                      <span className="text-gray-500">→</span>
                      <select
                        value={toVal}
                        onChange={(e) => changeTransitionStrap(t.key, 'to_status_id', e.target.value)}
                        disabled={transitionStrapPending}
                        className="min-w-44 border rounded-lg px-2 py-1.5 text-sm"
                      >
                        {visibleStatuses.map((s) => (
                          <option key={s.id} value={s.id}>
                            {s.name}
                          </option>
                        ))}
                      </select>
                      {toSt?.is_entry && (
                        <span className="text-[10px] font-semibold text-sky-700 border border-sky-200 rounded px-1 py-0.5">
                          開始
                        </span>
                      )}
                      {toSt?.is_terminal && (
                        <span className="text-[10px] font-semibold text-emerald-800 border border-emerald-200 rounded px-1 py-0.5">
                          終了
                        </span>
                      )}
                      {reason === 'same' && (
                        <span className="text-xs text-amber-700 border border-amber-300 rounded px-2 py-0.5">
                          無効（遷移前後が同一）
                        </span>
                      )}
                      {reason === 'duplicate' && (
                        <span className="text-xs text-amber-700 border border-amber-300 rounded px-2 py-0.5">
                          無効（重複）
                        </span>
                      )}
                      <button
                        type="button"
                        onClick={() => deleteTransitionStrapByKey(t.key)}
                        disabled={transitionStrapPending}
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

        {!orgMismatch && (
          <p className="mb-3 text-xs text-gray-600">
            開始は常に1件（ラジオで付け替え）。新規ワークフローの既定は表示順が最小の列。終了は複数（チェック）。並べ替えは左のハンドルをドラッグ。
          </p>
        )}

        {orgMismatch && (
          <p className="text-sm text-gray-500">
            選択中の組織に属するワークフローのみ、ステータス一覧を表示・編集できます。
          </p>
        )}

        {!orgMismatch && statusesLoading && (
          <p className="text-sm text-gray-500">ステータスを読み込み中...</p>
        )}
        {!orgMismatch && !statusesLoading && visibleStatuses.length === 0 && (
          <p className="text-sm text-gray-500">まだステータスがありません。「ステータスを追加」から作成できます。</p>
        )}
        {!orgMismatch && !statusesLoading && visibleStatuses.length > 0 && (
          <div className="overflow-x-auto rounded-lg border border-gray-200">
            <SortableDndProvider
              items={visibleStatuses}
              itemId={(s) => s.id}
              onReorder={(ids) => reorderDraftStatuses(ids)}
              disabled={saveEditorMutation.isPending || statusesLoading}
            >
              <table className="min-w-[720px] w-full text-sm">
                <thead className="bg-gray-50 text-left text-gray-600">
                  <tr>
                    <th className="w-10 px-2 py-2" aria-hidden />
                    <th className="px-3 py-2 font-medium">名前</th>
                    <th className="px-3 py-2 font-medium">色</th>
                    <th className="px-3 py-2 font-medium w-24">開始</th>
                    <th className="px-3 py-2 font-medium w-24">終了</th>
                    <th className="px-3 py-2 font-medium w-28">操作</th>
                  </tr>
                </thead>
                <SortableTbody
                  items={visibleStatuses}
                  itemId={(s) => s.id}
                  disabled={saveEditorMutation.isPending || statusesLoading}
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
                      <td className="px-3 py-2 font-medium text-gray-900">{s.name}</td>
                      <td className="px-3 py-2">
                        <span
                          className="inline-block w-6 h-6 rounded border border-gray-200 align-middle"
                          style={{ backgroundColor: s.color }}
                          title={s.color}
                        />
                        <span className="ml-2 text-gray-600 font-mono text-xs">{s.color}</span>
                      </td>
                      <td className="px-3 py-2 align-middle">
                        <input
                          type="radio"
                          name={entryRadioName}
                          checked={s.is_entry === true}
                          onChange={() => applyEntryStatus(s.id)}
                          disabled={saveEditorMutation.isPending}
                          aria-label={`${s.name} を開始にする`}
                        />
                      </td>
                      <td className="px-3 py-2 align-middle">
                        <input
                          type="checkbox"
                          checked={s.is_terminal === true}
                          onChange={() => toggleTerminalStatus(s)}
                          disabled={saveEditorMutation.isPending || s.is_entry === true}
                          aria-label={`${s.name} を終了にする`}
                        />
                      </td>
                      <td className="px-3 py-2">
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
                              saveEditorMutation.isPending ||
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
                                removeStatusFromDraft(s.id)
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
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">色</label>
                <input
                  type="color"
                  value={statusDialogForm.color}
                  onChange={(e) => setStatusDialogForm((f) => ({ ...f, color: e.target.value }))}
                  className="h-10 w-14 rounded border cursor-pointer"
                />
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
