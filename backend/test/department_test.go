package test

import (
	"fmt"
	"net/http"
	"testing"
)

func TestDepartment_Create(t *testing.T) {
	ts := newTestServer(t)
	status, resp := ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/departments", map[string]interface{}{
		"name":  "開発部",
		"order": 1,
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
		"name":  "営業部",
		"order": 0,
	})
	id := mustGetString(t, createResp, "data", "id")

	status, resp := ts.req(t, "PUT", "/api/v1/organizations/"+testOrgID+"/departments/"+id, map[string]interface{}{
		"name":  "営業本部",
		"order": 2,
	})
	assertStatus(t, status, http.StatusOK, "update department")
	assertField(t, mustGetString(t, resp, "data", "name"), "営業本部", "name")
}

func TestDepartment_Delete(t *testing.T) {
	ts := newTestServer(t)
	_, createResp := ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/departments", map[string]interface{}{
		"name":  "経理部",
		"order": 0,
	})
	id := mustGetString(t, createResp, "data", "id")

	status, _ := ts.req(t, "DELETE", "/api/v1/organizations/"+testOrgID+"/departments/"+id, nil)
	assertStatus(t, status, http.StatusNoContent, "delete department")
}

func TestDepartment_UserDepartments(t *testing.T) {
	ts := newTestServer(t)
	userID := createTestUser(t, ts, "User1", "user1@example.com")
	ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/users", map[string]interface{}{
		"user_id":      userID,
		"is_org_admin": false,
	})

	_, deptResp := ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/departments", map[string]interface{}{
		"name":  "開発部",
		"order": 0,
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
