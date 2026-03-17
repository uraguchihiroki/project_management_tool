// Package test はAPIエンドポイントのブラックボックステストを提供します。
// テスト用DBにSQLiteを使用するため、PostgreSQLの起動は不要です。
// 実行: go test ./test/... -v
package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/uraguchihiroki/project_management_tool/internal/handler"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
	"github.com/uraguchihiroki/project_management_tool/internal/service"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// testServer はテスト用のHTTPサーバーとDBを保持します
type testServer struct {
	server *httptest.Server
	db     *gorm.DB
}

// newTestServer はSQLiteを使ったテスト用サーバーを起動します
func newTestServer(t *testing.T) *testServer {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Project{},
		&model.Status{},
		&model.Issue{},
		&model.Comment{},
	); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	statusRepo := repository.NewStatusRepository(db)
	issueRepo := repository.NewIssueRepository(db)
	commentRepo := repository.NewCommentRepository(db)

	userSvc := service.NewUserService(userRepo)
	projectSvc := service.NewProjectService(projectRepo, statusRepo)
	issueSvc := service.NewIssueService(issueRepo, projectRepo)
	commentSvc := service.NewCommentService(commentRepo)

	userH := handler.NewUserHandler(userSvc)
	projectH := handler.NewProjectHandler(projectSvc)
	issueH := handler.NewIssueHandler(issueSvc)
	commentH := handler.NewCommentHandler(commentSvc)

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())

	api := e.Group("/api/v1")
	api.GET("/users", userH.List)
	api.POST("/users", userH.Create)
	api.GET("/users/:id", userH.Get)
	api.GET("/projects", projectH.List)
	api.POST("/projects", projectH.Create)
	api.GET("/projects/:id", projectH.Get)
	api.PUT("/projects/:id", projectH.Update)
	api.DELETE("/projects/:id", projectH.Delete)
	api.GET("/projects/:projectId/issues", issueH.List)
	api.POST("/projects/:projectId/issues", issueH.Create)
	api.GET("/projects/:projectId/issues/:number", issueH.Get)
	api.PUT("/projects/:projectId/issues/:number", issueH.Update)
	api.DELETE("/projects/:projectId/issues/:number", issueH.Delete)
	api.GET("/issues/:issueId/comments", commentH.List)
	api.POST("/issues/:issueId/comments", commentH.Create)
	api.PUT("/issues/:issueId/comments/:id", commentH.Update)
	api.DELETE("/issues/:issueId/comments/:id", commentH.Delete)

	srv := httptest.NewServer(e)
	t.Cleanup(func() {
		srv.Close()
	})

	return &testServer{server: srv, db: db}
}

// req はテスト用HTTPリクエストを送信し、レスポンスのbodyを返します
func (ts *testServer) req(t *testing.T, method, path string, body interface{}) (int, map[string]interface{}) {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}

	req, _ := http.NewRequest(method, ts.server.URL+path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return resp.StatusCode, result
}

// mustGetString はレスポンスのネストしたフィールドを取得します
func mustGetString(t *testing.T, m map[string]interface{}, keys ...string) string {
	t.Helper()
	var current interface{} = m
	for _, k := range keys {
		mp, ok := current.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map at key chain %v, got %T", keys, current)
		}
		current = mp[k]
	}
	s, ok := current.(string)
	if !ok {
		t.Fatalf("expected string at %v, got %T: %v", keys, current, current)
	}
	return s
}

// mustGetFloat はレスポンスの数値フィールドを取得します
func mustGetFloat(t *testing.T, m map[string]interface{}, keys ...string) float64 {
	t.Helper()
	var current interface{} = m
	for _, k := range keys {
		mp, ok := current.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map at key chain %v", keys)
		}
		current = mp[k]
	}
	f, ok := current.(float64)
	if !ok {
		t.Fatalf("expected float64 at %v, got %T: %v", keys, current, current)
	}
	return f
}

// mustGetArray はレスポンスの配列フィールドを取得します
func mustGetArray(t *testing.T, m map[string]interface{}, key string) []interface{} {
	t.Helper()
	arr, ok := m[key].([]interface{})
	if !ok {
		t.Fatalf("expected array at key %q, got %T: %v", key, m[key], m[key])
	}
	return arr
}

func assertStatus(t *testing.T, got, want int, context string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: status = %d, want %d", context, got, want)
	}
}

func assertField(t *testing.T, got, want, field string) {
	t.Helper()
	if got != want {
		t.Errorf("field %q = %q, want %q", field, got, want)
	}
}

func assertNotEmpty(t *testing.T, val, field string) {
	t.Helper()
	if val == "" {
		t.Errorf("field %q should not be empty", field)
	}
}

// createTestUser はテスト用ユーザーを作成しそのIDを返します
func createTestUser(t *testing.T, ts *testServer, name, email string) string {
	t.Helper()
	status, resp := ts.req(t, "POST", "/api/v1/users", map[string]string{
		"name":  name,
		"email": email,
	})
	assertStatus(t, status, http.StatusCreated, fmt.Sprintf("createUser(%s)", name))
	return mustGetString(t, resp, "data", "id")
}

// createTestProject はテスト用プロジェクトを作成しそのIDを返します
func createTestProject(t *testing.T, ts *testServer, key, name, ownerID string) string {
	t.Helper()
	status, resp := ts.req(t, "POST", "/api/v1/projects", map[string]string{
		"key":      key,
		"name":     name,
		"owner_id": ownerID,
	})
	assertStatus(t, status, http.StatusCreated, fmt.Sprintf("createProject(%s)", key))
	return mustGetString(t, resp, "data", "id")
}

// getFirstStatusID はプロジェクトの最初のステータスIDを返します
func getFirstStatusID(t *testing.T, ts *testServer, projectID string) string {
	t.Helper()
	status, resp := ts.req(t, "GET", "/api/v1/projects/"+projectID, nil)
	assertStatus(t, status, http.StatusOK, "getProject for status")
	data := resp["data"].(map[string]interface{})
	statuses := data["statuses"].([]interface{})
	if len(statuses) == 0 {
		t.Fatal("project has no statuses")
	}
	return statuses[0].(map[string]interface{})["id"].(string)
}
