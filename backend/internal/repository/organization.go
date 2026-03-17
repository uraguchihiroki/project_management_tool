package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

type OrganizationRepository interface {
	FindAll() ([]model.Organization, error)
	FindByID(id uuid.UUID) (*model.Organization, error)
	Create(org *model.Organization) error
	FindByUserID(userID uuid.UUID) ([]model.Organization, error)
	AddUser(orgUser *model.OrganizationUser) error
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

func (r *organizationRepository) Create(org *model.Organization) error {
	return r.db.Create(org).Error
}

func (r *organizationRepository) FindByUserID(userID uuid.UUID) ([]model.Organization, error) {
	var orgUsers []model.OrganizationUser
	err := r.db.
		Preload("Organization").
		Where("user_id = ?", userID).
		Find(&orgUsers).Error
	if err != nil {
		return nil, err
	}
	orgs := make([]model.Organization, 0, len(orgUsers))
	for _, ou := range orgUsers {
		orgs = append(orgs, ou.Organization)
	}
	return orgs, nil
}

func (r *organizationRepository) AddUser(orgUser *model.OrganizationUser) error {
	orgUser.JoinedAt = time.Now()
	return r.db.Where("organization_id = ? AND user_id = ?", orgUser.OrganizationID, orgUser.UserID).
		FirstOrCreate(orgUser).Error
}
