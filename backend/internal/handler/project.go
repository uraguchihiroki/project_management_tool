package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/uraguchihiroki/project_management_tool/internal/service"
)

type ProjectHandler struct {
	projectService service.ProjectService
}

func NewProjectHandler(projectService service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

func (h *ProjectHandler) List(c echo.Context) error {
	var orgID *uuid.UUID
	if raw := c.QueryParam("org_id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid org_id")
		}
		orgID = &parsed
	}
	projects, err := h.projectService.List(orgID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": projects})
}

func (h *ProjectHandler) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	project, err := h.projectService.Get(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "project not found")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": project})
}

func (h *ProjectHandler) Create(c echo.Context) error {
	type Request struct {
		Key            string  `json:"key" validate:"required,max=10"`
		Name           string  `json:"name" validate:"required,max=200"`
		Description    *string `json:"description"`
		OwnerID        string  `json:"owner_id" validate:"required,uuid"`
		OrganizationID string  `json:"organization_id"`
		StartDate      string  `json:"start_date"`
		EndDate        string  `json:"end_date"`
		Status         string  `json:"status"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid owner_id")
	}
	var orgID *uuid.UUID
	if req.OrganizationID != "" {
		parsed, err := uuid.Parse(req.OrganizationID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid organization_id")
		}
		orgID = &parsed
	}
	var startDate, endDate *time.Time
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			startDate = &t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			endDate = &t
		}
	}
	project, err := h.projectService.Create(service.CreateProjectInput{
		Key:            req.Key,
		Name:           req.Name,
		Description:    req.Description,
		OwnerID:        ownerID,
		OrganizationID: orgID,
		StartDate:      startDate,
		EndDate:        endDate,
		Status:         req.Status,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": project})
}

func (h *ProjectHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	type Request struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		StartDate   string  `json:"start_date"`
		EndDate     string  `json:"end_date"`
		Status      string  `json:"status"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	var startDate, endDate *time.Time
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			startDate = &t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			endDate = &t
		}
	}
	var status *string
	if req.Status != "" {
		status = &req.Status
	}
	project, err := h.projectService.Update(id, service.UpdateProjectInput{
		Name:        req.Name,
		Description: req.Description,
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      status,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "project not found")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": project})
}

func (h *ProjectHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	if err := h.projectService.Delete(id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "deleted"})
}

// GET /api/v1/organizations/:orgId/statuses
func (h *ProjectHandler) ListStatusesByOrg(c echo.Context) error {
	orgID, err := uuid.Parse(c.Param("orgId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid org id")
	}
	statuses, err := h.projectService.ListStatusesByOrg(orgID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": statuses})
}
