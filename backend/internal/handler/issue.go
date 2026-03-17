package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/uraguchihiroki/project_management_tool/internal/service"
)

type IssueHandler struct {
	issueService    service.IssueService
	approvalService service.ApprovalService
}

func NewIssueHandler(issueService service.IssueService, approvalService service.ApprovalService) *IssueHandler {
	return &IssueHandler{issueService: issueService, approvalService: approvalService}
}

func (h *IssueHandler) List(c echo.Context) error {
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	issues, err := h.issueService.List(projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": issues})
}

func (h *IssueHandler) Get(c echo.Context) error {
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	number, err := strconv.Atoi(c.Param("number"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid issue number")
	}
	issue, err := h.issueService.Get(projectID, number)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "issue not found")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": issue})
}

func (h *IssueHandler) Create(c echo.Context) error {
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	type Request struct {
		Title       string  `json:"title" validate:"required"`
		Description *string `json:"description"`
		StatusID    string  `json:"status_id" validate:"required,uuid"`
		Priority    string  `json:"priority"`
		AssigneeID  *string `json:"assignee_id"`
		ReporterID  string  `json:"reporter_id" validate:"required,uuid"`
		TemplateID  *uint   `json:"template_id"`
		WorkflowID  *uint   `json:"workflow_id"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	statusID, err := uuid.Parse(req.StatusID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid status_id")
	}
	reporterID, err := uuid.Parse(req.ReporterID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid reporter_id")
	}
	input := service.CreateIssueInput{
		Title:       req.Title,
		Description: req.Description,
		StatusID:    statusID,
		Priority:    req.Priority,
		ReporterID:  reporterID,
		TemplateID:  req.TemplateID,
		WorkflowID:  req.WorkflowID,
	}
	if req.AssigneeID != nil {
		aid, err := uuid.Parse(*req.AssigneeID)
		if err == nil {
			input.AssigneeID = &aid
		}
	}
	issue, err := h.issueService.Create(projectID, input)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	// ワークフローが紐付いている場合は承認レコードを自動生成
	if issue.WorkflowID != nil {
		if err := h.approvalService.InitializeForIssue(issue.ID, *issue.WorkflowID); err != nil {
			log.Printf("failed to initialize approvals for issue %s: %v", issue.ID, err)
		}
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": issue})
}

func (h *IssueHandler) Update(c echo.Context) error {
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	number, err := strconv.Atoi(c.Param("number"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid issue number")
	}
	type Request struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		StatusID    *string `json:"status_id"`
		Priority    *string `json:"priority"`
		AssigneeID  *string `json:"assignee_id"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	input := service.UpdateIssueInput{
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
	}
	if req.StatusID != nil {
		sid, err := uuid.Parse(*req.StatusID)
		if err == nil {
			input.StatusID = &sid
		}
	}
	if req.AssigneeID != nil {
		aid, err := uuid.Parse(*req.AssigneeID)
		if err == nil {
			input.AssigneeID = &aid
		}
	}
	issue, err := h.issueService.Update(projectID, number, input)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "issue not found")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": issue})
}

func (h *IssueHandler) Delete(c echo.Context) error {
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	number, err := strconv.Atoi(c.Param("number"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid issue number")
	}
	if err := h.issueService.Delete(projectID, number); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "deleted"})
}
