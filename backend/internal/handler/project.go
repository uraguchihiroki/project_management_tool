package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

type ProjectHandler struct {
	projectRepo repository.ProjectRepository
	statusRepo  repository.StatusRepository
}

func NewProjectHandler(projectRepo repository.ProjectRepository, statusRepo repository.StatusRepository) *ProjectHandler {
	return &ProjectHandler{projectRepo: projectRepo, statusRepo: statusRepo}
}

func (h *ProjectHandler) List(c echo.Context) error {
	projects, err := h.projectRepo.FindAll()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": projects})
}

func (h *ProjectHandler) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	project, err := h.projectRepo.FindByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "project not found")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": project})
}

func (h *ProjectHandler) Create(c echo.Context) error {
	type Request struct {
		Key         string  `json:"key" validate:"required,max=10"`
		Name        string  `json:"name" validate:"required,max=200"`
		Description *string `json:"description"`
		OwnerID     string  `json:"owner_id" validate:"required,uuid"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid owner_id")
	}
	project := &model.Project{
		ID:          uuid.New(),
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     ownerID,
		CreatedAt:   time.Now(),
	}
	if err := h.projectRepo.Create(project); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	// デフォルトステータスを作成
	defaultStatuses := []struct {
		Name  string
		Color string
		Order int
	}{
		{"未着手", "#6B7280", 1},
		{"進行中", "#3B82F6", 2},
		{"レビュー中", "#F59E0B", 3},
		{"完了", "#10B981", 4},
	}
	for _, s := range defaultStatuses {
		status := &model.Status{
			ID:        uuid.New(),
			ProjectID: project.ID,
			Name:      s.Name,
			Color:     s.Color,
			Order:     s.Order,
		}
		h.statusRepo.Create(status)
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": project})
}

func (h *ProjectHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	project, err := h.projectRepo.FindByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "project not found")
	}
	type Request struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = req.Description
	}
	if err := h.projectRepo.Update(project); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": project})
}

func (h *ProjectHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	if err := h.projectRepo.Delete(id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "deleted"})
}
