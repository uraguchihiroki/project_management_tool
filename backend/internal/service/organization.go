package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/pkg/keygen"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

type OrganizationService interface {
	List() ([]model.Organization, error)
	Get(id uuid.UUID) (*model.Organization, error)
	Create(name, adminEmail, adminName string) (*model.Organization, error)
	GetUserOrganizations(userID uuid.UUID) ([]model.Organization, error)
	AddUser(orgID uuid.UUID, existingUserID uuid.UUID, isOrgAdmin bool) (*model.User, error)
}

type organizationService struct {
	orgRepo    repository.OrganizationRepository
	userRepo   repository.UserRepository
	orgSeedSvc OrgSeedService
}

func NewOrganizationService(orgRepo repository.OrganizationRepository, userRepo repository.UserRepository, orgSeedSvc OrgSeedService) OrganizationService {
	return &organizationService{orgRepo: orgRepo, userRepo: userRepo, orgSeedSvc: orgSeedSvc}
}

func (s *organizationService) List() ([]model.Organization, error) {
	return s.orgRepo.FindAll()
}

func (s *organizationService) Get(id uuid.UUID) (*model.Organization, error) {
	return s.orgRepo.FindByID(id)
}

func (s *organizationService) Create(name, adminEmail, adminName string) (*model.Organization, error) {
	orgID := uuid.New()
	key := keygen.Slug(name)
	if key == "" {
		key = keygen.UUIDKey(orgID)
	}
	org := &model.Organization{
		ID:         orgID,
		Key:        key,
		Name:       name,
		AdminEmail: adminEmail,
		CreatedAt:  time.Now(),
	}
	if err := s.orgRepo.Create(org); err != nil {
		return nil, err
	}
	var ownerID *uuid.UUID
	if adminEmail != "" {
		// 同一メールでも組織が違えば別ユーザー。新組織用に新規ユーザーを作成
		userID := uuid.New()
		user := &model.User{
			ID:             userID,
			Key:            adminEmail,
			OrganizationID: org.ID,
			Name:           adminName,
			Email:          adminEmail,
			IsAdmin:        true,
			IsOrgAdmin:     true,
			JoinedAt:       time.Now(),
			CreatedAt:      time.Now(),
		}
		if adminName == "" {
			user.Name = adminEmail
		}
		if err := s.userRepo.Create(user); err != nil {
			return org, err
		}
		ownerID = &user.ID
	}
	// 組織作成時に初期データ（ステータス・役職・サンプルプロジェクト）を投入
	if err := s.orgSeedSvc.SeedNewOrganization(org.ID, ownerID); err != nil {
		return org, err // 組織は作成済みなのでエラーは返さず org を返すことも検討可。現状はエラーを伝播
	}
	return org, nil
}

func (s *organizationService) GetUserOrganizations(userID uuid.UUID) ([]model.Organization, error) {
	u, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.userRepo.FindAllByEmail(u.Email)
	if err != nil {
		return nil, err
	}
	orgs := make([]model.Organization, 0, len(rows))
	seen := make(map[uuid.UUID]struct{})
	for _, usr := range rows {
		if usr.Organization.ID == uuid.Nil {
			continue
		}
		if _, ok := seen[usr.Organization.ID]; ok {
			continue
		}
		seen[usr.Organization.ID] = struct{}{}
		orgs = append(orgs, usr.Organization)
	}
	return orgs, nil
}

func (s *organizationService) AddUser(orgID uuid.UUID, existingUserID uuid.UUID, isOrgAdmin bool) (*model.User, error) {
	existing, err := s.userRepo.FindByID(existingUserID)
	if err != nil {
		return nil, err
	}
	// 冪等: 既に同じemailのユーザーが組織にいればそれを返す
	if u, err := s.userRepo.FindByEmailAndOrg(orgID, existing.Email); err == nil && u != nil {
		return u, nil
	}
	// 同一人物を別組織に追加 = 新規ユーザーを作成（同じ name, email、異なる organization_id）
	newUserID := uuid.New()
	user := &model.User{
		ID:             newUserID,
		Key:            existing.Email,
		OrganizationID: orgID,
		Name:           existing.Name,
		Email:          existing.Email,
		IsOrgAdmin:     isOrgAdmin,
		JoinedAt:       time.Now(),
		CreatedAt:      time.Now(),
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}
