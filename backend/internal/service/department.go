package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

type DepartmentService interface {
	ListByOrganization(orgID uuid.UUID) ([]model.Department, error)
	Get(id uuid.UUID) (*model.Department, error)
	Create(orgID uuid.UUID, name string) (*model.Department, error)
	Update(id uuid.UUID, name string) (*model.Department, error)
	Delete(id uuid.UUID) error
	Reorder(orgID uuid.UUID, ids []uuid.UUID) error
	GetUserDepartments(orgID, userID uuid.UUID) ([]model.Department, error)
	SetUserDepartments(orgID, userID uuid.UUID, departmentIDs []uuid.UUID) error
}

type departmentService struct {
	deptRepo repository.DepartmentRepository
	orgRepo  repository.OrganizationRepository
}

func NewDepartmentService(deptRepo repository.DepartmentRepository, orgRepo repository.OrganizationRepository) DepartmentService {
	return &departmentService{deptRepo: deptRepo, orgRepo: orgRepo}
}

func (s *departmentService) ListByOrganization(orgID uuid.UUID) ([]model.Department, error) {
	return s.deptRepo.FindByOrganizationID(orgID)
}

func (s *departmentService) Get(id uuid.UUID) (*model.Department, error) {
	return s.deptRepo.FindByID(id)
}

func (s *departmentService) Create(orgID uuid.UUID, name string) (*model.Department, error) {
	maxOrder, err := s.deptRepo.GetMaxOrder(orgID)
	if err != nil {
		return nil, err
	}
	d := &model.Department{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           name,
		Order:          maxOrder + 1,
		CreatedAt:      time.Now(),
	}
	if err := s.deptRepo.Create(d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *departmentService) Update(id uuid.UUID, name string) (*model.Department, error) {
	d, err := s.deptRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		d.Name = name
	}
	if err := s.deptRepo.Update(d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *departmentService) Reorder(orgID uuid.UUID, ids []uuid.UUID) error {
	return s.deptRepo.Reorder(orgID, ids)
}

func (s *departmentService) Delete(id uuid.UUID) error {
	return s.deptRepo.Delete(id)
}

func (s *departmentService) GetUserDepartments(orgID, userID uuid.UUID) ([]model.Department, error) {
	return s.deptRepo.FindUserDepartments(orgID, userID)
}

func (s *departmentService) SetUserDepartments(orgID, userID uuid.UUID, departmentIDs []uuid.UUID) error {
	return s.deptRepo.SetUserDepartments(orgID, userID, departmentIDs)
}
