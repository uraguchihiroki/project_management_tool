package main

import (
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/uraguchihiroki/project_management_tool/internal/handler"
	appdb "github.com/uraguchihiroki/project_management_tool/internal/db"
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

	db.SetupJoinTable(&model.User{}, "Roles", &model.UserRole{})

	if err := appdb.PrepareStatusesWorkflowColumn(db); err != nil {
		log.Fatalf("failed to prepare statuses.workflow_id (legacy DB): %v", err)
	}

	if err := appdb.MigrateIssueProjectStatusSplitPre(db); err != nil {
		log.Fatalf("failed to migrate issue/project status split (pre): %v", err)
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
		&model.Group{},
		&model.UserGroup{},
		&model.IssueGroup{},
		&model.TransitionAlertRule{},
	); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	if err := appdb.MigrateProjectStatusSeed(db); err != nil {
		log.Fatalf("failed to migrate project status seed: %v", err)
	}

	if err := appdb.MigrateStatusDedupeAndUniqueIndex(db); err != nil {
		log.Fatalf("failed to migrate status dedupe / unique index: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	statusRepo := repository.NewStatusRepository(db)
	issueRepo := repository.NewIssueRepository(db)
	issueEventRepo := repository.NewIssueEventRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	userGroupRepo := repository.NewUserGroupRepository(db)
	issueGroupRepo := repository.NewIssueGroupRepository(db)
	alertRuleRepo := repository.NewTransitionAlertRuleRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	workflowRepo := repository.NewWorkflowRepository(db)
	transitionRepo := repository.NewWorkflowTransitionRepository(db)
	projectStatusRepo := repository.NewProjectStatusRepository(db)
	projectStatusTransitionRepo := repository.NewProjectStatusTransitionRepository(db)
	templateRepo := repository.NewTemplateRepository(db)
	orgRepo := repository.NewOrganizationRepository(db)
	superAdminRepo := repository.NewSuperAdminRepository(db)
	departmentRepo := repository.NewDepartmentRepository(db)

	userSvc := service.NewUserService(userRepo, orgRepo)
	projectSvc := service.NewProjectService(projectRepo, statusRepo, workflowRepo, transitionRepo, projectStatusRepo, projectStatusTransitionRepo)
	orgSeedSvc := service.NewOrgSeedService(orgRepo, statusRepo, roleRepo, projectRepo, departmentRepo, issueRepo, workflowRepo, transitionRepo, projectStatusRepo, projectStatusTransitionRepo)
	orgSvc := service.NewOrganizationService(orgRepo, userRepo, orgSeedSvc)
	superAdminSvc := service.NewSuperAdminService(superAdminRepo)
	departmentSvc := service.NewDepartmentService(departmentRepo, orgRepo)
	alertEval := &service.TransitionAlertEvaluator{Rules: alertRuleRepo, UG: userGroupRepo}
	issueSvc := service.NewIssueService(issueRepo, projectRepo, statusRepo, workflowRepo, transitionRepo, issueEventRepo, groupRepo, issueGroupRepo, alertEval)
	commentSvc := service.NewCommentService(commentRepo, issueRepo)
	roleSvc := service.NewRoleService(roleRepo)
	workflowSvc := service.NewWorkflowService(workflowRepo)
	templateSvc := service.NewTemplateService(templateRepo, projectRepo)
	statusSvc := service.NewStatusService(statusRepo, workflowRepo, transitionRepo)
	groupSvc := service.NewGroupService(groupRepo, userGroupRepo)

	userHandler := handler.NewUserHandler(userSvc)
	projectHandler := handler.NewProjectHandler(projectSvc)
	issueHandler := handler.NewIssueHandler(issueSvc, projectSvc)
	commentHandler := handler.NewCommentHandler(commentSvc)
	roleHandler := handler.NewRoleHandler(roleSvc, userSvc)
	workflowHandler := handler.NewWorkflowHandler(workflowSvc)
	templateHandler := handler.NewTemplateHandler(templateSvc, projectSvc)
	orgHandler := handler.NewOrganizationHandler(orgSvc)
	superAdminHandler := handler.NewSuperAdminHandler(superAdminSvc, orgSvc)
	departmentHandler := handler.NewDepartmentHandler(departmentSvc)
	statusHandler := handler.NewStatusHandler(statusSvc, workflowSvc)
	issueEventHandler := handler.NewIssueEventHandler(issueRepo, issueEventRepo)
	groupHandler := handler.NewGroupHandler(groupSvc)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		// 開発: LAN IP や別ポートの Next（WSL / Windows 混在）を許可。本番はリバースプロキシ同オリジンが一般的。
		AllowOriginFunc: func(origin string) (bool, error) {
			if origin == "" {
				return true, nil
			}
			u, err := url.Parse(origin)
			if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
				return false, nil
			}
			host := u.Hostname()
			if host == "localhost" || host == "127.0.0.1" || host == "::1" {
				return true, nil
			}
			if host == "frontend" {
				return true, nil
			}
			if strings.HasPrefix(host, "192.168.") || strings.HasPrefix(host, "10.") || strings.HasPrefix(host, "172.") {
				return true, nil
			}
			return false, nil
		},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	e.Match([]string{"GET", "HEAD"}, "/api/v1/health", func(c echo.Context) error { return c.NoContent(200) })

	api := e.Group("/api/v1")
	api.Use(authmw.RequireJWT)

	public := e.Group("/api/v1")
	public.Use(authmw.OptionalJWT)

	public.POST("/users", userHandler.Create)
	public.POST("/admin/login", userHandler.AdminLogin)
	public.POST("/super-admin/login", superAdminHandler.Login)
	api.GET("/users", userHandler.List)
	api.GET("/users/:id/groups", groupHandler.ListByUser)
	api.GET("/users/:id", userHandler.Get)
	api.POST("/admin/switch-organization", userHandler.SwitchOrganization)
	api.PUT("/users/:id/admin", userHandler.SetAdmin)
	api.GET("/users/:id/roles", roleHandler.GetUserRoles)
	api.PUT("/users/:id/roles", roleHandler.AssignRoles)

	api.GET("/roles", roleHandler.List)
	api.POST("/roles", roleHandler.Create)
	api.PUT("/roles/bulk/reorder", roleHandler.Reorder)
	api.PUT("/roles/:id", roleHandler.Update)
	api.DELETE("/roles/:id", roleHandler.Delete)

	api.GET("/workflows", workflowHandler.List)
	api.POST("/workflows", workflowHandler.Create)
	api.PUT("/workflows/reorder", workflowHandler.Reorder)
	api.GET("/workflows/:id", workflowHandler.Get)
	api.GET("/workflows/:id/statuses", statusHandler.ListByWorkflow)
	api.POST("/workflows/:id/statuses", statusHandler.CreateForWorkflow)
	api.PUT("/workflows/:id", workflowHandler.Update)
	api.DELETE("/workflows/:id", workflowHandler.Delete)

	api.GET("/templates", templateHandler.List)
	api.POST("/templates", templateHandler.Create)
	api.GET("/templates/:id", templateHandler.Get)
	api.PUT("/templates/:id", templateHandler.Update)
	api.DELETE("/templates/:id", templateHandler.Delete)
	api.GET("/projects/:projectId/templates", templateHandler.ListByProject)
	api.PUT("/projects/:projectId/templates/reorder", templateHandler.Reorder)

	api.GET("/organizations", orgHandler.List)
	api.POST("/organizations", orgHandler.Create)
	api.GET("/users/:id/organizations", orgHandler.ListByUser)
	api.POST("/organizations/:orgId/users", orgHandler.AddUser)

	api.GET("/organizations/:orgId/departments", departmentHandler.List)
	api.POST("/organizations/:orgId/departments", departmentHandler.Create)
	api.PUT("/organizations/:orgId/departments/reorder", departmentHandler.Reorder)
	api.PUT("/organizations/:orgId/departments/:id", departmentHandler.Update)
	api.DELETE("/organizations/:orgId/departments/:id", departmentHandler.Delete)
	api.GET("/users/:id/departments", departmentHandler.GetUserDepartments)
	api.PUT("/users/:id/departments", departmentHandler.SetUserDepartments)

	api.GET("/super-admin/organizations", superAdminHandler.ListOrganizations)
	api.POST("/super-admin/organizations", superAdminHandler.CreateOrganization)

	api.GET("/admin/users", userHandler.ListWithRoles)
	api.POST("/admin/users", userHandler.CreateForOrg)
	api.PUT("/admin/users/:id", userHandler.UpdateUser)
	api.DELETE("/admin/users/:id", userHandler.RemoveFromOrg)

	api.GET("/projects", projectHandler.List)
	api.GET("/projects/:id/project-statuses", projectHandler.ListProjectStatuses)
	api.PUT("/projects/:id/project-statuses/:statusId", projectHandler.UpdateProjectStatus)
	api.GET("/organizations/:orgId/statuses", projectHandler.ListStatusesByOrg)
	api.POST("/organizations/:orgId/statuses", statusHandler.Create)
	api.PUT("/statuses/:id", statusHandler.Update)
	api.DELETE("/statuses/:id", statusHandler.Delete)
	api.POST("/projects", projectHandler.Create)
	api.PUT("/projects/reorder", projectHandler.Reorder)
	api.GET("/projects/:id", projectHandler.Get)
	api.PUT("/projects/:id", projectHandler.Update)
	api.DELETE("/projects/:id", projectHandler.Delete)

	api.GET("/organizations/:orgId/groups", groupHandler.List)
	api.POST("/organizations/:orgId/groups", groupHandler.Create)
	api.GET("/groups/:id/members", groupHandler.ListMembers)
	api.PUT("/groups/:id/members", groupHandler.ReplaceMembers)
	api.GET("/groups/:id", groupHandler.Get)
	api.PUT("/groups/:id", groupHandler.Update)
	api.DELETE("/groups/:id", groupHandler.Delete)

	api.GET("/projects/:projectId/issues", issueHandler.List)
	api.POST("/projects/:projectId/issues", issueHandler.Create)
	api.GET("/organizations/:orgId/issues", issueHandler.ListByOrg)
	api.POST("/organizations/:orgId/issues", issueHandler.CreateForOrg)
	api.GET("/organizations/:orgId/issues/:number", issueHandler.GetByOrgAndNumber)
	api.PUT("/organizations/:orgId/issues/:number", issueHandler.UpdateByOrgAndNumber)
	api.DELETE("/organizations/:orgId/issues/:number", issueHandler.DeleteByOrgAndNumber)
	api.GET("/projects/:projectId/issues/:number/groups", issueHandler.ListIssueGroups)
	api.PUT("/projects/:projectId/issues/:number/groups", issueHandler.PutIssueGroups)
	api.GET("/projects/:projectId/issues/:number", issueHandler.Get)
	api.PUT("/projects/:projectId/issues/:number", issueHandler.Update)
	api.DELETE("/projects/:projectId/issues/:number", issueHandler.Delete)
	api.GET("/organizations/:orgId/issue-events", issueEventHandler.ListByOrganization)

	api.GET("/issues/:issueId/events", issueEventHandler.ListByIssue)
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
