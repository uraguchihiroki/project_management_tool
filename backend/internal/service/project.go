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
	OrganizationID *uuid.UUID
	StartDate      *time.Time
	EndDate        *time.Time
	Status         string
}

type UpdateProjectInput struct {
	Name        *string
	Description *string
	StartDate   *time.Time
	EndDate     *time.Time
	Status      *string
}

type ProjectService interface {
	List(orgID *uuid.UUID) ([]model.Project, error)
	Get(id uuid.UUID) (*model.Project, error)
	Create(input CreateProjectInput) (*model.Project, error)
	Update(id uuid.UUID, input UpdateProjectInput) (*model.Project, error)
	Delete(id uuid.UUID) error
	Reorder(orgID *uuid.UUID, ids []uuid.UUID) error
	ListStatusesByOrg(orgID uuid.UUID) ([]model.Status, error)
}

type projectService struct {
	projectRepo repository.ProjectRepository
	statusRepo  repository.StatusRepository
}

func NewProjectService(projectRepo repository.ProjectRepository, statusRepo repository.StatusRepository) ProjectService {
	return &projectService{projectRepo: projectRepo, statusRepo: statusRepo}
}

func (s *projectService) List(orgID *uuid.UUID) ([]model.Project, error) {
	if orgID != nil {
		return s.projectRepo.FindByOrg(*orgID)
	}
	return s.projectRepo.FindAll()
}

func (s *projectService) Get(id uuid.UUID) (*model.Project, error) {
	return s.projectRepo.FindByID(id)
}

func (s *projectService) Create(input CreateProjectInput) (*model.Project, error) {
	status := input.Status
	if status == "" {
		status = "none"
	}
	maxOrder, err := s.projectRepo.GetMaxOrder(input.OrganizationID)
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
		Status:         status,
		CreatedAt:      time.Now(),
	}
	if err := s.projectRepo.Create(project); err != nil {
		return nil, err
	}

	// デフォルトステータスを生成
	defaultStatuses := []struct {
		Name  string
		Color string
		Order int
	}{
		{"未着手", "#6B7280", 1},
		{"進行中", "#3B82F6", 2},
		{"レビュー中", "#F59E0B", 3},
		{"完了", "#10B981", 4},
	}
	for _, ds := range defaultStatuses {
		status := &model.Status{
			ID:        uuid.New(),
			ProjectID: &project.ID,
			Name:      ds.Name,
			Color:     ds.Color,
			Order:     ds.Order,
		}
		if err := s.statusRepo.Create(status); err != nil {
			return nil, err
		}
	}

	// ステータスを含めて再取得
	return s.projectRepo.FindByID(project.ID)
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
	if input.Status != nil {
		project.Status = *input.Status
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

func (s *projectService) ListStatusesByOrg(orgID uuid.UUID) ([]model.Status, error) {
	return s.statusRepo.FindByOrganizationID(orgID)
}
