package service

import (
	"log"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

// TransitionAlertEvaluator はステータス遷移後に遷移アラート条件を評価する（ブロックしない）
type TransitionAlertEvaluator struct {
	Rules repository.TransitionAlertRuleRepository
	// AlertFunc が非 nil のとき、マッチしたルール通知をログの代わりに呼ぶ（テスト用）
	AlertFunc func(rule *model.TransitionAlertRule, issue *model.Issue, actorID uuid.UUID)
}

// OnStatusChanged はインプリント記録後に呼ぶ。マッチしたルールを通知する（将来メール）
func (e *TransitionAlertEvaluator) OnStatusChanged(issue *model.Issue, fromStatusID uuid.UUID, toStatusID uuid.UUID, actorID uuid.UUID) {
	if e == nil || e.Rules == nil {
		return
	}
	from := fromStatusID
	rules, err := e.Rules.FindMatching(issue.OrganizationID, &from, toStatusID)
	if err != nil {
		log.Printf("[transition-alert] FindMatching: %v", err)
		return
	}
	for i := range rules {
		rule := &rules[i]
		if e.AlertFunc != nil {
			e.AlertFunc(rule, issue, actorID)
			continue
		}
		log.Printf("[transition-alert] rule=%q issue=%s actor=%s", rule.Name, issue.ID, actorID)
	}
}
