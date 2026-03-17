package test

import (
	"fmt"
	"net/http"
	"testing"
)

func TestRoleCreate(t *testing.T) {
	ts := newTestServer(t)

	t.Run("役職を作成できる", func(t *testing.T) {
		status, resp := ts.req(t, "POST", "/api/v1/roles", map[string]interface{}{
			"name":        "部長",
			"level":       7,
			"description": "部門の責任者",
		})
		assertStatus(t, status, http.StatusCreated, "create role")
		assertField(t, mustGetString(t, resp, "data", "name"), "部長", "name")
		assertNotEmpty(t, fmt.Sprintf("%v", mustGetFloat(t, resp, "data", "id")), "id")
	})

	t.Run("name未指定は400", func(t *testing.T) {
		status, _ := ts.req(t, "POST", "/api/v1/roles", map[string]interface{}{
			"level": 5,
		})
		assertStatus(t, status, http.StatusBadRequest, "create role without name")
	})
}

func TestRoleList(t *testing.T) {
	ts := newTestServer(t)

	// 複数役職を作成
	ts.req(t, "POST", "/api/v1/roles", map[string]interface{}{"name": "社長", "level": 10})
	ts.req(t, "POST", "/api/v1/roles", map[string]interface{}{"name": "課長", "level": 5})
	ts.req(t, "POST", "/api/v1/roles", map[string]interface{}{"name": "平社員", "level": 1})

	t.Run("役職一覧をlevel降順で取得できる", func(t *testing.T) {
		status, resp := ts.req(t, "GET", "/api/v1/roles", nil)
		assertStatus(t, status, http.StatusOK, "list roles")
		roles := mustGetArray(t, resp, "data")
		if len(roles) != 3 {
			t.Fatalf("expected 3 roles, got %d", len(roles))
		}
		// level降順になっているか確認
		first := roles[0].(map[string]interface{})
		if first["name"] != "社長" {
			t.Errorf("expected first role to be 社長 (highest level), got %v", first["name"])
		}
	})
}

func TestRoleUpdate(t *testing.T) {
	ts := newTestServer(t)

	_, createResp := ts.req(t, "POST", "/api/v1/roles", map[string]interface{}{
		"name":  "課長",
		"level": 5,
	})
	roleID := fmt.Sprintf("%.0f", mustGetFloat(t, createResp, "data", "id"))

	t.Run("役職を更新できる", func(t *testing.T) {
		status, resp := ts.req(t, "PUT", "/api/v1/roles/"+roleID, map[string]interface{}{
			"name":        "上級課長",
			"level":       6,
			"description": "シニア",
		})
		assertStatus(t, status, http.StatusOK, "update role")
		assertField(t, mustGetString(t, resp, "data", "name"), "上級課長", "name")
	})
}

func TestRoleDelete(t *testing.T) {
	ts := newTestServer(t)

	_, createResp := ts.req(t, "POST", "/api/v1/roles", map[string]interface{}{
		"name":  "削除テスト役職",
		"level": 2,
	})
	roleID := fmt.Sprintf("%.0f", mustGetFloat(t, createResp, "data", "id"))

	t.Run("役職を削除できる", func(t *testing.T) {
		status, _ := ts.req(t, "DELETE", "/api/v1/roles/"+roleID, nil)
		assertStatus(t, status, http.StatusNoContent, "delete role")
	})

	t.Run("削除後は一覧から消える", func(t *testing.T) {
		status, resp := ts.req(t, "GET", "/api/v1/roles", nil)
		assertStatus(t, status, http.StatusOK, "list after delete")
		roles := mustGetArray(t, resp, "data")
		if len(roles) != 0 {
			t.Errorf("expected 0 roles after delete, got %d", len(roles))
		}
	})
}

func TestUserRoleAssignment(t *testing.T) {
	ts := newTestServer(t)

	// ユーザーと役職を作成
	userID := createTestUser(t, ts, "田中太郎", "tanaka@example.com")

	_, r1 := ts.req(t, "POST", "/api/v1/roles", map[string]interface{}{"name": "課長", "level": 5})
	_, r2 := ts.req(t, "POST", "/api/v1/roles", map[string]interface{}{"name": "採用担当", "level": 3})
	role1ID := mustGetFloat(t, r1, "data", "id")
	role2ID := mustGetFloat(t, r2, "data", "id")

	t.Run("ユーザーに複数ロールを兼務で割り当てられる", func(t *testing.T) {
		status, resp := ts.req(t, "PUT", "/api/v1/users/"+userID+"/roles", map[string]interface{}{
			"role_ids": []float64{role1ID, role2ID},
		})
		assertStatus(t, status, http.StatusOK, "assign roles")
		assertField(t, mustGetString(t, resp, "message"), "roles assigned", "message")
	})

	t.Run("ユーザーの役職一覧を取得できる", func(t *testing.T) {
		status, resp := ts.req(t, "GET", "/api/v1/users/"+userID+"/roles", nil)
		assertStatus(t, status, http.StatusOK, "get user roles")
		roles := mustGetArray(t, resp, "data")
		if len(roles) != 2 {
			t.Fatalf("expected 2 roles, got %d", len(roles))
		}
	})

	t.Run("役職を空配列で更新すると全解除される", func(t *testing.T) {
		status, _ := ts.req(t, "PUT", "/api/v1/users/"+userID+"/roles", map[string]interface{}{
			"role_ids": []float64{},
		})
		assertStatus(t, status, http.StatusOK, "clear roles")

		_, resp := ts.req(t, "GET", "/api/v1/users/"+userID+"/roles", nil)
		roles := mustGetArray(t, resp, "data")
		if len(roles) != 0 {
			t.Errorf("expected 0 roles after clear, got %d", len(roles))
		}
	})
}

func TestUserIsAdmin(t *testing.T) {
	ts := newTestServer(t)

	t.Run("最初のユーザーは自動的に管理者になる", func(t *testing.T) {
		status, resp := ts.req(t, "POST", "/api/v1/users", map[string]string{
			"name":  "初代管理者",
			"email": "admin@example.com",
		})
		assertStatus(t, status, http.StatusCreated, "create first user")
		isAdmin, ok := resp["data"].(map[string]interface{})["is_admin"].(bool)
		if !ok || !isAdmin {
			t.Errorf("first user should be admin, got is_admin=%v", resp["data"].(map[string]interface{})["is_admin"])
		}
	})

	t.Run("2人目以降は管理者にならない", func(t *testing.T) {
		status, resp := ts.req(t, "POST", "/api/v1/users", map[string]string{
			"name":  "一般ユーザー",
			"email": "member@example.com",
		})
		assertStatus(t, status, http.StatusCreated, "create second user")
		isAdmin, _ := resp["data"].(map[string]interface{})["is_admin"].(bool)
		if isAdmin {
			t.Errorf("second user should not be admin")
		}
	})

	t.Run("管理者フラグを手動で変更できる", func(t *testing.T) {
		// 2人目を管理者に昇格
		_, resp := ts.req(t, "GET", "/api/v1/users", nil)
		users := mustGetArray(t, resp, "data")
		var memberID string
		for _, u := range users {
			user := u.(map[string]interface{})
			if user["email"] == "member@example.com" {
				memberID = user["id"].(string)
			}
		}

		status, _ := ts.req(t, "PUT", "/api/v1/users/"+memberID+"/admin", map[string]interface{}{
			"is_admin": true,
		})
		assertStatus(t, status, http.StatusOK, "set admin")
	})
}

func TestAdminUserList(t *testing.T) {
	ts := newTestServer(t)

	userID := createTestUser(t, ts, "テストユーザー", "test@example.com")
	ts.req(t, "POST", "/api/v1/roles", map[string]interface{}{"name": "エンジニア", "level": 3})
	_, rolesResp := ts.req(t, "GET", "/api/v1/roles", nil)
	roles := mustGetArray(t, rolesResp, "data")
	roleID := roles[0].(map[string]interface{})["id"].(float64)

	ts.req(t, "PUT", "/api/v1/users/"+userID+"/roles", map[string]interface{}{
		"role_ids": []float64{roleID},
	})

	t.Run("管理者用ユーザー一覧はロール情報を含む", func(t *testing.T) {
		status, resp := ts.req(t, "GET", "/api/v1/admin/users", nil)
		assertStatus(t, status, http.StatusOK, "admin user list")
		users := mustGetArray(t, resp, "data")
		if len(users) == 0 {
			t.Fatal("expected at least one user")
		}
		// ロール情報が含まれることを確認
		user := users[0].(map[string]interface{})
		if _, ok := user["roles"]; !ok {
			t.Errorf("admin user list should include roles field")
		}
	})
}
