package repository

import (
	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/pkg/keygen"
	"gorm.io/gorm"
)

type GroupRepository interface {
	Create(g *model.Group) error
	FindByID(id uuid.UUID) (*model.Group, error)
	ListByOrg(orgID uuid.UUID, kind *string) ([]model.Group, error)
	Update(g *model.Group) error
	Delete(id uuid.UUID) error
	FindByIDs(ids []uuid.UUID) ([]model.Group, error)
}

type groupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) GroupRepository {
	return &groupRepository{db: db}
}

func (r *groupRepository) Create(g *model.Group) error {
	return r.db.Create(g).Error
}

func (r *groupRepository) FindByID(id uuid.UUID) (*model.Group, error) {
	var g model.Group
	if err := r.db.First(&g, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *groupRepository) ListByOrg(orgID uuid.UUID, kind *string) ([]model.Group, error) {
	q := r.db.Where("organization_id = ?", orgID)
	if kind != nil && *kind != "" {
		q = q.Where("kind = ?", *kind)
	}
	var out []model.Group
	err := q.Order("display_order ASC, name ASC").Find(&out).Error
	return out, err
}

func (r *groupRepository) Update(g *model.Group) error {
	return r.db.Model(&model.Group{}).Where("id = ?", g.ID).Updates(map[string]interface{}{
		"name":        g.Name,
		"kind":        g.Kind,
		"display_order": g.DisplayOrder,
	}).Error
}

func (r *groupRepository) Delete(id uuid.UUID) error {
	_ = r.db.Delete(&model.UserGroup{}, "group_id = ?", id).Error
	_ = r.db.Delete(&model.IssueGroup{}, "group_id = ?", id).Error
	return r.db.Delete(&model.Group{}, "id = ?", id).Error
}

func (r *groupRepository) FindByIDs(ids []uuid.UUID) ([]model.Group, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var out []model.Group
	err := r.db.Where("id IN ?", ids).Find(&out).Error
	return out, err
}

// ---

type UserGroupRepository interface {
	ReplaceMembers(groupID uuid.UUID, userIDs []uuid.UUID) error
	ListMemberIDs(groupID uuid.UUID) ([]uuid.UUID, error)
	IsMember(userID, groupID uuid.UUID) bool
	ListGroupsByUser(userID uuid.UUID) ([]model.Group, error)
}

type userGroupRepository struct {
	db *gorm.DB
}

func NewUserGroupRepository(db *gorm.DB) UserGroupRepository {
	return &userGroupRepository{db: db}
}

func (r *userGroupRepository) ReplaceMembers(groupID uuid.UUID, userIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&model.UserGroup{}, "group_id = ?", groupID).Error; err != nil {
			return err
		}
		for _, uid := range userIDs {
			ug := model.UserGroup{
				UserID:  uid,
				GroupID: groupID,
				Key:     keygen.UUIDKey(uuid.New()),
			}
			if err := tx.Create(&ug).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *userGroupRepository) ListMemberIDs(groupID uuid.UUID) ([]uuid.UUID, error) {
	var rows []model.UserGroup
	if err := r.db.Where("group_id = ?", groupID).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]uuid.UUID, len(rows))
	for i := range rows {
		out[i] = rows[i].UserID
	}
	return out, nil
}

func (r *userGroupRepository) IsMember(userID, groupID uuid.UUID) bool {
	var n int64
	r.db.Model(&model.UserGroup{}).Where("user_id = ? AND group_id = ?", userID, groupID).Count(&n)
	return n > 0
}

func (r *userGroupRepository) ListGroupsByUser(userID uuid.UUID) ([]model.Group, error) {
	var groups []model.Group
	err := r.db.
		Joins("JOIN user_groups ON user_groups.group_id = groups.id").
		Where("user_groups.user_id = ?", userID).
		Order("groups.name ASC").
		Find(&groups).Error
	return groups, err
}

// ---

type IssueGroupRepository interface {
	ReplaceForIssue(issueID uuid.UUID, groupIDs []uuid.UUID) error
	ListGroupIDsByIssue(issueID uuid.UUID) ([]uuid.UUID, error)
	ListGroupsByIssue(issueID uuid.UUID) ([]model.Group, error)
}

type issueGroupRepository struct {
	db *gorm.DB
}

func NewIssueGroupRepository(db *gorm.DB) IssueGroupRepository {
	return &issueGroupRepository{db: db}
}

func (r *issueGroupRepository) ReplaceForIssue(issueID uuid.UUID, groupIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&model.IssueGroup{}, "issue_id = ?", issueID).Error; err != nil {
			return err
		}
		for _, gid := range groupIDs {
			ig := model.IssueGroup{
				IssueID: issueID,
				GroupID: gid,
				Key:     keygen.UUIDKey(uuid.New()),
			}
			if err := tx.Create(&ig).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *issueGroupRepository) ListGroupIDsByIssue(issueID uuid.UUID) ([]uuid.UUID, error) {
	var rows []model.IssueGroup
	if err := r.db.Where("issue_id = ?", issueID).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]uuid.UUID, len(rows))
	for i := range rows {
		out[i] = rows[i].GroupID
	}
	return out, nil
}

func (r *issueGroupRepository) ListGroupsByIssue(issueID uuid.UUID) ([]model.Group, error) {
	ids, err := r.ListGroupIDsByIssue(issueID)
	if err != nil || len(ids) == 0 {
		return nil, err
	}
	var groups []model.Group
	err = r.db.Where("id IN ?", ids).Order("name ASC").Find(&groups).Error
	return groups, err
}
