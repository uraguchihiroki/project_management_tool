package model

import (
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name       string    `gorm:"size:200;not null;uniqueIndex" json:"name"`
	AdminEmail string    `gorm:"size:255" json:"admin_email"`
	CreatedAt  time.Time `json:"created_at"`
}

type SuperAdmin struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Email     string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type OrganizationUser struct {
	OrganizationID uuid.UUID    `gorm:"type:uuid;not null;primaryKey" json:"organization_id"`
	Organization   Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	UserID         uuid.UUID    `gorm:"type:uuid;not null;primaryKey" json:"user_id"`
	User           User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	IsOrgAdmin     bool         `gorm:"default:false" json:"is_org_admin"`
	JoinedAt       time.Time    `json:"joined_at"`
}

type Department struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null" json:"organization_id"`
	Organization   Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Name           string    `gorm:"size:200;not null" json:"name"`
	Order          int       `gorm:"not null;default:0" json:"order"`
	CreatedAt      time.Time `json:"created_at"`
}

type OrganizationUserDepartment struct {
	OrganizationID uuid.UUID    `gorm:"type:uuid;not null;primaryKey" json:"organization_id"`
	UserID         uuid.UUID    `gorm:"type:uuid;not null;primaryKey" json:"user_id"`
	DepartmentID   uuid.UUID    `gorm:"type:uuid;not null;primaryKey" json:"department_id"`
	Department     Department   `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
}

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Email     string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	IsAdmin   bool      `gorm:"default:false" json:"is_admin"`
	Roles     []Role    `gorm:"many2many:user_roles;" json:"roles,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Role struct {
	ID             uint         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name           string       `gorm:"size:100;not null;uniqueIndex:idx_role_name_org" json:"name"`
	Level          int          `gorm:"not null;default:1" json:"level"`
	Order          int          `gorm:"column:display_order;not null;default:1" json:"-"` // 内部用、非表示
	Description    string       `gorm:"size:500" json:"description"`
	OrganizationID *uuid.UUID   `gorm:"type:uuid;uniqueIndex:idx_role_name_org" json:"organization_id,omitempty"`
	Organization   Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
}

type Project struct {
	ID             uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	Key            string       `gorm:"size:10;uniqueIndex;not null" json:"key"`
	Name           string       `gorm:"size:200;not null" json:"name"`
	Description    *string      `json:"description,omitempty"`
	OwnerID        uuid.UUID    `gorm:"type:uuid;not null" json:"owner_id"`
	Owner          User         `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	OrganizationID *uuid.UUID   `gorm:"type:uuid" json:"organization_id,omitempty"`
	Organization   Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Order          int          `gorm:"column:display_order;not null;default:1" json:"-"` // 内部用、非表示
	StartDate      *time.Time   `json:"start_date,omitempty"`
	EndDate        *time.Time   `json:"end_date,omitempty"`
	Status         string       `gorm:"size:20;not null;default:'none'" json:"status"`
	Statuses       []Status     `gorm:"foreignKey:ProjectID" json:"statuses,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
}

type Status struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID      *uuid.UUID `gorm:"type:uuid" json:"project_id,omitempty"`
	OrganizationID *uuid.UUID `gorm:"type:uuid" json:"organization_id,omitempty"`
	Name           string     `gorm:"size:50;not null" json:"name"`
	Color          string     `gorm:"size:7;not null" json:"color"`
	Order          int        `gorm:"not null" json:"order"`
	Type           string     `gorm:"size:20;not null;default:'issue'" json:"type"` // issue | project
}

type Issue struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Number         int        `gorm:"not null" json:"number"`
	Title          string     `gorm:"size:500;not null" json:"title"`
	Description    *string    `json:"description,omitempty"`
	StatusID       uuid.UUID  `gorm:"type:uuid;not null" json:"status_id"`
	Status         Status     `gorm:"foreignKey:StatusID" json:"status,omitempty"`
	Priority       string     `gorm:"size:20;not null;default:'medium'" json:"priority"`
	AssigneeID     *uuid.UUID `gorm:"type:uuid" json:"assignee_id,omitempty"`
	Assignee       *User      `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	ReporterID     uuid.UUID  `gorm:"type:uuid;not null" json:"reporter_id"`
	Reporter       User       `gorm:"foreignKey:ReporterID" json:"reporter,omitempty"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null" json:"organization_id"`
	ProjectID      *uuid.UUID `gorm:"type:uuid" json:"project_id,omitempty"`
	DueDate        *time.Time `json:"due_date,omitempty"`
	TemplateID     *uint      `json:"template_id,omitempty"`
	WorkflowID     *uint      `json:"workflow_id,omitempty"`
	Comments       []Comment  `gorm:"foreignKey:IssueID" json:"comments,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type IssueApproval struct {
	ID             uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	IssueID        uuid.UUID    `gorm:"type:uuid;not null" json:"issue_id"`
	WorkflowStepID uint         `gorm:"not null" json:"workflow_step_id"`
	WorkflowStep   WorkflowStep `gorm:"foreignKey:WorkflowStepID" json:"workflow_step,omitempty"`
	ApproverID     *uuid.UUID   `gorm:"type:uuid" json:"approver_id,omitempty"`
	Approver       *User        `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`
	Status         string       `gorm:"size:20;not null;default:'pending'" json:"status"`
	Comment        string       `gorm:"type:text" json:"comment"`
	ActedAt        *time.Time   `json:"acted_at,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
}

type IssueTemplate struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID       uuid.UUID `gorm:"type:uuid;not null" json:"project_id"`
	Project         Project   `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Name            string    `gorm:"size:200;not null" json:"name"`
	Description     string    `gorm:"size:500" json:"description"`
	Body            string    `gorm:"type:text" json:"body"`
	DefaultPriority string    `gorm:"size:20;not null;default:'medium'" json:"default_priority"`
	WorkflowID      *uint     `json:"workflow_id,omitempty"`
	Workflow        *Workflow `gorm:"foreignKey:WorkflowID" json:"workflow,omitempty"`
	Order           int       `gorm:"column:display_order;not null;default:1" json:"-"` // 内部用、非表示
	CreatedAt       time.Time `json:"created_at"`
}

type Comment struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	IssueID   uuid.UUID `gorm:"type:uuid;not null" json:"issue_id"`
	AuthorID  uuid.UUID `gorm:"type:uuid;not null" json:"author_id"`
	Author    User      `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Body      string    `gorm:"not null" json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Workflow struct {
	ID             uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	Organization   Organization   `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Name           string         `gorm:"size:200;not null" json:"name"`
	Description    string         `gorm:"size:500" json:"description"`
	Order          int            `gorm:"column:display_order;not null;default:1" json:"-"` // 内部用、非表示
	Steps          []WorkflowStep `gorm:"foreignKey:WorkflowID" json:"steps,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
}

type WorkflowStep struct {
	ID              uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	WorkflowID      uint       `gorm:"not null" json:"workflow_id"`
	Order           int        `gorm:"not null;default:1" json:"order"`
	Name            string     `gorm:"size:200;not null" json:"name"`
	RequiredLevel   int        `gorm:"not null;default:1" json:"required_level"`
	StatusID        *uuid.UUID `gorm:"type:uuid" json:"status_id,omitempty"`
	Status          *Status    `gorm:"foreignKey:StatusID" json:"status,omitempty"`
	ApproverType    string     `gorm:"size:20;not null;default:'role'" json:"approver_type"` // role / user / multiple
	ApproverUserID  *uuid.UUID `gorm:"type:uuid" json:"approver_user_id,omitempty"`
	MinApprovers    int        `gorm:"not null;default:1" json:"min_approvers"`
	ExcludeReporter bool       `gorm:"default:false" json:"exclude_reporter"`
	ExcludeAssignee bool       `gorm:"default:false" json:"exclude_assignee"`
}
