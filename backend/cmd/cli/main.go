// CLI エントリーポイント
//
// 使用例:
//
//	go run ./cmd/cli org seed --all
//	go run ./cmd/cli org seed --org-id=<uuid> [--owner-id=<uuid>]
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	appdb "github.com/uraguchihiroki/project_management_tool/internal/db"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
	"github.com/uraguchihiroki/project_management_tool/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	seedCmd := flag.NewFlagSet("org seed", flag.ExitOnError)
	seedAll := seedCmd.Bool("all", false, "全組織にseed投入")
	orgIDStr := seedCmd.String("org-id", "", "組織ID (UUID)")
	ownerIDStr := seedCmd.String("owner-id", "", "オーナーID (UUID)。指定時はサンプルプロジェクト・Issueも作成")

	if len(os.Args) < 2 || os.Args[1] != "org" {
		fmt.Fprintln(os.Stderr, "Usage: go run ./cmd/cli org seed --all")
		fmt.Fprintln(os.Stderr, "       go run ./cmd/cli org seed --org-id=<uuid> [--owner-id=<uuid>]")
		os.Exit(1)
	}
	if len(os.Args) < 3 || os.Args[2] != "seed" {
		fmt.Fprintln(os.Stderr, "Usage: go run ./cmd/cli org seed --all")
		fmt.Fprintln(os.Stderr, "       go run ./cmd/cli org seed --org-id=<uuid> [--owner-id=<uuid>]")
		os.Exit(1)
	}
	if err := seedCmd.Parse(os.Args[3:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if !*seedAll && *orgIDStr == "" {
		fmt.Fprintln(os.Stderr, "org-id or --all is required")
		os.Exit(1)
	}
	if *seedAll && *orgIDStr != "" {
		fmt.Fprintln(os.Stderr, "cannot use both --all and --org-id")
		os.Exit(1)
	}

	var orgID uuid.UUID
	var ownerID *uuid.UUID
	if !*seedAll {
		var err error
		orgID, err = uuid.Parse(*orgIDStr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid org-id:", err)
			os.Exit(1)
		}
		if *ownerIDStr != "" {
			oid, err := uuid.Parse(*ownerIDStr)
			if err != nil {
				fmt.Fprintln(os.Stderr, "invalid owner-id:", err)
				os.Exit(1)
			}
			ownerID = &oid
		}
	}

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

	if err := db.AutoMigrate(
		&model.Organization{},
		&model.SuperAdmin{},
		&model.Role{},
		&model.User{},
		&model.Department{},
		&model.OrganizationUserDepartment{},
		&model.Project{},
		&model.Status{},
		&model.WorkflowTransition{},
		&model.Issue{},
		&model.Comment{},
		&model.Workflow{},
		&model.IssueTemplate{},
		&model.IssueEvent{},
		&model.Group{},
		&model.UserGroup{},
		&model.IssueGroup{},
		&model.TransitionAlertRule{},
	); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	if err := appdb.MigrateStatusDedupeAndUniqueIndex(db); err != nil {
		log.Fatalf("failed to migrate status dedupe / unique index: %v", err)
	}

	orgRepo := repository.NewOrganizationRepository(db)
	statusRepo := repository.NewStatusRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	departmentRepo := repository.NewDepartmentRepository(db)
	issueRepo := repository.NewIssueRepository(db)
	workflowRepo := repository.NewWorkflowRepository(db)
	transitionRepo := repository.NewWorkflowTransitionRepository(db)
	orgSeedSvc := service.NewOrgSeedService(orgRepo, statusRepo, roleRepo, projectRepo, departmentRepo, issueRepo, workflowRepo, transitionRepo)

	var seedErr error
	if *seedAll {
		seedErr = orgSeedSvc.SeedAllOrganizations()
	} else {
		seedErr = orgSeedSvc.SeedNewOrganization(orgID, ownerID)
	}
	if seedErr != nil {
		log.Fatalf("seed failed: %v", seedErr)
	}

	fmt.Println("Seed completed successfully.")
}
