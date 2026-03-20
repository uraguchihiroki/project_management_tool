package main

import (
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/uraguchihiroki/project_management_tool/internal/handler"
	authmw "github.com/uraguchihiroki/project_management_tool/internal/middleware"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
	"github.com/uraguchihiroki/project_management_tool/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=pmt_user password=pmt_password dbname=pmt_db port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// user_roles 中間テーブルに key カラムを持たせるためカスタム JoinTable を設定
	db.SetupJoinTable(&model.User{}, "Roles", &model.UserRole{})

	// AutoMigrate
	if err := db.AutoMigrate(
		&model.Organization{},
		&model.SuperAdmin{},
		&model.Role{},
		&model.User{},
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
		log.Fatalf("failed to migrate: %v", err)
	}

	// Repositories
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	statusRepo := repository.NewStatusRepository(db)
	issueRepo := repository.NewIssueRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	workflowRepo := repository.NewWorkflowRepository(db)
	templateRepo := repository.NewTemplateRepository(db)
	approvalRepo := repository.NewApprovalRepository(db)
	orgRepo := repository.NewOrganizationRepository(db)
	superAdminRepo := repository.NewSuperAdminRepository(db)
	departmentRepo := repository.NewDepartmentRepository(db)

	// Services
	userSvc := service.NewUserService(userRepo, orgRepo)
	projectSvc := service.NewProjectService(projectRepo, statusRepo)
	orgSeedSvc := service.NewOrgSeedService(orgRepo, statusRepo, roleRepo, projectRepo, departmentRepo, issueRepo)
	orgSvc := service.NewOrganizationService(orgRepo, userRepo, orgSeedSvc)
	superAdminSvc := service.NewSuperAdminService(superAdminRepo)
	departmentSvc := service.NewDepartmentService(departmentRepo, orgRepo)
	issueSvc := service.NewIssueService(issueRepo, projectRepo)
	commentSvc := service.NewCommentService(commentRepo, issueRepo)
	roleSvc := service.NewRoleService(roleRepo)
	workflowSvc := service.NewWorkflowService(workflowRepo, statusRepo)
	templateSvc := service.NewTemplateService(templateRepo, projectRepo)
	approvalSvc := service.NewApprovalService(approvalRepo, workflowRepo, issueRepo, roleRepo)
	statusSvc := service.NewStatusService(statusRepo)

	// Handlers
	userHandler := handler.NewUserHandler(userSvc)
	projectHandler := handler.NewProjectHandler(projectSvc)
	issueHandler := handler.NewIssueHandler(issueSvc, approvalSvc, projectSvc)
	commentHandler := handler.NewCommentHandler(commentSvc)
	roleHandler := handler.NewRoleHandler(roleSvc, userSvc)
	workflowHandler := handler.NewWorkflowHandler(workflowSvc)
	templateHandler := handler.NewTemplateHandler(templateSvc, projectSvc)
	approvalHandler := handler.NewApprovalHandler(approvalSvc)
	orgHandler := handler.NewOrganizationHandler(orgSvc)
	superAdminHandler := handler.NewSuperAdminHandler(superAdminSvc, orgSvc)
	departmentHandler := handler.NewDepartmentHandler(departmentSvc)
	statusHandler := handler.NewStatusHandler(statusSvc)

	// Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		// localhost と 127.0.0.1 はブラウザ上で別オリジンになる（Playwright の baseURL 等）
		AllowOrigins: []string{"http://localhost:3000", "http://127.0.0.1:3000", "http://frontend:3000"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	// Health check (for Docker / E2E; wget --spider uses HEAD)
	e.Match([]string{"GET", "HEAD"}, "/api/v1/health", func(c echo.Context) error { return c.NoContent(200) })

	// ステップ更新: 最優先で登録（/workflows/:id との競合を完全回避）
	e.PUT("/api/v1/workflow-steps/:stepId", workflowHandler.UpdateStep, authmw.RequireJWT)

	// Routes
	api := e.Group("/api/v1")
	api.Use(authmw.RequireJWT)

	public := e.Group("/api/v1")
	public.Use(authmw.OptionalJWT)

	// Users
	public.POST("/users", userHandler.Create)
	public.POST("/admin/login", userHandler.AdminLogin)
	public.POST("/super-admin/login", superAdminHandler.Login)
	api.GET("/users", userHandler.List)
	api.GET("/users/:id", userHandler.Get)
	api.POST("/admin/switch-organization", userHandler.SwitchOrganization)
	api.PUT("/users/:id/admin", userHandler.SetAdmin)
	api.GET("/users/:id/roles", roleHandler.GetUserRoles)
	api.PUT("/users/:id/roles", roleHandler.AssignRoles)

	// Roles
	api.GET("/roles", roleHandler.List)
	api.POST("/roles", roleHandler.Create)
	api.PUT("/roles/bulk/reorder", roleHandler.Reorder)
	api.PUT("/roles/:id", roleHandler.Update)
	api.DELETE("/roles/:id", roleHandler.Delete)

	// Workflows（組織に属さない、グローバル）
	// /workflows/:id より具体的な /workflows/:id/steps/ を先に登録
	api.GET("/workflows", workflowHandler.List)
	api.POST("/workflows", workflowHandler.Create)
	api.PUT("/workflows/reorder", workflowHandler.Reorder)
	api.POST("/workflows/:id/steps", workflowHandler.AddStep)
	api.GET("/workflows/:id/steps/:stepId", workflowHandler.GetStep)
	api.PUT("/workflows/:id/steps/reorder", workflowHandler.ReorderSteps)
	api.PUT("/workflows/:id/steps/:stepId", workflowHandler.UpdateStep)
	api.DELETE("/workflows/:id/steps/:stepId", workflowHandler.DeleteStep)
	api.GET("/workflows/:id", workflowHandler.Get)
	api.PUT("/workflows/:id", workflowHandler.Update)
	api.DELETE("/workflows/:id", workflowHandler.Delete)

	// Templates
	api.GET("/templates", templateHandler.List)
	api.POST("/templates", templateHandler.Create)
	api.GET("/templates/:id", templateHandler.Get)
	api.PUT("/templates/:id", templateHandler.Update)
	api.DELETE("/templates/:id", templateHandler.Delete)
	api.GET("/projects/:projectId/templates", templateHandler.ListByProject)
	api.PUT("/projects/:projectId/templates/reorder", templateHandler.Reorder)

	// Approvals
	api.GET("/issues/:issueId/approvals", approvalHandler.List)
	api.POST("/approvals/:id/approve", approvalHandler.Approve)
	api.POST("/issues/:issueId/steps/:stepId/approve", approvalHandler.ApproveStep)
	api.POST("/approvals/:id/reject", approvalHandler.Reject)

	// Organizations
	api.GET("/organizations", orgHandler.List)
	api.POST("/organizations", orgHandler.Create)
	api.GET("/users/:id/organizations", orgHandler.ListByUser)
	api.POST("/organizations/:orgId/users", orgHandler.AddUser)

	// Departments
	api.GET("/organizations/:orgId/departments", departmentHandler.List)
	api.POST("/organizations/:orgId/departments", departmentHandler.Create)
	api.PUT("/organizations/:orgId/departments/reorder", departmentHandler.Reorder)
	api.PUT("/organizations/:orgId/departments/:id", departmentHandler.Update)
	api.DELETE("/organizations/:orgId/departments/:id", departmentHandler.Delete)
	api.GET("/users/:id/departments", departmentHandler.GetUserDepartments)
	api.PUT("/users/:id/departments", departmentHandler.SetUserDepartments)

	// Super Admin
	api.GET("/super-admin/organizations", superAdminHandler.ListOrganizations)
	api.POST("/super-admin/organizations", superAdminHandler.CreateOrganization)

	// Admin
	api.GET("/admin/users", userHandler.ListWithRoles)
	api.POST("/admin/users", userHandler.CreateForOrg)
	api.PUT("/admin/users/:id", userHandler.UpdateUser)
	api.DELETE("/admin/users/:id", userHandler.RemoveFromOrg)

	// Projects
	api.GET("/projects", projectHandler.List)
	api.GET("/organizations/:orgId/statuses", projectHandler.ListStatusesByOrg)
	api.POST("/organizations/:orgId/statuses", statusHandler.Create)
	api.PUT("/statuses/:id", statusHandler.Update)
	api.DELETE("/statuses/:id", statusHandler.Delete)
	api.POST("/projects", projectHandler.Create)
	api.PUT("/projects/reorder", projectHandler.Reorder)
	api.GET("/projects/:id", projectHandler.Get)
	api.PUT("/projects/:id", projectHandler.Update)
	api.DELETE("/projects/:id", projectHandler.Delete)

	// Issues
	api.GET("/projects/:projectId/issues", issueHandler.List)
	api.POST("/projects/:projectId/issues", issueHandler.Create)
	api.GET("/organizations/:orgId/issues", issueHandler.ListByOrg)
	api.POST("/organizations/:orgId/issues", issueHandler.CreateForOrg)
	api.GET("/organizations/:orgId/issues/:number", issueHandler.GetByOrgAndNumber)
	api.PUT("/organizations/:orgId/issues/:number", issueHandler.UpdateByOrgAndNumber)
	api.DELETE("/organizations/:orgId/issues/:number", issueHandler.DeleteByOrgAndNumber)
	api.GET("/projects/:projectId/issues/:number", issueHandler.Get)
	api.PUT("/projects/:projectId/issues/:number", issueHandler.Update)
	api.DELETE("/projects/:projectId/issues/:number", issueHandler.Delete)

	// Comments
	api.GET("/issues/:issueId/comments", commentHandler.List)
	api.POST("/issues/:issueId/comments", commentHandler.Create)
	api.PUT("/issues/:issueId/comments/:id", commentHandler.Update)
	api.DELETE("/issues/:issueId/comments/:id", commentHandler.Delete)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on :%s", port)
	e.Logger.Fatal(e.Start(":" + port))
}
