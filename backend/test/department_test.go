package test

import (
	"fmt"
	"net/http"
	"testing"
)

func TestDepartment_Create(t *testing.T) {
	ts := newTestServer(t)
	status, resp := ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/departments", map[string]interface{}{
		"name": "開発部",
	})
	assertStatus(t, status, http.StatusCreated, "create department")
	id := mustGetString(t, resp, "data", "id")
	assertNotEmpty(t, id, "id")
	assertField(t, mustGetString(t, resp, "data", "name"), "開発部", "name")
}

func TestDepartment_List(t *testing.T) {
	ts := newTestServer(t)
	status, resp := ts.req(t, "GET", "/api/v1/organizations/"+testOrgID+"/departments", nil)
	assertStatus(t, status, http.StatusOK, "list departments")
	_ = mustGetArray(t, resp, "data")
}

func TestDepartment_Update(t *testing.T) {
	ts := newTestServer(t)
	_, createResp := ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/departments", map[string]interface{}{
		"name": "営業部",
	})
	id := mustGetString(t, createResp, "data", "id")

	status, resp := ts.req(t, "PUT", "/api/v1/organizations/"+testOrgID+"/departments/"+id, map[string]interface{}{
		"name": "営業本部",
	})
	assertStatus(t, status, http.StatusOK, "update department")
	assertField(t, mustGetString(t, resp, "data", "name"), "営業本部", "name")
}

func TestDepartment_Delete(t *testing.T) {
	ts := newTestServer(t)
	_, createResp := ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/departments", map[string]interface{}{
		"name": "経理部",
	})
	id := mustGetString(t, createResp, "data", "id")

	status, _ := ts.req(t, "DELETE", "/api/v1/organizations/"+testOrgID+"/departments/"+id, nil)
	assertStatus(t, status, http.StatusNoContent, "delete department")
}

// TestDepartment_NormalFlow は部署管理の正常系ブラックボックステスト（一覧→作成→一覧反映→更新→削除）
func TestDepartment_NormalFlow(t *testing.T) {
	ts := newTestServer(t)

	// 1. 一覧取得（初期は空）
	status, listResp := ts.req(t, "GET", "/api/v1/organizations/"+testOrgID+"/departments", nil)
	assertStatus(t, status, http.StatusOK, "list departments (initial)")
	arr := mustGetArray(t, listResp, "data")
	if len(arr) != 0 {
		t.Errorf("expected 0 departments initially, got %d", len(arr))
	}

	// 2. 作成
	status, createResp := ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/departments", map[string]interface{}{
		"name": "取締役",
	})
	assertStatus(t, status, http.StatusCreated, "create department")
	id := mustGetString(t, createResp, "data", "id")
	assertNotEmpty(t, id, "id")
	assertField(t, mustGetString(t, createResp, "data", "name"), "取締役", "name")

	// 3. 一覧に反映されていること
	status, listResp = ts.req(t, "GET", "/api/v1/organizations/"+testOrgID+"/departments", nil)
	assertStatus(t, status, http.StatusOK, "list departments (after create)")
	arr = mustGetArray(t, listResp, "data")
	if len(arr) != 1 {
		t.Errorf("expected 1 department after create, got %d", len(arr))
	}
	assertField(t, mustGetString(t, arr[0].(map[string]interface{}), "name"), "取締役", "list[0].name")

	// 4. 更新
	status, updateResp := ts.req(t, "PUT", "/api/v1/organizations/"+testOrgID+"/departments/"+id, map[string]interface{}{
		"name": "取締役会",
	})
	assertStatus(t, status, http.StatusOK, "update department")
	assertField(t, mustGetString(t, updateResp, "data", "name"), "取締役会", "name after update")

	// 5. 一覧で更新が反映されていること
	status, listResp = ts.req(t, "GET", "/api/v1/organizations/"+testOrgID+"/departments", nil)
	assertStatus(t, status, http.StatusOK, "list departments (after update)")
	arr = mustGetArray(t, listResp, "data")
	if len(arr) != 1 {
		t.Errorf("expected 1 department after update, got %d", len(arr))
	}
	assertField(t, mustGetString(t, arr[0].(map[string]interface{}), "name"), "取締役会", "list[0].name after update")

	// 6. 削除
	status, _ = ts.req(t, "DELETE", "/api/v1/organizations/"+testOrgID+"/departments/"+id, nil)
	assertStatus(t, status, http.StatusNoContent, "delete department")

	// 7. 一覧が空に戻ること
	status, listResp = ts.req(t, "GET", "/api/v1/organizations/"+testOrgID+"/departments", nil)
	assertStatus(t, status, http.StatusOK, "list departments (after delete)")
	arr = mustGetArray(t, listResp, "data")
	if len(arr) != 0 {
		t.Errorf("expected 0 departments after delete, got %d", len(arr))
	}
}

func TestDepartment_Reorder(t *testing.T) {
	ts := newTestServer(t)
	_, r1 := ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/departments", map[string]interface{}{"name": "開発部"})
	_, r2 := ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/departments", map[string]interface{}{"name": "営業部"})
	id1 := mustGetString(t, r1, "data", "id")
	id2 := mustGetString(t, r2, "data", "id")

	status, _ := ts.req(t, "PUT", "/api/v1/organizations/"+testOrgID+"/departments/reorder", map[string]interface{}{
		"ids": []string{id2, id1},
	})
	assertStatus(t, status, http.StatusNoContent, "reorder departments")

	status, listResp := ts.req(t, "GET", "/api/v1/organizations/"+testOrgID+"/departments", nil)
	assertStatus(t, status, http.StatusOK, "list after reorder")
	arr := mustGetArray(t, listResp, "data")
	if len(arr) != 2 {
		t.Fatalf("expected 2 departments, got %d", len(arr))
	}
	assertField(t, mustGetString(t, arr[0].(map[string]interface{}), "name"), "営業部", "first after reorder")
	assertField(t, mustGetString(t, arr[1].(map[string]interface{}), "name"), "開発部", "second after reorder")
}

func TestDepartment_UserDepartments(t *testing.T) {
	ts := newTestServer(t)
	userID := createTestUser(t, ts, "User1", "user1@example.com")
	ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/users", map[string]interface{}{
		"user_id":      userID,
		"is_org_admin": false,
	})

	_, deptResp := ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/departments", map[string]interface{}{
		"name": "開発部",
	})
	deptID := mustGetString(t, deptResp, "data", "id")

	status, _ := ts.req(t, "PUT", fmt.Sprintf("/api/v1/users/%s/departments?org_id=%s", userID, testOrgID), map[string]interface{}{
		"department_ids": []string{deptID},
	})
	assertStatus(t, status, http.StatusOK, "set user departments")

	status, getResp := ts.req(t, "GET", fmt.Sprintf("/api/v1/users/%s/departments?org_id=%s", userID, testOrgID), nil)
	assertStatus(t, status, http.StatusOK, "get user departments")
	arr := mustGetArray(t, getResp, "data")
	if len(arr) != 1 {
		t.Errorf("expected 1 department, got %d", len(arr))
	}
}
