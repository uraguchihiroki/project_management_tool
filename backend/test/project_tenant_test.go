package test

import (
	"net/http"
	"testing"
)

// 目的: マトリクス GET /projects — 組織ユーザーは自組織のプロジェクトだけ一覧に含まれること。
// 期待: data 内の organization_id はすべて JWT の組織。
func TestProject_List_OrgUser_OnlyOwnOrg(t *testing.T) {
	ts := newTestServer(t)
	// --- Arrange ---
	ownerID := createTestUser(t, ts, "PO", "proj-list-owner@example.com")
	createTestProject(t, ts, "POWN", "自社プロジェクト", ownerID)

	_, otherOrgResp := ts.req(t, "POST", "/api/v1/organizations", map[string]interface{}{"name": "プロジェクト他社"})
	otherOrgID := mustGetString(t, otherOrgResp, "data", "id")
	_, addResp := ts.req(t, "POST", "/api/v1/organizations/"+otherOrgID+"/users", map[string]interface{}{
		"user_id": ownerID,
	})
	otherOwnerID := mustGetString(t, addResp, "data", "id")
	ts.req(t, "POST", "/api/v1/projects", map[string]interface{}{
		"key":             "POTH",
		"name":            "他社プロジェクト",
		"owner_id":        otherOwnerID,
		"organization_id": otherOrgID,
	})

	email := "proj-list-tenant@example.com"
	listUserID := createTestUser(t, ts, "一覧ユーザー", email)
	ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/users", map[string]interface{}{"user_id": listUserID})
	_, loginResp := ts.reqNoAuth(t, "POST", "/api/v1/admin/login", map[string]string{"email": email})
	token := mustGetString(t, loginResp, "data", "token")

	// --- Act ---
	st, resp := ts.reqWithToken(t, token, "GET", "/api/v1/projects", nil)
	t.Logf("GET /projects http=%d", st)
	// --- Assert ---
	assertStatus(t, st, http.StatusOK, "GET /projects")
	for i, p := range mustGetArray(t, resp, "data") {
		oid := p.(map[string]interface{})["organization_id"].(string)
		if oid != testOrgID {
			t.Fatalf("project[%d] organization_id=%q, want only %q", i, oid, testOrgID)
		}
	}
}

// 目的: マトリクス GET/PUT/DELETE /projects/:id — 他組織プロジェクトを組織ユーザーが触れないこと。
// 期待: 404。
func TestProject_Detail_OtherOrg_AsOrgUser(t *testing.T) {
	ts := newTestServer(t)
	_, otherOrgResp := ts.req(t, "POST", "/api/v1/organizations", map[string]interface{}{"name": "PJ詳細他社"})
	otherOrgID := mustGetString(t, otherOrgResp, "data", "id")
	seedOwner := createTestUser(t, ts, "シード", "proj-seed@example.com")
	_, addResp := ts.req(t, "POST", "/api/v1/organizations/"+otherOrgID+"/users", map[string]interface{}{
		"user_id": seedOwner,
	})
	otherOwnerID := mustGetString(t, addResp, "data", "id")
	st, prResp := ts.req(t, "POST", "/api/v1/projects", map[string]interface{}{
		"key":             "PX",
		"name":            "他社PX",
		"owner_id":        otherOwnerID,
		"organization_id": otherOrgID,
	})
	assertStatus(t, st, http.StatusCreated, "other org project")
	otherProjID := mustGetString(t, prResp, "data", "id")

	email := "proj-detail-tenant@example.com"
	u := createTestUser(t, ts, "ユーザー", email)
	ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/users", map[string]interface{}{"user_id": u})
	_, loginResp := ts.reqNoAuth(t, "POST", "/api/v1/admin/login", map[string]string{"email": email})
	token := mustGetString(t, loginResp, "data", "token")

	t.Run("GET", func(t *testing.T) {
		st, _ := ts.reqWithToken(t, token, "GET", "/api/v1/projects/"+otherProjID, nil)
		t.Logf("GET other project http=%d", st)
		assertStatus(t, st, http.StatusNotFound, "GET other org project")
	})
	t.Run("PUT", func(t *testing.T) {
		newName := "不正更新"
		st, _ := ts.reqWithToken(t, token, "PUT", "/api/v1/projects/"+otherProjID, map[string]interface{}{
			"name": newName,
		})
		t.Logf("PUT other project http=%d", st)
		assertStatus(t, st, http.StatusNotFound, "PUT other org project")
	})
	t.Run("DELETE", func(t *testing.T) {
		st, _ := ts.reqWithToken(t, token, "DELETE", "/api/v1/projects/"+otherProjID, nil)
		t.Logf("DELETE other project http=%d", st)
		assertStatus(t, st, http.StatusNotFound, "DELETE other org project")
	})
}

// 目的: マトリクス POST /projects — organization_id が JWT と異なると 403。
// 期待: 403。
func TestProject_Create_WrongOrganizationID_Forbidden(t *testing.T) {
	ts := newTestServer(t)
	_, otherOrgResp := ts.req(t, "POST", "/api/v1/organizations", map[string]interface{}{"name": "PJ作成他社"})
	otherOrgID := mustGetString(t, otherOrgResp, "data", "id")
	ownerID := createTestUser(t, ts, "オーナー", "proj-create-tenant@example.com")
	ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/users", map[string]interface{}{"user_id": ownerID})
	_, loginResp := ts.reqNoAuth(t, "POST", "/api/v1/admin/login", map[string]string{"email": "proj-create-tenant@example.com"})
	token := mustGetString(t, loginResp, "data", "token")

	// --- Act ---
	st, _ := ts.reqWithToken(t, token, "POST", "/api/v1/projects", map[string]interface{}{
		"key":             "BAD",
		"name":            "だめな作成",
		"owner_id":        ownerID,
		"organization_id": otherOrgID,
	})
	t.Logf("POST project wrong org http=%d", st)
	// --- Assert ---
	assertStatus(t, st, http.StatusForbidden, "POST project other org_id")
}
