import type { WorkflowEditorStatusPayload, WorkflowEditorTransitionPayload } from '@/lib/api'
import type { Status, Workflow, WorkflowTransition } from '@/types'

export type EditorDraftStatus = {
  id?: string
  client_id?: string
  name: string
  color: string
  is_entry: boolean
  is_terminal: boolean
}

export type EditorDraftTransition = {
  key: string
  from_ref: string
  to_ref: string
}

export type EditorSnapshot = {
  meta: { name: string; description: string }
  statuses: EditorDraftStatus[]
  transitions: EditorDraftTransition[]
}

export function normUuidLocal(id: string): string {
  return id.trim().toLowerCase()
}

export function snapshotFromServer(
  wf: Workflow,
  statuses: Status[],
  transitions: WorkflowTransition[]
): EditorSnapshot {
  const sorted = [...statuses].sort((a, b) => a.display_order - b.display_order)
  return {
    meta: { name: wf.name, description: wf.description ?? '' },
    statuses: sorted.map((s) => ({
      id: s.id,
      name: s.name,
      color: s.color,
      is_entry: s.is_entry === true,
      is_terminal: s.is_terminal === true,
    })),
    transitions: transitions.map((t) => ({
      key: `p-${t.id}`,
      from_ref: normUuidLocal(t.from_status_id),
      to_ref: normUuidLocal(t.to_status_id),
    })),
  }
}

export function serializeEditorSnapshot(s: EditorSnapshot): string {
  return JSON.stringify(s)
}

/** Sortable / 図用の仮 ID（安定） */
export function transitionKeyToNumericId(key: string): number {
  let h = 0
  for (let i = 0; i < key.length; i++) {
    h = (Math.imul(31, h) + key.charCodeAt(i)) | 0
  }
  return h >= 0 ? -h - 1 : h
}

export function statusesFromDraft(draft: EditorSnapshot): Status[] {
  return draft.statuses.map((st, i) => ({
    id: st.id ?? st.client_id!,
    name: st.name,
    color: st.color,
    display_order: i + 1,
    is_entry: st.is_entry,
    is_terminal: st.is_terminal,
  }))
}

export function transitionsFromDraft(draft: EditorSnapshot): WorkflowTransition[] {
  return draft.transitions.map((t, i) => ({
    id: transitionKeyToNumericId(t.key),
    workflow_id: 0,
    from_status_id: t.from_ref,
    to_status_id: t.to_ref,
    display_order: i + 1,
    created_at: '',
  }))
}

export function draftToPayload(s: EditorSnapshot): {
  name: string
  description: string
  statuses: WorkflowEditorStatusPayload[]
  transitions: WorkflowEditorTransitionPayload[]
} {
  return {
    name: s.meta.name.trim(),
    description: s.meta.description,
    statuses: s.statuses.map((st) => {
      const color = (st.color.trim() || '#6B7280').slice(0, 7)
      if (st.id) {
        return {
          id: st.id,
          name: st.name.trim(),
          color,
          is_entry: st.is_entry,
          is_terminal: st.is_terminal,
        }
      }
      return {
        client_id: st.client_id!,
        name: st.name.trim(),
        color,
        is_entry: st.is_entry,
        is_terminal: st.is_terminal,
      }
    }),
    transitions: s.transitions.map((t) => ({
      from_ref: normUuidLocal(t.from_ref),
      to_ref: normUuidLocal(t.to_ref),
    })),
  }
}
