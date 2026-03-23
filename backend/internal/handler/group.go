package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/uraguchihiroki/project_management_tool/internal/service"
)

type GroupHandler struct {
	groupService service.GroupService
}

func NewGroupHandler(groupService service.GroupService) *GroupHandler {
	return &GroupHandler{groupService: groupService}
}

// GET /organizations/:orgId/groups
func (h *GroupHandler) List(c echo.Context) error {
	orgID, _, authErr := requireOrgParam(c, "orgId")
	if authErr != nil {
		return authErr
	}
	var kind *string
	if k := c.QueryParam("kind"); k != "" {
		kind = &k
	}
	groups, err := h.groupService.List(orgID, kind)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": groups})
}

// POST /organizations/:orgId/groups
func (h *GroupHandler) Create(c echo.Context) error {
	orgID, _, authErr := requireOrgParam(c, "orgId")
	if authErr != nil {
		return authErr
	}
	var req struct {
		Name         string  `json:"name"`
		Kind         *string `json:"kind"`
		DisplayOrder *int    `json:"display_order"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	order := 0
	if req.DisplayOrder != nil {
		order = *req.DisplayOrder
	}
	g, err := h.groupService.Create(orgID, req.Name, req.Kind, order)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": g})
}

// GET /groups/:id
func (h *GroupHandler) Get(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	g, err := h.groupService.Get(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "group not found")
	}
	if !isSuperAdmin && (orgScope == nil || g.OrganizationID != *orgScope) {
		return echo.NewHTTPError(http.StatusNotFound, "group not found")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": g})
}

// PUT /groups/:id
func (h *GroupHandler) Update(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	g0, err := h.groupService.Get(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "group not found")
	}
	if !isSuperAdmin && (orgScope == nil || g0.OrganizationID != *orgScope) {
		return echo.NewHTTPError(http.StatusNotFound, "group not found")
	}
	var req struct {
		Name         string  `json:"name"`
		Kind         *string `json:"kind"`
		DisplayOrder *int    `json:"display_order"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	order := g0.DisplayOrder
	if req.DisplayOrder != nil {
		order = *req.DisplayOrder
	}
	g, err := h.groupService.Update(id, req.Name, req.Kind, order)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": g})
}

// DELETE /groups/:id
func (h *GroupHandler) Delete(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	g0, err := h.groupService.Get(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "group not found")
	}
	if !isSuperAdmin && (orgScope == nil || g0.OrganizationID != *orgScope) {
		return echo.NewHTTPError(http.StatusNotFound, "group not found")
	}
	if err := h.groupService.Delete(id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "deleted"})
}

// GET /groups/:id/members
func (h *GroupHandler) ListMembers(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	g0, err := h.groupService.Get(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "group not found")
	}
	if !isSuperAdmin && (orgScope == nil || g0.OrganizationID != *orgScope) {
		return echo.NewHTTPError(http.StatusNotFound, "group not found")
	}
	ids, err := h.groupService.ListMembers(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": ids})
}

// PUT /groups/:id/members — body: { "user_ids": ["uuid", ...] } 一括置換
func (h *GroupHandler) ReplaceMembers(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	g0, err := h.groupService.Get(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "group not found")
	}
	if !isSuperAdmin && (orgScope == nil || g0.OrganizationID != *orgScope) {
		return echo.NewHTTPError(http.StatusNotFound, "group not found")
	}
	var req struct {
		UserIDs []string `json:"user_ids"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	uids := make([]uuid.UUID, 0, len(req.UserIDs))
	for _, s := range req.UserIDs {
		uid, err := uuid.Parse(s)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid user_id")
		}
		uids = append(uids, uid)
	}
	if err := h.groupService.ReplaceMembers(id, uids); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "ok"})
}

// GET /users/:id/groups
func (h *GroupHandler) ListByUser(c echo.Context) error {
	_, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}
	actorID, err := actorIDFromClaims(c)
	if err != nil {
		return err
	}
	if !isSuperAdmin && userID != actorID {
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	}
	groups, err := h.groupService.ListGroupsByUser(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": groups})
}
