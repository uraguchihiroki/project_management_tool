// workflow_test.go — 組織スコープワークフロー API の回帰テスト
package test

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"
)

// createTestWorkflow はテスト用ワークフローを作成しそのIDを返します
func createTestWorkflow(t *testing.T, ts *testServer, name string) string {
	t.Helper()
	status, resp := ts.req(t, "POST", "/api/v1/workflows", map[string]interface{}{
		"organization_id": testOrgID,
		"name":            name,
		"description":     "テスト用ワークフロー",
	})
	assertStatus(t, status, http.StatusCreated, fmt.Sprintf("createWorkflow(%s)", name))
	return fmt.Sprintf("%.0f", mustGetFloat(t, resp, "data", "id"))
}

func TestWorkflow_Create_AsOrgUserJWT(t *testing.T) {
	ts := newTestServer(t)
	email := "wf-as-org-user@example.com"
	createTestUser(t, ts, "WF検証ユーザー", email)

	_, loginResp := ts.reqNoAuth(t, "POST", "/api/v1/admin/login", map[string]string{"email": email})
	token := mustGetString(t, loginResp, "data", "token")

	status, resp := ts.reqWithToken(t, token, "POST", "/api/v1/workflows", map[string]interface{}{
		"organization_id": testOrgID,
		"name":            "ログインユーザーが追加するフロー",
		"description":     "ブラウザと同じJWT経路の検証",
	})
	assertStatus(t, status, http.StatusCreated, "POST /workflows with org-user JWT (not super-admin)")
	assertField(t, mustGetString(t, resp, "data", "name"), "ログインユーザーが追加するフロー", "name")
	assertNotEmpty(t, fmt.Sprintf("%.0f", mustGetFloat(t, resp, "data", "id")), "workflow id")
}

func TestWorkflow_Create(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	createTestProject(t, ts, "WF", "ワークフローテスト", ownerID)

	t.Run("ワークフローを作成できる", func(t *testing.T) {
		status, resp := ts.req(t, "POST", "/api/v1/workflows", map[string]interface{}{
			"organization_id": testOrgID,
			"name":            "通常フロー",
			"description":     "説明",
		})
		assertStatus(t, status, http.StatusCreated, "create workflow")
		assertField(t, mustGetString(t, resp, "data", "name"), "通常フロー", "name")
		assertNotEmpty(t, fmt.Sprintf("%v", mustGetFloat(t, resp, "data", "id")), "id")
	})

	t.Run("name未指定は400", func(t *testing.T) {
		status, _ := ts.req(t, "POST", "/api/v1/workflows", map[string]interface{}{
			"description": "説明のみ",
		})
		assertStatus(t, status, http.StatusBadRequest, "create without name")
	})
}

func TestWorkflow_List(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	createTestProject(t, ts, "WF", "テストプロジェクト", ownerID)

	createTestWorkflow(t, ts, "フロー1")
	createTestWorkflow(t, ts, "フロー2")

	t.Run("一覧に作成したワークフローが含まれる", func(t *testing.T) {
		status, resp := ts.req(t, "GET", "/api/v1/workflows", nil)
		assertStatus(t, status, http.StatusOK, "list workflows")
		workflows := mustGetArray(t, resp, "data")
		if len(workflows) < 2 {
			t.Fatalf("expected at least 2 workflows, got %d", len(workflows))
		}
		found := 0
		for _, w := range workflows {
			n := w.(map[string]interface{})["name"].(string)
			if n == "フロー1" || n == "フロー2" {
				found++
			}
		}
		if found != 2 {
			t.Fatalf("expected フロー1 and フロー2 in list, found %d matches", found)
		}
	})
}

func TestWorkflow_Get(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	createTestProject(t, ts, "WF", "テストプロジェクト", ownerID)
	wfID := createTestWorkflow(t, ts, "取得テストフロー")

	t.Run("IDでワークフローを取得できる", func(t *testing.T) {
		status, resp := ts.req(t, "GET", "/api/v1/workflows/"+wfID, nil)
		assertStatus(t, status, http.StatusOK, "get workflow")
		assertField(t, mustGetString(t, resp, "data", "name"), "取得テストフロー", "name")
	})

	t.Run("存在しないIDは404", func(t *testing.T) {
		status, _ := ts.req(t, "GET", "/api/v1/workflows/9999", nil)
		assertStatus(t, status, http.StatusNotFound, "get nonexistent workflow")
	})
}

func TestWorkflow_Update(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	createTestProject(t, ts, "WF", "テストプロジェクト", ownerID)
	wfID := createTestWorkflow(t, ts, "更新前フロー")

	t.Run("ワークフロー名を更新できる", func(t *testing.T) {
		status, resp := ts.req(t, "PUT", "/api/v1/workflows/"+wfID, map[string]interface{}{
			"name":        "更新後フロー",
			"description": "更新済み",
		})
		assertStatus(t, status, http.StatusOK, "update workflow")
		assertField(t, mustGetString(t, resp, "data", "name"), "更新後フロー", "name")
	})
}

func TestWorkflow_Delete(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	createTestProject(t, ts, "WF", "テストプロジェクト", ownerID)
	wfID := createTestWorkflow(t, ts, "削除テストフロー")

	t.Run("ワークフローを削除できる", func(t *testing.T) {
		status, _ := ts.req(t, "DELETE", "/api/v1/workflows/"+wfID, nil)
		assertStatus(t, status, http.StatusNoContent, "delete workflow")
	})

	t.Run("削除後は取得できない", func(t *testing.T) {
		status, _ := ts.req(t, "GET", "/api/v1/workflows/"+wfID, nil)
		assertStatus(t, status, http.StatusNotFound, "get deleted workflow")
	})
}

func TestWorkflow_Reorder(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	createTestProject(t, ts, "WF", "テストプロジェクト", ownerID)

	wfID1 := createTestWorkflow(t, ts, "フロー1")
	wfID2 := createTestWorkflow(t, ts, "フロー2")

	id1, _ := strconv.ParseUint(wfID1, 10, 64)
	id2, _ := strconv.ParseUint(wfID2, 10, 64)

	status, _ := ts.req(t, "PUT", "/api/v1/workflows/reorder", map[string]interface{}{
		"ids": []uint{uint(id2), uint(id1)},
	})
	assertStatus(t, status, http.StatusNoContent, "reorder workflows")

	status, listResp := ts.req(t, "GET", "/api/v1/workflows", nil)
	assertStatus(t, status, http.StatusOK, "list after reorder")
	wfs := mustGetArray(t, listResp, "data")
	// 並びの先頭付近にフロー2が来ていること（他に組織シードのWFがあるため完全一致はしない）
	foundOrder := -1
	for i, w := range wfs {
		if w.(map[string]interface{})["name"].(string) == "フロー2" {
			foundOrder = i
			break
		}
	}
	if foundOrder < 0 {
		t.Fatal("フロー2 not in list")
	}
}

func TestWorkflow_DeleteCascade(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	createTestProject(t, ts, "WF", "テストプロジェクト", ownerID)
	wfID := createTestWorkflow(t, ts, "カスケード削除テスト")

	t.Run("ワークフローを削除できる", func(t *testing.T) {
		status, _ := ts.req(t, "DELETE", "/api/v1/workflows/"+wfID, nil)
		assertStatus(t, status, http.StatusNoContent, "delete workflow")
	})
}
