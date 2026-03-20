package repository

import (
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

type WorkflowRepository interface {
	FindAll() ([]model.Workflow, error)
	FindByID(id uint) (*model.Workflow, error)
	FindStepsByWorkflowID(workflowID uint) ([]model.WorkflowStep, error)
	Create(workflow *model.Workflow) error
	Update(workflow *model.Workflow) error
	Delete(id uint) error
	CreateStep(step *model.WorkflowStep) error
	UpdateStep(step *model.WorkflowStep) error
	DeleteStep(id uint) error
	FindStepByID(id uint) (*model.WorkflowStep, error)
	CountSteps(workflowID uint) (int64, error)
	Reorder(ids []uint) error
	ReorderSteps(workflowID uint, ids []uint) error
	ReorderStepsWithNextStatus(workflowID uint, ids []uint) error
	GetMaxOrder() (int, error)
	CreateApprovalObject(obj *model.ApprovalObject) error
	UpdateApprovalObject(obj *model.ApprovalObject) error
	DeleteApprovalObject(id uint) error
	DeleteApprovalObjectsByStepID(stepID uint) error
	CountApprovalObjects(stepID uint) (int64, error)
}

type workflowRepository struct {
	db *gorm.DB
}

func NewWorkflowRepository(db *gorm.DB) WorkflowRepository {
	return &workflowRepository{db: db}
}

func (r *workflowRepository) FindAll() ([]model.Workflow, error) {
	var workflows []model.Workflow
	// ステップ未追加のワークフローも一覧に含める（管理画面で作成直後に表示するため）。
	// 以前の INNER JOIN ではステップ0件の行が JOIN から落ち、一覧に出なかった。
	err := r.db.Order("workflows.display_order ASC").Find(&workflows).Error
	return workflows, err
}

func (r *workflowRepository) FindStepsByWorkflowID(workflowID uint) ([]model.WorkflowStep, error) {
	var steps []model.WorkflowStep
	err := r.db.Where("workflow_id = ?", workflowID).
		Preload("Status").
		Order("\"order\" ASC").
		Find(&steps).Error
	return steps, err
}

func (r *workflowRepository) FindByID(id uint) (*model.Workflow, error) {
	var workflow model.Workflow
	err := r.db.
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Status").Preload("NextStatus").Preload("ApprovalObjects", func(d *gorm.DB) *gorm.DB {
				return d.Order("sort_order ASC").Preload("Role").Preload("User")
			})
		}).
		First(&workflow, id).Error
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (r *workflowRepository) Create(workflow *model.Workflow) error {
	return r.db.Create(workflow).Error
}

func (r *workflowRepository) Update(workflow *model.Workflow) error {
	updates := map[string]interface{}{
		"name":        workflow.Name,
		"description": workflow.Description,
	}
	if workflow.Key != "" {
		updates["key"] = workflow.Key
	}
	return r.db.Model(&model.Workflow{}).Where("id = ?", workflow.ID).Updates(updates).Error
}

func (r *workflowRepository) Delete(id uint) error {
	// 承認オブジェクト→ステップ→ワークフローの順で削除
	var stepIDs []uint
	if err := r.db.Model(&model.WorkflowStep{}).Where("workflow_id = ?", id).Pluck("id", &stepIDs).Error; err != nil {
		return err
	}
	if len(stepIDs) > 0 {
		if err := r.db.Where("workflow_step_id IN ?", stepIDs).Delete(&model.ApprovalObject{}).Error; err != nil {
			return err
		}
	}
	if err := r.db.Where("workflow_id = ?", id).Delete(&model.WorkflowStep{}).Error; err != nil {
		return err
	}
	return r.db.Delete(&model.Workflow{}, id).Error
}

func (r *workflowRepository) CreateStep(step *model.WorkflowStep) error {
	return r.db.Create(step).Error
}

func (r *workflowRepository) UpdateStep(step *model.WorkflowStep) error {
	updates := map[string]interface{}{
		"status_id":        step.StatusID,
		"next_status_id":   step.NextStatusID,
		"description":      step.Description,
		"threshold":        step.Threshold,
		"exclude_reporter": step.ExcludeReporter,
		"exclude_assignee": step.ExcludeAssignee,
	}
	if step.Key != "" {
		updates["key"] = step.Key
	}
	return r.db.Model(&model.WorkflowStep{}).Where("id = ?", step.ID).Updates(updates).Error
}

func (r *workflowRepository) DeleteStep(id uint) error {
	if err := r.db.Where("workflow_step_id = ?", id).Delete(&model.ApprovalObject{}).Error; err != nil {
		return err
	}
	return r.db.Delete(&model.WorkflowStep{}, id).Error
}

func (r *workflowRepository) FindStepByID(id uint) (*model.WorkflowStep, error) {
	var step model.WorkflowStep
	err := r.db.
		Preload("Status").
		Preload("NextStatus").
		Preload("ApprovalObjects", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC").Preload("Role").Preload("User")
		}).
		First(&step, id).Error
	if err != nil {
		return nil, err
	}
	return &step, nil
}

func (r *workflowRepository) CountSteps(workflowID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.WorkflowStep{}).Where("workflow_id = ?", workflowID).Count(&count).Error
	return count, err
}

func (r *workflowRepository) GetMaxOrder() (int, error) {
	var maxOrder int
	err := r.db.Model(&model.Workflow{}).Select("COALESCE(MAX(display_order), 0)").Scan(&maxOrder).Error
	return maxOrder, err
}

func (r *workflowRepository) Reorder(ids []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&model.Workflow{}).Where("id = ?", id).
				Update("display_order", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *workflowRepository) ReorderSteps(workflowID uint, ids []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&model.WorkflowStep{}).
				Where("id = ? AND workflow_id = ?", id, workflowID).
				Update("order", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *workflowRepository) ReorderStepsWithNextStatus(workflowID uint, ids []uint) error {
	steps, err := r.FindStepsByWorkflowID(workflowID)
	if err != nil {
		return err
	}
	idSet := make(map[uint]bool)
	for _, id := range ids {
		idSet[id] = true
	}
	var stsStart, stsGoal *model.WorkflowStep
	var userSteps []model.WorkflowStep
	for i := range steps {
		s := &steps[i]
		if s.Status != nil {
			switch s.Status.StatusKey {
			case "sts_start":
				stsStart = s
			case "sts_goal":
				stsGoal = s
			default:
				if idSet[s.ID] {
					userSteps = append(userSteps, *s)
				}
			}
		}
	}
	// Build ordered list: sts_start, userSteps in ids order, sts_goal
	ordered := make([]*model.WorkflowStep, 0, len(steps))
	userStepByID := make(map[uint]*model.WorkflowStep)
	for i := range userSteps {
		userStepByID[userSteps[i].ID] = &userSteps[i]
	}
	if stsStart != nil {
		ordered = append(ordered, stsStart)
	}
	for _, id := range ids {
		if s := userStepByID[id]; s != nil {
			ordered = append(ordered, s)
		}
	}
	if stsGoal != nil {
		ordered = append(ordered, stsGoal)
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, step := range ordered {
			updates := map[string]interface{}{"order": i}
			if i+1 < len(ordered) {
				updates["next_status_id"] = ordered[i+1].StatusID
			} else {
				updates["next_status_id"] = nil
			}
			if err := tx.Model(&model.WorkflowStep{}).
				Where("id = ? AND workflow_id = ?", step.ID, workflowID).
				Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *workflowRepository) CreateApprovalObject(obj *model.ApprovalObject) error {
	return r.db.Create(obj).Error
}

func (r *workflowRepository) UpdateApprovalObject(obj *model.ApprovalObject) error {
	updates := map[string]interface{}{
		"sort_order":       obj.Order,
		"type":             obj.Type,
		"role_id":          obj.RoleID,
		"role_operator":    obj.RoleOperator,
		"user_id":          obj.UserID,
		"points":           obj.Points,
		"exclude_reporter": obj.ExcludeReporter,
		"exclude_assignee": obj.ExcludeAssignee,
	}
	if obj.Key != "" {
		updates["key"] = obj.Key
	}
	return r.db.Model(&model.ApprovalObject{}).Where("id = ?", obj.ID).Updates(updates).Error
}

func (r *workflowRepository) DeleteApprovalObject(id uint) error {
	return r.db.Delete(&model.ApprovalObject{}, id).Error
}

func (r *workflowRepository) DeleteApprovalObjectsByStepID(stepID uint) error {
	return r.db.Where("workflow_step_id = ?", stepID).Delete(&model.ApprovalObject{}).Error
}

func (r *workflowRepository) CountApprovalObjects(stepID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.ApprovalObject{}).Where("workflow_step_id = ?", stepID).Count(&count).Error
	return count, err
}
