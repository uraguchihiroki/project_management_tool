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

// GET /api/v1/organizations/:orgId/workflows
func (h *WorkflowHandler) ListByOrganization(c echo.Context) error {
	orgID, err := uuid.Parse(c.Param("orgId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid org id")
	}
	workflows, err := h.workflowService.ListByOrganization(orgID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": workflows})
}

// POST /api/v1/workflows
func (h *WorkflowHandler) Create(c echo.Context) error {
	type Request struct {
		OrganizationID string `json:"organization_id"`
		Name           string `json:"name"`
		Description    string `json:"description"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid organization_id")
	}
	workflow, err := h.workflowService.CreateWorkflow(orgID, req.Name, req.Description)
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

// PUT /api/v1/organizations/:orgId/workflows/reorder
func (h *WorkflowHandler) Reorder(c echo.Context) error {
	orgID, err := uuid.Parse(c.Param("orgId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid org id")
	}
	type Request struct {
		IDs []uint `json:"ids"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.workflowService.Reorder(orgID, req.IDs); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// PUT /api/v1/workflows/:id/steps/reorder
func (h *WorkflowHandler) ReorderSteps(c echo.Context) error {
	workflowID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workflow id")
	}
	type Request struct {
		IDs []uint `json:"ids"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.workflowService.ReorderSteps(uint(workflowID), req.IDs); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
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
		Name            string  `json:"name"`
		RequiredLevel   int     `json:"required_level"`
		StatusID        *string `json:"status_id"`
		ApproverType    string  `json:"approver_type"`
		ApproverUserID  *string `json:"approver_user_id"`
		MinApprovers    int     `json:"min_approvers"`
		ExcludeReporter bool    `json:"exclude_reporter"`
		ExcludeAssignee bool    `json:"exclude_assignee"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	if req.RequiredLevel < 0 || req.RequiredLevel > 9999 {
		return echo.NewHTTPError(http.StatusBadRequest, "必要レベルは0～9999の範囲で指定してください")
	}
	if req.MinApprovers < 0 || req.MinApprovers > 9999 {
		return echo.NewHTTPError(http.StatusBadRequest, "最小承認人数は0～9999の範囲で指定してください")
	}
	var statusID *uuid.UUID
	if req.StatusID != nil && *req.StatusID != "" {
		parsed, err := uuid.Parse(*req.StatusID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid status_id")
		}
		statusID = &parsed
	}
	var approverUserID *uuid.UUID
	if req.ApproverUserID != nil && *req.ApproverUserID != "" {
		parsed, err := uuid.Parse(*req.ApproverUserID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid approver_user_id")
		}
		approverUserID = &parsed
	}
	input := service.AddStepInput{
		Name:             req.Name,
		RequiredLevel:    req.RequiredLevel,
		StatusID:         statusID,
		ApproverType:     req.ApproverType,
		ApproverUserID:   approverUserID,
		MinApprovers:     req.MinApprovers,
		ExcludeReporter:  req.ExcludeReporter,
		ExcludeAssignee:  req.ExcludeAssignee,
	}
	step, err := h.workflowService.AddStep(uint(workflowID), input)
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
		Name            string  `json:"name"`
		RequiredLevel   int     `json:"required_level"`
		StatusID        *string `json:"status_id"`
		ApproverType    string  `json:"approver_type"`
		ApproverUserID  *string `json:"approver_user_id"`
		MinApprovers    int     `json:"min_approvers"`
		ExcludeReporter bool    `json:"exclude_reporter"`
		ExcludeAssignee bool    `json:"exclude_assignee"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	if req.RequiredLevel < 0 || req.RequiredLevel > 9999 {
		return echo.NewHTTPError(http.StatusBadRequest, "必要レベルは0～9999の範囲で指定してください")
	}
	if req.MinApprovers < 0 || req.MinApprovers > 9999 {
		return echo.NewHTTPError(http.StatusBadRequest, "最小承認人数は0～9999の範囲で指定してください")
	}
	var statusID *uuid.UUID
	if req.StatusID != nil && *req.StatusID != "" {
		parsed, err := uuid.Parse(*req.StatusID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid status_id")
		}
		statusID = &parsed
	}
	var approverUserID *uuid.UUID
	if req.ApproverUserID != nil && *req.ApproverUserID != "" {
		parsed, err := uuid.Parse(*req.ApproverUserID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid approver_user_id")
		}
		approverUserID = &parsed
	}
	input := service.UpdateStepInput{
		Name:             req.Name,
		RequiredLevel:   req.RequiredLevel,
		StatusID:        statusID,
		ApproverType:    req.ApproverType,
		ApproverUserID:  approverUserID,
		MinApprovers:    req.MinApprovers,
		ExcludeReporter: req.ExcludeReporter,
		ExcludeAssignee: req.ExcludeAssignee,
	}
	step, err := h.workflowService.UpdateStep(uint(stepID), input)
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
