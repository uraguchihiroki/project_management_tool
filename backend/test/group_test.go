package test

import (
	"fmt"
	"net/http"
	"testing"
)

func TestGroup_CRUD_and_IssueGroupIDs(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "Gオーナー", "grp-owner@example.com")
	projectID := createTestProject(t, ts, "GRP", "グループテスト", ownerID)
	statusID := getFirstStatusID(t, ts, projectID)

	// POST /organizations/:orgId/groups
	st, createG := ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/groups", map[string]interface{}{
		"name": "通知チーム",
		"kind": "notification",
	})
	assertStatus(t, st, http.StatusCreated, "POST group")
	groupID := mustGetString(t, createG, "data", "id")

	st, listG := ts.req(t, "GET", "/api/v1/organizations/"+testOrgID+"/groups", nil)
	assertStatus(t, st, http.StatusOK, "GET groups")
	arr := mustGetArray(t, listG, "data")
	if len(arr) < 1 {
		t.Fatal("want at least 1 group")
	}

	// メンバー置換
	st, _ = ts.req(t, "PUT", "/api/v1/groups/"+groupID+"/members", map[string]interface{}{
		"user_ids": []string{ownerID},
	})
	assertStatus(t, st, http.StatusOK, "PUT group members")

	// Issue 作成時に group_ids
	st, issueResp := ts.req(t, "POST", fmt.Sprintf("/api/v1/projects/%s/issues", projectID), map[string]interface{}{
		"title":       "グループ付きIssue",
		"status_id":   statusID,
		"reporter_id": ownerID,
		"group_ids":   []string{groupID},
	})
	assertStatus(t, st, http.StatusCreated, "create issue with groups")
	num := mustGetFloat(t, issueResp, "data", "number")

	st, getIssue := ts.req(t, "GET", fmt.Sprintf("/api/v1/projects/%s/issues/%d", projectID, int(num)), nil)
	assertStatus(t, st, http.StatusOK, "GET issue")
	data := getIssue["data"].(map[string]interface{})
	gs, ok := data["groups"].([]interface{})
	if !ok || len(gs) != 1 {
		t.Fatalf("want 1 group on issue, got %+v", data["groups"])
	}
	g0 := gs[0].(map[string]interface{})
	if g0["id"] != groupID {
		t.Fatalf("group id: got %v", g0["id"])
	}

	// GET /projects/:id/issues?group_id=
	st, listIss := ts.req(t, "GET", fmt.Sprintf("/api/v1/projects/%s/issues?group_id=%s", projectID, groupID), nil)
	assertStatus(t, st, http.StatusOK, "list issues by group")
	list := mustGetArray(t, listIss, "data")
	if len(list) < 1 {
		t.Fatal("want at least 1 issue in filtered list")
	}

	// PUT issue groups（置換）
	st, _ = ts.req(t, "PUT", fmt.Sprintf("/api/v1/projects/%s/issues/%d/groups", projectID, int(num)), map[string]interface{}{
		"group_ids": []string{},
	})
	assertStatus(t, st, http.StatusOK, "clear issue groups")

	st, getIssue2 := ts.req(t, "GET", fmt.Sprintf("/api/v1/projects/%s/issues/%d", projectID, int(num)), nil)
	assertStatus(t, st, http.StatusOK, "GET issue after clear groups")
	data2 := getIssue2["data"].(map[string]interface{})
	if g2, ok := data2["groups"].([]interface{}); ok && len(g2) > 0 {
		t.Fatalf("want no groups, got %v", g2)
	}
}

func TestGroup_Get_and_UserGroups(t *testing.T) {
	ts := newTestServer(t)
	userID := createTestUser(t, ts, "GU", "grp-user@example.com")

	st, createG := ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/groups", map[string]interface{}{
		"name": "ユーザーグループ",
	})
	assertStatus(t, st, http.StatusCreated, "POST group")
	groupID := mustGetString(t, createG, "data", "id")

	ts.req(t, "PUT", "/api/v1/groups/"+groupID+"/members", map[string]interface{}{
		"user_ids": []string{userID},
	})

	st, getG := ts.req(t, "GET", "/api/v1/groups/"+groupID, nil)
	assertStatus(t, st, http.StatusOK, "GET group")
	if mustGetString(t, getG, "data", "name") != "ユーザーグループ" {
		t.Fatal("group name mismatch")
	}

	st, ug := ts.req(t, "GET", "/api/v1/users/"+userID+"/groups", nil)
	assertStatus(t, st, http.StatusOK, "GET user groups")
	ugArr := mustGetArray(t, ug, "data")
	if len(ugArr) != 1 {
		t.Fatalf("want 1 user group, got %d", len(ugArr))
	}
}
