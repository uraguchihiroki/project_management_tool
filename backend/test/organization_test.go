package test

import (
	"net/http"
	"testing"
)

func TestOrganization_Create(t *testing.T) {
	ts := newTestServer(t)

	t.Run("組織を作成できる", func(t *testing.T) {
		status, resp := ts.req(t, "POST", "/api/v1/organizations", map[string]interface{}{
			"name": "テスト株式会社",
		})
		assertStatus(t, status, http.StatusCreated, "create org")
		assertField(t, mustGetString(t, resp, "data", "name"), "テスト株式会社", "name")
		assertNotEmpty(t, mustGetString(t, resp, "data", "id"), "id")
	})

	t.Run("nameが空の場合は400", func(t *testing.T) {
		status, _ := ts.req(t, "POST", "/api/v1/organizations", map[string]interface{}{
			"name": "",
		})
		assertStatus(t, status, http.StatusBadRequest, "create org without name")
	})
}

func TestOrganization_List(t *testing.T) {
	ts := newTestServer(t)

	ts.req(t, "POST", "/api/v1/organizations", map[string]interface{}{"name": "A社"})
	ts.req(t, "POST", "/api/v1/organizations", map[string]interface{}{"name": "B社"})

	t.Run("組織一覧を取得できる（FRS含む）", func(t *testing.T) {
		status, resp := ts.req(t, "GET", "/api/v1/organizations", nil)
		assertStatus(t, status, http.StatusOK, "list orgs")
		orgs := mustGetArray(t, resp, "data")
		// FRS(seed) + A社 + B社 = 3件
		if len(orgs) != 3 {
			t.Fatalf("expected 3 orgs, got %d", len(orgs))
		}
	})
}

func TestOrganization_UserMembership(t *testing.T) {
	ts := newTestServer(t)

	userID := createTestUser(t, ts, "山田太郎", "yamada@example.com")
	_, org1Resp := ts.req(t, "POST", "/api/v1/organizations", map[string]interface{}{"name": "所属会社1"})
	_, org2Resp := ts.req(t, "POST", "/api/v1/organizations", map[string]interface{}{"name": "所属会社2"})
	org1ID := mustGetString(t, org1Resp, "data", "id")
	org2ID := mustGetString(t, org2Resp, "data", "id")

	t.Run("ユーザーを組織に追加できる", func(t *testing.T) {
		status, _ := ts.req(t, "POST", "/api/v1/organizations/"+org1ID+"/users", map[string]interface{}{
			"user_id": userID,
		})
		assertStatus(t, status, http.StatusCreated, "add user to org1")
	})

	t.Run("ユーザーを複数組織に所属させられる", func(t *testing.T) {
		status, _ := ts.req(t, "POST", "/api/v1/organizations/"+org2ID+"/users", map[string]interface{}{
			"user_id": userID,
		})
		assertStatus(t, status, http.StatusCreated, "add user to org2")
	})

	t.Run("ユーザーの所属組織一覧を取得できる", func(t *testing.T) {
		status, resp := ts.req(t, "GET", "/api/v1/users/"+userID+"/organizations", nil)
		assertStatus(t, status, http.StatusOK, "get user orgs")
		orgs := mustGetArray(t, resp, "data")
		if len(orgs) != 2 {
			t.Fatalf("expected 2 orgs, got %d", len(orgs))
		}
	})

	t.Run("重複追加はエラーにならず冪等に処理される", func(t *testing.T) {
		status, _ := ts.req(t, "POST", "/api/v1/organizations/"+org1ID+"/users", map[string]interface{}{
			"user_id": userID,
		})
		assertStatus(t, status, http.StatusCreated, "add user to org1 again")
		// 所属数は変わらない
		_, resp := ts.req(t, "GET", "/api/v1/users/"+userID+"/organizations", nil)
		orgs := mustGetArray(t, resp, "data")
		if len(orgs) != 2 {
			t.Errorf("expected 2 orgs (no duplicate), got %d", len(orgs))
		}
	})
}

func TestProject_FilterByOrg(t *testing.T) {
	ts := newTestServer(t)

	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	_, org1Resp := ts.req(t, "POST", "/api/v1/organizations", map[string]interface{}{"name": "フィルタ会社1"})
	_, org2Resp := ts.req(t, "POST", "/api/v1/organizations", map[string]interface{}{"name": "フィルタ会社2"})
	org1ID := mustGetString(t, org1Resp, "data", "id")
	org2ID := mustGetString(t, org2Resp, "data", "id")

	// org1 に2件、org2 に1件プロジェクトを作成
	ts.req(t, "POST", "/api/v1/projects", map[string]interface{}{
		"key": "A01", "name": "org1プロジェクト1", "owner_id": ownerID, "organization_id": org1ID,
	})
	ts.req(t, "POST", "/api/v1/projects", map[string]interface{}{
		"key": "A02", "name": "org1プロジェクト2", "owner_id": ownerID, "organization_id": org1ID,
	})
	ts.req(t, "POST", "/api/v1/projects", map[string]interface{}{
		"key": "B01", "name": "org2プロジェクト1", "owner_id": ownerID, "organization_id": org2ID,
	})

	t.Run("org_idでフィルタするとその組織のプロジェクトだけ取得できる", func(t *testing.T) {
		status, resp := ts.req(t, "GET", "/api/v1/projects?org_id="+org1ID, nil)
		assertStatus(t, status, http.StatusOK, "filter by org1")
		projects := mustGetArray(t, resp, "data")
		if len(projects) != 2 {
			t.Fatalf("expected 2 projects for org1, got %d", len(projects))
		}
	})

	t.Run("org_id未指定は全件取得", func(t *testing.T) {
		status, resp := ts.req(t, "GET", "/api/v1/projects", nil)
		assertStatus(t, status, http.StatusOK, "all projects")
		projects := mustGetArray(t, resp, "data")
		if len(projects) != 3 {
			t.Fatalf("expected 3 projects total, got %d", len(projects))
		}
	})
}

func TestSuperAdmin_Login(t *testing.T) {
	ts := newTestServer(t)

	// スーパーアドミンを作成
	ts.req(t, "POST", "/api/v1/super-admin/organizations", map[string]interface{}{
		"name": "テスト組織",
	})

	// スーパーアドミンを直接DBに挿入
	ts.db.Exec("INSERT INTO super_admins (id, name, email, created_at) VALUES (?, ?, ?, datetime('now'))",
		"00000000-0000-0000-0000-000000000099", "SA管理者", "sa@example.com")

	t.Run("登録済みメールでログインできる", func(t *testing.T) {
		status, resp := ts.req(t, "POST", "/api/v1/super-admin/login", map[string]interface{}{
			"email": "sa@example.com",
		})
		assertStatus(t, status, http.StatusOK, "super admin login")
		assertField(t, mustGetString(t, resp, "data", "email"), "sa@example.com", "email")
	})

	t.Run("未登録メールは401", func(t *testing.T) {
		status, _ := ts.req(t, "POST", "/api/v1/super-admin/login", map[string]interface{}{
			"email": "unknown@example.com",
		})
		assertStatus(t, status, http.StatusUnauthorized, "unknown email")
	})

	t.Run("スーパーアドミンから組織を作成できる", func(t *testing.T) {
		status, resp := ts.req(t, "POST", "/api/v1/super-admin/organizations", map[string]interface{}{
			"name": "スーパー経由で作成",
		})
		assertStatus(t, status, http.StatusCreated, "create org via super admin")
		assertField(t, mustGetString(t, resp, "data", "name"), "スーパー経由で作成", "name")
	})
}
