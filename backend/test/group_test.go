package test

import (
	"fmt"
	"net/http"
	"testing"
)

func TestGroup_Create(t *testing.T) {
	ts := newTestServer(t)
	status, resp := ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/groups", map[string]interface{}{
		"name": "開発部",
	})
	assertStatus(t, status, http.StatusCreated, "create group")
	id := mustGetString(t, resp, "data", "id")
	assertNotEmpty(t, id, "id")
	assertField(t, mustGetString(t, resp, "data", "name"), "開発部", "name")
}

func TestGroup_List(t *testing.T) {
	ts := newTestServer(t)
	status, resp := ts.req(t, "GET", "/api/v1/organizations/"+testOrgID+"/groups", nil)
	assertStatus(t, status, http.StatusOK, "list groups")
	_ = mustGetArray(t, resp, "data")
}

func TestGroup_Update(t *testing.T) {
	ts := newTestServer(t)
	_, createResp := ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/groups", map[string]interface{}{
		"name": "営業部",
	})
	id := mustGetString(t, createResp, "data", "id")

	status, resp := ts.req(t, "PUT", "/api/v1/organizations/"+testOrgID+"/groups/"+id, map[string]interface{}{
		"name": "営業本部",
	})
	assertStatus(t, status, http.StatusOK, "update group")
	assertField(t, mustGetString(t, resp, "data", "name"), "営業本部", "name")
}

func TestGroup_Delete(t *testing.T) {
	ts := newTestServer(t)
	_, createResp := ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/groups", map[string]interface{}{
		"name": "経理部",
	})
	id := mustGetString(t, createResp, "data", "id")

	status, _ := ts.req(t, "DELETE", "/api/v1/organizations/"+testOrgID+"/groups/"+id, nil)
	assertStatus(t, status, http.StatusNoContent, "delete group")
}

// TestGroup_NormalFlow はグループ管理の正常系ブラックボックステスト（一覧→作成→一覧反映→更新→削除）
func TestGroup_NormalFlow(t *testing.T) {
	ts := newTestServer(t)

	// 1. 一覧取得（初期は空）
	status, listResp := ts.req(t, "GET", "/api/v1/organizations/"+testOrgID+"/groups", nil)
	assertStatus(t, status, http.StatusOK, "list groups (initial)")
	arr := mustGetArray(t, listResp, "data")
	if len(arr) != 0 {
		t.Errorf("expected 0 groups initially, got %d", len(arr))
	}

	// 2. 作成
	status, createResp := ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/groups", map[string]interface{}{
		"name": "取締役",
	})
	assertStatus(t, status, http.StatusCreated, "create group")
	id := mustGetString(t, createResp, "data", "id")
	assertNotEmpty(t, id, "id")
	assertField(t, mustGetString(t, createResp, "data", "name"), "取締役", "name")

	// 3. 一覧に反映されていること
	status, listResp = ts.req(t, "GET", "/api/v1/organizations/"+testOrgID+"/groups", nil)
	assertStatus(t, status, http.StatusOK, "list groups (after create)")
	arr = mustGetArray(t, listResp, "data")
	if len(arr) != 1 {
		t.Errorf("expected 1 group after create, got %d", len(arr))
	}
	assertField(t, mustGetString(t, arr[0].(map[string]interface{}), "name"), "取締役", "list[0].name")

	// 4. 更新
	status, updateResp := ts.req(t, "PUT", "/api/v1/organizations/"+testOrgID+"/groups/"+id, map[string]interface{}{
		"name": "取締役会",
	})
	assertStatus(t, status, http.StatusOK, "update group")
	assertField(t, mustGetString(t, updateResp, "data", "name"), "取締役会", "name after update")

	// 5. 一覧で更新が反映されていること
	status, listResp = ts.req(t, "GET", "/api/v1/organizations/"+testOrgID+"/groups", nil)
	assertStatus(t, status, http.StatusOK, "list groups (after update)")
	arr = mustGetArray(t, listResp, "data")
	if len(arr) != 1 {
		t.Errorf("expected 1 group after update, got %d", len(arr))
	}
	assertField(t, mustGetString(t, arr[0].(map[string]interface{}), "name"), "取締役会", "list[0].name after update")

	// 6. 削除
	status, _ = ts.req(t, "DELETE", "/api/v1/organizations/"+testOrgID+"/groups/"+id, nil)
	assertStatus(t, status, http.StatusNoContent, "delete group")

	// 7. 一覧が空に戻ること
	status, listResp = ts.req(t, "GET", "/api/v1/organizations/"+testOrgID+"/groups", nil)
	assertStatus(t, status, http.StatusOK, "list groups (after delete)")
	arr = mustGetArray(t, listResp, "data")
	if len(arr) != 0 {
		t.Errorf("expected 0 groups after delete, got %d", len(arr))
	}
}

func TestGroup_Reorder(t *testing.T) {
	ts := newTestServer(t)
	_, r1 := ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/groups", map[string]interface{}{"name": "開発部"})
	_, r2 := ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/groups", map[string]interface{}{"name": "営業部"})
	id1 := mustGetString(t, r1, "data", "id")
	id2 := mustGetString(t, r2, "data", "id")

	status, _ := ts.req(t, "PUT", "/api/v1/organizations/"+testOrgID+"/groups/reorder", map[string]interface{}{
		"ids": []string{id2, id1},
	})
	assertStatus(t, status, http.StatusNoContent, "reorder groups")

	status, listResp := ts.req(t, "GET", "/api/v1/organizations/"+testOrgID+"/groups", nil)
	assertStatus(t, status, http.StatusOK, "list after reorder")
	arr := mustGetArray(t, listResp, "data")
	if len(arr) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(arr))
	}
	assertField(t, mustGetString(t, arr[0].(map[string]interface{}), "name"), "営業部", "first after reorder")
	assertField(t, mustGetString(t, arr[1].(map[string]interface{}), "name"), "開発部", "second after reorder")
}

func TestGroup_UserGroups(t *testing.T) {
	ts := newTestServer(t)
	userID := createTestUser(t, ts, "User1", "user1@example.com")
	ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/users", map[string]interface{}{
		"user_id":      userID,
		"is_org_admin": false,
	})

	_, groupResp := ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/groups", map[string]interface{}{
		"name": "開発部",
	})
	groupID := mustGetString(t, groupResp, "data", "id")

	status, _ := ts.req(t, "PUT", fmt.Sprintf("/api/v1/users/%s/groups?org_id=%s", userID, testOrgID), map[string]interface{}{
		"group_ids": []string{groupID},
	})
	assertStatus(t, status, http.StatusOK, "set user groups")

	status, getResp := ts.req(t, "GET", fmt.Sprintf("/api/v1/users/%s/groups?org_id=%s", userID, testOrgID), nil)
	assertStatus(t, status, http.StatusOK, "get user groups")
	arr := mustGetArray(t, getResp, "data")
	if len(arr) != 1 {
		t.Errorf("expected 1 group, got %d", len(arr))
	}
}

