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
	GroupIDs    []uuid.UUID
}

type UpdateIssueInput struct {
	Title       *string
	Description *string
	StatusID    *uuid.UUID
	Priority    *string
	AssigneeID  *uuid.UUID
	GroupIDs    *[]uuid.UUID
}

type IssueService interface {
	List(projectID uuid.UUID, groupID *uuid.UUID) ([]model.Issue, error)
	ListByOrg(orgID uuid.UUID, groupID *uuid.UUID) ([]model.Issue, error)
	Get(projectID uuid.UUID, number int) (*model.Issue, error)
	GetByOrgAndNumber(orgID uuid.UUID, number int) (*model.Issue, error)
	GetWithGroups(projectID uuid.UUID, number int) (*model.Issue, []model.Group, error)
	GetByOrgAndNumberWithGroups(orgID uuid.UUID, number int) (*model.Issue, []model.Group, error)
	Create(projectID uuid.UUID, input CreateIssueInput) (*model.Issue, error)
	CreateForOrg(orgID uuid.UUID, input CreateIssueInput) (*model.Issue, error)
	Update(projectID uuid.UUID, number int, input UpdateIssueInput, actorID uuid.UUID) (*model.Issue, error)
	UpdateByOrgAndNumber(orgID uuid.UUID, number int, input UpdateIssueInput, actorID uuid.UUID) (*model.Issue, error)
	UpdateStatusWithImprint(issueID uuid.UUID, newStatusID uuid.UUID, actorID uuid.UUID) error
	SetIssueGroups(projectID uuid.UUID, number int, groupIDs []uuid.UUID) error
	SetIssueGroupsByOrg(orgID uuid.UUID, number int, groupIDs []uuid.UUID) error
	Delete(projectID uuid.UUID, number int) error
	DeleteByOrgAndNumber(orgID uuid.UUID, number int) error
}

type issueService struct {
	issueRepo      repository.IssueRepository
	projectRepo    repository.ProjectRepository
	statusRepo     repository.StatusRepository
	workflowRepo   repository.WorkflowRepository
	transitionRepo repository.WorkflowTransitionRepository
	eventRepo      repository.IssueEventRepository
	groupRepo      repository.GroupRepository
	issueGroupRepo repository.IssueGroupRepository
	alertEval      *TransitionAlertEvaluator
}

func NewIssueService(
	issueRepo repository.IssueRepository,
	projectRepo repository.ProjectRepository,
	statusRepo repository.StatusRepository,
	workflowRepo repository.WorkflowRepository,
	transitionRepo repository.WorkflowTransitionRepository,
	eventRepo repository.IssueEventRepository,
	groupRepo repository.GroupRepository,
	issueGroupRepo repository.IssueGroupRepository,
	alertEval *TransitionAlertEvaluator,
) IssueService {
	return &issueService{
		issueRepo:      issueRepo,
		projectRepo:    projectRepo,
		statusRepo:     statusRepo,
		workflowRepo:   workflowRepo,
		transitionRepo: transitionRepo,
		eventRepo:      eventRepo,
		groupRepo:      groupRepo,
		issueGroupRepo: issueGroupRepo,
		alertEval:      alertEval,
	}
}

func (s *issueService) workflowIDForNewIssue(project *model.Project, orgID uuid.UUID, statusID uuid.UUID) (uint, error) {
	st, err := s.statusRepo.FindByID(statusID)
	if err != nil {
		return 0, fmt.Errorf("status: %w", err)
	}
	if project != nil {
		if project.DefaultWorkflowID == nil {
			return 0, fmt.Errorf("project has no default workflow")
		}
		if st.WorkflowID != *project.DefaultWorkflowID {
			return 0, fmt.Errorf("status does not belong to project workflow")
		}
		return st.WorkflowID, nil
	}
	wf, err := s.workflowRepo.FindByOrgAndName(orgID, "組織Issue")
	if err != nil {
		return 0, fmt.Errorf("組織Issue workflow: %w", err)
	}
	if st.WorkflowID != wf.ID {
		return 0, fmt.Errorf("status does not belong to organization issue workflow")
	}
	return st.WorkflowID, nil
}

func (s *issueService) validateStatusChange(tx *gorm.DB, issue *model.Issue, newStatusID uuid.UUID) error {
	if issue.StatusID == newStatusID {
		return nil
	}
	sr := repository.NewStatusRepository(tx)
	tr := repository.NewWorkflowTransitionRepository(tx)
	newSt, err := sr.FindByID(newStatusID)
	if err != nil {
		return fmt.Errorf("status: %w", err)
	}
	if newSt.WorkflowID != issue.WorkflowID {
		return fmt.Errorf("status not in issue workflow")
	}
	if !tr.Exists(issue.WorkflowID, issue.StatusID, newStatusID) {
		return fmt.Errorf("transition not allowed")
	}
	return nil
}

func (s *issueService) validateGroupIDsForOrg(orgID uuid.UUID, ids []uuid.UUID) error {
	for _, id := range ids {
		g, err := s.groupRepo.FindByID(id)
		if err != nil {
			return fmt.Errorf("group not found: %w", err)
		}
		if g.OrganizationID != orgID {
			return fmt.Errorf("group %s does not belong to organization", id)
		}
	}
	return nil
}

func (s *issueService) List(projectID uuid.UUID, groupID *uuid.UUID) ([]model.Issue, error) {
	return s.issueRepo.FindByProject(projectID, groupID)
}

func (s *issueService) ListByOrg(orgID uuid.UUID, groupID *uuid.UUID) ([]model.Issue, error) {
	return s.issueRepo.FindByOrg(orgID, groupID)
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

func (s *issueService) Get(projectID uuid.UUID, number int) (*model.Issue, error) {
	return s.issueRepo.FindByNumber(projectID, number)
}

func (s *issueService) GetByOrgAndNumber(orgID uuid.UUID, number int) (*model.Issue, error) {
	return s.issueRepo.FindByOrgAndNumber(orgID, number)
}

func (s *issueService) GetWithGroups(projectID uuid.UUID, number int) (*model.Issue, []model.Group, error) {
	issue, err := s.issueRepo.FindByNumber(projectID, number)
	if err != nil {
		return nil, nil, err
	}
	groups, err := s.issueGroupRepo.ListGroupsByIssue(issue.ID)
	if err != nil {
		return issue, nil, err
	}
	return issue, groups, nil
}

func (s *issueService) GetByOrgAndNumberWithGroups(orgID uuid.UUID, number int) (*model.Issue, []model.Group, error) {
	issue, err := s.issueRepo.FindByOrgAndNumber(orgID, number)
	if err != nil {
		return nil, nil, err
	}
	groups, err := s.issueGroupRepo.ListGroupsByIssue(issue.ID)
	if err != nil {
		return issue, nil, err
	}
	return issue, groups, nil
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

	wfID, err := s.workflowIDForNewIssue(project, orgID, input.StatusID)
	if err != nil {
		return nil, err
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
		WorkflowID:     wfID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := s.issueRepo.Create(issue); err != nil {
		return nil, err
	}
	if len(input.GroupIDs) > 0 {
		if err := s.validateGroupIDsForOrg(orgID, input.GroupIDs); err != nil {
			return nil, err
		}
		if err := s.issueGroupRepo.ReplaceForIssue(issue.ID, input.GroupIDs); err != nil {
			return nil, err
		}
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

	wfID, err := s.workflowIDForNewIssue(nil, orgID, input.StatusID)
	if err != nil {
		return nil, err
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
		WorkflowID:     wfID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := s.issueRepo.Create(issue); err != nil {
		return nil, err
	}
	if len(input.GroupIDs) > 0 {
		if err := s.validateGroupIDsForOrg(orgID, input.GroupIDs); err != nil {
			return nil, err
		}
		if err := s.issueGroupRepo.ReplaceForIssue(issue.ID, input.GroupIDs); err != nil {
			return nil, err
		}
	}
	return s.issueRepo.FindByOrgAndNumber(orgID, issue.Number)
}

func (s *issueService) Update(projectID uuid.UUID, number int, input UpdateIssueInput, actorID uuid.UUID) (*model.Issue, error) {
	var out *model.Issue
	var oldStatus uuid.UUID
	var statusChanged bool
	err := s.issueRepo.DB().Transaction(func(tx *gorm.DB) error {
		ir := repository.NewIssueRepository(tx)
		er := repository.NewIssueEventRepository(tx)
		igr := repository.NewIssueGroupRepository(tx)
		issue, err := ir.FindByNumber(projectID, number)
		if err != nil {
			return err
		}
		oldStatus = issue.StatusID
		oldAssignee := issue.AssigneeID
		if input.Title != nil {
			issue.Title = *input.Title
		}
		if input.Description != nil {
			issue.Description = input.Description
		}
		if input.StatusID != nil {
			if err := s.validateStatusChange(tx, issue, *input.StatusID); err != nil {
				return err
			}
			issue.StatusID = *input.StatusID
		}
		if input.Priority != nil {
			issue.Priority = *input.Priority
		}
		if input.AssigneeID != nil {
			issue.AssigneeID = input.AssigneeID
		}
		issue.UpdatedAt = time.Now()
		statusChanged = oldStatus != issue.StatusID
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
		if input.GroupIDs != nil {
			if err := s.validateGroupIDsForOrg(issue.OrganizationID, *input.GroupIDs); err != nil {
				return err
			}
			if err := igr.ReplaceForIssue(issue.ID, *input.GroupIDs); err != nil {
				return err
			}
		}
		out, err = ir.FindByNumber(projectID, number)
		return err
	})
	if err == nil && statusChanged && s.alertEval != nil {
		s.alertEval.OnStatusChanged(out, oldStatus, out.StatusID, actorID)
	}
	return out, err
}

func (s *issueService) UpdateByOrgAndNumber(orgID uuid.UUID, number int, input UpdateIssueInput, actorID uuid.UUID) (*model.Issue, error) {
	var out *model.Issue
	var oldStatus uuid.UUID
	var statusChanged bool
	err := s.issueRepo.DB().Transaction(func(tx *gorm.DB) error {
		ir := repository.NewIssueRepository(tx)
		er := repository.NewIssueEventRepository(tx)
		igr := repository.NewIssueGroupRepository(tx)
		issue, err := ir.FindByOrgAndNumber(orgID, number)
		if err != nil {
			return err
		}
		oldStatus = issue.StatusID
		oldAssignee := issue.AssigneeID
		if input.Title != nil {
			issue.Title = *input.Title
		}
		if input.Description != nil {
			issue.Description = input.Description
		}
		if input.StatusID != nil {
			if err := s.validateStatusChange(tx, issue, *input.StatusID); err != nil {
				return err
			}
			issue.StatusID = *input.StatusID
		}
		if input.Priority != nil {
			issue.Priority = *input.Priority
		}
		if input.AssigneeID != nil {
			issue.AssigneeID = input.AssigneeID
		}
		issue.UpdatedAt = time.Now()
		statusChanged = oldStatus != issue.StatusID
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
		if input.GroupIDs != nil {
			if err := s.validateGroupIDsForOrg(issue.OrganizationID, *input.GroupIDs); err != nil {
				return err
			}
			if err := igr.ReplaceForIssue(issue.ID, *input.GroupIDs); err != nil {
				return err
			}
		}
		out, err = ir.FindByOrgAndNumber(orgID, number)
		return err
	})
	if err == nil && statusChanged && s.alertEval != nil {
		s.alertEval.OnStatusChanged(out, oldStatus, out.StatusID, actorID)
	}
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

func (s *issueService) SetIssueGroups(projectID uuid.UUID, number int, groupIDs []uuid.UUID) error {
	issue, err := s.issueRepo.FindByNumber(projectID, number)
	if err != nil {
		return err
	}
	if err := s.validateGroupIDsForOrg(issue.OrganizationID, groupIDs); err != nil {
		return err
	}
	return s.issueGroupRepo.ReplaceForIssue(issue.ID, groupIDs)
}

func (s *issueService) SetIssueGroupsByOrg(orgID uuid.UUID, number int, groupIDs []uuid.UUID) error {
	issue, err := s.issueRepo.FindByOrgAndNumber(orgID, number)
	if err != nil {
		return err
	}
	if err := s.validateGroupIDsForOrg(issue.OrganizationID, groupIDs); err != nil {
		return err
	}
	return s.issueGroupRepo.ReplaceForIssue(issue.ID, groupIDs)
}

func (s *issueService) UpdateStatusWithImprint(issueID uuid.UUID, newStatusID uuid.UUID, actorID uuid.UUID) error {
	var changed bool
	var oldSt uuid.UUID
	err := s.issueRepo.DB().Transaction(func(tx *gorm.DB) error {
		ir := repository.NewIssueRepository(tx)
		er := repository.NewIssueEventRepository(tx)
		issue, err := ir.FindByID(issueID)
		if err != nil {
			return err
		}
		if issue.StatusID == newStatusID {
			return nil
		}
		if err := s.validateStatusChange(tx, issue, newStatusID); err != nil {
			return err
		}
		changed = true
		oldSt = issue.StatusID
		issue.StatusID = newStatusID
		issue.UpdatedAt = time.Now()
		if err := ir.Update(issue); err != nil {
			return err
		}
		ev := newStatusImprint(issue, actorID, oldSt, newStatusID, issue.AssigneeID, time.Now().UTC())
		return er.Create(ev)
	})
	if err != nil {
		return err
	}
	if !changed || s.alertEval == nil {
		return nil
	}
	full, err := s.issueRepo.FindByID(issueID)
	if err != nil {
		return nil
	}
	s.alertEval.OnStatusChanged(full, oldSt, full.StatusID, actorID)
	return nil
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
