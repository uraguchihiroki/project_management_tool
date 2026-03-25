package repository

import (
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

type SuperAdminRepository interface {
	FindByEmail(email string) (*model.SuperAdmin, error)
	Create(admin *model.SuperAdmin) error
	Count() (int64, error)
}

type superAdminRepository struct {
	db *gorm.DB
}

func NewSuperAdminRepository(db *gorm.DB) SuperAdminRepository {
	return &superAdminRepository{db: db}
}

func (r *superAdminRepository) FindByEmail(email string) (*model.SuperAdmin, error) {
	var admin model.SuperAdmin
	err := r.db.Where("email = ?", email).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *superAdminRepository) Create(admin *model.SuperAdmin) error {
	return r.db.Create(admin).Error
}

func (r *superAdminRepository) Count() (int64, error) {
	var count int64
	r.db.Model(&model.SuperAdmin{}).Count(&count)
	return count, nil
}
