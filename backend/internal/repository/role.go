package repository

import (
	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

type RoleRepository interface {
	FindAll() ([]model.Role, error)
	FindByOrg(orgID uuid.UUID) ([]model.Role, error)
	FindByID(id uint) (*model.Role, error)
	Create(role *model.Role) error
	Update(role *model.Role) error
	Delete(id uint) error
	AssignRolesToUser(userID uuid.UUID, roleIDs []uint) error
	FindRolesByUserID(userID uuid.UUID) ([]model.Role, error)
	Reorder(orgID *uuid.UUID, ids []uint) error
	GetMaxOrder(orgID *uuid.UUID) (int, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) FindAll() ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Order("display_order ASC, level DESC").Find(&roles).Error
	return roles, err
}

func (r *roleRepository) FindByOrg(orgID uuid.UUID) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Where("organization_id = ?", orgID).Order("display_order ASC, level DESC").Find(&roles).Error
	return roles, err
}

func (r *roleRepository) FindByID(id uint) (*model.Role, error) {
	var role model.Role
	err := r.db.First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) Create(role *model.Role) error {
	return r.db.Create(role).Error
}

func (r *roleRepository) Update(role *model.Role) error {
	return r.db.Model(&model.Role{}).Where("id = ?", role.ID).Updates(map[string]interface{}{
		"name":        role.Name,
		"level":       role.Level,
		"description": role.Description,
	}).Error
}

func (r *roleRepository) Delete(id uint) error {
	return r.db.Delete(&model.Role{}, id).Error
}

func (r *roleRepository) AssignRolesToUser(userID uuid.UUID, roleIDs []uint) error {
	user := &model.User{ID: userID}
	if len(roleIDs) == 0 {
		return r.db.Model(user).Association("Roles").Clear()
	}
	var roles []model.Role
	if err := r.db.Find(&roles, roleIDs).Error; err != nil {
		return err
	}
	return r.db.Model(user).Association("Roles").Replace(roles)
}

func (r *roleRepository) FindRolesByUserID(userID uuid.UUID) ([]model.Role, error) {
	var user model.User
	if err := r.db.Preload("Roles").First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}
	return user.Roles, nil
}

func (r *roleRepository) GetMaxOrder(orgID *uuid.UUID) (int, error) {
	var maxOrder int
	q := r.db.Model(&model.Role{})
	if orgID != nil {
		q = q.Where("organization_id = ?", orgID)
	}
	err := q.Select("COALESCE(MAX(display_order), 0)").Scan(&maxOrder).Error
	return maxOrder, err
}

func (r *roleRepository) Reorder(orgID *uuid.UUID, ids []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			q := tx.Model(&model.Role{}).Where("id = ?", id)
			if orgID != nil {
				q = q.Where("organization_id = ?", orgID)
			}
			if err := q.Update("display_order", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
