package test

import (
	"net/http"
	"testing"
)

func TestProject_Create(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")

	t.Run("正常系: プロジェクト作成", func(t *testing.T) {
		status, resp := ts.req(t, "POST", "/api/v1/projects", map[string]string{
			"key":      "PROJ",
			"name":     "テストプロジェクト",
			"owner_id": ownerID,
		})
		assertStatus(t, status, http.StatusCreated, "POST /projects")
		assertNotEmpty(t, mustGetString(t, resp, "data", "id"), "id")
		assertField(t, mustGetString(t, resp, "data", "key"), "PROJ", "key")
		assertField(t, mustGetString(t, resp, "data", "name"), "テストプロジェクト", "name")
	})

	t.Run("正常系: 作成時にデフォルトステータスが自動生成される", func(t *testing.T) {
		status, resp := ts.req(t, "POST", "/api/v1/projects", map[string]string{
			"key": "AUTO", "name": "ステータス確認", "owner_id": ownerID,
		})
		assertStatus(t, status, http.StatusCreated, "POST /projects (statuses)")
		data := resp["data"].(map[string]interface{})
		statuses, ok := data["statuses"].([]interface{})
		if !ok || len(statuses) == 0 {
			t.Error("project creation should generate default statuses")
		}
	})

	t.Run("異常系: 同じキーのプロジェクトは重複不可", func(t *testing.T) {
		ts.req(t, "POST", "/api/v1/projects", map[string]string{
			"key": "DUP", "name": "最初", "owner_id": ownerID,
		})
		status, _ := ts.req(t, "POST", "/api/v1/projects", map[string]string{
			"key": "DUP", "name": "重複", "owner_id": ownerID,
		})
		if status == http.StatusCreated {
			t.Error("重複キーのプロジェクトが作成できてしまった")
		}
	})
}

func TestProject_List(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner2@example.com")

	t.Run("正常系: 空のリスト", func(t *testing.T) {
		status, resp := ts.req(t, "GET", "/api/v1/projects", nil)
		assertStatus(t, status, http.StatusOK, "GET /projects (empty)")
		arr := mustGetArray(t, resp, "data")
		if len(arr) != 0 {
			t.Errorf("expected 0 projects, got %d", len(arr))
		}
	})

	t.Run("正常系: 作成後に一覧に含まれる", func(t *testing.T) {
		createTestProject(t, ts, "PA", "プロジェクトA", ownerID)
		createTestProject(t, ts, "PB", "プロジェクトB", ownerID)
		status, resp := ts.req(t, "GET", "/api/v1/projects", nil)
		assertStatus(t, status, http.StatusOK, "GET /projects")
		arr := mustGetArray(t, resp, "data")
		if len(arr) != 2 {
			t.Errorf("expected 2 projects, got %d", len(arr))
		}
	})
}

func TestProject_Get(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner3@example.com")

	t.Run("正常系: IDで取得", func(t *testing.T) {
		projectID := createTestProject(t, ts, "GET1", "取得テスト", ownerID)
		status, resp := ts.req(t, "GET", "/api/v1/projects/"+projectID, nil)
		assertStatus(t, status, http.StatusOK, "GET /projects/:id")
		assertField(t, mustGetString(t, resp, "data", "id"), projectID, "id")
	})

	t.Run("異常系: 存在しないIDは404", func(t *testing.T) {
		status, _ := ts.req(t, "GET", "/api/v1/projects/00000000-0000-0000-0000-000000000000", nil)
		assertStatus(t, status, http.StatusNotFound, "GET /projects/:id (not found)")
	})
}

func TestProject_Update(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner4@example.com")

	t.Run("正常系: プロジェクト名を更新", func(t *testing.T) {
		projectID := createTestProject(t, ts, "UPD", "変更前", ownerID)
		status, resp := ts.req(t, "PUT", "/api/v1/projects/"+projectID, map[string]string{
			"name": "変更後",
		})
		assertStatus(t, status, http.StatusOK, "PUT /projects/:id")
		assertField(t, mustGetString(t, resp, "data", "name"), "変更後", "name")
	})

	t.Run("正常系: 更新後にGETで反映されている", func(t *testing.T) {
		projectID := createTestProject(t, ts, "UP2", "更新確認前", ownerID)
		ts.req(t, "PUT", "/api/v1/projects/"+projectID, map[string]string{"name": "更新確認後"})
		status, resp := ts.req(t, "GET", "/api/v1/projects/"+projectID, nil)
		assertStatus(t, status, http.StatusOK, "GET after PUT")
		assertField(t, mustGetString(t, resp, "data", "name"), "更新確認後", "name")
	})
}

func TestProject_Delete(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner5@example.com")

	t.Run("正常系: プロジェクト削除", func(t *testing.T) {
		projectID := createTestProject(t, ts, "DEL", "削除対象", ownerID)
		status, _ := ts.req(t, "DELETE", "/api/v1/projects/"+projectID, nil)
		assertStatus(t, status, http.StatusOK, "DELETE /projects/:id")

		getStatus, _ := ts.req(t, "GET", "/api/v1/projects/"+projectID, nil)
		assertStatus(t, getStatus, http.StatusNotFound, "GET after DELETE")
	})
}
