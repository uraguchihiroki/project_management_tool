package repository

import (
	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

type TransitionAlertRuleRepository interface {
	FindMatching(orgID uuid.UUID, fromStatusID *uuid.UUID, toStatusID uuid.UUID) ([]model.TransitionAlertRule, error)
	Create(r *model.TransitionAlertRule) error
}

type transitionAlertRuleRepository struct {
	db *gorm.DB
}

func NewTransitionAlertRuleRepository(db *gorm.DB) TransitionAlertRuleRepository {
	return &transitionAlertRuleRepository{db: db}
}

func (r *transitionAlertRuleRepository) FindMatching(orgID uuid.UUID, fromStatusID *uuid.UUID, toStatusID uuid.UUID) ([]model.TransitionAlertRule, error) {
	var rules []model.TransitionAlertRule
	q := r.db.Where("organization_id = ? AND to_status_id = ?", orgID, toStatusID)
	if err := q.Find(&rules).Error; err != nil {
		return nil, err
	}
	var out []model.TransitionAlertRule
	for i := range rules {
		rule := &rules[i]
		if rule.FromStatusID != nil {
			if fromStatusID == nil || *rule.FromStatusID != *fromStatusID {
				continue
			}
		}
		out = append(out, *rule)
	}
	return out, nil
}

func (r *transitionAlertRuleRepository) Create(rule *model.TransitionAlertRule) error {
	return r.db.Create(rule).Error
}
