export interface Organization {
  id: string
  name: string
  created_at: string
}

export interface Department {
  id: string
  organization_id: string
  name: string
  order: number
  created_at: string
}

export interface SuperAdmin {
  id: string
  name: string
  email: string
  created_at: string
}

export interface Role {
  id: number
  name: string
  level: number
  description: string
  created_at: string
}

export interface User {
  id: string
  name: string
  email: string
  avatar_url?: string
  is_admin: boolean
  roles?: Role[]
  created_at: string
}

/** プロジェクト進行（Workflow 不使用） */
export interface ProjectStatus {
  id: string
  project_id: string
  name: string
  color: string
  order: number
  status_key?: string
}

export interface Project {
  id: string
  key: string
  name: string
  description?: string
  owner_id: string
  owner: User
  organization_id?: string
  start_date?: string
  end_date?: string
  /** Issue 用（default_workflow の列） */
  statuses?: Status[]
  project_status_id?: string
  project_status?: ProjectStatus
  created_at: string
}

/** Issue 用（ワークフローに紐づく） */
export interface Status {
  id: string
  project_id?: string
  organization_id?: string
  workflow_id?: number
  name: string
  color: string
  display_order: number
  status_key?: string // sts_start, sts_goal。空=ユーザー定義
}

export type Priority = 'low' | 'medium' | 'high' | 'critical'

export interface IssueTemplate {
  id: number
  project_id: string
  project?: Project
  name: string
  description: string
  body: string
  default_priority: Priority
  workflow_id?: number
  workflow?: Workflow
  created_at: string
}

/** 組織スコープのグループ（Issue への紐付け・通知等） */
/** issue_events の1行（インプリント） */
export interface IssueEvent {
  id: string
  organization_id: string
  issue_id: string
  actor_id: string
  actor?: User
  event_type: string
  occurred_at: string
  from_status_id?: string
  to_status_id?: string
  assignee_id_at_occurred?: string
  payload?: Record<string, unknown>
}

export interface Issue {
  id: string
  number: number
  title: string
  description?: string
  status_id: string
  status: Status
  priority: Priority
  assignee_id?: string
  assignee?: User
  reporter_id: string
  reporter: User
  organization_id: string
  project_id?: string
  due_date?: string
  template_id?: number
  workflow_id?: number
  comments?: Comment[]
  created_at: string
  updated_at: string
}

export interface Comment {
  id: string
  issue_id: string
  author_id: string
  author: User
  body: string
  created_at: string
  updated_at: string
}

export interface WorkflowStep {
  id: number
  workflow_id: number
  order: number
  status_id: string
  status?: Status
  next_status_id?: string
  next_status?: Status
  description?: string
  threshold?: number
  exclude_reporter?: boolean
  exclude_assignee?: boolean
}

export interface Workflow {
  id: number
  organization_id: string
  name: string
  description: string
  steps?: WorkflowStep[]
  created_at: string
}

export interface WorkflowTransition {
  id: number
  workflow_id: number
  from_status_id: string
  to_status_id: string
  display_order: number
  created_at: string
}

export interface ApiResponse<T> {
  data: T
  message?: string
}

export interface ListResponse<T> {
  data: T[]
  total?: number
}

export const PRIORITY_LABELS: Record<Priority, string> = {
  low: '低',
  medium: '中',
  high: '高',
  critical: '緊急',
}

export const PRIORITY_COLORS: Record<Priority, string> = {
  low: 'bg-gray-100 text-gray-700',
  medium: 'bg-blue-100 text-blue-700',
  high: 'bg-orange-100 text-orange-700',
  critical: 'bg-red-100 text-red-700',
}
