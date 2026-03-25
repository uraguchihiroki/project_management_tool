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
	"github.com/uraguchihiroki/project_management_tool/internal/auth"
	appdb "github.com/uraguchihiroki/project_management_tool/internal/db"
	"github.com/uraguchihiroki/project_management_tool/internal/handler"
	authmw "github.com/uraguchihiroki/project_management_tool/internal/middleware"
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
	token  string
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

	db.SetupJoinTable(&model.User{}, "Roles", &model.UserRole{})

	if err := appdb.MigrateIssueProjectStatusSplitPre(db); err != nil {
		t.Fatalf("failed migrate issue/project status split (pre): %v", err)
	}

	if err := appdb.MigrateJunctionTablesSurrogatePK(db); err != nil {
		t.Fatalf("failed migrate junction tables surrogate PK: %v", err)
	}

	if err := db.AutoMigrate(
		&model.Organization{},
		&model.SuperAdmin{},
		&model.Role{},
		&model.User{},
		&model.Department{},
		&model.OrganizationUserDepartment{},
		&model.Project{},
		&model.Workflow{},
		&model.Status{},
		&model.WorkflowTransition{},
		&model.ProjectStatus{},
		&model.ProjectStatusTransition{},
		&model.Issue{},
		&model.Comment{},
		&model.IssueTemplate{},
		&model.IssueEvent{},
		&model.TransitionAlertRule{},
	); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	if err := appdb.MigrateDropLegacyBusinessUniqueIndexes(db); err != nil {
		t.Fatalf("failed to drop legacy unique indexes: %v", err)
	}

	if err := appdb.MigrateProjectStatusSeed(db); err != nil {
		t.Fatalf("failed migrate project status seed: %v", err)
	}

	if err := appdb.MigrateStatusOrderToDisplayOrder(db); err != nil {
		t.Fatalf("failed migrate status order column: %v", err)
	}

	if err := appdb.MigrateWorkflowTransitionDisplayOrder(db); err != nil {
		t.Fatalf("failed migrate workflow transition display_order: %v", err)
	}

	if err := appdb.MigrateStatusDedupe(db); err != nil {
		t.Fatalf("failed to migrate status dedupe: %v", err)
	}
	if err := appdb.MigrateJunctionOrganizationID(db); err != nil {
		t.Fatalf("failed migrate junction organization_id: %v", err)
	}
	if err := appdb.MigrateDropGroupTables(db); err != nil {
		t.Fatalf("failed migrate drop group tables: %v", err)
	}

	frsOrg := model.Organization{
		ID:        uuid.MustParse(testOrgID),
		Key:       testOrgID,
		Name:      "FRS",
		CreatedAt: time.Now(),
	}
	db.Create(&frsOrg)

	statusRepo := repository.NewStatusRepository(db)
	workflowRepo := repository.NewWorkflowRepository(db)
	transitionRepo := repository.NewWorkflowTransitionRepository(db)
	orgIssueWfID, orgIssueStatusIDs, err := service.CreateOrgIssueWorkflowWithDefaultStatuses(workflowRepo, statusRepo, frsOrg.ID, "組織Issue")
	if err != nil {
		t.Fatalf("seed 組織Issue workflow: %v", err)
	}
	if err := service.SeedDefaultIssueWorkflowTransitions(transitionRepo, orgIssueWfID, orgIssueStatusIDs); err != nil {
		t.Fatalf("seed 組織Issue workflow transitions: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	issueRepo := repository.NewIssueRepository(db)
	issueEventRepo := repository.NewIssueEventRepository(db)
	alertRuleRepo := repository.NewTransitionAlertRuleRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	templateRepo := repository.NewTemplateRepository(db)
	orgRepo := repository.NewOrganizationRepository(db)
	superAdminRepo := repository.NewSuperAdminRepository(db)
	departmentRepo := repository.NewDepartmentRepository(db)

	projectStatusRepo := repository.NewProjectStatusRepository(db)
	projectStatusTransitionRepo := repository.NewProjectStatusTransitionRepository(db)

	issueWFProv := service.NewIssueWorkflowProvisioner(projectRepo, workflowRepo, statusRepo, transitionRepo)

	userSvc := service.NewUserService(userRepo, orgRepo)
	projectSvc := service.NewProjectService(projectRepo, statusRepo, projectStatusRepo, projectStatusTransitionRepo)
	orgSeedSvc := service.NewOrgSeedService(orgRepo, statusRepo, roleRepo, projectRepo, departmentRepo, issueRepo, workflowRepo, transitionRepo, projectStatusRepo, issueWFProv)
	orgSvc := service.NewOrganizationService(orgRepo, userRepo, orgSeedSvc)
	superAdminSvc := service.NewSuperAdminService(superAdminRepo)
	departmentSvc := service.NewDepartmentService(departmentRepo, orgRepo)
	alertEval := &service.TransitionAlertEvaluator{Rules: alertRuleRepo}
	issueSvc := service.NewIssueService(issueRepo, projectRepo, statusRepo, workflowRepo, transitionRepo, issueEventRepo, alertEval, issueWFProv)
	commentSvc := service.NewCommentService(commentRepo, issueRepo)
	roleSvc := service.NewRoleService(roleRepo)
	workflowSvc := service.NewWorkflowService(workflowRepo)
	templateSvc := service.NewTemplateService(templateRepo, projectRepo)
	statusSvc := service.NewStatusService(statusRepo, workflowRepo, transitionRepo)

	userH := handler.NewUserHandler(userSvc)
	projectH := handler.NewProjectHandler(projectSvc, issueWFProv)
	issueH := handler.NewIssueHandler(issueSvc, projectSvc)
	commentH := handler.NewCommentHandler(commentSvc)
	roleH := handler.NewRoleHandler(roleSvc, userSvc)
	workflowH := handler.NewWorkflowHandler(workflowSvc)
	workflowTransitionH := handler.NewWorkflowTransitionHandler(workflowSvc, statusSvc, transitionRepo)
	templateH := handler.NewTemplateHandler(templateSvc, projectSvc)
	orgH := handler.NewOrganizationHandler(orgSvc)
	superAdminH := handler.NewSuperAdminHandler(superAdminSvc, orgSvc)
	departmentH := handler.NewDepartmentHandler(departmentSvc)
	statusH := handler.NewStatusHandler(statusSvc, workflowSvc)
	issueEventH := handler.NewIssueEventHandler(issueRepo, issueEventRepo)

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())

	superAdmin := model.SuperAdmin{
		ID:    uuid.New(),
		Email: "test-super-admin@example.com",
	}
	db.Create(&superAdmin)
	token, _ := auth.GenerateSuperAdminToken(superAdmin.ID)

	api := e.Group("/api/v1")
	api.Use(authmw.RequireJWT)
	public := e.Group("/api/v1")
	public.Use(authmw.OptionalJWT)
	api.GET("/users", userH.List)
	public.POST("/users", userH.Create)
	public.POST("/admin/login", userH.AdminLogin)
	public.POST("/super-admin/login", superAdminH.Login)
	api.POST("/admin/switch-organization", userH.SwitchOrganization)
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
	api.GET("/workflows/:id/statuses", statusH.ListByWorkflow)
	api.POST("/workflows/:id/statuses", statusH.CreateForWorkflow)
	api.PUT("/workflows/:id/statuses/reorder", statusH.ReorderForWorkflow)
	api.GET("/workflows/:id/transitions", workflowTransitionH.ListByWorkflow)
	api.PUT("/workflows/:id/transitions/reorder", workflowTransitionH.ReorderForWorkflow)
	api.POST("/workflows/:id/transitions", workflowTransitionH.CreateForWorkflow)
	api.PUT("/workflows/:id/transitions/:transitionId", workflowTransitionH.Update)
	api.DELETE("/workflows/:id/transitions/:transitionId", workflowTransitionH.Delete)
	api.PUT("/workflows/:id", workflowH.Update)
	api.DELETE("/workflows/:id", workflowH.Delete)
	api.GET("/templates", templateH.List)
	api.POST("/templates", templateH.Create)
	api.GET("/templates/:id", templateH.Get)
	api.PUT("/templates/:id", templateH.Update)
	api.DELETE("/templates/:id", templateH.Delete)
	api.GET("/projects/:projectId/templates", templateH.ListByProject)
	api.PUT("/projects/:projectId/templates/reorder", templateH.Reorder)
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
	api.GET("/super-admin/organizations", superAdminH.ListOrganizations)
	api.POST("/super-admin/organizations", superAdminH.CreateOrganization)
	api.GET("/admin/users", userH.ListWithRoles)
	api.POST("/admin/users", userH.CreateForOrg)
	api.PUT("/admin/users/:id", userH.UpdateUser)
	api.DELETE("/admin/users/:id", userH.RemoveFromOrg)
	api.GET("/projects", projectH.List)
	api.GET("/projects/:id/project-statuses", projectH.ListProjectStatuses)
	api.PUT("/projects/:id/project-statuses/:statusId", projectH.UpdateProjectStatus)
	api.GET("/organizations/:orgId/statuses", projectH.ListStatusesByOrg)
	api.POST("/organizations/:orgId/statuses", statusH.Create)
	api.PUT("/statuses/:id", statusH.Update)
	api.DELETE("/statuses/:id", statusH.Delete)
	api.POST("/projects", projectH.Create)
	api.POST("/projects/:id/default-issue-workflow", projectH.EnsureDefaultIssueWorkflow)
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
	api.GET("/organizations/:orgId/issue-events", issueEventH.ListByOrganization)
	api.GET("/issues/:issueId/events", issueEventH.ListByIssue)
	api.GET("/issues/:issueId/comments", commentH.List)
	api.POST("/issues/:issueId/comments", commentH.Create)
	api.PUT("/issues/:issueId/comments/:id", commentH.Update)
	api.DELETE("/issues/:issueId/comments/:id", commentH.Delete)

	srv := httptest.NewServer(e)
	t.Cleanup(func() {
		srv.Close()
	})

	return &testServer{server: srv, db: db, token: token}
}

// req はテスト用HTTPリクエストを送信し、レスポンスのbodyを返します
func (ts *testServer) req(t *testing.T, method, path string, body interface{}) (int, map[string]interface{}) {
	t.Helper()
	return ts.reqWithToken(t, ts.token, method, path, body)
}

// reqNoAuth は Authorization ヘッダなし（公開エンドポイントやログイン用）
func (ts *testServer) reqNoAuth(t *testing.T, method, path string, body interface{}) (int, map[string]interface{}) {
	t.Helper()
	return ts.reqWithToken(t, "", method, path, body)
}

func (ts *testServer) reqWithToken(t *testing.T, token, method, path string, body interface{}) (int, map[string]interface{}) {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}

	req, _ := http.NewRequest(method, ts.server.URL+path, bodyReader)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
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
	projectID := mustGetString(t, resp, "data", "id")
	st2, _ := ts.req(t, "POST", "/api/v1/projects/"+projectID+"/default-issue-workflow", nil)
	assertStatus(t, st2, http.StatusOK, "default-issue-workflow after createTestProject")
	return projectID
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

// getStatusIDs はプロジェクトのステータスIDを返します
func getStatusIDs(t *testing.T, ts *testServer, projectID string) []string {
	t.Helper()
	status, resp := ts.req(t, "GET", "/api/v1/projects/"+projectID, nil)
	assertStatus(t, status, http.StatusOK, "getProject for status")
	data := resp["data"].(map[string]interface{})
	statuses := data["statuses"].([]interface{})
	ids := make([]string, len(statuses))
	for i, s := range statuses {
		ids[i] = s.(map[string]interface{})["id"].(string)
	}
	return ids
}
