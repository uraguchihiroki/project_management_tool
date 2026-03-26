package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

type memRules struct {
	rules []model.TransitionAlertRule
	err   error
}

func (m *memRules) FindMatching(orgID uuid.UUID, fromStatusID *uuid.UUID, toStatusID uuid.UUID) ([]model.TransitionAlertRule, error) {
	if m.err != nil {
		return nil, m.err
	}
	var out []model.TransitionAlertRule
	for i := range m.rules {
		r := &m.rules[i]
		if r.OrganizationID != orgID || r.ToStatusID != toStatusID {
			continue
		}
		if r.FromStatusID != nil {
			if fromStatusID == nil || *r.FromStatusID != *fromStatusID {
				continue
			}
		}
		out = append(out, *r)
	}
	return out, nil
}

func (m *memRules) Create(r *model.TransitionAlertRule) error { return nil }

var _ repository.TransitionAlertRuleRepository = (*memRules)(nil)

func TestTransitionAlertEvaluator_OnStatusChanged_matchingRule_alerts(t *testing.T) {
	org := uuid.New()
	from := uuid.New()
	to := uuid.New()
	actor := uuid.New()

	rule := model.TransitionAlertRule{
		ID:              uuid.New(),
		OrganizationID:  org,
		Name:            "r1",
		FromStatusID:    &from,
		ToStatusID:      to,
	}
	var alertCalls int
	ev := &TransitionAlertEvaluator{
		Rules: &memRules{rules: []model.TransitionAlertRule{rule}},
		AlertFunc: func(*model.TransitionAlertRule, *model.Issue, uuid.UUID) {
			alertCalls++
		},
	}
	issue := &model.Issue{ID: uuid.New(), OrganizationID: org}
	ev.OnStatusChanged(issue, from, to, actor)
	if alertCalls != 1 {
		t.Fatalf("expected one alert for matching rule, got %d", alertCalls)
	}
}

func TestTransitionAlertEvaluator_OnStatusChanged_nonMatchingRule_skipped(t *testing.T) {
	org := uuid.New()
	from := uuid.New()
	to := uuid.New()
	actor := uuid.New()

	rule := model.TransitionAlertRule{
		ID:              uuid.New(),
		OrganizationID:  org,
		Name:            "r1",
		FromStatusID:    &from,
		ToStatusID:      uuid.New(),
	}
	alertCalls := 0
	ev := &TransitionAlertEvaluator{
		Rules: &memRules{rules: []model.TransitionAlertRule{rule}},
		AlertFunc: func(r *model.TransitionAlertRule, _ *model.Issue, a uuid.UUID) {
			alertCalls++
		},
	}
	issue := &model.Issue{ID: uuid.New(), OrganizationID: org}
	ev.OnStatusChanged(issue, from, to, actor)
	if alertCalls != 0 {
		t.Fatalf("want 0 alerts for non matching rule, got %d", alertCalls)
	}
}
