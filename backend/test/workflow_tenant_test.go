package test

import (
	"net/http"
	"testing"
)

// TestWorkflow_List_OrgUser_OnlyOwnOrg は組織ユーザ JWT で GET /workflows の data がすべて testOrgID であることを検証する
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
