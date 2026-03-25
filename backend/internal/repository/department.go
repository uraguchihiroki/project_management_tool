package repository

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

type DepartmentRepository interface {
	FindByOrganizationID(orgID uuid.UUID) ([]model.Department, error)
	FindByOrgAndName(orgID uuid.UUID, name string) (*model.Department, error)
	FindByID(id uuid.UUID) (*model.Department, error)
	Create(d *model.Department) error
	Update(d *model.Department) error
	Delete(id uuid.UUID) error
	AddUserToDepartment(orgID, userID, departmentID uuid.UUID) error
	RemoveUserFromDepartment(orgID, userID, departmentID uuid.UUID) error
	FindUserDepartments(orgID, userID uuid.UUID) ([]model.Department, error)
	SetUserDepartments(orgID, userID uuid.UUID, departmentIDs []uuid.UUID) error
	Reorder(orgID uuid.UUID, ids []uuid.UUID) error
	GetMaxOrder(orgID uuid.UUID) (int, error)
}

type departmentRepository struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) DepartmentRepository {
	return &departmentRepository{db: db}
}

func (r *departmentRepository) FindByOrganizationID(orgID uuid.UUID) ([]model.Department, error) {
	var depts []model.Department
	err := r.db.Where("organization_id = ?", orgID).Order("\"order\" ASC, name ASC").Find(&depts).Error
	return depts, err
}

func (r *departmentRepository) FindByOrgAndName(orgID uuid.UUID, name string) (*model.Department, error) {
	var d model.Department
	err := r.db.Where("organization_id = ? AND name = ?", orgID, name).First(&d).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *departmentRepository) FindByID(id uuid.UUID) (*model.Department, error) {
	var d model.Department
	err := r.db.First(&d, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *departmentRepository) Create(d *model.Department) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return r.db.Create(d).Error
}

func (r *departmentRepository) Update(d *model.Department) error {
	return r.db.Model(d).Updates(map[string]interface{}{
		"name": d.Name,
	}).Error
}

func (r *departmentRepository) GetMaxOrder(orgID uuid.UUID) (int, error) {
	var maxOrder int
	err := r.db.Model(&model.Department{}).Where("organization_id = ?", orgID).
		Select("COALESCE(MAX(\"order\"), 0)").Scan(&maxOrder).Error
	return maxOrder, err
}

func (r *departmentRepository) Reorder(orgID uuid.UUID, ids []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&model.Department{}).
				Where("id = ? AND organization_id = ?", id, orgID).
				Update("order", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *departmentRepository) Delete(id uuid.UUID) error {
	r.db.Where("department_id = ?", id).Delete(&model.OrganizationUserDepartment{})
	return r.db.Delete(&model.Department{}, "id = ?", id).Error
}

func (r *departmentRepository) AddUserToDepartment(orgID, userID, departmentID uuid.UUID) error {
	var n int64
	if err := r.db.Model(&model.OrganizationUserDepartment{}).
		Where("organization_id = ? AND user_id = ? AND department_id = ?", orgID, userID, departmentID).
		Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	oud := &model.OrganizationUserDepartment{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		DepartmentID:   departmentID,
		Key:            fmt.Sprintf("%s-%s-%s", orgID.String(), userID.String(), departmentID.String()),
	}
	return r.db.Create(oud).Error
}

func (r *departmentRepository) RemoveUserFromDepartment(orgID, userID, departmentID uuid.UUID) error {
	return r.db.Where("organization_id = ? AND user_id = ? AND department_id = ?", orgID, userID, departmentID).
		Delete(&model.OrganizationUserDepartment{}).Error
}

func (r *departmentRepository) FindUserDepartments(orgID, userID uuid.UUID) ([]model.Department, error) {
	var oudList []model.OrganizationUserDepartment
	err := r.db.Preload("Department").Where("organization_id = ? AND user_id = ?", orgID, userID).Find(&oudList).Error
	if err != nil {
		return nil, err
	}
	depts := make([]model.Department, 0, len(oudList))
	for _, oud := range oudList {
		depts = append(depts, oud.Department)
	}
	return depts, nil
}

func (r *departmentRepository) SetUserDepartments(orgID, userID uuid.UUID, departmentIDs []uuid.UUID) error {
	seen := map[uuid.UUID]struct{}{}
	var uniq []uuid.UUID
	for _, did := range departmentIDs {
		if _, dup := seen[did]; dup {
			continue
		}
		seen[did] = struct{}{}
		uniq = append(uniq, did)
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("organization_id = ? AND user_id = ?", orgID, userID).
			Delete(&model.OrganizationUserDepartment{}).Error; err != nil {
			return err
		}
		for _, did := range uniq {
			oud := &model.OrganizationUserDepartment{
				ID:             uuid.New(),
				OrganizationID: orgID,
				UserID:         userID,
				DepartmentID:   did,
				Key:            fmt.Sprintf("%s-%s-%s", orgID.String(), userID.String(), did.String()),
			}
			if err := tx.Create(oud).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
