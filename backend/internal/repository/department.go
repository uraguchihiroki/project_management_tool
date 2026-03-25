package repository

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

type GroupRepository interface {
	FindByOrganizationID(orgID uuid.UUID) ([]model.Group, error)
	FindByOrgAndName(orgID uuid.UUID, name string) (*model.Group, error)
	FindByID(id uuid.UUID) (*model.Group, error)
	Create(g *model.Group) error
	Update(g *model.Group) error
	Delete(id uuid.UUID) error
	AddUserToGroup(orgID, userID, groupID uuid.UUID) error
	RemoveUserFromGroup(orgID, userID, groupID uuid.UUID) error
	FindUserGroups(orgID, userID uuid.UUID) ([]model.Group, error)
	SetUserGroups(orgID, userID uuid.UUID, groupIDs []uuid.UUID) error
	Reorder(orgID uuid.UUID, ids []uuid.UUID) error
	GetMaxOrder(orgID uuid.UUID) (int, error)
}

type groupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) GroupRepository {
	return &groupRepository{db: db}
}

func (r *groupRepository) FindByOrganizationID(orgID uuid.UUID) ([]model.Group, error) {
	var groups []model.Group
	err := r.db.Where("organization_id = ?", orgID).Order("\"order\" ASC, name ASC").Find(&groups).Error
	return groups, err
}

func (r *groupRepository) FindByOrgAndName(orgID uuid.UUID, name string) (*model.Group, error) {
	var g model.Group
	err := r.db.Where("organization_id = ? AND name = ?", orgID, name).First(&g).Error
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *groupRepository) FindByID(id uuid.UUID) (*model.Group, error) {
	var g model.Group
	err := r.db.First(&g, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *groupRepository) Create(g *model.Group) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return r.db.Create(g).Error
}

func (r *groupRepository) Update(g *model.Group) error {
	return r.db.Model(g).Updates(map[string]interface{}{
		"name": g.Name,
	}).Error
}

func (r *groupRepository) GetMaxOrder(orgID uuid.UUID) (int, error) {
	var maxOrder int
	err := r.db.Model(&model.Group{}).Where("organization_id = ?", orgID).
		Select("COALESCE(MAX(\"order\"), 0)").Scan(&maxOrder).Error
	return maxOrder, err
}

func (r *groupRepository) Reorder(orgID uuid.UUID, ids []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&model.Group{}).
				Where("id = ? AND organization_id = ?", id, orgID).
				Update("order", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *groupRepository) Delete(id uuid.UUID) error {
	r.db.Where("group_id = ?", id).Delete(&model.OrganizationUserGroup{})
	return r.db.Delete(&model.Group{}, "id = ?", id).Error
}

func (r *groupRepository) AddUserToGroup(orgID, userID, groupID uuid.UUID) error {
	var n int64
	if err := r.db.Model(&model.OrganizationUserGroup{}).
		Where("organization_id = ? AND user_id = ? AND group_id = ?", orgID, userID, groupID).
		Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	oug := &model.OrganizationUserGroup{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		GroupID:        groupID,
		Key:            fmt.Sprintf("%s-%s-%s", orgID.String(), userID.String(), groupID.String()),
	}
	return r.db.Create(oug).Error
}

func (r *groupRepository) RemoveUserFromGroup(orgID, userID, groupID uuid.UUID) error {
	return r.db.Where("organization_id = ? AND user_id = ? AND group_id = ?", orgID, userID, groupID).
		Delete(&model.OrganizationUserGroup{}).Error
}

func (r *groupRepository) FindUserGroups(orgID, userID uuid.UUID) ([]model.Group, error) {
	var ougList []model.OrganizationUserGroup
	err := r.db.Preload("Group").Where("organization_id = ? AND user_id = ?", orgID, userID).Find(&ougList).Error
	if err != nil {
		return nil, err
	}
	groups := make([]model.Group, 0, len(ougList))
	for _, oug := range ougList {
		groups = append(groups, oug.Group)
	}
	return groups, nil
}

func (r *groupRepository) SetUserGroups(orgID, userID uuid.UUID, groupIDs []uuid.UUID) error {
	seen := map[uuid.UUID]struct{}{}
	var uniq []uuid.UUID
	for _, gid := range groupIDs {
		if _, dup := seen[gid]; dup {
			continue
		}
		seen[gid] = struct{}{}
		uniq = append(uniq, gid)
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("organization_id = ? AND user_id = ?", orgID, userID).
			Delete(&model.OrganizationUserGroup{}).Error; err != nil {
			return err
		}
		for _, gid := range uniq {
			oug := &model.OrganizationUserGroup{
				ID:             uuid.New(),
				OrganizationID: orgID,
				UserID:         userID,
				GroupID:        gid,
				Key:            fmt.Sprintf("%s-%s-%s", orgID.String(), userID.String(), gid.String()),
			}
			if err := tx.Create(oug).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
