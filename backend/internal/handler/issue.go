package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/uraguchihiroki/project_management_tool/internal/service"
	"gorm.io/gorm"
)

type IssueHandler struct {
	issueService   service.IssueService
	projectService service.ProjectService
}

func NewIssueHandler(issueService service.IssueService, projectService service.ProjectService) *IssueHandler {
	return &IssueHandler{issueService: issueService, projectService: projectService}
}

func (h *IssueHandler) List(c echo.Context) error {
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
	var groupID *uuid.UUID
	if g := c.QueryParam("group_id"); g != "" {
		gid, err := uuid.Parse(g)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid group_id")
		}
		groupID = &gid
	}
	issues, err := h.issueService.List(projectID, groupID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": issues})
}

func (h *IssueHandler) Get(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	number, err := strconv.Atoi(c.Param("number"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid issue number")
	}
	issue, groups, err := h.issueService.GetWithGroups(projectID, number)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "issue not found")
	}
	if !isSuperAdmin && (orgScope == nil || issue.OrganizationID != *orgScope) {
		return echo.NewHTTPError(http.StatusNotFound, "issue not found")
	}
	issue.Groups = groups
	return c.JSON(http.StatusOK, map[string]interface{}{"data": issue})
}

func (h *IssueHandler) Create(c echo.Context) error {
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
		Title       string   `json:"title" validate:"required"`
		Description *string  `json:"description"`
		StatusID    string   `json:"status_id" validate:"required,uuid"`
		Priority    string   `json:"priority"`
		AssigneeID  *string  `json:"assignee_id"`
		ReporterID  string   `json:"reporter_id" validate:"required,uuid"`
		TemplateID  *uint    `json:"template_id"`
		GroupIDs    []string `json:"group_ids"`
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
	}
	if req.AssigneeID != nil {
		aid, err := uuid.Parse(*req.AssigneeID)
		if err == nil {
			input.AssigneeID = &aid
		}
	}
	for _, gs := range req.GroupIDs {
		gid, err := uuid.Parse(gs)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid group_ids")
		}
		input.GroupIDs = append(input.GroupIDs, gid)
	}
	issue, err := h.issueService.Create(projectID, input)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": issue})
}

func (h *IssueHandler) Update(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	if !isSuperAdmin {
		project, err := h.projectService.Get(projectID)
		if err != nil || orgScope == nil || project.OrganizationID != *orgScope {
			return echo.NewHTTPError(http.StatusNotFound, "project not found")
		}
	}
	number, err := strconv.Atoi(c.Param("number"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid issue number")
	}
	type Request struct {
		Title       *string   `json:"title"`
		Description *string   `json:"description"`
		StatusID    *string   `json:"status_id"`
		Priority    *string   `json:"priority"`
		AssigneeID  *string   `json:"assignee_id"`
		GroupIDs    *[]string `json:"group_ids"`
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
	if req.GroupIDs != nil {
		ids := make([]uuid.UUID, 0, len(*req.GroupIDs))
		for _, gs := range *req.GroupIDs {
			gid, err := uuid.Parse(gs)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid group_ids")
			}
			ids = append(ids, gid)
		}
		input.GroupIDs = &ids
	}
	actorID, err := actorIDFromClaims(c)
	if err != nil {
		return err
	}
	issue, err := h.issueService.Update(projectID, number, input, actorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "issue not found")
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": issue})
}

func (h *IssueHandler) Delete(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	if !isSuperAdmin {
		project, err := h.projectService.Get(projectID)
		if err != nil || orgScope == nil || project.OrganizationID != *orgScope {
			return echo.NewHTTPError(http.StatusNotFound, "project not found")
		}
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

// GET /api/v1/organizations/:orgId/issues
func (h *IssueHandler) ListByOrg(c echo.Context) error {
	orgID, _, authErr := requireOrgParam(c, "orgId")
	if authErr != nil {
		return authErr
	}
	var groupID *uuid.UUID
	if g := c.QueryParam("group_id"); g != "" {
		gid, err := uuid.Parse(g)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid group_id")
		}
		groupID = &gid
	}
	issues, err := h.issueService.ListByOrg(orgID, groupID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": issues})
}

// POST /api/v1/organizations/:orgId/issues
func (h *IssueHandler) CreateForOrg(c echo.Context) error {
	orgID, _, authErr := requireOrgParam(c, "orgId")
	if authErr != nil {
		return authErr
	}
	type Request struct {
		Title       string   `json:"title" validate:"required"`
		Description *string  `json:"description"`
		StatusID    string   `json:"status_id" validate:"required,uuid"`
		Priority    string   `json:"priority"`
		AssigneeID  *string  `json:"assignee_id"`
		ReporterID  string   `json:"reporter_id" validate:"required,uuid"`
		TemplateID  *uint    `json:"template_id"`
		GroupIDs    []string `json:"group_ids"`
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
	}
	if req.AssigneeID != nil {
		aid, err := uuid.Parse(*req.AssigneeID)
		if err == nil {
			input.AssigneeID = &aid
		}
	}
	for _, gs := range req.GroupIDs {
		gid, err := uuid.Parse(gs)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid group_ids")
		}
		input.GroupIDs = append(input.GroupIDs, gid)
	}
	issue, err := h.issueService.CreateForOrg(orgID, input)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": issue})
}

// GET /api/v1/organizations/:orgId/issues/:number
func (h *IssueHandler) GetByOrgAndNumber(c echo.Context) error {
	orgID, _, authErr := requireOrgParam(c, "orgId")
	if authErr != nil {
		return authErr
	}
	number, err := strconv.Atoi(c.Param("number"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid issue number")
	}
	issue, groups, err := h.issueService.GetByOrgAndNumberWithGroups(orgID, number)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "issue not found")
	}
	issue.Groups = groups
	return c.JSON(http.StatusOK, map[string]interface{}{"data": issue})
}

// PUT /api/v1/organizations/:orgId/issues/:number
func (h *IssueHandler) UpdateByOrgAndNumber(c echo.Context) error {
	orgID, _, authErr := requireOrgParam(c, "orgId")
	if authErr != nil {
		return authErr
	}
	number, err := strconv.Atoi(c.Param("number"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid issue number")
	}
	type Request struct {
		Title       *string   `json:"title"`
		Description *string   `json:"description"`
		StatusID    *string   `json:"status_id"`
		Priority    *string   `json:"priority"`
		AssigneeID  *string   `json:"assignee_id"`
		GroupIDs    *[]string `json:"group_ids"`
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
	if req.GroupIDs != nil {
		ids := make([]uuid.UUID, 0, len(*req.GroupIDs))
		for _, gs := range *req.GroupIDs {
			gid, err := uuid.Parse(gs)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid group_ids")
			}
			ids = append(ids, gid)
		}
		input.GroupIDs = &ids
	}
	actorID, err := actorIDFromClaims(c)
	if err != nil {
		return err
	}
	issue, err := h.issueService.UpdateByOrgAndNumber(orgID, number, input, actorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "issue not found")
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": issue})
}

// DELETE /api/v1/organizations/:orgId/issues/:number
func (h *IssueHandler) DeleteByOrgAndNumber(c echo.Context) error {
	orgID, _, authErr := requireOrgParam(c, "orgId")
	if authErr != nil {
		return authErr
	}
	number, err := strconv.Atoi(c.Param("number"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid issue number")
	}
	if err := h.issueService.DeleteByOrgAndNumber(orgID, number); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "deleted"})
}

// GET /projects/:projectId/issues/:number/groups
func (h *IssueHandler) ListIssueGroups(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
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
	if !isSuperAdmin && (orgScope == nil || issue.OrganizationID != *orgScope) {
		return echo.NewHTTPError(http.StatusNotFound, "issue not found")
	}
	_, groups, err := h.issueService.GetWithGroups(projectID, number)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": groups})
}

// PUT /projects/:projectId/issues/:number/groups
func (h *IssueHandler) PutIssueGroups(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	if !isSuperAdmin {
		project, err := h.projectService.Get(projectID)
		if err != nil || orgScope == nil || project.OrganizationID != *orgScope {
			return echo.NewHTTPError(http.StatusNotFound, "project not found")
		}
	}
	number, err := strconv.Atoi(c.Param("number"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid issue number")
	}
	var req struct {
		GroupIDs []string `json:"group_ids"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ids := make([]uuid.UUID, 0, len(req.GroupIDs))
	for _, gs := range req.GroupIDs {
		gid, err := uuid.Parse(gs)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid group_ids")
		}
		ids = append(ids, gid)
	}
	if err := h.issueService.SetIssueGroups(projectID, number, ids); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "ok"})
}
