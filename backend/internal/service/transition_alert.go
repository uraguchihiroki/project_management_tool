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
	UG    repository.UserGroupRepository
}

// OnStatusChanged はインプリント記録後に呼ぶ。想定外 actor ならログに出す（将来メール）
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
		if rule.ExpectedGroupID == nil {
			continue
		}
		if e.UG != nil && e.UG.IsMember(actorID, *rule.ExpectedGroupID) {
			continue
		}
		log.Printf("[transition-alert] rule=%q issue=%s actor=%s not in expected_group=%s (notify_group=%v)",
			rule.Name, issue.ID, actorID, rule.ExpectedGroupID, rule.NotifyGroupID)
	}
}
