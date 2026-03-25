package test

import (
	"net/http"
	"testing"
)

// TestProject_NormalFlow はプロジェクト管理の正常系ブラックボックステスト（一覧→作成→一覧反映→取得→更新→取得で反映確認→削除）
func TestProject_NormalFlow(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "normalflow@example.com")

	// 1. 一覧取得（初期は空）
	status, listResp := ts.req(t, "GET", "/api/v1/projects", nil)
	assertStatus(t, status, http.StatusOK, "list projects (initial)")
	arr := mustGetArray(t, listResp, "data")
	if len(arr) != 0 {
		t.Errorf("expected 0 projects initially, got %d", len(arr))
	}

	// 2. 作成
	status, createResp := ts.req(t, "POST", "/api/v1/projects", map[string]interface{}{
		"key":             "FIRST",
		"name":            "初めてのプロジェクト",
		"owner_id":        ownerID,
		"organization_id": testOrgID,
	})
	assertStatus(t, status, http.StatusCreated, "create project")
	projectID := mustGetString(t, createResp, "data", "id")
	assertNotEmpty(t, projectID, "id")
	assertField(t, mustGetString(t, createResp, "data", "name"), "初めてのプロジェクト", "name")

	// 3. 一覧に反映されていること
	status, listResp = ts.req(t, "GET", "/api/v1/projects", nil)
	assertStatus(t, status, http.StatusOK, "list projects (after create)")
	arr = mustGetArray(t, listResp, "data")
	if len(arr) != 1 {
		t.Errorf("expected 1 project after create, got %d", len(arr))
	}
	assertField(t, mustGetString(t, arr[0].(map[string]interface{}), "name"), "初めてのプロジェクト", "list[0].name")

	// 4. 取得
	status, getResp := ts.req(t, "GET", "/api/v1/projects/"+projectID, nil)
	assertStatus(t, status, http.StatusOK, "get project")
	assertField(t, mustGetString(t, getResp, "data", "name"), "初めてのプロジェクト", "name")

	// 5. 更新
	status, updateResp := ts.req(t, "PUT", "/api/v1/projects/"+projectID, map[string]interface{}{
		"name":        "初めてのプロジェクト（更新後）",
		"description": "説明を追加",
	})
	assertStatus(t, status, http.StatusOK, "update project")
	assertField(t, mustGetString(t, updateResp, "data", "name"), "初めてのプロジェクト（更新後）", "name after update")

	// 6. 取得で更新が反映されていること
	status, getResp = ts.req(t, "GET", "/api/v1/projects/"+projectID, nil)
	assertStatus(t, status, http.StatusOK, "get project (after update)")
	assertField(t, mustGetString(t, getResp, "data", "name"), "初めてのプロジェクト（更新後）", "name after update")
	assertField(t, mustGetString(t, getResp, "data", "description"), "説明を追加", "description after update")

	// 7. 削除
	status, _ = ts.req(t, "DELETE", "/api/v1/projects/"+projectID, nil)
	assertStatus(t, status, http.StatusOK, "delete project")

	// 8. 取得で404になること
	status, _ = ts.req(t, "GET", "/api/v1/projects/"+projectID, nil)
	assertStatus(t, status, http.StatusNotFound, "get after delete")
}

func TestProject_Create(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")

	t.Run("正常系: プロジェクト作成", func(t *testing.T) {
		status, resp := ts.req(t, "POST", "/api/v1/projects", map[string]string{
			"key":             "PROJ",
			"name":            "テストプロジェクト",
			"owner_id":        ownerID,
			"organization_id": testOrgID,
		})
		assertStatus(t, status, http.StatusCreated, "POST /projects")
		assertNotEmpty(t, mustGetString(t, resp, "data", "id"), "id")
		assertField(t, mustGetString(t, resp, "data", "key"), "PROJ", "key")
		assertField(t, mustGetString(t, resp, "data", "name"), "テストプロジェクト", "name")
	})

	t.Run("正常系: 作成時にデフォルトステータスが自動生成される", func(t *testing.T) {
		status, resp := ts.req(t, "POST", "/api/v1/projects", map[string]string{
			"key": "AUTO", "name": "ステータス確認", "owner_id": ownerID, "organization_id": testOrgID,
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
			"key": "DUP", "name": "最初", "owner_id": ownerID, "organization_id": testOrgID,
		})
		status, _ := ts.req(t, "POST", "/api/v1/projects", map[string]string{
			"key": "DUP", "name": "重複", "owner_id": ownerID, "organization_id": testOrgID,
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

func TestProject_Reorder(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "reorder@example.com")
	_, r1 := ts.req(t, "POST", "/api/v1/projects", map[string]interface{}{
		"key": "PA", "name": "プロジェクトA", "owner_id": ownerID, "organization_id": testOrgID,
	})
	_, r2 := ts.req(t, "POST", "/api/v1/projects", map[string]interface{}{
		"key": "PB", "name": "プロジェクトB", "owner_id": ownerID, "organization_id": testOrgID,
	})
	id1 := mustGetString(t, r1, "data", "id")
	id2 := mustGetString(t, r2, "data", "id")

	status, _ := ts.req(t, "PUT", "/api/v1/projects/reorder?org_id="+testOrgID, map[string]interface{}{
		"ids": []string{id2, id1},
	})
	assertStatus(t, status, http.StatusNoContent, "reorder projects")

	status, listResp := ts.req(t, "GET", "/api/v1/projects?org_id="+testOrgID, nil)
	assertStatus(t, status, http.StatusOK, "list after reorder")
	arr := mustGetArray(t, listResp, "data")
	if len(arr) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(arr))
	}
	assertField(t, mustGetString(t, arr[0].(map[string]interface{}), "name"), "プロジェクトB", "first after reorder")
	assertField(t, mustGetString(t, arr[1].(map[string]interface{}), "name"), "プロジェクトA", "second after reorder")
}
