package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/uraguchihiroki/project_management_tool/internal/service"
)

type TemplateHandler struct {
	templateService service.TemplateService
}

func NewTemplateHandler(templateService service.TemplateService) *TemplateHandler {
	return &TemplateHandler{templateService: templateService}
}

// GET /api/v1/templates
func (h *TemplateHandler) List(c echo.Context) error {
	templates, err := h.templateService.ListAll()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": templates})
}

// GET /api/v1/projects/:projectId/templates
func (h *TemplateHandler) ListByProject(c echo.Context) error {
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	templates, err := h.templateService.ListByProject(projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": templates})
}

// POST /api/v1/templates
func (h *TemplateHandler) Create(c echo.Context) error {
	type Request struct {
		ProjectID       string `json:"project_id"`
		Name            string `json:"name"`
		Description     string `json:"description"`
		Body            string `json:"body"`
		DefaultPriority string `json:"default_priority"`
		WorkflowID      *uint  `json:"workflow_id"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project_id")
	}
	tmpl, err := h.templateService.CreateTemplate(projectID, req.Name, req.Description, req.Body, req.DefaultPriority, req.WorkflowID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": tmpl})
}

// GET /api/v1/templates/:id
func (h *TemplateHandler) Get(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid template id")
	}
	tmpl, err := h.templateService.GetTemplate(uint(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "template not found")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": tmpl})
}

// PUT /api/v1/templates/:id
func (h *TemplateHandler) Update(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid template id")
	}
	type Request struct {
		Name            string `json:"name"`
		Description     string `json:"description"`
		Body            string `json:"body"`
		DefaultPriority string `json:"default_priority"`
		WorkflowID      *uint  `json:"workflow_id"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	tmpl, err := h.templateService.UpdateTemplate(uint(id), req.Name, req.Description, req.Body, req.DefaultPriority, req.WorkflowID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": tmpl})
}

// PUT /api/v1/projects/:projectId/templates/reorder
func (h *TemplateHandler) Reorder(c echo.Context) error {
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	type Request struct {
		IDs []uint `json:"ids"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.templateService.Reorder(projectID, req.IDs); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// DELETE /api/v1/templates/:id
func (h *TemplateHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid template id")
	}
	if err := h.templateService.DeleteTemplate(uint(id)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}
