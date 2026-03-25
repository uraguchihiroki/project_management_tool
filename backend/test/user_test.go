package test

import (
	"fmt"
	"net/http"
	"testing"
)

func TestUser_Create(t *testing.T) {
	ts := newTestServer(t)

	t.Run("正常系: ユーザー作成", func(t *testing.T) {
		status, resp := ts.req(t, "POST", "/api/v1/users", map[string]string{
			"name":  "テストユーザー",
			"email": "test@example.com",
		})
		assertStatus(t, status, http.StatusCreated, "POST /users")
		assertNotEmpty(t, mustGetString(t, resp, "data", "id"), "id")
		assertField(t, mustGetString(t, resp, "data", "name"), "テストユーザー", "name")
		assertField(t, mustGetString(t, resp, "data", "email"), "test@example.com", "email")
	})

	t.Run("異常系: 同じメールで2回登録するとエラー", func(t *testing.T) {
		ts.req(t, "POST", "/api/v1/users", map[string]string{
			"name": "ユーザーA", "email": "dup@example.com",
		})
		status, _ := ts.req(t, "POST", "/api/v1/users", map[string]string{
			"name": "ユーザーB", "email": "dup@example.com",
		})
		if status == http.StatusCreated {
			t.Error("重複メールアドレスで作成できてしまった")
		}
	})
}

func TestUser_List(t *testing.T) {
	ts := newTestServer(t)

	t.Run("正常系: 空のリスト", func(t *testing.T) {
		status, resp := ts.req(t, "GET", "/api/v1/users", nil)
		assertStatus(t, status, http.StatusOK, "GET /users (empty)")
		arr := mustGetArray(t, resp, "data")
		if len(arr) != 0 {
			t.Errorf("empty DB should return 0 users, got %d", len(arr))
		}
	})

	t.Run("正常系: 作成後に一覧に含まれる", func(t *testing.T) {
		createTestUser(t, ts, "ユーザー1", "u1@example.com")
		createTestUser(t, ts, "ユーザー2", "u2@example.com")
		status, resp := ts.req(t, "GET", "/api/v1/users", nil)
		assertStatus(t, status, http.StatusOK, "GET /users")
		arr := mustGetArray(t, resp, "data")
		if len(arr) != 2 {
			t.Errorf("expected 2 users, got %d", len(arr))
		}
	})
}

func TestUser_Get(t *testing.T) {
	ts := newTestServer(t)

	t.Run("正常系: IDで取得", func(t *testing.T) {
		id := createTestUser(t, ts, "取得テスト", "get@example.com")
		status, resp := ts.req(t, "GET", "/api/v1/users/"+id, nil)
		assertStatus(t, status, http.StatusOK, "GET /users/:id")
		assertField(t, mustGetString(t, resp, "data", "id"), id, "id")
		assertField(t, mustGetString(t, resp, "data", "email"), "get@example.com", "email")
	})

	t.Run("異常系: 存在しないIDは404", func(t *testing.T) {
		status, _ := ts.req(t, "GET", fmt.Sprintf("/api/v1/users/%s", "00000000-0000-0000-0000-000000000000"), nil)
		assertStatus(t, status, http.StatusNotFound, "GET /users/:id (not found)")
	})

	t.Run("異常系: 不正なIDは400", func(t *testing.T) {
		status, _ := ts.req(t, "GET", "/api/v1/users/invalid-id", nil)
		assertStatus(t, status, http.StatusBadRequest, "GET /users/invalid-id")
	})
}
