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
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/uraguchihiroki/project_management_tool/internal/handler"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
	"github.com/uraguchihiroki/project_management_tool/internal/service"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// testOrgID はテスト用の固定組織ID
const testOrgID = "00000000-0000-0000-0000-000000000001"

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
		&model.Organization{},
		&model.SuperAdmin{},
		&model.Role{},
		&model.User{},
		&model.OrganizationUser{},
		&model.Department{},
		&model.OrganizationUserDepartment{},
		&model.Project{},
		&model.Status{},
		&model.Issue{},
		&model.Comment{},
		&model.Workflow{},
		&model.WorkflowStep{},
		&model.ApprovalObject{},
		&model.IssueTemplate{},
		&model.IssueApproval{},
	); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	// テスト用固定組織（FRS）をシード
	frsOrg := model.Organization{
		ID:        uuid.MustParse(testOrgID),
		Name:      "FRS",
		CreatedAt: time.Now(),
	}
	db.Create(&frsOrg)

	// 組織用デフォルトステータス（プロジェクト未割当Issue用）
	statusRepo := repository.NewStatusRepository(db)
	for _, ds := range []struct {
		Name  string
		Color string
		Order int
	}{
		{"未着手", "#6B7280", 1},
		{"進行中", "#3B82F6", 2},
		{"完了", "#10B981", 3},
	} {
		statusRepo.Create(&model.Status{
			ID:             uuid.New(),
			ProjectID:      nil,
			OrganizationID: &frsOrg.ID,
			Name:           ds.Name,
			Color:          ds.Color,
			Order:          ds.Order,
			Type:           "issue",
		})
	}

	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	issueRepo := repository.NewIssueRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	workflowRepo := repository.NewWorkflowRepository(db)
	templateRepo := repository.NewTemplateRepository(db)
	approvalRepo := repository.NewApprovalRepository(db)
	orgRepo := repository.NewOrganizationRepository(db)
	superAdminRepo := repository.NewSuperAdminRepository(db)
	departmentRepo := repository.NewDepartmentRepository(db)

	userSvc := service.NewUserService(userRepo, orgRepo)
	projectSvc := service.NewProjectService(projectRepo, statusRepo)
	orgSvc := service.NewOrganizationService(orgRepo, userRepo)
	superAdminSvc := service.NewSuperAdminService(superAdminRepo)
	departmentSvc := service.NewDepartmentService(departmentRepo, orgRepo)
	issueSvc := service.NewIssueService(issueRepo, projectRepo)
	commentSvc := service.NewCommentService(commentRepo)
	roleSvc := service.NewRoleService(roleRepo)
	workflowSvc := service.NewWorkflowService(workflowRepo)
	templateSvc := service.NewTemplateService(templateRepo)
	approvalSvc := service.NewApprovalService(approvalRepo, workflowRepo, issueRepo, roleRepo)
	statusSvc := service.NewStatusService(statusRepo)

	userH := handler.NewUserHandler(userSvc)
	projectH := handler.NewProjectHandler(projectSvc)
	issueH := handler.NewIssueHandler(issueSvc, approvalSvc)
	commentH := handler.NewCommentHandler(commentSvc)
	roleH := handler.NewRoleHandler(roleSvc, userSvc)
	workflowH := handler.NewWorkflowHandler(workflowSvc)
	templateH := handler.NewTemplateHandler(templateSvc)
	approvalH := handler.NewApprovalHandler(approvalSvc)
	orgH := handler.NewOrganizationHandler(orgSvc)
	superAdminH := handler.NewSuperAdminHandler(superAdminSvc, orgSvc)
	departmentH := handler.NewDepartmentHandler(departmentSvc)
	statusH := handler.NewStatusHandler(statusSvc)

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())

	api := e.Group("/api/v1")
	api.GET("/users", userH.List)
	api.POST("/users", userH.Create)
	api.GET("/users/:id", userH.Get)
	api.PUT("/users/:id/admin", userH.SetAdmin)
	api.GET("/users/:id/roles", roleH.GetUserRoles)
	api.PUT("/users/:id/roles", roleH.AssignRoles)
	api.GET("/roles", roleH.List)
	api.POST("/roles", roleH.Create)
	api.PUT("/roles/bulk/reorder", roleH.Reorder)
	api.PUT("/roles/:id", roleH.Update)
	api.DELETE("/roles/:id", roleH.Delete)
	api.GET("/workflows", workflowH.List)
	api.POST("/workflows", workflowH.Create)
	api.PUT("/workflows/reorder", workflowH.Reorder)
	api.GET("/workflows/:id", workflowH.Get)
	api.PUT("/workflows/:id", workflowH.Update)
	api.DELETE("/workflows/:id", workflowH.Delete)
	api.POST("/workflows/:id/steps", workflowH.AddStep)
	api.GET("/workflows/:id/steps/:stepId", workflowH.GetStep)
	api.PUT("/workflows/:id/steps/reorder", workflowH.ReorderSteps)
	api.PUT("/workflows/:id/steps/:stepId", workflowH.UpdateStep)
	api.DELETE("/workflows/:id/steps/:stepId", workflowH.DeleteStep)
	api.GET("/templates", templateH.List)
	api.POST("/templates", templateH.Create)
	api.GET("/templates/:id", templateH.Get)
	api.PUT("/templates/:id", templateH.Update)
	api.DELETE("/templates/:id", templateH.Delete)
	api.GET("/projects/:projectId/templates", templateH.ListByProject)
	api.PUT("/projects/:projectId/templates/reorder", templateH.Reorder)
	api.GET("/issues/:issueId/approvals", approvalH.List)
	api.POST("/approvals/:id/approve", approvalH.Approve)
	api.POST("/issues/:issueId/steps/:stepId/approve", approvalH.ApproveStep)
	api.POST("/approvals/:id/reject", approvalH.Reject)
	api.GET("/organizations", orgH.List)
	api.POST("/organizations", orgH.Create)
	api.GET("/users/:id/organizations", orgH.ListByUser)
	api.POST("/organizations/:orgId/users", orgH.AddUser)
	api.GET("/organizations/:orgId/departments", departmentH.List)
	api.POST("/organizations/:orgId/departments", departmentH.Create)
	api.PUT("/organizations/:orgId/departments/reorder", departmentH.Reorder)
	api.PUT("/organizations/:orgId/departments/:id", departmentH.Update)
	api.DELETE("/organizations/:orgId/departments/:id", departmentH.Delete)
	api.GET("/users/:id/departments", departmentH.GetUserDepartments)
	api.PUT("/users/:id/departments", departmentH.SetUserDepartments)
	api.POST("/super-admin/login", superAdminH.Login)
	api.GET("/super-admin/organizations", superAdminH.ListOrganizations)
	api.POST("/super-admin/organizations", superAdminH.CreateOrganization)
	api.GET("/admin/users", userH.ListWithRoles)
	api.POST("/admin/users", userH.CreateForOrg)
	api.PUT("/admin/users/:id", userH.UpdateUser)
	api.DELETE("/admin/users/:id", userH.RemoveFromOrg)
	api.GET("/projects", projectH.List)
	api.GET("/organizations/:orgId/statuses", projectH.ListStatusesByOrg)
	api.POST("/organizations/:orgId/statuses", statusH.Create)
	api.PUT("/statuses/:id", statusH.Update)
	api.DELETE("/statuses/:id", statusH.Delete)
	api.POST("/projects", projectH.Create)
	api.PUT("/projects/reorder", projectH.Reorder)
	api.GET("/projects/:id", projectH.Get)
	api.PUT("/projects/:id", projectH.Update)
	api.DELETE("/projects/:id", projectH.Delete)
	api.GET("/projects/:projectId/issues", issueH.List)
	api.POST("/projects/:projectId/issues", issueH.Create)
	api.GET("/organizations/:orgId/issues", issueH.ListByOrg)
	api.POST("/organizations/:orgId/issues", issueH.CreateForOrg)
	api.GET("/organizations/:orgId/issues/:number", issueH.GetByOrgAndNumber)
	api.PUT("/organizations/:orgId/issues/:number", issueH.UpdateByOrgAndNumber)
	api.DELETE("/organizations/:orgId/issues/:number", issueH.DeleteByOrgAndNumber)
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
		"key":             key,
		"name":            name,
		"owner_id":        ownerID,
		"organization_id": testOrgID,
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
