package test

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
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

// レガシーDB等で同一 workflow_id に (name,type,order) 重複行がある場合、一覧は1件にまとめる
func TestWorkflowStatuses_List_DedupesDuplicateNameTypeOrder(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "wf-st-dedup@example.com")
	createTestProject(t, ts, "WFD", "重複テスト", ownerID)
	wfID := createTestWorkflow(t, ts, "重複WF")
	wid64, err := strconv.ParseUint(wfID, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	wid := uint(wid64)
	for i := 0; i < 3; i++ {
		sid := uuid.New()
		if err := ts.db.Create(&model.Status{
			ID:         sid,
			Key:        "sts-dup-" + sid.String(),
			WorkflowID: wid,
			Name:       "未着手",
			Color:      "#6B7280",
			Order:      1,
			Type:       "issue",
		}).Error; err != nil {
			t.Fatal(err)
		}
	}
	status, resp := ts.req(t, "GET", "/api/v1/workflows/"+wfID+"/statuses", nil)
	assertStatus(t, status, http.StatusOK, "GET with dup rows")
	arr := mustGetArray(t, resp, "data")
	dup := 0
	for _, x := range arr {
		m := x.(map[string]interface{})
		if m["name"].(string) == "未着手" && int(m["order"].(float64)) == 1 {
			dup++
		}
	}
	if dup != 1 {
		t.Fatalf("expected 1 未着手 after dedupe, got %d (total rows %d)", dup, len(arr))
	}
}
