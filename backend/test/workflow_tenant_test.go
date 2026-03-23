package test

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"
)

// 目的: マトリクス GET /workflows（非SA）— 一覧に他組織の workflow が混ざらないこと。
// 期待: 各要素の organization_id は JWT の組織（testOrgID）のみ。
func TestWorkflow_List_OrgUser_OnlyOwnOrg(t *testing.T) {
	ts := newTestServer(t)
	email := "wf-list-tenant@example.com"
	createTestUser(t, ts, "一覧テナント", email)
	_, loginResp := ts.reqNoAuth(t, "POST", "/api/v1/admin/login", map[string]string{"email": email})
	token := mustGetString(t, loginResp, "data", "token")

	_, otherOrgResp := ts.req(t, "POST", "/api/v1/organizations", map[string]interface{}{"name": "別テナント社"})
	otherOrgID := mustGetString(t, otherOrgResp, "data", "id")

	ts.req(t, "POST", "/api/v1/workflows", map[string]interface{}{
		"organization_id": otherOrgID,
		"name":            "他社フロー",
		"description":     "",
	})
	createTestWorkflow(t, ts, "自社フロー一覧用")

	t.Run("一覧の organization_id はすべて JWT の組織", func(t *testing.T) {
		status, resp := ts.reqWithToken(t, token, "GET", "/api/v1/workflows", nil)
		t.Logf("GET /workflows http=%d", status)
		assertStatus(t, status, http.StatusOK, "GET /workflows as org user")
		workflows := mustGetArray(t, resp, "data")
		for i, w := range workflows {
			oid, ok := w.(map[string]interface{})["organization_id"].(string)
			if !ok || oid == "" {
				t.Fatalf("workflow[%d] missing organization_id", i)
			}
			if oid != testOrgID {
				t.Fatalf("workflow[%d] organization_id=%q, want only %q (tenant leak)", i, oid, testOrgID)
			}
		}
	})

	t.Run("org_id が JWT と異なると403", func(t *testing.T) {
		status, _ := ts.reqWithToken(t, token, "GET", "/api/v1/workflows?org_id="+otherOrgID, nil)
		assertStatus(t, status, http.StatusForbidden, "GET /workflows wrong org_id query")
	})

	t.Run("org_id が JWT と一致すれば200", func(t *testing.T) {
		status, resp := ts.reqWithToken(t, token, "GET", "/api/v1/workflows?org_id="+testOrgID, nil)
		assertStatus(t, status, http.StatusOK, "GET /workflows matching org_id")
		workflows := mustGetArray(t, resp, "data")
		for i, w := range workflows {
			oid := w.(map[string]interface{})["organization_id"].(string)
			if oid != testOrgID {
				t.Fatalf("workflow[%d] organization_id=%q", i, oid)
			}
		}
	})
}

// TestWorkflow_List_SuperAdmin_OrgQuery はスーパーアドミンが org_id で絞れることと、未指定で全件であることを検証する
func TestWorkflow_List_SuperAdmin_OrgQuery(t *testing.T) {
	ts := newTestServer(t)
	_, otherOrgResp := ts.req(t, "POST", "/api/v1/organizations", map[string]interface{}{"name": "SA絞り込み用"})
	otherOrgID := mustGetString(t, otherOrgResp, "data", "id")
	ts.req(t, "POST", "/api/v1/workflows", map[string]interface{}{
		"organization_id": otherOrgID,
		"name":            "SA他社WF",
		"description":     "",
	})

	t.Run("org_id 指定でその組織のみ", func(t *testing.T) {
		status, resp := ts.req(t, "GET", "/api/v1/workflows?org_id="+otherOrgID, nil)
		assertStatus(t, status, http.StatusOK, "GET /workflows?org_id= as super admin")
		workflows := mustGetArray(t, resp, "data")
		for i, w := range workflows {
			oid := w.(map[string]interface{})["organization_id"].(string)
			if oid != otherOrgID {
				t.Fatalf("workflow[%d] organization_id=%q want %q", i, oid, otherOrgID)
			}
		}
	})

	t.Run("org_id なしは全組織分を含む", func(t *testing.T) {
		status, resp := ts.req(t, "GET", "/api/v1/workflows", nil)
		assertStatus(t, status, http.StatusOK, "GET /workflows all orgs")
		workflows := mustGetArray(t, resp, "data")
		if len(workflows) < 2 {
			t.Fatalf("expected workflows from multiple orgs, got %d", len(workflows))
		}
		hasTest := false
		hasOther := false
		for _, w := range workflows {
			oid := w.(map[string]interface{})["organization_id"].(string)
			if oid == testOrgID {
				hasTest = true
			}
			if oid == otherOrgID {
				hasOther = true
			}
		}
		if !hasTest || !hasOther {
			t.Fatalf("expected both test org and other org workflows in unfiltered list: test=%v other=%v", hasTest, hasOther)
		}
	})

	t.Run("無効な org_id は400", func(t *testing.T) {
		status, _ := ts.req(t, "GET", "/api/v1/workflows?org_id=not-a-uuid", nil)
		assertStatus(t, status, http.StatusBadRequest, "invalid org_id")
	})
}

// TestWorkflow_List_InvalidOrgQuery は不正な org_id クエリを検証する
func TestWorkflow_List_InvalidOrgQuery(t *testing.T) {
	ts := newTestServer(t)
	email := "wf-bad-org-q@example.com"
	createTestUser(t, ts, "ユーザー", email)
	_, loginResp := ts.reqNoAuth(t, "POST", "/api/v1/admin/login", map[string]string{"email": email})
	token := mustGetString(t, loginResp, "data", "token")

	status, _ := ts.reqWithToken(t, token, "GET", "/api/v1/workflows?org_id=not-a-valid-uuid", nil)
	assertStatus(t, status, http.StatusBadRequest, "invalid org_id org user")
}

// 目的: マトリクス GET/PUT/DELETE /workflows/:id — 他組織のワークフローを組織ユーザーが取得・更新・削除できないこと。
// 期待: いずれも 404（workflow not found）。
func TestWorkflow_Detail_OtherOrg_AsOrgUser(t *testing.T) {
	ts := newTestServer(t)
	// --- Arrange ---
	_, otherOrgResp := ts.req(t, "POST", "/api/v1/organizations", map[string]interface{}{"name": "WF単体他社"})
	otherOrgID := mustGetString(t, otherOrgResp, "data", "id")
	st, wfResp := ts.req(t, "POST", "/api/v1/workflows", map[string]interface{}{
		"organization_id": otherOrgID,
		"name":            "他社WF単体",
		"description":     "",
	})
	assertStatus(t, st, http.StatusCreated, "create wf other org")
	otherWfID := fmt.Sprintf("%.0f", mustGetFloat(t, wfResp, "data", "id"))

	email := "wf-detail-tenant@example.com"
	userID := createTestUser(t, ts, "ユーザー", email)
	ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/users", map[string]interface{}{"user_id": userID})
	_, loginResp := ts.reqNoAuth(t, "POST", "/api/v1/admin/login", map[string]string{"email": email})
	token := mustGetString(t, loginResp, "data", "token")

	// --- Act / Assert ---
	t.Run("GET", func(t *testing.T) {
		st, _ := ts.reqWithToken(t, token, "GET", "/api/v1/workflows/"+otherWfID, nil)
		t.Logf("GET other wf http=%d", st)
		assertStatus(t, st, http.StatusNotFound, "GET other org workflow")
	})
	t.Run("PUT", func(t *testing.T) {
		st, _ := ts.reqWithToken(t, token, "PUT", "/api/v1/workflows/"+otherWfID, map[string]interface{}{
			"name": "hack", "description": "",
		})
		t.Logf("PUT other wf http=%d", st)
		assertStatus(t, st, http.StatusNotFound, "PUT other org workflow")
	})
	t.Run("DELETE", func(t *testing.T) {
		st, _ := ts.reqWithToken(t, token, "DELETE", "/api/v1/workflows/"+otherWfID, nil)
		t.Logf("DELETE other wf http=%d", st)
		assertStatus(t, st, http.StatusNotFound, "DELETE other org workflow")
	})
}

// 目的: マトリクス POST /workflows — 組織ユーザーは body の organization_id に騙されず JWT の組織にだけ作成できること。
// 期待: 201 かつ data.organization_id は testOrgID。
func TestWorkflow_Create_OrgUser_IgnoresForeignOrganizationID(t *testing.T) {
	ts := newTestServer(t)
	_, otherOrgResp := ts.req(t, "POST", "/api/v1/organizations", map[string]interface{}{"name": "WF作成騙し他社"})
	otherOrgID := mustGetString(t, otherOrgResp, "data", "id")

	email := "wf-create-tenant@example.com"
	userID := createTestUser(t, ts, "ユーザー", email)
	ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/users", map[string]interface{}{"user_id": userID})
	_, loginResp := ts.reqNoAuth(t, "POST", "/api/v1/admin/login", map[string]string{"email": email})
	token := mustGetString(t, loginResp, "data", "token")

	// --- Act ---
	st, resp := ts.reqWithToken(t, token, "POST", "/api/v1/workflows", map[string]interface{}{
		"organization_id": otherOrgID,
		"name":            "JWT側に作られるべき",
		"description":     "",
	})
	t.Logf("POST /workflows http=%d", st)
	// --- Assert ---
	assertStatus(t, st, http.StatusCreated, "POST workflow as org user")
	gotOrg := mustGetString(t, resp, "data", "organization_id")
	if gotOrg != testOrgID {
		t.Fatalf("organization_id=%q, want JWT org %q", gotOrg, testOrgID)
	}
}

// 目的: マトリクス PUT /workflows/reorder — 並び替えに他組織ワークフロー ID を混ぜると拒否されること。
// 期待: 403。
func TestWorkflow_Reorder_MixedOrgs_AsOrgUser_Forbidden(t *testing.T) {
	ts := newTestServer(t)
	wfOwn1 := createTestWorkflow(t, ts, "並替自社1")
	_, otherOrgResp := ts.req(t, "POST", "/api/v1/organizations", map[string]interface{}{"name": "WF並替他社"})
	otherOrgID := mustGetString(t, otherOrgResp, "data", "id")
	st, wfResp := ts.req(t, "POST", "/api/v1/workflows", map[string]interface{}{
		"organization_id": otherOrgID,
		"name":            "並替他社WF",
		"description":     "",
	})
	assertStatus(t, st, http.StatusCreated, "wf other org")
	otherWfID := fmt.Sprintf("%.0f", mustGetFloat(t, wfResp, "data", "id"))
	id1, _ := strconv.ParseUint(wfOwn1, 10, 64)
	idOther, _ := strconv.ParseUint(otherWfID, 10, 64)

	email := "wf-reorder-tenant@example.com"
	userID := createTestUser(t, ts, "ユーザー", email)
	ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/users", map[string]interface{}{"user_id": userID})
	_, loginResp := ts.reqNoAuth(t, "POST", "/api/v1/admin/login", map[string]string{"email": email})
	token := mustGetString(t, loginResp, "data", "token")

	// --- Act ---
	st, _ = ts.reqWithToken(t, token, "PUT", "/api/v1/workflows/reorder", map[string]interface{}{
		"ids": []uint{uint(id1), uint(idOther)},
	})
	t.Logf("PUT reorder mixed http=%d", st)
	// --- Assert ---
	assertStatus(t, st, http.StatusForbidden, "reorder with other org workflow id")
}
