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
  statuses?: Status[]
  created_at: string
}

export interface Status {
  id: string
  project_id?: string
  organization_id?: string
  name: string
  color: string
  order: number
  type?: 'issue' | 'project'
}

export type Priority = 'low' | 'medium' | 'high' | 'critical'

export interface IssueApproval {
  id: string
  issue_id: string
  workflow_step_id: number
  workflow_step: WorkflowStep
  approver_id?: string
  approver?: User
  status: 'pending' | 'approved' | 'rejected'
  comment: string
  acted_at?: string
  created_at: string
}

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

export interface ApprovalObject {
  id: number
  workflow_step_id: number
  order: number
  type: 'role' | 'user'
  role_id?: number
  role?: Role
  role_operator?: 'eq' | 'gte'
  user_id?: string
  user?: User
  points: number
  exclude_reporter: boolean
  exclude_assignee: boolean
}

export interface WorkflowStep {
  id: number
  workflow_id: number
  order: number
  step_type?: 'start' | 'normal' | 'goal'
  name: string
  description?: string
  threshold?: number
  status_id?: string
  status?: Status
  approval_objects?: ApprovalObject[]
  required_level?: number
  approver_type?: 'role' | 'user' | 'multiple'
  approver_user_id?: string
  min_approvers?: number
  exclude_reporter?: boolean
  exclude_assignee?: boolean
}

export interface Workflow {
  id: number
  name: string
  description: string
  steps?: WorkflowStep[]
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
