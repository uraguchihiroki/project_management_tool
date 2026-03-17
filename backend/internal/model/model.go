package model

import (
	"time"

	"github.com/google/uuid"
)

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
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Level       int       `gorm:"not null;default:1" json:"level"`
	Description string    `gorm:"size:500" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Project struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Key         string     `gorm:"size:10;uniqueIndex;not null" json:"key"`
	Name        string     `gorm:"size:200;not null" json:"name"`
	Description *string    `json:"description,omitempty"`
	OwnerID     uuid.UUID  `gorm:"type:uuid;not null" json:"owner_id"`
	Owner       User       `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Statuses    []Status   `gorm:"foreignKey:ProjectID" json:"statuses,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type Status struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null" json:"project_id"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	Color     string    `gorm:"size:7;not null" json:"color"`
	Order     int       `gorm:"not null" json:"order"`
}

type Issue struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Number      int        `gorm:"not null" json:"number"`
	Title       string     `gorm:"size:500;not null" json:"title"`
	Description *string    `json:"description,omitempty"`
	StatusID    uuid.UUID  `gorm:"type:uuid;not null" json:"status_id"`
	Status      Status     `gorm:"foreignKey:StatusID" json:"status,omitempty"`
	Priority    string     `gorm:"size:20;not null;default:'medium'" json:"priority"`
	AssigneeID  *uuid.UUID `gorm:"type:uuid" json:"assignee_id,omitempty"`
	Assignee    *User      `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	ReporterID  uuid.UUID  `gorm:"type:uuid;not null" json:"reporter_id"`
	Reporter    User       `gorm:"foreignKey:ReporterID" json:"reporter,omitempty"`
	ProjectID   uuid.UUID  `gorm:"type:uuid;not null" json:"project_id"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	TemplateID  *uint      `json:"template_id,omitempty"`
	WorkflowID  *uint      `json:"workflow_id,omitempty"`
	Comments    []Comment  `gorm:"foreignKey:IssueID" json:"comments,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
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
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID   uuid.UUID      `gorm:"type:uuid;not null" json:"project_id"`
	Project     Project        `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Name        string         `gorm:"size:200;not null" json:"name"`
	Description string         `gorm:"size:500" json:"description"`
	Steps       []WorkflowStep `gorm:"foreignKey:WorkflowID" json:"steps,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

type WorkflowStep struct {
	ID            uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	WorkflowID    uint       `gorm:"not null" json:"workflow_id"`
	Order         int        `gorm:"not null;default:1" json:"order"`
	Name          string     `gorm:"size:200;not null" json:"name"`
	RequiredLevel int        `gorm:"not null;default:1" json:"required_level"`
	StatusID      *uuid.UUID `gorm:"type:uuid" json:"status_id,omitempty"`
	Status        *Status    `gorm:"foreignKey:StatusID" json:"status,omitempty"`
}
