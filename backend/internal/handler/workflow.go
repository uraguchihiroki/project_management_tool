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
	orgID, err := uuid.Parse(req.OrganizationID)
	if !isSuperAdmin && (orgScope == nil || orgID != *orgScope) {
		return echo.NewHTTPError(http.StatusForbidden, "forbidden for this organization")
	}
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
	current, err := h.workflowService.GetWorkflow(uint(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "workflow not found")
	}
	if !isSuperAdmin && (orgScope == nil || current.OrganizationID != *orgScope) {
		return echo.NewHTTPError(http.StatusNotFound, "workflow not found")
	}
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

// PUT /api/v1/workflows/:id/steps/reorder
func (h *WorkflowHandler) ReorderSteps(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	workflowID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workflow id")
	}
	if !isSuperAdmin {
		wf, err := h.workflowService.GetWorkflow(uint(workflowID))
		if err != nil || orgScope == nil || wf.OrganizationID != *orgScope {
			return echo.NewHTTPError(http.StatusNotFound, "workflow not found")
		}
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
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if !isSuperAdmin {
		wf, err := h.workflowService.GetWorkflow(uint(id))
		if err != nil || orgScope == nil || wf.OrganizationID != *orgScope {
			return echo.NewHTTPError(http.StatusNotFound, "workflow not found")
		}
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workflow id")
	}
	if err := h.workflowService.DeleteWorkflow(uint(id)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// GET /api/v1/workflows/:id/steps/:stepId
func (h *WorkflowHandler) GetStep(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	stepID, err := strconv.ParseUint(c.Param("stepId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid step id")
	}
	step, err := h.workflowService.GetStep(uint(stepID))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "step not found")
	}
	if !isSuperAdmin && (orgScope == nil || step.OrganizationID != *orgScope) {
		return echo.NewHTTPError(http.StatusNotFound, "step not found")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": step})
}

type approvalObjectReq struct {
	Type            string  `json:"type"`
	RoleID          *uint   `json:"role_id"`
	RoleOperator    string  `json:"role_operator"`
	UserID          *string `json:"user_id"`
	Points          int     `json:"points"`
	ExcludeReporter bool    `json:"exclude_reporter"`
	ExcludeAssignee bool    `json:"exclude_assignee"`
}

// POST /api/v1/workflows/:id/steps
func (h *WorkflowHandler) AddStep(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	workflowID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if !isSuperAdmin {
		wf, err := h.workflowService.GetWorkflow(uint(workflowID))
		if err != nil || orgScope == nil || wf.OrganizationID != *orgScope {
			return echo.NewHTTPError(http.StatusNotFound, "workflow not found")
		}
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workflow id")
	}
	type Request struct {
		StatusID        string             `json:"status_id"` // このステップのステータス（必須）
		NextStatusID    *string            `json:"next_status_id"`
		Description     string             `json:"description"`
		Threshold       int                `json:"threshold"`
		ApprovalObjects []approvalObjectReq `json:"approval_objects"`
		ExcludeReporter bool               `json:"exclude_reporter"`
		ExcludeAssignee bool               `json:"exclude_assignee"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.StatusID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "status_id is required")
	}
	statusID, err := uuid.Parse(req.StatusID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid status_id")
	}
	if req.Threshold < 0 || req.Threshold > 99999 {
		return echo.NewHTTPError(http.StatusBadRequest, "閾値は0～99999の範囲で指定してください")
	}
	var nextStatusID *uuid.UUID
	if req.NextStatusID != nil && *req.NextStatusID != "" {
		parsed, err := uuid.Parse(*req.NextStatusID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid next_status_id")
		}
		nextStatusID = &parsed
	}
	aos := make([]service.ApprovalObjectInput, 0, len(req.ApprovalObjects))
	for _, ao := range req.ApprovalObjects {
		var userID *uuid.UUID
		if ao.UserID != nil && *ao.UserID != "" {
			parsed, err := uuid.Parse(*ao.UserID)
			if err != nil {
				continue
			}
			userID = &parsed
		}
		points := ao.Points
		if points < 1 {
			points = 1
		}
		aos = append(aos, service.ApprovalObjectInput{
			Type:            ao.Type,
			RoleID:          ao.RoleID,
			RoleOperator:    ao.RoleOperator,
			UserID:          userID,
			Points:          points,
			ExcludeReporter: ao.ExcludeReporter,
			ExcludeAssignee: ao.ExcludeAssignee,
		})
	}
	input := service.AddStepInput{
		StatusID:        statusID,
		NextStatusID:    nextStatusID,
		Description:     req.Description,
		Threshold:       req.Threshold,
		ApprovalObjects: aos,
		ExcludeReporter: req.ExcludeReporter,
		ExcludeAssignee: req.ExcludeAssignee,
	}
	step, err := h.workflowService.AddStep(uint(workflowID), input)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": step})
}

// PUT /api/v1/workflows/:id/steps/:stepId
func (h *WorkflowHandler) UpdateStep(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	stepID, err := strconv.ParseUint(c.Param("stepId"), 10, 64)
	if !isSuperAdmin {
		current, err := h.workflowService.GetStep(uint(stepID))
		if err != nil || orgScope == nil || current.OrganizationID != *orgScope {
			return echo.NewHTTPError(http.StatusNotFound, "step not found")
		}
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid step id")
	}
	type Request struct {
		StatusID        *string            `json:"status_id"`
		NextStatusID    *string            `json:"next_status_id"`
		Description     string             `json:"description"`
		Threshold       int                `json:"threshold"`
		ApprovalObjects []approvalObjectReq `json:"approval_objects"`
		ExcludeReporter bool               `json:"exclude_reporter"`
		ExcludeAssignee bool               `json:"exclude_assignee"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Threshold < 0 || req.Threshold > 99999 {
		return echo.NewHTTPError(http.StatusBadRequest, "閾値は0～99999の範囲で指定してください")
	}
	var statusID, nextStatusID *uuid.UUID
	if req.StatusID != nil && *req.StatusID != "" {
		parsed, err := uuid.Parse(*req.StatusID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid status_id")
		}
		statusID = &parsed
	}
	if req.NextStatusID != nil && *req.NextStatusID != "" {
		parsed, err := uuid.Parse(*req.NextStatusID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid next_status_id")
		}
		nextStatusID = &parsed
	}
	var aos []service.ApprovalObjectInput
	if req.ApprovalObjects != nil {
		aos = make([]service.ApprovalObjectInput, 0, len(req.ApprovalObjects))
		for _, ao := range req.ApprovalObjects {
			var userID *uuid.UUID
			if ao.UserID != nil && *ao.UserID != "" {
				parsed, err := uuid.Parse(*ao.UserID)
				if err != nil {
					continue
				}
				userID = &parsed
			}
			points := ao.Points
			if points < 1 {
				points = 1
			}
			aos = append(aos, service.ApprovalObjectInput{
				Type:            ao.Type,
				RoleID:          ao.RoleID,
				RoleOperator:    ao.RoleOperator,
				UserID:          userID,
				Points:          points,
				ExcludeReporter: ao.ExcludeReporter,
				ExcludeAssignee: ao.ExcludeAssignee,
			})
		}
	}
	input := service.UpdateStepInput{
		StatusID:        statusID,
		NextStatusID:    nextStatusID,
		Description:     req.Description,
		Threshold:       req.Threshold,
		ApprovalObjects: aos,
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
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	stepID, err := strconv.ParseUint(c.Param("stepId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid step id")
	}
	if !isSuperAdmin {
		current, err := h.workflowService.GetStep(uint(stepID))
		if err != nil || orgScope == nil || current.OrganizationID != *orgScope {
			return echo.NewHTTPError(http.StatusNotFound, "step not found")
		}
	}
	if err := h.workflowService.DeleteStep(uint(stepID)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}
