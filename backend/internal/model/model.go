package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Organization struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Key        string         `gorm:"size:255;not null" json:"key"`
	Name       string         `gorm:"size:200;not null;uniqueIndex" json:"name"`
	AdminEmail string         `gorm:"size:255" json:"admin_email"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

type SuperAdmin struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Key       string         `gorm:"size:255;not null" json:"key"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	Email     string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Department struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Key            string         `gorm:"size:255;not null" json:"key"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	Organization   Organization   `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Name           string         `gorm:"size:200;not null" json:"name"`
	Order          int            `gorm:"not null;default:0" json:"order"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type OrganizationUserDepartment struct {
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null;primaryKey" json:"organization_id"`
	UserID        uuid.UUID       `gorm:"type:uuid;not null;primaryKey" json:"user_id"`
	DepartmentID  uuid.UUID       `gorm:"type:uuid;not null;primaryKey" json:"department_id"`
	Key           string         `gorm:"size:255;not null" json:"key"`
	Department    Department     `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type User struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Key            string         `gorm:"size:255;not null" json:"key"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_user_org_email" json:"organization_id"`
	Organization   Organization   `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Name           string         `gorm:"size:100;not null" json:"name"`
	Email          string         `gorm:"size:255;not null;uniqueIndex:idx_user_org_email" json:"email"`
	AvatarURL      *string        `json:"avatar_url,omitempty"`
	IsAdmin        bool           `gorm:"default:false" json:"is_admin"`
	IsOrgAdmin     bool           `gorm:"default:false" json:"is_org_admin"`
	JoinedAt       time.Time      `json:"joined_at"`
	Roles          []Role         `gorm:"many2many:user_roles;joinForeignKey:UserID;joinReferences:RoleID" json:"roles,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type UserRole struct {
	UserID uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"user_id"`
	RoleID uint     `gorm:"not null;primaryKey" json:"role_id"`
	Key    string   `gorm:"size:255;not null" json:"key"`
}

func (UserRole) TableName() string { return "user_roles" }

type Role struct {
	ID             uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Key            string         `gorm:"size:255;not null" json:"key"`
	Name           string         `gorm:"size:100;not null;uniqueIndex:idx_role_name_org" json:"name"`
	Level          int            `gorm:"not null;default:1" json:"level"`
	Order          int            `gorm:"column:display_order;not null;default:1" json:"-"` // 内部用、非表示
	Description    string         `gorm:"size:500" json:"description"`
	OrganizationID *uuid.UUID     `gorm:"type:uuid;uniqueIndex:idx_role_name_org" json:"organization_id,omitempty"`
	Organization   Organization   `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type Project struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Key            string         `gorm:"size:10;not null;uniqueIndex:idx_project_org_key" json:"key"`
	Name           string         `gorm:"size:200;not null" json:"name"`
	Description    *string        `json:"description,omitempty"`
	OwnerID        uuid.UUID      `gorm:"type:uuid;not null" json:"owner_id"`
	Owner          User           `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_project_org_key" json:"organization_id"`
	Organization   Organization   `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Order          int            `gorm:"column:display_order;not null;default:1" json:"-"` // 内部用、非表示
	StartDate      *time.Time     `json:"start_date,omitempty"`
	EndDate        *time.Time     `json:"end_date,omitempty"`
	Statuses       []Status       `gorm:"foreignKey:ProjectID" json:"statuses,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type Status struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Key            string         `gorm:"size:255;not null" json:"key"`
	ProjectID      *uuid.UUID     `gorm:"type:uuid" json:"project_id,omitempty"`
	OrganizationID *uuid.UUID     `gorm:"type:uuid" json:"organization_id,omitempty"`
	Name           string         `gorm:"size:50;not null" json:"name"`
	Color          string         `gorm:"size:7;not null" json:"color"`
	Order          int            `gorm:"not null" json:"order"`
	Type           string         `gorm:"size:20;not null;default:'issue'" json:"type"` // issue | project
	StatusKey      string         `gorm:"size:50;index" json:"status_key,omitempty"` // sts_start, sts_goal。空=ユーザー定義
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
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
	WorkflowID     *uint          `json:"workflow_id,omitempty"`
	Comments       []Comment      `gorm:"foreignKey:IssueID" json:"comments,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type IssueApproval struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Key            string         `gorm:"size:255;not null" json:"key"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	IssueID        uuid.UUID      `gorm:"type:uuid;not null" json:"issue_id"`
	WorkflowStepID uint           `gorm:"not null" json:"workflow_step_id"`
	WorkflowStep   WorkflowStep   `gorm:"foreignKey:WorkflowStepID" json:"workflow_step,omitempty"`
	ApproverID     *uuid.UUID     `gorm:"type:uuid" json:"approver_id,omitempty"`
	Approver       *User          `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`
	Status         string         `gorm:"size:20;not null;default:'pending'" json:"status"`
	Comment        string         `gorm:"type:text" json:"comment"`
	ActedAt        *time.Time     `json:"acted_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
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
	WorkflowID      *uint          `json:"workflow_id,omitempty"`
	Workflow        *Workflow      `gorm:"foreignKey:WorkflowID" json:"workflow,omitempty"`
	Order           int            `gorm:"column:display_order;not null;default:1" json:"-"` // 内部用、非表示
	CreatedAt       time.Time      `json:"created_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type Comment struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Key            string         `gorm:"size:255;not null" json:"key"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	IssueID       uuid.UUID      `gorm:"type:uuid;not null" json:"issue_id"`
	AuthorID      uuid.UUID      `gorm:"type:uuid;not null" json:"author_id"`
	Author        User           `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Body          string         `gorm:"not null" json:"body"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type Workflow struct {
	ID             uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Key            string         `gorm:"size:255;not null" json:"key"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	Organization   Organization   `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Name           string         `gorm:"size:200;not null" json:"name"`
	Description    string         `gorm:"size:500" json:"description"`
	Order          int            `gorm:"column:display_order;not null;default:1" json:"-"` // 内部用、非表示
	Steps          []WorkflowStep `gorm:"foreignKey:WorkflowID" json:"steps,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type WorkflowStep struct {
	ID               uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Key              string     `gorm:"size:255;not null" json:"key"`
	OrganizationID   uuid.UUID  `gorm:"type:uuid;not null" json:"organization_id"`
	WorkflowID       uint       `gorm:"not null" json:"workflow_id"`
	StatusID         uuid.UUID  `gorm:"type:uuid;not null" json:"status_id"` // このステップのステータス。表示名は Status.Name
	Status           *Status    `gorm:"foreignKey:StatusID" json:"status,omitempty"`
	NextStatusID     *uuid.UUID `gorm:"type:uuid" json:"next_status_id,omitempty"` // 承認後ステータス。ゴールでは NULL
	NextStatus       *Status    `gorm:"foreignKey:NextStatusID" json:"next_status,omitempty"`
	Description      string     `gorm:"type:text" json:"description"`
	Threshold        int        `gorm:"not null;default:10" json:"threshold"`
	ApprovalObjects  []ApprovalObject `gorm:"foreignKey:WorkflowStepID" json:"approval_objects,omitempty"`
	ExcludeReporter  bool       `gorm:"default:false" json:"exclude_reporter"`
	ExcludeAssignee  bool       `gorm:"default:false" json:"exclude_assignee"`
	// 後方互換（Order は status チェーンで辿るため未使用だが既存データ用に残す）
	Order            int        `gorm:"not null;default:1" json:"order"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

type ApprovalObject struct {
	ID               uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Key              string     `gorm:"size:255;not null" json:"key"`
	OrganizationID   uuid.UUID  `gorm:"type:uuid;not null" json:"organization_id"`
	WorkflowStepID   uint       `gorm:"not null" json:"workflow_step_id"`
	Order          int        `gorm:"column:sort_order;not null;default:1" json:"order"`
	Type            string     `gorm:"size:20;not null" json:"type"` // role / user
	RoleID          *uint      `json:"role_id,omitempty"`
	Role            *Role      `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	RoleOperator    string     `gorm:"size:10" json:"role_operator,omitempty"` // eq / gte
	UserID          *uuid.UUID `gorm:"type:uuid" json:"user_id,omitempty"`
	User            *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Points          int            `gorm:"not null;default:1" json:"points"`
	ExcludeReporter bool           `gorm:"default:false" json:"exclude_reporter"`
	ExcludeAssignee bool           `gorm:"default:false" json:"exclude_assignee"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`
}
