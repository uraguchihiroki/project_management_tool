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
		"name":           name,
		"description":    "テスト用ワークフロー",
	})
	assertStatus(t, status, http.StatusCreated, fmt.Sprintf("createWorkflow(%s)", name))
	return fmt.Sprintf("%.0f", mustGetFloat(t, resp, "data", "id"))
}

// ブラウザと同じ経路: POST /admin/login で得た JWT（組織ユーザー）で POST /workflows する。
// スーパー管理者トークンではなく、一般ログインのトークンで検証する。
func TestWorkflow_Create_AsOrgUserJWT(t *testing.T) {
	ts := newTestServer(t)
	email := "wf-as-org-user@example.com"
	createTestUser(t, ts, "WF検証ユーザー", email)

	_, loginResp := ts.reqNoAuth(t, "POST", "/api/v1/admin/login", map[string]string{"email": email})
	token := mustGetString(t, loginResp, "data", "token")

	status, resp := ts.reqWithToken(t, token, "POST", "/api/v1/workflows", map[string]interface{}{
		"organization_id": testOrgID,
		"name":              "ログインユーザーが追加するフロー",
		"description":       "ブラウザと同じJWT経路の検証",
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
			"name":           "通常承認フロー",
			"description":    "一般的な承認フロー",
		})
		assertStatus(t, status, http.StatusCreated, "create workflow")
		assertField(t, mustGetString(t, resp, "data", "name"), "通常承認フロー", "name")
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
	projectID := createTestProject(t, ts, "WF", "テストプロジェクト", ownerID)
	statusIDs := getStatusIDs(t, ts, projectID)
	if len(statusIDs) < 1 {
		t.Fatal("project needs at least 1 status")
	}

	wf1 := createTestWorkflow(t, ts, "フロー1")
	wf2 := createTestWorkflow(t, ts, "フロー2")
	// ユーザーステップを1つ以上持つワークフローのみ一覧に表示される
	ts.req(t, "POST", "/api/v1/workflows/"+wf1+"/steps", map[string]interface{}{
		"status_id": statusIDs[0], "threshold": 10,
	})
	ts.req(t, "POST", "/api/v1/workflows/"+wf2+"/steps", map[string]interface{}{
		"status_id": statusIDs[0], "threshold": 10,
	})

	t.Run("全ワークフロー一覧を取得できる", func(t *testing.T) {
		status, resp := ts.req(t, "GET", "/api/v1/workflows", nil)
		assertStatus(t, status, http.StatusOK, "list workflows")
		workflows := mustGetArray(t, resp, "data")
		if len(workflows) != 2 {
			t.Fatalf("expected 2 workflows, got %d", len(workflows))
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

func TestWorkflowStep_AddAndList(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	projectID := createTestProject(t, ts, "WF", "テストプロジェクト", ownerID)
	wfID := createTestWorkflow(t, ts, "ステップテストフロー")
	statusIDs := getStatusIDs(t, ts, projectID)
	if len(statusIDs) < 2 {
		t.Fatal("project needs at least 2 statuses")
	}

	t.Run("ステップを追加できる", func(t *testing.T) {
		status, resp := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/steps", map[string]interface{}{
			"status_id": statusIDs[0],
			"threshold": 10,
		})
		assertStatus(t, status, http.StatusCreated, "add step")
		assertField(t, mustGetString(t, resp, "data", "status_id"), statusIDs[0], "status_id")
		assertField(t, mustGetString(t, resp, "data", "status", "name"), "未着手", "status.name")
	})

	t.Run("複数ステップを追加できる", func(t *testing.T) {
		ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/steps", map[string]interface{}{
			"status_id": statusIDs[1],
			"threshold": 10,
		})

		_, wfResp := ts.req(t, "GET", "/api/v1/workflows/"+wfID, nil)
		steps := wfResp["data"].(map[string]interface{})["steps"].([]interface{})
		// sts_start + user1 + user2 + sts_goal = 4 steps
		if len(steps) != 4 {
			t.Fatalf("expected 4 steps, got %d", len(steps))
		}
	})
}

func TestWorkflowStep_Update(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	projectID := createTestProject(t, ts, "WF", "テストプロジェクト", ownerID)
	wfID := createTestWorkflow(t, ts, "ステップ更新テスト")
	statusIDs := getStatusIDs(t, ts, projectID)

	_, stepResp := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/steps", map[string]interface{}{
		"status_id": statusIDs[0],
		"threshold": 10,
	})
	stepID := fmt.Sprintf("%.0f", mustGetFloat(t, stepResp, "data", "id"))

	t.Run("ステップを更新できる", func(t *testing.T) {
		status, resp := ts.req(t, "PUT", "/api/v1/workflows/"+wfID+"/steps/"+stepID, map[string]interface{}{
			"description": "更新後の説明",
			"threshold":   15,
		})
		assertStatus(t, status, http.StatusOK, "update step")
		assertField(t, mustGetString(t, resp, "data", "description"), "更新後の説明", "description")
		if mustGetFloat(t, resp, "data", "threshold") != 15 {
			t.Errorf("threshold = %v, want 15", mustGetFloat(t, resp, "data", "threshold"))
		}
	})
}

func TestWorkflowStep_Reorder(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	projectID := createTestProject(t, ts, "WF", "テストプロジェクト", ownerID)
	wfID := createTestWorkflow(t, ts, "ステップ並び替えテスト")
	statusIDs := getStatusIDs(t, ts, projectID)

	_, s1 := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/steps", map[string]interface{}{
		"status_id": statusIDs[0], "threshold": 10,
	})
	_, s2 := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/steps", map[string]interface{}{
		"status_id": statusIDs[1], "threshold": 10,
	})
	stepID1 := uint(mustGetFloat(t, s1, "data", "id"))
	stepID2 := uint(mustGetFloat(t, s2, "data", "id"))

	status, _ := ts.req(t, "PUT", "/api/v1/workflows/"+wfID+"/steps/reorder", map[string]interface{}{
		"ids": []uint{stepID2, stepID1},
	})
	assertStatus(t, status, http.StatusNoContent, "reorder steps")

	_, wfResp := ts.req(t, "GET", "/api/v1/workflows/"+wfID, nil)
	steps := wfResp["data"].(map[string]interface{})["steps"].([]interface{})
	// sts_start + user1 + user2 + sts_goal = 4 steps
	if len(steps) != 4 {
		t.Fatalf("expected 4 steps, got %d", len(steps))
	}
	// 並び替え後も2ステップ存在することを確認
	s0 := steps[0].(map[string]interface{})
	s1m := steps[1].(map[string]interface{})
	if s0["id"] == nil || s1m["id"] == nil {
		t.Error("steps should have id")
	}
}

func TestWorkflowStep_Delete(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	projectID := createTestProject(t, ts, "WF", "テストプロジェクト", ownerID)
	wfID := createTestWorkflow(t, ts, "ステップ削除テスト")
	statusIDs := getStatusIDs(t, ts, projectID)

	_, stepResp := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/steps", map[string]interface{}{
		"status_id": statusIDs[0], "next_status_id": statusIDs[1], "threshold": 10,
	})
	stepID := fmt.Sprintf("%.0f", mustGetFloat(t, stepResp, "data", "id"))

	t.Run("ステップを削除できる", func(t *testing.T) {
		status, _ := ts.req(t, "DELETE", "/api/v1/workflows/"+wfID+"/steps/"+stepID, nil)
		assertStatus(t, status, http.StatusNoContent, "delete step")
	})

	t.Run("削除後はワークフロー詳細にユーザーステップが含まれない", func(t *testing.T) {
		_, wfResp := ts.req(t, "GET", "/api/v1/workflows/"+wfID, nil)
		data := wfResp["data"].(map[string]interface{})
		steps, _ := data["steps"].([]interface{})
		// ユーザーステップを削除すると sts_start + sts_goal のみ残る（2ステップ）
		if len(steps) != 2 {
			t.Errorf("expected 2 steps (sts_start+sts_goal) after user step delete, got %d", len(steps))
		}
	})
}

func TestWorkflow_Reorder(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	projectID := createTestProject(t, ts, "WF", "テストプロジェクト", ownerID)
	statusIDs := getStatusIDs(t, ts, projectID)
	if len(statusIDs) < 1 {
		t.Fatal("project needs at least 1 status")
	}

	wfID1 := createTestWorkflow(t, ts, "フロー1")
	wfID2 := createTestWorkflow(t, ts, "フロー2")
	ts.req(t, "POST", "/api/v1/workflows/"+wfID1+"/steps", map[string]interface{}{
		"status_id": statusIDs[0], "threshold": 10,
	})
	ts.req(t, "POST", "/api/v1/workflows/"+wfID2+"/steps", map[string]interface{}{
		"status_id": statusIDs[0], "threshold": 10,
	})

	id1, _ := strconv.ParseUint(wfID1, 10, 64)
	id2, _ := strconv.ParseUint(wfID2, 10, 64)

	status, _ := ts.req(t, "PUT", "/api/v1/workflows/reorder", map[string]interface{}{
		"ids": []uint{uint(id2), uint(id1)},
	})
	assertStatus(t, status, http.StatusNoContent, "reorder workflows")

	status, listResp := ts.req(t, "GET", "/api/v1/workflows", nil)
	assertStatus(t, status, http.StatusOK, "list after reorder")
	wfs := mustGetArray(t, listResp, "data")
	if len(wfs) != 2 {
		t.Fatalf("expected 2 workflows, got %d", len(wfs))
	}
	assertField(t, mustGetString(t, wfs[0].(map[string]interface{}), "name"), "フロー2", "first after reorder")
	assertField(t, mustGetString(t, wfs[1].(map[string]interface{}), "name"), "フロー1", "second after reorder")
}

func TestWorkflow_DeleteCascade(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	projectID := createTestProject(t, ts, "WF", "テストプロジェクト", ownerID)
	wfID := createTestWorkflow(t, ts, "カスケード削除テスト")
	statusIDs := getStatusIDs(t, ts, projectID)
	ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/steps", map[string]interface{}{
		"status_id": statusIDs[0], "next_status_id": statusIDs[1], "threshold": 10,
	})
	ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/steps", map[string]interface{}{
		"status_id": statusIDs[1], "next_status_id": statusIDs[2], "threshold": 10,
	})

	t.Run("ステップが存在してもワークフローを削除できる", func(t *testing.T) {
		status, _ := ts.req(t, "DELETE", "/api/v1/workflows/"+wfID, nil)
		assertStatus(t, status, http.StatusNoContent, "delete workflow with steps")
	})
}
