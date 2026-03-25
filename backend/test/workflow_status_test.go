package test

import (
	"fmt"
	"net/http"
	"testing"
)

func TestWorkflowStatuses_ListAndCreate(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "wf-st-owner@example.com")
	createTestProject(t, ts, "WFS", "ステータスAPI", ownerID)
	wfID := createTestWorkflow(t, ts, "ステータス付きフロー")

	t.Run("新規ワークフローはステータス0件で取得できる", func(t *testing.T) {
		status, resp := ts.req(t, "GET", "/api/v1/workflows/"+wfID+"/statuses", nil)
		assertStatus(t, status, http.StatusOK, "GET workflow statuses empty")
		arr := mustGetArray(t, resp, "data")
		if len(arr) != 0 {
			t.Fatalf("expected 0 statuses, got %d", len(arr))
		}
	})

	t.Run("ステータスを追加できる", func(t *testing.T) {
		status, resp := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/statuses", map[string]interface{}{
			"name":  "カスタム",
			"color": "#FF00AA",
			"type":  "issue",
		})
		assertStatus(t, status, http.StatusCreated, "POST workflow status")
		assertField(t, mustGetString(t, resp, "data", "name"), "カスタム", "name")
	})

	t.Run("追加後は一覧に含まれる", func(t *testing.T) {
		status, resp := ts.req(t, "GET", "/api/v1/workflows/"+wfID+"/statuses", nil)
		assertStatus(t, status, http.StatusOK, "GET workflow statuses after create")
		arr := mustGetArray(t, resp, "data")
		if len(arr) != 1 {
			t.Fatalf("expected 1 status, got %d", len(arr))
		}
		n := arr[0].(map[string]interface{})["name"].(string)
		if n != "カスタム" {
			t.Fatalf("status name = %q, want カスタム", n)
		}
	})
}

func TestWorkflowStatuses_OrgIsolation(t *testing.T) {
	ts := newTestServer(t)

	_, otherOrgResp := ts.req(t, "POST", "/api/v1/organizations", map[string]interface{}{"name": "他社"})
	otherOrgID := mustGetString(t, otherOrgResp, "data", "id")

	status, wfResp := ts.req(t, "POST", "/api/v1/workflows", map[string]interface{}{
		"organization_id": otherOrgID,
		"name":            "他社専用フロー",
		"description":     "",
	})
	assertStatus(t, status, http.StatusCreated, "create workflow in other org")
	otherWfID := fmt.Sprintf("%.0f", mustGetFloat(t, wfResp, "data", "id"))

	userID := createTestUser(t, ts, "社内ユーザー", "wf-st-isolation@example.com")
	ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/users", map[string]interface{}{
		"user_id": userID,
	})
	_, loginResp := ts.reqNoAuth(t, "POST", "/api/v1/admin/login", map[string]string{"email": "wf-st-isolation@example.com"})
	token := mustGetString(t, loginResp, "data", "token")

	t.Run("別組織のワークフローのステータスは404", func(t *testing.T) {
		st, _ := ts.reqWithToken(t, token, "GET", "/api/v1/workflows/"+otherWfID+"/statuses", nil)
		assertStatus(t, st, http.StatusNotFound, "GET other org workflow statuses")
	})

	t.Run("別組織のワークフローへPOSTも404", func(t *testing.T) {
		st, _ := ts.reqWithToken(t, token, "POST", "/api/v1/workflows/"+otherWfID+"/statuses", map[string]interface{}{
			"name": "不正追加",
		})
		assertStatus(t, st, http.StatusNotFound, "POST other org workflow status")
	})
}

// 同一 workflow・同一 (name, display_order) は Service で拒否（DB 業務 UNIQUE は張らない）。
func TestWorkflowStatuses_DuplicateRejectedByService(t *testing.T) {
	ts := newTestServer(t)
	wfID := createTestWorkflow(t, ts, "重複拒否WF")
	body := map[string]interface{}{
		"name":          "かぶり",
		"color":         "#6B7280",
		"display_order": 5,
	}
	st1, _ := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/statuses", body)
	assertStatus(t, st1, http.StatusCreated, "first POST workflow status")
	st2, _ := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/statuses", body)
	assertStatus(t, st2, http.StatusBadRequest, "duplicate (name, display_order) should fail")
}
