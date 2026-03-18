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
		"name":        name,
		"description": "テスト用ワークフロー",
	})
	assertStatus(t, status, http.StatusCreated, fmt.Sprintf("createWorkflow(%s)", name))
	return fmt.Sprintf("%.0f", mustGetFloat(t, resp, "data", "id"))
}

func TestWorkflow_Create(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	createTestProject(t, ts, "WF", "ワークフローテスト", ownerID)

	t.Run("ワークフローを作成できる", func(t *testing.T) {
		status, resp := ts.req(t, "POST", "/api/v1/workflows", map[string]interface{}{
			"name":        "通常承認フロー",
			"description": "一般的な承認フロー",
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
	createTestProject(t, ts, "WF", "テストプロジェクト", ownerID)

	createTestWorkflow(t, ts, "フロー1")
	createTestWorkflow(t, ts, "フロー2")

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
	statusID := getFirstStatusID(t, ts, projectID)

	t.Run("ステップを追加できる", func(t *testing.T) {
		status, resp := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/steps", map[string]interface{}{
			"name":           "上司承認",
			"required_level": 5,
			"status_id":      statusID,
		})
		assertStatus(t, status, http.StatusCreated, "add step")
		assertField(t, mustGetString(t, resp, "data", "name"), "上司承認", "name")
	})

	t.Run("複数ステップを追加するとorderが連番になる", func(t *testing.T) {
		ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/steps", map[string]interface{}{
			"name": "部長承認", "required_level": 7,
		})

		_, wfResp := ts.req(t, "GET", "/api/v1/workflows/"+wfID, nil)
		steps := wfResp["data"].(map[string]interface{})["steps"].([]interface{})
		if len(steps) != 2 {
			t.Fatalf("expected 2 steps, got %d", len(steps))
		}
		// step1のorderが1, step2のorderが2であることを確認
		s1 := steps[0].(map[string]interface{})
		s2 := steps[1].(map[string]interface{})
		if s1["order"].(float64) != 1 {
			t.Errorf("step1 order = %v, want 1", s1["order"])
		}
		if s2["order"].(float64) != 2 {
			t.Errorf("step2 order = %v, want 2", s2["order"])
		}
	})
}

func TestWorkflowStep_Update(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	createTestProject(t, ts, "WF", "テストプロジェクト", ownerID)
	wfID := createTestWorkflow(t, ts, "ステップ更新テスト")

	_, stepResp := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/steps", map[string]interface{}{
		"name": "初期ステップ", "required_level": 3,
	})
	stepID := fmt.Sprintf("%.0f", mustGetFloat(t, stepResp, "data", "id"))

	t.Run("ステップを更新できる", func(t *testing.T) {
		status, resp := ts.req(t, "PUT", "/api/v1/workflows/"+wfID+"/steps/"+stepID, map[string]interface{}{
			"name":           "更新後ステップ",
			"required_level": 7,
		})
		assertStatus(t, status, http.StatusOK, "update step")
		assertField(t, mustGetString(t, resp, "data", "name"), "更新後ステップ", "name")
	})
}

func TestWorkflowStep_Reorder(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	createTestProject(t, ts, "WF", "テストプロジェクト", ownerID)
	wfID := createTestWorkflow(t, ts, "ステップ並び替えテスト")

	_, s1 := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/steps", map[string]interface{}{"name": "ステップ1", "required_level": 5})
	_, s2 := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/steps", map[string]interface{}{"name": "ステップ2", "required_level": 7})
	stepID1 := uint(mustGetFloat(t, s1, "data", "id"))
	stepID2 := uint(mustGetFloat(t, s2, "data", "id"))

	status, _ := ts.req(t, "PUT", "/api/v1/workflows/"+wfID+"/steps/reorder", map[string]interface{}{
		"ids": []uint{stepID2, stepID1},
	})
	assertStatus(t, status, http.StatusNoContent, "reorder steps")

	_, wfResp := ts.req(t, "GET", "/api/v1/workflows/"+wfID, nil)
	steps := wfResp["data"].(map[string]interface{})["steps"].([]interface{})
	if len(steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(steps))
	}
	assertField(t, mustGetString(t, steps[0].(map[string]interface{}), "name"), "ステップ2", "first after reorder")
	assertField(t, mustGetString(t, steps[1].(map[string]interface{}), "name"), "ステップ1", "second after reorder")
}

func TestWorkflowStep_Delete(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	createTestProject(t, ts, "WF", "テストプロジェクト", ownerID)
	wfID := createTestWorkflow(t, ts, "ステップ削除テスト")

	_, stepResp := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/steps", map[string]interface{}{
		"name": "削除対象ステップ", "required_level": 5,
	})
	stepID := fmt.Sprintf("%.0f", mustGetFloat(t, stepResp, "data", "id"))

	t.Run("ステップを削除できる", func(t *testing.T) {
		status, _ := ts.req(t, "DELETE", "/api/v1/workflows/"+wfID+"/steps/"+stepID, nil)
		assertStatus(t, status, http.StatusNoContent, "delete step")
	})

	t.Run("削除後はワークフロー詳細にステップが含まれない", func(t *testing.T) {
		_, wfResp := ts.req(t, "GET", "/api/v1/workflows/"+wfID, nil)
		data := wfResp["data"].(map[string]interface{})
		steps, _ := data["steps"].([]interface{})
		if len(steps) != 0 {
			t.Errorf("expected 0 steps after delete, got %d", len(steps))
		}
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
	if len(wfs) != 2 {
		t.Fatalf("expected 2 workflows, got %d", len(wfs))
	}
	assertField(t, mustGetString(t, wfs[0].(map[string]interface{}), "name"), "フロー2", "first after reorder")
	assertField(t, mustGetString(t, wfs[1].(map[string]interface{}), "name"), "フロー1", "second after reorder")
}

func TestWorkflow_DeleteCascade(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	createTestProject(t, ts, "WF", "テストプロジェクト", ownerID)
	wfID := createTestWorkflow(t, ts, "カスケード削除テスト")

	// ステップを追加してからワークフローを削除
	ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/steps", map[string]interface{}{
		"name": "ステップ1", "required_level": 5,
	})
	ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/steps", map[string]interface{}{
		"name": "ステップ2", "required_level": 7,
	})

	t.Run("ステップが存在してもワークフローを削除できる", func(t *testing.T) {
		status, _ := ts.req(t, "DELETE", "/api/v1/workflows/"+wfID, nil)
		assertStatus(t, status, http.StatusNoContent, "delete workflow with steps")
	})
}
