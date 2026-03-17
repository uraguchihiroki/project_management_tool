package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

type IssueHandler struct {
	issueRepo   repository.IssueRepository
	projectRepo repository.ProjectRepository
}

func NewIssueHandler(issueRepo repository.IssueRepository, projectRepo repository.ProjectRepository) *IssueHandler {
	return &IssueHandler{issueRepo: issueRepo, projectRepo: projectRepo}
}

func (h *IssueHandler) List(c echo.Context) error {
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	issues, err := h.issueRepo.FindByProject(projectID)
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
	issue, err := h.issueRepo.FindByNumber(projectID, number)
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
		DueDate     *string `json:"due_date"`
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
	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}
	nextNum, _ := h.issueRepo.NextNumber(projectID)
	issue := &model.Issue{
		ID:          uuid.New(),
		Number:      nextNum,
		Title:       req.Title,
		Description: req.Description,
		StatusID:    statusID,
		Priority:    priority,
		ReporterID:  reporterID,
		ProjectID:   projectID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if req.AssigneeID != nil {
		aid, err := uuid.Parse(*req.AssigneeID)
		if err == nil {
			issue.AssigneeID = &aid
		}
	}
	if err := h.issueRepo.Create(issue); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
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
	issue, err := h.issueRepo.FindByNumber(projectID, number)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "issue not found")
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
	if req.Title != nil {
		issue.Title = *req.Title
	}
	if req.Description != nil {
		issue.Description = req.Description
	}
	if req.StatusID != nil {
		sid, err := uuid.Parse(*req.StatusID)
		if err == nil {
			issue.StatusID = sid
		}
	}
	if req.Priority != nil {
		issue.Priority = *req.Priority
	}
	if req.AssigneeID != nil {
		aid, err := uuid.Parse(*req.AssigneeID)
		if err == nil {
			issue.AssigneeID = &aid
		}
	}
	issue.UpdatedAt = time.Now()
	if err := h.issueRepo.Update(issue); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
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
	issue, err := h.issueRepo.FindByNumber(projectID, number)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "issue not found")
	}
	if err := h.issueRepo.Delete(issue.ID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "deleted"})
}
