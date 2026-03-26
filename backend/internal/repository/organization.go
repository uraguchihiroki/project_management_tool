package repository

import (
	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

type OrganizationRepository interface {
	FindAll() ([]model.Organization, error)
	FindByID(id uuid.UUID) (*model.Organization, error)
	FindByName(name string) (*model.Organization, error)
	Create(org *model.Organization) error
	FindByUserID(userID uuid.UUID) ([]model.Organization, error)
	FindFirstOrgAdminID(orgID uuid.UUID) (*uuid.UUID, error)
}

type organizationRepository struct {
	db *gorm.DB
}

func NewOrganizationRepository(db *gorm.DB) OrganizationRepository {
	return &organizationRepository{db: db}
}

func (r *organizationRepository) FindAll() ([]model.Organization, error) {
	var orgs []model.Organization
	err := r.db.Order("created_at ASC").Find(&orgs).Error
	return orgs, err
}

func (r *organizationRepository) FindByID(id uuid.UUID) (*model.Organization, error) {
	var org model.Organization
	err := r.db.First(&org, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *organizationRepository) FindByName(name string) (*model.Organization, error) {
	var org model.Organization
	err := r.db.Where("name = ?", name).First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *organizationRepository) Create(org *model.Organization) error {
	return r.db.Create(org).Error
}

func (r *organizationRepository) FindByUserID(userID uuid.UUID) ([]model.Organization, error) {
	var user model.User
	err := r.db.Preload("Organization").First(&user, "id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	if user.Organization.ID == (uuid.UUID{}) {
		return []model.Organization{}, nil
	}
	return []model.Organization{user.Organization}, nil
}

func (r *organizationRepository) FindFirstOrgAdminID(orgID uuid.UUID) (*uuid.UUID, error) {
	var user model.User
	err := r.db.Where("organization_id = ? AND is_org_admin = ?", orgID, true).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user.ID, nil
}
