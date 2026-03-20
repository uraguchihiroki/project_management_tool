package test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/auth"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
)

func TestCrossOrganizationAuthorization(t *testing.T) {
	ts := newTestServer(t)

	// 2つの組織を作成
	_, org1Resp := ts.req(t, http.MethodPost, "/api/v1/organizations", map[string]interface{}{"name": "Org-1"})
	_, org2Resp := ts.req(t, http.MethodPost, "/api/v1/organizations", map[string]interface{}{"name": "Org-2"})
	org1ID := mustGetString(t, org1Resp, "data", "id")
	org2ID := mustGetString(t, org2Resp, "data", "id")

	org1UUID := uuid.MustParse(org1ID)
	org2UUID := uuid.MustParse(org2ID)

	owner1 := model.User{
		ID:             uuid.New(),
		Key:            "owner-org1",
		OrganizationID: org1UUID,
		Name:           "Owner 1",
		Email:          "owner1@example.com",
		IsAdmin:        true,
		JoinedAt:       time.Now(),
		CreatedAt:      time.Now(),
	}
	owner2 := model.User{
		ID:             uuid.New(),
		Key:            "owner-org2",
		OrganizationID: org2UUID,
		Name:           "Owner 2",
		Email:          "owner2@example.com",
		IsAdmin:        true,
		JoinedAt:       time.Now(),
		CreatedAt:      time.Now(),
	}
	if err := ts.db.Create(&owner1).Error; err != nil {
		t.Fatalf("create owner1: %v", err)
	}
	if err := ts.db.Create(&owner2).Error; err != nil {
		t.Fatalf("create owner2: %v", err)
	}

	// 各組織にステータス作成
	_, _ = ts.req(t, http.MethodPost, fmt.Sprintf("/api/v1/organizations/%s/statuses", org1ID), map[string]interface{}{
		"name": "org1-status", "color": "#3B82F6", "type": "issue", "order": 1,
	})
	_, org2StatusResp := ts.req(t, http.MethodPost, fmt.Sprintf("/api/v1/organizations/%s/statuses", org2ID), map[string]interface{}{
		"name": "org2-status", "color": "#EF4444", "type": "issue", "order": 1,
	})
	org2StatusID := mustGetString(t, org2StatusResp, "data", "id")

	// 各組織にプロジェクト作成
	_, _ = ts.req(t, http.MethodPost, "/api/v1/projects", map[string]interface{}{
		"key":             "ORG1",
		"name":            "Project Org1",
		"owner_id":        owner1.ID.String(),
		"organization_id": org1ID,
	})
	_, org2ProjectResp := ts.req(t, http.MethodPost, "/api/v1/projects", map[string]interface{}{
		"key":             "ORG2",
		"name":            "Project Org2",
		"owner_id":        owner2.ID.String(),
		"organization_id": org2ID,
	})
	org2ProjectID := mustGetString(t, org2ProjectResp, "data", "id")

	// org2 に issue 作成
	_, issueResp := ts.req(t, http.MethodPost, fmt.Sprintf("/api/v1/projects/%s/issues", org2ProjectID), map[string]interface{}{
		"title":       "Issue Org2",
		"status_id":   org2StatusID,
		"reporter_id": owner2.ID.String(),
	})
	issueNumber := int(mustGetFloat(t, issueResp, "data", "number"))

	org1Token, _ := auth.GenerateUserToken(owner1.ID, org1UUID, true)

	t.Run("org1 admin cannot list org2 statuses", func(t *testing.T) {
		status, _ := ts.reqWithToken(t, org1Token, http.MethodGet, fmt.Sprintf("/api/v1/organizations/%s/statuses", org2ID), nil)
		if status != http.StatusForbidden {
			t.Fatalf("status=%d, want 403", status)
		}
	})

	t.Run("org1 admin cannot access org2 project issue", func(t *testing.T) {
		status, _ := ts.reqWithToken(t, org1Token, http.MethodGet, fmt.Sprintf("/api/v1/projects/%s/issues/%d", org2ProjectID, issueNumber), nil)
		if status != http.StatusNotFound {
			t.Fatalf("status=%d, want 404", status)
		}
	})

	t.Run("super admin can access org2 statuses", func(t *testing.T) {
		status, resp := ts.req(t, http.MethodGet, fmt.Sprintf("/api/v1/organizations/%s/statuses", org2ID), nil)
		if status != http.StatusOK {
			t.Fatalf("status=%d, want 200", status)
		}
		items, ok := resp["data"].([]interface{})
		if !ok || len(items) == 0 {
			t.Fatalf("expected statuses")
		}
	})
}
