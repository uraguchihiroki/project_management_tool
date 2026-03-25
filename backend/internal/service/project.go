package service

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
	"gorm.io/gorm"
)

var projectStatusColorRegex = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

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
	Name            *string
	Description     *string
	StartDate       *time.Time
	EndDate         *time.Time
	ProjectStatusID *uuid.UUID
}

type ProjectService interface {
	List(orgID *uuid.UUID) ([]model.Project, error)
	Get(id uuid.UUID) (*model.Project, error)
	Create(input CreateProjectInput) (*model.Project, error)
	Update(id uuid.UUID, input UpdateProjectInput) (*model.Project, error)
	Delete(id uuid.UUID) error
	Reorder(orgID *uuid.UUID, ids []uuid.UUID) error
	ListStatusesByOrg(orgID uuid.UUID, statusType string, excludeSystem bool) ([]model.Status, error)
	ListProjectStatuses(projectID uuid.UUID) ([]model.ProjectStatus, error)
	UpdateProjectStatus(projectID, statusID uuid.UUID, name, color string, order int) (*model.ProjectStatus, error)
}

type projectService struct {
	projectRepo repository.ProjectRepository
	statusRepo  repository.StatusRepository
	psRepo      repository.ProjectStatusRepository
	pstRepo     repository.ProjectStatusTransitionRepository
}

func NewProjectService(
	projectRepo repository.ProjectRepository,
	statusRepo repository.StatusRepository,
	psRepo repository.ProjectStatusRepository,
	pstRepo repository.ProjectStatusTransitionRepository,
) ProjectService {
	return &projectService{
		projectRepo: projectRepo,
		statusRepo:  statusRepo,
		psRepo:      psRepo,
		pstRepo:     pstRepo,
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
	if _, err := s.projectRepo.FindByOrgAndKey(input.OrganizationID, input.Key); err == nil {
		return nil, ErrDuplicateProjectKey
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
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

	firstPS, err := SeedDefaultProjectStatuses(s.psRepo, project.ID)
	if err != nil {
		return nil, err
	}
	project.ProjectStatusID = &firstPS

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
	if input.ProjectStatusID != nil {
		newID := *input.ProjectStatusID
		cur := project.ProjectStatusID
		if cur == nil || *cur != newID {
			ps, err := s.psRepo.FindByID(newID)
			if err != nil {
				return nil, fmt.Errorf("project status not found")
			}
			if ps.ProjectID != id {
				return nil, fmt.Errorf("project status does not belong to this project")
			}
			if cur != nil && !s.pstRepo.Exists(id, *cur, newID) {
				return nil, fmt.Errorf("transition not allowed")
			}
			project.ProjectStatusID = &newID
		}
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
	// プロジェクト進行は project_statuses を参照（組織横断の Issue 用 statuses のみ返す）
	if statusType == "project" {
		return []model.Status{}, nil
	}
	if excludeSystem {
		return s.statusRepo.FindByOrganizationIDAndTypeExcludeSystem(orgID, statusType)
	}
	return s.statusRepo.FindByOrganizationIDAndType(orgID, statusType)
}

func (s *projectService) ListProjectStatuses(projectID uuid.UUID) ([]model.ProjectStatus, error) {
	return s.psRepo.FindByProjectID(projectID)
}

func (s *projectService) UpdateProjectStatus(projectID, statusID uuid.UUID, name, color string, order int) (*model.ProjectStatus, error) {
	if name == "" {
		return nil, fmt.Errorf("ステータス名は必須です")
	}
	if len(name) > 50 {
		return nil, fmt.Errorf("ステータス名は50文字以内で指定してください")
	}
	if color == "" || !projectStatusColorRegex.MatchString(color) {
		return nil, fmt.Errorf("色は#RRGGBB形式で指定してください")
	}
	ps, err := s.psRepo.FindByID(statusID)
	if err != nil {
		return nil, fmt.Errorf("project status not found")
	}
	if ps.ProjectID != projectID {
		return nil, fmt.Errorf("project status does not belong to this project")
	}
	if ps.StatusKey == "sts_start" || ps.StatusKey == "sts_goal" {
		return nil, fmt.Errorf("システムステータスは変更できません")
	}
	ps.Name = name
	ps.Color = color
	ps.Order = order
	if err := s.psRepo.Update(ps); err != nil {
		return nil, err
	}
	return s.psRepo.FindByID(statusID)
}
