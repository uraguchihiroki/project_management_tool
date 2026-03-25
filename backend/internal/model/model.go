package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// インプリント（issue_events）の event_type 定数
const (
	EventIssueStatusChanged   = "issue.status_changed"
	EventIssueAssigneeChanged = "issue.assignee_changed"
)

type Organization struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Key        string         `gorm:"size:255;not null" json:"key"`
	Name       string         `gorm:"size:200;not null;index" json:"name"`
	AdminEmail string         `gorm:"size:255" json:"admin_email"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

type SuperAdmin struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Key       string         `gorm:"size:255;not null" json:"key"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	Email     string         `gorm:"size:255;not null;index" json:"email"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Group struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Key            string         `gorm:"size:255;not null" json:"key"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	Organization   Organization   `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Name           string         `gorm:"size:200;not null" json:"name"`
	Order          int            `gorm:"not null;default:0" json:"order"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Group) TableName() string { return "groups" }

type OrganizationUserGroup struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_oug_trip,priority:1" json:"organization_id"`
	UserID         uuid.UUID      `gorm:"type:uuid;not null;index:idx_oug_trip,priority:2" json:"user_id"`
	GroupID        uuid.UUID      `gorm:"type:uuid;not null;index:idx_oug_trip,priority:3" json:"group_id"`
	Key            string         `gorm:"size:255;not null" json:"key"`
	Group          Group          `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (OrganizationUserGroup) TableName() string { return "organization_user_groups" }

type User struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Key            string         `gorm:"size:255;not null" json:"key"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_user_org_email" json:"organization_id"`
	Organization   Organization   `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Name           string         `gorm:"size:100;not null" json:"name"`
	Email          string         `gorm:"size:255;not null;index:idx_user_org_email" json:"email"`
	AvatarURL      *string        `json:"avatar_url,omitempty"`
	IsAdmin        bool           `gorm:"default:false" json:"is_admin"`
	IsOrgAdmin     bool           `gorm:"default:false" json:"is_org_admin"`
	JoinedAt       time.Time      `json:"joined_at"`
	Roles          []Role         `gorm:"many2many:user_roles;joinForeignKey:UserID;joinReferences:RoleID" json:"roles,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type UserRole struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_user_roles_org_pair,priority:1" json:"organization_id"`
	UserID         uuid.UUID      `gorm:"type:uuid;not null;index:idx_user_roles_org_pair,priority:2" json:"user_id"`
	RoleID         uint           `gorm:"not null;index:idx_user_roles_org_pair,priority:3" json:"role_id"`
	Key            string         `gorm:"size:255;not null" json:"key"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (UserRole) TableName() string { return "user_roles" }

type Role struct {
	ID             uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Key            string         `gorm:"size:255;not null" json:"key"`
	Name           string         `gorm:"size:100;not null;index:idx_role_name_org" json:"name"`
	Level          int            `gorm:"not null;default:1" json:"level"`
	Order          int            `gorm:"column:display_order;not null;default:1" json:"-"` // 内部用、非表示
	Description    string         `gorm:"size:500" json:"description"`
	OrganizationID *uuid.UUID     `gorm:"type:uuid;index:idx_role_name_org" json:"organization_id,omitempty"`
	Organization   Organization   `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type Project struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Key               string         `gorm:"size:10;not null;index:idx_project_org_key" json:"key"`
	Name              string         `gorm:"size:200;not null" json:"name"`
	Description       *string        `json:"description,omitempty"`
	OwnerID           uuid.UUID      `gorm:"type:uuid;not null" json:"owner_id"`
	Owner             User           `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	OrganizationID    uuid.UUID      `gorm:"type:uuid;not null;index:idx_project_org_key" json:"organization_id"`
	Organization      Organization   `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Order             int            `gorm:"column:display_order;not null;default:1" json:"-"` // 内部用、非表示
	StartDate         *time.Time     `json:"start_date,omitempty"`
	EndDate           *time.Time     `json:"end_date,omitempty"`
	DefaultWorkflowID *uint          `gorm:"index" json:"default_workflow_id,omitempty"`
	ProjectStatusID   *uuid.UUID     `gorm:"type:uuid;index" json:"project_status_id,omitempty"`
	ProjectStatus     *ProjectStatus `gorm:"foreignKey:ProjectStatusID" json:"project_status,omitempty"`
	Statuses          []Status       `gorm:"-" json:"statuses,omitempty"` // API 応答用（default_workflow の Issue 列）
	CreatedAt         time.Time      `json:"created_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

// ProjectStatus はプロジェクト進行用（Workflow は使わない）
type ProjectStatus struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Key       string         `gorm:"size:255;not null" json:"key"`
	ProjectID uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	Name      string         `gorm:"size:50;not null" json:"name"`
	Color     string         `gorm:"size:7;not null" json:"color"`
	Order     int            `gorm:"not null" json:"order"`
	StatusKey string         `gorm:"size:50;index" json:"status_key,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ProjectStatusTransition は同一プロジェクト内の許可遷移（from → to）。Workflow は使用しない。
type ProjectStatusTransition struct {
	ID                  uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Key                 string         `gorm:"size:255;not null" json:"key"`
	ProjectID           uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	FromProjectStatusID uuid.UUID      `gorm:"type:uuid;not null" json:"from_project_status_id"`
	ToProjectStatusID   uuid.UUID      `gorm:"type:uuid;not null" json:"to_project_status_id"`
	CreatedAt           time.Time      `json:"created_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

// Status は常に1ワークフローに属し、ワークフロー間で共有しない（Issue 専用）
type Status struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Key          string         `gorm:"size:255;not null" json:"key"`
	WorkflowID   uint           `gorm:"not null;index" json:"workflow_id"`
	Workflow     Workflow       `gorm:"foreignKey:WorkflowID" json:"workflow,omitempty"`
	Name         string         `gorm:"size:50;not null" json:"name"`
	Color        string         `gorm:"size:7;not null" json:"color"`
	DisplayOrder int            `gorm:"column:display_order;not null;default:1" json:"display_order"`
	StatusKey    string         `gorm:"size:50;index" json:"status_key,omitempty"` // ユーザー定義 key（空可）
	IsEntry      bool           `gorm:"not null;default:false" json:"is_entry"`
	IsTerminal   bool           `gorm:"not null;default:false" json:"is_terminal"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// WorkflowTransition は同一ワークフロー内の許可遷移（from → to）
type WorkflowTransition struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Key          string         `gorm:"size:255;not null" json:"key"`
	WorkflowID   uint           `gorm:"not null;index" json:"workflow_id"`
	FromStatusID uuid.UUID      `gorm:"type:uuid;not null" json:"from_status_id"`
	ToStatusID   uuid.UUID      `gorm:"type:uuid;not null" json:"to_status_id"`
	DisplayOrder int            `gorm:"column:display_order;not null;default:1" json:"display_order"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TransitionAlertRule は遷移アラート条件（想定外 actor 検知用）
type TransitionAlertRule struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Key             string         `gorm:"size:255;not null" json:"key"`
	OrganizationID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"organization_id"`
	Name            string         `gorm:"size:200;not null" json:"name"`
	FromStatusID    *uuid.UUID     `gorm:"type:uuid" json:"from_status_id,omitempty"`
	ToStatusID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"to_status_id"`
	CreatedAt       time.Time      `json:"created_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type Issue struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Key            string         `gorm:"size:255;not null" json:"key"`
	Number         int            `gorm:"not null" json:"number"`
	Title          string         `gorm:"size:500;not null" json:"title"`
	Description    *string        `json:"description,omitempty"`
	StatusID       uuid.UUID      `gorm:"type:uuid;not null" json:"status_id"`
	Status         Status         `gorm:"foreignKey:StatusID" json:"status,omitempty"`
	Priority       string         `gorm:"size:20;not null;default:'medium'" json:"priority"`
	AssigneeID     *uuid.UUID     `gorm:"type:uuid" json:"assignee_id,omitempty"`
	Assignee       *User          `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	ReporterID     uuid.UUID      `gorm:"type:uuid;not null" json:"reporter_id"`
	Reporter       User           `gorm:"foreignKey:ReporterID" json:"reporter,omitempty"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	ProjectID      *uuid.UUID     `gorm:"type:uuid" json:"project_id,omitempty"`
	DueDate        *time.Time     `json:"due_date,omitempty"`
	TemplateID     *uint          `json:"template_id,omitempty"`
	WorkflowID     uint           `gorm:"not null;index" json:"workflow_id"`
	Comments       []Comment      `gorm:"foreignKey:IssueID" json:"comments,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type IssueTemplate struct {
	ID              uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Key             string         `gorm:"size:255;not null" json:"key"`
	OrganizationID  uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	ProjectID       uuid.UUID      `gorm:"type:uuid;not null" json:"project_id"`
	Project         Project        `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Name            string         `gorm:"size:200;not null" json:"name"`
	Description     string         `gorm:"size:500" json:"description"`
	Body            string         `gorm:"type:text" json:"body"`
	DefaultPriority string         `gorm:"size:20;not null;default:'medium'" json:"default_priority"`
	Order           int            `gorm:"column:display_order;not null;default:1" json:"-"` // 内部用、非表示
	CreatedAt       time.Time      `json:"created_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type Comment struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Key            string         `gorm:"size:255;not null" json:"key"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	IssueID        uuid.UUID      `gorm:"type:uuid;not null" json:"issue_id"`
	AuthorID       uuid.UUID      `gorm:"type:uuid;not null" json:"author_id"`
	Author         User           `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Body           string         `gorm:"not null" json:"body"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type Workflow struct {
	ID             uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Key            string         `gorm:"size:255;not null" json:"key"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	Organization   Organization   `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Name           string         `gorm:"size:200;not null" json:"name"`
	Description    string         `gorm:"size:500" json:"description"`
	Order          int            `gorm:"column:display_order;not null;default:1" json:"-"` // 内部用、非表示
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// IssueEvent は Issue に対する操作事実の1記録（インプリント）。追記のみ。
type IssueEvent struct {
	ID                   uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Key                  string         `gorm:"size:255;not null" json:"key"`
	OrganizationID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"organization_id"`
	IssueID              uuid.UUID      `gorm:"type:uuid;not null;index" json:"issue_id"`
	ActorID              uuid.UUID      `gorm:"type:uuid;not null;index" json:"actor_id"`
	EventType            string         `gorm:"size:80;not null;index" json:"event_type"`
	OccurredAt           time.Time      `gorm:"not null;index" json:"occurred_at"`
	FromStatusID         *uuid.UUID     `gorm:"type:uuid" json:"from_status_id,omitempty"`
	ToStatusID           *uuid.UUID     `gorm:"type:uuid" json:"to_status_id,omitempty"`
	AssigneeIDAtOccurred *uuid.UUID     `gorm:"type:uuid" json:"assignee_id_at_occurred,omitempty"`
	Payload              datatypes.JSON `gorm:"type:json" json:"payload,omitempty"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

func (IssueEvent) TableName() string { return "issue_events" }
