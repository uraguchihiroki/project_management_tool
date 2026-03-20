package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	authmw "github.com/uraguchihiroki/project_management_tool/internal/middleware"
)

func requireClaims(c echo.Context) (*uuid.UUID, bool, error) {
	claims, ok := authmw.GetClaims(c)
	if !ok {
		return nil, false, echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	if claims.IsSuperAdmin {
		return nil, true, nil
	}
	if claims.OrganizationID == "" {
		return nil, false, echo.NewHTTPError(http.StatusForbidden, "organization scope is missing")
	}
	orgID, err := uuid.Parse(claims.OrganizationID)
	if err != nil {
		return nil, false, echo.NewHTTPError(http.StatusUnauthorized, "invalid organization scope")
	}
	return &orgID, false, nil
}

func requireOrgParam(c echo.Context, paramName string) (uuid.UUID, bool, error) {
	target, err := uuid.Parse(c.Param(paramName))
	if err != nil {
		return uuid.Nil, false, echo.NewHTTPError(http.StatusBadRequest, "invalid org id")
	}
	orgID, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return uuid.Nil, false, authErr
	}
	if !isSuperAdmin && (orgID == nil || *orgID != target) {
		return uuid.Nil, false, echo.NewHTTPError(http.StatusForbidden, "forbidden for this organization")
	}
	return target, isSuperAdmin, nil
}

func requireOrgQuery(c echo.Context, queryName string) (uuid.UUID, bool, error) {
	val := c.QueryParam(queryName)
	if val == "" {
		return uuid.Nil, false, echo.NewHTTPError(http.StatusBadRequest, queryName+" is required")
	}
	target, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, false, echo.NewHTTPError(http.StatusBadRequest, "invalid "+queryName)
	}
	orgID, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return uuid.Nil, false, authErr
	}
	if !isSuperAdmin && (orgID == nil || *orgID != target) {
		return uuid.Nil, false, echo.NewHTTPError(http.StatusForbidden, "forbidden for this organization")
	}
	return target, isSuperAdmin, nil
}
