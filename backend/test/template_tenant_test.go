package test

import (
	"fmt"
	"net/http"
	"testing"
)

// 目的: マトリクス GET /templates — 組織ユーザーは自組織プロジェクトのテンプレだけ一覧に含むこと。
// 期待: 他社プロジェクトに紐づくテンプレは data に含まれない。
func TestTemplate_List_OrgUser_ExcludesOtherOrg(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "TO", "tmpl-list-owner@example.com")
	ownProjectID := createTestProject(t, ts, "TL", "自社テンプレPJ", ownerID)
	tmplOwnID := createTestTemplate(t, ts, ownProjectID, "自社テンプレ")

	_, otherOrgResp := ts.req(t, "POST", "/api/v1/organizations", map[string]interface{}{"name": "テンプレ他社"})
	otherOrgID := mustGetString(t, otherOrgResp, "data", "id")
	_, addResp := ts.req(t, "POST", "/api/v1/organizations/"+otherOrgID+"/users", map[string]interface{}{
		"user_id": ownerID,
	})
	otherOwnerID := mustGetString(t, addResp, "data", "id")
	st, prResp := ts.req(t, "POST", "/api/v1/projects", map[string]interface{}{
		"key":             "TLO",
		"name":            "他社テンプレPJ",
		"owner_id":        otherOwnerID,
		"organization_id": otherOrgID,
	})
	assertStatus(t, st, http.StatusCreated, "other project")
	otherProjID := mustGetString(t, prResp, "data", "id")
	tmplOtherID := createTestTemplate(t, ts, otherProjID, "他社テンプレ")

	email := "tmpl-list-tenant@example.com"
	u := createTestUser(t, ts, "ユーザー", email)
	ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/users", map[string]interface{}{"user_id": u})
	_, loginResp := ts.reqNoAuth(t, "POST", "/api/v1/admin/login", map[string]string{"email": email})
	token := mustGetString(t, loginResp, "data", "token")

	// --- Act ---
	st, resp := ts.reqWithToken(t, token, "GET", "/api/v1/templates", nil)
	t.Logf("GET /templates http=%d", st)
	// --- Assert ---
	assertStatus(t, st, http.StatusOK, "list templates")
	foundOwn := false
	for _, x := range mustGetArray(t, resp, "data") {
		id := fmt.Sprintf("%.0f", x.(map[string]interface{})["id"].(float64))
		if id == tmplOtherID {
			t.Fatalf("other org template %s must not appear", tmplOtherID)
		}
		if id == tmplOwnID {
			foundOwn = true
		}
	}
	if !foundOwn {
		t.Fatalf("expected own template %s in list", tmplOwnID)
	}
}

// 目的: マトリクス GET /templates/:id — 他組織プロジェクトのテンプレは 404。
func TestTemplate_Get_OtherOrg_AsOrgUser_NotFound(t *testing.T) {
	ts := newTestServer(t)
	_, otherOrgResp := ts.req(t, "POST", "/api/v1/organizations", map[string]interface{}{"name": "TM取得他社"})
	otherOrgID := mustGetString(t, otherOrgResp, "data", "id")
	seed := createTestUser(t, ts, "S", "tmpl-get-seed@example.com")
	_, addResp := ts.req(t, "POST", "/api/v1/organizations/"+otherOrgID+"/users", map[string]interface{}{"user_id": seed})
	otherOwner := mustGetString(t, addResp, "data", "id")
	st, prResp := ts.req(t, "POST", "/api/v1/projects", map[string]interface{}{
		"key": "TG", "name": "他社", "owner_id": otherOwner, "organization_id": otherOrgID,
	})
	assertStatus(t, st, http.StatusCreated, "proj")
	otherProj := mustGetString(t, prResp, "data", "id")
	tmplID := createTestTemplate(t, ts, otherProj, "秘密テンプレ")

	email := "tmpl-get-tenant@example.com"
	u := createTestUser(t, ts, "U", email)
	ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/users", map[string]interface{}{"user_id": u})
	_, loginResp := ts.reqNoAuth(t, "POST", "/api/v1/admin/login", map[string]string{"email": email})
	token := mustGetString(t, loginResp, "data", "token")

	st, _ = ts.reqWithToken(t, token, "GET", "/api/v1/templates/"+tmplID, nil)
	t.Logf("GET other template http=%d", st)
	assertStatus(t, st, http.StatusNotFound, "GET other org template")
}

// 目的: マトリクス POST /templates — 他組織 project_id では 403。
func TestTemplate_Create_OtherOrgProject_Forbidden(t *testing.T) {
	ts := newTestServer(t)
	_, otherOrgResp := ts.req(t, "POST", "/api/v1/organizations", map[string]interface{}{"name": "TM作他社"})
	otherOrgID := mustGetString(t, otherOrgResp, "data", "id")
	seed := createTestUser(t, ts, "S", "tmpl-post-seed@example.com")
	_, addResp := ts.req(t, "POST", "/api/v1/organizations/"+otherOrgID+"/users", map[string]interface{}{"user_id": seed})
	otherOwner := mustGetString(t, addResp, "data", "id")
	st, prResp := ts.req(t, "POST", "/api/v1/projects", map[string]interface{}{
		"key": "TP", "name": "他社P", "owner_id": otherOwner, "organization_id": otherOrgID,
	})
	assertStatus(t, st, http.StatusCreated, "proj")
	otherProj := mustGetString(t, prResp, "data", "id")

	email := "tmpl-post-tenant@example.com"
	u := createTestUser(t, ts, "U", email)
	ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/users", map[string]interface{}{"user_id": u})
	_, loginResp := ts.reqNoAuth(t, "POST", "/api/v1/admin/login", map[string]string{"email": email})
	token := mustGetString(t, loginResp, "data", "token")

	st, _ = ts.reqWithToken(t, token, "POST", "/api/v1/templates", map[string]interface{}{
		"project_id": otherProj,
		"name":       "不正テンプレ",
	})
	t.Logf("POST template other project http=%d", st)
	assertStatus(t, st, http.StatusForbidden, "POST template other org project")
}
