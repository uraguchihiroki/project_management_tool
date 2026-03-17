package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/uraguchihiroki/project_management_tool/internal/service"
)

type RoleHandler struct {
	roleService service.RoleService
	userService service.UserService
}

func NewRoleHandler(roleService service.RoleService, userService service.UserService) *RoleHandler {
	return &RoleHandler{roleService: roleService, userService: userService}
}

// GET /api/v1/roles
func (h *RoleHandler) List(c echo.Context) error {
	var orgID *uuid.UUID
	if raw := c.QueryParam("org_id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid org_id")
		}
		orgID = &parsed
	}
	roles, err := h.roleService.ListRoles(orgID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": roles})
}

// POST /api/v1/roles
func (h *RoleHandler) Create(c echo.Context) error {
	type Request struct {
		Name           string `json:"name"`
		Level          int    `json:"level"`
		Description    string `json:"description"`
		OrganizationID string `json:"organization_id"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	var orgID *uuid.UUID
	if req.OrganizationID != "" {
		parsed, err := uuid.Parse(req.OrganizationID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid organization_id")
		}
		orgID = &parsed
	}
	role, err := h.roleService.CreateRole(req.Name, req.Level, req.Description, orgID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": role})
}

// PUT /api/v1/roles/:id
func (h *RoleHandler) Update(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid role id")
	}
	type Request struct {
		Name        string `json:"name"`
		Level       int    `json:"level"`
		Description string `json:"description"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	role, err := h.roleService.UpdateRole(uint(id), req.Name, req.Level, req.Description)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": role})
}

// DELETE /api/v1/roles/:id
func (h *RoleHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid role id")
	}
	if err := h.roleService.DeleteRole(uint(id)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// GET /api/v1/users/:id/roles
func (h *RoleHandler) GetUserRoles(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}
	roles, err := h.roleService.GetUserRoles(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": roles})
}

// PUT /api/v1/users/:id/roles
func (h *RoleHandler) AssignRoles(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}
	type Request struct {
		RoleIDs []uint `json:"role_ids"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.roleService.AssignRolesToUser(userID, req.RoleIDs); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "roles assigned"})
}
