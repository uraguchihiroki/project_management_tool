package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
	"gorm.io/gorm"
)

// OrgSeedService は組織作成時に初期データを投入する
type OrgSeedService interface {
	SeedNewOrganization(orgID uuid.UUID, ownerID *uuid.UUID) error
	SeedAllOrganizations() error
}

type orgSeedService struct {
	orgRepo        repository.OrganizationRepository
	statusRepo     repository.StatusRepository
	roleRepo       repository.RoleRepository
	projectRepo    repository.ProjectRepository
	departmentRepo repository.DepartmentRepository
	issueRepo    repository.IssueRepository
	workflowRepo repository.WorkflowRepository
	psRepo       repository.ProjectStatusRepository
}

func NewOrgSeedService(
	orgRepo repository.OrganizationRepository,
	statusRepo repository.StatusRepository,
	roleRepo repository.RoleRepository,
	projectRepo repository.ProjectRepository,
	departmentRepo repository.DepartmentRepository,
	issueRepo repository.IssueRepository,
	workflowRepo repository.WorkflowRepository,
	psRepo repository.ProjectStatusRepository,
) OrgSeedService {
	return &orgSeedService{
		orgRepo:        orgRepo,
		statusRepo:     statusRepo,
		roleRepo:       roleRepo,
		projectRepo:    projectRepo,
		departmentRepo: departmentRepo,
		issueRepo:      issueRepo,
		workflowRepo:   workflowRepo,
		psRepo:         psRepo,
	}
}

func (s *orgSeedService) ensureOrgIssueWorkflow(orgID uuid.UUID) error {
	_, err := s.workflowRepo.FindByOrgAndName(orgID, "組織Issue")
	if err == nil {
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	_, _, err = CreateWorkflowWithIssueStatuses(s.workflowRepo, s.statusRepo, orgID, "組織Issue")
	return err
}

func (s *orgSeedService) ensureProjectDefaultWorkflow(project *model.Project) error {
	if project.DefaultWorkflowID != nil {
		return nil
	}
	wfID, _, err := CreateWorkflowWithIssueStatuses(
		s.workflowRepo, s.statusRepo,
		project.OrganizationID,
		project.Name+" - Issue",
	)
	if err != nil {
		return err
	}
	project.DefaultWorkflowID = &wfID
	return s.projectRepo.Update(project)
}

func (s *orgSeedService) ensureProjectStatuses(project *model.Project) error {
	if project.ProjectStatusID != nil {
		return nil
	}
	firstID, err := SeedDefaultProjectStatuses(s.psRepo, project.ID)
	if err != nil {
		return err
	}
	project.ProjectStatusID = &firstID
	return s.projectRepo.Update(project)
}

// SeedNewOrganization は新規組織にステータス・役職・サンプルプロジェクトを投入する
func (s *orgSeedService) SeedNewOrganization(orgID uuid.UUID, ownerID *uuid.UUID) error {
	if err := s.ensureOrgIssueWorkflow(orgID); err != nil {
		return err
	}

	roles := []struct {
		Name  string
		Level int
	}{
		{"部長", 4},
		{"課長", 3},
		{"主任", 2},
		{"メンバー", 1},
	}
	for _, r := range roles {
		if err := s.upsertRole(orgID, r.Name, r.Level); err != nil {
			return err
		}
	}

	departments := []string{"開発部", "営業部", "管理部"}
	for i, name := range departments {
		if err := s.upsertDepartment(orgID, name, i+1); err != nil {
			return err
		}
	}

	if ownerID != nil {
		project, err := s.upsertSampleProject(orgID, *ownerID)
		if err != nil {
			return err
		}
		if project != nil {
			if err := s.ensureProjectDefaultWorkflow(project); err != nil {
				return err
			}
			if err := s.ensureProjectStatuses(project); err != nil {
				return err
			}
			if err := s.upsertSampleIssue(orgID, project.ID, *ownerID); err != nil {
				return err
			}
		}
	}

	return nil
}

// SeedAllOrganizations は全組織にseedデータを投入する
func (s *orgSeedService) SeedAllOrganizations() error {
	orgs, err := s.orgRepo.FindAll()
	if err != nil {
		return err
	}
	for _, org := range orgs {
		ownerID, _ := s.orgRepo.FindFirstOrgAdminID(org.ID)
		if err := s.SeedNewOrganization(org.ID, ownerID); err != nil {
			return err
		}
	}
	return nil
}

func (s *orgSeedService) upsertRole(orgID uuid.UUID, name string, level int) error {
	existing, err := s.roleRepo.FindByOrgAndName(orgID, name)
	if err == nil {
		existing.Level = level
		return s.roleRepo.Update(existing)
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	role := &model.Role{
		Name:           name,
		Level:          level,
		OrganizationID: &orgID,
		CreatedAt:      time.Now(),
	}
	if err := s.roleRepo.Create(role); err != nil {
		return err
	}
	role.Key = fmt.Sprintf("role-%d", role.ID)
	return s.roleRepo.Update(role)
}

func (s *orgSeedService) upsertSampleProject(orgID uuid.UUID, ownerID uuid.UUID) (*model.Project, error) {
	existing, err := s.projectRepo.FindByOrgAndName(orgID, "サンプルプロジェクト")
	if err == nil {
		existing.OwnerID = ownerID
		return existing, s.projectRepo.Update(existing)
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	orgIDPtr := &orgID
	keySuffix := strings.ReplaceAll(orgID.String(), "-", "")[:6]
	projectKey := "DEMO" + strings.ToUpper(keySuffix)
	maxOrder, _ := s.projectRepo.GetMaxOrder(orgIDPtr)
	project := &model.Project{
		ID:             uuid.New(),
		Key:            projectKey,
		Name:           "サンプルプロジェクト",
		OwnerID:        ownerID,
		OrganizationID: orgID,
		Order:          maxOrder + 1,
		CreatedAt:      time.Now(),
	}
	if err := s.projectRepo.Create(project); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *orgSeedService) upsertDepartment(orgID uuid.UUID, name string, _ int) error {
	_, err := s.departmentRepo.FindByOrgAndName(orgID, name)
	if err == nil {
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	maxOrder, _ := s.departmentRepo.GetMaxOrder(orgID)
	deptID := uuid.New()
	key := strings.ReplaceAll(strings.ToLower(name), " ", "-")
	if key == "" {
		key = deptID.String()
	}
	return s.departmentRepo.Create(&model.Department{
		ID:             deptID,
		Key:            key,
		OrganizationID: orgID,
		Name:           name,
		Order:          maxOrder + 1,
		CreatedAt:      time.Now(),
	})
}

func (s *orgSeedService) upsertSampleIssue(orgID uuid.UUID, projectID uuid.UUID, reporterID uuid.UUID) error {
	issues, err := s.issueRepo.FindByProject(projectID, nil)
	if err != nil || len(issues) > 0 {
		return err
	}
	project, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return err
	}
	if err := s.ensureProjectDefaultWorkflow(project); err != nil {
		return err
	}
	if err := s.ensureProjectStatuses(project); err != nil {
		return err
	}
	project, err = s.projectRepo.FindByID(projectID)
	if err != nil {
		return err
	}
	statuses, err := s.statusRepo.FindByProject(projectID)
	if err != nil || len(statuses) == 0 {
		return err
	}
	if project.DefaultWorkflowID == nil {
		return fmt.Errorf("project has no default workflow")
	}
	number, _ := s.issueRepo.NextNumber(projectID)
	now := time.Now()
	issueKey := uuid.New().String()
	if project != nil {
		issueKey = fmt.Sprintf("%s-%d", project.Key, number)
	}
	return s.issueRepo.Create(&model.Issue{
		ID:             uuid.New(),
		Key:            issueKey,
		Number:         number,
		Title:          "サンプルチケット",
		Description:    strPtr("動作確認用のサンプルです。"),
		StatusID:       statuses[0].ID,
		Priority:       "medium",
		ReporterID:     reporterID,
		OrganizationID: orgID,
		ProjectID:      &projectID,
		WorkflowID:     *project.DefaultWorkflowID,
		CreatedAt:      now,
		UpdatedAt:      now,
	})
}

func strPtr(s string) *string {
	return &s
}
