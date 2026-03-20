package service

import (
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
	issueRepo      repository.IssueRepository
}

func NewOrgSeedService(
	orgRepo repository.OrganizationRepository,
	statusRepo repository.StatusRepository,
	roleRepo repository.RoleRepository,
	projectRepo repository.ProjectRepository,
	departmentRepo repository.DepartmentRepository,
	issueRepo repository.IssueRepository,
) OrgSeedService {
	return &orgSeedService{
		orgRepo:        orgRepo,
		statusRepo:     statusRepo,
		roleRepo:       roleRepo,
		projectRepo:    projectRepo,
		departmentRepo: departmentRepo,
		issueRepo:      issueRepo,
	}
}

// SeedNewOrganization は新規組織にステータス・役職・サンプルプロジェクトを投入する
// 既存レコードは更新、なければ作成。開発中に何度でも実行可能（冪等）
// ownerID が nil の場合はサンプルプロジェクトは作成しない
func (s *orgSeedService) SeedNewOrganization(orgID uuid.UUID, ownerID *uuid.UUID) error {
	// 1. 組織用ステータス（Issue用: カンバン列）
	issueStatuses := []struct {
		Name  string
		Color string
		Order int
	}{
		{"未着手", "#6B7280", 1},
		{"進行中", "#3B82F6", 2},
		{"レビュー中", "#F59E0B", 3},
		{"完了", "#10B981", 4},
	}
	for _, st := range issueStatuses {
		if err := s.upsertOrgStatus(orgID, nil, st.Name, "issue", st.Color, st.Order); err != nil {
			return err
		}
	}

	// 2. 組織用ステータス（Project用: ライフサイクル）
	projectStatuses := []struct {
		Name  string
		Color string
		Order int
	}{
		{"計画中", "#6B7280", 1},
		{"進行中", "#3B82F6", 2},
		{"完了", "#10B981", 3},
	}
	for _, st := range projectStatuses {
		if err := s.upsertOrgStatus(orgID, nil, st.Name, "project", st.Color, st.Order); err != nil {
			return err
		}
	}

	// 3. 役職
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

	// 4. 部署
	departments := []string{"開発部", "営業部", "管理部"}
	for i, name := range departments {
		if err := s.upsertDepartment(orgID, name, i+1); err != nil {
			return err
		}
	}

	// 5. サンプルプロジェクト（owner がいる場合のみ）
	if ownerID != nil {
		project, err := s.upsertSampleProject(orgID, *ownerID)
		if err != nil {
			return err
		}
		if project != nil {
			// 6. サンプルプロジェクト用のIssueステータス
			issueStatusesForProject := []struct {
				Name  string
				Color string
				Order int
			}{
				{"未着手", "#6B7280", 1},
				{"進行中", "#3B82F6", 2},
				{"レビュー中", "#F59E0B", 3},
				{"完了", "#10B981", 4},
			}
			for _, st := range issueStatusesForProject {
				if err := s.upsertOrgStatus(orgID, &project.ID, st.Name, "issue", st.Color, st.Order); err != nil {
					return err
				}
			}
			// 7. サンプルIssue（プロジェクトに1件もない場合のみ作成）
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

func (s *orgSeedService) upsertOrgStatus(orgID uuid.UUID, projectID *uuid.UUID, name, statusType, color string, order int) error {
	existing, err := s.statusRepo.FindByOrgNameType(orgID, projectID, name, statusType)
	if err == nil {
		existing.Color = color
		existing.Order = order
		return s.statusRepo.Update(existing)
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	status := &model.Status{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      name,
		Color:     color,
		Order:     order,
		Type:      statusType,
	}
	if projectID == nil {
		status.OrganizationID = &orgID
	}
	return s.statusRepo.Create(status)
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
	return s.roleRepo.Create(&model.Role{
		Name:           name,
		Level:          level,
		OrganizationID: &orgID,
		CreatedAt:      time.Now(),
	})
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
	return s.departmentRepo.Create(&model.Department{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           name,
		Order:          maxOrder + 1,
		CreatedAt:      time.Now(),
	})
}

func (s *orgSeedService) upsertSampleIssue(orgID uuid.UUID, projectID uuid.UUID, reporterID uuid.UUID) error {
	issues, err := s.issueRepo.FindByProject(projectID)
	if err != nil || len(issues) > 0 {
		return err
	}
	statuses, err := s.statusRepo.FindByProject(projectID)
	if err != nil || len(statuses) == 0 {
		return err
	}
	number, _ := s.issueRepo.NextNumber(projectID)
	now := time.Now()
	return s.issueRepo.Create(&model.Issue{
		ID:             uuid.New(),
		Number:         number,
		Title:          "サンプルチケット",
		Description:    strPtr("動作確認用のサンプルです。"),
		StatusID:       statuses[0].ID,
		Priority:       "medium",
		ReporterID:     reporterID,
		OrganizationID: orgID,
		ProjectID:      &projectID,
		CreatedAt:      now,
		UpdatedAt:      now,
	})
}

func strPtr(s string) *string {
	return &s
}
