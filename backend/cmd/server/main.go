package main

import (
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	appdb "github.com/uraguchihiroki/project_management_tool/internal/db"
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

	db.SetupJoinTable(&model.User{}, "Roles", &model.UserRole{})

	if err := appdb.PrepareStatusesWorkflowColumn(db); err != nil {
		log.Fatalf("failed to prepare statuses.workflow_id (legacy DB): %v", err)
	}

	if err := appdb.MigrateIssueProjectStatusSplitPre(db); err != nil {
		log.Fatalf("failed to migrate issue/project status split (pre): %v", err)
	}

	if err := appdb.MigrateRenameDepartmentsToGroups(db); err != nil {
		log.Fatalf("failed migrate rename departments to groups: %v", err)
	}

	if err := appdb.MigrateJunctionTablesSurrogatePK(db); err != nil {
		log.Fatalf("failed to migrate junction tables surrogate PK: %v", err)
	}

	if err := db.AutoMigrate(
		&model.Organization{},
		&model.SuperAdmin{},
		&model.Role{},
		&model.User{},
		&model.Group{},
		&model.OrganizationUserGroup{},
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
		log.Fatalf("failed to migrate: %v", err)
	}

	if err := appdb.MigrateDropLegacyBusinessUniqueIndexes(db); err != nil {
		log.Fatalf("failed to drop legacy unique indexes: %v", err)
	}

	if err := appdb.MigrateProjectStatusSeed(db); err != nil {
		log.Fatalf("failed to migrate project status seed: %v", err)
	}

	if err := appdb.MigrateStatusOrderToDisplayOrder(db); err != nil {
		log.Fatalf("failed migrate status order column: %v", err)
	}

	if err := appdb.MigrateWorkflowTransitionDisplayOrder(db); err != nil {
		log.Fatalf("failed migrate workflow transition display_order: %v", err)
	}

	if err := appdb.MigrateStatusDedupe(db); err != nil {
		log.Fatalf("failed to migrate status dedupe: %v", err)
	}
	if err := appdb.MigrateJunctionOrganizationID(db); err != nil {
		log.Fatalf("failed migrate junction organization_id: %v", err)
	}
	if err := appdb.MigrateRemoveLegacyGlobalIssueStatuses(db); err != nil {
		log.Fatalf("failed migrate remove legacy global issue statuses: %v", err)
	}
	if err := appdb.MigrateStatusEntryUniqueIndex(db); err != nil {
		log.Fatalf("failed migrate status entry unique index: %v", err)
	}
	if err := appdb.MigrateEnsureDefaultIssueEntryStatus(db); err != nil {
		log.Fatalf("failed migrate ensure default issue entry status: %v", err)
	}
	if err := appdb.MigrateDropGroupTables(db); err != nil {
		log.Fatalf("failed migrate drop group tables: %v", err)
	}
	if err := appdb.MigrateDropApprovalTables(db); err != nil {
		log.Fatalf("failed migrate drop approval tables: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	statusRepo := repository.NewStatusRepository(db)
	issueRepo := repository.NewIssueRepository(db)
	issueEventRepo := repository.NewIssueEventRepository(db)
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
	groupRepo := repository.NewGroupRepository(db)

	userSvc := service.NewUserService(userRepo, orgRepo)
	issueWFProv := service.NewIssueWorkflowProvisioner(projectRepo, workflowRepo, statusRepo, transitionRepo)
	projectSvc := service.NewProjectService(projectRepo, statusRepo, projectStatusRepo, projectStatusTransitionRepo)
	orgSeedSvc := service.NewOrgSeedService(orgRepo, statusRepo, roleRepo, projectRepo, groupRepo, issueRepo, workflowRepo, transitionRepo, projectStatusRepo, issueWFProv)
	orgSvc := service.NewOrganizationService(orgRepo, userRepo, orgSeedSvc)
	superAdminSvc := service.NewSuperAdminService(superAdminRepo)
	groupSvc := service.NewGroupService(groupRepo, orgRepo)
	alertEval := &service.TransitionAlertEvaluator{Rules: alertRuleRepo}
	issueSvc := service.NewIssueService(issueRepo, projectRepo, statusRepo, workflowRepo, transitionRepo, issueEventRepo, alertEval, issueWFProv)
	commentSvc := service.NewCommentService(commentRepo, issueRepo)
	roleSvc := service.NewRoleService(roleRepo)
	workflowSvc := service.NewWorkflowService(workflowRepo)
	templateSvc := service.NewTemplateService(templateRepo, projectRepo)
	statusSvc := service.NewStatusService(statusRepo, workflowRepo, transitionRepo)

	userHandler := handler.NewUserHandler(userSvc)
	projectHandler := handler.NewProjectHandler(projectSvc, issueWFProv)
	issueHandler := handler.NewIssueHandler(issueSvc, projectSvc)
	commentHandler := handler.NewCommentHandler(commentSvc)
	roleHandler := handler.NewRoleHandler(roleSvc, userSvc)
	workflowHandler := handler.NewWorkflowHandler(workflowSvc)
	workflowTransitionHandler := handler.NewWorkflowTransitionHandler(workflowSvc, statusSvc, transitionRepo)
	templateHandler := handler.NewTemplateHandler(templateSvc, projectSvc)
	orgHandler := handler.NewOrganizationHandler(orgSvc)
	superAdminHandler := handler.NewSuperAdminHandler(superAdminSvc, orgSvc)
	groupHandler := handler.NewGroupHandler(groupSvc)
	statusHandler := handler.NewStatusHandler(statusSvc, workflowSvc)
	issueEventHandler := handler.NewIssueEventHandler(issueRepo, issueEventRepo)

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
	api.PUT("/workflows/:id/statuses/reorder", statusHandler.ReorderForWorkflow)
	api.GET("/workflows/:id/transitions", workflowTransitionHandler.ListByWorkflow)
	api.PUT("/workflows/:id/transitions/reorder", workflowTransitionHandler.ReorderForWorkflow)
	api.POST("/workflows/:id/transitions", workflowTransitionHandler.CreateForWorkflow)
	api.PUT("/workflows/:id/transitions/:transitionId", workflowTransitionHandler.Update)
	api.DELETE("/workflows/:id/transitions/:transitionId", workflowTransitionHandler.Delete)
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

	api.GET("/organizations/:orgId/groups", groupHandler.List)
	api.POST("/organizations/:orgId/groups", groupHandler.Create)
	api.PUT("/organizations/:orgId/groups/reorder", groupHandler.Reorder)
	api.PUT("/organizations/:orgId/groups/:id", groupHandler.Update)
	api.DELETE("/organizations/:orgId/groups/:id", groupHandler.Delete)
	api.GET("/users/:id/groups", groupHandler.GetUserGroups)
	api.PUT("/users/:id/groups", groupHandler.SetUserGroups)

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
	api.POST("/projects/:id/default-issue-workflow", projectHandler.EnsureDefaultIssueWorkflow)
	api.PUT("/projects/reorder", projectHandler.Reorder)
	api.GET("/projects/:id", projectHandler.Get)
	api.PUT("/projects/:id", projectHandler.Update)
	api.DELETE("/projects/:id", projectHandler.Delete)

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
