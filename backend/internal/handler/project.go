package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/uraguchihiroki/project_management_tool/internal/service"
)

type ProjectHandler struct {
	projectService service.ProjectService
}

func NewProjectHandler(projectService service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

func (h *ProjectHandler) List(c echo.Context) error {
	projects, err := h.projectService.List()
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
	project, err := h.projectService.Get(id)
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
	project, err := h.projectService.Create(service.CreateProjectInput{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     ownerID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": project})
}

func (h *ProjectHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	type Request struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	project, err := h.projectService.Update(id, service.UpdateProjectInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "project not found")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": project})
}

func (h *ProjectHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	if err := h.projectService.Delete(id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "deleted"})
}
