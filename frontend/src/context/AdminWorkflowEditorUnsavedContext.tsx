'use client'

import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
  type MouseEvent,
  type ReactNode,
} from 'react'
import { useRouter } from 'next/navigation'

const LEAVE_CONFIRM_MESSAGE = '保存していない変更があります。ページを離れますか？'

type AdminWorkflowEditorUnsavedContextValue = {
  workflowEditorUnsaved: boolean
  setWorkflowEditorUnsaved: (v: boolean) => void
}

const defaultValue: AdminWorkflowEditorUnsavedContextValue = {
  workflowEditorUnsaved: false,
  setWorkflowEditorUnsaved: () => {},
}

const AdminWorkflowEditorUnsavedContext = createContext<AdminWorkflowEditorUnsavedContextValue>(defaultValue)

export function AdminWorkflowEditorUnsavedProvider({ children }: { children: ReactNode }) {
  const [workflowEditorUnsaved, setWorkflowEditorUnsaved] = useState(false)
  const value = useMemo(
    () => ({ workflowEditorUnsaved, setWorkflowEditorUnsaved }),
    [workflowEditorUnsaved]
  )
  return (
    <AdminWorkflowEditorUnsavedContext.Provider value={value}>{children}</AdminWorkflowEditorUnsavedContext.Provider>
  )
}

export function useWorkflowEditorUnsaved() {
  return useContext(AdminWorkflowEditorUnsavedContext)
}

/** Provider 外では常に未保存なし。管理レイアウト内でワークフロー編集ページがフラグを立てる。 */
export function useNavigateAwayWithWorkflowGuard() {
  const router = useRouter()
  const { workflowEditorUnsaved } = useWorkflowEditorUnsaved()
  return useCallback(
    (href: string) => {
      if (workflowEditorUnsaved && !window.confirm(LEAVE_CONFIRM_MESSAGE)) {
        return
      }
      router.push(href)
    },
    [router, workflowEditorUnsaved]
  )
}

/** Link の onClick 用: キャンセル時のみ preventDefault */
export function useWorkflowEditorLeaveConfirmHandler() {
  const { workflowEditorUnsaved } = useWorkflowEditorUnsaved()
  return useCallback(
    (e: MouseEvent<HTMLAnchorElement>) => {
      if (workflowEditorUnsaved && !window.confirm(LEAVE_CONFIRM_MESSAGE)) {
        e.preventDefault()
      }
    },
    [workflowEditorUnsaved]
  )
}
