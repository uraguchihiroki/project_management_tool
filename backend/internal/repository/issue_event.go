package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

// IssueEventFilters は組織横断一覧のクエリ用
type IssueEventFilters struct {
	EventType      *string
	FromOccurredAt *time.Time
	ToOccurredAt   *time.Time
	ActorID        *uuid.UUID
	IssueID        *uuid.UUID
}

type IssueEventRepository interface {
	Create(ev *model.IssueEvent) error
	ListByIssueID(issueID uuid.UUID) ([]model.IssueEvent, error)
	ListByOrganization(orgID uuid.UUID, f IssueEventFilters) ([]model.IssueEvent, error)
}

type issueEventRepository struct {
	db *gorm.DB
}

func NewIssueEventRepository(db *gorm.DB) IssueEventRepository {
	return &issueEventRepository{db: db}
}

func (r *issueEventRepository) Create(ev *model.IssueEvent) error {
	return r.db.Create(ev).Error
}

func (r *issueEventRepository) ListByIssueID(issueID uuid.UUID) ([]model.IssueEvent, error) {
	var out []model.IssueEvent
	err := r.db.
		Where("issue_id = ?", issueID).
		Order("occurred_at ASC").
		Find(&out).Error
	return out, err
}

func (r *issueEventRepository) ListByOrganization(orgID uuid.UUID, f IssueEventFilters) ([]model.IssueEvent, error) {
	q := r.db.Where("organization_id = ?", orgID)
	if f.EventType != nil && *f.EventType != "" {
		q = q.Where("event_type = ?", *f.EventType)
	}
	if f.FromOccurredAt != nil {
		q = q.Where("occurred_at >= ?", *f.FromOccurredAt)
	}
	if f.ToOccurredAt != nil {
		q = q.Where("occurred_at <= ?", *f.ToOccurredAt)
	}
	if f.ActorID != nil {
		q = q.Where("actor_id = ?", *f.ActorID)
	}
	if f.IssueID != nil {
		q = q.Where("issue_id = ?", *f.IssueID)
	}
	var out []model.IssueEvent
	err := q.
		Order("occurred_at DESC").
		Find(&out).Error
	return out, err
}
