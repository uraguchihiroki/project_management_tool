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
	projectService  service.ProjectService
}

func NewTemplateHandler(templateService service.TemplateService, projectService service.ProjectService) *TemplateHandler {
	return &TemplateHandler{templateService: templateService, projectService: projectService}
}

// GET /api/v1/templates
func (h *TemplateHandler) List(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	templates, err := h.templateService.ListAll()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if !isSuperAdmin && orgScope != nil {
		filtered := make([]interface{}, 0, len(templates))
		for _, t := range templates {
			project, err := h.projectService.Get(t.ProjectID)
			if err == nil && project.OrganizationID == *orgScope {
				filtered = append(filtered, t)
			}
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"data": filtered})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": templates})
}

// GET /api/v1/projects/:projectId/templates
func (h *TemplateHandler) ListByProject(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	projectID, err := uuid.Parse(c.Param("projectId"))
	if !isSuperAdmin {
		project, err := h.projectService.Get(projectID)
		if err != nil || orgScope == nil || project.OrganizationID != *orgScope {
			return echo.NewHTTPError(http.StatusNotFound, "project not found")
		}
	}
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
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
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
	if !isSuperAdmin {
		project, err := h.projectService.Get(projectID)
		if err != nil || orgScope == nil || project.OrganizationID != *orgScope {
			return echo.NewHTTPError(http.StatusForbidden, "forbidden for this organization")
		}
	}
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
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid template id")
	}
	tmpl, err := h.templateService.GetTemplate(uint(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "template not found")
	}
	if !isSuperAdmin {
		project, err := h.projectService.Get(tmpl.ProjectID)
		if err != nil || orgScope == nil || project.OrganizationID != *orgScope {
			return echo.NewHTTPError(http.StatusNotFound, "template not found")
		}
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": tmpl})
}

// PUT /api/v1/templates/:id
func (h *TemplateHandler) Update(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid template id")
	}
	existing, err := h.templateService.GetTemplate(uint(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "template not found")
	}
	if !isSuperAdmin {
		project, err := h.projectService.Get(existing.ProjectID)
		if err != nil || orgScope == nil || project.OrganizationID != *orgScope {
			return echo.NewHTTPError(http.StatusNotFound, "template not found")
		}
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
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	projectID, err := uuid.Parse(c.Param("projectId"))
	if !isSuperAdmin {
		project, err := h.projectService.Get(projectID)
		if err != nil || orgScope == nil || project.OrganizationID != *orgScope {
			return echo.NewHTTPError(http.StatusNotFound, "project not found")
		}
	}
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
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid template id")
	}
	existing, err := h.templateService.GetTemplate(uint(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "template not found")
	}
	if !isSuperAdmin {
		project, err := h.projectService.Get(existing.ProjectID)
		if err != nil || orgScope == nil || project.OrganizationID != *orgScope {
			return echo.NewHTTPError(http.StatusNotFound, "template not found")
		}
	}
	if err := h.templateService.DeleteTemplate(uint(id)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}
