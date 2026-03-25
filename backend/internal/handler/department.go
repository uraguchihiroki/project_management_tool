package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/uraguchihiroki/project_management_tool/internal/service"
)

type DepartmentHandler struct {
	deptService service.DepartmentService
}

func NewDepartmentHandler(deptService service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{deptService: deptService}
}

// GET /api/v1/organizations/:orgId/departments
func (h *DepartmentHandler) List(c echo.Context) error {
	orgID, _, authErr := requireOrgParam(c, "orgId")
	if authErr != nil {
		return authErr
	}
	depts, err := h.deptService.ListByOrganization(orgID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": depts})
}

// POST /api/v1/organizations/:orgId/departments
func (h *DepartmentHandler) Create(c echo.Context) error {
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
		return echo.NewHTTPError(http.StatusBadRequest, "部署名は必須です")
	}
	if len(req.Name) > 200 {
		return echo.NewHTTPError(http.StatusBadRequest, "部署名は200文字以内で指定してください")
	}
	dept, err := h.deptService.Create(orgID, req.Name)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": dept})
}

// PUT /api/v1/organizations/:orgId/departments/:id
func (h *DepartmentHandler) Update(c echo.Context) error {
	orgID, _, authErr := requireOrgParam(c, "orgId")
	if authErr != nil {
		return authErr
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid department id")
	}
	type Request struct {
		Name string `json:"name"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	dept, err := h.deptService.Get(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "department not found")
	}
	if dept.OrganizationID != orgID {
		return echo.NewHTTPError(http.StatusNotFound, "department not found")
	}
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "部署名は必須です")
	}
	if len(req.Name) > 200 {
		return echo.NewHTTPError(http.StatusBadRequest, "部署名は200文字以内で指定してください")
	}
	updated, err := h.deptService.Update(id, req.Name)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": updated})
}

// PUT /api/v1/organizations/:orgId/departments/reorder
func (h *DepartmentHandler) Reorder(c echo.Context) error {
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
	if err := h.deptService.Reorder(orgID, ids); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// DELETE /api/v1/organizations/:orgId/departments/:id
func (h *DepartmentHandler) Delete(c echo.Context) error {
	orgID, _, authErr := requireOrgParam(c, "orgId")
	if authErr != nil {
		return authErr
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid department id")
	}
	dept, err := h.deptService.Get(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "department not found")
	}
	if dept.OrganizationID != orgID {
		return echo.NewHTTPError(http.StatusNotFound, "department not found")
	}
	if err := h.deptService.Delete(id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// GET /api/v1/users/:id/departments?org_id=xxx
func (h *DepartmentHandler) GetUserDepartments(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}
	orgID, _, authErr := requireOrgQuery(c, "org_id")
	if authErr != nil {
		return authErr
	}
	depts, err := h.deptService.GetUserDepartments(orgID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": depts})
}

// PUT /api/v1/users/:id/departments?org_id=xxx
func (h *DepartmentHandler) SetUserDepartments(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}
	orgID, _, authErr := requireOrgQuery(c, "org_id")
	if authErr != nil {
		return authErr
	}
	type Request struct {
		DepartmentIDs []string `json:"department_ids"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ids := make([]uuid.UUID, 0, len(req.DepartmentIDs))
	for _, s := range req.DepartmentIDs {
		id, err := uuid.Parse(s)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid department_id: "+s)
		}
		ids = append(ids, id)
	}
	if err := h.deptService.SetUserDepartments(orgID, userID, ids); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "departments updated"})
}
