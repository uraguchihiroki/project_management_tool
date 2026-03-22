package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

type CreateProjectInput struct {
	Key            string
	Name           string
	Description    *string
	OwnerID        uuid.UUID
	OrganizationID uuid.UUID
	StartDate      *time.Time
	EndDate        *time.Time
}

type UpdateProjectInput struct {
	Name        *string
	Description *string
	StartDate   *time.Time
	EndDate     *time.Time
}

type ProjectService interface {
	List(orgID *uuid.UUID) ([]model.Project, error)
	Get(id uuid.UUID) (*model.Project, error)
	Create(input CreateProjectInput) (*model.Project, error)
	Update(id uuid.UUID, input UpdateProjectInput) (*model.Project, error)
	Delete(id uuid.UUID) error
	Reorder(orgID *uuid.UUID, ids []uuid.UUID) error
	ListStatusesByOrg(orgID uuid.UUID, statusType string, excludeSystem bool) ([]model.Status, error)
}

type projectService struct {
	projectRepo    repository.ProjectRepository
	statusRepo     repository.StatusRepository
	workflowRepo   repository.WorkflowRepository
	transitionRepo repository.WorkflowTransitionRepository
}

func NewProjectService(
	projectRepo repository.ProjectRepository,
	statusRepo repository.StatusRepository,
	workflowRepo repository.WorkflowRepository,
	transitionRepo repository.WorkflowTransitionRepository,
) ProjectService {
	return &projectService{
		projectRepo:    projectRepo,
		statusRepo:     statusRepo,
		workflowRepo:   workflowRepo,
		transitionRepo: transitionRepo,
	}
}

func (s *projectService) List(orgID *uuid.UUID) ([]model.Project, error) {
	if orgID != nil {
		return s.projectRepo.FindByOrg(*orgID)
	}
	return s.projectRepo.FindAll()
}

func (s *projectService) Get(id uuid.UUID) (*model.Project, error) {
	p, err := s.projectRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if p.DefaultWorkflowID != nil {
		sts, err := s.statusRepo.FindByWorkflowID(*p.DefaultWorkflowID)
		if err != nil {
			return nil, err
		}
		p.Statuses = sts
	}
	return p, nil
}

func (s *projectService) Create(input CreateProjectInput) (*model.Project, error) {
	orgID := &input.OrganizationID
	maxOrder, err := s.projectRepo.GetMaxOrder(orgID)
	if err != nil {
		return nil, err
	}
	project := &model.Project{
		ID:             uuid.New(),
		Key:            input.Key,
		Name:           input.Name,
		Description:    input.Description,
		OwnerID:        input.OwnerID,
		OrganizationID: input.OrganizationID,
		Order:          maxOrder + 1,
		StartDate:      input.StartDate,
		EndDate:        input.EndDate,
		CreatedAt:      time.Now(),
	}
	if err := s.projectRepo.Create(project); err != nil {
		return nil, err
	}

	wfName := input.Name + " - Issue"
	wfID, _, err := CreateWorkflowWithIssueStatuses(s.workflowRepo, s.statusRepo, s.transitionRepo, input.OrganizationID, wfName)
	if err != nil {
		return nil, err
	}
	project.DefaultWorkflowID = &wfID
	if err := s.projectRepo.Update(project); err != nil {
		return nil, err
	}

	return s.Get(project.ID)
}

func (s *projectService) Update(id uuid.UUID, input UpdateProjectInput) (*model.Project, error) {
	project, err := s.projectRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		project.Name = *input.Name
	}
	if input.Description != nil {
		project.Description = input.Description
	}
	if input.StartDate != nil {
		project.StartDate = input.StartDate
	}
	if input.EndDate != nil {
		project.EndDate = input.EndDate
	}
	if err := s.projectRepo.Update(project); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *projectService) Delete(id uuid.UUID) error {
	return s.projectRepo.Delete(id)
}

func (s *projectService) Reorder(orgID *uuid.UUID, ids []uuid.UUID) error {
	return s.projectRepo.Reorder(orgID, ids)
}

func (s *projectService) ListStatusesByOrg(orgID uuid.UUID, statusType string, excludeSystem bool) ([]model.Status, error) {
	if excludeSystem {
		return s.statusRepo.FindByOrganizationIDAndTypeExcludeSystem(orgID, statusType)
	}
	return s.statusRepo.FindByOrganizationIDAndType(orgID, statusType)
}
