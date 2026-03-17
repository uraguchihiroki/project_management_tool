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
	workflows, err := h.workflowService.ListAll()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": workflows})
}

// GET /api/v1/projects/:projectId/workflows
func (h *WorkflowHandler) ListByProject(c echo.Context) error {
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	workflows, err := h.workflowService.ListByProject(projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": workflows})
}

// POST /api/v1/workflows
func (h *WorkflowHandler) Create(c echo.Context) error {
	type Request struct {
		ProjectID   string `json:"project_id"`
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
	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project_id")
	}
	workflow, err := h.workflowService.CreateWorkflow(projectID, req.Name, req.Description)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": workflow})
}

// GET /api/v1/workflows/:id
func (h *WorkflowHandler) Get(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workflow id")
	}
	workflow, err := h.workflowService.GetWorkflow(uint(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "workflow not found")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": workflow})
}

// PUT /api/v1/workflows/:id
func (h *WorkflowHandler) Update(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workflow id")
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

// DELETE /api/v1/workflows/:id
func (h *WorkflowHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workflow id")
	}
	if err := h.workflowService.DeleteWorkflow(uint(id)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// POST /api/v1/workflows/:id/steps
func (h *WorkflowHandler) AddStep(c echo.Context) error {
	workflowID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workflow id")
	}
	type Request struct {
		Name          string  `json:"name"`
		RequiredLevel int     `json:"required_level"`
		StatusID      *string `json:"status_id"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	var statusID *uuid.UUID
	if req.StatusID != nil && *req.StatusID != "" {
		parsed, err := uuid.Parse(*req.StatusID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid status_id")
		}
		statusID = &parsed
	}
	step, err := h.workflowService.AddStep(uint(workflowID), req.Name, req.RequiredLevel, statusID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": step})
}

// PUT /api/v1/workflows/:id/steps/:stepId
func (h *WorkflowHandler) UpdateStep(c echo.Context) error {
	stepID, err := strconv.ParseUint(c.Param("stepId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid step id")
	}
	type Request struct {
		Name          string  `json:"name"`
		RequiredLevel int     `json:"required_level"`
		StatusID      *string `json:"status_id"`
		Order         int     `json:"order"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	var statusID *uuid.UUID
	if req.StatusID != nil && *req.StatusID != "" {
		parsed, err := uuid.Parse(*req.StatusID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid status_id")
		}
		statusID = &parsed
	}
	step, err := h.workflowService.UpdateStep(uint(stepID), req.Name, req.RequiredLevel, statusID, req.Order)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": step})
}

// DELETE /api/v1/workflows/:id/steps/:stepId
func (h *WorkflowHandler) DeleteStep(c echo.Context) error {
	stepID, err := strconv.ParseUint(c.Param("stepId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid step id")
	}
	if err := h.workflowService.DeleteStep(uint(stepID)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}
