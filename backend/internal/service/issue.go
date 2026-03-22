package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/pkg/keygen"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CreateIssueInput struct {
	Title       string
	Description *string
	StatusID    uuid.UUID
	Priority    string
	AssigneeID  *uuid.UUID
	ReporterID  uuid.UUID
	TemplateID  *uint
	WorkflowID  *uint
}

type UpdateIssueInput struct {
	Title       *string
	Description *string
	StatusID    *uuid.UUID
	Priority    *string
	AssigneeID  *uuid.UUID
}

type IssueService interface {
	List(projectID uuid.UUID) ([]model.Issue, error)
	ListByOrg(orgID uuid.UUID) ([]model.Issue, error)
	Get(projectID uuid.UUID, number int) (*model.Issue, error)
	GetByOrgAndNumber(orgID uuid.UUID, number int) (*model.Issue, error)
	Create(projectID uuid.UUID, input CreateIssueInput) (*model.Issue, error)
	CreateForOrg(orgID uuid.UUID, input CreateIssueInput) (*model.Issue, error)
	Update(projectID uuid.UUID, number int, input UpdateIssueInput, actorID uuid.UUID) (*model.Issue, error)
	UpdateByOrgAndNumber(orgID uuid.UUID, number int, input UpdateIssueInput, actorID uuid.UUID) (*model.Issue, error)
	Delete(projectID uuid.UUID, number int) error
	DeleteByOrgAndNumber(orgID uuid.UUID, number int) error
}

type issueService struct {
	issueRepo   repository.IssueRepository
	projectRepo repository.ProjectRepository
	eventRepo   repository.IssueEventRepository
}

func NewIssueService(issueRepo repository.IssueRepository, projectRepo repository.ProjectRepository, eventRepo repository.IssueEventRepository) IssueService {
	return &issueService{issueRepo: issueRepo, projectRepo: projectRepo, eventRepo: eventRepo}
}

func assigneePtrEqual(a, b *uuid.UUID) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func uuidPtrStr(p *uuid.UUID) *string {
	if p == nil {
		return nil
	}
	s := p.String()
	return &s
}

func (s *issueService) List(projectID uuid.UUID) ([]model.Issue, error) {
	return s.issueRepo.FindByProject(projectID)
}

func (s *issueService) ListByOrg(orgID uuid.UUID) ([]model.Issue, error) {
	return s.issueRepo.FindByOrg(orgID)
}

func (s *issueService) Get(projectID uuid.UUID, number int) (*model.Issue, error) {
	return s.issueRepo.FindByNumber(projectID, number)
}

func (s *issueService) GetByOrgAndNumber(orgID uuid.UUID, number int) (*model.Issue, error) {
	return s.issueRepo.FindByOrgAndNumber(orgID, number)
}

func (s *issueService) Create(projectID uuid.UUID, input CreateIssueInput) (*model.Issue, error) {
	project, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return nil, err
	}
	orgID := project.OrganizationID
	// 採番
	nextNum, err := s.issueRepo.NextNumber(projectID)
	if err != nil {
		return nil, err
	}

	// デフォルト優先度
	priority := input.Priority
	if priority == "" {
		priority = "medium"
	}

	issueID := uuid.New()
	key := fmt.Sprintf("%s-%d", project.Key, nextNum)
	issue := &model.Issue{
		ID:             issueID,
		Key:            key,
		Number:         nextNum,
		Title:          input.Title,
		Description:    input.Description,
		StatusID:       input.StatusID,
		Priority:       priority,
		AssigneeID:     input.AssigneeID,
		ReporterID:     input.ReporterID,
		OrganizationID: orgID,
		ProjectID:      &projectID,
		TemplateID:     input.TemplateID,
		WorkflowID:     input.WorkflowID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := s.issueRepo.Create(issue); err != nil {
		return nil, err
	}
	// アソシエーションを含めて再取得
	return s.issueRepo.FindByNumber(projectID, issue.Number)
}

func (s *issueService) CreateForOrg(orgID uuid.UUID, input CreateIssueInput) (*model.Issue, error) {
	nextNum, err := s.issueRepo.NextNumberForOrg(orgID)
	if err != nil {
		return nil, err
	}

	priority := input.Priority
	if priority == "" {
		priority = "medium"
	}

	issueID := uuid.New()
	issue := &model.Issue{
		ID:             issueID,
		Key:            keygen.UUIDKey(issueID),
		Number:         nextNum,
		Title:          input.Title,
		Description:    input.Description,
		StatusID:       input.StatusID,
		Priority:       priority,
		AssigneeID:     input.AssigneeID,
		ReporterID:     input.ReporterID,
		OrganizationID: orgID,
		ProjectID:      nil,
		TemplateID:     input.TemplateID,
		WorkflowID:     input.WorkflowID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := s.issueRepo.Create(issue); err != nil {
		return nil, err
	}
	return s.issueRepo.FindByOrgAndNumber(orgID, issue.Number)
}

func (s *issueService) Update(projectID uuid.UUID, number int, input UpdateIssueInput, actorID uuid.UUID) (*model.Issue, error) {
	var out *model.Issue
	err := s.issueRepo.DB().Transaction(func(tx *gorm.DB) error {
		ir := repository.NewIssueRepository(tx)
		er := repository.NewIssueEventRepository(tx)
		issue, err := ir.FindByNumber(projectID, number)
		if err != nil {
			return err
		}
		oldStatus := issue.StatusID
		oldAssignee := issue.AssigneeID
		if input.Title != nil {
			issue.Title = *input.Title
		}
		if input.Description != nil {
			issue.Description = input.Description
		}
		if input.StatusID != nil {
			issue.StatusID = *input.StatusID
		}
		if input.Priority != nil {
			issue.Priority = *input.Priority
		}
		if input.AssigneeID != nil {
			issue.AssigneeID = input.AssigneeID
		}
		issue.UpdatedAt = time.Now()
		statusChanged := oldStatus != issue.StatusID
		assigneeChanged := !assigneePtrEqual(oldAssignee, issue.AssigneeID)
		if err := ir.Update(issue); err != nil {
			return err
		}
		now := time.Now().UTC()
		if statusChanged {
			ev := newStatusImprint(issue, actorID, oldStatus, issue.StatusID, issue.AssigneeID, now)
			if err := er.Create(ev); err != nil {
				return err
			}
		}
		if assigneeChanged {
			ev, err := newAssigneeImprint(issue, actorID, oldAssignee, issue.AssigneeID, now)
			if err != nil {
				return err
			}
			if err := er.Create(ev); err != nil {
				return err
			}
		}
		out, err = ir.FindByNumber(projectID, number)
		return err
	})
	return out, err
}

func (s *issueService) UpdateByOrgAndNumber(orgID uuid.UUID, number int, input UpdateIssueInput, actorID uuid.UUID) (*model.Issue, error) {
	var out *model.Issue
	err := s.issueRepo.DB().Transaction(func(tx *gorm.DB) error {
		ir := repository.NewIssueRepository(tx)
		er := repository.NewIssueEventRepository(tx)
		issue, err := ir.FindByOrgAndNumber(orgID, number)
		if err != nil {
			return err
		}
		oldStatus := issue.StatusID
		oldAssignee := issue.AssigneeID
		if input.Title != nil {
			issue.Title = *input.Title
		}
		if input.Description != nil {
			issue.Description = input.Description
		}
		if input.StatusID != nil {
			issue.StatusID = *input.StatusID
		}
		if input.Priority != nil {
			issue.Priority = *input.Priority
		}
		if input.AssigneeID != nil {
			issue.AssigneeID = input.AssigneeID
		}
		issue.UpdatedAt = time.Now()
		statusChanged := oldStatus != issue.StatusID
		assigneeChanged := !assigneePtrEqual(oldAssignee, issue.AssigneeID)
		if err := ir.Update(issue); err != nil {
			return err
		}
		now := time.Now().UTC()
		if statusChanged {
			ev := newStatusImprint(issue, actorID, oldStatus, issue.StatusID, issue.AssigneeID, now)
			if err := er.Create(ev); err != nil {
				return err
			}
		}
		if assigneeChanged {
			ev, err := newAssigneeImprint(issue, actorID, oldAssignee, issue.AssigneeID, now)
			if err != nil {
				return err
			}
			if err := er.Create(ev); err != nil {
				return err
			}
		}
		out, err = ir.FindByOrgAndNumber(orgID, number)
		return err
	})
	return out, err
}

func newStatusImprint(issue *model.Issue, actorID, from, to uuid.UUID, assigneeSnap *uuid.UUID, at time.Time) *model.IssueEvent {
	id := uuid.New()
	return &model.IssueEvent{
		ID:                   id,
		Key:                  keygen.UUIDKey(id),
		OrganizationID:       issue.OrganizationID,
		IssueID:              issue.ID,
		ActorID:              actorID,
		EventType:            model.EventIssueStatusChanged,
		OccurredAt:           at,
		FromStatusID:         &from,
		ToStatusID:           &to,
		AssigneeIDAtOccurred: assigneeSnap,
	}
}

func newAssigneeImprint(issue *model.Issue, actorID uuid.UUID, from, to *uuid.UUID, at time.Time) (*model.IssueEvent, error) {
	id := uuid.New()
	p := map[string]interface{}{
		"from_assignee_id": uuidPtrStr(from),
		"to_assignee_id":   uuidPtrStr(to),
	}
	raw, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return &model.IssueEvent{
		ID:             id,
		Key:            keygen.UUIDKey(id),
		OrganizationID: issue.OrganizationID,
		IssueID:        issue.ID,
		ActorID:        actorID,
		EventType:      model.EventIssueAssigneeChanged,
		OccurredAt:     at,
		Payload:        datatypes.JSON(raw),
	}, nil
}

func (s *issueService) Delete(projectID uuid.UUID, number int) error {
	issue, err := s.issueRepo.FindByNumber(projectID, number)
	if err != nil {
		return err
	}
	return s.issueRepo.Delete(issue.ID)
}

func (s *issueService) DeleteByOrgAndNumber(orgID uuid.UUID, number int) error {
	issue, err := s.issueRepo.FindByOrgAndNumber(orgID, number)
	if err != nil {
		return err
	}
	return s.issueRepo.Delete(issue.ID)
}
