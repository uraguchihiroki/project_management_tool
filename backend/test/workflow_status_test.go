package test

import (
	"fmt"
	"math"
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

func TestWorkflowStatuses_EntryTerminalMarkers(t *testing.T) {
	ts := newTestServer(t)
	wfID := createTestWorkflow(t, ts, "entry-terminal-WF")

	_, r1 := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/statuses", map[string]interface{}{
		"name": "Alpha", "color": "#111111",
	})
	id1 := mustGetString(t, r1, "data", "id")
	do1 := int(mustGetFloat(t, r1, "data", "display_order"))
	_, r2 := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/statuses", map[string]interface{}{
		"name": "Beta", "color": "#222222",
	})
	id2 := mustGetString(t, r2, "data", "id")
	do2 := int(mustGetFloat(t, r2, "data", "display_order"))

	t.Run("is_entry と is_terminal の同時指定は400", func(t *testing.T) {
		st, _ := ts.req(t, "PUT", "/api/v1/statuses/"+id1, map[string]interface{}{
			"display_order": do1,
			"is_entry":      true,
			"is_terminal":   true,
		})
		assertStatus(t, st, http.StatusBadRequest, "entry and terminal on same status")
	})

	t.Run("同一WFで開始は1件だけ保持される", func(t *testing.T) {
		s1, _ := ts.req(t, "PUT", "/api/v1/statuses/"+id1, map[string]interface{}{
			"display_order": do1,
			"is_entry":      true,
		})
		assertStatus(t, s1, http.StatusOK, "set entry on Alpha")
		s2, _ := ts.req(t, "PUT", "/api/v1/statuses/"+id2, map[string]interface{}{
			"display_order": do2,
			"is_entry":      true,
		})
		assertStatus(t, s2, http.StatusOK, "move entry to Beta")
		_, list := ts.req(t, "GET", "/api/v1/workflows/"+wfID+"/statuses", nil)
		arr := mustGetArray(t, list, "data")
		var alphaEntry, betaEntry bool
		for _, x := range arr {
			m := x.(map[string]interface{})
			id := m["id"].(string)
			ie, ok := m["is_entry"].(bool)
			if !ok {
				t.Fatalf("is_entry missing or not bool for %v", m)
			}
			switch id {
			case id1:
				alphaEntry = ie
			case id2:
				betaEntry = ie
			}
		}
		if alphaEntry || !betaEntry {
			t.Fatalf("expected only Beta is_entry: alpha=%v beta=%v", alphaEntry, betaEntry)
		}
	})

	t.Run("終了は複数立てられる", func(t *testing.T) {
		_, r3 := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/statuses", map[string]interface{}{
			"name": "Gamma", "color": "#333333",
		})
		id3 := mustGetString(t, r3, "data", "id")
		do3 := int(mustGetFloat(t, r3, "data", "display_order"))
		sa, _ := ts.req(t, "PUT", "/api/v1/statuses/"+id1, map[string]interface{}{
			"display_order": do1,
			"is_terminal":   true,
		})
		assertStatus(t, sa, http.StatusOK, "terminal Alpha")
		sb, _ := ts.req(t, "PUT", "/api/v1/statuses/"+id3, map[string]interface{}{
			"display_order": do3,
			"is_terminal":   true,
		})
		assertStatus(t, sb, http.StatusOK, "terminal Gamma")
	})
}

// 組織Issueブートストラップの3ステータスでは display_order 最小の1件だけ is_entry になる。
func TestOrgIssueWorkflow_DefaultEntryIsMinDisplayOrder(t *testing.T) {
	ts := newTestServer(t)
	st, wfListResp := ts.req(t, "GET", "/api/v1/workflows?org_id="+testOrgID, nil)
	assertStatus(t, st, http.StatusOK, "list workflows")
	workflows := mustGetArray(t, wfListResp, "data")
	var issueWfID string
	for _, w := range workflows {
		m := w.(map[string]interface{})
		if m["name"].(string) == "組織Issue" {
			issueWfID = fmt.Sprintf("%.0f", m["id"].(float64))
			break
		}
	}
	if issueWfID == "" {
		t.Fatal("組織Issue workflow not found")
	}
	st2, statResp := ts.req(t, "GET", "/api/v1/workflows/"+issueWfID+"/statuses", nil)
	assertStatus(t, st2, http.StatusOK, "GET 組織Issue statuses")
	arr := mustGetArray(t, statResp, "data")
	if len(arr) != 3 {
		t.Fatalf("組織Issue は3ステータス想定, got %d", len(arr))
	}
	minDO := math.MaxFloat64
	entryCount := 0
	var entryName string
	for _, x := range arr {
		m := x.(map[string]interface{})
		do := m["display_order"].(float64)
		if do < minDO {
			minDO = do
		}
		ie, ok := m["is_entry"].(bool)
		if !ok {
			t.Fatalf("is_entry missing or not bool: %v", m)
		}
		if ie {
			entryCount++
			entryName = m["name"].(string)
		}
	}
	if entryCount != 1 {
		t.Fatalf("is_entry true は 1 件期待, got %d", entryCount)
	}
	if entryName != "未着手" {
		t.Fatalf("開始は未着手（display_order 最小）期待, got name=%q minDO=%v", entryName, minDO)
	}
}
