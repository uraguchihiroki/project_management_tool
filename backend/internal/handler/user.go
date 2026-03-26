package handler

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/uraguchihiroki/project_management_tool/internal/auth"
	authmw "github.com/uraguchihiroki/project_management_tool/internal/middleware"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/service"
	"gorm.io/gorm"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) List(c echo.Context) error {
	orgID, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	if !isSuperAdmin && orgID != nil {
		users, err := h.userService.ListByOrg(*orgID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"data": users})
	}
	users, err := h.userService.List()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": users})
}

func (h *UserHandler) Get(c echo.Context) error {
	orgID, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}
	user, err := h.userService.Get(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	if !isSuperAdmin && (orgID == nil || user.OrganizationID != *orgID) {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": user})
}

func (h *UserHandler) Create(c echo.Context) error {
	type Request struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "名前は必須です")
	}
	if len(req.Name) > 100 {
		return echo.NewHTTPError(http.StatusBadRequest, "名前は100文字以内で指定してください")
	}
	if req.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "メールアドレスは必須です")
	}
	if len(req.Email) > 255 {
		return echo.NewHTTPError(http.StatusBadRequest, "メールアドレスは255文字以内で指定してください")
	}
	user, err := h.userService.Create(req.Name, req.Email)
	if err != nil {
		if errors.Is(err, service.ErrDuplicateEmailInOrg) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": user})
}

// POST /api/v1/admin/login
func (h *UserHandler) AdminLogin(c echo.Context) error {
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
	user, err := h.userService.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusUnauthorized, "メールアドレスが見つかりません")
		}
		// DB 障害などを 401 にしない（フロントが「未登録」と誤認しないよう明示）
		return echo.NewHTTPError(http.StatusInternalServerError, "ユーザー検索に失敗しました: "+err.Error())
	}
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "メールアドレスが見つかりません")
	}
	token, err := auth.GenerateUserToken(user.ID, user.OrganizationID, user.IsAdmin)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "トークン発行に失敗しました: "+err.Error())
	}
	// ネストした関連の JSON 変換トラブルを避けるためログイン応答はフラットな map のみ
	userOut := map[string]interface{}{
		"id":              user.ID.String(),
		"key":             user.Key,
		"organization_id": user.OrganizationID.String(),
		"name":            user.Name,
		"email":           user.Email,
		"is_admin":        user.IsAdmin,
		"is_org_admin":    user.IsOrgAdmin,
		"joined_at":       user.JoinedAt,
		"created_at":      user.CreatedAt,
	}
	if user.AvatarURL != nil {
		userOut["avatar_url"] = *user.AvatarURL
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"user":  userOut,
			"token": token,
		},
	})
}

// POST /api/v1/admin/switch-organization
// 同一メールで複数組織に所属する場合、JWT の組織スコープを切り替える。
func (h *UserHandler) SwitchOrganization(c echo.Context) error {
	claims, ok := authmw.GetClaims(c)
	if !ok || claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	if claims.IsSuperAdmin {
		return echo.NewHTTPError(http.StatusBadRequest, "super admin cannot use this endpoint")
	}
	type Request struct {
		OrganizationID string `json:"organization_id"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.OrganizationID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "organization_id is required")
	}
	targetOrgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid organization_id")
	}
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}
	currentUser, err := h.userService.Get(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	userInOrg, err := h.userService.FindByEmailAndOrg(targetOrgID, currentUser.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusForbidden, "not a member of this organization")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	token, err := auth.GenerateUserToken(userInOrg.ID, targetOrgID, userInOrg.IsAdmin)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to issue token")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"user":  userInOrg,
			"token": token,
		},
	})
}

// GET /api/v1/admin/users （ロール付きユーザー一覧、org_id で組織フィルタ）
func (h *UserHandler) ListWithRoles(c echo.Context) error {
	orgIDScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	var users []model.User
	var err error
	if orgIDStr := c.QueryParam("org_id"); orgIDStr != "" {
		orgID, parseErr := uuid.Parse(orgIDStr)
		if parseErr != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid org_id")
		}
		if !isSuperAdmin && (orgIDScope == nil || orgID != *orgIDScope) {
			return echo.NewHTTPError(http.StatusForbidden, "forbidden for this organization")
		}
		users, err = h.userService.ListByOrg(orgID)
	} else {
		if !isSuperAdmin && orgIDScope != nil {
			users, err = h.userService.ListByOrg(*orgIDScope)
		} else {
			users, err = h.userService.ListWithRoles()
		}
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": users})
}

// POST /api/v1/admin/users （組織にユーザーを追加）
func (h *UserHandler) CreateForOrg(c echo.Context) error {
	orgIDScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	type Request struct {
		OrgID string `json:"org_id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.OrgID == "" || req.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "org_id and email are required")
	}
	orgID, err := uuid.Parse(req.OrgID)
	if !isSuperAdmin && (orgIDScope == nil || orgID != *orgIDScope) {
		return echo.NewHTTPError(http.StatusForbidden, "forbidden for this organization")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid org_id")
	}
	user, err := h.userService.CreateForOrg(orgID, req.Name, req.Email)
	if err != nil {
		if errors.Is(err, service.ErrDuplicateEmailInOrg) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": user})
}

// PUT /api/v1/admin/users/:id （ユーザー名を更新）
func (h *UserHandler) UpdateUser(c echo.Context) error {
	orgIDScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	id, err := uuid.Parse(c.Param("id"))
	existing, err := h.userService.Get(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	if !isSuperAdmin && (orgIDScope == nil || existing.OrganizationID != *orgIDScope) {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
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
	if err := h.userService.Update(id, req.Name); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "updated"})
}

// DELETE /api/v1/admin/users/:id （組織からユーザーを除外）
func (h *UserHandler) RemoveFromOrg(c echo.Context) error {
	orgIDScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}
	orgIDStr := c.QueryParam("org_id")
	if orgIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "org_id is required")
	}
	orgID, err := uuid.Parse(orgIDStr)
	if !isSuperAdmin && (orgIDScope == nil || orgID != *orgIDScope) {
		return echo.NewHTTPError(http.StatusForbidden, "forbidden for this organization")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid org_id")
	}
	if err := h.userService.RemoveFromOrg(orgID, userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// PUT /api/v1/users/:id/admin
func (h *UserHandler) SetAdmin(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}
	type Request struct {
		IsAdmin bool `json:"is_admin"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.userService.SetAdmin(id, req.IsAdmin); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "updated"})
}
