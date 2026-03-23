package test

import (
	"net/http"
	"testing"
)

// 目的: マトリクス GET /users — 組織ユーザーは自組織のユーザーだけ一覧に含むこと。
func TestUser_List_OrgUser_OnlyOwnOrg(t *testing.T) {
	ts := newTestServer(t)
	u1 := createTestUser(t, ts, "U1", "user-tenant-a@example.com")
	_ = u1
	_, otherOrgResp := ts.req(t, "POST", "/api/v1/organizations", map[string]interface{}{"name": "ユーザー他社"})
	otherOrgID := mustGetString(t, otherOrgResp, "data", "id")
	_, addResp := ts.req(t, "POST", "/api/v1/organizations/"+otherOrgID+"/users", map[string]interface{}{
		"user_id": u1,
	})
	otherUserID := mustGetString(t, addResp, "data", "id")
	_ = otherUserID

	email := "user-list-tenant@example.com"
	listUID := createTestUser(t, ts, "一覧", email)
	ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/users", map[string]interface{}{"user_id": listUID})
	_, loginResp := ts.reqNoAuth(t, "POST", "/api/v1/admin/login", map[string]string{"email": email})
	token := mustGetString(t, loginResp, "data", "token")

	st, resp := ts.reqWithToken(t, token, "GET", "/api/v1/users", nil)
	t.Logf("GET /users http=%d", st)
	assertStatus(t, st, http.StatusOK, "GET /users")
	for i, x := range mustGetArray(t, resp, "data") {
		oid := x.(map[string]interface{})["organization_id"].(string)
		if oid != testOrgID {
			t.Fatalf("user[%d] organization_id=%q want %q", i, oid, testOrgID)
		}
	}
}

// 目的: マトリクス GET /users/:id — 他組織ユーザーは 404。
func TestUser_Get_OtherOrg_AsOrgUser_NotFound(t *testing.T) {
	ts := newTestServer(t)
	_, otherOrgResp := ts.req(t, "POST", "/api/v1/organizations", map[string]interface{}{"name": "ユーザー取得他社"})
	otherOrgID := mustGetString(t, otherOrgResp, "data", "id")
	seed := createTestUser(t, ts, "S", "user-get-seed@example.com")
	_, addResp := ts.req(t, "POST", "/api/v1/organizations/"+otherOrgID+"/users", map[string]interface{}{"user_id": seed})
	otherUserID := mustGetString(t, addResp, "data", "id")

	email := "user-get-tenant@example.com"
	u := createTestUser(t, ts, "U", email)
	ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/users", map[string]interface{}{"user_id": u})
	_, loginResp := ts.reqNoAuth(t, "POST", "/api/v1/admin/login", map[string]string{"email": email})
	token := mustGetString(t, loginResp, "data", "token")

	st, _ := ts.reqWithToken(t, token, "GET", "/api/v1/users/"+otherUserID, nil)
	t.Logf("GET other user http=%d", st)
	assertStatus(t, st, http.StatusNotFound, "GET other org user")
}
