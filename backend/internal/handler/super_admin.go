package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/uraguchihiroki/project_management_tool/internal/service"
)

type SuperAdminHandler struct {
	superAdminService service.SuperAdminService
	orgService        service.OrganizationService
}

func NewSuperAdminHandler(superAdminService service.SuperAdminService, orgService service.OrganizationService) *SuperAdminHandler {
	return &SuperAdminHandler{superAdminService: superAdminService, orgService: orgService}
}

// POST /api/v1/super-admin/login
func (h *SuperAdminHandler) Login(c echo.Context) error {
	type Request struct {
		Email string `json:"email"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email is required")
	}
	admin, err := h.superAdminService.FindByEmail(req.Email)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "メールアドレスが見つかりません")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": admin})
}

// GET /api/v1/super-admin/organizations
func (h *SuperAdminHandler) ListOrganizations(c echo.Context) error {
	orgs, err := h.orgService.List()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": orgs})
}

// POST /api/v1/super-admin/organizations
func (h *SuperAdminHandler) CreateOrganization(c echo.Context) error {
	type Request struct {
		Name       string `json:"name"`
		AdminEmail string `json:"admin_email"`
		AdminName  string `json:"admin_name"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	org, err := h.orgService.Create(req.Name, req.AdminEmail, req.AdminName)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": org})
}
