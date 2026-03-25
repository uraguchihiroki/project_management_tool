package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/uraguchihiroki/project_management_tool/internal/service"
)

type StatusHandler struct {
	statusService   service.StatusService
	workflowService service.WorkflowService
}

func NewStatusHandler(statusService service.StatusService, workflowService service.WorkflowService) *StatusHandler {
	return &StatusHandler{statusService: statusService, workflowService: workflowService}
}

func (h *StatusHandler) authorizeWorkflowAccess(c echo.Context) (uint, error) {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return 0, authErr
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid workflow id")
	}
	workflow, err := h.workflowService.GetWorkflow(uint(id))
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusNotFound, "workflow not found")
	}
	if !isSuperAdmin && (orgScope == nil || workflow.OrganizationID != *orgScope) {
		return 0, echo.NewHTTPError(http.StatusNotFound, "workflow not found")
	}
	return uint(id), nil
}

// GET /api/v1/workflows/:id/statuses
func (h *StatusHandler) ListByWorkflow(c echo.Context) error {
	wfID, err := h.authorizeWorkflowAccess(c)
	if err != nil {
		return err
	}
	statuses, err := h.statusService.ListByWorkflowID(wfID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": statuses})
}

// POST /api/v1/workflows/:id/statuses
func (h *StatusHandler) CreateForWorkflow(c echo.Context) error {
	wfID, err := h.authorizeWorkflowAccess(c)
	if err != nil {
		return err
	}
	type Request struct {
		Name          string `json:"name"`
		Color         string `json:"color"`
		DisplayOrder  int    `json:"display_order"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Color == "" {
		req.Color = "#6B7280"
	}
	status, err := h.statusService.CreateForWorkflow(wfID, req.Name, req.Color, req.DisplayOrder)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": status})
}

// POST /api/v1/organizations/:orgId/statuses
func (h *StatusHandler) Create(c echo.Context) error {
	orgID, _, authErr := requireOrgParam(c, "orgId")
	if authErr != nil {
		return authErr
	}
	type Request struct {
		Name         string `json:"name"`
		Color        string `json:"color"`
		DisplayOrder int    `json:"display_order"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Color == "" {
		req.Color = "#6B7280"
	}
	status, err := h.statusService.Create(orgID, req.Name, req.Color, req.DisplayOrder)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": status})
}

// PUT /api/v1/statuses/:id
func (h *StatusHandler) Update(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid status id")
	}
	current, err := h.statusService.Get(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "status not found")
	}
	if !isSuperAdmin {
		if current.Workflow.ID == 0 || orgScope == nil || current.Workflow.OrganizationID != *orgScope {
			return echo.NewHTTPError(http.StatusNotFound, "status not found")
		}
	}
	type Request struct {
		Name         string `json:"name"`
		Color        string `json:"color"`
		DisplayOrder int    `json:"display_order"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	status, err := h.statusService.Update(id, req.Name, req.Color, req.DisplayOrder)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": status})
}

// DELETE /api/v1/statuses/:id
func (h *StatusHandler) Delete(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid status id")
	}
	current, err := h.statusService.Get(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "status not found")
	}
	if !isSuperAdmin {
		if current.Workflow.ID == 0 || orgScope == nil || current.Workflow.OrganizationID != *orgScope {
			return echo.NewHTTPError(http.StatusNotFound, "status not found")
		}
	}
	if err := h.statusService.Delete(id); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// PUT /api/v1/workflows/:id/statuses/reorder
func (h *StatusHandler) ReorderForWorkflow(c echo.Context) error {
	wfID, err := h.authorizeWorkflowAccess(c)
	if err != nil {
		return err
	}
	type Request struct {
		StatusIDs []string `json:"status_ids"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ids := make([]uuid.UUID, 0, len(req.StatusIDs))
	for _, s := range req.StatusIDs {
		id, perr := uuid.Parse(s)
		if perr != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid status id")
		}
		ids = append(ids, id)
	}
	if err := h.statusService.ReorderForWorkflow(wfID, ids); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}
