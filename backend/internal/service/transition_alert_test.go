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

type memUG struct {
	pairs map[string]bool // "userID|groupID" -> member
}

func (m *memUG) ReplaceMembers(groupID uuid.UUID, userIDs []uuid.UUID) error { return nil }
func (m *memUG) ListMemberIDs(groupID uuid.UUID) ([]uuid.UUID, error)        { return nil, nil }
func (m *memUG) ListGroupsByUser(userID uuid.UUID) ([]model.Group, error)    { return nil, nil }
func (m *memUG) IsMember(userID, groupID uuid.UUID) bool {
	if m.pairs == nil {
		return false
	}
	return m.pairs[userID.String()+"|"+groupID.String()]
}

var _ repository.UserGroupRepository = (*memUG)(nil)

func TestTransitionAlertEvaluator_OnStatusChanged_expectedMember_noAlert(t *testing.T) {
	org := uuid.New()
	from := uuid.New()
	to := uuid.New()
	gid := uuid.New()
	actor := uuid.New()

	rule := model.TransitionAlertRule{
		ID:              uuid.New(),
		OrganizationID:  org,
		Name:            "r1",
		FromStatusID:    &from,
		ToStatusID:      to,
		ExpectedGroupID: &gid,
	}
	var alertCalls int
	ev := &TransitionAlertEvaluator{
		Rules: &memRules{rules: []model.TransitionAlertRule{rule}},
		UG: &memUG{pairs: map[string]bool{
			actor.String() + "|" + gid.String(): true,
		}},
		AlertFunc: func(*model.TransitionAlertRule, *model.Issue, uuid.UUID) {
			alertCalls++
		},
	}
	issue := &model.Issue{ID: uuid.New(), OrganizationID: org}
	ev.OnStatusChanged(issue, from, to, actor)
	if alertCalls != 0 {
		t.Fatalf("expected no alert when actor in group, got %d", alertCalls)
	}
}

func TestTransitionAlertEvaluator_OnStatusChanged_unexpectedMember_alertsOnce(t *testing.T) {
	org := uuid.New()
	from := uuid.New()
	to := uuid.New()
	gid := uuid.New()
	actor := uuid.New()

	rule := model.TransitionAlertRule{
		ID:              uuid.New(),
		OrganizationID:  org,
		Name:            "r1",
		FromStatusID:    &from,
		ToStatusID:      to,
		ExpectedGroupID: &gid,
	}
	var gotRule *model.TransitionAlertRule
	var gotActor uuid.UUID
	alertCalls := 0
	ev := &TransitionAlertEvaluator{
		Rules: &memRules{rules: []model.TransitionAlertRule{rule}},
		UG:    &memUG{pairs: map[string]bool{}}, // actor not in group
		AlertFunc: func(r *model.TransitionAlertRule, _ *model.Issue, a uuid.UUID) {
			alertCalls++
			gotRule = r
			gotActor = a
		},
	}
	issue := &model.Issue{ID: uuid.New(), OrganizationID: org}
	ev.OnStatusChanged(issue, from, to, actor)
	if alertCalls != 1 {
		t.Fatalf("want 1 alert, got %d", alertCalls)
	}
	if gotRule == nil || gotRule.Name != "r1" {
		t.Fatalf("unexpected rule: %+v", gotRule)
	}
	if gotActor != actor {
		t.Fatalf("actor: got %v want %v", gotActor, actor)
	}
}

func TestTransitionAlertEvaluator_OnStatusChanged_nilExpectedGroup_skipped(t *testing.T) {
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
		ExpectedGroupID: nil,
	}
	var alertCalls int
	ev := &TransitionAlertEvaluator{
		Rules: &memRules{rules: []model.TransitionAlertRule{rule}},
		UG:    &memUG{},
		AlertFunc: func(*model.TransitionAlertRule, *model.Issue, uuid.UUID) {
			alertCalls++
		},
	}
	issue := &model.Issue{ID: uuid.New(), OrganizationID: org}
	ev.OnStatusChanged(issue, from, to, actor)
	if alertCalls != 0 {
		t.Fatalf("rule without ExpectedGroupID should not alert, got %d", alertCalls)
	}
}
