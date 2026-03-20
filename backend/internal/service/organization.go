package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

type OrganizationService interface {
	List() ([]model.Organization, error)
	Get(id uuid.UUID) (*model.Organization, error)
	Create(name, adminEmail, adminName string) (*model.Organization, error)
	GetUserOrganizations(userID uuid.UUID) ([]model.Organization, error)
	AddUser(orgID, userID uuid.UUID, isOrgAdmin bool) error
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
	org := &model.Organization{
		ID:         uuid.New(),
		Name:       name,
		AdminEmail: adminEmail,
		CreatedAt:  time.Now(),
	}
	if err := s.orgRepo.Create(org); err != nil {
		return nil, err
	}
	var ownerID *uuid.UUID
	if adminEmail != "" {
		user, err := s.userRepo.FindByEmail(adminEmail)
		if err != nil {
			user = &model.User{
				ID:        uuid.New(),
				Name:      adminName,
				Email:     adminEmail,
				IsAdmin:   true,
				CreatedAt: time.Now(),
			}
			if adminName == "" {
				user.Name = adminEmail
			}
			if err := s.userRepo.Create(user); err != nil {
				return org, err
			}
		} else {
			_ = s.userRepo.UpdateAdmin(user.ID, true)
		}
		ownerID = &user.ID
		_ = s.orgRepo.AddUser(&model.OrganizationUser{
			OrganizationID: org.ID,
			UserID:         user.ID,
			IsOrgAdmin:     true,
		})
	}
	// 組織作成時に初期データ（ステータス・役職・サンプルプロジェクト）を投入
	if err := s.orgSeedSvc.SeedNewOrganization(org.ID, ownerID); err != nil {
		return org, err // 組織は作成済みなのでエラーは返さず org を返すことも検討可。現状はエラーを伝播
	}
	return org, nil
}

func (s *organizationService) GetUserOrganizations(userID uuid.UUID) ([]model.Organization, error) {
	return s.orgRepo.FindByUserID(userID)
}

func (s *organizationService) AddUser(orgID, userID uuid.UUID, isOrgAdmin bool) error {
	orgUser := &model.OrganizationUser{
		OrganizationID: orgID,
		UserID:         userID,
		IsOrgAdmin:     isOrgAdmin,
	}
	return s.orgRepo.AddUser(orgUser)
}
