package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/uraguchihiroki/project_management_tool/internal/service"
)

type StatusHandler struct {
	statusService service.StatusService
}

func NewStatusHandler(statusService service.StatusService) *StatusHandler {
	return &StatusHandler{statusService: statusService}
}

// POST /api/v1/organizations/:orgId/statuses
func (h *StatusHandler) Create(c echo.Context) error {
	orgID, _, authErr := requireOrgParam(c, "orgId")
	if authErr != nil {
		return authErr
	}
	type Request struct {
		Name        string `json:"name"`
		Color       string `json:"color"`
		Type        string `json:"type"`
		Order       int    `json:"order"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Color == "" {
		req.Color = "#6B7280"
	}
	status, err := h.statusService.Create(orgID, req.Name, req.Color, req.Type, req.Order)
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
		Name  string `json:"name"`
		Color string `json:"color"`
		Order int    `json:"order"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	status, err := h.statusService.Update(id, req.Name, req.Color, req.Order)
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
