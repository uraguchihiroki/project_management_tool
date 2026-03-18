package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/uraguchihiroki/project_management_tool/internal/service"
)

type ApprovalHandler struct {
	approvalService service.ApprovalService
}

func NewApprovalHandler(approvalService service.ApprovalService) *ApprovalHandler {
	return &ApprovalHandler{approvalService: approvalService}
}

// GET /api/v1/issues/:issueId/approvals
func (h *ApprovalHandler) List(c echo.Context) error {
	issueID, err := uuid.Parse(c.Param("issueId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid issue id")
	}
	approvals, err := h.approvalService.GetApprovals(issueID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": approvals})
}

// POST /api/v1/approvals/:id/approve
func (h *ApprovalHandler) Approve(c echo.Context) error {
	approvalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid approval id")
	}
	type Request struct {
		ApproverID string `json:"approver_id"`
		Comment    string `json:"comment"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	approverID, err := uuid.Parse(req.ApproverID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid approver_id")
	}
	approval, err := h.approvalService.Approve(approvalID, approverID, req.Comment)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": approval})
}

// POST /api/v1/issues/:issueId/steps/:stepId/approve
func (h *ApprovalHandler) ApproveStep(c echo.Context) error {
	issueID, err := uuid.Parse(c.Param("issueId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid issue id")
	}
	stepID, err := strconv.ParseUint(c.Param("stepId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid step id")
	}
	type Request struct {
		ApproverID string `json:"approver_id"`
		Comment    string `json:"comment"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	approverID, err := uuid.Parse(req.ApproverID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid approver_id")
	}
	approval, err := h.approvalService.ApproveStep(issueID, uint(stepID), approverID, req.Comment)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": approval})
}

// POST /api/v1/approvals/:id/reject
func (h *ApprovalHandler) Reject(c echo.Context) error {
	approvalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid approval id")
	}
	type Request struct {
		ApproverID string `json:"approver_id"`
		Comment    string `json:"comment"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	approverID, err := uuid.Parse(req.ApproverID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid approver_id")
	}
	approval, err := h.approvalService.Reject(approvalID, approverID, req.Comment)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": approval})
}
