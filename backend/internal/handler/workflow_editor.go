package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/uraguchihiroki/project_management_tool/internal/service"
)

type WorkflowEditorHandler struct {
	editor          service.WorkflowEditorService
	workflowService service.WorkflowService
}

func NewWorkflowEditorHandler(editor service.WorkflowEditorService, workflowService service.WorkflowService) *WorkflowEditorHandler {
	return &WorkflowEditorHandler{
		editor:          editor,
		workflowService: workflowService,
	}
}

// PUT /api/v1/workflows/:id/editor
func (h *WorkflowEditorHandler) Put(c echo.Context) error {
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

	var req service.WorkflowEditorSaveInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.editor.Save(uint(id), &req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}
