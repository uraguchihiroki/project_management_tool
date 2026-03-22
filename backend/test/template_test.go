package test

import (
	"fmt"
	"net/http"
	"testing"
)

func createTestTemplate(t *testing.T, ts *testServer, projectID, name string) string {
	t.Helper()
	body := map[string]interface{}{
		"project_id":       projectID,
		"name":             name,
		"description":      "テスト用テンプレート",
		"body":             "## 概要\n\n## 再現手順\n\n## 期待結果",
		"default_priority": "medium",
	}
	status, resp := ts.req(t, "POST", "/api/v1/templates", body)
	assertStatus(t, status, http.StatusCreated, fmt.Sprintf("createTemplate(%s)", name))
	return fmt.Sprintf("%.0f", mustGetFloat(t, resp, "data", "id"))
}

func TestTemplate_Create(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	projectID := createTestProject(t, ts, "TM", "テンプレートテスト", ownerID)

	t.Run("テンプレートを作成できる", func(t *testing.T) {
		status, resp := ts.req(t, "POST", "/api/v1/templates", map[string]interface{}{
			"project_id":       projectID,
			"name":             "バグ報告",
			"description":      "バグを報告するテンプレート",
			"body":             "## 概要\n\n## 再現手順",
			"default_priority": "high",
		})
		assertStatus(t, status, http.StatusCreated, "create template")
		assertField(t, mustGetString(t, resp, "data", "name"), "バグ報告", "name")
		assertField(t, mustGetString(t, resp, "data", "default_priority"), "high", "default_priority")
	})

	t.Run("name未指定は400", func(t *testing.T) {
		status, _ := ts.req(t, "POST", "/api/v1/templates", map[string]interface{}{
			"project_id": projectID,
		})
		assertStatus(t, status, http.StatusBadRequest, "create without name")
	})

	t.Run("default_priority未指定はmediumになる", func(t *testing.T) {
		status, resp := ts.req(t, "POST", "/api/v1/templates", map[string]interface{}{
			"project_id": projectID,
			"name":       "優先度デフォルトテスト",
		})
		assertStatus(t, status, http.StatusCreated, "create without priority")
		assertField(t, mustGetString(t, resp, "data", "default_priority"), "medium", "default_priority")
	})
}

func TestTemplate_List(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	projectID := createTestProject(t, ts, "TM", "テストプロジェクト", ownerID)

	createTestTemplate(t, ts, projectID, "バグ報告")
	createTestTemplate(t, ts, projectID, "機能要望")

	t.Run("全テンプレート一覧を取得できる", func(t *testing.T) {
		status, resp := ts.req(t, "GET", "/api/v1/templates", nil)
		assertStatus(t, status, http.StatusOK, "list templates")
		templates := mustGetArray(t, resp, "data")
		if len(templates) != 2 {
			t.Fatalf("expected 2 templates, got %d", len(templates))
		}
	})

	t.Run("プロジェクト別テンプレート一覧を取得できる", func(t *testing.T) {
		status, resp := ts.req(t, "GET", "/api/v1/projects/"+projectID+"/templates", nil)
		assertStatus(t, status, http.StatusOK, "list by project")
		templates := mustGetArray(t, resp, "data")
		if len(templates) != 2 {
			t.Fatalf("expected 2 templates, got %d", len(templates))
		}
	})
}

func TestTemplate_Get(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	projectID := createTestProject(t, ts, "TM", "テストプロジェクト", ownerID)
	tmplID := createTestTemplate(t, ts, projectID, "取得テスト")

	t.Run("IDでテンプレートを取得できる", func(t *testing.T) {
		status, resp := ts.req(t, "GET", "/api/v1/templates/"+tmplID, nil)
		assertStatus(t, status, http.StatusOK, "get template")
		assertField(t, mustGetString(t, resp, "data", "name"), "取得テスト", "name")
	})

	t.Run("存在しないIDは404", func(t *testing.T) {
		status, _ := ts.req(t, "GET", "/api/v1/templates/9999", nil)
		assertStatus(t, status, http.StatusNotFound, "get nonexistent")
	})
}

func TestTemplate_Update(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	projectID := createTestProject(t, ts, "TM", "テストプロジェクト", ownerID)
	tmplID := createTestTemplate(t, ts, projectID, "更新前テンプレート")

	t.Run("テンプレートを更新できる", func(t *testing.T) {
		status, resp := ts.req(t, "PUT", "/api/v1/templates/"+tmplID, map[string]interface{}{
			"name":             "更新後テンプレート",
			"description":      "更新済み",
			"body":             "新しい本文",
			"default_priority": "critical",
		})
		assertStatus(t, status, http.StatusOK, "update template")
		assertField(t, mustGetString(t, resp, "data", "name"), "更新後テンプレート", "name")
		assertField(t, mustGetString(t, resp, "data", "default_priority"), "critical", "default_priority")
	})
}

func TestTemplate_Delete(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	projectID := createTestProject(t, ts, "TM", "テストプロジェクト", ownerID)
	tmplID := createTestTemplate(t, ts, projectID, "削除テスト")

	t.Run("テンプレートを削除できる", func(t *testing.T) {
		status, _ := ts.req(t, "DELETE", "/api/v1/templates/"+tmplID, nil)
		assertStatus(t, status, http.StatusNoContent, "delete template")
	})

	t.Run("削除後は取得できない", func(t *testing.T) {
		status, _ := ts.req(t, "GET", "/api/v1/templates/"+tmplID, nil)
		assertStatus(t, status, http.StatusNotFound, "get deleted")
	})
}

func TestTemplate_Reorder(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	projectID := createTestProject(t, ts, "TM", "テンプレート並び替え", ownerID)
	_, r1 := ts.req(t, "POST", "/api/v1/templates", map[string]interface{}{
		"project_id": projectID, "name": "テンプレートA", "description": "A",
	})
	_, r2 := ts.req(t, "POST", "/api/v1/templates", map[string]interface{}{
		"project_id": projectID, "name": "テンプレートB", "description": "B",
	})
	tmplID1 := uint(mustGetFloat(t, r1, "data", "id"))
	tmplID2 := uint(mustGetFloat(t, r2, "data", "id"))

	status, _ := ts.req(t, "PUT", "/api/v1/projects/"+projectID+"/templates/reorder", map[string]interface{}{
		"ids": []uint{tmplID2, tmplID1},
	})
	assertStatus(t, status, http.StatusNoContent, "reorder templates")

	status, listResp := ts.req(t, "GET", "/api/v1/projects/"+projectID+"/templates", nil)
	assertStatus(t, status, http.StatusOK, "list after reorder")
	arr := mustGetArray(t, listResp, "data")
	if len(arr) != 2 {
		t.Fatalf("expected 2 templates, got %d", len(arr))
	}
	assertField(t, mustGetString(t, arr[0].(map[string]interface{}), "name"), "テンプレートB", "first after reorder")
	assertField(t, mustGetString(t, arr[1].(map[string]interface{}), "name"), "テンプレートA", "second after reorder")
}

func TestIssue_CreateFromTemplate(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	projectID := createTestProject(t, ts, "TM", "テストプロジェクト", ownerID)
	statusID := getFirstStatusID(t, ts, projectID)

	_, tmplResp := ts.req(t, "POST", "/api/v1/templates", map[string]interface{}{
		"project_id":       projectID,
		"name":             "バグ報告",
		"body":             "## 概要\n\n## 再現手順",
		"default_priority": "high",
	})
	tmplID := uint(mustGetFloat(t, tmplResp, "data", "id"))

	t.Run("テンプレートIDを指定してIssueを作成できる", func(t *testing.T) {
		status, resp := ts.req(t, "POST", "/api/v1/projects/"+projectID+"/issues", map[string]interface{}{
			"title":       "バグ: ログインできない",
			"status_id":   statusID,
			"reporter_id": ownerID,
			"priority":    "high",
			"template_id": tmplID,
		})
		assertStatus(t, status, http.StatusCreated, "create issue from template")
		data := resp["data"].(map[string]interface{})
		if data["template_id"] == nil {
			t.Errorf("template_id should not be nil")
		}
		if data["workflow_id"] == nil {
			t.Errorf("workflow_id should not be nil (derived from status)")
		}
	})
}
