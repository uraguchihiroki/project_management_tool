package main

import (
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/uraguchihiroki/project_management_tool/internal/handler"
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

	// AutoMigrate
	if err := db.AutoMigrate(
		&model.Role{},
		&model.User{},
		&model.Project{},
		&model.Status{},
		&model.Issue{},
		&model.Comment{},
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

	// Services
	userSvc := service.NewUserService(userRepo)
	projectSvc := service.NewProjectService(projectRepo, statusRepo)
	issueSvc := service.NewIssueService(issueRepo, projectRepo)
	commentSvc := service.NewCommentService(commentRepo)
	roleSvc := service.NewRoleService(roleRepo)

	// Handlers
	userHandler := handler.NewUserHandler(userSvc)
	projectHandler := handler.NewProjectHandler(projectSvc)
	issueHandler := handler.NewIssueHandler(issueSvc)
	commentHandler := handler.NewCommentHandler(commentSvc)
	roleHandler := handler.NewRoleHandler(roleSvc, userSvc)

	// Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	// Routes
	api := e.Group("/api/v1")

	// Users
	api.GET("/users", userHandler.List)
	api.POST("/users", userHandler.Create)
	api.GET("/users/:id", userHandler.Get)
	api.PUT("/users/:id/admin", userHandler.SetAdmin)
	api.GET("/users/:id/roles", roleHandler.GetUserRoles)
	api.PUT("/users/:id/roles", roleHandler.AssignRoles)

	// Roles
	api.GET("/roles", roleHandler.List)
	api.POST("/roles", roleHandler.Create)
	api.PUT("/roles/:id", roleHandler.Update)
	api.DELETE("/roles/:id", roleHandler.Delete)

	// Admin
	api.GET("/admin/users", userHandler.ListWithRoles)

	// Projects
	api.GET("/projects", projectHandler.List)
	api.POST("/projects", projectHandler.Create)
	api.GET("/projects/:id", projectHandler.Get)
	api.PUT("/projects/:id", projectHandler.Update)
	api.DELETE("/projects/:id", projectHandler.Delete)

	// Issues
	api.GET("/projects/:projectId/issues", issueHandler.List)
	api.POST("/projects/:projectId/issues", issueHandler.Create)
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
