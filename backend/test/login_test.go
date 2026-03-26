package test

import (
	"net/http"
	"testing"

	"github.com/uraguchihiroki/project_management_tool/internal/auth"
)

// 一般ユーザーログイン（POST /admin/login）と、そのJWTでの後続APIをブラックボックスで検証する。
func TestAdminLogin(t *testing.T) {
	ts := newTestServer(t)
	email := "login-happy-path@example.com"
	userID := createTestUser(t, ts, "ログイン検証ユーザー", email)

	t.Run("正常系: メールでログインし user と token が返る", func(t *testing.T) {
		status, resp := ts.reqNoAuth(t, "POST", "/api/v1/admin/login", map[string]string{"email": email})
		assertStatus(t, status, http.StatusOK, "POST /admin/login")
		token := mustGetString(t, resp, "data", "token")
		assertNotEmpty(t, token, "data.token")
		gotEmail := mustGetString(t, resp, "data", "user", "email")
		assertField(t, gotEmail, email, "data.user.email")
		gotID := mustGetString(t, resp, "data", "user", "id")
		assertField(t, gotID, userID, "data.user.id")

		claims, err := auth.ParseToken(token)
		if err != nil {
			t.Fatalf("parse token: %v", err)
		}
		if claims.UserID != userID {
			t.Errorf("JWT user_id = %q, want %q", claims.UserID, userID)
		}
	})

	t.Run("正常系: 発行JWTで GET /users/:id/organizations が取得できる", func(t *testing.T) {
		_, loginResp := ts.reqNoAuth(t, "POST", "/api/v1/admin/login", map[string]string{"email": email})
		token := mustGetString(t, loginResp, "data", "token")

		status, orgResp := ts.reqWithToken(t, token, "GET", "/api/v1/users/"+userID+"/organizations", nil)
		assertStatus(t, status, http.StatusOK, "GET /users/:id/organizations (with user JWT)")
		arr := mustGetArray(t, orgResp, "data")
		if len(arr) != 1 {
			t.Fatalf("expected 1 organization for user, got %d", len(arr))
		}
		orgID := arr[0].(map[string]interface{})["id"].(string)
		if orgID != testOrgID {
			t.Errorf("organization id = %q, want %q", orgID, testOrgID)
		}
	})

	t.Run("異常系: 未登録メールは401", func(t *testing.T) {
		status, _ := ts.reqNoAuth(t, "POST", "/api/v1/admin/login", map[string]string{
			"email": "no-such-user@example.com",
		})
		assertStatus(t, status, http.StatusUnauthorized, "POST /admin/login (unknown email)")
	})
}
