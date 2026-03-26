package handler

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	authmw "github.com/uraguchihiroki/project_management_tool/internal/middleware"
	"github.com/uraguchihiroki/project_management_tool/internal/service"
)

type OrganizationHandler struct {
	orgService service.OrganizationService
}

func NewOrganizationHandler(orgService service.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{orgService: orgService}
}

// GET /api/v1/organizations
func (h *OrganizationHandler) List(c echo.Context) error {
	orgID, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	if !isSuperAdmin && orgID != nil {
		org, err := h.orgService.Get(*orgID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "organization not found")
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"data": []interface{}{org}})
	}
	orgs, err := h.orgService.List()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": orgs})
}

// POST /api/v1/organizations
func (h *OrganizationHandler) Create(c echo.Context) error {
	_, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	if !isSuperAdmin {
		return echo.NewHTTPError(http.StatusForbidden, "only super admin can create organizations")
	}
	type Request struct {
		Name string `json:"name"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	org, err := h.orgService.Create(req.Name, "", "")
	if err != nil {
		if errors.Is(err, service.ErrDuplicateOrganizationName) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": org})
}

// GET /api/v1/users/:id/organizations
func (h *OrganizationHandler) ListByUser(c echo.Context) error {
	_, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	claims, ok := authmw.GetClaims(c)
	if !ok || claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}
	claimsUserID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}
	if !isSuperAdmin && userID != claimsUserID {
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	}
	orgs, err := h.orgService.GetUserOrganizations(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": orgs})
}

// POST /api/v1/organizations/:orgId/users
func (h *OrganizationHandler) AddUser(c echo.Context) error {
	orgID, _, authErr := requireOrgParam(c, "orgId")
	if authErr != nil {
		return authErr
	}
	type Request struct {
		UserID     string `json:"user_id"`
		IsOrgAdmin bool   `json:"is_org_admin"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user_id")
	}
	user, err := h.orgService.AddUser(orgID, userID, req.IsOrgAdmin)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": user, "message": "user added to organization"})
}
