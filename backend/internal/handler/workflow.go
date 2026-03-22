package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/uraguchihiroki/project_management_tool/internal/service"
)

type WorkflowHandler struct {
	workflowService service.WorkflowService
}

func NewWorkflowHandler(workflowService service.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{workflowService: workflowService}
}

// GET /api/v1/workflows
func (h *WorkflowHandler) List(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	workflows, err := h.workflowService.ListAll()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if !isSuperAdmin && orgScope != nil {
		filtered := make([]interface{}, 0, len(workflows))
		for _, wf := range workflows {
			if wf.OrganizationID == *orgScope {
				filtered = append(filtered, wf)
			}
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"data": filtered})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": workflows})
}

// POST /api/v1/workflows
func (h *WorkflowHandler) Create(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	type Request struct {
		OrganizationID string `json:"organization_id"`
		Name           string `json:"name"`
		Description    string `json:"description"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.OrganizationID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "organization_id is required")
	}
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	var orgID uuid.UUID
	if isSuperAdmin {
		parsed, err := uuid.Parse(req.OrganizationID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid organization_id")
		}
		orgID = parsed
	} else {
		if orgScope == nil {
			return echo.NewHTTPError(http.StatusForbidden, "organization scope is missing")
		}
		orgID = *orgScope
	}
	workflow, err := h.workflowService.CreateWorkflow(orgID, req.Name, req.Description)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": workflow})
}

// GET /api/v1/workflows/:id
func (h *WorkflowHandler) Get(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workflow id")
	}
	workflow, err := h.workflowService.GetWorkflow(uint(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "workflow not found")
	}
	if !isSuperAdmin && (orgScope == nil || workflow.OrganizationID != *orgScope) {
		return echo.NewHTTPError(http.StatusNotFound, "workflow not found")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": workflow})
}

// PUT /api/v1/workflows/:id
func (h *WorkflowHandler) Update(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workflow id")
	}
	current, err := h.workflowService.GetWorkflow(uint(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "workflow not found")
	}
	if !isSuperAdmin && (orgScope == nil || current.OrganizationID != *orgScope) {
		return echo.NewHTTPError(http.StatusNotFound, "workflow not found")
	}
	type Request struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	workflow, err := h.workflowService.UpdateWorkflow(uint(id), req.Name, req.Description)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": workflow})
}

// PUT /api/v1/workflows/reorder
func (h *WorkflowHandler) Reorder(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	type Request struct {
		IDs []uint `json:"ids"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if !isSuperAdmin && orgScope != nil {
		for _, id := range req.IDs {
			wf, err := h.workflowService.GetWorkflow(id)
			if err != nil || wf.OrganizationID != *orgScope {
				return echo.NewHTTPError(http.StatusForbidden, "forbidden workflow reorder")
			}
		}
	}
	if err := h.workflowService.Reorder(req.IDs); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// DELETE /api/v1/workflows/:id
func (h *WorkflowHandler) Delete(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workflow id")
	}
	if !isSuperAdmin {
		wf, err := h.workflowService.GetWorkflow(uint(id))
		if err != nil || orgScope == nil || wf.OrganizationID != *orgScope {
			return echo.NewHTTPError(http.StatusNotFound, "workflow not found")
		}
	}
	if err := h.workflowService.DeleteWorkflow(uint(id)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}
