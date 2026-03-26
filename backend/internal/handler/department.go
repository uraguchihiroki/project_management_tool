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

// GET /api/v1/organizations/:orgId/groups
func (h *GroupHandler) List(c echo.Context) error {
	orgID, _, authErr := requireOrgParam(c, "orgId")
	if authErr != nil {
		return authErr
	}
	groups, err := h.groupService.ListByOrganization(orgID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": groups})
}

// POST /api/v1/organizations/:orgId/groups
func (h *GroupHandler) Create(c echo.Context) error {
	orgID, _, authErr := requireOrgParam(c, "orgId")
	if authErr != nil {
		return authErr
	}
	type Request struct {
		Name string `json:"name"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "グループ名は必須です")
	}
	if len(req.Name) > 200 {
		return echo.NewHTTPError(http.StatusBadRequest, "グループ名は200文字以内で指定してください")
	}
	group, err := h.groupService.Create(orgID, req.Name)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": group})
}

// PUT /api/v1/organizations/:orgId/groups/:id
func (h *GroupHandler) Update(c echo.Context) error {
	orgID, _, authErr := requireOrgParam(c, "orgId")
	if authErr != nil {
		return authErr
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid group id")
	}
	type Request struct {
		Name string `json:"name"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	group, err := h.groupService.Get(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "group not found")
	}
	if group.OrganizationID != orgID {
		return echo.NewHTTPError(http.StatusNotFound, "group not found")
	}
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "グループ名は必須です")
	}
	if len(req.Name) > 200 {
		return echo.NewHTTPError(http.StatusBadRequest, "グループ名は200文字以内で指定してください")
	}
	updated, err := h.groupService.Update(id, req.Name)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": updated})
}

// PUT /api/v1/organizations/:orgId/groups/reorder
func (h *GroupHandler) Reorder(c echo.Context) error {
	orgID, _, authErr := requireOrgParam(c, "orgId")
	if authErr != nil {
		return authErr
	}
	type Request struct {
		IDs []string `json:"ids"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ids := make([]uuid.UUID, 0, len(req.IDs))
	for _, s := range req.IDs {
		id, err := uuid.Parse(s)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id: "+s)
		}
		ids = append(ids, id)
	}
	if err := h.groupService.Reorder(orgID, ids); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// DELETE /api/v1/organizations/:orgId/groups/:id
func (h *GroupHandler) Delete(c echo.Context) error {
	orgID, _, authErr := requireOrgParam(c, "orgId")
	if authErr != nil {
		return authErr
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid group id")
	}
	group, err := h.groupService.Get(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "group not found")
	}
	if group.OrganizationID != orgID {
		return echo.NewHTTPError(http.StatusNotFound, "group not found")
	}
	if err := h.groupService.Delete(id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// GET /api/v1/users/:id/groups?org_id=xxx
func (h *GroupHandler) GetUserGroups(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}
	orgID, _, authErr := requireOrgQuery(c, "org_id")
	if authErr != nil {
		return authErr
	}
	groups, err := h.groupService.GetUserGroups(orgID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": groups})
}

// PUT /api/v1/users/:id/groups?org_id=xxx
func (h *GroupHandler) SetUserGroups(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}
	orgID, _, authErr := requireOrgQuery(c, "org_id")
	if authErr != nil {
		return authErr
	}
	type Request struct {
		GroupIDs []string `json:"group_ids"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ids := make([]uuid.UUID, 0, len(req.GroupIDs))
	for _, s := range req.GroupIDs {
		id, err := uuid.Parse(s)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid group_id: "+s)
		}
		ids = append(ids, id)
	}
	if err := h.groupService.SetUserGroups(orgID, userID, ids); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "groups updated"})
}
