package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/uraguchihiroki/project_management_tool/internal/auth"
)

const claimsContextKey = "auth_claims"

func GetClaims(c echo.Context) (*auth.Claims, bool) {
	v := c.Get(claimsContextKey)
	claims, ok := v.(*auth.Claims)
	return claims, ok && claims != nil
}

func OptionalJWT(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		header := c.Request().Header.Get("Authorization")
		if header == "" {
			return next(c)
		}
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header")
		}
		token := strings.TrimPrefix(header, prefix)
		claims, err := auth.ParseToken(token)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}
		c.Set(claimsContextKey, claims)
		return next(c)
	}
}

func RequireJWT(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		header := c.Request().Header.Get("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
		}
		token := strings.TrimPrefix(header, prefix)
		claims, err := auth.ParseToken(token)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}
		c.Set(claimsContextKey, claims)
		return next(c)
	}
}
